set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/cache GOPATH=/tmp/gopath GOMEMLIMIT=4GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import shutil,json,hashlib
r=Path('/src');w=Path('/out');lab=Path('/tmp/full-map-gallery');lab.mkdir()
for d in ['cmd','internal']:shutil.copytree(r/d,lab/d)
for n in ['go.mod','go.sum']:shutil.copyfile(r/n,lab/n)
for n in ['assets','docs','tools']:(lab/n).symlink_to(r/n,target_is_directory=True)
shutil.copyfile(r/'docs/verification/hd-town-crests-runtime-20261015/gallery.go.txt',lab/'cmd/kol/crest_preview_setup_test.go')
gallery=(w/'maps-gallery.go.txt').read_text()
world_only=__import__('os').environ.get('KOL_MAP_WORLD_ONLY')=='1'
if world_only:gallery=gallery.replace('/out/maps-runtime','/out/maps-world-runtime').replace('/out/maps-captures.json','/out/maps-world-captures.json').replace('"regions":252','"regions":42').replace('"maps":6','"maps":1')
(lab/'cmd/kol/full_maps_gallery_test.go').write_text(gallery)
if __import__('os').environ.get('KOL_MAP_TASK')=='gui':shutil.copyfile(w/'maps-gui.go.txt',lab/'cmd/kol/full_maps_gui_test.go')
(lab/'internal/ui/hd_maps_verification.go').write_text('package ui\n// Disposable verification only.\nfunc(h *HD)SuspendTownMapVerification()func(){old:=h.townMaps;h.townMaps=nil;return func(){h.townMaps=old}}\n')
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in (r/d).rglob('*.go')]
if __import__('os').environ.get('KOL_MAP_TASK')!='gui':(w/('maps-world-source-inputs.json'if world_only else 'maps-source-inputs.json')).write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(paths)},indent=2)+'\n')
PY
cd /tmp/full-map-gallery
gofmt -w cmd/kol/full_maps_gallery_test.go cmd/kol/crest_preview_setup_test.go internal/ui/hd_maps_verification.go
mkdir -p /out/bin /out/maps-runtime
if [ "${KOL_MAP_TASK:-maps}" = gui ]; then
 gofmt -w cmd/kol/full_maps_gui_test.go
 go test -c -o /out/bin/maps-gui ./cmd/kol > /out/maps-gui-build.log 2>&1
 python3 /out/maps-gui.py
 exit
fi
binary=maps-gallery
if [ "${KOL_MAP_WORLD_ONLY:-0}" = 1 ]; then binary=maps-world; fi
go test -c -o "/out/bin/$binary" ./cmd/kol > "/out/$binary-build.log" 2>&1
Xvfb :97 -screen 0 2048x1536x24 -nolisten tcp > /out/maps-gallery-xvfb.log 2>&1 & xp=$!
trap 'kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:97 LIBGL_ALWAYS_SOFTWARE=1
if [ "${KOL_MAP_WORLD_ONLY:-0}" = 1 ]; then
 mkdir -p /out/maps-world-runtime
 go test -p 2 ./internal/ui > /out/maps-world-ui-tests.log 2>&1
 go test -p 2 -tags sceneverify ./cmd/kol > /out/maps-sceneverify.log 2>&1
fi
KOL_FULL_MAPS=1 timeout 1500 "/out/bin/$binary" -test.run '^TestHDFullMaps$' -test.v > "/out/$binary-run.log" 2>&1
