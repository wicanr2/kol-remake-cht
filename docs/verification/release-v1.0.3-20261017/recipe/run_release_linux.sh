#!/bin/bash
# 既有音訊image；/src唯讀、/out為交付目錄。
set -euo pipefail
test -f /.dockerenv
mkdir -p /tmp/kol-pulse
export XDG_RUNTIME_DIR=/tmp/kol-pulse PULSE_SERVER=unix:/tmp/kol-pulse/native PULSE_SINK=kol_music DISPLAY=:99 PYTHONDONTWRITEBYTECODE=1
Xvfb :99 -screen 0 1920x1200x24 -nolisten tcp >/out/smoke/xvfb-linux.log 2>&1 & xvfb=$!
pulseaudio -n --daemonize=no --use-pid-file=no --exit-idle-time=-1 --load="module-native-protocol-unix socket=/tmp/kol-pulse/native auth-anonymous=1" --load="module-null-sink sink_name=kol_music rate=44100 channels=2" >/out/smoke/pulse-linux.log 2>&1 & pulse=$!
trap 'kill "$xvfb" "$pulse" 2>/dev/null || true' EXIT
for i in $(seq 1 50); do pactl info >/dev/null 2>&1 && break; sleep .1; done
sleep .3
python3 /src/tools/verify_release_runtime.py linux
