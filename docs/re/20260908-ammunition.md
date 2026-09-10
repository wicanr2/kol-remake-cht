# 隊員彈藥欄位與射擊消耗補證

輸入 `KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
來源資料庫 `KOLTOWN-unpacked-ida94-derived.i64`，IDA Pro 9.4；以下皆為 IDA linear，
DGROUP 基底 `0x2DED0`。沿用非破壞性匯出 `panels-private.json` 的原始函式、指令與 bytes，
不改資料庫名稱。執行器為 dosgolem `fd746405` 加已記錄的本機相容性修正。

## 已證實

| 原始定位／bytes | 附加語意 | 證據 |
| --- | --- | --- |
| `sub_1E3A8:0x1E78D`，`c687126713` | 隊員首位元組初始化為 19 | 原指令 `mov byte ptr [bx+6712h],13h`；另一初始化路徑 `0x1E921` 同值 |
| `0x2088D`，`8a871267`；`0x20894`，`05ffff` | 讀首位元組後減一 | 消耗段 dosgolem 2,550 組比較 |
| `0x208A4`，`c7067ebbff00`；`0x208B8`，`c6472500` | 減至負值改為 255，清除指令 | 同上，含原值零邊界 |
| `0x208CE`，`888f1267` | 寫回 `[bx+6712h]`，不是唯讀人物種類 | 同上；保留原始位移 |
| `sub_228B7:0x229A5`，`3dff00` | 耗盡時不提供射擊／裝填 | 原始控制流 `0x22996..0x229AD` |
| `sub_22EFC:0x231AE..0x231D1` | 255 顯示零，其他值顯示加一 | 128 組原版標頭／數字出口測試 |

原版消耗測試只在 `sub_1E3A8` 入口建立有界堆疊並跳至 `0x2082C`，
於 `0x208D2` 返回。保留原函式段位址；扣除、種類判斷、裝填重設、耗盡及寫回
均執行原版 bytes。這不涵蓋彈道、命中、正常輸入或整場戰鬥。

## 舊結論訂正與實作缺口

原版首位元組不是所有陣營都表示種類：敵人仍是怪物索引，隊員則是彈藥減一。
既有 `Actor.Kind` 導覽名稱及「隊員填 0x13」註解不足以證明完整語意；
讀取、扣除、耗盡及顯示證據共同推翻將其視為固定隊員種類的解讀。
保留既有欄位，補明兩種語意，避免改動敵人肖像與資料格式。

稽核確認正式 `State.shoot` 未扣除該值，導致隊員可一直射擊。
SPEC-022 經 DRAFT／READY 審查後，將消耗接至真正射出後的共同收尾；
裝填提前返回，射空、阻擋及命中返回都會消耗。原版 `0x20824` 呼叫射擊結算
返回後無條件落入消耗段，沒有以命中與否跳過扣除。

本機初始原版收據 `workplace/ammunition-20260908/manifest.json`；
正式比較收據 `workplace/ammunition-implementation-20260908/oracle/manifest.json`。
彈藥標頭的畫面座標、正常玩家射擊及整體 AppImage 驗收仍待完成。

## 後續：彈藥文字與資訊框接線

`TestBattleAmmunitionTextOriginal` 使用已解碼 KOLDATA 0x101..0x103 原始文字，
執行真實模板、標頭與十進位數字常式；只替換字元輸出與無關物品解名。
六組確認模板後的 (24,60)、word_2DF54=1 與無空格的 ARROWS:／BOLTS: 數字。
不模擬字元前進，故不把此測試當成原版字寬或像素對拍。

128 組標頭／發數已直接比較正式 EquipmentPanel。首次原版比較漏設
Actor.Member 槽映射，導致第六隊員讀到第一隊員；修正測試後通過，正式程式未改。
首次失敗保存在 ammunition-panel-20260908/first-member-mapping，最新 oracle 全批 0.393 秒。

八筆彈藥附加語意經證據審查回填 battle-panel-address-index.json。
IDA Pro 9.4 以唯讀資料庫副本重生匯出，逐筆核對原始位址、bytes、原運算元、
已證實分級及證據出處，索引雜湊相符；不修改原始函式或資料庫名稱。
新的 panels-private.json 與 ida-ammunition.log 位於 ammunition-panel-20260908。
