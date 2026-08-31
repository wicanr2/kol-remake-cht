"""IDA Pro 9.4：驗證 KOLTOWN HEADPICS／TORPICS／LEGSPICS 配對契約。"""

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
TOKEN_TABLE_EA = DGROUP_EA + 0x05FA
FORMAT_STRING_EA = DGROUP_EA + 0x17DE
GEOMETRY_EA = DGROUP_EA + 0x017A
POSITION_EA = DGROUP_EA + 0x02BE
COMPONENTS = (
    {"component": 0, "token": "head", "archive": "HEADPICS.0"},
    {"component": 1, "token": "tor", "archive": "TORPICS.0"},
    {"component": 2, "token": "legs", "archive": "LEGSPICS.0"},
)
EXPECTED = {
    0x19029: "mov     ax, 9",
    0x1902C: "imul    [bp+var_A]",
    0x1903A: "mov     [bp+var_10], 2",
    0x19042: "mov     ax, [bp+var_10]",
    0x1904B: "shl     ax, 1",
    0x19051: "add     bx, 100h",
    0x19057: "mov     al, [bx-4478h]",
    0x1905E: "cmp     ax, 0",
    0x19075: "add     bx, 10Eh",
    0x1907B: "mov     al, [bx-4478h]",
    0x19082: "and     ax, 0Fh",
    0x19085: "add     ax, [bp+var_4]",
    0x19091: "mov     ax, 6",
    0x19094: "imul    [bp+var_6]",
    0x19097: "mov     cx, 2BEh",
    0x1909C: "mov     ax, [bp+var_10]",
    0x190D5: "mov     cx, 17Ah",
    0x19123: "mov     ax, 5",
    0x19126: "imul    [bp+var_10]",
    0x19129: "mov     cx, 5FAh",
    0x1912F: "mov     ax, 17DEh",
    0x19143: "mov     ax, [bp+var_6]",
    0x19146: "shl     ax, 1",
    0x19148: "add     ax, 1",
    0x1914C: "mov     ax, [bp+var_6]",
    0x1914F: "shl     ax, 1",
    0x19152: "call    sub_16FA6",
    0x1915A: "sub     [bp+var_10], 1",
    0x1915E: "cmp     [bp+var_10], 0",
}


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
    }


def export(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None or function.start_ea != start:
        raise RuntimeError(f"missing function at 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{start:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()
    token_bytes = ida_bytes.get_bytes(TOKEN_TABLE_EA, 15)
    format_bytes = ida_bytes.get_bytes(FORMAT_STRING_EA, 5)
    if token_bytes != b"head\x00tor\x00\x00legs\x00":
        rendered = "missing" if token_bytes is None else token_bytes.hex()
        raise RuntimeError(f"unexpected body token table: {rendered}")
    if format_bytes != b"%s%s\x00":
        rendered = "missing" if format_bytes is None else format_bytes.hex()
        raise RuntimeError(f"unexpected body filename format: {rendered}")
    for ea, expected in EXPECTED.items():
        actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        if actual != expected:
            raise RuntimeError(f"unexpected instruction at 0x{ea:x}: {actual!r}")

    geometry = ida_bytes.get_bytes(GEOMETRY_EA, 54 * 6)
    positions = ida_bytes.get_bytes(POSITION_EA, 54 * 6)
    if geometry is None or len(geometry) != 54 * 6 or any(value == 0 for value in geometry):
        raise RuntimeError("invalid body geometry table")
    if positions is None or len(positions) != 54 * 6:
        raise RuntimeError("missing body position table")

    archive_contracts = []
    for component in COMPONENTS:
        index = component["component"]
        dimensions = []
        for group in range(54):
            record = group * 6 + index * 2
            dimensions.append(
                {
                    "group": group,
                    "renderer_class": group // 9,
                    "variant": group % 9,
                    "color_image_index": 2 * group,
                    "mask_image_index": 2 * group + 1,
                    "width_bytes": geometry[record],
                    "width_pixels": geometry[record] * 8,
                    "height_pixels": geometry[record + 1],
                    "x": positions[record],
                    "y": positions[record + 1],
                }
            )
        archive_contracts.append(
            {
                **component,
                "payload_count": 108,
                "group_count": 54,
                "pairing": "color index 2*group, AND-mask index 2*group+1",
                "color_format": "4-bit EGA second-stage RLE",
                "mask_format": "MSB-first 1-bit EGA AND mask",
                "mask_semantics": "bit 1 preserves background; PNG alpha is 1-mask",
                "dimensions": dimensions,
                "confidence": "confirmed",
            }
        )

    summary = {
        "schema": "kol-ida-koltown-bodypics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "consumer": {"original_name": "sub_18F3A", "start_ea": "0x18f3a"},
        "paired_loader": {"original_name": "sub_16FA6", "start_ea": "0x16fa6"},
        "filename_contract": {
            "token_table_ea": f"0x{TOKEN_TABLE_EA:x}",
            "token_table_dgroup_offset": "0x05fa",
            "token_table_bytes": token_bytes.hex(),
            "format_ea": f"0x{FORMAT_STRING_EA:x}",
            "format_dgroup_offset": "0x17de",
            "format_bytes": format_bytes.hex(),
            "format": "%s%s",
            "expression": "component token concatenated with the active renderer suffix",
        },
        "selector_contract": {
            "component_iteration": "2, 1, 0",
            "presence_expression": "unsigned character byte at offset 0x100 + 16*component != 0",
            "variant_expression": "unsigned character byte at offset 0x10e + 16*component & 0x0f",
            "group_expression": "9*renderer_class + variant",
            "archive_color_expression": "2*group",
            "archive_mask_expression": "2*group+1",
            "observed_archive_group_domain": "0..53, represented as six renderer classes by nine table slots",
            "evidence_limit": "the instruction mask permits 0..15; valid stored variant values require separate character/item-data evidence",
            "confidence": "confirmed except valid stored variant domain",
        },
        "geometry_table": {
            "ea": f"0x{GEOMETRY_EA:x}",
            "dgroup_offset": "0x017a",
            "bytes": geometry.hex(),
            "record_layout": "54 groups * 3 components * (width_bytes,height_pixels)",
        },
        "position_table": {
            "ea": f"0x{POSITION_EA:x}",
            "dgroup_offset": "0x02be",
            "bytes": positions.hex(),
            "record_layout": "54 groups * 3 components * (x,y)",
        },
        "archive_contracts": archive_contracts,
        "contract_instructions": [instruction(ea) for ea in EXPECTED],
        "evidence_note": "原始函式名、線性 EA、bytes 與運算元均保留；component、renderer_class、variant 只作資料流導覽。",
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-bodypics-private-v1"
    private["functions"] = [export(ea) for ea in (0x16FA6, 0x18F3A, 0x1A4FE, 0x1BD17)]
    private_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN body PICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
