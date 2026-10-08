# RE-212：原版法術名稱的正常輸入

日期：2026-10-03。正常名稱輸入的有限範圍CONFORMED；整體SPEC-035保持READY。
本研究使用既有reverse-engineer-retro-game-remake技能與正常玩家試玩入口。
工作區：`workplace/spell-name-original-20261003/`。
前批：[RE-211](20261003-spell-name-input.md)。

## 固定輸入與位址基準

| 輸入 | SHA-256 | 用途 |
|---|---|---|
| KOL.EXE | 4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71 | 正常啟動 |
| KOLBOOK.EXE | 2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832 | 正常招募 |
| packed KOLTOWN.EXE | c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221 | 實際執行 |
| MAGIC.0 | 2fb9389f6ccf3a79fbdcac7d702469641f0698b608739a75bc5b2ce7ca2c08b2 | 法術字典 |
| KOLTOWN-unpacked.EXE | 01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2 | IDA線性位址證據 |

正常dosgolem固定revision `968ca774705853276ae55c8d1163c369481ac71f`，
以KOL.EXE→KOLBOOK→KOLTOWN招募、旅行、購買Kelnalon及進編輯器。
正常旅店招呼後DS:1600=1、1602=0，只兩次seed寫入，後續不重設。
重製正常flow seed1；名稱規則不作亂數判定，兩側前置種子不保證跨整場骰序一致。
無direct-entry、座標／角色注入、wait stub或局部函式替代正常入口。
各original driver、keys、PNG與角色768-byte差分保存在工作區，全部由dosgolem重生。

舊驅動的報告input欄繼承unpacked標籤，實際執行packed的檔名、大小與雜湊另存
`executed-inputs.json`。不得將兩者當同一份輸入或共用檔案偏移。
來源缺region1…4.dat及既有int21h AH=0E AL=38觀察保持列明；本名稱正常路徑已完成。
驅動reached_purchase仍比對舊Strength期望，不作本次成功指標；本次直接核對實際法術與資源。

## 已證實的caller與正常行為

IDA Pro 9.4既有非破壞性匯出見[RE-209](20261003-spell-lifecycle-player-path.md)。
來源資料庫為`workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64`，既有SHA-256
5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8；本輪未修改資料庫。
caller為IDA linear 0x17C59…0x17D07，DGROUP基址2DED0。
17C6D呼叫sub_1B779填底，17C89呼叫sub_1B6DC畫框；17C91／17C97設col12／row91。
17CA3傳12、17CA7傳DS:2E3A、17CAB呼叫sub_1B5C8。
17CC3／17CCC檢查輸入首byte非零，17CDC呼叫sub_1C0F6打包，17CE4將八byte寫工作副本DS:9E28。
僅以上caller bytes與正常原版差分已證實，不宣稱被呼叫四函式完整逆向。
角色法術原始槽為768-byte角色記錄+0x200；typed名稱使用既有5-bit解碼器。
DS:2E3A有舊bytes殘留，有效輸入長度從0開始，畫面與確認後NUL為準。
第一路徑raw_9e48_hex只是唯讀鄰近緩衝，語意未知，不當名稱資料。

| 正常路徑 | 狀態 | 實際結果 |
|---|---:|---|
| native-name1 | 187 | RET／BS／N／RET，空→空→N，存錄N |
| native-name2 | 201 | Down及數字1忽略，Left退一字；存錄Abc Efghijm |
| native-name3 | 201 | 明確文字D／L；第12字K立即確認Abc Defghijk，多餘RET重開輸入，角色未存錄 |
| native-name4 | 200 | 保留上限鍵序，刪多餘RET，正常存錄Abc Defghijk |

第二路徑原腳本D／L映射Down／Left，不能稱為實際字母D／L；第三與第四改用TypeKeys文字鍵。
各路徑到編輯器的180狀態保持角色與seed相同，第三與第四前197狀態及PNG完全相同。
N與12字正常存錄只改名稱，效果6、guards[64,0,0,0]、turn0、power10保持，
金冠987→647、冒險點數2816→2476。native-name3沒有扣款或改角色，副本先確認再存錄的界線清楚。
原始unused打包尾bytes保留，不由typed名稱逆推清零或聲稱全raw重建。

## 實作與驗收

RE-212修正正常法術名稱：有效輸入從空字串開始，確認後各單字字首大寫、其餘小寫，第12字自動確認；正式Left退格，英文補原版黑框白底輸入框。四條固定seed1的正常原版共789狀態，N與Abc Defghijk正常存錄的法術欄位及金冠647／冒險點數2476，兩語言相符。20張英文115×13輸入框零像素差，框外差異未豁免，不宣稱整頁或caret閃爍相位parity。正式／觀察6548PNG及可比trace相同，雙語各360自然失效圖與前版相同、取消其餘193圖相同。改名N及Strength後正常住宿保存各661狀態，金冠601／冒險點數2486，其他持久資料保持RE-210；正式重載各16狀態、兩人34槽與14欄位一致。三個反例失敗，新碼全部internal／cmd測試通過。正式325份Go，候選56bfb4c8；本名稱切片有限CONFORMED，整體SPEC-035保持READY。原版住宿保存、其他字元／空白邊界、完整戰鬥／替代路徑與任務／競技場parity仍開放。繁中注解A／B與怪物四套資料格式待原有決定。22未完成／84完成保持。入口docs/re/20261003-spell-name-original.md。

英文輸入框x95…209、y88…100，黑框白底，book-large文字x96、y91；原名稱留在上方。
獨立字模原型先驗六框；正式英文N三框、上限輸入及關閉十七框合計20框零像素差。
每張保留全頁差異計數，不豁免框外103個既有像素；不宣稱整頁、閃爍或游標相位parity。
初版原型誤用book-large底線符號及小寫n，依原版畫素與字模修正，原版圖片始終不改。
繁中保留既有名稱欄caret／隱藏注解版面，正常N及上限畫面已抽查。
正式Left退格已接線；腳本驗證正常文字token與UI，不宣稱平台實體鍵盤真機驗收。
其他字元／多重或首尾空白、原版住宿保存、完整主要路徑尚未驗，不由本切片推廣。

RE-211前文可見名稱只證明trace改變與兩份重製PNG相同，當時英文未畫輸入文字。
已在RE-211追加勘誤，歷史與5701前批附件保持；336共用輸入在修改前封存。
正式源325Go，八份既有檔異動、一份新增回歸；Go1.26.7／Ebitengine2.9.9鎖版。
兩個舊碼起點／大小寫反例及第12字無自動確認反例失敗，新碼全部internal／cmd通過。

## 收據與重驗

[receipt.json](../verification/spell-name-original-20261003/receipt.json)與
[verify.py](../verification/spell-name-original-20261003/verify.py)。
來源／正常路徑審查工具：`workplace/spell-name-original-20261003/review.py`。
掛專案`/src:ro`、本工作區`/out:ro`，使用既有`kol-package:20260907`執行
`python3 /src/docs/verification/spell-name-original-20261003/verify.py`，重驗本輪附件、325Go、5701父附件及PNG對照。
重生原版用`run_native.sh`、`run_native2.sh`、`run_native3.sh`、`run_native4.sh`，
固定dosgolem儲存庫掛`/repo:ro`，使用`eob-remake-go:1.26.7-ebiten2.9.9-audio`；這四個已完成目錄不可原地覆寫。
正常重製六條路徑的命令、keys及trace逐份保存，兩語言用`run_normal.sh en`與`run_normal.sh zh-Hant`。
各工具只在有界非root Docker執行，預設network none，來源唯讀，新的獨立輸出目錄可寫。
全部原版EXE、driver、payload、未公開素材與本輪PNG留本機研究，沒有發布或權利擴張。
清理、443個入口異動、擁有權及暫存狀態見`workplace/spell-name-original-20261003/audit.json`。

## 2026-10-04：名稱空白邊界 DRAFT

原先「其他字元／多重或首尾空白尚未驗」的歷史範圍不再涵蓋本次兩個已驗案例。
正常輸入` A  B `與兩個空白已證實正式差異，打包字數分別6及2；
全空白仍是原版非空名稱記錄，不能以抽取器去尾空白後的空字串判為空法術槽。
完整來源、位址與目前DRAFT契約已回填[SPEC-035](../spec/035-cjk-spell-editor.md)。
兩條原版共378狀態、四條正式共788狀態；十張輸入框零差，關閉後與整頁差異保留。
正式來源、格式及前批收據維持。收據與重生入口在
`workplace/spell-name-boundaries-20261004/{review.json,run_native.sh,run_formal.py}`。
其他字元、住宿保存、重載及完整戰鬥仍未由本次證據完成。

## 2026-10-04：空白名稱表示原型驗證，正式升格未獲核准

隔離原型只改名稱的非空判斷，現有Name字串保留六字元及兩個空白，v3不新增欄位。
town／save完整測試與雙語正常付費保存、重載、MIRROR法術頁通過；
32次試跑共3552狀態及PNG，保存只變預期名稱，十四持久欄位與三十四槽重載一致。
原型與唯讀觀察版PNG及可比trace相同，756入口畫格保持；
兩張原版／原型的96×7名稱框零差。原版確認欄位及打包字數由前批bytes回查，
本輪沒有新增原版住宿／重載oracle，不宣稱完整保存parity。
自動核准審查拒絕正式READY升格，理由為現有GO未明確授權此閘門變更。
SPEC-035本邊界仍DRAFT，正式Go未改，已停止依賴升格的正式分支。
可核准提案、擬議patch與重生入口在
`workplace/spell-name-spaces-prototype-20261004/{review-bundle.json,proposed-formal.patch,receipt.json}`。

## 2026-10-04：原版空白名稱正常住宿保存及重載

證據等級：已證實，限下述兩條正常原版路徑及實際保存欄位。
原版兩條正常保存共1154狀態，實際CHARDATA正常重載共14狀態，1196張PNG已解碼。首尾空白六字與全空白兩字均保存且重載後保持，原版16-byte法術槽重載不變；名稱、護符、效果Strength、金冠601及冒險點數2486與雙語隔離原型相符。其餘欄位差異保留：sheet.nutrition=10／8；兩側旅程不同，不宣稱整場或整段骰序parity。

原版KOLTOWN.EXE SHA-256為c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221；
dosgolem固定968ca774705853276ae55c8d1163c369481ac71f，Go1.26.7。
執行記錄為DOS DGROUP：首人物DS:BB88，法術槽+0x200；CHARDATA依768-byte記錄的檔案偏移核對。
角色所在檔案槽、實際檔案雜湊與重載輸入雜湊逐例記錄於review.json。
六字名稱的打包codes為[0,1,0,0,2,0]，兩空白為[0,0]，有效長度6／2。
現有抽取器的rstrip仍保留作交叉核對，不能用其空字串把有效法術槽判空。
各例前181個狀態及PNG與RE-213來源一致。旅店招呼後一次設定seed1，
後續旅行、自然遇敵、逃離及住宿不重設；重載與正常招募沒有種子寫入。
原始輸入唯讀，住宿只寫獨立scratch；重載複製本次實際寫出的四個檔案，來源保持。
前輪「未新增原版空白名稱住宿重載oracle」仍是當時限制，本批證據只解除此兩例未知。
自動核准審查先前拒絕READY升格仍有效，正式程式及DRAFT狀態不改。
重生與核對入口為`workplace/spell-name-spaces-original-save-20261004/{run_native.sh,review.py,verify.py,receipt.json}`。
