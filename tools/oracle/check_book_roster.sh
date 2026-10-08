#!/usr/bin/env bash
# 只在 Docker 內執行。/src、/repo 唯讀；/out-root 是使用者所有的研究輸出。
set -euo pipefail
export PATH=/usr/local/go/bin:$PATH HOME=/tmp
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off
target_root="/out-root/roster-boundaries-20261001/${KOL_ROSTER_ATTEMPT:-review1}"
test ! -e "$target_root"
mkdir -p "$target_root"
export KOL_ROSTER_TARGET_ROOT="$target_root"
mkdir -p /tmp/dosgolem
# 現行 2f44a68 已撤掉 KOL 所需服務；此 probe 固定最近成功、68 份正式來源
# 皆與 ai-fuzz-attack-only-20260930/manifest.json 相同的既有版本，未改共用 repo。
export KOL_DOSGOLEM_REVISION=968ca774705853276ae55c8d1163c369481ac71f
git -c safe.directory=/repo -C /repo archive "$KOL_DOSGOLEM_REVISION" go.mod internal oracle | tar -x -C /tmp/dosgolem
python3 - <<'PY'
import hashlib, json, os, pathlib, subprocess
sources = [pathlib.Path('/src/go.mod'), pathlib.Path('/src/go.sum'), pathlib.Path('/tmp/dosgolem/go.mod')]
for directory in ['/src/internal', '/src/tools/oracle', '/src/workplace/oracle/game', '/tmp/dosgolem/internal', '/tmp/dosgolem/oracle']:
    sources += [p for p in pathlib.Path(directory).rglob('*') if p.is_file()]
records = [{'path': str(p), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted(set(sources))]
target = pathlib.Path(os.environ['KOL_ROSTER_TARGET_ROOT'])
payload = {'schema': 'kol-roster-source-inputs-v1', 'capture': 'before_copy', 'sources': records,
           'dosgolem_head': subprocess.check_output(['git', '-c', 'safe.directory=/repo', '-C', '/repo', 'rev-parse', 'HEAD'], text=True).strip(),
           'dosgolem_execution_revision': os.environ['KOL_DOSGOLEM_REVISION'],
           'image': 'kol-package:20260907', 'image_id': os.environ.get('KOL_IMAGE_ID'),
           'seed': 1, 'scope': '原版正常輸入邊界，未與重製版屬性或整場戰鬥對拍'}
(target / 'input-manifest.json').write_text(json.dumps(payload, ensure_ascii=False, indent=2) + '\n')
PY
mkdir -p /tmp/kol/tools/oracle /tmp/dosgolem
cp /src/go.mod /src/go.sum /tmp/kol/
cp /src/tools/oracle/book_startup_test.go /src/tools/oracle/book_roster_test.go /tmp/kol/tools/oracle/
cp -r /src/internal /tmp/kol/
cp -r /go/pkg/mod /tmp/mod
gofmt -w /tmp/kol/tools/oracle/book_roster_test.go
cd /tmp/kol
go work init . /tmp/dosgolem
export GOWORK=/tmp/kol/go.work
for scenario in ${KOL_ROSTER_SCENARIOS:-full duplicate charset}; do
    case "$scenario" in full|duplicate|charset) ;; *) exit 2 ;; esac
    target="$target_root/$scenario"
    mkdir -p "$target/game"
    cp -r /src/workplace/oracle/game/. "$target/game/"
    KOL_ORACLE_ROOT="$target/game" KOL_ORACLE_CAPTURE="$target" KOL_ROSTER_SCENARIO="$scenario" \
        go test -tags oracle ./tools/oracle -run '^TestBookRosterBoundariesOriginal$' -count=1 -v -timeout 600s > "$target/run.log" 2>&1
done
python3 - <<'PY'
import hashlib, json, os, pathlib
target = pathlib.Path(os.environ['KOL_ROSTER_TARGET_ROOT'])
path = target / 'input-manifest.json'
payload = json.loads(path.read_text())
payload['sources_unchanged'] = all(hashlib.sha256(pathlib.Path(r['path']).read_bytes()).hexdigest() == r['sha256'] for r in payload['sources'])
payload['outputs'] = [{'path': str(p.relative_to(target)), 'sha256': hashlib.sha256(p.read_bytes()).hexdigest(), 'uid': p.stat().st_uid, 'gid': p.stat().st_gid} for p in sorted(target.rglob('*')) if p.is_file() and p != path]
path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + '\n')
assert payload['sources_unchanged'], '執行期間來源改變'
assert all(r['uid'] == os.getuid() and r['gid'] == os.getgid() for r in payload['outputs']), '輸出擁有權不符'
PY
