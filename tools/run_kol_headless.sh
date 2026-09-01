#!/usr/bin/env bash
# 在鎖定容器內以 Xvfb 執行 remake，供截圖與對拍使用。
#
#   tools/run_kol_headless.sh -scene contents -shot workplace/shots/contents.png
#
# Ebitengine 在套件初始化階段就會開 GLFW，因此即使只是列資產也需要 DISPLAY。
# 容器沒有 GPU，強制走 Mesa 軟體算繪，保證輸出與硬體無關、可重生。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"

exec timeout "${KOL_GO_TIMEOUT:-900}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=1g" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export PATH=$PATH:/usr/local/go/bin
    cp -a /go/pkg/mod /tmp/mod
    export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
    export GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    go build -o /tmp/kol ./cmd/kol
    # 映像沒有 xauth，xvfb-run 無法使用；直接起 Xvfb 並在結束後收掉。
    Xvfb :91 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    XPID=$!
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X91 ] && break; sleep 0.1; done
    set +e
    DISPLAY=:91 /tmp/kol "$@"
    STATUS=$?
    set -e
    kill $XPID 2>/dev/null || true
    exit $STATUS
  ' bash "$@"
