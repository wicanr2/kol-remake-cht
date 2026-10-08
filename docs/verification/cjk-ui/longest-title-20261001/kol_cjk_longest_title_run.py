from pathlib import Path
import json,hashlib,collections,sys,os,subprocess,time,copy
sys.path.insert(0,'tools')
from build_asset_contact_sheets import read_image
root=Path('workplace/cjk-longest-title-20261001');assert not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid());root.mkdir()
mother=Path('docs/verification/cjk-ui/natural-sign-20261001')
report=json.loads(Path('docs/verification/cjk-text-fit.json').read_text());population=[r for r in report['checks'] if r['kind']=='town-title'];maximum=max(r['width'] for r in population);assert maximum==160
all_max=[r for r in population if r['width']==maximum];assert len(all_max)==6
selected=next(r for r in all_max if r['en']=='docs/i18n/town-zh-Hant.json 城5 門9 person')
town=next(t for t in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns'] if t['town']==5)
door=next(d for d in town['doors'] if d['index']==9);goal=door['x'],door['y'];start=(40,19)
tokens=(mother/'script.txt').read_text().split();assert len(tokens)==3080
reference=json.loads((mother/'en-trace.json').read_text())['steps'];assert reference[-1]['town']['mode']==2 and reference[-1]['town']['door']['index']==8
tokens+=['SPACE','RET','SPACE'];leftstep=len(tokens)
gridpath=Path('assets/original-derived-png/ega/TOWN5/town5-map.png');grid=read_image(gridpath);tiles=grid.rgb()[::3];blocked=set(town['blocked_tiles']);prev={start:None};q=collections.deque([start])
while q:
    p=q.popleft()
    if p==goal:break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        n=p[0]+dx,p[1]+dy
        if 0<n[0]<63 and 0<n[1]<63 and n not in prev and tiles[n[1]*64+n[0]] not in blocked:prev[n]=p;q.append(n)
assert goal in prev;path=[];p=goal
while prev[p] is not None:path.append(p);p=prev[p]
path.reverse();current=start;doors={(d['x'],d['y']) for d in town['doors']};directions={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
for p in path:
    if current in doors and current!=start:tokens+=['>','RET']
    tokens.append(directions[p[0]-current[0],p[1]-current[1]]);current=p
doorstep=len(tokens);tokens+=['RET'];inside=len(tokens);shots=[3080,leftstep,doorstep,inside]
selection={'scope':'正式 town-title 204行全母體最寬之一；標題、正常六人招募、自然旅行、入第五城住宅，再正常離店步行到商行。不宣稱原版或其他動態欄位完成。',
    'population_count':len(population),'maximum_width':maximum,'all_maximum_candidates':all_max,'selected':selected,'town':5,'door':9,'position':goal,
    'door_name':door['name'],'person':door['person'],'picture':door['picture'],'portrait':door['portrait'],'mother_last_step':3080,'left_step':leftstep,'door_step':doorstep,'inside_step':inside,'shot_steps':shots,
    'walk_moves_after_residence':len(path),'seed':1,'seed_method':'執行前 -seed 1','load':None,'monsters':None,'ambush_rolls':None,
    'inputs_sha256':{str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in [mother/'receipt.json',mother/'script.txt',Path('docs/verification/cjk-text-fit.json'),gridpath,Path('assets/original-derived-data/town-buildings.json')]}}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n');(root/'script.txt').write_text(' '.join(tokens)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed';assert hashlib.sha256(Path(binary).read_bytes()).hexdigest()=='f969a63f195f29561b32f8bef333d0933d4d4d69641025981249d127017891d4'
for lang in ('en','zh-Hant'):
    folder=root/lang;folder.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),'-shot-dir',str(folder),'-shot-steps',','.join(map(str,shots)),'-trace',str(folder/'trace.json')]
    print('開始',lang,len(tokens),'個正常按鍵',goal,flush=True);started=time.monotonic()
    p=subprocess.run(args,capture_output=True,text=True,timeout=540);(folder/'engine.log').write_text(p.stdout+p.stderr)
    (folder/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-started,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n');assert p.returncode==0,p.stderr
    rows=json.loads((folder/'trace.json').read_text())['steps'];ref=json.loads((mother/f'{lang}-trace.json').read_text())['steps'];assert rows[:len(ref)]==ref
    assert rows[leftstep]['town']['mode']==0 and (rows[leftstep]['town']['x'],rows[leftstep]['town']['y'])==start
    last=rows[-1]['town'];assert last['town']==5 and last['mode']==2 and last['door']['index']==9 and (last['x'],last['y'])==goal,last
    print('完成',lang,len(rows),'筆狀態',round(time.monotonic()-started,2),'秒',flush=True)
en=json.loads((root/'en/trace.json').read_text())['steps'];zh=json.loads((root/'zh-Hant/trace.json').read_text())['steps'];assert len(en)==len(zh)
differing=[]
for a,b in zip(en,zh):
    ac,bc=copy.deepcopy(a),copy.deepcopy(b)
    for row in (ac,bc):
        if 'battle' in row:
            row['battle'].pop('exec_text',None);row['battle'].pop('encounter_text',None)
    assert ac==bc,(a['step'],ac,bc)
    if a!=b:differing.append(a['step'])
(root/'comparison.json').write_text(json.dumps({'states':len(en),'identical_mother_prefix_states':3081,'differing_display_steps':differing,'excluded_fields':['battle.exec_text','battle.encounter_text'],'other_differences':0},ensure_ascii=False,indent=2)+'\n')
print('英中其他狀態相同',len(en),'显示差異',len(differing),flush=True)
