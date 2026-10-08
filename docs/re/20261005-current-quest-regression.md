# RE-255：現行正式版長任務回歸

日期：2026-10-05。狀態：現行雙語長任務、保存重載、平台建置與Wine有限驗證通過。

沿[RE-246](20261005-quest-script-refresh.md)固定5971鍵、seed1及200個明示遭遇骰100，
重播目前正式Linuxc8764592。RE-246原收據屬Linux27212080，不能直接當新來源驗收。
本輪不改規則、程式、資料契約、操作或HD版面。

範圍為正常啟動、六人招募、長戰鬥、Fighting Bull第69回合取任務物品、
第109回合逃離、六人回報一次、正常住宿保存及新程序載入實際v3檔案。
執行前固定種子，不重播種，不注入人物、物品或座標。
明示遭遇控制沿既有預定plan，不稱自然完整旅程或原版整條任務parity。
不是24個任務的全部驗收，也不替代其他玩家分支、真機與封包。

正式338來源逐檔SHA-256與RE-254相符，正式Linux完整SHA為
`c87645925936137788443b6eda8b753842c0a725cbf4d7dd9139a631ffa302e5`。
研究入口`workplace/current-quest-regression-20261005/prepare.py`、`run.py`、`run.sh`。
固定Go映像`eob-remake-go:1.26.7-ebiten2.9.9-audio`，UID/GID1000:1000、network none，
來源唯讀、只有研究輸出可寫，Xvfb有就緒檢查與trap。
舊收據與這次輸出逐語言核對完整trace、完整保存、重載及最小充分PNG抽樣，
呈現改動與規則差異分開記錄，不忽略新反例。

同一338來源也重建Windows AMD64與macOS雙架構候選，沿用既有離線腳本與固定映像。
macOS使用`eob-remake-macos:1.26.7-ebiten2.9.9-audio`，Wine使用
`coab-wine-smoke:ubuntu-noble-20260826`。建置／結構與Wine正常有限流程分開記錄，
不當Windows或macOS真機驗收；候選留本機研究目錄，沒有正式封包或版號變更。

## 現行長任務結果

| 核對 | 結果 |
|---|---|
| 正常流程 | 每語言5972狀態，共11944；從標題招募六人，接任務、戰鬥、取物、逃離、回報及住宿 |
| 取物 | Fighting Bull第4781步、第69回合取Ship Wheel；第109回合逃離，第5583步回世界 |
| 回報 | 六人各一次，Treasure Map完整記錄保持 |
| 保存 | v3，城鎮7、入口3、57／8；完整保存只排saved_at與RE-246相同 |
| 新程序重載 | 每語言13狀態，共26；完整trace與RE-246相同，沒有改保存檔 |
| 舊版狀態差異 | 每語言82步，只差battle.cursor，均在stage2／prompt8下一回合問句；符合RE-248已證實修正 |
| 雙語 | 只排四種既有顯示文字，其他完整狀態相同；完整保存只排saved_at相同 |
| 圖像抽樣 | 40張完整PNG，28張與RE-246相同，12張不同；差分與原圖完整保留 |

四種顯示文字是encounter_text、exec_text、prompt_lines、spoil_lines。
沒有用整個battle物件或泛用游標遮罩隱藏其他差異。
圖片每語言六張改變：417、522、877、4780、4781、5582。
英文129至559像素、繁中268至2206像素不同。
這些是RE-248至RE-252的姓名字型、主列指令與訊息底色修正，
完整像素差分數與原圖保留，不稱40張全部相同。
取物與末端旅店繁中頁、877步前後姓名及主列圖示另作目視核對。

保存後六人金冠0／0／1721／0／386／0，傷勢3／1／26／1／1／10。
Olear與Fighting Bull各付56，其餘四人按缺錢流程略過，住宿佇列收完。
程式及338輸入保持，沿RE-252全量Go結果，不為無程式變動重跑單元測試。

## 平台建置與工具限制

同一338來源的Windows PE32+ AMD64、macOS ARM64／AMD64／universal已重建。
universal兩份thin、最低macOS11.0、系統相依與鎖定音訊依賴通過。
Linux沿既有正式執行檔逐byte身分。各平台完整雜湊見platform-build-checks.json。

Wine V1正常英文十三回合與真實重載的trace／PNG比較已通過，
收據卻把/out重載路徑轉成/src相對路徑而失敗。V1程式及輸出保留。
V2英文兩條流程通過，開始繁中前發現凍結的RE-252案例少四張預定PNG；
沒有啟動繁中遊戲，不能當產品缺陷或用少圖案例稱八圖通過。
沿同一現行Linux執行檔、同seed1與385正常鍵補八張繁中控制圖；
386完整狀態與原凍結trace相同，四張共同PNG逐檔相同。
V2保留後，V3重跑同一Wine9／DirectX條件與預定鍵序，四條流程全部通過。
雙語正常十三回合共772狀態，實際v3新程序重載共26狀態；
22組完整PNG與各自Linux控制圖相同，保存檔未改。
這些修正只改收據路徑與驗證輸入，不改程式、亂數或玩家操作。

## 遠端唯讀核對

依AGENTS.md先核主機gh auth status，使用主機gh取得目前Issue全文與儲存庫狀態。
目前仍為private，共106張worklist Issue，其中24 OPEN、82 CLOSED。
本機107項為18未完成、89完成；八個本機已有完成證據的舊Issue仍OPEN，
新建角邊界與HD工作尚無Issue，另有一張已合併的舊明細Issue CLOSED。
差異與各項驗證來源保存在remote-review.json，不由遠端歷史開關反向重開已完成功能。
沒有Issue寫入、留言、提交、推送或公開變更。

重生與核對入口：prepare.py、run.py／run.sh、review.py、prepare-platform.py、
build-windows.sh、mac-build.sh、check-platform.py、run-wine-controls.py、run-wine.py、
review-remote.py、seal.py、verify.py。原始與失敗產物都留在本機研究工作區。
seal.py封存338來源、平台執行檔、比較基線、研究附件及文件快照。
verify.py在來源與附件唯讀掛載下重算SHA-256、PNG解碼、尺寸及擁有權，
只有audit.json與verify.log寫入本輪輸出；receipt.json保存驗證範圍。
重驗入口為固定Go映像內的
`python3 /src/workplace/current-quest-regression-20261005/verify.py`，
環境`KOL_RE255_OUTPUT=/audit`；專案掛/src唯讀、本工作區掛/out唯讀，
本工作區另掛/audit可寫。UID/GID1000:1000、network none、1GiB、1CPU、
64程序與120秒上限。Wine臨時prefix清理、容器與專案擁有權檢查一併封存。
