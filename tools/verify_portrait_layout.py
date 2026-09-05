#!/usr/bin/env python3
"""驗證立繪版面表：每一格的尺寸要與抽出來的 PNG 逐張相符（RE-126）。

`extract_portrait_layout.py` 讀的是 KOLTOWN 資料段裡一串沒有標頭、沒有筆數、
只靠位址接龍的 byte。「解得開」不等於「解對了」，所以這一支拿另一個來源對答案：
版面表說第 n 張圖是 `w` 個 byte 欄寬、`h` 條掃描列，抽出來的 PNG 就必須是
`w × 8` 像素寬、`h` 像素高。222 張圖全部對得上，版面才算讀對。

  tools/run_python_tool.sh tools/verify_portrait_layout.py \
      --layout assets/original-derived-data/portrait-layout.json \
      --png assets/original-derived-png/ega
"""

from __future__ import annotations

import argparse
import json
import pathlib
import struct
import sys

# EGA 平面模式一個 byte 橫跨八個像素，所以表裡的寬是 byte 欄、高才是像素。
PIXELS_PER_BYTE = 8


def png_size(path: pathlib.Path) -> tuple[int, int]:
    raw = path.read_bytes()
    if raw[:8] != b"\x89PNG\r\n\x1a\n" or raw[12:16] != b"IHDR":
        raise SystemExit(f"{path} 不是 PNG")
    w, h = struct.unpack(">II", raw[16:24])
    return w, h


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--layout", required=True)
    ap.add_argument("--png", required=True)
    args = ap.parse_args()

    layout = json.loads(pathlib.Path(args.layout).read_text(encoding="utf-8"))
    root = pathlib.Path(args.png)
    bad: list[str] = []
    checked = 0

    def want(kind: str, path: pathlib.Path, w: int, h: int) -> None:
        nonlocal checked
        checked += 1
        got = png_size(path)
        if got != (w * PIXELS_PER_BYTE, h):
            bad.append(f"{kind} {path.name}：PNG {got[0]}×{got[1]}，"
                       f"表說 {w * PIXELS_PER_BYTE}×{h}")

    # 體型那一層的尺寸是「兩個角落相減」，不是獨立的寬高欄位。
    for v, b in enumerate(layout["body"]):
        want("體型", root / f"PERPICS/perpics-{v:03d}.png",
             b["x2"] - b["x"], b["y2"] - b["y"])
    for v, layers in enumerate(layout["layers"]):
        for k, name in enumerate(("HEADPICS/headpics", "TORPICS/torpics",
                                  "LEGSPICS/legspics")):
            want("外觀", root / f"{name}-{v:03d}.png", layers[k]["w"], layers[k]["h"])
    # 武器、盾與手的尺寸表與體型無關，位置表才分體型；拿第一組的尺寸驗就夠。
    for i, c in enumerate(layout["weapons"][0]):
        want("武器", root / f"WEAPPICS/weappics-{i:03d}.png", c["w"], c["h"])
    for i, c in enumerate(layout["shields"][0]):
        want("盾", root / f"SHLDPICS/shldpics-{i:03d}.png", c["w"], c["h"])
    for i, c in enumerate(layout["hands"]):
        want("手", root / f"HANDPICS/handpics-{i:03d}.png", c["w"], c["h"])

    # 尺寸表不分體型這件事本身也要驗，不能只是註解裡的說法。
    for key in ("weapons", "shields"):
        first = [(c["w"], c["h"]) for c in layout[key][0]]
        for v, rows in enumerate(layout[key][1:], start=1):
            if [(c["w"], c["h"]) for c in rows] != first:
                bad.append(f"{key} 第 {v} 個體型的尺寸與第 0 個不同")

    for line in bad:
        print(line, file=sys.stderr)
    print(f"對過 {checked} 張圖，{len(bad)} 處不符")
    return 1 if bad else 0


raise SystemExit(main())
