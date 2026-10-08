"""保留前版腳本，另建相同正常輸入的鏡像修正後重播。"""
from pathlib import Path
import json,hashlib,os
p=Path('/out');assert p.stat().st_uid==os.getuid() and p.stat().st_gid==os.getgid()
before='c1e3b410380c338d5d1e6c799daee00a75101ea688dfa836a9c94afdc4dd02e7'
after='4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4'
assert hashlib.sha256((p/'kol-items-mirror-linux').read_bytes()).hexdigest()==after
replacements=[('kol-items-actions-final-linux','kol-items-mirror-linux'),(before,after),
 ('normal-actions-after-','normal-mirror-'),('save-actions-','save-mirror-'),('reload-actions-','reload-mirror-'),
 ('actions-persistence-comparison.json','mirror-persistence-comparison.json')]
records=[]
for src,name in [('remake_actions_after.py','remake_mirror.py'),('save_actions.py','save_mirror.py'),('reload_actions.py','reload_mirror.py')]:
 text=(p/src).read_text()
 for a,b in replacements:text=text.replace(a,b)
 target=p/name;assert not target.exists();target.write_text(text)
 records.append(dict(source=src,source_sha256=hashlib.sha256((p/src).read_bytes()).hexdigest(),path=name,sha256=hashlib.sha256(target.read_bytes()).hexdigest()))
(p/'mirror-run-script-origins.json').write_text(json.dumps(records,ensure_ascii=False,indent=2)+'\n')
print('同一正常鍵序列另建修正後腳本；既有腳本與收據保留。')
