# 受阻打暈持物正式驗收

日期：2026-10-07。範圍：[SPEC-008 的受阻打暈保留持物子切片](../spec/008-combat-field.md)。
使用者確認採原版規則後，先審查 READY，再接正式來源及有限 CONFORMED。

| 驗證 | 結果 |
|---|---|
| 舊版反例 | 隊員／怪物與人物／地形阻擋四個案例均抓到錯誤掉物 |
| 正式測試 | 十九套件通過，五套件無測試 |
| 正常玩家操作 | 雙語各637狀態，正常招募、旅行、戰鬥、回城住宿與真正保存 |
| 新程序重載 | 各13狀態，真正v3保存不改寫；完整trace符合既有已驗原型 |
| 唯讀觀察 | 各637狀態；正式與觀察完整trace、PNG及排saved_at的保存一致 |
| 原版受阻事件 | 第317格映射重製320步；Olear完整手持、位置、指令、seed及返回暫存符合 |
| 持久欄位 | 每語言126項符合原版，包括六人各17個完整carried槽 |
| 戰鬥種子與返回暫存 | 每語言268點符合 |
| 世界位置與時間 | 每語言168點符合 |

原版使用 [RE-254](../re/20261005-blocked-knockback-return.md) 的封存 dosgolem
正常收據，revision `968ca774705853276ae55c8d1163c369481ac71f`。本輪沒有新增原版執行。
兩側seed2於啟動前或原版第一個延伸鍵前固定一次；不重播種、注入坐標或挑結果。
現行重製沿既有636個正常鍵，控制器及原版映射保持。

原版 KOLTOWN.EXE SHA-256：
`c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
IDA linear 位址與角色記錄內偏移分開記錄於SPEC-008和RE-253／254。
完整手持為 `72b0c013ee000000225828ff190be200`；返回暫存6，seed2126174334。
正式Linux SHA-256：`55613e676d6b218c5f077be52e448b28c2f9e97d1159a63c5720cd6ec4cb5547`。
Go1.26.7、Ebitengine2.9.9；Docker沿用 `eob-remake-go:1.26.7-ebiten2.9.9-audio`。

原版和現行英文320步、繁中320步已目視。持物規則通過；原版舊H名稱與現行空名
仍不同，不稱整頁相同。完整整場、未知角色記錄、其他操作、三平台真機、HD與發行仍開放。
原版與衍生圖片、payload均僅本機，不新增公開散布授權。

工作區入口：[預登記](local/blocked-stun-formal-20261007/plan.json)、
[正常回放](local/blocked-stun-formal-20261007/run.py)、
[逐項比較](local/blocked-stun-formal-20261007/review.json)、
[封存收據](local/blocked-stun-formal-20261007/receipt.json)、
[唯讀核對](local/blocked-stun-formal-20261007/verify.py)。
原工作區已按使用者要求清理。原回放、比較與核對腳本保持在壓縮證據中；
保存內容的唯讀核對改用 [保存工具與方法](retained-evidence.md)。
跨平台驗收另見 [現行收據](local/finish-platform-20261007/receipt.json)。
