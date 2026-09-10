#!/usr/bin/env bash
# Docker 內；受控存檔只縮短到場與升階資格準備，不冒稱正常取得資格。
set -euo pipefail
export HOME=/tmp DISPLAY=:99
test "$(stat -c %u:%g /out /out/kol | sort -u)" = "$(id -u):$(id -g)"
test -x /out/kol
Xvfb :99 -screen 0 1600x1000x24 -nolisten tcp >/out/xvfb.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
cd /src
for n in $(seq 1 50);do [ -e /tmp/.X11-unix/X99 ] && break;sleep .1;done
/out/kol -seed 1 -log-input -load /out/input-v1.json -save /out/result.json >/out/gui.log 2>&1 & pid=$!
for n in $(seq 1 100);do win=$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true);[ -n "$win" ] && break;sleep .1;done
test -n "$win";xdotool windowfocus "$win";sleep 3
key() {
 before=$(grep -c '^input scene=' /out/gui.log || true)
 xdotool keydown --window "$win" "$1";sleep .15;xdotool keyup --window "$win" "$1"
 for n in $(seq 1 60);do after=$(grep -c '^input scene=' /out/gui.log || true);[ "$after" -gt "$before" ] && break;sleep .1;done
 test "$after" -gt "$before";sleep .35
}
shot() { import -window "$win" "/out/$1.png"; }
shot import-notice
key space
shot loaded-world
key Right
shot arena-door
key Return
shot arena-choose
key Down;key Return
shot arena-confirm
key Return
shot arena-tunnel
key space
shot arena-bet
key greater;key Return
shot arena-intro
key space
shot arena-field
xdotool keydown --window "$win" ctrl+q;wait "$pid"
if grep -q '繪製失敗' /out/gui.log; then
 echo '競技場繪圖失敗，驗證不通過' >&2
 exit 1
fi
