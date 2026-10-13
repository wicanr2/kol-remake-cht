#!/usr/bin/env python3
"""正常玩家輸入＋實際Ebitengine/Oto/PulseAudio輸出；Docker限定。"""
import argparse,hashlib,json,os,re,signal,subprocess,time,wave
from pathlib import Path
import numpy as np
assert Path('/.dockerenv').exists()
ROOT=Path('/src');OUT=Path('/out');checks=[]
def waitfor(fn,seconds=10):
    until=time.monotonic()+seconds
    while time.monotonic()<until:
        value=fn()
        if value:return value
        time.sleep(.05)
    raise AssertionError('timed out')
def xdo(*args):return subprocess.check_output(['xdotool',*map(str,args)],text=True,timeout=10).strip()
def window(proc,text):
    assert proc.poll() is None,text()
    found=subprocess.run(['xdotool','search','--name','^武士傳說$'],capture_output=True,text=True,timeout=10)
    return found.stdout.splitlines()[-1] if found.returncode==0 else None
def key(w,*names):
    for n in names:xdo('keydown','--window',w,n);time.sleep(.12);xdo('keyup','--window',w,n);time.sleep(.12)
def click(w,x,y):xdo('mousemove','--window',w,x,y);xdo('click','1');time.sleep(.25)
def shot(name,w=1920,h=1200):subprocess.run(['ffmpeg','-y','-v','error','-f','x11grab','-draw_mouse','0','-video_size',f'{w}x{h}','-i',':99','-frames:v','1','-update','1',str(OUT/name)],check=True,timeout=10)
def quitgame(proc,w):
    xdo('keydown','--window',w,'ctrl');xdo('keydown','--window',w,'q')
    proc.wait(timeout=10)
    subprocess.run(['xdotool','keyup','q','ctrl'],check=True,timeout=10)
    assert proc.returncode==0
def record(name,body,load=None,missing=False,scene='game',extra=()):
    log=OUT/(name+'.log');wav=OUT/(name+'.wav');argv=[str(OUT/'kol'),'-scene','game','-lang','zh-Hant','-graphics','ega' if name=='title' else 'hd','-scale','1','-seed','1','-log-input','-text-log',str(OUT/(name+'-text.jsonl')),'-save',str(OUT/(name+'-save.json'))]
    argv[argv.index('-scene')+1]=scene
    argv+=list(map(str,extra))
    if load:argv+=['-load',str(load)]
    if missing:argv+=['-music-assets',str(OUT/'missing-music.json')]
    commands=[]
    with (OUT/(name+'-capture.log')).open('w') as stream,log.open('w') as glog:
        recorder=subprocess.Popen(['ffmpeg','-y','-nostdin','-v','error','-f','pulse','-sample_rate','44100','-channels','2','-i','kol_music.monitor','-c:a','pcm_s16le',str(wav)],stdout=stream,stderr=stream)
        time.sleep(.3);proc=subprocess.Popen(argv,cwd=ROOT,stdout=glog,stderr=subprocess.STDOUT)
        def text():return log.read_text()
        try:
            w=waitfor(lambda: window(proc,text),30)
            xdo('windowfocus','--sync',w);body(proc,w,text,commands)
            time.sleep(1.7);quitgame(proc,w);time.sleep(.4)
        finally:
            if proc.poll() is None:proc.terminate();proc.wait(timeout=5)
            recorder.send_signal(signal.SIGINT);recorder.wait(timeout=5)
    assert not (OUT/(name+'-save.json')).exists(),'presentation wrote save'
    with wave.open(str(wav),'rb') as f:
        assert f.getnchannels()==2 and f.getframerate()==44100 and f.getsampwidth()==2
        samples=np.frombuffer(f.readframes(f.getnframes()),dtype='<i2').reshape(-1,2)
    peak=int(np.max(np.abs(samples.astype(np.int32))));assert peak<32767
    report=dict(name=name,command=argv,markers=commands,seconds=len(samples)/44100,peak=peak,sha256=hashlib.sha256(wav.read_bytes()).hexdigest(),no_save_written=True)
    (OUT/(name+'-record.json')).write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    checks.append(report);return report,samples,log.read_text()

def title(proc,w,text,markers):
    waitfor(lambda:'started=23 remaining=0' in text(),110)
    assert len(re.findall(r'audio cue=title-thunder ',text()))==23
    key(w,'F10');waitfor(lambda:'display-frame hd=false open=true' in text())
    # 設定面板暫停標題動畫，音樂繼續，等待真實72秒循環跨點。
    def loop_position():
        key(w,'Return')
        rows=re.findall(r'music-position cue=title enabled=true title_ms=(\d+)',text())
        return rows and int(rows[-1])>73000
    waitfor(loop_position,30)
    key(w,'F10');waitfor(lambda:'display-frame hd=false open=false' in text())
    positions=re.findall(r'music-position cue=title enabled=true title_ms=(\d+)',text());assert int(positions[-1])>72000,'title did not cross actual loop'
    key(w,'space');waitfor(lambda:'music cue= enabled=true' in text());time.sleep(.6)
    markers.append(dict(action='normal 23 cues, real loop, keyboard exit',title_position_ms=int(positions[-1])))

def inn(proc,w,text,markers):
    key(w,'Left','Right','Return','space');waitfor(lambda:'music cue=inn enabled=true' in text(),15);started=time.monotonic();time.sleep(1)
    key(w,'F10','Up','Return');waitfor(lambda:'display-frame hd=false open=true cursor=0' in text())
    key(w,'Down','Return');waitfor(lambda:'display-frame hd=true open=true cursor=1' in text());key(w,'Escape')
    positions=[int(i) for i in re.findall(r'music-position cue=inn enabled=true title_ms=\d+ inn_ms=(\d+)',text())]
    assert len(positions)>=3 and all(b>=a for a,b in zip(positions,positions[1:])),'graphics switch restarted music'
    key(w,'F9');waitfor(lambda:'display action=music-false' in text());time.sleep(.65);key(w,'F9');waitfor(lambda:'display action=music-true' in text())
    shot('inn-controls-hd.png')
    xdo('windowsize',w,960,600);time.sleep(.4);click(w,792,588);time.sleep(.4);click(w,792,588);shot('inn-controls-small.png',960,600)
    rows=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',text());assert len(set(rows))==1,'F9/F10 altered game state'
    key(w,'F10','Up','Return');waitfor(lambda:'display-frame hd=false open=true cursor=0' in text())
    def loop_position():
        time.sleep(.8);key(w,'Return')
        rows=re.findall(r'music-position cue=inn enabled=true title_ms=\d+ inn_ms=(\d+)',text())
        return rows and int(rows[-1])>81000
    waitfor(loop_position,120)
    positions=re.findall(r'music-position cue=inn enabled=true title_ms=\d+ inn_ms=(\d+)',text());assert int(positions[-1])>80000,'inn did not cross actual loop'
    key(w,'Escape','Return','space');waitfor(lambda:'music cue= enabled=true' in text(),10)
    markers.append(dict(action='same state F9/F10 and resized mouse, normal inn exit',positions_ms=list(map(int,positions))))

def missing(proc,w,text,markers):
    waitfor(lambda:'配樂無法使用' in text(),10);key(w,'F9');waitfor(lambda:'display action=music-false' in text());shot('missing-music.png')
    assert 'music cue=' not in text();key(w,'F10','F10');markers.append(dict(action='missing package stays off, F10 remains usable'))

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--parts',nargs='+',choices=['title','inn','missing'],default=['title','inn','missing'])
    args=parser.parse_args()
    if 'title' in args.parts:
        report,samples,_=record('title',title);assert np.max(np.abs(samples[-44100:].astype(np.int32)))==0
    if 'inn' in args.parts:
        report,samples,_=record('inn',inn,OUT/'inputs/save.json');assert np.max(np.abs(samples[-44100:].astype(np.int32)))==0
    if 'missing' in args.parts:
        report,samples,log=record('missing',missing,OUT/'inputs/save.json',True);assert np.max(np.abs(samples.astype(np.int32)))==0
        rows=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',log);assert len(set(rows))==1
    (OUT/'runtime-audio.json').write_text(json.dumps(dict(schema='kol-music-runtime-audio/1',status='PASS',device='private PulseAudio null sink monitor; actual Ebitengine/Oto output at 44100Hz stereo',runs=checks,original_parity_claimed=False,real_machine_listening_claimed=False),ensure_ascii=False,indent=2)+'\n')
    print('RUNTIME AUDIO PASS: normal title/inn, loops, F9/F10, resize, missing package, exit silence',flush=True)
if __name__=='__main__':main()
