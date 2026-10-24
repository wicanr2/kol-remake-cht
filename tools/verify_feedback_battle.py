#!/usr/bin/env python3
"""SPEC-041：用 SPEC-040 受控場景走武器、部位、防禦三層正式選單。"""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--root', type=Path, default=Path('/repo'))
parser.add_argument('--bin', type=Path, required=True)
parser.add_argument('--out', type=Path, required=True)
args = parser.parse_args()
assert Path('/.dockerenv').exists()
root, binary, out = args.root.resolve(), args.bin.resolve(), args.out.resolve()
dest = out / 'battle'
dest.mkdir()
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
base = out / 'gui/config/kol-remake-cht/save.dat'
state = json.loads(base.read_text())
state['party'], state['speaker'] = [14], 0
state['characters'][14]['quests'] = [0, 1, 0, 0, 0, 0]
fixture = dest / 'party.json'
fixture.write_text(json.dumps(state))
sites = json.loads((root / 'assets/original-derived-data/quest-sites.json').read_text())
site = next(s for s in sites['sites'] if s['index'] == 4)
site.update(monster=15, monster_name='RUFFIAN', level=1, count=1,
            monster_x=[site['party_x'][0]-1], monster_y=[site['party_y'][0]])
site_path = dest / 'sites.json'
site_path.write_text(json.dumps(sites))
commands = []


def capture(name, script, language='en'):
    folder = dest / name
    folder.mkdir()
    command = [str(binary), '-lang', language, '-music', 'off', '-seed', '1',
               '-scene', 'play', '-load', str(fixture), '-quest-sites', str(site_path),
               '-verify-scenes', '-verify-lock-hp', '-verify-world-start', '48,9',
               '-play-script', ' '.join(script), '-shot-dir', str(folder),
               '-trace', str(folder / 'trace.json'), '-trace-detail']
    with (dest / (name + '.log')).open('w') as log:
        subprocess.run(command, cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=90)
    commands.append(command)
    trace = json.loads((folder / 'trace.json').read_text())['steps']
    return trace[-1]['battle']


script = ['RET', 'SPACE']
b = capture('entry', script)
assert b['prompt'] == 0 and 0x21 in b['bar'], b
script += ['>'] * b['bar'].index(0x21) + ['RET', 'RET']
b = capture('weapon', script)
assert b['prompt'] == 2, b
script += ['>', '>', '>', 'RET']
expected = [0x0F, 0x2F, 0x41, 0x42]
for language in ('en', 'zh-Hant'):
    b = capture('area-' + language, script, language)
    assert b['prompt'] == 13 and b['bar'] == expected, b
    guard_script = script + ['D', 'D', 'RET']
    b = capture('guard-' + language, guard_script, language)
    assert b['prompt'] == 5 and len(b['bar']) == 8, b
    b = capture('committed-' + language, guard_script + ['D', 'D', 'D', 'RET'], language)
    assert b['prompt'] == 0, b
report = dict(status='PASS', spec='docs/spec/041-player-feedback-fixes.md', seed=1,
              original_parity_claimed=False, binary_sha256=sha(binary),
              verifier_sha256=sha(Path(__file__)), input_save_sha256=sha(base),
              controlled_changes='角色14接任務4；任務4改為一名等級1 Ruffian、相鄰站位；SPEC-040免傷',
              scope='武器選目標→攻擊式→上中下→防禦→寫回；英文及繁中；方向鍵選項',
              commands=commands,
              outputs_sha256={str(p.relative_to(dest)): sha(p) for p in dest.rglob('*') if p.is_file()})
(dest / 'receipt.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print('PASS: SPEC-041 受控武器攻擊部位選單，雙語', flush=True)
