#!/usr/bin/env python3
"""IDA Pro 9.4：戰鬥模組的地圖——進入點、被呼叫的函式與用到的 UI 記錄。

野外遭遇談不攏就 `sub_2A4E5:0x2A94A` 呼叫 `sub_1E3A8(0)`，那是戰鬥的進入點。
本探針先畫出範圍：`sub_1E3A8` 的大小、它呼叫誰、每一支的大小與呼叫者，
以及整個模組畫過哪些 KOLDATA 記錄——用來決定接下來逐支細讀的順序。

  idat -A -S"tools/ida_probe_koltown_combat.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
COMBAT_ENTRY = 0x1E3A8
TEXT_DRAW = 0x1A8FD


def disassembly(ea: int) -> str:
    rendered = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return rendered.split(" ;", 1)[0].rstrip()


def callees(start: int) -> list[int]:
    out = []
    for ea in idautils.FuncItems(start):
        text = disassembly(ea)
        if not text.startswith("call"):
            continue
        for ref in idautils.CodeRefsFrom(ea, False):
            fn = ida_funcs.get_func(ref)
            if fn is not None and fn.start_ea != start and fn.start_ea not in out:
                out.append(fn.start_ea)
    return out


def summary(start: int) -> dict:
    fn = ida_funcs.get_func(start)
    if fn is None:
        return {"start_ea": f"0x{start:x}", "error": "no function"}
    items = list(idautils.FuncItems(start))
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{fn.end_ea:x}",
        "instruction_count": len(items),
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
        "callees": [f"0x{x:x}" for x in callees(start)],
    }


def text_records(funcs: set[int]) -> list[dict]:
    out = []
    for ea in idautils.CodeRefsTo(TEXT_DRAW, False):
        fn = ida_funcs.get_func(ea)
        if fn is None or fn.start_ea not in funcs:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        i = items.index(ea)
        record, seen_push = None, False
        for j in range(i - 1, max(-1, i - 8), -1):
            text = disassembly(items[j])
            if text.startswith("push"):
                seen_push = True
                continue
            if seen_push and text.startswith("mov     ax,"):
                if idc.get_operand_type(items[j], 1) == idc.o_imm:
                    record = idc.get_operand_value(items[j], 1)
                break
        out.append({"call_ea": f"0x{ea:x}", "function": f"0x{fn.start_ea:x}",
                    "record": None if record is None else f"0x{record:x}"})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    # 從進入點做兩層廣度優先，足以看出模組邊界而不會把整支 CRT 拖進來。
    level1 = callees(COMBAT_ENTRY)
    level2 = []
    for ea in level1:
        for callee in callees(ea):
            if callee not in level1 and callee not in level2:
                level2.append(callee)
    known = {COMBAT_ENTRY, *level1, *level2}

    dump = []
    for ea in idautils.FuncItems(0x10000):
        text = disassembly(ea)
        if 0x10c80 <= ea <= 0x10d60:
            dump.append({"ea": f"0x{ea:x}", "disassembly": text})
    for ea in idautils.FuncItems(COMBAT_ENTRY):
        pass

    receipt = {
        "schema": "kol-ida-koltown-combat-map-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：戰鬥模組的範圍與記錄；尚未鎖定任何斷言。",
        "entry": summary(COMBAT_ENTRY),
        "level1": [summary(ea) for ea in level1],
        "level2": [summary(ea) for ea in level2],
        "text_records": text_records(known),
        "entry_dump": dump,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
