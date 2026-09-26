#!/usr/bin/env python3
"""重跑 `docs/verification/battle-*/` 那幾份收據，逐 byte 比對畫格。

長路徑收據有 `replay_player_path.py`，短的戰鬥收據沒有——它們的重跑指令原本只寫在
收據的 `command` 散文欄裡，改了戰鬥數值之後要人照著散文猜規格檔。這一支把
`replay.spec` 跑起來，再拿輸出的 PNG 對收據 `frames` 的 SHA-256 與版控裡那一張。

  tools/replay_battle_receipts.py battle-flee
  tools/replay_battle_receipts.py --all

跑起來要 Docker（`run_kol_shots.sh` 自己會開容器），所以這一支在主機上執行。
存檔與換過的資料檔在 `workplace/`（不進版控），先用收據 `command` 欄寫的工具重生。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from mechanics import locate  # noqa: E402

ROOT = pathlib.Path("docs/verification")


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def with_shots(line: str, steps: list[int]) -> str:
    """把一行規格的 `-shot-steps` 換成指定步數（沒有就補上），並確保有 `-trace` 與
    `-trace-detail`（指令列、游標等欄位只在細節模式才記）。"""
    parts = line.split("|")
    out, i = [], 0
    while i < len(parts):
        if parts[i] in ("-shot-steps", "-trace"):
            i += 2
            continue
        if parts[i] == "-trace-detail":
            i += 1
            continue
        out.append(parts[i])
        i += 1
    shot_dir = pathlib.Path(parts[0]).parent
    out += ["-shot-steps", ",".join(str(s) for s in steps) or "0",
            "-trace", str(shot_dir / "trace.json"), "-trace-detail"]
    return "|".join(out)


def run_lines(lines: list[str], spec: pathlib.Path) -> bool:
    spec.parent.mkdir(parents=True, exist_ok=True)
    spec.write_text("\n".join(lines) + "\n", encoding="utf-8")
    run = subprocess.run(["tools/run_kol_shots.sh", str(spec)],
                         stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, text=True)
    if run.returncode != 0:
        print(run.stderr[-600:])
    return run.returncode == 0


def replay_mechanics(name: str, base: pathlib.Path, receipt: dict,
                     lines: list[str], write: bool) -> bool:
    """機制收據：跑一次拿軌跡，逐條找斷言；`--write` 時照找到的步數重截示意圖。

    截圖不參與通過與否（2026-09-26 定案：驗機制，不驗像素）。"""
    work = pathlib.Path("workplace/tmp/battle-mechanics") / name
    traced = [with_shots(l, [0]) for l in lines]
    if not run_lines(traced, work / "trace.spec"):
        print(f"{name}：跑不起來")
        return False
    ok = True
    located: list[list[int]] = []
    for n, run in enumerate(receipt["runs"]):
        trace = pathlib.Path(traced[n].split("|", 1)[0]).parent / "trace.json"
        steps = json.loads(trace.read_text(encoding="utf-8"))["steps"]
        found, failures = locate(run["checks"], steps)
        located.append(found)
        for f in failures:
            print(f"失敗：{name}（{run.get('name', n)}）{f}")
        ok = ok and not failures
        run["found"] = found
    if ok and write:
        shots = []
        for n, line in enumerate(lines):
            wanted = sorted({located[n][f["check"]] for f in receipt["frames"]
                             if f.get("run", 0) == n})
            shots.append(with_shots(line, wanted))
        if not run_lines(shots, work / "shots.spec"):
            print(f"{name}：截圖那一次跑不起來")
            return False
        for f in receipt["frames"]:
            n = f.get("run", 0)
            step = located[n][f["check"]]
            src = pathlib.Path(shots[n].split("|", 1)[0]).parent / f"step-{step:03d}.png"
            (base / f["file"]).write_bytes(src.read_bytes())
            f["step"] = step
            f["sha256"] = sha256(base / f["file"])
        (base / "receipt.json").write_text(
            json.dumps(receipt, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(f"{name}：{'通過' if ok else '失敗'}")
    return ok


def replay(name: str, write: bool = False) -> bool:
    base = ROOT / name
    spec = base / "replay.spec"
    receipt = json.loads((base / "receipt.json").read_text(encoding="utf-8"))
    if not spec.exists():
        print(f"{name}：沒有 replay.spec，無法重跑")
        return False
    if "runs" in receipt:
        lines = [l for l in spec.read_text(encoding="utf-8").splitlines()
                 if l.strip() and not l.startswith("#")]
        return replay_mechanics(name, base, receipt, lines, write)
    lines = [l for l in spec.read_text(encoding="utf-8").splitlines()
             if l.strip() and not l.startswith("#")]
    # 一份規格檔可以有好幾行（同一段路的不同分支），畫格用 `spec_line` 指定
    # 自己是哪一行跑出來的；沒寫就是第 0 行。
    out_dirs = [pathlib.Path(l.split("|", 1)[0]).parent for l in lines]
    try:
        subprocess.run(["tools/run_kol_shots.sh", str(spec)], check=True,
                       stdout=subprocess.DEVNULL)
    except subprocess.CalledProcessError:
        print(f"{name}：跑不起來")
        return False
    ok = True
    for frame in receipt["frames"]:
        got = out_dirs[frame.get("spec_line", 0)] / f"step-{frame['step']:03d}.png"
        if not got.exists():
            print(f"失敗：{name} 第 {frame['step']} 步沒有截到圖")
            ok = False
            continue
        digest = sha256(got)
        if digest != frame["sha256"]:
            print(f"失敗：{name} {frame['file']}（第 {frame['step']} 步）"
                  f"雜湊不同\n  收據 {frame['sha256']}\n  這一次 {digest}")
            ok = False
            continue
        kept = base / frame["file"]
        if sha256(kept) != digest:
            print(f"失敗：{name} 版控裡的 {frame['file']} 與收據的雜湊不同")
            ok = False
    print(f"{name}：{'通過' if ok else '失敗'}")
    return ok


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("names", nargs="*")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--write", action="store_true",
                    help="機制收據通過後，照找到的步數重截示意圖並寫回收據")
    args = ap.parse_args()
    names = args.names
    if args.all:
        names = sorted(p.parent.name for p in ROOT.glob("battle-*/replay.spec"))
    if not names:
        ap.error("要指定收據名稱或 --all")
    bad = [n for n in names if not replay(n, args.write)]
    if bad:
        print("沒過：" + "、".join(bad))
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
