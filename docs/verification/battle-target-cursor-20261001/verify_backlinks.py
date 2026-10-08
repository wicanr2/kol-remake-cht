"""在 Docker 內核對 RE-171 分級索引與舊規格回填；不改任何檔案。"""
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
verify(data['resolution_backlinks'])
negative = copy.deepcopy(data['resolution_backlinks'])
negative[0]['required_correction'] = '此標記故意不存在_RE171_negative'
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('缺勘誤標記反例未拒絕')
negative = copy.deepcopy(data['resolution_backlinks'])
negative[0]['input_sha256'] = '0' * 64
try:
    verify(negative)
except AssertionError:
    pass
else:
    raise AssertionError('不同原版模組反例未拒絕')
print('RE-171 反向引用通過；缺標記與不同模組雜湊反例均拒絕。')
