#!/usr/bin/env python3
"""從收據的 plan 重生截圖與軌跡，再驗一次。

長路徑的收據（`docs/verification/player-path/*.json` 裡有 `trace` 那幾份）本來就
帶著重跑所需的一切：腳本、種子、要讀的存檔、要截哪幾步。這支工具把它們組成
`run_kol_shots.sh` 的規格檔跑一遍，再交給 `verify_full_path.py` 產收據。

**為什麼要有這一支**：規格檔原本寫在 `workplace/`（不進版控），改了戰鬥數值之後
沒有東西可以重跑，只能照收據的散文描述猜。猜錯一個 token 就對不上——
`full-chain-ending` 的 `-quest-sites` 就是這樣漏掉的（描述有寫，機器讀不到），
於是重跑時打的是原版那十二隻 CYCLOPS，怎麼跑都跑不完。

  tools/replay_player_path.py full-chain arena
  tools/replay_player_path.py --all

必須在 Docker 內執行 Python。容器內設 `KOL_BIN` 可直接使用已建置遊戲；
未設定時沿用 Docker 包裝腳本。此工具會更新正式收據，審查回歸先用
`check_player_paths.py`；可寫目錄與既有目標須先核對 UID/GID。
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import subprocess
import sys

PLANS = pathlib.Path("docs/verification/player-path")
SPECS = pathlib.Path("workplace/tmp/replay")


def spec_line(plan: dict) -> str:
    """把 plan 組成一行規格檔。旗標的順序與手寫時相同，方便逐字比對。"""
    # 長路徑多半從標題走起（flow）；讀存檔直接在城裡開場的要在 plan 寫 `scene`，
    # 與 check_player_paths.py 的判斷相同。
    scene = plan.get("scene") or "flow"
    args = [f"{plan['shot_dir']}/step-000.png", "-scene", scene]
    if plan.get("seed") is not None:
        args += ["-seed", str(plan["seed"])]
    if plan.get("load"):
        args += ["-load", plan["load"]]
    if plan.get("save"):
        args += ["-save", plan["save"]["file"]]
    # 換過的資料檔要跟著 plan 走，不能只寫在散文裡。
    if plan.get("quest_sites"):
        args += ["-quest-sites", plan["quest_sites"]]
    if plan.get("monsters"):
        args += ["-monsters", plan["monsters"]]
    # 受控亂數（CLAUDE.md 允許的「明示的受控亂數輸入」）：plan 寫明才帶，收據照抄。
    if plan.get("ambush_rolls"):
        args += ["-ambush-rolls", plan["ambush_rolls"]]
    flag = "-flow-script" if scene == "flow" else "-play-script"
    args += [flag, plan["script"], "-shot-dir", plan["shot_dir"]]
    keep = sorted({c["step"] for c in plan.get("checkpoints", [])})
    if keep:
        args += ["-shot-steps", ",".join(str(k) for k in keep)]
    args += ["-trace", plan["trace"]]
    return "|".join(args)


def replay(name: str) -> bool:
    path = PLANS / f"{name}.json"
    plan = json.loads(path.read_text(encoding="utf-8"))
    if "trace" not in plan:
        print(f"{name}：不是長路徑收據（沒有 trace），改用 verify_player_path.py")
        return False
    SPECS.mkdir(parents=True, exist_ok=True)
    spec = SPECS / f"{name}.spec"
    runtime_plan = dict(plan)
    if plan.get("load"):
        # v1 匯入會在輸入旁保存轉換檔；保留觀察來源的原始位元組。
        copied = SPECS / f"{name}-input.json"
        copied.write_bytes(pathlib.Path(plan["load"]).read_bytes())
        runtime_plan["load"] = str(copied)
    line = spec_line(runtime_plan)
    spec.write_text(line + "\n", encoding="utf-8")
    binary = os.environ.get("KOL_BIN")
    command = ([binary, *line.split("|")[1:]] if binary else
               ["tools/run_kol_shots.sh", str(spec)])
    run = subprocess.run(command,
                         capture_output=True, text=True)
    (SPECS / f"{name}-engine.log").write_text(run.stdout + run.stderr, encoding="utf-8")
    if run.returncode != 0:
        print(f"{name}：跑不動\n{run.stderr[-800:]}")
        return False
    out = PLANS / f"{name}-receipt.json"
    check = subprocess.run(
        ["python3", "tools/verify_full_path.py", "--plan", str(path), "--out", str(out)],
        capture_output=True, text=True)
    print(f"{name}：{'通過' if check.returncode == 0 else '失敗'}")
    if check.returncode != 0:
        print(check.stderr[:2000])
    return check.returncode == 0


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("names", nargs="*", help="收據名，例如 full-chain")
    p.add_argument("--all", action="store_true", help="重跑所有長路徑收據")
    args = p.parse_args()

    names = args.names
    if args.all:
        names = sorted(p.stem for p in PLANS.glob("*.json")
                       if not p.stem.endswith("-receipt")
                       and "trace" in json.loads(p.read_text(encoding="utf-8")))
    if not names:
        p.error("要指定收據名，或加 --all")
    # 沿用檢查器的依賴排序；先產生存檔，才重播讀它的後續路徑。
    from check_player_paths import dependency_order
    try:
        names, _ = dependency_order(names)
    except ValueError as error:
        p.error(str(error))
    writers = {}
    parents = {}
    for name in names:
        plan = json.loads((PLANS / f"{name}.json").read_text(encoding="utf-8"))
        if plan.get("save"):
            writers[plan["save"]["file"]] = name
        parents[name] = plan.get("load")
    bad = []
    for name in names:
        parent = writers.get(parents[name])
        if parent in bad:
            print(f"{name}：本輪存檔來源 {parent} 未通過，不讀上一輪保存結果")
            bad.append(name)
        elif not replay(name):
            bad.append(name)
    if bad:
        print("沒過：" + "、".join(bad), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
