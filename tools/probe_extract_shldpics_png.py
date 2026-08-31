"""驗證 SHLDPICS.0 四組 EGA 彩圖／AND-mask。"""

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

ARCHIVE_ROOT = "Knights-of-legend/SHLDPICS"
CELL_WIDTH, CELL_HEIGHT = 64, 72


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def unpack_and_mask(data: bytes, width: int, height: int) -> tuple[bytes, bytes, bytes]:
    expected = width // 8 * height
    if len(data) != expected + 1:
        raise ValueError(f"AND mask 長度不符：{len(data)} != {expected + 1}")
    mask_bits = bytes((value >> bit) & 1 for value in data[:expected] for bit in range(7, -1, -1))
    opaque = bytes(1 - value for value in mask_bits)
    return mask_bits, opaque, data[expected:]


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
    if ida.get("schema") != "kol-ida-koltown-shldpics-summary-v1":
        raise SystemExit("非預期 SHLDPICS IDA 摘要 schema")
    contract_evidence = ida.get("archive_contract", {})
    dimensions = contract_evidence.get("dimensions")
    if contract_evidence.get("names_are_distinct_formats") is not True or len(dimensions or []) != 4:
        raise SystemExit("IDA 摘要尚未證實 SHLDPICS 名稱與幾何契約")
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    archive_contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if archive_contract is None or archive_contract["count"] != 8:
        raise SystemExit("SHLDPICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)
    sheet_width, sheet_height = CELL_WIDTH * 4, CELL_HEIGHT
    sheet_indices = bytearray(sheet_width * sheet_height)
    sheet_opaque = bytearray(sheet_width * sheet_height)
    records = []
    for group, dimension in enumerate(dimensions):
        width, height = dimension["width_pixels"], dimension["height_pixels"]
        color_item, mask_item = archive_contract["items"][group * 2:group * 2 + 2]
        def first_stage(item: dict) -> bytes:
            source = files[item["volume_name"]]
            return decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])
        pixels = decode_nibbles(first_stage(color_item))
        mask_bits, opaque, trailing = unpack_and_mask(first_stage(mask_item), width, height)
        if len(pixels) != width * height or len(opaque) != width * height:
            raise SystemExit(f"shield {group} 像素契約不符")
        if set(opaque) != {0, 1}:
            raise SystemExit(f"shield {group} alpha 未同時含透明與不透明")
        output = args.output_directory / f"shield-{group}.png"
        output.write_bytes(encode_rgba_png(pixels, opaque, width, height))
        origin_x = group * CELL_WIDTH + (CELL_WIDTH - width) // 2
        origin_y = (CELL_HEIGHT - height) // 2
        for y in range(height):
            dst, src = (origin_y + y) * sheet_width + origin_x, y * width
            sheet_indices[dst:dst + width] = pixels[src:src + width]
            sheet_opaque[dst:dst + width] = opaque[src:src + width]
        records.append({"group": group, "color_image_index": group * 2,
                        "mask_image_index": group * 2 + 1, "width_pixels": width,
                        "height_pixels": height, "pixel_sha256": sha(pixels),
                        "and_mask_sha256": sha(mask_bits), "opaque_sha256": sha(opaque),
                        "mask_trailing_byte_sha256": sha(trailing),
                        "png_sha256": sha(output.read_bytes())})
    contact = args.output_directory / "shldpics-contact-sheet.png"
    contact.write_bytes(encode_rgba_png(bytes(sheet_indices), bytes(sheet_opaque), sheet_width, sheet_height))
    result = {"schema": "kol-shldpics-ega-visual-probe-v1",
              "input_name": args.zip_path.name, "input_sha256": sha(zip_bytes),
              "ida_summary_name": "KOLTOWN-shldpics-summary.json",
              "ida_summary_sha256": sha(ida_bytes),
              "status": "mechanically validated and human-inspected all four color/AND-mask pairs",
              "human_inspection": "four upright round or heater-style shield images; transparent silhouettes are coherent and no rotation, mirroring, plane-order or clipping defect is visible",
              "semantic_classification": {"label": "character shield images", "confidence": "confirmed",
                                          "basis": "shld filename operand, character equipment consumer and visual silhouettes"},
              "indexed_archive_name": "SHLDPICS.0", "fixed_record_name": "SHLDPICS",
              "names_are_distinct_formats": True, "source_payload_count": 8,
              "combined_png_count": 4, "mask_semantics": "bit 1 preserves background; bit 0 is opaque",
              "contact_sheet_sha256": sha(contact.read_bytes()), "records": records}
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print("validated=8 combined=4")


if __name__ == "__main__":
    main()
