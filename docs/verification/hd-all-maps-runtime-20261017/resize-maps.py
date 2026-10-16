from pathlib import Path
from PIL import Image, __version__ as pillow_version
import hashlib, json, shutil

W = Path('/out')
assert W.stat().st_uid == 1000
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
plan = json.loads((W / 'maps-plan.json').read_text())
by_id = {r['id']: r for r in plan['regions']}
native_bytes = final_bytes = 0
records = []
for r in plan['regions']:
    if 'file' not in r:
        c = by_id['town7-r04-c04']
        assert r['archive'] == 'TOWN7'
        assert r['source_region_rgb_sha256'] == c['source_region_rgb_sha256']
        r.update({k: c[k] for k in ['file', 'png_sha256', 'rgba_sha256', 'native_size', 'prompt_file', 'prompt_sha256']})
        r['reuse_of'] = c['id']
        r['reuse_evidence'] = 'identical complete 192x192 source RGB SHA-256; independent position bounds retained'
    p = W / r['file']
    im = Image.open(p).convert('RGBA')
    assert sha(p) == r['png_sha256']
    assert hashlib.sha256(im.tobytes()).hexdigest() == r['rgba_sha256']
    assert im.size == tuple(r['native_size'])
    assert im.getchannel('A').getextrema() == (255, 255)
    out = W / (r['id'] + '-768.png')
    im.resize((768, 768), Image.Resampling.LANCZOS).save(out, optimize=True)
    final = Image.open(out).convert('RGBA')
    assert final.size == (768, 768) and final.getchannel('A').getextrema() == (255, 255)
    r['delivery_file'] = out.name
    r['delivery_png_sha256'] = sha(out)
    r['delivery_rgba_sha256'] = hashlib.sha256(final.tobytes()).hexdigest()
    r['delivery_size'] = [768, 768]
    r['status'] = 'DRAFT delivery resized; actual scene acceptance pending'
    record = {k: r[k] for k in ['id', 'archive', 'file', 'png_sha256', 'rgba_sha256', 'native_size', 'delivery_file', 'delivery_png_sha256', 'delivery_rgba_sha256', 'delivery_size']}
    record.update(tool='Pillow ' + pillow_version, resize='Image.Resampling.LANCZOS', native_bytes=p.stat().st_size, delivery_bytes=out.stat().st_size)
    if 'reuse_of' in r:
        record['reuse_of'] = r['reuse_of']
    records.append(record)
    native_bytes += p.stat().st_size
    final_bytes += out.stat().st_size
plan['generated'] = 240
plan['bindings'] = 252
plan['status'] = 'DRAFT complete new map delivery assets; geometry and runtime acceptance pending'
plan['capacity_decision'] = 'User chose smaller game assets and all future new images smaller; new position map delivery768x768; existing TOWN1/2 unchanged'
(W / 'maps-plan.json').write_text(json.dumps(plan, ensure_ascii=False, indent=2) + '\n')
receipt = dict(schema='kol-hd-map-resize/1', approval=plan['capacity_decision'], count=len(records), native_bytes=native_bytes, delivery_bytes=final_bytes, reduction=1-final_bytes/native_bytes, records=records)
(W / 'resize-receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
scope = json.loads((W / 'plan.json').read_text())
scope['capacity_decision'] = plan['capacity_decision']
scope['tools'] = sorted(set(scope.get('tools', []) + ['resize-maps.py', 'assemble-maps.py', 'resize-receipt.json', 'maps-plan.json']))
scope['remaining']['TOWN3-TOWN7'] = 'all paint generated; delivery768 ready; scene verification pending'
scope['remaining']['world-map'] = 'all paint generated; delivery768 ready; runtime and scene verification pending'
(W / 'plan.json').write_text(json.dumps(scope, ensure_ascii=False, indent=2) + '\n')
print(json.dumps({k: v for k, v in receipt.items() if k != 'records'}))
