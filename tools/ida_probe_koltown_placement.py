#!/usr/bin/env python3
"""IDA Pro 9.4：戰場站位表 `0x2E0C` 是誰填的。

任務地點把 payload `+3` 起 36 byte 複製到 `DGROUP+0x2E0C`（RE-090），那是 12 隻怪
與 6 名隊員的站位。隨機遭遇沒有那一段，站位一定是別處產生的。

`0x2E0C` 沒有具名 xref——**帶基底暫存器的存取（`mov [bx+2E0Ch], al`）靜態符號掃描
抓不到**（RE-092 的教訓）。所以這一支改成對**位址範圍**掃 operand：走過全部函式的
每一條指令，把任何一個 operand 的值落在指定範圍內的都記下來。

  idat -A -S"tools/ida_probe_koltown_placement.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
# 想找的 DGROUP 位移範圍（起、迄、說明）。
RANGES = (
    (0x2E0C, 0x2E30, "戰場站位表：12 隻怪與 6 名隊員，各兩個 byte"),
    (0x2BAE, 0x2BBE, "戰場物品的 X"),
    (0x2D9E, 0x2DAE, "戰場物品的 Y"),
    (0x322E, 0x3252, "法術效果 5 改的那個全域：一名參戰者兩個 byte（RE-100）"),
)
MAX_OPERANDS = 3


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    hits = {f"0x{lo:04x}-0x{hi:04x}": [] for lo, hi, _ in RANGES}
    for func_ea in idautils.Functions():
        name = idc.get_func_name(func_ea)
        for ea in idautils.FuncItems(func_ea):
            for n in range(MAX_OPERANDS):
                if idc.get_operand_type(ea, n) == idaapi.o_void:
                    break
                value = idc.get_operand_value(ea, n)
                for lo, hi, _ in RANGES:
                    if lo <= value < hi:
                        hits[f"0x{lo:04x}-0x{hi:04x}"].append({
                            "ea": f"0x{ea:x}", "func": name,
                            "operand": n, "value": f"0x{value:04x}",
                            "disassembly": disassembly(ea)})

    receipt = {
        "schema": "kol-ida-koltown-placement-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：對 DGROUP 位移做範圍掃描，找帶基底暫存器的存取；"
                "尚未鎖定任何斷言。",
        "ranges": [{"from": f"0x{lo:04x}", "to": f"0x{hi:04x}", "what": what}
                   for lo, hi, what in RANGES],
        "hits": hits,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
