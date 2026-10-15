set -eu
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/monster-baseline-cache GOPATH=/tmp/monster-baseline-gopath GOMEMLIMIT=2GiB
cp -a /go/pkg/mod /tmp/mod
python3 - <<'PY'
from pathlib import Path
import tarfile
r=Path('/src');w=Path('/out');s=Path('/tmp/monster-baseline-source');s.mkdir()
with tarfile.open(w/'baseline-source.tar')as t:
 assert all(not p.name.startswith('/') and '..' not in Path(p.name).parts and(p.isfile()or p.isdir())for p in t.getmembers())
 t.extractall(s)
for name in ['assets','docs','tools']:(s/name).symlink_to(r/name,target_is_directory=True)
PY
cd /tmp/monster-baseline-source
go build -trimpath -buildvcs=false -o /out/bin/baseline ./cmd/kol >/out/build-baseline.log 2>&1
sha256sum /out/bin/baseline
python3 /out/prepare-regression.py
python3 /out/normal-monsters.py winged-regression human-slots
python3 /out/save-monsters.py
KOL_MONSTER_GUI_CASE=info python3 /out/gui-monsters.py
KOL_MONSTER_GUI_CASE=encounter python3 /out/gui-monsters.py
