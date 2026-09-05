#!/usr/bin/env python3
"""IDA Pro 9.4：SPEC-012 第 10 節剩下的三個洞（RE-128 續）。

法術護符那四格已經由 RE-128 解掉（`MAGIC.0` 的 CLASS／SUBCLASS：法術對哪一類、
哪一種怪物有效）。剩下三個：

1. **效果 7／8 打的兩個全域**。`sub_26CD5` 把 `word_34A12` 與 `word_32380` 各減
   十倍，但那兩個是什麼、在哪裡被讀還沒查。用**線性位址**問 xref——同 RE-128 的
   `byte_32386`：DGROUP offset 的字面值在反組譯裡查不到 BSS 全域的寫入點。
2. **疲勞的分母**。`sub_21AC4` 填成「力量 ＋ 本體記錄的某一格」，隊員取 `+0x15`、
   怪物取 `+0x20`；那兩格是什麼還沒解，而 `+0x20` 在 RE-116 被讀成「手上的武器」
   ——兩者至少有一個是錯的。
3. **參戰者記錄 `+0x2C`**（效果 5／6 寫的那一格）的讀者。

  idat -A -S"tools/ida_probe_koltown_spell_effects.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x21AC4, 0x26CD5, 0x26AC2, 0x26C5F, 0x26D76, 0x23683)
# 效果 7／8 打的那兩個全域，以及疲勞的分子與分母。全部用線性位址。
GLOBALS = {
    "0x6b42_effect7": 0x34A12,
    "0x44b0_effect8": 0x32380,
    "0x14b0_fatigue_num": DGROUP_EA + 0x322E,
    "0x1c50_fatigue_den": DGROUP_EA + 0x1C50,
}
# 參戰者記錄那三格的讀者要靠 operand 文字找，因為它們是 `[bx+2Ch]` 這種相對存取。
FIELDS = ("+2Ch]", "+2Dh]", "+15h]", "+20h]")
# 疲勞那兩張表是用索引存取的陣列（`[bx + 322Eh]`），對單一位址問 xref 得到零個
# ——那是「查詢回空」的第四種形狀：過濾器本身有洞。改用 operand 文字找。
OPERANDS = ("322Eh", "1C50h", "34A12", "32380")
# `sub_1E3A8` 是戰鬥主迴圈，整支太大：只取把那兩個全域相加的那一行前後。
WINDOWS = ((0x21270, 0x212C0),)


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
        out.append({
            "from": f"0x{x.frm:x}",
            "function": idc.get_func_name(x.frm) or "",
            "type": idautils.XrefTypeName(x.type),
            "disassembly": disassembly(x.frm),
        })
    return {"ea": f"0x{ea:x}", "name": idc.get_name(ea) or "", "xrefs": out}


def field_hits(needles: tuple[str, ...]) -> dict:
    """整份程式碼裡用到 `[reg+NNh]` 的每一行。範圍限在有函式的位址。"""
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
        "schema": "kol-ida-koltown-spell-effects-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：效果 7／8 的兩個全域、疲勞的分母與參戰者 +0x2C；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "globals": {name: refs(ea) for name, ea in GLOBALS.items()},
        "field_hits": field_hits(FIELDS),
        "operand_hits": field_hits(OPERANDS),
        "windows": {f"0x{lo:x}-0x{hi:x}": window(lo, hi) for lo, hi in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
