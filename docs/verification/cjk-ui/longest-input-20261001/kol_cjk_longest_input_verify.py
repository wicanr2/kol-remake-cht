from pathlib import Path
import json,hashlib,os,sys,shutil
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-longest-input-20261001/attempt2');selection=json.loads((root/'selection.json').read_text())
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0];gs={g['character']:g for g in fm['glyphs']};fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath)
paperpath=Path('assets/original-derived-png/ega/BOOKPICS/bookpics-002.png');paper=read_png(paperpath);raw=paper.rgb()
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def width(s):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in s)
def base(x,y,w,h):
    out=bytearray()
    for yy in range(y,y+h):
        for xx in range(x,x+w):i=((yy//3)*paper.width+xx//3)*3;out+=raw[i:i+3]
    return out
def glyph_draw(buf,w,text,colour):
    pen=0
    for c in text:
        if c in gs:
            g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
            for yy in range(18):
                for xx in range(16):
                    a=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert a in (0,255)
                    if a:i=(yy*w+pen+xx)*3;buf[i:i+3]=colour
        pen+=gs[c]['advance'] if c in gs else fm['advance']
def compare(actual,expected):return sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3))
samples=[('letters',5,'',True),('letters',20,'W'*15,True),('letters',21,'W'*14,True),('letters',23,'W'*16,False),('letters',24,'W'*16,False),
    ('spaces',20,' '*15,True),('spaces',21,' '*15+'W',False),('spaces',22,' '*15+'W',False),('all-spaces',20,' '*15,True),('all-spaces',21,' '*16,False),('all-spaces',22,' '*16,False)]
checks=[]
for case,step,text,caret in samples:
    expected=base(72,138,438,18);pen=0
    for c in text:
        advance=gs[c]['advance'] if c in gs else fm['advance']
        if c!=' ' and c in gs:
            for y in range(18):
                for x in range(pen,pen+advance):i=(y*438+x)*3;expected[i:i+3]=b'\xff\xff\xff'
        pen+=advance
    glyph_draw(expected,438,text,b'\0\0\0')
    if caret:
        for y in (16,17):
            for x in range(pen,pen+8):i=(y*438+x)*3;expected[i:i+3]=b'\0\0\0'
    actual=read_png(root/case/'zh-Hant'/f'step-{step:03}.png').crop(72,138,438,18).rgb();d=compare(actual,expected);assert d==0,(case,step,d)
    assert pen+(8 if caret else 0)<=438
    checks.append({'case':case,'step':step,'text':text,'width':pen,'caret':caret,'caret_x':72+pen if caret else None,'pixel_differences':d})
unchanged=[];counts={}
for name,case in selection['cases'].items():
    en=json.loads((root/name/'en/trace.json').read_text());zh=json.loads((root/name/'zh-Hant/trace.json').read_text());assert en==zh
    assert all(row['scene']=='flow:newchar' for row in zh['steps'][5:]);counts[name]=len(zh['steps'])
    pair=(23,24) if name=='letters' else (21,22)
    for lang in ('en','zh-Hant'):
        a=read_png(root/name/lang/f'step-{pair[0]:03}.png').rgb();b=read_png(root/name/lang/f'step-{pair[1]:03}.png').rgb();assert a==b
        unchanged.append({'case':name,'language':lang,'steps':pair,'full_frame_pixel_differences':0})
portrait_checks=[]
for case,step in [('letters',29),('spaces',27)]:
    text=selection['cases'][case]['maximum_text'];expected=base(624,438,266,18)
    for yy in range(18):
        for xx in range(624,747):
            i=(yy*266+xx-624)*3;expected[i:i+3]=b'\0\0\0' if xx<745 else bytes([0,170,170])
    glyph_draw(expected,266,text,b'\xff\xff\xff')
    actual=read_png(root/case/'zh-Hant'/f'step-{step:03}.png').crop(624,438,266,18).rgb();d=compare(actual,expected);assert d==0,(case,d)
    over=max(0,width(text)-121);assert over>0
    portrait_checks.append({'case':case,'step':step,'text':text,'width':width(text),'name_start':[624,438],'black_backing_right_exclusive':745,'safe_width':121,'overflow':over,'pixel_differences':d,'display_pass':False})
controls=json.loads((root/'width-control.json').read_text());assert controls['changes'][0][1]['width']==256
target=Path('docs/verification/cjk-ui/longest-input-20261001');assert not target.exists()
assert (target.parent.stat().st_uid,target.parent.stat().st_gid)==(os.getuid(),os.getgid());target.mkdir();files={}
def copy(p,name):shutil.copyfile(p,target/name);files[name]=sha(target/name)
for name in ('selection.json','width-control.json','negative-layout.json','negative-before.json','negative-after.json'):copy(root/name,name)
copy(Path('docs/verification/cjk-text-fit.json'),'text-fit.json')
copy(Path('docs/verification/cjk-ui/font-width-20261001/text-fit.json'),'before-text-fit.json')
copy(Path('tools/verify_text_fit.py'),'verify_text_fit.py.txt')
for name,case in selection['cases'].items():
    copy(root/name/'script.txt',f'{name}-script.txt')
    for lang in ('en','zh-Hant'):
        for filename in ['trace.json','run.json','engine.log']+[f'step-{s:03}.png' for s in case['shots']]:copy(root/name/lang/filename,f'{name}-{lang}-{filename}')
for name in ('kol_cjk_longest_input_run.py','kol_cjk_allspace_run.py','kol_cjk_input_width_finish.py','kol_cjk_longest_input_verify.py','kol_nameplate_prototype.go'):copy(Path('/control')/name,name)
for c in ('A','B'):copy(root/'nameplate-prototype'/f'option-{c}.png',f'prototype-{c}.png')
receipt={'schema':'kol-cjk-longest-input-v1','date':'2026-10-01','status':'normal-input-confirmed-portrait-nameplate-failed','scope':selection['scope'],
    'selection':selection,'identical_en_zh_states_by_case':counts,'identical_en_zh_states_total':sum(counts.values()),'name_field_checks':checks,'extra_letter_ignored_checks':unchanged,
    'portrait_nameplate_checks':portrait_checks,'static_capacity_correction':{'before':240,'after':256,'checked':9635,'failed':0,'negative_before_exit':0,'negative_after_exit':1},
    'binary_sha256':json.loads((root/'letters/en/run.json').read_text())['binary_sha256'],'seed':{'remake':1,'original':None,'method':'執行前 -seed 1'},
    'geometry_evidence':'internal/book/layout_cjk.go、internal/ui/newcharacter_cjk.go、internal/font/font.go；輸入欄與名牌為兩個不同欄位。',
    'inputs_sha256':{str(p):sha(p) for p in [paperpath,fontpath,Path('assets/ui-font/manifest.json'),Path('internal/book/layout_cjk.go'),Path('internal/ui/newcharacter_cjk.go'),Path('internal/font/font.go'),Path('internal/charcreate/charcreate.go'),Path('internal/flow/flow.go'),Path('cmd/kol/input.go')]},
    'decision':{'status':'pending','scope':'繁中肖像名牌；不更動姓名容量、保存或重名政策','recommendation':'A：依實際字寬加寬名牌，完整顯示','alternatives':['A：加寬名牌','B：名牌內縮略加省略號'],'prototype_only':True},
    'limitations':'正式名牌未修，#34／#40 未完成；英文名牌現象亦保留，不宣稱原版對拍。未保存角色、測試中文輸入、重名、其他動態名稱或平台產物。',
    'toolchain':'eob-remake-go:1.26.7-ebiten2.9.9-audio；sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    'rights':'含原作衍生 PNG 的畫格與介面原型僅適用既有 private GitHub PNG 特例，不授權公開發布。',
    'files':files}
(target/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'receipt':str(target/'receipt.json'),'states':counts,'field_pixel_checks':len(checks),'portrait_overflows':[p['overflow'] for p in portrait_checks],'files':len(files)},ensure_ascii=False))
