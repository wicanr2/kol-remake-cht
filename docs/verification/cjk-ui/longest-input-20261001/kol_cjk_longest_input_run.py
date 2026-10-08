from pathlib import Path
import json,hashlib,os,subprocess
root=Path('workplace/cjk-longest-input-20261001/attempt2');assert not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid());root.mkdir()
fm=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
alphabet={g['character']:g['advance'] for g in fm['glyphs'] if len(g['character'])==1 and 'A'<=g['character']<='Z'}
assert max(alphabet,key=alphabet.get)=='W' and alphabet['W']==15
assert ' ' not in {g['character'] for g in fm['glyphs']} and fm['advance']==16
prefix=['SPACE','SPACE','5','D','RET']
cases={
    'letters':{'tokens':prefix+['W']*15+['BS','W','W','W','RET','RET','RET','D','RET'],'shots':[5,20,21,23,24,29],'maximum_text':'W'*16,'width':240},
    'spaces':{'tokens':prefix+['SPACE']*15+['W','W','RET','RET','RET','D','RET'],'shots':[5,20,21,22,27],'maximum_text':' '*15+'W','width':255}}
selection={'scope':'現行正式鍵盤可輸入的英文與空白；從標題正常建角核實滿16字、退格、游標、種族前進與肖像名牌。不宣稱中文輸入、重名政策或原版一致。',
    'seed':1,'seed_method':'執行前 -seed 1','capacity':16,'widest_letter':'W','letter_advance':15,'space_advance':16,
    'all_space_capacity_width':256,'cases':cases}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for name,case in cases.items():
    out=root/name;out.mkdir();(out/'script.txt').write_text(' '.join(case['tokens'])+'\n')
    for lang in ('en','zh-Hant'):
        folder=out/lang;folder.mkdir();args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(case['tokens']),'-shot-dir',str(folder),'-shot-steps',','.join(map(str,case['shots'])),'-trace',str(folder/'trace.json')]
        p=subprocess.run(args,capture_output=True,text=True,timeout=180);(folder/'engine.log').write_text(p.stdout+p.stderr)
        (folder/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n')
        assert p.returncode==0,p.stderr
        rows=json.loads((folder/'trace.json').read_text())['steps']
        assert rows[5]['scene']=='flow:newchar' and rows[-1]['scene']=='flow:newchar',rows[-1]
        print(name,lang,len(rows),rows[-1],flush=True)
    assert json.loads((out/'en/trace.json').read_text())==json.loads((out/'zh-Hant/trace.json').read_text())
