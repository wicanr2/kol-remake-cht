#!/usr/bin/env python3
"""IDA Pro 9.4：攻擊結算 `sub_26D76` 一次問完。

實跑已經問出入口與兩個結果（`docs/re/RESEARCH-LOG.md` 最後一節）：
`arg0 == 0` 走攻擊那一條、當前行動者索引在 `ds:9E5A`、結果落在 `ds:1C76`
（攻擊者 `+0x08` ÷10）與 `ds:44AC`（攻擊者 `+0x09`，不除）。

**問不出來的是對手**：改防禦者的欄位沒有反應，把整筆記錄塗成 0xFF 也一樣。
所以這一支要問：`sub_26D76` 讀哪些全域、指令碼怎麼分派、對手從哪裡來。

⚠ **一支探針一次問完。** 對同一個 `.i64` 連續開關會弄壞資料庫。壞掉的處置：
把 `.i64` 改名留證，`idat -A -B <EXE>` 重建，再拿一支跑過的探針比對輸出。

  idat -A -S"tools/ida_probe_koltown_attack_resolve.py OUT.json INPUT.EXE" INPUT.i64
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

RESOLVE_FN = 0x26D76
# 實跑已知的三個位址，要在反組譯裡找出它們的讀寫點。
CURRENT_SLOT = DGROUP_EA + 0x9E5A  # 當前行動者索引
SCORE_A = DGROUP_EA + 0x1C76       # 攻擊者 +0x08 ÷ 10
SCORE_B = DGROUP_EA + 0x44AC       # 攻擊者 +0x09
# RE-133 已經對過的子常式，看它們在結算裡的呼叫點。
SUBS = [0x26AC2, 0x24FD3, 0x23683, 0x2526C, 0x251B0, 0x22084]


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


def call_sites_in(host: int, callee: int) -> list[dict]:
    """callee 在 host 裡被呼叫的每一處，連前後文。"""
    f = ida_funcs.get_func(host)
    if f is None:
        return []
    items = list(idautils.FuncItems(host))
    out = []
    for i, ea in enumerate(items):
        if callee not in set(idautils.CodeRefsFrom(ea, False)):
            continue
        lo, hi = max(0, i - 12), min(len(items), i + 3)
        out.append({"ea": f"0x{ea:x}",
                    "context": [{"ea": f"0x{e:x}", "disassembly": disassembly(e)}
                                for e in items[lo:hi]]})
    return out


def touches(addr: int) -> list[dict]:
    out = []
    for ea in idautils.DataRefsTo(addr):
        f = ida_funcs.get_func(ea)
        out.append({"ea": f"0x{ea:x}",
                    "function": f"0x{f.start_ea:x}" if f else None,
                    "disassembly": disassembly(ea)})
    return out


def globals_used(start: int) -> list[dict]:
    """函式裡每一道指令參照到的資料位址（DGROUP 相對偏移一起算出來）。"""
    seen: dict[int, dict] = {}
    for ea in idautils.FuncItems(start):
        for ref in idautils.DataRefsFrom(ea):
            if ref in seen:
                seen[ref]["refs"].append(f"0x{ea:x}")
                continue
            seen[ref] = {"ea": f"0x{ref:x}",
                         "ds_offset": f"0x{ref - DGROUP_EA:x}" if ref >= DGROUP_EA else None,
                         "refs": [f"0x{ea:x}"]}
    return [seen[k] for k in sorted(seen)]


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-attack-resolve-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：攻擊結算讀哪些全域、指令碼怎麼分派、對手從哪裡來。",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "resolve_fn": summary(RESOLVE_FN),
        "resolve_listing": listing(RESOLVE_FN),
        "resolve_globals": globals_used(RESOLVE_FN),
        "sub_call_sites": {f"0x{s:x}": call_sites_in(RESOLVE_FN, s) for s in SUBS},
        "current_slot_refs": touches(CURRENT_SLOT),
        "score_a_refs": touches(SCORE_A),
        "score_b_refs": touches(SCORE_B),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2),
                        encoding="utf-8")


main()
