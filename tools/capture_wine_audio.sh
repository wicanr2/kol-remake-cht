#!/bin/bash
# 由既有音訊image提供隔離PulseAudio；Wine容器只共享本輪socket與輸出目錄。
set -euo pipefail
test -f /.dockerenv
mkdir /out/smoke/wine-pulse
export XDG_RUNTIME_DIR=/out/smoke/wine-pulse PULSE_SERVER=unix:/out/smoke/wine-pulse/native PULSE_SINK=kol_wine
pulseaudio -n --daemonize=no --use-pid-file=no --exit-idle-time=-1 \
 --load="module-native-protocol-unix socket=/out/smoke/wine-pulse/native auth-anonymous=1" \
 --load="module-null-sink sink_name=kol_wine rate=44100 channels=2" >/out/smoke/wine-pulse.log 2>&1 & pulse=$!
recorder=
trap 'test -z "$recorder" || kill -INT "$recorder" 2>/dev/null || true; kill "$pulse" 2>/dev/null || true' EXIT
for i in $(seq 1 50); do pactl info >/dev/null 2>&1 && break; sleep .1; done
pactl info >/out/smoke/wine-pulse-info.txt
ffmpeg -y -nostdin -v error -f pulse -sample_rate 44100 -channels 2 -i kol_wine.monitor -c:a pcm_s16le /out/smoke/windows-audio.wav >/out/smoke/windows-audio-capture.log 2>&1 & recorder=$!
touch /out/smoke/wine-audio-ready
for i in $(seq 1 9000); do
 if test -f /out/smoke/wine-audio-done; then sleep 2; kill -INT "$recorder"; wait "$recorder" || true; recorder=; exit 0; fi
 sleep .2
done
exit 1
