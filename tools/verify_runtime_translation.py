#!/usr/bin/env python3
"""以原始文字資料為母體，檢查繁中表是否覆蓋城鎮與建角文字。

這是靜態覆蓋檢查；通過後仍須驗證正常玩家路徑、文字安全矩形與
程式內動態字串，不能據此宣稱整個遊戲都已翻譯。
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys


def has_han(value: str) -> bool:
    return any("\u3400" <= char <= "\u9fff" for char in value)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--town-source", default="assets/original-derived-data/town-text.json")
    parser.add_argument("--buildings-source", default="assets/original-derived-data/town-buildings.json")
    parser.add_argument("--town-table", default="docs/i18n/town-zh-Hant.json")
    parser.add_argument("--newchar-table", default="docs/i18n/newchar-zh-Hant.json")
    parser.add_argument("--out")
    args = parser.parse_args()

    source = json.loads(pathlib.Path(args.town_source).read_text())
    buildings = json.loads(pathlib.Path(args.buildings_source).read_text())
    doors = {(town["town"], door["index"]): door
             for town in buildings["towns"] for door in town["doors"]}
    town_table = json.loads(pathlib.Path(args.town_table).read_text())
    newchar = json.loads(pathlib.Path(args.newchar_table).read_text())
    missing: list[dict] = []
    counts = {"town_messages": 0, "town_entries": 0, "town_names": 0,
              "town_people": 0, "newchar": 0}
    for town in source["towns"]:
        town_id = str(town["town"])
        translated_town = town_table["towns"].get(town_id, {})
        for door_id, building in enumerate(town["buildings"]):
            door = translated_town.get("doors", {}).get(str(door_id), {})
            where = {"town": town["town"], "door": door_id}
            raw_door = doors.get((town["town"], door_id))
            if raw_door is None or len(raw_door.get("icon_flags", [])) < 2:
                raise SystemExit(f"建築旗標缺失：town={town['town']} door={door_id}")
            for field, count in (("name", "town_names"), ("person", "town_people")):
                if raw_door.get(field):
                    counts[count] += 1
                    if not has_han(door.get(field, "")):
                        missing.append({**where, "field": field, "reason": "來源名稱缺譯或譯文沒有中文字"})
            entry_source = [m for m in building["messages"] if m["index"] in (0, 1)]
            if entry_source:
                counts["town_entries"] += 1
                if not door.get("entry") or not all(has_han(line) for line in door["entry"]):
                    missing.append({**where, "field": "entry", "reason": "進門敘述缺譯或含英文行"})
            messages = door.get("messages", {})
            for message in building["messages"]:
                index = message["index"]
                if index < 2 or not any(line.strip() for line in message["display"]):
                    continue
                counts["town_messages"] += 1
                translated = messages.get(str(index))
                # RE-139：已知原版直接訊息繪製路徑中，2–5 由接任務、6–7
                # 由任務物品分支讀取，兩者都受第二個 icon flag 的 bit 2 控制。
                # 8–9 沒有 sub_14234 的直接呼叫；這不排除尚未解出的間接讀取。
                if 2 <= index <= 7:
                    path = ("known_direct_path" if raw_door["icon_flags"][1] & 4
                            else "known_direct_path_blocked_by_flag")
                elif 8 <= index <= 9:
                    path = "no_direct_message_renderer_call_other_paths_unknown"
                else:
                    path = "not_assessed"
                if not translated:
                    missing.append({**where, "index": index, "field": "message",
                                    "reason": "缺譯", "known_path": path,
                                    "source": message["display"]})
                elif not all(has_han(line) for line in translated):
                    missing.append({**where, "index": index, "field": "message",
                                    "reason": "譯文有非中文行", "known_path": path,
                                    "translation": translated})

    seen: set[str] = set()
    for group in ("labels", "races", "secondary", "professions", "ui"):
        for row in newchar[group]:
            counts["newchar"] += 1
            key = row["en"]
            if key in seen:
                missing.append({"group": group, "source": key, "reason": "重複原文鍵"})
            seen.add(key)
            if not has_han(row.get("zh", "")):
                missing.append({"group": group, "source": key, "reason": "缺少中文字譯文"})

    result = {"tool": "tools/verify_runtime_translation.py",
              "scope": "城鎮原始文字與建角譯表；不含 Go 動態字串和畫面驗收",
              "sources": {"town": args.town_source, "buildings": args.buildings_source,
                          "town_translation": args.town_table,
                          "newchar_translation": args.newchar_table},
              "checked": counts, "missing": missing}
    if args.out:
        pathlib.Path(args.out).write_text(json.dumps(result, ensure_ascii=False, indent=1) + "\n")
    print(json.dumps({"checked": counts, "missing": len(missing)}, ensure_ascii=False))
    for item in missing[:20]:
        print(json.dumps(item, ensure_ascii=False), file=sys.stderr)
    return 1 if missing else 0


if __name__ == "__main__":
    sys.exit(main())
