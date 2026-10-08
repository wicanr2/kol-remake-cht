from pathlib import Path
import json, os, hashlib, collections, sys, subprocess, time
sys.path.insert(0, 'tools')
from build_asset_contact_sheets import read_image
root=Path('workplace/inventory-empty-slots-20261001/normal')
assert not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid())
root.mkdir()
def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
report=json.loads(Path('docs/verification/cjk-text-fit.json').read_text())
population=[r for r in report['checks'] if r['kind']=='display-name-shop-list']
maximum=max(r['width'] for r in population)
selected=[r for r in population if r['width']==maximum]
assert maximum==112 and len(selected)==1 and selected[0]['en']=='Girak Helbak'
town=next(t for t in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns'] if t['town']==7)
door=next(d for d in town['doors'] if d['index']==13)
good=next(g for g in door['goods'] if g['name']==selected[0]['en'])
assert good['index']==5 and good['kind']==0 and good['position']==10
gridpath=Path('assets/original-derived-png/ega/TOWN7/town7-map.png')
tiles=read_image(gridpath).rgb()[::3];blocked=set(town['blocked_tiles'])
tokens=Path('docs/verification/cjk-ui/natural-sign-20261001/script.txt').read_text().split()[:60]
start=(57,8);goal=(door['x'],door['y']);prev={start:None};queue=collections.deque([start])
while queue:
    p=queue.popleft()
    if p==goal:break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        n=p[0]+dx,p[1]+dy
        if 0<n[0]<63 and 0<n[1]<63 and n not in prev and tiles[n[1]*64+n[0]] not in blocked:prev[n]=p;queue.append(n)
assert goal in prev
path=[];p=goal
while prev[p] is not None:path.append(p);p=prev[p]
path.reverse();current=start;doors={(d['x'],d['y']) for d in town['doors']};dirs={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
for p in path:
    if current in doors and current!=start:tokens+=['>','RET']
    tokens.append(dirs[p[0]-current[0],p[1]-current[1]]);current=p
steps={'door':len(tokens)}
def add(label,keys):
    tokens.extend(keys);steps[label]=len(tokens)
add('entry',['RET'])
add('greeting',['SPACE','1'])
add('list',['>']*4+['RET'])
add('selected',['D']*good['index'])
add('shop_detail',['RET'])
add('eat_question',['RET'])
add('packed',['>','RET'])
add('end_shop',['>','RET'])
add('sheet',['>','RET'])
add('items',['>','RET'])
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())
leader=roster['slots'][0]
pack_count=sum(g is not None for g in leader['backpack'])
assert leader['name']=='Theodoric' and pack_count==0
add('pocket',['D']*(11+pack_count))
add('item_detail',['RET'])
add('leave_tavern',['RET','>','RET','RET','RET','SPACE'])
current=goal
for p in [*reversed(path[:-1]),start]:
    if current in doors and current!=goal:tokens+=['>','RET']
    tokens.append(dirs[p[0]-current[0],p[1]-current[1]]);current=p
add('inn_entry',['RET'])
add('inn_greeting',['SPACE','1'])
add('rest_whole',['>']*4+['RET'])
add('rest_price',['>','RET'])
add('saved',['RET'])
selection={'scope':'最長資料顯示名稱112像素的正常酒館購買與背包檢視；只驗商品清單、商品明細與背包小框／明細，不把酒館食品當武器。',
 'population_count':len(population),'maximum_width':maximum,'selected':selected,'town':7,'door':13,'position':goal,'good':good,
 'steps':steps,'shot_steps':list(steps.values()),'walk_moves':len(path),'seed':1,'seed_method':'執行前 -seed 1', 'load':None,'monsters':None,'ambush_rolls':None,
 'inputs_sha256':{str(p):sha(p) for p in [Path('docs/verification/cjk-text-fit.json'),Path('docs/i18n/display-names-zh-Hant.tsv'),Path('assets/original-derived-data/town-buildings.json'),Path('assets/original-derived-data/character-roster.json'),gridpath,Path('docs/verification/cjk-ui/natural-sign-20261001/script.txt')]}}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
(root/'script.txt').write_text(' '.join(tokens)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
assert sha(binary)=='f969a63f195f29561b32f8bef333d0933d4d4d69641025981249d127017891d4'
for lang in ('en','zh-Hant'):
    out=root/lang;out.mkdir()
    args=[binary,'-scene','flow','-seed','1','-lang',lang,'-save',str(out/'save.json'),'-flow-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps',','.join(map(str,steps.values())),'-trace',str(out/'trace.json')]
    started=time.monotonic();p=subprocess.run(args,capture_output=True,text=True,timeout=180)
    (out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-started,'binary_sha256':sha(binary)},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps']
    assert all(r['scene']=='flow:town' for r in rows[17:])
    assert rows[-1]['town']['door']['index']==5
    changes=[(r['step'],r['town']['party'][0]) for r in rows if 'party' in r.get('town',{})]
    assert any('gold=974' in text for step,text in changes),changes
    saved=json.loads((out/'save.json').read_text())
    (out/'save-summary.json').write_text(json.dumps({'keys':list(saved)},ensure_ascii=False,indent=2)+'\n')
    print(lang,len(rows),steps,changes[-3:],flush=True)
assert json.loads((root/'en/trace.json').read_text())==json.loads((root/'zh-Hant/trace.json').read_text())
print('英中全部狀態相同',len(tokens)+1,flush=True)
