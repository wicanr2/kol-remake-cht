from pathlib import Path
import json,subprocess,hashlib,time,os
root=Path('workplace/cjk-equipment-20261001')
selection=json.loads((root/'selection.json').read_text())
for p,h in selection['inputs_sha256'].items():assert hashlib.sha256(Path(p).read_bytes()).hexdigest()==h,p
mother=Path('docs/verification/cjk-ui/natural-sign-20261001')
original=(mother/'script.txt').read_text().split()
check_steps=[160,176,192,208,241,257]
tokens=[];checkpoints=[];mapping=[]
for step,key in enumerate(original[:max(check_steps)],1):
    tokens.append(key);mapping.append({'source_step':step,'replay_step':len(tokens)})
    if step in check_steps:
        before=len(tokens);tokens.extend(['SPACE','SPACE'])
        checkpoints.append({'source_step':step,'before':before,'equipment':before+1,'restored':before+2})
selection['checkpoints']=checkpoints;selection['source_state_mapping']=mapping
selection['prefix_inputs']=len(tokens)
selection['trace_snapshot_contract']='cmd/kol/main.go 只在stage改變時匯出battle.roster；相鄰切換驗證以該階段快照補齊省略值，其餘欄位逐欄相同。另以像素驗切換後復原。'
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
(root/'script.txt').write_text(' '.join(tokens)+'\n')
binary='workplace/cjk-edge-20261001/kol-edge-fixed'
assert hashlib.sha256(Path(binary).read_bytes()).hexdigest()=='f969a63f195f29561b32f8bef333d0933d4d4d69641025981249d127017891d4'
shots=sorted({n for c in checkpoints for k,n in c.items() if k!='source_step'})
for lang in ['en','zh-Hant']:
    out=root/(lang+'-current');assert not out.exists();out.mkdir()
    args=[binary,'-scene','flow','-seed','1','-lang',lang,'-flow-script',' '.join(tokens),'-shot-dir',str(out),'-shot-steps',','.join(map(str,shots)),'-trace',str(out/'trace.json')]
    begin=time.monotonic();p=subprocess.run(args,capture_output=True,text=True,timeout=150)
    (out/'engine.log').write_text(p.stdout+p.stderr)
    (out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-begin,'binary_sha256':hashlib.sha256(Path(binary).read_bytes()).hexdigest()},ensure_ascii=False,indent=2)+'\n')
    assert p.returncode==0,p.stderr
    rows=json.loads((out/'trace.json').read_text())['steps']
    reference=json.loads((mother/(lang+'-trace.json')).read_text())['steps']
    def normalized(row):return {k:v for k,v in row.items() if k not in ['step','name']}
    assert normalized(rows[0])==normalized(reference[0])
    for x in mapping:assert normalized(rows[x['replay_step']])==normalized(reference[x['source_step']]),(lang,x)
    for c in checkpoints:
        a,b,z=[rows[c[k]] for k in ['before','equipment','restored']]
        assert a['battle']['stage']==2 and a['battle']['prompt']==0,c
        snapshots=[normalized(row) for row in [a,b,z]]
        if 'roster' in snapshots[0]['battle']:
            for item in snapshots[1:]:
                if 'roster' not in item['battle']:item['battle']['roster']=snapshots[0]['battle']['roster']
        assert snapshots[0]==snapshots[1]==snapshots[2],c
        assert (out/f'step-{c["before"]:03d}.png').read_bytes()==(out/f'step-{c["restored"]:03d}.png').read_bytes(),c
    print(lang,len(rows),'筆；258個母路徑狀態及六次切換／復原全部一致。',flush=True)
en=json.loads((root/'en-current/trace.json').read_text())['steps'];zh=json.loads((root/'zh-Hant-current/trace.json').read_text())['steps']
differences=[]
for a,b in zip(en,zh):
    if a==b:continue
    a=json.loads(json.dumps(a));b=json.loads(json.dumps(b))
    assert a['step']==159 and a['battle']['stage']==b['battle']['stage']==1
    assert a['battle'].pop('encounter_text')==['YOU HAVE BEEN AMBUSHED','BY 3, LEVEL 1 CLIFF TROLLS!']
    assert b['battle'].pop('encounter_text')==json.loads((mother/'zh-Hant-trace.json').read_text())['steps'][159]['battle']['encounter_text']
    assert a==b;differences.append(159)
assert differences==[159],differences
print('英中270筆只有已知遭遇顯示譯文一筆不同；其他狀態、切換前後畫格一致。',flush=True)
