"""Docker 內驗證本輪 GUI 差異與收據擁有權，輸出雜湊清冊。"""
import hashlib
import json
import os
import subprocess
from pathlib import Path

root = Path('/out')
assert root.stat().st_uid == os.getuid()
def rgb(name):
    p = root / name
    size = subprocess.check_output(['identify', '-format', '%wx%h', str(p)]).decode()
    assert size == '1280x800', size
    return subprocess.check_output(['convert', str(p), '-depth', '8', 'rgb:-'])

a, b = rgb('weather-normal-1.png'), rgb('weather-normal-4.png')
changed = []
for pixel in range(1280*800):
    i = pixel*3
    if a[i:i+3] != b[i:i+3]:
        x, y = pixel % 1280, pixel // 1280
        assert 56 <= x < 324 and 324 <= y < 458, (x, y)
        changed.append((x, y))
assert changed, '正常野外天候未動畫'
assert (root/'goodbye-waiting-diff.txt').read_text().strip() == '0'
assert 'PASS' in (root/'oracle-weather-fixed/player-gaps.log').read_text()
oracle = json.loads((root/'oracle-weather-fixed/manifest.json').read_text())
assert oracle['source_changed_during_run'] is False
names = ['kol', 'tests.log', 'build.log', 'goodbye-gui.log', 'goodbye-waiting.png',
         'goodbye-outside.png', 'weather-gui.log', 'weather-normal-1.png',
         'weather-normal-4.png', 'arena-extraction.json', 'arena-overview.png',
         'oracle-single/manifest.json', 'oracle-party/manifest.json',
         'oracle-weather-fixed/manifest.json', 'oracle-weather-fixed/player-gaps.log']
files = []
for name in names:
    p = root/name
    st = p.stat()
    assert (st.st_uid, st.st_gid) == (os.getuid(), os.getgid()), name
    files.append(dict(path=name, sha256=hashlib.sha256(p.read_bytes()).hexdigest(),
                      uid=st.st_uid, gid=st.st_gid))
report = dict(schema='kol-player-gaps-verification-v1', files=files,
              weather_changed_pixels=len(changed),
              weather_bbox=[min(x for x,y in changed),min(y for x,y in changed),
                            max(x for x,y in changed)+1,max(y for x,y in changed)+1],
              weather_outside_sky_changed_pixels=0,
              note='原版受控像素比較與正常重製 GUI 分開驗證；不是完整原版正常野外同狀態。')
out = root/'summary.json'
if out.exists(): assert out.stat().st_uid == os.getuid()
out.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
print(json.dumps({k:v for k,v in report.items() if k!='files'}, ensure_ascii=False))
