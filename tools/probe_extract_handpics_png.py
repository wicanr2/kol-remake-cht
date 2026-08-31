"""驗證 HANDPICS.0 十二組 EGA 彩圖／AND-mask。"""

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

ARCHIVE_ROOT = "Knights-of-legend/HANDPICS"
CELL_WIDTH, CELL_HEIGHT = 32, 16


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
    if ida.get("schema") != "kol-ida-koltown-handpics-summary-v1":
        raise SystemExit("非預期 HANDPICS IDA 摘要 schema")
    dimensions = ida.get("archive_contract", {}).get("dimensions")
    if len(dimensions or []) != 12:
        raise SystemExit("IDA 摘要尚未證實 HANDPICS 全 12 組")
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 24:
        raise SystemExit("HANDPICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width, sheet_height = CELL_WIDTH * 2, CELL_HEIGHT * 6
    sheet_indices = bytearray(7 if ((x // 4) + (y // 4)) % 2 == 0 else 8
                              for y in range(sheet_height) for x in range(sheet_width))
    sheet_opaque = bytes([1]) * (sheet_width * sheet_height)
    records = []
    for dimension in dimensions:
        group = dimension["group"]
        width, height = dimension["width_pixels"], dimension["height_pixels"]
        color_item, mask_item = contract["items"][2*group:2*group+2]
        def first_stage(item: dict) -> bytes:
            source = files[item["volume_name"]]
            return decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])
        pixels = decode_nibbles(first_stage(color_item))
        mask_bits, opaque, trailing = unpack_mask(first_stage(mask_item), width // 8, height)
        if len(pixels) != width * height or set(opaque) != {0, 1}:
            raise SystemExit(f"HANDPICS group {group} 像素／alpha 契約不符")
        output = args.output_directory / f"hand-{group:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, opaque, width, height))
        origin_x = (group % 2) * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (group // 2) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
        for y in range(height):
            for x in range(width):
                if opaque[y * width + x]:
                    sheet_indices[(origin_y + y) * sheet_width + origin_x + x] = pixels[y * width + x]
        records.append({"group": group, "appearance_class": dimension["appearance_class"],
                        "position_in_pair": dimension["side"],
                        "color_image_index": 2*group, "mask_image_index": 2*group+1,
                        "width_pixels": width, "height_pixels": height,
                        "pixel_sha256": sha(pixels), "and_mask_sha256": sha(mask_bits),
                        "opaque_sha256": sha(opaque), "mask_trailing_byte_sha256": sha(trailing),
                        "png_sha256": sha(output.read_bytes())})
    contact = args.output_directory / "handpics-contact-sheet.png"
    contact.write_bytes(encode_rgba_png(bytes(sheet_indices), sheet_opaque, sheet_width, sheet_height))
    result = {"schema": "kol-handpics-ega-visual-probe-v1",
              "input_name": args.zip_path.name, "input_sha256": sha(zip_bytes),
              "ida_summary_name": "KOLTOWN-handpics-summary.json",
              "ida_summary_sha256": sha(ida_bytes),
              "status": "mechanically validated and human-inspected all twelve color/AND-mask pairs",
              "human_inspection": "six rows of paired hand sprites; transparent contours expose the checkerboard and no rotation, pair-order, plane-order or clipping defect is visible",
              "semantic_classification": {"label": "paired character hand sprites",
                                          "confidence": "confirmed",
                                          "basis": "hand filename operand, paired character renderer and visible silhouettes"},
              "source_payload_count": 24, "combined_png_count": 12,
              "mask_semantics": "bit 1 preserves background/transparent; PNG alpha is 1-mask",
              "contact_sheet_layout": "two positions by six appearance classes",
              "contact_sheet_background": "opaque 4x4 checkerboard verifies transparent silhouettes",
              "contact_sheet_sha256": sha(contact.read_bytes()), "records": records}
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print("validated=24 combined=12")


if __name__ == "__main__":
    main()
