#!/usr/bin/env python3
"""IDA Pro 9.4：敵人 AI 那三張基礎權重表是誰填的。

RE-081 把 `sub_2374C` 的決策樹解完了，但三張基礎權重表——`byte_3496A`／`3496B`／
`3496C`（移動三式）、`[0x6AD2]`（攻擊式）、`[0x1A9A]`（部位）——在解包後的靜態影像
裡是 `0xFF`，也就是 BSS。當時的結論是「要執行期取樣」。

**BSS 的內容只有兩種來源**：程式碼寫進去，或從檔案載入。這支探針把兩條路都掃：

1. 每一個 `sub_1A4FE(fmt, index, ds:dest, size)` 呼叫點往前回溯，取出 dest 與 size，
   算出那次載入涵蓋的 DGROUP 範圍，看有沒有蓋到這三張表。
2. 逐條掃全庫，找 operand 落在這三張表附近的**寫入**指令。

  idat -A -S"tools/ida_probe_koltown_weight_tables.py OUT.json INPUT.EXE" INPUT.i64
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
LOADER = 0x1A4FE  # sub_1A4FE(fmt, index, ds:dest, size)
# 三張表的 DGROUP 位移。byte_3496A 的絕對位址是 0x3496A。
TABLES = {
    "byte_3496A（移動三式）": 0x3496A - DGROUP_EA,
    "0x6AD2（攻擊式）": 0x6AD2,
    "0x1A9A（部位）": 0x1A9A,
}
# 三張表的**預設值**：`byte_399D6` 為零時從這裡複製過去（`0x1ECCF`…`0x1ED38`）。
# 這幾塊在靜態影像裡就有值，不是 BSS。
DEFAULTS = {
    "0x1092（攻擊式預設，10 筆）": (0x1092, 10),
    "0x10A6（部位預設，7 筆）": (0x10A6, 7),
    "0x10AE（移動三式預設，3 筆）": (0x10AE, 3),
    "0x1092 前後": (0x1080, 0x40),
    # byte_399D6 決定走哪一條；它只有一個讀者、沒有 writer，看它是不是 BSS。
    "byte_399D6（選表旗標）": (0x399D6 - DGROUP_EA, 0x10),
    "0x0FF4（攻擊式／部位的情境權重）": (0x0FF4, 0x30),
}
# 回溯這麼多條指令找 push 的參數；16-bit C 的呼叫序列很短。
LOOKBACK = 24
# 三張表都在這一段被填（0x1EC6D…0x1ED27），連同前面讀來源的那幾條一起看。
WINDOWS = ((0x1EB80, 0x1ED80),)
# 表的前後各看這麼多 byte，判斷載入範圍有沒有蓋到。
NEAR = 0x40


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def window(start: int, end: int) -> list[int]:
    out, ea = [], start
    while ea < end:
        out.append(ea)
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def preceding(ea: int, count: int) -> list[int]:
    """往前取 count 條指令的位址。"""
    out, cur = [], ea
    for _ in range(count):
        prev = idc.prev_head(cur)
        if prev == idaapi.BADADDR or prev < 0:
            break
        out.append(prev)
        cur = prev
    return list(reversed(out))


def loader_calls() -> list[dict]:
    """每一個 sub_1A4FE 呼叫點，連同前面那一段的反組譯。"""
    out = []
    for ea in idautils.CodeRefsTo(LOADER, False):
        func = ida_funcs.get_func(ea)
        out.append({
            "call_ea": f"0x{ea:x}",
            "func": f"0x{func.start_ea:x}" if func else None,
            "before": [{"ea": f"0x{i:x}", "disassembly": disassembly(i)}
                       for i in preceding(ea, LOOKBACK)],
        })
    return sorted(out, key=lambda x: x["call_ea"])


def flag_users() -> list[dict]:
    """誰碰 byte_399D6——它決定三張表要用怪物自帶的還是預設的。"""
    out = []
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            if "399D6" in text:
                out.append({"func": f"0x{start:x}", "ea": f"0x{ea:x}",
                            "disassembly": text})
    return out


def writers() -> dict:
    """逐條掃全庫，找 operand 落在三張表附近的指令，並分出讀寫。"""
    out: dict[str, list[dict]] = {name: [] for name in TABLES}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            mnem = idc.print_insn_mnem(ea)
            for n in range(3):
                t = idc.get_operand_type(ea, n)
                if t not in (idc.o_imm, idc.o_mem, idc.o_displ):
                    continue
                value = idc.get_operand_value(ea, n)
                for candidate in (value, value - DGROUP_EA):
                    for name, base in TABLES.items():
                        if base - NEAR <= candidate <= base + NEAR:
                            out[name].append({
                                "func": f"0x{start:x}",
                                "ea": f"0x{ea:x}",
                                "mnem": mnem,
                                "operand": n,
                                "dgroup_offset": f"0x{candidate:04x}",
                                "writes": n == 0 and mnem in
                                          ("mov", "add", "sub", "and", "or", "xor", "inc", "dec"),
                                "disassembly": disassembly(ea),
                            })
                            break
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
        "schema": "kol-ida-koltown-weight-tables-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：三張基礎權重表的寫入者與檔案載入範圍；尚未鎖定任何斷言。",
        "tables": {name: {"dgroup_offset": f"0x{off:04x}",
                          "absolute": f"0x{DGROUP_EA + off:x}",
                          "bytes": (ida_bytes.get_bytes(DGROUP_EA + off, 0x20)
                                    or b"").hex()}
                   for name, off in TABLES.items()},
        "windows": {f"0x{a:x}-0x{b:x}": [{"ea": f"0x{i:x}",
                                          "disassembly": disassembly(i)}
                                         for i in window(a, b)]
                    for a, b in WINDOWS},
        "defaults": {name: {"dgroup_offset": f"0x{off:04x}",
                            "bytes": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                      or b"").hex()}
                     for name, (off, n) in DEFAULTS.items()},
        "loader_calls": loader_calls(),
        "near_table": writers(),
        "flag_users": flag_users(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
