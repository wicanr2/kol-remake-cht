"""IDA Pro 9.4：匯出 KOLTOWN MISCPICS 路由、尺寸與 consumer 證據。"""

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
FUNCTION_EAS = [0x10000, 0x10DDC, 0x1103F, 0x11262, 0x1A429, 0x19D98, 0x19E9D, 0x2D3FA]
DGROUP_EA = 0x2DED0
MISCPICS_BUFFER_EA = 0x2FBB8
ROUTES = [
    {"request_code": 34, "image_index": 0, "x0": 13, "y0": 58, "x1": 37, "y1": 154},
    {"request_code": 35, "image_index": 1, "x0": 1, "y0": 164, "x1": 38, "y1": 196},
    {"request_code": 71, "image_index": 3, "x0": 9, "y0": 85, "x1": 16, "y1": 155},
    {"request_code": 72, "image_index": 4, "x0": 9, "y0": 85, "x1": 16, "y1": 155},
    {"request_code": 73, "image_index": 5, "x0": 9, "y0": 85, "x1": 16, "y1": 155},
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


def require_instruction(ea: int, mnemonic: str, fragment: str) -> None:
    if idc.print_insn_mnem(ea).lower() != mnemonic.lower():
        raise RuntimeError(f"unexpected mnemonic at {ea:#x}")
    disassembly = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    if fragment not in disassembly:
        raise RuntimeError(f"unexpected operand at {ea:#x}: {disassembly}")


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

    checks = [
        (0x1060D, "call", "sub_2D3FA"), (0x10621, "call", "sub_19D98"),
        (0x1108D, "mov", "word_3499E, 1"), (0x11093, "mov", "word_349A0, 0A4h"),
        (0x11099, "mov", "word_34A06, 26h"), (0x1109F, "mov", "word_34A0E, 0C4h"),
        (0x11111, "mov", "word_3499E, 0Dh"), (0x11117, "mov", "word_349A0, 3Ah"),
        (0x1111D, "mov", "word_34A06, 25h"), (0x11123, "mov", "word_34A0E, 9Ah"),
        (0x1113B, "mov", "word_3499E, 9"), (0x11141, "mov", "word_349A0, 55h"),
        (0x11147, "mov", "word_34A06, 10h"), (0x1114D, "mov", "word_34A0E, 9Bh"),
        (0x111A3, "sub", "[bp+arg_0], 1Ch"), (0x111AA, "sub", "[bp+arg_0], 3Eh"),
        (0x111C0, "push", "word_2FBB8"), (0x111C7, "add", "0FFFAh"),
        (0x111CB, "call", "sub_11262"), (0x11288, "call", "sub_1A429"),
        (0x112F4, "call", "sub_19E9D"),
    ]
    for check in checks:
        require_instruction(*check)
    archive_format = ida_bytes.get_strlit_contents(DGROUP_EA + 0x1678, -1, 0)
    if archive_format != b"misc%s.0":
        raise RuntimeError(f"unexpected MISCPICS filename format: {archive_format!r}")
    buffer_data_refs = sorted(idautils.DataRefsTo(MISCPICS_BUFFER_EA))
    expected_buffer_refs = [0x105EC, 0x10E62, 0x111C0]
    if buffer_data_refs != expected_buffer_refs:
        raise RuntimeError(
            f"unexpected MISCPICS buffer data refs: {[hex(x) for x in buffer_data_refs]}"
        )

    route_records = []
    for route in ROUTES:
        width_bytes = route["x1"] - route["x0"]
        height = route["y1"] - route["y0"]
        route_records.append({
            "request_code": route["request_code"],
            "image_index": route["image_index"],
            "x0": route["x0"],
            "y0": route["y0"],
            "x1_exclusive": route["x1"],
            "y1_exclusive": route["y1"],
            "width_bytes": width_bytes,
            "width_pixels": width_bytes * 8,
            "height_pixels": height,
            "status": "confirmed",
        })
    private = {
        "schema": "kol-ida-koltown-miscpics-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "miscpics_buffer_data_refs": [f"0x{x:x}" for x in buffer_data_refs],
        "functions": [function_record(ea) for ea in FUNCTION_EAS],
    }
    safe = {
        "schema": "kol-ida-koltown-miscpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "archive_load": {
            "filename_format": "misc%s.0",
            "filename_format_dgroup_offset": "0x1678",
            "format_call_ea": "0x1060d",
            "whole_file_reader_original_name": "sub_19D98",
            "whole_file_reader_call_ea": "0x10621",
            "archive_buffer_original_name": "word_2FBB8",
            "archive_buffer_ea": "0x2fbb8",
            "archive_buffer_data_refs": [f"0x{x:x}" for x in buffer_data_refs],
            "archive_buffer_ref_roles": {
                "0x105ec": "startup allocation/load destination selection",
                "0x10e62": "cleanup/free",
                "0x111c0": "only image consumer",
            },
            "status": "confirmed",
        },
        "router": {
            "original_name": "sub_1103F",
            "start_ea": "0x1103f",
            "misc_buffer_push_ea": "0x111c0",
            "image_index_expression": "mapped_request_code - 6",
            "index_adjust_instruction_ea": "0x111c7",
            "consumer_call_ea": "0x111cb",
            "routes": route_records,
            "unreferenced_placeholder_image_index": 2,
            "placeholder_reason": "the proven request-code mapping set is {6,7,9,10,11}; subtracting 6 never yields 2",
            "status": "confirmed",
        },
        "consumer": {
            "original_name": "sub_11262",
            "start_ea": "0x11262",
            "indexed_decoder_original_name": "sub_1A429",
            "indexed_decoder_call_ea": "0x11288",
            "direct_blitter_original_name": "sub_19E9D",
            "direct_blitter_call_ea": "0x112f4",
            "status": "confirmed",
        },
        "notes": [
            "完整函式、原始位址、運算元與 bytes 只保存在 private JSON。",
            "圖號 2 是一-byte placeholder；未被此 router 選取，不得輸出 PNG。",
            "二階 RLE、4-bit EGA 像素與方向仍須由解碼及視覺 probe 驗證。",
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
    print(f"KOLTOWN MISCPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
