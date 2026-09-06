#!/usr/bin/env python3
"""把原版 ZIP 解成一個可以餵給 dosgolem 的遊戲目錄。

dosgolem 的 `Load(exe, root)` 要一個「DOS 看到的當前目錄」——執行檔會用相對路徑
開 `TOWN0`、`KOLDATA` 那些檔。這支就是把 ZIP 攤平成那個樣子。

輸出**只留在本機**（`workplace/` 已經在 `.gitignore` 裡）：那是原版檔案本身，
不是衍生物，任何情況下都不進版控。

  tools/run_python_tool.sh tools/unpack_game_root.py --out workplace/oracle/game
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("workplace")
ROOT_IN_ZIP = "Knights-of-legend/"


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    out = pathlib.Path(args.out)
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(out.as_posix()).parents:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下（原版檔案不進版控）")
    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    out.mkdir(parents=True, exist_ok=True)
    written = []
    with zipfile.ZipFile(zip_path) as zf:
        for info in zf.infolist():
            if info.is_dir() or not info.filename.startswith(ROOT_IN_ZIP):
                continue
            name = info.filename[len(ROOT_IN_ZIP):]
            if "/" in name:  # 原版是平的一層，有子目錄就是解讀錯了
                raise SystemExit(f"ZIP 裡出現子目錄：{info.filename}")
            body = zf.read(info)
            (out / name).write_bytes(body)
            written.append({"name": name, "size": len(body),
                            "sha256": hashlib.sha256(body).hexdigest()})

    print(json.dumps({"files": len(written), "out": str(out),
                      "zip_sha256": digest}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
