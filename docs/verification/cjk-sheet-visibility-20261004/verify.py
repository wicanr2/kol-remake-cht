"""用kol-package:20260907將專案唯讀掛到/src，核對有限資料頁字模與真實視窗。"""
from pathlib import Path
import argparse,sys
r=Path('/src');o=r/'workplace/cjk-sheet-visibility-20261004';sys.path.insert(0,str(o))
from proof import check,sha,load
ap=argparse.ArgumentParser();ap.add_argument('--recorded-source',action='store_true');args=ap.parse_args()
receipt=load(r/'docs/verification/cjk-sheet-visibility-20261004/receipt.json')
for rel,h in receipt['attachments'].items():assert sha(r/rel)==h,rel
parent=r/receipt['parent']['path'];assert sha(parent)==receipt['parent']['sha256']
p=load(parent)
for rel,h in p['attachments'].items():assert sha(r/rel)==h,('parent',rel)
asset=r/receipt['asset_parent']['path'];assert sha(asset)==receipt['asset_parent']['sha256']
for rel,h in load(asset)['attachments'].items():assert sha(r/rel)==h,('asset parent',rel)
for rel,h in receipt['source_inputs'].items():
 assert sha(r/p['source_snapshot']/rel)==h,rel
 if not args.recorded_source:assert sha(r/rel)==h,('current source changed',rel)
assert sha(r/receipt['binary']['path'])==receipt['binary']['sha256']
assert check(r)==receipt['checks'] and receipt['correction']['status']=='refuted'
print('資料頁續核：新48欄與舊4欄完整字模零差；缺墨點／錯值反例拒絕。')
print('雙語各12正常鍵、13狀態；24真實鍵、14完整畫格、2正常退出0；保存SHA不變。')
print('前輪裁切說法撤回，原畫格與收據維持；其餘中文畫面及原版整場對拍未宣稱完成。')
