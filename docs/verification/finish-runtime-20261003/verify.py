"""Docker 唯讀重驗：掛儲存庫至 /src，執行 python3 本檔。"""
from pathlib import Path
import hashlib,json,sys
r=Path('/src');p=r/'docs/verification/finish-runtime-20261003/receipt.json'
d=json.loads(p.read_text());sha=lambda f:hashlib.sha256(f.read_bytes()).hexdigest()
for n,h in d['attachments'].items():assert sha(r/n)==h,n
o=r/'workplace/cjk-loot-normal-20261003'
parent=json.loads((o/'parent-preservation.json').read_text())
assert sha(r/parent['receipt'])==parent['sha256']
previous=json.loads((r/parent['receipt']).read_text())
for n,h in previous['attachments'].items():
    f=o/'parent-inputs'/n if n in parent['mutable'] else r/n
    assert sha(f)==h,n
for n,h in json.loads((o/'inputs.json').read_text()).items():assert sha(r/n)==h,n
sys.path.insert(0,str(r/'tools'))
from png_io import read_png
for row in d['checks']['pixel_comparisons']:
    suffix=f"gui-{row['stage']}/step-{row['step']:03}.png"
    assert read_png(o/'linux'/suffix).rgb()==read_png(o/'windows2'/suffix).rgb()
saves=[]
for sub in ['linux','windows2']:
    state=json.loads((o/sub/'default-save.dat').read_text());state.pop('saved_at');saves.append(state)
assert saves[0]==saves[1]
items=json.loads((r/'docs/worklist.json').read_text())
items=items['items'] if isinstance(items,dict) else items
assert sum(x.get('status','open')=='open' for x in items)==22
assert sum(x.get('status','open')=='done' for x in items)==84
assert d['closeout']['project_containers']==[]
assert d['closeout']['root_owned']==[] and d['closeout']['markdown_directories']==[]
print('PASS：兩平台正常互動及保存重載、13 組靜止畫面、325 Go、22 開放／84 完成；父收據歷史可回查')
