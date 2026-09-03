#!/usr/bin/env python3
"""IDA Pro 9.4：野外的時間推進、紮營與遭遇。

錨點（RE-069／RE-070）：`word_2DF40` 是時刻、`word_2DF42` 是時段，
`sub_297F2` 是野外主迴圈，payload 3 的 `0x323`／`0x33B` 是 24 個特殊地點。

本探針掃出所有讀寫 `word_2DF40`／`word_2DF44` 的指令，並倒出時間與遭遇相關的函式。

  idat -A -S"tools/ida_probe_koltown_wilderness.py OUT.json INPUT.EXE" INPUT.i64
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
DUMP_LIMIT = 4000

GLOBALS = {
    "clock_word_2DF40": 0x2DF40,
    "daypart_word_2DF42": 0x2DF42,
    "month_word_2FBB4": 0x2FBB4,
    "day_word_2FA2E": 0x2FA2E,
    "year_word_30A8E": 0x30A8E,
    "terrain_cost_word_2FA2C": 0x2FA2C,
    "weather_word_2DF5A": 0x2DF5A,
    "mount_word_39A54": 0x39A54,
    "encounter_x_byte_2FEE4": 0x2FEE4,
    "encounter_y_byte_2FEE5": 0x2FEE5,
    "party_x_word_2FB50": 0x2FB50,
}

# DGROUP 表：月名（12 × 4 byte）、移動成本、每月天數、時段門檻、天候兩張、每日營養。
DGROUP_DUMPS = ((0x063A, 48), (0x12B8, 20), (0x12C0, 12), (0x12CC, 10),
                (0x12D6, 12), (0x131A, 12), (0x1326, 12), (0x0070, 16),
                (0x12A0, 24), (0x1300, 96))

DUMP_FUNCTIONS = (0x2A4E5, 0x2A337, 0x2B0FF)


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


def global_users(addr: int) -> list[dict]:
    offset = addr - DGROUP_EA
    # 只認「用名字」與「方括號裡的位移」兩種寫法。單純的立即數（`add ax, 70h`）
    # 不算——那是別的東西剛好同值，全掃下來噪音比訊號多。
    needles = (f"_{addr:X}", f"_{addr:x}", f"+{offset:X}h]", f"[{offset:X}h]")
    out = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            text = disassembly(ea)
            if any(n in text for n in needles):
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

    globals_out = {name: global_users(addr) for name, addr in GLOBALS.items()}
    # 日期框的兩行小字畫在 byte 欄 3／列 0x3B 與 byte 欄 2／列 0x41（RE-068）。
    date_sites = []
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            text = disassembly(ea)
            if "word_2DF16, 3Bh" in text or "word_2DF16, 41h" in text:
                date_sites.append({"ea": f"0x{ea:x}", "function": f"0x{fn_ea:x}",
                                   "disassembly": text})
    owners = list(DUMP_FUNCTIONS)
    for uses in globals_out.values():
        for use in uses:
            ea = int(use["function"], 16)
            if ea not in owners:
                owners.append(ea)

    receipt = {
        "schema": "kol-ida-koltown-wilderness-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：野外時間、紮營與遭遇；尚未鎖定任何斷言。",
        "globals": globals_out,
        "date_box_sites": date_sites,
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
        "dumped_function_list": [f"0x{ea:x}" for ea in owners],
        "dumped_functions": [function_slice(ea) for ea in DUMP_FUNCTIONS],
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
