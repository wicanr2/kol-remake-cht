#!/usr/bin/env python3
"""Docker限定：將固定粉圓2.1字型離線轉為工具列PNG字模。"""
import argparse
import hashlib
import json
import os
import math
import re
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--root', type=Path, required=True)
p.add_argument('--font', type=Path, required=True)
p.add_argument('--license', type=Path, required=True)
a = p.parse_args()
assert Path('/.dockerenv').exists()
sha = lambda b: hashlib.sha256(b).hexdigest()
assert sha(a.font.read_bytes()) == '9d5bf4932d31fe94c18cd8cfddc98bc1b14ce10f4e354c682179db290a99c825'
assert sha(a.license.read_bytes()) == '5df6de36ba2fe26119ac161edbed97c62a9da3c1493f6930b9d9af7f5c93004b'
out = a.root/'cmd/kol/toolbarfont'
assert out.parent.stat().st_uid == os.getuid()
out.mkdir(exist_ok=True)
chars = set(chr(i) for i in range(32,127))
for name in ('cmd/kol/toolbar.go','cmd/kol/game_font.go','internal/town/assistance.go'):
    for literal in re.findall(r'"(?:[^"\\]|\\.)*"',(a.root/name).read_text()):
        chars.update(c for c in json.loads(literal) if ord(c) >= 0x3000)
chars.add('…')
chars = sorted(chars)
size = 20
cols = 32
rows = (len(chars)+cols-1)//cols
img = Image.new('RGBA',(cols*size,rows*size))
draw = ImageDraw.Draw(img)
face = ImageFont.truetype(str(a.font),size)
glyphs=[]
for i,c in enumerate(chars):
    x,y = (i%cols)*size,(i//cols)*size
    advance = max(1,math.ceil(face.getlength(c))) if ord(c)<127 else size
    cell = Image.new('RGBA',(size,size))
    pen = ImageDraw.Draw(cell)
    box = face.getbbox(c)
    pen.text((0,(size-(box[3]-box[1]))//2-box[1]),c,font=face,fill='white')
    img.alpha_composite(cell,(x,y))
    glyphs.append(dict(codepoint=ord(c),sheet_column=i%cols,sheet_row=i//cols,advance=advance))
img.save(out/'toolbar.png')
manifest = dict(schema='kol-font-manifest-v2',fonts=[dict(name='toolbar',file='toolbar.png',advance=size,first_codepoint=0,glyph_count=len(chars),cell=dict(width=size,height=size),sheet=dict(columns=cols,width=img.width,height=img.height),glyphs=glyphs)])
(out/'manifest.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
(out/'OFL.txt').write_bytes(a.license.read_bytes())
source = dict(name='jf-openhuninn-2.1',url='https://github.com/justfont/open-huninn-font/releases/download/v2.1/jf-openhuninn-2.1.ttf',license='OFL-1.1',source_sha256=sha(a.font.read_bytes()),license_sha256=sha(a.license.read_bytes()),png_sha256=sha((out/'toolbar.png').read_bytes()),generator='tools/bake_toolbar_font.py',runtime='PNG only; embedded in executable',glyph_count=len(chars))
(out/'source.json').write_text(json.dumps(source,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(source,ensure_ascii=False))
