#!/usr/bin/env python3
"""把城鎮畫面拆成「底圖 ＋ 疊在上面的矩形」。

城鎮畫面不像書頁是一張烘好的素材，而是底圖之上疊了好幾塊。要對拍就得先知道底圖是
哪一張、上面疊了哪些矩形。作法是拿抽出的全螢幕素材逐張與畫面相減，取差異最小的那張
當底圖，再對差異像素做四鄰連通分群；剩餘差異必須為零，否則底圖判斷就是錯的。

  python3 tools/decompose_town_frame.py --frame f.png \\
      --assets assets/original-derived-png/ega [--min-pixels 50] [--out 結果.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402


def best_background(frame, root: pathlib.Path, manifest: dict) -> tuple[str, int, int]:
    fr = frame.rgb()
    best = None
    for a in manifest["archives"]:
        for img in a["images"]:
            if img["width"] != frame.width or img["height"] != frame.height:
                continue
            im = read_png(root / img["file"])
            ir = im.rgb()
            diff = sum(1 for i in range(0, frame.width * frame.height * 3, 3)
                       if fr[i:i + 3] != ir[i:i + 3])
            if best is None or diff < best[2]:
                best = (a["archive"], img["unit_index"], diff, root / img["file"])
    if best is None:
        raise SystemExit("找不到同尺寸的全螢幕素材")
    return best


def components(frame, bg, min_pixels: int) -> list[dict]:
    fr, br = frame.rgb(), bg.rgb()
    W, H = frame.width, frame.height
    mask = bytearray(W * H)
    for y in range(H):
        for x in range(W):
            o = (y * W + x) * 3
            if fr[o:o + 3] != br[o:o + 3]:
                mask[y * W + x] = 1
    seen = bytearray(W * H)
    out = []
    for y in range(H):
        for x in range(W):
            if not mask[y * W + x] or seen[y * W + x]:
                continue
            stack = [(x, y)]
            seen[y * W + x] = 1
            x0 = x1 = x
            y0 = y1 = y
            n = 0
            while stack:
                cx, cy = stack.pop()
                n += 1
                x0, x1 = min(x0, cx), max(x1, cx)
                y0, y1 = min(y0, cy), max(y1, cy)
                for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                    nx, ny = cx + dx, cy + dy
                    if 0 <= nx < W and 0 <= ny < H and mask[ny * W + nx] \
                            and not seen[ny * W + nx]:
                        seen[ny * W + nx] = 1
                        stack.append((nx, ny))
            if n >= min_pixels:
                out.append({"pixels": n, "x": x0, "y": y0,
                            "width": x1 - x0 + 1, "height": y1 - y0 + 1})
    out.sort(key=lambda c: -c["pixels"])
    return out


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--frame", required=True)
    p.add_argument("--assets", default="assets/original-derived-png/ega")
    p.add_argument("--min-pixels", type=int, default=50)
    p.add_argument("--out", default=None)
    args = p.parse_args()

    root = pathlib.Path(args.assets)
    manifest = json.loads((root / "manifest.json").read_text())
    frame = read_png(args.frame)
    archive, index, diff, path = best_background(frame, root, manifest)
    bg = read_png(path)
    comps = components(frame, bg, args.min_pixels)
    covered = 0
    fr, br = frame.rgb(), bg.rgb()
    W, H = frame.width, frame.height
    for y in range(H):
        for x in range(W):
            o = (y * W + x) * 3
            if fr[o:o + 3] != br[o:o + 3] and any(
                    c["x"] <= x < c["x"] + c["width"] and c["y"] <= y < c["y"] + c["height"]
                    for c in comps):
                covered += 1
    receipt = {
        "schema": "kol-town-frame-decomposition-v1",
        "tool": "tools/decompose_town_frame.py",
        "frame": args.frame,
        "background": {"archive": archive, "index": index, "differing_pixels": diff},
        "components": comps,
        "differing_total": diff,
        "differing_inside_components": covered,
        "residual_outside_components": diff - covered,
    }
    if args.out:
        out = pathlib.Path(args.out)
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"background": f"{archive}[{index}]",
                      "components": len(comps),
                      "residual_outside_components": diff - covered}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
