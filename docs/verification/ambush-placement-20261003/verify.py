#!/usr/bin/env python3
"""Docker /src 唯讀重驗 RE-202；--seal 僅在 /out 寫入首次收據。"""
from pathlib import Path
import argparse,copy,hashlib,json
ROOT=Path(__file__).resolve().parents[3]
PRIVATE=ROOT/'workplace/ambush-placement-20261003'
HERE=Path(__file__).resolve().parent
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()

def validate():
    preservation=read(PRIVATE/'parent-preservation.json');parent=ROOT/preservation['parent_receipt']
    assert sha(parent)==preservation['parent_receipt_sha256']
    prior=read(parent);sources=preservation['sources']
    for name,h in sources.items():assert sha(PRIVATE/'parent-inputs'/name)==h
    for name,h in prior['attachments'].items():assert sha(PRIVATE/'parent-inputs'/name if name in sources else ROOT/name)==h,name
    deep=read(PRIVATE/'parent-check.json')
    assert deep['deep_parent_passed'] and deep['parent_receipt_sha256']==sha(parent) and deep['parent_attachments']==2317
    assert deep['archived_sources']==sources and prior['code_inputs']==319
    before=read(PRIVATE/'production-inputs-before.json')
    assert before==deep['code_inputs']==read(ROOT/'workplace/spell-effect-20261003/parent-check.json')['code_inputs'] and len(before)==319
    inputs={name:sha(ROOT/name)for name in before}
    changed=sorted(name for name,h in inputs.items() if before[name]!=h)
    assert changed==['cmd/kol/main.go','internal/battle/ambush.go','internal/battle/ambush_test.go','internal/battle/state.go']
    assert '進場位置槽 6' in (PRIVATE/'old-placement-negative.log').read_text()
    assert 'PASS: TestNormalAmbushPlacementOriginal' in (PRIVATE/'placement-targeted.log').read_text()
    rules=(PRIVATE/'rules.log').read_text();assert 'FAIL' not in rules and 'internal/battle' in rules and 'cmd/kol' in rules
    review=read(PRIVATE/'evidence-review.json');assert review['checked_instructions']==814 and review['candidate_comparisons']==24
    original=ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked.EXE'
    assert sha(original)=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    exe=original.read_bytes();header=int.from_bytes(exe[8:10],'little')*16;loaded=bytearray(exe)
    for i in range(int.from_bytes(exe[6:8],'little')):
        p=int.from_bytes(exe[24:26],'little')+4*i;at=header+int.from_bytes(exe[p:p+2],'little')+16*int.from_bytes(exe[p+2:p+4],'little')
        loaded[at:at+2]=((int.from_bytes(exe[at:at+2],'little')+0x1000)&0xffff).to_bytes(2,'little')
    for row in review['checks']:
        off=int(row['file_offset'],16);b=bytes.fromhex(row['ida_bytes'])
        assert int(row['ida_linear'],16)-0x10000+header==off
        assert loaded[off:off+len(b)]==b and exe[off:off+len(b)].hex()==row['file_bytes']
    native=read(PRIVATE/'native3/native-shop-private.json');old=read(ROOT/'workplace/spell-effect-20261003/effect2/native-shop-private.json')
    assert native['rows']==old['rows'] and len(native['rows'])==175
    assert native['normal_startup'] and not any(native[k]for k in ['state_injection','direct_entry','wait_stub'])
    assert native['unimplemented']==old['unimplemented']==['int 21h AH=0E AL=38 ×1']
    for i in range(175):assert sha(PRIVATE/'native3'/f'town-{i:03d}.png')==sha(ROOT/'workplace/spell-effect-20261003/effect2'/f'town-{i:03d}.png')
    driver=read(PRIVATE/'native3/driver.json')
    for name,key in [('instrumented-book-private.go','startup_sha256'),('native-driver.go','driver_sha256'),('keys.txt','keys_sha256')]:assert sha(PRIVATE/'native3'/name)==driver[key]
    text=(PRIVATE/'native3/native-driver.go').read_text();assert text.count('.WriteU16(')==2 and '.Stub(' not in text and '.Call(' not in text
    assert driver['keys_sha256']==sha(ROOT/'workplace/spell-effect-20261003/effect2/keys.txt')
    assert driver['startup_sha256']==sha(ROOT/'workplace/spell-effect-20261003/effect2/instrumented-book-private.go')
    phases=native['placement'];assert [x['label']for x in phases]==['before-placement','after-placement','before-first-planning','normal-first-party-prompt']
    local=read(PRIVATE/'placement-replay.json');assert len(local['candidates'])==24 and len(native['candidates'])==48
    assert (phases[0]['seed'],phases[1]['seed'])==(local['seed_before'],local['seed_after'])==(2052235245,4058265942)
    assert phases[0]['field_hex']==phases[1]['field_hex']==local['field_raw_hex']
    for i,c in enumerate(local['candidates']):
        a,b=native['candidates'][2*i:2*i+2]
        assert (a['phase'],b['phase'])==('before','after')
        assert a['slot']==b['slot']==c['slot'] and a['direction']==b['direction']==c['direction']
        assert [a['x'],a['y']]==c['before'] and [b['x'],b['y']]==c['after']
        assert [b['candidate_x'],b['candidate_y']]==c['candidate']
    expected=[[bytes.fromhex(phases[1]['actors'][i])[31],bytes.fromhex(phases[1]['actors'][i])[32]]for i in [0,1,6,7,8]]
    assert expected==[[33,30],[31,30],[35,32],[37,29],[36,34]]
    raw=bytes.fromhex(phases[0]['field_hex']);assert len(raw)==4096
    runs=[]
    prior_steps=read(ROOT/'workplace/spell-effect-20261003/remake-entry-en/trace.json')['steps']
    for lang in ['en','zh-Hant']:
        a=PRIVATE/f'observed-entry-{lang}';b=PRIVATE/f'formal-entry-{lang}';observed=read(a/'trace.json')['steps'];formal=read(b/'trace.json')['steps']
        assert len(observed)==len(formal)==184
        plain=copy.deepcopy(observed)
        for row in plain:
            row.get('town',{}).pop('probe_full_party',None)
            for key in ['probe_group','probe_weather','probe_world_seed']:row.get('world',{}).pop(key,None)
            for key in list(row.get('battle',{})):
                if key.startswith('probe_'):row['battle'].pop(key)
        assert plain==formal
        for i in range(184):assert sha(a/f'step-{i:03d}.png')==sha(b/f'step-{i:03d}.png')
        battle=observed[-1]['battle'];actors=battle['probe_actors']
        assert battle['stage']==2 and battle['round']==1 and battle['cursor']==1 and battle['monster']=='CLIFF TROLL'
        assert sum(x['Side']==1 for x in actors)==3
        assert [[actors[i]['X'],actors[i]['Y']]for i in [0,1,6,7,8]]==expected
        for i,tile in enumerate(raw):assert battle['probe_field']['Tiles'][(i%64)*64+i//64]==tile
        assert battle['probe_seed']==1323149826 and phases[-1]['seed']==2535178238
        if lang=='en':
            assert observed[:182]==prior_steps[:182]
            for i in range(183):assert sha(a/f'step-{i:03d}.png')==sha(ROOT/'workplace/spell-effect-20261003/remake-entry-en'/f'step-{i:03d}.png')
        for key in ['x','y','sub_x','sub_y','ticks','time_index','probe_world_seed']:assert observed[182]['world'][key]==prior_steps[182]['world'][key]
        runs.append(dict(language=lang,normal_keys=183,states=184,formal_png_equal_observer=184,field_cells=4096,positions=expected,current_actor=1,planning_seed=1323149826,original_first_planning_seed=2535178238))
    visual=read(PRIVATE/'field-visual.json')
    assert visual['native_sha256']==sha(PRIVATE/'native3/town-174.png') and visual['bounds']==[152,8,296,152]
    assert [(x['language'],x['different_pixels'],x['scale'])for x in visual['runs']]==[('en',0,1),('zh-Hant',55,3)]
    for x in visual['runs']:assert x['formal_sha256']==sha(PRIVATE/f"formal-entry-{x['language']}"/'step-183.png') and x['pixels']==20736
    formal=sha(PRIVATE/'kol-placement-linux');observer=sha(PRIVATE/'kol-placement-observer-linux')
    assert formal=='eed01abbbff335e4a0944d670a61c668d0be5ce5a93fbea12f2e9561e273b043'
    assert observer=='b5d8912e17b2461880ab5740aa158c19e95af8b9bc580fc46e9a2cd0f9fef1f4'
    return dict(schema='kol-normal-ambush-placement-v1',parent_receipt_sha256=sha(parent),parent_attachments=2317,deep_parent_recalculated=True,code_inputs=inputs,changed_code_inputs=changed,formal_sha256=formal,observer_sha256=observer,native_states=175,native_png_unchanged=175,candidate_steps=24,checked_instructions=814,runs=runs,visual=visual,scope='本場正常伏擊地形亂數交接與位置校正有限CONFORMED；完整首次規劃、裝備變體、其他入口、新保存與完整remake仍未完成')

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--seal',action='store_true');args=parser.parse_args()
    result=validate()
    if args.seal:
        path=Path('/out/receipt.json');assert not path.exists()
        names=['CONTEXT.md','WORKLOG.md','WORKLIST.md','docs/worklist.json','docs/re/RESEARCH-LOG.md','docs/re/20261003-ambush-placement.md','docs/spec/011-wilderness-ambush.md','docs/verification/ambush-placement-20261003/verify.py','docs/verification/ambush-placement-20261003/audit.py']
        attachments={name:sha(ROOT/name)for name in names}
        attachments.update(result['code_inputs'])
        for p in sorted(PRIVATE.rglob('*')):
            if p.is_file():attachments[str(p.relative_to(ROOT))]=sha(p)
        result['attachments']=attachments
        path.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        stored=read(HERE/'receipt.json')
        for name,h in stored['attachments'].items():assert sha(ROOT/name)==h,name
        assert result=={k:v for k,v in stored.items()if k!='attachments'}
    print(json.dumps({k:v for k,v in result.items()if k not in ['code_inputs','attachments']},ensure_ascii=False))

if __name__=='__main__':main()
