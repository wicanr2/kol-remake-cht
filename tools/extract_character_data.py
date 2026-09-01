#!/usr/bin/env python3
"""從固定原版輸入抽出角色建立所需的 typed data。

remake 的規則層只讀這份 JSON，不在執行期解析 `BOOKDATA`。內容全部來自
`tools/probe_kolbook_profession_data.py` 已證實的契約：item 2 是 39 筆 16-byte 職業
記錄（offset 0–6 骰式、8–14 內部基值），item 1 是 CR 分隔的文字資源。名稱順序、
selector 分組與骰面表都沿用該探針的常數，不另行推定。

  python3 tools/extract_character_data.py --zip kol.zip --repo . \\
      --out assets/original-derived-data/character-creation.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract  # noqa: E402
from probe_image_lzw import decode as lzw_decode  # noqa: E402
from probe_kolbook_profession_data import (  # noqa: E402
    DICE_SIDES, EXPECTED_DECODED_SHA256, EXPECTED_MENU_NAMES, EXPECTED_RACES,
    EXPECTED_SECONDARY, EXPECTED_TEXT_DECODED_SHA256, ITEM_INDEX, ROOT,
)

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")

# 主要屬性的畫面順序，出自原版 New Character 頁的左半頁標籤。
PRIMARY_ATTRIBUTES = ["STRENGTH", "QUICKNESS", "SIZE", "HEALTH",
                      "FORESIGHT", "CHARISMA", "INTELLECT"]
DERIVED_ATTRIBUTES = [
    {"name": "BALANCE", "from": ["STRENGTH", "INTELLECT"]},
    {"name": "ENDURANCE", "from": ["STRENGTH", "HEALTH"]},
    {"name": "BODY PTS.", "from": ["SIZE", "HEALTH"]},
]

# selector 分組與名稱映射沿用已證實的探針契約。
GROUPS = [
    {"race": 0, "secondary": 0, "first": 0, "last": 11},
    {"race": 0, "secondary": 1, "first": 12, "last": 15},
    {"race": 1, "secondary": 0, "first": 16, "last": 21},
    {"race": 1, "secondary": 1, "first": 22, "last": 27},
    {"race": 2, "secondary": None, "first": 28, "last": 35},
    {"race": 3, "secondary": None, "first": 36, "last": 38},
]


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--repo", default=".")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    repo = pathlib.Path(args.repo).resolve()
    out = pathlib.Path(args.out).resolve()
    try:
        relative = out.relative_to(repo)
    except ValueError:
        raise SystemExit("輸出必須位於 repo 內")
    if ALLOWED_OUTPUT_PARENT != pathlib.PurePosixPath(relative.as_posix()).parent:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")

    zip_path = pathlib.Path(args.zip)
    digest = sha(zip_path.read_bytes())
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")
    with zipfile.ZipFile(zip_path) as archive:
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}

    contract = logical_archive_contract(ROOT, files)
    if contract is None:
        raise SystemExit("BOOKDATA archive contract 不符")

    def payload(index: int) -> bytes:
        item = contract["items"][index]
        src = files[item["volume_name"]]
        return lzw_decode(src[item["offset"]:item["offset"] + item["compressed_bytes"]])

    records_raw = payload(ITEM_INDEX)
    if sha(records_raw) != EXPECTED_DECODED_SHA256:
        raise SystemExit("職業記錄 payload 已改變")
    text_raw = payload(1)
    if sha(text_raw) != EXPECTED_TEXT_DECODED_SHA256:
        raise SystemExit("文字資源 payload 已改變")

    lines = [line.decode("ascii") for line in text_raw.split(b"\r")]
    if lines[14:18] != EXPECTED_RACES or lines[18:20] != EXPECTED_SECONDARY:
        raise SystemExit("種族／secondary 文字順序改變")
    if lines[21:54] != EXPECTED_MENU_NAMES:
        raise SystemExit("職業選單文字順序改變")

    record_names = [
        *EXPECTED_MENU_NAMES[0:16],
        *EXPECTED_MENU_NAMES[16:22],
        *EXPECTED_MENU_NAMES[16:22],
        *EXPECTED_MENU_NAMES[22:33],
    ]
    if len(record_names) != 39:
        raise SystemExit("職業名稱映射數量錯誤")

    records = [records_raw[i * 16:(i + 1) * 16] for i in range(39)]
    if any(r[7] != 0 or r[15] != 0 for r in records):
        raise SystemExit("職業記錄保留 byte 改變")

    groups = []
    for g in GROUPS:
        professions = []
        for index in range(g["first"], g["last"] + 1):
            r = records[index]
            professions.append({
                "record_index": index,
                "name": record_names[index],
                "dice_codes": list(r[0:7]),
                "internal_bases": list(r[8:15]),
            })
        groups.append({"race": g["race"], "secondary": g["secondary"],
                       "professions": professions})

    data = {
        "schema": "kol-character-creation-data-v1",
        "tool": "tools/extract_character_data.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": ROOT, "record_item": ITEM_INDEX, "text_item": 1,
                           "record_sha256": sha(records_raw),
                           "text_sha256": sha(text_raw)},
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "dice_sides": DICE_SIDES,
        "primary_attributes": PRIMARY_ATTRIBUTES,
        "derived_attributes": DERIVED_ATTRIBUTES,
        "roll_rule": "每項內部結果為 internal_base + (dice_code & 0x0f) 次 "
                     "1..dice_sides[dice_code >> 4] 的擲骰總和",
        "derived_rule": "兩個來源屬性的內部無號整數平均，向下取整",
        "display_scale": {"numerator": 100, "denominator": 150,
                          "rule": "畫面值為 floor(內部值 × 100 / 150)"},
        "races": [{"value": i, "name": n,
                   "has_secondary": i in (0, 1)} for i, n in enumerate(EXPECTED_RACES)],
        "secondary": [{"value": i, "name": n} for i, n in enumerate(EXPECTED_SECONDARY)],
        "profession_groups": groups,
    }

    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
    os.replace(tmp, out)
    print(json.dumps({"groups": len(groups),
                      "professions": sum(len(g["professions"]) for g in groups),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
