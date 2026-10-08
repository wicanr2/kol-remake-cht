#!/usr/bin/env python3
"""Docker 內唯讀重驗 RE-201。--seal 只在 /out 建立新收據。"""
from pathlib import Path
import argparse,contextlib,copy,hashlib,io,json,os,runpy
ROOT=Path(__file__).resolve().parents[3]
PRIVATE=ROOT/'workplace/spell-effect-20261003'
HERE=Path(__file__).resolve().parent
read=lambda p:json.loads(p.read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
def validate(deep=False):
    parent=read(PRIVATE/'parent-preservation.json')
    p=ROOT/parent['parent_receipt'];assert sha(p)==parent['parent_receipt_sha256']
    receipt=read(p);sources=parent['sources']
    for n,h in sources.items():assert sha(PRIVATE/'parent-inputs'/n)==h,n
    for n,h in receipt['attachments'].items():assert sha(PRIVATE/'parent-inputs'/n if n in sources else ROOT/n)==h,n
    prior=read(PRIVATE/'parent-check.json');assert prior['deep_parent_passed']
    assert prior['parent_receipt_sha256']==sha(p) and prior['parent_attachments']==6635
    assert len(prior['code_inputs'])==319 and prior['changed_code_inputs']==[]
    for n,h in prior['code_inputs'].items():assert sha(ROOT/n)==h,n
    if deep:
        class ArchivedRoot:
            def __truediv__(self,n):return PRIVATE/'parent-inputs'/n if str(n) in sources else ROOT/n
        previous=runpy.run_path(str(ROOT/'docs/verification/native-scroll-20261003/verify.py'))['validate']
        previous.__globals__['ROOT']=ArchivedRoot()
        assert json.loads(json.dumps(previous()))=={k:v for k,v in receipt.items()if k!='attachments'}
    os.environ['KOL_VERIFY_ROOT']=str(ROOT)
    output=io.StringIO()
    with contextlib.redirect_stdout(output):runpy.run_path(str(PRIVATE/'check_exit.py'))
    departure=json.loads(output.getvalue())
    for name in ['exit1','effect1','effect2']:
        case=PRIVATE/name;d=read(case/'driver.json')
        for file,key in [('native-driver.go','native_driver_sha256'),('instrumented-book-private.go','startup_sha256'),('keys.txt','keys_sha256')]:assert sha(case/file)==d[key]
        driver=(case/'native-driver.go').read_text();assert driver.count('.WriteU16(')==2
        assert '.Stub(' not in driver and '.Call(' not in driver
        assert d['normal_startup'] and not d['state_injection'] and not d['wait_stub']
    failure=read(PRIVATE/'failure-review.json')
    assert not failure['native_receipt_written'] and failure['town_images']==239 and failure['battle_images']==0
    assert 'signal: killed' in (PRIVATE/'effect1/run.log').read_text()
    assert sha(PRIVATE/failure['reviewed_image'])==failure['reviewed_image_sha256']
    assert sha(PRIVATE/failure['last_image'])==failure['last_image_sha256']
    native=read(PRIVATE/'effect2/native-shop-private.json')
    assert len(native['rows'])==175 and native['rows'][:172]==read(PRIVATE/'exit1/native-shop-private.json')['rows']
    assert native['stopped_for_battle'] and not native['stopped_for_healing_question'] and native['reached_purchase']
    assert [x['key']for x in native['rows'][-3:]]==['Right','Right','Space']
    assert [x['record']for x in native['rows'][173]['printed']][-2:]==[0xE7,0xE8]
    battle=read(PRIVATE/'effect2/purchased-effect-private.json')
    assert len(battle['inputs'])==99 and battle['rounds']==4 and battle['answered_area'] and battle['applied_area']
    assert not battle['reached_battle_exit'] and not any(battle[k]for k in ['direct_entry','state_injection','wait_stub'])
    last=battle['inputs'][-1];before=last['state'];after=battle['final']
    assert last['key']=='2' and last['round']==4 and last['last_print']==0x16C
    assert before['actors']==after['actors'] and before['characters']==after['characters']
    assert before['counters']==after['counters'] and before['seed']==after['seed']==2128043714
    caster=bytes.fromhex(before['actors'][0]);target=bytes.fromhex(before['actors'][1])
    assert caster[0x27]==0 and target[0x13]==target[0x14]==16
    assert battle['wait_calls'][-1]['caller_ida']=='0x2855C' and battle['wait_calls'][-1]['last_print']==0x16C
    assert [(x['record'],x['caller_ida'])for x in battle['messages'][-4:]]==[(0x184,'0x284FC'),(0x185,'0x2853B'),(0x16C,'0x2854B'),(0x16D,'0x287F6')]
    lastrolls=battle['rolls'][-2:]
    assert [(x['n'],x['seed_before'],x['caller_ida'])for x in lastrolls]==[(2,3538685960,'0x283C7'),(11,1807519403,'0x28435')]
    cast=read(PRIVATE/'cast-check.json');state=3538685960;expected=[]
    for die in [2,11]:
        old=state;state=(state*0x343FD+0x269EC3)&0xffffffff
        expected.append(dict(n=die,seed_before=old,value=((state>>16)&0x7fff)%die+1,seed_after=state))
    assert cast['rolls']==expected and cast['turns']==1 and cast['power']==11 and cast['area_seed']==state==2128043714
    assert not cast['negative'] and cast['chest_before']==cast['chest_after']==16
    assert cast['body_parts']==[int(target[x])for x in [0x11,0x13,0x15,0x17,0x19,0x1b]]
    observed=read(PRIVATE/'remake-entry-en/trace.json')['steps'];formal=read(PRIVATE/'formal-entry-en/trace.json')['steps']
    assert len(observed)==len(formal)==184
    assert observed[182]['battle']['stage']==1 and observed[183]['battle']['stage']==2
    assert observed[182]['battle']['monster']==observed[183]['battle']['monster']=='CLIFF TROLL'
    plain=copy.deepcopy(observed)
    for row in plain:
        row.get('town',{}).pop('probe_full_party',None)
        for k in ['probe_group','probe_weather','probe_world_seed']:row.get('world',{}).pop(k,None)
        for k in list(row.get('battle',{})):
            if k.startswith('probe_'):row['battle'].pop(k)
    assert plain==formal
    for i in range(184):assert sha(PRIVATE/'remake-entry-en'/f'step-{i:03d}.png')==sha(PRIVATE/'formal-entry-en'/f'step-{i:03d}.png')
    for i in range(2):
        a=native['rows'][172+i]['state'];b=observed[181+i]['world']
        assert a['camera']==[b['x'],b['y']] and a['marker']==[b['sub_x'],b['sub_y']]
        assert (a['time'],a['part'],a['seed'])==(b['ticks'],b['time_index'],b['probe_world_seed'])
    original_positions=[[a[31],a[32]]for h in battle['initial']['actors']if (a:=bytes.fromhex(h))[1]==1 and a[17]>0]
    remake_positions=[[a['X'],a['Y']]for a in observed[183]['battle']['probe_actors']if a['Side']==1 and a['Head']>0]
    assert original_positions==[[35,32],[37,29],[36,34]]
    assert remake_positions==[[36,32],[36,30],[36,34]]
    assert battle['initial']['seed']==2535178238 and observed[183]['battle']['probe_seed']==3707996875
    positive=[]
    for lang in ['en','zh']:
        case=PRIVATE/('positive2-'+lang);trace=read(case/'trace.json')['steps']
        assert len(trace)==len(list(case.glob('step-*.png')))==302
        assert len(read(case/'command.json')[read(case/'command.json').index('-flow-script')+1].split())==301
        area=trace[298]['battle'];final=trace[301]['battle']
        assert trace[194]['battle']['probe_seed']==1
        assert area==trace[299]['battle']==trace[300]['battle']
        assert area['prompt']==12 and area['round']==6 and not area['probe_actors'][0]['Negative']
        assert area['probe_seed']==final['probe_seed']==410959018 and area['probe_rolls']==final['probe_rolls']
        assert area['probe_actors']==final['probe_actors'] and area['probe_actors'][1]['Chest']==16
        assert ('Select an area to heal' if lang=='en' else '選擇要治療的部位') in ' '.join(area['exec_text'])
        for row in trace:
            for key in ['exec_text','encounter_text']:row.get('battle',{}).pop(key,None)
        positive.append(trace)
    assert positive[0]==positive[1]
    return dict(schema='kol-purchased-spell-effect-receipt-v1',parent_receipt_sha256=sha(p),parent_attachments=6635,
                code_inputs=319,changed_code_inputs=[],departure=departure,
                normal_original=dict(prefix_states=175,battle_keys=99,round=4,negative=False,chest_before=16,chest_after=16,seed=2128043714),
                local_rule=dict(turns=1,power=11,rolls=expected,seed=state,chest=16),
                positive_ui=dict(keys_per_language=301,states_per_language=302,seed=410959018,first_planning_controlled_seed=1),
                encounter=dict(world_steps=2,formal_images=184,x=42,y=9,sub_x=5,sub_y=3,ticks=3340,part=1,seed=316395082,monster='CLIFF TROLL',count=3,original_positions=original_positions,remake_positions=remake_positions,original_planning_seed=2535178238,remake_planning_seed=3707996875),
                scope='限定正常出城、正常原版治療AREA、局部同狀態規則及雙語治療UI；遭遇種類／隻數相同，進場位置及seed不同，不宣稱完整戰鬥parity')
def attachments():
    paths=[ROOT/n for n in ['docs/re/20261003-purchased-spell-effect.md','CONTEXT.md','WORKLOG.md','WORKLIST.md','docs/worklist.json','docs/spec/012-combat-spells.md','docs/re/RESEARCH-LOG.md']]+[HERE/'verify.py']
    paths+=[p for p in PRIVATE.rglob('*')if p.is_file() and 'game' not in p.relative_to(PRIVATE).parts and p.name not in ['receipt.json','audit.json']]
    return {str(p.relative_to(ROOT)):sha(p)for p in sorted(set(paths))}
if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--seal',action='store_true');parser.add_argument('--deep',action='store_true');args=parser.parse_args()
    result=validate(deep=args.deep)
    if args.seal:
        result['attachments']=attachments();target=Path('/out/receipt.json');assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        stored=read(HERE/'receipt.json')
        assert result=={k:v for k,v in stored.items()if k!='attachments'}
        for n,h in stored['attachments'].items():assert sha(ROOT/n)==h,n
    print(json.dumps(result if not args.seal else dict(schema=result['schema'],attachments=len(result['attachments']),departure=result['departure']),ensure_ascii=False))
