#!/usr/bin/env bash
# 僅在 Docker 內執行；/src 唯讀，/out 是已建置 kol 所在的可寫驗證目錄。
set -eu
test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
if [ "${KOL_ENTER_BATTLE:-0}" = 1 ]; then
 world_steps="${KOL_WORLD_STEPS:-15}"
 [[ "$world_steps" =~ ^[1-9][0-9]?$|^100$ ]] || { echo 'KOL_WORLD_STEPS 必須介於 1 至 100' >&2; exit 2; }
fi
export HOME=/tmp DISPLAY=:94
Xvfb :94 -screen 0 1600x1000x24 -nolisten tcp >/out/inn-xvfb.log 2>&1 &
trap 'kill $(jobs -pr) 2>/dev/null || true' EXIT
for n in $(seq 1 50); do [ -e /tmp/.X11-unix/X94 ] && break; sleep 0.1; done
cd /src
extra=()
if [ "${KOL_ENTER_BATTLE:-0}" = 1 ]; then extra=(-seed 1); fi
/out/kol -log-input "${extra[@]}" >/out/inn-gui.log 2>&1 &
game_pid=$!
win=""
for n in $(seq 1 100); do
 win="$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true)"
 [ -n "$win" ] && break
 sleep 0.1
done
test -n "$win"
xdotool windowfocus "$win"
sleep 5
key() {
	before=$(grep -c '^input scene=' /out/inn-gui.log || true)
 xdotool keydown --window "$win" "$1"
 sleep 0.2
 xdotool keyup --window "$win" "$1"
	for n in $(seq 1 100); do
	 after=$(grep -c '^input scene=' /out/inn-gui.log || true)
	 [ "$after" -gt "$before" ] && break
	 sleep 0.2
	done
	test "$after" -gt "$before"
	sleep "${KOL_KEY_DELAY:-2}"
}
shot() { import -window "$win" "/out/$1.png"; }
key Return
key Return
key 6
for n in $(seq 1 "${KOL_SELECT_DOWN:-0}"); do key Down; done
key Return
shot inn-entry
key space
shot inn-greeting
key less
key Return
shot inn-recruit-question
key Return
shot inn-recruit-list
if [ "${KOL_RECRUIT_MOUSE:-0}" = 1 ]; then
 key greater
 # 繁中畫布第二列；依視窗大小換算，避免把實體像素當成畫布座標。
 width=$(xdotool getwindowgeometry --shell "$win" | awk -F= '/^WIDTH=/{print $2}')
 height=$(xdotool getwindowgeometry --shell "$win" | awk -F= '/^HEIGHT=/{print $2}')
 xdotool mousemove --window "$win" "$((220*width/640))" "$((196*height/400))"
 xdotool click --window "$win" 1
 sleep 3
 grep -F 'token=[D < RET]' /out/inn-gui.log >/dev/null
else
 key Return
fi
shot inn-two-members
key Return
shot inn-exit
if [ "${KOL_ENTER_BATTLE:-0}" = 1 ]; then
 key Right
 key Right
 for n in $(seq 1 9); do key Up; done
 shot town-edge
 key Return
 shot world-entry
 for n in $(seq 1 "$world_steps"); do key Right; done
 shot world-after-walk
 key space
 shot battle-entry
 if [ "${KOL_COMMAND_PANEL:-0}" = 1 ]; then
  # 正常選移動、方向與走路；再讓第二名休息，推進結算。
  key greater
  key greater
  key greater
  key Return
  shot command-target
  key Return
  shot command-move-form
  key greater
  key Return
  shot command-next-actor
  key less
  key Return
  shot command-round-ready
  for n in $(seq 1 12); do key Return; done
  shot command-after-round
  key greater
  key Return
  shot command-inspect-ally
  key greater
  key Return
  shot command-inspect-return
 fi
 if [ "${KOL_EQUIPMENT_PANEL:-0}" = 1 ]; then
  key space
  shot battle-portrait
  key space
  shot battle-equipment
  key space
  shot battle-portrait-return
 fi
 if [ "${KOL_BODY_INSPECTION:-0}" = 1 ]; then
  # 固定種子與上述隊伍下，第二格為徒手指令；進目標格後以返回圖示取消。
  # 是否真的進入目標選擇仍須檢視三張畫格，不以命令成功冒充 GUI 驗收。
  key greater
  key Return
  shot battle-target-body
  key greater
  key Return
  shot battle-body-return
 fi
 if [ "${KOL_AMMUNITION_SHOT:-0}" = 1 ]; then
  # Olear 的第二格是射擊；用正常目標游標選右側一格，不注入戰場狀態。
  key greater
  key Return
  key Right
  shot ammunition-target
  key Return
  shot ammunition-next-actor
  key less
  key Return
  shot ammunition-round-ready
  for n in $(seq 1 12); do key Return; done
  shot ammunition-after-round
  key space
  shot ammunition-after-shot
  # 下一輪先到 Theodoric；讓他休息後才檢查 Olear 的當前彈藥。
  key less
  key Return
  key space
  shot ammunition-current-actor
 fi
fi
xdotool keydown --window "$win" ctrl+q
wait "$game_pid"
