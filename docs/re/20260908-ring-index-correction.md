# 搜尋環索引與射擊門檻訂正

輸入 KOLTOWN-unpacked.EXE，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4，以下為 IDA linear，DGROUP 基底 0x2DED0。
執行器 dosgolem fd746405 加既有本機相容性修正，逐檔來源雜湊在收據 manifest。

## 新證據與舊結論訂正

SPEC-013 舊說法將原版 word_349B2 的 1 當成相鄰圈，因而把射擊門檻 4 減成 3。
首次原版測試直接讀 DS 相對 0x0FC1 得到 16，與此假設矛盾；
失敗收據保留在 workplace/ring-search-20260908/first-mapping。

修正假設後，TestBattleRingSearchOriginal 直接確認環 0..6 表值為
8、16、24、32、40、48、56，逐項與 remake 同號環格數一致。
原版 sub_2374C:0x23C1A 先設環號零，0x23C6A 搜尋鄰圈；
0x23DDE 的 1 是後續外圈起點。這兩條獨立證據足以否定「索引須減一」解讀。
以上皆為已證實，不以自訂欄位名稱作為語意依據。

## 有界執行

- sub_2531C 真實讀環長表，首次呼叫 sub_1B5A8 的上限是格數減一；
  最低／最高合法起點各跑完整遍歷，七環共 14 條，448 個候選與 14 次終止。
  續查不重新擲骰，環尾回繞，回到首次起點回 -1。
  只替換候選判斷 sub_24BB6，保留環表、骰面參數、游標與終止控制流。
- TestBattleEnemyShootRingOriginal 執行 0x23EBB..0x23F12，僅在射擊／移動出口
  回傳標記；六環 × 種類零／非零的 12 組通過，右臂固定有效。
  第 3 環仍可射擊，第 4 環改移動，無索引轉換。

原版與 remake 表長一致並不單獨證明每個環的座標順序、候選遮罩或整個 AI 決策一致。
本次修復只依 READY 增補訂正 findInRing 的起點骰面及 EnemyShootRing 門檻。
counterScan 的獨立起點迴圈仍須核對；不把本次 helper 證據自動延伸到未追查呼叫鏈。

收據：workplace/ring-search-20260908/manifest.json 與 boundary/manifest.json。
三筆已審查語意回填 battle-panel-address-index.json；由 IDA 唯讀資料庫副本匯出，
自動合併原位址、bytes、原運算元、已證實分級與證據來源。
