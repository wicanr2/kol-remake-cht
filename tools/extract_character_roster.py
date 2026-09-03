#!/usr/bin/env python3
"""從固定原版輸入抽出角色清單所需的 typed data。

`CHARDATA` 是 16 個 768-byte 槽位。名字在槽位起點、以 NUL 結尾；所在地是位移 64 起的
固定 10 bytes，城鎮代碼是位移 144 起的 3 bytes。這三個位置不是猜的：用原版角色清單
畫面（`docs/verification/character-ui/list-characters-run1/frames/frame-065.png`）以抽出的
原版字型讀回的 14 個名字，在 `CHARDATA` 內逐一唯一命中槽位起點；所在地與城鎮代碼再由
畫面上的 `<10 字所在地> <3 字城鎮>` 逐字比對確認。

頭像取自同一份存檔的 `CHARPICS`（16 × 760 byte，與模板庫 `CHARPICS.0` 同格式）。
16 個槽位的點陣圖逐一與抽出的 80 張模板 PNG 比對，**全部命中**，因此這裡記的是
模板圖號而不是另存一份點陣圖；有任何一格對不上模板就會失敗離開，不會默默略過。

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

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from png_io import read_png  # noqa: E402
from probe_extract_comb_png import PALETTE  # noqa: E402
from probe_extract_fixed_charpics_png import decode_record  # noqa: E402

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
# 位移 0x1A 是性別、0x1B 是傷勢。傷勢非 0 時旅店的招呼語換成
# 「"You are hurt! You should go to the abbey."」（KOLDATA 2Bh／2Ch，`sub_11A93:0x11F47`）。
FEMALE_OFFSET, WOUNDS_OFFSET = 0x1A, 0x1B
# 位移 0x2A–0x2C 是金幣，高位元組在前的 24-bit。`sub_14CCF` 讀進來、
# `sub_14D3A` 寫回去，客棧住宿與修道院治療花的就是這個欄位（RE-073）。
GOLD_OFFSET, GOLD_BYTES = 0x2A, 3
# 位移 0x2E／0x2F 是角色留宿的建築編號，來自建築 payload 的 0x3A／0x3B。
BUILDING_OFFSET = 0x2E
# 存檔的頭像檔：一格 760 byte，與模板庫同一種 packed-nibble 版面。
CHARPICS_NAME = "Knights-of-legend/CHARPICS"
PORTRAIT_BYTES = 760
TEMPLATE_COUNT = 80
TEMPLATE_DIR = "assets/original-derived-png/ega/CHARPICS"
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
        portraits = archive.read(CHARPICS_NAME)
    if hashlib.sha256(data).hexdigest() != EXPECTED_CHARDATA_SHA256:
        raise SystemExit("CHARDATA 已改變")
    if len(data) != SLOT_BYTES * SLOT_COUNT:
        raise SystemExit(f"CHARDATA 大小為 {len(data)}，不是 {SLOT_BYTES * SLOT_COUNT}")

    if len(portraits) != PORTRAIT_BYTES * SLOT_COUNT:
        raise SystemExit(f"CHARPICS 大小為 {len(portraits)}，"
                         f"不是 {PORTRAIT_BYTES * SLOT_COUNT}")
    templates = {}
    for index in range(TEMPLATE_COUNT):
        png = repo / TEMPLATE_DIR / f"charpics-{index:03d}.png"
        templates[read_png(png).rgb()] = index

    slots = []
    for index in range(SLOT_COUNT):
        base = index * SLOT_BYTES
        name = cstring(data[base + NAME_OFFSET:base + NAME_OFFSET + NAME_MAX])
        place = data[base + PLACE_OFFSET:base + PLACE_OFFSET + PLACE_BYTES]
        town = data[base + TOWN_OFFSET:base + TOWN_OFFSET + TOWN_BYTES]
        record = portraits[index * PORTRAIT_BYTES:(index + 1) * PORTRAIT_BYTES]
        rgb = b"".join(bytes(PALETTE[v]) for v in decode_record(record))
        portrait = templates.get(rgb)
        if portrait is None:
            raise SystemExit(f"槽位 {index} 的頭像不是任何一張模板；"
                             "頭像被編輯過就不能只記模板圖號")
        slots.append({
            "slot": index, "name": name,
            "place": place.decode("ascii"), "town": town.decode("ascii"),
            "occupied": name != "", "portrait": portrait,
            "female": data[base + FEMALE_OFFSET] != 0,
            "wounds": data[base + WOUNDS_OFFSET],
            "gold": int.from_bytes(
                data[base + GOLD_OFFSET:base + GOLD_OFFSET + GOLD_BYTES], "big"),
            "building": int.from_bytes(
                data[base + BUILDING_OFFSET:base + BUILDING_OFFSET + 2], "big"),
        })

    occupied = [s["name"] for s in slots if s["occupied"]]
    if occupied != EXPECTED_NAMES:
        raise SystemExit(f"槽位名字與原版畫面讀回的清單不符：{occupied}")

    doc = {
        "schema": "kol-character-roster-v1",
        "tool": "tools/extract_character_roster.py",
        "source_zip": {"name": zip_path.name, "sha256": digest},
        "source_file": {"name": CHARDATA_NAME, "sha256": EXPECTED_CHARDATA_SHA256},
        "portrait_file": {"name": CHARPICS_NAME,
                          "sha256": hashlib.sha256(portraits).hexdigest(),
                          "record_bytes": PORTRAIT_BYTES,
                          "rule": "每格 760 byte 解成 40×38，與 CHARPICS.0 的 80 張模板"
                                  "逐像素比對；記的是命中的模板圖號"},
        "evidence": "docs/verification/character-ui/list-characters-run{1,2}",
        "rights": {
            "classification": "private-repo 可保存；不得公開散布",
            "note": "衍生自原版資料檔，著作權仍屬原權利人；公開前須重新權利審查。",
        },
        "layout": {"slot_bytes": SLOT_BYTES, "slot_count": SLOT_COUNT,
                   "name_offset": NAME_OFFSET, "place_offset": PLACE_OFFSET,
                   "place_bytes": PLACE_BYTES, "town_offset": TOWN_OFFSET,
                   "town_bytes": TOWN_BYTES,
                   "female_offset": FEMALE_OFFSET, "wounds_offset": WOUNDS_OFFSET,
                   "gold_offset": GOLD_OFFSET, "gold_bytes": GOLD_BYTES,
                   "gold_rule": "高位元組在前的 24-bit 金幣（RE-073）",
                   "building_offset": BUILDING_OFFSET,
                   "building_rule": "留宿建築的編號，來自建築 payload 的 0x3A／0x3B"},
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
