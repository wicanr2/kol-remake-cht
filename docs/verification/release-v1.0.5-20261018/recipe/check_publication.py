"""Docker內核對公開Release七附件，匿名下載並驗證SHA-256。"""
import datetime
import hashlib
import json
from pathlib import Path
import urllib.request
from zoneinfo import ZoneInfo

assert Path('/.dockerenv').exists()
version = 'v.1.0.5-20261018'
source = Path('/src/dist-all') / version / 'release'
out = Path('/out')
remote = json.loads((out / 'publication-remote.json').read_text())
assert remote['tag_name'] == version and not remote['draft'] and not remote['prerelease']
published = datetime.datetime.fromisoformat(remote['published_at'].replace('Z', '+00:00'))
assert published.astimezone(ZoneInfo('Asia/Taipei')).strftime('%Y%m%d') == version[-8:]
local = json.loads((source / 'SHA256SUMS.json').read_text())['files']
expected = {Path(row['path']).name: row for row in local}
checksum = source / 'SHA256SUMS.txt'
expected[checksum.name] = dict(bytes=checksum.stat().st_size, sha256=hashlib.sha256(checksum.read_bytes()).hexdigest())
assets = {row['name']: row for row in remote['assets']}
assert len(remote['assets']) == len(assets) == 7 and assets.keys() == expected.keys()
assert remote['body'].strip() == (out / 'release-notes.md').read_text().strip()
downloads = []
for name, row in expected.items():
    asset = assets[name]
    assert asset['state'] == 'uploaded' and asset['size'] == row['bytes']
    assert asset['digest'] == 'sha256:' + row['sha256']
    url = f'https://github.com/wicanr2/kol-remake-cht/releases/download/{version}/{name}'
    assert asset['browser_download_url'] == url
    request = urllib.request.Request(url, headers={'User-Agent': 'kol-release-verifier/1'})
    digest, total = hashlib.sha256(), 0
    with urllib.request.urlopen(request, timeout=45) as stream:
        assert stream.status == 200
        while block := stream.read(1 << 20):
            total += len(block)
            assert total <= row['bytes']
            digest.update(block)
    assert total == row['bytes'] and digest.hexdigest() == row['sha256']
    downloads.append(dict(name=name, bytes=total, sha256=digest.hexdigest(), anonymous_http_status=200))
    print('PUBLIC DOWNLOAD PASS', name, flush=True)
report = dict(status='PASS', version=version, url=remote['html_url'], release_id=remote['id'],
    published_at=remote['published_at'], publication_timezone='Asia/Taipei',
    source_tag='bc087d31183407907c754e252727b195d860aa9f',
    expected_assets=7, no_full_or_promo_uploaded=True, notes_equal=True, downloads=downloads)
(out / 'publication.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
