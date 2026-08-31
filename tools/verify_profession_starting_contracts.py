"""以中文手冊職業表驗證 IDA 匯出的起始金冠、武器與技能契約。"""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


ROW = re.compile(r"^\|([^|]+)\|\s*\d+\s*\|\s*\d+\s*\|\s*\d+\s*\|\s*\d+\s*\|\s*\d+\s*\|\s*\d+\s*\|\s*\d+\s*\|([^|]+)\|$")
EQUIPMENT = re.compile(r"^([A-Z]{2})\s+(\d+)(?:/(\d+))?$")


def parse_manual(path: Path) -> list[dict]:
    text = path.read_text(encoding="utf-8")
    start = text.index("## 角色職業：男性人類")
    end = text.index("## 屬性", start)
    rows = []
    for line in text[start:end].splitlines():
        match = ROW.match(line.strip())
        if not match:
            continue
        name, equipment_text = (part.strip() for part in match.groups())
        parts = [part.strip() for part in equipment_text.split("；")]
        if not parts or not parts[0].startswith("GC "):
            raise ValueError(f"缺少 GC：{line}")
        gold = int(parts[0].split()[1])
        weapons = []
        for part in parts[1:]:
            item = EQUIPMENT.fullmatch(part)
            if item is None:
                raise ValueError(f"無法解析裝備：{part}")
            weapons.append({
                "weapon": item.group(1),
                "skills": [int(item.group(2)), int(item.group(3) or 0)],
            })
        rows.append({"manual_name": name, "gold": gold, "weapons": weapons})
    if len(rows) != 35:
        raise ValueError(f"手冊職業列數不是 35：{len(rows)}")
    return [*rows[:28], *rows[28:32], *rows[28:32], *rows[32:35]]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("summary", type=Path)
    parser.add_argument("manual", type=Path)
    args = parser.parse_args()

    expected = parse_manual(args.manual)
    summary = json.loads(args.summary.read_text(encoding="utf-8"))
    actual = summary["profession_starting_contracts"]
    if len(actual) != 39 or len(expected) != 39:
        raise SystemExit("職業契約數量不是 39")

    for index, (want, got) in enumerate(zip(expected, actual)):
        if got["record_index"] != index or got["gold_crowns"] != want["gold"]:
            raise SystemExit(f"record {index} 金冠不符：{want} != {got}")
        weapons = want["weapons"]
        if not weapons:
            raise SystemExit(f"record {index} 手冊沒有起始武器")
        if got["primary_weapon"] != weapons[0]["weapon"]:
            raise SystemExit(f"record {index} 主要武器不符")
        if got["primary_skill_pair_manual_units"] != weapons[0]["skills"]:
            raise SystemExit(f"record {index} 主要技能不符")
        secondary = weapons[1] if len(weapons) > 1 else None
        if got["secondary_weapon"] != (secondary["weapon"] if secondary else None):
            raise SystemExit(f"record {index} 次要武器不符")
        expected_secondary_skills = secondary["skills"] if secondary else [0, 0]
        if got["secondary_skill_pair_manual_units"] != expected_secondary_skills:
            raise SystemExit(f"record {index} 次要技能不符")
        if len(weapons) > 2:
            raise SystemExit(f"record {index} 出現超過兩件起始武器")
    print("39 筆起始金冠、武器與技能均符合中文手冊職業表")


if __name__ == "__main__":
    main()
