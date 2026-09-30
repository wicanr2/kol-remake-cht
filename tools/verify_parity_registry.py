#!/usr/bin/env python3
"""檢查 docs/verification/parity-registry.json：結構、引用的檔案與抽樣母體的數量。

  tools/run_python_tool.sh tools/verify_parity_registry.py

這支只驗登記資料自己站得住，不判斷任何階段有沒有通過。檢查項目：

  - 四個階段依序是 launchable、playable、approximate、strict，每個都有宣稱、母體、gate、
    硬失敗、平台矩陣與證據。
  - 平台矩陣三個平台都列出來。
  - evidence 引用的檔案都存在。
  - 有 glob 的母體，實際檔案數要等於登記的 count（母體被悄悄縮小會在這裡被抓到）。
  - status 只能是 not-claimed、linux-only、claimed-local。
"""
from __future__ import annotations

import json
import pathlib
import sys

REGISTRY = pathlib.Path("docs/verification/parity-registry.json")
STAGES = ["launchable", "playable", "approximate", "strict"]
PLATFORMS = {"linux-x86_64", "windows-amd64", "macos-universal"}
STATUS = {"not-claimed", "linux-only", "claimed-local"}
NEEDED = ["id", "name", "claim", "population", "gate", "hard_failures", "platform_matrix", "evidence", "status"]


def main() -> int:
    doc = json.loads(REGISTRY.read_text(encoding="utf-8"))
    problems: list[str] = []
    if doc.get("schema") != "kol-parity-registry/1":
        problems.append(f"schema 不是 kol-parity-registry/1：{doc.get('schema')}")
    stages = doc.get("stages", [])
    if [s.get("id") for s in stages] != STAGES:
        problems.append(f"階段順序應為 {STAGES}，實際 {[s.get('id') for s in stages]}")
    for s in stages:
        sid = s.get("id", "?")
        for key in NEEDED:
            if not s.get(key):
                problems.append(f"{sid}：缺 {key}")
        if s.get("status") not in STATUS:
            problems.append(f"{sid}：status {s.get('status')!r} 不在 {sorted(STATUS)}")
        got = {m.get("platform") for m in s.get("platform_matrix", [])}
        if got != PLATFORMS:
            problems.append(f"{sid}：平台矩陣是 {sorted(got)}，應為 {sorted(PLATFORMS)}")
        for path in s.get("evidence", []):
            if not pathlib.Path(path).exists():
                problems.append(f"{sid}：引用的檔案不存在 {path}")
        pop = s.get("population", {})
        if "glob" in pop:
            files = [p for p in pathlib.Path(".").glob(pop["glob"])
                     if not p.name.endswith(pop.get("exclude_suffix", "\0"))]
            if len(files) != pop.get("count"):
                problems.append(f"{sid}：母體 {pop['glob']} 登記 {pop.get('count')} 份，實際 {len(files)} 份")
        if "items" in pop and pop.get("count") != len(pop["items"]):
            problems.append(f"{sid}：母體 items 有 {len(pop['items'])} 項，登記 {pop.get('count')}")
        if "source" in pop and not pathlib.Path(pop["source"]).exists():
            problems.append(f"{sid}：母體來源不存在 {pop['source']}")
    if problems:
        print("\n".join(problems))
        return 1
    print(f"登記資料通過：{len(stages)} 個階段，"
          + "、".join(f"{s['id']}={s['status']}" for s in stages))
    return 0


if __name__ == "__main__":
    sys.exit(main())
