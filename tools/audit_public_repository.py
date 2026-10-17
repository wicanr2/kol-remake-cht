#!/usr/bin/env python3
"""公開化前的檔案範圍稽核；入口見docs/RELEASE-RIGHTS-REVIEW.md。"""
from pathlib import Path
import argparse
import collections
import hashlib
import json
import re
import subprocess

ROOT = Path('/src')
ASSET_ROOTS = ['assets/hd', 'assets/audio/hd']
def git(*args):
    return subprocess.check_output(['git', '-c', 'safe.directory=/src', '-C', str(ROOT), *args])
def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    assert Path('/.dockerenv').exists()
    head = git('rev-parse', 'HEAD').decode().strip()
    tracked = set(git('ls-files', '-z', '--', *ASSET_ROOTS).decode().strip('\0').split('\0'))
    disk = {str(p.relative_to(ROOT)) for name in ASSET_ROOTS for p in (ROOT/name).rglob('*') if p.is_file()}
    assert disk == tracked, (sorted(disk-tracked), sorted(tracked-disk))
    tree = {}
    for row in git('ls-tree', '-rz', 'HEAD', '--', *ASSET_ROOTS).split(b'\0'):
        if not row: continue
        info, path = row.split(b'\t', 1)
        tree[path.decode()] = info.split()[2].decode()
    resources_path = ROOT/'docs/verification/release-v1.0.3-20261017/build-inputs.json'
    resources = {row['path']: row for row in json.loads(resources_path.read_text())['resources']}
    assets = []
    for name in sorted(disk):
        content = (ROOT/name).read_bytes()
        blob = hashlib.sha1(b'blob '+str(len(content)).encode()+b'\0'+content).hexdigest()
        digest = hashlib.sha256(content).hexdigest()
        assert tree[name] == blob
        assert resources[name]['sha256'] == digest
        assets.append(dict(path=name, bytes=len(content), sha256=digest, git_blob=blob))
    forbidden_patterns = ['*.zip', '*.pdf', '*.exe', '*.EXE', '*.wav', '*.pcm', '*.dat', '*.DAT',
                          '*.tar', '*.xz', '*.pem', '*.key', '.env*', '*.i64', '*.bin', '*.dll']
    paths = sorted(set(git('log', '--all', '--format=', '--name-only', '--diff-filter=A', '--',
                           *forbidden_patterns).decode().splitlines())-{''})
    assert not paths, 'review historical raw or credential paths: '+str(paths)
    # 僅輸出命中檔名，不輸出疑似憑證內容。對全部可達歷史的文字blob做固定樣式檢查。
    extensions = {'.md','.json','.yaml','.yml','.toml','.txt','.log','.sh','.py','.go','.js','.cjs',
                  '.html','.svg','.ini','.cfg','.conf','.c','.h','.mod','.sum'}
    text_objects = {}
    for row in git('rev-list', '--objects', '--all').decode().splitlines():
        parts = row.split(' ', 1)
        if len(parts) != 2: continue
        oid, name = parts
        if Path(name).suffix in extensions or Path(name).name in ('LICENSE', '.gitconfig', '.npmrc', '.netrc'):
            text_objects[oid] = name
    pattern = re.compile(rb'gh[pousr]_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{40,}|'
                         rb'-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----|(?:AKIA|ASIA)[A-Z0-9]{16}')
    process = subprocess.Popen(['git','-c','safe.directory=/src','-C',str(ROOT),'cat-file','--batch'],
                               stdin=subprocess.PIPE, stdout=subprocess.PIPE)
    matches = []; inspected = 0; inspected_bytes = 0
    try:
        for oid, name in text_objects.items():
            process.stdin.write(oid.encode()+b'\n'); process.stdin.flush()
            fields = process.stdout.readline().split()
            assert len(fields) == 3 and fields[1] == b'blob', fields
            size = int(fields[2]); remaining = size; overlap = b''; hit = False
            while remaining:
                block = process.stdout.read(min(1024*1024, remaining))
                assert block
                hit = hit or bool(pattern.search(overlap+block))
                overlap = block[-256:]; remaining -= len(block)
            assert process.stdout.read(1) == b'\n'
            inspected += 1; inspected_bytes += size
            if hit: matches.append(dict(path=name, git_blob=oid))
    finally:
        process.stdin.close(); process.wait(timeout=10)
    assert not matches, 'review possible historical credential matches: '+str(matches)
    hd = [x for x in assets if x['path'].startswith('assets/hd/')]
    music = [x for x in assets if x['path'].startswith('assets/audio/hd/')]
    assert len(hd) == 832 and len(music) == 6
    report = dict(schema='kol-public-repository-audit/1', date='2026-10-17', status='PREPARED',
        repository='wicanr2/kol-remake-cht', source_head=head,
        visibility_before='private', visibility_requested='public',
        user_request='github repo 變成 public , hd 素材都要進入 github repo',
        scope='既有儲存庫及可達歷史公開；正式遊戲用HD圖像、配套清冊、四首HD配樂全部收錄。',
        copyright_status='公開收錄由使用者明確指定；未確認原作散布授權，不變更原作或第三方權利。',
        original_license_granted=False, source_license='RRSAL-1.0', license_sha256=sha(ROOT/'LICENSE'),
        hd_files=len(hd), hd_png_files=sum(Path(x['path']).suffix=='.png' for x in hd),
        music_files=len(music), music_ogg_files=sum(Path(x['path']).suffix=='.ogg' for x in music),
        total_files=len(assets), total_bytes=sum(x['bytes'] for x in assets),
        all_files_tracked=True, all_git_blobs_match=True, all_release_resource_hashes_match=True,
        resource_receipt=str(resources_path.relative_to(ROOT)), resource_receipt_sha256=sha(resources_path),
        historical_forbidden_paths=paths,
        credential_pattern_scan=dict(scope='all reachable historical text blobs; limited to four documented patterns',
            text_blobs=inspected, bytes=inspected_bytes, matches=matches,
            completeness_claim='不替代完整秘密掃描；只記錄本次實際檢查樣式與文字範圍。'),
        github_actions_runs=0, github_actions_source='host gh api repos/wicanr2/kol-remake-cht/actions/runs',
        legacy_manifests_preserved=True, source_asset_bytes_changed=False,
        full_packages_uploaded=False, promo_uploaded=False, published_tags_moved=False,
        tool=dict(path='tools/audit_public_repository.py',sha256=sha(Path(__file__)),
                  image='kol-package:1.26.7-ebiten2.9.9-20261011-r1'),
        assets=assets)
    args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps({k:report[k] for k in ['status','hd_files','hd_png_files','music_files','music_ogg_files','total_files','total_bytes']},ensure_ascii=False))
    print('HISTORY TEXT SCAN',inspected,inspected_bytes,'credential-pattern matches',len(matches))

if __name__ == '__main__': main()
