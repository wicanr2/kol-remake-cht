from pathlib import Path
import hashlib,json,runpy
r=Path('/src');here=Path(__file__).resolve().parent;private=r/'workplace/monster-equipment-20261003'
read=lambda p:json.loads(p.read_text());sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
receipt=read(here/'receipt.json');verified=runpy.run_path(str(here/'verify.py'))['validate']()
assert verified=={k:v for k,v in receipt.items()if k!='attachments'}
owners={}
for name,h in receipt['attachments'].items():
 p=r/name;assert p.is_file() and sha(p)==h,name
 key=f'{p.stat().st_uid}:{p.stat().st_gid}';owners[key]=owners.get(key,0)+1
assert set(owners)=={'1000:1000'}
bad=[str(p.relative_to(r))for p in r.rglob('*')if p.stat().st_uid==0 or p.is_dir()and p.name.endswith('.md')]
assert not bad,bad[:10]
state=read(private/'host-final-state.json');assert not state['project_containers'] and not state['staged'] and state['diff_check_exit_code']==0
assert state['all_initial_paths_retained']
body={**state,'schema':1,'receipt_sha256':sha(here/'receipt.json'),'attachments':len(receipt['attachments']),'recalculated_result':True,'owners':owners,'root_owned_or_md_directories':bad,'production_inputs':319,'changed_production_inputs':[],'goal_status':'active','format_decision':'pending','private_local_outputs':True}
path=Path('/seal/audit.json');assert not path.exists();path.write_text(json.dumps(body,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in body.items()if k not in['initial_paths','final_paths']},ensure_ascii=False))
