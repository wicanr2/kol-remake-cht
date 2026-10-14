# 法術卷軸HD正式接入

狀態：本切片**有限CONFORMED**。沿[SPEC-039](../spec/039-hd-art-ui.md#法術卷軸正式接入)接入正式素材載入與呈現。
基底90783017，正式376份Go／module，全套internal／cmd及sceneverify通過。
[正式比較頁](hd-spell-scroll-runtime-20261014/review.html)提供16張繁中畫格，來源與權利見[清冊](hd-spell-scroll-runtime-20261014/manifest.json)。
目前只驗Linux容器，舊Release保持；其他圖形仍續作，不代表全畫面HD或自然通關。

| 項目 | 結果 |
|---|---|
| 四組主路徑 | 十三欄編輯、名稱輸入、種族及亞種選單，每側107狀態相同 |
| EGA回歸 | 全部107張畫格與90783017相同 |
| HD與原型 | 四張完整畫格最多一色階，沒有更新原型掩蓋差異 |
| 返回建築 | 經既有不提交操作回到建築，配色與基底最多一色階 |
| 實際視窗 | 18個F10事件保持三組狀態，輸入緩衝保持；三組完整HD往返零差 |
| 鍵鼠及縮放 | 1920×1200、960×600、Return確認名稱、滑鼠選None與Ctrl+Q退出0 |
| 保存邊界 | 舊v1在啟動時轉v3，正式GUI輸入前後檔案SHA完全相同，編輯副本沒有提交 |
| 正常保存 | 住宿11狀態及另程序重載5狀態相同，完整v3只差saved_at，重載原檔SHA保持 |

結果見[畫面及保存檢查](hd-spell-scroll-runtime-20261014/check.json)與[GUI檢查](hd-spell-scroll-runtime-20261014/gui-check.json)。
服務路徑沿SPEC-040登記的準備舊存檔，再用既有操作入會、購買及編輯；沒有新增原版oracle。

單張卷軸以獨立`kol-hd-scroll/1`子組載入，核對固定來源圖號、檔名、數量、雜湊、實際尺寸及不透明性。
十項缺件、損壞、schema、來源、尺寸、路徑與透明度負例拒絕整組載入；繪製前另核實際`KOLPICS[1]`PNG指紋。
紙面用深墨，左右選單暫回深色面板墨色，離開時還原。法術規則、教派限制、名稱輸入與v3沒有改動。

GUI只在隔離副本加入Draw完成及頁面觀察，正式主程式不含診斷旗標或`/out`讀檔。
擷取先取X11原始畫格，再無損存PNG；本輪最長擷取0.34秒、PNG編碼15.97秒，畫格判準保持。
前置原型的提示消耗鍵、PNG編碼逾時、ESC與保存期望修正見[原型紀錄](hd-spell-scroll-20261014.md)；正式GUI乾淨執行通過。
音訊使用ALSA虛擬輸出，沒有新增音訊聆聽聲明。

素材與生成來源見[卷軸清冊](../../assets/hd/first-slice/scroll/manifest.json)及[generation.json](../../assets/hd/first-slice/scroll/generation.json)。
完整型別頁面、準備／保存檔及二進位只留本機；可提交收據只保留狀態雜湊與檢查結果。
來源見[376份輸入](hd-spell-scroll-runtime-20261014/formal-inputs.json)、[素材雜湊](hd-spell-scroll-runtime-20261014/asset-inputs.json)與[封存收據](hd-spell-scroll-runtime-20261014/receipt.json)。
頁面與卷軸清冊的status於驗收後更新；[前後雜湊](hd-spell-scroll-runtime-20261014/metadata-corrections.json)及原清冊保存，typed loader不讀status，全部呈現欄位與PNG保持，未重跑遊戲。
527份完整證據及三份程式雜湊另存[本機清冊](local/hd-spell-scroll-runtime-20261014/retention.json)。原工作區已逐檔核對後清理，解開沿[保存入口](retained-evidence.md)。
