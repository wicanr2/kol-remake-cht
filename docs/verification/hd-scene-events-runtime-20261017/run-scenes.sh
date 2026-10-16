set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/cache GOPATH=/tmp/gopath GOMEMLIMIT=3GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import shutil,json,hashlib
r=Path('/src');w=Path('/out');lab=Path('/tmp/event-gallery');lab.mkdir()
for d in ['cmd','internal']:shutil.copytree(r/d,lab/d)
for name in ['go.mod','go.sum']:shutil.copyfile(r/name,lab/name)
for name in ['assets','docs','tools']:(lab/name).symlink_to(r/name,target_is_directory=True)
shutil.copyfile(r/'docs/verification/hd-town-crests-runtime-20261015/gallery.go.txt',lab/'cmd/kol/crest_preview_setup_test.go')
shutil.copyfile(w/'maps-gallery.go.txt',lab/'cmd/kol/full_maps_gallery_test.go')
shutil.copyfile(w/'scene-events-gallery.go.txt',lab/'cmd/kol/scene_events_gallery_test.go')
(lab/'internal/ui/hd_events_verification.go').write_text('package ui\n// Verification-only, never production.\nfunc(h *HD)SuspendSceneEventsVerification()func(){old:=h.sceneEvents;h.sceneEvents=nil;return func(){h.sceneEvents=old}}\nfunc(h *HD)SuspendTownMapVerification()func(){old:=h.townMaps;h.townMaps=nil;return func(){h.townMaps=old}}\n')
(lab/'internal/battle/hd_events_verification.go').write_text('package battle\n// Prepared display state only, no gameplay RNG or production flags.\nfunc(s *State)PrepareHDResultVerification(index int){s.Stage=StageDone;s.Plan.Prompt=PromptExecuting;s.fledFromBar=index==0;s.RankFight=index==1||index==4;s.Verdict=VerdictLost;if index==0{s.Verdict=VerdictFled};if index==3||index==4{s.Verdict=VerdictWon}}\n')
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in (r/d).rglob('*.go')]
(w/'scene-events-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(paths)},indent=2)+'\n')
p=json.loads((w/'plan.json').read_text());p['tools']=sorted(set(p['tools']+['scene-events-gallery.go.txt','run-scene-events.sh']));(w/'plan.json').write_text(json.dumps(p,ensure_ascii=False,indent=2)+'\n')
PY
cd /tmp/event-gallery
gofmt -w cmd/kol/scene_events_gallery_test.go cmd/kol/full_maps_gallery_test.go cmd/kol/crest_preview_setup_test.go internal/ui/hd_events_verification.go internal/battle/hd_events_verification.go
cp cmd/kol/scene_events_gallery_test.go /out/scene-events-verification-gallery.go.txt
cp internal/ui/hd_events_verification.go /out/scene-events-verification-ui.go.txt
cp internal/battle/hd_events_verification.go /out/scene-events-verification-battle.go.txt
mkdir -p /out/bin /out/scene-events-runtime
go test -c -o /out/bin/scene-events-gallery ./cmd/kol > /out/scene-events-build.log 2>&1
Xvfb :96 -screen 0 2048x1536x24 -nolisten tcp > /out/scene-events-xvfb.log 2>&1 & xp=$!
trap 'kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:96 LIBGL_ALWAYS_SOFTWARE=1
go test ./internal/assets -run '^TestHDSceneEvents' -v > /out/scene-events-33-asset-tests.log 2>&1
go test -p 2 ./internal/ui > /out/scene-events-ui-tests.log 2>&1
KOL_SCENE_EVENTS=1 timeout 600 /out/bin/scene-events-gallery -test.run '^TestHDSceneEventGallery$' -test.v > /out/scene-events-run.log 2>&1
