# 2026-10-01 初始物品空格修正

已修正抽取器把空格尾端殘留誤當物品的缺陷。原版只檢查第一個位元組，
原抽取器則要求整格皆為零；固定名冊因而多出 21 件背包物品及 Mula 的一件手持物品。
[正式收據](inventory-empty-slots-20261001/receipt.json)保存來源雜湊、修正清單、
正常按鍵、三張目視畫格、工具及本機證據索引。

## 證據與實作

- 原版 ZIP SHA-256：`2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。
- CHARDATA SHA-256：`307ae9e2735b9f1d3d8008bcf018ed9125820af4e1ffadf078bc918be16f394e`。
- KOLTOWN-unpacked.EXE SHA-256：`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
- IDA Pro 9.4 重建 MZ 線性有效位址：購買 `sub_11A93:0x12F64..0x12F93`、
  物品頁 `sub_16986:0x16BF8..0x16C11`。原始函式、operand、bytes、xref 與分級註記
  留在 `workplace/inventory-empty-slots-20261001/ida-probe-reviewed.json`；正式資料庫未改。
- [RE-170](../re/RESEARCH-LOG.md#re-1702026-10-01-初始物品空格判準)與
  [SPEC-023 子契約](../spec/023-complete-save.md#2026-10-01-初始物品空格勘誤)先完成證據審查，
  再修 `tools/extract_character_roster.py` 的 `worn_slot`。

14 名角色共 238 格逐格核對；22 格改為空值，有效物品、其他角色欄位與原始 Carried 全部不變。
抽取前名冊 SHA-256 為 `f8f05ff26b1c278ab457616eb951fa1517c2e4ab8e7e8401a5c3151209fa972e`，
抽取後為 `d2b3666b606f7281845f9032673d80333132a65f938f7a5e1298a66ee2eb5413`。
三個空格／非空格邊界檢查通過，舊誤判資料作反面對照確實被拒絕。
重生入口為 `tools/extract_character_roster.py`；差分核對入口為
`tools/verify_inventory_empty_slots.py --before workplace/inventory-empty-slots-20261001/before-roster.json
--after assets/original-derived-data/character-roster.json --zip Knights-of-legend.zip`。
以上 Python 命令須在既有 `kol-package:20260907` Docker 容器中執行。

## 正常購買與保存

使用原 f969a63f 執行檔搭配新抽取資料，程式未重建；兩種語言皆在啟動前固定 seed 1。
從標題選角、旅店正常招募六人，再走至第七城第十三門，購買「吉拉克赫爾巴克」帶走。
Theodoric 初始背包為空，食品現在放第一格；檢視後回旅店住宿保存，再重新啟動讀該存檔。
沒有注入座標、隊伍或遭遇；保存格式仍為 v3。

| 驗證 | 結果 |
|---|---|
| 英中正常購買／住宿 | 各 199 筆狀態完全相同，金冠 1,009→974→918 |
| 英中正常重載 | 各 16 筆完全相同，第一格食品仍在 |
| 英中保存內容 | 排除 `saved_at` 後完全相同 |
| 六個文字框／四份來源圖 | 像素差 0，單墨跡像素反例拒絕；三張物品畫格目視可讀 |
| 舊 v3 相容性 | 舊收據存檔原樣讀入，21 筆狀態與舊重載收據相同，來源未改 |
| Go 產品測試 | Docker 內 `go test ./internal/... ./cmd/...` 通過 |
| 四條既有短回歸 | 商店 14、資料頁 10、住宿 11、讀檔 7 筆一致，畫格差 0 |

相容性首輪少給 `-shot-dir`，被 CLI 拒絕；補齊控制參數後在 `compat2` 乾淨重跑通過。
原失敗留本機 `compat/run.json`，不當成產品缺陷。

## 首輪完整鏈的基準差異（歷史）

`full-chain` 2,318 筆與 `full-chain-reload` 2,013 筆都抵達原完成點，所有抽樣畫格相同，
但嚴格軌跡比較各有二／三筆不同，仍列為未通過。分欄診斷確認只差 Carried 持物摘要；
軌跡僅在摘要改變時輸出 `town.party`，因此其中一步多出隊伍摘要。
其他狀態、摘要中的金冠／傷勢／任務值全部相同。

完整鏈保存差分也逐欄審查：只有空物品被移除、既有戰後 `SyncCarried` 隨之更新，
以及時間戳記；其他角色與頂層欄位相同。診斷不改比較器，不設忽略區、不覆寫舊收據。
完整診斷及原始失敗見本機 `battle-impact.json`／`battle-regression.log`，正式收據記錄其雜湊。
本地 #88 因新初始資料重新列待辦；下一閘門為保留舊證據、審查並重生受影響收據，再全量重播 29 條。

## 限制與清理

原版依據為靜態指令與固定資料，沒有新增原版動態同狀態對拍；不宣稱完整 parity。
不改法術空格、背包緊湊排列、其他交易規則，也不清洗玩家既有 v3 存檔。
舊商品與完整鏈資料按快照保留，不改名成新資料的驗收；本輪獨立重播的結果見下節。
原始 EXE、資料庫、名冊、存檔及完整軌跡只留本機；物品與天候衍生 PNG 適用既有 private GitHub
特例，沒有公開散布授權。Docker 採有界 `--rm` 與 UID/GID 1000:1000；最終清理結果記於 WORKLOG。

## 全量續核與結案（最新結果）

本地 #88 已重新結案。先保留 543 份舊基準及三份母存檔，再從固定新名冊重跑 29 條：
22 條軌跡相同，七條只有持物摘要不同。逐欄[審查結果](inventory-empty-slots-20261001/all-path-review.json)
與[採用紀錄](inventory-empty-slots-20261001/baseline-adoption.json)保留原值、新值及來源；
金冠、傷勢、任務與其他狀態不變，保存只差已證實的空物品、既有同步及時間戳記。
改金冠、改畫面模式與超出來源空物品數的持物差異，三種反例皆遭審查工具拒絕。
審查入口是 `tools/review_player_path_inventory_change.py`，在 `kol-package:20260907`
Docker 內執行，命令與固定輸入見正式收據。

更新七條摘要基準後，以原 f969a63f 程式、新名冊 d2b3666b 及原各路徑種子／受控輸入
再獨立重跑全部 29 條，14,875 筆嚴格狀態相同；讀檔依賴使用該次新保存。
[重播收據](inventory-empty-slots-20261001/replay-receipts-summary.json)及
[正式入口收據](inventory-empty-slots-20261001/canonical-receipts-summary.json)各 29 條、0 失敗。
住宿舊驗收仍要求 v1，已按 SPEC-023 的正式 v3 對應逐欄修正；沒有變更保存格式或玩家存檔。

首輪里程碑比圖另有五條引用較早的整圖天候。依已完成的 RE-164／SPEC-026，
保留舊圖，以七張正常路徑擷取另建引用，全部差異限於天候框；同執行檔、seed 1 與
90 步前綴的新舊名冊對照為 91 筆狀態及兩張圖完全相同，排除本次物品修正造成天候變化。
[天候引用勘誤](inventory-empty-slots-20261001/weather-reference-correction.json)保存圖號、
來源雜湊與 Group／Phase 候選。四份原版區域計畫的比較矩形、滑鼠例外及必要區域皆不變，
新畫格重驗原有 22 組比較全部通過；七張模板與兩張長鏈畫格完成來源合成核對，
其中四張模板已目視。原版完整天候畫面未重驗，不宣稱新原版 GUI 同狀態對拍。

首輪七條嚴格失敗、五條舊引用比圖失敗與原始快照仍留
`workplace/inventory-empty-slots-20261001/`；最新結果不抹去形成原因。
本輪未改正式 Go 程式、亂數條件、規則、保存格式或比較器；沒有提交、推送、遠端 Issue 寫入或發布。
