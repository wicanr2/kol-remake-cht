"""IDA 9.4：目標格標記的非破壞性證據匯出；payload 僅本機。"""
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
for name in ['armour-address-index.json','battle-panel-address-index.json','player-gaps-address-index.json','target-cursor-address-index.json']:
    p=Path('/src/docs/re')/name;d=json.loads(p.read_text(encoding='utf-8'))
    assert d['sha256']==digest and d['address_space']=='IDA linear'
    for a in d['entries'].values():
        assert a['confidence'] in ['已證實','強推論','假說','未知'] and a['semantic'] and a['evidence']
    indices.append((str(p),hashlib.sha256(p.read_bytes()).hexdigest(),d['entries']))
starts={0x1C683,0x24131,0x11390,0x1D21D,0x11722,0x228B7,0x249AE,
        ida_funcs.get_func(0x1F38D).start_ea}
f=ida_funcs.get_func(0x1C683);assert f and f.start_ea==0x1C683
for ea in idautils.FuncItems(f.start_ea):
    if idc.print_insn_mnem(ea).lower().startswith('call'):
        for target in idautils.CodeRefsFrom(ea,False):
            callee=ida_funcs.get_func(target)
            if callee:starts.add(callee.start_ea)
records=[]
for start in sorted(starts):
    f=ida_funcs.get_func(start);assert f
    rows=[]
    for ea in idautils.FuncItems(start):
        attached=[]
        for path,h,entries in indices:
            a=entries.get(hex(ea))
            if a:attached.append(dict(a,index=path,index_sha256=h,warning='' if a['confidence']=='已證實' else '警告：語意未證實'))
        rows.append({'original_ea':hex(ea),'original_name':idc.get_name(ea),
            'operand_text':ida_lines.tag_remove(idc.generate_disasm_line(ea,0) or ''),
            'bytes':(ida_bytes.get_bytes(ea,ida_bytes.get_item_size(ea)) or b'').hex(),
            'code_refs':[hex(x) for x in idautils.CodeRefsFrom(ea,False)],
            'data_refs':[hex(x) for x in idautils.DataRefsFrom(ea)],
            'annotations':attached,'semantic':None,'confidence':'未知',
            'warning':'未附加已證實語意；原始名稱僅供定位',
            'evidence':{'input_sha256':digest,'tool':'IDA Pro 9.4','address_space':'IDA linear','ea':hex(ea)}})
    records.append({'original_name':idc.get_func_name(start),'original_start':hex(f.start_ea),
        'original_end':hex(f.end_ea),'annotations':[], 'confidence':'未知',
        'warning':'函式語意尚未判定；保留原始名稱與邊界','instructions':rows})
out={'schema':'kol-target-cursor-ida-v1','input':source.name,'sha256':digest,
     'tool':'IDA Pro 9.4','address_space':'IDA linear',
     'database_sha256':hashlib.sha256(Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64').read_bytes()).hexdigest(),
     'indices':[{'path':p,'sha256':h} for p,h,e in indices],
     'xrefs_to_cursor_draw':[{'from':hex(x.frm),'to':hex(x.to),'iscode':x.iscode,'type':x.type} for x in idautils.XrefsTo(0x1C683)],
     'records':records}
Path('/out/cursor-annotated-private.json').write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
ida_pro.qexit(0)
