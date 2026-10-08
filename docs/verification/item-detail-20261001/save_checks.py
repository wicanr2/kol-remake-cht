"""物品檢視後正常返回旅店住宿保存；比較修正前後的持久資料。"""
from pathlib import Path
import json,hashlib,os,subprocess
root,out=Path('/src'),Path('/out')
assert out.stat().st_uid==os.getuid()
tail='RET > RET RET > > > > RET RET RET'.split()
base='SPACE SPACE 6 RET SPACE > RET > RET D RET'.split()
full=json.loads((out/'normal-after-en/command.json').read_text())
full=full[full.index('-flow-script')+1].split()
rows=[];saves=[]
for phase,binary,keys in [('before',root/'workplace/battle-cursor-20261001/kol-attack-frame-linux',base),
                          ('after',out/'kol-items-linux',full)]:
 for lang in ['en','zh-Hant']:
  dest=out/f'save-{phase}-{lang}';assert not dest.exists();dest.mkdir()
  save=dest/'save.json'
  cmd=[str(binary),'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(keys+tail),
   '-save',str(save),'-shot-dir',str(dest),'-shot-steps',str(len(keys+tail)),'-trace',str(dest/'trace.json')]
  (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
  run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=120)
  (dest/'run.log').write_text(run.stdout+run.stderr)
  assert run.returncode==0,(phase,lang,run.stderr[-1000:])
  assert save.is_file(),(phase,lang,'未產生住宿存檔')
  data=json.loads(save.read_text());print(phase,lang,list(data))
  saved_at=data.pop('saved_at',None)
  saves.append(data)
  rows.append(dict(phase=phase,lang=lang,keys=len(keys+tail),saved_at=saved_at,
   binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),save_sha256=hashlib.sha256(save.read_bytes()).hexdigest()))
assert all(x==saves[0] for x in saves)
(out/'save-comparison.json').write_text(json.dumps(dict(runs=rows,excluded_fields=['saved_at'],equal=True),ensure_ascii=False,indent=2)+'\n')
print('四份正常住宿存檔除saved_at外完全相同。')
