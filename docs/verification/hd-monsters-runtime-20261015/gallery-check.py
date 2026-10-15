from pathlib import Path
from PIL import Image
import numpy as np
import json,hashlib,html,sys
w=Path('/out');prefix=sys.argv[1] if len(sys.argv)>1 else 'all-monster-gallery-v1'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
capture=json.loads((w/(prefix+'.json')).read_text());folder=w/prefix
art={p['index']:p for p in json.loads((w/'generation-plan.json').read_text())['entries']}
groups={};cache={};records=[];negatives=[]
for frame in capture['frames']:
    assert sha(folder/frame['png'])==frame['sha256']
    groups.setdefault(frame['id'],{})[frame['variant']]=frame
def sample(index,width,height):
    key=index,width,height
    if key in cache:return cache[key]
    data=np.asarray(Image.open(w/art[index]['file']).convert('RGBA'),dtype=np.int64)
    data[:,:,:3]=(data[:,:,:3]*data[:,:,3:4]*257)//(255*256)
    sh,sw=data.shape[:2]
    x=np.clip((np.arange(width,dtype=np.float64)+.5)*sw/width-.5,0,sw-1)
    y=np.clip((np.arange(height,dtype=np.float64)+.5)*sh/height-.5,0,sh-1)
    x0=np.floor(x).astype(int);x1=np.minimum(x0+1,sw-1)
    y0=np.floor(y).astype(int);y1=np.minimum(y0+1,sh-1)
    fx=(x-x0)[None,:,None];fy=(y-y0)[:,None,None]
    a=data[y0[:,None],x0[None,:]];b=data[y0[:,None],x1[None,:]]
    c=data[y1[:,None],x0[None,:]];d=data[y1[:,None],x1[None,:]]
    out=np.floor((a*(1-fx)+b*fx)*(1-fy)+(c*(1-fx)+d*fx)*fy+.5).astype(np.int64)
    cache[key]=out;return out
def compose(frame,index=None,reverse=False):
    assert sha(folder/frame['underlay'])==frame['underlay_sha256']
    expected=np.asarray(Image.open(folder/frame['underlay']).convert('RGBA'),dtype=np.int64).copy()
    places=frame['placements'][::-1] if reverse else frame['placements']
    for original,fit in places:
        x0,y0,x1,y1=fit;source=sample(frame['index'] if index is None else index,x1-x0,y1-y0)
        dest=expected[y0:y1,x0:x1]
        expected[y0:y1,x0:x1]=np.clip(source+(dest*(255-source[:,:,3:4])+127)//255,0,255)
    return expected
for name,fs in groups.items():
    base=fs['baseline-ega'];frame=fs['prototype-hd']
    assert all(f['state_sha256']==base['state_sha256'] and f['text']==base['text'] and f['han_characters']>0 for f in fs.values()),name
    images={v:np.asarray(Image.open(folder/f['png']).convert('RGBA'),dtype=np.int64)for v,f in fs.items()}
    assert np.array_equal(images['baseline-ega'],images['prototype-ega']),name
    outside=np.abs(images['baseline-hd']-images['prototype-hd'])
    assert int(outside[:,:,3].max())==0
    for original,fit in frame['placements']:
        x0,y0,x1,y1=original;outside[y0:y1,x0:x1]=0
    outside_max=int(outside.max());assert outside_max<=1,(name,'outside',outside_max)
    diff=np.abs(compose(frame)-images['prototype-hd']);maximum=int(diff.max())
    assert maximum<=1,(name,'native',maximum)
    records.append(dict(id=name,index=frame['index'],kind=frame['kind'],count=frame['count'],ega_rgba_exact=True,
        serializable_state_equal=True,chinese_equal=True,han_characters=frame['han_characters'],
        outside_max=outside_max,native_max=maximum,alpha_max=int(diff[:,:,3].max()),pixels_over_one=int(np.count_nonzero(np.any(diff>1,axis=2)))))
    if frame['index']==0 and frame['count']==12:
        for mode,kwargs in [('wrong-drawing-order',dict(reverse=True)),('wrong-monster-identity',dict(index=1))]:
            bad=np.abs(compose(frame,**kwargs)-images['prototype-hd']);value=int(bad.max())
            assert value>1,(mode,'negative did not reject')
            negatives.append(dict(case=name,mutation=mode,max_difference=value,rejected=True))
    print(name,'PASS native',maximum,'outside',outside_max,flush=True)
assert len(negatives)==2
contracts=[Path('/go/pkg/mod/github.com/hajimehoshi/ebiten/v2@v2.9.9/imagetobytes.go'),Path('/usr/local/go/src/image/draw/draw.go'),Path('/usr/local/go/src/image/color/color.go')]
result=dict(status='PASS limited prepared remake rendering',scenes=len(groups),frames=len(capture['frames']),
    indices=sorted(set(q['index']for q in records)),native_max=max(q['native_max']for q in records),outside_max=max(q['outside_max']for q in records),
    records=records,negative_checks=negatives,ega_rgba_exact=True,scene_receipt_sha256=sha(w/(prefix+'.json')),
    method='Independent NumPy1.24.2 float64 pixel-center bilinear on Go1.26.7 premultiplied RGBA, then source-over in actual order',
    contracts={str(p):sha(p)for p in contracts},original_oracle=False,normal_player_path=False,gui_F10=False,v3=False)
(w/(prefix+'-check.json')).write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
parts=['<!doctype html><meta charset="utf-8"><title>怪物HD遊戲內比較</title><style>body{background:#21171b;color:#eeddb9;font:18px sans-serif;margin:24px}section{margin:28px 0}.pair{display:grid;grid-template-columns:1fr 1fr;gap:12px}img{width:100%}a{color:#e4bf79}</style>',
    '<h1>怪物HD：同狀態遊戲內EGA／HD比較</h1><p>使用既有怪物資料與中文函式，資訊框及1／3／12隻遭遇群像。受控重製驗證，沒有原版oracle或正常全旅程聲明。</p>',
    f'<p><a href="{prefix}-check.json">一色階、EGA、中文及負例核對</a> · <a href="{prefix}.json">擷取紀錄</a> · <a href="generation-plan.json">來源與提示</a></p>']
for name,fs in groups.items():
    parts.append(f'<section><h2>{html.escape(name)}</h2><div class="pair"><div>重製中文EGA<img loading="lazy" src="{prefix}/{fs["prototype-ega"]["png"]}"></div><div>HD候選<img loading="lazy" src="{prefix}/{fs["prototype-hd"]["png"]}"></div></div></section>')
(w/(prefix+'-review.html')).write_text('\n'.join(parts)+'\n')
print('PASS',len(groups),'scenes',len(capture['frames']),'frames native',result['native_max'],'negative',negatives,flush=True)
