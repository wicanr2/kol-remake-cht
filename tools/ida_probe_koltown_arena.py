#!/usr/bin/env python3
"""IDA Pro 9.4：競技場。

`word_345DE`（誰要升位階）唯一寫進真值的地方是 `sub_2AAA8:0x2AFE6`（RE-103），
所以升級掛在競技場那一側。這一支同時是競技場戰鬥的入口（`0x2B0D5` 呼叫
`sub_1E3A8`）與賭注那張表（`[0x1B2A]`）的來源。

  idat -A -S"tools/ida_probe_koltown_arena.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x2AAA8,)
MAX_INSNS = 1600
DGROUP_EA = 0x2DED0
# 賭注那張表與 `sub_14371`（實得位階，RE-074）用得到的區塊。
DUMP_BLOCKS = ((0x1B2A, 0x18),)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


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

    receipt = {
        "schema": "kol-ida-koltown-arena-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：競技場與升級；尚未鎖定任何斷言。",
        "functions": {f"0x{s:x}": {"name": idc.get_func_name(s),
                                   "instruction_count": len(list(idautils.FuncItems(s))),
                                   "callers": sorted({f"0x{e:x}"
                                                      for e in idautils.CodeRefsTo(s, False)})}
                      for s in sorted(set(wanted))
                      if ida_funcs.get_func(s) is not None},
        "listings": listings,
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
