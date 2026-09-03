#!/usr/bin/env python3
"""IDA Pro 9.4：AWARD 訓練的主流程與「要幾點技能」那一格的文字輸入。

錨點：訓練的問句是 KOLDATA 記錄 0x3F／0x40／0x41／0x6C／0xD0，全部經 `sub_1A8FD`
畫出。本探針反查這些記錄的呼叫點、倒出所在函式，並找出畫游標（字元 0x60）的位置。

  idat -A -S"tools/ida_probe_koltown_train.py OUT.json INPUT.EXE" INPUT.i64
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
TEXT_DRAW = 0x1A8FD
CHAR_DRAW = 0x1C2DA
WANTED_RECORDS = (0x3E, 0x3F, 0x40, 0x41, 0x6B, 0x6C, 0x6D, 0x6E, 0x6F, 0x70, 0xD0)
DUMP_LIMIT = 4000


def disassembly(ea: int) -> str:
    rendered = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return rendered.split(" ;", 1)[0].rstrip()


def function_slice(start: int, limit: int = DUMP_LIMIT) -> dict:
    fn = ida_funcs.get_func(start)
    if fn is None:
        return {"start_ea": f"0x{start:x}", "error": "no function"}
    items = list(idautils.FuncItems(start))
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{fn.end_ea:x}",
        "instruction_count": len(items),
        "truncated": len(items) > limit,
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
        "instructions": [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
                         for ea in items[:limit]],
    }


def immediate_before(items: list[int], i: int) -> int | None:
    seen_push = False
    for j in range(i - 1, max(-1, i - 8), -1):
        text = disassembly(items[j])
        if text.startswith("push"):
            seen_push = True
            continue
        if seen_push and text.startswith("mov     ax,"):
            if idc.get_operand_type(items[j], 1) == idc.o_imm:
                return idc.get_operand_value(items[j], 1)
            return None
    return None


def call_sites(target: int) -> list[dict]:
    out = []
    for ea in idautils.CodeRefsTo(target, False):
        fn = ida_funcs.get_func(ea)
        if fn is None:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        if ea not in items:
            continue
        value = immediate_before(items, items.index(ea))
        out.append({"call_ea": f"0x{ea:x}", "function": f"0x{fn.start_ea:x}",
                    "argument": None if value is None else f"0x{value:x}"})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    text_sites = call_sites(TEXT_DRAW)
    char_sites = call_sites(CHAR_DRAW)
    wanted = {f"0x{r:x}" for r in WANTED_RECORDS}
    owners: list[int] = []
    for site in text_sites:
        if site["argument"] in wanted:
            ea = int(site["function"], 16)
            if ea not in owners:
                owners.append(ea)
    for site in char_sites:
        if site["argument"] == "0x60":
            ea = int(site["function"], 16)
            if ea not in owners:
                owners.append(ea)

    for extra in (0x1B5C8, 0x1CA6E, 0x1D01D):
        if extra not in owners:
            owners.append(extra)

    receipt = {
        "schema": "kol-ida-koltown-train-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：AWARD 訓練問句與數字輸入格的游標；尚未鎖定任何斷言。",
        "train_record_sites": [s for s in text_sites if s["argument"] in wanted],
        "caret_char_sites": [s for s in char_sites if s["argument"] == "0x60"],
        "dumped_functions": [function_slice(ea) for ea in owners],
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
