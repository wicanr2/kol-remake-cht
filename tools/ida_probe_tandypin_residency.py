"""IDA Pro 9.4：固定 TANDYPIN 的 DOS 常駐記憶體行為。"""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_lines
import idaapi
import idautils
import idc
import ida_ua


EXPECTED_SHA256 = "036c39f330a4fe09bc2729756816b869f1154515bf4befba190e767d4e3cdce3"
EXPECTED = {
    # 檔案中的 b8 01 00 具有 MZ relocation；IDA 載入位址空間會成為 b8 01 10。
    0x10000: "b80110",
    0x10003: "8ed8",
    0x10005: "ba0104",
    0x10008: "b80031",
    0x1000B: "cd21",
}


def disassembly(ea: int) -> str:
    text = ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or "")
    return text.split(";", 1)[0].rstrip()


def main() -> None:
    if len(idc.ARGV) != 4:
        raise RuntimeError("usage: script.py PRIVATE.json SAFE.json INPUT.EXE")
    private_path, safe_path, input_path = map(Path, idc.ARGV[1:4])
    digest = hashlib.sha256(input_path.read_bytes()).hexdigest()
    if digest != EXPECTED_SHA256:
        raise RuntimeError(f"unexpected input SHA-256: {digest}")
    ida_auto.auto_wait()

    checks = []
    for ea, expected_bytes in EXPECTED.items():
        ida_ua.create_insn(ea)
        actual = disassembly(ea)
        size = ida_bytes.get_item_size(ea)
        raw = (ida_bytes.get_bytes(ea, size) or b"").hex()
        if raw != expected_bytes:
            private_path.write_text(
                json.dumps({"mismatch_ea": f"0x{ea:x}", "actual": raw, "expected": expected_bytes}) + "\n",
                encoding="utf-8",
            )
            raise RuntimeError(f"byte mismatch at 0x{ea:x}: {raw!r} != {expected_bytes!r}")
        checks.append({
            "ea": f"0x{ea:x}",
            "bytes": raw,
            "disassembly": actual,
            "code_refs_from": [f"0x{x:x}" for x in idautils.CodeRefsFrom(ea, False)],
        })

    shared = {
        "ida_version": idaapi.get_kernel_version(),
        "input_name": input_path.name,
        "input_sha256": digest,
        "address_space": "IDA linear effective address for TANDYPIN loaded MZ",
    }
    private = {"schema": "kol-ida-tandypin-private-v1", **shared, "instruction_checks": checks}
    safe = {
        "schema": "kol-ida-tandypin-summary-v1",
        **shared,
        "instruction_checks": [
            {"ea": row["ea"], "bytes": row["bytes"], "disassembly": row["disassembly"]}
            for row in checks
        ],
        "contract": {
            "dos_interrupt": "INT 21h, AH=31h terminate-and-stay-resident",
            "return_code_al": 0,
            "resident_paragraphs_dx": "0x0401",
            "speaker_or_pit_io_in_entry": False,
            "audio_asset_or_player": False,
            "status": "confirmed",
        },
        "limitation": "KOL 啟動它的條件與保留這段記憶體的產品目的仍需由 KOL.EXE caller 證據判定。",
    }
    private_path.write_text(json.dumps(private, indent=2) + "\n", encoding="utf-8")
    safe_path.write_text(json.dumps(safe, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    print(f"TANDYPIN residency probe failed: {exc}", file=sys.stderr)
    idc.qexit(1)
