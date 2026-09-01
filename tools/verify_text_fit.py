#!/usr/bin/env python3
"""中文模式的文字安全矩形檢查。

繁體中文字比原版英文字模寬，版面能不能容納不是看起來像就算數。這支工具用實際的
字模表字寬，算出每一個會畫到畫面上的字串的實際寬度，逐條檢查：

- 每個字元在字模表裡都有字模（缺字會被畫成空白，那是漏字不是排版問題）。
- 字串畫在指定起點時，右緣不超出所屬的文字安全矩形。

版面座標由 `cmd/kol -dump-cjk-layout` 匯出，因此不會和程式裡的版面表脫節。

  python3 tools/verify_text_fit.py --layout workplace/cjk-layout.json \\
      --font assets/ui-font/manifest.json --strings docs/i18n/newchar-zh-Hant.json \\
      --data assets/original-derived-data/character-creation.json \\
      --out docs/verification/cjk-text-fit.json
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--layout", required=True)
    p.add_argument("--font", required=True)
    p.add_argument("--strings", required=True)
    p.add_argument("--data", required=True)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    layout_doc = json.loads(pathlib.Path(args.layout).read_text())
    layout, box = layout_doc["layout"], layout_doc["text_box"]
    font = json.loads(pathlib.Path(args.font).read_text())["fonts"][0]
    advance = {chr(g["codepoint"]): g["advance"] for g in font["glyphs"]}
    default = font["advance"]
    table = json.loads(pathlib.Path(args.strings).read_text())
    data = json.loads(pathlib.Path(args.data).read_text())

    zh = {}
    for key in ("labels", "races", "secondary", "professions", "ui"):
        for item in table.get(key, []):
            zh[item["en"]] = item["zh"]

    def width(text: str) -> tuple[int, list[str]]:
        total, missing = 0, []
        for ch in text:
            # 空白沒有字模也不畫墨，但照樣前進；半形算半格、全形算一格。
            if ch == " ":
                total += default // 2
            elif ch == "\u3000":
                total += default
            elif ch in advance:
                total += advance[ch]
            else:
                missing.append(ch)
                total += default
        return total, missing

    checks = []

    def check(kind: str, en: str, x: int, right: int) -> None:
        text = zh.get(en, en)
        w, missing = width(text)
        checks.append({"kind": kind, "en": en, "zh": text, "x": x, "width": w,
                       "right_limit": right, "overflow": max(0, x + w - right),
                       "missing_glyphs": missing,
                       "pass": x + w <= right and not missing})

    left_right = box["LeftX"] + box["LeftW"]
    right_right = box["RightX"] + box["RightW"]

    check("heading", "New character", layout["HeadingX"], left_right)
    check("heading", "Select one", layout["RightX"], right_right)
    for item in table["labels"]:
        check("label", item["en"], layout["FieldX"], layout["ColonX"])
    for en in ("Character options", "LIST CHARACTERS", "CREATE CHARACTER",
               "DELETE CHARACTER", "EDIT PICTURE", "EDIT SHIELD", "HARBOR MASTER"):
        check("menu", en, layout["MenuX"], left_right)
    check("heading", "Character names", layout["HeadingX"], left_right)
    check("heading", "Location", layout["RightX"], right_right)
    for item in table["races"] + table["secondary"]:
        check("value", item["en"], layout["ValueX"], left_right)
    seen = set()
    for group in data["profession_groups"]:
        for prof in group["professions"]:
            if prof["name"] in seen:
                continue
            seen.add(prof["name"])
            check("value", prof["name"], layout["ValueX"], left_right)
            check("list", prof["name"], layout["RightX"], right_right)
    for en in ("REROLL STATS", "ACCEPT STATS", "USE ARROW KEYS", "TO SELECT A",
               "CHARACTER", "PRESS RETURN", "TO ACCEPT"):
        check("list", en, layout["RightX"], right_right)
    check("actions", "CANCEL", layout["ActionsX"], right_right)

    # 名字欄容量：16 個 ASCII 字元的最寬情形。
    widest = max(advance[chr(c)] for c in range(0x21, 0x7F) if chr(c) in advance)
    name_width = widest * 16
    checks.append({"kind": "name", "en": "name field capacity 16",
                   "zh": "", "x": layout["FieldX"], "width": name_width,
                   "right_limit": left_right,
                   "overflow": max(0, layout["FieldX"] + name_width - left_right),
                   "missing_glyphs": [],
                   "pass": layout["FieldX"] + name_width <= left_right})

    # 最後一列的底緣要留在文字安全矩形內。
    limit = box["LeftY"] + box["LeftH"]
    right_limit_y = box["RightY"] + box["RightH"]
    for name, bottom, lim in (
            ("attribute rows",
             layout["AttrY"] + 9 * layout["AttrStep"] + font["cell"]["height"], limit),
            ("right page list (12 items)",
             layout["RightListY"] + 11 * layout["RightStep"] + font["cell"]["height"],
             right_limit_y),
            ("character options menu",
             layout["MenuY"] + 5 * layout["MenuStep"] + font["cell"]["height"], limit),
            ("roster (16 slots)",
             layout["RosterY"] + 15 * layout["RosterStep"] + font["cell"]["height"],
             limit)):
        checks.append({"kind": "vertical", "en": name, "zh": "", "x": 0,
                       "width": bottom, "right_limit": lim,
                       "overflow": max(0, bottom - lim), "missing_glyphs": [],
                       "pass": bottom <= lim})
    bottom = 0


    failed = [c for c in checks if not c["pass"]]
    receipt = {"schema": "kol-cjk-text-fit-v1",
               "tool": "tools/verify_text_fit.py",
               "layout": args.layout, "font": args.font, "strings": args.strings,
               "checked": len(checks), "failed": len(failed),
               "criterion": "每個字元都要有字模，且字串右緣不得超出所屬的文字安全矩形",
               "failures": failed, "checks": checks}
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"checked": len(checks), "failed": len(failed)},
                     ensure_ascii=False))
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
