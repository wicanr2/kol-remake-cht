#!/usr/bin/env python3
"""檢查 UI 文字庫的譯文有沒有把控制碼弄丟。

原版的 UI 訊息不是完整句子，是靠記錄號串接的碎片，裡面帶三種東西：代換碼
`|A`…`|S`（執行期換成名字、代名詞、武器名）、版面碼 `~Vxx`／`~Hxx`／`~Cxx`
（掃描列、欄、置中）與換行 `@`。**譯文漏掉一個代換碼，畫面上就會少一個名字**，
而這種錯誤單元測試抓不到——規則層自己不讀譯文。

判準三條，任何一條不過就是失敗：

- 代換碼的**集合**要一致。位置可以改（中文語序不同），但不得憑空多出或消失。
- `~Vxx`（掃描列）的**集合**要一致。那是原版量出來的縱向座標，少一個就有一段畫錯位置。
- `~Hxx`／`~Cxx`（欄、置中）**可以少**但不得多。中文比英文短，兩行常併成一行，
  併掉的那一行不再需要把游標移回左緣；但憑空多一個欄座標會把文字推到別的地方。
- 換行 `@` 的**個數**可以少（中文較短），但不得變多——多一行會擠掉下一段。
- 每一行的實際像素寬度不得超過畫面寬。這是**弱檢查**：它只擋得住「畫到畫面外」，
  擋不住「畫到框外」——每一筆記錄畫在哪個框要等它接上畫面才知道，那時再用
  `tools/verify_text_fit.py` 的安全矩形檢查。

刻意留空的記錄要在譯文檔裡寫 `note` 說明理由，否則算漏譯。

  python3 tools/verify_ui_translation.py [--table docs/i18n/ui-zh-Hant.json]
      [--source assets/original-derived-data/ui-text.json] [--out 收據.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import re
import sys

SUBSTITUTION = re.compile(r"\|[A-Z]")
ROW = re.compile(r"~V[0-9A-Fa-f]{2}")
COLUMN = re.compile(r"~[HC][0-9A-Fa-f]{2}")
LAYOUT_ANY = re.compile(r"~[VHC][0-9A-Fa-f]{2}|\|[A-Z]")


def check(bank: str, index: str, row: dict) -> list[str]:
    en, zh, note = row["en"], row["zh"], row.get("note", "")
    if en in ("", "\x00", "\x1a\x00"):
        return []
    if zh == "":
        return [] if note else ["漏譯，而且沒有寫明為什麼留空"]
    problems = []
    for name, pattern, may_drop in (("代換碼", SUBSTITUTION, False),
                                    ("掃描列", ROW, False),
                                    ("欄", COLUMN, True)):
        a, b = set(pattern.findall(en)), set(pattern.findall(zh))
        lost, extra = sorted(a - b), sorted(b - a)
        bits = []
        if lost and not may_drop:
            bits.append(f"漏了 {' '.join(lost)}")
        if extra:
            bits.append(f"多了 {' '.join(extra)}")
        if bits:
            problems.append(f"{name}不一致（{'；'.join(bits)}）")
    if zh.count("@") > en.count("@"):
        problems.append(f"換行變多（原文 {en.count('@')} 個，譯文 {zh.count('@')} 個）")
    return problems


def width_of(line: str, advance: dict[str, int]) -> int:
    """一行的像素寬度。字模表查不到的字元（控制碼剝乾淨後不該有）算全形。"""
    return sum(advance.get(ch, 16) for ch in line)


def too_wide(zh: str, advance: dict[str, int], limit: int) -> list[str]:
    if not zh:
        return []
    plain = LAYOUT_ANY.sub("", zh).replace("]", " ")
    out = []
    for i, line in enumerate(plain.split("@")):
        w = width_of(line, advance)
        if w > limit:
            out.append(f"第 {i + 1} 行寬 {w} px，超過畫面的 {limit} px")
    return out


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--table", default="docs/i18n/ui-zh-Hant.json")
    ap.add_argument("--source", default="assets/original-derived-data/ui-text.json")
    ap.add_argument("--font", default="assets/ui-font/manifest.json")
    ap.add_argument("--screen-width", type=int, default=640,
                    help="中文模式的畫面寬度，預設 640")
    ap.add_argument("--out")
    args = ap.parse_args()

    advance = {}
    font = json.loads(pathlib.Path(args.font).read_text())
    for glyph in font["fonts"][0]["glyphs"]:
        advance[glyph["character"]] = glyph["advance"]

    table = json.loads(pathlib.Path(args.table).read_text())
    source = json.loads(pathlib.Path(args.source).read_text())
    banks = {b["bank"]: b for b in source["banks"]}

    failures, checked, intentional = [], 0, 0
    for name, rows in table["banks"].items():
        want = banks[name]
        if len(rows) != int(want["record_count"]):
            failures.append({"bank": name, "index": None,
                             "problems": [f"筆數 {len(rows)} 不等於原文的 {want['record_count']}"]})
        for index, row in rows.items():
            en = want["records"][int(index)]["text"]
            if row["en"] != en:
                failures.append({"bank": name, "index": index,
                                 "problems": ["原文與 ui-text.json 不一致，譯文檔過期了"]})
                continue
            problems = check(name, index, row)
            problems += too_wide(row["zh"], advance, args.screen_width)
            checked += 1
            if row["zh"] == "" and row.get("note"):
                intentional += 1
            if problems:
                failures.append({"bank": name, "index": index, "en": en,
                                 "zh": row["zh"], "problems": problems})

    receipt = {
        "tool": "tools/verify_ui_translation.py",
        "table": args.table,
        "source_banks": {n: b["sha256"] for n, b in banks.items()},
        "checked": checked,
        "intentionally_empty": intentional,
        "failures": failures,
    }
    if args.out:
        pathlib.Path(args.out).write_text(
            json.dumps(receipt, ensure_ascii=False, indent=1) + "\n")
    print(json.dumps({"checked": checked, "intentionally_empty": intentional,
                      "failures": len(failures)}, ensure_ascii=False))
    for f in failures[:20]:
        print(f"  {f['bank']} {f['index']}: {'；'.join(f['problems'])}", file=sys.stderr)
        if "en" in f:
            print(f"    原文 {f['en']!r}\n    譯文 {f['zh']!r}", file=sys.stderr)
    return 1 if failures else 0


sys.exit(main())
