# RE-252：戰鬥訊息底色與成功掉落名稱

日期：2026-10-05。狀態：有限 CONFORMED，完整介面及戰鬥仍開放。

路由為「原版GUI還原」與「逆向結論回填」，IDA Pro9.4工具鏈沿既有鎖定映像。
正常原版215格與重製218步的文字、黑字及位置一致，165個底色像素不同。
修正背景後只剩296格對299步的Broadsword遺漏，854個訊息像素不同。
先保留V1反例，再核原版掉落來源；V2通過READY後才接正式六個既有檔案。

## 輸入與定位

輸入KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
資料庫workplace/analysis/koltown/KOLTOWN-unpacked-ida94-derived.i64，SHA-256
`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
工具IDA Pro9.4；以下位址均為IDA linear，載入基底0x10000、DGROUP基底0x2DED0。
MZ標頭9216bytes；relocation segment加0x1000後逐項對原始bytes。
五個繪字函式389道與三個傷害／掉落／代換函式867道，共1256道相符。
原始名稱、位址、operand與bytes保留在本機匯出，正式資料庫唯讀。

## 已證實的有限契約

| 原始定位 | 語意與證據 |
|---|---|
| sub_1C1BF:0x1C1DA | word_2DF1C設FF，執行卷軸背景白色 |
| sub_1C311:0x1C327 | 模式0D呼叫實際EGA出口sub_1CE04；sub_1DE84屬低模式分支 |
| sub_1CE04:0x1CE55／0x1CE74 | 小字六列，大字才八列 |
| 0x1CE93／0x1CEA1 | 每列寫背景與反向字模；零字模空白也寫FF |
| sub_1C40D:0x1C40D | 8×8矩形只作交疊檢查，呼叫sub_1C3CB；該函式沒有填色，不是字格背景 |
| sub_277F4:0x27D9E | 成功位移分支先對目標呼叫sub_269BA，再印0x15B |
| sub_269BA:0x26A65／0x26A6B | 空地上槽成功時寫word_39A54及word_30D02 |
| sub_265F3:0x26769／0x26774 | 代換碼H讀word_30D02，索引DS9D20的16-byte地上記錄 |

舊RE-087把EGA出口及8×8交疊檢查誤當清底。實際小字格為8×6，包含空白。
RE-114／115與早期地圖的同名函式清底註記同樣不能當作背景寫入證據。
滑鼠遮蔽語意只作強推論，本輪沒有新增滑鼠行為或其正常驗收。
書本空白跳過底色的RE-043契約只適用該書本路徑，不能外推戰鬥。

## 玩家垂直鏈與正式變更

原始字模PNG及KOLDATA → Font／typed訊息 → 執行段斷行 → 雙語BattleField。
背景按既有字格填MessageBack，再畫原墨色。空白包含在內；字型、座標、換行保持。
成功掉落原始物品 → typed完整Item及Ground → forcedDrop回傳成功名稱 →
快速與逐句攻擊的既有ActResult.Item → KnockOnly返回前DisplayName → 卷軸。
公開Knockback介面保留；位移、掉落、清指令、骰序與v3資料契約沒有改動。
Broadsword與既有譯名「闊劍」恢復，繁中299步的變動只在卷軸矩形。

## 正常驗證

原版dosgolem revision968ca774705853276ae55c8d1163c369481ac71f，383狀態與5337次骰
沿RE-245不可變收據。兩側執行前固定seed1，385個等效鍵沿既有映射，未重擲或注入。
V1四次1544狀態保留212/213的失敗；V2三次1158狀態通過，另補300步繁中舊版反例。
只有299步exec_text改為完整掉落句；其餘typed狀態与39AI／13執行／13規劃事件相同。
213個英文訊息區逐像素零差。213張完整畫面均只差右下滑鼠67像素，無遮罩。
沒有一張整圖相同，完整UI仍未通過。繁中既有文字、字形與排版沿既定契約。

正式Linux SHA-256 `c87645925936137788443b6eda8b753842c0a725cbf4d7dd9139a631ffa302e5`，與V2候選位元組相同。
336Go／338輸入，六檔改動；十九套件通過。雙語1136正常保存狀態與26新程序重載通過。
完整v3只差saved_at，兩語言保存相同，重載trace相同。68張保存／重載PNG變動只在卷軸。
本輪各正常、反例與保存重載共4164狀態，不外推自然全旅程或原版整場規則parity。

## 仍開放的限制

- 0x27DD7至0x27DF4阻擋且持武器的打暈分支沒有呼叫sub_269BA，現行共同掉落處理可能不同。須另取正常原版阻擋樣本，本輪不改規則，也不猜H名稱。
- 地上十六槽全滿時原版保留舊H索引，remake未建模。本輪只記成功位移且實際掉落名稱。
- 原版方向確認及空隊伍，與現行ENTER／SPACE及保留六人仍不同。42射後別名、全50-byte終態與其他分支未過。
- 最新長任務仍需新版來源重跑；Windows／macOS真機、正式封包與HD解析度／介面仍未驗。

## 重現與封存

本機workplace/battle-message-background-20261005/入口：ida-export.py、ida-drop-export.py、
prepare.py／prepare-v2.py、build-v2-refined.sh、run.sh／run.py、promote.py、build-formal.sh、
run-save.sh、review-regression.py、update-docs.py、seal.py／verify.py。
V1反例、兩次IDA工具修正與原始輸出保留。第一次IDA因ASCII讀中文失敗，改明確UTF8後同鏈重跑；
第一次分支匯出缺EGA實際出口，補完整函式後才審查；promotion準備語法錯在寫入前停止，修正重跑。
本機原版資料與PNG沿既有權利分類。沒有commit、push、Issue寫入、公開發布或版號變更。
