#!/usr/bin/env python3
"""用原版字型把畫面上的一列文字讀回字串。

中文化需要原版每一句 UI 文字的原文；靠人眼抄容易出錯，靠猜更糟。這支工具對指定
起點的每個字模格逐一比對字型裡的 93 個字模：某個字模若能讓所有墨點同色、且所有
非墨點都不是該顏色，就是命中。全空的格子讀成空白。

  python3 tools/read_text_from_frame.py --frame f.png --x 176 --y 52 --cells 12
      [--rows 12 --row-step 6] [--font book-small] [--out 結果.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from locate_text_in_frame import glyph_ink, load_font  # noqa: E402
from png_io import read_png  # noqa: E402


def read_cell(frame, rgb, font, x: int, y: int) -> tuple[str, tuple | None]:
    e = font["entry"]
    cw, ch = e["cell"]["width"], e["cell"]["height"]
    if x + cw > frame.width or y + ch > frame.height:
        return "", None
    best = None
    for index in range(e["glyph_count"]):
        code = e["first_codepoint"] + index
        ink = glyph_ink(font, code)
        if not ink:
            continue
        colour = None
        ok = True
        for gx, gy in ink:
            o = ((y + gy) * frame.width + x + gx) * 3
            c = (rgb[o], rgb[o + 1], rgb[o + 2])
            if colour is None:
                colour = c
            elif c != colour:
                ok = False
                break
        if not ok:
            continue
        inkset = set(ink)
        for gy in range(ch):
            for gx in range(cw):
                if (gx, gy) in inkset:
                    continue
                o = ((y + gy) * frame.width + x + gx) * 3
                if (rgb[o], rgb[o + 1], rgb[o + 2]) == colour:
                    ok = False
                    break
            if not ok:
                break
        if ok and (best is None or len(ink) > best[1]):
            best = (chr(code), len(ink), colour)
    if best is None:
        return " ", None
    return best[0], best[2]


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--frame", required=True)
    p.add_argument("--fonts", default="assets/original-derived-png/fonts")
    p.add_argument("--font", default="book-small")
    p.add_argument("--x", type=int, required=True)
    p.add_argument("--y", type=int, required=True)
    p.add_argument("--cells", type=int, required=True)
    p.add_argument("--rows", type=int, default=1)
    p.add_argument("--row-step", type=int, default=6)
    p.add_argument("--out", default=None)
    args = p.parse_args()

    font = load_font(pathlib.Path(args.fonts), args.font)
    frame = read_png(args.frame)
    rgb = frame.rgb()
    advance = font["entry"]["advance"]

    lines = []
    for r in range(args.rows):
        y = args.y + r * args.row_step
        chars = []
        colours = []
        for c in range(args.cells):
            ch, colour = read_cell(frame, rgb, font, args.x + c * advance, y)
            chars.append(ch)
            if colour:
                colours.append(colour)
        text = "".join(chars).rstrip()
        counts: dict[tuple, int] = {}
        for c in colours:
            counts[c] = counts.get(c, 0) + 1
        dominant = max(counts, key=counts.get) if counts else None
        lines.append({"row": r, "y": y, "text": text,
                      "colour_rgb": list(dominant) if dominant else None})

    result = {
        "tool": "tools/read_text_from_frame.py",
        "frame": args.frame,
        "font": args.font,
        "origin": {"x": args.x, "y": args.y},
        "cells": args.cells,
        "row_step": args.row_step,
        "lines": lines,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    for line in lines:
        print(f"{line['y']:4d} {line['colour_rgb']} {line['text']!r}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
