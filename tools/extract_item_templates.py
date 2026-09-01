#!/usr/bin/env python3
"""從固定原版輸入抽出 `ITEM.LST` 的物品樣板，以及已觀測到的職業起始武器對應。

建立角色時，原版把起始武器的 `ITEM.LST` 記錄整筆抄進角色槽位。`ITEM.LST` 是 61 筆
16-byte 記錄，這一點由檔案長度與命中位置整除確認。

**職業 → 物品索引的通則尚未反追**：`weapon_selector_map` 給的是 selector → 手冊縮寫，
而 selector 與 `ITEM.LST` 索引之間不是算術關係（GS 4→7、HL 8→11、BR 9→5、LB 3→25），
EXE 內也找不到對應的位元組表。因此這裡只登記**實際觀測到**的職業，其餘留空，
讓 writer 對未知職業失敗而不是猜。

  python3 tools/extract_item_templates.py --zip kol.zip --repo . \\
      --out assets/original-derived-data/item-templates.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import sys
import zipfile

EXPECTED_ZIP_SHA256 = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
ITEM_NAME = "Knights-of-legend/ITEM.LST"
RECORD_BYTES = 16
ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")

# 由原版實際寫入的角色槽位觀測到的對應：把寫進槽位的武器樣板前綴拿回 ITEM.LST 比對，
# 每一筆都唯一命中。收據見 docs/verification/chardata-write.json。
OBSERVED = [
    {"profession_record": 0, "menu_name": "BARBARIAN",
     "primary_weapon": "GS", "primary_item": 7,
     "secondary_weapon": "HL", "secondary_item": 11,
     "evidence": "accept-and-save 情境寫入槽位 14 的實際 bytes"},
    {"profession_record": 1, "menu_name": "RANGER",
     "primary_weapon": "BR", "primary_item": 5,
     "secondary_weapon": "LB", "secondary_item": 25,
     "evidence": "accept-and-save-ranger 情境寫入槽位 14 的實際 bytes"},
]


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", required=True)
    p.add_argument("--repo", default=".")
    p.add_argument("--out", required=True)
    args = p.parse_args()

    repo = pathlib.Path(args.repo).resolve()
    out = pathlib.Path(args.out).resolve()
    try:
        relative = out.relative_to(repo)
    except ValueError:
        raise SystemExit("輸出必須位於 repo 內")
    if ALLOWED_OUTPUT_PARENT != pathlib.PurePosixPath(relative.as_posix()).parent:
        raise SystemExit(f"輸出必須位於 {ALLOWED_OUTPUT_PARENT}/ 底下")

    zip_path = pathlib.Path(args.zip)
    digest = hashlib.sha256(zip_path.read_bytes()).hexdigest()
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"非預期 ZIP SHA-256：{digest}")
    with zipfile.ZipFile(zip_path) as archive:
        raw = archive.read(ITEM_NAME)
    if len(raw) % RECORD_BYTES:
        raise SystemExit(f"{ITEM_NAME} 長度 {len(raw)} 不是 {RECORD_BYTES} 的倍數")
    count = len(raw) // RECORD_BYTES
    records = [raw[i * RECORD_BYTES:(i + 1) * RECORD_BYTES].hex() for i in range(count)]
    for entry in OBSERVED:
        for key in ("primary_item", "secondary_item"):
            if not 0 <= entry[key] < count:
                raise SystemExit(f"{entry['menu_name']} 的 {key} 超出範圍")

    doc = {
        "schema": "kol-item-templates-v1",
        "tool": "tools/extract_item_templates.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_file": {"name": ITEM_NAME, "sha256": hashlib.sha256(raw).hexdigest(),
                        "bytes": len(raw)},
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "record_bytes": RECORD_BYTES, "record_count": count,
        "records": records,
        "observed_profession_weapons": OBSERVED,
        "limitation": "職業 → ITEM.LST 索引的通則未知；只有觀測到的職業可以寫出武器樣板。",
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n")
    os.replace(tmp, out)
    print(json.dumps({"records": count, "observed": len(OBSERVED), "out": str(out)},
                     ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
