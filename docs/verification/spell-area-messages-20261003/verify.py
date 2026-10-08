from pathlib import Path
import hashlib, json
r=Path('/src')
o=r/'workplace/spell-area-messages-20261003'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(r/'docs/verification/spell-area-messages-20261003/receipt.json')
for name,h in receipt['attachments'].items():
    assert sha(r/name)==h,name
for name,h in receipt['production_go'].items():
    assert sha(r/name)==h,name
evidence_ns={'__name__':'frozen_review'}
body=(o/'review_evidence.py').read_text().split("(o/'evidence-review.json').write_text(",1)[0]
assert 'report=dict(' in body
exec(compile(body,str(o/'review_evidence.py'),'exec'),evidence_ns)
assert evidence_ns['report']==read(o/'evidence-review.json')
normal_ns={'__name__':'frozen_review'}
exec(compile((o/'review_normal.py').read_text(),str(o/'review_normal.py'),'exec'),normal_ns)
assert normal_ns['review']()==read(o/'normal-review.json')
baseline=(o/'baseline-tests.log').read_text()
assert '--- FAIL: TestSpellAreaDamageMessage' in baseline and 'The spell has a markedly Negative effect.' in baseline
for name in ['targeted-tests.log','all-tests.log']:
    text=(o/name).read_text(); assert 'FAIL' not in text and 'github.com/wicanr2/kol-remake-cht/cmd/kol' in text
for name in ['build.log','observer-build.log','gear-prototype-build.log']:
    assert (o/name).read_text()==''
archived=read(o/'parent-preservation.json')['sources']
changed=[name for name,h in receipt['production_go'].items() if name in archived and archived[name]!=h]
added=[name for name in receipt['production_go'] if name not in archived]
assert changed==receipt['changed_go'] and added==['cmd/kol/battle_spell_damage_test.go']
assert receipt['worklist']=={'open':22,'done':84}
print('PASS',len(receipt['attachments']),'attachments;',len(receipt['production_go']),'Go inputs; parent 6803; normal 2 x 632; reload 2 x 16; scope limited to damage message')
