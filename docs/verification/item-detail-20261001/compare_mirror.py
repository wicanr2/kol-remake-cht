"""核對同一104原版畫格，逐點保留剩餘清單差異。"""
from pathlib import Path
import json,sys
sys.path.insert(0,'/src/tools')
from png_io import read_png
p=Path('/out');rows=[]
for i in range(104):
 a=read_png(p/'gui-actions1'/f'item-{i:02d}.png').rgb();b=read_png(p/'normal-mirror-en'/f'step-{i+6:03d}.png').rgb();pts=[];cursor=0
 for n in range(64000):
  if a[n*3:n*3+3]==b[n*3:n*3+3]:continue
  x,y=n%320,n//320
  if 304<=x<318 and 185<=y<199:cursor+=1
  else:pts.append([x,y])
 rows.append(dict(frame=i,step=i+6,pixels=pts,cursor=cursor))
(p/'mirror-comparison-probe.json').write_text(json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
print(json.dumps(dict(zero=sum(not r['pixels'] for r in rows),pending=sum(bool(r['pixels']) for r in rows),
 residual_counts=sorted({len(r['pixels']) for r in rows if r['pixels']}),
 outside_x15=[r for r in rows if any(x!=15 for x,y in r['pixels'])]),ensure_ascii=False))
