#!/bin/bash
set -eu
cd /src
test "$(stat -c %u /out)" = 1000
cp -a /go/pkg/mod /tmp/mod
gofmt -w internal/assets/hd_actions.go internal/assets/hd_actions_test.go internal/ui/hd_actions.go internal/ui/hd_actions_test.go internal/ui/hd.go internal/ui/hd_pages.go internal/ui/hd_interface.go cmd/kol/hd.go cmd/kol/hd_test.go
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-formal-actions.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true' EXIT
export DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 GOWORK=off
for attempt in 1 2 3 4 5 6 7 8 9 10; do
 if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
 sleep .1
done
go test ./internal/... ./cmd/... >/out/go-test-formal-actions.log 2>&1
go test -tags sceneverify ./cmd/kol >/out/sceneverify-formal-actions.log 2>&1
go build -trimpath -o /out/bin/formal-actions ./cmd/kol >/out/build-formal-actions.log 2>&1
tail -4 /out/go-test-formal-actions.log
tail -4 /out/sceneverify-formal-actions.log
sha256sum /out/bin/formal-actions
