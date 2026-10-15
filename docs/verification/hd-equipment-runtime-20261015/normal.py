#!/usr/bin/env python3
"""Controlled town controller input, no production scene injection."""
from pathlib import Path
import hashlib
import json
import os
import subprocess
import sys
import time

ROOT = Path('/src')
OUT = Path('/out')
initial = 'L R RET SPACE > RET > RET'
human = 'SPACE SPACE 6 D D D D D RET SPACE > RET'
cases = {
    'winged-regression': ('play', initial + ' D'),
    'human-slots': ('flow', human + ' > RET ' + ' '.join(['D'] * 16)),
    'human-detail': ('flow', human + ' > RET D RET SPACE'),
    'human-party': ('flow', human + ' > > RET SPACE'),
}
for name, cursor in [('human-two-hand', 7), ('human-belt', 6)]:
    prefix = 'SPACE SPACE 6 ' + ' '.join(['D'] * cursor) + ' RET SPACE > RET'
    cases[name] = ('flow', prefix + ' > RET D')
for build, name, cursor in [(1,'Olear',2),(2,'Fighting Bull',4),(3,'Theodoric',0),(4,'Moraine',3),(5,'Shyla',9)]:
    prefix='SPACE SPACE 6 '+ ' '.join(['D']*cursor) +' RET SPACE > RET'
    cases[f'body{build}-slots']=('flow',prefix+' > RET '+ ' '.join(['D']*16))
    cases[f'body{build}-party']=('flow',prefix+' > > RET SPACE')
recruit='SPACE SPACE 6 RET SPACE'
for count in [1,1,1,1,4]:
    recruit+=' < RET RET '+' '.join(['D']*count)+' RET'
cases['six-build-party']=('flow',recruit+' > RET > > RET SPACE')
results = []
xlog = (OUT / 'xvfb-page-v7.log').open('w')
xserver = subprocess.Popen(['Xvfb', ':99', '-screen', '0', '2048x1536x24', '-nolisten', 'tcp'], stdout=xlog, stderr=subprocess.STDOUT)
env = dict(os.environ, DISPLAY=':99', LIBGL_ALWAYS_SOFTWARE='1')
try:
    for attempt in range(50):
        check = subprocess.run(['xdotool', 'getdisplaygeometry'], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        if check.returncode == 0:
            break
        time.sleep(.1)
    else:
        raise RuntimeError('Xvfb not ready')
    for case, (scene, script) in cases.items():
        if len(sys.argv) > 1 and case not in sys.argv[1:]:
            continue
        traces = {}
        for binary, graphics in [('baseline', 'ega'), ('prototype', 'ega'), ('prototype', 'hd'), ('baseline', 'hd')]:
            dest = OUT / os.environ.get('KOL_PAGE_OUTPUT', 'player-pages-v2') / case / (binary + '-' + graphics)
            dest.mkdir(parents=True, exist_ok=True)
            assert dest.stat().st_uid == 1000
            cmd = [str(OUT / 'bin' / ('equipment-formal' if binary == 'prototype' else 'formal-actions')), '-scene', scene, '-'+scene+'-script', script,
                   '-seed', '1', '-lang', 'zh-Hant', '-graphics', graphics,
                   '-music', 'off', '-scale', '1', '-save', str(dest / 'save.dat'),
                   '-shot-dir', str(dest / 'shots'), '-trace', str(dest / 'trace.json'),
                   '-text-log', str(dest / 'text.jsonl')]
            (dest / 'command.json').write_text(json.dumps(cmd, ensure_ascii=False, indent=2) + '\n')
            with (dest / 'run.log').open('w') as log:
                subprocess.run(cmd, cwd=ROOT, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=120, check=True)
            trace = json.loads((dest / 'trace.json').read_text())
            traces[binary + '-' + graphics] = trace
            results.append(dict(case=case, variant=binary+'-'+graphics,
                                trace_sha256=hashlib.sha256((dest / 'trace.json').read_bytes()).hexdigest(),
                                frames=len(list((dest / 'shots').glob('*.png')))))
            print(case, binary, graphics, 'rendered', flush=True)
        assert all(t == traces['baseline-ega'] for t in traces.values()), (case, 'controller state mismatch')
        print(case, 'all controller traces equal', flush=True)
    (OUT / os.environ.get('KOL_PAGE_RESULT', 'page-prototype-results.json')).write_text(json.dumps(dict(status='DRAFT', results=results), indent=2) + '\n')
finally:
    xserver.terminate()
    try:
        xserver.wait(timeout=5)
    except subprocess.TimeoutExpired:
        xserver.kill()
        xserver.wait()
    xlog.close()
