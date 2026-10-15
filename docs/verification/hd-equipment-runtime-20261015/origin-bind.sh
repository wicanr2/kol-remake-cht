set -eu
cd /src
cp -a /go/pkg/mod /tmp/mod
export GOWORK=off DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
gofmt -w internal/ui/hd_equipment.go
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-equipment-origin-bind.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in 1 2 3 4 5 6 7 8 9 10; do
 if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
 sleep .1
done
go test ./internal/ui ./cmd/kol >/out/go-test-equipment-origin-bind.log 2>&1
go build -trimpath -o /out/bin/equipment-formal-origin-bind ./cmd/kol >/out/build-equipment-origin-bind.log 2>&1
python3 - <<'PY'
from pathlib import Path
from PIL import Image,ImageChops
import subprocess,json,hashlib
r=Path('/src');w=Path('/out');d=w/'equipment-origin-bind';d.mkdir()
old=w/'equipment-formal-normal/human-slots/prototype-hd'
cmd=json.loads((old/'command.json').read_text());cmd[0]='/out/bin/equipment-formal-origin-bind'
for flag,path in [('-save',d/'save.dat'),('-shot-dir',d/'shots'),('-trace',d/'trace.json'),('-text-log',d/'text.jsonl')]:
 cmd[cmd.index(flag)+1]=str(path)
(d/'command.json').write_text(json.dumps(cmd,ensure_ascii=False,indent=2)+'\n')
with (d/'run.log').open('w') as log:subprocess.run(cmd,cwd=r,stdout=log,stderr=log,timeout=120,check=True)
assert json.loads((d/'trace.json').read_text())==json.loads((old/'trace.json').read_text())
assert (d/'text.jsonl').read_bytes()==(old/'text.jsonl').read_bytes()
maximum=0
for p in sorted((old/'shots').glob('*.png')):
 delta=ImageChops.difference(Image.open(p).convert('RGBA'),Image.open(d/'shots'/p.name).convert('RGBA'));mx=max(x[1] for x in delta.getextrema());assert mx<=1,(p.name,mx);assert delta.getchannel('A').getbbox() is None;maximum=max(maximum,mx)
layout=json.loads((r/'assets/original-derived-data/portrait-layout.json').read_text());assert layout['body'][0]['x']==15
result=dict(status='PASS',reason='Use runtime-derived origin for body0 instead of duplicated prototype literal; exact arithmetic for the pinned layout',normal_frames=31,states_equal=True,chinese_record_equal=True,full_hd_max_delta=maximum,pinned_origin_equal_to_prior_reference=True,binary_sha256=hashlib.sha256((w/'bin/equipment-formal-origin-bind').read_bytes()).hexdigest())
(w/'equipment-origin-bind-check.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps(result))
PY
