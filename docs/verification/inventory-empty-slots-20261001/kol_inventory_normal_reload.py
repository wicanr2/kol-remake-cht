from pathlib import Path
import json, os, hashlib, subprocess
root=Path('workplace/inventory-empty-slots-20261001/normal')
def sha(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
tokens=['1','>','RET']+['D']*11+['RET']
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for lang in ('en','zh-Hant'):
    savepath=root/lang/'save.json';saved=json.loads(savepath.read_text())
    leader=saved['characters'][saved['party'][0]]
    assert leader['name']=='Theodoric' and leader['gold']==918 and len(leader['backpack'])==1
    assert leader['backpack'][0]['name']=='Girak Helbak' and leader['backpack'][0]['record']=='c3a641581056082b002300370000a000'
    out=root/f'{lang}-reload2';assert not out.exists()
    assert (root.stat().st_uid,root.stat().st_gid)==(os.getuid(),os.getgid());out.mkdir()
    args=[binary,'-scene','play','-seed','1','-lang',lang,'-load',str(savepath),'-save',str(out/'unused-save.json'),'-play-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps','0,1,3,14,15','-trace',str(out/'trace.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=90)
    (out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'binary_sha256':sha(binary),'load_sha256_after':sha(savepath),'normal_save':str(savepath),'script':tokens},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps'];assert len(rows)==16
    assert rows[-1]['town']['outside_sheet']['page']=='items' and rows[-1]['town']['outside_sheet']['member']=='Theodoric'
    assert not (out/'unused-save.json').exists()
    print(lang,'正常保存食品後重載',len(rows),flush=True)
assert json.loads((root/'en-reload2/trace.json').read_text())==json.loads((root/'zh-Hant-reload2/trace.json').read_text())
print('英中重載全部狀態相同',flush=True)
