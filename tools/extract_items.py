#!/usr/bin/env python3
"""從 ITEM.LST 抽出戰場掉落物品表。

打贏一場戰鬥有 20% 機率掉一件東西（RE-091 的 `sub_289EE(0)`）：原版把
`item.lst` 整份載進記憶體，`rand(0x3D) − 1` 挑一筆 16 byte，複製到戰場物品格
`0x9D20` 的空槽。61 筆的第 0 筆是全零的空位。

一筆的版面與商店的商品記錄相同（RE-075），所以直接沿用 `extract_town_buildings`
的 `good()`。

  python3 tools/extract_items.py --out assets/original-derived-data/items.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from extract_town_buildings import good  # noqa: E402
from inventory_image_archives import EXPECTED_ZIP_SHA256  # noqa: E402

ENTRY = "Knights-of-legend/ITEM.LST"
RECORD_SIZE = 16
# `sub_289EE(0)` 的 `rand(0x3D)`：61 筆。
EXPECTED_RECORDS = 0x3D


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit("原版 ZIP SHA-256 不符")

    with zipfile.ZipFile(zip_path) as zf:
        body = zf.read(ENTRY)
    if len(body) != EXPECTED_RECORDS * RECORD_SIZE:
        raise SystemExit(f"ITEM.LST 是 {len(body)} byte，"
                         f"不是 {EXPECTED_RECORDS} × {RECORD_SIZE}")

    items = []
    for i in range(EXPECTED_RECORDS):
        rec = body[i * RECORD_SIZE:(i + 1) * RECORD_SIZE]
        if not any(rec):
            # 第 0 筆是全零的空位；原版擲到它就等於沒掉東西。
            items.append({"index": i, "name": "", "empty": True})
            continue
        items.append(good(rec, i))

    named = [x for x in items if x.get("name")]
    if len(named) < EXPECTED_RECORDS - 1:
        raise SystemExit(f"只有 {len(named)} 筆解出名字，預期 {EXPECTED_RECORDS - 1}")

    out = {
        "schema": "kol-battle-items-v1",
        "tool": "tools/extract_items.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_file": ENTRY,
        "evidence": "已證實：一筆 16 byte，版面與商店商品記錄相同（RE-075）；"
                    "61 筆對上 `sub_289EE(0)` 的 rand(0x3D)（RE-091）。",
        "rights": {"classification": "原版衍生 typed data；不得公開散布"},
        "record_size": RECORD_SIZE,
        "record_count": EXPECTED_RECORDS,
        "items": items,
    }
    pathlib.Path(args.out).write_text(
        json.dumps(out, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(json.dumps({"items": len(items), "named": len(named), "out": args.out},
                     ensure_ascii=False))
    return 0


sys.exit(main())
