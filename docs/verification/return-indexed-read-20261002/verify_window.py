"""RE-193正常記憶體窗重播、歷史保存與唯讀稽核。"""
import argparse
import hashlib
import json
import os
import runpy
from pathlib import Path

ROOT=Path('/src')
PRIVATE=ROOT/'workplace/return-indexed-read-20261002'
HERE=ROOT/'docs/verification/return-indexed-read-20261002'
PARENT=ROOT/'docs/verification/return-tracker-phases-20261002'
def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()

def replay(d,w):
    b=bytearray.fromhex(d['indexed_window_initial'])
    assert len(b)==0x1004
    writes=d['indexed_window_writes']; cursor=0; step=d['indexed_window_start_step'];rows=[]
    for event in w['events']:
        s=event['state']
        if event['ida']!='0x29800' or s['steps']<step: continue
        while cursor<len(writes) and writes[cursor]['step']<s['steps']:
            x=writes[cursor];assert x['step']>=step;step=x['step']
            off=int(x['off'],16)-0x12B7
            assert 0<=off<len(b) and b[off]==x['old']
            b[off]=x['new'];cursor+=1
        snapshot=bytes.fromhex(s['indexed_window_hex']);assert bytes(b)==snapshot
        address=(0x12B8+s['raw_bb84'])&0xFFFF
        assert 0x12B7<=address<0x22BB
        value=b[address-0x12B7];assert value==s['raw_12b8_indexed']
        assert int.from_bytes(b[0x1600-0x12B7:0x1604-0x12B7],'little')==s['seed']
        rows.append(dict(step=s['steps'],index=s['raw_bb84'],address=hex(address),
            indexed_byte=value,seed=s['seed'],cost=2*(value+s['movement']),
            writes_applied=cursor,window_sha256=hashlib.sha256(snapshot).hexdigest()))
    assert rows and cursor==len(writes)
    return rows

def validate():
    preserve=read(PRIVATE/'parent-preservation.json');saved=preserve['sources']
    assert preserve['schema']=='kol-return-indexed-read-parent-v1' and len(saved)==1348
    for name,digest in saved.items(): assert sha(PRIVATE/'parent-inputs'/name)==digest,name
    parent=read(PARENT/'receipt.json')
    assert sha(PARENT/'receipt.json')==preserve['receipt_sha256']
    for name,digest in parent['attachments'].items():
        assert sha(PRIVATE/'parent-inputs'/name if name in saved else ROOT/name)==digest,name
    class ArchivedRoot:
        def __truediv__(self,name):
            return PRIVATE/'parent-inputs'/name if str(name) in saved else ROOT/name
    prior=runpy.run_path(str(PARENT/'verify_phases.py'))['validate']
    prior.__globals__['ROOT']=ArchivedRoot()
    assert prior()=={k:v for k,v in parent.items() if k!='attachments'}
    code=[n for n in saved if n.startswith(('cmd/','internal/','tools/'))]
    assert len(code)==314
    for name in code: assert sha(ROOT/name)==saved[name],name
    production=ROOT/'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(production)=='dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    r=read(PRIVATE/'window-replay.json');assert r['status']=='PASS'
    summaries=[]
    for number,keys,writes,offsets,casts,results,seed in [
        (2,1895,76220,242,2,0,2068430026),
        (3,1959,62742,237,14,7,2840416083)]:
        label=f'source{number}';p=PRIVATE/label
        d=read(p/'complete-source-private.json');w=read(p/'world-return-private.json')
        assert d['normal_startup'] and not any(d[x] for x in ('direct_entry','state_injection','wait_stub'))
        assert d['seed']==d['initial']['seed']==1 and d['rounds']==65 and len(d['inputs'])==keys
        assert len(d['boundaries'])==130 and len(d['post_battle'])==8
        assert len(d['indexed_window_writes'])==writes
        assert len({x['off'] for x in d['indexed_window_writes']})==offsets
        costs=[x['return_index'] for x in d['cost_flow'] if x['ida']=='0x20d1b']
        assert costs==[720]*casts
        spell_messages=[x for x in d['messages'] if x['record'] in [361,362,363]]
        assert len(spell_messages)==results and all(x['record']==361 for x in spell_messages)
        assert any(x['record']==378 for x in d['messages'])
        if number==3:
            for turn in (2,4):
                b=next(x for x in d['boundaries'] if x['round']==turn and '0x1FD21' in x['at'])
                assert bytes.fromhex(b['state']['actors'][2])[35:37]==b'\x3e\x00'
        computed=replay(d,w);assert computed==r['results'][number-2]['checkpoints']
        assert computed[0]['index']==16 and computed[0]['cost']==100 and computed[0]['seed']==seed
        last=w['events'][-8:]
        assert [x['state']['time'] for x in last]==[3360]*6+[3260]*2
        assert all((x['state']['raw_bb84'],x['state']['raw_12b8_indexed'],x['state']['movement'])==(16,30,20) for x in last)
        assert all((x['state']['sub_x'],x['state']['sub_y'])==(4,4) for x in last)
        for filename,digest in r['results'][number-2]['inputs'].items(): assert sha(p/filename)==digest
        driver=read(p/'driver.json');assert driver['parent_sha256']==sha(ROOT/driver['parent_source'])
        assert driver['driver_sha256']==sha(p/'instrumented-indexed-read-private.go')
        log=(p/'run.log').read_text();assert '--- PASS: TestNormalBookEntry' in log and 'FAIL' not in log
        # 壞old值與壞快照必須遭拒，不只驗正向資料。
        bad=dict(d);bad['indexed_window_writes']=[dict(x) for x in d['indexed_window_writes']]
        bad['indexed_window_writes'][0]['old']^=1
        try: replay(bad,w)
        except AssertionError: pass
        else: raise AssertionError('錯誤old值被接受')
        badw=json.loads(json.dumps(w));target=next(x for x in badw['events'] if x['ida']=='0x29800' and x['state']['steps']>=d['indexed_window_start_step'])
        target['state']['indexed_window_hex']='ff'+target['state']['indexed_window_hex'][2:]
        try: replay(d,badw)
        except AssertionError: pass
        else: raise AssertionError('錯誤快照被接受')
        summaries.append(dict(source=label,rounds=65,keys=keys,changed_byte_writes=writes,
            unique_offsets=offsets,cast_cost_observations=casts,cast_result_messages=results,
            world_cost=100,world_times=[3360,3260],window_replay='passed',negative_controls='passed'))
    failure=(PRIVATE/'source/run.log').read_text()
    assert '畫面完全沒變' in failure and 'FAIL' in failure
    ida=read(PRIVATE/'indexed-window-ida-private.json')
    assert ida['schema']=='kol-return-window-ida-v1' and ida['tool']=='IDA Pro 9.4' and ida['function_count']==338
    assert ida['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert len(ida['indexed_direct_refs'])==3
    rows={x['original_ea']:x for r in ida['records'] for x in r['instructions']}
    assert '0A000h' in rows['0x1d509']['operand_text'] and '1AAAh' in rows['0x1d532']['operand_text']
    assert rows['0x1d541']['bytes']=='a4'
    assert {x['reference']['original_ea'] for x in ida['indexed_direct_refs']}=={'0x2d52c','0x2d544','0x2d553'}
    return dict(schema='kol-return-indexed-window-receipt-v1',historical_shared_inputs=1348,
        parent_attachments=len(parent['attachments']),parent_results_recomputed=True,
        unchanged_code_inputs=314,production_sha256=sha(production),sources=summaries,
        normal_out_of_range_return_observed=False,production_implemented=False,
        status='正常窗變值重播通過；正式返回成本與保存垂直鏈仍待驗。')

def attachments():
    names=set(read(PRIVATE/'parent-preservation.json')['sources'])
    names|={str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names|={'docs/re/20261002-indexed-return-window.md',
            'docs/verification/return-indexed-read-20261002/verify_window.py'}
    return {n:sha(ROOT/n) for n in sorted(names)}

if __name__=='__main__':
    ap=argparse.ArgumentParser();ap.add_argument('--seal',action='store_true');ap.add_argument('--audit',action='store_true');args=ap.parse_args()
    assert not(args.seal and args.audit);result=validate()
    if args.seal:
        result['attachments']=attachments();p=Path('/out/receipt.json');assert not p.exists()
        p.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt=read(HERE/'receipt.json')
        for name,digest in receipt['attachments'].items():assert sha(ROOT/name)==digest,name
        assert result=={k:v for k,v in receipt.items() if k!='attachments'}
        if args.audit:
            control=read(PRIVATE/'host-control.json')
            assert control['git_diff_check']=='passed' and control['git_staged']==control['project_containers']==[]
            bad=[];md=[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    q=Path(directory)/name
                    if q.lstat().st_uid==0:bad.append(str(q))
                    if q.is_dir() and q.name.endswith('.md'):md.append(str(q))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat();assert (q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-return-indexed-window-audit-v1',
                receipt_sha256=sha(HERE/'receipt.json'),attachments=len(receipt['attachments']),
                full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000],
                production_unchanged=True),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
