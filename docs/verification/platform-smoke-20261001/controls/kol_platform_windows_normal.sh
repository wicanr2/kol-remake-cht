set -euo pipefail
root=/work/workplace/platform-smoke-20261001/windows
wine_runtime=/usr/lib/wine/wine64
binary='Z:\work\workplace\platform-smoke-20261001\build\kol-windows.exe'
save="$WINEPREFIX/drive_c/users/ubuntu/AppData/Roaming/kol-remake-cht/save.dat"
test ! -e "$save"
mkdir -p "$root/normal" "$root/reload"
inputs=$(cat /work/workplace/platform-smoke-20261001/purchase-script.txt)
timeout 110s "$wine_runtime" "$binary" -scene flow -seed 1 -lang zh-Hant -flow-script "$inputs" -shot-dir 'Z:\work\workplace\platform-smoke-20261001\windows\normal' -shot-steps '0,103,104,111,116,117,120,126,137,138,198' -trace 'Z:\work\workplace\platform-smoke-20261001\windows\normal\trace.json' >"$root/normal/engine.log" 2>&1
test -s "$save"
sha256sum "$save" >"$root/save-before-reload.sha256"
cp "$save" "$root/default-save.dat"
timeout 80s "$wine_runtime" "$binary" -scene play -seed 1 -lang zh-Hant -load 'C:\users\ubuntu\AppData\Roaming\kol-remake-cht\save.dat' -play-script '1 > RET D D D D D D D D D D D RET' -shot-dir 'Z:\work\workplace\platform-smoke-20261001\windows\reload' -shot-steps '0,1,3,14,15' -trace 'Z:\work\workplace\platform-smoke-20261001\windows\reload\trace.json' >"$root/reload/engine.log" 2>&1
sha256sum -c "$root/save-before-reload.sha256"
printf 'Windows 正常購買、預設存檔與重載完成。\n'
