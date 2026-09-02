#!/usr/bin/env python3
"""城鎮的話題系統 typed data（RE-067）。

一棟建築的六個話題槽在它自己的 payload 位移 0x17…0x1C，每格一個 byte：
低 6 bit 是話題碼、高 2 bit 是類別。話題碼對應 `TOWNn.0` payload 20 的關鍵字表。

payload 20 的版面（由 `sub_11A93:0x122CE` 起的迴圈讀出）：

  0x000 + 4i    64 筆關鍵字：前三個 byte 是 mode 1 的 5-bit 打包字串（最多四個字元），
                第四個 byte 是碼（低 6 bit 是話題碼、高 2 bit 是類別）
  0x100 + 8i    16 筆人名，同樣是 mode 1 打包
  0x1C0 + 碼    這個話題該去問誰（人名索引；0xFF 代表沒有人知道）

`sub_1B94F` 的 mode 1 與 mode 0 不同：字數在第 0 個 byte 的**高 4 bit**，
位元流從同一個 byte 的 bit 3 開始，而且跳脫狀態一開始就是打開的（首字大寫）。

  python3 tools/extract_town_keywords.py --out assets/original-derived-data/town-keywords.json
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

from extract_town_text import render  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402
from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402
from probe_town_doors import (COUNT_OFFSET, DOOR_TABLE_PAYLOAD,  # noqa: E402
                              text_payload_index)

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
KEYWORD_PAYLOAD = 20
KEYWORD_BYTES, KEYWORD_COUNT = 4, 64
NAME_TABLE, NAME_BYTES, NAME_COUNT = 0x100, 8, 16
ASKWHO_TABLE = 0x1C0
TOPIC_SLOTS, TOPIC_COUNT = 0x17, 6
TOWN_FIRST, TOWN_LAST = 1, 7
NO_ONE = 0xFF


def decode_mode1(data: bytes, offset: int, span: int) -> str:
    """`sub_1B94F` 的 mode 1。字數在高 4 bit，位元流從 bit 3 起。"""
    count = data[offset] >> 4
    # 表尾是 0xFF 填充：高 4 bit 一律 15，而 span 個 byte 最多只裝得下
    # (span * 8 - 4) // 5 個字元，所以字數超過上限就是填充，不是字串。
    if count == 0 or count > (span * 8 - 4) // 5:
        return ""
    bits = [(data[offset] >> s) & 1 for s in (3, 2, 1, 0)]
    for b in range(1, span):
        byte = data[offset + b]
        bits.extend((byte >> s) & 1 for s in (7, 6, 5, 4, 3, 2, 1, 0))
    values = []
    for i in range(count):
        value = 0
        for bit in bits[i * 5:(i + 1) * 5]:
            value = (value << 1) | bit
        values.append(value)
    text, unknown = render(values)
    if unknown:
        raise SystemExit(f"位移 0x{offset:x} 出現未定的跳脫序列 {unknown}")
    # 原版在比對之前把第一個字元轉小寫（`byte_30C20 += 20h`）。
    return text[:1].lower() + text[1:]


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
    with zipfile.ZipFile(zip_path) as zf:
        for town in range(TOWN_FIRST, TOWN_LAST + 1):
            root = f"TOWN{town}"
            contract, parts = payloads(read_volumes(zf, root), root)
            table = lzw_decode(parts[KEYWORD_PAYLOAD])
            doors = lzw_decode(parts[DOOR_TABLE_PAYLOAD])

            keywords = []
            for i in range(KEYWORD_COUNT):
                off = i * KEYWORD_BYTES
                word = decode_mode1(table, off, KEYWORD_BYTES - 1)
                if not word:
                    break
                code = table[off + 3]
                keywords.append({"prefix": word, "code": code,
                                 "topic": code & 0x3F, "category": code >> 6})
            names = []
            for i in range(NAME_COUNT):
                off = NAME_TABLE + i * NAME_BYTES
                names.append(decode_mode1(table, off, NAME_BYTES))
            ask_who = {}
            for k in keywords:
                who = table[ASKWHO_TABLE + k["topic"]]
                if who != NO_ONE:
                    ask_who[str(k["topic"])] = who

            buildings = []
            for door in range(doors[COUNT_OFFSET] + 1):
                index = text_payload_index(door)
                if index >= contract["count"]:
                    continue
                body = lzw_decode(parts[index])
                slots = []
                for slot in range(TOPIC_COUNT):
                    raw = body[TOPIC_SLOTS + slot]
                    slots.append({"slot": slot, "raw": raw,
                                  "topic": raw & 0x3F, "category": raw >> 6})
                buildings.append({"door": door, "slots": slots})

            towns.append({
                "town": town,
                "table_sha256": hashlib.sha256(table).hexdigest(),
                "keyword_count": len(keywords),
                "keywords": keywords, "names": names,
                "ask_who": ask_who, "buildings": buildings,
            })

    doc = {
        "schema": "kol-town-keywords-data-v1",
        "tool": "tools/extract_town_keywords.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWNn",
                           "keyword_payload": KEYWORD_PAYLOAD},
        "evidence": "RE-067",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "keyword": f"payload{KEYWORD_PAYLOAD}[4i]：三個 byte 的 mode 1 打包字串"
                       "（最多四個字元，是玩家輸入的字首）＋一個碼 byte",
            "code": "低 6 bit 是話題碼、高 2 bit 是類別（LISTEN 的開場白 KOLDATA 0x9A + 類別）",
            "names": f"payload{KEYWORD_PAYLOAD}[0x{NAME_TABLE:X} + 8i]：16 個人名",
            "ask_who": f"payload{KEYWORD_PAYLOAD}[0x{ASKWHO_TABLE:X} + 話題碼]："
                       "沒有人回答時要指向誰；0xFF 代表沒有人知道",
            "slots": f"建築 payload[0x{TOPIC_SLOTS:X} + n]：六個話題槽，"
                     "命中第 n 槽就顯示訊息 20 + n",
        },
        "behaviour": {
            "listen": "隨機挑一個話題槽（`rand(6)`），空的或講過的就重來；"
                      "先講 KOLDATA 記錄 0x9A + 類別 的開場白，再顯示訊息 20 + 槽號。"
                      "講過的槽記在每棟建築進門時清零的旗標裡。",
            "mouth": "玩家輸入一句話，全部轉小寫；逐一比對關鍵字字首（以空白分詞），"
                     "命中就拿它的話題碼去比對這棟建築的六個槽；對上顯示訊息 20 + 槽號，"
                     "沒對上就用 KOLDATA 47／48／0x9E + 類別 組出「我聽過某某談起這件事」。"
                     "整張表都沒命中就顯示訊息 19。",
        },
        "towns": towns,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"towns": len(towns),
                      "keywords": sum(t["keyword_count"] for t in towns),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
