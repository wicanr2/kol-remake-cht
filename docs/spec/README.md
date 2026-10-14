# 規格索引

目前共有 40 個 ID、41 份規格，SPEC-002 分為抽取契約與 EGA 集合。
目前收尾與 HD 場景驗收採 [SPEC-040：受控場景驗收](040-controlled-scene-acceptance.md)。
目前程式、驗收範圍及待決事項見 [CONTEXT](../../CONTEXT.md)，
未完成工作的完成條件見 [WORKLIST](../../WORKLIST.md)。
任務入口共用亂數的最新有限正式驗收見[驗收入口](../verification/quest-rng-20261007.md)。
父規格的 READY／CONFORMED 登記不代表所有子分支已驗收；各節的 DRAFT 仍須先確認。

規則規格（005～029）末尾都有「說明書對照」表，連回說明書 PDF 頁碼與繁中章節。
規格狀態以下方登記與各文件的最新審查為準。
早期實作落點與當時缺口見 [原始碼稽核](../verification/spec-source-audit-20260907.md)
及 [後續稽核](../verification/spec-source-audit-20260908.md)，這兩份保存歷史範圍。
下表只修正已證實的現況，不自動升降規格狀態。
所有原版衍生行為必須依序經過：

`RE 證據 → DRAFT → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED`

在 M1 executable 與資料身分完成前，不建立玩法規格。

| ID | 規格 | 狀態 | 是否授權實作 |
|---|---|---|---|
| SPEC-040 | [受控場景與中文畫面驗收](040-controlled-scene-acceptance.md) | CONFORMED | 本輪收尾改採受控場景；指定畫面、獨立驗收模式及中文結局排版已驗 |
| SPEC-039 | [HD 手繪美術與介面工作計畫](039-hd-art-ui.md) | 整體DRAFT；首批、設定、介面／頭像及七段結局有限CONFORMED | [介面及NPC](../verification/hd-interface-runtime-20261012.md)、[七段結局](../verification/hd-ending-runtime-20261014.md)已接開發版，F10與存讀檔已驗；其他圖層依原閘門 |
| SPEC-038 | [套件邊界、固定步進與注入點](038-package-boundaries.md) | READY | remake 工程契約；`internal/layers` 四項守門測試擋依賴方向、Ebitengine 越界、規則層讀時鐘與全域亂數 |
| SPEC-037 | [建築進門後的發言者選擇與好惡](037-building-greeting.md) | READY | 多人隊伍先選發言者、被討厭與被欣賞的訊息、全隊被趕、SCROLL 教派；原版關鍵點三組共 36,176 組一致；滑鼠點肖像見 RE-167 |
| SPEC-036 | [圖像與盾牌編輯器](036-picture-shield-editor.md) | CONFORMED | 點陣圖編輯器已實作；鍵盤路徑與原版同狀態逐 byte 一致，滑鼠路徑依靜態讀碼、未與原版實跑比較 |
| SPEC-035 | [法術編輯器的繁中頁面](035-cjk-spell-editor.md) | READY | 編輯頁、扣款與名稱輸入已接；[RE-213](../re/20261003-spell-edit-inn-save.md)已驗原版住宿保存及重載的有限欄位，營養差異保留；名稱空白邊界仍DRAFT，整體未驗收 |
| SPEC-034 | [法術名的中文注解](034-cjk-spell-name-gloss.md) | CONFORMED | 法術名保留原文，依說明書音節表加注解；48 筆法術記錄交叉驗證，三處正常按鍵畫格已驗 |
| SPEC-033 | [中文畫面的資料名稱顯示](033-cjk-display-names.md) | READY | 顯示譯名已接；[RE-261](../re/20261006-cjk-v2-name-population.md)核對怪物v2四套裝備、475筆戰利品候選及正式字寬，既有最長名稱正常畫格與字模相符；其他母體及完整正常路徑仍待驗 |
| SPEC-031 | [資料頁物品選中後的明細層](031-item-detail-page.md) | READY | Theodoric兩條正常序列的完整清單、明細、搬移／丟棄與保存切片CONFORMED；等價shader的有限DirectX續核通過；其他物品與動作仍待驗 |
| SPEC-032 | [戰鬥戰利品與法術清單的繁中版面](032-cjk-battle-list-pages.md) | CONFORMED | 四種清單畫格與正常按鍵路徑已驗，戰利品採明示觀察用資料；不宣稱原版中文或整場戰鬥 parity |
| SPEC-030 | [中文模式的畫布倍率與角色資料頁](030-cjk-character-sheet.md) | CONFORMED | 六人、坐騎及換頁已驗；[RE-233](../re/20261005-sheet-party-header.md)修正六人英文肖像層，[RE-285](../re/20261006-sheet-subpages.md)續驗正常重載子頁；名牌與其他分支仍由原項追蹤 |
| SPEC-029 | [怪物遠程能力初始化](029-monster-ranged-initialization.md) | READY | 遠程機率與轉換骰已接入首次規劃；正式實作 768＋2304 組通過 |
| SPEC-028 | [三個可關閉的加速選項](028-optional-battle-speed.md) | DRAFT | 已決需求與現有接口已核對；細節／原型待確認，不授權先加空開關 |
| SPEC-027 | [戰鬥共同出口傷勢回寫](027-battle-exit-wounds.md) | READY | 傷勢折算及敗北回寫已接；4608 組原版比較、五條結算至 v2 重載通過，正常玩家鏈與額外掉落待驗 |
| SPEC-026 | [天候圖層與正常動畫](026-weather-animation.md) | CONFORMED | 六組資料、128 像素捲雲與雨雪輪播；原版四組合成、六組分派、節拍與相位逐次對拍通過（RE-164） |
| SPEC-024 | [建築告別與確認](024-building-goodbye.md) | READY | 單人直接退出、多人告別確認已實作；正常原版與繁中 GUI 已抽驗 |
| SPEC-025 | [競技場固定地形](025-arena-field.md) | READY | 4096格原版地形比較已過；[正常資格鏈](../verification/arena-natural-20261004/receipt.json)已驗付費訓練、下注、自然敗北、保存及重載；這條鏈沒有新增原版整場或勝利升階GUI證據 |
| SPEC-023 | [完整角色、名冊與續玩保存](023-complete-save.md) | READY | 正式kol-save-v3，v2遷移、v1匯入、建角、住宿交易及世界時間已接；[RE-279](../re/20261006-move-current-regression.md)驗現行正常任務保存與新程序重載；戰鬥中途保存不在已核准格式內 |
| SPEC-001 | [原版 Oracle 準備與位址契約](001-original-oracle-preparation.md) | READY | 僅授權研究 derivative，不授權玩法實作 |
| SPEC-002 | [PNG 素材抽取契約](002-png-asset-extraction.md) | READY | 僅授權已關閉像素格式的抽取器，其他 archive 仍須逐批驗證 |
| SPEC-002 | [EGA 抽取集合](002-ega-extraction-set.md) | READY | 僅授權本集合的 PNG 抽取與驗證 |
| SPEC-003 | [繁中說明書翻譯與校對契約](003-manual-translation.md) | CONFORMED | 僅授權文件翻譯，不授權玩法／runtime |
| SPEC-004 | [OGG 音訊目錄與執行期契約](004-ogg-audio-catalog.md) | 標題有限CONFORMED；HD配樂DRAFT | 原雷聲正式播放已驗；[原創室內樂試聽](../verification/hd-music-20261012/review.html)技術通過，人耳待確認 |
| SPEC-005 | [角色建立與角色儲存區](005-character-creation.md) | READY | 建角、刪除、穩定槽位與保存已接；[RE-169](../re/character-creation.md#2026-10-01-名冊邊界正常-ui-訂正re-169)已證實原版滿槽時點、重名及字元集合；正式差異待決，圖像編輯另見SPEC-036 |
| SPEC-006 | [住宿、治療與存檔格式](006-rest-and-save.md) | READY | 僅授權 REST 三條路、remake 存檔讀寫與離線匯入 |
| SPEC-007 | [野外的時刻、日曆與天候](007-wilderness-time.md) | READY | 僅授權時刻、日曆、天候與紮營畫面；遭遇見 SPEC-011 |
| SPEC-008 | [戰場的畫面、視線與回合順序](008-combat-field.md) | READY | 含戰場、回合、獎勵與 AI 主幹；受阻打暈保留持物子切片有限 CONFORMED，見[正常保存驗收](../verification/blocked-stun-20261007.md)；完整 AI 正常路徑待驗 |
| SPEC-009 | [戰鬥的攻擊結算](009-combat-attack.md) | 父規格READY；RE-180自然來源61個逐回合執行段窄切片CONFORMED | 近戰骰序／怪物負重修正與夾具倒地轉換已驗；完整GUI、戰後保存及其他情境仍由V01追蹤 |
| SPEC-010 | [戰場上的指令與選目標](010-combat-commands.md) | READY；施法／射擊另依 012／013 | 指令列、游標、物品動作、下馬及任務物品已接線；未驗收分支見本文 |
| SPEC-011 | [野外的隨機遭遇](011-wilderness-ambush.md) | READY | 含 RE-112／128 治療問答及敘述、怪物v2第13節有限CONFORMED；正常治療及其他入口仍待驗 |
| SPEC-012 | [施法](012-combat-spells.md) | READY | 含 RE-128～131 護符、效果 5／7／8 及疲勞鏈；選部位互動已做，正常 GUI 收據仍待補 |
| SPEC-013 | [射擊](013-combat-missiles.md) | READY（飛行動畫除外） | 授權彈道、目標選擇、攻防與訊息、裝填那一手，以及敵人 AI 的射擊分支 |
| SPEC-014 | [任務的完成與逃走](014-quest-completion.md) | READY（第 3–5、8 節與撿起 CONFORMED） | 撿物、帶物逃走、交付與保存鏈經 RE-162 原版關鍵點與 `quest-chain` 驗證；逃走／落敗掉落仍待證據 |
| SPEC-015 | [結局](015-ending.md) | READY | RE-107 回報計數、保存與戰鬥收尾觸發已接；完整正常玩家結局鏈仍待驗 |
| SPEC-016 | [開場旅店接線](016-starting-inn.md) | READY | 已核准選隊及跨城建築整合；完整操作對拍仍待完成 |
| SPEC-017 | [護甲減傷及命中後耐久](017-armour-reduction.md) | READY | 授權兩類護甲減傷、命中後耐久及近戰／射擊回寫；正常 GUI 與完整保存仍待驗收 |
| SPEC-018 | [戰鬥左下人形與疲勞框](018-battle-body-panel.md) | CONFORMED | 人形、傷勢遮罩、疲勞條與被檢視人物；不含尚未完成的指令圖示及左上框 |
| SPEC-019 | [戰鬥左上肖像與裝備框](019-battle-equipment-panel.md) | CONFORMED | 肖像、裝備列與模式切換已接；彈藥標題及數字由 SPEC-022 後續證據補足 |
| SPEC-020 | [規劃人物的檢視門檻生命週期](020-planning-inspection-roll.md) | READY | 門檻、跳過順序與取消保持已接，原版局部比較及回歸通過；完整戰鬥同狀態仍待驗 |
| SPEC-021 | [戰鬥左下命令面板](021-battle-command-panel.md) | READY | 授權唯讀命令、招式、部位、防禦及方向顯示；正常操作對拍仍待完成 |
| SPEC-022 | [隊員射擊彈藥消耗](022-battle-ammunition.md) | READY | 扣除、裝填、耗盡停止與資訊框已接；原版局部比較及 Olear 正常射擊 20→19 已驗，弩與耗盡 GUI 待驗 |

## 早期待查清單（歷史記錄，待依後續 RE 回填）

2026-09-08 新增 [SPEC-016：開場旅店接線](016-starting-inn.md)，狀態 READY。
只整合已載入城鎮的選隊與既有建築流程；跨城載入及完整同狀態驗證仍待完成。

以下是較早登記的缺口，已有多項被 RE-107…131 與程式覆蓋；**不得再從這份清單推定
它們現在沒有證據或沒有實作**。保留原定位，回填入口見上述稽核報告：

- 角色升級（`0x21611` 那一條）。RE-103 的 operand 掃描指出 `word_345DE` 只有競技場
  `sub_2AAA8:0x2AFE6` 會寫進真值，所以它屬於**競技場**那一側；升級本身是角色記錄
  `[+0x29] += 1`，等競技場那一輪再寫
- 逃走或落敗之後掉東西那一段（`0x21765`），要先有裝備欄十二格的 typed data
- 法術的**選部位**那一次互動，以及效果 5 的疲勞鏈（SPEC-012 第 8 節）
- 任務的回報（打完之後回去跟誰交）
- 結局的觸發那一格（`[0x1DD]`）是誰寫的——流程本身已由 SPEC-015 解出
SPEC-029：[怪物遠程能力初始化](029-monster-ranged-initialization.md) 已 READY 並接入正式
首次規劃；固定種子原版比較及完整回歸通過，正常整場尚未 CONFORMED。

SPEC-030：[中文模式的畫布倍率與角色資料頁](030-cjk-character-sheet.md) CONFORMED；
六人購馬、坐騎與隊員格、物品及隊伍換頁的繁中畫格與英中狀態已有收據。
中文模式**不對拍**——原版沒有中文畫面。倍率取三是算出來的：中文字模格高 18，
而原版最小行距 6，`6 × 3 = 18` 剛好；兩倍只給得起 12 放不下。三倍之後版面照抄就
成立，資料頁直接用 `SheetLayout × CJKScale`，沒有第二份座標表。

SPEC-031：[資料頁物品選中後的明細層](031-item-detail-page.md) READY，Theodoric兩條正常序列與既有明細／保存切片CONFORMED；其他物品及動作仍待驗。
選中非空物品後，原版由 `sub_16986:0x16BBF` 先以 `sub_1103F(0)` 蓋整頁，再由
`sub_1BA9D` 畫明細框；`sub_1BA9D:0x1BAF1` 必定呼叫 `sub_1B7E9`，因此明細框
有右／下四像素投影。原frame-072維持DOSBox-X `dosbox-bootstrap`輔助分類；
2026-10-01以固定dosgolem正常啟動重生六張英文明細／空格頁，除了實際滑鼠外均零差。
空格位置、字串接續及確認返回已修；RE-172首批完整清單續畫有1／8像素差異，不稱父規格全通過。

第9節另補正常搬移／丟棄確認與保存切片：23完整明細／等待及返回頁除滑鼠外零差，
四繁中字模框及十二重載選中頁通過；[收據](../verification/item-detail-20261001/actions-receipt.json)。
RE-174當時76清單畫格未通過、動作後有692像素殘留；後續第10／13節已閉合這些差異，父規格維持READY。

第10節RE-175修正物品鏡像及立繪回退：29完整頁除滑鼠外零差、24份typed／raw保存
符合原版同期動作記錄，[新收據](../verification/item-detail-20261001/mirror-receipt.json)。
前批692像素殘留已消除；當時清單、手持文字及繁中立繪缺口由後續第11～13節閉合。

第11節RE-176補繁中Figure貼圖：78張正常清單與固定原版同期記錄＋PNG的獨立三倍
合成全零差，32其他繁中頁／110英文頁及24份完整持久資料不變，
[新收據與重驗入口](../verification/item-detail-20261001/cjk-figure-receipt.json)已接入規格。
前批繁中漏圖已閉合；當時英文清單及手持文字缺口由第12／13節閉合，其他造型／物品動作仍由#64追蹤。

第12節RE-177修手持技能列號共用座標：三個英文小框全零差、三繁中圖與獨立字模重建
一致，其他107圖各語言不變；24份持久資料不變，
[新收據](../verification/item-detail-20261001/mark-receipt.json)。前批三框文字差異已閉合，
當時70張英文清單只剩x15續畫差異，已由第13節閉合；其他技能列／造型及物品動作仍由#64追蹤。

第13節RE-178先證實快取步長與還原生命週期，RE-179再完成暫態呈現：104動作及
61張十七欄／繞回英文整頁除既定滑鼠外全零差，110繁中圖、24份持久資料及十二重載
明細頁不變，雙語狀態一致。[正式收據](../verification/item-detail-20261001/edge-receipt.json)
及[重驗入口](../verification/item-detail-20261001/verify_edge.py)另存，窄切片CONFORMED。
父規格維持READY；Theodoric這兩條序列的x15差異已閉合，其他人物／全部動作、法術／
戰利品及整場戰鬥仍依#64與既有閘門驗收。

SPEC-032：[戰鬥戰利品與法術清單的繁中版面](032-cjk-battle-list-pages.md) CONFORMED。
本規格只處理 Issue #90 指出的兩個 CJK 清單頁；原版沒有中文 oracle，中文畫面不作
原版像素 parity 宣稱。
