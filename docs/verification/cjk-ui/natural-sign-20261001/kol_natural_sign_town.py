import pathlib,json,subprocess,collections,sys
sys.path.insert(0,'/src/tools')
from build_asset_contact_sheets import read_image
root=pathlib.Path('workplace/cjk-sign-natural-20261001')
tokens=(root/'world-route2.txt').read_text().split()+['RET']
binary='workplace/cjk-edge-20261001/kol-edge-fixed'

def run(tokens,name):
    p=subprocess.run([binary,'-scene','flow','-seed','1','-flow-script',' '.join(tokens),
                      '-shot',str(root/f'{name}.png'),'-trace',str(root/f'{name}.json')],
                     capture_output=True,text=True,timeout=180)
    (root/f'{name}.log').write_text(p.stdout+p.stderr)
    assert p.returncode==0,p.stderr
    return json.loads((root/f'{name}.json').read_text())['steps'][-1]

row=run(tokens,'town-entry')
print('ENTRY',row)
assert row['town']['town']==5 and not row['town']['left_town']
b=json.loads(pathlib.Path('assets/original-derived-data/town-buildings.json').read_text())
town=next(t for t in b['towns'] if t['town']==5)
door=next(d for d in town['doors'] if d['index']==8)
assert door['picture']==19 and (door['x'],door['y'])==(40,19)
grid=read_image(pathlib.Path('assets/original-derived-png/ega/TOWN5/town5-map.png')).rgb()[::3]
blocked=set(town['blocked_tiles'])
start=(row['town']['x'],row['town']['y']); goal=(40,19)
prev={start:None}; queue=collections.deque([start])
while queue:
    cur=queue.popleft()
    if cur==goal: break
    for dx,dy in ((-1,0),(0,-1),(1,0),(0,1)):
        nxt=cur[0]+dx,cur[1]+dy
        if 0<=nxt[0]<64 and 0<=nxt[1]<64 and nxt not in prev and grid[nxt[1]*64+nxt[0]] not in blocked:
            prev[nxt]=cur;queue.append(nxt)
assert goal in prev,'城鎮固定圖無法抵達住宅'
path=[];p=goal
while prev[p] is not None:
    path.append(p);p=prev[p]
path.reverse(); current=start
directions={(-1,0):'L',(0,-1):'U',(1,0):'R',(0,1):'D'}
doors={(d['x'],d['y']) for d in town['doors']}
for p in path:
    if current in doors and current!=start: tokens+=['>','RET']
    delta=p[0]-current[0],p[1]-current[1]
    tokens.append(directions[delta]);current=p
row=run(tokens,'residence-door')
print('DOOR',row)
assert row['town']['town']==5 and row['town']['mode']==1 and row['town']['at_door']
assert (row['town']['x'],row['town']['y'])==(40,19)
# door 是 Session 保留的上一個室內物件；此時門面由 Outdoor.AtDoor 繪製。
# 再經正式 DOOR 入屋，核對新室內建築確實是第 8 門，保留門口原圖畫格。
inside=run(tokens+['RET'],'residence-inside')
assert inside['town']['town']==5 and inside['town']['mode']==2 and inside['town']['door']['index']==8
print('INSIDE',inside)
(root/'residence-route.txt').write_text(' '.join(tokens)+'\n')
(root/'town-navigation.json').write_text(json.dumps({'town':5,'start':list(start),'goal':list(goal),
    'method':'固定格子 PNG 與 typed 不可通行圖塊的四方向最短路，遇門用正式 U-TURN 返回地圖',
    'steps':len(path),'world_route_tokens':len((root/'world-route2.txt').read_text().split()),
    'total_tokens':len(tokens)},ensure_ascii=False,indent=2)+'\n')
