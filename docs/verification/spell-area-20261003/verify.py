"""重製正常選部位提示、狀態保持與新保存鏈；原版正常施法仍未知。"""
from pathlib import Path
import argparse,base64,copy,hashlib,json,os,runpy
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/spell-area-20261003'
HERE=ROOT/'docs/verification/spell-area-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def canonical(x):return json.loads(json.dumps(x))
def semantic(rows):
 rows=copy.deepcopy(rows)
 for x in rows:
  for k in ['exec_text','encounter_text']:x.get('battle',{}).pop(k,None)
 return rows
def validate(deep=True):
 p=read(PRIVATE/'parent-preservation.json');sources=p['sources']
 assert len(sources)==1385 and p['parent_attachments']==1843
 for n,d in sources.items():assert sha(PRIVATE/'parent-inputs'/n)==d,n
 parent_path=PRIVATE/'parent-inputs/docs/verification/normal-flee-20261003/receipt.json'
 parent=read(parent_path);assert sha(parent_path)==p['parent_receipt_sha256']
 for n,d in parent['attachments'].items():assert sha(PRIVATE/'parent-inputs'/n if n in sources else ROOT/n)==d,n
 if deep:
  class ArchivedRoot:
   def __truediv__(self,n):return PRIVATE/'parent-inputs'/n if str(n)in sources else ROOT/n
  previous=runpy.run_path(str(PRIVATE/'parent-inputs/docs/verification/normal-flee-20261003/verify.py'))['validate']
  previous.__globals__['ROOT']=ArchivedRoot()
  assert canonical(previous())=={k:v for k,v in parent.items()if k!='attachments'}
  assert '最新續核RE-199'in(ROOT/'CONTEXT.md').read_text()
  assert '20261003-spell-area.md'in(ROOT/'docs/re/RESEARCH-LOG.md').read_text()
  assert 'RE-199：選部位提示的顯示接線勘誤'in(ROOT/'docs/spec/012-combat-spells.md').read_text()
  for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-verify-final.log','未完成 22 條'),('worklist-scan-final.log','0 行沒有')]:
   assert signal in(PRIVATE/filename).read_text(),filename
 bindings=read(PRIVATE/'source-bindings.json')
 assert len(bindings['code_inputs'])==319
 assert bindings['changed_code_inputs']==['cmd/kol/main.go']
 assert bindings['added_code_inputs']==['cmd/kol/battle_area_prompt_test.go']
 for n,d in {**bindings['code_inputs'],**bindings['inputs']}.items():assert sha(ROOT/n)==d,n
 assert read(PRIVATE/'evidence-review.json')['status']=='READY'
 review=read(PRIVATE/'code-change-review.json')
 assert review['status']=='PASS'and review['current_sha256']==sha(ROOT/'cmd/kol/main.go')
 old=(PRIVATE/'formatted-parent-main.go').read_text()
 assert sha(PRIVATE/'formatted-parent-main.go')==review['old_formatted_sha256']
 assert old.count(review['old_block'])==1
 assert old.replace(review['old_block'],review['new_block'],1)==(ROOT/'cmd/kol/main.go').read_text()
 assert '可接受輸入前缺提示'in(PRIVATE/'old-area-negative.log').read_text()
 assert '--- PASS: TestAreaPromptVisibleBeforeItAcceptsNumbers'in(PRIVATE/'area-targeted.log').read_text()
 assert 'github.com/wicanr2/kol-remake-cht/cmd/kol'in(PRIVATE/'rules.log').read_text()
 native=read(PRIVATE/'native-shop2/native-shop-private.json')
 assert native['tool']=='dosgolem'and native['revision']=='968ca774705853276ae55c8d1163c369481ac71f'
 assert native['normal_startup']and not any(native[k]for k in ['direct_entry','state_injection','wait_stub'])
 assert not native['reached_purchase']and native['final_spell_slot_hex']=='00000000000000000000000000000000'
 assert native['rows'][0]['state']['bounds']==[0,48,63,15]
 assert native['rows'][20]['state']['camera']==[61,15]and native['rows'][20]['state']['icon_count']==2
 assert '不認得按鍵 "."'in(PRIVATE/'native-shop1/run.log').read_text()
 for run in ['native-shop1','native-shop2']:
  d=read(PRIVATE/run/'driver.json')
  assert sha(PRIVATE/run/'shop-driver.go')==d['shop_driver_sha256']
  assert sha(PRIVATE/run/'instrumented-book-private.go')==d['startup_sha256']
 spell_record='85958e0b1ee00c814000000006000a22'
 buildings=read(ROOT/'assets/original-derived-data/town-buildings.json')
 hostel=next(x for x in buildings['towns'][-1]['doors']if x['index']==17)
 assert hostel['name']=='The Hostel'and hostel['spells'][2]['record']==spell_record
 outputs=[];traces=[];saves=[]
 item=runpy.run_path(str(ROOT/'docs/verification/unarmed-resolution-20261003/verify.py'))['item']
 for lang in ['en','zh-Hant']:
  old_path=PRIVATE/('normal-'+lang);path=PRIVATE/('fixed-normal-'+lang)
  old=read(old_path/'trace.json')['steps'];rows=read(path/'trace.json')['steps'];traces.append(rows)
  command=read(path/'command.json');keys=command[command.index('-flow-script')+1].split()
  assert len(keys)==303 and len(rows)==len(list(path.glob('*.png')))==304
  assert keys[-3:]==['0','7','2']and not any(k.startswith(('CELL:','ACTOR:'))for k in keys)
  assert semantic(rows)==semantic(old)
  assert rows[:300]==old[:300]and rows[303]==old[303]
  assert all((path/f'step-{i:03d}.png').read_bytes()==(old_path/f'step-{i:03d}.png').read_bytes()for i in list(range(300))+[303])
  assert all((path/f'step-{i:03d}.png').read_bytes()!=(old_path/f'step-{i:03d}.png').read_bytes()for i in [300,301,302])
  purchased=rows[61]['town']['probe_full_party'][0]
  assert purchased['gold']==987 and purchased['order']['name']=='Blue Gem'
  assert purchased['spells']==[dict(name='Kelnalon',effect=6,turn_code=0,power_code=10,guards=[64,0,0,0])]
  assert rows[194]['battle']['probe_seed']==1
  before=rows[300]['battle'];after=rows[303]['battle']
  assert before['prompt']==12 and before['round']==6 and before['probe_seed']==410959018
  assert before==rows[301]['battle']==rows[302]['battle']
  assert before['probe_actors'][0]['Negative']and before['probe_actors'][1]['Chest']==16
  assert after['probe_actors'][1]['Chest']==15 and after['probe_actors'][1]['Damage']==1
  assert after['prompt']==10 and after['probe_seed']==before['probe_seed']and after['probe_rolls']==before['probe_rolls']
  changed=[]
  for i,(a,b)in enumerate(zip(before['probe_actors'],after['probe_actors'])):
   changed.extend((i,k)for k in a if a[k]!=b[k])
  assert sorted(changed)==[(1,'Chest'),(1,'Damage'),(1,'Hurt')],changed
  assert before['probe_actors'][1]['Hurt']==0 and after['probe_actors'][1]['Hurt']==1
  text=' '.join(before['exec_text'])
  if lang=='en':
   assert all(x in text for x in ['Select an area to hurt','1 head','2 chest','6 left leg'])
  else:assert all(x in text for x in ['選擇要傷害的部位','1 頭部','2 胸口','6 左腿'])
  assert before['probe_rolls'][-2:]==[
   dict(N=2,Before=1764518064,After=3397395635,Round=6,Stage=2),
   dict(N=11,Before=3397395635,After=410959018,Round=6,Stage=2)]
  lodging=PRIVATE/('lodging-'+lang);lodgingrows=read(lodging/'trace.json')['steps']
  tokens=read(lodging/'driver-keys.json');save=read(lodging/'complete.json');saves.append(save)
  assert tokens[:303]==keys and lodgingrows[:304]==rows and len(lodgingrows)==len(tokens)+1
  assert len(tokens)<720 and save['schema']=='kol-save-v3'and len(save['characters'])==16
  assert save['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
  byname={m['name']:m for m in save['characters']if m}
  assert byname['Olear']['spells']==purchased['spells']and byname['Olear']['order']==purchased['order']
  assert byname['Olear']['gold']==931
  observed=PRIVATE/('observed-reload-'+lang);formal=PRIVATE/('formal-reload-'+lang)
  reloadrows=read(observed/'trace.json')['steps'];formalrows=read(formal/'trace.json')['steps']
  assert len(reloadrows)==len(formalrows)==16
  plain=copy.deepcopy(reloadrows)
  for x in plain:x.get('town',{}).pop('probe_full_party',None)
  assert plain==formalrows
  assert len(list(formal.glob('*.png')))==len(list(observed.glob('*.png')))==16
  for p in observed.glob('*.png'):assert p.read_bytes()==(formal/p.name).read_bytes()
  final=reloadrows[-1]['town']['probe_full_party'];assert[m['name']for m in final]==['Theodoric','Olear']
  for m in final:
   saved=byname[m['name']]
   assert[base64.b64decode(x).hex()for x in m['carried']]==saved['carried']
   for k in ['id','wounds','gold','place','town_code','building_id','sheet','quests','progress']:
    assert m.get(k,0)==saved.get(k,0),(lang,m['name'],k)
   assert(m.get('spells')or[])==(saved.get('spells')or[]),(lang,m['name'],'spells')
   assert m.get('order')==saved.get('order'),(lang,m['name'],'order')
   for k in ['in_hand','backpack','worn','token']:assert item(m.get(k))==item(saved.get(k)),(lang,m['name'],k)
  outputs.append(dict(language=lang,normal_keys=303,normal_states=304,normal_images=304,
   unchanged_images=301,area_round=6,invalid_keys=[0,7],selected_area=2,target='Theodoric',chest=[16,15],damage=[0,1],
   controlled_seed=1,area_seed=410959018,lodging_keys=len(tokens),lodging_states=len(lodgingrows),
   lodging_images=len(list(lodging.glob('*.png'))),save_gold=[byname[n]['gold']for n in ['Olear','Theodoric']],
   formal_reload_states=16,formal_reload_images=16))
 assert semantic(traces[0])==semantic(traces[1])
 assert{k:v for k,v in saves[0].items()if k!='saved_at'}=={k:v for k,v in saves[1].items()if k!='saved_at'}
 return dict(schema='kol-normal-spell-area-receipt-v1',historical_shared_inputs=1385,parent_attachments=1843,
  parent_results_recomputed=deep,code_inputs=319,changed_code_inputs=bindings['changed_code_inputs'],
  added_code_inputs=bindings['added_code_inputs'],normal=outputs,original_normal_shop=False,original_normal_area=False,
  full_entry_rng_verified=False,status='重製正常選部位問句顯示、數值與保存鏈已驗；原版正常施法及完整重製仍開放。')
def attachments():
 names={str(p.relative_to(ROOT))for p in PRIVATE.rglob('*')if p.is_file()}
 names|={'CONTEXT.md','WORKLIST.md','WORKLOG.md','docs/worklist.json','docs/re/RESEARCH-LOG.md',
  'docs/re/20261003-spell-area.md','docs/spec/012-combat-spells.md','cmd/kol/main.go','cmd/kol/battle_area_prompt_test.go',
  'docs/verification/spell-area-20261003/verify.py'}
 return{n:sha(ROOT/n)for n in sorted(names)}
if __name__=='__main__':
 ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--data',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
 assert sum([args.seal,args.data,args.audit])<=1
 result=validate(not args.data)
 if args.seal:
  target=Path('/out/receipt.json');assert not target.exists();result['attachments']=attachments();target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
 elif not args.data:
  saved=read(HERE/'receipt.json')
  for n,d in saved['attachments'].items():assert sha(ROOT/n)==d,n
  assert canonical(result)=={k:v for k,v in saved.items()if k!='attachments'}
  if args.audit:
   control=read(PRIVATE/'host-control.json');assert control['git_diff_check']=='passed'and control['git_staged']==control['project_containers']==[]
   assert control['head']=='1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
   bad=[];md=[]
   for directory,dirs,files in os.walk(ROOT):
    for name in dirs+files:
     p=Path(directory)/name
     if p.lstat().st_uid==0:bad.append(str(p))
     if p.is_dir()and p.name.endswith('.md'):md.append(str(p))
   assert not bad and not md
   for n in set(saved['attachments'])|set(read(PRIVATE/'source-bindings.json')['code_inputs']):
    st=(ROOT/n).stat();assert(st.st_uid,st.st_gid)==(1000,1000),n
   audit=Path('/out/audit.json');assert not audit.exists()
   audit.write_text(json.dumps(dict(status='PASS',receipt_sha256=sha(HERE/'receipt.json'),attachments_verified=len(saved['attachments']),
    parent_results_recomputed=True,root_owned_files=0,markdown_directories=0,control=control),ensure_ascii=False,indent=2)+'\n')
 print(json.dumps({k:v for k,v in result.items()if k!='attachments'},ensure_ascii=False))
