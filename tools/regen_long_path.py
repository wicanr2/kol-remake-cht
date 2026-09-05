#!/usr/bin/env python3
"""重生一條長路徑：戰鬥那幾段用 driver 重跑，其餘照抄舊腳本。

改了戰鬥的規則之後，長路徑收據的固定按鍵序列就作廢了——訊息少一則、敵人多走一步，
後面的按鍵全部錯位。這一支把重生自動化：

1. 讀 plan 與**舊軌跡**，依「這一步有沒有 `battle`」切成交錯的非戰鬥段與戰鬥段；
2. 非戰鬥段照抄舊 token（走路、進城、下注那些不受戰鬥規則影響）；
3. 戰鬥段交給 `tools/run_driver.sh` 現場打完；
4. 拼起來寫回 plan 的 `script`。

**軌跡要是舊的**：跑過 `replay_player_path.py` 之後那一份會被新行為覆寫，
切段就沒有意義了。先 `git checkout -- docs/verification/player-path/` 還原再跑。

  tools/regen_long_path.py arena
  tools/regen_long_path.py arena full-chain-ending

跑完要自己再跑一次 `tools/replay_player_path.py <名字>` 產收據並驗證。
"""

from __future__ import annotations

import json
import os
import pathlib
import subprocess
import sys

PLANS = pathlib.Path("docs/verification/player-path")
WORK = pathlib.Path("workplace/tmp/regen")
# driver 一段最多跑幾輪。一場戰鬥大約要一百多輪，留三倍餘裕。
DRIVE_LIMIT = 400


def segments(trace: list[dict]) -> list[tuple[bool, int, int]]:
    """把軌跡切成 (是不是戰鬥, 第一步, 最後一步)。"""
    out: list[list] = []
    for st in trace:
        inb = st.get("battle") is not None
        if not out or out[-1][0] != inb:
            out.append([inb, st["step"], st["step"]])
        else:
            out[-1][2] = st["step"]
    return [(a, b, c) for a, b, c in out]


def drive(name: str, plan: dict, prefix: list[str], seq: int) -> list[str]:
    """跑一段戰鬥。前綴要含到「進戰場」那一個 token。"""
    WORK.mkdir(parents=True, exist_ok=True)
    pre = WORK / f"{name}-{seq}-pre.txt"
    out = WORK / f"{name}-{seq}-out.txt"
    pre.write_text(" ".join(prefix) + "\n", encoding="utf-8")
    env = dict(os.environ)
    env["KOL_DRIVE_DIR"] = str(WORK / f"{name}-{seq}")
    if plan.get("quest_sites"):
        env["KOL_QUEST_SITES"] = plan["quest_sites"]
    cmd = ["tools/run_driver.sh", str(plan.get("seed", 1)), plan["load"],
           str(pre), str(out), str(DRIVE_LIMIT)]
    print("   driver:", " ".join(cmd))
    subprocess.run(cmd, check=True, env=env)
    return out.read_text(encoding="utf-8").split()


def regen(name: str) -> None:
    path = PLANS / f"{name}.json"
    plan = json.loads(path.read_text(encoding="utf-8"))
    trace = json.loads(pathlib.Path(plan["trace"]).read_text(encoding="utf-8"))["steps"]
    old = plan["script"].split()
    segs = segments(trace)
    print(f"== {name}：{len(old)} 個 token，切成 {len(segs)} 段")

    tokens: list[str] = []
    seq = 0
    for k, (inb, a, b) in enumerate(segs):
        if not inb:
            # 非戰鬥段照抄。step a…b 由 token[a−1 … b−1] 產生。
            #
            # **開頭那一個 token 不抄**（第一段除外）：它是把上一場戰鬥推出去的
            # 那一下，而 driver 已經自己按到「回野外／回城裡」才停。再按一次會落在
            # 戰鬥之外——野外的 token 表沒有 SPACE，整條腳本就死在那裡。
            #
            # 結尾那一個要抄：它是按下去進下一個戰場的那一下，正好給下一段當前綴。
            lo = a if k else 0
            hi = b + 1 if k + 1 < len(segs) else b
            chunk = old[lo:hi]
            tokens += chunk
            print(f"   照抄非戰鬥段 步 {a}…{b}（{len(chunk)} 個 token）")
            continue
        seq += 1
        before = len(tokens)
        tokens = drive(name, plan, tokens, seq)
        print(f"   第 {seq} 場：{len(tokens) - before} 個 token")

    plan["script"] = " ".join(tokens)
    path.write_text(json.dumps(plan, ensure_ascii=False, indent=1) + "\n",
                    encoding="utf-8")
    print(f"   寫回 {path}：{len(tokens)} 個 token（舊的是 {len(old)} 個）")


def main() -> int:
    if len(sys.argv) < 2:
        raise SystemExit("用法：regen_long_path.py <收據名字> [更多…]")
    for name in sys.argv[1:]:
        regen(name)
    return 0


sys.exit(main())
