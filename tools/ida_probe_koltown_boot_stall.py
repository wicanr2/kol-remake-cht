#!/usr/bin/env python3
"""IDA Pro 9.4：原版開機之後卡在哪一支。

dosgolem 把 `KOLTOWN-unpacked.EXE` 跑起來，十三個資料檔全部開得起來之後，程式停在
IDA `0x1B126`…`0x1B566` 那一段裡打轉。已經排除的：不是缺檔（`Missing()` 空的）、
不是等硬體（一個 I/O 埠都沒讀）、不是等鍵盤（七個鍵逐個送過都不動）。視訊模式
從頭到尾是 03h——它連圖形模式都還沒切進去。

這一支一次問完：那一段屬於哪個函式、誰呼叫它、它呼叫誰、停止點碰的兩個全域被誰
用，以及「誰設視訊模式」（那是它應該走到卻沒走到的地方）。

⚠ **對同一個 `.i64` 連續開關會把資料庫弄壞**，所以不要改一個參數重跑一次。

  idat -A -S"tools/ida_probe_koltown_boot_stall.py OUT.json INPUT.EXE" INPUT.i64
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

# dosgolem 兩百次取樣落到的位址（`internal/oracle/boot_test.go`）。
SAMPLES = [
    0x1B126, 0x1B12E, 0x1B133, 0x1B13B, 0x1B145, 0x1B158, 0x1B165,
    0x1B2D0, 0x1B2D9, 0x1B2F9, 0x1B304, 0x1B3A0, 0x1B3A6,
    0x1B4E8, 0x1B4F3, 0x1B4FE, 0x1B504, 0x1B50C, 0x1B54A, 0x1B553, 0x1B566,
]
# 停止點那兩道指令碰的全域（`mov word [ds:44AA]`、`cmp word [ds:44B4]`）。
STALL_GLOBALS = [DGROUP_EA + 0x44AA, DGROUP_EA + 0x44B4]
# 進入點與 `_main`，用來把呼叫鏈往上接。
MAIN = 0x10000


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


def owning_funcs(addrs: list[int]) -> dict:
    """每個取樣位址屬於哪個函式。**分布均勻不代表同一支**，要問過才知道。"""
    out = {}
    for ea in addrs:
        f = ida_funcs.get_func(ea)
        out[f"0x{ea:x}"] = {
            "function": f"0x{f.start_ea:x}" if f else None,
            "name": idc.get_func_name(ea) if f else None,
            "disassembly": disassembly(ea),
        }
    return out


def refs_to_globals(addrs: list[int]) -> dict:
    """誰碰這幾個全域。直接 xref 與名字比對都做——**只做一種會漏**。"""
    out = {}
    for g in addrs:
        name = idc.get_name(g) or ""
        direct = []
        for ea in idautils.DataRefsTo(g):
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
                        by_name.append({"ea": f"0x{ea:x}",
                                        "function": f"0x{start:x}",
                                        "disassembly": text})
        out[f"0x{g:x}"] = {"name": name, "direct_xrefs": direct, "by_name": by_name}
    return out


def video_mode_sites() -> list[dict]:
    """誰設視訊模式（`int 10h` 且 AH=0）。**那是它應該走到卻沒走到的地方。**"""
    out = []
    pattern = re.compile(r"int\s+10h")
    for start in idautils.Functions():
        items = list(idautils.FuncItems(start))
        for i, ea in enumerate(items):
            text = disassembly(ea)
            if not pattern.search(text):
                continue
            # 前四道一起帶出來，才看得到 AX 是怎麼設的。
            context = [{"ea": f"0x{e:x}", "disassembly": disassembly(e)}
                       for e in items[max(0, i - 4):i + 1]]
            out.append({"function": f"0x{start:x}", "ea": f"0x{ea:x}",
                        "context": context})
    return out


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    owners = owning_funcs(SAMPLES)
    stalled = sorted({int(v["function"], 16) for v in owners.values() if v["function"]})

    chain = list(stalled)
    for ea in stalled:
        chain += callees(ea)
        f = ida_funcs.get_func(ea)
        if f is not None:
            chain += [int(c, 16) for c in summary(ea)["callers"]]
    chain = sorted({ea for ea in chain if ida_funcs.get_func(ea) is not None})
    # 呼叫端可能落在函式中間，正規化到函式起點。
    chain = sorted({ida_funcs.get_func(ea).start_ea for ea in chain})

    receipt = {
        "schema": "kol-ida-koltown-boot-stall-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：原版開機之後卡住的那一段；尚未鎖定任何斷言。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "sample_owners": owners,
        "stalled_functions": [f"0x{ea:x}" for ea in stalled],
        "functions": [summary(ea) for ea in chain],
        "listings": {f"0x{ea:x}": listing(ea) for ea in stalled},
        "main_summary": summary(MAIN),
        "stall_globals": refs_to_globals(STALL_GLOBALS),
        "video_mode_sites": video_mode_sites(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
