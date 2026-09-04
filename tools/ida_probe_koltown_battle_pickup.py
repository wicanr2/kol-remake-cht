#!/usr/bin/env python3
"""IDA Pro 9.4：戰場上的撿與丟在規劃段哪裡處理。

RE-093 證實執行段的分派只認十個指令碼，撿（0x15）與丟（0x24）都不在——它們是規劃段
的即時動作。這支探針用**訊息記錄號**反查處理點，比掃指令碼準：

| 記錄 | 文字 |
|---|---|
| 0x10C | `cannot drop\|B\|E!` |
| 0x10D | `drops\|B\|E!` |
| 0x111 | `You have found a\|H.  Do you want to pick it up?` |
| 0x112 | `\|G, you must drop or belt your\|E to pick this up!` |
| 0x113 | `\|G, this is the quest item!  Run for the exit!` |
| 0x114 | `puts\|B\|I in\|B belt` |
| 0x183 | `You empty your hands and pick up the item.` |

`sub_1A8FD(記錄, 0x9E5C)` 的呼叫點前兩三條就會有 `mov ax, <記錄>`，逐條掃文字即可。

  idat -A -S"tools/ida_probe_koltown_battle_pickup.py OUT.json INPUT.EXE" INPUT.i64
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
# 撿與丟會用到的訊息記錄。
RECORDS = (0x10C, 0x10D, 0x111, 0x112, 0x113, 0x114, 0x183)
# 指令碼：撿、丟與三個語意還沒對上的。
COMMANDS = (0x15, 0x16, 0x24, 0x25, 0x26)
# 找到之後往前後各開這麼多條指令的窗。
CONTEXT = 12


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def neighbourhood(ea: int, before: int, after: int) -> list[dict]:
    items, cur = [], ea
    for _ in range(before):
        prev = idc.prev_head(cur)
        if prev == idaapi.BADADDR or prev < 0:
            break
        items.append(prev)
        cur = prev
    items = list(reversed(items)) + [ea]
    cur = ea
    for _ in range(after):
        nxt = idc.next_head(cur, cur + 0x40)
        if nxt == idaapi.BADADDR:
            break
        items.append(nxt)
        cur = nxt
    return [{"ea": f"0x{i:x}", "disassembly": disassembly(i)} for i in items]


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py OUT.json INPUT.EXE")
    out_path, input_path = Path(idc.ARGV[1]), Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    record_hits: dict[str, list[dict]] = {f"0x{r:x}": [] for r in RECORDS}
    command_hits: dict[str, list[dict]] = {f"0x{c:x}": [] for c in COMMANDS}
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            if idc.print_insn_mnem(ea) != "mov" and idc.print_insn_mnem(ea) != "cmp":
                continue
            for n in range(2):
                if idc.get_operand_type(ea, n) != idc.o_imm:
                    continue
                v = idc.get_operand_value(ea, n)
                if idc.print_insn_mnem(ea) == "mov" and v in RECORDS:
                    record_hits[f"0x{v:x}"].append({
                        "func": f"0x{start:x}", "ea": f"0x{ea:x}",
                        "context": neighbourhood(ea, 3, CONTEXT)})
                if idc.print_insn_mnem(ea) == "cmp" and v in COMMANDS \
                        and "31122" not in disassembly(ea):
                    command_hits[f"0x{v:x}"].append({
                        "func": f"0x{start:x}", "ea": f"0x{ea:x}",
                        "disassembly": disassembly(ea)})

    receipt = {
        "schema": "kol-ida-koltown-battle-pickup-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "note": "探索用輸出：撿與丟在規劃段的處理點；尚未鎖定任何斷言。",
        "by_record": record_hits,
        "by_command": command_hits,
    }
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n",
                        encoding="utf-8")
    idc.qexit(0)


main()
