#!/usr/bin/env python3
"""SPEC-042：Docker/Xvfb正式工具列操作、住宿保存與新程序重載。"""
import argparse
import ctypes
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time
from PIL import Image

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',type=Path,required=True)
p.add_argument('--bin',type=Path,required=True)
p.add_argument('--out',type=Path,required=True)
a=p.parse_args()
assert Path('/.dockerenv').exists() and not a.out.exists()
a.out.mkdir()
assert a.out.stat().st_uid==os.getuid()
root=a.root.resolve(); out=a.out.resolve(); binary=a.bin.resolve()
(out/'asound.conf').write_text('pcm.!default { type null }\n')
env=dict(os.environ,DISPLAY=':94',HOME='/tmp',XDG_CONFIG_HOME=str(out/'config'),ALSA_CONFIG_PATH=str(out/'asound.conf'),LIBGL_ALWAYS_SOFTWARE='1')
proc=None; stream=None; logfile=None; window=None; actions=[]; checks=[]
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
def xdo(*args): return subprocess.check_output(['xdotool',*map(str,args)],env=env,stderr=subprocess.DEVNULL).decode().strip()
def wait(fn,seconds=35):
    end=time.monotonic()+seconds
    while time.monotonic()<end:
        if fn(): return
        if proc is not None and proc.poll() is not None: raise RuntimeError(logfile.read_text())
        time.sleep(.05)
    raise TimeoutError(str(logfile))
def start(name,language='zh-Hant'):
    global proc,stream,logfile,window
    logfile=out/(name+'.log');stream=logfile.open('w')
    cmd=[str(binary),'-lang',language,'-scale','3' if language=='en' else '1','-seed','1','-music','off','-log-input']
    proc=subprocess.Popen(cmd,cwd=root,env=env,stdout=stream,stderr=stream)
    def visible():
        global window
        try: window=xdo('search','--onlyvisible','--name','武士傳說').splitlines()[0];return True
        except (subprocess.CalledProcessError,IndexError): return False
    wait(visible);xdo('windowmove',window,0,0);xdo('windowfocus',window);time.sleep(.8)
    actions.append(dict(start=name,command=cmd))
def key(*keys,game=False):
    for k in keys:
        before=logfile.read_text().count('input scene=')
        frame=logfile.read_text().count('toolbar-frame ')
        xdo('keydown','--window',window,k)
        if game: wait(lambda:logfile.read_text().count('input scene=')>before)
        elif k.startswith('F') or k=='Escape': wait(lambda:logfile.read_text().count('toolbar-frame ')>frame)
        else: time.sleep(.15)
        xdo('keyup','--window',window,k);time.sleep(.15)
        actions.append(dict(key=k,game=game,log=logfile.name))
def click(x,y):
    frame=logfile.read_text().count('toolbar-frame ')
    xdo('mousemove','--window',window,x,y);xdo('mousedown','--window',window,1)
    wait(lambda:logfile.read_text().count('toolbar-frame ')>frame)
    xdo('mouseup','--window',window,1);time.sleep(.25)
    actions.append(dict(click=[x,y],log=logfile.name))
def shot(name,height=696):
    subprocess.run(['ffmpeg','-y','-nostdin','-v','error','-f','x11grab','-draw_mouse','0','-video_size',f'960x{height}','-i',':94','-frames:v','1','-update','1',str(out/(name+'.png'))],env=env,check=True,timeout=15)
    actions.append(dict(screenshot=name+'.png',frame=len(actions),log=logfile.name))
    return out/(name+'.png')
def stop():
    global proc,stream
    xdo('keydown','--window',window,'ctrl+q');proc.wait(timeout=15)
    xdo('keyup','ctrl+q');assert proc.returncode==0,logfile.read_text();stream.close();proc=None
def party(): return json.loads([s.removeprefix('toolbar-party ') for s in logfile.read_text().splitlines() if s.startswith('toolbar-party ')][-1])
def create(*letters):
    key('Return',*letters,game=True)
    key('Return','Return','Return','Return','Down','Return','Return',game=True)
def input_field(x,y,value):
    click(x,y);key('ctrl+a');
    for c in value: key(c)
def latest_state(): return re.findall(r'display action=\S+ hd=\S+ state=(\w+)',logfile.read_text())[-1]
xlog=(out/'xvfb.log').open('w')
xvfb=subprocess.Popen(['Xvfb',':94','-screen','0','1280x800x24','-nolisten','tcp'],stdout=xlog,stderr=xlog)
try:
    time.sleep(.4)
    x11=ctypes.CDLL('libX11.so.6');x11.XOpenDisplay.restype=ctypes.c_void_p
    for f in ('XAutoRepeatOff','XFlush','XCloseDisplay'):getattr(x11,f).argtypes=[ctypes.c_void_p]
    display=x11.XOpenDisplay(b':94');assert display
    x11.XAutoRepeatOff(display);x11.XFlush(display)
    start('create-and-strengthen','en')
    key('Return','Return','5','Down',game=True)
    create('a','shift+b','c','shift+d');create('shift+z','y','shift+x')
    save=out/'config/kol-remake-cht/save.dat';data=json.loads(save.read_text())
    assert data['characters'][14]['name']=='aBcD' and data['characters'][15]['name']=='ZyX'
    key('Escape','6',*(['Down']*14),'Return','space',game=True)
    key('less','Return','Return',game=True)
    c=data['characters'][14]
    names=[v['name'] for v in data['characters'] if v and v['name']!=c['name'] and v['place']==c['place'] and v['town_code']==c['town_code']]
    key(*(['Down']*names.index('ZyX')),'Return',game=True)
    before_save=sha(save)
    key('F8');shot('english-quick-controls')
    assert 'enemy_multiplier=1' in logfile.read_text()
    click(80,72);click(280,72);click(460,72)
    assert 'no_fatigue=true enemy_multiplier=6 fear_immune=true' in logfile.read_text()
    # 設定開啟期間，按方向／Enter之外的遊戲鍵及長按方向不送入遊戲。
    state=latest_state();before_inputs=logfile.read_text().count('input scene=')
    key('1','space');xdo('keydown','--window',window,'Down');time.sleep(.7);xdo('keyup','--window',window,'Down');time.sleep(.2)
    assert logfile.read_text().count('input scene=')==before_inputs
    checks.append('輔助頁隔離遊戲鍵與長按方向')
    click(650,72)
    input_field(490,280,'110');click(720,280)
    first=party();speaker=next(i for i,m in enumerate(first) if m['speaker']);other=1-speaker
    assert first[speaker]['sheet']['stats']==[165]*7
    assert first[other]['sheet']['stats']==data['characters'][14+other]['sheet']['stats']
    click(720,340);first=party()
    assert all(w['offence']>=200 and w['defence_raw']>=2000 for w in first[speaker]['sheet']['weapons'])
    assert first[other]['sheet']['weapons']==data['characters'][14+other]['sheet']['weapons']
    click(720,400);assert party()[speaker]['gold']==9000 and party()[other]['gold']==1500
    click(650,220);click(720,400);assert [m['gold'] for m in party()]==[9000,9000]
    assert sha(save)==before_save,'永久修改在住宿之前不應自動寫遊戲保存'
    checks.append('目前發言者與全隊目標、屬性技能金幣永久修改、未住宿不寫保存')
    key('Escape','Escape')
    key('less','less','Return','Return','Return','Return',game=True)
    saved=json.loads(save.read_text());assert [saved['characters'][i]['gold'] for i in (14,15)]==[8940,8940]
    assert saved['characters'][14+speaker]['sheet']['stats']==[165]*7
    assert saved['characters'][14+other]['sheet']['stats']==data['characters'][14+other]['sheet']['stats']
    checks.append('正常旅館住宿扣款60並保存永久數值')
    stop();save_sha=sha(save)
    start('cold-reload')
    shot('toolbar',648)
    key('F8');shot('quick-controls')
    assert 'no_fatigue=false enemy_multiplier=1 fear_immune=false' in logfile.read_text()
    click(650,72);input_field(490,280,'110');shot('character-controls')
    key('Escape');click(80,72);click(280,72);click(460,72);shot('quick-controls-enabled')
    assert 'no_fatigue=true enemy_multiplier=6 fear_immune=true' in logfile.read_text()
    click(850,72);assert 'no_fatigue=false enemy_multiplier=1 fear_immune=false' in logfile.read_text()
    key('Escape')
    # 相同狀態切換原貌／HD，遊戲畫面與游標映射一併驗證。
    click(150,24);time.sleep(1);key('F8');shot('hd-quick-controls');key('Escape')
    click(40,24);key('F10');shot('display-settings',648);key('Escape')
    key('F9','F9','F7');shot('paused',648);key('F7')
    key('F12');shot('hidden',600);key('F12')
    # 正常人物頁快捷鍵須在設定返回後有效。
    key(str(speaker+1),game=True);shot('reloaded-character',648)
    assert sha(save)==save_sha
    stop()
    checks.append('新程序無-load重載、輔助重設關閉、EGA與HD切換、F7/F9/F10/F12、回到人物頁')
    # 確認正式PNG字模涵蓋全部工具列介面文案。
    manifest=json.loads((root/'cmd/kol/toolbarfont/manifest.json').read_text())
    chars={chr(g['codepoint']) for g in manifest['fonts'][0]['glyphs']}
    source=(root/'cmd/kol/toolbar.go').read_text()
    literals=''.join(json.loads(s) for s in re.findall(r'"(?:[^"\\]|\\.)*"',source))
    assert all(c in chars for c in literals if ord(c)>=0x3000)
    program={str(f.relative_to(root)):sha(f) for d in ('cmd','internal') for f in sorted((root/d).rglob('*.go'))}
    program.update({n:sha(root/n) for n in ('go.mod','go.sum')})
    program.update({str(f.relative_to(root)):sha(f) for f in sorted((root/'cmd/kol/toolbarfont').glob('*')) if f.is_file()})
    receipt=dict(status='PASS',spec='docs/spec/042-assist-toolbar.md',binary_sha256=sha(binary),verifier_sha256=sha(Path(__file__)),program_inputs=program,seed=1,platform='Linux Docker/Xvfb',original_parity_claimed=False,checks=checks,actions=actions,screenshots={f.name:sha(f) for f in out.glob('*.png')},save_sha256=sha(save),source_zip_sha256=sha(root/'Knights-of-legend.zip'),rights='原作衍生遊戲畫面PNG，依2026-10-17公開化決定保存；原作權利未改變',capture='正式互動程式X11畫面，frame為actions序號；沒有使用原版oracle',font_source_sha256=sha(root/'cmd/kol/toolbarfont/source.json'))
    (out/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(status='PASS',checks=checks),ensure_ascii=False))
finally:
    if proc is not None and proc.poll() is None:proc.terminate();proc.wait(timeout=10)
    if stream is not None and not stream.closed:stream.close()
    xvfb.terminate();xvfb.wait(timeout=10);xlog.close()
