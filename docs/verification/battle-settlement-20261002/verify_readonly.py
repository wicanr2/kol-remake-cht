"""正式唯讀入口；沿用鎖定的來源驗證，正規化 JSON tuple/list 表示。"""
from pathlib import Path
import json, runpy

root = Path('/src')
here = root / 'docs/verification/battle-settlement-20261002'
driver = runpy.run_path(str(here / 'verify.py'))
actual = json.loads(json.dumps(driver['result'](), ensure_ascii=False))
stored = driver['read'](here / 'receipt.json')
for record in stored.pop('attachments'):
    path = root / record['path']
    assert driver['sha'](path) == record['sha256'], path
    assert path.stat().st_uid == path.stat().st_gid == 1000, path
assert actual == stored
print('正常61回合與戰後來源正式唯讀重驗通過；JSON表示正規化不更改比較欄位或原始收據。')
