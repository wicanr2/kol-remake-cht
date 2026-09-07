#!/usr/bin/env python3
"""驗收 `dist-all/<版本>/packages/` 的六個發行檔。

玩家拿到的是這幾個檔，所以結構錯了在這裡就要紅，而不是等對方回報「打不開」。
檢查的是**打開之後拿得到什麼**：執行檔在不在、有沒有執行位元、素材有沒有跟著、
release 那一份是不是真的不含原作衍生位元組。

    verify_packages.py <dist 目錄> <版本>
"""

import os
import sys
import zipfile

# release 不得含原作衍生位元組。圖示是 remake 自己畫的純幾何圖形，是唯一的例外。
DERIVED_SUFFIXES = (".png", ".ogg", ".exe.bak", ".pdf", ".jpg", ".jpeg", ".icns")
ICON_NAMES = ("kol.png", ".DirIcon", "kol.icns")


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


def main() -> int:
    globals()["FAILURES"] = 0
    dist, version = sys.argv[1], sys.argv[2]
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
