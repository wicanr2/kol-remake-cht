set -euo pipefail
root=/work/workplace/platform-smoke-20261001/linux
binary=/work/workplace/platform-smoke-20261001/build/kol-linux
save="$XDG_CONFIG_HOME/kol-remake-cht/save.dat"
test ! -e "$save"
mkdir -p "$root/normal" "$root/reload"
"$binary" -version >"$root/version.txt"
timeout 50s "$binary" -scene cover -lang zh-Hant -shot "$root/cover.png" >"$root/cover.log" 2>&1
inputs=$(cat /work/workplace/platform-smoke-20261001/purchase-script.txt)
timeout 110s "$binary" -scene flow -seed 1 -lang zh-Hant -flow-script "$inputs" -shot-dir "$root/normal" -shot-steps '0,103,104,111,116,117,120,126,137,138,198' -trace "$root/normal/trace.json" >"$root/normal/engine.log" 2>&1
test -s "$save"
sha256sum "$save" >"$root/save-before-reload.sha256"
cp "$save" "$root/default-save.dat"
timeout 80s "$binary" -scene play -seed 1 -lang zh-Hant -load "$save" -play-script '1 > RET D D D D D D D D D D D RET' -shot-dir "$root/reload" -shot-steps '0,1,3,14,15' -trace "$root/reload/trace.json" >"$root/reload/engine.log" 2>&1
sha256sum -c "$root/save-before-reload.sha256"
printf 'Linux 正常購買、預設存檔與重載完成。\n'
