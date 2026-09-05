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

ROOT = pathlib.Path("docs/verification")


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def replay(name: str) -> bool:
    base = ROOT / name
    spec = base / "replay.spec"
    receipt = json.loads((base / "receipt.json").read_text(encoding="utf-8"))
    if not spec.exists():
        print(f"{name}：沒有 replay.spec，無法重跑")
        return False
    line = [l for l in spec.read_text(encoding="utf-8").splitlines()
            if l.strip() and not l.startswith("#")][0]
    out_dir = pathlib.Path(line.split("|", 1)[0]).parent
    try:
        subprocess.run(["tools/run_kol_shots.sh", str(spec)], check=True,
                       stdout=subprocess.DEVNULL)
    except subprocess.CalledProcessError:
        print(f"{name}：跑不起來")
        return False
    ok = True
    for frame in receipt["frames"]:
        got = out_dir / f"step-{frame['step']:03d}.png"
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
    args = ap.parse_args()
    names = args.names
    if args.all:
        names = sorted(p.parent.name for p in ROOT.glob("battle-*/replay.spec"))
    if not names:
        ap.error("要指定收據名稱或 --all")
    bad = [n for n in names if not replay(n)]
    if bad:
        print("沒過：" + "、".join(bad))
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
