# RE-268：首名怪物裝填修正後的長任務與平台回歸

日期：2026-10-06。狀態：現行來源有限回歸通過。

沿 [RE-267](20261006-monster-first-reload.md) 的正式來源，337 個 Go 檔與 339 個輸入。
使用 [RE-259](20261006-ambush-current-regression.md) 既定的 5971 鍵、seed1 及 200 個明示遭遇骰 100。
途中沒有重播種、狀態代換或更換操作。正式程式與已核准資料契約保持。

| 驗證 | 結果 | 限制 |
|---|---|---|
| 雙語長任務 | 11944 正常狀態、26 新程序重載通過 | 固定控制遭遇的有限任務鏈 |
| 完整 trace | 雙語與 RE-259 全部相同 | 同語言所有欄位納入比對 |
| 抽樣整圖 | 40／40 PNG 與 RE-259 逐 byte 相同 | 正常流程的有限抽樣 |
| 實際人物 v3 保存 | 完整保存只排 saved_at 與 RE-259 相同，重載未改檔 | 原版整條任務仍未驗 |
| Windows AMD64 | 離線建置、PE32+、音訊依賴與 Wine 9 啟動通過 | Windows 真機未驗 |
| macOS | AMD64、ARM64、universal 結構，最低 11.0 與系統相依通過 | 尚無 macOS 執行 |
| Wine 正常首場 | 雙語 772 狀態、12 PNG 與 RE-267 現行 Linux 相同 | seed1、385 正常鍵，無遭遇覆寫 |
| Wine 真正保存重載 | 雙語 26 狀態、6 PNG 與本輪 Linux 相同，保存檔未改 | 從本輪實際 v3 檔以新程序重載 |

第 4781 步、第 69 回合 Fighting Bull 取得 Ship Wheel，第 109 回合逃離，第 5583 步回世界。
六人各回報一次，最終 location 為 town7、entry3、57,8。Treasure Map 與完整物品記錄都納入保存比對。
跨語言比較沿既定四個顯示欄：encounter_text、exec_text、prompt_lines、spoil_lines。
同語言與父收據的完整 trace 比對沒有排除欄位。
200 個明示遭遇骰 100 沿既定計畫，不宣稱自然完整旅程、全 24 任務或原版整條任務對拍。

## 執行檔與保存檔身分

| 執行檔 | SHA-256 |
|---|---|
| kol-linux-amd64 | dba813be51cc7fd59b0c081b50f9993db32ae5bc2a0a984f54dcf7e2ef9a8170 |
| kol-macos-amd64 | 1099fa7c4ca88eee0414e1d64f328ca155f214a851db97deafdcafcf603ef417 |
| kol-macos-arm64 | fee38aea33a70b700a1ceb2e71478631585e899fa0dffbf83f7731d4348c14ab |
| kol-macos-universal | a017d9a49b1484ff9c1912dc52da1ffdd3da7a02f778a2e756b34d838ec34739 |
| kol-windows-amd64.exe | 1e129a4059feed268718b16a97dbddc4199268f9919b758ce7fc1f23bbfff88b |

Linux 使用 RE-267 正式執行檔，逐 byte 身分核實，本輪沒有重編 Linux。
Windows 與 macOS 由同 339 個來源離線重建。
en 實際 v3 SHA-256：`36e9c7657eb6bd6083329f9e54884351c57c62310a9671354a063ac22fab114b`。
zh-Hant 實際 v3 SHA-256：`fc5d65ec96ccf8756b9584e932caf6e20c794769cac72f31ecf16f915b3562c7`。

Go 1.26.7、Ebitengine 2.9.9 與音訊依賴保持鎖版，資訊保存於 platform-build-info.json。
既有 Go、macOS 與 Wine 固定映像的 ID 保存於 tool-images.txt，沒有建立替代映像。
Wine 啟動、版本及 Xvfb logs 依 normal／reload 階段分開保存。

## 收據與重現

工作區為 `workplace/monster-reload-current-regression-20261006/`，入口 `prepare.py`。
run.py／run.sh 驅動雙語長任務、實際保存及新程序重載，review.py 比對 RE-259 的完整 trace、保存與 PNG。
build-windows.sh、mac-build.sh、check-platform.py 建置並驗證結構。
run-wine.py／run-wine.sh 分 normal、reload 階段，比對現行 Linux 收據。
finalize.py 合併收據，核對 UID 後只移除本輪 wineprefix-current，再更新目前文件。
seal.py 封存附件、來源、圖片來源分類及雜湊；verify.py 唯讀重驗完整 trace、保存與 PNG CRC、解碼及擁有權。

使用相同固定 Go 映像，專案唯讀掛 /src，工作區唯讀掛 /out，執行 python3 /out/verify.py。
重驗輸出寫 /tmp/audit.json。設定 UID/GID 1000:1000、無網路、2 GiB、2 CPU、64 程序及 120 秒逾時。
需要保留 audit.json 時，將已存在的本輪工作區另掛為可寫 /audit，設定 KOL_RE268_OUTPUT=/audit。
每個執行批次使用 --rm，Xvfb 等待 socket 並以 trap 收尾，Wine 服務同步清理。
PNG 與人物存檔只留本機，沒有公開散布或正式封包變更。

## 仍開放的完成條件

正式來源沒有修改，十九個套件沿 RE-267 最近正式測試，不重跑同一份測試。
本輪只補齊新來源的有限任務與平台回歸。Windows 真機、macOS 執行、正式封包仍待驗。
其他怪物裝填別名、受阻持物 DRAFT、完整 50-byte、完整 UI／操作與勝利 S 結尾正常 GUI 保持開放。
HD 已核准手繪媒材；解析度、介面、人物造型仍待確認。本機 18 未完成／89 完成、14 人工保持。
沒有 Issue 存取或寫入、提交、推送、版號、封包、發行及正式 HD 素材變更。
