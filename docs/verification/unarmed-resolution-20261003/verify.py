"""徒手結算、本場正常勝利及新住宿重載的有限唯讀驗收。"""
from pathlib import Path
import argparse,base64,copy,hashlib,json,os,runpy
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/unarmed-resolution-20261003'
HERE=ROOT/'docs/verification/unarmed-resolution-20261003'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def canonical(value):return json.loads(json.dumps(value))
def item(value):
 if value is None:return None
 if isinstance(value,list):return[item(v)for v in value]
 value=copy.deepcopy(value);value.setdefault('picture',0);value.setdefault('movement',0);return value
def validate():
 research=(ROOT/'docs/re/20261003-unarmed-resolution.md').read_text()
 for address in ['0x26FCA','0x26C2D','0x27579','0x27648','0x2736B','0x2755E']:assert address in research,address
 for name,marker in [('009-combat-attack.md','RE-197：有限CONFORMED'),('011-wilderness-ambush.md','RE-197：本場徒手差異閉合'),('027-battle-exit-wounds.md','RE-197：本場徒手差異閉合')]:
  assert marker in(ROOT/'docs/spec'/name).read_text(),name
 preservation=read(PRIVATE/'parent-preservation.json');sources=preservation['sources'];assert len(sources)==1375
 for name,digest in sources.items():assert sha(PRIVATE/'parent-inputs'/name)==digest,name
 parent=read(PRIVATE/'parent-inputs/docs/verification/normal-victory-20261003/receipt2.json')
 assert len(parent['attachments'])==5343
 for name,digest in parent['attachments'].items():assert sha(PRIVATE/'parent-inputs'/name if name in sources else ROOT/name)==digest,name
 class ArchivedRoot:
  def __truediv__(self,name):return PRIVATE/'parent-inputs'/name if str(name)in sources else ROOT/name
 previous=runpy.run_path(str(ROOT/'docs/verification/normal-victory-20261003/verify_final.py'))['validate']
 previous.__globals__['ROOT']=ArchivedRoot()
 assert canonical(previous())=={k:v for k,v in parent.items()if k!='attachments'}
 built=read(PRIVATE/'production-inputs.json');prior=read(ROOT/'workplace/normal-victory-20261003/save-production-inputs.json')
 assert len(built['code_inputs'])==318
 for name,digest in built['code_inputs'].items():assert sha(ROOT/name)==digest,name
 assert [n for n in built['code_inputs']if built['code_inputs'][n]!=prior['code_inputs'][n]]==built['changed_inputs']==['cmd/kol/main.go','internal/battle/attack.go','internal/battle/attack_test.go']
 assert sha(PRIVATE/'kol-unarmed-linux')==built['production_sha256']
 assert sha(PRIVATE/'kol-unarmed-observer-linux')==built['observer_sha256']
 review=read(PRIVATE/'evidence-review.json');ida=read(PRIVATE/'ida-unarmed.json')
 assert review['checked_instructions']==1045 and sha(PRIVATE/'ida-unarmed.json')==review['ida_json_sha256']
 assert ida['ida_version']=='9.4'and ida['input_sha256']==review['input_sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
 exe=(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked.EXE').read_bytes();assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked.EXE')==review['input_sha256']
 loaded=bytearray(exe);table=int.from_bytes(exe[24:26],'little');header=int.from_bytes(exe[8:10],'little')*16
 for i in range(int.from_bytes(exe[6:8],'little')):
  p=table+4*i;target=header+int.from_bytes(exe[p+2:p+4],'little')*16+int.from_bytes(exe[p:p+2],'little')
  loaded[target:target+2]=((int.from_bytes(loaded[target:target+2],'little')+0x1000)&0xffff).to_bytes(2,'little')
 assert header==9216
 flat=[x for f in ida['functions']for x in f['listing']]
 assert len(flat)==len(review['checks'])==1045
 for instruction,check in zip(flat,review['checks']):
  offset=int(check['file_offset'],16);value=bytes.fromhex(instruction['bytes'])
  assert int(instruction['ea'],16)-0x10000+header==offset
  assert loaded[offset:offset+len(value)]==value
  assert exe[offset:offset+len(value)].hex()==check['file_bytes']
 assert '第 2 次擲骰面數 3'in(PRIVATE/'old-unarmed-negative.log').read_text()
 assert 'FAIL'not in(PRIVATE/'rules.log').read_text()
 for name in ['TestResolveUnarmedOriginalBranches','TestResolveUnarmedMissDoesNotRollStyle','TestUnarmedBrokenArmourRetainsZeroCodeDie','TestResolveRandomStyleOrder']:
  assert '--- PASS: '+name in(PRIVATE/'targeted.log').read_text()
 native_path=ROOT/'workplace/battle-victory-20261002/source4/complete-source-private.json';native=read(native_path)
 assert sha(native_path)==review['source_sha256']
 assert native['normal_startup']and not any(native[k]for k in ['direct_entry','state_injection','wait_stub'])
 assert native['seed']==1 and native['rounds']==4 and len(native['inputs'])==147
 assert bytes.fromhex(native['initial']['characters'][0])[0x150:0x160]==bytes(16)
 actor_diffs=runpy.run_path(str(ROOT/'docs/verification/battle-settlement-20261002/verify.py'))['actor_diffs']
 offsets=[0x30]+list(range(0xa0,0x1a0,16));outputs=[];traces=[];saves=[]
 for lang in ['en','zh-Hant']:
  path=PRIVATE/('normal-'+lang);rows=read(path/'trace.json')['steps'];traces.append(rows)
  old=ROOT/('workplace/normal-victory-20261003/normal3-'+lang)
  command=read(path/'command.json');keys=command[command.index('-flow-script')+1].split()
  prior_command=read(old/'command.json');assert keys==prior_command[prior_command.index('-flow-script')+1].split()
  assert len(keys)==269 and not any(k.startswith(('CELL:','ACTOR:'))for k in keys)
  assert len(rows)==len(list(path.glob('*.png')))==270
  for image in path.glob('*.png'):assert image.read_bytes()==(old/image.name).read_bytes(),(lang,image.name)
  entries=[]
  for boundary in native['boundaries']:
   if not boundary['at'].startswith('執行段入口'):continue
   matches=[r for r in rows if r.get('battle',{}).get('prompt')==8 and r['battle']['round']==boundary['round']]
   assert len(matches)==1;row=matches[0]
   assert row['battle']['probe_seed']==boundary['state']['seed']
   assert actor_diffs(boundary['state'],row['battle']['probe_actors'])==[]
   entries.append(row['step'])
  end=next(r for r in rows if r.get('battle',{}).get('stage')==5);b=end['battle']
  assert (b['round'],b['verdict'],b['probe_return_scratch'])==(4,2,1)
  assert actor_diffs(native['battle_exit_state'],b['probe_actors'])==[]
  assert len(b['probe_rolls'])==99
  assert [(x['N'],x['Before'])for x in b['probe_rolls']]==[(x['n'],x['seed_before'])for x in native['rolls'][:-1]]
  world=rows[-1]['world'];assert [world[k]for k in ['x','y','sub_x','sub_y','ticks']]==[49,23,4,4,3300]
  assert world['probe_world_roll_count']==100 and world['probe_world_seed']==1084083701
  assert (world['probe_world_last_roll']['N'],world['probe_world_last_roll']['Before'])==(native['rolls'][-1]['n'],native['rolls'][-1]['seed_before'])
  assert all(r['world']['ticks']==3360 for r in rows if r.get('battle'))
  party=rows[-1]['town']['probe_full_party']
  for member,rawhex in zip(party,native['post_battle'][4]['state']['characters']):
   raw=bytes.fromhex(rawhex)
   assert [base64.b64decode(x).hex()for x in member['carried']]==[raw[o:o+16].hex()for o in offsets]
   assert member['wounds']==raw[27]and member['gold']==int.from_bytes(raw[42:45],'big')
   assert member['sheet']['adventure_points']==int.from_bytes(raw[38:40],'big')
  assert party[0]['backpack'][-1]['name']=='War Hammer'
  lodging=PRIVATE/('lodging-'+lang);lodgingrows=read(lodging/'trace.json')['steps'];save=read(lodging/'complete.json');saves.append(save)
  assert len(lodgingrows)==508 and len(read(lodging/'driver-keys.json'))==507 and lodgingrows[:270]==rows
  assert save['schema']=='kol-save-v3'and len(save['characters'])==16 and save['settings']['faithful_inn_stay']
  assert save['location']==dict(kind='town',town=7,entry_point=3,x=57,y=8)
  assert save['clock']==dict(ticks=3360,part=0,day=2,month=3,year=0,weather=1)
  byname={m['name']:m for m in save['characters']if m};beforestay=lodgingrows[494]['town']['probe_full_party']
  route_diffs=[(m['name'],i)for m,s in zip(party,beforestay)for i,(a,z)in enumerate(zip(m['carried'],s['carried']))if a!=z]
  assert route_diffs==[('Theodoric',0)]and base64.b64decode(beforestay[1]['carried'][0])==bytes(16)
  for m in beforestay:
   s=byname[m['name']];assert s['carried']==[base64.b64decode(x).hex()for x in m['carried']]
   assert s['wounds']==m['wounds']==0 and s['gold']==m['gold']-(56 if m['name']=='Olear'else 0)
   assert s['sheet']['adventure_points']==m['sheet']['adventure_points']
  observed=PRIVATE/('observed-reload-'+lang);formal=PRIVATE/('formal-reload-'+lang)
  reloadrows=read(observed/'trace.json')['steps'];formalrows=read(formal/'trace.json')['steps']
  assert len(reloadrows)==len(formalrows)==16
  plain=copy.deepcopy(reloadrows)
  for row in plain:row.get('town',{}).pop('probe_full_party',None)
  assert plain==formalrows
  assert len(list(formal.glob('*.png')))==16
  for image in observed.glob('*.png'):assert image.read_bytes()==(formal/image.name).read_bytes()
  final=reloadrows[-1]['town']['probe_full_party'];assert[m['name']for m in final]==['Theodoric','Olear']
  for m in final:
   s=byname[m['name']];assert[base64.b64decode(x).hex()for x in m['carried']]==s['carried']
   for key in ['id','wounds','gold','place','town_code','building_id','sheet','quests','progress']:assert m.get(key,0)==s.get(key,0),(lang,m['name'],key)
   for key in ['in_hand','backpack','worn']:assert item(m.get(key))==item(s.get(key)),(lang,m['name'],key)
  outputs.append(dict(language=lang,normal_keys=269,states=270,images_unchanged=270,execution_entries=entries,
   rolls=100,roll_faces_and_seeds_match=True,final_sampled_actor_differences=[],last_hurt=25,carried_slots=34,
   world=[49,23,4,4,3300],gold=[m['gold']for m in party],adventure_points=[m['sheet']['adventure_points']for m in party],
   lodging_keys=507,formal_reload_states=16,formal_reload_images=16))
 for a,b in zip(*traces):
  for k in ['probe_actors','probe_party','probe_rolls','probe_seed','probe_return_scratch']:assert a.get('battle',{}).get(k)==b.get('battle',{}).get(k)
  assert a.get('town',{}).get('probe_full_party')==b.get('town',{}).get('probe_full_party')and a.get('world')==b.get('world')
 assert {k:v for k,v in saves[0].items()if k!='saved_at'}=={k:v for k,v in saves[1].items()if k!='saved_at'}
 for filename,signal in [('worklist-selftest-final.log','正反對照通過'),('worklist-verify-final.log','未完成 22 條'),('worklist-scan-final.log','0 行沒有')]:
  assert signal in(PRIVATE/filename).read_text(),filename
 return dict(schema='kol-unarmed-resolution-receipt-v1',historical_shared_inputs=1375,parent_attachments=5343,parent_results_recomputed=True,
  code_inputs=318,changed_code_inputs=built['changed_inputs'],production_sha256=built['production_sha256'],normal=outputs,
  ida_instructions_checked=1045,resolution_backlinks_verified=True,old_negative_rejected=True,full_entry_rng_verified=False,full_actor_record_parity=False,
  status='本場徒手分支、100骰、既定角色抽樣及正常新保存鏈有限CONFORMED；其他玩家分支與完整重製仍開放。')
def attachments():
 names={str(p.relative_to(ROOT))for p in PRIVATE.rglob('*')if p.is_file()}
 names|={'CONTEXT.md','WORKLIST.md','WORKLOG.md','docs/worklist.json','docs/re/RESEARCH-LOG.md','docs/re/20261003-unarmed-resolution.md',
  'docs/spec/009-combat-attack.md','docs/spec/011-wilderness-ambush.md','docs/spec/027-battle-exit-wounds.md','docs/verification/unarmed-resolution-20261003/verify.py'}
 return{n:sha(ROOT/n)for n in sorted(names)}
if __name__=='__main__':
 ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--check',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
 assert sum([args.seal,args.check,args.audit])<=1
 result=validate()
 if args.seal:
  target=Path('/out/receipt.json');assert not target.exists();result['attachments']=attachments();target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
 elif not args.check:
  saved=read(HERE/'receipt.json')
  for name,digest in saved['attachments'].items():assert sha(ROOT/name)==digest,name
  assert canonical(result)=={k:v for k,v in saved.items()if k!='attachments'}
  if args.audit:
   control=read(PRIVATE/'host-control.json')
   assert control['git_diff_check']=='passed'and control['git_staged']==control['project_containers']==[]
   assert control['head']=='1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
   bad=[];md=[]
   for directory,dirs,files in os.walk(ROOT):
    for name in dirs+files:
     p=Path(directory)/name
     if p.lstat().st_uid==0:bad.append(str(p))
     if p.is_dir()and p.name.endswith('.md'):md.append(str(p))
   assert not bad and not md
   names=set(saved['attachments'])|set(read(PRIVATE/'production-inputs.json')['code_inputs'])
   for name in names:
    s=(ROOT/name).stat();assert(s.st_uid,s.st_gid)==(1000,1000),name
   print(json.dumps(dict(schema='kol-unarmed-resolution-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),attachments=len(saved['attachments']),
    code_inputs=318,parent_results_recomputed=True,full_readonly_verification='passed',host_control=control,
    root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000]),ensure_ascii=False))
 print(json.dumps({k:v for k,v in result.items()if k!='attachments'},ensure_ascii=False))
