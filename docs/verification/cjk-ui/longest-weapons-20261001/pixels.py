from pathlib import Path
import json,csv,sys
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-weapon-20261001')
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
font=read_png(Path('assets/ui-font')/fm['file'])
gs={g['character']:g for g in fm['glyphs']}
bank=json.loads(Path('docs/i18n/ui-zh-Hant.json').read_text())['banks']['KOLDATA']
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
marks=json.loads(Path('assets/original-derived-data/portrait-layout.json').read_text())['slot_marks']
names={r['en'].casefold():r['zh-Hant'] for r in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
labels={x['en']:x['zh'] for x in json.loads(Path('docs/i18n/newchar-zh-Hant.json').read_text())['ui']}
def width(t):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in t)
def value(i):return bank[str(i)]['zh'].replace('@','').strip()
def box(x,y,w,h):
    p=bytearray(w*h*3)
    for yy in range(h):
        for xx in range(w):
            if xx in (0,w-1) or yy in (0,h-1):
                i=(yy*w+xx)*3;p[i:i+3]=b'\xff\xff\xff'
    return {'rect':[x,y,w,h],'pixels':p,'ink':[],'lines':[]}
def draw(b,x,y,t):
    if not t:return
    bx,by,bw,bh=b['rect'];pen=x-bx;top=y-by
    assert 0<pen and pen+width(t)<=bw-1 and 0<top and top+18<=bh-1,(t,b['rect'])
    b['lines'].append({'text':t,'x':x,'y':y,'width':width(t)})
    for c in t:
        if c not in gs:assert c==' ';pen+=fm['advance'];continue
        g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
        for yy in range(18):
            for xx in range(16):
                a=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert a in (0,255)
                if a:
                    i=((top+yy)*bw+pen+xx)*3;b['pixels'][i:i+3]=b'\xff\xff\xff';b['ink'].append(i)
        pen+=g['advance']
checks=[];frames=[]
def check(b,path,label):
    png=read_png(path);assert (png.width,png.height)==(960,600)
    actual=png.crop(*b['rect']).rgb();expected=b['pixels']
    d=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert d==0,(label,d)
    bad=expected.copy();i=b['ink'][0];bad[i:i+3]=b'\0\0\0'
    assert sum(actual[i:i+3]!=bad[i:i+3] for i in range(0,len(actual),3))==1
    checks.append({'label':label,'file':str(path),'rect':b['rect'],'lines':b['lines'],
                   'pixel_differences':0,'missing_ink_pixel_rejected':True})
def same(a,b):
    pa=read_png(a);pb=read_png(b)
    assert (pa.width,pa.height)==(pb.width,pb.height) and pa.rgb()==pb.rgb(),(a,b)
    frames.append({'a':str(a),'b':str(b),'pixel_differences':0})
good=json.loads((root/'player-plan.json').read_text())['good'];c=roster[4]
assert c['name']=='Fighting Bull' and good['name']=='War Maul'
text=names[good['name'].casefold()];assert width(text)==64 and width(text)>63
def mark(path,row,x,label):
    y=marks[c['sheet']['race']][row]*3;b=box(x,y,300,66)
    draw(b,x+6,y+6,value(142+min(row,11))+':');draw(b,x+6,y+24,text)
    draw(b,x+6,y+42,f"{good['dice']['low']}-{good['dice']['high']}")
    v=str(good['encumbrance']);draw(b,x+294-width(v),y+42,v);check(b,path,label)
def detail(path,label):
    b=box(186,174,588,204);namex=(14+(24-len(good['name']))//2)*24-120
    draw(b,namex,198,text);draw(b,264,246,value(163)+value(164+good['kind']))
    draw(b,504,246,value(171)+value(187+good['position']))
    draw(b,264,270,value(172)+f"{good['dice']['low']}-{good['dice']['high']}")
    draw(b,504,270,value(174)+str(good['encumbrance']));draw(b,264,294,value(175)+str(good['price']))
    draw(b,504,294,value(177)+value(183+good['condition']//64));check(b,path,label)
normal=root/'player-zh-Hant';reload=root/'player-zh-Hant-reload'
mark(normal/'step-076.png',13,18,'正常背包標記');detail(normal/'step-077.png','正常背包明細')
mark(normal/'step-082.png',10,330,'正常手持標記');detail(normal/'step-083.png','正常手持明細')
mark(reload/'step-013.png',10,330,'重載手持標記');detail(reload/'step-014.png','重載手持明細')
same(normal/'step-082.png',reload/'step-013.png');same(normal/'step-083.png',reload/'step-014.png')
saved=json.loads((root/'player-zh-Hant-save.json').read_text())
fighter=next(x for x in saved['characters'] if x['name']=='Fighting Bull')
def panel(path,who,items,label):
    b=box(48,6,363,225)
    if who:draw(b,72,21,who.upper())
    for i,name in enumerate(items):
        draw(b,72,72+18*i,labels[['W:','H:','T:','L:','S:','B:'][i]])
        if name:draw(b,144,72+18*i,names[name.casefold()])
    check(b,path,label)
items=[fighter['in_hand']['name']]+[(fighter['worn'][i] or {}).get('name','') for i in [0,1,2,6,9]]
panel(normal/'step-188.png','Fighting Bull',items,'玩家戰鬥武器欄')
monster=json.loads(Path('assets/original-derived-data/monsters.json').read_text())['monsters'][19]
assert monster['name']=='BINDERAK' and monster['weapon']['name']=='Spiney Mace'
assert width(names['spiney mace'])==80 and width(names['spiney mace'])>79
cp=json.loads((root/'enemy-checkpoint.json').read_text())
enemyitems=[monster['weapon']['name']]+[monster['worn'][i]['name'] for i in [0,1,2]]+[(monster['shield'] or {}).get('name','')]
# SPEC-019 怪物裝備面板沒有角色姓名列。
panel(root/'enemy-zh-Hant'/f"step-{cp['equipment']:03}.png",'',enemyitems,'最長敵方戰鬥武器欄')
for lang in ['en','zh-Hant']:
    d=root/f'player-{lang}';same(d/'step-187.png',d/'step-189.png')
    d=root/f'enemy-{lang}';same(d/f"step-{cp['before']:03}.png",d/f"step-{cp['restored']:03}.png")
(root/'pixel-checks.json').write_text(json.dumps({'checks':checks,'whole_frame_comparisons':frames,
    'narrow_63_player_and_79_enemy_widths_rejected':True,
    'basis':'固定來源物品、外部譯表、正式字模與 SPEC-019／SPEC-033 座標；未呼叫正式 renderer。'},ensure_ascii=False,indent=2)+'\n')
print('八個完整文字框、六個全畫格比較均為零差分；縮窄字寬與缺一墨點的負向檢查遭拒。')
