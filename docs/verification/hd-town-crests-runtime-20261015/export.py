"""匯出瀏覽器原生 Canvas PNG；本程式只解開 data URL，不繪圖或縮放。"""
from pathlib import Path
import base64, hashlib, html, json, re, shutil, subprocess

root=Path('/repo'); out=root/'workplace/hd-town-crests-20261015'
assert out.stat().st_uid==1000 and (out/'crests.js').stat().st_uid==1000
command=['chromium','--headless','--no-sandbox','--disable-dev-shm-usage','--disable-gpu','--user-data-dir=/tmp/kol-town-crests-chromium','--allow-file-access-from-files','--dump-dom','file:///repo/workplace/hd-town-crests-20261015/review.html']
result=subprocess.run(command,capture_output=True,text=True,timeout=45)
(out/'chromium.log').write_text(result.stderr)
assert result.returncode==0, result.stderr
match=re.search(r'<pre id="exports">(.*?)</pre>',result.stdout,re.S)
assert match, 'Canvas exports missing'
exports=json.loads(html.unescape(match.group(1)))
assert len(exports)==6
for name,url in exports.items():
 assert re.fullmatch(r'twn[1-6]pics-000.png',name) and url.startswith('data:image/png;base64,')
 data=base64.b64decode(url.split(',',1)[1],validate=True)
 assert data.startswith(b'\x89PNG\r\n\x1a\n')
 (out/name).write_bytes(data)
shutil.copyfile(root/'assets/hd/first-slice/interface/crest.png',out/'twn7pics-000.png')
record={'schema':'kol-hd-town-crests-canvas/1','command':command,'browser':subprocess.check_output(['chromium','--version'],text=True).strip(),'creator':'code-native Canvas extension of approved crest()','rights':'private-source-bound PNG; local/private only; no public grant','source_count':7,'new_count':6,'preserved':'twn7pics-000.png','creator_sha256':hashlib.sha256((out/'crests.js').read_bytes()).hexdigest(),'files':{n:hashlib.sha256((out/n).read_bytes()).hexdigest() for n in sorted(set(exports)|{'twn7pics-000.png'})}}
(out/'generation.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'exported':len(record['files']),'browser':record['browser']},ensure_ascii=False))
