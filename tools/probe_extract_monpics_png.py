"""驗證 MONPICS EGA 像素，輸出本機研究 PNG 與可版控統計。"""

from __future__ import annotations

import argparse
import hashlib
import json
import struct
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_comb_png import PALETTE, chunk, decode_nibbles
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/MONPICS"
BACKGROUND = 0
CELL_WIDTH = 96
CELL_HEIGHT = 80
GRID_COLUMNS = 8


def encode_rgba_png(indices: bytes, opaque: bytes, width: int, height: int) -> bytes:
    rows = []
    for y in range(height):
        colors = indices[y * width : (y + 1) * width]
        alphas = opaque[y * width : (y + 1) * width]
        row = bytearray(b"\0")
        for index, alpha in zip(colors, alphas):
            row.extend((*PALETTE[index], 0xFF if alpha else 0x00))
        rows.append(bytes(row))
    header = struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0)
    import zlib

    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(
        b"IDAT", zlib.compress(b"".join(rows), 9)
    ) + chunk(b"IEND", b"")


def unpack_mask(data: bytes, width_bytes: int, height: int) -> tuple[bytes, bytes]:
    expected = width_bytes * height
    if len(data) != expected + 1:
        raise ValueError(f"遮罩長度不符：{len(data)} != {expected + 1}")
    pixels = bytearray()
    for value in data[:expected]:
        pixels.extend((value >> bit) & 1 for bit in range(7, -1, -1))
    return bytes(pixels), data[expected:]


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
    dimensions = ida_summary["dimension_table"]["dimensions"]
    if len(dimensions) != 32:
        raise SystemExit("非預期 MONPICS 尺寸表長度")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 64:
        raise SystemExit("非預期 MONPICS archive contract")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width = GRID_COLUMNS * CELL_WIDTH
    sheet_height = 4 * CELL_HEIGHT
    sheet_indices = bytearray([BACKGROUND]) * (sheet_width * sheet_height)
    sheet_opaque = bytearray([0]) * (sheet_width * sheet_height)
    records = []
    for monster_index in range(32):
        dimension = dimensions[monster_index]
        width = dimension["width_pixels"]
        height = dimension["height_pixels"]
        width_bytes = dimension["width_bytes"]
        color_item = contract["items"][monster_index * 2]
        mask_item = contract["items"][monster_index * 2 + 1]
        color_source = files[color_item["volume_name"]]
        mask_source = files[mask_item["volume_name"]]
        color_first_stage = decode(color_source[
            color_item["offset"] : color_item["offset"] + color_item["compressed_bytes"]
        ])
        mask_first_stage = decode(mask_source[
            mask_item["offset"] : mask_item["offset"] + mask_item["compressed_bytes"]
        ])
        pixels = decode_nibbles(color_first_stage)
        opaque, mask_trailing = unpack_mask(mask_first_stage, width_bytes, height)
        expected_pixels = width * height
        if len(pixels) != expected_pixels:
            raise SystemExit(
                f"monster {monster_index} 像素數不符：{len(pixels)} != {expected_pixels}"
            )
        output_path = args.output_directory / f"monster-{monster_index:02d}.png"
        output_path.write_bytes(encode_rgba_png(pixels, opaque, width, height))

        grid_index = monster_index
        origin_x = (grid_index % GRID_COLUMNS) * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (grid_index // GRID_COLUMNS) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
        for y in range(height):
            destination = (origin_y + y) * sheet_width + origin_x
            source_row = y * width
            sheet_indices[destination : destination + width] = pixels[source_row : source_row + width]
            sheet_opaque[destination : destination + width] = opaque[source_row : source_row + width]
        records.append({
            "monster_index": monster_index,
            "color_image_index": monster_index * 2,
            "mask_image_index": monster_index * 2 + 1,
            "width_pixels": width,
            "height_pixels": height,
            "pixel_count": len(pixels),
            "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
            "mask_sha256": hashlib.sha256(opaque).hexdigest(),
            "mask_trailing_byte_sha256": hashlib.sha256(mask_trailing).hexdigest(),
            "color_count": len(set(pixels)),
            "png_sha256": hashlib.sha256(output_path.read_bytes()).hexdigest(),
        })

    sheet_path = args.output_directory / "monpics-contact-sheet.png"
    sheet_path.write_bytes(
        encode_rgba_png(sheet_indices, sheet_opaque, sheet_width, sheet_height)
    )
    result = {
        "schema": "kol-monpics-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(args.ida_summary_path.read_bytes()).hexdigest(),
        "status": "mechanically validated EGA color and 1-bit mask candidates; mask semantics require human inspection",
        "archive_root": ARCHIVE_ROOT,
        "monster_count": 32,
        "source_payload_count": 64,
        "combined_png_count": 32,
        "all_pixel_counts_match": True,
        "contact_sheet_sha256": hashlib.sha256(sheet_path.read_bytes()).hexdigest(),
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=64 contact_sheet={sheet_path}")


if __name__ == "__main__":
    main()
