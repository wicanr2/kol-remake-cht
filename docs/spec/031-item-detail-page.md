# SPEC-031：資料頁物品選中後的明細層

狀態：**READY**（Theodoric兩條正常序列的完整清單、明細、搬移／丟棄與保存切片CONFORMED；2026-10-04另完成等價shader跨平台續核。其他物品及動作仍待驗）

2026-09-14曾依DOSBox-X輔助收據標CONFORMED；原收據與形成原因保留於第6節。
RE-172首批正常清單有1／8像素差異，RE-173為8／9；RE-174動作後另有692像素
框形殘留，均保留於各批收據。第13節已閉合Theodoric兩條正常序列，不能外推全部物品及動作。

本規格只處理資料頁（`sub_16986`）選中一件非空物品後的畫面重畫，以及與這個
畫面直接相鄰的空格訊息、動作訊息與返回生命週期。它不是整個物品資料模型或
動作規則的重製規格。

## 1. 範圍與排除

納入：

- 選中物品後的 320×200 英文頁面圖層順序。
- 立繪、木屋背板、十七列清單與外框在選中後是否仍存在。
- 明細框 `(62,58)`–`(257,125)` 的填色、框線與右／下四像素投影。
- 空格訊息、動作完成訊息、送禮候選人與 U-turn 返回所需的既有狀態轉換。
- 中文模式沿用同一狀態分層並以 `book.CJKScale` 放大；中文沒有原版 oracle，
  不宣稱中文像素 parity。

排除：

- 物品記錄格式、十七格索引、穿戴／丟棄／送禮／食用規則本身。
- 重新設計英文或繁體中文文字、資料頁座標、圖示列或玩家操作方式。
- 未取得公開散布權的原版素材與執行檔。

## 2. 證據輸入與定位契約

| 輸入／工具 | 定位與雜湊 | 用途 | 證據等級 |
|---|---|---|---|
| `KOLTOWN-unpacked.EXE` | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`；IDA 線性有效位址，重建 unpacked MZ | 原版控制流與呼叫順序 | 已證實 |
| `workplace/analysis/koltown/KOLTOWN-unpacked.EXE.asm` | IDA Pro 9.4、`ida-pro-9.4-idapython:locked-v1`；保留 `sub_16986`、`sub_1BA9D`、`sub_1B7E9` 原名 | 可回查的反組譯證據 | 已證實 |
| `workplace/analysis/koltown/mark-box.json` | 同一輸入 SHA、IDA 9.4；`sub_1BA9D` `0x1BA9D..0x1BAFF`、`sub_1B7E9` `0x1B7E9..0x1B8B8` | 函式邊界與呼叫端索引 | 已證實 |
| `docs/re/RESEARCH-LOG.md` RE-127、RE-137 | 物品頁控制流、`sub_1BA9D` 複合框與原版畫格觀察 | 證據交接與既有勘誤 | 已證實 |
| `docs/verification/character-ui/town-sheet-items-run1/frames/frame-070.png`、`frame-072.png` | 原版 ZIP SHA-256 `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`；DOSBox-X 2026.06.02、fixed 3000 | 目前可重生的輔助畫面 oracle | 已證實（輔助；非 dosgolem 收據） |

反組譯中的位址都是 **IDA linear effective address for reconstructed unpacked MZ**；
不能與檔案偏移或其他工具的虛擬位址混用。新增語意保留原始函式名、位址與
operand；本規格沒有把推測性名稱寫回 IDA 資料庫。

## 3. 原版行為

### 3.1 完整物品頁

`sub_16986` 的完整頁分支在 `0x16995` 先呼叫 `sub_1103F(0)`，再依
`word_2DF2C` 呼叫 `sub_1103F(0x20)` 或 `sub_1103F(0x21)`／`sub_1103F(0x64)`，
並呼叫 `sub_18F3A(0)` 貼立繪。主迴圈 `0x16A11` 另外畫右側清單欄、外框、分隔線、
外框投影、動作圖示與 `sub_196A9(..., 0x11, ...)` 的十七列清單。

### 3.2 選中非空物品

在 `sub_16986:0x16BBF`，原版：

1. 呼叫 `sub_1103F(0)`，覆蓋完整 320×200 畫布；這一步之後，先前的立繪、木屋
   背板、十七列清單、外框與分隔線不再是畫面內容。
2. 以五個參數呼叫 `sub_1BA9D`：`(0x3E, 0x3A, 0x101, 0x7D, 0x0F)`，其幾何範圍
   是 `(62,58)`–`(257,125)`。
3. 設定 `word_2DF1A=0x0F`、`word_39A56=選中列`；非空記錄複製 16 bytes 後，
   呼叫 `sub_153FF(8)` 與 `sub_18857(選中列)` 印明細並建立動作圖示列。

這條分支沒有再次呼叫 `sub_1103F(0x20)`／`(0x21)`／`(0x64)`、`sub_18F3A(0)`、
`sub_196A9` 或完整頁的外框／分隔線序列。這是「選中後只剩木紋底圖與明細層」的
控制流證據，不是由畫面外觀反推。

### 3.3 `sub_1BA9D` 的複合框

`sub_1BA9D:0x1BA9D..0x1BAFF` 的順序是：

1. 以內縮一像素的參數呼叫 `sub_1B779` 填滿內部。
2. 以原始五個參數呼叫 `sub_1B6DC` 畫框線。
3. 以前四個參數呼叫 `sub_1B7E9`；`sub_1B7E9` 在其 `0x1B7E9..0x1B8B8` 內，
   以 `+4` 的右側與下側迴圈畫出各四像素的投影。

因此「明細框沒有投影」與原版控制流矛盾；這不是待猜的美術選擇。

### 3.4 空格與動作返回

空格會從 `0x16C66` 起印出 `0xCA`、`0xCB`／`0xCC` 與部位名稱，最後在
`0x16CE7` 等待一次按鍵；它與非空明細共用前面的 `sub_1103F(0)`，但不代表要
把選中後的十七列清單重新畫回來。

非空動作回到 `0x16CEC` 檢查 `byte_2DF50`；非零時清零並跳回 `0x16995`，也就是
重新建立完整頁。這是動作完成後的 full redraw，不能與選中當下的 blank-base
分支混用。

## 4. 重製版規格

### 4.1 畫面分層

英文 `drawItemPage` 與中文 `drawItemPageCJK` 都使用下列狀態判斷：

```text
interactionBase = p.Acting || p.Detail != nil || len(p.Message) > 0
```

- `interactionBase == false`：保留既有完整物品頁流程——整頁底圖、木屋背板、立繪、
  右欄、外框／分隔線／投影、十七列與左上小框。
- `interactionBase == true`：只先畫 `drawSheetPageFull`（`sub_1103F(0)` 對應的
  全頁底圖），不畫木屋背板、立繪、右欄、十七列、左上小框或完整頁外框。
- 明細／訊息／Note／送禮候選人的既有文字狀態與輸入結果不改；它們繼續在目前
  的明細區域繪製。`p.Note` 且 `p.Acting == false` 的動作完成後狀態走完整頁，
  對應原版 `byte_2DF50` 的 full redraw。

### 4.2 明細框

繪製 `l.Detail` 時固定依序執行：內部填色、`frame`、`shadow`。`shadow` 的右側與
下側必須各為四像素，使用現有 `town.Shadow`／`shadow` 原語；不得以文字裁切或
全頁差分例外取代。英文框仍為 `(62,58)`–`(257,125)`，中文框為同一份
`town.ItemPageLayout.Detail` 乘 `book.CJKScale`。

英文明細文字使用既有 `book-large` 8×8 字模；frame-072 的第一段字形起於
`y=66`，由 `goodDetail` 的原始列 `102` 經模式 8 的文字基準 `50` 換算取得。
這項位置與字模高度由 frame-072 像素列量測交叉確認；不是為了避開差異而加入的
允許區。

### 4.3 玩家狀態與保存

不改 `internal/town/itempage.go` 的 `AcceptItemSlot`、`afterItemAction`、
`closeItemAction` 或保存格式。選中、空格、動作、給隊友、U-turn 都必須仍由正常
玩家按鍵進入；截圖腳本只固定輸入，不得使用 direct-entry、座標注入或強制勝利。

## 5. 驗收契約

1. 既有 frame-070 路徑 `L R RET SPACE > RET > RET D` 必須維持通過；它保護完整
   物品頁、十七列與左上小框。
2. 新增正常路徑 `L R RET SPACE > RET > RET D RET`，產生選中後 320×200 PNG，
   與原版 `frame-072.png` 整頁逐點比較。
3. 比較只可明列原版滑鼠游標矩形 `(158,100,14,14)`；frame-070 的既有
   `(15,7,1,1)` 灰點只屬完整清單 pair 的允許區。選中 pair 的明細框四邊與
   右／下四像素投影列為必須相同區，允許區外差異必須為 0。
4. 收據記錄重製建置／輸入腳本、兩側 PNG SHA-256、工具版本、尺寸、比較矩形、
   允許區、區外差異、首個差異座標，以及 dosgolem 能力檢查結果。DOSBox-X 畫面
   若仍是唯一原版輸入，標為輔助／bootstrap，不冒稱 strict dosgolem parity。
5. 回歸至少包含：完整頁、選中後清除立繪／清單、四像素投影、明細、Note、Message、
   Giving 與返回；再跑相關 Go 測試、完整測試與 Docker 內建置／GUI smoke。

## 6. 生命週期

2026-09-14 證據審查已完成：

- `sub_16986` 選中分支的 `0x16BBF`、`sub_1BA9D` 呼叫與 `0x16CEC` 返回控制流；
- `sub_1BA9D:0x1BAF1` 對 `sub_1B7E9` 的直接呼叫及四像素幾何；
- dosgolem 的能力檢查結果為 `false`：目前沒有 KOL 專屬適配器，因此 frame-072
  明確分類為 DOSBox-X `dosbox-bootstrap`，不作 strict dosgolem parity；
- 正常玩家路徑、CJK 非 parity 限制與現有 itempage 狀態測試的涵蓋範圍。

以上證據足以授權本切片實作，但不授權把 DOSBox-X 圖片包裝成 dosgolem 收據。
實作與驗收已完成：`internal/ui/itempage.go`／`itempage_cjk.go` 依上述分層繪製，
英文 frame-070／frame-072 正常玩家路徑與原版逐點收據均通過，明細框四邊及右／下
四像素投影列入必須相同區；相關 Go 產品測試、建置與 CJK 文字安全檢查也通過。
收據為 `docs/verification/remake-parity/receipt-town-sheet-items.json`，其中
frame-072 的允許區外差異為 0。由於目前 dosgolem 沒有 KOL 適配器，原版畫面仍是
DOSBox-X `dosbox-bootstrap` 輸入；本規格不宣稱 strict dosgolem parity。未通過的
差異仍應回到 RE／DRAFT，不在程式中默默增加例外。

## 7. 2026-10-01：空格訊息窄續核（CONFORMED）

原版已能由固定dosgolem `968ca774705853276ae55c8d1163c369481ac71f` 的正常啟動鏈
重生；第6節的能力判定是2026-09-14歷史，不作目前能力判定。舊DOSBox-X收據保持輔助分類。
新研究與定位入口為[物品明細續核](../re/20261001-item-detail.md)。本節先審查READY再實作，
最後以正常畫格與[正式收據](../verification/item-detail-20261001/receipt.json)升CONFORMED。

已證實：正常胸甲／腿甲的模式8明細與目前英文頁面一致；空的BODY／HANDS／FEET
卻呈現訊息起點、大寫、標點、圖示與確認返回差異。只授權下列窄修正：

- 英文空格訊息起點為字格欄8（x64）、像素列76；保留記錄0xCA的四個前導空白與`@`，
  將0xCB、部位名稱及原版額外印的`!`接在同一列。資料頁大寫旗標沿用既定契約。
- 背包空格使用0xCA＋0xCC；原版表外索引上限仍為0x0B。本次不改十七格映射。
- `sub_144C3`清除舊圖示、只加入SPACEBAR（id1），以`sub_11722(0)`等待。
  窄查證實ENTER與SPACE確認，其他方向／字元鍵不應移動物品清單或離開此層。
  第3.4節「等待一次按鍵」在此訂正為該確認等待，不能解作任意鍵。
- 確認後清除訊息、重建清單圖示與同一選中格。非空物品及保存格式不改。
- 中文使用同一訊息狀態、既定三倍起點及既有譯文：身上為「你的〈部位〉上沒有東西！」，
  背包為「你的口袋裡沒有東西！」。不能按英文順序把「上沒有東西」接在部位之前。

驗收預登記：相同正常前綴與21個後續鍵；英／繁中各27筆狀態，修正前後保存對照。
英文模式8非空／空格明細完整320×200逐像素比較，唯一允許區是實際可見原版滑鼠
矩形（由來源畫格與正常鼠移事件定位）；明細框、投影、文字與圖示均不能豁免。
初次「末角只剩一個像素」假設已被原版自身的夾限推翻，實際箭頭在(304,185)；
不把該探測判為通過。清單外框另有1／8像素差異，須保留，不稱完整清單已對拍。
相關狀態／輸入測試及Go建置必須通過，再升本節CONFORMED；父規格與#64整項不得借此結案。

最終程式為`6ffca0ddcaacd88b0d0bca899d4c23aad613654b8ba0c230f802dca72f929bd2`。
兩張正常模式8胸甲／腿甲、四張BODY／HANDS／FEET及返回HANDS空格確認頁，
英文完整320×200比較除了實際滑鼠矩形外均零差，框線、投影、標點與SPACEBAR列均納入。
原版箭頭每框實際67差異像素；不豁免文字或框。四個繁中字模框零差，六張中文頁目視可讀。
原版沒有繁中oracle；不宣稱中文像素parity。框邊缺像素、舊錯位頁與字模缺墨點均拒絕。
全部`internal/...`及`cmd/kol`Go測試通過，含104圓環框回歸、ENTER／SPACE返回及方向鍵等待。
四次物品檢視各27筆記錄欄位一致；軌跡不記資料頁暫態，實際返回由畫格及輸入測試驗證。
四次正常住宿保存除saved_at外完全相同：前版先檢視胸甲22鍵，新版含空格流程37鍵，
不把不同按鍵長度說成逐步同狀態。完整清單11張仍有1／8像素差異，收據列為未通過。
第7節的原版空口袋畫格證據已由第8節接續補足；其他物品種類及全部動作仍由#64追蹤。

重驗入口：[verify_items.py](../verification/item-detail-20261001/verify_items.py)。
在`kol-package:20260907`有界Docker中掛專案至`/src:ro`、已存在的
`workplace/item-detail-20261001`至`/out:ro`，執行`python3 /src/docs/verification/item-detail-20261001/verify_items.py`。
使用UID/GID、`--rm --network none --memory 1g --cpus 1 --pids-limit 64`及120秒外層逾時。
正式原版重生使用同收據目錄`original.sh`與`prepare.py`掛`/scratch:ro`、固定dosgolem
儲存庫掛`/repo:ro`、本機專用輸出掛`/out:rw`，設`KOL_ITEM_ATTEMPT`為未存在名稱；
`KOL_ITEM_KEYS`依收據original_keys送入。來源檔與資料庫唯讀，所有掛載先驗存在與型態。

## 8. 2026-10-01：十七欄位與名稱置中續核（CONFORMED）

沿用第7節固定輸入與工具身分；新序列由正常書本啟動選Theodoric，逐一選中、確認並
返回17欄，最後DOWN從16繞至0、UP繞至16、再確認空口袋，U-TURN返回資料頁及旅店。
入口為既有研究的[本輪追加](../re/20261001-item-detail.md#十七欄位續核與名稱置中re-173)，
本機重生腳本`workplace/item-detail-20261001/original_all.sh`、`prepare_all.py`、
`all-keys.json`及`remake_all.py`。原版61個後續鍵，重製版含啟動前綴66鍵；固定seed1。

已證實：原版gui3正常畫格中，頭盔、胸甲、腿甲及13個空欄的明細與現行英文程式
除既定滑鼠外相同。六個空口袋均使用IN YOUR POCKET，空手武器欄使用ON YOUR WEAPON。
巨劍明細另有220個名稱像素差異，全部在y66..72、x120..206；其他欄位與動作圖示相同。
清單續畫仍有8／9像素差異，本節不豁免或修正它們。

名稱置中已由IDA原始運算及正常巨劍畫格交叉證實：`sub_153FF`在`0x15420`先增加
計數再於`0x15424..0x1542B`檢查結尾NUL，因此計數為名稱長度＋1；`0x15436..0x15448`
執行有號除法，字格欄為`mode + 1 + (24 - (length + 1)) / 2`。商店mode13、物品頁
mode8使用同一函式。修正前`goodDetail`省略NUL，偶數長度名稱因此多偏右一格。
只授權補回此計數；名稱原文、翻譯、角色記錄、物品身分及保存格式不變。
定位索引為[item-name-address-index.json](../re/item-name-address-index.json)，來源
KOLTOWN-unpacked.EXE雜湊與IDA linear位址基準同第7節；不改正式.i64或原始名稱。

驗收預登記：重製版兩語言重跑相同66鍵／67筆記錄；18張實際選中頁（17欄＋重訪末格）
英文完整320×200除既定14×14滑鼠矩形外必須零差。至少驗偶數巨劍與奇數盔甲名稱，
所有空欄中文字模及安全矩形、正常返回與繞回；缺框像素、舊巨劍錯位為拒絕反例。
商店與物品頁共用名稱置中，須以既有商店正常選品測試核對奇偶長度及實際欄位。
全內部分層測試及104圓環框回歸通過後，才可把本節升CONFORMED；#64整項仍開放。

最終程式為`351464d1bbff6848defd94e2346b7c4c4741e7299f0842abe340e855f6b10ec2`，
位置`workplace/item-detail-20261001/kol-items-name-linux`。18張實際選中頁及返回資料頁／
旅店兩張完整英文320×200均除原版滑鼠外零差；舊巨劍220像素錯位與框邊缺一像素皆拒絕。
十四個空欄／重訪口袋繁中字模框零差，中文巨劍名稱只在預期24像素平移範圍內改變，
其餘66張每語言畫格保持相同。四組各67筆記錄欄位一致；四份各73鍵正常檢視後住宿
保存除saved_at外完全相同。全內部分層及cmd/kol測試通過，含商店四種原名欄位及104圓環框。

正式[十七欄收據](../verification/item-detail-20261001/all-slots-receipt.json)保存72附件，
其中61張PNG；[重驗入口](../verification/item-detail-20261001/verify_slots.py)在第7節相同
有界、無網路Docker掛載中執行。原版重生入口改用同目錄`original_all.sh`及`prepare_all.py`，
將前者作入口並依收據original_keys設定KOL_ITEM_KEYS；使用未存在的KOL_ITEM_ATTEMPT。
四筆名稱索引與十筆既有語意自動合併至本機`item-name-annotated-private.json`，正式.i64未改。
完整清單續畫仍未通過：37張有8／9像素差異，未新增允許區。只完成Theodoric四件非空
物品及十三空欄，不外推全部物品種類、所有動作、#64或整個重製版完成。

## 9. 2026-10-01：正常搬移與丟棄確認（明細與保存切片CONFORMED）

本節先READY才實作；已通過範圍限下列明細／等待／物品保存，完整清單續畫仍未通過。

RE-174沿用本文來源雜湊、IDA Pro 9.4／IDA linear與固定dosgolem身分。
原版由正常書本選Theodoric進旅店，共104後續鍵：腰帶巨劍→手上→腰帶→第一口袋
→手上→丟棄提問並選NO→再提問選YES→確認訊息→檢視空武器→返回旅店。
唯讀活動記錄基址為`(DS:BB82 + BB88h) mod 65536`，依`0x16BF8..0x16C05`
原始取址式及每張畫格的Theodoric名稱核實。記錄與原版payload只留本機。
入口見[研究追加](../re/20261001-item-detail.md#正常搬移與丟棄確認re-174)與
[分級索引](../re/item-action-address-index.json)。

已證實：正常拿取、腰帶與收納由sub_18CD5交換完整16 bytes；空物品仍以首byte為零
判斷，不能要求未用尾端也歸零。每次動作返回完整頁時`0x16CFF`跳16995，169B9重設
選中欄0；單純UTURN或空欄確認沿原有清單返回。取消丟棄仍有1898F的重畫旗標，
不等同動作層UTURN。本輪只授權成功搬移及此丟棄流程的重設，不外推食用等其他分支。
18986..1898F明確讓圖示3繞過重建旗標；維持食用的既有選取，不把本輪修改套入該支。

丟棄契約：選圖示0x15後先以18D2A在明細下方疊框`(56,138)..(266,152)`及四像素
投影，文字欄8（x64）、列141；保留上方物品明細。印記錄D1，144EF列YES／NO（id7／8），
預設YES。ENTER確認所選，方向鍵／SPACE不確認或移物品欄。NO不改物品，返回首欄。
YES才清空選中物品，疊印D2並只列SPACEBAR；此時上方明細仍保留，角色資料已清空。
ENTER／SPACE確認後回首欄。中文沿用209／210譯文、同一幾何三倍與既有字模，不新增用詞。

目前程式351464d1在第一動作後仍停第9欄，後續正常鍵因此選錯；丟棄直接清空、沒有
提問或確認等待。只授權補內部暫態、等待與上述疊框，不改持久格式、角色名或物品record。
驗收預登記：兩語言重播109正常鍵／110筆記錄、動作前後五條搬移／取消／清除狀態，
英文明細及丟棄疊框整頁除既定滑鼠外零差，中文確認／完成字模及安全矩形，正常住宿
保存與重載。原版清單x15的差異仍不得豁免；父規格及#64依原範圍保留。

本輪正式驗證入口為[verify_actions.py](../verification/item-detail-20261001/verify_actions.py)，
收據寫入同目錄`actions-receipt.json`。重生所需原版記錄及完整存檔只留本機；
工具逐鍵重算並明列未通過清單，不以明細頁零差替代完整清單驗收。

最終程式`c1e3b410380c338d5d1e6c799daee00a75101ea688dfa836a9c94afdc4dd02e7`，
本機`workplace/item-detail-20261001/kol-items-actions-final-linux`。
[正式收據](../verification/item-detail-20261001/actions-receipt.json)記錄23個英文完整頁
除實際滑鼠外零差、四個繁中確認／完成字模框零差，雙語各110筆狀態相同。
六個階段、雙語各自正常住宿及重載再住宿共24次；十七物品欄非空16-byte記錄與
原版同期動作後記錄一致，空欄依首byte正規化。十二重載選中頁與相應正常動作頁
完全相同，每階段英中完整保存僅排除saved_at即相同；不外推原版金幣／時間。
前版同鍵序列的提問頁有36,394像素差異，反例拒絕；明細框缺一像素及中文字模缺墨點
亦拒絕。全部內部分層與UI／cmd測試及104圓環框回歸通過。

仍有76個完整清單畫格未通過。搬移後除x15外另有`(17,125)..(149,152)`框形殘留
692像素差異，來源／還原生命週期尚未證實；未將它加入允許區，也未自行猜補畫面。
此新缺口列入#64，父規格仍READY；不宣稱所有物品／背包洞位／食用／送禮完成。

## 10. 2026-10-01：物品動作後鏡像與立繪（鏡像與殘留切片CONFORMED）

RE-175的來源雜湊、工具與位址基準沿用本文，入口為[研究追加](../re/20261001-item-detail.md#動作後物品鏡像與立繪re-175)
及[七筆分級索引](../re/item-figure-address-index.json)。已證實：原版18CD5交換同一角色
記錄兩欄的16 bytes，18F3A再讀該記錄，首byte零即跳過裝備圖。重製正常保存的typed
腰帶／手持與Carried鏡像不同；Figure回退舊腰帶，動作返回頁出現692像素多餘圖形。

契約：物品動作完成時，以現有SyncCarried及Good.Record更新發言者十七欄鏡像，
必須先於refreshItemFigure；YES清除後在確認等待前即同步，NO不清除物品。
送禮成功也同步接收者；保留原有動作、圖示及等待規則。立繪原始回退記錄少於16 bytes
或首byte零時回nil，不以未用尾端判定非空。使用完整record，沒有來源record時沿用
既有SyncCarried空記錄契約，不猜補未知bytes。

不新增保存欄位、不改schema／十七欄順序／物品名稱與記錄、不改背包緊湊模型；
多物品洞位及食用／送禮的正常原版流程仍依父層待辦，不能宣稱本節驗了全部動作。
legacy原始記錄回退仍保留，僅修空格判定；正常完整保存的typed／raw應相符。

驗收預登記：同一104原版鍵與109重製鍵，英中各110筆狀態相同；23已驗明細／等待
完整頁保持除滑鼠外零差。四次成功搬移、NO返回及YES確認後六完整重建頁應移除
692像素圖形；殘餘x15照實記錄，不加入允許區。六階段英中正常住宿及重載再住宿，
typed與raw十七欄非空16-byte記錄均與原版同期動作結果一致，空欄按首byte正規化；
其他角色／規則資料與前版同階段保存相同，只允許發言者Carried鏡像改為現在物品。
加入原始空格尾端非零與有效記錄、搬移前後鏡像／Figure及YES／NO回歸，跑全部
內部分層／cmd測試及104圓環框。所有證據匯出另建收據，原有收據與來源不覆寫。

正式重驗入口：[verify_mirror.py](../verification/item-detail-20261001/verify_mirror.py)，
收據寫入同目錄`mirror-receipt.json`；逐點保留清單未通過結果，原始記錄僅本機。

先READY再修正，最終程式`4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4`，
本機`workplace/item-detail-20261001/kol-items-mirror-linux`。
[正式收據](../verification/item-detail-20261001/mirror-receipt.json)保存29個英文完整明細／
等待／六次重建及返回頁，除實際滑鼠外全零差；前版692像素反例拒絕。
104原版畫格全部逐點核對，其中34張零差、70張清單未通過，未擴大允許區。
英中各110筆狀態相同；前後除摘要持物數及鏡像以外規則資料相同。
24份六階段正常保存／重載再保存的typed與raw十七欄均與原版同期物品記錄一致；
20份前版過期raw鏡像反例拒絕。其他持久欄位只排除saved_at／發言者Carried即與前版相同。
全部內部分層／UI／cmd測試及104圓環框通過；三個舊行為回歸確實失敗。

RE-174的692像素未知原因已由本節閉合：舊腰帶鏡像造成重畫時仍貼巨劍，修正後歸零。
剩餘x15的1／8像素及三張手持小框22文字像素照實列未通過；文字差異已存在前版，
不由本修正引入，待依sub_16D0B續核。繁中本輪104張畫面保持前版；目前
drawItemPageCJK沒有迭代Figure貼立繪，尚待補正式正常顯示驗證，不能宣稱繁中立繪完成。
上述缺口由#64追蹤，父規格維持READY；背包洞位與其他動作規則不在本節完成範圍。

## 11. 2026-10-01：繁中物品頁立繪（正常顯示切片CONFORMED）

證據審查：RE-175已證實正常搬移後英文立繪與原版六次重建頁零差；同一正常
流程的繁中step-021畫面沒有人物，drawItemPageCJK也未讀取ItemPage.Figure。
這是已證實的顯示遺漏。來源、雜湊、七層合成及位址證據沿用第10節，倍率沿用
[SPEC-030](030-cjk-character-sheet.md)的book.CJKScale，不新增美術或操作決策。

契約：完整清單層在背板之後、右欄／框線／小框之前，依Figure順序呼叫既有
pieceImage；X取byte欄乘PixelsPerByte再乘倍率，Y及圖像尺寸乘同一倍率，使用
最近鄰縮放。自訂盾牌與素材失敗均沿用pieceImage的處理；缺圖回傳錯誤。
互動底層依既有itemPageUsesInteractionBase判定，明細／空格／丟棄等待不新增立繪。
規則、typed data、原始鏡像、輸入與存檔均不改。空Figure不貼圖。

驗收預登記：同一正常109鍵，英中各110筆狀態與RE-175相同；全部110英文畫面
不變。繁中完整清單以原版同期角色記錄、固定portrait-layout及已驗PNG獨立合成
三倍立繪，逐像素核對，立繪之外所有文字／控制區與前版相同；互動頁完全不變。
六階段雙語正常住宿及重載再住宿共24份持久資料，排除saved_at後與RE-175相同。
至少保留初始腰帶、手持、收納、取消及丟棄後的正常畫面，並拒絕舊版漏圖、位移
與明細誤貼反例。跑全部內部分層／cmd及104圓環框回歸。英文70張清單差異仍保留，
原版無繁中畫面，本節只證明原作美術三倍顯示與狀態一致，不宣稱中文原版像素對拍。

正式重驗入口：[verify_cjk_figure.py](../verification/item-detail-20261001/verify_cjk_figure.py)，
收據另存同目錄cjk-figure-receipt.json；正常操作與保存重生腳本入口見WORKLOG。

先READY再實作，最終程式325ba73b2617743dd84697adc50e31c68e16152f6bf0093d3f967140be4d015a，
本機workplace/item-detail-20261001/kol-items-cjk-figure-linux。
[正式收據](../verification/item-detail-20261001/cjk-figure-receipt.json)保存43 PNG附件：
78張繁中完整清單與原版同期512-byte角色記錄、固定版面及已驗PNG獨立三倍合成全零差；
所有文字／小框／圖示保留前版。32張其他繁中頁及110張英文頁完全不變，雙語各110筆
狀態與RE-175一致。六階段24份正常住宿／重載再保存排除saved_at後完整持久資料不變，
十二重載明細頁亦不變。舊版漏圖、三倍立繪右移1px及明細誤貼人物反例均拒絕。
全部內部／UI／cmd測試及104圓環框通過；未新增字模、美術、規則或保存欄位。

第10節的繁中漏圖缺口由本節閉合，歷史收據保留。這只涵蓋Theodoric正常流程中的
腰帶／手持／收納／取消／丟棄造型，不外推自訂盾牌或其他角色全部造型驗收。
英文70張清單差異、三張手持小框文字與其他動作仍未完成，父規格READY、#64保持開放。

## 12. 2026-10-01：手持小框技能列號座標（正常顯示切片CONFORMED）

證據審查：來源與版本沿用本文，IDA Pro9.4／IDA linear的sub_16D0B資料流已證實
技能列號的值是首個匹配列索引+1，位置重新算floor(框X/8)+8，而非骰式起點再加8。
見[RE-177](../re/20261001-item-detail.md#手持小框技能列號位置re-177)與
[七筆索引](../re/item-mark-address-index.json)。正常三張手持框的類別87與第一技能列87
一致，列號1不改；沒有新規則、資料格式或忠實度取捨。

契約：ItemMark共用一個技能列號像素X計算，等於(X/CellWidth+8)*CellWidth；
英文傳入對應字格欄，繁中傳入該像素X乘book.CJKScale，不乘中文字模Advance。
手持X110時英文x168、繁中x504。其餘字串、骰式、重量、文字列及框線不改；Skill空字串
仍不輸出。只修顯示座標，不改武器技能搜尋、物品動作、typed／raw資料或存檔。

驗收預登記：同一正常109鍵，英中各110筆狀態不變；英文三個手持框22文字像素差異
消除，僅剩原x15差異，不豁免。其他107英文頁不變。繁中三框以既有字模獨立核對
x504的新列號，x464舊列號清除；其餘107畫面完全不變，所有文字變動限該兩字模格。
六階段24份正常住宿／重載再住宿排除saved_at後完整持久資料不變；十二重載明細頁
不變。前版偏移與缺墨點反例拒絕；全部內部／cmd及104圓環框通過。原版無繁中畫面，
繁中只驗字模與三倍幾何；其他人物技能列及完整清單續畫仍依父層驗收。

正式重驗入口：[verify_mark.py](../verification/item-detail-20261001/verify_mark.py)，
收據另存mark-receipt.json；原版記錄及正常重生腳本沿WORKLOG入口，歷史收據不覆寫。

先READY再實作；最終程式821ee273ce3d78f658d4efa83b2716ec89a12e4002fa0365b7c8bf17fc86b78a，
本機workplace/item-detail-20261001/kol-items-mark-linux。
[正式收據](../verification/item-detail-20261001/mark-receipt.json)保存27 PNG附件：三個英文
手持小框完整100×22區域與原版零差，前版每框22文字像素偏移消除；整頁仍有x15的一像素，
未豁免。三繁中畫格與字模獨立重建完整畫面一致，舊x464列號移至x504；其他107畫格
兩語言各自完全不變。雙語110筆狀態不變，六階段24份完整持久資料排除saved_at後不變，
十二重載明細頁不變。偏移與缺墨點反例拒絕，全部內部／cmd與104圓環框通過。

104原版畫格逐張核對，34零差、70未通過，剩餘差異僅x15的1／8像素；第10／11節的
手持小框文字缺口由本節閉合，歷史收據與形成原因保留。其他人物技能列母體、全部物品
動作及清單續畫仍未全驗，父規格READY、#64不結案。

## 13. 2026-10-01：英文清單快取還原（正常顯示切片CONFORMED）

證據審查：RE-178的固定dosgolem正常104鍵、327次唯讀呼叫與四平面欄快照，
配合IDA Pro 9.4／IDA linear原始定位，已證實x15差異是快取步長及還原生命週期。
七次正常重建的快取修改均符合8160步長，8000步長反例均拒絕；104原版PNG及
角色／UI狀態與既有gui-actions1不變。完整78清單的獨立曝光模型整頁除既定滑鼠
外全零差，解釋現行70張未通過頁。見[RE-178](../re/20261001-item-detail.md#清單左緣快取與曝光re-178)、
[九筆索引](../re/item-edge-address-index.json)及[來源收據](../verification/item-detail-20261001/edge-source-receipt.json)。
實作前，模型通過只證實來源；當時程式821ee273及其70張差異尚未改，不能升CONFORMED。

英文契約：以目前背景、背板及Figure順序合成清單快取；原版模式13在x15的快取
位元修改依RE-178四段範圍重現，原始欄位保留，不把8160誤改成8000。
初入及重建用新合成的背景畫完整頁，之後更新快取供還原；換列先從該快取還原
舊框byte矩形及26列，再畫新框，包含同列初始化。原版未要求重建的互動返回，
用已修改的快取畫完整頁，再畫框、清單、圖示底板及目前小框；動作要求重建時
依既有重設首欄流程重新合成快取。圖示底板覆蓋列164..195，不能遺漏此覆蓋。
資料頁關閉、角色改變及立繪重建要建立相應的新快取；還原歷史是畫面暫態，不能
加入持久存檔或改角色記錄、物品動作、選取、文字、素材及儲存schema。
實作從幾何、位元運算與有效狀態推導，不能依驗收畫格號硬填色或增加允許區。

本節依既有英文完整頁對拍閘門修正呈現。繁中維持已驗的三倍版面及字模；本節
不將EGA快取的特定像素缺陷直接外推為新的繁中美術要求。兩語言共用遊戲狀態，
繁中無原版像素oracle。其他人物／野外頁及其他物品動作仍須另做正常抽樣，
不能以Theodoric的固定來源替代全部母體。

驗收預登記：重製兩語言以同一seed1、109正常鍵重跑各110筆狀態，狀態與RE-177
相同；英文104整頁除已證實原版滑鼠矩形外全零差，特別包含首次換列、七次重建、
空武器確認返回及U-TURN選取。繁中110圖保持既有呈現。補跑正常十七欄與上下繞回
序列，防止只對固定動作序列硬補；原版來源及seed條件按既有收據重生。
六階段雙語正常住宿及重載再保存24份完整持久資料排除saved_at後與RE-177相同，
十二重載明細頁不變。前版每幀重畫、錯誤8000步長、錯誤重設還原歷史及少一個
框像素都應被拒絕；跑全部內部分層／cmd測試及104圓環框回歸，再升本節CONFORMED。
父規格及#64的其他物品、動作與正常完整戰鬥仍保持原有閘門。

來源重驗入口：[verify_edge_source.py](../verification/item-detail-20261001/verify_edge_source.py)。
以kol-package:20260907有界、無網路Docker執行，專案掛/src:ro、已存在的本機
workplace/item-detail-20261001掛/out:ro；使用UID/GID1000:1000、--rm、--memory 1g、
--cpus 1、--pids-limit 64、120秒外層逾時，執行python3該工具，不加--export。
原版重生入口是同一本機目錄original_edge.sh及prepare_edge.py；來源唯讀，設未存在的
KOL_ITEM_ATTEMPT與actions-keys.json的104鍵，用固定dosgolem儲存庫及現有Go工具鏈。
工具僅本機保留快取快照，八張已驗PNG依現有private GitHub特例保存，不能公開發布。

先READY再實作；最終程式fb185a0a1a9217feb25b66b2458cde1f6c15ae328f87ed79a9b6d291b57f5fba，
本機workplace/item-detail-20261001/kol-items-edge-final-linux。
ItemPage.RestoredEdge只記錄暫態還原範圍，GPU由框線之前的PNG合成取得x15欄，再按
四個已證實平面範圍修改色盤位元；換列、原快取返回及動作重建依各自生命週期處理。
圖示底板按唯一ItemPageLayout.ControlBack覆蓋，繁中呈現與保存格式不改。

[正式收據](../verification/item-detail-20261001/edge-receipt.json)與
[重驗入口](../verification/item-detail-20261001/verify_edge.py)記錄兩條正常序列：104動作
畫格及61張十七欄／繞回畫格，全部320×200英文整頁除實際原版滑鼠外零差。
沒有新增允許區；滑鼠區仍為(304,185,14,14)，每張實際67點。前版70張x15差異已閉合，
其餘英文像素不變；110張繁中畫面與前版完全相同，兩語言110／67筆狀態各自一致。
六階段雙語正常住宿及重載再住宿共24份完整持久資料，只排除saved_at後與RE-177相同；
十二重載明細頁不變。實際前版70個失敗頁、首版漏底板的兩像素與整頁框線單點損壞
均經正式比較拒絕。全部內部／cmd測試及104圓環框通過，暫態返回與重建回歸通過。

產品重驗使用上述同一隔離方式，外層逾時300秒，執行verify_edge.py、不加--export；
正常重生入口見WORKLOG的prepare_edge_final.py及四個雙語操作／保存工具。
15張新PNG另存，舊來源、收據及保存鎖定不改；收尾稽核另存edge-final-audit.json。
本節只閉合Theodoric兩條正常序列。其他人物、野外、全部物品動作及原版完整保存／
時間／金幣未全驗，父規格維持READY，#64與整場戰鬥的既有閘門仍開放。

## 14. 2026-10-04：物品頁shader等價語法與DirectX續核

本節沿用第13節已READY並通過有限對拍的四平面還原契約。只改
`internal/ui/itempage_edge.go`的shader語法，不新增原版行為或資料格式決定。
所有來源、二進位、腳本、失敗log及畫格由[收據](../verification/directx-item-edge-20261004/receipt.json)
固定SHA-256，327份Go中僅此檔與修改前不同；來源版本分別保存於本輪before-source及source-final。

同一Wine9、Ebitengine2.9.9及預設DirectX重現原候選D3DCompile E5017。
原先逐分量的`mod`改用`x - y * floor(x / y)`，與[Ebitengine Kage公式](https://ebitengine.org/en/documents/shader.html)相符。
只替換取餘數後仍觸發`node->type == HLSL_IR_CONSTANT`編譯器斷言；
把四平面動態索引區塊改成固定`PlaneRows[0..3]`及`bits.x/y/z/w`後，隔離候選同一26鍵通過。
證據足以支持這個等價區塊避開鎖版編譯器故障，不拆分宣稱某一次索引讀取或寫入是唯一原因。
16個色盤索引與16組平面遮罩共256個代數輸入相符；此核對不代替所有GPU輸入驗收。

正式採用前先核對隔離候選來源、27個完整狀態及8張英文畫格；通過後套入production，
再跑全部`internal/...`及`cmd/...`測試，結果通過。Linux及Wine預設DirectX兩語言各完成
189正常任務返回鍵與26正常戰利品重載鍵，八條完整軌跡及52張配對畫格均與前輪相同。
Windows兩語言各4,076正常UI腳本鍵從標題、六人招募、任務戰鬥、拾取與腰帶切換到旅館保存，
各4,077完整狀態及20張抽樣完整畫格與前輪Linux相同；完整`kol-save-v3`只排除`saved_at`即相同。
沒有注入座標、角色、物品或遭遇，不宣稱4,076鍵全由真實X11輸入。

另以新Windows保存各26真實X11鍵重載，共14張按鍵後完整畫格通過，
兩個遊戲程序正常關窗退出0，原保存SHA不變。既有正常保存另有相同52鍵／14畫格／兩次正常退出。
期望畫格在實際輸入前固定，每鍵只送一次，完整畫格30秒內未出現即失敗，沒有像素豁免。
保存中的發言者變更及完整畫格單像素變更反例均拒絕。

現行Linux候選SHA前綴b87ad738、Windows f525f6ff、macOS universal fcd417a8。
macOS雙架構及universal只通過建置、最低11.0與系統依賴等結構檢查，未執行或真機驗收。
原版任務返回75附件仍完整，舊收據綁定before-source；本節不增加原版完整戰鬥對拍範圍。
本節確認既有shader在上述有限正常路徑的等價運算與跨平台執行，父規格仍READY，#64維持開放。

重驗入口：[verify.py](../verification/directx-item-edge-20261004/verify.py)。
在`kol-package:20260907`容器將專案唯讀掛載至`/src`，執行
`python3 /src/docs/verification/directx-item-edge-20261004/verify.py`。
使用UID/GID1000:1000、`--rm --network none --memory 1g --cpus 1 --pids-limit 64`及120秒外層逾時。
正常重生腳本及固定輸入在`workplace/directx-item-edge-20261004/`，其中`final-wine.sh`執行返回、重載
及完整拾取保存，`gui-own-wine.sh`／`gui-own-control.sh`執行真實新保存重載；來源唯讀、輸出限本輪工作目錄。
1,965附件及1,096張PNG／資產快照全部留本機，沒有公開權利擴張、正式版號或封包。


## 15. 2026-10-04：繁中送禮候選人的既有呈現契約續核

沿用第4節READY的共同ItemPage狀態及同一幾何。正常標題、旅店招募、切回Theodoric、
MIRROR及物品頁證實英文Giving列出候選人，繁中只印Note。修正只在drawItemPageCJK
傳入既有Party並畫編號及原名，不改送禮規則、輸入、人物順序或保存格式。

[本機收據](../../workplace/item-give-player-path-20261004/receipt.json)封存1,442附件、658PNG，
含前後329份Go／module、920份視覺及工具輸入雜湊、鍵序、二進位與失敗分類。
Linux SHA-256為b6f8d01ad1a0d64ca0069061f306534ba10d72f0dd95076ba676172951d2c093。
20次正常UI程序共560筆狀態及PNG，正式與唯讀觀察版畫格及可比trace相同。
英文34張不變，繁中只有step-032新增兩個候選人。完整畫格由舊畫格加固定字模重建為零差：
1 Theodoric位於(192,198)、108×18、280墨點，2 Poontang位於(192,222)、103×18、301墨點。
兩框均在明細安全矩形內，舊漏名反例拒絕；目視兩名清晰可讀。

正常數字鍵2將同一16-byte巨劍記錄轉交Poontang第三口袋。使用正式預設住宿模式付費保存，
修正前後完整kol-save-v3只差saved_at。實際新保存退出重載，人物裝備、背包及攜帶鏡像相符；
再走26狀態正常旅店、發言者2及物品頁，第三口袋顯示巨劍。全部internal/...及cmd/kol測試通過。
重生入口在同一本機工作區：build_after.sh、run_baseline.sh、run_after.sh；
[唯讀重驗器](../../workplace/item-give-player-path-20261004/verify.py)於kol-package:20260907有界Docker，
唯讀掛專案至/src:ro，執行python3 /src/workplace/item-give-player-path-20261004/verify.py。
使用UID/GID1000:1000、--rm --network none、120秒、1GiB、1CPU及64程序；
來源變動後加--recorded-source只驗歷史。素材及收據僅本機，沒有新增公開權利。

沒有新增原版送禮oracle，不稱送禮規則parity。六人送禮的畫面還沒對拍；本批為正常UI腳本，
不稱全程真實X11按鍵。Windows／macOS未重建或執行，父規格仍READY、整項待辦維持。
SPEC-035空白名稱仍DRAFT，兩份待核准修正未套用。

## 16. 2026-10-04：送禮修正的現行跨平台續核

第15節修正由相同329份Go／module與920份視覺／工具輸入重建三平台。
Linux逐byte保持b6f8d01a；Windows amd64為86fcdb0b，macOS universal為108fb763，
完整SHA在[本機跨平台收據](../../workplace/item-give-platform-20261004/receipt.json)。
macOS的arm64、x86_64、最低版本11.0、arm64臨時簽章及系統相依庫檢查通過，未執行。

Wine9預設DirectX從標題、雙人招募、送禮、付費住宿到本次保存重載口袋，
兩語六條腳本共208筆狀態及PNG，完整軌跡與Linux同期完全相同。
另以136真實X11鍵走四次正式遊戲，24完整960×600畫格與預定Linux畫格相同，
四次均透過WM_DELETE_WINDOW正常退出0。兩語腳本及GUI完整v3只差saved_at，
重載不改原保存，收件人第三口袋巨劍可見。舊漏名畫格581差異像素反例拒絕。
15次第一次擷取仍為前頁，既有30秒完整畫格閘門均通過；沒有重送按鍵或像素允許區。

封存709附件與280PNG，Go來源及第15節原收據未改。重生腳本與來源在同一工作區；
[唯讀重驗](../../workplace/item-give-platform-20261004/verify.py)使用kol-package:20260907，
唯讀掛專案至/src:ro，執行python3 /src/workplace/item-give-platform-20261004/verify.py。
UID/GID1000:1000、--rm --network none、120秒、1GiB、1CPU及64程序上限；
來源更新後加--recorded-source只驗歷史。Windows Wine不代表真機，macOS仍僅結構；
原版送禮oracle未增加，六人送禮的畫面還沒對拍，父規格與整項待辦保持。
本批未選正式版號、打包或發布。

## 17. 2026-10-04：送禮確認與搬移時點續核（有限 CONFORMED）

來源與有限證據見[原版正常送禮續核](../re/20261001-item-detail.md#2026-10-04原版正常送禮與逐人保存續核)。
使用者於2026-10-04明確採用原版確認流程；本節已審查為READY。
正式接線只採下列成功轉交五檔，人物保存v3保持。
正式接線及新來源驗收見 [RE-220](../re/20261004-gift-confirmation.md)。
雙語300狀態、216真實鍵、26完整畫格、v3保存與重載通過。
有限範圍為兩人巨劍；父規格其他分支保持原狀態。第15、16節沒有新增原版oracle
是當時結果；本節新增六條正常原版鏈，不能反向改寫歷史或宣稱其餘動作完成。

| 狀態 | 原版已證實 | 修正前重製版 | 本輪窄修正 |
|---|---|---|---|
| 選收件人 | 按2後兩人角色不變，訊息／SPACEBAR等待 | 立即搬移並返回清單 | 增加不持久化的確認暫態 |
| 確認 | ENTER／SPACE才完整搬移；DOWN不確認 | 沒有此等待 | 正常輸入及SPACEBAR共用同一確認 |
| 完成 | 來源腰帶清零、第三口袋完整16 bytes、回首欄 | 物品結果相符，保留Note | 確認後清訊息並回首欄 |
| 保存 | 兩人各自確認後實際CHARDATA重載保持 | 現行v3保存重載保持 | 不改v3格式或住宿規則 |

可核對提案為[review.json](../../workplace/item-give-original-20261004/review.json)，
預覽為[繁中確認畫面](../../workplace/item-give-original-20261004/prototype2/gates-RET-zh-Hant/step-036.png)。
隔離原型由現行329輸入複製，五份變更限town/itempage.go、keys.go、session.go及
ui/itempage.go、itempage_cjk.go；readonly觀察main.go不進正式修正。
新GiftConfirm／GiftTarget／GiftMessage只為暫態，候選姓名與譯文沿用現行資料。
確認等待的訊息沿原版已量得(56,138)..(266,152)框，繁中依既有三倍版面。
雙語原型300狀態、完整保存除saved_at相同及52張重載圖相同，不能取代正式驗收。

本輪已依READY完成的有限驗收契約：同一固定seed及初始兩人，正常招募／選巨劍／選收件人，
確認前整份人物資料不變，ENTER與SPACE才搬移，DOWN、自選與越界數字不搬移。
雙語文字／安全矩形、正常真實GUI按鍵、正式保存與重載第三口袋都要驗；
跑相關與完整Go測試，其他正式玩家路徑回歸保持。核准只適用第17節，17.1及18節未獲核准。
只可升這條窄鏈，六人、滿背包、其他物品、野外送禮及完整畫面仍由原待辦追蹤。

唯讀重驗：使用kol-package:20260907，有界Docker將專案掛/src:ro、此既有工作區
掛/out:ro，執行 `python3 -B /out/verify.py`；UID/GID1000:1000、--rm --network none、
60秒外層逾時、1GiB、1CPU及64程序。工具核對1624附件與所有來源，不能只看檔案存在。
重生原版入口為run.sh及prepare.py／probe.go，再依各案例plan.json由run_case.sh送鍵；
設KOL_GIFT_CASE為新名稱，重載設KOL_GIFT_RELOAD=1及KOL_GIFT_PARENT為實際保存案例。
用同一Go image、原版遊戲/src/workplace/oracle/game唯讀、固定dosgolem/repo唯讀及
新的明確/out可寫目錄；掛載前驗存在與UID/GID。原型由prototype.sh、prepare_prototype.py
及run_prototype.py產生，指定KOL_GIFT_PROTO為新輸出名稱；先有對應原版gates收據。
來源與全部產物僅本機，沒有版號、公開散布或正式打包決定。

### 17.1 滿背包送禮與內層取消的新增證據（DRAFT）

新增[研究](../re/20261001-item-detail.md#2026-10-04滿背包送禮取消與六口袋保存續核)
與[本機收據](../../workplace/item-give-full-pack-20261004/receipt.json)。
本節只記新增已證實事實；第17節成功確認五檔已獲核准；本17.1節仍DRAFT，不沿新證據擴張正式程式。

- 六次正常轉交把兩個背包、手持、頭／軀幹／腿的16-byte物品填入六口袋，現行17欄結果相符。
- 第七件披風失敗時角色完全不變，原版SPACEBAR等待，DOWN無效，SPACE返回同一收件人提問。
- 收件人提問中U-TURN取消會要求完整重建首欄0；現行取消保留3，確認訊息也未回到提問。
- 原版兩人實際CHARDATA及現行v3正常保存重載保持全部六口袋與未轉交披風。
- 正常披風原版無防護列，現行PRT:1-6；全部Kind0及列距待獨立窄RE審查。

這些差異列入原UI／物品待辦，不稱滿背包完整CONFORMED。
若後續採用原版流程，需分清成功確認、失敗確認後重問及內層取消三種暫態，
不把一般動作U-TURN的保留選中欄規則套入送禮內層取消。
繁中沿現有「他的背包滿了！」譯文，文字／框與圖示依同一狀態切換；目前正式來源保持。
六人隊伍、其他物品、野外及完整畫面仍依既有工作清單，不降低完成條件。

唯讀重驗：kol-package:20260907，有界Docker將專案/src:ro及此既有工作區/out:ro，
執行`python3 -B /out/verify.py`；UID/GID1000:1000、--rm --network none、
120秒外層逾時、1GiB、1CPU、64程序。工具實際解碼1153PNG並重算全部物品與保存。
原版重生入口prepare.py／run_native.sh，保存重載為prepare_save.py／run_native_save.sh；
原版archive測試工具與run_case.sh由前輪封存工具驗SHA後重用，原版資料/src/workplace/oracle/game唯讀。
現行重生為run_remake.sh／run_remake.py及run_remake_save.sh／run_remake_save.py；
來源與全部產物只留本機，不發布或定版。

### 17.2 六人隊伍確認等待的漏繪肖像（有限CONFORMED）

RE-232沿用使用者已批准的第17節「選收件人後保留上圖明細與訊息」。
本節修正該確認畫面的漏繪，不新增送禮操作、資料格式或美術決策。
六人候選清單的既有呈現保持；其與原版的差異不在本節宣稱通過。

原版輸入為KOLTOWN.EXE，SHA-256
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
dosgolem固定968ca774、正常啟動觀察器7cdb3533，正常招募六人、選巨劍並送Kenneth。
旅店招呼後且新送鍵前設定seed1，只設定一次，不改角色、物品、座標或入口。
原版`gates-RET`與`save-SPACE`第63／64格證實確認等待保留六人肖像、物品明細及訊息。
實際原版215狀態包含六次12288-byte CHARDATA寫出及正常重載。
證據入口：[RE-232](../re/20261005-gift-six-party.md)，原始payload及新PNG僅本機。

已證實：Inn先畫隊伍，但物品頁重畫底圖後未重畫確認等待的隊伍肖像。
原版確認圖與修正前英文有11916個非滑鼠差異點，皆落在上方肖像區。
最小修正限英文與繁中物品繪圖：`GiftConfirm`為真時，在底圖之後、明細之前
呼叫既有隊伍肖像繪圖；PNG索引或素材載入錯誤沿原有error傳回。
轉交時點、完整16-byte物品、候選清單、資料v2、人物v3及其他頁都保持。

隔離驗收預先限定：雙語ENTER與SPACE正常鏈、全隊付費保存及新程序重載。
正式與唯讀觀察版的狀態／PNG相同，改動圖限第68／69步。
英文原版完整確認圖除既有滑鼠區(304,185,14,14)外零差，舊漏肖像圖應被拒絕。
10812組原始物品槽、原版持久欄位與真正v3重載相符；完整v3只排除saved_at後保持。
人物頁另有第六名肖像殘留，候選清單亦非原版像素版面，兩者保留FAIL與未完成狀態。
後續RE-233以六個正常發言者、31張英文整頁閉合本文MIRROR的肖像殘留與錯留隊長；
當時FAIL收據保持，候選清單的版面差異仍未完成。見SPEC-030的RE-233節。
不把本節通過擴張成整條UI流程parity、滿背包、所有物品、野外或HD驗收。

正式兩檔修正後的338輸入已鎖定。全量internal／cmd測試通過，乾淨正式Linux
`b810e7fc60ab9b2ae8f1ae6a4481c547edbc3f17d970a6e586779d9b5e71a77e`
與已驗760狀態的候選逐位元組相同。雙語兩人確認回歸各39狀態，只有第36／37張變動；
兩張英文完整確認頁與原版除同一滑鼠區外零差。雙語城鎮／戰鬥各184狀態及四張抽樣PNG保持。
共446正式回歸狀態。Windows與macOS雙架構及universal建置、架構與相依核對通過；真機未驗。
新原版PNG、資料與候選只留本機，沒有改版號、正式封包或公開素材契約。


### 17.3 選收件人版面比較（DRAFT，待共同決策）

此節只記隔離原型與已證實差異，不改第17節已批准的ENTER／SPACE確認。
現行正式Giving畫面用編號及完整姓名清單蓋掉物品明細；原版保留六人肖像與明細，
下方提示GIVE TO WHOM?。原版與現行仍用1至6選人，選人後才進確認等待。
新版面是否移除完整姓名清單，尚未獲使用者選擇，本節不升READY。

來源與原版正常六人鏈沿第17.2節固定KOLTOWN.EXE c2442c84、dosgolem968ca774、
工具7cdb3533與seed1。沒有新原版實驗，既有66／87／62狀態由封存收據及雜湊沿用。
原型來源為RE-233現行338輸入、正式Linux dbc465a3，只覆蓋兩份物品UI繪圖檔。
Giving期間在整頁底圖後畫全隊肖像，改走既有明細分支，把Note畫到下方已量得訊息框；
英文沿原版大寫顯示，繁中沿既有譯文。不改typed data、按鍵、轉交時點、資料v2或人物v3。

正常六人招募、Theodoric巨劍送Kenneth的ENTER／SPACE鏈與真正保存、新程序重載，
雙語新舊正式／唯讀觀察共1520狀態，整份typed trace保持；完整v3只排saved_at後相同。
各鏈只改第64至67四張選人圖，確認與轉交圖保持，重載27張圖完全相同。
53張英文原版整頁除既定滑鼠(304,185,14,14)外零差，八張舊選人圖被拒絕。
10812組完整16-byte物品槽含空槽逐位元組相同，沒有省略空槽尾端。
這是可丟棄原型證據，不能取代後續READY審查、正式接線及正常GUI驗收。

[雙語操作比較頁](../../workplace/gift-recipient-layout-20261005/review.html)可選不同操作步驟；
[RE-234](../re/20261005-gift-recipient-layout.md)保存工具、範圍與限制。
只核鍵盤原型，滑鼠語意仍由原有分支追蹤。滿背包、內層取消、其他物品、野外及HD未新增驗收。
生成工具確認HTML引用的所有PNG存在；尚未在瀏覽器實際操作比較頁。
新原版引用、原型與所有輸出僅本機，正式程式與封包保持。

## 18. 2026-10-04：類型0明細欄位與列距（DRAFT）

第17.1節披風差異已追到原版顯示分支。本節DRAFT等待明確核准，正式來源保持。
來源為KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`；
IDA Pro9.4、IDA linear，正式資料庫SHA-256
`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
以正式資料庫新副本匯出sub_153FF共283指令，原名、運算元、bytes與xref均符合既有匯出。

已證實：0x1549C取byte_37D04低2bit，0x1549F與0比較，為0時0x154A4
跳至0x155AF。此分支略過部位、傷害／防護與0x1551E的列加8；
負重仍在類別原列。0x155DD才將全部類型的列加8，印價值及狀況／營養。
營養的既有條件仍是部位高nibble10、類型0；其他印狀況。本提案只補漏掉的顯示閘門。

建議改動限internal/town/shop.go的goodDetail：Kind非0才換列與印Range；
Kind0不印Range，沿用原名、譯文、負重、價值、狀況／營養計算。
不改物品16-byte記錄、傷害／防護規則、轉交時點或kol-save-v3。
單檔提案為[review.json](../../workplace/item-misc-detail-20261004/review.json)，
預覽為[繁中披風明細](../../workplace/item-misc-detail-20261004/save-zh-Hant/step-086.png)。

隔離原型兩語四次正常UI程序，266狀態。完整軌跡只改披風明細的五段欄位；
兩語各104住宿狀態及29重載狀態，角色、圖示、六口袋及保存除saved_at保持。
每語只改正常披風step-086／087兩張圖，其餘262張保持。英文完整明細框196×68
與原版native-full/gift-087.png零差；繁中用前版實際字模獨立搬移，完整畫格零差。
舊英文明細及單像素損壞反例拒絕，沒有增加允許區或宣稱中文原版parity。
來源329份Go／module及920份視覺／資料輸入保持，原型單檔加唯讀main觀察入口。

核准後才完成READY審查與正式修改，跑相關及完整Go檢查，再補食品及其他類型0
正常入口、保存重載抽樣。此批原版正常路徑只覆蓋披風，不稱全物品CONFORMED。
成功轉交、滿背包確認、內層取消與六人分支仍由第17節及原待辦追蹤。

本機[收據](../../workplace/item-misc-detail-20261004/receipt.json)與
[唯讀重驗](../../workplace/item-misc-detail-20261004/verify.py)同區。
重驗使用kol-package:20260907，把專案/src:ro，執行python3 -B
/src/workplace/item-misc-detail-20261004/verify.py。使用UID/GID1000:1000、
--rm --network none、--memory 768m、--cpus 1、--pids-limit 64及120秒外層逾時。
正常重生入口prepare.py→prototype.sh→run.py；固定原版畫格與鍵序由父收據驗SHA後沿用。
IDA用locked-v1與disposable2.i64、ida.py；正式來源唯讀，輸出限已存在的本工作區。
首IDA的ASCII及驗證座標／路徑／譯文錯誤保留初稿與更正紀錄，乾淨重驗通過。
原版payload、原型與全部新產物僅本機，沒有新增散布權、封包或版號。
