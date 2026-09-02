#!/usr/bin/env python3
"""探測 `TOWNn.0` 城鎮資料檔：容器契約、payload 解碼，以及畫面磚塊的比對。

`TOWNn.0` 從來沒有被分析過——既有的清冊工具只認名字含 `PIC` 的檔案，
所以七個 50–65 KB 的城鎮資料檔一直不在 inventory 裡。這支探針做三件事：

1. 用 `*PICS` 同一套容器契約（20 byte header、2 byte 索引、最多五個分卷）解析。
2. 用 `sub_1A26B` 的 LZW 解碼每個 payload，回報長度分佈。
3. 把原版戶外畫面上的 16×16 磚塊轉成 high-nibble-first 的 4-bit 封裝 bytes，
   在解碼後的 payload 裡搜尋——命中就證實磚圖住在城鎮資料檔裡。

只輸出長度、雜湊、命中位置與統計，不輸出原版素材本身。

  python3 tools/probe_town_data.py --zip /in/kol.zip --town TOWN7 \\
      --frame docs/verification/.../frame-077.png --out 收據.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import struct
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from inventory_image_archives import EXPECTED_ZIP_SHA256, archive_contract  # noqa: E402
from png_io import read_png  # noqa: E402
from probe_extract_comb_png import PALETTE  # noqa: E402
from probe_image_lzw import decode as lzw_decode  # noqa: E402

VIEWPORT = (104, 58, 192, 96)
TILE = 16


def nibble_rle(payload: bytes) -> bytes:
    """KOLTOWN `sub_1C80F` 的 nibble RLE。

    版面：`[0..2)` 是 big-endian 的 N，`[2..N)` 是控制位元組，`[N..)` 是 nibble 資料流。
    控制位元組最高位為 0 時，`(值+1)` 是重複次數，重複的是資料流的下一個 nibble；
    最高位為 1 時，`(值-0x7f)` 是照抄的 nibble 數。輸出把 nibble 兩兩打包成 byte，
    高位在前。
    """
    if len(payload) < 4:
        return b""
    end = int.from_bytes(payload[0:2], "big")
    if not 2 < end <= len(payload):
        return b""
    control = payload[2:end]
    data = payload[end:]
    out = bytearray()
    pending = None          # 尚未輸出的高位 nibble
    src_pos = 0
    src_high = True         # 下一個要讀的 nibble 是不是高位

    def next_nibble() -> int | None:
        nonlocal src_pos, src_high
        if src_pos >= len(data):
            return None
        byte = data[src_pos]
        if src_high:
            src_high = False
            return byte >> 4
        src_high = True
        src_pos += 1
        return byte & 0x0F

    def emit(value: int) -> None:
        nonlocal pending
        if pending is None:
            pending = value
        else:
            out.append((pending << 4) | value)
            pending = None

    for code in control:
        if code & 0x80:
            count = code - 0x7F
            for _ in range(count):
                n = next_nibble()
                if n is None:
                    return bytes(out)
                emit(n)
        else:
            count = code + 1
            n = next_nibble()
            if n is None:
                return bytes(out)
            for _ in range(count):
                emit(n)
    if pending is not None:
        out.append(pending << 4)
    return bytes(out)


def read_volumes(zf: zipfile.ZipFile, root: str) -> dict[str, bytes]:
    out = {}
    for info in zf.infolist():
        name = info.filename.split("/")[-1]
        if name.startswith(root + "."):
            out[name] = zf.read(info)
    return out


def payloads(volumes: dict[str, bytes], root: str) -> tuple[dict, list[bytes]]:
    first = volumes[f"{root}.0"]
    contract = archive_contract(first)
    if contract is None:
        raise SystemExit(f"{root}.0 不符合既有的容器契約")
    count = contract["count"]
    offsets = list(struct.unpack_from(f"<{count}H", first, 20))
    last_length = contract["last_payload_length"]
    # 分卷：payload 依序接在 .0 之後的 .1、.2…，與 *PICS 相同。
    blob = bytearray(first)
    volume_order = sorted(k for k in volumes if k != f"{root}.0")
    for name in volume_order:
        blob.extend(volumes[name])
    out = []
    for i, off in enumerate(offsets):
        end = offsets[i + 1] if i + 1 < count else off + last_length
        out.append(bytes(blob[off:end]))
    return contract, out


def palette_index() -> dict[tuple[int, int, int], int]:
    return {tuple(rgb): i for i, rgb in enumerate(PALETTE)}


def tile_indices(rgb: bytes, width: int, x0: int, y0: int,
                 index: dict[tuple[int, int, int], int]) -> list[list[int]] | None:
    """把畫面上的 16×16 方塊讀成色盤索引矩陣。"""
    rows = []
    for y in range(TILE):
        row = []
        for x in range(TILE):
            o = ((y0 + y) * width + x0 + x) * 3
            key = (rgb[o], rgb[o + 1], rgb[o + 2])
            if key not in index:
                return None
            row.append(index[key])
        rows.append(row)
    return rows


def encode(rows: list[list[int]], kind: str) -> bytes:
    """把索引矩陣編成候選的儲存格式。磚圖到底是哪一種要靠搜尋命中來決定。"""
    out = bytearray()
    if kind in ("packed-high", "packed-low"):
        high = kind == "packed-high"
        for row in rows:
            for x in range(0, TILE, 2):
                a, b = row[x], row[x + 1]
                out.append((a << 4) | b if high else (b << 4) | a)
        return bytes(out)
    if kind in ("planar-plane-major", "planar-row-major"):
        planes = []
        for plane in range(4):
            plane_rows = []
            for row in rows:
                for x in range(0, TILE, 8):
                    byte = 0
                    for bit in range(8):
                        if row[x + bit] >> plane & 1:
                            byte |= 0x80 >> bit
                    plane_rows.append(byte)
            planes.append(plane_rows)
        if kind == "planar-plane-major":
            for plane_rows in planes:
                out.extend(plane_rows)
        else:
            per_row = TILE // 8
            for r in range(TILE):
                for plane_rows in planes:
                    out.extend(plane_rows[r * per_row:(r + 1) * per_row])
        return bytes(out)
    raise SystemExit(f"未知的編碼 {kind}")


ENCODINGS = ("packed-high", "packed-low", "planar-plane-major", "planar-row-major")


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--town", default="TOWN7")
    p.add_argument("--frame", default=None)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    raw = pathlib.Path(args.zip).read_bytes()
    if hashlib.sha256(raw).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")
    with zipfile.ZipFile(args.zip) as zf:
        volumes = read_volumes(zf, args.town)
        contract, blobs = payloads(volumes, args.town)

    decoded = []
    for i, blob in enumerate(blobs):
        try:
            data = lzw_decode(blob)
            decoded.append({"index": i, "compressed_bytes": len(blob),
                            "decoded_bytes": len(data),
                            "sha256": hashlib.sha256(data).hexdigest()})
        except Exception as exc:                       # noqa: BLE001
            decoded.append({"index": i, "compressed_bytes": len(blob),
                            "error": type(exc).__name__})
            data = b""
        decoded[-1]["_data"] = data

    matches = []
    tiles_checked = 0
    if args.frame:
        frame = read_png(args.frame)
        rgb = frame.rgb()
        index = palette_index()
        X0, Y0, MW, MH = VIEWPORT
        for ty in range(MH // TILE):
            for tx in range(MW // TILE):
                rows = tile_indices(rgb, frame.width, X0 + tx * TILE, Y0 + ty * TILE, index)
                if rows is None:
                    continue
                tiles_checked += 1
                for kind in ENCODINGS:
                    needle = encode(rows, kind)
                    hit = None
                    for rec in decoded:
                        pos = rec["_data"].find(needle)
                        if pos >= 0:
                            hit = {"tile_x": tx, "tile_y": ty, "encoding": kind,
                                   "payload": rec["index"], "offset": pos}
                            break
                    if hit:
                        matches.append(hit)
                        break

    # 第二階段：磚塊編號排成的地圖。把畫面上一列 12 個磚的編號當指紋，
    # 在 2048 byte 的 payload 裡找連續出現的位置——命中就代表那些 payload 是地圖配置。
    layout = {"searched": False}
    if matches:
        by_pos = {(m["tile_y"], m["tile_x"]): m for m in matches}
        rows_found = []
        for ty in range(MH // TILE):
            row = [by_pos.get((ty, tx)) for tx in range(MW // TILE)]
            if all(r is not None for r in row):
                rows_found.append((ty, row))
        def gid(m: dict) -> int:
            return m["payload"] * 96 + m["offset"] // 128

        def local(m: dict) -> int:
            return m["offset"] // 128

        def bank_hi(m: dict) -> int:
            return (m["payload"] << 8) | (m["offset"] // 128)

        candidates = {
            "byte-local": (lambda m: bytes([local(m) & 0xFF])),
            "byte-global-96": (lambda m: bytes([gid(m) & 0xFF])),
            "u16le-global-96": (lambda m: gid(m).to_bytes(2, "little")),
            "u16be-global-96": (lambda m: gid(m).to_bytes(2, "big")),
            "u16le-bank-hi": (lambda m: bank_hi(m).to_bytes(2, "little")),
            "u16be-bank-hi": (lambda m: bank_hi(m).to_bytes(2, "big")),
        }
        hits = []
        for name, fn in candidates.items():
            for ty, row in rows_found:
                needle = b"".join(fn(m) for m in row)
                for rec in decoded:
                    if len(rec["_data"]) < 256:
                        continue
                    pos = rec["_data"].find(needle)
                    if pos >= 0:
                        hits.append({"numbering": name, "tile_row": ty,
                                     "payload": rec["index"], "offset": pos})
                        break
        layout = {"searched": True, "rows_available": [t for t, _ in rows_found],
                  "hits": hits}

    # 第三階段：不靠編號，只靠「哪幾格畫同一塊磚」的相等關係去找地圖。
    # 如果地圖是每格一個索引（不管索引怎麼編），畫面上相同的磚在資料裡也會是相同的值。
    pattern = {"searched": False}
    if matches:
        by_pos = {(m["tile_y"], m["tile_x"]): (m["payload"], m["offset"] // 128)
                  for m in matches}
        cols, rows_n = MW // TILE, MH // TILE
        cells = [(ty, tx) for ty in range(rows_n) for tx in range(cols)
                 if (ty, tx) in by_pos]
        # 相等關係：把每一格對映成一個群組編號
        groups: dict[tuple[int, int], int] = {}
        for pos in cells:
            groups.setdefault(by_pos[pos], len(groups))
        want = {pos: groups[by_pos[pos]] for pos in cells}
        hits = []
        for rec in decoded:
            data = rec["_data"]
            if len(data) < 512:
                continue
            for cell_bytes in (1, 2):
                for stride in (16, 24, 32, 40, 48, 64, 96, 128):
                    span = stride * cell_bytes
                    for start in range(0, len(data) - span * rows_n):
                        seen: dict[int, tuple[int, int]] = {}
                        ok = True
                        for (ty, tx), g in want.items():
                            off = start + ty * span + tx * cell_bytes
                            v = int.from_bytes(data[off:off + cell_bytes], "little")
                            if v in seen and seen[v] != (g, v):
                                ok = False
                                break
                            prev = [k for k, val in seen.items() if val[0] == g]
                            if prev and prev[0] != v:
                                ok = False
                                break
                            seen[v] = (g, v)
                        if ok:
                            hits.append({"payload": rec["index"], "cell_bytes": cell_bytes,
                                         "stride": stride, "start": start})
                            break
                    if hits and hits[-1]["payload"] == rec["index"]:
                        break
                if hits and hits[-1]["payload"] == rec["index"]:
                    break
        pattern = {"searched": True, "distinct_tiles": len(groups),
                   "cells": len(cells), "hits": hits[:20], "hit_count": len(hits)}

    expanded = []
    for rec in decoded:
        data = rec["_data"]
        if len(data) in (2049, 513):
            out = nibble_rle(data)
            expanded.append({"payload": rec["index"], "payload_bytes": len(data),
                             "control_end": int.from_bytes(data[0:2], "big"),
                             "expanded_bytes": len(out),
                             "sha256": hashlib.sha256(out).hexdigest()})

    receipt = {
        "schema": "kol-town-data-probe-v1",
        "nibble_rle_expansion": expanded,
        "tile_layout_search": layout,
        "tile_pattern_search": pattern,
        "tool": "tools/probe_town_data.py",
        "town": args.town,
        "volumes": {k: len(v) for k, v in sorted(volumes.items())},
        "container_contract": contract,
        "payloads": [{k: v for k, v in rec.items() if not k.startswith("_")}
                     for rec in decoded],
        "frame": args.frame,
        "tiles_checked": tiles_checked,
        "tile_matches": matches,
        "criterion": "畫面磚塊轉成 4-bit 封裝 bytes 後，若能在解碼 payload 裡找到，"
                     "就證實磚圖住在城鎮資料檔裡",
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"payloads": len(decoded), "tiles_checked": tiles_checked,
                      "tile_matches": len(matches)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
