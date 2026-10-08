from pathlib import Path
import json, hashlib, subprocess, os

root=Path('/src')
out=Path('/out')
assert out.stat().st_uid==os.getuid()
plan=json.loads((root/'docs/verification/player-path/quest-chain.json').read_text())
prefix=plan['script'].split()[:417]+['RET']
# 首位 Olear 按 YES，下一位 Poontang：武器→目標→攻擊式→防禦式→主列摘要→移動→摘要→取消→徒手。
tail='> RET D RET > > > RET > > > RET > > > > > RET L RET > RET > RET > RET > > RET RET > > RET > > > RET'.split()
keys=prefix+tail
shots=[417,418,420,421,422,426,430,436,437,438,440,442,444,447,448,451,455]
shots=[n for n in shots if n<=len(keys)]
results=[]
for version,binary in [('before',root/'workplace/platform-smoke-20261001/build/kol-linux'),('after',out/'kol-ring-linux')]:
 for lang in ('en','zh-Hant'):
  dest=out/f'normal2-{version}-{lang}'
  assert not dest.exists()
  dest.mkdir()
  command=[str(binary),'-scene','flow','-seed','1','-lang',lang,
           '-ambush-rolls',plan['ambush_rolls'],'-flow-script',' '.join(keys),
           '-shot-dir',str(dest),'-shot-steps',','.join(map(str,shots)),
           '-trace',str(dest/'trace.json')]
  (dest/'command.json').write_text(json.dumps(command,ensure_ascii=False,indent=2)+'\n')
  run=subprocess.run(command,cwd=root,capture_output=True,text=True,timeout=350)
  (dest/'run.log').write_text(run.stdout+run.stderr)
  assert run.returncode==0,(version,lang,run.returncode,run.stderr[-1200:])
  results.append(dict(version=version,lang=lang,steps=len(keys),binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),trace_sha256=hashlib.sha256((dest/'trace.json').read_bytes()).hexdigest()))
def load(version,lang):return json.loads((out/f'normal2-{version}-{lang}'/'trace.json').read_text())['steps']
for lang in ('en','zh-Hant'):
 a,b=load('before',lang),load('after',lang)
 assert a==b,[(i,x,y) for i,(x,y) in enumerate(zip(a,b)) if x!=y][:2]
assert load('after','en')==load('after','zh-Hant')
prompts={z.get('battle',{}).get('prompt') for z in load('after','en')[418:]}
assert {0,1,2,3,4,5}.issubset(prompts),prompts
summary=dict(schema='kol-target-normal-player-v1',keys=len(keys),runs=results,
             same_states=True,normal_prefix='quest-chain前417個正常按鍵：標題、招募、接任務、步行、答應參戰',
             seed=1,ambush_control=plan['ambush_rolls'],state_injection=False,direct_entry=False,
             original_battle_parity=False,covered_prompts=sorted(prompts),limits=['只驗Poontang近戰／徒手／移動規劃UI；狀態軌跡不含目標游標座標，不宣稱完整規劃狀態／結算對拍'])
(out/'normal2-player-receipt.json').write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(summary,ensure_ascii=False))
