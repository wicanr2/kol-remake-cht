from pathlib import Path
import hashlib,json,unicodedata
import numpy as np
from PIL import Image

r=Path('/src'); w=Path('/out')
def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.file_digest(p.open('rb'),'sha256').hexdigest()
def pixels(p): return np.asarray(Image.open(p).convert('RGBA'),dtype=np.int16)
def diff(a,b):
    assert a.shape==b.shape
    d=np.abs(a-b)
    return dict(max_channel_delta=int(d.max()),different_pixels=int(np.count_nonzero(d.max(2))))
dirs=[w/'scenes-linux',w/'scenes-windows']
records=[read(d/'motion-captures.json') for d in dirs]
assert all(d['scenes']==13 and len(d['frames'])==52 for d in records)
maps=[{(f['Name'],f['Variant']):f for f in d['frames']} for d in records]
assert set(maps[0])==set(maps[1]) and len(maps[0])==52
glyphs=read(r/'assets/hd/first-slice/interface/font-atlas.json')['glyphs']
comparisons=[]; old_comparisons=[]; roundtrips=[]; missing=set()
for key in sorted(maps[0]):
    a,b=maps[0][key],maps[1][key]
    assert a['StateSHA']==b['StateSHA'] and a['Text']==b['Text'],key
    limit=0 if key[1].endswith('ega') else 1
    check=diff(pixels(dirs[0]/'motion-runtime'/a['PNG']),pixels(dirs[1]/'motion-runtime'/b['PNG']))
    comparisons.append(dict(name=key[0],variant=key[1],limit=limit,**check))
    assert check['max_channel_delta']<=limit,(key,check)
    for directory,row in zip(dirs,[a,b]):
        assert sha(directory/'motion-runtime'/row['PNG'])==row['SHA']
        for text in row['Text'] or []:
            missing.update(c for c in text if unicodedata.category(c)=='Lo' and '\u4e00'<=c<='\u9fff' and c not in glyphs)
        prior=r/'dist-all/v.1.0.3-20261017/smoke'/('hd-scenes-linux' if directory==dirs[0] else 'hd-scenes-windows')
        old=diff(pixels(prior/'motion-runtime'/row['PNG']),pixels(directory/'motion-runtime'/row['PNG']))
        old_comparisons.append(dict(platform=directory.name,name=key[0],variant=key[1],limit=limit,**old))
        assert old['max_channel_delta']<=limit,(directory,key,old)
for directory,rows in zip(dirs,maps):
    for name in sorted({n for n,v in rows}):
        frames=[rows[name,v] for v in ['first-ega','first-hd','restored-ega','restored-hd']]
        assert len({f['StateSHA'] for f in frames})==1
        assert all(f['Text']==frames[0]['Text'] for f in frames)
        for theme,a,b,limit in [('ega',frames[0],frames[2],0),('hd',frames[1],frames[3],1)]:
            d=diff(pixels(directory/'motion-runtime'/a['PNG']),pixels(directory/'motion-runtime'/b['PNG']))
            roundtrips.append(dict(platform=directory.name,name=name,theme=theme,limit=limit,**d))
            assert d['max_channel_delta']<=limit,(name,theme,d)
assert not missing,missing
result=dict(status='PASS',scenes=13,frames_per_platform=52,cross_platform=comparisons,
            previous_release=old_comparisons,roundtrips=roundtrips,missing_glyphs=[],
            model_and_chinese_unchanged=True,original_parity=False,real_windows_machine=False)
(w/'captures-comparison.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('PASS',max(x['max_channel_delta'] for x in comparisons),max(x['max_channel_delta'] for x in old_comparisons))
