"""RE-185封存後稽核；主收據只讀，所有清理結果另記。"""
from pathlib import Path
import argparse
import hashlib
import json
import runpy

ROOT = Path('/src')
HERE = ROOT / 'docs/verification/battle-exits-20261002'
PRIVATE = ROOT / 'workplace/battle-exits-20261002'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def validate():
    receipt = read(HERE / 'receipt.json')
    result = runpy.run_path(str(HERE / 'verify_exits.py'))['validate']()
    assert result == {k: v for k, v in receipt.items() if k != 'attachments'}
    for name, digest in receipt['attachments'].items():
        path = ROOT / name
        assert sha(path) == digest, name
        assert path.stat().st_uid == path.stat().st_gid == 1000, name
    root_owned, md_dirs = [], []
    for path in ROOT.rglob('*'):
        if path.lstat().st_uid == 0:
            root_owned.append(str(path.relative_to(ROOT)))
        if path.is_dir() and path.name.endswith('.md'):
            md_dirs.append(str(path.relative_to(ROOT)))
    assert not root_owned and not md_dirs
    control = read(PRIVATE / 'final-control.json')
    assert control['git_diff_check_exit'] == 0
    assert control['staged_files'] == [] and control['related_containers'] == []
    rows = read(ROOT / 'docs/worklist.json')['items']
    done = sum(r.get('status') == 'done' for r in rows)
    assert done == 84 and len(rows) - done == 22
    database = ROOT / 'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64'
    assert sha(database) == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    return dict(schema='kol-battle-exits-audit-v1', date='2026-10-02',
        receipt_sha256=sha(HERE / 'receipt.json'), verifier_sha256=sha(HERE / 'verify_exits.py'),
        auditor_sha256=sha(HERE / 'audit_exits.py'), attachments=len(receipt['attachments']),
        preserved_shared_inputs=1309, uid=1000, gid=1000, root_owned=root_owned,
        md_same_name_directories=md_dirs, formal_database_unchanged=True,
        git_diff_check_exit=0, staged_files=[], related_containers=[],
        worklist=dict(open=22, done=84), product_sha256=result['product_sha256'],
        cleanup='一次性Docker均--rm；Xvfb有trap與wait；沒有新映像，原始payload、新PNG與程式僅本機。',
        remaining='次格返回有限CONFORMED；世界時間、勝利暫存與亂數接續維持父閘門。')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--export', action='store_true')
    args = parser.parse_args()
    result = validate()
    if args.export:
        target = Path('/out/audit.json')
        assert not target.exists()
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    else:
        assert result == read(HERE / 'audit.json')
    print(json.dumps(result, ensure_ascii=False))


if __name__ == '__main__':
    main()
