"""Docker 唯讀重驗：儲存庫掛 /src，python3 執行本檔。"""
from pathlib import Path
import copy,hashlib,json
r=Path('/src');d=r/'docs/verification/quest-return-20261003'
receipt=json.loads((d/'receipt.json').read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
for n,h in receipt['attachments'].items():assert sha(r/n)==h,n
for n,h in receipt['input_hashes'].items():assert sha(r/receipt['source_snapshot']/n)==h,n
def normal(rows):
 rows=copy.deepcopy(rows)
 for row in rows:
  row.get('world',{}).pop('prompt_lines',None)
  for key in ['encounter_text','exec_text']:row.get('battle',{}).pop(key,None)
 return rows
o=r/'workplace/quest-return-20261003'
states=[json.loads((o/f'formal-fixed-{lang}/trace.json').read_text())['steps'] for lang in ['en','zh-Hant']]
assert len(states[0])==len(states[1])==190 and normal(states[0])==normal(states[1])
for i,xy in zip(range(185,190),[(4,4),(3,4),(4,4),(4,3),(4,4)]):
 w=states[0][i]['world'];assert (w['sub_x'],w['sub_y'])==xy and w['mode']==0
assert states[0][185]['world']['ticks']==760 and states[0][189]['world']['ticks']==540
bad=copy.deepcopy(states[1]);bad[186]['world']['sub_x']+=1;assert normal(states[0])!=normal(bad)
for p in receipt['png_provenance']:assert sha(d/p['file'])==p['sha256']
closeout=json.loads((d/'closeout.json').read_text())
assert closeout['project_containers']==[] and closeout['root_owned']==[] and closeout['markdown_directories']==[]
print('PASS：來源快照、189正常鍵、雙語190狀態、四方向與位置反例、18張PNG；原版動態parity未驗')
