#!/usr/bin/env python3
"""IDA Pro 9.4：所有 archive 載入點的緩衝位址（RE-112 續）。

治療那一條的六段敘述在 DGROUP `0x2350`，而那一塊在 BSS、**零個參照點**——
既沒有 xref 也沒有 operand 命中。零不是答案，是「過濾器有洞」的訊號：緩衝可能
由更大的一次載入涵蓋，位址寫成別的形式。

所以換一個過濾器：把 `sub_1A4FE`（RE-128 證實的 archive 載入器，
`sub_1A4FE(檔名, entry, ds:緩衝, 大小)`）的每一個呼叫點連同前面幾行倒出來，
看哪一次的緩衝涵蓋 `0x2350`。同一份輸出也列出全檔 `mov ax, offset ...` 取的
每一個 DGROUP 位址，範圍限在 `0x30000`…`0x30400`（＝ DGROUP `0x2130`…`0x2530`）。

  idat -A -S"tools/ida_probe_koltown_loaders.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
DGROUP_EA = 0x2DED0
LOADER = 0x1A4FE
# 呼叫點前面幾行就是推參數的地方。Borland 小模式一個參數一行 push。
CONTEXT = 14
# 取址掃描的範圍：DGROUP 0x2130…0x2530，把 0x2350 夾在中間。
OFFSET_LO, OFFSET_HI = DGROUP_EA + 0x2130, DGROUP_EA + 0x2530


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    calls = []
    for frm in idautils.CodeRefsTo(LOADER, False):
        fn = ida_funcs.get_func(frm)
        before = []
        if fn is not None:
            items = list(idautils.FuncItems(fn.start_ea))
            if frm in items:
                i = items.index(frm)
                before = [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
                          for ea in items[max(0, i - CONTEXT):i]]
        calls.append({"call_ea": f"0x{frm:x}",
                      "function": idc.get_func_name(frm) or "",
                      "before": before})

    offsets = []
    for fn in idautils.Functions():
        for ea in idautils.FuncItems(fn):
            text = disassembly(ea)
            if "offset" not in text:
                continue
            for op in (0, 1):
                v = idc.get_operand_value(ea, op)
                if OFFSET_LO <= v <= OFFSET_HI:
                    offsets.append({"ea": f"0x{ea:x}",
                                    "function": idc.get_func_name(fn) or "",
                                    "value": f"0x{v:x}",
                                    "dgroup_offset": f"0x{v - DGROUP_EA:04x}",
                                    "disassembly": text})
                    break

    receipt = {
        "schema": "kol-ida-koltown-loaders-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：archive 載入點的緩衝與一段 DGROUP 的取址點；尚未鎖定任何斷言。",
        "loader": f"0x{LOADER:x}",
        "calls": sorted(calls, key=lambda c: c["call_ea"]),
        "offsets_in_range": offsets,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
