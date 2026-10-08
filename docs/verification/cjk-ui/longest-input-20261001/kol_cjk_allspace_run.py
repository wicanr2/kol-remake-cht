from pathlib import Path
import json,subprocess,hashlib,os
root=Path('workplace/cjk-longest-input-20261001/attempt2');selection=json.loads((root/'selection.json').read_text());assert 'all-spaces' not in selection['cases']
case={'tokens':['SPACE','SPACE','5','D','RET']+['SPACE']*16+['W'],'shots':[5,20,21,22],'maximum_text':' '*16,'width':256}
folder=root/'all-spaces';assert not folder.exists() and (root.stat().st_uid,root.stat().st_gid)==(os.getuid(),os.getgid());folder.mkdir();(folder/'script.txt').write_text(' '.join(case['tokens'])+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for lang in ('en','zh-Hant'):
    out=folder/lang;out.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(case['tokens']),'-shot-dir',str(out),'-shot-steps',','.join(map(str,case['shots'])),'-trace',str(out/'trace.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=180);(out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n');assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps'];assert rows[-1]['scene']=='flow:newchar';print(lang,len(rows),flush=True)
assert json.loads((folder/'en/trace.json').read_text())==json.loads((folder/'zh-Hant/trace.json').read_text())
selection['cases']['all-spaces']=case;(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
