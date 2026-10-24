#!/usr/bin/env python3
"""SPEC-043：Docker/Xvfb正式鍵鼠切換字型、存檔與偏好重載。"""
import argparse
import ctypes
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time

import numpy as np
from PIL import Image

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--root', type=Path, required=True)
p.add_argument('--bin', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
assert Path('/.dockerenv').exists() and not a.out.exists()
a.out.mkdir()
assert a.out.stat().st_uid == os.getuid()
root, out, binary = a.root.resolve(), a.out.resolve(), a.bin.resolve()
(out / 'asound.conf').write_text('pcm.!default { type null }\n')
env = dict(os.environ, DISPLAY=':94', HOME='/tmp', XDG_CONFIG_HOME=str(out / 'config'),
           ALSA_CONFIG_PATH=str(out / 'asound.conf'), LIBGL_ALWAYS_SOFTWARE='1')
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
proc = stream = logfile = window = None
actions, checks = [], []
save = out / 'config/kol-remake-cht/save.dat'
prefs = out / 'config/kol-remake-cht/toolbar.json'


def xdo(*args):
    return subprocess.check_output(['xdotool', *map(str, args)], env=env, stderr=subprocess.DEVNULL).decode().strip()


def wait(fn, seconds=45):
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        if fn(): return
        if proc is not None and proc.poll() is not None: raise RuntimeError(logfile.read_text())
        time.sleep(.05)
    raise TimeoutError(str(logfile))


def start(name, language='zh-Hant'):
    global proc, stream, logfile, window
    logfile = out / (name + '.log'); stream = logfile.open('w')
    cmd = [str(binary), '-lang', language, '-scale', '3' if language == 'en' else '1',
           '-seed', '1', '-music', 'off', '-log-input']
    proc = subprocess.Popen(cmd, cwd=root, env=env, stdout=stream, stderr=stream)
    def visible():
        global window
        try: window = xdo('search', '--onlyvisible', '--name', '武士傳說').splitlines()[0]; return True
        except (subprocess.CalledProcessError, IndexError): return False
    wait(visible)
    xdo('windowmove', window, 0, 0); xdo('windowfocus', window); time.sleep(1)
    actions.append(dict(start=name, command=cmd))


def key(*keys, game=False, event=False):
    for k in keys:
        before = logfile.read_text().count('input scene=')
        frame = logfile.read_text().count('toolbar-frame ')
        xdo('keydown', '--window', window, k)
        try:
            if game: wait(lambda: logfile.read_text().count('input scene=') > before)
            elif event or k.startswith('F') or k == 'Escape':
                wait(lambda: logfile.read_text().count('toolbar-frame ') > frame)
            else: time.sleep(.18)
        finally: xdo('keyup', '--window', window, k)
        time.sleep(.2)
        actions.append(dict(key=k, game=game, log=logfile.name))


def geometry():
    raw = dict(line.split('=', 1) for line in xdo('getwindowgeometry', '--shell', window).splitlines())
    return int(raw['WIDTH']), int(raw['HEIGHT'])


def resize_window(width, height):
    xdo('windowsize', window, width, height)
    wait(lambda: geometry() == (width, height))
    # X11尺寸已更新不代表GL前緩衝已交換；正常F10往返取得重排後的繪製回條。
    key('F10'); key('Escape'); time.sleep(.4)


def click(x, y):
    width, _ = geometry()
    frame = logfile.read_text().count('toolbar-frame ')
    xdo('mousemove', '--window', window, round(x * width / 960), round(y * width / 960))
    xdo('mousedown', '--window', window, 1)
    try: wait(lambda: logfile.read_text().count('toolbar-frame ') > frame)
    finally: xdo('mouseup', '--window', window, 1)
    time.sleep(.3)
    actions.append(dict(click=[x, y], width=width, log=logfile.name))


def shot(name):
    width, height = geometry()
    # 游標固定到工具列上緣，避免字型前後比較包含不同的遊戲游標位置。
    xdo('mousemove', '--window', window, width - 1, 0)
    time.sleep(.2)
    path = out / (name + '.png')
    subprocess.run(['ffmpeg', '-y', '-nostdin', '-v', 'error', '-f', 'x11grab', '-draw_mouse', '0',
                    '-video_size', f'{width}x{height}', '-i', ':94', '-frames:v', '1', '-update', '1', str(path)],
                   env=env, check=True, timeout=15)
    actions.append(dict(screenshot=path.name, log=logfile.name))
    return np.asarray(Image.open(path).convert('RGB')).astype(np.int16)


def stop():
    global proc, stream
    xdo('keydown', '--window', window, 'ctrl+q'); proc.wait(timeout=15)
    xdo('keyup', 'ctrl+q')
    assert proc.returncode == 0, logfile.read_text()
    stream.close(); proc = None


def states():
    return re.findall(r'display action=(\S+) hd=(\w+) state=(\w+)', logfile.read_text())


def create(*letters):
    key('Return', *letters, game=True)
    key('Return', 'Return', 'Return', 'Return', 'Down', 'Return', 'Return', game=True)


def choose_font(rounded, keyboard=False):
    key('F10')
    if keyboard:
        key('Right', event=True)
        # tab switch sets cursor to the currently active font.
        if rounded != (json.loads(prefs.read_text())['game_font'] == 'rounded' if prefs.exists() else False):
            key('Down' if rounded else 'Up')
        key('Return', event=True)
    else:
        click(780, 260); click(500, 406 if rounded else 352)
    assert json.loads(prefs.read_text())['game_font'] == ('rounded' if rounded else 'current')
    shot(('rounded' if rounded else 'current') + '-settings-' + str(len(actions)))
    key('Escape'); time.sleep(.8)


xlog = (out / 'xvfb.log').open('w')
xvfb = subprocess.Popen(['Xvfb', ':94', '-screen', '0', '2560x1600x24', '-nolisten', 'tcp'], stdout=xlog, stderr=xlog)
try:
    time.sleep(.4)
    x11 = ctypes.CDLL('libX11.so.6'); x11.XOpenDisplay.restype = ctypes.c_void_p
    for fn in ('XAutoRepeatOff', 'XFlush', 'XCloseDisplay'): getattr(x11, fn).argtypes = [ctypes.c_void_p]
    display = x11.XOpenDisplay(b':94'); assert display
    x11.XAutoRepeatOff(display); x11.XFlush(display)
    start('normal-player-path')
    key('Return', 'Return', '5', 'Down', game=True)
    create('a', 'shift+b', 'c', 'shift+d'); create('shift+z', 'y', 'shift+x')
    data = json.loads(save.read_text())
    key('Escape', '6', *(['Down'] * 14), 'Return', 'space', game=True)
    key('less', 'Return', 'Return', game=True)
    c = data['characters'][14]
    names = [v['name'] for v in data['characters'] if v and v['name'] != c['name'] and v['place'] == c['place'] and v['town_code'] == c['town_code']]
    key(*(['Down'] * names.index('ZyX')), 'Return', game=True)
    initial_save = sha(save)
    current = shot('ega-current-inn')
    choose_font(True)
    rounded = shot('ega-rounded-inn')
    assert np.count_nonzero(current != rounded) > 100
    assert sha(save) == initial_save
    # Same paused game snapshot across both font choices, no gameplay input leakage.
    key('F10'); click(780, 260)
    state = states()[-1][2]; before = logfile.read_text().count('input scene=')
    key('1', 'space'); xdo('keydown', '--window', window, 'Down'); time.sleep(.7); xdo('keyup', '--window', window, 'Down')
    assert logfile.read_text().count('input scene=') == before
    click(500, 352); assert states()[-1][2] == state
    click(500, 406); assert states()[-1][2] == state
    shot('font-settings-ega')
    key('Escape'); time.sleep(.8)
    choose_font(False)
    restored = shot('ega-restored-inn')
    assert np.array_equal(current, restored), '回到現有字型應逐像素恢復'
    checks.append('EGA粉圓文字變化、原字模像素還原、設定期間狀態／保存與遊戲輸入隔離')
    choose_font(True)
    # Graphics switches from the new shared desktop panel, then native 1920 screenshot.
    key('F10'); click(500, 260); click(500, 406); time.sleep(1)
    assert states()[-1][1] == 'true'
    key('Escape'); time.sleep(.8)
    resize_window(1920, 1296)
    hd_rounded = shot('hd-rounded-inn')
    choose_font(False); hd_current = shot('hd-current-inn')
    assert np.count_nonzero(hd_current != hd_rounded) > 100
    choose_font(True); hd_again = shot('hd-rounded-restored-inn')
    assert np.max(np.abs(hd_again - hd_rounded)) <= 1
    key('F10'); click(780, 260); shot('font-settings-hd'); key('Escape'); time.sleep(.8)
    checks.append('HD旅店比例字型切換，現有／粉圓往返，1920原生尺寸設定頁')
    # Hidden toolbar and keyboard-only font choice; desktop coordinates remain independent.
    click(40, 24); time.sleep(.5)
    resize_window(960, 648)
    key('F12'); key('F10'); click(780, 260); shot('font-settings-hidden'); key('Escape'); time.sleep(.8); key('F12')
    gstate = states()[-1][2]
    key('F10'); key('Left', event=True); key('Escape'); time.sleep(.8)
    choose_font(False, keyboard=True)
    choose_font(True)
    assert sha(save) == initial_save
    # Inn save follows the existing normal user flow, after all modal inputs released.
    key('less', 'less', 'Return', 'Return', 'Return', 'Return', game=True)
    saved = json.loads(save.read_text())
    assert [saved['characters'][i]['gold'] for i in (14, 15)] == [1440, 1440]
    save_sha = sha(save)
    stop()
    start('cold-reload')
    assert json.loads(prefs.read_text())['game_font'] == 'rounded'
    shot('reloaded-rounded-party')
    key('F10'); click(780, 260); shot('font-settings-reloaded')
    assert sha(save) == save_sha
    checks.append('收合工具列、鍵盤／滑鼠、正常住宿保存扣款60及新程序無-load續玩與字型偏好重載')
    key('Escape'); time.sleep(.8)
    # 數字鍵於城鎮戶外開人物頁；建築內只切換發言者，不能誤當成已開頁。
    key('1', game=True)
    sheet_rounded = shot('ega-rounded-sheet')
    assert np.count_nonzero(np.all(sheet_rounded[462:484, 116:866] == [0, 170, 170], axis=2)) > 10000, '未抵達人物頁的傷勢／營養條'
    choose_font(False); sheet_current = shot('ega-current-sheet')
    assert np.count_nonzero(sheet_current != sheet_rounded) > 100
    assert np.array_equal(sheet_current[218:411, 282:355], sheet_rounded[218:411, 282:355]), '屬性數字變化'
    assert np.array_equal(sheet_current[60:90, 210:750], sheet_rounded[60:90, 210:750]), '英文字母姓名變化'
    choose_font(True); sheet_again = shot('ega-rounded-restored-sheet')
    assert np.array_equal(sheet_again, sheet_rounded)
    key('2', game=True); shot('ega-rounded-second-sheet'); key('1', game=True)
    click(150, 24); time.sleep(.8)
    assert states()[-1][1] == 'true'
    resize_window(1920, 1296)
    hd_sheet = shot('hd-rounded-sheet')
    choose_font(False); hd_sheet_current = shot('hd-current-sheet')
    assert np.count_nonzero(hd_sheet != hd_sheet_current) > 100
    choose_font(True); hd_sheet_again = shot('hd-rounded-restored-sheet')
    assert np.max(np.abs(hd_sheet_again - hd_sheet)) <= 1
    assert sha(save) == save_sha
    checks.append('正常戶外人物頁、1／2切換人物；EGA英文字母姓名及屬性數字像素不變，EGA／HD原生頁字型往返還原')
    stop()
    start('english', 'en')
    english = shot('english-current')
    choose_font(False)
    english_again = shot('english-font-change')
    assert np.array_equal(english, english_again), '英文遊戲字模應保持'
    choose_font(True)
    assert sha(save) == save_sha
    stop()
    checks.append('英文模式字模不變、字型偏好仍可設定、不寫遊戲保存')
    program = {str(f.relative_to(root)): sha(f) for d in ('cmd', 'internal') for f in sorted((root / d).rglob('*.go'))}
    program.update({n: sha(root / n) for n in ('go.mod', 'go.sum')})
    for folder in ('toolbarfont', 'roundedfont'):
        program.update({str(f.relative_to(root)): sha(f) for f in sorted((root / 'cmd/kol' / folder).glob('*')) if f.is_file()})
    receipt = dict(status='PASS', spec='docs/spec/043-rounded-game-font.md', binary_sha256=sha(binary),
                   verifier_sha256=sha(Path(__file__)), program_inputs=program, seed=1,
                   platform='Linux Docker/Xvfb', checks=checks, actions=actions,
                   screenshots={f.name: sha(f) for f in out.glob('*.png')}, save_sha256=sha(save),
                   source_zip_sha256=sha(root / 'Knights-of-legend.zip'),
                   original_parity_claimed=False, capture='正式互動程式X11畫面；正常建角、招募及住宿，未使用場景注入',
                   rights='原作衍生遊戲畫面PNG，依2026-10-17公開化決定保存；原作權利未改變')
    (out / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(dict(status='PASS', checks=checks), ensure_ascii=False))
finally:
    if proc is not None and proc.poll() is None: proc.terminate(); proc.wait(timeout=10)
    if stream is not None and not stream.closed: stream.close()
    xvfb.terminate(); xvfb.wait(timeout=10); xlog.close()
