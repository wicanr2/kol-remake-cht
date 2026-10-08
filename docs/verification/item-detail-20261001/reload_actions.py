"""從六份正常住宿存檔重新啟動；沿標題與資料頁查看後再次住宿。"""
from pathlib import Path
import json,hashlib,os,subprocess
root,out=Path('/src'),Path('/out')
assert out.stat().st_uid==os.getuid() and out.stat().st_gid==os.getgid()
binary=out/'kol-items-actions-final-linux';rows=[]
stages=[(15,10),(28,9),(42,11),(56,10),(72,10),(88,10)]
def goods(c):
 v=c['worn']+[c['in_hand']]+(c['backpack'] or [])
 return v+[None]*(17-len(v))
oracle=json.loads((out/'gui-actions1/item-calls-private.json').read_text())['states']
offsets=[256+16*i for i in range(10)]+[48]+[160+16*i for i in range(6)]
for frame,slot in stages:
 for lang in ['en','zh-Hant']:
  source=out/f'save-actions-{frame:02d}-{lang}/save.json'
  data=json.loads(source.read_text());expected=data['characters'][0]
  raw=bytes.fromhex(oracle[frame]['character_record_hex'])
  wanted=[raw[o:o+16].hex() if raw[o] else None for o in offsets]
  assert [g['record'] if g else None for g in goods(expected)]==wanted
  dest=out/f'reload-actions-{frame:02d}-{lang}';assert not dest.exists();dest.mkdir()
  keys=['SPACE','SPACE','6','RET','SPACE','>','RET','>','RET']+['D']*slot+['RET']
  detail_step=len(keys)
  keys+=['RET','>','RET','RET']+['>']*4+['RET']*3
  saved=dest/'save.json'
  cmd=[str(binary),'-scene','flow','-load',str(source),'-save',str(saved),'-seed','1','-lang',lang,
   '-flow-script',' '.join(keys),'-shot-dir',str(dest),'-shot-steps',f'{detail_step},{len(keys)}','-trace',str(dest/'trace.json')]
  (dest/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
  run=subprocess.run(cmd,cwd=root,capture_output=True,text=True,timeout=180)
  (dest/'run.log').write_text(run.stdout+run.stderr)
  assert run.returncode==0,(frame,lang,run.stderr[-1000:])
  assert saved.is_file(),(frame,lang,'重載後正常住宿沒有存檔')
  actual=json.loads(saved.read_text())['characters'][0]
  assert goods(actual)==goods(expected),(frame,lang,'重載改變物品')
  assert [g['record'] if g else None for g in goods(actual)]==wanted
  rows.append(dict(frame=frame,lang=lang,slot=slot,keys=len(keys),detail_step=detail_step,
   original_nonempty_slots=[i for i,g in enumerate(wanted) if g],
   original_normalized_items_equal=True,loaded_typed_items_equal=True,
   source_save_sha256=hashlib.sha256(source.read_bytes()).hexdigest(),
   save_sha256=hashlib.sha256(saved.read_bytes()).hexdigest()))
  print(frame,lang,'重載與原版十七欄一致',flush=True)
for frame,slot in stages:
 for prefix in ['save','reload']:
  a=json.loads((out/f'{prefix}-actions-{frame:02d}-en/save.json').read_text());a.pop('saved_at')
  b=json.loads((out/f'{prefix}-actions-{frame:02d}-zh-Hant/save.json').read_text());b.pop('saved_at')
  assert a==b,(frame,prefix,'語言影響持久資料')
(out/'actions-persistence-comparison.json').write_text(json.dumps(dict(runs=rows,excluded_fields=['saved_at'],
 languages_equal=True,normal_startup=True,state_injection=False,
 limits='只對拍十七物品欄；住宿後時間、金幣與原版同期狀態未對拍。'),ensure_ascii=False,indent=2)+'\n')
