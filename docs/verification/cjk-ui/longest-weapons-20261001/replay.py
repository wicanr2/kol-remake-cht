from pathlib import Path
import json,hashlib,subprocess,os
root=Path('workplace/cjk-weapon-20261001')
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
selection=json.loads((root/'selection.json').read_text())
for p,h in selection['inputs_sha256'].items():assert sha(p)==h,p
assert sha(selection['binary']['path'])==selection['binary']['sha256']
assert root.stat().st_uid==os.getuid()
for r in json.loads((root/'replay-plan.json').read_text()):
    out=root/r['label']
    if out.exists():
        assert json.loads((out/'run.json').read_text())==r
        assert len(json.loads((out/'trace.json').read_text())['steps'])==r['expected_states']
        continue
    out.mkdir()
    args=r['command'];assert args[args.index('-seed')+1]=='1'
    load=args[args.index('-load')+1] if '-load' in args else None
    old=sha(load) if load else None
    p=subprocess.run(args,capture_output=True,text=True,timeout=75)
    (out/'engine.log').write_text(p.stdout+p.stderr)
    assert p.returncode==0,(r['label'],p.stderr)
    assert len(json.loads((out/'trace.json').read_text())['steps'])==r['expected_states']
    if load:assert sha(load)==old
    (out/'run.json').write_text(json.dumps(r,ensure_ascii=False,indent=2)+'\n')
    print(r['label'],r['expected_states'],flush=True)
def compare(a,b,path):
    if path.endswith('.encounter_text') or path.endswith('.exec_text'):return
    if isinstance(a,dict) and isinstance(b,dict):
        assert a.keys()==b.keys(),path
        for k in a:compare(a[k],b[k],path+'.'+k)
    elif isinstance(a,list) and isinstance(b,list):
        assert len(a)==len(b),path
        for i,(x,y) in enumerate(zip(a,b)):compare(x,y,path+f'[{i}]')
    elif a!=b:assert '.encounter_text[' in path or '.exec_text[' in path,(path,a,b)
for family in ['player','player-save','player-reload','enemy']:
    def folder(lang):
        return root/(('player-'+lang+'-'+family.split('-')[1]) if '-' in family else family+'-'+lang)
    compare(json.loads((folder('en')/'trace.json').read_text()),
            json.loads((folder('zh-Hant')/'trace.json').read_text()),family)
data=[]
for lang in ['en','zh-Hant']:
    save=json.loads((root/f'player-{lang}-save.json').read_text());save.pop('saved_at');data.append(save)
assert data[0]==data[1]
fighter=next(c for c in data[0]['characters'] if c['name']=='Fighting Bull')
good=json.loads((root/'player-plan.json').read_text())['good']
assert fighter['in_hand']['name']==good['name'] and fighter['in_hand']['record']==good['record']
source=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots'][4]
assert len(fighter['backpack'])==6
assert {g['record'] for g in fighter['backpack']}==({g['record'] for g in source['backpack']}-{good['record']}|{source['in_hand']['record']})
print('英中非顯示狀態、正常保存與完整物品記錄通過。')
