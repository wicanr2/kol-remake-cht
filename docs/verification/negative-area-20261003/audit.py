from pathlib import Path
import hashlib, json, os, stat, subprocess
r = Path('/src'); here = r / 'docs/verification/negative-area-20261003'
read = lambda p: json.loads(p.read_text())
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
host = read(here / 'host-state.json'); receipt = read(here / 'receipt.json')
initial = {line[3:] for line in host['git_initial'].splitlines() if line}
final = {line[3:] for line in host['git_final'].splitlines() if line}
assert initial <= final
assert not [line for line in host['git_final'].splitlines() if line and line[0] not in [' ', '?']]
assert host['diff_check'] == 0
for name, h in receipt['attachments'].items():
    p = r / name
    assert sha(p) == h, name
    assert p.lstat().st_uid == p.lstat().st_gid == 1000, name
owned = subprocess.run(['find', str(r), '-user', 'root', '-o', '-type', 'd', '-name', '*.md'],
                       text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True)
assert not owned.stdout, owned.stdout
project = [c for c in host['docker_mount_records'] if c['name'].startswith('/kol-') or
           any(m['Source'] == '/home/anr2/cht/kol' or m['Source'].startswith('/home/anr2/cht/kol/') for m in c['mounts'])]
assert not project
for name in ['remote_writes', 'commit', 'push', 'packaging', 'publishing']:
    assert not host[name]
work = read(r / 'docs/worklist.json')
assert sum(i.get('status') == 'done' for i in work['items']) == 84
assert sum(i.get('status') != 'done' for i in work['items']) == 22
body = dict(schema=1, receipt_sha256=sha(here / 'receipt.json'),
            verified_attachments=len(receipt['attachments']), ownership='1000:1000',
            root_owned_or_md_directories=0, symlink_policy='find不跟隨symlink；附件用lstat',
            git_initial_paths=len(initial), git_final_paths=len(final), retained_initial_paths=True,
            staged=[], diff_check=0, project_containers=project,
            inspected_other_containers=len(host['docker_mount_records']),
            production_changed=False, worklist_open=22, worklist_done=84,
            format_decision='pending', goal_status='active',
            original_area='normal-negative-confirmed', remake_direction_key_gap='open',
            audit_tool='docs/verification/negative-area-20261003/audit.py', audit_tool_sha256=sha(here / 'audit.py'),
            host_state='docs/verification/negative-area-20261003/host-state.json', host_state_sha256=sha(here / 'host-state.json'),
            cleanup='本輪容器均--rm；Xvfb由trap結束，未改其他專案資源；workplace保存僅本機研究產物',
            git_or_external_writes=False)
(Path('/audit') / 'audit.json').write_text(json.dumps(body, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(body, ensure_ascii=False))
