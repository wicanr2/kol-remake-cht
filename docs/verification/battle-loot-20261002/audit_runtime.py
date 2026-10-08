"""補充執行資料的雜湊鎖定；保留既有收據，--export 只新增一次稽核。"""
from pathlib import Path
import argparse, hashlib, json, runpy, sys

root = Path('/src')
proof = root/'docs/verification/battle-loot-20261002'
parser = argparse.ArgumentParser()
parser.add_argument('--export',action='store_true')
args = parser.parse_args()
target = Path('/out/runtime-audit.json') if args.export else proof/'runtime-audit.json'
if args.export:
    assert not target.exists() and target.parent.stat().st_uid==1000
    sys.argv = [str(proof/'audit.py')]
    runpy.run_path(str(proof/'audit.py'),run_name='__main__')
else:
    sys.argv = [str(proof/'verify_loot.py')]
    runpy.run_path(str(proof/'verify_loot.py'),run_name='__main__')

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

receipt = json.loads((proof/'receipt.json').read_text())
audit = json.loads((proof/'audit.json').read_text())
for name,digest in audit['locks'].items():
    assert sha(root/name)==digest, name
paths = [p for p in (root/'docs/i18n').rglob('*') if p.is_file()]
paths += list((root/'docs/re').glob('*.json'))
paths = sorted(set(p for p in paths if str(p.relative_to(root)) not in receipt['attachments']))
assert len(paths)==65
assert all(p.stat().st_uid==p.stat().st_gid==1000 for p in paths)
binary = root/'workplace/battle-loot-20261002/fix2/kol-battle-loot-linux'
assert binary.stat().st_mode & 0o777 == 0o755
result = dict(schema='kol-battle-loot-runtime-audit-v1',date='2026-10-02',
    receipt_sha256=sha(proof/'receipt.json'),audit_sha256=sha(proof/'audit.json'),
    verifier_sha256=sha(proof/'audit_runtime.py'),production_mode='0755',
    additional_runtime_and_index_inputs={str(p.relative_to(root)):sha(p) for p in paths},
    note='補充語言資料、標題／城鎮樣本及研究索引鎖定；既有5258附件與歷史收據不覆寫。')
if args.export:
    target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
else:
    assert json.loads(target.read_text())==result
print(json.dumps(dict(additional_inputs=len(paths),production_mode='0755',readonly=not args.export),ensure_ascii=False))
