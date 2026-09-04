#!/usr/bin/env python3
"""IDA Pro 9.4：任務旗標怎麼標成完成。

角色記錄 `0x1E2` 起的六個 byte 是 24 個任務地點的旗標，一個 byte 塞四個 2-bit
（RE-079 只解了「怎麼讀出來判斷有沒有接這個任務」）。對那個位移做範圍掃描找到
七個使用者，其中 `sub_28F5D` 正是打贏那一刻呼叫的（`0x211CB`）。

一次倒完那幾支與相關的視窗。**不要改一個常數重跑一次**（見 `run_ida_probe.sh`）。

  idat -A -S"tools/ida_probe_koltown_quest_flag.py OUT.json INPUT.EXE" INPUT.i64
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
# 打贏時呼叫的那一支、城鎮那兩支、任務地點那一支。
FUNCS = (0x28F5D, 0x14564, 0x167D9, 0x2B0FF)
WINDOWS = (
    # sub_11A93 裡兩處讀旗標的段落。
    (0x11C60, 0x11CE0),
)
# 除數表：任務編號 & 3 → 1／4／16／64（RE-079）。
DUMP_BLOCKS = ((0x007C, 4),)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def line(ea: int) -> dict:
    return {"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append(line(ea))
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

    receipt = {
        "schema": "kol-ida-koltown-quest-flag-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：任務旗標的讀寫；尚未鎖定任何斷言。",
        "funcs": {f"0x{ea:x}": {
            "name": idc.get_func_name(ea),
            "instruction_count": len(list(idautils.FuncItems(ea))),
            "callers": sorted({f"0x{c:x}" for c in idautils.CodeRefsTo(ea, False)}),
            "listing": [line(i) for i in idautils.FuncItems(ea)],
        } for ea in FUNCS if ida_funcs.get_func(ea) is not None},
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
