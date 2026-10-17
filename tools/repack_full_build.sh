#!/bin/bash
# 僅容器內使用；/src 唯讀，/work 為明確的本輪輸出目錄。
set -euo pipefail
test -f /.dockerenv
mode=${1:?native 或 macos}
version=${2:?完整版號}
[[ $version =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]
export PATH=/usr/local/go/bin:$PATH GOWORK=off GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
cp -a /go/pkg/mod /tmp/mod
mkdir /tmp/kol-source
cp -a /src/cmd /src/internal /src/go.mod /src/go.sum /tmp/kol-source/
cd /tmp/kol-source
go version > /work/$mode-go-version.txt
if [ "$mode" = native ]; then
  go build -trimpath -buildvcs=false -ldflags "-X main.buildVersion=$version" -o /work/kol ./cmd/kol > /work/linux-build.log 2>&1
  xvfb-run -a /work/kol -version > /work/linux-version.txt
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "-X main.buildVersion=$version" -o /work/kol.exe ./cmd/kol > /work/windows-build.log 2>&1
elif [ "$mode" = macos ]; then
  eval "$(osxcross-conf)"
  export CGO_ENABLED=1 GOOS=darwin MACOSX_DEPLOYMENT_TARGET=11.0
  for arch in arm64 amd64; do
    case $arch in
      arm64) prefix=arm64-apple-$OSXCROSS_TARGET ;;
      amd64) prefix=x86_64-apple-$OSXCROSS_TARGET ;;
    esac
    env GOARCH=$arch CC=$prefix-clang CXX=$prefix-clang++ CGO_CFLAGS=-mmacosx-version-min=11.0 CGO_LDFLAGS=-mmacosx-version-min=11.0 \
      go build -trimpath -buildvcs=false -ldflags "-X main.buildVersion=$version" -o /tmp/kol-$arch ./cmd/kol > /work/macos-$arch-build.log 2>&1
  done
  prefix=x86_64-apple-$OSXCROSS_TARGET
  $prefix-lipo -create /tmp/kol-arm64 /tmp/kol-amd64 -output /work/kol-macos
  $prefix-lipo -info /work/kol-macos > /work/macos-architectures.txt
  for arch in arm64 x86_64; do
    $prefix-lipo -thin $arch /work/kol-macos -output /tmp/kol-thin
    $prefix-otool -l /tmp/kol-thin > /work/macos-$arch-load.txt
    $prefix-otool -L /tmp/kol-thin > /work/macos-$arch-libs.txt
  done
else
  exit 2
fi
