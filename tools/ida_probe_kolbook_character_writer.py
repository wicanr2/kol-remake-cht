"""IDA Pro 9.4：固定 KOLBOOK 角色建立 writer 與職業靜態表。"""

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


EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"
PROFESSION_META_EA = 0x19348
PROFESSION_META_BYTES = 39 * 5
PROFESSION_WORD_EA = 0x1940C
PROFESSION_WORD_BYTES = 39 * 2
EXPECTED_PROFESSION_META_SHA256 = "c15a9bfa5b026607e4377f52c01c4796bc8d3df685efcca1c523df016a264945"
EXPECTED_PROFESSION_WORD_SHA256 = "93845e8224b196978ccea8d0d5804ac8e4158cf751d63b85bf82f9c0064a842e"
EXPECTED_STORAGE_STRINGS = ["chardata", "charpics", "shldpics", "region%d.dat"]
PROFESSION_NAMES = [
    "BARBARIAN", "RANGER", "WARRION", "SQUIRE", "DARKGUARD", "WATCHMAN",
    "PLAINSMAN", "HUNTER", "REGULAR", "HIGHWAYMAN", "PIRATE", "ROGUE",
    "TIGRESS", "AMAZON", "HUNTRESS", "PLAINSWOMAN", "BREKLAND", "KLVAR",
    "MELOD", "PYAR", "THISM", "USIP", "BREKLAND", "KLVAR", "MELOD", "PYAR",
    "THISM", "USIP", "TUNNELER", "RATGUARD", "TROLLBANE", "MILITIA", "DIGGER",
    "SPIDER GUARD", "ORCBANE", "LEVY", "CLIFF GUARD", "ROCK RANGER", "FAR SEEKER",
]
WEAPON_SELECTORS = {
    1: "LO", 2: "BA", 3: "LB", 4: "GS", 5: "LC", 6: "WH", 7: "SC",
    8: "HL", 9: "BR", 10: "SB", 11: "MA", 12: "SS", 13: "LS", 14: "EB",
}
EXPECTED_INSTRUCTION_BYTES = {
    0x12064: "b90003",
    0x1206D: "b9f802",
    0x12075: "bf4418",
    0x12088: "b97c01",
    0x120B8: "8b5ef2",
    0x120C6: "8887bea6",
    0x120D7: "8b46fe",
    0x120DE: "83c324",
    0x120E1: "8887bea6",
    0x120EC: "83c31a",
    0x120EF: "8887bea6",
    0x120FA: "83c31c",
    0x120FD: "8887bea6",
    0x12117: "83c31d",
    0x1211A: "8887bea6",
    0x12130: "8b87cc01",
    0x12144: "83c32b",
    0x12147: "8887bea6",
    0x12158: "83c32c",
    0x1215B: "8897bea6",
    0x1216B: "8a870801",
    0x121BA: "83c357",
    0x121C0: "8887bea6",
    0x1220A: "83c34a",
    0x12210: "8887bea6",
    0x122A0: "83c330",
    0x122A6: "8887bea6",
    0x12327: "83c328",
    0x1232A: "8887bea6",
    0x1232E: "8b1ebaa6",
    0x12332: "83c37f",
    0x12335: "c687bea603",
    0x1249C: "83c32e",
    0x1249F: "8887bea6",
    0x124B1: "83c32f",
    0x124B4: "8887bea6",
    0x124D1: "83c340",
    0x124D7: "8887bea6",
    0x12501: "81c39000",
    0x12508: "8887bea6",
    0x125F4: "e8cb0e",
    0x125FA: "b80030",
    0x12609: "e8b411",
    0x12620: "b8802f",
    0x1263D: "8b5ef2",
    0x12640: "8a871a09",
    0x1264A: "8887c847",
    0x1266B: "e85211",
}


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def insn(ea: int) -> dict:
    size = ida_bytes.get_item_size(ea)
    return {
        "ea": f"0x{ea:x}",
        "bytes": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ""),
        "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [f"0x{x:x}" for x in idautils.DataRefsFrom(ea)],
    }


def function_record(start: int) -> dict:
    function = ida_funcs.get_func(start)
    if function is None:
        raise RuntimeError(f"missing function 0x{start:x}")
    return {
        "original_name": idc.get_func_name(start),
        "start_ea": f"0x{function.start_ea:x}",
        "end_ea": f"0x{function.end_ea:x}",
        "callers": [f"0x{x:x}" for x in sorted(idautils.CodeRefsTo(function.start_ea, False))],
        "instructions": [insn(ea) for ea in idautils.FuncItems(function.start_ea)],
    }


def operand_sites(*needles: str) -> list[dict]:
    lowered = tuple(needle.lower() for needle in needles)
    rows = []
    for start in idautils.Functions():
        for ea in idautils.FuncItems(start):
            row = insn(ea)
            if any(needle in row["disassembly"].lower() for needle in lowered):
                rows.append({"original_function": idc.get_func_name(start), "function_start_ea": f"0x{start:x}", **row})
    return rows


def c_string(ea: int) -> str:
    data = bytearray()
    for offset in range(128):
        value = ida_bytes.get_byte(ea + offset)
        if value == 0:
            return data.decode("ascii")
        data.append(value)
    raise RuntimeError(f"unterminated string at 0x{ea:x}")


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = sha(input_path.read_bytes())
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    profession_meta = ida_bytes.get_bytes(PROFESSION_META_EA, PROFESSION_META_BYTES)
    profession_words = ida_bytes.get_bytes(PROFESSION_WORD_EA, PROFESSION_WORD_BYTES)
    if profession_meta is None or profession_words is None:
        raise RuntimeError("missing profession table bytes")
    if sha(profession_meta) != EXPECTED_PROFESSION_META_SHA256:
        raise RuntimeError("profession meta table changed")
    if sha(profession_words) != EXPECTED_PROFESSION_WORD_SHA256:
        raise RuntimeError("profession word table changed")

    checks = []
    for ea, expected in EXPECTED_INSTRUCTION_BYTES.items():
        row = insn(ea)
        if row["bytes"] != expected:
            raise RuntimeError(f"instruction bytes mismatch at 0x{ea:x}: {row['bytes']} != {expected}")
        checks.append({"ea": row["ea"], "bytes": row["bytes"], "disassembly": row["disassembly"]})

    gold_values = [int.from_bytes(profession_words[i:i + 2], "little") for i in range(0, len(profession_words), 2)]
    starting_contracts = []
    for index, name in enumerate(PROFESSION_NAMES):
        meta = profession_meta[index * 5:(index + 1) * 5]
        secondary_selector = meta[0] >> 4
        primary_selector = meta[0] & 0x0F
        if primary_selector not in WEAPON_SELECTORS:
            raise RuntimeError(f"invalid primary weapon selector at profession {index}")
        if secondary_selector not in ({0} | set(WEAPON_SELECTORS)):
            raise RuntimeError(f"invalid secondary weapon selector at profession {index}")
        if any(value % 10 for value in meta[1:5]):
            raise RuntimeError(f"non-decimal skill unit at profession {index}")
        starting_contracts.append({
            "record_index": index,
            "menu_name": name,
            "gold_crowns": gold_values[index],
            "primary_weapon": WEAPON_SELECTORS[primary_selector],
            "primary_skill_pair_internal": [meta[3], meta[4]],
            "primary_skill_pair_manual_units": [meta[3] // 10, meta[4] // 10],
            "secondary_weapon": WEAPON_SELECTORS.get(secondary_selector),
            "secondary_skill_pair_internal": [meta[1], meta[2]],
            "secondary_skill_pair_manual_units": [meta[1] // 10, meta[2] // 10],
        })
    storage_string_eas = [0x19912, 0x1991C, 0x19926, 0x19930]
    storage_strings = [c_string(ea) for ea in storage_string_eas]
    if storage_strings != EXPECTED_STORAGE_STRINGS:
        raise RuntimeError(f"storage route strings changed: {storage_strings}")

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for KOLBOOK loaded MZ",
        "profession_meta": {
            "ea": f"0x{PROFESSION_META_EA:x}",
            "record_count": 39,
            "record_stride": 5,
            "sha256": sha(profession_meta),
        },
        "profession_word_table": {
            "ea": f"0x{PROFESSION_WORD_EA:x}",
            "record_count": 39,
            "record_stride": 2,
            "sha256": sha(profession_words),
        },
        "instruction_checks": checks,
        "storage_route_strings": [
            {"ea": f"0x{ea:x}", "value": value}
            for ea, value in zip(storage_string_eas, storage_strings)
        ],
    }
    private = {
        "schema": "kol-ida-kolbook-character-writer-private-v1",
        **shared,
        "profession_meta_hex": profession_meta.hex(),
        "profession_words_hex": profession_words.hex(),
        "character_creation_function": function_record(0x11820),
        "slot_finder_function": function_record(0x126F6),
        "save_function": function_record(0x137C0),
        "candidate_block_loader_function": function_record(0x135DB),
        "candidate_block_operand_sites": operand_sites("5a3eh", "a5c2", "23802"),
    }
    safe = {
        "schema": "kol-ida-kolbook-character-writer-summary-v1",
        **shared,
        "record_storage_contract": {
            "chardata_buffer_ea": "0x238fe",
            "slot_count": 16,
            "record_stride": 768,
            "fields": [
                {"offset": "0x00", "bytes": 16, "semantic": "name", "status": "confirmed"},
                {"offset": "0x1a", "bytes": 1, "semantic": "secondary selector", "status": "confirmed"},
                {"offset": "0x1c", "bytes": 1, "semantic": "profession record index", "status": "confirmed"},
                {"offset": "0x1d", "bytes": 7, "semantic": "primary internal stats", "status": "confirmed"},
                {"offset": "0x24", "bytes": 1, "semantic": "race enum", "status": "confirmed"},
                {"offset": "0x28", "bytes": 1, "semantic": "(race enum + 1) * 16 selector", "status": "confirmed"},
                {"offset": "0x29", "bytes": 1, "semantic": "floor(sum of four profession weapon-skill bytes / 200)", "status": "confirmed"},
                {"offset": "0x2b", "bytes": 2, "semantic": "starting gold crowns little-endian", "status": "confirmed"},
                {"offset": "0x30", "bytes": 16, "semantic": "primary starting weapon template", "status": "confirmed"},
                {"offset": "0x4a", "bytes": 8, "semantic": "primary starting weapon template prefix", "status": "confirmed"},
                {"offset": "0x52", "bytes": 1, "semantic": "primary starting weapon template byte 13", "status": "confirmed"},
                {"offset": "0x54", "bytes": 1, "semantic": "primary weapon first skill internal", "status": "confirmed"},
                {"offset": "0x56", "bytes": 1, "semantic": "primary weapon second skill internal", "status": "confirmed"},
                {"offset": "0x57", "bytes": 8, "semantic": "secondary starting weapon template prefix", "status": "confirmed"},
                {"offset": "0x5f", "bytes": 1, "semantic": "secondary weapon template byte 13", "status": "confirmed"},
                {"offset": "0x61", "bytes": 1, "semantic": "secondary weapon first skill internal", "status": "confirmed"},
                {"offset": "0x63", "bytes": 1, "semantic": "secondary weapon second skill internal", "status": "confirmed"},
                {"offset": "0x7f", "bytes": 1, "semantic": "constant 3; gameplay meaning unknown", "status": "confirmed write, unknown semantics"},
            ],
            "region_metadata_copy_map": [
                {"region_source_offset": "0x01", "record_offset": "0x2e", "bytes": 1},
                {"region_source_offset": "0x02", "record_offset": "0x2f", "bytes": 1},
                {"region_source_offset": "0x03", "record_offset": "0x40", "bytes": 10},
                {"region_source_offset": "0x0d", "record_offset": "0x90", "bytes": 3},
                {"region_source_offset": "0x10", "record_offset": "0x93", "bytes": 1},
                {"region_source_offset": "0x11", "record_offset": "0x94", "bytes": 1},
                {"region_source_offset": "0x12", "record_offset": "0x97", "bytes": 1},
                {"region_source_offset": "0x13", "record_offset": "0x13", "bytes": 1},
                {"region_source_offset": "0x15", "record_offset": "0x99", "bytes": 1},
                {"region_source_offset": "0x18", "record_offset": "0x9c", "bytes": 1},
                {"region_source_offset": "0x1a", "record_offset": "0x14", "bytes": 1},
                {"region_source_offset": "0x1b", "record_offset": "0x15", "bytes": 1},
                {"region_source_offset": "0x1d", "record_offset": "0x17", "bytes": 1},
                {"region_source_offset": "0x1e", "record_offset": "0x18", "bytes": 1}
            ],
            "region_metadata_semantics": "unknown; byte routing from selected REGION%d.DAT record is confirmed",
            "save_bytes": 12288,
        },
        "charpics_storage_contract": {
            "buffer_ea": "0x1aa84",
            "slot_count": 16,
            "record_stride": 760,
            "save_bytes": 12160,
        },
        "shldpics_storage_contract": {
            "buffer_ea": "0x1da08",
            "slot_count": 16,
            "record_stride": 760,
            "initial_template_source_ea": "0x19b5a",
            "copied_bytes_on_character_commit": 760,
            "save_bytes": 12160,
        },
        "weapon_selector_map": [{"selector": key, "manual_abbreviation": value} for key, value in WEAPON_SELECTORS.items()],
        "profession_starting_contracts": starting_contracts,
        "writer_coverage": "after clearing the 768-byte slot, all direct CHARDATA writes in sub_11820 are covered by fields plus region_metadata_copy_map; no additional profession-derived starting item or skill write exists in this function",
        "status": "writer offsets and starting gold/weapon/skill contracts confirmed; unlisted offsets remain zero after creation or have no direct write in sub_11820",
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"KOLBOOK character writer probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
