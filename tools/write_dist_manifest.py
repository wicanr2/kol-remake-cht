#!/usr/bin/env python3
"""替 `dist-all/<版本>/` 的三個包各寫一份逐檔 SHA-256 的 manifest。

專案契約要求每次打包驗證內容、manifest、SHA-256、平台／架構、存檔位置、
素材 deny-list 與授權文件。這一支負責 manifest 那一項，其餘由
`tools/build_dist.sh` 做。

  tools/run_python_tool.sh tools/write_dist_manifest.py --dist dist-all/v.1.0.0-20260901 --version v.1.0.0-20260901
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import re

# 三個包各自的權利分類，寫進 manifest 免得日後靠記憶。
RIGHTS = {
    "full": "含原作衍生素材，僅本機保存；不得散布。",
    "release": "只有 remake 自己的程式碼、文件與抽取器，不含原版衍生位元組。",
    "promo": "推廣素材；每一項都要另外標來源與影音驗收。",
}


def digest(path: pathlib.Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--dist", required=True)
    p.add_argument("--version", required=True)
    p.add_argument("--platform", action="append", required=True,
                   choices=("linux/amd64", "windows/amd64", "darwin/amd64", "darwin/arm64"))
    args = p.parse_args()
    if not re.fullmatch(r"v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}", args.version):
        p.error("版號必須使用 v.<主版>.<次版>.<修訂版>-YYYYMMDD")

    root = pathlib.Path(args.dist)
    for bundle in ("full", "release", "promo"):
        base = root / bundle
        if not base.is_dir():
            continue
        files = []
        for path in sorted(base.rglob("*")):
            if not path.is_file() or path == base / "manifest.json":
                continue
            files.append({"path": str(path.relative_to(base)),
                          "bytes": path.stat().st_size,
                          "sha256": digest(path)})
        manifest = {
            "schema": "kol-dist-manifest-v1",
            "version": args.version,
            "bundle": bundle,
            "rights": RIGHTS[bundle],
            "platform": args.platform,
            "validation": "清單僅記錄檔案與目標平台；啟動、存讀檔及真機驗收須另附收據。",
            "save_location": "互動模式預設使用作業系統的使用者設定目錄下 kol-remake-cht/save.dat；可用 -save 指定。",
            "file_count": len(files),
            "files": files,
        }
        (base / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
        print(f"{bundle}: {len(files)} 個檔案")
    return 0


raise SystemExit(main())
