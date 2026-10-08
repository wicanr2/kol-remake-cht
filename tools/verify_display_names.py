#!/usr/bin/env python3
"""核對顯示譯名 TSV 與怪物、裝備及任務資料的實際英文名稱。"""

from __future__ import annotations

import argparse
import csv
import json
import pathlib
import sys

HEADER = ["key", "en", "zh-Hant", "status", "source"]


def names(node):
    if isinstance(node, dict):
        for key, value in node.items():
            if key == "name" and isinstance(value, str) and value:
                yield value
            yield from names(value)
    elif isinstance(node, list):
        for value in node:
            yield from names(value)


def required_names(root: pathlib.Path) -> set[str]:
    monsters = json.loads((root / "monsters.json").read_text())["monsters"]
    items = json.loads((root / "items.json").read_text())["items"]
    towns = json.loads((root / "town-buildings.json").read_text())["towns"]
    roster = json.loads((root / "character-roster.json").read_text())["slots"]
    required = set(names(monsters))
    required.update(item["name"] for item in items if item["name"])
    # 初始持有物不一定出現在共用物品表或商店（SPEC-033，RE-170）。
    # 只掃有效角色的物品；玩家姓名與規則用法術名不屬於顯示譯名。
    for character in roster:
        if not character or not character["occupied"]:
            continue
        goods = [character["in_hand"], *character["equipment"],
                 *character["backpack"]]
        required.update(good["name"] for good in goods if good and good.get("name"))
    for town in towns:
        for door in town["doors"]:
            required.update(good["name"] for good in door.get("goods", [])
                            if good.get("name"))
            for field in ("quest_wants", "quest_gives"):
                if door.get(field):
                    required.add(door[field])
            # 教派名、頭銜與信物（RE-121／200）：對話、資料頁及物品明細。
            if door.get("order"):
                required.update(door["order"][f] for f in ("name", "title", "token") if door["order"].get(f))
            # 名冊右欄的所在地鍵：角色記錄存的是建築名前 10 個大寫字元（`placeKey`）。
            if door.get("name"):
                required.add(door["name"].upper()[:10])
    # MAGIC.0 的亞種名（RE-128）：法術注解與法術編輯器的 SUBCLASS 欄。
    magic = json.loads((root / "magic.json").read_text())
    for group in magic["subclass"]:
        required.update(name.upper() for name in group if name not in ("None", "All"))
    # 法術編輯器的 CLASS／STAT 選項與 SUBCLASS 的 All（RE-128，SPEC-035）。
    required.update(magic["class"])
    required.update(name for name in magic["stat"] if name != "None")
    required.add("All")
    world = json.loads((root / "world-map.json").read_text())
    required.update(town["code"] for town in world["towns"])
    return required


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--table", default="docs/i18n/display-names-zh-Hant.tsv")
    parser.add_argument("--source-root", default="assets/original-derived-data")
    args = parser.parse_args()
    problems = []
    seen_keys = set()
    english_to_chinese = {}
    with pathlib.Path(args.table).open(encoding="utf-8", newline="") as handle:
        rows = csv.reader(handle, delimiter="\t")
        if next(rows, None) != HEADER:
            problems.append("TSV 欄位不符")
        for line_number, row in enumerate(rows, 2):
            if len(row) != len(HEADER):
                problems.append(f"第 {line_number} 行欄數不是 5")
                continue
            key, en, zh, status, source = row
            if not key or key in seen_keys:
                problems.append(f"第 {line_number} 行 key 為空或重複：{key}")
            seen_keys.add(key)
            if not en or not zh or not source or status not in ("核准", "暫譯"):
                problems.append(f"第 {line_number} 行缺原名／譯名／來源或狀態不正確")
            if not any("\u3400" <= char <= "\u9fff" for char in zh):
                problems.append(f"第 {line_number} 行沒有中文字：{en}")
            previous = english_to_chinese.setdefault(en, zh)
            if previous != zh:
                problems.append(f"原名 {en} 有衝突譯名：{previous}／{zh}")
    required = required_names(pathlib.Path(args.source_root))
    missing = sorted(required - english_to_chinese.keys())
    orphan = sorted(english_to_chinese.keys() - required)
    if missing:
        problems.append(f"缺譯原名：{missing}")
    if orphan:
        problems.append(f"孤兒原名：{orphan}")
    print(json.dumps({"required_unique_names": len(required),
                      "catalog_keys": len(seen_keys),
                      "catalog_unique_names": len(english_to_chinese),
                      "problems": len(problems)}, ensure_ascii=False))
    for problem in problems:
        print(problem, file=sys.stderr)
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
