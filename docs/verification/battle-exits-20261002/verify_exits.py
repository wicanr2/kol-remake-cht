"""RE-185有限次格修正與正常出口來源；封存後可全唯讀重驗。"""
from pathlib import Path
import argparse
import hashlib
import json
import runpy
import sys

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/battle-exits-20261002'
HERE = ROOT / 'docs/verification/battle-exits-20261002'
INPUT = '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
REVISION = '968ca774705853276ae55c8d1163c369481ac71f'
PRODUCT = '9c64e34362cee13c5abbea95174f46f7c78f90d1f33968947c9ded5704c4d05f'
PROBE = 'cd1f16f00f2d15b118ff21b5ca34aeb72dfc5e51fc71b79f4c014bbedda7fa1e'
sys.path.insert(0, str(ROOT / 'tools'))
from png_io import read_png


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def native(directory, expected, rounds, keys, scratch, elapsed):
    source = read(PRIVATE / directory / 'world-return-private.json')
    complete = read(PRIVATE / directory / 'complete-source-private.json')
    for data in [source, complete]:
        assert data['normal_startup'] and data['input_sha256'] == INPUT
        assert data['revision'] == REVISION
        assert not any(data[k] for k in ['direct_entry', 'state_injection', 'wait_stub'])
    assert complete['reached_battle_exit'] and complete['rounds'] == rounds
    assert len(complete['inputs']) == keys
    events = source['events']
    index = next(i for i, r in enumerate(events) if r['ida'] == '0x1E3A8')
    fields = ['x', 'y', 'sub_x', 'sub_y', 'time']
    assert [events[index]['state'][k] for k in fields] == [49, 23, 2, 3, 3360]
    tail = events[index + 1:]
    chain, offset = [], 0
    for address in ['0x2A94F', '0x2A95E', '0x29800', '0x2981F', '0x29823', '0x29981']:
        n = next(i for i in range(offset, len(tail)) if tail[i]['ida'] == address)
        chain.append(tail[n]['state'])
        offset = n + 1
    assert chain[0]['raw_bb84'] == scratch
    assert [chain[1][k] for k in fields] == [49, 23, 4, 4, 3360]
    assert [chain[-1][k] for k in fields] == expected
    assert chain[3]['time'] - chain[4]['time'] == elapsed
    assert 2 * (chain[2]['movement'] + chain[2]['raw_12b8_indexed']) == elapsed
    assert all(r['raw_bb84'] == scratch for r in chain)
    initial = complete['initial']
    baseline = read(ROOT / 'workplace/battle-return-20261002/source1/complete-source-private.json')['initial']
    # 本輪跳過舊版先行規劃展示。角色／計數／seed相同，原始指令欄與步數不冒充相同。
    for key in ['characters', 'counters', 'seed']:
        assert initial[key] == baseline[key], key
    assert initial == read(PRIVATE / 'flee4/complete-source-private.json')['initial']
    actors = [bytes.fromhex(a) for a in complete['battle_exit_state']['actors']]
    if directory == 'flee4':
        assert [r['key'] for r in complete['inputs']] == [',', 'Return', 'Return'] * 2 + ['.', '.', 'Return', 'Y']
        assert complete['inputs'][-1]['last_print'] == 265
        assert [actors[i][0x11] for i in [0, 1]] == [15, 16]
        assert all(r['down_party_670a'] == r['down_enemy_6a96'] == 255 for r in complete['exit_control'])
        assert all(r['raw_bb84'] == 0 for r in complete['exit_control'])
    else:
        assert not any(r['key'] == 'Y' for r in complete['inputs'])
        assert all(actors[i][0x11] == 255 for i in [0, 1])
    return dict(round_counter=rounds, battle_keys=keys, scratch=scratch,
        indexed_byte=chain[2]['raw_12b8_indexed'], elapsed=elapsed,
        returned=expected, normal_startup=True, waits_unmodified=True,
        initial_relation='兩條新來源完整初態相同；與RE-184角色／計數／seed相同，先行規劃指令與步數不同。')


def marker_difference(current, previous, scale):
    a, b = read_png(current), read_png(previous)
    assert (a.width, a.height) == (b.width, b.height) == (320 * scale, 200 * scale)
    aa, bb = a.rgb(), b.rgb()
    old_origin = (188 * scale, 112 * scale)
    new_origin = (192 * scale, 114 * scale)
    expected = {(x0 + x, y0 + y) for x0, y0 in [old_origin, new_origin]
                for x in range(2 * scale) for y in range(2 * scale)}
    # 比較完整RGB，只允許來源已證實的舊／新點位兩個矩形。
    points = {(i // 3 % a.width, i // 3 // a.width)
              for i in range(0, len(aa), 3) if aa[i:i + 3] != bb[i:i + 3]}
    assert points == expected, current.name
    return len(points)


def validate():
    historical = runpy.run_path(str(PRIVATE / 'history_check.py'))['validate']()
    product = PRIVATE / 'kol-world-return-linux'
    probe = PRIVATE / 'kol-settlement-probe2-linux'
    assert sha(product) == PRODUCT and sha(probe) == PROBE
    assert product.stat().st_mode & 0o777 == 0o755
    code = (ROOT / 'cmd/kol/main.go').read_text()
    manifest = read(PRIVATE / 'probe2-source.json')
    assert manifest['parent_sha256'] == sha(ROOT / 'cmd/kol/main.go')
    assert 'KOL_PROBE_PLAN_SEED' not in code
    assert sha(PRIVATE / 'probe2-main-private.go') == manifest['probe_sha256']
    log = (PRIVATE / 'center-rules.log').read_text()
    assert 'FAIL' not in log and 'cmd/kol' in log and 'internal/save' in log
    assert (PRIVATE / 'center-build.log').read_text() == ''
    ida = read(PRIVATE / 'battle-exits-ida-private.json')
    assert ida['schema'] == 'kol-battle-exits-ida-v1' and ida['sha256'] == INPUT
    assert ida['tool'] == 'IDA Pro 9.4' and ida['address_space'] == 'IDA linear'
    assert ida['function_count'] == 338 and len(ida['threshold_direct_refs']) == 120
    assert [len(r['instructions']) for r in ida['records']] == [672, 128, 58]
    assert ida['records'][0]['original_name'] == ida['records'][1]['original_name'] == 'sub_1E3A8'
    assert ida['database_sha256'] == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    natural = native('flee3', [49, 23, 4, 4, 3260], 5, 193, 16, 100)
    escaped = native('flee4', [49, 23, 4, 4, 3300], 1, 10, 0, 60)
    assert natural['elapsed'] != escaped['elapsed']
    results = {}
    first_party = None
    for language, scale in [('en', 1), ('zh-Hant', 3)]:
        current = PRIVATE / ('gui-center-' + language)
        previous = ROOT / 'workplace/battle-world-20261002' / ('gui-complete3-' + language)
        rows = read(current / 'trace.json')['steps']
        old = read(previous / 'trace.json')['steps']
        assert len(rows) == 493 and len(old) == 489
        command, old_command = read(current / 'command.json'), read(previous / 'command.json')
        keys = command[command.index('-flow-script') + 1].split()
        old_keys = old_command[old_command.index('-flow-script') + 1].split()
        assert keys == old_keys + ['R', 'D', 'L', 'U']
        assert len(keys) == 492 and '-load' not in command
        assert command[command.index('-scene') + 1] == 'flow'
        metadata = read(current / 'binary.json')
        assert metadata['sha256'] == PROBE and metadata['production_sha256'] == PRODUCT
        assert metadata['normal_startup'] and not metadata['direct_entry'] and not metadata['state_injection']
        assert rows[:480] == old[:480]
        assert rows[119]['battle']['probe_seed'] == 1
        assert max(r['battle']['round'] for r in rows if 'battle' in r) == 61
        image_count = 0
        for i in range(480):
            assert sha(current / f'step-{i:03}.png') == sha(previous / f'step-{i:03}.png'), (language, i)
            image_count += 1
        differences = []
        for i in range(480, 489):
            modified = json.loads(json.dumps(rows[i]))
            assert [modified['world'][k] for k in ['x', 'y', 'sub_x', 'sub_y', 'ticks']] == [49, 23, 4, 4, 3320]
            modified['world']['sub_x'], modified['world']['sub_y'] = 2, 3
            assert modified == old[i], (language, i)
            differences.append(marker_difference(current / f'step-{i:03}.png', previous / f'step-{i:03}.png', scale))
        assert [rows[i]['world']['sub_x'] for i in range(489, 493)] == [5, 5, 4, 4]
        assert [rows[i]['world']['sub_y'] for i in range(489, 493)] == [4, 5, 5, 4]
        assert all([rows[i]['world'][k] for k in ['x', 'y']] == [49, 23] for i in range(489, 493))
        party = rows[480]['town']['probe_full_party']
        assert party == old[480]['town']['probe_full_party']
        if first_party is None:
            first_party = party
        else:
            assert party == first_party
        reload = PRIVATE / ('production-reload-' + language)
        old_reload = ROOT / 'workplace/battle-loot-20261002/fix2' / ('production-reload2-' + language)
        assert read(reload / 'trace.json') == read(old_reload / 'trace.json')
        assert len(read(reload / 'trace.json')['steps']) == 16
        reload_command = read(reload / 'command.json')
        assert reload_command[0] == '/out/kol-world-return-linux'
        save_path = Path(reload_command[reload_command.index('-load') + 1])
        assert read(reload / 'binary.json')['save_sha256'] == sha(save_path)
        for image in reload.glob('*.png'):
            assert sha(image) == sha(old_reload / image.name), image.name
        assert len(list(reload.glob('*.png'))) == 16
        assert rows[480]['world']['ticks'] != 3260
        results[language] = dict(normal_keys=492, states=493, battle_rounds=61,
            unchanged_pre_exit_states=480, unchanged_pre_exit_images=image_count,
            returned=[49, 23, 4, 4, 3320], changed_images=9,
            changed_pixels_per_image=differences, post_exit_moves=4,
            complete_party_unchanged=True, formal_reload_states=16, formal_reload_images=16,
            formal_reload_unchanged=True, full_world_time_parity=False)
    return dict(schema='kol-battle-exits-verification-v1', date='2026-10-02',
        product_sha256=PRODUCT, observation_sha256=PROBE, input_sha256=INPUT,
        dosgolem_revision=REVISION, historical=historical,
        source=dict(natural_defeat=natural, confirmed_menu_escape=escaped), product=results,
        scope='SPEC-011 RE-185次格返回獨立切片；其他世界／戰鬥返回仍DRAFT。',
        limitations=['產品返回時刻3320與原版3260仍不同。',
            '勝利最終暫存及完整世界／戰鬥亂數接續尚未驗。',
            '正式重載使用RE-182正常住宿存檔；本輪沒有從置中後另造住宿存檔。',
            '正常完整GUI涵蓋自然敗北；勝利／逃走置中另以正式共同出口測試驗證。'])


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--export', action='store_true')
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    result = validate()
    if args.export:
        target = Path('/out/receipt.json')
        assert not target.exists()
        paths = [p for p in PRIVATE.rglob('*') if p.is_file()]
        shared = read(PRIVATE / 'parent-preservation.json')['sources']
        paths += [ROOT / name for name in shared]
        paths += [ROOT / 'cmd/kol/battle_world_return_test.go', ROOT / 'tools/png_io.py',
            HERE / 'verify_exits.py', HERE / 'audit_exits.py',
            ROOT / 'docs/re/20261002-battle-exits.md']
        assert all(p.stat().st_uid == p.stat().st_gid == 1000 for p in paths)
        result['attachments'] = {str(p.relative_to(ROOT)): sha(p) for p in sorted(set(paths))}
        target.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    elif not args.check:
        expected = read(HERE / 'receipt.json')
        for name, digest in expected['attachments'].items():
            p = ROOT / name
            assert sha(p) == digest, name
            assert p.stat().st_uid == p.stat().st_gid == 1000, name
        assert result == {k: v for k, v in expected.items() if k != 'attachments'}
    print(json.dumps({k: v for k, v in result.items() if k != 'attachments'}, ensure_ascii=False))


if __name__ == '__main__':
    main()
