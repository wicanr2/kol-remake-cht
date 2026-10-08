"""IDA 9.4：非破壞性匯出物品明細與空格等待分支；完整payload僅本機。"""
from pathlib import Path
import json,hashlib
import ida_auto,ida_bytes,ida_funcs,ida_lines,ida_nalt,idaapi,idautils,idc,ida_pro
ida_auto.auto_wait()
source=Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked.EXE')
digest=hashlib.sha256(source.read_bytes()).hexdigest()
assert digest=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
assert ida_nalt.retrieve_input_file_sha256().hex()==digest
assert idaapi.get_kernel_version()=='9.4'
indices=[]
for name in ['armour-address-index.json','battle-panel-address-index.json','player-gaps-address-index.json','target-cursor-address-index.json','item-detail-address-index.json','item-name-address-index.json']:
 p=Path('/src/docs/re')/name;d=json.loads(p.read_text(encoding='utf-8'))
 assert d['sha256']==digest and d['address_space']=='IDA linear'
 for a in d['entries'].values():assert a['confidence'] in ['已證實','強推論','假說','未知'] and a['evidence']
 indices.append((str(p),hashlib.sha256(p.read_bytes()).hexdigest(),d['entries']))
records=[]
for start in [0x16986,0x153FF,0x144C3,0x11722,0x1449C]:
 f=ida_funcs.get_func(start);assert f and f.start_ea==start
 rows=[]
 for ea in idautils.FuncItems(start):
  attached=[]
  for path,h,entries in indices:
   a=entries.get(hex(ea))
   if a:attached.append(dict(a,index=path,index_sha256=h,warning='' if a['confidence']=='已證實' else '警告：語意未證實'))
  rows.append(dict(original_ea=hex(ea),original_name=idc.get_name(ea),
   operand_text=ida_lines.tag_remove(idc.generate_disasm_line(ea,0) or ''),
   bytes=(ida_bytes.get_bytes(ea,ida_bytes.get_item_size(ea)) or b'').hex(),
   code_refs=[hex(x) for x in idautils.CodeRefsFrom(ea,False)],
   data_refs=[hex(x) for x in idautils.DataRefsFrom(ea)],annotations=attached,
   semantic=None,confidence='未知',warning='未附加已證實語意；原名只作定位',
   evidence=dict(input_sha256=digest,tool='IDA Pro 9.4',address_space='IDA linear',ea=hex(ea))))
 records.append(dict(original_start=hex(f.start_ea),original_end=hex(f.end_ea),
  original_name=idc.get_func_name(start),instructions=rows))
out=dict(schema='kol-item-detail-ida-v1',input=source.name,sha256=digest,tool='IDA Pro 9.4',
 address_space='IDA linear',database_sha256=hashlib.sha256(Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64').read_bytes()).hexdigest(),
 records=records)
Path('/out/item-name-annotated-private.json').write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)
