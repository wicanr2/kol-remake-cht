#!/usr/bin/env python3
"""IDA Pro 9.4：任務物品放在戰場哪一格（RE-113 續）。

RE-113 記著「`byte_30C6E`／`byte_30A7E` 在 KOLTOWN 裡只有寫沒有讀，靜態 grep 到
此為止」。**只有寫沒有讀是「查詢回空」的形狀，不是答案**：xref 只看得到絕對定址
的存取，`[bx + 2D9Eh]` 這種相對定址一個都不會出現。所以這一支換一個過濾器——
用 DGROUP offset 的字面值掃 operand 文字，範圍是那兩個全域與它們前後幾格。

同一份輸出順便倒出 `sub_22229`（RE-113 記著「決定搬到 `+0x20` 還是 `+0x30` 的
判準未解」，而 RE-122 的疲勞分母也卡在同一支）。

  idat -A -S"tools/ida_probe_koltown_quest_item.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x22229, 0x223C0, 0x269BA)
GLOBALS = {"byte_30C6E": 0x30C6E, "byte_30A7E": 0x30A7E}
# 那兩個全域的 DGROUP offset：0x2D9E 與 0x2BAE。相對定址寫成這種字面值。
OPERANDS = ("2D9Eh", "2BAEh", "30C6E", "30A7E")
# `sub_1E3A8`（戰場那一支）與 `sub_228B7` 都太大，只取用到那兩個陣列的地方前後。
WINDOWS = ((0x1E550, 0x1E590), (0x201D0, 0x20340), (0x204C0, 0x20500),
           (0x22C00, 0x22C60))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    if ida_funcs.get_func(start) is None:
        return []
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def window(lo: int, hi: int) -> list[dict]:
    out, ea = [], lo
    while ea < hi:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{f.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
        "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)}),
    }


def refs(ea: int) -> dict:
    out = []
    for x in idautils.XrefsTo(ea, 0):
        out.append({"from": f"0x{x.frm:x}",
                    "function": idc.get_func_name(x.frm) or "",
                    "type": idautils.XrefTypeName(x.type),
                    "disassembly": disassembly(x.frm)})
    return {"ea": f"0x{ea:x}", "name": idc.get_name(ea) or "", "xrefs": out}


def operand_hits(needles: tuple[str, ...]) -> dict:
    out: dict[str, list[dict]] = {n: [] for n in needles}
    for fn in idautils.Functions():
        for ea in idautils.FuncItems(fn):
            text = disassembly(ea)
            for n in needles:
                if n in text:
                    out[n].append({"ea": f"0x{ea:x}",
                                   "function": idc.get_func_name(fn) or "",
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
        "schema": "kol-ida-koltown-quest-item-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：任務物品那兩個全域的相對定址讀者；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "globals": {name: refs(ea) for name, ea in GLOBALS.items()},
        "operand_hits": operand_hits(OPERANDS),
        "windows": {f"0x{lo:x}-0x{hi:x}": window(lo, hi) for lo, hi in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
