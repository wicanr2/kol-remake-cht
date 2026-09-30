#!/usr/bin/env bash
# 跑「拿原版收據當預期值」的那幾支對拍測試。
#
#   tools/run_original_parity_tests.sh [<收據目錄>]
#
# 收據目錄預設 workplace/battle-planning-20260911/，裡面要有 dosgolem 正常路徑
# 產出的 battle-terrain.json；隊員 fixture 由 prepare_battle_state_fixture.py
# 產在 workplace/planning-parity-20260911/party-fixture.json。
#
# cmd/kol 連著 Ebitengine，套件初始化就要連得上 X（即使測試根本不開視窗），
# 所以照樣要 Xvfb。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
RECEIPTS="${1:-workplace/battle-planning-20260911}"
FIXTURE="${KOL_PARTY_FIXTURE:-workplace/planning-parity-20260911/party-fixture.json}"
# 多回合執行段收據（#9）：`tools/oracle/book_startup_test.go` 的 KOL_BOOK_EXECUTE 模式產出。
ROUNDS="${KOL_ROUNDS_RECEIPT:-workplace/battle-panels-20260930/battle-rounds.json}"
# 點陣圖編輯器收據（SPEC-036）：`tools/oracle/book_editor_test.go`（KOL_CHECK=editor）產出。
EDITOR="${KOL_EDITOR_RECEIPT:-workplace/bitmap-editor-20260929/bitmap-editor.json}"

for f in "$REPO_ROOT/$FIXTURE" "$REPO_ROOT/$RECEIPTS/battle-terrain.json" "$REPO_ROOT/$ROUNDS" "$REPO_ROOT/$EDITOR" \
         "$REPO_ROOT/$RECEIPTS/battle-state-records.json"; do
  [ -f "$f" ] || { echo "缺少 $f" >&2; exit 2; }
done

exec timeout "${KOL_TEST_TIMEOUT:-900}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp \
  -e "KOL_PARTY_STATE_INPUT=/src/$FIXTURE" \
  -e "KOL_TERRAIN_INPUT=/src/$RECEIPTS/battle-terrain.json" \
  -e "KOL_STATE_RECORDS_INPUT=/src/$RECEIPTS/battle-state-records.json" \
  -e "KOL_ROUNDS_INPUT=/src/$ROUNDS" \
  -e "KOL_BITMAP_EDITOR_INPUT=/src/$EDITOR" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=2g" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export PATH=$PATH:/usr/local/go/bin
    cp -a /go/pkg/mod /tmp/mod
    export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
    export GOPROXY=off GOSUMDB=off GOWORK=off GOFLAGS=-mod=mod
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    Xvfb :94 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X94 ] && break; sleep 0.1; done
    export DISPLAY=:94
    go test ./cmd/kol/ -run TestOriginalNormal -v -count=1 2>&1 | grep -vE "^XGB"
    go test ./internal/bitmapedit/ -run TestOriginalEditor -v -count=1
  '
