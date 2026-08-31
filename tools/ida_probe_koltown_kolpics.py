"""IDA Pro 9.4：固定 KOLTOWN 對 KOLPICS.0 的消費契約。"""

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
EXPECTED_STRINGS = {
    0x166A: b"kol%s.0\0",
    0x1672: b"kol%s\0",
    0x16BA: b"kol%s\0",
    0x1802: b"kol%s\0",
    0x184A: b"kol%s\0",
}
FUNCTIONS = [
    0x10000, 0x10F9C, 0x1103F, 0x11262, 0x19B29, 0x19D98,
    0x19E9D, 0x1A429, 0x1A4FE, 0x1BD17, 0x1E3A8, 0x2D3FA,
]
EXPECTED_INSTRUCTIONS = {
    0x10570: "mov     ax, 166Ah", 0x10578: "call    sub_2D3FA",
    0x10580: "push    word_30A6C", 0x10584: "push    word_30A6A",
    0x1058C: "call    sub_19D98", 0x105AD: "mov     ax, 1672h",
    0x105BD: "mov     ax, word_34A0A", 0x105CC: "mov     ax, 0",
    0x105D4: "call    sub_1A4FE", 0x105DC: "mov     ax, 500h",
    0x105E0: "push    word_39824", 0x105E4: "call    sub_19B29",
    0x10FC0: "push    [bp+arg_0]", 0x10FC3: "push    [bp+arg_2]",
    0x10FC6: "call    sub_1A4FE", 0x1101D: "push    word_34A0E",
    0x1102D: "push    word_2F968", 0x11031: "call    sub_19E9D",
    0x11189: "cmp     [bp+arg_0], 64h", 0x11192: "mov     [bp+arg_0], 0Dh",
    0x111A3: "sub     [bp+arg_0], 1Ch", 0x111AA: "sub     [bp+arg_0], 3Eh",
    0x111AE: "cmp     [bp+arg_0], 5", 0x111B7: "cmp     [bp+arg_0], 0Ch",
    0x111C0: "push    word_2FBB8", 0x111C7: "add     ax, 0FFFAh",
    0x111CB: "call    sub_11262", 0x111F9: "mov     ax, 16BAh",
    0x1120D: "push    [bp+arg_0]", 0x11210: "call    sub_10F9C",
    0x1A548: "mov     ax, 1802h", 0x1A575: "push    word_2DF74",
    0x1A57D: "push    [bp+var_12]", 0x1A580: "call    sub_1A429",
    0x1E4C0: "mov     word_39A4E, 8", 0x1E4D0: "mov     word_39A4E, 0Ch",
    0x1E4F2: "mov     ax, 184Ah", 0x1E511: "push    word_39A4E",
    0x1E519: "call    sub_1A4FE", 0x1E521: "mov     ax, 8000h",
    0x1E530: "call    sub_19B29",
    0x1EAFE: "mov     cx, 6", 0x1EB04: "shl     ax, cl",
}


def disassembly(ea: int) -> str:
    rendered = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return rendered.split(" ;", 1)[0].rstrip()


def instruction(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disassembly(ea),
        "code_refs_to": [f"0x{x:x}" for x in idautils.CodeRefsTo(ea, False)],
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
        "callers": [f"0x{x:x}" for x in idautils.CodeRefsTo(start, False)],
        "instructions": [instruction(ea) for ea in idautils.FuncItems(start)],
    }


def archive_records() -> list[dict]:
    records = []
    rle = {1: (320, 200, 29), 2: (296, 32, 30), 4: (288, 155, 32),
           5: (288, 77, 33), 13: (288, 78, 100)}
    raw = {0: (1281, 256, 5, "startup", "strong inference", "strong inference"),
           8: (10241, 64, 160, "default", "confirmed", "strong inference"),
           12: (10241, 64, 160, "alternate", "confirmed", "strong inference")}
    for index in range(14):
        if index in rle:
            width, height, request = rle[index]
            records.append({"index": index, "kind": "rle", "width": width,
                            "height": height, "request": request,
                            "geometry_grade": "confirmed"})
        elif index in raw:
            decoded_bytes, width, rows, role, width_grade, height_grade = raw[index]
            records.append({"index": index, "kind": "raw-planar", "role": role,
                            "decoded_bytes": decoded_bytes,
                            "source_stride_bytes": 256, "planar_bytes_per_stride": 128,
                            "width": width, "height": rows,
                            "layout_grade": "confirmed",
                            "width_grade": width_grade, "height_grade": height_grade})
        else:
            records.append({"index": index, "kind": "placeholder", "decoded_bytes": 1})
    return records


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SUMMARY.json INPUT.EXE")
    private_path, summary_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    strings = []
    for offset, expected in EXPECTED_STRINGS.items():
        ea = DGROUP_EA + offset
        actual = ida_bytes.get_bytes(ea, len(expected))
        if actual != expected:
            raise RuntimeError(f"string mismatch at DGROUP+0x{offset:x}")
        strings.append({"ea": f"0x{ea:x}", "dgroup_offset": f"0x{offset:x}",
                        "bytes": actual.hex(), "text": expected[:-1].decode("ascii")})

    checks = []
    for ea, expected in EXPECTED_INSTRUCTIONS.items():
        actual = disassembly(ea)
        if actual != expected:
            raise RuntimeError(f"instruction mismatch at 0x{ea:x}: {actual!r} != {expected!r}")
        checks.append(instruction(ea))

    address_space = "IDA linear effective address for reconstructed unpacked MZ"
    records = archive_records()
    contract = {
        "source_stride_bytes": 256, "planar_bytes_per_stride": 128,
        "plane_bytes_per_row": 32, "plane_count": 4,
        "converted_pixels_per_block": 256, "bit_order": "most-significant bit first",
        "color_index": "plane3..plane0 form bits 3..0",
        "layout_grade": "confirmed by sub_19B29",
        "row_count_grade": "strong inference from decoded data length, sub_19B29 source stride and visual probe",
        "limitation": "index 8/12 are consumed with a 64-column offset; their 160-row extent and index 0 geometry remain inferred from decoded length",
    }
    shared = {
        "ida_version": idaapi.get_kernel_version(), "input_name": input_path.name,
        "input_sha256": digest, "address_space": address_space,
        "active_indices": [0, 1, 2, 4, 5, 8, 12, 13],
        "placeholder_indices": [3, 6, 7, 9, 10, 11],
        "rle_indices": [1, 2, 4, 5, 13], "raw_planar_indices": [0, 8, 12],
        "archive_records": records, "raw_planar_contract": contract,
        "evidence_notes": [
            "DGROUP+0x166a/0x1672 的其他近字組命中位於既有指令內，屬位元組巧合而非消費端。",
            "sub_1E3A8 傳給 sub_19B29 的 0x8000 不是 KOLPICS 索引 8/12 的資料長度證明。",
        ],
    }
    private = {"schema": "kol-ida-koltown-kolpics-private-v1", **shared,
               "strings": strings, "instruction_checks": checks,
               "functions": [function_record(ea) for ea in FUNCTIONS]}
    summary = {"schema": "kol-ida-koltown-kolpics-summary-v1", **shared,
               "strings": strings,
               "instruction_checks": [{"ea": x["ea"], "bytes": x["bytes"],
                                       "disassembly": x["disassembly"]} for x in checks],
               "function_starts": [f"0x{x:x}" for x in FUNCTIONS]}
    private_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    private_path.write_text(json.dumps(private, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    summary_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLTOWN KOLPICS probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
