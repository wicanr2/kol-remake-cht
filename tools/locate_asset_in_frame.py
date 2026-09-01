#!/usr/bin/env python3
"""在原版執行畫面中找出某張抽取資產被繪製的位置。

對拍需要知道每個素材在畫面上的座標。與其猜，不如直接搜尋：把資產在指定範圍內逐格
比對，回報差異最少的位置與差異像素數。完全命中（差異 0）代表該資產就是原版當下畫的
那塊像素，位置即為原版的 blit 座標。

資產若帶 alpha，透明像素不參與比對。

  python3 tools/locate_asset_in_frame.py --asset a.png --frame f.png
      [--x0 0 --y0 0 --x1 320 --y1 200] [--out 結果.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--asset", required=True)
    p.add_argument("--frame", required=True)
    p.add_argument("--x0", type=int, default=0)
    p.add_argument("--y0", type=int, default=0)
    p.add_argument("--x1", type=int, default=None)
    p.add_argument("--y1", type=int, default=None)
    p.add_argument("--out", default=None)
    args = p.parse_args()

    asset = read_png(args.asset)
    frame = read_png(args.frame)
    x1 = args.x1 if args.x1 is not None else frame.width - asset.width
    y1 = args.y1 if args.y1 is not None else frame.height - asset.height
    x1 = min(x1, frame.width - asset.width)
    y1 = min(y1, frame.height - asset.height)
    if x1 < args.x0 or y1 < args.y0:
        raise SystemExit("搜尋範圍容不下這張資產")

    arg = asset.rgb()
    alpha = ([asset.pixels[i * 4 + 3] for i in range(asset.width * asset.height)]
             if asset.channels == 4 else [255] * (asset.width * asset.height))
    frg = frame.rgb()
    opaque_indices = [i for i, a in enumerate(alpha) if a]

    best = None
    for oy in range(args.y0, y1 + 1):
        for ox in range(args.x0, x1 + 1):
            diff = 0
            for i in opaque_indices:
                ax, ay = i % asset.width, i // asset.width
                fo = ((oy + ay) * frame.width + ox + ax) * 3
                ao = i * 3
                if frg[fo:fo + 3] != arg[ao:ao + 3]:
                    diff += 1
                    if best is not None and diff >= best["differing"]:
                        break
            if best is None or diff < best["differing"]:
                best = {"x": ox, "y": oy, "differing": diff}
                if diff == 0:
                    break
        if best and best["differing"] == 0:
            break

    result = {
        "tool": "tools/locate_asset_in_frame.py",
        "asset": args.asset,
        "frame": args.frame,
        "asset_size": [asset.width, asset.height],
        "compared_pixels": len(opaque_indices),
        "best_offset": {"x": best["x"], "y": best["y"]},
        "differing_pixels": best["differing"],
        "exact": best["differing"] == 0,
    }
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(text)
    return 0 if result["exact"] else 1


if __name__ == "__main__":
    sys.exit(main())
