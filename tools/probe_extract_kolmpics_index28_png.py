"""驗證 KOLMPICS 圖號 28 的 1-bit EGA 遮罩候選。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_monpics_png import encode_rgba_png, unpack_mask
from probe_image_lzw import decode


ARCHIVE_ROOT = "Knights-of-legend/KOLMPICS"
IMAGE_INDEX = 28


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
    if ida_summary.get("schema") != "kol-ida-koltown-kolmpics-index28-summary-v1":
        raise SystemExit("非預期 KOLMPICS 圖號 28 IDA 摘要 schema")
    consumer = ida_summary["consumer"]
    if consumer["image_index"] != IMAGE_INDEX:
        raise SystemExit("IDA 摘要圖號不符")
    rectangle = consumer["source_rectangle"]
    width_bytes = rectangle["width_bytes"]
    width = rectangle["width_pixels"]
    height = rectangle["height_pixels"]

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ARCHIVE_ROOT, files)
    if contract is None or contract["count"] != 83:
        raise SystemExit("非預期 KOLMPICS archive contract")
    item = contract["items"][IMAGE_INDEX]
    source = files[item["volume_name"]]
    first_stage = decode(source[item["offset"] : item["offset"] + item["compressed_bytes"]])
    mask, trailing = unpack_mask(first_stage, width_bytes, height)
    expected_pixels = width * height
    if len(mask) != expected_pixels:
        raise SystemExit(f"遮罩像素數不符：{len(mask)} != {expected_pixels}")

    args.output_directory.mkdir(parents=True, exist_ok=True)
    bits_path = args.output_directory / "kolmpics-index-0028-mask-bits.png"
    alpha_path = args.output_directory / "kolmpics-index-0028-composite-alpha.png"
    bits_path.write_bytes(encode_rgba_png(bytes(15 if bit else 0 for bit in mask), bytes([1]) * len(mask), width, height))
    composite_opaque = bytes(1 - bit for bit in mask)
    alpha_path.write_bytes(encode_rgba_png(bytes([15]) * len(mask), composite_opaque, width, height))
    result = {
        "schema": "kol-kolmpics-index28-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_summary_bytes).hexdigest(),
        "status": "mechanically validated MSB-first EGA AND-mask; orientation requires human inspection",
        "archive_root": ARCHIVE_ROOT,
        "image_index": IMAGE_INDEX,
        "width_pixels": width,
        "height_pixels": height,
        "decoded_bytes": len(first_stage),
        "visible_mask_bytes": width_bytes * height,
        "trailing_bytes": len(trailing),
        "trailing_byte_sha256": hashlib.sha256(trailing).hexdigest(),
        "preserve_transparent_bit_count": sum(mask),
        "replace_opaque_bit_count": len(mask) - sum(mask),
        "composite_formula": "destination = (destination AND mask) OR color",
        "mask_sha256": hashlib.sha256(mask).hexdigest(),
        "bits_png_sha256": hashlib.sha256(bits_path.read_bytes()).hexdigest(),
        "alpha_png_sha256": hashlib.sha256(alpha_path.read_bytes()).hexdigest(),
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated=1 mask={bits_path}")


if __name__ == "__main__":
    main()
