#!/usr/bin/env python3
"""把 `MAGIC.0` 抽成法術的參數字典（RE-128）。

法術編輯器（SCROLL 選單第 1 列）十三個欄位的選項全部來自這個檔案：開機時
`sub_10000` 用 `sub_1A4FE("magic", 0, ds:byte_32386)` 把它載進 DGROUP+0x44B6，
`sub_18350` 再拿記錄裡的半位元組去查。

欄位 0 是法術**名稱**（`sub_1B5C8` 收 12 個字、`sub_1C0F6` 打包成 8 個 byte 的
5-bit 名），不查這個檔案；欄位 1…4／5…8 是四格護符的高／低 nibble，9 是屬性，
0x0A 是 `+0x0E` 的 bit7，0x0B／0x0C 是兩個骰式。

檔案結構：一個 payload，用 `TOWN0` 那一套 LZW 解開得到 4,097 個 byte，分成
0x100 一段的區塊。每一段的頭 16 個 byte 是字串位址的低半、接著 16 個是高半，
之後才是字串本身；位址換算是 **`低 + (高 − 8) × 256`**（高位那 8 是載入時的
segment 偏移）。成本那幾段沒有字串，直接是 16-bit 的數字。

  tools/run_python_tool.sh tools/extract_magic.py --zip /in/kol.zip \
      --out assets/original-derived-data/magic.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import pathlib
import sys
import zipfile

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from probe_town_data import lzw_decode, payloads, read_volumes  # noqa: E402

ARCHIVE = "MAGIC"
# 解開之後的長度。對不上就代表換了一份輸入或解碼壞了，直接失敗離開。
EXPECTED_BYTES = 4097
# 字串位址的高半要減掉的 segment 偏移。
STRING_BIAS = 8
# 一段裡的位址表有幾筆。
ENTRIES = 16

# 有名稱的那幾段：(名字, 低半位址表, 高半位址表, 最多幾筆)。
NAME_TABLES = (
    # 欄位 0…4：法術對哪些種族有效。
    ("class", 0x000, 0x010, ENTRIES),
    # 欄位 9：影響哪一個屬性。
    ("stat", 0x100, 0x110, ENTRIES),
    # 欄位 0x0B／0x0C：骰式。
    ("dice", 0x260, 0x290, 0x30),
)
# 欄位 5…8 的選項按高 nibble 分成十六組，每組十六筆。
SUBCLASS_LO, SUBCLASS_HI, SUBCLASS_GROUPS = 0x400, 0x500, 16
# 兩種成本：索引是那一格的整個 byte（0…255）。
COST_TABLES = (
    ("adventure", 0x600, 0x700),
    ("gold", 0x800, 0x900),
)
# `+0x0C` 那一格自己的兩種成本。
STAT_COST = (("adventure", 0x120, 0x130), ("gold", 0x140, 0x150))
COST_ENTRIES = 0x100
# 骰式選單挑第 n 項時真正寫進記錄的值（`0x18109`：`[0x200 + n]`）。記錄那一格
# 只換中間六個位元，bit0 與 bit7 留著——`+0x0E` 的 bit7 是「要不要先選目標」。
DICE_CODE = 0x200
DICE_ENTRIES = 0x30
# 骰式的成本（`sub_18296`）：骰數與骰面分開算，兩張表各八筆、都是單一 byte。
# 索引是從記錄那一格拆出來的 `(值 & 0x70) >> 4` 與 `(值 & 0x0E) >> 1`。
DICE_COST = (("count_adventure", 0x240), ("count_gold", 0x248),
             ("sides_adventure", 0x250), ("sides_gold", 0x258))
DICE_COST_ENTRIES = 8
# 欄位 0x0A（`+0x0E` 的 bit7）沒有自己的表：兩種選擇一律是 100 × 值 ＋ 100，
# 冒險點數與金冠同額（`0x1864B`…`0x18666`）。
TARGET_FLAG_COST_STEP = 100


def text_at(data: bytes, lo: int, hi: int) -> str | None:
    """把一組低／高 byte 換算成字串。兩個都是 0 代表這一格空著。"""
    if lo == 0 and hi == 0:
        return None
    off = lo + ((hi - STRING_BIAS) << 8)
    if off < 0 or off >= len(data):
        raise SystemExit(f"字串位址 0x{off:04x} 落在檔案外")
    raw = data[off:]
    for end, b in enumerate(raw):
        if b in (0, 0x0D):
            return raw[:end].decode("ascii", "replace")
    raise SystemExit(f"字串位址 0x{off:04x} 沒有結尾")


def names(data: bytes, lo_base: int, hi_base: int, count: int) -> list[str]:
    """一段的名稱清單。原版是「遇到空字串就停」，這裡照做。"""
    out: list[str] = []
    for i in range(count):
        s = text_at(data, data[lo_base + i], data[hi_base + i])
        if not s:
            break
        out.append(s)
    return out


def costs(data: bytes, lo_base: int, hi_base: int, count: int) -> list[int]:
    return [data[lo_base + i] | (data[hi_base + i] << 8) for i in range(count)]


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--zip", required=True)
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    with zipfile.ZipFile(args.zip) as zf:
        volumes = read_volumes(zf, ARCHIVE)
        contract, parts = payloads(volumes, ARCHIVE)
        if len(parts) != 1:
            raise SystemExit(f"{ARCHIVE} 有 {len(parts)} 個 payload，預期 1 個")
        raw = parts[0]
        data = lzw_decode(raw)
    if len(data) != EXPECTED_BYTES:
        raise SystemExit(f"解開之後是 {len(data)} 個 byte，預期 {EXPECTED_BYTES}")

    doc = {
        "schema": "kol-magic-dictionary-v1",
        "tool": pathlib.Path(__file__).name,
        "source": f"Knights-of-legend/{ARCHIVE}.0",
        "source_sha256": hashlib.sha256(raw).hexdigest(),
        "decoded_bytes": len(data),
        "decoded_sha256": hashlib.sha256(data).hexdigest(),
        "contract": contract,
        "note": "法術編輯器十三個欄位的選項與成本（RE-128）。位址換算：低 + (高 − 8) × 256。",
    }
    for name, lo, hi, count in NAME_TABLES:
        doc[name] = names(data, lo, hi, count)
    doc["subclass"] = [names(data, SUBCLASS_LO + g * ENTRIES,
                             SUBCLASS_HI + g * ENTRIES, ENTRIES)
                       for g in range(SUBCLASS_GROUPS)]
    doc["cost"] = {name: costs(data, lo, hi, COST_ENTRIES)
                   for name, lo, hi in COST_TABLES}
    doc["stat_cost"] = {name: costs(data, lo, hi, ENTRIES)
                        for name, lo, hi in STAT_COST}
    doc["dice_code"] = list(data[DICE_CODE:DICE_CODE + DICE_ENTRIES])
    doc["dice_cost"] = {name: list(data[base:base + DICE_COST_ENTRIES])
                        for name, base in DICE_COST}
    doc["target_flag_cost_step"] = TARGET_FLAG_COST_STEP

    out = pathlib.Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    print(json.dumps({"out": str(out), "class": len(doc["class"]),
                      "stat": len(doc["stat"]), "dice": len(doc["dice"]),
                      "subclass_groups": sum(1 for g in doc["subclass"] if g)},
                     ensure_ascii=False))
    return 0


raise SystemExit(main())
