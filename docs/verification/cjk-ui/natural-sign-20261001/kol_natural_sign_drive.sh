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
export KOL_BIN=/src/workplace/cjk-edge-20261001/kol-edge-fixed
export KOL_DRIVE_DIR=workplace/cjk-sign-natural-20261001/driver
export KOL_SCENE=flow KOL_SCRIPT_FLAG=-flow-script
export KOL_QUEST_SITES= KOL_MONSTERS= KOL_AMBUSH_ROLLS= KOL_STOP_AFTER_BATTLE= KOL_DRIVE_QUEST=
python3 - <<'PY'
import pathlib,json,hashlib
root=pathlib.Path('workplace/cjk-sign-natural-20261001')
tokens=(root/'party-prefix.txt').read_text().split()+['RET','SPACE','R','R']+['U']*9+['RET']
(root/'world-prefix.txt').write_text(' '.join(tokens)+'\n')
record={'seed':1,'seed_method':'-seed 1','initial_state':'正式標題、固定原版名冊、正常旅店招募六人，再按鍵出城',
        'input_prefix':' '.join(tokens),'world_target':[37,38],'town_target':[40,19],
        'load':None,'quest_sites':None,'monsters':None,'ambush_rolls':None,
        'binary_sha256':hashlib.sha256(pathlib.Path('workplace/cjk-edge-20261001/kol-edge-fixed').read_bytes()).hexdigest(),
        'driver_sha256':hashlib.sha256(pathlib.Path('tools/drive_battle_script.py').read_bytes()).hexdigest(),
        'comparison_scope':'僅重製正常玩家路徑，沒有原版亂數對拍或完整原版 parity 聲明'}
(root/'travel-run.json').write_text(json.dumps(record,ensure_ascii=False,indent=2)+'\n')
PY
python3 -u tools/drive_battle_script.py 1 - workplace/cjk-sign-natural-20261001/world-prefix.txt workplace/cjk-sign-natural-20261001/world-route.txt 1000 37,38 | tee workplace/cjk-sign-natural-20261001/driver.log
