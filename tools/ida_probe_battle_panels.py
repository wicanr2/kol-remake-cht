"""由既有 IDA 資料庫找出 0x101 裝備文字的使用端；不修改原始定位。"""
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
provenance = {'input': source.name, 'sha256': sha,
              'tool': 'IDA Pro ' + idaapi.get_kernel_version(), 'address_space': 'IDA linear'}
index_path = Path('/src/docs/re/battle-panel-address-index.json')
index = json.loads(index_path.read_text(encoding='utf-8'))
assert index['sha256'] == sha and index['address_space'] == provenance['address_space']
assert index['tool'] == provenance['tool']
for entry in index['entries'].values():
    assert entry['confidence'] in ('已證實', '強推論', '假說', '未知')
    assert entry['semantic'] and entry['evidence']

def annotate(row):
    entry = index['entries'].get(row['ea'])
    if entry:
        row.update(entry)
        row['warning'] = '' if entry['confidence'] == '已證實' else '警告：語意尚未證實'
        row['annotation_index'] = 'docs/re/battle-panel-address-index.json'
    return row
anchors = []
starts = set()
for start in idautils.Functions():
    for ea in idautils.FuncItems(start):
        if any(idc.get_operand_type(ea, op) == idc.o_imm and
               idc.get_operand_value(ea, op) == 0x101 for op in range(2)):
            anchors.append(hex(ea))
            starts.add(start)
records = []
starts.update((0x2324F, 0x27E20, 0x11722, 0x24131, 0x1E3A8,
               0x1103F, 0x11262, 0x1C94E, 0x1C9AB, 0x1D44A, 0x228B7, 0x1C53D,
               0x23638, 0x1A8FD, 0x252CB, 0x2374C, 0x2531C, 0x253E7))
for ea in idautils.DataRefsTo(0x3B264):
    f = ida_funcs.get_func(ea)
    if f:
        starts.add(f.start_ea)
for start in idautils.Functions():
    for ea in idautils.FuncItems(start):
        if any(idc.get_operand_type(ea, op) == idc.o_imm and
               idc.get_operand_value(ea, op) == 0x3254 for op in range(2)):
            starts.add(start)
for ea in idautils.DataRefsTo(0x2FBB8):
    f = ida_funcs.get_func(ea)
    if f:
        starts.add(f.start_ea)
for start in sorted(starts):
    f = ida_funcs.get_func(start)
    records.append({'original_name': idc.get_func_name(start), 'start': hex(start),
        'end': hex(f.end_ea), 'callers': [hex(x.frm) for x in idautils.XrefsTo(start) if x.iscode],
        'instructions': [
            annotate({'ea': hex(ea), 'bytes': (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b'').hex(),
             'text': ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ''),
             'code_refs': [hex(x) for x in idautils.CodeRefsFrom(ea, False)],
             'semantic': None, 'confidence': '未知', 'warning': '原始指令，尚未附加已證實語意',
             'source': provenance})
            for ea in idautils.FuncItems(start)]})
Path('/out/panels-private.json').write_text(json.dumps({'schema': 'kol-battle-panels-v1',
    'source': provenance, 'immediate_0101_anchors': anchors, 'records': records,
    'annotation_index_sha256': hashlib.sha256(index_path.read_bytes()).hexdigest(),
    'mode_references': [annotate({'ea': hex(x.frm), 'original_name': idc.get_func_name(x.frm),
        'bytes': (ida_bytes.get_bytes(x.frm, ida_bytes.get_item_size(x.frm)) or b'').hex(),
        'text': ida_lines.tag_remove(idc.generate_disasm_line(x.frm, 0) or ''),
        'semantic': None, 'confidence': '未知', 'warning': '僅交叉參照，尚未確認讀寫語意',
        'source': provenance}) for x in idautils.XrefsTo(0x303FC)
        if ida_bytes.is_code(ida_bytes.get_flags(x.frm))],
    'body_rectangles': annotate({'ea': '0x2f12c', 'source': provenance,
        'bytes': ida_bytes.get_bytes(0x2F12C, 72).hex(),
        'semantic': None, 'confidence': '未知', 'warning': '原始資料，尚待呼叫端核對'}),
    'body_cache_filename': annotate({'ea': '0x2f548', 'source': provenance,
        'bytes': ida_bytes.get_bytes(0x2F548, 10).hex(),
        'semantic': None, 'confidence': '未知', 'warning': '原始格式字串，須連同 0x1060D 呼叫解讀'}),
    'equipment_offsets': annotate({'ea': '0x2ea50', 'source': provenance,
        'bytes': ida_bytes.get_bytes(0x2EA50, 12).hex(),
        'semantic': None, 'confidence': '未知', 'warning': '原始資料，尚待呼叫端核對'})},
    ensure_ascii=False, indent=2), encoding='utf-8')
ida_pro.qexit(0)
