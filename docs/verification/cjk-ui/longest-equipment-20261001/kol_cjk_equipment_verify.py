from pathlib import Path
import json,csv,hashlib,sys,os,shutil
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-equipment-20261001')
selection=json.loads((root/'selection.json').read_text())
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath)
assert font.channels==4
gs={g['character']:g for g in fm['glyphs']}
names={r['en'].casefold():r['zh-Hant'] for r in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
labels={x['en']:x['zh'] for x in json.loads(Path('docs/i18n/newchar-zh-Hant.json').read_text())['ui']}
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
party=['Theodoric','Poontang','Olear','Moraine','Fighting Bull','Kenneth']
members={c['name']:c for c in roster if c and c['occupied'] and c['name'] in party}
assert len(members)==6
geometry=[48,6,363,225]
x0,y0,w,h=geometry
def width(t):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in t)
def expected(c):
    image=bytearray(w*h*3)
    for y in range(h):
        for x in range(w):
            if x in (0,w-1) or y in (0,h-1):
                pos=(y*w+x)*3;image[pos:pos+3]=b'\xff\xff\xff'
    lines=[];ink=[]
    def draw(x,y,text,safe_right=408):
        if not text:return
        assert x+width(text)<=safe_right and y+18<=y0+h-1,(text,x,width(text),safe_right)
        assert x>x0 and y>y0
        lines.append({'text':text,'position':[x,y],'width':width(text),'right':x+width(text),'bottom':y+18,'safe_right':safe_right})
        pen=x-x0;top=y-y0
        for char in text:
            if char not in gs:assert char==' ';pen+=fm['advance'];continue
            g=gs[char];sx=g['sheet_column']*16;sy=g['sheet_row']*18
            for yy in range(18):
                for xx in range(16):
                    alpha=font.pixels[((sy+yy)*font.width+sx+xx)*4+3]
                    assert alpha in (0,255)
                    if alpha:
                        pos=((top+yy)*w+pen+xx)*3;image[pos:pos+3]=b'\xff\xff\xff';ink.append(pos)
            pen+=g['advance']
    draw(72,21,c['name'].upper())
    eq=c['equipment']
    item_names=[(c['in_hand'] or {}).get('name','')]+[(eq[i] or {}).get('name','') for i in [0,1,2,6,9]]
    for i,(key,name) in enumerate(zip(['W:','H:','T:','L:','S:','B:'],item_names)):
        draw(72,72+18*i,labels[key],144)
        if name:draw(144,72+18*i,names[name.casefold()])
    if c['name']=='Olear':draw(72,180,labels['ARROWS:']+'20')
    return image,lines,ink
templates={name:expected(c) for name,c in members.items()}
checks=[]
for c in selection['checkpoints']:
    p=root/'zh-Hant-current'/f'step-{c["equipment"]:03d}.png'
    png=read_png(p);assert (png.width,png.height)==(960,600)
    actual=png.crop(*geometry).rgb()
    differences={name:sum(actual[i:i+3]!=tpl[0][i:i+3] for i in range(0,len(actual),3)) for name,tpl in templates.items()}
    matches=[name for name,n in differences.items() if n==0]
    assert len(matches)==1,(str(p),differences)
    name=matches[0];image,lines,ink=templates[name]
    assert ink
    bad=image.copy();bad[ink[0]:ink[0]+3]=b'\0\0\0'
    assert sum(actual[i:i+3]!=bad[i:i+3] for i in range(0,len(actual),3))==1
    # 以正式264像素安全寬故意縮成95，96像素最寬護甲必須遭拒。
    selected=[line for line in lines if line['text']==selection['selected']['zh']]
    if selected:assert any(line['width']>95 for line in selected)
    checks.append({'source_step':c['source_step'],'step':c['equipment'],'actor':name,'panel':geometry,'lines':lines,
                   'pixel_differences':0,'single_missing_ink_pixel_rejected':True,'wide_armour_present':bool(selected),
                   'before_step':c['before'],'restored_step':c['restored']})
assert set(c['actor'] for c in checks)==set(party)
assert sum(c['wide_armour_present'] for c in checks)==1
out=root/'pixel-checks.json'
out.write_text(json.dumps({'schema':'kol-cjk-equipment-pixel-check-v1','checks':checks,'selection':selection,'narrow_95_pixel_armour_rejected':True,
                         'independent_basis':'固定初始角色實際物品、外部譯表、正式字模與SPEC-019版面座標；未呼叫正式renderer。'},ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'panels':len(checks),'pixel_differences':0,'actors':[c['actor'] for c in checks],'maximum_armour':selection['selected'],'narrow_95_rejected':True},ensure_ascii=False),flush=True)
