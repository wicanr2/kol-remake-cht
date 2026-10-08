# RE-210：正常持續法術與住宿保存

日期：2026-10-03。正常原版自然失效與重製版正常保存鏈有限 CONFORMED。
正式 323 份 Go 與候選 b3a5d06b 保持 RE-209；SPEC-035 因名稱輸入接線缺口保持 READY。
工作區為 `workplace/sustained-spell-chain-20261003/`。
承接 [RE-209](20261003-spell-lifecycle-player-path.md) 已驗的正常 Strength 修改與扣款。
本輪從正常購買、修改後出城施法，觀察有效倒數與失效，再走正常住宿保存及正式重載。
原版固定 dosgolem revision 與 seed1；不注入角色、位置、物品、骰結果或勝利。
原版怪物四套資料格式仍待決，整場狀態若因此不可比，保留差異並限制驗收範圍。
所有原版 payload、私用驅動、收據與未驗圖片僅本機。正式發布、格式與版面決策保持。

## 原版正常失效證據

推論等級：本次正常相鄰可見目標已證實。工具為 dosgolem 固定 revision
`968ca774705853276ae55c8d1163c369481ac71f`，正常 KOL.EXE／KOLBOOK／KOLTOWN 啟動。
正常旅店招呼後設定 DS:1600=1、1602=0；重製以 `-seed 1` 啟動。
後續不重設種子、不注入角色、位置、物品、回傳值或等待。
來源 `KOLTOWN-unpacked.EXE` SHA-256：
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
位址空間分開記錄：IDA Pro 9.4 linear；狀態為 DGROUP 相對偏移，基底 0x2DED0。

正常 Strength 修改後依實際 IconDoor=2 離店，再步行出城遇伏擊。
Olear 正常選負面法術，對相鄰 Theodoric 施法後兩人 REST。
原版力量 128→117，原始啟用旗標 actor+0x0D 為 1，倒數 actor+0x30 為 2→1→0。
第 5 回合力量恢復 128、旗標清零，實際印出 KOLDATA 0x173，
呼叫端為 IDA linear 0x20F68。共 141 個戰鬥輸入，未走勝負出口。

`native-chain3` 在印字呼叫鉤子即停止，最後圖仍為前頁；`native-chain4` 讓原版再執行
20,000,000 步，不送任何新鍵，實際顯示「THEODORIC'S SPELL WEARS OFF!」。
兩批初末 actors、characters、counters、seed 與 141 鍵逐項相同。
這是正常原版自然顯示的收據，不用 stub 或複製頁面。
來源 prefix 有 201 個狀態；修改後法術原始槽為 `85958e0b1ee00c814000000001000a22`。

## 正式重製正常分支

| 分支 | 每語言正常鍵 | 每語言狀態 | 已驗結果 |
|---|---:|---:|---|
| 修改後施法、REST 至自然失效 | 359 | 360 | 力量128→120→128，倒數3→2→1→0，失效頁 |
| 修改後施法、正常逃走、回城住宿 | 656 | 657 | 全新v3保存，法術效果1、金冠601／冒險點數2486 |
| 另起正式程式重載、招募Olear | 15 | 16 | 兩人34物品槽與14持久欄位相符 |
| 正常開名稱輸入、取消與拒絕保存 | 196 | 197 | 原名Kelnalon／效果6、金冠987／冒險點數2816保持 |

英文與繁中、正式與唯讀觀察各跑上述分支，共 4920 張 PNG 逐張相同。
失效句繁中為「Theodoric 的法術效力消退了！」，已目視安全範圍。
名稱輸入頁顯示 caret 並隱藏注解；但正常 BS／N 的 typing 一直是 Kelnalon。
直接呼叫 `TypeSpellName` 的既有單元測試沒有涵蓋 `Session.Key` 正式路由。
正常名稱輸入尚未完成，這是本輪新確認產品缺口。取消 PASS 不代表文字輸入 PASS。

保存分支在自然失效前逃走，由既有出口還原暫態屬性；自然失效和保存分開驗證。
保存法術的 guards=[64,0,0,0]、effect=1、turn_code=0、power_code=10 保持。
Blue Gem／Blue Chain 及頸部原始信物記錄、時鐘420／Part4／Day2／Month3／Weather2保持。
英文與繁中保存除時間戳相同，正式重載保留 name／gold／wounds／carried／spells／
token／order／quests／progress／sheet／worn／backpack／in_hand／steed。

## 限制與驗證問題

原版持續2／力量差11，重製持續3／力量差8。相同 seed 數字不等於相同骰條件。
已知初始怪物四套資料格式仍待決，本輪不是整場同狀態或像素 parity。
原版住宿保存、其他視野邊界、全法術類型與其他主要玩家路徑仍未驗。
原有繁中注解 A／B、資料格式與發行決定保持待答，不由自動續輪核准。

初次原版離店鍵錯誤且程序被 kill，確切終止原因未知，末張 PNG 不完整。
第二次驅動期待 UTurn，被實際 icon guard 拒絕；查明 Door=2 後同正常路徑通過。
重製探查的離店選中 SCROLL、世界 Y／非露營 SPACE、raw grid／PNG 雜湊混淆，
以及 prepare 失敗後提前啟動相依 probe，均為驗證腳本問題。
繁中初驗誤期待英文失效句，程式已正常完成；依譯表371修正檢查後重驗通過。
無效探索與不完整 PNG 僅本機留存，不登錄為已驗圖片。

## 收據與重驗

收據：[receipt.json](../verification/sustained-spell-chain-20261003/receipt.json)。
唯讀入口：[verify.py](../verification/sustained-spell-chain-20261003/verify.py)。
在 Docker 掛專案 `/src:ro`、本工作區 `/out:ro`，執行
`kol-package:20260907 python3 /src/docs/verification/sustained-spell-chain-20261003/verify.py`。
它核對本輪附件、323份未變Go、前批4356附件與335份封存輸入、4920張正式／觀察PNG。
原版payload／研究驅動與探索圖片僅本機；沒有提交、推送或公開交付。
入口Git異動、Docker清理與擁有權見本工作區 `audit.json`。
