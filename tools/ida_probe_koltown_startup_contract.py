"""IDA Pro 9.4：匯出 KOLTOWN startup／main 的非破壞性啟動契約證據。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_bytes
import ida_funcs
import ida_lines
import ida_nalt
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
TARGETS = [0x10000, 0x2B9E6, 0x2BAB1, 0x2BBD0, 0x2BD64]
ARG_GLOBALS = [0x2F32C, 0x2F32E, 0x2F330]


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:x}",
        "bytes": raw.hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [f"0x{x:x}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea: int) -> dict:
    function = ida_funcs.get_func(ea)
    if function is None:
        return {"requested_ea": f"0x{ea:x}", "missing": True}
    items = list(idautils.FuncItems(function.start_ea))
    return {
        "requested_ea": f"0x{ea:x}",
        "original_name": idc.get_func_name(function.start_ea),
        "start_ea": f"0x{function.start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(function.start_ea, False)],
        "instructions": [instruction(item) for item in items],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    summary_path = Path(idc.ARGV[2])
    input_path = Path(idc.ARGV[3])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")

    functions = [function_record(ea) for ea in TARGETS]
    globals_record = []
    for ea in ARG_GLOBALS:
        globals_record.append({
            "ea": f"0x{ea:x}",
            "original_name": idc.get_name(ea),
            "raw_word": ida_bytes.get_word(ea),
            "xrefs_to": [f"0x{x.frm:x}" for x in idautils.XrefsTo(ea, 0)],
        })

    summary = {
        "schema": "kol-ida-koltown-startup-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "targets": [
            {
                "requested_ea": record["requested_ea"],
                "original_name": record.get("original_name"),
                "start_ea": record.get("start_ea"),
                "end_ea": record.get("end_ea"),
                "instruction_count": len(record.get("instructions", [])),
                "caller_count": len(record.get("callers", [])),
            }
            for record in functions
        ],
        "argument_globals": globals_record,
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-startup-private-v1"
    private["functions"] = functions

    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN startup probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
