#!/usr/bin/env python3
"""IDA Pro 9.4：接任務那一段的文字、問句與一覽畫面的畫格。

`sub_14564` 已經讀完（接任務的門檻、旗標怎麼設成「接了」），剩下四個小洞：

  * `sub_14234(位移, 旗標)` 連放四段（0x100／0x140／0x180／0x1C0），像是任務敘述；
  * `sub_144EF` 是那個「要接嗎」的問句，回傳決定接或不接；
  * `[0x7C]` 那張四格表是任務旗標的除數（1／4／16／64）；
  * `sub_168F8(編號, 群組, x, y)` 是一覽畫面上的一格。

外加一次 operand 值掃描找 `[0xB926]` 的寫入者——上一支用 DGROUP 相對值掃不到，
因為反組譯把它寫成絕對位址 `0x2DED0 + 0xB926 = 0x397F6`。

  idat -A -S"tools/ida_probe_koltown_quest_offer.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x14234, 0x144EF, 0x1452D, 0x144C3, 0x168F8, 0x1449C)
# 接任務的呼叫者：什麼時候會問「要不要接」。
WINDOWS = (0x1242E,)
BEFORE, AFTER = 0x120, 0x80
# `[0xB926]` 與 `[0xB956]` 的絕對位址。
WANTED = (DGROUP_EA + 0xB926, DGROUP_EA + 0xB956, DGROUP_EA + 0x7C)
BLOCKS = ((0x7C, 0x10), (0x100, 0x10), (0x140, 0x10))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


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


def scan_operands(wanted: set[int]) -> dict:
    hits: dict[str, list[dict]] = {f"0x{v:x}": [] for v in wanted}
    for start in idautils.Functions():
        name = idc.get_func_name(start)
        for ea in idautils.FuncItems(start):
            for n in range(2):
                if idc.get_operand_value(ea, n) in wanted:
                    hits[f"0x{idc.get_operand_value(ea, n):x}"].append({
                        "ea": f"0x{ea:x}", "function": name, "operand": n,
                        "disassembly": disassembly(ea)})
    return hits


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-quest-offer-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：接任務的文字與問句、一覽畫面的畫格；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS
                     if len(list(idautils.FuncItems(ea))) < 400},
        "windows": {f"0x{ea:x}": window(ea - BEFORE, ea + AFTER) for ea in WINDOWS},
        "operand_hits": scan_operands(set(WANTED)),
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
