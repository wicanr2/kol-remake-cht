from pathlib import Path
import hashlib,json,subprocess
r=Path('/src');here=r/'docs/verification/spell-list-input-20261003'
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
host=read(here/'host-state.json'); receipt=read(here/'receipt.json')
initial={line[3:] for line in host['git_initial'].splitlines() if line}
final={line[3:] for line in host['git_final'].splitlines() if line}
assert initial<=final
assert not [line for line in host['git_final'].splitlines() if line and line[0] not in [' ','?']]
assert host['diff_check']==0
for name,h in receipt['attachments'].items():
    p=r/name
    assert sha(p)==h,name
    assert p.lstat().st_uid==p.lstat().st_gid==1000,name
root_owned=subprocess.run(['find',str(r),'-user','root','-o','-type','d','-name','*.md'],text=True,capture_output=True,check=True)
assert not root_owned.stdout,root_owned.stdout
project=[c for c in host['docker_mount_records'] if c['name'].startswith('/kol-') or any(m['Source']=='/home/anr2/cht/kol' or m['Source'].startswith('/home/anr2/cht/kol/') for m in c['mounts'])]
assert not project,project
for name in ['remote_writes','commit','push','packaging','publishing']:assert not host[name]
work=read(r/'docs/worklist.json')
assert sum(i.get('status')=='done' for i in work['items'])==84
assert sum(i.get('status')!='done' for i in work['items'])==22
body=dict(schema=1,receipt_sha256=sha(here/'receipt.json'),verified_attachments=len(receipt['attachments']),
          ownership='1000:1000',root_owned_or_md_directories=0,symlink_policy='lstat; find不跟隨symlink',
          git_initial_paths=len(initial),git_final_paths=len(final),retained_initial_paths=True,staged=[],diff_check=0,
          project_containers=project,inspected_other_containers=len(host['docker_mount_records']),
          production_changed=True,changed_go=receipt['changed_go'],format_decision='pending',gloss_layout_decision='pending',
          worklist_open=22,worklist_done=84,goal_status='active',
          audit_tool='docs/verification/spell-list-input-20261003/audit.py',audit_tool_sha256=sha(here/'audit.py'),
          host_state='docs/verification/spell-list-input-20261003/host-state.json',host_state_sha256=sha(here/'host-state.json'),
          cleanup='本輪容器均--rm，Xvfb由trap結束；原版／正常保存／原型僅本機，不清理其他專案。',
          commit=False,push=False,remote_writes=False,packaging=False,publishing=False)
(Path('/audit')/'audit.json').write_text(json.dumps(body,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(body,ensure_ascii=False))
