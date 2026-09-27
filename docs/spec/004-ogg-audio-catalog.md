# SPEC-004：OGG 音訊目錄與執行期契約

狀態：**DRAFT**  
是否授權實作：**否**

## 目的

把原版可聽事件轉成可追溯的 OGG 資產目錄，並保證 remake 執行期只讀 OGG，不直接模擬
DOS 音訊硬體或讀取原版 executable。此規格未達 READY 前，不建立正式轉檔器、合成器或
Ebitengine 音訊路徑。

## 已固定證據

- 固定 ZIP 沒有獨立的 WAV、VOC、MIDI 或已辨識音訊 archive。
- `TITLE.EXE` 的 `sub_10F75` 已證實控制 PC Speaker 與 PIT channel 2：共兩段音爆，每段
  更新 100 次除數；`sub_11953` 是 `rand() % N + 1`，`sub_11904` 是依全域校正常數忙等。
- `TANDYPIN.EXE` 入口執行 DOS `INT 21h`、`AH=31h` 常駐記憶體，保留 `0x0401`
  paragraphs；它不是獨立音訊素材或音效播放器。
- 六個 executable 均以兩個全新 IDA DB 窮舉已定義 code 的 x86 `IN`／`OUT`。只有
  `TITLE.EXE` 命中 `0x61`／`0x42`；其餘五個沒有音訊候選。六個保守未解析的 `DX`
  I/O 點經人工審查全是既有圖片 renderer 的 EGA graph-register writes。
- `sub_10F75` 的八個 caller 全在 `_main`、`sub_10AF0`、`sub_10C86`，且都緊接
  `sub_11133` 圖片 blit；單次 cue 路徑隨後進入鍵盤輪詢。這證實 cue 位於正常 TITLE
  presentation flow，但尚未把各張畫面推定成原版語意名稱。
- 依 DOSBox-X 官方 `DX-CAPTURE /A` 契約，在正常 `TITLE.EXE` 流程取得 48 kHz、stereo、
  PCM WAV。38.229333 秒觀察窗內有四段非靜音輸出，約為 0.935、1.057、0.942、0.910 秒；
  peak 為 -10.317 dB。這是 emulator machine capture，不冒充人耳驗收或原硬體時鐘 parity。
- 原始定位、bytes、caller 與工具版本見 `docs/re/ida-fingerprints/`；目錄見
  `docs/re/audio-inventory.json`。

## DRAFT 資產契約

每個可交付 cue 至少記錄：穩定 ID、玩家觸發條件、原版函式／位址、證據等級、生成方法、
取樣率、聲道、時長、loop、音量基準、來源權利、OGG SHA-256 與人耳驗收結果。檔名只用
穩定 ID，不把推測性語意當作原版名稱。

執行期允許副檔名只有 `.ogg`；原版 EXE、archive、即時 PIT 合成與未列入 catalog 的檔案
一律失敗即關閉。技術解碼通過不等於人耳驗收。

## 硬體時序停止線

不再反組譯 DAC／PIT／busy-wait 的逐週期 wall-clock。若重建 TITLE cue，只依公開 PIT
規格與成熟模擬器契約求合理的 sample duration 與播放完成閘門，並標記為
`hardware-spec approximation`；不得宣稱逐波形或實機時鐘 parity。

## 已由使用者決定的路線

2026-09-01：TITLE 的可聽事件採**規格近似合成 OGG**，不散布原版錄音。

## 已固定的編碼與權利契約

- 編碼：Vorbis、44 100 Hz、單聲道、`ffmpeg libvorbis -q:a 5`。
- loop：TITLE 雷聲不循環；每次觸發播放一次。
- 音量基準：合成峰值對齊 DOSBox-X 擷取量測的 −10.317 dB。
- 權利分類：remake 自有、可公開——波形由公開的 PC Speaker／PIT 規格與已證實的演算法
  產生，沒有取樣或轉換任何原版音訊。
- 目錄與逐 cue 欄位在 `assets/audio/manifest.json`，技術驗收收據在
  `docs/verification/audio/`。

## READY gate（人耳驗收由 #4 追蹤）

1. **人耳驗收尚未執行**，需使用者實際聆聽 `assets/audio/title-thunder.ogg`。技術解碼
   通過不等於聽起來對。
2. 證據審查後才可把本規格升為 READY；在那之前不建立 Ebitengine 執行期音訊路徑。

## 目前已產出的候選資產

`assets/audio/title-thunder.ogg`：`sub_10F75` 的規格近似合成，時長 0.961451 秒，與原版
DOSBox-X 擷取四段非靜音視窗的平均值差 0.000581 秒。這是**待驗收的候選**，不是已授權的
執行期資產。
