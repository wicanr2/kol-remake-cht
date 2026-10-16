from pathlib import Path
from PIL import Image
import hashlib,json,shutil
R,W=Path('/src'),Path('/out');V=R/'docs/verification/hd-all-maps-runtime-20261017'
assert V.parent.stat().st_uid==W.stat().st_uid==1000
full=json.loads((W/'maps-check.json').read_text());world=json.loads((W/'maps-world-check.json').read_text())
assert full['scenes']==504 and world['scenes']==84 and world['delivery_composite_max']<=1
gui={kind:json.loads((W/('formal-gui-'+kind)/'results.json').read_text())for kind in ['town','world']}
assert all(len(v['display_events'])==12 and v['exit_code']==0 for v in gui.values())
V.mkdir(exist_ok=True);assert V.stat().st_uid==1000
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
files={}
names={'maps-check.json':'full-check.json','maps-world-check.json':'world-point-check.json','maps-captures.json':'full-captures.json','maps-world-captures.json':'world-point-captures.json','maps-source-inputs.json':'pre-marker-production-inputs.json','maps-world-source-inputs.json':'production-inputs.json','resize-receipt.json':'resize-receipt.json','previous-manifest.json':'previous-manifest.json','maps-full-tests.log':'full-tests.log','maps-world-ui-tests.log':'world-ui-tests.log','maps-sceneverify.log':'sceneverify.log'}
for source,name in names.items():shutil.copyfile(W/source,V/name)
for name in ['resize-maps.py','install-maps.py','assemble-maps.py','prepare-maps.py','persist-maps.py','maps-gallery.go.txt','maps-gui.go.txt','run-map-gallery.sh','maps-gui.py','town2-formal-gui.py','check-maps.py','publish-maps-evidence.py','pre-marker-hd_town_maps.go.txt','pre-marker-hd_town_maps_test.go.txt']:
 shutil.copyfile(W/name,V/name)
for kind in gui:
 shutil.copyfile(W/('formal-gui-'+kind)/'results.json',V/('gui-'+kind+'.json'))
 for p in (W/('formal-gui-'+kind)).glob('*-state.json'):shutil.copyfile(p,V/(kind+'-'+p.name))
rows=[]
for town,x,y,label in [(3,23,29,'第三城街區'),(4,23,29,'第四城高地'),(5,23,29,'第五城城堡'),(6,11,53,'第六城木橋'),(7,47,11,'霍布城建築'),(0,23,29,'世界山脈與河岸')]:
 case=f'town{town}-{x:02}-{y:02}';row={'case':case,'label':label,'images':{}}
 for mode in ['ega','hd']:
  runtime='maps-world-runtime'if town==0 else 'maps-runtime';source=W/runtime/(case+'-'+mode+'.png');im=Image.open(source).convert('RGB');name=case+'-'+mode+'.png';im.resize((768,480),Image.Resampling.LANCZOS).save(V/name,optimize=True)
  row['images'][mode]={'file':name,'sha256':sha(V/name),'source_file':str(source.relative_to(R))if source.is_relative_to(R)else 'workplace/hd-full-20261016/'+str(source.relative_to(W)),'source_sha256':sha(source),'source_size':list(im.size),'preview_size':[768,480],'tool':'Pillow9.4.0 LANCZOS'}
 rows.append(row)
html='<!doctype html><html lang="zh-Hant"><meta charset="utf-8"><title>七城與世界連續手繪地圖</title><style>body{background:#21191e;color:#eee;font:18px sans-serif;margin:24px}section{margin:32px 0}figure{margin:0}div{display:grid;grid-template-columns:1fr 1fr;gap:12px}img{width:100%;height:auto}</style><h1>七城與世界地圖</h1><p>正式重製程式的同一受控狀態。左：繁中EGA；右：HD。這些是768×480預覽；原始畫布與雜湊另存於清冊。原版oracle與正常全旅程不在本批範圍。天候及城鎮頭盔仍續作。</p>'
for row in rows:html+='<section><h2>'+row['label']+'</h2><div>'+''.join('<figure><figcaption>'+mode.upper()+'</figcaption><img src="'+row['images'][mode]['file']+'"></figure>'for mode in ['ega','hd'])+'</div></section>'
(V/'review.html').write_text(html+'\n')
receipt={'schema':'kol-hd-all-maps-acceptance/1','status':'PASS limited CONFORMED position backgrounds and native world point','maps_total':8,'regions_total':336,'new_regions':252,'stages':{'full_group':full,'world_point':world},'GUI_F10':gui,'gui_scope':'actual1920x1200 keyboard and960x600 mouse,24 events,total states equal; GUI binary precedes RGB point correction; controls unchanged; final point separately verified by84 full native roundtrips','GUI_binary_sha256':sha(W/'bin/maps-gui'),'capacity':{k:v for k,v in json.loads((W/'resize-receipt.json').read_text()).items()if k!='records'},'comparisons':rows,'original_oracle':False,'normal_player_path':False,'platform':'Linux Docker only','source_zip_sha256':'2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46','toolchain':'kol-package:1.26.7-ebiten2.9.9-20261011-r1;Go1.26.7;Ebitengine2.9.9;Pillow9.4.0','scope':'Seven town and world position backgrounds; new five towns/world accepted; first two retained; town helmets/weather/battle/intro and other HD remain open; v3 and rules unchanged; source-bound art private repository only.'}
(V/'receipt.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
files={p.name:sha(p)for p in sorted(V.iterdir())if p.is_file()and p.name!='manifest.json'}
(V/'manifest.json').write_text(json.dumps({'schema':'kol-hd-all-maps-evidence/1','rights':'private GitHub preservation only; no public distribution rights asserted','files':files,'comparisons':rows},ensure_ascii=False,indent=2)+'\n')
scope=json.loads((W/'plan.json').read_text());scope['completed']+=['TOWN3','TOWN4','TOWN5','TOWN6','TOWN7','TOWN0-position-background','world-native-RGB-point'];scope['remaining'].pop('TOWN3-TOWN7',None);scope['remaining']['world-map']='position background accepted;25 location/camping events generated in part, runtime pending';scope['tools']=sorted(set(scope['tools']+['maps-gui.go.txt','maps-gui.py','check-maps.py','publish-maps-evidence.py']));scope['maps_evidence']='docs/verification/hd-all-maps-runtime-20261017.md';(W/'plan.json').write_text(json.dumps(scope,ensure_ascii=False,indent=2)+'\n')
print('Published limited full-map evidence',len(files),'files; private only; global HD active')
