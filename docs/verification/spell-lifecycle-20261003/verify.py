"""Docker 唯讀重驗。掛載專案至 /src、本切片工作區至 /out，皆為唯讀。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
read=lambda p:json.loads(p.read_text())
receipt=read(r/'docs/verification/spell-lifecycle-20261003/receipt.json')
for name,h in receipt['attachments'].items():assert sha(r/name)==h,name
for name,h in receipt['production_inputs'].items():assert sha(r/name)==h,name
parent_path=r/receipt['parent_receipt']
assert sha(parent_path)==receipt['parent_receipt_sha256']
parent=read(parent_path);archive=read(o/'parent-preservation.json')
for name,h in parent['attachments'].items():
    p=o/'parent-inputs'/name if name in archive['sources'] else r/name
    assert sha(p)==h,('parent',name)
for name,h in read(o/'build-inputs.json')['inputs'].items():assert sha(r/name)==h,name
assert read(o/'evidence-review.json')['status']==read(o/'normal-review.json')['status']=='PASS'
assert read(o/'evidence-review.json')['original_cases']==21
assert read(o/'evidence-review.json')['checked_instructions']==1313
assert len(receipt['production_go'])==322
assert receipt['worklist']==dict(open=22,done=84)
assert sha(o/'kol-spell-lifecycle-linux')==receipt['formal_sha256']
assert sha(o/'kol-spell-lifecycle-observer-linux')==receipt['observer_sha256']
for name in ['targeted-tests.log','all-tests.log']:
    text=(o/name).read_text()
    assert 'FAIL' not in text and 'ok  ' in text,name
print('PASS',len(receipt['attachments']),'attachments;',len(parent['attachments']),'preserved parent attachments;',len(receipt['production_go']),'Go inputs; 21 original cases; bilingual normal path/save/reload')
