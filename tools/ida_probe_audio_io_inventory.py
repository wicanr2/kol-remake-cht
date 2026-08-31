"""IDA Pro 9.4：窮舉 executable 的 x86 port I/O，保守標記音訊候選。"""

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


EXPECTED_SHA256 = {
    "KOL.EXE": "4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71",
    "KOLBOOK.EXE": "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832",
    "KOLTOWN-unpacked.EXE": "01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2",
    "MAKEMODE.EXE": "3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56",
    "TANDYPIN.EXE": "036c39f330a4fe09bc2729756816b869f1154515bf4befba190e767d4e3cdce3",
    "TITLE.EXE": "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6",
}
AUDIO_PORTS = {0x42, 0x43, 0x61, 0xC0}
IO_MNEMONICS = {"in", "out", "ins", "insb", "insw", "outs", "outsb", "outsw"}


def disassembly(ea: int) -> str:
    return ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disassembly(ea),
    }


def immediate_port(ea: int) -> int | None:
    mnemonic = idc.print_insn_mnem(ea).lower()
    index = 1 if mnemonic == "in" else 0
    operand = idc.print_operand(ea, index).lower().replace(" ", "")
    if operand in {"dx", "edx"}:
        return None
    value = idc.get_operand_value(ea, index)
    return value if 0 <= value <= 0xFFFF else None


def nearest_dx_setup(ea: int, function: ida_funcs.func_t | None) -> dict | None:
    if function is None:
        return None
    cursor = ida_bytes.prev_head(ea, function.start_ea)
    for _ in range(32):
        if cursor < function.start_ea:
            break
        destination = idc.print_operand(cursor, 0).lower().replace(" ", "")
        mnemonic = idc.print_insn_mnem(cursor).lower()
        if destination in {"dx", "edx"} and mnemonic in {"mov", "xor", "pop", "lea"}:
            row = instruction(cursor)
            row["operand_value"] = idc.get_operand_value(cursor, 1) if mnemonic in {"mov", "lea"} else None
            return row
        cursor = ida_bytes.prev_head(cursor, function.start_ea)
    return None


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if EXPECTED_SHA256.get(input_path.name) != digest:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    hits = []
    for ea in idautils.Heads():
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        mnemonic = idc.print_insn_mnem(ea).lower()
        if mnemonic not in IO_MNEMONICS:
            continue
        function = ida_funcs.get_func(ea)
        port = immediate_port(ea)
        setup = nearest_dx_setup(ea, function) if port is None else None
        setup_value = setup.get("operand_value") if setup else None
        candidate_reason = None
        if port in AUDIO_PORTS:
            candidate_reason = f"immediate port 0x{port:x}"
        elif setup_value in AUDIO_PORTS:
            candidate_reason = f"nearest direct DX setup 0x{setup_value:x}"
        hits.append({
            **instruction(ea),
            "original_function": idc.get_func_name(function.start_ea) if function else "",
            "function_start_ea": f"0x{function.start_ea:x}" if function else None,
            "immediate_port": f"0x{port:x}" if port is not None else None,
            "nearest_dx_setup": setup,
            "audio_candidate_reason": candidate_reason,
        })

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for the named loaded MZ image",
        "io_instruction_count": len(hits),
        "audio_candidate_count": sum(row["audio_candidate_reason"] is not None for row in hits),
    }
    private = {"schema": "kol-ida-audio-io-private-v1", **shared, "hits": hits}
    safe = {
        "schema": "kol-ida-audio-io-summary-v1",
        **shared,
        "audio_candidates": [
            {
                "ea": row["ea"],
                "bytes": row["bytes"],
                "disassembly": row["disassembly"],
                "original_function": row["original_function"],
                "function_start_ea": row["function_start_ea"],
                "reason": row["audio_candidate_reason"],
            }
            for row in hits if row["audio_candidate_reason"] is not None
        ],
        "method_limit": (
            "窮舉 IDA 已定義 code 的 x86 IN/OUT；DX 僅以同函式向前 32 個 instruction 的最近直接設定"
            "作候選，不把它當完整資料流證明。無候選不單獨證明沒有間接 port route。"
        ),
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"audio I/O inventory probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
