#!/usr/bin/env python3
"""IDA Pro 9.4：訊息記錄 → 使用位置的全域索引，外加戰後結算的完整反組譯。

`sub_1A8FD(記錄號, 緩衝)` 是唯一的訊息輸出入口（RE-060）。呼叫點前面一定是
`mov ax, imm` 把記錄號推上堆疊，所以往回掃兩三條就能把「哪一段程式印哪一句」整份
建出來。有了這份索引，要找某個玩家看得到的字是誰印的，就不必再逐支反組譯。

同時倒出 `sub_289EE`（戰後結算）與 `sub_22866`（隊員能不能收下戰利品）。

  idat -A -S"tools/ida_probe_koltown_message_index.py OUT.json INPUT.EXE" INPUT.i64
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
PRINT_MESSAGE = 0x1A8FD
FUNCS = (0x289EE, 0x22866)
# 往回看幾條指令找記錄號。`push ax` 之前的 `mov ax, imm` 通常就在兩條之內。
LOOKBACK = 6


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def line(ea: int) -> dict:
    return {"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}


def preceding(ea: int, n: int) -> list[int]:
    """往回列出 n 條指令的位址（靠函式的指令串，避免猜長度）。"""
    f = ida_funcs.get_func(ea)
    if f is None:
        return []
    items = list(idautils.FuncItems(f.start_ea))
    try:
        i = items.index(ea)
    except ValueError:
        return []
    return items[max(0, i - n):i]


def record_of(call_ea: int) -> dict:
    """回推這個呼叫印的是哪一號訊息。"""
    for ea in reversed(preceding(call_ea, LOOKBACK)):
        if idc.print_insn_mnem(ea) != "mov":
            continue
        if idc.print_operand(ea, 0) != "ax":
            continue
        if idc.get_operand_type(ea, 1) != idaapi.o_imm:
            # `mov ax, word_30A78` 這種是加上語言基底的間接記錄號，記下來但不當常數。
            return {"record": None, "via": disassembly(ea)}
        return {"record": idc.get_operand_value(ea, 1), "via": disassembly(ea)}
    return {"record": None, "via": None}


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    index = []
    for ea in sorted(idautils.CodeRefsTo(PRINT_MESSAGE, False)):
        info = record_of(ea)
        f = ida_funcs.get_func(ea)
        index.append({"ea": f"0x{ea:x}",
                      "func": idc.get_func_name(ea) if f else None,
                      "record": (f"0x{info['record']:x}" if info["record"] is not None
                                 else None),
                      "via": info["via"]})

    receipt = {
        "schema": "kol-ida-koltown-message-index-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：sub_1A8FD 呼叫點回推的訊息記錄號；記錄號為 None 表示"
                "由暫存器或語言基底算出，需個別看上下文。",
        "print_message": f"0x{PRINT_MESSAGE:x}",
        "call_count": len(index),
        "message_index": index,
        "funcs": {f"0x{ea:x}": [line(i) for i in idautils.FuncItems(ea)] for ea in FUNCS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
