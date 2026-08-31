"""IDA Pro 9.4：驗證 KOLTOWN WEAPPICS.0 武器彩圖／AND-mask 契約。"""

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
WEAP_TOKEN_EA = DGROUP_EA + 0x0609
FORMAT_STRING_EA = DGROUP_EA + 0x17EA
SELECTOR_BASE_EA = DGROUP_EA + 0x05F2
GEOMETRY_EA = DGROUP_EA + 0x0402
POSITION_EA = DGROUP_EA + 0x0472
EXPECTED = {
    0x19303: "mov     al, [bx-4478h]",
    0x1930A: "mov     cx, 20h ; ' '",
    0x19322: "and     ax, 0Fh",
    0x19328: "cmp     [bp+var_C], 5",
    0x19331: "cmp     [bp+var_E], 2",
    0x1933A: "cmp     [bp+var_E], 3",
    0x19343: "xor     [bp+var_E], 1",
    0x19347: "mov     ax, [bp+var_E]",
    0x1934A: "shl     ax, 1",
    0x19352: "mov     al, [bx+5F2h]",
    0x1935C: "add     ax, cx",
    0x19361: "mov     cx, 2",
    0x19369: "div     cx",
    0x19374: "mov     ax, 38h ; '8'",
    0x19377: "imul    word_3AC60",
    0x1937B: "mov     cx, 472h",
    0x19380: "mov     ax, [bp+var_A]",
    0x19383: "shl     ax, 1",
    0x193B4: "mov     bx, [bp+var_A]",
    0x193B7: "shl     bx, 1",
    0x193B9: "mov     al, [bx+402h]",
    0x193CF: "mov     cx, 402h",
    0x193F7: "mov     ax, 17EAh",
    0x1940B: "mov     ax, [bp+var_8]",
    0x1940E: "add     ax, 1",
    0x19412: "push    [bp+var_8]",
    0x19415: "call    sub_16FA6",
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
    base_name_bytes = ida_bytes.get_bytes(WEAP_TOKEN_EA, 5)
    format_bytes = ida_bytes.get_bytes(FORMAT_STRING_EA, 5)
    if base_name_bytes != b"weap\x00":
        rendered = "missing" if base_name_bytes is None else base_name_bytes.hex()
        raise RuntimeError(f"unexpected weap token bytes: {rendered}")
    if format_bytes != b"%s%s\x00":
        rendered = "missing" if format_bytes is None else format_bytes.hex()
        raise RuntimeError(f"unexpected filename format bytes: {rendered}")
    for ea, expected in EXPECTED.items():
        actual = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
        if actual != expected:
            raise RuntimeError(f"unexpected instruction at 0x{ea:x}: {actual!r}")

    selector_bases = ida_bytes.get_bytes(SELECTOR_BASE_EA, 6)
    geometry = ida_bytes.get_bytes(GEOMETRY_EA, 56)
    positions = ida_bytes.get_bytes(POSITION_EA, 6 * 56)
    if selector_bases is None or len(selector_bases) != 6:
        raise RuntimeError("missing WEAPPICS selector-base table")
    if geometry is None or len(geometry) != 56 or any(value == 0 for value in geometry):
        raise RuntimeError("invalid WEAPPICS geometry table")
    if positions is None or len(positions) != 6 * 56:
        raise RuntimeError("missing WEAPPICS position table")

    dimensions = [
        {
            "group": group,
            "color_image_index": 2 * group,
            "mask_image_index": 2 * group + 1,
            "width_bytes": geometry[2 * group],
            "width_pixels": geometry[2 * group] * 8,
            "height_pixels": geometry[2 * group + 1],
        }
        for group in range(28)
    ]
    position_records = [
        {
            "renderer_class": renderer_class,
            "group": group,
            "x": positions[renderer_class * 56 + 2 * group],
            "y": positions[renderer_class * 56 + 2 * group + 1],
        }
        for renderer_class in range(6)
        for group in range(28)
    ]
    summary = {
        "schema": "kol-ida-koltown-weappics-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "base_name": {
            "ea": f"0x{WEAP_TOKEN_EA:x}",
            "dgroup_offset": "0x0609",
            "bytes": "7765617000",
            "text": "weap",
        },
        "filename_format": {
            "ea": f"0x{FORMAT_STRING_EA:x}",
            "dgroup_offset": "0x17ea",
            "bytes": "2573257300",
            "text": "%s%s",
            "expression": "weap token concatenated with the active renderer suffix",
        },
        "consumer": {"original_name": "sub_18F3A", "start_ea": "0x18f3a"},
        "paired_loader": {"original_name": "sub_16FA6", "start_ea": "0x16fa6"},
        "selector_contract": {
            "family_expression": "unsigned equipment byte at character offset 0x3d / 0x20",
            "variant_expression": "unsigned equipment byte at character offset 0x3e & 0x0f",
            "special_case": "when family is 5 and variant is 2 or 3, variant ^= 1",
            "selector_expression": "selector_base[family] + 2*variant",
            "group_expression": "selector / 2",
            "selector_base_table": {
                "ea": f"0x{SELECTOR_BASE_EA:x}",
                "dgroup_offset": "0x05f2",
                "bytes": selector_bases.hex(),
                "values": list(selector_bases),
            },
            "confidence": "confirmed",
        },
        "archive_contract": {
            "indexed_name": "WEAPPICS.0",
            "payload_count": 56,
            "group_count": 28,
            "pairing": "color index selector, AND-mask index selector+1; selector is even",
            "color_format": "4-bit EGA second-stage RLE",
            "mask_format": "MSB-first 1-bit EGA AND mask",
            "mask_semantics": "bit 1 preserves background; PNG alpha is 1-mask",
            "dimensions": dimensions,
            "confidence": "confirmed",
        },
        "geometry_table": {"ea": f"0x{GEOMETRY_EA:x}", "bytes": geometry.hex()},
        "position_table": {
            "ea": f"0x{POSITION_EA:x}",
            "renderer_class_count": 6,
            "bytes_per_renderer_class": 56,
            "records": position_records,
        },
        "contract_instructions": [instruction(ea) for ea in EXPECTED],
        "evidence_note": "原始函式名、線性 EA、bytes 與運算元均保留；family、variant、renderer_class 僅為導覽標籤，未推定遊戲內分類名稱。",
    }
    private = dict(summary)
    private["schema"] = "kol-ida-koltown-weappics-private-v1"
    private["functions"] = [export(ea) for ea in (0x16FA6, 0x18F3A, 0x1A4FE, 0x1BD17)]
    private_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN WEAPPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
