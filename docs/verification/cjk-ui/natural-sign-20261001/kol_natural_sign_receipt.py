import pathlib,json,hashlib,copy,shutil,sys,subprocess
sys.path.insert(0,'/src/tools')
from png_io import read_png

root=pathlib.Path('workplace/cjk-sign-natural-20261001')
en=json.loads((root/'replay-en/trace.json').read_text())['steps']
zh=json.loads((root/'replay-zh-Hant/trace.json').read_text())['steps']
assert len(en)==len(zh)==3081
differences=[]; display=0; encounter_display=[]
for i,(a,b) in enumerate(zip(en,zh)):
    a,b=copy.deepcopy(a),copy.deepcopy(b)
    if a.get('battle') is not None and b.get('battle') is not None:
        display+=a['battle'].get('exec_text')!=b['battle'].get('exec_text')
        a['battle'].pop('exec_text',None);b['battle'].pop('exec_text',None)
        if a['battle'].get('encounter_text') != b['battle'].get('encounter_text'):
            assert i in (159,1653),('未登記的遭遇譯文差異',i)
            encounter_display.append({'step':i,'en':a['battle']['encounter_text'],'zh-Hant':b['battle']['encounter_text']})
        a['battle'].pop('encounter_text',None);b['battle'].pop('encounter_text',None)
    if a!=b: differences.append((i,[k for k in set(a)|set(b) if a.get(k)!=b.get(k)]))
assert not differences,differences[:8]
for rows in (en,zh):
    assert len(rows[58]['town']['party'])==6
    assert rows[3079]['town']['town']==5 and rows[3079]['town']['mode']==1
    assert (rows[3079]['town']['x'],rows[3079]['town']['y'])==(40,19)
    assert rows[3080]['town']['mode']==2 and rows[3080]['town']['door']['index']==8

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
source=pathlib.Path('assets/original-derived-png/ega/TWN5PICS/twn5pics-019.png')
pic=read_png(source); raw=pic.rgb()
pixel_checks={}
for lang,scale,x0,y0 in [('en',1,104,58),('zh-Hant',3,312,174)]:
    frame=read_png(root/f'replay-{lang}/step-3079.png');rgb=frame.rgb(); errors=0
    for y in range(pic.height*scale):
        for x in range(pic.width*scale):
            a=((y+y0)*frame.width+x+x0)*3; b=((y//scale)*pic.width+x//scale)*3
            errors+=rgb[a:a+3]!=raw[b:b+3]
    assert errors==0,(lang,errors)
    pixel_checks[lang]=errors
encounters=[]
for i,row in enumerate(en):
    b=row.get('battle')
    if b and b['stage']==1:
        encounters.append({'start_step':i,'monster':b['monster'],'world':row['world']})
    if b and b['stage']==5 and encounters and 'verdict' not in encounters[-1]:
        encounters[-1].update(result_step=i,round=b['round'],verdict=b['verdict'])
assert [(x['monster'],x['round'],x['verdict']) for x in encounters]==[('CLIFF TROLL',18,1),('ORC',13,1)]

target=pathlib.Path('docs/verification/cjk-ui/natural-sign-20261001')
assert not target.exists(),'正式收據目錄已存在，拒絕覆寫'
assert target.parent.stat().st_uid==1000
target.mkdir(); files={}
def save(src,name):
    shutil.copyfile(src,target/name);files[name]=sha(target/name)
for lang in ('en','zh-Hant'):
    save(root/f'replay-{lang}/trace.json',f'{lang}-trace.json')
    save(root/f'replay-{lang}/engine.log',f'{lang}-engine.log')
    save(root/f'replay-{lang}/run.json',f'{lang}-run.json')
save(root/'final-script.txt','script.txt')
save(root/'travel-run.json','travel-run.json')
save(root/'town-navigation.json','town-navigation.json')
for lang,step,name in [('en',3079,'door-en.png'),('zh-Hant',58,'party-zh-Hant.png'),
    ('zh-Hant',159,'cliff-troll-zh-Hant.png'),('zh-Hant',2976,'town-entrance-zh-Hant.png'),
    ('zh-Hant',2977,'town-map-zh-Hant.png'),('zh-Hant',3079,'door-zh-Hant.png'),
    ('zh-Hant',3080,'inside-zh-Hant.png')]:
    save(root/f'replay-{lang}/step-{step:03}.png',name)
for name in ['kol_natural_sign_prepare.py','kol_natural_sign_party.py','kol_natural_sign_drive.sh',
             'kol_natural_world_route.py','kol_natural_sign_town.py','kol_natural_sign_replay.py','kol_natural_sign_receipt.py']:
    save(pathlib.Path('/control')/name,name)
report={'schema':'kol-cjk-natural-sign-reachability-v1','date':'2026-10-01',
    'scope':'固定種子 1、正常標題／名冊／旅店招募、自然遭遇與入城後走到住宅門口；只驗重製可達性及英中語意，不宣稱原版同狀態對拍',
    'status':'remake-normal-player-path-confirmed',
    'goal_status':'#35 圖片告示英文仍可见，翻譯未實作，未結案',
    'source_mapping':{'town':5,'archive':'TWN5PICS','picture':19,'door':8,'position':[40,19],
        'building':'Residence','person':'Chauncey Lawnwatcher','world_entrance':[37,38]},
    'semantic_comparison':{'states_per_language':3081,'differing_display_steps':display,
        'excluded_fields':['battle.exec_text','battle.encounter_text'],
        'display_field_evidence':'cmd/kol/main.go:1273–1283 的顯示文字匯出；遭遇欄位僅允許已登記第 159、1653 步譯文差異',
        'encounter_display_differences':encounter_display,'other_differences':0},
    'pixel_checks':pixel_checks,'encounters':encounters,
    'seed':{'en':1,'zh-Hant':1,'method':'執行前 -seed 1','original':None,
        'original_scope':'本次沒有執行原版亂數對拍，不將相同 remake seed 冒稱原版 parity'},
    'load':None,'ambush_rolls':None,'monsters':None,'quest_sites':None,
    'limitations':'門口的 town.door 仍保留 Session 上一間室內建築，不能把它當 Outdoor.AtDoor；本次另以座標、來源圖片逐像素及正常入屋後 door=8 核實。未改正式程式；未驗本次鏈的住宿保存、原版或平台封包。',
    'git_head':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),
    'binary':{'path':'workplace/cjk-edge-20261001/kol-edge-fixed','sha256':sha(pathlib.Path('workplace/cjk-edge-20261001/kol-edge-fixed'))},
    'toolchain':{'image':'eob-remake-go:1.26.7-ebiten2.9.9-audio','image_id':'sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa'},
    'inputs':{str(p):sha(p) for p in [pathlib.Path('Knights-of-legend.zip'),source,
        pathlib.Path('assets/original-derived-data/world-map.json'),pathlib.Path('assets/original-derived-data/town-buildings.json'),
        pathlib.Path('assets/original-derived-data/character-roster.json'),pathlib.Path('assets/ui-font/manifest.json'),
        pathlib.Path('assets/ui-font/zh-hant-16.png')]},
    'rights':'原作衍生 PNG 僅適用既有 private GitHub 特例；不授權公開發布',
    'visual_review':{'status':'pending','files':['door-en.png','party-zh-Hant.png','cliff-troll-zh-Hant.png',
        'town-entrance-zh-Hant.png','town-map-zh-Hant.png','door-zh-Hant.png','inside-zh-Hant.png']},
    'files':files}
report['goal_status']=report['goal_status'].replace('可见','可見')
(target/'receipt.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'states':len(en),'display':display,'semantic_differences':len(differences),
    'pixels':pixel_checks,'encounters':encounters,'receipt':str(target/'receipt.json')},ensure_ascii=False))
