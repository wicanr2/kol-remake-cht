# 試玩修正、操作影片與三平台完整版

驗收狀態：PASS。本機版本為 `v.1.0.6-20261024`，原始碼提交為 `a091378b1e20ba9be760b98c272e2236926cd3dd`。
本版六個完整版及83秒影片只留本機，沒有建立GitHub Release或移動舊tag。
入口見[README](../../README.md#開始遊玩)，[正式收據](repack-v1.0.6-20261024/receipt.json)記錄封包與平台範圍。

## 本版內容

- 自動讀回住宿保存，人物及裝備頁以數字切換，名稱保留大小寫。
- 商店販賣及清單點選、法術修改、騎馬遭遇方向鍵、武器攻擊上中下選擇。
- 地圖長按連走，等待頁要求放開方向鍵，滾輪不略過遭遇與戰鬥訊息。
- 護甲鐵砧既有功能完成實際點選驗證。

規格與測試見[SPEC-041](../spec/041-player-feedback-fixes.md)及[試玩修正驗證](player-feedback-20261024.md)。
443份Go與module輸入逐檔對應該輪通過收據，1744份素材與既有固定清冊相同。
本輪沒有變更正式程式、kol-save-v3、圖稿或配樂，只重編、注入版號並分裝。
錄影工具調整後，另以實際封包中的Linux執行檔重跑雙語鍵鼠流程，
[回歸收據](repack-v1.0.6-20261024/gui-regression.json)通過。

## 封包與操作驗收

[六包大小與SHA-256](repack-v1.0.6-20261024/SHA256SUMS.json)、
[包內manifest及逐檔檢查](repack-v1.0.6-20261024/package-content.json)全部通過。
純EGA保留912份資源，移除832份HD圖像與配套；EGA＋HD保留1744份。
兩版都有四首配樂、F9及本次修正，HD版可用F10切換。

Linux實際AppImage解包經AppRun、Windows實際ZIP經Wine9，兩種圖形版本都通過
開場、建角選單、組隊、旅店、F9、F10及無`-load`續玩抽測。
續玩使用實際建角住宿產生的保存，兩名隊員與各1440金冠保持，檢視後存檔逐位元組不變。
Linux驗證使用虛擬ALSA，Wine使用獨立PulseAudio錄製實際輸出。
兩平台各限4GiB且無交換空間，OOM kill均0。
固定seed1抽樣24張圖，跨平台EGA最大差0、HD最大差1，文字與模型相同。
音訊非靜音、峰值14854、未裁切，退出末秒靜音；本輪未重新人耳聆聽。

macOS為arm64／x86_64雙架構，最低11.0、arm64簽章結構與系統動態庫通過。
Windows與macOS尚無本輪真機驗收，Linux未驗主機FUSE。
本輪抽樣不代表自然全任務通關或新增原版parity。

## 操作推廣片

[83秒1080p影片](../../dist-all/v.1.0.6-20261024/promo/kol-v.1.0.6-20261024-feedback-promo-1080p.mp4)
含11段實際鍵鼠操作：[章節](repack-v1.0.6-20261024/promo-chapters.json)、
[錄影收據](repack-v1.0.6-20261024/recording.json)、[影音檢查](repack-v1.0.6-20261024/promo-metadata.json)。
建角、住宿及重開依一般互動入口操作；商店、法術、野外及戰鬥使用準備資料，片中標示受控場景。
影片呈現販賣、護甲減重、法術修改、連走、騎馬選項、上中下攻擊部位及HD切換。
使用已核准標題室內樂，技術檢查涵蓋解析度、25fps、音軌、非靜音、峰值、黑幀、凍結及字幕邊界。
靜態字卡及停留閱讀頁會觸發凍結偵測，另以章節抽樣及操作紀錄核對。

## 重建與保留

- [三平台編譯](../../tools/repack_full_build.sh)、[六包分裝](../../tools/repack_full_variants.py)、[實際封包驗證](../../tools/verify_full_variants_runtime.py)。
- [操作錄影](../../tools/verify_feedback_gui.py)、[影片合成](../../tools/compose_feedback_promo.py)。
- Go1.26.7、Ebitengine2.9.9及[四個既有映像](repack-v1.0.6-20261024/images.json)，全程容器、原始輸入唯讀、UID/GID1000、無網路且有資源與逾時上限。
- 舊工作區及本輪工作區依[保存入口](retained-evidence.md)完整封存、逐檔核對後清理。原始輸入、舊交付與本版六包及影片保留。

錄影曾因固定等待時間在高負載下漏鍵，改成接收回執後放鍵；文字仍短按。
驗證另修正ALSA設定及共用中的保存副本，改用固定住宿保存；同封包乾淨重跑通過。
這些失敗及原始錄音、畫格與工具副本保存在本機完整封存，不列作產品缺陷。

完整封存實際還原通過，見[還原抽測](repack-v1.0.6-20261024/restore-smoke.json)及[跨目錄／權限／連結測試](repack-v1.0.6-20261024/restore-tool-tests.json)。

清理結果見[空間與保存清冊](repack-v1.0.6-20261024/cleanup-summary.json)及[衛生檢查](repack-v1.0.6-20261024/hygiene.json)。工作區已清空，完整封存及可讀副本約1.71GiB；原始輸入與交付包保持。目錄配置差額不作共用主機free空間增幅。

舊比較頁另保存圖片及資料依賴，逐檔核對後保留可直接查看的12頁，見[比較頁核對](repack-v1.0.6-20261024/readable-pages.json)。相對連結調整的HTML另記原始及可讀副本雜湊，原始封存不改。
