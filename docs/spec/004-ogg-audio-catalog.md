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

## 尚未通過的 READY gate

1. 逐一掃描其餘 executable 的 speaker／PIT 與可聽事件路徑，關閉 cue 母體。
2. 以正常玩家路徑確認各 cue 的觸發情境；caller 不可只靠導覽標籤推定。
3. 由使用者決定已證實演算法 cue 採「規格近似合成 OGG」、新作替代音效或靜音；目前不代決。
4. 固定 OGG 編碼參數、loop 契約、權利分類與人耳抽樣方式。
5. 證據審查後才可把本規格升為 READY。
