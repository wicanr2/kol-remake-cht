"""IDA Pro 9.4：匯出 KOLBOOK 760-byte 角色圖片記錄的直接繪圖 consumer。"""

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
ROOT_FUNCTION_EAS = (0x15335, 0x153EA)
MAX_CALL_DEPTH = 3


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


def discover_functions() -> dict[int, int]:
    depths = {ea: 0 for ea in ROOT_FUNCTION_EAS}
    pending = list(ROOT_FUNCTION_EAS)
    while pending:
        start = pending.pop(0)
        depth = depths[start]
        function = ida_funcs.get_func(start)
        if not function or depth >= MAX_CALL_DEPTH:
            continue
        for ea in idautils.FuncItems(function.start_ea):
            for target in idautils.CodeRefsFrom(ea, False):
                callee = ida_funcs.get_func(target)
                if not callee or callee.start_ea == function.start_ea:
                    continue
                if callee.start_ea not in depths:
                    depths[callee.start_ea] = depth + 1
                    pending.append(callee.start_ea)
    return depths


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    summary_path = Path(idc.ARGV[2])
    input_path = Path(idc.ARGV[3])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    depths = discover_functions()
    functions = []
    for start, depth in sorted(depths.items()):
        function = ida_funcs.get_func(start)
        if not function:
            raise RuntimeError(f"missing function at 0x{start:x}")
        functions.append(
            {
                "original_name": idc.get_func_name(start),
                "start_ea": f"0x{start:x}",
                "end_ea": f"0x{function.end_ea:x}",
                "discovery_depth": depth,
                "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
                "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
            }
        )

    by_name = {item["original_name"]: item for item in functions}
    direct = [item["disassembly"] for item in by_name["sub_15AEC"]["instructions"]]
    enlarged = [item["disassembly"] for item in by_name["sub_15B6E"]["instructions"]]
    required_direct = {
        "mov     ch, 26h ; '&'",
        "mov     cl, 14h",
        "mov     bl, [si]",
        "shr     al, 1",
        "inc     si",
    }
    required_enlarged = required_direct | {
        "mov     ah, es:[di]",
        "mov     ah, es:[di+28h]",
        "mov     ah, es:[di+50h]",
        "mov     ah, es:[di+78h]",
        "add     di, 0A0h",
    }
    if not required_direct.issubset(direct):
        raise RuntimeError("sub_15AEC no longer matches the 40x38 packed-nibble loop")
    if not required_enlarged.issubset(enlarged):
        raise RuntimeError("sub_15B6E no longer matches the 4x enlarged packed-nibble loop")

    common = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for loaded KOLBOOK MZ",
        "root_original_functions": [idc.get_func_name(ea) for ea in ROOT_FUNCTION_EAS],
        "max_call_depth": MAX_CALL_DEPTH,
    }
    private = dict(common)
    private["schema"] = "kol-ida-kolbook-charpic-record-private-v1"
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")

    summary = dict(common)
    summary["schema"] = "kol-ida-kolbook-charpic-record-summary-v1"
    summary["record_contract"] = {
        "status": "confirmed",
        "record_bytes": 760,
        "width_pixels": 40,
        "height_pixels": 38,
        "bits_per_pixel": 4,
        "pixels_per_byte": 2,
        "packing": "high nibble first, then low nibble",
        "direct_consumer": {
            "original_function": "sub_15AEC",
            "ea_range": "0x15aec..0x15b6e",
            "row_count_instruction_ea": "0x15b1d",
            "bytes_per_row_instruction_ea": "0x15b20",
        },
        "four_times_enlarged_consumer": {
            "original_function": "sub_15B6E",
            "ea_range": "0x15b6e..0x15c14",
            "row_count_instruction_ea": "0x15b9f",
            "bytes_per_row_instruction_ea": "0x15ba2",
        },
    }
    summary["functions"] = [
        {
            "original_name": item["original_name"],
            "start_ea": item["start_ea"],
            "end_ea": item["end_ea"],
            "discovery_depth": item["discovery_depth"],
            "instruction_count": len(item["instructions"]),
            "caller_count": len(item["callers"]),
        }
        for item in functions
    ]
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK charpic record probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
