"""驗證 AWD1PICS 全 24 張 24×21 EGA 圖片。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_comb_png import decode_nibbles
from probe_extract_monpics_png import encode_rgba_png
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/AWD1PICS"
WIDTH, HEIGHT = 24, 21


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("ida_summary_path", type=Path)
    parser.add_argument("output_directory", type=Path)
    parser.add_argument("summary_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_directory.resolve().parts:
        raise SystemExit("研究 PNG 只可寫入 ignored workplace")
    zip_bytes = args.zip_path.read_bytes()
    if sha(zip_bytes) != EXPECTED_ZIP_SHA256:
        raise SystemExit("非預期 ZIP SHA-256")
    ida_bytes = args.ida_summary_path.read_bytes()
    ida = json.loads(ida_bytes)
    if ida.get("schema") != "kol-ida-koltown-awdpics-summary-v1":
        raise SystemExit("非預期 AWD1PICS IDA 摘要 schema")
    expected = ida.get("archive_contract", {})
    if (expected.get("image_indices") != list(range(24)) or
            (expected.get("width_pixels"), expected.get("height_pixels")) != (WIDTH, HEIGHT)):
        raise SystemExit("IDA 摘要尚未證實 AWD1PICS 全 archive 幾何")
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 24:
        raise SystemExit("AWD1PICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)
    records, decoded = [], []
    for item in contract["items"]:
        source = files[item["volume_name"]]
        stage1 = decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])
        pixels = decode_nibbles(stage1)
        if len(pixels) != WIDTH * HEIGHT or any(value > 15 for value in pixels):
            raise SystemExit(f"AWD1PICS[{item['index']}] 像素契約不符")
        decoded.append(pixels)
        output = args.output_directory / f"awd1pics-{item['index']:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))
        records.append({"image_index": item["index"], "width_pixels": WIDTH,
                        "height_pixels": HEIGHT, "color_count": len(set(pixels)),
                        "pixel_sha256": sha(pixels), "png_sha256": sha(output.read_bytes())})
    sheet_width, sheet_height = WIDTH * 8, HEIGHT * 3
    sheet = bytearray(sheet_width * sheet_height)
    for index, pixels in enumerate(decoded):
        tile_x, tile_y = index % 8, index // 8
        for y in range(HEIGHT):
            start = (tile_y * HEIGHT + y) * sheet_width + tile_x * WIDTH
            sheet[start:start + WIDTH] = pixels[y * WIDTH:(y + 1) * WIDTH]
    contact = args.output_directory / "awd1pics-contact-sheet.png"
    contact.write_bytes(encode_rgba_png(bytes(sheet), bytes([1]) * len(sheet), sheet_width, sheet_height))
    result = {"schema": "kol-awd1pics-ega-visual-probe-v1",
              "input_name": args.zip_path.name, "input_sha256": sha(zip_bytes),
              "ida_summary_name": args.ida_summary_path.name,
              "ida_summary_sha256": sha(ida_bytes),
              "status": "mechanically validated all 24 direct-blit 4-bit EGA images; orientation and sequence require human inspection",
              "human_inspection": "24 upright medal, trophy, shield, ring and other heraldic icons; no rotation, mirroring or plane-order defect observed",
              "semantic_classification": {"label": "character quest-medal icons",
                                          "confidence": "strong inference",
                                          "cross_reference": "Traditional Chinese manual page 41 and Apple II reference card Medal command"},
              "source_payload_count": 24, "png_count": 24,
              "width_pixels": WIDTH, "height_pixels": HEIGHT,
              "contact_sheet_layout": "8 columns by 3 rows in archive index order",
              "contact_sheet_sha256": sha(contact.read_bytes()), "records": records}
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print("validated=24")


if __name__ == "__main__":
    main()
