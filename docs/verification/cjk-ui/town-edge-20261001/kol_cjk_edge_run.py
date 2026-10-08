import hashlib
import json
import pathlib
import subprocess
import sys

root = pathlib.Path('workplace/cjk-edge-20261001')
phase, binary = sys.argv[1:]
base = json.loads(pathlib.Path('docs/verification/player-path/full.json').read_text())
prefix = base['script'].split()[:19]
edge = json.loads(pathlib.Path('docs/verification/player-path/town-edge.json').read_text())
plans = {'edge': ' '.join(prefix + edge['script'].split())}
if phase == 'after':
    plans['building'] = base['script']
    plans['leave'] = ' '.join(prefix + edge['script'].split()[:11] + ['RET'])
root.mkdir(exist_ok=True)
for case, script in plans.items():
  for lang in ('en', 'zh-Hant'):
    target = root / phase / lang if case == 'edge' else root / phase / case / lang
    target.mkdir(parents=True, exist_ok=True)
    args = [binary, '-scene', 'flow', '-seed', '1', '-lang', lang,
            '-flow-script', script, '-shot-dir', str(target),
            '-shot-steps', '19,21,22,23,24,26,29,30,31,32', '-trace', str(target / 'trace.json')]
    run = subprocess.run(args, capture_output=True, text=True, timeout=90)
    (target / 'engine.log').write_text(run.stdout + run.stderr)
    record = {'command': args, 'returncode': run.returncode,
              'binary_sha256': hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest(),
              'seed': 1, 'seed_method': '-seed 1',
              'scope': '從標題、名冊與旅店正常按鍵走到城門；無讀檔、座標注入或遭遇控制'}
    (target / 'run.json').write_text(json.dumps(record, ensure_ascii=False, indent=2) + '\n')
    if run.returncode:
        raise RuntimeError(run.stderr[-2000:])
    rows = json.loads((target / 'trace.json').read_text())['steps']
    print(phase, case, lang, len(rows), [(r['step'], r.get('town',{}).get('mode'),
          r.get('town',{}).get('at_edge'),r.get('town',{}).get('x'),r.get('town',{}).get('y'))
          for r in rows if r['step'] >= 29])
