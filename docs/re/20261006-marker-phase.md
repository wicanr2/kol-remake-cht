# RE-277：正常戰鬥圓環的捕捉相位

2026-10-06。已證實：RE-276八張正常MOVE畫格的288個非滑鼠差異均來自捕捉相位。
按原版最後繪製呼叫指定相位，英文及繁中共16個完整圓環框零差。
正式來源仍為342輸入／340Go、Linux97201bdb，沒有產品修改或新保存／長任務驗收。

## 來源與方法

原版沿[RE-273正常MOVE](20261006-ai-flank-normal.md)固定seed1、632狀態與正常策略。
新觀察只讀 `sub_1CB28` 且 caller return為0x114AB的呼叫，記下引數、格子與相位。
每張畫格另記最後繪製呼叫及捕捉時的可見磚圖／精靈，保留其原始值。
原版正常入口、不替換等待、不注入角色／指令／座標、不重擲。

原始封裝SHA-256 `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。
此字串沿父收據於封存核對；完整輸入與工具入口見[RE-171](20261001-target-cursor.md)。
dosgolem來源為固定 `968ca774705853276ae55c8d1163c369481ac71f` 封存，
由既有Go1.26.7映像建置，沒有修改共用dosgolem工作樹。
動態位址用IDA線性基準：runtime linear減DGROUP segment×16與0x2DED0的差。
0x114AB是far call的回傳位置，call指令為0x114A6；不是原始檔案偏移。
來源已證實語意沿[圓環索引](target-cursor-address-index.json)，未新增硬體時序逆向。

新原版完整rows、events、rolls及所有舊輸出字段與RE-273相同；所有原版PNG逐byte相同。
唯讀hook沒有改CPU步數、骰序、角色狀態或畫面。只多出marker_calls與marker_frames。

## 原因與驗證

正式 `cmd/kol/main.go` 的jobs截圖分支先返回，未呼叫advanceBattleMarker。
BattleTiles初始相位為0，因此截圖取圖集單元5；互動遊玩才依SPEC-010推進動畫。
該規格已明列三BIOS tick等待的現代近似，允許腳本截圖固定相位，不宣稱精確wall-clock。

原版最後繪製的引數決定單元，不由圖片擬合或選擇最接近的相位。
每格底圖按原版可見磚圖及精靈合成，依RE-171小端16-bit遮罩畫圓環。
八個原版框與其實際相位模型全符；八個正式截圖框與固定單元5模型全符。
兩模型之間的差異位置逐點等於整頁除既有滑鼠外的全部差異，合計288像素。
沒有遮罩新區域，原始35頁的差異數仍為288；這批結論說明跨相位比較，未修改原版圖片。

現行共用renderer另用觀察到的原版相位測試，英文16×16及繁中48×48各八框全符。
fixture只為同相位呈現測試，不當作正常全遊戲或互動等待時序對拍。
實際原版正常路徑與已封存重製正常路徑另保留於收據。

## 重現與限制

本機工作區 `workplace/marker-phase-20261006/`：prepare.py複製鎖定工具與加唯讀hook，
build-original.sh建置，run-original.sh正常重播，analyze.py核所有原版輸出與相位模型，
test-renderer.sh驗現行共用renderer的16倍率框；update-docs.py回填，seal.py封存與重核。
finish-docs.py登記已由本證據解除的歷史引文，再核清單；claim-check.py核正反對照。
原版執行匯出、PNG、fixture與完整角色記錄均留本機，沒有新增公開散布權。

本輪只確認八個已知跨相位差異。其他等待層、完整UI、完整50-byte、勝利正常GUI、
操作與既有DRAFT決策、平台真機及正式封包保持未驗。沒有新來源、資料格式或HD正式素材。
沒有Issue、提交、推送、版號或發行變更。容器與Xvfb有界且收尾移除；UID/GID1000。
