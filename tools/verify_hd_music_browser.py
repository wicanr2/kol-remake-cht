#!/usr/bin/env python3
"""只驗瀏覽器音訊載入與播放控制，不宣稱人耳驗收。"""
import ast,base64,hashlib,json,os,socket,struct,subprocess,time,urllib.request
from pathlib import Path
from urllib.parse import urlparse
assert Path('/.dockerenv').exists()
ROOT=Path('/src');OUT=Path('/out')
source=ROOT/'docs/verification/hd-ui-theme-20261012/verify_browser.py'
cls=next(n for n in ast.parse(source.read_text()).body if isinstance(n,ast.ClassDef) and n.name=='CDP')
exec(compile(ast.Module(body=[cls],type_ignores=[]),str(source),'exec'))
profile=Path('/tmp/kol-music-review-browser');profile.mkdir()
with (OUT/'browser.log').open('w') as log:
    chrome=subprocess.Popen(['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--mute-audio','--allow-file-access-from-files','--remote-debugging-port=0','--user-data-dir='+str(profile),'about:blank'],stdout=log,stderr=log)
    try:
        for _ in range(100):
            if (profile/'DevToolsActivePort').exists():break
            assert chrome.poll() is None;time.sleep(.05)
        port=int((profile/'DevToolsActivePort').read_text().splitlines()[0]);pages=json.load(urllib.request.urlopen('http://127.0.0.1:'+str(port)+'/json'))
        c=CDP(next(p['webSocketDebuggerUrl'] for p in pages if p['type']=='page'))
        c.call('Page.navigate',{'url':(ROOT/'docs/verification/hd-music-20261012/review.html').as_uri()})
        for _ in range(100):
            ready=c.evaluate('[...document.querySelectorAll("audio")].map(a=>({id:a.id,ready:a.readyState,duration:a.duration,error:a.error&&a.error.code}))')
            if len(ready)==3 and all(a['ready']>=1 for a in ready):break
            time.sleep(.05)
        assert len(ready)==3 and all(a['ready']>=1 and not a['error'] for a in ready),ready
        assert abs(next(a['duration'] for a in ready if a['id']=='title')-72)<.01
        assert abs(next(a['duration'] for a in ready if a['id']=='inn')-80)<.01
        def click(selector):
            c.evaluate('document.querySelector('+json.dumps(selector)+').scrollIntoView()')
            r=c.evaluate('(()=>{const r=document.querySelector('+json.dumps(selector)+').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()')
            for t in ('mousePressed','mouseReleased'):c.call('Input.dispatchMouseEvent',dict(type=t,x=r['x'],y=r['y'],button='left',clickCount=1))
        click('#preview');time.sleep(3.5)
        playing=c.evaluate('[...document.querySelectorAll("audio")].map(a=>({id:a.id,time:a.currentTime,paused:a.paused,error:a.error&&a.error.code}))')
        assert next(a for a in playing if a['id']=='title')['time']>2
        assert next(a for a in playing if a['id']=='thunder')['time']>0
        click('#stop');time.sleep(.2);assert c.evaluate('[...document.querySelectorAll("audio")].every(a=>a.paused&&a.currentTime===0)')
        report=dict(schema='kol-hd-music-browser/1',status='PASS',browser=c.call('Browser.getVersion')['product'],metadata=ready,real_click_playback=playing,stop_all=True,human_listening=False,audio_output_muted=True)
        (OUT/'browser-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('BROWSER MUSIC PASS: OGG metadata, real play/stop, thunder timer; no human listening claim')
    finally:
        chrome.terminate()
        try:chrome.wait(timeout=5)
        except subprocess.TimeoutExpired:chrome.kill();chrome.wait(timeout=5)
