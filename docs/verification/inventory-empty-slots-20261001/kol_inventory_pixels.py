from pathlib import Path
import json, os, hashlib, sys, shutil, csv
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/inventory-empty-slots-20261001/normal')
selection=json.loads((root/'selection.json').read_text());steps=selection['steps']
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
en=json.loads((root/'en/trace.json').read_text());zh=json.loads((root/'zh-Hant/trace.json').read_text())
assert en==zh and len(en['steps'])==199
re=json.loads((root/'en-reload2/trace.json').read_text());rz=json.loads((root/'zh-Hant-reload2/trace.json').read_text())
assert re==rz and len(re['steps'])==16
saves=[json.loads((root/lang/'save.json').read_text()) for lang in ('en','zh-Hant')]
for saved in saves:saved.pop('saved_at')
assert saves[0]==saves[1]
leader=saves[0]['characters'][saves[0]['party'][0]];good=leader['backpack'][0]
assert leader['gold']==918 and good['record']==selection['good']['record'] and good['name']=='Girak Helbak'
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath);gs={g['character']:g for g in fm['glyphs']}
bank=json.loads(Path('docs/i18n/ui-zh-Hant.json').read_text())['banks']['KOLDATA']
names={row['en'].casefold():row['zh-Hant'] for row in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
def width(s):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in s)
checks=[]
def box(x,y,w,h):
    pixels=bytearray(w*h*3)
    for yy in range(h):
        for xx in range(w):
            if xx in (0,w-1) or yy in (0,h-1):i=(yy*w+xx)*3;pixels[i:i+3]=b'\xff\xff\xff'
    return {'rect':[x,y,w,h],'pixels':pixels,'ink':[],'lines':[]}
def draw(b,x,y,text,colour=(255,255,255)):
    bx,by,bw,bh=b['rect'];pen=x-bx;top=y-by
    assert pen>=1 and pen+width(text)<=bw-1 and top>=1 and top+18<=bh-1,(text,b['rect'],x,y)
    b['lines'].append({'text':text,'position':[x,y],'width':width(text),'bottom':y+18})
    for c in text:
        if c not in gs:assert c==' ';pen+=fm['advance'];continue
        g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
        for yy in range(18):
            for xx in range(16):
                a=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert a in (0,255)
                if a:
                    i=((top+yy)*bw+pen+xx)*3;b['pixels'][i:i+3]=bytes(colour);b['ink'].append(i)
        pen+=g['advance']
def check(b,path,label):
    actual=read_png(path).crop(*b['rect']).rgb();expected=b['pixels']
    d=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert d==0,(label,d)
    i=b['ink'][0];wrong=expected.copy();wrong[i:i+3]=b'\0\0\0'
    assert sum(actual[j:j+3]!=wrong[j:j+3] for j in range(0,len(actual),3))==1
    checks.append({'label':label,'file':str(path),'rect':b['rect'],'lines':b['lines'],'pixel_differences':d,'single_missing_ink_pixel_rejected':True})
towns=json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns']
town=next(t for t in towns if t['town']==7)
shop_population=[{'town':t['town'],'door':d['index'],'good':g['index'],'name':g['name'],'zh':names[g['name'].casefold()],'width':width(names[g['name'].casefold()])} for t in towns for d in t['doors'] for g in d.get('goods',[])]
assert len(shop_population)==503 and max(r['width'] for r in shop_population)==112
assert [r for r in shop_population if r['width']==112]==[{'town':7,'door':13,'good':5,'name':'Girak Helbak','zh':'吉拉克赫爾巴克','width':112}]
door=next(d for d in town['doors'] if d['index']==13)
b=box(306,252,585,204)
for i,g in enumerate(door['goods']):draw(b,315,258+i*16,names[g['name'].casefold()],(255,255,255) if i==5 else (0,170,170))
check(b,root/'zh-Hant'/f'step-{steps["selected"]:03}.png','商品清單')
def value(n):return bank[str(n)]['zh'].replace('@','').strip()
segments=[(20,102,names[good['name'].casefold()]),(16,118,value(163)+value(164)),(16,126,value(173)+'1-3'),(26,126,value(174)+'0'),(16,134,value(175)+'35'),(26,134,value(176)+value(181))]
b=box(306,252,585,204)
draw(b,315,258,value(49))
for col,row,text in segments:draw(b,col*24,row*3,text)
check(b,root/'zh-Hant'/f'step-{steps["shop_detail"]:03}.png','商品明細')
marky=json.loads(Path('assets/original-derived-data/portrait-layout.json').read_text())['slot_marks'][3][11]*3
def mark(path,label):
    b=box(18,marky,300,66);draw(b,24,marky+6,value(153)+':');draw(b,24,marky+24,names[good['name'].casefold()]);draw(b,24,marky+42,'1-3');draw(b,318-6-width('0'),marky+42,'0');check(b,path,label)
def detail(path,label):
    b=box(186,174,588,204)
    for col,row,text in segments:draw(b,192+(col*8-104)*3,150+(row-86)*3,text)
    check(b,path,label)
mark(root/'zh-Hant'/f'step-{steps["pocket"]:03}.png','背包小框')
detail(root/'zh-Hant'/f'step-{steps["item_detail"]:03}.png','背包明細')
mark(root/'zh-Hant-reload2/step-014.png','重載背包小框')
detail(root/'zh-Hant-reload2/step-015.png','重載背包明細')
assert checks[2]['lines'][1]['width']==112 and checks[4]['lines'][1]['width']==112
pictures=[];sources=[]
for step,index,xyen,xyzh in [(steps['door'],29,(104,58),(312,174)),(steps['entry'],28,(8,64),(24,192))]:
    source=Path(f'assets/original-derived-png/ega/TWN7PICS/twn7pics-{index:03}.png');pic=read_png(source);raw=pic.rgb();sources.append(source)
    for lang,scale,xy in [('en',1,xyen),('zh-Hant',3,xyzh)]:
        expected=bytearray()
        for yy in range(pic.height*scale):
            for xx in range(pic.width*scale):i=((yy//scale)*pic.width+xx//scale)*3;expected+=raw[i:i+3]
        actual=read_png(root/lang/f'step-{step:03}.png').crop(*xy,pic.width*scale,pic.height*scale).rgb();assert actual==expected
        pictures.append({'step':step,'language':lang,'source':str(source),'scale':scale,'pixel_differences':0})

(root.parent / 'pixel-checks.json').write_text(json.dumps({'checks':checks,'pictures':pictures},ensure_ascii=False,indent=2)+'\n')
print('六個文字框與四份圖片均零像素差')
