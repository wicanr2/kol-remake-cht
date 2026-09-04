#!/usr/bin/env python3
"""IDA Pro 9.4：幾個全域是誰寫的。

帶基底暫存器的存取（`mov [bx+29h], al`）沒有具名 xref，所以只能走過每一支函式的
每一條指令，比對 operand 的**值**。這一支掃下面幾個目前卡住的全域：

  word_37CF2  0x37CF2  規劃段收尾寫進 +0x25 的指令碼（`0x1FA6D`）
  word_345DE  0x345DE  哪一名角色該升位階（`0x21607`、`0x21636`，RE-103）
  word_345DA  0x345DA  倒下的隊員數；逃走時變成 0xFF
  word_34966  0x34966  倒下的敵人數；逃走時變成 0xFF
  byte_3AC7D  0x3AC7D  逃走或落敗掉東西的機率（`0x21765`）
  [0xB926]    0x3B7F6  任務要的那件東西的名字（`sub_2886C` 比的對象，RE-103）

  idat -A -S"tools/ida_probe_koltown_writers.py OUT.json INPUT.EXE" INPUT.i64
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
DGROUP_EA = 0x2DED0
# 名字 → 絕對位址。DGROUP 位移是位址減 DGROUP_EA。
TARGETS = {
    "word_345DE": 0x345DE,
    "word_345DA": 0x345DA,
    "word_34966": 0x34966,
    "byte_3AC7D": 0x3AC7D,
    "quest_item_0xB926": DGROUP_EA + 0xB926,
    # 規劃段收尾寫進 +0x25 的指令碼（`0x1FA6D`）。分派表沒有 0x22，
    # 要知道 default 那一條寫進去的是不是選中的圖號（RE-105）。
    "word_37CF2": 0x37CF2,
}
# 附近幾個 byte 也算命中，`mov [bx+2], al` 這種寫法會落在後面。
SPAN = 4
# RAW_VALUES 是「直接找這個 operand 值」的目標，給角色記錄裡的位移用——
# 那種存取寫成 `mov al, [bx+1DDh]`，位址是基底暫存器加位移，沒有絕對位址可比。
RAW_VALUES = {
    # 結局的判準那一格（`sub_2B0FF:0x2B392`，RE-106）。
    "char_0x1DD": 0x1DD,
    # 任務旗標的群組（`0x28F73`、`0x2B396`）。
    "word_2DF70": 0x2DF70,
}


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

    # 兩種寫法都要抓：絕對位址（`mov word_345DE, ax`）與 DGROUP 位移
    # （`mov cx, 670Eh` 之後拿 bx 當基底）。
    wanted = {}
    for name, ea in TARGETS.items():
        for delta in range(SPAN):
            wanted[ea + delta] = name
            wanted[ea + delta - DGROUP_EA] = name

    for name, v in RAW_VALUES.items():
        wanted[v] = name

    hits = {name: [] for name in list(TARGETS) + list(RAW_VALUES)}
    scanned = 0
    for func_ea in idautils.Functions():
        f = ida_funcs.get_func(func_ea)
        if f is None:
            continue
        scanned += 1
        for ea in idautils.FuncItems(func_ea):
            for op in range(3):
                t = idc.get_operand_type(ea, op)
                if t == idc.o_void:
                    break
                if t not in (idc.o_mem, idc.o_imm, idc.o_displ, idc.o_phrase):
                    continue
                v = idc.get_operand_value(ea, op)
                name = wanted.get(v)
                if name is None:
                    continue
                hits[name].append({
                    "func": f"0x{func_ea:x}",
                    "func_name": idc.get_func_name(func_ea),
                    "ea": f"0x{ea:x}",
                    "operand": op,
                    "disassembly": disassembly(ea),
                })

    receipt = {
        "schema": "kol-ida-koltown-writers-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：幾個全域的 operand 值掃描；尚未鎖定任何斷言。",
        "functions_scanned": scanned,
        "targets": {k: f"0x{v:x}" for k, v in TARGETS.items()},
        "raw_values": {k: f"0x{v:x}" for k, v in RAW_VALUES.items()},
        "hits": hits,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
