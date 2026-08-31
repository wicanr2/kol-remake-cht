"""IDA Pro 9.4：匯出 KOLBOOK 的 AWD／角色圖片檔讀寫 consumer。"""

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
KEYWORDS = ("awd", "charpics", "shldpics", "steedpic")
EXPLICIT_FUNCTION_EAS = (0x1374E, 0x137C0)


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
        near_occurrences = []
        if 0 <= relative_offset <= 0xFFFF:
            needle = struct.pack("<H", relative_offset)
            cursor = 0
            while True:
                found = load_image.find(needle, cursor)
                if found < 0:
                    break
                occurrence_ea = 0x10000 + found
                function = ida_funcs.get_func(occurrence_ea)
                if function:
                    consumer_starts.add(function.start_ea)
                    near_occurrences.append(
                        {
                            "ea": f"0x{occurrence_ea:x}",
                            "source_original_function": idc.get_func_name(function.start_ea),
                            "source_disassembly": ida_lines.tag_remove(
                                idc.generate_disasm_line(occurrence_ea, 0) or ""
                            ),
                        }
                    )
                cursor = found + 1
        records.append(
            {
                "string_ea": f"0x{string.ea:x}",
                "text": text,
                "dgroup_relative_offset": relative_offset,
                "near_offset_occurrences": near_occurrences,
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

    summary = {
        "schema": "kol-ida-kolbook-asset-io-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for loaded KOLBOOK MZ",
        "dgroup_ea": f"0x{dgroup_ea:x}",
        "strings": [
            {
                "string_ea": record["string_ea"],
                "text": record["text"],
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
    }
    private = dict(summary)
    private["schema"] = "kol-ida-kolbook-asset-io-private-v1"
    private["strings"] = records
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK asset I/O probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
