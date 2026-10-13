#!/bin/bash
# Wine9 image；音訊連線只到同輪捕捉容器的Unix socket。
set -euo pipefail
test -f /.dockerenv
export DISPLAY=:99 WINEPREFIX=/tmp/kol-wine WINEDEBUG=-all WINEDLLOVERRIDES='mscoree,mshtml=' PULSE_SERVER=unix:/out/smoke/wine-pulse/native PULSE_SINK=kol_wine
mkdir -p /tmp/kol-home "$WINEPREFIX"
export HOME=/tmp/kol-home
Xvfb :99 -screen 0 1920x1200x24 -nolisten tcp >/out/smoke/xvfb-wine.log 2>&1 & xvfb=$!
test ! -f /sys/fs/cgroup/memory.events || cat /sys/fs/cgroup/memory.events >/out/smoke/wine-memory-before.txt
trap 'test ! -f /sys/fs/cgroup/memory.events || cat /sys/fs/cgroup/memory.events >/out/smoke/wine-memory-after.txt; test ! -f /sys/fs/cgroup/memory.peak || cat /sys/fs/cgroup/memory.peak >/out/smoke/wine-memory-peak.txt; touch /out/smoke/wine-audio-done; wineserver -k 2>/dev/null || true; kill "$xvfb" 2>/dev/null || true' EXIT
for i in $(seq 1 100); do test -f /out/smoke/wine-audio-ready && break; sleep .1; done
test -S /out/smoke/wine-pulse/native
sleep .3
/usr/lib/wine/wine64 reg add 'HKCU\Software\Wine\Drivers' /v Audio /t REG_SZ /d pulse /f >/out/smoke/wine-driver.log 2>&1
python3 /src/tools/verify_release_runtime.py windows
