"""IDA Pro 9.4：匯出 KOLTOWN KOLMPICS 圖號 28 consumer 證據。"""

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
FUNCTION_EAS = [0x10000, 0x14047, 0x1A429, 0x1BD17, 0x19D98, 0x2D3FA]
DGROUP_EA = 0x2DED0


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


def require_instruction(ea: int, mnemonic: str, operand_fragment: str) -> None:
    if idc.print_insn_mnem(ea).lower() != mnemonic.lower():
        raise RuntimeError(f"unexpected mnemonic at {ea:#x}")
    disassembly = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    if operand_fragment not in disassembly:
        raise RuntimeError(f"unexpected operand at {ea:#x}: {disassembly}")


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path = Path(idc.ARGV[1])
    safe_path = Path(idc.ARGV[2])
    input_path = Path(idc.ARGV[3])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    require_instruction(0x1037A, "call", "sub_2D3FA")
    require_instruction(0x1038E, "call", "sub_19D98")
    require_instruction(0x14150, "push", "word_39820")
    require_instruction(0x14154, "push", "word_39820")
    require_instruction(0x14158, "mov", "1Ch")
    require_instruction(0x1415C, "call", "sub_1A429")
    require_instruction(0x1417B, "call", "sub_1BD17")
    require_instruction(0x1BF4F, "and", "dl, [bx]")
    require_instruction(0x1BF51, "or", "dl, [si]")
    archive_format = ida_bytes.get_strlit_contents(DGROUP_EA + 0x165A, -1, 0)
    if archive_format != b"kolm%s.0":
        raise RuntimeError(f"unexpected KOLMPICS filename format: {archive_format!r}")

    private = {
        "schema": "kol-ida-koltown-kolmpics-index28-private-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "data_windows": [{
            "label": "startup_asset_filename_formats",
            "dgroup_start_offset": "0x164a",
            "dgroup_end_offset_exclusive": "0x1698",
            "ida_start_ea": f"0x{DGROUP_EA + 0x164A:x}",
            "bytes": (ida_bytes.get_bytes(DGROUP_EA + 0x164A, 0x4E) or b"").hex(),
        }],
        "functions": [function_record(ea) for ea in FUNCTION_EAS],
    }
    safe = {
        "schema": "kol-ida-koltown-kolmpics-index28-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for reconstructed unpacked MZ",
        "archive_load": {
            "filename_format": "kolm%s.0",
            "filename_format_dgroup_offset": "0x165a",
            "format_call_ea": "0x1037a",
            "whole_file_reader_original_name": "sub_19D98",
            "whole_file_reader_call_ea": "0x1038e",
            "archive_buffer_original_name": "word_39820",
            "status": "confirmed",
        },
        "consumer": {
            "original_name": "sub_14047",
            "start_ea": "0x14047",
            "archive_buffer_push_eas": ["0x14150", "0x14154"],
            "image_index": 28,
            "image_index_instruction_ea": "0x14158",
            "indexed_decoder_original_name": "sub_1A429",
            "indexed_decoder_call_ea": "0x1415c",
            "blitter_original_name": "sub_1BD17",
            "blitter_call_ea": "0x1417b",
            "ega_composite": {
                "destination_read_ea": "0x1bf4c",
                "and_mask_ea": "0x1bf4f",
                "or_color_ea": "0x1bf51",
                "destination_write_ea": "0x1bf53",
                "formula": "destination = (destination AND mask) OR color",
                "mask_bit_1": "preserve destination (transparent)",
                "mask_bit_0": "clear destination before OR color (opaque)",
                "status": "confirmed",
            },
            "source_rectangle": {
                "x0_byte_column": 0,
                "y0_scanline": 0,
                "x1_byte_column_exclusive": 7,
                "y1_scanline_exclusive": 57,
                "width_bytes": 7,
                "width_pixels": 56,
                "height_pixels": 57,
            },
            "status": "confirmed",
        },
        "notes": [
            "原始函式名、線性位址、運算元與 bytes 保留在 private JSON。",
            "本摘要只證實 KOLMPICS 圖號 28，不把契約泛化到其他 82 個 payload。",
            "MSB-first 位元順序、畫面方向與尾 byte 仍須由解碼及目視 probe 驗證。",
        ],
    }
    private_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.parent.mkdir(parents=True, exist_ok=True)
    safe_path.write_text(json.dumps(safe, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN KOLMPICS index 28 probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
