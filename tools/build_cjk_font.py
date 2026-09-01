#!/usr/bin/env python3
"""把中文 UI 字串會用到的字元，從系統 TrueType 字型烘成點陣字模表。

remake 的中文模式不在執行期做字型排版：字模先烘成 PNG 字模表，執行期只讀 PNG，
與原版抽出的英文字型走同一條路。烘焙用 ffmpeg 的 drawtext（libfreetype），
再以固定門檻二值化，因此同一組輸入每次都得到同一張表。

  python3 tools/build_cjk_font.py --strings docs/i18n/newchar-zh-Hant.json \\
      --out assets/ui-font --name zh-hant-16 --family "Noto Sans CJK TC" --size 16

字型本體不入庫；入庫的是烘出來的字模表與 manifest，manifest 記下字型家族、授權與
烘焙參數，好讓同一張表可以重生。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import struct
import subprocess
import sys
import zlib

CELL_PITCH = 32   # 烘焙時的格距
DRAW_MARGIN = 8   # 每格內縮的繪製起點；鄰格字模的側邊突出才不會被量進來
ASCII_MAX = 0x7F


def _walk(node, seen: set[str]) -> None:
    """把任意巢狀 JSON 裡的字串值都掃過一遍，收集非 ASCII 字元。"""
    if isinstance(node, str):
        seen.update(ch for ch in node if ord(ch) > 0x7F)
    elif isinstance(node, list):
        for item in node:
            _walk(item, seen)
    elif isinstance(node, dict):
        for item in node.values():
            _walk(item, seen)


def collect(strings_path: pathlib.Path,
            extra: list[pathlib.Path] | None = None) -> list[str]:
    data = json.loads(strings_path.read_text())
    seen: set[str] = set()
    for key in ("labels", "races", "secondary", "professions", "ui"):
        for item in data.get(key, []):
            seen.update(item["zh"])
    # 譯文不只住在字串表：城鎮畫面的對拍樣本自帶預先斷行的中文台詞，
    # 那些字沒烘進來就會在畫面上變成空白。
    for path in extra or []:
        _walk(json.loads(path.read_text()), seen)
    # 玩家輸入的名字與屬性數值都是 ASCII，因此整段可見 ASCII 一律烘進來，
    # 中文模式才不需要在兩套字型之間切換。
    seen.update(chr(c) for c in range(0x21, 0x7F))
    seen.discard(" ")
    return sorted(seen)


def render(chars: list[str], family: str, size: int, columns: int,
           work: pathlib.Path) -> bytes:
    """把每個字元烘到固定格點上。

    字元經由 `textfile=` 傳給 drawtext 並關掉展開，否則 `%`、`\\`、引號這些字元會被
    drawtext 的參數解析或 text expansion 吃掉——那會安靜地烘出錯字模。
    """
    rows = (len(chars) + columns - 1) // columns
    width, height = columns * CELL_PITCH, rows * CELL_PITCH
    text_dir = work / "chars"
    text_dir.mkdir(parents=True, exist_ok=True)
    filters = []
    for index, ch in enumerate(chars):
        x = (index % columns) * CELL_PITCH + DRAW_MARGIN
        y = (index // columns) * CELL_PITCH + DRAW_MARGIN
        path = text_dir / f"{ord(ch):04x}.txt"
        path.write_text(ch, encoding="utf-8")
        filters.append(
            f"drawtext=font={family}:textfile={path}:expansion=none"
            f":fontsize={size}:fontcolor=white:x={x}:y={y}")
    graph = ",".join(filters)
    out = work / "sheet-raw.png"
    result = subprocess.run(
        ["ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
         "-f", "lavfi", "-i", f"color=c=black:s={width}x{height}:d=1",
         "-vf", graph, "-frames:v", "1", str(out)],
        capture_output=True, text=True)
    if result.returncode != 0:
        raise SystemExit("ffmpeg 烘焙失敗：" + result.stderr.strip()[:400])
    return out.read_bytes()


def decode_png_gray(raw: bytes) -> tuple[int, int, list[int]]:
    """只解 ffmpeg 產出的那一種 PNG：8-bit RGB／RGBA、無 interlace。"""
    if raw[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit("不是 PNG")
    pos, idat, width, height, colour, depth = 8, bytearray(), 0, 0, 0, 0
    while pos < len(raw):
        length = struct.unpack(">I", raw[pos:pos + 4])[0]
        kind = raw[pos + 4:pos + 8]
        body = raw[pos + 8:pos + 8 + length]
        if kind == b"IHDR":
            width, height, depth, colour = struct.unpack(">IIBB", body[:10])
        elif kind == b"IDAT":
            idat += body
        elif kind == b"IEND":
            break
        pos += 12 + length
    if depth != 8 or colour not in (2, 6):
        raise SystemExit(f"未預期的 PNG 格式 depth={depth} colour={colour}")
    channels = 3 if colour == 2 else 4
    data = zlib.decompress(bytes(idat))
    stride = width * channels
    grey: list[int] = []
    previous = bytearray(stride)
    pos = 0
    for _ in range(height):
        filter_type = data[pos]
        line = bytearray(data[pos + 1:pos + 1 + stride])
        pos += 1 + stride
        for i in range(stride):
            a = line[i - channels] if i >= channels else 0
            b = previous[i]
            c = previous[i - channels] if i >= channels else 0
            if filter_type == 1:
                line[i] = (line[i] + a) & 0xFF
            elif filter_type == 2:
                line[i] = (line[i] + b) & 0xFF
            elif filter_type == 3:
                line[i] = (line[i] + (a + b) // 2) & 0xFF
            elif filter_type == 4:
                p = a + b - c
                pa, pb, pc = abs(p - a), abs(p - b), abs(p - c)
                pred = a if (pa <= pb and pa <= pc) else (b if pb <= pc else c)
                line[i] = (line[i] + pred) & 0xFF
        grey.extend(line[i] for i in range(0, stride, channels))
        previous = line
    return width, height, grey


def write_png_rgba(path: pathlib.Path, width: int, height: int, pixels: bytes) -> None:
    raw = bytearray()
    for y in range(height):
        raw.append(0)
        raw += pixels[y * width * 4:(y + 1) * width * 4]
    def chunk(kind: bytes, body: bytes) -> bytes:
        return (struct.pack(">I", len(body)) + kind + body
                + struct.pack(">I", zlib.crc32(kind + body) & 0xFFFFFFFF))
    path.write_bytes(b"\x89PNG\r\n\x1a\n"
                     + chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0))
                     + chunk(b"IDAT", zlib.compress(bytes(raw), 9))
                     + chunk(b"IEND", b""))


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--strings", required=True)
    p.add_argument("--extra-json", action="append", default=[],
                   help="另外掃描這些 JSON 裡所有字串值的非 ASCII 字元（可重複）")
    p.add_argument("--out", required=True)
    p.add_argument("--name", default="zh-hant-16")
    p.add_argument("--family", default="Noto Sans CJK TC")
    p.add_argument("--license", default="SIL Open Font License 1.1")
    p.add_argument("--size", type=int, default=16)
    p.add_argument("--cell", type=int, default=16)
    p.add_argument("--cell-height", type=int, default=0,
                   help="字模格高度，預設與寬度相同；ASCII 有下伸部時需要多留幾列")
    p.add_argument("--columns", type=int, default=16)
    p.add_argument("--threshold", type=int, default=96)
    p.add_argument("--work", default="/tmp/cjk-font")
    args = p.parse_args()

    chars = collect(pathlib.Path(args.strings),
                    [pathlib.Path(x) for x in args.extra_json])
    work = pathlib.Path(args.work)
    work.mkdir(parents=True, exist_ok=True)
    raw = render(chars, args.family, args.size, args.columns, work)
    width, height, grey = decode_png_gray(raw)

    # 先量出所有字模的共同邊界，再用同一個位移把每個字模裁進固定字模格，
    # 這樣基線一致，不會逐字浮動。
    cell_h = args.cell_height or args.cell
    top, bottom, left, right = height, 0, width, 0
    for index in range(len(chars)):
        ox = (index % args.columns) * CELL_PITCH + DRAW_MARGIN
        oy = (index // args.columns) * CELL_PITCH + DRAW_MARGIN
        for y in range(-DRAW_MARGIN, CELL_PITCH - DRAW_MARGIN):
            for x in range(-DRAW_MARGIN, CELL_PITCH - DRAW_MARGIN):
                if grey[(oy + y) * width + ox + x] >= args.threshold:
                    top, bottom = min(top, y), max(bottom, y)
                    left, right = min(left, x), max(right, x)
    if bottom - top + 1 > cell_h or right - left + 1 > args.cell:
        raise SystemExit(f"字模超出 {args.cell}x{cell_h}："
                         f"高 {bottom - top + 1}、寬 {right - left + 1}")
    yoff, xoff = top, left

    cell = args.cell
    columns = args.columns
    rows = (len(chars) + columns - 1) // columns
    sheet_w, sheet_h = columns * cell, rows * cell_h
    pixels = bytearray(sheet_w * sheet_h * 4)
    glyphs = []
    for index, ch in enumerate(chars):
        ox = (index % columns) * CELL_PITCH + DRAW_MARGIN + xoff
        oy = (index // columns) * CELL_PITCH + DRAW_MARGIN + yoff
        cx = (index % columns) * cell
        cy = (index // columns) * cell_h
        bits, ink_right = [], -1
        for y in range(cell_h):
            value = 0
            for x in range(cell):
                on = grey[(oy + y) * width + ox + x] >= args.threshold
                value = (value << 1) | (1 if on else 0)
                if on:
                    ink_right = max(ink_right, x)
                    o = ((cy + y) * sheet_w + cx + x) * 4
                    pixels[o:o + 4] = b"\xff\xff\xff\xff"
            bits.append(value)
        # 漢字是全形，固定一格；ASCII 是比例字，字寬取墨跡右緣加一個間隙，
        # 否則以半形固定字距畫會把相鄰字模切掉。
        if ord(ch) > ASCII_MAX:
            advance = cell
        else:
            advance = min(cell, max(4, ink_right + 2))
        glyphs.append({
            "codepoint": ord(ch), "character": ch,
            "sheet_column": index % columns, "sheet_row": index // columns,
            "advance": advance,
            "rows_hex": "".join("%04x" % v for v in bits),
        })

    out = pathlib.Path(args.out)
    out.mkdir(parents=True, exist_ok=True)
    sheet = out / f"{args.name}.png"
    write_png_rgba(sheet, sheet_w, sheet_h, bytes(pixels))
    manifest = {
        "schema": "kol-font-manifest-v2",
        "tool": "tools/build_cjk_font.py",
        "note": "中文 UI 字模表。執行期只讀 PNG，不在執行期做字型排版。",
        "rights": {"source_family": args.family, "source_license": args.license,
                   "classification": "字型著作權屬原作者；散布時必須隨附授權全文",
                   "derived": "本表是指定字元在固定大小下的點陣化結果"},
        "bake": {"family": args.family, "size": args.size, "threshold": args.threshold,
                 "cell_pitch": CELL_PITCH, "draw_margin": DRAW_MARGIN,
                 "vertical_offset": yoff, "horizontal_offset": xoff},
        "fonts": [{
            "name": args.name, "file": sheet.name,
            "advance": cell, "cell": {"width": cell, "height": cell_h},
            "sheet": {"columns": columns, "width": sheet_w, "height": sheet_h},
            "glyph_count": len(glyphs),
            "png_sha256": hashlib.sha256(sheet.read_bytes()).hexdigest(),
            "glyphs": glyphs,
        }],
    }
    (out / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"characters": len(chars), "sheet": str(sheet),
                      "sheet_size": [sheet_w, sheet_h],
                      "offset": [xoff, yoff], "cell": [cell, cell_h]},
                     ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
