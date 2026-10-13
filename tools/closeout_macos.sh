#!/bin/bash
# 僅容器內執行；專案/src唯讀，交付/out可寫。
set -euo pipefail
test -f /.dockerenv
version=$(cat /out/full/VERSION)
eval "$(osxcross-conf)"
cp -a /go/pkg/mod /tmp/mod
mkdir /tmp/kol-source
cp -a /src/cmd /src/internal /src/go.mod /src/go.sum /tmp/kol-source/
cd /tmp/kol-source
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off CGO_ENABLED=1 GOOS=darwin MACOSX_DEPLOYMENT_TARGET=11.0
go version >/out/smoke/macos-go-version.txt
for arch in arm64 amd64; do
  case $arch in
    arm64) prefix=arm64-apple-$OSXCROSS_TARGET ;;
    amd64) prefix=x86_64-apple-$OSXCROSS_TARGET ;;
  esac
  env GOARCH=$arch CC=$prefix-clang CXX=$prefix-clang++ CGO_CFLAGS=-mmacosx-version-min=11.0 CGO_LDFLAGS=-mmacosx-version-min=11.0 \
    go build -trimpath -buildvcs=false -ldflags "-X main.buildVersion=$version" -o /tmp/kol-$arch ./cmd/kol >"/out/smoke/macos-$arch-build.log" 2>&1
done
prefix=x86_64-apple-$OSXCROSS_TARGET
$prefix-lipo -create /tmp/kol-arm64 /tmp/kol-amd64 -output /out/full/kol-macos
$prefix-lipo -info /out/full/kol-macos >/out/smoke/macos-architectures.txt
for arch in arm64 x86_64; do
  $prefix-lipo -thin $arch /out/full/kol-macos -output /tmp/kol-thin
  $prefix-otool -l /tmp/kol-thin >"/out/smoke/macos-$arch-load.txt"
  $prefix-otool -L /tmp/kol-thin >"/out/smoke/macos-$arch-libs.txt"
done
cat /out/smoke/macos-architectures.txt
