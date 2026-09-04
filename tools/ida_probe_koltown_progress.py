#!/usr/bin/env python3
"""IDA Pro 9.4：角色記錄 `[0x1DD]` 那一格是誰寫的。

結局的判準是 `[0x1DD + word_2DF70] == 0x17`（RE-106），但那一格是誰寫的沒解。
operand 掃描找到五個使用點，其中 `sub_11A93` 是建築內的主迴圈（RE-074 的訓練
就在那一支裡），所以「任務的回報」很可能也在那裡。本探針倒出五處的前後文。

  idat -A -S"tools/ida_probe_koltown_progress.py OUT.json INPUT.EXE" INPUT.i64
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
# 五個使用點各看前 0xA0、後 0x60 個 byte。
USES = (0x11CC8, 0x145DC, 0x160CD, 0x167F6, 0x2B392)
BEFORE, AFTER = 0xA0, 0x60


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
        "schema": "kol-ida-koltown-progress-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：角色記錄 [0x1DD] 的五個使用點；尚未鎖定任何斷言。",
        "windows": {f"0x{ea:x}": window(ea - BEFORE, ea + AFTER) for ea in USES},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
