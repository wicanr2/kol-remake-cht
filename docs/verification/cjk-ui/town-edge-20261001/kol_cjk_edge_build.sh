set -euo pipefail
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=mod
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
cp -a /go/pkg/mod /tmp/mod
export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
Xvfb :94 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in {1..50}; do
    [ -e /tmp/.X11-unix/X94 ] && break
    sleep 0.1
done
export DISPLAY=:94
gofmt -w internal/ui/outdoor.go internal/ui/idle_party_test.go
go test ./cmd/kol ./internal/... | tee workplace/cjk-edge-20261001/go-test.log
go build -trimpath -o workplace/cjk-edge-20261001/kol-edge-fixed ./cmd/kol
python3 /control/kol_cjk_edge_run.py after workplace/cjk-edge-20261001/kol-edge-fixed
KOL_BIN=/src/workplace/cjk-edge-20261001/kol-edge-fixed KOL_REPLAY_CHECK_DIR=workplace/cjk-edge-20261001/english-regression python3 tools/check_player_paths.py full town town-edge town-leave | tee workplace/cjk-edge-20261001/english-regression.log
