"""機械驗證 KOLBOOK 的 BOOKDATA 職業記錄；不輸出原版 raw bytes。"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract
from probe_image_lzw import decode


ROOT = "Knights-of-legend/BOOKDATA"
ITEM_INDEX = 2
EXPECTED_COMPRESSED_BYTES = 424
EXPECTED_DECODED_BYTES = 625
EXPECTED_DECODED_SHA256 = "848ad877f8176b8b3d950d8853b36cf7e0271629c8822ce15c7f420ea2e7b321"
DICE_SIDES = [3, 4, 6, 8, 10, 12, 20, 100]
EXPECTED_TEXT_DECODED_SHA256 = "0378dedd8b41e7dc07a4e154fe38a662ae53bb30e02c924fce3745a4f1d36b41"
EXPECTED_WEAPON_DECODED_SHA256 = "65650081e865b3b23f29d1a555d3d49bd59baf5f9801a9b3e849f32acadbb77f"
EXPECTED_SHIELD_DECODED_SHA256 = "d8004e1c4a71bece38f848a73624867dcfe6119efc6aadaebd0219bcd811a562"
EXPECTED_RACES = ["HUMAN", "ELVEN", "DWARVEN", "KELDEN"]
EXPECTED_SECONDARY = ["MALE", "FEMALE"]
EXPECTED_MENU_NAMES = [
    "BARBARIAN", "RANGER", "WARRION", "SQUIRE", "DARKGUARD", "WATCHMAN",
    "PLAINSMAN", "HUNTER", "REGULAR", "HIGHWAYMAN", "PIRATE", "ROGUE",
    "TIGRESS", "AMAZON", "HUNTRESS", "PLAINSWOMAN", "BREKLAND", "KLVAR",
    "MELOD", "PYAR", "THISM", "USIP", "TUNNELER", "RATGUARD", "TROLLBANE",
    "MILITIA", "DIGGER", "SPIDER GUARD", "ORCBANE", "LEVY", "CLIFF GUARD",
    "ROCK RANGER", "FAR SEEKER",
]


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("zip_path", type=Path)
    parser.add_argument("safe_output", type=Path)
    parser.add_argument("private_output", type=Path)
    args = parser.parse_args()

    zip_bytes = args.zip_path.read_bytes()
    if sha(zip_bytes) != EXPECTED_ZIP_SHA256:
        raise SystemExit("非預期原版 ZIP SHA-256")

    with zipfile.ZipFile(args.zip_path) as archive:
        files = {
            info.filename: archive.read(info)
            for info in archive.infolist()
            if not info.is_dir()
        }
    contract = logical_archive_contract(ROOT, files)
    if contract is None or not contract["all_payload_bounds_valid"]:
        raise SystemExit("BOOKDATA archive 契約無效")
    if contract["count"] != 5:
        raise SystemExit(f"非預期 BOOKDATA 項目數：{contract['count']}")

    item = contract["items"][ITEM_INDEX]
    if item["compressed_bytes"] != EXPECTED_COMPRESSED_BYTES:
        raise SystemExit("職業項目壓縮長度改變")
    source = files[item["volume_name"]]
    payload = source[item["offset"] : item["offset"] + item["compressed_bytes"]]
    decoded = decode(payload)
    if len(decoded) != EXPECTED_DECODED_BYTES or sha(decoded) != EXPECTED_DECODED_SHA256:
        raise SystemExit("職業項目解碼結果改變")
    if decoded[-1] != 0:
        raise SystemExit("職業項目尾 byte 不再是 0")

    text_item = contract["items"][1]
    text_source = files[text_item["volume_name"]]
    text_payload = text_source[
        text_item["offset"] : text_item["offset"] + text_item["compressed_bytes"]
    ]
    text_decoded = decode(text_payload)
    if sha(text_decoded) != EXPECTED_TEXT_DECODED_SHA256:
        raise SystemExit("BOOKDATA 文字項目解碼結果改變")
    lines = [line.decode("ascii") for line in text_decoded.split(b"\r")]
    if lines[14:18] != EXPECTED_RACES or lines[18:20] != EXPECTED_SECONDARY:
        raise SystemExit("種族／secondary 選單文字順序改變")
    if lines[21:54] != EXPECTED_MENU_NAMES:
        raise SystemExit("職業選單文字順序改變")

    decoded_auxiliary = {}
    for index, expected_length, expected_hash, label in [
        (3, 225, EXPECTED_WEAPON_DECODED_SHA256, "weapon_templates"),
        (4, 761, EXPECTED_SHIELD_DECODED_SHA256, "initial_shield_template"),
    ]:
        auxiliary_item = contract["items"][index]
        auxiliary_source = files[auxiliary_item["volume_name"]]
        auxiliary_payload = auxiliary_source[
            auxiliary_item["offset"] : auxiliary_item["offset"] + auxiliary_item["compressed_bytes"]
        ]
        auxiliary_decoded = decode(auxiliary_payload)
        if len(auxiliary_decoded) != expected_length or sha(auxiliary_decoded) != expected_hash:
            raise SystemExit(f"{label} 解碼結果改變")
        if auxiliary_decoded[-1] != 0:
            raise SystemExit(f"{label} 尾 byte 不再是 0")
        decoded_auxiliary[label] = {
            "archive_item_index": index,
            "compressed_bytes": auxiliary_item["compressed_bytes"],
            "decoded_bytes": len(auxiliary_decoded),
            "decoded_sha256": sha(auxiliary_decoded),
            "trailing_byte_hex": "00",
        }

    records = [decoded[index * 16 : (index + 1) * 16] for index in range(39)]
    if any(len(record) != 16 for record in records):
        raise SystemExit("職業記錄 stride 改變")
    if any(record[7] != 0 or record[15] != 0 for record in records):
        raise SystemExit("職業記錄保留 byte 改變")
    if any((code >> 4) >= len(DICE_SIDES) for record in records for code in record[:7]):
        raise SystemExit("職業骰面 selector 超界")
    duplicate_pairs = [[left, right] for left, right in zip(range(28, 32), range(32, 36))]
    if any(records[left] != records[right] for left, right in duplicate_pairs):
        raise SystemExit("矮人雙名稱記錄不再成對相同")

    group_contract = [
        {"selector": "race=0 HUMAN, secondary=0 MALE", "record_indices": [0, 11], "count": 12},
        {"selector": "race=0 HUMAN, secondary=1 FEMALE", "record_indices": [12, 15], "count": 4},
        {"selector": "race=1 ELVEN, secondary=0 MALE", "record_indices": [16, 21], "count": 6},
        {"selector": "race=1 ELVEN, secondary=1 FEMALE", "record_indices": [22, 27], "count": 6},
        {"selector": "race=2 DWARVEN", "record_indices": [28, 35], "count": 8},
        {"selector": "race=3 KELDEN", "record_indices": [36, 38], "count": 3},
    ]
    record_names = [
        *EXPECTED_MENU_NAMES[0:16],
        *EXPECTED_MENU_NAMES[16:22],
        *EXPECTED_MENU_NAMES[16:22],
        *EXPECTED_MENU_NAMES[22:33],
    ]
    if len(record_names) != 39:
        raise SystemExit("職業名稱映射數量錯誤")
    safe = {
        "schema": "kol-kolbook-profession-data-summary-v1",
        "input_name": "Knights-of-legend.zip",
        "input_sha256": sha(zip_bytes),
        "archive_root": ROOT,
        "archive_item_index": ITEM_INDEX,
        "compressed_bytes": len(payload),
        "decoded_bytes": len(decoded),
        "decoded_sha256": sha(decoded),
        "record_count": len(records),
        "record_stride": 16,
        "trailing_bytes": 1,
        "trailing_byte_hex": "00",
        "dice_sides": DICE_SIDES,
        "record_layout": {
            "dice_codes": {"offset": 0, "count": 7},
            "reserved_zero_0": {"offset": 7, "count": 1},
            "internal_bases": {"offset": 8, "count": 7},
            "reserved_zero_1": {"offset": 15, "count": 1},
        },
        "text_item_decoded_sha256": sha(text_decoded),
        "weapon_template_contract": {
            **decoded_auxiliary["weapon_templates"],
            "template_count": 14,
            "template_stride": 16,
            "non_template_trailing_bytes": 1,
        },
        "initial_shield_template_contract": {
            **decoded_auxiliary["initial_shield_template"],
            "pixel_bytes_consumed": 760,
            "non_pixel_trailing_bytes": 1,
        },
        "race_enum": [{"value": index, "name": name} for index, name in enumerate(EXPECTED_RACES)],
        "secondary_enum": [{"value": index, "name": name} for index, name in enumerate(EXPECTED_SECONDARY)],
        "selector_groups": group_contract,
        "record_name_map": [{"record_index": index, "menu_name": name} for index, name in enumerate(record_names)],
        "byte_identical_record_pairs": duplicate_pairs,
        "status": "mechanically confirmed structure and profession menu-name order",
    }
    private = {
        "schema": "kol-kolbook-profession-data-private-v1",
        **safe,
        "records": [
            {
                "index": index,
                "raw_hex": record.hex(),
                "dice_codes": list(record[:7]),
                "internal_bases": list(record[8:15]),
            }
            for index, record in enumerate(records)
        ],
    }
    args.safe_output.parent.mkdir(parents=True, exist_ok=True)
    args.private_output.parent.mkdir(parents=True, exist_ok=True)
    args.safe_output.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    args.private_output.write_text(json.dumps(private, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
