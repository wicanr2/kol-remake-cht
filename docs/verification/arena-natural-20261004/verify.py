"""在 /src:ro 的 kol-package:20260907 Docker 內執行；只核對有限正常玩家鏈。"""
from pathlib import Path
import argparse,json,sys
r=Path('/src');o=r/'workplace/arena-natural-20261004';sys.path.insert(0,str(o))
from proof import validate,digest,load
ap=argparse.ArgumentParser();ap.add_argument('--recorded-source',action='store_true',help='只核对已封存來源，不宣稱當前程式回歸');args=ap.parse_args()
receipt=load(r/'docs/verification/arena-natural-20261004/receipt.json')
for path,h in receipt['attachments'].items():assert digest(r/path)==h,path
for path,h in receipt['source_inputs'].items():
 assert digest(r/receipt['source_snapshot']/path)==h,path
 if not args.recorded_source:assert digest(r/path)==h,('current source changed',path)
assert digest(r/receipt['binary']['path'])==receipt['binary']['sha256']
parent=r/receipt['parent']['path'];assert digest(parent)==receipt['parent']['sha256']
for path,h in load(parent)['attachments'].items():assert digest(r/path)==h,('parent',path)
assert validate(r)==receipt['checks'] and receipt['normal_arena_chain_verified']
assert receipt['seed']['original'] is None and not receipt['seed']['original_execution']
print('有限正常競技場鏈通過：雙語各4204正常鍵、4205狀態；46真實重載／重存鍵、16完整畫格、2正常退出。')
print('附件與來源雜湊、正常訓練資格、下注／自然敗北結算、完整v3保存及反例通過；原版整場對拍未驗。')
