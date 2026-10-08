"""手持小框列號：原版完整畫面、獨立中文字模及持久資料核對。"""
from pathlib import Path
import argparse,hashlib,importlib.util,json,shutil,sys
ROOT,OUT=Path('/src'),Path('/out');PROOF=ROOT/'docs/verification/item-detail-20261001'
sys.path.insert(0,str(ROOT/'tools'))
from png_io import read_png
AFTER='821ee273ce3d78f658d4efa83b2716ec89a12e4002fa0365b7c8bf17fc86b78a'
BEFORE='325ba73b2617743dd84697adc50e31c68e16152f6bf0093d3f967140be4d015a'
FRAMES=[25,66,82];STAGES=[15,28,42,56,72,88]
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text(encoding='utf-8'))
def path(p):
    if p.is_relative_to(OUT):return str(Path('workplace/item-detail-20261001')/p.relative_to(OUT))
    return str(p.relative_to(ROOT))
def main():
    ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
    old=load(PROOF/'cjk-figure-receipt.json');mirror=load(PROOF/'mirror-receipt.json')
    assert sha(PROOF/'cjk-figure-receipt.json')=='e419f1e8f1bbb1775a27695c18a2e49652ce966b1c2964ed952a4f12c582eec2'
    assert sha(PROOF/'mirror-receipt.json')=='82d4e21da58dbeb0ecc86e2b6e0376679f4154a41025facf273443361f97ee01'
    for receipt in [old,mirror]:
        for a in receipt['attachments']:assert sha(PROOF/a['path'])==a['sha256']
    assert sha(OUT/'kol-items-mark-linux')==AFTER and sha(OUT/'kol-items-cjk-figure-linux')==BEFORE
    assert sha(ROOT/'Knights-of-legend.zip')==old['source_zip_sha256']
    assert sha(PROOF/'verify_actions.py')==load(PROOF/'actions-receipt.json')['verifier_sha256']
    spec=importlib.util.spec_from_file_location('historical_actions',PROOF/'verify_actions.py')
    legacy=importlib.util.module_from_spec(spec);spec.loader.exec_module(legacy)
    for k in ['source_inputs_private','readonly_actor_private']:assert sha(ROOT/mirror[k]['path'])==mirror[k]['sha256']
    oracle=load(ROOT/mirror['readonly_actor_private']['path']);states=oracle['states'];assert len(states)==104
    annotated=load(OUT/'item-mark-annotated-private.json');assert annotated['tool']=='IDA Pro 9.4' and annotated['address_space']=='IDA linear'
    assert annotated['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64')==annotated['database_sha256']
    rows=annotated['records'][0]['instructions'];assert len(rows)==231
    attached=[a for r in rows for a in r['annotations']];assert len(attached)==7
    for a in attached:assert a['confidence']=='已證實' and a['evidence'] and sha(Path(a['index']))==a['index_sha256']
    by_ea={r['original_ea']:r for r in rows}
    assert by_ea['0x16f36']['bytes']=='050800' and by_ea['0x16f3f']['bytes']=='053100'
    assert by_ea['0x16f52']['bytes']=='837efa04'
    fm=load(ROOT/'assets/ui-font/manifest.json')['fonts'][0]
    font=read_png(ROOT/'assets/ui-font'/fm['file']);assert sha(ROOT/'assets/ui-font'/fm['file'])==fm['png_sha256']
    glyph=next(g for g in fm['glyphs'] if g['character']=='1')
    assert fm['cell']==dict(width=16,height=18) and fm['advance']==16 and font.channels==4
    want=bytearray(16*18*3);ink=[]
    for y in range(18):
        for x in range(16):
            n=((glyph['sheet_row']*18+y)*font.width+glyph['sheet_column']*16+x)*4
            alpha=font.pixels[n+3];assert alpha in [0,255]
            if alpha:want[(y*16+x)*3:(y*16+x+1)*3]=b'\xff\xff\xff';ink.append((x,y))
    assert ink
    expected_keys=['SPACE','SPACE','6','RET','SPACE']+[{'.':'>','Return':'RET','Down':'D','Up':'U'}[k] for k in old['original_keys']]
    traces=[];checks=[];pending=[];zero=[];copies=[]
    for lang in ['en','zh-Hant']:
        d=OUT/f'normal-mark-{lang}';cmd=load(d/'command.json')
        assert cmd[cmd.index('-flow-script')+1].split()==expected_keys and len(expected_keys)==109
        assert cmd[cmd.index('-seed')+1]=='1'
        trace=load(d/'trace.json')['steps'];assert len(trace)==110
        assert trace==load(OUT/f'normal-cjk-figure-{lang}/trace.json')['steps'];traces.append(trace)
        for step in range(110):
            current=read_png(d/f'step-{step:03d}.png');previous=read_png(OUT/f'normal-cjk-figure-{lang}'/f'step-{step:03d}.png')
            assert (current.width,current.height)==((320,200) if lang=='en' else (960,600))
            frame=step-6
            if frame not in FRAMES:
                assert current.rgb()==previous.rgb(),(lang,step,'非目標頁改變')
            elif lang=='zh-Hant':
                expected=bytearray(previous.rgb())
                for y in range(162,180):
                    a=(y*960+464)*3;expected[a:a+16*3]=bytes(16*3)
                    a=(y*960+504)*3;expected[a:a+16*3]=want[(y-162)*16*3:(y-161)*16*3]
                assert current.rgb()==expected,(frame,'中文字模完整畫面不符')
                assert current.crop(504,162,16,18).rgb()==want
                assert current.crop(464,162,16,18).rgb()==bytes(16*18*3)
                assert previous.rgb()!=current.rgb(), '前版偏移未拒絕'
                wrong=bytearray(want);x,y=ink[0];wrong[(y*16+x)*3:(y*16+x+1)*3]=bytes(3)
                assert wrong!=current.crop(504,162,16,18).rgb()
            else:
                original=read_png(OUT/'gui-actions1'/f'item-{frame:02d}.png')
                assert original.crop(110,39,100,22).rgb()==current.crop(110,39,100,22).rgb()
                bad,_=legacy.diff(original,previous);assert len(bad)==23
                pts,_=legacy.diff(original,current);assert pts==[[15,7]]
                a,b=previous.rgb(),current.rgb()
                changes=[[n%320,n//320] for n in range(64000) if a[n*3:n*3+3]!=b[n*3:n*3+3]]
                assert len(changes)==22 and all(170<=x<=180 and 54<=y<=58 for x,y in changes)
                raw=bytes.fromhex(states[frame]['character_record_hex'])
                skill_codes=[raw[82+13*n] for n in range(4)]
                assert skill_codes==[87,0,0,0] and raw[61]==87
                checks.append(dict(original_frame=frame,remake_step=step,skill='1',source_codes=skill_codes,
                                   english_x=168,chinese_x=504,y=54,original_mark_rect=[110,39,100,22],
                                   mark_pixel_differences=0,previous_text_differences=22,remaining_full_frame_pixels=pts))
                mark=bytearray(current.crop(110,39,100,22).rgb())
                first=next(n for n in range(2200) if 58<=n%100<66 and 15<=n//100<21 and mark[n*3:n*3+3]==b'\xff\xff\xff')
                mark[first*3:first*3+3]=bytes(3);assert mark!=original.crop(110,39,100,22).rgb()
            if lang=='en' and 0<=frame<104:
                pts,cursor=legacy.diff(read_png(OUT/'gui-actions1'/f'item-{frame:02d}.png'),current);assert cursor==67
                if pts:
                    assert len(pts) in (1,8) and all(x==15 for x,y in pts)
                    pending.append(dict(original_frame=frame,remake_step=step,pixels=pts,status='未通過，未加入允許區'))
                else:zero.append(frame)
            if frame in FRAMES:
                copies.append((d/f'step-{step:03d}.png',f'mark-{lang}-{step:03d}.png'))
                copies.append((OUT/f'normal-cjk-figure-{lang}'/f'step-{step:03d}.png',f'mark-before-{lang}-{step:03d}.png'))
    assert traces[0]==traces[1] and len(checks)==3 and len(pending)==70 and len(zero)==34
    for f in FRAMES:copies.append((OUT/'gui-actions1'/f'item-{f:02d}.png',f'mark-original-{f:02d}.png'))
    persistence=[]
    for r in old['persistence_checks']:
        baseline=ROOT/r['path'];assert sha(baseline)==r['sha256']
        phase,frame,lang=r['phase'],r['frame'],r['language']
        p=OUT/f'{phase}-mark-{frame:02d}-{lang}/save.json'
        a,b=load(p),load(baseline);a.pop('saved_at');b.pop('saved_at');assert a==b
        persistence.append(dict(frame=frame,language=lang,phase=phase,path=path(p),sha256=sha(p),previous_sha256=sha(baseline),
                                complete_persistence_equal_except_saved_at=True))
        if phase=='reload':
            d=OUT/f'reload-mark-{frame:02d}-{lang}';cmd=load(d/'command.json')
            step=int(cmd[cmd.index('-shot-steps')+1].split(',')[0]);p=d/f'step-{step:03d}.png'
            assert read_png(p).rgb()==read_png(OUT/f'reload-cjk-figure-{frame:02d}-{lang}'/p.name).rgb()
            copies.append((p,f'mark-reload-{frame:02d}-{lang}.png'))
    assert len(persistence)==24
    attachments=[]
    for src,name in copies:
        target=PROOF/name
        if args.export:assert not target.exists();shutil.copyfile(src,target)
        assert sha(target)==sha(src)
        attachments.append(dict(path=name,sha256=sha(target),source_path_private=path(src),rights='private-repo可保存；不得公開散布'))
    inputs={r['path']:sha(ROOT/r['path']) for r in old['current_inputs']}
    for p in [Path(__file__),ROOT/'docs/re/item-mark-address-index.json']:inputs[path(p)]=sha(p)
    receipt=dict(schema='kol-item-mark-v1',date='2026-10-01',status='手持小框技能列號座標窄切片通過',
                 binary=dict(path=path(OUT/'kol-items-mark-linux'),sha256=AFTER),before_binary_sha256=BEFORE,
                 source_zip_sha256=old['source_zip_sha256'],source_executable_sha256=old['source_executable_sha256'],
                 original_revision=old['original_revision'],original_seed=1,remake_seed=1,seed_setup=old['seed_setup'],
                 normal_startup=True,state_injection=False,direct_entry=False,original_keys=old['original_keys'],remake_keys_count=109,
                 language_states_equal=110,other_english_frames_unchanged=107,other_chinese_frames_unchanged=107,
                 checks=checks,unverified_list_frames=pending,all_zero_frames=zero,persistence_checks=persistence,
                 current_inputs=[dict(path=p,sha256=h) for p,h in sorted(inputs.items())],attachments=attachments,
                 source_receipt=dict(path='cjk-figure-receipt.json',sha256=sha(PROOF/'cjk-figure-receipt.json')),
                 annotations=dict(input=annotated['input'],sha256=annotated['sha256'],database_sha256=annotated['database_sha256'],
                                  tool=annotated['tool'],address_space=annotated['address_space'],new_confirmed=7),
                 negative_controls=['前版英文22像素偏移拒絕','前版繁中x464偏移拒絕','英／繁中文字模缺墨點拒絕'],
                 tool=dict(go='1.26.7',ebitengine='2.9.9',image='eob-remake-go:1.26.7-ebiten2.9.9-audio'),
                 verifier_sha256=sha(Path(__file__)),
                 limits='只閉合Theodoric此正常手持框的座標；英文70張x15續畫差異、其他造型與全部物品動作尚未完成。原版無繁中畫面。')
    target=PROOF/'mark-receipt.json'
    if args.export:assert not target.exists();target.write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    else:assert load(target)==receipt,'收據重算不符'
    print(json.dumps(dict(status='通過',marks=3,unverified_list_frames=70,persistence=24,attachments=len(attachments)),ensure_ascii=False))
if __name__=='__main__':main()
