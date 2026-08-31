"""驗證 MISCPICS 五張 EGA 彩圖與未引用 placeholder。"""

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


ARCHIVE_ROOT = "Knights-of-legend/MISCPICS"
PLACEHOLDER_INDEX = 2
CELL_WIDTH = 304
CELL_HEIGHT = 104
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
    if ida_summary.get("schema") != "kol-ida-koltown-miscpics-summary-v1":
        raise SystemExit("非預期 MISCPICS IDA 摘要 schema")
    routes = ida_summary["router"]["routes"]
    if [route["image_index"] for route in routes] != [0, 1, 3, 4, 5]:
        raise SystemExit("非預期 MISCPICS route index set")
    if ida_summary["router"]["unreferenced_placeholder_image_index"] != PLACEHOLDER_INDEX:
        raise SystemExit("非預期 MISCPICS placeholder index")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 6:
        raise SystemExit("非預期 MISCPICS archive contract")

    placeholder_item = contract["items"][PLACEHOLDER_INDEX]
    placeholder_source = files[placeholder_item["volume_name"]]
    placeholder_decoded = decode(placeholder_source[
        placeholder_item["offset"] : placeholder_item["offset"] + placeholder_item["compressed_bytes"]
    ])
    if len(placeholder_decoded) != 1:
        raise SystemExit("MISCPICS placeholder 解碼長度不是 1")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 3 * CELL_HEIGHT
    sheet_indices = bytearray(sheet_width * sheet_height)
    sheet_opaque = bytearray(sheet_width * sheet_height)
    records = []
    for grid_index, route in enumerate(routes):
        item = contract["items"][route["image_index"]]
        source = files[item["volume_name"]]
        first_stage = decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])
        pixels = decode_nibbles(first_stage)
        width = route["width_pixels"]
        height = route["height_pixels"]
        expected_pixels = width * height
        if len(pixels) != expected_pixels:
            raise SystemExit(
                f"MISCPICS index {item['index']} 像素數不符：{len(pixels)} != {expected_pixels}"
            )
        if any(index > 15 for index in pixels):
            raise SystemExit(f"MISCPICS index {item['index']} EGA 色彩索引越界")
        opaque = bytes([1]) * len(pixels)
        output_path = args.output_directory / f"miscpics-index-{item['index']:04d}.png"
        output_path.write_bytes(encode_rgba_png(pixels, opaque, width, height))
        origin_x = (grid_index % GRID_COLUMNS) * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (grid_index // GRID_COLUMNS) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
        for y in range(height):
            destination = (origin_y + y) * sheet_width + origin_x
            source_row = y * width
            sheet_indices[destination : destination + width] = pixels[source_row : source_row + width]
            sheet_opaque[destination : destination + width] = opaque[source_row : source_row + width]
        records.append({
            "request_code": route["request_code"],
            "image_index": item["index"],
            "width_pixels": width,
            "height_pixels": height,
            "pixel_count": len(pixels),
            "color_count": len(set(pixels)),
            "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
            "png_sha256": hashlib.sha256(output_path.read_bytes()).hexdigest(),
        })

    sheet_path = args.output_directory / "miscpics-contact-sheet.png"
    sheet_path.write_bytes(encode_rgba_png(
        sheet_indices, sheet_opaque, sheet_width, sheet_height
    ))
    result = {
        "schema": "kol-miscpics-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_summary_bytes).hexdigest(),
        "status": "mechanically validated five 4-bit EGA candidates; orientation requires human inspection",
        "archive_root": ARCHIVE_ROOT,
        "source_payload_count": 6,
        "png_count": 5,
        "placeholder": {
            "image_index": PLACEHOLDER_INDEX,
            "disposition": "excluded_unreferenced_placeholder",
            "decoded_bytes": len(placeholder_decoded),
            "decoded_sha256": hashlib.sha256(placeholder_decoded).hexdigest(),
            "output_png": None,
        },
        "all_pixel_counts_match": True,
        "contact_sheet_sha256": hashlib.sha256(sheet_path.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=5 placeholder=1 contact_sheet={sheet_path}")


if __name__ == "__main__":
    main()
