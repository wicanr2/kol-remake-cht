#!/usr/bin/env python3
"""IDA Pro 9.4：馬廄那一頁的坐騎圖——從哪個檔案讀、畫在哪、買下之後搬去哪。

RE-119 讀出四匹的圖號、名稱與價錢，但**圖本身還沒接上**：明細那一頁呼叫
`sub_1A4FE`（`0x12a47`）畫圖，而那支的資料來源、座標與尺寸都還沒讀。原版 ZIP 裡
與坐騎有關的檔案有兩個：`STEEDPIC`（16 槽 × 760 byte，是角色自己那一張）與
`STEEDS.0`（四筆 LZW payload，各解出 761 byte），要分清楚誰是模板誰是角色槽。

一次問完：
  * `sub_1a4fe` 全文（畫圖）與 `sub_12a14`…`0x12d17` 這一段（明細、折抵、買下）。
  * `word_2fb5c` 這個緩衝區指標的每一個讀寫點。
  * 開檔那幾支：'steeds'、'steedpic' 兩個字串的參照。

  idat -A -S"tools/ida_probe_koltown_steed.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
DGROUP_EA = 0x2DED0
# 畫圖那一支，以及它上游的載入／解壓。
FUNCTIONS = (0x1A4FE,)
# 馬廄那一段：清單、明細、折抵、買下、搬圖。
WINDOWS = ((0x1291F, 0x12D20),)
# 緩衝區指標與圖號暫存。
GLOBALS = {
    "pic_buffer_word_2fb5c": 0x2FB5C,
    "steed_movement_word_39a54": 0x39A54,
    "party_movement_word_2fa2c": 0x2FA2C,
}
# 檔名字串（DGROUP 位移）：'steeds'、'steedpic'、'steedpics'。
STRINGS = ("steeds", "steedpic", "steedpics")


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def listing(start: int) -> list[dict]:
    if ida_funcs.get_func(start) is None:
        return []
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{f.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
        "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)}),
    }


def refs_to(ea: int) -> list[dict]:
    out = []
    for xref in idautils.XrefsTo(ea, 0):
        out.append({"from": f"0x{xref.frm:x}", "type": xref.type,
                    "disassembly": disassembly(xref.frm),
                    "function": idc.get_func_name(xref.frm) or ""})
    return out


def find_strings() -> dict:
    found = {}
    for s in idautils.Strings():
        text = str(s)
        if text.strip("\x00") in STRINGS:
            found.setdefault(text.strip("\x00"), []).append({
                "ea": f"0x{s.ea:x}", "refs": refs_to(s.ea)})
    return found


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-steed-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：馬廄那一頁的坐騎圖來源、座標與搬運；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "globals": {name: refs_to(ea) for name, ea in GLOBALS.items()},
        "strings": find_strings(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
