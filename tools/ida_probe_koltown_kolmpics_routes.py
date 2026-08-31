"""IDA Pro 9.4：固定 KOLTOWN KOLMPICS request router 與所有 caller。"""

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
ROUTER_EA = 0x1103F
DYNAMIC_HELPERS = [0x11722, 0x23638, 0x27E20]
EXPECTED_INSTRUCTIONS = {
    0x10359: "mov     ax, word_39820",
    0x1038E: "call    sub_19D98",
    0x11048: "cmp     [bp+arg_0], 2",
    0x11051: "cmp     [bp+arg_0], 1Dh",
    0x11072: "cmp     [bp+arg_0], 2",
    0x1107B: "cmp     [bp+arg_0], 1Eh",
    0x11084: "cmp     [bp+arg_0], 23h",
    0x11153: "cmp     [bp+arg_0], 64h",
    0x1115C: "cmp     [bp+arg_0], 1Ch",
    0x11165: "cmp     [bp+arg_0], 24h",
    0x1116E: "cmp     [bp+arg_0], 1Fh",
    0x11177: "cmp     [bp+arg_0], 46h",
    0x11180: "cmp     [bp+arg_0], 4Ah",
    0x1124D: "push    word_39820",
    0x11254: "call    sub_11262",
    0x11746: "add     ax, 4",
    0x11753: "add     ax, 7",
    0x11763: "add     ax, 3",
    0x140BB: "mov     word_34A06, 7",
    0x140C1: "mov     word_34A0E, 39h",
    0x14158: "mov     ax, 1Ch",
    0x1415C: "call    sub_1A429",
    0x1417B: "call    sub_1BD17",
    0x2365F: "add     ax, 3",
    0x23668: "add     ax, 14h",
    0x23671: "add     ax, 3",
    0x27E29: "push    [bp+arg_0]",
    0x27E2C: "call    sub_1103F",
    0x27E34: "sub     [bp+arg_0], 47h",
}


def normalized_disassembly(ea: int) -> str:
    text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return text.split(" ;", 1)[0].rstrip()


def instruction(ea: int) -> dict:
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
        raise RuntimeError(f"missing function at 0x{start:x}")
    return {
        "original_name": idc.get_func_name(function.start_ea),
        "start_ea": f"0x{function.start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instructions": [instruction(ea) for ea in idautils.FuncItems(function.start_ea)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    checks = []
    for ea, expected in EXPECTED_INSTRUCTIONS.items():
        actual = normalized_disassembly(ea)
        if actual != expected:
            raise RuntimeError(f"instruction mismatch at 0x{ea:x}: {actual!r} != {expected!r}")
        checks.append(instruction(ea))
    calls = sorted(idautils.CodeRefsTo(ROUTER_EA, False))
    starts = {ROUTER_EA, 0x11262, 0x14047, 0x19E9D, 0x1A429, *DYNAMIC_HELPERS}
    for call in calls:
        function = ida_funcs.get_func(call)
        if function is None:
            raise RuntimeError(f"router call outside function at 0x{call:x}")
        starts.add(function.start_ea)
    caller_windows = []
    for call in calls:
        function = ida_funcs.get_func(call)
        items = list(idautils.FuncItems(function.start_ea))
        position = items.index(call)
        caller_windows.append({
            "call_ea": f"0x{call:x}",
            "source_original_function": idc.get_func_name(function.start_ea),
            "source_start_ea": f"0x{function.start_ea:x}",
            "window": [instruction(ea) for ea in items[max(0, position - 20):position + 3]],
        })
    helper_calls = {}
    helper_caller_windows = []
    for helper in DYNAMIC_HELPERS:
        refs = sorted(idautils.CodeRefsTo(helper, False))
        helper_calls[f"0x{helper:x}"] = [f"0x{x:x}" for x in refs]
        for call in refs:
            function = ida_funcs.get_func(call)
            if function is None:
                raise RuntimeError(f"helper call outside function at 0x{call:x}")
            starts.add(function.start_ea)
            items = list(idautils.FuncItems(function.start_ea))
            position = items.index(call)
            helper_caller_windows.append({
                "helper_ea": f"0x{helper:x}", "call_ea": f"0x{call:x}",
                "source_original_function": idc.get_func_name(function.start_ea),
                "source_start_ea": f"0x{function.start_ea:x}",
                "window": [instruction(ea) for ea in items[max(0, position - 28):position + 3]],
            })
    output = {
        "schema": "kol-ida-koltown-kolmpics-routes-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "router_original_name": idc.get_func_name(ROUTER_EA),
        "router_ea": f"0x{ROUTER_EA:x}",
        "router_call_eas": [f"0x{x:x}" for x in calls],
        "caller_windows": caller_windows,
        "dynamic_helper_call_eas": helper_calls,
        "dynamic_helper_caller_windows": helper_caller_windows,
        "instruction_checks": checks,
        "functions": [function_record(ea) for ea in sorted(starts)],
    }
    safe = {
        "schema": "kol-ida-koltown-kolmpics-routes-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": output["address_space"],
        "archive_load": {
            "filename_format": "kolm%s.0", "archive_buffer_original_name": "word_39820",
            "whole_file_reader_call_ea": "0x1038e", "status": "confirmed",
        },
        "router": {
            "original_name": "sub_1103F", "start_ea": "0x1103f",
            "kolmpics_candidate_indices_within_archive": [
                *range(0, 29), 31, *range(36, 71), *range(74, 83)
            ],
            "requests_routed_elsewhere": [29, 30, 32, 33, 34, 35, 71, 72, 73, 100],
            "status": "confirmed",
        },
        "geometry_contracts": [
            {"request_or_index": 0, "width_pixels": 320, "height_pixels": 200,
             "basis": "sub_1103F default rectangle", "status": "confirmed"},
            {"request_or_index": 2, "width_pixels": 296, "height_pixels": 32,
             "basis": "sub_1103F request 2 rectangle", "status": "confirmed"},
            {"request_or_index": "runtime table value + 3", "width_pixels": 24, "height_pixels": 20,
             "basis": "sub_11722 fixed rectangle", "status": "confirmed"},
            {"request_or_index": "arg_6 + 3", "width_pixels": 24, "height_pixels": 20,
             "basis": "sub_23638 fixed rectangle", "status": "confirmed"},
            {"request_or_index": "sub_27E20 arg_0", "width_pixels": None, "height_pixels": None,
             "basis": "geometry inherited from caller; 24x20 requires archive/visual evidence",
             "status": "unknown in this IDA slice"},
            {"request_or_index": 28, "width_pixels": 56, "height_pixels": 57,
             "basis": "sub_14047 AND-mask rectangle", "status": "confirmed"},
        ],
        "router_call_eas": output["router_call_eas"],
        "dynamic_helper_call_eas": helper_calls,
        "instruction_checks": [
            {"ea": x["ea"], "bytes": x["bytes"],
             "disassembly": normalized_disassembly(int(x["ea"], 16))} for x in checks
        ],
        "limits": [
            "候選 route 不等於每個索引都有可達 caller；placeholder 與二階像素數由 archive probe 分類。",
            "sub_27E20 的 24x20 幾何不能只由本 IDA 摘要宣稱。",
        ],
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
    safe_path.parent.mkdir(parents=True, exist_ok=True)
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN KOLMPICS routes probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
