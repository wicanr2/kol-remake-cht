# 逆向研究紀錄

## RE-001：固定原版輸入與封裝身分

### 目標

- 遊戲／build：*Knights of Legend*，封裝確切來源與修改歷史未知
- 平台：DOS（`強推論`；來自 MZ executable 與安裝批次檔，尚未動態啟動）
- 原始資料位置：專案根目錄，Docker 唯讀掛載 `/kol`
- 可寫實驗位置：尚未建立；本輪沒有解包到工作樹
- 工具：Python 3.13 standard library（ZIP／hash／header）、Poppler `pdfinfo` 22.12.0、
  Pillow 11 image metadata；全部在 `--network none` 一次性 Docker 容器執行

### 問題

目前有哪些固定研究輸入？ZIP 是否完整？內含哪些 executable 與資料家族？

### 先前主張

| 主張 | 來源 | 證據狀態 | 衝突 |
|---|---|---|---|
| ZIP 是 DOS 版《武士傳說》 | 檔名與使用者提供內容 | 假說 | 尚未固定 executable 與版本 |
| 素材可抽成 PNG | 專案目標 | 待證實 | 尚未知來源格式與 palette |

### 證據

| Artifact | 觀察 | 獨立檢查 |
|---|---|---|
| `Knights-of-legend.zip`，SHA-256 `2b455e…9f46` | 76 項、無 comment，逐項 CRC 通過 | Python `ZipFile.testzip()` 回傳 `None` |
| ZIP 內 6 個 `.EXE` | 全部以 `4d5a`（MZ）開頭；逐檔雜湊見 generated manifest | 待 IDA Pro 9.4 header／segment 重驗 |
| ZIP timestamps | 多數是 1989；少數 gameplay／資料檔為 1998 | manifest 保留逐項 DOS timestamp |
| manual PDF，SHA-256 `4c80f5…7086a` | 79 頁、PDF 1.6、379.776×600.096 pt | Poppler 22.12.0 |
| reference card，SHA-256 `4815bc…abd3` | 5 頁、PDF 1.6、382.32×598.8 pt | Poppler 22.12.0 |
| 三張 JPEG | 地圖 625×625；盒面 644×960；盒背 664×961；均 RGB | Pillow 11 metadata |

### 結論

- 狀態：`強推論`
- 行為／身分：目前封裝含 DOS MZ 程式與配套資料，是 DOS 遊戲集合；CRC 完整。
- remake 對應：將此 ZIP SHA-256 作為目前唯一研究基準，runtime 不直接讀它。
- 剩餘不確定：1998 timestamps 的成因、是否含玩家存檔／修改資料、發行 revision、
  compiler／linker／overlay、圖像與音訊格式均未知。

### 重生方式

完整命令由 `tools/inventory_original.py` 實作，於唯讀 Docker mount 執行，輸出
`docs/re/original-inventory.json`。任何輸入雜湊、項目數、CRC 或逐項 SHA-256 改變都應
視為另一份 build，不得沿用後續位址與語意。

### 舊引用更新

- [x] `CONTEXT.md`
- [x] `WORKLIST.md`
- [x] `REMAKE-PLAN.md`
- [ ] 格式規格（尚未建立）
- [ ] 程式與測試（尚未開始）

## RE-002：六個 DOS executable 的 IDA 第一層指紋

### 目標

- 輸入：RE-001 manifest 中六個固定 SHA-256 executable
- 工具：IDA Pro 9.4，image `ida-pro-9.4-idapython:locked-v1`，image ID
  `sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`
- 位址空間：IDA linear effective address；各檔 load base `0x10000`
- 輸出：`docs/re/ida-fingerprints/*.json`，schema `kol-ida-fingerprint-v1`

### 問題

IDA loader 如何分類六個 executable？哪些 segment／入口／library signature 可作後續
工具鏈分流？是否已有足夠證據宣稱 compiler 或 overlay？

### 證據

| 檔案／SHA-256 短碼 | IDA range／entry CS:IP | Segment／函式 | FLIRT log | 結論等級 |
|---|---|---|---|---|
| `KOL.EXE` `4f798b…d7b71` | `0x10000..0x132e8`；`1000:0248` | 4／2 | Quick C 1/2.01＋MSC 5.1 runtime candidate | 強推論 |
| `KOLBOOK.EXE` `2cada0…5832` | `0x10000..0x270e0`；`1000:6d5e` | 3／2 | 同上 | 強推論 |
| `KOLTOWN.EXE` `c2442c…b221` | `0x10000..0x30cc0`；`2f51:0012` | 3／0 | 無 | 未知 |
| `MAKEMODE.EXE` `3051cc…6c56` | `0x10000..0x226d8`；`1000:1964` | 3／2 | 同上 | 強推論 |
| `TANDYPIN.EXE` `036c39…cce3` | `0x10000..0x10070`；`1000:0000` | 2／0 | 無 | 未知 |
| `TITLE.EXE` `8a2879…fad6` | `0x10000..0x17248`；`1000:1cde` | 3／2 | 同上 | 強推論 |

六個 MZ header 的 `overlay_number` 都是 0。依 MZ page／last-page 欄位計算，`KOLBOOK`、
`KOLTOWN`、`MAKEMODE` 分別有 55、81、63 bytes 位於宣告映像尾端之後；這些小尾端目前
只列為 appended payload／padding 候選，不能稱為 overlay。其餘三檔差額為 0。

### 結論

- 狀態：部分 `強推論`、部分 `未知`。
- 四個 executable 可確認 IDA 找到同一 Microsoft runtime signature 候選與 `_main`
  入口；不能由 FLIRT 年代範圍鎖定精確 compiler／linker。
- `KOLTOWN` 的 0 relocation、非零初始 CS、巨大 data segment 與零函式結果需要以 IDA
  entry bytes／控制流另查，可能解釋為特殊 layout、packing 或自解碼，但目前都只是候選。
- `TANDYPIN` 僅 624 bytes，功能與工具鏈未知，不從名稱推定 Tandy pin driver 語意。

### KOLTOWN entry 窄追查

固定原始位置為 IDA linear EA `0x2f522`。含原始 bytes 與 32 條有限線性解碼的
`ida-fingerprints/KOLTOWN-entry.json` 只留本機並由 Git 精確排除；其中 stub 直接讀取：

- `[CS:0004] = 0x0000`：寫入 `ES+0x10`；
- `[CS:0006] = 0x131f`：`rep movsb` 長度；
- `[CS:000c] = 0x1f92`：加到 load segment 形成目的 segment；
- `[CS:000e] = 0x0001`：後續保留於 `DX`。

已證實的控制流是：`DS=CS`，以 `STD; REP MOVSB` 反向複製 `0x131f` bytes，然後 push
目的 segment 與 `0x0038`，用 `RETF` 轉移。這證實 entry 是搬移／轉接 stub；它是否同時
解壓、packer 的名稱及第二階段真正入口仍未知。IDA loader 的「Possibly packed file」與
0 relocation 可支持 packed／self-relocating 的強推論，不能單獨證明 packer 家族。

MZ 宣告檔長為 133,679 bytes，實際 133,760 bytes；尾端 81 bytes 全為 `0x1A`。因此
RE-002 先前的 appended payload／overlay 候選已被推翻，訂正為 DOS EOF padding（已證實）。

### 重生與工具勘誤

第一次最小探針錯用不存在的 `ida_kernwin.qexit`；雖產生 JSON，仍判定失敗且不採用。
改用 `idc.qexit` 後在全新一次性 DB 重跑，確認 schema、IDA 9.4、輸入 SHA-256、非空 JSON
與 UID/GID。正式證據只保存 JSON；暫存 `.i64`／log 不提交，避免攜入授權環境資訊。
entry probe 又經兩次失敗即關閉訂正：先修 `generate_disasm_line` 所屬 module，再修 MZ
last-page／pages 的 offset；最終輸出以 `trailing_bytes=81`、尾端全 `0x1A`、128 entry bytes
及輸入 SHA-256 作硬閘。

## RE-003：KOLTOWN EXEPACK 與 unpacked derivative

### 問題

Stage 2 是否為已知 packer？能否建立具 relocation、真正入口及 IDA 函式圖的研究映像？

### 固定證據

- 原始輸入：`KOLTOWN.EXE` SHA-256 `c2442c…b221`。
- IDA：9.4 locked-v1；packed 位址空間。
- Stage 2：packed EA `0x2f548..0x2f627`，104 instructions、12 direct xrefs。
- Marker：stub base `+0x117` 的 `Packed file is corrupt`。
- 解壓命令：`0xB0/0xB1` fill、`0xB2/0xB3` copy、low bit 是 last-block flag。
- Relocation：stub offset `0x12d` 起的 16-section table；逐項加 load segment。
- Final transfer：`CS:IP=1b8f:001c`，`SS:SP=2b48:07d0`。

上述 marker、資料格式、控制流及成功 consumer 共同證實 Microsoft EXEPACK；精確工具版本
未知。安全摘要見 `ida-fingerprints/KOLTOWN-stage2-summary.json`；含完整 bytes／disassembly
的 private JSON 只留本機 ignored。

### 解包與 IDA consumer

沿用既有 CC0 `unexepack.c`，來源 SHA-256 `0e1da11a…b03c`，在 GCC 13 容器原樣編譯。
輸出 derivative SHA-256 `01f6f9a9…d1d2`、138,528 bytes、2,281 relocations。IDA 9.4
重載入口為 unpacked EA `0x2b90c`；手動建立入口函式後辨識 337 functions，入口函式
`0x2b90c..0x2b9af` 有 63 instructions／7 call sites。

### 結論與 remake mapping

- Packer：Microsoft EXEPACK，`已證實`；版本 `未知`。
- Unpacked derivative：可作後續 IDA xref／函式邊界 oracle，仍待 SPEC-001 的 runtime
  memory 三點核對與 garbage-byte consumer gate，故 spec 保持 DRAFT。
- Remake 不移植 EXEPACK／DOS relocation／compiler startup；只使用它們抵達產品程式。
- 未達 READY 前不建立玩法 implementation。

### 勘誤

先前把 `KOLTOWN` 81-byte 尾端列為 appended payload／overlay 候選；已由全 `0x1A` 原始
bytes 推翻，改為 DOS EOF padding。先前「packer 未知」也由本節證據升級為已證實 EXEPACK。

## RE-004：DOSBox-X 執行期 oracle 啟動切片

### 固定工具與問題

- DOSBox-X 2026.06.02，既有 image `rich2-dosbox-dbg:latest`，原生唯讀 AI bridge。
- 原始 ZIP 仍以唯讀方式掛載；解壓副本只存在 `/tmp`，不加入版控。
- 問題：能否在 packed 程式自行解包後，擷取載入映像並與 CC0 derivative 比較？

### 目前結果

`-break-start` 實際停在 BIOS `F000:FFF0`，不是 MZ entry；已停止把該狀態誤當程式入口。
直接執行 `KOLTOWN.EXE` 與暫定 `KOLTOWN.EXE 0` 都會在三秒觀察窗前退出，傳統記憶體內
找不到檔案中固定的 `Packed file is corrupt` marker。因此本輪沒有 runtime bytes，不能
關閉 SPEC-001 的三點核對 gate。

這項結果只證實 `KOLTOWN` 不是可用上述命令獨立維持的 oracle；「必須由主程式帶哪些
狀態／參數呼叫」仍為 `未知`。下一步由 IDA 9.4 追查啟動參數 consumer 或主程式 exec
contract，不以猜參數的方式繼續。
