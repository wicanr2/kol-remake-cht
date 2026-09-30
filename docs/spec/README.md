# 規格索引

2026-09-08 現況訂正：共有 26 個 ID、27 份規格；最新實作與剩餘缺口以
[本輪稽核](../verification/spec-source-audit-20260908.md) 為準。
SPEC-019 的彈藥排除由 SPEC-022 補充；SPEC-020 已實作，SPEC-022 彈藥畫面與
Olear 正常射擊已驗。SPEC-023 已補 v1 匯入、建角、住宿交易及世界時間。
下表的早期授權限制不可直接當成目前未實作清單；本次不自動升 CONFORMED。

規則規格（005～029）末尾都有「說明書對照」表，連回說明書 PDF 頁碼與繁中章節。
規格狀態以下方登記與各文件的最新審查為準。2026-09-07 原始碼稽核
確認多份索引／本文保留過時的未實作或禁作條款，不能直接當成現行工作清單。
逐份實作落點、缺口與回填證據見
[原始碼／規格稽核](../verification/spec-source-audit-20260907.md)。
以下保留既有授權登記供回查；本次不自動升降 READY／CONFORMED，矛盾條款須逐項審查。
所有原版衍生行為必須依序經過：

`RE 證據 → DRAFT → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED`

在 M1 executable 與資料身分完成前，不建立玩法規格。

| ID | 規格 | 狀態 | 是否授權實作 |
|---|---|---|---|
| SPEC-038 | [套件邊界、固定步進與注入點](038-package-boundaries.md) | READY | remake 工程契約；`internal/layers` 四項守門測試擋依賴方向、Ebitengine 越界、規則層讀時鐘與全域亂數 |
| SPEC-037 | [建築進門後的發言者選擇與好惡](037-building-greeting.md) | READY | 多人隊伍先選發言者、被討厭與被欣賞的訊息、全隊被趕、SCROLL 教派；原版關鍵點三組共 36,176 組一致；滑鼠點肖像見 RE-167 |
| SPEC-036 | [圖像與盾牌編輯器](036-picture-shield-editor.md) | CONFORMED | 點陣圖編輯器已實作；鍵盤路徑與原版同狀態逐 byte 一致，滑鼠路徑依靜態讀碼、未與原版實跑比較 |
| SPEC-035 | [法術編輯器的繁中頁面](035-cjk-spell-editor.md) | READY | 中文模式原本看不到編輯頁，已補；四張正常按鍵畫格已驗，選值寫回與英中同狀態待驗 |
| SPEC-034 | [法術名的中文注解](034-cjk-spell-name-gloss.md) | CONFORMED | 法術名保留原文，依說明書音節表加注解；48 筆法術記錄交叉驗證，三處正常按鍵畫格已驗 |
| SPEC-033 | [中文畫面的資料名稱顯示](033-cjk-display-names.md) | READY | 491 個來源原名已有暫譯並接顯示端（含所在地、城鎮代碼、教派與 MAGIC.0 亞種）；待英中同狀態與正常玩家畫格驗收 |
| SPEC-031 | [資料頁物品選中後的明細層](031-item-detail-page.md) | CONFORMED | 選中物品頁的圖層與投影已完成；frame-072 是 DOSBox-X 輔助收據，非 strict dosgolem parity |
| SPEC-032 | [戰鬥戰利品與法術清單的繁中版面](032-cjk-battle-list-pages.md) | CONFORMED | 四種清單畫格與正常按鍵路徑已驗，戰利品採明示觀察用資料；不宣稱原版中文或整場戰鬥 parity |
| SPEC-030 | [中文模式的畫布倍率與角色資料頁](030-cjk-character-sheet.md) | CONFORMED | 六人與坐騎的正常按鍵畫格、物品與隊伍換頁及英中語意狀態已驗；中文模式不對拍，英文嚴格對拍那一版不動 |
| SPEC-029 | [怪物遠程能力初始化](029-monster-ranged-initialization.md) | READY | 遠程機率與轉換骰已接入首次規劃；正式實作 768＋2304 組通過 |
| SPEC-028 | [三個可關閉的加速選項](028-optional-battle-speed.md) | DRAFT | 已決需求與現有接口已核對；細節／原型待確認，不授權先加空開關 |
| SPEC-027 | [戰鬥共同出口傷勢回寫](027-battle-exit-wounds.md) | READY | 傷勢折算及敗北回寫已接；4608 組原版比較、五條結算至 v2 重載通過，正常玩家鏈與額外掉落待驗 |
| SPEC-026 | [天候圖層與正常動畫](026-weather-animation.md) | CONFORMED | 六組資料、128 像素捲雲與雨雪輪播；原版四組合成、六組分派、節拍與相位逐次對拍通過（RE-164） |
| SPEC-024 | [建築告別與確認](024-building-goodbye.md) | READY | 單人直接退出、多人告別確認已實作；正常原版與繁中 GUI 已抽驗 |
| SPEC-025 | [競技場固定地形](025-arena-field.md) | READY | 固定底圖及繁中對話已接；4096 格原版比較、地圖隔離與受控資格 GUI 進場通過，正常取得資格與結算鏈仍待驗 |
| SPEC-023 | [完整角色、名冊與續玩保存](023-complete-save.md) | READY | 完整 v2、v1 匯入、建角、住宿交易及世界時間已接線；保存與 GUI 驗證見本輪收據 |
| SPEC-001 | [原版 Oracle 準備與位址契約](001-original-oracle-preparation.md) | READY | 僅授權研究 derivative，不授權玩法實作 |
| SPEC-002 | [PNG 素材抽取契約](002-png-asset-extraction.md) | READY | 僅授權已關閉像素格式的抽取器，其他 archive 仍須逐批驗證 |
| SPEC-002 | [EGA 抽取集合](002-ega-extraction-set.md) | READY | 僅授權本集合的 PNG 抽取與驗證 |
| SPEC-003 | [繁中說明書翻譯與校對契約](003-manual-translation.md) | CONFORMED | 僅授權文件翻譯，不授權玩法／runtime |
| SPEC-004 | [OGG 音訊目錄與執行期契約](004-ogg-audio-catalog.md) | DRAFT | 否；先完成 cue 母體與產品決策 |
| SPEC-005 | [角色建立與角色儲存區](005-character-creation.md) | READY | 建角、刪除、穩定槽位與保存已接；重製版滿槽 GUI 已驗，原版滿槽對拍及圖像編輯仍未完成 |
| SPEC-006 | [住宿、治療與存檔格式](006-rest-and-save.md) | READY | 僅授權 REST 三條路、remake 存檔讀寫與離線匯入 |
| SPEC-007 | [野外的時刻、日曆與天候](007-wilderness-time.md) | READY | 僅授權時刻、日曆、天候與紮營畫面；遭遇見 SPEC-011 |
| SPEC-008 | [戰場的畫面、視線與回合順序](008-combat-field.md) | READY | 含戰場、回合、獎勵與 AI 主幹；敵人拾取已依 RE-116 接線，卡住重試仍有有界近似，完整 AI 正常路徑待驗 |
| SPEC-009 | [戰鬥的攻擊結算](009-combat-attack.md) | 近戰與**擊退／打暈** READY | 僅授權近戰攻擊與擊退的規則層，不含 UI |
| SPEC-010 | [戰場上的指令與選目標](010-combat-commands.md) | READY；施法／射擊另依 012／013 | 指令列、游標、物品動作、下馬及任務物品已接線；未驗收分支見本文 |
| SPEC-011 | [野外的隨機遭遇](011-wilderness-ambush.md) | READY | 含 RE-112／128 治療問答及敘述；正常治療與保存往返仍待驗 |
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

SPEC-031：[資料頁物品選中後的明細層](031-item-detail-page.md) CONFORMED。
選中非空物品後，原版由 `sub_16986:0x16BBF` 先以 `sub_1103F(0)` 蓋整頁，再由
`sub_1BA9D` 畫明細框；`sub_1BA9D:0x1BAF1` 必定呼叫 `sub_1B7E9`，因此明細框
有右／下四像素投影。實作與收據已完成，frame-072 允許區外差異為 0；dosgolem
目前沒有 KOL 適配器，因此該畫面只作 DOSBox-X `dosbox-bootstrap` 輔助收據。

SPEC-032：[戰鬥戰利品與法術清單的繁中版面](032-cjk-battle-list-pages.md) CONFORMED。
本規格只處理 Issue #90 指出的兩個 CJK 清單頁；原版沒有中文 oracle，中文畫面不作
原版像素 parity 宣稱。
