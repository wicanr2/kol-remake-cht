#!/usr/bin/env python3
"""檢查 docs/spec 的狀態登記：索引、各檔狀態行與 CONFORMED 的證據要一致。

  tools/run_python_tool.sh tools/verify_spec_states.py

CLAUDE.md 的順序是「RE 證據 → DRAFT → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED」。
這支不判斷證據夠不夠，只擋登記自己的矛盾：

  1. docs/spec 底下每份 NNN-*.md 都在 README.md 的索引裡有一列，索引指到的檔案都存在。
  2. 索引每列的狀態欄要出現 DRAFT、READY 或 CONFORMED，以第一個出現的為主狀態。
  3. 規格檔的「狀態：」行若有，主狀態要與索引相同。
  4. 主狀態是 CONFORMED 的規格，內文要指向驗證證據：`verification/`、`tools/verify_` 或某個 `_test.go`。
"""
from __future__ import annotations

import pathlib
import re
import sys

SPEC = pathlib.Path("docs/spec")
KEYWORD = re.compile(r"DRAFT|READY|CONFORMED")
ROW = re.compile(r"^\| (SPEC-\d+) \| \[[^\]]+\]\(([^)]+)\) \| ([^|]+) \|", re.M)
EVIDENCE = re.compile(r"verification/|tools/verify_|_test\.go")


def primary(text: str) -> str | None:
    m = KEYWORD.search(text)
    return m.group(0) if m else None


def main() -> int:
    problems: list[str] = []
    index = ROW.findall((SPEC / "README.md").read_text(encoding="utf-8"))
    listed: dict[str, tuple[str, str]] = {}
    for sid, path, status in index:
        if not (SPEC / path).exists():
            problems.append(f"{sid}：索引指到不存在的檔案 {path}")
            continue
        state = primary(status)
        if state is None:
            problems.append(f"{sid}：索引的狀態欄沒有 DRAFT／READY／CONFORMED：{status.strip()[:40]}")
            continue
        listed[path] = (sid, state)

    for f in sorted(SPEC.glob("[0-9][0-9][0-9]-*.md")):
        if f.name not in listed:
            problems.append(f"{f.name}：不在 README.md 的索引裡")
            continue
        sid, state = listed[f.name]
        text = f.read_text(encoding="utf-8")
        line = re.search(r"^狀態[：:]\s*(.+)$", text, re.M)
        if line:
            own = primary(line.group(1))
            if own and own != state:
                problems.append(f"{sid}（{f.name}）：索引是 {state}，檔內狀態行是 {own}")
        if state == "CONFORMED" and not EVIDENCE.search(text):
            problems.append(f"{sid}（{f.name}）：標 CONFORMED，內文沒有指向驗證證據")

    if problems:
        print("\n".join(problems))
        return 1
    print(f"規格狀態登記一致：索引 {len(index)} 列，{len(listed)} 份檔案")
    return 0


if __name__ == "__main__":
    sys.exit(main())
