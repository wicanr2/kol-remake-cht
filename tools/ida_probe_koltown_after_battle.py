#!/usr/bin/env python3
"""IDA Pro 9.4：戰後結算——冒險點數、金冠、戰利品與升級。

獎勵整段都在 `sub_1E3A8` 裡（`0x21258`…`0x214B7`）：冒險點數把每一個參戰敵人的
攻防技能加起來除以 2 再除隊伍人數，金冠只讀第一隻怪的骰式擲 `word_3B26A` 輪。
`sub_289EE(0)` 接在後面，前半擲掉落物、後半是「Items Found:」那張清單。
`word_345DE` 不是 `0xFFFF` 時整段獎勵跳過，改走 `0x21611` 的升級——那是城鎮
訓練，不是野外戰鬥。結論見 RE-096。

本探針倒出那四段與它們用到的記錄指標、旗標，以及 `sub_22296`（怪物記錄）與
`sub_22866`（能不能收下戰利品）。

  idat -A -S"tools/ida_probe_koltown_after_battle.py OUT.json INPUT.EXE" INPUT.i64
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
# 想追的全域：記錄指標（0x39A48）、升級的觸發（0x345DE）、金冠的外層輪數
# （0x3B26A）、參戰者上界（0x2FBC0）、隊伍人數（0x2DF2A）與敵人數（0x30C6C）。
GLOBALS = (0x2DED0 + 0x6A98,)
# 想倒的函式（整支）與視窗（起、迄）。
FUNCS = ()
WINDOWS = (
    # 打贏那三句、冒險點數與金冠，最後呼叫 sub_289EE(0)。
    (0x211C0, 0x21500),
    # 打輸、逃走與升級。
    (0x21500, 0x21700),
    # sub_289EE 前半：掉落物；後半：兩張清單與領取。
    (0x289EE, 0x28F60),
    # sub_1E3A8 的戰場初始化：地圖載入、地形灑點與兩套站位。
    (0x1E41A, 0x1EA10),
    # 任務地點把站位複製到 0x2E0C 的那一段。
    (0x2B1A0, 0x2B2D0),
)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def line(ea: int) -> dict:
    return {"ea": f"0x{ea:x}", "disassembly": disassembly(ea)}


def window(start: int, end: int) -> list[dict]:
    out, ea = [], start
    while ea < end:
        out.append(line(ea))
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def refs(ea: int) -> dict:
    """對一個全域，把讀與寫分開列出，並標出所在函式。"""
    reads, writes = [], []
    for x in idautils.XrefsTo(ea, 0):
        f = ida_funcs.get_func(x.frm)
        item = {"ea": f"0x{x.frm:x}",
                "func": idc.get_func_name(x.frm) if f else None,
                "disassembly": disassembly(x.frm),
                "type": idautils.XrefTypeName(x.type)}
        (writes if x.type in (idaapi.dr_W,) else reads).append(item)
    return {"write": writes, "read": reads}


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-after-battle-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：戰後結算的冒險點數、掉落物與升級；尚未鎖定任何斷言。",
        "globals": {f"0x{ea:x}": refs(ea) for ea in GLOBALS},
        "callers": {f"0x{ea:x}": sorted({f"0x{c:x}" for c in idautils.CodeRefsTo(ea, False)})
                    for ea in FUNCS},
        "funcs": {f"0x{ea:x}": [line(i) for i in idautils.FuncItems(ea)] for ea in FUNCS},
        "windows": {f"0x{a:x}-0x{b:x}": window(a, b) for a, b in WINDOWS},
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
