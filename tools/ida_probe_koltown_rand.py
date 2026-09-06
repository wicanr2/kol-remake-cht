#!/usr/bin/env python3
"""IDA Pro 9.4：原版的亂數產生器。

`sub_1B5A8(n)` 是 remake 一路寫成 `rand(n)` 的那一支（回傳 1…n）。它底下呼叫的
才是 C 函式庫的 `rand`。要把 remake 的亂數與原版對齊，得先知道那一支的
**乘數、增量與取出哪幾個位元**，以及種子放在哪、誰播種。

  idat -A -S"tools/ida_probe_koltown_rand.py OUT.json INPUT.EXE" INPUT.i64
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
# `rand(n)` 的包裝；底下那一支由 call 追出來。
WRAPPER = 0x1B5A8


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


def callees(start: int) -> list[int]:
    out = []
    for ea in idautils.FuncItems(start):
        if idc.print_insn_mnem(ea) != "call":
            continue
        for t in idautils.CodeRefsFrom(ea, False):
            f = ida_funcs.get_func(t)
            if f is not None and f.start_ea == t:
                out.append(t)
    return sorted(set(out))


def seed_sites() -> list[dict]:
    """誰播種：掃全庫找 `int 1Ah`（BIOS tick，KOLBOOK 那一側 RE-035 看過同一套）。"""
    out = []
    pattern = re.compile(r"int\s+1Ah")
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            if pattern.search(text):
                out.append({"function": f"0x{start:x}", "ea": f"0x{ea:x}",
                            "disassembly": text})
    return out


def seed_writers() -> list[dict]:
    """誰碰種子那兩個 word（`word_2F4D0`／`word_2F4D2`，DGROUP 0x1600）。"""
    out = []
    pattern = re.compile(r"word_2F4D[02]")
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            if pattern.search(text):
                out.append({"function": f"0x{start:x}", "ea": f"0x{ea:x}",
                            "disassembly": text})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    chain = [WRAPPER, 0x10000, 0x1A8DF, 0x2D526] + callees(WRAPPER)
    # 再往下一層：`rand` 自己可能還呼叫別的（例如長整數乘法的 helper）。
    deep = []
    for ea in chain[1:]:
        deep += [c for c in callees(ea) if c not in chain]
    chain += sorted(set(deep))

    receipt = {
        "schema": "kol-ida-koltown-rand-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：原版亂數產生器與播種點；尚未鎖定任何斷言。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "functions": [summary(ea) for ea in chain],
        "listings": {f"0x{ea:x}": listing(ea) for ea in chain},
        "seed_sites": seed_sites(),
        "seed_writers": seed_writers(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
