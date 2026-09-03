#!/usr/bin/env python3
"""把原版角色檔匯入成 remake 的存檔（SPEC-006）。

原版的存檔是一人一筆的 `CHARDATA`（16 × 768 byte）加上三份同槽位的圖檔。
remake 用的是自己的 JSON 格式，一檔一隊。這支工具做的就是這一次轉換：

  * 原版檔案**只讀不寫**。remake 在任何情況下都不回寫 `CHARDATA`。
  * 只搬有 RE 證據的欄位（SPEC-006 的欄位表）。沒有證據的位移不猜、不補零。
  * 頭像逐像素與 80 張模板比對；對不上就失敗，不會默默留白。

  python3 tools/import_original_characters.py --zip kol.zip --repo . \\
      --names Theodoric Olear --out saves/slot1.json
"""

from __future__ import annotations

import argparse
import datetime
import hashlib
import json
import os
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_character_roster import (  # noqa: E402
    BUILDING_OFFSET, CHARDATA_NAME, CHARPICS_NAME, FEMALE_OFFSET, GOLD_BYTES,
    GOLD_OFFSET, NAME_MAX, NAME_OFFSET, PLACE_BYTES, PLACE_OFFSET,
    PORTRAIT_BYTES, SLOT_BYTES, SLOT_COUNT, TEMPLATE_COUNT, TEMPLATE_DIR,
    TOWN_BYTES, TOWN_OFFSET, WOUNDS_OFFSET, cstring)
from png_io import read_png  # noqa: E402
from probe_extract_comb_png import PALETTE  # noqa: E402
from probe_extract_fixed_charpics_png import decode_record  # noqa: E402

SAVE_SCHEMA = "kol-save-v1"
MAX_PARTY = 6
# 城內座標與入城點在角色記錄的位移（`sub_14DB1:0x1512E` 起）。
TOWN_NUMBER_OFFSET = 0x97
ENTRY_POINT_OFFSET = 0x13
TOWN_X_OFFSET, TOWN_Y_OFFSET = 0x99, 0x9C


def load_templates(repo: pathlib.Path) -> dict[bytes, int]:
    templates: dict[bytes, int] = {}
    for index in range(TEMPLATE_COUNT):
        path = repo / TEMPLATE_DIR / f"charpics-{index:03d}.png"
        templates[read_png(path).rgb()] = index
    return templates


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", help="原版 ZIP（唯讀）")
    p.add_argument("--chardata", help="直接指定 CHARDATA 檔案")
    p.add_argument("--charpics", help="直接指定 CHARPICS 檔案")
    p.add_argument("--repo", default=".")
    p.add_argument("--names", nargs="+", required=True,
                   help=f"要帶進隊伍的角色名，最多 {MAX_PARTY} 位；第一位是發言者")
    p.add_argument("--out", required=True)
    p.add_argument("--saved-at", default=None,
                   help="存檔時間戳；不給就用現在的 UTC 時間")
    p.add_argument("--location", default=None,
                   help="覆寫開場位置，格式 `城鎮,入城點,X,Y`；"
                        "不給就用角色記錄裡留宿的那一間")
    args = p.parse_args()

    if len(args.names) > MAX_PARTY:
        raise SystemExit(f"隊伍最多 {MAX_PARTY} 位，給了 {len(args.names)} 位")
    repo = pathlib.Path(args.repo).resolve()

    if args.zip:
        zip_path = pathlib.Path(args.zip)
        source = {"name": zip_path.name,
                  "sha256": hashlib.sha256(zip_path.read_bytes()).hexdigest()}
        with zipfile.ZipFile(zip_path) as archive:
            data = archive.read(CHARDATA_NAME)
            portraits = archive.read(CHARPICS_NAME)
    elif args.chardata and args.charpics:
        data = pathlib.Path(args.chardata).read_bytes()
        portraits = pathlib.Path(args.charpics).read_bytes()
        source = {"name": pathlib.Path(args.chardata).name,
                  "sha256": hashlib.sha256(data).hexdigest()}
    else:
        raise SystemExit("要嘛給 --zip，要嘛同時給 --chardata 與 --charpics")

    if len(data) != SLOT_BYTES * SLOT_COUNT:
        raise SystemExit(f"CHARDATA 大小為 {len(data)}，不是 {SLOT_BYTES * SLOT_COUNT}")
    if len(portraits) != PORTRAIT_BYTES * SLOT_COUNT:
        raise SystemExit(f"CHARPICS 大小為 {len(portraits)}，"
                         f"不是 {PORTRAIT_BYTES * SLOT_COUNT}")

    templates = load_templates(repo)
    slots: dict[str, dict] = {}
    for index in range(SLOT_COUNT):
        base = index * SLOT_BYTES
        name = cstring(data[base + NAME_OFFSET:base + NAME_OFFSET + NAME_MAX])
        if not name:
            continue
        record = portraits[index * PORTRAIT_BYTES:(index + 1) * PORTRAIT_BYTES]
        rgb = b"".join(bytes(PALETTE[v]) for v in decode_record(record))
        portrait = templates.get(rgb)
        if portrait is None:
            raise SystemExit(f"槽位 {index}（{name}）的頭像不是任何一張模板；"
                             "頭像被編輯過就沒有可搬的圖號，這裡不猜")
        slots[name] = {
            "name": name,
            "portrait": portrait,
            "female": data[base + FEMALE_OFFSET] != 0,
            "wounds": data[base + WOUNDS_OFFSET],
            "gold": int.from_bytes(
                data[base + GOLD_OFFSET:base + GOLD_OFFSET + GOLD_BYTES], "big"),
            "place": data[base + PLACE_OFFSET:
                          base + PLACE_OFFSET + PLACE_BYTES].decode("ascii"),
            "town_code": data[base + TOWN_OFFSET:
                              base + TOWN_OFFSET + TOWN_BYTES].decode("ascii"),
            "building_id": int.from_bytes(
                data[base + BUILDING_OFFSET:base + BUILDING_OFFSET + 2], "big"),
            "renown": 0,
            "spellcaster": False,
            "speaker": False,
            "_slot": index,
        }

    party = []
    for i, name in enumerate(args.names):
        if name not in slots:
            raise SystemExit(f"角色檔裡沒有 {name}；有的是 {sorted(slots)}")
        member = dict(slots[name])
        slot = member.pop("_slot")
        member["speaker"] = i == 0
        party.append(member)
        if i == 0:
            base = slot * SLOT_BYTES
            location = {"kind": "town",
                        "town": data[base + TOWN_NUMBER_OFFSET],
                        "entry_point": data[base + ENTRY_POINT_OFFSET],
                        "x": data[base + TOWN_X_OFFSET],
                        "y": data[base + TOWN_Y_OFFSET]}

    if args.location:
        try:
            town, entry, x, y = (int(v) for v in args.location.split(","))
        except ValueError:
            raise SystemExit("--location 的格式是 `城鎮,入城點,X,Y`")
        location = {"kind": "town", "town": town, "entry_point": entry, "x": x, "y": y}

    saved_at = args.saved_at or datetime.datetime.now(
        datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    doc = {
        "schema": SAVE_SCHEMA,
        "saved_at": saved_at,
        "origin": {"kind": "imported", "source_sha256": source["sha256"],
                   "tool": "tools/import_original_characters.py",
                   "note": f"由 {source['name']} 的原版角色檔匯入；原檔未被改寫"},
        "settings": {"faithful_inn_stay": False},
        # 野外座標與世界時間在原版記錄裡都還沒定位過，匯入的存檔一律從城鎮開始，
        # 沒有證據的欄位不猜也不補（SPEC-006）。
        "world": {"x": 0, "y": 0, "sub_x": 3, "sub_y": 3},
        "location": location,
        "party": party,
    }

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n",
                   encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"party": [m["name"] for m in party],
                      "source_sha256": source["sha256"],
                      "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
