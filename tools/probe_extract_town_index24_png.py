"""輸出 TWN*PICS 全域圖號 24 的本機測試 PNG；不得作 production extractor。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_extract_comb_png import decode_nibbles, encode_rgb_png
from probe_image_lzw import decode


IMAGE_INDEX = 24


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("output_dir", type=Path)
    parser.add_argument("summary_path", type=Path)
    args = parser.parse_args()
    if "workplace" not in args.output_dir.resolve().parts:
        raise SystemExit("測試 PNG 只可寫入 ignored workplace")

    zip_bytes = args.zip_path.read_bytes()
    digest = hashlib.sha256(zip_bytes).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")

    records = []
    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
        for town_index in range(8):
            width, height = (192, 96) if town_index == 0 else (88, 64)
            root = f"Knights-of-legend/TWN{town_index}PICS"
            contract = logical_archive_contract(root, files)
            if contract is None or not contract["all_payload_bounds_valid"]:
                raise SystemExit(f"分卷契約失敗：{root}")
            item = contract["items"][IMAGE_INDEX]
            source = files[item["volume_name"]]
            start = item["offset"]
            end = start + item["compressed_bytes"]
            first_stage = decode(source[start:end])
            indices = decode_nibbles(first_stage)
            if len(indices) != width * height:
                raise SystemExit(
                    f"{root}[{IMAGE_INDEX}] 非預期像素數：{len(indices)}"
                )
            if not indices or max(indices) > 15:
                raise SystemExit(f"{root}[{IMAGE_INDEX}] 色彩索引超出 4-bit 範圍")
            output = args.output_dir / f"twn{town_index}-index-{IMAGE_INDEX:04d}.png"
            output.parent.mkdir(parents=True, exist_ok=True)
            output.write_bytes(encode_rgb_png(indices, width, height))
            records.append(
                {
                    "source_root": root,
                    "source_index": IMAGE_INDEX,
                    "volume_index": item["volume_index"],
                    "volume_name": item["volume_name"],
                    "compressed_offset": start,
                    "compressed_bytes": item["compressed_bytes"],
                    "first_stage_bytes": len(first_stage),
                    "pixel_count": len(indices),
                    "width": width,
                    "height": height,
                    "minimum_palette_index": min(indices),
                    "maximum_palette_index": max(indices),
                    "distinct_palette_indices": len(set(indices)),
                    "pixel_indices_sha256": hashlib.sha256(indices).hexdigest(),
                    "png_sha256": hashlib.sha256(output.read_bytes()).hexdigest(),
                }
            )

    summary = {
        "schema": "kol-town-index24-png-probe-summary-v1",
        "input_name": args.zip_path.name,
        "input_sha256": digest,
        "status": "disposable RE probe; PNG files remain in ignored workplace",
        "evidence_scope": "TWN0..7 PICS global image index 24, EGA mode 0Dh",
        "records": records,
    }
    args.summary_path.parent.mkdir(parents=True, exist_ok=True)
    args.summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
