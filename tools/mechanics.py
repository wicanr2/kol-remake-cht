"""收據的機制斷言：在 `-trace` 軌跡上找「這件事有沒有照規則發生」，不比畫面。

2026-09-26 定案：對拍與 remake 收據以遊戲機制為準，不要求截圖逐位元相同
（CONTEXT.md 已定案決策）。亂數一位移，事件發生在第幾步就會變，所以斷言不綁
固定步數，而是**依序找第一個符合條件的步**：

    {"find": {"battle": {"stage": 1, "site": -1}}, "time_part_change": true,
     "note": "時段變換那一下被伏擊"}

- `find`：條件；寫法同 `verify_full_path.py` 的里程碑（值為 null 表示「這一步沒有
  這一段」），另外支援 `{"$re": "正規式"}`（字串，或清單裡任一行符合）、
  `{"$in": [...]}`、`{"$gte": n}`、`{"$lte": n}`。
- 從上一個斷言找到的那一步**之後**往下找；`within` 限制最多再走幾步。
- `time_part_change`：那一步的 `world.time_index` 要與前一步不同（遭遇只在時段
  變換時擲，SPEC-011 §1）。
- `same_step_as_previous`：要與上一個斷言落在同一步（同一步上的兩個條件分開寫）。
"""

from __future__ import annotations

import re


def match(want, got, where: str, failures: list[str]) -> None:
    if isinstance(want, dict) and any(k.startswith("$") for k in want):
        for op, arg in want.items():
            if op == "$re":
                lines = got if isinstance(got, list) else [got]
                if not any(isinstance(x, str) and re.search(arg, x) for x in lines):
                    failures.append(f"{where}：{got!r} 不符合 /{arg}/")
            elif op == "$in":
                if got not in arg:
                    failures.append(f"{where}：{got!r} 不在 {arg!r}")
            elif op == "$gte":
                if not isinstance(got, (int, float)) or got < arg:
                    failures.append(f"{where}：{got!r} 小於 {arg!r}")
            elif op == "$lte":
                if not isinstance(got, (int, float)) or got > arg:
                    failures.append(f"{where}：{got!r} 大於 {arg!r}")
            else:
                failures.append(f"{where}：不認得的條件 {op}")
        return
    if isinstance(want, dict):
        if not isinstance(got, dict):
            failures.append(f"{where}：期望物件，實際 {got!r}")
            return
        for key, value in want.items():
            if value is None:
                if got.get(key) is not None:
                    failures.append(f"{where}.{key}：期望這一步沒有這一段，實際還在")
                continue
            if key not in got:
                failures.append(f"{where}.{key}：軌跡裡沒有這個欄位")
                continue
            match(value, got[key], f"{where}.{key}", failures)
        return
    if want != got:
        failures.append(f"{where}：期望 {want!r}，實際 {got!r}")


def _time_index(step: dict):
    return (step.get("world") or {}).get("time_index")


def fits(check: dict, steps: list[dict], i: int) -> list[str]:
    """第 i 步符不符合這一個斷言；回傳不符的原因（空清單表示符合）。"""
    failures: list[str] = []
    match(check.get("find", {}), steps[i], f"第 {i} 步", failures)
    if check.get("time_part_change"):
        if i == 0 or _time_index(steps[i]) == _time_index(steps[i - 1]):
            failures.append(f"第 {i} 步不是時段變換的那一步")
    return failures


def locate(checks: list[dict], steps: list[dict]) -> tuple[list[int], list[str]]:
    """依序找每個斷言落在哪一步。找不到的記成 -1 並寫進失敗原因。"""
    found: list[int] = []
    failures: list[str] = []
    cursor = 0
    for n, check in enumerate(checks):
        label = check.get("note") or f"第 {n + 1} 個斷言"
        if check.get("same_step_as_previous") and found:
            i = found[-1]
            why = fits(check, steps, i) if i >= 0 else ["前一個斷言沒找到"]
            found.append(i if not why else -1)
            failures.extend(f"{label}：{w}" for w in why)
            continue
        limit = len(steps)
        if "within" in check:
            limit = min(limit, cursor + int(check["within"]) + 1)
        hit = -1
        last_why: list[str] = []
        for i in range(cursor, limit):
            why = fits(check, steps, i)
            if not why:
                hit = i
                break
            last_why = why
        found.append(hit)
        if hit < 0:
            failures.append(f"{label}：找不到符合的步（最後一次：{'；'.join(last_why[:2])}）")
            continue
        cursor = hit + 1
    return found, failures
