#!/usr/bin/env python3
"""IDA Pro 9.4：接任務那一側（`sub_14564`）與任務一覽（`sub_167D9`）。

回報那一側解完了（RE-107／RE-109），接取那一側只讀到門檻判定
（`0x1459C`…`0x145F2`：旗標要是 0，最終任務另外要求已回報數 ≥ 0x17）。
當時要查的是整段接取：誰能給任務、給哪一個、旗標怎麼從「沒接」變成「接了」、
`[0xB926]` 那十六個 byte（戰鬥中比對「手上那件是不是任務要的」）是誰填的。

本探針一次問完，因為同一個 `.i64` 連續開關會弄壞它：

  * `sub_14564` 與 `sub_167D9`、`sub_16239` 的完整反組譯與呼叫者；
  * 對 `[0xB926]`、`[0xB956]` 與任務旗標 `[0x1E2]` 做 **operand 值掃描**——
    帶基底暫存器的存取（`mov [bx+B926h], al`）沒有具名 xref，只能走過每一支
    函式的每一條指令比對 operand 的值；
  * 任務 payload 的相關 DGROUP 區塊。

  idat -A -S"tools/ida_probe_koltown_quest_take.py OUT.json INPUT.EXE" INPUT.i64
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
# 要倒全文的幾支。
FUNCTIONS = (0x14564, 0x167D9, 0x16239)
# operand 要比對的值：角色記錄裡的任務欄位與那兩塊十六 byte。
WANTED = (0x1E2, 0x1DD, 0xB926, 0xB956, 0x2E0, 0x2F0)
# 要倒的 DGROUP 區塊：兩塊十六 byte 的暫存區。
BLOCKS = ((0xB926, 0x20), (0xB956, 0x20))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    f = ida_funcs.get_func(start)
    if f is None:
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


def scan_operands(wanted: set[int]) -> dict:
    """走過每一支函式的每一條指令，比對兩個 operand 的**值**。"""
    hits: dict[str, list[dict]] = {f"0x{v:x}": [] for v in wanted}
    for start in idautils.Functions():
        name = idc.get_func_name(start)
        for ea in idautils.FuncItems(start):
            for n in range(2):
                v = idc.get_operand_value(ea, n)
                if v in wanted:
                    hits[f"0x{v:x}"].append({
                        "ea": f"0x{ea:x}", "function": name, "operand": n,
                        "disassembly": disassembly(ea)})
    return hits


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-quest-take-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：接任務與任務一覽；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "operand_hits": scan_operands(set(WANTED)),
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
