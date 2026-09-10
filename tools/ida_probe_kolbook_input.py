"""一次性 IDA 資料庫：匯出書本讀鍵候選函式與直接呼叫關係，不附推測命名。"""
import hashlib
import json
from pathlib import Path
import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import ida_pro
import idaapi
import idautils
import idc

ida_auto.auto_wait()
source=Path(idc.get_input_file_path())
sha=hashlib.sha256(source.read_bytes()).hexdigest()
assert sha=="2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832",sha
starts=set()
for ea in (0x187F5,0x14B43,0x14BAA):
    f=ida_funcs.get_func(ea)
    if f: starts.add(f.start_ea)
for f in idautils.Functions():
    for ea in idautils.FuncItems(f):
        if idc.print_insn_mnem(ea)=="int" and idc.get_operand_value(ea,0) in (0x16,0x21):
            starts.add(f)
records=[]
for start in sorted(starts):
    f=ida_funcs.get_func(start)
    records.append({"original_name":idc.get_func_name(start),"start":hex(start),"end":hex(f.end_ea),
        "semantic":"未判定的輸入候選或服務常式","confidence":"unknown","warning":"不得將候選位置視為已證實玩法語意",
        "source":{"input_sha256":sha,"tool":"IDA Pro "+idaapi.get_kernel_version(),"address_space":"IDA linear"},
        "callers":[hex(x.frm) for x in idautils.XrefsTo(start) if x.iscode],
        "instructions":[{"ea":hex(ea),"bytes":(ida_bytes.get_bytes(ea,ida_bytes.get_item_size(ea)) or b'').hex(),
          "text":ida_lines.tag_remove(idc.generate_disasm_line(ea,0) or ''),
          "code_refs":[hex(x) for x in idautils.CodeRefsFrom(ea,False)]} for ea in idautils.FuncItems(start)]})
Path('/out/kolbook-input-private.json').write_text(json.dumps({"schema":"kol-ida-input-probe-v1","input":source.name,"sha256":sha,"tool":idaapi.get_kernel_version(),"segments":[{"start":hex(s),"name":idc.get_segm_name(s)} for s in idautils.Segments()],"function_count":sum(1 for _ in idautils.Functions()),"records":records},ensure_ascii=False,indent=2),encoding="utf-8")
ida_pro.qexit(0)
