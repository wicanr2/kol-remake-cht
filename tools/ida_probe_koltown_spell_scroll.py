#!/usr/bin/env python3
"""IDA Pro 9.4：SCROLL 那一格，以及法術剩下的三個洞。

兩件事一次問完（同一個 `.i64` 連續開關會弄壞它）：

  * **SCROLL**（建築內圖示 0x0D，派工表指到 `sub_17318`，RE-074）：整支還沒讀過，
    remake 按下去只記一筆未實作。
  * **法術剩下的**（SPEC-012 第 10 節）：疲勞折算的基準 `[0x1C50]`、疲勞累計
    `[0x322E + 2 × 參戰者]` 的讀寫、以及參戰者記錄 `+0x2C`／`+0x2E`／`+0x2F`
    三個欄位的語意（效果 7／8 打的就是後兩格）。

`+0x2C` 這種小位移沒辦法用 operand 值去掃——`add sp, 2Ch` 也會中。這裡改成
掃反組譯字串裡的 `+2Ch]` 形式，只收真的帶基底暫存器的記憶體存取。

  idat -A -S"tools/ida_probe_koltown_spell_scroll.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
import re
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
# 要倒全文的幾支：SCROLL、疲勞折算、部位效果。
FUNCTIONS = (0x17318, 0x23683, 0x277F4)
# operand 值掃描：疲勞累計表、效果欄位表、法術區。
WANTED = (0x322E, 0x12A4, 0x1C50)
# 反組譯字串掃描：參戰者記錄的三個未解欄位。
FIELD_PATTERNS = ("+2Ch]", "+2Eh]", "+2Fh]")
# 要倒的 DGROUP 區塊。
BLOCKS = ((0x1C50, 0x10), (0x12A4, 0x10), (0x322E, 0x20))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


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


def scan() -> tuple[dict, dict]:
    """一次走完所有函式：operand 值與反組譯字串兩種比對同時做。"""
    wanted = set(WANTED)
    values: dict[str, list[dict]] = {f"0x{v:x}": [] for v in wanted}
    fields: dict[str, list[dict]] = {p: [] for p in FIELD_PATTERNS}
    for start in idautils.Functions():
        name = idc.get_func_name(start)
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            for n in range(2):
                v = idc.get_operand_value(ea, n)
                if v in wanted:
                    values[f"0x{v:x}"].append({
                        "ea": f"0x{ea:x}", "function": name, "operand": n,
                        "disassembly": text})
            for pattern in FIELD_PATTERNS:
                if pattern in text:
                    fields[pattern].append({
                        "ea": f"0x{ea:x}", "function": name, "disassembly": text})
    return values, fields


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    values, fields = scan()
    receipt = {
        "schema": "kol-ida-koltown-spell-scroll-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：SCROLL 那一格與法術剩下的三個洞；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "operand_hits": values,
        "field_hits": fields,
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
