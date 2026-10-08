"""從正常標題核對17個欄位、繞回及退出；不注入角色狀態。"""
from pathlib import Path
import hashlib,json,os,subprocess
root=Path('/src');out=Path('/out')
assert out.stat().st_uid==os.getuid() and out.stat().st_gid==os.getgid()
binary=out/'kol-items-actions-final-linux'
assert hashlib.sha256(binary.read_bytes()).hexdigest()=='c1e3b410380c338d5d1e6c799daee00a75101ea688dfa836a9c94afdc4dd02e7'
original=json.loads((out/'actions-keys.json').read_text())
assert len(original)==104
equivalent={'.':'>','Return':'RET','Down':'D','Up':'U'}
keys=['SPACE','SPACE','6','RET','SPACE']+[equivalent[k] for k in original]
for lang in ['en','zh-Hant']:
 dest=out/('normal-actions-after-'+lang)
 assert not dest.exists();dest.mkdir()
 cmd=[str(binary),'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(keys),
  '-shot-dir',str(dest),'-shot-steps',','.join(map(str,range(len(keys)+1))),'-trace',str(dest/'trace.json')]
 (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
 run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=180)
 (dest/'run.log').write_text(run.stdout+run.stderr)
 assert run.returncode==0,(lang,run.returncode,run.stderr[-1000:])
 states=json.loads((dest/'trace.json').read_text())['steps']
 assert len(states)==110
 print(lang,json.dumps(states[-1],ensure_ascii=False))
