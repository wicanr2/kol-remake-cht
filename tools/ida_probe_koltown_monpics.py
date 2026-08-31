"""IDA Pro 9.4：匯出 KOLTOWN MONPICS 尺寸表與消費端證據。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
DGROUP_EA = 0x2DED0
MON_DIMENSIONS_OFFSET = 0x1370
MONSTER_COUNT = 32
MONPICS_IMAGE_COUNT = 64
CONSUMERS = (0x22EFC, 0x2A975)
GENERIC_IMAGE_LOADER = 0x10F9C


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


def function_record(start_ea: int) -> dict:
    function = ida_funcs.get_func(start_ea)
    if function is None or function.start_ea != start_ea:
        raise RuntimeError(f"missing function at {start_ea:#x}")
    return {
        "original_name": idc.get_func_name(start_ea),
        "start_ea": f"0x{start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start_ea, False)],
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start_ea)],
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

    table_ea = DGROUP_EA + MON_DIMENSIONS_OFFSET
    table = ida_bytes.get_bytes(table_ea, MONSTER_COUNT * 2)
    if table is None or len(table) != MONSTER_COUNT * 2:
        raise RuntimeError("MONPICS dimension table unavailable")
    dimensions = [
        {
            "monster_index": index,
            "width_bytes": table[index * 2],
            "width_pixels": table[index * 2] * 8,
            "height_pixels": table[index * 2 + 1],
        }
        for index in range(MONSTER_COUNT)
    ]
    if any(item["width_bytes"] == 0 or item["height_pixels"] == 0 for item in dimensions):
        raise RuntimeError("MONPICS dimension table contains zero-sized entry")

    functions = [function_record(ea) for ea in (*CONSUMERS, GENERIC_IMAGE_LOADER)]
    summary = {
        "schema": "kol-ida-koltown-monpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "dimension_table": {
            "dgroup_relative_offset": MON_DIMENSIONS_OFFSET,
            "ida_linear_ea": f"0x{table_ea:x}",
            "entry_count": MONSTER_COUNT,
            "entry_contract": "two unsigned bytes: width in 8-pixel bytes, height in pixels",
            "dimensions": dimensions,
        },
        "archive_contract": {
            "image_count": MONPICS_IMAGE_COUNT,
            "images_per_monster": 2,
            "image_index_contract": "2 * monster_index and 2 * monster_index + 1",
        },
        "consumer_contracts": [
            {
                "original_function": "sub_22EFC",
                "start_ea": "0x22efc",
                "archive_name_offset": 0x1850,
                "archive_pattern": "mon%s",
                "image_index_expression": "2 * word_3B262",
                "dimension_index_expression": "2 * word_3B262",
                "generic_loader": "sub_10F9C",
                "generic_loader_call_ea": "0x22fb9",
                "evidence_grade": "confirmed",
            },
            {
                "original_function": "sub_2A975",
                "start_ea": "0x2a975",
                "dimension_index_expression": "2 * word_3B262",
                "rectangle_copy_function": "sub_16FA6",
                "rectangle_copy_call_ea": "0x2aa0f",
                "evidence_grade": "confirmed",
            },
        ],
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-monpics-private-v1"
    private["dimension_table_bytes"] = table.hex()
    private["functions"] = functions
    private_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN MONPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
