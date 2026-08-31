"""IDA Pro 9.4：非破壞性匯出 DOS executable 基本指紋。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_entry
import ida_funcs
import ida_ida
import ida_nalt
import ida_segment
import idaapi
import idc


SCHEMA = "kol-ida-fingerprint-v1"


def main() -> None:
    if len(idc.ARGV) != 2:
        raise RuntimeError("usage: ida_export_fingerprint.py OUTPUT.json")
    output = Path(idc.ARGV[1])
    input_path = Path(ida_nalt.get_input_file_path())
    input_bytes = input_path.read_bytes()

    segments = []
    segment = ida_segment.get_first_seg()
    while segment is not None:
        segments.append(
            {
                "name": ida_segment.get_segm_name(segment),
                "class": ida_segment.get_segm_class(segment),
                "start_ea": f"0x{segment.start_ea:x}",
                "end_ea": f"0x{segment.end_ea:x}",
                "size": segment.end_ea - segment.start_ea,
                "bitness": segment.bitness,
                "permissions": segment.perm,
            }
        )
        segment = ida_segment.get_next_seg(segment.start_ea)

    entries = []
    for index in range(ida_entry.get_entry_qty()):
        ordinal = ida_entry.get_entry_ordinal(index)
        ea = ida_entry.get_entry(ordinal)
        entries.append(
            {
                "ordinal": ordinal,
                "ea": f"0x{ea:x}",
                "name": ida_entry.get_entry_name(ordinal),
            }
        )

    result = {
        "schema": SCHEMA,
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_bytes": len(input_bytes),
        "input_sha256": hashlib.sha256(input_bytes).hexdigest(),
        "loader_file_type": idaapi.get_file_type_name(),
        "file_type_code": ida_ida.inf_get_filetype(),
        "processor": ida_ida.inf_get_procname(),
        "address_space": "IDA linear effective address",
        "min_ea": f"0x{ida_ida.inf_get_min_ea():x}",
        "max_ea": f"0x{ida_ida.inf_get_max_ea():x}",
        "start_ip": f"0x{ida_ida.inf_get_start_ip():x}",
        "start_cs": f"0x{ida_ida.inf_get_start_cs():x}",
        "function_count": ida_funcs.get_func_qty(),
        "segments": segments,
        "entries": entries,
    }
    output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOL fingerprint export failed: {exc}", file=sys.stderr)
    idc.qexit(1)
