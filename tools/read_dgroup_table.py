#!/usr/bin/env python3
"""從 unpacked 的 KOLTOWN.EXE 讀 DGROUP 的初始化資料。

`.i64` 沒有把 DGROUP 全部倒成有內容的 `db`，但初始化資料本來就在執行檔裡——
只要知道 DGROUP 在檔案裡的起點就讀得到。起點用一張**已知內容**的表定位：任務旗標
的除數表（DGROUP `0x7C` = `01 04 10 40`，RE-105）在檔內只出現一次。

定位之後先跑正對照：隨機遭遇那六組敵人站位（`0x0BD4` 欄／`0x0B8C` 列，各 72 byte）
必須與 `internal/battle/ambush.go` 裡那份表逐 byte 相同。對不上就不要相信後面讀到
的任何一個 byte——位置錯了一樣讀得出「看起來像資料」的東西。

  tools/run_python_tool.sh tools/read_dgroup_table.py \\
      --exe workplace/analysis/koltown/KOLTOWN-unpacked.EXE 0x12B2:6 0x12AC:6
"""

from __future__ import annotations

import argparse
import pathlib
import sys

# 定位用的錨：DGROUP 位移 → 內容。
ANCHOR_OFFSET = 0x7C
ANCHOR_BYTES = bytes([0x01, 0x04, 0x10, 0x40])

# 正對照：隨機遭遇的六組敵人站位（RE-098）。
AMBUSH_ENEMY_X_OFF, AMBUSH_ENEMY_Y_OFF = 0x0BD4, 0x0B8C
AMBUSH_ENEMY_X = [
    [28, 28, 28, 28, 27, 27, 27, 26, 26, 26, 26, 26],
    [31, 33, 29, 35, 32, 30, 34, 28, 36, 31, 33, 32],
    [36, 36, 36, 36, 37, 37, 37, 38, 38, 38, 38, 38],
    [31, 33, 29, 35, 32, 30, 34, 28, 36, 31, 33, 32],
    [28, 36, 27, 37, 27, 37, 28, 36, 28, 36, 27, 37],
    [32, 32, 30, 30, 34, 34, 29, 29, 35, 35, 32, 32],
]
AMBUSH_ENEMY_Y = [
    [32, 30, 34, 28, 31, 29, 33, 30, 32, 28, 34, 31],
    [37, 37, 37, 37, 38, 38, 38, 36, 36, 39, 39, 39],
    [32, 30, 34, 28, 31, 29, 33, 30, 32, 28, 34, 31],
    [27, 27, 27, 27, 26, 26, 26, 28, 28, 25, 25, 25],
    [32, 32, 30, 30, 34, 34, 29, 29, 35, 35, 32, 32],
    [27, 37, 26, 38, 26, 38, 27, 37, 27, 37, 26, 38],
]


def locate(blob: bytes) -> int:
    """回傳 DGROUP 在檔案裡的起點。錨不是唯一的就失敗，不猜。"""
    hits = [i for i in range(len(blob) - len(ANCHOR_BYTES) + 1)
            if blob[i:i + len(ANCHOR_BYTES)] == ANCHOR_BYTES]
    if len(hits) != 1:
        raise SystemExit(f"錨 {ANCHOR_BYTES.hex(' ')} 出現 {len(hits)} 次，無法定位")
    return hits[0] - ANCHOR_OFFSET


def check(blob: bytes, base: int) -> None:
    for name, off, want in (("敵人站位欄", AMBUSH_ENEMY_X_OFF, AMBUSH_ENEMY_X),
                            ("敵人站位列", AMBUSH_ENEMY_Y_OFF, AMBUSH_ENEMY_Y)):
        flat = [v for row in want for v in row]
        got = list(blob[base + off: base + off + len(flat)])
        if got != flat:
            raise SystemExit(f"正對照失敗：{name}（DGROUP 0x{off:04X}）"
                             f"讀到 {got[:12]}…，期望 {flat[:12]}…")
        print(f"正對照 {name}（DGROUP 0x{off:04X}，{len(flat)} byte）：相同")


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--exe", required=True, help="unpacked 的 KOLTOWN.EXE")
    p.add_argument("ranges", nargs="+", metavar="OFF:LEN",
                   help="要讀的 DGROUP 位移與長度，例如 0x12B2:6")
    args = p.parse_args()

    blob = pathlib.Path(args.exe).read_bytes()
    base = locate(blob)
    print(f"{args.exe}：{len(blob)} byte，DGROUP 檔內起點 0x{base:05X}")
    check(blob, base)
    print()
    for spec in args.ranges:
        text, _, count = spec.partition(":")
        off, n = int(text, 0), int(count or "16", 0)
        got = blob[base + off: base + off + n]
        print("DGROUP 0x%04X  %-*s  十進位 %s" % (
            off, n * 3, " ".join("%02X" % b for b in got),
            " ".join(str(b) for b in got)))
    return 0


sys.exit(main())
