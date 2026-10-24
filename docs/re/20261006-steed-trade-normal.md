# RE-293：正常購馬、折抵與保存路徑

日期：2026-10-06。狀態：單人單城正常保存重載已抽驗；購馬確認流程原型 DRAFT。
不改正式程式或原完成條件。

既有 RE-119／120 與 SPEC-030 已有資料、折抵及坐騎呈現契約，舊正常按鍵抽樣
從加金冠的觀察存檔開始。本輪補合法原版名冊、正常起始與真正保存重載的有限證據。
既有觀察存檔收據仍保留其原範圍，不冒稱自然取得。

## 預先登記

- 固定原始 ZIP、原版遊戲唯讀；選名冊第13槽原有 ddduhh，原始金冠3000、城1旅店記錄14。正常書本角色清單下移13次，沒有建角、改名、改金冠或注入位置。
- dosgolem 使用既有固定 revision 968ca774705853276ae55c8d1163c369481ac71f 與正常啟動觀察器 7cdb3533；旅店招呼後且新送鍵前固定 seed 1 一次。原始輸入及完整SHA在工作區。
- 正常從 Trollsbane Inn 出門，以實際64×64地圖產生21方向鍵至 Holcroft Stables。先買 Lgt Warhorse 400，再用舊馬半價折抵 Draft Horse，零差額分支預期實收50。繁中正式顯示名為「馱馬」。
- 核對購買、交韁繩、坐騎圖、人物頁與移動率，正常走回旅店付60保存。預期金冠3000→2600→2550→2490；使用真正CHARDATA及v3另起程序重載。
- remake 使用現行342輸入／340Go、Linux991046ab，正常標題、角色清單、城內按鍵與 faithful_inn_stay。唯讀觀察副本只增加序列化欄，不進正式來源。
- 首版計畫78個原版延伸鍵；單次原版上限160鍵／540秒。若按鍵段數或控制器錯誤，保留首次失敗與同種子修訂，不改角色、資金或挑亂數結果。

## 範圍

本輪限單人、城1、輕型戰馬換駄馬及真正保存。沒有全隊騎走、荒野、其他馬廄、
HD素材、平台真機或正式封包驗收。原版沒有中文畫面；繁中另驗字模、安全矩形及正常操作。
若兩側資料更新時點不同，保留原始差異，不能以最後值一致冒稱全部逐步同狀態。

準備入口：[腳本](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006-prepare.py)。
工作區：[預先計畫](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/plan.json)；
原版 run-native.sh／run_case.sh、現行 build.sh／run-formal.sh／run-formal.py 在同區。
來源與全部新產物僅本機，沒有新增散布權或資料格式決定。

## 正常路徑結果

來源及原版執行器已逐檔核對。原版 KOLTOWN.EXE SHA-256：
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
角色欄位使用執行期 DGROUP 相對位移：六槽基址 `+0xBB88`、步幅 `0x300`，
錢包 `+0x2A…0x2C`、坐騎 `+0x1B0…0x1BF`。不是 IDA 線性位址。
既有 RE-119／120 的 IDA 資料流及其原始位址保留，不以本輪推測性名稱覆蓋。

| 抽驗 | 結果 | 證據等級與限制 |
|---|---|---|
| 正常原版路徑 | 原名冊第13槽、步行、輕型戰馬400、折抵馱馬50、回旅店付60；金冠3000→2600→2550→2490 | 已證實，單人單城，未注入金冠或位置 |
| 真正原版保存 | CHARDATA 12288 bytes，STEEDPIC／CHARPICS／SHLDPICS各12160 bytes；另起程序讀這四份真正保存 | 已證實，未從觀察記錄製作匯入保存 |
| 原版新程序重載 | 四個有效城鎮狀態均保持2490金冠、完整16-byte坐騎及移動率10；沒有再次寫檔 | 已證實，正常角色清單再選同一人 |
| 現行版雙語路徑 | 正式／唯讀觀察各93狀態，四程序均正常保存；另起程序各22狀態讀v3 | 已證實，英文及繁中，現行流程與原版仍不同 |
| 原型雙語路徑 | 各97狀態及各22重載狀態；156個映射狀態、754項金冠、完整坐騎、移動率、隊伍數、室內圖示及選框核對零差 | 已證實於列出的欄位，DRAFT原型，不是全畫面或全記錄parity |
| 觀察副本中立 | 正式與觀察的230個狀態去除新增觀察欄後相同；230組PNG逐byte相同 | 已證實，沒有改寫正式來源 |
| 保存相容性 | 正式／觀察／原型雙語六份完整v3，除saved_at外所有欄位相同 | 已證實，不改怪物v2或人物v3 |
| 馱馬圖 | 原版正常人物頁、原版重載、正式英文頁與原型英文頁均1520像素符合來源PNG | 已證實，本輪只驗馱馬 |
| 中文名牌 | 正式、原型及兩種觀察重載共四頁「馱馬」字模零差；寬32px、安全104×18px；所驗四段文字log缺字0 | 已證實，不等於全部中文或所有最長字串完成 |

本輪完整原版角色保存相對原始 CHARDATA 只有第13槽13個byte改變。
除金冠及坐騎之外，`+0x97` 為 `0→1`。差分保留，語意未在本切片判定，
不猜成正式v3欄位，不宣稱整份768-byte記錄一致。
正常離開時第78格已換成 KOLBOOK，舊觀察器的城鎮DS欄位失效。
該格的PNG及raw資料保留，但不解讀其隊伍數54783等數字。
同狀態映射只採城鎮第0…77格；第78格只作正常返回書本的畫面證據。

## 首次差異與修訂

首版原版79狀態、正式／觀察雙語各97狀態，全部保留。
按連續商品清單設計的鍵序沒有完成任何保存。
原版輕型戰馬在第32格已扣金冠，但坐騎16-byte仍空；第33格讀完交付才寫入，
回一般招呼頁五圖示。現行版第50格立即寫坐騎，第51格回兩圖示商品清單。
此差異已證實，不能只寫成驗證腳本錯誤。

修訂鍵序分開兩側控制器，角色、seed1、資料、路線及上限不變。
原版重新進交易；第41格只顯示交舊韁繩、第42格顯示接新韁繩，
第43格才換16-byte坐騎並回招呼頁。現行版兩句合為一頁，確認一次回清單。
兩側正常保存重載的終值一致，不能據此宣稱中間每一步相同。

準備時第一次把父收據的附件物件當成SHA字串，失敗版本及輸入保留在
`failed-preflight/`。重載包裝腳本第一次把正式trace只在變更事件列出的party欄
當成每格必有，驗證器KeyError；實際程序退出0、保存可讀，首次產物保留。
修訂驗證器對唯讀觀察查全隊欄，再對正式／觀察PNG及trace驗中立，
同容器映像重跑正常重載。這兩件屬工具問題，不當作產品缺陷。

首次唯讀封存核對在重算中遇到矩形tuple與JSON陣列的型別比較差異。
來源及逐格值已核對，失敗收據、工具及當時文件原文保存於
`failed-readonly-audit-v1/`。修訂為先依JSON契約正規化再比對，重封存並用
同映像、同唯讀命令重跑，不改原始值、比較欄位或玩家程式。

## DRAFT操作契約與選項

依 [grilling](/home/anr2/.codex/skills/grilling/SKILL.md) 的確認閘門，這一段只登記
待選方案，不是READY或正式規則。正式程式仍回商品清單。

1. 建議採原版：答應先扣錢；有舊馬時交舊、接新各確認一次；最後才更新坐騎
   及資料頁移動率，再回一般招呼頁。繼續買馬須重新選交易。
2. 保留現行：購買時立即換坐騎，兩句合併、確認一次，回商品清單。連買較快，
   明示為玩家流程的忠實度差異。

可丟棄原型只以Go overlay替換馬廄購買函式，沿同typed商品及PNG，
沒有寫回 `internal/town/shop.go`，沒有新增runtime資產或格式。
原型SHA-256：`9bad76eb440f7dbce96477d05a2f5275f95dc3a7817d48c5ff1f2a0754198b79`。
完整v3終值相同，變更可撤回；目前仍等待玩家路徑取捨。
若選原版，後續依DRAFT→審查→READY→正式接線→同狀態驗證流程處理。

入口：[三方並列畫面](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/review.html)、
[核對結果](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/comparison.json)、
[逐格映射](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/mapped-states.json)、
[修訂計畫](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/corrected-plan.json)、
[原型準備](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/prepare-prototype.py)、
[唯讀原始與現行資料核對](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/analyze.py)。
執行分析時使用既有 `eob-remake-go:1.26.7-ebiten2.9.9-audio`、
專案 `/src:ro`、本工作區 `/out:rw`，命令 `python3 -B /out/analyze.py`。
UID/GID1000:1000、無網路、1GiB／1CPU／64程序、外層180秒。

封存後使用[唯讀核對](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/verify.py)檢查
[收據](../../docs/verification/local/workspace-cleanup-20261024/browse/workplace/steed-trade-normal-20261006/receipt.json)的全部附件、外部來源及
原始映射重算。專案仍掛`/src:ro`，工作區改掛`/out:ro`，
Codex技能目錄唯讀掛`/codex:ro`，命令`python3 -B /out/verify.py`。
資源與UID/GID沿上段；需要保存稽核JSON時，只開放指定`audit.json`檔。

HD仍只核准手繪媒材；既有解析度與版面原型沿原問題待選。
本輪不結案完整CJK、全部坐騎、全隊騎走、野外、平台真機或十八項完成條件。
