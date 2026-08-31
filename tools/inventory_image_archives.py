"""盤點固定《武士傳說》ZIP 內的圖片候選，不輸出任何原始 payload。"""

from __future__ import annotations

import argparse
import hashlib
import json
import struct
import zipfile
from pathlib import Path


EXPECTED_ZIP_SHA256 = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
EXACT_NAMES = {"CHARPICS", "SHLDPICS", "STEEDPIC"}


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def is_candidate(name: str) -> bool:
    base = Path(name).name.upper()
    return "PIC" in base or base in EXACT_NAMES


def archive_contract(data: bytes) -> dict | None:
    if len(data) < 22:
        return None
    count = struct.unpack_from("<H", data, 0)[0]
    table_end = 20 + count * 2
    if count == 0 or count > 4096 or table_end > len(data):
        return None
    offsets = list(struct.unpack_from(f"<{count}H", data, 20))
    last_length = struct.unpack_from("<H", data, 10)[0]
    monotonic = all(left < right for left, right in zip(offsets, offsets[1:]))
    first_after_table = offsets[0] >= table_end
    payload_end = offsets[-1] + last_length
    payload_within_file = payload_end <= len(data)
    exact_file_end = payload_end == len(data)
    in_bounds = monotonic and first_after_table and payload_within_file
    lengths = [right - left for left, right in zip(offsets, offsets[1:])]
    lengths.append(last_length)
    return {
        "count": count,
        "fixed_header_bytes": 20,
        "index_entry_bytes": 2,
        "index_start": 20,
        "index_end": table_end,
        "first_payload_offset": offsets[0],
        "last_payload_offset": offsets[-1],
        "last_payload_length": last_length,
        "payload_end": payload_end,
        "trailing_bytes_after_indexed_payloads": len(data) - payload_end,
        "strictly_increasing_offsets": monotonic,
        "first_payload_after_index": first_after_table,
        "payload_end_equals_file_size": exact_file_end,
        "payload_end_within_file": payload_within_file,
        "all_payload_bounds_valid": in_bounds,
        "minimum_compressed_payload_bytes": min(lengths),
        "maximum_compressed_payload_bytes": max(lengths),
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("output_path", type=Path)
    args = parser.parse_args()

    zip_bytes = args.zip_path.read_bytes()
    digest = sha256(zip_bytes)
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")

    records = []
    with zipfile.ZipFile(args.zip_path) as archive:
        bad = archive.testzip()
        if bad is not None:
            raise SystemExit(f"ZIP CRC 失敗：{bad}")
        for info in archive.infolist():
            if info.is_dir() or not is_candidate(info.filename):
                continue
            data = archive.read(info)
            records.append(
                {
                    "name": info.filename,
                    "size": len(data),
                    "sha256": sha256(data),
                    "archive_contract": archive_contract(data),
                }
            )

    result = {
        "schema": "kol-image-asset-inventory-v1",
        "input_name": args.zip_path.name,
        "input_sha256": digest,
        "candidate_rule": "basename contains PIC (case-insensitive), plus exact legacy names",
        "candidate_count": len(records),
        "records": sorted(records, key=lambda item: item["name"].upper()),
    }
    args.output_path.parent.mkdir(parents=True, exist_ok=True)
    args.output_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
