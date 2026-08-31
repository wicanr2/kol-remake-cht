"""驗證 COMBPICS 全七張 EGA 彩圖，輸出本機 PNG 與安全統計。"""

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


ARCHIVE_ROOT = "Knights-of-legend/COMBPICS"
CELL_WIDTH = 304
CELL_HEIGHT = 152
GRID_COLUMNS = 2


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
    zip_digest = hashlib.sha256(zip_bytes).hexdigest()
    if zip_digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{zip_digest}")
    ida_summary_bytes = args.ida_summary_path.read_bytes()
    ida_summary = json.loads(ida_summary_bytes)
    if ida_summary.get("schema") != "kol-ida-koltown-combpics-summary-v1":
        raise SystemExit("非預期 COMBPICS IDA 摘要 schema")
    groups = ida_summary["consumer_groups"]
    dimensions = {}
    for group in groups:
        rectangle = group["rectangle"]
        for image_index in group["image_indices"]:
            if image_index in dimensions:
                raise SystemExit(f"COMBPICS 圖號 {image_index} 重複 consumer")
            dimensions[image_index] = (
                rectangle["width_pixels"], rectangle["height_pixels"], group["original_name"]
            )
    if sorted(dimensions) != list(range(7)):
        raise SystemExit("COMBPICS consumer 未完整覆蓋圖號 0..6")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 7:
        raise SystemExit("非預期 COMBPICS archive contract")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 4 * CELL_HEIGHT
    sheet_indices = bytearray(sheet_width * sheet_height)
    sheet_opaque = bytearray(sheet_width * sheet_height)
    records = []
    for image_index, item in enumerate(contract["items"]):
        source = files[item["volume_name"]]
        first_stage = decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])
        pixels = decode_nibbles(first_stage)
        width, height, consumer_name = dimensions[image_index]
        expected_pixels = width * height
        if len(pixels) != expected_pixels:
            raise SystemExit(
                f"COMBPICS index {image_index} 像素數不符：{len(pixels)} != {expected_pixels}"
            )
        if any(index > 15 for index in pixels):
            raise SystemExit(f"COMBPICS index {image_index} EGA 色彩索引越界")
        opaque = bytes([1]) * len(pixels)
        output_path = args.output_directory / f"combpics-index-{image_index:04d}.png"
        output_path.write_bytes(encode_rgba_png(pixels, opaque, width, height))
        origin_x = (image_index % GRID_COLUMNS) * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (image_index // GRID_COLUMNS) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
        for y in range(height):
            destination = (origin_y + y) * sheet_width + origin_x
            source_row = y * width
            sheet_indices[destination : destination + width] = pixels[source_row : source_row + width]
            sheet_opaque[destination : destination + width] = opaque[source_row : source_row + width]
        records.append({
            "image_index": image_index,
            "consumer_original_name": consumer_name,
            "width_pixels": width,
            "height_pixels": height,
            "pixel_count": len(pixels),
            "color_count": len(set(pixels)),
            "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
            "png_sha256": hashlib.sha256(output_path.read_bytes()).hexdigest(),
        })

    sheet_path = args.output_directory / "combpics-all-contact-sheet.png"
    sheet_path.write_bytes(encode_rgba_png(
        sheet_indices, sheet_opaque, sheet_width, sheet_height
    ))
    result = {
        "schema": "kol-combpics-all-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_summary_bytes).hexdigest(),
        "status": "mechanically validated all seven 4-bit EGA images; orientation requires human inspection",
        "archive_root": ARCHIVE_ROOT,
        "source_payload_count": 7,
        "png_count": 7,
        "all_archive_indices_have_consumer": True,
        "all_pixel_counts_match": True,
        "contact_sheet_sha256": hashlib.sha256(sheet_path.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=7 contact_sheet={sheet_path}")


if __name__ == "__main__":
    main()
