#!/usr/bin/env python3
"""IDA Pro 9.4：攻防技能的罰值（參戰者 `+0x2E`／`+0x2F`）與疲勞的分母。

法術效果 7、8 寫的是參戰者記錄的 `+0x2E`／`+0x2F`（`0x28704` 查
`[0x12A4 + 效果 − 1]`）。要接進 remake 之前得先確定三件事：

1. `0x28704` 那一段的加減方向與夾限——正面加、負面減，還是反過來。
2. `sub_26AC2`／`sub_26CD5` 怎麼把那兩格折進 `word_34A12`（攻擊技能）與
   `word_32380`（防禦技能）：加還是減、倍率多少、有沒有下限。
3. 疲勞折算 `sub_23683` 的分母 `[0x1C50 + 2i]` 是誰填的。

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
DUMP_FUNCTIONS = (0x26AC2, 0x26CD5, 0x26C5F, 0x23683, 0x22229, 0x22296, 0x22278,
                  # 法術落到欄位的那一支整支倒出來（效果 1…8 都在裡面）。
                  0x2818D,
                  # 疲勞分母 `[0x1C50 + 2i]` 的三個寫入點。
                  0x21AC4, 0x2324F, 0x236AB,
                  # 攻擊結算：攻防技能與疲勞折算實際被用掉的地方。
                  0x26D76,
                  # 法術過期時把屬性搬回來的兩支：隊員與敵人各一。
                  0x2198A, 0x2215A)
# 法術效果落到欄位的那一段落在哪一支裡，由 get_func 補；窗口另外取。
WINDOWS = {
    "cast_stat": (0x28704, 0x28790),
    "cast_fatigue": (0x286AB, 0x28704),
}
# 誰碰這幾個全域；直接掃反組譯文字，連相對定址的字面值一起找。
TRACKED_TOKENS = ("word_34A12", "word_32380", "word_3B26A", "word_2FB4A",
                  "1C50h", "12A4h", "2Eh]", "2Fh]",
                  # 參戰者記錄的基底：誰整批清過這一段。
                  "6712h")
# 全庫掃描：直接以 `[reg+2Eh]`／`[reg+2Fh]` 觸碰參戰者那兩格的指令。`322Eh` 這種
# 字尾相同的位移會混進純字串比對，所以用正規式框住前面的加號（「查詢回空」的正對照：
# `sub_26CD5` 的兩條讀取必須出現在結果裡，沒出現就是這個掃描有洞）。
DIRECT_FIELD_RE = r"\+2[EF]h\]"
DGROUP_DUMPS = (
    (0x12A4, 8),    # 效果 → 參戰者記錄欄位偏移
    (0x1C50, 40),   # 疲勞折算的分母，一個參戰者兩個 byte
)


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
        "schema": "kol-ida-koltown-skill-penalty-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：攻防技能罰值與疲勞分母；尚未鎖定任何斷言。",
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
