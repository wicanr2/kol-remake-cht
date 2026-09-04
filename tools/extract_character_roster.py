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

from extract_town_buildings import good  # noqa: E402
from extract_town_text import render  # noqa: E402
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
# 任務旗標：六個 byte，一個 byte 塞四個 2-bit 欄位，對應 24 個任務地點（RE-079）。
QUEST_OFFSET, QUEST_BYTES = 0x1E2, 6
# 結局的判準：`[0x1DD + 群組]` 等於 0x17 就播（`sub_2B0FF:0x2B38F`，RE-106）。
# 那一格是誰寫的還沒解，先原樣帶出來。
PROGRESS_OFFSET, PROGRESS_BYTES = 0x1DD, 5
# 交任務物品時要逐 byte 比對的那十七格（`sub_11A93:0x11B85`…`0x11CF7`，RE-109）：
# 0x30 起每 0x10 一格、不到 0x1A0 為止，跳過 0x40…0x9F 那一段武器技能。
CARRIED_FIRST, CARRIED_LAST, CARRIED_STRIDE = 0x30, 0x190, 0x10
CARRIED_SKIP_LO, CARRIED_SKIP_HI = 0x40, 0x9F
# 穿戴欄：部位 0…8 在 0x100 + 16 × 部位，部位 9（手上那件武器）在 0x30，
# 一格 16 byte，版面與商店的商品記錄相同（RE-075／RE-080）。
WORN_SLOTS, WORN_STRIDE = 10, 16
WORN_BASE, WEAPON_SLOT_OFFSET = 0x100, 0x30
# 位移 0x2E／0x2F 是角色留宿的建築編號，來自建築 payload 的 0x3A／0x3B。
BUILDING_OFFSET = 0x2E
# MIRROR 角色資料頁用得到的欄位（RE-074）。位移 0x1D–0x23 是七個內部屬性，
# 順序與說明書相同：力量、敏捷、體型、健康、先見、魅力、智力。
STATS_OFFSET, STATS_COUNT = 0x1D, 7
PROFESSION_OFFSET = 0x1C      # ＋6 之後查 KOLDATA；非人族與女性精靈再減 6
RACE_OFFSET = 0x24            # 0…3 ＝ Human／Elven／Dwarven／Kelden
RANK_OFFSET = 0x29            # KOLDATA 0x74 + (值/5)*5，餘數非 0 再畫 0x74 + 值
ADVENTURE_OFFSET = 0x26       # 高位元組在前的 16-bit
MOVEMENT_OFFSET = 0x1BA       # 0 當成 20，畫面上印 chr(0x41 + 值)
NUTRITION_OFFSET = 0x9F
ARMOUR_OFFSET, SHIELD_OFFSET = 0x1A0, 0x1A8
# 法術：角色記錄 0x200 起、每 16 byte 一筆、共十六筆（RE-100）。版面與商品記錄
# 相同，所以名字與骰式都用同一支解。
SPELL_OFFSET, SPELL_STRIDE, SPELL_COUNT = 0x200, 16, 16
WEAPON_OFFSET, WEAPON_STRIDE, WEAPON_ROWS = 0x4A, 13, 4
WEAPON_SKILL_OFFSET = 0x53    # OFF 在 +0／+1、DEF 在 +2／+3，都是大端 16-bit 除以 10
# 身高體重表在 KOLTOWN 的 DGROUP，索引都是 2 × 性別 + 種族（RE-074）。
HEIGHT_BASE = (0x1C, 0x10, 0x1C, 0x1C, 0x1E, 0x00, 0x14)
WEIGHT_BASE = (0x03, 0x03, 0x02, 0x03, 0x03, 0x00, 0x03)
WEIGHT_PER_FOOT = (0x41, 0x32, 0x4B, 0x46, 0x28, 0x00, 0x23)


def equipment(rec: bytes) -> list[dict | None]:
    """穿戴欄十格。空的那一格是 null；其餘欄位與商店的商品記錄同名。"""
    out = []
    for pos in range(WORN_SLOTS):
        base = WEAPON_SLOT_OFFSET if pos == WORN_SLOTS - 1 else WORN_BASE + WORN_STRIDE * pos
        block = rec[base:base + WORN_STRIDE]
        if len(block) < WORN_STRIDE or not any(block):
            out.append(None)
            continue
        item = good(block, pos)
        item["offset"] = base
        out.append(item)
    return out


def carried(rec: bytes) -> list[dict]:
    """交任務物品時會被掃到的那十七格，原樣帶出來（RE-109）。

    比對是逐 byte 相同，所以 typed 的穿戴欄不夠用——那一份丟掉了原始位元組。"""
    out = []
    off = CARRIED_FIRST
    while off <= CARRIED_LAST:
        if not CARRIED_SKIP_LO <= off <= CARRIED_SKIP_HI:
            out.append({"offset": off,
                        "bytes": rec[off:off + CARRIED_STRIDE].hex()})
        off += CARRIED_STRIDE
    return out


def spells(rec: bytes) -> list[dict | None]:
    """法術十六格（RE-100）。空的那一格是 null。

    程式只讀四處：+0x08…+0x0B 的四格護符、+0x0C 的效果、+0x0D 的持續回合骰式、
    +0x0E 的強度骰式（bit 7 是「要不要先選目標格」）。其餘欄位與商品記錄同名，
    照樣解出來當交叉檢查。"""
    out = []
    for i in range(SPELL_COUNT):
        base = SPELL_OFFSET + SPELL_STRIDE * i
        block = rec[base:base + SPELL_STRIDE]
        if len(block) < SPELL_STRIDE or not any(block):
            out.append(None)
            continue
        item = good(block, i)
        item["offset"] = base
        item["effect"] = block[0x0C]
        item["turn_code"] = block[0x0D]
        item["power_code"] = block[0x0E]
        item["needs_target"] = bool(block[0x0E] & 0x80)
        item["guards"] = list(block[0x08:0x0C])
        out.append(item)
    return out


def packed_name(raw: bytes) -> str:
    """解 5-bit 壓縮的裝備名稱：第一個 byte 的高 nibble 是字數，
    字元從第 4 個 bit 起每 5 個 bit 一個（RE-074）。

    字元走 `sub_1B94F` 的 mode 1（RE-075）：每個字的第一個字母轉大寫，
    所以資料裡是 `Greatsword`。角色資料頁另外整段轉大寫，那是畫面的事。"""
    count = raw[0] >> 4
    values = []
    for i in range(count):
        bit = 4 + i * 5
        j = bit >> 3
        word = (raw[j] << 8) | (raw[j + 1] if j + 1 < len(raw) else 0)
        values.append((word >> (11 - (bit & 7))) & 0x1F)
    text, unknown = render(values, capitalise=True)
    if unknown:
        raise SystemExit(f"名稱裡有未定的跳脫序列 {unknown}")
    return text.rstrip()


def character_sheet(rec: bytes) -> dict:
    """MIRROR 那一頁要用的欄位。算式與版面見 RE-074。"""
    gender, race = rec[FEMALE_OFFSET], rec[RACE_OFFSET]
    index = 2 * gender + race
    if index >= len(HEIGHT_BASE):
        raise SystemExit(f"性別 {gender} 種族 {race} 的身高索引 {index} 超出表")
    inches = rec[STATS_OFFSET + 2] - HEIGHT_BASE[index]
    feet, rest = divmod(inches, 12)
    weight = (feet - WEIGHT_BASE[index]) * WEIGHT_PER_FOOT[index] + rest * 4
    weapons = []
    for i in range(WEAPON_ROWS):
        base = WEAPON_OFFSET + WEAPON_STRIDE * i
        if rec[base] == 0:
            continue
        skill = WEAPON_SKILL_OFFSET + WEAPON_STRIDE * i
        weapons.append({
            "row": i,
            "name": packed_name(rec[base:base + WEAPON_STRIDE]),
            # 位移 +8 是武器識別，訓練時用它比對這名角色有沒有練過（RE-076）。
            "weapon": rec[base + 8],
            "offence": (rec[skill] << 8 | rec[skill + 1]) // 10,
            "defence": (rec[skill + 2] << 8 | rec[skill + 3]) // 10,
            "offence_raw": rec[skill] << 8 | rec[skill + 1],
            "defence_raw": rec[skill + 2] << 8 | rec[skill + 3],
        })
    return {
        "stats": list(rec[STATS_OFFSET:STATS_OFFSET + STATS_COUNT]),
        "race": race,
        "profession": rec[PROFESSION_OFFSET],
        "rank": rec[RANK_OFFSET],
        "adventure_points": rec[ADVENTURE_OFFSET] << 8 | rec[ADVENTURE_OFFSET + 1],
        "movement_rate": rec[MOVEMENT_OFFSET],
        "nutrition": rec[NUTRITION_OFFSET],
        "height_feet": feet,
        "height_inches": rest,
        "weight": weight,
        "weapons": weapons,
        "armour": packed_name(rec[ARMOUR_OFFSET:ARMOUR_OFFSET + 8]),
        "shield": packed_name(rec[SHIELD_OFFSET:SHIELD_OFFSET + 8]),
    }
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
            "quests": list(data[base + QUEST_OFFSET:
                                base + QUEST_OFFSET + QUEST_BYTES]),
            "progress": list(data[base + PROGRESS_OFFSET:
                                  base + PROGRESS_OFFSET + PROGRESS_BYTES]),
            "sheet": character_sheet(data[base:base + SLOT_BYTES]) if name else None,
            # 穿戴欄是角色記錄自己的欄位，不屬於 MIRROR 資料頁，所以放在槽位這一層
            # （`internal/roster` 的 Slot.Equipment 就是照這個位置讀）。
            "equipment": equipment(data[base:base + SLOT_BYTES]) if name else None,
            # 法術同樣是角色記錄自己的欄位（RE-100）。
            "spells": spells(data[base:base + SLOT_BYTES]) if name else None,
            # 交任務物品要逐 byte 比，所以那十七格原樣帶出來（RE-109）。
            "carried": carried(data[base:base + SLOT_BYTES]) if name else None,
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
                   "building_rule": "留宿建築的編號，來自建築 payload 的 0x3A／0x3B",
                   "stats_offset": STATS_OFFSET, "stats_count": STATS_COUNT,
                   "stats_rule": "七個內部屬性：力量、敏捷、體型、健康、先見、魅力、智力；"
                                 "畫面值是內部值 × 100 / 150（RE-074）",
                   "sheet_rule": "MIRROR 角色資料頁的欄位與算式見 RE-074"},
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
