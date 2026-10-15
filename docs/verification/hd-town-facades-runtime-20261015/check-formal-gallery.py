from pathlib import Path
from PIL import Image, ImageChops
import hashlib
import json

w = Path('/out')
r = Path('/src')
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
current = json.loads((w / 'formal-gallery-captures.json').read_text())
candidate = json.loads((w / 'full-group-v2-captures.json').read_text())
sign = json.loads((w / 'sign-native-v3-captures.json').read_text())
group = lambda frames: {key: {f['Variant']: f for f in frames if f['Key'] == key} for key in {f['Key'] for f in frames}}
cases, old, signed = group(current['frames']), group(candidate['frames']), group(sign['frames'])
assert len(cases) == 103 and len(current['frames']) == 412
manifest = json.loads((r / 'assets/hd/first-slice/town-facades/manifest.json').read_text())
art = {p['file'][:-4]: p for p in manifest['pictures']}
records = []
for key, frames in sorted(cases.items()):
    reference = old[key]
    for variant, f in frames.items():
        assert sha(w / 'formal-gallery-gallery' / f['PNG']) == f['SHA']
        assert f['StateSHA'] == reference['baseline-ega']['StateSHA']
        assert f['Text'] == reference['baseline-ega']['Text'] and f['Han'] > 0
    images = {v: Image.open(w / 'formal-gallery-gallery' / f['PNG']).convert('RGBA') for v, f in frames.items()}
    old_ega = Image.open(w / 'full-group-v2-gallery' / reference['baseline-ega']['PNG']).convert('RGBA')
    assert images['baseline-ega'].tobytes() == images['prototype-ega'].tobytes() == old_ega.tobytes()
    path = w / 'full-group-v2-gallery' / reference['prototype-hd']['PNG']
    if key == 'twn5pics-019':
        path = w / 'sign-native-v3-gallery' / signed[key]['prototype-hd']['PNG']
    expected = Image.open(path).convert('RGBA')
    delta = ImageChops.difference(expected, images['prototype-hd'])
    full_max = max(q[1] for q in delta.getextrema())
    assert full_max <= 1, (key, 'full HD versus candidate', full_max)
    repeat = ImageChops.difference(images['baseline-hd'], images['prototype-hd'])
    assert max(q[1] for q in repeat.getextrema()) == 0
    old_hd = Image.open(w / 'full-group-v2-gallery' / reference['baseline-hd']['PNG']).convert('RGBA')
    outside = ImageChops.difference(old_hd, images['prototype-hd'])
    outside.paste((0, 0, 0, 0), tuple(current['native_rect']))
    assert max(q[1] for q in outside.getextrema()) == 0
    raw = Image.open(r / 'assets/hd/first-slice/town-facades' / art[key]['file']).convert('RGBA')
    sampled = raw.transform((1152, 576), Image.Transform.AFFINE,
                            (raw.width/1152,0,0,0,raw.height/576,0),Image.Resampling.BILINEAR)
    native = ImageChops.difference(sampled, images['prototype-hd'].crop(tuple(current['native_rect'])))
    if key == 'twn5pics-019':
        for b in json.loads((w / 'residence-sign-hd.json').read_text())['hd_clip']:
            native.paste((0,0,0,0),(b[0]-624,b[1]-348,b[2]-624,b[3]-348))
    native_max = max(q[1] for q in native.getextrema())
    assert native_max <= 1, (key, 'native painting', native_max)
    records.append(dict(key=key, native_painting_max=native_max, full_candidate_hd_max=full_max,
                        outside_max=0, ega_rgba_exact=True, complete_serializable_Outdoor_equal=True,
                        chinese_equal=True, han_characters=frames['prototype-hd']['Han']))
font = json.loads((r / 'assets/hd/first-slice/interface/font-atlas.json').read_text())
chars = set(c for f in current['frames'] for text in f['Text'] for c in text if '\u3400' <= c <= '\u9fff')
missing = sorted(chars - set(font['glyphs']))
assert not missing
result = dict(status='PASS limited formal103facade/edge scenes', scenes=103, frames=412,
              native_painting_max=max(q['native_painting_max'] for q in records),
              full_candidate_hd_max=max(q['full_candidate_hd_max'] for q in records), outside_max=0,
              ega_rgba_exact=True, complete_serializable_Outdoor_equal=True, chinese_equal=True,
              actual_han_characters=sum(f['Han'] for f in current['frames']), unique_han=len(chars), missing_glyphs=missing,
              records=records, source_inputs_sha256=sha(w/'formal-source-inputs.json'),
              verification_source_inputs_sha256=sha(w/'formal-verification-source-inputs.json'),
              assets_inputs_sha256=sha(w/'formal-asset-inputs.json'), captures_sha256=sha(w/'formal-gallery-captures.json'),
              residence_note='Whole HD matches the native32pixel candidate within1; painting-only CPU comparison excludes the three native text faces. Optional independent font-filter model remains unresolved and is not used as parity proof.',
              normal_player_path=False, original_oracle=False, platform='Linux Docker only')
(w/'formal-gallery-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k!='records'},ensure_ascii=False),flush=True)
