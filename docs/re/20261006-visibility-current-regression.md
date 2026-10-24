# RE-291：移動可見性修正後的長任務與平台回歸

日期：2026-10-06。狀態：現行正式版有限任務與平台回歸通過，完整原版 UI 驗收保持開放。

本輪使用 RE-290 現行正式來源，342 份輸入、340 份 Go 檔案。Linux SHA-256 為
`991046abc4fb68103a2d5b7fae8f93a305aa5f2ff40a5fa33129968df85b8742`。
不改正式程式、資料格式、任務亂數、版號或發行包。

## 預先登記的驗證

- 沿用 RE-279 的 5971 鍵、執行前 seed 1、200 個明示遭遇骰 100。不得重播挑結果、重新播種或替換遊戲狀態。
- 英文與繁中各走正常招募、旅行、任務戰鬥、取物、逃離、回報、旅館保存，再以新程序載入實際 v3 存檔。
- 比較完整軌跡、只省略 saved_at 的完整存檔與 40 張 PNG。若舊按鍵受到修正影響，保留首次失敗，先找最早差異，再登記正常按鍵修訂。
- 由同一來源建置 Windows AMD64、macOS AMD64／ARM64／universal，檢查架構、依賴與最低系統版本。Wine 正常操作與實際任務保存重載必須符合現行 Linux。
- 重用 RE-290 已封存的十九套件測試，不重跑未變來源。不把明示遭遇控制、Wine 或交叉建置稱為自然全旅程、原版完整任務 parity 或平台真機驗收。

工作區入口：[prepare.py](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/visibility-current-regression-20261006/prepare.py)。
基準：[RE-279](20261006-move-current-regression.md)。現行修正：[RE-290](20261006-move-message-visibility.md)。
同輪新增本檔並掛入逆向索引；驗證結果、輸入雜湊與清理紀錄將封存於工作區。

## 首次反例與腳本修訂

原 5971 鍵在兩語言都退出 2，原始輸出留在 old-plan-failed-en／zh-Hant。
4781 鍵的正常前綴定位首個完整狀態差異為第 864 步。第四回合原先省略的怪物
移動訊息恢復後，確認鍵不足，後續規劃錯位並在未取物時逃走。

隔離控制器只將原 110 段執行提示的 SPACE 長度同步到現行 Prompt 10。
32 段長度改變，總鍵數增加 16 為 5987。其他正常動作、預先固定的種子、
明示遭遇骰與初始狀態保持。110 個段末的全部戰鬥狀態欄除 exec_text、一次性
walls 輸出相同；完整實際 v3 只差 saved_at。這是腳本產生證據，尚非正式 GUI 驗收。

正式重跑預先登記為 5987 鍵，取物預期 4789 步、第 69 回合，逃離問句 5598 步、
第 109 回合，5599 步返回世界。所有正常動作逐步對應記在 step-mapping.json。
40 個畫面索引沿原抽樣映射。完整舊 trace 因新增確認訊息不再要求逐步相同，
改核對映射後非執行訊息狀態及完整保存；exec_text 的原始差異完整保留。
新正式雙語 trace 仍須比較全部非顯示欄位，Wine 仍須與同版 Linux 完整相同。

## 正式結果

| 驗證 | 結果 | 限制 |
|---|---|---|
| 雙語正常任務 | 11976 狀態、取物、逃離、六人回報及旅館實際保存通過 | 固定 seed 1、200 個明示遭遇骰 100 |
| 映射後完整狀態 | 每語言 5001 個非執行訊息狀態與 RE-279 相符 | 只排 step／name 步驟 metadata 與既定顯示欄；本次非訊息狀態的顯示文字也零差 |
| 新程序重載 | 雙語 26 狀態與舊正常保存基準相符 | 實際本輪 v3，沒有替換資料 |
| 完整 v3 | 與 RE-279 只差 saved_at，兩語言持久資料相同 | 包括任務、完整物品、金幣與傷勢 |
| 抽樣畫面 | 40 張中 38 張相同；兩張逃離頁差異完整保留 | 英文 13066 像素、繁中 117814 像素，全部在戰場視圖 |
| Windows | 本輪 PE32+ AMD64 建置及鎖版音訊依賴通過 | Wine 證據不代替真機 |
| macOS | AMD64／ARM64／universal 與最低 11.0 相依檢查通過 | 本輪沒有 macOS 執行或真機驗收 |
| Wine 9 正常首場 | 雙語 772 狀態、12 張 PNG 與現行 Linux 完全相同 | 沿既有 385 鍵正常首場 |
| Wine 9 實際保存重載 | 雙語 26 狀態、6 張 PNG 與本輪 Linux 完全相同 | 本輪保存未改，另起程序 |

第 4789 步、第 69 回合 Fighting Bull 取得 Ship Wheel，第 109 回合逃離，
第 5599 步回世界。六人各回報一次，Treasure Map 納入完整保存。
最終位置為 town 7、entry 3、57,8。新正式版八個驗收程序共 12800 狀態、58 張 PNG。
另保留兩個舊按鍵失敗程序、4782 狀態的正常診斷前綴與隔離產生器的一個最末狀態；
全部 PNG 共 82 張。失敗程序未輸出完整 trace，不虛構其狀態數。

兩張逃離頁的問句及戰場外面板零差，已目視雙語取物、逃離、保存與重載頁。
本輪訊息可讀，未見新增裁切或圖像損壞。旅館與重載頁的肖像名牌保持既有截斷，完整名稱版面仍待確認。差異與 RE-290 保存上一張戰場的修正相符，
此因果分類為強推論；沒有本頁原版 oracle，不宣稱完整 UI parity。
兩張完整原始圖片、像素位置與判讀在 view-difference-review.json；不裁掉差異或改存檔讓它相同。

## 執行檔與保存檔

| 檔案 | SHA-256 |
|---|---|
| kol-linux-amd64 | `991046abc4fb68103a2d5b7fae8f93a305aa5f2ff40a5fa33129968df85b8742` |
| kol-macos-amd64 | `1545ef11436cfb021ae612901d26a5996a272a9a8491ddd5a11179bbcda7c89d` |
| kol-macos-arm64 | `b5705ec542316fa23e73501b50d7a5eb4eee327ab9e3da455157f86dcd7cbb64` |
| kol-macos-universal | `bbd6379b20b6b7b11841e8c0f06c1e5acce7334b6be00907a7ed94911d7b49d1` |
| kol-windows-amd64.exe | `68dfd3cb0b4ffc75ac01f30ef10fe5fe8d332ebbabd8c594f3c69c9e34353d7f` |
| formal-en/complete.json | `64e54925fa9e4fc566825a42ce9f2ab191407eec7b42335b65ac57f025d4cb74` |
| formal-zh-Hant/complete.json | `4ad1fd670dcb831e6277d51365d6140e2baa481a61a44a1decb66299a0b81d77` |

Linux 沿用 RE-290 正式程式；Windows 與 macOS 由同 342 份來源建置。
Go 1.26.7、Ebitengine 2.9.9 與音訊依賴保持鎖版，資訊在 platform-build-info.json。
RE-290 十九套件的 current-product-tests.log 隨外部來源雜湊保存，不重跑相同來源。

## 重驗與清理

工作區為 `workplace/visibility-current-regression-20261006/`，入口 prepare.py。
run.py／run.sh 重播新固定普通按鍵；review-mapped.py 核對映射、完整保存與圖片。
prepare-generator.py 的唯一修改是隔離副本 cmd/kol/main.go 的等待鍵控制器。
generator-inputs.json 保存完整來源與 110 段舊／新長度，正式 342 份來源保持。
build-windows.sh／mac-build.sh／check-platform.py 提供平台建置與結構檢查。
run-wine.sh／run-wine.py 分 normal、reload 兩階段；finalize.py 檢查本輪前綴擁有權後移除。
seal.py 保存附件、外部來源與 82 張 PNG 權利清冊，verify.py 為獨立唯讀核對入口。

重驗用固定 Go 映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`，專案唯讀掛 /src、
工作區唯讀掛 /out，以 UID/GID 1000:1000 執行 python3 /out/verify.py。
設定無網路、2 GiB、2 CPU、64 程序及 180 秒外層逾時，結果寫 /tmp/audit.json。
完整映像 ID 見 tool-images.txt。所有程序在 Docker 執行，使用 --rm、有界逾時與 Xvfb／Wine trap。
本輪 wineprefix-current 已核對擁有權後移除；容器終檢、無 root-owned 檔案及無 .md 目錄檢查隨收據保存。
畫面及保存屬僅本機資料，沒有公開散布授權；不變更版號、封包或發布。

清單保持 18 未完成、89 已完成、14 人工。任務共用亂數仍 DRAFT，HD 手繪媒材已選，
解析度與介面仍待確認。完整戰鬥／原版 UI、成功怪物拾取、真機、正式封包與發行沿原有完成條件。
