# 左下命令面板實作

日期：2026-09-08。SPEC-021 先建立 DRAFT，依原版參數／文字／方向及敵人閘門證據
審查升 READY，再接正式 BattleField。SPEC-021 尚未 CONFORMED。

新增 State.CommandPanel，區分當前 Planning 與其他 Actor 的命令、招式、防禦來源。
方向採被檢視者 Facing，敵人檢定讀行動者智力、敵人等級與已保存 InspectionRoll，
不另擲亂數。空人物與執行階段不顯示。原版五種三字母標籤依已驗座標畫出，
KOLMPICS 圖片要求 24×20，缺檔及尺寸錯誤回報錯誤。

回歸驗證包含當前規劃來源、敵人記錄來源、嚴格大於門檻、防禦跳號、方向與清框。
既有完整 `go test ./internal/... ./cmd/kol -count=1` 與建置通過。
產物位於 `workplace/command-panel-implementation-20260908/`，tests.exit 為零。
容器為既有 kol-package:20260907，輸出擁有權 1000:1000。

正常 GUI 顯示／選目標／取消與原版同狀態仍待驗證；3,168 組原版參數證據目前
尚未直接逐項比對此新增 helper。完整繁中戰鬥排版、彈藥數字、保存與 AppImage
驗收亦未完成，不能宣稱完整 A10 或整體操作等價。
未提交／推送，容器自動移除。

## 後續原版參數直接比較

`TestBattleCommandPanelOriginal` 已直接呼叫正式 State.CommandPanel，
比較原版 sub_2324F 的繪圖出口與 remake 的標籤記錄、圖碼、座標、順序。
3,168 組全部通過，整批 panels 0.272 秒。原版側仍保留獨立固定期望表，
不以 remake 產生原版答案。

範圍：兩種人物來源、八類命令、招式 0..10、防禦 0／1／6、方向 0／7／8 及
顯示／不顯示。顯示關閉組以 remake 空人物狀態比較無輸出；它不證明敵方檢定
閘門本身等價，閘門的原版及 remake 邊界證據仍分別回查既有測試。
此為參數契約比較，不是原版正常玩家畫面的像素對拍。

收據 `workplace/command-panel-implementation-20260908/oracle/manifest.json` 與
`panels.log`，執行前來源清冊未變，輸出雜湊與 1000:1000 核對通過。
本輪只修改原版對照測試，不重建未改動的正式執行檔；正常 GUI 尚待驗證。

## 正常一輪結算抽測（命令標籤尚未驗收）

新增 KOL_COMMAND_PANEL=1 腳本分支，從正常開場進移動規劃、選走路、下一名休息，
再推進結算。實際畫格 command-round-ready 顯示兩人規劃完成；command-after-round
已回到 Theodoric 的下一輪主指令，人物相對位置改變，傷勢遮罩及疲勞條更新。
初始相鄰格為空，命令面板依規則清框；本輪沒有取得可見命令標籤，不能標為 GUI 完成。
下一步在結算後兩名隊員相鄰的狀態下檢視隊友，驗證其已保存命令及返回。

收據 gui-round-receipt.json 保存輸入數及截圖雜湊，執行檔／腳本執行前後一致，
輸出 1000:1000。這是 remake 正常一輪結算證據，不是原版同狀態或完整戰鬥勝負驗收。

## 隊友命令可見性及取消恢復

沿用同一執行檔與正常開場／一輪操作，在下一輪選徒手進目標格，檢視相鄰 Poontang。
人工確認 command-inspect-ally.png 顯示 Poontang 肖像、人形，以及左下 OPT 標籤與
休息床圖示；代表正式玩家路徑已讀取其他隊員保存的命令，不是測試注入畫面。
選返回後，command-inspect-return.png 與檢視前 command-after-round.png 整張像素差異零。

148 個輸入完成，收據於 `workplace/command-panel-implementation-20260908/ally-gui/`，
`gui-verification.json` 保存圖檔雜湊與結果。執行前後 binary／腳本雜湊一致、擁有權
1000:1000。前次僅完成一輪但未顯示命令的收據保留，未覆寫成成功。
此抽測只驗 OPT／休息與隊友來源；ATK／LOC／DEF／DIR、敵人閘門及原版同狀態 GUI
仍須各自驗收，SPEC-021 維持 READY。

## 真實原版檢定與正式面板可見性

新增 TestBattleCommandVisibilityOriginal，真實執行 sub_24131 及 sub_252CB，
不替換檢定結果。行動者智力 60、等級 9，敵人等級 2，門檻 39／40／41；
目標槽 0／6／17 共九組。隊員皆顯示，敵人只有門檻 39 顯示，等於及高於 40 隱藏。
原版結果與正式 State.CommandPanel 的可見性直接比較全部通過，亂數出口若被呼叫
即失敗。這補足前次關閉組使用空人物狀態的證據限制，但不代替敵人正常 GUI。

整批 panels 0.240 秒通過，收據 `visibility-oracle/manifest.json` 與 `panels.log`。
執行前來源未變、輸出雜湊及 1000:1000 核對通過。正式程式未改，不重建執行檔。
