# 完成比較後的證據保存與清理

使用者於 2026-10-07 要求比較完成後刪除 `workplace/` 的對應工作目錄。
先保存必要證據、核對 SHA-256 並修正入口，再刪除該次已完成的工作區。
其他任務的原始輸入、有效研究工作區及交付包不在此清理範圍。

| 比較或驗證 | 現行入口 | 本機壓縮證據 |
|---|---|---|
| 任務亂數原型與圓環 | [比較頁](local/quest-entry-rng-prototype-20261006/review.html) | [原型保存清冊](local/quest-entry-rng-prototype-20261006/retention.json) |
| 圓環捕捉相位 | [比較頁的相位證據](local/quest-entry-rng-prototype-20261006/review.html#markers) | [相位保存清冊](local/quest-marker-phase-20261006/retention.json) |
| 移動修正後的組合候選 | [最新比較](local/quest-entry-rng-prototype-20261006/review.html) | [保存清冊](local/move-message-visibility-20261006/retention.json) |
| 受阻打暈、回城保存及重載 | [正式驗收](blocked-stun-20261007.md) | [保存清冊](local/blocked-stun-formal-20261007/retention.json) |
| 前版跨平台抽樣 | [原收據](local/finish-platform-20261007/receipt.json) | [保存清冊](local/finish-platform-20261007/retention.json) |
| 任務亂數正式接續及正常保存 | [正式驗收](quest-rng-20261007.md) | [保存清冊](local/quest-rng-formal-20261007/retention.json) |
| 任務亂數後的三平台與Windows回歸 | [平台驗收](platform-quest-rng-20261007.md) | [保存清冊](local/platform-quest-rng-20261007/retention.json) |
| 原版C首批matching試點 | [試點與C源檔](../re/matching-decompilation.md) | [保存清冊](local/matching-pilot-20261008/retention.json) |
| 原版C增量連結與跨段呼叫 | [增量證據與C源檔](../re/matching-decompilation.md#增量連結與第四個函式) | [保存清冊](local/matching-link-20261008/retention.json) |
| 完整原版EXEPACK重建基準 | [封裝與來源覆蓋](../re/matching-decompilation.md#完整packed-exe的增量重建基準) | [保存清冊](local/matching-pack-20261008/retention.json) |
| 第一批46個C與八個library來源替換 | [第一批紀錄](../re/matching-decompilation.md#第一批c來源還原紀錄) | [保存清冊](local/matching-c-20261008/retention.json) |
| 第二批52個C與42個函式庫函式來源替換 | [第二批紀錄](../re/matching-decompilation.md#第二批c與函式庫還原紀錄) | [保存清冊](local/matching-c2-20261008/retention.json) |
| 第三批KOLBOOK九個C與36個函式庫函式 | [目前來源進度](../re/matching-decompilation.md#持續c來源還原) | [保存清冊](local/matching-c3-20261008/retention.json) |

各目錄的 `evidence.tar.xz` 保留完整狀態紀錄、圖片、保存、腳本、輸入清冊與原收據。
執行檔屬可重建產物，移除位元組並保留原 SHA-256。原收據不改寫，內部路徑仍指向
當時的工作區；`retention.json` 記錄每個舊路徑在壓縮檔中的相對位置及雜湊。
小型 JSON、腳本及測試 log 另有可直接閱讀的副本。

本機證據在 `.gitignore` 排除。含原版衍生資料，不新增公開散布授權。
完整性核對只證明保存內容未改，不能代替重新執行遊戲或擴大原版對拍範圍。

## 在 Docker 內核對

沿用[目前恢復的Go revision](platform-quest-rng-20261007.md)。專案唯讀掛 `/src`，
`docs/verification/local/` 唯讀掛 `/out`；UID/GID 1000:1000、無網路、
1 GiB、1 CPU、96 程序及 60 秒外層逾時。

```sh
python3 -B /src/tools/retain_verification.py verify \
  quest-entry-rng-prototype-20261006 quest-marker-phase-20261006 \
  move-message-visibility-20261006 blocked-stun-formal-20261007 \
  finish-platform-20261007 quest-rng-formal-20261007 platform-quest-rng-20261007 \
  matching-pilot-20261008 matching-link-20261008 matching-pack-20261008 matching-c-20261008 \
  matching-c2-20261008 matching-c3-20261008
```

需重查大檔時，工具的 `extract` 可解到容器的暫存目錄。只抽出原證據，不產生新收據。
歷史回放腳本使用原路徑，重播前需在容器內掛回對應路徑並重建執行檔。
工作完成後刪除暫存解壓目錄，不把它重新留在專案 `workplace/`。

## 本輪清理結果

2026-10-07第一批五個工作目錄已刪除。刪除前共約756 MiB；
壓縮證據與可直接閱讀副本約18.6 MiB，約釋放737.8 MiB。
原收據的全部保留附件逐檔核對SHA-256，執行檔只保留原雜湊。
目前Markdown入口沒有指向已刪目錄的連結。正式342份程式輸入保持，
該次清理時任務亂數為DRAFT；後續正式狀態見[最新驗收](quest-rng-20261007.md)。
五個工作區的完整清理數據見[本機收據](local/cleanup-20261007.json)。

同日接手後，任務亂數正式驗收工作區另保存1217份檔案與四份可重建執行檔的雜湊。
清理前236791517 bytes，保存約4801811 bytes，釋放約221 MiB。
原收據、原版引用、完整v3重載及新連結經完全唯讀掛載獨立核對通過後，才刪除該工作目錄。

2026-10-08平台回歸另保存1158份檔案與七份可重建執行檔的雜湊。
134757126 bytes縮為約6366771 bytes；兩個Wine前綴另釋放1440804232 bytes。
根收據、8508狀態、完整v3及連結經完全唯讀核對後，才清理本次工作目錄。

同日matching首批試點另保存138份檔案，約1.75 MiB；44份可重建或按固定commit重取的
執行檔只保留來源與hash。三個純C函式268 bytes及乾淨重編鏈、原始file bytes與新連結
均經完全唯讀核對後，才清理約5.71 MiB的試點工作目錄。
這批包含匹配C、原版code／DB衍生證據與候選工具，只留本機；完整EXE尚未匹配。

同日增量連結另保存395份檔案，18份可重建或按固定commit重取的執行檔只保留hash。
13107888 bytes縮為1987250 bytes；全部檔案、兩次OMF重生的完整初始化映像與2281筆
重定位均經全唯讀核對後，才刪除`workplace/matching-link-20261008/`。
四個純C函式共300 bytes已原位匹配，129012 bytes仍屬原版片段，完整MZ與EXEPACK尚未匹配。

後續完整EXEPACK研究保存557份檔案，35份可重生或固定commit重取的執行檔只保留hash。
17909645 bytes縮為2349194 bytes；封存、原版full-hash、完整OMF與342份正式輸入均經
全唯讀核對後，才刪除`workplace/matching-pack-20261008/`。生成的完整packed EXE與原版相同，
但129012 bytes仍是原版片段；匹配C、原始code與工具仍只留本機。

持續C批次另保存805份檔案，31份可重建執行檔只保留hash；37125058 bytes保存為16841622 bytes。
兩份獨立compiler／library來源核對、46C／4309 bytes、8library／372 bytes、原版packed hash、
六EXE與284個未還原函式清冊均經全唯讀核對後，才清理`workplace/matching-c-20261008/`。
完整目標仍active；下一批由此源碼接續，保留117895個未還原CODE bytes及其他五EXE。

第二批保存1014份檔案，35份可重建執行檔只保留hash；58514028 bytes保存為19148105 bytes。
三份獨立來源核對、52C／5146 bytes與42個函式庫函式／4214 bytes、兩支完整EXE雜湊及
十三項乾淨重編均經全唯讀核對。清理範圍限`workplace/matching-c2-20261008/`。
KOLTOWN仍有245函式／113305 CODE bytes、KOLBOOK仍有190函式／37351 CODE bytes，
其餘四支EXE尚未還原；完整matching目標保持進行中。

第三批保存1108份檔案，18份可重建執行檔只保留hash；62672804 bytes保存為7518680 bytes。
KOLBOOK九C／594 bytes與36個函式庫函式／3705 bytes、六項乾淨重編及六項反例均通過，
兩支程式各兩份來源核對經全唯讀封存重算後，才清理`workplace/matching-c3-20261008/`。
合計60C／5651 bytes與78個函式庫函式／7919 bytes；KOLBOOK未還原降為146函式／33141 CODE bytes。
KOLTOWN與其餘四支EXE的來源限制保持，完整matching目標仍未完成。
