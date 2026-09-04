#!/usr/bin/env python3
"""替 `dist-all/<版本>/` 的三個包各寫一份逐檔 SHA-256 的 manifest。

專案契約要求每次打包驗證內容、manifest、SHA-256、平台／架構、存檔位置、
素材 deny-list 與授權文件。這一支負責 manifest 那一項，其餘由
`tools/build_dist.sh` 做。

  tools/run_python_tool.sh tools/write_dist_manifest.py --dist dist-all/0.1.0 --version 0.1.0
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib

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
    args = p.parse_args()

    root = pathlib.Path(args.dist)
    for bundle in ("full", "release", "promo"):
        base = root / bundle
        if not base.is_dir():
            continue
        files = []
        for path in sorted(base.rglob("*")):
            if not path.is_file() or path.name == "manifest.json":
                continue
            files.append({"path": str(path.relative_to(base)),
                          "bytes": path.stat().st_size,
                          "sha256": digest(path)})
        manifest = {
            "schema": "kol-dist-manifest-v1",
            "version": args.version,
            "bundle": bundle,
            "rights": RIGHTS[bundle],
            "platform": ["linux/amd64（容器裡原生建置，Ebitengine 在 Linux 要 CGO）",
                         "windows/amd64（交叉編譯，Ebitengine 在 Windows 走 syscall "
                         "不用 CGO；**沒有真機驗收**）"],
            "save_location": "遊戲以 `-save` 指定；預設寫在工作目錄，不碰使用者家目錄",
            "file_count": len(files),
            "files": files,
        }
        (base / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
        print(f"{bundle}: {len(files)} 個檔案")
    return 0


raise SystemExit(main())
