"""Docker 唯讀重驗：儲存庫掛 /src，python3 執行本檔。"""
from pathlib import Path
import hashlib,json,runpy
r=Path('/src');p=r/'docs/verification/loot-belt-20261003/receipt.json'
d=json.loads(p.read_text())
def sha(p):
 h=hashlib.sha256()
 with p.open('rb')as f:
  for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
 return h.hexdigest()
for n,h in d['attachments'].items():assert sha(r/n)==h,n
ns=runpy.run_path(str(r/'workplace/cjk-loot-belt-fixed-20261003/checks.py'))
assert ns['compute']()==d['checks']
for row in d['png_provenance']:assert sha(p.parent/row['file'])==row['sha256']
assert d['closeout']['project_containers']==[] and d['closeout']['root_owned']==[] and d['closeout']['markdown_directories']==[]
print('PASS：正式雙語正常腰帶保存及重載、來源與PNG雜湊、三處80像素完整字模；完整remake尚未完成')
