"""IDA Pro 9.4：匯出 KOLTOWN 圖像載入／bitstream pipeline 的原始函式證據。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_bytes
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2"
TARGETS = [
    0x10F9C,
    0x16FA6,
    0x19D98,
    0x19E12,
    0x19E9D,
    0x1A04E,
    0x1A26B,
    0x1A429,
    0x1A4E4,
    0x1A4FE,
    0x1A866,
    0x1BD17,
    0x2D3FA,
]


def insn(ea: int) -> dict:
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

    functions = []
    for requested_ea in TARGETS:
        function = ida_funcs.get_func(requested_ea)
        if function is None:
            functions.append({"requested_ea": f"0x{requested_ea:x}", "missing": True})
            continue
        items = list(idautils.FuncItems(function.start_ea))
        functions.append({
            "requested_ea": f"0x{requested_ea:x}",
            "original_name": idc.get_func_name(function.start_ea),
            "start_ea": f"0x{function.start_ea:x}",
            "end_ea": f"0x{function.end_ea:x}",
            "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(function.start_ea, False)],
            "instructions": [insn(ea) for ea in items],
        })

    summary = {
        "schema": "kol-ida-koltown-image-pipeline-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "functions": [
            {
                "requested_ea": f["requested_ea"],
                "original_name": f.get("original_name"),
                "start_ea": f.get("start_ea"),
                "end_ea": f.get("end_ea"),
                "instruction_count": len(f.get("instructions", [])),
                "caller_count": len(f.get("callers", [])),
            }
            for f in functions
        ],
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-image-pipeline-private-v1"
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN image pipeline probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
