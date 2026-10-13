#!/usr/bin/env python3
"""容器內驗證三平台封包；入口見 docs/verification/release-v1.0.2-20261013.md。"""
import argparse
import hashlib
import json
import re
import os
from pathlib import Path
import shutil
import struct
import subprocess
import tarfile
import time
import zipfile

ROOT, OUT = Path('/src'), Path('/out')


def sha(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def run(argv, cwd, log, timeout=120):
    with log.open('w') as stream:
        subprocess.run(list(map(str, argv)), cwd=cwd, stdout=stream,
                       stderr=subprocess.STDOUT, check=True, timeout=timeout)


def check_version(command, cwd, log):
    result = subprocess.run(list(map(str, command)) + ['-version'], cwd=cwd,
                            capture_output=True, check=True, timeout=60)
    log.write_bytes(result.stdout + result.stderr)
    assert result.stdout.decode().strip() == VERSION


def structures():
    exe = (OUT / 'full/kol.exe').read_bytes()
    assert exe[:2] == b'MZ'
    offset = struct.unpack_from('<I', exe, 0x3c)[0]
    assert exe[offset:offset+4] == b'PE\0\0'
    assert struct.unpack_from('<H', exe, offset+4)[0] == 0x8664
    assert (OUT / 'full/kol').read_bytes()[:5] == b'\x7fELF\x02'
    architectures = (OUT / 'smoke/macos-architectures.txt').read_text()
    assert 'arm64' in architectures and 'x86_64' in architectures
    for arch in ('arm64', 'x86_64'):
        load = (OUT / f'smoke/macos-{arch}-load.txt').read_text()
        if arch == 'arm64':
            assert 'LC_CODE_SIGNATURE' in load
        assert 'minos 11.0' in load
        lines = (OUT / f'smoke/macos-{arch}-libs.txt').read_text().splitlines()[1:]
        assert lines and all(s.strip().split()[0].startswith(('/usr/lib/', '/System/Library/')) for s in lines)
    write(OUT / 'smoke/structures.json', dict(status='PASS', linux='ELF64', windows='AMD64 PE',
          macos=['arm64', 'x86_64'], arm64_ad_hoc_signature=True, macos_execution=False))


def linux():
    base = OUT / 'smoke'
    extracted = Path('/tmp/release-extract')
    extracted.mkdir()
    image = OUT / 'release' / f'kol-{VERSION}-release-linux-x86_64.AppImage'
    run([image, '--appimage-extract'], extracted, base / 'release-extract.log')
    engine = extracted / 'squashfs-root'
    assert not (engine / 'assets').exists() and not (engine / 'docs/manual').exists()
    # 錯誤來源須在目標寫入之前被拒絕。
    rejected = Path('/tmp/rejected-import')
    result = subprocess.run(['python3', str(engine / 'tools/import_game_data.py'),
                             '--source', '/tmp/missing-data', '--target', str(rejected)],
                            capture_output=True, timeout=30)
    (base / 'import-negative.log').write_bytes(result.stdout + result.stderr)
    assert result.returncode != 0 and not rejected.exists()
    target = Path('/tmp/imported')
    run(['python3', engine / 'tools/import_game_data.py', '--source', OUT / 'full',
         '--target', target], Path('/tmp'), base / 'release-import.log')
    required = json.loads((OUT / 'release/import-manifest.json').read_text())['files']
    assert all(sha(target / row['path']) == row['sha256'] for row in required)
    binary = engine / 'usr/bin/kol'
    check_version([binary], target, base / 'release-version.log')
    for graphics in ('ega', 'hd'):
        run([binary, '-scene', 'cover', '-lang', 'zh-Hant', '-graphics', graphics,
             '-shot', base / f'release-cover-{graphics}.png'], target, base / f'release-{graphics}.log')
    gui = base / 'linux-gui'
    gui.mkdir()
    (gui / 'inputs').mkdir()
    retained = ROOT / 'docs/verification/local/display-switch-runtime-20261012/evidence.tar.xz'
    with tarfile.open(retained) as archive:
        for name in ('save.json', 'party6.json', 'stay-v3.json'):
            (gui / 'inputs' / name).write_bytes(archive.extractfile('inputs/' + name).read())
    wrapper = gui / 'run-full'
    wrapper.write_text('#!/bin/sh\nexec /delivery/full/kol-' + VERSION + '-full-linux-x86_64.AppImage --appimage-extract-and-run "$@"\n')
    wrapper.chmod(0o755)
    write(base / 'release-smoke.json', dict(status='PASS', version=VERSION, imported_files=len(required),
          source=sha(image), invalid_import_rejected_before_writing=True, ega_hd_startup=True))
    # 同一容器內使用剛匯入的引擎包，錄製full／release的實際音訊與控制。
    run(['python3', ROOT/'tools/verify_package_controls.py'], ROOT, base/'linux-controls.log', timeout=180)


def winpath(path):
    return 'Z:' + str(path).replace('/', '\\')


def windows():
    base = OUT / 'smoke'
    wine = '/usr/lib/wine/wine64'
    fixture = Path('/tmp/windows-input-save.json')
    with tarfile.open(ROOT / 'docs/verification/local/display-switch-runtime-20261012/evidence.tar.xz') as archive:
        fixture.write_bytes(archive.extractfile('inputs/save.json').read())
    checks = []
    for bundle in ('full', 'release'):
        unpacked = Path('/tmp') / bundle
        unpacked.mkdir()
        package = OUT / bundle / f'kol-{VERSION}-{bundle}-windows-amd64.zip'
        with zipfile.ZipFile(package) as archive:
            archive.extractall(unpacked)
        engine = next(unpacked.iterdir())
        if bundle == 'release':
            assert not (engine / 'assets').exists() and not (engine / 'docs/manual').exists()
            run(['python3', engine / 'tools/import_game_data.py', '--source', OUT / 'full',
                 '--target', engine], engine, base / 'wine-release-import.log')
        command = [wine, winpath(engine / 'kol.exe')]
        check_version(command, engine, base / f'wine-{bundle}-version.log')
        for graphics in ('ega', 'hd'):
            run(command + ['-scene', 'play', '-lang', 'zh-Hant', '-graphics', graphics,
                           '-load', winpath(fixture), '-seed', '1',
                           '-play-script', 'L R RET SPACE', '-shot-steps', '0,1,2,3,4',
                           '-shot-dir', winpath(base / f'wine-{bundle}-{graphics}')],
                engine, base / f'wine-{bundle}-{graphics}.log')
        # 使用真實視窗，驗 F10、模式選擇及返回；以 Draw 完成事件作同步。
        log = base / f'wine-{bundle}-gui.log'
        with log.open('w') as stream:
            proc = subprocess.Popen(command + ['-scene', 'game', '-lang', 'zh-Hant', '-graphics', 'ega',
                '-scale', '2', '-seed', '1', '-load', winpath(fixture),
                '-save', winpath(Path('/tmp') / f'{bundle}-gui-save.json'), '-log-input'],
                cwd=engine, stdout=stream, stderr=subprocess.STDOUT)
            def wait_for(text):
                for _ in range(150):
                    if text in log.read_text():
                        return
                    assert proc.poll() is None, log.read_text()
                    time.sleep(.1)
                raise AssertionError(('window event timeout', text, log.read_text()))
            def key(name):
                subprocess.run(['xdotool', 'keydown', name], check=True, timeout=10)
                time.sleep(.18)
                subprocess.run(['xdotool', 'keyup', name], check=True, timeout=10)
                time.sleep(.18)
            try:
                for _ in range(150):
                    found = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name', '^武士傳說$'], capture_output=True, text=True)
                    if found.returncode == 0:
                        window = found.stdout.splitlines()[-1]
                        break
                    assert proc.poll() is None, log.read_text()
                    time.sleep(.1)
                else:
                    raise AssertionError(log.read_text())
                subprocess.run(['xdotool', 'windowfocus', '--sync', window], check=True, timeout=10)
                for token in ('Left','Right','Return','space'):
                    key(token)
                wait_for('music cue=inn enabled=true')
                time.sleep(1.5)
                key('F9'); wait_for('display action=music-false')
                key('F9'); wait_for('display action=music-true')
                key('F10'); wait_for('display-frame hd=false open=true')
                key('Down'); key('Return'); wait_for('display-frame hd=true open=true')
                key('Escape'); wait_for('display-frame hd=true open=false')
                key('F10'); key('Up'); key('Return')
                wait_for('display action=ega hd=false')
                subprocess.run(['xdotool', 'keydown', 'ctrl', 'q'], check=True, timeout=10)
                proc.wait(timeout=15)
                subprocess.run(['xdotool', 'keyup', 'q', 'ctrl'], check=True, timeout=10)
                assert proc.returncode == 0
            finally:
                if proc.poll() is None:
                    proc.terminate(); proc.wait(timeout=10)
        assert not (Path('/tmp') / f'{bundle}-gui-save.json').exists()
        events=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',log.read_text())
        assert len(events)>=7 and len(set(events))==1,'F9/F10 altered game state'
        positions=re.findall(r'music-position cue=inn enabled=true title_ms=\d+ inn_ms=(\d+)',log.read_text())
        assert positions and max(map(int,positions))>0,'Wine music player did not advance'
        checks.append(dict(bundle=bundle, package_sha256=sha(package),
                           ega_hd_rendered=True, in_game_switch=True, f9=True, music_default_on=True, normal_inn_music=True, save_unchanged=True))
    write(base / 'windows-smoke.json', dict(status='PASS', version=VERSION, checks=checks,
          windows_real_machine=False, runtime='Wine9 / Xvfb'))


def pixels():
    import numpy as np
    from PIL import Image
    import wave
    base=OUT/'smoke';fixture=base/'linux-gui/inputs/save.json'
    image=OUT/'full'/f'kol-{VERSION}-full-linux-x86_64.AppImage'
    comparisons=[]
    for graphics in ('ega','hd'):
        dest=base/('linux-reference-'+graphics);dest.mkdir()
        argv=[image,'--appimage-extract-and-run','-scene','play','-lang','zh-Hant','-graphics',graphics,
              '-load',fixture,'-seed','1','-play-script','L R RET SPACE','-shot-steps','0,1,2,3,4','-shot-dir',dest]
        write(dest/'command.json',list(map(str,argv)));run(argv,Path('/tmp'),dest/'run.log')
        for bundle in ('full','release'):
            for p in sorted(dest.glob('step-*.png')):
                a=np.asarray(Image.open(p).convert('RGB')).astype(np.int16)
                b=np.asarray(Image.open(base/f'wine-{bundle}-{graphics}'/p.name).convert('RGB')).astype(np.int16)
                assert a.shape==b.shape
                diff=np.abs(a-b);maximum=int(diff.max());count=int(np.count_nonzero(diff.max(2)))
                comparisons.append(dict(bundle=bundle,graphics=graphics,frame=p.name,max_channel_delta=maximum,different_pixels=count,
                                        predeclared_numeric_bound_met=maximum <= (0 if graphics=='ega' else 1)))
    with wave.open(str(base/'windows-audio.wav'),'rb') as wav:
        assert wav.getnchannels()==2 and wav.getframerate()==44100 and wav.getsampwidth()==2
        samples=np.frombuffer(wav.readframes(wav.getnframes()),dtype='<i2').astype(np.int32)
    peak=int(np.max(np.abs(samples)));tail=int(np.max(np.abs(samples[-88200:])))
    assert 100<peak<32767 and tail==0,(peak,tail)
    status='PASS' if all(r['predeclared_numeric_bound_met'] for r in comparisons) else 'NUMERIC_BOUND_UNMET'
    write(base/'windows-pixels.json',dict(status=status,comparisons=comparisons,reference_appimage_sha256=sha(image),
          audio=dict(device='isolated PulseAudio null sink; actual Windows PE via Wine9',sha256=sha(base/'windows-audio.wav'),peak=peak,last_second_peak=tail),windows_real_machine=False))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=('structures', 'linux', 'windows', 'pixels'))
    args = parser.parse_args()
    assert Path('/.dockerenv').exists()
    VERSION = json.loads((OUT / 'build-inputs.json').read_text())['version']
    globals()[args.phase]()
    if args.phase=='pixels':
        print('pixels',json.loads((OUT/'smoke/windows-pixels.json').read_text())['status'],flush=True)
    else:
        print(args.phase, 'PASS', flush=True)
