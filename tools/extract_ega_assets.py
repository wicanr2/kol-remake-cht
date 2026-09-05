#!/usr/bin/env python3
"""SPEC-002 的 production EGA 抽取器：把固定原版 ZIP 轉成 runtime 專用 PNG。

與 `probe_extract_*.py` 的差別在於這支是交付用的：原子化輸出、寫 provenance
manifest、重讀每張 PNG 驗證，並對照已版控的 visual probe 紀錄逐張核對 SHA-256。
任何一項不符即整批失敗，不留半套資產。

抽取計畫直接讀 `docs/re/*-visual-probe.json`：那些紀錄已經過證據審查，帶著自己的
IDA 摘要雜湊，逐張記錄尺寸與 PNG SHA-256。因此本工具不重新推導幾何，只負責可重生地
產出並驗證；上游 IDA 證據的入口記在 manifest 裡。

  python3 tools/extract_ega_assets.py --zip kol.zip --repo . --out assets/original-derived-png/ega

輸出根目錄必須位於 `assets/original-derived-png/`；runtime 只讀這裡的 PNG，
不得在執行時解析原版 archive。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import shutil
import struct
import sys
import zipfile
import zlib

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract  # noqa: E402
from probe_extract_comb_png import PALETTE, decode_nibbles  # noqa: E402
from probe_extract_fixed_charpics_png import decode_record  # noqa: E402
from probe_image_lzw import decode as lzw_decode  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-png")


def direct(archive: str, count: int, probe: str, *, index_key: str = "image_index",
           packed: bool = False, width: int | None = None, height: int | None = None,
           record_bytes: int | None = None, trailing: int = 0) -> dict:
    return {
        "archive": archive, "root": f"Knights-of-legend/{archive}",
        "payload_count": count, "kind": "direct_packed" if packed else "direct_rle",
        "probe": probe, "records": "records", "index_key": index_key,
        "fixed_size": (width, height) if width else None,
        "record_bytes": record_bytes, "trailing_bytes": trailing,
    }


def mask_pair(archive: str, count: int, probe: str, index_key: str,
              records_selector: dict | None = None) -> dict:
    return {
        "archive": archive, "root": f"Knights-of-legend/{archive}",
        "payload_count": count, "kind": "mask_pair",
        "probe": probe, "records": "records", "index_key": index_key,
        "records_selector": records_selector,
    }


FAMILIES: list[dict] = [
    # 直接彩圖（LZW → 二階 RLE nibbles → 4-bit EGA）
    direct("BOOKPICS", 14, "kolbook-bookpics-ega-visual-probe.json"),
    direct("BACKPICS", 14, "title-backpics-ega-visual-probe.json"),
    direct("ENDPICS", 7, "koltown-endpics-ega-visual-probe.json"),
    direct("COMBPICS", 7, "combpics-all-ega-visual-probe.json"),
    direct("AWD1PICS", 24, "awd1pics-ega-visual-probe.json"),
    direct("MISCPICS", 6, "miscpics-ega-visual-probe.json"),
    # 固定 40×38 packed nibbles，每筆末 1 byte 不進像素
    direct("CHARPICS", 80, "charpics-archive-ega-visual-probe.json",
           packed=True, width=40, height=38, record_bytes=760, trailing=1),
    # 馬廄賣的四匹（RE-120）。版面與 CHARPICS.0 相同，來源是 STEEDS.0——
    # `STEEDPIC` 是玩家角色自己那一張的 16 槽表，不是這一份。
    direct("STEEDS", 4, "steeds-ega-visual-probe.json",
           packed=True, width=40, height=38, record_bytes=760, trailing=1),
    # 彩圖＋AND-mask 配對
    mask_pair("MONPICS", 64, "monpics-ega-visual-probe.json", "monster_index"),
    mask_pair("PERPICS", 12, "perpics-ega-visual-probe.json", "portrait_index"),
    mask_pair("WEAPPICS", 56, "weappics-ega-visual-probe.json", "group"),
    mask_pair("HANDPICS", 24, "handpics-ega-visual-probe.json", "group"),
    mask_pair("SHLDPICS", 8, "shldpics-ega-visual-probe.json", "group"),
]
for _name in ("HEADPICS", "TORPICS", "LEGSPICS"):
    FAMILIES.append(mask_pair(_name, 108, "bodypics-ega-visual-probe.json", "group",
                              records_selector={"list": "archives", "field": "archive",
                                                "value": f"{_name}.0"}))

FAMILIES.append({
    "archive": "KOLMPICS", "root": "Knights-of-legend/KOLMPICS",
    "payload_count": 83, "kind": "kolmpics",
    "probe": "kolmpics-all-ega-visual-probe.json",
    "records": "records", "index_key": "image_index",
    "fixed_size": None, "record_bytes": None, "trailing_bytes": 0,
})
for _town in range(8):
    FAMILIES.append({
        "archive": f"TWN{_town}PICS", "root": f"Knights-of-legend/TWN{_town}PICS",
        "payload_count": (25 if _town == 0 else None), "kind": "twn",
        "probe": "twn-archives-ega-visual-probe.json",
        "records": "records", "index_key": "image_index",
        "record_filter": {"field": "town_index", "value": _town},
        "fixed_size": None, "record_bytes": None, "trailing_bytes": 0,
    })

FAMILIES.append({
    "archive": "KOLPICS", "root": "Knights-of-legend/KOLPICS",
    "payload_count": 14, "kind": "kolpics",
    "probe": "kolpics-ega-visual-probe.json",
    "records": "records", "index_key": "image_index",
    "fixed_size": None, "record_bytes": None, "trailing_bytes": 0,
})
FAMILIES.append({
    "archive": "WEATPICS", "root": "Knights-of-legend/WEATPICS",
    "payload_count": 18, "kind": "weatpics",
    "probe": "koltown-weatpics-layered-visual-probe.json",
    "records": "records", "index_key": "image_index",
    "fixed_size": None, "record_bytes": None, "trailing_bytes": 0,
})

for _name in ("SET0PICS", "SET1PICS", "SET2PICS", "FACEPICS"):
    FAMILIES.append({
        "archive": _name, "root": f"Knights-of-legend/{_name}",
        "payload_count": 6, "kind": "direct_rle",
        "probe": "title-set-face-ega-visual-probe.json",
        "records": "records", "index_key": "image_index",
        "record_filter": {"field": "archive_root", "value": _name},
        "fixed_size": None, "record_bytes": None, "trailing_bytes": 0,
    })


# --------------------------------------------------------------------------
# PNG
# --------------------------------------------------------------------------

def _chunk(tag: bytes, body: bytes) -> bytes:
    payload = tag + body
    return (struct.pack(">I", len(body)) + payload
            + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))


def encode_rgba_png(indices: bytes, opaque: bytes, width: int, height: int) -> bytes:
    """與已驗證的 probe 抽取器同一套編碼，確保 production PNG 雜湊可對照。"""
    rows = []
    for y in range(height):
        colors = indices[y * width:(y + 1) * width]
        alphas = opaque[y * width:(y + 1) * width]
        row = bytearray(b"\0")
        for index, alpha in zip(colors, alphas):
            row.extend((*PALETTE[index], 0xFF if alpha else 0x00))
        rows.append(bytes(row))
    header = struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0)
    return (b"\x89PNG\r\n\x1a\n" + _chunk(b"IHDR", header)
            + _chunk(b"IDAT", zlib.compress(b"".join(rows), 9)) + _chunk(b"IEND", b""))


def verify_png_bytes(data: bytes, width: int, height: int) -> None:
    """重讀 PNG：檢查簽章、chunk CRC、IHDR 與解壓後的列數。"""
    if data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit("PNG 簽章不符")
    pos, idat, seen_ihdr = 8, b"", False
    while pos < len(data):
        length = struct.unpack_from(">I", data, pos)[0]
        tag = data[pos + 4:pos + 8]
        body = data[pos + 8:pos + 8 + length]
        crc = struct.unpack_from(">I", data, pos + 8 + length)[0]
        if zlib.crc32(tag + body) & 0xFFFFFFFF != crc:
            raise SystemExit(f"PNG chunk {tag!r} CRC 不符")
        if tag == b"IHDR":
            w, h, depth, ctype = struct.unpack(">IIBB", body[:10])
            if (w, h) != (width, height) or depth != 8 or ctype != 6:
                raise SystemExit(f"PNG IHDR 不符：{w}x{h} depth={depth} type={ctype}")
            seen_ihdr = True
        elif tag == b"IDAT":
            idat += body
        pos += 12 + length
    if not seen_ihdr:
        raise SystemExit("PNG 缺少 IHDR")
    if len(zlib.decompress(idat)) != height * (width * 4 + 1):
        raise SystemExit("PNG 解壓長度不符")


# --------------------------------------------------------------------------
# 抽取
# --------------------------------------------------------------------------

def load_probe(repo: pathlib.Path, family: dict) -> tuple[list[dict], dict]:
    path = repo / "docs/re" / family["probe"]
    raw = path.read_bytes()
    doc = json.loads(raw)
    selector = family.get("records_selector")
    if selector:
        entry = next((e for e in doc[selector["list"]]
                      if e[selector["field"]] == selector["value"]), None)
        if entry is None:
            raise SystemExit(f"{family['probe']} 沒有 {selector['value']} 的紀錄")
        records = entry["records"]
    else:
        records = doc[family["records"]]
    filt = family.get("record_filter")
    if filt:
        records = [r for r in records if r[filt["field"]] == filt["value"]]
    provenance = {
        "probe_file": family["probe"],
        "probe_sha256": hashlib.sha256(raw).hexdigest(),
        "probe_schema": doc.get("schema"),
        "ida_summary_name": doc.get("ida_summary_name"),
        "ida_summary_sha256": doc.get("ida_summary_sha256"),
    }
    return records, provenance


def payload(files: dict[str, bytes], item: dict) -> bytes:
    source = files[item["volume_name"]]
    return lzw_decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])


def unpack_and_mask(data: bytes, width: int, height: int) -> bytes:
    """AND-mask：bit 1 保留背景（透明），PNG alpha 取 1-mask；末 1 byte 不進遮罩。"""
    if width % 8:
        raise SystemExit(f"遮罩寬度 {width} 不是 8 的倍數")
    expected = width // 8 * height
    if len(data) != expected + 1:
        raise SystemExit(f"AND-mask 長度不符：{len(data)} != {expected + 1}")
    return bytes(1 - ((value >> bit) & 1)
                 for value in data[:expected] for bit in range(7, -1, -1))


def decode_planar(source: bytes, pixel_count: int) -> bytes:
    """KOLPICS 的 raw-planar renderer 資料：每 256 像素一個 4-plane 區塊。"""
    if pixel_count % 256:
        raise SystemExit("raw-planar 像素數不是 256 的倍數")
    block_count = pixel_count // 256
    if len(source) != block_count * 256 + 1:
        raise SystemExit("raw-planar 長度契約不符")
    output = bytearray(pixel_count)
    for block in range(block_count):
        row = source[block * 256:block * 256 + 128]
        planes = [row[p * 32:(p + 1) * 32] for p in range(4)]
        for x in range(256):
            bit = 0x80 >> (x & 7)
            output[block * 256 + x] = sum(
                (1 << plane) if planes[plane][x >> 3] & bit else 0 for plane in range(4))
    return bytes(output)


def kolmpics_shared_mask(files: dict[str, bytes]) -> bytes:
    """TWN 各城圖號 0 共用 KOLMPICS 圖號 28 的 56×57 AND-mask。"""
    contract = logical_archive_contract("Knights-of-legend/KOLMPICS", files)
    if contract is None:
        raise SystemExit("KOLMPICS archive contract 不符")
    return unpack_and_mask(payload(files, contract["items"][28]), 56, 57)


def extract_family(family: dict, files: dict[str, bytes], repo: pathlib.Path,
                   staging: pathlib.Path, shared: dict) -> dict:
    contract = logical_archive_contract(family["root"], files)
    if contract is None:
        raise SystemExit(f"{family['archive']} archive contract 不符")
    if family["payload_count"] and contract["count"] != family["payload_count"]:
        raise SystemExit(f"{family['archive']} archive contract 不符")
    items = contract["items"]
    records, provenance = load_probe(repo, family)

    outdir = staging / family["archive"]
    outdir.mkdir(parents=True, exist_ok=True)
    slug = family["archive"].lower()
    images = []

    skipped = []
    for record in records:
        index = record[family["index_key"]]
        if family["kind"] == "kolmpics":
            if record["kind"] == "placeholder":
                skipped.append({"unit_index": index, "reason": "一-byte placeholder"})
                continue
            width, height = record["width_pixels"], record["height_pixels"]
            stage1 = payload(files, items[index])
            if record["kind"] == "and-mask":
                opaque = unpack_and_mask(stage1, width, height)
                pixels = bytes([15]) * len(opaque)
            else:
                pixels = decode_nibbles(stage1)
                opaque = bytes([1]) * len(pixels)
            sources = [{"image_index": index, "role": record["kind"]}]
        elif family["kind"] == "kolpics":
            if record["kind"] == "placeholder":
                skipped.append({"unit_index": index, "reason": "一-byte placeholder"})
                continue
            width, height = record["width_pixels"], record["height_pixels"]
            stage1 = payload(files, items[index])
            pixels = (decode_planar(stage1, width * height)
                      if record["kind"] == "raw-planar" else decode_nibbles(stage1))
            opaque = bytes([1]) * len(pixels)
            sources = [{"image_index": index, "role": record["kind"]}]
        elif family["kind"] == "weatpics":
            width, height = record["width_pixels"], record["height_pixels"]
            stage1 = payload(files, items[index])
            if record["role"] == "and_mask_1bit":
                if len(stage1) != 497:
                    raise SystemExit(f"WEATPICS[{index}] 1-bit mask 長度不符")
                pixels = bytes(15 if stage1[y * 8 + x // 8] & (0x80 >> (x % 8)) else 0
                               for y in range(height) for x in range(width))
            else:
                pixels = decode_nibbles(stage1)
            opaque = bytes([1]) * len(pixels)
            sources = [{"image_index": index, "role": record["role"]}]
        elif family["kind"] == "twn":
            if record["disposition"] != "png":
                skipped.append({"unit_index": index, "reason": record["disposition"]})
                continue
            width, height = record["width_pixels"], record["height_pixels"]
            pixels = decode_nibbles(payload(files, items[index]))
            if record["alpha_source"] == "opaque":
                opaque = bytes([1]) * len(pixels)
                sources = [{"image_index": index, "role": "color"}]
            else:
                opaque = shared["kolmpics_28_mask"]
                sources = [{"image_index": index, "role": "color"},
                           {"archive": "KOLMPICS", "image_index": 28,
                            "role": "shared_and_mask"}]
        elif family["kind"] == "mask_pair":
            width = record["width_pixels"]
            height = record["height_pixels"]
            color = payload(files, items[record["color_image_index"]])
            mask = payload(files, items[record["mask_image_index"]])
            pixels = decode_nibbles(color)
            opaque = unpack_and_mask(mask, width, height)
            sources = [
                {"image_index": record["color_image_index"], "role": "color"},
                {"image_index": record["mask_image_index"], "role": "and_mask"},
            ]
        else:
            item = items[index]
            stage1 = payload(files, item)
            if family["fixed_size"]:
                width, height = family["fixed_size"]
            else:
                width, height = record["width_pixels"], record["height_pixels"]
            if family["record_bytes"] is not None:
                want = family["record_bytes"] + family["trailing_bytes"]
                if len(stage1) != want:
                    raise SystemExit(
                        f"{family['archive']}[{index}] 解碼長度不符：{len(stage1)} != {want}")
                stage1 = stage1[:family["record_bytes"]]
            pixels = (decode_record(stage1) if family["kind"] == "direct_packed"
                      else decode_nibbles(stage1))
            opaque = bytes([1]) * len(pixels)
            sources = [{"image_index": index, "role": "color"}]

        if len(pixels) != width * height:
            raise SystemExit(
                f"{family['archive']}[{index}] 像素數不符：{len(pixels)} != {width * height}")
        if any(value > 15 for value in pixels):
            raise SystemExit(f"{family['archive']}[{index}] 出現超出 4-bit 的索引")
        if len(opaque) != len(pixels):
            raise SystemExit(f"{family['archive']}[{index}] 遮罩長度與像素不符")

        png = encode_rgba_png(pixels, opaque, width, height)
        verify_png_bytes(png, width, height)
        digest = hashlib.sha256(png).hexdigest()
        if digest != record["png_sha256"]:
            raise SystemExit(
                f"{family['archive']}[{index}] PNG SHA-256 與已版控 visual probe 不符")
        suffix = f"-{record['role']}" if family["kind"] == "weatpics" else ""
        name = f"{slug}-{index:03d}{suffix}.png"
        (outdir / name).write_bytes(png)
        images.append({
            "unit_index": index,
            "file": f"{family['archive']}/{name}",
            "width": width,
            "height": height,
            "has_transparency": any(v == 0 for v in opaque),
            "png_sha256": digest,
            "source_payloads": sources,
        })

    unplanned = contract["count"] - len(records)
    return {
        "archive": family["archive"],
        "archive_root": family["root"],
        "source_payload_count": contract["count"],
        "plan_record_count": len(records),
        "payloads_without_plan_record": unplanned,
        "png_count": len(images),
        "skipped_count": len(skipped),
        "skipped": skipped,
        "composition": family["kind"],
        "evidence": provenance,
        "images": images,
    }


def extract_town_maps(files: dict[str, bytes], repo: pathlib.Path,
                      staging: pathlib.Path) -> list[dict]:
    """八個城鎮的地圖格子陣列與磚塊 atlas（RE-059 的段內版面）。"""
    from probe_extract_town_map_png import (  # 延後載入避免與本模組互相 import
        ATLAS_COLUMNS, MAP_SIDE, TILE_COUNT, TILE_SIDE, atlas, build_segment,
        encode_gray_png, map_image, verify_gray_png)

    probe_path = repo / "docs/re/town-map-ega-visual-probe.json"
    probe = json.loads(probe_path.read_text(encoding="utf-8"))
    provenance = {
        "visual_probe": probe_path.name,
        "visual_probe_sha256": hashlib.sha256(probe_path.read_bytes()).hexdigest(),
        "evidence": probe["evidence"],
        "status": probe["status"],
    }
    out = []
    for record in probe["records"]:
        town = record["town_index"]
        root = f"TOWN{town}"
        volumes = {name.split("/")[-1]: body for name, body in files.items()
                   if name.split("/")[-1].startswith(root + ".")}
        _, seg, detail = build_segment(volumes, root)
        if [d["decoded_sha256"] for d in detail] != \
                [d["decoded_sha256"] for d in record["segment_payloads"]]:
            raise SystemExit(f"{root} payload 解碼雜湊與 visual probe 不符")

        outdir = staging / root
        outdir.mkdir(parents=True, exist_ok=True)
        images = []

        grid = map_image(seg)
        grid_png = encode_gray_png(grid, MAP_SIDE, MAP_SIDE)
        verify_gray_png(grid_png, MAP_SIDE, MAP_SIDE, grid)
        digest = hashlib.sha256(grid_png).hexdigest()
        if digest != record["map"]["png_sha256"]:
            raise SystemExit(f"{root} 地圖 PNG SHA-256 與 visual probe 不符")
        (outdir / record["map"]["file"]).write_bytes(grid_png)
        images.append({
            "unit_index": 0, "file": f"{root}/{record['map']['file']}",
            "width": MAP_SIDE, "height": MAP_SIDE, "has_transparency": False,
            "png_sha256": digest, "png_format": "8-bit 灰階（colour type 0）",
            "role": "map_cell_array", "cell_index": "(X << 6) + Y",
            "source_payloads": [{"payload_index": 0, "role": "segment 0x0000-0x0FFF"}],
        })

        pixels = atlas(seg)
        side = ATLAS_COLUMNS * TILE_SIDE
        tiles_png = encode_rgba_png(pixels, bytes([1]) * len(pixels), side, side)
        verify_png_bytes(tiles_png, side, side)
        digest = hashlib.sha256(tiles_png).hexdigest()
        if digest != record["tiles"]["png_sha256"]:
            raise SystemExit(f"{root} 磚塊 atlas SHA-256 與 visual probe 不符")
        (outdir / record["tiles"]["file"]).write_bytes(tiles_png)
        images.append({
            "unit_index": 1, "file": f"{root}/{record['tiles']['file']}",
            "width": side, "height": side, "has_transparency": False,
            "png_sha256": digest, "role": "tile_atlas",
            "tile_count": TILE_COUNT, "tile_side": TILE_SIDE,
            "atlas_columns": ATLAS_COLUMNS,
            "source_payloads": [{"payload_index": i, "role": "segment tiles"}
                                for i in (0, 1, 2)],
        })

        out.append({
            "archive": root, "archive_root": f"Knights-of-legend/{root}",
            "source_payload_count": record["archive_payload_count"],
            "plan_record_count": 2, "payloads_without_plan_record": 0,
            "png_count": len(images), "skipped_count": 0, "skipped": [],
            "composition": "town_map", "evidence": provenance, "images": images,
        })
    return out


def extract_party_markers(files: dict[str, bytes], repo: pathlib.Path,
                          staging: pathlib.Path) -> dict:
    """KOLPICS payload 0 的十個隊伍標記精靈（RE-060）。"""
    from probe_extract_party_marker_png import (  # 延後載入避免互相 import
        ARCHIVE_ROOT, PAYLOAD_INDEX, UNIT_COUNT, UNIT_SIDE, build)

    probe_path = repo / "docs/re/party-marker-ega-visual-probe.json"
    probe = json.loads(probe_path.read_text(encoding="utf-8"))
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None:
        raise SystemExit("KOLPICS archive contract 不符")
    data = payload(files, contract["items"][PAYLOAD_INDEX])
    if hashlib.sha256(data).hexdigest() != probe["source"]["decoded_sha256"]:
        raise SystemExit("KOLPICS payload 0 解碼雜湊與 visual probe 不符")

    pixels, _ = build(data)
    width, height = UNIT_SIDE, UNIT_SIDE * UNIT_COUNT
    png = encode_rgba_png(pixels, bytes([1]) * len(pixels), width, height)
    verify_png_bytes(png, width, height)
    digest = hashlib.sha256(png).hexdigest()
    if digest != probe["atlas"]["png_sha256"]:
        raise SystemExit("隊伍標記 atlas SHA-256 與 visual probe 不符")

    name = "KOLPICS-MARKERS"
    outdir = staging / name
    outdir.mkdir(parents=True, exist_ok=True)
    (outdir / probe["atlas"]["file"]).write_bytes(png)
    return {
        "archive": name,
        "archive_root": ARCHIVE_ROOT,
        "source_payload_count": contract["count"],
        "plan_record_count": 1, "payloads_without_plan_record": 0,
        "png_count": 1, "skipped_count": 0, "skipped": [],
        "composition": "party_marker_atlas",
        "note": "來源是 KOLPICS 的 payload 0；這個 archive 名稱是 remake 的排列，"
                "不是原版另一個 archive。既有的 KOLPICS 抽取把同一段當 renderer 資料"
                "排成 256×5，每 256 byte 只取前半，看不到單元 1、3、5、7、9。",
        "evidence": {
            "visual_probe": probe_path.name,
            "visual_probe_sha256": hashlib.sha256(probe_path.read_bytes()).hexdigest(),
            "evidence": probe["evidence"], "status": probe["status"],
        },
        "images": [{
            "unit_index": 0, "file": f"{name}/{probe['atlas']['file']}",
            "width": width, "height": height, "has_transparency": False,
            "png_sha256": digest, "role": "party_marker_atlas",
            "unit_count": UNIT_COUNT, "unit_side": UNIT_SIDE,
            "source_payloads": [{"image_index": PAYLOAD_INDEX, "role": "party markers"}],
        }],
    }


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--repo", default=".")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    repo = pathlib.Path(args.repo).resolve()
    out = pathlib.Path(args.out).resolve()
    try:
        relative = out.relative_to(repo)
    except ValueError:
        raise SystemExit("輸出根目錄必須位於 repo 內")
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(relative.as_posix()).parents:
        raise SystemExit(f"輸出根目錄必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")

    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")
    with zipfile.ZipFile(zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}

    staging = out.parent / f".staging-{os.getpid()}"
    if staging.exists():
        shutil.rmtree(staging)
    staging.mkdir(parents=True)
    try:
        shared = {"kolmpics_28_mask": kolmpics_shared_mask(files)}
        archives = [extract_family(f, files, repo, staging, shared) for f in FAMILIES]
        archives.extend(extract_town_maps(files, repo, staging))
        archives.append(extract_party_markers(files, repo, staging))
        manifest = {
            "schema": "kol-ega-production-manifest-v2",
            "tool": "tools/extract_ega_assets.py",
            "source_zip": {"name": zip_path.name, "sha256": digest},
            "palette": "BIOS mode 0Dh EGA 16 色，索引即原版 4-bit 值",
            "png_format": "8-bit RGBA（colour type 6）",
            "plan_source": "docs/re/*-visual-probe.json 的逐張紀錄；每張 PNG 的 "
                           "SHA-256 都與該紀錄核對，上游 IDA 摘要雜湊記在各 archive 的 evidence",
            "rights": {
                "classification": "private-repo 可保存；不得公開散布",
                "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
            },
            "coverage_note": "涵蓋 SPEC-002 的全部 752 個已證實 PNG 單位，"
                             "外加 TOWN0–TOWN7 的地圖格子陣列與磚塊 atlas（RE-059）"
                             "與 KOLPICS payload 0 的隊伍標記 atlas（RE-060）。"
                             "地圖那張是資料的無損 8-bit 灰階 PNG 表示，像素值即磚塊編號，"
                             "不是原版顯示矩形。"
                             "KOLPICS 的 raw-planar 與 WEATPICS 的遮罩層是 renderer 資料，"
                             "不是原版顯示矩形。`skipped` 是計畫內明確排除的 placeholder；"
                             "`payloads_without_plan_record` 是連紀錄都沒有的 payload"
                             "（MISCPICS 的一-byte placeholder 屬此類）。",
            "archive_count": len(archives),
            "png_count": sum(a["png_count"] for a in archives),
            "archives": archives,
        }
        (staging / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
        if out.exists():
            # 這個根目錄還住著別支抽取器的產物（`extract_battle_tiles.py` 的
            # `BATTLE/`）。原子替換是整個目錄換掉，所以先把不在本次計畫裡的
            # 子目錄搬進 staging——否則跑一次這支就把別人的 PNG 全刪了。
            keep = {a["archive"] for a in archives}
            for entry in sorted(out.iterdir()):
                if entry.name in keep or entry.name == "manifest.json":
                    continue
                dest = staging / entry.name
                if dest.exists():
                    continue
                if entry.is_dir():
                    shutil.copytree(entry, dest)
                else:
                    shutil.copy2(entry, dest)
            shutil.rmtree(out)
        os.replace(staging, out)
    finally:
        if staging.exists():
            shutil.rmtree(staging, ignore_errors=True)

    print(json.dumps({"archives": manifest["archive_count"],
                      "pngs": manifest["png_count"], "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
