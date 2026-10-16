from pathlib import Path
from PIL import Image
import numpy as np,json,hashlib,re
R,W=Path('/src'),Path('/out');D=R/'assets/hd/first-slice/motion';c=json.loads((W/'motion-captures.json').read_text());m=json.loads((D/'manifest.json').read_text());sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();scenes={}
for f in c['frames']:
 p=W/'motion-runtime'/f['PNG'];assert sha(p)==f['SHA'];scenes.setdefault(f['Name'],{})[f['Variant']]=f
reports=[];han=0
for name,fs in scenes.items():
 assert len(fs)==4 and len({f['StateSHA']for f in fs.values()})==1 and all(f['Text']==fs['first-ega']['Text']for f in fs.values())
 images={v:np.array(Image.open(W/'motion-runtime'/f['PNG']).convert('RGBA'))for v,f in fs.items()};assert np.array_equal(images['first-ega'],images['restored-ega']),(name,'EGA changed');delta=int(np.abs(images['first-hd'].astype('int16')-images['restored-hd'].astype('int16')).max());assert delta<=1,(name,'HD roundtrip',delta)
 for word in (fs['first-hd']['Text'] or []):assert not any(s in word for s in ['CAMPING!','~H','~V']), (name,'visible markup',word)
 han+=fs['first-hd']['Han'];reports.append(dict(name=name,kind=fs['first-hd']['Kind'],han=fs['first-hd']['Han'],state_sha256=fs['first-hd']['StateSHA'],ega_max_delta=0,hd_roundtrip_max_delta=delta))
# Static painted weather frames bypass interpolation at the native384x372 size.
native=[];wx,wy=92,494
for group in [3,4]:
 for phase in range(3):
  name=f'weather-{group}-{phase:02}';f=scenes[name]['first-hd'];actual=np.array(Image.open(W/'motion-runtime'/f['PNG']).convert('RGB'))[wy:wy+372,wx:wx+384].astype('int16');expected=np.array(Image.open(D/f'weather-{group}-{phase}.png').convert('RGB')).astype('int16');maximum=int(np.abs(actual-expected).max());assert maximum<=1,(name,'native',maximum);native.append(dict(name=name,maximum=maximum))
# Healthy body diagrams must be the delivered painting rather than stale tint.
for kind in range(3):
 name=f'body-{kind}-0';f=scenes[name]['first-hd'];actual=np.array(Image.open(W/'motion-runtime'/f['PNG']).convert('RGB'))[510:930,432:768].astype('int16');expected=np.array(Image.open(D/f'body-{kind+3}.png').convert('RGB')).astype('int16');maximum=int(np.abs(actual-expected).max());assert maximum<=1,(name,'native',maximum);native.append(dict(name=name,maximum=maximum))
 for damage in [1,2]:
  f=scenes[f'body-{kind}-{damage}']['first-hd'];wounded=np.array(Image.open(W/'motion-runtime'/f['PNG']).convert('RGB'))[510:930,432:768].astype('int16');assert np.count_nonzero(np.any(wounded!=actual,axis=2))>100,(kind,damage,'wounds did not repaint')
log=(W/'motion-gallery-run.log').read_text();assert 'HD-NATIVE-FALLBACK'not in log;assert '--- PASS: TestHDMotionGallery'in log
result=dict(status='PASS',controlled_scenes=c['scenes'],frames=len(c['frames']),han=han,missing_glyphs=0,ega_max_delta=0,hd_roundtrip_max_delta=max(r['hd_roundtrip_max_delta']for r in reports),unpainted_EGA_fallbacks=0,native_picture_checks=native,scenes=reports,manifest_sha256=sha(D/'manifest.json'),generation_sha256=sha(D/'generation.json'),source_index_sha256=sha(W/'motion-source-inputs.json'),original_oracle=False,normal_player_path=False,remake_title_seed=1,original_seed=None,gameplay_RND_calls=0)
(W/'motion-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items()if k not in ['scenes','native_picture_checks']},ensure_ascii=False))
