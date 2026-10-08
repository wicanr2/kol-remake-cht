# RE-257：怪物名稱修正後的長任務與平台回歸

日期：2026-10-06。狀態：現行來源的有限回歸通過。

沿[RE-256](20261005-monster-plural-ending.md)正式336個Go檔與338個輸入，
Linux執行檔420e0e47保持。本輪沒有修改正式程式、資料契約或玩家操作。

| 驗證 | 結果 | 限制 |
|---|---|---|
| 雙語長任務 | 11,944個正常狀態、26個新程序重載狀態；完整trace與RE-255相同 | 5971個預定鍵、seed1與200個明示遭遇骰100，途中不重播種 |
| 取物與返回 | 第69回合Fighting Bull取Ship Wheel，第109回合逃離，六人各回報一次 | 有限任務鏈，不代表自然完整旅程或全24任務 |
| 實際v3保存 | 完整保存只排saved_at與RE-255相同；兩份檔案經新程序重載 | 不把單元測試或注入狀態當正常保存 |
| 長任務圖片 | 40張完整PNG與RE-255逐byte相同 | 沒有新增原版整條任務對拍 |
| Windows建置 | PE32+ AMD64、離線建置與音訊依賴通過 | Wine9實行，尚無Windows真機 |
| macOS建置 | AMD64、ARM64及universal，最低11.0、系統相依通過 | 沒有macOS實際啟動 |
| Wine正常路徑 | 雙語872個狀態、16張PNG與RE-256現行Linux相同 | seed2、435正常鍵、十四回合首場；沒有遭遇覆寫 |
| Wine實際重載 | 雙語26個狀態、6張PNG與本輪Linux相同；保存檔未改 | 使用本輪兩份實際v3檔案，各以新程序啟動 |

最終六人金冠為0／0／1721／0／386／0，傷勢為3／1／26／1／1／10。
Olear與Fighting Bull各付56住宿，其餘四人依正常不足提示略過；旅店保存佇列結束。
完整保存比對涵蓋所有欄位與物品記錄，只排saved_at。
雙語語意比較只排encounter_text、exec_text、prompt_lines、spoil_lines四個既定顯示欄。

## 輸入與輸出身分

| 執行檔 | SHA-256 |
|---|---|
| Linux AMD64，沿正式執行檔 | 420e0e4744c8151b16c3631b61b47f56ce095455c89114e4452ed637fd8853d4 |
| Windows AMD64 | a099c0b0a32f12ff96d741262308dbffb17199f1d3561a69ffef7f135db8d1e9 |
| macOS AMD64 | 31b469977e32a43b418ba71ab2b4cd9b9cb5309996110b13b7aab81086bea981 |
| macOS ARM64 | 2003123243683a37104dc1ddf35d48f76c371d3d374d1907f4ea432e26a89f45 |
| macOS universal | bd50132f9f7c009de3f099a6357b8a6f142775d60cda4c066963d6e39b9fd278 |

Go1.26.7、Ebitengine2.9.9與音訊依賴沿既有鎖版。
固定映像與完整ID保存tool-images.txt：
eob-remake-go:1.26.7-ebiten2.9.9-audio、
eob-remake-macos:1.26.7-ebiten2.9.9-audio、coab-wine-smoke:ubuntu-noble-20260826。
英文v3檔案SHA為011a6567d9aa92619ef0f09664a878584a741f2ab06e94294a6a2620675086fb，
繁中為776128dbfe5708d19fda46dff9944986f6c814e5d80a133459ad36597f8dac0c。

## 收據與重現

工作區為`workplace/plural-current-regression-20261006/`。
入口prepare.py固定338個來源、正式執行檔及RE-255既定計畫；run.py／run.sh跑長任務，
review.py逐項比對前次完整trace、保存及PNG。
build-windows.sh、mac-build.sh、check-platform.py建置並驗證架構。
run-wine.py／run-wine.sh分normal與reload兩階段核對Linux控制。
finalize.py合併Wine收據、核對UID後只移除本輪wineprefix-current，並更新目前文件。
seal.py封存；verify.py重算附件、外部輸入、PNG解碼與擁有權。
重驗時來源唯讀掛至/src、此工作區唯讀掛至/out，同工作區可寫別名掛至/audit，
設定KOL_RE257_OUTPUT=/audit，在固定Go映像內執行python3 /out/verify.py。
verify.py只回寫audit.json，所有負載均在有界Docker、UID/GID1000:1000、network none執行。
Xvfb與Wine有清理trap；容器清冊、檔案擁有權、git diff --check與清單核對另留收據。
62張PNG及保存資料僅本機，沒有公開散布權。

## 現況與未完成項

GitHub依AGENTS.md例外在主機先核auth status，再唯讀取得106張Issue。
儲存庫仍private，24張OPEN／82張CLOSED；本機18未完成／89完成、14人工保持。
八張已有完成紀錄仍OPEN，建角邊界與HD兩項尚未建Issue，差異保留remote-review.json。
沒有Issue寫入、提交、推送、版號、封包或發布變更。

RE-256十九套件通過，正式來源未變，本輪沒有重跑相同Go測試。
受阻持物候選DRAFT、完整原版UI與操作、舊H、42別名、全50-byte、其他戰鬥分支仍未閉合。
HD手繪媒材已核准，解析度、介面及人物造型仍待確認。真機與正式交付保持原閘門。
本輪結果只證明現行重製來源的有限回歸，不宣稱原版完整parity或remake完成。
