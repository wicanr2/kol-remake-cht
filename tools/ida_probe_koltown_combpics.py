"""IDA Pro 9.4：匯出 KOLTOWN COMBPICS 七張 EGA 圖 consumer 證據。"""

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
DGROUP_EA = 0x2DED0
FUNCTION_EAS = [
    0x1E3A8, 0x288D4, 0x2AAA8, 0x2B454, 0x1A4FE, 0x10F9C,
    0x1C907, 0x1A2BB, 0x19E9D, 0x2D3FA,
]
LARGE_CALLS = [(0x2125C, 3), (0x2157B, 2), (0x215AC, 0), (0x2161F, 4), (0x216B5, 1)]
SMALL_CALLS = [(0x2ABD2, 6), (0x2ADCB, 5)]


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


def require_instruction(ea: int, mnemonic: str, fragment: str) -> None:
    if idc.print_insn_mnem(ea).lower() != mnemonic.lower():
        raise RuntimeError(f"unexpected mnemonic at {ea:#x}")
    disassembly = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    if fragment not in disassembly:
        raise RuntimeError(f"unexpected operand at {ea:#x}: {disassembly}")


def require_call_map(function_ea: int, expected: list[tuple[int, int]]) -> list[dict]:
    refs = sorted(idautils.CodeRefsTo(function_ea, False))
    expected_refs = [ea for ea, _ in expected]
    if refs != expected_refs:
        raise RuntimeError(f"unexpected callers for {function_ea:#x}: {[hex(x) for x in refs]}")
    records = []
    for call_ea, image_index in expected:
        require_instruction(call_ea, "call", idc.get_func_name(function_ea))
        previous = idc.prev_head(call_ea)
        require_instruction(previous, "push", "ax")
        load = idc.prev_head(previous)
        require_instruction(load, "mov", f"ax, {image_index:X}h" if image_index > 9 else f"ax, {image_index}")
        records.append({"call_ea": f"0x{call_ea:x}", "image_index": image_index})
    return records


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

    large_calls = require_call_map(0x288D4, LARGE_CALLS)
    small_calls = require_call_map(0x2B454, SMALL_CALLS)
    checks = [
        (0x288EC, "mov", "word_3499E, 1"), (0x288F2, "mov", "word_34A06, 26h"),
        (0x288F8, "mov", "word_349A0, 8"), (0x288FE, "mov", "word_34A0E, 98h"),
        (0x28937, "call", "sub_1A4FE"), (0x28984, "call", "sub_1C907"),
        (0x289AC, "call", "sub_19E9D"),
        (0x2B45D, "mov", "word_3499E, 1"), (0x2B463, "mov", "word_349A0, 40h"),
        (0x2B469, "mov", "word_34A06, 0Ch"), (0x2B46F, "mov", "word_34A0E, 80h"),
        (0x2B49A - 1, "call", "sub_10F9C"),
    ]
    for check in checks:
        require_instruction(*check)
    for offset in (0x1856, 0x18A8):
        if ida_bytes.get_strlit_contents(DGROUP_EA + offset, -1, 0) != b"comb%s":
            raise RuntimeError(f"unexpected COMBPICS format at DGROUP+{offset:#x}")

    private = {
        "schema": "kol-ida-koltown-combpics-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "functions": [function_record(ea) for ea in FUNCTION_EAS],
    }
    safe = {
        "schema": "kol-ida-koltown-combpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "archive": {"filename_format": "comb%s", "payload_count": 7, "status": "confirmed"},
        "consumer_groups": [
            {
                "original_name": "sub_288D4",
                "start_ea": "0x288d4",
                "filename_format_dgroup_offset": "0x1856",
                "calls": large_calls,
                "image_indices": [0, 1, 2, 3, 4],
                "rectangle": {
                    "x0_byte_column": 1, "y0_scanline": 8,
                    "x1_byte_column_exclusive": 38, "y1_scanline_exclusive": 152,
                    "width_bytes": 37, "width_pixels": 296, "height_pixels": 144,
                },
                "loader_original_name": "sub_1A4FE", "loader_call_ea": "0x28937",
                "second_stage_original_name": "sub_1C907", "second_stage_call_ea": "0x28984",
                "blitter_original_name": "sub_19E9D", "blitter_call_ea": "0x289ac",
                "status": "confirmed",
            },
            {
                "original_name": "sub_2B454",
                "start_ea": "0x2b454",
                "filename_format_dgroup_offset": "0x18a8",
                "calls": small_calls,
                "image_indices": [5, 6],
                "rectangle": {
                    "x0_byte_column": 1, "y0_scanline": 64,
                    "x1_byte_column_exclusive": 12, "y1_scanline_exclusive": 128,
                    "width_bytes": 11, "width_pixels": 88, "height_pixels": 64,
                },
                "loader_and_blit_original_name": "sub_10F9C",
                "loader_and_blit_call_ea": "0x2b499",
                "status": "confirmed",
            },
        ],
        "coverage": {
            "consumer_image_indices": [0, 1, 2, 3, 4, 5, 6],
            "archive_image_indices": [0, 1, 2, 3, 4, 5, 6],
            "all_archive_indices_have_consumer": True,
            "status": "confirmed",
        },
        "notes": [
            "完整函式、原始位址、運算元與 bytes 只保存於 private JSON。",
            "二階 RLE、4-bit EGA pixels、palette bounds 與方向仍須由解碼及視覺 probe 驗證。",
        ],
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.parent.mkdir(parents=True, exist_ok=True)
    safe_path.write_text(json.dumps(safe, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN COMBPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
