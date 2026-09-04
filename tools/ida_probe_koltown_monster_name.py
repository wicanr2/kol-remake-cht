#!/usr/bin/env python3
"""IDA Pro 9.4：戰場上的敵人是誰、名字從哪來。

原版執行段的訊息印出 `THE RUFFIAN`，而任務地點記錄寫的是 BINDERAK（RE-089）。
`|F`／`|G` 對敵人印的是 `sub_1A8FD(0, 0xB967)`——記錄號 0、文字庫 DGROUP+0xB967。
本探針倒出名字那一段（`sub_2687F` 與 `sub_265F3` 的 `|F` 分支），掃全庫找誰寫
DGROUP+0xB967 的名字緩衝區與 DGROUP+0xCD92 的怪物記錄區，並附上這兩塊的靜態內容。
輸出只有位址、反組譯文字與資料位元組。

  idat -A -S"tools/ida_probe_koltown_monster_name.py OUT.json INPUT.EXE" INPUT.i64
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
NAME_BUFFER = 0xB967       # `|F`／`|G` 印敵人名字時傳給 sub_1A8FD 的文字庫
MONSTER_RECORDS = 0xCD92   # 參戰敵人的怪物記錄，一筆 128 byte（sub_22296）
# 名字那一段：印名字的那一支，加上 5-bit 解碼與訊息輸出。
DUMP_FUNCS = (0x2687F, 0x2B0FF, 0x1A4FE, 0x2A4E5)
# `|F` 的分支落在 sub_265F3 中段，另外整段窗開出來對照。
WINDOWS = ((0x26714, 0x26800), (0x29440, 0x294D0),
           (0x1E3A8, 0x1E4A0), (0x2A620, 0x2A6A0), (0x2B010, 0x2B060))
# 這兩塊的位址在指令裡多半以立即值出現，逐條掃文字比 xref 可靠。
TRACKED_IMMEDIATES = (NAME_BUFFER, MONSTER_RECORDS)
# 名字緩衝區是某塊更大的緩衝的一部分：改掃整段範圍才看得到 base+offset 的存取。
SCAN_RANGES = ((0xB900, 0xBA00), (0xCD00, 0xCE00))
DUMP_BLOCKS = ((NAME_BUFFER - 0x20, 0x60), (MONSTER_RECORDS, 0x80),
               # sub_1A4FE 的第一個參數：0x1844（戰鬥）與 0x1896（另一處）。
               (0x1844, 0x40), (0x1896, 0x40))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def summary(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
        "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)}),
    }


def listing(start: int, end: int | None = None) -> list[dict]:
    if end is None:
        items = idautils.FuncItems(start)
    else:
        items = []
        ea = start
        while ea < end:
            items.append(ea)
            size = ida_bytes.get_item_size(ea)
            ea += size if size > 0 else 1
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)} for ea in items]


def immediate_users() -> dict:
    """逐條掃全庫，找出指令裡帶這兩個 DGROUP 位移的位置。"""
    out: dict[str, list[dict]] = {f"0x{v:04x}": [] for v in TRACKED_IMMEDIATES}
    absolute = {f"0x{v:04x}": DGROUP_EA + v for v in TRACKED_IMMEDIATES}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            for n in range(3):
                if idc.get_operand_type(ea, n) not in (idc.o_imm, idc.o_mem,
                                                       idc.o_displ, idc.o_phrase):
                    continue
                value = idc.get_operand_value(ea, n)
                for key, offset in ((k, v) for k, v in
                                    ((k, int(k, 16)) for k in out)):
                    if value == offset or value == absolute[key]:
                        out[key].append({
                            "func": f"0x{start:x}",
                            "func_name": idc.get_func_name(start),
                            "ea": f"0x{ea:x}",
                            "disassembly": disassembly(ea),
                        })
    return out


# 緩衝區多半以 `offset byte_39836` 的形式當參數傳出去，operand 掃描看不到，
# 逐條比對反組譯文字才抓得到。
TRACKED_NAMES = ("byte_39836", "byte_39837", "byte_39838", "byte_39826",
                 "word_3B262", "word_3B26A", "word_30C6C", "sub_1A4FE")


def name_users() -> dict:
    out: dict[str, list[dict]] = {n: [] for n in TRACKED_NAMES}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            for name in TRACKED_NAMES:
                if name in text:
                    out[name].append({"func": f"0x{start:x}", "ea": f"0x{ea:x}",
                                      "disassembly": text})
    return out


def range_users() -> dict:
    """掃全庫，列出 operand 落在指定 DGROUP 範圍內的指令（含絕對位址寫法）。"""
    out: dict[str, list[dict]] = {}
    for lo, hi in SCAN_RANGES:
        key = f"0x{lo:04x}-0x{hi:04x}"
        hits: list[dict] = []
        for start in idautils.Functions():
            for ea in idautils.FuncItems(start):
                for n in range(3):
                    if idc.get_operand_type(ea, n) not in (idc.o_imm, idc.o_mem,
                                                           idc.o_displ):
                        continue
                    value = idc.get_operand_value(ea, n)
                    for candidate in (value, value - DGROUP_EA):
                        if lo <= candidate < hi:
                            hits.append({
                                "func": f"0x{start:x}",
                                "ea": f"0x{ea:x}",
                                "dgroup_offset": f"0x{candidate:04x}",
                                "disassembly": disassembly(ea),
                            })
                            break
        out[key] = hits
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
        "schema": "kol-ida-koltown-monster-name-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：敵人名字緩衝區與怪物記錄的寫入者；尚未鎖定任何斷言。",
        "functions": [summary(ea) for ea in DUMP_FUNCS],
        "listings": {f"0x{ea:x}": listing(ea) for ea in DUMP_FUNCS},
        "windows": {f"0x{a:x}-0x{b:x}": listing(a, b) for a, b in WINDOWS},
        "immediate_users": immediate_users(),
        "range_users": range_users(),
        "name_users": name_users(),
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                    or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
