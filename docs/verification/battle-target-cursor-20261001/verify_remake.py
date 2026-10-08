"""Docker 內唯讀核對 RE-171 最終窄收據、來源及擁有權。"""
from pathlib import Path
import copy
import hashlib
import json
import os
import sys

ROOT = Path('/src')
sys.path.insert(0, str(ROOT / 'tools'))
from png_io import read_png

DEST = ROOT / 'docs/verification/battle-target-cursor-20261001'
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()


def verify(receipt):
    assert receipt['binary']['sha256'] == sha(ROOT / receipt['binary']['path'])
    for key in ['original_base', 'original_lifecycle', 'superseded']:
        row = receipt[key]
        assert sha(DEST / row['path']) == row['sha256']
    for row in receipt['inputs']:
        assert sha(ROOT / row['path']) == row['sha256'], row['path']
    for row in receipt['attachments']:
        p = DEST / row['file']
        assert sha(p) == row['sha256'], p
        assert p.stat().st_uid == row['uid'] == os.getuid(), p
        assert p.stat().st_gid == row['gid'] == os.getgid(), p
    for row in receipt['normal_changes']:
        assert row['changed_outside_field'] == 0
        if row['file'].endswith(('step-417.png', 'step-418.png')):
            assert row['changed_in_field'] == 0
    assert len(receipt['normal_changes']) == 34
    assert receipt['summary']['annotation_entries'] == 14
    assert receipt['summary']['remake_scaled_boxes'] == 104
    assert receipt['summary']['different_pixels'] == 0
    assert sha(DEST / receipt['verifier']['path']) == receipt['verifier']['sha256']
    provenance = receipt['provenance']
    assert sha(ROOT / provenance['ida_private_output']) == provenance['ida_private_sha256']


receipt = json.loads((DEST / 'remake-receipt.json').read_text())
verify(receipt)
for field in ['binary', 'original_lifecycle']:
    bad = copy.deepcopy(receipt)
    bad[field]['sha256'] = '0' * 64
    try:
        verify(bad)
    except AssertionError:
        pass
    else:
        raise AssertionError('不同來源反例未拒絕：' + field)

base = json.loads((DEST / 'receipt.json').read_text())
for row in base['attachments']:
    assert sha(DEST / row['file']) == row['sha256'], row['file']
life = json.loads((DEST / 'lifecycle-receipt.json').read_text())
assert len(life['checks']) == 52
for row in life['checks']:
    assert row['different_pixels'] == 0
    p = DEST / row['frame']
    assert sha(p) == row['frame_sha256']
    img = read_png(p)
    assert (img.width, img.height) == (320, 200)

pngs = [r for r in receipt['attachments'] if r['file'].endswith('.png')]
assert len(pngs) == 138
for row in pngs:
    img = read_png(DEST / row['file'])
    if row['file'].startswith('normal-final-'):
        expected = (960, 600) if 'zh-Hant' in row['file'] else (320, 200)
    else:
        scale = 3 if row['file'].endswith('-x3.png') else 1
        expected = (16 * scale, 16 * scale)
    assert (img.width, img.height) == expected, row['file']

bad_owned = [str(p.relative_to(ROOT)) for p in ROOT.rglob('*')
             if p.lstat().st_uid == 0 or
             (not p.is_symlink() and p.is_dir() and p.name.endswith('.md'))]
assert not bad_owned, bad_owned[:20]
print(json.dumps(dict(status='通過', final_attachments=len(receipt['attachments']),
    first_batch_attachments=len(base['attachments']), original_lifecycle_boxes=52,
    remake_pngs=len(pngs), different_pixels_outside_field=0,
    wrong_binary_and_original_hash_rejected=True, root_owned_or_md_directories=0,
    uid=os.getuid(), gid=os.getgid(), receipt_sha256=sha(DEST / 'remake-receipt.json'),
    verifier_sha256=sha(Path(__file__))), ensure_ascii=False))
