#!/usr/bin/env python3
"""Docker限定：正式存檔路徑與準備世界場景的實際配樂輸出。"""
import argparse,hashlib,json,re,subprocess,time
from pathlib import Path
from PIL import Image,ImageChops
from verify_music_runtime import ROOT,OUT,key,xdo,window,waitfor,quitgame,record,shot,np
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,data):p.write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
commands=[]
parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--reuse-saves',action='store_true');args=parser.parse_args()
for enabled in (False,True):
    name='stay-music-'+str(enabled).lower();save=OUT/(name+'.json')
    log=OUT/(name+'.log')
    argv=[str(OUT/'kol'),'-scene','game','-lang','zh-Hant','-graphics','hd','-scale','1','-seed','1','-music','hd' if enabled else 'off','-load',str(OUT/'inputs/stay-v3.json'),'-save',str(save),'-log-input']
    if args.reuse_saves:
        assert save.exists() and 'token=[RET]' in log.read_text();commands.append(argv);continue
    assert not save.exists()
    with log.open('w') as stream:
        proc=subprocess.Popen(argv,cwd=ROOT,stdout=stream,stderr=subprocess.STDOUT)
        try:
            w=waitfor(lambda:window(proc,log.read_text),30);xdo('windowfocus','--sync',w);time.sleep(.7)
            key(w,'Left','Right','Return','space')
            if enabled:
                waitfor(lambda:'music cue=inn enabled=true' in log.read_text())
                key(w,'F9','F9','F10','Up','Return','Down','Return','Escape')
            key(w,'period','period','period','period','Return','Return')
            waitfor(save.exists,15);quitgame(proc,w)
        finally:
            if proc.poll() is None:proc.terminate();proc.wait(timeout=5)
    commands.append(argv)
saves=[json.loads((OUT/('stay-music-'+str(e).lower()+'.json')).read_text()) for e in (False,True)]
for data in saves:data.pop('saved_at',None)
assert saves[0]==saves[1],'music/settings changed actual v3 save'
for enabled in (False,True):
    dest=OUT/('reload-'+str(enabled).lower());dest.mkdir(exist_ok=args.reuse_saves)
    argv=[str(OUT/'kol'),'-scene','play','-lang','zh-Hant','-graphics','hd','-scale','1','-seed','1','-load',str(OUT/('stay-music-'+str(enabled).lower()+'.json')),'-music','hd' if enabled else 'off','-play-script','L R RET SPACE > RET','-trace',str(dest/'trace.json'),'-text-log',str(dest/'text.jsonl'),'-shot-dir',str(dest)]
    subprocess.run(argv,cwd=ROOT,check=True,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,timeout=60);commands.append(argv)
a,b=OUT/'reload-false',OUT/'reload-true'
assert (a/'trace.json').read_bytes()==(b/'trace.json').read_bytes()
assert (a/'text.jsonl').read_bytes()==(b/'text.jsonl').read_bytes()
frames=list(a.glob('step-*.png'));assert len(frames)==7
frame_checks=[]
for p in frames:
    left=np.asarray(Image.open(p).convert('RGB')).astype(np.int16);right=np.asarray(Image.open(b/p.name).convert('RGB')).astype(np.int16)
    delta=np.abs(left-right);count=int(np.count_nonzero(delta.max(2)));maximum=int(delta.max())
    # 原生HD線性過濾的軟體GPU舍入；存檔、狀態與文字仍要求完全相同。
    assert maximum<=1 and count<=1000,(p.name,maximum,count)
    frame_checks.append(dict(frame=p.name,exact=p.read_bytes()==(b/p.name).read_bytes(),max_channel_delta=maximum,different_pixels=count))

# 準備既有受控任務的相鄰世界落點，再透過正常方向鍵與確認進入戰鬥。
# 這是場景準備，不能證明從城鎮正常走完整趟旅程或原版parity。
fixture=json.loads((OUT/'inputs/final.json').read_text());fixture['location']={'kind':'world'}
fixture['world']={'x':47,'y':9,'sub_x':7,'sub_y':4}
prepared=OUT/'inputs/music-world-v1.json';write(prepared,fixture)
converted=OUT/'inputs/music-world-v3.json'
argv=[str(OUT/'kol'),'-scene','play','-lang','zh-Hant','-music','off','-load',str(prepared),'-save',str(converted),'-play-script','>','-shot-dir',str(OUT/'world-import')]
with (OUT/'world-import.log').open('w') as log:
    subprocess.run(argv,cwd=ROOT,check=True,stdout=log,stderr=subprocess.STDOUT,timeout=60)
commands.append(argv)
assert json.loads(converted.read_text())['schema']=='kol-save-v3'
def adventure(proc,w,text,markers):
    waitfor(lambda:'music cue=exploration enabled=true' in text(),15);time.sleep(3);shot('world-music-hd.png')
    key(w,'Right','Return','space');waitfor(lambda:'music cue=combat enabled=true' in text(),15);time.sleep(4);shot('combat-music-hd.png')
    key(w,'F10','Up','Return','Down','Return','Escape','F9');time.sleep(.7);key(w,'F9');time.sleep(2)
    # HD軟體繪圖的淡出更新與裝置佇列要完成，才截取靜音尾端。
    key(w,'F9');time.sleep(3.5)
    markers.append(dict(action='prepared world -> normal Right, Return, Space -> encounter/field; F10 both themes; F9 mute/resume',prepared_scene=True,original_parity_claimed=False))
report,samples,log=record('adventure',adventure,converted,extra=['-quest-sites',OUT/'inputs/sites-final.json'])
assert np.max(np.abs(samples.astype(np.int32)))>100
assert np.max(np.abs(samples[-44100:].astype(np.int32)))==0
rows=re.findall(r'display action=\S+ hd=\S+ state=([0-9a-f]+)',log);assert len(set(rows))==1
write(OUT/'player-paths.json',dict(schema='kol-music-player-paths/1',status='PASS',commands=commands,normal_stay_save_equal_except_saved_at=True,reused_same_build_saves=args.reuse_saves,new_process_reload_states=7,new_process_reload_frames=frame_checks,prepared_inputs={p.name:sha(p) for p in [prepared,converted,OUT/'inputs/sites-final.json']},adventure=report,original_parity_claimed=False,limits='住宿與重載為正常UI；HD線性過濾容許每通道1階且每圖最多1000像素的GPU舍入差；探索／戰鬥由準備存檔開始，只驗場景配樂、中文與控制，不宣稱全旅程或原版parity。'))
print('PLAYER PATHS PASS: normal stay/save/reload; world -> combat real audio, F9/F10, same state',flush=True)
