#!/usr/bin/env python3
"""IDA 9.4：列出 KOLTOWN 建築訊息繪製器的原始交叉參照與鄰近指令。

探索輸出只放本機；不得把含原版指令 bytes 的 JSON 加入 Git。
idat -A -S"SCRIPT OUT.json INPUT.EXE" 一次性副本.i64
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
TARGET = 0x14234
EXTRA_WINDOWS = ((0x11B4B, 0x11BE9), (0x11C2E, 0x11D20))


def row(ea: int) -> dict:
    size = max(1, ida_bytes.get_item_size(ea))
    raw = ida_bytes.get_bytes(ea, min(size, 12)) or b""
    return {
        "ea": f"0x{ea:x}",
        "bytes": raw.hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "operand_values": [idc.get_operand_value(ea, n) for n in range(2)],
    }


def context(ea: int) -> dict:
    func = ida_funcs.get_func(ea)
    if func is None:
        return {"xref_ea": f"0x{ea:x}", "error": "xref outside recognized function"}
    addresses = list(idautils.FuncItems(func.start_ea))
    if ea not in addresses:
        return {"xref_ea": f"0x{ea:x}", "error": "xref absent from function items"}
    index = addresses.index(ea)
    lo, hi = max(0, index - 18), min(len(addresses), index + 6)
    return {
        "xref_ea": f"0x{ea:x}",
        "original_function_name": idc.get_func_name(func.start_ea),
        "function_start_ea": f"0x{func.start_ea:x}",
        "function_end_ea": f"0x{func.end_ea:x}",
        "function_instruction_count": len(addresses),
        "window": [row(address) for address in addresses[lo:hi]],
    }


def bounded_window(start: int, end: int) -> dict:
    items = []
    ea = start
    while ea < end:
        items.append(row(ea))
        next_ea = idc.next_head(ea, end)
        if next_ea <= ea:
            raise RuntimeError(f"no next head after 0x{ea:x}")
        ea = next_ea
    return {"start_ea": f"0x{start:x}", "end_ea": f"0x{end:x}", "rows": items}


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    output, original = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(original.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    target_function = ida_funcs.get_func(TARGET)
    if target_function is None or target_function.start_ea != TARGET:
        raise RuntimeError(f"expected function missing at 0x{TARGET:x}")
    code_refs = sorted(set(idautils.CodeRefsTo(TARGET, False)))
    data_refs = sorted(set(idautils.DataRefsTo(TARGET)))
    receipt = {
        "schema": "kol-ida-message-consumers-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": original.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "target": {"original_name": idc.get_func_name(TARGET), "ea": f"0x{TARGET:x}"},
        "evidence_level": "探索用 raw xref；語意待人工審查，間接呼叫仍須另查",
        "code_ref_count": len(code_refs),
        "code_refs": [context(ea) for ea in code_refs],
        "data_refs": [f"0x{ea:x}" for ea in data_refs],
        "extra_windows": [bounded_window(a, b) for a, b in EXTRA_WINDOWS],
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


main()
