"""IDA Pro 9.4：固定 KOLTOWN 的最小角色 reader 契約。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
EXPECTED_STRINGS = {
    0x2F600: "chardata",
    0x2F60A: "chardata",
    0x2F61E: "charpics",
    0x2F628: "charpics",
    0x2F632: "shldpics",
    0x2F63C: "shldpics",
}
EXPECTED_INSTRUCTIONS = {
    0x14DE7: "8b1e82bb", 0x14DEB: "83c324", 0x14DF1: "8a8788bb",
    0x14E05: "8b1e82bb", 0x14E09: "83c31c", 0x14E0F: "8a8788bb",
    0x14FAD: "8e063a6b", 0x14FB7: "b91000", 0x14FBA: "268a04",
    0x14FCB: "8b1e82bb", 0x14FD2: "8a8788bb",
    0x14CF1: "8b1e82bb", 0x14CF5: "83c32b", 0x14CFE: "8a8788bb",
    0x14D0A: "8b1e82bb", 0x14D0E: "83c32c", 0x14D14: "8a8788bb",
    0x19167: "8b1e82bb", 0x1916B: "83c330", 0x1916E: "8a8788bb",
}


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def insn(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [f"0x{x:x}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None:
        raise RuntimeError(f"missing function 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{function.start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in sorted(idautils.CodeRefsTo(function.start_ea, False))],
        "instructions": [insn(ea) for ea in idautils.FuncItems(function.start_ea)],
    }


def c_string(ea: int) -> str:
    data = bytearray()
    for offset in range(64):
        value = ida_bytes.get_byte(ea + offset)
        if value == 0:
            return data.decode("ascii")
        data.append(value)
    raise RuntimeError(f"unterminated string at 0x{ea:x}")


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = sha(input_path.read_bytes())
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    strings = []
    for ea, expected in EXPECTED_STRINGS.items():
        actual = c_string(ea)
        if actual != expected:
            raise RuntimeError(f"string mismatch at 0x{ea:x}: {actual!r}")
        strings.append({"ea": f"0x{ea:x}", "value": actual})
    checks = []
    for ea, expected in EXPECTED_INSTRUCTIONS.items():
        row = insn(ea)
        if row["bytes"] != expected:
            raise RuntimeError(f"instruction mismatch at 0x{ea:x}: {row['bytes']} != {expected}")
        checks.append({"ea": row["ea"], "bytes": row["bytes"], "disassembly": row["disassembly"]})

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for fixed unpacked KOLTOWN MZ derivative",
        "storage_route_strings": strings,
        "instruction_checks": checks,
    }
    private = {
        "schema": "kol-ida-koltown-character-reader-private-v1",
        **shared,
        "character_roster_function": function_record(0x147DD),
        "character_transfer_function": function_record(0x14DB1),
        "gold_read_function": function_record(0x14CCF),
        "weapon_render_function": function_record(0x18F3A),
    }
    safe = {
        "schema": "kol-ida-koltown-character-reader-summary-v1",
        **shared,
        "reader_contracts": [
            {"record_offset": "0x00..0x0f", "consumer": "sub_147DD", "sites": ["0x14fad..0x14fc1", "0x14fcb..0x14fdd"], "status": "confirmed 16-byte name roster and comparison"},
            {"record_offset": "0x24", "consumer": "sub_14DB1", "sites": ["0x14de7..0x14dfd"], "status": "confirmed race eligibility input"},
            {"record_offset": "0x1c", "consumer": "sub_14DB1", "sites": ["0x14e05..0x14e23"], "status": "confirmed profession eligibility input"},
            {"record_offset": "0x2b..0x2c", "consumer": "sub_14CCF", "sites": ["0x14cf1..0x14d1e"], "status": "confirmed little-endian gold reader"},
            {"record_offset": "0x30", "consumer": "sub_18F3A", "sites": ["0x19167..0x1917e"], "status": "confirmed primary weapon template presence and renderer route"},
        ],
        "limitation": "This is the minimal reader cross-check for creation fields, not a complete 768-byte schema. Weapon-skill values remain writer-plus-manual confirmed; this probe does not assign broader skill semantics.",
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN character reader probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
