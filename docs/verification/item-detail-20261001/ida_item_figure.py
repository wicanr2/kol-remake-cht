"""IDA 9.4：立繪空物品與搬移原址證據；只匯出暫存資料庫副本。"""
from pathlib import Path
import json,hashlib
import ida_auto,ida_bytes,ida_funcs,ida_lines,ida_nalt,idaapi,idautils,idc,ida_pro
ida_auto.auto_wait()
src=Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked.EXE')
digest=hashlib.sha256(src.read_bytes()).hexdigest()
assert digest=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
assert ida_nalt.retrieve_input_file_sha256().hex()==digest and idaapi.get_kernel_version()=='9.4'
indices=[]
for name in ['item-detail-address-index.json','item-name-address-index.json','item-action-address-index.json','item-figure-address-index.json']:
 p=Path('/src/docs/re')/name
 if not p.exists():continue
 d=json.loads(p.read_text(encoding='utf-8'));assert d['sha256']==digest and d['address_space']=='IDA linear'
 indices.append((str(p),hashlib.sha256(p.read_bytes()).hexdigest(),d['entries']))
records=[]
for start in [0x18F3A,0x18CD5]:
 f=ida_funcs.get_func(start);assert f and f.start_ea==start;rows=[]
 for ea in idautils.FuncItems(start):
  attached=[]
  for path,h,entries in indices:
   if a:=entries.get(hex(ea)):
    attached.append(dict(a,index=path,index_sha256=h,warning='' if a['confidence']=='已證實' else '警告：語意未證實'))
  rows.append(dict(original_ea=hex(ea),original_name=idc.get_name(ea),
   operand_text=ida_lines.tag_remove(idc.generate_disasm_line(ea,0) or ''),
   bytes=(ida_bytes.get_bytes(ea,ida_bytes.get_item_size(ea)) or b'').hex(),
   code_refs=[hex(x) for x in idautils.CodeRefsFrom(ea,False)],data_refs=[hex(x) for x in idautils.DataRefsFrom(ea)],annotations=attached,
   semantic=None,confidence='未知',warning='未附加已證實語意；原名只作定位',
   evidence=dict(input_sha256=digest,tool='IDA Pro 9.4',address_space='IDA linear',ea=hex(ea))))
 records.append(dict(original_start=hex(f.start_ea),original_end=hex(f.end_ea),original_name=idc.get_func_name(start),instructions=rows))
out=dict(schema='kol-item-figure-ida-v1',input=src.name,sha256=digest,tool='IDA Pro 9.4',address_space='IDA linear',
 database_sha256=hashlib.sha256(Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64').read_bytes()).hexdigest(),records=records)
target=Path('/out/item-figure-annotated-private.json' if any(p.endswith('/item-figure-address-index.json') for p,h,e in indices) else '/out/item-figure-ida-private.json')
assert not target.exists();target.write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)
