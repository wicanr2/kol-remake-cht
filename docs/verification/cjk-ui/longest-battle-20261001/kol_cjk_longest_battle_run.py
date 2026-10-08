from pathlib import Path
import json,hashlib,os,subprocess,time
root=Path('workplace/cjk-longest-battle-20261001')
assert not root.exists();assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid());root.mkdir()
source=Path('docs/verification/cjk-ui/natural-sign-20261001')
receipt=json.loads((source/'receipt.json').read_text())
for name in ['zh-Hant-trace.json','script.txt']:
    assert hashlib.sha256((source/name).read_bytes()).hexdigest()==receipt['files'][name]
rows=json.loads((source/'zh-Hant-trace.json').read_text())['steps']
font=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
adv={g['character']:g['advance'] for g in font['glyphs']}
def joined(lines):
    result=''
    def word(c):return c.isascii() and (c.isalnum() or c=="'")
    for line in lines:
        if result and line and word(result[-1]) and word(line[0]):result+=' '
        result+=line
    return result
candidates=[]
for row in rows:
    b=row.get('battle') or {}
    if b.get('prompt') in (9,10,11,12) and b.get('exec_text'):
        text=joined(b['exec_text']);missing=[c for c in text if c not in adv and not c.isspace()];assert not missing,missing
        candidates.append({'step':row['step'],'pixel_width':sum(adv.get(c,font['advance']) for c in text),'text':text,'lines':b['exec_text']})
maximum=max(x['pixel_width'] for x in candidates)
selected=sorted(candidates,key=lambda x:(-x['pixel_width'],x['step']))[:3]
assert maximum==564 and selected[0]['step']==2566
shots=sorted({x['step'] for x in selected}|{2565,2567})
tokens=(source/'script.txt').read_text().split()[:max(shots)]
binary=Path(receipt['binary']['path']);assert hashlib.sha256(binary.read_bytes()).hexdigest()==receipt['binary']['sha256']
selection={'scope':'既有固定種子 1、正常六人招募與兩次自然戰鬥的 3,081 筆完整路徑；先選全母體最寬可見訊息，再重播相同前綴擷取畫格。不宣稱全遊戲最長內容。',
    'source':str(source/'receipt.json'),'source_trace_sha256':receipt['files']['zh-Hant-trace.json'],
    'candidate_count':len(candidates),'unique_messages':len(set(x['text'] for x in candidates)),
    'maximum_width':maximum,'selected':selected,'shot_steps':shots,'input_count':len(tokens),
    'seed':1,'seed_method':'執行前 -seed 1','ambush_rolls':None,'load':None,'monsters':None,'quest_sites':None,
    'font_sha256':hashlib.sha256(Path('assets/ui-font/manifest.json').read_bytes()).hexdigest()}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
(root/'script.txt').write_text(' '.join(tokens)+'\n')
args=[str(binary),'-scene','flow','-seed','1','-lang','zh-Hant','-flow-script',' '.join(tokens),'-shot-dir',str(root/'frames'),'-shot-steps',','.join(map(str,shots)),'-trace',str(root/'trace.json')]
started=time.monotonic();p=subprocess.run(args,capture_output=True,text=True,timeout=360)
(root/'engine.log').write_text(p.stdout+p.stderr)
(root/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'seconds':time.monotonic()-started,'binary_sha256':receipt['binary']['sha256']},ensure_ascii=False,indent=2)+'\n')
assert p.returncode==0,p.stderr[-2500:]
actual=json.loads((root/'trace.json').read_text())['steps'];assert actual==rows[:len(actual)]
assert len(actual)==len(tokens)+1
print(json.dumps({'states':len(actual),'maximum_width':maximum,'selected_steps':[x['step'] for x in selected],'identical_reference_prefix':True},ensure_ascii=False),flush=True)
