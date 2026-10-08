# SPEC-032：戰鬥戰利品與法術清單的繁中版面

狀態：**CONFORMED**（2026-09-24；2026-09-14 由 DRAFT 證據審查升 READY）。
本規格固定 CJK 畫面、文字安全矩形與輸入命中區。四種畫格已由正常按鍵路徑產生並目視
檢查；戰利品路徑使用明示的觀察用怪物／任務資料。原版沒有中文畫面，因此本規格不宣稱
中文像素 parity，也不證明整場戰鬥規則一致。

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

- `cmd/kol/battle_render.go`：2026-09-14 READY 審查時，CJK 分派只涵蓋遭遇與
  主戰場，戰利品／法術清單仍落到英文 renderer 再放大；後續實作已補。
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
  既有 `clip` 逐字截斷；不得換行或縮小字級。物品名依 SPEC-033 的暫譯表只在
  中文繪製端替換，規則資料與存檔仍保留原文；角色自取姓名保持輸入原樣。
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

## 7. 完成收據與限制

- 法術與正負效果兩頁：`docs/verification/cjk-ui/battle-lists-receipt.json`；
  重製版正常按鍵路徑，畫格、工具與來源雜湊在收據內。
- 戰利品物品與角色兩頁：`docs/verification/cjk-ui/battle-spoils-winning-receipt.json`；
  固定 seed 39、六人存檔、觀察用怪物與任務地點資料，第 22 回合勝利後擷取。
  收據延伸驗證取消清單後返回野外及英中結束狀態一致。
- `cmd/kol/world_battle_mouse_test.go` 覆蓋清單列、token、CJK 座標、列間重疊與
  畫布邊界；2026-09-24 Docker 內完整 Go 測試通過。
- 字模覆蓋 5,785 筆文字、缺字 0；`docs/verification/cjk-ui/text-fit.json`
  的文字安全矩形 9,301 項失敗 0。這些只證明重製版中文顯示範圍，
  不代替英文原版畫面或戰鬥規則的同狀態對拍。

## 正常自然勝利的清單續核

[現行自然勝利收據](../verification/cjk-victory-list-20261004/receipt.json)與[唯讀重驗器](../verification/cjk-victory-list-20261004/verify.py)固定 seed 1，從真正標題正常招募 Olear／Theodoric。現行程式的兩語各 460 正常鍵在第 6 回合自然擊敗 SLEDGE，第 457 步是「找到的物品：」及戰錘，第 458 步是「角色：」及兩名隊員；第 459 步正常領取，第 460 步返回世界。沒有載入勝利存檔、指定遭遇、替換怪物／任務或強制結算。

繁中實際視窗從標題送 460 真實鍵，勝利、報酬、兩階段清單、領取及返回共六張完整 RGB 閘門相符，正常關窗退出 0。兩種語言各 698 正常鍵完成回城住宿保存；各 24 鍵重載查看戰錘及再住宿，另 48 真實鍵、14 完整畫格、兩次正常退出 0。完整 v3 除時間戳一致，兩人 34 槽物品、戰錘原名與 16-byte record、冒險點數 2891／749 保留；每次住宿各扣 56 金冠。五個清單欄位在腳本與實際視窗共十框字模零差，保留選中白字、未選中青字及同一安全矩形。未改正式 Go、字型、譯表、規則或保存格式。

這次只驗一件戰利品與兩名隊員，不宣稱最長清單或其他母體皆完成。正常拾取與保存另有有限證據，仍不把它擴張成本規格之外的全部掉落／交換規則或原版完整戰鬥 parity。

第 2 節與第 5 節的 dosgolem 支援限制屬當時快照。現行已有[原版任務返回的有限收據](../verification/original-quest-return-20261004/receipt.json)，能力與現況由 [CONTEXT](../../CONTEXT.md) 指向的收據決定。這份清單續核沒有執行原版，原版 seed 為 null；不宣稱原版中文 oracle 或整場戰鬥能力已完成。

## 現行最長名稱的正常勝利清單

[最長勝利清單收據](../verification/cjk-longest-victory-20261004/receipt.json)與[唯讀重驗](../verification/cjk-longest-victory-20261004/verify.py)重算現行940筆typed記錄中的316個可達候選，名稱上界80px。這個母體排除法術槽、坐騎與不能正常拿起的物品，不涵蓋尚未接線的四套怪物資料。

來源是此前4076正常鍵由標題產生的保存，本輪唯讀載入。固定seed1，使用已核准的原版住宿模式，正常逐人住宿離隊、招募Shyla並拿起Kenneth腰帶上的棘刺釘頭鎚。Kenneth正常DROP，兩語各570正常鍵在SLEDGE第7回合自然勝利。第565步物品清單含棘刺釘頭鎚與戰錘，第566步選Kenneth，第567步正常領回，第570步返回世界。六人首次第31回合敗北的診斷保留，沒有挑選種子或替換遭遇。

選中最長名稱、未選中戰錘、兩個標題與Kenneth／Shyla姓名，在逐鍵呈現與真實視窗共十四框完整字模零差。各1914正常鍵經自然GREMLIN第20回合勝利回城，Kenneth住宿保存；各45鍵重載後正常再招募、查看棘刺釘頭鎚明細、再住宿。完整v3除時間戳相同，再保存只差Kenneth正常56金冠住宿，原名及16-byte記錄b9c12e2e40d08ca0000023ff2b3c9100保留。

繁中570真實鍵驗勝利清單，雙語重載各45真實鍵驗明細與保存，共23完整RGB閘門、三次正常關窗退出0。1914鍵是正常逐鍵呈現腳本，沒有將全部回城路線稱為真實X11操作。使用faithful_inn_stay=true，住宿會離隊；未改預設不離隊模式、字型、譯表、規則、資料或保存格式。本規格維持CONFORMED，這份有限驗收不代表四套裝備、所有清單或原版整場parity。
