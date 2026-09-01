#!/usr/bin/env python3
"""標題序列的畫面對拍：原版每一個擷取畫格都要能在 remake 的一輪播放裡找到。

標題序列有兩處亂數、原版取樣又是 0.3 秒一張的去重畫格，因此逐幀位置對位沒有意義。
可以要求的是**集合包含**：原版擷取到的每一張畫面，remake 都要能逐像素產生一模一樣的
一張。找不到完全相同的，再檢查是不是擷取在整張搬移的中途（上半是一張、下半是另一張），
那是擷取造成的撕裂，不是 remake 的差異；仍找不到就記為未解釋，不含糊帶過。

  python3 tools/verify_title_parity.py --remake workplace/shots/title \\
      --original docs/verification/title-sequence/observe-run1/frames \\
      --out docs/verification/title-sequence/parity-run1.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from png_io import read_png  # noqa: E402

WIDTH, HEIGHT = 320, 200


def load(path: pathlib.Path) -> bytes:
    image = read_png(path)
    if (image.width, image.height) != (WIDTH, HEIGHT):
        raise SystemExit(f"{path} 不是 {WIDTH}x{HEIGHT}")
    return bytes(image.rgb())


def row(data: bytes, y: int) -> bytes:
    return data[y * WIDTH * 3:(y + 1) * WIDTH * 3]


def matching_suffix_row(frame: bytes, other: bytes) -> int:
    """回傳最小的列號 r，使得第 r 列（含）之後的每一個 byte 都相同；沒有就回 HEIGHT。"""
    stride = WIDTH * 3
    if frame == other:
        return 0
    row_index = HEIGHT
    while row_index > 0 and frame[(row_index - 1) * stride:] == other[(row_index - 1) * stride:]:
        row_index -= 1
    return row_index


def matching_prefix_row(frame: bytes, other: bytes) -> int:
    """回傳最大的列號 r，使得第 0 列到第 r-1 列的每一個 byte 都相同。"""
    stride = WIDTH * 3
    row_index = 0
    while row_index < HEIGHT and frame[row_index * stride:(row_index + 1) * stride] \
            == other[row_index * stride:(row_index + 1) * stride]:
        row_index += 1
    return row_index


def find_partial_update(frame: bytes, remake: dict[str, bytes]) -> dict | None:
    """畫面是否為「下半已更新完成、上半還在更新中途」的擷取。

    整張搬移不是瞬間完成的，取樣落在搬移途中就會得到這種畫面。這裡只斷言可以量到的
    事實：存在一個分割列，其後的每一列都逐像素等於 remake 的某一張；分割列之前的
    部分另外記下它與最接近的 remake 畫面差幾個像素，不對成因下結論。
    """
    best = None
    for name, data in remake.items():
        row_index = matching_suffix_row(frame, data)
        if row_index >= HEIGHT:
            continue
        if best is None or row_index < best[0]:
            best = (row_index, name)
    if best is None:
        best = (HEIGHT, None)
    if best[0] == 0:
        return None
    split, name = best
    prefix = max(((matching_prefix_row(frame, data), other_name)
                  for other_name, data in remake.items()))
    if prefix[0] > HEIGHT - split:
        rest = min((sum(1 for i in range(prefix[0] * WIDTH * 3, len(frame), 3)
                        if frame[i:i + 3] != other[i:i + 3]), other_name)
                   for other_name, other in remake.items())
        return {"complete_part": "prefix", "prefix_frame": prefix[1],
                "split_row": prefix[0],
                "rest_closest_remake": rest[1], "rest_pixel_difference": rest[0]}
    head = min((sum(1 for i in range(0, split * WIDTH * 3, 3)
                    if frame[i:i + 3] != other[i:i + 3]), other_name)
               for other_name, other in remake.items())
    return {"complete_part": "suffix", "suffix_frame": name, "split_row": split,
            "rest_closest_remake": head[1], "rest_pixel_difference": head[0]}


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--remake", required=True, action="append",
                   help="remake 畫面目錄，可給多次；判準是所有目錄的聯集")
    p.add_argument("--original", required=True)
    p.add_argument("--out", required=True)
    p.add_argument("--skip-first", type=int, default=0,
                   help="略過原版擷取開頭的 N 張（TITLE.EXE 還沒開始畫的畫面）")
    args = p.parse_args()

    remake = {}
    for directory in args.remake:
        root = pathlib.Path(directory)
        for q in sorted(root.glob("title-*.png")):
            remake[f"{root.name}/{q.name}"] = load(q)
    if not remake:
        raise SystemExit("remake 畫面目錄是空的")
    by_hash: dict[str, str] = {}
    for name, data in remake.items():
        by_hash.setdefault(hashlib.sha256(data).hexdigest(), name)

    results, matched, torn, unmatched = [], 0, 0, 0
    originals = sorted(pathlib.Path(args.original).glob("frame-*.png"))
    for index, q in enumerate(originals):
        if index < args.skip_first:
            results.append({"original": q.name, "status": "skipped",
                            "reason": "TITLE.EXE 開始繪製之前的畫面"})
            continue
        data = load(q)
        name = by_hash.get(hashlib.sha256(data).hexdigest())
        if name:
            matched += 1
            results.append({"original": q.name, "status": "exact", "remake": name})
            continue
        partial = find_partial_update(data, remake)
        if partial:
            torn += 1
            results.append({"original": q.name, "status": "partial_update_capture",
                            **partial})
            continue
        best = min(((sum(1 for i in range(0, len(data), 3)
                         if data[i:i + 3] != other[i:i + 3]), other_name)
                    for other_name, other in remake.items()))
        unmatched += 1
        results.append({"original": q.name, "status": "unmatched",
                        "closest_remake": best[1], "pixel_difference": best[0]})

    receipt = {
        "schema": "kol-title-parity-v1",
        "tool": "tools/verify_title_parity.py",
        "remake_frames": len(remake),
        "remake_dirs": args.remake,
        "original_frames": len(originals),
        "exact": matched, "partial_update_capture": torn, "unmatched": unmatched,
        "skipped": args.skip_first,
        "criterion": "原版每一張擷取畫格都必須逐像素等於 remake 一輪播放裡的某一張；"
                     "取樣落在整張搬移途中的畫格另記為更新中途，並量出分割列。",
        "results": results,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({k: receipt[k] for k in
                      ("remake_frames", "original_frames", "exact",
                       "partial_update_capture", "unmatched", "skipped")}, ensure_ascii=False))
    return 1 if unmatched else 0


if __name__ == "__main__":
    sys.exit(main())
