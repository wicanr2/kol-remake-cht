# 戰鬥肖像與裝備框修復驗證

日期：2026-09-08。狀態：基本功能已實作；SPEC-019 維持 READY，尚未 CONFORMED。

## 實作與證據

依 [原版證據](../re/20260908-equipment-panel.md) 及
[SPEC-019](../spec/019-battle-equipment-panel.md)，補上左上隊員肖像、敵人圖像、
六列隊員／五列敵人裝備，以及主指令和人物目標格的空白鍵切換。
空白鍵不再誤當確認指令；取消子選單保留模式，輪到下一名隊員時恢復肖像。
護甲名稱讀取目前可變裝備，破損後不回退顯示初始名稱。

新版 dosgolem `fd746405` 加現用分支本機修正，直接執行固定原版常式：
十二組模式／初值、人物與空格、四個隊員／敵人槽位及兩個肖像位置測試通過。
連同人形與疲勞測試，`oracle-source/panels.log` 結果通過（0.064 秒）。
首輪測試誤把 DS:2DAE 與 IDA 線性位址 30C7E 當不同欄位，造成圖示值被覆蓋；
已修正測試前置並保留 `first-icon-alias/`，沒有因此修改 CPU 或遊戲行為。

原版輸入為 `KOLTOWN-unpacked.EXE`，SHA-256：
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4 採線性位址，DGROUP 基底 0x2DED0。重新匯出的
`workplace/battle-panels-20260908/panels-private.json` 已合併全部 28 筆分級註記；
索引 SHA-256 為 `a0f53205374b85fd08ffe2e6dcc5c7114a3b38d8fc25b08b01686b9765da88f7`。

## 重製版驗證

`kol-package:20260907` 容器內完整執行 `go test ./internal/... ./cmd/kol -count=1`
及建置通過。新增測試涵蓋破損、換手、丟棄、取消保留模式、新角色重設、
敵人五列、肖像及盾牌轉接，以及空白鍵不得提交命令。

正常 GUI 使用 `tools/verify_starting_inn_gui.sh`，參數為
`KOL_ENTER_BATTLE=1 KOL_WORLD_STEPS=100 KOL_KEY_DELAY=0.3 KOL_EQUIPMENT_PANEL=1`。
從標題選 Theodoric、招募 Poontang、離店、出城步行遇敵，再切換面板。
126 個正常輸入完成；人工檢視肖像與裝備清單。兩次肖像畫面整張像素差異為零，
兩次裝備畫面亦為零，證明本次切換沒有推進戰況或留下繪圖殘影。

產物：`workplace/equipment-panel-20260908/`。執行檔 SHA-256：
`24c113944b57707ebcef84c69f687f7d94b71ad9eccbaaf6cb0a26e9460ffb60`。
執行前雜湊在 `gui-inputs.sha256`，結果、逐檔雜湊與擁有權在 `gui-verification.json`。
截圖為 `battle-portrait.png`、`battle-equipment.png`、`battle-portrait-return.png`。

## 限制與清理

原版常式測試與 remake 正常 GUI 是不同層級證據，不能合稱原版正常戰鬥同狀態對拍。
彈藥標題／數字語意、左下命令／招式圖示、正常攻擊後破損顯示、完整保存及
AppImage 重建驗收仍未完成。未改存檔格式，未將 SPEC 升為 CONFORMED。

檔案擁有權 1000:1000，`git diff --check` 通過；本批 KOL 容器全部退出並自動移除。
未提交、推送或發布；原版與衍生研究輸出保留既有權利邊界。
