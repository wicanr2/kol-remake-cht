# AppImage 與 dosgolem 操作驗證

後續訂正（同日）：使用者已確認滑鼠可取代鍵盤，以結果等價為準。因此下列裝置／按鍵
映射差異不能單獨用來否決重製版；完整等價仍未驗證。後續工作已轉為
[原始碼與規格稽核](spec-source-audit-20260907.md)，新發現的規則與保存缺口另列於該報告。
下文保留本次原始實驗與當時判讀，不將舊結果冒稱為新判準下的通過證據。

日期：2026-09-07。結論：**未通過「操作感一致」驗收**。AppImage 基本鍵鼠流程可用，
地圖點擊的操作語意已有差異證據；正常原版啟動鏈尚未走到城鎮，不能宣稱完整同狀態對拍。
本輪只驗證，未修改遊戲或重建封包。

## 目前狀態與判準

| 項目 | 結果 | 證據等級／限制 |
|---|---|---|
| AppImage 啟動、選角、進旅店、出門、城鎮移動 | 通過抽測 | 已證實；直接執行既有包，正常鍵鼠輸入 |
| 選單第三列的滑鼠與鍵盤等價 | 通過 | 已證實；兩條操作所得畫面 SHA-256 相同 |
| 地圖滑鼠操作一致 | 不通過目前驗收 | AppImage 點地圖送 `RET` 已證實；原版可點地圖移動為輔助實驗支持的強推論，未完成正常 EGA 同狀態驗證 |
| 長按與操作節奏 | 未知 | AppImage 長按 Down 1 秒僅記一次 `D`；原版長按未驗，不從單次鍵盤佇列推導實體鍵盤重複速度 |
| 原版正常啟動至城鎮 | 本輪未達成 | dosgolem 可從 TITLE 送 Enter 至 KOLBOOK 載入，但後續按鍵未推到城鎮 |
| 戰鬥、存讀檔、音訊與完整主觀手感 | 未驗 | 不用舊像素收據或單元測試替代本輪 AppImage 驗收 |

驗收要求是相同玩家動作有相同操作語意、可觀察狀態轉換與可比較節奏；
畫面變動或探測測試印出 PASS 本身不代表通過這個要求。

## 固定輸入

- AppImage：`dist-all/0.2.1/packages/kol-0.2.1-full-linux-x86_64.AppImage`。
  SHA-256：`af36f9a6958a69a783ca638d45bfc3f142660aa314139a3407cfed87d576868f`。
  使用 `--appimage-extract-and-run`，因此未涵蓋主機 FUSE 掛載與桌面整合。
- 原始 ZIP：`Knights-of-legend.zip`，SHA-256：
  `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。
- dosgolem：`/home/anr2/cht/dosgolem-kolrand`，提交
  `7f30ad8427086738b037c50a417baf3d0d56a5d2`。
- 專案 HEAD：`ecd9113e4601ad117e0349096f6fe2415d043204`。
- 容器：`kol-package:20260907`，image ID
  `sha256:3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193`，Go 1.26.7。
- 其餘來源、探測腳本、截圖與日誌 SHA-256 見本機
  `workplace/appimage-dosgolem-audit-20260907/manifest.json`。原版衍生畫面僅留本機。

## 實際操作與結果

AppImage 使用 `tools/verify_interactive.sh` 的容器內測試主體，補上背景程序清理與日誌保存。
從標題按 Return、翻到目錄、5 進角色選項、方向鍵選列、Escape 返回，並以滑鼠重走
同一列；再回目錄、6 選角、Return 進旅店、點 DOOR 出門、Down 移動。
`interactive.log` 全部通過，截圖 `s15.png`、`s16.png` 目視確認隊伍標記向下移動。
另驗批次截圖 83 張及兩張標題圖，僅作既有封包功能抽測。

補充實驗在城鎮地圖視窗座標 `(900,400)` 點左鍵，輸入日誌為
`游標=(450,200) 可點=0 token=[RET]`；接著長按 Down 1 秒只記一筆 `token=[D]`。
完整輸入見 `interaction-probe.log`。目前 `cmd/kol/input.go` 的 `flowTargets` 也只列城鎮圖示，
不列地圖移動點；野外與戰鬥回傳空命中區。這是限制線索，不把未實跑戰鬥寫成實測。

dosgolem 重跑 `TestDumpTownMove`：直接啟動 `KOLTOWN-unpacked.EXE`，設定
BIOS 資料區（BDA）`0040:0049=9`，再以滑鼠關閉文字及點 U-TURN 出門。
八個方向／數字鍵與六個地圖點擊都有畫面變動；目視 `move-00-map.png` 與
`move-01-右一格.png` 可見地圖與隊伍位置變化。**此組是 Tandy16 解碼及初始化輔助實驗**，
不是正常 EGA 玩家啟動收據，不以兩張不同狀態的畫面宣稱逐像素或方向完全一致。
既有遊戲副本的 75 個檔案逐一與 ZIP 比對相同，見 `original-inputs.json`。

另從 ZIP 在容器內重新抽出資料，以未替換的 `KOL.EXE` 正常啟動；無記憶體注入、
無替代 TITLE。每次執行四千萬道指令，輸入依序為等待、Enter、Enter、6、Enter、Enter。
第一次 Enter 後實際載入 `KOLBOOK.EXE`；之後停在 dosgolem 執行期
`0110:3CA4…3CC5` 附近，未新增 BOOKDATA 或 KOLTOWN 的開檔紀錄。
這些是 dosgolem 執行期 `segment:offset`，不是 IDA 位址。
未實作服務報告包含 `int 21h AH=29` 與 `AH=67`，**尚未證實它們是停滯根因**。
`normal-keys.log` 的 PASS 只表示探測沒有拋出錯誤，不表示正常玩家鏈通過。

## 下一個驗證入口與清理

先讓 dosgolem 從 KOLBOOK 正常載入書本並能選角，再重生正常 EGA 城鎮收據，
用相同起始隊伍與位置比較地圖點擊、取消、長按與戰鬥選目標。
現階段不得把舊 DOSBox 圖片登錄成 dosgolem 收據，亦不應宣稱操作感已一致。

本輪容器皆 `--rm`、`--network none`、UID/GID 1000:1000、有記憶體／CPU／程序數及外層逾時限制。
原始資料唯讀，遊戲副本寫容器 `/tmp`；37 個保存檔案擁有權抽查全部為 1000:1000。
暫存證據保留在上述本機目錄；未提交或推送，未修改其他專案或清理其資源。
