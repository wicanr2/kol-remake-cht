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
AWD_VARIANT_MARKERS = ("PICS", "PCGA", "HERC")


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def is_candidate(name: str) -> bool:
    base = Path(name).name.upper()
    return (
        "PIC" in base
        or base in EXACT_NAMES
        or (base.startswith("AWD") and any(marker in base for marker in AWD_VARIANT_MARKERS))
    )


def fixed_record_contract(name: str, data: bytes) -> dict | None:
    base = Path(name).name.upper()
    if base not in EXACT_NAMES or len(data) != 0x2F80:
        return None
    record_bytes = 0x2F8
    return {
        "record_count": len(data) // record_bytes,
        "record_bytes": record_bytes,
        "file_bytes": len(data),
        "record_sha256": [
            sha256(data[offset : offset + record_bytes])
            for offset in range(0, len(data), record_bytes)
        ],
    }


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


def logical_archive_contract(root_name: str, files: dict[str, bytes]) -> dict | None:
    """依 KOLTOWN sub_1A4FE 契約解析 `.0` header 與最多五個實體分卷。"""

    first_name = f"{root_name}.0"
    first = files.get(first_name)
    if first is None:
        return None
    first_contract = archive_contract(first)
    if first_contract is None or not first_contract["all_payload_bounds_valid"]:
        return None

    volume_names = []
    for volume_index in range(5):
        name = f"{root_name}.{volume_index}"
        if name not in files:
            break
        volume_names.append(name)
    if not volume_names:
        return None
    if any(f"{root_name}.{index}" in files for index in range(len(volume_names) + 1, 5)):
        return None

    thresholds = list(struct.unpack_from("<5H", first, 0))
    last_lengths = list(struct.unpack_from("<5H", first, 10))
    first_payload_offset = struct.unpack_from("<H", first, 20)[0]
    if first_payload_offset < 22 or (first_payload_offset - 20) % 2:
        return None
    table_slots = (first_payload_offset - 20) // 2
    offsets = list(struct.unpack_from(f"<{table_slots}H", first, 20))

    if len(volume_names) == 1:
        total_count = thresholds[0]
        if total_count == 0 or total_count > table_slots:
            return None
        boundaries = [0, total_count]
    else:
        used_thresholds = thresholds[: len(volume_names) - 1]
        if (
            any(value == 0 for value in used_thresholds)
            or any(left >= right for left, right in zip(used_thresholds, used_thresholds[1:]))
            or any(thresholds[len(volume_names) - 1 :])
        ):
            return None
        final_start = used_thresholds[-1]
        final_size = len(files[volume_names[-1]])
        sentinel_candidates = [
            index
            for index in range(final_start + 1, table_slots)
            if offsets[index] == final_size and not any(offsets[index + 1 :])
        ]
        if len(sentinel_candidates) != 1:
            return None
        total_count = sentinel_candidates[0]
        boundaries = [0, *used_thresholds, total_count]

    items = []
    volume_records = []
    all_bounds_valid = True
    for volume_index, volume_name in enumerate(volume_names):
        data = files[volume_name]
        start_index = boundaries[volume_index]
        end_index = boundaries[volume_index + 1]
        local_offsets = offsets[start_index:end_index]
        monotonic = all(left < right for left, right in zip(local_offsets, local_offsets[1:]))
        expected_first = first_payload_offset if volume_index == 0 else 0
        starts_at_expected_offset = bool(local_offsets) and local_offsets[0] == expected_first
        lengths = [right - left for left, right in zip(local_offsets, local_offsets[1:])]
        final_length = last_lengths[volume_index]
        lengths.append(final_length)
        payload_end = local_offsets[-1] + final_length if local_offsets else 0
        if volume_index == len(volume_names) - 1 and len(volume_names) > 1:
            sentinel_matches_end = offsets[total_count] == len(data)
        elif len(volume_names) == 1:
            sentinel_matches_end = payload_end <= len(data)
        else:
            sentinel_matches_end = payload_end == len(data)
        bounds_valid = (
            monotonic
            and starts_at_expected_offset
            and all(length > 0 for length in lengths)
            and payload_end <= len(data)
            and sentinel_matches_end
        )
        all_bounds_valid = all_bounds_valid and bounds_valid
        volume_records.append(
            {
                "name": volume_name,
                "size": len(data),
                "sha256": sha256(data),
                "volume_index": volume_index,
                "first_image_index": start_index,
                "image_count": end_index - start_index,
                "first_payload_offset": local_offsets[0] if local_offsets else None,
                "last_payload_length": final_length,
                "payload_end": payload_end,
                "trailing_bytes_after_indexed_payloads": len(data) - payload_end,
                "strictly_increasing_local_offsets": monotonic,
                "starts_at_expected_offset": starts_at_expected_offset,
                "payload_end_equals_file_size": payload_end == len(data),
                "payload_end_within_file": payload_end <= len(data),
                "all_payload_bounds_valid": bounds_valid,
            }
        )
        for image_index, (offset, length) in enumerate(
            zip(local_offsets, lengths), start=start_index
        ):
            items.append(
                {
                    "index": image_index,
                    "volume_index": volume_index,
                    "volume_name": volume_name,
                    "offset": offset,
                    "compressed_bytes": length,
                }
            )

    return {
        "root_name": root_name,
        "count": total_count,
        "volume_count": len(volume_names),
        "fixed_header_bytes": 20,
        "index_entry_bytes": 2,
        "index_table_slots": table_slots,
        "cumulative_volume_thresholds": thresholds,
        "last_payload_lengths": last_lengths,
        "all_payload_bounds_valid": all_bounds_valid,
        "volumes": volume_records,
        "items": items,
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
                    "fixed_record_contract": fixed_record_contract(info.filename, data),
                }
            )

        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }

    logical_archives = []
    for name in sorted(files, key=str.upper):
        if not name.upper().endswith(".0") or not is_candidate(name):
            continue
        contract = logical_archive_contract(name[:-2], files)
        if contract is not None:
            logical_archives.append(contract)

    result = {
        "schema": "kol-image-asset-inventory-v2",
        "input_name": args.zip_path.name,
        "input_sha256": digest,
        "candidate_rule": (
            "basename contains PIC (case-insensitive), plus exact legacy names, plus AWD PCGA/HERC variants"
        ),
        "candidate_count": len(records),
        "logical_archive_count": len(logical_archives),
        "logical_image_count": sum(item["count"] for item in logical_archives),
        "logical_archives": logical_archives,
        "records": sorted(records, key=lambda item: item["name"].upper()),
    }
    volume_index = {
        volume["name"]: (contract["root_name"], contract["volume_count"], volume["volume_index"])
        for contract in logical_archives
        for volume in contract["volumes"]
    }
    for record in result["records"]:
        logical = volume_index.get(record["name"])
        if logical is None:
            record["logical_archive_root"] = None
            record["logical_volume_index"] = None
            record["archive_contract_scope"] = None
            continue
        root_name, volume_count, index = logical
        record["logical_archive_root"] = root_name
        record["logical_volume_index"] = index
        if record["archive_contract"] is not None:
            record["archive_contract_scope"] = (
                "complete_single_volume" if volume_count == 1 else "first_volume_only"
            )
        else:
            record["archive_contract_scope"] = "continuation_volume"
    args.output_path.parent.mkdir(parents=True, exist_ok=True)
    args.output_path.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
