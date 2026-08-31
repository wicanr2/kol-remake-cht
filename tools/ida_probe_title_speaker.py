"""IDA Pro 9.4：固定 TITLE 的 PC speaker／PIT 函式與 caller。"""

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


EXPECTED_SHA256 = "8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6"
ROOTS = [0x10F33, 0x10F48, 0x10F5D, 0x10F75, 0x11904, 0x11953]
EXPECTED_INSTRUCTIONS = {
    0x10F3C: "in      al, 61h",
    0x10F3E: "or      al, 3",
    0x10F40: "out     61h, al",
    0x10F51: "in      al, 61h",
    0x10F53: "and     al, 0FCh",
    0x10F55: "out     61h, al",
    0x10F66: "mov     ax, [bp+arg_0]",
    0x10F69: "out     42h, al",
    0x10F6D: "out     42h, al",
    0x10F86: "call    sub_10F33",
    0x10F9A: "mov     ax, 2710h",
    0x10F9E: "call    sub_11953",
    0x10FAD: "call    sub_10F5D",
    0x10FE1: "cmp     [bp+var_2], 64h",
    0x10FEA: "call    sub_10F48",
    0x11013: "cmp     [bp+var_4], 2",
}


def disassembly(ea: int) -> str:
    text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return text.split(";", 1)[0].rstrip()


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disassembly(ea),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [f"0x{x:x}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None or function.start_ea != start:
        raise RuntimeError(f"missing function at 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in sorted(idautils.CodeRefsTo(start, False))],
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    checks = []
    for ea, expected in EXPECTED_INSTRUCTIONS.items():
        actual = disassembly(ea)
        if actual != expected:
            raise RuntimeError(f"instruction mismatch at 0x{ea:x}: {actual!r} != {expected!r}")
        checks.append(instruction(ea))

    starts = set(ROOTS)
    caller_sites: dict[str, list[str]] = {}
    for root in ROOTS:
        refs = sorted(idautils.CodeRefsTo(root, False))
        caller_sites[f"0x{root:x}"] = [f"0x{x:x}" for x in refs]
        for ref in refs:
            function = ida_funcs.get_func(ref)
            if function is None:
                raise RuntimeError(f"caller outside function at 0x{ref:x}")
            starts.add(function.start_ea)

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for TITLE loaded MZ",
        "root_original_functions": [idc.get_func_name(ea) for ea in ROOTS],
        "root_start_eas": [f"0x{x:x}" for x in ROOTS],
        "caller_sites": caller_sites,
    }
    private = {
        "schema": "kol-ida-title-speaker-private-v1",
        **shared,
        "instruction_checks": checks,
        "functions": [function_record(ea) for ea in sorted(starts)],
    }
    safe = {
        "schema": "kol-ida-title-speaker-summary-v1",
        **shared,
        "primitive_contract": {
            "speaker_enable_original_function": "sub_10F33",
            "speaker_disable_original_function": "sub_10F48",
            "pit_channel_2_divisor_original_function": "sub_10F5D",
            "status": "confirmed",
        },
        "cue_contract": {
            "navigation_label": "title-random-tone-burst",
            "original_function": "sub_10F75",
            "start_ea": "0x10f75",
            "outer_burst_count": 2,
            "divisor_updates_per_burst": 100,
            "random_helper_original_function": "sub_11953",
            "random_helper_contract": "rand() % N + 1",
            "busy_wait_helper_original_function": "sub_11904",
            "speaker_order": ["sub_10F33", "sub_10F5D repeated", "sub_10F48"],
            "caller_sites": caller_sites["0x10f75"],
            "status": "confirmed",
            "label_is_evidence": False,
        },
        "instruction_checks": [
            {"ea": row["ea"], "bytes": row["bytes"], "disassembly": row["disassembly"]}
            for row in checks
        ],
        "limitation": (
            "已確認 TITLE 呼叫順序與演算法形狀；未宣稱逐週期 PIT／busy-wait 時長、"
            "實機波形或精確 wall-clock parity。後續 OGG 只能標為 hardware-spec approximation。"
        ),
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.parent.mkdir(parents=True, exist_ok=True)
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"TITLE speaker probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
