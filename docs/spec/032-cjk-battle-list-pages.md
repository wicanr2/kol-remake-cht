# SPEC-032：戰鬥戰利品與法術清單的繁中版面

狀態：**READY**（2026-09-14；由 DRAFT 證據審查升級）。本規格固定 CJK 畫面、文字
安全矩形與輸入命中區，授權進入 implementation。原版沒有中文畫面，因此本規格不宣稱
中文像素 parity。

## 1. 範圍與排除

本規格只處理兩個已存在於戰鬥狀態機的呈現頁：

- 戰後戰利品的 `Items Found:`／`Characters:` 清單（`StageSpoils`）。
- 規劃段的 `Which Spell?`／`Select Affect` 清單（`PromptSpell`／`PromptAffect`）。

戰鬥規則、掉落判定、法術資料、清單順序、英文 320×200 版面、原版字模與狀態轉移
不在本規格內。不得用 direct-entry、座標注入、強制勝利或受控存檔來代替正常玩家路徑
的完成證據。

## 2. 證據與既有契約

Issue #90（`wicanr2/kol-remake-cht`）目前要求兩頁有中文畫面與截圖；遠端 Issue
狀態由 host `gh issue view 90` 於 2026-09-14 確認。

既有來源與落點：

- `cmd/kol/battle_render.go`：目前 CJK 分派只涵蓋遭遇與主戰場，戰利品／法術清單
  仍落到英文 renderer 再放大。
- `internal/ui/battle.go`：英文 renderer 的框、標題、列、選中樣式與戰利品圖示列。
- `internal/battle/spoils.go`：`SpoilBox`、`SpoilList`、兩階段標題與 `SpoilIcons`。
- `internal/battle/spell.go`：`SpellListBox`、`SpellListRows`、法術／正負效果標題。
- `docs/re/RESEARCH-LOG.md` 的 RE-115：戰利品框、列起點、行距、最多 16 列、選中列
  換色及底部 YES／U-TURN 圖示的已證實靜態證據；該段沒有原版戰利品畫格。
- `docs/i18n/ui-zh-Hant.json`：記錄 0x181／0x182 的「找到的物品：」／「角色：」、
  0x104／0x105 的法術清單標題、0x106／0x107 的正負效果文字。
- `docs/spec/030-cjk-character-sheet.md` 與 `docs/verification/cjk-ui/README.md`：
  中文畫布為 960×600，`book.CJKScale = 3`，中文字模格為 16×18，固定列只可截斷，
  不換行、不縮字。

目前沒有 KOL 的 dosgolem adapter，也沒有原版中文 oracle；因此收據必須把這兩項限制
寫清楚，不能把 DOSBox 或英文畫面改名成中文 parity 收據。

## 3. CJK 呈現契約

所有座標都從現有的 320×200 版面乘以 `book.CJKScale`，先呼叫
`BattleFieldCJK` 畫共同底層，再在同一張 960×600 畫布上覆蓋清單頁。

### 3.1 戰利品頁

- 框使用 `battle.SpoilBox`：原版 `(4,7)`、`125×146`，CJK 對應每個座標與尺寸乘三。
- 標題仍使用 `town.UISegments` 的定位段；譯文段直接以 CJK 字模畫在其指定欄列，
  不重新猜測 `~H02~V09`。
- 清單第一列是 `battle.SpoilList` 的欄 1、列 `0x13`，行距 8，最多顯示 16 列；
  CJK 起點、行距與欄寬全部乘三。
- 選中列保留英文 renderer 與 RE-115 的已知語意：選中為白字，其他列為青字，
  不額外畫白框。
- 清單文字不得越過框內安全右緣。安全右緣取框右側保留 4 個原版像素，超出時使用
  既有 `clip` 逐字截斷；不得換行或縮小字級。物品名與角色名仍沿用 typed data 的
  原文，未有翻譯證據時不得擅自翻譯。
- 清除 `battle.MessageBox` 對應的底圖區域後，使用原有 YES／U-TURN 圖示與
  `town.IconBar` 的相對位置，所有座標、圖示與選中框乘三。此區域仍覆蓋報獎勵卷軸，
  不新增第二個訊息框。

### 3.2 法術／正負效果頁

- 框使用 `battle.SpellListBox == battle.SheetBox`：原版 `(16,2)`、`121×75`，
  CJK 對應每個座標與尺寸乘三。
- 標題仍使用 `town.UISegments` 的 `~V07~H03` 定位段；法術標題與正負效果標題
  依狀態機既有 `SpellListTitle` 選擇，不改變資料或狀態。
- 清單使用 `battle.SpellListRows`：欄 3、列 `0x0F`、行距 8、最多 7 列；上述
  座標乘三後，以既有 `clip` 保護框內安全右緣。
- 法術名沿用角色資料的原文；正面／負面列沿用 `battleRecord` 已取得的譯文。兩者
  都不得以 `strings.ToUpper` 破壞中文內容。
- 目前英文 renderer 使用白色框表示 `Plan.ListPick`；其原版游標外觀仍未完全對拍，
  因此 CJK 只保留這個既有重製版視覺契約並等比放大，不把白框宣稱為原版已證實語意。

## 4. 輸入命中區

- 戰利品每一列的 `input.Row` 必須與 renderer 共用 `SpoilList`，中文模式將 x、y、
  寬度與行距同乘三；英文模式的現有座標與 token 不變。
- `PromptSpell`／`PromptAffect` 必須列出可見清單列的 `input.Row`，游標移動仍只送
  `U`／`D`／`RET`，不得新增直接設游標的 API。可見列數分別取既有法術清單長度與
  `battle.AffectRecords()`，並受 `SpellListRows.Visible` 限制。
- CJK 命中矩形必須在 960×600 畫布內、列間不重疊，並與畫面上的文字安全矩形同一
  起點；底部圖示沿用既有 CJK `iconTargets`。

## 5. 驗收閘門

升 CONFORMED 前必須全部通過：

1. `cmd/kol/battle_render.go` 對 `StageSpoils`、`PromptSpell`、`PromptAffect` 皆
   分派至 CJK renderer；英文分派與狀態機測試保持通過。
2. 新增純幾何／輸入測試，覆蓋中文戰利品列、法術列、正負效果列、選中 token、列間
   不重疊與 CJK 畫布邊界。
3. 在 Docker 內完成 Go 測試、CJK 字元覆蓋／文字安全矩形檢查與乾淨建置。
4. 以正常玩家操作路徑產生並檢視至少 `Items Found:`、`Characters:`、`Which Spell?`
   與 `Select Affect` 的繁中畫格；若某一頁只能用既有觀察回放，必須明列為觀察證據，
   不得冒充無受控輸入的正常玩家完成證據。
5. 收據記錄輸入檔名／雜湊、工具版本、畫布尺寸、命令、輸出 PNG 雜湊、證據等級與
   權利分類；不宣稱原版中文 parity，也不宣稱 dosgolem 已支援 KOL。

## 6. 不在本規格內

- 原版中文畫面、中文像素 parity、法術或物品名稱的重新翻譯。
- 法術選目標格、指定部位的規則與游標外觀的未證實原版細節。
- 戰利品掉落機率、背包交換、角色資料回寫與戰後保存。
- 英文 renderer 的既有版面修正；若英文證據另有缺口，另立規格處理。
