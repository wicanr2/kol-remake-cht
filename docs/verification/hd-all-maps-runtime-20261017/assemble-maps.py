from pathlib import Path
from PIL import Image,ImageDraw
import json,sys
W=Path('/out');plan=json.loads((W/'maps-plan.json').read_text());side=32
for archive in sys.argv[1:] or sorted({r['archive'] for r in plan['regions'] if 'file'in r}):
 name=archive.lower();regions=[r for r in plan['regions'] if r['archive']==archive]
 assembled=Image.new('RGB',(64*side,64*side),(46,40,55))
 for r in regions:
  if 'file'not in r:continue
  im=Image.open(W/r['file']).convert('RGB').resize((12*side,12*side),Image.Resampling.BILINEAR)
  x0,y0,x1,y1=r['visible_bounds'];sx,sy,_,_=r['source_bounds']
  assembled.paste(im.crop(((x0-sx)*side,(y0-sy)*side,(x1-sx)*side,(y1-sy)*side)),(x0*side,y0*side))
 assembled.save(W/(name+'-assembly-diagnostic.png'))
 source=Image.open(W/(name+'-source-map.png')).convert('RGB').resize((64*side,64*side),Image.Resampling.NEAREST)
 for i,(top,bottom)in enumerate([(0,15),(15,27),(27,39),(39,51),(51,64)]):
  pair=Image.new('RGB',(2048,2*(bottom-top)*side+32),(28,24,32));d=ImageDraw.Draw(pair)
  d.text((8,8),archive+' SOURCE above / POSITION-BOUND CANDIDATES below; DRAFT geometry and seams',fill='white')
  pair.paste(source.crop((0,top*side,2048,bottom*side)),(0,32));pair.paste(assembled.crop((0,top*side,2048,bottom*side)),(0,32+(bottom-top)*side))
  pair.save(W/f'{name}-row{i}-diagnostic.png')
 print(archive,sum('file'in r for r in regions),'/42')
