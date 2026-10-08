#!/usr/bin/env python3
"""把收據的 plan 重跑一次，比對狀態軌跡——**不覆寫收據**。

`verify_full_path.py` 比的是收據目錄裡那幾張圖與軌跡自不自洽，而它們全部來自
同一次產生，所以程式改了它照樣通過。`replay_player_path.py` 會重跑，但它把結果
寫回 `docs/verification/`，跑完收據就被換掉了，看不出有沒有變。

這一支重跑到 `workplace/replay-check/<名字>/`，拿產出的**狀態軌跡**（`-trace`）與收據
目錄裡的參考軌跡逐步比：場景、城鎮／野外位置與模式、提示、戰鬥階段與訊息、隊伍數值、
結局。不同就列出來並回非零。2026-09-26 定案：收據驗機制，不要求畫面逐位元相同——
截圖差異只列出數量當參考，不算失敗。

短路徑原本沒有軌跡；`--bootstrap` 把這一次跑出來的軌跡寫成參考（只補缺少的）。

  tools/check_player_paths.py world-quest town-mirror
  tools/check_player_paths.py --all

必須在 Docker 內執行 Python；容器內設 `KOL_BIN` 直接使用已建置的執行檔，
`KOL_REPLAY_CHECK_DIR` 指定可寫輸出。`--all` 依存檔依賴排序，續玩讀當輪新存檔。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import subprocess
import sys

PLANS = pathlib.Path("docs/verification/player-path")
OUT = pathlib.Path(os.environ.get("KOL_REPLAY_CHECK_DIR", "workplace/replay-check"))


def spec_line(plan: dict, shot_dir: pathlib.Path) -> str:
    """組一行規格檔。除了輸出目錄，其餘旗標與 replay_player_path.py 相同。"""
    # plan 帶齊重跑所需的旗標才重得出來（README 的教訓）。`scene` 是其中一個：
    # `full` 沒有 trace 卻要走 flow（腳本從標題開始），光看 trace 猜會猜錯。
    scene = plan.get("scene") or ("flow" if "trace" in plan else "play")
    args = [f"{shot_dir}/step-000.png", "-scene", scene]
    if plan.get("seed") is not None:
        args += ["-seed", str(plan["seed"])]
    if plan.get("load"):
        args += ["-load", plan["load"]]
    if plan.get("quest_sites"):
        args += ["-quest-sites", plan["quest_sites"]]
    if plan.get("monsters"):
        args += ["-monsters", plan["monsters"]]
    # 受控亂數（CLAUDE.md 允許的「明示的受控亂數輸入」）：plan 寫明才帶，收據照抄。
    if plan.get("ambush_rolls"):
        args += ["-ambush-rolls", plan["ambush_rolls"]]
    # 會寫存檔的那幾條要把 -save 帶上，否則下一條讀不到它的產出
    # （`town-load` 讀的就是 `town-rest` 寫出來的那一份）。寫到暫存目錄，
    # 不動 workplace/saves 裡那幾份現成的。
    if plan.get("save"):
        args += ["-save", str(shot_dir / "save.json")]
    flag = "-flow-script" if scene == "flow" else "-play-script"
    args += [flag, plan["script"], "-shot-dir", str(shot_dir)]
    # 短路徑逐步存圖、逐張比；只有長路徑（有 trace）才只截檢查點那幾張。
    steps = sorted({c["step"] for c in plan.get("checkpoints", [])})
    if steps and plan.get("trace"):
        args += ["-shot-steps", ",".join(str(s) for s in steps)]
    args += ["-trace", f"{shot_dir}/trace.json"]
    return "|".join(args)


def reference_trace(plan: dict) -> pathlib.Path:
    return pathlib.Path(plan.get("trace") or f"{plan['shot_dir']}/trace.json")


def state(step: dict) -> dict:
    """一步的狀態：軌跡整筆去掉步名（步名只是畫格檔名）。"""
    return {k: v for k, v in step.items() if k != "name"}


def compare_traces(ref: list[dict], got: list[dict]) -> list[str]:
    out = []
    if len(ref) != len(got):
        out.append(f"軌跡長度不同：參考 {len(ref)} 步、這一次 {len(got)} 步")
    for i, (a, b) in enumerate(zip(ref, got)):
        a, b = state(a), state(b)
        if a == b:
            continue
        keys = sorted(k for k in set(a) | set(b) if a.get(k) != b.get(k))
        detail = "；".join(f"{k}: {json.dumps(a.get(k), ensure_ascii=False)[:120]} → "
                          f"{json.dumps(b.get(k), ensure_ascii=False)[:120]}" for k in keys[:2])
        out.append(f"第 {i} 步 {detail}")
        if len(out) >= 4:
            break
    return out


def digest(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def dependency_order(names: list[str]) -> tuple[list[str], dict[str, str]]:
    """先執行存檔產生者；拒絕重複寫入者與循環，不偷偷讀上一輪存檔。"""
    plans = {n: json.loads((PLANS / f"{n}.json").read_text(encoding="utf-8")) for n in names}
    writers = {}
    for name, plan in plans.items():
        save = plan.get("save", {}).get("file")
        if save:
            if save in writers:
                raise ValueError(f"同一存檔有多個產生者：{save}")
            writers[save] = name
    ordered, visiting, visited = [], set(), set()

    def visit(name):
        if name in visiting:
            raise ValueError(f"玩家路徑存檔依賴有循環：{name}")
        if name in visited:
            return
        visiting.add(name)
        parent = writers.get(plans[name].get("load"))
        if parent:
            visit(parent)
        visiting.remove(name)
        visited.add(name)
        ordered.append(name)

    for name in names:
        visit(name)
    return ordered, {path: str(OUT / name / "save.json") for path, name in writers.items()}


def check(name: str, bootstrap: bool = False, fresh_saves: dict[str, str] | None = None) -> bool:
    path = PLANS / f"{name}.json"
    plan = json.loads(path.read_text(encoding="utf-8"))
    if fresh_saves and plan.get("load") in fresh_saves:
        plan["load"] = fresh_saves[plan["load"]]
        independent_save = False
    else:
        independent_save = bool(plan.get("load"))
    shot_dir = OUT / name
    shot_dir.mkdir(parents=True, exist_ok=True)
    if independent_save:
        # v1 匯入會在輸入旁寫轉換檔。複製相同位元組到輸出區，保留觀察來源唯讀。
        source = pathlib.Path(plan["load"])
        copied = shot_dir / "input-save.json"
        copied.write_bytes(source.read_bytes())
        plan["load"] = str(copied)
    spec = shot_dir / "shots.spec"
    spec.write_text(spec_line(plan, shot_dir) + "\n", encoding="utf-8")

    binary = os.environ.get("KOL_BIN")
    command = ([binary, *spec_line(plan, shot_dir).split("|")[1:]] if binary else
               ["tools/run_kol_shots.sh", str(spec)])
    run = subprocess.run(command, capture_output=True, text=True)
    (shot_dir / "engine.log").write_text(run.stdout + run.stderr, encoding="utf-8")
    if run.returncode != 0:
        print(f"{name}：跑不動\n{run.stderr[-800:]}")
        return False

    got_path = shot_dir / "trace.json"
    got = json.loads(got_path.read_text(encoding="utf-8"))["steps"]
    ref_path = reference_trace(plan)
    if not ref_path.exists():
        if not bootstrap:
            print(f"{name}：沒有參考軌跡 {ref_path}（先跑 --bootstrap）")
            return False
        ref_path.write_bytes(got_path.read_bytes())
        print(f"{name}：寫出參考軌跡 {ref_path}（{len(got)} 步）")
    ref = json.loads(ref_path.read_text(encoding="utf-8"))["steps"]
    problems = compare_traces(ref, got)

    # 截圖只當參考：數一下有幾張與收據目錄不同，不算失敗。
    kept = pathlib.Path(plan["shot_dir"])
    shots = sorted(shot_dir.glob("step-*.png"))
    pixel = sum(1 for shot in shots
                if (kept / shot.name).exists() and digest(shot) != digest(kept / shot.name))
    ok = not problems
    print(f"{name}：{'狀態一致' if ok else '狀態不同'}（{len(got)} 步；"
          f"截圖與收據不同 {pixel} 張，僅供參考）")
    for line in problems:
        print(f"    ✗ {line}")
    return ok


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("names", nargs="*", help="收據名，例如 world-quest")
    p.add_argument("--all", action="store_true", help="檢查全部")
    p.add_argument("--bootstrap", action="store_true",
                   help="缺少參考軌跡的路徑，把這一次的軌跡寫成參考")
    args = p.parse_args()
    names = args.names
    if args.all:
        names = sorted(f.stem for f in PLANS.glob("*.json")
                       if not f.stem.endswith("-receipt"))
    if not names:
        p.error("要指定收據名，或加 --all")
    try:
        names, fresh_saves = dependency_order(names)
    except ValueError as error:
        p.error(str(error))
    bad = [n for n in names if not check(n, args.bootstrap, fresh_saves)]
    if bad:
        print(f"\n對不上的：{'、'.join(bad)}")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
