"""世界進場差異的唯讀重驗；歷史輸入從封存副本驗證，收據只建立一次。"""
from pathlib import Path
import argparse, hashlib, json, runpy

ROOT = Path('/src')
PRIVATE = ROOT/'workplace/battle-world-20261002'
HERE = ROOT/'docs/verification/battle-world-20261002'
LEGACY = ROOT/'docs/verification/battle-loot-20261002'

def read(path):
    return json.loads(path.read_text())

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def history():
    parent = read(PRIVATE/'parent-preservation.json')
    assert parent['schema'] == 'kol-world-parent-preservation-v1'
    assert len(parent['sources']) == 1299
    for name,digest in parent['sources'].items():
        p = PRIVATE/'parent-inputs'/name
        assert sha(p) == digest,name
        assert p.stat().st_uid == p.stat().st_gid == 1000,name
    receipt = read(LEGACY/'receipt.json')
    assert sha(LEGACY/'receipt.json') == parent['receipt_sha256']
    def check(locks):
        for name,digest in locks.items():
            p = PRIVATE/'parent-inputs'/name if name in parent['sources'] else ROOT/name
            assert sha(p) == digest,name
    check(receipt['attachments'])
    check(read(LEGACY/'audit.json')['locks'])
    runtime = read(LEGACY/'runtime-audit.json')
    assert sha(LEGACY/'audit.json') == runtime['audit_sha256']
    assert sha(LEGACY/'audit_runtime.py') == runtime['verifier_sha256']
    check(runtime['additional_runtime_and_index_inputs'])
    prior = runpy.run_path(str(LEGACY/'verify_loot.py'))['validate']()
    assert prior == {k:v for k,v in receipt.items() if k!='attachments'}
    return dict(preserved_shared_inputs=1299,legacy_attachments=len(receipt['attachments']),
        legacy_supplemental_inputs=len(runtime['additional_runtime_and_index_inputs']),
        legacy_result_unchanged=True,scope='RE-182歷史輸入；不把封存文件當作RE-183目前現況。')

def validate():
    historical = history()
    source = read(PRIVATE/'source2/world-prefix-private.json')
    assert source['normal_startup'] and not any(source[k] for k in ['direct_entry','state_injection','wait_stub'])
    assert source['revision'] == '968ca774705853276ae55c8d1163c369481ac71f'
    assert len(source['inputs']) == 163 and len(source['events']) == 184
    for name in ['battle-state-records.json','battle-planning.json','battle-alignment.json','battle-terrain.json']:
        assert sha(PRIVATE/'source2'/name) == sha(ROOT/'workplace/battle-loot-20261002/loot1'/name),name
    battle_entry = next(x['state'] for x in source['events'] if x['ida']=='0x1E3A8')
    assert (battle_entry['x'],battle_entry['y'],battle_entry['sub_x'],battle_entry['sub_y'],battle_entry['time']) == (49,23,2,3,3360)
    before = [x for x in source['inputs'] if x['key'].startswith('before-walker:')]
    after = [x for x in source['inputs'] if x['key'].startswith('after-walker:')]
    assert len(before) == len(after) == 22
    fields = ['x','y','sub_x','sub_y']
    completed = [a['key'].split(':')[1] for a,b in zip(before,after)
                 if any(a['state'][k]!=b['state'][k] for k in fields)]
    assert completed == ['Right','Down','Left','Up']*4
    results = {}
    for lang in ['en','zh-Hant']:
        old_folder = ROOT/'workplace/battle-loot-20261002/fix2'/('gui-'+lang)
        old = read(old_folder/'trace.json')['steps']
        folder = PRIVATE/('gui-complete3-'+lang)
        rows = read(folder/'trace.json')['steps']
        command = read(folder/'command.json')
        keys = command[command.index('-flow-script')+1].split()
        assert len(keys) == 488 and len(rows) == 489
        assert len(list(folder.glob('step-*.png'))) == 489
        assert '-load' not in command and '-scene' in command
        assert keys[30:101] == ['R']*71 and keys[101]=='SPACE'
        assert keys[102:118] == ['R','D','L','U']*4
        for a,b in zip(old[141:],rows[129:]):
            assert a.get('battle') == b.get('battle'),(lang,a['step'],b['step'])
            assert a.get('town',{}).get('probe_full_party') == b.get('town',{}).get('probe_full_party'),(lang,a['step'])
        assert rows[119]['battle']['probe_seed'] == 1
        assert rows[118]['battle']['stage'] == 1
        world = rows[118]['world']
        assert (world['x'],world['y'],world['sub_x'],world['sub_y'],world['ticks']) == (49,23,2,3,3320)
        assert (rows[-1]['world']['x'],rows[-1]['world']['y']) == (49,23)
        assert (old[117]['world']['x'],old[117]['world']['y']) == (51,23)
        exit_party = rows[480]['town']['probe_full_party']
        assert [x['wounds'] for x in exit_party] == [24,19]
        assert [x.get('gold',0) for x in exit_party] == [1777,0]
        # 首兩版漏播種／切片錯誤必須被拒絕，保留實際反例。
        for phase,point in [('gui-complete1-',135),('gui-complete2-',129)]:
            bad = read(PRIVATE/(phase+lang)/'trace.json')['steps']
            assert bad[point]['battle']['probe_seed'] != 1
        results[lang] = dict(keys=488,states=489,frames=489,normal_startup=True,
            controlled_first_planning_seed=1,battle_states_matching_legacy=360,
            exit_inventory_slots=34,wounds=[24,19],gold=[1777,0],
            world_entry=[49,23],world_exit=[49,23],entry_time=3320)
    return dict(schema='kol-world-entry-verification-v1',date='2026-10-02',
        original_input=dict(filename='KOLTOWN-unpacked.EXE',
            sha256=source['input_sha256'],archive='Knights-of-legend.zip',
            archive_sha256='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46',
            dosgolem_revision=source['revision'],address_space='IDA linear；狀態DGROUP相對位移'),
        historical=historical,normal_source=dict(inputs=163,observations=184,step_clock_calls=87,
            battle_entry=[49,23],entry_time=3360,completed_right_steps=71,completed_walker_steps=16,
            unchanged_initial_battle_files=4),normal_gui=results,
        controlled_random='沿用RE-182旅店後原版seed1及首次規劃兩側seed1；正式程式不鎖種子。',
        production_unchanged=True,
        limitations=['完成移動對齊；原版100個送鍵與22次walker不等於重製逐鍵輸入。',
            '兩世界大格差異已定位為進場前路徑不同；相同完成步可到相同49,23。',
            '原版進場時刻3360、重製3320，仍待查扣時刻的返回時點。',
            '戰鬥期間原版DS:0066改為4；欄位復用與正常返回次格未在本收據證實。',
            '其他戰鬥／住宿完整世界同狀態及父規格仍開放。'])

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--export',action='store_true')
    args = parser.parse_args()
    result = validate()
    if args.export:
        target = Path('/out/receipt.json')
        assert not target.exists()
        paths = [p for p in PRIVATE.rglob('*') if p.is_file()]
        paths += [ROOT/'docs/re/20261002-battle-world.md',ROOT/'docs/spec/011-wilderness-ambush.md',
                  ROOT/'CONTEXT.md',ROOT/'docs/re/RESEARCH-LOG.md',ROOT/'docs/worklist.json',ROOT/'WORKLIST.md',
                  HERE/'verify_world.py']
        paths += [p for folder in ['cmd','internal','assets','docs/i18n'] for p in (ROOT/folder).rglob('*') if p.is_file()]
        paths += [ROOT/'go.mod',ROOT/'go.sum']
        paths += [ROOT/name for name in read(PRIVATE/'parent-preservation.json')['sources']]
        assert all(p.stat().st_uid == p.stat().st_gid == 1000 for p in paths)
        result['attachments'] = {str(p.relative_to(ROOT)):sha(p) for p in sorted(set(paths))}
        target.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    else:
        expected = read(HERE/'receipt.json')
        for name,digest in expected['attachments'].items():
            assert sha(ROOT/name)==digest,name
        assert result == {k:v for k,v in expected.items() if k!='attachments'}
    print(json.dumps(result if not args.export else {k:v for k,v in result.items() if k!='attachments'},ensure_ascii=False))

if __name__=='__main__':
    main()
