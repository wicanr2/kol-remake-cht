# 七段結局 HD 接入

狀態：**有限 CONFORMED**。使用者已核准首張候選及同畫風續作，七段均已接入開發版。
契約見[SPEC-039](../spec/039-hd-art-ui.md#七段結局手繪接入契約)，原版來源與首張候選見[前批比較](hd-completion-20261014.md)。
現行發行包仍為 v.1.0.2-20261013；本批未打包新版本。本批只驗七段呈現，正常完整通關仍不在本批聲明內。

## 畫面與來源

[七段比較頁](hd-ending-runtime-20261014/review.html)逐段列出原版 DOS、重製 EGA 與手繪 HD，包含結局內容。
HD 使用同一紅黑銅框、32px 原生中文字及原圖框，插畫等比置中。
第0段與核准候選完整畫格相同；其餘六段保留原人物、構圖與事件次序。

素材位於[結局清冊](../../assets/hd/first-slice/events/manifest.json)，
完整生成提示與工具路徑見[generation.json](../../assets/hd/first-slice/events/generation.json)。
本批使用內建 `image_gen.imagegen`，沒有改用 CLI。原圖來源固定為 `Knights-of-legend/ENDPICS.0`，
SHA-256 `7644948e4adb3ff45f54501f005de42c927f766c447404a5bc4bde8f6286896e`。
PNG、實際尺寸、來源指紋、圖號、權利分類及執行畫格由[截圖清冊](hd-ending-runtime-20261014/manifest.json)索引。

原版沿用前批鎖定 dosgolem 968ca774 的七段受控執行收據，未重新抽樣亂數。
呼叫前 seed1，段內亂數呼叫0次；原版圖片矩形皆符合原始解碼 PNG。
原版白底黑字與重製版紅黑底分開標示，不宣稱整頁原版 parity 或自然通關。

## 正式接入與驗收

| 檢查 | 結果 |
|---|---|
| 素材載入 | 七圖完整驗證雜湊、PNG尺寸、圖號、來源指紋後原子載入；缺圖、損壞、錯尺寸、重號、重複指紋及不安全路徑負例拒絕 |
| 中文呈現 | 七段真實譯文擷取通過；缺字、畫布界線與圖片重疊檢查通過，逐段目視可讀 |
| EGA 回歸 | 正式英文／繁中共14張結局畫格與來源基底 bd9a2045 相同 |
| 正常保存 | 正式程式從標題沿既有固定輸入走1046狀態，再以新程序重載13狀態；EGA基底與本批HD trace、完整v3除saved_at均相同 |
| 遊戲內切換 | 七段各一次 EGA／HD 往返，末段另驗960×600；共48個設定事件，七個不同段落狀態均保持 |
| 圖像復原 | HD往返完整畫格逐通道最多一色階，第0段為1，其餘六段為0；維持使用者原訂標準 |
| 工具鏈 | Go1.26.7、Ebitengine2.9.9，鎖版package image；正式366份Go／module清冊保持，全套internal／cmd及sceneverify測試通過 |

詳細結果見[檢查收據](hd-ending-runtime-20261014/check.json)、[封存收據](hd-ending-runtime-20261014/receipt.json)與[正式輸入清冊](hd-ending-runtime-20261014/formal-inputs.json)。
真正視窗的切換驗證使用隔離副本，僅注入起始結局狀態；正式呈現、字模、素材、F10及輸入程式保持。
標題至保存與新程序重載使用正式程式。此方法依 SPEC-040 驗場景，不證明正常全旅程或其他事件。
本批只在 Linux 容器執行；Windows／macOS與其他12類HD仍未完成。

## 重生與保存

工作區名稱為 `hd-ending-runtime-20261014`。比較腳本、完整輸出、失敗紀錄、GUI隔離差異及執行檔雜湊保存在
[本機保存清冊](local/hd-ending-runtime-20261014/retention.json)，核對與解開方式見[證據保存入口](retained-evidence.md)。
還原後依序執行 `build.sh`、`build-gui.sh`、`verify.sh`，沿用原鎖版Docker、唯讀專案及可寫輸出目錄。
正常HD流程初次200秒預算不足，沿同一輸入改600秒乾淨重跑通過；原逾時輸出保留為工具預算限制。
原始ZIP／EXE／archive不進Git或Release，插畫及截圖只依目前private儲存庫範圍保存。
