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
# 儲存庫可見度閘門：不是 private 或查不到就停（失敗即關閉，`docs/RELEASE-RIGHTS-REVIEW.md`）。
"$REPO_ROOT/tools/check_repo_visibility.sh" || exit $?
DIST="dist-all/$VERSION"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"

for bundle in full release promo; do
  [ -f "$REPO_ROOT/$DIST/$bundle/manifest.json" ] || {
    echo "$bundle 少了 manifest.json" >&2; exit 2; }
done
# 三個包與版本目錄的根都要有條款：拿到 tar.gz 的人看不到儲存庫，
# 包裡沒有條款就等於沒被授權（`rulebook/85`）。
[ -f "$REPO_ROOT/$DIST/LICENSE" ] || {
  echo "版本目錄的根少了 LICENSE" >&2; exit 2; }
for bundle in full release promo; do
  [ -f "$REPO_ROOT/$DIST/$bundle/LICENSE" ] || {
    echo "$bundle 少了 LICENSE" >&2; exit 2; }
done
# release 一個原版衍生位元組都不能有。
if find "$REPO_ROOT/$DIST/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
     -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) -print -quit | grep -q .; then
  echo "release 裡有不該散布的檔案" >&2; exit 3
fi

# Windows 那一份只驗得到「是不是 PE 執行檔」——真機驗收要另外做。
for bundle in full release; do
  exe="$REPO_ROOT/$DIST/$bundle/kol.exe"
  [ -f "$exe" ] || { echo "$bundle 少了 kol.exe" >&2; exit 2; }
  # 同樣避開 `… | grep -q`（見 verify_macos.sh 的註解）。
  [ "$(head -c 2 "$exe")" = MZ ] || { echo "$bundle 的 kol.exe 不是 PE" >&2; exit 3; }
done

# macOS 那一份走 `verify_macos.sh` 的四道（雙弧、arm64 的 ad-hoc 簽章、最低系統
# 版本、動態相依只落在系統目錄）。同樣**不是真機驗收**。
for bundle in full release; do
  mac="$DIST/$bundle/kol-macos"
  [ -f "$REPO_ROOT/$mac" ] || { echo "$bundle 少了 kol-macos" >&2; exit 2; }
  "$REPO_ROOT/tools/verify_macos.sh" "$mac" >/dev/null     || { echo "$bundle 的 kol-macos 沒過靜態驗收" >&2; exit 3; }
done

# 三個平台各自獨立的發行檔。玩家拿到的是這幾個，所以驗收要驗它們，
# 不只是驗 build 目錄。
PKG="$REPO_ROOT/$DIST/packages"
[ -d "$PKG" ] || { echo "少了 packages/，先跑 tools/package_dist.sh" >&2; exit 2; }
for bundle in full release; do
  for f in "kol-$VERSION-$bundle-linux-x86_64.AppImage" \
           "kol-$VERSION-$bundle-windows-amd64.zip" \
           "kol-$VERSION-$bundle-macos-universal.zip"; do
    [ -s "$PKG/$f" ] || { echo "少了 $f" >&2; exit 2; }
  done
done
"$REPO_ROOT/tools/verify_packages.py" "$REPO_ROOT/$DIST" "$VERSION" || exit 3

# promo 要有片子與它的來源說明。
[ -s "$REPO_ROOT/$DIST/promo/kol-promo.mp4" ] || { echo "promo 少了影片" >&2; exit 2; }
[ -s "$REPO_ROOT/$DIST/promo/sources.json" ] || { echo "promo 少了 sources.json" >&2; exit 2; }

# 啟動驗收：拿包裡的 Linux 執行檔畫一張封面，畫得出來才算啟動成功。
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
