#!/usr/bin/env python3
"""Docker內補建已驗完整版的三平台無素材引擎包；不重編或覆寫完整版。"""
import argparse
import hashlib
import json
from pathlib import Path
import plistlib
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

import repack_full_variants as full

ROOT = Path('/src')
VERSION = 'v.1.0.5-20261018'


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def read(path):
    return json.loads(path.read_text())


def run(args, **kwargs):
    return subprocess.run(list(map(str, args)), check=True, timeout=240, **kwargs)


def prepare(out, work):
    assert not any((out / 'release').iterdir()), '拒絕覆寫既有引擎交付'
    assert not work.exists()
    proof = read(ROOT / 'docs/verification/repack-v1.0.5-20261018/receipt.json')
    assert proof['status'] == 'PASS' and proof['version'] == VERSION
    assert full.program_inputs() == proof['program_inputs'], '正式程式已改變'
    for row in proof['packages']:
        path = out / row['path']
        assert path.stat().st_size == row['bytes'] and full.sha(path) == row['sha256']
    inputs = read(out / 'build-inputs.json')
    work.mkdir()
    (work / 'smoke').mkdir()
    for platform, name in (('windows-amd64', 'kol.exe'), ('macos-universal', 'kol-macos')):
        package = out / 'full' / f'kol-{VERSION}-full-ega-hd-{platform}.zip'
        suffix = '/kol.exe' if platform.startswith('windows') else '/Contents/MacOS/kol'
        with zipfile.ZipFile(package) as z:
            member = next(n for n in z.namelist() if n.endswith(suffix))
            with z.open(member) as source, (work / name).open('wb') as dest:
                shutil.copyfileobj(source, dest)
        assert full.sha(work / name) == inputs['binaries'][platform]['sha256']
    with tempfile.TemporaryDirectory(prefix='kol-release-input-') as tmp:
        path = out / 'full' / f'kol-{VERSION}-full-ega-linux-x86_64.AppImage'
        run([path, '--appimage-extract'], cwd=tmp, stdout=subprocess.DEVNULL)
        shutil.copy2(Path(tmp) / 'squashfs-root/usr/bin/kol', work / 'kol')
    assert full.sha(work / 'kol') == inputs['binaries']['linux-x86_64']['sha256']
    # 沿已核准的引擎包內容，工具與必要文件須仍與目前來源完全相同。
    common = {}
    old = ROOT / 'dist-all/v.1.0.4-20261017/release/kol-v.1.0.4-20261017-release-ega-windows-amd64.zip'
    with zipfile.ZipFile(old) as z:
        for name in z.namelist():
            relative = name.split('/', 1)[1]
            if not (relative == 'LICENSE' or relative.startswith(('docs/', 'tools/'))):
                continue
            source = ROOT / relative
            assert source.is_file() and source.read_bytes() == z.read(name), relative
            common[relative] = full.sha(source)
    record = dict(version=VERSION, source_tag='bc087d31183407907c754e252727b195d860aa9f',
        full_receipt_sha256=full.sha(ROOT / 'docs/verification/repack-v1.0.5-20261018/receipt.json'),
        program_inputs=inputs['program_inputs'], binaries=inputs['binaries'], resources=inputs['resources'],
        common_inputs=common, full_packages=proof['packages'], binary_rebuilt=False)
    write(out / 'release-inputs.json', record)
    print('PREPARED: exact validated binaries; full packages unchanged', flush=True)


def instructions(theme, platform):
    label = '純 EGA' if theme == 'ega' else 'EGA＋HD'
    if platform == 'linux-x86_64':
        how = f'''先將AppImage加上執行權限，執行 --appimage-extract。
匯入：python3 squashfs-root/tools/import_game_data.py --source "合法抽取資料目錄" --target "$PWD/game-data"
啟動：KOL_DATA_DIR="$PWD/game-data" ./kol-{VERSION}-release-{theme}-linux-x86_64.AppImage --appimage-extract-and-run
存檔：$XDG_CONFIG_HOME/kol-remake-cht/save.dat，未設定時 ~/.config/kol-remake-cht/save.dat。'''
    elif platform == 'macos-universal':
        how = '''匯入：python3 武士傳說.app/Contents/Resources/tools/import_game_data.py --source "合法抽取資料目錄" --target 武士傳說.app/Contents/Resources
完成後開啟武士傳說.app，支援Apple Silicon／Intel，macOS 11以上。
沒有Apple開發者簽章，首次開啟需依macOS安全設定允許。
存檔：~/Library/Application Support/kol-remake-cht/save.dat。'''
    else:
        how = '''先解壓，在本包目錄執行：python tools/import_game_data.py --source "合法抽取資料目錄" --target "."
完成後開啟kol.exe。
存檔：%AppData%\\kol-remake-cht\\save.dat；雙擊啟動日誌為同目錄kol.log。'''
    return f'''《武士傳說》{VERSION} {label} 引擎包
本包不含遊戲圖像、配樂或原版資料，須自行匯入合法持有且已按本專案流程抽取的資料。
資料來源須有assets目錄，所有檔案須符合import-manifest.json的大小與SHA-256。
{how}
純EGA匯入912份資料，不複製手繪圖像，F10的HD選項停用。
EGA＋HD匯入1744份資料，可在遊戲中以F10切換；四首配樂以F9獨立開關。
三平台均按需讀圖，快取有容量上限，久未使用及回EGA時釋放；首次HD仍需載入。
方向鍵移動，Enter執行，Esc返回，Ctrl+Q離開；住宿保存，kol-save-v3格式保持。
Windows使用Wine驗證，macOS僅結構驗證，尚無真機驗收；Linux主機FUSE未驗。
引擎授權見LICENSE。原作及第三方素材權利不變，匯入後的資料僅本機使用。
'''


def assemble(out, work):
    inputs = read(out / 'release-inputs.json')
    icon = work / 'kol.png'
    full.run(['convert', '-size', '256x256', 'xc:#140606', '-fill', '#c8992e', '-draw',
              'polygon 128,18 220,60 220,140 128,238 36,140 36,60', icon])
    packages = []
    for theme in ('ega', 'ega-hd'):
        selected = [r for r in inputs['resources'] if theme == 'ega-hd' or not r['path'].startswith('assets/hd/')]
        for platform, binary in (('linux-x86_64', 'kol'), ('windows-amd64', 'kol.exe'), ('macos-universal', 'kol-macos')):
            stem = f'kol-{VERSION}-release-{theme}-{platform}'
            with tempfile.TemporaryDirectory(prefix='kol-engine-') as tmp:
                staged = Path(tmp) / stem
                staged.mkdir()
                base = staged
                if platform == 'macos-universal':
                    contents = staged / '武士傳說.app/Contents'
                    base = contents / 'Resources'
                    base.mkdir(parents=True)
                    (contents / 'MacOS').mkdir()
                    shutil.copy2(work / binary, contents / 'MacOS/kol')
                    (contents / 'MacOS/kol').chmod(0o755)
                    (contents / 'Info.plist').write_bytes(plistlib.dumps(dict(
                        CFBundleName='武士傳說', CFBundleDisplayName='武士傳說', CFBundleExecutable='kol',
                        CFBundleIdentifier='tw.wicanr2.kol-remake-cht', CFBundleVersion='1.0.5',
                        CFBundleShortVersionString='1.0.5', KOLReleaseVersion=VERSION,
                        CFBundlePackageType='APPL', CFBundleIconFile='kol',
                        LSMinimumSystemVersion='11.0', NSHighResolutionCapable=True)))
                    full.run(['convert', icon, base / 'kol.icns'])
                elif platform == 'linux-x86_64':
                    (base / 'usr/bin').mkdir(parents=True)
                    shutil.copy2(work / binary, base / 'usr/bin/kol')
                    (base / 'AppRun').write_text('''#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
if [ -n "${KOL_DATA_DIR:-}" ]; then cd "$KOL_DATA_DIR" || exit 1; else cd "$HERE" || exit 1; fi
exec "$HERE/usr/bin/kol" "$@"
''')
                    (base / 'AppRun').chmod(0o755)
                    (base / 'kol.desktop').write_text('[Desktop Entry]\nType=Application\nName=武士傳說\nExec=AppRun\nIcon=kol\nCategories=Game;RolePlaying;\n')
                    shutil.copy2(icon, base / 'kol.png')
                    shutil.copy2(icon, base / '.DirIcon')
                else:
                    shutil.copy2(work / binary, base / binary)
                for name, expected in inputs['common_inputs'].items():
                    assert full.sha(ROOT / name) == expected
                    dest = base / name
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(ROOT / name, dest)
                (base / 'README.txt').write_bytes(b'\xef\xbb\xbf' + instructions(theme, platform).replace('\n', '\r\n').encode())
                (base / 'KNOWN-LIMITATIONS.txt').write_text('Windows／macOS尚無真機驗收；完整原版parity及自然全任務通關未宣稱。\n')
                (base / 'VERSION').write_text(VERSION + '\n')
                write(base / 'import-manifest.json', dict(schema='kol-required-data/1', version=VERSION,
                    graphics=theme, source='使用者合法持有且已抽取的遊戲資料', files=selected,
                    rights='只有檔案大小與雜湊，不含素材位元組；匯入後資料不得散布。'))
                if platform == 'macos-universal':
                    for name in ('README.txt', 'LICENSE'):
                        shutil.copy2(base / name, staged / name)
                files = [dict(path=str(p.relative_to(staged)), bytes=p.stat().st_size, sha256=full.sha(p))
                         for p in sorted(staged.rglob('*')) if p.is_file()]
                write(base / 'manifest.json', dict(schema='kol-release-variant/1', version=VERSION, bundle='release',
                    graphics=theme, platform=platform, files=files, required_resource_files=len(selected)))
                suffix = '.AppImage' if platform == 'linux-x86_64' else '.zip'
                target = out / 'release' / (stem + suffix)
                assert not target.exists()
                if suffix == '.AppImage':
                    squash = Path(tmp) / 'engine.squashfs'
                    with (work / (stem + '-squashfs.log')).open('w') as log:
                        full.run(['mksquashfs', staged, squash, '-noappend', '-comp', 'gzip', '-processors', '2'], stdout=log, stderr=subprocess.STDOUT)
                    with target.open('wb') as dest:
                        for source in (Path('/opt/runtime-x86_64'), squash):
                            with source.open('rb') as stream:
                                shutil.copyfileobj(stream, dest)
                    target.chmod(0o755)
                else:
                    full.make_zip(target, staged, stem, (2026, 10, 18))
                packages.append(dict(path=str(target.relative_to(out)), bytes=target.stat().st_size,
                    sha256=full.sha(target), platform=platform, graphics=theme, bundle='release',
                    required_resource_files=len(selected), assets_included=False))
                print('PACKAGED', target.name, target.stat().st_size, flush=True)
    write(out / 'release/SHA256SUMS.json', dict(version=VERSION, files=packages))
    (out / 'release/SHA256SUMS.txt').write_text(''.join(f"{r['sha256']}  {Path(r['path']).name}\n" for r in packages))


def verify(out, work):
    inputs = read(out / 'release-inputs.json')
    packages = read(out / 'release/SHA256SUMS.json')['files']
    results = []
    for row in packages:
        package = out / row['path']
        assert full.sha(package) == row['sha256'] and package.stat().st_size == row['bytes']
        with tempfile.TemporaryDirectory(prefix='kol-engine-check-') as tmp:
            dest = Path(tmp)
            if package.suffix == '.AppImage':
                run([package, '--appimage-extract'], cwd=dest, stdout=subprocess.DEVNULL)
                staged = dest / 'squashfs-root'
                base, executable = staged, staged / 'usr/bin/kol'
            else:
                with zipfile.ZipFile(package) as z:
                    assert z.testzip() is None
                    z.extractall(dest)
                    if row['platform'] == 'macos-universal':
                        exe = next(n for n in z.namelist() if n.endswith('/Contents/MacOS/kol'))
                        assert (z.getinfo(exe).external_attr >> 16) & 0o111
                staged = next(dest.iterdir())
                if row['platform'] == 'macos-universal':
                    base = staged / '武士傳說.app/Contents/Resources'
                    executable = staged / '武士傳說.app/Contents/MacOS/kol'
                    plist = plistlib.loads((staged / '武士傳說.app/Contents/Info.plist').read_bytes())
                    assert plist['KOLReleaseVersion'] == VERSION and plist['LSMinimumSystemVersion'] == '11.0'
                else:
                    base, executable = staged, staged / 'kol.exe'
            assert full.sha(executable) == inputs['binaries'][row['platform']]['sha256']
            manifest = read(base / 'manifest.json')
            for item in manifest['files']:
                path = staged / item['path']
                assert path.stat().st_size == item['bytes'] and full.sha(path) == item['sha256']
            for path in staged.rglob('*'):
                if not path.is_file():
                    continue
                n = str(path.relative_to(staged))
                assert '/assets/' not in '/' + n and '/docs/manual/' not in '/' + n
                if path.suffix.lower() in ('.png', '.icns'):
                    assert n in ('kol.png', '.DirIcon', '武士傳說.app/Contents/Resources/kol.icns')
                else:
                    assert path.suffix.lower() not in ('.ogg', '.wav', '.pdf', '.jpg', '.jpeg', '.dat', '.zip')
                if path.suffix.lower() == '.exe':
                    assert path == executable
            required = read(base / 'import-manifest.json')
            expected = [r for r in inputs['resources'] if row['graphics'] == 'ega-hd' or not r['path'].startswith('assets/hd/')]
            assert required['files'] == expected and len(expected) == row['required_resource_files']
            assert (base / 'VERSION').read_text().strip() == VERSION
            assert (base / 'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
            assert (base / 'LICENSE').read_bytes() == (ROOT / 'LICENSE').read_bytes()
            if row['platform'] == 'macos-universal':
                run(['python3', base / 'tools/import_game_data.py', '--source', ROOT, '--target', base])
                assert (base / 'assets/hd').exists() == (row['graphics'] == 'ega-hd')
                assert read(base / 'data-import-receipt.json')['files'] == len(expected)
            results.append(dict(package=package.name, status='PASS', asset_free=True,
                binary_equal_to_full=True, required_resource_files=len(expected),
                macos_import_verified=row['platform'] == 'macos-universal'))
        print('VERIFIED', package.name, flush=True)
    write(work / 'package-content.json', dict(status='PASS', version=VERSION, packages=results))


def finish(out, work):
    import numpy as np
    from PIL import Image
    import io
    import wave
    inputs = read(out / 'release-inputs.json')
    assert full.program_inputs() == inputs['program_inputs']
    reports = {p: read(work / (p + '-runtime.json')) for p in ('linux', 'windows')}
    assert all(r['status'] == 'PASS' and 'oom_kill 0' in r['memory_events'] for r in reports.values())
    comparisons = []
    archive = ROOT / 'docs/verification/local/repack-20261018/evidence.tar.xz'
    retention = read(ROOT / 'docs/verification/repack-v1.0.5-20261018/retention.json')
    assert full.sha(archive) == retention['archive']['sha256']
    with tarfile.open(archive, 'r:xz') as old:
        for platform in ('linux', 'windows'):
            for theme in ('ega', 'ega-hd'):
                name = platform + '-' + theme
                for p in sorted((work / (name + '-flow')).glob('*.png')):
                    original = old.extractfile(name + '-flow/' + p.name).read()
                    a = np.asarray(Image.open(p).convert('RGB'))
                    b = np.asarray(Image.open(io.BytesIO(original)).convert('RGB'))
                    assert a.shape == b.shape
                    maximum = int(np.max(np.abs(a.astype(np.int16) - b.astype(np.int16))))
                    assert maximum == 0, (name, p.name, maximum)
                    comparisons.append(dict(platform=platform, graphics=theme, frame=p.name, maximum=maximum, sha256=full.sha(p)))
                for suffix in ('-trace.json', '-text.jsonl'):
                    assert (work / (name + suffix)).read_bytes() == old.extractfile(name + suffix).read()
    assert len(comparisons) == 24
    with wave.open(str(work / 'smoke/windows-audio.wav'), 'rb') as f:
        assert (f.getnchannels(), f.getframerate(), f.getsampwidth()) == (2, 44100, 2)
        samples = np.frombuffer(f.readframes(f.getnframes()), dtype='<i2').astype(np.int32)
    peak = int(np.max(np.abs(samples)))
    assert 100 < peak < 32767 and not np.any(samples[-88200:])
    packages = read(out / 'release/SHA256SUMS.json')['files']
    assert read(work / 'package-content.json')['status'] == 'PASS'
    for row in inputs['full_packages']:
        assert full.sha(out / row['path']) == row['sha256'], '完整版被改動'
    write(work / 'comparisons.json', dict(status='PASS', reference='同版已驗完整版', frames=comparisons, text_and_model_equal=True))
    receipt = dict(schema='kol-engine-release/1', status='PASS', version=VERSION, verified_on='2026-10-18',
        source_tag=inputs['source_tag'], program_inputs=inputs['program_inputs'], binaries=inputs['binaries'],
        compiled_binary_unchanged=True, full_packages_unchanged=True, packages=packages, runtime=reports,
        import_resource_counts=[912,1744], compared_frames=24, full_same_platform_maximum=0,
        audio=dict(peak=peak, seconds=len(samples)/88200, non_silent=True, last_second_silent=True, human_relisten=False),
        macos_validation='同版已驗universal執行檔、bundle結構及兩種實際匯入；沒有真機執行',
        windows_real_machine=False, macos_real_machine=False, original_parity_claimed=False,
        upload_allowlist=[Path(r['path']).name for r in packages]+['SHA256SUMS.txt'],
        rights='公開只上傳六個無素材引擎包及SHA256SUMS.txt；完整版與影片留本機。',
        tool_sha256={n:full.sha(ROOT/'tools'/n) for n in ('pack_release_variants.py','verify_full_variants_runtime.py','repack_full_variants.py')})
    receipt['attachments'] = {str(p.relative_to(work)): dict(size=p.stat().st_size, sha256=full.sha(p))
                              for p in sorted(work.rglob('*')) if p.is_file() and p.name != 'receipt.json'}
    write(work / 'receipt.json', receipt)
    write(out / 'SHA256SUMS.json', dict(version=VERSION, files=inputs['full_packages']+packages))
    print('ENGINE RELEASE PASS: 6 packages, imported gameplay, 24 exact frames', flush=True)


def main():
    assert Path('/.dockerenv').exists()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=('prepare', 'assemble', 'verify', 'finish'))
    parser.add_argument('--out', type=Path, required=True)
    parser.add_argument('--work', type=Path, required=True)
    args = parser.parse_args()
    assert read(args.out / 'build-inputs.json')['version'] == VERSION
    globals()[args.phase](args.out, args.work)


if __name__ == '__main__':
    main()
