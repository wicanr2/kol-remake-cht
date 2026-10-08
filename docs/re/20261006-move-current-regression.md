# RE-279：最新 UI 修正後的長任務與平台回歸

日期：2026-10-06。狀態：現行來源有限回歸通過。

沿 [RE-278](20261006-full-move-ui.md) 的 340 份 Go、342 份來源與 Linux d43a6385，
重跑 [RE-274](20261006-missile-current-regression.md) 既定 5,971 鍵、seed 1 與 200 個明示遭遇判定 100。
本輪沒有改正式來源、途中重設種子或代換狀態。怪物 v2 與人物 v3 保持。

| 驗證 | 結果 | 範圍 |
|---|---|---|
| 雙語長任務 | 11,944 正常狀態、26 新程序重載通過 | 沿既定遭遇控制條件 |
| 完整狀態 | 同語言完整 trace 與 RE-274 相同 | 沒有排除狀態欄位 |
| 整張畫面 | 40 張 PNG 與 RE-274 逐位元組相同 | 預先登記的流程抽樣 |
| 實際人物 v3 | 完整資料只差 saved_at，重載未改檔 | 本次任務與六人隊伍 |
| Windows AMD64 | 建置、PE32+、音訊依賴及 Wine 9 通過 | Windows 真機未驗 |
| macOS | AMD64、ARM64、universal 結構與系統相依通過 | 最低 11.0，未執行或真機驗收 |
| Wine 正常首場 | 雙語 772 狀態、12 PNG 符合 RE-278 正式 Linux | seed 1 與 385 正常按鍵，沒有遭遇覆寫 |
| Wine 重載 | 雙語 26 狀態、6 PNG 符合本輪 Linux | 本輪實際 v3，另起程序 |

第 4,781 步、第 69 回合 Fighting Bull 取得 Ship Wheel，第 109 回合逃離，
第 5,583 步回世界。六人各回報一次，最後位於 town 7、entry 3、57,8。
Treasure Map 與完整物品記錄納入保存比較。
跨語言只排既定顯示欄 exec_text、encounter_text、prompt_lines、spoil_lines；保存資料全符。
這是新版有限回歸，不宣稱自然完整旅程、全 24 任務或原版整條任務對拍。

## 程式與保存檔

| 執行檔 | SHA-256 |
|---|---|
| kol-linux-amd64 | `d43a6385a9c35486cc2cd383f78ffce23d70820f1fa113c98559d1760ff2b647` |
| kol-macos-amd64 | `5c291d6b40656c009996140951fe59d7aa2095d78c25869b86ab220435f562d6` |
| kol-macos-arm64 | `1c4b531f6d2496b77ce2d3f2ad87958cc89103d6e4f5b893a7d49742caac8990` |
| kol-macos-universal | `5ab98dda03e7175caabd41add0e419eb7baf84880c4bce88b4e0687c00577335` |
| kol-windows-amd64.exe | `a51d1a12bab64c1d8c3439a73df54ce0b4bbe7e54ff80d29db81964d161dc35d` |

Linux 沿用 RE-278 正式程式並逐位元組核對，Windows 與 macOS 從同 342 份來源重建。
en 實際 v3 SHA-256：`c06ad7d53fa80c2f805204b1b2dff665dba02779699e105448e7f324682f5471`。
zh-Hant 實際 v3 SHA-256：`6d3786a06d2f15b1f2be1a7502763a4593cf8e38f1a4812fc735b75c446cc906`。

Go 1.26.7、Ebitengine 2.9.9 與音訊依賴保持鎖版，資訊在 platform-build-info.json。
固定映像 ID 在 tool-images.txt。十九套件沿 RE-278 正式收據，不重複測試相同來源。

## 收據與重驗

本機工作區：`workplace/move-current-regression-20261006/`。
prepare.py 核對現行來源與固定 plan，run.py／run.sh 執行雙語任務、保存及新程序重載。
review.py 比較完整 trace、完整保存資料及所有抽樣 PNG。
build-windows.sh、mac-build.sh、check-platform.py 驗證跨平台程式結構與依賴。
run-wine.py／run-wine.sh 分 normal、reload 兩階段核對正式 Linux。
finalize.py 先驗前綴擁有權，再移除本輪 wineprefix-current。
seal.py 保存附件、來源與權利清冊，verify.py 獨立核對雜湊、完整狀態、保存、PNG CRC 與解碼、入口及擁有權。

重驗使用固定 Go 映像，專案唯讀掛 /src、工作區唯讀掛 /out，執行 python3 /out/verify.py。
輸出在 /tmp/audit.json；UID/GID 1000:1000、無網路、2 GiB、2 CPU、64 程序及 120 秒外層逾時。
所有容器使用 --rm 與 --init，Xvfb／Wine 以 trap 清理。原版衍生畫面與存檔僅本機，沒有公開散布授權。

## 仍開放

成功側翼、撤退、撿武器、其他戰局、完整原始記錄與玩家操作仍沿既有驗收項目。
RE-278 的有限同相位畫格證據不外推本輪長任務或其他戰鬥。
真機、macOS 執行、正式封包及既有產品決策保持待驗。HD 手繪媒材已選，解析度與介面仍待核定。
清單保持 18 未完成、89 完成、14 人工。沒有 Issue 寫入、提交、推送、版號或公開發行。
