#!/usr/bin/env python3
"""SPEC 用的城鎮建築 typed data 抽取器（RE-061）。

把八座城鎮的門座標、入城點與建築資料轉成 runtime 可讀的 JSON。版面由 KOLTOWN 的
`sub_10000` 掃門迴圈與 `sub_1198E`／`sub_11302` 讀出，證據見 `docs/re/town-doors.json`。

  python3 tools/extract_town_buildings.py --zip /in/kol.zip \\
      --out assets/original-derived-data/town-buildings.json

輸出根目錄必須位於 `assets/original-derived-data/`；runtime 只讀這份 JSON，
不在執行期解析原版 archive。
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
from probe_town_doors import (COUNT_OFFSET, DOOR_OFFSET, DOOR_TABLE_PAYLOAD,  # noqa: E402
                              ENTRY_OFFSET, NAME_OFFSET, PERSON_OFFSET,
                              pascal_high_bit, text_payload_index)

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
TOWN_COUNT = 8
# TOWN0 是野外地圖，沒有城鎮建築；它的 payload 3 門數是 0。
TOWN_FIRST = 1
GRID_SIDE = 64
# payload 3 的前 32 個 byte 是「磚塊編號 → 不可通行」的位元圖，一個 bit 一個磚塊。
# `sub_10000:0x1082D` 查的是 `payload3[tile >> 3] & (1 << (tile & 7))`（RE-065）。
BLOCK_BITMAP_BYTES = 32
# 建築 payload 的位移 0x3C／0x3D 是圖示列的旗標（`byte_2FC00`／`byte_2FC01`）。
# `sub_11A93:0x120B8` 起逐位元決定要推哪些圖示（RE-067）。
ICON_FLAGS = 0x3C
# 位移 0x38／0x39 是價錢，0x3A／0x3B 是建築編號，兩組都是高位元組在前。
# 客棧（旗標 bit3）住一晚的價錢是 `payload[0x38] << 8 | payload[0x39]`
# （`sub_14DB1:0x14E38`）；修道院（bit6）治療則只用 `payload[0x39]`，
# 每一點傷勢乘一次（`sub_11A93:0x126A3`）。住宿存檔時把 0x3A／0x3B 寫進
# 角色記錄位移 0x2E／0x2F（`sub_14DB1:0x1505B`）。
PRICE_OFFSET = 0x38
RECORD_ID_OFFSET = 0x3A


def be16(body: bytes, offset: int) -> int:
    return body[offset] << 8 | body[offset + 1]


def blocked_tiles(table: bytes) -> list[int]:
    return [t for t in range(BLOCK_BITMAP_BYTES * 8)
            if table[t >> 3] & (1 << (t & 7))]


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
        for town in range(TOWN_FIRST, TOWN_COUNT):
            root = f"TOWN{town}"
            contract, parts = payloads(read_volumes(zf, root), root)
            table = lzw_decode(parts[DOOR_TABLE_PAYLOAD])
            count = table[COUNT_OFFSET]

            entries = []
            k = 0
            while ENTRY_OFFSET + 2 * k + 1 < DOOR_OFFSET:
                x, y = table[ENTRY_OFFSET + 2 * k], table[ENTRY_OFFSET + 2 * k + 1]
                if (x, y) == (0, 0):
                    break
                entries.append({"x": x, "y": y})
                k += 1

            doors = []
            for i in range(count + 1):          # 原版迴圈是 `<=`，實際讀到 count + 1 筆
                ox = DOOR_OFFSET + 2 * i
                x, y = table[ox], table[ox + 1]
                if not (0 <= x < GRID_SIDE and 0 <= y < GRID_SIDE):
                    raise SystemExit(f"{root} 門 {i} 的座標 ({x},{y}) 超出 64×64")
                index = text_payload_index(i)
                door = {"index": i, "x": x, "y": y,
                        "picture": 2 * i + 3, "portrait": 2 * i + 2}
                if index < contract["count"]:
                    body = lzw_decode(parts[index])
                    door["icon_flags"] = [body[ICON_FLAGS], body[ICON_FLAGS + 1]]
                    name, ok_name = pascal_high_bit(body, NAME_OFFSET)
                    person, ok_person = pascal_high_bit(body, PERSON_OFFSET)
                    if not (ok_name and ok_person):
                        raise SystemExit(f"{root} 門 {i} 的字串不合版面")
                    door["name"] = name
                    door["person"] = person
                    door["price"] = be16(body, PRICE_OFFSET)
                    door["heal_rate"] = body[PRICE_OFFSET + 1]
                    door["record_id"] = be16(body, RECORD_ID_OFFSET)
                else:
                    # TOWN7 的門 21 有圖與肖像，卻沒有對應的文字 payload。
                    door["text_missing"] = True
                doors.append(door)

            towns.append({
                "town": town, "archive": f"TWN{town}PICS",
                "map_archive": f"TOWN{town}",
                "door_count_byte": count,
                "blocked_bitmap": table[:BLOCK_BITMAP_BYTES].hex(),
                "blocked_tiles": blocked_tiles(table),
                "table_sha256": hashlib.sha256(table).hexdigest(),
                "entry_points": entries, "doors": doors,
            })

    doc = {
        "schema": "kol-town-buildings-data-v1",
        "tool": "tools/extract_town_buildings.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWNn", "payload": DOOR_TABLE_PAYLOAD},
        "evidence": "RE-061；docs/verification/character-ui/town-enter-building-run1",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "count": f"payload{DOOR_TABLE_PAYLOAD}[0x{COUNT_OFFSET:02X}]",
            "entry_points": f"payload{DOOR_TABLE_PAYLOAD}[0x{ENTRY_OFFSET:02X} + 2k]",
            "doors": f"payload{DOOR_TABLE_PAYLOAD}[0x{DOOR_OFFSET:02X} + 2i]",
            "blocked": f"payload{DOOR_TABLE_PAYLOAD}[0x00…0x1F]：磚塊編號的不可通行位元圖，"
                       "查 `bitmap[tile >> 3] & (1 << (tile & 7))`",
            "icon_flags": f"建築 payload[0x{ICON_FLAGS:02X}／0x{ICON_FLAGS + 1:02X}]："
                          "圖示列旗標；bit0 或 flags[1] 的 bit2 給 LISTEN＋MOUTH、"
                          "bit1/4/5 給買賣、bit2 給 id 11、bit3 給 REST＋COMPANION、"
                          "bit6 給 REST",
            "price": f"建築 payload[0x{PRICE_OFFSET:02X}／0x{PRICE_OFFSET + 1:02X}]："
                     "高位元組在前的價錢。bit3 的客棧是住一晚的總價（0 代表免費），"
                     "bit6 的修道院只用低位元組，當每一點傷勢的價錢。",
            "record_id": f"建築 payload[0x{RECORD_ID_OFFSET:02X}／0x{RECORD_ID_OFFSET + 1:02X}]："
                         "住宿存檔時寫進角色記錄位移 0x2E／0x2F 的建築編號。",
            "picture": "TWN{n}PICS[2i+3]，192×96，畫在地圖視窗 (104,58)",
            "portrait": "TWN{n}PICS[2i+2]，88×64，畫在 (8,64)",
            "text": "TOWNn.0 payload i+4（索引 ≥ 0x14 再加 1）；名稱在 0、人物在 29，"
                    "都是長度前綴且末字元設高位元",
        },
        "door_icons": {"door": {"id": 2, "kolmpics": 5},
                       "u_turn": {"id": 15, "kolmpics": 18},
                       "rule": "圖示列畫的是 KOLMPICS[id + 3]"},
        "towns": towns,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    os.replace(tmp, out)
    print(json.dumps({"towns": len(towns),
                      "doors": sum(len(t["doors"]) for t in towns),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
