# 規劃段逐格走完（2026-09-27，#84）

## 結果

同一場正常遭遇裡，Theodoric 主指令列上的每一個指令、徒手攻擊式與防禦方式的每一格、
移動方式的每一格，原版與 remake 照同一串鍵走，**每一鍵之後的圖示列格數、白框位置，
以及寫回的指令碼、攻擊選項、防禦選項與方向全部一致**。比的是機制，不比畫面。

```
tools/run_original_parity_tests.sh      # TestOriginalNormalPlanningCells
```

第一次比對時移動方式那一列原版 7 格、remake 3 格，追出兩個 remake 的錯：參戰者
`+0x03` 其實由種族決定（remake 一律當 0），以及磚塊 0x0C 在移動方式列上是**擋住**
後三格而不是解鎖。兩者依 RE-148 修正後全數一致。

## 走了哪些格

Theodoric 的主指令列是 `YES、0x29 徒手、0x16 切換、0x27 移動、5 休息`，他是 Kelden、
手上沒有武器。

| 組 | 路徑 | 寫回（指令／攻擊／防禦） | 驗到的規則 |
|---|---|---|---|
| `u-s1-g1` | 徒手 → 攻擊式 0x2A → 防禦式 0x2A | 5／0／0 | 兩個都是 0 改成休息 |
| `u-s2-g2` | 徒手 → 0x37 → 0x30 | 0x29／0／1 | 防禦選項 1 時攻擊選項歸零 |
| `u-s3-g4` | 徒手 → 0x38 → 0x32 | 0x29／2／3 | 圖號 − 0x36、圖號 − 0x2F |
| `u-s4-g5` | 徒手 → 0x39 → 0x33 | 0x29／3／4 | 同上 |
| `u-s5-g6` | 徒手 → 0x3A → 0x34 | 0x29／4／5 | 同上 |
| `u-s1-g7` | 徒手 → 0x2A → 0x36 | 0x29／0／6 | 攻擊式 0x2A 夾成 0；防禦 7 減 1 |
| `switch` | 切換 | 0x16／0／0 | 直接收尾 |
| `move-1`…`move-6` | 移動 → 選目標格 → 移動方式第 1…6 格 | 0x27／圖號／0x3C 時 2 | 選中的圖號本身；走路那一格防禦 2 |
| `rest` | 休息 | 5／0／0 | 直接收尾 |

選目標格一律按 Return 確定面前那一格，方向寫回 0。

## 原版那一側

`tools/oracle/book_startup_test.go` 的 `KOL_BOOK_PLAN` 模式，按鍵由
`KOL_BOOK_PLAN_KEYS` 給；dosgolem `968ca77`（`~/cht/dosgolem-kolrand`），seed 1、正常
路徑，每一組獨立跑一次：

```
docker run --rm --network none --memory 4g --cpus 2 --pids-limit 512 \
  -u "$(id -u):$(id -g)" -e HOME=/tmp -e KOL_CHECK=book \
  -e KOL_BOOK_SELECT_DOWN=2 -e KOL_BOOK_BATTLE=1 -e KOL_BOOK_WALK=1 -e KOL_BOOK_NEXT_DAY=1 \
  -e KOL_BOOK_ENTER_FIGHT=1 -e KOL_BOOK_PLAN=1 -e KOL_BOOK_PLAN_KEYS=<序列> \
  -v "$PWD:/src:ro" -v ~/cht/dosgolem-kolrand:/repo:ro -v <輸出>:/out \
  --tmpfs "/tmp:exec,uid=$(id -u),gid=$(id -g),size=2g" \
  eob-remake-go:1.26.7-ebiten2.9.9-audio bash /src/tools/oracle/check_updated_dosgolem.sh
```

輸出在 `workplace/planning-cells-20260927/<組>/`（畫格與原始記錄，不進版控）。
`tools/summarize_planning_probes.py` 先確認每一組的進場狀態
（`battle-state-records.json`）與 2026-09-11 那一場逐 byte 相同，再摘成
`planning-cells.json`；不同就整份拒寫。

## remake 那一側

`TestOriginalNormalPlanningCells` 用同一份隊員資料與地形建場，把規劃推到槽 1，
`.`／`,` 對到 `MoveSelection(±1)`、Return 對到 `Choose`，每一鍵之後比格數與白框，
走完比四個值。寫回之後原版換到下一名的指令列、remake 只排了一名，所以最後一鍵只比
四個值。

## 正反對照

修正之前這一支測試在三組移動方式上失敗，訊息列出 remake 的 3 格與 4 格圖示列——
訊號真的會開口，失敗也正是這一輪修正的來源。

## 不宣稱

- 武器攻擊式（0x21）、射擊（0x22／0x23）、施法（0x0D）與丟、撿不在 Theodoric 的
  指令列上，這一場走不到；換算規則由 `commands_test.go` 逐圖號驗算。
- 選目標格只按 Return，沒有移動游標；游標規則見 SPEC-010 §3 的單元測試與
  `battle-target-mouse` 收據。
- 只比規劃段寫回的值，不含執行段的結算。
