"""收尾稽核：保留歷史鎖定，確認新分級匯出、擁有權及可回查入口。"""
from pathlib import Path
import hashlib, json, os, runpy

root = Path('/src')
proof = root / 'docs/verification/battle-settlement-20261002'
runpy.run_path(str(proof / 'verify_readonly.py'))

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def read(path):
    return json.loads(path.read_text())

old = read(root / 'docs/verification/battle-complete-20261001/receipt.json')
for records in [old['current_inputs'], old['attachments']]:
    for name, expected in records.items():
        path = root / name
        assert sha(path) == expected, path
        assert path.stat().st_uid == path.stat().st_gid == 1000, path
index_path = root / 'docs/re/battle-settlement-address-index.json'
index = read(index_path)
export_path = root / 'workplace/battle-settlement-20261001/settlement-index-final-private.json'
export = read(export_path)
assert export['function_count'] == 338 and export['tool'] == index['tool'] == 'IDA Pro 9.4'
assert export['sha256'] == index['sha256'] and export['address_space'] == index['address_space'] == 'IDA linear'
rows = {row['original_ea']: row for record in export['records'] for row in record['instructions']}
for address, note in index['entries'].items():
    row = rows[address]
    assert row['bytes'] and row['operand_text']
    assert any(a['semantic'] == note['semantic'] and a['confidence'] == '已證實'
               and a['index_sha256'] == sha(index_path) for a in row['annotations']), address
root_owned, md_dirs = [], []
for parent, directories, files in os.walk(root):
    for name in directories + files:
        path = Path(parent) / name
        if path.lstat().st_uid == 0:
            root_owned.append(str(path.relative_to(root)))
        if name.endswith('.md') and path.is_dir():
            md_dirs.append(str(path.relative_to(root)))
assert not root_owned and not md_dirs, (root_owned, md_dirs)
worklist = read(root / 'docs/worklist.json')
assert sum(x.get('status') != 'done' for x in worklist['items']) == 22
assert sum(x.get('status') == 'done' for x in worklist['items']) == 84
locks = [proof / 'receipt.json', proof / 'verify.py', proof / 'verify_readonly.py', proof / 'audit.py',
         index_path, export_path, root / 'docs/re/20261002-battle-settlement.md',
         root / 'docs/spec/027-battle-exit-wounds.md', root / 'docs/worklist.json', root / 'WORKLIST.md',
         proof / 'worklist-final.log']
for path in locks:
    assert path.stat().st_uid == path.stat().st_gid == 1000, path
target = Path('/out/audit.json')
assert not target.exists() and target.parent.stat().st_uid == os.getuid() == 1000
summary = dict(schema='kol-normal-settlement-audit-v1', date='2026-10-02',
               preserved_re180_inputs=len(old['current_inputs']), preserved_re180_attachments=len(old['attachments']),
               re181_attachments=len(read(proof / 'receipt.json')['attachments']),
               merged_annotations=len(index['entries']), uid=1000, gid=1000,
               root_owned=root_owned, md_same_name_directories=md_dirs, worklist=dict(open=22, done=84),
               locks={str(p.relative_to(root)): sha(p) for p in locks},
               cleanup='一次性容器均使用--rm；批末主機docker ps -a核實，GUI的Xvfb由trap與wait清理。',
               remaining='SPEC-027搜刮切片DRAFT、正式產品及正常保存鏈仍開放；未提交或發布。')
target.write_text(json.dumps(summary, ensure_ascii=False, indent=2)+'\n')
print(json.dumps({k: v for k, v in summary.items() if k != 'locks'}, ensure_ascii=False))
