# RE-275：戰鬥目標資料框辨識

日期：2026-10-06。狀態：初始倒地格資訊框有限CONFORMED，完整UI保持開放。

沿 [RE-273](20261006-ai-flank-normal.md) 的正常六人移動戰鬥與固定seed 1。
原版第339畫格／重製第342步，Moraine選MOVE，初始目標格為倒地Olear。
原版左上與左下資訊框皆為空，重製卻顯示Olear的肖像及傷勢。
根因是inspectBodyTarget按Actors第一筆相同座標取人，未按戰場可見精靈辨識。

## 原始定位與證據

輸入KOLTOWN-unpacked.EXE SHA-256：
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
正式資料庫SHA-256：`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
IDA Pro 9.4固定映像沿既有locked-v1，只開一次性副本，保留原名、bytes與xref。
位址為IDA線性位址，載入基底0x10000，DGROUP基底0x2DED0；原版GUI工具與seed方法沿RE-273。
本輪沒有新原版GUI實驗。既有正常畫格及原始記錄與父收據雜湊核對後重用。

同一份164道指令核對原始MZ檔時，8個far CALL的segment word須依檔案重定位表與IDA載入基底0x10000換算。首次稽核直接比未重定位檔案bytes而失敗，已保留此工具前提錯誤；mz-relocations.json分列檔案偏移、檔案bytes、IDA線性位址與載入bytes，全部164指令仍逐條核對，未放寬差異。

| 原始定位 | 已證實語意 | 證據 |
|---|---|---|
| sub_24131:0x24161、0x2417A、0x24196 | 新進目標層重設檢視旗標並清兩框 | 164道新IDA指令與既有匯出bytes完全相符 |
| 0x245FB..0x24600 | 視窗精靈圖號零，走空格清框 | 跳至0x246EC，原始分支保持 |
| 0x24643..0x24648 | 非零圖號>=8跳0x246E9，保留先前框 | 原始cmp ax,8及分支，不把倒地格當空格 |
| 0x24672..0x24696 | 圖號1..6取隊員；7由格上的敵人索引取人 | 原始索引與暫設word_37D2A資料流 |
| 0x24714／0x24730 | 空格清兩框 | 原始矩形及呼叫參數 |

新進倒地格因前置清框而保持空框；游標從人物移到倒地或物品格則保留先前框。
空格清框，同格站立者按SPEC-008既定可見精靈優先度選取。
資訊檢視不改行動者、亂數、正式指令、主選單或法術問部位顯示。
語意索引見battle-panel-address-index.json，新結論附加於原始定位，不更名或改正式資料庫。

## 規格、實作與正常驗證

spec-draft.md、spec-ready.md及ready-review.json保存DRAFT、審查與READY閘門。
隔離Go覆蓋只改body_panel.go、planning.go及既有body_panel_test.go。
舊碼在初始倒地格、同格站立者與物品保留三個反例失敗，候選通過。
候選正常及保存驗證後才接正式，SPEC-018補充有限CONFORMED。

| 驗證 | 結果 |
|---|---|
| 正式測試 | 十九套件通過，五套件無測試 |
| 正式與候選 | Linux逐位元組相同，342來源／340Go；三份來源變更 |
| 雙語正常移動戰鬥 | 每版1270狀態、76PNG；完整trace與RE-273相同 |
| 視覺變更 | 每版只改兩張第342步資料框；英文2275像素，繁中沿放大後畫布 |
| 原版資料框 | 兩框20495像素零差，剩餘20像素均在圓環 |
| REST、正常保存與重載 | 每版1934狀態、28PNG保持，完整人物v3只差saved_at |
| 全輪 | 候選與正式共6408狀態、16程序、208PNG |

正式Linux SHA-256：`e7e2ccc0217184d74a4c3eadda05d3ce259c54ea46fe2181d12b362181cbd21f`。
現行來源是本輪current-inputs.json。保存格式仍為kol-save-v3，怪物v2、資料及裝備不改。

## 畫面限制與下一步

原35組英文頁的非滑鼠差異由3579降為1304像素。沒有放寬既有滑鼠矩形。
第339畫格已解決資料框，仍有20個圓環像素；其他八組差異保持。
小型差異位於目標圓環，部分是相位候選，尚未逐相位證實。不得加遮罩或稱整場相同。
第631畫格有1016像素位於結果圖上下的資料框邊緣，仍需追清框與貼圖順序。
原版目標格保留分支由IDA證實；本輪正常GUI只驗初始倒地格，不宣稱全部後續分支GUI完成。
AI成功側翼、撤退、撿武器及原始50-byte仍開放。
[RE-274](20261006-missile-current-regression.md)長任務與平台屬前版Linux03509ac1，本輪新版未重跑。
HD、正式封包、真機與既有使用者決策維持原閘門。

## 本機收據與重生

工作區：`workplace/target-inspection-20261006/`。
ida-export.py以唯讀來源的臨時副本匯出164道指令；第一次未明示UTF-8中止，失敗log保留，修正後同映像乾淨重跑。
prepare-prototype.py、build-prototype.sh、run.py／run.sh及compare.py建立候選、舊碼反例並比較同一路徑與原版畫格。
regress.py／regress.sh核正常REST及真正保存、新程序重載。
ready.py審查後更新READY規格，promote.py才寫三份來源，build-current.sh跑正式全量測試與建置。
update-docs.py回填目前狀態、既有工作項及規格。seal.py封存，verify.py核對來源、附件、PNG完整CRC／解碼、索引及擁有權。
所有工作在固定Docker、UID/GID1000、network none、有界資源及外層逾時下執行；Xvfb由trap清理。
圖片、原版執行資料及人物存檔只留本機。沒有Issue、提交、推送、版號、公開發行或HD正式素材變更。
