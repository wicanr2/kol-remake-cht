from pathlib import Path
from PIL import Image
import numpy as np,json,hashlib
w=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();normal=[]
for name in ['winged-regression','human-slots']:
    root=w/'monster-normal-v2'/name
    traces={};text={};shots={}
    for v in ['baseline-ega','prototype-ega','baseline-hd','prototype-hd']:
        d=root/v;traces[v]=json.loads((d/'trace.json').read_text());text[v]=(d/'text.jsonl').read_text();shots[v]=sorted((d/'shots').glob('*.png'))
    assert all(q==traces['baseline-ega']for q in traces.values())
    assert len(set(text.values()))==1
    maxima={}
    for theme in ['ega','hd']:
        a,b=shots['baseline-'+theme],shots['prototype-'+theme];assert len(a)==len(b)
        maximum=0
        for p,q in zip(a,b):
            x=np.asarray(Image.open(p).convert('RGBA'),dtype=np.int16);y=np.asarray(Image.open(q).convert('RGBA'),dtype=np.int16)
            assert x.shape==y.shape;maximum=max(maximum,int(np.abs(x-y).max()))
        assert maximum<=(1 if theme=='hd'else 0),(name,theme,maximum);maxima[theme]=maximum
    normal.append(dict(case=name,states=len(traces['baseline-ega']['steps']),chinese_equal=True,controller_trace_equal=True,ega_rgba_exact=True,hd_max=maxima['hd']))
gui=[]
for name in ['info','encounter']:
    p=w/('gui-'+name)/'results.json';q=json.loads(p.read_text())
    assert q['status'].startswith('PASS') and q['exit_code']==0 and q['serializable_battle_state_unchanged']
    assert len(q['display_events'])==12 and len(q['rounds'])==2 and all(x['hd_max_channel_delta']<=1 for x in q['rounds'])
    gui.append(dict(case=name,receipt_sha256=sha(p),events=12,rounds=q['rounds']))
save=w/'monster-formal-save/results.json';s=json.loads(save.read_text());assert s['normal_save_equal_except_saved_at']and s['reload_file_unchanged']
result=dict(status='PASS limited normal regression, prepared actual F10 GUI and v3',normal=normal,normal_states_per_variant=sum(q['states']for q in normal),
    gui=gui,F10_events=24,window_sizes=[[1920,1200],[960,600]],serializable_battle_state_preserved=True,save_sha256=sha(save),save=s,original_oracle=False,normal_completion=False)
(w/'monster-regression-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print('Normal',result['normal_states_per_variant'],'states; F10',24,'events; save',s['rest_states'],'reload',s['reload_states'],'PASS',flush=True)
