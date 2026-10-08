"""有界Docker重算十七物品欄位窄收據；不豁免清單色差。"""
from pathlib import Path
import argparse,hashlib,json,os,re,shutil,sys
sys.path.insert(0,'/src/tools')
from png_io import read_png
ROOT=Path('/src');OUT=Path('/out');PROOF=ROOT/'docs/verification/item-detail-20261001'
BEFORE='6ffca0ddcaacd88b0d0bca899d4c23aad613654b8ba0c230f802dca72f929bd2'
AFTER='351464d1bbff6848defd94e2346b7c4c4741e7299f0842abe340e855f6b10ec2'
OLD_RECEIPT='3d498b4f902d50e6ff2e5bd3b1236cdd7c5175508742fa494dc10e3d4d2a3120'
SELECTED=[(4+3*r,r) for r in range(17)]+[(56,16)]
EMPTY=set(range(3,9))|set(range(10,17))
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text())
def diff(a,b):
 assert (a.width,a.height,b.width,b.height)==(320,200,320,200)
 ar,br=a.rgb(),b.rgb();pts=[];pointer=0
 for n in range(64000):
  if ar[n*3:n*3+3]==br[n*3:n*3+3]:continue
  x,y=n%320,n//320
  if 304<=x<318 and 185<=y<199:pointer+=1
  else:pts.append([x,y])
 return pts,pointer
def main():
 ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
 assert sha(PROOF/'receipt.json')==OLD_RECEIPT
 old=load(PROOF/'receipt.json')
 for a in old['attachments']:assert sha(PROOF/a['path'])==a['sha256']
 assert sha(OUT/'kol-items-linux')==BEFORE and sha(OUT/'kol-items-name-linux')==AFTER
 assert sha(ROOT/'Knights-of-legend.zip')==old['source_zip_sha256']
 source=load(OUT/'gui3/inputs.json');oracle=load(OUT/'gui3/item-calls-private.json')
 assert source['revision']==old['original_revision'] and source['seed']==oracle['seed']==1
 assert source['normal_startup'] and oracle['normal_inputs']
 assert not source['state_injection'] and not oracle['state_injection'] and not oracle['direct_entry']
 keys=['.','Return','.','Return']
 for row in range(17):
  keys+=['Return','Return']
  if row<16:keys+=['Down']
 keys+=['Down','Up','Return','Return','.','Return','Return']
 assert len(keys)==61 and oracle['keys']==keys==load(OUT/'all-keys.json')
 assert source['keys'].split(',')==keys and oracle['mouse_api']=='MoveMouse(638,199)'
 exe=next(p for p in (OUT/'gui3/game').iterdir() if p.name.upper()=='KOLTOWN.EXE')
 assert sha(exe)==old['source_executable_sha256']
 calls=[c for c in oracle['calls'] if c['ida_ea']=='0x153FF']
 assert len(calls)==4 and all(c['args'][0]==8 and c['caller_ida']=='0x16C55' for c in calls)
 log=(OUT/'gui3/run.log').read_text();assert '--- PASS: TestNormalBookEntry' in log and 'FAIL' not in log
 pairs=re.findall(r'DS:BB86\(word_39A56\)=(\d+) 圖示數DS:6C=(\d+) 圖示選取DS:6A=(\d+) 圖形模式DS:4E=(\d+)\n[^\n]*物品頁 step=(\d+)',log)
 assert len(pairs)==61
 ds={int(s):dict(row=int(r),icons=int(c),icon_selected=int(i),mode=int(m)) for r,c,i,m,s in pairs}
 for frame,row in SELECTED:
  assert ds[frame]['row']==row and ds[frame]['mode']==13
  assert ds[frame]['icons']==1 if row in EMPTY else ds[frame]['icons']>1
 index=load(ROOT/'docs/re/item-name-address-index.json')
 assert index['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
 annotated=load(OUT/'item-name-annotated-private.json')
 attached=[a for f in annotated['records'] for r in f['instructions'] for a in r['annotations']]
 assert len(attached)==14 and sum(a['index'].endswith('/item-name-address-index.json') for a in attached)==4
 for a in attached:
  assert a['index_sha256']==sha(Path(a['index'])) and a['confidence']=='已證實' and a['evidence']
 assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64')==annotated['database_sha256']
 checks=[];pending=[];copies=[]
 selected={f:r for f,r in SELECTED}
 for i in range(61):
  a=read_png(OUT/'gui3'/f'item-{i:02d}.png')
  b=read_png(OUT/'normal-all-after-en'/f'step-{i+6:03d}.png')
  pts,pointer=diff(a,b);assert pointer==67
  if i in selected or i in [0,1,2,3,59,60]:
   assert not pts,(i,pts[:12])
   if i in selected or i in [59,60]:
    checks.append(dict(original_frame=i,remake_step=i+6,row=selected.get(i),empty=selected.get(i) in EMPTY,
     compare=[0,0,320,200],outside_cursor_pixels=0,cursor_pixels=67))
    for src,name in [(OUT/'gui3'/f'item-{i:02d}.png',f'all-original-{i:02d}.png'),
      (OUT/'normal-all-after-en'/f'step-{i+6:03d}.png',f'all-en-{i+6:03d}.png'),
      (OUT/'normal-all-after-zh-Hant'/f'step-{i+6:03d}.png',f'all-zh-Hant-{i+6:03d}.png')]:copies.append((src,name))
   if i in selected:
    pos=(58*320+62)*b.channels;raw=bytearray(b.pixels)
    assert raw[pos:pos+3]==b'\xff\xff\xff';raw[pos:pos+3]=b'\0\0\0'
    assert diff(a,type(b)(b.width,b.height,b.channels,bytes(raw)))==([[62,58]],67)
  else:
   assert len(pts) in [8,9] and all(x==15 for x,y in pts),(i,pts[:12])
   pending.append(dict(original_frame=i,remake_step=i+6,pixels=pts,status='未通過，未加入允許區'))
 assert len(pending)==37 and len(checks)==20
 old_weapon=read_png(OUT/'normal-all-en/step-037.png');source_weapon=read_png(OUT/'gui3/item-31.png')
 bad,pointer=diff(source_weapon,old_weapon)
 assert len(bad)==220 and all(120<=x<=206 and 66<=y<=72 for x,y in bad) and pointer==67
 copies.append((OUT/'normal-all-en/step-037.png','all-before-en-037.png'))
 fm=load(ROOT/'assets/ui-font/manifest.json')['fonts'][0]
 font=read_png(ROOT/'assets/ui-font'/fm['file']);assert sha(ROOT/'assets/ui-font'/fm['file'])==fm['png_sha256']
 gs={g['character']:g for g in fm['glyphs']};bank=load(ROOT/'docs/i18n/ui-zh-Hant.json')['banks']['KOLDATA']
 glyph_checks=[]
 for frame,row in SELECTED:
  if row not in EMPTY:continue
  text=bank['204']['zh'].strip() if row>10 else bank['202']['zh']+bank[str(142+row)]['zh']+bank['203']['zh']+'！'
  width=sum(gs[c]['advance'] if c in gs else fm['advance'] for c in text)
  assert 192+width<257*3 and 228+18<125*3
  want=bytearray(width*18*3);pen=0;ink=[]
  for c in text:
   if c not in gs:assert c.isspace();pen+=fm['advance'];continue
   g=gs[c];sx,sy=g['sheet_column']*16,g['sheet_row']*18
   for y in range(18):
    for x in range(16):
     alpha=font.pixels[((sy+y)*font.width+sx+x)*4+3];assert alpha in [0,255]
     if alpha:
      n=(y*width+pen+x)*3;want[n:n+3]=b'\xff\xff\xff';ink.append(n)
   pen+=g['advance']
  actual=read_png(OUT/'normal-all-after-zh-Hant'/f'step-{frame+6:03d}.png').crop(192,228,width,18).rgb()
  assert actual==want,(frame,row,'中文字模不符')
  wrong=want.copy();wrong[ink[0]:ink[0]+3]=b'\0\0\0';assert actual!=wrong
  glyph_checks.append(dict(frame=frame,row=row,text=text,rect=[192,228,width,18],pixel_differences=0))
 assert len(glyph_checks)==14
 for i in range(67):
  for lang in ['en','zh-Hant']:
   a=read_png(OUT/f'normal-all-{lang}'/f'step-{i:03d}.png')
   b=read_png(OUT/f'normal-all-after-{lang}'/f'step-{i:03d}.png')
   if i!=37:assert a.rgb()==b.rgb(),(i,lang,'名稱以外畫格改動')
   elif lang=='zh-Hant':
    aa,bb=a.rgb(),b.rgb();count=0
    for n in range(a.width*a.height):
     if aa[n*3:n*3+3]!=bb[n*3:n*3+3]:
      assert 360<=n%a.width<416 and 198<=n//a.width<216;count+=1
    assert count>0
 states=[load(OUT/f'{phase}-{lang}/trace.json')['steps'] for phase in ['normal-all','normal-all-after'] for lang in ['en','zh-Hant']]
 assert all(len(s)==67 and s==states[0] for s in states)
 saves=[];save_meta=load(OUT/'save-all-comparison.json')
 assert save_meta['equal'] and save_meta['excluded_fields']==['saved_at']
 for r in save_meta['runs']:
  p=OUT/f"save-all-{r['phase']}-{r['lang']}"/'save.json';assert sha(p)==r['save_sha256'] and r['keys']==73
  d=load(p);d.pop('saved_at');saves.append(d)
 assert len(saves)==4 and all(d==saves[0] for d in saves)
 assert 'FAIL' not in (OUT/'tests-name.log').read_text()
 assert all(x in (OUT/'tests-name.log').read_text() for x in ['internal/town','internal/ui','cmd/kol'])
 for name in ['original_all.sh','prepare_all.py','all-keys.json','remake_all.py','remake_all_after.py','save_all.py','save-all-comparison.json','tests-name.log','ida_items_name.py']:
  copies.append((OUT/name,name))
 for lang in ['en','zh-Hant']:
  copies.append((OUT/f'normal-all-after-{lang}/command.json',f'all-command-{lang}.json'))
 inputs=[dict(path=r['path'],sha256=sha(ROOT/r['path'])) for r in old['product_inputs']]
 inputs.append(dict(path='docs/re/item-name-address-index.json',sha256=sha(ROOT/'docs/re/item-name-address-index.json')))
 if args.export:
  assert PROOF.stat().st_uid==os.getuid() and PROOF.stat().st_gid==os.getgid()
  for src,name in copies:
   assert src.stat().st_uid==os.getuid() and src.stat().st_gid==os.getgid()
   assert not (PROOF/name).exists(),name
  for src,name in copies:shutil.copyfile(src,PROOF/name)
 else:
  for src,name in copies:assert sha(src)==sha(PROOF/name)
 receipt=dict(schema='kol-item-slots-dosgolem-v1',date='2026-10-01',status='十八選中頁與返回資料頁／旅店通過；完整清單仍有差異',
  binary=dict(path='workplace/item-detail-20261001/kol-items-name-linux',sha256=AFTER),before_binary_sha256=BEFORE,
  original_revision=source['revision'],original_seed=1,remake_seed=1,seed_setup='原版正常旅店後DS:1600=1、DS:1602=0；重製版-seed 1',
  source_zip_sha256=old['source_zip_sha256'],source_executable_sha256=old['source_executable_sha256'],
  source_inputs_private=dict(path='workplace/item-detail-20261001/gui3/inputs.json',sha256=sha(OUT/'gui3/inputs.json')),
  original_keys=keys,remake_keys_count=66,allowed_cursor_rect=[304,185,14,14],checks=checks,glyph_checks=glyph_checks,
  selected_readonly_ds={str(k):v for k,v in ds.items()},mode8_call_count=4,unverified_list_frames=pending,
  old_weapon_rejected=dict(frame=31,remake_step=37,pixels=220,bounding_rect=[120,66,87,7]),
  normal_record_states_each=67,normal_record_fields_equal=True,trace_limit='記錄未含物品暫態；返回與繞回以實際畫格、逐鍵比較及唯讀選中觀察核實。完整清單仍未通過。',
  save_comparison=save_meta,product_inputs=inputs,old_receipt_sha256=OLD_RECEIPT,old_attachments_preserved=True,
  annotations=dict(total=14,new=4,index_sha256=sha(ROOT/'docs/re/item-name-address-index.json'),private_payload_sha256=sha(OUT/'item-name-annotated-private.json')),
  tests=dict(command='GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1',log='tests-name.log',circle_frames=104),
  attachments=[dict(path=name,sha256=sha(PROOF/name),uid=(PROOF/name).stat().st_uid,gid=(PROOF/name).stat().st_gid) for src,name in copies],
  rights='PNG僅依既有private GitHub特例保存；沒有公開散布權，原版payload與完整保存資料僅本機。',
  limits=['Theodoric四件非空物品及十三空欄；未驗全部物品種類與動作。','37清單畫格仍有8／9像素差異，未豁免，#64未結案。','原版無中文oracle；中文驗字模、安全框、正常狀態與保存，不冒稱中文parity。'],
  verifier_sha256=sha(Path(__file__)))
 if args.export:(PROOF/'all-slots-receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
 else:assert receipt==load(PROOF/'all-slots-receipt.json')
 assert all(a['uid']==1000 and a['gid']==1000 for a in receipt['attachments'])
 print(json.dumps(dict(status=receipt['status'],checks=len(checks),glyph_checks=len(glyph_checks),pending=len(pending),attachments=len(copies),save_equal=True),ensure_ascii=False))
if __name__=='__main__':main()
