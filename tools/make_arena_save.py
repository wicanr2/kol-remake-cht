#!/usr/bin/env python3
"""把一份存檔改成「有人該去競技場升階」的觀察輸入。

競技場的候選條件是「帳面位階與實得位階不同、而且帳面位階還沒到 24」（RE-108）。
原版名冊裡的角色兩個位階都是齊的——他們是遊戲開始前的狀態，還沒打過任何一場——
所以照原樣走進競技場只會聽到「None of you are ready」。這支工具把指定那幾名的
帳面位階降下來，好讓「選誰上場 → 確定 → 開打 → 升一階」那一條走得完。

只產生獨立觀察輸入，不修改來源。v1／v2 均保留格式；v2 依隊伍索引找到名冊角色。
這不是正常遊玩取得升階資格的證據，也不保證指定的位階與實得位階一定不同。

  python3 tools/make_arena_save.py --in workplace/saves/arena.json \\
      --out workplace/observe/arena-ready.json --members 0 --rank 0
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
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

    src, out = pathlib.Path(args.src), pathlib.Path(args.out)
    if src.resolve() == out.resolve() or (out.exists() and src.samefile(out)):
        raise SystemExit("輸出不可覆寫來源存檔")
    if not 0 <= args.rank <= 24:
        raise SystemExit("位階必須介於 0 至 24")
    source_bytes = src.read_bytes()
    doc = json.loads(source_bytes)
    if not isinstance(doc, dict) or doc.get("schema") not in ("kol-save-v1", "kol-save-v2"):
        raise SystemExit("只接受 kol-save-v1 或 kol-save-v2")
    party = doc.get("party")
    if not isinstance(party, list) or not party:
        raise SystemExit("來源存檔沒有活動隊伍")
    changed = []
    for i in dict.fromkeys(args.members):
        if i < 0 or i >= len(party):
            raise SystemExit(f"隊伍裡沒有第 {i} 名")
        if doc["schema"] == "kol-save-v2":
            characters = doc.get("characters")
            slot = party[i]
            if (not isinstance(characters, list) or type(slot) is not int or
                    not 0 <= slot < len(characters) or not isinstance(characters[slot], dict)):
                raise SystemExit(f"隊員 {i} 的名冊索引無效")
            member = characters[slot]
            sheet = member.get("sheet")
            if not isinstance(sheet, dict):
                raise SystemExit(f"隊員 {i} 缺少完整資料頁")
            sheet["rank"] = args.rank
        else:
            member = party[i]
            if not isinstance(member, dict):
                raise SystemExit(f"隊員 {i} 格式錯誤")
            member["rank"] = args.rank
        changed.append(member.get("name", ""))

    out.parent.mkdir(parents=True, exist_ok=True)
    if out.parent.stat().st_uid != os.getuid() or (out.exists() and out.stat().st_uid != os.getuid()):
        raise SystemExit("輸出檔與目錄必須由目前使用者擁有")
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(json.dumps({
        "source": args.src,
        "source_sha256": hashlib.sha256(source_bytes).hexdigest(),
        "schema": doc["schema"],
        "evidence": "受控修改位階的觀察輸入，不是正常玩家路徑收據",
        "members": changed, "rank": args.rank,
        "out": args.out,
        "out_sha256": hashlib.sha256(out.read_bytes()).hexdigest(),
    }, ensure_ascii=False))
    return 0


sys.exit(main())
