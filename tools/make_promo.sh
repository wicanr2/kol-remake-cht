#!/usr/bin/env bash
# 推廣片：把實跑截圖 ＋ 遊戲自己的音效合成一支 66 秒的短片。
#
#   tools/make_promo.sh <輸出目錄>
#
# 素材界線（`rulebook/93`）：
#   * 畫面**全部是 remake 的實機截圖**，來自 `docs/verification/` 的可重播收據，
#     不是 mockup、不重畫。每一張的來源寫在輸出的 `sources.json`。
#   * 配樂**沒有**。KOL 的 DOS 版只有 TITLE 的 PC Speaker 音爆，全庫普查過
#     （WORKLIST 的音訊列），**原版沒有背景音樂可用**。鐵則是「找不到就回報，
#     不可自己合成頂替」，所以這支片只鋪遊戲自己的雷聲 cue，其餘留白。
#   * 那個雷聲（`assets/audio/title-thunder.ogg`）是依公開的 PIT channel 2 規格
#     產生的，manifest 分類為「remake 自有；可公開」——不是原版音訊的取樣。
#
# 合成在 `game-video:latest`（既有的共用 image，只讀不改）裡跑，限 2 核；
# 一律靜態圖 ＋ 淡入淡出，**不用 zoompan**（那個濾鏡的 `d` 是「每個輸入幀輸出
# d 幀」，配上前置 `fps` 會把 6 秒算成兩萬幀）。
set -euo pipefail

OUT="${1:?用法：make_promo.sh <輸出目錄>}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${KOL_VIDEO_IMAGE:-game-video:latest}"
case "$OUT" in /*) ABS="$OUT" ;; *) ABS="$REPO_ROOT/$OUT" ;; esac
mkdir -p "$ABS"

timeout "${KOL_PROMO_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory 3g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp \
  -v "$REPO_ROOT/docs:/docs:ro" \
  -v "$REPO_ROOT/assets:/assets:ro" \
  -v "$ABS:/out" \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=1g" \
  "$IMAGE" bash -c '
    set -euo pipefail
    bash /docs/promo/compose.sh
    cp /docs/promo/theme.sh /docs/promo/compose.sh /out/
  '

echo "$ABS"
