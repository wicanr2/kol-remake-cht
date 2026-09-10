"""敗北共同出口窄查；保留原始定位，自動合併已審查索引，不改資料庫名稱。"""
import json
import os
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
              tool='IDA Pro '+idaapi.get_kernel_version(), address_space='IDA linear')
index = json.loads(Path('/src/docs/re/player-gaps-address-index.json').read_text(encoding='utf-8'))
assert index['sha256'] == sha and index['address_space'] == source['address_space']
assert index['input'] == source['input'] and index['tool'] == source['tool']

def row(ea):
    r = dict(ea=hex(ea), original_name=idc.get_func_name(ea),
             bytes=(ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b'').hex(),
             text=ida_lines.tag_remove(idc.generate_disasm_line(ea, 0) or ''),
             code_refs=[hex(x) for x in idautils.CodeRefsFrom(ea, False)],
             semantic=None, confidence='未知', evidence=source,
             warning='警告：原始指令，語意未經審查')
    note = index['entries'].get(hex(ea))
    if note:
        assert set(note) == {'semantic', 'confidence', 'evidence'}
        assert note['confidence'] in ('已證實','強推論','假說','未知') and note['evidence']
        r.update(note)
        r['warning'] = '' if note['confidence'] == '已證實' else '警告：語意尚未證實'
    return r

f = ida_funcs.get_func(0x21765)
assert f and f.start_ea == 0x1E3A8
exit_rows = [row(ea) for ea in idautils.FuncItems(f.start_ea) if ea >= 0x21000]
refs = [row(x.frm) for x in idautils.XrefsTo(0x3AC7D)]
callers = []
for x in idautils.XrefsTo(f.start_ea):
    if x.iscode:
        caller = ida_funcs.get_func(x.frm)
        callers.append(dict(call=hex(x.frm), start=hex(caller.start_ea),
                            end=hex(caller.end_ea), instructions=[row(ea) for ea in idautils.FuncItems(caller.start_ea)]))
# 取址線索也需查；未命中不能證明沒有間接寫入或整區覆寫。
address_candidates = []
for start in idautils.Functions():
    for ea in idautils.FuncItems(start):
        if any(idc.get_operand_type(ea, n) and
               idc.get_operand_value(ea, n) & 0xffff in (0xcdad, 0x3253, 0xac7d)
               for n in range(3)):
            address_candidates.append(row(ea))
out = dict(schema='kol-battle-exit-re-v1', source=source,
           helpers=[dict(start=hex(start), instructions=[row(ea) for ea in idautils.FuncItems(start)]) for start in (0x2886C,)],
           function=dict(start=hex(f.start_ea), end=hex(f.end_ea)),
           exit=exit_rows, direct_refs=refs, address_candidates=address_candidates, callers=callers,
           limitation='位址候選掃描不是完整間接資料流證明；未命中不能升格為恆零')
Path('/out/battle-exit-private.json').write_text(json.dumps(out,ensure_ascii=False,indent=2),encoding='utf-8')
if os.getenv('KOL_RE_MONSTER_INIT') == '1':
    ai = ida_funcs.get_func(0x23EBB)
    assert ai
    monster = dict(source=source,
                   conversion=[row(ea) for fn in (0x222BF, 0x22319, 0x22376) for ea in idautils.FuncItems(fn)],
                   initial=[row(ea) for ea in idautils.FuncItems(f.start_ea) if ea < 0x1F100],
                   ai=[row(ea) for ea in idautils.FuncItems(ai.start_ea)],
                   panel=[row(ea) for ea in idautils.FuncItems(0x22EFC)],
                   limitation='原始指令窄查；未知語意不自動升格')
    Path('/out/monster-kind-private.json').write_text(json.dumps(monster,ensure_ascii=False,indent=2),encoding='utf-8')
if os.getenv('KOL_RE_FIELD_POINTER') == '1':
    field = dict(source=source,
                 functions=[dict(start=hex(fn), instructions=[row(ea) for ea in idautils.FuncItems(fn)])
                            for fn in (0x223C0, 0x2B3D1)],
                 limitation='原始指令窄查；未知語意不自動升格')
    Path('/out/field-pointer-private.json').write_text(json.dumps(field,ensure_ascii=False,indent=2),encoding='utf-8')
ida_pro.qexit(0)
