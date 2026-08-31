"""IDA Pro 9.4：匯出 MAKEMODE 顯示模式與素材 suffix consumer。"""

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


EXPECTED_SHA256 = "3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56"
KEYWORDS = ("pics", "pcga", "herc", "mode", "ega", "cga")
EXPLICIT_FUNCTION_EAS = (0x11242, 0x1142C, 0x11497, 0x114F4, 0x11734, 0x118A8)


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

    data_segments = []
    for start in idautils.Segments():
        segment = ida_segment.getseg(start)
        if segment and ida_segment.get_segm_class(segment) == "DATA":
            data_segments.append(start)
    if len(data_segments) != 1:
        raise RuntimeError(f"expected one DATA segment, got {data_segments}")
    dgroup_ea = data_segments[0]

    raw_input = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw_input, 0x08)[0] * 16
    load_image = raw_input[header_size:]
    records = []
    consumer_starts = set(EXPLICIT_FUNCTION_EAS)
    for string in idautils.Strings():
        text = str(string)
        if not any(keyword in text.lower() for keyword in KEYWORDS):
            continue
        relative_offset = string.ea - dgroup_ea
        occurrences = []
        if 0 <= relative_offset <= 0xFFFF:
            needle = struct.pack("<H", relative_offset)
            cursor = 0
            while True:
                found = load_image.find(needle, cursor)
                if found < 0:
                    break
                ea = 0x10000 + found
                function = ida_funcs.get_func(ea)
                if function:
                    consumer_starts.add(function.start_ea)
                    occurrences.append(
                        {
                            "ea": f"0x{ea:x}",
                            "source_original_function": idc.get_func_name(function.start_ea),
                            "source_disassembly": ida_lines.tag_remove(
                                idc.generate_disasm_line(ea, 0) or ""
                            ),
                        }
                    )
                cursor = found + 1
        records.append(
            {
                "string_ea": f"0x{string.ea:x}",
                "text": text,
                "dgroup_relative_offset": relative_offset,
                "near_offset_occurrences": occurrences,
            }
        )

    functions = []
    for start in sorted(consumer_starts):
        function = ida_funcs.get_func(start)
        functions.append(
            {
                "original_name": idc.get_func_name(start),
                "start_ea": f"0x{start:x}",
                "end_ea": f"0x{function.end_ea:x}",
                "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
                "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
            }
        )

    conversion_calls = []
    main_function = ida_funcs.get_func(0x10010)
    if main_function is None:
        raise RuntimeError("missing _main at 0x10010")
    main_items = list(idautils.FuncItems(main_function.start_ea))
    for position, ea in enumerate(main_items):
        if 0x10464 not in list(idautils.CodeRefsFrom(ea, False)):
            continue
        if position < 6:
            raise RuntimeError(f"short conversion call context at {ea:#x}")
        context = main_items[position - 6 : position]
        if [idc.print_insn_mnem(item) for item in context] != [
            "mov", "push", "mov", "push", "mov", "push"
        ]:
            raise RuntimeError(f"unexpected conversion call context at {ea:#x}")
        transform_case = idc.get_operand_value(context[0], 1)
        image_count = idc.get_operand_value(context[2], 1)
        base_offset = idc.get_operand_value(context[4], 1)
        base_bytes = ida_bytes.get_strlit_contents(dgroup_ea + base_offset, -1, 0)
        if not base_bytes:
            raise RuntimeError(f"missing base string at DGROUP+{base_offset:#x}")
        conversion_calls.append(
            {
                "call_ea": f"0x{ea:x}",
                "base_name": base_bytes.decode("ascii"),
                "base_dgroup_offset": base_offset,
                "image_count": image_count,
                "transform_case": transform_case,
            }
        )

    summary = {
        "schema": "kol-ida-makemode-display-selection-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for loaded MAKEMODE MZ",
        "dgroup_ea": f"0x{dgroup_ea:x}",
        "strings": [
            {
                "string_ea": record["string_ea"],
                "text": (
                    record["text"]
                    if len(record["text"]) <= 32 and "\n" not in record["text"]
                    else None
                ),
                "text_length": len(record["text"]),
                "text_sha256": hashlib.sha256(record["text"].encode("utf-8")).hexdigest(),
                "dgroup_relative_offset": record["dgroup_relative_offset"],
                "near_offset_occurrence_count": len(record["near_offset_occurrences"]),
                "consumer_original_functions": sorted(
                    {
                        occurrence["source_original_function"]
                        for occurrence in record["near_offset_occurrences"]
                    }
                ),
            }
            for record in records
        ],
        "consumer_functions": [
            {
                "original_name": function["original_name"],
                "start_ea": function["start_ea"],
                "end_ea": function["end_ea"],
                "instruction_count": len(function["instructions"]),
                "caller_count": len(function["callers"]),
            }
            for function in functions
        ],
        "conversion_calls": conversion_calls,
    }
    private = dict(summary)
    private["schema"] = "kol-ida-makemode-display-selection-private-v1"
    private["strings"] = records
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"MAKEMODE display-selection probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
