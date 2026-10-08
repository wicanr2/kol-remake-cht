"""封存後稽核；不改主收據，預設可在全唯讀容器重跑。"""
from pathlib import Path
import argparse
import hashlib
import json
import runpy

ROOT = Path('/src')
HERE = ROOT / 'docs/verification/battle-return-20261002'
PRIVATE = ROOT / 'workplace/battle-return-20261002'


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate():
    runpy.run_path(str(HERE / 'verify_return.py'))['validate']()
    receipt = read(HERE / 'receipt.json')
    for name, digest in receipt['attachments'].items():
        path = ROOT / name
        assert sha(path) == digest, name
        assert path.stat().st_uid == path.stat().st_gid == 1000, name
    root_owned = []
    md_dirs = []
    for path in ROOT.rglob('*'):
        if path.lstat().st_uid == 0:
            root_owned.append(str(path.relative_to(ROOT)))
        if path.is_dir() and path.name.endswith('.md'):
            md_dirs.append(str(path.relative_to(ROOT)))
    assert not root_owned and not md_dirs
    control = read(PRIVATE / 'final-control.json')
    assert control['git_diff_check_exit'] == 0 and control['staged_files'] == []
    assert control['related_containers'] == []
    rows = read(ROOT / 'docs/worklist.json')['items']
    done = sum(r.get('status') == 'done' for r in rows)
    assert done == 84 and len(rows) - done == 22
    database = ROOT / 'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64'
    assert sha(database) == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    return dict(schema='kol-world-return-source-audit-v1', date='2026-10-02',
        receipt_sha256=sha(HERE / 'receipt.json'), verifier_sha256=sha(HERE / 'verify_return.py'),
        auditor_sha256=sha(HERE / 'audit_return.py'), attachments=len(receipt['attachments']),
        preserved_shared_inputs=1303, uid=1000, gid=1000, root_owned=root_owned,
        md_same_name_directories=md_dirs, formal_database_unchanged=True,
        git_diff_check_exit=0, staged_files=[], related_containers=[],
        worklist=dict(open=22, done=84), production_unchanged=True,
        cleanup='一次性Docker --rm；本輪無背景GUI程序，沒有新映像；原始payload與未驗PNG留本機。',
        remaining='返回子切片DRAFT；先確認正常勝利／逃走最後暫存，產品世界同狀態仍開放。')


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
