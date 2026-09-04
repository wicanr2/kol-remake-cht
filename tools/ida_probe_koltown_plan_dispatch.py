#!/usr/bin/env python3
"""IDA Pro 9.4：規劃段選完圖示之後的分派與收尾。

RE-082 的分派表（`0x1FA02`）沒有列 0x22，但執行段的 `0x20799` 對 0x22 與 0x23
都做裝填檢查，而指令列那一格的出現條件正好是「還沒裝填」（RE-083）。0x22 到底
有沒有被寫進 +0x25，要看分派的 default 那一條。

順便倒出兩處三個全域的寫回點：敵人 `0x1F104` 與玩家 `0x1FA6D`，好定 `word_30916`
的語意（RE-104 標為未解）。

  idat -A -S"tools/ida_probe_koltown_plan_dispatch.py OUT.json INPUT.EXE" INPUT.i64
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_lines
import idaapi
import idc

EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
WINDOWS = ((0x1F9E0, 0x1FAB0), (0x1F0D0, 0x1F130), (0x1F780, 0x1F7D0))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
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
        "schema": "kol-ida-koltown-plan-dispatch-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：規劃段的分派 default 與三個全域的寫回；尚未鎖定任何斷言。",
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
