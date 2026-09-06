#!/usr/bin/env python3
"""IDA Pro 9.4：`sub_2215A`（怪物記錄 bit0 的唯一讀取點）與雲層動畫的呼叫端，一次問完。

上一支探針（`ida_probe_koltown_clouds_and_monster_flag.py`）掃出七個「讀某個暫存器
`+10h`」的位置，其中 `0x22172` 把整個旗標 byte 餵進 `sub_22376(值, 0, 1)`——依 RE-118
已解的公式那就是 **bit0**，也就是 CYCLOPS／CLIFF TROLL／SLEDGE／MIST GRUB 那一位的
唯一讀取點。但 `sub_2215A` 只有摘要沒有 listing，語意還是懸著。

dosgolem 那邊實跑過一次（設 `word_39A48` 指的記錄 `+0x10` 為 0 vs 1，差分整個
DGROUP）：**兩組零差異**，連 `word_2FB4E` 都沒變——代表 `sub_22296` 就把預算吃光了，
差分問不到答案。所以回頭靜態讀。

同一支順便問雲層動畫（`sub_2A18E`）的**呼叫端條件**：版面已經讀出來了（EGA 每列
右旋 1 位元、Tandy 每列右旋 1 個 nibble），缺的是「什麼狀態下會被呼叫」，那要看
`0x11593` 的前後文與 `word_34A1A`／`word_2DF1E` 的寫入點。

⚠ **一支探針一次問完。** 對同一個 `.i64` 連續開關會弄壞資料庫。

  idat -A -S"tools/ida_probe_koltown_iconbar_and_cloudcaller.py OUT.json INPUT.EXE" INPUT.i64
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

# bit0 那一條鏈。
ICONBAR_FN = 0x2215A      # 讀 word_39A48+0x10 → sub_22376(值, 0, 1)
ICONBAR_HELPER = 0x22296  # sub_2215A 進門就呼叫的那一支
BITFIELD_FN = 0x22376     # 位元擷取器（RE-118 解過，這裡取原文覆核）
GATE_FN = 0x22229         # 「碎成一地」那一句的前置判斷
ICONBAR_CALLERS = (0x20F30, 0x22132)

# 雲層動畫那一條鏈。
CLOUD_FN = 0x2A18E
CLOUD_CALLER = 0x11593
VIDEO_MODE_EA = 0x2DF1E   # word_2DF1E
CLOUD_BUF_EA = 0x34A1A    # word_34A1A


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


def globals_used(start: int) -> list[dict]:
    seen: dict[int, dict] = {}
    for ea in idautils.FuncItems(start):
        for ref in idautils.DataRefsFrom(ea):
            if ref in seen:
                seen[ref]["refs"].append(f"0x{ea:x}")
                continue
            seen[ref] = {"ea": f"0x{ref:x}",
                         "name": idc.get_name(ref) or None,
                         "ds_offset": f"0x{ref - DGROUP_EA:x}" if ref >= DGROUP_EA else None,
                         "refs": [f"0x{ea:x}"]}
    return [seen[k] for k in sorted(seen)]


def context(ea: int, before: int = 12, after: int = 12) -> list[dict]:
    f = ida_funcs.get_func(ea)
    if f is None:
        return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}]
    items = list(idautils.FuncItems(f.start_ea))
    if ea not in items:
        return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}]
    i = items.index(ea)
    lo, hi = max(0, i - before), min(len(items), i + after + 1)
    return [{"ea": f"0x{e:x}", "disassembly": disassembly(e)} for e in items[lo:hi]]


def refs_to_global(ea: int) -> list[dict]:
    """全域的讀寫點：xref 對絕對定址是有效的，附前後文才看得出是讀是寫。"""
    out = []
    for ref in idautils.DataRefsTo(ea):
        fn = ida_funcs.get_func(ref)
        out.append({"ea": f"0x{ref:x}",
                    "function": f"0x{fn.start_ea:x}" if fn else None,
                    "disassembly": disassembly(ref),
                    "context": context(ref, before=6, after=6)})
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
        "schema": "kol-ida-koltown-iconbar-cloudcaller-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "怪物記錄 +0x10 bit0 的唯一讀取點 sub_2215A，以及雲層動畫的呼叫端條件。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "iconbar_fn": summary(ICONBAR_FN),
        "iconbar_listing": listing(ICONBAR_FN),
        "iconbar_globals": globals_used(ICONBAR_FN),
        "iconbar_caller_contexts": [
            {"ea": f"0x{ea:x}", "context": context(ea, before=20, after=12)}
            for ea in ICONBAR_CALLERS],
        "helper_fn": summary(ICONBAR_HELPER),
        "helper_listing": listing(ICONBAR_HELPER),
        "helper_globals": globals_used(ICONBAR_HELPER),
        "bitfield_fn": summary(BITFIELD_FN),
        "bitfield_listing": listing(BITFIELD_FN),
        "gate_fn": summary(GATE_FN),
        "gate_listing": listing(GATE_FN),
        "cloud_fn": summary(CLOUD_FN),
        "cloud_caller_context": context(CLOUD_CALLER, before=40, after=24),
        "video_mode_refs": refs_to_global(VIDEO_MODE_EA),
        "cloud_buf_refs": refs_to_global(CLOUD_BUF_EA),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
