set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/map-cache GOPATH=/tmp/map-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import shutil,json,hashlib
r=Path('/src');w=Path('/out');lab=Path('/tmp/town2-preview');lab.mkdir();sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
inputs=json.loads((w/'formal-source-inputs.json').read_text())
for name,expected in inputs.items():
 assert sha(r/name)==expected;p=lab/name;p.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(r/name,p)
for name in ['assets','docs','tools']:(lab/name).symlink_to(r/name,target_is_directory=True)
shutil.copyfile(r/'docs/verification/hd-town-crests-runtime-20261015/gallery.go.txt',lab/'cmd/kol/crest_preview_setup_test.go')
s=(r/'docs/verification/hd-town-maps-runtime-20261016/gallery.go.txt').read_text().replace('/out/position-city-assets','/out/candidate-assets')
s=s.replace('Decode("TOWN1",0)','Decode("TOWN2",0)').replace('buildings.Town(1)','buildings.Town(2)').replace('out.Town=1','out.Town=2').replace('prepared real TOWN1','prepared real TOWN2')
s=s.replace('candidate:=base;raw','candidate:=base;raw')
s=s.replace('out:=positionCityOutdoor(t,m,47,9)','out:=positionCityOutdoor(t,m,37,45)').replace('第一城連續手繪HD切換驗收','第二城連續手繪HD切換驗收')
(lab/'cmd/kol/position_city_gallery_test.go').write_text(s);(w/'formal-gallery.go.txt').write_text(s)
helper='package ui\n// Disposable verification only. Never enters production.\nfunc(h *HD)SuspendTownMapVerification()func(){old:=h.townMaps;h.townMaps=nil;return func(){h.townMaps=old}}\n'
(lab/'internal/ui/hd_town_maps_verification.go').write_text(helper);(w/'verification-helper.go.txt').write_text(helper)
PY
cd /tmp/town2-preview
gofmt -w cmd/kol/position_city_gallery_test.go cmd/kol/crest_preview_setup_test.go internal/ui/hd_town_maps_verification.go
mkdir -p /out/bin /out/formal-city-runtime
go test -c -o /out/bin/town2-gallery ./cmd/kol > /out/formal-build.log 2>&1
mkdir -p /tmp/.X11-unix
chmod 1777 /tmp/.X11-unix
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp > /out/formal-xvfb.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
go test -p 2 ./internal/... ./cmd/... > /out/formal-tests.log 2>&1
go test -p 2 -tags sceneverify ./cmd/kol > /out/formal-sceneverify.log 2>&1
go build -o /out/bin/formal-town2 ./cmd/kol > /out/formal-program-build.log 2>&1
KOL_POSITION_CITY=1 timeout 600 /out/bin/town2-gallery -test.run '^TestHDPositionCity$' -test.v > /out/formal-run.log 2>&1
python3 - <<'PY'
from pathlib import Path
import hashlib,json
lab=Path('/tmp/town2-preview');paths=[lab/'go.mod',lab/'go.sum']+[p for d in ['cmd','internal']for p in (lab/d).rglob('*.go')]
Path('/out/formal-verification-source-inputs.json').write_text(json.dumps({str(p.relative_to(lab)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(paths)},indent=2)+'\n')
PY
printf '%s\n' 'Second city complete gallery captured'
