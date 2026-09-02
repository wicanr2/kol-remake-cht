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
# 一則訊息固定 4 格 ＝ 64 byte；原版對話框一次顯示一或兩則（最多八行）。
SLOTS_PER_MESSAGE = 4
LETTERS = "abcdefghijklmnopqrstuvwxyz"
# 30 是跳脫記號：接字母就轉大寫，接符號就換一套符號。
# 依據是全七城 191,278 個字元的上下文統計（RE-062）：
# `30 27` 有 1,721/1,796 次落在整格開頭且後面接字母——那是開頭引號；
# `31` 後面幾乎只有 s／t／v／m／l（`'s`、`'t`、`'ve`、`'m`、`'ll`）。
SHIFT = 30
SYMBOLS = {27: "!", 28: ",", 29: ".", 31: "'"}
ESCAPED = {27: '"', 31: "?"}
TOWN_FIRST, TOWN_LAST = 1, 7


def slot_values(data: bytes, offset: int, count: int) -> list[int]:
    out = []
    for i in range(count):
        bit = (offset + 1) * 8 + i * 5
        j = bit >> 3
        word = (data[j] << 8) | (data[j + 1] if j + 1 < len(data) else 0)
        out.append((word >> (11 - (bit & 7))) & 0x1F)
    return out


def render(values: list[int]) -> tuple[str, list[int]]:
    """回傳（文字, 未定的跳脫序列）。未定的照原樣留一個記號，不猜。"""
    out, unknown = [], []
    escaped = False
    for v in values:
        if v == SHIFT and not escaped:
            escaped = True
            continue
        if v == 0:
            out.append(" ")
        elif 1 <= v <= 26:
            ch = LETTERS[v - 1]
            out.append(ch.upper() if escaped else ch)
        elif escaped:
            if v in ESCAPED:
                out.append(ESCAPED[v])
            else:
                out.append(f"<esc{v}>")
                unknown.append(v)
        else:
            out.append(SYMBOLS[v])
        escaped = False
    return "".join(out), unknown


def display_text(text: str) -> str:
    """資料 → 畫面。由原版畫格 072 的八行逐格讀回比對出來的兩條規則：
    句號後補兩個空白並把下一個字母轉大寫；逗號後補一個空白。
    （資料裡已有的空白照留，所以「,␣」會變成「,␣␣」——與原版一致。）"""
    out = []
    capitalise = False
    for ch in text:
        if capitalise and ch.isalpha():
            out.append(ch.upper())
            capitalise = False
            continue
        capitalise = False
        out.append(ch)
        if ch == ".":
            out.append("  ")
            capitalise = True
        elif ch == ",":
            out.append(" ")
    return "".join(out).rstrip()


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
        text, unknown = render(values)
        entry = {"slot": slot, "offset": offset, "chars": count,
                 "text": text, "display": display_text(text)}
        if unknown:
            entry["unknown_escapes"] = unknown
        lines.append(entry)
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
                messages = []
                for start in range(0, len(data) // SLOT_BYTES, SLOTS_PER_MESSAGE):
                    group = [l for l in lines
                             if start <= l["slot"] - TEXT_BASE // SLOT_BYTES
                             < start + SLOTS_PER_MESSAGE]
                    if group and any(l["text"].strip() for l in group):
                        messages.append({"index": start // SLOTS_PER_MESSAGE,
                                         "lines": [l["text"] for l in group],
                                         "display": [l["display"] for l in group]})
                buildings.append({
                    "messages": messages,
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
        "message_grouping": {
            "slots_per_message": SLOTS_PER_MESSAGE,
            "bytes_per_message": SLOT_BYTES * SLOTS_PER_MESSAGE,
            "note": "訊息 0 與 1 合起來是進門的敘述（八行）；其後是人物的多種反應。"
                    "畫面上第五行之後的問句來自 KOLDATA（例如記錄 45 "
                    '「"What can I do for you?"」）。',
        },
        "encoding": {
            "slot_bytes": SLOT_BYTES, "slot_chars": SLOT_CHARS,
            "count_byte": "每格的第 0 個 byte",
            "bits_per_char": 5, "bit_order": "MSB-first",
            "alphabet": {"0": "空白", "1-26": "a-z", "27": "!", "28": ",",
                         "29": ".", "31": "'"},
            "display_rule": "句號後補兩個空白並把下一個字母轉大寫；逗號後補一個空白。"
                            "引號前後的空白還沒定案（見 RE-062）。",
            "escape": {"code": SHIFT, "letter": "轉大寫",
                       "27": '"（開頭引號）', "31": "?"},
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
