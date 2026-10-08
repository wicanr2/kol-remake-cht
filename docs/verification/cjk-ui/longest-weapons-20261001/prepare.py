from pathlib import Path
import json,csv,hashlib,os
root=Path('workplace/cjk-weapon-20261001')
assert root.parent.is_dir() and not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid())
root.mkdir()
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
names={r['en']:r['zh-Hant'] for r in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
adv={g['character']:g['advance'] for g in fm['glyphs']}
def width(text):
    assert all(c in adv or c==' ' for c in text)
    return sum(adv.get(c,fm['advance']) for c in text)
rows=[]
def add(g,source):
    if g and g.get('kind') in (1,3) and g.get('name'):
        rows.append({'source':source,'name':g['name'],'zh':names[g['name']],
                     'width':width(names[g['name']]),'record':g['record']})
for g in json.loads(Path('assets/original-derived-data/items.json').read_text())['items']:add(g,f'items[{g["index"]}]')
for t in json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns']:
    for d in t['doors']:
        for g in d.get('goods',[]):add(g,f'town{t["town"]}.door{d["index"]}.good{g["index"]}')
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
for c in roster:
    if not c or not c['occupied']:continue
    for field in ('in_hand','equipment','backpack'):
        for i,g in enumerate([c[field]] if field=='in_hand' else c[field]):add(g,f'roster{c["slot"]}.{field}[{i}]')
player_count=len(rows)
monsters=json.loads(Path('assets/original-derived-data/monsters.json').read_text())['monsters']
for m in monsters:
    add(m['weapon'],f'monster{m["index"]}.weapon')
    for i,g in enumerate(m['worn']):add(g,f'monster{m["index"]}.worn[{i}]')
maximum=max(r['width'] for r in rows)
assert player_count==197 and len(rows)==282 and maximum==80
selected=next(m for m in monsters if m['name']=='BINDERAK')
assert names[selected['weapon']['name']]=='棘刺釘頭鎚' and width(names[selected['weapon']['name']])==maximum
site=json.loads(Path('assets/original-derived-data/quest-sites.json').read_text())['sites'][5]
poontang=next(c for c in roster if c and c.get('name')=='Poontang')
assert site['monster']==selected['index']==19 and (site['x'],site['y'])==(48,21)
assert (poontang['quests'][1]//site['quest_divisor'])&3==1
mother=Path('docs/verification/cjk-ui/natural-sign-20261001/script.txt').read_text().split()
scout=mother[:72]+['S']*12+['R']*49
(root/'scout-script.txt').write_text(' '.join(scout)+'\n')
paths=['assets/original-derived-data/items.json','assets/original-derived-data/town-buildings.json',
       'assets/original-derived-data/character-roster.json','assets/original-derived-data/monsters.json',
       'assets/original-derived-data/world-map.json','assets/original-derived-data/quest-sites.json',
       'assets/ui-font/manifest.json','assets/ui-font/zh-hant-16.png',
       'docs/i18n/display-names-zh-Hant.tsv','docs/i18n/newchar-zh-Hant.json','docs/i18n/ui-zh-Hant.json',
       'docs/verification/cjk-ui/natural-sign-20261001/script.txt',
       'internal/battle/equipment_panel.go','internal/ui/battle_cjk.go']
selection={'schema':'kol-normal-longest-weapon-selection-v1','date':'2026-10-01',
    'selected_before_execution':True,'population_count':len(rows),'player_source_count':player_count,
    'monster_source_count':len(rows)-player_count,'maximum_width':maximum,
    'classification':'kind 1／3 的共用物品、商店、有效角色持物及怪物武器／持物；不含玩家自訂名稱',
    'population':rows,'selected':{'monster_index':19,'monster_name':'BINDERAK',
        'weapon_name':selected['weapon']['name'],'zh':names[selected['weapon']['name']],
        'record':selected['weapon']['record'],'width':maximum,'quest':5,'world':[48,21]},
    'seed':1,'load':None,'strategy':'正常標題、六人招募、城門離城，再走到初始 Poontang 已持有的任務5；不改種子、遭遇、座標或 typed data。',
    'initial_quest_flag':{'character':'Poontang','byte_index':1,'raw':4,'divisor':4,'value':1},
    'inputs_sha256':{p:sha(p) for p in paths},
    'binary':{'path':'workplace/platform-smoke-20261001/build/kol-linux',
        'sha256':sha('workplace/platform-smoke-20261001/build/kol-linux')}}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
print('已固定282筆來源，最大80；初始任務5的Binderak，seed1，133個正常輸入。',flush=True)
