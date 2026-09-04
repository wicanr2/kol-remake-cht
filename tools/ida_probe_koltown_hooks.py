#!/usr/bin/env python3
"""IDA Pro 9.4：任務回報與競技場的入口。

RE-107 找到回報那一段在 `sub_11A93:0x11C28`，RE-108 找到競技場是
`sub_2AAA8`（唯一呼叫點 `sub_29004:0x29442`）。要接進 remake 得先知道**什麼時候
走到那裡**，所以這一支倒出兩處往上的路：回報那一段前面那一大塊，
以及 `sub_29004` 的呼叫點與那一段。

  idat -A -S"tools/ida_probe_koltown_hooks.py OUT.json INPUT.EXE" INPUT.i64
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
WINDOWS = ((0x11B20, 0x11C40), (0x29380, 0x29460))
# 這幾支要知道誰呼叫它們。
CALLERS_OF = (0x29004, 0x11A93)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    callers = {}
    for start in CALLERS_OF:
        rows = []
        for ea in idautils.CodeRefsTo(start, False):
            f = ida_funcs.get_func(ea)
            rows.append({"ea": f"0x{ea:x}",
                         "func": idc.get_func_name(f.start_ea) if f else "",
                         "disassembly": disassembly(ea)})
        callers[f"0x{start:x}"] = sorted(rows, key=lambda r: r["ea"])

    receipt = {
        "schema": "kol-ida-koltown-hooks-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：任務回報與競技場的入口；尚未鎖定任何斷言。",
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "callers": callers,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
