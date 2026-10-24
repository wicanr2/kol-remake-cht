#!/usr/bin/env python3
"""SPEC-041：Docker／Xvfb 中的實際鍵鼠、預設保存與無 -load 重啟。"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--root', type=Path, default=Path('/repo'))
parser.add_argument('--bin', type=Path, required=True)
parser.add_argument('--out', type=Path, required=True)
args = parser.parse_args()
assert Path('/.dockerenv').exists()
ROOT, OUT, BIN = args.root.resolve(), args.out.resolve(), args.bin.resolve()
assert OUT.is_dir() and (OUT.stat().st_uid, OUT.stat().st_gid) == (os.getuid(), os.getgid())
GUI = OUT / 'gui'
GUI.mkdir()  # 必須是新的輸出，避免舊圖被當成本輪結果。
(GUI / 'asound.conf').write_text('pcm.!default { type null }\n')
env = dict(os.environ, DISPLAY=':92', LIBGL_ALWAYS_SOFTWARE='1', HOME='/tmp',
           XDG_CONFIG_HOME=str(GUI / 'config'), ALSA_CONFIG_PATH=str(GUI / 'asound.conf'))
proc = None
log = None
stream = None
window = None
actions = []


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def xdo(*argv):
    return subprocess.check_output(['xdotool', *map(str, argv)], env=env, stderr=subprocess.DEVNULL).decode().strip()


def waitfor(fn, seconds=15):
    end = time.monotonic() + seconds
    while time.monotonic() < end:
        if fn():
            return
        if proc is not None and proc.poll() is not None:
            raise RuntimeError(f'遊戲提早結束：{log.read_text()}')
        time.sleep(.05)
    raise TimeoutError(str(log))


def start(name, *extra, language='en'):
    global proc, log, stream, window
    log = GUI / (name + '.log')
    stream = log.open('w')
    command = [str(BIN), '-lang', language, '-scale', '3' if language == 'en' else '1',
               '-seed', '1', '-music', 'off', '-log-input', *map(str, extra)]
    proc = subprocess.Popen(command, cwd=ROOT, env=env, stdout=stream, stderr=stream)
    def find_window():
        global window
        try:
            window = xdo('search', '--onlyvisible', '--name', '武士傳說').splitlines()[0]
            return True
        except (subprocess.CalledProcessError, IndexError):
            return False
    waitfor(find_window)
    xdo('windowmove', window, 0, 0)
    xdo('windowfocus', window)
    time.sleep(.6)
    actions.append(dict(start=name, command=command))


def count_inputs():
    return log.read_text().count('input scene=')


def key(*keys):
    for k in keys:
        n = count_inputs()
        xdo('keydown', '--window', window, k)
        time.sleep(.12)
        xdo('keyup', '--window', window, k)
        waitfor(lambda: count_inputs() > n)
        time.sleep(.08)
        actions.append(dict(key=k, log=log.name))


def click(x, y):
    n = count_inputs()
    xdo('mousemove', '--window', window, x, y)
    xdo('click', '--window', window, '1')
    waitfor(lambda: count_inputs() > n)
    time.sleep(.15)
    actions.append(dict(click=[x, y], log=log.name))


def shot(name):
    subprocess.run(['ffmpeg', '-y', '-nostdin', '-v', 'error', '-f', 'x11grab', '-draw_mouse', '0',
                    '-video_size', '960x600', '-i', ':92', '-frames:v', '1', '-update', '1',
                    str(GUI / (name + '.png'))], env=env, check=True, timeout=15)


def stop():
    global proc, stream
    xdo('keydown', '--window', window, 'ctrl+q')
    proc.wait(timeout=10)
    assert proc.returncode == 0, log.read_text()
    stream.close()
    proc = None
    # 控制鍵的釋放送到 X server，不向已消失的視窗送事件。
    xdo('keyup', 'ctrl+q')


def create(*letters):
    key('Return', *letters)
    shot('typing-' + str(len(actions)))
    key('Return', 'Return', 'Return', 'Return', 'Down', 'Return', 'Return')


xlog = (GUI / 'xvfb.log').open('w')
xvfb = subprocess.Popen(['Xvfb', ':92', '-screen', '0', '1280x800x24', '-nolisten', 'tcp'],
                        stdout=xlog, stderr=xlog)
try:
    time.sleep(.3)
    start('creation')
    key('Return', 'Return', '5', 'Down')
    create('a', 'shift+b', 'c', 'shift+d')
    create('shift+z', 'y', 'shift+x')
    save = GUI / 'config/kol-remake-cht/save.dat'
    data = json.loads(save.read_text())
    assert data['characters'][14]['name'] == 'aBcD'
    assert data['characters'][15]['name'] == 'ZyX'
    assert not data['party']
    shot('created-menu')
    stop()
    before = sha(save)

    # 真正的新程序，不帶 -load 及 -save。
    start('automatic-reload')
    shot('automatic-reload-menu')
    key('Escape')
    assert 'input scene=menu ' in log.read_text(), log.read_text()
    key('6', *(['Down'] * 14), 'Return', 'space')
    shot('created-character-inn')
    key('less', 'Return', 'Return')
    c = data['characters'][14]
    names = [v['name'] for v in data['characters'] if v is not None and v['name'] != c['name']
             and v['place'] == c['place'] and v['town_code'] == c['town_code']]
    key(*(['Down'] * names.index('ZyX')), 'Return')
    shot('recruited-second-character')
    key('less', 'less', 'Return', 'Return', 'Return', 'Return')
    data = json.loads(save.read_text())
    assert data['party'] == [14, 15]
    assert [data['characters'][i]['gold'] for i in (14, 15)] == [1440, 1440]
    assert before != sha(save)
    shot('paid-and-saved')
    stop()
    before = sha(save)
    start('automatic-party-reload')
    shot('automatic-party-reload')
    key('1', '2')
    shot('sheet-second-character')
    key('period', 'Return', '1')
    shot('equipment-first-character')
    key('period', 'Return', 'Return')
    n = count_inputs()
    xdo('keydown', '--window', window, 'Right')
    time.sleep(.8)
    xdo('keyup', '--window', window, 'Right')
    assert count_inputs() - n >= 3, '長按未產生連續移動'
    shot('held-map-movement')
    stop()
    assert sha(save) == before, '未住宿卻改了存檔'

    # 受控店鋪輸入只用於場景驗收，不冒稱自然遊玩取得物品。
    buildings = json.loads((ROOT / 'assets/original-derived-data/town-buildings.json').read_text())
    town = next(t for t in buildings['towns'] if t['town'] == 1)
    door = next(d for d in town['doors'] if (d['x'], d['y']) == (23, 52))
    fixture = copy.deepcopy(data)
    idx = next(i for i, c in enumerate(fixture['characters']) if c and c['name'] == 'ddduhh')
    fixture['party'], fixture['speaker'] = [idx], 0
    fixture['location'] = dict(kind='town', town=1, entry_point=0, x=24, y=52)
    c = fixture['characters'][idx]
    c['gold'] = 3000
    def complete_item(good):
        item = dict(index=0, name='', price=0, kind=0, damage=False, hands=0, **{'class': 0},
                    position=0, encumbrance=0, condition=0, percent=0, size_class=0,
                    record='', picture=0, movement=0)
        item.update(copy.deepcopy(good))
        item['dice'] = dict(count=0, sides=0, bonus=0, low=0, high=0) | good.get('dice', {})
        return item
    c['backpack'] = [complete_item(door['goods'][0]), complete_item(door['goods'][1])]
    for i in range(6):
        c['carried'][1+i] = c['backpack'][i]['record'] if i < 2 else '00' * 16
    cap = copy.deepcopy(c['backpack'][0])
    cap['encumbrance'] = 200
    raw = bytearray.fromhex(cap['record'])
    raw[10] = 200
    cap['record'] = raw.hex()
    c['worn'][0], c['carried'][7] = cap, cap['record']
    fixture_path = GUI / 'shop-fixture.json'
    fixture_path.write_text(json.dumps(fixture))
    for language in ('en', 'zh-Hant'):
        start('shop-' + language, '-load', fixture_path, language=language)
        key('Left', 'Return', 'space')
        # 圖示位置兩種語言共用實體 960x600 座標。
        click(612, 540)  # SELL，圖示列第六格。
        shot('sell-list-' + language)
        click(330, 282 if language == 'en' else 280)  # 第二列。
        shot('sell-offer-' + language)
        key('Return')
        shot('sell-done-' + language)
        key('space')
        click(708, 540)  # ANVIL。
        shot('anvil-' + language)
        key('space')
        shot('anvil-offer-' + language)
        key('Return')
        click(228, 540)  # MIRROR。
        key('period', 'Return', 'Return')
        shot('armour-adjusted-' + language)
        stop()

    hostel = next(d for t in buildings['towns'] if t['town'] == 7
                  for d in t['doors'] if (d['x'], d['y']) == (59, 28))
    magic_fixture = copy.deepcopy(fixture)
    magic_fixture['location'] = dict(kind='town', town=7, entry_point=7, x=59, y=29)
    c = magic_fixture['characters'][idx]
    c['order'] = {k: copy.deepcopy(hostel['order'][k]) for k in
                  ('name', 'title', 'intellect', 'allow', 'record', 'token', 'token_record')}
    sp = hostel['spells'][0]
    c['spells'] = [{k: sp[k] for k in ('name', 'effect', 'turn_code', 'power_code', 'guards')}]
    c['spellcaster'] = True
    c['sheet']['adventure_points'] = 3000
    c['sheet']['stats'][3] = 150
    magic_path = GUI / 'magic-fixture.json'
    magic_path.write_text(json.dumps(magic_fixture))
    for language in ('en', 'zh-Hant'):
        start('magic-' + language, '-load', magic_path, language=language)
        key('Up', 'Return', 'space')
        click(516, 540)  # SCROLL。
        click(330, 282 if language == 'en' else 280)  # Modify a spell。
        key('Return')  # 第一筆既有法術。
        shot('magic-editor-' + language)
        # 第二欄是第一個種族欄；從原本矮人改成同教派允許的凱爾登。
        click(180, 195)
        shot('magic-options-' + language)
        click(330 if language == 'en' else 540, 357 if language == 'en' else 273)
        shot('magic-selected-' + language)
        stop()

    # 壞預設保存必須拒絕啟動，且原檔不得被新名冊覆寫。
    original = save.read_bytes()
    save.write_text('broken save')
    result = subprocess.run([str(BIN), '-music', 'off'], cwd=ROOT, env=env,
                            capture_output=True, timeout=20)
    assert result.returncode != 0 and save.read_text() == 'broken save'
    save.write_bytes(original)
    (GUI / 'bad-save.log').write_bytes(result.stdout + result.stderr)
    files = [p for p in GUI.rglob('*') if p.is_file()]
    assert all((p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()) for p in files)
    receipt = dict(status='PASS', spec='docs/spec/041-player-feedback-fixes.md', seed=1,
                   binary_sha256=sha(BIN), verifier_sha256=sha(Path(__file__)),
                   source_zip_sha256=sha(ROOT / 'Knights-of-legend.zip'),
                   scope=['名稱大小寫', '建角與住宿預設存檔', '兩次無-load新程序續玩',
                          '人物及裝備數字切換', '方向長按', '雙語販賣清單與鐵砧', '雙語法術編輯清單', '壞檔拒絕啟動'],
                   original_parity_claimed=False, real_platform='Linux Docker/Xvfb',
                   controlled_inputs='shop-fixture.json 與 magic-fixture.json；位置、隊伍、金冠、背包、過大護甲、教派及法術',
                   actions=actions, outputs_sha256={str(p.relative_to(GUI)): sha(p) for p in files})
    (GUI / 'receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
    print('PASS: SPEC-041 實際互動、預設保存及重啟', flush=True)
finally:
    if proc is not None and proc.poll() is None:
        proc.terminate()
        proc.wait(timeout=10)
    if stream is not None:
        stream.close()
    xvfb.terminate()
    xvfb.wait(timeout=10)
    xlog.close()
