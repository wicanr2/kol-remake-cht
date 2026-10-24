# RE-234：送禮選收件人版面原型

日期：2026-10-05。狀態：DRAFT隔離原型通過有限驗證，版面仍待使用者選擇。
正式來源與RE-233的Linux dbc465a3保持。

已證實：原版選收件人時保留上方六人肖像與物品明細，下方提示給誰；
現行畫面顯示編號及完整姓名清單，蓋掉明細。兩者都用1至6選人，之後仍需ENTER／SPACE確認。
這次只比較版面，已批准的確認時點保持。沒有把建議當成使用者選擇。

## 來源與有限驗證

- 原版KOLTOWN.EXE SHA-256：`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
- 原版沿RE-232的封存gates-RET／save-SPACE／reload，66／87／62狀態，共215；沒有新oracle實驗。
- dosgolem固定`968ca774705853276ae55c8d1163c369481ac71f`，正常啟動工具`7cdb353304421eab815e7fde5a52b656a0cefc20388615fe28daadfef058e3fd`。原版正常招募六人，seed1在旅店招呼後首次新送鍵前設定一次。重製從標題前固定seed1，不改人物、座標或入口。
- 現行338輸入全部以SHA驗證，正式Linux：`dbc465a3386734a37815b4639dd1783f80f17421b439558ce1c66daf18f5836b`。
- 原型：`7860a10b048c61077198db1b113be4d137da87bf7b405780d074cc46ac7bb4a6`。只覆蓋internal/ui/itempage.go與itempage_cjk.go，唯讀觀察main.go另覆蓋，不進正式路徑。

| 項目 | 結果 |
|---|---|
| 正常範圍 | 六人招募，Theodoric巨劍送Kenneth，無效自選／隊伍外數字、DOWN，RET與SPACE確認，真正保存與新程序重載 |
| 新舊正式／觀察 | 共1520狀態與PNG，整份typed trace相同，沒有省略人物欄位 |
| 畫面改動 | 每條正常送禮鏈只改64至67四張選人等待圖；確認與轉交圖保持，27張重載圖全部相同 |
| 原版畫面 | 53張英文完整頁除既定滑鼠(304,185,14,14)外零差，八張舊選人圖被拒絕 |
| 原始物品 | 10812個完整16-byte槽含空槽與實際重載，逐位元組相同；6360個非空槽 |
| 保存 | 完整v3只排saved_at後相同，新程序重載不改檔 |
| 正式來源 | 338來源與正式候選保持；沒有正式修改、READY升格或封包更新 |

原型Giving時畫全隊肖像，沿既有明細與訊息框繪圖；Note移到下方、英文沿原版大寫，
繁中沿原譯文。角色、物品、選人、確認與保存規則保持，沒有新增資料欄位。
滿背包、內層取消、所有收件人、全部物品、野外與滑鼠語意未新增驗收。
候選版面不自動成為正式規則；若選定，還需READY審查、正式接線與正常GUI回歸。

## 比較與重生入口

[本機比較頁](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/gift-recipient-layout-20261005/review.html)可切換繁中／英文與四個操作步驟，
並列現行完整姓名清單、原版明細版面及同狀態原版圖。點圖可查看原尺寸。
使用者選定前保持DRAFT；HD手繪媒材已定，HD比例仍是另一個待選分支。

工作區`workplace/gift-recipient-layout-20261005/`：prepare.py驗來源及前輪附件，
build.sh建立兩UI覆蓋與唯讀觀察版本，run.sh before／candidate依plans.json與reload-plan.json重播，
compare.py核整頁、完整typed trace、原始槽與保存，build-review.py核所有引用PNG並生成比較頁。
比較頁的HTML引用與實際遊戲PNG已驗，但尚未在瀏覽器實際操作頁面，不稱瀏覽器工具驗收。
原型沒有改正式Go，依比例原則未重跑來源未變的全量Go測試或三平台建置；正式證據仍屬RE-233。

比較摘要初版將10812母批數量與10608轉交槽數的差額誤寫為204個空槽差異。
這是計數誤判，已撤回；加入重載204槽後10812組原始bytes全部相同。
原錯摘要與count-correction.json保留為非驗收，不改原版、正式碼、資料或物品規則。

seal.py保存文件快照、來源與附件SHA；verify.py只讀驗SHA、PNG CRC／解壓及有限範圍。
先確認現有工作樹路徑，再/src:ro掛載，使用eob-remake-go:1.26.7-ebiten2.9.9-audio，
UID/GID1000:1000、--rm --network none、--memory 1g、--cpus 1、--pids-limit 64、120秒外層逾時，
執行python3 -B /src/workplace/gift-recipient-layout-20261005/verify.py。
來源後續改變用--recorded-source只核封存附件；receipt.json與audit.json同區。
所有新PNG、原型、來源引用與payload僅本機，沒有新增授權、提交、推送、Issue寫入、版號、發布或新image。
原有18未完成／89完成／14人工項保持，不宣稱HD或完整remake完成。
