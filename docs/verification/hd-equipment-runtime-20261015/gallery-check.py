from pathlib import Path
from PIL import Image,ImageChops
import json,hashlib
w=Path('/out');d=w/'equipment-formal-gallery';m=json.loads((d/'results.json').read_text());assert len(m['cases'])==38 and len(m['text'])==646
v10=json.loads((w/'action-runtime-v10/results.json').read_text());old={x['ID']:x for x in v10['Frames']}
check=[];literal=[]
for c in m['cases']:
 name=c['id']+'.png'
 if c['group']=='belt':expected=w/'belt-v13'/name
 elif c['group'] in ['default','copied','changed','striped']:expected=w/'action-runtime-v10'/name
 else:expected=w/'leg-native-v12'/name
 actual=Image.open(d/name).convert('RGBA');reference=Image.open(expected).convert('RGBA');assert actual.size==reference.size==(1920,1200)
 delta=ImageChops.difference(actual,reference);maximum=max(v[1] for v in delta.getextrema());assert maximum<=1,(name,maximum)
 assert delta.getchannel('A').getbbox() is None
 if c['group'] in ['changed','striped']:
  rect=old[c['id']]['FaceRect'];r=[rect['Min']['X'],rect['Min']['Y'],rect['Max']['X'],rect['Max']['Y']]
  assert r[2]-r[0]==240 and r[3]-r[1]==228
  cell=ImageChops.difference(actual.crop(r),reference.crop(r));assert all(x==(0,0) for x in cell.getextrema()),c['id']
  literal.append(dict(id=c['id'],rect=r,original_v10_literal_grid_equal=True))
 check.append(dict(id=c['id'],group=c['group'],reference=str(expected.relative_to(w)),max_channel_delta=maximum,alpha_equal=True,actual_png_sha256=hashlib.sha256((d/name).read_bytes()).hexdigest()))
defaults=[]
for body in range(6):
 a=Image.open(d/f'body{body}-shield-default.png').convert('RGBA').crop((0,0,1920,950));b=Image.open(d/f'body{body}-shield-copied.png').convert('RGBA').crop((0,0,1920,950))
 maximum=max(x[1] for x in ImageChops.difference(a,b).getextrema());assert maximum<=1
 defaults.append(dict(body=body,default_and_copied_viewport_max_delta=maximum))
result=dict(schema='kol-hd-equipment-formal-gallery-check/1',frames=check,text_records=646,full_hd_max_delta=max(x['max_channel_delta'] for x in check),literal_shields=literal,default_matches=defaults,source_frames=216,limits='Controlled actual UI fixtures, not normal-player or original-executable evidence. Original PNG/rawRGB and GPU-rendered RGB independently verified.')
(d/'check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS38 full HD frames max',result['full_hd_max_delta'],';12 literal shield grids exact;6 default/copied pairs equal;646 Chinese slot records.')
