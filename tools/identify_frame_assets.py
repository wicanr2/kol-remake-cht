#!/usr/bin/env python3
"""辨識原版畫面是由哪些抽取資產組成的。

動畫對拍需要先知道原版每一幀畫了什麼。與其目視猜，不如拿抽取資產逐張去對：全螢幕
素材直接在原點比對，較小的素材在指定範圍內搜尋 blit 位置。只回報逐像素完全命中的
結果；有遮罩的素材只比對不透明像素。

  python3 tools/identify_frame_assets.py --frames 目錄 --archives BACKPICS,SET0PICS \\
      --assets assets/original-derived-png/ega [--search-small] [--out 結果.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import Image, read_png  # noqa: E402


def load_assets(root: pathlib.Path, archives: list[str]) -> list[dict]:
    manifest = json.loads((root / "manifest.json").read_text())
    out = []
    for a in manifest["archives"]:
        if a["archive"] not in archives:
            continue
        for img in a["images"]:
            out.append({"archive": a["archive"], "index": img["unit_index"],
                        "file": img["file"], "width": img["width"],
                        "height": img["height"]})
    return out


def match_at(frame: Image, frgb: bytes, asset: Image, argb: bytes,
             alpha: list[int] | None, ox: int, oy: int) -> bool:
    for i in range(asset.width * asset.height):
        if alpha is not None and not alpha[i]:
            continue
        ax, ay = i % asset.width, i // asset.width
        fo = ((oy + ay) * frame.width + ox + ax) * 3
        ao = i * 3
        if frgb[fo:fo + 3] != argb[ao:ao + 3]:
            return False
    return True


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--frames", required=True, help="含畫格 PNG 的目錄")
    p.add_argument("--assets", default="assets/original-derived-png/ega")
    p.add_argument("--archives", required=True, help="逗號分隔的 archive 名稱")
    p.add_argument("--search-small", action="store_true",
                   help="非全螢幕素材也搜尋 blit 位置（慢）")
    p.add_argument("--out", default=None)
    args = p.parse_args()

    root = pathlib.Path(args.assets)
    assets = load_assets(root, args.archives.split(","))
    cache: dict[str, tuple[Image, bytes, list[int] | None]] = {}
    for a in assets:
        img = read_png(root / a["file"])
        alpha = None
        if img.channels == 4:
            alpha = [img.pixels[i * 4 + 3] for i in range(img.width * img.height)]
            if all(alpha):
                alpha = None
        cache[a["file"]] = (img, img.rgb(), alpha)

    rows = []
    frames = sorted(pathlib.Path(args.frames).glob("frame-*.png"))
    for path in frames:
        frame = read_png(path)
        frgb = frame.rgb()
        hits = []
        for a in assets:
            img, argb, alpha = cache[a["file"]]
            if img.width > frame.width or img.height > frame.height:
                continue
            full = (img.width, img.height) == (frame.width, frame.height)
            if full or not args.search_small:
                if match_at(frame, frgb, img, argb, alpha, 0, 0):
                    hits.append({"archive": a["archive"], "index": a["index"],
                                 "x": 0, "y": 0})
                continue
            found = False
            for oy in range(frame.height - img.height + 1):
                for ox in range(frame.width - img.width + 1):
                    if match_at(frame, frgb, img, argb, alpha, ox, oy):
                        hits.append({"archive": a["archive"], "index": a["index"],
                                     "x": ox, "y": oy})
                        found = True
                        break
                if found:
                    break
        rows.append({"frame": path.name, "matches": hits})

    result = {
        "tool": "tools/identify_frame_assets.py",
        "frames_dir": args.frames,
        "archives": args.archives.split(","),
        "searched_small_assets": args.search_small,
        "frame_count": len(rows),
        "frames": rows,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    matched = sum(1 for r in rows if r["matches"])
    print(json.dumps({"frames": len(rows), "frames_with_match": matched},
                     ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
