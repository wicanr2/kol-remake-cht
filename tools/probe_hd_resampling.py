#!/usr/bin/env python3
"""隔離的HD整數雙線性取樣原型；入口：docs/verification/release-v1.0.2-20261013.md。"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT=Path('/src')
OUT=Path('/probe')

PATCH=r'''
type hdSampleKey struct {
    source *ebiten.Image
    width, height int
}
var hdSampleCache = map[hdSampleKey]*ebiten.Image{}

// 目的像素中心對應來源像素中心；整數權重避免驅動取樣精度差。
func hdSample(src *ebiten.Image, width, height int) *ebiten.Image {
    key:=hdSampleKey{src,width,height}
    if result:=hdSampleCache[key]; result!=nil { return result }
    b:=src.Bounds(); sw,sh:=b.Dx(),b.Dy()
    raw:=make([]byte,sw*sh*4); src.ReadPixels(raw)
    result:=image.NewRGBA(image.Rect(0,0,width,height))
    axis:=func(p,n,d int)(int,int,int64,int64) {
        denominator:=int64(2*d)
        numerator:=int64((2*p+1)*n-d)
        if numerator<=0 { return 0,0,0,denominator }
        if numerator>=int64(n-1)*denominator { return n-1,n-1,0,denominator }
        start:=int(numerator/denominator)
        return start,start+1,numerator%denominator,denominator
    }
    for y:=0;y<height;y++ {
        y0,y1,fy,dy:=axis(y,sh,height)
        for x:=0;x<width;x++ {
            x0,x1,fx,dx:=axis(x,sw,width)
            divisor:=dx*dy
            for channel:=0;channel<4;channel++ {
                top:=int64(raw[(y0*sw+x0)*4+channel])*(dx-fx)+int64(raw[(y0*sw+x1)*4+channel])*fx
                bottom:=int64(raw[(y1*sw+x0)*4+channel])*(dx-fx)+int64(raw[(y1*sw+x1)*4+channel])*fx
                result.Pix[y*result.Stride+x*4+channel]=byte((top*(dy-fy)+bottom*fy+divisor/2)/divisor)
            }
        }
    }
    cached:=ebiten.NewImageFromImage(result); hdSampleCache[key]=cached
    return cached
}
'''

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


assert Path('/.dockerenv').exists()
assert OUT.stat().st_uid==os.getuid()
if len(sys.argv)>1 and sys.argv[1]=='render':
    case=sys.argv[2]
    assert case in ('linux-debian','linux-ubuntu','windows-ubuntu')
    formal=len(sys.argv)>3 and sys.argv[3]=='formal'
    data=Path('/delivery/full');fixture=Path('/delivery/smoke/linux-gui/inputs/save.json')
    def platform_path(path):
        return 'Z:'+str(path).replace('/','\\') if case.startswith('windows') else str(path)
    binary=OUT/(('formal-' if formal else '')+('kol.exe' if case.startswith('windows') else 'kol'))
    prefix=['/usr/lib/wine/wine64'] if case.startswith('windows') else []
    for graphics in ('ega','hd'):
        dest=OUT/(('formal-' if formal else '')+case+'-'+graphics);dest.mkdir()
        command=prefix+[platform_path(binary),'-scene','play','-lang','zh-Hant',
            '-graphics',graphics,'-load',platform_path(fixture),'-seed','1',
            '-play-script','L R RET SPACE','-shot-steps','0,1,2,3,4','-shot-dir',platform_path(dest)]
        (dest/'command.json').write_text(json.dumps(dict(argv=command,binary_sha256=digest(binary),
            fixture_sha256=digest(fixture),graphics_library='auto'),indent=2)+'\n')
        with (dest/'run.log').open('w') as log:
            subprocess.run(command,cwd=data,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=120)
        print(case,graphics,'PASS',flush=True)
    sys.exit(0)
if len(sys.argv)>1 and sys.argv[1]=='compare':
    import numpy as np
    from PIL import Image
    formal=len(sys.argv)>2 and sys.argv[2]=='formal'
    cases=tuple(('formal-' if formal else '')+n for n in ('linux-debian','linux-ubuntu','windows-ubuntu'))
    rows=[]
    for graphics in ('ega','hd'):
        for left,right in ((cases[0],cases[1]),(cases[0],cases[2]),(cases[1],cases[2])):
            for path in sorted((OUT/(left+'-'+graphics)).glob('step-*.png')):
                target=OUT/(right+'-'+graphics)/path.name
                a=np.asarray(Image.open(path).convert('RGBA')).astype(np.int16)
                b=np.asarray(Image.open(target).convert('RGBA')).astype(np.int16)
                diff=np.abs(a-b);maximum=int(diff.max())
                rows.append(dict(left=left,right=right,graphics=graphics,frame=path.name,
                    max_channel_delta=maximum,different_pixels=int(np.count_nonzero(diff.max(2))),
                    numeric_bound_met=maximum <= (0 if graphics=='ega' else 1),
                    left_sha256=digest(path),right_sha256=digest(target)))
    assert len(rows)==30
    status='PASS' if all(r['numeric_bound_met'] for r in rows) else 'NUMERIC_BOUND_UNMET'
    receipt=dict(schema='kol-hd-fixed-sampling/1',status=status,production=formal,
        comparisons=rows,numeric_bound={'ega':0,'hd':1},original_parity=False)
    (OUT/('formal-receipt.json' if formal else 'receipt.json')).write_text(json.dumps(receipt,indent=2)+'\n')
    print(status, 'EGAmax',max(r['max_channel_delta'] for r in rows if r['graphics']=='ega'),
          'HDmax',max(r['max_channel_delta'] for r in rows if r['graphics']=='hd'),flush=True)
    sys.exit(0)
stage=Path('/tmp/kol-sample-probe')
stage.mkdir()
for name in ('cmd','internal','go.mod','go.sum'):
    source=ROOT/name;dest=stage/name
    if source.is_dir():shutil.copytree(source,dest)
    else:shutil.copy2(source,dest)
base_commit='670ba55b392a3be548780bb807184e67514c0fb5'
for name in ('internal/ui/hd.go','internal/ui/hd_interface.go'):
    (stage/name).write_bytes(subprocess.check_output(['git','-C',str(ROOT),'show',base_commit+':'+name]))
for name in ('internal/ui/hd_sampling.go','internal/ui/hd_sampling_test.go'):
    (stage/name).unlink(missing_ok=True)
base_inputs={str(p.relative_to(stage)):digest(p) for folder in ('cmd','internal') for p in sorted((stage/folder).rglob('*.go'))}
base_inputs.update({name:digest(stage/name) for name in ('go.mod','go.sum')})
assert base_inputs==json.loads((ROOT/'docs/verification/hd-music-runtime-20261013/receipt.json').read_text())['program_inputs']
if len(sys.argv)>1 and sys.argv[1]=='validate':
    print('Pinned prototype base PASS',len(base_inputs),'inputs',flush=True)
    sys.exit(0)
path=stage/'internal/ui/hd.go';text=path.read_text()
start=text.index('func drawHD(');end=text.index('\nfunc (h *HD) party(',start)
replacement='''func drawHD(dst, src *ebiten.Image, rect image.Rectangle) {
    sampled:=hdSample(src,rect.Dx(),rect.Dy())
    var op ebiten.DrawImageOptions
    op.GeoM.Translate(float64(rect.Min.X),float64(rect.Min.Y))
    op.Filter=ebiten.FilterNearest
    dst.DrawImage(sampled,&op)
}
'''
path.write_text(text[:start]+replacement+text[end:]+PATCH)
path=stage/'internal/ui/hd_interface.go';text=path.read_text()
needle='\th.style = s'
text=text.replace(needle,'\ts.images["innkeeper-cropped.png"] = s.images["innkeeper-painted-v1.png"].SubImage(image.Rect(0, 90, 1254, 1084)).(*ebiten.Image)\n'+needle)
text=text.replace('pic = s.images["innkeeper-painted-v1.png"].SubImage(image.Rect(0, 90, 1254, 1084)).(*ebiten.Image)','pic = s.images["innkeeper-cropped.png"]')
path.write_text(text)
shutil.copytree('/go/pkg/mod','/tmp/mod')
env=dict(os.environ,PATH='/usr/local/go/bin:'+os.environ['PATH'],GOWORK='off',GOPROXY='off',
    GOSUMDB='off',GOMODCACHE='/tmp/mod',GOCACHE='/tmp/gocache',GOPATH='/tmp/gopath')
subprocess.run(['gofmt','-w','internal/ui/hd.go','internal/ui/hd_interface.go'],cwd=stage,env=env,check=True)
for name in ('hd.go','hd_interface.go'):
    shutil.copy2(stage/'internal/ui'/name,OUT/('prototype-'+name))
for binary,more in (('kol',{}),('kol.exe',dict(GOOS='windows',GOARCH='amd64',CGO_ENABLED='0'))):
    with (OUT/(binary+'-build.log')).open('w') as log:
        subprocess.run(['go','build','-trimpath','-buildvcs=false','-o',str(OUT/binary),'./cmd/kol'],
            cwd=stage,env=dict(env,**more),stdout=log,stderr=subprocess.STDOUT,check=True,timeout=500)
    print(binary,digest(OUT/binary),flush=True)
inputs={str(p.relative_to(stage)):digest(p) for root in ('cmd','internal') for p in sorted((stage/root).rglob('*.go'))}
inputs.update({name:digest(stage/name) for name in ('go.mod','go.sum')})
(OUT/'build-inputs.json').write_text(json.dumps(dict(schema='kol-hd-resampling-prototype/1',
    production=False,base_commit=base_commit,program_inputs=inputs,binaries={n:digest(OUT/n) for n in ('kol','kol.exe')}),indent=2)+'\n')
