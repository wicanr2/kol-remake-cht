"""Docker唯讀重驗：專案/src、本工作區/out均唯讀。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(r/'docs/verification/spell-name-input-20261003/receipt.json')
for n,h in receipt['attachments'].items():assert sha(r/n)==h,n
for n,h in receipt['production_go'].items():assert sha(r/n)==h,n
assert len(receipt['production_go'])==324
archive=read(o/'parent-preservation.json');pp=r/archive['parent_receipt'];assert sha(pp)==archive['parent_receipt_sha256']
parent=read(pp);assert len(parent['attachments'])==9313
for n,h in parent['attachments'].items():assert sha(o/'parent-inputs'/n if n in archive['sources'] else r/n)==h,n
for lang in ['en','zh-Hant']:
    report=read(o/f'normal-{lang}-review.json');assert report['status']=='PASS' and report['formal_observer_pngs']==2468
    assert report['name_only_save_change'] and report['cancel_preserves'] and report['reload_item_slots']==34
    for label,count in [('expiry',360),('cancel',197),('save',661),('reload',16)]:
        pairs=[]
        for mode in ['formal','observed']:
            p=o/f'{mode}-{label}-{lang}';assert len(read(p/'trace.json')['steps'])==count
            shots={q.name:sha(q) for q in p.glob('step-*.png')};assert len(shots)==count;pairs.append(shots)
        assert pairs[0]==pairs[1],(lang,label)
assert sha(o/'kol-spell-player-linux')==receipt['formal_sha256']
assert sha(o/'kol-spell-player-observer-linux')==receipt['observer_sha256']
assert receipt['worklist']==dict(open=22,done=84) and not receipt['original_name_same_state']
print('PASS',len(receipt['attachments']),'attachments; 9313 preserved parent inputs; 324 Go; 4936 matching PNGs; limited name-input wiring')
