#!/usr/bin/env python3
"""替 remake 展示截圖產生 provenance manifest 並就地驗證。

展示圖不是隨手截的：每一張都由 `docs/screenshots/shots.spec` 的一行指令重生，
manifest 記下產生指令、輸入資產的雜湊、PNG 尺寸與 SHA-256、以及權利分類。
驗證在這裡一起做——PNG 解得開、尺寸符合模式、顏色數不是一片死白，
並且與已通過對拍的同狀態截圖逐 byte 相同（有對應的話）。

  python3 tools/build_screenshot_manifest.py --spec docs/screenshots/shots.spec \\
      --out docs/screenshots/MANIFEST.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402

# 與已通過對拍的同狀態截圖比對：展示圖不得和對拍過的畫面不一樣。
PARITY_TWINS = {
    "docs/screenshots/en-book-cover.png":
        ("docs/verification/remake-parity/shots/cover.png",
         "docs/verification/remake-parity/book-screens.json"),
    "docs/screenshots/en-book-contents.png":
        ("docs/verification/remake-parity/shots/contents.png",
         "docs/verification/remake-parity/book-screens.json"),
    "docs/screenshots/en-town-inn.png":
        ("docs/verification/remake-parity/shots/inn.png",
         "docs/verification/remake-parity/town-inn.json"),
    "docs/screenshots/en-town-inn-party2.png":
        ("docs/verification/remake-parity/shots/town-inn/recruited.png",
         "docs/verification/remake-parity/town-inn.json"),
    "docs/screenshots/en-book-harbor.png":
        ("docs/verification/remake-parity/shots/harbor.png",
         "docs/verification/remake-parity/harbor.json"),
}

EXPECTED_SIZE = {"en": (320, 200), "zh": (640, 400)}


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def distinct_colours(path: pathlib.Path) -> int:
    img = read_png(path)
    rgb = img.rgb()
    return len({bytes(rgb[i:i + 3]) for i in range(0, len(rgb), 3)})


def parse_spec(path: pathlib.Path) -> list[dict]:
    out = []
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split("|")
        out.append({"file": parts[0], "args": parts[1:]})
    return out


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--spec", default="docs/screenshots/shots.spec")
    p.add_argument("--out", default="docs/screenshots/MANIFEST.json")
    p.add_argument("--assets", default="assets/original-derived-png/ega/manifest.json")
    p.add_argument("--fonts", default="assets/original-derived-png/fonts/manifest.json")
    p.add_argument("--ui-fonts", default="assets/ui-font/manifest.json")
    args = p.parse_args()

    entries = []
    failures = []
    for item in parse_spec(pathlib.Path(args.spec)):
        path = pathlib.Path(item["file"])
        if not path.exists():
            failures.append({"file": item["file"], "reason": "檔案不存在"})
            continue
        img = read_png(path)
        mode = "zh" if path.name.startswith("zh-") else "en"
        want = EXPECTED_SIZE[mode]
        colours = distinct_colours(path)
        entry = {
            "file": item["file"],
            "mode": "中文模式（不對拍）" if mode == "zh" else "英文模式（逐畫面對拍）",
            "command": "cmd/kol " + " ".join(item["args"]),
            "width": img.width, "height": img.height,
            "distinct_colours": colours,
            "sha256": sha256(path),
        }
        ok = True
        if (img.width, img.height) != want:
            entry["error"] = f"尺寸 {img.width}x{img.height} 不是 {want[0]}x{want[1]}"
            ok = False
        if colours < 4:
            entry["error"] = f"只有 {colours} 種顏色，可能是空白畫面"
            ok = False
        twin = PARITY_TWINS.get(item["file"])
        if twin:
            twin_path = pathlib.Path(twin[0])
            entry["parity_twin"] = twin[0]
            entry["parity_receipt"] = twin[1]
            entry["identical_to_parity_shot"] = (
                twin_path.exists() and sha256(twin_path) == entry["sha256"])
            if not entry["identical_to_parity_shot"]:
                entry["error"] = "與對拍過的同狀態截圖不同"
                ok = False
        entry["pass"] = ok
        if not ok:
            failures.append(entry)
        entries.append(entry)

    commit = subprocess.run(["git", "rev-parse", "HEAD"], capture_output=True,
                            text=True).stdout.strip()
    receipt = {
        "schema": "kol-remake-screenshots-v1",
        "tool": "tools/build_screenshot_manifest.py",
        "spec": args.spec,
        "runner": "tools/run_kol_shots.sh",
        "producer": {
            "container_image": "eob-remake-go:1.26.7-ebiten2.9.9-audio",
            "renderer": "Xvfb ＋ Mesa llvmpipe 軟體算繪，320×200／640×400 離屏畫布",
            "commit": commit,
        },
        "inputs": {
            "ega_assets_manifest_sha256": sha256(pathlib.Path(args.assets)),
            "fonts_manifest_sha256": sha256(pathlib.Path(args.fonts)),
            "ui_font_manifest_sha256": sha256(pathlib.Path(args.ui_fonts)),
        },
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "畫面由 remake 繪製，但內容含原版衍生素材（BOOKPICS／CHARPICS／"
                    "KOLMPICS／TWNxPICS 等抽出的 PNG）與原版點陣字型，著作權仍屬原權利人。"
                    "使用者 2026-09-01 的 private-repo 例外只涵蓋 PNG，不等於公開散布權；"
                    "repo 轉 public、public fork、GitHub Release 或公開 release 之前必須"
                    "重新做權利審查。中文字模表另衍生自 Noto Sans CJK（OFL 1.1）。",
        },
        "criterion": "每張圖都要能解碼、尺寸符合該模式、顏色數不少於 4；"
                     "有對應對拍收據的畫面必須與那張逐 byte 相同",
        "count": len(entries),
        "failed": len(failures),
        "screenshots": entries,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"count": len(entries), "failed": len(failures)}))
    return 0 if not failures else 1


if __name__ == "__main__":
    raise SystemExit(main())
