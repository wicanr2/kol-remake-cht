# 三平台 HD 介面與配樂更新

版本：`v.1.0.2-20261013`，發行時區Asia/Taipei。狀態：六包、76秒影片及原定一色階閘門全部通過；使用者最終以go解除暫緩，[private Release](https://github.com/wicanr2/kol-remake-cht/releases/tag/v.1.0.2-20261013)已發布，四個遠端附件大小及SHA-256相符。tag固定指向`46a01f116f579806187bd7fbbfa6d0a5f7f6ce7d`，後續文件提交不移動tag。正式結果見[完整收據](release-v1.0.2-20261013/receipt.json)。

本版接續已核准的HD操作介面、人物頭像與四首室內樂，提供三平台完整版、引擎包及新版推廣片。前版保持，不修改玩法、v3格式或未核准的美術。以[配樂接入收據](hd-music-runtime-20261013/receipt.json)的361份Go／module為基準，只修改兩份HD呈現檔並新增整數取樣與測試，共363份。來源與正式三環境結果見[精度修正收據](release-v1.0.2-20261013/precision-fix.json)。

唯一交付根目錄為[dist-all/v.1.0.2-20261013/](../../dist-all/v.1.0.2-20261013/)。依本專案AGENTS.md沿用`full/`、`release/`、`promo/`與`smoke/`。完整版及影片僅本機保存；引擎包只含程式、工具、必要翻譯、條款與匯入清單，不含素材或完整手冊譯文。新版權利審查另由[發行審查](../RELEASE-RIGHTS-REVIEW.md)追蹤。

| 執行前固定閘門 | 判準 |
|---|---|
| 來源 | 363份Go／module與精度修正收據完全相同；只差兩份HD檔及新增取樣／測試，四首OGG與核准catalog雜湊相符 |
| 建置與回歸 | Go1.26.7、Ebitengine2.9.9；Linux AMD64、Windows AMD64、macOS Intel／ARM universal；正式套件通過，既有兩則vet警示另列 |
| 固定正常抽樣 | 沿前版唯一seed1鍵序重播一次；1046狀態及保存後13狀態重載；語意與完整v3除saved_at相同，不新增原版抽樣 |
| Linux封包 | 實際AppImage、匯入後引擎包啟動；正常旅店內F9、F10、鍵鼠及縮放，錄製實際配樂與關閉後靜音 |
| Windows封包 | 實際ZIP於Wine9啟動，正常旅店觸發配樂，F9與F10切換；EGA畫格精確比較，HD逐通道最多1色階的GPU取樣差另列 |
| macOS | universal、最低11.0、ARM ad-hoc簽章、系統動態相依與app執行位元；不宣稱真機執行 |
| 六包 | 逐檔內容、平台、必要素材、版號、UTF-8、SHA-256、授權及Release deny-list通過 |
| 影片 | 本版封包新畫格，76秒1080p／25fps、H.264／AAC；使用已核准原創室內樂與雷聲，無削波或長黑／長靜音；逐幕目視字幕與畫面 |
| 外部發布 | 僅指定private儲存庫的三個引擎包及checksum；完整版、原始資料及影片不上傳，舊tag／Release保持 |

來源工具：[交付流水線](../../tools/closeout_release.py)、[封包啟動](../../tools/verify_release_runtime.py)、[封包控制](../../tools/verify_package_controls.py)、[影片合成](../../tools/compose_release_promo.py)。全程在有界Docker內執行，原始來源唯讀，輸出UID/GID1000。建置、驗證與失敗分類追加至[WORKLOG](../../WORKLOG.md)。

Windows／Linux的[同主題實際封包對比](release-v1.0.2-20261013/review.html)與[來源清冊](release-v1.0.2-20261013/manifest.json)已保存。首次候選EGA十組逐像素相同，HD十組逐通道最大差2／255；未達預登記的一色階數值標準，原結果保留。中文、幾何與畫面目視相同。使用者當時選擇「維持一色階標準，暫緩發布」，首次候選未建立tag或Release，不把結構驗證當成發行就緒。下方修正後完整重建並通過數值閘門，使用者才在最後確認解除暫緩。

色差診斷沿用同一正式Windows PE、seed1、保存及五狀態鍵序，由[診斷工具](../../tools/diagnose_release_pixels.py)在Wine9內指定auto、DirectX及OpenGL後端。輸出保存於`dist-all/v.1.0.2-20261013/smoke/pixel-diagnosis/`；屬可撤回診斷，不變更正式渲染設定或原有數值閘門。

固定整數雙線性取樣由[隔離原型工具](../../tools/probe_hd_resampling.py)先建立在`workplace/hd-pixel-consistency-20261013/`。只修圖片取樣的數值可重現性，核准素材與位置保持。原型三十組、正式程式三十組均為EGA零差、HD最多一色階；[正式同主題比較](release-v1.0.2-20261013/fixed-sampling-comparison.png)已目視核對。像素中心、有理數權重、預乘RGBA與素材組內快取依[SPEC-039](../spec/039-hd-art-ui.md#hd圖片取樣的跨平台數值一致性)。正式套件及sceneverify通過，封包與正常GUI另驗。此工作區含首次未發布候選，完成比較並保存雜湊後清理。

診斷分級：相同Linux執行檔在Mesa22.3.6／25.2.8的畫格差兩色階已證實；相同Mesa下Linux／Windows最高一色階亦已證實。差異來自驅動取樣實作屬強推論，不宣稱已定位驅動內部指令。強制DirectX探針以-9退出且沒有畫格，結果未知；不當成產品故障或成功證據。後端選擇方式依[Ebitengine官方文件](https://ebitengine.org/en/documents/2.5.html)與容器內鎖定v2.9.9的`internal/ui/graphics.go`核對。

## 修正後的結果

| 項目 | 實際結果 |
|---|---|
| 正式來源 | 363份Go／module逐檔核對；Linux及Windows兩次獨立建置SHA-256相同，核准圖片與四首OGG保持 |
| 測試 | 正式套件及sceneverify通過；vet保留兩則既有未具名欄位警示，不宣稱全綠 |
| 固定正常路徑 | 相同seed1及唯一鍵序作修正後回歸，1046狀態／40PNG相同；新程序重載13狀態／3PNG相同，完整v3只排除saved_at |
| Linux實際封包 | 完整版AppImage及匯入後引擎包啟動、正常旅店、預設配樂、F9、F10、兩尺寸鍵鼠、遊戲狀態保持及實際音訊／關閉後靜音通過 |
| Windows實際封包 | full及release ZIP於Wine9啟動、正常旅店、預設配樂、F9、F10、狀態保持通過；實際PE音訊峰值12894，最後一秒為0 |
| 最終跨環境像素 | Debian Mesa22.3.6實際AppImage對Ubuntu Mesa25.2.8的兩個Windows ZIP，共二十組；EGA十組零差，HD十組逐通道最大1／255，未放寬閘門 |
| macOS | Intel／ARM universal、最低11.0、ARM ad-hoc簽章、app執行位元及僅系統動態相依通過，沒有執行或真機聲明 |
| 六包 | 逐包平台、版號、內容、素材、UTF-8、manifest、SHA-256及素材deny-list通過；使用者解除暫緩，四附件雜湊與授權相符，ready_for_publication為true |
| 影片 | 76秒、1920×1080、25fps、H.264／AAC；四首核准室內樂與雷聲，峰值−5.6dB，無指定閾值的長黑或長靜音；十幕中文及畫面比例目視通過 |

| 平台 | 本機完整版 | 需匯入資料的引擎包 |
|---|---|---|
| Linux AMD64 | [AppImage](../../dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-linux-x86_64.AppImage) | [AppImage](../../dist-all/v.1.0.2-20261013/release/kol-v.1.0.2-20261013-release-linux-x86_64.AppImage) |
| Windows AMD64 | [ZIP](../../dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-windows-amd64.zip) | [ZIP](../../dist-all/v.1.0.2-20261013/release/kol-v.1.0.2-20261013-release-windows-amd64.zip) |
| macOS Intel／ARM | [Universal ZIP](../../dist-all/v.1.0.2-20261013/full/kol-v.1.0.2-20261013-full-macos-universal.zip) | [Universal ZIP](../../dist-all/v.1.0.2-20261013/release/kol-v.1.0.2-20261013-release-macos-universal.zip) |

[76秒推廣影片](../../dist-all/v.1.0.2-20261013/promo/kol-v.1.0.2-20261013-promo.mp4)、[十幕檢查圖](../../dist-all/v.1.0.2-20261013/promo/contact-sheet.png)、[引擎包校驗碼](../../dist-all/v.1.0.2-20261013/release/SHA256SUMS.txt)及[完整交付雜湊](../../dist-all/v.1.0.2-20261013/SHA256SUMS.json)均已保存。完整遊戲、影片與原始音訊捕捉僅本機；公開收據不含原始保存或WAV。

首次候選及三環境修正的完整證據由[本機保存清冊](local/hd-pixel-consistency-20261013/retention.json)索引，原失敗及原雜湊保持。Linux僅解包執行，未驗主機FUSE；Windows僅Wine，macOS未執行。此結果不涵蓋所有GPU、完整正常通關或新增原版parity。
