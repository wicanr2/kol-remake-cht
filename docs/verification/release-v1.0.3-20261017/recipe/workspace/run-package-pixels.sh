set -euo pipefail
test -f /.dockerenv
export DISPLAY=:99 GOMEMLIMIT=3GiB GOMAXPROCS=3
Xvfb :99 -screen 0 1920x1200x24 -nolisten tcp >/out/smoke/pixels-xvfb.log 2>&1 & xp=$!
trap 'kill "$xp" 2>/dev/null || true; wait "$xp" 2>/dev/null || true' EXIT INT TERM
for i in $(seq 1 50); do xdpyinfo >/dev/null 2>&1 && break; sleep .1; done
xdpyinfo >/dev/null
python3 /src/tools/verify_release_runtime.py pixels
