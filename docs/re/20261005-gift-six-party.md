# RE-232：六人送禮確認肖像、保存與重載

日期：2026-10-05。狀態：SPEC-031第17.2節有限CONFORMED。

已證實正常選Kenneth後的確認等待漏繪隊伍肖像。原版保留六人肖像、巨劍明細及訊息；
ENTER／SPACE才搬移完整物品。原因是旅店先畫隊伍，物品頁的底圖又覆蓋上方肖像。
按已核准的「選收件人後保留上圖明細與訊息」修正，不新增操作、美術或資料契約。

## 來源與正常路徑

- KOLTOWN.EXE：`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
- dosgolem：`968ca774705853276ae55c8d1163c369481ac71f`；沿用固定正常啟動觀察器`7cdb353304421eab815e7fde5a52b656a0cefc20388615fe28daadfef058e3fd`。
- 原版由KOL.EXE→TITLE.EXE→KOLBOOK.EXE→KOLTOWN.EXE正常招募六人，旅店招呼後首次新送鍵前只設定一次DS:1600／1602為seed1。重製從標題啟動前固定seed1；不替換人物、物品、座標或入口。
- 預先鎖定gates-RET、save-SPACE及實際保存重載，共215原版狀態。原版實際寫六次12288-byte CHARDATA；四份保存檔按雜湊原樣交給新程序。
- 只將Theodoric腰帶巨劍送Kenneth第四口袋，物品為`a3c8a1a4eef9100301904bff4f579600`。自選、隊伍外數字與DOWN均不轉交，ENTER及SPACE都正常轉交。

## 修正與驗收

| 項目 | 結果 |
|---|---|
| 前後正式輸入 | 各338，336份Go；只改internal/ui/itempage.go與itempage_cjk.go |
| 修正 | GiftConfirm時，底圖後以既有函式重畫隊伍肖像，再畫明細 |
| 舊版負對照 | 六人英文確認頁11916個非滑鼠差異點，舊漏圖被拒絕 |
| 候選驗證 | 雙語ENTER、SPACE及真正重載，正式／唯讀觀察760狀態和PNG相符 |
| 狀態保持 | 同步前版軌跡，只有第68／69張確認圖改變；完整v3僅排除saved_at後保持 |
| 物品與保存 | 10812組完整16-byte物品槽、六人原版持久欄位及真正重載相符 |
| 英文整頁 | 六人鏈41張加兩人確認2張，除既定滑鼠(304,185,14,14)外零差 |
| 正式程式 | b810e7fc60ab9b2ae8f1ae6a4481c547edbc3f17d970a6e586779d9b5e71a77e，逐位元組等於已驗候選 |
| 正式回歸 | 雙語兩人各39狀態、城鎮／戰鬥各184狀態，共446；全量internal／cmd測試通過 |
| 平台建置 | Windows amd64、macOS amd64／arm64／universal結構、最低11.0及鎖定相依通過；真機未驗 |

原版與重製保存後金冠為953、158、1721、2093、386、759，Poontang傷勢1、另五人0。
核對原版人物資料、法術、金冠、傷勢、任務、進度與所有物品；v3完整人物另核對保存與重載。
原版旅店保存會移除當前隊員，人物v3保持六人，沿既有契約，不稱整個旅店介面一致。
比對器僅沿父契約正規化省略的零值picture／movement及空法術null／空陣列，完整原始物品不放寬。

## 仍未完成

原版人物頁第44／45格與現行程式相差2959個非滑鼠像素，現行保留右側第六名肖像。
送禮候選清單在第59至62格與原版相差17170個非滑鼠像素，現行保留六人清單。
兩者記錄在additional-screen-differences.json為FAIL，不加遮罩、不稱完整UI流程通過。
本次有限修正未涵蓋其他收件人逐一轉交、滿背包、全部物品、野外、HD或完整remake。
RE-231全任務回歸屬舊0aca3427來源，本輪沒有在新b810e7fc重跑全部7477狀態。

## 重生與收據入口

工作區`workplace/gift-six-party-20261005/`僅本機。先讀prepare.py及plans.json，再依序執行
run-native.sh、build-observer.sh／run-remake.sh、reload-native.sh、
prepare-correction.py／build-correction.sh、run-candidate.sh、compare.py。
promotion.json記錄READY審查及兩檔前後雜湊，正式重建入口build-formal.sh；
補充run-regression.sh、build-windows.sh／mac-build.sh、check-platform.py與review-formal.py。
正式來源已變更後勿直接重跑會覆寫產物的prepare.py；先開新工作區並重新鎖輸入。

seal.py保存文件快照與來源、附件雜湊；verify.py只讀收據並核對PNG CRC、解壓及SHA。
重驗以已存在工作樹/src:ro掛載，使用eob-remake-go:1.26.7-ebiten2.9.9-audio，
UID/GID1000:1000、--rm --network none、--memory 1g、--cpus 1、--pids-limit 64、120秒外層逾時，
執行python3 -B /src/workplace/gift-six-party-20261005/verify.py；來源後續變動時用--recorded-source核對封存附件。
收據為同區receipt.json，清理與所有權為audit.json。所有原版payload、新PNG與程式僅本機；
未提交、推送、寫Issue、定新版號、打包或公開發布。沒有新增image，沿用既有Go與osxcross工具鏈。

初次候選定位CJK函式遇重複needle、比較器包含真實MIRROR差異、保存映射欄位差異，
以及回歸畫格補零位數錯誤皆保留原稿與失敗輸出。只修檢查器的表示或有限契約，真實UI差異保持FAIL。
一處文件腳本換行筆誤修正後同image乾淨重驗通過，未改產品驗收標準。
