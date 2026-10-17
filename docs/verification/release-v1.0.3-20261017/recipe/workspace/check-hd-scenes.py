from pathlib import Path
import json,hashlib,sys,unicodedata
import numpy as np
from PIL import Image

R,O=Path('/src'),Path('/out')
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def pixels(p):return np.asarray(Image.open(p).convert('RGBA')).astype(np.int16)
def difference(a,b):
    assert a.shape==b.shape,(a.shape,b.shape)
    d=np.abs(a-b)
    return dict(max_channel_delta=int(d.max()),different_pixels=int(np.count_nonzero(d.max(2))))

native=O/'smoke/hd-scenes-linux';wine=O/'smoke/hd-scenes-windows'
expected=read(R/'docs/verification/hd-full-runtime-20261017/source-inputs.json')
assert read(native/'source-inputs.json')==expected
assert all(sha(R/n)==h for n,h in expected.items())
records=[read(d/'motion-captures.json') for d in [native,wine]]
assert all(d['scenes']==13 and len(d['frames'])==52 and d['controlled'] is True for d in records)
keys=[{(f['Name'],f['Variant']):f for f in d['frames']} for d in records]
assert set(keys[0])==set(keys[1]) and len(keys[0])==52
hd_glyphs=read(R/'assets/hd/first-slice/interface/font-atlas.json')['glyphs']
missing=set();comparisons=[];roundtrips=[];models=True
for key in sorted(keys[0]):
    left,right=keys[0][key],keys[1][key]
    for directory,row in [(native,left),(wine,right)]:
        assert sha(directory/'motion-runtime'/row['PNG'])==row['SHA']
        for text in row['Text'] or []:
            missing.update(c for c in text if '\u4e00'<=c<='\u9fff' and c not in hd_glyphs)
    models=models and left['StateSHA']==right['StateSHA'] and left['Text']==right['Text']
    diff=difference(pixels(native/'motion-runtime'/left['PNG']),pixels(wine/'motion-runtime'/right['PNG']))
    limit=0 if key[1].endswith('ega') else 1
    comparisons.append(dict(name=key[0],variant=key[1],bound=limit,passed=diff['max_channel_delta']<=limit,**diff))
for directory,frames in zip([native,wine],keys):
    for name in sorted({k[0] for k in frames}):
        rows=[frames[name,v] for v in ['first-ega','first-hd','restored-ega','restored-hd']]
        models=models and len({r['StateSHA'] for r in rows})==1 and all(r['Text']==rows[0]['Text'] for r in rows)
        for theme,a,b,limit in [('ega',rows[0],rows[2],0),('hd',rows[1],rows[3],1)]:
            diff=difference(pixels(directory/'motion-runtime'/a['PNG']),pixels(directory/'motion-runtime'/b['PNG']))
            roundtrips.append(dict(platform=directory.name,name=name,theme=theme,bound=limit,passed=diff['max_channel_delta']<=limit,**diff))
passed=models and not missing and all(x['passed'] for x in comparisons+roundtrips)
result=dict(schema='kol-release-hd-scenes/1',status='PASS' if passed else 'NUMERIC_OR_STATE_GATE_UNMET',
            version='v.1.0.3-20261017',scenes=13,frames_per_platform=52,models_and_chinese_equal=models,
            missing_glyphs=sorted(missing),observed_Han=sum(f['Han'] for f in records[0]['frames'] if f['Variant']=='first-ega'),
            comparisons=comparisons,roundtrips=roundtrips,source_receipt_sha256=sha(R/'docs/verification/hd-full-runtime-20261017/receipt.json'),
            production_changed=False,original_oracle=False,normal_player_path=False,
            environment={'GOMEMLIMIT':'3GiB','GOMAXPROCS':'3','limit_memory':'12GiB','driver_entry':'main OS thread'})
(O/'smoke/hd-scenes-comparison.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'status':result['status'],'scenes':13,'cross_platform_max':max(x['max_channel_delta'] for x in comparisons),
                  'roundtrip_max':max(x['max_channel_delta'] for x in roundtrips),'models_equal':models,'missing_glyphs':sorted(missing)},ensure_ascii=False))
sys.exit(0 if passed else 1)
