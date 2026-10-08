# RE-256：怪物名稱末字與英文複數

開工日期：2026-10-05。完成本切片：2026-10-06。狀態：英文名稱末字有限CONFORMED。

RE-254正常seed2第432格戰敗圖實際顯示CYCLOPES；同狀態候選第435步為CYCLOPSES。
前輪摘要把原圖名稱寫成CYCLOPS，原始圖片與收據保持，本輪另核末字處理。
問題限定英文名稱拼接，不按一般英文文法推測原版，也不改怪物識別值、物品、亂數或保存。

固定來源KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro9.4，linear load base0x10000、DGROUP0x2DED0；正式.i64唯讀，新副本匯出。
工作區`workplace/monster-plural-ending-20261005/`，入口ida-export.py；
固定映像ida-pro-9.4-idapython:locked-v1、UID/GID1000:1000、network none及有界資源。
新證據通過原始bytes核對後才草擬規格與隔離原型，未達READY不接正式。

## 原始來源與隔離契約（DRAFT）

199條重定位bytes與固定EXE相同，最小IDA探針338函式與輸入雜湊通過。
已證實0x2120C與0x21524先退word_345E0；sub_1C2DA的0x1C2F0至0x1C2F9
以該游標覆寫DS:1B60輸出緩衝，再遞增游標。勝利寫ES，戰敗寫E後接記錄0x178。
遭遇0x2A8FD退word_2DF14，寫ES覆寫末S。
原文CYCLOPS變CYCLOPES。單隻不退欄，非S結尾多隻只補S。
型別資料的怪物原名母體全大寫，目前只有CYCLOPS末S；
假想MUCK THINGS等測例只作字串邊界，不當正常原版遭遇證據。

候選將名稱底字與既有後綴分開拼接，英文勝利代換快照、戰敗句與遭遇句共用。
傳入Monster.Name與Actor.Name保持原值，繁中沿原譯文，count與其他規則不變。
空名稱與count不大於1沿既有guard保持，不稱原版無效輸入已驗。
以正常標題六人招募、seed2、預定435鍵，比較前版與候選雙語各436完整狀態，
只准英文第201步encounter_text及第434步exec_text改為已證實名稱，PNG變動限訊息框；
原版正常第432格與候選第434步完整訊息區需零差，繁中全狀態與整圖保持。
通過後才審查READY；正式接線還需正常seed1保存與新程序重載、全量測試。

首次IDA匯出在ASCII環境寫入中文註記失敗，零byte JSON與腳本／log保留。
改用明確UTF-8後同映像、同資料庫來源乾淨重跑。正式.i64與EXE未改。
本輪自2026-10-05跨日續核至2026-10-06，工作區與來源檔名沿開工日期保持。

## RE-256 READY審查（2026-10-06）

199條原始bytes、1744正常狀態與32完整PNG、兩個舊版名稱反例、戰敗訊息區零差通過。
正常seed1保存1136狀態與26實際v3新程序重載、完整保存只差saved_at通過。
契約READY：英文count>1且末S時只在顯示快照去末S，再沿原後綴；單隻、非S、繁中保持。
正式規則、Monster.Name、Actor.Name、亂數、輸入與人物保存v3不變。
勝利S結尾正常原版GUI及遭遇原版完整畫格尚未取得，不外推完整UI或所有名稱分支。
來源、定位、正常路徑與工具限制見[RE-256](20261005-monster-plural-ending.md)。

## 正式結果與限制

READY之後只改victory.go、既有victory_test.go與main.go，338來源／336Go保持。
正式Linux與已驗候選逐byte相同，完整SHA為`420e0e4744c8151b16c3631b61b47f56ce095455c89114e4452ed637fd8853d4`。
十九套件全部通過。正式雙語872正常／26實際重載狀態與22組PNG，
逐檔與候選控制相同。候選seed1的1136保存狀態／26重載與前正式c8764592完整相同，
保存只排saved_at保持；正式新程序再載入同一實際v3檔，沒有改檔。
seed2前版／候選1744正常狀態，英文只改201步遭遇名與434步戰敗名，
繁中完整狀態及八張PNG保持。英文兩張圖各改42／262像素，限各自名稱框。
戰敗訊息區262差異降至0；不稱原版全戰鬥終態或完整UI相同。

原版沒有另捕獲遭遇句整頁，勝利S結尾正常GUI也未走到。
字串邊界與199條來源bytes支持三出口的名稱處理，不能代替其他分支正常原版畫格。
現行受阻forcedDrop、舊H、方向確認、末端隊伍、滑鼠及42別名仍沿既有閘門。
RE-254受阻候選保持DRAFT，沒有併入本次正式程式。
RE-255長任務與Win／Mac／Wine屬c8764592；新420e0e47未借舊收據稱已重跑。
HD手繪媒材保持，比例、介面、真機、封包與發行保持既有待確認／待驗項。

首次比較器漏列encounter_text並用RE-254受阻候選435步當現行結果頁。
現行正式的結果頁為434步，435已回世界。差分只在兩個名稱欄位；
保留首稿與24張不足範圍圖片，同一程式／種子／435鍵重跑，補抓201與434。
單次文件patch引用已不存在的行而拒絕，未變動正式程式。

重生入口：ida-export.py、prepare.py、build.sh、run.py／run.sh、review.py、
run-save.py、promote.py、build-formal.sh、run-formal.py／run-formal.sh、update-docs.py。
seal.py封存正式／前版來源、原版bytes、基線、PNG、失敗輸出與文件快照；
verify.py重算附件及外部輸入雜湊、PNG尺寸／CRC、擁有權與反向勘誤。
重驗使用固定Go映像、UID/GID1000:1000、network none、1GiB／1CPU／64程序及120秒，
專案掛/src唯讀、工作區掛/out唯讀，另將本工作區掛/audit可寫，
執行`KOL_RE256_OUTPUT=/audit python3 /src/workplace/monster-plural-ending-20261005/verify.py`。
原版衍生PNG與payload僅本機，不公開、不加入封包。

## RE-258 遭遇名稱列勘誤（2026-10-06，有限CONFORMED）

RE-256當輪的原版遭遇畫格限制已由RE-258解除。正常原版198個狀態與既有前綴相同，
第197格CYCLOPES三行可見；0x2A8BE另加7及0x2A8C3回欄8的bytes核實。
英文名稱列y24，訊息矩形328差異降0；完整頁只差既有右下滑鼠67像素，未遮罩。
候選正常872、實際保存1136及26重載，正式872／26與22PNG通過，十九套件通過。
原版完整畫格缺項在本輪解除；完整UI／滑鼠與勝利S結尾正常GUI仍待驗。
資料契約、繁中、任務、規則、亂數、輸入與其他未驗分支保持，受阻原型仍DRAFT。
來源與原始定位見[RE-258](20261006-ambush-name-frame.md)。
