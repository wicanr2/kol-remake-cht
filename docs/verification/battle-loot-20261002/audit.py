"""唯讀核對新收據、歷史保存、正式資料庫及檔案衛生；稽核另存 /out。"""
from pathlib import Path
import hashlib, json, os, runpy, sys

root = Path('/src')
proof = root / 'docs/verification/battle-loot-20261002'
sys.argv = [str(proof/'verify_loot.py')]
runpy.run_path(str(proof/'verify_loot.py'),run_name='__main__')

def read(path):
    return json.loads(path.read_text())

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

receipt = read(proof/'receipt.json')
for name in receipt['attachments']:
    p = root/name
    assert p.stat().st_uid == p.stat().st_gid == 1000, name
old_receipt = root/'docs/verification/battle-settlement-20261002/receipt.json'
assert sha(old_receipt) == 'c6a39a1f789b56c0473ed0e240b9967250845e70092cb3efcd3c8161efc2ff83'
database = root/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64'
assert sha(database) == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
root_owned, md_dirs = [], []
for parent, directories, files in os.walk(root):
    for name in directories+files:
        p = Path(parent)/name
        if p.lstat().st_uid == 0:
            root_owned.append(str(p.relative_to(root)))
        if name.endswith('.md') and p.is_dir():
            md_dirs.append(str(p.relative_to(root)))
assert not root_owned and not md_dirs, (root_owned,md_dirs)
worklist = read(root/'docs/worklist.json')
counts = dict(open=sum(x.get('status')!='done' for x in worklist['items']),
              done=sum(x.get('status')=='done' for x in worklist['items']))
assert counts == dict(open=22,done=84)
remote = read(root/'workplace/battle-loot-20261002/remote-open-20261002.json')
assert len(remote)==24 and {6,44,33,28,23}.issubset({x['number'] for x in remote})
control = read(root/'workplace/battle-loot-20261002/final-control.json')
assert control['related_containers'] == [] and control['git_diff_check_exit'] == 0
assert control['staged_files'] == []
locks = [proof/'receipt.json',proof/'verify_loot.py',proof/'audit.py',old_receipt,database,
         root/'docs/worklist.json',root/'WORKLIST.md',root/'CONTEXT.md',
         root/'docs/re/RESEARCH-LOG.md',root/'tools/worklist.py',
         root/'workplace/battle-loot-20261002/final-control.json']
assert all(p.stat().st_uid==p.stat().st_gid==1000 for p in locks)
target = Path('/out/audit.json')
assert not target.exists() and target.parent.stat().st_uid==os.getuid()==1000
summary = dict(schema='kol-battle-loot-audit-v1',date='2026-10-02',
    receipt_attachments=len(receipt['attachments']),preserved_parent_inputs=1213,
    unchanged_re181_receipt=True,unchanged_formal_database=True,
    source_cases=288,drop_source_cases=12,production_matching_reload_frames=32,
    uid=1000,gid=1000,root_owned=root_owned,md_same_name_directories=md_dirs,
    worklist=counts,remote_open=24,related_containers=control['related_containers'],
    git_diff_check_exit=control['git_diff_check_exit'],staged_files=control['staged_files'],
    locks={str(p.relative_to(root)):sha(p) for p in locks},
    remaining='搜刮／掉落子切片有限CONFORMED；父規格及其他戰鬥／世界／平台交付閘門仍開放。',
    cleanup='一次性容器--rm；Xvfb由trap及wait清理；主機控制面核實專案容器0。')
target.write_text(json.dumps(summary,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in summary.items() if k!='locks'},ensure_ascii=False))
