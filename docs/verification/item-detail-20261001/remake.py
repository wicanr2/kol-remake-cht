"""從正常標題到物品明細；沿用目前程式，不注入玩家狀態。"""
from pathlib import Path
import hashlib,json,os,subprocess
root=Path('/src');out=Path('/out')
assert out.stat().st_uid==os.getuid()
binary=root/'workplace/battle-cursor-20261001/kol-attack-frame-linux'
assert hashlib.sha256(binary.read_bytes()).hexdigest()=='e5be211397bbb2eb11e355a77bbe490cc8a1cd8c618c049679dd97a1456bdc7a'
keys='SPACE SPACE 6 RET SPACE > RET > RET D RET RET D RET RET D RET RET D RET RET D RET SPACE U RET'.split()
for lang in ['en','zh-Hant']:
 dest=out/('normal-full-'+lang)
 assert not dest.exists();dest.mkdir()
 cmd=[str(binary),'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(keys),
  '-shot-dir',str(dest),'-shot-steps',','.join(map(str,range(len(keys)+1))),'-trace',str(dest/'trace.json')]
 (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
 run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=120)
 (dest/'run.log').write_text(run.stdout+run.stderr)
 assert run.returncode==0,(lang,run.returncode,run.stderr[-1000:])
 states=json.loads((dest/'trace.json').read_text())['steps']
 print(lang,json.dumps(states[-1],ensure_ascii=False))
