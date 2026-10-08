#!/usr/bin/env bash
# 在 Linux 上交叉編出 macOS 的 `kol`（arm64 ＋ x86_64，最後 lipo 成 universal）。
#
#   tools/build_macos.sh <輸出路徑>
#   KOL_BUILD_VERSION=<完整版號> tools/build_macos.sh <輸出路徑>
#
# 為什麼要單獨一支：Ebitengine 的 macOS 後端走 Cocoa／Metal，**一定要 CGO**，
# 所以 `CC` 得指到 osxcross 的 clang wrapper；而那份工具鏈的執行檔要 glibc 2.38，
# 本專案鎖定的 Go image 是 Debian 12（2.36）跑不動，只好另外一個 image
#（`docker/Dockerfile.osxcross-go`，Ubuntu 24.04 起底再把 Go 與 module cache 搬過去）。
#
# 前綴帶 SDK 次版號（SDK 15.5 → `darwin24.5`），**從 `osxcross-conf` 讀，不寫死**。
#
# ⚠ **Linux 上執行不了 macOS binary。** 這支跑完只做得到靜態驗收
#（`tools/verify_macos.sh`）：雙弧、arm64 有 ad-hoc 簽章、最低系統版本、
# 沒有連到系統目錄以外的動態庫。**那不等於功能正常**，真機驗收要另外做。
set -euo pipefail

OUT="${1:?用法：build_macos.sh <輸出路徑>}"
if [ -n "${KOL_BUILD_VERSION:-}" ] && \
   ! [[ "$KOL_BUILD_VERSION" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]; then
  echo "版號必須使用 v.<主版>.<次版>.<修訂版>-YYYYMMDD" >&2
  exit 2
fi
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_MAC_IMAGE:-kol-osxcross-go:20260907}"
CPUS="${KOL_GO_CPUS:-2}"
MIN="${KOL_MACOS_MIN:-11.0}"

case "$OUT" in /*) ABS="$OUT" ;; *) ABS="$REPO_ROOT/$OUT" ;; esac
mkdir -p "$(dirname "$ABS")"
REL="${ABS#"$REPO_ROOT"/}"

exec timeout "${KOL_MAC_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory 6g --cpus "$CPUS" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e "KOL_MACOS_MIN=$MIN" -e "KOL_OUT=$REL" \
  -e "KOL_BUILD_VERSION=${KOL_BUILD_VERSION:-}" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=2g" \
  "$IMAGE" \
  bash -c '
    set -euo pipefail
    eval "$(osxcross-conf)"
    cp -a /go/pkg/mod /tmp/mod
    export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
    export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off GOFLAGS=-mod=mod
    export CGO_ENABLED=1 GOOS=darwin
    export MACOSX_DEPLOYMENT_TARGET=$KOL_MACOS_MIN
    link_flags=()
    if [ -n "${KOL_BUILD_VERSION:-}" ]; then
      link_flags=(-ldflags "-X main.buildVersion=$KOL_BUILD_VERSION")
    fi
    for arch in arm64 amd64; do
      case $arch in
        arm64)  pre=arm64-apple-$OSXCROSS_TARGET ;;
        amd64)  pre=x86_64-apple-$OSXCROSS_TARGET ;;
      esac
      echo "== $arch（$pre）"
      env GOARCH=$arch CC=$pre-clang CXX=$pre-clang++ \
          CGO_CFLAGS="-mmacosx-version-min=$KOL_MACOS_MIN" \
          CGO_LDFLAGS="-mmacosx-version-min=$KOL_MACOS_MIN" \
        go build -trimpath "${link_flags[@]}" -o /tmp/kol-$arch ./cmd/kol
    done
    x86_64-apple-$OSXCROSS_TARGET-lipo -create /tmp/kol-arm64 /tmp/kol-amd64 \
      -output "/src/$KOL_OUT"
    x86_64-apple-$OSXCROSS_TARGET-lipo -info "/src/$KOL_OUT"
  '
