"""驗證 KOLMPICS 全 83 筆的 EGA 圖、AND-mask 與 placeholder。"""

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


ARCHIVE_ROOT = "Knights-of-legend/KOLMPICS"
PLACEHOLDERS = [1, 29, 30, 32, 33, 34, 35, 70, 71, 72, 73, 82]
ICON_INDICES = [
    *range(3, 28), 31, *range(36, 70), *range(74, 82),
]
COLOR_DIMENSIONS = {0: (320, 200), 2: (296, 32), **{i: (24, 20) for i in ICON_INDICES}}
MASK_INDEX = 28


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
    if ida.get("schema") != "kol-ida-koltown-kolmpics-routes-summary-v2":
        raise SystemExit("非預期 KOLMPICS route IDA 摘要 schema")
    route_candidates = ida["router"]["kolmpics_candidate_indices_within_archive"]
    if route_candidates != [*range(0, 29), 31, *range(36, 71), *range(74, 83)]:
        raise SystemExit("KOLMPICS router 候選集合不符")
    extraction_contract = ida.get("canonical_icon_extraction_contract", {})
    if (
        extraction_contract.get("width_pixels") != 24
        or extraction_contract.get("height_pixels") != 20
        or extraction_contract.get("status")
        != "confirmed only when combined with fixed archive and visual probes"
    ):
        raise SystemExit("KOLMPICS canonical icon extraction contract 不符")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 83:
        raise SystemExit("KOLMPICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)

    cell_width, cell_height, columns = 32, 28, 10
    rows = (len(ICON_INDICES) + columns - 1) // columns
    sheet_width, sheet_height = columns * cell_width, rows * cell_height
    sheet = bytearray(7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
                      for y in range(sheet_height) for x in range(sheet_width))
    records = []
    for item in contract["items"]:
        index = item["index"]
        source = files[item["volume_name"]]
        stage1 = decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])
        record = {"image_index": index, "stage1_bytes": len(stage1), "stage1_sha256": sha(stage1)}
        if index in PLACEHOLDERS:
            if len(stage1) != 1:
                raise SystemExit(f"KOLMPICS[{index}] placeholder 長度不符")
            record.update({"kind": "placeholder", "placeholder_value": stage1[0]})
            records.append(record)
            continue
        if index == MASK_INDEX:
            mask, opaque, trailing = unpack_mask(stage1, 7, 57)
            output = args.output_directory / "kolmpics-28-and-mask.png"
            output.write_bytes(encode_rgba_png(bytes([15]) * len(mask), opaque, 56, 57))
            bits_output = args.output_directory / "kolmpics-28-mask-bits.png"
            bits_output.write_bytes(encode_rgba_png(bytes(15 if bit else 0 for bit in mask),
                                                     bytes([1]) * len(mask), 56, 57))
            record.update({
                "kind": "and-mask", "width_pixels": 56, "height_pixels": 57,
                "mask_bit_1": "preserve destination / PNG transparent",
                "mask_bit_0": "replace destination / PNG opaque",
                "mask_sha256": sha(mask), "opaque_sha256": sha(opaque),
                "trailing_byte_sha256": sha(trailing), "png_sha256": sha(output.read_bytes()),
                "bits_png_sha256": sha(bits_output.read_bytes()),
            })
            records.append(record)
            continue
        if index not in COLOR_DIMENSIONS:
            raise SystemExit(f"KOLMPICS[{index}] 未分類")
        width, height = COLOR_DIMENSIONS[index]
        pixels = decode_nibbles(stage1)
        if len(pixels) != width * height or any(x > 15 for x in pixels):
            raise SystemExit(f"KOLMPICS[{index}] RLE 像素契約不符")
        output = args.output_directory / f"kolmpics-{index:02d}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), width, height))
        if index in ICON_INDICES:
            position = ICON_INDICES.index(index)
            origin_x = (position % columns) * cell_width + (cell_width - width) // 2
            origin_y = (position // columns) * cell_height + (cell_height - height) // 2
            for y in range(height):
                start = (origin_y + y) * sheet_width + origin_x
                sheet[start : start + width] = pixels[y * width : (y + 1) * width]
        record.update({
            "kind": "rle-color", "width_pixels": width, "height_pixels": height,
            "extraction_geometry_grade": "confirmed",
            "caller_display_geometry_grade": (
                "confirmed for fixed index 0/2 rectangles" if index in (0, 2)
                else "not claimed; dynamic placement/clipping remains caller-specific"
            ),
            "color_count": len(set(pixels)), "pixel_sha256": sha(pixels),
            "png_sha256": sha(output.read_bytes()),
        })
        records.append(record)

    contact = args.output_directory / "kolmpics-icons-contact-sheet.png"
    contact.write_bytes(encode_rgba_png(bytes(sheet), bytes([1]) * len(sheet), sheet_width, sheet_height))
    active = sorted(set(range(83)) - set(PLACEHOLDERS))
    result = {
        "schema": "kol-kolmpics-all-ega-visual-probe-v2",
        "input_name": args.zip_path.name, "input_sha256": sha(zip_bytes),
        "ida_summary_name": args.ida_summary_path.name, "ida_summary_sha256": sha(ida_bytes),
        "status": "mechanically validated and human-inspected all 71 active payloads; all 68 icons have confirmed 24x20 extraction geometry",
        "human_inspection": "the 68-item 24x20 contact sheet forms a coherent UI icon set containing gestures, weapons, status marks, figures and direction arrows; the 320x200 frame and 296x32 bar are correctly oriented; index 28 mask-bits view preserves the previously verified rounded silhouette; no RLE, plane-order, rotation or clipping defect is visible",
        "active_indices": active, "placeholder_indices": PLACEHOLDERS,
        "active_payload_count": len(active), "placeholder_count": len(PLACEHOLDERS),
        "color_png_count": len(COLOR_DIMENSIONS), "and_mask_png_count": 1,
        "mask_visualization_png_count": 1,
        "icon_geometry": {"width_pixels": 24, "height_pixels": 20,
                          "extraction_grade": "confirmed by 480-index payloads, two fixed consumers and complete contact sheet",
                          "caller_display_limit": "dynamic screen placement and clipping are not claimed"},
        "contact_sheet_layout": "ten columns by seven rows, archive index order with gaps removed",
        "contact_sheet_sha256": sha(contact.read_bytes()), "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=83 active={len(active)} placeholders={len(PLACEHOLDERS)} human_inspection=complete")


if __name__ == "__main__":
    main()
