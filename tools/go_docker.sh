#!/usr/bin/env bash
# 在鎖定的 Go／Ebitengine 容器內執行 go 指令。
#
#   tools/go_docker.sh build ./...
#   tools/go_docker.sh test ./...
#
# 映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`（Go 1.26.7、Ebitengine v2.9.9）是既有的
# 共用映像，本專案只讀不改。映像內的 module cache 屬 root，因此每次執行先複製到容器
# tmpfs（約 55 MB）再指向該副本，避免以 root 執行而在專案裡留下 root-owned 檔案。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
CPUS="${KOL_GO_CPUS:-2}"

exec timeout "${KOL_GO_TIMEOUT:-900}" docker run --rm \
  --network none \
  --memory 3g --cpus "$CPUS" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp -e "KOL_GO_ENV=${KOL_GO_ENV:-}" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=1g" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export PATH=$PATH:/usr/local/go/bin
    cp -a /go/pkg/mod /tmp/mod
    export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
    export GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
    # KOL_GO_ENV 讓呼叫端加交叉編譯用的環境變數（例如 GOOS=windows）。
    if [ -n "${KOL_GO_ENV:-}" ]; then export ${KOL_GO_ENV}; fi
    go "$@"
  ' bash "$@"
