from pathlib import Path
import hashlib, json, subprocess
root=Path('workplace/inventory-empty-slots-20261001')
out=root/'compat2'; assert not out.exists(); out.mkdir()
source=Path('docs/verification/cjk-ui/longest-shop-20261001/en-save.json')
before=hashlib.sha256(source.read_bytes()).hexdigest()
args=['workplace/cjk-edge-20261001/kol-edge-fixed','-scene','play','-seed','1',
      '-lang','en','-load',str(source),'-save',str(out/'unused.json'),
      '-play-script','1 > RET '+'D '*16+'RET','-trace',str(out/'trace.json'),
      '-shot-dir',str(out),'-shot-steps','20']
p=subprocess.run(args,capture_output=True,text=True,timeout=90)
(out/'run.json').write_text(json.dumps({'command':args,'returncode':p.returncode,'input_sha256_before':before},indent=2)+'\n')
(out/'engine.log').write_text(p.stdout+p.stderr)
assert p.returncode==0,p.stderr
old=json.loads(Path('docs/verification/cjk-ui/longest-shop-20261001/en-reload-trace.json').read_text())
assert json.loads((out/'trace.json').read_text())==old
assert not (out/'unused.json').exists()
assert hashlib.sha256(source.read_bytes()).hexdigest()==before
print('舊 v3 存檔原樣載入，21 筆狀態與舊收據相同；來源檔未改。')
