"""Docker/Xvfb 內正常重播武器與防禦全列選取，不改遊戲狀態。"""
from pathlib import Path
import hashlib
import json
import os
import subprocess

root, out = Path('/src'), Path('/out')
assert out.stat().st_uid == os.getuid()
binary = out / 'kol-attack-frame-linux'
assert hashlib.sha256(binary.read_bytes()).hexdigest() == 'e5be211397bbb2eb11e355a77bbe490cc8a1cd8c618c049679dd97a1456bdc7a'
plan = json.loads((root / 'docs/verification/player-path/quest-chain.json').read_text())
# Olear 正常 YES 換人；Poontang 武器→目標→武器式，六次右移繞回，再進防禦式繞回。
keys = plan['script'].split()[:417] + ['RET'] + ' > RET D RET'.split()
keys += ['>'] * 6 + ['>', 'RET'] + ['>'] * 8
assert len(keys) == 438
shots = list(range(422, 439))
rows = []
for lang in ['en', 'zh-Hant']:
    dest = out / ('attack-row-fixed-' + lang)
    assert not dest.exists()
    dest.mkdir()
    cmd = [str(binary), '-scene', 'flow', '-seed', '1', '-lang', lang,
           '-ambush-rolls', plan['ambush_rolls'], '-flow-script', ' '.join(keys),
           '-shot-dir', str(dest), '-shot-steps', ','.join(map(str, shots)),
           '-trace', str(dest / 'trace.json')]
    (dest / 'command.json').write_text(json.dumps(cmd, ensure_ascii=False, indent=2)+'\n')
    run = subprocess.run(cmd, cwd=root, capture_output=True, text=True, timeout=350)
    (dest / 'run.log').write_text(run.stdout + run.stderr)
    assert run.returncode == 0, (lang, run.returncode, run.stderr[-1000:])
    states = json.loads((dest / 'trace.json').read_text())['steps']
    assert len(states) == 439
    for n in shots:
        assert states[n]['battle']['prompt'] == (2 if n < 430 else 5), (lang, n)
    rows.append(states)
assert rows[0] == rows[1]
receipt = dict(schema='kol-attack-row-fixed-normal-v1', seed=1, binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
    runs=2, inputs_each=len(keys), states_each=439, equal_language_states=True,
    shot_steps=shots, normal_prefix='既有quest-chain正常標題、六人招募、接任務、步行至site6、答應參戰',
    ambush_rolls=plan['ambush_rolls'], direct_entry=False, state_injection=False,
    scope='同圖示列與選取位置的呈現；不宣稱兩側整場戰鬥初始狀態相同')
(out / 'attack-row-fixed-normal-receipt.json').write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+'\n')
print(json.dumps(receipt, ensure_ascii=False))
