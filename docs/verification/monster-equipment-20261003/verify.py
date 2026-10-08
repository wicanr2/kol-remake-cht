#!/usr/bin/env python3
"""Docker /src、/out 唯讀核對；--seal 只首次寫入指定 /seal。"""
from pathlib import Path
import argparse,contextlib,hashlib,io,json,runpy
ROOT=Path(__file__).resolve().parents[3]
PRIVATE=ROOT/'workplace/monster-equipment-20261003'
HERE=Path(__file__).resolve().parent
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def validate():
 preservation=read(PRIVATE/'parent-preservation.json')
 parent=ROOT/preservation['parent_receipt']
 assert sha(parent)==preservation['parent_receipt_sha256']
 prior=read(parent);sources=preservation['sources']
 assert len(sources)==330
 for name,h in sources.items():assert sha(PRIVATE/'parent-inputs'/name)==h,name
 for name,h in prior['attachments'].items():
  assert sha(PRIVATE/'parent-inputs'/name if name in sources else ROOT/name)==h,name
 before=read(PRIVATE/'production-inputs-before.json')
 assert before==prior['code_inputs'] and len(before)==319
 assert all(sha(ROOT/name)==h for name,h in before.items())
 for name in ['tools/extract_monsters.py','assets/original-derived-data/monsters.json']:
  assert sha(ROOT/name)==sources[name]
 assert sha(ROOT/'workplace/ambush-placement-20261003/kol-placement-linux')==prior['formal_sha256']
 assert prior['formal_sha256']=='eed01abbbff335e4a0944d670a61c668d0be5ce5a93fbea12f2e9561e273b043'
 # 原研究腳本的寫入只收集在記憶體，禁止修改或重封原始收據。
 captured={};original=Path.write_text
 def capture(path,text,*args,**kwargs):
  assert str(path)in['/out/evidence-review.json','/out/prototype-check.json'],str(path)
  captured[path.name]=json.loads(text);return len(text)
 try:
  Path.write_text=capture
  with contextlib.redirect_stdout(io.StringIO()):
   runpy.run_path(str(PRIVATE/'review.py'))
   runpy.run_path(str(PRIVATE/'check_prototype.py'))
 finally:Path.write_text=original
 for name,body in captured.items():assert body==read(PRIVATE/name),name
 assert set(captured)=={'evidence-review.json','prototype-check.json'}
 evidence=captured['evidence-review.json'];prototype=captured['prototype-check.json']
 assert evidence['checked_instructions']==1003 and evidence['source_variants']==128
 assert not evidence['production_ready'] and not prototype['production']
 inputs=read(PRIVATE/'prototype-inputs.json')
 for name,h in inputs['files'].items():assert sha(PRIVATE/'prototype-inputs'/name)==h
 assert 'PASS: TestEquipmentPrototypeNormalOriginal' in (PRIVATE/'prototype-targeted.log').read_text()
 assert 'FAIL' not in (PRIVATE/'prototype-targeted.log').read_text()
 names=read(PRIVATE/'display-names-audit.json')
 assert names['source_variants']==128 and len(names['missing_current'])==7 and not names['missing_proposed']
 assert sha(ROOT/names['table'])==names['table_sha256']
 assert sha(PRIVATE/'display-names-proposed.tsv')==names['proposed_table_sha256']
 deep=read(PRIVATE/'parent-check.json')
 assert deep['deep_parent_passed'] and deep['parent_attachments']==1557
 assert deep['parent_receipt_sha256']==sha(parent)
 return dict(schema='kol-monster-equipment-prototype-v1',production=False,formal_sha256=prior['formal_sha256'],production_inputs=before,changed_production_inputs=[],parent_receipt_sha256=sha(parent),parent_attachments=1557,deep_parent_passed=True,archived_sources=330,normal_original_states=175,checked_instructions=1003,source_variants=128,selected_variants=[2,0,3],prototype=prototype,scope='原版裝備選擇已證實，本場首次提示隔離原型比較通過；格式尚待決定，未接正式路徑，不結案完整remake')

def main():
 parser=argparse.ArgumentParser();parser.add_argument('--seal',action='store_true');args=parser.parse_args()
 result=validate()
 if args.seal:
  path=Path('/seal/receipt.json');assert not path.exists()
  names=['CONTEXT.md','WORKLOG.md','WORKLIST.md','docs/worklist.json','docs/re/RESEARCH-LOG.md','docs/re/20261003-monster-equipment.md','docs/verification/monster-equipment-20261003/verify.py','docs/verification/monster-equipment-20261003/audit.py']
  attachments={name:sha(ROOT/name)for name in names}
  attachments.update(result['production_inputs'])
  for p in sorted(PRIVATE.rglob('*')):
   if p.is_file():attachments[str(p.relative_to(ROOT))]=sha(p)
  result['attachments']=attachments
  path.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
 else:
  stored=read(HERE/'receipt.json')
  for name,h in stored['attachments'].items():assert sha(ROOT/name)==h,name
  assert result=={k:v for k,v in stored.items()if k!='attachments'}
 print(json.dumps({k:v for k,v in result.items()if k not in['production_inputs','prototype','attachments']},ensure_ascii=False))
if __name__=='__main__':main()
