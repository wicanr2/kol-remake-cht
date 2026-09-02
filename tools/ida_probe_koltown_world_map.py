#!/usr/bin/env python3
"""IDA Pro 9.4：城鎮走到邊界之後怎麼切到野外地圖。

已知的錨點：`sub_10000` 在 `0x1068d` 用 `town%d` 組出地圖檔名，`sub_15920` 同時用到
`town%d.0`、`twn%d%s.0`、`twn%d%s.1` 與 `town%d`（像是換圖／換城的那一支）。
本探針把這兩支完整倒出來，並列出提供城號的那個全域的所有讀寫點。

  idat -A -S"tools/ida_probe_koltown_world_map.py OUT.json INPUT.EXE" INPUT.i64
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

# 完整倒出的函式：城鎮主迴圈與疑似換圖路徑。
DUMP_FUNCTIONS = (0x29987, 0x29B40, 0x2A18E, 0x2A01C)
DUMP_LIMIT = 20000

GLOBALS = {
    "camera_x_word_2FB50": 0x2FB50,
    "camera_y_word_2FB56": 0x2FB56,
    "marker_col_word_2DF36": 0x2DF36,
    "marker_row_word_2DF38": 0x2DF38,
    "clamp_word_2DF66": 0x2DF66,
    "clamp_word_2DF68": 0x2DF68,
    "clamp_word_2DF6A": 0x2DF6A,
    "clamp_word_2DF6C": 0x2DF6C,
    "entry_point_word_2DF62": 0x2DF62,
    "town_number_word_2DF2C": 0x2DF2C,
    "edge_flag_word_2DF60": 0x2DF60,
    "icon_choice_word_3AC5A": 0x3AC5A,
    "icon_count_word_2DF3C": 0x2DF3C,
    "icon_cursor_word_2DF3A": 0x2DF3A,
    "building_index_word_2FBC2": 0x2FBC2,
}

# `town%d` 的四個 DGROUP 位移；誰載入它們就是誰在組地圖檔名。
TOWN_FORMAT_OFFSETS = {0x168C, 0x16C8, 0x1794, 0x1808}


def disassembly(ea: int) -> str:
    rendered = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return rendered.split(" ;", 1)[0].rstrip()


def call_sites(target: int, window: int = 14) -> list[dict]:
    """列出所有呼叫點，並附上呼叫前 window 條指令——引數是 push 上去的。"""
    out = []
    for ea in idautils.CodeRefsTo(target, False):
        fn = ida_funcs.get_func(ea)
        if fn is None:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        if ea not in items:
            continue
        i = items.index(ea)
        lo = max(0, i - window)
        out.append({
            "call_ea": f"0x{ea:x}",
            "function": f"0x{fn.start_ea:x}",
            "context": [{"ea": f"0x{items[j]:x}",
                         "disassembly": disassembly(items[j])}
                        for j in range(lo, i + 1)],
        })
    return out


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


def global_users(addr: int) -> list[dict]:
    """16-bit real mode 下 DGROUP 全域是用位移定址，xref 不可靠——直接掃指令文字。"""
    offset = addr - DGROUP_EA
    needles = (f"{offset:X}h", f"_{addr:X}", f"_{addr:x}")
    out = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            text = disassembly(ea)
            if any(n in text for n in needles):
                out.append({"ea": f"0x{ea:x}", "function": f"0x{fn_ea:x}",
                            "disassembly": text})
    return out


def town_number_sources() -> list[dict]:
    """找出載入 `town%d` 位移的指令，並附前後 10 條——城號就在附近被 push。"""
    out = []
    for fn_ea in idautils.Functions():
        items = list(idautils.FuncItems(fn_ea))
        for i, ea in enumerate(items):
            hit = False
            for op in range(2):
                if idc.get_operand_type(ea, op) == idc.o_imm and \
                        idc.get_operand_value(ea, op) in TOWN_FORMAT_OFFSETS:
                    hit = True
            if not hit:
                continue
            lo, hi = max(0, i - 10), min(len(items), i + 6)
            out.append({
                "function": f"0x{fn_ea:x}", "ea": f"0x{ea:x}",
                "context": [{"ea": f"0x{items[j]:x}",
                             "disassembly": disassembly(items[j])}
                            for j in range(lo, hi)],
            })
    return out


def ffff_stores() -> list[dict]:
    """所有把 0FFFFh 寫進記憶體的指令——城鎮邊界那條路徑的線索。"""
    out = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            text = disassembly(ea)
            if "0FFFFh" in text and text.startswith("mov"):
                out.append({"ea": f"0x{ea:x}", "function": f"0x{fn_ea:x}",
                            "disassembly": text})
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
        "schema": "kol-ida-koltown-world-map-explore-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：城鎮主迴圈與換圖路徑的完整反組譯；尚未鎖定任何斷言。",
        "town_number_sources": town_number_sources(),
        "message_draw_calls": call_sites(0x14234, 26),
        "icon_push_calls": call_sites(0x11703, 8),
        "ffff_stores": ffff_stores(),
        "globals": {name: global_users(addr) for name, addr in GLOBALS.items()},
        "dumped_functions": [function_slice(ea) for ea in DUMP_FUNCTIONS],
        "dgroup_bit_mask_table_0x5E": (ida_bytes.get_bytes(DGROUP_EA + 0x5E, 16) or b"").hex(),
        "dgroup_town0_literals": {
            f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, 16) or b"").hex()
            for off in (0x1844, 0x1868, 0x186E, 0x1884, 0x1890, 0x1896, 0x189C, 0x18A2)},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
