"""IDA Pro 9.4：匯出 KOLTOWN PERPICS consumer 與 caller 證據。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_bytes
import ida_auto
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
PERPICS_CONSUMER = 0x18F3A
RELATED_FUNCTIONS = [0x1CBA9, 0x16FA6, 0x2D3FA]
DGROUP_EA = 0x2DED0
DATA_WINDOWS = [
    ("index_tables_0168_0680", 0x0168, 0x06A0),
    ("format_strings_17d8_1802", 0x17D8, 0x1802),
]


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


def data_window_record(label: str, start_offset: int, end_offset: int) -> dict:
    start_ea = DGROUP_EA + start_offset
    size = end_offset - start_offset
    raw = ida_bytes.get_bytes(start_ea, size)
    if raw is None or len(raw) != size:
        raise RuntimeError(f"missing DGROUP data window {label} at {start_ea:#x}")
    return {
        "label": label,
        "dgroup_start_offset": f"0x{start_offset:x}",
        "dgroup_end_offset_exclusive": f"0x{end_offset:x}",
        "ida_start_ea": f"0x{start_ea:x}",
        "bytes": raw.hex(),
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    safe_path = Path(idc.ARGV[2])
    input_path = Path(idc.ARGV[3])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")

    ida_auto.auto_wait()

    caller_eas = sorted({
        ida_funcs.get_func(ref).start_ea
        for ref in idautils.CodeRefsTo(PERPICS_CONSUMER, False)
        if ida_funcs.get_func(ref) is not None
    })
    function_eas = list(dict.fromkeys([PERPICS_CONSUMER, *caller_eas, *RELATED_FUNCTIONS]))
    functions = [function_record(ea) for ea in function_eas]
    output = {
        "schema": "kol-ida-koltown-perpics-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "consumer_original_name": idc.get_func_name(PERPICS_CONSUMER),
        "consumer_start_ea": f"0x{PERPICS_CONSUMER:x}",
        "caller_start_eas": [f"0x{ea:x}" for ea in caller_eas],
        "functions": functions,
        "data_windows": [data_window_record(*window) for window in DATA_WINDOWS],
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")

    dimension_bytes = ida_bytes.get_bytes(DGROUP_EA + 0x168, 18)
    if dimension_bytes is None or len(dimension_bytes) != 18:
        raise RuntimeError("missing PERPICS dimension table")
    dimensions = []
    for index in range(6):
        x0, y0, x1 = dimension_bytes[index * 3 : index * 3 + 3]
        if x1 <= x0 or y0 >= 0x9A:
            raise RuntimeError(f"invalid PERPICS dimension record {index}")
        dimensions.append({
            "portrait_index": index,
            "color_image_index": index,
            "mask_image_index": index + 6,
            "x0": x0,
            "y0": y0,
            "x1_exclusive": x1,
            "y1_exclusive": 0x9A,
            "width_bytes": x1 - x0,
            "width_pixels": (x1 - x0) * 8,
            "height_pixels": 0x9A - y0,
        })
    safe = {
        "schema": "kol-ida-koltown-perpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "evidence_contract": {
            "filename_format_dgroup_offset": "0x17d8",
            "filename_format": "per%s",
            "filename_format_call_ea": "0x18fdd",
            "consumer_original_name": "sub_18F3A",
            "consumer_start_ea": "0x18f3a",
            "rectangle_pair_call_ea": "0x18ff3",
            "rectangle_pair_original_name": "sub_16FA6",
            "shared_loader_original_name": "sub_1A4FE",
            "color_index_expression": "portrait_index",
            "mask_index_expression": "portrait_index + 6",
            "dimensions_dgroup_offset": "0x168",
            "fixed_y1_exclusive": 154,
            "status": "confirmed",
        },
        "dimensions": dimensions,
        "notes": [
            "原始函式名、線性位址與呼叫點均保留；完整指令與 bytes 僅存 private JSON。",
            "sub_1CBA9 是媒體／磁碟提示流程，不是 PERPICS 圖片 loader。",
            "後六個 payload 的 1-bit 遮罩排列與尾 byte 語意仍須由解碼探針驗證。",
        ],
    }
    safe_path.parent.mkdir(parents=True, exist_ok=True)
    safe_path.write_text(json.dumps(safe, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN PERPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
