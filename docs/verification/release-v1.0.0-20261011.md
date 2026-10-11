# v1.0.0 本機收尾交付

版本：`v.1.0.0-20261011`。狀態：本機交付驗收通過。

使用者於 2026-10-11 授權判斷目前狀況，最多補一次抽樣，製作完整版、release 及推廣影片。
本次沿現行 EGA／Go 正式來源，不更改規則、保存格式、HD 或加速設定。
完整原版 parity、全部中文字版面與平台真機的既有未驗項保持。

交付入口為 [dist-all/v.1.0.0-20261011/](../../dist-all/v.1.0.0-20261011/)。
專案 AGENTS.md 明定本專案使用 `full/`、`release/`、`promo/`，因此沿用既有名稱；
`smoke/` 保存驗收，根目錄保存 `build-inputs.json` 與 `SHA256SUMS.json`。
完整版含遊戲資料，僅本機使用。release 是需匯入合法資料的引擎包。
影片含原作衍生畫面，亦留本機；本次沒有 GitHub Release、tag 或外部上傳。

| 項目 | 預先固定的完成條件 |
|---|---|
| 正式來源 | 342 份 Go／module 輸入與最近平台收據逐項相符 |
| 唯一抽樣 | seed1，繁中正常任務戰鬥、逃離、自然遭遇、回城住宿保存、新程序重載 |
| 比較 | 與最近正式來源的完整 trace 相同；v3 除 saved_at 相同；重載不改保存 |
| 硬失敗 | 崩潰、卡住、狀態差異、保存損壞即停止交付，不換種子重抽 |
| 封包 | Linux AppImage、Windows AMD64、macOS universal，各有 full／release |
| release | 無原版素材；資料匯入先核全部 SHA-256，缺檔或不同版本拒絕 |
| 影片 | 本版實際畫格、逐幕來源、codec／音量／黑幀／靜音／凍結檢查及目視 |
| 清理 | 本專案容器歸零、UID/GID 正確、既有研究工作區保持 |

抽樣基準沿 [任務亂數正式驗收](quest-rng-20261007.md)，不增加原版覆蓋，也不宣稱全程通關。
抽樣、封包與影片收據保存於交付目錄，不把生成二進位或原版資料加入 Git。

| 實際驗收 | 結果 |
|---|---|
| 正式來源與測試 | 342 份輸入與最近平台收據相同；十九套件通過 |
| 唯一正常抽樣 | 1046 狀態、40 PNG；trace 完全相同；完整 v3 只差 saved_at |
| 同一鏈的新程序重載 | 13 狀態、3 PNG；trace 相同、輸入存檔未改 |
| 六個封包 | 三平台各有 full／release；執行檔、完整資料、版號、條款、ZIP UTF-8、BOM 與 macOS 執行位元通過 |
| release 匯入 | 全資料先核 SHA-256；錯誤雜湊在寫入前拒絕；Linux／Wine 啟動成功 |
| Windows | 實際 ZIP 解壓，full 與匯入後 release 版號相同、中文封面 PNG 相同；Wine9，非真機 |
| macOS | ARM64／Intel、最低11.0、ARM64 ad-hoc 簽章與系統相依通過；沒有執行 |
| 影片 | 64秒、1920×1080、25fps、H.264／AAC；峰值 -9.8dB，八幕中心與字幕目視通過 |
| 音畫邊界 | 沒有達門檻的黑幀；截圖停留6至10秒及雷聲外靜音依分鏡，不冒稱連續實機錄影 |
| manifest | full／release／promo 逐檔獨立重算，包含巢狀來源 manifest；根目錄 SHA256SUMS.json 保存全部輸出 |

正式 Linux 執行檔 SHA-256：`12f62105ad16117689d3650c66bddf675afe5d31eb003c10f3c6f4eef550a6ce`。
Windows：`610fd057143da3d808ae705514cd7aed177ddc71213c688c361a1902db62c667`。
macOS universal：`336b9d7499e3099c61fd322721b0bcc69fdf5ebffea96c69360c6130d5956d78`。
六個可交付封包與影片的完整雜湊以交付根目錄 `SHA256SUMS.json` 為準。

可重現入口為 [closeout_release.py](../../tools/closeout_release.py)，資料匯入入口為
[import_game_data.py](../../tools/import_game_data.py)，影片入口為
[compose_release_promo.py](../../tools/compose_release_promo.py)。所有分析、建置與驗收只能在有界 Docker 中執行。
來源掛 `/src:ro`，全新版本目錄掛 `/out`，目前 UID/GID，預設 `--network none`。
依序執行 `prepare --git-head <HEAD>`、`native`、macOS 雙架構建置、`assemble`、`sample`、影片合成、`verify`。
`sample` 需有已就緒的 Xvfb、DISPLAY 及 XAUTHORITY。Xvfb 由有界容器的擁有程序及 trap 回收。
不得重跑此 phase 挑結果。正式測試只選 `./cmd/... ./internal/...`，避免掃進本機研究原型。

工具鏈沿 `docker/Dockerfile.package` 建立
`kol-package:1.26.7-ebiten2.9.9-20261011-r1`，取代已不存在的 `kol-package:20260907`。
它使用現有 Go1.26.7／Ebitengine2.9.9 驗證 image 與固定 SHA-256 的 AppImage runtime。
macOS 沿 `kol-osxcross-go:20261007-r1`；Windows 啟動沿 `kol-wine-verify:ubuntu-noble-20261007-r1`。
Linux 以解包模式驗收，Wine 不是真機，macOS 只核結構。影片僅使用已接受的遊戲雷聲，不新增背景曲。

正式十九套件測試通過。`go vet` 有兩則既有警示，位於 `internal/ui/bitmap_editor.go` 的 56、57 行，
原因是 `book.Label` 使用未具名欄位。警示原文保存在 `smoke/vet.log`，不宣稱靜態檢查全綠。
本次保留正式來源與既有驗收可比性，不為封包另改 UI 程式。
