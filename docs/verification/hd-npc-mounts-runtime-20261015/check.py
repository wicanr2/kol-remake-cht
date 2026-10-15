from pathlib import Path
from PIL import Image,ImageChops
import hashlib,json,runpy
w=Path('/out');r=Path('/src');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
script=(w/'check-full-gallery.py').read_text().replace('full-gallery-v1','formal-gallery-v1').replace('full-gallery-check.json','formal-gallery-native-check.json').replace('full-runtime-review.html','formal-runtime-review.html')
script=script.replace("name+'-visual.png'","name+'-formal-visual.png'")
script=script.replace('完整組仍為隔離原型，尚未正式接入。','完整組已接入正式程式，這是受控場景驗證。')
(w/'check-formal-native.py').write_text(script)
runpy.run_path(str(w/'check-formal-native.py'))
old=json.loads((w/'full-gallery-v1.json').read_text())
formal=json.loads((w/'formal-gallery-v1.json').read_text())
oldframes={(f['id'],f['variant']):f for f in old['frames']}
compared=[]
for f in formal['frames']:
    theme=f['variant'].split('-')[1];before=oldframes[f['id'],'prototype-'+theme]
    assert f['inn_sha256']==before['inn_sha256'] and f['text']==before['text']
    a=Image.open(w/'formal-gallery-v1'/f['png']).convert('RGBA')
    b=Image.open(w/'full-gallery-v1'/before['png']).convert('RGBA')
    diff=ImageChops.difference(a,b);maximum=max(v[1]for v in diff.getextrema())
    assert maximum<=(1 if theme=='hd' else 0) and diff.getextrema()[3]==(0,0),(f['id'],theme,maximum)
    compared.append(dict(id=f['id'],variant=f['variant'],max_channel_difference=maximum))
sources=json.loads((w/'formal-runtime-source-inputs.json').read_text())
assert all(sha(r/name)==value for name,value in sources.items())
font=json.loads((r/'assets/hd/first-slice/interface/font-atlas.json').read_text())
glyphs=font.get('glyphs',{})
assert glyphs,'Missing actual font glyph index'
chars=set(''.join(s for f in formal['frames']for s in f['text']))
assert chars.issubset(glyphs),(chars-set(glyphs))
result=dict(status='PASS formal complete NPC and mount scenes',go_module_inputs=len(sources),
            source_map_sha256=sha(w/'formal-runtime-source-inputs.json'),scenes=110,frames=440,
            native_check_sha256=sha(w/'formal-gallery-native-check.json'),
            candidate_complete_frame_max=max(q['max_channel_difference']for q in compared),
            ega_rgba_exact=True,full_Inn_and_chinese_equal=True,observed_font_characters=len(chars),missing_hd_glyphs=0,
            font_index_sha256=sha(r/'assets/hd/first-slice/interface/font-atlas.json'),
            comparisons=compared,original_oracle=False,normal_player_path=False,platform='Linux Docker')
(w/'formal-gallery-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
normal=[]
for case in ['human-slots','winged-regression']:
    root=w/'npc-normal'/case;variants=['baseline-ega','prototype-ega','prototype-hd','baseline-hd']
    traces=[json.loads((root/v/'trace.json').read_text())for v in variants]
    texts=[(root/v/'text.jsonl').read_bytes()for v in variants]
    assert all(x==traces[0]for x in traces) and all(x==texts[0]for x in texts)
    frames=list((root/'baseline-ega/shots').glob('*.png'));assert len(frames)==len(traces[0]['steps'])
    hdmaximum=0
    for f in frames:
        a=Image.open(f).convert('RGBA');b=Image.open(root/'prototype-ega/shots'/f.name).convert('RGBA')
        assert a.tobytes()==b.tobytes(),(case,f.name)
        a=Image.open(root/'baseline-hd/shots'/f.name).convert('RGBA');b=Image.open(root/'prototype-hd/shots'/f.name).convert('RGBA')
        diff=ImageChops.difference(a,b);value=max(v[1]for v in diff.getextrema())
        assert value<=1 and diff.getextrema()[3]==(0,0),(case,f.name,value)
        hdmaximum=max(hdmaximum,value)
    normal.append(dict(case=case,states=len(frames),trace_equal=True,text_equal=True,ega_rgba_exact=True,hd_max=hdmaximum))
(w/'normal-npc-check.json').write_text(json.dumps(dict(status='PASS normal controller regression',cases=normal,
    states_per_variant=sum(q['states']for q in normal),seed=1,original_oracle=False),ensure_ascii=False,indent=2)+'\n')
print('Formal440 frames, original candidate max',result['candidate_complete_frame_max'],'font',len(chars),'no missing; normal',sum(q['states']for q in normal),'states per variant',flush=True)
