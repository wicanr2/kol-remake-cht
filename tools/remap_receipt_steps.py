#!/usr/bin/env python3
"""重生腳本之後，把收據的里程碑與檢查點對到新的步號。

`regen_long_path.py` 換掉戰鬥那幾段的按鍵之後，步號整條位移了。里程碑本來就寫著
「這一步該長什麼樣」（`battle.stage`、`world.x`…），所以這一支拿那些條件去新軌跡裡
**依序**找第一個符合的步，再把檢查點照同一份對照表搬過去。

依序找是必要的：`battle.stage == 5` 在一條有三場戰鬥的路徑上會出現三次，
只有「上一個里程碑之後」那個才是要的。

  tools/replay_player_path.py arena      # 先跑一次產生新軌跡
  tools/remap_receipt_steps.py arena
  tools/replay_player_path.py arena      # 再跑一次，這次要全過

找不到對應的里程碑會**留在原地並印出來**——那表示新打法真的沒走過那個狀態，
要人去看，不能默默改掉。
"""

from __future__ import annotations

import json
import pathlib
import sys

PLANS = pathlib.Path("docs/verification/player-path")


def match(node: dict, cond: dict) -> bool:
    for k, v in cond.items():
        # null ＝「這一步沒有這一段」，與 verify_full_path.py 的 subset 同一套。
        if v is None:
            if node.get(k) is not None:
                return False
            continue
        if k not in node:
            return False
        if isinstance(v, dict):
            if not isinstance(node[k], dict) or not match(node[k], v):
                return False
        elif node[k] != v:
            return False
    return True


def remap(name: str) -> int:
    path = PLANS / f"{name}.json"
    plan = json.loads(path.read_text(encoding="utf-8"))
    trace = json.loads(pathlib.Path(plan["trace"]).read_text(encoding="utf-8"))["steps"]
    mapping: dict[int, int] = {}
    missing = 0
    last = -1
    for m in plan.get("milestones", []):
        cond = {k: v for k, v in m.items() if k not in ("step", "note")}
        hit = None
        for st in trace:
            if st["step"] <= last:
                continue
            if match(st, cond):
                hit = st["step"]
                break
        if hit is None:
            print(f"   找不到里程碑：步 {m['step']} {m.get('note', '')}")
            missing += 1
            continue
        mapping[m["step"]] = hit
        m["step"] = hit
        last = hit
    for c in plan.get("checkpoints", []):
        old = c["step"]
        if old not in mapping:
            print(f"   檢查點步 {old} 沒有對應的里程碑，留在原地")
            missing += 1
            continue
        c["step"] = mapping[old]
        if isinstance(c.get("file"), str) and c["file"].startswith("step-"):
            c["file"] = f"step-{mapping[old]:03d}.png"
    path.write_text(json.dumps(plan, ensure_ascii=False, indent=1) + "\n",
                    encoding="utf-8")
    print(f"== {name}：對上 {len(mapping)} 個里程碑，{missing} 個沒對上")
    return missing


def main() -> int:
    if len(sys.argv) < 2:
        raise SystemExit("用法：remap_receipt_steps.py <收據名字> [更多…]")
    bad = 0
    for name in sys.argv[1:]:
        bad += remap(name)
    return 1 if bad else 0


sys.exit(main())
