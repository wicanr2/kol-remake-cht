#!/usr/bin/env python3
"""IDA Pro 9.4：`word_2FB4A`（疲勞折算）在哪幾個地方被讀掉。

`sub_23683` 算的是 `word_2FB4A = 50 × [0x322E + 2i] ÷ [0x1C50 + 2i]`，而 remake
把它實作成 `50 × 累積傷害 ÷ 昏迷門檻`（參戰者記錄 `+0x0E` ÷ `+0x10`）。RE-129 已經
證實分母是「力量 ＋ 健康」，與 `+0x10`（(體型 ＋ 健康) ÷ 2）不是同一個東西，
所以分子也得重查：`[0x322E + 2i]` 是不是就是 `+0x0E`，還是另一個累計。

`[0x322E + 2i]` 的寫入點已經普查完（只有法術效果 5 會加）。接著要知道
`word_2FB4A` 被**讀**到哪幾條式子裡，才能把 remake 那幾個折算點逐一對上。
掃全庫每一條提到它的指令，每一條各帶前後文——一次問完，不要為了看上下文再開一次
資料庫。

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
DUMP_FUNCTIONS = (0x23683,)
# 法術效果落到欄位的那一段落在哪一支裡，由 get_func 補；窗口另外取。
WINDOWS = {}
# 誰碰這幾個全域；直接掃反組譯文字，連相對定址的字面值一起找。
TRACKED_TOKENS = ("word_2FB4A",)
# 全庫掃描：提到 `word_2FB4A` 的每一條指令，各帶前後文。正對照是 `sub_23683`
# 那條唯一的寫入，沒出現就是這個掃描有洞。
DIRECT_FIELD_RE = r"word_2FB4A"
# 每一條命中前後各帶這麼多條指令，好一次看清楚它被算進哪一條式子。
CONTEXT = 14
DGROUP_DUMPS = ()


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
        items = list(idautils.FuncItems(start))
        texts = [disassembly(ea) for ea in items]
        for n, text in enumerate(texts):
            if not pattern.search(text):
                continue
            lo, hi = max(0, n - CONTEXT), min(len(items), n + CONTEXT + 1)
            out.append({
                "function": f"0x{start:x}", "ea": f"0x{items[n]:x}",
                "disassembly": text,
                "context": [{"ea": f"0x{items[k]:x}", "disassembly": texts[k]}
                            for k in range(lo, hi)],
            })
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
        "schema": "kol-ida-koltown-fatigue-readers-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：`word_2FB4A` 的讀取點普查；尚未鎖定任何斷言。",
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
