# RE-233：人物資料頁肖像殘留與錯留隊長

日期：2026-10-05。狀態：SPEC-030的RE-233肖像層有限CONFORMED。

已證實：原版資料頁會把目前查看角色的肖像及名牌畫到左側第一格，只留該名角色。
舊正式Inn與InnCJK先畫全隊，再蓋資料頁抬頭，導致第六人與框邊殘留；其他發言者又錯留隊長。
本輪只改繪圖層，資料頁數值、玩家操作、資料v2與人物保存v3保持。

## 原版證據

- KOLTOWN.EXE SHA-256：`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
- dosgolem：`968ca774705853276ae55c8d1163c369481ac71f`；正常啟動觀察器`7cdb353304421eab815e7fde5a52b656a0cefc20388615fe28daadfef058e3fd`。
- 原始遊戲唯讀，KOL.EXE→TITLE.EXE→KOLBOOK.EXE→KOLTOWN.EXE正常招募六名預設角色，逐人選發言者、開MIRROR、正常返回。seed1在旅店招呼後首次送新鍵前設定DS:1600／1602一次。
- 預先鎖定70正常鍵／71狀態，六次開頁為44、49、54、59、64、69；speaker_offset依序0、768、1536、2304、3072、3840。完整六人768-byte記錄在第40至70格全部保持。
- 原版本輪沒有住宿寫檔；保存驗收重用RE-232原版已封存六次實際CHARDATA與真正重載的契約，再於新來源重跑真正v3保存／重載。

## 修正與驗收

| 項目 | 結果 |
|---|---|
| 正式來源 | 前後338輸入、336份Go；只改internal/ui/town.go |
| 規格流程 | 原版六名入口→DRAFT原型→證據審查READY→正式修正→有限CONFORMED |
| 正式行為 | sheetVisibleParty在資料頁取目前Speaker一名，供既有隊員格繪圖；空隊伍與預設發言者沿既有契約 |
| 原型 | 雙語正式／唯讀觀察304狀態，完整trace與舊版相同，只改49、54、59、64、69、74六張圖 |
| 原版畫面 | 31張完整英文頁除既定滑鼠(304,185,14,14)外零差，六張舊錯圖都被拒絕 |
| 物品 | 6324組完整16-byte槽相符，完整原版人物記錄保持 |
| 繁中 | 六名肖像與名牌由前版開頁前的正常選角畫格搬移比較零差；抬頭字模與下方數值保持 |
| 正式Linux | dbc465a3386734a37815b4639dd1783f80f17421b439558ce1c66daf18f5836b，與已驗候選逐位元組相同 |
| 正式回歸 | 雙語兩人送禮、六人保存與新程序重載、城鎮／戰鬥共684狀態；全量internal／cmd測試通過 |
| 保存 | 新v3完整payload只排saved_at後與前版相同；新程序重載不改保存檔 |
| 平台 | Windows AMD64、macOS雙架構／universal重建與結構通過；最低11.0與鎖定相依保持，真機未驗 |

物品與隊伍立繪換頁沿各頁原有繪圖；送禮確認的全隊肖像保持。
原版第六名殘留的2959像素差異已由新完整頁驗收閉合，RE-232的FAIL收據保持歷史。
首隔離提案只取隊長，逐人原版畫面立即推翻該解釋；原稿與建置保留initial-proposal，不進正式路徑。

本輪不宣稱法術／任務子頁、坐騎與戶外／野外所有邊界完成。送禮Giving候選清單與原版版面差異仍開放。
RE-231整條任務7477狀態與RE-230 Wine執行證據仍屬舊0aca3427，未冒稱新dbc465a3已重跑。
HD手繪媒材已確認，畫面比例未定；未改HD素材、字型或操作版面。

## 重生與收據

本機工作區`workplace/sheet-party-header-20261005/`。prepare.py先驗RE-232收據及338來源，
plan.json預鎖輸入，run-native.sh重生原版；prepare-observer.py與build-observer.sh建立只讀觀察副本。
run-views.sh before／candidate各驗兩語言正式與觀察；prepare-candidate.py與build-candidate.sh產生隔離原型，
compare.py驗完整英文頁、原始物品與狀態，check-cjk.py驗獨立肖像搬移。
promote.py只在READY後修改一份正式Go，build-formal.sh全量測試與重建；
run-regression.sh、build-windows.sh／mac-build.sh、check-platform.py與review-formal.py補回歸和平台證據。
正式來源已變後勿直接重跑會覆寫輸出的prepare.py或promote.py；重生另開工作區並重新鎖定來源。

seal.py保存文件快照與全部附件，verify.py只讀核對SHA、PNG CRC／解壓與有限範圍。
重驗先確認現有專案路徑，再/src:ro掛載，使用eob-remake-go:1.26.7-ebiten2.9.9-audio，
UID/GID1000:1000、--rm --network none、--memory 1g、--cpus 1、--pids-limit 64及120秒外層逾時，
執行python3 -B /src/workplace/sheet-party-header-20261005/verify.py；後續來源變動用--recorded-source只核封存附件。
receipt.json為本機收據，audit.json記錄UID／Docker清理。所有新原版PNG、payload、候選與研究工具僅本機。
沒有提交、推送、Issue寫入、新版號、正式打包、發布或新增image。完整remake仍未完成。
