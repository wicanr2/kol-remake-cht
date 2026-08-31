"""IDA Pro 9.4：探索 KOLBOOK.EXE 的 BOOKPICS 載入與 consumer。"""

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
import ida_segment
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
STRING_TEXT = "book%s.0"
BOOK_BUFFER_SEGMENT_EA = 0x1A472
CONSUMER_EA = 0x1137B
EXPLICIT_FUNCTION_EAS = (CONSUMER_EA, 0x13843, 0x13D7D, 0x146CE, 0x15702)


def instruction(ea: int) -> dict:
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
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def describe(function) -> dict | None:
    if function is None:
        return None
    return {"original_name": idc.get_func_name(function.start_ea),
            "start_ea": f"0x{function.start_ea:x}", "end_ea": f"0x{function.end_ea:x}"}


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    data_segments = [start for start in idautils.Segments()
                     if ida_segment.get_segm_class(ida_segment.getseg(start)) == "DATA"]
    if len(data_segments) != 1:
        raise RuntimeError(f"unexpected DATA segments: {data_segments}")
    dgroup_ea = data_segments[0]
    strings = [string for string in idautils.Strings() if str(string) == STRING_TEXT]
    if len(strings) != 1:
        raise RuntimeError(f"unexpected book string count: {len(strings)}")
    string = strings[0]
    relative = string.ea - dgroup_ea
    raw = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    load_image = raw[header_size:]
    needle = struct.pack("<H", relative)
    occurrences = [0x10000 + i for i in range(len(load_image)) if load_image.startswith(needle, i)]
    records = []
    starts = set()
    for ea in occurrences:
        function = ida_funcs.get_func(ea)
        if function:
            starts.add(function.start_ea)
        records.append({"ea": f"0x{ea:x}", "containing_function": describe(function),
                        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")})
    buffer_refs = []
    for ea in idautils.DataRefsTo(BOOK_BUFFER_SEGMENT_EA):
        function = ida_funcs.get_func(ea)
        if function:
            starts.add(function.start_ea)
        buffer_refs.append({"ea": f"0x{ea:x}", "function": describe(function),
                            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")})
    starts.update(EXPLICIT_FUNCTION_EAS)
    consumer_call_eas = sorted(idautils.CodeRefsTo(CONSUMER_EA, False))
    call_records = []
    for call_ea in consumer_call_eas:
        function = ida_funcs.get_func(call_ea)
        if function:
            starts.add(function.start_ea)
        previous = idc.prev_head(call_ea)
        argument = idc.prev_head(previous)
        constant = None
        if idc.print_insn_mnem(previous) == "push" and idc.print_insn_mnem(argument) == "mov":
            constant = idc.get_operand_value(argument, 1)
        call_records.append({
            "call_ea": f"0x{call_ea:x}", "caller": describe(function),
            "argument_constant": constant,
            "argument_instructions": [instruction(argument), instruction(previous)],
        })
    constants = [record["argument_constant"] for record in call_records]
    if any(value is None for value in constants) or sorted(set(constants)) != list(range(14)) + [100]:
        raise RuntimeError(f"unexpected BOOKPICS consumer arguments: {sorted(set(constants))}")
    expected_operands = {
        0x11396: 100, 0x1139F: 9,
        0x113BD: 0, 0x113C3: 0, 0x113C9: 40, 0x113CF: 200,
        0x113DE: 4, 0x113E4: 16, 0x113EA: 18, 0x113F0: 47,
        0x11408: 22, 0x1140E: 36,
        0x1141D: 4, 0x11423: 170, 0x11429: 35, 0x1142F: 192,
    }
    for ea, value in expected_operands.items():
        if idc.get_operand_value(ea, 1) != value:
            raise RuntimeError(f"unexpected BOOKPICS geometry operand at {ea:#x}")
    consumer_groups = []
    for image_index in range(14):
        if image_index <= 3:
            x0, y0, x1, y1, branch = 0, 0, 40, 200, "default_full_screen"
        elif image_index in (9, 11):
            x0, y0, x1, y1, branch = 22, 16, 36, 47, "small_right"
        elif image_index == 13:
            x0, y0, x1, y1, branch = 4, 170, 35, 192, "wide_bottom"
        else:
            x0, y0, x1, y1, branch = 4, 16, 18, 47, "small_left"
        consumer_groups.append({
            "image_index": image_index, "geometry_branch": branch,
            "rectangle": {"x0_byte": x0, "y0": y0, "x1_byte": x1, "y1": y1,
                          "width_pixels": (x1 - x0) * 8, "height_pixels": y1 - y0},
        })
    summary = {
        "schema": "kol-ida-kolbook-bookpics-summary-v1",
        "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
        "input_sha256": digest, "address_space": "IDA linear effective address for loaded KOLBOOK MZ",
        "dgroup_ea": f"0x{dgroup_ea:x}", "string_ea": f"0x{string.ea:x}",
        "dgroup_relative_offset": relative, "raw_occurrences": records,
        "book_buffer_segment_ea": f"0x{BOOK_BUFFER_SEGMENT_EA:x}",
        "book_buffer_data_refs": buffer_refs,
        "consumer_ea": f"0x{CONSUMER_EA:x}",
        "consumer_call_records": [{
            "call_ea": record["call_ea"], "caller": record["caller"],
            "argument_constant": record["argument_constant"],
        } for record in call_records],
        "consumer_groups": consumer_groups,
        "special_argument_alias": {"argument": 100, "image_index": 9,
                                   "comparison_ea": "0x11396", "assignment_ea": "0x1139f"},
        "pipeline_functions": [
            {"original_name": "sub_13D7D", "start_ea": "0x13d7d", "role": "indexed LZW decode"},
            {"original_name": "sub_15702", "start_ea": "0x15702", "role": "second-stage RLE and EGA planes"},
            {"original_name": "sub_13843", "start_ea": "0x13843", "role": "rectangle blit"},
        ],
        "semantic_assessment": {
            "whole_archive_load": "confirmed", "all_consumer_calls": "confirmed",
            "image_indices_and_geometry": "confirmed", "lzw_and_second_stage_ega": "confirmed",
            "rgb_palette": "hardware-spec approximation",
        },
    }
    private = dict(summary)
    private["schema"] = "kol-ida-kolbook-bookpics-private-v1"
    private["functions"] = [export_function(ea) for ea in sorted(starts)]
    private["consumer_call_records"] = call_records
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK BOOKPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
