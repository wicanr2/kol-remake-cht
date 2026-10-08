"""十七欄位正常檢視後住宿保存；核對名稱置中修正不影響持久資料。"""
from pathlib import Path
import json,hashlib,os,subprocess
root,out=Path('/src'),Path('/out')
assert out.stat().st_uid==os.getuid() and out.stat().st_gid==os.getgid()
tail=['>']*4+['RET']*3
cmd=json.loads((out/'normal-all-en/command.json').read_text())
keys=cmd[cmd.index('-flow-script')+1].split()+tail
assert len(keys)==73
rows=[];saves=[]
for phase,binary in [('before',out/'kol-items-linux'),('after',out/'kol-items-name-linux')]:
 for lang in ['en','zh-Hant']:
  dest=out/f'save-all-{phase}-{lang}';assert not dest.exists();dest.mkdir()
  save=dest/'save.json'
  cmd=[str(binary),'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(keys),
   '-save',str(save),'-shot-dir',str(dest),'-shot-steps','73','-trace',str(dest/'trace.json')]
  (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
  run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=180)
  (dest/'run.log').write_text(run.stdout+run.stderr)
  assert run.returncode==0,(phase,lang,run.stderr[-1000:])
  assert save.is_file(),(phase,lang,'正常住宿沒有存檔')
  data=json.loads(save.read_text());saved_at=data.pop('saved_at')
  saves.append(data)
  rows.append(dict(phase=phase,lang=lang,keys=73,saved_at=saved_at,
   binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),save_sha256=hashlib.sha256(save.read_bytes()).hexdigest()))
assert len(saves)==4 and all(x==saves[0] for x in saves)
(out/'save-all-comparison.json').write_text(json.dumps(dict(runs=rows,excluded_fields=['saved_at'],equal=True),ensure_ascii=False,indent=2)+'\n')
print('四份正常十七欄位檢視後住宿存檔除saved_at外完全相同。')
