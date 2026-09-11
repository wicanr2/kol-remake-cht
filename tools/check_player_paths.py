#!/usr/bin/env python3
"""把收據的 plan 重跑一次，比對每個檢查點的截圖——**不覆寫收據**。

`verify_full_path.py` 比的是收據目錄裡那幾張圖與軌跡自不自洽，而它們全部來自
同一次產生，所以程式改了它照樣通過。`replay_player_path.py` 會重跑，但它把結果
寫回 `docs/verification/`，跑完收據就被換掉了，看不出有沒有變。

這一支重跑到 `workplace/replay-check/<名字>/`，再拿產出的截圖與收據目錄裡同名那
幾張逐 byte 比。不同就列出來並回非零。

  tools/check_player_paths.py world-quest town-mirror
  tools/check_player_paths.py --all

跑起來要 Docker（`run_kol_shots.sh` 自己會開容器），所以這一支在主機上執行。
長路徑一條要跑十幾二十分鐘，`--all` 會很久。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import subprocess
import sys

PLANS = pathlib.Path("docs/verification/player-path")
OUT = pathlib.Path("workplace/replay-check")


def spec_line(plan: dict, shot_dir: pathlib.Path) -> str:
    """組一行規格檔。除了輸出目錄，其餘旗標與 replay_player_path.py 相同。"""
    scene = "flow" if "trace" in plan else "play"
    args = [f"{shot_dir}/step-000.png", "-scene", scene]
    if plan.get("seed") is not None:
        args += ["-seed", str(plan["seed"])]
    if plan.get("load"):
        args += ["-load", plan["load"]]
    if plan.get("quest_sites"):
        args += ["-quest-sites", plan["quest_sites"]]
    if plan.get("monsters"):
        args += ["-monsters", plan["monsters"]]
    flag = "-flow-script" if scene == "flow" else "-play-script"
    args += [flag, plan["script"], "-shot-dir", str(shot_dir)]
    steps = sorted({c["step"] for c in plan.get("checkpoints", [])})
    if steps:
        args += ["-shot-steps", ",".join(str(s) for s in steps)]
    if plan.get("trace"):
        args += ["-trace", f"{shot_dir}/trace.json"]
    return "|".join(args)


def digest(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def check(name: str) -> bool:
    path = PLANS / f"{name}.json"
    plan = json.loads(path.read_text(encoding="utf-8"))
    shot_dir = OUT / name
    shot_dir.mkdir(parents=True, exist_ok=True)
    spec = shot_dir / "shots.spec"
    spec.write_text(spec_line(plan, shot_dir) + "\n", encoding="utf-8")

    run = subprocess.run(["tools/run_kol_shots.sh", str(spec)],
                         capture_output=True, text=True)
    if run.returncode != 0:
        print(f"{name}：跑不動\n{run.stderr[-800:]}")
        return False

    kept = pathlib.Path(plan["shot_dir"])
    same, diff, missing = 0, [], []
    for shot in sorted(shot_dir.glob("step-*.png")):
        old = kept / shot.name
        if not old.exists():
            missing.append(shot.name)
            continue
        if digest(shot) == digest(old):
            same += 1
        else:
            diff.append(shot.name)
    ok = not diff and not missing
    print(f"{name}：{'一致' if ok else '對不上'}"
          f"（相同 {same}、不同 {len(diff)}、收據裡沒有 {len(missing)}）")
    for n in diff[:5]:
        print(f"    ✗ {n}")
    for n in missing[:5]:
        print(f"    ? {n}（重跑產出，收據裡沒有）")
    return ok


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("names", nargs="*", help="收據名，例如 world-quest")
    p.add_argument("--all", action="store_true", help="檢查全部")
    args = p.parse_args()
    names = args.names
    if args.all:
        names = sorted(f.stem for f in PLANS.glob("*.json")
                       if not f.stem.endswith("-receipt"))
    if not names:
        p.error("要指定收據名，或加 --all")
    bad = [n for n in names if not check(n)]
    if bad:
        print(f"\n對不上的：{'、'.join(bad)}")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
