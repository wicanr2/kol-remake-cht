"""IDA Pro 9.4：驗證 KOLTOWN WEATPICS 六組圖層的載入與合成契約。"""

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
STRING_TEXT = "weat%s.0"
BUFFER_SEGMENT_EA = 0x2DF7C
INDEXED_LOADER_EA = 0x1A4FE
WEATHER_GROUP_LOADER_EA = 0x2A0C6
EXPLICIT_FUNCTION_EAS = (0x2A01C, WEATHER_GROUP_LOADER_EA, 0x2A18E)


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {"ea": f"0x{ea:x}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
            "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
            "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)]}


def describe(function) -> dict | None:
    if function is None:
        return None
    return {"original_name": idc.get_func_name(function.start_ea),
            "start_ea": f"0x{function.start_ea:x}", "end_ea": f"0x{function.end_ea:x}"}


def export_function(start: int) -> dict:
    function = ida_funcs.get_func(start)
    return {**describe(function), "instructions": [instruction(ea) for ea in idautils.FuncItems(start)]}


def selected_instructions(addresses: tuple[int, ...]) -> list[dict]:
    return [instruction(ea) for ea in addresses]


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    strings = [string for string in idautils.Strings() if str(string) == STRING_TEXT]
    if len(strings) != 1:
        raise RuntimeError(f"unexpected weather string count: {len(strings)}")
    string = strings[0]
    relative = string.ea - DGROUP_EA
    raw = input_path.read_bytes()
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    load_image = raw[header_size:]
    needle = struct.pack("<H", relative)
    occurrences = [0x10000 + i for i in range(len(load_image)) if load_image.startswith(needle, i)]
    records, starts = [], set()
    for ea in occurrences:
        function = ida_funcs.get_func(ea)
        if function:
            starts.add(function.start_ea)
        records.append({"ea": f"0x{ea:x}", "containing_function": describe(function),
                        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")})
    buffer_refs = []
    for ea in idautils.DataRefsTo(BUFFER_SEGMENT_EA):
        function = ida_funcs.get_func(ea)
        if function:
            starts.add(function.start_ea)
        buffer_refs.append({"ea": f"0x{ea:x}", "function": describe(function),
                            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")})
    loader_calls = []
    for call_ea in sorted(idautils.CodeRefsTo(INDEXED_LOADER_EA, False)):
        function = ida_funcs.get_func(call_ea)
        if function:
            starts.add(function.start_ea)
            items = list(idautils.FuncItems(function.start_ea))
            index = items.index(call_ea)
            preceding = [instruction(ea) for ea in items[max(0, index - 12):index]]
        else:
            preceding = []
        loader_calls.append({"call_ea": f"0x{call_ea:x}", "caller": describe(function),
                             "preceding_instructions": preceding})
    starts.update(EXPLICIT_FUNCTION_EAS)
    weather_calls = []
    for call_ea in sorted(idautils.CodeRefsTo(WEATHER_GROUP_LOADER_EA, False)):
        function = ida_funcs.get_func(call_ea)
        if function:
            starts.add(function.start_ea)
            items = list(idautils.FuncItems(function.start_ea))
            index = items.index(call_ea)
            preceding = [instruction(ea) for ea in items[max(0, index - 8):index]]
        else:
            preceding = []
        weather_calls.append({"call_ea": f"0x{call_ea:x}", "caller": describe(function),
                              "preceding_instructions": preceding})
    if [item["call_ea"] for item in weather_calls] != ["0x290f5", "0x298a1", "0x29900", "0x29b1f"]:
        raise RuntimeError("unexpected WEATPICS group-loader call sites")
    loader_contract_eas = (0x2A0D5, 0x2A0D8, 0x2A0DE, 0x2A0E2, 0x2A0ED, 0x2A0F4,
                           0x2A100, 0x2A109, 0x2A112, 0x2A119, 0x2A128, 0x2A12E,
                           0x2A160, 0x2A166, 0x2A173, 0x2A179, 0x2A17E, 0x2A185)
    compositor_contract_eas = (0x2A1D1, 0x2A1D4, 0x2A1D6, 0x2A1D9, 0x2A1DB,
                               0x2A1DE, 0x2A1E0, 0x2A1E4, 0x2A1E8, 0x2A1F1,
                               0x2A1F9, 0x2A26B, 0x2A270, 0x2A275, 0x2A277,
                               0x2A27B, 0x2A27D, 0x2A281, 0x2A283, 0x2A293)
    summary = {"schema": "kol-ida-koltown-weatpics-discovery-summary-v1",
               "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
               "input_sha256": digest,
               "address_space": "IDA linear effective address for reconstructed unpacked MZ",
               "string_ea": f"0x{string.ea:x}", "dgroup_relative_offset": relative,
               "raw_occurrences": records, "buffer_segment_ea": f"0x{BUFFER_SEGMENT_EA:x}",
               "buffer_data_refs": buffer_refs,
               "indexed_loader_call_count": len(loader_calls),
               "weather_group_loader_ea": f"0x{WEATHER_GROUP_LOADER_EA:x}",
               "weather_group_call_count": len(weather_calls),
               "group_contract": {
                   "group_count_from_archive": 6, "items_per_group": 3,
                   "width_pixels": 64, "height_pixels": 62,
                   "roles": ["base_color", "or_color", "and_mask"],
                   "item_index_formula": "3 * group + role_offset",
                   "four_plane_mask_groups": [3, 4], "one_bit_mask_groups": [0, 1, 2, 5],
                   "composition": "result = (base_color AND and_mask) OR or_color",
                   "confidence": "confirmed",
               },
               "loader_contract_instructions": selected_instructions(loader_contract_eas),
               "compositor_contract_instructions": selected_instructions(compositor_contract_eas),
               "evidence_note": "Original names, linear EAs, bytes and disassembly are retained; semantic labels are additive.",
               "semantic_classification": {"label": "wilderness weather/time-of-day background candidates",
                                           "confidence": "strong inference",
                                           "basis": "six human-inspected composites; not asserted by code names alone"}}
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-weatpics-discovery-private-v1"
    private["functions"] = [export_function(ea) for ea in sorted(starts)]
    private["indexed_loader_calls"] = loader_calls
    private["weather_group_calls"] = weather_calls
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN WEATPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
