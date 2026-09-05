#!/usr/bin/env python3
"""24 個任務地點的 typed data（RE-079）。

`sub_29004:0x29484` 在世界座標變動之後比對 `payload3[0x323 + i]`／`[0x33B + i]`，
命中就呼叫 `sub_2B0FF(i)`。那一支先看隊上有沒有人帶著這個任務
（角色 `[0x1E2 + i / 4]` 除以 `[0x7C + (i & 3)]` 之後低兩位是 1），
再讀 `TOWN0` 的 payload `0x25 + i`：第 0 個 byte 是怪物編號、第 1 個是等級、
第 2 個是數量（`0x2B19B`…`0x2B1AD`）。

  python3 tools/extract_quest_sites.py --out assets/original-derived-data/quest-sites.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_town_buildings import good, packed_name  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from extract_town_text import render  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
TABLE_PAYLOAD, SITE_PAYLOAD_BASE, SITE_COUNT = 3, 0x25, 24
SITE_X, SITE_Y = 0x323, 0x33B
FIRST_MONSTER_PAYLOAD, MONSTER_COUNT = 5, 32
# 任務旗標：一個 byte 塞四個 2-bit 欄位，除數表在 KOLTOWN 的 DGROUP+0x7C。
QUEST_FIELD_BASE, QUEST_DIVISORS = 0x1E2, (1, 4, 16, 64)
# 站位：記錄 +3 起 36 byte 複製到 DGROUP+0x2E0C，戰鬥再由那裡取
# （`sub_2B0FF:0x2B204`、`sub_1E3A8:0x1E870`／`0x1E7C5`）。
PLACEMENT_OFFSET, MONSTER_SLOTS, PARTY_SLOTS = 3, 12, 6
# 敘述：記錄 0x80 起八列，一列 16 byte——第 0 個 byte 是字數，
# 之後從第 8 個 bit 起每 5 個 bit 一個字（`sub_2B0FF:0x2B23F` 的 `sub_1B94F`）。
TEXT_OFFSET, TEXT_LINES, TEXT_STRIDE, TEXT_MAX = 0x80, 8, 0x10, 24
# 任務要的那一件東西：記錄 +0x29 起十六個 byte，版面與商店的商品記錄相同（RE-113）。
# `sub_2B0FF:0x2B1C4` 把它複製到 DGROUP `0xB956`，戰鬥中 `sub_2886C` 拿它跟角色
# 手上那件逐 byte 比——帶著它逃出戰場就算任務完成（RE-103）。
QUEST_ITEM_OFFSET, QUEST_ITEM_BYTES = 0x29, 0x10
# 任務物品放在戰場哪一格：`+0x27` 是列、`+0x28` 是欄。`sub_2B0FF:0x2B1F0` 在
# `byte_39826`（＝任務物品那十六個 byte 的第 0 格）非零時，把這兩個 byte 抄進
# `byte_30C6E`／`byte_30A7E`——而那兩個是**戰場物品第 0 件的列與欄**，各自是一張
# 陣列的第 0 格（`0x2D9E + i`／`0x2BAE + i`）。讀的地方是 `sub_223C0:0x226C5`
# 起那一段：掃全部物品，欄與列都對上行動者站的那一格，才把地圖磚換成 `0x27`
#（RE-128 那一輪解出，更正 RE-113 的「只有寫沒有讀」）。戰場是 64 × 64，
# 二十四個地點的值全部落在 0…63。
QUEST_ITEM_SPOT_OFFSET = 0x27


def quest_item(raw: bytes) -> dict:
    """任務要的那一件東西：十六個 byte 的物品記錄加它擺在戰場哪一格。

    記錄的版面與商店的商品相同（RE-113），所以直接走 `good`——撿起來的人要拿它
    去打，只留名字的話傷害骰、負重與種類全都沒有。"""
    rec = bytes(raw[QUEST_ITEM_OFFSET:QUEST_ITEM_OFFSET + QUEST_ITEM_BYTES])
    item = good(rec, 0)
    item.pop("index", None)
    item["spot"] = {"y": raw[QUEST_ITEM_SPOT_OFFSET],
                    "x": raw[QUEST_ITEM_SPOT_OFFSET + 1]}
    return item


def line_values(data: bytes, base: int, count: int) -> list[int]:
    out = []
    for i in range(count):
        bit = base * 8 + 8 + i * 5
        j = bit >> 3
        word = (data[j] << 8) | (data[j + 1] if j + 1 < len(data) else 0)
        out.append((word >> (11 - (bit & 7))) & 0x1F)
    return out


def narrative(record: bytes) -> list[str]:
    lines = []
    for i in range(TEXT_LINES):
        base = TEXT_OFFSET + i * TEXT_STRIDE
        count = record[base]
        if not 0 < count <= TEXT_MAX:
            raise SystemExit(f"敘述第 {i} 列的字數不合理：{count}")
        text, unknown = render(line_values(record, base, count))
        if unknown:
            raise SystemExit(f"敘述第 {i} 列有未定的跳脫序列：{unknown}")
        lines.append(text)
    return lines


def monster_name(record: bytes) -> str:
    length = record[0]
    if not 0 < length <= 24:
        raise SystemExit(f"怪物名稱長度不合理：{length}")
    return record[1:1 + length].decode("ascii")


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
        sites_raw = [lzw_decode(parts[SITE_PAYLOAD_BASE + i]) for i in range(SITE_COUNT)]
        names = [monster_name(lzw_decode(parts[FIRST_MONSTER_PAYLOAD + i]))
                 for i in range(MONSTER_COUNT)]

    sites = []
    for i, raw in enumerate(sites_raw):
        monster = raw[0]
        if monster >= MONSTER_COUNT:
            raise SystemExit(f"任務地點 {i} 的怪物編號 {monster} 超出範圍")
        sites.append({
            "index": i,
            "x": table[SITE_X + i], "y": table[SITE_Y + i],
            "monster": monster, "monster_name": names[monster],
            "level": raw[1], "count": raw[2],
            "quest_byte": QUEST_FIELD_BASE + i // 4,
            "quest_divisor": QUEST_DIVISORS[i & 3],
            "payload": SITE_PAYLOAD_BASE + i,
            "payload_sha256": hashlib.sha256(raw).hexdigest(),
            "monster_y": list(raw[PLACEMENT_OFFSET:PLACEMENT_OFFSET + MONSTER_SLOTS]),
            "monster_x": list(raw[PLACEMENT_OFFSET + MONSTER_SLOTS:
                                  PLACEMENT_OFFSET + 2 * MONSTER_SLOTS]),
            "party_y": list(raw[PLACEMENT_OFFSET + 2 * MONSTER_SLOTS:
                                PLACEMENT_OFFSET + 2 * MONSTER_SLOTS + PARTY_SLOTS]),
            "party_x": list(raw[PLACEMENT_OFFSET + 2 * MONSTER_SLOTS + PARTY_SLOTS:
                                PLACEMENT_OFFSET + 2 * MONSTER_SLOTS + 2 * PARTY_SLOTS]),
            "narrative": narrative(raw),
            "quest_item": quest_item(raw),
        })

    doc = {
        "schema": "kol-quest-site-data-v1",
        "tool": "tools/extract_quest_sites.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWN0",
                           "table_payload": TABLE_PAYLOAD,
                           "site_payloads": [SITE_PAYLOAD_BASE,
                                             SITE_PAYLOAD_BASE + SITE_COUNT - 1]},
        "evidence": "RE-079",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "coordinates": f"payload{TABLE_PAYLOAD}[0x{SITE_X:X} + i]／[0x{SITE_Y:X} + i]",
            "site_record": f"payload {SITE_PAYLOAD_BASE} + i：怪物、等級、數量",
            "quest_flag": "角色 [0x1E2 + i / 4] 除以 (1, 4, 16, 64)[i & 3] 之後低兩位是 1",
            "placement": "記錄 +3 起 36 byte：12 敵人 Y、12 敵人 X、6 隊員 Y、6 隊員 X",
            "narrative": "記錄 0x80 起八列，一列 16 byte：第 0 個 byte 是字數，"
                         "之後從第 8 個 bit 起每 5 個 bit 一個字",
        },
        "sites": sites,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"sites": len(sites), "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
