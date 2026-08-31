"""驗證 TITLE 的 SET0/1/2 與 FACE 共 24 張 EGA 圖，輸出本機 PNG 與安全統計。"""

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


ARCHIVES = ("SET0PICS", "SET1PICS", "SET2PICS", "FACEPICS")


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
    if ida.get("schema") != "kol-ida-title-image-consumers-summary-v1":
        raise SystemExit("非預期 TITLE IDA 摘要 schema")
    contracts = {item["archive_root"]: item for item in ida["archive_contracts"]}
    if tuple(contracts) != ARCHIVES:
        raise SystemExit("TITLE archive contract 集合或順序不符")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    args.output_directory.mkdir(parents=True, exist_ok=True)
    records = []
    for root in ARCHIVES:
        evidence = contracts[root]
        rectangle = evidence["rectangle"]
        width = rectangle["width_pixels"]
        height = rectangle["height_pixels"]
        archive = logical_archive_contract(f"Knights-of-legend/{root}", files)
        if archive is None or archive["count"] != 6:
            raise SystemExit(f"{root} archive contract 不符")
        if evidence["image_indices"] != list(range(6)):
            raise SystemExit(f"{root} consumer 未完整覆蓋 0..5")
        for item in archive["items"]:
            source = files[item["volume_name"]]
            stage1 = decode(source[item["offset"]:item["offset"] + item["compressed_bytes"]])
            pixels = decode_nibbles(stage1)
            if len(pixels) != width * height or any(value > 15 for value in pixels):
                raise SystemExit(f"{root}[{item['index']}] 像素契約不符")
            opaque = bytes([1]) * len(pixels)
            output = args.output_directory / f"{root.lower()}-{item['index']:02d}.png"
            output.write_bytes(encode_rgba_png(pixels, opaque, width, height))
            records.append({
                "archive_root": root,
                "image_index": item["index"],
                "width_pixels": width,
                "height_pixels": height,
                "pixel_count": len(pixels),
                "color_count": len(set(pixels)),
                "pixel_sha256": hashlib.sha256(pixels).hexdigest(),
                "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
            })
    result = {
        "schema": "kol-title-set-face-ega-visual-probe-v1",
        "input_name": args.zip_path.name,
        "input_sha256": zip_digest,
        "ida_summary_name": args.ida_summary_path.name,
        "ida_summary_sha256": hashlib.sha256(ida_bytes).hexdigest(),
        "status": "mechanically validated 24 direct-blit 4-bit EGA images; orientation requires human inspection",
        "archive_count": 4,
        "source_payload_count": len(records),
        "png_count": len(records),
        "all_archive_indices_have_consumer": True,
        "all_pixel_counts_match": True,
        "records": records,
    }
    args.summary_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"validated={len(records)}")


if __name__ == "__main__":
    main()
