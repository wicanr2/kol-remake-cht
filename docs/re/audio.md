# 音訊

**這款遊戲只有一個音效。**整份原版沒有音樂、沒有語音、沒有音效檔——六個
executable 裡只有 `TITLE.EXE` 碰得到音訊硬體，而它發出的是標題畫面翻頁時的一聲
雷響。知道這件事之前，很容易花很多力氣去找不存在的音樂檔。

## 怎麼確定「只有一個」

不是找不到就當作沒有，是**把母體窮舉完**（RE-033）。x86 要發聲一定得走 `IN`／`OUT`
指令碰 I/O 埠，所以拿 IDA 掃六個 executable 裡每一條已定義的 `IN`／`OUT`：

| executable | I/O 指令數 | 音訊候選 |
|---|---:|---:|
| KOL.EXE | 6 | 0 |
| KOLBOOK.EXE | 43 | 0 |
| KOLTOWN-unpacked.EXE | 54 | 0 |
| MAKEMODE.EXE | 0 | 0 |
| TANDYPIN.EXE | 0 | 0 |
| TITLE.EXE | 12 | **6** |

只有 `TITLE.EXE` 的六條指令命中音訊埠 `0x61`（speaker gate）與 `0x42`
（PIT channel 2）。其餘六個「保守未解析」的 `DX` 間接 I/O 點逐一人工看過，全都在
既有的 EGA 圖片 renderer 垂直鏈上，是 graph-register 寫入，不是音訊。

每個 executable 都用**兩個全新的 IDA DB** 各跑一次，兩份輸出逐 byte 相同。
安全摘要在 `ida-fingerprints/*-audio-io-summary.json`，目錄在
[audio-inventory.json](audio-inventory.json)。

**`TANDYPIN.EXE` 不是音訊播放器。**名字看起來像 Tandy 音效驅動，實際上入口就是
`INT 21h`／`AH=31h`，以 `DX=0x0401` paragraphs 常駐記憶體，入口沒有任何 speaker
或 PIT I/O（RE-032）。這是二手推論（檔名）輸給一手證據（指令）的例子。

## 那一聲雷是怎麼發的

`TITLE.EXE` 的三支：

| 函式 | 做什麼 |
|---|---|
| `sub_10F33` | 讀 `0x61`、設 bit 0 與 1，打開 speaker gate |
| `sub_10F48` | 清掉那兩個 bit，關掉 speaker |
| `sub_10F5D` | 把除數寫進 PIT channel 2（`0x42`），決定頻率 |

`sub_10F75` 是 cue 本體：**兩段音爆，每段更新 100 次除數**，除數由
`sub_11953`（`rand() % N + 1`）隨機取，每次之間用 `sub_11904` 忙等。
隨機掃過的頻率聽起來就是一陣雜訊般的隆隆聲——雷。

八個呼叫點全在 `_main`、`sub_10AF0`、`sub_10C86`，而且**每一個都緊接
`sub_11133` 的圖片 blit**，之後進 `_kbhit`。所以它是標題序列換頁時的音效，
不是背景音樂（RE-033）。

## 原版實際聽起來如何

用 DOSBox-X 2026.06.02 的官方 `DX-CAPTURE /A /-D` 契約，跑**正常的** TITLE 流程
錄下來（不是叫函式，是照玩家會看到的順序跑）：

- 48 kHz、stereo、PCM，38.229333 秒
- 四段非靜音窗，長度約 0.935、1.057、0.942、0.910 秒
- peak −10.317424 dB

明細在 [title-audio-oracle.json](title-audio-oracle.json)。WAV 本身是本機 oracle，
不進 Git。

**機械擷取證實「有聲音、不是靜音」，不等於人耳驗收。**DOSBox-X 的時序也不冒充
原機的 wall-clock。

## remake 怎麼處理

使用者於 2026-09-01 決定走**規格近似合成**：依公開的 PC Speaker／PIT 規格與上面
已證實的演算法自行合成波形，**不散布原版錄音**。

理由是權利：錄下來的 WAV 是原版執行結果的轉換產物，著作權仍屬原權利人；自行合成的
波形則是 remake 自有、可公開。這個分界和圖像那邊不同——圖像沒有「用規格重畫」這條路，
所以只能留在 private repo。

已產出的候選是 `assets/audio/title-thunder.ogg`：時長 0.961451 秒，與原版四段非靜音
窗的平均值差 0.000581 秒。編碼契約是 Vorbis／44,100 Hz／單聲道／`ffmpeg libvorbis
-q:a 5`，不循環，峰值對齊擷取量到的 −10.317 dB。

### 硬體時序的停止線

**不追逐逐週期的 PIT 與忙等 parity。**只依公開規格與成熟模擬器的契約求出合理的
sample duration 與「播放完成」的控制流閘門，標記為 `hardware-spec approximation`。
玩家看得見的 cue 與等待播放完成的行為證實之後就停止深挖。

## 現況與還沒過的閘門

**遊戲裡目前一個音都沒有。**SPEC-004 仍是 DRAFT，因此還沒建立 Ebitengine 的音訊
路徑——規格沒到 READY 就不接進 runtime，這是專案的硬規則。

擋在 READY 前面的是一件事：**人耳驗收還沒做**，需要使用者實際聽過
`assets/audio/title-thunder.ogg`。技術上解得開、不是靜音、時長對得上，都不代表
聽起來像那一聲雷。

驗收通過之後要做的：升 SPEC-004 為 READY → 建立 OGG 轉換／合成管線的驗證收據 →
接上 Ebitengine 的播放路徑 → 在標題序列的三個呼叫點觸發。

## 相關

- 研究時間線：RE-032（cue 與 TANDYPIN 排除）、RE-033（I/O 母體與 WAV oracle）
- 規格：[../spec/004-ogg-audio-catalog.md](../spec/004-ogg-audio-catalog.md)
- 標題序列本身怎麼跑：RE-044、RE-047
