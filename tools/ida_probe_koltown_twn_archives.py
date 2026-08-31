"""IDA Pro 9.4：匯出 KOLTOWN 的 TWN0..7PICS 全包幾何與圖號 0 consumer。"""

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
ROOT_FUNCTION_EAS = (0x11302, 0x14047, 0x15920)


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


def require_instruction(ea: int, expected: str) -> None:
    actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    if actual != expected:
        raise RuntimeError(f"0x{ea:x}: expected {expected!r}, got {actual!r}")


def caller_context(call_ea: int, before: int = 12, after: int = 3) -> dict:
    function = ida_funcs.get_func(call_ea)
    if not function:
        raise RuntimeError(f"caller function missing at 0x{call_ea:x}")
    items = list(idautils.FuncItems(function.start_ea))
    index = items.index(call_ea)
    selected = items[max(0, index - before) : min(len(items), index + after + 1)]
    return {
        "call_ea": f"0x{call_ea:x}",
        "caller_original_name": idc.get_func_name(function.start_ea),
        "caller_start_ea": f"0x{function.start_ea:x}",
        "instructions": [instruction(ea) for ea in selected],
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
    ida_auto.auto_wait()

    required = {
        0x1132C: "mov     word_3499E, 1",
        0x11332: "mov     word_349A0, 40h ; '@'",
        0x11338: "mov     word_34A06, 0Ch",
        0x1133E: "mov     word_34A0E, 80h",
        0x11344: "mov     al, [bp+arg_0]",
        0x1134A: "and     ax, 1",
        0x11355: "cmp     word_2DF2C, 0",
        0x1135F: "mov     word_3499E, 0Dh",
        0x11365: "mov     word_349A0, 3Ah ; ':'",
        0x1136B: "mov     word_34A06, 25h ; '%'",
        0x11371: "mov     word_34A0E, 9Ah",
        0x1137B: "mov     al, [bp+arg_0]",
        0x11382: "call    sub_10F9C",
        0x1409F: "mov     ax, 0",
        0x140A7: "call    sub_1A4FE",
        0x140AF: "mov     word_3499E, 0",
        0x140B5: "mov     word_349A0, 0",
        0x140BB: "mov     word_34A06, 7",
        0x140C1: "mov     word_34A0E, 39h ; '9'",
        0x14158: "mov     ax, 1Ch",
        0x1415C: "call    sub_1A429",
        0x1417B: "call    sub_1BD17",
        0x159AB: "mov     ax, 1780h",
        0x159EC: "mov     ax, 178Ah",
    }
    for ea, expected in required.items():
        require_instruction(ea, expected)

    functions = []
    call_sites = []
    for start in ROOT_FUNCTION_EAS:
        function = ida_funcs.get_func(start)
        if not function:
            raise RuntimeError(f"missing root function at 0x{start:x}")
        callers = sorted(idautils.CodeRefsTo(start, False))
        functions.append(
            {
                "original_name": idc.get_func_name(start),
                "start_ea": f"0x{start:x}",
                "end_ea": f"0x{function.end_ea:x}",
                "callers": [f"0x{x:x}" for x in callers],
                "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
            }
        )
        call_sites.extend(caller_context(ea) for ea in callers)

    common = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed KOLTOWN MZ",
    }
    private = dict(common)
    private["schema"] = "kol-ida-koltown-twn-archives-private-v1"
    private["functions"] = functions
    private["call_sites"] = call_sites
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")

    summary = dict(common)
    summary["schema"] = "kol-ida-koltown-twn-archives-summary-v1"
    summary["geometry_contract"] = {
        "status": "confirmed",
        "general_consumer_original_function": "sub_11302",
        "general_consumer_ea_range": "0x11302..0x11390",
        "town_zero_rectangle": {
            "condition": "word_2DF2C == 0",
            "width_pixels": 192,
            "height_pixels": 96,
        },
        "nonzero_town_even_rectangle": {
            "condition": "word_2DF2C != 0 and image_index is even",
            "width_pixels": 88,
            "height_pixels": 64,
        },
        "nonzero_town_odd_rectangle": {
            "condition": "word_2DF2C != 0 and image_index is odd",
            "width_pixels": 192,
            "height_pixels": 96,
        },
        "town_nonzero_image_zero": {
            "consumer_original_function": "sub_14047",
            "consumer_ea_range": "0x14047..0x14189",
            "image_index": 0,
            "width_pixels": 56,
            "height_pixels": 57,
            "and_mask_archive": "KOLMPICS.0",
            "and_mask_image_index": 28,
            "composition": "(destination AND mask) OR color",
        },
        "archive_loader_original_function": "sub_15920",
        "archive_loader_ea_range": "0x15920..0x15abf",
        "filename_patterns": ["twn%d%s.0", "twn%d%s.1"],
        "limitation": "fixed archive counts, image 1 placeholders and per-item decoded pixel counts require fixed-file evidence",
    }
    summary["functions"] = [
        {
            "original_name": item["original_name"],
            "start_ea": item["start_ea"],
            "end_ea": item["end_ea"],
            "instruction_count": len(item["instructions"]),
            "caller_count": len(item["callers"]),
            "callers": item["callers"],
        }
        for item in functions
    ]
    summary["call_sites"] = [
        {
            "call_ea": item["call_ea"],
            "caller_original_name": item["caller_original_name"],
            "caller_start_ea": item["caller_start_ea"],
        }
        for item in call_sites
    ]
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN TWN archive probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
