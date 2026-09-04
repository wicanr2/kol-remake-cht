#!/usr/bin/env python3
"""從一份存檔造出「最後一件任務就差回報」的觀察輸入。

結局的觸發條件是角色記錄 `0x1DD` 那一格等於 0x17——已回報的任務數滿 23 件
（RE-106／RE-107）。remake 目前做得到「打完標成完成」與「交到指定建築算回報」，
**接任務那一側（`sub_14564`）還沒實作**，所以沒有辦法用正常路徑把那一格累積上去。
這支工具把一份存檔改成「已回報 22 件、最後一件已完成、東西就在身上」，
好讓最後那一步——走進集會所交出去、看到結局——可以用正常按鍵走完。

改的是存檔本身，不是遊戲：產出的檔案仍是 `kol-save-v1`，遊戲照常讀。

  python3 tools/make_final_quest_save.py --in workplace/saves/full-chain.json \\
      --out workplace/observe/final-quest.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

# 最終任務那一筆：TOWN7 的第 14 扇門（The Assembly Building），要 Shackles。
FINAL_SITE = 4
FINAL_TOWN, FINAL_DOOR = 7, 14
CARRIED_SLOTS = 17
CARRIED_STRIDE = 16
# 已回報 22 件；交出去那一刻加一，正好是觸發結局的 0x17。
REPORTS_BEFORE = 0x16
QUEST_DONE = 2


def quest_byte(site: int) -> tuple[int, int]:
    """任務旗標的位置：第幾個 byte、除數是多少（一個 byte 四個 2-bit 欄位）。"""
    return site // 4, 4 ** (site % 4)


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--in", dest="src", required=True)
    p.add_argument("--out", required=True)
    p.add_argument("--buildings", default="assets/original-derived-data/town-buildings.json")
    p.add_argument("--member", type=int, default=0, help="要改哪一名隊員（索引）")
    p.add_argument("--flag", type=int, default=QUEST_DONE,
                   help="最終任務的旗標：0 沒接、1 接了、2 完成、3 回報過")
    p.add_argument("--reports", type=int, default=REPORTS_BEFORE,
                   help="已回報的任務數；0x17（23）是結局的判準")
    args = p.parse_args()

    doc = json.loads(pathlib.Path(args.src).read_text(encoding="utf-8"))
    buildings = json.loads(pathlib.Path(args.buildings).read_text(encoding="utf-8"))

    wants = None
    for town in buildings.get("towns", []):
        if town.get("town") != FINAL_TOWN:
            continue
        for door in town.get("doors", []):
            if door.get("index") == FINAL_DOOR and door.get("quest_final"):
                wants = door.get("quest_wants_bytes")
    if not wants or len(wants) != CARRIED_STRIDE * 2:
        raise SystemExit("找不到最終任務那一扇門要的東西")

    m = doc["party"][args.member]
    flags = list(m.get("quests") or [0] * 6)
    while len(flags) < 6:
        flags.append(0)
    index, div = quest_byte(FINAL_SITE)
    # 把最終任務那一欄設成「已完成」，其餘欄不動。
    field = (flags[index] // div) % 4
    flags[index] += (args.flag - field) * div
    m["quests"] = flags

    progress = list(m.get("progress") or [0] * 5)
    while len(progress) < 5:
        progress.append(0)
    progress[0] = args.reports
    m["progress"] = progress

    carried = list(m.get("carried") or [])
    while len(carried) < CARRIED_SLOTS:
        carried.append("0" * (CARRIED_STRIDE * 2))
    carried[0] = wants
    m["carried"] = carried[:CARRIED_SLOTS]

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    body = json.dumps(doc, ensure_ascii=False, indent=1) + "\n"
    out.write_text(body, encoding="utf-8")
    print(json.dumps({
        "source": args.src,
        "source_sha256": hashlib.sha256(pathlib.Path(args.src).read_bytes()).hexdigest(),
        "member": m["name"],
        "quests": flags, "progress": progress, "wants": wants,
        "out": args.out,
        "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
    }, ensure_ascii=False))
    return 0


sys.exit(main())
