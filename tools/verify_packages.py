#!/usr/bin/env python3
"""驗收 `dist-all/<版本>/packages/` 的六個發行檔。

玩家拿到的是這幾個檔，所以結構錯了在這裡就要紅，而不是等對方回報「打不開」。
檢查的是**打開之後拿得到什麼**：執行檔在不在、有沒有執行位元、素材有沒有跟著、
release 那一份是不是真的不含原作衍生位元組。

    verify_packages.py <dist 目錄> <版本>
    verify_packages.py --manifests-only <dist 目錄>   # 只重算三包的 SHA-256
"""

import hashlib
import json
import os
import re
import sys
import zipfile

# release 不得含原作衍生位元組。圖示是 remake 自己畫的純幾何圖形，是唯一的例外。
DERIVED_SUFFIXES = (".png", ".ogg", ".exe.bak", ".pdf", ".jpg", ".jpeg", ".icns")
ICON_NAMES = ("kol.png", ".DirIcon", "kol.icns")
VERSION_PATTERN = r"v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}"


def fail(msg: str) -> None:
    print(f"  ✗ {msg}")
    globals()["FAILURES"] += 1


def ok(msg: str, before: int) -> None:
    """這一項全過才報 ✓。失敗之後還印一行 ✓ 只會讓人以為它過了。

    判準是「這一項有沒有讓失敗數從 before 增加」。用布林旗標不行——它一旦翻起來
    就分不出是這一項失敗還是前一項失敗，兩種情況都會印出誤導的 ✓。
    """
    if globals()["FAILURES"] == before:
        print(f"  ✓ {msg}")


def check_zip(path: str, want_exec: str, bundle: str) -> None:
    name = os.path.basename(path)
    before = globals()["FAILURES"]
    with zipfile.ZipFile(path) as z:
        names = z.namelist()
        infos = {i.filename: i for i in z.infolist()}
    hit = [n for n in names if n.endswith(want_exec)]
    if not hit:
        fail(f"{name}：找不到 {want_exec}")
        return
    exe = hit[0]
    mode = (infos[exe].external_attr >> 16) & 0o777
    # Windows 不看執行位元，macOS 的 .app 沒有它就會被 Finder 判成損毀。
    if want_exec == "Contents/MacOS/kol" and not mode & 0o111:
        fail(f"{name}：{exe} 沒有執行位元（{oct(mode)}），Finder 會說檔案損毀")
    if want_exec == "Contents/MacOS/kol" and not any(
        n.endswith("Contents/Info.plist") for n in names
    ):
        fail(f"{name}：.app 少了 Info.plist")
    assets = [n for n in names if "/assets/" in n]
    if bundle == "full" and not assets:
        fail(f"{name}：full 應該帶素材，一個都沒有")
    if bundle == "release":
        bad = [
            n
            for n in names
            if n.lower().endswith(DERIVED_SUFFIXES)
            and os.path.basename(n) not in ICON_NAMES
        ]
        if bad:
            fail(f"{name}：release 帶了不該散布的檔案，例如 {bad[:3]}")
    if not any(n.endswith("如何開始.txt") for n in names):
        fail(f"{name}：少了給玩家的說明檔")
    ok(f"{name}：{len(names)} 個項目，素材 {len(assets)}", before)


def check_appimage(path: str, bundle: str) -> None:
    name = os.path.basename(path)
    before = globals()["FAILURES"]
    if not os.access(path, os.X_OK):
        fail(f"{name}：沒有執行位元，玩家下載後不能直接跑")
    with open(path, "rb") as fh:
        head = fh.read(4)
    # type-2 AppImage 是 ELF；前面接的 runtime 就是一支 static-pie 執行檔。
    if head[:4] != b"\x7fELF":
        fail(f"{name}：不是 ELF，runtime 沒有嵌進去")
        return
    ok(f"{name}：{os.path.getsize(path) // 1024} KB", before)


def check_manifest(dist: str, bundle: str, version: str | None = None) -> None:
    """重算每個檔的 SHA-256 與大小，與 `<bundle>/manifest.json` 逐筆比對。

    manifest 由 `write_dist_manifest.py` 在打包時寫下；這裡回頭驗它，打包後被改過、
    少了、多了的檔案都算失敗。manifest 自己不在清單裡。
    """
    base = os.path.join(dist, bundle)
    before = globals()["FAILURES"]
    try:
        with open(os.path.join(base, "manifest.json"), encoding="utf-8") as fh:
            manifest = json.load(fh)
    except (OSError, ValueError) as err:
        fail(f"{bundle}/manifest.json 讀不到：{err}")
        return
    if manifest.get("schema") != "kol-dist-manifest-v1":
        fail(f"{bundle}：manifest schema 不符")
    if manifest.get("bundle") != bundle:
        fail(f"{bundle}：manifest bundle 不符")
    declared_version = manifest.get("version")
    if not isinstance(declared_version, str) or not re.fullmatch(VERSION_PATTERN, declared_version):
        fail(f"{bundle}：manifest 版號不符合完整版號契約")
    elif version is not None and declared_version != version:
        fail(f"{bundle}：manifest 版號 {declared_version} 與要求的 {version} 不符")
    listed = {row["path"]: row for row in manifest.get("files", [])}
    if manifest.get("file_count") != len(listed):
        fail(f"{bundle}：file_count {manifest.get('file_count')} 與清單 {len(listed)} 筆不符")
    on_disk = set()
    for root, _, names in os.walk(base):
        for n in names:
            rel = os.path.relpath(os.path.join(root, n), base)
            if rel != "manifest.json":
                on_disk.add(rel)
    for rel in sorted(on_disk - listed.keys()):
        fail(f"{bundle}：{rel} 不在 manifest 裡")
    for rel, row in sorted(listed.items()):
        path = os.path.join(base, rel)
        if not os.path.isfile(path):
            fail(f"{bundle}：manifest 列了 {rel}，檔案不存在")
            continue
        h = hashlib.sha256()
        with open(path, "rb") as fh:
            for chunk in iter(lambda: fh.read(1 << 20), b""):
                h.update(chunk)
        if h.hexdigest() != row.get("sha256"):
            fail(f"{bundle}：{rel} 的 SHA-256 與 manifest 不符")
        if os.path.getsize(path) != row.get("bytes"):
            fail(f"{bundle}：{rel} 的大小與 manifest 不符")
    ok(f"{bundle}/manifest.json：{len(listed)} 個檔案的 SHA-256 與大小相符", before)


def main() -> int:
    globals()["FAILURES"] = 0
    if len(sys.argv) == 3 and sys.argv[1] == "--manifests-only":
        for bundle in ("full", "release", "promo"):
            if os.path.isdir(os.path.join(sys.argv[2], bundle)):
                check_manifest(sys.argv[2], bundle)
        print("manifest 驗收" + ("未通過" if globals()["FAILURES"] else "通過"))
        return 1 if globals()["FAILURES"] else 0
    if len(sys.argv) != 3 or not re.fullmatch(VERSION_PATTERN, sys.argv[2]):
        print("用法：verify_packages.py <dist 目錄> v.<主版>.<次版>.<修訂版>-YYYYMMDD", file=sys.stderr)
        return 2
    dist, version = sys.argv[1], sys.argv[2]
    for bundle in ("full", "release", "promo"):
        print(f"== {bundle} manifest")
        check_manifest(dist, bundle, version)
    pkg = os.path.join(dist, "packages")
    for bundle in ("full", "release"):
        print(f"== {bundle}")
        check_appimage(
            os.path.join(pkg, f"kol-{version}-{bundle}-linux-x86_64.AppImage"), bundle
        )
        check_zip(
            os.path.join(pkg, f"kol-{version}-{bundle}-windows-amd64.zip"),
            "kol.exe",
            bundle,
        )
        check_zip(
            os.path.join(pkg, f"kol-{version}-{bundle}-macos-universal.zip"),
            "Contents/MacOS/kol",
            bundle,
        )
    if globals()["FAILURES"]:
        print(f"發行檔驗收未通過：{globals()['FAILURES']} 項")
        return 1
    print("發行檔驗收通過")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
