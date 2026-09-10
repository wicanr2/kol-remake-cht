# 跨城選角接線驗證

日期：2026-09-08。SPEC-016 接線擴充 READY，未宣稱 CONFORMED。

角色所在地代碼透過已載入世界資料尋找唯一城鎮入口，載入對應建築及地圖，
在暫存狀態驗證旅店與座標後才切換正式 Session。失敗保持原隊伍、城鎮與戶外指標。
同時修正野外入城未更新 `TownCode` 的遺漏，避免當地招募仍使用上一城代碼。
位置／鏡頭／邊界沿用 RE-069 與現有 EnterFromWorld 契約，不新增推測規則。

## 驗證結果

- 城鎮、flow、世界回歸通過；新增跨城成功／錯誤建築保留原狀，以及未知／重複城鎮代碼測試。
- `kol-package:20260907` 執行完整 `go test ./internal/... ./cmd/kol -count=1` 通過並建置成功。
- `KOL_SELECT_DOWN=12` 執行 `tools/verify_starting_inn_gui.sh`，從正常標題、目錄、
  選角頁選 Mula，不使用直接進場或座標注入。人工確認四張截圖：巨魔剋星旅店入口、
  候選人僅 ddduhh、兩人同隊、離店後正確城鎮地圖與徽章。
- 本機輸出 `workplace/cross-town-20260908/`，含測試紀錄、建置、GUI 輸入紀錄及截圖。
  `manifest.json` 是執行後盤點的檔案雜湊，不單獨作歷史建置來源證明。

本輪只證明 remake 跨城接線與正常 GUI 路徑；尚未用 dosgolem 重跑相同 Mula 狀態，
不能宣稱原版嚴格等價。完整角色保存、中文資料頁、其他旅店操作與 AppImage 驗收仍未完成。
所有輸出 UID/GID 為 1000:1000；一次性容器已清理，未提交、推送或打包。

## 後續 dosgolem 正常路徑驗證

同日追加：上方「尚未用 dosgolem 重跑」是第一輪交接狀態，現已完成跨城入口重跑。
沿用 `tools/oracle/check_updated_dosgolem.sh`，設定 `KOL_CHECK=book`、
`KOL_BOOK_SELECT_DOWN=12`。從唯讀固定原版 `KOL.EXE` 啟動，片頭以 DOS Enter 推進，
書本空白／6，再以 BIOS Down 十二次選 Mula，Enter／空白進入旅店；沒有修改記憶體或直接進場。

`workplace/dosgolem-mula-20260908/` 保存執行前來源清冊、紀錄及 dosgolem 自行重生的圖。
測試通過（48.963 秒）；`input_capture=before_copy`、`source_changed_during_run=false`。
原版實際開啟 `TOWN1.0`、`TWN1PICS.0`、`TWN1PICS.1`，不能把合成單元測試中的 Town=2
誤當成 BRT 的實際資料。

人工核對原版 `book-selection.png`：Mula 位於 TROLLSBANE／BRT。
原版 `book-final.png` 與 remake `cross-town-20260908/inn-greeting.png` 的相符範圍是：
Mula 頭像及隊長、Trollsbane Inn 及店主肖像、戰士招呼語的中文語意、六個圖示順序。
中文布局允許不同；此為入口可見狀態的抽樣，沒有藉此宣稱隱藏欄位、長按節奏、
招募、住宿、完整存讀檔或 AppImage 已完成同狀態驗收。
缺少 region1…4 的既有探測仍保留在原版紀錄，不補造原始檔案。
