"""驗證 PERPICS EGA 像素，輸出本機研究 PNG 與可版控統計。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_comb_png import decode_nibbles
from probe_extract_monpics_png import encode_rgba_png, unpack_mask
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/PERPICS"
CELL_WIDTH = 120
CELL_HEIGHT = 160
GRID_COLUMNS = 3


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
    ida_summary = json.loads(args.ida_summary_path.read_text(encoding="utf-8"))
    if ida_summary.get("schema") != "kol-ida-koltown-perpics-summary-v1":
        raise SystemExit("非預期 PERPICS IDA 摘要 schema")
    dimensions = ida_summary["dimensions"]
    if len(dimensions) != 6:
        raise SystemExit("非預期 PERPICS 尺寸表長度")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 12:
        raise SystemExit("非預期 PERPICS archive contract")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 2 * CELL_HEIGHT
    sheet_indices = bytearray(sheet_width * sheet_height)
    sheet_opaque = bytearray(sheet_width * sheet_height)
    records = []
    for dimension in dimensions:
        portrait_index = dimension["portrait_index"]
        width = dimension["width_pixels"]
        height = dimension["height_pixels"]
        width_bytes = dimension["width_bytes"]
        color_item = contract["items"][dimension["color_image_index"]]
        mask_item = contract["items"][dimension["mask_image_index"]]

        def first_stage(item: dict) -> bytes:
            source = files[item["volume_name"]]
            return decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])

        pixels = decode_nibbles(first_stage(color_item))
        opaque, mask_trailing = unpack_mask(first_stage(mask_item), width_bytes, height)
        expected_pixels = width * height
        if len(pixels) != expected_pixels or len(opaque) != expected_pixels:
            raise SystemExit(f"portrait {portrait_index} 像素數不符")
        if any(index > 15 for index in pixels):
            raise SystemExit(f"portrait {portrait_index} EGA 色彩索引越界")

        output_path = args.output_directory / f"portrait-{portrait_index:02d}.png"
        output_path.write_bytes(encode_rgba_png(pixels, opaque, width, height))
        origin_x = (portrait_index % GRID_COLUMNS) * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (portrait_index // GRID_COLUMNS) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
        for y in range(height):
            destination = (origin_y + y) * sheet_width + origin_x
            source_row = y * width
            sheet_indices[destination : destination + width] = pixels[source_row : source_row + width]
            sheet_opaque[destination : destination + width] = opaque[source_row : source_row + width]
        records.append({
            "portrait_index": portrait_index,
            "color_image_index": color_item["index"],
            "mask_image_index": mask_item["index"],
            "width_pixels": width,
            "height_pixels": height,
            "pixel_count": len(pixels),
            "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
            "mask_sha256": hashlib.sha256(opaque).hexdigest(),
            "mask_trailing_byte_sha256": hashlib.sha256(mask_trailing).hexdigest(),
            "color_count": len(set(pixels)),
            "png_sha256": hashlib.sha256(output_path.read_bytes()).hexdigest(),
        })

    sheet_path = args.output_directory / "perpics-contact-sheet.png"
    sheet_path.write_bytes(encode_rgba_png(
        sheet_indices, sheet_opaque, sheet_width, sheet_height
    ))
    result = {
        "schema": "kol-perpics-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(args.ida_summary_path.read_bytes()).hexdigest(),
        "status": "mechanically validated EGA color and 1-bit mask candidates; orientation and mask semantics require human inspection",
        "archive_root": ARCHIVE_ROOT,
        "portrait_count": 6,
        "source_payload_count": 12,
        "combined_png_count": 6,
        "all_pixel_counts_match": True,
        "contact_sheet_sha256": hashlib.sha256(sheet_path.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=12 contact_sheet={sheet_path}")


if __name__ == "__main__":
    main()
