from pathlib import Path
from PIL import Image,ImageChops
import hashlib,json
w=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
results=[]
for name in ['human-slots','six-build-party']:
 root=w/'formal-action-normal'/name
 variants=['baseline-ega','prototype-ega','prototype-hd','baseline-hd']
 traces=[json.loads((root/v/'trace.json').read_text()) for v in variants]
 texts=[(root/v/'text.jsonl').read_bytes() for v in variants]
 assert all(x==traces[0] for x in traces) and all(x==texts[0] for x in texts)
 base=root/'baseline-ega/shots';formal=root/'prototype-ega/shots'
 files=sorted(base.glob('*.png'));assert len(files)==len(traces[0]['steps'])
 for p in files:
  delta=ImageChops.difference(Image.open(p).convert('RGBA'),Image.open(formal/p.name).convert('RGBA'))
  assert all(x==(0,0) for x in delta.getextrema()),(name,p.name)
 changed=[]
 for p in sorted((root/'prototype-hd/shots').glob('*.png')):
  old=root/'baseline-hd/shots'/p.name
  mx=max(x[1] for x in ImageChops.difference(Image.open(p).convert('RGBA'),Image.open(old).convert('RGBA')).getextrema())
  if mx:changed.append(dict(file=p.name,max_delta=mx))
 results.append(dict(case=name,states=len(traces[0]['steps']),trace_equal=True,text_equal=True,ega_rgb_and_alpha_equal=True,trace_sha256=sha(root/'prototype-hd/trace.json'),text_sha256=sha(root/'prototype-hd/text.jsonl'),intentional_hd_changes=changed))
 print(name,len(traces[0]['steps']),'states; EGA exact; Chinese record equal;',len(changed),'changed HD frames',flush=True)
(w/'formal-action-normal-check.json').write_text(json.dumps(dict(schema='kol-hd-action-normal-check/1',cases=results,normal_from_title=True,seed=1,source_original_seed=None,limits='Remake normal-controller regression; no original oracle. Full equipment rendering remains DRAFT.'),ensure_ascii=False,indent=2)+'\n')
