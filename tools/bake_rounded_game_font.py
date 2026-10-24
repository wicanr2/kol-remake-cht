#!/usr/bin/env python3
"""Docker限定：由固定粉圓2.1及原中文manifest製作雙尺寸遊戲PNG字模。"""
import argparse
import hashlib
import json
import os
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--root', type=Path, required=True)
p.add_argument('--font', type=Path, required=True)
p.add_argument('--out', type=Path, required=True)
a = p.parse_args()
assert Path('/.dockerenv').exists()
sha = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
assert sha(a.font) == '9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825'
assert a.out.parent.stat().st_uid == os.getuid()
a.out.mkdir(exist_ok=True)
base_path = a.root / 'assets/ui-font/manifest.json'
base = json.loads(base_path.read_text())['fonts'][0]
entries, outputs, missing = [], {}, []
for scale in (1, 2):
    size, cw, height = 16 * scale, 16 * scale, 18 * scale
    face = ImageFont.truetype(str(a.font), size * 4)
    asc, _ = face.getmetrics()
    notdef = bytes(face.getmask('\U0010ffff'))
    chars = [g for g in base['glyphs'] if g['codepoint'] >= 0x2e80]
    glyphs = []
    for g in chars:
        if bytes(face.getmask(chr(g['codepoint']))) == notdef:
            if scale == 1:
                missing.append(g['codepoint'])
            continue
        glyphs.append(g)
    columns = 32
    sheet = Image.new('RGBA', (columns * cw, ((len(glyphs) + columns - 1) // columns) * height))
    mapping = []
    for i, g in enumerate(glyphs):
        ch = chr(g['codepoint'])
        tile = Image.new('L', (cw * 4, height * 4))
        ImageDraw.Draw(tile).text((0, (height - size) * 2 + size * 4 - asc), ch, font=face, fill=255)
        alpha = tile.resize((cw, height), Image.Resampling.LANCZOS)
        rgba = Image.new('RGBA', (cw, height), 'white')
        rgba.putalpha(alpha)
        sheet.alpha_composite(rgba, ((i % columns) * cw, (i // columns) * height))
        mapping.append(dict(codepoint=g['codepoint'], sheet_column=i % columns, sheet_row=i // columns, advance=g['advance'] * scale))
    name = f'rounded-{size}'
    file = name + '.png'
    sheet.save(a.out / file)
    entries.append(dict(name=name, file=file, advance=base['advance'] * scale, first_codepoint=0,
                        glyph_count=len(mapping), cell=dict(width=cw, height=height),
                        sheet=dict(columns=columns, width=sheet.width, height=sheet.height), glyphs=mapping))
    outputs[file] = sha(a.out / file)
(a.out / 'manifest.json').write_text(json.dumps(dict(schema='kol-font-manifest-v2', fonts=entries), ensure_ascii=False, indent=2) + '\n')
outputs['manifest.json'] = sha(a.out / 'manifest.json')
record = dict(font='jf open 粉圓 2.1', source_sha256=sha(a.font),
              source_url='https://github.com/justfont/open-huninn-font/releases/download/v2.1/jf-openhuninn-2.1.ttf',
              license='OFL-1.1', license_path='cmd/kol/toolbarfont/OFL.txt',
              license_sha256=sha(a.root / 'cmd/kol/toolbarfont/OFL.txt'),
              original_manifest_sha256=sha(base_path), generator='tools/bake_rounded_game_font.py',
              generator_sha256=sha(Path(__file__)), outputs=outputs,
              method='4x supersampling; fixed original baseline/advance; grayscale alpha; PNG only at runtime',
              fallback_codepoints=missing, glyph_count=entries[0]['glyph_count'])
(a.out / 'source.json').write_text(json.dumps(record, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(record, ensure_ascii=False))
