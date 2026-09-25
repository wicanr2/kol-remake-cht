#!/usr/bin/env python3
"""IDA 9.4：查 KOLTOWN 建築訊息 2–9 的間接讀取與任務旗標寫入點（RE-140）。

RE-139 只列了 `sub_14234` 的直接呼叫。這支一次問完剩下的可達性問題：
訊息緩衝區還有誰讀、建築旗標 `byte_2FC01` 附近有誰寫、全程式哪裡出現
訊息 8／9 的位移 0x280／0x2C0、`sub_14234` 的段內位移是否被當成函式指標，
以及所有間接呼叫點。探索輸出含原版指令 bytes，只放本機，不得加入 Git。

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
TARGET = 0x14234
FLAG_RANGE = (0x2FBF0, 0x2FC20)
MESSAGE_OFFSETS = (0x280, 0x2C0)


def row(ea: int) -> dict:
    size = max(1, ida_bytes.get_item_size(ea))
    raw = ida_bytes.get_bytes(ea, min(size, 12)) or b""
    return {
        "ea": f"0x{ea:x}",
        "func": idc.get_func_name(ea) or None,
        "bytes": raw.hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
    }


def window(ea: int, before: int = 10, after: int = 4) -> list:
    func = ida_funcs.get_func(ea)
    if func is None:
        return [row(ea)]
    addresses = list(idautils.FuncItems(func.start_ea))
    if ea not in addresses:
        return [row(ea)]
    i = addresses.index(ea)
    return [row(a) for a in addresses[max(0, i - before):i + after + 1]]


def code_heads():
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        for ea in idautils.Heads(seg.start_ea, seg.end_ea):
            if ida_bytes.is_code(ida_bytes.get_flags(ea)):
                yield ea


def memory_targets(func_start: int) -> list[int]:
    """sub_14234 內所有記憶體運算元的目標位址。"""
    out = set()
    for ea in idautils.FuncItems(func_start):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea):
            continue
        for op in insn.ops:
            if op.type == ida_ua.o_void:
                break
            if op.type == ida_ua.o_mem:
                out.add(op.addr)
        for ref in idautils.DataRefsFrom(ea):
            out.add(ref)
    return sorted(out)


def refs_to(target: int) -> list:
    rows = []
    for ea in sorted(set(idautils.DataRefsTo(target))):
        rows.append({"ref_ea": f"0x{ea:x}", "func": idc.get_func_name(ea) or None,
                     "window": window(ea, 6, 3)})
    return rows


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    output, original = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(original.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    func = ida_funcs.get_func(TARGET)
    if func is None or func.start_ea != TARGET:
        raise RuntimeError(f"expected function missing at 0x{TARGET:x}")
    seg = ida_segment.getseg(TARGET)
    seg_base = seg.start_ea
    near_offset = TARGET - seg_base

    body = [row(ea) for ea in idautils.FuncItems(TARGET)]
    targets = memory_targets(TARGET)

    flag_refs = {}
    for ea in range(*FLAG_RANGE):
        refs = refs_to(ea)
        if refs:
            flag_refs[f"0x{ea:x}"] = refs

    immediates = {f"0x{v:x}": [] for v in MESSAGE_OFFSETS}
    near_ptr_code = []
    indirect_calls = []
    for ea in code_heads():
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea):
            continue
        mnem = insn.get_canon_mnem()
        for op in insn.ops:
            if op.type == ida_ua.o_void:
                break
            value = None
            if op.type == ida_ua.o_imm:
                value = op.value & 0xFFFF
            elif op.type == ida_ua.o_displ:
                value = op.addr & 0xFFFF
            if value in MESSAGE_OFFSETS:
                immediates[f"0x{value:x}"].append(
                    {"ea": f"0x{ea:x}", "func": idc.get_func_name(ea) or None,
                     "operand_type": op.type, "window": window(ea, 8, 4)})
            if op.type == ida_ua.o_imm and (op.value & 0xFFFF) == near_offset:
                near_ptr_code.append({"ea": f"0x{ea:x}", "row": row(ea)})
        if mnem in ("call", "callfi", "callni", "jmpni", "jmpfi"):
            first = insn.ops[0]
            if first.type in (ida_ua.o_reg, ida_ua.o_phrase, ida_ua.o_displ) or (
                    first.type == ida_ua.o_mem and mnem != "call"):
                indirect_calls.append(row(ea))
            elif first.type == ida_ua.o_mem and ida_bytes.get_byte(ea) == 0xFF:
                indirect_calls.append(row(ea))

    # 資料區裡等於 sub_14234 段內位移的 16 位元字（可能是函式指標表）。
    near_ptr_data = []
    for seg_ea in idautils.Segments():
        s = ida_segment.getseg(seg_ea)
        ea = s.start_ea
        while ea + 1 < s.end_ea:
            if not ida_bytes.is_code(ida_bytes.get_flags(ea)) and \
                    ida_bytes.get_word(ea) == near_offset:
                near_ptr_data.append({"ea": f"0x{ea:x}",
                                      "segment": ida_segment.get_segm_name(s),
                                      "next_word": f"0x{ida_bytes.get_word(ea + 2):x}"})
            ea += 1

    receipt = {
        "schema": "kol-ida-message-reach-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": original.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "evidence_level": "探索用 raw 掃描；語意待人工審查",
        "target": {"original_name": idc.get_func_name(TARGET), "ea": f"0x{TARGET:x}",
                   "segment": ida_segment.get_segm_name(seg),
                   "segment_base": f"0x{seg_base:x}", "near_offset": f"0x{near_offset:x}"},
        "target_body": body,
        "target_memory_operands": [
            {"ea": f"0x{t:x}", "name": idc.get_name(t) or None, "refs": refs_to(t)}
            for t in targets],
        "flag_range": [f"0x{FLAG_RANGE[0]:x}", f"0x{FLAG_RANGE[1]:x}"],
        "flag_refs": flag_refs,
        "message_offset_immediates": immediates,
        "near_offset_immediates_in_code": near_ptr_code,
        "near_offset_words_in_data": near_ptr_data,
        "indirect_calls": indirect_calls,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


main()
