"""IDA Pro 9.4：探索 KOLTOWN AWD1PICS consumer 與像素契約。"""

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
TARGET_EAS = (0x168F8, 0x1A4FE, 0x1C907, 0x19E9D)
EXPECTED_DISASSEMBLY = {
    0x16839: "mov     [bp+var_4], 0", 0x16851: "mov     [bp+var_E], 0",
    0x1685C: "mov     dx, [bp+var_4]", 0x1685F: "shl     dx, 1",
    0x16861: "shl     dx, 1", 0x16863: "add     dx, ax",
    0x168BA: "add     [bp+var_E], 1", 0x168BE: "cmp     [bp+var_E], 4",
    0x168C7: "add     [bp+var_4], 1", 0x168CB: "cmp     [bp+var_4], 6",
    0x168A2: "push    [bp+var_2]", 0x168A5: "push    [bp+var_6]",
    0x168A8: "call    sub_168F8", 0x168D9: "add     [bp+var_2], 1",
    0x168DD: "cmp     [bp+var_2], 5",
    0x1690E: "mov     ax, [bp+arg_2]", 0x16911: "add     ax, 1",
    0x16925: "mov     ax, [bp+arg_4]", 0x16931: "mov     ax, [bp+arg_4]",
    0x16934: "add     ax, 3", 0x1692B: "mov     ax, [bp+arg_6]",
    0x1693A: "mov     ax, [bp+arg_6]", 0x1693D: "add     ax, 15h",
    0x16971: "mov     ax, 2D50h", 0x16975: "push    [bp+arg_0]",
    0x16978: "call    sub_10F9C",
}


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
    if function is None:
        raise RuntimeError(f"missing function at 0x{start:x}")
    return {**describe(function), "instructions": [instruction(ea) for ea in idautils.FuncItems(start)]}


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    for ea, expected in EXPECTED_DISASSEMBLY.items():
        actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        if actual != expected:
            raise RuntimeError(f"unexpected instruction at 0x{ea:x}: {actual!r}")
    starts = set(TARGET_EAS)
    target = ida_funcs.get_func(0x168F8)
    callers = []
    for call_ea in sorted(idautils.CodeRefsTo(target.start_ea, False)):
        caller = ida_funcs.get_func(call_ea)
        if caller:
            starts.add(caller.start_ea)
            items = list(idautils.FuncItems(caller.start_ea))
            index = items.index(call_ea)
            preceding = [instruction(ea) for ea in items[max(0, index - 16):index]]
        else:
            preceding = []
        callers.append({"call_ea": f"0x{call_ea:x}", "caller": describe(caller),
                        "preceding_instructions": preceding})
    target_calls = []
    for ea in idautils.FuncItems(target.start_ea):
        refs = list(idautils.CodeRefsFrom(ea, False))
        if refs:
            target_calls.append({"instruction": instruction(ea),
                                 "targets": [describe(ida_funcs.get_func(ref)) for ref in refs]})
            for ref in refs:
                function = ida_funcs.get_func(ref)
                if function:
                    starts.add(function.start_ea)
    if [item["call_ea"] for item in callers] != ["0x168a8"]:
        raise RuntimeError("unexpected AWD consumer caller set")
    contract_eas = tuple(EXPECTED_DISASSEMBLY)
    summary = {"schema": "kol-ida-koltown-awdpics-summary-v1",
               "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
               "input_sha256": digest,
               "address_space": "IDA linear effective address for reconstructed unpacked MZ",
               "consumer": describe(target), "caller_count": len(callers),
               "callers": callers, "consumer_direct_calls": target_calls,
               "archive_contract": {"archive_family": "AWD%dPICS", "bundled_archive": "AWD1PICS",
                                    "image_indices": list(range(24)), "width_pixels": 24,
                                    "height_pixels": 21, "pixel_format": "4-bit EGA direct-blit",
                                    "transparency": "none", "confidence": "confirmed"},
               "index_derivation": "outer 0..5 and inner 0..3 produce 4*outer+inner = 0..23",
               "filename_number_derivation": "caller slot 0..4 becomes AWD1..AWD5",
               "contract_instructions": [instruction(ea) for ea in contract_eas],
               "evidence_note": "Original names, linear EAs, bytes and disassembly are retained; no semantic rename is evidence."}
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-awdpics-private-v1"
    private["functions"] = [export_function(ea) for ea in sorted(starts)]
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN AWD1PICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
