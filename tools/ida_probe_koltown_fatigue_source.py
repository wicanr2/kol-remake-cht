#!/usr/bin/env python3
"""IDA Pro 9.4：`[0x322E + 2i]` 到底是誰在寫（RE-129 的續問）。

`sub_23683` 算的是 `word_2FB4A = 50 × [0x322E + 2i] ÷ [0x1C50 + 2i]`，而 remake
把它實作成 `50 × 累積傷害 ÷ 昏迷門檻`（參戰者記錄 `+0x0E` ÷ `+0x10`）。RE-129 已經
證實分母是「力量 ＋ 健康」，與 `+0x10`（(體型 ＋ 健康) ÷ 2）不是同一個東西，
所以分子也得重查：`[0x322E + 2i]` 是不是就是 `+0x0E`，還是另一個累計。

判準是**誰在寫它**。全庫掃 `322Eh` 的每一條指令，再把傷害落到身上那一支
（`sub_277F4`）與回合收尾整段倒出來逐條看。

  idat -A -S"tools/ida_probe_koltown_skill_penalty.py OUT.json INPUT.EXE" INPUT.i64
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
# 攻防技能的取用點、本體記錄的讀取、疲勞折算。
DUMP_FUNCTIONS = (0x23683, 0x277F4, 0x21AC4)
# 法術效果落到欄位的那一段落在哪一支裡，由 get_func 補；窗口另外取。
WINDOWS = {
    # 每回合的衰減與休息那一條，都掛在 `sub_1E3A8` 裡。
    "per_turn_decay": (0x20D40, 0x20DC0),
    "rest_command": (0x20AE0, 0x20B60),
}
# 誰碰這幾個全域；直接掃反組譯文字，連相對定址的字面值一起找。
TRACKED_TOKENS = ("word_2FB4A", "1C50h", "322Eh")
# 全庫掃描：碰 `[bx+322Eh]` 的每一條指令。正對照是 `sub_23683` 那條讀取與
# `sub_21AC4` 的歸零——兩條都必須出現，沒出現就是這個掃描有洞。
DIRECT_FIELD_RE = r"322Eh\]"
DGROUP_DUMPS = ((0x322E, 40),)  # 分子，一個參戰者兩個 byte


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def listing(start: int) -> list[dict]:
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}
            for ea in idautils.FuncItems(start)]


def window(start: int, end: int) -> list[dict]:
    out = []
    ea = start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def summary(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
        "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)}),
    }


def token_owners() -> dict:
    out: dict[str, dict[str, int]] = {}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            for name in TRACKED_TOKENS:
                if name in text:
                    bucket = out.setdefault(name, {})
                    key = f"0x{start:x}"
                    bucket[key] = bucket.get(key, 0) + 1
    return out


def direct_field_touches() -> list[dict]:
    out = []
    pattern = re.compile(DIRECT_FIELD_RE)
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

    hosts = {}
    for name, (start, _) in WINDOWS.items():
        function = ida_funcs.get_func(start)
        hosts[name] = (summary(function.start_ea) if function
                       else {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"})

    receipt = {
        "schema": "kol-ida-koltown-fatigue-source-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：`[0x322E + 2i]` 的寫入點普查；尚未鎖定任何斷言。",
        "functions": [summary(ea) for ea in DUMP_FUNCTIONS],
        "listings": {f"0x{ea:x}": listing(ea) for ea in DUMP_FUNCTIONS},
        "window_hosts": hosts,
        "windows": {name: window(*bounds) for name, bounds in WINDOWS.items()},
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
        "token_owners": token_owners(),
        "direct_field_touches": direct_field_touches(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
