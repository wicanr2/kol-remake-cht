# RE-270：第二名怪物裝填修正後的來源回歸

日期：2026-10-06。狀態：現行來源有限回歸通過。

沿[RE-269](20261006-monster-second-reload.md)的339個Go檔、341份來源與Linux340c8492，
依[RE-268](20261006-monster-reload-current-regression.md)既定5971鍵、seed1與200明示遭遇骰100重跑。
途中沒有額外重播種或狀態代換，怪物v2／人物v3及正式程式保持。

| 驗證 | 結果 | 範圍 |
|---|---|---|
| 雙語長任務 | 11944正常狀態、26新程序重載通過 | 既定有限任務鏈，控制遭遇明示 |
| 完整trace | 每語言與RE-268全欄位相同 | 沒有排除同語言欄位 |
| 整張圖片抽樣 | 40／40PNG逐byte相同 | 正常流程有限抽樣 |
| 真正人物v3保存 | 全部資料只排saved_at相同，重載未改檔 | 全24任務與原版整條任務未驗 |
| Windows AMD64 | 離線建置、PE32+、音訊依賴與Wine9通過 | Windows真機未驗 |
| macOS | AMD64、ARM64與universal結構、最低11.0及系統相依通過 | 沒有macOS執行或真機驗收 |
| Wine正常首場 | 雙語772狀態、12PNG符合現行RE-269 Linux | 固定seed1、385正常鍵，無遭遇覆寫 |
| Wine保存重載 | 雙語26狀態、6PNG符合本輪Linux，保存未改 | 本輪實際v3檔，另起程序 |

第4781步、第69回合Fighting Bull取得Ship Wheel，第109回合逃離，第5583步回世界。
六人各回報一次，最後位於town7、entry3、57,8；Treasure Map與完整物品記錄納入保存比較。
跨語言只排既定四個顯示欄encounter_text、exec_text、prompt_lines、spoil_lines。
200明示遭遇骰沿既定計畫，不稱自然完整旅程、全24任務或原版整條任務對拍。

## 程式與保存檔

| 執行檔 | SHA-256 |
|---|---|
| kol-linux-amd64 | `340c8492aafa789f480609a8b14d9b7dec2dd269f2e9e1e925f3fafab11c5051` |
| kol-macos-amd64 | `f983c3b474126f6bd2160d0e9f59ad17a06436f7a1d6f193bbb3253dfb427319` |
| kol-macos-arm64 | `54a3c1c6fec99aca143ef750a0f81797b3671a57230422f583e36af22d3a10e8` |
| kol-macos-universal | `f23bfb385968e30b742cc6f66dc12cc24c3b8d6cd3cca730b39e3a1a1816f1d6` |
| kol-windows-amd64.exe | `f9106204367304b4636e5e10614b16121a2795f43a25284e2c0a8294e8bc969c` |

Linux沿用RE-269正式程式，逐byte核對，本輪不重編。Windows與macOS從同341份來源重建。
en實際v3 SHA-256：`74a1d8f1b0b077531193fe080cec71d76a146202e8e93cf80f76603f3e0833a0`。
zh-Hant實際v3 SHA-256：`b54c5623d4cc5957ec5f7127a326ca41f86d147961925b4fc98f7ed129d08f0a`。

Go1.26.7／Ebitengine2.9.9及音訊依賴保持鎖版，資訊在platform-build-info.json。
Go、macOS、Wine固定映像ID在tool-images.txt；未建立替代映像。
十九個套件沿RE-269最新正式收據，不對未改來源重複測試。

## 收據與重現

工作區為`workplace/second-reload-current-regression-20261006/`，入口prepare.py。
run.py／run.sh驅動雙語正常任務、實際保存與新程序重載，review.py比較完整trace、v3與PNG。
build-windows.sh、mac-build.sh、check-platform.py建置並核對結構、系統相依與音訊依賴。
run-wine.py／run-wine.sh分normal、reload階段，比對現行Linux正常首場及本輪實際保存重載。
finalize.py合併收據，核對整個前綴UID/GID後只移除本輪wineprefix-current。
update-docs.py更新目前狀態與既有工作項；seal.py封存來源、圖片來源分類、附件及雜湊。
verify.py以唯讀來源核對完整trace、保存、PNG CRC／解碼、入口與擁有權。

重驗沿固定Go映像，專案唯讀掛/src、工作區唯讀掛/out，執行python3 /out/verify.py。
輸出寫/tmp/audit.json；UID/GID1000:1000、無網路、2GiB、2CPU、64程序，120秒逾時。
每批用--rm，Xvfb有socket正對照與trap，Wine服務同步清理。
圖片與人物存檔只留本機，不包含公開散布授權；正式封包與版號保持。

## 仍開放

此收據只補齊新來源有限長任務與平台回歸。真機、macOS執行、正式封包仍待驗。
首名／第二名來源的有限驗收見RE-269；其他別名、原始＋0x2C／＋0x31、完整UI與整場保持。
受阻持物DRAFT、勝利S結尾正常GUI及既有操作問題不因本輪關閉。
HD沿已確認手繪媒材，解析度、版面與人物造型仍待既有決策。
清單18未完成／89完成、14人工保持。沒有Issue寫入、提交、推送、公開發行或HD正式素材修改。
