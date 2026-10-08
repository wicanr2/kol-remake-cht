# RE-250：戰鬥人物名稱的大字型

日期：2026-10-05。狀態：姓名字型與大小寫有限CONFORMED；完整UI與操作仍開放。

原版正常規劃頁197格的Poontang保留大小寫，正式200步轉成POONTANG。
兩側疲勞條的511個像素完全相同。首個目視判讀誤認直條未畫，像素量測已否定，
不把這項工具判讀寫成產品缺陷。

來源KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
沿既有IDA Pro 9.4匯出panels-private.json，先用MZ relocation核對原始指令bytes。
位址空間為IDA linear，載入基底0x10000，DGROUP基底0x2DED0。

- 已證實：sub_22EFC:0x23098設word_2DF18=1，0x230A8／0x230AE設欄3、行7，
  0x230C0讀角色名稱。裝備分支0x23107另切小字，0x23101／0x23243控制轉大寫。
- 已知缺口：BattleSheet把名稱與裝備共用小字型及ToUpper。原型分開名稱字型與裝備字型。
- 既有SPEC-033要求玩家姓名照輸入顯示；繁中名稱保留現有字模、座標及安全矩形，移除名稱ToUpper。
  不翻譯玩家姓名，不改原名、資料、規則、輸入或人物v3保存。

場景鏈為正常標題、六人招募、自然遭遇、人物／裝備資訊框、下一回合清框、戰敗、保存與新程序重載。
候選與唯讀觀察副本須保持完整狀態及已核亂數；英文名稱框對原版逐像素核對，
繁中只容許名稱區改變。疲勞條、傷勢、人形、指令、戰場與裝備文字保持。
READY審查後才可接正式。完整UI、滑鼠、按鍵取捨、其他分支與HD仍開放。

工作區`workplace/battle-name-font-20261005/`；入口prepare.py、build.sh、run.sh與compare.py。
原始指令不改名；新語意另列分級與來源。原版資料及衍生圖只留本機，不新增公開交付。

## 正式驗收

283道既有IDA指令經9216-byte MZ header與0x1000 relocation核對，全部相符。
原始bytes、原名與原始分級保持，七筆附加語意保存在evidence-review.json。
大字使用RE-086／087已核的book-large，沒有新字型或素材重生。
六名隊員首個姓名框原有107至250像素差，修正後全部0；213個正常戰鬥左上119×73框全部相同。
五組正常執行1930狀態，其中候選／唯讀1158；各語言完整trace與RE-249自己的基線相同，
39AI／13執行／13規劃的全部唯讀事件亦相同。英文386張、繁中25張的前後比較，
所有新差異限制在姓名區(24,7)112×8乘各語言倍率，疲勞、裝備、指令與戰場保持。
正常Olear前172鍵，再SPACE切裝備、SPACE復原的四組700狀態與12圖通過。
裝備列及彈藥文字保持，兩次切換後完整PNG復原。這是remake正常切換回歸，
不稱取得新的原版裝備模式畫格；原版契約沿SPEC-019／022及0x23101／0x23107。

首次比較腳本把兩種語言的遭遇／執行譯文要求byte相同，於110步報錯。
此為驗證腳本問題，保留compare-cross-language-rejected.py與分類記錄，
改成每語言自己的完整基線後重核同一批收據。沒有重跑遊戲或把譯文排成產品錯誤。

READY審查後才修改三個既有來源：battle_sheet.go、battle_render.go與battle_cjk.go。
NameFont由cmd繪圖端注入，不入typed遊戲狀態或保存；小字裝備仍走原出口。
正式336 Go／338輸入、Linux完整SHA為
`d7faa6c925f829fa9c7c6b6d20cc4dcc6522b9ad5951fa2cad4abeca5de5ea0d`，與原型逐byte相同。
全部19套件、雙語1136正常保存／26新程序重載通過。各語言完整trace保持，
完整人物v3只差saved_at；六人持久欄與34物品槽保持。68張保存／重載圖的新差異也只在姓名區。

完整213頁未套遮罩仍有67至1471像素差，0頁整圖相同。
其中66頁只差右下滑鼠游標；其他147頁仍含左下指令、訊息或其他區域差異。
因此不宣稱完整UI。原版按鍵、住宿末端隊伍、42射後別名及全50-byte仍開放。
長任務、Win／Mac、真機與正式封包未重跑，不把RE-246舊27212080收據當成本輪驗收。
HD只核准手繪媒材，解析度及版面仍待選。沒有提交／推送、Issue寫入、版號或公開方式變更。

工作區另有run-mode.sh、promote.py、build-formal.sh、run-save.sh、review-regression.py、
update-docs.py、seal.py／verify.py。before-source與formal-source分別保存三檔前後來源。
初始原型只在隔離副本使用before-source重跑，不覆寫現行工作樹。

## 回填索引

工作區resolution-backlinks.json與update-docs.py使用DOS、固定EXE雜湊、IDA linear原始位址作不可變鍵。
0x23098、0x230C0與0x23107回指SPEC-019的「RE-250：人物姓名大字與原始大小寫有限CONFORMED」。
缺勘誤、錯模組雜湊與缺文件三個反例均拒絕；既有RE-249與新RE-250入口正對照通過。
SPEC-033只回填既有姓名照輸入顯示契約，未擴張必譯集合或名字接受集。

後續指令區主列重建與標籤字型已有限修正，見[RE-251](20261005-command-panel-lifecycle.md)。本文姓名證據保持，完整UI差異以新收據為準。
