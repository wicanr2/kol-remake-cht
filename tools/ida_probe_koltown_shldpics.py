"""IDA Pro 9.4：探索 KOLTOWN SHLDPICS.0 與無副檔名 shldpics consumer。"""

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
EXPECTED_DISASSEMBLY = {
    0x1920E: "and     ax, 0Fh", 0x19211: "mov     [bp+var_8], ax",
    0x1925F: "mov     bx, [bp+var_8]", 0x19262: "shl     bx, 1",
    0x19264: "mov     al, [bx+43Ah]", 0x19275: "mov     ax, [bp+var_8]",
    0x19278: "shl     ax, 1", 0x19281: "mov     al, [bx+1]",
    0x1929E: "mov     ax, 60Eh", 0x192B6: "mov     ax, [bp+var_8]",
    0x192B9: "shl     ax, 1", 0x192BB: "add     ax, 1",
    0x192BF: "mov     ax, [bp+var_8]", 0x192C2: "shl     ax, 1",
    0x192C5: "call    sub_16FA6", 0x16FDD: "push    [bp+arg_0]",
    0x16FE0: "push    [bp+arg_4]", 0x16FE3: "call    sub_1A4FE",
    0x17065: "push    [bp+arg_2]", 0x17068: "push    [bp+arg_4]",
    0x1706B: "call    sub_1A4FE", 0x17087: "call    sub_1BD17",
}


def insn(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {"ea": f"0x{ea:x}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
            "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
            "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)]}


def desc(function) -> dict | None:
    if function is None:
        return None
    return {"original_name": idc.get_func_name(function.start_ea),
            "start_ea": f"0x{function.start_ea:x}", "end_ea": f"0x{function.end_ea:x}"}


def export(start: int) -> dict:
    function = ida_funcs.get_func(start)
    return {**desc(function), "instructions": [insn(ea) for ea in idautils.FuncItems(start)]}


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    raw = input_path.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    for ea, expected in EXPECTED_DISASSEMBLY.items():
        actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        if actual != expected:
            raise RuntimeError(f"unexpected instruction at 0x{ea:x}: {actual!r}")
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    load = raw[header_size:]
    starts, strings = set(), []
    for string in idautils.Strings():
        text = str(string)
        if "shld" not in text.lower():
            continue
        relative = string.ea - DGROUP_EA
        needle = struct.pack("<H", relative)
        occurrences = []
        for offset in range(len(load)):
            if load.startswith(needle, offset):
                ea = 0x10000 + offset
                function = ida_funcs.get_func(ea)
                if function:
                    starts.add(function.start_ea)
                occurrences.append({"ea": f"0x{ea:x}", "function": desc(function),
                                    "instruction": insn(ea)})
        strings.append({"string_ea": f"0x{string.ea:x}", "text": text,
                        "dgroup_relative_offset": relative, "raw_occurrences": occurrences})
    short_ea = 0x2E4DE
    short_bytes = ida_bytes.get_bytes(short_ea, 5)
    if short_bytes != b"shld\x00":
        raise RuntimeError(f"unexpected short SHLD bytes: {short_bytes!r}")
    relative = short_ea - DGROUP_EA
    needle = struct.pack("<H", relative)
    occurrences = []
    for offset in range(len(load)):
        if load.startswith(needle, offset):
            ea = 0x10000 + offset
            function = ida_funcs.get_func(ea)
            item_head = ida_bytes.get_item_head(ea)
            if function:
                starts.add(function.start_ea)
            occurrences.append({"ea": f"0x{ea:x}", "function": desc(function),
                                "is_instruction_head": ida_bytes.get_item_head(ea) == ea,
                                "item_head_ea": f"0x{item_head:x}",
                                "item_head_instruction": insn(item_head)})
    strings.append({"string_ea": f"0x{short_ea:x}", "text": "shld",
                    "dgroup_relative_offset": relative, "raw_occurrences": occurrences,
                    "note": "four-character base name omitted by the default IDA Strings minimum length"})
    starts.update((0x16FA6, 0x1A4FE, 0x1BD17, 0x1C907))
    geometry_bytes = ida_bytes.get_bytes(DGROUP_EA + 0x43A, 16)
    if geometry_bytes is None or len(geometry_bytes) != 16:
        raise RuntimeError("cannot read SHLDPICS geometry table")
    if not strings:
        raise RuntimeError("no SHLD strings found")
    operand_consumers = [item for item in occurrences
                         if item["item_head_instruction"]["disassembly"] == "mov     ax, 60Eh"]
    if [(item["ea"], item["item_head_ea"]) for item in operand_consumers] != [("0x1929f", "0x1929e")]:
        raise RuntimeError(f"unexpected short SHLD operand consumers: {operand_consumers}")
    private = {"schema": "kol-ida-koltown-shldpics-private-v1",
               "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
               "input_sha256": digest,
               "address_space": "IDA linear effective address for reconstructed unpacked MZ",
               "strings": strings,
               "shld_geometry_table": {"ea": f"0x{DGROUP_EA + 0x43A:x}",
                                       "bytes": geometry_bytes.hex()},
               "archive_contract": {
                   "indexed_name": "SHLDPICS.0", "fixed_record_name": "SHLDPICS",
                   "names_are_distinct_formats": True, "group_count": 4,
                   "pairing": "color index 2*i, AND-mask index 2*i+1",
                   "dimensions": [{"group": i, "width_pixels": geometry_bytes[2*i] * 8,
                                   "height_pixels": geometry_bytes[2*i+1]}
                                  for i in range(4)],
                   "color_format": "4-bit EGA second-stage RLE",
                   "mask_format": "MSB-first 1-bit EGA AND mask",
                   "composition_consumer": {"original_name": "sub_1BD17", "call_ea": "0x17087"},
                   "confidence": "confirmed",
               },
               "contract_instructions": [insn(ea) for ea in EXPECTED_DISASSEMBLY],
               "functions": [export(ea) for ea in sorted(starts)]}
    summary = dict(private)
    summary["schema"] = "kol-ida-koltown-shldpics-summary-v1"
    summary.pop("functions")
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN SHLDPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
