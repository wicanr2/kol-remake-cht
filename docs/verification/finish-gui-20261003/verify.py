"""Docker 內唯讀重驗：kol-package:20260907，儲存庫掛 /src。

執行 python3 /src/docs/verification/finish-gui-20261003/verify.py。
正常遊戲重播腳本及中間建置在收據索引的 workplace；不宣稱原版或真機驗收。
"""
from pathlib import Path
import hashlib
import json

r=Path('/src')
work=r/'workplace/cjk-monster-loot-20261003'
record=r/'docs/verification/finish-gui-20261003/receipt.json'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(record)
for name,digest in receipt['attachments'].items():
    assert sha(r/name)==digest,name
for name,digest in read(work/'inputs.json').items():
    assert sha(r/name)==digest,name
assert sum(name.endswith('.go') for name in read(work/'inputs.json'))==325
parent=read(work/'parent-preservation.json')
assert sha(r/parent['parent_receipt'])==parent['parent_sha256']
old=read(r/parent['parent_receipt'])
assert len(old['attachments'])==parent['attachments']==36
for name,digest in old['attachments'].items():
    path=work/'parent-inputs'/name if name in parent['sources'] else r/name
    assert sha(path)==digest,name
assert read(r/old['asset_audit'])['pngs_rebuilt_and_hash_equal']==773
checks=read(work/'checks.json')
assert checks['inspection_states_per_language']==2751
assert checks['all_non_display_fields_equal'] and checks['extension_prefix_equal']
assert checks['name_width']==80 and checks['normal_party']==6
assert len(checks['png_checks'])==3
assert all(row['complete_glyph_pixels'] and row['missing_ink_pixel_rejected'] and row['narrow_width_rejected'] for row in checks['png_checks'])
assert receipt['seed']==1 and receipt['original_seed'] is None
assert receipt['production_go_unchanged']==325
assert receipt['binary_sha256']==checks['current_binary_sha256']=='56bfb4c8d2d83f8df1ee459c080b38c7975ad99d47eeff9c526e35a2a7cb2403'
population=read(work/'ground-population.json')
assert population['candidate_records']==316 and population['maximum']==80
assert population['excluded_example'][0]['kind']==0
for row in receipt['screenshots']:
    assert sha(record.parent/row['file'])==row['png_sha256']
    assert row['source_sha256']=='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46'
    assert row['classification']=='private PNG exception; no public redistribution'
assert len(receipt['screenshots'])==10
build=read(work/'build/checks.json')
assert build['windows_pe_amd64']
assert set(build['macos'])=={'arm64','x86_64'}
assert build['macos']['arm64']['signature']
before=read(work/'failed-cross-python-image/checks.json')
for name in ['kol-windows-amd64.exe','kol-macos-arm64','kol-macos-amd64','kol-macos-universal']:
    assert sha(work/'build'/name)==build['sha256'][name]==before['sha256'][name],name
items=read(r/'docs/worklist.json')['items']
assert sum(item.get('status')=='done' for item in items)==84
assert sum(item.get('status')!='done' for item in items)==22
previous=read(work/'parent-inputs/docs/worklist.json')['items']
assert {item['id']:(item.get('status'),item['acceptance']) for item in items}=={item['id']:(item.get('status'),item['acceptance']) for item in previous}
closeout=read(work/'closeout.json')
assert not closeout['root_owned'] and not closeout['markdown_directories']
assert closeout['linked_records_exist']
print('PASS：2751狀態雙語機制一致；三個文字框完整字模；四個交叉建置重複雜湊一致；325 Go不變；22開放／84完成')
