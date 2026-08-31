"""IDA Pro 9.4：匯出 KOL／KOLTOWN BIOS video interrupt 與直接 port I/O 證據。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_bytes
import ida_auto
import ida_funcs
import ida_lines
import idaapi
import idautils
import idc


EXPECTED_SHA256 = {
    "KOL.EXE": "4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71",
    "KOLTOWN-unpacked.EXE": "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2",
    "MAKEMODE.EXE": "3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56",
    "TITLE.EXE": "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6",
}


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    raw = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:x}",
        "bytes": raw.hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    summary_path = Path(idc.ARGV[2])
    input_path = Path(idc.ARGV[3])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    expected = EXPECTED_SHA256.get(input_path.name)
    if expected is None or digest != expected:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    all_code = [ea for ea in idautils.Heads() if ida_bytes.is_code(ida_bytes.get_flags(ea))]
    ordinal = {ea: index for index, ea in enumerate(all_code)}
    hits = []
    for ea in all_code:
        mnemonic = idc.print_insn_mnem(ea).lower()
        operand = idc.print_operand(ea, 0).lower().replace(" ", "")
        category = None
        if mnemonic == "int" and operand in {"10h", "0x10"}:
            category = "bios_int_10h"
        elif mnemonic in {"in", "out", "ins", "insb", "insw", "outs", "outsb", "outsw"}:
            category = "direct_port_io"
        if category is None:
            continue
        function = ida_funcs.get_func(ea)
        index = ordinal[ea]
        window = all_code[max(0, index - 12) : min(len(all_code), index + 13)]
        hits.append(
            {
                "ea": f"0x{ea:x}",
                "category": category,
                "original_function": idc.get_func_name(function.start_ea) if function else "",
                "function_start_ea": f"0x{function.start_ea:x}" if function else None,
                "window": [instruction(item) for item in window],
            }
        )

    summary = {
        "schema": "kol-ida-video-io-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for the named loaded MZ image",
        "hits": [
            {
                "ea": hit["ea"],
                "category": hit["category"],
                "original_function": hit["original_function"],
                "function_start_ea": hit["function_start_ea"],
            }
            for hit in hits
        ],
    }
    private = dict(summary)
    private["schema"] = "kol-ida-video-io-private-v1"
    private["hits"] = hits
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOL video I/O probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
