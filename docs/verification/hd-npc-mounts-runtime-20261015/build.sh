set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/npc-formal-cache GOPATH=/tmp/npc-formal-gopath
cp -a /go/pkg/mod /tmp/mod
gofmt -w /src/internal/assets/hd_npc_mounts.go /src/internal/assets/hd_npc_mounts_test.go /src/internal/ui/hd_npc_mounts.go /src/internal/ui/hd_npc_mounts_test.go /src/internal/ui/hd_pages.go /src/cmd/kol/hd.go
python3 - <<'PY'
from pathlib import Path
import shutil,json,hashlib
r=Path('/src');s=Path('/tmp/npc-formal-source');w=Path('/out');s.mkdir()
files=[r/'go.mod',r/'go.sum']+[p for d in ['cmd','internal']for p in (r/d).rglob('*.go')]
for p in files:
 t=s/p.relative_to(r);t.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,t)
inputs={str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest()for p in files}
(w/'formal-runtime-source-inputs.json').write_text(json.dumps(inputs,indent=2)+'\n')
for name in ['assets','docs','tools']:
 (s/name).symlink_to(r/name,target_is_directory=True)
gallery=(w/'full-gallery_test.go.txt').read_text().replace('useNPCPrototype=i==1;','').replace('/out/full-gallery-v1','/out/formal-gallery-v1')
gallery=gallery.replace('func TestFullNPCGallery(t *testing.T){','func TestFullNPCGallery(t *testing.T){\n if os.Getenv("KOL_FULL_NPC_GALLERY")!="1" {t.Skip("explicit controlled gallery only")}\n')
(s/'cmd/kol/full_gallery_test.go').write_text(gallery)
print('Formal source snapshot',len(inputs),'files',flush=True)
PY
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/formal-xvfb.log 2>&1 &
xpid=$!
trap 'kill "$xpid" 2>/dev/null || true; wait "$xpid" 2>/dev/null || true' EXIT INT TERM
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
cd /tmp/npc-formal-source
go build -trimpath -buildvcs=false -o /out/bin/formal-npc ./cmd/kol >/out/build-formal-npc.log 2>&1
sha256sum /out/bin/formal-npc
go test -count=1 ./internal/... ./cmd/... >/out/test-formal-all.log 2>&1
printf '%s\n' 'Formal all packages PASS'
go test -tags sceneverify -count=1 ./cmd/kol >/out/test-formal-sceneverify.log 2>&1
printf '%s\n' 'Formal sceneverify PASS'
go test -c -o /out/bin/formal-npc-gallery ./cmd/kol >/out/build-formal-gallery.log 2>&1
KOL_FULL_NPC_GALLERY=1 timeout 420s /out/bin/formal-npc-gallery -test.run '^TestFullNPCGallery$' -test.v >/out/run-formal-gallery-v1.log 2>&1
printf '%s\n' 'Formal full gallery rendered'
