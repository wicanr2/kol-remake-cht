# v.1.0.5-20261018 三平台雙版本引擎發布

驗收及正式發布：**PASS**。Release ID415941791已公開並設為最新版，七附件大小及SHA-256全符。

Release入口：[v.1.0.5-20261018](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.5-20261018)。
本機交付與公開附件以[六包校驗清冊](release-v1.0.5-20261018/SHA256SUMS.json)為準。

本輪依使用者要求更新README並發布Release。沿既有tag `v.1.0.5-20261018`，
固定`bc087d31183407907c754e252727b195d860aa9f`，不移動tag或重編正式程式。
從[已驗六個完整版](repack-v1.0.5-20261018.md)取出相同執行檔，建立三平台純EGA、EGA＋HD引擎包。
兩版分別要求912／1744份合法抽取資料，包內不含素材，完整版本仍留本機。

封裝入口為[pack_release_variants.py](../../tools/pack_release_variants.py)，
實際匯入操作沿[封包驗證工具](../../tools/verify_full_variants_runtime.py)的`--bundle release`。
來源與輸出集中於`dist-all/v.1.0.5-20261018/`，
本輪原工作區為`workplace/release-v1.0.5-20261018/`，封存核對後已清理。
來源、正式收據與遠端回讀保存於[本版證據目錄](release-v1.0.5-20261018/)。
[發布說明](release-v1.0.5-20261018/release-notes.md)沿同一版號與發布範圍。
[遠端核對工具](release-v1.0.5-20261018/recipe/check_publication.py)檢查七附件與匿名下載SHA-256。

公開範圍沿[權利審查](../RELEASE-RIGHTS-REVIEW.md)，只上傳六個無素材引擎包及SHA256SUMS.txt。
三平台使用已驗按需載入與快取釋放；Windows使用Wine，macOS只有結構驗證，均無真機驗收。
既有圖稿、配樂、規則與kol-save-v3保持，沒有新增原版parity或自然全任務通關聲明。

## 驗收結果

[正式收據](release-v1.0.5-20261018/receipt.json)核對435份正式來源與同版執行檔，
資料、工具及輸入指紋見[發布輸入](release-v1.0.5-20261018/release-inputs.json)與[工具鏈](release-v1.0.5-20261018/images.json)。
六包的manifest、匯入清單、授權、大小、SHA-256、平台、版本與素材排除均通過，見[包內核對](release-v1.0.5-20261018/package-content.json)。
原先六個完整版的SHA-256全部保持，本輪沒有重新編譯或修改素材。

[Windows／Wine](release-v1.0.5-20261018/windows-runtime.json)及[Linux](release-v1.0.5-20261018/linux-runtime.json)
四個實際引擎包匯入後通過開場、書封、目錄、建角選單進入／取消、組隊、旅店、F9及F10。
純EGA拒絕HD，EGA＋HD釋放後能重新載入，F10不改正式模型且過程不寫保存。
Linux另以`KOL_DATA_DIR`使用外部資料目錄。兩平台4 GiB且無交換空間，OOM及OOM kill皆0。
macOS兩包實際匯入912／1744份資料通過，沿已驗universal執行檔與bundle結構；沒有真機執行。

固定seed1的同一短流程24張圖與同版完整版逐一整圖零差，文字與模型逐位元組相同，見[比較收據](release-v1.0.5-20261018/comparisons.json)。
實際音訊峰值14664、非靜音、未裁切，退出末秒靜音；本輪未重做人耳驗收。
首次Wine腳本在HD首次繪圖時漏掉第二次固定延遲F9，遊戲仍存活而工具逾時。
改為等待操作及畫面回條後，以同版本、容器與seed重跑同組案例通過；正式遊戲不改。
失敗分類見[輸入時序收據](release-v1.0.5-20261018/input-timing-adjustment.json)，原始腳本與紀錄留本機。

工作清單及文件核對見[檢查紀錄](release-v1.0.5-20261018/worklist-check.log)。
原始畫格、錄音與失敗紀錄由[保存工具](release-v1.0.5-20261018/recipe/retain.py)封存，見[保存清冊](release-v1.0.5-20261018/retention.json)。
71份驗收檔案與3份可重建程式雜湊已逐檔核對，原工作區已清理；460個本機入口、工具語法、擁有權與專案衛生通過，見[交接檢查](release-v1.0.5-20261018/hygiene.json)。本輪容器無殘留。

## 遠端發布

[草稿回讀](release-v1.0.5-20261018/draft-remote.json)先核對七個授權附件的名稱、大小及GitHub digest。
公開後[完整回讀](release-v1.0.5-20261018/publication-remote.json)與[匿名下載驗證](release-v1.0.5-20261018/publication.json)核對全部HTTP200、大小與SHA-256，說明文字亦與本機相同。
台北發行日2026-10-18符合版號，正式tag仍固定bc087d31；沒有上傳完整版或影片，舊tag／Release保持。
發布後465個本機入口、新增遠端證據擁有權及專案衛生再次核對通過；本輪容器仍為0。
