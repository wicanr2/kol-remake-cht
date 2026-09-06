#!/usr/bin/env python3
"""IDA Pro 9.4：誰該設定 `word_2FB54`（搬移迴圈的目標段）。

`sub_1DC26` 是 `mov es, [ds:1C84]` ／ `rep movsw` 的搬移常式。執行期量到那個全域
**從頭到尾是 0**（`internal/oracle` 的 `TestProbeWhoSetsDestSegment`），於是 `ES = 0`，
`rep movsw` 把中斷向量表整片寫掉；下一次計時器中斷就把 `CS:IP` 送進垃圾。

原版在真 DOS 上跑得起來，所以那個全域**應該**在更早的某處被設好。這一支問：
誰碰它、`sub_1DC26` 是誰呼叫的、以及寫 `ds:1C86` 那一道（`0x10410`）在做什麼。

⚠ **對同一個 `.i64` 連續開關會把資料庫弄壞**，所以一次問完，不要改參數重跑。

  idat -A -S"tools/ida_probe_koltown_dest_segment.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
import re
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

# 搬移迴圈那一支（`push bp ／ mov bp, sp ／ mov es, [1C84]`）。
MOVER = 0x1DC26
# 目標段與它隔壁那一格（後者在執行期被寫成 8）。
DEST_SEG = DGROUP_EA + 0x1C84
DEST_NEXT = DGROUP_EA + 0x1C86
# 寫 `ds:1C86` 那一道所在的位置。
WRITER_SITE = 0x10410


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    return [{"ea": f"0x{ea:x}",
             "bytes": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
             "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {"original_name": idc.get_func_name(start),
            "start_ea": f"0x{start:x}", "end_ea": f"0x{f.end_ea:x}",
            "instruction_count": len(list(idautils.FuncItems(start))),
            "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)})}


def touches(addr: int) -> dict:
    """誰碰這個全域。**直接 xref 與名字掃描都做**——只做一種會漏：
    走 `[bx+常數]` 的間接存取在 xref 裡看不到（RE-113 那次就是這樣漏掉的）。"""
    name = idc.get_name(addr) or ""
    direct = []
    for ea in idautils.DataRefsTo(addr):
        f = ida_funcs.get_func(ea)
        direct.append({"ea": f"0x{ea:x}",
                       "function": f"0x{f.start_ea:x}" if f else None,
                       "disassembly": disassembly(ea)})
    by_name = []
    if name:
        pattern = re.compile(re.escape(name))
        for start in idautils.Functions():
            for ea in idautils.FuncItems(start):
                text = disassembly(ea)
                if pattern.search(text):
                    by_name.append({"ea": f"0x{ea:x}", "function": f"0x{start:x}",
                                    "disassembly": text})
    return {"address": f"0x{addr:x}", "name": name,
            "direct_xrefs": direct, "by_name": by_name}


def context(ea: int, before: int = 12, after: int = 6) -> list[dict]:
    """某一道指令前後的碼。函式邊界之外不取。"""
    f = ida_funcs.get_func(ea)
    if f is None:
        return []
    items = list(idautils.FuncItems(f.start_ea))
    if ea not in items:
        return []
    i = items.index(ea)
    lo, hi = max(0, i - before), min(len(items), i + after + 1)
    return [{"ea": f"0x{e:x}", "disassembly": disassembly(e)} for e in items[lo:hi]]


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    mover = ida_funcs.get_func(MOVER)
    mover_start = mover.start_ea if mover else MOVER
    callers = [int(c, 16) for c in summary(mover_start).get("callers", [])]
    caller_funcs = sorted({ida_funcs.get_func(ea).start_ea
                           for ea in callers if ida_funcs.get_func(ea)})

    receipt = {
        "schema": "kol-ida-koltown-dest-segment-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：搬移常式的目標段從哪來；尚未鎖定任何斷言。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "mover": summary(mover_start),
        "mover_listing": listing(mover_start),
        "mover_callers": [summary(ea) for ea in caller_funcs],
        "dest_segment": touches(DEST_SEG),
        "dest_next": touches(DEST_NEXT),
        "writer_site": context(WRITER_SITE, before=20, after=8),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
