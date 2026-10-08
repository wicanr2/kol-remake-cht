from pathlib import Path
import json,csv,sys
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-held-names-20261001')
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
font=read_png(Path('assets/ui-font')/fm['file'])
gs={g['character']:g for g in fm['glyphs']}
bank=json.loads(Path('docs/i18n/ui-zh-Hant.json').read_text())['banks']['KOLDATA']
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
marks=json.loads(Path('assets/original-derived-data/portrait-layout.json').read_text())['slot_marks']
names={r['en']:r['zh-Hant'] for r in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
def width(text):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in text)
def value(index):return bank[str(index)]['zh'].replace('@','').strip()
def box(x,y,w,h):
    pixels=bytearray(w*h*3)
    for yy in range(h):
        for xx in range(w):
            if xx in (0,w-1) or yy in (0,h-1):
                i=(yy*w+xx)*3;pixels[i:i+3]=b'\xff\xff\xff'
    return {'rect':[x,y,w,h],'pixels':pixels,'ink':[],'lines':[]}
def draw(b,x,y,text):
    bx,by,bw,bh=b['rect'];pen=x-bx;top=y-by
    assert pen>=1 and pen+width(text)<=bw-1 and top>=1 and top+18<=bh-1,(text,b['rect'],x,y)
    b['lines'].append({'text':text,'x':x,'y':y,'width':width(text)})
    for c in text:
        if c not in gs:assert c==' ';pen+=fm['advance'];continue
        g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
        for yy in range(18):
            for xx in range(16):
                alpha=font.pixels[((sy+yy)*font.width+sx+xx)*4+3]
                assert alpha in (0,255)
                if alpha:
                    i=((top+yy)*bw+pen+xx)*3;b['pixels'][i:i+3]=b'\xff\xff\xff';b['ink'].append(i)
        pen+=g['advance']
checks=[];diffs=[]
def check(b,path,label):
    actual=read_png(path).crop(*b['rect']).rgb()
    d=sum(actual[i:i+3]!=b['pixels'][i:i+3] for i in range(0,len(actual),3))
    assert d==0,(label,d)
    wrong=b['pixels'].copy();i=b['ink'][0];wrong[i:i+3]=b'\0\0\0'
    assert sum(actual[i:i+3]!=wrong[i:i+3] for i in range(0,len(actual),3))==1
    checks.append({'label':label,'file':str(path),'rect':b['rect'],'lines':b['lines'],
                   'pixel_differences':0,'missing_ink_pixel_rejected':True})
def differences(before,after,allowed):
    a=read_png(before);b=read_png(after);assert (a.width,a.height)==(b.width,b.height)
    ar=a.rgb();br=b.rgb(); count=0
    for i in range(0,len(ar),3):
        if ar[i:i+3]==br[i:i+3]:continue
        count+=1;x=(i//3)%a.width;y=(i//3)//a.width
        assert any(rx<=x<rx+rw and ry<=y<ry+rh for rx,ry,rw,rh in allowed),(before,x,y)
    diffs.append({'before':str(before),'after':str(after),'allowed_rects':allowed,
                  'changed_pixels':count,'outside_allowed_pixels':0})
for kind,slot,row,field,itemindex,markstep,detailstep,rmark,rdetail in [
    ('maul',2,15,'backpack',4,78,79,18,19),
    ('warder',8,7,'equipment',7,25,26,10,11)]:
    character=roster[slot];good=character[field][itemindex];label=names[good['name']]
    assert label==('大槌' if kind=='maul' else '黑色頸鍊')
    y=marks[character['sheet']['race']][row]*3
    col=14+(24-len(good['name']))//2;namex=col*24-120
    def mark(path,tag):
        b=box(18,y,300,66);draw(b,24,y+6,value(142+min(row,11))+':')
        draw(b,24,y+24,label);draw(b,24,y+42,f"{good['dice']['low']}-{good['dice']['high']}")
        w=str(good['encumbrance']);draw(b,312-width(w),y+42,w);check(b,path,tag)
    def detail(path,tag):
        b=box(186,174,588,204);draw(b,namex,198,label)
        draw(b,264,246,value(163)+value(164+good['kind']))
        draw(b,504,246,value(171)+value(187+good['position']))
        draw(b,264,270,value(172 if good['damage'] else 173)+f"{good['dice']['low']}-{good['dice']['high']}")
        draw(b,504,270,value(174)+str(good['encumbrance']))
        draw(b,264,294,value(175)+str(good['price']))
        draw(b,504,294,value(177)+value(183+good['condition']//64));check(b,path,tag)
    suffix='-v2' if kind=='warder' else ''
    after=root/(f'after-zh-Hant-{kind}-save'+suffix)
    reload=root/(f'after-zh-Hant-{kind}-reload'+suffix)
    mark(after/f'step-{markstep:03}.png',kind+'正常標記框')
    detail(after/f'step-{detailstep:03}.png',kind+'正常明細框')
    mark(reload/f'step-{rmark:03}.png',kind+'重載標記框')
    detail(reload/f'step-{rdetail:03}.png',kind+'重載明細框')
    for language in ['en','zh-Hant']:
        before=root/(f'before-{language}-{kind}-save'+suffix)
        current=root/(f'after-{language}-{kind}-save'+suffix)
        for oldpath in before.glob('*.png'):
            step=int(oldpath.stem.split('-')[1]);allowed=[]
            if language=='zh-Hant' and step in (markstep,detailstep):
                x,yy=(24,y+24) if step==markstep else (namex,198)
                allowed=[[x,yy,max(width(good['name'].upper()),width(label)),18]]
            differences(oldpath,current/oldpath.name,allowed)
            if allowed:assert diffs[-1]['changed_pixels']>0
    for a,b in [(markstep,rmark),(detailstep,rdetail)]:
        differences(after/f'step-{a:03}.png',reload/f'step-{b:03}.png',[])
(root/'pixel-checks.json').write_text(json.dumps({'checks':checks,'whole_frame_comparisons':diffs},ensure_ascii=False,indent=2)+'\n')
print('八個完整文字框符合字模；英中前後及重載畫格差分限於四個譯名區域。')
