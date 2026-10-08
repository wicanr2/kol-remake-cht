"""Docker唯讀重驗；儲存庫掛/src，kol-package:20260907執行本檔。"""
from pathlib import Path
import hashlib,json,runpy
r=Path('/src');p=r/'docs/verification/current-platform-20261004/receipt.json'
def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
    return h.hexdigest()
d=json.loads(p.read_text())
for n,h in d['attachments'].items():assert sha(r/n)==h,n
ns=runpy.run_path(str(r/'workplace/current-platform-20261003/verification.py'))
assert ns['compute']()==d['checks']
for row in d['png_provenance']:assert sha(p.parent/row['file'])==row['sha256']
assert d['closeout']['project_containers']==[]
assert d['closeout']['root_owned']==[] and d['closeout']['markdown_directories']==[]
print('PASS：收據可重驗；327份來源建置、正常保存與有限雙語回歸通過。13組X11畫格全部相同='+str(d['checks']['all_13_frames_equal'])+'；完整remake仍未完成。')
