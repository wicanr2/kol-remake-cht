#!/usr/bin/env python3
"""Docker 內從固定來源重建三平台純 EGA／EGA＋HD 完整版。"""
import argparse
import datetime
import hashlib
import json
from pathlib import Path
import plistlib
import re
import shutil
import subprocess
import tempfile
import wave
import zipfile

ROOT = Path('/src')


def sha(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def program_inputs():
    result = {str(p.relative_to(ROOT)): sha(p) for d in ('cmd', 'internal')
              for p in sorted((ROOT / d).rglob('*.go'))}
    result.update({n: sha(ROOT / n) for n in ('go.mod', 'go.sum')})
    return result


def prepare(args):
    assert not args.out.exists(), '拒絕覆寫交付'
    proof_path = args.acceptance or ROOT / 'docs/verification/windows-memory-20261017/receipt.json'
    proof = json.loads(proof_path.read_text())
    assert proof['status'] == 'PASS'
    program = program_inputs()
    accepted_program = proof.get('program_inputs')
    if accepted_program is None:
        source = proof_path.parent / 'program-inputs.json'
        assert sha(source) == proof['artifacts']['program-inputs.json'], '驗收來源清冊雜湊不符'
        accepted_program = json.loads(source.read_text())['program']
    assert program == accepted_program, '正式程式已改變，須重新驗收'
    baseline = ROOT / 'docs/verification/release-v1.0.3-20261017/build-inputs.json'
    resources = json.loads(baseline.read_text())['resources']
    for row in resources:
        source = ROOT / row['path']
        assert source.stat().st_size == row['bytes'] and sha(source) == row['sha256'], source
    args.out.mkdir()
    for name in ('full', 'release', 'promo', 'smoke'):
        (args.out / name).mkdir()
    args.work.mkdir()
    write(args.out / 'build-inputs.json', dict(version=args.version, source_commit=args.git_head,
          program_inputs=program, acceptance_receipt=str(proof_path.relative_to(ROOT)),
          acceptance_sha256=sha(proof_path), resources=resources,
          resource_baseline_sha256=sha(baseline), full_only=True))
    print('PREPARED', len(program), len(resources), flush=True)


def run(command, **kwargs):
    return subprocess.run(list(map(str, command)), check=True, timeout=600, **kwargs)


def instructions(version, theme, platform):
    launch, save = {
        'windows-amd64': ('解壓後開啟 kol.exe。', '%AppData%\\kol-remake-cht\\save.dat'),
        'linux-x86_64': ('賦予 AppImage 執行權限後開啟；也可使用 --appimage-extract-and-run。', '$XDG_CONFIG_HOME/kol-remake-cht/save.dat，未設定時為 ~/.config/kol-remake-cht/save.dat'),
        'macos-universal': ('解壓後開啟「武士傳說.app」。支援 Apple Silicon 及 Intel，macOS 11 以上。', '~/Library/Application Support/kol-remake-cht/save.dat'),
    }[platform]
    return f'''《武士傳說》{version} {'純 EGA' if theme == 'ega' else 'EGA＋HD'} 完整版
{launch}
已包含遊戲資料。純 EGA 不含手繪圖像，F10 的 HD 選項停用。
EGA＋HD 可在遊戲中以 F10 切換；圖片按需讀取，久未使用的快取自動釋放。
回到 EGA 時釋放 HD 圖片。首次切換 HD 仍需等候載入。
兩版均保留四首配樂，F9 獨立開關，預設開啟。
方向鍵移動及選擇選項，逗號／句號移動圖示白框，Enter 執行，Esc 返回，Ctrl+Q 離開。
地圖方向鍵可長按連走，遇到問句後須放開再按；滾輪不會跳過遭遇與戰鬥訊息。
人物／裝備頁按1至6切換隊員。名稱支援Shift及Caps Lock大小寫。
住宿會保存，重新開啟會自動讀回。另讀指定存檔可使用 -load 路徑；-load "" 開始新遊戲。
存檔：{save}
Windows 雙擊啟動日誌：%AppData%\\kol-remake-cht\\kol.log。
遊戲規則及 kol-save-v3 存檔格式保持。原版完整 parity 與自然全任務通關未宣稱。
Windows 使用 Wine 驗證；macOS 僅結構驗證，尚無真機驗收。Linux 主機 FUSE 未驗。
含原作衍生資料，僅本機保存，不得上傳或散布。引擎授權見 LICENSE。
'''


def make_zip(target, staged, stem, date):
    with zipfile.ZipFile(target, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=6, allowZip64=True) as z:
        for p in sorted(staged.rglob('*')):
            if not p.is_file():
                continue
            info = zipfile.ZipInfo(stem + '/' + str(p.relative_to(staged)), (*date, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = (0o100755 if p.stat().st_mode & 0o111 else 0o100644) << 16
            with z.open(info, 'w', force_zip64=True) as dest, p.open('rb') as source:
                shutil.copyfileobj(source, dest, 1 << 20)


def assemble(args):
    inputs = json.loads((args.out / 'build-inputs.json').read_text())
    assert inputs['program_inputs'] == program_inputs()
    binaries = {p: args.work / n for p, n in (
        ('windows-amd64', 'kol.exe'), ('linux-x86_64', 'kol'), ('macos-universal', 'kol-macos'))}
    for path in binaries.values():
        assert path.is_file()
    inputs['binaries'] = {p: dict(bytes=b.stat().st_size, sha256=sha(b)) for p, b in binaries.items()}
    write(args.out / 'build-inputs.json', inputs)
    date = datetime.datetime.strptime(args.version[-8:], '%Y%m%d').timetuple()[:3]
    icon = args.work / 'kol.png'
    run(['convert', '-size', '256x256', 'xc:#140606', '-fill', '#c8992e', '-draw',
         'polygon 128,18 220,60 220,140 128,238 36,140 36,60', icon])
    packages = []
    for theme in ('ega', 'ega-hd'):
        selected = [r for r in inputs['resources'] if theme == 'ega-hd' or not r['path'].startswith('assets/hd/')]
        for platform, binary in binaries.items():
            stem = f'kol-{args.version}-full-{theme}-{platform}'
            with tempfile.TemporaryDirectory(prefix='kol-full-', dir='/tmp') as tmp:
                staged = Path(tmp) / stem
                staged.mkdir()
                base = staged
                if platform == 'macos-universal':
                    contents = staged / '武士傳說.app/Contents'
                    base = contents / 'Resources'
                    (contents / 'MacOS').mkdir(parents=True)
                    base.mkdir()
                    shutil.copy2(binary, contents / 'MacOS/kol')
                    (contents / 'MacOS/kol').chmod(0o755)
                    (contents / 'Info.plist').write_bytes(plistlib.dumps(dict(
                        CFBundleName='武士傳說', CFBundleDisplayName='武士傳說', CFBundleExecutable='kol',
                        CFBundleIdentifier='tw.wicanr2.kol-remake-cht', CFBundleVersion=args.version[2:].split('-')[0],
                        CFBundleShortVersionString=args.version[2:].split('-')[0], KOLReleaseVersion=args.version,
                        CFBundlePackageType='APPL', CFBundleIconFile='kol',
                        LSMinimumSystemVersion='11.0', NSHighResolutionCapable=True)))
                    run(['convert', icon, base / 'kol.icns'])
                elif platform == 'linux-x86_64':
                    (base / 'usr/bin').mkdir(parents=True)
                    shutil.copy2(binary, base / 'usr/bin/kol')
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
                    shutil.copy2(binary, base / 'kol.exe')
                for name in ('LICENSE', 'docs/i18n', 'docs/manual', 'docs/re/title-sequence.json',
                             'docs/re/town-inn-sample.json', 'docs/re/town-outdoor-sample.json',
                             'docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json'):
                    source, dest = ROOT / name, base / name
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    if source.is_dir():
                        shutil.copytree(source, dest)
                    else:
                        shutil.copy2(source, dest)
                for row in selected:
                    dest = base / row['path']
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(ROOT / row['path'], dest)
                (base / 'VERSION').write_text(args.version + '\n')
                (base / 'README.txt').write_bytes(b'\xef\xbb\xbf' + instructions(args.version, theme, platform).replace('\n', '\r\n').encode())
                (base / 'DO-NOT-DISTRIBUTE.txt').write_text('含原作衍生資料，僅本機保存，不得上傳或散布。\n')
                if platform == 'macos-universal':
                    for name in ('README.txt', 'LICENSE', 'DO-NOT-DISTRIBUTE.txt'):
                        shutil.copy2(base / name, staged / name)
                rows = [dict(path=str(p.relative_to(staged)), bytes=p.stat().st_size, sha256=sha(p))
                        for p in sorted(staged.rglob('*')) if p.is_file()]
                write(base / 'manifest.json', dict(schema='kol-full-variant/1', version=args.version,
                      platform=platform, graphics=theme, bundle='full', local_only=True,
                      resource_files=len(selected), files=rows))
                suffix = '.AppImage' if platform == 'linux-x86_64' else '.zip'
                target = args.out / 'full' / (stem + suffix)
                assert not target.exists()
                if suffix == '.AppImage':
                    squash = Path(tmp) / 'data.squashfs'
                    with (args.out / 'smoke' / (stem + '-squashfs.log')).open('w') as log:
                        run(['mksquashfs', staged, squash, '-noappend', '-comp', 'gzip', '-processors', '2'], stdout=log, stderr=subprocess.STDOUT)
                    with target.open('wb') as stream:
                        for source in (Path('/opt/runtime-x86_64'), squash):
                            with source.open('rb') as src:
                                shutil.copyfileobj(src, stream, 1 << 20)
                    target.chmod(0o755)
                else:
                    make_zip(target, staged, stem, date)
                packages.append(dict(path=str(target.relative_to(args.out)), bytes=target.stat().st_size,
                    sha256=sha(target), graphics=theme, platform=platform, bundle='full',
                    resource_files=len(selected), local_only=True))
                print('PACKAGED', target.name, target.stat().st_size, flush=True)
    write(args.out / 'SHA256SUMS.json', dict(version=args.version, files=packages))
    (args.out / 'full/SHA256SUMS.txt').write_text(''.join(f"{p['sha256']}  {Path(p['path']).name}\n" for p in packages))


def verify(args):
    inputs = json.loads((args.out / 'build-inputs.json').read_text())
    packages = json.loads((args.out / 'SHA256SUMS.json').read_text())['files']
    results = []
    for row in packages:
        path = args.out / row['path']
        assert path.stat().st_size == row['bytes'] and sha(path) == row['sha256']
        with tempfile.TemporaryDirectory(prefix='kol-verify-') as tmp:
            root = Path(tmp)
            if path.suffix == '.AppImage':
                run([path, '--appimage-extract'], cwd=root, stdout=subprocess.DEVNULL)
                staged = root / 'squashfs-root'
                executable, base = staged / 'usr/bin/kol', staged
                assert (staged / 'AppRun').stat().st_mode & 0o111
            else:
                with zipfile.ZipFile(path) as z:
                    assert z.testzip() is None
                    z.extractall(root)
                    for info in z.infolist():
                        if not info.filename.isascii():
                            assert info.flag_bits & 0x800
                    staged = next(root.iterdir())
                    if row['platform'] == 'macos-universal':
                        name = next(n for n in z.namelist() if n.endswith('/Contents/MacOS/kol'))
                        assert (z.getinfo(name).external_attr >> 16) & 0o111
                base = staged / '武士傳說.app/Contents/Resources' if row['platform'] == 'macos-universal' else staged
                executable = staged / '武士傳說.app/Contents/MacOS/kol' if row['platform'] == 'macos-universal' else staged / 'kol.exe'
                if row['platform'] == 'macos-universal':
                    plist = plistlib.loads((staged / '武士傳說.app/Contents/Info.plist').read_bytes())
                    assert plist['KOLReleaseVersion'] == args.version and plist['LSMinimumSystemVersion'] == '11.0'
            assert sha(executable) == inputs['binaries'][row['platform']]['sha256']
            manifest = json.loads((base / 'manifest.json').read_text())
            assert manifest['version'] == args.version and manifest['graphics'] == row['graphics']
            for item in manifest['files']:
                p = staged / item['path']
                assert p.stat().st_size == item['bytes'] and sha(p) == item['sha256'], item['path']
            assert (base / 'VERSION').read_text().strip() == args.version
            assert (base / 'LICENSE').read_bytes() == (ROOT / 'LICENSE').read_bytes()
            assert (base / 'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
            assert (base / 'assets/hd').exists() == (row['graphics'] == 'ega-hd')
            selected = [r for r in inputs['resources'] if row['graphics'] == 'ega-hd' or not r['path'].startswith('assets/hd/')]
            assert len(selected) == row['resource_files']
            for item in selected:
                assert sha(base / item['path']) == item['sha256']
            assert not [p for p in staged.rglob('*') if p.is_file() and p.suffix.lower() in ('.pdf', '.wav', '.dat', '.rom')]
            results.append(dict(package=path.name, status='PASS', manifest_files=len(manifest['files']),
                resource_files=len(selected), hashes_checked=True, binary_hash_verified=True))
        print('VERIFIED', path.name, flush=True)
    write(args.out / 'smoke/package-content.json', dict(status='PASS', version=args.version, packages=results))


def finish(args):
    import numpy as np
    from PIL import Image
    inputs = json.loads((args.out / 'build-inputs.json').read_text())
    assert inputs['program_inputs'] == program_inputs()
    content = json.loads((args.out / 'smoke/package-content.json').read_text())
    assert content['status'] == 'PASS' and len(content['packages']) == 6
    runtimes = {}
    for platform in ('linux', 'windows'):
        report = json.loads((args.work / (platform + '-runtime.json')).read_text())
        assert report['status'] == 'PASS' and report['version'] == args.version
        assert 'oom_kill 0' in report['memory_events'] and len(report['runs']) == 2
        runtimes[platform] = report
    comparisons = []
    for theme in ('ega', 'ega-hd'):
        left = args.work / ('linux-' + theme + '-flow')
        right = args.work / ('windows-' + theme + '-flow')
        paths = sorted(left.glob('*.png'))
        assert len(paths) == 6 and [p.name for p in paths] == sorted(p.name for p in right.glob('*.png'))
        for p in paths:
            a, b = np.asarray(Image.open(p).convert('RGB')), np.asarray(Image.open(right / p.name).convert('RGB'))
            assert a.shape == b.shape
            maximum = int(np.max(np.abs(a.astype(np.int16) - b.astype(np.int16))))
            assert maximum <= (0 if theme == 'ega' else 1), (theme, p.name, maximum)
            comparisons.append(dict(graphics=theme, frame=p.name, maximum=maximum,
                linux_sha256=sha(p), windows_sha256=sha(right / p.name)))
        for suffix in ('-trace.json', '-text.jsonl'):
            assert (args.work / ('linux-' + theme + suffix)).read_bytes() == (args.work / ('windows-' + theme + suffix)).read_bytes(), suffix
        text = (args.work / ('linux-' + theme + '-text.jsonl')).read_text()
        drawn = {json.loads(line)['text'] for line in text.splitlines() if line}
        assert {'目錄', '建立角色', '新角色', '人類'} <= drawn, '流程缺少預期中文'
        # HD 書封中文先由 DrawNativeSubtitle 繪製，不進點陣 f.Draw 紀錄。
        if theme == 'ega':
            assert '武士傳說' in drawn
    mac = {}
    architecture = (args.work / 'macos-architectures.txt').read_text()
    assert 'arm64' in architecture and 'x86_64' in architecture
    for arch in ('arm64', 'x86_64'):
        load = (args.work / f'macos-{arch}-load.txt').read_text()
        libs = (args.work / f'macos-{arch}-libs.txt').read_text()
        assert re.search(r'\bminos 11\.0\b', load)
        if arch == 'arm64':
            assert 'LC_CODE_SIGNATURE' in load
        dependencies = [line.strip().split()[0] for line in libs.splitlines()[1:] if line.strip()]
        assert dependencies and all(n.startswith(('/usr/lib/', '/System/Library/')) for n in dependencies)
        mac[arch] = dict(minimum='11.0', code_signature_present='LC_CODE_SIGNATURE' in load, dependencies=dependencies)
    with wave.open(str(args.work / 'smoke/windows-audio.wav'), 'rb') as f:
        assert f.getnchannels() == 2 and f.getframerate() == 44100 and f.getsampwidth() == 2
        samples = np.frombuffer(f.readframes(f.getnframes()), dtype='<i2')
    peak = int(np.max(np.abs(samples.astype(np.int32))))
    assert 100 < peak < 32767
    assert not np.any(samples[-88200:]), '退出後仍有音訊'
    audio = dict(seconds=len(samples)/88200, peak=peak, non_silent=True, last_second_silent=True,
                 sha256=sha(args.work / 'smoke/windows-audio.wav'), human_relisten=False)
    write(args.work / 'comparisons.json', dict(status='PASS', frames=comparisons, text_and_model_equal=True))
    write(args.work / 'macos-structure.json', dict(status='PASS', architectures=mac, real_machine=False))
    write(args.work / 'audio-check.json', audio)
    receipt = dict(schema='kol-full-repack/1', status='PASS', verified_on=args.version[-8:-4]+'-'+args.version[-4:-2]+'-'+args.version[-2:],
        version=args.version, source_commit=inputs['source_commit'], program_inputs=inputs['program_inputs'],
        program_files=len(inputs['program_inputs']), resource_files=len(inputs['resources']),
        resources_unchanged=True, memory_fix_reused=True, save_format_unchanged=True,
        player_feedback_fixes_included='player-feedback-' in inputs['acceptance_receipt'],
        prior_acceptance=inputs['acceptance_receipt'], prior_acceptance_sha256=inputs['acceptance_sha256'],
        packages=json.loads((args.out / 'SHA256SUMS.json').read_text())['files'],
        runtime=runtimes, compared_frames=len(comparisons)*2,
        ega_max=max(r['maximum'] for r in comparisons if r['graphics']=='ega'),
        hd_max=max(r['maximum'] for r in comparisons if r['graphics']=='ega-hd'),
        text_and_model_equal=True, macos=mac, macos_real_machine=False,
        windows_real_machine=False, audio=audio, publication_scope='六個完整版僅本機，未上傳；影片另有獨立驗收。',
        tools={n: sha(ROOT/'tools'/n) for n in ('repack_full_build.sh','repack_full_variants.py','verify_full_variants_runtime.py')})
    receipt['attachments'] = {str(p.relative_to(args.work)): dict(size=p.stat().st_size, sha256=sha(p))
        for p in sorted(args.work.rglob('*')) if p.is_file() and p.name != 'receipt.json'}
    write(args.work / 'receipt.json', receipt)
    for name in ('receipt.json', 'comparisons.json', 'macos-structure.json', 'audio-check.json', 'linux-runtime.json', 'windows-runtime.json'):
        shutil.copy2(args.work / name, args.out / 'smoke' / name)
    print('SIX FULL PACKAGES PASS', flush=True)


def main():
    assert Path('/.dockerenv').exists(), '只允許在 Docker 執行'
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=('prepare', 'assemble', 'verify', 'finish'))
    parser.add_argument('--version', required=True)
    parser.add_argument('--out', type=Path, required=True)
    parser.add_argument('--work', type=Path, required=True)
    parser.add_argument('--git-head')
    parser.add_argument('--acceptance', type=Path, help='已通過且鎖定程式來源的驗收收據')
    args = parser.parse_args()
    assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}', args.version)
    globals()[args.phase](args)


if __name__ == '__main__':
    main()
