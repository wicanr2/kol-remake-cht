"""封存後稽核；Git與Docker控制面由主機實際查核後傳入。"""
from pathlib import Path
import argparse
import hashlib
import json
import os

ROOT = Path('/src')
HERE = ROOT / 'docs/verification/world-clock-20261002'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--head', required=True)
    parser.add_argument('--project-containers', type=int, required=True)
    parser.add_argument('--staged-empty', action='store_true', required=True)
    parser.add_argument('--diff-check', action='store_true', required=True)
    args = parser.parse_args()
    assert args.project_containers == 0
    receipt_path = HERE / 'receipt.json'
    receipt = json.loads(receipt_path.read_text())
    for name, digest in receipt['attachments'].items():
        path = ROOT / name
        assert sha(path) == digest, name
        assert (path.stat().st_uid, path.stat().st_gid) == (1000, 1000), name
    assert (receipt_path.stat().st_uid, receipt_path.stat().st_gid) == (1000, 1000)
    roots, md_dirs = [], []
    for base, dirs, files in os.walk(ROOT):
        for name in dirs + files:
            path = Path(base) / name
            if path.is_symlink():
                continue
            if path.stat().st_uid == 0:
                roots.append(str(path.relative_to(ROOT)))
            if path.is_dir() and path.name.endswith('.md'):
                md_dirs.append(str(path.relative_to(ROOT)))
    assert not roots, roots[:20]
    assert not md_dirs, md_dirs[:20]
    result = dict(schema='kol-world-time-audit-v1', head=args.head,
                  receipt_sha256=sha(receipt_path), attachments=len(receipt['attachments']),
                  attachment_ownership='1000:1000', root_owned=0, md_named_directories=0,
                  project_containers=0, staged_empty=True, git_diff_check=True,
                  rights='新原版副本、payload、PNG與候選程式僅本機；未提交或發布。',
                  temporary_outputs='本輪輸出限world-clock-20261002，歷史來源及封存副本保留。')
    target = Path('/out/audit.json')
    assert not target.exists()
    target.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(result, ensure_ascii=False))


if __name__ == '__main__':
    main()
