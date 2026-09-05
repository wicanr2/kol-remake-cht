#!/usr/bin/env python3
"""把一份存檔改成「站在某扇門口、身上有錢」的觀察輸入。

原版名冊的角色都住在城 7 的半途之家，錢包只有兩三百枚金幣——走不到別座城的
馬廄（RE-119），也付不起入會費與法術（RE-121）。這支工具把開場位置換到指定的
門口並把發言者的錢包補到指定金額，好讓那幾條路徑走得完。

預設是城 1 的 Holcroft Stables（門 8，在 41, 54）。城 7 的 The Hostel 是
`--town 7 --x 59 --y 28`。

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
    p.add_argument("--town", type=int, default=TOWN, help="城鎮編號")
    p.add_argument("--x", type=int, default=DOOR_X, help="門的 X")
    p.add_argument("--y", type=int, default=DOOR_Y, help="門的 Y")
    p.add_argument("--entry", type=int, default=0,
                   help="入城點編號；鏡頭的夾限用它，選錯玩家會被推回可視範圍")
    args = p.parse_args()

    doc = json.loads(pathlib.Path(args.src).read_text(encoding="utf-8"))
    doc["location"] = {"kind": "town", "town": args.town,
                       "entry_point": args.entry,
                       "x": args.x, "y": args.y + 1}
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
        "town": args.town, "x": args.x, "y": args.y + 1, "gold": args.gold,
        "out": args.out,
        "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
    }, ensure_ascii=False))
    return 0


sys.exit(main())
