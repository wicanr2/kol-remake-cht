#!/usr/bin/env bash
# 驗收 `dist-all/<版本>/`：啟動、manifest、授權文件與素材 deny-list。
#
#   tools/verify_dist.sh 0.1.0
#
# 專案契約要求每次打包都驗證這幾項。**這一支只驗 Linux 那一份**——Windows 與
# macOS 要真機（或至少對應的工具鏈），還沒做。
set -euo pipefail

VERSION="${1:?用法：verify_dist.sh <版本>}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="dist-all/$VERSION"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"

for bundle in full release promo; do
  [ -f "$REPO_ROOT/$DIST/$bundle/manifest.json" ] || {
    echo "$bundle 少了 manifest.json" >&2; exit 2; }
done
for bundle in full release; do
  [ -f "$REPO_ROOT/$DIST/$bundle/LICENSE" ] || {
    echo "$bundle 少了 LICENSE" >&2; exit 2; }
done
# release 一個原版衍生位元組都不能有。
if find "$REPO_ROOT/$DIST/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
     -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) -print -quit | grep -q .; then
  echo "release 裡有不該散布的檔案" >&2; exit 3
fi

# 啟動驗收：拿包裡的執行檔畫一張封面，畫得出來才算啟動成功。
exec timeout "${KOL_GO_TIMEOUT:-900}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -w "/src/$DIST/full" \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    Xvfb :93 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X93 ] && break; sleep 0.1; done
    DISPLAY=:93 ./kol -scene cover -shot /tmp/cover.png >/dev/null
    DISPLAY=:93 ./kol -scene ending -shot-dir /tmp/end >/dev/null
    test -s /tmp/cover.png
    test -s /tmp/end/panel-6.png
    echo "啟動驗收通過：封面與結局都畫得出來"
  '
