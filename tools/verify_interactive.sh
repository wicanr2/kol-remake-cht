#!/usr/bin/env bash
# 驗證互動遊玩真的吃鍵盤與滑鼠。
#
#   tools/verify_interactive.sh [工作目錄] [啟動命令]
#
# 啟動命令預設是 `./kol`。要驗打包後的東西就指定它，例如
#   tools/verify_interactive.sh dist-all/0.2.1/packages \
#     "./kol-0.2.1-full-linux-x86_64.AppImage --appimage-extract-and-run"
# ——那才證明**發行包本身**能玩，而不只是儲存庫裡的執行檔能玩。
#
# 判準是**畫面雜湊必須改變**：送一個鍵進去、截一張圖，和上一張比。按鍵沒作用時
# 兩張會一模一樣，測試就紅——這正是 0.2.0 回報的症狀，所以這支腳本擋得住它復發。
#
# 在鎖定容器內以 Xvfb 跑，xdotool 送事件，ImageMagick 截圖。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKDIR="${1:-.}"
KOL_CMD="${2:-./kol}"
IMAGE="${KOL_GO_IMAGE:-eob-remake-go:1.26.7-ebiten2.9.9-audio}"

exec timeout "${KOL_GO_TIMEOUT:-600}" docker run --rm \
  --network none \
  --memory 3g --cpus "${KOL_GO_CPUS:-2}" --pids-limit 512 \
  --log-opt max-size=10m --log-opt max-file=3 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp -e "KOL_WORKDIR=$WORKDIR" \
  -e "KOL_CMD=$KOL_CMD" \
  -v "$REPO_ROOT:/src" -w /src \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=512m" \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    cd "/src/$KOL_WORKDIR"
    export DISPLAY=:94
    Xvfb :94 -screen 0 1600x1000x24 -nolisten tcp >/tmp/xvfb.log 2>&1 &
    for _ in $(seq 1 50); do [ -e /tmp/.X11-unix/X94 ] && break; sleep 0.1; done

    $KOL_CMD -log-input >/tmp/kol.log 2>&1 &
    KOL_PID=$!
    # 等視窗真的映射出來，不要用固定 sleep。
    WIN=""
    for _ in $(seq 1 100); do
      WIN="$(xdotool search --onlyvisible --name . 2>/dev/null | head -1 || true)"
      [ -n "$WIN" ] && break
      sleep 0.1
    done
    [ -n "$WIN" ] || { echo "視窗沒有出現"; cat /tmp/kol.log; exit 3; }
    xdotool windowfocus "$WIN" 2>/dev/null || true
    xdotool windowactivate "$WIN" 2>/dev/null || true
    sleep 1

    shot() { import -window "$WIN" "/tmp/$1.png" 2>/dev/null; }
    hash_of() { sha256sum "/tmp/$1.png" | cut -c1-16; }

    fail=0
    step=0
    prev=""

    # 等畫面真的變，不要用固定 sleep。合成事件送到 ebiten、ebiten 重畫、X 的
    # 視窗內容更新這三段各有延遲，固定 sleep 會在事件生效前就截到舊畫面
    # ——那會把「按鍵有效」誤判成無效。
    wait_change() {                 # wait_change <說明>
      step=$((step+1))
      local h=""
      for _ in $(seq 1 25); do
        shot "s$step"; h="$(hash_of "s$step")"
        [ "$h" != "$prev" ] && break
        sleep 0.2
      done
      if [ "$h" = "$prev" ]; then
        echo "  ✗ $1：等了 5 秒畫面仍無變化（$prev）"; fail=1
      else
        echo "  ✓ $1：$prev → $h"
      fi
      prev="$h"
    }
    baseline() { step=$((step+1)); shot "s$step"; prev="$(hash_of "s$step")"; }

    echo "== 鍵盤"
    # 標題序列自己會動，先離開它再取基準，否則「畫面有變」證明不了任何事。
    xdotool key --window "$WIN" Return; sleep 1.5
    baseline
    xdotool key --window "$WIN" Return
    wait_change "Return 翻到目錄頁"
    xdotool key --window "$WIN" 5
    wait_change "數字 5 進角色選項"
    xdotool key --window "$WIN" Down
    wait_change "方向鍵下移游標"

    # 現在停在角色選項頁、游標在第 1 項。先用鍵盤走到第 2 項並進去，記下那張畫面；
    # 待會用滑鼠點同一列，兩張必須一模一樣。**「畫面有變」是太寬鬆的判準**
    # ——點歪了也會變（落到 fallback 的 Return），只有「與鍵盤等價」才證明
    # 可點區域真的對上版面常數。
    echo "== 滑鼠與鍵盤等價"
    xdotool key --window "$WIN" Down
    wait_change "鍵盤：游標到第 2 項"
    xdotool key --window "$WIN" Return
    wait_change "鍵盤：進入第 2 項"
    KB="$prev"

    xdotool key --window "$WIN" Escape
    wait_change "Escape 回角色選項頁"
    xdotool key --window "$WIN" Up
    wait_change "游標上移"
    xdotool key --window "$WIN" Up
    wait_change "游標回第 0 項"

    # 中文模式：選單第一列在畫布 (48,104) 起、行距 24，畫布放大兩倍。
    # 第 2 列的中心在畫布 (100, 104+2*24+12)=(100,164) → 視窗 (200,328)。
    xdotool mousemove --window "$WIN" 200 328; sleep 0.3
    xdotool click --window "$WIN" 1
    wait_change "滑鼠：點第 2 列"
    if [ "$KB" = "$prev" ]; then
      echo "  ✓ 點第 2 列與鍵盤走到第 2 列是同一張畫面（$prev）"
    else
      echo "  ✗ 滑鼠點到別的地方：鍵盤 $KB，滑鼠 $prev"; fail=1
    fi

    xdotool click --window "$WIN" 3
    wait_change "右鍵回上一頁"

    # 真正的問題是「能不能玩」，所以要走到會動的畫面為止：選角 → 旅店 →
    # 出門 → 在城鎮裡走一步。
    echo "== 走進遊戲"
    xdotool key --window "$WIN" Escape
    wait_change "回目錄頁"
    xdotool key --window "$WIN" 6
    wait_change "目錄頁 6 進角色清單"
    xdotool key --window "$WIN" Return
    wait_change "選定角色進旅店"
    # 旅店圖示列第 0 格是 DOOR，用滑鼠點它出門——順便驗證圖示列的可點區域。
    # 樣本的白框停在第 5 格（COMPANION），所以這一點要送出五個 `<` 再 Return。
    # 中文模式圖示列在畫布 (64,340) 起、格距 64、格寬 48、格高 44，
    # 第 0 格中心 (88,362) → 視窗 (176,724)。
    xdotool mousemove --window "$WIN" 176 724; sleep 0.3
    xdotool click --window "$WIN" 1
    wait_change "滑鼠點 DOOR 出門到戶外地圖"
    xdotool key --window "$WIN" Down
    wait_change "在城鎮裡往下走一步"

    kill $KOL_PID 2>/dev/null || true
    wait $KOL_PID 2>/dev/null || true
    if [ $fail -ne 0 ]; then
      echo "-- 收到的 token --"; grep "^input " /tmp/kol.log || true
      echo "互動驗收未通過"; exit 4
    fi
    echo "互動驗收通過：鍵盤與滑鼠都推得動畫面"
  '
