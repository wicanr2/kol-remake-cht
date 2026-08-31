"""IDA Pro 9.4：固定 KOLBOOK 對 chardata 的檔案路由與 caller。"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
DGROUP_EA = 0x19240
CHARDATA_STRING_EA = 0x1945A


def insn(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
    }


def function_record(start: int) -> dict:
    f = ida_funcs.get_func(start)
    if f is None:
        raise RuntimeError(f"missing function 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{f.end_ea:x}",
        "callers": [f"0x{x:x}" for x in sorted(idautils.CodeRefsTo(start, False))],
        "instructions": [insn(ea) for ea in idautils.FuncItems(start)],
    }


def caller_record(call_ea: int) -> dict:
    f = ida_funcs.get_func(call_ea)
    if f is None:
        raise RuntimeError(f"caller outside function 0x{call_ea:x}")
    items = list(idautils.FuncItems(f.start_ea))
    index = items.index(call_ea)
    context_eas = items[max(0, index - 8):index + 3]
    immediate_arg0 = None
    previous = items[index - 1] if index >= 1 else None
    previous2 = items[index - 2] if index >= 2 else None
    if previous is not None and idc.print_insn_mnem(previous).lower() == "push":
        operand = idc.print_operand(previous, 0).lower().replace(" ", "")
        if operand not in {"ax", "eax"}:
            immediate_arg0 = idc.get_operand_value(previous, 0)
        elif previous2 is not None and idc.print_insn_mnem(previous2).lower() == "mov":
            if idc.print_operand(previous2, 0).lower().replace(" ", "") in {"ax", "eax"}:
                immediate_arg0 = idc.get_operand_value(previous2, 1)
    return {
        "call_ea": f"0x{call_ea:x}",
        "caller_original_function": idc.get_func_name(f.start_ea),
        "caller_function_start_ea": f"0x{f.start_ea:x}",
        "immediate_arg0": immediate_arg0,
        "context": [insn(ea) for ea in context_eas],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    string_ea = CHARDATA_STRING_EA
    raw_string = ida_bytes.get_bytes(string_ea, 9)
    if raw_string != b"chardata\x00":
        private_path.write_text(json.dumps({"ea": hex(string_ea), "raw": raw_string.hex() if raw_string else None}) + "\n")
        raise RuntimeError(f"chardata bytes mismatch: {raw_string!r}")
    text = "chardata"
    relative = string_ea - DGROUP_EA

    raw_input = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw_input, 0x08)[0] * 16
    load_image = raw_input[header_size:]
    occurrences = []
    starts = set()
    needle = struct.pack("<H", relative)
    cursor = 0
    while True:
        found = load_image.find(needle, cursor)
        if found < 0:
            break
        if found == 0 or load_image[found - 1] != 0xB9:
            raise RuntimeError(f"unexpected chardata pointer encoding at load offset 0x{found:x}")
        ea = 0x10000 + found - 1
        is_code = ida_bytes.is_code(ida_bytes.get_flags(ea))
        if is_code:
            f = ida_funcs.get_func(ea)
            if f is not None:
                starts.add(f.start_ea)
            occurrences.append({
                **insn(ea),
                "original_function": idc.get_func_name(f.start_ea) if f else "",
                "function_start_ea": f"0x{f.start_ea:x}" if f else None,
            })
        else:
            occurrences.append({
                "ea": f"0x{ea:x}",
                "bytes": (ida_bytes.get_bytes(ea, 16) or b"").hex(),
                "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
                "original_function": "",
                "function_start_ea": None,
                "is_code": False,
            })
        cursor = found + 1
    if not occurrences:
        private_path.write_text(json.dumps({"string_ea": hex(string_ea), "relative": relative, "occurrences": []}) + "\n")
        raise RuntimeError("no DGROUP-relative chardata occurrence")

    route_start = next(iter(starts))
    callers = [caller_record(ea) for ea in sorted(idautils.CodeRefsTo(route_start, False))]

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for KOLBOOK loaded MZ",
        "dgroup_ea": f"0x{DGROUP_EA:x}",
        "string_ea": f"0x{string_ea:x}",
        "string_text": text,
        "dgroup_relative_offset": relative,
        "occurrences": occurrences,
    }
    private = {
        "schema": "kol-ida-kolbook-chardata-private-v1",
        **shared,
        "caller_sites": callers,
        "functions": [function_record(x) for x in sorted(starts)],
    }
    safe = {
        "schema": "kol-ida-kolbook-chardata-summary-v1",
        **shared,
        "consumer_original_functions": [idc.get_func_name(x) for x in sorted(starts)],
        "route_call_site_count": len(callers),
        "chardata_call_sites": [
            {
                "call_ea": row["call_ea"],
                "caller_original_function": row["caller_original_function"],
                "caller_function_start_ea": row["caller_function_start_ea"],
            }
            for row in callers if row["immediate_arg0"] == 0
        ],
        "limitation": "檔案路由已固定；768-byte record 內欄位仍需 writer／reader 資料流證據。",
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK chardata probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
