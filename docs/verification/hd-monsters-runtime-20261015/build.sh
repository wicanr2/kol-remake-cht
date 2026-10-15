set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/monster-gallery3-cache GOPATH=/tmp/monster-gallery3-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import hashlib,json,shutil
r=Path('/src');w=Path('/out');s=Path('/tmp/monster-gallery3-source');s.mkdir()
files=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
inputs={str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in files}
assert inputs==json.loads((w/'final-source-inputs-v2.json').read_text())
for p in files:
 t=s/p.relative_to(r);t.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,t)
for name in ['assets','docs','tools']:(s/name).symlink_to(r/name,target_is_directory=True)
text=(w/'monster-all-gallery_test.go.txt').read_text().replace('useMonsterPrototype=i==1','').replace('PrototypeMonsterPlacements','MonsterPlacements').replace('all-monster-gallery-v1','formal-monster-gallery-v3')
text=text.replace(';i,j int',';underlay *ebiten.Image;i,j int').replace('for i:=0;i<2;i++','for i:=0;i<1;i++').replace('gallery.models[i]=model','gallery.models[i]=model;gallery.models[1]=model')
text=text.replace('frame.UnderlaySHA,g.err=monsterPNG(filepath.Join(g.dir,frame.Underlay),monsterPrototypeUnderlays[model]);','if g.underlay==nil{g.underlay=ebiten.NewImage(1920,1200)};underlay:=g.underlay;underlay.Clear();var op ebiten.DrawImageOptions;op.GeoM.Scale(2,2);underlay.DrawImage(model.hdMonsterUnderlay.Image(),&op);frame.UnderlaySHA,g.err=monsterPNG(filepath.Join(g.dir,frame.Underlay),underlay);')
(s/'cmd/kol/monster_gallery_test.go').write_text(text)
gui=(w/'monster-gui_test.go.txt').read_text().replace('for i:=0;i<2;i++','for i:=0;i<1;i++').replace('gallery.models[i]=model','gallery.models[i]=model;gallery.models[1]=model')
(s/'cmd/kol/monster_gui_test.go').write_text(gui)
(w/'gallery-v3-fixture.go.txt').write_text(text);(w/'gui-v3-fixture.go.txt').write_text(gui)
print('Same402 formal sources; one model and reusable capture surface',flush=True)
PY
cd /tmp/monster-gallery3-source
gofmt -w cmd/kol/monster_gallery_test.go cmd/kol/monster_gui_test.go
go test -c -o /out/bin/formal-monster-gui-v3 ./cmd/kol >/out/build-formal-gui-v3.log 2>&1
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-gallery-v3.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 KOL_MONSTER_GALLERY=1
python3 /out/monitor-gallery.py
