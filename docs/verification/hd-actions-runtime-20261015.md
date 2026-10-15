# HD動作圖示正式接入

狀態：68個圖示的有限CONFORMED。[SPEC-039](../spec/039-hd-art-ui.md#動作圖示獨立子組)沿已核准的銅色Canvas圖示接入正式程式；完整物品／隊伍與裝備合成仍DRAFT。
[比較頁](hd-actions-runtime-20261015/review.html)提供來源PNG圖示展示、實際GPU原生圖示及正常旅店EGA／HD。[截圖清冊](hd-actions-runtime-20261015/manifest.json)區分展示與正常路徑，沒有新增原版DOS執行收據。

| 驗證 | 結果 |
|---|---|
| 素材與圖號 | 68張24×20來源，固定非連續集合3…27、31、36…69、74…81；單張1296×960原生圖集，每格144×120 |
| 嚴格載入 | 14項素材負例、實際來源PNG修改、nil來源、錯ZIP身分與70張archive拒絕；五項UI負例保持原子替換 |
| 實際GPU | 68圖示及12頁工具列，共14張完整畫格與V10原型最多一色階，alpha零差；來源展示完全相同 |
| 中文 | 原有五個非空工具列標籤皆繁中；其餘63項原本為純圖示，不新增推測名稱。空白鍵圖示改用「空白」，正常旅店顯示「繼續」 |
| 正常控制器 | 從標題進入Kenneth物品17槽及正常招募六型隊伍，共71狀態；修改前後及EGA／HD的完整trace與中文字紀錄相同 |
| EGA回歸 | 71張畫格RGB與alpha逐像素相同 |
| 遊戲內切換 | Shyla物品／隊伍、1920×1200及960×600、鍵鼠共24個F10設定事件；完整Inn與隊伍保持，HD往返零差，Ctrl+Q退出0 |
| 存讀檔 | 正常住宿11狀態、另程序重載5狀態相同，完整v3只差saved_at；修改前程式產生的舊v3亦可讀且檔案SHA保持 |
| 可重建性 | 既有Chromium映像重建圖集逐位元相同，最終manifest相同；正式384份Go／module、91份HD資產有完整雜湊 |

結果見[畫格](hd-actions-runtime-20261015/gallery-check.json)、[正常路徑](hd-actions-runtime-20261015/normal-check.json)、[GUI](hd-actions-runtime-20261015/gui-check.json)、[存讀檔](hd-actions-runtime-20261015/save-check.json)及[正式收據](hd-actions-runtime-20261015/receipt.json)。

`KOLMPICS[28]`實際為56×57城鎮徽章AND遮罩，依[RE-012](../re/RESEARCH-LOG.md)及`internal/town/outdoor.go`保留。先前69項計數包含此遮罩，現訂正為68；舊矩陣與coverage.json保持，勘誤見[工作歷程](../../WORKLOG.md)。

正式子組位於[actions/manifest.json](../../assets/hd/first-slice/actions/manifest.json)。啟用HD前驗完整圖集、原生RGBA與alpha範圍、授權檔及實際EGA庫71張圖號、68張來源PNG／RGB；全組成功才建立快取與換畫布。GPU首次繪圖再驗來源。圖示同時供原生資料頁與工具列；完整裝備頁尚未啟用原生HD呈現，不擴大本批結論。

重建入口為[build_hd_action_icons.py](../../tools/build_hd_action_icons.py)，圖形來源為[hd_action_icons.js](../../tools/hd_action_icons.js)，沿既有核准Canvas家族與Noto SIL OFL1.1。工具檔首列Docker映像、掛載、資源限制及命令。正式runtime只讀PNG與manifest，沒有原型絕對路徑、archive讀取或診斷旗標。

重跑方法保存於[tests.sh](hd-actions-runtime-20261015/tests.sh)、[gallery.sh](hd-actions-runtime-20261015/gallery.sh)與[gallery.go.txt](hd-actions-runtime-20261015/gallery.go.txt)、[正常路徑](hd-actions-runtime-20261015/normal.py)、[核對](hd-actions-runtime-20261015/normal-check.py)、[GUI](hd-actions-runtime-20261015/gui.py)及[保存](hd-actions-runtime-20261015/save.py)。在既有工具映像中將本工作區掛至`/out`，正式或凍結副本掛至`/src`；觀察式GUI只覆蓋main.go，不改正式主程式。來源由[正式384份](hd-actions-runtime-20261015/formal-inputs.json)、[V10原型382份](hd-actions-runtime-20261015/prototype-inputs-v10.json)及[HD素材](hd-actions-runtime-20261015/asset-inputs.json)定位。

首輪GUI已完成四組與24事件，短Ctrl+Q被遊戲取樣為單獨Q。修正自動化輸入為跨完整Draw保持Control後按Q，同一映像與相同驗證乾淨重跑通過；產品程式未改。最初命令列換行及將純圖示當成必有中文標籤的核對錯誤亦已訂正，不放寬一色階門檻。

僅驗Linux容器；舊Release保持。本工作區仍供裝備原型續作，尚未封存清理。完整角色模型、存檔、原錨點、原始archive及完整GUI狀態只留本機；四張PNG依現行private儲存庫例外保存。
