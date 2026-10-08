"""RE-189施法暫存證據與歷史來源的封存／唯讀重驗。"""
import argparse
import hashlib
import json
import os
import runpy
from pathlib import Path

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/return-cost-gate-20261002'
HERE = ROOT / 'docs/verification/return-cost-gate-20261002'
PARENT = ROOT / 'docs/verification/return-cost-20261002'


def read(p):
    return json.loads(p.read_text())


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def validate():
    preservation = read(PRIVATE / 'parent-preservation.json')
    assert preservation['schema'] == 'kol-return-cost-gate-parent-v1'
    assert len(preservation['sources']) == 1332
    saved_names = preservation['sources']
    for name, digest in saved_names.items():
        p = PRIVATE / 'parent-inputs' / name
        assert sha(p) == digest, name
        assert (p.stat().st_uid, p.stat().st_gid) == (1000, 1000), name
    parent = read(PARENT / 'receipt2.json')
    assert sha(PARENT / 'receipt2.json') == preservation['receipt_sha256']
    for name, digest in parent['attachments'].items():
        p = PRIVATE / 'parent-inputs' / name if name in saved_names else ROOT / name
        assert sha(p) == digest, name

    class ArchivedRoot:
        def __truediv__(self, name):
            return PRIVATE / 'parent-inputs' / name if str(name) in saved_names else ROOT / name

    old = runpy.run_path(str(PARENT / 'verify_cost.py'))['validate']
    old.__globals__['ROOT'] = ArchivedRoot()
    assert old() == {k: v for k, v in parent.items() if k != 'attachments'}
    unchanged = 0
    for name, digest in saved_names.items():
        if name.startswith(('cmd/', 'internal/', 'tools/')):
            assert sha(ROOT / name) == digest, name
            unchanged += 1
    assert unchanged == 314
    production = ROOT / 'workplace/world-clock-20261002/kol-world-time-linux'
    assert sha(production) == 'dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    ida = read(ROOT / 'workplace/return-cost-20261002/cost-flow-ida-private.json')
    ins = {i['original_ea']: i for f in ida['records'] for i in f['instructions']}
    for ea, code in {'0x20c31':'c70684bb0800', '0x20c73':'a384bb',
                     '0x20cb8':'890e84bb', '0x20d17':'010684bb'}.items():
        assert ins[ea]['bytes'] == code
    direction = read(PRIVATE / 'spell-alias-cases-private.json')
    assert direction['schema'] == 'kol-world-cost-spell-alias-v1'
    expected = [-1, 0, 1, 16, 840]
    assert [(r['Index'], r['Seed']) for r in direction['rows']] == [(i,s) for i in expected for s in [1,2]]
    for r in direction['rows']:
        byte = {-1:0, 0:10, 1:10, 16:30, 840:r['Seed']}[r['Index']]
        assert r['ReadOffset'] == 0x12B8+r['Index'] and r['Byte'] == byte
        assert r['Before'] == 4200 and r['After'] == 4200-2*(20+byte)
        assert r['Cost'] == 2*(20+byte) and r['Calls'] == 0
        assert r['FinalSeed'] == r['Seed']
    spells = read(PRIVATE / 'spell-cast-cost-private.json')
    assert spells['schema'] == 'kol-spell-scratch-world-cost-v1'
    assert [(r['Index'], r['Seed']) for r in spells['rows']] == [(i,s) for i in [840,150,1400,586] for s in [1,2]]
    configurations = [([32,0,0,0],4,0x72),([0,0,0,0],0,0),
                      ([32,0,0,0],4,0xF2),([49,32,0,0],7,0x16)]
    for n, r in enumerate(spells['rows']):
        guards, effect, power = configurations[n//2]
        assert r['Spell']['Guards'] == guards
        assert (r['Spell']['Effect'],r['Spell']['PowerCode']) == (effect,power)
        byte = {840:r['Seed'],150:1,1400:100,586:0}[r['Index']]
        assert r['Byte'] == byte and r['ReadOffset'] == 0x12B8+r['Index']
        assert r['Before'] == 4200 and r['After'] == 4200-2*(20+byte)
        assert r['Cost'] == 2*(20+byte) and r['Calls'] == 0
        assert r['FinalSeed'] == r['Seed']
    for name, test in [('alias-cases2.log','TestWorldCostSpellAlias'),
                       ('spell-cast-cost.log','TestSpellScratchReachesWorldCost')]:
        log = (PRIVATE / name).read_text()
        assert f'--- PASS: {test}' in log and 'FAIL' not in log
    failure = (PRIVATE / 'alias-cases.log').read_text()
    assert '--- FAIL: TestWorldCostSpellAlias' in failure and '0x297f2' in failure
    return dict(schema='kol-return-cost-gate-receipt-v1', native_cost_cases=10,
                native_spell_cost_cases=8, historical_shared_inputs=1332,
                parent_attachments=len(parent['attachments']), parent_results_recomputed=True,
                unchanged_code_inputs=unchanged, production_sha256=sha(production),
                rng_alias_costs=[42,44], normal_spell_return_verified=False,
                status='成本DRAFT；原版指令切片已證實，正常施法返回仍待驗。')


def attachments():
    names = set(read(PRIVATE / 'parent-preservation.json')['sources'])
    names |= {str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names |= {'docs/re/20261002-return-cost-spells.md',
              'docs/verification/return-cost-gate-20261002/verify_gate.py'}
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
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    else:
        receipt = read(HERE / 'receipt.json')
        for name, digest in receipt['attachments'].items():
            assert sha(ROOT / name) == digest, name
        assert result == {k:v for k,v in receipt.items() if k != 'attachments'}
        if args.audit:
            control = read(PRIVATE / 'host-control.json')
            assert control['git_head'] == '1384e508f70e8a5d81bc050ea79f66d3f0d38fe6'
            assert control['git_diff_check'] == 'passed'
            assert control['git_staged'] == control['project_containers'] == []
            bad_owners, md_dirs = [], []
            for parent, dirs, files in os.walk(ROOT):
                for name in dirs + files:
                    p = Path(parent) / name
                    if p.lstat().st_uid == 0: bad_owners.append(str(p))
                    if p.is_dir() and p.name.endswith('.md'): md_dirs.append(str(p))
            assert not bad_owners and not md_dirs
            for name in receipt['attachments']:
                p = ROOT / name
                assert (p.stat().st_uid, p.stat().st_gid) == (1000,1000), name
            audit = dict(schema='kol-return-cost-gate-audit-v1',
                         receipt_sha256=sha(HERE / 'receipt.json'),
                         attachments=len(receipt['attachments']),
                         full_readonly_verification='passed', host_control=control,
                         root_owned_files=bad_owners, md_directories=md_dirs,
                         artifact_uid_gid=[1000,1000], production_unchanged=True)
            target = Path('/out/audit.json')
            assert not target.exists()
            target.write_text(json.dumps(audit,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k != 'attachments'}, ensure_ascii=False))
