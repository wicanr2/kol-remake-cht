from pathlib import Path
from PIL import Image,ImageChops,ImageDraw
import hashlib,json
w=Path('/out');r=Path('/src');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
formal=json.loads((w/'formal-city-captures.json').read_text());old=json.loads((w/'position-city-captures.json').read_text());v2=json.loads((w/'position-city-v2/position-city-captures.json').read_text());m=json.loads((r/'assets/hd/first-slice/town-maps/manifest.json').read_text());paint={};owners={}
for p in m['regions']:
    native=r/'assets/hd/first-slice/town-maps'/p['file'];assert sha(native)==m['files'][p['file']]['sha256'];im=Image.open(native).convert('RGBA');paint[p['id']]=im.transform((1152,1152),Image.Transform.AFFINE,(im.width/1152,0,0,0,im.height/1152,0),Image.Resampling.BILINEAR);x0,y0,x1,y1=p['visible_bounds'];sx,sy,_,_=p['source_bounds']
    for y in range(y0,y1):
        for x in range(x0,x1):assert(x,y)not in owners;owners[x,y]=(p['id'],x-sx,y-sy)
assert len(owners)==4096 and len(paint)==42
marker=Image.open(r/'assets/original-derived-png/ega/KOLPICS-MARKERS/kolpics-markers.png').convert('RGBA').crop((0,16,16,32));overlay=[]
for y in range(16):
    for x in range(16):
        lo,hi=(0,7)if x<8 else(8,15);occupied=lambda px:any(marker.getpixel((px,y))[:3])
        if occupied(x)or(x>lo and occupied(x-1))or(x<hi and occupied(x+1)):overlay.append((x,y,marker.getpixel((x,y))if occupied(x)else(0,0,0,255)))
base={(f['Case'],f['Variant']):f for f in old['frames']};revised={(f['Case'],f['Variant']):f for f in v2['frames']};records=[]
names=sorted({f['Case']for f in formal['frames']});assert len(names)==23 and len(formal['frames'])==92
for name in names:
    fs={f['Variant']:f for f in formal['frames']if f['Case']==name};assert set(fs)=={'candidate-ega','candidate-hd','baseline-ega','baseline-hd'}
    for f in fs.values():
        ref=base[name,'baseline-ega'];assert f['StateSHA']==ref['StateSHA']and f['Text']==ref['Text']and f['Han']>0;assert sha(w/'formal-city-runtime'/f['PNG'])==f['SHA']
    ega=Image.open(w/'formal-city-runtime'/fs['candidate-ega']['PNG']).convert('RGBA');ega0=Image.open(w/'position-city-runtime'/base[name,'baseline-ega']['PNG']).convert('RGBA');ega1=Image.open(w/'formal-city-runtime'/fs['baseline-ega']['PNG']).convert('RGBA');assert ega.tobytes()==ega0.tobytes()==ega1.tobytes()
    hd=Image.open(w/'formal-city-runtime'/fs['candidate-hd']['PNG']).convert('RGBA');baseline=Image.open(w/'formal-city-runtime'/fs['baseline-hd']['PNG']).convert('RGBA');expected=baseline.copy()
    is_map=name.startswith('camera-')
    if is_map:
        _,cx,cy=name.split('-');cx,cy=int(cx),int(cy)
        for row in range(6):
            for col in range(12):
                x,y=max(0,min(63,cx-5+col)),max(0,min(63,cy-2+row));key,sx,sy=owners[x,y];expected.paste(paint[key].crop((sx*96,sy*96,(sx+1)*96,(sy+1)*96)),(624+col*96,348+row*96))
        d=ImageDraw.Draw(expected)
        for x,y,c in overlay:d.rectangle((624+5*96+x*6,348+2*96+y*6,624+5*96+x*6+5,348+2*96+y*6+5),fill=c)
    native_max=max(p[1]for p in ImageChops.difference(hd,expected).getextrema());assert native_max<=(1 if is_map else 0),(name,native_max)
    outside=ImageChops.difference(hd,baseline);outside.paste((0,0,0,0),(624,348,1776,924));assert max(p[1]for p in outside.getextrema())==0
    prior=revised.get((name,'candidate-hd'),base[name,'candidate-hd']);parent=w/'position-city-v2/position-city-runtime'if(name,'candidate-hd')in revised else w/'position-city-runtime';candidate=Image.open(parent/prior['PNG']).convert('RGBA');whole_max=max(p[1]for p in ImageChops.difference(hd,candidate).getextrema());assert whole_max<=1,(name,whole_max)
    records.append({'case':name,'EGA_RGBA_exact':True,'serializable_Outdoor_equal':True,'Chinese_equal':True,'native_composite_max':native_max,'whole_candidate_HD_max':whole_max,'outside_viewport_max':0,'comparator':'same execution and draw sequence; verification-only pointer suspension outside production'})
font=json.loads((r/'assets/hd/first-slice/interface/font-atlas.json').read_text());chars={c for f in formal['frames']for s in f['Text']for c in s if '\u3400'<=c<='\u9fff'};assert not chars-set(font['glyphs'])
sources=json.loads((w/'formal-city-source-inputs.json').read_text());assets=json.loads((w/'formal-city-asset-inputs.json').read_text());assert len(sources)==417 and len(assets)==436 and all(sha(r/p)==v for p,v in sources.items())and all(sha(r/p)==v for p,v in assets.items())
result={'status':'PASS limited formal first-city sample','scenes':len(names),'frames':len(formal['frames']),'observed_Han_characters':sum(f['Han']for f in formal['frames']),'missing_glyphs':0,'source_count':len(sources),'HD_asset_count':len(assets),'verification_source_count':len(json.loads((w/'formal-city-verification-inputs.json').read_text())),'records':records,'program_sha256':sha(w/'bin/formal-city'),'gallery_sha256':sha(w/'bin/formal-city-gallery'),'source_index_sha256':sha(w/'formal-city-source-inputs.json'),'asset_index_sha256':sha(w/'formal-city-asset-inputs.json'),'verification_index_sha256':sha(w/'formal-city-verification-inputs.json'),'captures_sha256':sha(w/'formal-city-captures.json'),'source_manifest_sha256':sha(r/'assets/hd/first-slice/town-maps/manifest.json'),'original_oracle':False,'normal_player_path':False,'limits':'Single formal map/corner/door/edge sample; prior complete4096-cell and changed-bridge receipts retained separately; marker and other towns not HD-complete.'}
(w/'formal-city-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items()if k!='records'}))
