#!/bin/bash
set -eu
cd /src
cp -a /go/pkg/mod /tmp/mod
export GOWORK=off DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
gofmt -w /out/formal-action-gallery_test.go.txt
go test -c -trimpath -overlay /out/formal-action-gallery-overlay.json -o /out/bin/formal-action-gallery.test ./internal/ui >/out/build-formal-action-gallery.log 2>&1
go build -trimpath -overlay /out/formal-action-gui-overlay.json -o /out/bin/gui-formal-actions ./cmd/kol >/out/build-formal-action-gui.log 2>&1
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-formal-action-gallery.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in 1 2 3 4 5 6 7 8 9 10; do
 if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
 sleep .1
done
/out/bin/formal-action-gallery.test -test.run '^TestFormalHDActionGallery$' -test.v -test.timeout 120s >/out/run-formal-action-gallery.log 2>&1
tail -4 /out/run-formal-action-gallery.log
