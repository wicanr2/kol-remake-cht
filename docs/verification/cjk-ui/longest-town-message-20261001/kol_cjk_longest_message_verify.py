from pathlib import Path
import json, hashlib, os, sys, shutil
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-longest-message-20261001');selection=json.loads((root/'selection.json').read_text())
en=json.loads((root/'en/trace.json').read_text());zh=json.loads((root/'zh-Hant/trace.json').read_text());assert en==zh and len(zh['steps'])==95
lines=[r['zh'] for r in selection['selected']]
assert lines==json.loads(Path('docs/i18n/town-zh-Hant.json').read_text())['towns']['7']['doors']['17']['messages']['17']
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0];glyphs={g['character']:g for g in fm['glyphs']}
fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath);assert font.channels==4
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def width(s):return sum(glyphs[c]['advance'] if c in glyphs else fm['advance'] for c in s)
x0,y0,w,h=306,252,585,204;expected=bytearray(w*h*3)
for y in range(h):
    for x in range(w):
        if x in (0,w-1) or y in (0,h-1):i=(y*w+x)*3;expected[i:i+3]=b'\xff\xff\xff'
checks=[];ink=[]
for row,line in enumerate(lines):
    pen=315-x0;top=258-y0+row*24;assert width(line)<=575 and top+18<h-1
    for char in line:
        assert char in glyphs,char
        g=glyphs[char];sx=g['sheet_column']*16;sy=g['sheet_row']*18
        for yy in range(18):
            for xx in range(16):
                a=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert a in (0,255)
                if a:
                    i=((top+yy)*w+pen+xx)*3;assert expected[i:i+3]==b'\0\0\0';expected[i:i+3]=b'\xff\xff\xff';ink.append(i)
        pen+=g['advance']
    checks.append({'text':line,'width':width(line),'start':[315,258+row*24],'bottom':258+row*24+18})
actual=read_png(root/'zh-Hant/step-093.png').crop(x0,y0,w,h).rgb()
differences=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert differences==0,differences
corrupt=expected.copy();corrupt[ink[0]:ink[0]+3]=b'\0\0\0'
assert sum(actual[i:i+3]!=corrupt[i:i+3] for i in range(0,len(actual),3))==1
picture_checks=[];source_paths=[]
for step,index,start_en,start_zh in [(90,37,(104,58),(312,174)),(91,36,(8,64),(24,192)),(92,36,(8,64),(24,192)),(93,36,(8,64),(24,192))]:
    source=Path(f'assets/original-derived-png/ega/TWN7PICS/twn7pics-{index:03}.png');pic=read_png(source)
    if source not in source_paths:source_paths.append(source)
    for lang,scale,xy in [('en',1,start_en),('zh-Hant',3,start_zh)]:
        raw=pic.rgb();scaled=bytearray()
        for y in range(pic.height*scale):
            for x in range(pic.width*scale):i=((y//scale)*pic.width+x//scale)*3;scaled+=raw[i:i+3]
        actualpic=read_png(root/lang/f'step-{step:03}.png').crop(*xy,pic.width*scale,pic.height*scale).rgb()
        d=sum(actualpic[i:i+3]!=scaled[i:i+3] for i in range(0,len(scaled),3));assert d==0,(source,lang,step,d)
        picture_checks.append({'step':step,'language':lang,'source':str(source),'scale':scale,'position':xy,'pixel_differences':d})
target=Path('docs/verification/cjk-ui/longest-town-message-20261001');assert not target.exists()
assert (target.parent.stat().st_uid,target.parent.stat().st_gid)==(os.getuid(),os.getgid());target.mkdir();files={}
def copy(p,name):shutil.copyfile(p,target/name);files[name]=sha(target/name)
for name in ('selection.json','script.txt'):copy(root/name,name)
for lang in ('en','zh-Hant'):
    for name in ['trace.json','run.json','engine.log']+[f'step-{s:03}.png' for s in selection['shot_steps']]:copy(root/lang/name,f'{lang}-{name}')
for name in ('kol_cjk_longest_message_run.py','kol_cjk_longest_message_verify.py'):copy(Path('/control')/name,name)
run=json.loads((root/'en/run.json').read_text());assert run['returncode']==0
r={'schema':'kol-cjk-longest-town-message-v1','date':'2026-10-01','status':'normal-player-path-pixel-and-visual-confirmed',
    'scope':selection['scope'],'selection':selection,'selected_before_execution':True,'identical_en_zh_states':95,'binary_sha256':run['binary_sha256'],
    'seed':{'remake':1,'original':None,'method':'執行前 -seed 1'},'geometry':{'dialog':[x0,y0,w,h],'text_start':[315,258],'safe_width':575,'line_step':24,'font_height':18},
    'geometry_evidence':'internal/town/layout.go CJKInn 及 internal/ui/town.go InnCJK。',
    'normal_path_evidence':'標題→新隊伍→正常六人招募→城內走路→門口→進門敘述→問誰開口→選門離店→訊息17→確認返回原門格；internal/town/session.go leaveDoor 的多人告別規則。',
    'line_checks':checks,'dialog_pixel_differences':differences,'single_missing_ink_pixel_rejected':True,'picture_checks':picture_checks,
    'visual_review':{'status':'passed','method':'主代理檢視五張中文完整畫格與英文告別畫格；告別兩行完整可讀、無缺字裁切，確認可返回原門格。'},
    'inputs_sha256':{str(p):sha(p) for p in source_paths+[fontpath,Path('assets/ui-font/manifest.json'),Path('docs/i18n/town-zh-Hant.json'),Path('docs/verification/cjk-text-fit.json'),Path('internal/town/layout.go'),Path('internal/town/session.go'),Path('internal/ui/town.go'),Path('assets/original-derived-data/town-buildings.json'),Path('assets/original-derived-png/ega/TOWN7/town7-map.png')]},
    'toolchain':'eob-remake-go:1.26.7-ebiten2.9.9-audio；sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    'rights':'重製執行畫面含原作衍生 PNG，沿用既有 private GitHub PNG 特例；沒有公開散布授權。',
    'limitations':'只驗正式 town-message 全母體最寬之一；另外三則同寬訊息、其他欄位、最長輸入、原版與平台封包未驗。#34／#40 不結案。',
    'files':files}
(target/'receipt.json').write_text(json.dumps(r,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'receipt':str(target/'receipt.json'),'states':95,'widths':[x['width'] for x in checks],'dialog_pixels':differences,'picture_checks':len(picture_checks),'files':len(files)},ensure_ascii=False))
