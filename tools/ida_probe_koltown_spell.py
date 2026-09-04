#!/usr/bin/env python3
"""IDA Pro 9.4：施法那一手。

執行段的指令碼 `0x0D` 分派到 `sub_2818D`（RE-093）。訊息索引指出它印
0x169『casts a spell, but nothing is there!』、0x16A『…but it has no effect!』、
0x16B『casts|S at|F.』、0x16C 的部位選單、0x16D＋0x172『The spell has a … effect.』、
以及 0x184／0x185／0x186 的「Select an area to heal／hurt」。

一次倒完整支與它的呼叫關係，另外附上角色記錄 `0x1A0`（會不會法術，RE-067）
與法術名稱可能落在的幾個 DGROUP 區塊。**不要改一個常數重跑一次**——那會把 `.i64`
弄壞（見 `tools/run_ida_probe.sh` 的標頭）。

  idat -A -S"tools/ida_probe_koltown_spell.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x2818D, 0x265F3, 0x277F4,
         # 施法用到的判斷與收尾：目標是隊員還是敵人、腿治好之後、找目標格兩支、
         # 法術失效那一輪的兩支。
         0x22229, 0x21D94, 0x24BB6, 0x24C35, 0x25D0F, 0x223C0)
# 回合開始時印 0x173（法術失效）的那一段。
WINDOWS = ((0x20F00, 0x20FA0),)
# 想倒的 DGROUP 區塊（位移、長度）。
DUMP_BLOCKS = (
    (0x12A4, 8),   # 效果 1…8 → 參戰者記錄的欄位偏移
    (0x00D4, 8),   # 骰面表（2 3 5 7 9 11 19 99），當驗證用
)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def line(ea: int) -> dict:
    return {"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append(line(ea))
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def callees(start: int) -> list[str]:
    out = set()
    for ea in idautils.FuncItems(start):
        if idc.print_insn_mnem(ea) != "call":
            continue
        for t in idautils.CodeRefsFrom(ea, False):
            f = ida_funcs.get_func(t)
            if f is not None and f.start_ea == t:
                out.add(f"0x{t:x}")
    return sorted(out)


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-spell-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：施法那一手；尚未鎖定任何斷言。",
        "funcs": {f"0x{ea:x}": {
            "name": idc.get_func_name(ea),
            "end_ea": f"0x{ida_funcs.get_func(ea).end_ea:x}",
            "instruction_count": len(list(idautils.FuncItems(ea))),
            "callers": sorted({f"0x{c:x}" for c in idautils.CodeRefsTo(ea, False)}),
            "callees": callees(ea),
            "listing": [line(i) for i in idautils.FuncItems(ea)],
        } for ea in FUNCS if ida_funcs.get_func(ea) is not None},
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
