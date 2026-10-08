"""戰鬥返回來源的唯讀重驗；歷史輸入使用封存副本，收據只建立一次。"""
from pathlib import Path
from collections import Counter
import argparse
import hashlib
import json
import runpy

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/battle-return-20261002'
HERE = ROOT / 'docs/verification/battle-return-20261002'
LEGACY = ROOT / 'docs/verification/battle-world-20261002'
INPUT_SHA = '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
REVISION = '968ca774705853276ae55c8d1163c369481ac71f'


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def history():
    parent = read(PRIVATE / 'parent-preservation.json')
    assert parent['schema'] == 'kol-world-return-parent-v1'
    assert len(parent['sources']) == 1303
    for name, digest in parent['sources'].items():
        p = PRIVATE / 'parent-inputs' / name
        assert sha(p) == digest, name
        assert p.stat().st_uid == p.stat().st_gid == 1000, name
    prior = read(LEGACY / 'receipt.json')
    assert sha(LEGACY / 'receipt.json') == parent['receipt_sha256']
    assert sha(LEGACY / 'audit.json') == parent['audit_sha256']
    for name, digest in prior['attachments'].items():
        path = PRIVATE / 'parent-inputs' / name if name in parent['sources'] else ROOT / name
        assert sha(path) == digest, name
    calculated = runpy.run_path(str(LEGACY / 'verify_world.py'))['validate']()
    assert calculated == {k: v for k, v in prior.items() if k != 'attachments'}
    # 本輪只改文件與觀察工具，不能悄悄替換產品輸入。
    for name, digest in parent['sources'].items():
        if name.startswith(('cmd/', 'internal/', 'assets/', 'docs/i18n/')) or name in ['go.mod', 'go.sum']:
            assert sha(ROOT / name) == digest, name
    return dict(preserved_shared_inputs=1303, legacy_attachments=len(prior['attachments']),
                legacy_results_recomputed=True, production_inputs_unchanged=True)


def validate():
    historical = history()
    source = read(PRIVATE / 'source1/world-return-private.json')
    assert source['schema'] == 'kol-normal-world-return-v1'
    assert source['normal_startup']
    assert not any(source[k] for k in ['direct_entry', 'state_injection', 'wait_stub'])
    assert source['input_sha256'] == INPUT_SHA and source['revision'] == REVISION
    assert len(source['inputs']) == 162 and len(source['events']) == 549
    complete = read(PRIVATE / 'source1/complete-source-private.json')
    old = read(ROOT / 'workplace/battle-loot-20261002/loot1/complete-source-private.json')
    fields = ['initial', 'inputs', 'boundaries', 'rolls', 'messages', 'post_battle',
              'battle_exit_state', 'battle_exit_counts']
    for key in fields:
        assert complete[key] == old[key], key
    assert complete['rounds'] == 61 and len(complete['inputs']) == 351
    assert len(complete['rolls']) == 1491 and complete['battle_exit_counts']['rolls'] == 1458
    for name in ['battle-state-records.json', 'battle-planning.json', 'battle-alignment.json', 'battle-terrain.json']:
        assert sha(PRIVATE / 'source1' / name) == sha(ROOT / 'workplace/battle-loot-20261002/loot1' / name)
    events = source['events']
    entry_index = next(i for i, e in enumerate(events) if e['ida'] == '0x1E3A8')
    entry = events[entry_index]['state']
    assert [entry[k] for k in ['x', 'y', 'sub_x', 'sub_y', 'time', 'movement', 'raw_bb84']] == [49, 23, 2, 3, 3360, 20, 65535]
    after = events[entry_index + 1:]
    chain = ['0x2A94F', '0x2A952', '0x2A958', '0x2A95E', '0x29A43',
             '0x29A83', '0x29800', '0x2981F', '0x29823', '0x29981']
    picked = []
    offset = 0
    for address in chain:
        n = next(i for i in range(offset, len(after)) if after[i]['ida'] == address)
        picked.append(after[n])
        offset = n + 1
    for event in picked:
        s = event['state']
        assert [s[k] for k in ['x', 'y', 'raw_bb84', 'raw_12b8_indexed', 'movement']] == [49, 23, 16, 30, 20]
        assert s['time'] == (3260 if event['ida'] in ['0x29823', '0x29981'] else 3360)
    assert [picked[0]['state'][k] for k in ['sub_x', 'sub_y']] == [4, 3]
    assert [picked[3]['state'][k] for k in ['sub_x', 'sub_y']] == [4, 4]
    assert [picked[-1]['state'][k] for k in ['sub_x', 'sub_y']] == [4, 4]
    assert picked[5]['state']['seed'] == 4284445587
    assert picked[6]['state']['seed'] == 3077853194
    cost = 2 * (picked[6]['state']['movement'] + picked[6]['state']['raw_12b8_indexed'])
    assert entry['time'] - picked[-1]['state']['time'] == cost == 100
    # 進場方向40與提前扣時都不能通過相同返回條件。
    assert 2 * (entry['movement'] + entry['raw_12b8_indexed']) == 40 != cost
    assert entry['time'] - 40 != picked[-1]['state']['time']
    old_gui = {}
    for language in ['en', 'zh-Hant']:
        old_rows = read(ROOT / 'workplace/battle-world-20261002' / ('gui-complete3-' + language) / 'trace.json')['steps']
        previous = old_rows[480]['world']
        actual = [previous[k] for k in ['x', 'y', 'sub_x', 'sub_y', 'ticks']]
        assert actual == [49, 23, 2, 3, 3320]
        assert actual != [49, 23, 4, 4, 3260]
        old_gui[language] = actual
    cases = read(PRIVATE / 'loot-return-scratch-cases-private.json')
    assert cases['schema'] == 'kol-loot-return-scratch-source-cases-v1'
    assert cases['sha256'] == INPUT_SHA and cases['no_stubs'] and not cases['normal_player_path']
    assert len(cases['cases']) == 432
    counts = Counter()
    for row in cases['cases']:
        assert row['Seed'] == 1232439027
        initial = row['InitialScratch']
        assert initial in [-1, 0, 1]
        assert row['FinalScratch'] == row['WantScratch']
        assert row['FinalScratch'] in [initial, 16]
        assert len(row['Before']) == len(row['After']) == 6
        assert all(len(bytes.fromhex(v)) == 768 for v in row['Before'] + row['After'])
        counts[(initial, row['FinalScratch'])] += 1
    for initial in [-1, 0, 1]:
        assert counts[(initial, initial)] == 136 and counts[(initial, 16)] == 8
    ida = read(PRIVATE / 'world-return-scratch-ida-private.json')
    assert ida['schema'] == 'kol-world-return-scratch-ida-v1'
    assert ida['sha256'] == INPUT_SHA and ida['tool'] == 'IDA Pro 9.4'
    assert ida['address_space'] == 'IDA linear' and ida['function_count'] == 338
    assert len(ida['threshold_direct_refs']) == 120
    expected_bytes = {'0x29803': '8b1e84bb', '0x2980a': '8a87b812', '0x2981f': '29067000'}
    rows = {r['original_ea']: r for block in ida['records'] for r in block['instructions']}
    for address, raw in expected_bytes.items():
        assert rows[address]['bytes'] == raw, address
    reviewed = read(PRIVATE / 'world-return-reviewed-ida-private.json')
    index = read(ROOT / 'docs/re/world-return-address-index.json')
    merged = {}
    all_rows = [r for b in reviewed['records'] for r in b['instructions']]
    all_rows += [r['reference'] for r in reviewed['threshold_direct_refs']]
    all_rows += [r for ref in reviewed['threshold_direct_refs'] for r in ref['context']]
    for row in all_rows:
        for annotation in row['annotations']:
            if annotation['index'].endswith('/world-return-address-index.json'):
                assert annotation['index_sha256'] == sha(ROOT / 'docs/re/world-return-address-index.json')
                assert annotation['confidence'] == '已證實' and not annotation['warning']
                merged[row['original_ea']] = annotation
    assert set(merged) == set(index['entries'])
    product = ROOT / 'workplace/battle-loot-20261002/fix2/kol-battle-loot-linux'
    assert sha(product) == '7bc7d34a76270fe355c268393ac68308cead4d9c7f59ed960c6681cc311495d4'
    assert product.stat().st_mode & 0o777 == 0o755
    return dict(schema='kol-world-return-source-verification-v1', date='2026-10-02',
                input_sha256=INPUT_SHA, dosgolem_revision=REVISION, historical=historical,
                normal_source=dict(inputs_observed=162, events=549, unchanged_battle_fields=fields,
                    rounds=61, battle_keys=351, battle_rolls=1458, total_observed_rolls=1491,
                    entry=[49, 23, 2, 3, 3360],
                    returned=[49, 23, 4, 4, 3260], scratch=16, indexed_byte=30, elapsed=100,
                    weather_seed_before=4284445587, weather_seed_after=3077853194),
                scratch_cases=dict(cases=432, seed=1232439027, initial_scratch=[-1, 0, 1],
                    preserved_per_initial=136, cleared_per_initial=8, no_stubs=True, normal_player_path=False),
                reviewed_address_entries=len(merged), direct_scratch_refs=120, production_unchanged=True,
                previous_product_exit_rejected=old_gui,
                limitations=['只證實本場正常原版返回，沒有新產品GUI／保存驗收。',
                    '432組只證實搜刮常式，不能推定整場其他出口保留進場方向。',
                    '勝利／逃走及世界／戰鬥亂數接續尚未驗，返回子切片DRAFT、父閘門開放。'])


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--export', action='store_true')
    args = parser.parse_args()
    result = validate()
    if args.export:
        target = Path('/out/receipt.json')
        assert not target.exists()
        paths = [p for p in PRIVATE.rglob('*') if p.is_file()]
        paths += [HERE / 'verify_return.py', HERE / 'audit_return.py', ROOT / 'docs/re/20261002-battle-return.md',
                  ROOT / 'docs/re/world-return-address-index.json', ROOT / 'docs/re/RESEARCH-LOG.md',
                  ROOT / 'docs/spec/011-wilderness-ambush.md', ROOT / 'docs/spec/027-battle-exit-wounds.md',
                  ROOT / 'CONTEXT.md', ROOT / 'WORKLIST.md', ROOT / 'docs/worklist.json']
        assert all(p.stat().st_uid == p.stat().st_gid == 1000 for p in paths)
        result['attachments'] = {str(p.relative_to(ROOT)): sha(p) for p in sorted(set(paths))}
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    else:
        expected = read(HERE / 'receipt.json')
        for name, digest in expected['attachments'].items():
            p = ROOT / name
            assert sha(p) == digest, name
            assert p.stat().st_uid == p.stat().st_gid == 1000, name
        assert result == {k: v for k, v in expected.items() if k != 'attachments'}
    print(json.dumps({k: v for k, v in result.items() if k != 'attachments'}, ensure_ascii=False))


if __name__ == '__main__':
    main()
