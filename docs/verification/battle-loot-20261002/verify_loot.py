"""戰後搜刮正式規則與正常GUI唯讀重驗；--export僅建立一次新收據。"""
from pathlib import Path
import argparse, base64, hashlib, json, os, runpy

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/battle-loot-20261002'
RUN = PRIVATE / 'fix2'
HERE = ROOT / 'docs/verification/battle-loot-20261002'

def read(path):
    return json.loads(path.read_text(encoding='utf-8'))

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def validate_save_paths():
    results = {}
    saves = []
    for lang, lodging in [('en','inn4-en'),('zh-Hant','inn5-zh-Hant')]:
        folder = RUN / lodging
        command = read(folder / 'command.json')
        assert command[command.index('-lang')+1] == lang
        assert '-load' not in command and '-faithful-inn-stay' in command
        keys = command[command.index('-flow-script')+1].split()
        prefix = read(RUN / ('gui-'+lang) / 'command.json')
        assert keys[:500] == prefix[prefix.index('-flow-script')+1].split()
        steps = read(folder / 'trace.json')['steps']
        original_steps = read(RUN / ('gui-'+lang) / 'trace.json')['steps']
        assert len(steps) == len(keys)+1 == 769
        for a,b in zip(steps[:501],original_steps):
            assert a.get('battle') == b.get('battle'), (lang,a['step'])
            assert a['town'].get('probe_full_party') == b['town'].get('probe_full_party') if 'town' in a and 'town' in b else True
        saved = read(folder / 'complete.json')
        assert saved['schema'] == 'kol-save-v3' and len(saved['characters']) == 16
        assert saved['party'] == [0] and saved['settings']['faithful_inn_stay']
        assert saved['location'] == dict(kind='town',town=7,entry_point=3,x=57,y=8)
        assert (saved['world']['x'],saved['world']['y']) == (40,23)
        by_name = {m['name']:m for m in saved['characters'] if m is not None}
        after = original_steps[492]['town']['probe_full_party']
        for m in after:
            s = by_name[m['name']]
            assert s['carried'] == [base64.b64decode(x).hex() for x in m['carried']]
            assert s['wounds'] == m['wounds']
            assert s['gold'] == m.get('gold',0)-(56 if m['name']=='Olear' else 0)
        assert by_name['Olear']['gold'] == 1721 and by_name['Theodoric']['gold'] == 0
        assert len(by_name['Olear']['backpack']) == 5 and by_name['Olear']['backpack'][2]['name'] == ''
        assert by_name['Theodoric']['backpack'] is None
        reload_folder = RUN / ('reload1-'+lang)
        reload_command = read(reload_folder / 'command.json')
        assert reload_command[reload_command.index('-load')+1] == str(Path('/out')/lodging/'complete.json')
        assert reload_command[reload_command.index('-scene')+1] == 'play'
        assert reload_command[reload_command.index('-play-script')+1] == 'L R RET SPACE SPACE > > > > > RET RET D RET SPACE'
        rows = read(reload_folder / 'trace.json')['steps']
        assert len(rows) == 16 and len(list(reload_folder.glob('step-*.png'))) == 16
        assert [m['name'] for m in rows[0]['town']['probe_full_party']] == ['Theodoric']
        final = rows[-1]['town']['probe_full_party']
        assert [m['name'] for m in final] == ['Theodoric','Olear']
        for m in final:
            s = by_name[m['name']]
            assert [base64.b64decode(x).hex() for x in m['carried']] == s['carried']
            for key in ['id','wounds','place','town_code','building_id','sheet','quests','progress']:
                assert m[key] == s[key], (lang,m['name'],key)
            assert m.get('gold',0) == s['gold']
            for key in ['in_hand','backpack','worn']:
                expected = s[key] if isinstance(s[key],list) else [s[key]]
                actual = m.get(key) if isinstance(m.get(key),list) else [m.get(key)]
                simplify = lambda xs: [(x.get('name',''),x.get('record',''),x.get('condition',0)) if x else ('','',0) for x in xs]
                assert simplify(actual) == simplify(expected), (lang,m['name'],key)
        production_folder = RUN / ('production-reload2-'+lang)
        production_command = read(production_folder/'command.json')
        assert production_command[0] == '/out/kol-battle-loot-linux'
        production_rows = read(production_folder/'trace.json')['steps']
        assert len(production_rows) == 16
        for row, production_row in zip(rows,production_rows):
            observed = dict(row)
            observed['town'] = dict(row['town'])
            observed['town'].pop('probe_full_party')
            assert observed == production_row, (lang,row['step'])
            image_name = f"step-{row['step']:03d}.png"
            assert sha(reload_folder/image_name) == sha(production_folder/image_name), (lang,image_name)
        comparable = dict(saved)
        comparable.pop('saved_at')
        saves.append(comparable)
        results[lang] = dict(normal_keys=len(keys),states=len(steps),saved_roster_slots=16,
            saved_characters=len(by_name),
            reload_states=len(rows),recruited='Olear',matched_inventory_slots=34,
            wounds=[24,19],gold=[1721,0],lodging_price=56,
            process_restart=True,production_matching_frames=16,save_schema='kol-save-v3')
    assert saves[0] == saves[1]
    return results

def validate():
    source = read(PRIVATE / 'loot1/complete-source-private.json')
    old = read(ROOT / 'workplace/battle-settlement-20261001/post1/complete-source-private.json')
    for key in ['initial','inputs','boundaries','rolls','messages','post_battle']:
        assert source[key] == old[key], key
    assert source['rounds'] == 61 and len(source['inputs']) == 351 and len(source['rolls']) == 1491
    assert source['normal_startup'] and not any(source[k] for k in ['state_injection','direct_entry','wait_stub'])
    controls = {r['ida']:r for r in source['exit_control']}
    for point in ['0x21607','0x2175B','0x21765']:
        assert controls[point]['scratch_cd8a'] == 0
    assert controls['0x2186E']['seed'] == 4284445587
    cases = read(PRIVATE / 'loot-source-cases-private.json')
    assert cases['no_stubs'] and not cases['normal_player_path'] and len(cases['cases']) == 144
    assert '144組' in (PRIVATE / 'source-cases.log').read_text()
    residual = read(PRIVATE / 'loot-residual-cases-private.json')
    assert residual['no_stubs'] and len(residual['cases']) == 144
    assert '144組' in (PRIVATE / 'residual-source-cases.log').read_text()
    drops = read(PRIVATE / 'drop-source-cases-private.json')
    assert drops['no_stubs'] and len(drops['cases']) == 12
    log = (PRIVATE / 'rules2.log').read_text()
    assert 'FAIL' not in log and 'cmd/kol' in log and 'internal/town' in log
    focused_log = (PRIVATE / 'rules3.log').read_text()
    assert 'FAIL' not in focused_log and 'TestForcedDropOriginalCases' in focused_log and 'TestBattleLoot' in focused_log
    parent = read(PRIVATE / 'parent-preservation.json')
    assert parent['re180_inputs'] == 1213 and len(parent['sources']) == 1223
    for r in parent['sources']:
        assert sha(PRIVATE / 'parent-inputs' / r['path']) == r['sha256'], r['path']
    helper = runpy.run_path(str(ROOT / 'docs/verification/battle-settlement-20261002/verify.py'))
    actor_diffs = helper['actor_diffs']
    results = {}
    traces = {}
    for lang in ['en','zh-Hant']:
        folder = RUN / ('gui-' + lang)
        steps = read(folder / 'trace.json')['steps']
        assert len(steps) == 501 and len(list(folder.glob('step-*.png'))) == 501
        traces[lang] = steps
        entries = []
        for boundary in source['boundaries']:
            if not boundary['at'].startswith('執行段入口'):
                continue
            state = boundary['state']
            rows = [r for r in steps if r.get('battle',{}).get('round') == boundary['round']
                    and r.get('battle',{}).get('probe_seed') == state['seed']
                    and r.get('battle',{}).get('prompt') == 8]
            assert len(rows) == 1 and not actor_diffs(state, rows[0]['battle']['probe_actors']), boundary['round']
            entries.append(rows[0]['step'])
        assert len(entries) == 61
        final = next(r for r in steps if r.get('battle',{}).get('stage') == 5)
        b = final['battle']
        assert (b['round'],b['verdict']) == (61,1) and b['probe_seed'] == 1232439027
        assert actor_diffs(source['battle_exit_state'],b['probe_actors']) == [(6,'Command',0,33)]
        after_drop = next(r for r in steps if r.get('battle',{}).get('round') == 25 and r['battle']['stage'] == 2)
        assert after_drop['battle']['probe_party'][0]['HandItem'] is None
        assert not after_drop['battle']['probe_party'][0]['HasWeapon']
        assert [(r['N'],r['Before']) for r in b['probe_rolls']] == [(r['n'],r['seed_before']) for r in source['rolls'][:1458]]
        after = steps[final['step']+1]['town']['probe_full_party']
        original = source['post_battle'][1]['state']['characters']
        offsets = [0x30]+list(range(0xA0,0x1A0,16))
        for slot,m in enumerate(after):
            raw = bytes.fromhex(original[slot])
            assert m['wounds'] == raw[0x1B]
            assert m.get('gold',0) == int.from_bytes(raw[0x2A:0x2D],'big')
            carried = [base64.b64decode(x) for x in m['carried']]
            assert len(carried) == 17
            for off, actual in zip(offsets,carried):
                assert actual == raw[off:off+16], (lang,slot,hex(off),actual.hex(),raw[off:off+16].hex())
            assert base64.b64decode(m['carried'][3]) == bytes(16)
        assert [m['wounds'] for m in after] == [24,19] and after[1].get('gold',0) == 0
        results[lang] = dict(states=501,execution_entries=61,combat_rolls=1458,
            defeat_step=final['step'],world_step=final['step']+1,
            wounds=[24,19],gold=[m.get('gold',0) for m in after],matched_inventory_slots=34,
            final_command_differences=[[6,'Command',0,33]])
    for a,b in zip(traces['en'],traces['zh-Hant']):
        assert a['town'].get('probe_full_party') == b['town'].get('probe_full_party') if 'town' in a and 'town' in b else True
    index_path = ROOT / 'docs/re/battle-loot-address-index.json'
    index = read(index_path)
    export = read(PRIVATE / 'loot-control-annotated2-private.json')
    assert export['function_count'] == 338 and export['tool'] == index['tool'] == 'IDA Pro 9.4'
    assert export['sha256'] == index['sha256'] and export['address_space'] == index['address_space'] == 'IDA linear'
    hand_export = read(PRIVATE / 'hand-drop-annotated-private.json')
    rows = {r['original_ea']:r for e in [export,hand_export] for f in e['records'] for r in f['instructions']}
    for address,note in index['entries'].items():
        r = rows[address]
        assert r['bytes'] and r['operand_text'] and any(a['semantic']==note['semantic'] and
            a['index_sha256']==sha(index_path) and a['confidence']=='已證實' for a in r['annotations'])
    negative = {}
    for lang in ['en','zh-Hant']:
        old_steps = read(PRIVATE / ('gui-'+lang) / 'trace.json')['steps']
        final = next(r for r in old_steps if r.get('battle',{}).get('stage') == 5)
        after = old_steps[final['step']+1]['town']['probe_full_party']
        differences = [(slot,hex(off)) for slot,m in enumerate(after)
            for off,encoded in zip(offsets,m['carried']) if base64.b64decode(encoded) !=
            bytes.fromhex(original[slot])[off:off+16]]
        assert differences == [(0,'0x30'),(1,'0xa0'),(1,'0xb0'),(1,'0xd0'),(1,'0xe0')]
        negative[lang] = [list(difference) for difference in differences]
    return dict(schema='kol-battle-loot-verification-v1',date='2026-10-02',
        original_input=dict(filename='KOLTOWN-unpacked.EXE',
            sha256='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2',
            archive='Knights-of-legend.zip',archive_sha256='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',
            disassembler='IDA Pro 9.4',address_space='IDA linear',
            dosgolem_revision='968ca774705853276ae55c8d1163c369481ac71f'),
        controlled_random=dict(normal_original_seed=1,normal_remake_seed=1,
            seed_point='每場首次規劃開始；原版DS:1600／1602，重製注入LCG；正式程式不鎖種子',
            loot_cases_initial_seed=1232439027,drop_cases_initial_seed=1,
            comparison='相同初態、等價正常按鍵及同狀態入口；不是只以seed數字相同判定通過'),
        production_sha256=sha(RUN/'kol-battle-loot-linux'),probe_sha256=sha(RUN/'kol-settlement-probe2-linux'),
        source_cases=288,drop_source_cases=12,preserved_parent_inputs=1213,source_prefix_unchanged=True,merged_annotations=9,
        rejected_old_inventory=negative,
        normal_save_reload=validate_save_paths(),
        normal_gui=results,limitations=['首次規劃相同時點固定測試seed；正式程式不永久固定種子。',
        '正常GUI確認戰後傷勢、錢包與兩人17格物品；最後32骰正常GUI未另加逐骰觀察，規則骰序由原版288組案例比較。',
        '原版世界49,23、重製51,23及終態指令暫存差異仍明列，非全世界或全欄位parity。',
        '敗北後雙語正常住宿保存、結束行程、重載及重新招募已閉合；其他戰鬥分支未閉合，父規格維持READY。',
        '住宿續行中的自然遇敵只驗重製路徑；研究副本每場首次規劃固定seed 1，不外推為原版對拍。'])

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('--export',action='store_true')
    args=parser.parse_args()
    result=validate()
    target=HERE/'receipt.json'
    if args.export:
        assert not target.exists()
        paths=[p for p in PRIVATE.rglob('*') if p.is_file()]
        paths += [ROOT/'docs/re/20261002-battle-loot.md',ROOT/'docs/spec/027-battle-exit-wounds.md',
                  ROOT/'docs/re/battle-loot-address-index.json',HERE/'verify_loot.py',HERE/'audit.py']
        paths += [p for folder in ['cmd','internal','assets'] for p in (ROOT/folder).rglob('*') if p.is_file()]
        paths += [ROOT/'go.mod',ROOT/'go.sum']
        result['attachments']={str(p.relative_to(ROOT)):sha(p) for p in sorted(set(paths))}
        assert all(p.stat().st_uid==p.stat().st_gid==1000 for p in paths)
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    else:
        expected=read(target)
        for name,digest in expected['attachments'].items():
            assert sha(ROOT/name)==digest,name
        assert {k:v for k,v in expected.items() if k!='attachments'} == result
    print(json.dumps({k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))

if __name__=='__main__':
    main()
