# RE-274：同格射擊修正後的長任務與平台回歸

日期：2026-10-06。狀態：現行來源有限回歸通過。

沿 [RE-273](20261006-ai-flank-normal.md) 的340份Go、342份來源與Linux03509ac1，
重跑 [RE-270](20261006-second-reload-current-regression.md) 既定5,971鍵、seed 1與200個明示遭遇判定100。
途中沒有重新設定種子或代換狀態。本輪正式來源未改，怪物v2與人物v3保持。

| 驗證 | 結果 | 限制 |
|---|---|---|
| 雙語長任務 | 11,944正常狀態與26新程序重載通過 | 遭遇判定沿既定控制條件 |
| 完整狀態 | 每語言完整trace與RE-270相同 | 同語言沒有排除欄位 |
| 整張畫面 | 40張PNG逐位元組相同 | 正常流程有限抽樣 |
| 人物v3存檔 | 完整資料只排saved_at相同，重載未改檔 | 未驗全24任務或原版整條任務 |
| Windows AMD64 | 離線建置、PE32+、音訊依賴及Wine9通過 | Windows真機未驗 |
| macOS | AMD64、ARM64與universal結構、最低11.0及系統相依通過 | 沒有macOS執行或真機驗收 |
| Wine正常首場 | 雙語772狀態、12PNG符合RE-273正式Linux | seed 1及385正常鍵，沒有遭遇覆寫 |
| Wine保存重載 | 雙語26狀態、6PNG符合本輪Linux，存檔未改 | 本輪實際v3檔，另起程序 |

第4,781步、第69回合Fighting Bull取得Ship Wheel，第109回合逃離，第5,583步回世界。
六人各回報一次，最後位於town 7、entry 3、57,8。Treasure Map與完整物品記錄納入保存比較。
跨語言只排既定四個顯示欄encounter_text、exec_text、prompt_lines、spoil_lines。
200個明示遭遇判定沿既定計畫，不稱自然完整旅程、全24任務或原版整條任務對拍。

## 程式與保存檔

| 執行檔 | SHA-256 |
|---|---|
| kol-linux-amd64 | `03509ac1818f59ba05fb854815e2d32df25fa2ea6d1433075fc439ef081a6ad6` |
| kol-macos-amd64 | `e05e5cc08832bc34914ae52f143f01ff9aca8252411e766147940b3163bbdc34` |
| kol-macos-arm64 | `a0aa150e6856634ff008863d007a70aa097c601d0d15f67910fa737474df44f4` |
| kol-macos-universal | `121e6e0634033b1fc994fa6379627ee3097aa7c3569c99fdbf479d1c3ffea453` |
| kol-windows-amd64.exe | `f1a5a1fa857cfe8a9d803d81b287071bdf1a17ab729f207e8cbfd0f6a2a784f4` |

Linux沿用RE-273正式程式，逐位元組核對，本輪不重編。Windows與macOS由同342份來源重建。
en實際v3 SHA-256：`7d5bb0a0184038b65b69785f3c44376b5f42c745342e407d0b7ad1023523c437`。
zh-Hant實際v3 SHA-256：`453345540cd2620f84c6bb53b7c92bd9cd862b5d0464da1f810db6007a6ac7ea`。

Go 1.26.7、Ebitengine 2.9.9與音訊依賴保持鎖版，資訊在platform-build-info.json。
固定映像ID在tool-images.txt。十九個套件沿RE-273最新正式收據，本輪未改來源，不重複測試。

## 本機收據與重驗

工作區為`workplace/missile-current-regression-20261006/`。
program-inputs.json核對342份來源。plan.json保留預先固定的操作、種子、遭遇判定及前版程式雜湊。
run.py／run.sh執行雙語正常任務、實際保存與新程序重載。
review.py比較完整trace、人物v3與PNG。
build-windows.sh、mac-build.sh及check-platform.py建置並核對結構、系統相依及音訊依賴。
run-wine.py／run-wine.sh分normal、reload階段，比對RE-273正常首場及本輪實際保存重載。
finalize.py核對整個前綴UID/GID後，只移除本輪wineprefix-current。
seal.py封存來源、畫面權利分類、附件與雜湊。verify.py獨立核對完整狀態、保存、PNG CRC與解碼、入口及擁有權。

重驗使用固定Go映像，專案唯讀掛/src、工作區唯讀掛/out，執行python3 /out/verify.py。
輸出寫/tmp/audit.json，UID/GID 1000:1000、無網路、2GiB、2CPU、64程序及120秒外層逾時。
工作使用--rm；Xvfb與Wine有trap清理。圖片與人物存檔只留本機，沒有公開散布授權。

## 仍開放

本輪只補齊RE-273新來源的有限長任務及平台回歸。成功側翼、撤退、撿武器、整場UI與原始50-byte仍開放。
RE-273的九組英文非滑鼠差異保持，未放寬允許矩形。
真機、macOS執行、正式封包及既有產品決策仍待驗。
HD已確認手繪媒材，解析度、版面與人物造型仍待既有決策。
清單18未完成／89完成、14人工保持。沒有Issue寫入、提交、推送、公開發行、版號或HD正式素材變更。
