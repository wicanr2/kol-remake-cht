set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/crest-cache GOPATH=/tmp/crest-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
cd /src
gofmt -w internal/assets/hd_town_crests.go internal/assets/hd_town_crests_test.go internal/ui/hd_town_crests.go internal/ui/hd_town_crests_test.go cmd/kol/hd_town_crests_test.go internal/ui/hd.go internal/ui/hd_interface.go internal/ui/hd_pages.go cmd/kol/hd.go
python3 - <<'PY'
from pathlib import Path
import hashlib,json,shutil,tarfile
r=Path('/src');w=Path('/out');sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
assert all(p.stat().st_uid==p.stat().st_gid==1000 for p in paths)
(w/'source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):sha(p) for p in paths},indent=2)+'\n')
(w/'asset-inputs.json').write_text(json.dumps({str(p.relative_to(r)):sha(p) for p in(r/'assets/hd').rglob('*')if p.is_file()},indent=2)+'\n')
for dest in [Path('/tmp/crest-formal'),Path('/tmp/crest-baseline')]:dest.mkdir()
s=Path('/tmp/crest-formal')
for p in paths:
 t=s/p.relative_to(r);t.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,t)
old=Path('/tmp/crest-baseline')
with tarfile.open(w/'baseline-source.tar')as t:
 assert all(not p.name.startswith('/') and '..'not in Path(p.name).parts and(p.isfile()or p.isdir())for p in t.getmembers())
 t.extractall(old)
for dest in [s,old]:
 for name in ['assets','docs','tools']:(dest/name).symlink_to(r/name,target_is_directory=True)
 shutil.copyfile(w/'gallery.go.txt',dest/'cmd/kol/crest_gallery_test.go')
PY
mkdir -p /out/bin /out/gallery
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-tests.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
go test -p 2 ./internal/... ./cmd/... >/out/full-test.log 2>&1
go test -p 2 -tags sceneverify ./cmd/kol >/out/sceneverify.log 2>&1
go build -trimpath -buildvcs=false -o /out/bin/formal ./cmd/kol >/out/build-formal.log 2>&1
printf '%s\n' 'Formal tests, sceneverify and runtime build PASS'
cd /tmp/crest-formal
gofmt -w cmd/kol/crest_gallery_test.go
go test -c -o /out/bin/formal-gallery ./cmd/kol >/out/build-formal-gallery.log 2>&1
python3 - <<'PY'
from pathlib import Path
import hashlib,json
r=Path('/tmp/crest-formal');paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
Path('/out/verification-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths},indent=2)+'\n')
PY
cd /tmp/crest-baseline
go build -trimpath -buildvcs=false -o /out/bin/baseline ./cmd/kol >/out/build-baseline.log 2>&1
gofmt -w cmd/kol/crest_gallery_test.go
go test -c -o /out/bin/baseline-gallery ./cmd/kol >/out/build-baseline-gallery.log 2>&1
export KOL_CREST_GALLERY=1
KOL_CREST_VARIANT=baseline timeout 120s /out/bin/baseline-gallery -test.run '^TestCrestGallery$' -test.v >/out/gallery-baseline.log 2>&1
KOL_CREST_VARIANT=formal timeout 120s /out/bin/formal-gallery -test.run '^TestCrestGallery$' -test.v >/out/gallery-formal.log 2>&1
kill "$xpid" 2>/dev/null || true
wait "$xpid" 2>/dev/null || true
trap - EXIT INT TERM
export KOL_PAGE_OUTPUT=normal KOL_PAGE_RESULT=normal-results.json
python3 /out/normal.py winged-regression human-slots
python3 /out/save.py
for crest_case in twn1pics-000 twn7pics-000; do
 KOL_CREST_GUI_CASE="$crest_case" python3 /out/gui.py
done
