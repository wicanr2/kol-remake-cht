"""IDA Pro 9.4：探索 KOLTOWN ENDPICS 字串與未切函式的使用點。"""

from __future__ import annotations

import hashlib
import json
import struct
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
STRING_TEXT = "end%s"
CONSUMER_EA = 0x2B80A
LOADER_EA = 0x10F9C
CONTRACTS = (
    (0x2B525, 0, (0x2B506, 1), (0x2B50C, 13), (0x2B512, 22), (0x2B518, 135)),
    (0x2B57D, 1, (0x2B55E, 15), (0x2B564, 41), (0x2B56A, 39), (0x2B570, 152)),
    (0x2B5D5, 2, (0x2B5B6, 1), (0x2B5BC, 50), (0x2B5C2, 38), (0x2B5C8, 143)),
    (0x2B62D, 3, (0x2B60E, 1), (0x2B614, 50), (0x2B61A, 23), (0x2B620, 147)),
    (0x2B695, 4, (0x2B676, 17), (0x2B67C, 12), (0x2B682, 38), (0x2B688, 135)),
    (0x2B6ED, 5, (0x2B6CE, 0), (0x2B6D4, 46), (0x2B6DA, 24), (0x2B6E0, 148)),
    (0x2B755, 6, (0x2B736, 16), (0x2B73C, 45), (0x2B742, 39), (0x2B748, 143)),
)


def describe_function(function) -> dict | None:
    if function is None:
        return None
    return {
        "original_name": idc.get_func_name(function.start_ea),
        "start_ea": f"0x{function.start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
    }


def head(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
    }


def export_function(start: int) -> dict:
    function = ida_funcs.get_func(start)
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instructions": [head(ea) for ea in idautils.FuncItems(start)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    matches = [string for string in idautils.Strings() if str(string) == STRING_TEXT]
    if len(matches) != 1:
        raise RuntimeError(f"unexpected end string count: {len(matches)}")
    string = matches[0]
    relative_offset = string.ea - DGROUP_EA
    raw = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    load_image = raw[header_size:]
    needle = struct.pack("<H", relative_offset)
    occurrences = []
    cursor = 0
    while True:
        found = load_image.find(needle, cursor)
        if found < 0:
            break
        occurrences.append(0x10000 + found)
        cursor = found + 1
    if len(occurrences) != 1:
        raise RuntimeError(f"unexpected end offset occurrences: {occurrences}")
    occurrence = occurrences[0]
    consumer = ida_funcs.get_func(CONSUMER_EA)
    if consumer is None or consumer.start_ea != CONSUMER_EA:
        raise RuntimeError("ENDPICS consumer boundary not found")
    call_eas = sorted(idautils.CodeRefsTo(CONSUMER_EA, False))
    caller_starts = sorted({ida_funcs.get_func(ea).start_ea for ea in call_eas})
    if call_eas != [item[0] for item in CONTRACTS]:
        raise RuntimeError(f"unexpected ENDPICS call set: {call_eas}")
    consumer_groups = []
    for call_ea, image_index, x0_pair, y0_pair, x1_pair, y1_pair in CONTRACTS:
        previous = idc.prev_head(call_ea)
        argument = idc.prev_head(previous)
        if idc.print_insn_mnem(previous) != "push" or idc.get_operand_value(argument, 1) != image_index:
            raise RuntimeError(f"unexpected image argument at {call_ea:#x}")
        for assignment_ea, expected_value in (x0_pair, y0_pair, x1_pair, y1_pair):
            if idc.print_insn_mnem(assignment_ea) != "mov" or idc.get_operand_value(assignment_ea, 1) != expected_value:
                raise RuntimeError(f"unexpected rectangle value at {assignment_ea:#x}")
        x0, y0, x1, y1 = x0_pair[1], y0_pair[1], x1_pair[1], y1_pair[1]
        consumer_groups.append({
            "call_ea": f"0x{call_ea:x}",
            "image_index": image_index,
            "rectangle_assignment_eas": [f"0x{x[0]:x}" for x in (x0_pair, y0_pair, x1_pair, y1_pair)],
            "rectangle": {"x0_byte": x0, "y0": y0, "x1_byte": x1, "y1": y1,
                          "width_pixels": (x1 - x0) * 8, "height_pixels": y1 - y0},
        })
    surrounding = []
    ea = occurrence
    for _ in range(50):
        ea = idc.prev_head(ea)
    for _ in range(110):
        surrounding.append(head(ea))
        ea = idc.next_head(ea)
    summary = {
        "schema": "kol-ida-koltown-endpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "string_ea": f"0x{string.ea:x}",
        "dgroup_relative_offset": relative_offset,
        "raw_occurrence_ea": f"0x{occurrence:x}",
        "containing_function": describe_function(ida_funcs.get_func(occurrence)),
        "previous_function": describe_function(ida_funcs.get_prev_func(occurrence)),
        "next_function": describe_function(ida_funcs.get_next_func(occurrence)),
        "consumer": describe_function(consumer),
        "consumer_call_eas": [f"0x{ea:x}" for ea in call_eas],
        "consumer_callers": [describe_function(ida_funcs.get_func(ea)) for ea in caller_starts],
        "consumer_groups": consumer_groups,
        "pipeline_functions": [
            {"original_name": "sub_10F9C", "start_ea": "0x10f9c", "role": "indexed image route"},
            {"original_name": "sub_1A4FE", "start_ea": "0x1a4fe", "role": "archive selection and LZW"},
            {"original_name": "sub_1C907", "start_ea": "0x1c907", "role": "display-mode conversion route"},
            {"original_name": "sub_1A2BB", "start_ea": "0x1a2bb", "role": "second-stage RLE and EGA planes"},
            {"original_name": "sub_19E9D", "start_ea": "0x19e9d", "role": "rectangle blit"},
        ],
        "semantic_assessment": {
            "filename_consumer_and_call_set": "confirmed",
            "image_indices_and_rectangles": "confirmed",
            "lzw_and_second_stage_ega": "confirmed by shared pipeline plus payload validation",
            "rgb_palette": "hardware-spec approximation",
        },
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-endpics-private-v1"
    private["surrounding_heads"] = surrounding
    private["functions"] = ([export_function(CONSUMER_EA), export_function(LOADER_EA)]
                            + [export_function(ea) for ea in caller_starts])
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN ENDPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
