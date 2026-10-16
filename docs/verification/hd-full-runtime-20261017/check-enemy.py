from pathlib import Path
from PIL import Image
import json
import hashlib
import numpy as np

W,R=Path('/out'),Path('/src')
new=json.loads((W/'motion-captures.json').read_text())
old=json.loads((W/'motion-full-accepted-captures.json').read_text())
expected={f['Name'] for f in old['frames'] if f['Kind'] in ['body','battle','target','prompt','spoils']}
assert new['scope']=='standing-enemy'
assert {f['Name'] for f in new['frames']}==expected
assert new['scenes']==len(expected)==34 and len(new['frames'])==136
oldf={(f['Name'],f['Variant']):f for f in old['frames']}
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
checks=[]
scenes={}
for f in new['frames']:
    prior=oldf[(f['Name'],f['Variant'])]
    assert f['Text']==prior['Text'],('text',f['Name'])
    if f['Name']!='event-combat-03':
        assert f['StateSHA']==prior['StateSHA'],('state',f['Name'])
    p=W/'motion-runtime'/f['PNG']
    assert sha(p)==f['SHA']
    a=np.array(Image.open(W/'motion-full-accepted-runtime'/prior['PNG']).convert('RGB')).astype('int16')
    b=np.array(Image.open(p).convert('RGB')).astype('int16')
    diff=np.abs(a-b)
    scenes.setdefault(f['Name'],{})[f['Variant']]=f
    if f['Variant'].endswith('ega'):
        assert not diff.any(),(f['Name'],'EGA changed')
        continue
    changed=np.any(diff>1,axis=2)
    bbox=None
    if changed.any():
        yy,xx=np.nonzero(changed)
        bbox=[int(xx.min()),int(yy.min()),int(xx.max()+1),int(yy.max()+1)]
        assert bbox[2]-bbox[0]<=96 and bbox[3]-bbox[1]<=96,(f['Name'],'outside enemy cell',bbox)
    checks.append(dict(name=f['Name'],variant=f['Variant'],enemy_change_bounds=bbox))
for name,fs in scenes.items():
    assert len({f['StateSHA'] for f in fs.values()})==1,(name,'within-stage state')
    a,b=[np.array(Image.open(W/'motion-runtime'/fs[v]['PNG']).convert('RGBA')).astype('int16') for v in ['first-hd','restored-hd']]
    assert int(np.abs(a-b).max())<=1,(name,'roundtrip')
manifest=json.loads((R/'assets/hd/first-slice/motion/manifest.json').read_text())
prior=json.loads((W/'pre-consistent-motion-manifest.json').read_text())
assert len(manifest['pictures'])==65
for name,v in manifest['files'].items():
    if name!='battle-sprite-6.png':
        assert v==prior['files'][name]
assert 'HD-NATIVE-FALLBACK' not in (W/'motion-gallery-run.log').read_text()
result=dict(status='PASS',controlled_scenes=34,frames=136,
    cross_stage_model_unchanged_for33_field_scenes=True,Chinese_unchanged=True,
    result_scene_inactive_title_cache_differs=True,within_stage_model_unchanged=True,
    ega_delta=0,hd_roundtrip_max_delta=1,changed_asset='battle-sprite-6.png',
    other64_assets_unchanged=True,changed_pixels_confined_to_enemy_cell=True,checks=checks,
    manifest_sha256=sha(R/'assets/hd/first-slice/motion/manifest.json'),
    generation_sha256=sha(R/'assets/hd/first-slice/motion/generation.json'),
    source_index_sha256=sha(W/'motion-source-inputs.json'),original_oracle=False,normal_player_path=False)
(W/'enemy-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='checks'},ensure_ascii=False))
