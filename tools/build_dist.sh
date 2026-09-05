#!/usr/bin/env bash
# 從乾淨的工作樹建出 `dist-all/<版本>/` 的三個包（專案契約：`full`／`release`／`promo`）。
#
#   tools/build_dist.sh 0.1.0
#
# 建置在鎖定的 Go 容器裡跑，輸出以目前 UID/GID 寫入。三個包的界線：
#
#   full/     含原作衍生素材的完整版，**只留本機**，不得上傳、不得散布。
#   release/  通過權利審查的那一份：只有 remake 自己的程式碼、文件與抽取器，
#             **一個原版衍生位元組都不帶**；玩家用自己的原版磁片跑抽取器產生素材。
#   promo/    影片、metadata、來源與影音驗收。目前只有骨架。
#
# 每次都從乾淨輸入重建：先刪掉那個版本的目錄再建，免得留下上一次的殘骸。
set -euo pipefail

VERSION="${1:?用法：build_dist.sh <版本>}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$REPO_ROOT/dist-all/$VERSION"

if [ -n "$(git -C "$REPO_ROOT" status --porcelain --untracked-files=no)" ]; then
  echo "工作樹有未提交的改動；打包要從乾淨的輸入開始。" >&2
  exit 2
fi

rm -rf "$OUT"
mkdir -p "$OUT"/{full,release,promo}

# 1. 建置。linux/amd64 在容器裡原生建（Ebitengine 在 Linux 要 CGO）；
#    windows/amd64 交叉編得出來——**Ebitengine 在 Windows 不用 CGO**，走的是
#    syscall。macOS 要 Cocoa，得另外的工具鏈（osxcross），還沒做。
"$REPO_ROOT/tools/go_docker.sh" build -trimpath -o /src/dist-all/"$VERSION"/full/kol ./cmd/kol
KOL_GO_ENV="GOOS=windows GOARCH=amd64 CGO_ENABLED=0" \
  "$REPO_ROOT/tools/go_docker.sh" build -trimpath \
  -o /src/dist-all/"$VERSION"/full/kol.exe ./cmd/kol
cp "$OUT/full/kol" "$OUT/full/kol.exe" "$OUT/release/"

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

# 6. promo 的骨架。
printf '%s\n' \
  "推廣素材放這裡：影片、截圖、metadata、來源與影音驗收。" \
  "目前是空的——推廣片還沒做（見 WORKLIST 的 M6）。" \
  > "$OUT/promo/README.txt"

# 7. deny-list：release 裡不得有任何原版衍生位元組。
if find "$OUT/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
      -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) -print -quit | grep -q .; then
  echo "release 裡出現了不該散布的檔案：" >&2
  find "$OUT/release" \( -name '*.png' -o -name '*.ogg' -o -name '*.EXE' \
      -o -name '*.zip' -o -name '*.pdf' -o -name '*.jpg' \) >&2
  exit 3
fi

# 8. manifest：每個包逐檔 SHA-256。
"$REPO_ROOT/tools/run_python_tool.sh" tools/write_dist_manifest.py \
  --dist "dist-all/$VERSION" --version "$VERSION"

echo "$OUT"
