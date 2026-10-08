import pathlib,json,subprocess,hashlib,time
root=pathlib.Path('workplace/cjk-sign-natural-20261001')
tokens=(root/'residence-route.txt').read_text().split()+['RET']
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
steps=[18,58,60,72,159,2976,2977,len(tokens)-1,len(tokens)]
for lang in ('en','zh-Hant'):
    out=root/f'replay-{lang}'
    out.mkdir(exist_ok=True)
    args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),
          '-shot-dir',str(out),'-shot-steps',','.join(map(str,sorted(set(steps)))),
          '-trace',str(out/'trace.json')]
    started=time.monotonic()
    p=subprocess.run(args,capture_output=True,text=True,timeout=480)
    (out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-started,
        'seed':1,'seed_method':'-seed 1','binary_sha256':hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest()},
        ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr[-2500:]
    rows=json.loads((out/'trace.json').read_text())['steps']
    print(lang,'states',len(rows),'last',rows[-1]['town'],flush=True)
(root/'final-script.txt').write_text(' '.join(tokens)+'\n')
