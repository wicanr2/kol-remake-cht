#!/usr/bin/env python3
"""IDA Pro 9.4：戰鬥規劃段裡敵人那一手的決策函式。

規劃段輪到敵人時，`sub_1E3A8:0x1F0FC` 呼叫 `sub_2374C`，由它把指令碼、攻擊選項與
防禦選項寫進參戰者記錄的 +0x25／+0x23／+0x24（RE-080）。本探針逐條倒出這一支，
附上它呼叫的每一支的大小與呼叫者，並把它讀到的 DGROUP 位址與附近位元組原樣附上，
供逐條對碼。輸出只有位址、反組譯文字與資料表位元組，不含原版可執行內容。

  idat -A -S"tools/ida_probe_koltown_enemy_ai.py OUT.json INPUT.EXE" INPUT.i64
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
ENEMY_AI = 0x2374C
DGROUP_EA = 0x2DED0
# 決策細節在這幾支裡；rand、畫框、long 乘除與已解出的視線／方向格不重倒。
DUMP_CALLEES = (0x22296, 0x23683, 0x236AB, 0x24E2B, 0x24EF9, 0x251B0, 0x2526C,
                0x252CB, 0x2531C, 0x253E7, 0x2578F, 0x2590F, 0x25D0F,
                # 第二層：輪盤、環格查詢與被抄來的目標索引。
                0x25A3A, 0x24BB6, 0x24C35, 0x25653)
# 要原樣倒出來的 DGROUP 區塊：(位移, 位元組數)。
DGROUP_DUMPS = (
    (0x0FC0, 8),     # 每一環的格數（`sub_2531C:0x25332`）
    (0x0FC8, 10),    # 磚塊 → 擋路的位元圖（`sub_24E2B:0x24ED3`）
    (0x0FD2, 6),     # 徒手選項 → 裝備欄位移（`sub_251B0:0x25226`）
    (0x0FD8, 14),    # 每一環的 X 位移陣列指標
    (0x0FE6, 14),    # 每一環的 Y 位移陣列指標
    (0x0FF4, 160),   # 攻擊式／部位權重，一列 10 byte
    (0x109C, 16),    # 拳腳可用性遮罩
    (0x10B2, 16),    # 指令的時間成本
    (0x1A9A, 8),     # 部位的基礎權重（`sub_2590F:0x25923`）
    (0x6AD2, 16),    # 攻擊式的基礎權重（`sub_2578F:0x257B0`）
    (0x10C6, 16),    # 加權輪盤的等級除數（`sub_25A3A:0x25AD1`）
    (0x10D8, 12),    # 相對格差 → 方向索引（`sub_2374C:0x23FDF`）
)
# 環位移陣列本身在指標後面，要解一層才看得到。
RING_POINTERS = (0x0FD8, 0x0FE6)
RING_COUNT_TABLE = 0x0FC0
RING_COUNT = 7
# 這幾個執行期變數決定移動與權重，但寫在別支函式裡；掃全庫找出誰碰它們。
TRACKED_GLOBALS = ("byte_3496A", "byte_3496B", "byte_3496C", "byte_37D11",
                   "byte_2EF75", "word_349B2", "word_2FBC0", "word_34A08")
# 呼叫端：規劃段輪到敵人的那一段，看三個全域怎麼寫回參戰者記錄。
CALLER_WINDOW = (0x1F0C0, 0x1F170)
# 資料參照往後倒這麼多位元組，足以看出一張表的頭幾筆。
DATA_WINDOW = 32


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


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


def caller_window(start: int, end: int) -> list[dict]:
    out = []
    ea = start
    while ea < end:
        out.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        size = ida_bytes.get_item_size(ea)
        ea += size if size > 0 else 1
    return out


def ring_offsets() -> dict:
    """環位移表：每一環的格數在 0x0FC0，X／Y 位移各由一張指標表指出去。"""
    out = {}
    for ring in range(RING_COUNT):
        count = ida_bytes.get_byte(DGROUP_EA + RING_COUNT_TABLE + ring)
        entry = {"count": count}
        for label, table in (("x", RING_POINTERS[0]), ("y", RING_POINTERS[1])):
            pointer = ida_bytes.get_word(DGROUP_EA + table + 2 * ring)
            entry[f"{label}_pointer"] = f"0x{pointer:04x}"
            entry[label] = [ida_bytes.get_word(DGROUP_EA + pointer + 2 * i)
                            for i in range(count)]
        out[str(ring)] = entry
    return out


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

    function = ida_funcs.get_func(ENEMY_AI)
    if function is None or function.start_ea != ENEMY_AI:
        raise RuntimeError(f"function boundary not found at 0x{ENEMY_AI:x}")

    listing = []
    callees: list[int] = []
    data_refs: dict[int, list[str]] = {}
    for ea in idautils.FuncItems(ENEMY_AI):
        listing.append({"ea": f"0x{ea:x}", "disassembly": disassembly(ea)})
        if idc.print_insn_mnem(ea) == "call":
            for target in idautils.CodeRefsFrom(ea, False):
                callee = ida_funcs.get_func(target)
                if callee is not None and callee.start_ea == target:
                    callees.append(target)
        for ref in idautils.DataRefsFrom(ea):
            data_refs.setdefault(ref, []).append(f"0x{ea:x}")

    receipt = {
        "schema": "kol-ida-koltown-enemy-ai-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：規劃段敵人決策函式的逐條反組譯；尚未鎖定任何斷言。",
        "root": summary(ENEMY_AI),
        "callees": [summary(ea) for ea in sorted(set(callees))],
        "data_refs": [
            {
                "ea": f"0x{ref:x}",
                "dgroup_offset": (f"0x{ref - DGROUP_EA:04x}"
                                  if ref >= DGROUP_EA else None),
                "name": idc.get_name(ref) or None,
                "read_at": sorted(set(sites)),
                "bytes": (ida_bytes.get_bytes(ref, DATA_WINDOW) or b"").hex(),
            }
            for ref, sites in sorted(data_refs.items())
        ],
        "listing": listing,
        "callee_listings": {
            f"0x{ea:x}": [{"ea": f"0x{i:x}", "disassembly": disassembly(i)}
                          for i in idautils.FuncItems(ea)]
            for ea in DUMP_CALLEES
        },
        "caller_window": caller_window(*CALLER_WINDOW),
        "dgroup_dumps": {f"0x{off:04x}": (ida_bytes.get_bytes(DGROUP_EA + off, n)
                                          or b"").hex()
                         for off, n in DGROUP_DUMPS},
        "ring_offsets": ring_offsets(),
        "global_owners": global_owners(),
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
