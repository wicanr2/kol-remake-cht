# RE-191：返回暫存的追蹤邊界

日期：2026-10-02。正式程式仍為RE-187的`dea837bc`，成本仍DRAFT。
本輪有可丟棄Go追蹤副本，沒有將它接到正式路徑。
驗證入口為[verify_tracker.py](../verification/return-tracker-20261002/verify_tracker.py)。

## 固定來源與工具

| 項目 | 身分 |
|---|---|
| 原版 | KOLTOWN-unpacked.EXE，SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 正式IDA資料庫 | KOLTOWN-unpacked-ida94-derived.i64，SHA-256 `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`，未修改 |
| 靜態工具 | IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1`，image ID `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780` |
| 動態工具 | dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f` |
| Go工具鏈 | `eob-remake-go:1.26.7-ebiten2.9.9-audio`，沿用鎖版工具鏈 |
| 位址基準 | 指令及xref為IDA linear；BB84／BB85為DGROUP相對位移，對應原始word_39A54 |

IDA只在一次性資料庫匯出word_39A54的120個直接交叉參照及左右28bytes指令窗，
去重後1197條指令，正式資料庫338函式保持。120是參照數，不是寫入數。
匯出保留原始名字、位址、operand、bytes、已分級索引與未知警示；直接xref不能
排除間接寫入。初始化與世界方向處理只分類，不深入無關helper。

## 正常來源與隔離Go副本

| 來源 | 比較點 | 原版筆數 |
|---|---|---:|
| RE-188正常勝利 | `0x20BAD`負重乘數寫入後 | 11 |
| RE-190正常施法後敗北 | `0x20BAD`負重乘數寫入後 | 94 |
| RE-190兩次施法 | `0x20D1B`施法成本寫入後 | 2 |

107例全部由原版唯讀觀察建立，包含同值寫入，沒有只取變值監看的70筆當完整來源。
隔離副本在既有spendTime中依序追蹤乘數及CastTime覆寫，105個中間值與兩個720
都符合原版。這是Go函式與原版觀察點的比較，沒有驗整輪、GUI、世界成本或保存。
副本的ReturnScratch與ScratchWrites只在容器暫存State中，不是正式資料格式。
314個cmd／internal／tools輸入及正式程式雜湊保持。

首版把變值筆數誤當案例數，83筆斷言在啟動Go前失敗。第二版漏設Actor.Member，
使Warder取到錯誤隊員，施法覆寫未執行。修正後用同一批107個來源、乾淨副本重跑
通過。兩次失敗輸入／log保留，屬準備與夾具問題，不是產品規則缺陷。

## 正常法術規劃後逃走

預登記單一路徑：書本正常選Warder，旅店招募Olear，兩人首回合丟弓；第二回合
Warder只規劃原有Vartwelyrmi的Negative，Olear休息，從Begin next round的逃走格離開。
種子在旅店招呼後及首次規劃前固定為1，DS:1600=1、1602=0；沒有換種子、重排或
依結果改策略。沒有角色、法術、座標、指令注入、direct-entry或等待替換。

原版實際只執行第一回合，第二回合在規劃後逃走。43個按鍵、28個出口前亂數與
八次原生等待完成返回。兩人手空且未倒地，法術沒有執行；不能稱為施法生效驗收。
角色+0x25..27的`0d0001`證實已規劃第一筆法術Negative。

| 原始寫入／觀察 | 已證實結果 |
|---|---|
| `0x1F502`，bytes `a384bb` | 清單回傳寫BB84；本場觀察到`0x0700`，不是法術成本 |
| `0x1F60D`，bytes `a384bb` | 正負清單回傳寫BB84；本場Negative留下`0x0701` |
| `0x1F186`，bytes `c70684bb0000` | 主列重畫先清為0；`0x1F18C`變值觀察確認兩個byte都清除 |
| `0x1F1E4`，bytes `830684bb01` | 有目標標記時加1；本場Warder主列暫為1 |
| 最後Olear主列 | 先清0，後選REST；本場最後BB84=0，沒有保留`0x0701` |
| 世界`0x29800..29823` | 索引0讀byte10，移動率20，成本60；3360降3300，次格4,4 |

本場天候前種子1226606285，天候後1292271196，扣時刻仍在天候後。
首版對逃走原生字元等待送Return，4500鍵上界後停止，未返回；第二版只改回答y，
並收緊上界120，正常返回。首版完整payload／畫格保留，不選取其中片段當通過收據。

此結果證明本路徑的選單殘值會被主列重畫覆寫。不能從單一選單回傳推定世界成本，
也不能把所有清單回傳當索引乘256。沒有外推全部規劃取消、滑鼠輸入或所有逃走。

## 下一個實作閘門

成本追蹤應保持原始word的16位元暫存身分，按真正執行階段寫入。已證實來源包括
主列重畫、執行排序、成功掉落、回合疲勞／施法、持任務物計數及搜刮清槽。
正式垂直鏈仍須把這些寫入與世界延遲扣時刻串起來，驗三出口、後續移動與保存。
法術清單的中間值已證實在本場被覆寫，不深挖無關文字helper。

表外成本的動態讀值仍未完整契約化；RE-189的840亂數讀值不能改成固定byte。
本輪不升READY、不把隔離副本接正式路徑、不以Verdict或靜態終態固定成本。
父SPEC-011／027、正常勝利GUI及完整進場亂數限制保持。

## 封存入口

工作區`workplace/return-tracker-20261002`，入口為preserve_parent.py、ida_tracker.py、
tracker_cases.sh及menu_flee2.sh。1340份共同輸入先保存，RE-190原3094附件與下層
歷史保持；原版副本、IDA payload、Go副本及新畫格僅本機。

在既有容器將專案唯讀掛到`/src`，執行
`python3 /src/docs/verification/return-tracker-20261002/verify_tracker.py`重驗。
第一次封存用`--seal`及唯一輸出`/out/receipt.json`；完整撤去可寫掛載後用`--audit`。
收據為[receipt.json](../verification/return-tracker-20261002/receipt.json)，
批末稽核為[audit.json](../verification/return-tracker-20261002/audit.json)。
