#!/usr/bin/env python3
"""驗配樂控制候選的真實鍵鼠與連續播放，不聲稱正式遊戲。"""
import ast,base64,hashlib,json,os,socket,struct,subprocess,time,urllib.request
from pathlib import Path
from urllib.parse import urlparse
assert Path('/.dockerenv').exists()
ROOT=Path('/src');OUT=Path('/out');source=ROOT/'docs/verification/hd-ui-theme-20261012/verify_browser.py'
cls=next(n for n in ast.parse(source.read_text()).body if isinstance(n,ast.ClassDef) and n.name=='CDP');exec(compile(ast.Module(body=[cls],type_ignores=[]),str(source),'exec'))
profile=Path('/tmp/kol-music-control');profile.mkdir();checks=[]
with (OUT/'browser.log').open('w') as log:
    chrome=subprocess.Popen(['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--mute-audio','--allow-file-access-from-files','--remote-debugging-port=0','--user-data-dir='+str(profile),'about:blank'],stdout=log,stderr=log)
    try:
        for _ in range(100):
            if (profile/'DevToolsActivePort').exists():break
            assert chrome.poll() is None;time.sleep(.05)
        port=int((profile/'DevToolsActivePort').read_text().splitlines()[0]);pages=json.load(urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/json'));c=CDP(next(p['webSocketDebuggerUrl'] for p in pages if p['type']=='page'))
        def nav(capture=False):
            c.call('Emulation.setDeviceMetricsOverride',dict(width=1920,height=1200,deviceScaleFactor=1,mobile=False));c.call('Page.navigate',{'url':(ROOT/'docs/verification/hd-music-control-20261013/review.html').as_uri()+('?capture=1' if capture else '')})
            for _ in range(100):
                if c.evaluate('document.body&&document.body.dataset.ready')=='true':break
                time.sleep(.05)
            assert c.evaluate('document.body.dataset.ready')=='true'
        def click(selector):
            c.evaluate('document.querySelector('+json.dumps(selector)+').scrollIntoView({block:"center"})');r=c.evaluate('(()=>{const r=document.querySelector('+json.dumps(selector)+').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()')
            for t in ('mousePressed','mouseReleased'):c.call('Input.dispatchMouseEvent',dict(type=t,**r,button='left',clickCount=1))
        def key():
            for kind in ('keyDown','keyUp'):c.call('Input.dispatchKeyEvent',dict(type=kind,key='F9',windowsVirtualKeyCode=120))
        def ok(name,value):checks.append(dict(name=name,passed=bool(value)));assert value,name
        nav();click('#start');time.sleep(.4);before=c.evaluate('document.querySelector("audio").currentTime');identity=c.evaluate('musicReview.identity()');ok('actual audio plays',before>0)
        click('#graphics');time.sleep(.3);ok('A EGA keeps audio',c.evaluate('!document.querySelector("audio").paused&&document.querySelector("audio").currentTime>='+str(before)))
        key();ok('F9 off',c.evaluate('!musicReview.state.enabled&&document.querySelector("audio").paused'));key();time.sleep(.2);ok('F9 resumes',c.evaluate('musicReview.state.enabled&&!document.querySelector("audio").paused'))
        click('[data-mode="theme"]');ok('B EGA silent',c.evaluate('document.querySelector("audio").paused'));click('#graphics');time.sleep(.2);ok('B HD plays',c.evaluate('!document.querySelector("audio").paused'))
        ok('scene identity retained',identity==c.evaluate('musicReview.identity()'))
        nav(True);screens={}
        for mode,gfx,name in [('independent','hd','a-hd.png'),('independent','ega','a-ega.png'),('theme','hd','b-hd.png')]:
            c.evaluate('Object.assign(musicReview.state,'+json.dumps(dict(mode=mode,graphics=gfx))+');musicReview.sync()');data=c.call('Page.captureScreenshot',dict(format='png',captureBeyondViewport=False))['data'];(OUT/name).write_bytes(base64.b64decode(data));screens[name]=hashlib.sha256((OUT/name).read_bytes()).hexdigest()
        report=dict(schema='kol-music-control-prototype/1',status='PASS',production_changed=False,checks=checks,screenshots=screens,browser=c.call('Browser.getVersion')['product'],human_listening=False,original_parity_claimed=False)
        (OUT/'browser-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('PROTOTYPE PASS',len(checks),'checks')
    finally:
        chrome.terminate()
        try:chrome.wait(timeout=5)
        except subprocess.TimeoutExpired:chrome.kill();chrome.wait(timeout=5)
