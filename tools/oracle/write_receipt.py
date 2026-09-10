#!/usr/bin/env python3
"""容器內建立原版驗證輸入／輸出雜湊；本清冊不代替測試與畫面驗收。"""
import hashlib
import json
import subprocess
import sys
from pathlib import Path


def record(path):
    return {
        "path": str(path),
        "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
        "uid": path.stat().st_uid,
        "gid": path.stat().st_gid,
    }


out = Path("/out")
inputs = set()
for root in (
    "/src/internal", "/src/tools/oracle", "/repo/internal", "/repo/oracle",
):
    inputs.update(p for p in Path(root).rglob("*") if p.is_file() and p.suffix in (".go", ".py", ".sh"))
inputs.update(p for p in Path("/src/workplace/oracle/game").rglob("*") if p.is_file())
for name in ("/src/go.mod", "/src/go.sum", "/repo/go.mod",
             "/src/workplace/analysis/koltown/KOLTOWN-unpacked.EXE"):
    inputs.add(Path(name))
head = subprocess.check_output(
    ["git", "-c", "safe.directory=/repo", "-C", "/repo", "rev-parse", "HEAD"], text=True).strip()
data = {
    "schema": "kol-dosgolem-verification-inputs-v1",
    "dosgolem_head": head,
    "note": "包含未提交修正的來源雜湊；本清冊不判定測試通過或原版一致",
    "rights": "原版檔案及擷取圖片僅本機研究，未授權公開散布",
    "inputs": [record(p) for p in sorted(inputs)],
    "input_capture": "執行後工作樹清冊，不能單獨證明歷史執行來源",
    "outputs": [record(p) for p in sorted(out.iterdir()) if p.is_file() and p.name != "manifest.json"],
}
manifest = out / "manifest.json"
if "--start" in sys.argv:
    data["input_capture"] = "before_copy"
    data["outputs"] = []
elif manifest.exists():
    previous = json.loads(manifest.read_text())
    if previous.get("input_capture") == "before_copy":
        data["source_changed_during_run"] = previous["inputs"] != data["inputs"]
        data["inputs"] = previous["inputs"]
        data["input_capture"] = "before_copy"
(out / "manifest.json").write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")
