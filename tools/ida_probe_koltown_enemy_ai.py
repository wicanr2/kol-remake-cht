#!/usr/bin/env python3
"""IDA Pro 9.4：敵人 AI 的主幹。

RE-081 解出決策樹的形狀，但撤退（+0x29／+0x2A）、反制掃描 `sub_253E7`、
側翼 `sub_25653` 與射擊那一條（指令碼 0x23）都還沒對到逐條的碼。本探針倒出
`sub_2374C` 全文與它呼叫的每一支，另加決策樹那一段的視窗。

  idat -A -S"tools/ida_probe_koltown_enemy_ai.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x2374C,)
MAX_INSNS = 1400
WINDOWS = ((0x23C1A, 0x24060),)
# AI 讀的幾張表：`0x10D8` 是相對格差換方向，其餘是加權輪盤（RE-081）。
DUMP_BLOCKS = ((0x10D8, 0x10), (0x0B6C, 0x20), (0x0FC8, 0x10))


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

    receipt = {
        "schema": "kol-ida-koltown-enemy-ai-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：敵人 AI 的主幹與它的 callee；尚未鎖定任何斷言。",
        "functions": {f"0x{s:x}": {"name": idc.get_func_name(s),
                                   "instruction_count": len(list(idautils.FuncItems(s))),
                                   "callers": sorted({f"0x{e:x}"
                                                      for e in idautils.CodeRefsTo(s, False)})}
                      for s in sorted(set(wanted))
                      if ida_funcs.get_func(s) is not None},
        "listings": listings,
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
