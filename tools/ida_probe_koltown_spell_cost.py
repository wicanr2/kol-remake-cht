#!/usr/bin/env python3
"""IDA Pro 9.4：法術編輯器的骰式成本與教派允許表（RE-128 續）。

`sub_18350` 把四格護符與屬性那一格的成本都查表加總，唯獨兩個骰式欄位交給
`sub_18296` ——它同時負責印出骰式字串與把成本累加進去，所以那兩欄要花多少錢
只有讀它才知道。

另一件事是 `[0x1CF4 + 0x60 + n]`：欄位 1…4 選 CLASS 時那一格是零就不讓選。
`byte_2FBC4` 在載入映像裡整段是 `FF`（BSS），全檔只有四個參照點，其中兩個是
`mov ax, offset byte_2FBC4` ——**要看的是取了位址之後傳給誰**，所以連同那兩個
呼叫點附近的視窗一起倒出來。

  idat -A -S"tools/ida_probe_koltown_spell_cost.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x18296, 0x152C6, 0x29004, 0x11703, 0x11722, 0x1B90B)
# `sub_11A93` 是主迴圈，整支太大：只取取址那一行前後。
WINDOWS = ((0x11AD0, 0x11B40),)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    if ida_funcs.get_func(start) is None:
        return []
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def window(lo: int, hi: int) -> list[dict]:
    out, ea = [], lo
    while ea < hi:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{f.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
    }


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-spell-cost-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：骰式成本與教派允許表的來源；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "windows": {f"0x{lo:x}-0x{hi:x}": window(lo, hi) for lo, hi in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
