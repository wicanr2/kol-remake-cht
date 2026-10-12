#!/usr/bin/env python3
"""Docker 內由鎖定 Noto 字型烘製原型用的原生 HD 字模，不更改正式字型。"""
from pathlib import Path
import hashlib,json
from PIL import Image,ImageDraw,ImageFont

assert Path('/.dockerenv').exists()
out=Path('/out')
source=Path('/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc')
fonts=[ImageFont.truetype(str(source),64,index=i) for i in range(5)]
index=next(i for i,f in enumerate(fonts) if 'TC' in f.getname()[0])
font=fonts[index]
chars=sorted(set((out/'review.js').read_text())|set('…0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'))
chars=[c for c in chars if c.isprintable() or c in (' ','　')]
atlas=Image.new('RGBA',(1536,((len(chars)+15)//16)*96),(0,0,0,0))
draw=ImageDraw.Draw(atlas);glyphs={}
for i,ch in enumerate(chars):
    x,y=(i%16)*96,(i//16)*96
    draw.text((x+10,y+76),ch,font=font,fill='white',anchor='ls')
    glyphs[ch]={'x':x,'y':y,'advance':font.getlength(ch)}
atlas.save(out/'font-atlas.png')
license=Path('/usr/share/doc/fonts-noto-cjk/copyright')
assert license.is_file()
(out/'font-LICENSE.txt').write_bytes(license.read_bytes())
data={'schema':'kol-hd-review-font/1','source':str(source),'source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),'face':font.getname(),'face_index':index,'size':64,'cell':96,'baseline':76,'glyphs':glyphs,'license':'font-LICENSE.txt','scope':'prototype only; not the production font inventory'}
(out/'font-atlas.json').write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
print('FONT PASS',font.getname(),len(chars),atlas.size)
