#!/usr/bin/env bash
set -eu

export DISPLAY=:99
export TERM=xterm
OUT=/out
EXPECTED_KOLTOWN_SHA256=c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221

cleanup() {
  kill "${DOSBOX_PID:-}" "${XVFB_PID:-}" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

actual_sha256=$(sha256sum /game/Knights-of-legend/KOLTOWN.EXE | cut -d' ' -f1)
if [ "$actual_sha256" != "$EXPECTED_KOLTOWN_SHA256" ]; then
  echo "KOLTOWN.EXE SHA-256 不符：$actual_sha256" >&2
  exit 1
fi

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
mount c /game/Knights-of-legend
c:
KOLTOWN.EXE A
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

if ! python3 /repo/tools/oracle/koltown_runtime_oracle.py \
  "$OUT/KOLTOWN-runtime-load.bin" "$OUT/KOLTOWN-runtime-receipt.json"; then
  import -window root "$OUT/KOLTOWN-oracle-failure.png" 2>/dev/null || true
  cp /tmp/kol-dosbox.log "$OUT/KOLTOWN-oracle-failure.log" || true
  tail -60 /tmp/kol-dosbox.log >&2
  exit 1
fi

tail -20 /tmp/kol-dosbox.log
