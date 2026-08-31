"""IDA Pro 9.4：匯出 entry 原始 bytes 與有限線性解碼，不推測語意。"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path

import ida_bytes
import ida_entry
import ida_nalt
import ida_ua
import idaapi
import idc


def main() -> None:
    if len(idc.ARGV) != 2:
        raise RuntimeError("usage: ida_probe_entry.py OUTPUT.json")
    output = Path(idc.ARGV[1])
    input_path = Path(ida_nalt.get_input_file_path())
    raw = input_path.read_bytes()
    last, pages = struct.unpack_from("<HH", raw, 2)
    declared = (pages - 1) * 512 + (last or 512)

    ordinal = ida_entry.get_entry_ordinal(0)
    entry_ea = ida_entry.get_entry(ordinal)
    entry_bytes = ida_bytes.get_bytes(entry_ea, 128) or b""
    bytes_before_entry = ida_bytes.get_bytes(entry_ea - 18, 18) or b""
    instructions = []
    ea = entry_ea
    for _ in range(32):
        size = ida_ua.create_insn(ea)
        if size <= 0:
            instructions.append(
                {"ea": f"0x{ea:x}", "bytes": "", "disassembly": None}
            )
            break
        data = ida_bytes.get_bytes(ea, size) or b""
        instructions.append(
            {
                "ea": f"0x{ea:x}",
                "bytes": data.hex(),
                "disassembly": idc.generate_disasm_line(ea, 0),
            }
        )
        ea += size

    result = {
        "schema": "kol-ida-entry-probe-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_bytes": len(raw),
        "input_sha256": hashlib.sha256(raw).hexdigest(),
        "address_space": "IDA linear effective address",
        "entry_ea": f"0x{entry_ea:x}",
        "bytes_before_entry_18": bytes_before_entry.hex(),
        "entry_bytes_128": entry_bytes.hex(),
        "linear_decode": instructions,
        "mz_declared_bytes": declared,
        "trailing_bytes": len(raw) - declared,
        "trailing_hex": raw[declared:].hex(),
    }
    output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOL entry probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
