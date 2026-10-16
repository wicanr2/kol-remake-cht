set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/cache GOPATH=/tmp/gopath GOMEMLIMIT=3GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import json,hashlib,shutil
r,w=Path('/src'),Path('/out');lab=Path('/tmp/motion-gallery');lab.mkdir()
for d in ['cmd','internal']:shutil.copytree(r/d,lab/d)
for p in ['go.mod','go.sum']:shutil.copyfile(r/p,lab/p)
for p in ['assets','docs','tools']:(lab/p).symlink_to(r/p,target_is_directory=True)
shutil.copyfile(r/'docs/verification/hd-town-crests-runtime-20261015/gallery.go.txt',lab/'cmd/kol/crest_preview_setup_test.go')
shutil.copyfile(w/'maps-gallery.go.txt',lab/'cmd/kol/full_maps_gallery_test.go')
shutil.copyfile(w/'scene-events-gallery.go.txt',lab/'cmd/kol/scene_events_gallery_test.go')
shutil.copyfile(w/'motion-gallery.go.txt',lab/'cmd/kol/motion_gallery_test.go')
shutil.copyfile(w/'scene-events-verification-ui.go.txt',lab/'internal/ui/hd_events_verification.go')
shutil.copyfile(w/'scene-events-verification-battle.go.txt',lab/'internal/battle/hd_events_verification.go')
(lab/'internal/ui/hd_motion_verification.go').write_text('package ui\nfunc(b *BattleTiles)PrepareMotionRingVerification(p int){b.ringPhase=p}\n')
p=lab/'internal/ui/hd_pages.go';s=p.read_text();a=s.index('\n\tvar op ebiten.DrawImageOptions',s.index('if painted != src'));s=s[:a]+'\n\tfmt.Printf("HD-NATIVE-FALLBACK key=%s bounds=%v pos=%d,%d scale=%d\\n",key,src.Bounds(),x,y,scale)'+s[a:];p.write_text(s)
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')];(w/'motion-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(paths)},indent=2)+'\n')
scope=json.loads((w/'plan.json').read_text());scope['tools']=sorted(set(scope['tools']+['motion-gallery.go.txt','run-motion-gallery.sh']));(w/'plan.json').write_text(json.dumps(scope,ensure_ascii=False,indent=2)+'\n')
PY
cd /tmp/motion-gallery
gofmt -w cmd/kol/motion_gallery_test.go cmd/kol/scene_events_gallery_test.go
cp cmd/kol/motion_gallery_test.go /out/motion-verification-gallery.go.txt
go test -c -o /out/bin/motion-gallery ./cmd/kol > /out/motion-gallery-build.log 2>&1
Xvfb :94 -screen 0 2048x1536x24 -nolisten tcp > /out/motion-gallery-xvfb.log 2>&1 & xp=$!
trap 'kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true; cat /sys/fs/cgroup/memory.peak > /out/motion-memory-peak.txt; cat /sys/fs/cgroup/memory.events > /out/motion-memory-events.txt' EXIT INT TERM
export DISPLAY=:94 LIBGL_ALWAYS_SOFTWARE=1
mkdir -p /out/motion-runtime
KOL_MOTION=1 timeout 600 /out/bin/motion-gallery -test.run '^TestHDMotionGallery$' -test.v > /out/motion-gallery-run.log 2>&1
python3 -c 'from pathlib import Path;print(Path("/out/motion-gallery-run.log").read_text()[-3000:])'
