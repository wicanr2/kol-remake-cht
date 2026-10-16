from pathlib import Path
from PIL import Image
import json,hashlib,shutil,sys
W=Path('/out');G=Path('/generated');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
assert W.stat().st_uid==1000
plan=json.loads((W/'maps-plan.json').read_text());regions={r['id']:r for r in plan['regions']}
for rec in json.loads((W/sys.argv[1]).read_text()):
 src=G/Path(rec['path']).relative_to('/home/anr2/.codex/generated_images');target=W/rec.get('file',rec['id']+'-native.png')
 if target.exists():assert sha(target)==sha(src)
 else:shutil.copyfile(src,target)
 im=Image.open(target).convert('RGBA');assert im.width==im.height and 1152<=im.width<=4096 and im.getchannel('A').getextrema()==(255,255)
 r=regions[rec['id']];pf=rec.get('prompt_file',r['id']+'-prompt.txt')
 if 'reason'in rec:r.setdefault('revisions',[]).append(dict(replaced_file=r.get('file'),replacement=target.name,reason=rec['reason']))
 r.update(file=target.name,png_sha256=sha(target),rgba_sha256=hashlib.sha256(im.tobytes()).hexdigest(),native_size=list(im.size),prompt_file=pf,prompt_sha256=sha(W/pf),status='DRAFT native verified; visual/source geometry review pending')
plan['generated']=sum('file'in r for r in plan['regions']);(W/'maps-plan.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n');print('saved native maps',plan['generated'],'/252')
