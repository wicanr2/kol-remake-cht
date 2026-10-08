import pathlib,json,subprocess
root=pathlib.Path('workplace/cjk-sign-natural-20261001')
tokens=(root/'prefix.txt').read_text().split()
tokens += (['>']*5+['RET','RET','RET'])*4
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
for phase, extra in [('party',[]),('door',['RET']),('leave',['RET','SPACE'])]:
    args=[binary,'-scene','flow','-seed','1','-flow-script',' '.join(tokens+extra),
          '-shot',str(root/f'{phase}.png'),'-trace',str(root/f'{phase}.json')]
    p=subprocess.run(args,capture_output=True,text=True,timeout=90)
    (root/f'{phase}.log').write_text(p.stdout+p.stderr)
    assert p.returncode==0,p.stderr
    row=json.loads((root/f'{phase}.json').read_text())['steps'][-1]
    print(phase,row)
(root/'party-prefix.txt').write_text(' '.join(tokens)+'\n')
