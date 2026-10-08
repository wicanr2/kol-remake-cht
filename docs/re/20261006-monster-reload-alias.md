# RE-264：怪物射後裝填別名的正常觀察

狀態：本場唯讀觀察通過；完整射後裝填別名仍待驗。

沿用 RE-260 的固定 seed2近戰與RE-244的固定seed1射擊流程，保持各自正常入口與操作控制器，只增加唯讀觀察。
工作入口為 `workplace/monster-reload-alias-20261006/prepare.py`，正常原版執行器為固定
dosgolem `968ca774705853276ae55c8d1163c369481ac71f`。

先驗證新增觀察不改變既有完整狀態、事件、骰序與 PNG，再分別記錄實際讀取值與最後寫入者。
本輪不修改正式程式，不以手持武器種類代替尚未閉合的原版讀取來源。


## 輸入與驗證

| 項目 | 結果 |
|---|---|
| 原版EXE | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| IDA Pro 9.4資料庫 | `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 位址空間 | IDA linear基址0x10000；DGROUP0x2DED0；寫入事件的ida已是IDA linear |
| Go image | eob-remake-go:1.26.7-ebiten2.9.9-audio；ID 3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa |
| 舊收據比較 | 1190狀態、304事件、10544骰序及919張PNG與各自RE-260／RE-244完全相同 |
| 新增觀察 | 131個讀取／階段點、16193筆寫入 |
| 靜態來源 | 338正式輸入、75原版檔案、169固定dosgolem來源、173道IDA載入指令相符 |

## 實際讀取與來源

RE-260近戰場沒有射擊，只作觀察中立對照。以下是RE-244正常REST場射後0x20842的觀察，非全部怪物或跨場資料契約。

| 角色槽：完整byte值 | 次數 |
|---|---|
| 6:33 | 13 |
| 7:0 | 8 |

| 角色槽 | 實際完整byte | 最後寫入IDA位址 | 射後讀取次數 |
|---|---|---|---|
| 6 | 33 | 0x1EC29 | 13 |
| 7 | 0 | 本場起點後未見寫入 | 8 |

最後寫入來源保存在`ranged/review.json`的read_summary與原始alias_writes。沒有觀察到寫入的值只記為本場初值，不推定跨場常數。

| 角色槽 | DS位址 | 對應128-byte怪物裝備記錄 |
|---|---|---|
| 6 | 0xCDC5 | 第0筆＋0x33，即首個頭部裝備記錄＋3 |
| 7 | 0xD0C5 | 第6筆＋0x33，本場未使用的裝備記錄 |

RE-244首場有三名怪物。槽7的位址在觀察起點已有完整零值，21次射後讀取中對應的8次均維持零，未記錄到該位址寫入。
此結果不能證明跨場清零，也不能把其他角色槽的別名當作零。

0x1EC29的原始mov [bx],al與前置位址計算包含於173道IDA來源bytes。本場其他raw寫入事件只保留工具的位址換算，未逐一歸類函式，也不把它們全部認作本EXE已證實的資料來源。

## 證據等級與限制

- 已證實：新增唯讀hook保持全部父收據欄位與PNG；原版完整byte暫存與讀值相符，非6／10／11的射後Command與Flag1D符合既有真實指令。
- 已證實範圍：本場實際讀值及記錄到的最後寫入指令。其餘別名位址、跨場緩衝與中途保存的正式模型保持未知。
- 42是比較欄位數，非物品或裝填種類值。歷史收據不重寫。
- 兩場各自維持十四／十三回合戰敗，未新增原版勝利S結尾、正式程式、保存、完整UI或平台驗收。
- 完整射後裝填別名仍待驗。SPEC-022父狀態及未完成條目保持，既有受阻規則確認與HD後續問題保持。

## 重現入口

在既有Go Docker image內依序執行prepare.py、build-original.sh、run-original.py、review.py。再執行prepare-ranged.py，將ranged子目錄掛為/out，依序執行其build-original.sh、run-original.py、review.py。兩個原版控制器與種子各自沿父收據。原版與專案來源唯讀掛載，只有工作輸出可寫。
封存後以專案及工作區唯讀掛載執行verify.py；RE264_REPORT供review.py把重新核對結果寫到容器/tmp。
靜態失敗與修正分別保留static-audit-v1-unrelocated.json、static-audit.json；seal.json及audit.json保存最終核對。

## 2026-10-06 RE-265追加勘誤

RE-264把42解釋為比較欄位數，有誤。RE-245 actor-contract.py共40個比較欄位；compare.py把各畫面retained_alias_differences筆數相加得42，全部是Command，Flag1D當時相符。原有收據與封存程式保持，詳見[RE-265](20261006-actor-record-coverage.md)。
