# 完整自然戰鬥的執行段續核（RE-180）

本輪完成兩處已證實規則修正及 61 個執行段同狀態比較。原版從正常選角、招募、旅行
自然遇到 SLEDGE，未替換等待常式；重製版逐回合從原版入口接手，所以不把本結果
當成完整正常 GUI、規劃段連續對拍、戰後回寫或保存鏈完成。

## 輸入與工具

| 輸入 | SHA-256 |
|---|---|
| Knights-of-legend.zip | `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` |
| KOLTOWN.EXE | `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221` |
| KOLTOWN-unpacked.EXE | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| KOLTOWN-unpacked-ida94-derived.i64 | `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |

原版執行器為 dosgolem `968ca774705853276ae55c8d1163c369481ac71f`，Go 1.26.7；
正常流程在旅店招呼後及首次規劃入口將 DS:1600／1602 設為 1／0。除此測試種子設定，
不注入角色、指令或座標。351 次真實按鍵、350 次等待，61 回合到 IDA 線性位址
`0x211AD` 的戰鬥出口。記錄包含 122 邊界、1458 次亂數呼叫及 1318 則訊息。

IDA Pro 9.4 沿用正式資料庫的暫存副本。所有函式、指令、運算元、bytes、xref
保持原始定位；新增七筆[分級索引](battle-order-address-index.json)由匯出工具自動合併，
未知定位仍帶警示。IDA 位址採 `IDA linear`；原版記錄採 DGROUP 相對位移，
執行期呼叫位址依固定 DGROUP 基準轉換，兩種空間不混用。

容器沿用 `kol-package:20260907`、`eob-remake-go:1.26.7-ebiten2.9.9-audio` 與
`ida-pro-9.4-idapython:locked-v1`。原始輸入／正式資料庫唯讀；一次性容器設 UID/GID
1000:1000、資源限制、逾時及無網路。完整 payload、原版 EXE／資料副本只留本機。

## 已證實差異與勘誤

1. 近戰選項 1：原版在 `sub_26D76:0x26E8E` 先擲 `rand(100)`，再在 `0x26FF0`
   擲 `rand(3)` 選列。舊 Resolve 反過來；第 15 回合才首次揭露，舊六回合不能排除此錯誤。
   [SPEC-009](../spec/009-combat-attack.md) 第 3 節本來就有正確順序，先續核 READY 再修正。
2. 怪物負重：`sub_24FD3:0x24FEA` 讀參戰者 `+0x02`，係數為
   `70-(等級-1)*5`；它沒有按陣營改讀遭遇等級。正常 SLEDGE 遭遇等級 3、參戰者等級 1，
   重量 50、護甲 100、無盾，實際負重 87／速度 35。舊 NewQuest／LoadLevel 是 75／30，
   第 22、34 回合改變出手先後。舊註解的不同尺度解釋由直接資料流與 183 次唯讀觀察否定；
   保留本段勘誤與失敗日誌，正式攻防技能及獎勵仍按原本遭遇等級。
3. 夾具轉換：第 24 回合 Olear 頭／胸原始 `0xFF` 表示倒地，正式 Actor 使用 0。
   只對六個部位轉換，並保留原始 byte；其他欄位不轉換。這是比較輸入問題，沒有更改
   正式健康表示或存檔格式。

原版 original2 新增負重觀察後，122 邊界、351 輸入及 1318 訊息與 original1 完全相同；
1458 次亂數呼叫除新增觀測欄位外相同。前六回合的種子／18 名原始記錄／計數器亦與
RE-158 歷史收據相同，沒有覆寫舊收據。

## 實作與驗證邊界

正式修正在 `internal/battle/attack.go`、`state.go` 與 `battle.go`；夾具在
`cmd/kol/original_rounds_test.go`。兩個新回歸均先在未修程式上拒絕實際舊行為。
原版每回合入口的 32 位元亂數狀態交給重製版同一亂數實作，比較每一次骰面順序、
結束種子、位置、六部位、累積傷害、挨傷、招架、方向及疲勞，61 個執行段全部通過。
其餘欄位及完整規劃流程沒有因這個比較而升格。

本輪建置 `b1264c4e12452cad59ed015bab02bd339c967b3da41bbab037019faf6d28f058`，
本機 `workplace/battle-complete-20261001/kol-battle-rules-linux`。
完整內部／cmd 測試、104 圓環框及舊正常五次 AI／六執行段抽樣通過。
證據入口為[正式收據](../verification/battle-complete-20261001/receipt.json)與
[唯讀重驗工具](../verification/battle-complete-20261001/verify_rounds.py)。
新程式雙語各110正常物品畫格與110筆狀態保持RE-179結果；
[收尾稽核](../verification/battle-complete-20261001/audit.json)另鎖定重驗程式與正式證據。

本機重生入口都在 `workplace/battle-complete-20261001/`：`original_complete.sh`、
`prepare_original.py`、`ida_battle_order.py`、`compare_rounds.sh`、`check_rules.sh`、
`check_normal_items.py`。原版停在出口入口，未執行戰後回寫；V01 與工作清單第 6 項
保持開放，不宣稱整個 SPEC-009 或完整 remake 已完成，也未打包／發布。
