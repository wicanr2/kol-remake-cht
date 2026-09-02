#!/usr/bin/env python3
"""探針：城鎮的 MOUTH 關鍵字表（`TOWNn.0` 的 payload 20）。

`sub_11A93:0x122CE` 的迴圈說明版面：`var_2` 從 0 到 0x100 步進 4，每一輪
`sub_1B94F(0D39Eh + var_2, 1)` 解出一個關鍵字，比對玩家輸入；命中就讀
`[0D39Eh + var_2 + 3]` 這個碼，拆成 `碼 / 0x40` 與 `碼 & 0x3F`，
再拿 `碼 & 0x3F` 去比對建築 payload 位移 0x17…0x1C 的六個 byte（各取低 6 bit），
對上第 var_A 個就顯示訊息 20 + var_A。沒對上就看 `[0D39Eh + 0x1C0 + 碼]`。

`0D39Eh` 是 DGROUP 位移，`sub_10000:0x106C9` 把 payload 0x14 讀進這裡。

  tools/run_python_tool.sh tools/probe_town_keywords.py --town 1
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from extract_town_text import render  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402
from probe_town_doors import text_payload_index  # noqa: E402

KEYWORD_PAYLOAD = 20
ENTRY_BYTES = 4
ENTRY_COUNT = 64
TOPIC_TABLE = 0x100      # 每筆 8 byte 的第二張文字表
TOPIC_BYTES = 8
CODE_TABLE = 0x1C0       # 沒對上時的後備碼表
BUILDING_TOPICS = 0x17   # 建築 payload 位移 0x17…0x1C：六個主題碼


def decode_mode1(data: bytes, offset: int) -> str:
    """`sub_1B94F` 的 mode 1：字數放在第 0 個 byte 的高 4 bit，位元流從同一個 byte
    的 bit 3 開始，而且跳脫狀態一開始就是打開的（第一個字母大寫）。"""
    count = data[offset] >> 4
    if count == 0:
        return ""
    bits = []
    for b in range(4):
        byte = data[offset + b]
        for shift in (3, 2, 1, 0) if b == 0 else (7, 6, 5, 4, 3, 2, 1, 0):
            bits.append((byte >> shift) & 1)
    values = []
    for i in range(count):
        v = 0
        for bit in bits[i * 5:(i + 1) * 5]:
            v = (v << 1) | bit
        values.append(v)
    text, _ = render(values)
    return text


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--town", type=int, required=True)
    args = p.parse_args()
    zip_path = pathlib.Path(args.zip)
    if hashlib.sha256(zip_path.read_bytes()).hexdigest() != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")
    root = f"TOWN{args.town}"
    with zipfile.ZipFile(zip_path) as zf:
        contract, parts = payloads(read_volumes(zf, root), root)
        table = lzw_decode(parts[KEYWORD_PAYLOAD])
        buildings = {}
        for door in range(0, 24):
            index = text_payload_index(door)
            if index >= contract["count"]:
                continue
            body = lzw_decode(parts[index])
            buildings[door] = list(body[BUILDING_TOPICS:BUILDING_TOPICS + 6])

    entries = []
    for i in range(ENTRY_COUNT):
        off = i * ENTRY_BYTES
        entries.append({
            "index": i, "offset": off,
            "bytes": table[off:off + ENTRY_BYTES].hex(),
            "keyword": decode_mode1(table, off),
            "code": table[off + 3],
            "code_high": table[off + 3] >> 6,
            "code_low": table[off + 3] & 0x3F,
        })
    topics = []
    for i in range(32):
        off = TOPIC_TABLE + i * TOPIC_BYTES
        if off + TOPIC_BYTES > len(table):
            break
        topics.append({"index": i, "offset": off,
                       "bytes": table[off:off + TOPIC_BYTES].hex(),
                       "text": decode_mode1(table, off)})
    print(json.dumps({
        "town": args.town, "payload": KEYWORD_PAYLOAD, "length": len(table),
        "sha256": hashlib.sha256(table).hexdigest(),
        "entries": entries, "topics": topics,
        "code_table": table[CODE_TABLE:CODE_TABLE + 64].hex(),
        "building_topics": buildings,
    }, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
