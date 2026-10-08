from pathlib import Path
import json,hashlib,os,sys,shutil
sys.path.insert(0,'tools')
from png_io import read_png
root=Path('workplace/cjk-longest-dialog-20261001');selection=json.loads((root/'selection.json').read_text())
en=json.loads((root/'en/trace.json').read_text());zh=json.loads((root/'zh-Hant/trace.json').read_text());assert en==zh and len(zh['steps'])==169
messages=json.loads(Path('docs/i18n/town-zh-Hant.json').read_text())['towns']['7']['doors']['8']['messages']
bank=json.loads(Path('docs/i18n/ui-zh-Hant.json').read_text())['banks']['KOLDATA']
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0];gs={g['character']:g for g in fm['glyphs']};fontpath=Path('assets/ui-font')/fm['file'];font=read_png(fontpath)
def sha(p):return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def width(s):return sum(gs[c]['advance'] if c in gs else fm['advance'] for c in s)
checks=[]
for cycle,category,message in zip(selection['cycles'],[3,1,0],[20,24,23]):
    lead=bank[str(0x9A+category)]['zh'];assert '@' not in lead and '~' not in lead
    for step,kind,lines in [(cycle['prompt_step'],'question',messages['18']+[lead]),(cycle['answer_step'],'answer',messages[str(message)])]:
        w,h=585,204;expected=bytearray(w*h*3);ink=[]
        for yy in range(h):
            for xx in range(w):
                if xx in (0,w-1) or yy in (0,h-1):i=(yy*w+xx)*3;expected[i:i+3]=b'\xff\xff\xff'
        for row,line in enumerate(lines):
            assert width(line)<=575 and 258+row*24+18<455
            pen=9;top=6+row*24
            for c in line:
                if c not in gs:assert c in (' ','\u3000');pen+=fm['advance'];continue
                g=gs[c];sx=g['sheet_column']*16;sy=g['sheet_row']*18
                for yy in range(18):
                    for xx in range(16):
                        a=font.pixels[((sy+yy)*font.width+sx+xx)*4+3];assert a in (0,255)
                        if a:i=((top+yy)*w+pen+xx)*3;expected[i:i+3]=b'\xff\xff\xff';ink.append(i)
                pen+=g['advance']
        actual=read_png(root/'zh-Hant'/f'step-{step:03}.png').crop(306,252,w,h).rgb()
        d=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,len(actual),3));assert d==0,(step,d)
        wrong=expected.copy();wrong[ink[0]:ink[0]+3]=b'\0\0\0';assert sum(actual[i:i+3]!=wrong[i:i+3] for i in range(0,len(actual),3))==1
        checks.append({'step':step,'kind':kind,'cycle':cycle['cycle'],'category':category,'record':0x9A+category if kind=='question' else message,
            'lines':lines,'widths':[width(s) for s in lines],'bottom':258+(len(lines)-1)*24+18,'pixel_differences':d,'single_missing_ink_pixel_rejected':True})
assert checks[0]['lines'][-1]==selection['selected'][0]['zh'] and width(checks[0]['lines'][-1])==240
source=Path('assets/original-derived-png/ega/TWN7PICS/twn7pics-018.png');pic=read_png(source);raw=pic.rgb();pictures=[]
for check in checks:
    for lang,scale,xy in [('en',1,(8,64)),('zh-Hant',3,(24,192))]:
        expected=bytearray()
        for yy in range(pic.height*scale):
            for xx in range(pic.width*scale):i=((yy//scale)*pic.width+xx//scale)*3;expected+=raw[i:i+3]
        actual=read_png(root/lang/f'step-{check["step"]:03}.png').crop(*xy,pic.width*scale,pic.height*scale).rgb();assert actual==expected
        pictures.append({'step':check['step'],'language':lang,'source':str(source),'scale':scale,'pixel_differences':0})
target=Path('docs/verification/cjk-ui/longest-dialog-20261001');assert not target.exists()
assert (target.parent.stat().st_uid,target.parent.stat().st_gid)==(os.getuid(),os.getgid());target.mkdir();files={}
def copy(p,name):shutil.copyfile(p,target/name);files[name]=sha(target/name)
for name in ('selection.json','script.txt'):copy(root/name,name)
for lang in ('en','zh-Hant'):
    for name in ['trace.json','run.json','engine.log']+[f'step-{s:03}.png' for s in selection['shot_steps']]:copy(root/lang/name,f'{lang}-{name}')
for name in ('kol_cjk_longest_dialog_run.py','kol_cjk_longest_dialog_verify.py'):copy(Path('/control')/name,name)
r={'schema':'kol-cjk-longest-dialog-v1','date':'2026-10-01','status':'normal-player-path-pixel-and-visual-confirmed','scope':selection['scope'],'selection':selection,
    'identical_en_zh_states':169,'checks':checks,'picture_checks':pictures,'binary_sha256':json.loads((root/'en/run.json').read_text())['binary_sha256'],
    'seed':{'remake':1,'original':None,'method':'執行前 -seed 1；一次走完三則固定順序傳聞，未重擲挑結果。'},
    'geometry':{'dialog':[306,252,585,204],'text_start':[315,258],'safe_width':575,'line_step':24,'font_height':18},
    'visual_review':{'status':'passed','method':'主代理檢視三張中文問句、三張中文答覆及英文最長問句完整畫格；三次正常聆聽與兩段確認完整、可讀、無裁切。'},
    'inputs_sha256':{str(p):sha(p) for p in [source,fontpath,Path('assets/ui-font/manifest.json'),Path('docs/i18n/town-zh-Hant.json'),Path('docs/i18n/ui-zh-Hant.json'),Path('assets/original-derived-data/town-keywords.json'),Path('assets/original-derived-data/town-buildings.json'),Path('internal/town/session.go'),Path('internal/town/topics.go'),Path('internal/ui/town.go'),Path('internal/town/layout.go')]},
    'toolchain':'eob-remake-go:1.26.7-ebiten2.9.9-audio；sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
    'rights':'原作衍生 PNG 僅適用既有 private GitHub 特例，不授權公開發布。',
    'limitations':'只驗正式 dialog-ui 母體最寬問句與所列三則正常聆聽；不宣稱原版亂數、同狀態或全話題對拍。肖像名牌、其他欄位、最長標題及平台仍未完成；#34／#40保持開放。',
    'files':files}
(target/'receipt.json').write_text(json.dumps(r,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'receipt':str(target/'receipt.json'),'states':169,'dialog_checks':len(checks),'source_picture_checks':len(pictures),'files':len(files)},ensure_ascii=False))
