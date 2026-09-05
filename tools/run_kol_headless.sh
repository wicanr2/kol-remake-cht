#!/usr/bin/env bash
# 在容器裡跑事先編好的 kol，**標準輸出原樣送回來**。
#
#   tools/go_docker.sh build -o workplace/tmp/kol ./cmd/kol
#   tools/run_kol_headless.sh -scene play -load workplace/saves/hostel.json -route-to 59,28
#
# `run_kol_shots_fast.sh` 把 stdout 丟掉（它要的是截圖檔），所以會印東西的旗標
# （`-route-to`、`-dump-cjk-layout`）需要另一支。Ebitengine 在套件初始化就要
# 連得上 X，即使旗標在開視窗之前就 return，所以照樣要 Xvfb。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
BIN="${KOL_BIN:-workplace/tmp/kol}"

if [ ! -x "$REPO_ROOT/$BIN" ]; then
  echo "找不到 $BIN，先跑 tools/go_docker.sh build -o $BIN ./cmd/kol" >&2
  exit 2
fi
STALE="$(find "$REPO_ROOT/cmd" "$REPO_ROOT/internal" -name '*.go' \
  -newer "$REPO_ROOT/$BIN" -print -quit 2>/dev/null || true)"
if [ -n "$STALE" ]; then
  echo "$BIN 比原始碼舊（$STALE 在它之後改過），先重編" >&2
  exit 2
fi

exec timeout "${KOL_GO_TIMEOUT:-900}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    Xvfb :94 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    XPID=$!
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X94 ] && break; sleep 0.1; done
    DISPLAY=:94 "/src/'"$BIN"'" "$@"
    status=$?
    kill $XPID 2>/dev/null || true
    exit $status
  ' -- "$@"
