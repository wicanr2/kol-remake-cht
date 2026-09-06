#!/usr/bin/env python3
"""IDA Pro 9.4：誰呼叫 `sub_22084`（扣裝備耐久），扣幾次。

扣法已經實跑量到（`internal/oracle` 的 `TestProbeDurabilityRoutine`）：

    arg0 是物品格相對 `[ds:BB78]` 的位移，arg1 是次數
    記錄 `+0x0F` 的 bit 5 沒設就不扣
    新耐久 = 記錄`+0x0B` − 32 × arg1
    新耐久為負 → 整格 16 byte 清成 0；剛好 0 不清

缺的是**呼叫端**：什麼時候扣、arg1 傳什麼。那一段跑不到（攻擊結算含畫面，
`sub_26D76` 直接呼叫會跑不完），所以只能靜態讀。

⚠ 這是本輪對同一個 `.i64` 的第三次開啟。**一次問完**，不要改參數重跑。

  idat -A -S"tools/ida_probe_koltown_durability.py OUT.json INPUT.EXE" INPUT.i64
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

DURABILITY_FN = 0x22084
# `[ds:BB78]`：指向當前記錄的指標。誰設它決定「扣的是誰的裝備」。
RECORD_PTR = DGROUP_EA + 0xBB78


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


def context(ea: int, before: int = 16, after: int = 4) -> list[dict]:
    """呼叫點前後的碼。**參數是往前推的 push，所以 before 要夠長。**"""
    f = ida_funcs.get_func(ea)
    if f is None:
        return []
    items = list(idautils.FuncItems(f.start_ea))
    if ea not in items:
        return []
    i = items.index(ea)
    lo, hi = max(0, i - before), min(len(items), i + after + 1)
    return [{"ea": f"0x{e:x}", "disassembly": disassembly(e)} for e in items[lo:hi]]


def touches(addr: int) -> list[dict]:
    out = []
    for ea in idautils.DataRefsTo(addr):
        f = ida_funcs.get_func(ea)
        out.append({"ea": f"0x{ea:x}",
                    "function": f"0x{f.start_ea:x}" if f else None,
                    "disassembly": disassembly(ea)})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    info = summary(DURABILITY_FN)
    callers = [int(c, 16) for c in info.get("callers", [])]

    receipt = {
        "schema": "kol-ida-koltown-durability-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：扣裝備耐久的呼叫端；尚未鎖定任何斷言。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "durability_fn": info,
        "durability_listing": listing(DURABILITY_FN),
        "call_sites": {f"0x{ea:x}": context(ea) for ea in callers},
        "record_ptr_writers": touches(RECORD_PTR),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
