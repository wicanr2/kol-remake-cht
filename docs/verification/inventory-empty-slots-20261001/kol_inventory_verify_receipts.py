from pathlib import Path
import json,subprocess
root=Path('workplace/inventory-empty-slots-20261001')
out=root/'verifier-final';assert not out.exists();out.mkdir()
reports=[]
for path in sorted(Path('docs/verification/player-path').glob('*.json')):
    if path.name.endswith('-receipt.json'):continue
    plan=json.loads(path.read_text());name=path.stem
    plan['shot_dir']=str(root/'all-after'/name)
    if plan.get('trace'):plan['trace']=str(root/'all-after'/name/'trace.json')
    if plan.get('save'):plan['save']['file']=str(root/'all-after'/name/'save.json')
    cloned=out/(name+'-plan.json');cloned.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    receipt=out/(name+'-receipt.json')
    tool='tools/verify_full_path.py' if plan.get('trace') else 'tools/verify_player_path.py'
    p=subprocess.run(['python3',tool,'--plan',str(cloned),'--out',str(receipt)],capture_output=True,text=True)
    (out/(name+'.log')).write_text(p.stdout+p.stderr)
    d=json.loads(receipt.read_text());reports.append({'name':name,'tool':tool,'returncode':p.returncode,'failures':d['failures']})
    print(name,p.returncode,d['failures'][:2],flush=True)
assert len(reports)==29
(root/'verifier-final-summary.json').write_text(json.dumps({'schema':'kol-player-path-verifier-summary-v1','paths':reports},ensure_ascii=False,indent=2)+'\n')

assert all(x['returncode']==0 and not x['failures'] for x in reports), reports
