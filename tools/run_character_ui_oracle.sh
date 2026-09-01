#!/usr/bin/env bash
# 以有界 docker run 執行 KOLBOOK 角色建立 UI oracle。
#
#   tools/run_character_ui_oracle.sh <輸出目錄> [dosbox_character_ui_oracle.py 參數...]
#
# 原版 ZIP 唯讀掛載；遊戲副本建在容器 tmpfs，容器結束即消失。輸出目錄先由主機
# 以目前 UID/GID 建立，避免產生 root-owned 專案檔案。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_DOSBOX_IMAGE:-rich2-dosbox-dbg:latest}"
ZIP="${KOL_ORIGINAL_ZIP:-$REPO_ROOT/Knights-of-legend.zip}"

if [ $# -lt 1 ]; then
  echo "用法：$0 <輸出目錄> [oracle 參數...]" >&2
  exit 2
fi

OUT_DIR="$1"; shift
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"

STEPS_MOUNT=()
if [ -n "${KOL_STEPS_DIR:-}" ]; then
  STEPS_MOUNT=(-v "$(cd "$KOL_STEPS_DIR" && pwd):/steps:ro")
fi

timeout 900 docker run --rm \
  "${STEPS_MOUNT[@]}" \
  --network none \
  --memory 1g --cpus 1 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -v "$ZIP:/in/kol.zip:ro" \
  -v "$REPO_ROOT/tools/dosbox_character_ui_oracle.py:/opt/kol/dosbox_character_ui_oracle.py:ro" \
  -v "$REPO_ROOT/tools/x11_key_inject.py:/opt/kol/x11_key_inject.py:ro" \
  -v "$OUT_DIR:/out" \
  --tmpfs "/work:exec,uid=$(id -u),gid=$(id -g),size=256m" \
  "$IMAGE" \
  python3 /opt/kol/dosbox_character_ui_oracle.py --zip /in/kol.zip --work /work --out /out "$@"
