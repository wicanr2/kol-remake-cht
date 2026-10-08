# RE-197：徒手攻擊的分支與空槽傷害骰

日期：2026-10-03。入口為[目前狀態](../../CONTEXT.md)與
[SPEC-009](../spec/009-combat-attack.md)。研究工作區為
`workplace/unarmed-resolution-20261003/`，原始輸入與正式資料庫唯讀。

驗證入口為[有限唯讀重驗](../verification/unarmed-resolution-20261003/verify.py)，
正式保存鏈與批末稽核結果在同目錄`receipt.json`與`audit.json`。

## 來源

| 項目 | 身分 |
|---|---|
| 原版 | `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 正式資料庫 | `workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64`，SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 工具 | IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1`，image SHA-256 `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780` |
| 新匯出 | `ida_unarmed.py`／`ida-unarmed.json`，一次性資料庫查詢 `sub_26AC2`、`sub_26D76` 的1045道指令，保留名稱、bytes及交叉參照 |
| 正常原版來源 | RE-186的`source4/complete-source-private.json`，dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f`，首次規劃seed 1，正常147鍵、4回合勝利 |
| 重製反例 | RE-196正常雙語勝利，Olear第4回合指令41、選項1；兩筆骰面100／2變成3／100，終態LastHurt 25變23 |
| 保存 | 前輪5343附件全部核對，1375份可能異動的共享輸入另存`parent-inputs/`，舊收據與程式不覆寫 |

位址均為IDA linear；DGROUP基準為`0x2DED0`。參戰者記錄位移與DS位移另列，
不與IDA位址混用。IDA程序離開碼1，但輸出JSON存在、schema、版本、輸入雜湊及
函式邊界正確，1045道指令已讀回。原始匯出只留本機。

## 已證實的控制流

| 原始定位 | 證據與語意 |
|---|---|
| `sub_26D76:0x26FCA..0x26FDA` | `cmp word_31122,29h`後徒手直接將攻擊骰分數存入`word_2FB46`並跳`0x27079`，不走選列、選項加分或武器遮罩減半 |
| `0x26FE7..0x26FF0` | `rand(3)`僅在前述跳轉未走、選項1時執行，是持械分支 |
| `sub_26AC2:0x26BDA..0x26BF9` | 隊員徒手依參戰者選項`+0x23`查DS:0x0FD2，再加角色槽位移0x100；選項1拿FEET槽 |
| `0x26C2D..0x26C40` | 不檢查物品名稱或槽是否空，直接讀所選槽`+0x0C`傷害碼 |
| `0x274D6..0x27518`、`0x27556..0x27570` | 傷害碼0仍算1個2面骰，再進正常傷害迴圈；全零物品不代表略過擲骰 |
| `0x27579..0x2758E` | 攻擊式傷害加成只限指令0x21；徒手0x29不加表DS:0x1200 |
| `0x27648..0x27665` | 命中部位列先設1，只有指令0x21才查攻擊式表；徒手固定列1 |

正常來源第4回合索引73在IDA返回位置`0x2736B`擲100，索引74在`0x2755E`擲2。
重製版Resolve未區分徒手便呼叫ModifierRow，額外擲3；AttackWeapon回傳零值後
傷害迴圈次數0，漏掉2面骰。多一骰與少一骰恰好抵消次數，故100個seed_before仍
一致，骰面卻不同。空槽只指本場及零值Weapon表示，不由名稱空白推定所有原始
16-byte槽全零。

以上先形成READY契約；後續正常重播確認LastHurt25，結果見下節。
停止線為修正上述已證實分支及其正常玩家鏈；不追無關helper或完整亂數內部。

## 正式修正與驗收

證據審查先升SPEC-009勘誤為READY，再改AttackWeapon、Resolve及傷害／部位的
徒手分支。正式318碼輸入只改attack.go、既有attack_test.go及main.go的錯誤註解。
Go 1.26.7／Ebitengine 2.9.9，既有音訊建置image；未改存檔、資產、UI或速度契約。
新正式候選SHA-256為`ca2ff7bfad06aced4dec33b9876d480696b5bbd4a73442134a24ec80dcf5a5d7`，
觀察副本為`d2a6fc45a65139bddcf7be81c8b213e4b1aa03e7a088450d2d7b3176b80f5433`。

| 驗收 | 結果與範圍 |
|---|---|
| 規則 | 舊版負對照在第2骰多擲3失敗；四式徒手、落空、破甲空槽及既有持械骰序通過，全部internal／cmd通過 |
| 正常勝利 | 雙語各269鍵／270狀態與PNG，全部270張PNG逐byte保持RE-196已目視畫面；四個原版執行入口相符 |
| 本場亂數 | 99筆戰中／報酬與1筆天候的n、seed_before全部相符；沒有完整進場亂數或完整角色記錄的聲明 |
| 本場終態 | 既定角色抽樣零差異，LastHurt25；原兩筆骰面與LastHurt25／23反例閉合 |
| 返回／物品 | 世界49,23、次格4,4、時刻3300；34個完整16-byte槽、戰鎚、金額1835／1067及冒險點數2891／749符合原版 |
| 新保存鏈 | 雙語各507正常鍵、508狀態及27張檢查圖；v3保存保留合法3360／0、16名冊槽與實際住宿前34槽 |
| 正式重載 | 新正式候選及觀察副本各16狀態／PNG相同，正常招募Olear，完整物品與持久欄位還原 |

回城另一次正常逃走仍只改Theodoric手持大劍槽，保存金額1779／1067。回城可玩性
只驗重製，不外推原版全路線；typed物品只補town.Good省略的picture／movement零值。
所有原版payload、PNG、觀察程式與保存僅本機。

## 回填與勘誤索引

不可變鍵為DOS＋上述固定KOLTOWN-unpacked.EXE＋IDA linear位址。本表只附語意，
不覆寫原始名稱；推論等級均為已證實。自動重驗要求舊規格出現RE-197勘誤標記。

| 原始定位 | 語意 | 證據／舊規格 | 必要勘誤標記 |
|---|---|---|---|
| sub_26D76:0x26FCA | 徒手跳過選项修飾 | 本文及SPEC-009 | RE-197：有限CONFORMED |
| sub_26AC2:0x26C2D | 所選槽直接讀傷害碼 | 本文及SPEC-009 | RE-197：有限CONFORMED |
| sub_26D76:0x27579 | 徒手不加持械傷害表 | 本文及SPEC-009 | RE-197：有限CONFORMED |
| sub_26D76:0x27648 | 徒手命中部位固定列1 | 本文及SPEC-009 | RE-197：有限CONFORMED |
| 正常第4回合0x2736B／0x2755E | 兩筆骰面與LastHurt差異閉合 | 本文及SPEC-011／027 | RE-197：本場徒手差異閉合 |

初次交叉驗證直接比較IDA載入bytes與檔案bytes，首個far call因MZ relocation失敗。
依檔案重定位表加分析載入segment0x1000後，同一命令乾淨重跑，1045道指令全部
相符；原始與載入bytes另列於evidence-review.json。這是驗證工具的位址處理勘誤。

父規格與工作清單第6／44項仍開放；下一步正常逃走GUI與其餘玩家功能。
批末完整唯讀重驗、擁有權與Docker／Git控制面另見本輪audit.json。
