#!/usr/bin/env bash
# Docker 內，/src 唯讀、/out 為本輪輸出。
set -euo pipefail
export HOME=/tmp DISPLAY=:97
test "$(stat -c %u:%g /out /out/kol | sort -u)" = "$(id -u):$(id -g)"
Xvfb :97 -screen 0 1600x1000x24 -nolisten tcp >/out/xvfb-goodbye.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
cd /src
for n in $(seq 1 50);do [ -e /tmp/.X11-unix/X97 ] && break;sleep .1;done
/out/kol -seed 1 -log-input -save /out/goodbye-save.json >/out/goodbye-gui.log 2>&1 & pid=$!
for n in $(seq 1 100);do win=$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true);[ -n "$win" ] && break;sleep .1;done
test -n "$win";xdotool windowfocus "$win";sleep 3
key() {
 before=$(grep -c '^input scene=' /out/goodbye-gui.log || true)
 xdotool keydown --window "$win" "$1";sleep .15;xdotool keyup --window "$win" "$1"
 for n in $(seq 1 60);do after=$(grep -c '^input scene=' /out/goodbye-gui.log || true);[ "$after" -gt "$before" ] && break;sleep .1;done
 test "$after" -gt "$before";sleep .3
}
shot() { import -window "$win" "/out/$1.png"; }
key Return;key Return;key 6;key Return;key space
shot single-inn
key Return
shot single-outside
# 回到門格重新進店，讀入店敘述後招募正常候選人。
key Left;key Right;key Return;key space
key less;key Return;key Return;key Return
shot two-members
key Return
shot goodbye-waiting
key Right
shot goodbye-ignored-move
key space
shot goodbye-outside
xdotool keydown --window "$win" ctrl+q;wait "$pid"
compare -metric AE /out/goodbye-waiting.png /out/goodbye-ignored-move.png null: 2>/out/goodbye-waiting-diff.txt
echo '正常離店 GUI 腳本完成；單人／多人畫格須目視確認'
