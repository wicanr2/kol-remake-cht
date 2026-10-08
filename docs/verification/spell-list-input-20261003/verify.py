from pathlib import Path
import hashlib, json, runpy
r, o = Path('/src'), Path('/out')
here=r/'docs/verification/spell-list-input-20261003'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(here/'receipt.json')
for name,h in receipt['attachments'].items():
    assert sha(r/name)==h,name
preserved=read(o/'parent-preservation.json')
for name,h in preserved['sources'].items():assert sha(o/'parent-inputs'/name)==h,name
source=(o/'review_input.py').read_text()
source='\n'.join(line for line in source.splitlines() if not line.startswith("(o / 'input-review.json').write_text") and not line.startswith('print('))
ns={'__name__':'readonly_input_review'}
exec(compile(source,str(o/'review_input.py'),'exec'),ns)
assert ns['body']==read(o/'input-review.json')
normal=runpy.run_path(str(o/'review_normal.py'),run_name='readonly_normal_review')['review']()
assert normal==read(o/'normal-review.json')
gloss=runpy.run_path(str(o/'review_gloss.py'),run_name='readonly_gloss_review')['review']()
assert gloss==read(o/'gloss-prototype-review.json')
assert sha(o/'input-spec-ready.md')==read(o/'ready-contract.json')['ready_spec_sha256']
assert sha(o/'kol-spell-list-linux')=='53e335fe9c99891def1f8a66465bd5596e06b1fd078fc3a571eb33a1d9bf25b8'
baseline=(o/'baseline-recheck.log').read_text(); green=(o/'selector-tests.log').read_text()
for name in ['TestSpellListMouseChoosesNegative','TestSpellListCancelReturnsToPreviousLayer']:
    assert '--- FAIL: '+name in baseline
    assert '--- PASS: '+name in green
for name in ['TestSpellListKeyboardSeparatesRowsAndIcons','TestSpellListMouseCancelsBothLayers']:
    assert '--- PASS: '+name in green
assert '\nFAIL' not in (o/'all-tests.log').read_text()
assert len(receipt['production_go'])==319
changed=[name for name,h in receipt['production_go'].items() if preserved['sources'][name]!=h]
assert len(changed)==8,changed
assert receipt['format_decision']=='pending' and receipt['worklist']=={'open':22,'done':84}
print(json.dumps(dict(status='PASS',attachments=len(receipt['attachments']),archived_inputs=len(preserved['sources']),
                      parent_attachments=ns['body']['parent_attachments'],changed_go=changed,
                      normal=[{k:v for k,v in x.items() if k not in ['selection_pixels','mouse']} for x in normal['formal']],scope=normal['scope'],remaining=normal['remaining']),ensure_ascii=False))
