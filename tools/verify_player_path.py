#!/usr/bin/env python3
"""驗證城鎮正常玩家路徑的截圖序列。

路徑由 `cmd/kol -scene play -play-script …` 產生：全部經過 `town.Session` 這個狀態機，
沒有 direct-entry、沒有座標注入、沒有測試旗標。本工具檢查：

1. 每一步的截圖存在、解得開、尺寸正確；
2. 相鄰兩步不得完全相同（每個 token 都要有可見效果）；
3. 指定的檢查點與已通過對拍的截圖**逐 byte 相同**——可玩路徑與對拍證據因此連上。

  python3 tools/verify_player_path.py --plan docs/verification/player-path/town.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from png_io import read_png  # noqa: E402


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--plan", required=True)
    p.add_argument("--root", default=".")
    p.add_argument("--out", default=None)
    args = p.parse_args()

    root = pathlib.Path(args.root)
    plan = json.loads(pathlib.Path(args.plan).read_text(encoding="utf-8"))
    tokens = plan["script"].split()
    shots = [root / plan["shot_dir"] / f"step-{i:03d}.png" for i in range(len(tokens) + 1)]

    steps, failures = [], []
    previous = None
    for i, shot in enumerate(shots):
        if not shot.exists():
            failures.append(f"缺少 {shot}")
            continue
        image = read_png(shot)
        size = [image.width, image.height]
        if size != plan["size"]:
            failures.append(f"{shot.name} 尺寸 {size} 不是 {plan['size']}")
        digest = sha256(shot)
        entry = {"step": i, "token": "（起始）" if i == 0 else tokens[i - 1],
                 "file": str(shot.relative_to(root)), "sha256": digest}
        if previous is not None and digest == previous:
            entry["unchanged"] = True
            failures.append(f"{shot.name} 與前一步完全相同：token 沒有效果")
        previous = digest
        steps.append(entry)

    for check in plan.get("checkpoints", []):
        shot = root / plan["shot_dir"] / f"step-{check['step']:03d}.png"
        twin = root / check["identical_to"]
        if not (shot.exists() and twin.exists()):
            failures.append(f"檢查點 {check['step']} 的檔案缺少")
            continue
        if sha256(shot) != sha256(twin):
            failures.append(f"檢查點 {check['step']} 與 {check['identical_to']} 不同")

    receipt = {
        "schema": "kol-player-path-v1",
        "name": plan["name"],
        "description": plan["description"],
        "script": plan["script"],
        "step_count": len(steps),
        "checkpoints": plan.get("checkpoints", []),
        "failures": failures,
        "steps": steps,
    }
    out = pathlib.Path(args.out or plan["receipt"])
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"steps": len(steps), "failures": len(failures)}, ensure_ascii=False))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
