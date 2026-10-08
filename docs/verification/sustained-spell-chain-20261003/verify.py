"""Docker 唯讀重驗：專案 /src、本輪工作區 /out 都掛唯讀。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(r/'docs/verification/sustained-spell-chain-20261003/receipt.json')
for n,h in receipt['attachments'].items():assert sha(r/n)==h,n
for n,h in receipt['production_go'].items():assert sha(r/n)==h,n
assert len(receipt['production_go'])==323 and not receipt['changed_go']
archive=read(o/'parent-preservation.json');pp=r/archive['parent_receipt']
assert sha(pp)==archive['parent_receipt_sha256']
parent=read(pp)
for n,h in parent['attachments'].items():
    assert sha(o/'parent-inputs'/n if n in archive['sources'] else r/n)==h,('parent',n)
assert len(parent['attachments'])==4356 and len(archive['sources'])==335
report=read(o/'normal-review.json');assert report['status']=='PASS'
assert report['formal_observer_pngs']==4920 and not report['original_same_state']
assert report['original']['capture_after_print_does_not_change_actors_characters_counters_or_seed']
for lang in ['en','zh-Hant']:
    assert read(o/f'normal-{lang}-review.json')['status']=='PASS'
    for label,count in [('expiry',360),('save',657),('reload',16),('name-cancel',197)]:
        pairs=[]
        for mode in ['formal','observed']:
            p=o/f'{mode}-{label}-{lang}'
            assert len(read(p/'trace.json')['steps'])==count
            shots={q.name:sha(q) for q in p.glob('step-*.png')};assert len(shots)==count
            pairs.append(shots)
        assert pairs[0]==pairs[1],(label,lang)
name=read(o/'name-cancel-review.json');assert not name['actual_typing_passed']
assert receipt['worklist']==dict(open=22,done=84)
old=r/'workplace/spell-lifecycle-player-path-20261003'
assert sha(old/'kol-spell-player-linux')==receipt['formal_sha256']
assert sha(old/'kol-spell-player-observer-linux')==receipt['observer_sha256']
print('PASS',len(receipt['attachments']),'attachments; 4356 parent inputs; 323 unchanged Go; 4920 matching formal/observer PNGs; actual name typing remains open')
