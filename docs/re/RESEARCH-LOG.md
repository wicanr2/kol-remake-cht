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

`-break-start` 實際停在 BIOS `F000:FFF0`，不是 MZ entry。IDA unpacked EA `0x103b3` 讀
第二個 C 參數 `argv`，`0x103b6` 取 `argv[1]`，`0x103b9..0x103bc` 以首字元減 ASCII `A`
形成 0..15 索引；因此最小合法命令列為 `KOLTOWN.EXE A`。startup 原始函式與全域摘要見
`ida-fingerprints/KOLTOWN-startup-summary.json`；完整指令只留 private JSON。

第一次啟動探針有三項已保留的勘誤：

1. 0.1／1／2／3 秒採樣時 DOS 尚未完成 BIOS／COMMAND.COM 啟動，不能解釋成 KOLTOWN 退出。
2. `cd Knights-of-legend` 依賴 DOS 8.3 長檔名映射，實際未進入遊戲目錄；改為直接把該
   目錄掛為 `C:`。
3. derivative MZ header 欄位 `0x0240` 的單位是 paragraph，故 header 是 `0x2400`
   bytes，不是 `0x240`；錯誤 dump 向前偏 `0x21c` 段，已作廢並重生。

修正後，第一輪持續執行採樣由 Microsoft runtime marker 推得 load segment `0824`。
第二次乾淨啟動預先設定 `23B3:001C`（`0824+1B8F:001C`）斷點，精確命中 EXEPACK final
transfer 後、C startup 前。runtime marker 再次獨立反推 `0824`，129,312-byte load image
SHA-256 為 `df79480d…67f6`。

比較器依 derivative 的 2,281-entry MZ relocation table，逐 word 驗證 runtime 值等於
derivative 值加 `0824`，全部通過；扣除 relocation 後整份 load image **0 byte 差異**。
另有五個分散 32-byte 錨點（unpacked IDA EA `0x10000`、`0x18000`、`0x20000`、
`0x28000`、`0x2ded8`）全部吻合，超過 SPEC-001 的三點最低要求。

### 結論

- KOLTOWN 的合法最小啟動參數 `A`：`已證實`（IDA argv consumer＋DOS 正常執行）。
- CC0 derivative 對此固定 build 的 canonical startup load image：`已證實`逐 byte 一致。
- 上游對一般 EXEPACK garbage bytes 的警告仍保留；本 build 在 startup 前實測零差異，故
  不存在會進入 code、relocation 或玩家資料 consumer 的實際差異。
- SPEC-001 可由 DRAFT 升 READY；只授權 derivative 作後續研究 oracle。

## RE-005：圖片容器索引與 LZW 解碼核心

### 問題與固定證據

如何由 `*PICS*` 候選選取單張圖？壓縮資料如何解碼？目前證據是否足以產生正式 PNG？

- 輸入 ZIP：SHA-256 `2b455e3c…9f46`；KOLTOWN derivative：`01f6f9…d1d2`。
- IDA Pro 9.4 locked-v1，unpacked linear EA；完整指令在本機 `*-private.json`。
- `sub_1A429`／`sub_1A4FE` 證實 `20 + index*2` 的 word offset table、byte 0 count、
  相鄰 offset 差與 offset 10 的最末 payload 長度。
- `sub_1A04E`／`sub_1A26B` 顯示 9-bit 起始、LSB-first、`0x100` 碼寬增加標記、
  `0x101` 起字典與 LZW 特例；語意為 `強推論`。
- 當時清冊有 42 個候選；後續 RE-007 補入 AWD PCGA／HERC 變體後為 44 個。30 個符合
  容器契約，共 960 個 payload。全部 offset 嚴格遞增且 payload 有界；當時
  `AWD1PICS.0` 的 520-byte 尾區尚未知，後由 RE-007 訂正分類。
- 一次性探針對 960 個 payload 完成解碼且無無效字典碼；只證明機械自洽，不證明像素。

### 結論與停止線

容器 index 契約為 `已證實`；LZW 變體為 `強推論`。尺寸、顯示 suffix、位平面、palette、
透明色、畫面方向及尾端 520 bytes 在 RE-005 當時都未關閉，因此尚不能產生正式 PNG，
也不能接 Go runtime。格式細節、固定顯示契約與後續訂正集中於 `image-formats.md`；當時
的下一窄任務是沿一個 consumer 完成來源到 blit 的垂直鏈並做原版視覺交叉驗證。

## RE-006：COMBPICS EGA 垂直鏈與外部視覺 Oracle

### 原版資料流

- KOL loaded-MZ EA `0x100d1..0x10219` 以 BIOS EGA query 選 mode `0x0d`，再呼叫
  `INT 10h` 設定模式；KOLTOWN EA `0x10032..0x10053` 由 BDA `0040:0049` 讀回模式。
- DGROUP offset `0x0670` 是 `pics\0pcga\0herc\0`；mode 4 選 PCGA、mode 7 選 HERC，
  mode 0Dh 使用 PICS。
- `sub_288D4` 載入 `COMBPICS` 後設定 `x=[1,38)` byte-columns、`y=[8,152)`，經
  `sub_1C907` → `sub_1A2BB` 做二階轉換，再以 `sub_19E9D` blit。
- `sub_1A2BB` 證實 big-endian 16-bit stream 分界、high-nibble-first 4-bit 色彩、repeat／
  literal RLE 與四 EGA bitplanes。前五張均展開為 `296×144=42,624` pixels。

### Palette 停止線與視覺對照

KOL／TITLE／KOLTOWN 的 IDA-defined code `INT 10h`／port I/O 清冊未見 palette 寫入；
TITLE 只重設模式與操作 speaker／timer／EGA plane，`MAKEMODE` 沒有 video I/O。依硬體
規格停止線，RGB 對應採成熟 DOSBox-X commit `63a4357` 的 mode 0Dh 與 `ega_palette`
契約（來源 SHA-256 `42044a6d…4060`），換算 6-bit 值為
`00/55/aa/ff`。此色彩是 `hardware-spec approximation`，不宣稱類比螢幕逐波形一致。

可丟棄探針產生 `COMBPICS[0]`、`[1]` 296×144 PNG，分別與魔旅士文章的同場景 JPEG
匹配。固定 crop／BICUBIC 的 RGB MAE 正配為 `20.12`、`19.17`，交叉錯配為 `62.00`、
`62.88`。原圖只留 ignored 工作區；安全 URL、SHA-256 與矩陣見
`combpics-visual-oracle.json`。

### 結論

COMBPICS 前五項的 LZW、二階 RLE、4-bit 索引、296×144 幾何、方向、EGA palette mapping
與同場景身分均為 `已證實`；JPEG 不能證明 byte-exact RGB。此結論只關閉單一 PICS／
mode 0Dh 分支，不能泛化到其他 archive、PCGA、HERC、透明圖或無副檔名 runtime 檔。
SPEC-002 當時仍為 DRAFT；RE-007 接續追查 `AWD1PICS.0` 尾端與三個無副檔名檔案。

## RE-007：AWD 工作區快照與角色圖片記錄表

### AWD 固定快照邊界

- 輸入 KOLBOOK SHA-256 `2cada028…e5832`；IDA Pro 9.4 loaded-MZ linear EA。
- KOLBOOK `_main` 以 `awd%dpics.0`／`awd%dpcga.0`／`awd%dherc.0` 選檔，分別從 far
  offset `0x0100`／`0x1000`／`0x2000` 呼叫 `sub_137C0`，長度一律為 `0x0fa0`。
- `sub_137C0`（`0x137c0..0x13843`）使用 `__dos_creat`、指定長度的 `__dos_write`、
  短寫檢查與 `__dos_close`；`sub_1374E`（`0x1374e..0x137c0`）則以最大 `0xffff`
  bytes 做整檔讀取。原始名稱、位址、指令與呼叫端保存在 ignored private JSON。
- KOLTOWN `sub_168F8` 建立 `awd%d%s` 檔名後呼叫索引 loader `sub_1A4FE`，故 PICS 的
  玩家顯示路徑只消費索引 payload，止於 byte 3480。

結論：三個 AWD 變體都是固定 4000-byte 編輯器工作區快照，`已證實`。PICS 在索引資料後
的 520 bytes 已排除於 KOLTOWN 圖片流，但內容仍未解析；分類為 opaque workspace tail，
不可改稱 padding，也不可輸出成 PNG。

### 固定角色圖片記錄表

- `CHARPICS`、`SHLDPICS`、`STEEDPIC` 實檔皆為 `0x2f80` bytes。
- KOLBOOK `sub_135DB` 整檔讀入 CHARPICS／SHLDPICS；`sub_11820`、`sub_1293E` 以
  `sub_137C0` 固定寫回 `0x2f80` bytes。
- KOLTOWN `sub_14DB1` 以 CHARDATA 的 `0x300`-byte 槽位計算圖片槽位 `0x2f8` bytes，
  每次複製 `0x17c` words，再依 DGROUP offsets `0x1744..0x176c` 寫 STEEDPIC、讀寫
  CHARPICS、讀寫 SHLDPICS，每次寫檔長度均為 `0x2f80`。
- `steedpics` 是另一個複數檔名；`sub_14779` 以最大 `0x7530` bytes 載入，不得和單數
  `STEEDPIC` 的固定記錄表混稱同一檔案。

結論：三個單數／無副檔名檔是 16 筆、每筆 760 bytes 的可變角色圖片記錄表，`已證實`；
記錄內像素寬高、位平面及透明規則仍為 `未知`。SPEC-002 的兩項 READY gate 已關閉，
但完整命名、manifest schema、失敗政策與權利審查未完成，故仍維持 DRAFT。

### 失敗紀錄與證據邊界

第一版 KOLBOOK IDA probe 使用 IDA 9.4 不存在的 `idc.get_segm_class`，未產生可採用輸出；
改用 `ida_segment.get_segm_class` 後在全新資料庫重跑。之後再明確匯出兩個底層 helper，
避免僅由呼叫端常數推測其 I/O 語意。安全摘要不含原版 bytes 或 disassembly。

## RE-008：TWN 邏輯分卷與全域圖號 24

### 舊結論勘誤

舊 v1 清冊把 `TWN*PICS.1/.2` 當成不符合 archive 形狀的獨立檔，因此只解碼 `.0` 第一卷，
報告 30 archive／960 payload。這個數字只覆蓋第一卷，不是完整邏輯 archive。舊摘要仍留在
Git 歷史；現行 v2 清冊改以 KOLTOWN consumer 的分卷契約重生。

### 分卷 reader 證據

- 固定 KOLTOWN derivative SHA-256 `01f6f9…d1d2`，IDA Pro 9.4，unpacked loaded-MZ
  linear EA；完整指令留在 `KOLTOWN-image-pipeline-private.json`。
- `sub_1A4FE` 在 `0x1a71c..0x1a770` 從 `.0` header 讀全域 offset 與最多五個累計門檻；
  `0x1a770..0x1a783` 以門檻選 `.0` 至 `.4`；`0x1a786..0x1a7c3` 對卷界最後一項改讀
  header offsets `10+volume*2` 的 last-length；其餘以相鄰 offset 差求長度。
- 固定 ZIP 中 TWN0..6 各有 `.0/.1`，TWN7 有 `.0/.1/.2`。第一卷 payload 從 byte 200
  起，續卷從 byte 0 起；每卷最後一項加 last-length 都精確等於實體檔案大小。
- 實際邏輯圖數為 25、34、30、26、26、30、26、46，合計比舊清冊多 62 張；30 個邏輯
  archive 總數為 1,022。v2 LZW 探針全部解碼，無越界或無效字典碼。

結論：`.1/.2` 是同一 TWN archive 的 continuation volume，分卷選擇、全域圖號、offset
基準與末項長度為 `已證實`。LZW 語意仍依既有證據標為 `強推論`，不能從 1,022 次機械
成功外推所有像素形狀。

### 全域圖號 24 垂直鏈

- `sub_2AAA8` `0x2aab1..0x2aab5` 傳常數 24 給 `sub_11302`。
- `sub_11302` 以 `twn%d%s` 選當前 town／顯示 suffix；圖號 24 在八組 archive 全部由
  `.1` 提供。town 0 設矩形 byte-column `[13,37)`、scanline `[58,154)`；其餘 town 設
  `[1,12)`、`[64,128)`。
- `sub_10F9C` 完整串接 `sub_1A4FE` → `sub_1C907`／`sub_1A2BB` → `sub_19E9D`。
  因 mode 0Dh 每 byte-column 為 8 pixels，兩種輸出分別為 192×96 與 88×64。
- 可丟棄探針對八張圖的像素數全部精確相符，palette index 均在 0..15；兩次 PNG 逐 byte
  相同。接觸表目視為 town 0 道路／提示畫面及七張人物肖像，無鏡像、旋轉、空白或明顯
  scanline 錯位。PNG 與接觸表只留 ignored 工作區，安全摘要 SHA-256 `642a68c6…ed54`。

結論：TWN 圖號 24 的分卷、LZW、二階 RLE、4-bit indices、幾何、方向與 blit 為
`已證實`；EGA RGB mapping 維持 `hardware-spec approximation`。圖片的劇情／人物語意
未由文字 consumer 證實，因此只以「全域圖號 24」稱呼，不作推測性命名。

## RE-009：MAKEMODE 的 CGA／Hercules 素材生成鏈

### 問題與固定證據

固定 ZIP 為何只有一般 `*PICS*`，KOLTOWN 卻有 `pics/pcga/herc` suffix consumer？缺少的
PCGA／HERC 是遺失素材、另一版發行物，還是可由工具生成？

- 輸入 `MAKEMODE.EXE`，SHA-256 `3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56`。
- IDA Pro 9.4 `locked-v1`，image ID `6f6d59af…e2780`，loaded-MZ linear EA；DGROUP EA
  `0x13210`。完整指令只留 ignored private JSON。
- `_main` 的模式提示 consumer 明確提供 CGA／Hercules 轉換；mode 1 suffix 為 `pcga`，
  mode 2 suffix 為 `herc`。長篇原版提示在版控摘要只留長度與 SHA-256，不提交全文。
- `_main` `0x1019d..0x103c4` 共 28 次呼叫 `sub_10464`。安全 probe 機械匯出每次的
  原始 call EA、base、image count 與 transform case；28 個 count 全部與 v2 邏輯清冊
  相符。表中不含 AWD 工作區與動態角色 CHARPICS。

### 轉換資料流

- `sub_10464` `0x10491..0x104e5` 以 `%spics.0` 讀來源 header；`0x10544..0x10565`
  依全域圖號選來源分卷並整檔讀取。
- `sub_114F4`（`0x114f4..0x116e4`）是 9-bit 起始、`0x100` 增寬標記的 LZW 解碼器，
  與 KOLTOWN decoder 結構獨立吻合。
- `sub_10464` 依 0..22 transform case 建立矩形；`0x10fb5..0x10fe0` 計算面積並呼叫
  `sub_11734`。後者將二階 RLE／nibble stream 展開為四個 EGA bitplanes。
- `sub_118A8`（`0x118a8..0x11964`）逐八像素合成 4-bit EGA index，套 16-entry mapping，
  再打包為兩個 2-bit bytes；`sub_11242` 接著重壓縮。
- `sub_10464` 以 `%s%s.%d`／`%s%s.0` 重建 header 與分卷，`sub_1142C` 透過
  `__dos_creat`／`__dos_write`／短寫檢查寫出 `basepcga.N` 或 `baseherc.N`。

### 結論與停止線

- PCGA／HERC 的一般圖片不是固定 ZIP 中遺失的靜態輸入，而是官方 MAKEMODE 由 EGA
  PICS 生成的衍生 archive：`已證實`。
- 28 組來源基名、圖數與輸出命名：`已證實`；全部圖數與獨立 v2 清冊一致。
- 解碼 → transform case → 4-plane 至 2-bit → 重壓縮的整體靜態鏈：`強推論`。尚未用
  MAKEMODE 動態輸出做 byte-level／runtime round-trip，Hercules 的單色／抖色規則也未關閉。
- 是否讓 CGA／Hercules 成為首版必要 theme 是產品範圍，不由 RE 代決。共同決策前停止
  深挖 23 個 transform case；仍可繼續不依賴該答案的 EGA PICS consumer 研究。

## RE-010：MONPICS 彩圖／遮罩垂直鏈

### IDA 尺寸與 consumer

- 輸入 `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4
  `locked-v1`，unpacked loaded-MZ linear EA。正式 `.i64` 唯讀，探針只在 `/tmp` 副本執行。
- DGROUP offset `0x1370`、IDA EA `0x2f240` 至下一張表前共有 32 組兩-byte 記錄。
  `sub_22EFC` 以 `2 * word_3B262` 取出 8-pixel 寬度單位與 scanline 高度，建立置中矩形；
  `0x22fb9` 以同一偶數索引呼叫 `sub_10F9C` 載入 `mon%s`。
- `sub_2A975` 使用同一尺寸表，並在 `0x2aa0f` 令 `sub_16FA6` 消費
  `[2 * monster_index, 2 * monster_index + 1]`。這解釋實檔 64 個 payload 與 32 組尺寸。

### 像素驗證與勘誤

- 第一版探針錯把 64 個 payload 當 64 個怪物，尺寸表後半出現零值而失敗；回查下一張表
  `0x13b0` 與兩條索引運算後，訂正為 32 個怪物、每個兩筆資料。
- 第二版又把 COMBPICS 的二階 RLE 套到每一筆；奇數 payload 首二 bytes 不構成合法分界，
  探針在 index 1 失敗。逐筆長度與 `sub_16FA6` 範圍證明偶數是彩圖、奇數是遮罩。
- 32 個偶數 payload 全部精確展開為 `width_bytes * 8 * height` 個 4-bit index。32 個奇數
  payload 均為 `width_bytes * height + 1` bytes；前段以 MSB-first 展開為 1-bit 遮罩，
  多出的單一 byte 不進入可見矩形。
- 兩筆合成的 32 張 RGBA PNG 與接觸表均只留 ignored `workplace/`。人工抽查顯示完整、
  可辨識的怪物，沒有鏡像、反相、位元錯位或遮罩偏移。

### 結論與限制

32 組尺寸、偶數 4-bit EGA 彩圖、奇數 MSB-first 1-bit 遮罩、配對、方向與透明邊界為
`已證實`。奇數 LZW 輸出末端多出的單一 byte 不影響 consumer 尺寸，但成因未由終止碼
證實，維持 `強推論：壓縮流終止／padding artifact`。安全摘要只保存尺寸、雜湊與統計；
完整指令與 table bytes 在 ignored private JSON。SPEC-002 仍因其餘預定格式與 theme
範圍未關閉而維持 DRAFT，不接 production extractor 或 Go runtime。

## RE-011：PERPICS 六組人物彩圖／遮罩垂直鏈

### IDA consumer 與索引

- 輸入 `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4 `locked-v1`，
  unpacked loaded-MZ linear EA。探針在全新 `/tmp` DB 執行，完整指令與 bytes 只存 private。
- 第一輪自動化因未等待 IDA auto-analysis 而只產生 `.i64`、沒有 JSON，不列成功證據；
  加入 `ida_auto.auto_wait()` 後在全新 DB 重跑通過。後續擴充探針也另用全新 DB 驗證。
- `sub_18F3A:0x18fdd` 直接消費 DGROUP `0x17d8` 的 `per%s`；`0x18ff3` 以
  `[portrait_index, portrait_index+6]` 呼叫 `sub_16FA6`。`sub_16FA6:0x16fe3/0x1706b`
  兩次呼叫 `sub_1A4FE`，`0x17087` 再呼叫 `sub_1BD17` 做遮罩 blit。
- 舊 fingerprint 的 near-offset occurrence 本身不足以證實 consumer；本輪參數資料流才將
  `per%s` 歸屬升為 `已證實`。另追查的 `sub_1CBA9` 已排除為媒體／磁碟提示流程。

### 幾何、像素與目視

- DGROUP `0x0168` 六筆三-byte 記錄配合固定 `y1=154`，給出 72×103、64×111、72×64、
  112×148、72×103、56×110 六種矩形；原始 operand、table bytes 與位址保留 private JSON。
- `PERPICS.0` 前六筆經 LZW 與二階 RLE 全部精確得到矩形所需 4-bit EGA pixels；後六筆
  全部等於 `width_bytes*height+1`，以前段 MSB-first bits 作遮罩後與彩圖逐像素吻合。
- 六張 RGBA PNG 與接觸表只留 ignored `workplace/`；人工抽查為完整直立人物，沒有鏡像、
  反相、plane 錯排、遮罩位移或裁切。安全摘要保存尺寸、雜湊與統計，不含原版 bytes。

### 結論與限制

六組尺寸、`[i,i+6]` 彩圖／遮罩配對、4-bit EGA、MSB-first 遮罩、方向及透明邊界為
`已證實`。每個遮罩末端多出的單一零 byte 不進入矩形，成因維持
`強推論：壓縮流終止／padding artifact`。SPEC-002 仍因其他 archive 與 theme 範圍未
關閉而保持 DRAFT；本結論不授權 production extractor 或 Go runtime。
