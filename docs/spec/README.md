# 規格索引

十一份規格：一份 CONFORMED、六份整份 READY、三份部分 READY（未解的段落標明
未開始），一份 DRAFT 不授權實作。READY 的授權範圍逐份限定，見下表。
所有原版衍生行為必須依序經過：

`RE 證據 → DRAFT → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED`

在 M1 executable 與資料身分完成前，不建立玩法規格。

| ID | 規格 | 狀態 | 是否授權實作 |
|---|---|---|---|
| SPEC-001 | [原版 Oracle 準備與位址契約](001-original-oracle-preparation.md) | READY | 僅授權研究 derivative，不授權玩法實作 |
| SPEC-002 | [PNG 素材抽取契約](002-png-asset-extraction.md) | READY | 僅授權已關閉像素格式的抽取器，其他 archive 仍須逐批驗證 |
| SPEC-002 | [EGA 抽取集合](002-ega-extraction-set.md) | READY | 僅授權本集合的 PNG 抽取與驗證 |
| SPEC-003 | [繁中說明書翻譯與校對契約](003-manual-translation.md) | CONFORMED | 僅授權文件翻譯，不授權玩法／runtime |
| SPEC-004 | [OGG 音訊目錄與執行期契約](004-ogg-audio-catalog.md) | DRAFT | 否；先完成 cue 母體與產品決策 |
| SPEC-005 | [角色建立與角色儲存區](005-character-creation.md) | READY | 僅授權本規格涵蓋的角色建立與刪除流程；16 槽全滿分支仍須標為未驗證 |
| SPEC-006 | [住宿、治療與存檔格式](006-rest-and-save.md) | READY | 僅授權 REST 三條路、remake 存檔讀寫與離線匯入 |
| SPEC-007 | [野外的時刻、日曆、天候與遭遇](007-wilderness-time.md) | 時刻／日曆／天候 READY；遭遇 DRAFT | 僅授權時刻、日曆、天候與紮營畫面 |
| SPEC-008 | [戰場的畫面、視線與回合順序](008-combat-field.md) | 畫面／視線／規劃順序／執行順序 READY；移動 READY；**戰利品與升級未開始** | 僅授權戰場繪製、視線、站位、兩段順序與移動 |
| SPEC-009 | [戰鬥的攻擊結算](009-combat-attack.md) | 近戰 READY；射擊、法術與敵人 AI DRAFT | 僅授權近戰攻擊的規則層，不含 UI 與 AI |
| SPEC-010 | [戰場上的指令與選目標](010-combat-commands.md) | 主指令列／選目標格／三列子選單 READY；背包用物品與五個未定語意的圖示未開始 | 僅授權指令列條件、選目標游標與選項換算 |

## 未開始的規格

這些是玩家會遇到、但**還沒有 RE 證據**因此連 DRAFT 都還不存在的行為，列在這裡
免得被當成「已經有草案」：

- 戰鬥打贏之後的戰利品、經驗值與升級（入口 `sub_289EE` 已定位，未解讀）
- 法術系統
- 任務的回報與完成、結局的觸發
- 野外的隨機遭遇**實作**（規則 RE-078 已解，尚未寫成 spec）
