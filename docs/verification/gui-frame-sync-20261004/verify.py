"""儲存庫唯讀掛/src，以kol-package:20260907執行本檔。"""
from pathlib import Path
import hashlib,json,runpy
r=Path('/src');p=r/'docs/verification/gui-frame-sync-20261004/receipt.json'
def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
    return h.hexdigest()
d=json.loads(p.read_text());parent=json.loads((r/d['parent']['receipt']).read_text())
assert sha(r/d['parent']['receipt'])==d['parent']['sha256']
for n,h in parent['attachments'].items():assert sha(r/n)==h,n
for n,h in d['attachments'].items():assert sha(r/n)==h,n
ns=runpy.run_path(str(r/'workplace/gui-frame-sync-20261004/checks.py'))
assert ns['compute']()==d['checks']
for row in d['png_provenance']:assert sha(p.parent/row['file'])==row['sha256']
assert all(d['run_results'][n]['exit_code']==0 for n in ['guarded_linux','guarded_wine','guarded_wine_controller','formal_expected'])
assert d['closeout']['project_containers']==[] and d['closeout']['root_owned']==[] and d['closeout']['markdown_directories']==[]
print('PASS：213真實鍵、13按鍵後畫格、完整v3保存、正常關窗；三組舊差異為擷取過早，完整remake仍未完成。')
