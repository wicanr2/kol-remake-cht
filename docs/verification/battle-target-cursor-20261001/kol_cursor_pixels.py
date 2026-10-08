from pathlib import Path
import json,hashlib,sys,os
sys.path.insert(0,'/src/tools')
from png_io import read_png
root=Path('/out');assets=Path('/src/assets/original-derived-png/ega')
marker=read_png(assets/'KOLPICS-MARKERS/kolpics-markers.png')
atlas=read_png(assets/'BATTLE/battle-tiles-08.png')
assert (marker.width,marker.height)==(16,160)
assert (atlas.width,atlas.height)==(160,128)
def bit(x):return 1<<(7-x if x<8 else 23-x)
def tile(i):
 assert 0<=i<80
 return atlas.crop(i%10*16,i//10*16,16,16).rgb()
def composite(bg,sprite):
 assert len(bg)==len(sprite)==768
 result=bytearray(bg)
 for y in range(16):
  mask=0
  for x in range(16):
   if sprite[(y*16+x)*3:(y*16+x)*3+3]!=b'\0\0\0':mask|=bit(x)
  wide=(mask|mask<<1|mask>>1)&65535
  for x in range(16):
   if wide&bit(x):
    i=(y*16+x)*3;result[i:i+3]=sprite[i:i+3]
 return bytes(result)
checks=[];sequences=[];retained=[]
for name in sys.argv[1:]:
 run=root/name;calls=json.loads((run/'cursor-calls.json').read_text())['calls']
 layers={s.get('frame'):s['icon_slots'] for s in json.loads((run/'battle-planning.json').read_text())['steps']}
 renders=[x for x in calls if x['ida_ea']=='0x1CB28' and x['caller_ida']=='0x114AB']
 sequences.append({'run':name,'render_count':len(renders),'units':sorted(set(x['args'][0] for x in renders)),
   'first_16':[x['args'][0] for x in renders[:16]]})
 for state_path in sorted(run.glob('plan-*.png.state.json')):
  state=json.loads(state_path.read_text())
  if layers.get(state['frame'])!=2:continue
  assert state['cursor_mode']==5
  row,col=state['cursor_row'],state['cursor_col'];offset=0x153+640*row+2*col
  candidates=[x for x in renders if x['step']<=state['step'] and x['args'][1]==offset]
  assert candidates,(name,state_path.name)
  call=candidates[-1];unit=call['args'][0];assert 5<=unit<=9
  pos=row*9+col;tid=bytes.fromhex(state['visible_tiles'])[pos];sid=bytes.fromhex(state['visible_sprites'])[pos]
  base=tile(tid)
  if sid:base=composite(base,tile(sid-1));assert call['args'][3]==100
  else:assert call['args'][3]==tid
  expected=composite(base,marker.crop(0,unit*16,16,16).rgb())
  image=read_png(run/state['frame']);actual=image.crop(152+col*16,8+row*16,16,16).rgb()
  diff=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,768,3))
  assert diff==0,(name,state['frame'],unit,tid,sid,diff)
  bad=bytearray(expected);bad[0]^=1
  assert bad!=actual,'缺一位元反例沒有被拒絕'
  checks.append({'run':name,'frame':state['frame'],'frame_sha256':hashlib.sha256((run/state['frame']).read_bytes()).hexdigest(),
   'capture_step':state['step'],'last_draw_step':call['step'],'call_args':call['args'],'row':row,'col':col,'tile':tid,'sprite_code':sid,
   'unit':unit,'rect':[152+col*16,8+row*16,16,16],'pixels':256,'different_pixels':diff,
   'rgb_sha256':hashlib.sha256(actual).hexdigest(),'one_bit_negative_rejected':True})
 for frame in ({'gui4':['plan-06.png','plan-07.png','plan-08.png'],'gui5':['plan-31.png']}.get(name,[])):
  state=json.loads((run/(frame+'.state.json')).read_text())
  assert state['cursor_mode']==5
  call=next(x for x in reversed(renders) if x['step']<=state['step'])
  unit,offset,_,bgid,_,_=call['args'];x=(offset%40)*8;y=offset//40;col=(x-152)//16;row=(y-8)//16
  pos=row*9+col;tid=bytes.fromhex(state['visible_tiles'])[pos];sid=bytes.fromhex(state['visible_sprites'])[pos]
  base=tile(tid)
  if sid:base=composite(base,tile(sid-1))
  expected=composite(base,marker.crop(0,unit*16,16,16).rgb())
  actual=read_png(run/frame).crop(x,y,16,16).rgb()
  diff=sum(actual[i:i+3]!=expected[i:i+3] for i in range(0,768,3));assert diff==0,(name,frame,diff)
  retained.append({'run':name,'frame':frame,'capture_step':state['step'],'last_draw_step':call['step'],
    'unit':unit,'rect':[x,y,16,16],'sprite_code':sid,'cursor_mode':5,'different_pixels':diff,
    'rgb_sha256':hashlib.sha256(actual).hexdigest(),'frame_sha256':hashlib.sha256((run/frame).read_bytes()).hexdigest()})
assert checks
out={'schema':'kol-original-target-cursor-pixels-v1','method':'原版實際最後一次 sub_11390→sub_1CB28 呼叫相位，獨立逐列16-bit遮罩合成，核對完整16×16框',
 'assets':[{'path':str(p.relative_to(Path('/src'))),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in [assets/'KOLPICS-MARKERS/kolpics-markers.png',assets/'BATTLE/battle-tiles-08.png']],
 'checks':checks,'retained_in_submenus':retained,'sequences':sequences,'remake_compared':False,
 'limits':['只證明原版游標的來源、相位及合成；未修改或驗收重製版','實際等待秒數與動畫速度尚未作重製版契約']}
(root/'original-pixel-checks.json').write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'checks':len(checks),'occupied':sum(c['sprite_code']!=0 for c in checks),'empty':sum(c['sprite_code']==0 for c in checks),'different_pixels':0}))
