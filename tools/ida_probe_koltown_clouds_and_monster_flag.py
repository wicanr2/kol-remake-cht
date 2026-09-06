#!/usr/bin/env python3
"""IDA Pro 9.4：雲層動畫（`sub_2A18E`）與怪物記錄 `+0x10` 的低位，一次問完。

兩件事都卡在「靜態讀不到」：

1. **雲層動畫**（`sub_2A18E`）：野外的天候圖層畫得出來，但動畫的版面沒定案
   （`WORKLIST.md`「還沒接進遊戲的」）。要的是完整反組譯、它參照的每一個
   DGROUP 位址，以及誰呼叫它。
2. **怪物記錄 `+0x10` 的 bit0**：bit7 是「碎成一地」那一句（RE-131），
   而 CYCLOPS／CLIFF TROLL／SLEDGE／MIST GRUB 的 bit0 是 1，語意未解。
   怪物記錄走 `[reg + 位移]` 這種相對定址，**xref 一個都看不到**——所以改成
   掃全庫：每一處「讀某個暫存器 `+10h` 的 byte」都列出來，連前後文。

⚠ **一支探針一次問完。** 對同一個 `.i64` 連續開關會弄壞資料庫。

  idat -A -S"tools/ida_probe_koltown_clouds_and_monster_flag.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
import re
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

CLOUD_FN = 0x2A18E
# 「碎成一地」那一句所在的函式（bit7 的讀取點，RE-131）。
SHATTER_EA = 0x20D9A
# 怪物記錄的技能讀取（RE-133 的 `sub_26C5F(0x17)`），拿來認記錄基底。
MONSTER_FIELD_FN = 0x26C5F

# 「讀某個暫存器 +10h 的 byte」：mov al, [bx+10h] / [si+10h] / [di+10h] / [bp+10h]
PLUS10 = re.compile(r"\[\s*(bx|si|di|bp)\s*\+\s*10h\s*\]", re.I)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    return [{"ea": f"0x{ea:x}",
             "bytes": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
             "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {"original_name": idc.get_func_name(start),
            "start_ea": f"0x{start:x}", "end_ea": f"0x{f.end_ea:x}",
            "instruction_count": len(list(idautils.FuncItems(start))),
            "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)})}


def globals_used(start: int) -> list[dict]:
    seen: dict[int, dict] = {}
    for ea in idautils.FuncItems(start):
        for ref in idautils.DataRefsFrom(ea):
            if ref in seen:
                seen[ref]["refs"].append(f"0x{ea:x}")
                continue
            seen[ref] = {"ea": f"0x{ref:x}",
                         "ds_offset": f"0x{ref - DGROUP_EA:x}" if ref >= DGROUP_EA else None,
                         "refs": [f"0x{ea:x}"]}
    return [seen[k] for k in sorted(seen)]


def context(ea: int, before: int = 8, after: int = 8) -> list[dict]:
    f = ida_funcs.get_func(ea)
    if f is None:
        return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}]
    items = list(idautils.FuncItems(f.start_ea))
    if ea not in items:
        return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}]
    i = items.index(ea)
    lo, hi = max(0, i - before), min(len(items), i + after + 1)
    return [{"ea": f"0x{e:x}", "disassembly": disassembly(e)} for e in items[lo:hi]]


def scan_plus10() -> list[dict]:
    """全庫掃「讀某個暫存器 +10h」的地方。**xref 看不到相對定址，只能掃。**"""
    out = []
    for fn in idautils.Functions():
        for ea in idautils.FuncItems(fn):
            dis = disassembly(ea)
            if not PLUS10.search(dis):
                continue
            out.append({"ea": f"0x{ea:x}", "function": f"0x{fn:x}",
                        "disassembly": dis, "context": context(ea)})
            if len(out) >= 120:
                return out
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
        "schema": "kol-ida-koltown-clouds-monster-flag-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：雲層動畫的版面，以及怪物記錄 +0x10 低位的讀取點。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "cloud_fn": summary(CLOUD_FN),
        "cloud_listing": listing(CLOUD_FN),
        "cloud_globals": globals_used(CLOUD_FN),
        "shatter_context": context(SHATTER_EA, before=24, after=24),
        "monster_field_fn": summary(MONSTER_FIELD_FN),
        "monster_field_listing": listing(MONSTER_FIELD_FN),
        "plus10_sites": scan_plus10(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
