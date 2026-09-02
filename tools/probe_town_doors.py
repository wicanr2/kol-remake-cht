#!/usr/bin/env python3
"""解出城鎮的門座標表與建築資料（RE-061）。

`TOWNn.0` 的 payload 3 被 `sub_10000:0x106ab` 讀進 DGROUP 位移 `4430h`。
`sub_10000:0x10A6D` 的迴圈說明了它的版面：

  `[0x20]`            門數（迴圈條件是 `index <= 門數 × 2`，所以實際讀到門數 + 1 筆）
  `[0x27 + 2k]`       第 k 個入城點的 X，`[0x28 + 2k]` 是 Y（`word_2DF62` 選第幾個）
  `[0x47 + 2i]`       第 i 個門的 X，`[0x48 + 2i]` 是 Y

隊伍走到某個門的座標時 `word_2FBC2` 被設成 `2i`，接著 `sub_1198E`：

  `sub_11302(2i + 3)`  把 `TWN{n}PICS[2i+3]`（192×96）畫進地圖視窗 (104,58)-(296,154)
  `sub_11703(2)`       掛上 DOOR 圖示（`KOLMPICS[2+3]` ＝ 5）
  `sub_11703(0Fh)`     掛上 U-TURN 圖示（`KOLMPICS[15+3]` ＝ 18）
  選 DOOR → `sub_11A93` 讀 `TOWNn.0` 的 payload `i + 4`（≥ 0x14 時再加 1）

該 payload 的位移 0 是建築名、位移 29 是人物名，都是「長度前綴 ＋ 末字元設高位元」。
偶數的 `TWN{n}PICS[2i+2]` 是同一棟建築的 88×64 人物肖像，畫在 (8,64)-(96,128)。

  python3 tools/probe_town_doors.py --out docs/re/town-doors.json
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

DOOR_TABLE_PAYLOAD = 3
COUNT_OFFSET = 0x20
ENTRY_OFFSET = 0x27
DOOR_OFFSET = 0x47
TEXT_BASE = 4
TEXT_SKIP = 0x14
NAME_OFFSET = 0
PERSON_OFFSET = 29
TOWN_COUNT = 8


def pascal_high_bit(data: bytes, offset: int) -> tuple[str, bool]:
    """長度前綴字串；最後一個字元的 bit 7 是結束記號。"""
    length = data[offset]
    if length == 0 or offset + 1 + length > len(data):
        return "", False
    body = bytearray(data[offset + 1:offset + 1 + length])
    terminated = bool(body[-1] & 0x80)
    body[-1] &= 0x7F
    text = body.decode("ascii", "replace")
    return text, terminated and all(0x20 <= c < 0x7F for c in body)


def text_payload_index(door: int) -> int:
    index = door + TEXT_BASE
    if index >= TEXT_SKIP:
        index += 1
    return index


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    towns = []
    with zipfile.ZipFile(zip_path) as zf:
        for town in range(TOWN_COUNT):
            root = f"TOWN{town}"
            volumes = read_volumes(zf, root)
            contract, parts = payloads(volumes, root)
            table = lzw_decode(parts[DOOR_TABLE_PAYLOAD])
            count = table[COUNT_OFFSET]

            entries = []
            k = 0
            while ENTRY_OFFSET + 2 * k + 1 < DOOR_OFFSET:
                x, y = table[ENTRY_OFFSET + 2 * k], table[ENTRY_OFFSET + 2 * k + 1]
                if (x, y) == (0, 0):
                    break
                entries.append({"index": k, "x": x, "y": y})
                k += 1

            doors = []
            # 迴圈條件是 `<=`，所以讀到的是 count + 1 筆。
            for i in range(count + 1):
                ox = DOOR_OFFSET + 2 * i
                if ox + 1 >= len(table):
                    break
                x, y = table[ox], table[ox + 1]
                index = text_payload_index(i)
                entry = {"door": i, "x": x, "y": y,
                         "picture_index": 2 * i + 3, "portrait_index": 2 * i + 2,
                         "text_payload": index}
                if index < contract["count"]:
                    body = lzw_decode(parts[index])
                    name, ok_name = pascal_high_bit(body, NAME_OFFSET)
                    person, ok_person = pascal_high_bit(body, PERSON_OFFSET)
                    entry["name"] = name
                    entry["person"] = person
                    entry["strings_well_formed"] = ok_name and ok_person
                else:
                    entry["text_payload_missing"] = True
                doors.append(entry)

            towns.append({
                "town": town,
                "archive_payload_count": contract["count"],
                "door_count_byte": count,
                "doors_read": len(doors),
                "table_bytes": len(table),
                "table_sha256": hashlib.sha256(table).hexdigest(),
                "entry_points": entries,
                "doors": doors,
            })

    receipt = {
        "schema": "kol-town-doors-v1",
        "input_name": "Knights-of-legend.zip",
        "input_sha256": digest,
        "evidence": "RE-061：sub_10000 的門座標掃描迴圈與 sub_1198E／sub_11302 的配對",
        "layout": {"count": f"payload3[0x{COUNT_OFFSET:02X}]",
                   "entry_points": f"payload3[0x{ENTRY_OFFSET:02X} + 2k]",
                   "doors": f"payload3[0x{DOOR_OFFSET:02X} + 2i]",
                   "picture": "TWN{n}PICS[2i+3]，192×96，畫在 (104,58)",
                   "portrait": "TWN{n}PICS[2i+2]，88×64，畫在 (8,64)",
                   "text": "TOWNn.0 payload i+4（≥0x14 再加 1）；名稱在 0、人物在 29"},
        "towns": towns,
    }
    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    for t in towns:
        bad = [d for d in t["doors"] if not d.get("strings_well_formed")]
        print("TOWN%d：門數 byte %2d、讀到 %2d 筆、入城點 %2d 個、字串異常 %d 筆"
              % (t["town"], t["door_count_byte"], t["doors_read"],
                 len(t["entry_points"]), len(bad)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
