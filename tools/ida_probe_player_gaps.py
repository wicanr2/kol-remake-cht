"""最小玩家缺口查詢；唯讀資料庫副本，不改名，不從散文匯入語意。"""
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
sha = '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
assert ida_nalt.retrieve_input_file_sha256().hex() == sha
source = dict(input='KOLTOWN-unpacked.EXE', sha256=sha,
              tool='IDA Pro ' + idaapi.get_kernel_version(), address_space='IDA linear')
index_file = Path('/src/docs/re/player-gaps-address-index.json')
index = json.loads(index_file.read_text(encoding='utf-8')) if index_file.exists() else {'entries': {}}
if index_file.exists():
    assert index['sha256'] == sha and index['address_space'] == 'IDA linear'
    assert index['input'] == source['input'] and index['tool'] == 'IDA Pro 9.4'
records = []
starts = {0x11390, 0x11A93, 0x2AAA8, 0x144C3, 0x1A8DF, 0x13E82, 0x2B3D1, 0x2A18E, 0x19E9D, 0x2A39B, 0x29B40}
for ref in idautils.XrefsTo(0x34A1A):
    func = ida_funcs.get_func(ref.frm)
    if func:
        starts.add(func.start_ea)
for start in sorted(starts):
    f = ida_funcs.get_func(start)
    assert f and f.start_ea == start
    rows = []
    for ea in idautils.FuncItems(start):
        row = dict(ea=hex(ea), original_name=idc.get_func_name(ea),
                   bytes=(ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b'').hex(),
                   text=ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ''),
                   code_refs=[hex(x) for x in idautils.CodeRefsFrom(ea, False)],
                   semantic=None, confidence='未知', evidence=source,
                   warning='原始指令；未附加已證實語意')
        if row['ea'] in index['entries']:
            note = index['entries'][row['ea']]
            assert note['confidence'] in ('已證實', '強推論', '假說', '未知') and note['evidence']
            assert set(note) == {'semantic', 'confidence', 'evidence'}
            row.update(note)
            row['warning'] = '' if note['confidence'] == '已證實' else '警告：語意尚未證實'
        rows.append(row)
    records.append(dict(start=hex(start), end=hex(f.end_ea), original_name=idc.get_func_name(start),
                        callers=[hex(x.frm) for x in idautils.XrefsTo(start) if x.iscode], instructions=rows))
refs = []
for x in idautils.XrefsTo(0x2DF2A):
    refs.append(dict(ea=hex(x.frm), original_name=idc.get_func_name(x.frm),
        text=ida_lines.tag_remove(idc.generate_disasm_line(x.frm, 0) or ''),
        bytes=(ida_bytes.get_bytes(x.frm, ida_bytes.get_item_size(x.frm)) or b'').hex(),
        semantic=None, confidence='未知', evidence=source, warning='隊員數欄位交叉參照；各讀寫端用途尚待逐項審查'))
Path('/out/player-gaps-private.json').write_text(json.dumps(dict(schema='kol-player-gaps-re-v1',
    source=source, records=records, mode_refs=refs), ensure_ascii=False, indent=2), encoding='utf-8')
ida_pro.qexit(0)
