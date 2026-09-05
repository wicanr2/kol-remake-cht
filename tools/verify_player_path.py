#!/usr/bin/env python3
"""驗證城鎮正常玩家路徑的截圖序列。

路徑由 `cmd/kol -scene play -play-script …` 產生：全部經過 `town.Session` 這個狀態機，
沒有 direct-entry、沒有座標注入、沒有測試旗標。本工具檢查：

1. 每一步的截圖存在、解得開、尺寸正確；
2. 相鄰兩步不得完全相同（每個 token 都要有可見效果）；
3. 指定的檢查點與已通過對拍的截圖**逐 byte 相同**——可玩路徑與對拍證據因此連上；
4. 路徑若會寫存檔（`save` 區塊），把寫出來的存檔讀回來逐欄比對——存檔的內容
   因此也是玩家路徑的一部分，不是另一條沒人走的支線；
5. 檢查點的 `same_as_except` 用來證明「只有這一塊變了」：指定矩形之外必須逐像素
   相同，矩形之內必須真的不同。讀檔路徑靠它證明載入的存檔確實換掉了畫面上的數字。

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


def save_digest(doc: dict) -> str:
    """存檔內容的雜湊，**不含 `saved_at`**（同 `verify_full_path.py`）。

    整份檔案的 SHA-256 每跑一次就換一個，因為 `saved_at` 是寫出來那一刻的時間；
    記進收據等於每次重跑都多一行沒有意義的差異，而那一欄也重生不出來。
    """
    stable = {k: v for k, v in doc.items() if k != "saved_at"}
    body = json.dumps(stable, ensure_ascii=False, sort_keys=True,
                      separators=(",", ":"))
    return hashlib.sha256(body.encode("utf-8")).hexdigest()


def dig(doc, path: str):
    """取 `party[0].gold` 這種欄位路徑；取不到就丟 KeyError／IndexError。"""
    node = doc
    for part in path.split("."):
        while "[" in part:
            name, rest = part.split("[", 1)
            index, part = rest.split("]", 1)
            if name:
                node = node[name]
            node = node[int(index)]
        if part:
            node = node[part]
    return node


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
        if "identical_to" in check:
            twin = root / check["identical_to"]
            if not (shot.exists() and twin.exists()):
                failures.append(f"檢查點 {check['step']} 的檔案缺少")
                continue
            if sha256(shot) != sha256(twin):
                failures.append(f"檢查點 {check['step']} 與 {check['identical_to']} 不同")
        if "same_as_except" in check:
            spec = check["same_as_except"]
            twin = root / spec["file"]
            if not (shot.exists() and twin.exists()):
                failures.append(f"檢查點 {check['step']} 的檔案缺少")
                continue
            a, b = read_png(shot), read_png(twin)
            if (a.width, a.height) != (b.width, b.height):
                failures.append(f"檢查點 {check['step']} 與 {spec['file']} 尺寸不同")
                continue
            pa, pb, w = a.rgb(), b.rgb(), a.width
            r = spec["rect"]
            outside, inside = 0, 0
            for y in range(a.height):
                for x in range(w):
                    i = (y * w + x) * 3
                    if pa[i:i + 3] == pb[i:i + 3]:
                        continue
                    if (r["x"] <= x < r["x"] + r["width"]
                            and r["y"] <= y < r["y"] + r["height"]):
                        inside += 1
                    else:
                        outside += 1
            check["differing_inside"] = inside
            check["differing_outside"] = outside
            if outside:
                failures.append(
                    f"檢查點 {check['step']} 在指定矩形之外有 {outside} 個像素與"
                    f"{spec['file']} 不同")
            if inside == 0:
                failures.append(
                    f"檢查點 {check['step']} 與 {spec['file']} 完全相同；"
                    "指定的矩形本來應該不一樣")

    save_check = None
    if "save" in plan:
        spec = plan["save"]
        path = root / spec["file"]
        fields = []
        if not path.exists():
            failures.append(f"路徑應該寫出存檔 {spec['file']}，但檔案不在")
        else:
            doc = json.loads(path.read_text(encoding="utf-8"))
            for field, want in spec["expect"].items():
                try:
                    got = dig(doc, field)
                except (KeyError, IndexError, TypeError):
                    got = None
                ok = got == want
                fields.append({"field": field, "want": want, "got": got, "pass": ok})
                if not ok:
                    failures.append(f"存檔的 {field} 是 {got!r}，應為 {want!r}")
        save_check = {"file": spec["file"], "fields": fields,
                      "sha256_without_saved_at":
                          save_digest(doc) if path.exists() else None}

    receipt = {
        "schema": "kol-player-path-v1",
        "name": plan["name"],
        "description": plan["description"],
        "script": plan["script"],
        "step_count": len(steps),
        "checkpoints": plan.get("checkpoints", []),
        "save": save_check,
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
