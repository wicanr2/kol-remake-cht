set -eu
export DISPLAY=:99 WINEPREFIX=/tmp/kol-hd-scenes WINEDEBUG=-all WINEDLLOVERRIDES='mscoree,mshtml='
export XDG_CACHE_HOME=/tmp/kol-wine-cache
export GOMEMLIMIT=3GiB
mkdir -p "$WINEPREFIX" "$XDG_CACHE_HOME"
Xvfb :99 -screen 0 2048x1536x24 -nolisten tcp >/out/xvfb-windows.log 2>&1 & xp=$!
trap 'wineserver -k 2>/dev/null || true; kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true; cat /sys/fs/cgroup/memory.peak >/out/memory-peak.txt' EXIT INT TERM
for i in $(seq 1 50); do xdotool getdisplaygeometry >/dev/null 2>&1 && break; sleep .1; done
xdotool getdisplaygeometry >/dev/null
cd /src
KOL_RELEASE_MAIN=1 timeout 900 /usr/lib/wine/wine64 'Z:\capture\bin\hd-scenes.exe' >/out/run-windows.log 2>&1
tail -n 18 /out/run-windows.log
