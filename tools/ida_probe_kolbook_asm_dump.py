"""IDA Pro 9.4：把 KOLBOOK 整份反組譯寫成文字（只讀，不改名）。

  idat -A -S"tools/ida_probe_kolbook_asm_dump.py OUT.asm OUT.json IN.EXE" IN.EXE

輸出的 `.asm` 只留本機（`workplace/`）；JSON 只記雜湊、函式數與段落。
"""

import hashlib
import json
import sys

import ida_auto
import ida_nalt
import ida_pro
import idaapi
import idautils
import idc

EXPECTED_SHA256 = "2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832"

ida_auto.auto_wait()
out_asm, out_json = idc.ARGV[1], idc.ARGV[2]
sha = ida_nalt.retrieve_input_file_sha256().hex()
assert sha == EXPECTED_SHA256, sha
idc.gen_file(idc.OFILE_ASM, out_asm, 0, idc.BADADDR, 0)
body = open(out_asm, "rb").read()
summary = {
    "schema": "kol-ida-asm-dump-v1", "input": "KOLBOOK.EXE", "sha256": sha,
    "tool": "IDA Pro " + idaapi.get_kernel_version(), "address_space": "IDA linear",
    "functions": len(list(idautils.Functions())),
    "segments": [{"start": hex(s), "name": idc.get_segm_name(s)} for s in idautils.Segments()],
    "asm_sha256": hashlib.sha256(body).hexdigest(), "asm_bytes": len(body),
}
json.dump(summary, open(out_json, "w", encoding="utf-8"), ensure_ascii=False, indent=2)
ida_pro.qexit(0)
