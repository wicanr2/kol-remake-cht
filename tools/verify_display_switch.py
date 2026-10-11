#!/usr/bin/env python3
"""Docker內驗證正式遊戲畫面設定；/src為專案，/out為含inputs的本輪輸出。"""
from pathlib import Path
import argparse, hashlib, json, re, subprocess, time
from PIL import Image, ImageChops

root=Path('/src'); out=Path('/out')
assert Path('/.dockerenv').exists()
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--binary',type=Path,default=out/'kol')
args=parser.parse_args()
events=[]
def xdo(*args):
    events.append(['xdotool',*map(str,args)])
    return subprocess.check_output(events[-1],text=True,timeout=10).strip()
def shot(name,size=(1920,1200)):
    subprocess.run(['ffmpeg','-y','-loglevel','error','-f','x11grab','-draw_mouse','0','-video_size',f'{size[0]}x{size[1]}','-i',':99','-frames:v','1','-update','1',str(out/name)],check=True,timeout=10)
def click(w,x,y):
    xdo('mousemove','--window',w,x,y);xdo('click','1');time.sleep(.3)
def key(w,*names):
    # 軟體HD繪圖可能跨過短按鍵；按下與放開均保留數個更新幀。
    for name in names:
        xdo('keydown','--window',w,name);time.sleep(.12)
        xdo('keyup','--window',w,name);time.sleep(.12)
    time.sleep(.3)
def dimensions(w):
    text=xdo('getwindowgeometry','--shell',w)
    return [int(re.search(k+r'=(\d+)',text)[1]) for k in ('WIDTH','HEIGHT')]
def execute(name,missing=False):
    log=out/(name+'.log')
    argv=[str(args.binary),'-scene','game','-lang','zh-Hant','-graphics','ega','-scale','2','-seed','1','-load',str(out/'inputs/save.json'),'-save',str(out/(name+'-save.json')),'-log-input','-text-log',str(out/(name+'-text.jsonl'))]
    if missing:argv+=['-hd-assets',str(out/'missing-hd')]
    def action(fn,frame):
        before=log.stat().st_size
        fn()
        for _ in range(100):
            if frame in log.read_text()[before:]:
                time.sleep(.4)
                return
            time.sleep(.1)
        raise AssertionError(('frame not rendered',frame,log.read_text()[before:]))
    with log.open('w') as stream:
        proc=subprocess.Popen(argv,cwd=root,stdout=stream,stderr=subprocess.STDOUT)
        try:
            w=None
            for _ in range(80):
                time.sleep(.1)
                found=subprocess.run(['xdotool','search','--name','^武士傳說$'],capture_output=True,text=True)
                if found.returncode==0:w=found.stdout.splitlines()[-1];break
                assert proc.poll() is None,log.read_text()
            assert w,'window not found'
            xdo('windowfocus','--sync',w);time.sleep(.7)
            key(w,'Left','Right');click(w,264,1086);key(w,'space')
            shot(name+'-ega.png')
            action(lambda:click(w,1790,1175),'display-frame hd=false open=true cursor=0');shot(name+'-menu-ega.png')
            assert dimensions(w)==[1920,1200]
            action(lambda:click(w,1500,735),f'display-frame hd={str(not missing).lower()} open=true cursor=1');shot(name+'-menu-hd.png')
            assert dimensions(w)==[1920,1200],'switch resized the window'
            key(w,'1','Left','Right')
            action(lambda:key(w,'Escape'),f'display-frame hd={str(not missing).lower()} open=false');shot(name+'-closed.png')
            if missing:
                a=Image.open(out/(name+'-ega.png')).convert('RGB')
                b=Image.open(out/(name+'-closed.png')).convert('RGB')
                assert not ImageChops.difference(a,b).getbbox(),'missing pack changed EGA'
            else:
                click(w,500,150)
                action(lambda:key(w,'F10'),'display-frame hd=true open=true cursor=1')
                key(w,'Up')
                action(lambda:key(w,'Return'),'display-frame hd=false open=true cursor=0');shot(name+'-keyboard-ega.png')
                key(w,'Down')
                action(lambda:key(w,'Return'),'display-frame hd=true open=true cursor=1')
                key(w,'Down')
                action(lambda:key(w,'Return'),'display-frame hd=true open=false cursor=2')
                xdo('windowsize',w,960,600);time.sleep(.5)
                action(lambda:key(w,'F10'),'display-frame hd=true open=true cursor=1')
                action(lambda:click(w,750,285),'display-frame hd=false open=true cursor=0')
                action(lambda:click(w,750,370),'display-frame hd=true open=true cursor=1')
                action(lambda:click(w,800,495),'display-frame hd=true open=false');click(w,120,75)
                shot(name+'-resized.png',(960,600))
                action(lambda:key(w,'F10'),'display-frame hd=true open=true cursor=1')
                action(lambda:xdo('click','3'),'display-frame hd=true open=false')
                action(lambda:key(w,'F10'),'display-frame hd=true open=true cursor=1')
                action(lambda:key(w,'F10'),'display-frame hd=true open=false')
            xdo('keydown','--window',w,'ctrl');time.sleep(.1);xdo('keydown','--window',w,'q')
            proc.wait(timeout=10);xdo('keyup','q','ctrl')
            assert proc.returncode==0,log.read_text()
        finally:
            if proc.poll() is None:proc.terminate();proc.wait(timeout=5)
    assert not (out/(name+'-save.json')).exists(),'display operations wrote a save'
    text=log.read_text()
    rows=re.findall(r'display action=(\S+) hd=(\S+) state=([0-9a-f]+)',text)
    assert len(rows)>=3,text
    assert len({s for _,_,s in rows[:3]})==1,'menu input changed game state'
    if missing:
        assert rows[1][0:2]==('hd','false'),rows
        assert 'HD素材載入失敗' in text
        labels=(out/(name+'-text.jsonl')).read_text()
        assert 'HD 圖片無法使用。' in labels and '已保留目前畫面，請稍後再試。' in labels
    else:
        assert rows[1][0:2]==('hd','true'),rows
        assert 'token=[2]' in text and 'token=[1]' in text,text
        assert len(rows)==15,rows
        assert len({s for _,_,s in rows[3:]})==1,'second settings session changed state'
        a=Image.open(out/(name+'-ega.png')).convert('RGB')
        b=Image.open(out/(name+'-closed.png')).convert('RGB')
        for i in range(6):
            r=(88+i*288,10,336+i*288,242)
            assert ImageChops.difference(a.crop(r),b.crop(r)).getbbox(),('missing HD portrait',i)
        assert not ImageChops.difference(a.crop((0,245,1920,1150)),b.crop((0,245,1920,1150))).getbbox(),'switch changed scene or text outside portraits'
    return dict(command=argv,display_events=rows,passed=True)

def stay(switched):
    name='stay-switched' if switched else 'stay-baseline'
    log=out/(name+'.log'); save=out/(name+'.json')
    assert not save.exists()
    argv=[str(args.binary),'-scene','game','-lang','zh-Hant','-graphics','ega','-scale','1','-seed','1','-load',str(out/'inputs/stay-v3.json'),'-save',str(save),'-log-input']
    with log.open('w') as stream:
        proc=subprocess.Popen(argv,cwd=root,stdout=stream,stderr=subprocess.STDOUT)
        try:
            w=None
            for _ in range(80):
                time.sleep(.1)
                found=subprocess.run(['xdotool','search','--name','^武士傳說$'],capture_output=True,text=True)
                if found.returncode==0:w=found.stdout.splitlines()[-1];break
                assert proc.poll() is None,log.read_text()
            assert w
            xdo('windowfocus','--sync',w);time.sleep(.7)
            key(w,'Left','Right','Return','space')
            if switched:
                key(w,'F10');click(w,750,370)
                for _ in range(100):
                    if 'display-frame hd=true open=true' in log.read_text():break
                    time.sleep(.1)
                else:raise AssertionError(log.read_text())
                key(w,'Escape')
            key(w,'period','period','period','period','Return','Return')
            for _ in range(50):
                if save.exists():break
                time.sleep(.1)
            assert save.exists(),log.read_text()
            xdo('keydown','--window',w,'ctrl');time.sleep(.1);xdo('keydown','--window',w,'q')
            proc.wait(timeout=10);xdo('keyup','q','ctrl')
            assert proc.returncode==0,log.read_text()
        finally:
            if proc.poll() is None:proc.terminate();proc.wait(timeout=5)
    return argv

checks=[execute('game'),execute('missing',True)]
stay_commands=[stay(False),stay(True)]
saved=[json.loads((out/(n+'.json')).read_text()) for n in ('stay-baseline','stay-switched')]
for data in saved:data.pop('saved_at',None)
assert saved[0]==saved[1],'switching changed persisted game data'
for name in ('game-menu-ega.png','game-menu-hd.png','missing-menu-hd.png'):
    before=Image.open(root/'docs/verification/display-switch-20261011'/name).convert('RGB')
    actual=Image.open(out/name).convert('RGB')
    assert not ImageChops.difference(before,actual).getbbox(),('approved prototype changed',name)
(out/'gui-check.json').write_text(json.dumps(dict(passed=True,checks=checks,events=events,stay_commands=stay_commands,stay_v3_equal_except_saved_at=True,approved_screens_unchanged=3,limits='正式程式的畫面設定驗收；未新增原版parity或平台真機證據'),ensure_ascii=False,indent=2)+'\n')
print('GUI PASS: mouse, F10, keyboard, resized click, right-click, no observed state or save changes, missing-pack fallback, three approved screens unchanged',flush=True)
