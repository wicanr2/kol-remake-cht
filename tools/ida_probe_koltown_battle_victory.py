#!/usr/bin/env python3
"""IDA Pro 9.4：戰鬥打贏之後——怪物計數、勝負判定、戰利品與升級。

remake 目前只判「隊伍全倒」（SPEC-008），把怪物全部打死沒有判定，也沒有戰後結算。
入口是 `sub_289EE`（戰利品與升級），訊息在 KOLDATA 372…384（`You have defeated|F`、
`receives|N adventure points, and`、`|N gold crowns!`、`you have gained a level!`）。
本探針倒出那一支與它的 callee，並開三個窗看剩餘怪物數 `word_3B26A` 怎麼加減，
另外掃全庫找誰碰勝負相關的幾個全域。輸出只有位址、反組譯文字與資料位元組。

  idat -A -S"tools/ida_probe_koltown_battle_victory.py OUT.json INPUT.EXE" INPUT.i64
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
VICTORY = 0x289EE            # 戰利品與升級
DUMP_FUNCS = (VICTORY, 0x28F5D, 0x288D4, 0x26C5F, 0x22278)
# 剩餘怪物數在 `sub_1E3A8` 裡被減；三個窗分別是減的那一段與它前後的判定。
WINDOWS = ((0x21360, 0x21420), (0x211A0, 0x21260), (0x21420, 0x21520),
           (0x21600, 0x216B0), (0x21260, 0x21300),
           (0x20FA0, 0x21040), (0x21300, 0x21365), (0x21520, 0x21600),
           (0x28B1B, 0x28BE0))
# 勝負相關的全域：剩餘怪物數、隊伍人數、兩個還沒定語意的旗標。
TRACKED = ("word_3B26A", "word_30C6C", "word_34966", "word_345DE",
           "word_345E0", "word_2FBBC", "sub_289EE", "sub_28F5D")
DATA_WINDOW = 32
# 要原樣倒出來的 DGROUP 區塊：(位移, 位元組數)。
DGROUP_DUMPS = ((0x00D4, 8), (0x185E, 16), (0x1B2A, 16))


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def summary(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        return {"start_ea": f"0x{start:x}", "note": "沒有函式邊界"}
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{f.end_ea:x}",
        "instruction_count": len(list(idautils.FuncItems(start))),
        "callers": sorted({f"0x{ea:x}" for ea in idautils.CodeRefsTo(start, False)}),
    }


def listing(start: int, end: int | None = None) -> list[dict]:
    if end is None:
        items = list(idautils.FuncItems(start))
    else:
        items, ea = [], start
        while ea < end:
            items.append(ea)
            size = ida_bytes.get_item_size(ea)
            ea += size if size > 0 else 1
    return [{"ea": f"0x{ea:x}", "disassembly": disassembly(ea)} for ea in items]


def name_users() -> dict:
    out: dict[str, list[dict]] = {n: [] for n in TRACKED}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            for name in TRACKED:
                if name in text:
                    out[name].append({"func": f"0x{start:x}", "ea": f"0x{ea:x}",
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

    callees: list[int] = []
    data_refs: dict[int, list[str]] = {}
    for ea in idautils.FuncItems(VICTORY):
        if idc.print_insn_mnem(ea) == "call":
            for target in idautils.CodeRefsFrom(ea, False):
                f = ida_funcs.get_func(target)
                if f is not None and f.start_ea == target:
                    callees.append(target)
        for ref in idautils.DataRefsFrom(ea):
            data_refs.setdefault(ref, []).append(f"0x{ea:x}")

    receipt = {
        "schema": "kol-ida-koltown-battle-victory-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：戰鬥勝負判定與戰後結算；尚未鎖定任何斷言。",
        "functions": [summary(ea) for ea in DUMP_FUNCS],
        "callees": [summary(ea) for ea in sorted(set(callees))],
        "listings": {f"0x{ea:x}": listing(ea) for ea in DUMP_FUNCS},
        "windows": {f"0x{a:x}-0x{b:x}": listing(a, b) for a, b in WINDOWS},
        "data_refs": [
            {
                "ea": f"0x{ref:x}",
                "dgroup_offset": (f"0x{ref - DGROUP_EA:04x}" if ref >= DGROUP_EA else None),
                "name": idc.get_name(ref) or None,
                "read_at": sorted(set(sites)),
                "bytes": (ida_bytes.get_bytes(ref, DATA_WINDOW) or b"").hex(),
            }
            for ref, sites in sorted(data_refs.items())
        ],
        "name_users": name_users(),
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
