#!/usr/bin/env python3
"""IDA 9.4：傾印 KOLTOWN 指定函式的全部指令（原始名稱、位址、bytes、反組譯）。

用來讀單一函式的分支細節；函式清單寫在 FUNCTIONS。輸出含原版指令 bytes，只放本機，
不得加入 Git。

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
FUNCTIONS = (0x24131,)


def main() -> None:
    output, original = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(original.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    out = {"schema": "kol-ida-function-dump-v1", "ida_version": idaapi.get_kernel_version(),
           "input_sha256": digest,
           "address_space": "IDA linear effective address for reconstructed unpacked MZ",
           "functions": {}}
    for start in FUNCTIONS:
        func = ida_funcs.get_func(start)
        rows = []
        for ea in idautils.FuncItems(func.start_ea):
            size = max(1, ida_bytes.get_item_size(ea))
            rows.append({"ea": f"0x{ea:x}", "bytes": (ida_bytes.get_bytes(ea, min(size, 12)) or b"").hex(),
                         "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")})
        out["functions"][f"0x{start:x}"] = {"original_name": idc.get_func_name(start),
                                            "end_ea": f"0x{func.end_ea:x}", "rows": rows}
    output.write_text(json.dumps(out, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    idc.qexit(0)


main()
