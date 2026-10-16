from pathlib import Path
from PIL import Image,ImageDraw,ImageChops
import hashlib,json,os
R,W=Path('/src'),Path('/out');D=R/'assets/hd/first-slice/town-maps'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
world_only=os.environ.get('KOL_MAP_WORLD_ONLY')=='1'
capture_file='maps-world-captures.json'if world_only else 'maps-captures.json'
runtime='maps-world-runtime'if world_only else 'maps-runtime'
m=json.loads((D/'manifest.json').read_text());c=json.loads((W/capture_file).read_text());expected_count=84 if world_only else 504;assert c['scenes']==expected_count and len(c['frames'])==expected_count
font=json.loads((R/'assets/hd/first-slice/interface/font-atlas.json').read_text());glyphs=set(font['glyphs'])
marker=Image.open(R/'assets/original-derived-png/ega/KOLPICS-MARKERS/kolpics-markers.png').convert('RGBA').crop((0,16,16,32));overlay=[]
for y in range(16):
 for x in range(16):
  lo,hi=(0,7)if x<8 else(8,15);occupied=lambda px:any(marker.getpixel((px,y))[:3])
  if occupied(x)or(x>lo and occupied(x-1))or(x<hi and occupied(x+1)):overlay.append((x,y,marker.getpixel((x,y))if occupied(x)else(0,0,0,255)))
records=[];counts={};max_delta=0
for archive in (['TOWN0']if world_only else ['TOWN3','TOWN4','TOWN5','TOWN6','TOWN7','TOWN0']):
 paints={};owners={};used=set();covered=set();rs=[r for r in m['regions']if r['archive']==archive];assert len(rs)==42
 for r in rs:
  p=D/r['file'];assert sha(p)==m['files'][r['file']]['sha256'];im=Image.open(p).convert('RGBA');assert im.size==(768,768)
  paints[r['id']]=im.transform((1152,1152),Image.Transform.AFFINE,(im.width/1152,0,0,0,im.height/1152,0),Image.Resampling.BILINEAR)
  x0,y0,x1,y1=r['visible_bounds'];sx,sy,_,_=r['source_bounds']
  for y in range(y0,y1):
   for x in range(x0,x1):assert(x,y)not in owners;owners[x,y]=(r['id'],x-sx,y-sy)
 assert len(owners)==4096
 frames=[f for f in c['frames']if f['Town']==int(archive[-1])];assert len(frames)==84
 for f in frames:
  assert f['Han']>0 and f['OutsideMax']==0 and f['RoundtripMax']<=1
  chars={ch for s in f['Text']for ch in s if '\u3400'<=ch<='\u9fff'};assert not chars-glyphs
  p=W/runtime/f['ViewportPNG'];assert sha(p)==f['ViewportSHA'];actual=Image.open(p).convert('RGBA');assert actual.size==(1152,576)
  expected=Image.new('RGBA',(1152,576))
  for row in range(6):
   for col in range(12):
    x,y=max(0,min(63,f['X']-5+col)),max(0,min(63,f['Y']-2+row));covered.add((x,y));key,sx,sy=owners[x,y];used.add(key);expected.paste(paints[key].crop((sx*96,sy*96,(sx+1)*96,(sy+1)*96)),(col*96,row*96))
  if archive!='TOWN0':
   d=ImageDraw.Draw(expected)
   for x,y,color in overlay:d.rectangle((5*96+x*6,2*96+y*6,5*96+x*6+5,2*96+y*6+5),fill=color)
  if archive=='TOWN0' and world_only:
   point=(516,324,528,336);expected.paste(ImageChops.invert(expected.crop(point).convert('RGB')).convert('RGBA'),point)
  delta=ImageChops.difference(actual,expected)
  if archive=='TOWN0' and not world_only:
   # World marker remains a separate EGA slice; verify its position/state through Go, exclude only its12x12 image rectangle.
   delta.paste((0,0,0,0),(516,324,528,336))
  maximum=max(e[1]for e in delta.getextrema());assert maximum<=1,(f['Name'],maximum);max_delta=max(max_delta,maximum)
  records.append(dict(case=f['Name'],source_cells=72,delivery_composite_max=maximum,outside_viewport_max=f['OutsideMax'],HD_roundtrip_max=f['RoundtripMax'],Chinese_present=True,missing_glyphs=0))
 assert len(covered)==4096 and len(used)==42
 counts[archive]=dict(cameras=84,source_cells=len(covered),regions=len(used))
previous=json.loads((W/'previous-manifest.json').read_text());current={r['id']:r for r in m['regions']}
for r in previous['regions']:assert r==current[r['id']] and sha(D/r['file'])==previous['files'][r['file']]['sha256']
source_file='maps-world-source-inputs.json'if world_only else 'maps-source-inputs.json'
inputs=json.loads((W/source_file).read_text());assert all(sha(R/p)==s for p,s in inputs.items())
result=dict(status='PASS controlled world background and native RGB point'if world_only else 'PASS controlled five-town and world position backgrounds',scenes=expected_count,regions=42 if world_only else 252,maps=counts,observed_Han_characters=sum(f['Han']for f in c['frames']),missing_glyphs=0,delivery_composite_max=max_delta,outside_viewport_max=0,HD_roundtrip_max=max(f['RoundtripMax']for f in c['frames']),previous84_native_bytes_and_bindings_unchanged=True,source_count=len(inputs),source_index_sha256=sha(W/source_file),captures_sha256=sha(W/capture_file),gallery_sha256=sha(W/('bin/maps-world'if world_only else 'bin/maps-gallery')),manifest_sha256=sha(D/'manifest.json'),generation_sha256=sha(D/'generation.json'),records=records,original_oracle=False,normal_player_path=False,GUI_F10=False,scope='World native RGB point included in full viewport composition; prepared scenes do not prove original parity.'if world_only else 'Position backgrounds only; original walking/world markers and weather not counted as HD; independent Pillow composition ignores only world marker12x12 rectangle; prepared scenes do not prove original parity.')
(W/('maps-world-check.json'if world_only else 'maps-check.json')).write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items()if k!='records'},ensure_ascii=False))
