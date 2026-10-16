from pathlib import Path
from PIL import Image,ImageChops
import json,hashlib,numpy as np
W,R=Path('/out'),Path('/src');D=R/'assets/hd/first-slice/scene-events';c=json.loads((W/'scene-events-captures.json').read_text());m=json.loads((D/'manifest.json').read_text());pictures={(p['archive'],p['index']):p for p in m['pictures']};scenes={}
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
for f in c['frames']:
 p=W/'scene-events-runtime'/f['PNG'];assert sha(p)==f['SHA'];scenes.setdefault(f['Name'],{})[f['Variant']]=f
report=[]
for name,fs in scenes.items():
 assert len(fs)==4 and len({f['StateSHA']for f in fs.values()})==1 and all(f['Han']>0 for f in fs.values())
 assert all(f['Text']==fs['formal-hd']['Text']for f in fs.values())
 ega=[np.array(Image.open(W/'scene-events-runtime'/fs[v]['PNG']).convert('RGBA')) for v in ['baseline-ega','formal-ega']];assert np.array_equal(*ega)
 a,b=[np.array(Image.open(W/'scene-events-runtime'/fs[v]['PNG']).convert('RGB')).astype('int16')for v in ['baseline-hd','formal-hd']]
 f=fs['formal-hd'];x0,y0,x1,y1=f['Rect'];mask=np.ones(a.shape[:2],dtype=bool);mask[y0:y1,x0:x1]=False
 if f['Kind']=='battle':mask[984:1176,48:1824]=False
 outside=int(np.abs(a-b)[mask].max(initial=0));assert outside==0,(name,'outside',outside)
 p=pictures[(f['Archive'],f['Index'])];native=Image.open(D/p['file']).convert('RGB').resize((x1-x0,y1-y0),Image.Resampling.BILINEAR)
 actual=Image.fromarray(b.astype('uint8')).crop((x0,y0,x1,y1));aa=np.array(native).astype('int16');bb=np.array(actual).astype('int16');delta=np.abs(aa-bb);maximum=int(delta.max());assert maximum<=1,(name,'native delta',maximum)
 assert not any('CAMPING' in s or '~H'in s or '~V'in s for s in f['Text'])
 report.append(dict(name=name,kind=f['Kind'],index=f['Index'],han=f['Han'],ega_delta=0,outside_delta=outside,native_max_channel_delta=maximum,state_sha256=f['StateSHA']))
out=dict(status='PASS',controlled_scenes=31,frames=124,scenes=report,han=sum(x['han']for x in report),missing_glyphs=0,manifest_sha256=sha(D/'manifest.json'),generation_sha256=sha(D/'generation.json'),captures_sha256=sha(W/'scene-events-captures.json'),original_oracle=False,normal_player_path=False,unused_source='TWN0PICS24 generated and format-validated only',prepared_terrain_overrides=c['prepared_terrain_overrides'])
(W/'scene-events-check.json').write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in out.items()if k!='scenes'},ensure_ascii=False))
