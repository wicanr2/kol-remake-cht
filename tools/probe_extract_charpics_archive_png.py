"""驗證 CHARPICS.0 全 80 張 40×38 EGA 角色圖片。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_fixed_charpics_png import HEIGHT, RECORD_BYTES, WIDTH, decode_record
from probe_extract_monpics_png import encode_rgba_png
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/CHARPICS"
IMAGE_COUNT = 80
GRID_COLUMNS = 10
CELL_WIDTH = 48
CELL_HEIGHT = 46


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("archive_ida_summary_path", type=Path)
    parser.add_argument("record_ida_summary_path", type=Path)
    parser.add_argument("output_directory", type=Path)
    parser.add_argument("summary_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_directory.resolve().parts:
        raise SystemExit("研究 PNG 只可寫入 ignored workplace")

    zip_bytes = args.zip_path.read_bytes()
    zip_digest = hashlib.sha256(zip_bytes).hexdigest()
    if zip_digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{zip_digest}")
    archive_ida_bytes = args.archive_ida_summary_path.read_bytes()
    archive_ida = json.loads(archive_ida_bytes)
    archive_contract = archive_ida.get("archive_consumer_contract", {})
    if (
        archive_contract.get("status") != "confirmed"
        or archive_contract.get("indexed_archive_name") != "CHARPICS.0"
        or archive_contract.get("index_entry_offset_formula") != "20 + 2 * image_index"
        or archive_contract.get("subsequent_display_original_function") != "sub_15335"
    ):
        raise SystemExit("IDA 摘要尚未證實 CHARPICS.0 archive consumer")
    record_ida_bytes = args.record_ida_summary_path.read_bytes()
    record_ida = json.loads(record_ida_bytes)
    record_contract = record_ida.get("record_contract", {})
    if (
        record_contract.get("status") != "confirmed"
        or record_contract.get("record_bytes") != RECORD_BYTES
        or record_contract.get("width_pixels") != WIDTH
        or record_contract.get("height_pixels") != HEIGHT
        or record_contract.get("packing") != "high nibble first, then low nibble"
    ):
        raise SystemExit("IDA 摘要尚未證實 40×38 packed-nibble 下游 consumer")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != IMAGE_COUNT:
        raise SystemExit("CHARPICS.0 archive contract 不符")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 8 * CELL_HEIGHT
    sheet_pixels = bytearray(
        7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
        for y in range(sheet_height)
        for x in range(sheet_width)
    )
    records = []
    for item in contract["items"]:
        source = files[item["volume_name"]]
        compressed = source[item["offset"] : item["offset"] + item["compressed_bytes"]]
        stage1 = decode(compressed)
        if len(stage1) != RECORD_BYTES + 1:
            raise SystemExit(
                f"CHARPICS[{item['index']}] 解碼長度不符：{len(stage1)} != {RECORD_BYTES + 1}"
            )
        pixels = decode_record(stage1[:RECORD_BYTES])
        output = args.output_directory / f"charpics-{item['index']:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))

        origin_x = (item["index"] % GRID_COLUMNS) * CELL_WIDTH + 4
        origin_y = (item["index"] // GRID_COLUMNS) * CELL_HEIGHT + 4
        for y in range(HEIGHT):
            source_row = y * WIDTH
            destination = (origin_y + y) * sheet_width + origin_x
            sheet_pixels[destination : destination + WIDTH] = pixels[source_row : source_row + WIDTH]
        records.append(
            {
                "image_index": item["index"],
                "compressed_bytes": item["compressed_bytes"],
                "decoded_bytes": len(stage1),
                "pixel_bytes_consumed": RECORD_BYTES,
                "trailing_byte_hex": stage1[RECORD_BYTES:].hex(),
                "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
                "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
                "color_count": len(set(pixels)),
            }
        )

    sheet = args.output_directory / "charpics-contact-sheet.png"
    sheet.write_bytes(
        encode_rgba_png(
            bytes(sheet_pixels), bytes([1]) * len(sheet_pixels), sheet_width, sheet_height
        )
    )
    result = {
        "schema": "kol-charpics-archive-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "archive_ida_summary_name": args.archive_ida_summary_path.name,
        "archive_ida_summary_sha256": hashlib.sha256(archive_ida_bytes).hexdigest(),
        "record_ida_summary_name": args.record_ida_summary_path.name,
        "record_ida_summary_sha256": hashlib.sha256(record_ida_bytes).hexdigest(),
        "status": "mechanically validated all 80 images; orientation and semantic groups require human inspection",
        "source_archive": "CHARPICS.0",
        "image_count": len(records),
        "width_pixels": WIDTH,
        "height_pixels": HEIGHT,
        "packing": "high nibble first, then low nibble",
        "decoded_bytes_per_item": RECORD_BYTES + 1,
        "pixel_bytes_per_item": RECORD_BYTES,
        "trailing_byte_disposition": "excluded_non_pixel; downstream 40x38 consumer reads exactly 760 bytes",
        "png_count": len(records),
        "contact_sheet_sha256": hashlib.sha256(sheet.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={len(records)} contact_sheet={sheet}")


if __name__ == "__main__":
    main()
