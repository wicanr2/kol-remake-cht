"""IDA Pro 9.4：驗證 KOLTOWN HANDPICS.0 左右手彩圖／AND-mask 契約。"""

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
DGROUP_EA = 0x2DED0
HAND_STRING_EA = 0x2E4E3
GEOMETRY_EA = DGROUP_EA + 0x5C2
EXPECTED = {
    0x1941D: "mov     ax, word_3AC60", 0x19420: "shl     ax, 1",
    0x19422: "mov     [bp+var_A], ax", 0x19454: "mov     bx, [bp+var_A]",
    0x19457: "shl     bx, 1", 0x19459: "mov     al, [bx+5C2h]",
    0x1946A: "mov     ax, [bp+var_A]", 0x1946D: "shl     ax, 1",
    0x19476: "mov     al, [bx+1]", 0x19493: "mov     ax, 613h",
    0x194AB: "mov     ax, [bp+var_A]", 0x194AE: "shl     ax, 1",
    0x194B0: "add     ax, 1", 0x194B4: "mov     ax, [bp+var_A]",
    0x194B7: "shl     ax, 1", 0x194BA: "call    sub_16FA6",
    0x194C2: "add     [bp+var_A], 1", 0x19556: "mov     ax, 613h",
    0x1956E: "mov     ax, [bp+var_A]", 0x19571: "shl     ax, 1",
    0x19573: "add     ax, 1", 0x19577: "mov     ax, [bp+var_A]",
    0x1957A: "shl     ax, 1", 0x1957D: "call    sub_16FA6",
    0x16FDD: "push    [bp+arg_0]", 0x16FE3: "call    sub_1A4FE",
    0x17065: "push    [bp+arg_2]", 0x1706B: "call    sub_1A4FE",
    0x17087: "call    sub_1BD17",
}


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {"ea": f"0x{ea:x}", "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
            "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
            "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
            "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)]}


def describe(function) -> dict:
    return {"original_name": idc.get_func_name(function.start_ea),
            "start_ea": f"0x{function.start_ea:x}", "end_ea": f"0x{function.end_ea:x}"}


def export(start: int) -> dict:
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
    if ida_bytes.get_bytes(HAND_STRING_EA, 5) != b"hand\x00":
        raise RuntimeError("unexpected hand base-name bytes")
    for ea, expected in EXPECTED.items():
        actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        if actual != expected:
            raise RuntimeError(f"unexpected instruction at 0x{ea:x}: {actual!r}")
    geometry = ida_bytes.get_bytes(GEOMETRY_EA, 24)
    if geometry is None or len(geometry) != 24 or any(geometry[i] == 0 for i in range(24)):
        raise RuntimeError("invalid HANDPICS geometry table")
    dimensions = [{"group": i, "appearance_class": i // 2,
                   "side": "first" if i % 2 == 0 else "second",
                   "width_bytes": geometry[2*i], "width_pixels": geometry[2*i] * 8,
                   "height_pixels": geometry[2*i+1]} for i in range(12)]
    summary = {"schema": "kol-ida-koltown-handpics-summary-v1",
               "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
               "input_sha256": digest,
               "address_space": "IDA linear effective address for reconstructed unpacked MZ",
               "base_name": {"ea": f"0x{HAND_STRING_EA:x}", "dgroup_offset": "0x0613",
                             "bytes": "68616e6400", "text": "hand"},
               "consumer": {"original_name": "sub_18F3A", "start_ea": "0x18f3a"},
               "paired_loader": {"original_name": "sub_16FA6", "start_ea": "0x16fa6"},
               "archive_contract": {"indexed_name": "HANDPICS.0", "group_count": 12,
                                    "appearance_class_count": 6,
                                    "groups_per_appearance_class": 2,
                                    "pairing": "color index 2*group, AND-mask index 2*group+1",
                                    "color_format": "4-bit EGA second-stage RLE",
                                    "mask_format": "MSB-first 1-bit EGA AND mask",
                                    "mask_semantics": "bit 1 preserves background; PNG alpha is 1-mask",
                                    "dimensions": dimensions, "confidence": "confirmed"},
               "geometry_table": {"ea": f"0x{GEOMETRY_EA:x}", "bytes": geometry.hex()},
               "contract_instructions": [instruction(ea) for ea in EXPECTED],
               "evidence_note": "Original names, linear EAs, bytes and operands are retained; first/second are positional labels, not inferred handedness."}
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-handpics-private-v1"
    private["functions"] = [export(ea) for ea in (0x16FA6, 0x18F3A, 0x1A4FE, 0x1BD17)]
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN HANDPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
