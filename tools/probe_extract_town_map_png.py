#!/usr/bin/env python3
"""把 `TOWNn.0` 的地圖格子陣列與磚塊集合抽成 PNG（RE-059 的版面）。

段內配置由 `sub_115D1` ＋ `sub_1D1C6` 定案：

  `0x0000`–`0x0FFF`  64×64 的格子陣列，索引 `(X << 6) + Y`，一格一 byte ＝ 磚塊編號
  `0x1000`–`0x8FFF`  256 塊磚，每塊 128 byte ＝ 4 個 EGA 平面 × 32 byte，plane-major

payload 0／1／2 解碼後首尾相接（0x0000／0x3000／0x6000）就是這塊區域。

輸出兩種 PNG：
  `town{n}-tiles.png`  256×256 的磚塊 atlas，每列 16 塊，EGA 16 色
  `town{n}-map.png`    64×64 的 8-bit 灰階，像素值 ＝ 該格的磚塊編號（X 橫向、Y 縱向）

地圖那張是**資料的無損 PNG 表示**，不是原版畫面；runtime 只讀 PNG 的契約因此維持。

  python3 tools/probe_extract_town_map_png.py --zip /in/kol.zip \\
      --out workplace/town-map --probe-out docs/re/town-map-ega-visual-probe.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import struct
import sys
import zipfile
import zlib

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_ega_assets import encode_rgba_png, verify_png_bytes  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

TOWN_COUNT = 8
PAYLOAD_DECODED = 12289        # 12,288 ＋ LZW 終止 byte
SEGMENT_PARTS = ((0, 0x0000), (1, 0x3000), (2, 0x6000))
MAP_BYTES = 0x1000
MAP_SIDE = 64
TILE_BASE = 0x1000
TILE_BYTES = 128
TILE_SIDE = 16
TILE_COUNT = 256
ATLAS_COLUMNS = 16


def build_segment(volumes: dict[str, bytes], root: str) -> tuple[dict, bytes, list[dict]]:
    contract, parts = payloads(volumes, root)
    seg = bytearray(0x9000)
    detail = []
    for index, offset in SEGMENT_PARTS:
        decoded = lzw_decode(parts[index])
        if len(decoded) != PAYLOAD_DECODED:
            raise SystemExit(f"{root} payload {index} 解碼長度 {len(decoded)} 不是 {PAYLOAD_DECODED}")
        seg[offset:offset + 0x3000] = decoded[:0x3000]
        detail.append({
            "payload_index": index,
            "segment_offset": f"0x{offset:04x}",
            "compressed_bytes": len(parts[index]),
            "decoded_bytes": len(decoded),
            "decoded_sha256": hashlib.sha256(decoded).hexdigest(),
        })
    return contract, bytes(seg), detail


def tile_pixels(seg: bytes, index: int) -> bytes:
    """一塊磚 ＝ 4 個平面 × 32 byte（16 列 × 2 byte），plane-major。"""
    base = TILE_BASE + index * TILE_BYTES
    planes = [seg[base + p * 32:base + (p + 1) * 32] for p in range(4)]
    out = bytearray(TILE_SIDE * TILE_SIDE)
    for n in range(TILE_SIDE * TILE_SIDE):
        bit = 0x80 >> (n & 7)
        out[n] = sum((1 << p) if planes[p][n >> 3] & bit else 0 for p in range(4))
    return bytes(out)


def atlas(seg: bytes) -> bytes:
    rows = ATLAS_COLUMNS * TILE_SIDE
    out = bytearray(rows * rows)
    for index in range(TILE_COUNT):
        px = tile_pixels(seg, index)
        ox = (index % ATLAS_COLUMNS) * TILE_SIDE
        oy = (index // ATLAS_COLUMNS) * TILE_SIDE
        for y in range(TILE_SIDE):
            start = (oy + y) * rows + ox
            out[start:start + TILE_SIDE] = px[y * TILE_SIDE:(y + 1) * TILE_SIDE]
    return bytes(out)


def _chunk(tag: bytes, body: bytes) -> bytes:
    payload_bytes = tag + body
    return (struct.pack(">I", len(body)) + payload_bytes
            + struct.pack(">I", zlib.crc32(payload_bytes) & 0xFFFFFFFF))


def encode_gray_png(values: bytes, width: int, height: int) -> bytes:
    if len(values) != width * height:
        raise SystemExit("灰階 PNG 的像素數不符")
    raw = bytearray()
    for y in range(height):
        raw.append(0)
        raw.extend(values[y * width:(y + 1) * width])
    return (b"\x89PNG\r\n\x1a\n"
            + _chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 0, 0, 0, 0))
            + _chunk(b"IDAT", zlib.compress(bytes(raw), 9))
            + _chunk(b"IEND", b""))


def verify_gray_png(data: bytes, width: int, height: int, values: bytes) -> None:
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit("灰階 PNG 簽章不符")
    w, h, depth, colour = struct.unpack(">IIBB", data[16:26])
    if (w, h, depth, colour) != (width, height, 8, 0):
        raise SystemExit("灰階 PNG 表頭不符")
    idat = b""
    pos = 8
    while pos < len(data):
        length = struct.unpack(">I", data[pos:pos + 4])[0]
        tag = data[pos + 4:pos + 8]
        if tag == b"IDAT":
            idat += data[pos + 8:pos + 8 + length]
        pos += 12 + length
    raw = zlib.decompress(idat)
    stride = width + 1
    for y in range(height):
        if raw[y * stride] != 0:
            raise SystemExit("灰階 PNG 出現非零 filter")
        if raw[y * stride + 1:(y + 1) * stride] != values[y * width:(y + 1) * width]:
            raise SystemExit("灰階 PNG 重讀的像素與來源不符")


def map_image(seg: bytes) -> bytes:
    """X 橫向、Y 縱向；索引 `(X << 6) + Y`。"""
    out = bytearray(MAP_SIDE * MAP_SIDE)
    for y in range(MAP_SIDE):
        for x in range(MAP_SIDE):
            out[y * MAP_SIDE + x] = seg[(x << 6) + y]
    return bytes(out)


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    p.add_argument("--probe-out", required=True)
    args = p.parse_args()

    raw = pathlib.Path(args.zip).read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    outdir = pathlib.Path(args.out)
    outdir.mkdir(parents=True, exist_ok=True)
    records = []
    with zipfile.ZipFile(args.zip) as zf:
        for town in range(TOWN_COUNT):
            root = f"TOWN{town}"
            volumes = read_volumes(zf, root)
            if not volumes:
                raise SystemExit(f"{root} 不在固定 ZIP 裡")
            contract, seg, detail = build_segment(volumes, root)

            grid = map_image(seg)
            grid_png = encode_gray_png(grid, MAP_SIDE, MAP_SIDE)
            verify_gray_png(grid_png, MAP_SIDE, MAP_SIDE, grid)

            pixels = atlas(seg)
            side = ATLAS_COLUMNS * TILE_SIDE
            tiles_png = encode_rgba_png(pixels, bytes([1]) * len(pixels), side, side)
            verify_png_bytes(tiles_png, side, side)

            (outdir / f"town{town}-map.png").write_bytes(grid_png)
            (outdir / f"town{town}-tiles.png").write_bytes(tiles_png)

            unique = len({bytes(seg[TILE_BASE + n * TILE_BYTES:
                                    TILE_BASE + (n + 1) * TILE_BYTES])
                          for n in range(TILE_COUNT)})
            used = sorted(set(grid))
            records.append({
                "town_index": town,
                "archive_volumes": sorted(volumes),
                "archive_payload_count": contract["count"],
                "segment_payloads": detail,
                "map": {
                    "file": f"town{town}-map.png",
                    "width_pixels": MAP_SIDE, "height_pixels": MAP_SIDE,
                    "cell_bytes": 1, "index": "(X << 6) + Y",
                    "distinct_tile_ids": len(used),
                    "tile_id_range": [used[0], used[-1]],
                    "png_sha256": hashlib.sha256(grid_png).hexdigest(),
                },
                "tiles": {
                    "file": f"town{town}-tiles.png",
                    "width_pixels": side, "height_pixels": side,
                    "tile_count": TILE_COUNT, "tile_side": TILE_SIDE,
                    "atlas_columns": ATLAS_COLUMNS,
                    "byte_unique_tiles": unique,
                    "colour_count": len(set(pixels)),
                    "png_sha256": hashlib.sha256(tiles_png).hexdigest(),
                },
            })

    probe = {
        "schema": "kol-town-map-ega-visual-probe-v1",
        "input_name": "Knights-of-legend.zip",
        "input_sha256": digest,
        "evidence": "RE-059：sub_115D1 的格子索引與 sub_1D1C6 的 shl ax,7 ＋ 段 +0x100",
        "status": "八個城鎮的 payload 0/1/2 都解出 12,289 byte；版面契約逐城驗過",
        "human_inspection": "TOWN7 的格子陣列印成字元圖可見建築矩形、道路與外圍空地；"
                            "12×6 視窗在鏡頭 (57, 8) 與原版畫面 67/67 相同",
        "rights_class": "private-repo 可保存的原版衍生 PNG；不含公開散布權",
        "town_count": TOWN_COUNT,
        "records": records,
    }
    out = pathlib.Path(args.probe_out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(probe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"towns": TOWN_COUNT,
                      "map_png": [r["map"]["png_sha256"][:12] for r in records],
                      "tiles_png": [r["tiles"]["png_sha256"][:12] for r in records],
                      "distinct_tile_ids": [r["map"]["distinct_tile_ids"] for r in records],
                      "byte_unique_tiles": [r["tiles"]["byte_unique_tiles"] for r in records]},
                     ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
