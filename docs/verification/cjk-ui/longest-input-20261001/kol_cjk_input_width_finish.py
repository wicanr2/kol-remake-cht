from pathlib import Path
import json, os, hashlib, subprocess
root=Path('workplace/cjk-longest-input-20261001/attempt2')
assert (root.stat().st_uid,root.stat().st_gid)==(os.getuid(),os.getgid())
old=Path('docs/verification/cjk-ui/font-width-20261001/text-fit.json');current=Path('docs/verification/cjk-text-fit.json')
assert hashlib.sha256(current.read_bytes()).hexdigest()==hashlib.sha256(old.read_bytes()).hexdigest()
args=['--layout','workplace/cjk-layout.json','--font','assets/ui-font/manifest.json','--strings','docs/i18n/newchar-zh-Hant.json','--bank-strings','docs/i18n/ui-zh-Hant.json','--ui-i18n','docs/i18n/ui-zh-Hant.json','--data','assets/original-derived-data/character-creation.json','--display-names','docs/i18n/display-names-zh-Hant.tsv','--town-i18n','docs/i18n/town-zh-Hant.json']
run=subprocess.run(['python3','tools/verify_text_fit.py',*args,'--out',str(current)],capture_output=True,text=True)
assert run.returncode==0,(run.stdout,run.stderr)
new=json.loads(current.read_text());previous=json.loads(old.read_text());assert len(new['checks'])==len(previous['checks'])==9635
changed=[(a,b) for a,b in zip(previous['checks'],new['checks']) if a!=b]
assert len(changed)==1 and changed[0][0]['kind']=='name' and changed[0][0]['width']==240 and changed[0][1]['width']==256
assert not new['failed'] if 'failed' in new else all(x['pass'] for x in new['checks'])
layout=json.loads(Path('workplace/cjk-layout.json').read_text());layout['text_box']['LeftW']=248
(root/'negative-layout.json').write_text(json.dumps(layout,ensure_ascii=False,indent=2)+'\n')
base=Path('docs/verification/cjk-ui/font-width-20261001/verify_text_fit.py.txt')
negative_args=list(args);negative_args[1]=str(root/'negative-layout.json')
controls=[]
for label,tool in [('before',str(base)),('after','tools/verify_text_fit.py')]:
    out=root/f'negative-{label}.json';command=['python3',tool,*negative_args,'--out',str(out)]
    result=subprocess.run(command,capture_output=True,text=True);doc=json.loads(out.read_text())
    failed=[r for r in doc['checks'] if not r['pass']]
    controls.append({'label':label,'command':command,'returncode':result.returncode,'stdout':result.stdout,'stderr':result.stderr,'failed':failed})
assert controls[0]['returncode']==0 and not controls[0]['failed'],controls[0]
assert controls[1]['returncode']==1 and len(controls[1]['failed'])==1 and controls[1]['failed'][0]['kind']=='name',controls[1]
(root/'width-control.json').write_text(json.dumps({'command':['python3','tools/verify_text_fit.py',*args,'--out',str(current)],'returncode':run.returncode,'stdout':run.stdout,'changes':changed,'negative_controls':controls},ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'checked':9635,'failed':0,'old_capacity':240,'new_capacity':256,'negative_before':0,'negative_after':1},ensure_ascii=False))
