#!/usr/bin/env bash
# 在有界容器內烘中文 UI 字模表。
#
#   tools/run_cjk_font_build.sh [build_cjk_font.py 參數...]
#
# 映像 `eob-remake-audio:20260901` 是既有的共用映像（內含 ffmpeg），本專案只讀不改；
# 映像本身沒有中文字型，因此把主機的 Noto 字型目錄唯讀掛進來，並在容器 tmpfs 重建
# fontconfig 快取。容器不碰主機字型、不寫主機家目錄，結束即刪。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_FFMPEG_IMAGE:-eob-remake-audio:20260901}"
FONT_DIR="${KOL_FONT_DIR:-/usr/share/fonts/opentype/noto}"

if [ ! -d "$FONT_DIR" ]; then
  echo "找不到字型目錄 $FONT_DIR" >&2
  exit 2
fi

exec timeout "${KOL_FONT_TIMEOUT:-600}" docker run --rm \
  --network none \
  --memory 1g --cpus 1 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" \
  -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w /src \
  -v "$FONT_DIR:/usr/share/fonts/opentype/noto:ro" \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    fc-cache -f >/dev/null 2>&1 || true
    python3 tools/build_cjk_font.py "$@"
  ' bash "$@"
