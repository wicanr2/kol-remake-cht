#!/usr/bin/env python3
"""在 Docker 內重生 SPEC-040 場景；原始輸入唯讀，輸出僅限本輪工作區。"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

ROOT = Path('/src')
OUT = Path('/out')


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def write(p, value):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def run(argv, name, seconds=120):
    with (OUT / (name + '.log')).open('w') as f:
        result = subprocess.run(list(map(str, argv)), cwd=ROOT, stdout=f,
                                stderr=subprocess.STDOUT, timeout=seconds)
    if result.returncode:
        raise RuntimeError(f'{name}: {result.returncode}; see log')


def prepare():
    (OUT / 'inputs').mkdir(exist_ok=True)
    imports = {
        'party6': (['Poontang', 'Moraine', 'Fighting Bull', 'Kenneth', 'Sharakan', 'Olear'], '7,3,57,8'),
        'shop': (['ddduhh'], '1,0,24,52'),
        'train': (['ddduhh'], '1,0,53,52'),
        'stable': (['Poontang'], '1,0,41,55'),
        'hostel': (['Poontang'], '7,7,59,29'),
        'quest-take': (['Poontang', 'Moraine', 'Fighting Bull', 'Kenneth', 'Sharakan', 'Olear'], '7,4,9,24'),
        'final': (['Poontang'], '7,3,57,8'),
    }
    records = []
    for name, (names, location) in imports.items():
        dest = OUT / 'inputs' / (name + '.json')
        argv = ['python3', '-B', ROOT / 'tools/import_original_characters.py',
                '--zip', ROOT / 'Knights-of-legend.zip', '--repo', ROOT,
                '--names', *names, '--location', location, '--out', dest,
                '--saved-at', '2026-10-11T00:00:00Z']
        run(argv, 'prepare-' + name)
        original = json.loads(dest.read_text())
        data = copy.deepcopy(original)
        if name in ('stable', 'hostel'):
            data['party'][0]['gold'] = 3000
        if name == 'final':
            member = data['party'][0]
            member['quests'] = [0, 1, 0, 0, 0, 0]
            member['progress'] = [23, 0, 0, 0, 0]
        write(dest, data)
        records.append(dict(name=name, command=list(map(str, argv)),
            changes='gold=3000' if name in ('stable', 'hostel') else
                    'final quest accepted, reported=23' if name == 'final' else 'location only',
            imported_sha256=hashlib.sha256(json.dumps(original, sort_keys=True).encode()).hexdigest(),
            output_sha256=sha(dest)))
    quest_list = json.loads((OUT / 'inputs/quest-take.json').read_text())
    quest_list['party'][0]['quests'] = [0, 3, 0, 0, 0, 0]
    quest_list['party'][0]['progress'] = [5, 0, 0, 0, 0]
    write(OUT / 'inputs/quest-list.json', quest_list)
    sites_path = ROOT / 'assets/original-derived-data/quest-sites.json'
    sites = json.loads(sites_path.read_text())
    site = next(s for s in sites['sites'] if s['index'] == 4)
    site['count'] = 1
    site['monster_x'] = [site['party_x'][0] - 1]
    site['monster_y'] = [site['party_y'][0]]
    sites['note_observation_only'] = 'SPEC-040：最終任務一隻同等級獨眼巨人，相鄰站位；非原版對拍資料。'
    write(OUT / 'inputs/sites-final.json', sites)
    write(OUT / 'predeclared.json', dict(schema='kol-controlled-scenes/1', seed=1,
        spec='docs/spec/040-controlled-scene-acceptance.md', lock_hp=True,
        party_damage_multiplier=20, final_world_start=[48, 9], original_parity_claimed=False,
        source_zip_sha256=sha(ROOT / 'Knights-of-legend.zip'),
        source_sites_sha256=sha(sites_path), imports=records,
        substituted_sites_sha256=sha(OUT / 'inputs/sites-final.json'),
        program_inputs={str(p.relative_to(ROOT)): sha(p) for sub in ('cmd', 'internal')
                        for p in sorted((ROOT / sub).rglob('*.go'))},
        targets=['book/creation', 'town/services', 'world/quests',
                 'encounter', 'planning', 'execution', 'victory', 'spoils', 'ending panels 0..6']))
    wrapper = OUT / 'kol-controlled'
    wrapper.write_text('#!/bin/sh\nexec /out/kol-sceneverify -verify-scenes -verify-lock-hp '
                       '-verify-attack 20 -verify-world-start 48,9 "$@"\n')
    wrapper.chmod(0o755)
    (OUT / 'initial.txt').write_text('RET SPACE\n')
    print('PREPARED', len(imports), 'fresh imports', flush=True)


def captures():
    jobs = []
    for name, fixture in [('town-shop', 'shop'), ('town-train', 'train'),
                          ('town-stable', 'stable'), ('town-scroll', 'hostel'),
                          ('quest-take', 'quest-take'), ('quest-list', 'quest-list')]:
        plan = json.loads((ROOT / 'docs/verification/player-path' / (name + '.json')).read_text())
        jobs.append(dict(name=name, fixture=fixture, scene='play', script=plan['script'], controlled=True))
    for name in ('full', 'town-mirror', 'town-recruit', 'town-round-trip'):
        plan = json.loads((ROOT / 'docs/verification/player-path' / (name + '.json')).read_text())
        jobs.append(dict(name=name, scene=plan.get('scene', 'play'), script=plan['script'], controlled=False))
    write(OUT / 'capture-plan.json', jobs)
    results = []
    for job in jobs:
        dest = OUT / 'scenes' / job['name']
        dest.mkdir(parents=True, exist_ok=True)
        argv = [OUT / ('kol-sceneverify' if job['controlled'] else 'kol-normal'),
                '-lang', 'zh-Hant', '-seed', '1', '-scene', job['scene'],
                '-flow-script' if job['scene'] == 'flow' else '-play-script', job['script'],
                '-shot-dir', dest, '-trace', dest / 'trace.json', '-text-log', dest / 'text.jsonl']
        if job['controlled']:
            argv += ['-verify-scenes', '-verify-lock-hp', '-verify-attack', '20',
                     '-load', OUT / 'inputs' / (job['fixture'] + '.json')]
        run(argv, 'capture-' + job['name'])
        trace = json.loads((dest / 'trace.json').read_text())['steps']
        results.append(dict(name=job['name'], states=len(trace), command=list(map(str, argv)),
                            final=trace[-1], screenshots=len(list(dest.glob('*.png')))))
        print('CAPTURE', job['name'], len(trace), flush=True)
    write(OUT / 'capture-results.json', results)


def battle():
    env = dict(os.environ, KOL_BIN='/out/kol-controlled', KOL_SCENE='play',
               KOL_SCRIPT_FLAG='-play-script', KOL_QUEST_SITES='/out/inputs/sites-final.json',
               KOL_DRIVE_DIR='/out/driver', KOL_STOP_AFTER_BATTLE='1')
    with (OUT / 'battle-driver.log').open('w') as f:
        subprocess.run(['python3', '-B', ROOT / 'tools/drive_battle_script.py', '1',
                        '/out/inputs/final.json', '/out/initial.txt', '/out/battle-script.txt', '110'],
                       cwd=ROOT, env=env, stdout=f, stderr=subprocess.STDOUT, check=True, timeout=480)
    script = (OUT / 'battle-script.txt').read_text().strip()
    for language in ('en', 'zh-Hant'):
        dest = OUT / 'battle' / language
        dest.mkdir(parents=True, exist_ok=True)
        argv = [OUT / 'kol-controlled', '-lang', language, '-scene', 'play', '-seed', '1',
                '-load', OUT / 'inputs/final.json', '-quest-sites', OUT / 'inputs/sites-final.json',
                '-play-script', script,
                '-shot-dir', dest, '-trace', dest / 'trace.json', '-text-log', dest / 'text.jsonl']
        run(argv, 'battle-' + language, 180)
        steps = json.loads((dest / 'trace.json').read_text())['steps']
        panels = sorted({s['ending']['panel'] for s in steps if s.get('scene') == 'ending'
                         and 0 <= s['ending']['panel'] < 7})
        verdicts = sorted({s['battle']['verdict'] for s in steps if s.get('battle')})
        write(dest / 'summary.json', dict(states=len(steps), ending_panels=panels, verdicts=verdicts,
                                         command=list(map(str, argv))))
        assert 2 in verdicts, '未到勝利，不能以腳本退出成功判通過'
        assert set(range(7)) <= set(panels), '結局七段尚未全部抵達'
        print('BATTLE', language, len(steps), 'ending', panels, flush=True)


def guards():
    base = ['-verify-scenes', '-scene', 'play', '-play-script', 'SPACE',
            '-trace', '/out/forbidden-trace.json', '-shot-dir', '/out/forbidden-shots']
    cases = [
        ('formal-rejects-verification', 'kol-normal', ['-verify-scenes']),
        ('helper-needs-mode', 'kol-sceneverify', ['-verify-lock-hp']),
        ('verification-rejects-save', 'kol-sceneverify', base + ['-save', '/out/forbidden-save.json']),
        ('verification-rejects-range', 'kol-sceneverify', base + ['-verify-attack', '101']),
    ]
    results = []
    for name, binary, args in cases:
        argv = [str(OUT / binary), *args]
        result = subprocess.run(argv, cwd=ROOT, capture_output=True, text=True, timeout=15)
        assert result.returncode == 2, (name, result.returncode, result.stderr)
        results.append(dict(name=name, command=argv, exit_code=2, stderr=result.stderr, passed=True))
    assert not any(OUT.glob('forbidden-*')), '拒絕參數後不可寫入驗收或正式輸出'
    write(OUT / 'cli-guards.json', results)
    print('GUARDS', len(results), 'PASS', flush=True)


def seal():
    from PIL import Image
    from verify_cjk_closeout import differences
    public = ROOT / 'docs/verification/controlled-scenes-20261011'
    hd = ROOT / 'docs/verification/hd-first-slice-20261011'
    public.mkdir(exist_ok=True)
    a = json.loads((OUT / 'battle/en/trace.json').read_text())['steps']
    b = json.loads((OUT / 'battle/zh-Hant/trace.json').read_text())['steps']
    def strip(value):
        if isinstance(value, dict):
            return {k: strip(v) for k, v in value.items() if k not in ('exec_text', 'encounter_text')}
        if isinstance(value, list):
            return [strip(v) for v in value]
        return value
    diff = differences(strip(a), strip(b))
    assert not diff, diff[:5]
    hp = []
    for row in b:
        for actor in (row.get('battle') or {}).get('roster', []):
            if ' party ' in actor:
                hp.append(re.search(r'head=(\d+) chest=(\d+)', actor).groups())
    assert hp and len(set(hp)) == 1
    assert any((s.get('battle') or {}).get('stage') == 4 for s in b), '缺戰利品'
    ending = [s for s in b if s['scene'] == 'ending' and s['ending']['panel'] < 7]
    assert {s['ending']['panel'] for s in ending} == set(range(7))
    english_unchanged = []
    for p in sorted((OUT / 'battle/en').glob('*.png')):
        old = OUT / 'battle-before-ending-fix/en' / p.name
        assert old.exists() and sha(old) == sha(p), p
        english_unchanged.append(p.name)
    assert not list((OUT / 'inputs').glob('*.v2.json')), '驗收不可產生正式保存'
    selected = [('shop', 'scenes/town-shop/step-010.png'),
                ('train', 'scenes/town-train/step-012.png'),
                ('stable', 'scenes/town-stable/step-008.png'),
                ('spells', 'scenes/town-scroll/step-020.png'),
                ('quest-take', 'scenes/quest-take/step-013.png'),
                ('quest-list', 'scenes/quest-list/step-010.png')]
    for name, stage in [('world-quest', None), ('encounter', 1), ('planning', 2), ('spoils', 4)]:
        step = 0 if stage is None else next(s['step'] for s in b if (s.get('battle') or {}).get('stage') == stage)
        selected.append((name, f'battle/zh-Hant/step-{step:03d}.png'))
    for row in ending:
        selected.append((f'ending-{row["ending"]["panel"]}', f'battle/zh-Hant/step-{row["step"]:03d}.png'))
    pictures = []
    for name, relative in selected:
        src, dst = OUT / relative, public / (name + '.png')
        assert Image.open(src).size == (960, 600)
        shutil.copy2(src, dst)
        pictures.append(dict(file=dst.name, sha256=sha(dst), frame=relative,
                             tool='kol sceneverify / SPEC-040', reviewed=True,
                             rights='private GitHub 可保存；無公開散布權',
                             source_entry='Knights-of-legend.zip 原版角色匯入＋已驗證 PNG 素材',
                             source_sha256=sha(ROOT / 'Knights-of-legend.zip')))
    hd_checks = []
    for scene in ('inn', 'battle'):
        dom = (OUT / f'hd-{"browser" if scene == "inn" else "battle"}-dom.html').read_text()
        assert 'data-tests="PASS"' in dom and 'data-check-count="44"' in dom
        src, dst = OUT / f'hd-{scene}-preview.png', hd / f'{scene}-preview.png'
        assert Image.open(src).size == (1920, 1200)
        shutil.copy2(src, dst)
        hd_checks.append(dict(scene=scene, checks=44, passed=True, screenshot=dst.name,
                             screenshot_sha256=sha(dst)))
    referenced = [ROOT / p for p in (
        'docs/screenshots/zh-town-door.png', 'docs/verification/cjk-closeout-20261011/zh-party.png',
        'docs/verification/cjk-closeout-20261011/zh-battle.png',
        'workplace/hd-art-ui-20261004/painted-direction.png')]
    write(hd / 'manifest.json', dict(schema='kol-hd-prototype/1', resolution=[1920, 1200],
        stage='disposable-prototype', rights='private GitHub 可保存；原版參考衍生，無公開散布權',
        source_zip_sha256=sha(ROOT / 'Knights-of-legend.zip'), tool='內建 image_gen.imagegen；canvas 原型；Chromium 154.0.8037.92',
        prompts='prompts.json', sources={str(p.relative_to(ROOT)): sha(p) for p in referenced},
        assets=[dict(file=p.name, sha256=sha(p), size=list(Image.open(p).size),
                     source_entry='prompts.json 對應生成輸出' if '-v1' in p.stem else '既有核准風格參考或本輪原型畫格',
                     evidence='prototype only; not final art approval') for p in sorted(hd.glob('*.png'))],
        browser_checks=hd_checks, review_html_sha256=sha(hd / 'review.html'),
        limits=['門面1774×887尚未達目標素材尺寸', '首位人物手臂及盾牌有生成偏差',
                '外框、NPC、圖示、地圖及傷勢人形尚未HD重繪', '原型不執行遊戲規則，不改v3']))
    inputs = {str(p.relative_to(ROOT)): sha(p) for sub in ('cmd', 'internal')
              for p in sorted((ROOT / sub).rglob('*.go')) if not p.name.endswith('_test.go')}
    inputs.update({n: sha(ROOT / n) for n in ('go.mod', 'go.sum', 'docs/i18n/ui-zh-Hant.json')})
    guard_results = json.loads((OUT / 'cli-guards.json').read_text())
    assert len(guard_results) == 4 and all(row['passed'] for row in guard_results)
    receipt = dict(schema='kol-controlled-scenes/1', date='2026-10-11',
        controlled_scene_acceptance=True, scope='SPEC-040 指定場景與中文畫面', original_parity_claimed=False,
        seed=1, lock_hp=True, party_damage_multiplier=20,
        states_per_language=len(a), non_display_differences=diff, party_head_chest_values=hp,
        english_render_unchanged_frames=len(english_unchanged), ending_panels=list(range(7)),
        capture_results=json.loads((OUT / 'capture-results.json').read_text()),
        reused_fixed_scenes='docs/verification/cjk-closeout-20261011.md：148張展示／建築及正常初次規劃',
        reviewed_pngs=pictures, program_inputs=inputs,
        verification_tool_sha256=sha(Path(__file__)),
        binaries={n: sha(OUT / n) for n in ('kol-normal', 'kol-sceneverify')},
        tests=['go test ./cmd/... ./internal/...', 'go test -tags sceneverify ./cmd/kol ./internal/battle',
               'go test ./internal/ui -run TestCJKEndingTextOutsidePictures -count=1'],
        cli_guard_checks=[{key: row[key] for key in ('name', 'exit_code', 'passed')} for row in guard_results],
        hd_browser_checks=hd_checks, formal_save_written=False,
        limits=['姓名、玩家輸入及精靈語法術名依現行契約保留', '終端結局完成後空白畫格不列場景通過',
                '不宣稱所有亂數分支、正常全旅程或原版整場parity', '既有v1封包未重包'])
    write(public / 'receipt.json', receipt)
    write(public / 'manifest.json', dict(schema='kol-scene-png/1', entries=pictures,
        source_asset_manifest='docs/verification/cjk-closeout-20261011/manifest.json',
        source_asset_manifest_sha256=sha(ROOT / 'docs/verification/cjk-closeout-20261011/manifest.json')))
    write(OUT / 'bilingual-comparison.json', dict(states_per_language=len(a), differences=diff,
        english_unchanged=english_unchanged, hp=hp))
    receipt['attachments'] = {str(p.relative_to(OUT)): dict(size=p.stat().st_size, sha256=sha(p))
                              for p in sorted(OUT.rglob('*')) if p.is_file() and p.name != 'receipt.json'}
    write(OUT / 'receipt.json', receipt)
    print('SEALED', len(pictures), 'reviewed screenshots;', len(a), 'states/language; HD 44 checks/scene', flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('phase', choices=['prepare', 'captures', 'battle', 'guards', 'seal'])
    globals()[p.parse_args().phase]()
