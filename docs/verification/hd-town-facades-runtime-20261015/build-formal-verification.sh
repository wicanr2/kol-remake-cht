set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/facade-verify-cache GOPATH=/tmp/facade-verify-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
python3 /out/prepare-formal-verification.py
python3 - <<'PY'
from pathlib import Path
import hashlib,json,shutil,tarfile
r=Path('/src');w=Path('/out');s=Path('/tmp/formal-source');s.mkdir()
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
inputs={str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths}
assert inputs==json.loads((w/'formal-source-inputs.json').read_text())
for p in paths:
 t=s/p.relative_to(r);t.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,t)
for name in ['assets','docs','tools']:(s/name).symlink_to(r/name,target_is_directory=True)
for name,target in [('formal-gallery.go.txt','facade_formal_gallery_test.go'),('formal-gui.go.txt','facade_formal_gui_test.go')]:shutil.copyfile(w/name,s/'cmd/kol'/target)
old=Path('/tmp/baseline-source');old.mkdir()
with tarfile.open(w/'baseline-source.tar')as t:
 assert all(not p.name.startswith('/') and '..'not in Path(p.name).parts and(p.isfile()or p.isdir())for p in t.getmembers())
 t.extractall(old)
for name in ['assets','docs','tools']:(old/name).symlink_to(r/name,target_is_directory=True)
PY
cd /tmp/formal-source
gofmt -w cmd/kol/facade_formal_gallery_test.go cmd/kol/facade_formal_gui_test.go
go test -c -o /out/bin/formal-facade-gui ./cmd/kol >/out/build-formal-gallery.log 2>&1
python3 - <<'PY'
from pathlib import Path
import hashlib,json
s=Path('/tmp/formal-source');paths=[s/'go.mod',s/'go.sum']+[p for d in ['cmd','internal']for p in(s/d).rglob('*.go')]
Path('/out/formal-verification-source-inputs.json').write_text(json.dumps({str(p.relative_to(s)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths},indent=2)+'\n')
PY
cd /tmp/baseline-source
go build -trimpath -buildvcs=false -o /out/bin/baseline ./cmd/kol >/out/build-facade-baseline.log 2>&1
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-formal-gallery.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 KOL_FACADE_GALLERY=1 KOL_FACADE_REQUIRE_ALL=1
timeout 300s /out/bin/formal-facade-gui -test.run '^TestFacadeGallery$' -test.v >/out/run-formal-gallery.log 2>&1
kill "$xpid" 2>/dev/null || true
wait "$xpid" 2>/dev/null || true
trap - EXIT INT TERM
export KOL_PAGE_OUTPUT=facade-normal KOL_PAGE_RESULT=facade-normal-results.json
python3 /out/normal-facades.py winged-regression human-slots
python3 /out/save-facades.py
for facade_case in twn1pics-003 twn5pics-019 twn7pics-013 miscpics-000; do
 KOL_FACADE_GUI_CASE="$facade_case" python3 /out/gui-facades.py
done
