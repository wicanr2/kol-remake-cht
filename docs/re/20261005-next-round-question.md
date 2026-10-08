# RE-249：下一回合問句的繪圖接線

日期：2026-10-05。狀態：十三頁問句與繪圖錯誤返回有限CONFORMED；完整UI與操作仍開放。

RE-248已修正規劃末端索引及鏡頭，十三張確認頁仍缺左上問句。
來源KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4 linear基底0x10000；沿RE-248的333道已核對指令bytes與封存原版正常收據。

- 已證實：sub_1E3A8:0x1FB04清左上框，0x1FB20清左下框，
  0x1FB2C推入0x108、0x1FB30呼叫sub_1A8FD，再畫YES／U-TURN／逃離圖示。
- 已證實：0x1FB6F等待選擇，0x1FB9B再次清左上框；問句不延續到執行段。
- 已證實：現有抽取文字庫已保存記錄264，原文
  `~V16~H03Begin next@round?`，既有繁中譯文為`~V16~H03開始下@一回合？`。
  UISegments把版面碼解析為X24、Y22／30；繁中沿同一座標乘既定3倍。
- remake缺口：目前StageField／PromptNextRound清兩框及畫圖示，沒有讀取或繪製該記錄。

第一版原型錯用小字型，十三頁各差293像素。0x1FAEA先寫word_2DF18=1，
RE-086／087已核該值選8×8大字、@行距8；原版畫格亦符合。
改用既有book-large，繁中沿現有字模與座標。
缺記錄的正常217鍵反例另揭露Draw只重複印錯誤，Update等待drawn而不結束，120秒後由工具停止。
V2補暫態drawErr，下一Update把錯誤交回RunGame；不保存該欄，也不宣稱原版壞資料行為。
這是既有失敗即關閉契約的接線修正。V1畫格、程式、比較及逾時記錄完整保留。

工作區`workplace/next-round-question-20261005/`，入口prepare.py。
只做可丟棄繪圖overlay，保持目前338來源、規則、亂數、輸入、怪物v2與人物v3。
正常385鍵、seed1固定一次，不重跑原版抽樣，不注入狀態或覆寫遭遇。
先核十三頁左上框英文像素、繁中原譯文及安全矩形；其他頁、戰場與完整狀態須保持。
完整頁面、其他問句、方向確認、住宿末端隊伍與射後別名不因此升格。
原版資料及衍生圖僅本機。READY審查後才可接正式，不改HD設計、版號或公開方式。

## 正式驗收

先封存V1的十三頁各293像素差異及缺記錄逾時，V2改用book-large。
十三頁英文左上119×73全部逐像素相同；繁中沿原譯文顯示兩行，文字安全區與字模均通過。
雙語候選及唯讀副本1544狀態，整份trace與RE-248完全相同，全部觀察事件也完全相同。
39次AI、13次執行與13次規劃入口、比較點seed及已解欄位保持。
20張每語言正常抽樣中，只有十三張問句頁改變，全部新像素限制在左上安全矩形。
其餘七張對照頁及戰場保持，不由這40張抽樣宣稱所有遊戲畫面已通過。

缺記錄、超出左上框與缺字模的三個受控資料反例，沿正常首輪217鍵到問句，
V2各在5至6秒回報對應錯誤並退出1。資料副本只改UI記錄264，正式資料保持。
壞資料是現代runtime邊界驗證，不稱原版parity或正常遊戲成功路徑。
第一次邊界腳本缺shot-dir，在參數驗證退出2；首版及紀錄另存，不算產品反例。

READY審查後才改cmd/kol/battle_render.go與main.go。338輸入、336 Go保持，
規則、亂數、輸入、怪物v2與人物v3保持。新Linux與已驗原型逐byte相同：
`599620c7ee2b6422213929dfb51ff4edd3c90c73492e569752fa7994872e1f85`。
全部19套件通過。雙語1136正常返回／保存狀態與RE-248完全相同，
完整v3只差saved_at，26個新程序重載狀態完全相同；六人持久欄及34物品槽保持。

工作區入口prepare.py／build.sh、run-remake.sh、compare.py、run-boundaries.py、
review-ready.py、promote.py、build-formal.sh、prepare-regression.py／run-save.sh、
review-regression.py、seal.py／verify.py。before-source與formal-source保存兩檔前後來源。
要重跑初始原型，在隔離副本使用before-source；不覆寫正式工作樹。
Ebitengine Draw錯誤只暫存在game，下一Update返回RunGame；不寫進保存。

尚未重跑長任務、Win／Mac、真機與封包。RE-246平台及長任務收據仍屬Linux27212080，
不稱為新599620c7來源的驗收。方向確認、住宿末端隊伍、完整頁面、射後別名與全50-byte仍開放。
HD手繪媒材已核准，比例及介面仍待選。沒有提交／推送、Issue寫入、版號、封包或公開方式變更。

## 回填索引

不可變鍵為DOS、固定輸入SHA-256、IDA Pro 9.4 linear原始位址。
工作區resolution-backlinks.json／review-backlinks.py核對0x1FAEA與0x1FB30，
並拒絕缺勘誤、缺文件及錯模組雜湊。SPEC-008勘誤標記為「RE-249：下一回合問句有限CONFORMED」。
RE-248以追加回指保留歷史；20260908-command-panel的0x1FAEA測試記憶體CB返回未覆蓋問句，
不因該局部測試推定問句已完成。RE-086／087及其他分支的共用字型旗標契約保持。
本輪未重命名原始函式、全域或運算元。原始匯出仍標unknown，另以semantic-review.json
逐筆附加已證實的有限語意、原始bytes與收據來源，兩者分開保存。

## 後續補接：RE-250戰鬥姓名

[RE-250](20261005-battle-name-font.md)另修正人物姓名字型與大小寫，213個左上框相同。
本文Linux599620c7與原有收據保持；新來源保存回歸由RE-250核對，完整UI仍開放。
