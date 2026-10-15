# 完成比較後的證據保存與清理

使用者於 2026-10-07 要求比較完成後刪除 `workplace/` 的對應工作目錄。
先保存必要證據、核對 SHA-256 並修正入口，再刪除該次已完成的工作區。
其他任務的原始輸入、有效研究工作區及交付包不在此清理範圍。

| 比較或驗證 | 現行入口 | 本機壓縮證據 |
|---|---|---|
| 完整物品／隊伍、裝備原型及正式接入、68圖示 | [正式裝備驗收](hd-equipment-runtime-20261015.md)與[圖示驗收](hd-actions-runtime-20261015.md) | [保存清冊](local/hd-equipment-pages-20261015/retention.json)；完整模型與原錨點只留本機 |
| 全畫面HD盤點、原版七段事件與第0段手繪候選 | [盤點及事件比較](hd-completion-20261014.md) | [保存清冊](local/hd-completion-20261014/retention.json)，107檔與兩份可重建執行檔雜湊；候選待確認 |
| HD跨平台一色階修正、首次失敗與取樣原型 | [交付與精度修正](release-v1.0.2-20261013.md) | [保存清冊](local/hd-pixel-consistency-20261013/retention.json)，1667檔與12份可重建執行檔雜湊；首次候選保留 |
| HD操作介面、人物頭像與同主題對比候選 | [可操作原型](hd-ui-theme-20261012/review.html) | 小型原型直接保存於同目錄，[來源及逐檔雜湊](hd-ui-theme-20261012/manifest.json)；已核准接入 |
| HD介面、旅店頭像及正常切換存讀檔 | [正式驗收](hd-interface-runtime-20261012.md) | [保存清冊](local/hd-interface-runtime-20261012/retention.json) |
| 原創室內樂標題與旅店試聽 | [試聽頁](hd-music-20261012/review.html) | MIDI、OGG、音符及[雜湊清冊](hd-music-20261012/manifest.json)直接保存；人耳待確認 |
| 正式遊戲內HD／EGA切換與住宿存讀檔 | [正式驗收](display-switch-runtime-20261012.md) | [保存清冊](local/display-switch-runtime-20261012/retention.json) |
| 遊戲內畫面設定與切換候選 | [候選驗證](display-switch-20261011.md) | [保存清冊](local/display-switch-20261011/retention.json) |
| 首批HD正式接入、EGA回歸、點選及存讀檔 | [首批驗收](hd-runtime-20261011.md) | [保存清冊](local/hd-runtime-20261011/retention.json) |
| 受控場景、中文結局及HD第一批原型 | [目前驗收](controlled-scenes-20261011.md)與[HD原型](hd-first-slice-20261011/review.html) | [保存清冊](local/controlled-scenes-20261011/retention.json) |
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
| 第三批KOLBOOK九個C與36個函式庫函式 | [第三批紀錄](../re/matching-decompilation.md#第三批kolbook來源還原紀錄) | [保存清冊](local/matching-c3-20261008/retention.json) |
| 第四批TANDYPIN完整來源及三支函式庫替換 | [第四批紀錄](../re/matching-decompilation.md#第四批六支程式來源基準紀錄) | [保存清冊](local/matching-c4-20261008/retention.json) |
| 第五批KOL完整來源還原 | [第五批紀錄](../re/matching-decompilation.md#第五批kol完整來源還原紀錄) | [保存清冊](local/matching-c5-20261008/retention.json) |
| 第六批三支runtime及四個Maker C | [第六批來源](../re/matching-decompilation.md#第六批來源還原紀錄) | [保存清冊](local/matching-c6-20261008/retention.json) |
| 第七批Maker九個C／ASM函式 | [第七批來源](../re/matching-decompilation.md#第七批來源還原紀錄) | [保存清冊](local/matching-c7-20261009/retention.json) |
| 第八批MAKEMODE完整CODE／DATA來源 | [第八批來源](../re/matching-decompilation.md#第八批來源還原紀錄) | [保存清冊](local/matching-c8-20261009/retention.json) |
| 第九批TITLE完整來源及未命名CODE | [第九批來源](../re/matching-decompilation.md#第九批來源還原紀錄) | [保存清冊](local/matching-c9-20261009/retention.json) |
| 第十批KOLBOOK新增75個原始來源函式 | [第十批來源](../re/matching-decompilation.md#第十批來源還原紀錄) | [保存清冊](local/matching-c10-20261009/retention.json) |
| 第十一批KOLBOOK新增18個原始來源函式 | [第十一批來源](../re/matching-decompilation.md#第十一批來源還原紀錄) | [保存清冊](local/matching-c11-20261009/retention.json) |
| 第十二批KOLBOOK主函式及全部初始化DATA | [第十二批來源](../re/matching-decompilation.md#第十二批來源還原紀錄) | [保存清冊](local/matching-c12-20261009/retention.json) |
| 第十三批角色建立與411個額外CODE bytes | [第十三批來源](../re/matching-decompilation.md#第十三批來源還原紀錄) | [保存清冊](local/matching-c13-20261009/retention.json) |
| 第十四批KOLTOWN完整LIB模組與33原始函式 | [第十四批來源](../re/matching-decompilation.md#第十四批來源還原紀錄) | [保存清冊](local/matching-c14-20261009/retention.json) |
| 第十五批KOLTOWN新增30個純C原始函式 | [第十五批來源](../re/matching-decompilation.md#第十五批來源還原紀錄) | [保存清冊](local/matching-c15-20261009/retention.json) |
| 第十六批KOLTOWN新增25個純C原始函式 | [第十六批來源](../re/matching-decompilation.md#第十六批來源還原紀錄) | [保存清冊](local/matching-c16-20261009/retention.json) |
| 第十七批KOLTOWN新增12純C與31明列C／ASM | [第十七批來源](../re/matching-decompilation.md#第十七批來源還原紀錄) | [保存清冊](local/matching-c17-20261009/retention.json) |
| 第十八批KOLTOWN全部初始化DATA來源 | [第十八批來源](../re/matching-decompilation.md#第十八批來源還原紀錄) | [保存清冊](local/matching-c18-20261009/retention.json) |
| 第十九批KOLTOWN新增24個CODE函式 | [第十九批來源](../re/matching-decompilation.md#第十九批來源還原紀錄) | [保存清冊](local/matching-c19-20261009/retention.json) |
| 第二十批KOLTOWN新增16個CODE函式 | [第二十批來源](../re/matching-decompilation.md#第二十批來源還原紀錄) | [保存清冊](local/matching-c20-20261010/retention.json) |
| 第二十一批KOLTOWN新增14個純C函式 | [第二十一批來源](../re/matching-decompilation.md#第二十一批來源還原紀錄) | [保存清冊](local/matching-c21-20261010/retention.json) |
| 第二十二批KOLTOWN新增14個來源函式 | [第二十二批來源](../re/matching-decompilation.md#第二十二批來源還原紀錄) | [保存清冊](local/matching-c22-20261010/retention.json) |
| 第二十三批KOLTOWN主函式、選單與輸入 | [第二十三批來源](../re/matching-decompilation.md#第二十三批來源還原紀錄) | [保存清冊](local/matching-c23-20261010/retention.json) |
| 第二十四批KOLTOWN完整城鎮互動控制函式 | [第二十四批來源](../re/matching-decompilation.md#第二十四批來源還原紀錄) | [保存清冊](local/matching-c24-20261010/retention.json) |
| 第二十五批KOLTOWN完整戰鬥主控制與多CODE物件 | [第二十五批來源](../re/matching-decompilation.md#第二十五批來源還原紀錄) | [保存清冊](local/matching-c25-20261010/retention.json) |
| 第二十六批KOLTOWN圖像／文字／磁碟四函式 | [第二十六批來源](../re/matching-decompilation.md#第二十六批來源還原紀錄) | [保存清冊](local/matching-c26-20261010/retention.json) |
| 第二十七批KOLTOWN選項／物品／圖像四純C函式 | [第二十七批來源](../re/matching-decompilation.md#第二十七批來源還原紀錄) | [保存清冊](local/matching-c27-20261010/retention.json) |
| 第二十八批KOLTOWN名冊／保存／資料頁四函式 | [第二十八批來源](../re/matching-decompilation.md#第二十八批來源還原紀錄) | [保存清冊](local/matching-c28-20261010/retention.json) |
| 第二十九批KOLTOWN角色／選單／欄位編輯四函式 | [第二十九批來源](../re/matching-decompilation.md#第二十九批來源還原紀錄) | [保存清冊](local/matching-c29-20261010/retention.json) |
| 第三十批KOLTOWN欄位／圖像／戰鬥四函式 | [第三十批來源](../re/matching-decompilation.md#第三十批來源還原紀錄) | [保存清冊](local/matching-c30-20261010/retention.json) |
| 第三十一批KOLTOWN戰鬥顯示／選擇／輸入四函式 | [第三十一批來源](../re/matching-decompilation.md#第三十一批來源還原紀錄) | [保存清冊](local/matching-c31-20261010/retention.json) |
| 第三十二批KOLTOWN目標／權重／攻擊／傷害四純C函式 | [第三十二批來源](../re/matching-decompilation.md#第三十二批來源還原紀錄) | [保存清冊](local/matching-c32-20261010/retention.json) |
| 第三十三批KOLTOWN效果／掉物／世界／遭遇四純C函式 | [第三十三批來源](../re/matching-decompilation.md#第三十三批來源還原紀錄) | [保存清冊](local/matching-c33-20261010/retention.json) |
| 第三十四批KOLTOWN圖像／人物選擇／事件／分頁四函式 | [目前來源進度](../re/matching-decompilation.md#持續c來源還原) | [保存清冊](local/matching-c34-20261010/retention.json) |

2026-10-11 的[中文與功能有限核查](cjk-closeout-20261011.md)另保存於
[本機清冊](local/cjk-closeout-20261011/retention.json)。包含通過、缺輸入、逾時及舊里程碑差異，
核對名稱為 `cjk-closeout-20261011`，同用下方工具的 `verify`。

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
  matching-c2-20261008 matching-c3-20261008 matching-c4-20261008 matching-c5-20261008 matching-c6-20261008 matching-c7-20261009 matching-c8-20261009 matching-c9-20261009 matching-c10-20261009 matching-c11-20261009 matching-c12-20261009 matching-c13-20261009 matching-c14-20261009 matching-c15-20261009 matching-c16-20261009 matching-c17-20261009 matching-c18-20261009 matching-c19-20261009 matching-c20-20261010 matching-c21-20261010 matching-c22-20261010 matching-c23-20261010 matching-c24-20261010 matching-c25-20261010 matching-c26-20261010 matching-c27-20261010 matching-c28-20261010 matching-c29-20261010 matching-c30-20261010 matching-c31-20261010 matching-c32-20261010 matching-c33-20261010 matching-c34-20261010
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

第四批保存1288份檔案，28份可重建執行檔只保留hash；77436265 bytes保存為8459328 bytes。
TANDYPIN完整來源與KOL／TITLE／MAKEMODE新增101個函式庫函式，16項乾淨產物及九項反例通過。
原始六EXE、四支新程式全檔雜湊、新舊十二份來源證明與342份正式輸入經全唯讀封存核對後，
才清理`workplace/matching-c4-20261008/`。清冊明列五支部分來源、559函式及176286 CODE／13788 DATA bytes未還原。

第五批保存1438份檔案，20份可重建執行檔只保留hash；91189586 bytes保存為9237881 bytes。
KOL全部76函式、CODE與DATA來源、兩次完整EXE、八項乾淨產物及八項反例均通過，
全唯讀封存再次重算兩份完整來源與正式342份輸入後，才清理`workplace/matching-c5-20261008/`。
KOL與TANDYPIN完整來源，其餘四支仍尚518函式、169979 CODE bytes及12553 DATA bytes未還原。

第六批保存1706份檔案，29份可重建執行檔只保留hash；115356323 bytes保存為9779916 bytes。
三支完整runtime來源、四個Maker純C、22項乾淨產物、八項反例及六份來源核對通過，
全唯讀封存再次重算六份source proof與正式342份程式輸入後，才清理`workplace/matching-c6-20261008/`。
完整目標仍有四支部分來源、388函式、153311 CODE bytes及12553 DATA bytes未還原。

2026-10-09第七批保存2365份檔案，23份可重建執行檔只保留hash；
133308658 bytes保存為10777487 bytes。Maker新增一個純C與八個明列C／ASM函式，
32項乾淨產物、十項反例與兩份完整來源核對通過；六原始輸入、所有附件及342份Go輸入
經全唯讀封存重算後，清理範圍限`workplace/matching-c7-20261009/`。
目前379函式、151554 CODE及12553 DATA bytes未還原；完整matching目標維持未完成。

同日第八批保存3679份檔案，26份可重建執行檔只保留hash；
165621187 bytes保存為12273981 bytes。MAKEMODE全部78函式、12816 CODE及2353初始化DATA bytes
有明列來源，原始CODE／DATA片段均為0。36項乾淨產物、十三項反例與兩份完整來源核對通過，
六原始輸入、全部附件與342份Go輸入經全唯讀封存重算後，清理範圍限`workplace/matching-c8-20261009/`。
剩餘三支部分來源、378函式、148146 CODE及10200 DATA bytes；完整目標保持active。
DS+4E2的0x1F00初值用途未知，原始C宣告為假說，限制隨來源清冊保持。

第九批保存5918份檔案，26份可重建執行檔只保留hash；
195783496 bytes保存為14613501 bytes。TITLE全部115個原始函式、另列118個C CODE bytes、
16880 CODE與1279初始化DATA bytes均有來源，原始CODE／DATA片段為0。
77項乾淨產物、十五項反例、原始IDA DB關係及兩份完整來源經全唯讀封存重算後，
清理範圍限`workplace/matching-c9-20261009/`。原始799函式不變，兩個初值用途未知。
KOLBOOK與KOLTOWN仍有345函式、140738 CODE及8921 DATA bytes未還原，完整目標保持active。

第十批保存8893份檔案，23份可重建執行檔只保留hash；
221612068 bytes保存為18252743 bytes。KOLBOOK新增75個完整原始函式來源8485 bytes，
二十個新純C與55個混合C／ASM分列；原有九個C及共享OBJ保持。
159項乾淨產物、九項反例、兩份獨立來源及342份Go輸入經全唯讀封存重算後，
清理範圍限`workplace/matching-c10-20261009/`。四支完整來源保持，
Book25函式與2185 DATA bytes仍未完成；全六支剩270函式、132253 CODE及8921 DATA bytes，goal維持active。

同日第十一批保存14488份檔案，25份可重建EXE只保留hash；255367655 bytes保存為20822756 bytes。
Book新增18個完整原始函式6012 bytes，195項乾淨產物、九反例與兩份來源證明通過。
六原始輸入、所有附件及342份正式輸入經全唯讀封存重算退出0後，才清理本批工作區。
四個未匹配候選與DATA來源候選不計完成；全六支仍有252函式、126241 CODE及8921 DATA bytes，完整目標保持active。

同日第十二批保存20049份檔案，25份可重建EXE只留hash；297887791 bytes保存為23667775 bytes。
Book完整主函式4789 bytes與16-byte尾碼、全部2185初始化DATA來源均已核對。
201項乾淨產物、十五反例、兩份獨立來源及六原始輸入、342份Go輸入經全唯讀重驗退出0後，才清理本批工作區。
DATA完成與CODE完成分開核對；Book六函式及8085 CODE bytes仍待還原，全六支仍251函式、121390 CODE與6736 DATA bytes。

同日第十三批保存25224份檔案，24份可重建EXE只留hash；341055815 bytes保存為27165226 bytes。
角色建立3671 bytes及三個額外區段411 bytes有完整來源，209項乾淨產物、十七反例與兩份來源證明通過。
六原始輸入、全部附件與342份正式輸入經全唯讀重驗退出0後，才精確清理本批工作區。
原始191／799函式數保持；Book五函式4003 CODE bytes、全六支250函式117308 CODE及6736 DATA bytes仍待還原，完整目標active。

同日第十四批保存25680份檔案，23份可重建執行檔只留hash；382660051 bytes保存為28480804 bytes。
Town52個完整LIB模組覆蓋75原始函式，新增33；全部51個C保持，十項乾淨產物及十一反例通過。
六原始輸入、兩份獨立來源、2281重定位及完整packed雜湊、342正式輸入經全唯讀重驗退出0後，才精確清理本批工作區。
Book五／Town212仍217函式、111877 CODE及6736 DATA bytes未還原，完整目標active。

同日第十五批保存25949份檔案，35份可重建執行檔只留hash；418724543 bytes保存為28225759 bytes。
Town新增30個純C原始函式3203 bytes，現81純C8260 bytes；原有75個LIB原始函式與52模組保持。
十項乾淨產物、十三反例、六原始輸入、兩份完整來源、2281重定位與342正式輸入，
均已從壓縮證據全唯讀重驗通過。清理範圍限`workplace/matching-c15-20261009/`，約釋放372 MiB。
Book五／Town182共187函式、108674 CODE及6736 DATA bytes未還原，完整目標保持進行中。

同日第十六批保存26206份檔案，23份可重建執行檔只留hash；451089376 bytes保存為29925298 bytes。
Town新增25個純C原始函式5595 bytes，現106純C13855 bytes；81個既有C與75個LIB原始函式保持。
十項乾淨產物、十四反例、六原始輸入、兩份完整來源、2281重定位與342正式輸入，
均已從壓縮證據全唯讀重驗通過。清理範圍限`workplace/matching-c16-20261009/`，約釋放402 MiB。
Book五／Town157共162函式、103079 CODE及6736 DATA bytes未還原，完整目標保持進行中。

同日第十七批保存26561份檔案，31份可重建執行檔只留hash；512598542 bytes保存為31852025 bytes。
Town新增12純C1905及31明列C／ASM2265 bytes，共43個原始函式4170 bytes。
31混合函式分列465個C框架與1800個ASM bytes，來源分類器拒絕錯誤分類、範圍與普通C插入ASM。
十項乾淨產物、21反例、六原始輸入、兩份完整來源、2281重定位及342正式輸入，
均從壓縮證據全唯讀重驗通過。清理範圍限`workplace/matching-c17-20261009/`，約釋放458 MiB。
Book五／Town114共119函式、98909 CODE及6736 DATA bytes未還原，完整目標保持進行中。

同日第十八批保存27053份檔案，27份可重建執行檔只留hash；610853612 bytes保存為35088749 bytes。
Town全部6736初始化DATA已具來源：5600 compiler初值、991 LIB實際初值、88 LIB宣告零與57原生零bytes。
246個C符號、31個LIB範圍、原始DATA片段0、15項乾淨產物、31反例與兩份完整來源均已核對。
六原始輸入、338函式／271項DATA／1359筆IDA直接引用、2281重定位與342正式輸入，
均從壓縮證據全唯讀重驗通過。清理範圍限`workplace/matching-c18-20261009/`，約釋放549 MiB。
Book五／Town114仍119函式、98909 CODE bytes未還原；未還原DATA為0，完整目標保持進行中。

2026-10-10第十九批保存27719份檔案，23份可重建執行檔只留hash；741855232 bytes保存為37318330 bytes。
Town新增24個原始CODE函式4337 bytes，包含3純C466與21明列混合3871 bytes。
全部149個既有compiler函式與6736 DATA來源保持；15項乾淨產物、35反例與兩份完整來源通過。
六原始輸入、173 compiler函式、2281重定位、九個DATA relocation與342正式輸入，
均從壓縮證據全唯讀重驗通過。清理範圍限`workplace/matching-c19-20261009/`，約釋放672 MiB。
Book五／Town90仍95函式、94572 CODE bytes未還原；DATA來源保持，完整目標仍進行中。


2026-10-10 第二十批保存 29104 份檔案，23 份可重建執行檔只留 hash；1029842224 bytes 保存為 39736863 bytes。
Town 新增 16 個原始 CODE 函式，共 3823 bytes，包含 1 純 C 276 與 15 混合 C／ASM 3547 bytes。
全部 173 個既有 compiler 函式與 6736 DATA 來源保持；15 項乾淨產物、39 反例及兩份完整來源核對通過。
六原始輸入、189 compiler 函式、2281 重定位、九個 DATA relocation 與 342 正式輸入，
均從壓縮證據全唯讀重驗通過。清理範圍限 `workplace/matching-c20-20261010/`，約釋放 944 MiB。
Book 五／Town 74 仍有 79 函式、90749 CODE bytes 未還原；DATA 來源保持，完整目標繼續進行。


2026-10-10 第二十一批保存 29814 份檔案，23 份可重建執行檔只留 hash；1164323031 bytes 保存為 41950914 bytes。
Town 新增 14 個純 C 原始函式，共 5268 bytes，沒有新增 ASM。全部 189 個既有 compiler 函式與完整 DATA 鏈保持。
15 項乾淨產物、43 反例、兩份獨立來源核對、六原始輸入、203 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c21-20261010/`，約釋放 1070 MiB。
Book 五／Town 60 仍有 65 函式、85481 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十二批保存 30519 份檔案，23 份可重建執行檔只留 hash；1315556443 bytes 保存為 44247523 bytes。
Town 新增 14 個原始來源函式，共 6851 bytes，包含 12 純 C 5944 與 2 混合 907 bytes，混合分列 493 C／414 ASM。
全部 203 個既有 compiler 函式與完整 DATA 鏈保持；15 項乾淨產物、47 反例、兩份獨立來源核對、六原始輸入、217 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c22-20261010/`，約釋放 1212 MiB。
Book 五／Town 46 仍有 51 函式、78630 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十三批保存 31276 份檔案，23 份可重建執行檔只留 hash；1486813342 bytes 保存為 46609859 bytes。
Town 主函式、選單與輸入新增三個完整來源函式，共 6095 bytes，包含 1 純 C 1355 與 2 混合 4740 bytes，混合分列 4533 C／207 ASM。
全部 217 個既有 compiler 函式與完整 DATA 鏈保持；15 項乾淨產物、51 反例、兩份獨立來源核對、六原始輸入、220 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c23-20261010/`，約釋放 1373 MiB。
Book 五／Town 43 仍有 48 函式、72535 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十四批保存 32070 份檔案，23 份可重建執行檔只留 hash；1680414770 bytes 保存為 49143160 bytes。
Town 完整城鎮互動控制 `sub_11A93` 新增 8967 bytes，分列 8939 C／28 ASM；原始 2874 條指令及全部 220 個既有 compiler 函式保持。
15 項乾淨產物、55 反例、兩份獨立來源核對、六原始輸入、221 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c24-20261010/`，約釋放 1555 MiB。
Book 五／Town 42 仍有 47 函式、63568 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十五批保存 33149 份檔案，23 份可重建執行檔只留 hash；1964222568 bytes 保存為 53385800 bytes。
Town 完整戰鬥控制 `sub_1E3A8` 新增 13650 bytes，分列 13634 C／16 ASM；原始 4389 條指令及全部 221 個既有 compiler 函式保持。
實際兩個 CODE 編譯物件分別為 53366 與 13650 bytes；合併 67016-byte 視圖不作 compiler 輸入，保留每個物件的來源與 PUBLIC 位置。
18 項乾淨產物、62 反例、兩份獨立來源核對、六原始輸入、222 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c25-20261010/`，約釋放 1806 MiB。
壓縮證據 SHA-256 為 `47789f97f6ed157c15f3580828b39f1e9db0cad50a9941400b9cee87eae397b5`。
Book 五／Town 41 仍有 46 函式、49918 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十六批保存 34481 份檔案，23 份可重建執行檔只留 hash；2279245607 bytes 保存為 75755036 bytes。
Town 圖像轉換、矩形搬移、文字 codec 與磁碟重試新增四個完整來源函式，共 1707 bytes，分列 1012 C／695 ASM。
全部 222 個既有 compiler 函式、兩個既有 CODE 來源／物件及完整 DATA 保持；新增第三個 CONTROL 物件 1707 bytes。
20 項乾淨產物、68 反例、兩份獨立來源核對、六原始輸入、226 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。只刪除 `workplace/matching-c26-20261010/`，約釋放 2084 MiB。
壓縮證據 SHA-256 為 `490f9cdbf70c1552925a88327ded087d666760f32cf1c7cd10da003d05090478`。
Book 五／Town 37 仍有 42 函式、48211 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十七批保存 37324 份檔案，23 份可重建執行檔只留 hash；2929551494 bytes 保存為 123774040 bytes。
Town 選項重繪、文字選擇、物品命令與圖像載入新增四個完整純 C 函式，共 3679 bytes，1236 條原始指令匹配，沒有新增 ASM。
全部 226 個既有 compiler 函式、三個既有 CODE 來源／物件及完整 DATA 保持；新增第四個 MENU 物件 3679 bytes。
22 項乾淨產物、74 反例、兩份獨立來源核對、六原始輸入、230 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。原反例腳本失敗、69 個案例及欄位取值修正也保留；只刪除 `workplace/matching-c27-20261010/`，約釋放 2657 MiB。
壓縮證據 SHA-256 為 `5cc4873442c4f9bc2f6230a2630d03cd9f7d1899fca5cee9ec2d672ccbd524ce`。
Book 五／Town 33 仍有 38 函式、44532 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十八批保存 40676 份檔案，23 份可重建執行檔只留 hash；3655820291 bytes 保存為 179009348 bytes。
Town 名冊、保存、物品明細與角色資料頁新增四個完整來源函式，共 5221 bytes，分列 5033 C／188 ASM，1756 條原始指令匹配。
兩個純 C 2654 與兩個混合 2567 bytes；13 個 DOS／ES 範圍預先明列。全部 230 個既有 compiler 函式、四個既有 CODE 來源／物件及完整 DATA 保持。
24 項乾淨產物、80 反例、兩份獨立來源核對、六原始輸入、234 compiler 函式、2281 重定位及 342 正式輸入，
均從壓縮證據全唯讀重驗通過。長度反例的 801／800 由真實物件另核；原失敗腳本與 73 個案例保留。只刪除 `workplace/matching-c28-20261010/`，約釋放 3295 MiB。
壓縮證據 SHA-256 為 `15cff0a2a9687006e79c983d13c2b1b44d3ac385ba34472410ef06fe4ad58de3`。
Book 五／Town 29 仍有 34 函式、39311 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第二十九批保存 42812 份檔案，23 份可重建執行檔只留 hash；4094764232 bytes 保存為 213044472 bytes。
Town 角色摘要、物品選單、購買與欄位編輯新增四個完整來源函式，共 6038 bytes，分列 5911 C／127 ASM，2023 條原始指令匹配。
兩個純 C 3966 與兩個混合 2072 bytes；四個 DS／ES 範圍預先明列。全部 234 個既有 compiler 函式、五個既有 CODE 來源／物件及完整 DATA 保持。
26 項乾淨產物、86 反例、兩份獨立來源核對、六原始輸入、238 compiler 函式、2281 重定位及 342 正式輸入，均從壓縮證據全唯讀重驗通過。
前批 801／800 長度反例與兩批失敗紀錄保持；只刪除 `workplace/matching-c29-20261010/`，約釋放 3680 MiB。
壓縮證據 SHA-256 為 `4edce41204ecc7777c284934e248fde87bf1b9e13677d33155523c5b997260a3`。
Book 五／Town 25 仍有 30 函式、33273 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第三十批保存 45279 份檔案，23 份可重建執行檔只留 hash；4581218518 bytes 保存為 251167288 bytes。
Town 欄位、圖像、格點與戰鬥命令新增四個完整來源函式，共 5554 bytes，分列 5539 C／15 ASM，1864 條原始指令匹配。
三個純 C 4530 與一個混合 1024 bytes；唯一 ES／SI 範圍預先明列。全部 238 個既有 compiler 函式、六個既有 CODE 來源／物件與完整 DATA 保持。
28 項乾淨產物、92 反例、兩份獨立來源核對、六原始輸入、242 compiler 函式、2281 重定位與 342 正式輸入，均從壓縮證據全唯讀重驗通過。
前批 801／800 長度反例與兩批失敗紀錄保持；只刪除 `workplace/matching-c30-20261010/`，約釋放 4107 MiB。
壓縮證據 SHA-256 為 `5efa472e33b8f116deaa750d836ff183c1369fe9745029f3f68fd4fe3f48153a`。
Book 五／Town 21 仍有 26 函式、27719 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。


2026-10-10 第三十一批保存 48117 份檔案，23 份可重建執行檔只留 hash；5120704523 bytes 保存為 293534172 bytes。
Town 戰鬥顯示、狀態數值、戰鬥選擇與目標輸入新增四個完整函式，共 6558 bytes，分列 6553 C／5 ASM，2182 條原始指令匹配。
三個純 C 5707 與一個混合 851 bytes；唯一 DS 範圍預先明列。全部 242 個既有 compiler 函式、七個既有 CODE 來源／物件與完整 DATA 保持。
30 項乾淨產物、98 反例、兩份獨立來源核對、六原始輸入、246 compiler 函式、2281 重定位與 342 Go／module 輸入均從壓縮證據全唯讀重驗通過。
前批 801／800 長度反例與兩批失敗紀錄保持；只刪除 `workplace/matching-c31-20261010/`，約釋放 4579 MiB。
壓縮證據 SHA-256 為 `a08959c8db2c45e551eee28038473b0bed673a1fd0382d64890cb2844bde779a`。
Book 五／Town 17 仍有 22 函式、21161 CODE bytes 未還原；DATA 來源保持完整，目標繼續進行。

第三十二批另保存 51296 份檔案，23 份可重建 EXE 只留 hash。
5713959497 bytes 壓縮為 340051632 bytes，SHA-256 為 `c058068d8fc9e24b89411a3b4f39802ce6fddc9114bccd58c6a504395dc7e0c9`。
四個純 C 函式 5611 bytes、九個真實 CODE 物件、完整 DATA、32 項乾淨產物與 104 反例，
已由壓縮檔全唯讀獨立核對。新入口核對後只刪除本批工作區，約釋放 5099 MiB。
剩餘 18 函式／15550 CODE bytes，未還原 DATA 為 0，完整目標仍開放。

第三十三批另保存 54871 份檔案，23 份可重建 EXE 只留 hash。
6368282319 bytes 壓縮為 391374076 bytes，SHA-256 為 `04efa2a5a4b0382229d72e95b03a9d5f697a3e3f5c27e6c9d06e6b4ca624f76a`。
四個純 C 函式 6348 bytes、十個真實 CODE 物件、完整 DATA、34 項乾淨產物與 110 反例，
已由壓縮檔全唯讀獨立核對，另重算部分 CRT 與 41 個函式外未還原 bytes。
新入口核對後只刪除本批工作區，約釋放 5673 MiB。剩餘 14 函式／9202 CODE bytes，DATA 未還原為 0，完整目標仍開放。

第三十四批另保存 59003 份檔案，26 份可重建或按固定 Git blob 重取的執行檔只留 hash。
7077133575 bytes 壓縮為 446953976 bytes，SHA-256 為 `b89bcf54287983fae8b5580eb11919307d6f38e806a565eea72d2fc63063852c`。
四函式 4203 bytes、十一個真實 CODE 物件、完整 DATA、36 項乾淨產物與 116 反例，
已從壓縮檔全唯讀獨立核對；12 個未匹配試驗與八份替代工具／資源的 pins 另核。
新入口核對後只刪除本批工作區，約釋放 6293 MiB。剩餘 10 函式／4999 CODE bytes，DATA 未還原為 0，完整目標仍開放。

## HD配樂正式接入

[正式驗收](hd-music-runtime-20261013.md)與[本機保存清冊](local/hd-music-runtime-20261013/retention.json)保存2026-10-13的四首室內樂、獨立F9與實際音訊收據。
本輪保存82份檔案；可重建遊戲程式只留SHA-256，Go快取及已核對的重複音樂中間檔已移除。
壓縮檔23568120 bytes，SHA-256 `60e73d18327825b625ca6cd7c146f05a35b0658fed39f7b21c52158be0240149`。原始WAV、正常住宿v3、七狀態重載、準備世界輸入與日誌僅本機保留。
核對使用`tools/retain_verification.py verify hd-music-runtime-20261013`；在上述Docker條件下將專案唯讀掛`/src`、`docs/verification/local/`唯讀掛`/out`。還原用同工具`extract`及不存在的絕對`--destination`，輸出另給可寫掛載。
全部封存雜湊重驗後，只刪除`workplace/hd-music-runtime-20261013/`；其他工作區保持。

## 七段結局HD正式接入

[正式驗收](hd-ending-runtime-20261014.md)及[本機保存清冊](local/hd-ending-runtime-20261014/retention.json)
保存七段中文、14張EGA保持、正常1046狀態、13狀態重載及48個F10設定事件。
223份檔案已封存並逐項核對，三份可重建執行檔只留雜湊；原始比較工作區已清理。
核對使用既有Docker工具鏈執行 `tools/retain_verification.py verify hd-ending-runtime-20261014`；
掛載與解開方法沿本頁通用入口。初次200秒正常HD流程逾時保留，同輸入600秒重跑通過。

## 書本與操作子頁HD原型

[原型驗收](hd-native-ui-20261014.md)及[本機保存清冊](local/hd-native-ui-20261014/retention.json)
保存十五組每側145狀態、38張比較圖、30個F10事件與鍵鼠收據。
2190份檔案壓縮為349451720 bytes，八份可重建程式只留雜湊。
壓縮SHA-256為`c85df11dd1eed2e6e82734a28636fbd09cf7f4d3b976b16bec952fbec9ca20f6`。
在本頁既有Docker工具鏈與掛載下，執行`tools/retain_verification.py verify hd-native-ui-20261014`核對；
`delete hd-native-ui-20261014`另核原工作區逐檔雜湊後才精確刪除。
第五輪一般畫面與第六輪兩個中文問句分開保存。正式接入另見[本批收據](hd-native-ui-runtime-20261014.md)。

## 書本與操作子頁HD正式接入

[正式驗收](hd-native-ui-runtime-20261014.md)及[本機清冊](local/hd-native-ui-runtime-20261014/retention.json)
保存374份正式來源、十五組每側145狀態、30個F10事件、正常住宿與重載，以及邊框／腳本失敗輸出。
1195份檔案壓縮為267176440 bytes，三份可重建程式只留雜湊。
壓縮SHA-256為`08498b066fcc4599af9902bb4448c456f23537c31003af4343c9ca13786bfced`。
核對沿本頁Docker掛載執行`tools/retain_verification.py verify hd-native-ui-runtime-20261014`；
`delete hd-native-ui-runtime-20261014`再核原工作區後只清理該批，其他工作保持。

## 法術卷軸HD原型

[原型驗收](hd-spell-scroll-20261014.md)與[本機清冊](local/hd-spell-scroll-20261014/retention.json)
保存四組每側107狀態、十八個F10事件、生成前後及工具失敗紀錄。
395份檔案壓縮為86121424 bytes，三份可重建程式只留雜湊。
壓縮SHA-256為`dc3186b9d3ff46873b87ba20c3dc9960f4d18e740f2a8222bb5fbf85e85011b7`。
核對沿本頁Docker掛載執行`tools/retain_verification.py verify hd-spell-scroll-20261014`；
`delete hd-spell-scroll-20261014`再核原工作區後精確清理，正式接入由[後續收據](hd-spell-scroll-runtime-20261014.md)追蹤。

## 法術卷軸HD正式接入

[正式驗收](hd-spell-scroll-runtime-20261014.md)與[本機清冊](local/hd-spell-scroll-runtime-20261014/retention.json)
保存376份來源、四組107狀態、十八個F10事件、正常住宿與重載，以及完整型別頁面和存檔。
527份檔案壓縮為84581860 bytes，三份可重建程式只留雜湊。
壓縮SHA-256為`960e88c7fc0beb73b1bae8392b401100e4b7eda201499264e7a08206e55e0cd5`。
核對沿本頁Docker掛載執行`tools/retain_verification.py verify hd-spell-scroll-runtime-20261014`；
`delete hd-spell-scroll-runtime-20261014`核對原工作區後精確清理。完整型別資料只留本機，Git收據只保存雜湊與檢查結果。

## 八十張角色模板HD

[正式驗收](hd-character-templates-20261014.md)與[本機清冊](local/hd-character-templates-20261014/retention.json)
保存二十組生成與修正版、80張像素核對、380份來源、正式52張畫格、最長姓名、24個F10事件及新角色保存重載。
1325份檔案已保留，七份可重建程式只留雜湊；壓縮檔460399808 bytes，SHA-256為`bb4874024c7cd14c97338e077cb55c8747f37f0a8745720d393e1846d84dd534`。原工作區逐檔核對後已清理。
完整建角JSON、v3存檔與原始失敗輸出僅本機；Git只保存PNG、來源及精簡結果。
核對沿本頁既有Docker掛載執行`tools/retain_verification.py verify hd-character-templates-20261014`；
`delete hd-character-templates-20261014`逐檔核對後只清理該工作區。
原型fixture符號連結已另記於`fixture-links.json`，封存前只移除連結本身；來源未刪除。
不同姓名牌修正階段的來源雜湊與替代單檔索引保存在`source-input-map.json`。

## 完整物品與隊伍HD正式接入

[正式裝備驗收](hd-equipment-runtime-20261015.md)、[圖示驗收](hd-actions-runtime-20261015.md)與
[本機保存清冊](local/hd-equipment-pages-20261015/retention.json)保存本批所有原型、修正版、來源階段、生成提示、失敗輸出與正式驗證。
正式391份Go／module、185份HD資產；38張GPU畫格、正常106狀態、24個F10及v3保存重載均保留各自範圍。
本批保存7125份檔案，28份可重建執行檔只留雜湊；原4591404728 bytes壓縮為2863918172 bytes。
壓縮SHA-256為`db7c87878ef2a15e71835a1ab4e221a20e173cd0d7f627636ed812a3542cd123`，已獨立唯讀核對全部7125份內容與入口副本。
原模型、錨點、v3及完整GUI狀態只留本機，Git只保存經核對的PNG與精簡結果。
核對沿本頁Docker掛載執行`tools/retain_verification.py verify hd-equipment-pages-20261015`；
`delete hd-equipment-pages-20261015`再次逐檔核對後，只清理該工作區。其他任務保持。
