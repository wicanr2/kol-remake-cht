"""IDA Pro 9.4：固定 KOLBOOK 角色建立 UI 返回與 commit 邊界。"""

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


EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
EXPECTED = {
    0x11907: "b81000", 0x11915: "a05014", 0x1194D: "833e500000", 0x11957: "e9170d",
    0x119A4: "e81229", 0x119AD: "833e500000", 0x119ED: "e90706",
    0x11A87: "e82f28", 0x11ABE: "833e500000", 0x11AC8: "e98ffe",
    0x11BD4: "e8e226", 0x11C11: "833e500000", 0x11C24: "e92bfe", 0x11C2A: "e92dfd",
    0x11C9D: "8b46fa", 0x11DA9: "e80d25", 0x11DB2: "833e500000", 0x11DC5: "e9d5fe",
    0x11DF6: "833e500000", 0x11E00: "e908fd",
    0x11F6A: "817ef4004b", 0x11F74: "817ef40048", 0x11F85: "3dffff", 0x11F8D: "c746f24f00",
    0x11F92: "817ef4004d", 0x11F9C: "817ef40050", 0x11FAD: "3d5000", 0x11FB5: "c746f20000",
    0x11FBA: "833e500000", 0x11FC4: "837ef40d", 0x11FD7: "b8b400", 0x11FF1: "e995fc",
    0x12007: "833ebaa6ff", 0x12011: "c746f20000", 0x12034: "8146f20003",
    0x12039: "817ef20030", 0x12043: "833ebaa6ff", 0x1204D: "b83800",
    0x1205B: "e87026", 0x12061: "e90d06", 0x12064: "b90003",
}


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
    }


def function_record(start: int) -> dict:
    func = ida_funcs.get_func(start)
    if func is None:
        raise RuntimeError(f"missing function 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{func.start_ea:x}",
        "end_ea": f"0x{func.end_ea:x}",
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    input_hash = digest(input_path.read_bytes())
    if input_hash != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {input_hash}")
    ida_auto.auto_wait()
    checks = []
    for ea, expected in EXPECTED.items():
        row = instruction(ea)
        if row["bytes"] != expected:
            raise RuntimeError(f"instruction bytes mismatch at 0x{ea:x}: {row['bytes']} != {expected}")
        checks.append(row)

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": input_hash,
        "address_space": "IDA linear effective address for KOLBOOK loaded MZ",
        "instruction_checks": checks,
    }
    private = {
        "schema": "kol-ida-kolbook-character-ui-private-v1",
        **shared,
        "character_creation_function": function_record(0x11820),
        "name_input_function": function_record(0x14C02),
        "menu_function": function_record(0x142B9),
        "key_input_function": function_record(0x146CE),
    }
    safe = {
        "schema": "kol-ida-kolbook-character-ui-summary-v1",
        **shared,
        "contracts": [
            {"semantic": "name input capacity is 16 bytes; empty first byte loops; cancel exits creation", "status": "confirmed", "eas": ["0x11907", "0x11915", "0x1194d", "0x11957"]},
            {"semantic": "race cancel returns to name", "status": "confirmed", "eas": ["0x119a4", "0x119ad", "0x119ed", "0x11ff7"]},
            {"semantic": "secondary menu cancel returns to race", "status": "confirmed", "eas": ["0x11a87", "0x11abe", "0x11ac8"]},
            {"semantic": "profession cancel returns to secondary menu for selector below 2, otherwise race", "status": "confirmed", "eas": ["0x11bd4", "0x11c11", "0x11c24", "0x11c2a"]},
            {"semantic": "attribute menu result zero rerolls; cancel returns to profession", "status": "confirmed", "eas": ["0x11c9d", "0x11da9", "0x11db2", "0x11dc5", "0x11df6", "0x11e00"]},
            {"semantic": "portrait arrows wrap over indices 0 through 79; Return accepts; cancel returns before commit", "status": "confirmed", "eas": ["0x11f6a", "0x11f74", "0x11f85", "0x11f8d", "0x11f92", "0x11f9c", "0x11fad", "0x11fb5", "0x11fba", "0x11fc4", "0x11ff1"]},
            {"semantic": "slot scan tests 16 records at stride 0x300; full storage reports and exits before record clearing", "status": "confirmed", "eas": ["0x12007", "0x12011", "0x12034", "0x12039", "0x12043", "0x1204d", "0x1205b", "0x12061", "0x12064"]},
        ],
        "limitations": [
            "Static control flow does not replace a normal DOS UI oracle.",
            "Duplicate-name handling and save-failure rollback remain unknown.",
            "word_19290 is described as a cancel flag from consistent control-flow use; original address is retained.",
        ],
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK character UI probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
