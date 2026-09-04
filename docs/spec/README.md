# 規格索引

十三份規格：一份 CONFORMED、八份整份 READY、三份部分 READY（未解的段落標明
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
| SPEC-007 | [野外的時刻、日曆與天候](007-wilderness-time.md) | READY | 僅授權時刻、日曆、天候與紮營畫面；遭遇見 SPEC-011 |
| SPEC-008 | [戰場的畫面、視線與回合順序](008-combat-field.md) | 畫面／視線／兩段順序／移動／訊息／勝負判定／**冒險點數、金冠與戰利品** READY；升級與敵人的撤退、反制掃描未開始 | 僅授權戰場繪製、視線、站位、兩段順序、移動、這一輪的勝負收尾與戰後的獎勵、戰利品 |
| SPEC-009 | [戰鬥的攻擊結算](009-combat-attack.md) | 近戰與**擊退／打暈** READY | 僅授權近戰攻擊與擊退的規則層，不含 UI |
| SPEC-010 | [戰場上的指令與選目標](010-combat-commands.md) | 主指令列／選目標格／三列子選單／**丟、撿、切換** READY；施法與投擲未開始 | 僅授權指令列條件、選目標游標、選項換算與三手物品動作 |
| SPEC-011 | [野外的隨機遭遇](011-wilderness-ambush.md) | READY（遭遇圖的版面與任務怪那一條除外） | 授權擲點、觸發、站位、地形與開打；遭遇圖版面與任務怪那一條除外 |
| SPEC-012 | [施法](012-combat-spells.md) | READY（效果 5 的疲勞鏈、效果 7／8 與選部位的互動除外） | 授權擲骰、擋招、效果、倒數、資料抽取與兩張清單 |
| SPEC-013 | [射擊](013-combat-missiles.md) | READY（投擲 0x22 與飛行動畫除外） | 授權彈道、目標選擇、攻防與訊息，以及敵人 AI 的射擊分支 |
| SPEC-014 | [任務的完成與逃走](014-quest-completion.md) | READY（掉東西與競技場的升級除外） | 授權旗標的標記、逃走的條件與那一問 |

## 未開始的規格

這些是玩家會遇到、但**還沒有 RE 證據**因此連 DRAFT 都還不存在的行為，列在這裡
免得被當成「已經有草案」：

- 角色升級（`0x21611` 那一條）。RE-103 的 operand 掃描指出 `word_345DE` 只有競技場
  `sub_2AAA8:0x2AFE6` 會寫進真值，所以它屬於**競技場**那一側；升級本身是角色記錄
  `[+0x29] += 1`，等競技場那一輪再寫
- 逃走或落敗之後掉東西那一段（`0x21765`），要先有裝備欄十二格的 typed data
- 投擲（指令碼 0x22）——`sub_26D76` 裡沒有一處比它，行為還沒有原版證據（SPEC-013 第 1 節）
- 法術的**選部位**那一次互動，以及效果 5 的疲勞鏈（SPEC-012 第 8 節）
- 任務的回報與完成、結局的觸發
