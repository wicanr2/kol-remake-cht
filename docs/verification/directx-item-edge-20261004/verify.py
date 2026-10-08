from pathlib import Path
import copy,hashlib,json,sys
r=Path(__file__).resolve().parents[3]
receipt=json.loads(Path(__file__).with_name('receipt.json').read_text())
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
for path,h in receipt['attachments'].items():assert sha(r/path)==h,path
o=r/'workplace/directx-item-edge-20261004'
sys.path.insert(0,str(o/'controls'))
from png_io import read_png
for name,h in receipt['source_inputs'].items():assert sha(o/'source-final'/name)==h,name
parent=r/receipt['original_quest_parent']['receipt']
assert sha(parent)==receipt['original_quest_parent']['sha256']
old=json.loads(parent.read_text())
for group in ['attachments','original_inputs']:
 for path,h in old[group].items():assert sha(r/path)==h,path
for name,h in old['production_inputs'].items():assert sha(o/'before-source'/name)==h,name
assert receipt['checks']['changed']==['internal/ui/itempage_edge.go']
for platform in ['linux','windows']:
 for lang in ['en','zh-Hant']:
  for stage in ['return','reload']:
   a=json.loads((o/f'regression-{platform}/{stage}-{lang}/trace.json').read_text())
   b=json.loads((o/f'reference/{stage}-{lang}-trace.json').read_text());assert a==b
for lang in ['en','zh-Hant']:
 actual=json.loads((o/f'full-windows/{lang}/trace.json').read_text())
 expected=json.loads((o/f'reference/full-{lang}-trace.json').read_text())
 assert actual==expected and len(actual['steps'])==4077
 a=json.loads((o/f'full-windows/{lang}-save.json').read_text())
 b=json.loads((o/f'reference/full-{lang}-save.json').read_text())
 a.pop('saved_at');b.pop('saved_at');assert a==b and a['schema']=='kol-save-v3'
 bad=copy.deepcopy(a);bad['speaker']+=1;assert bad!=b
 for directory in ['gui-windows','gui-own-save']:
  root=o/directory/lang;data=json.loads((root/'inputs.json').read_text())
  assert data['tokens']==(o/f'controls/reload-{lang}.txt').read_text().split() and len(data['records'])==26
  assert int((root/'process-exit-status').read_text())==0
  assert json.loads((root/'window-close.json').read_text())['window_destroyed']
  for row in data['records']:
   if 'post_draw_equal' not in row:continue
   assert row['post_draw_equal']
   p=o/f'expected/{lang}/step-{row["step"]:03}.png';img=read_png(p);rgb=img.rgb()
   if img.width==320:rgb=b''.join(b''.join(rgb[(y*320+x)*3:(y*320+x)*3+3]*3 for x in range(320))*3 for y in range(200))
   result=read_png(root/f'step-{row["step"]:03}.png').rgb();assert result==rgb
   changed=bytearray(result);changed[0]^=1;assert bytes(changed)!=rgb
assert receipt['checks']['real_windows_execution'] is False
assert receipt['checks']['macos_execution'] is False
assert 'E5017' in (o/'before-directx/engine.log').read_text()
assert 'node->type == HLSL_IR_CONSTANT' in (o/'mod-only-regression-windows/reload-en/engine.log').read_text()
print('PASS immutable shader fix, normal DirectX paths, own-save reload, PNG and negative controls; no real-machine claim')
