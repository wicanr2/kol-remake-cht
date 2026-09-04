#!/usr/bin/env python3
"""IDA Pro 9.4：投擲／射擊那一手，以及任務旗標的另一個呼叫點。

`sub_26D76(0)` 是投擲與射擊（`sub_26D76(1)` 是近戰，RE-084／RE-085 已解）。
本探針倒出它的全文與它呼叫的每一支，另加 `0x21100`…`0x211A0` 這段——
`sub_28F5D`（任務標成完成，RE-101）有兩個呼叫點，`0x211CB` 是打贏那一條，
`0x21176` 的上下文還沒看過——記錄 0x109 是 `Do you want to flee?`，所以那一條在
逃走上。視窗一次涵蓋 `0x20F00`…`0x21800`（逃走、打贏結算、競技場賭金與升級）。

  idat -A -S"tools/ida_probe_koltown_throw.py OUT.json INPUT.EXE" INPUT.i64
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
DGROUP_EA = 0x2DED0
FUNCS = (0x26D76, 0x1AAE8)
# 戰鬥收尾整段：逃走的問句、打贏的結算、競技場的賭金與升級。
WINDOWS = ((0x20F00, 0x21800),)
MAX_INSNS = 1000


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
        items = list(idautils.FuncItems(start))
        if len(items) > MAX_INSNS:
            continue
        listings[f"0x{start:x}"] = listing(start)

    receipt = {
        "schema": "kol-ida-koltown-throw-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：投擲／射擊那一手與任務旗標的另一個呼叫點；尚未鎖定任何斷言。",
        "functions": {f"0x{s:x}": {"name": idc.get_func_name(s),
                                   "instruction_count": len(list(idautils.FuncItems(s))),
                                   "callers": sorted({f"0x{e:x}"
                                                      for e in idautils.CodeRefsTo(s, False)})}
                      for s in sorted(set(wanted))
                      if ida_funcs.get_func(s) is not None},
        "listings": listings,
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
