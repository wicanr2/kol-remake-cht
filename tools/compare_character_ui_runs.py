#!/usr/bin/env python3
"""比對同一角色建立 UI 情境兩次獨立擷取的檢查點畫面。

用途是把「哪些畫面可逐 byte 重生、哪些本來就會逐次不同」講清楚：屬性頁由 BIOS tick
播種，名字欄有閃爍游標，TITLE 標題序列是循環動畫——這三類不可能逐 byte 相同，
其餘檢查點必須相同，否則就是擷取或流程走偏。

  python3 tools/compare_character_ui_runs.py 第一輪目錄 第二輪目錄 [--out 輸出.json]
"""

from __future__ import annotations

import argparse
import json
import pathlib
import struct
import sys
import zlib


def read_png_rgb(path: pathlib.Path) -> tuple[int, int, bytes]:
    """讀 8-bit truecolor PNG，回傳 (寬, 高, RGB bytes)。"""
    data = path.read_bytes()
    pos, idat, width, height = 8, b"", 0, 0
    while pos < len(data):
        length = struct.unpack(">I", data[pos:pos + 4])[0]
        tag = data[pos + 4:pos + 8]
        body = data[pos + 8:pos + 8 + length]
        if tag == b"IHDR":
            width, height, depth, ctype = struct.unpack(">IIBB", body[:10])
            if depth != 8 or ctype != 2:
                raise SystemExit(f"{path} 不是 8-bit truecolor PNG")
        elif tag == b"IDAT":
            idat += body
        pos += 12 + length
    raw = zlib.decompress(idat)
    stride = width * 3
    prev = bytearray(stride)
    out = bytearray()
    i = 0
    for _ in range(height):
        ftype = raw[i]
        i += 1
        line = bytearray(raw[i:i + stride])
        i += stride
        if ftype == 1:
            for x in range(3, stride):
                line[x] = (line[x] + line[x - 3]) & 0xFF
        elif ftype == 2:
            for x in range(stride):
                line[x] = (line[x] + prev[x]) & 0xFF
        elif ftype == 3:
            for x in range(stride):
                left = line[x - 3] if x >= 3 else 0
                line[x] = (line[x] + ((left + prev[x]) >> 1)) & 0xFF
        elif ftype == 4:
            for x in range(stride):
                a = line[x - 3] if x >= 3 else 0
                b = prev[x]
                c = prev[x - 3] if x >= 3 else 0
                p = a + b - c
                pa, pb, pc = abs(p - a), abs(p - b), abs(p - c)
                pred = a if (pa <= pb and pa <= pc) else (b if pb <= pc else c)
                line[x] = (line[x] + pred) & 0xFF
        elif ftype != 0:
            raise SystemExit(f"{path} 有未支援的 filter {ftype}")
        out += line
        prev = line
    return width, height, bytes(out)


def compare(run_a: pathlib.Path, run_b: pathlib.Path) -> dict:
    ra = json.loads((run_a / "receipt.json").read_text())
    rb = json.loads((run_b / "receipt.json").read_text())
    if ra["scenario"] != rb["scenario"]:
        raise SystemExit(f"情境不同：{ra['scenario']} vs {rb['scenario']}")
    if len(ra["checkpoints"]) != len(rb["checkpoints"]):
        raise SystemExit("兩輪檢查點數量不同，流程已走偏")

    rows = []
    for x, y in zip(ra["checkpoints"], rb["checkpoints"]):
        if x["label"] != y["label"]:
            raise SystemExit(f"檢查點標籤不同：{x['label']} vs {y['label']}")
        row = {"index": x["index"], "label": x["label"],
               "identical": x["png_sha256"] == y["png_sha256"]}
        if not row["identical"]:
            wa, ha, pa = read_png_rgb(run_a / x["file"])
            wb, hb, pb = read_png_rgb(run_b / y["file"])
            if (wa, ha) != (wb, hb):
                row["note"] = f"尺寸不同 {wa}x{ha} vs {wb}x{hb}"
                row["diff_byte_ratio"] = 1.0
            else:
                diff = sum(1 for p, q in zip(pa, pb) if p != q)
                row["diff_byte_ratio"] = round(diff / len(pa), 6)
        rows.append(row)

    identical = sum(1 for r in rows if r["identical"])
    return {
        "tool": "tools/compare_character_ui_runs.py",
        "scenario": ra["scenario"],
        "run_a": {"generated_utc": ra["generated_utc"],
                  "dosbox_conf_sha256": ra["dosbox_conf_sha256"],
                  "character_files_changed": ra["character_files_changed"]},
        "run_b": {"generated_utc": rb["generated_utc"],
                  "dosbox_conf_sha256": rb["dosbox_conf_sha256"],
                  "character_files_changed": rb["character_files_changed"]},
        "checkpoint_count": len(rows),
        "identical_count": identical,
        "differing_count": len(rows) - identical,
        "checkpoints": rows,
    }


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("run_a")
    p.add_argument("run_b")
    p.add_argument("--out", default=None)
    args = p.parse_args()
    result = compare(pathlib.Path(args.run_a), pathlib.Path(args.run_b))
    text = json.dumps(result, ensure_ascii=False, indent=2) + "\n"
    if args.out:
        pathlib.Path(args.out).write_text(text)
    print(text)
    return 0


if __name__ == "__main__":
    sys.exit(main())
