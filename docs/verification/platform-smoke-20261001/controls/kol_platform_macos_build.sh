set -euo pipefail
eval "$(osxcross-conf)"
cp -a /go/pkg/mod /tmp/mod
export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=mod
export CGO_ENABLED=1 GOOS=darwin MACOSX_DEPLOYMENT_TARGET=11.0
go version
for arch in arm64 amd64; do
  case "$arch" in
    arm64) task_compiler=arm64-apple-$OSXCROSS_TARGET ;;
    amd64) task_compiler=x86_64-apple-$OSXCROSS_TARGET ;;
  esac
  env GOARCH=$arch CC=$task_compiler-clang CXX=$task_compiler-clang++ \
    CGO_CFLAGS=-mmacosx-version-min=11.0 CGO_LDFLAGS=-mmacosx-version-min=11.0 \
    go build -trimpath -o "/work/workplace/platform-smoke-20261001/build/kol-macos-$arch" ./cmd/kol
done
task_apple_tool=x86_64-apple-$OSXCROSS_TARGET
$task_apple_tool-lipo -create /work/workplace/platform-smoke-20261001/build/kol-macos-arm64 /work/workplace/platform-smoke-20261001/build/kol-macos-amd64 -output /work/workplace/platform-smoke-20261001/build/kol-macos
$task_apple_tool-lipo -info /work/workplace/platform-smoke-20261001/build/kol-macos
for arch in arm64 amd64; do
  $task_apple_tool-otool -l /work/workplace/platform-smoke-20261001/build/kol-macos-$arch > "/work/workplace/platform-smoke-20261001/macos/$arch-load-commands.txt"
  $task_apple_tool-otool -L /work/workplace/platform-smoke-20261001/build/kol-macos-$arch > "/work/workplace/platform-smoke-20261001/macos/$arch-libraries.txt"
done
sha256sum /work/workplace/platform-smoke-20261001/build/kol-macos*
