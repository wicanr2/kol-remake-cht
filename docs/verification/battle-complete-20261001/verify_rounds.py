"""唯讀重驗 RE-180 的原始來源、61 回合結果與輸入鎖定；不擴張完整玩家鏈聲明。"""
from pathlib import Path
import argparse, hashlib, json, os, re

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--root', default='/src')
ap.add_argument('--export', action='store_true')
args = ap.parse_args()
root = Path(args.root)
work = root / 'workplace/battle-complete-20261001'
proof = root / 'docs/verification/battle-complete-20261001'

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def load(path):
    return json.loads(path.read_text(encoding='utf-8'))

def source_contract(data):
    assert data['schema'] == 'kol-normal-complete-source-v1'
    assert data['tool'] == 'dosgolem' and data['revision'] == '968ca774705853276ae55c8d1163c369481ac71f'
    assert data['input_sha256'] == '01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
    assert data['seed'] == 1 and data['normal_startup'] and data['reached_battle_exit']
    assert not data['state_injection'] and not data['direct_entry'] and not data['wait_stub']
    assert data['rounds'] == 61 and len(data['boundaries']) == 122
    assert len(data['inputs']) == 351 and len(data['wait_calls']) == 350
    assert len(data['rolls']) == 1458 and len(data['messages']) == 1318
    for number in range(1, 62):
        start, end = data['boundaries'][2 * (number - 1):2 * number]
        assert start['round'] == end['round'] == number
        assert start['at'] == '執行段入口 0x1FD21'
        assert end['at'] == ('戰鬥出口 0x211AD' if number == 61 else '下一回合規劃 0x1EF3A')
        seed = start['state']['seed']
        for roll in data['rolls']:
            if roll['round'] == number and roll['phase'] == 'exec':
                assert roll['seed_before'] == seed
                seed = (seed * 0x343FD + 0x269EC3) & 0xFFFFFFFF
        assert seed == end['state']['seed']

first = load(work / 'original1/complete-source-private.json')
second = load(work / 'original2/complete-source-private.json')
source_contract(first)
source_contract(second)
for key in ['boundaries', 'inputs', 'messages', 'wait_calls', 'initial']:
    assert first[key] == second[key], key
assert first['rolls'] == [{k: v for k, v in r.items() if k != 'execution_gear'} for r in second['rolls']]
gear = [r['execution_gear'] for r in second['rolls'] if 'execution_gear' in r]
assert len(gear) == 183
monster = [g for g in gear if g['actor'] == 6]
assert len(monster) == 61 and all((g['level_byte_02'], g['encumbrance'], g['weapon_speed']) == (1, 87, 35) for g in monster)
adapter = load(work / 'rounds-adapter-private.json')
assert adapter == dict(schema='kol-battle-rounds-v1', boundaries=first['boundaries'], rolls=first['rolls'])
historical = load(root / 'workplace/battle-first-round-20260928/battle-rounds.json')
for a, b in zip(first['boundaries'][:12], historical['boundaries'][:12]):
    assert a['round'] == b['round'] and a['at'] == b['at']
    for key in ['seed', 'actors', 'counters']:
        assert a['state'][key] == b['state'][key]
log = (work / 'compare-rounds-final.log').read_text()
assert '--- PASS: TestOriginalNormalRoundsExecution' in log and '\nFAIL' not in log
assert sorted(map(int, re.findall(r'--- PASS: TestOriginalNormalRoundsExecution/第(\d+)回合', log))) == list(range(1, 62))
for name in ['tests-rules.log', 'tests-old-normal.log']:
    text = (work / name).read_text()
    assert '\nFAIL' not in text and 'ok  ' in text
assert '建場：負重等級 3，負重 75，速度 30' in (work / 'monster-load-before2.log').read_text()
assert '第 1 次擲骰面數 3' in (work / 'attack-order-before.log').read_text()
for name in ['TestMonsterGearUsesCombatLevel', 'TestResolveRandomStyleOrder']:
    assert name in (root / 'internal/battle/attack_test.go').read_text()
items = load(work / 'normal-items-comparison.json')
assert len(items['records']) == 2 and {r['language'] for r in items['records']} == {'en', 'zh-Hant'}
gui_paths = []
for record in items['records']:
    old = root / record['baseline']
    new = work / record['output']
    assert record['states'] == record['unchanged_frames'] == 110
    assert load(new / 'trace.json')['steps'] == load(old / 'trace.json')['steps']
    shots = sorted(new.glob('step-*.png'))
    assert len(shots) == 110
    for shot in shots:
        assert shot.read_bytes() == (old / shot.name).read_bytes()
    gui_paths += shots + [new / 'command.json', new / 'trace.json', old / 'command.json', old / 'trace.json']
index = root / 'docs/re/battle-order-address-index.json'
annotated = load(work / 'battle-order-annotated-private.json')
assert annotated['function_count'] == 338 and annotated['database_sha256'] == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
joined = {row['original_ea']: row for f in annotated['records'] for row in f['instructions']}
for address, semantics in load(index)['entries'].items():
    row = joined[address]
    assert row['bytes'] and row['operand_text'] and any(a['semantic'] == semantics['semantic'] and a['confidence'] == '已證實' and a['index_sha256'] == digest(index) for a in row['annotations'])
for altered in [dict(first, seed=2), dict(first, wait_stub=True)]:
    try:
        source_contract(altered)
    except AssertionError:
        pass
    else:
        raise AssertionError('錯誤來源反例未拒絕')
binary = work / 'kol-battle-rules-linux'
assert digest(binary) == 'b1264c4e12452cad59ed015bab02bd339c967b3da41bbab037019faf6d28f058'
current_paths = [root / 'go.mod', root / 'go.sum', index]
for parent, pattern in [('internal', '*.go'), ('cmd/kol', '*.go'), ('assets', '*')]:
    current_paths += [p for p in (root / parent).rglob(pattern) if p.is_file()]
private_paths = [binary, work / 'rounds-adapter-private.json', work / 'party-fixture.json',
                 work / 'battle-order-ida-private.json', work / 'battle-order-annotated-private.json',
                 work / 'comparison-adapter-inputs.json', work / 'normal-items-comparison.json']
private_paths += list(work.glob('*.log')) + list(work.glob('*.py')) + list(work.glob('*.sh'))
for name in ['original1', 'original2']:
    private_paths += [work / name / leaf for leaf in ['complete-source-private.json', 'inputs.json', 'instrumented-book-startup-private.go', 'run.log', 'battle-state-records.json', 'battle-alignment.json', 'battle-terrain.json']]
private_paths += gui_paths

def locks(paths):
    result = {}
    for p in sorted(set(paths)):
        assert p.stat().st_uid == os.getuid() == 1000 and p.stat().st_gid == os.getgid() == 1000, p
        result[str(p.relative_to(root))] = digest(p)
    return result

receipt = dict(schema='kol-complete-battle-execution-receipt-v1', date='2026-10-01',
               confidence='已證實：此自然來源的61個逐回合執行段；非完整正常重製玩家鏈',
               binary_sha256=digest(binary), original_seed=1,
               remake_seed_setup='每回合執行前固定為來源入口的32位元狀態；同一LCG實作',
               original_normal_inputs=351, original_rounds=61, compared_execution_rounds=61,
               source_unchanged_boundaries=122, gear_observations=183,
               monster_gear=dict(level=1, encumbrance=87, weapon_speed=35),
               normal_items=items['records'], current_inputs=locks(current_paths), attachments=locks(private_paths),
               limitations=['重製版每回合接手來源狀態，未驗連續GUI及所有規劃分支',
                            '原版在0x211AD入口停止，戰後回寫與保存未驗',
                            '全部怪物／勝利／戰利品及其他戰鬥種類仍由V01追蹤'],
               rights='完整原版payload與暫存只留本機；沒有正式封包或發布',
               tools=dict(dosgolem='968ca774705853276ae55c8d1163c369481ac71f',
                          ida='IDA Pro 9.4／IDA linear', go='1.26.7', ebitengine='2.9.9',
                          package_image='sha256:3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193',
                          build_image='sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa',
                          ida_image='sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780'))
target = proof / 'receipt.json'
if args.export:
    assert not target.exists() and proof.stat().st_uid == os.getuid()
    target.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
else:
    assert load(target) == receipt, '收據或輸入鎖定不同；不得覆寫歷史基準'
print('61執行段、183負重觀察、220正常物品畫格／狀態與七筆IDA索引核實通過；完整玩家鏈未宣稱完成。')
