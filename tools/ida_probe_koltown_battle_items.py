#!/usr/bin/env python3
"""IDA Pro 9.4：戰場上的撿、丟與背包三手。

執行段的分派（`0x2065E`…`0x2070F`）把指令碼分到各自的 handler。攻擊、移動、休息、
怯戰都解完了（RE-084／RE-085／SPEC-009），剩下 `CmdPack`（0x0D）、`CmdPickUp`
（0x15）、`CmdDrop`（0x24）與三個語意還沒對上的（0x16／0x25／0x26）。

本探針倒出分派那一段與它跳到的每一支，並附上訊息記錄與相關 DGROUP 區塊。
訊息已知的有：0x10C『cannot drop|B|E!』、0x10D『drops|B|E!』、0x111『You have
found a|H…』、0x112『…must drop or belt your|E…』、0x114『puts|B|I in|B belt』、
0x183『You empty your hands and pick up the item.』。

  idat -A -S"tools/ida_probe_koltown_battle_items.py OUT.json INPUT.EXE" INPUT.i64
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
# 執行段的分派表。
DISPATCH = (0x20150, 0x20660)
# `0x9D20` 是戰場上的物品格（`|H` 讀的那張表，RE-086／RE-091）。
ITEM_TABLE = 0x9D20
DUMP_BLOCKS = ((ITEM_TABLE, 0x40),)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
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


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    dispatch = window(*DISPATCH)
    # 分派段裡出現的 call 目標，就是各手的 handler。
    callees = []
    ea = DISPATCH[0]
    while ea < DISPATCH[1]:
        if idc.print_insn_mnem(ea) == "call":
            for t in idautils.CodeRefsFrom(ea, False):
                f = ida_funcs.get_func(t)
                if f is not None and f.start_ea == t:
                    callees.append(t)
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1

    receipt = {
        "schema": "kol-ida-koltown-battle-items-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：執行段的分派與撿／丟／背包三手；尚未鎖定任何斷言。",
        "dispatch": dispatch,
        "handlers": [summary(ea) for ea in sorted(set(callees))],
        "handler_listings": {
            f"0x{ea:x}": [{"ea": f"0x{i:x}", "disassembly": disassembly(i)}
                          for i in idautils.FuncItems(ea)]
            for ea in sorted(set(callees))
            if ida_funcs.get_func(ea) is not None
            and len(list(idautils.FuncItems(ea))) < 400
        },
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
