#!/usr/bin/env python3
"""城鎮文字的 5-bit 解碼與抽取（RE-062）。

`TOWNn.0` 的每個建築 payload 從位移 128 起是一連串 **16 byte 的文字格**：

  byte 0        這一格的字元數（1…24）
  byte 1…15     24 個 5-bit 字元，MSB-first（15 × 8 ÷ 5 ＝ 24）

24 正是城鎮對話框的欄數，所以一格就是畫面上的一行。字元表只有 32 個位置，
因此小寫是預設、大寫靠一個記號：

  0        空白
  1…26     a…z
  27       `"`
  28       `,`
  29       `.`
  30       大寫記號：下一個字母轉大寫
  31       `'`

  python3 tools/extract_town_text.py --out assets/original-derived-data/town-text.json
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
from probe_town_doors import (COUNT_OFFSET, DOOR_OFFSET,  # noqa: E402
                              DOOR_TABLE_PAYLOAD, text_payload_index)

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
TEXT_BASE = 128          # 文字格的起點
SLOT_BYTES = 16
SLOT_CHARS = 24
LETTERS = "abcdefghijklmnopqrstuvwxyz"
SYMBOLS = {27: '"', 28: ",", 29: ".", 31: "'"}
SHIFT = 30
TOWN_FIRST, TOWN_LAST = 1, 7


def slot_values(data: bytes, offset: int, count: int) -> list[int]:
    out = []
    for i in range(count):
        bit = (offset + 1) * 8 + i * 5
        j = bit >> 3
        word = (data[j] << 8) | (data[j + 1] if j + 1 < len(data) else 0)
        out.append((word >> (11 - (bit & 7))) & 0x1F)
    return out


def render(values: list[int]) -> str:
    out = []
    upper = False
    for v in values:
        if v == SHIFT:
            upper = True
            continue
        if v == 0:
            out.append(" ")
        elif 1 <= v <= 26:
            ch = LETTERS[v - 1]
            out.append(ch.upper() if upper else ch)
        else:
            out.append(SYMBOLS[v])
        upper = False
    return "".join(out)


def building_lines(data: bytes) -> tuple[list[dict], list[dict]]:
    """回傳（文字行, 略過的格）。字數 byte 超出 24 的格不是文字格——
    這些格要列出來而不是默默丟掉，否則看不出模型有沒有洞。"""
    lines, skipped = [], []
    for slot in range(TEXT_BASE // SLOT_BYTES, len(data) // SLOT_BYTES):
        offset = slot * SLOT_BYTES
        count = data[offset]
        if count == 0:
            continue
        if count > SLOT_CHARS:
            skipped.append({"slot": slot, "offset": offset, "count_byte": count})
            continue
        values = slot_values(data, offset, count)
        lines.append({"slot": slot, "offset": offset, "chars": count,
                      "text": render(values)})
    return lines, skipped


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

    towns = []
    total_lines = 0
    total_skipped = 0
    with zipfile.ZipFile(zip_path) as zf:
        for town in range(TOWN_FIRST, TOWN_LAST + 1):
            root = f"TOWN{town}"
            contract, parts = payloads(read_volumes(zf, root), root)
            table = lzw_decode(parts[DOOR_TABLE_PAYLOAD])
            buildings = []
            for door in range(table[COUNT_OFFSET] + 1):
                index = text_payload_index(door)
                if index >= contract["count"]:
                    continue
                data = lzw_decode(parts[index])
                lines, skipped = building_lines(data)
                total_lines += len(lines)
                total_skipped += len(skipped)
                buildings.append({
                    "skipped_slots": skipped,
                    "door": door,
                    "x": table[DOOR_OFFSET + 2 * door],
                    "y": table[DOOR_OFFSET + 2 * door + 1],
                    "text_payload": index,
                    "payload_sha256": hashlib.sha256(data).hexdigest(),
                    "lines": lines,
                })
            towns.append({"town": town, "building_count": len(buildings),
                          "buildings": buildings})

    doc = {
        "schema": "kol-town-text-data-v1",
        "tool": "tools/extract_town_text.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWNn",
                           "text_start_offset": TEXT_BASE},
        "evidence": "RE-062",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "encoding": {
            "slot_bytes": SLOT_BYTES, "slot_chars": SLOT_CHARS,
            "count_byte": "每格的第 0 個 byte",
            "bits_per_char": 5, "bit_order": "MSB-first",
            "alphabet": {"0": "空白", "1-26": "a-z", "27": '"', "28": ",",
                         "29": ".", "30": "大寫記號", "31": "'"},
        },
        "town_count": len(towns),
        "line_count": total_lines,
        "skipped_slot_count": total_skipped,
        "skipped_note": "字數 byte 超出 24 的格：那個位置不是文字格，"
                        "多半是同一份記錄裡的商店或數值欄位。",
        "towns": towns,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"towns": len(towns), "lines": total_lines,
                      "skipped": total_skipped,
                      "buildings": sum(t["building_count"] for t in towns),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
