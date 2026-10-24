# v1.0.7 兩類完整版與更新推廣片

狀態：PASS。版本`v.1.0.7-20261024`，遊戲來源`308ecfdf`。
交付根目錄為`dist-all/v.1.0.7-20261024/`，完整版及影片只留本機。

本輪使用[三平台建置](../../tools/repack_full_build.sh)、[雙版本分裝與內容核對](../../tools/repack_full_variants.py)、
[實際封包操作](../../tools/verify_full_variants_runtime.py)與[更新影片合成](../../tools/compose_update_promo.py)。
影片重新錄製v1.0.6修正及v1.0.7新增操作，分鏡、來源與影音檢查隨交付保存。
既有[SPEC-041](../spec/041-player-feedback-fixes.md)、[SPEC-042](../spec/042-assist-toolbar.md)及
[工具列驗收](assist-toolbar-20261024.md)為功能入口，不重開原版未知研究。

## 六個完整版

兩類實體目錄為`full/ega/`與`full/ega-hd/`，各有Linux x86_64 AppImage、Windows AMD64 ZIP、
macOS ARM64／Intel通用ZIP。純EGA約19至27 MB，EGA＋HD約1.27至1.30 GB。
兩類都包含v1.0.6修正、v1.0.7工具列、可選輔助、永久角色強化、中文及四首配樂。
純EGA不含832份HD圖像及配套檔案，保留912份資源；HD版含1744份資源。

[正式收據](repack-v1.0.7-20261024/receipt.json)、[六包SHA-256](repack-v1.0.7-20261024/SHA256SUMS.json)及
[包內逐檔核對](repack-v1.0.7-20261024/package-content.json)全部通過。
程式由453份已驗收Go／module／內嵌字模輸入乾淨重編，再注入同一完整版號；素材逐檔對應固定清冊。
OFL字型授權與來源清冊隨包保存。內含manifest、平台／架構、版本、必要素材、完整授權及原版資料deny-list已核對。
內容與實際啟動驗收後，六包移入兩類目錄，[分類收據](repack-v1.0.7-20261024/category-layout.json)
記錄移動前後SHA相同；打包腳本已直接輸出此結構，實際執行工具由校驗碼清冊定位封包。

## 實際封包與保存

[Linux](repack-v1.0.7-20261024/linux-runtime.json)實際AppImage解包經AppRun、
[Windows](repack-v1.0.7-20261024/windows-runtime.json)實際ZIP經Wine9，兩類都通過：

- 開場、建角選單、選角與旅店、F9配樂及F10切換；純EGA正確拒絕缺少的HD。
- F8工具列預設關閉輔助、玩家不疲勞／敵疲勞6倍／免怯戰及歸零。
- 從實際建角、住宿產生的保存不帶-load續玩，兩人及各1440金幣保持；檢視不改寫保存。

兩平台各在4GiB記憶體上限且無交換空間下執行，OOM與OOM kill皆0。
固定seed1共24張抽樣，EGA最大像素差0、HD最大差1，文字與模型相同，見
[比較收據](repack-v1.0.7-20261024/comparisons.json)。
[實際音訊](repack-v1.0.7-20261024/audio-check.json)由隔離PulseAudio錄製，非靜音、峰值14052、未裁切，退出末秒靜音。
本輪沒有重新人耳聆聽。

[macOS結構](repack-v1.0.7-20261024/macos-structure.json)通過雙架構、最低11.0、ARM64簽章結構與系統動態庫核對。
Windows／macOS沒有本輪真機驗收，Linux沒有主機FUSE驗收；這些抽樣不代表原版parity或自然全任務通關。

## 更新推廣影片

影片為124秒、1920×1080、25fps，17段實際鍵鼠章節加首尾字卡。v1.0.6段落展示建角大小寫、
住宿與自動重載、人物／裝備數字切換、販賣、鐵砧、法術修改、長按移動、騎馬遭遇、攻擊部位與HD切換。
v1.0.7段落展示輔助控制、永久屬性與技能200／金幣9000、住宿扣款保存、重開輔助關閉，以及系統快捷鍵。

兩組錄影都來自本輪相同Linux正式程式，與包內執行檔SHA相符，沒有重用前版影片。
遊戲畫面完整等比縮放，保留工具列及原有視口；商店、法術、野外及攻擊的準備資料在片中標為受控場景。
配樂使用正式遊戲中已核准的標題室內樂，來源與SHA依音訊manifest核對，沒有聲稱錄影原音。
完整影片、章節、FFprobe、抽樣畫格、contact sheet及影音metadata保存於本版`promo/`。
[124秒影片](../../dist-all/v.1.0.7-20261024/promo/kol-v.1.0.7-20261024-updates-promo-1080p.mp4)、
[章節](repack-v1.0.7-20261024/promo-chapters.json)、[影音驗收](repack-v1.0.7-20261024/promo-metadata.json)及
[錄影來源一](repack-v1.0.7-20261024/recording-feedback.json)／[二](repack-v1.0.7-20261024/recording-assist.json)可回查。
全片音軌非靜音、峰值12585、無裁切；遊戲區黑幀檢查通過。19章各抽五張畫格，17段操作皆有畫面變化，
字幕邊界與contact sheet目視通過；靜態字卡、暫停及閱讀造成的凍結偵測由章節與操作紀錄區分。

## 保留與清理

完整版與影片不加入Git或GitHub Release，已發布tag／Release保持。
依使用者只留最新版的決定，新版全數通過後已刪除v1.0.6本機交付約3.98GB，見
[清理清冊](repack-v1.0.7-20261024/prior-local-prune.json)。本輪381項工作資料完整封存並逐檔驗證後清除，
[本機保存清冊](local/release-v1.0.7-20261024/retention.json)記錄原始271014253 bytes、封存101410092 bytes，沒有省略檔案。
workspace已空。核對及還原沿用[保存入口](retained-evidence.md)的工具，還原名稱為`release-v1.0.7-20261024`。
研究輸入、原始素材及既有研究封存保持。收尾狀態依[工作紀錄](../../WORKLOG.md)及[目前狀態](../../CONTEXT.md)。
