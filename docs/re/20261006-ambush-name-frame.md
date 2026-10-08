# RE-258：正常遭遇提示的原版整頁

日期：2026-10-06。狀態：英文伏擊名稱列有限CONFORMED。

RE-256已證實三個英文名稱出口先覆寫末S，但正常遭遇提示的原版整頁尚缺。
本輪沿RE-254固定dosgolem968ca774、原版輸入與seed2，正常啟動、招募六人及旅行。
第197格出現兩隻怪物提示，在確認前補抓完整PNG並停止，不重跑整場或改正式程式。
重複前198個完整狀態與既有原版前綴需一致，只允許第197格image欄新增檔名。
原版197與現行正式201的世界位置、次格、時間及怪物數需可比；
先核對英文完整頁與訊息框，再揭露差分，不把局部文字相同當完整UI或整場parity。

工作區`workplace/ambush-name-frame-20261006/`，入口prepare.py、build-original.sh、run-original.py。
固定Go映像eob-remake-go:1.26.7-ebiten2.9.9-audio，UID/GID1000:1000、network none、
原版與來源唯讀，只有工作區可寫；12GiB／2CPU／128程序與有界逾時。
原版兩個seed word只在旅店招呼後設定一次；不改角色、物品、座標、入口、等待或骰式。
原版輸入、工具版本、SHA與位址來源沿RE-254與RE-256收據，原始輸出僅本機。

## RE-258 READY審查（2026-10-06）

原版198狀態、138事件、23亂數與前綴相同，199條原始bytes及E7／E8解碼記錄核實。
英文候選正常872狀態只改201步分列，實際保存1136狀態只改英文171／457遭遇分列；
繁中全部保持，26真實新程序重載與完整v3保持。
遭遇訊息框328像素差異降0，整頁只差既有右下滑鼠67像素，未遮罩。
契約READY：只修英文伏擊三行及第三列y24，任務、繁中、規則、亂數、輸入與保存格式保持。
既定嚴格原版忠實度目標內的局部修正，沒有新的產品取捨；受阻DRAFT、完整UI與其他分支保持。

## 原版與版面證據

dosgolem固定968ca774705853276ae55c8d1163c369481ac71f，正式收據由它自行重生。
原版程序152.372秒完成198個正常狀態，六人招募、旅行與seed2沿RE-254計畫。
198個完整row、138事件、23亂數與既有前綴相同，只新增197格image檔名。
23亂數是此正常前綴的觀察數，沒有要求跨遊戲完整骰序。
世界32／14、次格3／0、時間1680、時段3、六人及兩敵人與正式201步可比。
原版197格seed為1239767523；原版兩個seed word在旅店招呼後設定一次，沒有重播種。
新原版完整PNG為17ba4a368464299989199ae2d359ba78149f148dc28fe78acd6e61cb9f14ec88。

原版畫格明確顯示CYCLOPES，修正前正式兩行、原版三行，訊息框仍差328像素。
來源KOLTOWN-unpacked.EXE SHA為01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2。
IDA Pro9.4，linear load base0x10000、DGROUP0x2DED0，資料庫SHA為
5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8。
重新核對RE-256匯出的199道重定位bytes，0x2A8BE的原始operand
add word_2DF16,7與0x2A8C3的mov word_2DF14,8已證實：等級後明確另起名稱列。
KOLDATA.0實際archive count為2，payload1按既有LZW抽取器解碼，完整body SHA與typed bank相同。
E7原文~V0B~H08YOU HAVE BEEN AMBUSHED@BY]、E8原文, LEVEL]核實。
第一、二列y11／17；來源另加7使第三列y24，x64。此處不作依字串長度自動換行。

原版四個region1至4缺檔與int21h AH0E AL38五次未支援訊號沿既有基準保留。
只驗本輪已重生的EGA玩家前綴，不稱dosgolem支援所有遊戲功能。
正常輸入檔、固定tar及工具hash保存在original-inputs.json、oracle-inputs.json與original/run-inputs.json。
所有研究輸入唯讀；本輪沒有DOS保存寫入，原版scratch另留本機。

## 正式結果

READY後只改main.go的英文伏擊分列及internal/ui/battle.go的英文第三列y位置。
338輸入、336Go保持，正式Linux與候選逐byte相同，完整SHA為`1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274`。
十九套件全部通過，五個無測試套件另列，不把它們當成通過的產品測例。
候選雙語正常872狀態只差英文201步encounter_text的分列與一張PNG；繁中保持。
候選seed1真正保存1136狀態只差英文171／457步遭遇分列，完整v3只排saved_at保持。
26真實新程序重載保持；正式再跑872正常／26重載及22張PNG與候選相同，保存檔未改。
英文遭遇整個60／8／194／32訊息矩形由328像素差異降0。
完整原版／候選頁只差右下滑鼠67像素，座標304至312／184至195，沒有遮罩。
繁中、任務遭遇、名稱語意、規則、seed消耗、輸入、monster v2與人物v3保持。

RE-256「尚無原版遭遇完整畫格」的當輪限制由本輪解除；英文遭遇名稱列有限CONFORMED。
勝利S結尾正常原版GUI、完整UI／滑鼠、舊H、42別名、全50-byte、末端隊伍與其他戰鬥仍開放。
受阻持物候選仍DRAFT，沒有併入正式；HD媒材保持，解析度與操作仍待選。
RE-257長任務與Win／Mac／Wine屬420e0e47，本輪新來源不借前版收據稱已重跑。
Windows／macOS真機、正式封包、版號與發布保持原閘門。

## 工具問題與重驗入口

初次文件patch索引標題不符被拒，沒有建立半成品。
比較器先把未發生的messages／drops null與篩選舊資料的[]直接比較，保留拒絕稿後
只對這兩個空採集欄允許null與空清單；198原始row、events與rolls不省略任何欄。
後續檢視對null求len失敗也保留記錄。第二個零差假設被原版三行反例否定，完整反例保留。
來源記錄驗證先誤把壓縮archive當CR明文，第二稿又誤用payloads參數型別。
讀既有extract_ui_text.py／probe_town_data.py契約後，按volume mapping與payload1解碼通過。
上述均為工具或驗收輸入問題，沒有改原版與正式規則，也沒有重跑已通過的原版程序。

入口prepare.py、build-original.sh、run-original.py、compare.py、prepare-candidate.py、
build-candidate.sh、run-candidate.py／run-candidate.sh、review-candidate.py、promote.py、
build-formal.sh、run-formal.py、finalize.py及seal.py／verify.py。
review.html列未遮罩的原版、前版與現行候選原圖，供回查名稱與換行。
封存含before／candidate／formal來源、READY快照、失敗工具、原版、PNG及外部輸入hash。
唯讀重驗使用固定Go映像，專案掛/src、工作區掛/out均唯讀，工作區另掛/audit可寫，
設定KOL_RE258_OUTPUT=/audit，執行python3 /out/verify.py。UID/GID1000:1000、network none、
2GiB／2CPU／64程序與120秒；只回寫audit.json。
全部原版PNG與payload僅本機，不公開、不加入發行包；本輪沒有Issue寫入、提交或推送。
