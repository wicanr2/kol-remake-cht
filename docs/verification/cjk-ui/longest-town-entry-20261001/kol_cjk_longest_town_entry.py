from pathlib import Path
import json,hashlib,collections,os,subprocess,time,sys
sys.path.insert(0,'tools')
from build_asset_contact_sheets import read_image
root=Path('workplace/cjk-longest-town-20261001')
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid())
if root.exists():
    assert root.is_dir() and not list(root.iterdir())
    assert (root.stat().st_uid,root.stat().st_gid)==(os.getuid(),os.getgid())
else:
    root.mkdir()
report=json.loads(Path('docs/verification/cjk-text-fit.json').read_text());population=[x for x in report['checks'] if x['kind']=='town-entry']
maximum=max(x['width'] for x in population);assert maximum==240
selected=[x for x in population if x['en']=='docs/i18n/town-zh-Hant.json 城7 門21'];assert max(x['width'] for x in selected)==maximum
town=next(x for x in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns'] if x['town']==7)
door=next(x for x in town['doors'] if x['index']==21);goal=(door['x'],door['y']);assert goal==(20,36)
mother=Path('docs/verification/cjk-ui/natural-sign-20261001');tokens=(mother/'script.txt').read_text().split()[:60]
reference=json.loads((mother/'en-trace.json').read_text())['steps'][60]['town'];assert (reference['x'],reference['y'],reference['mode'])==(57,8,0)
gridpath=Path('assets/original-derived-png/ega/TOWN7/town7-map.png');grid=read_image(gridpath);assert (grid.width,grid.height)==(64,64)
tiles=grid.rgb()[::3];blocked=set(town['blocked_tiles']);start=(57,8);prev={start:None};q=collections.deque([start])
while q:
    p=q.popleft()
    if p==goal:break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        n=(p[0]+dx,p[1]+dy)
        if 0<n[0]<63 and 0<n[1]<63 and n not in prev and tiles[n[1]*64+n[0]] not in blocked:prev[n]=p;q.append(n)
assert goal in prev
path=[];p=goal
while prev[p] is not None:path.append(p);p=prev[p]
path.reverse();current=start;doors={(d['x'],d['y']) for d in town['doors']};dirs={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
for p in path:
    if current in doors and current!=start:tokens+=['>','RET']
    tokens.append(dirs[(p[0]-current[0],p[1]-current[1])]);current=p
doorstep=len(tokens);tokens+=['RET'];insidestep=len(tokens)
(root/'script.txt').write_text(' '.join(tokens)+'\n')
selection={'scope':'正式 town-entry 全母體最寬之一，由標題、正常六人招募與城內方向鍵抵達；不宣稱其他欄位或原版完成',
    'population_count':len(population),'maximum_width':maximum,'selected':selected,'town':7,'door':21,'position':goal,
    'door_step':doorstep,'inside_step':insidestep,'start':start,'walk_moves':len(path),'seed':1,
    'seed_method':'執行前 -seed 1','load':None,'monsters':None,'ambush_rolls':None,
    'report_sha256':hashlib.sha256(Path('docs/verification/cjk-text-fit.json').read_bytes()).hexdigest(),
    'map_sha256':hashlib.sha256(gridpath.read_bytes()).hexdigest()}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for lang in ('en','zh-Hant'):
    out=root/lang;out.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps',f'{doorstep},{insidestep}','-trace',str(out/'trace.json')]
    started=time.monotonic();p=subprocess.run(args,capture_output=True,text=True,timeout=180)
    (out/'engine.log').write_text(p.stdout+p.stderr);(out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-started,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr[-2000:]
    rows=json.loads((out/'trace.json').read_text())['steps'];last=rows[-1]['town'];assert last['town']==7 and last['mode']==2 and last['door']['index']==21,last
    print(lang,len(rows),last,flush=True)
assert json.loads((root/'en/trace.json').read_text())==json.loads((root/'zh-Hant/trace.json').read_text())
