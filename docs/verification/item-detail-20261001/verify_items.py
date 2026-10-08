"""有界Docker中重算物品明細與空格確認窄收據；--export僅保存已驗PNG。"""
from pathlib import Path
import argparse,hashlib,json,os,shutil,sys
sys.path.insert(0,'/src/tools')
from png_io import read_png
ROOT=Path('/src'); OUT=Path('/out')
PROOF=ROOT/'docs/verification/item-detail-20261001'
SELECTED=[5,8,11,14,17,20]
BEFORE='e5be211397bbb2eb11e355a77bbe490cc8a1cd8c618c049679dd97a1456bdc7a'
AFTER='6ffca0ddcaacd88b0d0bca899d4c23aad613654b8ba0c230f802dca72f929bd2'
ZIP='2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46'
CURSOR=(304,185,14,14)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text())
def diff(a,b):
 assert (a.width,a.height,b.width,b.height)==(320,200,320,200)
 ar,br=a.rgb(),b.rgb(); pts=[];pointer=0
 for n in range(64000):
  if ar[3*n:3*n+3]==br[3*n:3*n+3]:continue
  x,y=n%320,n//320
  if 304<=x<318 and 185<=y<199:pointer+=1
  else:pts.append((x,y))
 return pts,pointer
def main():
 export=argparse.ArgumentParser();export.add_argument('--export',action='store_true');args=export.parse_args()
 assert sha(ROOT/'Knights-of-legend.zip')==ZIP
 exe=[p for p in (OUT/'gui2/game').iterdir() if p.name.upper()=='KOLTOWN.EXE']
 assert len(exe)==1 and sha(exe[0])=='c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221'
 assert sha(ROOT/'workplace/battle-cursor-20261001/kol-attack-frame-linux')==BEFORE
 assert sha(OUT/'kol-items-linux')==AFTER
 source=load(OUT/'gui2/inputs.json')
 assert source['seed']==1 and source['revision']=='968ca774705853276ae55c8d1163c369481ac71f'
 assert source['normal_startup'] and not source['state_injection'] and not source['direct_entry']
 original=load(OUT/'gui2/item-calls-private.json')
 assert original['seed']==1 and original['normal_inputs'] and not original['state_injection']
 assert original['mouse_api']=='MoveMouse(638,199)' and len(original['keys'])==21
 calls=[x for x in original['calls'] if x['ida_ea']=='0x153FF']
 assert len(calls)==2 and all(x['args'][0]==8 and x['caller_ida']=='0x16C55' for x in calls)
 index=load(ROOT/'docs/re/item-detail-address-index.json')
 assert index['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
 assert index['address_space']=='IDA linear' and len(index['entries'])==6
 assert all(v['confidence']=='已證實' and v['evidence'] for v in index['entries'].values())
 annotated=load(OUT/'item-annotated-private.json')
 attached=[a for f in annotated['records'] for r in f['instructions'] for a in r['annotations']]
 attachments_in_export=sum(a['index'].endswith('/item-detail-address-index.json') for a in attached)
 assert attachments_in_export==6 and annotated['sha256']==index['sha256']
 assert len(attached)==10
 assert all(a['index_sha256']==sha(Path(a['index'])) for a in attached)
 assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64')==annotated['database_sha256']
 assert 'FAIL' not in (OUT/'tests-clean.log').read_text()
 assert 'internal/town' in (OUT/'tests-clean.log').read_text() and 'internal/ui' in (OUT/'tests-clean.log').read_text()
 checks=[];unverified=[];copies=[]
 for i in range(21):
  a=read_png(OUT/'gui2'/f'item-{i:02d}.png')
  b=read_png(OUT/'normal-after-en'/f'step-{i+6:03d}.png')
  pts,pointer=diff(a,b)
  assert pointer==67,(i,pointer)
  if i in SELECTED:
   assert not pts,(i,pts[:8])
   checks.append(dict(original_frame=i,remake_step=i+6,compare=[0,0,320,200],
    outside_cursor_pixels=0,cursor_pixels=pointer,mode8=(i in [5,8]),empty=(i not in [5,8])))
   # 對實際原版畫格拿掉一個框邊像素，必須拒絕；不測比較器自己的常數。
   raw=bytearray(b.pixels);n=(58*320+62)*b.channels;raw[n:n+3]=b'\0\0\0'
   wrong_image=type(b)(b.width,b.height,b.channels,bytes(raw))
   assert diff(a,wrong_image)==([(62,58)],67)
   ar,br=a.rgb(),b.rgb();pos=(58*320+62)*3
   assert ar[pos:pos+3]==br[pos:pos+3]==b'\xff\xff\xff'
   for src,name in [(OUT/'gui2'/f'item-{i:02d}.png',f'original-{i:02d}.png'),
                    (OUT/'normal-after-en'/f'step-{i+6:03d}.png',f'en-{i+6:03d}.png'),
                    (OUT/'normal-after-zh-Hant'/f'step-{i+6:03d}.png',f'zh-Hant-{i+6:03d}.png')]:
    copies.append((src,name))
   if i not in [5,8]:
    old=read_png(OUT/'normal-full-en'/f'step-{i+6:03d}.png')
    assert diff(a,old)[0],('舊錯位畫格沒有被拒絕',i)
  else:
   expected=0 if i in [0,1,2,3] else (1 if i==4 else 8)
   assert len(pts)==expected,(i,len(pts))
   if pts:unverified.append(dict(original_frame=i,remake_step=i+6,pixels=[list(p) for p in pts],status='未通過，未加允許區'))
 # 中文字模由既有譯表與字模PNG獨立合成，不以Go繪製結果當預期值。
 fm=load(ROOT/'assets/ui-font/manifest.json')['fonts'][0]
 font=read_png(ROOT/'assets/ui-font'/fm['file']);assert sha(ROOT/'assets/ui-font'/fm['file'])==fm['png_sha256']
 gs={g['character']:g for g in fm['glyphs']}
 bank=load(ROOT/'docs/i18n/ui-zh-Hant.json')['banks']['KOLDATA']
 glyph_checks=[]
 for step,label in [(17,145),(20,146),(23,147),(26,146)]:
  text=bank['202']['zh']+bank[str(label)]['zh']+bank['203']['zh']+'！'
  width=sum(gs[c]['advance'] if c in gs else fm['advance'] for c in text)
  assert 192+width<257*3 and 228+18<125*3
  want=bytearray(width*18*3);pen=0;ink=[]
  for c in text:
   if c not in gs:assert c.isspace();pen+=fm['advance'];continue
   g=gs[c];sx,sy=g['sheet_column']*16,g['sheet_row']*18
   for y in range(18):
    for x in range(16):
     alpha=font.pixels[((sy+y)*font.width+sx+x)*4+3]
     assert alpha in [0,255]
     if alpha:
      n=(y*width+pen+x)*3;want[n:n+3]=b'\xff\xff\xff';ink.append(n)
   pen+=g['advance']
  actual=read_png(OUT/'normal-after-zh-Hant'/f'step-{step:03d}.png').crop(192,228,width,18).rgb()
  assert actual==want,(step,'中文字模不符')
  wrong=want.copy();wrong[ink[0]:ink[0]+3]=b'\0\0\0';assert actual!=wrong
  glyph_checks.append(dict(step=step,text=text,rect=[192,228,width,18],pixel_differences=0,missing_ink_pixel_rejected=True))
 # 正常記錄欄位沒有資料頁暫態，僅作該欄位範圍的一致性護欄。
 states=[load(OUT/f'{phase}-{lang}/trace.json')['steps'] for phase in ['normal-full','normal-after'] for lang in ['en','zh-Hant']]
 assert all(len(s)==27 and s==states[0] for s in states)
 saves=[];save_meta=load(OUT/'save-comparison.json')
 for run in save_meta['runs']:
  p=OUT/f"save-{run['phase']}-{run['lang']}"/'save.json';assert sha(p)==run['save_sha256']
  d=load(p);d.pop('saved_at');saves.append(d)
 assert len(saves)==4 and all(d==saves[0] for d in saves)
 old=load(ROOT/'docs/verification/battle-target-cursor-20261001/attack-row-receipt.json')
 inputs=[dict(path=r['path'],sha256=sha(ROOT/r['path'])) for r in old['product_inputs']]
 for p in ['docs/re/item-detail-address-index.json','docs/i18n/ui-zh-Hant.json','tools/png_io.py']:
  if p not in [r['path'] for r in inputs]:inputs.append(dict(path=p,sha256=sha(ROOT/p)))
 for name in ['original.sh','prepare.py','remake.py','remake_after.py','save_checks.py','tests-clean.log','save-comparison.json']:
  copies.append((OUT/name,name))
 copies.extend([(OUT/'normal-after-en/command.json','command-en.json'),(OUT/'normal-after-zh-Hant/command.json','command-zh-Hant.json')])
 if args.export:
  for src,name in copies:
   assert src.stat().st_uid==os.getuid() and src.stat().st_gid==os.getgid()
   shutil.copyfile(src,PROOF/name)
 else:
  for src,name in copies:assert sha(src)==sha(PROOF/name)
 artifacts=[dict(path=name,sha256=sha(PROOF/name),uid=(PROOF/name).stat().st_uid,gid=(PROOF/name).stat().st_gid) for src,name in copies]
 receipt=dict(schema='kol-item-detail-dosgolem-v1',date='2026-10-01',status='窄契約通過，完整清單仍有差異',
  before_binary=dict(path='workplace/battle-cursor-20261001/kol-attack-frame-linux',sha256=BEFORE),
  binary=dict(path='workplace/item-detail-20261001/kol-items-linux',sha256=AFTER),source_zip_sha256=ZIP,
  source_executable_sha256='c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221',
  source_inputs_private=dict(path='workplace/item-detail-20261001/gui2/inputs.json',sha256=sha(OUT/'gui2/inputs.json')),
  original_revision=source['revision'],original_seed=1,remake_seed=1,original_keys=original['keys'],
  mouse_event='MoveMouse(638,199)，原版自身夾限至可見(304,185)',allowed_cursor_rect=list(CURSOR),
  checks=checks,glyph_checks=glyph_checks,unverified_list_frames=unverified,
  normal_record_states_each=27,normal_record_fields_equal=True,normal_trace_limit='軌跡未記資料頁暫態；不單靠軌跡證明確認或移列，使用畫格及輸入測試',
  save_comparison=save_meta,product_inputs=inputs,attachments=artifacts,
  semantic_index_sha256=sha(ROOT/'docs/re/item-detail-address-index.json'),semantic_annotations_merged=attachments_in_export,
  annotated_private=dict(path='workplace/item-detail-20261001/item-annotated-private.json',sha256=sha(OUT/'item-annotated-private.json')),
  tests=dict(command='GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1',log='tests-clean.log',circle_frames=104),
  rights='PNG僅依2026-09-01 private GitHub特例保存；沒有公開散布權，原版payload／角色存檔只留本機',
  limits=['只核胸甲／腿甲模式8與BODY／HANDS／FEET空格頁；未驗其他物品類型、空口袋正常原版畫格或所有動作',
   '原版沒有中文畫面；中文驗既有三倍座標、四框字模及正常語意／保存',
   '清單1／8像素差異保留，不稱完整清單或#64完成；舊DOSBox收據維持輔助分類'],
  verifier_sha256=sha(Path(__file__)))
 if args.export:(PROOF/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
 else:assert receipt==load(PROOF/'receipt.json')
 assert all(r['uid']==1000 and r['gid']==1000 for r in artifacts)
 print(json.dumps(dict(status=receipt['status'],attachments=len(artifacts),original_frames=6,remake_frames=12,
  full_canvas_checks=6,glyph_checks=4,current_inputs=len(inputs),unverified_list_frames=len(unverified),save_equal=True),ensure_ascii=False))
if __name__=='__main__':main()
