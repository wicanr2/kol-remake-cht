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
