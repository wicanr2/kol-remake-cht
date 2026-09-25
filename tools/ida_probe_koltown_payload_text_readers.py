#!/usr/bin/env python3
"""IDA 9.4：列出 KOLTOWN 文字解碼器的所有呼叫點與建築 payload 緩衝區的位址運算元（RE-140）。

建築 payload 緩衝區在 DGROUP:0x1CF4；訊息 n 位於 0x1CF4 + 0x80 + n*0x40。
這支回答「訊息 8／9（0x1F74..0x1FF3）除了 `sub_14234` 以外還有沒有讀取者」：
一是 `sub_1B94F` 的每個呼叫點與其前面推入的引數，二是全程式中立即值、
記憶體位址或位移落在 0x1CF4..0x1FF4 的指令。
探索輸出含原版指令 bytes，只放本機，不得加入 Git。

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
import ida_segment
import ida_ua
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
DECODER = 0x1B94F
BUFFER = 0x1CF4
RANGE = (0x1CF4, 0x1FF4)


def row(ea: int) -> dict:
    size = max(1, ida_bytes.get_item_size(ea))
    raw = ida_bytes.get_bytes(ea, min(size, 12)) or b""
    return {"ea": f"0x{ea:x}", "bytes": raw.hex(),
            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")}


def window(ea: int, before: int, after: int) -> list:
    func = ida_funcs.get_func(ea)
    if func is None:
        return [row(ea)]
    addresses = list(idautils.FuncItems(func.start_ea))
    if ea not in addresses:
        return [row(ea)]
    i = addresses.index(ea)
    return [row(a) for a in addresses[max(0, i - before):i + after + 1]]


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    output, original = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(original.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    decoder = ida_funcs.get_func(DECODER)
    if decoder is None or decoder.start_ea != DECODER:
        raise RuntimeError(f"expected function missing at 0x{DECODER:x}")

    callers = []
    for ea in sorted(set(idautils.CodeRefsTo(DECODER, False))):
        callers.append({"ea": f"0x{ea:x}", "func": idc.get_func_name(ea) or None,
                        "window": window(ea, 8, 1)})

    in_range = []
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        for ea in idautils.Heads(seg.start_ea, seg.end_ea):
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
                continue
            insn = ida_ua.insn_t()
            if not ida_ua.decode_insn(insn, ea):
                continue
            for op in insn.ops:
                if op.type == ida_ua.o_void:
                    break
                if op.type == ida_ua.o_imm:
                    value = op.value & 0xFFFF
                elif op.type in (ida_ua.o_mem, ida_ua.o_displ):
                    value = op.addr & 0xFFFF
                else:
                    continue
                if RANGE[0] <= value < RANGE[1]:
                    in_range.append({"ea": f"0x{ea:x}", "func": idc.get_func_name(ea) or None,
                                     "operand_type": op.type, "value": f"0x{value:x}",
                                     "payload_offset": f"0x{value - BUFFER:x}",
                                     "window": window(ea, 6, 4)})
                    break

    receipt = {
        "schema": "kol-ida-payload-text-readers-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": original.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ；"
                         "運算元值為 DGROUP 內 16 位元位移",
        "evidence_level": "探索用 raw 掃描；語意待人工審查",
        "decoder": {"original_name": idc.get_func_name(DECODER), "ea": f"0x{DECODER:x}"},
        "decoder_callers": callers,
        "range": [f"0x{RANGE[0]:x}", f"0x{RANGE[1]:x}"],
        "operands_in_range": in_range,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


main()
