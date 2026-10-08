"""RE-187正常世界時間接續，封存後全唯讀重驗。"""
from pathlib import Path
import argparse
import copy
import hashlib
import json
import runpy
import sys

ROOT = Path('/src')
PRIVATE = ROOT / 'workplace/world-clock-20261002'
HERE = ROOT / 'docs/verification/world-clock-20261002'
sys.path.insert(0, str(ROOT / 'tools'))
from png_io import read_png


def read(path):
    return json.loads(path.read_text())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate():
    historical = runpy.run_path(str(PRIVATE / 'history_check.py'))['validate']()
    paths = dict(victory='workplace/battle-victory-20261002/source4',
                 defeat='workplace/battle-exits-20261002/flee3',
                 flee='workplace/battle-exits-20261002/flee4')
    native = {}
    for case, baseline in paths.items():
        data = read(PRIVATE / case / 'complete-source-private.json')
        previous = read(ROOT / baseline / 'complete-source-private.json')
        observed = data.pop('return_writes') or []
        data.pop('return_write_method')
        assert data == previous, case
        assert (PRIVATE / case / 'world-return-private.json').read_bytes() == (ROOT / baseline / 'world-return-private.json').read_bytes()
        assert len(observed) == dict(victory=57, defeat=120, flee=0)[case]
        world = read(PRIVATE / case / 'world-return-private.json')
        after_entry = next(i for i, r in enumerate(world['events']) if r['ida'] == '0x1E3A8')
        event = next(r for r in world['events'][after_entry+1:] if r['ida'] == '0x29800')
        before_cost = [r for r in observed if r['step'] < event['state']['steps']]
        if case != 'flee':
            last = before_cost[-1]
            assert last['ip_ida'] == dict(victory='0x20BAD', defeat='0x21821')[case]
            assert last['new'] == dict(victory=1, defeat=16)[case]
        native[case] = dict(changed_byte_writes=len(observed), returned_index=event['state']['raw_bb84'],
                            complete_source_unchanged=True)
    assert sha(PRIVATE / 'kol-world-time-linux') == 'dea837bc052786c3fa90b2672dd462e81ca768d5c44202f6b2333d12164c12b9'
    assert sha(PRIVATE / 'kol-settlement-probe2-linux') == 'b4b5f65b1a7dd4770a3996764ea086b58b7a0002c6932a113ac7237b9e99f7b5'
    assert (PRIVATE / 'kol-world-time-linux').stat().st_mode & 0o777 == 0o755
    assert 'FAIL' not in (PRIVATE / 'time-rules.log').read_text()
    assert (PRIVATE / 'time-build.log').read_text() == ''
    manifest = read(PRIVATE / 'probe2-source.json')
    assert manifest['parent_sha256'] == sha(ROOT / 'cmd/kol/main.go')
    assert manifest['probe_sha256'] == sha(PRIVATE / 'probe2-main-private.go')
    assert 'KOL_PROBE_PLAN_SEED' not in (ROOT / 'cmd/kol/main.go').read_text()
    source = read(ROOT / 'workplace/battle-return-20261002/source1/world-return-private.json')
    weather = source['rolls'][-1]
    assert weather['n'] == 128 and weather['caller_ida'] == '0x29A92'
    native_after_weather = next(r['state']['seed'] for r in reversed(source['events']) if r['ida'] == '0x29981')
    gui = {}
    english_weather_changes = {}
    layout = (ROOT / 'internal/world/cjk.go').read_text()
    assert 'WeatherPicX: 46, WeatherPicY: 247,' in layout
    for lang, scale in [('en', 1), ('zh-Hant', 3)]:
        current = PRIVATE / ('gui-time-' + lang)
        baseline = ROOT / 'workplace/battle-exits-20261002' / ('gui-center-' + lang)
        now, old = read(current / 'trace.json')['steps'], read(baseline / 'trace.json')['steps']
        assert len(now) == len(old) == 493
        count = 0
        for row, prior in zip(now, old):
            assert row.get('battle') == prior.get('battle')
            comparable = copy.deepcopy(row)
            if 'world' in comparable:
                w = comparable['world']
                for name in list(w):
                    if name.startswith('probe_'):
                        del w[name]
                if 'battle' in row:
                    assert w['ticks'] == 3360 and prior['world']['ticks'] == 3320
                    assert 'probe_world_roll_count' not in row['world']
                    w['ticks'] = 3320
                    count += 1
            assert comparable == prior, row['step']
        assert count == 362
        w = now[480]['world']
        assert w['probe_world_roll_count'] == 1491
        assert w['probe_world_last_roll'] == dict(N=128, Before=weather['state']['seed'],
                                               After=native_after_weather, Round=61, Stage=5)
        assert w['probe_world_seed'] == native_after_weather
        assert [w[k] for k in ['x', 'y', 'sub_x', 'sub_y', 'ticks']] == [49, 23, 4, 4, 3320]
        pictures = sorted(current.glob('*.png'))
        assert len(pictures) == 493
        changed = []
        for p in pictures:
            q = baseline / p.name
            if p.read_bytes() == q.read_bytes():
                continue
            index = int(p.stem.split('-')[1])
            assert 480 <= index <= 492
            a, b = read_png(p), read_png(q)
            aa, bb = a.rgb(), b.rgb()
            assert (a.width, a.height) == (b.width, b.height) == (320*scale, 200*scale)
            points = [(i//3 % a.width, i//3 // a.width) for i in range(0, len(aa), 3)
                      if aa[i:i+3] != bb[i:i+3]]
            origin_x, origin_y = (16, 76) if lang == 'en' else (46, 247)
            assert all(origin_x <= x < origin_x+64*scale and
                       origin_y <= y < origin_y+62*scale for x, y in points)
            if lang == 'en':
                english_weather_changes[index] = set(points)
            else:
                expected = {(46+(x-16)*3+dx, 247+(y-76)*3+dy)
                            for x, y in english_weather_changes[index]
                            for dx in range(3) for dy in range(3)}
                assert set(points) == expected
            changed.append(dict(step=index, pixels=len(points)))
        assert len(changed) == 13
        reload = PRIVATE / ('production-reload-' + lang)
        old_reload = ROOT / 'workplace/battle-exits-20261002' / ('production-reload-' + lang)
        assert read(reload / 'trace.json') == read(old_reload / 'trace.json')
        reload_pictures = sorted(reload.glob('*.png'))
        assert len(reload_pictures) == 16
        for image in reload_pictures:
            assert image.read_bytes() == (old_reload / image.name).read_bytes()
        gui[lang] = dict(states=493, battle_states=362, battle_unchanged=True,
                         pre_exit_pngs_identical=480, changed_weather_pictures=changed,
                         final_world_seed=w['probe_world_seed'], reload_states=16, reload_pngs=16)
    return dict(schema='kol-world-time-receipt-v1', native=native, gui=gui, historical=historical,
                product_sha256=sha(PRIVATE / 'kol-world-time-linux'),
                limitation='更新順序與戰後天候亂數有限CONFORMED；成本40／100差異及完整進場亂數仍DRAFT。')


def attachments():
    parent = read(PRIVATE / 'parent-preservation.json')
    paths = [ROOT / name for name in parent['sources']]
    paths += [p for p in PRIVATE.rglob('*') if p.is_file()]
    paths += [HERE / 'verify_time.py', HERE / 'audit_time.py', ROOT / 'docs/re/20261002-world-clock.md',
              ROOT / 'internal/world/time_part_test.go', ROOT / 'cmd/kol/battle_world_time_test.go']
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
        receipt = read(HERE / 'receipt.json')
        for name, digest in receipt['attachments'].items():
            assert sha(ROOT / name) == digest, name
        assert result == {k: v for k, v in receipt.items() if k != 'attachments'}
    print(json.dumps({k: v for k, v in result.items() if k != 'attachments'}, ensure_ascii=False))
