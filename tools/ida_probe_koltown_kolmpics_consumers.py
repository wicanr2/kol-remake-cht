"""IDA Pro 9.4：盤點 KOLMPICS 預載 buffer 與 indexed decoder 的所有 consumer。"""

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
INDEXED_DECODER_EA = 0x1A429
KOLMPICS_BUFFER_EA = 0x39820


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


def function_record(start_ea: int) -> dict:
    function = ida_funcs.get_func(start_ea)
    if function is None or function.start_ea != start_ea:
        raise RuntimeError(f"missing function at {start_ea:#x}")
    return {
        "original_name": idc.get_func_name(start_ea),
        "start_ea": f"0x{start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start_ea, False)],
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start_ea)],
    }


def containing_function_eas(refs: list[int]) -> list[int]:
    result = set()
    for ref in refs:
        function = ida_funcs.get_func(ref)
        if function is not None:
            result.add(function.start_ea)
    return sorted(result)


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: script.py PRIVATE.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    input_path = Path(idc.ARGV[2])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    decoder_call_refs = sorted(idautils.CodeRefsTo(INDEXED_DECODER_EA, False))
    buffer_data_refs = sorted(idautils.DataRefsTo(KOLMPICS_BUFFER_EA))
    function_eas = sorted(set(
        containing_function_eas(decoder_call_refs)
        + containing_function_eas(buffer_data_refs)
        + [INDEXED_DECODER_EA]
    ))
    output = {
        "schema": "kol-ida-koltown-kolmpics-consumers-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "indexed_decoder_original_name": idc.get_func_name(INDEXED_DECODER_EA),
        "indexed_decoder_start_ea": f"0x{INDEXED_DECODER_EA:x}",
        "indexed_decoder_call_refs": [f"0x{x:x}" for x in decoder_call_refs],
        "kolmpics_buffer_original_name": idc.get_name(KOLMPICS_BUFFER_EA),
        "kolmpics_buffer_ea": f"0x{KOLMPICS_BUFFER_EA:x}",
        "kolmpics_buffer_data_refs": [f"0x{x:x}" for x in buffer_data_refs],
        "functions": [function_record(ea) for ea in function_eas],
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(output, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN KOLMPICS consumer probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
