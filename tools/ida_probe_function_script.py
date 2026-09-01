"""IDA Pro 9.4：把指定函式與其直接被呼叫者匯出成可讀的呼叫腳本。

用途是把「一段線性程式碼就是流程本身」的情形讀成資料：匯出根函式的每一個呼叫點，
並把 Borland 小模式的 `mov reg, IMM` + `push reg` 樣式解析成引數值。安全摘要不帶
原版 bytes 或反組譯文字，只留位址、目標與解析後的引數。

  idat -A -S"tools/ida_probe_function_script.py PRIVATE.json SUMMARY.json IN.EXE \\
      <root_ea_hex> [extra_ea_hex ...]" IN.EXE
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import ida_pro
import idautils
import idc


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "mnemonic": idc.print_insn_mnem(ea),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
    }


def export_function(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None:
        raise RuntimeError(f"no function at 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def call_targets(start: int) -> list[int]:
    out = []
    for ea in idautils.FuncItems(start):
        if idc.print_insn_mnem(ea) != "call":
            continue
        for target in idautils.CodeRefsFrom(ea, False):
            function = ida_funcs.get_func(target)
            if function is not None and function.start_ea == target:
                out.append(target)
    return sorted(set(out))


def call_script(start: int) -> list[dict]:
    script: list[dict] = []
    pending: list[dict] = []
    reg_imm: dict[str, int] = {}
    for ea in idautils.FuncItems(start):
        mnem = idc.print_insn_mnem(ea)
        if mnem == "mov" and idc.get_operand_type(ea, 1) == idc.o_imm:
            reg_imm[idc.print_operand(ea, 0)] = idc.get_operand_value(ea, 1)
            continue
        if mnem == "push":
            operand = idc.print_operand(ea, 0)
            if idc.get_operand_type(ea, 0) == idc.o_imm:
                pending.append({"kind": "imm", "value": idc.get_operand_value(ea, 0)})
            elif operand in reg_imm:
                pending.append({"kind": "imm", "value": reg_imm[operand],
                                "via_register": operand})
            else:
                pending.append({"kind": "operand", "text": operand})
            continue
        if mnem == "call":
            targets = list(idautils.CodeRefsFrom(ea, False))
            script.append({
                "call_ea": f"0x{ea:x}",
                "target_ea": f"0x{targets[0]:x}" if targets else None,
                "target_name": idc.get_func_name(targets[0]) if targets else None,
                "arguments": list(reversed(pending)),
            })
            pending = []
            reg_imm.clear()
            continue
        if mnem and idc.get_operand_type(ea, 0) == idc.o_reg:
            reg_imm.pop(idc.print_operand(ea, 0), None)
    return script


def main() -> None:
    if len(idc.ARGV) < 5:
        raise RuntimeError("usage: script.py PRIVATE SUMMARY INPUT ROOT_EA [EXTRA...]")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    root = int(idc.ARGV[4], 16)
    extra_eas = [int(a, 16) for a in idc.ARGV[5:]]
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    ida_auto.auto_wait()

    function = ida_funcs.get_func(root)
    if function is None or function.start_ea != root:
        raise RuntimeError(f"root function boundary not found at 0x{root:x}")
    targets = call_targets(root)
    extra = [ea for ea in extra_eas if ea not in targets and ea != root]
    functions = ([export_function(root)]
                 + [export_function(ea) for ea in targets]
                 + [export_function(ea) for ea in extra])
    script = call_script(root)

    common = {
        "ida_version": "9.4",
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for the named loaded MZ image",
        "root_ea": f"0x{root:x}",
        "root_end_ea": f"0x{function.end_ea:x}",
        "root_original_name": idc.get_func_name(root),
    }
    private_path.write_text(json.dumps(
        {"schema": "kol-ida-function-script-private-v1", **common,
         "functions": functions, "call_script": script},
        ensure_ascii=False, indent=2) + "\n")
    summary_path.write_text(json.dumps(
        {"schema": "kol-ida-function-script-summary-v1", **common,
         "root_instruction_count": len(functions[0]["instructions"]),
         "functions": [{"original_name": f["original_name"], "start_ea": f["start_ea"],
                        "end_ea": f["end_ea"],
                        "instruction_count": len(f["instructions"])}
                       for f in functions[1:]],
         "call_script": script},
        ensure_ascii=False, indent=2) + "\n")
    ida_pro.qexit(0)


main()
