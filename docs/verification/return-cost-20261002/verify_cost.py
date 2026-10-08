"""RE-188正常來源與成本資料流的封存／唯讀重驗。"""
from pathlib import Path
import argparse
import hashlib
import json
import runpy

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/return-cost-20261002'
HERE = ROOT / 'docs/verification/return-cost-20261002'
LEGACY = ROOT / 'docs/verification/world-clock-20261002'


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate():
    fix = read(PRIVATE / 'audit-fix-preservation.json')
    first = read(HERE / 'receipt.json')
    assert sha(HERE / 'receipt.json') == fix['receipt_sha256']
    for name, digest in first['attachments'].items():
        path = PRIVATE / 'audit-fix-inputs' / name if name in fix['sources'] else ROOT / name
        assert sha(path) == digest, name
    parent = read(PRIVATE / 'parent-preservation.json')
    assert parent['schema'] == 'kol-return-cost-parent-v1' and len(parent['sources']) == 1326
    for name, digest in parent['sources'].items():
        saved = PRIVATE / 'parent-inputs' / name
        assert sha(saved) == digest, name
        assert (saved.stat().st_uid, saved.stat().st_gid) == (1000, 1000)
    receipt = read(LEGACY / 'receipt.json')
    assert sha(LEGACY / 'receipt.json') == parent['receipt_sha256']
    for name, digest in receipt['attachments'].items():
        path = PRIVATE / 'parent-inputs' / name if name in parent['sources'] else ROOT / name
        assert sha(path) == digest, name

    class ArchivedRoot:
        def __truediv__(self, name):
            return PRIVATE / 'parent-inputs' / name if str(name) in parent['sources'] else ROOT / name

    old = runpy.run_path(str(LEGACY / 'verify_time.py'))['validate']
    old.__globals__['ROOT'] = ArchivedRoot()
    result = old()
    assert result == {k: v for k, v in receipt.items() if k != 'attachments'}
    unchanged = 0
    for name, digest in parent['sources'].items():
        if name.startswith(('cmd/', 'internal/', 'tools/')):
            assert sha(ROOT / name) == digest, name
            unchanged += 1
    source = read(PRIVATE / 'victory/complete-source-private.json')
    observations = source.pop('cost_flow')
    assert source == read(ROOT / 'workplace/world-clock-20261002/victory/complete-source-private.json')
    assert (PRIVATE / 'victory/world-return-private.json').read_bytes() == (
        ROOT / 'workplace/world-clock-20261002/victory/world-return-private.json').read_bytes()
    assert len(observations) == 39
    costs = [r for r in observations if r['ida'] == '0x20bad']
    assert len(costs) == 11
    assert [(r['actor'], r['return_index']) for r in costs[-2:]] == [(0, 1), (1, 1)]
    assert costs[-2]['state']['steps'] < costs[-1]['state']['steps']
    case_data = read(PRIVATE / 'fatigue-cases-private.json')
    assert case_data['source_sha256'] == sha(PRIVATE / 'victory/complete-source-private.json')
    for r, case in zip(costs, case_data['cases']):
        raw = bytes.fromhex(r['state']['actors'][r['actor']])
        assert case == dict(round=r['round'], actor=r['actor'], command=raw[0x25],
                            style=raw[0x23], guard=raw[0x24], flag1e=raw[0x1E], want=r['return_index'])
    assert len(case_data['cases']) == 11
    log = (PRIVATE / 'fatigue-cases.log').read_text()
    assert '--- PASS: TestReturnCostNativeFatigue' in log and 'FAIL' not in log
    native_log = (PRIVATE / 'victory/run.log').read_text()
    assert '--- PASS: TestNormalBookEntry' in native_log and 'FAIL' not in native_log
    table = read(PRIVATE / 'cost-table-private.json')
    assert table['bytes'] == [0,10,10,10,20,6,14,30,8,31,28,31,30,31,30,31,31,30]
    assert table['dgroup_file_base'] == 0x202D0 and table['source_file_offset'] == 0x21587
    exe = ROOT / 'workplace/analysis/koltown/KOLTOWN-unpacked.EXE'
    assert sha(exe) == table['input_sha256']
    assert list(exe.read_bytes()[0x21587:0x21599]) == table['bytes']
    ida = read(PRIVATE / 'cost-flow-ida-private.json')
    assert ida['schema'] == 'kol-return-cost-ida-v1' and ida['function_count'] == 338
    assert ida['sha256'] == table['input_sha256'] and ida['address_space'] == 'IDA linear'
    db = ROOT / 'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64'
    assert sha(db) == ida['database_sha256']
    instructions = {r['original_ea']: r for section in ida['records'] for r in section['instructions']}
    assert instructions['0x20baa']['bytes'] == instructions['0x26a65']['bytes'] == 'a384bb'
    assert instructions['0x20b1f']['code_refs'] == ['0x20d8e']
    assert sum(len(r['instructions']) for r in ida['records']) == 648
    return dict(schema='kol-return-cost-receipt-v1', observations=39, native_fatigue_cases=11,
                complete_source_unchanged=True, last_same_value_writer=1, cost_bytes=18,
                historical_shared_inputs=1326, legacy_attachments=len(receipt['attachments']),
                legacy_results_recomputed=True, unchanged_code_inputs=unchanged,
                status='成本DRAFT；正式程式不改，更新順序保持RE-187有限CONFORMED。')


def attachments():
    names = set(read(PRIVATE / 'parent-preservation.json')['sources'])
    names |= {str(p.relative_to(ROOT)) for p in PRIVATE.rglob('*') if p.is_file()}
    names |= {'docs/re/20261002-return-cost.md', str((HERE / 'verify_cost.py').relative_to(ROOT))}
    names |= {str((HERE / 'audit_cost.py').relative_to(ROOT))}
    return {name: sha(ROOT / name) for name in sorted(names)}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--seal', action='store_true')
    parser.add_argument('--receipt', choices=['receipt.json', 'receipt2.json'], default='receipt2.json')
    args = parser.parse_args()
    result = validate()
    if args.seal:
        target = Path('/out') / args.receipt
        assert not target.exists()
        result['attachments'] = attachments()
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    else:
        receipt = read(HERE / args.receipt)
        fix = read(PRIVATE / 'audit-fix-preservation.json')
        for name, digest in receipt['attachments'].items():
            path = PRIVATE / 'audit-fix-inputs' / name if args.receipt == 'receipt.json' and name in fix['sources'] else ROOT / name
            assert sha(path) == digest, name
        assert result == {k:v for k,v in receipt.items() if k != 'attachments'}
    print(json.dumps({k:v for k,v in result.items() if k != 'attachments'}, ensure_ascii=False))
