"""IDA Pro 9.4：匯出 TITLE.EXE `_main` 的標題序列腳本。

標題序列在原版是一段線性程式碼，不是資料表：`_main` 依序呼叫載圖、疊圖與等待常式。
畫面觀察只能得到牆鐘時間與去重畫格，重現不了原版的實際順序與參數，因此這裡直接把
`_main` 及其所有直接被呼叫者匯出，並把每個呼叫點之前的 push 立即值一併記錄，
讓序列可以逐呼叫重建。

  idat -A -S"tools/ida_probe_title_sequence.py PRIVATE.json SUMMARY.json TITLE.EXE" TITLE.EXE
"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import ida_pro
import idautils
import idc

EXPECTED_SHA256 = "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6"
MAIN_EA = 0x10010


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
    """依位址順序列出 `_main` 的呼叫點，並把引數解析成可讀值。

    Borland 小模式把引數逐一 push 上堆疊，常見樣式是 `mov ax, IMM` 後 `push ax`，
    或直接 `push word_XXXX`。這裡對每個呼叫點往回掃描，收集自上一個呼叫（或函式起點）
    以來的 push，並用最近一次寫入同一暫存器的立即值把 `push reg` 解析成數值。
    解析不出來的保留原始 operand 文字，不猜。
    """
    items = list(idautils.FuncItems(start))
    script: list[dict] = []
    pending: list[dict] = []
    reg_imm: dict[str, int] = {}
    for ea in items:
        mnem = idc.print_insn_mnem(ea)
        text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
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
                "disassembly": text,
            })
            pending = []
            reg_imm.clear()
            continue
        # 任何其他寫入暫存器的指令都讓既有的立即值失效，避免解析出錯誤的引數。
        if mnem and idc.get_operand_type(ea, 0) == idc.o_reg:
            reg_imm.pop(idc.print_operand(ea, 0), None)
    return script


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    main_function = ida_funcs.get_func(MAIN_EA)
    if main_function is None or main_function.start_ea != MAIN_EA:
        raise RuntimeError("_main boundary not found")
    targets = call_targets(MAIN_EA)
    functions = [export_function(MAIN_EA)] + [export_function(ea) for ea in targets]
    script = call_script(MAIN_EA)

    private = {
        "schema": "kol-ida-title-sequence-private-v1",
        "ida_version": "9.4",
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for the named loaded MZ image",
        "main_ea": f"0x{MAIN_EA:x}",
        "functions": functions,
        "call_script": script,
    }
    # 安全摘要不帶原版 bytes 或反組譯文字，只留位址、目標與解析後的引數語意。
    safe_script = [{k: v for k, v in call.items() if k != "disassembly"}
                   for call in script]
    summary = {
        "schema": "kol-ida-title-sequence-summary-v1",
        "ida_version": "9.4",
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for the named loaded MZ image",
        "main_ea": f"0x{MAIN_EA:x}",
        "main_end_ea": f"0x{main_function.end_ea:x}",
        "main_instruction_count": len(functions[0]["instructions"]),
        "direct_callees": [
            {"original_name": f["original_name"], "start_ea": f["start_ea"],
             "end_ea": f["end_ea"], "instruction_count": len(f["instructions"])}
            for f in functions[1:]
        ],
        "call_script": safe_script,
    }
    private_path.write_text(json.dumps(private, ensure_ascii=False, indent=2) + "\n")
    summary_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n")
    ida_pro.qexit(0)


main()
