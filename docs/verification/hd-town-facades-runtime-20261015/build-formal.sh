set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/facade-formal-cache GOPATH=/tmp/facade-formal-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import hashlib,json
r=Path('/src');w=Path('/out')
paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
assert all(p.stat().st_uid==p.stat().st_gid==1000 for p in paths)
(w/'formal-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths},indent=2)+'\n')
paths=[p for p in(r/'assets/hd').rglob('*')if p.is_file()]
(w/'formal-asset-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths},indent=2)+'\n')
PY
cd /src
gofmt -w internal/assets/hd_town_facades.go internal/assets/hd_town_facades_test.go internal/ui/hd_town_facades.go internal/ui/hd_town_facades_test.go cmd/kol/hd_town_facades_test.go internal/ui/hd.go internal/ui/hd_interface.go internal/ui/hd_pages.go cmd/kol/hd.go
python3 - <<'PY'
from pathlib import Path
import hashlib,json
r=Path('/src');w=Path('/out');paths=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in(r/d).rglob('*.go')]
(w/'formal-source-inputs.json').write_text(json.dumps({str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in paths},indent=2)+'\n')
PY
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-formal.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
go test -p 2 ./internal/... ./cmd/... >/out/full-test-formal.log 2>&1
go test -p 2 -tags sceneverify ./cmd/kol >/out/sceneverify-formal.log 2>&1
mkdir -p /out/bin
go build -o /out/bin/kol-formal-town-facades ./cmd/kol >/out/build-formal.log 2>&1
printf '%s\n' 'Formal source tests, sceneverify and runtime build passed'
