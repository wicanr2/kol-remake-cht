#!/usr/bin/env python3
"""探針：野外地圖 `TOWN0` 的 payload 3 版面（RE-066 候選）。

`sub_29004` 把 `town0` 的 payload 3 讀進 `byte_2FBC4`（DGROUP 位移 0x1CF4），
其後用到的位移有：

  0x200 + tile   逐磚塊的一個 byte（移動前先看，非零就不做「退回原位」）
  0x300 + tile/8 磚塊的不可通行位元圖（與城鎮相同的查法，只是基底不同）
  0x323 + i      24 個特殊地點的 X（0x33B + i 是 Y）→ sub_2B0FF(i)
  0x393 + 16i    22 筆 16 byte 的記錄，前兩個 byte 是 X／Y → 進城

本探針把這幾塊原樣倒出來，讓下一輪可以判定 16 byte 記錄裡還有什麼。

  tools/run_python_tool.sh tools/probe_world_map_data.py
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

TILE_FLAGS = 0x200
BLOCK_BITMAP = 0x300
SPECIAL_X, SPECIAL_Y, SPECIAL_COUNT = 0x323, 0x33B, 24
TOWN_RECORDS, TOWN_RECORD_BYTES, TOWN_RECORD_COUNT = 0x393, 16, 22


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    args = p.parse_args()
    zip_path = pathlib.Path(args.zip)
    if hashlib.sha256(zip_path.read_bytes()).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(zip_path) as zf:
        contract, parts = payloads(read_volumes(zf, "TOWN0"), "TOWN0")
        table = lzw_decode(parts[3])
        extra = lzw_decode(parts[4])

    def ascii_of(raw: bytes) -> str:
        return "".join(chr(b & 0x7F) if 32 <= (b & 0x7F) < 127 else "." for b in raw)

    doc = {
        "payload_count": contract["count"],
        "payload3_len": len(table),
        "payload4_len": len(extra),
        "payload3_sha256": hashlib.sha256(table).hexdigest(),
        "tile_flags": table[TILE_FLAGS:TILE_FLAGS + 256].hex(),
        "block_bitmap": table[BLOCK_BITMAP:BLOCK_BITMAP + 32].hex(),
        "special": [{"index": i, "x": table[SPECIAL_X + i], "y": table[SPECIAL_Y + i]}
                    for i in range(SPECIAL_COUNT)],
        "town_records": [
            {"index": i,
             "x": table[TOWN_RECORDS + TOWN_RECORD_BYTES * i],
             "y": table[TOWN_RECORDS + TOWN_RECORD_BYTES * i + 1],
             "bytes": table[TOWN_RECORDS + TOWN_RECORD_BYTES * i:
                            TOWN_RECORDS + TOWN_RECORD_BYTES * (i + 1)].hex(),
             "ascii": ascii_of(table[TOWN_RECORDS + TOWN_RECORD_BYTES * i:
                                     TOWN_RECORDS + TOWN_RECORD_BYTES * (i + 1)])}
            for i in range(TOWN_RECORD_COUNT)],
        "head_hex": table[:0x200].hex(),
        "payload4_head": extra[:256].hex(),
        "payload4_ascii": ascii_of(extra[:512]),
    }
    print(json.dumps(doc, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
