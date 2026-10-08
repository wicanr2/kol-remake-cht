"""唯讀重驗正常連續戰鬥與戰後來源；--export只建立一次新收據。"""
from pathlib import Path
import argparse, hashlib, json, sys, zipfile

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/battle-settlement-20261001'
HERE = ROOT / 'docs/verification/battle-settlement-20261002'
FIELDS = dict(X=31, Y=32, Damage=14, Head=17, Chest=19, RightArm=21, LeftArm=23,
              RightLeg=25, LeftLeg=27, Parries=15, Facing=40, Hurt=45, Command=37,
              Style=35, Guard=36, Flag1D=29, Flag1E=30, LastHurt=42, PickedUp=43)
PARTS = {'Head', 'Chest', 'RightArm', 'LeftArm', 'RightLeg', 'LeftLeg'}

def read(path):
    return json.loads(path.read_text())

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def slim_state(state):
    return {k: v for k, v in state.items() if k != 'exit_context'}

def actor_diffs(state, actors, fields=FIELDS):
    errors = []
    for slot in [0, 1, 6]:
        raw = bytes.fromhex(state['actors'][slot])
        a = actors[slot]
        for key, off in fields.items():
            value = raw[off]
            if key in PARTS and value == 255:
                value = 0
            if value != a[key]:
                errors.append((slot, key, value, a[key]))
        for key in ['fatigue', 'stamina']:
            if state['counters'][slot][key] != a[key.title()]:
                errors.append((slot, key))
    return errors

def continuous(source, steps):
    entries = []
    for boundary in source['boundaries']:
        if not boundary['at'].startswith('執行段入口'):
            continue
        s = boundary['state']
        candidates = [row for row in steps if row.get('battle', {}).get('round') == boundary['round']
                      and row.get('battle', {}).get('probe_seed') == s['seed']
                      and row.get('battle', {}).get('prompt') == 8]
        assert len(candidates) == 1, boundary['round']
        row = candidates[0]
        assert not actor_diffs(s, row['battle']['probe_actors']), boundary['round']
        entries.append(dict(round=boundary['round'], step=row['step']))
    assert len(entries) == 61
    final = next(row for row in steps if row.get('battle', {}).get('stage') == 5)
    b = final['battle']
    assert (b['round'], b['verdict']) == (61, 1)
    assert b['probe_seed'] == source['battle_exit_state']['seed']
    final_diffs = actor_diffs(source['battle_exit_state'], b['probe_actors'])
    assert final_diffs == [(1, 'Command', 0, 41), (6, 'Command', 0, 33)]
    assert [(r['N'], r['Before']) for r in b['probe_rolls']] == [
        (r['n'], r['seed_before']) for r in source['rolls'][:1458]]
    core = {k: v for k, v in FIELDS.items() if k in PARTS or k in ['X', 'Y', 'Damage', 'Parries', 'Hurt']}
    for boundary in source['boundaries']:
        if not boundary['at'].startswith('下一回合規劃'):
            continue
        n = boundary['round'] + 1
        row = next(row for row in steps if row.get('battle', {}).get('round') == n
                   and row.get('battle', {}).get('stage') == 2)
        assert not actor_diffs(boundary['state'], row['battle']['probe_actors'], core), n
        waiting = next(r for r in source['inputs'] if r['round'] == n and r['phase'] == 'plan')
        assert row['battle']['probe_seed'] == waiting['state']['seed'], n
    party = steps[final['step'] + 1]['town']['party']
    assert 'wounds=24' in party[0] and 'wounds=19' in party[1]
    # 保留產品缺口：沒有源自原版的32次搜刮，Theodoric的1009金幣仍在。
    assert 'gold=1009' in party[1]
    return dict(execution_entries=entries, next_planning_states=60, rolls=1458,
                final_step=final['step'], final_seed=b['probe_seed'], wounds=[24, 19],
                final_command_differences=final_diffs, loot_implemented=False)

def source_model(source, omit_loss=False):
    start = source['battle_exit_state']
    ctx = start['exit_context']
    assert ctx['protect_word_6a98'] == 0
    chars = [bytearray.fromhex(x) for x in start['characters']]
    seed, roll_index = start['seed'], 1458
    lost = []
    def roll(caller):
        nonlocal seed, roll_index
        r = source['rolls'][roll_index]
        assert r['n'] == 100 and r['caller_ida'] == caller and r['seed_before'] == seed
        roll_index += 1
        seed = (seed * 214013 + 2531011) & 0xFFFFFFFF
        return ((seed >> 16) & 32767) % 100 + 1
    for slot in range(6):
        a = bytes.fromhex(start['actors'][slot])
        if a[1] != 0:
            continue
        if a[17] in [0, 255] or a[19] in [0, 255]:
            for off in range(0xA0, 0x1A0, 16):
                if off == 0x170:
                    continue
                value = roll('0x217E1')
                if value < ctx['threshold_byte_cdad'] and not omit_loss:
                    if any(chars[slot][off:off+16]):
                        lost.append(dict(slot=slot, offset=hex(off), roll=value))
                    chars[slot][off:off+16] = bytes(16)
            value = roll('0x21844')
            if value < ctx['threshold_byte_cdad'] and not omit_loss:
                chars[slot][0x2A:0x2D] = bytes(3)
                lost.append(dict(slot=slot, gold=True, roll=value))
        chars[slot][0x1B] = (chars[slot][0x1B] + ((a[14]-chars[slot][0x1B]) & 65535)//2) & 255
    return [x.hex() for x in chars], seed, roll_index-1458, lost

def result():
    one = read(PRIVATE / 'post1/complete-source-private.json')
    source = read(PRIVATE / 'post2/complete-source-private.json')
    assert source['normal_startup'] and not any(source[k] for k in ['direct_entry', 'state_injection', 'wait_stub'])
    assert source['battle_exit_counts'] == dict(rounds=61, inputs=351, waits=350, rolls=1458, messages=1318, boundaries=122)
    for key in ['inputs', 'boundaries']:
        stripped = [{**r, 'state': slim_state(r['state'])} for r in source[key]]
        assert stripped == one[key], key
    assert source['rolls'] == one['rolls'] and source['messages'] == one['messages']
    assert slim_state(source['battle_exit_state']) == one['battle_exit_state']
    assert len(source['post_battle']) == 8
    assert source['post_battle'][0]['last_print'] == 378
    assert source['post_battle'][1]['last_print'] == 225
    assert [bytes.fromhex(x)[27] for x in source['post_battle'][1]['state']['characters']][:2] == [24, 19]
    chars, loot_seed, count, lost = source_model(source)
    assert chars == source['post_battle'][1]['state']['characters'] and count == 32
    assert source_model(source, omit_loss=True)[0] != chars
    assert source['rolls'][1490]['n'] == 128 and source['rolls'][1490]['seed_before'] == loot_seed
    sys.path.insert(0, str(ROOT / 'tools'))
    from probe_town_data import read_volumes, payloads, lzw_decode
    archive = ROOT / 'Knights-of-legend.zip'
    assert sha(archive) == '2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46'
    with zipfile.ZipFile(archive) as z:
        _, pieces = payloads(read_volumes(z, 'TOWN0'), 'TOWN0')
    monster = lzw_decode(pieces[32])
    config = bytes.fromhex(source['battle_exit_state']['exit_context']['config_cd8a_hex'])
    assert monster[:32] == config[8:40] and monster[27] == 20
    assert sha(ROOT / 'workplace/battle-complete-20261001/kol-battle-rules-linux') == 'b1264c4e12452cad59ed015bab02bd339c967b3da41bbab037019faf6d28f058'
    assert sha(PRIVATE / 'kol-settlement-probe2-linux') == '03b33761128fdbb3468631f6976db1e2ea000acbf9acd3a88dfb2cf35f8dfc79'
    traces = [read(PRIVATE / name / 'trace.json')['steps'] for name in ['gui4', 'gui5-zh']]
    comparisons = [continuous(source, trace) for trace in traces]
    for en, zh in zip(*traces):
        for k in ['probe_actors', 'probe_seed', 'probe_rolls', 'probe_party']:
            assert en.get('battle', {}).get(k) == zh.get('battle', {}).get(k), (en['step'], k)
    assert len(traces[0]) == len(traces[1]) == 501
    ida = read(PRIVATE / 'settlement-ida2-private.json')
    assert ida['tool'] == 'IDA Pro 9.4' and ida['address_space'] == 'IDA linear' and ida['function_count'] == 338
    assert ida['database_sha256'] == '5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8'
    assert sha(ROOT / 'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64') == ida['database_sha256']
    return dict(schema='kol-normal-battle-settlement-v1', date='2026-10-02', production_changed=False,
                original=dict(revision=source['revision'], seed=1, setup=source['seed_setup'], rounds=61,
                              waits_before_exit=350, post_waits=8, last_post_key_consumed=False),
                remake=dict(probe_only=True, seed=1, seed_setup='首次規劃入口原地重設共用LCG，沒有重排初始隊伍',
                            observations={'en': comparisons[0], 'zh-Hant': comparisons[1]}),
                source_settlement=dict(characters_exact_bytes=6*768, loot_rolls=count, loot_seed=loot_seed,
                                       threshold=20, lost=lost, prototype_only=True),
                negative_control='省略搜刮即無法重建正常原版角色；首版觀察器gui3不作驗收。',
                limitations=['正式程式尚未實作32次搜刮；未驗住宿保存／重載',
                             '正常操作持續且未注入角色或座標；測試副本只控制與原版相同的播種時點',
                             '原版世界位置49,23；重製51,23，世界完整同狀態尚未證實',
                             '訊息動畫分段不同，不宣稱逐輸入中間畫格一致',
                             'word_cd8a是重用暫存，post2歷史鍵名arena_word_cd8a不能當固定競技場旗標',
                             '未外推任務保護／競技場或其他怪物正常搜刮路徑'])

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--export', action='store_true')
    args = parser.parse_args()
    actual = result()
    if args.export:
        target = Path('/out/receipt.json')
        assert not target.exists()
        paths = sorted(p for p in PRIVATE.rglob('*') if p.is_file())
        paths += [HERE / 'verify.py']
        actual['attachments'] = [dict(path=str(p.relative_to(ROOT)), sha256=sha(p)) for p in paths]
        target.write_text(json.dumps(actual, ensure_ascii=False, indent=2)+'\n')
    else:
        stored = read(HERE / 'receipt.json')
        for record in stored.pop('attachments'):
            p = ROOT / record['path']
            assert sha(p) == record['sha256'], p
            assert p.stat().st_uid == p.stat().st_gid == 1000, p
        assert stored == actual
    print('正常連續61回合、雙語61入口／60次新規劃／1458骰及戰後來源重建通過；產品搜刮與保存鏈仍待完成。')

if __name__ == '__main__':
    main()
