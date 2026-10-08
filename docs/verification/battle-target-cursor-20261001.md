# 戰鬥目標游標原版驗證（2026-10-01）

最新完整列已通過，程式為`e5be2113`，本地#33按兩塊原始條件結案，完整戰鬥仍待驗。
[攻擊／防禦完整列收據](battle-target-cursor-20261001/attack-row-receipt.json)與末節保存新建置範圍。
下述圓環批使用`ac79bb47`，新建置已再跑104框回歸；兩版結果各自保留。

**圓環窄契約已實作並通過同相位抽驗；完整戰鬥仍待驗。**
新增原版52框、重製版英／繁中倍率104框皆零差異；正常四次455鍵重播，
456筆已記錄狀態相同，34張更新後畫面只在戰場格內改變，四種子選單均實際走到。
[重製版收據](battle-target-cursor-20261001/remake-receipt.json)列明來源、相位、時鐘近似與限制。

首批原版來源驗證：12 個目標框與4個子選單框逐像素一致。
9 個空格、3 個有人格，各完整 16×16、差異 0。四個攻擊／防禦子選單仍持續播放圓環。
原版來源與分級原始定位見 [RE-171](../re/20261001-target-cursor.md)，
輸入、命令、腳本、PNG SHA-256 與限制見[收據](battle-target-cursor-20261001/receipt.json)。

原版以固定 dosgolem 版本 `968ca774705853276ae55c8d1163c369481ac71f` 重生。
seed 1 在旅店後、出城與遭遇前設定，未掃種子、重擲、強制遭遇、注入座標或替換等鍵常式。
`gui4` 從正常兩人隊伍自然伏擊進武器目標層；`gui5` 同前綴，再以正常移動、休息、
回合確認及空白鍵結算一回合，由 Theodoric 徒手目標層檢視相鄰 Poontang。

合成器讀既有十格標記 PNG 與戰場 PNG；以原版捕捉時點之前的最後 `sub_1CB28`
呼叫取得相位，以同時記錄的格圖與人物編號合成。有人格先合成人物，再疊圓環。
整框比較，未裁掉黑邊、未選最接近的相位；單位元差異反例被拒絕。
等待迴圈實際呼叫涵蓋單元 5–9，來回方向符合 IDA 資料流。

收據保存可重播腳本，需在 Docker 中執行：將儲存庫掛 `/src:ro`、dosgolem 儲存庫掛
`/repo:ro`、本收據目錄掛 `/scratch:ro`，另將已存在且由目前使用者擁有的研究目錄掛
`/out`。使用 `kol-package:20260907`、UID/GID 1000:1000、無網路與相稱資源／逾時；
命令及環境變數以 `receipt.json` 的 `replay` 為準。原始遊戲須已依專案流程準備於
`workplace/oracle/game/`。`kol_cursor_pixels.py gui4 gui5` 重驗捕捉相位及像素，
`verify_backlinks.py` 檢查分級索引、舊規格與舊研究的訂正入口。

首輪 `gui1` 沒進目標層，只列觀察；`gui2` 少捕捉時點，保留後以 `gui4` 正式重生。
`gui3` 核實正常移動的一回合，再以 `gui5` 保存每張畫格時點。首次 IDA 匯出遇 ASCII
預設編碼，改明確 UTF-8 後以新副本成功；舊副本與 log 保留本機。
像素腳本初次把模式 5 誤當目標層入口，後改用實際 YES／U-TURN 兩格圖示列。
子選單查證又發現模式仍為 5，已依實際呼叫證實其動畫持續；沒有放寬像素判準。

原版 ZIP、EXE、archive、完整人物快照、反組譯 payload 與 `.i64` 不入正式目錄。
PNG 沿用既有 private GitHub 例外，不包含公開散布權。首批沒有正式程式變更、
重製版種子對拍、遠程／取消／摘要完整生命週期、全遊戲 parity 或正式封包。
#33／#64 維持待辦；首批下一閘門的生命週期及同相位呈現已由後述窄收據補齊。

生命週期續核入口：[verify_lifecycle.py](battle-target-cursor-20261001/verify_lifecycle.py)。
以相同 Docker 掛載執行，逐框核對取消清除、空／有人摘要、子選單與遠程外圈；
輸出 `lifecycle-receipt.json` 及逐張 SHA-256。原始快照與完整呼叫仍只留本機。

正式呈現續驗：[同狀態 GPU 測試](../../internal/ui/battle_marker_test.go)核對相同格圖、人物、
相位的英文／繁中倍率；正常玩家重播腳本暫存於
`workplace/battle-cursor-20261001/normal_replay.py`，從標題沿既有任務鏈前綴進場，
只送正式按鍵，原版整場戰鬥對拍仍另列限制。

重製版正式窄收據：[remake-receipt.json](battle-target-cursor-20261001/remake-receipt.json)。
彙整器入口為 [finish_remake.py](battle-target-cursor-20261001/finish_remake.py)，
正常重播入口為 [normal_final_replay.py](battle-target-cursor-20261001/normal_final_replay.py)。
兩者都在上述有界 Docker 中執行；彙整器先核對 GPU 結果與正常路徑，才複製已驗 PNG。

原版生命週期重播使用 `lifecycle-kol_cursor_gui.sh` 及同前綴兩支 Python，掛本目錄
至 `/scratch:ro`；`remake-receipt.json` 的 `original_replay` 保存各次按鍵與名冊選擇。
`gui9／10` 使用名冊繪製後才送方向鍵的修正版；`gui6／7` 當時的完整生成Go來源與
來源雜湊保留本機。新重播以實際最後繪製相位核對，不宣稱不同腳本版本的指令步數相同。

此次正式修正包含圓環、首次遠程起點與空摘要條件。正常首輪誤把第一位 Olear
當近戰；改先按YES、再由 Poontang 走武器／徒手／移動／防禦，第二輪各455鍵通過。
首輪遠程起點修正後的空資訊框與原版一致，不列成產品缺陷；第二輪非戰場區仍嚴格零差。
GPU測試因 RunGame 結束後其他測試仍建圖片而失敗，改隔離子行程後同命令乾淨重跑通過。
觀察、來源與失敗均保留，不將綠色測試或此抽樣稱為整場原版 parity。

摘要來源續核由 RE-171-source 證實 `sub_249AE:0x249CB` 重載目前人物指令。
前版5bf08653只完成呈現，初次／換人來源仍錯；
[前版收據](battle-target-cursor-20261001/remake-receipt-render-only.json)與其PNG保留。
正式來源修正後再用同一正常路徑、種子與完整框零差異判準驗收，不混用兩版結果。

最終建置 `ac79bb47` 已完成上述四次455鍵重播；初次／換人主列整圖零差，
34張完整畫面非戰場區零差。完整程式雜湊、14筆自動註記來源與前版收據雜湊
列於最終收據；[verify_remake_backlinks.py](battle-target-cursor-20261001/verify_remake_backlinks.py)
核對 RE-160 來源勘誤並拒絕缺標記、
錯誤模組雜湊與錯誤等級，不讓舊結論繼續流入實作。

收尾稽核入口：[verify_remake.py](battle-target-cursor-20261001/verify_remake.py)，
在上述 Docker 掛載中執行，核對程式／來源／附件雜湊、PNG格式、UID/GID與舊收據，
並拒絕錯誤程式及原版雜湊反例；結果保存於同目錄 `remake-final-audit.json`。

完整攻擊／防禦圖示列續核入口：[verify_attack_rows.py](battle-target-cursor-20261001/verify_attack_rows.py)，
獨立像素合成器為 [attack_rows_pixels.py](battle-target-cursor-20261001/attack_rows_pixels.py)。
正常原版重播與英／繁中修正前後重播在既有有界Docker執行；`--export` 只保存全部
零差後的PNG及 `attack-row-receipt.json`，預設唯讀重驗。舊薄框畫格仍保留本機，
不修改首批或圓環收據；原版箭頭以正式 `MoveMouse` 事件移到右上角後重生。

完整列已完成，見[attack-row-receipt.json](battle-target-cursor-20261001/attack-row-receipt.json)：
17原版狀態與34英／繁中正常畫格的296×32完整比較區均零差，保留全部圖示、外框、
間隙與尾端。六個武器／八個防禦選取及右移繞回已驗，單位元／錯白框／缺圖反例拒絕。
新程式`e5be2113`只修繁中戰鬥圖示列框線；英文17整圖不變、繁中差異只在各框線。
兩語言各438正常鍵／439筆狀態與修正前相同；104個圓環框回歸及Go測試通過。
原版沒有繁中畫面，本次核對的是既定三倍呈現，不稱整場戰鬥同狀態。

本地#33按原始「兩塊都有逐畫面收據」條件結案；早先「#33仍待辦」是前批歷史。
其他圓環等待層及法術／戰利品／物品頁仍由#64與完整戰鬥驗收追蹤，沒有遠端Issue寫入。
來源重播入口為 `attack-kol_cursor_gui.sh` 及同前綴兩支Python；
重製版為 `attack_after_replay.py`，舊薄框重播為 `attack_normal_replay.py`。
沿用前述Docker掛載，原版設`KOL_CURSOR_ATTEMPT=gui11`，按鍵以收據`original_keys`為準。
新程式位於`workplace/battle-cursor-20261001/kol-attack-frame-linux`；舊ac79程式與原始薄框保留，
不把舊圓環收據的建置雜湊換成新程式，也不宣稱正式平台封包完成。
新附件、來源與擁有權稽核保存於 [attack-final-audit.json](battle-target-cursor-20261001/attack-final-audit.json)；
首批與圓環附件仍按各自雜湊核對。
