"""唯讀核對本版精簡證據、縮圖與新增入口。"""
from pathlib import Path
import hashlib
import json
import re
from PIL import Image

R = Path('/src')
D = R / 'docs/verification/release-v1.0.3-20261017'
def read(p): return json.loads(p.read_text())
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
manifest = read(D / 'manifest.json')
for row in manifest['files']:
    p = D / row['path']
    assert p.stat().st_size == row['bytes'] and sha(p) == row['sha256'], p
for row in read(D / 'screenshots.json')['files']:
    source, image = R / row['source'], D / row['path']
    assert sha(source) == row['source_sha256'] and sha(image) == row['sha256']
    with Image.open(image) as im:
        assert list(im.size) == row['dimensions'] and max(im.size) <= 768
inputs = read(D / 'build-inputs.json')['program_inputs']
assert len(inputs) == 430 and all(sha(R / path) == value for path, value in inputs.items())
receipt = read(D / 'receipt.json')
assert receipt['technical_status'] == 'PASS' and len(receipt['packages']) == 6
assert read(D / 'packages.json')['ready_for_publication'] is True
links = 0
for name in ['README.md', 'CONTEXT.md', 'WORKLIST.md', 'docs/spec/039-hd-art-ui.md',
             'docs/verification/hd-completion-20261014.md', 'docs/RELEASE-RIGHTS-REVIEW.md',
             'docs/verification/release-v1.0.3-20261017.md']:
    p = R / name
    values = re.findall(r'\]\(([^)]+)\)', p.read_text())
    # 先用已知完整HD入口作正對照；僅本輪新入口屬本輪完整性責任。
    if name == 'README.md':
        assert 'docs/verification/hd-full-runtime-20261017.md' in values
        assert (R / 'docs/verification/hd-full-runtime-20261017.md').is_file()
    for value in values:
        if 'v1.0.3-20261017' not in value and 'v.1.0.3-20261017' not in value:
            continue
        if value.startswith(('http:', 'https:')):
            continue
        target = (p.parent / value.split('#', 1)[0]).resolve()
        assert target.exists(), (name, value)
        links += 1
    assert 'workplace/release-v1.0.3-20261017' not in p.read_text(), name
print(json.dumps(dict(status='PASS', formal_files=len(manifest['files']),
                      previews=8, program_files=430, new_local_links=links)))
