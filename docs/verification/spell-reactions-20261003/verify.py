from pathlib import Path
import hashlib, json
r,o=Path('/src'),Path('/out')
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
read=lambda p:json.loads(p.read_text())
receipt=read(r/'docs/verification/spell-reactions-20261003/receipt.json')
assert receipt['schema']=='kol-spell-reactions-20261003-v1'
for name,h in receipt['attachments'].items(): assert sha(r/name)==h,name
for name,h in receipt['production_go'].items(): assert sha(r/name)==h,name
assert len(receipt['production_go'])==320 and len(receipt['production_inputs'])==322
assert sha(r/receipt['parent_receipt'])==receipt['parent_receipt_sha256']
assert sha(o/'kol-spell-reaction-linux')==receipt['formal_sha256']
assert sha(o/'kol-spell-reaction-observer-linux')==receipt['observer_sha256']
ns={}
exec((o/'review_evidence.py').read_text().split("(o/'evidence-review.json').write_text",1)[0],ns)
assert ns['report']==read(o/'evidence-review.json')
assert len(ns['report']['cases'])==8
ns={}
exec((o/'review_normal.py').read_text().split("(o/'normal-review.json').write_text",1)[0],ns)
assert ns['report']==read(o/'normal-review.json')==receipt['normal_review']
draft=(o/'spec-draft.md').read_text()
assert (o/'spec-ready.md').read_text()==draft.replace('# RE-207 DRAFT','# RE-207 READY',1)+'\n證據審查：evidence-review.json PASS。僅授權以上部位法術修正。\n'
assert 'FAIL' in (o/'baseline-tests.log').read_text()
assert 'part:0' in (o/'limb-counterexample.log').read_text()
assert '--- PASS: TestSpellReactionOriginalCases' in (o/'original-compare.log').read_text()
assert (o/'original-compare.log').read_text().count('--- PASS: TestSpellReactionOriginalCases/')==8
assert (o/'targeted-tests.log').read_text().count('--- PASS: TestSpellReactionDamagePageThenKnockback/')==4
assert 'FAIL' not in (o/'all-tests.log').read_text()
work=read(r/'docs/worklist.json')['items']
assert sum(x.get('status')!='done' for x in work)==22 and sum(x.get('status')=='done' for x in work)==84
print('PASS',len(receipt['attachments']),'attachments; 320 Go; 8 original cases; both normal 631-key paths, saves and reloads unchanged')
