set -euo pipefail
wine_runtime=/usr/lib/wine/wine64
binary='Z:\work\workplace\platform-smoke-20261001\build\kol-windows.exe'
root=/work/workplace/platform-smoke-20261001/windows
$wine_runtime "$binary" -version >"$root/version.txt" 2>"$root/version.log"
timeout 80s $wine_runtime "$binary" -scene cover -lang zh-Hant -shot 'Z:\work\workplace\platform-smoke-20261001\windows\cover.png' >"$root/cover.log" 2>&1
test -s "$root/cover.png"
printf 'Windows中文封面擷取通過。\n'
