#!/usr/bin/env python3
"""從固定原版輸入抽出兩套內嵌點陣字型，輸出 runtime 用的 PNG 字模表與度量。

`BOOKDATA.0` 與 `KOLDATA.0` 的第一個 payload 解壓後是同一份 1,303-byte 區塊
（SHA-256 相同），內容依序是：

| 位移 | 內容 |
|---|---|
| `0x000`–`0x2e7` | 大字型，93 個字模、每模 8 bytes、8 列 |
| `0x2e8`–`0x515` | 小字型，93 個字模、每模 6 bytes、6 列 |
| `0x516` | 一個尾端 byte，不屬任何字模 |

兩套都由 ASCII 32 起算，逐 byte 是一列點陣、MSB 在左。小字型就是書本內文所用者：
由原版畫面取出的 `HUMAN` 五個字模在此逐 byte 命中。

輸出 RGBA 字模表：墨色為不透明白，其餘透明，runtime 自行上色。

  python3 tools/extract_fonts.py --zip kol.zip --repo . --out assets/original-derived-png/fonts
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
from probe_image_lzw import decode as lzw_decode  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-png")
SOURCE_ROOTS = ("Knights-of-legend/BOOKDATA", "Knights-of-legend/KOLDATA")
PAYLOAD_BYTES = 1303
FIRST_CODEPOINT = 32
GLYPH_COUNT = 93
SHEET_COLUMNS = 16

FONTS = [
    {"name": "book-large", "offset": 0x000, "bytes_per_glyph": 8, "width": 8, "height": 8,
     "note": "大字型；本輪尚未指認其在畫面上的使用位置"},
    {"name": "book-small", "offset": 0x2E8, "bytes_per_glyph": 6, "width": 8, "height": 6,
     "note": "書本內文字型；HUMAN 五字模已由原版畫面逐 byte 命中"},
]


def chunk(tag: bytes, body: bytes) -> bytes:
    payload = tag + body
    return (struct.pack(">I", len(body)) + payload
            + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))


def encode_rgba(w: int, h: int, ink: list[int]) -> bytes:
    rows = []
    for y in range(h):
        row = bytearray(b"\0")
        for x in range(w):
            on = ink[y * w + x]
            row += bytes((0xFF, 0xFF, 0xFF, 0xFF) if on else (0x00, 0x00, 0x00, 0x00))
        rows.append(bytes(row))
    return (b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 6, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(b"".join(rows), 9))
            + chunk(b"IEND", b""))


def payload_zero(files: dict[str, bytes], root: str) -> bytes:
    contract = logical_archive_contract(root, files)
    if contract is None or contract["count"] < 1:
        raise SystemExit(f"{root} archive contract 不符")
    item = contract["items"][0]
    src = files[item["volume_name"]]
    return lzw_decode(src[item["offset"]:item["offset"] + item["compressed_bytes"]])


def build_sheet(data: bytes, font: dict) -> tuple[bytes, list[dict], int, int]:
    cw, ch = font["width"], font["height"]
    rows = (GLYPH_COUNT + SHEET_COLUMNS - 1) // SHEET_COLUMNS
    sw, sh = SHEET_COLUMNS * cw, rows * ch
    ink = [0] * (sw * sh)
    glyphs = []
    for i in range(GLYPH_COUNT):
        off = font["offset"] + i * font["bytes_per_glyph"]
        raw = data[off:off + font["bytes_per_glyph"]]
        if len(raw) != font["bytes_per_glyph"]:
            raise SystemExit(f"{font['name']} 字模 {i} 超出 payload")
        gx, gy = (i % SHEET_COLUMNS) * cw, (i // SHEET_COLUMNS) * ch
        for y, value in enumerate(raw):
            for x in range(cw):
                if (value >> (7 - x)) & 1:
                    ink[(gy + y) * sw + gx + x] = 1
        glyphs.append({
            "codepoint": FIRST_CODEPOINT + i,
            "sheet_column": i % SHEET_COLUMNS,
            "sheet_row": i // SHEET_COLUMNS,
            "source_offset": f"0x{off:04x}",
            "rows_hex": raw.hex(),
        })
    return encode_rgba(sw, sh, ink), glyphs, sw, sh


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

    payloads = {root: payload_zero(files, root) for root in SOURCE_ROOTS}
    data = payloads[SOURCE_ROOTS[0]]
    if len(data) != PAYLOAD_BYTES:
        raise SystemExit(f"payload 長度不符：{len(data)}")
    if any(v != data for v in payloads.values()):
        raise SystemExit("BOOKDATA 與 KOLDATA 的第一個 payload 不相同，字型來源假設不成立")
    payload_sha = hashlib.sha256(data).hexdigest()

    staging = out.parent / f".staging-fonts-{os.getpid()}"
    if staging.exists():
        shutil.rmtree(staging)
    staging.mkdir(parents=True)
    try:
        entries = []
        for font in FONTS:
            png, glyphs, sw, sh = build_sheet(data, font)
            name = f"{font['name']}-{font['width']}x{font['height']}.png"
            (staging / name).write_bytes(png)
            entries.append({
                "name": font["name"],
                "file": name,
                "note": font["note"],
                "cell": {"width": font["width"], "height": font["height"]},
                "advance": font["width"],
                "bytes_per_glyph": font["bytes_per_glyph"],
                "source_offset": f"0x{font['offset']:04x}",
                "first_codepoint": FIRST_CODEPOINT,
                "glyph_count": GLYPH_COUNT,
                "sheet": {"columns": SHEET_COLUMNS, "width": sw, "height": sh},
                "png_sha256": hashlib.sha256(png).hexdigest(),
                "glyphs": glyphs,
            })
        used = FONTS[-1]["offset"] + GLYPH_COUNT * FONTS[-1]["bytes_per_glyph"]
        manifest = {
            "schema": "kol-font-manifest-v1",
            "tool": "tools/extract_fonts.py",
            "source_zip": {"name": zip_path.name, "sha256": digest},
            "source_payload": {
                "archives": [r.split("/")[-1] for r in SOURCE_ROOTS],
                "payload_index": 0,
                "bytes": len(data),
                "sha256": payload_sha,
                "identical_across_archives": True,
                "trailing_bytes": len(data) - used,
            },
            "encoding": "每 byte 一列點陣，MSB 在左；墨色輸出為不透明白，其餘透明",
            "rights": {
                "classification": "private-repo 可保存；不得公開散布",
                "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
            },
            "fonts": entries,
        }
        (staging / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
        if out.exists():
            shutil.rmtree(out)
        os.replace(staging, out)
    finally:
        if staging.exists():
            shutil.rmtree(staging, ignore_errors=True)

    print(json.dumps({"fonts": len(entries), "glyphs": GLYPH_COUNT,
                      "payload_sha256": payload_sha, "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
