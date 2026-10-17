set -euo pipefail
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOCACHE=/out/gocache GOPATH=/tmp/gopath GOMODCACHE=/tmp/mod GOPROXY=off GOSUMDB=off
cp -a /go/pkg/mod /tmp/mod
mkdir /tmp/kol-source
cp -a /src/cmd /src/internal /src/go.mod /src/go.sum /tmp/kol-source/
cd /tmp/kol-source
gofmt -w cmd/kol/hd*.go cmd/kol/main.go cmd/kol/runtime_log.go cmd/kol/display_settings.go internal/assets/hd*.go internal/ui/hd*.go internal/ui/display_settings.go
go build -trimpath -buildvcs=false -ldflags '-X main.buildVersion=v.1.0.4-20261017' -o /out/kol ./cmd/kol
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags '-X main.buildVersion=v.1.0.4-20261017' -o /out/kol.exe ./cmd/kol
ln -s /src/assets assets
ln -s /src/docs docs
xvfb-run -a go test ./internal/assets -run TestDeferredHD -count=1
xvfb-run -a go test ./internal/ui -run TestHDCache -count=1
