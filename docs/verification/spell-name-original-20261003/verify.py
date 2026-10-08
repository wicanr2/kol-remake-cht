"""Docker唯讀重驗：專案/src、本工作區/out均唯讀。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(r/'docs/verification/spell-name-original-20261003/receipt.json')
for n,h in receipt['attachments'].items():assert sha(r/n)==h,n
for n,h in receipt['production_go'].items():assert sha(r/n)==h,n
assert len(receipt['production_go'])==325
archive=read(o/'parent-preservation.json');pp=r/archive['parent_receipt']
assert sha(pp)==archive['parent_receipt_sha256']
parent=read(pp);assert len(parent['attachments'])==5701
for n,h in parent['attachments'].items():assert sha(o/'parent-inputs'/n if n in archive['sources'] else r/n)==h,n
for lang in ['en','zh-Hant']:
    report=read(o/f'normal-{lang}-review.json');assert report['status']=='PASS' and report['formal_observer_pngs']==3274
    assert report['name_only_save_change'] and report['cancel_preserves'] and report['reload_item_slots']==34
    for label,count in [('expiry',360),('cancel',197),('save',661),('reload',16),('name-only',195),('cap',208)]:
        pairs=[]
        for mode in ['formal','observed']:
            p=o/f'{mode}-{label}-{lang}';assert len(read(p/'trace.json')['steps'])==count
            shots={q.name:sha(q) for q in p.glob('step-*.png')};assert len(shots)==count;pairs.append(shots)
        assert pairs[0]==pairs[1],(lang,label)
source=read(o/'original-name-review.json');assert source['status']=='PASS' and len(source['source'])==4
assert source['normal_name_fields_same_state'] and len(source['pixels'])==20
assert all(p['input_box_difference']==0 and not p['whole_page_parity'] for p in source['pixels'])
assert not source['whole_page_parity'] and not source['physical_keyboard_test']
assert sha(o/'kol-spell-player-linux')==receipt['formal_sha256']
assert sha(o/'kol-spell-player-observer-linux')==receipt['observer_sha256']
assert receipt['worklist']==dict(open=22,done=84)
print('PASS',len(receipt['attachments']),'attachments; 5701 preserved parent inputs; 325 Go; 6548 matching PNGs; finite original name-input parity')
