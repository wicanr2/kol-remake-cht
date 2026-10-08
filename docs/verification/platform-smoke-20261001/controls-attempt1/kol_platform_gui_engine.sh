set -euo pipefail
platform=$1
root=/work/workplace/platform-smoke-20261001/$platform
out="$root/gui-normal"
mkdir -p "$out"
if [ "$platform" = windows ]; then
  save="$WINEPREFIX/drive_c/users/ubuntu/AppData/Roaming/kol-remake-cht/save.dat"
  binary='Z:\work\workplace\platform-smoke-20261001\build\kol-windows.exe'
  runtime=/usr/lib/wine/wine64
else
  save="$XDG_CONFIG_HOME/kol-remake-cht/save.dat"
  binary=/work/workplace/platform-smoke-20261001/build/kol-linux
  runtime=
fi
test ! -e "$save"
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
$runtime "$binary" -seed 1 -log-input >"$out/engine.log" 2>&1 & game_pid=$!
for n in $(seq 1 1800); do
  if [ -e "$out/controller-done" ]; then break; fi
  kill -0 "$game_pid"
  sleep .1
done
test -e "$out/controller-done"
wait "$game_pid"; game_pid=
test -s "$save"
cp "$save" "$root/default-save.dat"
sha256sum "$save" >"$root/gui-save.sha256"
printf '%s 預設存檔互動流程完成。\n' "$platform"
