# 最新版正常玩家路徑抽驗

日期：2026-09-08。結果：remake 正常射擊流程通過；dosgolem 原版 Olear 旅店入口通過。
本輪不是完整戰鬥同狀態對拍，也不是 AppImage 驗收。

## 執行檔與條件

- 以 `kol-package:20260907` 容器重建目前工作樹；完整
  `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 及建置通過。
- 執行檔：`workplace/fixed-seed-gui-20260908/kol`。
- SHA-256：`e1256c3c7d03f08950d071b235988fb0839f431dd41462bba8d72adcf833ce98`。
- remake 固定 seed=1；既有 GUI 腳本設 SELECT_DOWN=2、WORLD_STEPS=100、
  EQUIPMENT_PANEL=1、AMMUNITION_SHOT=1、KEY_DELAY=0.3，均使用 `KOL_` 前綴。
- dosgolem 使用 `check_updated_dosgolem.sh`，KOL_CHECK=book、KOL_BOOK_SELECT_DOWN=2。
  原版這條腳本未設定遊戲 seed，只驗選角及旅店入口，不將它用作亂數相關戰鬥收據。

## 實際結果

remake 經標題、書本、選 Olear、旅店招募、離店、出城、探索、進戰場與射擊結算，
完成 150 個正常輸入。人工檢視 `battle-equipment.png` 與
`ammunition-current-actor.png`：Olear 的 ARROWS:20 更新為 ARROWS:19，頭部傷勢
與疲勞同步顯示。這次未新增遊戲修正。

dosgolem 從正常 KOL.EXE 入口進入 KOLBOOK.EXE，再選 Olear 進入 KOLTOWN.EXE。
TestNormalBookEntry 通過（36.97 秒）。檢視 `original/book-final.png` 與
`inn-greeting.png`，角色 Olear、Halfway House／半途之家、店主肖像、招呼語內容
及六個圖示順序相符；中文畫面採既有排版，沒有宣稱像素相等。
日誌仍列 region1.dat…region4.dat 缺檔探測，本輪入口成功，不推論其他區域已驗收。

## 收據與限制

輸出根目錄：`workplace/fixed-seed-gui-20260908/`。
`gui-receipt.json` 保存 25 個輸出雜湊及輸入數，執行檔與腳本前後雜湊一致。
`original/manifest.json` 保存八個輸出；來源變動旗標為 false。
所有輸出擁有權為 1000:1000。首次原版容器漏掛 /out，未執行測試；修正掛載後，
以相同工具鏈與腳本乾淨重跑成功，分類為驗證環境錯誤。

剩餘：原版戰鬥固定 seed／可比狀態設定、異種武器拾取、弩與彈藥耗盡、完整存讀檔
及新版 AppImage。既有固定 seed 的常式測試可沿用，不新增自然 RND() 精確骰序閘門。
容器與背景 Xvfb 已結束，專案容器清單為空；未提交、推送或打包。
