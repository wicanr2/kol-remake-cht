#!/usr/bin/env bash
# 與 run_kol_shots.sh 相同，但用事先編好的執行檔，不在容器裡重編。
#
#   tools/go_docker.sh build -o workplace/tmp/kol ./cmd/kol
#   tools/run_kol_shots_fast.sh <規格檔>
#
# 用途是「同一條腳本要跑很多次」的探索：每一次重編一遍 Go 是主要成本，
# 而探索期間程式碼並沒有變。正式收據仍然用 run_kol_shots.sh 從原始碼重建。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
BIN="${KOL_BIN:-workplace/tmp/kol}"

if [ $# -lt 1 ]; then
  echo "用法：$0 <規格檔>" >&2
  exit 2
fi
if [ ! -x "$REPO_ROOT/$BIN" ]; then
  echo "找不到 $BIN，先跑 tools/go_docker.sh build -o $BIN ./cmd/kol" >&2
  exit 2
fi
SPECS="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"

exec timeout "${KOL_GO_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w /src \
  -v "$SPECS:/specs.txt:ro" \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    Xvfb :93 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    XPID=$!
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X93 ] && break; sleep 0.1; done
    status=0
    while IFS= read -r line; do
      case "$line" in ""|\#*) continue;; esac
      IFS="|" read -r -a parts <<< "$line"
      out="${parts[0]}"
      args=("${parts[@]:1}")
      mkdir -p "$(dirname "$out")"
      if ! DISPLAY=:93 "/src/'"$BIN"'" "${args[@]}" >/dev/null 2>/tmp/shot.err; then
        echo "失敗：$out" >&2; cat /tmp/shot.err >&2; status=1; continue
      fi
      echo "$out"
    done < /specs.txt
    kill $XPID 2>/dev/null || true
    exit $status
  '
