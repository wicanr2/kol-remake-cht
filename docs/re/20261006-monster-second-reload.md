# RE-269：第二名怪物的裝填來源

日期：2026-10-06。狀態：裝備池頭部記錄生命週期切片CONFORMED，完整別名與整場保持開放。

沿 [RE-267](20261006-monster-first-reload.md) 的現行 339 個輸入及
[RE-268](20261006-monster-reload-current-regression.md) 的長任務與平台回歸。
剩餘每語言 16 筆 Command 差異都在第二名怪物。

本輪核對原版 DS:D0C5 的資料來源。它等於怪物裝備池 DS:CD92 的row6記錄＋0x33。
row6是零起算第六列，也就是第七筆128-byte記錄。此位置可被第七名怪物的裝備覆寫，不能固定成零。

入口為 `workplace/monster-second-reload-20261006/ida-runner.py`，查詢在 `ida-export.py`。
使用既有 IDA Pro 9.4 固定映像，正式資料庫唯讀，對 /tmp 一次性副本查詢。
原始名稱、位址、位元組與操作數保留，語意依證據分級。

## 固定輸入與工具

| 輸入 | SHA-256／版本 |
|---|---|
| `workplace/analysis/koltown/KOLTOWN-unpacked.EXE` | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 同目錄`KOLTOWN-unpacked-ida94-derived.i64` | `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| dosgolem | `968ca774705853276ae55c8d1163c369481ac71f` |
| IDA | Pro9.4，`ida-pro-9.4-idapython:locked-v1`，UID1000，338函式 |
| Go | `eob-remake-go:1.26.7-ebiten2.9.9-audio` |

IDA位址為linear，載入基址`0x10000`；DS偏移另以DGROUP基址`0x2DED0`定位。
MZ檔頭9216bytes，2281筆重定位，以載入segment`0x1000`換算後，3,020道唯一指令bytes相符。
輸出只留本機工作區；原始EXE、DB與資料未變更，不包含公開散布授權。

## 來源結論

| 原始定位 | 已證實的範圍 |
|---|---|
| IDA`0x2B973..0x2B980` | `push ss; pop es`後，`rep stosb`清零DGROUP偏移`[0x1A42,0xD5B0)`，包含整個1536-byte裝備池 |
| IDA`0x1EB9A..0x1EC46` | 逐actor6至最後參戰者，複製32-byte標頭及所選96-byte裝備至`CD92+(actor-6)*128`；未使用列保留 |
| IDA`0x1EC29` | actor12的頭部記錄＋3寫入DS:D0C5，成為actor7射前／射後來源 |
| `sub_22296:0x222B0..0x222B5` | 依actor選擇怪物128-byte本體指標，原始`word_39A48`保留 |
| `sub_22084:0x220E9` | 耐久成負時清除所選完整16-byte物品；頭部名稱＋3亦清零。未破損只修改＋0x0B |
| `sub_269BA:0x269D4..0x26AB9` | 怪物強制掉落只處理手持＋0x20的16bytes，不清頭部＋0x30 |

對42,940道函式指令作窄查詢，找到九處真正裝備池定位及93個共用指標交叉參照。
這些是搜尋範圍，不宣稱直接xref涵蓋所有間接存取。派生名稱只供導覽。
已分級語意回填[怪物投射位址索引](monster-ranged-address-index.json)，窄匯出`ida-writers-indexed.json`合併註記與原始bytes。

首次查詢漏掉已知`0x1EC19`正對照。IDA把16位立即數`0xCD92`延伸成uint64；
原始值保留，候選按16位正規化後正對照通過。首次腳本與`ida-error.json`保留。
這是查詢問題，沒有原版或remake缺陷。另一次技能掛載未跨過既有symlink，改掛實際唯讀來源後通過。

## 原版生命週期補驗

`run-lifetime.sh`從固定dosgolem commit匯出隔離副本。
`lifetime_oracle_test.go`先由真實MZ啟動到main，核實整個1536-byte池為零。
複製常式只替換入口堆疊與出口返回，待驗`0x1EB9A..0x1EC46`指令保持；補丁前後bytes記錄於`lifetime-original.json`。
測試的32-byte標頭為受控零值，96-byte裝備來自現行v2；不以注入標頭證明能力欄位。

- 32種真實v2資料各驗六名不覆寫row6、七名覆寫、後續三名保留。
- 原版與Go在獨立比較前固定seed1；七次選擇及終seed相符，沒有亂數stub或挑選結果。
- 18組頭部來源0／6／10／11／33／255與耐久31／32／255核實。耐久只扣32或清整筆，後續三名保留。
- 封存正常三場實際數量為3、4、1，共八筆完整裝備複製；第二名51個讀取／階段點均零，觀察期無該別名寫入。

有界常式證明來源與寫入邊界，不當作正常第七名GUI或完整戰局驗收。
原型的`TestRE269V2HistoryAgainstOriginal`消費實際原版32種收據，同seed比較選擇、來源與跨場保留。
空槽填充、未綁暫存、格式不符、完整byte分支、頭部清除與其他別名未知另有邊界檢查。
現行128套只有一個空頭部槽，其來源＋3與耐久旗標均零；不以合成填充邊界宣稱新的資料母體已驗。

## 原型、審查與接線

先保留正式339輸入，`prepare.py`建立唯讀覆蓋原型。
第七名怪物選定裝備後保存原始頭部來源，並觀察該actor可變頭部槽。槽被清除後來源變零。
暫存延續於同一程序，伏擊、任務、競技場共用；新程序從啟動零值開始，不新增人物v3或怪物v2欄位。
首名來源也補上非空頭部清除後讀零；初始空槽的填充仍保留。

原型雙語正常／觀察1544狀態、實際保存1136狀態、新程序重載26狀態，共2706。
每語言213seed與1917角色記錄的50語意欄位零差，第二名原有16筆Command差異歸零。
完整正常trace保持；觀察只排`re265_inspection`後與無觀察版本相符。
完整v3只排`saved_at`，28張比較PNG保持；實際PNG檔含觀察副本，共40張。

`ready.py`封存`spec-draft.md`，`prepare-formal.py`保存`prototype-review.json`。審查通過後才將
[SPEC-022](../spec/022-battle-ammunition.md)增補升READY。
`promote.py`只改四份既有Go檔及新增暫存與回歸檢查，339Go／341輸入。
這是已選原版行為的來源修復；不涉及新的產品、操作、資料格式或美術取捨。

正式Linux SHA-256：`340c8492aafa789f480609a8b14d9b7dec2dd269f2e9e1e925f3fafab11c5051`。
與審查原型逐byte相同。十九套件通過，五個套件無測試檔。
正式1934正常、實際保存與新程序重載狀態、28PNG通過，完整v3只差saved_at，收據`formal-review.json`。
全輪原型與正式共4640狀態、68PNG。`seal.py`封存附件、外部輸入與341來源，`verify.py`獨立核對。

## 範圍與下一閘門

50語意欄位包括既定投影，不等於全部50-byte原始記錄相同。
原始＋0x2C／＋0x31、其他裝填別名、完整UI／滑鼠與整場完成條件仍開放。
RE-268長任務／平台屬舊dba813be，不能借作新340c8492驗收。下一步以既定流程補跑新來源有限回歸。
HD仍沿使用者確認的手繪媒材，解析度與版面待既有決策。
