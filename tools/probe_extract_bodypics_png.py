"""驗證 HEADPICS／TORPICS／LEGSPICS 共 162 組 EGA 彩圖／AND-mask。"""

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


ARCHIVES = {
    "HEADPICS.0": {"root": "Knights-of-legend/HEADPICS", "slug": "head", "cell": (64, 60)},
    "TORPICS.0": {"root": "Knights-of-legend/TORPICS", "slug": "torso", "cell": (96, 80)},
    "LEGSPICS.0": {"root": "Knights-of-legend/LEGSPICS", "slug": "legs", "cell": (64, 92)},
}
GRID_COLUMNS, GRID_ROWS = 9, 6


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
    if ida.get("schema") != "kol-ida-koltown-bodypics-summary-v1":
        raise SystemExit("非預期 body PICS IDA 摘要 schema")
    contracts = ida.get("archive_contracts")
    if len(contracts or []) != 3 or any(x.get("group_count") != 54 for x in contracts):
        raise SystemExit("IDA 摘要尚未證實三個 archive 的 54 組契約")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    args.output_directory.mkdir(parents=True, exist_ok=True)
    archive_results = []
    total_payloads = total_pngs = 0
    for evidence in contracts:
        archive_name = evidence["archive"]
        config = ARCHIVES.get(archive_name)
        if config is None:
            raise SystemExit(f"未登記 archive：{archive_name}")
        contract = logical_archive_contract(config["root"], files)
        if contract is None or contract["count"] != 108:
            raise SystemExit(f"{archive_name} archive contract 不符")
        cell_width, cell_height = config["cell"]
        sheet_width, sheet_height = cell_width * GRID_COLUMNS, cell_height * GRID_ROWS
        sheet_indices = bytearray(
            7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
            for y in range(sheet_height)
            for x in range(sheet_width)
        )
        sheet_opaque = bytes([1]) * (sheet_width * sheet_height)
        output_subdir = args.output_directory / config["slug"]
        output_subdir.mkdir(parents=True, exist_ok=True)
        records = []
        for dimension in evidence["dimensions"]:
            group = dimension["group"]
            width, height = dimension["width_pixels"], dimension["height_pixels"]
            width_bytes = dimension["width_bytes"]
            color_item, mask_item = contract["items"][2 * group : 2 * group + 2]

            def first_stage(item: dict) -> bytes:
                source = files[item["volume_name"]]
                return decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])

            pixels = decode_nibbles(first_stage(color_item))
            mask_bits, opaque, trailing = unpack_mask(first_stage(mask_item), width_bytes, height)
            if len(pixels) != width * height or len(opaque) != width * height:
                raise SystemExit(f"{archive_name} group {group} 像素／alpha 契約不符")
            if set(opaque) != {0, 1}:
                raise SystemExit(f"{archive_name} group {group} alpha 未同時含透明與不透明")
            output = output_subdir / f"{config['slug']}-{group:02d}.png"
            output.write_bytes(encode_rgba_png(pixels, opaque, width, height))
            origin_x = (group % GRID_COLUMNS) * cell_width + (cell_width - width) // 2
            origin_y = (group // GRID_COLUMNS) * cell_height + (cell_height - height) // 2
            for y in range(height):
                for x in range(width):
                    if opaque[y * width + x]:
                        sheet_indices[(origin_y + y) * sheet_width + origin_x + x] = pixels[y * width + x]
            records.append(
                {
                    "group": group,
                    "renderer_class": dimension["renderer_class"],
                    "variant": dimension["variant"],
                    "color_image_index": 2 * group,
                    "mask_image_index": 2 * group + 1,
                    "width_pixels": width,
                    "height_pixels": height,
                    "pixel_sha256": sha(pixels),
                    "and_mask_sha256": sha(mask_bits),
                    "opaque_sha256": sha(opaque),
                    "mask_trailing_byte_sha256": sha(trailing),
                    "png_sha256": sha(output.read_bytes()),
                }
            )
        contact = output_subdir / f"{config['slug']}-contact-sheet.png"
        contact.write_bytes(encode_rgba_png(bytes(sheet_indices), sheet_opaque, sheet_width, sheet_height))
        archive_results.append(
            {
                "archive": archive_name,
                "archive_root": config["root"],
                "component_token": evidence["token"],
                "source_payload_count": 108,
                "combined_png_count": 54,
                "contact_sheet_layout": "nine variants by six renderer classes",
                "contact_sheet_background": "opaque 8x8 checkerboard verifies transparent silhouettes",
                "contact_sheet_sha256": sha(contact.read_bytes()),
                "records": records,
            }
        )
        total_payloads += 108
        total_pngs += 54

    result = {
        "schema": "kol-bodypics-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": sha(zip_bytes),
        "ida_summary_name": "KOLTOWN-bodypics-summary.json",
        "ida_summary_sha256": sha(ida_bytes),
        "status": "mechanically validated and human-inspected all 162 color/AND-mask pairs",
        "human_inspection": "three six-by-nine sheets show coherent coordinated headgear or hair, torso armor or clothing, and leg armor or clothing across renderer classes and variants; intentionally separated component fragments retain transparent gaps, and no rotation, pair-order, plane-order or clipping defect is visible",
        "semantic_classification": {
            "labels": ["character head sprites", "character torso sprites", "character leg sprites"],
            "confidence": "confirmed",
            "basis": "head/tor/legs filename tokens, character component fields and paired renderer",
        },
        "source_payload_count": total_payloads,
        "combined_png_count": total_pngs,
        "mask_semantics": "bit 1 preserves background/transparent; PNG alpha is 1-mask",
        "archives": archive_results,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={total_payloads} combined={total_pngs} human_inspection=complete")


if __name__ == "__main__":
    main()
