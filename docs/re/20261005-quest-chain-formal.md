# RE-235：最新正式程式的完整任務鏈

日期：2026-10-05。狀態：有限正常任務、完整人物保存與重載回歸通過。

正式程式保持RE-233的336 Go／338輸入，Linux SHA-256為
`dbc465a3386734a37815b4639dd1783f80f17421b439558ce1c66daf18f5836b`。
與RE-231的來源差異僅itempage.go、itempage_cjk.go、town.go三份UI檔。
沒有改正式來源、規則、素材、怪物v2、人物v3或封包。

## 已證實的有限範圍

- 從正式標題狀態機正常招募六人，沿RE-231封存7476鍵完成任務，再由新程序讀取真正存檔。
- seed1在執行前固定，途中不重設；沿既定200個遭遇控制骰100。沒有人物、座標、怪物、地點或戰局替換。
- 雙語各7477任務狀態與13重載狀態。每種語言整份trace與RE-231相同；英中只排既定exec_text、encounter_text、prompt_lines、spoil_lines後相同。
- Moraine先正常DROP騰手，第6073步／第80回合撿Ship Wheel，第7090步／第114回合逃走。第7249步六人各回報一次，第7472步實際寫出完整v3。
- 五人缺錢依正常提示略過，只有Moraine支付56住宿。保存六人，不把保存人數當成住宿人數。金冠0／0／0／2093／0／0，傷勢3／15／19／0／3／9保持。
- 完整v3只排saved_at後與前輪及另一語言相同，新程序重載不改存檔雜湊；Moraine手中Treasure Map的16-byte記錄保持`ca48a19d64503430000a000000000000`。
- 32張前後完整PNG只兩張重載資料頁肖像區改變，正文、數值與其他圖相同。已核准RE-233把目前Moraine移至第一格並移除第六名殘影。差分限六個原有肖像格與投影，英文明牌從y1至53不含端點，繁中從y3至149不含端點，依正式layout與drawPartySlots公式計算。
- 沿用前輪真實失敗負對照：未取得物品卻逃跑、已取物但戰敗均拒絕為任務完成，沒有偽造任務旗標。

這是目前正式程式的受控遭遇任務回歸，不是新原版全任務或整場戰鬥對拍。
RE-162局部原版交付證據沿原收據保持，沒有新dosgolem實驗。
自然全旅程、正常結局、全CJK／其他UI、Windows及macOS真機、新版封包都未新增驗收。
HD手繪媒材已確認，比例及送禮選人版面仍待共同決策；本輪不改其DRAFT狀態。

## 遠端狀態核對

先在主機執行gh auth status核實憑證，再唯讀擷取GitHub Issue清單及#6／#44全文。
共106張，24張仍開放。本機18未完成／89完成不等於遠端開關已同步。
#2／#4／#23／#24／#25／#28／#33／#36在本機已有完成紀錄，遠端仍開放；
剩餘戰鬥#6與#44開放且沒有留言。本輪不修改、留言、關閉或同步任何Issue。
remote-issues.json、remote-issue6.json、remote-issue44.json與remote-review.json保存此次快照。

## 重生與收據入口

本機工作區`workplace/quest-chain-formal-20261005/`：

- prepare.py核對RE-231全部封存附件、RE-233正式來源、執行檔與既定按鍵；只換驗收執行檔路徑。
- run.sh／run.py以正常標題流程各重播7476鍵，保存後啟動新程序重載。current-en／current-zh-Hant保存完整trace、真正v3、process.json與16張抽樣圖。
- compare.py核完整狀態、保存與32張前後PNG；current-review.json與visual-review.json保存有限通過事實，不將圖片差異當成資料差異。
- 初次檢查誤將英文y52作排他邊界，漏掉名牌底下4像素投影。完整狀態與保存已通過，comparison-initial-failure.json記錄畫面邊界失敗。初次腳本未留快照，compare-initial.py依失敗條件重建，註記於compare-initial-reconstruction-note.json，只作錯誤定位，不冒稱原始腳本收據。沿正式PartySlot、Shadow與CJKInn公式修正為六格矩形，英文下緣53、繁中149；沒有改正式碼、畫格或放寬任意遮罩。visual-geometry.json與visual-boundary-investigation.json保存依據。
- review-remote.py核遠端快照與本機差異；update-docs.py回填目前真相與原工作項，不結案其他閘門。
- seal.py保存文件快照、來源與附件SHA；verify.py只讀驗SHA、PNG CRC／解壓、現行338來源與有限範圍；receipt.json與audit.json同區。

工具沿既有eob-remake-go:1.26.7-ebiten2.9.9-audio，映像ID
`sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa`。
先確認現有工作樹與工作區路徑，/src:ro及/out:rw掛載，UID/GID1000:1000，
--rm --network none、3g記憶體、4CPU、256程序限額、1860秒外層逾時，
容器內執行bash /out/run.sh en或zh-Hant；Xvfb有退出清理。
封存後以1g記憶體、1CPU、128程序限額、120秒外層逾時、整個工作樹只讀掛/src，
執行python3 -B /src/workplace/quest-chain-formal-20261005/verify.py。
後續來源改變時加--recorded-source只核封存附件，不能據此稱新來源已跑過。
所有新輸出與原始payload只留本機；沒有提交、推送、發行、定版號或新增映像。
來源未改，依比例原則沿用RE-233全量Go與三平台建置證據，不重跑無關測試。
Docker容器清理、檔案UID與暫存狀態見audit.json。
