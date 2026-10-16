#!/usr/bin/env python3
from pathlib import Path
from PIL import Image, ImageChops
import hashlib
import json
import os
import re
import subprocess
import threading
import time

ROOT, OUT = Path('/src'), Path('/out')
DEST = OUT / 'formal-gui-events'
DEST.mkdir(exist_ok=True)
env = dict(os.environ, DISPLAY=':99', LIBGL_ALWAYS_SOFTWARE='1', KOL_SCENE_EVENTS_GUI='1')
condition = threading.Condition()
latest, serial, events = None, 0, []
process = None

def tool(*args, **kw):
    return subprocess.run(list(args), env=env, check=True, timeout=kw.pop('timeout', 15), **kw)

def read_frames():
    global latest, serial
    with (DEST / 'session.log').open('w') as log:
        for line in process.stdout:
            log.write(line)
            if line.startswith('display action='):
                events.append(line.strip())
            if line.startswith('equipment-frame '):
                state = json.loads(line[len('equipment-frame '):])
                with condition:
                    latest, serial = state, serial + 1
                    condition.notify_all()

def wait_frame(predicate=lambda x: True, after=-1, timeout=15):
    deadline = time.monotonic() + timeout
    with condition:
        while time.monotonic() < deadline:
            if latest is not None and serial > after and predicate(latest):
                return json.loads(json.dumps(latest))
            if process.poll() is not None:
                raise RuntimeError('game exited: ' + str(process.returncode))
            condition.wait(.1)
    raise RuntimeError('GUI frame timeout; last frame=' + str(latest)[:900])

def key(value, predicate=lambda x: True):
    old = serial
    tool('xdotool', 'key', '--window', window, '--clearmodifiers', value)
    return wait_frame(predicate, old + 1)

def field(mapping, name, default=None):
    snake = re.sub(r'(?<!^)(?=[A-Z])', '_', name).lower()
    return mapping.get(name, mapping.get(snake, mapping.get(name.lower(), default)))

def inn(state):
    return state.get('inside') or {}

def settings_open(state):
    return bool(field(state.get('settings') or {}, 'Open', False))

def model(state):
    return {k: v for k, v in state.items() if k not in ['hd', 'settings']}

def geometry():
    result = tool('xdotool', 'getwindowgeometry', '--shell', window, capture_output=True, text=True)
    data = dict(line.split('=', 1) for line in result.stdout.splitlines() if '=' in line)
    return int(data['WIDTH']), int(data['HEIGHT'])

def capture(name):
    # Draw queues GPU work before the X11 window presents it. Wait after observing
    # the closed state, rather than counting frames from before the input event.
    target_hd = latest['hd']
    observed = serial
    wait_frame(lambda s: s['hd'] == target_hd and not settings_open(s), after=observed + 2)
    ppm = DEST / (name + '.ppm')
    png = DEST / (name + '.png')
    tool('import', '-window', window, str(ppm), timeout=25)
    Image.open(ppm).convert('RGB').save(png, compress_level=1)
    ppm.unlink()
    return png

def click(x, y):
    width, height = geometry()
    old = serial
    tool('xdotool', 'mousemove', '--window', window, str(round(x * width / 960)), str(round(y * height / 600)), 'click', '1')
    return wait_frame(after=old + 1)

def cycle(label, mouse):
    before = wait_frame(lambda s: s['hd'] and not settings_open(s))
    original = capture(label + '-before')
    key('F10', settings_open)
    if mouse:
        click(758, 285)
    else:
        key('Up')
        key('Return', lambda s: not s['hd'])
    key('Escape', lambda s: not settings_open(s) and not s['hd'])
    ega = capture(label + '-ega')
    assert model(latest) == model(before), label + ' EGA changed complete controller/UI state'
    key('F10', settings_open)
    if mouse:
        click(758, 365)
        click(816, 494)
        wait_frame(lambda s: s['hd'] and not settings_open(s))
    else:
        key('Down')
        key('Return', lambda s: s['hd'])
        key('Escape', lambda s: not settings_open(s))
    after = wait_frame(lambda s: s['hd'] and not settings_open(s))
    assert model(after) == model(before), label + ' HD changed complete controller/UI state'
    restored = capture(label + '-restored')
    a, b = Image.open(original).convert('RGB'), Image.open(restored).convert('RGB')
    assert a.size == b.size
    delta = ImageChops.difference(a, b)
    maximum = max(x[1] for x in delta.getextrema())
    assert maximum <= 1, (label, maximum)
    (DEST / (label + '-state.json')).write_text(json.dumps(model(before), ensure_ascii=False, indent=2) + '\n')
    return dict(scene=label, mouse=mouse, size=a.size, hd_max_channel_delta=maximum,
                state_sha256=hashlib.sha256(json.dumps(model(before), sort_keys=True).encode()).hexdigest(),
                files=[p.name for p in [original, ega, restored]])


Path('/tmp/.X11-unix').mkdir(exist_ok=True)
os.chmod('/tmp/.X11-unix',0o1777)
xlog = (DEST / 'xvfb.log').open('w')
xserver = subprocess.Popen(['Xvfb', ':99', '-screen', '0', '2048x1536x24', '-nolisten', 'tcp'], stdout=xlog, stderr=subprocess.STDOUT)
try:
    for _ in range(50):
        if subprocess.run(['xdotool','getdisplaygeometry'],env=env,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode == 0:
            break
        time.sleep(.1)
    else:
        raise RuntimeError('Xvfb not ready')
    process = subprocess.Popen([str(OUT/'bin/scene-events-gui'), '-test.run', '^TestHDSceneEventGUI$', '-test.v'], cwd=ROOT, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True, bufsize=1)
    reader = threading.Thread(target=read_frames, daemon=True);reader.start()
    before=wait_frame(lambda s:s['hd'] and not settings_open(s),timeout=240)
    result=tool('xdotool','search','--onlyvisible','--pid',str(process.pid),capture_output=True,text=True)
    window=result.stdout.splitlines()[-1];tool('xdotool','windowfocus',window)
    output=[]
    for i,name in enumerate(['world-05','combat-03','arena-05']):
        wait_frame(lambda state:state['case']==name and state['hd'] and not settings_open(state))
        old=serial;tool('xdotool','windowsize',window,'1920','1200');wait_frame(after=old+2)
        output.append(cycle(name+'-large',False))
        old=serial;tool('xdotool','windowsize',window,'960','600');wait_frame(after=old+2)
        output.append(cycle(name+'-small',True))
        if i<2:key('F8',lambda state:state['case']!=name)
    held=serial;tool('xdotool','keydown','Control_L');wait_frame(after=held+1)
    try:
        tool('xdotool','keydown','q');process.wait(timeout=15)
    finally:
        tool('xdotool','keyup','q','Control_L')
    reader.join(timeout=3)
    assert process.returncode==0 and len(events)==36
    result=dict(status='PASS controlled actual F10 GUI',case='world-result-arena',rounds=output,display_events=events,serializable_scene_state_unchanged=True,exit_code=process.returncode,original_oracle=False,normal_player_path=False)
    (DEST/'results.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('GUI',result['case'],len(output),'rounds',len(events),'events; serializable scene preserved; HD roundtrip max',max(q['hd_max_channel_delta']for q in output),flush=True)
finally:
    if process is not None and process.poll() is None:
        process.terminate()
        try:process.wait(timeout=5)
        except subprocess.TimeoutExpired:process.kill();process.wait()
    xserver.terminate()
    try:xserver.wait(timeout=5)
    except subprocess.TimeoutExpired:xserver.kill();xserver.wait()
    xlog.close()
