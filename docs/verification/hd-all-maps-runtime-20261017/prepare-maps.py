from pathlib import Path
from PIL import Image
import json,hashlib,shutil
R=Path('/src');W=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert W.stat().st_uid==W.stat().st_gid==1000
base=json.loads((R/'assets/hd/first-slice/town-maps/manifest.json').read_text());regions=[];requests=[];sources=[]
style='approved-town-style.png';shutil.copyfile(R/'assets/hd/first-slice/town-maps/town1-r01-c04.png',W/style)
waterstyle='approved-water-style.png';shutil.copyfile(R/'assets/hd/first-slice/town-maps/town2-r03-c03.png',W/waterstyle)
raw=json.loads((R/'assets/original-derived-png/ega/manifest.json').read_text())
for n in [3,4,5,6,7,0]:
 archive=f'TOWN{n}';entries=next(x for x in raw['archives']if x['archive']==archive)['images'];assert len(entries)==2
 pp=R/'assets/original-derived-png/ega'/archive;grid=Image.open(pp/f'town{n}-map.png');atlas=Image.open(pp/f'town{n}-tiles.png');assert grid.mode=='L' and grid.size==(64,64) and atlas.mode=='RGBA' and atlas.size==(256,256)
 source={'archive':archive,'archive_count':2}
 for index,label in [(0,'grid'),(1,'atlas')]:
  entry=next(x for x in entries if x['unit_index']==index);p=R/'assets/original-derived-png/ega'/entry['file'];im=Image.open(p);assert sha(p)==entry['png_sha256']
  source[label]={'archive':archive,'index':index,'source_png_sha256':sha(p),'source_rgb_sha256':hashlib.sha256(im.convert('RGB').tobytes()).hexdigest()};source[label+'_rgba_sha256']=hashlib.sha256(im.convert('RGBA').tobytes()).hexdigest()
 sources.append(source)
 assembled=Image.new('RGB',(1024,1024))
 for y in range(64):
  for x in range(64):
   tile=grid.getpixel((x,y));assembled.paste(atlas.crop((tile%16*16,tile//16*16,tile%16*16+16,tile//16*16+16)),(x*16,y*16))
 assembled.save(W/f'town{n}-source-map.png')
 for original in base['regions'][:42]:
  region={k:v for k,v in original.items()if k in ['id','archive','source_bounds','visible_bounds']};region['id']=region['id'].replace('town1',f'town{n}');region['archive']=archive
  name=region['id'];x0,y0,x1,y1=region['source_bounds'];crop=assembled.crop((x0*16,y0*16,x1*16,y1*16));ref=name+'-reference.png';large=name+'-reference-enlarged.png';crop.save(W/ref);crop.resize((768,768),Image.Resampling.NEAREST).save(W/large)
  region.update(reference=ref,enlarged_reference=large,reference_sha256=sha(W/ref),status='DRAFT')
  blue=sum(b>100 and b>r*1.5 and b>g*1.5 for r,g,b in crop.getdata());red=sum(r>100 and r>g*1.5 and g==b for r,g,b in crop.getdata())
  prompt=("Use case: stylized-concept. Paint ONE continuous native HD overhead-isometric RPG map scene on an opaque square1254x1254 canvas. "
   "References1 and2 are the EXACT same192x192 map source, original and enlarged. Reference3 is approved brushwork/material STYLE ONLY. Never copy its house or layout. "
   "Preserve ONLY objects evident in the FIRST reference, at EXACT normalized coordinates and silhouettes. This is a12x12 map-cell window; a cell shift is unacceptable. "
   "Source coordinates0..192 map proportionally to0..1254. Objects cut by any edge must remain cut there; never move or shrink a building, wall, bridge, road, canopy or shore to complete it. "
   "Restrained fine hand-painted terracotta, gray stone, white plaster/timber and rounded green trees ONLY where these actually appear. Newly paint native detail; no pixel-art upscale. "
   "Ground, roofs and paths must be coherent continuous surfaces across imaginary cell boundaries. Black is dark land shadow, not a rectangular hole or water. "
   "Do not invent any structures, props, people, labels, text, grid, gutters, UI or border. Do not reinterpret the source into a new composition. ")
  refs=[ref,large,style]
  if blue:
   prompt+="Reference4 is WATER MATERIAL STYLE ONLY. Source blue areas alone are calm evenly lit muted blue-green water. Plain water stays EMPTY: no new islands, towers, walls, bridges, boats, rocks, reflections or breaking waves. Preserve every shoreline and all partial edge fragments exactly. ";refs.append(waterstyle)
  else:prompt+="This exact source contains ZERO blue water pixels: paint NO lake, NO moat, NO canal, NO blue reflection anywhere. All dark corridors are land shadow. "
  if red<40:prompt+="This source has no red roof: paint NO red-roof building and never copy the style-reference house. "
  if blue>192*192*.92:prompt+="Almost the entire source is EMPTY WATER. Keep exactly that huge empty area, with land fragments ONLY at the reference edge locations. Do not add any other object. "
  prompt+='Exact source region '+name+'.'
  pf=name+'-prompt.txt';(W/pf).write_text(prompt+'\n');requests.append(dict(id=name,archive=archive,prompt=prompt,prompt_file=pf,references=refs));regions.append(region)
(W/'maps-plan.json').write_text(json.dumps({'schema':'kol-hd-full-position-maps/1','status':'DRAFT source audited,252 position paintings pending','source_zip_sha256':base['source_zip_sha256'],'maps':sources,'regions':regions,'generated':0},ensure_ascii=False,indent=2)+'\n')
(W/'maps-requests.json').write_text(json.dumps(requests,ensure_ascii=False,indent=2)+'\n')
print('six64x64 maps,252 regions, source archives count2 and all hashes verified')
