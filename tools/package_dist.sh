#!/usr/bin/env bash
# 把 `dist-all/<版本>/{full,release}` 封成三個平台各自獨立的發行檔。
#
#   tools/package_dist.sh v.1.0.0-20260901
#
# 產物放在 `dist-all/<版本>/packages/`：
#
#   kol-<版本>-<包>-linux-x86_64.AppImage    單檔，chmod +x 就能跑
#   kol-<版本>-<包>-windows-amd64.zip        解開後雙擊 kol.exe
#   kol-<版本>-<包>-macos-universal.zip      解開後是 `武士傳說.app`
#
# `<包>` 是 full（含原作衍生素材，只留本機）或 release（通過權利審查，可散布）。
# 三個平台分開，因為一個包裡塞三支不同作業系統的執行檔對玩家沒有意義，
# 而且會讓下載量變成三倍。
#
# 全程在鎖定容器內進行。映像是 `docker/Dockerfile.package`（在建置映像上補
# `file`／`zip`，appimagetool 缺 `file` 就不動）。appimagetool 本身由
# `workplace/tools/` 提供（本機、不進版控），版本與雜湊記在 WORKLOG。
set -euo pipefail

check_version() {
  if ! [[ "$1" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]; then
    echo "版號必須使用 v.<主版>.<次版>.<修訂版>-YYYYMMDD" >&2
    return 2
  fi
}

if [ "${1:-}" = "--in-container" ]; then
  VERSION="${2:?缺版本}"
  check_version "$VERSION"
  DIST="/src/dist-all/$VERSION"
  OUT="$DIST/packages"
  if [ -e "$OUT" ] || [ -L "$OUT" ]; then
    echo "封包目錄已存在，拒絕覆寫：$OUT" >&2
    exit 2
  fi
  mkdir "$OUT"

  # 圖示是 remake 自己畫的純幾何盾牌與劍——**不含任何原版衍生像素**，所以
  # release 那一份也帶得走。沒有 CJK 字型可用，因此不放文字。
  ICON=/tmp/kol-icon.png
  convert -size 256x256 xc:'#0b1020' \
    -fill '#c9a227' -draw "polygon 128,18 220,60 220,140 128,238 36,140 36,60" \
    -fill '#0b1020' -draw "polygon 128,40 202,74 202,134 128,214 54,134 54,74" \
    -fill '#c9a227' -draw "rectangle 122,64 134,188" \
    -fill '#c9a227' -draw "rectangle 90,98 166,110" \
    "$ICON"

  for bundle in full release; do
    SRC="$DIST/$bundle"
    [ -d "$SRC" ] || continue
    echo "== $bundle"

    # ---------- Linux：AppImage ----------
    APPDIR="/tmp/$bundle.AppDir"
    rm -rf "$APPDIR"
    mkdir -p "$APPDIR/usr/bin"
    cp -a "$SRC"/. "$APPDIR/"
    mv "$APPDIR/kol" "$APPDIR/usr/bin/kol"
    rm -f "$APPDIR/kol.exe" "$APPDIR/kol-macos"
    cat > "$APPDIR/AppRun" <<'RUN'
#!/bin/sh
# AppImage 掛在暫時目錄，資源要從這裡找而不是玩家當下的工作目錄。
HERE="$(dirname "$(readlink -f "$0")")"
cd "$HERE" || exit 1
exec "$HERE/usr/bin/kol" "$@"
RUN
    chmod +x "$APPDIR/AppRun" "$APPDIR/usr/bin/kol"
    cat > "$APPDIR/kol.desktop" <<'DESK'
[Desktop Entry]
Type=Application
Name=武士傳說
Comment=Knights of Legend 繁體中文重製
Exec=AppRun
Icon=kol
Categories=Game;RolePlaying;
Terminal=false
DESK
    cp "$ICON" "$APPDIR/kol.png"
    cp "$ICON" "$APPDIR/.DirIcon"
    # `--runtime-file` 是必要的：容器沒有網路，appimagetool 預設會去 GitHub
    # 抓 runtime，抓不到就整個失敗。
    ARCH=x86_64 /opt/appimagetool --appimage-extract-and-run \
      --runtime-file /opt/runtime-x86_64 \
      "$APPDIR" "$OUT/kol-$VERSION-$bundle-linux-x86_64.AppImage"
    rm -rf "$APPDIR"

    # ---------- Windows：zip ----------
    WNAME="kol-$VERSION-$bundle-windows-amd64"
    WD="/tmp/win-$bundle/$WNAME"
    rm -rf "/tmp/win-$bundle"
    mkdir -p "$WD"
    cp -a "$SRC"/. "$WD/"
    rm -f "$WD/kol" "$WD/kol-macos"
    cat > "$WD/如何開始.txt" <<'TXT'
《武士傳說》繁體中文重製版 — Windows

1. 把整個資料夾解開到任何位置。
2. 雙擊 kol.exe。
3. 操作：方向鍵移動，A S Z X 斜走，逗號與句號移動下方圖示列的白框，
   Enter 執行，Esc 回上一頁。滑鼠可以直接點選單、清單與圖示列，
   右鍵等於 Esc，滾輪上下移動游標。
4. 住宿存檔會寫到 %AppData%\kol-remake-cht\save.dat。
   要讀回來請用 kol.exe -load "%AppData%\kol-remake-cht\save.dat"。
5. 介面預設繁體中文。要看與原版逐畫面對拍的英文版面，加 -lang en。

這一份沒有經過 Windows 真機驗收。
TXT
    python3 /src/tools/make_zip.py "$OUT/$WNAME.zip" "$WD" "$WNAME"
    rm -rf "/tmp/win-$bundle"

    # ---------- macOS：含 .app 的 zip ----------
    MNAME="kol-$VERSION-$bundle-macos-universal"
    MD="/tmp/mac-$bundle/$MNAME"
    APP="$MD/武士傳說.app"
    rm -rf "/tmp/mac-$bundle"
    mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
    cp "$SRC/kol-macos" "$APP/Contents/MacOS/kol"
    chmod +x "$APP/Contents/MacOS/kol"
    # 資源放進 Resources；程式啟動時工作目錄是 `/`，靠 useResourceDir 找過來。
    for item in assets docs LICENSE README.md DO-NOT-DISTRIBUTE.txt tools; do
      if [ -e "$SRC/$item" ]; then cp -a "$SRC/$item" "$APP/Contents/Resources/"; fi
    done
    convert "$ICON" "$APP/Contents/Resources/kol.icns"
    sed -e "s/@VERSION@/$VERSION/g" > "$APP/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>武士傳說</string>
  <key>CFBundleDisplayName</key><string>武士傳說</string>
  <key>CFBundleExecutable</key><string>kol</string>
  <key>CFBundleIdentifier</key><string>tw.wicanr2.kol-remake-cht</string>
  <key>CFBundleVersion</key><string>@VERSION@</string>
  <key>CFBundleShortVersionString</key><string>@VERSION@</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleIconFile</key><string>kol</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST
    cat > "$MD/如何開始.txt" <<'TXT'
《武士傳說》繁體中文重製版 — macOS

1. 解開之後把「武士傳說.app」拖到「應用程式」或任何位置。
2. 這個 app 沒有 Apple 開發者簽章，第一次開啟會被攔下來。放行方式：
   在終端機執行
       xattr -dr com.apple.quarantine "/path/to/武士傳說.app"
   或在「系統設定 → 隱私權與安全性」按「仍要打開」。
3. 操作：方向鍵移動，A S Z X 斜走，逗號與句號移動下方圖示列的白框，
   Enter 執行，Esc 回上一頁。滑鼠可以直接點選單、清單與圖示列，
   右鍵等於 Esc，滾輪上下移動游標。
4. 住宿存檔會寫到 ~/Library/Application Support/kol-remake-cht/save.dat。
5. 介面預設繁體中文。要看與原版逐畫面對拍的英文版面，加 -lang en。

這一份是在 Linux 上交叉編譯的 universal binary（arm64 + x86_64），
只通過靜態驗收，**沒有經過 macOS 真機驗收**。
TXT
    python3 /src/tools/make_zip.py "$OUT/$MNAME.zip" "$MD" "$MNAME"
    rm -rf "/tmp/mac-$bundle"
  done

  ls -la "$OUT"
  exit 0
fi

VERSION="${1:?用法：package_dist.sh <版本>}"
check_version "$VERSION"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# 儲存庫可見度閘門：不是 private 或查不到就停（失敗即關閉，`docs/RELEASE-RIGHTS-REVIEW.md`）。
"$REPO_ROOT/tools/check_repo_visibility.sh" || exit $?
IMAGE="${KOL_PACKAGE_IMAGE:-kol-package:20260907}"
TOOL="${KOL_APPIMAGETOOL:-$REPO_ROOT/workplace/tools/appimagetool}"
RUNTIME="${KOL_APPIMAGE_RUNTIME:-$REPO_ROOT/workplace/tools/runtime-x86_64}"

[ -d "$REPO_ROOT/dist-all/$VERSION" ] || {
  echo "沒有 dist-all/$VERSION，先跑 tools/build_dist.sh" >&2; exit 2; }
[ -x "$TOOL" ] || {
  echo "缺 appimagetool：$TOOL" >&2
  echo "放一份可執行的 appimagetool 到 workplace/tools/ 再跑一次。" >&2; exit 2; }
[ -f "$RUNTIME" ] || {
  echo "缺 AppImage runtime：$RUNTIME" >&2
  echo "從 https://github.com/AppImage/type2-runtime/releases 取 runtime-x86_64" >&2
  echo "放到 workplace/tools/ 再跑一次（容器無網路，不能讓它自己下載）。" >&2; exit 2; }

exec timeout "${KOL_GO_TIMEOUT:-1800}" docker run --rm \
  --network none \
  --memory 4g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp \
  -v "$REPO_ROOT:/src" -v "$TOOL:/opt/appimagetool:ro" \
  -v "$RUNTIME:/opt/runtime-x86_64:ro" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=3g" \
  "$IMAGE" \
  bash /src/tools/package_dist.sh --in-container "$VERSION"
