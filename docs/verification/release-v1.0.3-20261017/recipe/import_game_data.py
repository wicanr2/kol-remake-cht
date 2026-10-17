#!/usr/bin/env python3
"""把合法持有、已抽取的武士傳說資料匯入引擎包；用法見封包 README.txt。"""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import shutil
import tempfile


def sha(path):
    return hashlib.file_digest(path.open('rb'), 'sha256').hexdigest()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--source', type=Path, required=True,
                   help='合法持有的已抽取遊戲目錄，內含 assets/；不是原版 ZIP')
    p.add_argument('--target', type=Path, required=True,
                   help='Windows 解壓目錄、macOS Contents/Resources 或 Linux 外部資料目錄')
    args = p.parse_args()
    engine = Path(__file__).resolve().parent.parent
    manifest = json.loads((engine / 'import-manifest.json').read_text())
    source, target = args.source.resolve(), args.target.resolve()
    checked = []
    for row in manifest['files']:
        relative = PurePosixPath(row['path'])
        if relative.is_absolute() or '..' in relative.parts or relative.parts[0] != 'assets':
            p.error('資料清單含不安全路徑')
        path = source / relative
        if not path.is_file() or path.is_symlink() or not path.resolve().is_relative_to(source):
            p.error(f'缺少合法抽取資料：{relative}')
        if path.stat().st_size != row['bytes'] or sha(path) != row['sha256']:
            p.error(f'資料版本或雜湊不符：{relative}')
        checked.append((path, relative))
    if (target / 'assets').exists():
        p.error('目標已有 assets/，請選新的空目錄，或先自行備份既有資料')
    target.mkdir(parents=True, exist_ok=True)
    # 全部來源驗證後才寫入，暫存與目標在同一檔案系統。
    with tempfile.TemporaryDirectory(prefix='.kol-import-', dir=target) as tmp:
        staged = Path(tmp)
        for path, relative in checked:
            dst = staged / relative
            dst.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, dst)
            if sha(dst) != sha(path):
                p.error(f'複製驗證失敗：{relative}')
        (staged / 'assets').rename(target / 'assets')
    if engine != target:
        for name in ('docs', 'LICENSE', 'README.txt', 'KNOWN-LIMITATIONS.txt', 'VERSION'):
            src, dst = engine / name, target / name
            if src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
            elif src.is_file():
                shutil.copyfile(src, dst)
    receipt = dict(schema='kol-data-import/1', version=manifest['version'],
                   files=len(checked), rights='原作衍生資料僅供本機使用，不得散布。')
    (target / 'data-import-receipt.json').write_text(
        json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print(f'已驗證並匯入 {len(checked)} 份資料：{target}')


if __name__ == '__main__':
    main()
