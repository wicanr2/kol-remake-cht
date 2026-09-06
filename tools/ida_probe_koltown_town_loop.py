#!/usr/bin/env python3
"""IDA Pro 9.4：載入完成之後空轉的那一支。

原版走到 `TOWN7.0`、畫出木紋框架之後，停在 IDA `0x1A1BC`…`0x1A1EE` 的迴圈裡跑滿
2.4 億道指令沒有離開。那一段跟著一條鏈結串列走到 `-1`，而節點永遠是 `0x07A0`
——表的 `+0x0F40`（＝ `0x07A0 × 2`）存著 `0x07A0`，指向自己。

寫入序列已經量到（`internal/oracle` 的 `TestProbeWhoWritesSelfLoop`）：那塊記憶體
被 `0x1C873` 寫成過 ASCII 的 `'DDD'`——**是共用暫存區，不是專用的鏈結表**，
而最後把那一格寫成自己的是走訪迴圈同一支的 `0x1A246`。

這一支問完剩下的：兩支的完整反組譯、誰呼叫它們、誰設那個 far 指標。

⚠ 這是本輪對同一個 `.i64` 的第四次開啟。**一次問完**。壞掉的處置：把 `.i64` 改名
留證，`idat -A -B <EXE>` 重建，再拿一支跑過的探針比對輸出逐條相同才繼續。

  idat -A -S"tools/ida_probe_koltown_town_loop.py OUT.json INPUT.EXE" INPUT.i64
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

# 走訪迴圈與它自己的寫入點。
LOOP_EA = 0x1A1BC
WRITER_EA = 0x1A246
# 把那塊記憶體寫成 'DDD' 的那一支。
TEXT_WRITER_EA = 0x1C873
# 表的 far 指標（`ds:2528` 位移、`ds:252A` 段）。
TABLE_PTR_OFF = DGROUP_EA + 0x2528
TABLE_PTR_SEG = DGROUP_EA + 0x252A


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


def owner(ea: int) -> int | None:
    f = ida_funcs.get_func(ea)
    return f.start_ea if f else None


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

    loop_fn = owner(LOOP_EA)
    text_fn = owner(TEXT_WRITER_EA)

    receipt = {
        "schema": "kol-ida-koltown-town-loop-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：載入完成後空轉的那一支；尚未鎖定任何斷言。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "loop_function": summary(loop_fn) if loop_fn else None,
        "loop_listing": listing(loop_fn) if loop_fn else [],
        "writer_owner": f"0x{owner(WRITER_EA):x}" if owner(WRITER_EA) else None,
        "text_writer_function": summary(text_fn) if text_fn else None,
        "text_writer_listing": listing(text_fn) if text_fn else [],
        "table_ptr_off": touches(TABLE_PTR_OFF),
        "table_ptr_seg": touches(TABLE_PTR_SEG),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
