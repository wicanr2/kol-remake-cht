#!/usr/bin/env python3
"""驗證 `MAGIC.0` 的字典解對了：SUBCLASS 那幾組就是怪物表裡的怪（RE-128）。

`MAGIC.0` 是一串沒有標頭的位址表加字串，「解得開」不等於「解對了」。這一支拿另一份
typed data 對答案：法術的 CLASS 分成九類，而 SUBCLASS 依 CLASS 分成十六組——
Elemental／Giant／Legendary／Undead 那四組列出來的名字，應該一個不差地出現在
`monsters.json` 的三十二隻裡。

  tools/run_python_tool.sh tools/verify_magic_dictionary.py \
      --magic assets/original-derived-data/magic.json \
      --monsters assets/original-derived-data/monsters.json
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

# 原版自己的拼字不一致：`MAGIC.0` 寫 Salamandar、`TOWN0` 的怪物表寫 SALAMANDER。
# 兩邊都是原版資料，不是解碼錯誤，所以列成已知例外而不是失敗。
SPELLING = {"salamandar": "salamander"}
# 這幾個不是怪物名，是「沒有」與「全部」。
PLACEHOLDERS = {"none", "all"}


def monster_names(doc) -> set[str]:
    out: set[str] = set()

    def walk(node) -> None:
        if isinstance(node, dict):
            name = node.get("name")
            # 怪物名在原版是全大寫；同一份檔案裡的武器名是混合大小寫。
            if isinstance(name, str) and name and name == name.upper():
                out.add(name.lower())
            for value in node.values():
                walk(value)
        elif isinstance(node, list):
            for value in node:
                walk(value)

    walk(doc)
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--magic", required=True)
    ap.add_argument("--monsters", required=True)
    args = ap.parse_args()

    magic = json.loads(pathlib.Path(args.magic).read_text(encoding="utf-8"))
    monsters = monster_names(json.loads(
        pathlib.Path(args.monsters).read_text(encoding="utf-8")))

    bad: list[str] = []
    checked = 0
    for group, rows in enumerate(magic["subclass"]):
        for name in rows:
            key = name.lower()
            if key in PLACEHOLDERS:
                continue
            checked += 1
            if SPELLING.get(key, key) not in monsters:
                bad.append(f"第 {group} 組的 {name!r} 不在怪物表裡")

    # CLASS 那九筆與 STAT 那九筆的第一項都要是 None——原版靠它當「沒有」。
    for key in ("class", "stat"):
        if not magic[key] or magic[key][0] != "None":
            bad.append(f"{key} 的第一項是 {magic[key][:1]}，應該是 None")
    # 骰式那一串要看得出來是 `低-高`。
    for row in magic["dice"]:
        lo, _, hi = row.partition("-")
        if not lo.isdigit() or not hi.isdigit():
            bad.append(f"骰式 {row!r} 不是「低-高」")

    for line in bad:
        print(line, file=sys.stderr)
    print(f"對過 {checked} 個 SUBCLASS 名稱與 {len(magic['dice'])} 個骰式，"
          f"{len(bad)} 處不符")
    return 1 if bad else 0


raise SystemExit(main())
