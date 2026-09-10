#!/usr/bin/env python3
"""在 Docker 內核對完整保存 GUI 的來源與產物，產生本機收據。"""
import argparse
import hashlib
import json
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument("--root", type=Path, default=Path("/src"))
p.add_argument("--out", type=Path, default=Path("/out"))
a = p.parse_args()

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

before = json.loads((a.out / "source-before.json").read_text())
changed = [name for name, value in before.items() if digest(a.root / name) != value]
assert not changed, changed
assert (a.out / "legacy.json").read_bytes() == (a.out / "legacy-before.json").read_bytes()
paid = json.loads((a.out / "before-delete.json").read_text())
assert paid["characters"][14]["gold"] == 1440  # 畫面報價 60，起始金冠 1500。
assert not paid["party"]
results = (a.out / "gui-result.log").read_text()
for line in ("正常建角兩人、重啟選角與忠實住宿保存通過",
             "刪除不搬槽、重用空槽產生新身分及滿槽流程通過",
             "v1 匯入保留原檔、零值、傷勢與未知設定通過"):
    assert line in results
files = [f for f in a.out.iterdir() if f.is_file() and f.name != "receipt.json"]
owners = {str(f.relative_to(a.out)): [f.stat().st_uid, f.stat().st_gid] for f in files}
assert all(value == [1000, 1000] for value in owners.values()), owners
receipt = dict(schema="kol-complete-save-verification-v1", date="2026-09-08",
               image="kol-package:20260907",
               image_id="sha256:3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193",
               source_changed_during_run=False, sources_sha256=before,
               outputs_sha256={f.name: digest(f) for f in files}, owners=owners,
               input_counts={name: sum(line.startswith("input scene=") for line in
                             (a.out / (name + ".log")).read_text().splitlines())
                             for name in ["creation", "reload", "delete", "import-v1", "clock-reload"]},
               seed=1, seed_scope="remake GUI 重播；驗證保存，不宣稱亂數或原版同狀態對拍",
               visual_review=["新角色進入 BRT Trollsbane", "住宿報價 60、金冠 1500→1440、離隊詢問",
                              "v1 限制提示無缺字", "滿槽提示", "1102 年 2 月 28 日→3 月 1 日",
                              "世界重載道路與路標圖示"],
               limitations=["非 AppImage 驗收", "非原版戰鬥對拍", "世界時間 GUI 使用明示的受控 v2 輸入"],
               receipt_tool_sha256=digest(Path(__file__)))
target = a.out / "receipt.json"
if target.exists():
    assert target.stat().st_uid == 1000 and target.stat().st_gid == 1000
target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({"binary_sha256": digest(a.out / "kol"), "source_changed": False,
                  "input_counts": receipt["input_counts"], "output_count": len(files)}, ensure_ascii=False))
