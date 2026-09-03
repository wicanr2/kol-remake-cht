#!/usr/bin/env python3
"""野外怪物與遭遇機率的 typed data（RE-078）。

`sub_2A4E5` 擲遭遇時讀兩處：

  payload4[季節 × 8 + 地形]（夜裡再加 0x20）  遭遇機率；程式除以 2 之後才當百分比
  payload4[0x40 + 怪物]                      怪物出沒的地形位元遮罩
  payload4[0x60 + 怪物]                      怪物出沒的季節位元遮罩

遮罩的位元順序來自 KOLTOWN 的 DGROUP：地形是 `0x1312` 的 `80 40 20 10 08 04 02 01`
（bit 7 是地形 0），季節是 `0x1316` 的 `08 04 02 01`（bit 3 是冬）。

怪物本身是 `TOWN0` 的 payload 5…36，一筆 512 byte，第 0 個 byte 是名稱長度、
名稱從第 1 個 byte 起（`sub_2A4E5:0x2A64D` 把它讀進 `byte_39836`，
畫的時候 `sub_1BA60(0xB967, 長度)`）。名稱之後 +0x11 起
是六個屬性（力量、快捷、洞察、體型、健康、智力），戰鬥的先攻與昏迷門檻
從這裡取（RE-080）。

  python3 tools/extract_monsters.py --out assets/original-derived-data/monsters.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_town_buildings import good  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
TABLE_PAYLOAD, FIRST_MONSTER_PAYLOAD, MONSTER_COUNT = 4, 5, 32
TERRAIN_MASKS, SEASON_MASKS = 0x40, 0x60
NIGHT_OFFSET, SEASON_COUNT, TERRAIN_COUNT = 0x20, 4, 8
NAME_LIMIT = 24
# 怪物記錄的屬性欄位（RE-080）。名稱欄固定 16 byte，屬性從 +0x11 起：
# `sub_2215A` 把 +0x11…+0x16 搬進參戰者記錄的 +0x08…+0x0C，
# `sub_21AC4` 另外用 +0x15 算昏迷門檻。
STAT_OFFSETS = {"strength": 0x11, "quickness": 0x12, "foresight": 0x13,
                "size": 0x14, "health": 0x15, "intellect": 0x16}
# 戰鬥要用的其餘欄位（RE-080）：+0x17 起是攻防技能（大端 16-bit 各一組），
# +0x20 是手上的武器，+0x30／+0x40／+0x50 是三件護甲，+0x60 是盾牌，
# 每一格 16 byte，版面與商店的商品記錄相同。
SKILL_OFFSET = 0x17
WEAPON_OFFSET = 0x20
ARMOUR_OFFSETS = (0x30, 0x40, 0x50)
SHIELD_OFFSET = 0x60
TERRAIN_NAMES = ("DESERT", "RIVER", "SWAMP", "HILL",
                 "ROAD", "FOREST", "SHORE", "PLAINS")
SEASON_NAMES = ("WINTER", "SPRING", "SUMMER", "AUTUMN")
# 月份 → 季節（KOLTOWN DGROUP+0x1332）。
MONTH_SEASON = (0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 0)


def bits(mask: int, count: int, names: tuple[str, ...]) -> list[str]:
    """位元遮罩攤成名稱清單。bit 7 是第 0 項，依 DGROUP 的遮罩表往下數。"""
    return [names[i] for i in range(count) if mask & (0x80 >> i)]


def item(record: bytes, offset: int) -> dict | None:
    """怪物身上的一格裝備；空的回 None。"""
    block = record[offset:offset + 16]
    if len(block) < 16 or not any(block):
        return None
    out = good(block, 0)
    out["offset"] = offset
    return out


def monster_name(record: bytes) -> str:
    length = record[0]
    if not 0 < length <= NAME_LIMIT:
        raise SystemExit(f"怪物名稱長度不合理：{length}")
    raw = record[1:1 + length]
    if any(b < 0x20 or b > 0x7E for b in raw):
        raise SystemExit(f"怪物名稱有不可列印的 byte：{list(raw)}")
    return raw.decode("ascii")


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    out = pathlib.Path(args.out)
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(out.as_posix()).parents:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")
    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(zip_path) as zf:
        _, parts = payloads(read_volumes(zf, "TOWN0"), "TOWN0")
        table = lzw_decode(parts[TABLE_PAYLOAD])
        records = [lzw_decode(parts[FIRST_MONSTER_PAYLOAD + i])
                   for i in range(MONSTER_COUNT)]

    monsters = []
    for i, rec in enumerate(records):
        terrain, season = table[TERRAIN_MASKS + i], table[SEASON_MASKS + i]
        monsters.append({
            "index": i,
            "payload": FIRST_MONSTER_PAYLOAD + i,
            "name": monster_name(rec),
            "terrain_mask": terrain,
            "terrain": bits(terrain, TERRAIN_COUNT, TERRAIN_NAMES),
            "season_mask": season,
            "season": bits(season << 4, SEASON_COUNT, SEASON_NAMES),
            **{name: rec[off] for name, off in STAT_OFFSETS.items()},
            "attack_skill": rec[SKILL_OFFSET] << 8 | rec[SKILL_OFFSET + 1],
            "parry_skill": rec[SKILL_OFFSET + 2] << 8 | rec[SKILL_OFFSET + 3],
            "weapon": item(rec, WEAPON_OFFSET),
            "armour_weight": sum(
                rec[off + 10] for off in ARMOUR_OFFSETS if any(rec[off:off + 16])),
            "shield": item(rec, SHIELD_OFFSET),
            "record_sha256": hashlib.sha256(rec).hexdigest(),
        })

    chance = []
    for season in range(SEASON_COUNT):
        chance.append({
            "season": SEASON_NAMES[season],
            "day": [table[season * TERRAIN_COUNT + t] for t in range(TERRAIN_COUNT)],
            "night": [table[NIGHT_OFFSET + season * TERRAIN_COUNT + t]
                      for t in range(TERRAIN_COUNT)],
        })

    doc = {
        "schema": "kol-monster-data-v1",
        "tool": "tools/extract_monsters.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWN0",
                           "table_payload": TABLE_PAYLOAD,
                           "monster_payloads": [FIRST_MONSTER_PAYLOAD,
                                                FIRST_MONSTER_PAYLOAD + MONSTER_COUNT - 1]},
        "evidence": "RE-078",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "chance": "payload4[季節 × 8 + 地形]；夜裡（天候第 5 組）再加 0x20；"
                      "程式除以 2 之後才當百分比",
            "terrain_mask": f"payload4[0x{TERRAIN_MASKS:X} + 怪物]",
            "season_mask": f"payload4[0x{SEASON_MASKS:X} + 怪物]",
            "monster_record": "payload 5 + 怪物：第 0 byte 是名稱長度，名稱接在後面；"
                              "其餘欄位要等戰鬥解出來才寫",
        },
        "names": {"terrain": list(TERRAIN_NAMES), "season": list(SEASON_NAMES),
                  "month_season": list(MONTH_SEASON)},
        "table_sha256": hashlib.sha256(table).hexdigest(),
        "encounter_chance": chance,
        "monsters": monsters,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n",
                   encoding="utf-8")
    print(json.dumps({"monsters": len(monsters), "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
