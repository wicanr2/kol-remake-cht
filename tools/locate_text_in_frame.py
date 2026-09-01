#!/usr/bin/env python3
"""用抽出的原版字型，在原版畫面中機械定位某個字串的座標與顏色。

中文化要把譯文放進原版的同一個文字框，因此必須知道原版每個字串畫在哪、用什麼顏色。
與其目視量測，不如拿字型自己去比對：把字串依字模展開成墨點集合，逐位置檢查畫面上
所有墨點是否同一個顏色，且該顏色不等於字模外緣的顏色。

  python3 tools/locate_text_in_frame.py --frame f.png --font book-small --text "STRENGTH"
      [--fonts assets/original-derived-png/fonts] [--out 結果.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402


def load_font(root: pathlib.Path, name: str) -> dict:
    manifest = json.loads((root / "manifest.json").read_text())
    entry = next((f for f in manifest["fonts"] if f["name"] == name), None)
    if entry is None:
        raise SystemExit(f"字型 manifest 沒有 {name}")
    sheet = read_png(root / entry["file"])
    return {"entry": entry, "sheet": sheet}


def glyph_ink(font: dict, codepoint: int) -> list[tuple[int, int]]:
    e = font["entry"]
    index = codepoint - e["first_codepoint"]
    if not 0 <= index < e["glyph_count"]:
        raise SystemExit(f"字型沒有 codepoint {codepoint}")
    cw, ch = e["cell"]["width"], e["cell"]["height"]
    gx = (index % e["sheet"]["columns"]) * cw
    gy = (index // e["sheet"]["columns"]) * ch
    sheet = font["sheet"]
    ink = []
    for y in range(ch):
        for x in range(cw):
            o = ((gy + y) * sheet.width + gx + x) * 4
            if sheet.pixels[o + 3]:
                ink.append((x, y))
    return ink


def text_ink(font: dict, text: str) -> tuple[list[tuple[int, int]], list[tuple[int, int]], int, int]:
    """回傳 (墨點, 非墨點, 文字框寬, 文字框高)。

    只比對墨點會在單色區塊上大量誤命中（整片紅木桌背景的每個位置都「同色」），
    因此同時回傳非墨點，讓呼叫端要求文字框內確實有對比。
    """
    e = font["entry"]
    advance = e["advance"]
    cw, ch = e["cell"]["width"], e["cell"]["height"]
    ink: list[tuple[int, int]] = []
    hole: list[tuple[int, int]] = []
    for i, c in enumerate(text):
        on = set(glyph_ink(font, ord(c)))
        for y in range(ch):
            for x in range(cw):
                (ink if (x, y) in on else hole).append((i * advance + x, y))
    return ink, hole, advance * len(text), e["cell"]["height"]


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--frame", required=True)
    p.add_argument("--fonts", default="assets/original-derived-png/fonts")
    p.add_argument("--font", default="book-small")
    p.add_argument("--text", required=True)
    p.add_argument("--out", default=None)
    p.add_argument("--max-hits", type=int, default=20)
    p.add_argument("--min-contrast", type=float, default=0.6,
                   help="文字框內非墨點必須有多少比例不是墨色")
    args = p.parse_args()

    font = load_font(pathlib.Path(args.fonts), args.font)
    frame = read_png(args.frame)
    rgb = frame.rgb()
    ink, hole, w, h = text_ink(font, args.text)
    if not ink:
        raise SystemExit("這串文字沒有任何墨點，無法定位")

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
                # 文字框內的非墨點必須大多不是墨色，否則整片單色區塊都會命中。
                contrast = 0
                for x, y in hole:
                    o = ((oy + y) * frame.width + ox + x) * 3
                    if (rgb[o], rgb[o + 1], rgb[o + 2]) != colour:
                        contrast += 1
                ok = contrast >= len(hole) * args.min_contrast
            if ok:
                hits.append({"x": ox, "y": oy,
                             "colour": {"r": colour[0], "g": colour[1], "b": colour[2]}})
                if len(hits) >= args.max_hits:
                    break
        if len(hits) >= args.max_hits:
            break

    result = {
        "tool": "tools/locate_text_in_frame.py",
        "frame": args.frame,
        "font": args.font,
        "text": args.text,
        "text_box": {"width": w, "height": h},
        "ink_pixels": len(ink),
        "hole_pixels": len(hole),
        "min_contrast": args.min_contrast,
        "hits": hits,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(text)
    return 0 if hits else 1


if __name__ == "__main__":
    sys.exit(main())
