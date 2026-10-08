"""繁中物品立繪：原版同期記錄＋固定PNG獨立合成，完整畫面及存檔核對。"""
from pathlib import Path
import argparse, hashlib, json, os, shutil, sys
ROOT, OUT = Path('/src'), Path('/out')
PROOF = ROOT/'docs/verification/item-detail-20261001'
sys.path.insert(0, str(ROOT/'tools'))
from png_io import read_png

AFTER = '325ba73b2617743dd84697adc50e31c68e16152f6bf0093d3f967140be4d015a'
BEFORE = '4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4'
STAGES = [15,28,42,56,72,88]
# 同一正常操作的清單等待區間；提問YES/NO雖然也有兩圖示，仍屬明細底層。
LIST = {}
for start,end in [(3,12),(15,25),(28,37),(42,53),(56,66),(72,82),(88,98)]:
    LIST.update({i:i-start for i in range(start,end+1)})
LIST.update({100:10,101:10})
FOCAL = [3,15,25,28,42,56,72,82,88,100]

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p): return json.loads(p.read_text(encoding='utf-8'))
def project_path(p):
    if p.is_relative_to(OUT):
        return str(Path('workplace/item-detail-20261001')/p.relative_to(OUT))
    return str(p.relative_to(ROOT))
def restore_rect(dst, before, width, x,y,w,h):
    for yy in range(y,y+h):
        a=(yy*width+x)*3; b=a+w*3
        dst[a:b]=before[a:b]
def mark_rects(layout, row, scale):
    x=(110 if row==10 else 6)*scale
    y=layout['slot_marks'][3][row]*scale
    w,h=100*scale,22*scale
    # 現有中文框線與投影是1px／4px；本節不改它們。
    return [(x,y,w,h),(x+w,y+4,4,h),(x+4,y+h,w,4)]
def piece_plan(layout, raw):
    # 正常角色Theodoric：前版正常保存已確認男性、種族3；不注入新角色。
    v=3; body=layout['body'][v]
    result=[('PERPICS',v,dict(x=body['x'],y=body['y'],w=body['x2']-body['x'],h=body['y2']-body['y']))]
    for layer,archive in [(2,'LEGSPICS'),(1,'TORPICS'),(0,'HEADPICS')]:
        rec=raw[256+16*layer:272+16*layer]
        if rec[0]:
            index=v*9+(rec[14]&15)
            result.append((archive,index,layout['layers'][index][layer]))
    held=raw[48:64]
    has_weapon=bool(held[0] and held[12]&3 in (1,3))
    two_hand=has_weapon and bool(held[13]&2)
    shield=raw[352:368]
    assert not shield[0], '此正常母體沒有盾牌；其他造型需另列母體'
    def weapon_index(rec):
        kind,style=rec[13]>>5,rec[14]&15
        if kind==5 and style in (2,3): style^=1
        return (layout['weapon_base'][kind]+2*style)//2
    if has_weapon:
        index=weapon_index(held)
        result.append(('WEAPPICS',index,layout['weapons'][v][index]))
    belt=raw[400:416]
    if belt[0]:
        index=weapon_index(belt); c=layout['weapons'][0][index]
        result.append(('WEAPPICS',index,dict(x=2,y=153-c['h'],w=c['w'],h=c['h'])))
    for k in range(2 if two_hand else 1):
        index=v*2+k; result.append(('HANDPICS',index,layout['hands'][index]))
    return result

def compose(before, plan, scale, pngs, row, layout, shift=0):
    width,height=320*scale,200*scale
    assert (before.width,before.height)==(width,height)
    expected=bytearray(before.rgb())
    for archive,index,c in plan:
        img=pngs[(archive,index)]
        assert (img.width,img.height)==(c['w']*8,c['h']) and img.channels==4
        for y in range(img.height):
            for x in range(img.width):
                pos=(y*img.width+x)*4; alpha=img.pixels[pos+3]
                assert alpha in (0,255)
                if not alpha: continue
                rgb=img.pixels[pos:pos+3]
                for sy in range(scale):
                    yy=(c['y']+y)*scale+sy
                    xx=(c['x']*8+x)*scale+shift
                    assert 0<=xx and xx+scale<=width and 0<=yy<height
                    p=(yy*width+xx)*3;expected[p:p+scale*3]=rgb*scale
    for r in mark_rects(layout,row,scale):restore_rect(expected,before.rgb(),width,*r)
    return bytes(expected)

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
    old=load(PROOF/'mirror-receipt.json')
    assert sha(PROOF/'mirror-receipt.json')=='82d4e21da58dbeb0ecc86e2b6e0376679f4154a41025facf273443361f97ee01'
    for a in old['attachments']: assert sha(PROOF/a['path'])==a['sha256']
    assert sha(OUT/'kol-items-cjk-figure-linux')==AFTER
    assert sha(OUT/'kol-items-mirror-linux')==BEFORE
    assert sha(ROOT/'Knights-of-legend.zip')==old['source_zip_sha256']
    for k in ['source_inputs_private','readonly_actor_private']:
        assert sha(ROOT/old[k]['path'])==old[k]['sha256']
    states=load(ROOT/old['readonly_actor_private']['path'])['states']
    assert len(states)==104
    layout=load(ROOT/'assets/original-derived-data/portrait-layout.json')
    assert layout['source_sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    manifest=load(ROOT/'assets/original-derived-png/ega/manifest.json')
    assert manifest['source_zip']['sha256']==old['source_zip_sha256']
    pngs={};input_assets={}
    for a in manifest['archives']:
        if a['archive'] not in ['PERPICS','HEADPICS','TORPICS','LEGSPICS','WEAPPICS','HANDPICS']:continue
        for img in a['images']:
            p=ROOT/'assets/original-derived-png/ega'/img['file']
            assert sha(p)==img['png_sha256']
            pngs[(a['archive'],img['unit_index'])]=read_png(p)
            input_assets[str(p.relative_to(ROOT))]=sha(p)
    traces=[];copies=[];checks=[];unchanged=[];changed_pixels=0
    for lang in ['en','zh-Hant']:
        d=OUT/f'normal-cjk-figure-{lang}'
        cmd=load(d/'command.json');keys=cmd[cmd.index('-flow-script')+1].split()
        mapping={'.':'>','Return':'RET','Down':'D','Up':'U'}
        assert keys==['SPACE','SPACE','6','RET','SPACE']+[mapping[k] for k in old['original_keys']]
        assert cmd[cmd.index('-seed')+1]=='1' and len(keys)==109
        trace=load(d/'trace.json')['steps'];assert len(trace)==110
        assert trace==load(OUT/f'normal-mirror-{lang}/trace.json')['steps'];traces.append(trace)
        for step in range(110):
            now=read_png(d/f'step-{step:03d}.png')
            before=read_png(OUT/f'normal-mirror-{lang}'/f'step-{step:03d}.png')
            if lang=='en' or step-6 not in LIST:
                assert now.rgb()==before.rgb(),(lang,step,'不應改變的畫面')
                if lang=='zh-Hant':unchanged.append(step)
                continue
            frame=step-6;raw=bytes.fromhex(states[frame]['character_record_hex'])
            assert raw[:10].split(b'\0')[0]==b'Theodoric'
            plan=piece_plan(layout,raw)
            expected=compose(before,plan,3,pngs,LIST[frame],layout)
            assert now.rgb()==expected,(frame,'PNG獨立合成不符')
            n=sum(now.rgb()[p:p+3]!=before.rgb()[p:p+3] for p in range(0,len(expected),3))
            assert n>0,(frame,'仍漏人物')
            changed_pixels+=n
            checks.append(dict(original_frame=frame,remake_step=step,selected_row=LIST[frame],
                               canvas=[960,600],scale=3,independent_composite_pixels_different=0,
                               restored_figure_pixels=n,pieces=[dict(archive=a,index=i,**c) for a,i,c in plan]))
            if frame in FOCAL:
                for p,name in [(d/f'step-{step:03d}.png',f'cjk-figure-zh-Hant-{step:03d}.png'),
                               (OUT/'normal-cjk-figure-en'/f'step-{step:03d}.png',f'cjk-figure-en-{step:03d}.png'),
                               (OUT/'gui-actions1'/f'item-{frame:02d}.png',f'cjk-figure-original-{frame:02d}.png')]:
                    copies.append((p,name))
    assert traces[0]==traces[1] and len(checks)==78 and len(unchanged)==32
    # 反例：實際舊版漏圖與三倍立繪整體右移1px，必須被完整畫面核對拒絕。
    ref=read_png(OUT/'normal-cjk-figure-zh-Hant/step-021.png')
    prior=read_png(OUT/'normal-mirror-zh-Hant/step-021.png')
    plan=piece_plan(layout,bytes.fromhex(states[15]['character_record_hex']))
    assert ref.rgb()!=prior.rgb()
    assert compose(prior,plan,3,pngs,0,layout,shift=1)!=ref.rgb()
    # 明細等待頁保持原樣；誤貼一層的反例需拒絕。
    detail=read_png(OUT/'normal-cjk-figure-zh-Hant/step-020.png')
    assert compose(detail,plan[:1],3,pngs,9,layout)!=detail.rgb()
    copies.append((OUT/'normal-mirror-zh-Hant/step-021.png','cjk-figure-before-zh-Hant-021.png'))
    persistence=[]
    for frame in STAGES:
        for lang in ['en','zh-Hant']:
            for phase in ['save','reload']:
                p=OUT/f'{phase}-cjk-figure-{frame:02d}-{lang}/save.json'
                a,b=load(p),load(OUT/f'{phase}-mirror-{frame:02d}-{lang}/save.json')
                a.pop('saved_at');b.pop('saved_at');assert a==b,(frame,lang,phase,'持久資料變動')
                c=a['characters'][0]
                assert not c['female'] and c['sheet']['race']==3 and c['shield'] is None
                persistence.append(dict(frame=frame,language=lang,phase=phase,path=project_path(p),sha256=sha(p),
                                        compared_with_RE175_except_saved_at=True))
            d=OUT/f'reload-cjk-figure-{frame:02d}-{lang}'
            cmd=load(d/'command.json');detail_step=int(cmd[cmd.index('-shot-steps')+1].split(',')[0])
            p=d/f'step-{detail_step:03d}.png'
            assert read_png(p).rgb()==read_png(OUT/f'reload-mirror-{frame:02d}-{lang}'/p.name).rgb()
            copies.append((p,f'cjk-figure-reload-{frame:02d}-{lang}.png'))
    assert len(persistence)==24
    inputs={v['path']:sha(ROOT/v['path']) for v in old['product_inputs']}
    inputs.update(input_assets)
    inputs.update({str(Path(__file__).relative_to(ROOT)):sha(Path(__file__))})
    attachments=[]
    for src,name in copies:
        target=PROOF/name
        if args.export:
            assert not target.exists();shutil.copyfile(src,target)
        assert sha(src)==sha(target)
        attachments.append(dict(path=name,sha256=sha(target),rights='private-repo可保存；不得公開散布',
                                source_path_private=project_path(src)))
    receipt=dict(schema='kol-item-cjk-figure-v1',date='2026-10-01',status='繁中物品立繪正常顯示窄切片通過',
                 binary=dict(path=project_path(OUT/'kol-items-cjk-figure-linux'),sha256=AFTER),
                 before_binary_sha256=BEFORE,source_zip_sha256=old['source_zip_sha256'],
                 source_executable_sha256=old['source_executable_sha256'],original_revision=old['original_revision'],
                 original_seed=1,remake_seed=1,seed_setup=old['seed_setup'],normal_startup=True,state_injection=False,direct_entry=False,
                 source_receipt=dict(path='mirror-receipt.json',sha256=sha(PROOF/'mirror-receipt.json')),
                 original_keys=old['original_keys'],remake_keys_count=109,english_unchanged_frames=110,
                 language_states_equal=110,checks=checks,cjk_unchanged_steps=unchanged,
                 restored_pixels_total=changed_pixels,persistence_checks=persistence,
                 negative_controls=['實際前版漏圖拒絕','三倍立繪右移1px拒絕','明細誤貼人物拒絕'],
                 english_pending_frames=old['unverified_list_frames'],original_zero_frames=old['all_zero_frames'],
                 tool=dict(go='1.26.7',ebitengine='2.9.9',image='eob-remake-go:1.26.7-ebiten2.9.9-audio'),
                 current_inputs=[dict(path=p,sha256=h) for p,h in sorted(inputs.items())],attachments=attachments,
                 verifier_sha256=sha(Path(__file__)),
                 limits='原版無繁中畫面；只證明同期美術三倍合成及正常狀態／保存不變。英文70張清單差異、其他物品動作與背包洞位仍未完成。')
    target=PROOF/'cjk-figure-receipt.json'
    if args.export:
        assert not target.exists();target.write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    else:assert load(target)==receipt,'收據重算不符'
    print(json.dumps(dict(status='通過',figure_frames=len(checks),english_unchanged=110,
                          cjk_other_unchanged=len(unchanged),persistence=len(persistence),attachments=len(attachments)),ensure_ascii=False))

if __name__=='__main__':main()
