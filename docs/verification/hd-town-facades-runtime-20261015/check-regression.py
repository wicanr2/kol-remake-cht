from pathlib import Path
from PIL import Image, ImageChops
import json
import hashlib

w = Path('/out')
normal = []
for name in ['winged-regression', 'human-slots']:
    traces, text, shots = {}, {}, {}
    for v in ['baseline-ega', 'prototype-ega', 'baseline-hd', 'prototype-hd']:
        d = w / 'facade-normal' / name / v
        traces[v] = json.loads((d / 'trace.json').read_text())
        text[v] = (d / 'text.jsonl').read_text()
        shots[v] = sorted((d / 'shots').glob('*.png'))
    assert all(x == traces['baseline-ega'] for x in traces.values())
    assert len(set(text.values())) == 1
    maxima = {}
    for theme in ['ega', 'hd']:
        a, b = shots['baseline-' + theme], shots['prototype-' + theme]
        assert len(a) == len(b)
        maximum = 0
        for left, right in zip(a, b):
            x, y = Image.open(left).convert('RGBA'), Image.open(right).convert('RGBA')
            assert x.size == y.size
            maximum = max(maximum, max(q[1] for q in ImageChops.difference(x, y).getextrema()))
        assert maximum <= (1 if theme == 'hd' else 0), (name, theme, maximum)
        maxima[theme] = maximum
    normal.append(dict(case=name, states=len(traces['baseline-ega']['steps']), chinese_equal=True,
                       controller_trace_equal=True, ega_rgba_exact=True, hd_max=maxima['hd']))
gui = []
for key in ['twn1pics-003', 'twn5pics-019', 'twn7pics-013', 'miscpics-000']:
    p = w / ('gui-' + key) / 'results.json'
    q = json.loads(p.read_text())
    assert q['status'].startswith('PASS') and q['exit_code'] == 0 and q['serializable_Outdoor_unchanged']
    assert len(q['display_events']) == 12 and len(q['rounds']) == 2
    assert all(x['hd_max_channel_delta'] <= 1 for x in q['rounds'])
    gui.append(dict(case=key, events=12, rounds=q['rounds'], receipt_sha256=hashlib.sha256(p.read_bytes()).hexdigest()))
saved = w / 'facade-formal-save/results.json'
save = json.loads(saved.read_text())
assert save['normal_save_equal_except_saved_at'] and save['reload_file_unchanged']
result = dict(status='PASS limited normal, controlled F10 and v3 regression', normal=normal,
              states_per_variant=sum(q['states'] for q in normal), gui=gui, F10_events=48,
              window_sizes=[[1920,1200],[960,600]], serializable_Outdoor_unchanged=True,
              prepared_GUI_marker_clock='walkTick=-1073741824; static marker phase only; production rules unchanged',
              save=save, save_receipt_sha256=hashlib.sha256(saved.read_bytes()).hexdigest(),
              original_oracle=False, normal_completion=False, platform='Linux Docker only')
(w/'regression-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(result,ensure_ascii=False),flush=True)
