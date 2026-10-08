set -euo pipefail
export PATH=$PATH:/usr/local/go/bin
cp -a /go/pkg/mod /tmp/mod
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=mod
go version
for platform in linux windows; do
  case "$platform" in
    linux) export GOOS=linux GOARCH=amd64 CGO_ENABLED=1; suffix= ;;
    windows) export GOOS=windows GOARCH=amd64 CGO_ENABLED=0; suffix=.exe ;;
  esac
  go build -trimpath -o "/work/workplace/platform-smoke-20261001/build/kol-$platform$suffix" ./cmd/kol
  go build -trimpath -o "/work/workplace/platform-smoke-20261001/build/kol-$platform-second$suffix" ./cmd/kol
  cmp "/work/workplace/platform-smoke-20261001/build/kol-$platform$suffix" "/work/workplace/platform-smoke-20261001/build/kol-$platform-second$suffix"
  sha256sum "/work/workplace/platform-smoke-20261001/build/kol-$platform$suffix"
  go version -m "/work/workplace/platform-smoke-20261001/build/kol-$platform$suffix"
done
