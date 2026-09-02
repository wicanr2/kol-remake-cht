"""驗證 KOLTOWN 的城鎮地圖格子陣列版面（RE-059）。

反組譯給的契約（`sub_115D1` ＋ `sub_1D1C6`）是：

  磚塊來源段 ＝ `word_30C8A + 0x100`，也就是段內位移 **0x1000**；磚距 `shl ax,7` ＝ 128。
  格子索引 ＝ `(X << 6) + Y`，X 由 `word_2DF68`／`word_2DF6A` 夾限、Y 由
  `word_2DF66`／`word_2DF6C` 夾限，而 `word_2DF6C` 靜態就是 `3Fh`、`word_2DF68` 是 0。

`sub_15920` 把 `TOWNn.0` 的 payload 0／1／2 接到同一段的 0x0000／0x3000／0x6000。
兩件事合起來的推論是：**0x0000–0x0FFF 是 64×64 的格子陣列，0x1000–0x8FFF 是 256 塊磚**。

本工具用資料驗這個推論：把畫面上量到的 12×6 磚塊編號當指紋，在格子陣列裡找鏡頭位置。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import zipfile
from pathlib import Path

from probe_town_data import lzw_decode, payloads, read_volumes

MAP_BYTES = 0x1000       # 格子陣列大小
TILE_BASE = 0x1000       # 磚塊區起點（＝ segment +0x100 paragraph）
TILE_BYTES = 128
GRID_STRIDE = 64         # (X << 6)
VIEW_COLS, VIEW_ROWS = 12, 6
COL_BIAS, ROW_BIAS = -5, -2   # 0FFFBh / 0FFFEh


def build_segment(zip_path: Path, town: str) -> tuple[dict, bytes]:
    with zipfile.ZipFile(zip_path) as zf:
        volumes = read_volumes(zf, town)
        contract, parts = payloads(volumes, town)
    seg = bytearray(0x9000)
    sizes = {}
    for index, offset in ((0, 0x0000), (1, 0x3000), (2, 0x6000)):
        chunk = lzw_decode(parts[index])       # sub_1A4FE 解出來的才是段內內容
        sizes[index] = len(chunk)
        seg[offset:offset + len(chunk)] = chunk
    contract = dict(contract, decoded_payload_bytes=sizes)
    return contract, bytes(seg)


def tile_classes(seg: bytes) -> dict[int, int]:
    """位元組完全相同的磚塊要視為同一塊——比對器只會報出編號最小的那一個。"""
    first: dict[bytes, int] = {}
    out = {}
    for n in range((0x9000 - TILE_BASE) // TILE_BYTES):
        body = seg[TILE_BASE + n * TILE_BYTES:TILE_BASE + (n + 1) * TILE_BYTES]
        out[n] = first.setdefault(body, n)
    return out


def screen_tiles(probe: Path) -> dict[tuple[int, int], int]:
    data = json.loads(probe.read_text(encoding="utf-8"))
    out = {}
    for m in data["tile_matches"]:
        absolute = m["payload"] * 0x3000 + m["offset"]
        if absolute < TILE_BASE:
            raise SystemExit("有磚塊命中落在格子陣列區，版面推論不成立")
        out[(m["tile_y"], m["tile_x"])] = (absolute - TILE_BASE) // TILE_BYTES
    return out


def search(grid: bytes, want: dict[tuple[int, int], int], classes: dict[int, int],
           max_x: int, min_y: int, max_y: int) -> tuple[list[dict], int]:
    hits = []
    best = 0
    for cam_x in range(0, max_x + 1):
        for cam_y in range(min_y, max_y + 1):
            score = 0
            for (row, col), tile in want.items():
                x = min(max(cam_x + col + COL_BIAS, 0), max_x)
                y = min(max(cam_y + row + ROW_BIAS, min_y), max_y)
                if classes[grid[(x << 6) + y]] == tile:
                    score += 1
            best = max(best, score)
            if score == len(want):
                hits.append({"camera_x": cam_x, "camera_y": cam_y})
    return hits, best


def render(grid: bytes, cam_x: int, cam_y: int,
           max_x: int, min_y: int, max_y: int) -> list[list[int]]:
    out = []
    for row in range(VIEW_ROWS):
        line = []
        for col in range(VIEW_COLS):
            x = min(max(cam_x + col + COL_BIAS, 0), max_x)
            y = min(max(cam_y + row + ROW_BIAS, min_y), max_y)
            line.append(grid[(x << 6) + y])
        out.append(line)
    return out


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--town", default="TOWN7")
    p.add_argument("--probe", default="workplace/town-data/town7-probe.json")
    p.add_argument("--max-x", type=int, default=63)
    p.add_argument("--min-y", type=int, default=0)
    p.add_argument("--max-y", type=int, default=63)
    p.add_argument("--out", required=True)
    args = p.parse_args()

    zip_path = Path(args.zip)
    contract, seg = build_segment(zip_path, args.town)
    grid = seg[:MAP_BYTES]
    want = screen_tiles(Path(args.probe))
    classes = tile_classes(seg)
    hits, best = search(grid, want, classes, args.max_x, args.min_y, args.max_y)

    used = sorted({grid[i] for i in range(MAP_BYTES)})
    receipt = {
        "schema": "kol-town-map-grid-v1",
        "town": args.town,
        "zip_sha256": hashlib.sha256(zip_path.read_bytes()).hexdigest(),
        "archive_contract": contract,
        "layout": {
            "map_bytes": MAP_BYTES,
            "grid_stride": GRID_STRIDE,
            "grid_rows_x": MAP_BYTES // GRID_STRIDE,
            "tile_base": TILE_BASE,
            "tile_bytes": TILE_BYTES,
            "tile_count": (0x9000 - TILE_BASE) // TILE_BYTES,
        },
        "screen_cells_known": len(want),
        "distinct_screen_tiles": len(set(want.values())),
        "screen_tile_range": [min(want.values()), max(want.values())],
        "grid_distinct_values": len(used),
        "grid_value_range": [used[0], used[-1]],
        "duplicate_tile_ids": {str(k): v for k, v in classes.items() if v != k},
        "best_score": best,
        "camera_hits": hits,
    }
    if len(hits) == 1:
        cam = hits[0]
        receipt["rendered_view"] = render(grid, cam["camera_x"], cam["camera_y"],
                                          args.max_x, args.min_y, args.max_y)
        receipt["measured_view"] = [
            [want.get((r, c)) for c in range(VIEW_COLS)] for r in range(VIEW_ROWS)]
    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                   encoding="utf-8")
    print(json.dumps({k: v for k, v in receipt.items()
                      if k not in ("rendered_view", "measured_view",
                                   "archive_contract")},
                     ensure_ascii=False, indent=2))
    return 0


raise SystemExit(main())
