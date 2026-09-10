#!/usr/bin/env python3
"""IDA Pro 9.4：騎乘那一手、逃走之後掉東西，以及角色立繪的合成。

三件事一次問完（同一個 `.i64` 連續開關會弄壞它）：

  * **騎乘**（指令碼 0x28 → `sub_227C0`，RE-093／RE-094）：指令列那一格已經會出現，
    當時執行段尚未接上（後由 RE-124 解出，`CmdLand` 已實作）。
  * **逃走之後掉東西**（`0x21765` 那一段，RE-095 只讀到「機率讀 `byte_3AC7D`，
    那個機率是誰算的還沒對上」）：連同它所在的那一支一起倒出來。
  * **角色立繪**（`sub_18F3A`）：資料頁的物品與隊伍那兩頁都從它開始（RE-123），
    要知道它怎麼疊 `HEADPICS`／`TORPICS`／`LEGSPICS`。

順便補 `sub_2625A`——它在執行段三個地方當「要不要印訊息」的開關，語意一直未定。

  idat -A -S"tools/ida_probe_koltown_mount_drop.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCTIONS = (0x227C0, 0x18F3A, 0x2625A)
# 逃走之後掉東西那一段。
WINDOWS = ((0x21700, 0x21800),)
# 掉落機率那一格與它前後（`byte_3AC7D` 在 DGROUP + 0xCEAD）。
BLOCKS = ((0xCEA0, 0x20),)


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


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-mount-drop-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：騎乘、逃走掉東西與立繪合成；尚未鎖定任何斷言。",
        "functions": {f"0x{ea:x}": summary(ea) for ea in FUNCTIONS},
        "listings": {f"0x{ea:x}": listing(ea) for ea in FUNCTIONS},
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
