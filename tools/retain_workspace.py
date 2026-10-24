#!/usr/bin/env python3
"""Docker限定：完整封存舊workplace，逐檔核對後才清除；不省略任何執行檔。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import stat
import tarfile
import tempfile


def digest(stream):
    return hashlib.file_digest(stream, 'sha256').hexdigest()


def inventory(root, excluded=()):
    rows = {}
    paths=[]
    for child in root.iterdir():
        if child.name in excluded:
            continue
        paths.append(child)
        if child.is_dir() and not child.is_symlink():
            paths.extend(child.rglob('*'))
    for p in sorted(paths):
        name = str(p.relative_to(root))
        if name.split('/')[0] in excluded:
            continue
        s = p.lstat()
        assert (s.st_uid, s.st_gid) == (os.getuid(), os.getgid()), p
        row = dict(mode=stat.S_IMODE(s.st_mode))
        if p.is_symlink():
            row.update(kind='symlink', target=os.readlink(p))
        elif p.is_dir():
            row.update(kind='directory')
        elif p.is_file():
            with p.open('rb') as f:
                row.update(kind='file', size=s.st_size, sha256=digest(f))
        else:
            raise AssertionError(f'非一般檔案，須先結束擁有程序：{p}')
        rows[name] = row
    return rows


def verify(dest):
    info = json.loads((dest/'retention.json').read_text())
    archive = dest/'evidence.tar.xz'
    with archive.open('rb') as f:
        assert digest(f) == info['archive_sha256']
    observed, hashes = {}, {}
    with tarfile.open(archive, 'r:xz') as packed:
        for m in packed:
            row = info['entries'][m.name]
            assert m.name not in observed
            assert m.mode == row['mode'], m.name
            if m.isfile():
                with packed.extractfile(m) as f:
                    actual = digest(f)
                assert m.size == row['size'] and actual == row['sha256'], m.name
                hashes[m.name] = actual
            elif m.islnk():
                assert hashes[m.linkname] == row['sha256'], m.name
                hashes[m.name] = hashes[m.linkname]
            elif m.issym():
                assert row['kind'] == 'symlink' and m.linkname == row['target']
            else:
                assert m.isdir() and row['kind'] == 'directory'
            observed[m.name] = row
    assert observed == info['entries']
    return info


def main():
    assert Path('/.dockerenv').exists()
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('phase', choices=('archive', 'verify', 'clean', 'restore'))
    p.add_argument('--source', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--exclude', action='append', default=[])
    p.add_argument('--name', help='restore時只還原此頂層工作目錄')
    a = p.parse_args()
    assert a.source.is_dir() and a.out.parent.is_dir()
    if a.phase == 'archive':
        assert not a.out.exists()
        rows = inventory(a.source, a.exclude)
        assert rows
        a.out.mkdir()
        canonical = {}
        with tarfile.open(a.out/'evidence.tar.xz', 'w:xz', preset=1) as packed:
            for name, row in rows.items():
                path = a.source/name
                m = packed.gettarinfo(str(path), arcname=name)
                if row['kind'] == 'file':
                    identity = (row['size'], row['sha256'])
                    if identity in canonical:
                        m.type, m.linkname, m.size = tarfile.LNKTYPE, canonical[identity], 0
                        packed.addfile(m)
                    else:
                        canonical[identity] = name
                        with path.open('rb') as f:
                            packed.addfile(m, f)
                else:
                    packed.addfile(m)
        with (a.out/'evidence.tar.xz').open('rb') as f:
            sha = digest(f)
        report = dict(status='PASS', schema='kol-complete-workspace-retention/1',
                      source=str(a.source), excluded=a.exclude, entries=rows,
                      original_bytes=sum(r.get('size', 0) for r in rows.values()),
                      archive_bytes=(a.out/'evidence.tar.xz').stat().st_size,
                      archive_sha256=sha, omitted_files=[],
                      rights='local-only; includes original-derived research',
                      limits='歷史命令的原始路徑保留；restore可完整重建任一頂層目錄。')
        (a.out/'retention.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
        verify(a.out)
        print('ARCHIVED VERIFIED', len(rows), report['original_bytes'], report['archive_bytes'],flush=True)
    elif a.phase == 'verify':
        r = verify(a.out)
        print('VERIFIED',len(r['entries']),flush=True)
    elif a.phase == 'clean':
        r = verify(a.out)
        assert inventory(a.source,r['excluded']) == r['entries'], '來源在封存後改變，拒絕刪除'
        names = {n.split('/')[0] for n in r['entries']}
        for name in sorted(names):
            path = a.source/name
            if path.is_dir() and not path.is_symlink():
                shutil.rmtree(path)
            else:
                path.unlink()
        print('CLEANED',len(names),r['original_bytes'],flush=True)
    else:
        r = verify(a.out)
        assert a.name and '/' not in a.name and a.name not in ('.','..')
        assert not (a.source/a.name).exists()
        needed={(v['size'],v['sha256']) for k,v in r['entries'].items()
                if k.startswith(a.name+'/') and v['kind']=='file'}
        # 串流一次解壓，暫存要還原的內容。xz內的跨目錄硬連結不反覆倒帶解壓。
        with tempfile.TemporaryDirectory(prefix='kol-restore-') as temp:
            with tarfile.open(a.out/'evidence.tar.xz','r|xz') as packed:
                for m in packed:
                    row=r['entries'][m.name]
                    content=None
                    if row['kind']=='file':
                        identity=(row['size'],row['sha256'])
                        content=Path(temp)/(str(row['size'])+'-'+row['sha256'])
                        if m.isfile() and identity in needed:
                            with packed.extractfile(m) as src,content.open('wb') as out:
                                shutil.copyfileobj(src,out)
                    if m.name.split('/')[0] != a.name:
                        continue
                    dest=a.source/m.name
                    assert not m.name.startswith('/') and '..' not in Path(m.name).parts
                    if row['kind']=='directory':
                        dest.mkdir(parents=True,exist_ok=True)
                    elif row['kind']=='symlink':
                        dest.parent.mkdir(parents=True,exist_ok=True)
                        dest.symlink_to(row['target'])
                    else:
                        dest.parent.mkdir(parents=True,exist_ok=True)
                        assert content.is_file(), m.name
                        shutil.copyfile(content,dest)
                    if not dest.is_symlink():
                        dest.chmod(row['mode'])
        assert inventory(a.source/a.name)=={k[len(a.name)+1:]:v for k,v in r['entries'].items() if k.startswith(a.name+'/')}
        print('RESTORED',a.name,flush=True)


if __name__ == '__main__':
    main()
