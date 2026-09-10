# 玩家缺口補修：告別、競技場與天候

日期：2026-09-08。此收據只涵蓋下列邊界，不宣稱所有缺口已補齊。

## 告別流程（G04）

依 [SPEC-024](../spec/024-building-goodbye.md) 與
[原版證據](../re/20260908-building-goodbye.md)，單人直接離店；多人顯示建築訊息 17，
清除畫面上的發言人物框，等待空白鍵、確認鍵或等價圖示操作後回到門格。
等待期間的移動、詢問及取消不會提前退出，人物與存檔資料不變。

dosgolem `fd746405cf0ef9e535121e215b935b0bf59cd1de` 連同本機相容修正，
由正常 `KOL.EXE` 書本入口重新擷取單人及多人離店；收據分別在
`workplace/player-gaps-20260908/oracle-single/` 與 `oracle-party/`。
多人原版為 Olear／Theodoric，重製版為 Theodoric／Poontang；此處驗證人數分支、
相同旅店的告別文字與確認流程，不宣稱完整角色同狀態或像素相同。
原版多人 `battle-path-04.png` 顯示告別，`05` 返回城鎮。

新建置經 `tools/verify_goodbye_gui.sh` 正常操作：選角、單人退出、重新入店、招募、
多人告別、嘗試移動、確認退出。已目視告別文字及單一確認圖示；等待畫面與按移動後
像素差異為 0。`TestGoodbyeOnlyForMultipleMembers` 另驗 1／2／6 人與兩種確認鍵，
`TestDoorwayCancellationHasNoGoodbye` 驗門口取消不誤觸告別。
收據不涵蓋全部城鎮逐店文字，也未逐一抽測滑鼠點選。

## 競技場（G03：已實作，正常流程待驗）

[SPEC-025](../spec/025-arena-field.md) 已 READY。抽取器新增 `TOWN0` payload 85，
正式 runtime 讀 `battle-map-24.png`，每場先複製再走既有地形散布。
原 24 張地圖與其他既有 PNG 雜湊未改；新增圖已以實際磚塊合成全圖目視確認圍牆與地面。

dosgolem `TestArenaMapOriginal` 執行原版 RLE／轉置，4096 格一致；它以固定來源
替代 archive 讀取，不是正常競技場到場。重製版回歸驗兩次開戰不污染固定底圖，
非地面格不受散布影響；散布使用測試 seed=1，但本次未宣稱原版亂數結果相同。

## 天氣（G02）與失敗勘誤

`oracle-rules/player-gaps.log` 的六組天氣分派／等待參數通過；後續
`oracle-weather/player-gaps.log` 的合成比較失敗。原版使用 128 像素循環列並只重合成
上方 31 列；已依 [SPEC-026](../spec/026-weather-animation.md) 修正既有 64 像素模型，
預載六組圖層、雨雪三畫格及正常動畫節拍。主畫面以結局是否達成判斷是否暫停，
不能以結局物件存在判斷；正常 GUI 發現這個條件錯誤後已修正重驗。

原版初始化會先複製完整底圖到結果緩衝區；首版測試漏做，且僅配置 0x1900 bytes，
不足容納結果尾端 0x1930，造成剩餘約 73 個像素假失敗。兩項測試錯誤均已訂正，
不把它們列為產品缺陷。`oracle-weather-fixed/player-gaps.log` 最終四組 64×62
合成（phase=1）、六組分派及競技場 4096 格全部通過，manifest 記錄來源未變。

正常 `tools/verify_weather_gui.sh` 從書本選角、單人離店、出城後擷取四張畫面。
第一次腳本漏按城門確認，分類為驗證腳本錯誤，修正後同容器重跑。
最終首末畫面 2928 個像素改變，邊界為實體畫面 `(62,358)` 至 `(318,422)`（右下不含）的天空區，
天空外差異 0；日期與地圖未變。`tools/verify_player_gaps_receipt.py` 重驗這些條件，
`summary.json` 保存 15 個輸出的 SHA-256 與 UID/GID。
本次不是原版正常野外的同相位畫面對拍，雨雪正常 GUI 仍未抽驗；節拍為名義 BIOS
頻率近似，不宣稱硬體逐週期一致。

完整 `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 與建置已於
`kol-package:20260907` 容器重跑通過。測試、建置與 GUI 輸出位於
`workplace/player-gaps-20260908/`；不是新版 AppImage。繁中資料頁、編輯器、音訊、
加速選項、完整戰鬥同狀態及正式交付仍各自保留待辦。

結束檢查：本輪輸出均為 UID/GID 1000:1000，無本輪 Docker 容器殘留；
暫存研究、原型與快取保留於本機忽略目錄，其他專案資源未動。
