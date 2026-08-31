"""核對 SPEC-002 EGA extraction set 表格與固定圖片清冊。"""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


ROW = re.compile(
    r"^\| (?P<name>[A-Z0-9]+PICS) \| (?P<payload>[0-9,]+) \| "
    r"(?P<confirmed>[0-9,]+) \| (?P<pending>[0-9,]+) \| "
    r"(?P<non_image>[0-9,]+) \| (?P<png>[0-9,]+) \|"
)
EXPECTED_TOTALS = (1022, 996, 0, 26, 752)


def number(value: str) -> int:
    return int(value.replace(",", ""))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("inventory", type=Path)
    parser.add_argument("extraction_set", type=Path)
    args = parser.parse_args()

    inventory = json.loads(args.inventory.read_text(encoding="utf-8"))
    archives = {
        record["root_name"].rsplit("/", 1)[-1]: record["count"]
        for record in inventory["logical_archives"]
    }
    rows: dict[str, tuple[int, int, int, int, int]] = {}
    for line in args.extraction_set.read_text(encoding="utf-8").splitlines():
        match = ROW.match(line)
        if not match:
            continue
        name = match["name"]
        if name in rows:
            raise SystemExit(f"重複 extraction set row：{name}")
        rows[name] = tuple(
            number(match[key])
            for key in ("payload", "confirmed", "pending", "non_image", "png")
        )

    if rows.keys() != archives.keys():
        missing = sorted(archives.keys() - rows.keys())
        extra = sorted(rows.keys() - archives.keys())
        raise SystemExit(f"archive 集合不符：missing={missing} extra={extra}")
    for name, values in rows.items():
        payload, confirmed, pending, non_image, _ = values
        if payload != archives[name]:
            raise SystemExit(f"{name} payload 不符：{payload} != {archives[name]}")
        if confirmed + pending + non_image != payload:
            raise SystemExit(f"{name} disposition 未完整覆蓋 payload")

    totals = tuple(sum(values[index] for values in rows.values()) for index in range(5))
    if totals != EXPECTED_TOTALS:
        raise SystemExit(f"總計不符：{totals} != {EXPECTED_TOTALS}")
    print(
        f"validated_archives={len(rows)} payloads={totals[0]} "
        f"confirmed={totals[1]} pending={totals[2]} non_image={totals[3]} png={totals[4]}"
    )


if __name__ == "__main__":
    main()
