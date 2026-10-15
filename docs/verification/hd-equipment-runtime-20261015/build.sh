set -eu
cd /src
cp -a /go/pkg/mod /tmp/mod
export GOWORK=off DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1
gofmt -w internal/hdart/equipment.go internal/assets/hd_equipment.go internal/ui/hd_equipment.go internal/ui/hd.go internal/ui/hd_pages.go internal/ui/itempage_cjk.go cmd/kol/main.go cmd/kol/hd.go cmd/kol/hd_equipment.go cmd/kol/hd_pages.go cmd/kol/hd_pages_test.go cmd/kol/hd_test.go
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-equipment-formal-build.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in 1 2 3 4 5 6 7 8 9 10; do
 if xdotool getdisplaygeometry >/dev/null 2>&1; then break; fi
 sleep .1
done
go test ./internal/... ./cmd/... >/out/go-test-equipment-formal.log 2>&1
go test -tags sceneverify ./cmd/kol >/out/sceneverify-equipment-formal.log 2>&1
go build -trimpath -o /out/bin/equipment-formal ./cmd/kol >/out/build-equipment-formal.log 2>&1
tail -5 /out/go-test-equipment-formal.log
tail -3 /out/sceneverify-equipment-formal.log
sha256sum /out/bin/equipment-formal
