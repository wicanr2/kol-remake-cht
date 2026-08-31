"""IDA Pro 9.4：匯出 KOLBOOK CHARPICS.0 索引項抽取與角色圖組合 consumer。"""

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


EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
ROOT_FUNCTION_EAS = (0x13D7D,)
CALL_SITE_EAS = (
    0x11E43,
    0x11E46,
    0x11E49,
    0x11F2F,
    0x11F33,
    0x11F37,
    0x11F3B,
    0x11F3F,
    0x11F42,
    0x11F50,
    0x11F53,
    0x11F57,
)
MAX_CALL_DEPTH = 4


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


def discover_functions() -> dict[int, int]:
    depths = {ea: 0 for ea in ROOT_FUNCTION_EAS}
    pending = list(ROOT_FUNCTION_EAS)
    while pending:
        start = pending.pop(0)
        depth = depths[start]
        function = ida_funcs.get_func(start)
        if not function or depth >= MAX_CALL_DEPTH:
            continue
        for ea in idautils.FuncItems(function.start_ea):
            for target in idautils.CodeRefsFrom(ea, False):
                callee = ida_funcs.get_func(target)
                if not callee or callee.start_ea == function.start_ea:
                    continue
                if callee.start_ea not in depths:
                    depths[callee.start_ea] = depth + 1
                    pending.append(callee.start_ea)
    return depths


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

    call_site = [instruction(ea) for ea in CALL_SITE_EAS]
    expected_calls = {
        "0x11e43": "mov     ax, word_20C2A",
        "0x11e46": "add     ax, 400h",
        "0x11e49": "mov     word_1A472, ax",
        "0x11f2f": "push    word_1A472",
        "0x11f33": "push    word_1A470",
        "0x11f37": "push    word_19B52",
        "0x11f3b": "push    word_19B52",
        "0x11f3f": "push    [bp+var_E]",
        "0x11f42": "call    sub_13D7D",
        "0x11f50": "mov     ax, word_20C2A",
        "0x11f53": "add     ax, 400h",
        "0x11f57": "call    sub_15335",
    }
    if any(item["disassembly"] != expected_calls[item["ea"]] for item in call_site):
        raise RuntimeError("CHARPICS.0 call-site contract changed")

    functions = []
    for start, depth in sorted(discover_functions().items()):
        function = ida_funcs.get_func(start)
        if not function:
            raise RuntimeError(f"missing function at 0x{start:x}")
        functions.append(
            {
                "original_name": idc.get_func_name(start),
                "start_ea": f"0x{start:x}",
                "end_ea": f"0x{function.end_ea:x}",
                "discovery_depth": depth,
                "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
                "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
            }
        )

    common = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for loaded KOLBOOK MZ",
        "root_original_functions": [idc.get_func_name(ea) for ea in ROOT_FUNCTION_EAS],
        "max_call_depth": MAX_CALL_DEPTH,
    }
    private = dict(common)
    private["schema"] = "kol-ida-kolbook-charpics-archive-private-v1"
    private["call_site"] = call_site
    private["functions"] = functions
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")

    summary = dict(common)
    summary["schema"] = "kol-ida-kolbook-charpics-archive-summary-v1"
    summary["archive_consumer_contract"] = {
        "status": "confirmed",
        "indexed_archive_name": "CHARPICS.0",
        "index_entry_offset_formula": "20 + 2 * image_index",
        "index_entry_formula_instruction_range": "0x13d8f..0x13d97",
        "payload_decoder_original_function": "sub_139A8",
        "payload_decoder_call_ea": "0x13e19",
        "source_segment": "word_19B52",
        "destination_segment": "word_20C2A + 0x400 paragraphs",
        "destination_offset": 0,
        "subsequent_display_original_function": "sub_15335",
        "subsequent_display_call_ea": "0x11f57",
        "limitation": "archive count, 40x38 geometry and trailing-byte exclusion require fixed-file and downstream-consumer evidence",
    }
    summary["call_site"] = [
        {"ea": item["ea"], "disassembly": item["disassembly"]} for item in call_site
    ]
    summary["functions"] = [
        {
            "original_name": item["original_name"],
            "start_ea": item["start_ea"],
            "end_ea": item["end_ea"],
            "discovery_depth": item["discovery_depth"],
            "instruction_count": len(item["instructions"]),
            "caller_count": len(item["callers"]),
        }
        for item in functions
    ]
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK CHARPICS archive probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
