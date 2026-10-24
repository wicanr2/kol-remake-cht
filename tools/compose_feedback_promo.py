#!/usr/bin/env python3
"""Docker限定：把修正版的實際操作錄影剪成1080p推廣片，保存章節及影音檢查。"""
import argparse
import hashlib
import json
import math
import os
import re
import shutil
from pathlib import Path
import subprocess
from PIL import Image, ImageDraw
import numpy as np

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--root', type=Path, default=Path('/src'))
parser.add_argument('--out', type=Path, default=Path('/out'))
parser.add_argument('--capture', type=Path, required=True)
parser.add_argument('--work', type=Path, required=True)
parser.add_argument('--reuse-from', type=Path, help='重用本輪同一錄影的首輪片段，再完整核對影音')
args = parser.parse_args()
assert Path('/.dockerenv').exists()
ROOT, OUT, CAP, WORK = args.root, args.out, args.capture, args.work
P = OUT/'promo'
W, H, FPS = 1920, 1080, 25
version = json.loads((OUT/'build-inputs.json').read_text())['version']
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
write = lambda p,d: p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
chapters = json.loads((CAP/'chapters.json').read_text())
assert json.loads((CAP/'receipt.json').read_text())['status']=='PASS'
WORK.mkdir()
atlas_dir=ROOT/'assets/hd/first-slice/interface'
font=json.loads((atlas_dir/'font-atlas.json').read_text())
atlas=Image.open(atlas_dir/'font-atlas.png').convert('RGBA')
bitmap=json.loads((ROOT/'assets/ui-font/manifest.json').read_text())['fonts'][0]
glyphs={g['character']:g for g in bitmap['glyphs']}


def text(text, size, color=(244,232,200)):
    scale=size/font['size']
    width=math.ceil(sum(font['glyphs'][c]['advance']*scale if c in font['glyphs'] else size for c in text))+32
    canvas=Image.new('RGBA',(width,math.ceil(font['cell']*scale)+32))
    pen=16.
    for c in text:
        if c in font['glyphs']:
            g=font['glyphs'][c]; cell=font['cell']
            alpha=atlas.crop((g['x'],g['y'],g['x']+cell,g['y']+cell)).getchannel('A')
            side=round(cell*scale)
            alpha=alpha.resize((side,side),Image.Resampling.LANCZOS)
            x=round(pen-font['xpad']*scale)
            pen+=g['advance']*scale
        else:
            g=glyphs[c]
            alpha=Image.new('L',(16,18))
            pixels=alpha.load(); raw=bytes.fromhex(g['rows_hex'])
            for y in range(18):
                row=int.from_bytes(raw[2*y:2*y+2],'big')
                for x in range(16): pixels[x,y]=255 if row & (1<<(15-x)) else 0
            alpha=alpha.resize((size,size),Image.Resampling.NEAREST)
            x=round(pen);pen+=size
        tile=Image.new('RGBA',alpha.size,(*color,255));tile.putalpha(alpha)
        canvas.alpha_composite(tile,(x,16))
    box=canvas.getbbox()
    return canvas.crop(box) if box else Image.new('RGBA',(1,size))


def overlay(caption, controlled=False):
    layer=Image.new('RGBA',(W,H))
    draw=ImageDraw.Draw(layer)
    draw.rectangle((0,961,W,H),fill=(32,20,20,255))
    draw.line((120,963,W-120,963),fill=(200,153,46,255),width=2)
    label=text(caption,40)
    assert label.width<W-160 and label.height<70,(caption,label.size)
    layer.alpha_composite(label,((W-label.width)//2,982))
    sub=text('受控場景' if controlled else '實際鍵鼠操作',22,(186,172,145))
    layer.alpha_composite(sub,(80,1043))
    v=text(version,22,(186,172,145));layer.alpha_composite(v,(W-v.width-80,1043))
    return layer


def run(command, log=None, timeout=180):
    result=subprocess.run(list(map(str,command)),capture_output=True,timeout=timeout)
    if log: Path(log).write_bytes(result.stdout+result.stderr)
    if result.returncode: raise RuntimeError(result.stderr.decode(errors='replace')[-2000:])
    return result.stdout


def card(path, headline, subtitle, source):
    background=Image.new('RGBA',(W,H),(32,20,20,255))
    frame=Image.open(source).convert('RGBA').resize((1536,960),Image.Resampling.NEAREST)
    # 輸入畫面只等比縮放，文字卡的暗幕明確屬影片，不改遊戲素材。
    background.alpha_composite(frame,(192,0))
    background.alpha_composite(Image.new('RGBA',(W,H),(18,8,8,205)))
    title=text(headline,104,(224,180,82));sub=text(subtitle,44)
    background.alpha_composite(title,((W-title.width)//2,380))
    background.alpha_composite(sub,((W-sub.width)//2,550))
    background.convert('RGB').save(path)


durations=dict(name=6,save=6,resume=6,switch=6,sell=6,armour=8,magic=8,walk=5,ride=7,attack=9,hd=7)
assert set(durations)=={c['name'] for c in chapters}
segments=[];timeline=[];now=0.
transitions=[]
for index, chapter in enumerate([dict(name='intro',seconds=4),*chapters,dict(name='end',seconds=5)]):
    name=chapter['name']; dest=WORK/(name+'.mp4')
    if name in ('intro','end'):
        duration=chapter['seconds']; png=WORK/(name+'.png')
        card(png,'武士傳說', '操作修正實錄' if name=='intro' else '原影像版 / HD 版 · 三平台完整版',
             CAP/('created-menu.png' if name=='intro' else 'attack-area-hd-live.png'))
        if args.reuse_from:
            shutil.copy2(args.reuse_from/dest.name,dest)
        else:
            run(['ffmpeg','-y','-v','error','-loop','1','-i',png,'-t',duration,'-r',FPS,
                 '-c:v','libx264','-preset','veryfast','-crf','18','-threads','2','-pix_fmt','yuv420p',dest])
        caption='武士傳說：操作修正實錄' if name=='intro' else version+' 三平台完整版'
    else:
        duration=durations[name]; caption=chapter['caption'].replace('販賣','出售')
        paint=WORK/(name+'-overlay.png');overlay(caption,chapter['controlled']).save(paint)
        factor=duration/chapter['duration']
        filters=(f'[0:v]trim=duration={chapter["duration"]},setpts=(PTS-STARTPTS)*{factor:.9f},'
                 'scale=1536:960:flags=neighbor,pad=1920:1080:192:0:color=0x201414,'
                 f'fps={FPS},tpad=stop_mode=clone:stop_duration=1[v];[v][1:v]overlay=0:0,format=yuv420p[out]')
        if args.reuse_from:
            shutil.copy2(args.reuse_from/dest.name,dest)
        else:
            run(['ffmpeg','-y','-v','error','-ss',chapter['start'],'-i',CAP/'live.mp4','-loop','1','-i',paint,
                 '-filter_complex_threads','1','-filter_complex',filters,'-map','[out]',
                 '-frames:v',duration*FPS,'-an','-c:v','libx264','-preset','veryfast','-crf','18','-threads','2',dest],WORK/(name+'-encode.log'))
    segment_probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',dest]))
    assert abs(float(segment_probe['format']['duration'])-duration)<.05
    assert segment_probe['streams'][0]['width']==W and segment_probe['streams'][0]['height']==H
    check=WORK/(name+'-black.log')
    run(['ffmpeg','-v','info','-i',dest,'-vf','crop=1536:960:192:0,blackdetect=d=0.1:pix_th=0.02:pic_th=0.98','-an','-f','null','-'],check)
    black=[tuple(map(float,m)) for m in re.findall(r'black_start:([0-9.]+) black_end:([0-9.]+) black_duration:([0-9.]+)',check.read_text())]
    if black:
        assert len(black)==1 and black[0][1]>=duration-.08 and black[0][2]<1, (name,black)
        # 遊戲已關窗的片尾不是操作內容；保留最後實際畫面，補足章節既定時長。
        original=WORK/(name+'-window-close-tail.mp4');dest.rename(original)
        cutoff=max(0,black[0][0]-1/FPS)
        run(['ffmpeg','-y','-v','error','-i',original,'-vf',
             f'trim=end={cutoff},setpts=PTS-STARTPTS,tpad=stop_mode=clone:stop_duration=2',
             '-frames:v',duration*FPS,'-an','-c:v','libx264','-preset','veryfast','-crf','18','-threads','2',dest])
        transitions.append(dict(chapter=name,black_tail=black[0],last_game_frame_held_seconds=duration-cutoff,
                                source_sha256=sha(original)))
    segments.append(dest)
    timeline.append(dict(name=name,caption=caption,start=now,seconds=duration,
                         source=chapter if name not in ('intro','end') else 'title card',
                         sha256=sha(dest)))
    now+=duration
    print('ENCODED',name,duration,flush=True)
concat=WORK/'concat.txt';concat.write_text(''.join(f"file '{p}'\n" for p in segments))
silent=WORK/'silent.mp4'
run(['ffmpeg','-y','-v','error','-f','concat','-safe','0','-i',concat,'-c','copy',silent])
music_manifest=ROOT/'assets/audio/hd/manifest.json'
music=json.loads(music_manifest.read_text())
track=next(t for t in music['tracks'] if t['id']=='title')
audio=music_manifest.parent/track['file'];assert sha(audio)==track['sha256']
target=P/f'kol-{version}-feedback-promo-1080p.mp4'
assert not target.exists()
run(['ffmpeg','-y','-v','error','-i',silent,'-stream_loop','-1','-i',audio,
     '-filter_complex',f'[1:a]atrim=0:{now},afade=t=in:d=1.2,afade=t=out:st={now-2}:d=2,volume=0.75[a]',
     '-map','0:v','-map','[a]','-c:v','copy','-c:a','aac','-b:a','192k','-t',now,'-movflags','+faststart',target])
probe=json.loads(run(['ffprobe','-v','error','-show_streams','-show_format','-of','json',target]))
video=next(s for s in probe['streams'] if s['codec_type']=='video')
sound=next(s for s in probe['streams'] if s['codec_type']=='audio')
assert video['width']==W and video['height']==H and abs(float(probe['format']['duration'])-now)<.15
assert video['r_frame_rate']=='25/1' and sound['sample_rate']=='44100'
pcm=run(['ffmpeg','-v','error','-i',target,'-map','0:a','-f','s16le','-ac','2','pipe:1'])
samples=np.frombuffer(pcm,dtype='<i2').astype(np.int32)
peak=int(np.abs(samples).max());rms=float(np.sqrt(np.mean(samples.astype(float)**2)))
assert 100<peak<32767 and rms>20
run(['ffmpeg','-v','info','-i',target,'-vf','crop=1536:960:192:0,blackdetect=d=0.15:pix_th=0.02:pic_th=0.98,freezedetect=n=-70dB:d=4','-an','-f','null','-'],WORK/'frame-analysis.log')
assert 'black_start:' not in (WORK/'frame-analysis.log').read_text(), '遊戲畫面區出現黑幀'
# 逐段抽樣；靜態字卡預先登記，操作段必須有實際畫面變化。
frames=[]
for scene in timeline:
    for n, part in enumerate((.15,.5,.85)):
        t=scene['start']+scene['seconds']*part
        p=P/f'{scene["name"]}-{n}.png'
        run(['ffmpeg','-y','-v','error','-ss',t,'-i',target,'-frames:v','1','-update','1',p])
        frames.append((scene['name'],p))
    if scene['name'] not in ('intro','end'):
        images=[Image.open(p).convert('RGB').tobytes() for owner,p in frames if owner==scene['name']]
        assert len(set(images))>1,('操作段沒有畫面變化',scene['name'])
sheet=Image.new('RGB',(4*480,math.ceil(len(timeline)/4)*270),(32,20,20))
for i,scene in enumerate(timeline):
    p=P/f'{scene["name"]}-1.png'
    sheet.paste(Image.open(p).resize((480,270)),((i%4)*480,(i//4)*270))
sheet.save(P/'contact-sheet.png')
write(P/'ffprobe.json',probe)
write(P/'chapters.json',timeline)
shutil_log=P/'frame-analysis.log';shutil_log.write_bytes((WORK/'frame-analysis.log').read_bytes())
write(P/'metadata.json',dict(status='PASS',version=version,seconds=now,fps=FPS,resolution=[W,H],
    video=target.name,video_sha256=sha(target),source_recording_sha256=sha(CAP/'live.mp4'),
    source_receipt_sha256=sha(CAP/'receipt.json'),program=OUT.joinpath('build-inputs.json').as_posix(),
    audio=dict(source=str(audio.relative_to(ROOT)),sha256=sha(audio),rights=music['rights'],peak=peak,rms=rms,
               human_listening=track['human_listening'],relistened_this_run=False),
    remake_capture=True,controlled_scene_labels=True,original_parity_claimed=False,
    window_close_tails_removed=transitions,
    theme='原影像、紅磚暗底、金色字幕線；實際鍵鼠動作剪輯',
    differences_from_previous=['實錄操作取代靜態場景巡禮','11項操作章節與受控場景標示','展示存檔重啟及實際滑鼠選項'],
    limits='字卡與停留閱讀頁會觸發凍結偵測；逐段三個抽樣與操作紀錄另核對，平台真機未擴大宣稱。',
    tool_sha256=sha(Path(__file__))))
print('PROMO PASS',target,now,'seconds',flush=True)
