"""RE-192原版階段觀察及隔離追蹤的封存／唯讀重驗。"""
import argparse,collections,hashlib,json,os,runpy
from pathlib import Path
ROOT=Path('/src')
PRIVATE=ROOT/'workplace/return-tracker-phases-20261002'
HERE=ROOT/'docs/verification/return-tracker-phases-20261002'
PARENT=ROOT/'docs/verification/return-tracker-20261002'
def read(p):return json.loads(p.read_text())
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def validate():
    preservation=read(PRIVATE/'parent-preservation.json');saved=preservation['sources']
    assert preservation['schema']=='kol-return-tracker-phases-parent-v1' and len(saved)==1344
    for name,digest in saved.items():assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parent=read(PARENT/'receipt.json')
    assert sha(PARENT/'receipt.json')==preservation['receipt_sha256']
    for name,digest in parent['attachments'].items():
        assert sha(PRIVATE/'parent-inputs'/name if name in saved else ROOT/name)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    old=runpy.run_path(str(PARENT/'verify_tracker.py'))['validate'];old.__globals__['ROOT']=ArchivedRoot()
    assert old()=={k:v for k,v in parent.items() if k!='attachments'}
    code=[n for n in saved if n.startswith(('cmd/','internal/','tools/'))]
    assert len(code)==314
    for name in code:assert sha(ROOT/name)==saved[name],name
    production=ROOT/'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(production)=='dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    source=PRIVATE/'source'
    s=read(source/'complete-source-private.json')
    original=read(ROOT/'workplace/spell-return-20261002/source3/complete-source-private.json')
    assert sha(source/'complete-source-private.json')=='0694189048f33a8fec722cbda5c1726cb87d0922775f3e0755291883769bbc0b'
    for k in original:
        if k not in ['cost_flow']:assert s[k]==original[k],k
    addresses={x['ida'] for x in original['cost_flow']}
    prior=[{k:v for k,v in r.items() if k not in ['regs','local_index','arg_actor']} for r in s['cost_flow'] if r['ida'] in addresses]
    assert prior==original['cost_flow']
    assert s['normal_startup'] and not any(s[k] for k in ['direct_entry','state_injection','wait_stub'])
    assert len(s['inputs'])==921 and s['rounds']==42 and len(s['post_battle'])==8
    assert s['seed']==s['initial']['seed']==1
    assert len(s['cost_flow'])==944 and len(s['return_writes'])==561
    ws=s['return_writes'];word={off:next(r['old'] for r in ws if r['off']==off) for off in ['0xbb84','0xbb85']};at=0
    for r in s['cost_flow']:
        while at<len(ws) and ws[at]['step']<=r['state']['steps']:
            word[ws[at]['off']]=ws[at]['new'];at+=1
        assert word['0xbb84']+256*word['0xbb85']==r['return_index']
    wrong=0
    quick=[r for r in s['cost_flow'] if r['ida']=='0x1fd9d']
    assert len(quick)==126 and all(r['local_index']!=r['actor'] for r in quick)
    for r in quick:
        i=r['actor'];a=bytes.fromhex(r['state']['actors'][i]);c=r['state']['counters'][i]
        assert c['fatigue']>=0
        ratio=50*c['fatigue']//c['stamina'] if c['stamina']>0 else 0
        wrong+=a[9]*ratio//200!=r['return_index']
    assert wrong==94
    drop=[r for r in s['cost_flow'] if r['ida']=='0x20234'];forced=[r for r in s['cost_flow'] if r['ida']=='0x26a68']
    assert len(drop)==len(forced)==1 and drop[0]['actor']==1 and drop[0]['return_index']==816
    assert forced[0]['arg_actor']==0 and forced[0]['local_index']==forced[0]['return_index']==1
    cases=read(PRIVATE/'phase-cases-private.json')
    assert cases['source_sha256']==sha(source/'complete-source-private.json') and len(cases['cases'])==903
    rows=[r for r in s['cost_flow'] if r['ida']!='0x21045'];clear=0;ground=[]
    for c,r in zip(cases['cases'],rows):
        ea=r['ida'];i=r['local_index'] if ea in ['0x1fd9d','0x1fe53','0x1fede'] else r['arg_actor'] if ea=='0x26a68' else r['actor']
        assert (c['ida'],c['actor'],c['want'],c['round'])==(ea,i,r['return_index'],r['round'])
        assert c['raw']==r['state']['actors'][i]
        if 'character' in c:assert c['character']==r['state']['characters'][i]
        if 'counter' in c:assert c['counter']==r['state']['counters'][i]
        if 'actors' in c:assert c['actors']==r['state']['actors'] and c['characters']==r['state']['characters']
        if ea in ['0x20234','0x26a68']:assert c['ground_before']==ground
        if ea=='0x20234':ground.append(dict(slot=r['local_index'],record=r['state']['characters'][i][0x30*2:0x40*2]))
        if ea=='0x21821':assert c['clear_iteration']==clear%17;clear+=1
    assert clear==85
    for filename,needle in [('prototype.log','--- PASS: TestReturnPhaseNativeObservations'),('battle-suite.log','ok  '),('source/run.log','--- PASS: TestNormalBookEntry')]:
        log=(PRIVATE/filename).read_text();assert needle in log and 'FAIL' not in log
    assert 'cannot use []int' in (PRIVATE/'prototype.log.attempt1').read_text()
    assert 'no such file or directory' in (PRIVATE/'battle-suite.log.attempt1').read_text()
    driver=read(source/'driver.json');assert driver['parent_sha256']==sha(ROOT/driver['parent_source'])
    assert driver['driver_sha256']==sha(source/'instrumented-phases-private.go')
    w=read(source/'world-return-private.json');last=w['events'][-8:]
    assert sha(source/'world-return-private.json')=='0626a51fdeabb5a5f344beccdc8ed27b53b0bfb510e6bd7647cd2315a143a84f'
    assert [r['state']['time'] for r in last]==[3360]*6+[3260]*2
    assert all((r['state']['raw_bb84'],r['state']['raw_12b8_indexed'],r['state']['movement'])==(16,30,20) for r in last)
    return dict(schema='kol-return-tracker-phases-receipt-v1',historical_shared_inputs=1344,parent_attachments=len(parent['attachments']),parent_results_recomputed=True,unchanged_code_inputs=314,production_sha256=sha(production),native_observations=944,prototype_comparisons=903,stale_global_index_value_counterexamples=94,old_normal_inputs_unchanged=True,prototype_only=True,battle_suite='passed',world_cost=100,world_times=[3360,3260],status='返回階段903個局部比較通過；正式成本與正常重製GUI及保存仍待驗。')
def attachments():
    names=set(read(PRIVATE/'parent-preservation.json')['sources'])
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'docs/re/20261002-return-tracker-phases.md','docs/verification/return-tracker-phases-20261002/verify_phases.py'}
    return {name:sha(ROOT/name) for name in sorted(names)}
if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        result['attachments']=attachments();target=Path('/out/receipt.json');assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt=read(HERE/'receipt.json')
        for name,digest in receipt['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in receipt.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json');assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            bad,md=[],[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0:bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'):md.append(str(p))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat();assert (q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-return-tracker-phases-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),attachments=len(receipt['attachments']),full_readonly_verification='passed',host_control=control,root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000],production_unchanged=True),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
