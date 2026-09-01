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

## RE-012：KOLMPICS 圖號 28 的 EGA AND-mask

### archive 到 consumer

- 固定輸入與工具：`KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`、IDA Pro 9.4
  `locked-v1`、unpacked loaded-MZ linear EA；每次探針修改都在全新 `/tmp` DB 重跑。
- `sub_10000:0x1037a` 消費 DGROUP `0x165a` 的 `kolm%s.0`，`0x1038e` 以 `sub_19D98`
  把 archive 讀入原始全域 `word_39820`。完整運算元與 bytes 只留 private JSON。
- `sub_14047:0x14150/0x14154` 傳入 `word_39820`，`0x14158` 放入常數 `0x1c`，並於
  `0x1415c` 呼叫 `sub_1A429`；因此圖號 28 與 buffer consumer 為 `已證實`，不是 near-offset
  猜測。來源矩形 `[0,7)×[0,57)` 對應 56×57 pixels。

### 遮罩語意與視覺驗證

- 圖號 28 LZW 輸出 400 bytes；前 399 bytes 恰為 `7*57`，MSB-first 展開得到 3,192 bits，
  末端一個零 byte 不進入 consumer 矩形。
- 初版透明預覽沿用其他 archive 的 1=opaque，幾乎全空；未採為證據。回查
  `sub_1BD17:0x1bf4c..0x1bf53` 證實 EGA 式為 `(destination AND mask) OR color`，訂正為
  bit 1 保留背景／透明、bit 0 清除後覆寫／不透明。
- 修正後 bit-map 目視為右上切角、圓弧下緣的連續遮罩，沒有鏡像、scanline 錯位或亂碼。
  PNG 只留 ignored `workplace/`，安全摘要保存尺寸、bit count 與 SHA-256。

### 結論與限制

KOLMPICS 圖號 28 的來源、索引、56×57 幾何、MSB-first 位序、EGA AND-mask 語意與方向為
`已證實`。末端零 byte 的成因仍是 `強推論：壓縮流終止／padding artifact`。圖號 28 是
合成遮罩，不是可獨立當彩圖輸出的素材；其餘 82 個 KOLMPICS payload 保持未知，SPEC-002
仍為 DRAFT。

## RE-013：MISCPICS 完整 archive 垂直鏈

### 唯一 consumer 與路由

- 固定輸入 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4
  `locked-v1`，unpacked loaded-MZ linear EA。探索與專用探針均使用全新 `/tmp` DB。
- `sub_10000:0x1060d/0x10621` 以 `misc%s.0` 載入 `word_2FBB8`。該全域全部 data refs
  僅 `0x105ec` 啟動配置、`0x10e62` cleanup、`0x111c0` consumer，排除旁路直接讀取。
- `sub_1103F` 證實 request code 34／35／71／72／73 依序映到 image 0／1／3／4／5；
  `0x111c7` 的 `mapped-6` 後索引集合不含 2。`sub_11262:0x11288` 解碼，`0x112f4`
  直接 blit。原始函式名、位址、operand 與 bytes 保留 private JSON。

### 幾何、像素與 placeholder

- IDA 矩形依序固定為 192×96、296×32、56×70、56×70、56×70。五個 payload 經 LZW
  與二階 RLE 後分別得到 18,432、9,472 與三組 3,920 個 4-bit indices，全部精確吻合。
- 接觸表目視為一張場景、一張橫幅及三張人物／標記圖；方向、scanline、plane 與裁切正常。
  consumer 是 direct blit，沒有 mask 證據，因此黑色索引保留為色彩，不轉 alpha。
- image 2 是唯一一-byte payload；控制流不會選到，且 buffer 沒有其他 consumer。它分類為
  `已證實：未引用非圖片 placeholder`，正式 manifest 應是 `excluded_non_image`、無 PNG。

### 結論

MISCPICS 六個 payload 已完整分類：五張已證實 EGA 彩圖與一筆已證實非圖片 placeholder。
來源、索引、尺寸、二階 RLE、4-bit palette bounds、方向及 direct-blit 透明政策均已關閉。
PNG 與接觸表只留 ignored `workplace/`；安全摘要只保存雜湊與統計。這不改變 SPEC-002
整體 DRAFT 狀態，也不授權 production importer。

## RE-014：COMBPICS 圖號 5／6 與全 archive 完成

### 舊結論範圍與新 consumer

- RE-006 只證實 `sub_288D4` 使用的前五張 296×144 圖，不能解釋 archive 的七筆；舊證據
  保留，不改寫成當時已完成。本輪另追 `sub_2B454`，補齊圖號 5、6。
- 固定輸入 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`、IDA Pro 9.4 `locked-v1`、
  unpacked loaded-MZ linear EA。完整函式與 bytes 只留 private JSON。
- `sub_288D4` 全部 call sites `0x2125c/0x2157b/0x215ac/0x2161f/0x216b5` 分別傳
  3／2／0／4／1；矩形固定 296×144。`sub_2B454` 全部 call sites `0x2abd2/0x2adcb`
  分別傳 6／5；矩形固定 88×64。兩組集合恰為 archive 0..6，沒有未歸屬 payload。

### 像素與視覺

- 七筆 LZW 後皆通過同一 big-endian 分界、high-nibble-first 二階 RLE；0..4 各得到
  42,624 pixels，5..6 各得到 5,632 pixels，所有 index 均在 0..15。
- 七張 ignored PNG 接觸表顯示五張戰鬥結果／場景與兩張肖像，方向、scanline、plane、
  裁切皆正常。圖號 0、1 仍引用 RE-006 的魔旅士文章同場景獨立交叉證據。

### 訂正後結論

COMBPICS 不再只是「前五項已證實」，而是全七筆的 consumer、尺寸、二階 RLE、4-bit EGA、
方向與 direct-blit 契約均為 `已證實`。這是對舊結論範圍的追加，不抹除 RE-006 當時限制；
SPEC-002 仍因其他 archive 與 theme 決策維持 DRAFT。

## RE-015：TITLE 的 SET0／SET1／SET2／FACE 全 archive

- 固定輸入 `TITLE.EXE` SHA-256 `8a2879d9…ad6`；IDA Pro 9.4 `locked-v1`、loaded-MZ
  linear EA。每次探針修改均建立全新 `/tmp` DB；完整指令、bytes 與 `.i64` 不入 Git。
- `_main` 在 `0x1021b/0x103e0/0x10558/0x10723` 將四個格式化檔名送入整檔 reader
  `sub_11058`。`sub_11666`／`sub_1170E` 證實索引 offset table，`sub_11133` 證實 mode
  0Dh 四 plane 矩形 blit；原始函式名與位址均保留在安全摘要。
- SET0／SET1／SET2 各由 `sub_10AF0` 完整使用 0..5，尺寸依序 88×70、104×84、
  160×125；FACE 的 18 個呼叫完整涵蓋 0..5，固定 112×80。
- 24 筆 LZW＋二階 RLE 像素數全部精確吻合，索引皆在 0..15。代表性 PNG 目視為山谷騎士
  動畫與正向臉部影格，方向、scanline、plane 與裁切正常；PNG 只留 ignored 工作區。
- 四個 archive 的來源、consumer、完整索引、尺寸、二階 RLE、4-bit EGA、方向與
  direct-blit 政策升為 `已證實`。SPEC-002 整體仍為 DRAFT，不授權 production importer。

## RE-016：TITLE 的 BACKPICS 全 14 張

- 固定 `TITLE.EXE` SHA-256 `8a2879d9…ad6`；IDA Pro 9.4 `locked-v1`、loaded-MZ linear
  EA。探索與最終探針每次修改均用全新 DB；完整指令與 `.i64` 只留 `/tmp`。
- `sub_10E9F:0x10e9f` 以 `back%s` 選檔；IDA 全部 30 個 code refs 只屬 `_main` 與
  `sub_10C86`，且每個呼叫均傳立即值圖號。集合恰為 0..13，沒有動態索引旁路。
- `sub_11728 → sub_11291 → sub_11A69 → sub_114F8` 串起 archive 選取、LZW 與二階
  RLE／四 EGA planes；`sub_11133` 提供 mode 0Dh blit 契約。
- 14 筆全部精確展開為 320×200、4-bit index 0..15。代表性 PNG 顯示山谷雷雨、閃電
  過場及製作名單，方向、scanline、plane 與裁切正常；PNG 僅留 ignored 工作區。
- BACKPICS 全 archive 的來源、consumer、完整索引、像素格式、尺寸、方向與無透明
  direct-blit 政策為 `已證實`。SPEC-002 仍維持 DRAFT，不建立 production importer。

## RE-017：KOLTOWN 的 ENDPICS 全七張

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4 `locked-v1`、
  unpacked loaded-MZ linear EA。字串探針最初未將 `end%s` 歸入函式，但原始 DGROUP offset
  唯一出現在 `0x2b821`；回查 IDA DB 證實它位於 `sub_2B80A`，不是未使用字串。
- `sub_2B80A` 的全部七個 caller 都來自 `sub_2B4A7`，立即值圖號依序為 0..6。
  探針同時硬閘門每個 call 前四筆矩形 assignment，避免人工抄錄成為唯一證據。
- 七張尺寸依序 168×122、192×111、296×93、176×97、168×123、192×102、184×98；
  LZW＋二階 RLE 像素數全數吻合，索引皆在 0..15。
- 七張 ignored PNG 目視為結局敘事人物／場景，方向、scanline、plane 與裁切正常。
  ENDPICS 全 archive 的 consumer、索引、幾何、像素格式與無透明 direct-blit 政策為
  `已證實`；SPEC-002 仍為 DRAFT。

## RE-018：KOLBOOK 的 BOOKPICS 全 14 張

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…5832`；IDA Pro 9.4 `locked-v1`、loaded-MZ
  linear EA。每次探針修改均使用全新 DB；完整指令與 `.i64` 僅留 `/tmp`。
- `_main:0x10292..0x102ac` 以 `book%s.0` 建檔名並整檔載入；另一個 raw word 命中
  `sub_172CC:0x172e2 mov al,5`，已排除為巧合，沒有冒充字串 consumer。
- `sub_1137B` 全部 caller 的常數集合為 0..13 與 100；100 在函式入口映為 9。
  圖號 0..3 是 320×200，4..12 是 112×31，13 是 248×22；幾何 operand 由探針硬閘門。
- 14 筆 LZW／二階 RLE 全部精確吻合矩形，索引在 0..15。PNG 目視為書本介面、章節標題
  與圖示列，方向、scanline、plane、裁切正常；PNG 僅留 ignored 工作區。
- BOOKPICS 全 archive 的來源、consumer、alias、索引、幾何、4-bit EGA 與無透明
  direct-blit 政策為 `已證實`。SPEC-002 整體仍為 DRAFT。

## RE-019：KOLTOWN WEATPICS 六組分層背景

### 問題與證據

固定 KOLTOWN derivative（SHA-256 `01f6f9a9…d1d2`）的 `weat%s.0` 是否可安全轉為
runtime-only PNG？IDA Pro 9.4 位址空間為 reconstructed unpacked MZ linear EA；原始名稱、
EA、bytes 與 disassembly 保留於 ignored private JSON，版控摘要為
`koltown-weatpics-layered-visual-probe.json`。

- `sub_2A0C6:0x2a0d5..0x2a11d` 將 group 乘 3，依序載入三個 payload。
- 全 18 筆分成六組 64×62；前兩層為 4-bit 色彩。group 0／1／2／5 第三層
  為 496-byte MSB-first 1-bit mask，group 3／4 為二階 RLE 後的四平面 mask。
- `sub_2A18E:0x2a1de..0x2a1e8` 已證實合成公式為 `(base AND mask) OR overlay`。
- 六張合成圖目視為晴空、雲景、風暴、降雨、降雪與夜景；故「野外天候／時段」
  為 `強推論`，不由縮寫檔名單獨宣稱。

### 勘誤、結論與 remake mapping

早期僅根據 caller 位置曾短暫歸為「戰鬥效果」；六張合成圖的直接視覺證據否定此說，
改為天候／時段背景候選。這項訂正保留於此，不重寫為從未誤判。

圖層角色、尺寸、像素值與合成公式為 `已證實`。mask PNG 是運算資料，不是 RGBA
alpha；runtime 可只讀 PNG，但必須依 manifest role 合成，不得把 mask 單獨顯示。
兩次全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
`11ef4d0e…bc35`、`638bce39…ae4a`；兩輪 18 張圖層 PNG、6 張 composite 與摘要也逐 byte 相同，
摘要 SHA-256 `7bb4e514…bed9`。SPEC-002 維持 DRAFT，未授權 production importer。

## RE-020：AWD1PICS 全 24 張勳章圖示

### 原版資料流

- 輸入：`KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4，reconstructed
  unpacked MZ linear EA。安全摘要為 `ida-fingerprints/KOLTOWN-awdpics-summary.json`。
- `sub_167D9:0x16839..0x168cb` 的 6×4 迴圈以 `4*outer+inner` 完整產生圖號 0..23；
  `0x168fc..0x168dd` 另遍歷五個角色槽。每個 packed 2-bit 值為 3 時才畫出對應圖示。
- `sub_168F8:0x16925..0x16940` 將矩形設為 3 byte-columns×21 scanlines，即 24×21；
  `0x16975..0x16978` 把圖號交給共用 loader `sub_10F9C`，再經 LZW、二階 RLE、
  四 EGA planes 與 direct blit。
- 固定 ZIP 只附 `AWD1PICS`；consumer 的角色槽 0..4 會建立 `AWD1..AWD5` 檔名。
  本切片只證實 bundled `AWD1PICS` 內容，不宣稱其餘四檔與它 byte-identical。

### 格式、視覺與結論

24 個 payload 均精確展開為 `24×21=504` 個 4-bit 色彩索引，無獨立 mask，故索引 0
也是色彩而不得自動轉透明。8×3 接觸表目視為正向的勳章、獎盃、盾牌、戒指與
紋章圖示，無鏡像、旋轉、plane 錯排或裁切。中文手冊第 41 頁與 Apple II 參考卡另證實
`Medal` 命令會顯示角色完成任務贏得的所有勳章。因此像素契約為 `已證實`，
「角色任務勳章圖示」為具獨立文件交叉支持的 `強推論`。

兩次全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
`5684459b…df8`、`445767bb…af2`；兩輪 24 張 PNG、接觸表與視覺摘要也逐 byte 相同，
摘要與接觸表 SHA-256 為 `8c9ed5fe…b5fe`、`b11f2e05…14d3`。SPEC-002 維持 DRAFT。

## RE-021：SHLDPICS.0 四組角色盾牌

### 名稱分流與 consumer

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4，位址空間為
  reconstructed unpacked MZ linear EA。
- 無副檔名 `SHLDPICS`（12,160 bytes）只走 `sub_13D9A`／`sub_14DB1` 的固定角色記錄表
  讀寫；它與有 20-byte index header 的 `SHLDPICS.0` 是兩種格式，不得互相代替。
- 圖片基名 `shld` 只有四字元，因 IDA `Strings()` 預設最短長度而曾漏列。它在
  DGROUP EA `0x2e4de`、offset `0x060e`；真正 operand consumer 是
  `sub_18F3A:0x1929e mov ax,60Eh`。raw word 另兩個命中只是遠呼叫位址 bytes 巧合。
- `sub_18F3A:0x192b6..0x192c5` 將盾牌型別 `i` 映為 `[2*i,2*i+1]`；
  `sub_16FA6` 分別載入彩圖與 mask，最後在 `0x17087` 呼叫 `sub_1BD17` 合成。

### 幾何、遮罩與視覺

DGROUP `0x2e30a` 的前四組 byte-column／scanline 配對為 `(5,29)`、`(7,45)`、
`(6,63)`、`(6,61)`，即 40×29、56×45、48×63、48×61。偶數 payload 的二階 RLE
均精確展開為對應 4-bit EGA pixels；奇數 payload 則是 MSB-first 1-bit AND mask，
有效長度後各多一 byte。

`sub_1BD17` 的 `(destination AND mask) OR color` 證實 mask bit 1 保留背景／透明，
bit 0 清除後覆寫／不透明；PNG alpha 必須是 `1-mask` 而不是直接複製 mask。
四張 RGBA PNG 目視為兩張圓盾與兩張徽章形盾，邊界連續，無鏡像、旋轉、plane 錯排或裁切。
「角色盾牌圖」與像素契約為 `已證實`。

兩次全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
`bfd58326…34fe`、`739147f4…6bcd`；兩輪四張 PNG、接觸表與摘要也逐 byte 相同，
路徑無關的最終摘要與接觸表 SHA-256 為 `13dd97ae…a6ef`、`079a90ad…0da9`。
早期摘要因寫入暫存 basename 而無法跨路徑重生，已作廢。SPEC-002 維持 DRAFT。

### 連帶勘誤：MONPICS／PERPICS PNG alpha

SHLDPICS 的同一 `sub_1BD17` consumer 證據顯示，MONPICS／PERPICS 早期探針的
`unpack_mask` 極性顫倒：它把 bit 1 當不透明，實際上 bit 1 是保留背景／透明。黑底
接觸表曾掩蓋不透明黑矩形，因此舊 PNG 與舊視覺摘要作廢。

兩探針已改為 `opaque=1-mask`，並用 8×8 棋盤背景重生 32 張怪物與六張人物。
輪廓正確露出棋盤，且兩輪全部 PNG、接觸表與摘要逐 byte 相同。MONPICS 摘要／接觸表
SHA-256 為 `7a5a5258…3b29`／`934f1eb3…f2b5`；PERPICS 為
`d6d87f46…9fc5`／`b1a90a2f…9ff8`。訂正不改變原始尺寸與配對結論，但取代所有舊 alpha 產物。

## RE-022：HANDPICS.0 十二組手部 sprite

### selector 與幾何

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4，位址空間為
  reconstructed unpacked MZ linear EA。
- `hand\0` 位於 DGROUP EA `0x2e4e3`、offset `0x0613`；`sub_18F3A:0x19493`與
  `0x19556` 以它建立 archive 檔名。
- `0x1941d..0x19422` 將 renderer class 乘 2，選擇同類別的兩個手部位置組；
  `0x194ab..0x194ba` 與 `0x1956e..0x1957d` 再各以 `2*group` 與 `2*group+1`
  載入彩圖、AND-mask。六個 renderer class 因此完整對應 archive 的 12 組／24 筆。
- DGROUP EA `0x2e492` 的 24-byte 表依序給出 12 組 byte-column／scanline。尺寸為
  class 0／1 各兩張 16×8，class 2 兩張 16×6，class 3 兩張 24×8，
  class 4 兩張 24×7，class 5 兩張 16×8。

### 像素、alpha 與結論

24 個 payload 全部依尺寸表精確展開；每組偶數索引為 4-bit EGA 彩圖，奇數索引為
MSB-first 1-bit AND-mask，且尾端多一 byte。`sub_16FA6`→`sub_1BD17` 證實
bit 1 保留背景／透明，PNG 使用 `alpha=1-mask`。

6×2 棋盤接觸表目視為六種角色外觀的成對手部 sprite，每列兩個相對位置的輪廓
正確露出棋盤，無旋轉、pair-order、plane 或裁切錯誤。但「第一／第二位置」不直接
宣稱為角色解剖學左／右手。格式、尺寸、配對、alpha 與手部身分為 `已證實`。

兩次全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
`8067fa1f…d88b`、`ddd7637a…ce18`；兩輪 12 張 PNG、接觸表與摘要也逐 byte 相同，
摘要／接觸表 SHA-256 為 `a180983f…1b03`、`6ed86cb0…75b2`。SPEC-002 維持 DRAFT。

## RE-023：WEAPPICS.0 全 28 組角色武器 sprite

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4、重建 unpacked MZ
  linear EA。第一版探針把 `0x17ea` 猜成 `weap`，實際 bytes 是 `%s%s\0`，因此失敗即關閉
  且未產生摘要；回查指令後修正為 DGROUP `0x0609` 的 `weap\0` token 加 active renderer
  suffix，保留這項勘誤而不把舊假設藏起來。
- `sub_18F3A:0x19303..0x19415` 由角色裝備欄位算出 family、variant 與偶數 selector；
  六值基表為 `[0,10,14,28,38,42]`，family 5 的 variant 2／3 另有交換。這些名稱只作
  導覽；遊戲內武器分類名稱與各 family 合法 variant 範圍未由本切片宣稱。
- 56 筆完整配成 `[selector,selector+1]` 的 28 組彩圖／AND-mask。DGROUP `0x0402`
  幾何表逐組固定 48–192×5–63 pixels；二階 RLE 彩圖、MSB-first mask 與一-byte 非像素尾端
  全部吻合。六個 renderer class 的 28 組相對位置表亦收進安全摘要。
- 28 張 RGBA PNG 依 `sub_1BD17` 使用 `alpha=1-mask`；4×7 棋盤接觸表顯示弓、刃器、斧、
  長柄武器、錘與連枷，透明輪廓連續，未見旋轉、pair-order、plane-order 或裁切錯誤。
- 兩個獨立全新 IDA DB 的 safe／private JSON 均逐 byte 相同，SHA-256 分別為
  `ea84a195…4831`、`fa282fe0…17f9`；兩輪 28 張 PNG、接觸表與視覺摘要也逐 byte 相同，
  摘要／接觸表 SHA-256 為 `f3f48278…f9fb`／`9af11e5a…7346`。
- 新增可重生 IDA／PNG 探針與安全摘要；SPEC-002 只新增此窄分支，仍為 DRAFT，未接
  production importer 或 Go／Ebitengine runtime。

## RE-024：HEADPICS／TORPICS／LEGSPICS 全 162 組角色分層 sprite

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4、重建 unpacked MZ
  linear EA。`sub_18F3A:0x1903a..0x19164` 以 component 2→1→0 處理 legs／tor／head。
- 初版探針由 archive 名稱猜 `0x17de` 是 `%spics`，實際 bytes 為 `%s%s\0`，所以
  fail-closed 且未產生摘要。最終證據固定 DGROUP `0x05fa` 的
  `head\0tor\0\0legs\0` token 表、`%s%s` 與 active renderer suffix 的組合資料流。
- 每部件由角色存在欄位、`variant=byte[0x10e+16*component]&0x0f` 與
  `group=9*renderer_class+variant` 選 `[2*group,2*group+1]`。三 archive 各 54 組／108 筆；
  0..53 table slots 對應六類×九格，但合法存檔 variant 是否只取 0..8 仍未證實。
- DGROUP `0x017a`／`0x02be` 提供 162 組寬高／位置。HEAD 尺寸 24–56×8–52、TOR
  56–88×28–71、LEGS 48–56×13–84；324 筆彩圖、mask 與尾 byte 全部吻合。
- 162 張 RGBA PNG 使用 `alpha=1-mask`。三份 6×9 棋盤接觸表呈現協調的髮型／頭盔、
  衣甲／軀幹與褲裙／腿甲；分離片段保留合理透明間隙，未見旋轉、pair-order、
  plane-order 或裁切錯誤。
- 兩個最終全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  `87151ef1…1b1b`、`20a82ec0…c932`；兩輪 165 個 PNG 與視覺摘要也逐 byte 相同。
  視覺摘要／HEAD／TOR／LEGS 接觸表 SHA-256 分別為 `d01b16e1…f3ba`、
  `3a4bd8a0…92a3`、`3dc7994e…58bc`、`fc0c8382…d53e`。
- 新增可重生 IDA／PNG 探針與安全摘要；SPEC-002 維持 DRAFT，未接 production importer。

## RE-025：KOLPICS.0 混合 archive 與 raw-planar 勘誤

- 固定輸入是 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`，IDA Pro 9.4、重建
  unpacked MZ linear EA。先檢查同名 `KOL.EXE`，已證實它只是顯示模式偵測／launcher，
  不是 KOLPICS consumer；此排除結論保留，避免日後因名稱相同重走錯路。
- startup、`sub_1103F` 與 `sub_1E3A8` 完整覆蓋 active indices
  `[0,1,2,4,5,8,12,13]`；`[3,6,7,9,10,11]` 解碼後都只有一 byte，且沒有 consumer，
  分類為 placeholder、不輸出 PNG。
- 索引 1／2／4／5／13 的 caller 矩形分別是 320×200、296×32、288×155、288×77、
  288×78，二階 RLE 像素數完全吻合，為 `已證實`。
- `sub_19B29:0x19b41..0x19ba9` 每次從 source stride 256 的前 128 bytes 取四個 32-byte
  planes，MSB-first 組成 256 個 4-bit chunky pixels。索引 8／12 的下游
  `0x1eafe..0x1eb06` 以 `y<<6+x` 存取，因此 64 欄為 `已證實`；160 列只由 10,241-byte
  有效長度與目視推得，仍是 `強推論`。索引 0 的 256×5 也只列為 `強推論`。
- 初版曾把每 128 bytes 當相鄰一列而輸出 256×80，接觸表呈明顯交錯雜訊；這項目視反證
  促使回查 in-place source stride 與下游 64 欄 consumer。舊幾何已作廢，不得當證據。
- 兩份全新 IDA DB 的安全／私有 JSON 逐 byte 一致，SHA-256 為 `af3fe195…9e2`／
  `25193b3d…58d`；兩輪 PNG 與摘要亦逐 byte 一致。安全摘要見
  `ida-fingerprints/KOLTOWN-kolpics-summary.json`，視覺摘要見
  `kolpics-ega-visual-probe.json`。SPEC-002 維持 DRAFT。

## RE-026：KOLMPICS.0 全 83 筆

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`、IDA Pro 9.4、重建 unpacked MZ
  linear EA。startup 以 `kolm%s.0` 預載 archive；`word_39820` 的直接 data refs 只進
  `sub_1103F`／`sub_11262` 與既有圖號 28 的 `sub_14047`，排除其他直接 consumer。
- router 證實 KOLMPICS 候選 request 集合為 0..28、31、36..70、74..82；29／30／
  32..35／71..73／100 轉交 KOLPICS 或 MISCPICS。archive 的 12 個一-byte payload
  `[1,29,30,32,33,34,35,70,71,72,73,82]` 分類為非圖片 placeholder。
- 圖號 0／2 依 caller 矩形與二階 RLE 精確得到 320×200／296×32；圖號 28 維持
  56×57 AND-mask、bit 1 透明。其餘 68 筆全部精確得到 480 pixels，並與兩個固定
  24×20 helper 契約吻合；`sub_27E20` 繼承 caller 幾何，因此相關個別圖號仍標強推論。
- 68 圖示接觸表呈現連續的手勢、武器、狀態符號、人物與方向箭頭，未見 RLE、plane-order、
  旋轉或裁切錯誤。這份目視支持共同 24×20 分組，但不取代動態 caller 證據。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為 `79fceaaf…e5ba`／
  `c39152e8…71d2`；兩輪 PNG／摘要亦逐 byte 相同。安全摘要見
  `ida-fingerprints/KOLTOWN-kolmpics-routes-summary.json`，視覺摘要見
  `kolmpics-all-ega-visual-probe.json`。SPEC-002 維持 DRAFT，未接 production importer。

## RE-027：固定角色圖片記錄的 40×38 EGA 像素契約

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832`；IDA Pro 9.4、loaded-MZ linear EA。
  KOLBOOK `sub_1293E` 先以 `rep movsw` 複製 `0x17c` words 的單筆記錄到暫存 segment，
  再將同一 segment 分別交給原尺寸與 4 倍顯示 wrapper。
- 原尺寸 EGA consumer `sub_15AEC:0x15aec..0x15b6e` 設 `CH=0x26` 列、每列
  `CL=0x14` bytes；每次先將 source byte 右移四次取 high nibble，再使用原 byte 取 low
  nibble。故單筆為 `38×20×2 = 1520` pixels，即 40×38、high-nibble-first 4-bit EGA，
  恰好耗盡 760 bytes。
- 4 倍顯示 consumer `sub_15B6E:0x15b6e..0x15c14` 使用同一 38×20 source loop，並將
  每個 nibble 擴展到四條 scanline，提供同程式內的獨立放大路徑交叉驗證。
- 兩份全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  `db71511e…b316`／`8bcbfc6e…815a`。三個來源各 16 筆皆機械解成 40×38 PNG；48 張
  ignored PNG 與接觸表已檢查，接觸表可見有效角色合成圖，也有空白或雜訊槽。
- 視覺摘要／接觸表 SHA-256 為 `7eede045…bb55`／`fd552865…37d1`。像素格式為
  `已證實`；逐槽角色、盾牌或坐騎語意及未使用槽狀態仍為 `未知`，不由目視自動命名。
  SPEC-002 仍為 DRAFT，未建立 production importer 或 Go／Ebitengine runtime。

## RE-028：CHARPICS.0 全 80 張角色模板

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832`；IDA Pro 9.4、loaded-MZ linear EA。
  `sub_11820` 載入 `CHARPICS.0` 後，以 `sub_13D7D` 將選定圖號解碼到
  `word_20C2A+0x400`，再立即交給 `sub_15335` 顯示。
- `sub_13D7D:0x13d7d..0x13e25` 以 `20+2*image_index` 取得 offset，呼叫
  `sub_139A8` 解碼。固定 archive 恰有 80 筆，全部機械解碼為 761 bytes；RE-027 的
  40×38 consumer 只消費前 760 bytes，最後一 byte 全為 `0x00`，分類為非像素尾端。
- 80 張 ignored PNG 全部含 9–15 種 EGA 色彩；10×8 接觸表顯示方向一致的完整人物、
  武器、盾牌、服裝與坐騎組合，未見 nibble、列寬、旋轉或裁切錯誤。此目視支持「完整角色
  模板」分類，但不證實逐張職業、裝備或姿勢名稱。
- 兩份全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `b4877736…ff30`／`18d44d22…f74b`；兩輪 80 張 PNG、接觸表與視覺摘要亦逐 byte 相同。
  視覺摘要／接觸表 SHA-256 為 `0bcb3709…7575`／`cd0a65db…a5a4`。
- 新增可重生 IDA／PNG 探針與安全摘要；SPEC-002 仍為 DRAFT，未接 production importer。

## RE-029：TWN0..7PICS 全 243 筆城鎮圖像

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4、重建
  unpacked MZ linear EA。安全探針保留原函式名、EA、operand 與 caller；完整指令只存
  本機 ignored private JSON。
- `sub_11302:0x11302..0x11390` 證實一般城鎮圖矩形：目前城鎮 selector 為 0 或圖號為
  奇數時使用 192×96；非零城鎮的偶數圖號使用 88×64。`sub_14047:0x14047..0x14189`
  另證實非零城鎮圖號 0 是 56×57 彩圖，與 `KOLMPICS.0` 圖號 28 的 AND-mask 依
  `(destination AND mask) OR color` 合成。`sub_15920` 證實 `twn%d%s` 檔名路由；`.1/.2`
  延續卷仍沿用 RE-008 已證實的 `sub_1A4FE` 分卷契約。
- 固定八包共 243 筆：TWN0 的 25 筆全部精確展開為 192×96；TWN1..7 的圖號 0 都是
  56×57，圖號 1 都是單一 byte 的非圖片 placeholder，偶數圖號 2 起都是 88×64，奇數
  圖號 3 起都是 192×96。故輸出 236 張 PNG，七筆 placeholder 明列排除。
- 三份接觸表人工抽查涵蓋 TWN0、TWN1、TWN7：TWN0 為連續場景；後兩包呈交錯人像、
  建築與場景，方向、色盤、透明輪廓及裁切一致。外觀只作視覺驗證，不據此命名 NPC、
  地點或劇情用途。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `123cf073…ca52`／`76b2f51d…f360`。兩輪 236 張個別 PNG、八張接觸表與摘要亦逐 byte
  相同；視覺摘要 SHA-256 為 `13a3b6d8…dce3`。安全摘要見
  `ida-fingerprints/KOLTOWN-twn-archives-summary.json`，視覺摘要見
  `twn-archives-ega-visual-probe.json`。
- TWN 的 EGA payload、幾何、遮罩與排除項已關閉；SPEC-002 仍為 DRAFT，因首版顯示
  theme 範圍尚待使用者決定，且尚未建立正式、原子化的全包抽取器。

## RE-030：KOLPICS raw-planar extraction geometry 勘誤

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4、重建
  unpacked MZ linear EA。這一節追加勘誤，不重寫 RE-025 當時的保守結論。
- `sub_19B29:0x19b41..0x19ba9` 每輪從 256-byte source stride 的前 128 bytes 讀四個
  32-byte planes，寫出 256 個 chunky color indices，並以 output offset 與 caller endpoint
  比較決定下一輪。圖號 0 caller 的 endpoint 是 `0x500`，因此正好是五個 converter block；
  其無損資料 PNG 形狀固定為 256×5。這是 renderer data serialization，不是螢幕矩形。
- 圖號 8／12 各有 10,241 decoded bytes：40 個完整 256-byte stride 加一個共同 `0x00`
  非像素尾 byte，故來源恰產生 10,240 indices。下游 `0x1eafe..0x1eb06` 以
  `y<<6+x` 存取，固定 64 欄；完整來源 grid 因此是 64×160。這只宣稱來源資料形狀，
  不宣稱單次玩家路徑會觸及每一格。
- 兩份全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `a8ee8476…e63d`／`c6538475…45a5`。兩輪八張 PNG、接觸表與 v2 視覺摘要亦逐 byte相同；
  摘要／接觸表 SHA-256 為 `82281d7a…8f67`／`d6f3a288…524e`。接觸表與舊正確排列完全
  相同，未見 serialization 或 plane-order 退化。
- RE-025 的「160 列與圖號 0 幾何仍為強推論」現訂正為：original display geometry
  不適用；production extraction geometry 為已證實的 renderer-data PNG。六筆 placeholder
  仍排除，KOLPICS 全包不再有 `excluded_unknown`。SPEC-002 尚因 KOLMPICS 與 theme 決策
  維持 DRAFT。

## RE-031：KOLMPICS 68 筆圖示的 canonical extraction shape

- 固定 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`；IDA Pro 9.4、重建
  unpacked MZ linear EA。`sub_11722` 以三個 byte-columns×20 scanlines、`sub_23638`
  以 `x+3`／`y+20` 建立兩條獨立 24×20 consumer 契約。
- 排除圖號 0／2、圖號 28 AND-mask 與 12 筆 placeholder 後，其餘 68 筆全部經相同二階
  RLE 精確得到 480 個 4-bit indices。完整 archive-order 接觸表在 24×20 下呈現連續的
  手勢、武器、狀態、人物與方向圖示，沒有 row-width、plane-order、旋轉或裁切異常。
- 三項證據共同證實 68 筆的 canonical 24×20 lossless extraction shape。這不等於每個
  dynamic caller 都使用相同螢幕座標或 clipping rectangle；`sub_27E20` 的 caller-specific
  display geometry 仍未由本切片宣稱，不能拿 extraction PNG 尺寸冒充畫面 parity。
- 兩份全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `93a4a3be…f6ef`／`cf44a187…a160`。兩輪 73 個 PNG 與 v2 摘要逐 byte 相同；摘要／接觸表
  SHA-256 為 `2647e6a6…572e`／`2f335771…15a`。
- KOLMPICS 全 83 筆現為 71 筆已證實來源與 12 筆 `excluded_non_image`，不再有
  `excluded_unknown`。完整 EGA extraction set 的 pending 因此歸零；SPEC-002 仍等待
  顯示 theme 決策與 production extractor 設計，未接 Go runtime。

## RE-032：TITLE PC Speaker cue 與 TANDYPIN 排除

- 輸入：`TITLE.EXE` SHA-256 `8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6`；
  `TANDYPIN.EXE` SHA-256 `036c39f330a4fe09bc2729756816b869f1154515bf4befba190e767d4e3cdce3`。
- 工具／位址空間：IDA Pro 9.4；DOS MZ 載入後的 IDA linear effective address。安全摘要為
  `ida-fingerprints/TITLE-speaker-summary.json` 與 `TANDYPIN-residency-summary.json`。
- **已證實**：TITLE `sub_10F33` 啟用 PC Speaker、`sub_10F48` 關閉 speaker、`sub_10F5D`
  將除數寫入 PIT channel 2。`sub_10F75` 執行兩段音爆，每段 100 次除數更新；
  `sub_11953` 呼叫 `_rand` 並回傳 `rand() % N + 1`，`sub_11904` 為忙等延遲。
- **已證實**：`TANDYPIN.EXE` 入口執行 `INT 21h`、`AH=31h`，以 `DX=0x0401` terminate-and-
  stay-resident；入口沒有 speaker／PIT I/O，因此排除為獨立音訊素材或音效播放器。
- **限制**：`title-random-tone-burst` 只是導覽標籤；不宣稱原版語意名稱。依硬體時序停止線，
  不追逐週期 PIT／busy-wait parity。其餘 executable 尚未完成 cue 母體掃描。
- 可重生性：TITLE 與 TANDYPIN 各以兩個全新 IDA DB 執行，兩份 safe JSON 都逐 byte 相同。

## RE-033：全 executable 音訊 I/O 母體與正常流程 WAV oracle

- 六個 executable 均以兩個全新 IDA DB 窮舉 IDA-defined code 的 x86 `IN`／`OUT`。
  KOL／KOLBOOK／KOLTOWN-unpacked／MAKEMODE／TANDYPIN／TITLE 的 I/O 指令數依序為
  6／43／54／0／0／12，只有 TITLE 六個 instruction 命中 `0x61`／`0x42` 音訊埠。
- **已證實**：KOLBOOK 四個、KOLTOWN 兩個保守未解析的 `DX` I/O 點都有原始 EGA
  graph-register 註記，且位於 RE-027 等既有圖片 renderer 垂直鏈，排除為音訊。
- **已證實**：TITLE cue 八個 caller 位於 `_main`、`sub_10AF0`、`sub_10C86`，全都緊接
  `sub_11133` 圖片 blit；單次 cue 路徑隨後進入 `_kbhit`。這是正常 TITLE presentation flow。
- DOSBox-X 2026.06.02 依官方 `DX-CAPTURE /A /-D` 契約取得正常流程 WAV：48 kHz stereo
  PCM、38.229333 秒、7,340,076 bytes、peak -10.317424 dB，SHA-256
  `f2c5191545b3bae1f066c9328db184c9406541941c8a008426d95a8f2260936b`。四段非靜音窗見
  `title-audio-oracle.json`。
- **限制**：機械擷取證實非靜音輸出，不等於人耳驗收；DOSBox-X timing 不冒充原硬體
  wall-clock parity。WAV 為本機 ignored oracle，不加入 Git。

## RE-034：角色建立手冊契約與 CHARDATA 路由

- 手冊實體頁 4–5、27–31 與 Apple II 參考卡實體頁 2 固定角色建立順序、四種族、性別限制、
  不限次 reroll、圖像接受、16 名角色儲存上限與六人隊伍上限。
- 固定 ZIP 的 `CHARDATA` 是 12,288 bytes＝16×768；`CHARPICS` 是 12,160 bytes＝16×760。
  這與手冊 16 名上限交叉吻合，但不證實 record 內欄位。
- **已證實**：KOLBOOK loaded MZ 的 `chardata` 位於 EA `0x1945a`、DGROUP offset `0x021a`；
  `sub_134C2` 以 `arg_0<<4` 選 16-byte filename entry。五個立即 `arg_0=0` caller 位於
  `_main`、`sub_11820`、`sub_1293E`、`sub_135DB`。
- 兩個全新 IDA DB 的 safe JSON 逐 byte 相同，SHA-256 `e71a6ca3…ec5d`。
- **限制**：職業表數字是手冊標示的平均值，不能當亂數上下界。RNG、初始技能／物品 writer、
  768-byte 欄位與保存 transaction 仍未知；SPEC-005 維持 DRAFT。

## RE-035：KOLBOOK 職業屬性亂數與衍生公式

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832`；IDA Pro 9.4 loaded-MZ linear EA。
  `_main:0x1001e..0x1015b` 以 BIOS `INT 1Ah` 的低 tick word 為一般 seed；顯示模式值 7
  分支改傳 16。`_rand` 唯一 caller `sub_14BE4` 已證實回傳 `rand()%N+1`。
- 本節初次把兩個 `[bx+7772h]` load 誤讀為同一 record offset，因而寫成骰式 byte 同時
  是基值。後續回查前一指令 `0x11cb6 add bx,8` 已否定該結論：offset 0..6 是骰式，
  offset 8..14 才是獨立內部基值。形成原因與訂正保留於 RE-036，不把舊誤判藏起來。
- EA `0x1a47b`、`0x1a47c`、`0x1a47d` 分別是力量／智力、力量／健康、體型／健康的
  無號整數平均，全部向下取整。欄位語意由手冊順序與公式交叉分級，不以改名代替證據。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  當輪 `f8ff1103…187b`／`9575adcf…d7f`；此版本現由 RE-036 的修正版摘要取代。完整
  指令及 `.i64` 不進版控；本輪比對完成後已清除，可由固定探針與輸入重生。
- **限制（當時）**：尚未關閉 35 個職業 enum 與 16-byte 記錄逐項對應，也未證實 DOS library
  `_rand` 的完整序列；SPEC-005 維持 DRAFT，不進 production implementation。

## RE-036：職業記錄、名稱 selector 與屬性顯示量尺

- `BOOKDATA` 項目 2 固定為 424-byte 壓縮流；LZW 解成 625 bytes、SHA-256
  `848ad877…b321`，結構是 39×16-byte 記錄加一個零尾 byte。每筆 offset 0..6 是七個
  骰式 `D`，8..14 是七個內部基值 `B`，7／15 固定為零。內部公式訂正為
  `B + Σ roll(dice[D>>4])`，次數 `D&0x0f`。
- `sub_12677:0x126b3..0x126c4` 對十項內部屬性顯示
  `floor(value×100/150)`；因此手冊的 0–100 平均表不能直接與 raw byte 比較。
- `sub_14101` 以 CR 遍歷 `BOOKDATA` 文字項目；行 14..19 固定 HUMAN／ELVEN／DWARVEN／
  KELDEN 與 MALE／FEMALE enum，行 21..53 固定 33 個 menu names。`sub_11820` 的 selector
  將它們映到 12＋4＋6＋6＋8＋3＝39 筆記錄；男女精靈共用名稱而使用不同記錄。
- 矮人記錄 `[28..31]` 與 `[32..35]` 成對逐 byte 相同，對應手冊把八個 menu names 合併
  為四列共同數值；故「35 個手冊列」與「39 個 executable 記錄」不再互相矛盾。
- 最終兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `e3309787…1116`／`a6e07f2a…8401`；兩輪 archive safe／private 為
  `aaa112ef…cb2d`／`7d13f6ea…9b36`。private raw records 與 `.i64` 不進版控。
- **限制**：本輪尚未關閉初始技能／物品 writer、768-byte 存檔欄位與 DOS `_rand` 完整
  序列；SPEC-005 維持 DRAFT，不授權 production implementation。

## RE-037：角色 writer、起始武器與三檔保存順序

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832`；IDA Pro 9.4 loaded-MZ linear EA。
  `sub_11820` 清空 768-byte CHARDATA slot，寫入名字、selector、職業、七項屬性、種族、
  起始金冠、武器模板與武器技能；同槽另寫 760-byte CHARPICS 與 SHLDPICS。
- EA `0x1940c` 的 39×2-byte 金冠表及 EA `0x19348` 的 39×5-byte 武器／技能 metadata，
  配合 BOOKDATA 項目 3 的 14×16-byte 武器模板，全部 39 筆均與中文手冊職業表一致。
  項目 4 則提供 760-byte 初始盾牌模板。
- 三份工作區依序保存為 `chardata` 0x3000 bytes、`charpics` 0x2f80 bytes、`shldpics`
  0x2f80 bytes。這只證實呼叫順序，不宣稱原版具跨檔案 transaction 或 rollback。
- `region%d.dat` 每筆 48 bytes、最多五筆；固定 ZIP 只有 REGION0。固定 source→record
  offset 路由已收入安全摘要，所有 region 欄位語意仍標未知。
- 最終兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `39beef72…b178`／`099e0d2b…9f73`；archive safe／private 為
  `ae711e3b…2176`／`f42fbbf1…97bc`。完整函式、operand 掃描及 `.i64` 不進版控。
- **限制**：只關閉起始金冠、武器與武器技能，不能推廣成全部一般技能／物品。最小 writer
  offset 已建立，其他欄位 reader、圖像 selector、UI 邊界與失敗路徑仍未知；SPEC-005
  維持 DRAFT，不授權 production implementation。

## RE-038：writer coverage 與 KOLTOWN 最小 reader

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832` 與 unpacked KOLTOWN derivative
  `01f6f9…d1d2`；兩者皆使用 IDA Pro 9.4 linear EA，從全新一次性 DB 匯出。
- **勘誤**：RE-037 將 CHARDATA `0x4a..0x51` 暫稱主要武器輔助區塊；回查 source
  `BOOKDATA` item 3 與完整 copy loop，證實它是主要武器模板 bytes 0..7，`0x52` 是同模板
  byte 13。舊稱不再作語意依據。
- writer coverage 固定 `0x28=(race+1)*16`、`0x29=floor(sum(meta[1:5])/200)` 與
  `0x7f=3` 等先前漏列的直接寫入。由於 slot 先完整清零，`sub_11820` 其後全部直接寫入均
  落在欄位表或 REGION route；沒有其他職業衍生起始技能／物品 writer。
- KOLTOWN `sub_147DD` 讀 16-byte 名字，`sub_14DB1` 讀 race／profession eligibility，
  `sub_14CCF` 讀 little-endian gold，`sub_18F3A` 讀主要武器模板並路由 renderer。這是角色
  建立所需最小 reader 交叉查核，不是完整 768-byte schema。
- KOLBOOK safe／private 兩輪 SHA-256 為 `2884a1ae…c129`／`6c00db3a…fdb6`；KOLTOWN
  safe／private 為 `16f8b24d…67ba`／`dcfa8c59…18e1`，各自逐 byte 相同。
- **限制**：武器技能仍由 writer＋手冊交叉證實，本切片不替其他一般技能命名；REGION
  欄位、正常 UI 邊界、圖像 selector 與失敗 rollback 仍未知。SPEC-005 保持 DRAFT。

## RE-039：KOLBOOK 角色建立 UI 返回與滿槽邊界

- 固定 `KOLBOOK.EXE` SHA-256 `2cada028…e5832`；IDA Pro 9.4 loaded-MZ linear EA，所有
  附加語意均保留原始 `sub_11820`、全域名稱、EA、instruction bytes 與控制流目的地。
- 名字 helper 容量為 `0x10` bytes；空第一 byte 且未取消會回到名字輸入，取消則離開建立。
  種族取消回名字、secondary 取消回種族；職業依 selector 回 secondary 或種族；屬性選項
  0 回完整擲骰起點，取消回職業。
- 肖像接受左／上／右／下，selector 在 `0..79` 循環；Return 接受，取消回 commit 前流程。
  slot scan 以 `0x300` stride 走到 `0x3000`，全 16 槽非空時於 `0x12061` 退出，早於
  `0x12064` 的記錄清零與後續 writes。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `9ecd4b76…aa82`／`ef84a41e…50c`。private 完整函式與 `.i64` 不進版控。
- DOSBox-X 實驗訂正啟動方式：直接 `KOLBOOK.EXE A` 只停在黑畫面；`KOL.EXE` 才是正常
  玩家入口並進入 EGA 開場動畫。尚未走到主選單，因此本節只宣稱靜態控制流已證實，正常
  UI 截圖、重名、16 字元終止呈現、保存失敗 rollback 仍未知；SPEC-005 維持 DRAFT。

## RE-040：正常 DOS 玩家路徑的角色建立 UI oracle

- 輸入為固定 `Knights-of-legend.zip`（SHA-256 `2b455e3c…9f46`），唯讀掛載後於容器
  tmpfs 重建遊戲目錄。容器映像 `rich2-dosbox-dbg:latest`（DOSBox-X 2026.06.02 SDL2），
  機器設定固定 `machine=ega`、`memsize=4`、`core=normal`、`cputype=386`、
  `cycles=fixed 3000`；設定檔 SHA-256 記在每份收據內。
- 擷取器 `tools/dosbox_character_ui_oracle.py` 以 X11 XTEST（`tools/x11_key_inject.py`）
  送鍵，並在畫面靜止到指定秒數後才送下一鍵。先前用 DOSBox-X `AUTOTYPE` 的固定時序
  盲打，同一顆按鍵在不同執行的反應延遲介於 1.5 秒到 11 秒，無法對齊選單層；改成畫面
  驅動之後四個情境全部穩定完成。
- 正常入口鏈已證實：`KOL.EXE` → TITLE 標題序列（循環動畫）→ 任意鍵離開 → KOLBOOK
  書本封面 → 任意鍵 → 目錄頁 → `5` → Character options → 反白條下移一格 →
  `CREATE CHARACTER` → New Character 名字層。
- **名字欄容量 16 個可見字元，輸入第 16 個字元即前進到種族層，不需 Return。**
  連送 17 個小寫字母時第 15 個字元顯示 `abcdefghijklmno`、第 16 個顯示
  `abcdefghijklmnop` 並同時開啟種族選單；第 17 個字元與其後兩次 Backspace 落在種族層
  而無作用。26 個檢查點在兩次全新執行之間只有一張名字層畫面因游標相位不同
  （0.0125%），其餘逐 byte 相同。擷取器送出的是未加 Shift 的小寫字母，畫面與
  `CHARDATA` 中的名字也是小寫，因此不對大小寫保存行為下任何結論。
- **空名按 Return 不前進**，畫面停在名字層，與 `sub_11820` 的「輸入 buffer 首 byte 為零
  時回到同一輸入點」一致。
- 取消目的地：種族→名字、性別→種族、職業→性別、屬性→職業，再往上依序回到性別、
  種族、名字層；回到名字層時名字與右頁摘要欄位皆清空。與 RE-039 的靜態控制流一致。
- `REROLL STATS` 重擲九個顯示值；HUMAN／MALE／BARBARIAN 連續兩次 reroll 的力量為
  88 → 85 → 89、體型 74 → 72 → 73、健康 75 → 75 → 74。
- 四個情境全程未接受屬性或角色圖像，`CHARDATA`、`CHARPICS`、`SHLDPICS` 的前後
  SHA-256 均未變動。
- **角色圖像 selector 恰好循環 80 張**：由 selector 0 向左一次得到的畫面，與由 0
  連續向右 80 次得到的畫面相同；再向右一次回到 selector 0。這在正常玩家路徑上證實
  RE-039 由靜態控制流得到的 `0..79` 循環。
- **接受圖像是保存點**：走完「ACCEPT STATS → 接受圖像」後畫面回到 Character options，
  `CHARDATA` 出現變動。兩輪都只寫入 slot 14，`CHARPICS`／`SHLDPICS` 的 SHA-256 未變
  （本批取 selector 0，其模板與該槽原有內容相同）。逐 offset 差異見
  `docs/verification/asset-parity/chardata-write-run{1,2}.json`；差異清單是「值有變動的
  byte」，不是 writer 覆蓋率。
- 六個情境各執行兩次全新流程並以 `tools/compare_character_ui_runs.py` 比對：
  `blank-name` 10/10、`name-length` 25/26、`escape-layers` 21/28、`reroll` 12/20、
  `accept-and-save` 12/16、`picture-cycle` 10/99 檢查點逐 byte 相同。全部差異都落在三類
  本來就不可能相同的畫面，且差異量已量化——TITLE 循環動畫落點 13.98%、名字欄閃爍
  游標 0.0125%、BIOS tick 播種的屬性數值 0.1%–0.25%。`picture-cycle` 的多數檢查點屬
  第三類，其角色圖像區塊另由 RE-041 的 parity 收據確認兩輪都 80/80 相同。
- **限制**：16 槽全滿的訊息與退出行為只有靜態控制流證據。固定 ZIP 的 `CHARDATA`
  十六槽都非全零，前十四槽有名字，第 15、16 槽名字為空但記錄非零；本批保存落在
  slot 14，支持「以名字首 byte 判斷槽位可用」的假說，但滿槽分支本身未由 UI 證實。
  重名處理、名字欄字元集合與大小寫保存方式仍未知。1,368 個去重畫格與收據已入
  private repo 的 `docs/verification/character-ui/`。

## RE-041：EGA production 抽取器與原版同狀態資產 parity

- `tools/extract_ega_assets.py` 由固定 ZIP（SHA-256 `2b455e3c…9f46`）產出
  `assets/original-derived-png/ega/`：30 個邏輯 archive、752 張 PNG，等於 SPEC-002
  已證實 PNG 單位的全部。輸出先建在暫存目錄、整批驗證後才替換，輸出根目錄被限制在
  `assets/original-derived-png/`。
- 抽取計畫直接取自 `docs/re/*-visual-probe.json` 的逐張紀錄；每張 PNG 都重讀驗證
  （簽章、chunk CRC、IHDR、解壓列數）並與該紀錄的 `png_sha256` 核對，任何一項不符即
  整批失敗。manifest 記錄各 archive 的 probe 檔與其上游 IDA 摘要雜湊。
- 組成方式分六類：direct RLE nibbles、固定 40×38 packed nibbles、彩圖＋AND-mask 配對、
  KOLMPICS 的 rle-color／and-mask／placeholder 混合、TWN 的共享 KOLMPICS 圖號 28 遮罩、
  KOLPICS raw-planar 與 WEATPICS 分層資料。26 筆非圖片 placeholder 明列排除。
- **原版同狀態 parity**：80 張 `CHARPICS` 模板對原版角色圖像選擇畫面，在
  `tools/locate_asset_in_frame.py` 搜出的 blit 座標 (208, 106) 上，兩輪擷取各 80/80
  逐像素相同（差異 0）。`BOOKPICS[0]` 對書本封面畫面 64,000 像素中只有 50 個不同，
  集中在 (160, 100) 起的 8×11 區塊，即原版畫的滑鼠游標。
- **限制**：其餘 29 個 archive 目前只有「與 visual probe 紀錄雜湊相符」這層驗證，
  沒有各自的原版同狀態畫面收據。KOLPICS 的 raw-planar 與 WEATPICS 的遮罩層是 renderer
  資料，不是原版顯示矩形。PCGA／HERC 未納入。

## RE-042：原版內嵌點陣字型

- 書本內文的字模由原版畫面反推：`RACE : HUMAN` 一行的 `HUMAN` 五個字模逐 byte 取出，
  在固定 ZIP 的所有檔案與所有 archive payload 中搜尋，命中 `BOOKDATA.0` 與 `KOLDATA.0`
  的第一個 payload。兩者解壓後為同一份 1,303-byte 區塊，SHA-256 `819e4f82…d318`。
- 該區塊的結構已固定：`0x000`–`0x2e7` 是 93 個 8-byte 字模（8 列），`0x2e8`–`0x515`
  是 93 個 6-byte 字模（6 列），最後 1 個 byte 不屬任何字模。兩套都由 ASCII 32 起算，
  逐 byte 一列點陣、MSB 在左，涵蓋 ASCII 32–124。小字型即書本內文所用者。
- `tools/extract_fonts.py` 產出 `assets/original-derived-png/fonts/`：兩張 RGBA 字模表
  （墨色不透明白、其餘透明）與含逐字模 hex、來源位移、字距的 manifest。
- 大字型已指認的畫面用途是角色圖像層右側的上下箭頭：codepoint 94（`^`）與 95（`_`）
  的字模就是箭頭圖形，逐 byte 命中（RE-046）。
- **限制**：大字型是否另有其他畫面用途尚未盤點。字模表以外的 744 bytes 之前區段沒有
  其他推測。ASCII 125 以上與非 ASCII 沒有字模，中文化必須另行處理。

## RE-043：書本文字版面與原版繪字規則

- `tools/locate_text_in_frame.py` 用抽出的字型在原版畫面上逐位置比對墨點顏色一致性，
  並要求文字框內的非墨點大多不同色（否則單色背景會整片誤命中）。
  `tools/extract_text_layout.py` 以同一機制產出 `docs/re/text-layout.json`，
  每個字串必須唯一命中，否則整批失敗。
- New Character 頁左半頁的靜態版面已固定：標籤全在 x=24；NAME 52、RACE 68、SEX 84、
  CLASS 100（行距 16）；屬性列自 STRENGTH 118 起每 6 px 一列到 BODY PTS. 172。
  冒號不緊接標籤而在固定字元欄：名字／種族／性別／職業列在字元欄 5（x=64），
  屬性列在字元欄 10（x=104）。顏色為 EGA 青（0,170,170），作用中的名字列為
  棕（170,85,0）。標題圖是 `BOOKPICS[10]`，位置 (32,16)，由資產搜尋逐像素命中。
- **原版繪字規則**已由逐像素比對證實：每個字模格先清成白底再畫墨色，但**空白字元
  不清底**——字串內的空白處會保留書頁的黃色斑點。最後一列的實際文字是 `BODY PTS.`。
- remake 依此規則重繪整頁後，與原版空白 New Character 畫面只差 31 個像素，全部落在
  名字欄位置：原版在種族層取消後於該處殘留一個字元（本例為前次名字首字 `t`，
  字模已逐 byte 命中）；成因未確認，remake 也尚未實作名字輸入狀態。
- **限制**：其餘畫面（角色選項、屬性數值、右頁選單）的文字版面尚未量測。中文化的
  字框限制已浮現：原版字模格是 8×6、字距 8，繁體中文字形無法放進同一格，
  版面策略必須另行決策。

## RE-044：TITLE 標題序列的原版畫面時間軸

- 觀察模式（全程不送按鍵）兩輪各擷取 100 秒，得 40／41 個去重畫格；收據與畫格存於
  `docs/verification/title-sequence/`。
- `tools/identify_frame_assets.py` 以抽取資產逐張逐像素比對，9 個畫格與 `BACKPICS`
  的全螢幕素材完全相同：圖號 4、5、6、7、8、9、10、11、12。
- 序列結構已固定：`BACKPICS[6]` 反覆作為段落之間的畫面，其間依序出現
  7、8 →（閃爍段）→ 6 → 9、10 →（閃爍段）→ 6 → 11、12 →（閃爍段，含 4、5）→
  10 → 6，約 86 秒後回到同一畫面重新開始。
- **限制**：閃爍段在少數畫格之間高頻交替，機制未確認，不由外觀推定是重繪、色盤動畫
  或逐段清除。畫格時間戳是主機牆鐘、取樣間隔 0.3 秒，因此時間是觀察值而非原版設計
  時序；逐幀時序對拍需另外固定週期與取樣方式。remake 尚未實作標題序列。

## RE-045：角色建立逐畫面對拍與展開跨頁的貼圖殘影

- remake 以固定腳本 `t e s t RET RET RET RET`（seed 1）重播名字→種族→性別→職業→
  擲屬性，九張畫面對原版 `reroll-run1` 的同狀態畫格。計畫
  `docs/verification/remake-parity/plan-newchar-steps.json`、收據 `newchar-steps.json`，
  9 對全通過，允許差異區以外零像素差。
- 殘餘差異只有四類，每一類都有機械證據：
  1. **頁尾殘影 112 px**（每一張都有）。原版把 `BOOKPICS[2]` 貼到前一個書本畫面之上時，
     左頁底緣有一條未被覆蓋的細帶：y=190 的 x=24–159、(162,192)、(163,197)、(164,197)、
     y=198 的 x=151–169。這 112 個像素**每一個都等於 `BOOKPICS[0]`**；封面與目錄頁在該帶
     內完全相同，因此分辨不出壓在下面的是哪一張。差異是雙向的——有的位置原版紅、素材黑，
     有的相反——所以不是解碼錯誤；`BOOKPICS[2]` 解出 64000 像素且與探針記錄的
     `pixel_sha256` 相符。以 `BOOKPICS[2]` 的黑色像素為條件全畫面檢查，只有 100 個
     透出前一畫面，因此也不是「索引 0 當透明」的整體規則。
  2. **名字輸入游標 8 px**：游標會閃爍，取樣落在哪一相位由畫面靜止的時點決定。
  3. **DOSBox 滑鼠指標 55 px**：remake 沒有這個元素。
  4. **屬性數值欄 163 px**：原版以 BIOS tick 播種，remake 以固定 seed 重播；骰式相同、
     點數不同。標籤與冒號不在允許區內，仍逐像素驗證。
- **標籤高亮規則**由整輪畫面簽章證實（`tools` 逐列統計棕／青像素，`picture-cycle-run1`
  與 `reroll-run1` 一致）：名字／種族／性別／職業層各自只有對應那一列是棕色；
  **擲屬性層十列屬性同時變棕**（frame-082，1668 px 由青轉棕）；**圖像層沒有作用中的列**
  （frame-084 起十四列全青）。
- **限制**：貼圖殘影的成因（貼圖矩形怎麼切）未反組譯確認，只證實「殘留內容等於前一畫面」。
  remake 直接畫 `BOOKPICS[2]`，因此不重現該細帶；要逐像素重現得先反組譯貼圖常式。

## RE-046：角色圖像層的版面與完整流程逐畫面對拍

- remake 以固定腳本 `z e d RET RET RET RET DOWN RET LEFT` 加 81 次 `RIGHT`（seed 1）
  重播完整建立流程並走完 80 張圖像的循環與兩端回繞，92 張畫面對原版
  `picture-cycle-run1` 的同狀態畫格。計畫
  `docs/verification/remake-parity/plan-newchar-pictures.json`、收據
  `newchar-pictures.json`，92 對全通過，允許差異區以外**零像素差**。
- 圖像層的版面由原版畫面機械量得：
  - 右頁上半是五行說明，青色，起點 (176,52)、行距 6：`USE ARROW KEYS` /
    `TO SELECT A` / `CHARACTER` / `PRESS RETURN` / `TO ACCEPT`。
  - 右頁下方是棕色的 `CANCEL   ACCEPT`，起點 (176,166)。
  - 角色圖像放在一個青色雙格外框內：框線在 x=206 與 x=248、y=105／144／152，
    兩格內部填黑。上格內部 41×38，圖像 40×38 貼在 (208,106)，左側留一欄黑。
  - 下格是名牌：角色名以書本內文字型（8×6）白字畫在黑底上，起點 (208,146)，
    不另清字模格。
  - 右側 (264,105) 與 (264,144) 各一個 8×8 字模，分別是大字型的 codepoint 94 與 95，
    逐 byte 命中；畫法與書頁文字相同（字模格先清成白底再畫黑墨）。
- 全部字串都是用 `tools/read_text_from_frame.py` 以抽出的原版字型逐格讀回的，不是人眼抄。
- 殘餘差異仍只有 RE-045 的四類：頁尾殘影 112 px、名字游標閃爍 8 px、滑鼠指標 55 px、
  屬性數值欄 152 px。80 張圖像的 blit 矩形 (208,106,40×38) 列為必須相同區，全部逐像素相同。
- **限制**：圖像層的接受動作（寫入 CHARDATA）尚未實作，因此 `accept-and-save-run1` 的
  三檔前後 SHA-256 收據還沒有 remake 側的對照。

## RE-047：TITLE 標題序列是一段可靜態讀出的線性腳本

- `TITLE.EXE` 的 `_main`（`0x10010`–`0x10af0`，1051 條指令）就是整段標題序列，不是資料表。
  新增 `tools/ida_probe_title_sequence.py` 匯出 `_main` 與其全部直接被呼叫者，並把每個
  呼叫點的引數解析出來（Borland 小模式的 `mov ax, IMM` + `push ax` 樣式）。兩個全新
  IDA DB 的 safe／private JSON 逐 byte 相同，安全摘要在
  `docs/re/ida-fingerprints/TITLE-sequence-summary.json`。
- **時間是精確的 BIOS tick，不是牆鐘**：`sub_11022`（`0x11022`）記下目前 `int 1Ah` 的
  tick，`sub_11039(n)`（`0x11039`）忙等到 tick ≥ 記錄值 + n。因此每一段等待都能靜態讀出。
  全螢幕畫面之間一律是 75 tick（≈4.12 秒）。
- 序列結構（`byte_1423A = 1` 的路徑，`0x101a4` 設定，代表第八次 `__dos_allocmem` 成功）：
  閃電 → SET0 動畫 → `BACKPICS[7]` → `[8]` → SET1 動畫 → `[9]` → `[10]` → SET2 動畫 →
  `[11]` 或 `[13]` → `[12]` → FACE 動畫 → `[10]` → 跳回開頭循環（`0x10a06`）。
  `byte_1423A = 0` 的路徑畫面順序與等待完全相同，只是少一份常駐緩衝。
- 先前記為「機制未確認的閃爍段」是 SET 動畫：`sub_10AF0(set)`（`0x10af0`）跑 4 圈，
  每圈把 SET 的第 0 張疊到 `BACKPICS[0]` 的離屏副本後整張送出，接著把第 1、2、3 張
  直接畫到顯示記憶體，等待依序是 3、4、4、2 tick；然後 `rand()%2` 選第 4 張疊
  `BACKPICS[1]` 或第 5 張疊 `BACKPICS[2]`，放一次雷聲，再等 5 tick。四圈之後接一次閃電。
  疊圖矩形隨 set 固定：SET0 (112,53,88×70)、SET1 (96,53,104×84)、SET2 (48,53,160×125)、
  FACE (104,80,112×80)，與四個 archive 的實際尺寸完全吻合。
- **兩處亂數**：SET 動畫每圈的閃光張數（第 4 或第 5 張）、以及 `[11]` 與 `[13]` 的選擇，
  都由 `sub_11953(2)`（`rand()%2+1`）決定，而 `_srand` 的種子是開機時的 BIOS tick
  （`0x100a8`／`0x100b2`）。**因此原版每次播放的閃電並不相同，逐幀對拍不可能要求完全一致。**
- 雷聲 `sub_10F75`（`0x10f75`）是喇叭噪音：兩輪、每輪 100 步，頻率為
  `500 × 輪次 + rand(10000) + 1000`；步進間隔是用開機期校正值算出的忙等，不是 tick 數，
  所以雷聲長度不能靜態讀出。
- **與觀察時間軸交叉核對**：靜態讀出的順序完整預測 `docs/verification/title-sequence/`
  observe-run1 的每一個轉換點（閃電 7.98／SET0 10.1–16.2／閃電 17.30／`[7]` 19.13／
  `[8]` 23.06／SET1 27.3–33.8／…／`[10]` 81.41／下一輪 85.48）。畫面間距觀察值
  3.80–4.08 秒與 75 tick 的 4.12 秒相差都在 0.3 秒的取樣間隔內。
- **限制**：雷聲長度、開機期的忙等校正值（`sub_118D1`）與 `sub_11AAC` 的作用尚未讀出。
  remake 尚未實作這段序列。

## RE-048：翻書動畫是表驅動的逐欄捲動

- `KOLBOOK.EXE` 的 `sub_114F5`（`0x114f5`–`0x11635`）是封面翻開成目錄頁的動畫。
  新增 `tools/ida_probe_function_script.py`（通用的呼叫腳本匯出）與
  `tools/ida_probe_kolbook_book_animation.py`（讀出兩張參數表）。
- 動畫由兩段迴圈驅動，參數表放在 DGROUP：
  - 第一段 `dgroup+0x64`，17 筆、每筆 2 bytes `(a, b)`，呼叫
    `sub_15745(a+0x14, b, a+0x14+b, 來源緩衝)`。
  - 中間一步 `sub_15745(0x14, 0, 0x14, 來源緩衝)`。
  - 第二段 `dgroup+0x86`，18 筆、每筆 3 bytes `(c, d, e)`，呼叫
    `sub_15745(c+2, d, e+2, 來源緩衝)`。
  - 每一步之後 `sub_14094(3)`，等 3 個 BIOS tick。
- `sub_15745` 依顯示模式分派；EGA（mode 0Dh）走 `sub_168B0`。`sub_168B0(arg0, arg2,
  arg4, buf)` 先呼叫 `sub_1690A(arg0, arg2)`，再把來源緩衝的第 `arg4` 個 byte 欄複製到
  畫面的第 `arg0+arg2` 個 byte 欄。
- `sub_1690A(start, span)` 把畫面上第 `start` 欄起的 `span` 個 byte 欄**整體左移一欄**：
  設 sequencer map mask 為 `0x0F`、graphics 寫入模式 1（latched copy），逐列 `movsb`。
- 只動第 10 列起的 180 列（偏移 `0x190 = 400 = 10×40`，計數 `0xB4 = 180`），
  欄的單位是 EGA 的 byte column（8 像素），四個平面一起搬（寫入模式 1 用 latch）。
- **模擬驗證**：以封面為起始畫面、目錄頁為來源，照上述規則跑 36 步，與原版
  `accept-and-save-run1` 的動畫畫格比對——18 張裡有 16 張**逐像素完全相同**；
  frame-008 與 frame-026 只差原版滑鼠游標（50／55 px）；frame-020 是取樣落在
  第 25 步與第 26 步之間的更新中途（前 84 列等於第 25 步、後 116 列等於第 26 步）。
- remake 端 `internal/book/animation.go` 依 `tools/extract_book_animation.py` 抽出的
  typed data 重播，`plan-book-open.json` 18 對全通過。
- **限制**：來源緩衝 `word_1A63C` 的填法尚未反追（實測內容等於目錄頁）；其他畫面之間
  是否也用同一組表未盤點；第二段的 `e` 欄與 `c+d` 不相等的用意未讀出，只照表重播。
