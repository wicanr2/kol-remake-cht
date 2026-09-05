#!/usr/bin/env python3
"""把一份存檔改成「站在馬廄門口、身上有錢」的觀察輸入。

八座城裡有五間馬廄（RE-119），一間都不在城 7——而原版名冊的角色都住在城 7 的
半途之家，錢也不夠買最便宜那一匹（200 金幣）。這支工具把開場位置換到城 1 的
Holcroft Stables 門口，並把發言者的錢包補到指定金額，好讓「進門 → TRADE →
挑一匹 → 付錢 → 資料頁看得到牠」那一條走得完。

改的是存檔本身，不是遊戲：產出的檔案仍是 `kol-save-v1`，遊戲照常讀。

  python3 tools/make_stable_save.py --in workplace/saves/quest.json \\
      --out workplace/saves/stable.json --gold 3000
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

# Holcroft Stables 是城 1 的第 8 扇門，門在 (41, 54)；玩家站在門的正下方。
TOWN, DOOR_X, DOOR_Y = 1, 41, 54


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--in", dest="src", required=True)
    p.add_argument("--out", required=True)
    p.add_argument("--gold", type=int, default=3000, help="發言者的錢包")
    args = p.parse_args()

    doc = json.loads(pathlib.Path(args.src).read_text(encoding="utf-8"))
    doc["location"] = {"kind": "town", "town": TOWN, "entry_point": 0,
                       "x": DOOR_X, "y": DOOR_Y + 1}
    for member in doc["party"]:
        if member.get("speaker"):
            member["gold"] = args.gold
            break
    else:
        doc["party"][0]["gold"] = args.gold

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(json.dumps({
        "source": args.src,
        "source_sha256": hashlib.sha256(pathlib.Path(args.src).read_bytes()).hexdigest(),
        "town": TOWN, "x": DOOR_X, "y": DOOR_Y + 1, "gold": args.gold,
        "out": args.out,
        "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
    }, ensure_ascii=False))
    return 0


sys.exit(main())
