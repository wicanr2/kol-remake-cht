#!/usr/bin/env python3
"""IDA Pro 9.4：結局的流程與觸發。

RE-017 定位到 `sub_2B4A7` 連著呼叫 `sub_2B80A`（載 ENDPICS）七次，圖號 0…6，
所以那一支就是結局。本探針倒出它的全文、它呼叫的每一支，以及**誰呼叫它**——
觸發條件在呼叫點那一側。

  idat -A -S"tools/ida_probe_koltown_ending.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x2B4A7,)
MAX_INSNS = 1200
# 呼叫點前後各看這麼多 byte，觸發條件在那裡。
CALLER_SPAN = 0xC0


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def callees(start: int) -> list[int]:
    out = []
    for ea in idautils.FuncItems(start):
        if idc.print_insn_mnem(ea) != "call":
            continue
        for t in idautils.CodeRefsFrom(ea, False):
            f = ida_funcs.get_func(t)
            if f is not None and f.start_ea == t:
                out.append(t)
    return sorted(set(out))


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    wanted = list(FUNCS)
    for start in FUNCS:
        wanted.extend(callees(start))
    listings = {}
    for start in sorted(set(wanted)):
        f = ida_funcs.get_func(start)
        if f is None or f.start_ea != start:
            continue
        if len(list(idautils.FuncItems(start))) > MAX_INSNS:
            continue
        listings[f"0x{start:x}"] = listing(start)

    callers = {}
    for start in FUNCS:
        for ea in idautils.CodeRefsTo(start, False):
            callers[f"0x{ea:x}"] = window(ea - CALLER_SPAN, ea + 0x20)

    receipt = {
        "schema": "kol-ida-koltown-ending-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：結局的流程與呼叫點；尚未鎖定任何斷言。",
        "functions": {f"0x{s:x}": {"name": idc.get_func_name(s),
                                   "instruction_count": len(list(idautils.FuncItems(s))),
                                   "callers": sorted({f"0x{e:x}"
                                                      for e in idautils.CodeRefsTo(s, False)})}
                      for s in sorted(set(wanted))
                      if ida_funcs.get_func(s) is not None},
        "listings": listings,
        "caller_windows": callers,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
