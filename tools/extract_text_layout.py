#!/usr/bin/env python3
"""用原版字型在原版畫面上機械定位一組字串，輸出可版控的文字版面表。

中文化要把譯文放進原版的同一個文字框，remake 也要照原版座標畫字，因此需要一份
「哪個字串畫在哪、什麼顏色」的表。這支工具用抽出的字型逐位置比對，命中必須唯一；
任何字串命中 0 次或多於一次都算失敗，避免把猜測寫成版面資料。

  python3 tools/extract_text_layout.py --plan 版面計畫.json --out 版面.json
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from locate_text_in_frame import load_font, text_ink  # noqa: E402
from png_io import read_png  # noqa: E402


def locate(frame, rgb, font, text: str, min_contrast: float) -> list[dict]:
    ink, hole, w, h = text_ink(font, text)
    hits = []
    for oy in range(frame.height - h + 1):
        for ox in range(frame.width - w + 1):
            colour = None
            ok = True
            for x, y in ink:
                o = ((oy + y) * frame.width + ox + x) * 3
                c = (rgb[o], rgb[o + 1], rgb[o + 2])
                if colour is None:
                    colour = c
                elif c != colour:
                    ok = False
                    break
            if ok and hole:
                contrast = sum(
                    1 for x, y in hole
                    if (rgb[((oy + y) * frame.width + ox + x) * 3],
                        rgb[((oy + y) * frame.width + ox + x) * 3 + 1],
                        rgb[((oy + y) * frame.width + ox + x) * 3 + 2]) != colour)
                ok = contrast >= len(hole) * min_contrast
            if ok:
                hits.append({"x": ox, "y": oy, "colour": list(colour),
                             "width": w, "height": h})
    return hits


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--plan", required=True)
    p.add_argument("--fonts", default="assets/original-derived-png/fonts")
    p.add_argument("--out", default=None)
    p.add_argument("--min-contrast", type=float, default=0.6)
    args = p.parse_args()

    plan = json.loads(pathlib.Path(args.plan).read_text())
    fonts_root = pathlib.Path(args.fonts)
    screens = []
    failures = []
    for spec in plan["screens"]:
        frame = read_png(spec["frame"])
        rgb = frame.rgb()
        font = load_font(fonts_root, spec.get("font", "book-small"))
        items = []
        for text in spec["strings"]:
            hits = locate(frame, rgb, font, text, args.min_contrast)
            if len(hits) != 1:
                failures.append({"screen": spec["name"], "text": text,
                                 "hit_count": len(hits)})
                items.append({"text": text, "hit_count": len(hits), "hits": hits[:4]})
                continue
            h = hits[0]
            items.append({
                "text": text,
                "x": h["x"], "y": h["y"],
                "width": h["width"], "height": h["height"],
                "colour_rgb": h["colour"],
            })
        screens.append({
            "name": spec["name"],
            "frame": spec["frame"],
            "font": spec.get("font", "book-small"),
            "note": spec.get("note"),
            "strings": items,
        })

    result = {
        "tool": "tools/extract_text_layout.py",
        "purpose": "原版文字的座標與顏色，供 remake 照座標繪字與中文化排版沿用",
        "font_source": "assets/original-derived-png/fonts/manifest.json",
        "unique_hit_required": True,
        "failure_count": len(failures),
        "failures": failures,
        "screens": screens,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(json.dumps({"screens": len(screens), "failures": len(failures)},
                     ensure_ascii=False))
    return 0 if not failures else 1


if __name__ == "__main__":
    sys.exit(main())
