"""正常標題／物品動作／住宿，逐階段核對十七欄位並從存檔重載。"""
from pathlib import Path
import json,hashlib,os,subprocess
root,out=Path('/src'),Path('/out')
assert out.stat().st_uid==os.getuid() and out.stat().st_gid==os.getgid()
binary=out/'kol-items-mirror-linux'
assert hashlib.sha256(binary.read_bytes()).hexdigest()=='4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4'
source=json.loads((out/'actions-keys.json').read_text())
mapping={'.':'>','Return':'RET','Down':'D','Up':'U'}
prefix=['SPACE','SPACE','6','RET','SPACE']
close_and_rest=['>','RET','RET']+['>']*4+['RET']*3
rows=[]
for frame in [15,28,42,56,72,88]:
 for lang in ['en','zh-Hant']:
  dest=out/f'save-mirror-{frame:02d}-{lang}';assert not dest.exists();dest.mkdir()
  keys=prefix+[mapping[k] for k in source[:frame+1]]+close_and_rest
  save=dest/'save.json'
  cmd=[str(binary),'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(keys),
   '-save',str(save),'-shot-dir',str(dest),'-shot-steps',str(len(keys)),'-trace',str(dest/'trace.json')]
  (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
  run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=180)
  (dest/'run.log').write_text(run.stdout+run.stderr)
  assert run.returncode==0,(frame,lang,run.stderr[-1000:])
  assert save.is_file(),(frame,lang,'正常住宿沒有存檔')
  rows.append(dict(frame=frame,lang=lang,keys=len(keys),save_sha256=hashlib.sha256(save.read_bytes()).hexdigest()))
  print(frame,lang,'保存完成',flush=True)
(out/'save-mirror-runs-private.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
