#!/usr/bin/env python3
"""野外地圖的 typed data（RE-066／RE-068）。

`TOWN0` 的 payload 3 由 `sub_29004:0x29031` 讀進 `byte_2FBC4`。已讀出的版面：

  0x100 + 磚塊編號   地形位元；最低的那個 1 的位序就是地形碼，全為 0 時是 7
  0x200 + 磚塊編號   道路旗標（`word_2FA2A`；非零時圖示列多出 ROAD／SIGNPOST）
  0x300 + 磚塊/8     不可通行位元圖（查法與城鎮相同，只是基底不同）
  0x323 + i          24 個特殊地點的 X（0x33B + i 是 Y）
  0x393 + 16i        22 筆城鎮記錄（見 RE-066 的欄位表）

地圖格子陣列與磚圖沿用 `TOWN0` 的 payload 0…2（與城鎮同一個版面），
runtime 讀的是抽取出來的 PNG，這裡只放數值表。

  python3 tools/extract_world_map.py --out assets/original-derived-data/world-map.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
MAP_PAYLOAD, TABLE_PAYLOAD = 0, 3
TERRAIN_TABLE, ROAD_TABLE, BLOCK_BITMAP = 0x100, 0x200, 0x300
SPECIAL_X, SPECIAL_Y, SPECIAL_COUNT = 0x323, 0x33B, 24
TOWN_RECORDS, TOWN_RECORD_BYTES, TOWN_RECORD_COUNT = 0x393, 16, 22
TILE_COUNT, GRID_SIDE = 256, 64
NO_ONE = 0xFF
# 地形碼 → KOLDATA 記錄 0xDA + 碼；時段是記錄 0xD5 + 時段。
TERRAIN_RECORD_BASE, TIME_RECORD_BASE = 0xDA, 0xD5
TERRAIN_NAMES = ("DESERT", "RIVER", "SWAMP", "HILL",
                 "ROAD", "FOREST", "SHORE", "PLAINS")
TIME_NAMES = ("DAWN", "MID MORN", "NOON", "AFT NOON", "TWILIGHT")


def terrain_code(bits: int) -> int:
    """`sub_29987:0x299C7`：全為 0 是 7，否則是最低那個 1 的位序。"""
    if bits == 0:
        return 7
    code = 0
    while bits & 1 == 0:
        bits >>= 1
        code += 1
    return code


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    out = pathlib.Path(args.out)
    if ALLOWED_OUTPUT_PARENT not in pathlib.PurePosixPath(out.as_posix()).parents:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")
    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(zip_path) as zf:
        _, parts = payloads(read_volumes(zf, "TOWN0"), "TOWN0")
        table = lzw_decode(parts[TABLE_PAYLOAD])
        grid = lzw_decode(parts[MAP_PAYLOAD])

    blocked = [t for t in range(TILE_COUNT)
               if table[BLOCK_BITMAP + (t >> 3)] & (1 << (t & 7))]
    terrain = [terrain_code(table[TERRAIN_TABLE + t]) for t in range(TILE_COUNT)]
    road = [table[ROAD_TABLE + t] for t in range(TILE_COUNT)]

    towns = []
    for i in range(TOWN_RECORD_COUNT):
        base = TOWN_RECORDS + TOWN_RECORD_BYTES * i
        raw = table[base:base + TOWN_RECORD_BYTES]
        if raw[0] == 0 and raw[1] == 0:
            continue
        code = "".join(chr(b & 0x7F) for b in raw[9:12])
        towns.append({
            "record": i, "x": raw[0], "y": raw[1], "index": raw[2],
            "camera": {"x": raw[3], "y": raw[4]},
            "bounds": {"min_x": raw[5], "max_x": raw[6],
                       "min_y": raw[7], "max_y": raw[8]},
            "code": code,
            "disk": f"{chr(raw[12])}{chr(raw[13])}",
            # 記錄 i 進的是 TOWN{i+1}；i ≥ 6 全部進 TOWN7，用入城點 i−6 區分。
            "town": i + 1 if i < 6 else 7,
            "entry_point": 0 if i < 6 else i - 6,
        })

    special = [{"index": i, "x": table[SPECIAL_X + i], "y": table[SPECIAL_Y + i]}
               for i in range(SPECIAL_COUNT)]

    used = sorted({grid[i] for i in range(GRID_SIDE * GRID_SIDE)})
    doc = {
        "schema": "kol-world-map-data-v1",
        "tool": "tools/extract_world_map.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWN0",
                           "map_payload": MAP_PAYLOAD, "table_payload": TABLE_PAYLOAD},
        "evidence": "RE-066；RE-068",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "terrain": f"payload{TABLE_PAYLOAD}[0x{TERRAIN_TABLE:X} + 磚塊編號]："
                       "地形位元，最低的 1 的位序就是地形碼；全為 0 時是 7",
            "road": f"payload{TABLE_PAYLOAD}[0x{ROAD_TABLE:X} + 磚塊編號]：道路旗標",
            "blocked": f"payload{TABLE_PAYLOAD}[0x{BLOCK_BITMAP:X}…]：不可通行位元圖",
            "towns": f"payload{TABLE_PAYLOAD}[0x{TOWN_RECORDS:X} + 16i]：城鎮記錄",
            "special": f"payload{TABLE_PAYLOAD}[0x{SPECIAL_X:X}／0x{SPECIAL_Y:X} + i]",
        },
        "names": {
            "terrain": list(TERRAIN_NAMES), "time": list(TIME_NAMES),
            "terrain_record_base": TERRAIN_RECORD_BASE,
            "time_record_base": TIME_RECORD_BASE,
            "note": "名稱本身是 KOLDATA 的記錄，這裡只記對照；譯文走 i18n 表。",
        },
        "grid": {"side": GRID_SIDE, "png": "assets/original-derived-png/ega/TOWN0/town0-map.png",
                 "sha256": hashlib.sha256(grid[:GRID_SIDE * GRID_SIDE]).hexdigest(),
                 "used_tiles": len(used)},
        "table_sha256": hashlib.sha256(table).hexdigest(),
        "blocked_tiles": blocked,
        "terrain_by_tile": terrain,
        "road_by_tile": road,
        "towns": towns,
        "special_locations": special,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"towns": len(towns), "blocked": len(blocked),
                      "used_tiles": len(used), "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
