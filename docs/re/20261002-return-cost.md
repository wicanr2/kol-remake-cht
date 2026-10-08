# RE-188：返回成本的最後寫入與原始表讀取

日期：2026-10-02。正式程式保持RE-187的dea837bc，本輪補成本資料流證據。
入口為[唯讀重驗](../verification/return-cost-20261002/verify_cost.py)。

## 固定來源

KOLTOWN-unpacked.EXE SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro9.4，IDA linear；正式資料庫SHA-256
`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
dosgolem revision `968ca774705853276ae55c8d1163c369481ac71f`，正常seed1、147鍵第四回合勝利。
新原生來源只加指令位址觀察，完整JSON移除cost_flow後與RE-187逐項相同，
world-return-private.json逐byte相同。沒有換種子、改按鍵或替換等待。

## 已證實的因果順序

| 定位 | 限定結論 |
|---|---|
| IDA linear 0x20BA5／20BAA，bytes `9a34853a1e`／`a384bb` | sub_268D4的負重乘數回傳寫入word_39A54；0x20BAD觀察到已寫入結果 |
| IDA linear 0x20AF0..20B1F | Hurt為0且指令REST時跳過乘數計算；不能從最後角色靜態欄位猜測每手都寫入 |
| IDA linear 0x26A65，bytes `a384bb` | 成功掉落時把選中的地上槽號寫入word_39A54；本場第4回合觀察值0 |
| IDA linear 0x1FD9A／1FE50 | 執行順序準備也寫入這個暫存；本場準備尾值7，之後仍被掉落與疲勞覆寫 |
| 正常來源，機器step1223232645／1223233296 | Olear、Theodoric先後在0x20BAD寫1；真正最後寫入是Theodoric，前輪變值監看只能看到Olear那一次 |

OnCall記錄39個指定點；其中11次0x20BAD的角色、指令、式、Guard及Flag1E直接轉
Actor後，既有正式encumbranceMultiplier全部等於原版。包括步行2、快跑6、慢跑3、
特殊指令1與最後徒手／武器攻擊1。這是正常來源到正式規則的比較，不只檢查公式自洽。
沒有因此宣稱整個返回暫存追蹤已實作。

觀察欄actor保留DS:9E5A當下值。0x1FD9D尚未更新目前角色，掉落函式也可能處理
另一個arg_0角色；這兩處不把actor欄冒稱實際受處理的角色。0x20BAD角色已由
正式回合收尾資料流確認。WatchWrites只保存變值，不能當作所有賦值的完整序列。

## 成本讀取範圍

用既有tools/read_dgroup_table.py的唯一錨定位，再以兩組各72個站位byte作正對照。
DGROUP在檔案中起點0x202D0；DGROUP 0x12B7對應file offset 0x21587。
保存18個byte，對應返回索引−1..16：

`0,10,10,10,20,6,14,30,8,31,28,31,30,31,30,31,31,30`。

0..7是地形成本表範圍；8..16是表後的實際資料，這裡只證實被讀取的byte，
不為相鄰資料猜補語意。索引1讀10、16讀30，移動率20時分別為60、100。
不能把所有非方向索引夾到−1..1，或將表外一律當同一成本。

## 下一閘門

成本子切片仍DRAFT。正式追蹤須依實際寫入順序，涵蓋休息跳過、掉落、
回合末逃走計數及搜刮清槽；不可只按Verdict或最後Actor推定。正常敗北／勝利／逃走
與未進戰鬥路徑都要驗，並保持RE-187戰後天候順序及存檔契約。
沒有新增runtime資料格式、回呼或保存欄位，本輪不改production path。

1326份RE-187共享輸入先保存於workplace/return-cost-20261002/parent-inputs。
來源與IDA輸出、原始payload、新PNG及研究測試只留本機；舊來源與收據不覆寫。
重生入口為同工作區source_flow.sh、prepare_cases.py及fatigue_cases.sh，
都在既有Docker image執行。收據由verify_cost.py保存，清理稽核見
[audit_cost.py](../verification/return-cost-20261002/audit_cost.py)及同目錄audit.json。
