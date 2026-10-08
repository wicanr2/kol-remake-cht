"""RE-190正常原版施法返回的封存與唯讀重驗。"""
import argparse
import hashlib
import json
import os
import runpy
from pathlib import Path

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/spell-return-20261002'
HERE = ROOT / 'docs/verification/spell-return-20261002'
PARENT = ROOT / 'docs/verification/return-cost-gate-20261002'


def read(p):
    return json.loads(p.read_text())


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def validate():
    preservation = read(PRIVATE / 'parent-preservation.json')
    saved = preservation['sources']
    assert preservation['schema'] == 'kol-spell-return-parent-v1'
    assert len(saved) == 1336
    for name, digest in saved.items():
        assert sha(PRIVATE / 'parent-inputs' / name) == digest, name
    parent = read(PARENT / 'receipt.json')
    assert sha(PARENT / 'receipt.json') == preservation['receipt_sha256']
    for name, digest in parent['attachments'].items():
        p = PRIVATE / 'parent-inputs' / name if name in saved else ROOT / name
        assert sha(p) == digest, name

    class ArchivedRoot:
        def __truediv__(self, name):
            return PRIVATE / 'parent-inputs' / name if str(name) in saved else ROOT / name

    old = runpy.run_path(str(PARENT / 'verify_gate.py'))['validate']
    old.__globals__['ROOT'] = ArchivedRoot()
    assert old() == {k:v for k,v in parent.items() if k != 'attachments'}
    code_names = [name for name in saved if name.startswith(('cmd/', 'internal/', 'tools/'))]
    assert len(code_names) == 314
    for name in code_names:
        assert sha(ROOT / name) == saved[name], name
    production = ROOT / 'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(production) == 'dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    for n in [1,2]:
        log = (PRIVATE / f'source{n}/run.log').read_text()
        assert '--- FAIL: TestNormalBookEntry' in log
        assert '正常招募未選到Warder' in log
        assert not (PRIVATE / f'source{n}/complete-source-private.json').exists()
    source = PRIVATE / 'source3'
    log = (source / 'run.log').read_text()
    assert '--- PASS: TestNormalBookEntry' in log and 'FAIL' not in log
    assert '正常旅店已招募原始Warder' in log
    driver = read(source / 'driver.json')
    assert driver['generated_sha256'] == sha(source / 'instrumented-spell-private.go')
    assert driver['parent_sha256'] == sha(ROOT / driver['parent_path'])
    assert driver['preparation_sha256'] == sha(PRIVATE / 'prepare_spell3.py')
    s = read(source / 'complete-source-private.json')
    w = read(source / 'world-return-private.json')
    assert s['schema'] == 'kol-normal-spell-return-v1'
    for payload in [s,w]:
        assert payload['normal_startup']
        assert not any(payload[k] for k in ['direct_entry','state_injection','wait_stub'])
        assert payload['revision'] == '968ca774705853276ae55c8d1163c369481ac71f'
        assert payload['input_sha256'] == '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert s['seed'] == s['initial']['seed'] == 1
    assert s['reached_battle_exit'] and s['rounds'] == 42
    initial = [bytes.fromhex(h) for h in s['initial']['characters']]
    assert initial[0].split(b'\0')[0] == b'Olear'
    assert initial[1].split(b'\0')[0] == b'Warder'
    raw_spell = bytes.fromhex('bb0654b9599935207400000005008a53')
    assert initial[1][0x200:0x210] == raw_spell
    assert initial[1][0x30:0x40] == bytes.fromhex('72b0c013ee000000225828ff190be200')
    casts = [r for r in s['cost_flow'] if r['ida'] == '0x20d1b']
    assert [(r['round'],r['actor'],r['return_index']) for r in casts] == [(2,1,720),(3,1,720)]
    for r in casts:
        ch = bytes.fromhex(r['state']['characters'][1])
        assert ch[0x200:0x210] == raw_spell and ch[0x30] == 0
    cast_messages = [m for m in s['messages'] if 0x169 <= m['record'] <= 0x172]
    assert [(m['round'],m['record']) for m in cast_messages] == [(2,0x169),(3,0x169)]
    assert [r['action'] for r in s['normal_plans'] if r['actor']==1][:3] == ['DROP','CAST','CAST']
    actors = [bytes.fromhex(h) for h in s['battle_exit_state']['actors']]
    assert all(actors[i][0x11] == 0xff for i in [0,1])
    assert actors[6][0x11] != 0xff and actors[6][0x13] != 0xff
    counts = s['battle_exit_counts']
    assert (len(s['inputs']),len(s['boundaries']),len(s['cost_flow'])) == (921,84,475)
    assert (counts['rolls'],len(s['rolls']),len(s['post_battle'])) == (988,1021,8)
    assert len(s['return_writes']) == 561
    assert s['cost_flow'][-1]['ida'] == '0x21821' and s['cost_flow'][-1]['return_index'] == 16
    last = w['events'][-8:]
    assert [r['ida'] for r in last] == ['0x2A95E','0x29A43','0x29987','0x29A83',
                                      '0x29800','0x2981F','0x29823','0x29981']
    for r in last:
        q = r['state']
        assert (q['x'],q['y'],q['sub_x'],q['sub_y']) == (49,23,4,4)
        assert (q['raw_bb84'],q['raw_12b8_indexed'],q['movement']) == (16,30,20)
    assert last[5]['state']['time'] == 3360 and last[6]['state']['time'] == 3260
    assert last[3]['state']['seed'] == 397419821
    assert last[4]['state']['seed'] == 4068287292
    rules = (PRIVATE / 'cast-rules.log').read_text()
    assert '--- PASS: TestNormalWarderCastTime' in rules and 'FAIL' not in rules
    return dict(schema='kol-normal-spell-return-receipt-v1',historical_shared_inputs=1336,
                parent_attachments=len(parent['attachments']),parent_results_recomputed=True,
                unchanged_code_inputs=314,production_sha256=sha(production),
                native_normal_rounds=42,native_normal_inputs=921,cost_observations=475,
                native_spell_rounds=[2,3],native_spell_cost=720,spell_verdict='NothingThere',
                battle_rolls=988,total_rolls=1021,post_battle_waits=8,
                final_index=16,final_byte=30,world_cost=100,world_times=[3360,3260],
                normal_spell_return_verified=True,spell_cost_survived_to_world=False,
                status='正常丟弓、施法與敗北返回已證實；正式返回成本仍DRAFT。')


def attachments():
    names = set(read(PRIVATE / 'parent-preservation.json')['sources'])
    names |= {str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names |= {'docs/re/20261002-normal-spell-return.md',
              'docs/verification/spell-return-20261002/verify_spell_return.py'}
    return {name:sha(ROOT / name) for name in sorted(names)}


if __name__ == '__main__':
    ap = argparse.ArgumentParser()
    ap.add_argument('--seal', action='store_true')
    ap.add_argument('--audit', action='store_true')
    args = ap.parse_args()
    assert not (args.seal and args.audit)
    result = validate()
    if args.seal:
        result['attachments'] = attachments()
        target = Path('/out/receipt.json')
        assert not target.exists()
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        receipt = read(HERE / 'receipt.json')
        for name,digest in receipt['attachments'].items():
            assert sha(ROOT / name) == digest,name
        assert result == {k:v for k,v in receipt.items() if k != 'attachments'}
        if args.audit:
            control = read(PRIVATE / 'host-control.json')
            assert control['git_diff_check'] == 'passed'
            assert control['git_staged'] == control['project_containers'] == []
            bad,md = [],[]
            for directory,dirs,files in os.walk(ROOT):
                for name in dirs+files:
                    p = Path(directory)/name
                    if p.lstat().st_uid == 0: bad.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'): md.append(str(p))
            assert not bad and not md
            for name in receipt['attachments']:
                q=(ROOT/name).stat()
                assert (q.st_uid,q.st_gid)==(1000,1000),name
            print(json.dumps(dict(schema='kol-normal-spell-return-audit-v1',
                receipt_sha256=sha(HERE/'receipt.json'),attachments=len(receipt['attachments']),
                full_readonly_verification='passed',host_control=control,
                root_owned_files=bad,md_directories=md,artifact_uid_gid=[1000,1000],
                production_unchanged=True),ensure_ascii=False))
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))
