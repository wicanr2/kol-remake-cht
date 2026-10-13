#!/bin/bash
# 固定流程只重播一次；Xvfb明確等待就緒，不依賴xvfb-run訊號握手。
set -euo pipefail
test -f /.dockerenv
export DISPLAY=:99
Xvfb :99 -screen 0 1920x1200x24 -nolisten tcp >/out/smoke/xvfb-sample.log 2>&1 & xvfb=$!
trap 'kill "$xvfb" 2>/dev/null || true' EXIT
for i in $(seq 1 50); do xdpyinfo >/dev/null 2>&1 && break; sleep .1; done
xdpyinfo >/dev/null
version=$(cat /out/full/VERSION)
python3 /src/tools/closeout_release.py sample --version "$version"
