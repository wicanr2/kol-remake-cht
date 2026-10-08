from pathlib import Path
import json, hashlib, os, sys, collections, subprocess
sys.path.insert(0,'tools')
from build_asset_contact_sheets import read_image
root=Path('workplace/cjk-longest-message-20261001')
assert root.is_dir() and not list(root.iterdir()) and (root.stat().st_uid,root.stat().st_gid)==(os.getuid(),os.getgid())
report=json.loads(Path('docs/verification/cjk-text-fit.json').read_text());population=[r for r in report['checks'] if r['kind']=='town-message']
maximum=max(r['width'] for r in population);assert maximum==368
selected=[r for r in population if r['en']=='docs/i18n/town-zh-Hant.json 城7 門17 訊息17'];assert max(r['width'] for r in selected)==maximum
town=next(x for x in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns'] if x['town']==7)
door=next(x for x in town['doors'] if x['index']==17);goal=(door['x'],door['y']);assert goal==(59,28)
tokens=Path('docs/verification/cjk-ui/natural-sign-20261001/script.txt').read_text().split()[:60]
gridpath=Path('assets/original-derived-png/ega/TOWN7/town7-map.png');grid=read_image(gridpath);tiles=grid.rgb()[::3]
blocked=set(town['blocked_tiles']);start=(57,8);prev={start:None};q=collections.deque([start])
while q:
    p=q.popleft()
    if p==goal:break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        n=p[0]+dx,p[1]+dy
        if 0<n[0]<63 and 0<n[1]<63 and n not in prev and tiles[n[1]*64+n[0]] not in blocked:prev[n]=p;q.append(n)
assert goal in prev
path=[];p=goal
while prev[p] is not None:path.append(p);p=prev[p]
path.reverse();current=start;doors={(d['x'],d['y']) for d in town['doors']};dirs={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
for p in path:
    if current in doors and current!=start:tokens+=['>','RET']
    tokens.append(dirs[(p[0]-current[0],p[1]-current[1])]);current=p
doorstep=len(tokens);tokens+=['RET'];entry=len(tokens);tokens+=['SPACE'];who=len(tokens);tokens+=['RET'];farewell=len(tokens);tokens+=['SPACE'];back=len(tokens)
shots=[doorstep,entry,who,farewell,back]
selection={'scope':'正式 town-message 全母體最寬之一；由正常六人招募與第七城旅舍離店告別觸發，不宣稱其他欄位或原版完成。',
    'population_count':len(population),'maximum_width':maximum,'selected':selected,'all_maximum_candidates':[r for r in population if r['width']==maximum],
    'town':7,'door':17,'message':17,'position':goal,'door_step':doorstep,'entry_step':entry,'who_step':who,'farewell_step':farewell,'back_step':back,'shot_steps':shots,
    'walk_moves':len(path),'seed':1,'seed_method':'執行前 -seed 1','load':None,'monsters':None,'ambush_rolls':None,
    'report_sha256':hashlib.sha256(Path('docs/verification/cjk-text-fit.json').read_bytes()).hexdigest()}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n');(root/'script.txt').write_text(' '.join(tokens)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for lang in ('en','zh-Hant'):
    out=root/lang;out.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps',','.join(map(str,shots)),'-trace',str(out/'trace.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=180);(out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps'];assert rows[farewell]['town']['mode']==2 and rows[farewell]['town']['door']['index']==17
    assert rows[-1]['town']['mode']==0 and (rows[-1]['town']['x'],rows[-1]['town']['y'])==goal
    print(lang,len(rows),selection['shot_steps'],rows[-1]['town'],flush=True)
assert json.loads((root/'en/trace.json').read_text())==json.loads((root/'zh-Hant/trace.json').read_text())
