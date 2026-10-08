"""本場正常逃走與新住宿／正式重載的有限驗收；歷史輸入由封存副本重驗。"""
from pathlib import Path
import argparse,base64,copy,hashlib,json,os,runpy
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/normal-flee-20261003'
HERE=ROOT/'docs/verification/normal-flee-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def canonical(v):return json.loads(json.dumps(v))
def validate(deep=True):
 if deep:
  research=(ROOT/'docs/re/20261003-normal-flee.md').read_text()
  for address in ['0x211AD','0x29800','0x29981']:assert address in research,address
  for n in ['011-wilderness-ambush.md','027-battle-exit-wounds.md']:
   assert 'RE-198：本場正常逃走有限CONFORMED'in(ROOT/'docs/spec'/n).read_text(),n
  assert '20261003-normal-flee.md'in(ROOT/'docs/re/RESEARCH-LOG.md').read_text()
  assert '最新續核RE-198'in(ROOT/'CONTEXT.md').read_text()
  for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-verify-final.log','未完成 22 條'),('worklist-scan-final.log','0 行沒有')]:
   assert signal in(PRIVATE/filename).read_text(),filename
 preservation=read(PRIVATE/'parent-preservation.json');sources=preservation['sources']
 assert len(sources)==1380 and preservation['parent_attachments']==2115
 for n,d in sources.items():assert sha(PRIVATE/'parent-inputs'/n)==d,n
 parent=read(PRIVATE/'parent-inputs/docs/verification/unarmed-resolution-20261003/receipt.json')
 assert sha(PRIVATE/'parent-inputs/docs/verification/unarmed-resolution-20261003/receipt.json')==preservation['parent_receipt_sha256']
 for n,d in parent['attachments'].items():assert sha(PRIVATE/'parent-inputs'/n if n in sources else ROOT/n)==d,n
 if deep:
  class ArchivedRoot:
   def __truediv__(self,n):return PRIVATE/'parent-inputs'/n if str(n)in sources else ROOT/n
  previous=runpy.run_path(str(ROOT/'docs/verification/unarmed-resolution-20261003/verify.py'))['validate']
  previous.__globals__['ROOT']=ArchivedRoot()
  assert canonical(previous())=={k:v for k,v in parent.items()if k!='attachments'}
 bindings=read(PRIVATE/'source-bindings.json')
 assert len(bindings['code_inputs'])==318 and bindings['changed_code_inputs']==[]
 for n,d in {**bindings['code_inputs'],**bindings['inputs']}.items():assert sha(ROOT/n)==d,n
 native_path='workplace/battle-exits-20261002/flee4/complete-source-private.json'
 native=read(ROOT/native_path);world_native=read(ROOT/'workplace/battle-exits-20261002/flee4/world-return-private.json')
 assert native['tool']=='dosgolem'and native['revision']=='968ca774705853276ae55c8d1163c369481ac71f'
 assert native['input_sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
 for source in [native,world_native]:
  assert source['normal_startup']and not any(source[k]for k in ['direct_entry','state_injection','wait_stub'])
 assert native['seed']==1 and native['rounds']==1 and len(native['inputs'])==10
 assert [x['key']for x in native['inputs']]==[',','Return','Return']*2+['.','.','Return','Y']
 assert len(native['rolls'])==8 and native['battle_exit_counts']['rolls']==7
 assert all(x['down_party_670a']==x['down_enemy_6a96']==255 and x['raw_bb84']==0 for x in native['exit_control'])
 defeat=read(ROOT/'workplace/battle-exits-20261002/flee3/complete-source-private.json')
 assert defeat['rounds']==5 and len(defeat['inputs'])==193 and not any(x['key']=='Y'for x in defeat['inputs'])
 assert defeat['initial']['actors']==native['initial']['actors']and defeat['initial']['characters']==native['initial']['characters']
 monitored=read(ROOT/'workplace/world-clock-20261002/flee/complete-source-private.json')
 assert not monitored.pop('return_writes');monitored.pop('return_write_method');assert monitored==native
 entered=next(i for i,x in enumerate(world_native['events'])if x['ida']=='0x1E3A8')
 cost=next(x['state']for x in world_native['events'][entered+1:]if x['ida']=='0x29800')
 returned=next(x['state']for x in reversed(world_native['events'])if x['ida']=='0x29981')
 assert [cost[k]for k in ['raw_bb84','raw_12b8_indexed','movement','time']]==[0,10,20,3360]
 assert [returned[k]for k in ['x','y','sub_x','sub_y','time','seed']]==[49,23,4,4,3300,1924036713]
 module=runpy.run_path(str(ROOT/'docs/verification/battle-settlement-20261002/verify.py'))
 actor_diffs=module['actor_diffs'];item=runpy.run_path(str(ROOT/'docs/verification/unarmed-resolution-20261003/verify.py'))['item']
 offsets=[0x30]+list(range(0xa0,0x1a0,16));outputs=[];traces=[];saves=[]
 for lang in ['en','zh-Hant']:
  path=PRIVATE/('normal-'+lang);rows=read(path/'trace.json')['steps'];traces.append(rows)
  command=read(path/'command.json');keys=command[command.index('-flow-script')+1].split()
  old_path=ROOT/('workplace/unarmed-resolution-20261003/normal-'+lang)
  old=read(old_path/'command.json');oldkeys=old[old.index('-flow-script')+1].split()
  assert keys==oldkeys[:119]+'< RET RET < RET RET > > RET Y RET RET RET RET'.split()
  assert len(keys)==133 and not any(k.startswith(('CELL:','ACTOR:'))for k in keys)
  assert len(rows)==len(list(path.glob('*.png')))==134
  assert rows[:120]==read(old_path/'trace.json')['steps'][:120]
  for i in range(120):assert(path/f'step-{i:03d}.png').read_bytes()==(old_path/f'step-{i:03d}.png').read_bytes()
  initial=rows[119]['battle'];assert initial['probe_seed']==1
  assert actor_diffs(native['initial'],initial['probe_actors'])==[]
  assert (rows[125]['battle']['prompt'],rows[128]['battle']['prompt'],rows[129]['battle']['stage'])==(8,9,5)
  assert all(rows[i]['battle']['round']==1 for i in range(119,130))
  assert not any(x.get('battle',{}).get('stage')==4 for x in rows)
  end=rows[129]['battle'];assert(end['verdict'],end['probe_return_scratch'],end['probe_seed'])==(3,0,752224798)
  assert actor_diffs(native['battle_exit_state'],end['probe_actors'])==[]
  assert[(x['N'],x['Before'])for x in end['probe_rolls']]==[(x['n'],x['seed_before'])for x in native['rolls'][:-1]]
  w=rows[130]['world'];assert[w[k]for k in ['x','y','sub_x','sub_y','ticks','probe_return_scratch','probe_world_seed']]==[49,23,4,4,3300,0,1924036713]
  assert w['probe_world_roll_count']==8
  assert(w['probe_world_last_roll']['N'],w['probe_world_last_roll']['Before'])==(native['rolls'][-1]['n'],native['rolls'][-1]['seed_before'])
  assert all(x['world']['ticks']==3360 for x in rows if x.get('battle'))
  party=rows[-1]['town']['probe_full_party']
  for m,rawhex in zip(party,native['post_battle'][0]['state']['characters']):
   raw=bytes.fromhex(rawhex)
   assert[base64.b64decode(x).hex()for x in m['carried']]==[raw[a:a+16].hex()for a in offsets]
   assert m['wounds']==raw[27]==0 and m['gold']==int.from_bytes(raw[42:45],'big')
   assert m['sheet']['adventure_points']==int.from_bytes(raw[38:40],'big')
  lodging=PRIVATE/('lodging-'+lang);lodgingrows=read(lodging/'trace.json')['steps'];save=read(lodging/'complete.json');saves.append(save)
  tokens=read(lodging/'driver-keys.json')
  assert len(tokens)==395 and len(lodgingrows)==396 and lodgingrows[:134]==rows
  assert tokens[:133]==keys and len(list(lodging.glob('*.png')))==35
  assert save['schema']=='kol-save-v3'and len(save['characters'])==16 and save['settings']['faithful_inn_stay']
  assert save['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
  assert save['clock']==dict(ticks=3000,part=1,day=2,month=3,year=0,weather=2)
  assert [(x['step'],x['battle']['verdict'])for x in lodgingrows[134:]if x.get('battle',{}).get('stage')==5]==[(182,3),(232,3)]
  byname={m['name']:m for m in save['characters']if m};beforestay=lodgingrows[382]['town']['probe_full_party']
  assert [(m['name'],i)for m,s in zip(party,beforestay)for i,(a,b)in enumerate(zip(m['carried'],s['carried']))if a!=b]==[]
  for m in beforestay:
   s=byname[m['name']];assert s['carried']==[base64.b64decode(x).hex()for x in m['carried']]
   assert s['wounds']==m['wounds']==0 and s['gold']==m['gold']-(56 if m['name']=='Olear'else 0)
   assert s['sheet']['adventure_points']==m['sheet']['adventure_points']
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
   s=byname[m['name']];assert[base64.b64decode(x).hex()for x in m['carried']]==s['carried']
   for k in ['id','wounds','gold','place','town_code','building_id','sheet','quests','progress']:assert m.get(k,0)==s.get(k,0),(lang,m['name'],k)
   for k in ['in_hand','backpack','worn']:assert item(m.get(k))==item(s.get(k)),(lang,m['name'],k)
  outputs.append(dict(language=lang,normal_keys=133,normal_states=134,normal_images=134,prior_images_identical=120,
   controlled_seed=1,rolls=8,roll_faces_and_seeds_match=True,initial_and_exit_sampled_actor_differences=[],
   world=[49,23,4,4,3300],carried_slots=34,gold=[m['gold']for m in party],adventure_points=[m['sheet']['adventure_points']for m in party],
   lodging_keys=395,lodging_states=396,lodging_images=35,return_route_additional_flee=2,
   save_gold=[byname[m['name']]['gold']for m in party],formal_reload_states=16,formal_reload_images=16))
 for a,b in zip(*traces):
  for k in ['probe_actors','probe_party','probe_rolls','probe_seed','probe_return_scratch']:assert a.get('battle',{}).get(k)==b.get('battle',{}).get(k)
  assert a.get('town',{}).get('probe_full_party')==b.get('town',{}).get('probe_full_party')and a.get('world')==b.get('world')
 assert{k:v for k,v in saves[0].items()if k!='saved_at'}=={k:v for k,v in saves[1].items()if k!='saved_at'}
 return dict(schema='kol-normal-flee-receipt-v1',historical_shared_inputs=1380,parent_attachments=2115,parent_results_recomputed=deep,
  code_inputs=318,changed_code_inputs=[],normal=outputs,full_entry_rng_verified=False,full_actor_record_parity=False,
  status='本場正常逃走、8骰、既定角色抽樣、返回成本及新保存鏈有限CONFORMED；其他玩家分支與完整重製仍開放。')
def attachments():
 names={str(p.relative_to(ROOT))for p in PRIVATE.rglob('*')if p.is_file()}
 names|={'CONTEXT.md','WORKLIST.md','WORKLOG.md','docs/worklist.json','docs/re/RESEARCH-LOG.md','docs/re/20261003-normal-flee.md',
  'docs/spec/011-wilderness-ambush.md','docs/spec/027-battle-exit-wounds.md','docs/verification/normal-flee-20261003/verify.py'}
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
