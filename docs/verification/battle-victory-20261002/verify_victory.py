"""RE-186 正常勝利來源的封存與全唯讀重驗。"""
from pathlib import Path
import argparse
import hashlib
import json
import runpy

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/battle-victory-20261002'
HERE = ROOT / 'docs/verification/battle-victory-20261002'


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate():
    history = runpy.run_path(str(PRIVATE / 'history_check.py'))['validate']()
    baseline = read(ROOT / 'workplace/battle-exits-20261002/flee4/complete-source-private.json')['initial']
    attempts = []
    for n, rounds, keys, done in [(1, 63, 2000, False), (2, 1, 14, False),
                                   (3, 81, 1738, False), (4, 4, 147, True)]:
        path = PRIVATE / f'source{n}'
        data = read(path / 'complete-source-private.json')
        assert data['schema'] == 'kol-normal-victory-attempt-v1'
        assert data['normal_startup'] and data['seed'] == 1
        assert data['revision'] == '968ca774705853276ae55c8d1163c369481ac71f'
        assert data['input_sha256'] == '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
        assert not any(data[k] for k in ['direct_entry', 'state_injection', 'wait_stub'])
        assert data['initial'] == baseline
        assert (data['rounds'], len(data['inputs']), data['reached_battle_exit']) == (rounds, keys, done)
        log = (path / 'run.log').read_text()
        assert ('--- PASS:' in log) == done
        assert not any(r['key'] == 'Y' for r in data['inputs'])
        attempts.append(dict(source=f'source{n}', round_counter=rounds, keys=keys,
                             reached_exit=done, initial_identical=True))
    failed = read(PRIVATE / 'source3/complete-source-private.json')
    first, after = failed['boundaries'][:2]
    assert bytes.fromhex(first['state']['actors'][0])[31:33] == bytes([33, 30])
    assert bytes.fromhex(after['state']['actors'][0])[31:33] == bytes([34, 29])
    data = read(PRIVATE / 'source4/complete-source-private.json')
    first, after = data['boundaries'][:2]
    assert bytes.fromhex(after['state']['actors'][0])[31:33] == bytes([32, 31])
    # Theo真正先換副手武器；用完整16bytes驗證，不以指令名稱推定。
    initial = bytes.fromhex(data['initial']['characters'][1])
    switched = bytes.fromhex(after['state']['characters'][1])
    assert initial[0x30:0x40] == bytes(16)
    assert switched[0x30:0x40] == initial[0x190:0x1A0]
    assert switched[0x190:0x1A0] == initial[0x30:0x40]
    exit_row = next(r for r in data['exit_control'] if r['ida'] == '0x211AD')
    assert exit_row['down_enemy_6a96'] == 1 and exit_row['down_party_670a'] == 0
    assert exit_row['raw_bb84'] == 1
    actors = [bytes.fromhex(r) for r in data['battle_exit_state']['actors']]
    assert [actors[i][17] for i in [0, 1]] == [15, 16]
    assert actors[6][19] == 255
    assert [r['n'] for r in data['rolls'] if r['phase'] == 'post'] == [19] * 12 + [100, 128]
    assert data['battle_exit_counts']['rolls'] == 86
    for address, schema in [('controls', 'kol-victory-controls-ida-v1'),
                             ('turn-order', 'kol-victory-turn-order-ida-v1')]:
        ida = read(PRIVATE / f'{address}-ida-private.json')
        assert ida['schema'] == schema and ida['tool'] == 'IDA Pro 9.4'
        assert ida['function_count'] == 338
        assert ida['address_space'] == 'IDA linear'
        assert ida['sha256'] == data['input_sha256']
        assert ida['database_sha256'] == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    world = read(PRIVATE / 'source4/world-return-private.json')
    assert world['normal_startup'] and not any(world[k] for k in ['direct_entry', 'state_injection', 'wait_stub'])
    events = world['events']
    start = next(i for i, r in enumerate(events) if r['ida'] == '0x1E3A8')
    fields = ['x', 'y', 'sub_x', 'sub_y', 'time']
    assert [events[start]['state'][k] for k in fields] == [49, 23, 2, 3, 3360]
    chain, offset = [], start + 1
    for address in ['0x2A94F', '0x2A95E', '0x29800', '0x2981F', '0x29823', '0x29981']:
        index = next(i for i in range(offset, len(events)) if events[i]['ida'] == address)
        chain.append(events[index]['state'])
        offset = index + 1
    assert all(r['raw_bb84'] == 1 and r['raw_12b8_indexed'] == 10 for r in chain)
    assert [chain[1][k] for k in fields] == [49, 23, 4, 4, 3360]
    assert [chain[-1][k] for k in fields] == [49, 23, 4, 4, 3300]
    assert 2 * (chain[2]['movement'] + chain[2]['raw_12b8_indexed']) == 60
    assert chain[3]['time'] - chain[4]['time'] == 60
    assert chain[1]['seed'] == 1502690714 and chain[2]['seed'] == 1084083701
    assert world['rolls'][-1]['caller_ida'] == '0x29A92' and world['rolls'][-1]['n'] == 128
    assert world['rolls'][-1]['state']['time'] == 3360
    assert sha(ROOT / 'cmd/kol/main.go') == sha(PRIVATE / 'parent-inputs/cmd/kol/main.go')
    parent = read(PRIVATE / 'parent-preservation.json')
    for name, digest in parent['sources'].items():
        if name.startswith(('cmd/', 'internal/', 'tools/')):
            assert sha(ROOT / name) == digest, name
    return dict(schema='kol-victory-source-receipt-v1', attempts=attempts,
        native_victory=dict(rounds=4, battle_keys=147, battle_random_calls=86,
                            exit_scratch=1, returned_scratch=1, indexed_byte=10,
                            movement=20, elapsed=60, world=[49, 23, 4, 4, 3300],
                            exit_seed=775978695, before_weather_seed=1502690714,
                            returned_seed=1084083701, post_random_calls=14),
        historical=history, production_unchanged=True,
        limitation='原版正常來源；未驗重製勝利GUI、完整亂數接續或其他勝利場次，不升正式時間規格。')


def attachments():
    paths = [p for p in PRIVATE.rglob('*') if p.is_file()]
    paths += [ROOT / 'CONTEXT.md', ROOT / 'WORKLIST.md', ROOT / 'docs/worklist.json', ROOT / 'docs/re/RESEARCH-LOG.md',
              ROOT / 'docs/re/20261002-battle-victory.md',
              ROOT / 'docs/spec/011-wilderness-ambush.md', ROOT / 'docs/spec/027-battle-exit-wounds.md',
              HERE / 'verify_victory.py']
    return {str(p.relative_to(ROOT)): sha(p) for p in sorted(set(paths))}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--seal', action='store_true')
    args = parser.parse_args()
    result = validate()
    if args.seal:
        target = Path('/out/receipt.json')
        assert not target.exists()
        result['attachments'] = attachments()
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    else:
        saved = read(HERE / 'receipt.json')
        for name, digest in saved['attachments'].items():
            assert sha(ROOT / name) == digest, name
        assert result == {k: v for k, v in saved.items() if k != 'attachments'}
    print(json.dumps(result if not args.seal else {k: v for k, v in result.items() if k != 'attachments'}, ensure_ascii=False))
