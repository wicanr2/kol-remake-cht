#!/usr/bin/env python3
"""把 dosgolem 逐格走規劃段的輸出摘成一份可進版控的收據（#84）。

  python3 tools/summarize_planning_probes.py <輸出 JSON> <基準目錄> <名字>=<目錄> ...

每個目錄是一次 `KOL_BOOK_PLAN=1 KOL_BOOK_PLAN_KEYS=…` 的輸出，裡面要有
`battle-planning.json` 與 `battle-state-records.json`。基準目錄是 2026-09-11 那一場
（`workplace/battle-planning-20260911/`）：每一組的進場狀態都要與它逐 byte 相同，
否則不是同一場遭遇，比了也沒有意義——這一條失敗就整份拒寫。

摘要只留機制：每一鍵之後的圖示列格數與白框位置，以及槽 1（Theodoric）最後寫回的
指令碼、攻擊選項、防禦選項與方向。畫面與原始記錄留在 workplace/，不進版控。
"""
from __future__ import annotations

import hashlib
import json
import pathlib
import sys

ACTOR = 1


def sha(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> int:
    if len(sys.argv) < 4:
        print(__doc__, file=sys.stderr)
        return 2
    out = pathlib.Path(sys.argv[1])
    base = pathlib.Path(sys.argv[2])
    base_state = sha(base / "battle-state-records.json")
    cases = []
    for arg in sys.argv[3:]:
        name, _, folder = arg.partition("=")
        d = pathlib.Path(folder)
        state = sha(d / "battle-state-records.json")
        if state != base_state:
            print(f"{name}：進場狀態與基準不同（{state[:12]} ≠ {base_state[:12]}），拒寫",
                  file=sys.stderr)
            return 1
        probe_path = d / "battle-planning.json"
        probe = json.loads(probe_path.read_text(encoding="utf-8"))
        steps = probe["steps"]
        if steps[0]["step"] != -1 or steps[0]["icon_slots"] != 5:
            print(f"{name}：規劃入口不是五格的主指令列", file=sys.stderr)
            return 1
        rec = steps[-1]["actors"][ACTOR]
        cases.append({
            "name": name,
            "probe_sha256": sha(probe_path),
            "steps": [{"key": s["key"], "icon_slots": s["icon_slots"],
                       "framed_slot": s["framed_slot"]} for s in steps[1:]],
            "final": {k: rec[k] for k in ("command_code", "attack_option",
                                          "defence_option", "facing")},
        })
    body = {
        "schema": "kol-planning-cells-v1",
        "source": "dosgolem-normal-player-path",
        "actor_slot": ACTOR,
        "alignment_seed": 1,
        "entry_state_sha256": base_state,
        "method": "每一組從規劃入口送一串鍵；比每一鍵之後的圖示列格數、白框位置與槽 1 寫回的四個值",
        "cases": cases,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(body, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(f"寫出 {out}：{len(cases)} 組")
    return 0


if __name__ == "__main__":
    sys.exit(main())
