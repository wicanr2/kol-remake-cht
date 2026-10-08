"""在 Docker 內唯讀核對收尾稽核，來源掛 /src，無須執行遊戲。

工具鏈：kol-package:20260907，Python 3.11 標準函式庫。
在容器內執行 python3 /src/docs/verification/finish-audit-20261003/verify.py。
"""
from pathlib import Path
import hashlib
import json


ROOT = Path('/src')
WORK = ROOT / 'workplace/finish-audit-20261003'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def check(condition, message):
    if not condition:
        raise SystemExit(message)


receipt = read(ROOT / 'docs/verification/finish-audit-20261003/receipt.json')
for name, digest in receipt['attachments'].items():
    check(sha(ROOT / name) == digest, f'本輪附件變更：{name}')

preserved = read(WORK / 'parent-preservation.json')
parent_path = ROOT / preserved['parent_receipt']
check(sha(parent_path) == preserved['parent_receipt_sha256'], '前批收據變更')
parent = read(parent_path)
check(len(parent['attachments']) == preserved['attachments'] == 5078,
      '前批附件數不符')
for name, digest in parent['attachments'].items():
    path = (WORK / 'parent-inputs' / name
            if name in preserved['sources'] else ROOT / name)
    check(sha(path) == digest, f'前批附件變更：{name}')
for name, digest in parent['production_go'].items():
    check(sha(ROOT / name) == digest, f'正式 Go 變更：{name}')
check(len(parent['production_go']) == 325, '正式 Go 數量不符')
for name, digest in read(WORK / 'inputs-before.json').items():
    check(sha(ROOT / name) == digest, f'資產稽核來源變更：{name}')

assets = read(WORK / 'asset-audit.json')
check(assets['inventory_identical'], '資產清冊不相符')
check(assets['pngs_rebuilt_and_hash_equal'] == 773, '重建 PNG 數量不符')
check(assets['manifest_equal_except_input_basename'], 'manifest 不相符')
check(all(row['exit_code'] == 0 for row in assets['commands']), '資產命令失敗')
for row in assets['images']:
    for base in (ROOT / 'assets/original-derived-png/ega',
                 WORK / 'rebuild-inputs/assets/original-derived-png/ega'):
        check(sha(base / row['file']) == row['sha256'], f'PNG 變更：{row["file"]}')

text = read(WORK / 'cjk-text-fit.json')
check(text['checked'] == 9675 and text['failed'] == 0, '文字矩形失敗')
coverage = read(WORK / 'cjk-coverage.json')
check(coverage['checked_strings'] == 6101 and not coverage['missing'], '缺字')
check(not read(WORK / 'runtime-translation.json')['missing'], '來源缺譯')
check(all(row['exit_code'] == 0 for row in read(WORK / 'text-audit.json')['commands']),
      '文字命令失敗')
data = read(ROOT / 'docs/worklist.json')
check(sum(row.get('status') == 'done' for row in data['items']) == 84, '完成數漂移')
check(sum(row.get('status') != 'done' for row in data['items']) == 22, '開放數漂移')
before = read(WORK / 'before/docs/worklist.json')
check({row['id']: (row.get('status'), row['acceptance']) for row in data['items']}
      == {row['id']: (row.get('status'), row['acceptance']) for row in before['items']},
      '完成條件或開放狀態被改動')
print('PASS：773 PNG、9675 文字矩形、6101 字串；5078 前批附件；325 Go 未變；22 開放／84 完成')
