#!/usr/bin/env bash
# 從乾淨的工作樹建出 `dist-all/<版本>/` 的三個包（專案契約：`full`／`release`／`promo`）。
#
#   tools/build_dist.sh v.1.0.0-20260901
#
# 建置在鎖定的 Go 容器裡跑，輸出以目前 UID/GID 寫入。三個包的界線：
#
#   full/     含原作衍生素材的完整版，**只留本機**，不得上傳、不得散布。
#   release/  通過權利審查的那一份：只有 remake 自己的程式碼、文件與抽取器，
#             **一個原版衍生位元組都不帶**；玩家用自己的原版磁片跑抽取器產生素材。
#   promo/    影片、metadata、來源與影音驗收。目前只有骨架。
#
# 從乾淨輸入建到新的版本目錄；既有版本保留，不覆寫。
set -euo pipefail

check_version() {
  if ! [[ "$1" =~ ^v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}$ ]]; then
    echo "版號必須使用 v.<主版>.<次版>.<修訂版>-YYYYMMDD" >&2
    return 2
  fi
}

# 本機 Linux 驗收包：由呼叫端提供隔離容器與唯讀 /src、可寫 /out。
if [ "${1:-}" = "--local-linux-in-container" ]; then
  check_version "${2:?缺版本}"
  exec python3 /src/tools/build_local_appimage.py --version "${2:?缺版本}" --output /out
fi

# 內部組裝入口：主機已完成可見度與乾淨工作樹檢查；/src唯讀、/out可寫。
if [ "${1:-}" = "--stage-in-container" ]; then
  VERSION="${2:?缺版本}"
  check_version "$VERSION"
  PHASE="${3:?缺組裝階段}"
  [ -f /.dockerenv ] && [ -d /src ] && [ -d /out ] || {
    echo "組裝只能在隔離Docker的/src與/out內執行" >&2; exit 2; }
  [ "$(stat -c '%u:%g' /out)" = "$(id -u):$(id -g)" ] || {
    echo "輸出目錄UID/GID不符" >&2; exit 2; }
  REPO_ROOT=/src
  OUT=/out
  case "$PHASE" in
    prepare)
      [ -z "$(find "$OUT" -mindepth 1 -maxdepth 1 -print -quit)" ] || {
        echo "版本輸出必須為全新空目錄" >&2; exit 2; }
      mkdir "$OUT"/{full,release,promo}
      ;;
    files)
      cp "$OUT/full/kol" "$OUT/full/kol.exe" "$OUT/full/kol-macos" "$OUT/release/"

      # 2. 兩邊都帶的文件。條款另外放一份在版本目錄的根，因為拿到 tar.gz 的人看不到
      #    儲存庫——包裡沒有條款就等於沒被授權（`rulebook/85`）。
      cp "$REPO_ROOT/LICENSE" "$OUT/"
      cp "$REPO_ROOT/LICENSE" "$OUT/promo/"
      for dir in full release; do
        cp "$REPO_ROOT/LICENSE" "$OUT/$dir/"
        cp "$REPO_ROOT/README.md" "$OUT/$dir/"
        mkdir -p "$OUT/$dir/docs"
        cp -a "$REPO_ROOT/docs/manual" "$OUT/$dir/docs/"
      done

      # 3. runtime 要的幾個 JSON（`cmd/kol` 的旗標預設值）。它們是 remake 自己的
      #    typed data 與譯文，兩邊都帶；`character-creation.json` 是原版衍生的，
      #    只進 full。
      for dir in full release; do
        mkdir -p "$OUT/$dir/docs/re" "$OUT/$dir/docs/i18n"
        cp "$REPO_ROOT"/docs/re/title-sequence.json \
           "$REPO_ROOT"/docs/re/town-inn-sample.json \
           "$REPO_ROOT"/docs/re/town-outdoor-sample.json "$OUT/$dir/docs/re/"
        cp -a "$REPO_ROOT"/docs/i18n/. "$OUT/$dir/docs/i18n/"
      done

      # 4. release 另外帶抽取器——玩家拿自己的原版磁片產生素材。
      mkdir -p "$OUT/release/tools"
      cp "$REPO_ROOT"/tools/*.py "$OUT/release/tools/"
      cp "$REPO_ROOT"/tools/run_python_tool.sh "$OUT/release/tools/"

      # 5. full 帶素材（只留本機）。
      mkdir -p "$OUT/full/assets"
      cp -a "$REPO_ROOT/assets/original-derived-png" "$OUT/full/assets/"
      cp -a "$REPO_ROOT/assets/original-derived-data" "$OUT/full/assets/"
      [ -d "$REPO_ROOT/assets/ui-font" ] && cp -a "$REPO_ROOT/assets/ui-font" "$OUT/full/assets/"
      [ -d "$REPO_ROOT/assets/audio" ] && cp -a "$REPO_ROOT/assets/audio" "$OUT/full/assets/"
      printf '%s\n' \
        "這一份含原作衍生素材，只留本機。" \
        "不得上傳、不得散布、不得放進 GitHub Release。" \
        > "$OUT/full/DO-NOT-DISTRIBUTE.txt"

      # 7. deny-list：release 裡不得有任何原版衍生位元組。
      if find "$OUT/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
            -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) -print -quit | grep -q .; then
        echo "release 裡出現了不該散布的檔案：" >&2
        find "$OUT/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
            -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) >&2
        exit 3
      fi

      ;;
    manifest)
      python3 /src/tools/write_dist_manifest.py \
        --dist /out --version "$VERSION" \
        --platform linux/amd64 --platform windows/amd64 \
        --platform darwin/amd64 --platform darwin/arm64
      ;;
    *) echo "未知組裝階段：$PHASE" >&2; exit 2 ;;
  esac
  exit 0
fi

VERSION="${1:?用法：build_dist.sh <版本>}"
check_version "$VERSION"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# 儲存庫可見度閘門：不符合目前public決定或查不到就停（docs/RELEASE-RIGHTS-REVIEW.md）。
"$REPO_ROOT/tools/check_repo_visibility.sh" || exit $?
OUT="$REPO_ROOT/dist-all/$VERSION"

if [ -n "$(git -C "$REPO_ROOT" status --porcelain --untracked-files=no)" ]; then
  echo "工作樹有未提交的改動；打包要從乾淨的輸入開始。" >&2
  exit 2
fi

if [ -e "$OUT" ] || [ -L "$OUT" ]; then
  echo "版本目錄已存在，拒絕覆寫：$OUT" >&2
  exit 2
fi
mkdir -p "$REPO_ROOT/dist-all"
mkdir "$OUT"

run_stage() {
  timeout "${KOL_GO_TIMEOUT:-900}" docker run --rm \
    --network none --memory 2g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -u "$(id -u):$(id -g)" -e HOME=/tmp \
    -v "$REPO_ROOT:/src:ro" -v "$OUT:/out" -w /src \
    "${KOL_PACKAGE_IMAGE:-kol-package:20260907}" \
    bash /src/tools/build_dist.sh --stage-in-container "$VERSION" "$1"
}
run_stage prepare

# 1. 建置三個平台。
#    linux/amd64   在容器裡原生建（Ebitengine 在 Linux 要 CGO）。
#    windows/amd64 交叉編得出來——**Ebitengine 在 Windows 不用 CGO**，走的是 syscall。
#    darwin        要 Cocoa／Metal，所以要 CGO ＋ osxcross 的 clang；那一段在
#                  `tools/build_macos.sh`，產出的是 arm64 ＋ x86_64 的 universal。
#
#    ⚠ **Linux 容器裡只驗得了結構**：`verify_macos.sh` 的四道與 Windows 的檔頭
#    都不是真機驗收。三個平台裡只有 linux/amd64 這一份真的跑得起來。
"$REPO_ROOT/tools/go_docker.sh" build -trimpath -ldflags "-X main.buildVersion=$VERSION" \
  -o /src/dist-all/"$VERSION"/full/kol ./cmd/kol
KOL_GO_ENV="GOOS=windows GOARCH=amd64 CGO_ENABLED=0" \
  "$REPO_ROOT/tools/go_docker.sh" build -trimpath -ldflags "-X main.buildVersion=$VERSION" \
  -o /src/dist-all/"$VERSION"/full/kol.exe ./cmd/kol
KOL_BUILD_VERSION="$VERSION" "$REPO_ROOT/tools/build_macos.sh" "dist-all/$VERSION/full/kol-macos"
"$REPO_ROOT/tools/verify_macos.sh" "dist-all/$VERSION/full/kol-macos"
run_stage files
# 6. promo：推廣片與它的來源、metadata 與影音驗收。
"$REPO_ROOT/tools/make_promo.sh" "$OUT/promo"

run_stage manifest

echo "$OUT"
