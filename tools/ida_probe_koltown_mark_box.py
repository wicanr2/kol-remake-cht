#!/usr/bin/env python3
"""IDA Pro 9.4：物品頁左上那個小框與投影原語（RE-128 對拍用）。

資料頁那兩頁對原版畫面時剩下的差異有結構：小框右邊 `x = 106…109`、下面
`y = 62…65` 各一條四像素寬的帶，原版是黑的而 remake 讓背景透了出來——同一個
形狀在整頁的外框上已經證實是 `sub_1B7E9` 的投影。這一支把畫小框那支
（`sub_16D0B`）、投影原語本身與小框用的那個填色框（`sub_1BA9D`）倒出來，
確認那兩條帶的來源與偏移量。

  idat -A -S"tools/ida_probe_koltown_mark_box.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x16D0B, 0x1B7E9, 0x1BA9D, 0x18F3A, 0x1C70E, 0x16FA6)


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


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-mark-box-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：物品頁左上小框與投影原語；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
