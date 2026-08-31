"""IDA Pro 9.4：在受驗證的 KOLTOWN 解包衍生檔建立真正入口函式。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_entry
import ida_funcs
import ida_lines
import ida_nalt
import ida_ua
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: ida_probe_unpacked_entry.py PRIVATE.json SUMMARY.json")
    private_output = Path(idc.ARGV[1])
    summary_output = Path(idc.ARGV[2])
    input_path = Path(ida_nalt.get_input_file_path())
    raw = input_path.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")

    ordinal = ida_entry.get_entry_ordinal(0)
    entry_ea = ida_entry.get_entry(ordinal)
    if entry_ea != 0x2B90C:
        raise RuntimeError(f"unexpected entry EA: 0x{entry_ea:x}")

    ida_bytes.del_items(entry_ea, ida_bytes.DELIT_SIMPLE, 0x1000)
    if ida_ua.create_insn(entry_ea) <= 0:
        raise RuntimeError("cannot create entry instruction")
    ida_funcs.add_func(entry_ea, idaapi.BADADDR)
    ida_auto.auto_wait()
    function = ida_funcs.get_func(entry_ea)
    if function is None:
        raise RuntimeError("IDA did not create entry function")

    items = list(idautils.FuncItems(function.start_ea))
    instructions = []
    calls = []
    for ea in items:
        size = ida_bytes.get_item_size(ea)
        data = ida_bytes.get_bytes(ea, size) or b""
        text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        instructions.append({"ea": f"0x{ea:x}", "bytes": data.hex(), "disassembly": text})
        if idc.print_insn_mnem(ea).lower() in ("call", "callf"):
            targets = [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)]
            calls.append({"ea": f"0x{ea:x}", "targets": targets})

    summary = {
        "schema": "kol-ida-unpacked-entry-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "input_derivation": "CC0 unexepack source SHA-256 0e1da11a... applied to fixed KOLTOWN.EXE",
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "entry_ea": f"0x{entry_ea:x}",
        "function_start_ea": f"0x{function.start_ea:x}",
        "function_end_ea": f"0x{function.end_ea:x}",
        "instruction_count": len(items),
        "direct_call_site_count": len(calls),
        "database_function_count_after_probe": ida_funcs.get_func_qty(),
    }
    private = dict(summary)
    private["schema"] = "kol-ida-unpacked-entry-private-v1"
    private["instructions"] = instructions
    private["calls"] = calls

    private_output.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_output.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN unpacked entry probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
