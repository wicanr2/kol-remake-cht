#!/usr/bin/env bash
# 在有界容器內執行本專案的純 Python 工具（解包、比對、驗證）。
#
#   tools/run_python_tool.sh tools/xxx.py [參數...]
#
# 專案規則要求分析、解包、轉檔與驗證都在容器內進行。這支包裝用既有的鎖定映像
# （內含 Python 3.11，只讀不改），把工作樹掛進 /src，原版 ZIP 唯讀掛在 /in/kol.zip，
# 輸出以目前 UID/GID 寫入，不會留下 root-owned 檔案。預設無網路。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_PY_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"
ZIP="${KOL_ORIGINAL_ZIP:-$REPO_ROOT/Knights-of-legend.zip}"

ZIP_MOUNT=()
if [ -f "$ZIP" ]; then
  ZIP_MOUNT=(-v "$ZIP:/in/kol.zip:ro")
fi

exec timeout "${KOL_PY_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory "${KOL_PY_MEMORY:-2g}" --cpus "${KOL_PY_CPUS:-2}" --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  "${ZIP_MOUNT[@]}" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  python3 "$@"
