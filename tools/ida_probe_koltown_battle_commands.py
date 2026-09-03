#!/usr/bin/env python3
"""IDA Pro 9.4：戰場上玩家那一手——主指令列、選目標格與三個子選單。

`sub_228B7` 逐項看裝備、傷勢與腳下有什麼來決定主指令列要放哪幾格，回傳玩家選中的
9×9 視窗格索引（圖號留在 `word_37CF2`）。`sub_24131` 是選目標格的互動迴圈。攻擊式、
移動方式與防禦方式各有一列子選單，掛在 `sub_1E3A8` 裡。本探針把這些逐條倒出來，
附上圖示列元件與相關 DGROUP 表，供逐條對碼。

  idat -A -S"tools/ida_probe_koltown_battle_commands.py OUT.json INPUT.EXE" INPUT.i64
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
# 主指令列、選目標格、圖示列元件與幾支條件判斷。
DUMP_FUNCTIONS = (0x228B7, 0x24131, 0x11703, 0x11722, 0x1449C, 0x227C0, 0x22278,
                  0x24AF8, 0x24A02,
                  # 執行段：移動那一手。
                  0x26384,
                  # 訊息入口：`|X` 代換碼與 `]` 都在這一支處理。
                  0x1A8FD,
                  # `|X` 的代換：戰鬥訊息的名字與代名詞都由這一支填。
                  0x265F3,
                  # 每一手開頭印行動者名字的那一支；`|G` 走的也是它。
                  0x2685A, 0x2687F,
                  # 訊息框：清、送字元、收尾。
                  0x1C1A4, 0x1C1BF, 0x1C2DA, 0x1C311, 0x1103F,
                  # 傷害落到部位，順便決定傷害輕重那一句。
                  0x277F4)
# 訊息框的游標與版面放在這幾個全域裡；掃全庫找出誰設它們。
TRACKED_GLOBALS = ("word_2DF14", "word_2DF16", "word_2DF18", "word_2FB58",
                   "word_345E0", "word_2E57C")
# 三個子選單掛在 `sub_1E3A8` 裡，用位址視窗取。
WINDOWS = {
    # 主指令列選完之後的分派表把圖號送到這幾段。
    "command_branches": (0x1F3B0, 0x1F6BA),
    "unarmed_forms": (0x1F6BA, 0x1F7D9),
    "movement_forms": (0x1F7D9, 0x1F940),
    "weapon_forms": (0x1F940, 0x1FB38),
    # 執行段：另外排一次順序，鍵值大的先動（RE-080）。
    "execution": (0x1FD00, 0x20900),
    # 攻擊結算的開頭：目標是怎麼找出來的。
    "attack_target": (0x26D76, 0x26E60),
    # EGA 的字元繪製：word_2DF14／word_2DF16 的單位在這裡定。
    "glyph_draw": (0x1DE84, 0x1DF40),
    # 攻擊訊息：哪一句對到哪一種攻擊式。
    "attack_messages": (0x270A0, 0x27200),
    "attack_hit": (0x27200, 0x27620),
}
# 要原樣倒出來的 DGROUP 區塊：(位移, 位元組數)。
DGROUP_DUMPS = (
    (0x0B42, 8),     # 徒手圖號 − 0x2A → 攻擊選項
    (0x0B76, 9),     # 方向 0…7 → 3×3 格編號
    (0x0C3C, 18),    # 環 0 的 Y 位移，多倒一格看第 9 項是什麼
    (0x0C4E, 18),    # 環 0 的 X 位移，同上
    (0x2BAE, 16),    # 地上東西的 X（`sub_228B7:0x22C1D`）
    (0x2D9E, 16),    # 地上東西的 Y（`sub_228B7:0x22C39`）
    (0x0B6C, 9),     # 3×3 格編號 → 方向，0x0B76 的反表（`sub_24131:0x24987`）
    (0x110A, 128),   # 攻擊選項修正，一組 14 byte
    (0x10B2, 16),    # 指令的時間成本（執行段的鍵值要用）
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


def global_owners() -> dict:
    out: dict[str, dict[str, int]] = {}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            text = disassembly(ea)
            for name in TRACKED_GLOBALS:
                if name in text:
                    bucket = out.setdefault(name, {})
                    key = f"0x{start:x}"
                    bucket[key] = bucket.get(key, 0) + 1
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
        "schema": "kol-ida-koltown-battle-commands-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：玩家那一手的指令列與子選單；尚未鎖定任何斷言。",
        "functions": [summary(ea) for ea in DUMP_FUNCTIONS],
        "listings": {f"0x{ea:x}": listing(ea) for ea in DUMP_FUNCTIONS},
        "windows": {name: window(*bounds) for name, bounds in WINDOWS.items()},
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
        "global_owners": global_owners(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
