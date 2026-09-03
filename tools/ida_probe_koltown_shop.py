#!/usr/bin/env python3
"""IDA Pro 9.4：客棧 REST（住宿＋存檔）與修道院 REST（治療）的價格與存檔路徑。

已知錨點（sub_11A93）：
  * `0x1265f` 起是 `byte_2FC00 & 0x40` 的分支——修道院治療，價格 = 角色 +0x1B × byte_2FBFD。
  * `0x127b1` 起是沒有 0x40 的分支——客棧；問記錄 0x35「Whole party?」之後
    對每個成員呼叫 `sub_14DB1`，價格與「Save character to hard disk?」都在那支裡。

本探針把 `sub_14DB1` 與週邊幾支完整倒出，並列出價格／金幣／角色指標的所有讀寫點。

  idat -A -S"tools/ida_probe_koltown_rest_save.py OUT.json INPUT.EXE" INPUT.i64
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

DUMP_FUNCTIONS = (0x14371, 0x17318, 0x1716A)
DUMP_LIMIT = 9000

GLOBALS = {
    "text_col_word_2DF14": 0x2DF14,
    "text_row_word_2DF16": 0x2DF16,
    "sheet_mode_word_2DF46": 0x2DF46,
}

TEXT_DRAW = 0x1A8FD


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
    """16-bit real mode 的 DGROUP 全域用位移定址，xref 不可靠——直接掃指令文字。"""
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


def text_record_index() -> list[dict]:
    """每個 `call sub_1A8FD` 畫的是哪一筆 UI 記錄。

    呼叫慣例是右到左，所以最靠近 call 的那個 `push ax` 才是記錄編號；
    往前找最近的 `mov ax, imm`。取不到常數就記 null（編號是算出來的）。
    """
    out = []
    for ea in idautils.CodeRefsTo(TEXT_DRAW, False):
        fn = ida_funcs.get_func(ea)
        if fn is None:
            continue
        items = list(idautils.FuncItems(fn.start_ea))
        if ea not in items:
            continue
        i = items.index(ea)
        index = None
        seen_push = False
        for j in range(i - 1, max(-1, i - 8), -1):
            text = disassembly(items[j])
            if text.startswith("push"):
                seen_push = True
                continue
            if seen_push and text.startswith("mov     ax,"):
                operand = idc.get_operand_value(items[j], 1)
                if idc.get_operand_type(items[j], 1) == idc.o_imm:
                    index = operand
                break
        out.append({"call_ea": f"0x{ea:x}", "function": f"0x{fn.start_ea:x}",
                    "record": None if index is None else f"0x{index:x}"})
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
        "schema": "kol-ida-koltown-award-explore-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：客棧住宿／存檔與修道院治療的價格路徑；尚未鎖定任何斷言。",
        "text_record_index": text_record_index(),
        "globals": {name: global_users(addr) for name, addr in GLOBALS.items()},
        "dumped_functions": [function_slice(ea) for ea in DUMP_FUNCTIONS],
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in ((0x0130, 128),)},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
