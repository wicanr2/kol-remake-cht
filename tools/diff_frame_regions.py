#!/usr/bin/env python3
"""兩張同尺寸畫格差在哪：逐列的區段、每列每欄的計數，以及指定位置的顏色。

`compare_frames.py` 回答「差多少」，這一支回答「差在哪、差成什麼樣」。對拍對不上
的時候，看到的通常不是一團亂——而是「某一列整條」「某四欄貫穿整張」這種有結構的
差異，順著它就找得到版面上少算的那幾個像素。

  tools/run_python_tool.sh tools/diff_frame_regions.py --a remake.png --b 原版.png
  tools/run_python_tool.sh tools/diff_frame_regions.py --a a.png --b b.png \\
      --rows 157,158 --cols 20,100,200,300
"""

from __future__ import annotations

import argparse
import collections
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from png_io import read_png  # noqa: E402

# 每一段摘要最多列幾筆，免得整份輸出被單一維度塞滿。
TOP = 12
# 逐列區段最多印幾列。
MAX_ROWS = 40


def runs(xs: list[int]) -> list[tuple[int, int]]:
    """把一列裡的欄號壓成連續區段。"""
    out, start, prev = [], xs[0], xs[0]
    for x in xs[1:]:
        if x != prev + 1:
            out.append((start, prev))
            start = x
        prev = x
    out.append((start, prev))
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--a", required=True)
    ap.add_argument("--b", required=True)
    ap.add_argument("--rows", default=None, help="另外印這幾列的顏色（逗號分隔）")
    ap.add_argument("--cols", default="20,100,200,300",
                    help="`--rows` 印哪幾欄的顏色")
    args = ap.parse_args()

    a, b = read_png(pathlib.Path(args.a)), read_png(pathlib.Path(args.b))
    if (a.width, a.height) != (b.width, b.height):
        raise SystemExit(f"尺寸不同：{a.width}×{a.height} 與 {b.width}×{b.height}")
    ra, rb, w = a.rgb(), b.rgb(), a.width

    rows: dict[int, list[int]] = collections.defaultdict(list)
    for i in range(0, len(ra), 3):
        if ra[i:i + 3] != rb[i:i + 3]:
            p = i // 3
            rows[p // w].append(p % w)

    total = sum(len(v) for v in rows.values())
    print(f"差 {total} 個像素，分布在 {len(rows)} 列")
    if not rows:
        return 0

    per_row = collections.Counter({y: len(v) for y, v in rows.items()})
    per_col: collections.Counter = collections.Counter()
    for xs in rows.values():
        per_col.update(xs)
    print("差最多的列：", per_row.most_common(TOP))
    print("差最多的欄：", per_col.most_common(TOP))
    band: collections.Counter = collections.Counter()
    for y, n in per_row.items():
        band[y // 25 * 25] += n
    print("每 25 列：", sorted(band.items()))

    print("逐列區段：")
    for y in sorted(rows)[:MAX_ROWS]:
        segs = "  ".join(f"x={s}" if s == e else f"x={s}..{e}"
                         for s, e in runs(rows[y]))
        print(f"  y={y:3d}  {segs}")
    if len(rows) > MAX_ROWS:
        print(f"  …另外 {len(rows) - MAX_ROWS} 列")

    if args.rows:
        cols = [int(x) for x in args.cols.split(",")]
        print("指定位置的顏色：")
        for y in (int(v) for v in args.rows.split(",")):
            for label, raw in (("remake", ra), ("原版  ", rb)):
                px = [tuple(raw[((y * w) + x) * 3:((y * w) + x) * 3 + 3])
                      for x in cols]
                print(f"  y={y:3d} {label} {px}")
    return 0


raise SystemExit(main())
