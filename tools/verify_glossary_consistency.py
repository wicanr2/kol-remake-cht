#!/usr/bin/env python3
"""檢查遊戲譯表是否與術語表一致（#39：術語表是唯一來源）。

術語表 `docs/manual/glossary.md` 每列是「英文｜繁中｜狀態｜備註」；英文與繁中可用
`A / B` 對 `甲／乙` 並列多個詞。拿每個英文詞去比對遊戲裡**原文完全相同**的譯文：

- `docs/i18n/display-names-zh-Hant.tsv` 的資料名稱（不分大小寫）
- `docs/i18n/newchar-zh-Hant.json` 的欄名、種族、次要、職業與介面字串
- `docs/i18n/ui-zh-Hant.json` 三個文字庫中原文只有那一個詞的記錄（去掉 `]`、`@`、
  標點與版面碼）

譯法不同就列出來。狀態不是核准或暫譯的詞不比對。有不一致時以非零結束。

  python3 tools/verify_glossary_consistency.py [--out 報告.json]
"""

from __future__ import annotations

import argparse
import csv
import json
import pathlib
import re
import sys

LAYOUT = re.compile(r"~[VHC][0-9A-Fa-f]{2}")


def glossary(path: pathlib.Path) -> dict[str, tuple[str, str, int]]:
    """英文詞（小寫）→（繁中、狀態、行號）。"""
    terms: dict[str, tuple[str, str, int]] = {}
    for n, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        if len(cells) < 3 or not line.startswith("| ") or cells[0] in ("英文", "---"):
            continue
        en, zh, status = cells[0], cells[1], cells[2]
        if status not in ("核准", "暫譯"):
            continue
        ens = [e.strip() for e in en.split("/")]
        zhs = [z.strip() for z in re.split(r"[／/]", zh)]
        if len(ens) != len(zhs):
            if len(ens) > 1:
                continue  # 並列數不對等（例如英文有別名、中文只有一個）就不拆
            zhs = [zh]
        for e, z in zip(ens, zhs):
            e = e.strip("`").strip()
            z = re.sub(r"（.*?）", "", z).strip()
            if e and z and not re.search(r"[`…]", e):
                terms.setdefault(e.lower(), (z, status, n))
    return terms


def bare(text: str) -> str:
    text = LAYOUT.sub("", text)
    return re.sub(r"[\]@\"'.,:!?！：，。？]", "", text).strip()


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--glossary", default="docs/manual/glossary.md")
    p.add_argument("--exceptions", default="docs/i18n/glossary-exceptions.json")
    p.add_argument("--out")
    args = p.parse_args()
    terms = glossary(pathlib.Path(args.glossary))
    found: list[dict] = []

    def compare(source: str, en: str, zh: str) -> None:
        key = bare(en).lower()
        if key in terms and zh:
            want, status, line = terms[key]
            got = bare(zh)
            found.append({"source": source, "en": en, "zh": got, "glossary": want,
                          "status": status, "glossary_line": line, "match": got == want})

    with open("docs/i18n/display-names-zh-Hant.tsv", encoding="utf-8", newline="") as fh:
        for row in csv.DictReader(fh, delimiter="\t"):
            compare(f"display-names:{row['key']}", row["en"], row["zh-Hant"])
    newchar = json.loads(pathlib.Path("docs/i18n/newchar-zh-Hant.json").read_text())
    for group in ("labels", "races", "secondary", "professions", "ui"):
        for row in newchar[group]:
            compare(f"newchar:{group}:{row['en']}", row["en"], row.get("zh", ""))
    ui = json.loads(pathlib.Path("docs/i18n/ui-zh-Hant.json").read_text())
    for bank, rows in ui["banks"].items():
        for index, row in rows.items():
            compare(f"ui:{bank}:{index}", row.get("en", ""), row.get("zh", ""))

    exceptions = json.loads(pathlib.Path(args.exceptions).read_text())["exceptions"]
    raw = [f for f in found if not f["match"]]
    conflicts = [f for f in raw if f["source"] not in exceptions]
    stale = sorted(set(exceptions) - {f["source"] for f in raw})
    for source in stale:
        conflicts.append({"source": source, "en": "", "zh": "", "glossary": "", "status": "",
                          "glossary_line": 0, "match": False, "stale_exception": True})
    report = {"glossary_terms": len(terms), "compared": len(found), "conflicts": conflicts}
    if args.out:
        pathlib.Path(args.out).write_text(json.dumps(report, ensure_ascii=False, indent=1) + "\n")
    print(json.dumps({"glossary_terms": len(terms), "compared": len(found),
                      "conflicts": len(conflicts)}, ensure_ascii=False))
    for c in conflicts:
        if c.get("stale_exception"):
            print(f"  例外清單的 {c['source']} 已經沒有不一致，請移除", file=sys.stderr)
            continue
        print(f"  {c['source']}: {c['en']!r} 譯作「{c['zh']}」，術語表第 {c['glossary_line']} 行是"
              f"「{c['glossary']}」（{c['status']}）", file=sys.stderr)
    return 1 if conflicts else 0


if __name__ == "__main__":
    sys.exit(main())
