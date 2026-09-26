#!/usr/bin/env bash
# 在**單一容器**裡跑 `drive_battle_script.py`。
#
#   tools/go_docker.sh build -o workplace/tmp/kol ./cmd/kol
#   tools/run_driver.sh <seed> <存檔> <起始 token 檔> <輸出> [輪數] [世界目標] [城鎮目標] [城鎮邊界]
#
# 為什麼要有這一支：driver 每一輪都要跑一次完整腳本，而每一輪開一個容器要三秒。
# 一場戰鬥上百輪，光是開關容器就佔掉大半時間。這裡把容器開一次，Xvfb 起一次，
# driver 在裡面直接呼叫預編執行檔。
#
# 環境變數與 `drive_battle_script.py` 相同（`KOL_DRIVE_DIR`、`KOL_QUEST_SITES`），
# 另外 `KOL_DRIVE_TIMEOUT` 是整支的逾時（預設 3600 秒）。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
BIN="${KOL_BIN:-workplace/tmp/kol}"

if [ ! -x "$REPO_ROOT/$BIN" ]; then
  echo "找不到 $BIN，先跑 tools/go_docker.sh build -o $BIN ./cmd/kol" >&2
  exit 2
fi
# 與 run_kol_shots_fast.sh 同一道保險：執行檔比原始碼舊就拒絕跑。
STALE="$(find "$REPO_ROOT/cmd" "$REPO_ROOT/internal" -name '*.go' \
  -newer "$REPO_ROOT/$BIN" -print -quit 2>/dev/null || true)"
if [ -n "$STALE" ]; then
  echo "$BIN 比原始碼舊（$STALE 在它之後改過），先重編：" >&2
  echo "  tools/go_docker.sh build -o $BIN ./cmd/kol" >&2
  exit 2
fi

exec timeout "${KOL_DRIVE_TIMEOUT:-3600}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -e "KOL_BIN=/src/$BIN" \
  -e "KOL_DRIVE_DIR=${KOL_DRIVE_DIR:-workplace/tmp/drive}" \
  -e "KOL_QUEST_SITES=${KOL_QUEST_SITES:-}" \
  -e "KOL_SCENE=${KOL_SCENE:-flow}" \
  -e "KOL_SCRIPT_FLAG=${KOL_SCRIPT_FLAG:--flow-script}" \
  -e "KOL_MONSTERS=${KOL_MONSTERS:-}" \
  -e "KOL_AMBUSH_ROLLS=${KOL_AMBUSH_ROLLS:-}" \
  -e "KOL_STOP_AFTER_BATTLE=${KOL_STOP_AFTER_BATTLE:-}" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
    Xvfb :94 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X94 ] && break; sleep 0.1; done
    export DISPLAY=:94
    exec python3 tools/drive_battle_script.py "$@"
  ' -- "$@"
