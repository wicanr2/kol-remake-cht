"""IDA Pro 9.4：窮舉 KOLBOOK 的 rand／srand 路徑與 caller。"""

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
DICE_SIDES_EA = 0x1929C
EXPECTED_DICE_SIDES = bytes([3, 4, 6, 8, 10, 12, 20, 100])
EXPECTED_INSTRUCTION_BYTES = {
    0x1001E: "cd1a",
    0x10020: "8916e679",
    0x1011F: "c706e6790000",
    0x10157: "ff36e679",
    0x1015B: "e8c487",
    0x14BED: "e8443d",
    0x14BF1: "f77e04",
    0x14BF4: "8bc2",
    0x14BF6: "050100",
    0x11CBC: "8a877277",
    0x11CCC: "8a877277",
    0x11CB6: "83c308",
    0x11CE1: "8a875c00",
    0x11CEB: "8366ec0f",
    0x11CFA: "e8e72e",
    0x11D00: "0146f6",
    0x11D18: "88873412",
    0x11D44: "a23b12",
    0x11D62: "a23c12",
    0x11D80: "a23d12",
    0x11BDA: "0346f8",
    0x11BDD: "05ebff",
    0x11C85: "8346fa06",
    0x126B3: "8a873412",
    0x126BA: "b96400",
    0x126C4: "f7f1",
}


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


def operand_sites(*needles: str) -> list[dict]:
    rows = []
    lowered = tuple(needle.lower() for needle in needles)
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            row = insn(ea)
            if any(needle in row["disassembly"].lower() for needle in lowered):
                rows.append({
                    "original_function": idc.get_func_name(start),
                    "function_start_ea": f"0x{start:x}",
                    **row,
                })
    return rows


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    checks = []
    for ea, expected in EXPECTED_INSTRUCTION_BYTES.items():
        row = insn(ea)
        if row["bytes"] != expected:
            raise RuntimeError(f"instruction bytes mismatch at 0x{ea:x}: {row['bytes']} != {expected}")
        checks.append({"ea": row["ea"], "bytes": row["bytes"], "disassembly": row["disassembly"]})
    dice_sides = ida_bytes.get_bytes(DICE_SIDES_EA, len(EXPECTED_DICE_SIDES))
    if dice_sides != EXPECTED_DICE_SIDES:
        raise RuntimeError(f"dice sides mismatch: {dice_sides!r}")

    rng_functions = []
    for ea in idautils.Functions():
        name = idc.get_func_name(ea)
        if "rand" in name.lower():
            rng_functions.append(ea)
    if not rng_functions:
        raise RuntimeError("no rand-named functions")

    caller_starts = set()
    rand_wrapper_starts = set()
    wrapper_routes = []
    routes = []
    for target in sorted(rng_functions):
        refs = sorted(idautils.CodeRefsTo(target, False))
        sites = []
        for ref in refs:
            f = ida_funcs.get_func(ref)
            if f is None:
                raise RuntimeError(f"RNG caller outside function 0x{ref:x}")
            caller_starts.add(f.start_ea)
            if idc.get_func_name(target) == "_rand":
                rand_wrapper_starts.add(f.start_ea)
            sites.append({
                "call_ea": f"0x{ref:x}",
                "caller_original_function": idc.get_func_name(f.start_ea),
                "caller_function_start_ea": f"0x{f.start_ea:x}",
            })
        routes.append({
            "target_original_name": idc.get_func_name(target),
            "target_start_ea": f"0x{target:x}",
            "call_sites": sites,
        })

    second_level_starts = set()
    for wrapper_start in sorted(rand_wrapper_starts):
        sites = []
        for ref in sorted(idautils.CodeRefsTo(wrapper_start, False)):
            f = ida_funcs.get_func(ref)
            if f is None:
                raise RuntimeError(f"wrapper caller outside function 0x{ref:x}")
            second_level_starts.add(f.start_ea)
            sites.append({
                "call_ea": f"0x{ref:x}",
                "caller_original_function": idc.get_func_name(f.start_ea),
                "caller_function_start_ea": f"0x{f.start_ea:x}",
            })
        wrapper_routes.append({
            "wrapper_original_name": idc.get_func_name(wrapper_start),
            "wrapper_start_ea": f"0x{wrapper_start:x}",
            "call_sites": sites,
        })

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for KOLBOOK loaded MZ",
        "rng_routes": routes,
        "wrapper_routes": wrapper_routes,
        "instruction_checks": checks,
    }
    private = {
        "schema": "kol-ida-kolbook-rng-private-v1",
        **shared,
        "rng_functions": [function_record(x) for x in sorted(rng_functions)],
        "caller_functions": [function_record(x) for x in sorted(caller_starts)],
        "wrapper_caller_functions": [function_record(x) for x in sorted(second_level_starts)],
        "bookdata_loader_function": function_record(0x13E3F),
        "bookdata_postload_function": function_record(0x13FFC),
        "bookdata_buffer_operand_sites": operand_sites("7772h", "209b2"),
        "character_stat_display_function": function_record(0x12677),
        "text_resource_function": function_record(0x14101),
    }
    safe = {
        "schema": "kol-ida-kolbook-rng-summary-v1",
        **shared,
        "caller_function_count": len(caller_starts),
        "wrapper_caller_function_count": len(second_level_starts),
        "confirmed_contract": {
            "seed": {
                "normal_path": "BIOS INT 1Ah low tick word DX passed to srand",
                "mode_7_path": "global is reset and incremented to 16 before srand",
                "status": "confirmed",
            },
            "bounded_roll": "sub_14BE4(N) = rand() % N + 1",
            "dice_sides_table_ea": f"0x{DICE_SIDES_EA:x}",
            "dice_sides": list(dice_sides),
            "primary_internal_stat_contract": (
                "for each of 7 fields: record byte[8+i] + low_nibble of byte[i] rolls of "
                "sub_14BE4(dice_sides[high_nibble])"
            ),
            "primary_output_eas": [f"0x{x:x}" for x in range(0x1A474, 0x1A47B)],
            "derived_output_contracts": [
                {"ea": "0x1a47b", "formula": "floor((EA 0x1a474 + EA 0x1a47a) / 2)"},
                {"ea": "0x1a47c", "formula": "floor((EA 0x1a474 + EA 0x1a477) / 2)"},
                {"ea": "0x1a47d", "formula": "floor((EA 0x1a476 + EA 0x1a477) / 2)"},
            ],
            "display_contract": "floor(internal_stat * 100 / 150)",
            "profession_record_groups": [
                {"selector": "race=0 HUMAN, secondary=0 MALE", "indices": [0, 11]},
                {"selector": "race=0 HUMAN, secondary=1 FEMALE", "indices": [12, 15]},
                {"selector": "race=1 ELVEN, secondary=0 MALE", "indices": [16, 21]},
                {"selector": "race=1 ELVEN, secondary=1 FEMALE", "indices": [22, 27]},
                {"selector": "race=2 DWARVEN", "indices": [28, 35]},
                {"selector": "race=3 KELDEN", "indices": [36, 38]},
            ],
            "status": "confirmed",
        },
        "limitation": "本摘要固定亂數、職業 selector、內部屬性與顯示換算；初始技能／物品與存檔 writer 另作審查。",
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK RNG probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
