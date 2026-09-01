#!/usr/bin/env python3
"""從固定原版輸入抽出起始武器樣板與每個職業的武器 selector。

建立角色時，原版把起始武器的 16-byte 樣板抄進角色槽位。樣板不是從 `ITEM.LST` 取的
（雖然同樣的位元組也出現在那裡），而是來自 `BOOKDATA` 的第 3 個項目——`_main` 在
`0x10257` 把它載進 dgroup+0x79EE，`sub_11820` 再以 `[bx+79EEh]` 定址。這一點由
反組譯讀出、再以四筆實際寫入的槽位逐 byte 命中確認（RE-050）。

規則：
- 職業 metadata（`BOOKDATA` 第 2 項之外的 dgroup+0x108，39×5 bytes）第一個 byte 的
  低 nibble 是主武器 selector、高 nibble 是副武器 selector（0 表示沒有副武器）。
- 武器表記錄索引 ＝ selector − 1，每筆 16 bytes。

  python3 tools/extract_weapon_templates.py --zip kol.zip --repo . \\
      --out assets/original-derived-data/weapon-templates.json
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

from inventory_image_archives import EXPECTED_ZIP_SHA256, logical_archive_contract  # noqa: E402
from probe_image_lzw import decode as lzw_decode  # noqa: E402
from probe_kolbook_profession_data import ROOT  # noqa: E402

ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")
EXE_NAME = "Knights-of-legend/KOLBOOK.EXE"
EXPECTED_EXE_SHA256 = \
    "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
WEAPON_ITEM_INDEX = 3          # BOOKDATA 的第 3 個項目，載到 dgroup+0x79EE
WEAPON_RECORD_BYTES = 16
WEAPON_RECORD_COUNT = 14       # 與 weapon_selector_map 的 14 個 selector 一致
DGROUP_FILE_BASE = 37952       # dgroup 偏移 0 在 EXE 內的位置（由動畫參數表定位確認）
META_OFFSET = 0x108            # 職業 metadata，39×5 bytes
META_STRIDE = 5
PROFESSION_COUNT = 39


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
        files = {i.filename: archive.read(i) for i in archive.infolist() if not i.is_dir()}
    exe = files[EXE_NAME]
    if hashlib.sha256(exe).hexdigest() != EXPECTED_EXE_SHA256:
        raise SystemExit("KOLBOOK.EXE 已改變")

    contract = logical_archive_contract(ROOT, files)
    if contract is None:
        raise SystemExit("BOOKDATA archive contract 不符")
    item = contract["items"][WEAPON_ITEM_INDEX]
    raw = lzw_decode(files[item["volume_name"]]
                     [item["offset"]:item["offset"] + item["compressed_bytes"]])
    need = WEAPON_RECORD_BYTES * WEAPON_RECORD_COUNT
    if len(raw) < need:
        raise SystemExit(f"武器表只有 {len(raw)} bytes，不足 {need}")
    records = [raw[i * WEAPON_RECORD_BYTES:(i + 1) * WEAPON_RECORD_BYTES].hex()
               for i in range(WEAPON_RECORD_COUNT)]

    professions = []
    for index in range(PROFESSION_COUNT):
        base = DGROUP_FILE_BASE + META_OFFSET + index * META_STRIDE
        meta = exe[base:base + META_STRIDE]
        primary = meta[0] & 0x0F
        secondary = meta[0] >> 4
        if not 1 <= primary <= WEAPON_RECORD_COUNT:
            raise SystemExit(f"職業 {index} 的主武器 selector {primary} 超出範圍")
        if secondary and not 1 <= secondary <= WEAPON_RECORD_COUNT:
            raise SystemExit(f"職業 {index} 的副武器 selector {secondary} 超出範圍")
        professions.append({
            "record_index": index,
            "primary_selector": primary, "primary_record": primary - 1,
            "secondary_selector": secondary,
            "secondary_record": secondary - 1 if secondary else None,
            "secondary_skills": [meta[1], meta[2]],
            "primary_skills": [meta[3], meta[4]],
        })

    doc = {
        "schema": "kol-weapon-templates-v1",
        "tool": "tools/extract_weapon_templates.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source": {"archive": ROOT, "item_index": WEAPON_ITEM_INDEX,
                   "item_sha256": hashlib.sha256(raw).hexdigest(),
                   "loaded_at": "dgroup+0x79ee",
                   "exe": EXE_NAME, "exe_sha256": EXPECTED_EXE_SHA256},
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔與執行檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "rule": "職業 metadata 第一個 byte 的低 nibble 是主武器 selector、高 nibble 是副"
                "武器 selector（0 表示沒有副武器）；武器表記錄索引 ＝ selector − 1。",
        "record_bytes": WEAPON_RECORD_BYTES, "record_count": WEAPON_RECORD_COUNT,
        "records": records,
        "professions": professions,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n")
    os.replace(tmp, out)
    print(json.dumps({"weapons": WEAPON_RECORD_COUNT, "professions": len(professions),
                      "without_secondary": sum(1 for p in professions
                                               if p["secondary_selector"] == 0),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
