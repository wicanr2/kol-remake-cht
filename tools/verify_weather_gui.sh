#!/usr/bin/env bash
# Docker 內使用；正常單人選角、離店與出城，不注入世界座標。
set -euo pipefail
export HOME=/tmp DISPLAY=:98
test "$(stat -c %u:%g /out /out/kol | sort -u)" = "$(id -u):$(id -g)"
Xvfb :98 -screen 0 1600x1000x24 -nolisten tcp >/out/weather-xvfb.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
cd /src
for n in $(seq 1 50);do [ -e /tmp/.X11-unix/X98 ] && break;sleep .1;done
/out/kol -seed 1 -log-input -save /out/weather-save.json >/out/weather-gui.log 2>&1 & pid=$!
for n in $(seq 1 100);do win=$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true);[ -n "$win" ] && break;sleep .1;done
test -n "$win";xdotool windowfocus "$win";sleep 3
key() {
 before=$(grep -c '^input scene=' /out/weather-gui.log || true)
 xdotool keydown --window "$win" "$1";sleep .15;xdotool keyup --window "$win" "$1"
 for n in $(seq 1 60);do after=$(grep -c '^input scene=' /out/weather-gui.log || true);[ "$after" -gt "$before" ] && break;sleep .1;done
 test "$after" -gt "$before";sleep .25
}
key Return;key Return;key 6;key Return;key space;key Return
key Right;key Right
for n in $(seq 1 9);do key Up;done
key Return
key Return
for n in $(seq 1 4);do import -window "$win" "/out/weather-normal-$n.png";sleep .4;done
xdotool keydown --window "$win" ctrl+q;wait "$pid"
