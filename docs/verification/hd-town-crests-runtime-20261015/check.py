from pathlib import Path
from PIL import Image, ImageChops
import hashlib,json
w=Path('/out');r=Path('/src');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
base=json.loads((w/'baseline-captures.json').read_text());formal=json.loads((w/'formal-captures.json').read_text())
assert len(base['frames'])==len(formal['frames'])==14
frames=base['frames']+formal['frames'];rows=[]
bg=Image.open(r/'assets/hd/first-slice/interface/background.png').convert('RGBA')
for i in range(1,8):
 key=f'twn{i}pics-000';fs={f['Variant']:f for f in frames if f['Key']==key}
 assert len(fs)==4
 for f in fs.values():assert sha(w/'gallery'/f['PNG'])==f['SHA'] and f['StateSHA']==fs['baseline-ega']['StateSHA'] and f['Text']==fs['baseline-ega']['Text'] and f['Han']>0
 im={v:Image.open(w/'gallery'/f['PNG']).convert('RGBA')for v,f in fs.items()}
 assert im['baseline-ega'].tobytes()==im['formal-ega'].tobytes()
 diff=ImageChops.difference(im['baseline-hd'],im['formal-hd']);whole=max(q[1]for q in diff.getextrema())
 diff.paste((0,0,0,0),(120,490,550,940));outside=max(q[1]for q in diff.getextrema());assert outside<=1,(key,'outside',outside)
 if i==7:assert whole<=1,(key,'preserved',whole)
 native=Image.open(r/f'assets/hd/first-slice/town-crests/{key}.png').convert('RGBA')
 expected=Image.alpha_composite(bg.crop((130,502,450,822)),native)
 nd=ImageChops.difference(expected,im['formal-hd'].crop((130,502,450,822)));native_max=max(q[1]for q in nd.getextrema());assert native_max<=1,(key,'native',native_max)
 rows.append(dict(key=key,ega_rgba_exact=True,complete_serializable_Outdoor_equal=True,chinese_equal=True,native_composite_max=native_max,outside_old_crest_rect_max=outside,whole_baseline_hd_max=whole,preserved_whole_hd_checked=i==7))
font=json.loads((r/'assets/hd/first-slice/interface/font-atlas.json').read_text());chars={c for f in frames for s in f['Text']for c in s if '\u3400'<=c<='\u9fff'};missing=sorted(chars-set(font['glyphs']));assert not missing
normal=[]
for name in ['winged-regression','human-slots']:
 dirs={v:w/'normal'/name/v for v in ['baseline-ega','prototype-ega','baseline-hd','prototype-hd']}
 traces={v:json.loads((d/'trace.json').read_text())for v,d in dirs.items()};assert all(q==traces['baseline-ega']for q in traces.values())
 assert len({(d/'text.jsonl').read_text()for d in dirs.values()})==1
 maxima={}
 for theme in ['ega','hd']:
  a,b=sorted((dirs['baseline-'+theme]/'shots').glob('*.png')),sorted((dirs['prototype-'+theme]/'shots').glob('*.png'));assert len(a)==len(b)
  maximum=max(max(q[1]for q in ImageChops.difference(Image.open(x).convert('RGBA'),Image.open(y).convert('RGBA')).getextrema())for x,y in zip(a,b))
  assert maximum<=(0 if theme=='ega' else 1),(name,theme,maximum);maxima[theme]=maximum
 normal.append(dict(case=name,states=len(traces['baseline-ega']['steps']),trace_equal=True,chinese_equal=True,ega_max=maxima['ega'],hd_max=maxima['hd']))
gui=[]
for key in ['twn1pics-000','twn7pics-000']:
 p=w/f'gui-{key}/results.json';q=json.loads(p.read_text());assert q['exit_code']==0 and q['serializable_Outdoor_unchanged'] and len(q['display_events'])==12 and len(q['rounds'])==2
 assert all(x['hd_max_channel_delta']<=1 for x in q['rounds'])
 gui.append(dict(case=key,events=12,rounds=q['rounds'],receipt_sha256=sha(p)))
saved=json.loads((w/'crest-save/results.json').read_text());assert saved['normal_save_equal_except_saved_at'] and saved['reload_file_unchanged']
result=dict(status='PASS limited seven-crest native render, Chinese, normal operations, F10 and v3',scenes=7,frames=28,records=rows,native_composite_max=max(q['native_composite_max']for q in rows),outside_max=max(q['outside_old_crest_rect_max']for q in rows),actual_han_characters=sum(f['Han']for f in frames),unique_han=len(chars),missing_glyphs=missing,normal=normal,states_per_variant=sum(q['states']for q in normal),gui=gui,F10_events=24,window_sizes=[[1920,1200],[960,600]],saved=saved,source_inputs_sha256=sha(w/'source-inputs.json'),asset_inputs_sha256=sha(w/'asset-inputs.json'),verification_source_inputs_sha256=sha(w/'verification-source-inputs.json'),formal_binary_sha256=sha(w/'bin/formal'),baseline_commit='bc309434b479e2785ba99fdd70ae088044584d40',platform='Linux Docker only',original_oracle=False,natural_completion=False,prepared_GUI_marker_clock='walkTick=-1073741824; production rules unchanged')
(w/'check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items()if k not in ['records','normal','gui','saved']},ensure_ascii=False),flush=True)
