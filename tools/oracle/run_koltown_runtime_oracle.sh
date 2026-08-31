#!/usr/bin/env bash
set -eu

export DISPLAY=:99
export TERM=xterm
OUT=/out

cleanup() {
  kill "${DOSBOX_PID:-}" "${XVFB_PID:-}" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

Xvfb :99 -screen 0 800x600x24 -ac >/tmp/kol-xvfb.log 2>&1 &
XVFB_PID=$!

cat >/tmp/kol-oracle.conf <<'CONF'
[sdl]
fullscreen=false
output=surface
waitonerror=false
[dosbox]
machine=svga_et4000
memsize=16
[cpu]
core=normal
cputype=386
cycles=fixed 20000
[mixer]
nosound=true
[autoexec]
mount c /game
c:
cd Knights-of-legend
KOLTOWN.EXE 0
CONF

script -q -f -c "dosbox-x -conf /tmp/kol-oracle.conf -break-start" \
  /tmp/kol-dosbox.log >/dev/null 2>&1 &
DOSBOX_PID=$!

connected=0
for _ in $(seq 1 20); do
  if python3 -c 'import socket; socket.create_connection(("127.0.0.1",9876),1).close()' \
      2>/dev/null; then
    connected=1
    break
  fi
  sleep 1
done

if [ "$connected" -ne 1 ]; then
  tail -40 /tmp/kol-dosbox.log >&2
  exit 1
fi

python3 /repo/tools/oracle/koltown_runtime_oracle.py \
  "$OUT/KOLTOWN-runtime-load.bin" "$OUT/KOLTOWN-runtime-receipt.json"

tail -20 /tmp/kol-dosbox.log
