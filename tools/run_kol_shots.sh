#!/usr/bin/env bash
# 在鎖定容器內一次產生多張 remake 截圖。
#
#   tools/run_kol_shots.sh <規格檔>
#
# 規格檔每行是「輸出路徑|參數|參數…」，欄位以 `|` 分隔，因此參數本身可以含空白
# （例如操作腳本）。空行與 `#` 開頭的行會略過。
#
# 一次啟動一個容器、建置一次、起一次 Xvfb，再逐行執行；分開跑十幾個容器只是把同樣的
# 建置成本重複十幾遍。容器與 tools/run_kol_headless.sh 用同一個鎖定映像與同一組資源
# 上限，輸出以目前 UID/GID 寫入，不會留下 root-owned 檔案。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"

if [ $# -lt 1 ]; then
  echo "用法：$0 <規格檔>" >&2
  exit 2
fi
SPECS="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"

# KOL_BIN 指向一個已建好的執行檔（例如從 AppImage 解出來的 usr/bin/kol）時不從原始碼
# 建置，直接用它跑；用來驗證封包裡的那一份。工作目錄仍是 /src，素材讀的是工作樹，
# 與封包內素材是否相同另由 verify_local_appimage.py 逐檔核對。
BIN_MOUNT=()
if [ -n "${KOL_BIN:-}" ]; then
  [ -f "$KOL_BIN" ] || { echo "找不到 KOL_BIN：$KOL_BIN" >&2; exit 2; }
  BIN_MOUNT=(-v "$(cd "$(dirname "$KOL_BIN")" && pwd)/$(basename "$KOL_BIN"):/prebuilt/kol:ro" -e KOL_PREBUILT=1)
fi

exec timeout "${KOL_GO_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w /src \
  -v "$SPECS:/specs.txt:ro" \
  "${BIN_MOUNT[@]}" \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=1g" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export PATH=$PATH:/usr/local/go/bin
    cp -a /go/pkg/mod /tmp/mod
    export GOMODCACHE=/tmp/mod GOCACHE=/tmp/gocache GOPATH=/tmp/gopath
    export GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
    # 這個容器沒有掛 dosgolem，`go.work` 的 `use /dosgolem` 會指到不存在的路徑；
    # 而且 workspace 模式不接受 `-mod=mod`。這裡不需要 oracle，直接關掉。
    export GOWORK=off
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    if [ -n "${KOL_PREBUILT:-}" ]; then
      cp /prebuilt/kol /tmp/kol
    else
      go build -o /tmp/kol ./cmd/kol
    fi
    Xvfb :92 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    XPID=$!
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X92 ] && break; sleep 0.1; done
    status=0
    while IFS= read -r line; do
      case "$line" in ""|\#*) continue;; esac
      IFS="|" read -r -a parts <<< "$line"
      out="${parts[0]}"
      args=("${parts[@]:1}")
      mkdir -p "$(dirname "$out")"
      if ! DISPLAY=:92 /tmp/kol "${args[@]}" >/dev/null 2>/tmp/shot.err; then
        echo "失敗：$out" >&2; cat /tmp/shot.err >&2; status=1; continue
      fi
      echo "$out"
    done < /specs.txt
    kill $XPID 2>/dev/null || true
    exit $status
  '
