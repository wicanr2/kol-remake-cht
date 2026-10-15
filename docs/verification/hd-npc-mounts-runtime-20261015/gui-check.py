from pathlib import Path
from PIL import Image,ImageChops
import json,hashlib
w=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
results=[]
for name in ['npc','native-npc','stable','sheet']:
    p=w/('gui-'+name)/'results.json';q=json.loads(p.read_text())
    assert q['status'].startswith('PASS') and q['exit_code']==0 and q['full_Inn_unchanged']
    assert len(q['display_events'])==12 and len(q['rounds'])==2
    assert all(x['hd_max_channel_delta']<=1 for x in q['rounds'])
    results.append(dict(case=name,receipt_sha256=sha(p),events=12,rounds=q['rounds']))
m=json.loads((w/'complete-creature-prototype.json').read_text())
row=next(q for q in m['portraits']if q['source']['archive']=='TWN5PICS' and q['source']['index']==22)
source=Image.open(w/row['file']).convert('RGBA').crop(row['rect'])
expected=source.transform((528,384),Image.Transform.AFFINE,(source.width/528,0,0,0,source.height/384,0),Image.Resampling.BILINEAR)
actual=Image.open(w/'gui-native-npc/large-before.png').convert('RGBA').crop((48,384,576,768))
diff=ImageChops.difference(expected.crop((2,2,526,382)),actual.crop((2,2,526,382)))
native_max=max(v[1]for v in diff.getextrema());assert native_max<=1,native_max
result=dict(status='PASS actual game.Update, F10 handler, render and presentation on prepared flow.SceneInn',
            cases=4,rounds=8,events=48,window_sizes=[[1920,1200],[960,600]],full_Inn_preserved=True,
            hd_roundtrip_max=max(x['hd_max_channel_delta']for q in results for x in q['rounds']),
            native_service_npc_interior_max=native_max,native_rectangle=[48,384,576,768],border_excluded=2,
            results=results,original_oracle=False,normal_player_path=False,
            limits='Prepared real door and mount display state; GUI tests display controls only. Normal regression/save checks are separate.')
(w/'npc-gui-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('F10',result['events'],'events; roundtrip',result['hd_roundtrip_max'],'native service image',native_max,flush=True)
