from pathlib import Path
import json,hashlib,os,sys,shutil
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-longest-title-20261001');selection=json.loads((root/'selection.json').read_text());comparison=json.loads((root/'comparison.json').read_text())
assert comparison['states']==3091 and comparison['identical_mother_prefix_states']==3081 and comparison['other_differences']==0
town=json.loads(Path('docs/i18n/town-zh-Hant.json').read_text())['towns']['5']['doors']['9'];lines=[town['name'],town['person']]
assert lines[1]==selection['selected']['zh']
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0];gs={g['character']:g for g in fm['glyphs']};fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath)
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def width(s):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in s)
w,h=585,60;expected=bytearray(w*h*3);ink=[];line_checks=[]
for y in range(h):
    for x in range(w):
        if x in (0,w-1) or y in (0,h-1):i=(y*w+x)*3;expected[i:i+3]=b'\xff\xff\xff'
for row,line in enumerate(lines):
    tx=306+(585-width(line))//2;ty=175+row*24;pen=tx-306;top=ty-174
    assert 307<=tx and tx+width(line)<=890 and ty+18<=233
    for c in line:
        assert c in gs,c
        g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
        for y in range(18):
            for x in range(16):
                a=font.pixels[((sy+y)*font.width+sx+x)*4+3];assert a in (0,255)
                if a:i=((top+y)*w+pen+x)*3;expected[i:i+3]=b'\xff\xff\xff';ink.append(i)
        pen+=g['advance']
    line_checks.append({'text':line,'width':width(line),'start':[tx,ty],'right':tx+width(line),'bottom':ty+18})
frame=read_png(root/'zh-Hant'/f'step-{selection["inside_step"]:03}.png');actual=frame.crop(306,174,w,h).rgb()
d=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert d==0,d
wrong=expected.copy();wrong[ink[0]:ink[0]+3]=b'\0\0\0';assert sum(actual[i:i+3]!=wrong[i:i+3] for i in range(0,len(actual),3))==1
picture_checks=[];sources=[]
for step,index,xy_en,xy_zh in [(selection['door_step'],selection['picture'],(104,58),(312,174)),(selection['inside_step'],selection['portrait'],(8,64),(24,192))]:
    source=Path(f'assets/original-derived-png/ega/TWN5PICS/twn5pics-{index:03}.png');pic=read_png(source);raw=pic.rgb();sources.append(source)
    for lang,scale,xy in [('en',1,xy_en),('zh-Hant',3,xy_zh)]:
        target_pixels=bytearray()
        for y in range(pic.height*scale):
            for x in range(pic.width*scale):i=((y//scale)*pic.width+x//scale)*3;target_pixels+=raw[i:i+3]
        actual_pixels=read_png(root/lang/f'step-{step:03}.png').crop(*xy,pic.width*scale,pic.height*scale).rgb();assert actual_pixels==target_pixels,(step,lang)
        picture_checks.append({'step':step,'language':lang,'source':str(source),'scale':scale,'position':xy,'pixel_differences':0})
target=Path('docs/verification/cjk-ui/longest-title-20261001');assert not target.exists()
assert (target.parent.stat().st_uid,target.parent.stat().st_gid)==(os.getuid(),os.getgid());target.mkdir();files={}
def copy(p,name):shutil.copyfile(p,target/name);files[name]=sha(target/name)
for name in ('selection.json','script.txt','comparison.json'):copy(root/name,name)
for lang in ('en','zh-Hant'):
    for name in ['trace.json','run.json','engine.log']+[f'step-{s:03}.png' for s in selection['shot_steps']]:copy(root/lang/name,f'{lang}-{name}')
for name in ('kol_cjk_longest_title_run.py','kol_cjk_longest_title_verify.py'):copy(Path('/control')/name,name)
r={'schema':'kol-cjk-longest-title-v1','date':'2026-10-01','status':'pixel-confirmed-visual-review-pending','scope':selection['scope'],'selection':selection,
    'comparison':comparison,'line_checks':line_checks,'title_pixel_differences':d,'single_missing_ink_pixel_rejected':True,'picture_checks':picture_checks,
    'geometry':{'title':[306,174,585,60],'inner_safe_rect':[307,175,583,58],'line_step':24,'font_height':18,'centred':True},
    'geometry_evidence':'internal/town/layout.go CJKInn 與 internal/ui/town.go InnCJK，標題按正式字模 Measure 置中。',
    'binary_sha256':json.loads((root/'en/run.json').read_text())['binary_sha256'],'seed':{'remake':1,'original':None,'method':'執行前 -seed 1；既有完整自然按鍵加正常離店步行，未重新挑種子。'},
    'inputs_sha256':{str(p):sha(p) for p in sources+[fontpath,Path('assets/ui-font/manifest.json'),Path('docs/i18n/town-zh-Hant.json'),Path('internal/town/layout.go'),Path('internal/ui/town.go'),Path('assets/original-derived-data/town-buildings.json')]},
    'toolchain':'eob-remake-go:1.26.7-ebiten2.9.9-audio；sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    'rights':'原作衍生 PNG 僅適用既有 private GitHub 特例，不授權公開發布。',
    'visual_review':{'status':'pending','files':[f'zh-Hant-step-{s:03}.png' for s in selection['shot_steps']]+[f'en-step-{selection["inside_step"]:03}.png']},
    'limitations':'只驗正式 town-title 全母體最寬之一及所列正常自然旅程。不宣稱另外五則同寬標題、原版亂數／同狀態、存檔或平台完成。肖像名牌與其他動態欄位仍待驗，#34／#40不結案。',
    'files':files}
(target/'receipt.json').write_text(json.dumps(r,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'receipt':str(target/'receipt.json'),'states':comparison['states'],'widths':[c['width'] for c in line_checks],'title_pixels':d,'picture_checks':len(picture_checks),'files':len(files)},ensure_ascii=False))
