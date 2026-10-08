import json
import pathlib
import subprocess
import hashlib

root = pathlib.Path('workplace/cjk-sign-natural-20261001')
root.mkdir(exist_ok=True)
base = json.loads(pathlib.Path('docs/verification/player-path/full.json').read_text())['script'].split()[:18]
tokens = base + ['>']*5 + ['RET','RET','RET']
binary = 'workplace/cjk-edge-20261001/kol-edge-fixed'
run = subprocess.run([binary,'-scene','flow','-seed','1','-flow-script',' '.join(tokens),
                      '-shot',str(root/'recruit-first.png'),'-trace',str(root/'recruit-first.json')],
                      capture_output=True,text=True,timeout=90)
(root/'recruit-first.log').write_text(run.stdout+run.stderr)
(root/'seed-and-input.json').write_text(json.dumps({'seed':1,'seed_method':'-seed 1',
    'binary_sha256':hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest(),
    'script':' '.join(tokens),'scope':'正常標題與名冊開場後，透過旅店正式招募操作；不讀觀察存檔、不注入座標或遭遇骰值'},ensure_ascii=False,indent=2)+'\n')
assert run.returncode == 0,run.stderr
r = json.loads((root/'recruit-first.json').read_text())['steps']
print('LAST',r[-1])
print('PARTY_CHANGES',[(x['step'],x['town']['party']) for x in r if x.get('town',{}).get('party')])
(root/'prefix.txt').write_text(' '.join(tokens)+'\n')
