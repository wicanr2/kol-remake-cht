"""IDA Pro 9.4：匯出 TITLE.EXE BACKPICS selector 的全部 caller。"""

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


EXPECTED_SHA256 = "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6"
SELECTOR_EA = 0x10E9F
PIPELINE_EAS = (0x11133, 0x11291, 0x114F8, 0x11728, 0x11A69)


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
    }


def export_function(start: int) -> dict:
    function = ida_funcs.get_func(start)
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    selector = ida_funcs.get_func(SELECTOR_EA)
    if selector is None or selector.start_ea != SELECTOR_EA:
        raise RuntimeError("BACKPICS selector boundary not found")
    call_eas = sorted(idautils.CodeRefsTo(SELECTOR_EA, False))
    caller_starts = sorted({ida_funcs.get_func(ea).start_ea for ea in call_eas})
    functions = ([export_function(SELECTOR_EA)]
                 + [export_function(ea) for ea in PIPELINE_EAS]
                 + [export_function(ea) for ea in caller_starts])
    calls = []
    image_indices = []
    for call_ea in call_eas:
        function = ida_funcs.get_func(call_ea)
        items = list(idautils.FuncItems(function.start_ea))
        index = items.index(call_ea)
        if index < 3 or idc.print_insn_mnem(items[index - 2]) != "mov":
            raise RuntimeError(f"unexpected selector argument setup at {call_ea:#x}")
        image_index = idc.get_operand_value(items[index - 2], 1)
        if idc.print_insn_mnem(items[index - 1]) != "push":
            raise RuntimeError(f"missing selector index push at {call_ea:#x}")
        image_indices.append(image_index)
        calls.append({
            "call_ea": f"0x{call_ea:x}",
            "caller_original_name": idc.get_func_name(function.start_ea),
            "image_index": image_index,
            "preceding_instructions": [instruction(ea) for ea in items[max(0, index - 5):index]],
        })
    if sorted(set(image_indices)) != list(range(14)):
        raise RuntimeError(f"BACKPICS caller coverage mismatch: {sorted(set(image_indices))}")
    summary = {
        "schema": "kol-ida-title-backpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "selector_original_name": idc.get_func_name(SELECTOR_EA),
        "selector_start_ea": f"0x{SELECTOR_EA:x}",
        "selector_end_ea": f"0x{selector.end_ea:x}",
        "call_count": len(call_eas),
        "call_eas": [f"0x{ea:x}" for ea in call_eas],
        "image_indices": sorted(set(image_indices)),
        "call_records": [{
            "call_ea": call["call_ea"],
            "caller_original_name": call["caller_original_name"],
            "image_index": call["image_index"],
        } for call in calls],
        "caller_functions": [{
            "original_name": idc.get_func_name(ea),
            "start_ea": f"0x{ea:x}",
            "end_ea": f"0x{ida_funcs.get_func(ea).end_ea:x}",
        } for ea in caller_starts],
        "pipeline_functions": [{
            "original_name": idc.get_func_name(ea),
            "start_ea": f"0x{ea:x}",
            "end_ea": f"0x{ida_funcs.get_func(ea).end_ea:x}",
        } for ea in PIPELINE_EAS],
        "archive_contract": {
            "archive_root": "BACKPICS",
            "image_indices": list(range(14)),
            "width_pixels": 320,
            "height_pixels": 200,
            "color_index_bits": 4,
            "transparent": False,
            "display_policy": "full-screen direct blit",
        },
        "semantic_assessment": {
            "selector_and_all_callers": "confirmed",
            "lzw_sub_11291": "confirmed",
            "second_stage_rle_ega_sub_114F8": "confirmed",
            "mode_0x0d_geometry_and_blit_sub_11133": "confirmed",
            "rgb_palette": "hardware-spec approximation",
        },
    }
    private = dict(summary)
    private["schema"] = "kol-ida-title-backpics-private-v1"
    private["calls"] = calls
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"TITLE BACKPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
