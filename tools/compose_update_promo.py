#!/usr/bin/env python3
"""Docker限定：合成同輪重新錄製的v1.0.6修正及v1.0.7工具列操作影片。"""
import argparse
import hashlib
import json
import math
from pathlib import Path
import re
import subprocess
import numpy as np
from PIL import Image,ImageDraw,ImageFont

p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--root',type=Path,required=True)
p.add_argument('--out',type=Path,required=True)
p.add_argument('--feedback',type=Path,required=True)
p.add_argument('--assist',type=Path,required=True)
p.add_argument('--work',type=Path,required=True)
p.add_argument('--font',type=Path,required=True)
a=p.parse_args()
assert Path('/.dockerenv').exists() and not a.work.exists()
a.work.mkdir()
root,out,work=a.root,a.out,a.work
promo=out/'promo';W,H,FPS=1920,1080,25
sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
write=lambda f,d:f.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
inputs=json.loads((out/'build-inputs.json').read_text());version=inputs['version']
assert re.fullmatch(r'v\.1\.0\.7-\d{8}',version)
assert sha(a.font)=='9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825'
def run(cmd,log=None,seconds=240):
    r=subprocess.run(list(map(str,cmd)),capture_output=True,timeout=seconds)
    if log:Path(log).write_bytes(r.stdout+r.stderr)
    if r.returncode:raise RuntimeError(r.stderr.decode(errors='replace')[-2500:])
    return r.stdout
def text(s,size,color=(244,232,200)):
    f=ImageFont.truetype(str(a.font),size);box=f.getbbox(s)
    im=Image.new('RGBA',(max(1,box[2]-box[0]),max(1,box[3]-box[1])))
    ImageDraw.Draw(im).text((-box[0],-box[1]),s,font=f,fill=(*color,255))
    assert im.width<=W-160 and im.height<=150,(s,im.size)
    return im
def centered(im,s,y,size=40,color=(244,232,200)):
    label=text(s,size,color);assert y>=0 and y+label.height<=H
    im.alpha_composite(label,((W-label.width)//2,y))
def overlay(c):
    im=Image.new('RGBA',(W,H));d=ImageDraw.Draw(im)
    d.rectangle((0,934,W,H),fill=(32,20,20,255));accent=(104,177,194) if c['section']=='v1.0.7' else (200,153,46)
    d.line((80,936,W-80,936),fill=(*accent,255),width=2)
    centered(im,c['caption'],964,40)
    tag=text(c['section'],28,accent);im.alpha_composite(tag,(48,100))
    tag=text('新增' if c['section']=='v1.0.7' else '修正',38,accent);im.alpha_composite(tag,(48,145))
    tag=text('受控場景' if c['controlled'] else '實際鍵鼠操作',22,(186,172,145));im.alpha_composite(tag,(80,1038))
    tag=text(version,22,(186,172,145));im.alpha_composite(tag,(W-tag.width-80,1038))
    return im
def card(name,source):
    im=Image.new('RGBA',(W,H),(32,20,20,255));image=Image.open(source).convert('RGBA')
    image.thumbnail((1440,900));im.alpha_composite(image,((W-image.width)//2,(H-image.height)//2))
    im.alpha_composite(Image.new('RGBA',(W,H),(18,8,8,200)))
    centered(im,'武士傳說',260,104,(224,180,82));centered(im,version,410,42)
    if name=='intro':
        centered(im,'v1.0.6 操作修正 ＋ v1.0.7 工具列與可選輔助',530,44)
        centered(im,'正式程式重新錄製',620,34)
    else:
        centered(im,'Windows ／ Linux ／ macOS',530,46)
        centered(im,'純 EGA 完整版　｜　EGA＋HD 完整版',620,44)
        centered(im,'輔助預設關閉；永久強化由住宿保存',710,34)
    path=work/(name+'.png');im.convert('RGB').save(path);return path
old_durations=dict(name=5,save=5,resume=5,switch=5,sell=6,armour=7,magic=7,walk=5,ride=6,attack=8,hd=6)
new_durations={'assist':8,'permanent':12,'assist-save':6,'assist-resume':7,'toolbar-hd':7,'shortcuts':8}
chapters=[];sources=[]
for cap,section,durations in [(a.feedback,'v1.0.6',old_durations),(a.assist,'v1.0.7',new_durations)]:
    receipt=json.loads((cap/'receipt.json').read_text());assert receipt['status']=='PASS'
    assert receipt['binary_sha256']==inputs['binaries']['linux-x86_64']['sha256']
    raw=json.loads((cap/'chapters.json').read_text());assert {c['name'] for c in raw}==set(durations)
    sources.append(dict(section=section,recording_sha256=sha(cap/'live.mp4'),receipt_sha256=sha(cap/'receipt.json'),chapters_sha256=sha(cap/'chapters.json')))
    chapters += [dict(c,section=section,capture=cap,seconds=durations[c['name']]) for c in raw]
segments=[];timeline=[];transitions=[];now=0
for c in [dict(name='intro',seconds=5),*chapters,dict(name='end',seconds=6)]:
    name=c['name'];duration=c['seconds'];dest=work/(name+'.mp4')
    if name in ('intro','end'):
        path=card(name,a.feedback/'created-menu.png' if name=='intro' else a.assist/'hd-quick-controls.png')
        run(['ffmpeg','-y','-v','error','-loop','1','-i',path,'-t',duration,'-r',FPS,'-c:v','libx264','-preset','veryfast','-crf','18','-threads','2','-pix_fmt','yuv420p',dest])
        caption='修正與新增功能實錄' if name=='intro' else '三平台、純EGA與EGA＋HD兩類完整版'
    else:
        caption=c['caption'];layer=work/(name+'-overlay.png');overlay(c).save(layer)
        factor=duration/c['duration']
        filters=(f'[0:v]trim=duration={c["duration"]},setpts=(PTS-STARTPTS)*{factor:.9f},scale=-2:900:flags=neighbor,'
                 f'pad=1920:1080:(ow-iw)/2:30:color=0x201414,fps={FPS},tpad=stop_mode=clone:stop_duration=1[v];'
                 '[v][1:v]overlay=0:0,format=yuv420p[out]')
        run(['ffmpeg','-y','-v','error','-ss',c['start'],'-i',c['capture']/'live.mp4','-loop','1','-i',layer,
             '-filter_complex_threads','1','-filter_complex',filters,'-map','[out]','-frames:v',duration*FPS,
             '-an','-c:v','libx264','-preset','veryfast','-crf','18','-threads','2',dest],work/(name+'-encode.log'))
        # 在實際遊戲區內檢查黑幀，不能靠外框或字幕掩蓋關窗。
        width=1440 if c['section']=='v1.0.6' else 1240
        crop=f'crop={width-20}:880:{(W-width)//2+10}:40'
        analysis=work/(name+'-black.log')
        run(['ffmpeg','-v','info','-i',dest,'-vf',crop+',blackdetect=d=0.1:pix_th=0.02:pic_th=0.98','-an','-f','null','-'],analysis)
        black=[tuple(map(float,m)) for m in re.findall(r'black_start:([0-9.]+) black_end:([0-9.]+) black_duration:([0-9.]+)',analysis.read_text())]
        if black:
            assert len(black)==1 and black[0][1]>=duration-.08 and black[0][2]<1,(name,black)
            original=work/(name+'-close-tail.mp4');dest.rename(original);cut=max(0,black[0][0]-1/FPS)
            run(['ffmpeg','-y','-v','error','-i',original,'-vf',f'trim=end={cut},setpts=PTS-STARTPTS,tpad=stop_mode=clone:stop_duration=2',
                 '-frames:v',duration*FPS,'-an','-c:v','libx264','-preset','veryfast','-crf','18','-threads','2',dest])
            transitions.append(dict(name=name,removed_black_tail=black[0],last_frame_held=duration-cut))
    segments.append(dest)
    timeline.append(dict(name=name,caption=caption,start=now,seconds=duration,section=c.get('section','card'),controlled=c.get('controlled',False),source_start=c.get('start'),source_duration=c.get('duration'),sha256=sha(dest)))
    now+=duration;print('ENCODED',name,duration,flush=True)
concat=work/'concat.txt';concat.write_text(''.join(f"file '{s}'\n" for s in segments))
silent=work/'silent.mp4';run(['ffmpeg','-y','-v','error','-f','concat','-safe','0','-i',concat,'-c','copy',silent])
music_path=root/'assets/audio/hd/manifest.json';music=json.loads(music_path.read_text());track=next(t for t in music['tracks'] if t['id']=='title')
audio=music_path.parent/track['file'];assert sha(audio)==track['sha256']
target=promo/f'kol-{version}-updates-promo-1080p.mp4';assert not target.exists()
run(['ffmpeg','-y','-v','error','-i',silent,'-stream_loop','-1','-i',audio,'-filter_complex',
     f'[1:a]atrim=0:{now},afade=t=in:d=1.2,afade=t=out:st={now-2}:d=2,volume=0.75[a]',
     '-map','0:v','-map','[a]','-c:v','copy','-c:a','aac','-b:a','192k','-t',now,'-movflags','+faststart',target])
probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',target]))
video=next(s for s in probe['streams'] if s['codec_type']=='video');sound=next(s for s in probe['streams'] if s['codec_type']=='audio')
assert (video['width'],video['height'],video['r_frame_rate'])==(W,H,'25/1')
assert abs(float(probe['format']['duration'])-now)<.15 and sound['sample_rate']=='44100'
pcm=run(['ffmpeg','-v','error','-i',target,'-map','0:a','-f','s16le','-ac','2','pipe:1'])
samples=np.frombuffer(pcm,dtype='<i2').astype(np.int32);peak=int(np.abs(samples).max());rms=float(np.sqrt(np.mean(samples.astype(float)**2)))
assert 100<peak<32767 and rms>20
run(['ffmpeg','-v','info','-i',target,'-vf','crop=1220:880:350:40,blackdetect=d=0.15:pix_th=0.02:pic_th=0.98,freezedetect=n=-70dB:d=4','-an','-f','null','-'],promo/'frame-analysis.log')
assert 'black_start:' not in (promo/'frame-analysis.log').read_text()
for c in timeline:
    frames=[]
    for n,fraction in enumerate((.08,.3,.5,.7,.92)):
        frame=promo/f'{c["name"]}-{n}.png'
        run(['ffmpeg','-y','-v','error','-ss',c['start']+c['seconds']*fraction,'-i',target,'-frames:v','1','-update','1',frame])
        frames.append(Image.open(frame).convert('RGB').crop((350,40,1570,920)).tobytes())
    if c['section']!='card':assert len(set(frames))>1,('操作章節沒有畫面變化',c['name'])
sheet=Image.new('RGB',(1920,math.ceil(len(timeline)/4)*270),(32,20,20))
for i,c in enumerate(timeline):sheet.paste(Image.open(promo/f'{c["name"]}-2.png').resize((480,270)),((i%4)*480,(i//4)*270))
sheet.save(promo/'contact-sheet.png');write(promo/'chapters.json',timeline);write(promo/'ffprobe.json',probe)
write(promo/'metadata.json',dict(status='PASS',version=version,video=target.name,video_sha256=sha(target),seconds=now,fps=FPS,resolution=[W,H],
    source_recordings=sources,binary_sha256=inputs['binaries']['linux-x86_64']['sha256'],audio=dict(source=str(audio.relative_to(root)),sha256=sha(audio),rights=music['rights'],peak=peak,rms=rms,human_listening=track['human_listening'],relistened_this_run=False),
    original_parity_claimed=False,remake_capture=True,controlled_scene_labels=True,window_close_tails_removed=transitions,
    font=dict(source='jf-openhuninn-2.1',sha256=sha(a.font),license='OFL-1.1'),subtitle_bounds_checked=True,
    differences_from_previous=['v1.0.6修正與v1.0.7新增分段標示','保留完整工具列與遊戲視口等比呈現','新增永久強化、住宿重載與輔助啟動狀態實錄'],
    limits='靜態字卡、暫停與閱讀頁會觸發凍結偵測；各操作章節五張抽樣及操作紀錄另核對。平台真機範圍未擴大。',tool_sha256=sha(Path(__file__))))
print('PROMO PASS',target,now,'seconds',flush=True)
