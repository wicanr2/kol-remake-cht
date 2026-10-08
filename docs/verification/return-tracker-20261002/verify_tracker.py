"""RE-191隔離追蹤與正常規劃後逃走來源的封存／唯讀重驗。"""
import argparse,hashlib,json,os,runpy
from pathlib import Path

ROOT=Path('/src')
PRIVATE=ROOT/'workplace/return-tracker-20261002'
HERE=ROOT/'docs/verification/return-tracker-20261002'
PARENT=ROOT/'docs/verification/spell-return-20261002'

def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()

def validate():
    preservation=read(PRIVATE/'parent-preservation.json')
    saved=preservation['sources']
    assert preservation['schema']=='kol-return-tracker-parent-v1' and len(saved)==1340
    for name,digest in saved.items(): assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parent=read(PARENT/'receipt.json')
    assert sha(PARENT/'receipt.json')==preservation['receipt_sha256']
    for name,digest in parent['attachments'].items():
        p=PRIVATE/'parent-inputs'/name if name in saved else ROOT/name
        assert sha(p)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):
            return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    old=runpy.run_path(str(PARENT/'verify_spell_return.py'))['validate']
    old.__globals__['ROOT']=ArchivedRoot()
    assert old()=={k:v for k,v in parent.items() if k!='attachments'}
    code_names=[n for n in saved if n.startswith(('cmd/','internal/','tools/'))]
    assert len(code_names)==314
    for name in code_names: assert sha(ROOT/name)==saved[name],name
    production=ROOT/'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(production)=='dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    d=read(PRIVATE/'scratch-xrefs-ida-private.json')
    assert d['schema']=='kol-return-tracker-ida-v1' and d['function_count']==338
    assert d['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert d['database_sha256']=='5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    refs=d['scratch_direct_refs']
    assert len(refs)==120 and all(r['target']=='0x39a54' for r in refs)
    ins={i['original_ea']:i for r in refs for i in r['context']}
    assert len(ins)==1197
    for ea,raw in {'0x1f502':'a384bb','0x1f60d':'a384bb',
                   '0x1f186':'c70684bb0000','0x1f1e4':'830684bb01'}.items():
        assert ins[ea]['bytes']==raw
    cases=read(PRIVATE/'fatigue-tracker-cases-private.json')
    assert cases['schema']=='kol-fatigue-tracker-cases-v1' and len(cases['cases'])==107
    assert sum(r['ida']=='0x20bad' for r in cases['cases'])==105
    assert sum(r['ida']=='0x20d1b' for r in cases['cases'])==2
    for name,digest in cases['sources'].items(): assert sha(ROOT/name)==digest,name
    native_sources={name:read(ROOT/name) for name in cases['sources']}
    for r in cases['cases']:
        p=native_sources[r['source']]
        native=[n for n in p['cost_flow'] if n['ida']==r['ida'] and n['round']==r['round'] and n['actor']==r['actor']]
        assert len(native)==1 and native[0]['return_index']==r['want']
        raw=bytes.fromhex(native[0]['state']['actors'][r['actor']])
        assert [r['command'],r['style'],r['guard'],r['flag1e'],r['spell_pick']]==[raw[0x25],raw[0x23],raw[0x24],raw[0x1e],raw[0x26]]
        if r['spell']:
            sp=bytes.fromhex(native[0]['state']['characters'][r['actor']])[0x200+16*raw[0x26]:0x210+16*raw[0x26]]
            assert r['spell']==dict(Guards=list(sp[8:12]),Effect=sp[12],PowerCode=sp[14])
    log=(PRIVATE/'tracker-cases.log').read_text()
    assert '--- PASS: TestReturnTrackerNativeWrites' in log and 'FAIL' not in log
    assert '--- FAIL: TestReturnTrackerNativeWrites' in (PRIVATE/'tracker-cases.log.attempt2').read_text()
    assert read(PRIVATE/'preparation-attempt1.json')['game_or_tests_started'] is False
    source=PRIVATE/'menu-flee2'
    s,w=read(source/'complete-source-private.json'),read(source/'world-return-private.json')
    assert s['schema']=='kol-normal-menu-flee-v1'
    for payload in [s,w]:
        assert payload['normal_startup'] and not any(payload[k] for k in ['direct_entry','state_injection','wait_stub'])
        assert payload['revision']=='968ca774705853276ae55c8d1163c369481ac71f'
        assert payload['input_sha256']==d['sha256']
    assert s['seed']==s['initial']['seed']==1
    names=[bytes.fromhex(h).split(b'\0')[0] for h in s['initial']['characters'][:2]]
    assert names==[b'Warder',b'Olear']
    assert s['reached_battle_exit'] and s['rounds']==2
    assert len(s['inputs'])==43 and s['inputs'][-1]['key']=='y'
    assert s['battle_exit_counts']['rolls']==28 and len(s['post_battle'])==8
    assert [r['action'] for r in s['normal_plans']]==['DROP','DROP','CAST-NEGATIVE','REST']
    assert not any(r['ida']=='0x20d1b' for r in s['cost_flow'])
    assert bytes.fromhex(s['inputs'][32]['state']['actors'][0])[0x25:0x28]==bytes.fromhex('0d0001')
    writes=s['return_writes']
    assert any(r['ip_ida']=='0x1F505' and r['off']=='0xbb85' and r['new']==7 for r in writes)
    assert any(r['ip_ida']=='0x1F610' and r['off']=='0xbb84' and r['new']==1 for r in writes)
    assert writes[-1]['ip_ida']=='0x1F18C' and writes[-1]['new']==0
    last=w['events'][-8:]
    for r in last:
        q=r['state']
        assert (q['raw_bb84'],q['raw_12b8_indexed'],q['movement'])==(0,10,20)
        assert (q['x'],q['y'],q['sub_x'],q['sub_y'])==(49,23,4,4)
    assert last[5]['state']['time']==3360 and last[6]['state']['time']==3300
    assert last[3]['state']['seed']==1226606285 and last[4]['state']['seed']==1292271196
    native_log=(source/'run.log').read_text()
    assert '--- PASS: TestNormalBookEntry' in native_log and 'FAIL' not in native_log
    failed=read(PRIVATE/'menu-flee/complete-source-private.json')
    assert not failed['reached_battle_exit'] and len(failed['inputs'])==4500
    assert '--- FAIL: TestNormalBookEntry' in (PRIVATE/'menu-flee/run.log').read_text()
    driver=read(source/'driver.json')
    assert driver['parent_sha256']==sha(ROOT/driver['parent_path'])
    assert driver['generated_sha256']==sha(source/'instrumented-menu-flee-private.go')
    assert driver['preparation_sha256']==sha(PRIVATE/'prepare_menu_flee2.py')
    return dict(schema='kol-return-tracker-receipt-v1',historical_shared_inputs=1340,
                parent_attachments=len(parent['attachments']),parent_results_recomputed=True,
                unchanged_code_inputs=314,production_sha256=sha(production),
                ida_direct_references=120,ida_unique_instructions=1197,
                prototype_native_cases=107,prototype_only=True,
                normal_menu_flee_inputs=43,executed_rounds=1,planning_rounds=2,
                spell_executed=False,menu_scratch=0x701,final_scratch=0,
                world_cost=60,world_times=[3360,3300],
                status='正常選單殘值覆寫及107個隔離追蹤比較已證實；正式返回成本仍DRAFT。')

def attachments():
    names=set(read(PRIVATE/'parent-preservation.json')['sources'])
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'docs/re/20261002-return-scratch-tracker.md','docs/verification/return-tracker-20261002/verify_tracker.py'}
    return {name:sha(ROOT/name) for name in sorted(names)}

if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true')
    args=ap.parse_args();assert not(args.seal and args.audit)
    result=validate()
    if args.seal:
        result['attachments']=attachments()
        target=Path('/out/receipt.json');assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt=read(HERE/'receipt.json')
        for name,digest in receipt['attachments'].items(): assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in receipt.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json')
            assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            bad,md=[],[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p=Path(directory)/name
                    if p.lstat().st_uid==0: bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'): md.append(str(p))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat();assert (q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-return-tracker-audit-v1',receipt_sha256=sha(HERE/'receipt.json'),
                attachments=len(receipt['attachments']),full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000],production_unchanged=True),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
