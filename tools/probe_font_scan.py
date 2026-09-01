#!/usr/bin/env python3
"""把 executable 當成 1bpp 點陣圖攤開，用來目視搜尋內嵌字型表。

點陣字型在檔案裡是連續的 bit 圖樣，攤成固定寬度的 1bpp 影像時會出現肉眼可辨的字形
列。這支只是搜尋用的探針：它不宣稱找到的是什麼，只把候選位置顯示出來，實際結論仍
要由 IDA 的繪字 consumer 證實。

  python3 tools/probe_font_scan.py --zip kol.zip --entry KOLBOOK.EXE \\
      --out workplace/fontscan --row-bytes 8 --rows 256
"""

from __future__ import annotations

import argparse
import pathlib
import struct
import sys
import zipfile
import zlib


def write_png(path: pathlib.Path, w: int, h: int, rgb: bytes) -> None:
    raw = b"".join(b"\x00" + rgb[y * w * 3:(y + 1) * w * 3] for y in range(h))

    def chunk(tag: bytes, body: bytes) -> bytes:
        payload = tag + body
        return (struct.pack(">I", len(body)) + payload
                + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))

    path.write_bytes(b"\x89PNG\r\n\x1a\n"
                     + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0))
                     + chunk(b"IDAT", zlib.compress(raw, 9))
                     + chunk(b"IEND", b""))


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--entry", required=True, help="ZIP 內的檔名尾段，例如 KOLBOOK.EXE")
    p.add_argument("--out", required=True)
    p.add_argument("--row-bytes", type=int, default=8)
    p.add_argument("--rows", type=int, default=256)
    p.add_argument("--gap", type=int, default=4, help="欄之間的分隔像素")
    args = p.parse_args()

    outdir = pathlib.Path(args.out)
    if "workplace" not in outdir.resolve().parts:
        raise SystemExit("探針輸出只可寫入 ignored workplace")
    outdir.mkdir(parents=True, exist_ok=True)

    with zipfile.ZipFile(args.zip) as zf:
        name = next((i.filename for i in zf.infolist()
                     if i.filename.split("/")[-1] == args.entry), None)
        if name is None:
            raise SystemExit(f"ZIP 沒有 {args.entry}")
        data = zf.read(name)

    row_w = args.row_bytes * 8
    per_col = args.rows * args.row_bytes
    columns = (len(data) + per_col - 1) // per_col
    width = columns * (row_w + args.gap) - args.gap
    height = args.rows
    rgb = bytearray(width * height * 3)
    for i in range(width * height):
        rgb[i * 3] = rgb[i * 3 + 1] = rgb[i * 3 + 2] = 0x30      # 分隔底色

    for col in range(columns):
        base = col * per_col
        x0 = col * (row_w + args.gap)
        for y in range(height):
            off = base + y * args.row_bytes
            for bx in range(args.row_bytes):
                if off + bx >= len(data):
                    break
                value = data[off + bx]
                for bit in range(8):
                    on = (value >> (7 - bit)) & 1
                    x = x0 + bx * 8 + bit
                    o = (y * width + x) * 3
                    v = 0xFF if on else 0x00
                    rgb[o] = rgb[o + 1] = rgb[o + 2] = v

    out = outdir / f"{args.entry.lower()}-{args.row_bytes}x{args.rows}.png"
    write_png(out, width, height, bytes(rgb))
    print(f"{out} {width}x{height} columns={columns} "
          f"bytes_per_column={per_col} file_bytes={len(data)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
