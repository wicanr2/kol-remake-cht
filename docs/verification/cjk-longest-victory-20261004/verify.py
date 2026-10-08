"""有界 kol-package:20260907 Docker 內，專案唯讀掛載 /src 後重驗。"""
from pathlib import Path
import argparse,subprocess,sys
r=Path('/src');o=r/'workplace/cjk-longest-victory-20261004';sys.path.insert(0,str(o))
from proof import check,load,sha
ap=argparse.ArgumentParser();ap.add_argument('--recorded-source',action='store_true');args=ap.parse_args()
d=load(r/'docs/verification/cjk-longest-victory-20261004/receipt.json')
for rel,h in d['attachments'].items():assert sha(r/rel)==h,rel
parent=r/d['parent']['path'];assert sha(parent)==d['parent']['sha256']
subprocess.run([sys.executable,str(parent.with_name('verify.py')),'--recorded-source'],check=True)
for rel,h in d['source_inputs'].items():
    assert sha(r/d['source_snapshot']/rel)==h,rel
    if not args.recorded_source:assert sha(r/rel)==h,('current source changed',rel)
assert sha(r/d['binary']['path'])==d['binary']['sha256']
assert check(r,args.recorded_source)==d['checks']
print('最長勝利清單：316候選上界80px；正常Spiney Mace清單、領取及保存。')
print('雙語各570正常鍵自然SLEDGE第7回合勝利；各1914鍵保存、45鍵重載。')
print('完整v3只排除時間戳相同；再保存只有Kenneth正常56住宿差，16-byte稀有記錄保留。')
print('660真實X11鍵、23完整RGB閘門、三次正常退出0；十四個字模安全矩形零差。')
print('使用既有原版住宿模式；沒有新增原版oracle，不宣稱四套怪物資料或全remake完成。')
