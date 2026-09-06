#!/usr/bin/env python3
"""IDA Pro 9.4：一手的時間額度 `word_30D06` 從算出來到被用掉的那一整段。

`sub_23683` 算的是 `word_2FB4A = 50 × [0x322E + 2i] ÷ [0x1C50 + 2i]`，而 remake
把它實作成 `50 × 累積傷害 ÷ 昏迷門檻`（參戰者記錄 `+0x0E` ÷ `+0x10`）。RE-129 已經
證實分母是「力量 ＋ 健康」，與 `+0x10`（(體型 ＋ 健康) ÷ 2）不是同一個東西，
所以分子也得重查：`[0x322E + 2i]` 是不是就是 `+0x0E`，還是另一個累計。

RE-130 把疲勞的來源縮到兩條：法術效果 5，以及每一手收尾的
`疲勞 −= word_30D06 ÷ 10`（額度為負就加倍，所以疲勞反而上升）。後者是「受傷 →
疲勞 → 撤退」那條迴路的唯一入口，不接的話敵人永遠不會撤退。

`word_30D06` 由 `sub_236AB` 在一手開始時算好（`0x20B22`），到 `0x20D40` 才被讀，
中間那五百多個 byte 就是花掉它的地方。把那一整段連同提到它的每一條指令倒出來，
才知道 remake 要不要整套時間額度。

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
DUMP_FUNCTIONS = (0x236AB, 0x2526C, 0x251B0,
                  # 指令成本的第二個乘數，以及兩支拿額度當權重的抽樣器。
                  0x268D4, 0x2578F, 0x2590F)
# 法術效果落到欄位的那一段落在哪一支裡，由 get_func 補；窗口另外取。
WINDOWS = {
    # 額度算好之後到被讀掉之間那一段。
    "spend": (0x20B22, 0x20D40),
    # 比例到 100 之後那一句（`0x20D9A` 的分支）。
    "exhausted": (0x20DA4, 0x20E70),
}
# 誰碰這幾個全域；直接掃反組譯文字，連相對定址的字面值一起找。
TRACKED_TOKENS = ("word_30D06", "word_2FBB6", "word_2F940")
# 全庫掃描：提到 `word_30D06`／`word_2FBB6` 的每一條指令，各帶前後文。正對照是
# `sub_236AB` 的兩條寫入與 `0x20D40` 的讀取，沒出現就是這個掃描有洞。
DIRECT_FIELD_RE = r"word_30D06|word_2FBB6"
# 每一條命中前後各帶這麼多條指令，好一次看清楚它被算進哪一條式子。
CONTEXT = 14
DGROUP_DUMPS = (
    # 指令與姿態的時間成本表，一整塊倒出來免得為了少一張表再開一次資料庫。
    (0x0FD2, 16),   # 移動方式圖號 → 坐騎記錄裡的位移（`sub_251B0:0x25226`）
    (0x10B2, 112),  # 0x10B2 指令成本、0x10BE 姿態乘數、0x10E4 移動方式乘數、
                    # 0x10EA 護符成本、0x10F8 效果成本、0x1102 骰式成本
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
        "schema": "kol-ida-koltown-time-allowance-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：一手的時間額度；尚未鎖定任何斷言。",
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
