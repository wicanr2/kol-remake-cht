"""IDA Pro 9.4：重建 KOLTOWN 第一段搬移映射並探查第二階段 code。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_entry
import ida_funcs
import ida_ida
import ida_lines
import ida_nalt
import ida_segment
import ida_ua
import idaapi
import idautils
import idc


EXPECTED_SHA256 = "c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221"


def word(ea: int) -> int:
    return ida_bytes.get_word(ea)


def instruction_record(ea: int) -> dict[str, object]:
    size = ida_bytes.get_item_size(ea)
    data = ida_bytes.get_bytes(ea, size) or b""
    return {
        "ea": f"0x{ea:x}",
        "bytes": data.hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
    }


def main() -> None:
    if len(idc.ARGV) != 3:
        raise RuntimeError("usage: ida_probe_koltown_stage2.py PRIVATE.json SUMMARY.json")
    private_output = Path(idc.ARGV[1])
    summary_output = Path(idc.ARGV[2])
    input_path = Path(ida_nalt.get_input_file_path())
    raw = input_path.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")

    ordinal = ida_entry.get_entry_ordinal(0)
    entry_ea = ida_entry.get_entry(ordinal)
    segment = ida_segment.getseg(entry_ea)
    if segment is None:
        raise RuntimeError("entry segment missing")
    stub_base = segment.start_ea
    load_base = ida_ida.inf_get_min_ea()

    copy_length = word(stub_base + 0x06)
    destination_paragraph_delta = word(stub_base + 0x0C)
    transfer_ip = 0x0038
    source_stage2_ea = stub_base + transfer_ip
    destination_base_ea = load_base + destination_paragraph_delta * 16
    runtime_stage2_ea = destination_base_ea + transfer_ip
    mapping_delta = destination_base_ea - stub_base
    final_ip = word(stub_base + 0x00)
    final_cs_delta = word(stub_base + 0x02)
    final_sp = word(stub_base + 0x08)
    final_ss_delta = word(stub_base + 0x0A)
    final_entry_ea = load_base + final_cs_delta * 16 + final_ip
    error_marker = ida_bytes.get_bytes(stub_base + 0x117, 0x16) or b""

    if copy_length != 0x131F or destination_paragraph_delta != 0x1F92:
        raise RuntimeError("stage-1 constants changed")
    if source_stage2_ea != 0x2F548 or runtime_stage2_ea != 0x2F958:
        raise RuntimeError("stage-2 mapping changed")
    if error_marker != b"Packed file is corrupt":
        raise RuntimeError(f"unexpected packer marker: {error_marker!r}")

    ida_bytes.del_items(source_stage2_ea, ida_bytes.DELIT_SIMPLE, 0x400)
    if ida_ua.create_insn(source_stage2_ea) <= 0:
        raise RuntimeError("cannot create first stage-2 instruction")
    ida_funcs.add_func(source_stage2_ea, idaapi.BADADDR)
    ida_auto.auto_wait()
    function = ida_funcs.get_func(source_stage2_ea)
    if function is None:
        raise RuntimeError("IDA did not create stage-2 function")

    instruction_eas = list(idautils.FuncItems(function.start_ea))
    instructions = [instruction_record(ea) for ea in instruction_eas]
    code_xrefs = []
    for ea in instruction_eas:
        for target in idautils.CodeRefsFrom(ea, False):
            code_xrefs.append({"from": f"0x{ea:x}", "to": f"0x{target:x}"})

    common = {
        "schema": "kol-ida-koltown-stage2-summary-v1",
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address; runtime relocated EA shown separately",
        "stage1_entry_ea": f"0x{entry_ea:x}",
        "source_stub_base_ea": f"0x{stub_base:x}",
        "copy_length": copy_length,
        "copy_source_range": [f"0x{stub_base:x}", f"0x{stub_base + copy_length:x}"],
        "copy_destination_range": [
            f"0x{destination_base_ea:x}",
            f"0x{destination_base_ea + copy_length:x}",
        ],
        "mapping_delta": f"0x{mapping_delta:x}",
        "source_stage2_ea": f"0x{source_stage2_ea:x}",
        "runtime_stage2_ea": f"0x{runtime_stage2_ea:x}",
        "decompressed_entry": {
            "relative_cs": f"0x{final_cs_delta:x}",
            "ip": f"0x{final_ip:x}",
            "ida_linear_ea": f"0x{final_entry_ea:x}",
        },
        "decompressed_stack": {
            "relative_ss": f"0x{final_ss_delta:x}",
            "sp": f"0x{final_sp:x}",
        },
        "packer_error_marker": error_marker.decode("ascii"),
        "decompression_command_families": {
            "0xb0_0xb1": "repeat one byte for count",
            "0xb2_0xb3": "copy count bytes",
            "low_bit": "last-block flag",
        },
        "relocation_table_source_offset": "0x12d",
        "stage2_function_start_ea": f"0x{function.start_ea:x}",
        "stage2_function_end_ea": f"0x{function.end_ea:x}",
        "stage2_instruction_count": len(instruction_eas),
        "stage2_code_xref_count": len(code_xrefs),
    }
    private = dict(common)
    private["schema"] = "kol-ida-koltown-stage2-private-v1"
    private["instructions"] = instructions
    private["code_xrefs"] = code_xrefs

    private_output.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    summary_output.write_text(json.dumps(common, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN stage-2 probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
