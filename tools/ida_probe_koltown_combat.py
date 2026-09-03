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
import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
COMBAT_ENTRY = 0x1E3A8
# 想逐條讀的函式；探針每次只倒這幾支，免得輸出爆掉。
DUMP_FUNCTIONS = (0x26AC2, 0x24FD3, 0x277F4, 0x24BB6, 0x24C35, 0x268D4, 0x26D3C)
DGROUP_EA = 0x2DED0
# 要原樣倒出來的 DGROUP 區塊：(位移, 位元組數)。
DGROUP_DUMPS = (
    (0x10CE, 16),   # 磚塊 → 擋視線／擋路的位元圖（`sub_25F8A:0x26011`）
    (0x1312, 8),    # 位元遮罩 80 40 20 10 08 04 02 01
    (0x00D4, 16),   # 傷害骰面數表（`sub_26D76:0x27502`）
    (0x109C, 16),   # 拳腳可用性遮罩（`sub_26D76:0x27532`）
    (0x110A, 128),  # 攻擊選項修正（一組 14 byte，`sub_26D76:0x2700B`）
    (0x11B2, 128),  # 招架訊息表（一列 11 byte，`sub_26D76:0x273D1`）
    (0x1200, 16),   # 武器攻擊式的傷害加成（`sub_26D76:0x27587`）
    (0x0FD8, 32),   # 方向樣板：X 位移陣列的指標表
    (0x0FE6, 32),   # 方向樣板：Y 位移陣列的指標表
    (0x0B76, 16),   # 方向 → 視窗格（`sub_1E3A8:0x1F2E7`）
    (0x0B60, 16),   # 一般遭遇的隊伍站位
    (0x0B8C, 72),   # 敵人陣型 Y
    (0x0BD4, 72),   # 敵人陣型 X
)
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

    owners_of = {}
    for fn_ea in idautils.Functions():
        for ea in idautils.FuncItems(fn_ea):
            text = disassembly(ea)
            for name in ("word_30A6A", "word_30C8A", "6712h"):
                if name in text:
                    owners_of.setdefault(name, {}).setdefault(f"0x{fn_ea:x}", 0)
                    owners_of[name][f"0x{fn_ea:x}"] += 1

    dump = []
    for start in DUMP_FUNCTIONS:
        for ea in idautils.FuncItems(start):
            dump.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})

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
        "global_owners": owners_of,
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
        "entry_dump": dump,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
