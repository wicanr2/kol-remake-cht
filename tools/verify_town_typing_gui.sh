#!/usr/bin/env bash
# 城鎮打字問話的真實鍵盤驗收（#61）：Xvfb 裡開互動遊玩，xdotool 送按鍵，
# 每一鍵都等 `-log-input` 多出一行才送下一鍵。僅限 Docker（kol-package 映像）：
#   /src 唯讀掛工作樹、/w 可寫；/w/gui/start.json 是起始存檔；LANG_MODE=zh-Hant|en。
# 先用 tools/go_docker.sh build -o workplace/tmp/kol ./cmd/kol 編好執行檔。
set -e
STALE="$(find /src/cmd /src/internal -name '*.go' -newer /src/workplace/tmp/kol -print -quit)"
if [ -n "$STALE" ]; then echo "workplace/tmp/kol 比原始碼舊（$STALE），先重編" >&2; exit 2; fi
export HOME=/tmp DISPLAY=:97
Xvfb :97 -screen 0 1600x1000x24 -nolisten tcp >/tmp/x.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
for n in $(seq 1 50); do [ -e /tmp/.X11-unix/X97 ] && break; sleep .1; done
cp /w/gui/start.json /tmp/save.json
cd /src
log=/w/gui/kol3.log
/src/workplace/tmp/kol -lang "$LANG_MODE" -seed 1 -log-input -load /tmp/save.json -save /tmp/save.json > $log 2>&1 &
pid=$!
win=""; for n in $(seq 1 100); do win=$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true); [ -n "$win" ] && break; sleep .1; done
test -n "$win"; xdotool windowfocus "$win"; sleep 3
k(){ before=$(grep -c '^input scene=' $log || true); xdotool keydown --window "$win" "$1"; sleep .15; xdotool keyup --window "$win" "$1"
  for n in $(seq 1 60); do after=$(grep -c '^input scene=' $log || true); [ "$after" -gt "$before" ] && break; sleep .1; done; test "$after" -gt "$before"; sleep .3; }
s(){ import -window "$win" /w/gui/$LANG_MODE-$1.png; }
k Escape; k 6; for n in $(seq 1 14); do k Down; done; k Return; k space
for x in greater greater greater; do k $x; done
k Return; s prompt
for x in e l v x; do k $x; done; s typed
k BackSpace; s backspaced
k e; k Return; s answer
k space; s after-answer
for x in greater greater greater; do k $x; done
k Return; s prompt2
k Return; s empty-submit
kill $pid 2>/dev/null || true
