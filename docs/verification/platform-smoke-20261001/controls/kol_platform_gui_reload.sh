set -euo pipefail
platform=$1
root=/work/workplace/platform-smoke-20261001/$platform
out="$root/gui-reload"
mkdir -p "$out" "$root/reload"
if [ "$platform" = windows ]; then
  save="$WINEPREFIX/drive_c/users/ubuntu/AppData/Roaming/kol-remake-cht/save.dat"
  load='C:\users\ubuntu\AppData\Roaming\kol-remake-cht\save.dat'
  binary='Z:\work\workplace\platform-smoke-20261001\build\kol-windows.exe'
  runtime=/usr/lib/wine/wine64
  shotdir='Z:\work\workplace\platform-smoke-20261001\windows\reload'
  trace='Z:\work\workplace\platform-smoke-20261001\windows\reload\trace.json'
else
  save="$XDG_CONFIG_HOME/kol-remake-cht/save.dat"
  load="$save"
  binary=/work/workplace/platform-smoke-20261001/build/kol-linux
  runtime=
  shotdir="$root/reload"
  trace="$root/reload/trace.json"
fi
test -s "$save"
sha256sum "$save" >"$root/save-before-reload.sha256"
Xvfb "$DISPLAY" -screen 0 1280x800x24 -nolisten tcp >"$out/xvfb.log" 2>&1 & xvfb_pid=$!
game_pid=
cleanup() {
  if [ -n "$game_pid" ]; then kill "$game_pid" 2>/dev/null || true; wait "$game_pid" 2>/dev/null || true; fi
  if [ "$platform" = windows ]; then /usr/lib/wine/wineserver -k || true; fi
  kill "$xvfb_pid" 2>/dev/null || true
  wait "$xvfb_pid" 2>/dev/null || true
}
trap cleanup EXIT
sleep 1
$runtime "$binary" -seed 1 -log-input -load "$load" >"$out/engine.log" 2>&1 & game_pid=$!
for n in $(seq 1 1000); do
  if [ -e "$out/controller-done" ]; then break; fi
  kill -0 "$game_pid"
  sleep .1
done
test -e "$out/controller-done"
wait "$game_pid"; game_pid=
sha256sum -c "$root/save-before-reload.sha256"
timeout 60s $runtime "$binary" -scene play -seed 1 -lang zh-Hant -load "$load" -play-script '1 > RET D D D D D D D D D D D RET' -shot-dir "$shotdir" -shot-steps '0,1,3,14,15' -trace "$trace" >"$root/reload/engine.log" 2>&1
sha256sum -c "$root/save-before-reload.sha256"
printf '%s 互動與逐步重載均完成，保存未改写。\n' "$platform"
