#!/usr/bin/env python3
"""IDA Pro 9.4：彈道上一格的判定 `sub_25F8A`。

`sub_25D7A`（RE-102）是 Bresenham：從行動者往記住的目標格逐格走，每一格呼叫
`sub_25F8A(模式, 第幾步)`，回非零就停下，障礙種類寫在 `word_30A74`
（記錄 0x121 之後接 0x122 `tree!` 或 0x123 `wall!`）。本探針倒出那一支與它
呼叫的每一支，另加它讀到的 DGROUP 區塊。

  idat -A -S"tools/ida_probe_koltown_missile.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x25F8A,)
MAX_INSNS = 1000
# 戰場地形的磚塊分類表候選（RE-099 的地形是逐格擲出來的，判定表還沒對上）。
DUMP_BLOCKS = ((0x0B6C, 0x20), (0x00D4, 0x10), (0x109C, 0x20), (0x1200, 0x40))


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
        "schema": "kol-ida-koltown-missile-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：彈道上一格的判定；尚未鎖定任何斷言。",
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
