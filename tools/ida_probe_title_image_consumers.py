"""IDA Pro 9.4：探索 TITLE.EXE 的圖像檔名字串與近指標 consumer。"""

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


EXPECTED_SHA256 = "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6"
TARGET_PREFIXES = ("set0", "set1", "set2", "face", "back")
# 字串 consumer 緊接著呼叫的候選載入／繪圖函式；僅供探索，語意不由名稱推定。
ADDITIONAL_FUNCTION_EAS = (
    0x10AF0, 0x10E04, 0x11022, 0x11058, 0x11133, 0x11666, 0x1170E,
    0x11728, 0x11A69,
)


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
    raw_input = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw_input, 0x08)[0] * 16
    load_image = raw_input[header_size:]
    strings = [
        string for string in idautils.Strings()
        if any(prefix in str(string).lower() for prefix in TARGET_PREFIXES)
    ]
    if not strings:
        raise RuntimeError("target strings not found")
    segment_starts = {
        ida_segment.getseg(string.ea).start_ea
        for string in strings if ida_segment.getseg(string.ea)
    }
    if len(segment_starts) != 1:
        raise RuntimeError(f"target strings span unexpected segments: {segment_starts}")
    dgroup_ea = next(iter(segment_starts))

    records = []
    consumer_starts = set()
    for string in strings:
        text = str(string)
        lowered = text.lower()
        prefix_positions = [lowered.index(prefix) for prefix in TARGET_PREFIXES if prefix in lowered]
        effective_ea = string.ea + min(prefix_positions)
        xrefs = []
        for xref in idautils.XrefsTo(effective_ea, 0):
            function = ida_funcs.get_func(xref.frm)
            if function:
                consumer_starts.add(function.start_ea)
            xrefs.append({
                "from_ea": f"0x{xref.frm:x}",
                "xref_type": xref.type,
                "source_original_function": idc.get_func_name(function.start_ea) if function else "",
                "source_disassembly": ida_lines.tag_remove(idc.generate_disasm_line(xref.frm, 0) or ""),
            })
        relative_offset = effective_ea - dgroup_ea
        near_occurrences = []
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
                near_occurrences.append({
                    "ea": f"0x{occurrence_ea:x}",
                    "source_original_function": idc.get_func_name(function.start_ea),
                    "source_disassembly": ida_lines.tag_remove(idc.generate_disasm_line(occurrence_ea, 0) or ""),
                })
            cursor = found + 1
        records.append({
            "string_ea": f"0x{string.ea:x}",
            "effective_string_ea": f"0x{effective_ea:x}",
            "text": text,
            "dgroup_relative_offset": relative_offset,
            "xrefs": xrefs,
            "near_offset_occurrences": near_occurrences,
        })

    consumer_starts.update(
        function.start_ea
        for ea in ADDITIONAL_FUNCTION_EAS
        if (function := ida_funcs.get_func(ea)) is not None
    )
    functions = []
    for start in sorted(consumer_starts):
        function = ida_funcs.get_func(start)
        functions.append({
            "original_name": idc.get_func_name(start),
            "start_ea": f"0x{start:x}",
            "end_ea": f"0x{function.end_ea:x}",
            "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
            "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
        })

    summary = {
        "schema": "kol-ida-title-image-consumers-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "dgroup_ea": f"0x{dgroup_ea:x}",
        "strings": [{
            "string_ea": record["string_ea"],
            "effective_string_ea": record["effective_string_ea"],
            "text": record["text"],
            "dgroup_relative_offset": record["dgroup_relative_offset"],
            "xref_count": len(record["xrefs"]),
            "near_offset_occurrence_count": len(record["near_offset_occurrences"]),
            "consumer_original_functions": sorted({
                item["source_original_function"]
                for item in record["xrefs"] + record["near_offset_occurrences"]
                if item["source_original_function"]
            }),
        } for record in records],
        "consumer_functions": [{
            "original_name": function["original_name"],
            "start_ea": function["start_ea"],
            "end_ea": function["end_ea"],
            "instruction_count": len(function["instructions"]),
            "caller_count": len(function["callers"]),
        } for function in functions],
        "archive_contracts": [
            {
                "archive_root": "SET0PICS",
                "load_call_ea": "0x1021b",
                "presentation_call_ea": "0x102a6",
                "presentation_argument": 0,
                "image_indices": [0, 1, 2, 3, 4, 5],
                "rectangle": {"x0_byte": 14, "y0": 53, "x1_byte": 25, "y1": 123,
                              "width_pixels": 88, "height_pixels": 70},
            },
            {
                "archive_root": "SET1PICS",
                "load_call_ea": "0x103e0",
                "presentation_call_ea": "0x103f8",
                "presentation_argument": 1,
                "image_indices": [0, 1, 2, 3, 4, 5],
                "rectangle": {"x0_byte": 12, "y0": 53, "x1_byte": 25, "y1": 137,
                              "width_pixels": 104, "height_pixels": 84},
            },
            {
                "archive_root": "SET2PICS",
                "load_call_ea": "0x10558",
                "presentation_call_ea": "0x10570",
                "presentation_argument": 2,
                "image_indices": [0, 1, 2, 3, 4, 5],
                "rectangle": {"x0_byte": 6, "y0": 53, "x1_byte": 26, "y1": 178,
                              "width_pixels": 160, "height_pixels": 125},
            },
            {
                "archive_root": "FACEPICS",
                "load_call_ea": "0x10723",
                "presentation_call_eas": [
                    "0x1074b", "0x10785", "0x1079c", "0x107b3", "0x107ca", "0x107e1",
                    "0x10833", "0x1086d", "0x10884", "0x1089b", "0x108b2", "0x10904",
                    "0x1093e", "0x10955", "0x1096c", "0x10983", "0x1099a", "0x109b1",
                ],
                "image_indices": [0, 1, 2, 3, 4, 5],
                "rectangle": {"x0_byte": 13, "y0": 80, "x1_byte": 27, "y1": 160,
                              "width_pixels": 112, "height_pixels": 80},
            },
        ],
        "semantic_assessment": {
            "whole_file_loader_sub_11058": "confirmed",
            "indexed_payload_selector_sub_11666": "confirmed",
            "little_endian_word_reader_sub_1170E": "confirmed",
            "rectangle_blit_sub_11133": "confirmed for mode 0x0d geometry",
            "second_stage_rle_and_ega_pixels": "strong inference pending payload validation",
        },
    }
    private = dict(summary)
    private["schema"] = "kol-ida-title-image-consumers-private-v1"
    private["strings"] = records
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"TITLE image consumer probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
