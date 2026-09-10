# 怪物參戰 byte +0 初始化訂正

狀態：初始化分支已證實；非零值轉換語意仍待查。日期：2026-09-08。

輸入 `KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
工具 IDA Pro 9.4，位址空間為 IDA 線性；DGROUP 基底 `0x2DED0`。
原始資料庫使用副本，未改名或改寫來源。匯出保留原始位址、bytes、函式名及分級，
並自動合併既有受審索引。完整原始指令在 `workplace/monster-kind-20260908/monster-kind-private.json`。

## 已證實

`sub_1E3A8` 的 `0x1ED38..0x1EDD9` 逐名處理敵人：

- `0x1ED69..0x1ED78` 以本場怪物記錄索引乘 128，加 DS:CD92，讀 `[bx+1Eh]`。
- `0x1ED85` 呼叫 `sub_1B5A8(100)`；`0x1ED90..0x1ED94` 以無號 `JA` 判斷：原值必須嚴格大於骰值。
- 通過時 `0x1EDA6` 讀 `[bx+1Fh]`，`0x1EDAD` 呼叫 `sub_222BF`，`0x1EDC4` 將返回低 byte 寫到參戰者 `+0`。
- 未通過時 `0x1EDD4` 寫參戰者 `+0 = 0`。

因此「敵人參戰 +0 就是怪物種類索引」已被寫入端推翻。非零值經 `sub_222BF` 轉換，
尚不命名其完整語意或推測分布。既有 `TestBattleEnemyShootRingOriginal` 已確認原版射擊
分支在環號 <4、該 byte 非零、右臂可用時通往射擊出口。

正常 SLEDGE 快照的怪物記錄 `+0x1E=0`、參戰 `+0=0`，符合必走零值分支。
來源為 `workplace/battle-state-records-20260908/` 的 dosgolem 正常玩家入口收據，
不是以直接呼叫或強制改值產生的正常畫面。

## 正式實作缺口

`internal/battle/state.go` 目前以 `site.Monster` 填入敵人 `Actor.Kind`；
`internal/battle/ai.go` 的射擊資格直接判斷 `a.Kind != 0`，因而可能給予原版沒有的射擊資格。
`equipment_panel.go` 同時將 Kind 用於怪物顯示，必須分開核對用途，不能直接把全部怪物改成 0。
前輪對齊測試排除該欄位的限制保留，不能以 59 欄通過掩蓋此正式玩法缺口。

下一步：查 `sub_222BF`，固定兩邊測試種子驗證機率／轉換邊界，補 READY 規格後接入
正式建構及顯示，並回到正常戰鬥比較。地形追查本輪尚未展開。
## 同日追加：轉換常式已驗證

`sub_222BF` 呼叫 `sub_22376(code,5,3)` 取 `(code>>3)&7`，
再以 `sub_22376(code,2,3)` 取 `code&7`，交給 `sub_22319`。
後者以第一個值加一作擲骰次數，骰面讀 DS:00D4 的 `2,3,5,7,9,11,19,99`，
每次加總只寫回低 byte，故總和模 256。原始 IDA 線性定位：
`0x222D3`、`0x222E9`、`0x22302`、`0x22332`、`0x22350`、`0x22360`。
工具、輸入與位址基準同上；完整匯出在 `workplace/monster-conversion-chain-20260908/`。

`TestMonsterRangedConversionOriginal` 實際呼叫原版 `sub_222BF`，256 種代碼 ×
固定種子 0、1、65535，共 768 組返回值及種子消耗相符。未替換原版 RND 常式，
沒有追求跨正常遊玩流程的自然亂數序列。收據在
`workplace/monster-conversion-oracle-20260908/`，同批 4608 組傷勢及 258 組任務物品回歸通過。
原先「轉換語意待查」由本證據訂正；初始化機率閘門的執行對拍與顯示用途仍待完成。
## 同日追加：初始化機率閘門通過

`TestMonsterRangedGateOriginal` 實際執行 IDA `0x1ED38..0x1EDD9`，僅以入口橋建立
區域變數堆疊、在下一階段前返回；未替換 `sub_1B5A8` 或 `sub_222BF`。
門檻 0…255 × 代碼 0／0x3F／0xFF × 固定種子 0／1／65535，共 2304 組通過。
比較內容包含實際寫入參戰 byte +0 與最終種子，故涵蓋：
門檻必須嚴格大於 RND(100)、失敗覆寫 0、每次必擲機率骰、只在通過時擲轉換骰。
結果位於 `workplace/monster-gate-oracle-20260908/battle-exit.log`，來源清單同目錄保存。
本次只驗單名敵人的初始化區段，不宣稱完整正常戰鬥或怪物顯示已通過。
## 同日追加：面板取圖與能力分離

IDA `sub_22EFC` 的 `0x22F3B/0x22F60/0x22F77/0x22FB3` 讀取 `word_3B262`，
作尺寸表索引並乘二選取成對 MONPICS payload，未讀參戰 byte +0。
原始匯出在 `workplace/monster-panel-20260908/monster-kind-private.json`，輸入與工具基準同上。
SPEC-019 READY 增補據此修正正式面板為 `State.Monster.Index`；能力 0、1、99、255
均保留同一肖像。完整測試與編譯通過，輸出在 `workplace/monster-portrait-fixed-20260908/`。
首輪新測試漏設 StageField、舊測試漏填獨立索引的失敗保留於 `monster-portrait-fix-20260908/`。
這完成顯示用途分離，能力初始化正式接線仍未完成，AppImage 尚未包含本輪修正。
