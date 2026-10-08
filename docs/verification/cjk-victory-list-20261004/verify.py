"""在有界 kol-package:20260907 Docker 內，專案唯讀掛載 /src 後重驗。"""
from pathlib import Path
import argparse,subprocess,sys
r=Path('/src');o=r/'workplace/cjk-victory-list-20261004';sys.path.insert(0,str(o))
from proof import check,load,sha
ap=argparse.ArgumentParser();ap.add_argument('--recorded-source',action='store_true');args=ap.parse_args()
d=load(r/'docs/verification/cjk-victory-list-20261004/receipt.json')
for rel,h in d['attachments'].items():assert sha(r/rel)==h,rel
parent=r/d['parent']['path'];assert sha(parent)==d['parent']['sha256']
subprocess.run([sys.executable,str(parent.with_name('verify.py')),'--recorded-source'],check=True)
for rel,h in d['source_inputs'].items():
    assert sha(r/d['source_snapshot']/rel)==h,rel
    if not args.recorded_source:assert sha(r/rel)==h,('current source changed',rel)
assert sha(r/d['binary']['path'])==d['binary']['sha256']
assert check(r)==d['checks']
print('現行自然勝利：雙語各460正常鍵、461狀態；第6回合SLEDGE、戰錘正常領取。')
print('雙語各698鍵至正常新保存、24鍵重載；完整v3除時間戳一致，34格及戰錘record保持。')
print('508真實X11鍵、20完整RGB閘門、三次正常退出0；十個清單字模安全矩形零差。')
print('沒有新增原版整場oracle，不宣稱四套怪物資料、最長清單或全remake完成。')
