from pathlib import Path
import json,hashlib,os,shutil,sys
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-longest-battle-20261001')
selection=json.loads((root/'selection.json').read_text());run=json.loads((root/'run.json').read_text());assert run['returncode']==0
rows=json.loads((root/'trace.json').read_text())['steps']
mother=Path('docs/verification/cjk-ui/natural-sign-20261001')
reference=json.loads((mother/'zh-Hant-trace.json').read_text())['steps'];assert rows==reference[:len(rows)]
assert len(rows)==2568
fm=json.loads(Path('assets/ui-font/manifest.json').read_text());f=fm['fonts'][0];glyphs={g['character']:g for g in f['glyphs']}
fontpath=Path('assets/ui-font')/f['file'];font=read_png(fontpath);assert font.channels==4
source=Path('assets/original-derived-png/ega/MISCPICS/miscpics-001.png');paper=read_png(source);assert (paper.width,paper.height)==(296,32)
raw=paper.rgb();x0,y0,w,h=24,492,888,96;tx,ty,safe=72,504,816
def width(text):return sum(glyphs[c]['advance'] if c in glyphs else f['advance'] for c in text)
def word(c):return c.isascii() and (c.isalnum() or c=="'")
def wrap(lines):
    text=''
    for line in lines:
        if text and line and word(text[-1]) and word(line[0]):text+=' '
        text+=line
    result=[]
    while text:
        n=len(text)
        while n>1 and width(text[:n])>safe:n-=1
        if n<len(text) and word(text[n-1]) and word(text[n]):
            for k in range(n-1,0,-1):
                if text[k]==' ':n=k;break
        result.append(text[:n].rstrip(' '));text=text[n:].lstrip(' ')
    return result
checks=[]
for step in selection['shot_steps']:
    state=rows[step]['battle'];assert state['prompt'] in (9,10,11,12)
    lines=wrap(state['exec_text']);assert len(lines)<=4
    frame=read_png(root/'frames'/f'step-{step:03}.png');actual=frame.crop(x0,y0,w,h).rgb()
    expected=bytearray()
    for y in range(h):
        for x in range(w):
            pos=((y//3)*paper.width+x//3)*3;expected+=raw[pos:pos+3]
    changed=[]
    for row,line in enumerate(lines):
        assert width(line)<=safe
        pen=tx-x0;top=ty-y0+row*18
        for char in line:
            if char not in glyphs:
                assert char in (' ','\u3000');pen+=f['advance'];continue
            g=glyphs[char];sx=g['sheet_column']*16;sy=g['sheet_row']*18
            for yy in range(18):
                for xx in range(16):
                    alpha=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert alpha in (0,255)
                    if alpha:
                        pos=((top+yy)*w+pen+xx)*3
                        if expected[pos:pos+3]!=b'\0\0\0':changed.append(pos)
                        expected[pos:pos+3]=b'\0\0\0'
            pen+=g['advance']
    differences=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert differences==0,(step,differences)
    assert changed
    corrupt=expected.copy();p=changed[0];corrupt[p:p+3]=b'\xff\xff\xff'
    assert sum(actual[i:i+3]!=corrupt[i:i+3] for i in range(0,len(actual),3))==1
    checks.append({'step':step,'source_lines':state['exec_text'],'rendered_lines':lines,'widths':[width(x) for x in lines],
        'bottom':ty+(len(lines)-1)*18+18,'scroll_bottom':y0+h,'pixel_differences':differences,
        'ink_pixels':len(changed),'single_missing_ink_pixel_rejected':True})
target=Path('docs/verification/cjk-ui/longest-battle-20261001');assert not target.exists()
assert (target.parent.stat().st_uid,target.parent.stat().st_gid)==(os.getuid(),os.getgid());target.mkdir()
files={}
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def copy(p,name):shutil.copyfile(p,target/name);files[name]=sha(target/name)
for name in ['selection.json','run.json','engine.log','trace.json','script.txt']:copy(root/name,name)
for step in selection['shot_steps']:copy(root/'frames'/f'step-{step:03}.png',f'step-{step:03}.png')
for name in ['kol_cjk_longest_battle_run.py','kol_cjk_longest_battle_verify.py']:copy(Path('/control')/name,name)
receipt={'schema':'kol-cjk-longest-natural-battle-v1','date':'2026-10-01','status':'pixel-confirmed-visual-review-pending',
    'scope':selection['scope'],'source_receipt':str(mother/'receipt.json'),'selected_before_execution':True,
    'selection':selection,'identical_reference_states':len(rows),'geometry':{'scroll':[x0,y0,w,h],'text_start':[tx,ty],'safe_width':safe,'line_height':18,'max_lines':4},
    'geometry_evidence':'internal/battle/message.go MessageBox／MessageArea 與 internal/ui/battle_cjk.go BattleFieldCJK 的原座標乘三；正式 renderer 不改。',
    'checks':checks,'binary_sha256':run['binary_sha256'],'seed':{'remake':1,'original':None,'method':'執行前 -seed 1'},
    'inputs_sha256':{str(p):sha(p) for p in [source,Path('assets/ui-font/manifest.json'),fontpath,Path('internal/ui/battle_cjk.go'),Path('internal/font/font.go'),mother/'zh-Hant-trace.json',mother/'script.txt']},
    'toolchain':'eob-remake-go:1.26.7-ebiten2.9.9-audio；sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    'rights':'重製執行畫面含原作衍生圖，沿用既有 private GitHub PNG 特例；沒有公開散布授權。',
    'limitations':'全遊戲其他訊息／最長輸入與其他畫面仍未全覆蓋；#34／#40 不結案。原版、存檔與平台封包未驗。',
    'visual_review':{'status':'pending','files':[f'step-{s:03}.png' for s in selection['shot_steps']]},'files':files}
(target/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'states':len(rows),'candidate_count':selection['candidate_count'],'maximum_width':selection['maximum_width'],'checks':checks,'receipt':str(target/'receipt.json')},ensure_ascii=False))
