#!/usr/bin/env python3
"""IDA Pro 9.4：野外隨機遭遇的擲點流程、六張小表與遭遇圖。

規則見 RE-078：時段變換之後擲一次，機率查 `TOWN0` 的 payload 4，怪物用地形與季節
兩張遮罩篩，數量與等級由隊伍平均位階和人數決定。機率表與遮罩已經抽成
`assets/original-derived-data/monsters.json`，剩下三張 DGROUP 小表還沒：

  0x133E  10 byte  平均位階 1…10 走任務怪物那一條的機率
  0x1348  5 word   級距 → 等級修正（−2 −1 0 +1 +2）
  0x1352  30 byte  級距 × 隊伍人數 → 數量上限

同時倒出 `sub_2A4E5`（遭遇）、`sub_2A337`（隊伍平均位階）與 `sub_2A975`（遭遇圖）
整支，加上三個視窗：`sub_1E3A8` 開頭那個「參數為 0 就跳過預設值」的閘門、敵人
等級寫進參戰者記錄的那一段，以及 `sub_29987` 裡呼叫遭遇的那一行。結論見 RE-097。

  idat -A -S"tools/ida_probe_koltown_ambush.py OUT.json INPUT.EXE" INPUT.i64
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
FUNCS = (0x2A4E5, 0x2A337, 0x2A975)
WINDOWS = (
    # sub_1E3A8 的閘門：參數為 0 就跳過整段預設值（怪物 0x11、一隻、等級 1）。
    (0x1E3A8, 0x1E420),
    # 敵人的參戰者等級 = word_3B26A ÷ 5 ＋ 1，上限 5。
    (0x1EE00, 0x1EE64),
    # 時段變了才擲遭遇，接在天候（sub_29B40）後面。
    (0x299F0, 0x29A60),
)
# 位移與長度，都以 DGROUP 為基準。
DUMP_BLOCKS = (
    (0x1312, 8),   # 地形位元表 80 40 20 10 08 04 02 01
    (0x1316, 4),   # 季節位元表 08 04 02 01
    (0x1332, 12),  # 月份 → 季節
    (0x133E, 10),  # 平均位階 → 任務怪物那一條的機率
    (0x1348, 10),  # 級距 → 等級修正（五個 word）
    (0x1352, 30),  # 級距 × 人數 → 數量上限
    (0x0B60, 6),   # 隨機遭遇的隊員站位：列
    (0x0B66, 6),   # 隨機遭遇的隊員站位：欄
    (0x0B8C, 72),  # 隨機遭遇的敵人站位：列，每張地圖 12 個
    (0x0BD4, 72),  # 隨機遭遇的敵人站位：欄
    (0x1370, 64),  # 遭遇圖：怪物 → 圖的偏移（每隻兩個 byte）
    (0x13B0, 24),  # 遭遇圖：第 i 隻畫在哪（每隻兩個 byte）
    (0x1840, 112),  # 檔名表：0x1844 與 0x184A 都在這一段
    (0x0670, 40),  # 副檔名表（sub_2D3FA 用 5 × word_2DF64 取）
)


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    receipt = {
        "schema": "kol-ida-koltown-ambush-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：隨機遭遇的三張小表與擲點流程；斷言見 RE-078。",
        "dgroup": f"0x{DGROUP_EA:x}",
        "blocks": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n) or b"").hex()
                   for off, n in DUMP_BLOCKS},
        "windows": {f"0x{a:x}-0x{b:x}": [{"ea": f"0x{i:x}", "disassembly": disassembly(i)}
                                        for i in range(a, b)
                                        if ida_bytes.is_head(ida_bytes.get_flags(i))]
                    for a, b in WINDOWS},
        "funcs": {
            f"0x{ea:x}": [{"ea": f"0x{i:x}", "disassembly": disassembly(i)}
                          for i in idautils.FuncItems(ea)]
            for ea in FUNCS if ida_funcs.get_func(ea) is not None
        },
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
