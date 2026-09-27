# 野外時刻的每步成本（2026-09-27，#65）

原版在平原上四個方向各走八步（往西十六步），每一步的時刻、時段、`word_39A54`、隊伍移動率與地形碼
（`world-clock-probe.json`）。每步扣的是 `2 × ([0x12B8 + dy] + 移動率)`：往北 40、
往南與東西 60，與腳下地形無關（RE-151）。

重跑（dosgolem `968ca77`，seed 1、正常路徑）：

```
docker run --rm --network none --memory 4g --cpus 2 --pids-limit 512 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp -e KOL_CHECK=book \
  -e KOL_BOOK_SELECT_DOWN=2 -e KOL_BOOK_BATTLE=1 -e KOL_BOOK_WALK=1 -e KOL_BOOK_NEXT_DAY=1 \
  -e KOL_BOOK_CLOCK_PROBE=1 \
  -v "$PWD:/src:ro" -v ~/cht/dosgolem-kolrand:/repo:ro -v <輸出>:/out \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=2g" \
  eob-remake-go:1.26.7-ebiten2.9.9-audio bash /src/tools/oracle/check_updated_dosgolem.sh
```

remake 那一側由 `internal/world/clock_test.go` 的 `TestClockStepCost`、
`TestClockPartChangesOneStepLate` 與 `TestRoadRunClock` 驗同一組數字與時段晚一步換的
規則（後兩支對照原版錄影 `world-clock-run2`、`world-road-run1`，見 RE-151）。

往西那一組走十六步並讀時段 `word_2DF42`（`DS:0x72`）：第 14 步時刻 3360 已跨過 DAWN 的
下限，時段仍是 0；第 15 步開頭換成 1——時段在扣之前判（RE-151）。第 15 步之後時刻沒有
再扣、位置也不再動：換時段那一下擲出了遭遇，探針讀記憶體時原版正停在遭遇裡等鍵，這兩步
的數字不代表一般步伐。

第五組（往北十二步再按九次 Return）在這條路徑上沒有意義：這一次的出城點 (49,23) 不在
道路上，Return 按的不是 ROAD，時刻不動。循路的成本由 remake 的循路軌跡與舊錄影
`world-road-run1` 對照（`TestRoadRunClock`）。

不宣稱：只量了平原與移動率 20；其他地形依靜態讀法同樣不影響成本，沒有逐一實測。

`screen_changed` 一欄在這一段全為 false：那個指紋取的是書本畫面的緩衝區，沒有涵蓋
野外地圖，不當證據。步伐有沒有走出去看的是時刻有沒有扣。`word_39A4E` 在換格那一步
讀到 1251，是這一步之後被別的程式借用覆寫，與成本無關。
