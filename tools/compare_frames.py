#!/usr/bin/env python3
"""同狀態畫面對拍：比較兩張 PNG 的像素差異並輸出可版控的判讀結果。

三種用途共用同一支工具：
1. 抽取資產 對 原版執行畫面（驗證抽取器解出來的就是遊戲畫的）。
2. remake 畫面 對 原版執行畫面（remake parity）。
3. 同一側兩次執行（重生性）。

比較只在 RGB 通道進行；alpha 不參與，因為原版畫面沒有 alpha。差異以像素為單位，
另給出差異的外接矩形，方便判斷「只有文字不同」還是「整個版面錯位」。

  python3 tools/compare_frames.py --a 甲.png --b 乙.png [--crop-a x,y,w,h]
      [--crop-b x,y,w,h] [--diff-png 輸出.png] [--out 結果.json] [--label 名稱]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import Image, read_png, write_png_rgb  # noqa: E402


def parse_box(text: str | None) -> tuple[int, int, int, int] | None:
    if not text:
        return None
    parts = [int(v) for v in text.split(",")]
    if len(parts) != 4:
        raise SystemExit("裁切格式必須是 x,y,w,h")
    return tuple(parts)  # type: ignore[return-value]


def compare(a: Image, b: Image) -> dict:
    if (a.width, a.height) != (b.width, b.height):
        return {"size_match": False,
                "a_size": [a.width, a.height], "b_size": [b.width, b.height],
                "identical": False}
    ra, rb = a.rgb(), b.rgb()
    total = a.width * a.height
    differing = 0
    min_x, min_y, max_x, max_y = a.width, a.height, -1, -1
    for i in range(total):
        o = i * 3
        if ra[o:o + 3] != rb[o:o + 3]:
            differing += 1
            x, y = i % a.width, i // a.width
            min_x = min(min_x, x)
            max_x = max(max_x, x)
            min_y = min(min_y, y)
            max_y = max(max_y, y)
    result = {
        "size_match": True,
        "size": [a.width, a.height],
        "identical": differing == 0,
        "pixel_total": total,
        "pixel_differing": differing,
        "pixel_diff_ratio": round(differing / total, 6),
    }
    if differing:
        result["diff_bounding_box"] = {"x": min_x, "y": min_y,
                                       "width": max_x - min_x + 1,
                                       "height": max_y - min_y + 1}
    return result


def diff_image(a: Image, b: Image) -> bytes:
    """相同像素轉為暗灰，不同像素塗紅，方便一眼看出差異落在哪。"""
    ra, rb = a.rgb(), b.rgb()
    out = bytearray(len(ra))
    for i in range(a.width * a.height):
        o = i * 3
        if ra[o:o + 3] == rb[o:o + 3]:
            grey = (ra[o] + ra[o + 1] + ra[o + 2]) // 6
            out[o] = out[o + 1] = out[o + 2] = grey
        else:
            out[o], out[o + 1], out[o + 2] = 0xFF, 0x00, 0x00
    return bytes(out)


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--a", required=True)
    p.add_argument("--b", required=True)
    p.add_argument("--crop-a", default=None)
    p.add_argument("--crop-b", default=None)
    p.add_argument("--diff-png", default=None)
    p.add_argument("--out", default=None)
    p.add_argument("--label", default=None)
    args = p.parse_args()

    a = read_png(args.a)
    b = read_png(args.b)
    box_a, box_b = parse_box(args.crop_a), parse_box(args.crop_b)
    if box_a:
        a = a.crop(*box_a)
    if box_b:
        b = b.crop(*box_b)

    result = compare(a, b)
    result.update({
        "tool": "tools/compare_frames.py",
        "label": args.label,
        "a": {"path": args.a, "crop": box_a},
        "b": {"path": args.b, "crop": box_b},
    })
    if args.diff_png and result.get("size_match") and not result["identical"]:
        result["diff_png_sha256"] = write_png_rgb(
            args.diff_png, a.width, a.height, diff_image(a, b))

    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(text)
    return 0 if result.get("identical") else 1


if __name__ == "__main__":
    sys.exit(main())
