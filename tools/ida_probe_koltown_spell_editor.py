#!/usr/bin/env python3
"""IDA Pro 9.4：法術編輯器那一頁剩下的三塊（RE-128）。

`sub_178E0` 的主迴圈已經倒完（`scroll-order.json`），但它每一輪呼叫的兩支還沒讀：
`sub_1874E` 畫整頁、`sub_18350` 算成本並標出游標停在哪一欄。離開時的結算另外用到
四支帳目函式與四個全域，其中 `[0x1CF4 + 0x60 + n]` 是「這個教派准不准這一類」的
允許表——那一格是零就不讓選，而全檔沒有任何一支在 `0x1CF4` 字面值上寫入，所以
**改用線性位址** DGROUP+0x1CF4 ＝ `0x2FBC4` 問 xref（同 RE-128 的 `byte_32386`）。

  idat -A -S"tools/ida_probe_koltown_spell_editor.py OUT.json INPUT.EXE" INPUT.i64
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
# 畫頁、算成本，以及離開時結算用到的帳目函式。
FUNCTIONS = (0x1874E, 0x18350, 0x14427, 0x14465, 0x144EF, 0x144C3, 0x1A8FD)
# 要問寫入者的全域，全部用線性位址。DGROUP offset 附在名字裡。
GLOBALS = {
    "0x1cf4_order_allow": DGROUP_EA + 0x1CF4,
    "0x1c78_gold_cost": DGROUP_EA + 0x1C78,
    "0x1cec_ap_cost_lo": DGROUP_EA + 0x1CEC,
    "0x1cee_ap_cost_hi": DGROUP_EA + 0x1CEE,
    "0x2b9e_ap_have_lo": DGROUP_EA + 0x2B9E,
    "0x2ba0_ap_have_hi": DGROUP_EA + 0x2BA0,
}
# 教派允許表那一段：0x1CF4 起 0x80 個 byte，含 `+0x60` 那十六格。
BLOCKS = ((0x1CF4, 0x80),)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    if ida_funcs.get_func(start) is None:
        return []
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


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
    """一個全域的每一個參照點，附該行反組譯——讀跟寫要看得出來。"""
    out = []
    for x in idautils.XrefsTo(ea, 0):
        out.append({
            "from": f"0x{x.frm:x}",
            "function": idc.get_func_name(x.frm) or "",
            "type": idautils.XrefTypeName(x.type),
            "disassembly": disassembly(x.frm),
        })
    return {"ea": f"0x{ea:x}", "name": idc.get_name(ea) or "", "xrefs": out}


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-spell-editor-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：法術編輯器的版面、成本與教派允許表；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "globals": {name: refs(ea) for name, ea in GLOBALS.items()},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
