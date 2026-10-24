#!/usr/bin/env python3
"""Docker限定：六個完整版中的 Linux／Windows 正常操作與固定流程抽樣。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time
import zipfile


def main():
    assert Path('/.dockerenv').exists()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--platform', choices=('linux', 'windows'), required=True)
    parser.add_argument('--bundle', choices=('full', 'release'), default='full')
    parser.add_argument('--out', type=Path, default=Path('/out'))
    parser.add_argument('--work', type=Path, default=Path('/work'))
    parser.add_argument('--resume-save', type=Path, help='實際建角及住宿產生的保存，用來核對無-load續玩')
    args = parser.parse_args()
    out, work = args.out, args.work
    version = json.loads((out / 'build-inputs.json').read_text())['version']
    windows = args.platform == 'windows'
    env = dict(os.environ, DISPLAY=':99', PULSE_SERVER='unix:/work/smoke/wine-pulse/native', PULSE_SINK='kol_wine')
    env.update(HOME='/tmp/kol-home',XDG_CONFIG_HOME='/tmp/kol-config')
    Path(env['HOME']).mkdir();Path(env['XDG_CONFIG_HOME']).mkdir()
    if windows:
        env.update(WINEPREFIX='/tmp/kol-repack-wine', WINEDEBUG='-all', WINEDLLOVERRIDES='mscoree,mshtml=')
        Path(env['WINEPREFIX']).mkdir()
    else:
        # Linux操作驗收使用虛擬ALSA；實際音訊另由Wine／PulseAudio錄製核對。
        alsa=work/'linux-asound.conf'
        alsa.write_text('pcm.!default { type null }\n')
        env['ALSA_CONFIG_PATH']=str(alsa)
    x = subprocess.Popen(['Xvfb', ':99', '-screen', '0', '1920x1200x24', '-nolisten', 'tcp'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    proc = None
    runs = []

    def xdo(*values):
        return subprocess.check_output(['xdotool', *map(str, values)], env=env, text=True, timeout=10).strip()

    def wait(fn, log, seconds=60):
        until = time.monotonic() + seconds
        while time.monotonic() < until:
            assert proc.poll() is None, log.read_text(errors='replace')
            if fn():
                return
            time.sleep(.1)
        raise AssertionError('timeout: ' + log.read_text(errors='replace'))

    def key(window, name):
        xdo('keydown', '--window', window, name)
        time.sleep(.8)
        xdo('keyup', '--window', window, name)
        time.sleep(.8)

    def inputkey(window, name, scene, log):
        marker = 'input scene=' + scene + ' '
        count = log.read_text(errors='replace').count(marker)
        xdo('keydown', '--window', window, name)
        try:
            wait(lambda: log.read_text(errors='replace').count(marker) > count, log, 45)
        finally:
            xdo('keyup', '--window', window, name)
        time.sleep(1)

    def displaykey(window, name, action, log):
        count = log.read_text(errors='replace').count(action)
        frames = log.read_text(errors='replace').count('display-frame ')
        xdo('keydown', '--window', window, name)
        try:
            wait(lambda: log.read_text(errors='replace').count(action) > count, log, 90)
        finally:
            xdo('keyup', '--window', window, name)
        wait(lambda: log.read_text(errors='replace').count('display-frame ') > frames, log, 90)
        time.sleep(1)

    def patharg(path):
        return 'Z:' + str(path).replace('/', '\\') if windows else str(path)

    try:
        time.sleep(.5)
        assert (work / 'smoke/wine-pulse/native').is_socket(), '先啟動隔離音訊容器'
        if windows:
            subprocess.run(['/usr/lib/wine/wine64', 'reg', 'add', r'HKCU\Software\Wine\Drivers', '/v', 'Audio', '/t', 'REG_SZ', '/d', 'pulse', '/f'], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True, timeout=90)
        for theme in ('ega', 'ega-hd'):
            suffix = 'windows-amd64.zip' if windows else 'linux-x86_64.AppImage'
            entries=json.loads((out/'SHA256SUMS.json').read_text())['files']
            package=out/next(row['path'] for row in entries if row['bundle']==args.bundle and row['graphics']==theme and row['platform']==('windows-amd64' if windows else 'linux-x86_64'))
            dest = Path('/tmp') / ('game-' + theme)
            dest.mkdir()
            if windows:
                with zipfile.ZipFile(package) as z:
                    z.extractall(dest)
                root = next(dest.iterdir())
                command = ['/usr/lib/wine/wine64', str(root / 'kol.exe')]
            else:
                subprocess.run([str(package), '--appimage-extract'], cwd=dest, check=True, stdout=subprocess.DEVNULL, timeout=120)
                root = dest / 'squashfs-root'
                command = [str(root / 'AppRun')]
            if args.bundle == 'release':
                target = root if windows else Path('/tmp') / ('data-' + theme)
                subprocess.run(['python3', str(root / 'tools/import_game_data.py'), '--source', '/src',
                                '--target', str(target)], check=True, timeout=180)
                if not windows:
                    env['KOL_DATA_DIR'] = str(target)
                assert (target / 'assets/hd').exists() == (theme == 'ega-hd')
            else:
                target = root
            assert (target / 'assets/hd').exists() == (theme == 'ega-hd')
            actual_version = subprocess.check_output([*command, '-version'], cwd=root, env=env, text=True, timeout=30).strip()
            assert actual_version == version, actual_version
            name = args.platform + '-' + theme
            log = work / (name + '-gui.log')
            save = Path('/tmp') / (name + '-save.json')
            with log.open('w') as stream:
                proc = subprocess.Popen([*command, '-seed', '1', '-log-input', '-save', patharg(save)], cwd=root, env=env, stdout=stream, stderr=subprocess.STDOUT)
                text = lambda: log.read_text(errors='replace')
                win = []

                def find():
                    found = subprocess.run(['xdotool', 'search', '--onlyvisible', '--name', '^武士傳說$'], env=env, capture_output=True, text=True, timeout=10)
                    if found.returncode == 0:
                        win[:] = found.stdout.splitlines()[-1:]
                    return bool(win)

                wait(find, log)
                window = win[0]
                xdo('windowfocus', '--sync', window)
                wait(lambda: 'music cue=title' in text(), log)
                inputkey(window, 'space', 'title', log)
                inputkey(window, 'Return', 'cover', log)
                inputkey(window, '5', 'contents', log)
                displaykey(window,'F8','display action=toolbar-2',log)
                assert 'no_fatigue=false enemy_multiplier=1 fear_immune=false' in text()
                displaykey(window,'Return','display action=toolbar-8',log)
                key(window,'Right');displaykey(window,'Return','display action=toolbar-9',log)
                key(window,'Right');displaykey(window,'Return','display action=toolbar-10',log)
                assert 'no_fatigue=true enemy_multiplier=6 fear_immune=true' in text()
                key(window,'Right');key(window,'Right')
                displaykey(window,'Return','display action=toolbar-12',log)
                assert re.findall(r'toolbar action=12[^\n]*',text())[-1].endswith('no_fatigue=false enemy_multiplier=1 fear_immune=false')
                displaykey(window,'Escape','display action=toolbar-back',log)
                key(window, 'F10')
                wait(lambda: 'display-frame hd=false open=true' in text(), log)
                key(window, 'Down')
                key(window, 'Return')
                if theme == 'ega':
                    wait(lambda: 'display action=hd-unavailable hd=false' in text(), log)
                    assert 'display-frame hd=true' not in text()
                else:
                    wait(lambda: 'display-frame hd=true open=true cursor=1' in text(), log, 90)
                    key(window, 'Up')
                    key(window, 'Return')
                    wait(lambda: 'display-frame hd=false open=true cursor=0' in text(), log)
                    count = text().count('display-frame hd=true open=true cursor=1')
                    key(window, 'Down')
                    key(window, 'Return')
                    wait(lambda: text().count('display-frame hd=true open=true cursor=1') > count, log, 90)
                states = re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)', text())
                assert states and len(set(states)) == 1, 'F10 改變正式狀態'
                key(window, 'Escape')
                inputkey(window, 'Down', 'menu', log)
                inputkey(window, 'Return', 'menu', log)
                inputkey(window, 'Escape', 'newchar', log)
                inputkey(window, 'Escape', 'menu', log)
                inputkey(window, '6', 'contents', log)
                inputkey(window, 'Return', 'select', log)
                displaykey(window, 'F9', 'display action=music-false', log)
                displaykey(window, 'F9', 'display action=music-true', log)
                wait(lambda: 'display action=music-true' in text(), log)
                wait(lambda: 'music cue=inn' in text(), log)
                xdo('keydown', '--window', window, 'ctrl')
                xdo('keydown', '--window', window, 'q')
                code = proc.wait(timeout=20)
                xdo('keyup', 'q', 'ctrl')
                assert code == 0 and not save.exists()
                assert 'panic:' not in text() and 'HD畫面載入失敗' not in text()
            if args.resume_save:
                fixture = json.loads(args.resume_save.read_text())
                assert fixture['party'] == [14, 15]
                assert [fixture['characters'][i]['gold'] for i in (14,15)] == [1440,1440]
                shutil.copy2(args.resume_save, save)
                before = save.read_bytes()
                resume_log = work/(name+'-resume.log')
                with resume_log.open('w') as stream:
                    proc = subprocess.Popen([*command,'-seed','1','-music','off','-log-input',
                        '-save',patharg(save)],cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
                    win.clear()
                    wait(find,resume_log)
                    window=win[0]
                    xdo('windowfocus','--sync',window)
                    inputkey(window,'1','town',resume_log)
                    inputkey(window,'2','town',resume_log)
                    assert 'input scene=title' not in resume_log.read_text(errors='replace')
                    xdo('keydown','--window',window,'ctrl')
                    xdo('keydown','--window',window,'q')
                    assert proc.wait(timeout=20)==0
                    xdo('keyup','q','ctrl')
                assert save.read_bytes()==before, '檢視角色改變了保存'
            shots = work / (name + '-flow')
            shots.mkdir()
            with (work / (name + '-flow.log')).open('w') as stream:
                subprocess.run([*command, '-scene', 'flow', '-lang', 'zh-Hant', '-graphics', 'hd' if theme == 'ega-hd' else 'ega',
                    '-seed', '1', '-flow-script', 'RET RET 5 D RET', '-shot-dir', patharg(shots),
                    '-trace', patharg(work / (name + '-trace.json')), '-text-log', patharg(work / (name + '-text.jsonl'))],
                    cwd=root, env=env, stdout=stream, stderr=subprocess.STDOUT, check=True, timeout=120)
            assert len(list(shots.glob('*.png'))) == 6
            runs.append(dict(package=package.name, bundle=args.bundle, sha256=hashlib.file_digest(package.open('rb'), 'sha256').hexdigest(),
                graphics=theme, version=actual_version, exit_code=code, normal_intro=True, party_inn=True,
                create_menu_enter_cancel=True, f9=True, f10_state_unchanged=True, hd_disabled=theme == 'ega',
                assistance_toolbar=True,assistance_startup_off=True,assistance_toggles_and_reset=True,
                hd_release_reload=theme == 'ega-hd', no_save_written=True, flow_frames=6))
            if args.resume_save:
                runs[-1]['automatic_resume_no_load']=True
                runs[-1]['resume_party']=['aBcD','ZyX']
                runs[-1]['resume_gold']=[1440,1440]
            print('RUNTIME PASS', package.name, flush=True)
        (work / (args.platform + '-runtime.json')).write_text(json.dumps(dict(status='PASS', version=version,
            runs=runs, memory_limit=Path('/sys/fs/cgroup/memory.max').read_text().strip(),
            memory_peak=int(Path('/sys/fs/cgroup/memory.peak').read_text()),
            memory_events=Path('/sys/fs/cgroup/memory.events').read_text(),
            real_windows_machine=False, host_fuse=False), ensure_ascii=False, indent=2) + '\n')
    finally:
        if proc is not None and proc.poll() is None:
            proc.terminate()
            proc.wait(timeout=10)
        if windows:
            subprocess.run(['wineserver', '-k'], env=env, check=False)
        x.terminate()
        x.wait(timeout=10)


if __name__ == '__main__':
    main()
