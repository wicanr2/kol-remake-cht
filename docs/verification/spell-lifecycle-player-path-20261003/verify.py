"""Docker 唯讀重驗：專案 /src、本輪工作區 /out 都掛唯讀。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(r/'docs/verification/spell-lifecycle-player-path-20261003/receipt.json')
for name,h in receipt['attachments'].items():assert sha(r/name)==h,name
for name,h in receipt['production_go'].items():assert sha(r/name)==h,name
for name,h in read(o/'build-inputs.json')['inputs'].items():assert sha(r/name)==h,name
assert len(receipt['production_go'])==323
archive=read(o/'parent-preservation.json');parent_path=r/archive['parent_receipt']
assert sha(parent_path)==archive['parent_receipt_sha256']
parent=read(parent_path)
for name,h in parent['attachments'].items():
    p=o/'parent-inputs'/name if name in archive['sources'] else r/name
    assert sha(p)==h,('parent',name)
assert len(parent['attachments'])==3342
assert read(o/'menu-evidence-review.json')['checked_instructions']==1283
report=read(o/'menu-normal-review.json')
assert report['status']=='PASS' and len(report['cases'])==4
assert report['old_menu_counterexamples']==13 and report['formal_observer_images']==1600
assert receipt['worklist']==dict(open=22,done=84)
assert sha(o/'kol-spell-player-linux')==receipt['formal_sha256']
assert sha(o/'kol-spell-player-observer-linux')==receipt['observer_sha256']
print('PASS',len(receipt['attachments']),'attachments; 3342 preserved parent inputs; 323 Go inputs; 4 bilingual original/remake edit cases; 1600 identical observer/formal PNGs')
