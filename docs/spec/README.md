# 規格索引

目前有七份 READY 規格、一份 CONFORMED 規格與一份不授權實作的 DRAFT 規格。READY 的
授權範圍逐份限定，見下表。所有原版衍生行為必須依序經過：

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
| SPEC-008 | [戰場的畫面、視線與回合順序](008-combat-field.md) | 畫面／視線／規劃順序 READY；指令、執行順序與命中 DRAFT | 僅授權戰場繪製、視線、站位與規劃順序 |
