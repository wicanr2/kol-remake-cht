"""在Docker內唯讀重驗；專案掛/src，RE-213工作區掛/out。"""
from pathlib import Path
import hashlib,json
r,o=Path('/src'),Path('/out')
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
d=read(r/'docs/verification/spell-edit-inn-save-20261003/receipt.json')
for n,h in d['attachments'].items():assert sha(r/n)==h,n
for n,h in d['production_go'].items():assert sha(r/n)==h,n
assert len(d['production_go'])==325
a=read(o/'parent-preservation.json');parent=read(r/a['parent_receipt'])
assert sha(r/a['parent_receipt'])==a['parent_receipt_sha256']
assert len(parent['attachments'])==9313 and len(a['sources'])==338
for n,h in parent['attachments'].items():assert sha(o/'parent-inputs'/n if n in a['sources'] else r/n)==h,n
assert sha(o/'kol-spell-player-linux')==d['formal_sha256']
assert sha(o/'kol-spell-player-observer-linux')==d['observer_sha256']
assert read(o/'original-save-review.json')['status']=='PASS finite original spell edit/inn save/reload'
assert read(o/'remake-save-review.json')['formal_observer_pngs']==1584
for lang in ['en','zh-Hant']:
 for label,count in [('save',380),('reload',16)]:
  pairs=[]
  for mode in ['formal','observed']:
   p=o/f'{mode}-{label}-{lang}';assert len(read(p/'trace.json')['steps'])==count
   images={q.name:sha(q) for q in p.glob('step-*.png')};assert len(images)==count;pairs.append(images)
  assert pairs[0]==pairs[1]
assert not d['whole_battle_parity'] and not d['whole_page_parity']
assert d['worklist']==dict(open=22,done=84)
print('PASS',len(d['attachments']),'attachments; 9313 preserved parent attachments; 325 unchanged Go; original normal spell edit/inn save/reload; 1584 paired remake PNGs')
