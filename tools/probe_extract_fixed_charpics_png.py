"""驗證三個 16×760-byte 固定角色圖片表，輸出本機研究 PNG。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256
from probe_extract_monpics_png import encode_rgba_png


SOURCE_NAMES = ("CHARPICS", "SHLDPICS", "STEEDPIC")
ARCHIVE_PREFIX = "Knights-of-legend/"
RECORD_COUNT = 16
RECORD_BYTES = 760
WIDTH = 40
HEIGHT = 38
GRID_COLUMNS = 8
CELL_WIDTH = 48
CELL_HEIGHT = 46


def decode_record(data: bytes) -> bytes:
    if len(data) != RECORD_BYTES:
        raise ValueError(f"固定角色圖片記錄長度不符：{len(data)}")
    pixels = bytearray()
    for value in data:
        pixels.extend((value >> 4, value & 0x0F))
    if len(pixels) != WIDTH * HEIGHT:
        raise AssertionError("固定角色圖片像素數不符")
    return bytes(pixels)


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
    ida_bytes = args.ida_summary_path.read_bytes()
    ida = json.loads(ida_bytes)
    contract = ida.get("record_contract", {})
    expected_contract = {
        "status": "confirmed",
        "record_bytes": RECORD_BYTES,
        "width_pixels": WIDTH,
        "height_pixels": HEIGHT,
        "bits_per_pixel": 4,
        "pixels_per_byte": 2,
        "packing": "high nibble first, then low nibble",
    }
    if any(contract.get(key) != value for key, value in expected_contract.items()):
        raise SystemExit("IDA 摘要尚未證實固定角色圖片的 40×38 packed-nibble 契約")

    with zipfile.ZipFile(args.zip_path) as archive:
        sources = {
            name: archive.read(ARCHIVE_PREFIX + name)
            for name in SOURCE_NAMES
        }
    if any(len(data) != RECORD_COUNT * RECORD_BYTES for data in sources.values()):
        raise SystemExit("固定角色圖片表不是 16×760 bytes")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 6 * CELL_HEIGHT
    sheet_pixels = bytearray(
        7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
        for y in range(sheet_height)
        for x in range(sheet_width)
    )
    records = []
    for source_index, name in enumerate(SOURCE_NAMES):
        data = sources[name]
        for record_index in range(RECORD_COUNT):
            start = record_index * RECORD_BYTES
            raw = data[start : start + RECORD_BYTES]
            pixels = decode_record(raw)
            output = args.output_directory / f"{name.lower()}-{record_index:02d}.png"
            output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))

            grid_index = source_index * RECORD_COUNT + record_index
            origin_x = (grid_index % GRID_COLUMNS) * CELL_WIDTH + 4
            origin_y = (grid_index // GRID_COLUMNS) * CELL_HEIGHT + 4
            for y in range(HEIGHT):
                source_row = y * WIDTH
                destination = (origin_y + y) * sheet_width + origin_x
                sheet_pixels[destination : destination + WIDTH] = pixels[source_row : source_row + WIDTH]
            records.append(
                {
                    "source_name": name,
                    "record_index": record_index,
                    "record_sha256": hashlib.sha256(raw).hexdigest(),
                    "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
                    "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
                    "color_count": len(set(pixels)),
                }
            )

    sheet = args.output_directory / "fixed-charpics-contact-sheet.png"
    sheet.write_bytes(
        encode_rgba_png(
            bytes(sheet_pixels), bytes([1]) * len(sheet_pixels), sheet_width, sheet_height
        )
    )
    result = {
        "schema": "kol-fixed-charpics-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_bytes).hexdigest(),
        "status": "mechanically validated; orientation and semantic identities require human inspection",
        "source_names": list(SOURCE_NAMES),
        "source_sha256": {name: hashlib.sha256(data).hexdigest() for name, data in sources.items()},
        "record_count_per_source": RECORD_COUNT,
        "record_bytes": RECORD_BYTES,
        "width_pixels": WIDTH,
        "height_pixels": HEIGHT,
        "packing": "high nibble first, then low nibble",
        "png_count": len(records),
        "contact_sheet_sha256": hashlib.sha256(sheet.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={len(records)} contact_sheet={sheet}")


if __name__ == "__main__":
    main()
