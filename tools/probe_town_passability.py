#!/usr/bin/env python3
"""探針：城鎮地圖的可通行位元圖（`TOWNn.0` payload 3 的位移 0x00–0x1F）。

`sub_10000:0x1082D` 在移動前先算出目的地格子的磚塊編號，再查兩張表：

    mask = byte_2DF2E[tile & 7]        ; DGROUP 0x5E ＝ 01 02 04 08 10 20 40 80
    if payload3[tile >> 3] & mask: 不能走

所以 payload 3 的前 32 個 byte 是「磚塊編號 → 可否通行」的位元圖，一個 bit 一個磚塊，
LSB 先。本探針把八座城的位元圖解出來，並與地圖上實際用到的磚塊交叉比對。

  tools/run_python_tool.sh tools/probe_town_passability.py
"""

from __future__ import annotations

import argparse
import collections
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402
from probe_town_doors import COUNT_OFFSET, DOOR_OFFSET  # noqa: E402

DOOR_TABLE_PAYLOAD = 3
BLOCK_BITMAP_BYTES = 32
MAP_PAYLOAD = 0            # payload 0：64×64 的格子陣列
GRID_SIDE = 64


def blocked_tiles(table: bytes) -> list[int]:
    return [t for t in range(BLOCK_BITMAP_BYTES * 8)
            if table[t >> 3] & (1 << (t & 7))]


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    args = p.parse_args()
    zip_path = pathlib.Path(args.zip)
    if hashlib.sha256(zip_path.read_bytes()).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    out = []
    with zipfile.ZipFile(zip_path) as zf:
        for town in range(0, 8):
            root = f"TOWN{town}"
            contract, parts = payloads(read_volumes(zf, root), root)
            table = lzw_decode(parts[DOOR_TABLE_PAYLOAD])
            grid = lzw_decode(parts[MAP_PAYLOAD])
            used = collections.Counter(grid[:GRID_SIDE * GRID_SIDE])
            blocked = set(blocked_tiles(table))
            used_blocked = sorted(t for t in used if t in blocked)
            used_open = sorted(t for t in used if t not in blocked)
            doors = []
            for i in range(table[COUNT_OFFSET] + 1):
                dx, dy = table[DOOR_OFFSET + 2 * i], table[DOOR_OFFSET + 2 * i + 1]
                tile = grid[dx * GRID_SIDE + dy]
                doors.append({"door": i, "x": dx, "y": dy, "tile": tile,
                              "blocked": tile in blocked})
            out.append({
                "doors_on_blocked_tiles": [d for d in doors if d["blocked"]],
                "door_count": len(doors),
                "town": town,
                "bitmap": table[:BLOCK_BITMAP_BYTES].hex(),
                "blocked_count": len(blocked),
                "used_tiles": len(used),
                "used_blocked": used_blocked,
                "used_open": used_open,
                "blocked_cells": sum(n for t, n in used.items() if t in blocked),
                "open_cells": sum(n for t, n in used.items() if t not in blocked),
            })
    print(json.dumps(out, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
