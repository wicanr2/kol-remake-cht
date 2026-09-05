#!/usr/bin/env python3
"""IDA Pro 9.4：SCROLL 那一支的週邊，以及疲勞與那兩個未解欄位的讀者。

`sub_17318` 已經倒完（`spell-scroll.json`），但它呼叫的七支還沒讀：教派名怎麼印、
「已經是成員了嗎」怎麼判、選單第二列做什麼、錢包那兩支、以及講一句話就結束的
那一支。另外兩件順路問完：`sub_26CD5` 是全檔唯一讀參戰者記錄 `+0x2E`／`+0x2F`
的地方，`sub_21AC4` 是唯一寫疲勞基準 `[0x1C50]` 的地方。

  idat -A -S"tools/ida_probe_koltown_scroll_order.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x172E9, 0x1429C, 0x178E0, 0x14CCF, 0x14D3A, 0x172C1,
             0x14234, 0x26CD5, 0x21AC4, 0x2324F, 0x236AB)
# `%ld` 那兩個格式字串（`0x17CC`／`0x17D0`）與教派名可能落腳的區塊。
BLOCKS = ((0x17CC, 0x10), (0x1D2A, 0x10))


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


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-scroll-order-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：SCROLL 週邊與疲勞、+0x2E／+0x2F 的讀者；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
