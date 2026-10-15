from pathlib import Path
from PIL import Image
import numpy as np,json,hashlib,sys
w=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
prefix=sys.argv[1] if len(sys.argv)>1 else 'formal-monster-gallery-v3'
formal=json.loads((w/(prefix+'.json')).read_text());proto=json.loads((w/'all-monster-gallery-v1.json').read_text())
old={(f['id'],f['variant']):f for f in proto['frames']};records=[]
for f in formal['frames']:
    assert sha(w/prefix/f['png'])==f['sha256']
    q=old[f['id'],f['variant']]
    assert f['state_sha256']==q['state_sha256'] and f['text']==q['text'] and f['han_characters']==q['han_characters']
    # Both formal HD variants now contain monster art. Compare to the actual candidate.
    if f['variant'].endswith('hd'):q=old[f['id'],'prototype-hd']
    a=np.asarray(Image.open(w/prefix/f['png']).convert('RGBA'),dtype=np.int16)
    b=np.asarray(Image.open(w/'all-monster-gallery-v1'/q['png']).convert('RGBA'),dtype=np.int16)
    maximum=int(np.abs(a-b).max());assert maximum<=(1 if f['variant'].endswith('hd')else 0),(f['id'],f['variant'],maximum)
    records.append(dict(id=f['id'],variant=f['variant'],max=maximum,state_and_chinese_equal=True))
assert len(records)==512
result=dict(status='PASS formal vs isolated actual-game candidate',frames=512,scenes=128,ega_rgba_exact=True,hd_max=max(q['max']for q in records),records=records,source_inputs_sha256=sha(w/'final-source-inputs-v2.json'),assets_sha256=sha(w/'formal-asset-inputs.json'),original_oracle=False)
(w/'formal-vs-candidate-check.json').write_text(json.dumps(result,indent=2)+'\n')
print('Formal/candidate512frames PASS max',result['hd_max'],flush=True)
