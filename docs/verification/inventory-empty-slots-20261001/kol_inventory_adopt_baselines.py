from pathlib import Path
import copy,hashlib,json,os,shutil,subprocess,sys
sys.path.insert(0,'tools')
from review_player_path_inventory_change import stripped
def save_digest(doc):
    body=json.dumps({k:v for k,v in doc.items() if k!='saved_at'},ensure_ascii=False,
                    sort_keys=True,separators=(',',':'))
    return hashlib.sha256(body.encode('utf-8')).hexdigest()

root=Path('workplace/inventory-empty-slots-20261001')
review=json.loads((root/'all-review.json').read_text())
inputs=json.loads((root/'all-inputs.json').read_text())
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
for path,digest in inputs['inputs_sha256'].items():assert sha(path)==digest,path
for relative,digest in inputs['baseline_sha256'].items():
    assert sha(Path('docs/verification/player-path')/relative)==digest,relative
weather=json.loads((root/'weather-review.json').read_text())
assert all(r['outside_weather_differences']==0 for r in weather['rows'])
for step in (45,90):
    assert (root/f'weather-control/old/step-{step:03d}.png').read_bytes()==(root/f'weather-control/new/step-{step:03d}.png').read_bytes()
# 比較工具的反面對照：金冠、一般狀態或超出空格數的持物變動都必須拒絕。
from review_player_path_inventory_change import review_trace,review_save
sample=[{'step':0,'town':{'mode':0,'party':['Theodoric gold=1009 carried=9']}}]
for mutated in [
    [{'step':0,'town':{'mode':0,'party':['Theodoric gold=1008 carried=9']}}],
    [{'step':0,'town':{'mode':1,'party':['Theodoric gold=1009 carried=9']}}],
    [{'step':0,'town':{'mode':0,'party':['Theodoric gold=1009 carried=3']}}],
]:
    try:review_trace(sample,mutated,{'Theodoric':5})
    except AssertionError:pass
    else:raise AssertionError('反例被錯誤接受')
save_backup=root/'baseline-saves-before';assert not save_backup.exists();save_backup.mkdir()
updates=[]
for r in review['paths']:
    if not r['changed_steps']:continue
    name=r['name'];planpath=Path('docs/verification/player-path')/(name+'.json')
    plan=json.loads(planpath.read_text());steps=json.loads((root/'all-before'/name/'trace.json').read_text())['steps']
    assert sha(Path(plan['trace']))==r['old_trace_sha256']
    changed_milestones=[]
    for m in plan.get('milestones',[]):
        if 'party' not in m.get('town',{}):continue
        actual=steps[m['step']]['town']['party']
        assert stripped(m['town']['party'])==stripped(actual)
        if m['town']['party']!=actual:
            changed_milestones.append(m['step']);m['town']['party']=actual
    planpath.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
    shutil.copy2(root/'all-before'/name/'trace.json',plan['trace'])
    for c in plan.get('checkpoints',[]):
        source=root/'all-before'/name/c['file'];dest=Path(plan['shot_dir'])/c['file']
        assert source.exists()
        if sha(source)!=sha(dest):assert name=='full-chain-ending' and c['step'] in (45,90)
        shutil.copy2(source,dest)
    if plan.get('save'):
        dest=Path(plan['save']['file']);assert dest.stat().st_uid==os.getuid()
        shutil.copy2(dest,save_backup/dest.name)
        shutil.copy2(root/'all-before'/name/'save.json',dest)
    out=Path('docs/verification/player-path')/(name+'-receipt.json')
    run=subprocess.run(['python3','tools/verify_full_path.py','--plan',str(planpath),'--out',str(out)],capture_output=True,text=True)
    assert run.returncode==0,run.stderr
    manifest_path=Path(plan['shot_dir'])/'build-manifest.json'
    manifest=json.loads(manifest_path.read_text()) if manifest_path.exists() else {'schema':'kol-player-path-build/1','path':name}
    manifest.update({'date':'2026-10-01','base_head':'1384e508f70e8a5d81bc050ea79f66d3f0d38fe6','dirty_worktree':True,
        'build_sha256':inputs['binary_sha256'],
        'evidence':'RE-170初始空物品修正後，正常按鍵／受控輸入與里程碑驗證；不宣稱原版完整對拍',
        'rights':'原作衍生PNG僅限既有private GitHub特例；原始輸入與存檔留本機',
        'toolchain':{'image':'eob-remake-go:1.26.7-ebiten2.9.9-audio','image_id':'sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa','go':'1.26.7','ebitengine':'2.9.9'},
        'plan_sha256':sha(planpath),'trace_sha256':sha(plan['trace']),'receipt_sha256':sha(out),
        'inventory_review':'docs/verification/inventory-empty-slots-20261001/all-path-review.json',
        'previous_baseline':str(root/'baseline-before'/name),
        'run_spec':str(root/'all-before'/name/'shots.spec'),'run_spec_sha256':sha(root/'all-before'/name/'shots.spec'),
        'inputs_sha256':inputs['inputs_sha256'],
        'frames':[{'step':c['step'],'file':c['file'],'sha256':sha(Path(plan['shot_dir'])/c['file'])} for c in plan.get('checkpoints',[])]})
    inp=manifest.setdefault('input',{});inp.update({'source_name':'Knights-of-legend.zip','source_sha256':'2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',
        'character_roster_sha256':inputs['inputs_sha256']['assets/original-derived-data/character-roster.json'],
        'seed':plan.get('seed'),'controlled_ambush':bool(plan.get('ambush_rolls'))})
    if plan.get('ambush_rolls'):inp['ambush_rolls']=plan['ambush_rolls']
    spec=(root/'all-before'/name/'shots.spec').read_text().strip().split('|')
    if '-load' in spec:
        load=spec[spec.index('-load')+1];inp['effective_load']=load;inp['effective_load_sha256']=sha(load)
    if plan.get('save'):manifest['save_sha256_without_saved_at']=save_digest(json.loads(Path(plan['save']['file']).read_text()))
    manifest_path.write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    updates.append({'name':name,'milestones_changed':changed_milestones,'trace_sha256':sha(plan['trace']),
                    'manifest':str(manifest_path),'manifest_sha256':sha(manifest_path)})
# SPEC-023 正式保存已是v3，修正尚在要求v1的住宿驗收條件。
p=Path('docs/verification/player-path/town-rest.json');plan=json.loads(p.read_text())
old_expect=copy.deepcopy(plan['save']['expect']);assert old_expect['schema']=='kol-save-v1'
expected={}
for key,value in old_expect.items():
    if key=='schema':expected[key]='kol-save-v3'
    elif key=='party[0].speaker':expected['speaker']=0
    elif key.startswith('party[0].'):expected[key.replace('party[0].','characters[0].')]=value
    else:expected[key]=value
expected['party[0]']=0
plan['save']['expect']=expected;p.write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
dest=Path(plan['save']['file']);shutil.copy2(dest,save_backup/dest.name);shutil.copy2(root/'all-before/town-rest/save.json',dest)
run=subprocess.run(['python3','tools/verify_player_path.py','--plan',str(p)],capture_output=True,text=True)
assert run.returncode==0,run.stdout+run.stderr
result={'schema':'kol-inventory-baseline-adoption-v1','updates':updates,'review_sha256':sha(root/'all-review.json'),
        'negative_controls_rejected':3,'town_rest_v3_expectations':{'old':old_expect,'new':expected},
        'save_backups_sha256':{p.name:sha(p) for p in save_backup.iterdir()},
        'scope':'只更新經逐欄審查的七條基準與正式v3保存條件；比較器、按鍵、seed及其他條件不改'}
(root/'baseline-adoption.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('七條基準及里程碑已審查更新，舊資料完整備份；正式住宿v3保存驗證通過。')
