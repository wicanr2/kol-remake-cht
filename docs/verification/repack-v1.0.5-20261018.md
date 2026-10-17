# 三平台雙版本完整版重建

驗收狀態：**PASS**。六個完整版已完成，下載入口見 [README](../../README.md#開始遊玩)。

本輪入口為[正式收據](repack-v1.0.5-20261018/receipt.json)與[本機保存清冊](repack-v1.0.5-20261018/retention.json)。
來源與工具鏈見[重建輸入](repack-v1.0.5-20261018/build-inputs.json)及[映像檔指紋](repack-v1.0.5-20261018/images.json)。

| 平台 | 純 EGA | EGA＋HD | 驗證 |
|---|---:|---:|---|
| Linux x86_64 AppImage | 19,732,984 bytes | 1,270,413,816 bytes | 實際解包後經 AppRun 執行 |
| Windows AMD64 ZIP | 18,859,981 bytes | 1,292,965,138 bytes | 實際 ZIP 解壓後經 Wine 9 執行 |
| macOS universal ZIP | 26,369,876 bytes | 1,300,538,361 bytes | arm64／x86_64、arm64 簽章結構、macOS 11、系統動態庫 |

[六包清冊](repack-v1.0.5-20261018/SHA256SUMS.json)與[包內逐檔核對](repack-v1.0.5-20261018/package-content.json)記錄大小、SHA-256、manifest及資源數。
Linux／Wine四個實際封包均通過正常開場、書封、目錄、建角進入／取消、組隊、旅店、F9及F10。
分平台收據：[Linux](repack-v1.0.5-20261018/linux-runtime.json)、[Wine](repack-v1.0.5-20261018/windows-runtime.json)、[macOS結構](repack-v1.0.5-20261018/macos-structure.json)。
純EGA拒絕HD切換，完整圖像版可回到EGA釋放後再載入，F10不改變正式模型，過程不寫存檔。
兩平台均限制4 GiB且不使用交換空間，OOM與OOM kill皆0；包含軟體繪圖的容器峰值約3.63／4.29 GB，不等於遊戲程序或真機用量。

固定seed1的同一短流程共24張畫格，跨平台EGA零色差、HD最多一色階，文字與模型逐位元組相同，見[畫格比較](repack-v1.0.5-20261018/comparisons.json)。
角色選單、建角頁與書封中文已目視確認。腳本曾誤要求HD原生副標出現在點陣文字紀錄；
查證`plate`先呼叫`DrawNativeSubtitle`後返回，修正腳本預期，正式遊戲不改，詳見[調整收據](repack-v1.0.5-20261018/verification-adjustment.json)。
實際配樂輸出非靜音、峰值14031、未裁切，退出末秒靜音；本輪未重新人耳聆聽，見[音訊檢查](repack-v1.0.5-20261018/audio-check.json)。

本輪依使用者要求重建 `v.1.0.5-20261018` 三平台完整版，各有純 EGA、EGA＋HD 兩種，共六包。
435 份正式 Go／module 與[記憶體修復收據](windows-memory-20261017/receipt.json)完全相同；
1744 份資源沿用[前版固定清冊](release-v1.0.3-20261017/build-inputs.json)。
遊戲規則、圖稿、配樂及 kol-save-v3 格式保持，只重新編譯並注入新版號。

純 EGA 收錄 912 份資源，移除 `assets/hd/` 的 832 份圖片與配套，HD 選項停用。
EGA＋HD 收錄完整 1744 份資源，可用 F10 切換；兩版均保留四首配樂及獨立 F9。
三平台均使用按需解碼、容量受限快取與回到 EGA 時釋放圖片的正式程式。

## 重建入口

- [封包工具](../../tools/repack_full_variants.py)：依序執行 prepare、assemble、verify、finish。
- [編譯工具](../../tools/repack_full_build.sh)：native 編 Linux／Windows，macos 編雙架構。
- [實際封包操作](../../tools/verify_full_variants_runtime.py)：Linux AppImage 解包後經 AppRun、Windows ZIP 經 Wine 執行。
- 唯一輸出為 `dist-all/v.1.0.5-20261018/`；重建工作區為 `workplace/repack-20261018/`，完成後已封存並清理。

沿用 Go 1.26.7、Ebitengine 2.9.9、既有 package／osxcross／Wine 9 工具映像檔，
原始碼唯讀掛載，UID／GID 1000，網路關閉，每批容器均設定資源與逾時上限。
來源與既有收據相同，本輪以編譯、包內逐檔雜湊、實際封包操作及固定流程抽樣驗證。
前輪單元測試與素材深驗保留原收據，不聲稱本輪重新執行完整測試套件。

## 權利與驗證限制

六個完整版包含原作衍生資料，只留本機，不上傳 GitHub。
本輪未重做引擎 Release 或推廣影片；既有公開引擎包、76 秒影片與舊 tag 均保持。
Windows 使用 Wine，macOS 僅雙架構、簽章結構、最低版本與動態庫檢查，均無真機驗收。
Linux 以 AppImage 解包執行，未驗主機 FUSE。流程抽樣不代表自然全任務通關或新增原版 parity。

75份驗收檔案已壓縮保存，3份可重建執行檔僅保留雜湊；本輪工作區已逐檔核對後刪除。工具與保存方式見[證據索引](retained-evidence.md)。
封存入口為[本輪保存工具](repack-v1.0.5-20261018/recipe/retain.py)。[工作清單核對](repack-v1.0.5-20261018/worklist-check.log)及[交接檢查](repack-v1.0.5-20261018/hygiene.json)通過，本輪Docker容器無殘留。
