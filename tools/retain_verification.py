#!/usr/bin/env python3
"""在 Docker 內保存完成的比較證據，再刪除已核對的工作目錄。"""
import argparse
import hashlib
import json
import shutil
import tarfile
from html.parser import HTMLParser
from pathlib import Path

ROOT = Path('/src')
OUT = Path('/out')
NAMES = (
    'hd-npc-mounts-20261015',
    'hd-equipment-pages-20261015',
    'hd-character-templates-20261014',
    'hd-spell-scroll-runtime-20261014',
    'hd-spell-scroll-20261014',
    'hd-native-ui-runtime-20261014',
    'hd-native-ui-20261014',
    'hd-ending-runtime-20261014',
    'hd-completion-20261014',
    'hd-pixel-consistency-20261013',
    'hd-music-runtime-20261013',
    'display-switch-runtime-20261012',
    'display-switch-20261011',
    'hd-runtime-20261011',
    'controlled-scenes-20261011',
    'cjk-closeout-20261011',
    'quest-entry-rng-prototype-20261006',
    'quest-marker-phase-20261006',
    'move-message-visibility-20261006',
    'blocked-stun-formal-20261007',
    'finish-platform-20261007',
    'quest-rng-formal-20261007',
    'platform-quest-rng-20261007',
    'matching-pilot-20261008',
    'matching-link-20261008',
    'matching-pack-20261008',
    'matching-c-20261008',
    'matching-c2-20261008',
    'matching-c3-20261008',
    'matching-c4-20261008',
    'matching-c5-20261008',
    'matching-c6-20261008',
    'matching-c7-20261009',
    'matching-c8-20261009',
    'matching-c9-20261009',
    'matching-c10-20261009',
    'matching-c11-20261009',
    'matching-c12-20261009',
    'matching-c13-20261009',
    'matching-c14-20261009',
    'matching-c15-20261009',
    'matching-c16-20261009',
    'matching-c17-20261009',
    'matching-c18-20261009',
    'matching-c19-20261009',
    'matching-c20-20261010',
    'matching-c21-20261010',
    'matching-c22-20261010',
    'matching-c23-20261010',
    'matching-c24-20261010',
    'matching-c25-20261010',
    'matching-c26-20261010',
    'matching-c27-20261010',
    'matching-c28-20261010',
    'matching-c29-20261010',
    'matching-c30-20261010',
    'matching-c31-20261010',
    'matching-c32-20261010',
    'matching-c33-20261010',
    'matching-c34-20261010',
)
EXECUTABLE_MAGIC = (b'\x7fELF', b'MZ', bytes.fromhex('cafebabe'),
                    bytes.fromhex('cffaedfe'), bytes.fromhex('feedfacf'))


def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


def record(path):
    st = path.stat()
    assert st.st_uid == st.st_gid == 1000, path
    return {'size': st.st_size, 'sha256': sha(path)}


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    assert not path.exists(), path
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def sealed_matches(actual, expected):
    if isinstance(expected, str):
        return actual['sha256'] == expected
    return actual == expected


def source_files(source):
    assert source.is_dir() and not source.is_symlink(), source
    result = {}
    for path in sorted(source.rglob('*')):
        assert not path.is_symlink(), path
        if path.is_file():
            result[str(path.relative_to(source))] = record(path)
        elif path.is_dir():
            assert path.stat().st_uid == path.stat().st_gid == 1000, path
        else:
            raise AssertionError(path)
    return result


def archive(name):
    source, dest = ROOT / 'workplace' / name, OUT / name
    assert OUT.is_dir() and OUT.stat().st_uid == OUT.stat().st_gid == 1000
    assert not dest.exists(), dest
    files = source_files(source)
    receipt = read(source / 'receipt.json')
    sealed = receipt.get('attachments', receipt.get('artifacts'))
    assert sealed, 'missing sealed receipt'
    for path, expected in sealed.items():
        assert sealed_matches(files[path], expected), path
    omitted = {}
    for path in files:
        with (source / path).open('rb') as stream:
            magic = stream.read(4)
        if any(magic.startswith(x) for x in EXECUTABLE_MAGIC):
            omitted[path] = files[path]
    retained = {k: v for k, v in files.items() if k not in omitted}
    dest.mkdir()
    with tarfile.open(dest / 'evidence.tar.xz', 'w:xz', preset=3) as packed:
        for path in retained:
            packed.add(source / path, arcname=path, recursive=False)
    # 小型入口可直接閱讀；完整紀錄及歷史腳本只保存一份壓縮副本。
    visible = [p for p in retained if '/' not in p and Path(p).suffix
               in ('.json', '.py', '.sh', '.log', '.txt', '.diff', '.c')]
    for path in visible:
        shutil.copy2(source / path, dest / path)
    index = {
        'schema': 'kol-retained-verification/1',
        'date': receipt.get('verified_on', receipt.get('date', '2026-10-07')),
        'source': 'workplace/' + name,
        'rights': 'local-only original-derived evidence; no publication',
        'receipt_sha256': files['receipt.json']['sha256'],
        'archive': record(dest / 'evidence.tar.xz'),
        'retained': retained, 'readable_copies': visible,
        'removed_rebuildable_executables': omitted,
        'original_bytes': sum(v['size'] for v in files.values()),
        'old_receipt_modified': False, 'new_game_execution': False,
        'limits': 'Historical receipts retain their original paths and hashes. '
                  'Executable bytes were removed; their hashes remain. '
                  'This checks retained evidence integrity, not a new game run.',
    }
    write(dest / 'retention.json', index)
    verify(name)
    print(name, 'archived', len(retained), 'files;', len(omitted),
          'rebuildable executables omitted', flush=True)


def verify(name):
    dest = OUT / name
    index = read(dest / 'retention.json')
    assert index['source'] == 'workplace/' + name
    assert record(dest / 'evidence.tar.xz') == index['archive']
    seen = {}
    with tarfile.open(dest / 'evidence.tar.xz', 'r:xz') as packed:
        for member in packed:
            path = Path(member.name)
            assert member.isfile() and not path.is_absolute()
            assert '..' not in path.parts and member.name not in seen
            with packed.extractfile(member) as stream:
                h = hashlib.sha256()
                size = 0
                for block in iter(lambda: stream.read(1024 * 1024), b''):
                    h.update(block)
                    size += len(block)
            seen[member.name] = {'size': size, 'sha256': h.hexdigest()}
    assert seen == index['retained']
    assert record(dest / 'receipt.json')['sha256'] == index['receipt_sha256']
    for path in index['readable_copies']:
        assert record(dest / path) == seen[path], path
    receipt = read(dest / 'receipt.json')
    sealed = receipt.get('attachments', receipt.get('artifacts'))
    original = {**seen, **index['removed_rebuildable_executables']}
    assert all(sealed_matches(original[p], v) for p, v in sealed.items())
    view = dest / 'view-manifest.json'
    if view.exists():
        manifest = read(view)
        for path, expected in manifest['files'].items():
            assert record(dest / path) == expected, path
        for path, expected in manifest['sources'].items():
            source_name = Path(path).parent.name
            assert source_name in NAMES
            assert sha(OUT / source_name / 'receipt.json') == expected
        class Links(HTMLParser):
            def handle_starttag(self, tag, attrs):
                for key, value in attrs:
                    if key in ('href', 'src'):
                        assert value and not value.startswith(('http:', 'https:', 'data:'))
                        target = (ROOT / 'docs/verification/local' / name / value.split('#')[0]).resolve()
                        assert target.is_relative_to(ROOT.resolve()) and target.is_file(), value
        Links().feed((dest / 'review.html').read_text())
    return index


def delete(name):
    index = verify(name)
    source = ROOT / 'workplace' / name
    assert source_files(source) == {
        **index['retained'], **index['removed_rebuildable_executables']}
    assert source.parent.resolve() == (ROOT / 'workplace').resolve()
    shutil.rmtree(source)
    assert not source.exists()
    print(name, 'removed after archive hash verification', flush=True)


def extract(name, dest):
    index = verify(name)
    assert dest is not None and dest.is_absolute() and not dest.exists()
    dest.mkdir(parents=True)
    with tarfile.open(OUT / name / 'evidence.tar.xz', 'r:xz') as packed:
        for member in packed:
            target = dest / member.name
            target.parent.mkdir(parents=True, exist_ok=True)
            with packed.extractfile(member) as src, target.open('xb') as dst:
                shutil.copyfileobj(src, dst)
            assert record(target) == index['retained'][member.name]
    print(name, 'evidence extracted; executable bytes excluded', flush=True)


def main():
    assert Path('/.dockerenv').exists(), 'Docker required'
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('stage', choices=['archive', 'verify', 'delete', 'extract'])
    parser.add_argument('names', nargs='+', choices=NAMES)
    parser.add_argument('--destination', type=Path)
    args = parser.parse_args()
    if args.stage == 'extract':
        assert len(args.names) == 1
    for name in args.names:
        if args.stage == 'archive':
            archive(name)
        elif args.stage == 'delete':
            delete(name)
        elif args.stage == 'extract':
            extract(name, args.destination)
        else:
            index = verify(name)
            print(name, 'PASS', len(index['retained']), 'retained hashes', flush=True)


if __name__ == '__main__':
    main()
