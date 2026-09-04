#!/usr/bin/env python3
"""把一份存檔改成「有人該去競技場升階」的觀察輸入。

競技場的候選條件是「帳面位階與實得位階不同、而且帳面位階還沒到 24」（RE-108）。
原版名冊裡的角色兩個位階都是齊的——他們是遊戲開始前的狀態，還沒打過任何一場——
所以照原樣走進競技場只會聽到「None of you are ready」。這支工具把指定那幾名的
帳面位階降下來，好讓「選誰上場 → 確定 → 開打 → 升一階」那一條走得完。

改的是存檔本身，不是遊戲：產出的檔案仍是 `kol-save-v1`，遊戲照常讀。

  python3 tools/make_arena_save.py --in workplace/saves/arena.json \\
      --out workplace/observe/arena-ready.json --members 0 --rank 0
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--in", dest="src", required=True)
    p.add_argument("--out", required=True)
    p.add_argument("--members", type=int, nargs="+", default=[0],
                   help="要降位階的隊員索引")
    p.add_argument("--rank", type=int, default=0, help="降到多少")
    args = p.parse_args()

    doc = json.loads(pathlib.Path(args.src).read_text(encoding="utf-8"))
    changed = []
    for i in args.members:
        if i < 0 or i >= len(doc["party"]):
            raise SystemExit(f"隊伍裡沒有第 {i} 名")
        doc["party"][i]["rank"] = args.rank
        changed.append(doc["party"][i]["name"])

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(json.dumps({
        "source": args.src,
        "source_sha256": hashlib.sha256(pathlib.Path(args.src).read_bytes()).hexdigest(),
        "members": changed, "rank": args.rank,
        "out": args.out,
        "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
    }, ensure_ascii=False))
    return 0


sys.exit(main())
