from pathlib import Path
import json,sys,hashlib
sys.path.insert(0,'tools')
from build_asset_contact_sheets import read_image
from png_io import read_png
root=Path('workplace/inventory-empty-slots-20261001')
source=Path('assets/original-derived-png/ega/WEATPICS')
images=[next(source.glob(f'weatpics-{i:03d}-*.png')) for i in range(18)]
raw=[read_image(p).rgb() for p in images]
def composite(group,phase):
    if group in (3,4):return raw[group*3+phase%3]
    base,overlay,mask=raw[group*3:group*3+3]
    out=bytearray(base)
    for y in range(31):
        for x in range(64):
            src=(y*128+(x-phase+128)%128)*3
            if mask[src:src+3]==b'\0\0\0':
                i=(y*64+x)*3;out[i:i+3]=overlay[src:src+3]
    return bytes(out)
rows=[]
for step in (45,90):
    paths=[root/'baseline-before/full-chain-ending'/f'step-{step:03d}.png',
           root/'all-before/full-chain-ending'/f'step-{step:03d}.png']
    a,b=[read_png(p) for p in paths]
    assert(a.width,a.height)==(b.width,b.height)==(320,200)
    pa,pb=a.rgb(),b.rgb()
    differences=[i//3 for i in range(0,len(pa),3) if pa[i:i+3]!=pb[i:i+3]]
    assert all(16<=i%320<80 and 76<=i//320<138 for i in differences)
    candidates=[]
    for img in (a,b):
        crop=img.crop(16,76,64,62).rgb()
        found=[{'group':g,'phase':p} for g in range(6) for p in range(64)
               if composite(g,p)==crop]
        assert found,(step,'無法從正式天候圖與合成規則重生')
        candidates.append(found)
    rows.append({'step':step,'difference_pixels':len(differences),'outside_weather_differences':0,
                 'old_candidates':candidates[0],'new_candidates':candidates[1],
                 'images_sha256':[hashlib.sha256(p.read_bytes()).hexdigest() for p in paths]})
result={'schema':'kol-weather-image-review-v1','rows':rows,
        'source_sha256':{str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in images},
        'scope':'只核對已存兩張天候框；不是原版時序、天候序列或其他畫格的驗收'}
(root/'weather-review.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(rows,ensure_ascii=False,indent=2))
