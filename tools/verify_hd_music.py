#!/usr/bin/env python3
"""最終OGG及可用分軌組合驗證。技術PASS不等於人耳通過。"""
import hashlib,itertools,json,subprocess
from pathlib import Path
import numpy as np
assert Path('/.dockerenv').exists()
ROOT=Path('/src');OUT=Path('/out');DEST=ROOT/'docs/verification/hd-music-20261012'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def decode(path):return np.frombuffer(subprocess.check_output(['ffmpeg','-v','error','-i',str(path),'-f','f32le','-acodec','pcm_f32le','-'],timeout=60),dtype='<f4').reshape(-1,2)
checks=[]
for kind,seconds in [('title',72),('inn',80)]:
    stems=[decode(DEST/(kind+'-'+part+'.ogg')) for part in ('lead','harmony','pulse')]
    assert all(len(s)==seconds*44100 for s in stems)
    for bits in itertools.product((0,1),repeat=3):
        if not any(bits):continue
        mixed=sum(s for b,s in zip(bits,stems) if b)
        peak=float(np.max(np.abs(mixed)));assert peak<.99
        checks.append(dict(track=kind,stems=bits,peak=peak,seconds=seconds,decoded_ogg=True))
    triple=decode(DEST/(kind+'-three-loops.ogg'));assert len(triple)==seconds*44100*3
    assert 0<float(np.sqrt(np.mean(triple**2))) and np.max(np.abs(triple))<.99
    del triple
preview=DEST/'title-with-thunder.ogg'
subprocess.run(['ffmpeg','-y','-v','error','-i',str(DEST/'title.ogg'),'-i',str(ROOT/'assets/audio/title-thunder.ogg'),'-filter_complex','[1:a]adelay=3000:all=1[sfx];[0:a][sfx]amix=inputs=2:duration=first:normalize=0','-c:a','libvorbis','-q:a','6',str(preview)],check=True,timeout=60)
samples=decode(preview);assert len(samples)==72*44100 and np.max(np.abs(samples))<.99
report=dict(schema='kol-hd-music-qa/1',status='TECHNICAL_PASS_LISTENING_PENDING',checks=checks,three_loops_seconds={'title':216,'inn':240},sfx_preview=dict(file=preview.name,sha256=sha(preview),peak=float(np.max(np.abs(samples))),source='assets/audio/title-thunder.ogg',source_sha256=sha(ROOT/'assets/audio/title-thunder.ogg'),delay_seconds=3),listening=dict(identity_blind_test='pending',third_loop_fatigue='pending',dialogue_and_sfx_audibility='pending'),production_enabled=False)
(DEST/'technical-check.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('MUSIC QA PASS',len(checks),'decoded stem combinations, both three-loop files, thunder mix',flush=True)
