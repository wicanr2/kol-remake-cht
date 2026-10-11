#!/usr/bin/env python3
"""Docker 內的中文收尾核查，範圍與限制見 docs/verification/cjk-closeout-20261011.md。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import zipfile

SRC, OUT = Path('/src'), Path(os.environ.get('KOL_CJK_OUT', '/out'))
BIN = Path(os.environ.get('KOL_CJK_BINARY', str(SRC / 'dist-all/v.1.0.0-20261011/full/kol')))
sys.path.insert(0, str(SRC / 'tools'))
import check_player_paths as paths


def sha(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def write(p, value):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def run(argv, log, timeout=180):
    with log.open('w') as f:
        try:
            p = subprocess.run(argv, cwd=SRC, stdout=f, stderr=subprocess.STDOUT, timeout=timeout)
        except subprocess.TimeoutExpired:
            f.write(f'\n驗證工具逾時 {timeout} 秒；不判為產品失敗。\n')
            return 124
    return p.returncode


def inventory():
    assert not (OUT / 'predeclared.json').exists()
    source = {str(p.relative_to(SRC)): sha(p) for n in ('cmd', 'internal')
              for p in sorted((SRC / n).rglob('*.go'))}
    source.update({n: sha(SRC / n) for n in ('go.mod', 'go.sum')})
    built = json.loads((BIN.parent.parent / 'build-inputs.json').read_text())
    assert source == built['program_inputs']
    plans = []
    for p in sorted((SRC / paths.PLANS).glob('*.json')):
        if p.stem.endswith('-receipt'):
            continue
        d = json.loads(p.read_text())
        plans.append(dict(name=p.stem, path=str(p.relative_to(SRC)), sha256=sha(p),
            seed=d.get('seed', 1), scene=d.get('scene', 'play'),
            keys=len(d['script'].split()), load=d.get('load'),
            input_available=not d.get('load') or (SRC / d['load']).is_file(),
            controlled_rolls=d.get('ambush_rolls'), checkpoints=d.get('checkpoints', [])))
    enums = {}
    for name in ('internal/flow/flow.go', 'internal/town/session.go',
                 'internal/world/session.go', 'internal/battle/planning.go'):
        p = SRC / name
        if p.exists():
            enums[name] = [dict(line=i, text=s.strip()) for i, s in
                enumerate(p.read_text().splitlines(), 1)
                if re.match(r'\s*(Scene|Mode|Stage|Prompt)[A-Z]\w*', s)]
    write(OUT / 'predeclared.json', dict(schema='kol-cjk-closeout/1',
        version='v.1.0.0-20261011', binary_sha256=sha(BIN), program_inputs=source,
        plans=plans, enums=enums, original_seed=1,
        original_seed_setup='既有正常入口工具：旅店招呼後、任何招募／野外亂數前，DS+0x1600/0x1602；規劃比較點另明示重新對齊。',
        display_difference_policy='僅 battle.exec_text 與 battle.encounter_text 顯示文字可以不同；來源 cmd/kol/main.go exportTraceStep。其他狀態差異全部列出。',
        fail='崩潰、缺譯、缺字、非顯示差異或未訪分支均不得列為全部中文通過',
        limits='展示／觀察輸入與正常新遊戲分列；缺資料不能跳過後稱全覆蓋；不覆寫已交付包。'))
    print('INVENTORY', len(source), 'program inputs;', len(plans), 'plans', flush=True)


def static():
    d = OUT / 'static'
    d.mkdir(exist_ok=True)
    cmds = [
        ['verify_runtime_translation.py', '--out', str(d / 'runtime.json')],
        ['verify_ui_translation.py', '--screen-width', '960', '--out', str(d / 'ui.json')],
        ['verify_display_names.py'],
        ['verify_world_prompts.py'],
        ['verify_cjk_coverage.py', '--report', str(d / 'glyphs.json')],
        ['verify_glossary_consistency.py', '--out', str(d / 'glossary.json')],
    ]
    results = []
    for cmd in cmds:
        log = d / (Path(cmd[0]).stem + '.log')
        code = run(['python3', '-B', str(SRC / 'tools' / cmd[0]), *cmd[1:]], log)
        results.append(dict(command=cmd, exit=code, log=str(log.relative_to(OUT))))
        print('STATIC', cmd[0], code, flush=True)
    code = run([str(BIN), '-dump-cjk-layout', str(d / 'layout.json')], d / 'layout.log')
    assert code == 0
    cmd = ['python3', '-B', str(SRC / 'tools/verify_text_fit.py'),
        '--layout', str(d / 'layout.json'), '--font', 'assets/ui-font/manifest.json',
        '--strings', 'docs/i18n/newchar-zh-Hant.json',
        '--bank-strings', 'docs/i18n/ui-zh-Hant.json',
        '--ui-i18n', 'docs/i18n/ui-zh-Hant.json',
        '--data', 'assets/original-derived-data/character-creation.json',
        '--town-i18n', 'docs/i18n/town-zh-Hant.json',
        '--display-names', 'docs/i18n/display-names-zh-Hant.tsv',
        '--out', str(d / 'text-fit.json')]
    code = run(cmd, d / 'text-fit.log')
    results.append(dict(command=cmd, exit=code, log='static/text-fit.log'))
    write(d / 'checks.json', results)
    print('STATIC TEXT FIT', code, flush=True)


def captures():
    jobs = []
    for line in (SRC / 'docs/screenshots/shots.spec').read_text().splitlines():
        if not line or line.startswith('#'):
            continue
        parts = line.split('|')
        name = Path(parts[0]).stem
        argv = parts[1:]
        # 輸出只寫本輪目錄；歷史規格一字不改。
        for flag in ('-shot', '-shot-dir', '-trace', '-save', '-text-log'):
            if flag in argv:
                i = argv.index(flag)
                argv[i + 1] = str(OUT / 'gallery' / (name + '.png'))
        if '-seed' not in argv:
            argv += ['-seed', '1']
        argv += ['-text-log', str(OUT / 'gallery' / (name + '.jsonl'))]
        jobs.append(dict(name=name, argv=argv, kind='展示規格；不證明正常入口可達'))
    # 逐個建築繪製門面與進門文字，確保七城101個已抽取建築都納入。
    buildings = json.loads((SRC / 'assets/original-derived-data/town-buildings.json').read_text())
    for town in buildings['towns']:
        for door in town['doors']:
            if not door.get('name'):
                continue
            name = f'zh-building-{town["town"]:02}-{door["index"]:02}'
            base = OUT / 'gallery' / name
            jobs.append(dict(name=name, kind='建築呈現；觀察入口', argv=[
                '-lang', 'zh-Hant', '-seed', '1', '-scene', 'building',
                '-building-town', str(town['town']), '-building-door', str(door['index']),
                '-shot', str(base.with_suffix('.png')), '-text-log', str(base.with_suffix('.jsonl'))]))
    (OUT / 'gallery').mkdir(exist_ok=True)
    write(OUT / 'capture-plan.json', jobs)
    results = []
    for i, job in enumerate(jobs):
        code = run([str(BIN), *job['argv']], OUT / 'gallery' / (job['name'] + '.log'))
        results.append(dict(**job, exit=code))
        if i % 20 == 0:
            print('CAPTURE', i + 1, '/', len(jobs), job['name'], code, flush=True)
    write(OUT / 'captures.json', results)
    print('CAPTURES DONE', len(results), 'failed', sum(r['exit'] != 0 for r in results), flush=True)


def differences(a, b, path=''):
    if type(a) != type(b):
        return [dict(path=path, en=a, zh=b)]
    if isinstance(a, dict):
        out = []
        for k in sorted(set(a) | set(b)):
            out += differences(a.get(k), b.get(k), path + '.' + k)
        return out
    if isinstance(a, list):
        out = []
        if len(a) != len(b):
            out.append(dict(path=path + '.length', en=len(a), zh=len(b)))
        for i, (x, y) in enumerate(zip(a, b)):
            out += differences(x, y, path + f'[{i}]')
        return out
    return [] if a == b else [dict(path=path, en=a, zh=b)]


def players():
    pre = json.loads((OUT / 'predeclared.json').read_text())
    report, saves = [], {}
    # 所有現存輸入都列入；缺少輸入原樣登記，絕不產生新「基準」冒充回歸。
    names, _ = paths.dependency_order([p['name'] for p in pre['plans']])
    selected = set(os.environ.get('KOL_CJK_CASES', '').split(',')) - {''}
    report_path = OUT / os.environ.get('KOL_CJK_REPORT', 'players.json')
    for name in names:
        if selected and name not in selected:
            continue
        plan = json.loads((SRC / paths.PLANS / (name + '.json')).read_text())
        loaded = plan.get('load')
        if loaded and not (SRC / loaded).is_file() and loaded not in saves:
            report.append(dict(name=name, status='MISSING_INPUT', path=loaded))
            print('PLAYER MISSING_INPUT', name, loaded, flush=True)
            continue
        runs = []
        for lang in ('en', 'zh-Hant'):
            d = OUT / 'players' / name / lang
            d.mkdir(parents=True, exist_ok=True)
            p = dict(plan)
            if loaded:
                original = saves[loaded][lang] if loaded in saves else SRC / loaded
                copied = d / 'input-save.json'
                copied.write_bytes(original.read_bytes())
                p['load'] = str(copied)
            # 保留原計畫所有關鍵檢查點，只截這些點，仍記每個狀態與文字呼叫。
            argv = paths.spec_line(p, d).split('|')[1:]
            if '-shot-steps' not in argv:
                argv += ['-shot-steps', ','.join(str(c['step']) for c in plan.get('checkpoints', [])) or '0']
            if '-seed' not in argv:
                argv += ['-seed', '1']
            argv += ['-lang', lang, '-text-log', str(d / 'text.jsonl')]
            write(d / 'command.json', dict(argv=argv, seed=plan.get('seed', 1),
                source_plan=pre['plans'][next(i for i,p in enumerate(pre['plans']) if p['name']==name)],
                input_save_sha256=sha(d / 'input-save.json') if loaded else None))
            code = run([str(BIN), *argv], d / 'engine.log', timeout=600)
            runs.append(dict(lang=lang, exit=code))
        result = dict(name=name, runs=runs, scene=plan.get('scene', 'play'), seed=plan.get('seed', 1),
            controlled_rolls=plan.get('ambush_rolls'), observation_input=bool(loaded))
        if any(r['exit'] for r in runs):
            result['status'] = 'EXECUTION_FAILED'
        else:
            base = OUT / 'players' / name
            en = json.loads((base / 'en/trace.json').read_text())['steps']
            zh = json.loads((base / 'zh-Hant/trace.json').read_text())['steps']
            diff = differences(en, zh)
            visible, other = [], []
            for row in diff:
                (visible if re.fullmatch(r'\[\d+\]\.battle\.(exec_text|encounter_text)(?:\[\d+\]|\.length)?', row['path']) else other).append(row)
            result.update(states=len(en), difference_count=len(diff), display_differences=visible,
                          unexpected_differences=other, status='PASS' if not other else 'STATE_DIFFERENCE')
            saved = plan.get('save', {}).get('file')
            if saved:
                saves[saved] = {lang: base / lang / 'save.json' for lang in ('en','zh-Hant')}
        report.append(result)
        write(report_path, report)
        print('PLAYER', name, result['status'], result.get('states'), flush=True)
    write(report_path, report)


def text_report():
    player_names = {p['name'] for p in json.loads((SRC / 'assets/original-derived-data/character-roster.json').read_text())['slots'] if p['name']}
    records = {}
    for p in sorted(OUT.rglob('*.jsonl')):
        for line in p.read_text().splitlines():
            row = json.loads(line)
            value = row['text']
            if re.search('[A-Za-z]', value):
                records.setdefault(value, set()).add(str(p.relative_to(OUT)))
    write(OUT / 'latin-text.json', dict(player_names=sorted(player_names),
        strings=[dict(text=t, logs=sorted(ps)) for t,ps in sorted(records.items())],
        limits='只掃字模繪製呼叫。玩家名、法術音節、單位、版權與圖片英文必須逐類人工判讀。'))
    print('LATIN TEXT', len(records), flush=True)


def candidate():
    """新程式只重驗受影響的告示，另以既有等價輸入重生有限原版比較。"""
    plan = json.loads((SRC / 'workplace/normal-six-battle-20261005/remake-plan.json').read_text())
    cases = [('normal-six', plan['script'], '0,45,59,171,172'),
             ('residence', (SRC / 'docs/verification/cjk-ui/natural-sign-20261001/script.txt').read_text().strip(),
              '18,58,60,72,159,2976,2977,3078,3079,3080')]
    results = []
    for name, script, shots in cases:
        for lang in ('en', 'zh-Hant'):
            d = OUT / 'candidate' / name / lang
            d.mkdir(parents=True, exist_ok=True)
            argv = [str(BIN), '-scene', 'flow', '-seed', '1', '-lang', lang,
                    '-flow-script', script, '-trace', str(d / 'trace.json'), '-trace-detail',
                    '-shot-dir', str(d), '-shot-steps', shots, '-text-log', str(d / 'text.jsonl')]
            write(d / 'command.json', dict(argv=argv, binary_sha256=sha(BIN),
                seed=1, method='啟動前固定；正常標題與按鍵；無存檔／座標注入／亂數覆寫'))
            code = run(argv, d / 'engine.log', timeout=900)
            results.append(dict(name=name, language=lang, exit=code))
            write(OUT / 'candidate-runs.json', results)
            print('CANDIDATE', name, lang, code, flush=True)


def sign():
    """觀察入口只驗告示繪製；不能稱正常新遊戲可達。"""
    sample = json.loads((SRC / 'docs/re/town-outdoor-sample.json').read_text())
    sample.update(town=5, camera=dict(x=40, y=19), at_door=True, at_edge=False,
                  entry_point=0, note='本輪明示的告示繪圖觀察輸入；不是正常新遊戲證據')
    write(OUT / 'sign-input.json', sample)
    for name, binary in [('before', SRC / 'dist-all/v.1.0.0-20261011/full/kol'), ('after', BIN)]:
        for lang in ('en', 'zh-Hant'):
            d = OUT / 'sign' / name / lang
            d.mkdir(parents=True, exist_ok=True)
            argv = [str(binary), '-scene', 'outdoor', '-lang', lang, '-seed', '1',
                    '-outdoor', str(OUT / 'sign-input.json'), '-shot', str(d / 'door.png'),
                    '-text-log', str(d / 'text.jsonl')]
            write(d / 'command.json', dict(argv=argv, binary_sha256=sha(binary),
                  classification='觀察入口，未宣稱正常新遊戲路徑'))
            code = run(argv, d / 'engine.log')
            assert code == 0
    print('SIGN four captures completed', flush=True)


def report():
    """只整理已執行的結果，不重播、不補造缺失的輸入。"""
    from PIL import Image
    import shutil

    allowed = r'\[\d+\]\.battle\.(exec_text|encounter_text)(?:\[\d+\]|\.length)?'
    rows = {}
    for name in ('players.json', 'players-remaining.json'):
        for row in json.loads((OUT / name).read_text()):
            rows[row['name']] = row
    for row in rows.values():
        if row.get('status') == 'STATE_DIFFERENCE':
            all_diff = row['display_differences'] + row['unexpected_differences']
            row['display_differences'] = [x for x in all_diff if re.fullmatch(allowed, x['path'])]
            row['unexpected_differences'] = [x for x in all_diff if not re.fullmatch(allowed, x['path'])]
            row['status'] = 'PASS' if not row['unexpected_differences'] else 'STATE_DIFFERENCE'
        if row.get('status') == 'PASS':
            row['status'] = 'LANGUAGE_MATCH'
    rows['quest-chain'] = dict(name='quest-chain', status='INCOMPLETE',
        reason='本批 zh-Hant 程序達600秒工具上限；兩側均沒有完整trace或save，不判產品通過或失敗。')
    text_report()
    def subset(expected, actual, path=''):
        found = []
        if isinstance(expected, dict):
            for k, value in expected.items():
                if k in ('step','note'):
                    continue
                if not isinstance(actual,dict) or k not in actual:
                    found.append(dict(path=path+'.'+k,status='UNOBSERVED'))
                else:
                    found += subset(value,actual[k],path+'.'+k)
        elif expected != actual:
            found.append(dict(path=path,status='DIFFERENCE',expected=expected,actual=actual))
        return found
    milestone_audit = []
    for d in sorted((OUT/'players').iterdir()):
        f=d/'en/trace.json'
        if not f.exists():
            continue
        plan=json.loads((SRC/paths.PLANS/(d.name+'.json')).read_text())
        trace=json.loads(f.read_text())['steps']
        checks=[]
        for m in plan.get('milestones',[]):
            i=m['step']
            checks += subset(m,trace[i],f'step{i}') if i<len(trace) else [dict(status='MISSING_STEP',step=i)]
        milestone_audit.append(dict(name=d.name,milestones=len(plan.get('milestones',[])),
                                   checks=checks,last_scene=trace[-1].get('scene')))
    write(OUT/'milestone-audit.json',milestone_audit)
    write(OUT / 'player-summary.json', dict(cases=list(rows.values()), milestone_audit=milestone_audit,
        counts={s: sum(x['status'] == s for x in rows.values()) for s in sorted({x['status'] for x in rows.values()})},
        complete_coverage=False, display_policy=allowed,
        limits='缺少歷史存檔14條，長任務1條未完成。競技場舊里程碑76個已觀測欄位不同；英中一致不代表該功能驗收通過。'))

    en = json.loads((OUT / 'contents-normal-en/trace.json').read_text())['steps']
    zh = json.loads((OUT / 'contents-normal-zh/trace.json').read_text())['steps']
    pixel_diff = lambda a,b: sum(x != y for x,y in zip(Image.open(a).convert('RGB').getdata(), Image.open(b).convert('RGB').getdata()))
    contents = dict(states_per_language=len(en), state_differences=differences(en, zh),
        english_pixel_differences=pixel_diff(OUT / 'gallery/en-book-contents.png', OUT / 'contents-normal-en/step-002.png'),
        return_page_pixel_differences=pixel_diff(OUT / 'contents-normal-zh/step-002.png', OUT / 'contents-normal-zh/step-004.png'),
        safe_rectangle=[528,48,354,402], labels=['目錄','製作名單','前言','獻詞','安裝新區域','角色選項','開始遊戲','播放開場動畫'],
        longest_label_bound=6*16, number_right=882,
        limits='正常靜態頁及返回；翻書動畫另列，未宣稱動畫文字全中文。')
    write(OUT / 'contents-comparison.json', contents)
    plan=json.loads((SRC/'workplace/normal-six-battle-20261005/remake-plan.json').read_text())
    original=json.loads((OUT/'fresh-pinned/normal-six-private.json').read_text())
    previous=json.loads((SRC/'workplace/normal-six-battle-20261005/corrected-original/normal-six-private.json').read_text())
    write(OUT/'fresh-original-comparison.json',dict(exact=original==previous,
          differences=differences(original,previous),states=len(original['rows'])))
    normal_en=json.loads((OUT/'candidate/normal-six/en/trace.json').read_text())['steps']
    normal_zh=json.loads((OUT/'candidate/normal-six/zh-Hant/trace.json').read_text())['steps']
    world_differences=[]; samples=0
    for m in plan['mapping']:
        i,j=m['original_row'],m['remake_step']
        if i<54:
            continue
        a,b=original['rows'][i]['state'],normal_en[j]['world'];samples+=1
        x=a['camera']+a['marker']+[a['time']]
        y=[b['x'],b['y'],b['sub_x'],b['sub_y'],b['ticks']]
        if x!=y:
            world_differences.append(dict(original_row=i,remake_step=j,original=x,remake=y))
    initial=json.loads((OUT/'fresh-pinned/first-battle-private.json').read_text())['state']
    entry=normal_en[-1]['battle']; actors=[]
    for line in entry['roster']:
        m=re.fullmatch(r'(\d+) (.+) (party|foe) (\d+),(\d+) head=(\d+) chest=(\d+)',line)
        assert m
        raw=bytes.fromhex(initial['actors'][int(m[1])])
        actors.append(dict(slot=int(m[1]),actual=[int(m[k]) for k in (4,5,6,7)],
                           expected=[raw[31],raw[32],raw[17],raw[19]]))
    crop=(152,8,296,152)
    a=Image.open(OUT/'fresh-pinned/normal-169.png').convert('RGB').crop(crop)
    b=Image.open(OUT/'candidate/normal-six/en/step-172.png').convert('RGB').crop(crop)
    diff=differences(normal_en,normal_zh)
    write(OUT/'normal-six-comparison.json',dict(original_states=len(original['rows']),
          world_samples=samples,world_differences=world_differences,initial_actors=actors,
          actor_differences=[a for a in actors if a['actual']!=a['expected']],
          cursor=[entry['cursor'],initial['actor']],round=[entry['round'],initial['round']],
          bar_matches=entry['bar']==list(bytes.fromhex(initial['icons_hex'])),
          viewport_pixel_differences=sum(x!=y for x,y in zip(a.getdata(),b.getdata())),
          language_states=len(normal_en),display_difference_count=sum(bool(re.fullmatch(allowed,x['path'])) for x in diff),
          unexpected_differences=[x for x in diff if not re.fullmatch(allowed,x['path'])],complete_battle_parity=False))
    program = {str(p.relative_to(SRC)): sha(p) for n in ('cmd','internal') for p in sorted((SRC/n).rglob('*.go'))}
    program.update({n: sha(SRC/n) for n in ('go.mod','go.sum')})
    write(OUT / 'final-program-inputs.json', dict(program_inputs=program,
        binary_sha256=sha(OUT/'kol-cjk-final'), version='dev-cjk-20261011',
        release_unchanged=sha(SRC/'dist-all/v.1.0.0-20261011/full/kol') == json.loads((OUT/'predeclared.json').read_text())['binary_sha256']))

    dest = SRC / 'docs/verification/cjk-closeout-20261011'
    dest.mkdir(exist_ok=True)
    assert dest.stat().st_uid == os.getuid()
    shots = [
        ('zh-contents.png', OUT/'contents-normal-zh/step-002.png', '正常入口／目錄 step2', 'kol-cjk-final'),
        ('zh-newchar.png', OUT/'gallery/zh-newchar-picture.png', '展示入口／建角', 'v.1.0.0-20261011'),
        ('zh-party.png', OUT/'candidate/normal-six/zh-Hant/step-045.png', '正常入口／六人 step45', 'kol-cjk'),
        ('zh-battle.png', OUT/'candidate/normal-six/zh-Hant/step-172.png', '正常入口／戰鬥 step172', 'kol-cjk'),
        ('en-battle.png', OUT/'candidate/normal-six/en/step-172.png', '正常入口／戰鬥 step172', 'kol-cjk'),
        ('original-battle.png', OUT/'fresh-pinned/normal-169.png', 'dosgolem 正常入口／normal169', 'dosgolem 968ca774705853276ae55c8d1163c369481ac71f'),
        ('zh-residence.png', OUT/'sign/after/zh-Hant/door.png', '觀察入口／告示', 'kol-cjk'),
        ('approved-sign-A.png', SRC/'workplace/cjk-residence-sign-prototype-20261006/A.png', '使用者核准原型 A；歷史正常畫格衍生', 'RE-271'),
    ]
    records = []
    for name, source, frame, producer in shots:
        with Image.open(source) as im:
            im.verify()
        with Image.open(source) as im:
            dimensions = list(im.size)
        shutil.copyfile(source, dest/name)
        records.append(dict(file=name, sha256=sha(dest/name), dimensions=dimensions,
            source_path=str(source), source_sha256=sha(source), frame=frame, producer=producer,
            evidence='已核對格式、雜湊與目視；正常／觀察範圍依 frame 分類',
            rights='private-repo 可保存；不得公開散布'))
    source_manifests = json.loads((SRC/'workplace/hd-art-ui-20261004/inventory.json').read_text())['source_manifests']
    for path, expected in source_manifests.items():
        assert sha(SRC/path) == expected, path
    with zipfile.ZipFile(SRC/'Knights-of-legend.zip') as z:
        entries = {n: hashlib.sha256(z.read(n)).hexdigest() for n in z.namelist() if not n.endswith('/')}
    write(dest/'manifest.json', dict(schema='kol-cjk-screenshots/1', tool='tools/verify_cjk_closeout.py report',
        tool_sha256=sha(SRC/'tools/verify_cjk_closeout.py'),
        source_zip=dict(name='Knights-of-legend.zip',sha256=sha(SRC/'Knights-of-legend.zip')),
        source_entries=entries, asset_manifests=source_manifests,
        docker_image='kol-package:1.26.7-ebiten2.9.9-20261011-r1',
        docker_image_id='0b933af1238ab5839ad7a64b6ae05f069f57cb85da5917dfbe59730f6d7ba53f',
        go='1.26.7', ebiten='2.9.9', python='3.11', seed=1,
        commands='完整命令、trace、原版 seed 設定與輸入保存在 ../local/cjk-closeout-20261011/evidence.tar.xz',
        files=records, all_chinese=False,
        limits=['14條歷史計畫缺輸入','長任務工具逾時','翻書動畫、其他圖片文字與未訪動態分支未全覆蓋',
                '原作標誌、角色名、法術音節及按鍵名不以英文字母出現即判缺譯']))
    layout=SRC/'workplace/hd-art-ui-20261004/layout-review.html'
    write(dest/'hd-decision.json',dict(date='2026-10-11',authority='使用者明確回覆',
          answer='A：1920×1200，保持原比例（建議）',style='手繪，沿既有已確認決定',
          resolution=[1920,1200],aspect='16:10',logical_layout='保持原比例',
          display_1080p=dict(content=[1728,1080],side_bars=96),
          source_prototype=str(layout.relative_to(SRC)),source_prototype_sha256=sha(layout),
          scope='只確認媒材與版面方向；造型、構圖、素材契約與正式整合未核准'))
    write(OUT/'receipt.json', dict(schema='kol-cjk-closeout-receipt/1', date='2026-10-11', status='有限核查完成，非全畫面中文驗收',
        artifacts={str(p.relative_to(OUT)): dict(size=p.stat().st_size,sha256=sha(p))
                   for p in sorted(OUT.rglob('*')) if p.is_file() and p.name!='receipt.json'},
        complete_parity=False, complete_chinese=False))
    print('REPORT', len(rows), 'plans;', len(records), 'verified PNGs', flush=True)


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('phase', choices=('inventory', 'static', 'captures', 'players', 'text_report', 'candidate', 'sign', 'report'))
    args = p.parse_args()
    assert Path('/.dockerenv').exists() and SRC.is_dir() and OUT.is_dir()
    assert (OUT.stat().st_uid, OUT.stat().st_gid) == (os.getuid(), os.getgid())
    globals()[args.phase]()
