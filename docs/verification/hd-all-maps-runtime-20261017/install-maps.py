from pathlib import Path
from PIL import Image
import hashlib, json, shutil

R, W = Path('/src'), Path('/out')
D = R / 'assets/hd/first-slice/town-maps'
assert W.stat().st_uid == D.stat().st_uid == 1000
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
plan = json.loads((W / 'maps-plan.json').read_text())
manifest = json.loads((D / 'manifest.json').read_text())
generation = json.loads((D / 'generation.json').read_text())
assert len(manifest['maps']) == 2 and len(manifest['regions']) == 84
(W / 'previous-manifest.json').write_bytes((D / 'manifest.json').read_bytes())
(W / 'previous-generation.json').write_bytes((D / 'generation.json').read_bytes())
generation['previous_generation_sha256'] = sha(D / 'generation.json')
coverage = {s['archive']: [0]*4096 for s in plan['maps']}
for r in plan['regions']:
    p = W / r['delivery_file']
    assert sha(p) == r['delivery_png_sha256']
    im = Image.open(p).convert('RGBA')
    assert im.size == (768, 768) and im.getchannel('A').getextrema() == (255, 255)
    assert hashlib.sha256(im.tobytes()).hexdigest() == r['delivery_rgba_sha256']
    name = r['id'] + '.png'
    assert not (D / name).exists()
    shutil.copyfile(p, D / name)
    manifest['files'][name] = dict(sha256=sha(p), width=768, height=768)
    manifest['regions'].append({k: r[k] for k in ['id', 'archive', 'source_bounds', 'visible_bounds']} | dict(file=name, native_rgba_sha256=r['delivery_rgba_sha256']))
    record = {k: r[k] for k in ['id', 'archive', 'source_bounds', 'visible_bounds', 'source_region_rgb_sha256', 'reference_sha256']}
    record.update(native_sha256=r['png_sha256'], native_rgba_sha256=r['rgba_sha256'], native_size=r['native_size'], native_file=r['file'], prompt=(W/r['prompt_file']).read_text(), prompt_sha256=r['prompt_sha256'], delivery_file=name, delivery_sha256=r['delivery_png_sha256'], delivery_rgba_sha256=r['delivery_rgba_sha256'], delivery_size=[768,768], tool='built-in imagegen; original bytes retained locally; Pillow9.4.0 LANCZOS to768x768 optimize=True', evidence='READY source/format and geometry reviewed; controlled runtime acceptance pending')
    if 'reuse_of' in r:
        record.update(reuse_of=r['reuse_of'], reuse_evidence=r['reuse_evidence'])
    generation['regions'].append(record)
    x0,y0,x1,y1 = r['visible_bounds']
    for y in range(y0,y1):
        for x in range(x0,x1): coverage[r['archive']][y*64+x] += 1
assert all(set(v) == {1} for v in coverage.values())
manifest['maps'].extend(plan['maps'])
generation['maps'].extend(plan['maps'])
generation['capacity_decision'] = plan['capacity_decision']
generation['status'] = 'READY all eight position map groups; controlled scene acceptance pending; other HD remains open'
(D/'generation.json').write_text(json.dumps(generation, ensure_ascii=False, indent=2)+'\n')
manifest['generation_sha256'] = sha(D/'generation.json')
(D/'manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2)+'\n')
for r in json.loads((W/'previous-manifest.json').read_text())['regions']:
    assert sha(D/r['file']) == manifest['files'][r['file']]['sha256']
scope = json.loads((W/'plan.json').read_text())
scope['tools'] = sorted(set(scope['tools']+['install-maps.py', 'previous-manifest.json', 'previous-generation.json']))
(W/'plan.json').write_text(json.dumps(scope, ensure_ascii=False, indent=2)+'\n')
print('Installed8 maps336 regions;252 new768x768 PNG; previous84 bytes and bindings retained')
