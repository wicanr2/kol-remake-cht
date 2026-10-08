"""由有界Docker重算正常物品動作窄收據；清單差異一律列未通過。"""
from pathlib import Path
import argparse,hashlib,json,os,shutil,sys
sys.path.insert(0,'/src/tools')
from png_io import read_png
ROOT=Path('/src');OUT=Path('/out');PROOF=ROOT/'docs/verification/item-detail-20261001'
BEFORE='351464d1bbff6848defd94e2346b7c4c4741e7299f0842abe340e855f6b10ec2'
AFTER='c1e3b410380c338d5d1e6c799daee00a75101ea688dfa836a9c94afdc4dd02e7'
CHECKED=[13,14,26,27,38,39,40,41,54,55,67,68,69,70,71,83,84,85,86,87,99,102,103]
OFFSETS=[256+16*i for i in range(10)]+[48]+[160+16*i for i in range(6)]
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text())
def diff(a,b):
 assert (a.width,a.height,b.width,b.height)==(320,200,320,200)
 aa,bb=a.rgb(),b.rgb();points=[];cursor=0
 for n in range(64000):
  if aa[n*3:n*3+3]==bb[n*3:n*3+3]:continue
  x,y=n%320,n//320
  if 304<=x<318 and 185<=y<199:cursor+=1
  else:points.append([x,y])
 return points,cursor
def goods(c):
 v=c['worn']+[c['in_hand']]+(c['backpack'] or [])
 assert len(v)<=17
 return v+[None]*(17-len(v))
def source_items(state):
 raw=bytes.fromhex(state['character_record_hex']);assert len(raw)==512 and raw[:9]==b'Theodoric'
 return [raw[o:o+16].hex() if raw[o] else None for o in OFFSETS]
def main():
 ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
 historical=load(PROOF/'all-slots-receipt.json')
 assert sha(PROOF/'all-slots-receipt.json')=='4b21c2f87edf61a58804925f2e5ecb9fde2f4e3cdcd7b044e933a5ce547f1943'
 for name in ['receipt.json','all-slots-receipt.json']:
  for a in load(PROOF/name)['attachments']:assert sha(PROOF/a['path'])==a['sha256']
 assert sha(OUT/'kol-items-name-linux')==BEFORE and sha(OUT/'kol-items-actions-final-linux')==AFTER
 assert sha(ROOT/'Knights-of-legend.zip')==historical['source_zip_sha256']
 source=load(OUT/'gui-actions1/inputs.json');oracle=load(OUT/'gui-actions1/item-calls-private.json')
 assert source['revision']==historical['original_revision'] and source['seed']==oracle['seed']==1
 assert source['normal_startup'] and oracle['normal_inputs']
 assert not source['state_injection'] and not source['direct_entry'] and not oracle['state_injection'] and not oracle['direct_entry']
 keys=load(OUT/'actions-keys.json');assert len(keys)==104 and oracle['keys']==keys and source['keys'].split(',')==keys
 exe=next(p for p in (OUT/'gui-actions1/game').iterdir() if p.name.upper()=='KOLTOWN.EXE')
 assert sha(exe)==historical['source_executable_sha256']
 log=(OUT/'gui-actions1/run.log').read_text();assert '--- PASS: TestNormalBookEntry' in log and 'FAIL' not in log
 states=oracle['states'];assert len(states)==104 and [r['frame'] for r in states]==list(range(104))
 assert all(r['graphics_mode']==13 and r['character_dgroup_offset']=='0xBB88' for r in states)
 observed=[]
 stages=[(15,[0,1,2,10]),(28,[0,1,2,9]),(42,[0,1,2,11]),(56,[0,1,2,10]),(72,[0,1,2,10]),(88,[0,1,2])]
 for frame,slots in stages:
  values=source_items(states[frame]);assert [i for i,v in enumerate(values) if v]==slots
  assert states[frame]['selected_row']==0 and states[frame]['icon_count']==2
  observed.append(dict(frame=frame,selected_row=0,nonempty_slots=slots))
 assert source_items(states[70])==source_items(states[71])==source_items(states[72])
 before_yes=source_items(states[86]);after_yes=source_items(states[87])
 assert before_yes[10] and after_yes[10] is None and before_yes[:10]+before_yes[11:]==after_yes[:10]+after_yes[11:]
 assert bytes.fromhex(states[87]['character_record_hex'])[48:64]==bytes(16)
 for f,icons,selected in [(70,2,0),(71,2,1),(86,2,0),(87,1,0),(99,1,0)]:
  assert states[f]['icon_count']==icons and states[f]['icon_selected']==selected and states[f]['selected_row']==10
 calls=[c for c in oracle['calls'] if c['ida_ea']=='0x18CD5']
 assert [c['args'][:2] for c in calls]==[[400,48],[48,400],[400,160],[160,48]]
 annotations=load(OUT/'item-actions-annotated-private.json')
 attached=[a for f in annotations['records'] for r in f['instructions'] for a in r['annotations']]
 assert len(attached)==13 and sum(a['index'].endswith('/item-action-address-index.json') for a in attached)==8
 for a in attached:assert sha(Path(a['index']))==a['index_sha256'] and a['confidence']=='已證實' and a['evidence']
 assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64')==annotations['database_sha256']
 mapping={'.':'>','Return':'RET','Down':'D','Up':'U'}
 equivalent=['SPACE','SPACE','6','RET','SPACE']+[mapping[k] for k in keys]
 traces=[]
 for lang in ['en','zh-Hant']:
  cmd=load(OUT/f'normal-actions-after-{lang}/command.json')
  assert cmd[cmd.index('-flow-script')+1].split()==equivalent and len(equivalent)==109
  trace=load(OUT/f'normal-actions-after-{lang}/trace.json')['steps'];assert len(trace)==110
  assert trace[-1]['scene']=='flow:town' and trace[-1]['town']['mode']==2
  traces.append(trace)
 assert traces[0]==traces[1]
 checks=[];pending=[];copies=[]
 for i in range(104):
  a=read_png(OUT/'gui-actions1'/f'item-{i:02d}.png');b=read_png(OUT/'normal-actions-after-en'/f'step-{i+6:03d}.png')
  pts,cursor=diff(a,b);assert cursor==67
  if i in CHECKED or i<4:
   assert not pts,(i,pts[:10])
   if i in CHECKED:
    checks.append(dict(original_frame=i,remake_step=i+6,compare=[0,0,320,200],outside_cursor_pixels=0,cursor_pixels=67))
    for src,name in [(OUT/'gui-actions1'/f'item-{i:02d}.png',f'actions-original-{i:02d}.png'),
      (OUT/'normal-actions-after-en'/f'step-{i+6:03d}.png',f'actions-en-{i+6:03d}.png'),
      (OUT/'normal-actions-after-zh-Hant'/f'step-{i+6:03d}.png',f'actions-zh-Hant-{i+6:03d}.png')]:copies.append((src,name))
  elif pts:
   pending.append(dict(original_frame=i,remake_step=i+6,pixels=len(pts),
    bounding_rect=[min(x for x,y in pts),min(y for x,y in pts),max(x for x,y in pts)+1,max(y for x,y in pts)+1],
    coordinates_sha256=hashlib.sha256(json.dumps(pts,separators=(',',':')).encode()).hexdigest(),status='未通過，未加入允許區'))
 assert len(checks)==23 and len(pending)==76
 for f in CHECKED:
  if f in [102,103]:continue  # 資料頁與旅館不含物品明細框。
  b=read_png(OUT/'normal-actions-after-en'/f'step-{f+6:03d}.png')
  raw=bytearray(b.pixels);n=(58*320+62)*b.channels;assert raw[n:n+3]==b'\xff\xff\xff';raw[n:n+3]=b'\0\0\0'
  assert diff(read_png(OUT/'gui-actions1'/f'item-{f:02d}.png'),type(b)(b.width,b.height,b.channels,bytes(raw)))[0]==[[62,58]]
 # 前版在同一正常鍵序列缺少提問；這個反例應被整頁比較拒絕。
 old=read_png(OUT/'normal-actions-before-en/step-076.png')
 bad,cursor=diff(read_png(OUT/'gui-actions1/item-70.png'),old);assert bad and cursor==67
 copies.append((OUT/'normal-actions-before-en/step-076.png','actions-before-en-076.png'))
 fm=load(ROOT/'assets/ui-font/manifest.json')['fonts'][0];font=read_png(ROOT/'assets/ui-font'/fm['file'])
 assert sha(ROOT/'assets/ui-font'/fm['file'])==fm['png_sha256']
 glyphs={g['character']:g for g in fm['glyphs']};bank=load(ROOT/'docs/i18n/ui-zh-Hant.json')['banks']['KOLDATA'];glyph_checks=[]
 for frame,record in [(70,'209'),(71,'209'),(86,'209'),(87,'210')]:
  text=bank[record]['zh'];width=sum(glyphs[c]['advance'] for c in text);want=bytearray(width*18*3);pen=0;ink=[]
  for c in text:
   g=glyphs[c];sx,sy=g['sheet_column']*16,g['sheet_row']*18
   for y in range(18):
    for x in range(16):
     alpha=font.pixels[((sy+y)*font.width+sx+x)*4+3];assert alpha in [0,255]
     if alpha:
      n=(y*width+pen+x)*3;want[n:n+3]=b'\xff\xff\xff';ink.append(n)
   pen+=g['advance']
  assert 192+width<266*3 and 423+18<152*3
  actual=read_png(OUT/'normal-actions-after-zh-Hant'/f'step-{frame+6:03d}.png').crop(192,423,width,18).rgb()
  assert actual==want,(frame,'繁中字模不符')
  wrong=want.copy();wrong[ink[0]:ink[0]+3]=b'\0\0\0';assert actual!=wrong
  glyph_checks.append(dict(original_frame=frame,record=record,text=text,rect=[192,423,width,18],pixel_differences=0))
 persistence=load(OUT/'actions-persistence-comparison.json');assert persistence['languages_equal'] and persistence['normal_startup'] and not persistence['state_injection']
 assert len(persistence['runs'])==12 and persistence['excluded_fields']==['saved_at']
 frame_map={15:26,28:38,42:54,56:67,72:83,88:99}
 for r in persistence['runs']:
  frame,lang=r['frame'],r['lang'];wanted=source_items(states[frame]);expected=None
  for phase,key in [('save','source_save_sha256'),('reload','save_sha256')]:
   p=OUT/f'{phase}-actions-{frame:02d}-{lang}/save.json';assert sha(p)==r[key]
   c=load(p)['characters'][0];actual=goods(c)
   assert [g['record'] if g else None for g in actual]==wanted
   if expected is None:expected=actual
   else:assert actual==expected
  reloaded=OUT/f'reload-actions-{frame:02d}-{lang}'/f"step-{r['detail_step']:03d}.png"
  assert read_png(reloaded).rgb()==read_png(OUT/f'normal-actions-after-{lang}'/f'step-{frame_map[frame]+6:03d}.png').rgb()
  copies.append((reloaded,f'actions-reload-{frame:02d}-{lang}.png'))
 for frame,slots in stages:
  for phase in ['save','reload']:
   a=load(OUT/f'{phase}-actions-{frame:02d}-en/save.json');a.pop('saved_at')
   b=load(OUT/f'{phase}-actions-{frame:02d}-zh-Hant/save.json');b.pop('saved_at');assert a==b
 tests=(OUT/'tests-actions-final.log').read_text();assert 'FAIL' not in tests
 assert all(x in tests for x in ['internal/town','internal/ui','cmd/kol'])
 for name in ['original_actions.sh','prepare_actions.py','actions-keys.json','remake_actions_after.py','save_actions.py','reload_actions.py',
   'actions-persistence-comparison.json','tests-actions-final.log','ida_items_actions.py','ida_item_actions_annotated.py']:
  copies.append((OUT/name,name))
 for lang in ['en','zh-Hant']:
  copies.append((OUT/f'normal-actions-after-{lang}/command.json',f'actions-command-{lang}.json'))
 inputs=[dict(path=r['path'],sha256=sha(ROOT/r['path'])) for r in historical['product_inputs']]
 inputs.append(dict(path='docs/re/item-action-address-index.json',sha256=sha(ROOT/'docs/re/item-action-address-index.json')))
 if args.export:
  assert PROOF.stat().st_uid==os.getuid() and PROOF.stat().st_gid==os.getgid()
  for src,name in copies:
   assert src.stat().st_uid==os.getuid() and src.stat().st_gid==os.getgid() and not (PROOF/name).exists(),name
  for src,name in copies:shutil.copyfile(src,PROOF/name)
 else:
  for src,name in copies:assert sha(src)==sha(PROOF/name)
 receipt=dict(schema='kol-item-actions-dosgolem-v1',date='2026-10-01',status='搬移／丟棄明細與保存切片通過；完整清單未通過',
  binary=dict(path='workplace/item-detail-20261001/kol-items-actions-final-linux',sha256=AFTER),before_binary_sha256=BEFORE,
  source_zip_sha256=historical['source_zip_sha256'],source_executable_sha256=historical['source_executable_sha256'],
  original_revision=source['revision'],original_seed=1,remake_seed=1,seed_setup=historical['seed_setup'],
  source_inputs_private=dict(path='workplace/item-detail-20261001/gui-actions1/inputs.json',sha256=sha(OUT/'gui-actions1/inputs.json')),
  readonly_actor_private=dict(path='workplace/item-detail-20261001/gui-actions1/item-calls-private.json',sha256=sha(OUT/'gui-actions1/item-calls-private.json'),
    bytes=512,dgroup_offset='0xBB88',item_offsets=OFFSETS,empty_rule='首byte為0；未用尾端不要求歸零'),
  original_keys=keys,remake_keys_count=109,normal_record_states_each=110,normal_record_fields_equal=True,
  normal_startup=True,state_injection=False,direct_entry=False,allowed_cursor_rect=[304,185,14,14],checks=checks,glyph_checks=glyph_checks,
  original_item_stages=observed,move_pairs=[c['args'][:2] for c in calls],drop_no_preserves_items=True,drop_yes_clears_16_bytes=True,
  unverified_list_frames=pending,old_query_rejected=dict(original_frame=70,remake_step=76,outside_cursor_pixels=len(bad)),
  persistence=persistence,product_inputs=inputs,historical_receipt_sha256=sha(PROOF/'all-slots-receipt.json'),historical_attachments_preserved=True,
  annotations=dict(total=13,new=8,index_sha256=sha(ROOT/'docs/re/item-action-address-index.json'),private_payload_sha256=sha(OUT/'item-actions-annotated-private.json')),
  tests=dict(command='GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1',log='tests-actions-final.log',circle_frames=104),
  attachments=[dict(path=name,sha256=sha(PROOF/name),uid=(PROOF/name).stat().st_uid,gid=(PROOF/name).stat().st_gid) for src,name in copies],
  rights='PNG僅依既有private GitHub特例保存；原版payload與完整存檔僅本機，沒有公開散布權。',
  limits=['只驗Theodoric巨劍成功搬移／NO／YES及既有空武器，其他物品與動作未全驗。',
   '76清單畫格仍有差異，搬移後另有692像素框形殘留；未豁免，SPEC-031父規格及#64仍開放。',
   '原版無繁中oracle；繁中只驗字模、安全矩形、狀態及保存，不冒稱原版像素一致。',
   '住宿與重載只對拍十七物品欄；未宣稱原版完整角色、時間、金幣或亂數內部一致。'],
  verifier_sha256=sha(Path(__file__)))
 assert all(a['uid']==1000 and a['gid']==1000 for a in receipt['attachments'])
 if args.export:
  assert not (PROOF/'actions-receipt.json').exists()
  (PROOF/'actions-receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
 else:assert receipt==load(PROOF/'actions-receipt.json')
 print(json.dumps(dict(status=receipt['status'],checks=len(checks),glyph_checks=len(glyph_checks),pending=len(pending),attachments=len(copies),
  normal_saves=12,normal_reloads=12,inputs=len(inputs)),ensure_ascii=False))
if __name__=='__main__':main()
