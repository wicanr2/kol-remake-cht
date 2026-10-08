from pathlib import Path
import json,os,hashlib,collections,sys,subprocess
sys.path.insert(0,'tools')
from build_asset_contact_sheets import read_image
root=Path('workplace/cjk-longest-dialog-20261001');assert not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid());root.mkdir()
report=json.loads(Path('docs/verification/cjk-text-fit.json').read_text());pop=[r for r in report['checks'] if r['kind']=='dialog-ui'];maximum=max(r['width'] for r in pop);assert maximum==240
selected=[r for r in pop if r['width']==maximum];assert len(selected)==1 and '0x9d' in selected[0]['en']
town=next(t for t in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns'] if t['town']==7)
door=next(d for d in town['doors'] if d['index']==8);goal=door['x'],door['y'];assert goal==(4,25)
td=next(t for t in json.loads(Path('assets/original-derived-data/town-keywords.json').read_text())['towns'] if t['town']==7)
slots=next(b['slots'] for b in td['buildings'] if b['door']==8);assert sorted(s['category'] for s in slots if s['topic']==0)==[0,1,3]
grid=read_image(Path('assets/original-derived-png/ega/TOWN7/town7-map.png'));tiles=grid.rgb()[::3];blocked=set(town['blocked_tiles']);start=(57,8);prev={start:None};q=collections.deque([start])
while q:
    p=q.popleft()
    if p==goal:break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        n=p[0]+dx,p[1]+dy
        if 0<n[0]<63 and 0<n[1]<63 and n not in prev and tiles[n[1]*64+n[0]] not in blocked:prev[n]=p;q.append(n)
assert goal in prev;path=[];p=goal
while prev[p] is not None:path.append(p);p=prev[p]
path.reverse();tokens=Path('docs/verification/cjk-ui/natural-sign-20261001/script.txt').read_text().split()[:60];current=start;doors={(d['x'],d['y']) for d in town['doors']};dirs={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
for p in path:
    if current in doors and current!=start:tokens+=['>','RET']
    tokens.append(dirs[p[0]-current[0],p[1]-current[1]]);current=p
tokens+=['RET'];entry=len(tokens);tokens+=['SPACE','1'];greeting=len(tokens);cycles=[]
for i in range(3):
    tokens+=['>','>','RET'];prompt=len(tokens);tokens+=['SPACE'];answer=len(tokens);tokens+=['SPACE'];idle=len(tokens)
    cycles.append({'cycle':i+1,'prompt_step':prompt,'answer_step':answer,'idle_step':idle})
shots=[entry,greeting]+[s for c in cycles for s in (c['prompt_step'],c['answer_step'],c['idle_step'])]
selection={'scope':'正式 dialog-ui 全母體最寬問句；從標題正常六人招募、船屋選人與三次聆聽，依固定種子既定順序完整走三則傳聞，不反覆抽取挑選。',
    'population_count':len(pop),'maximum_width':maximum,'selected':selected,'town':7,'door':8,'position':goal,'entry_step':entry,'greeting_step':greeting,'cycles':cycles,'shot_steps':shots,
    'seed':1,'seed_method':'執行前 -seed 1','load':None,'monsters':None,'ambush_rolls':None,'report_sha256':hashlib.sha256(Path('docs/verification/cjk-text-fit.json').read_bytes()).hexdigest()}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n');(root/'script.txt').write_text(' '.join(tokens)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for lang in ('en','zh-Hant'):
    out=root/lang;out.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps',','.join(map(str,shots)),'-trace',str(out/'trace.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=180);(out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n');assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps'];last=rows[-1]['town'];assert last['mode']==2 and last['door']['index']==8
    print(lang,len(rows),cycles,flush=True)
assert json.loads((root/'en/trace.json').read_text())==json.loads((root/'zh-Hant/trace.json').read_text())
