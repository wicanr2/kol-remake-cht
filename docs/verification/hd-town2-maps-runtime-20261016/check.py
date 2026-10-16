from pathlib import Path
from PIL import Image, ImageChops, ImageDraw
import json, hashlib
R=Path('/src');W=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
manifest=json.loads((W/'candidate-assets/manifest.json').read_text());capture=json.loads((W/'formal-city-captures.json').read_text())
paint={};owners={}
for r in manifest['regions']:
 if r['archive']!='TOWN2':continue
 p=W/'candidate-assets'/r['file'];assert sha(p)==manifest['files'][r['file']]['sha256'];im=Image.open(p).convert('RGBA')
 paint[r['id']]=im.transform((1152,1152),Image.Transform.AFFINE,(im.width/1152,0,0,0,im.height/1152,0),Image.Resampling.BILINEAR)
 x0,y0,x1,y1=r['visible_bounds'];sx,sy,_,_=r['source_bounds']
 for y in range(y0,y1):
  for x in range(x0,x1):assert(x,y)not in owners;owners[x,y]=(r['id'],x-sx,y-sy)
assert len(owners)==4096 and len(paint)==42
marker=Image.open(R/'assets/original-derived-png/ega/KOLPICS-MARKERS/kolpics-markers.png').convert('RGBA').crop((0,16,16,32));overlay=[]
for y in range(16):
 for x in range(16):
  lo,hi=(0,7)if x<8 else(8,15);occupied=lambda px:any(marker.getpixel((px,y))[:3])
  if occupied(x)or(x>lo and occupied(x-1))or(x<hi and occupied(x+1)):overlay.append((x,y,marker.getpixel((x,y))if occupied(x)else(0,0,0,255)))
names=sorted({f['Case']for f in capture['frames']});assert len(names)==99 and len(capture['frames'])==396
records=[];covered=set()
for name in names:
 fs={f['Variant']:f for f in capture['frames']if f['Case']==name};assert set(fs)=={'baseline-ega','baseline-hd','candidate-ega','candidate-hd'}
 ref=fs['baseline-ega']
 for f in fs.values():assert f['StateSHA']==ref['StateSHA'] and f['Text']==ref['Text'] and f['Han']>0 and sha(W/'formal-city-runtime'/f['PNG'])==f['SHA']
 images={k:Image.open(W/'formal-city-runtime'/f['PNG']).convert('RGBA')for k,f in fs.items()}
 assert images['baseline-ega'].tobytes()==images['candidate-ega'].tobytes()
 hd=images['candidate-hd'];base=images['baseline-hd'];expected=base.copy();is_map=name.startswith('camera-')
 if is_map:
  _,cx,cy=name.split('-');cx,cy=int(cx),int(cy)
  for row in range(6):
   for col in range(12):
    x,y=max(0,min(63,cx-5+col)),max(0,min(63,cy-2+row));covered.add((x,y));key,sx,sy=owners[x,y];expected.paste(paint[key].crop((sx*96,sy*96,(sx+1)*96,(sy+1)*96)),(624+col*96,348+row*96))
  d=ImageDraw.Draw(expected)
  for x,y,c in overlay:d.rectangle((624+5*96+x*6,348+2*96+y*6,624+5*96+x*6+5,348+2*96+y*6+5),fill=c)
 delta=max(v[1]for v in ImageChops.difference(hd,expected).getextrema());assert delta<=(1 if is_map else 0),(name,delta)
 outside=ImageChops.difference(hd,base);outside.paste((0,0,0,0),(624,348,1776,924));assert max(v[1]for v in outside.getextrema())==0,name
 records.append(dict(case=name,EGA_RGBA_exact=True,Chinese_equal=True,serializable_Outdoor_equal=True,native_composite_max=delta,outside_viewport_max=0))
assert len(covered)==4096
prior_capture=json.loads((W/'preview-captures.json').read_text());old_frames={(f['Case'],f['Variant']):f for f in prior_capture['frames']}
whole_max=0
for f in capture['frames']:
 old=old_frames[f['Case'],f['Variant']];assert f['StateSHA']==old['StateSHA'] and f['Text']==old['Text']
 a=Image.open(W/'formal-city-runtime'/f['PNG']).convert('RGBA');b=Image.open(W/'preview-runtime'/old['PNG']).convert('RGBA');maximum=max(v[1]for v in ImageChops.difference(a,b).getextrema());assert maximum<=(1 if f['Variant'].endswith('hd')else 0),(f['Case'],maximum);whole_max=max(whole_max,maximum)

font=json.loads((R/'assets/hd/first-slice/interface/font-atlas.json').read_text());chars={c for f in capture['frames']for s in f['Text']for c in s if '\u3400'<=c<='\u9fff'};assert not chars-set(font['glyphs'])
prior=json.loads((W/'town1-original-manifest.json').read_text());current={r['id']:r for r in manifest['regions']}
assert manifest['maps'][0]==prior['maps'][0]
for r in prior['regions']:assert r==current[r['id']] and sha(W/'candidate-assets'/r['file'])==prior['files'][r['file']]['sha256']
inputs=json.loads((W/'formal-source-inputs.json').read_text());assert all(sha(R/p)==v for p,v in inputs.items())
result=dict(status='PASS limited second-city controlled complete map gallery',scenes=99,frames=396,map_cameras=84,source_cells=4096,regions=42,doors=14,observed_Han_characters=sum(f['Han']for f in capture['frames']),missing_glyphs=0,first_city_native_bytes_and_bindings_unchanged=True,source_count=len(inputs),source_index_sha256=sha(W/'formal-source-inputs.json'),asset_index_sha256=sha(W/'candidate-assets-inputs.json'),captures_sha256=sha(W/'formal-city-captures.json'),gallery_sha256=sha(W/'bin/town2-gallery'),records=records,original_oracle=False,normal_player_path=False,GUI_F10=False,whole_preview_max=whole_max,formal_asset_index_sha256=sha(W/'formal-asset-inputs.json'),scope='Prepared TOWN2 complete map, typed doors and edge, actual production rendering; original EGA walking marker retained; other HD remains open.')
(W/'formal-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items()if k!='records'},ensure_ascii=False))
