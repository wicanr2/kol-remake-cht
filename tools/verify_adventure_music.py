#!/usr/bin/env python3
"""驗最終探索／戰鬥OGG、分軌組合與三輪輸出，保留既有人耳決定。"""
import hashlib,itertools,json,subprocess,gzip
from pathlib import Path
import numpy as np
assert Path('/.dockerenv').exists()
d=Path('/src/docs/verification/hd-adventure-music-20261013');checks=[]
def decode(p):
    r=subprocess.run(['ffmpeg','-v','error','-i',str(p),'-f','f32le','-acodec','pcm_f32le','-'],capture_output=True,check=True,timeout=60)
    assert not r.stderr,r.stderr.decode();return np.frombuffer(r.stdout,dtype='<f4').reshape(-1,2)
for track,seconds in [('exploration',90),('combat',60)]:
    stems=[decode(d/(track+'-'+name+'.ogg')) for name in ('lead','harmony','pulse')]
    assert all(len(s)==seconds*44100 for s in stems)
    for bits in itertools.product((0,1),repeat=3):
        if not any(bits):continue
        mix=sum(s for bit,s in zip(bits,stems) if bit);peak=float(np.max(np.abs(mix)));assert peak<.99
        checks.append(dict(track=track,stems=bits,peak=peak,loop_boundary_delta=float(np.max(np.abs(mix[0]-mix[-1])))))
    triple=decode(d/(track+'-three-loops.ogg'));assert len(triple)==seconds*44100*3 and 0<float(np.sqrt(np.mean(triple**2))) and np.max(np.abs(triple))<.99
    del triple
for p in d.glob('*.log'):
    (d/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0));p.unlink()
m=json.loads((d/'manifest.json').read_text());m.update(toolchain='kol-music:timidity2.14-fluid3.1-r1',technical_qa=dict(status='PASS',decoded_stem_combinations=checks,three_loops_pass=True,no_decoder_warnings=True),human_listening=m.get('human_listening','pending'),production_enabled=m.get('production_enabled',False),files={p.name:dict(sha256=hashlib.sha256(p.read_bytes()).hexdigest(),size=p.stat().st_size) for p in sorted(d.iterdir()) if p.name!='manifest.json'})
(d/'manifest.json').write_text(json.dumps(m,ensure_ascii=False,indent=2)+'\n');print('ADVENTURE QA PASS',len(checks),'decoded stem combinations; listening decision preserved')
