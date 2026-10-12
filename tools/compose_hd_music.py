#!/usr/bin/env python3
"""原創室內樂試聽。只在Docker執行；標準MIDI，不引用原作旋律或音訊。"""
import argparse, hashlib, itertools, json, math, struct, subprocess, wave
from pathlib import Path
import numpy as np

assert Path('/.dockerenv').exists()
TPQ=480; RATE=44100
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def vlq(n):
    out=[n&127];n>>=7
    while n:out.insert(0,(n&127)|128);n>>=7
    return bytes(out)
def midi(path,notes,bpm,bars):
    events=[(0,0,b'\xff\x51\x03'+int(60000000/bpm).to_bytes(3,'big')),(0,1,b'\xff\x58\x04\x04\x02\x18\x08')]
    for channel,program,pan in [(0,73,76),(1,41,48),(2,42,60),(3,24,84),(9,0,64)]:
        events.extend([(0,2,bytes([0xc0|channel,program])),(0,2,bytes([0xb0|channel,10,pan])),(0,2,bytes([0xb0|channel,91,36])),(0,2,bytes([0xb0|channel,93,0]))])
    for ch,pitch,start,length,velocity in notes:
        events.extend([(round(start*TPQ),4,bytes([0x90|ch,pitch,velocity])),(round((start+length)*TPQ),3,bytes([0x80|ch,pitch,0]))])
    events.append((bars*4*TPQ,9,b'\xff\x2f\x00'));events.sort(key=lambda e:(e[0],e[1]));last=0;track=b''
    for tick,_,data in events:track+=vlq(tick-last)+data;last=tick
    path.write_bytes(b'MThd'+struct.pack('>IHHH',6,0,1,TPQ)+b'MTrk'+struct.pack('>I',len(track))+track)

def score(kind):
    title=kind=='title';bpm=80 if title else 72;bars=24
    # D Aeolian，借用E作柔和懸念。三個八小節段落，每段呼吸四小節。
    chords=[(50,57,65),(48,55,64),(46,53,62),(48,55,64),(50,57,65),(53,60,69),(43,50,58),(45,52,60)]*3
    notes=[]
    def add(ch,p,beat,dur,v):notes.append((ch,p,beat,dur,v))
    for bar,chord in enumerate(chords):
        beat=bar*4
        add(2,chord[0]-12,beat+.04,3.65,43 if title else 36)
        # 室內樂弓弦每兩小節交替留白；不鋪整片交響弦樂。
        if bar%4!=3:
            add(1,chord[1],beat+.12,2.65,39 if title else 31)
            if title and bar>=8:add(1,chord[2],beat+2.75,.9,30)
        for offset,index in ([(.1,1),(2.1,2)] if title else [(.1,1),(1.6,2),(3.1,1)]):
            add(3,chord[index],beat+offset,.7,35 if title else 30)
        # 手鼓只在中段出現，每兩小節一次；保留對話及確認音的空間。
        if 8<=bar<16 and bar%2==0:
            add(9,63,beat+.02,.16,28);add(9,62,beat+2.6,.12,19)
    motif=[74,81,72,70,76,74] # D5 A5 C5 Bb4 E5 D5，新作，不是原版旋律。
    rhythms=[(0,1.5),(2,1),(4,1.5),(6,1),(8,1.5),(10,2)]
    for phrase,start in enumerate([0,8,16]):
        if title: pitches=motif if phrase!=1 else [77,81,79,76,72,74]
        else:pitches=[p-12 for p in motif] if phrase!=1 else [69,72,74,77,76,72]
        for p,(offset,dur) in zip(pitches,rhythms):
            add(0,p,start*4+2+offset,dur,52 if title else 43)
        # 短回應與停頓，末段不以強終止式卡住循環。
        if phrase<2:
            for j,p in enumerate(([69,72,74] if title else [65,64,62])):
                add(1,p,start*4+19+j*1.5,1.1,38 if title else 31)
    return bpm,bars,notes

def run(argv):return subprocess.run(list(map(str,argv)),check=True,capture_output=True,text=True,timeout=120)
def main():
    ap=argparse.ArgumentParser();ap.add_argument('--out',type=Path,required=True);a=ap.parse_args();a.out.mkdir(exist_ok=True)
    font=Path('/usr/share/sounds/sf2/FluidR3_GM.sf2');assert font.is_file()
    config=a.out/'timidity.cfg';config.write_text('soundfont '+str(font)+'\n')
    licensepath=Path('/usr/share/doc/fluid-soundfont-gm/copyright');(a.out/'soundfont-LICENSE.txt').write_bytes(licensepath.read_bytes())
    checks=[];catalog={}
    for kind in ('title','inn'):
        bpm,bars,notes=score(kind);duration=bars*4*60/bpm;frames=round(duration*RATE);stems=[]
        midi(a.out/(kind+'.mid'),notes,bpm,bars)
        scoredata=dict(label='重製版原創室內樂 MIDI 試聽稿',bpm=bpm,bars=bars,time_signature='4/4',motif=['D5','A5','C5','Bb4','E5','D5'],notes=[dict(channel=ch,pitch=p,start_beat=s,duration_beats=d,velocity=v) for ch,p,s,d,v in notes])
        (a.out/(kind+'-score.json')).write_text(json.dumps(scoredata,ensure_ascii=False,indent=2)+'\n')
        for name,channels in [('lead',{0}),('harmony',{1,2}),('pulse',{3,9})]:
            stemnotes=[n for n in notes if n[0] in channels];mid=a.out/(kind+'-'+name+'.mid');midi(mid,stemnotes,bpm,bars)
            raw=a.out/(kind+'-'+name+'-raw.wav')
            run(['timidity','-c',config,'-OwS','-s',str(RATE),'-o',raw,mid])
            with wave.open(str(raw),'rb') as w:
                assert w.getnchannels()==2 and w.getsampwidth()==2 and w.getframerate()==RATE
                samples=np.frombuffer(w.readframes(w.getnframes()),dtype='<i2').reshape(-1,2).astype(np.float64)/32768
            # 把尾韻折回開頭，整首與各聲部都使用相同小節長度。
            folded=np.zeros((frames,2),dtype=np.float64)
            for start in range(0,len(samples),frames):
                part=samples[start:start+frames];folded[:len(part)]+=part
            # 10ms的極輕接點窗，只作用於循環邊界，防止起迄不連續。
            fade=round(.01*RATE);folded[:fade]*=np.linspace(0,1,fade)[:,None];folded[-fade:]*=np.linspace(1,0,fade)[:,None]
            stems.append((name,folded));raw.unlink();mid.unlink()
        mix=sum(s for _,s in stems);rms=math.sqrt(float(np.mean(mix**2)));gain=min(.70/max(float(np.max(np.abs(mix))),1e-8),10**(-24/20)/rms)
        for name,samples in [('mix',mix),*stems]:
            samples=samples*gain;wav=a.out/(kind+'-'+name+'.wav')
            with wave.open(str(wav),'wb') as w:w.setnchannels(2);w.setsampwidth(2);w.setframerate(RATE);w.writeframes(np.round(np.clip(samples,-1,1)*32767).astype('<i2').tobytes())
            ogg=a.out/(kind+('.ogg' if name=='mix' else '-'+name+'.ogg'))
            run(['ffmpeg','-y','-v','error','-i',wav,'-c:a','libvorbis','-q:a','6',ogg]);wav.unlink()
            # 從最終OGG解碼驗，而非只驗編碼前WAV。
            decoded=subprocess.check_output(['ffmpeg','-v','error','-i',str(ogg),'-f','f32le','-acodec','pcm_f32le','-'],timeout=60)
            audio=np.frombuffer(decoded,dtype='<f4').reshape(-1,2)
            report=dict(path=ogg.name,sha256=sha(ogg),duration=len(audio)/RATE,sample_rate=RATE,channels=2,peak=float(np.max(np.abs(audio))),rms=float(np.sqrt(np.mean(audio**2))),boundary_delta=float(np.max(np.abs(audio[0]-audio[-1]))))
            assert abs(report['duration']-duration)<.01 and 0<report['rms'] and report['peak']<.99
            checks.append(report)
        for bits in itertools.product((0,1),repeat=3):
            if not any(bits):continue
            combination=sum(s for bit,(_,s) in zip(bits,stems) if bit)*gain
            assert np.max(np.abs(combination))<.99
        ebur=run(['ffmpeg','-nostats','-i',a.out/(kind+'.ogg'),'-filter_complex','ebur128=peak=true','-f','null','-'])
        (a.out/(kind+'-loudness.log')).write_text(ebur.stderr)
        triple=a.out/(kind+'-three-loops.ogg')
        run(['ffmpeg','-y','-v','error','-i',a.out/(kind+'.ogg'),'-af',f'aloop=loop=2:size={frames}', '-t',str(duration*3),'-c:a','libvorbis','-q:a','6',triple])
        catalog[kind]=dict(bpm=bpm,bars=bars,duration=duration,midi_sha256=sha(a.out/(kind+'.mid')),allowed_stem_combinations='all seven nonempty subsets; same gain/grid; verified unclipped')
    report=dict(schema='kol-hd-music-candidate/1',status='TECHNICAL_PASS_LISTENING_PENDING',music_identity='new-remake-original',source_melody='project-authored MIDI; no original melody sampled',soundfont=dict(path=str(font),sha256=sha(font),package='fluid-soundfont-gm 3.1-5.3',license='MIT; see soundfont-LICENSE.txt'),synth='TiMidity++ 2.14.0-8.1',codec='Ogg Vorbis q6; 44100Hz stereo',tracks=catalog,checks=checks,production_enabled=False)
    (a.out/'manifest.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('MUSIC TECHNICAL PASS; listening pending',flush=True)
if __name__=='__main__':main()
