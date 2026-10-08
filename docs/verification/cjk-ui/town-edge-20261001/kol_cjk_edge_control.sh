set -euo pipefail
export LIBGL_ALWAYS_SOFTWARE=1 GALLIUM_DRIVER=llvmpipe
Xvfb :94 -screen 0 1280x800x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
xvfb_pid=$!
trap 'kill "$xvfb_pid" 2>/dev/null || true' EXIT
for attempt in {1..50}; do
    [ -e /tmp/.X11-unix/X94 ] && break
    sleep 0.1
done
export DISPLAY=:94
python3 /control/kol_cjk_edge_run.py "$@"
