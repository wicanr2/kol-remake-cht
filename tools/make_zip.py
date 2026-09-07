#!/usr/bin/env python3
"""把一個目錄封成 zip，保留 Unix 執行權限。

容器裡沒有 `zip`，而且 `zip` 預設也不會保留 macOS `.app` 內那支執行檔的執行位元
——解開之後不能執行的 `.app`，Finder 只會說檔案損毀。這支用 zipfile 明寫
external_attr，並把目錄項也寫進去，好讓解開後的結構與權限與來源一致。

    make_zip.py <輸出.zip> <來源目錄> [封裝內的頂層名稱]
"""

import os
import sys
import zipfile


def add(zf: zipfile.ZipFile, src: str, arcroot: str) -> None:
    for root, dirs, files in os.walk(src):
        dirs.sort()
        rel = os.path.relpath(root, src)
        base = arcroot if rel == "." else os.path.join(arcroot, rel)
        for name in sorted(dirs):
            info = zipfile.ZipInfo(os.path.join(base, name) + "/")
            info.external_attr = (0o40755 << 16) | 0x10
            zf.writestr(info, b"")
        for name in sorted(files):
            path = os.path.join(root, name)
            if os.path.islink(path):
                info = zipfile.ZipInfo(os.path.join(base, name))
                info.create_system = 3
                # 0xA1FF = symlink + 0777，解開時才會還原成連結而不是一般檔案。
                info.external_attr = 0xA1FF0000
                zf.writestr(info, os.readlink(path))
                continue
            mode = os.stat(path).st_mode & 0o7777
            info = zipfile.ZipInfo.from_file(path, os.path.join(base, name))
            info.external_attr = (mode << 16) | 0o600
            info.compress_type = zipfile.ZIP_DEFLATED
            with open(path, "rb") as fh:
                zf.writestr(info, fh.read())


def main() -> int:
    if len(sys.argv) < 3:
        print(__doc__, file=sys.stderr)
        return 2
    out, src = sys.argv[1], sys.argv[2]
    arcroot = sys.argv[3] if len(sys.argv) > 3 else os.path.basename(src.rstrip("/"))
    os.makedirs(os.path.dirname(os.path.abspath(out)), exist_ok=True)
    tmp = out + ".partial"
    with zipfile.ZipFile(tmp, "w", zipfile.ZIP_DEFLATED) as zf:
        add(zf, src, arcroot)
    os.replace(tmp, out)
    print(out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
