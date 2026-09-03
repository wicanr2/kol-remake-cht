#!/usr/bin/env python3
"""24 個任務地點的 typed data（RE-079）。

`sub_29004:0x29484` 在世界座標變動之後比對 `payload3[0x323 + i]`／`[0x33B + i]`，
命中就呼叫 `sub_2B0FF(i)`。那一支先看隊上有沒有人帶著這個任務
（角色 `[0x1E2 + i / 4]` 除以 `[0x7C + (i & 3)]` 之後低兩位是 1），
再讀 `TOWN0` 的 payload `0x25 + i`：第 0 個 byte 是怪物編號、第 1 個是等級、
第 2 個是數量（`0x2B19B`…`0x2B1AD`）。

  python3 tools/extract_quest_sites.py --out assets/original-derived-data/quest-sites.json
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

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
TABLE_PAYLOAD, SITE_PAYLOAD_BASE, SITE_COUNT = 3, 0x25, 24
SITE_X, SITE_Y = 0x323, 0x33B
FIRST_MONSTER_PAYLOAD, MONSTER_COUNT = 5, 32
# 任務旗標：一個 byte 塞四個 2-bit 欄位，除數表在 KOLTOWN 的 DGROUP+0x7C。
QUEST_FIELD_BASE, QUEST_DIVISORS = 0x1E2, (1, 4, 16, 64)


def monster_name(record: bytes) -> str:
    length = record[0]
    if not 0 < length <= 24:
        raise SystemExit(f"怪物名稱長度不合理：{length}")
    return record[1:1 + length].decode("ascii")


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
        sites_raw = [lzw_decode(parts[SITE_PAYLOAD_BASE + i]) for i in range(SITE_COUNT)]
        names = [monster_name(lzw_decode(parts[FIRST_MONSTER_PAYLOAD + i]))
                 for i in range(MONSTER_COUNT)]

    sites = []
    for i, raw in enumerate(sites_raw):
        monster = raw[0]
        if monster >= MONSTER_COUNT:
            raise SystemExit(f"任務地點 {i} 的怪物編號 {monster} 超出範圍")
        sites.append({
            "index": i,
            "x": table[SITE_X + i], "y": table[SITE_Y + i],
            "monster": monster, "monster_name": names[monster],
            "level": raw[1], "count": raw[2],
            "quest_byte": QUEST_FIELD_BASE + i // 4,
            "quest_divisor": QUEST_DIVISORS[i & 3],
            "payload": SITE_PAYLOAD_BASE + i,
            "payload_sha256": hashlib.sha256(raw).hexdigest(),
        })

    doc = {
        "schema": "kol-quest-site-data-v1",
        "tool": "tools/extract_quest_sites.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_archive": {"root": "Knights-of-legend/TOWN0",
                           "table_payload": TABLE_PAYLOAD,
                           "site_payloads": [SITE_PAYLOAD_BASE,
                                             SITE_PAYLOAD_BASE + SITE_COUNT - 1]},
        "evidence": "RE-079",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {
            "coordinates": f"payload{TABLE_PAYLOAD}[0x{SITE_X:X} + i]／[0x{SITE_Y:X} + i]",
            "site_record": f"payload {SITE_PAYLOAD_BASE} + i：怪物、等級、數量",
            "quest_flag": "角色 [0x1E2 + i / 4] 除以 (1, 4, 16, 64)[i & 3] 之後低兩位是 1",
        },
        "sites": sites,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"sites": len(sites), "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
