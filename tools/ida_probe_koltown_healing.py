#!/usr/bin/env python3
"""IDA Pro 9.4：野外治療那一條剩下的洞（RE-112 續）。

怪物編號 0xFF 那一條是「Will you be healed?」（記錄 0xE6），敘述走 `sub_2A46E`
但基底換成 `0x2350 + (索引 << 6)`。**那六段敘述從哪裡載入還沒解**——`0x2350` 在
BSS，而 DGROUP offset 的字面值在反組譯裡查不到寫入點（同 RE-128 的 `byte_32386`
教訓），所以這裡改用**線性位址** DGROUP+0x2350 ＝ `0x30220` 問 xref，並一併倒出
擲這一條的機率表 `0x133E` 與算平均位階那一支。

  idat -A -S"tools/ida_probe_koltown_healing.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x2A46E, 0x2A337, 0x2A437, 0x2A4E5)
# 六段敘述的緩衝與擲這一條的機率表，都用線性位址。
GLOBALS = {
    "0x2350_healing_text": DGROUP_EA + 0x2350,
    "0x133e_healing_odds": DGROUP_EA + 0x133E,
}
# 那六段敘述在載入映像裡的內容（0x2350 起 6 × 0x40）。
BLOCKS = ((0x2350, 6 * 0x40), (0x133E, 0x20))
# 相對定址查不到 xref，另外用 operand 文字掃一次。
OPERANDS = ("2350h", "133Eh", "30220", "3020E")


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
    out = []
    for x in idautils.XrefsTo(ea, 0):
        out.append({
            "from": f"0x{x.frm:x}",
            "function": idc.get_func_name(x.frm) or "",
            "type": idautils.XrefTypeName(x.type),
            "disassembly": disassembly(x.frm),
        })
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
        "schema": "kol-ida-koltown-healing-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：治療那一條的六段敘述來源與機率表；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "globals": {name: refs(ea) for name, ea in GLOBALS.items()},
        "operand_hits": operand_hits(OPERANDS),
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
