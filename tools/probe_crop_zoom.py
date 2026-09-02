#!/usr/bin/env python3
"""把畫面的一塊裁下來放大，方便肉眼判讀小元素（游標、標記、圖示）。

  python3 tools/probe_crop_zoom.py --in 畫面.png --box x,y,w,h --scale 6 --out 放大.png
"""

from __future__ import annotations

import argparse
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png, write_png_rgb  # noqa: E402


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--in", dest="src", required=True)
    p.add_argument("--box", required=True, help="x,y,w,h")
    p.add_argument("--scale", type=int, default=6)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    img = read_png(pathlib.Path(args.src))
    x, y, w, h = (int(v) for v in args.box.split(","))
    rgb = img.crop(x, y, w, h).rgb()
    s = args.scale
    out = bytearray()
    for row in range(h):
        line = bytearray()
        for col in range(w):
            px = rgb[(row * w + col) * 3:(row * w + col) * 3 + 3]
            line.extend(px * s)
        out.extend(bytes(line) * s)
    write_png_rgb(pathlib.Path(args.out), w * s, h * s, bytes(out))
    print(args.out)
    return 0


if __name__ == "__main__":
    sys.exit(main())
