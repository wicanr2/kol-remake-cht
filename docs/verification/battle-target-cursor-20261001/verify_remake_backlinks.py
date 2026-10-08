"""在 Docker 內核對 RE-171 續核來源勘誤；首批檢查器另行保留。"""
from pathlib import Path
import copy
import json

ROOT = Path('/src')
RECEIPT = ROOT / 'docs/verification/battle-target-cursor-20261001/receipt.json'


def verify(rows):
    for row in rows:
        assert row['module'] == 'DOS/KOLTOWN-unpacked.EXE'
        assert row['address_space'] == 'IDA linear'
        assert row['confidence'] == '已證實'
        index = json.loads((ROOT / row['index']).read_text())
        assert index['sha256'] == row['input_sha256']
        assert index['address_space'] == row['address_space']
        entry = index['entries'][row['address']]
        assert entry['confidence'] == row['confidence'] and entry['evidence']
        evidence = (ROOT / row['evidence']).read_text()
        assert row['address'].lower() in evidence.lower()
        assert row['input_sha256'] in evidence
        assert row['required_correction'] in (ROOT / row['older_spec']).read_text()


data = json.loads(RECEIPT.read_text())
current = json.loads(RECEIPT.with_name('remake-receipt.json').read_text())
rows = data['resolution_backlinks'] + current['resolution_backlinks']
assert len(rows) == 3
verify(rows)
source = rows[-1]
assert source['address'] == '0x249cb'
assert 'sub_249AE' in (ROOT / source['evidence']).read_text()
assert '0x1F16A' in (ROOT / source['evidence']).read_text()
assert '摘要來源勘誤' in (ROOT / 'docs/spec/010-combat-commands.md').read_text()
negative = copy.deepcopy(rows)
negative[0]['required_correction'] = '此標記故意不存在_RE171_negative'
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('缺勘誤標記反例未拒絕')
negative = copy.deepcopy(rows)
negative[0]['input_sha256'] = '0' * 64
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('不同原版模組反例未拒絕')
negative = copy.deepcopy(rows)
negative[-1]['required_correction'] = '故意刪除來源勘誤_RE171_source_negative'
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('RE-160缺來源勘誤反例未拒絕')
negative = copy.deepcopy(rows)
negative[-1]['confidence'] = '假說'
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('來源證據等級反例未拒絕')
print('RE-171三筆反向引用通過；缺標記、模組雜湊、來源勘誤與等級反例均拒絕。')
