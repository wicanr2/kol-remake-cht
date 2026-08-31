"""驗證 KOLTOWN WEATPICS 18 筆圖層與 6 張合成 PNG。"""

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


ARCHIVE_ROOT = "Knights-of-legend/WEATPICS"
WIDTH, HEIGHT = 64, 62
FOUR_PLANE_MASK_GROUPS = {3, 4}


def digest(data: bytes) -> str:
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
    if digest(zip_bytes) != EXPECTED_ZIP_SHA256:
        raise SystemExit("非預期 ZIP SHA-256")
    ida_bytes = args.ida_summary_path.read_bytes()
    ida = json.loads(ida_bytes)
    if ida.get("schema") != "kol-ida-koltown-weatpics-discovery-summary-v1":
        raise SystemExit("非預期 WEATPICS IDA 摘要 schema")
    contract = ida.get("group_contract", {})
    if contract.get("composition") != "result = (base_color AND and_mask) OR or_color":
        raise SystemExit("IDA 摘要尚未證實 WEATPICS 合成契約")
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    archive_contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if archive_contract is None or archive_contract["count"] != 18:
        raise SystemExit("WEATPICS archive contract 不符")
    args.output_directory.mkdir(parents=True, exist_ok=True)
    decoded: list[bytes] = []
    records = []
    for item in archive_contract["items"]:
        index, group, role_offset = item["index"], item["index"] // 3, item["index"] % 3
        source = files[item["volume_name"]]
        stage1 = decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])
        if role_offset == 2 and group not in FOUR_PLANE_MASK_GROUPS:
            if len(stage1) != 497:
                raise SystemExit(f"WEATPICS[{index}] 1-bit mask 長度不符")
            pixels = bytes(15 if stage1[y * 8 + x // 8] & (0x80 >> (x % 8)) else 0
                           for y in range(HEIGHT) for x in range(WIDTH))
            role = "and_mask_1bit"
        else:
            pixels = decode_nibbles(stage1)
            role = ("base_color", "or_color", "and_mask_4plane")[role_offset]
        if len(pixels) != WIDTH * HEIGHT or any(value > 15 for value in pixels):
            raise SystemExit(f"WEATPICS[{index}] 像素契約不符")
        decoded.append(pixels)
        output = args.output_directory / f"weatpics-{index:02d}-{role}.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))
        records.append({"image_index": index, "group": group, "role": role,
                        "width_pixels": WIDTH, "height_pixels": HEIGHT,
                        "pixel_sha256": digest(pixels), "png_sha256": digest(output.read_bytes())})
    composites = []
    for group in range(6):
        base, overlay, mask = decoded[group * 3:group * 3 + 3]
        pixels = bytes((a & m) | b for a, b, m in zip(base, overlay, mask))
        output = args.output_directory / f"weatpics-group-{group}-composite.png"
        output.write_bytes(encode_rgba_png(pixels, bytes([1]) * len(pixels), WIDTH, HEIGHT))
        composites.append({"group": group, "formula": "(base_color AND and_mask) OR or_color",
                           "pixel_sha256": digest(pixels), "png_sha256": digest(output.read_bytes())})
    result = {"schema": "kol-koltown-weatpics-layered-visual-probe-v1",
              "input_name": args.zip_path.name, "input_sha256": digest(zip_bytes),
              "ida_summary_name": args.ida_summary_path.name,
              "ida_summary_sha256": digest(ida_bytes),
              "status": "mechanically validated all 18 source layers and 6 exact composites; semantic labels remain evidence-graded",
              "width_pixels": WIDTH, "height_pixels": HEIGHT,
              "source_payload_count": 18, "source_png_count": 18, "composite_png_count": 6,
              "runtime_png_policy": "mask PNG values are data, not RGBA transparency",
              "records": records, "composites": composites}
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print("validated=18 composites=6")


if __name__ == "__main__":
    main()
