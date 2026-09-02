#!/usr/bin/env python3
"""從原版戶外畫面反解鏡頭座標與隊伍標記畫格（RE-059／RE-060 的驗證入口）。

畫面上 12×6 格的磚塊像素直接對回 `TOWNn.0` 的磚塊集合，得到一組磚塊編號；
再拿這組編號去 64×64 的格子陣列裡找唯一的鏡頭位置。隊伍那一格會被標記蓋掉，
所以改用十個標記精靈逐一合成後比對，同時解出動畫畫格。

  python3 tools/solve_outdoor_camera.py --frames <目錄> --out 結果.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_ega_assets import payload  # noqa: E402
from inventory_image_archives import (EXPECTED_ZIP_SHA256,  # noqa: E402
                                      logical_archive_contract)
from png_io import read_png  # noqa: E402
from probe_extract_comb_png import PALETTE  # noqa: E402
from probe_extract_party_marker_png import UNIT_COUNT, unit_pixels  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402
from probe_town_map_grid import (COL_BIAS, MAP_BYTES, ROW_BIAS,  # noqa: E402
                                 TILE_BASE, TILE_BYTES, VIEW_COLS, VIEW_ROWS)

VIEW_X, VIEW_Y, TILE = 104, 58, 16
# 認得出磚塊的格數下限：71 格扣掉滑鼠游標最多蓋住的 6 格還有餘裕，
# 低於這個數就不是戶外畫面，不能拿來解鏡頭。
MIN_KNOWN_CELLS = 50
# 圖示列第一格：外框 24×22 在 (32,170)，框內 22×20 對圖示的第 1..22 欄。
ICON_X, ICON_Y, ICON_W, ICON_H = 32, 170, 24, 22
MARKER_ROW, MARKER_COL = -ROW_BIAS, -COL_BIAS


def town_segment(zip_path: pathlib.Path, town: str) -> bytes:
    with zipfile.ZipFile(zip_path) as zf:
        volumes = read_volumes(zf, town)
        _, parts = payloads(volumes, town)
    seg = bytearray(0x9000)
    for index, offset in ((0, 0x0000), (1, 0x3000), (2, 0x6000)):
        seg[offset:offset + 0x3000] = lzw_decode(parts[index])[:0x3000]
    return bytes(seg)


def tile_table(seg: bytes) -> dict[bytes, int]:
    """位元組相同的磚塊只留編號最小的那一個。"""
    table: dict[bytes, int] = {}
    for n in range((0x9000 - TILE_BASE) // TILE_BYTES):
        table.setdefault(planar(seg, TILE_BASE + n * TILE_BYTES), n)
    return table


def planar(buf: bytes, base: int) -> bytes:
    planes = [buf[base + p * 32:base + (p + 1) * 32] for p in range(4)]
    out = bytearray(256)
    for n in range(256):
        bit = 0x80 >> (n & 7)
        out[n] = sum((1 << p) if planes[p][n >> 3] & bit else 0 for p in range(4))
    return bytes(out)


def composite(sprite: bytes, tile: bytes) -> bytes:
    """`sub_1D21D`：精靈佔用位圖左右各膨脹一像素（不跨位元組），其餘保留背景。"""
    out = bytearray(tile)
    for y in range(16):
        for x in range(16):
            lo, hi = (0, 7) if x < 8 else (8, 15)
            near = sprite[y * 16 + x] != 0
            if not near and x - 1 >= lo:
                near = sprite[y * 16 + x - 1] != 0
            if not near and x + 1 <= hi:
                near = sprite[y * 16 + x + 1] != 0
            if near:
                out[y * 16 + x] = sprite[y * 16 + x]
    return bytes(out)


def frame_tiles(rgb: bytes, width: int, index: dict) -> list[list[bytes]] | None:
    rows = []
    for ty in range(VIEW_ROWS):
        line = []
        for tx in range(VIEW_COLS):
            cell = bytearray(256)
            for y in range(TILE):
                base = ((VIEW_Y + ty * TILE + y) * width + VIEW_X + tx * TILE) * 3
                for x in range(TILE):
                    key = tuple(rgb[base + x * 3:base + x * 3 + 3])
                    value = index.get(key)
                    if value is None:
                        return None
                    cell[y * TILE + x] = value
            line.append(bytes(cell))
        rows.append(line)
    return rows


def solve(cells, table, grid, markers, max_x=63, max_y=63):
    """先用非標記格解鏡頭，再解標記畫格。"""
    # 隊伍標記會蓋住一格，原版的滑鼠游標會蓋住另外幾格；認不出磚塊的格就跳過。
    known = {}
    for row in range(VIEW_ROWS):
        for col in range(VIEW_COLS):
            tile = table.get(cells[row][col])
            if tile is not None:
                known[(row, col)] = tile
    if len(known) < MIN_KNOWN_CELLS:
        return None
    hits = []
    for cam_x in range(max_x + 1):
        for cam_y in range(max_y + 1):
            ok = True
            for (row, col), tile in known.items():
                x = min(max(cam_x + col + COL_BIAS, 0), max_x)
                y = min(max(cam_y + row + ROW_BIAS, 0), max_y)
                if table.get(planar_cache(grid, x, y), -1) != tile:
                    ok = False
                    break
            if ok:
                hits.append((cam_x, cam_y))
    if len(hits) != 1:
        return {"camera_hits": hits}
    cam_x, cam_y = hits[0]
    # 標記不一定在正中格：`word_2DF36`／`word_2DF38` 是它在視窗內的欄／列偏移，
    # 所以逐格試合成，命中的那一格就是隊伍所在。
    for row in range(VIEW_ROWS):
        for col in range(VIEW_COLS):
            if (row, col) in known:
                continue
            x = min(max(cam_x + col + COL_BIAS, 0), max_x)
            y = min(max(cam_y + row + ROW_BIAS, 0), max_y)
            under = grid[(x << 6) + y]
            for unit in range(UNIT_COUNT):
                if composite(markers[unit], TILE_PIXELS[under]) == cells[row][col]:
                    return {"camera_x": cam_x, "camera_y": cam_y,
                            "marker_unit": unit, "marker_row": row, "marker_col": col,
                            "marker_row_offset": row - MARKER_ROW,
                            "marker_col_offset": col - MARKER_COL,
                            "tile_under_marker": under, "known_cells": len(known)}
    return {"camera_x": cam_x, "camera_y": cam_y, "marker_unit": None,
            "known_cells": len(known)}


TILE_PIXELS: dict[int, bytes] = {}
_GRID_TILE: dict[tuple[int, int], bytes] = {}


def planar_cache(grid: bytes, x: int, y: int) -> bytes:
    key = (x, y)
    if key not in _GRID_TILE:
        _GRID_TILE[key] = TILE_PIXELS[grid[(x << 6) + y]]
    return _GRID_TILE[key]


def icon_slot(rgb: bytes, width: int, assets: pathlib.Path,
              background: bytes) -> int | None:
    """回傳圖示列第一格的 KOLMPICS 圖號；整格與底圖相同時回傳 None（沒有圖示）。"""
    same = True
    for y in range(ICON_H):
        for x in range(ICON_W):
            a = rgb[((ICON_Y + y) * width + ICON_X + x) * 3:
                    ((ICON_Y + y) * width + ICON_X + x) * 3 + 3]
            b = background[((ICON_Y + y) * 320 + ICON_X + x) * 3:
                           ((ICON_Y + y) * 320 + ICON_X + x) * 3 + 3]
            if a != b:
                same = False
                break
        if not same:
            break
    if same:
        return None
    for path in sorted(assets.glob("kolmpics-*.png")):
        image = read_png(path)
        if (image.width, image.height) != (ICON_W, ICON_H - 2):
            continue
        px = image.rgb()
        if all(rgb[((ICON_Y + 1 + y) * width + ICON_X + 1 + x) * 3:
                   ((ICON_Y + 1 + y) * width + ICON_X + 1 + x) * 3 + 3]
               == px[(y * image.width + 1 + x) * 3:(y * image.width + 1 + x) * 3 + 3]
               for y in range(ICON_H - 2) for x in range(ICON_W - 2)):
            return int(path.stem.split("-")[-1])
    return -1


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--town", default="TOWN7")
    p.add_argument("--frames", required=True)
    p.add_argument("--assets", default="assets/original-derived-png/ega")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    zip_path = pathlib.Path(args.zip)
    if hashlib.sha256(zip_path.read_bytes()).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")
    seg = town_segment(zip_path, args.town)
    grid = seg[:MAP_BYTES]
    table = tile_table(seg)
    for n in range((0x9000 - TILE_BASE) // TILE_BYTES):
        TILE_PIXELS[n] = planar(seg, TILE_BASE + n * TILE_BYTES)

    with zipfile.ZipFile(zip_path) as zf:
        files = {i.filename: zf.read(i) for i in zf.infolist() if not i.is_dir()}
    kolpics = payload(files, logical_archive_contract(
        "Knights-of-legend/KOLPICS", files)["items"][0])
    markers = [unit_pixels(kolpics, i) for i in range(UNIT_COUNT)]

    background = read_png(pathlib.Path(args.assets) / "KOLMPICS/kolmpics-000.png").rgb()
    index = {tuple(rgb): i for i, rgb in enumerate(PALETTE)}
    results = []
    for path in sorted(pathlib.Path(args.frames).glob("frame-*.png")):
        image = read_png(path)
        cells = frame_tiles(image.rgb(), image.width, index)
        if cells is None:
            results.append({"frame": path.name, "outdoor": False})
            continue
        solved = solve(cells, table, grid, markers)
        entry = {"frame": path.name, "outdoor": solved is not None}
        if solved:
            entry.update(solved)
            entry["icon"] = icon_slot(image.rgb(), image.width,
                                      pathlib.Path(args.assets) / "KOLMPICS", background)
        results.append(entry)

    solved = [r for r in results if r.get("camera_x") is not None]
    states = sorted({(r["camera_x"], r["camera_y"], r["marker_unit"],
                      r.get("marker_row_offset"), r.get("marker_col_offset"))
                     for r in solved},
                    key=lambda s: tuple(-1 if v is None else v for v in s))
    receipt = {
        "schema": "kol-outdoor-camera-solve-v1",
        "town": args.town,
        "frames_dir": args.frames,
        "frame_count": len(results),
        "outdoor_frames": len(solved),
        "unsolved_frames": [r["frame"] for r in results
                            if r.get("outdoor") and r.get("camera_x") is None],
        "distinct_states": [{"camera_x": s[0], "camera_y": s[1], "marker_unit": s[2],
                             "marker_row_offset": s[3], "marker_col_offset": s[4]}
                            for s in states],
        "frames": results,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({k: v for k, v in receipt.items() if k != "frames"},
                     ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
