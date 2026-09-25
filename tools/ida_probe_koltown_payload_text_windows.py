#!/usr/bin/env python3
"""IDA 9.4：補看建築 payload 文字讀取點的長前文與兩支印出函式的呼叫者（RE-140）。

`ida_probe_koltown_payload_text_readers.py` 找到以 0x1CF4 加變數位移解碼文字的
呼叫點；這支取其中位移來源不明者的較長前文，並列出 `sub_14189`、`sub_2A46E`
的全部直接呼叫者，用來判斷訊息 8／9（payload 0x280..0x2FF）是否可能被讀出。
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
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
LONG_WINDOWS = (0x12D4D, 0x138BC, 0x176A1, 0x2A4AF)
CALLEES = (0x14189, 0x2A46E)
FULL_FUNCTIONS = (0x14189, 0x2A46E)


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
    receipt = {
        "schema": "kol-ida-payload-text-windows-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": original.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "evidence_level": "探索用 raw 前文；語意待人工審查",
        "long_windows": {f"0x{ea:x}": {"func": idc.get_func_name(ea) or None,
                                        "rows": window(ea, 60, 2)} for ea in LONG_WINDOWS},
        "callers": {},
        "functions": {},
    }
    for callee in CALLEES:
        rows = []
        for ea in sorted(set(idautils.CodeRefsTo(callee, False))):
            rows.append({"ea": f"0x{ea:x}", "func": idc.get_func_name(ea) or None,
                         "window": window(ea, 10, 1)})
        receipt["callers"][f"0x{callee:x}"] = {
            "original_name": idc.get_func_name(callee),
            "data_refs": [f"0x{ea:x}" for ea in sorted(set(idautils.DataRefsTo(callee)))],
            "code_refs": rows}
    for start in FULL_FUNCTIONS:
        func = ida_funcs.get_func(start)
        receipt["functions"][f"0x{start:x}"] = [row(ea) for ea in idautils.FuncItems(func.start_ea)] \
            if func is not None else None
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


main()
