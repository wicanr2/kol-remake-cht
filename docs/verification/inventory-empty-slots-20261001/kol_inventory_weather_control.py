from pathlib import Path
import hashlib,json,subprocess,shutil
root=Path('workplace/inventory-empty-slots-20261001')
out=root/'weather-control';assert not out.exists();out.mkdir()
plan=json.loads(Path('docs/verification/player-path/full-chain-ending.json').read_text())
tokens=plan['script'].split()[:90]
for label,roster in [('old',root/'before-roster.json'),('new',Path('assets/original-derived-data/character-roster.json'))]:
    dest=out/label;dest.mkdir();load=dest/'input-save.json';shutil.copy2(plan['load'],load)
    args=['workplace/cjk-edge-20261001/kol-edge-fixed','-scene','play','-seed',str(plan['seed']),
          '-load',str(load),'-roster',str(roster),'-quest-sites',plan['quest_sites'],
          '-ambush-rolls',plan['ambush_rolls'],'-play-script',' '.join(tokens),
          '-shot-dir',str(dest),'-shot-steps','45,90','-trace',str(dest/'trace.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=90)
    (dest/'engine.log').write_text(p.stdout+p.stderr)
    (dest/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,
         'roster_sha256':hashlib.sha256(roster.read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr
for step in (45,90):
    a=out/'old'/f'step-{step:03d}.png';b=out/'new'/f'step-{step:03d}.png'
    assert a.read_bytes()==b.read_bytes()
    assert b.read_bytes()==(root/'all-before/full-chain-ending'/b.name).read_bytes()
assert (out/'old/trace.json').read_bytes()==(out/'new/trace.json').read_bytes()
print('相同執行檔／seed下，新舊名冊的90步前綴與兩張天候畫格完全相同；該圖差異不由空物品修正引起。')
