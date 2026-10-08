"""正常物品鏡像／立繪窄收據；不豁免剩餘清單字元與x15差異。"""
from pathlib import Path
import argparse,hashlib,importlib.util,json,os,shutil,sys
ROOT=Path('/src');OUT=Path('/out');PROOF=ROOT/'docs/verification/item-detail-20261001'
sys.path.insert(0,'/src/tools')
from png_io import read_png
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def load(p):return json.loads(p.read_text())
old=load(PROOF/'actions-receipt.json')
assert sha(PROOF/'verify_actions.py')==old['verifier_sha256']
spec=importlib.util.spec_from_file_location('historical_actions',PROOF/'verify_actions.py')
legacy=importlib.util.module_from_spec(spec);spec.loader.exec_module(legacy)
AFTER='4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4'
FOCAL=sorted(set(legacy.CHECKED+[15,28,42,56,72,88]))
def normalized_raw(c):
 assert len(c['carried'])==17
 raw=[bytes.fromhex(v) for v in c['carried']];assert all(len(v)==16 for v in raw)
 rows=[v.hex() if v[0] else None for v in raw]
 return rows[7:]+rows[:1]+rows[1:7]
def main():
 ap=argparse.ArgumentParser();ap.add_argument('--export',action='store_true');args=ap.parse_args()
 assert sha(PROOF/'actions-receipt.json')=='d2d13cb2054d707450ce08df70fdc2a4cc85597c29ff9bc30c64f89255ddc049'
 for name in ['receipt.json','all-slots-receipt.json','actions-receipt.json']:
  for a in load(PROOF/name)['attachments']:assert sha(PROOF/a['path'])==a['sha256']
 assert sha(OUT/'kol-items-actions-final-linux')==old['binary']['sha256'] and sha(OUT/'kol-items-mirror-linux')==AFTER
 assert sha(ROOT/'Knights-of-legend.zip')==old['source_zip_sha256']
 source=load(OUT/'gui-actions1/inputs.json');oracle=load(OUT/'gui-actions1/item-calls-private.json')
 assert sha(OUT/'gui-actions1/inputs.json')==old['source_inputs_private']['sha256']
 assert sha(OUT/'gui-actions1/item-calls-private.json')==old['readonly_actor_private']['sha256']
 assert source['revision']==old['original_revision'] and source['seed']==oracle['seed']==1
 assert source['normal_startup'] and oracle['normal_inputs'] and not source['state_injection'] and not oracle['state_injection']
 assert not source['direct_entry'] and not oracle['direct_entry']
 assert oracle['keys']==old['original_keys']==load(OUT/'actions-keys.json') and len(oracle['keys'])==104
 exe=next(p for p in (OUT/'gui-actions1/game').iterdir() if p.name.upper()=='KOLTOWN.EXE')
 assert sha(exe)==old['source_executable_sha256']
 states=oracle['states'];assert len(states)==104
 annotated=load(OUT/'item-figure-annotated-private.json')
 assert annotated['sha256']=='01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2'
 assert annotated['tool']=='IDA Pro 9.4' and annotated['address_space']=='IDA linear'
 rows=[r for f in annotated['records'] for r in f['instructions']]
 attached=[a for r in rows for a in r['annotations']];assert len(attached)==7
 for a in attached:assert a['confidence']=='已證實' and a['evidence'] and a['index_sha256']==sha(Path(a['index']))
 index=load(ROOT/'docs/re/item-figure-address-index.json');assert len(index['entries'])==7
 assert set(index['entries'])=={r['original_ea'] for r in rows if r['annotations']}
 by_ea={r['original_ea']:r for r in rows}
 assert by_ea['0x19596']['bytes']=='8a8788bb' and by_ea['0x18d1b']['bytes']=='837efe10'
 assert sha(ROOT/'workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64')==annotated['database_sha256']
 copies=[];checks=[];pending=[];zero=[]
 for i in range(104):
  original=read_png(OUT/'gui-actions1'/f'item-{i:02d}.png')
  current=read_png(OUT/'normal-mirror-en'/f'step-{i+6:03d}.png')
  pts,cursor=legacy.diff(original,current);assert cursor==67
  if not pts:zero.append(i)
  if i in FOCAL:
   assert not pts,(i,pts[:10])
   checks.append(dict(original_frame=i,remake_step=i+6,compare=[0,0,320,200],outside_cursor_pixels=0,cursor_pixels=67))
   for src,name in [(OUT/'gui-actions1'/f'item-{i:02d}.png',f'mirror-original-{i:02d}.png'),
     (OUT/'normal-mirror-en'/f'step-{i+6:03d}.png',f'mirror-en-{i+6:03d}.png'),
     (OUT/'normal-mirror-zh-Hant'/f'step-{i+6:03d}.png',f'mirror-zh-Hant-{i+6:03d}.png')]:copies.append((src,name))
  elif pts:
   assert len(pts) in [1,8,23]
   if len(pts)==23:assert i in [25,66,82] and all(x==15 or (170<=x<=180 and 54<=y<=58) for x,y in pts)
   else:assert all(x==15 for x,y in pts)
   pending.append(dict(original_frame=i,remake_step=i+6,pixels=pts,status='未通過，未加入允許區'))
  before=read_png(OUT/'normal-actions-after-en'/f'step-{i+6:03d}.png')
  aa,bb=before.rgb(),current.rgb();changed=[]
  for n in range(64000):
   if aa[n*3:n*3+3]!=bb[n*3:n*3+3]:changed.append([n%320,n//320])
  assert all(16<=x<152 and 124<=y<153 for x,y in changed),(i,changed[:8])
  # 中文此切片的畫面不改；資料／立繪計畫另由正常保存及規則測試驗證。
  assert read_png(OUT/'normal-actions-after-zh-Hant'/f'step-{i+6:03d}.png').rgb()==read_png(OUT/'normal-mirror-zh-Hant'/f'step-{i+6:03d}.png').rgb()
 assert len(checks)==29 and len(pending)==70 and len(zero)==34
 before=read_png(OUT/'normal-actions-after-en/step-021.png');original=read_png(OUT/'gui-actions1/item-15.png')
 bad,cursor=legacy.diff(original,before);assert len(bad)==692 and cursor==67
 copies.append((OUT/'normal-actions-after-en/step-021.png','mirror-before-en-021.png'))
 traces=[]
 for lang in ['en','zh-Hant']:
  cmd=load(OUT/f'normal-mirror-{lang}/command.json');assert len(cmd[cmd.index('-flow-script')+1].split())==109
  trace=load(OUT/f'normal-mirror-{lang}/trace.json')['steps'];assert len(trace)==110;traces.append(trace)
  previous=load(OUT/f'normal-actions-after-{lang}/trace.json')['steps']
  for a,b in zip(previous,trace):
   a=json.loads(json.dumps(a));b=json.loads(json.dumps(b));a.get('town',{}).pop('party',None);b.get('town',{}).pop('party',None);assert a==b
  assert trace[93]['town']['party'][0].endswith('carried=3')
 assert traces[0]==traces[1]
 persistence=load(OUT/'mirror-persistence-comparison.json');assert len(persistence['runs'])==12 and persistence['languages_equal']
 assert persistence['normal_startup'] and not persistence['state_injection']
 raw_checks=[];prior_wrong=[];frame_map={15:26,28:38,42:54,56:67,72:83,88:99}
 for r in persistence['runs']:
  frame,lang=r['frame'],r['lang'];wanted=legacy.source_items(states[frame]);reference=None
  for phase,hash_key in [('save','source_save_sha256'),('reload','save_sha256')]:
   p=OUT/f'{phase}-mirror-{frame:02d}-{lang}/save.json';assert sha(p)==r[hash_key]
   data=load(p);c=data['characters'][0]
   assert [g['record'] if g else None for g in legacy.goods(c)]==wanted and normalized_raw(c)==wanted
   if reference is None:reference=legacy.goods(c)
   else:assert legacy.goods(c)==reference
   previous=load(OUT/f'{phase}-actions-{frame:02d}-{lang}/save.json')
   wrong=sum(a!=b for a,b in zip(normalized_raw(previous['characters'][0]),wanted))
   if wrong:prior_wrong.append(dict(frame=frame,lang=lang,phase=phase,wrong_slots=wrong))
   data.pop('saved_at');previous.pop('saved_at');data['characters'][0].pop('carried');previous['characters'][0].pop('carried');assert data==previous
   raw_checks.append(dict(frame=frame,lang=lang,phase=phase,typed_equal=True,raw_equal=True,other_persistent_fields_equal=True))
  shot=OUT/f'reload-mirror-{frame:02d}-{lang}'/f"step-{r['detail_step']:03d}.png"
  assert read_png(shot).rgb()==read_png(OUT/f'normal-mirror-{lang}'/f'step-{frame_map[frame]+6:03d}.png').rgb()
  copies.append((shot,f'mirror-reload-{frame:02d}-{lang}.png'))
 for frame in frame_map:
  for phase in ['save','reload']:
   a=load(OUT/f'{phase}-mirror-{frame:02d}-en/save.json');b=load(OUT/f'{phase}-mirror-{frame:02d}-zh-Hant/save.json')
   a.pop('saved_at');b.pop('saved_at');assert a==b
 assert len(raw_checks)==24 and len(prior_wrong)==20
 tests=(OUT/'tests-mirror.log').read_text();assert 'FAIL' not in tests and all(x in tests for x in ['internal/town','internal/ui','cmd/kol'])
 negative=(OUT/'tests-mirror-before.log').read_text();assert all('--- FAIL: '+n in negative for n in ['TestGiveItem','TestFigureIgnoresEmptyRecordTails','TestItemActionsSynchronizeCarriedBeforeFigure'])
 for name in ['prepare_mirror_runs.py','remake_mirror.py','save_mirror.py','reload_mirror.py','compare_mirror.py','tests-mirror.log','tests-mirror-before.log',
   'mirror-persistence-comparison.json','mirror-build-private.json','mirror-run-script-origins.json','ida_item_figure.py']:
  copies.append((OUT/name,name))
 for lang in ['en','zh-Hant']:copies.append((OUT/f'normal-mirror-{lang}/command.json',f'mirror-command-{lang}.json'))
 inputs=[dict(path=r['path'],sha256=sha(ROOT/r['path'])) for r in old['product_inputs']]
 inputs.append(dict(path='docs/re/item-figure-address-index.json',sha256=sha(ROOT/'docs/re/item-figure-address-index.json')))
 if args.export:
  assert PROOF.stat().st_uid==os.getuid() and PROOF.stat().st_gid==os.getgid()
  for src,name in copies:assert src.stat().st_uid==os.getuid() and src.stat().st_gid==os.getgid() and not (PROOF/name).exists(),name
  for src,name in copies:shutil.copyfile(src,PROOF/name)
 else:
  for src,name in copies:assert sha(src)==sha(PROOF/name)
 receipt=dict(schema='kol-item-mirror-dosgolem-v1',date='2026-10-01',status='物品鏡像與立繪殘留修正通過；完整清單仍有差異',
  binary=dict(path='workplace/item-detail-20261001/kol-items-mirror-linux',sha256=AFTER),before_binary_sha256=old['binary']['sha256'],
  source_zip_sha256=old['source_zip_sha256'],source_executable_sha256=old['source_executable_sha256'],
  original_revision=old['original_revision'],original_seed=1,remake_seed=1,seed_setup=old['seed_setup'],
  source_inputs_private=old['source_inputs_private'],readonly_actor_private=old['readonly_actor_private'],
  original_keys=oracle['keys'],remake_keys_count=109,normal_startup=True,state_injection=False,direct_entry=False,
  allowed_cursor_rect=[304,185,14,14],checks=checks,all_zero_frames=zero,unverified_list_frames=pending,
  old_belt_rejected=dict(original_frame=15,remake_step=21,pixels=692,bounding_rect=[17,125,133,28]),
  changed_english_rect=[16,124,136,29],chinese_normal_frames_unchanged=104,
  normal_record_states_each=110,normal_record_fields_equal=True,trace_change='前版摘要的持物數以過期Carried計算；本版第93步正確變為3，其餘場景／規則欄位相同。',
  persistence=persistence,raw_item_checks=raw_checks,old_raw_mirror_rejected=prior_wrong,excluded_before_after_fields=['saved_at','characters[0].carried'],
  product_inputs=inputs,historical_receipt_sha256=sha(PROOF/'actions-receipt.json'),historical_verifier_sha256=sha(PROOF/'verify_actions.py'),
  annotations=dict(total=7,new=7,index_sha256=sha(ROOT/'docs/re/item-figure-address-index.json'),private_payload_sha256=sha(OUT/'item-figure-annotated-private.json')),
  tests=dict(command='GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1',log='tests-mirror.log',circle_frames=104,old_failures=3),
  attachments=[dict(path=name,sha256=sha(PROOF/name),uid=(PROOF/name).stat().st_uid,gid=(PROOF/name).stat().st_gid) for src,name in copies],
  rights='原版PNG僅依既有private GitHub特例保存；原版payload、完整存檔及輸入只留本機，沒有公開散布權。',
  limits=['只驗Theodoric巨劍成功搬移、丟棄與同期物品鏡像；其他物品、送禮／食用正常原版流程未全驗。',
   '70清單畫格未通過，含x15的1／8像素與三張手持小框22文字像素，不列允許區；#64與父規格仍開放。',
   '繁中沿原版幾何共用立繪計畫；本切片畫面保持前版，不宣稱繁中原版像素一致或整個繁中立繪已驗。',
   '只對拍十七物品欄；原版完整角色、時間、金幣與亂數內部未對拍。'],verifier_sha256=sha(Path(__file__)))
 assert all(a['uid']==1000 and a['gid']==1000 for a in receipt['attachments'])
 if args.export:
  assert not (PROOF/'mirror-receipt.json').exists();(PROOF/'mirror-receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
 else:assert receipt==load(PROOF/'mirror-receipt.json')
 print(json.dumps(dict(status=receipt['status'],checks=len(checks),pending=len(pending),raw_save_and_reload_checks=len(raw_checks),
  old_raw_rejected=len(prior_wrong),attachments=len(copies),pngs=sum(n.endswith('.png') for s,n in copies),inputs=len(inputs)),ensure_ascii=False))
if __name__=='__main__':main()
