# 開場旅店修復與實際操作驗證

日期：2026-09-08。規格：SPEC-016 READY；尚未 CONFORMED。

## 修正

開場選角改由角色城鎮代碼、留宿建築編號與所在地名稱定位正式建築，建立獨立隊員後
進入同一個 `town.Session`。對話、招募、住宿與離店不再走只支援 DOOR 的靜態樣本。
不一致所在地保留選角頁及原隊伍；跨城載入仍未完成，不能把此拒絕分支視為完成。

實際操作另發現 `InnCJK` 漏畫 `Inn.List`，造成有候選人卻看不見名字。
已在既有對話框補八列、選中顏色及捲動指示，名稱超寬逐字截斷。

## 驗證

- Docker 映像：`kol-package:20260907`。完整 `go test ./internal/... ./cmd/kol -count=1`
  通過；新流程回歸含標題至旅店、招募、離店與錯誤所在地保留原狀。
- 建置使用 `workplace/spec-fixes-20260907/run.sh`；一般 GUI 使用
  `tools/verify_starting_inn_gui.sh`，Xvfb 與 xdotool 從正常視窗輸入，沒有 direct-entry。
- 人工逐張確認本機 `workplace/starting-inn-20260908/` 的 `inn-entry.png`、
  `inn-recruit-list.png`、`inn-two-members.png`、`inn-exit.png`：半途之家敘述、
  八列候選人（包括 Fighting Bull）、Theodoric 與 Poontang 同隊、兩人走到戶外。
- 第一輪空白清單及過早擷取入口保留在 `first-gui-attempt/`。
  後者為腳本等待不足，改成等待輸入紀錄後再抓圖；同容器命令乾淨重跑確認。
- 原版入口證據由 dosgolem 自行重生，見 `docs/re/20260908-kolbook-input-bridge.md`。
  本輪沒有逐項對照原版招募、住宿與存讀檔，不能宣稱完整操作等價。

收據 `manifest.json` 記錄本機建置與擷取雜湊；來源清冊為執行後盤點，不能單獨證明
歷史建置輸入。建置後只有 UI 原始碼的 gofmt 格式整理，未再改行為。
所有輸出 UID/GID 為 1000:1000，一次性容器已清理，未提交、推送或更新 AppImage。

## 剩餘工作

跨城選角、完整角色保存、中文資料頁與其他未呈現的內頁、旅店各功能的原版同狀態抽測，
以及乾淨重打包後 AppImage 正常玩家路徑驗收仍未完成。

## 招募滑鼠接線後續驗證

2026-09-08 追加：候選人列原先沒有命中範圍，點擊落到 Return fallback，會確認原本
反白者。已補中英文列範圍，將點擊轉成既有上下移動、回到確認圖示及 Return。
完整 `go test ./internal/... ./cmd/kol -count=1` 與乾淨建置通過。

以 `KOL_RECRUIT_MOUSE=1` 執行 `tools/verify_starting_inn_gui.sh`：正常開場進半途之家，
招募清單初始反白 Poontang，先將圖示選到返回，再實際點第二列 Olear。
輸入紀錄為畫布座標 `(220,196)`、`token=[D < RET]`；人工確認
`workplace/recruit-mouse-20260908/inn-two-members.png` 與 `inn-exit.png` 均有
Theodoric／Olear。沒有直接指定角色或注入隊伍。

本輪 GUI 抽測繁中、未捲動清單；英文及捲動後列差換算目前只有回歸證據，
不擴張為所有滑鼠路徑或 AppImage 驗收通過。本機 manifest 保存建置、紀錄與 PNG 雜湊，
來源清冊標為執行後盤點。所有輸出 UID/GID 1000:1000，一次性容器已清理。
