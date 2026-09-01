#!/usr/bin/env python3
"""從固定原版輸入抽出角色清單所需的 typed data。

`CHARDATA` 是 16 個 768-byte 槽位。名字在槽位起點、以 NUL 結尾；所在地是位移 64 起的
固定 10 bytes，城鎮代碼是位移 144 起的 3 bytes。這三個位置不是猜的：用原版角色清單
畫面（`docs/verification/character-ui/list-characters-run1/frames/frame-065.png`）以抽出的
原版字型讀回的 14 個名字，在 `CHARDATA` 內逐一唯一命中槽位起點；所在地與城鎮代碼再由
畫面上的 `<10 字所在地> <3 字城鎮>` 逐字比對確認。

  python3 tools/extract_character_roster.py --zip kol.zip --repo . \\
      --out assets/original-derived-data/character-roster.json
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
CHARDATA_NAME = "Knights-of-legend/CHARDATA"
EXPECTED_CHARDATA_SHA256 = \
    "307ae9e2735b9f1d3d8008bcf018ed9125820af4e1ffadf078bc918be16f394e"
ALLOWED_OUTPUT_PARENT = pathlib.PurePosixPath("assets/original-derived-data")

SLOT_BYTES = 768
SLOT_COUNT = 16
NAME_OFFSET, NAME_MAX = 0, 24
PLACE_OFFSET, PLACE_BYTES = 64, 10
TOWN_OFFSET, TOWN_BYTES = 144, 3
# 畫面上的 14 個名字，用來驗證槽位起點；由 tools/read_text_from_frame.py 讀回。
EXPECTED_NAMES = ["Theodoric", "Poontang", "Olear", "Moraine", "Fighting Bull",
                  "Kenneth", "Jax", "Sharakan", "Warder", "Shyla", "Aventhalas",
                  "One tooth", "Mula", "ddduhh"]


def cstring(raw: bytes) -> str:
    end = raw.find(b"\x00")
    return raw[:end if end >= 0 else len(raw)].decode("ascii")


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
        data = archive.read(CHARDATA_NAME)
    if hashlib.sha256(data).hexdigest() != EXPECTED_CHARDATA_SHA256:
        raise SystemExit("CHARDATA 已改變")
    if len(data) != SLOT_BYTES * SLOT_COUNT:
        raise SystemExit(f"CHARDATA 大小為 {len(data)}，不是 {SLOT_BYTES * SLOT_COUNT}")

    slots = []
    for index in range(SLOT_COUNT):
        base = index * SLOT_BYTES
        name = cstring(data[base + NAME_OFFSET:base + NAME_OFFSET + NAME_MAX])
        place = data[base + PLACE_OFFSET:base + PLACE_OFFSET + PLACE_BYTES]
        town = data[base + TOWN_OFFSET:base + TOWN_OFFSET + TOWN_BYTES]
        slots.append({
            "slot": index, "name": name,
            "place": place.decode("ascii"), "town": town.decode("ascii"),
            "occupied": name != "",
        })

    occupied = [s["name"] for s in slots if s["occupied"]]
    if occupied != EXPECTED_NAMES:
        raise SystemExit(f"槽位名字與原版畫面讀回的清單不符：{occupied}")

    doc = {
        "schema": "kol-character-roster-v1",
        "tool": "tools/extract_character_roster.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_file": {"name": CHARDATA_NAME, "sha256": EXPECTED_CHARDATA_SHA256},
        "evidence": "docs/verification/character-ui/list-characters-run{1,2}",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {"slot_bytes": SLOT_BYTES, "slot_count": SLOT_COUNT,
                   "name_offset": NAME_OFFSET, "place_offset": PLACE_OFFSET,
                   "place_bytes": PLACE_BYTES, "town_offset": TOWN_OFFSET,
                   "town_bytes": TOWN_BYTES},
        "location_rule": "畫面上的所在地欄是「10 byte 所在地 + 一個空白 + 3 byte 城鎮代碼」，"
                         "共 14 個字元；所在地欄不以 NUL 結尾，取固定長度。",
        "slots": slots,
    }
    out.parent.mkdir(parents=True, exist_ok=True)
    tmp = out.with_suffix(out.suffix + f".tmp-{os.getpid()}")
    tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=2) + "\n")
    os.replace(tmp, out)
    print(json.dumps({"slots": SLOT_COUNT, "occupied": len(occupied),
                      "out": str(out)}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
