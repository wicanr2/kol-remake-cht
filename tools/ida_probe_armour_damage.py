"""唯讀既有資料庫副本，匯出護甲結算及耐久呼叫端的原始定位。"""
import hashlib
import json
from pathlib import Path
import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import ida_nalt
import ida_pro
import idaapi
import idautils
import idc

ida_auto.auto_wait()
source = Path('/src/workplace/analysis/koltown/KOLTOWN-unpacked.EXE')
sha = hashlib.sha256(source.read_bytes()).hexdigest()
assert sha == '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
assert ida_nalt.retrieve_input_file_sha256().hex() == sha
index_path = Path('/src/docs/re/armour-address-index.json')
index = json.loads(index_path.read_text(encoding='utf-8'))
assert index['sha256'] == sha and index['address_space'] == 'IDA linear'
assert index['tool'] == 'IDA Pro ' + idaapi.get_kernel_version()
for annotation in index['entries'].values():
    assert annotation['confidence'] in ('已證實', '強推論', '假說', '未知')
    assert annotation['semantic'] and annotation['evidence']
starts = {0x26D76, 0x22084, 0x26D3C, 0x22229, 0x22278, 0x22296, 0x277F4}
callers = [x.frm for x in idautils.XrefsTo(0x22084) if x.iscode]
for ea in callers:
    f = ida_funcs.get_func(ea)
    if f:
        starts.add(f.start_ea)
records = []
for start in sorted(starts):
    f = ida_funcs.get_func(start)
    assert f is not None
    records.append({'original_name': idc.get_func_name(start), 'start': hex(f.start_ea),
        'end': hex(f.end_ea), 'instructions': [
            {'ea': hex(ea), 'bytes': (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b'').hex(),
             'text': ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ''),
             'code_refs': [hex(x) for x in idautils.CodeRefsFrom(ea, False)],
             'semantic': None, 'confidence': 'unknown', 'warning': '原始指令，尚未附加已證實語意',
             'source': {'sha256': sha, 'tool': idaapi.get_kernel_version(), 'address_space': 'IDA linear'}}
            for ea in idautils.FuncItems(start)]})
for record in records:
    for instruction in record['instructions']:
        annotation = index['entries'].get(instruction['ea'])
        if annotation:
            instruction.update(annotation)
            instruction['warning'] = '' if annotation['confidence'] == '已證實' else '警告：語意尚未證實'
            instruction['annotation_index'] = 'docs/re/armour-address-index.json'
Path('/out/armour-private.json').write_text(json.dumps({'schema': 'kol-armour-ida-v1',
    'input': source.name, 'sha256': sha, 'tool': idaapi.get_kernel_version(),
    'annotation_index_sha256': hashlib.sha256(index_path.read_bytes()).hexdigest(),
    'durability_callers': [hex(ea) for ea in callers], 'records': records},
    ensure_ascii=False, indent=2), encoding='utf-8')
ida_pro.qexit(0)
