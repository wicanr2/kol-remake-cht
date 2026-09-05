#!/usr/bin/env python3
"""驗證一條長的正常玩家路徑。

`verify_player_path.py` 逐步比對截圖，適合十幾二十步的短路徑；一條從標題走到
存檔的完整路徑有上千步，逐步存圖既不合理也沒有意義。這一支改成檢查兩樣東西：

1. **狀態軌跡**（`cmd/kol -trace` 寫出來的 JSON）：里程碑那幾步停在哪一個畫面、
   隊伍在城裡或世界地圖的哪一格、戰鬥走到哪一階段。軌跡是遊戲自己report 的狀態，
   不是從畫面猜的。
2. **檢查點截圖**：只留下幾張關鍵畫格，逐 byte 比對已經通過對拍的那一批。

另外，路徑若會寫存檔，把存檔讀回來逐欄比對。

  python3 tools/verify_full_path.py --plan docs/verification/player-path/full-chain.json
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
    """存檔內容的雜湊，**不含 `saved_at`**。

    存檔裡的 `saved_at` 是寫出來那一刻的時間，所以整份檔案的 SHA-256 每跑一次
    就換一個——記進收據等於每次重跑都製造一行沒有意義的差異，而且那一欄根本
    驗不了任何東西（重生不出同一個值）。改成把時間戳拿掉再算，收據才重生得出來。
    """
    stable = {k: v for k, v in doc.items() if k != "saved_at"}
    body = json.dumps(stable, ensure_ascii=False, sort_keys=True,
                      separators=(",", ":"))
    return hashlib.sha256(body.encode("utf-8")).hexdigest()


def dig(doc, path: str):
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


def subset(want, got, where: str, failures: list[str]) -> None:
    """want 裡出現的每個欄位都要與 got 相同；got 多出來的欄位不管。"""
    if isinstance(want, dict):
        if not isinstance(got, dict):
            failures.append(f"{where}：期望物件，實際 {got!r}")
            return
        for key, value in want.items():
            if key not in got:
                failures.append(f"{where}.{key}：軌跡裡沒有這個欄位")
                continue
            subset(value, got[key], f"{where}.{key}", failures)
        return
    if want != got:
        failures.append(f"{where}：期望 {want!r}，實際 {got!r}")


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--plan", required=True)
    p.add_argument("--root", default=".")
    p.add_argument("--out", default=None)
    args = p.parse_args()

    root = pathlib.Path(args.root)
    plan = json.loads(pathlib.Path(args.plan).read_text(encoding="utf-8"))
    tokens = plan["script"].split()
    failures: list[str] = []

    trace_path = root / plan["trace"]
    steps = []
    if not trace_path.exists():
        failures.append(f"缺少軌跡 {trace_path}")
    else:
        steps = json.loads(trace_path.read_text(encoding="utf-8"))["steps"]
        if len(steps) != len(tokens) + 1:
            failures.append(
                f"軌跡有 {len(steps)} 筆，腳本是 {len(tokens)} 個 token（應該 {len(tokens) + 1} 筆）")

    checked = []
    for m in plan.get("milestones", []):
        i = m["step"]
        if i >= len(steps):
            failures.append(f"里程碑第 {i} 步超出軌跡")
            continue
        want = {k: v for k, v in m.items() if k not in ("step", "note")}
        subset(want, steps[i], f"第 {i} 步", failures)
        entry = {"step": i, "token": "（起始）" if i == 0 else tokens[i - 1],
                 "note": m.get("note", "")}
        entry.update({k: steps[i].get(k) for k in ("scene", "town", "world", "battle")
                      if steps[i].get(k) is not None})
        checked.append(entry)

    shots = []
    for c in plan.get("checkpoints", []):
        path = root / plan["shot_dir"] / c["file"]
        if not path.exists():
            failures.append(f"缺少檢查點截圖 {path}")
            continue
        image = read_png(path)
        if [image.width, image.height] != plan["size"]:
            failures.append(f"{c['file']} 尺寸 {[image.width, image.height]} 不是 {plan['size']}")
        digest = sha256(path)
        row = {"step": c["step"], "file": c["file"], "sha256": digest,
               "shows": c.get("shows", "")}
        if "identical_to" in c:
            other = root / c["identical_to"]
            if not other.exists():
                failures.append(f"對拍基準不存在 {other}")
            else:
                same = sha256(other)
                row["identical_to"] = c["identical_to"]
                row["identical"] = same == digest
                if same != digest:
                    failures.append(f"{c['file']} 與 {c['identical_to']} 不同")
        shots.append(row)

    saved = None
    if "save" in plan:
        path = root / plan["save"]["file"]
        if not path.exists():
            failures.append(f"缺少存檔 {path}")
        else:
            doc = json.loads(path.read_text(encoding="utf-8"))
            saved = {"file": plan["save"]["file"],
                     "sha256_without_saved_at": save_digest(doc), "fields": {}}
            for key, want in plan["save"]["expect"].items():
                try:
                    got = dig(doc, key)
                except (KeyError, IndexError, TypeError):
                    failures.append(f"存檔缺少欄位 {key}")
                    continue
                saved["fields"][key] = got
                if got != want:
                    failures.append(f"存檔 {key} 是 {got!r}，期望 {want!r}")

    receipt = {
        "schema": "kol-full-path-receipt-v1",
        "name": plan["name"],
        "description": plan.get("description", ""),
        "seed": plan.get("seed"),
        "load": plan.get("load"),
        "token_count": len(tokens),
        "trace": plan["trace"],
        "trace_sha256": sha256(trace_path) if trace_path.exists() else None,
        "milestones": checked,
        "checkpoints": shots,
        "save": saved,
        "failures": failures,
    }
    # 這幾塊由 plan 帶進來，驗證器不檢查，只負責原樣抄進收據：換過的輸入與
    # 不宣稱的事情要跟著結論走，不能只留在腳本裡。
    for key in ("substituted_inputs", "not_claimed", "rights", "command"):
        if key in plan:
            receipt[key] = plan[key]
    text = json.dumps(receipt, ensure_ascii=False, indent=1) + "\n"
    if args.out:
        out = pathlib.Path(args.out)
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(text, encoding="utf-8")
    else:
        print(text, end="")
    if failures:
        for f in failures:
            print("失敗：" + f, file=sys.stderr)
        return 1
    return 0


sys.exit(main())
