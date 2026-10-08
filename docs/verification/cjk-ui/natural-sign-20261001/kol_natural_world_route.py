import heapq
import hashlib
import json
import os
import pathlib
import sys

sys.path.insert(0,'/src/tools')
from build_asset_contact_sheets import read_image

root=pathlib.Path('workplace/cjk-sign-natural-20261001')
data=json.loads(pathlib.Path('assets/original-derived-data/world-map.json').read_text())
grid_path=pathlib.Path(data['grid']['png'])
manifest=json.loads(pathlib.Path('assets/original-derived-png/ega/manifest.json').read_text())
archive=next(a for a in manifest['archives'] if a['archive']=='TOWN0')
entry=next(i for i in archive['images'] if i['role']=='map_cell_array')
assert hashlib.sha256(grid_path.read_bytes()).hexdigest()==entry['png_sha256']
grid=read_image(grid_path)
pixels=grid.rgb()[::3]
# typed grid.sha256 是原始 column-major 索引 bytes；PNG 自己的 hash 由素材 manifest 提供。
original=bytes(pixels[y*64+x] for x in range(64) for y in range(64))
assert hashlib.sha256(original).hexdigest()==data['grid']['sha256']
blocked=set(data['blocked_tiles'])

def first_direction(w,target):
    start=(w['x']*8+w['sub_x'],w['y']*8+w['sub_y'])
    tx,ty=target
    def goal(p): return p[0]//8==tx and p[1]//8==ty
    def heuristic(p):
        return max(max(tx*8-p[0],0,p[0]-(tx*8+7)),max(ty*8-p[1],0,p[1]-(ty*8+7)))
    def free(p):
        x,y=p[0]//8,p[1]//8
        return pixels[min(63,y+1)*64+x] not in blocked
    queue=[(heuristic(start),0,start)]
    costs={start:0}
    prev={start:None}
    while queue:
        _,cost,cur=heapq.heappop(queue)
        if cost!=costs[cur]: continue
        if goal(cur):
            if cur==start: return None
            while prev[cur]!=start: cur=prev[cur]
            return cur[0]-start[0],cur[1]-start[1]
        for dx in (-1,0,1):
            for dy in (-1,0,1):
                p=cur[0]+dx,cur[1]+dy
                if not (dx or dy) or not (0<=p[0]<512 and 0<=p[1]<512): continue
                # 按正式 Travel.walk：只有跨世界格才檢查腳下 Y+1 的地形。
                if (p[0]//8,p[1]//8)!=(cur[0]//8,cur[1]//8) and not free(p): continue
                new=cost+1
                if new<costs.get(p,10**9):
                    costs[p]=new; prev[p]=cur
                    heapq.heappush(queue,(new+heuristic(p),new,p))
    raise RuntimeError('固定世界地圖無法抵達目標')

source=pathlib.Path('tools/drive_battle_script.py').read_text()
assert source.endswith('main()\n')
namespace={'__file__':'/src/tools/drive_battle_script.py','__name__':'kol_natural_route_driver'}
exec(compile(source[:-len('main()\n')],'/src/tools/drive_battle_script.py','exec'),namespace)
original=namespace['world_step']

def world_step(tokens,row,target):
    w=row.get('world')
    if w is None or target is None or w.get('camping') or w.get('heal_prompt') or w.get('quest_prompt') or (w['x'],w['y'])==tuple(target):
        return original(tokens,row,target)
    d=first_direction(w,target)
    tokens.append(namespace['DIRS'][d])
    return True

namespace['world_step']=world_step
namespace['main']()
