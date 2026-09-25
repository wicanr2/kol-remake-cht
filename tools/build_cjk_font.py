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
import csv
import hashlib
import json
import os
import pathlib
import struct
import subprocess
import sys
import zlib

CELL_PITCH = 48   # 烘焙時的格距：一格要放得下「基準字＋目標字」兩個字
DRAW_MARGIN = 8   # 每格內縮的繪製起點；鄰格字模的側邊突出才不會被量進來
ASCII_MAX = 0x7F

# drawtext 是**依照它這一次要畫的字串**的最大 ascender 對齊的，所以一個字單獨畫時，
# 它的墨跡頂端一律落在 y —— 逗號、句號這種本來坐在基線上的字會被拉到格子最上面。
# 解法是每一格都先畫一個基準字再畫目標字：兩個字在同一次 drawtext 裡就共用基線。
# 基準字的墨跡頂端在每一格量一次；目標字比基準字高時整串會下移，那個位移量正好由
# 基準字自己的墨跡頂端反映出來，因此不必假設基準字是最高的。
# 基準字挑全形驚嘆號：advance 是一整個 em，墨跡卻只有中間兩三個像素寬，
# 因此和目標字的墨跡不會在分界線上黏在一起。
BASELINE_REF = "！"
REF_SPLIT = 14    # 基準字（全形，advance ＝ 1 em）與目標字的分界 x


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
            extra: list[pathlib.Path] | None = None,
            extra_tsv: list[pathlib.Path] | None = None) -> list[str]:
    data = json.loads(strings_path.read_text())
    seen: set[str] = set()
    for key in ("labels", "races", "secondary", "professions", "ui"):
        for item in data.get(key, []):
            seen.update(item["zh"])
    # 譯文不只住在字串表：城鎮畫面的對拍樣本自帶預先斷行的中文台詞，
    # 那些字沒烘進來就會在畫面上變成空白。
    for path in extra or []:
        _walk(json.loads(path.read_text()), seen)
    for path in extra_tsv or []:
        with path.open(encoding="utf-8", newline="") as handle:
            rows = csv.DictReader(handle, delimiter="\t")
            if rows.fieldnames != ["key", "en", "zh-Hant", "status", "source"]:
                raise SystemExit(f"資料名稱譯表欄位不符：{path}")
            for row in rows:
                seen.update(ch for ch in row["zh-Hant"] if ord(ch) > 0x7F)
    # 玩家輸入的名字與屬性數值都是 ASCII，因此整段可見 ASCII 一律烘進來，
    # 中文模式才不需要在兩套字型之間切換。
    seen.update(chr(c) for c in range(0x21, 0x7F))
    seen.discard(" ")
    return sorted(seen)


# 一次交給 ffmpeg 的字模列數。整張表一次烘會踩到兩個上限：單一命令列參數 128 KB
# （`MAX_ARG_STRLEN`），以及上千個 drawtext 濾鏡各自載入字型的記憶體。分批烘再拼接，
# 兩者都不會碰到，而且輸出仍然逐位元可重生。
ROWS_PER_BATCH = 16


def render_batch(chars: list[str], family: str, size: int, columns: int,
                 work: pathlib.Path, tag: str) -> bytes:
    """把一批字元烘到固定格點上，回傳 PNG bytes。

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
        path.write_text(BASELINE_REF + ch, encoding="utf-8")
        filters.append(
            f"drawtext=font={family}:textfile={path}:expansion=none"
            f":fontsize={size}:fontcolor=white:x={x}:y={y}")
    graph_path = work / f"filtergraph-{tag}.txt"
    graph_path.write_text(",".join(filters), encoding="utf-8")
    out = work / f"sheet-{tag}.png"
    result = subprocess.run(
        ["ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
         "-f", "lavfi", "-i", f"color=c=black:s={width}x{height}:d=1",
         "-filter_script:v", str(graph_path), "-frames:v", "1", str(out)],
        capture_output=True, text=True)
    if result.returncode != 0:
        # 取尾端不是開頭：ffmpeg 先印一整段輸入資訊，真正的錯誤在最後面。
        raise SystemExit("ffmpeg 烘焙失敗：\n" + result.stderr.strip()[-2000:])
    return out.read_bytes()


def render(chars: list[str], family: str, size: int, columns: int,
           work: pathlib.Path) -> tuple[int, int, list[int]]:
    """分批烘出整張字模表，回傳（寬, 高, 灰階像素）。"""
    per_batch = ROWS_PER_BATCH * columns
    width = columns * CELL_PITCH
    grey: list[int] = []
    for start in range(0, len(chars), per_batch):
        batch = chars[start:start + per_batch]
        raw = render_batch(batch, family, size, columns, work,
                           f"{start // per_batch:03d}")
        bw, bh, part = decode_png_gray(raw)
        if bw != width:
            raise SystemExit(f"批次寬度 {bw} 不是 {width}")
        grey.extend(part)
    rows = (len(chars) + columns - 1) // columns
    height = rows * CELL_PITCH
    if len(grey) != width * height:
        raise SystemExit(f"拼接後像素數 {len(grey)} 不是 {width * height}")
    return width, height, grey


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
    p.add_argument("--extra-tsv", action="append", default=[],
                   help="另掃描資料名稱 TSV 的繁中譯名（可重複）")
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
                    [pathlib.Path(x) for x in args.extra_json],
                    [pathlib.Path(x) for x in args.extra_tsv])
    work = pathlib.Path(args.work)
    work.mkdir(parents=True, exist_ok=True)
    width, height, grey = render(chars, args.family, args.size, args.columns, work)

    # 每一格畫的是「基準字＋目標字」。先量出基準字的墨跡頂端，那就是這一格的基線；
    # 目標字的縱向位置一律換算成「相對基線」，逗號、句號才會坐回原來的高度。
    cell_h = args.cell_height or args.cell

    def ink_box(index: int, x0: int, x1: int) -> tuple[int, int, int, int] | None:
        ox = (index % args.columns) * CELL_PITCH + DRAW_MARGIN
        oy = (index // args.columns) * CELL_PITCH + DRAW_MARGIN
        t, b, l, r = height, -height, width, -width
        for y in range(-DRAW_MARGIN, CELL_PITCH - DRAW_MARGIN):
            for x in range(x0, x1):
                if grey[(oy + y) * width + ox + x] >= args.threshold:
                    t, b = min(t, y), max(b, y)
                    l, r = min(l, x), max(r, x)
        return (t, b, l, r) if b >= t else None

    ref_tops = []
    for index in range(len(chars)):
        ref = ink_box(index, 0, REF_SPLIT)
        if ref is None:
            raise SystemExit(f"第 {index} 格量不到基準字 {BASELINE_REF} 的墨跡")
        if ref[3] >= REF_SPLIT - 1:
            raise SystemExit(f"基準字墨跡右緣 {ref[3]} 太靠近分界 {REF_SPLIT}")
        ref_tops.append(ref[0])

    # 每個字模自己的墨跡邊界，縱向已經扣掉該格基準字的頂端，因此全部同基線。
    extents = []
    for index in range(len(chars)):
        box = ink_box(index, REF_SPLIT, CELL_PITCH - DRAW_MARGIN)
        if box is None:
            extents.append(None)
            continue
        t, b, l, r = box
        extents.append((t - ref_tops[index], b - ref_tops[index],
                        l - REF_SPLIT, r - REF_SPLIT))

    top, bottom, left, right = height, -height, width, -width
    for e in extents:
        if e is None:
            continue
        top, bottom = min(top, e[0]), max(bottom, e[1])
        left, right = min(left, e[2]), max(right, e[3])

    oversize = [(chars[i], e) for i, e in enumerate(extents)
                if e and (e[1] - e[0] + 1 > cell_h or e[3] - e[2] + 1 > args.cell)]
    if oversize:
        raise SystemExit(
            f"有字模本身就超出 {args.cell}x{cell_h}，只能改小 --size 或加大字格：\n"
            + "\n".join(f"  {c}：高 {e[1] - e[0] + 1}、寬 {e[3] - e[2] + 1}"
                        for c, e in oversize[:8]))
    if bottom - top + 1 > cell_h:
        raise SystemExit(f"字模共同高度 {bottom - top + 1} 超出 {cell_h}，"
                         f"基線位移無法逐字調整")

    yoff, xoff = top, left

    def glyph_xoff(index: int) -> int:
        """全域水平位移；字模因此超出字格時夾回剛好裝得下的位置。"""
        e = extents[index]
        if e is None:
            return xoff
        return min(max(xoff, e[3] - args.cell + 1), e[2])

    clamped = [chars[i] for i in range(len(chars)) if glyph_xoff(i) != xoff]

    cell = args.cell
    columns = args.columns
    rows = (len(chars) + columns - 1) // columns
    sheet_w, sheet_h = columns * cell, rows * cell_h
    pixels = bytearray(sheet_w * sheet_h * 4)
    glyphs = []
    for index, ch in enumerate(chars):
        ox = (index % columns) * CELL_PITCH + DRAW_MARGIN + REF_SPLIT + glyph_xoff(index)
        oy = (index // columns) * CELL_PITCH + DRAW_MARGIN + ref_tops[index] + yoff
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
        "inputs": [{"path": str(path), "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
                   for path in [pathlib.Path(args.strings)]
                   + [pathlib.Path(x) for x in args.extra_json]
                   + [pathlib.Path(x) for x in args.extra_tsv]],
        "rights": {"source_family": args.family, "source_license": args.license,
                   "classification": "字型著作權屬原作者；散布時必須隨附授權全文",
                   "derived": "本表是指定字元在固定大小下的點陣化結果"},
        "bake": {"family": args.family, "size": args.size, "threshold": args.threshold,
                 "cell_pitch": CELL_PITCH, "draw_margin": DRAW_MARGIN,
                 "vertical_offset": yoff, "horizontal_offset": xoff,
                 "horizontally_clamped_glyphs": clamped},
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
