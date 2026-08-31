"""驗證 TWN0..TWN7 全部城鎮圖片，輸出本機研究 PNG 與可版控統計。"""

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


EXPECTED_COUNTS = (25, 34, 30, 26, 26, 30, 26, 46)
GRID_COLUMNS = 5
CELL_WIDTH = 200
CELL_HEIGHT = 104


def decoded_item(contract: dict, files: dict[str, bytes], index: int) -> tuple[dict, bytes]:
    item = contract["items"][index]
    source = files[item["volume_name"]]
    compressed = source[item["offset"] : item["offset"] + item["compressed_bytes"]]
    return item, decode(compressed)


def geometry(town: int, index: int) -> tuple[int, int] | None:
    if town == 0:
        return 192, 96
    if index == 0:
        return 56, 57
    if index == 1:
        return None
    if index % 2 == 0:
        return 88, 64
    return 192, 96


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
    if ida.get("schema") != "kol-ida-koltown-twn-archives-summary-v1":
        raise SystemExit("非預期 KOLTOWN TWN IDA 摘要 schema")
    contract_evidence = ida.get("geometry_contract", {})
    if contract_evidence.get("status") != "confirmed":
        raise SystemExit("TWN 幾何契約尚未證實")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }

    mask_contract = logical_archive_contract("Knights-of-legend/KOLMPICS", files)
    if mask_contract is None or mask_contract["count"] != 83:
        raise SystemExit("非預期 KOLMPICS archive contract")
    _, mask_stage1 = decoded_item(mask_contract, files, 28)
    mask_bits, portrait_opaque, mask_trailing = unpack_mask(mask_stage1, 7, 57)
    if len(portrait_opaque) != 56 * 57:
        raise SystemExit("TWN 圖號 0 的共享遮罩尺寸不符")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    records: list[dict] = []
    contact_sheets: list[dict] = []
    png_count = 0
    excluded_count = 0
    for town, expected_count in enumerate(EXPECTED_COUNTS):
        root = f"Knights-of-legend/TWN{town}PICS"
        archive_contract = logical_archive_contract(root, files)
        if archive_contract is None or archive_contract["count"] != expected_count:
            raise SystemExit(f"TWN{town} archive contract 不符")

        rows = (expected_count + GRID_COLUMNS - 1) // GRID_COLUMNS
        sheet_width = GRID_COLUMNS * CELL_WIDTH
        sheet_height = rows * CELL_HEIGHT
        sheet_pixels = bytearray(
            7 if ((x // 8) + (y // 8)) % 2 == 0 else 8
            for y in range(sheet_height)
            for x in range(sheet_width)
        )
        for index in range(expected_count):
            item, stage1 = decoded_item(archive_contract, files, index)
            size = geometry(town, index)
            base = {
                "town_index": town,
                "image_index": index,
                "volume_name": item["volume_name"],
                "compressed_bytes": item["compressed_bytes"],
                "decoded_bytes": len(stage1),
            }
            if size is None:
                if len(stage1) != 1:
                    raise SystemExit(f"TWN{town}[1] placeholder 長度不符")
                records.append({
                    **base,
                    "disposition": "excluded_non_image_placeholder",
                    "decoded_hex": stage1.hex(),
                    "decoded_sha256": hashlib.sha256(stage1).hexdigest(),
                })
                excluded_count += 1
                continue

            width, height = size
            pixels = decode_nibbles(stage1)
            expected_pixels = width * height
            if len(pixels) != expected_pixels:
                raise SystemExit(
                    f"TWN{town}[{index}] 像素數不符：{len(pixels)} != {expected_pixels}"
                )
            opaque = portrait_opaque if town > 0 and index == 0 else bytes([1]) * len(pixels)
            output = args.output_directory / f"twn{town}-{index:02d}.png"
            output.write_bytes(encode_rgba_png(pixels, opaque, width, height))
            png_count += 1

            origin_x = (index % GRID_COLUMNS) * CELL_WIDTH + (CELL_WIDTH - width) // 2
            origin_y = (index // GRID_COLUMNS) * CELL_HEIGHT + (CELL_HEIGHT - height) // 2
            for y in range(height):
                source_row = y * width
                destination = (origin_y + y) * sheet_width + origin_x
                for x in range(width):
                    if opaque[source_row + x]:
                        sheet_pixels[destination + x] = pixels[source_row + x]
            records.append({
                **base,
                "disposition": "png",
                "width_pixels": width,
                "height_pixels": height,
                "pixel_count": len(pixels),
                "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
                "color_count": len(set(pixels)),
                "alpha_source": "KOLMPICS image 28; alpha=1-mask" if town > 0 and index == 0 else "opaque",
                "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
            })

        sheet = args.output_directory / f"twn{town}-contact-sheet.png"
        sheet.write_bytes(
            encode_rgba_png(
                bytes(sheet_pixels), bytes([1]) * len(sheet_pixels), sheet_width, sheet_height
            )
        )
        contact_sheets.append({
            "town_index": town,
            "width_pixels": sheet_width,
            "height_pixels": sheet_height,
            "png_sha256": hashlib.sha256(sheet.read_bytes()).hexdigest(),
        })

    result = {
        "schema": "kol-twn-archives-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_bytes).hexdigest(),
        "status": "mechanically validated all TWN payloads; orientation and scene semantics require human inspection",
        "source_archives": [f"TWN{town}PICS.0" for town in range(8)],
        "source_payload_count": sum(EXPECTED_COUNTS),
        "png_count": png_count,
        "excluded_placeholder_count": excluded_count,
        "geometry_contract": {
            "town_0_all": "192x96",
            "town_1_to_7_index_0": "56x57 color composed with KOLMPICS image 28 AND-mask",
            "town_1_to_7_index_1": "one-byte non-image placeholder",
            "town_1_to_7_even_index_from_2": "88x64",
            "town_1_to_7_odd_index_from_3": "192x96",
        },
        "mask_sha256": hashlib.sha256(mask_bits).hexdigest(),
        "mask_trailing_byte_hex": mask_trailing.hex(),
        "contact_sheets": contact_sheets,
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={len(records)} png={png_count} excluded={excluded_count}")


if __name__ == "__main__":
    main()
