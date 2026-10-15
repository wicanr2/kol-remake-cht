set -eu
cd /src
cp -a /go/pkg/mod /tmp/mod
export GOWORK=off DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
gofmt -w /out/equipment-formal-gallery_test.go.txt /out/gui-main-equipment-formal.go.txt
go test -c -trimpath -overlay /out/equipment-formal-gallery-overlay.json -o /out/bin/equipment-formal-gallery.test ./internal/ui >/out/build-equipment-formal-gallery.log 2>&1
go build -trimpath -overlay /out/gui-equipment-formal-overlay.json -o /out/bin/gui-equipment-formal ./cmd/kol >/out/build-gui-equipment-formal.log 2>&1
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-equipment-formal-gallery.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in 1 2 3 4 5 6 7 8 9 10; do
 if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
 sleep .1
done
/out/bin/equipment-formal-gallery.test -test.run '^TestHDEquipmentMatrixV7$' -test.v -test.timeout 180s >/out/run-equipment-formal-gallery.log 2>&1
tail -5 /out/run-equipment-formal-gallery.log
