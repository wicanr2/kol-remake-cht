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

## RE-049：角色選項頁、角色清單與 CHARDATA 的槽位版面

- 新增正常 DOS UI 情境 `list-characters`（目錄頁按 5 → 角色選項選單選 LIST CHARACTERS →
  角色清單 → 取消），兩輪的六個 checkpoint 畫面 SHA-256 逐一相同，三個角色檔全程未變。
- **角色選項頁**：標題圖是 `BOOKPICS[8]`，位置 (32,16)；選單在左頁 x=24、y=52、行距 6，
  由上而下是 `LIST CHARACTERS`、`CREATE CHARACTER`、`DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER`，反白項為黑色、其餘青色。
- **角色清單頁**：左標題圖 `BOOKPICS[12]` 在 (32,16)、右標題圖 `BOOKPICS[9]` 在 (176,16)；
  名字在 x=24、所在地在 x=176，兩欄都是 y=52、行距 6、黑色。
- **`CHARDATA` 的槽位版面**（12,288 bytes ＝ 16 槽 × 768 bytes）：名字在槽位起點、
  以 NUL 結尾；所在地是位移 64 起的**固定 10 bytes**（不以 NUL 結尾）；城鎮代碼是位移
  144 起的 3 bytes。畫面上的所在地欄是「10 byte 所在地 ＋ 一個空白 ＋ 3 byte 城鎮代碼」，
  共 14 個字元。這三個位置不是猜的：用畫面讀回的 14 個名字在 `CHARDATA` 內搜尋，
  每一個都唯一命中槽位起點（偏移 0），所在地與城鎮再由畫面逐字比對確認。
  原版磁碟上 16 槽中有 14 槽已使用，另外 2 槽名字為空——與畫面上的 14 列一致。
- **兩段翻頁沿用 RE-048 的同一張參數表**，只是起點與來源不同。逐像素比對逼出兩條原版
  行為：**反白是動畫結束之後才畫上去的**（動畫期間整份選單都是青色），以及**離開選單
  進入清單時，原版先把反白拿掉再開始捲動**（frame-045 還有反白，frame-046 起沒有）。
  兩者都是把反白放錯位置時逐張差 281 px、放對之後差 0 px 得到的。
- 對拍收據：`plan-book-menu.json` 16 對、`plan-book-roster.json` 20 對，全部通過。
- **限制**：槽位其他欄位（種族、職業、屬性、金冠、裝備）的位移尚未反追；
  `HARBOR MASTER` 的功能未探；remake 建立的角色還不會寫回 `CHARDATA`。

## RE-050：角色寫入的逐 byte 重現與武器樣板來源

- 以 `--save-character-files` 重跑 `accept-and-save`，取得原版寫入前後的 `CHARDATA`。
  原版把角色寫進槽位 14，**只改動 58 個 byte**，全部落在槽位的 0x00–0x63 之間。
- 對照畫面上讀回的十個屬性值，槽位 `0x1d`–`0x23` 存的是**內部值**（畫面值＝
  `floor(內部值 × 100 / 150)`），逐一吻合。
- **金冠是 big-endian**：職業字表（dgroup+0x1CC）裡是 little-endian 的 `dc 05`＝1500，
  寫進槽位 `0x2b` 時是 `05 DC`。`KOLBOOK-character-writer-summary.json` 記成
  little-endian，該處為誤；big-endian 與原版讀圖片張數時的 `xchg al, ah` 一致。
- **職業 metadata（dgroup+0x108，39×5）的第一個 byte 是兩個 selector 的 nibble**：
  高 nibble 是副武器 selector、低 nibble 是主武器 selector（BARBARIAN `0x84` →
  HL(8)/GS(4)；RANGER `0x39` → LB(3)/BR(9)）。其後四個 byte 依序是副武器兩個技能值、
  主武器兩個技能值。兩筆記錄都吻合。
- **起始武器樣板來自 `BOOKDATA` 的第 3 個項目**：`_main` 在 `0x10257` 把它載到
  dgroup+0x79EE，`sub_11820` 以 `[bx+79EEh]` 定址。該項目解出 225 bytes ＝ 14 筆
  16-byte 記錄，與 `weapon_selector_map` 的 14 個 selector 一致。
- **武器 selector 到記錄索引的規則已讀出**（`0x12181`–`0x121f1`）：
  `var_12 = meta[i][0] & 0x0F − 1` 是主武器記錄索引；`var_14 = meta[i][0] >> 4`，
  非零時 `(var_14 − 1) × 16` 是副武器記錄的位移，為零則整段跳過——**沒有副武器的職業
  不會寫 `0x57` 與 `0x5f`**。39 筆職業裡有 21 筆沒有副武器。
- 樣板複製的去向：主武器整筆 16 bytes 到 `0x30`、前 8 bytes 到 `0x4a`、第 13 byte 到
  `0x52`；副武器前 8 bytes 到 `0x57`、第 13 byte 到 `0x5f`。
- **一個被推翻的推論，以及它為什麼撐得住那麼久**：只有 BARBARIAN 一筆資料時，把寫進槽位
  的武器 bytes 拿去搜尋，在 `ITEM.LST` 裡唯一命中記錄 7；GS(4)→7、HL(8)→11 兩點都成立，
  看起來就是「`ITEM.LST` 索引 ＝ selector ＋ 3」。跑第二個職業才發現 BR(9)→5、LB(3)→25
  不符。真正的來源根本不是 `ITEM.LST`——同樣的武器記錄同時出現在 `ITEM.LST` 與
  `BOOKDATA` 第 3 項，所以位元組搜尋會命中一個**正確但無關**的位置。
  **位元組相同不等於來源相同**；要確定來源只能讀定址的那段程式碼。
- **對拍結果**：`internal/chardata` 的 writer 對 BARBARIAN 與 RANGER 各 58/58 個 byte
  逐一相同，寫錯 0、多寫 0（`docs/verification/chardata-write/`）。兩筆是不同職業、
  不同武器、不同擲骰結果。
- **限制**：REGION 中繼資料的來源（`region_metadata_copy_map`）在本例中與槽位樣板相同，
  因此無法分辨原版是否真的寫過；武器記錄 16 個 byte 的內部欄位語意未知；
  39 筆職業都能寫出樣板，但只有 BARBARIAN 與 RANGER 有原版寫入可對照。

## RE-051：DELETE CHARACTER 的畫面、確認視窗與一個 byte 的刪除

- 新增三個正常 DOS UI 情境：`delete-character-probe`（進刪除畫面後直接取消）、
  `delete-character`（反白條上下移動、叫出確認視窗）、`delete-character-confirm`
  （實際回答 YES 刪除，再回答 NO 驗證取消分支）。後兩個各跑兩輪，13/13 與 15/15 個
  checkpoint 畫面 SHA-256 逐一相同——包含 TITLE 標題序列那一張。
- **刪除畫面就是角色清單頁**，標題圖、兩欄起點與行距都和 `LIST CHARACTERS` 相同，
  差別只有顏色語意：反白那一列是黑字、其餘青字，左右兩欄同時反白。反白條起點是第 0 列，
  上下移動不繞圈（送 Up 到第 0 列、Down 到末列都停住）。
- **確認視窗**：黑框白底 132×24 位於 (94,58)，投影是往右下各偏移 4 px、厚 4 px 的黑塊。
  兩行暗紅（EGA 170,0,0）文字 `ARE YOU SURE?` 在 (104,64)、`YES/NO` 在 (128,70)，
  由 `tools/read_text_from_frame.py` 以抽出的原版字型逐格讀回，不是人眼抄的。
  兩輪執行、清單長度不同（14 列與 13 列）時視窗位置完全相同。
- **視窗只認字母鍵**：送出 Down 與 Return 之後畫面逐像素沒有任何變化（`delete-character`
  的 checkpoint 10–12 是同一張 PNG）；`y` 才會刪除，`n` 取消。視窗開著時清單沒有任何
  一列反白。兩種回答都回到刪除清單，且反白條重置為第 0 列。
- **刪除只寫一個 byte**：把該槽位名字的第一個 byte 寫成 0。兩輪的 12288 bytes `CHARDATA`
  差分完全相同——只有 slot 1 的 offset `0x000` 由 `0x50` 變成 `0x00`，`CHARPICS` 與
  `SHLDPICS` 的 SHA-256 沒變。畫面上名冊往上補齊是「只列出已建立的槽位」造成的，
  不是資料搬移。
- **對拍結果**：`plan-delete.json` 8 對、`plan-delete-move.json` 5 對全部通過，允許差異
  只有原版滑鼠游標（55 px）與展開跨頁的頁尾殘影（112 px）。remake 的 `DeleteSlot`
  對同一份 `CHARDATA` 產出與原版**逐 byte 相同**的檔案
  （`docs/verification/chardata-delete/`）。
- **限制**：確認視窗按 ESC 的行為沒有量測；`Y`／`N` 大寫、滑鼠按鍵與「名冊為空時進入
  刪除畫面」都未測；`EDIT PICTURE`、`EDIT SHIELD`、`HARBOR MASTER` 仍未探。

## RE-052：書本剩下三項功能的入口——共用的角色選擇清單

- 新增六個正常 DOS UI 情境：`edit-picture-probe`／`edit-shield-probe`／
  `harbor-master-probe`（進入即取消）與對應的 `-enter`（選第一名角色後進入功能本身）。
  三個 `-enter` 情境都保存了角色檔，`CHARDATA`／`CHARPICS`／`SHLDPICS` 的 SHA-256
  全程未變。
- **四個功能共用同一張角色選擇清單**：`DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER` 進入後的第一張畫面是**同一個 PNG**
  （SHA-256 `36eba7c1…`，四份收據完全相同）——角色清單頁加一條反白條，起點第 0 列。
  差別只在選定角色之後走到哪裡。
- **角色選項選單的 ESC 回到目錄頁**：由清單 ESC 回到選單（反白重置為第 0 項），
  再 ESC 回到目錄頁。三個 probe 情境都走出同一條路徑。
- **HARBOR MASTER 在只安裝一個 region 時是一張訊息頁**：展開跨頁上三行黑字
  `ONLY ONE REGION`(24,52)、`IS INSTALLED.`(24,58)、`PRESS ANY KEY.`(24,70)，
  第三行前空一列，沒有標題圖。按鍵回到角色選項。多個 region 的畫面沒有量測——
  那需要另一張角色磁碟。
- **`EDIT PICTURE` 與 `EDIT SHIELD` 是同一個點陣圖編輯器**，只是編輯對象不同
  （`CHARPICS` 對 `SHLDPICS`）。畫面由左側大編輯區、右上預覽與目前顏色、右側 16 色
  色盤、底部 8 個工具圖示組成，全部畫在紅磚底上而不是書頁上。**ESC 沒有作用**——
  連送三次畫面逐像素完全沒變，因此離開方式與工具操作都要另查參考卡，而且原版是
  滑鼠驅動的，現有的鍵盤 oracle 走不完。
- **對拍結果**：`plan-harbor.json` 1 對通過，允許差異只有滑鼠游標 55 px 與頁尾殘影
  112 px。編輯器沒有實作，因此沒有對拍。
- **限制**：編輯器的工具語意、色盤索引、寫回 `CHARPICS`／`SHLDPICS` 的時機與格式都未知；
  要往下做需要能送滑鼠事件的 oracle。

## RE-053：走出書本——Play the game 到旅店組隊

- 新增四個正常 DOS UI 情境：`play-the-game`、`town-icon-bar`、`town-icon-move`、
  `town-recruit-companion`。這是本專案第一次把 oracle 開出 `KOLBOOK` 之外，
  進入 `KOLTOWN` 的城鎮畫面。三個角色檔全程未變。
- **進入遊戲的鏈**：目錄頁按 `6`（Play the game）→ 共用的角色選擇清單（**和
  `DELETE CHARACTER` 等四個功能是同一張畫面**，SHA-256 相同）→ 選第一名角色當領隊 →
  載入城鎮，停在半途之家旅店（`Halfway House`，店主 `Devin Torlock`）。
  畫面左上是隊員頭像與名字前四個字母（手冊說的「只顯示前四個字母」得到證實），
  左側是店主肖像，中央是對話框，底部是圖示列。
- **圖示列不吃方向鍵**：在旅店送 Right／Left 之後畫面逐像素完全沒變。改用參考卡指定的
  `<`／`>` 才會移動白框——`>` 往右、`<` 往左，走到第 6 格再按 `>` 會**繞回第 1 格**。
  Return 執行白框所在的圖示。旅店的六個圖示依序是 DOOR、MIRROR、EAR、MOUTH、BED、
  COMPANION，和手冊列出的順序一致。
- **送 `<`／`>` 需要送鍵層自己補 Shift**：US 排列上它們是逗號／句號的 Shift 位，
  只查 keysym→keycode 會送成 `,`／`.`。`tools/x11_key_inject.py` 因此改成從
  keyboard mapping 的奇數位置判斷哪些 keysym 需要 Shift。
- **組隊流程**：COMPANION（第 6 格）→ 先問 `Look for characters on hard disk?`
  （兩個圖示，白框在第一個）→ Return → 列出**該旅店裡不在隊伍中的角色**
  （半途之家有 8 人，與 `CHARDATA` 中所在地為 `HALFWAY HO` 的 11 人扣掉領隊後一致）→
  上下方向鍵移動反白、Return 加入。加入後隊員頭像出現在畫面頂端領隊旁邊，
  回到旅店的一般圖示列。
- **招募不寫磁碟**：整個組隊流程跑完，`CHARDATA`／`CHARPICS`／`SHLDPICS` 的 SHA-256
  都沒有變。這與手冊一致——遊戲只在旅店選 `REST/SAVE/HEAL` 時才保存。
- **DOOR 直接離開旅店**：Return 在第 1 格會走到布雷特爾的戶外地圖（日期框、城鎮徽章、
  俯視地圖），也就是手冊印刷頁 15 描述的畫面。
- **限制**：圖示列的其餘五個圖示、戶外地圖的移動、隊伍上限六人、發言者切換（白框頭像）
  與 `REST/SAVE/HEAL` 的寫入都還沒量測；remake 也還沒有任何城鎮畫面。

## RE-054：旅店畫面的組成與圖示列的素材

- 新增 `tools/decompose_town_frame.py`：用抽出的全螢幕素材逐張與畫面相減取最小差異當
  底圖，再對差異像素做四鄰連通分群，並要求**分群之外的殘餘差異為零**。這個檢查是判斷
  底圖對不對的關鍵——只看「差異最小」會接受一張只是碰巧接近的圖。
- **旅店畫面的底圖是 320×200 的紅磚畫面**，`BOOKPICS[3]` 與 `KOLMPICS[0]` 是同一張圖
  （PNG SHA-256 `83e871eb…fe73` 相同），書本與城鎮各存一份。殘餘差異 0 像素，因此
  「底圖 ＋ 11 個矩形」就是這張畫面的完整組成。
- **圖示列已完全對上**：六格在 x = 32、64、96、128、160、192，y = 170，外框 24×22，
  框內 22×20 是 `KOLMPICS` 圖示的第 1..22 欄。六格逐像素零差異命中
  `KOLMPICS[5]`（DOOR）、`[20]`（MIRROR）、`[7]`（EAR）、`[6]`（MOUTH）、
  `[8]`（BED）、`[9]`（COMPANION），順序與手冊列出的一致。
  **注意：整張 24×20 圖示在畫面上會被外框蓋掉左右各一欄**，直接拿 24×20 去全畫面搜尋
  會零命中——這也是第一次搜尋失敗的原因。
- 其餘四塊的矩形：對話框 (102,84) 199×72、店主肖像 (8,64) 92×68、標題框 (102,58)
  199×24、隊員頭像 (14,1) 47×52（`CHARPICS[47]` 畫在 (16,2)，下方是名字前五個字元）。
- **店主肖像是 `TWN7PICS[12]`（88×64）畫在 (8,64)**，88×64 個像素全部相同，外框是素材
  自帶的，畫面上只多一層投影。
  找法值得記下來：先用「尺寸範圍 ＋ 位移搜尋」找了三輪都零命中，改成**從畫面上取一列
  60 個像素當指紋，去所有素材的像素緩衝區裡 `find`**，一次就命中。
  **不確定素材放在哪裡時，用畫面上的一列去搜尋素材，比拿素材去搜尋畫面便宜得多**——
  前者是一次字串搜尋，後者是尺寸猜測乘上位移猜測。
- **標題框與對話框**：兩個框都是白框黑底，右下同樣是偏移 4、厚 4 的投影。標題框
  (102,58) 195×20、對話框 (102,84) 195×68。**框內文字用 8×8 大字型**（書頁內文用的是
  8×6 小字型），字格從 x=104 起、行距 8、每行 24 格；標題兩行置中（起始格 ＝
  (24 − 字數) / 2 向下取整），台詞靠左。隊員名牌則用 8×6 小字型，只畫名字前五個字元。
- **隊員格**：外框 43×40（上格圖像）＋ 43×9（下格名牌）共用邊，整塊先填黑再畫
  `CHARPICS` 圖像與名字，右下同樣有偏移 4 的投影。**白框是發言者、亮藍框是其餘隊員**；
  招募之後新加入的成員成為發言者（frame-081 中第 1 格轉藍、第 2 格轉白）。
- **限制**：戶外地圖、商店、修道院尚未拆解；圖示列其餘五個圖示的行為未測。

## RE-055：旅店畫面的 remake 對拍

- `internal/town` 保存量得的版面與旅店狀態，`internal/ui` 依同一組數字重繪，
  `cmd/kol -scene inn` 輸出截圖。對拍計畫 `plan-town-inn.json` 八對：六個圖示白框位置、
  單人隊伍與招募後的兩人隊伍，**允許差異區之外全部零像素**，允許區只有原版滑鼠游標
  （每張 28–31 px）。
- 兩個原版收據不同（`town-icon-move-run1` 與 `town-recruit-companion-run1`），
  因此不是同一張畫面對自己。
- 中文模式一併補上（`-lang zh-Hant -scene inn`）：素材與版面兩倍放大，文字改用 16×18
  中文字格，台詞是預先斷行的暫譯。中文模式不對拍。
- **烘字模表的來源不只字串表**：城鎮樣本自帶的中文台詞不在 `docs/i18n/` 裡，
  第一次烘出來的表沒有那些字，畫面上會變成空白。`tools/build_cjk_font.py` 因此加了
  `--extra-json`，掃描任意巢狀 JSON 的字串值收集非 ASCII 字元。

## RE-056：城鎮戶外地圖的拆解——捲動的 16×16 磚塊圖

- 新增情境 `town-outdoor-move`：在旅店按 DOOR 走到戶外地圖，再用方向鍵移動幾步。
  三個角色檔全程未變。
- **戶外畫面的組成**（`tools/decompose_town_frame.py`，底圖同樣是那張紅磚圖）：
  地圖視窗 (103,57) 194×98（白框 1 px，框內 192×96 從 (104,58) 起）、城鎮徽章
  (24,85) 56×57、日期框 (14,57) 67×15、隊員頭像格 (14,1) 47×52、單一圖示格
  (32,170) 24×22。三個框都帶偏移 4、厚 4 的黑色投影，和旅店同一套。
- **徽章是 `TWN7PICS[0]`**（56×57，含 `KOLMPICS[28]` 共用 AND-mask）畫在 (24,85)，
  2,754 個不透明像素零差異。圖示是 `KOLMPICS[5]`（DOOR），零差異。
  日期框內是小字型白字：`Apr 1` 在 (24,59)、`AD  1100` 在 (16,65)。
- **地圖視窗不是一張固定圖，是捲動的磚塊圖**。三條證據：
  1. 從地圖上取八段 24 像素的指紋去 752 張 PNG 搜尋，**全部零命中**；同尺寸 192×96
     素材逐張對位比較，最好的一張也差 12,535／18,432 個像素。
  2. 移動時視窗以 **16 px** 為單位捲動：往右一步視窗不動（只有隊伍圖示動，98.7%
     相同），往右兩步水平位移 16 px 後 97.9% 相同，往下兩步垂直位移 16 px 後 97.7%
     相同。
  3. 以 16×16 切格時，位移 (0,0) 的 72 格只有 47 種不同內容（65%），其餘 255 種對齊
     都在 87% 以上——磚格對齊視窗原點。
  合起來：視窗是 12×6 格、每格 16×16 的捲動地圖。
- **磚圖來源仍未知**（假說）：`KOLPICS` 的 0／8／12 是 `sub_19B29` 轉換的 raw-planar
  資料，RE-030 已證實**不是** display rectangle，因此拿它們的 PNG 像素去搜尋本來就搜
  不到——這是磚圖來源的第一候選，但要讀 KOLTOWN 的地圖繪製路徑才能確定。
- **remake 還不能畫這張畫面**：底圖、徽章、日期框、頭像格與圖示格都對得上，但地圖
  視窗佔 18,432 個像素，沒有磚圖就畫不出來。把它列成「允許差異區」等於用允許掩蓋
  未實作，違反對拍判準，因此這一輪不實作、不對拍。
- **限制**：磚圖與地圖配置的格式、捲動的邊界規則、隊伍圖示的動畫都還沒查。

## RE-057：城鎮地圖的磚圖在 `TOWNn.0` 裡，格式是 16×16、四平面、plane-major

- **`TOWN0.0`…`TOWN7.0` 從來沒有被分析過**。既有的清冊工具 `is_candidate()` 只認名字
  含 `PIC` 的檔案，七個 50–65 KB 的城鎮資料檔因此一直不在 inventory 裡——
  這是**自己的過濾器有洞**，不是資料不存在。
- `TOWN7.0`（＋分卷 `TOWN7.1`）用 `*PICS` 同一套容器契約解得開：26 個 payload，
  LZW 解碼後的長度分佈是 3 × 12,289、21 × 2,049、1 × 513、1 × 116
  （尾端多一個 byte 是既有的 LZW 終止 byte）。
- **磚圖在 payload 0／1／2**：這三段解碼後首尾相接成一塊 36,864 byte 的區域
  （段內位移 0x0000／0x3000／0x6000），其中**前 4,096 byte 是地圖格子陣列**、
  其後 32,768 byte 是 **256 塊磚**（RE-059 由繪製迴圈定案）。把原版戶外畫面上
  72 個 16×16 磚塊轉成候選格式去搜尋，**67 塊命中**，而且：
  - 全部命中 `planar-plane-major`：16×16 磚 ＝ 4 個 EGA 平面，每平面 32 bytes
    （16 列 × 2 bytes），四個平面依序排列共 128 bytes。
    另外三種候選（high／low nibble 封裝、planar row-major）一塊都沒中。
  - **全部的 offset 都是 128 的倍數**，磚距乾淨。
  - 沒命中的五塊是畫面上 (2,3)–(3,4) 那一小片，正是隊伍圖示與其覆蓋處。
- **當時的編號假設是錯的**：本輪把三個 payload 當成各自 96 塊、合計 288 塊，於是推論
  「一個 byte 放不下」。實際的磚塊區從段內位移 0x1000 起算、共 256 塊，
  一格確實就是一個 byte（RE-059）。當時的六種指紋搜尋也全部只掃 2,048 byte 的
  payload，沒有掃 payload 0 的開頭——格子陣列正好在那裡。
- **限制**：21 個 2,049 byte payload 與 116 byte 表的用途未知；磚圖的色盤語意、
  動畫磚、以及其他六個城鎮是否共用同一格式都還沒查。

## RE-058：城鎮磚圖的載入路徑，以及兩個被推翻的假設

新增探索用探針 `tools/ida_probe_koltown_town_map.py`（安全摘要見
`ida-fingerprints/KOLTOWN-town-map-summary.json`）。輸入是本機的 unpacked
`KOLTOWN.EXE`（SHA-256 `01f6f9…d1d2`），338 個函式。

- **只看 xref 會得到假零**：`town%d.0`、`twn%d%s.0`、`twn%d%s.1`、`town%d` 這些字串的
  `DataRefsTo` 全部是空的。16-bit real mode 下程式碼是把 DGROUP 位移當**立即值**載入
  （`mov ax, 1776h`），IDA 不會因此產生 data xref。改掃立即值之後，18 個使用點一次浮現。
- **`sub_15920` 是城鎮資料載入器**：依目前城鎮編號（`word_2DF2C`）組出檔名，
  用 `sub_19D98` 把 `town%d.0`／`twn%d%s.0`／`twn%d%s.1` 讀進記憶體，
  然後用 `sub_1A4FE` 從 `town%d` 這個 archive 取出**payload 0、1、2**，
  分別放到同一個 segment（`word_30C8A`）的位移 **0x0000、0x3000、0x6000**，
  最後對 `word_30C8A+0x100` 呼叫 `sub_19B29`（RE-030 已證實的 raw-planar → 色盤索引轉換）。
  - 0x3000 ＝ 12,288，正是每個磚圖 payload 的解碼長度，**三個 payload 首尾相接**。
  - 這與 RE-057 的位元組證據互相印證：磚塊是一個平坦陣列，第 *n* 塊的位移是
    `n × 128`。陣列的起點與筆數由 RE-059 定案為位移 0x1000、共 256 塊。
- **`sub_1C80F` 是 nibble RLE**：版面是 `[0..2)` big-endian 的 N、`[2..N)` 控制位元組、
  `[N..)` nibble 資料流；控制位元組最高位為 0 時 `(值+1)` 是重複次數，為 1 時
  `(值−0x7f)` 是照抄的 nibble 數，輸出兩兩打包成 byte。`sub_1C907` 依 `word_2DF1E`
  在它與 `sub_1A2BB` 之間分流。

### 兩個被推翻的假設

1. **「`sub_2A0C6` 載入的是 `TOWNn.0` 的地圖區段」是錯的。** 把 sprintf 的格式字串
   立即值解回文字之後，`sub_2A01C`（0x1874）與 `sub_2A0C6`（0x187C）用的都是
   **`weat%s`**——那是天氣圖 `WEATPICS`，與城鎮地圖無關。它一次取 `3n`、`3n+1`，
   在 `n` 為 9 或 12 時另取 `3n+2`。
2. **「21 個 2,049 byte 的 `TOWN7.0` payload 是 nibble-RLE 的地圖區段」也是錯的。**
   把 `sub_1C80F` 的版面實作出來套上去，每一個 payload 的 big-endian 表頭都指向
   2,132–16,589，全部超出 payload 長度，一個都展不開。它們是別的東西。

### 六種資料側問法全部零命中（成因見 RE-059）

格子陣列後來由繪製迴圈解出（RE-059）：它在 `TOWN7.0` payload 0 的**開頭 4,096 byte**，
而下表六種問法全部只掃 21 個 2,049 byte 的 payload，或是套用了「288 塊磚」這個錯的
筆數。搜尋範圍與編號兩邊同時偏掉，才會六種問法一起回零。
（**寫下來是為了不要有人再問一次同一個問題**）：

| 問法 | 範圍 | 結果 |
|---|---|---|
| 一格一 byte 的磚塊編號 | `TOWN7.0` 全部 payload | 零命中 |
| 一格兩 byte（LE／BE）的磚塊編號 | 同上 | 零命中 |
| 只看「哪幾格相同」的相等關係 | 21 個 2,049 byte payload × 2 種格寬 × 8 種列距 × 全部起點 | 零命中 |
| 偏移不變的差分（編號整體加常數） | `TOWN7.0` ＋ `TWN7PICS` 全部 payload，byte／word LE／BE | 零命中 |
| `sub_1C80F` 的 nibble RLE | 21 個 2,049 byte payload | 表頭全部越界，展不開 |
| 掃「乘以 128」的算式找繪製迴圈 | 全部 338 個函式 | 只有 `mov cx,7` 的角色資料 stride，沒有地圖繪製 |

`word_30C8A` 也不是磚塊專用緩衝區：`sub_16986` 先把它當繪圖目標（`word_2DF12`），
之後才換成 `0A000h`（EGA 視訊段），所以它是**離屏畫面緩衝區**，城鎮的地圖與磚圖
載入時借用它的前 36 KB。

## RE-059：城鎮地圖格子陣列——`sub_115D1` 的繪製迴圈與 64×64 的一格一 byte

輸入是本機的 unpacked `KOLTOWN.EXE`（SHA-256 `01f6f9…d1d2`），IDA Pro 9.4
`ida-pro-9.4-idapython:locked-v1`，探針 `tools/ida_probe_koltown_town_map.py`；
資料側驗證用 `tools/probe_town_map_grid.py`，收據 `docs/re/town-map-grid.json`。

### 說明書先說了功能，反組譯再說怎麼做

`docs/manual/chapters/02-poking-around-town.md`：從旅店 DOOR 出來後，左邊窗格是日期與
城鎮徽章，大窗格是戶外地圖；隊伍是一頂會左右張望的頭盔，用方向鍵移動（參考卡另給
`A`／`S`／`Z`／`X` 斜走）。有路通到建築就進得去；站在門口時地圖會換成建築物的圖，
並出現 DOOR 與 U-TURN 圖示。走到城鎮邊緣會看到野外，此時 DOOR 切換到野外地圖。
`ROAD` 與 `SIGNPOST` 兩個圖示只在路上有作用。

### `sub_115D1`：12×6 的雙重迴圈（`已證實`）

```
sub_1C40D(0x68, 0x3A, 0x128, 0x9A)            ; 清出 (104,58)-(296,154) 的地圖視窗
for row = 0..5:                                ; 0x116c5 cmp var_2, 6
  for col = 0..11:                             ; 0x116b8 cmp var_4, 0Ch
    X = word_2FB50 + col + 0FFFBh              ; 鏡頭 X，欄位偏移 -5
    Y = word_2FB56 + row + 0FFFEh              ; 鏡頭 Y，列偏移 -2
    X = clamp(X, word_2DF68, word_2DF6A)       ; 0x1161a / 0x1163c
    Y = clamp(Y, word_2DF66, word_2DF6C)       ; 0x1162b / 0x1164d
    word_3AC5A = (X << 6) + Y                  ; 0x1165e mov cx,6 / shl / add
    tile = byte ptr word_30C8A:[word_3AC5A]    ; 0x11674 mov al, es:[si]
    byte ptr [2A48h + 12*row + col] = tile     ; 0x11690 畫面上 12×6 的磚塊快取
    sub_1CAC2(tile, 0x280*row + 91Dh + 2*col, word_30C8A)
```

- 目的位移 `0x280*row + 0x91D + 2*col` 是 320×200 EGA 平面的位址：每列 40 byte，
  `0x91D = 58*40 + 13` ＝ 螢幕 (104, 58)；`0x280 = 640 = 16 列 × 40`；一格橫向 2 byte
  ＝ 16 像素。與 RE-056 量到的視窗位置、格數完全一致。
- `word_2DF68`（X 下限）在 `0x109c4`／`0x10bfa` 靜態就是 `0`，
  `word_2DF6C`（Y 上限）在 `0x104e3`／`0x109d0`／`0x10c06` 靜態就是 `3Fh`。
  **Y 只有 64 個值**，這是格子陣列版面的關鍵。

### `sub_1CAC2` → `sub_1D1C6`：磚塊來源在段內位移 0x1000（`已證實`）

`sub_1CAC2` 只是依 `word_2DF1E` 分流的顯示模式派送：`0Dh` → `sub_1D1C6`（EGA）、
`9` → `sub_1D994`、`< 8` → `sub_1E211`（CGA）。EGA 那支的核心是：

```
0x1d1d8 mov cx, 7 / shl ax, cl        ; 來源位移 = tile * 128
0x1d1f2 mov cx, [bp+arg_4] / add cx, 100h / mov ds, cx   ; 來源段 = 緩衝段 + 0x100 paragraph
0x1d206 out dx, ax                    ; 3C4h sequencer map mask，ax 由 0102h 逐平面左移
0x1d207 movsb / movsb / add di, 26h   ; 每列 2 byte，跳到下一列（2 + 38 = 40）
```

`+0x100` paragraph ＝ **+0x1000 byte**。所以離屏段的前 0x1000 byte 不是磚圖：

| 段內位移 | 大小 | 內容 |
|---|---|---|
| `0x0000`–`0x0FFF` | 4,096 | 地圖格子陣列，索引 `(X << 6) + Y`，一格一 byte |
| `0x1000`–`0x8FFF` | 32,768 | 256 塊磚，每塊 128 byte，四平面 plane-major |

`sub_15920` 把 `TOWNn.0` 的 payload 0／1／2 放到 `0x0000`／`0x3000`／`0x6000`，
最後對 `word_30C8A + 0x100` 呼叫 `sub_19B29`，長度剛好是 `0x8000` ＝ 256 × 128——
**它處理的範圍與磚塊區一個 byte 不差**，格子陣列被排除在外。

### 資料側驗證：67 格全中，鏡頭唯一解（`已證實`）

把 `TOWN7.0` 的三個 payload 解碼後接成 0x9000 的緩衝區，取前 4,096 byte 當格子陣列，
磚塊編號改用 `(payload*0x3000 + payload 內位移 − 0x1000) / 128`，再用 RE-057 在原版
畫面上量到的 67 格當指紋，掃過 64×64 個鏡頭位置：

- **唯一命中 `(word_2FB50, word_2FB56) = (57, 8)`，67/67 全中**；次佳只有 12 格，
  沒有第二個接近的候選。
- 首輪比對有 7 格看起來不符（畫面 193、地圖 210），追下去發現**磚 193 與磚 210 的
  128 byte 完全相同**，比對器只是回報編號較小的那一個。整組磚有 3 群位元組相同的
  重複磚（最大一群 15 塊），共 19 塊、唯一磚 240 塊。
- 磚塊編號範圍 30–254，全部落在一個 byte 內；重新編號之後 payload 0 沒有任何一塊
  磚落在 `< 0x1000`，與版面表一致。
- 把陣列印成 64×64 的字元圖，可以直接看出建築的矩形、道路、以及城鎮外圍成片的
  空地磚（值 254，出現 1,680 次）。

### 這一輪修正了什麼

| 先前斷言 | 出處 | 現況 |
|---|---|---|
| `sub_2A0C6` 讀 `twn` archive | RE-058 | 格式字串是 `weat%s`，讀的是天氣圖 |
| 磚塊共 288 塊，一格放不進一個 byte | RE-057 | 256 塊，一格就是一個 byte |
| 格子陣列不在已知 payload 裡 | RE-058 | 在 payload 0 的開頭 4,096 byte |

### 還沒查的

其他六個城鎮是否共用同一版面、`word_2DF6A`（X 上限）在各城鎮的值、21 個 2,049 byte
payload 的用途、隊伍頭盔的動畫與覆蓋規則、以及走到邊緣切換野外地圖的條件。

## RE-060：隊伍標記的精靈、動畫與那圈算出來的黑色外框

### 精靈來源：KOLPICS payload 0 的十個 128 byte 單元（`已證實`）

`sub_10000:0x105d4` 用 `sub_1A4FE` 把 `kol%s` archive 的 payload 0 讀進段 `word_39824`
的位移 0，接著 `sub_19B29(word_39824, 0x500)`——**0x500 ＝ 1,280 byte ＝ 10 × 128**，
與城鎮磚塊同一種 16×16、四平面 plane-major 的格式。

`sub_11390` 選其中一個：索引是 `word_2DF48 + word_2DF24`。城鎮戶外畫面走
`word_2DF48 == 0` 那條分支（`0x113c4`），所以用的是單元 0–4；`word_2DF24` 在
`0x114cc`–`0x114f3` 由 `word_2DF26` 推進，碰到 4 就改 −1、碰到 0 就改 +1，
是**五格來回**的動畫，正好對上手冊寫的「頭盔會左右張望」。

既有的 `KOLPICS` 抽取把同一段當 renderer 資料排成 256×5，
`decode_planar` 每 256 byte 只取前 128 byte，因此看不到單元 1、3、5、7、9——
畫面上那一格用的正是單元 1。新增 `tools/probe_extract_party_marker_png.py`
把十個單元排成 16×160 的 atlas。

### 遮罩不是存的，是算出來的（`已證實`）

`sub_1D21D` 是 EGA 的合成路徑：

```
0x1d24d  ds:2A9Ah[0..15] = 0
0x1d25d  for plane in 1..4: for r in 0..15: 2A9Ah[r] |= sprite_word         ; 佔用位圖
0x1d279  for r: ax = w[r]; ax |= (ax << 1) | (ax >> 1); w[r] = ~ax           ; 膨脹後反相
0x1d29e  for plane, r: [2A9Ah+20h] = mask[r] & tile_word                     ; AND 背景
0x1d2c9  for 40h words: [var_2] = [2A9Ah+20h] | sprite_word                  ; OR 精靈
0x1d2e2  sub_1D1C6(0Ah, dest, seg)                                           ; 當一般磚塊畫出去
```

也就是「把精靈四個平面 OR 成佔用位圖 → 左右各膨脹一像素 → 反相當 AND 遮罩 →
`(磚塊 AND 遮罩) OR 精靈`」。膨脹範圍內、精靈自己沒用到的像素會變成黑色，
這就是標記四周那一圈外框。合成結果先寫進精靈段的第 `0x0A` 塊暫存磚，再照一般磚塊畫。

**位移是在 little-endian 的 16-bit word 上做的**，而畫面上左半 8 像素是低位元組——
所以 `<<1`／`>>1` 不會跨越位元組邊界，第 7 欄與第 8 欄之間不互相膨脹。
先照「一整列 16 像素」實作時，原版畫面上剛好有一個像素對不上（255/256）；
補上這條邊界規則之後是 256/256。

### 資料側驗證：從原版畫面反解狀態（`已證實`）

`tools/solve_outdoor_camera.py` 把畫面上 12×6 格的像素對回磚塊編號，
再用格子陣列解鏡頭，最後逐格試合成解出標記所在與動畫畫格。對
`town-outdoor-move-run1` 的 94 張畫格：

- 22 張是戶外畫面，**每一張的鏡頭都是唯一解**；其餘 72 張不是戶外畫面。
- 標記不一定在正中格：`word_2DF36`／`word_2DF38` 是它在視窗內的欄／列偏移。
  這一輪量到的序列是鏡頭 (57,8) 偏移 (0,0) → 偏移 (0,1) → 鏡頭 (58,8) 偏移 (0,1)
  → (1,1) → 鏡頭 (58,9) 偏移 (1,1) → (0,1)。
  **隊伍先在視窗內走，走到一定位置鏡頭才跟著捲一格**，兩者要分開記。
- 動畫畫格 0、1、2、3、4 全部出現過，與來回規則一致。
- 圖示列是看情境出現的：站在門口的畫格有 `KOLMPICS[5]`（DOOR），
  走開之後整格是底圖紅磚，連外框都沒有。

收據 `docs/verification/town-ui/outdoor-camera-solve.json`。

### remake 對拍

`docs/verification/remake-parity/plan-town-outdoor.json`：21 對逐畫面對拍全過，
允許區只有原版的滑鼠游標。排除兩張原版正在重繪的畫格（隊員格／徽章／日期框還沒畫完）。

### 還沒查的

鏡頭跟隨的確切條件（視窗內走到第幾格才捲動）、斜走鍵、`word_2DF48 == 5` 那組標記
用在哪個畫面、以及 `[bx+6AE4h]` 這個逐格旗標把背景換成磚塊 100 的時機。

## RE-061：城鎮的門座標表、建築配對，以及門口狀態

### 門的判定（`已證實`）

`sub_10000:0x106AB` 把 `TOWNn.0` 的 payload 3 讀進 DGROUP 位移 `4430h`
（IDA 線性位址 `0x32300` ＝ `DGROUP 0x2DED0 + 0x4430`）。`0x10A6D` 起的迴圈說明版面：

```
word_39A4E = word_2FB50 + word_2DF36          ; 隊伍地圖 X ＝ 鏡頭 X ＋ 標記欄偏移
word_39A54 = word_2FB56 + word_2DF38          ; 隊伍地圖 Y ＝ 鏡頭 Y ＋ 標記列偏移
for idx = 0, 2, 4 … while idx <= byte_32320 * 2:
    if [4430h + idx + 47h] == 隊伍X and [4430h + idx + 48h] == 隊伍Y:
        word_2FBC2 = idx                      ; ＝ 門序號 × 2
        sub_1198E()
```

| payload 3 位移 | 內容 |
|---|---|
| `0x20` | 門數（迴圈是 `<=`，實際讀到門數 ＋ 1 筆） |
| `0x27 + 2k` | 第 k 個城門在**世界地圖**上的 X／Y，由 `word_2DF62` 選（TOWN7 有 10 個，其餘 1 個）|
| `0x47 + 2i` | 第 i 個門的 X／Y |

**掃門迴圈接在移動後面**，所以「站在門的格子上」不等於「門口狀態」——
剛從建築物走出來時隊伍就在門的座標上，畫面仍是地圖。這一點由原版畫格證實：
`town-outdoor-move-run1` 的 068–076 站在 (57,8)（旅店門）卻顯示地圖與單一 DOOR 圖示。

### 建築的四份資料是同一個索引推出來的（`已證實`）

`sub_1198E` 對門口狀態做的事：

```
sub_11302(word_2FBC2 + 3)      ; 建築門面圖
word_2DF3C = 0                 ; 清空圖示列
sub_11703(2)                   ; DOOR
sub_11703(0Fh)                 ; U-TURN
var_2 = sub_11722(3)           ; 等玩家選；選 DOOR → sub_11A93() 進去
```

`sub_11302` 依索引奇偶決定貼圖矩形：奇數是 `(13, 58, 37, 154)` ＝ 螢幕
(104,58)–(296,154)，正好是地圖視窗；偶數是 `(1, 40h, 0Ch, 80h)` ＝ (8,64)–(96,128)，
正好是旅店老闆肖像的位置。`sub_11722` 畫圖示時是 `KOLMPICS[id + 3]`，
所以 DOOR（id 2）是 `KOLMPICS[5]`、U-TURN（id 15）是 `KOLMPICS[18]`。

因此門 *i* 對應：

| 資料 | 來源 |
|---|---|
| 門面圖 192×96 | `TWN{n}PICS[2i + 3]` |
| 人物肖像 88×64 | `TWN{n}PICS[2i + 2]` |
| 建築名與人物名 | `TOWNn.0` 的 payload `i + 4`（索引 ≥ `0x14` 再加 1） |

文字是「長度前綴 ＋ 末字元設高位元」：payload 9 的位移 0 是 `0D "Halfway Hous" E5`
（`E5` ＝ `'e' | 0x80`），位移 29 是 `Devin Torlock`——與旅店畫面標題框的兩行相同。

### 跨城交叉檢查

八座城的門數與 `TWN{n}PICS` 的偶奇配對數逐一相等：TOWN1 16／16、TOWN2 14／14、
TOWN3 12／12、TOWN4 12／12、TOWN5 14／14、TOWN6 12／12、TOWN7 22／22。
TOWN0 是野外地圖，門數 0。

**一個原版資料缺口**：TOWN7 的門 21（20, 36）有門面圖 45 與肖像 44，
但它的文字 payload 會是 26，而 `TOWN7.0` 只有 0–25。其餘 21 個門的文字全部對得上。

### 原版觀察與對拍

新增 oracle 情境 `town-enter-building`：離開門口再走回來。畫格 071 證實門口狀態——
地圖視窗換成 `TWN7PICS[13]`（除滑鼠游標外逐像素相同）、圖示列變成
`KOLMPICS[5]` 白框與 `KOLMPICS[18]` 藍框兩格。remake 的門口狀態對這一張畫格
逐畫面對拍通過，允許區只有滑鼠游標。

### 還沒解出來的：建築內文

畫格 072 顯示按下 DOOR 之後是**旅店那張版面**：肖像在 (8,64)、標題框兩行是
payload 9 的兩個字串、對話框是一段敘述文。那段敘述文**不以明文存在任何原始檔裡**——
`Once inside`、`fortress`、`counter`、`Presently` 等字在 75 個原始檔中全部零命中，
連 `TOWNn.0` 的 payload 也沒有。payload 9 的 2,049 byte 大半是零，
像是商店欄位而不是文字。**下一步是解這套文字編碼**；在那之前不實作進入建築。

## RE-062：城鎮文字是 5-bit 打包的——9,515 行全部解出

### 先排除的三條路

建築內文在**任何原始檔裡都沒有明文**：`Once inside`、`fortress`、`counter`、
`Presently` 這些字在 75 個檔案中全部零命中。接著試過的也都零命中：

| 問法 | 範圍 | 結果 |
|---|---|---|
| 全檔 XOR 1–255 | 75 個原始檔 | 零命中 |
| 全檔位移 ±1–255 | 同上 | 零命中 |
| 5／6 bit 打包，固定字母表 | 整檔、未定起點 | 零命中 |
| 掃 `and ..., 1Fh` 找解碼器 | 全部 338 個函式 | **一個都沒有** |

最後一列值得記下來：**解碼器不存在於 KOLTOWN.EXE 的這個形式裡**，
所以「找解碼器再讀格式」這條路在這裡走不通，得由資料自己說話。

### 三個文字庫（`已證實`）

`sub_1A8FD(記錄號, 文字庫)` 依 `0x0D` 分隔取第 N 段，是全遊戲的訊息入口
（KOLTOWN 裡有 331 個呼叫點）。文字庫是 archive 的 payload，LZW 解碼後才是明文：

| 檔案 | payload | 長度 | 內容 |
|---|---|---|---|
| `KOLDATA` | 1 | 6,486 | 395 筆城鎮／戰鬥 UI 訊息 |
| `BOOKDATA` | 1 | 2,674 | 73 筆書本 UI 與工作人員名單 |
| `ENDTXT` | 0 | 3,381 | 18 筆結局敘事 |

標記語法：`~Vxx`／`~Hxx` 是十六進位的垂直／水平座標，`~Cxx` 是顏色，
`@` 是換行，`|X` 是代換記號（`sub_1A8FD` 在 `0x1A961` 處理 `7Ch` ＝ `|`）。

### 建築內文的版面（`已證實`）

用畫面上的原文當 crib 去掃**位元偏移**，在 `TOWN7.0` payload 9 的位移 129 命中
`ONCE INSIDE THE WALLS`。命中點全部落在 `16k + 1`，而 15 byte × 8 ÷ 5 ＝ **24**，
正是城鎮對話框的欄數——版面因此確定：

```
位移 128 起，每 16 byte 一格：
  byte 0      這一格的字元數（1…24）
  byte 1…15   24 個 5-bit 字元，MSB-first
```

一格就是畫面上的一行。字元表只有 32 個位置，所以小寫是預設、大寫靠一個記號：

| 值 | 意義 | 依據 |
|---|---|---|
| 0 | 空白 | 13.74%，最高頻 |
| 1–26 | a–z | 頻率與英文吻合（e 9.2%、t 6.2%、o 6.7%、j 0.11%、q 0.06%）|
| 27 | `"` | 常結束整格，其後最常接大寫記號（開引號）|
| 28 | `,` | 其後接空白或大寫記號 |
| 29 | `.` | 其後最常接 `"`，其次是句尾 |
| 30 | 跳脫記號 | 3.70%，見下表 |
| 31 | `'` | 後面幾乎只接 s／t（`'s`、`'t`）|

跳脫之後是同一顆鍵的 shift 位置，與打字機一致：

| 序列 | 符號 | 證據 |
|---|---|---|
| `[30]` ＋ 字母 | 轉大寫 | 每句開頭與專名 |
| `[30][28]` | `;` | `me; all that shuffling` |
| `[30][29]` | `:` | `reads:`、`specialty:`、`Yelthusa: Tleslass` |
| `[30][30]` | `-` | `well-dressed`、`low-roofed`、`run-down`（49 例）|
| `[30][31]` | `?` | 句尾與引號前 |
| `[27][27]` | `!"` | `says"Hi, Fighter!"`、`"Welcome to Kazhad!"` |

套上完整符號表之後，9,515 行**沒有任何未定的跳脫序列**。

驗證：`TOWN7.0` payload 9 的前八格解出

> Once inside the walls／of the small fortress／you approach the larger／
> of two buildings.you／enter and stand at a／scarred wooden counter.／
> Presently a man appears／from a back room.

與原版畫格 072 的對話框逐行相同；第 48–51 格解出的正是旅店老闆那段開場白。

### 訊息的分組與畫面規則（`已證實`）

**一則訊息固定 4 格 ＝ 64 byte**，對話框一次顯示兩則（八行）。半途之家的訊息 0 與 1
合起來正是進門那段敘述，訊息 10 起是店主的十幾種反應，每則自成四行、語意完整。

畫面規則由原版畫格 072 的八行逐格讀回比對出來，只有兩條：

| 資料 | 畫面 |
|---|---|
| `.` | `.` ＋兩個空白，且下一個字母轉大寫 |
| `,` | `,` ＋一個空白 |

套上之後八行與原版**逐字相同**（`of two buildings.you` → `of two buildings.  You`）。
資料裡已有的空白照留，所以 `,␣` 會變成 `,␣␣`——與原版畫面一致。

### 還沒定案的

- **引號前後的空白**：`says␣[30][27][30]hi…` 在畫面上是 `says"Hi, …`，
  引號前那個空白被吃掉。符號數與畫面字數對得上（21 個符號扣掉 3 個跳脫記號 ＝ 18 字），
  所以是位置規則而不是解碼錯誤。
- **610 個略過的格**（總格數 10,125 的 6%）：字數 byte 超出 24，那些位置不是文字格，
  多半是同一份記錄裡的商店或數值欄位。已逐格列在收據裡，不是默默丟掉。

### remake 對拍

`docs/verification/remake-parity/plan-town-building.json`：半途之家的進門畫面
（肖像 `TWN7PICS[12]`、標題框兩行、對話框八行、圖示 `KOLMPICS[4]`）對原版畫格 072
逐畫面對拍通過，允許區只有滑鼠游標。畫面完全由三份 typed data 組出來，沒有寫死的字串。

### 產出

`tools/extract_town_text.py` → `assets/original-derived-data/town-text.json`：
七座城、101 棟建築、**9,515 行**。這是繁體中文化真正要翻的城鎮文字母體。

## RE-063：城鎮文字的繁體中文化——布瑞提爾 21 棟建築

原文的斷行是原版 24 欄的結果，中文模式一行只有 22 個 16×18 字格，所以譯文自己帶斷行，
不在執行期做斷字。譯文表 `docs/i18n/town-zh-Hant.json` 的結構是「城／門／欄位」，
欄位是建築名、人物名與進門敘述。

- 布瑞提爾（TOWN7）21 棟建築的名稱、人物與進門敘述共 168 行已翻好。
- `internal/town/i18n.go` 把譯文套進進門畫面：沒有譯文的欄位保留原文，
  畫面不會出現空白。
- 字模表 371 → **770 字**；文字安全矩形檢查 125 → **286 項**全過
  （新增標題框置中寬度、對話框行寬與行數）。
- 英文模式不受影響：十二份對拍收據 203 對全過。

還沒翻的：其餘六座城的 80 棟建築、建築內的後續對話（每棟十幾則）、
以及 `KOLDATA`／`ENDTXT` 兩份 UI 文字庫。

## RE-064：解碼器讀出來了——RE-062 的字元表由統計升級為已證實

RE-062 的字元表是用畫面 crib 加上下文統計反解的，當時掃 `and ..., 1Fh` 找不到解碼器。
這一輪換個入口找到了：從 `sub_11A93` 裡「和 0x40／0x80 有關的算式」往下追。

### `sub_14234(位移, 旗標)`：訊息繪製迴圈（`已證實`）

```
if 旗標 == 0: sub_1452D()                 ; 先清空對話框
word_349AC = 0
loop:
    sub_1B94F(1CF4h + word_349AC + 位移, 0)   ; 解一格文字進緩衝區 2D50h
    word_2DF14 = 0Dh                          ; 螢幕 byte 欄 13 ＝ x 104
    sub_1A8FD(0, 2D50h)                       ; 畫那一行
    word_2DF16 += 8                           ; 行距 8
    word_349AC += 10h                         ; 下一格，16 byte
while word_349AC < 40h
```

`1CF4h` 就是 `byte_2FBC4`（DGROUP 0x2DED0 + 0x1CF4 = 0x2FBC4）。迴圈上限 `40h` ＝ 64，
每格 `10h` ＝ 16——**一則訊息剛好 4 格**，與 RE-062 從資料反解的完全相同。
呼叫點傳進來的位移都是 0x40 的倍數：`0x80`／`0xC0` 是進門敘述的兩則，
其餘出現過 `0x100`…`0x540`，以及一個由 `add [bp+var_6], 40h` 算出來的變動值。

### `sub_1B94F(位址, 模式)`：5-bit 解碼器（`已證實`）

輸出緩衝是 `2D50h`，界線 `2D68h`——差 `18h` ＝ **24 個字元**，正是對話框欄數。
取位元的迴圈是 `test [si],dl` ＋ `rcl al,1` ＋ `ror dl,1`（MSB-first，每次 5 位元），
所以掃 `and ..., 1Fh` 當然找不到——**它根本沒有用遮罩**。

`ah` 是跳脫狀態（0 或 FF），逐值的處理：

| 值 | 未跳脫 | 已跳脫 |
|---|---|---|
| `0` | 空白（**不動跳脫狀態**）| 同左 |
| `1`–`1A` | `al \| 60h` ＝ 小寫 | `al \| 40h` ＝ 大寫，跳脫關閉 |
| `1B` | `!` ＋兩個空白，跳脫**打開** | `"`，並**往回刪掉前面的空白** |
| `1C` | `,` ＋一個空白 | 原始 byte `1Ch`（原版字型的自訂字符）|
| `1D` | `.` ＋兩個空白，跳脫**打開** | `:` |
| `1E` | 跳脫記號本身，不輸出 | `-` |
| `1F` | `'` | `?` ＋兩個空白，跳脫**保持打開** |

RE-062 用統計猜的七個符號**全部命中**。兩件當時列為「還沒定案」的也一併解開：

- **引號前的空白為什麼消失**：`FF1Bh` 那條分支有 `dec di / cmp [di],20h / jz` 的迴圈，
  它把輸出尾端的空白往回刪掉再寫 `"`。
- **`[1B][1B]` 為什麼是 `!"`**：第一個 `1B` 給 `!` 與兩個空白並打開跳脫，
  第二個在跳脫狀態下變成 `"` 並把那兩個空白吃回去。

因此 `s a y s _ [1E] [1B] [1E] h i [1C] [1E] f i g h t e r [1B] [1B]` 解出來是
`says"Hi, Fighter!"`——與原版畫面**逐字元相同**。

### 修正

抽取器改成照解碼器實作，`display` 不再另外補空白。先前兩個實作錯誤：
空白被誤寫成會清掉跳脫狀態（於是 `.  ` 之後的 `"` 變成 `!`），
以及刪空白時把兩個空白當成一個元素而刪不掉。9,515 行重解之後沒有任何未定跳脫序列。

## RE-065：城鎮地形擋路、地圖邊界的城門提示，以及圖示列的生命週期

固定輸入 `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a9…d1d2`，IDA Pro 9.4
`ida-pro-9.4-idapython:locked-v1`，探針 `tools/ida_probe_koltown_world_map.py`。

### 地形擋路是磚塊編號的位元圖（`已證實`）

移動之前，`sub_10000:0x1082D` 先算出目的地那一格：

```
ax = (dy + word_2DF38) * 12                  ; dy ＋ 標記列偏移
bx = dx + 1Dh + ax + word_2DF36              ; dx ＋ 標記欄偏移；1Dh ＝ 列 2 欄 5
al = [bx + 2A48h]                            ; 畫面上 12×6 格的磚塊快取
mask = [(tile & 7) + 5Eh]                    ; DGROUP 0x5E ＝ 01 02 04 08 10 20 40 80
if [(tile >> 3) + 4430h] & mask: jmp 10A6D   ; 不能走，整段移動作廢
```

`4430h` 是 `TOWNn.0` payload 3 載入的位址（RE-061），所以**位移 0x00–0x1F 是
「磚塊編號 → 不可通行」的位元圖**，一個 bit 一個磚塊、LSB 先。0x20 起是門數，
與既有版面接得上。

交叉檢查兩項：TOWN0（野外地圖）的位元圖**全為零**，也就是沒有任何磚塊擋路；
七座城各擋住 125–141 個磚塊編號，落在地圖上是 723–2,041 格。
八座城**全部 103 個門的座標都落在可通行的磚塊上**——踩不到的門是不會有的。

### 夾限的是鏡頭，不是隊伍（`已證實`）

移動之後 `0x1093D` 起把 `word_2FB50`／`word_2FB56` 夾限在四個界限之內，
每夾一次就把 `word_39A4E` 加一。夾限發生時：

```
word_2FBC2 = 0FFFFh
word_345DA = sub_1198E()
if word_345DA != 0: word_2DF60 = 1; 回移動迴圈          ; U-TURN
else:                                                    ; DOOR：離開城鎮
    word_2DF68 = word_2DF66 = 0; word_2DF6A = word_2DF6C = 3Fh
    word_2FB50 = [4430h + 27h + 2k]; word_2FB56 = [4430h + 28h + 2k]
    sub_29004()                                          ; 載入 town0 的野外地圖
```

夾限的是鏡頭，隊伍位置是鏡頭加上 0…1 的標記偏移，所以**隊伍最遠可以站到界限＋1**。
四個界限來自角色記錄（`word_39A52` 基底的位移 0x14／0x15／0x17／0x18）；
兩個上限都是 0 時整組換成 0…63，也就是新角色拿到的是完整 64×64。

`sub_1198E` 是門口與城門共用的同一支程序，差別只在 `word_2FBC2`：
不是 0FFFFh 就畫 `sub_11302(word_2FBC2 + 3)` 的建築門面圖，是 0FFFFh 就畫
`sub_1103F(22h)` ——MISCPICS 的圖 0，192×96，正好是地圖視窗（RE-013）。
說明書把它描述成「地圖會換成林地中的布雷特爾圖像」，兩邊對得上。

### 圖示列是一個會被清空重建的堆疊（`已證實`）

- `sub_11703(id)`：`icons[word_2DF3C++] = id`，表在 DGROUP 0x2DAE。
- `sub_11722(arg)`：逐格畫 `sub_1103F(icons[i] + 3)`，白框畫在 `word_2DF3A`；
  `arg == 1` 只重畫就回來，其餘進等待迴圈。
- `sub_1449C()`：`word_2DF3C = 0`、`word_2DF3A = 0`，再 `sub_1103F(2)` 重畫地圖視窗。

於是四個轉換都定下來了：

| 時機 | 圖示列 |
|---|---|
| 移動之後掃門落空（`0x10B0B`）| 清空，並重畫地圖視窗 |
| 踩到門或鏡頭被夾限（`sub_1198E`）| 清空 → DOOR → U-TURN，白框回第一格 |
| 選 U-TURN | `sub_1449C()` → DOOR，只剩一格 |
| 從建築物出來（`0x10C32`）| 先清空 → DOOR，只剩一格 |

`<`／`>`（`0x2C`／`0x3C`、`0x2E`／`0x3E`）在地圖迴圈與門口迴圈都會移動白框並繞回，
條件都是圖示列非空。

### 原版觀察與對拍

新增 oracle 情境 `town-map-edge`：從旅店門口往東兩步再一路往北走到地圖上緣
（路徑由格子陣列與剛解出的位元圖算出），再多按一次方向鍵、白框移到 U-TURN、回地圖。
`tools/solve_outdoor_camera.py` 對這一輪 124 張畫格反解，**每一張戶外畫格的鏡頭都是唯一解**，
序列是 (58,5)→(58,4)→(58,3)→(58,2)→(58,1)→(58,0)，標記偏移固定 (0,1)。
畫格 118／119 不是戶外畫面（換成林地圖），120 之後回到 (58,0) 且圖示列剩一格 DOOR。

remake 對拍 `docs/verification/remake-parity/plan-town-edge.json`：四對逐畫面全過，
允許區只有原版的滑鼠游標。玩家路徑 `docs/verification/player-path/town-edge.json`
用同一串按鍵走完 13 步，其中兩步與已對拍截圖逐 byte 相同。

### 還沒實作的

選 DOOR 之後的野外地圖本身。`sub_29004` 已經讀出輪廓：`word_2DF2C = 0`、
載入 `town0` 的 payload 3 與 4、標記固定在視窗的 (3,3)、八方向移動、
偏移在 0…7 之間繞回時鏡頭走 `word_322F6`（初值 1），並先掃六名隊員記錄的
位移 0x1BA 取最大值（沒有就 20）當旅行速度。remake 目前把「在城門選了 DOOR」
記成 `Session.LeftTown`，交給上層，不假裝已經走出去。

## RE-066：世界地圖的城鎮記錄——七座城的位置、界限、三字母代碼與磁碟

`sub_29004` 把 `TOWN0` 的 payload 3 讀進 `byte_2FBC4`（DGROUP 位移 `0x1CF4`）。
迴圈 `0x294A8` 逐一比對 22 筆 16 byte 的記錄；座標吻合就設 `word_2FBC2 = i` 進城。
`sub_29CA3` 把那一筆記錄拆開（`已證實`，位移直接來自反組譯）：

| 記錄位移 | 寫進 | 意義 |
|---:|---|---|
| +0／+1 | — | 這個地點在世界地圖上的 X／Y |
| +2 | — | 索引（城鎮 10–15、礦坑 0–9）|
| +3／+4 | `word_2FB50`／`word_2FB56` | 進城之後的鏡頭座標 |
| +5／+6 | `word_2DF68`／`word_2DF6A` | 城鎮地圖的 X 界限 |
| +7／+8 | `word_2DF66`／`word_2DF6C` | 城鎮地圖的 Y 界限 |
| +9…+11 | `byte_349AE`…`byte_349B0` | 三字母代碼 |
| +12／+13 | — | 磁碟代號；`word_3B266 = [+13] - 30h` |

結尾是：

```
word_2DF36 = word_2DF38 = 0
word_2DF2C = word_2FBC2 + 1
if word_2DF2C > 6:
    word_2DF62 = word_2DF2C - 7     ; 入城點索引
    word_2DF2C = 7
```

也就是**記錄 i 對應 `TOWN{i+1}`**，而記錄 6 之後全部落在 `TOWN7`，
用入城點索引 0–9 區分——這正好解釋 RE-061 量到的「TOWN7 有 10 個入城點」。

解出來的十六筆有資料的記錄：

| 記錄 | 世界座標 | 代碼 | 城鎮地圖界限 | 對應 |
|---:|---|---|---|---|
| 0 | (59, 37) | `BRT` | 0–63 × 0–63 | TOWN1 |
| 1 | (51, 4) | `HTN` | 0–63 × 0–63 | TOWN2 |
| 2 | (15, 10) | `THM` | 0–63 × 0–63 | TOWN3 |
| 3 | (6, 29) | `POT` | 0–63 × 0–63 | TOWN4 |
| 4 | (37, 38) | `SHL` | 0–63 × 0–63 | TOWN5 |
| 5 | (24, 57) | `OLN` | 0–63 × 0–63 | TOWN6 |
| 6–15 | 十處 | `MIN` | 各 16×16 一塊 | TOWN7 的入城點 0–9 |

十個礦坑把 `TOWN7` 的 64×64 地圖切成 16×16 的十六宮格中的十塊
（記錄 6 是 0–15 × 0–15、記錄 7 是 16–31 × 0–15，依此類推）。
說明書寫的 `Trollsbane, BRT` 就是這裡的三字母代碼。

### 城名

代碼配上城鎮對話裡的一手證據，六個名字全部對上：

- `BRT` ＝ Brettle（TOWN1 的「Fortress of Brettle」）
- `HTN` ＝ Htron——TOWN7 門 3 的台詞「south is Brettle, west is Thimblewald and
  the Hobe, north is Htron」，配上世界座標 (51, 4) 在最北。
- `THM` ＝ Thimblewald——同一句話說它在西邊，座標 (15, 10) 也在西北。
- `POT` ＝ Poitle's Lock（TOWN4 門 3 明寫 "Poitle's Lock"，門 11 是 Lockkeeper）
- `SHL` ＝ Shellernoon（TOWN5 門 0「Lord of Shellernoon」）。
  **原版資料一律拼作 Shellernoon**，說明書的 `Shellemoon` 是異拼；一手資料為準。
- `OLN` ＝ Olanth（TOWN6 門 0「The Abbey of Olanthen」）
- `MIN` ＝ 礦坑，也就是 the Hobe（TOWN7 門 0「The Hobe's Main Gate」）

### 野外地圖的其他版面

同一份 payload 3 裡：位移 `0x200 + 磚塊編號` 是逐磚塊的一個 byte；
`0x300` 起 32 byte 是野外的不可通行位元圖（查法與城鎮相同，只是基底不同）；
`0x323`／`0x33B` 各 24 個 byte 是特殊地點的 X／Y，命中會呼叫 `sub_2B0FF(i)`。

野外的移動有**次格座標**：`word_2DF36`／`word_2DF38` 在 0…7 之間繞回，
繞回時世界座標才走 `word_322F6`（初值 1）。隊伍標記是 2×2 像素的點，
畫在視窗第 3 列第 5 欄那一格內、偏移 `次格座標 × 2` 像素——
由 `town-leave-run1` 的畫格量到：按一次方向鍵，點只移動兩個像素，地圖不動。

## RE-067：城鎮居民的話題系統——LISTEN 與 MOUTH 怎麼挑訊息

RE-064 解出了文字怎麼存，但沒解出「哪一則什麼時候出現」。這一輪從
`sub_11A93` 的圖示分派往下追，整套機制都出來了。

### 圖示分派（`已證實`）

`sub_11A93` 建築內迴圈的分派表（`0x13D06` 起）與 `sub_11703` 推進圖示列的
編號對得上。圖示畫的是 `KOLMPICS[編號 + 3]`：

| 編號 | KOLMPICS | 圖 | 跳往 |
|---:|---:|---|---|
| 1 | 4 | 手按鍵（SPACEBAR）| — |
| 2 | 5 | 門（DOOR）| `0x13D64` 離開 |
| 3 | 6 | 嘴（MOUTH）| `0x1223B` 輸入問題 |
| 4 | 7 | 耳（LISTEN）| `0x1254E` 聽傳聞 |
| 5 | 8 | 床（REST）| `0x1264E` |
| 6 | 9 | 人（COMPANION）| `0x12903` |
| 0x11 | 20 | 鏡（MIRROR）| `0x13CF6` → `sub_15ABF` 角色資料 |

### 六個話題槽（`已證實`）

每棟建築的 payload 位移 `0x17`…`0x1C` 是六個 byte：**低 6 bit 是話題碼、
高 2 bit 是類別**。命中第 n 槽就顯示訊息 `20 + n`（`sub_14234(580h + n*40h, 0)`）。

### LISTEN

```
sub_14234(500h, 0)                     ; 訊息 18：問話的敘述
n = sub_1B5A8(6) - 1                   ; rand(6)：隨機挑一個起點
if (payload[17h + n] & 3Fh) != 0: 換下一格   ; 有關鍵字的槽要用 MOUTH 問
if [1AA4h + n] != 0: 換下一格          ; 這個槽已經講過
[1AA4h + n]++
sub_1A8FD(9Ah + (payload[17h + n] >> 6), KOLDATA)   ; 依類別選開場白
sub_14234(580h + n*40h, 0)             ; 訊息 20 + n
```

找不到就 `n--`（繞回 5），回到起點還沒找到就講 KOLDATA 記錄 `0x2E`／`0x5B`
（`"Any other gossip?"`／`"Not lately."`）。**LISTEN 只講話題碼為 0 的槽**——
掛了關鍵字的槽要玩家用 MOUTH 問到那個字才會講。`0x1258C` 是 `jz` 才往下走，
非零就換下一格；把這個條件讀反會讓 LISTEN 講出應該要問才有的內容。

`[1AA4h + n]` 這六個旗標在進門時清零（`sub_11A93:0x11AB6`）。
KOLDATA 記錄 `0x9A`…`0x9D` 是四句開場白：
`"What do you know?"`、`"How are things?"`、`"Any news in town?"`、
`"What news have you of the world?"`。

### MOUTH

```
sub_14234(500h, 0)                     ; 訊息 18
sub_1B5C8(2E3Ah, 17h)                  ; 讀入最多 23 個字元
整串轉小寫
for var_2 = 0; var_2 < 100h; var_2 += 4:
    sub_1B94F(0D39Eh + var_2, 1)       ; 解出一個關鍵字（最多四個字元）
    if 輸入的任何一個詞以這個關鍵字開頭:
        碼 = [0D39Eh + var_2 + 3]
        for n = 5 downto 0:
            if (payload[17h + n] & 3Fh) == (碼 & 3Fh):
                sub_14234(580h + n*40h, 0)     ; 訊息 20 + n
                回主迴圈
        who = [0D39Eh + 1C0h + (碼 & 3Fh)]
        if who != 0FFh and 城號 != 7:
            「I heard <who> talking about <that/them/him>.」
sub_14234(540h, 0)                     ; 訊息 19：我什麼都不知道
```

比對是**字首比對**：關鍵字最多四個字元，玩家輸入以空白分詞後逐詞比對。

### `sub_1B94F` 的 mode 1（`已證實`）

RE-064 讀的是 mode 0。mode 1 不一樣，差別全在 `0x1B962` 起那幾行：

| | mode 0 | mode 1 |
|---|---|---|
| 字數 | 第 0 個 byte 整個 | 第 0 個 byte 的**高 4 bit** |
| 位元流起點 | 第 1 個 byte 的 MSB | **第 0 個 byte 的 bit 3** |
| 跳脫狀態 | 關 | **開**（首字大寫）|

### payload 20 的版面（`已證實`）

`sub_10000:0x106C9` 把 `TOWNn.0` 的 payload 20 讀進 DGROUP 位移 `0xD39E`：

| 位移 | 內容 |
|---|---|
| `0x000 + 4i` | 64 筆關鍵字：三個 byte 的 mode 1 字串 ＋ 一個碼 byte |
| `0x100 + 8i` | 16 個人名，同樣是 mode 1 |
| `0x1C0 + 話題碼` | 這個話題該去問誰（人名索引，`0xFF` ＝ 沒有人知道）|

七座城共 305 個關鍵字，全部是四字元以內的字首，而且**按字典序排好**：
`aklo`、`band`、`bron`、`butl`、`dark`、`duke`、`elro`、`elve`…
人名解出來是 `benjamin`、`elmu`、`friar huxley`、`sir celegorn` 等，
與門的人物名對得上。

### 交叉檢查

101 棟建築共 **409 個非空話題槽**：每一個的話題碼都在該城的關鍵字表裡找得到，
而且每一個都有對應的訊息 `20 + 槽號`。**零例外。**

抽取器 `tools/extract_town_keywords.py` → `assets/original-derived-data/town-keywords.json`；
三份 UI 文字庫另由 `tools/extract_ui_text.py` 抽成 `ui-text.json`
（KOLDATA 395、BOOKDATA 73、ENDTXT 18 筆）。

### 招呼語（`已證實`）

讀完進門敘述按空白鍵之後（`sub_11A93:0x12017`）：

```
tier = sub_14371() / 6                 ; 聲望
if tier > 3: tier = 3
if tier != 0:
    sub_14234(700h + tier*40h, 1)      ; 訊息 26 + tier → 27／28／29
else:
    idx = [39A52 + 1Ah] << 7           ; 角色記錄位移 0x1A ＝ 性別
    if [39A52 + 1A0h] != 0: idx += 40h ; 會法術
    sub_14234(300h + idx, 1)           ; 訊息 10／11／12／13
```

也就是「男 ＝ 10、男法術 ＝ 11、女 ＝ 12、女法術 ＝ 13」；聲望不為零就整個換成
訊息 26 ＋ 等第。後面接 KOLDATA 記錄 45 的 `"What can I do for you?"`。
隊伍不只一人時另有訊息 16（`word_2DF2A != 1`）。
**這一刻白框也回到隊員身上**——進門那一段說話的是建築裡的人，所以隊員外框是藍的。

### 對拍

原版情境 `town-listen`（DOSBox-X），三對逐畫面對拍全過，允許區只有滑鼠游標：
招呼語與六格圖示列、LISTEN 的第一段（訊息 18 ＋ 開場白）、第二段（訊息 20）。
圖示列在 `y = 150…200` 這一段**逐 byte 相同**，六個圖示的順序與白框位置都對上。

## RE-068：野外地圖的畫面——版面、地形標籤與那個會閃的點

`sub_2A39B` 是野外畫面的重畫入口，矩形全部是它與 `sub_29B40` 的引數（`已證實`）：

| 元素 | 呼叫 | 矩形 |
|---|---|---|
| 底圖 | `sub_1103F(0)` | KOLMPICS 圖 0，與城鎮共用 |
| 地圖視窗 | `sub_1B7E9(67h,39h,128h,9Ah)` | (103,57)–(296,154)，與城鎮同一個 |
| 磚塊 | `sub_115D1()` | 共用同一支繪製迴圈，偏移一樣是 −5／−2 |
| 天候圖框 | `sub_1B7E9(0Fh,4Bh,50h,8Ah)` | (15,75)–(80,138)，內容區 (16,76) 64×62 |
| 標籤框 | `sub_1BA9D(0Eh,8Eh,50h,9Ch,0Fh)` | (14,142)–(80,156) |

### 標籤的兩行（`已證實`）

`sub_29B40` 把游標放在 byte 欄 2／列 0x90 與 byte 欄 3／列 0x96，各畫一筆 KOLDATA：

```
sub_1A8FD(word_2DF42 + 0D5h, KOLDATA)    ; 時段：DAWN／MID MORN／NOON／AFT NOON／TWILIGHT
sub_1A8FD(word_3499A + 0DAh, KOLDATA)    ; 地形：DESERT／RIVER／SWAMP／HILL／ROAD／FOREST／SHORE／PLAINS
```

**記錄自帶空白**（`"  DAWN ]"`），置中是靠那些空白做的；抽取時把空白去掉，
畫出來就會往左移兩個字。remake 因此另留一條「保留前後空白」的取字路徑。

地形碼由 `sub_29987` 算出來：`payload3[0x100 + 磚塊編號]` 取**最低那個 1 的位序**，
整個 byte 為 0 時是 7（PLAINS）。同一支還算出 `word_2FA2A = payload3[0x200 + 磚塊]`，
那是道路旗標，非零時圖示列會多出 ROAD 與 SIGNPOST。
時段 `word_2DF42` 是「`word_2DF40` 落在 `[12CCh + 2i]` 這五個門檻的哪一段」。

### 隊伍那個點（`已證實`）

視窗第 3 列第 5 欄那一格內，位置是「次格座標 × 2」像素；次格座標在 0…7 之間繞回，
繞回時世界座標才走一格。點是 2×2，而且**不是塗白，是把底下四個像素的 EGA 索引
與 15 互斥或**——`town-leave-run1` 的畫格量到 `0 → 15` 與 `8 → 7` 兩組對應，
後者排除了「OR 15」的解釋。它每隔幾個 tick 閃一次。

### 對拍

`docs/verification/remake-parity/plan-world.json` 四對全過：隊員格與日期框、
地圖視窗與那個點、標籤框、圖示列，允許區只有原版滑鼠游標。

### 還沒實作的：雲層動畫

`sub_2A18E:0x2A1B6` 每個 tick 把覆蓋層**整段以 16 byte 為單位向右旋轉一個位元**
（`rcr` 帶進位，從單位末端繞回開頭），再照 `(底圖 AND 遮罩) OR 覆蓋層` 合成。
旋轉範圍是 0x9B × 16 ＝ 2,480 byte，從覆蓋層起點算起正好蓋到 1-bit 遮罩的末端。
把它照「每列 64 像素繞回」或「兩列 128 像素繞回」重現，對原版畫格的最佳吻合都還差
四百個像素，**緩衝區的位元組版面因此尚未定案**。remake 目前畫的是未旋轉的合成圖，
對拍計畫用 `compare` 明寫每一對涵蓋的矩形，天候圖那一塊列在 `not_claimed`——
不是允許差異區，是還沒宣稱。

## RE-069：城門與世界地圖的座標是同一組

RE-065 讀到「在城門選 DOOR 之後 `word_2FB50 = [4430h + 27h + 2k]`」，
而 RE-066 讀到世界地圖城鎮記錄的前兩個 byte 是那個地點的世界座標。
兩張表逐一比對，**16 個入城點全部相同**：

| 城 | payload3 的 `0x27 + 2k` | 世界地圖記錄 |
|---|---|---|
| 1 | (59, 37) | 記錄 0 `BRT` (59, 37) |
| 2 | (51, 4) | 記錄 1 `HTN` (51, 4) |
| 3 | (15, 10) | 記錄 2 `THM` (15, 10) |
| 4 | (6, 29) | 記錄 3 `POT` (6, 29) |
| 5 | (37, 38) | 記錄 4 `SHL` (37, 38) |
| 6 | (24, 57) | 記錄 5 `OLN` (24, 57) |
| 7 | 十組，(30,24)…(4,3) | 記錄 6–15 `MIN`，逐一相同 |

所以那一欄不是「入城點在城鎮地圖上的位置」，而是**這個城門在世界地圖上的格子**。
進城的落點是另一組欄位（世界地圖記錄的 +3／+4，由 `sub_29CA3` 讀）。

`town-leave-run1` 那一輪的 `word_2DF62` 是 3，離開霍布之後落在 (40,23)，
與記錄 9 相同——remake 的玩家路徑走完同一串按鍵，畫面與野外對拍截圖逐 byte 相同。


## RE-070：野外走到城鎮入口——外觀圖、腳下那一格與進城的整組欄位

輸入：`KOLTOWN-unpacked.EXE`（SHA-256 `01f6f9a9…d1e2`）的 IDA Pro 9.4 資料庫，
探針 `tools/ida_probe_koltown_world_map.py` 經 `tools/run_ida_probe.sh` 於
`ida-pro-9.4-idapython:locked-v1` 執行；證據等級 `已證實`（逐指令讀出並以原版實跑對拍）。

### 走到城鎮入口那一格會發生什麼

`sub_29004` 的主迴圈在**世界座標真的變了**之後（`word_3AC5A != 0`，`0x293B7`
的分支否則整段跳過）依序做三件事：查地形擋路、掃 24 個特殊地點、掃 22 筆城鎮記錄。
第三段命中時：

```
word_2FBC2 = i                       ; 記錄編號
ax = sub_29BBB()
if ax == 0FFFFh: 離開野外迴圈（進城）
if ax == 0:      整頁重畫
else:            只重畫地圖視窗，回到等鍵盤
```

`sub_29BBB` 自己是那個提示：

```
sub_11302(payload3[0x395 + 16i] + 8) ; TWN0PICS[記錄索引 + 8]，192×96，畫滿地圖視窗
sub_1C14E()                          ; 清鍵盤緩衝區
sub_11703(2); sub_11703(0Fh)         ; DOOR、U-TURN
var_2 = sub_11722(3)
if var_2 == 2:  sub_29CA3(); return 0FFFFh          ; DOOR → 進城
sub_1449C()                                          ; 清圖示列
if word_2FA2A != 0: sub_11703(17h); sub_11703(18h)   ; 站在路上就補 ROAD、SIGNPOST
if var_2 == 0Fh: sub_11703(2); sub_11722(1); return 0FFh   ; U-TURN → 只剩一格 DOOR
```

`TWN0PICS` 有 25 張圖，全部 192×96。記錄索引 0–9 是礦坑與洞窟（圖 8–17），
10–15 是六座城（圖 18–23）；抽出來的 `twn0pics-011.png` 與原版提示畫面的
地圖視窗逐 byte 相同（差異 76 px 全部落在 DOSBox-X 的滑鼠游標上）。

### `sub_29CA3`：進城要搬的九個欄位

記錄長 16 byte，基底 `payload3[0x393 + 16i]`：

| 記錄位移 | 去向 | 意義 |
|---|---|---|
| +0／+1 | 比對 `word_2FB50`／`word_2FB56` | 這個地點在世界地圖的格子 |
| +2 | `sub_11302(值 + 8)` | 外觀圖索引 |
| +3／+4 | `word_2FB50`／`word_2FB56` | 進城之後的鏡頭座標 |
| +5／+6 | `word_2DF68`／`word_2DF6A` | 城鎮地圖的 X 界限 |
| +7／+8 | `word_2DF66`／`word_2DF6C` | 城鎮地圖的 Y 界限 |
| +9…+11 | `byte_349AE`…`byte_349B0` | 三字母代號（BRT、MIN…） |
| +13 | `word_3B266 = 值 − 30h` | 磁片編號（ASCII 數字） |

接著 `word_2DF36 = word_2DF38 = 0`（標記偏移歸零，隊伍就站在鏡頭那一格），
`word_2DF2C = i + 1`；`i + 1 > 6` 時再 `word_2DF62 = (i+1) − 7`、`word_2DF2C = 7`。
所以六座城是記錄 0–5，十個礦坑與洞窟共用 TOWN7 的地圖，用入城點 0–9 區分——
與 RE-069 由另一端（城門的離城座標）得到的對應完全一致。

### 腳下那一格是 (X, Y+1)

`sub_29987` 取地形的方式是：

```
word_2FA2A = payload3[0x200 + byte_30941]    ; 道路方向遮罩
v = payload3[0x100 + byte_30941]
word_3499A = (v == 0) ? 7 : v 最低位的 1 的位序
```

`byte_30941` 不是另外算出來的座標，而是**視窗緩衝區的第 0x29 個 byte**
（DGROUP 0x30918 + 0x29；0x29 ＝ 3×12＋5，第 3 列第 5 欄）。繪製迴圈與城鎮共用
`sub_115D1`，列偏移固定 −2，因此第 3 列畫的是 `Y + 1`：**野外的地形名稱、
道路旗標與地形擋路看的都是 `(X, Y+1)`，不是鏡頭座標那一格**。
擋路那一段讀 `viewbuf[0x29 + dx + 12*dy]` 也是同一個基底。

原版實跑印證：`world-enter-town-run1` 走到 (41,23)，標籤是 `FOREST`；
`grid[41][23]` 是磚塊 250（道路），`grid[41][24]` 是磚塊 85（森林）。
把 `Tile()` 改成 `(X, Y+1)` 之後，兩個畫面的標籤框與圖示列都與原版逐 byte 相同。

標籤本身放在 `word_3499A`，只有 `sub_297F2` → `sub_29B40` 會重畫
（`sub_29B40` 畫的是 KOLDATA 記錄 `0D5h + 時段` 與 `0DAh + 地形碼`）。
而走到城鎮入口的那一步會跳進 `sub_29BBB`，這段整個被略過——所以提示畫面上的
地形標籤停在**上一格**的地形。remake 也照這樣處理。

### 順帶讀出、還沒實作的兩件事

- ROAD（圖示 17h）與 SIGNPOST（18h）都走 `sub_29DDE`，差別是引數 0 或 1：
  `word_2FA2A` 是八個方向的道路位元遮罩，它先看目前方向 `word_2FB5E` 還通不通，
  不通就往兩側各找三格；SIGNPOST 先把目前方向記進 `word_2FBBA`，再從那裡逆時針
  掃一圈找**另一條**路。找到就設 `word_37D28 = 1` 並回傳一個模擬的方向鍵碼，
  由主迴圈當成玩家按鍵處理。`word_37D28` 設起來之後只准走道路磚塊
  （`0x2930C` 那段會把踏上非道路的一步退回）。
- `word_322F6` 是「次格繞回時世界座標走幾格」，起始 1。`0x29691` 的
  `xor word_322F6, 0Bh` 把它在 1 與 10 之間切換，而觸發它的是**鍵碼 2**，
  不是任何圖示；同一段還有鍵碼 1 直接把 `word_2FB50`／`word_2FB56` 印出來。
  兩個都是開發用的按鍵，不是玩家功能。

### 一個判讀陷阱

原版提示畫面比 remake 多一個 14×11 的白色左箭頭，位置在 (162,102)。
那不是遊戲元素，是 **DOSBox-X 的滑鼠游標**——同一輪的城門提示畫面同樣有它，
而更早的 `town-leave-run1`（游標停在別處）沒有。參考卡第 11 章寫的
「滑鼠：將箭頭移到圖示」就是這個游標。對拍計畫把它列為允許差異，不是遊戲行為。

## RE-071：道路是方向遮罩，ROAD 與 SIGNPOST 是同一支自動循路

輸入與工具同 RE-070；證據等級 `已證實`（表格逐 byte 倒出，並以原版實跑逐畫面對拍）。

### 道路表存的是八個方向

`payload3[0x200 + 磚塊編號]` 不是「這格是不是路」的布林值，而是**八個方向的位元遮罩**。
`sub_29DDE` 用 `byte_1312` 那張表取位元，四張表原樣倒出來是：

| DGROUP | 內容 | 值 |
|---|---|---|
| `0x1312` | 方向 → 位元 | `80 40 20 10 08 04 02 01`，方向 d 對應 `1 << (7 − d)` |
| `0x12F2` | 方向 → X 位移 | `0, +1, +1, +1, 0, −1, −1, −1` |
| `0x1302` | 方向 → Y 位移 | `−1, −1, 0, +1, +1, +1, 0, −1` |
| `0x12E2` | 方向 → 鍵碼 | `4800h 4900h 4D00h 5100h 5000h 4F00h 4B00h 4700h` |

方向 0 是正北，順時針數到 7 是西北；鍵碼就是 BIOS 的方向鍵掃描碼
（48h 上、49h PgUp、4Dh 右、51h PgDn、50h 下、4Fh End、4Bh 左、47h Home）。
IDA 另外命名的 `word_2F1B4`／`2F1B8`／`2F1BC`／`2F1C0` 其實是同一張鍵碼表的
第 1、3、5、7 項，也就是四個斜向。

### 選方向的規則

```
ROAD(17h)      → sub_29DDE(0)
SIGNPOST(18h)  → word_2FBBA = word_2FB5E; sub_29DDE(1)
```

- 目前方向 `word_2FB5E` 還有路就沿用；沒有就 `(dir − k) & 7`、`(dir + k) & 7`
  逐一試 k ＝ 1…3，兩側都沒有就把方向清成 0FFh。
- 引數非 0（SIGNPOST）或方向是 0FFh 時改掃一圈：由 `word_2FBBA`（沒有就 7）
  逆時針走，跳過 `word_2FBBA` 自己，取第一條有路的。SIGNPOST 因此是
  「這個路口的**另一條**路」。
- 找不到就回傳 0，什麼都不做。

### 走的那一步會先把隊伍拉回中線

正東西南北的四個方向另有一段修正：隊伍的次格座標偏離中線時改走一個斜向。
以正東為例，次格 Y > 4 就走東北、< 3 就走東南，3 與 4 視為已經在中線上。
北、南看的是次格 X，東、西看的是次格 Y。斜向方向不修正。

選定的那一步設 `word_37D28 = 1`，只准落在道路磚塊上——`sub_29004:0x2930C`
會把踏出道路的一步逐軸退回。這個旗標每次讀鍵前都清成 0（`0x29120`），
所以只作用在模擬出來的那一步。

### 圖示列的框色有三種

`sub_11722` 畫圖示列時，白框那一格是 0Fh、其餘是 9（亮藍）。但野外主迴圈自己
處理游標，`0x2973B` 與 `0x29765` 把**離開的那一格**畫成 **3（青色）**。
所以三種狀態分別是：目前這一格白、被選過又離開的青、還沒被選過的亮藍。
按下 Return 之後白框一律回到第 0 格（`0x29783`）。

### 對拍

`world-road-run1`：出城 → 往北十二步站上 (40,21)（腳下那一格是 (40,22)，磚塊 251，
遮罩 12h ＝ 方向 3 與 6）→ 連按七次 ROAD → SIGNPOST。remake 走同一串按鍵，
地圖視窗、地形標籤與圖示列在四個取樣點都與原版逐 byte 相同
（`plan-world-road`，10 對全過）。**唯一的差別是時段**：第八步時原版由 DAWN
變成 MID MORN，那是還沒實作的時間推進（`sub_297F2` 的 `word_2DF40` 倒數），
已列為不宣稱。

## RE-072：招募同伴——候選人怎麼挑、清單怎麼捲、受傷的招呼語

輸入與工具同 RE-070；`sub_11A93` 完整倒出（2,874 條指令）。
證據等級 `已證實`（清單行為由原版實跑逐畫面確認，受傷分支逐指令讀出並以資料驗證）。

### 兩個畫面

按下 COMPANION（圖示 6）之後：

1. 對話框畫 KOLDATA 記錄 392（`Look for characters on@~H0Dhard disk?`），
   圖示列換成 **KOLMPICS 10／11**——大拇指朝上與朝下，也就是 id 7 與 8。
2. 回答之後對話框改畫候選人清單，圖示列換成大拇指朝上與 U-TURN。

清單的版面就是對話框自己的字格：起點 (104, 86)、行距 8、一次八列。
反白那一列是白字，其餘是 **EGA 3（青色）**；右側第 23 格畫大字型碼位
94／95 的上下箭頭，**顏色是 EGA 4（紅色），而且兩個箭頭一直都在**，
不隨捲動位置消失（`town-recruit-list-run1` 的三個取樣點逐 byte 相同）。

上下方向鍵移動反白：**兩端都夾住，不繞回**；反白走到第八列之後清單一次捲一列。
半途之家有 11 位候選人，按 14 次向下之後停在第 11 位，再按沒有反應。

### 候選人是誰

角色檔位移 64 的 10 byte 所在地欄，與建築名稱轉大寫後的前 10 個字元相同：
`Halfway House` → `HALFWAY HO`。再比對位移 144 的三字母城鎮代碼，
並扣掉已經在隊伍裡的人。半途之家（TOWN7 入城點 3，代碼 `MIN`）因此得到
12 − 1 ＝ 11 位，與原版畫面上的 11 列逐一相同。

頭像取自存檔的 `CHARPICS`（16 × 760 byte）。16 格逐像素與 `CHARPICS.0` 的
80 張模板比對，**全部命中**，因此 typed data 記的是模板圖號；抽取器對不上就失敗。
槽位 0 命中模板 47，與旅店畫面對拍用的圖號相同。

### 受傷的招呼語

`sub_11A93:0x11F47` 讀的是**角色記錄位移 0x1B**：

```
if party[speaker].0x1B != 0:
    if byte_2FC00 & 40h: sub_1A8FD(72h)          ; 這棟建築能治療
    else: sub_1A8FD(2Bh); sub_1A8FD(2Ch)         ; "You are hurt! You should go to the abbey."
else:
    ...一般招呼語（RE-067）
```

`byte_2FC00` 是建築 payload 位移 0x3C 的旗標，bit6 就是給 REST 的那一位。
記錄 72h 是 `"Those wounds need tending."`。

資料側對得上：`CHARDATA` 位移 0x1B 只有 Poontang（1）、One tooth（1）與
Warder（20）非 0。原版招募 Poontang 之後的畫面正是那兩行，而領隊 Theodoric
（0x1B ＝ 0）進門時是一般招呼語。位移 0x1A 同樣讀出來是性別，
Moraine 與 Shyla 為 1，與 RE-067 由招呼語推出的性別一致。

### 對拍

`plan-town-recruit` 九對全過：問句、清單、反白移動、加入之後的隊員格、
標題框、招呼語與圖示列，逐 byte 相同，允許差異只有原版的滑鼠游標。
玩家路徑 `town-recruit` 14 步。

## RE-073：客棧住宿與修道院治療——價錢從哪來、存檔寫了什麼

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_rest_save.py`，
把 `sub_14DB1`（415 條）與 `sub_13E82`、`sub_144EF`、`sub_14D3A`、`sub_14CCF`
完整倒出，並附每一個 `sub_1A8FD` 呼叫點畫的是哪一筆 UI 記錄。
證據等級 `已證實`（價錢公式由 15 棟建築的資料與原版實跑一致確認，
存檔欄位逐指令讀出）。

### REST 是三條路，由建築旗標分岔

圖示列上的 REST 在兩種建築都會出現（`payload[0x3C]` 的 bit3 與 bit6，
見 RE-067）。按下去之後 `sub_11A93:0x1264E` 先看 bit6：

| 旗標 | 建築 | 行為 |
|---|---|---|
| bit6（0x40） | 修道院 | 治療。傷勢 0 就只講記錄 0x62，不收錢。 |
| bit3（0x08） | 客棧 | 住一晚：收錢、寫存檔、角色離隊。 |

### 修道院：每一點傷勢一個價

```
wounds = party[speaker][0x1B]
if wounds == 0: say(62h); return                 ; 0x12675
price  = wounds * payload[0x39]                  ; 0x126A3，8-bit 乘法
say(39h); say(53h); print(price); say(54h)       ; "If you need healing…it will cost you…gold, interested?"
if answer != YES: return                         ; 0x12717
if price > gold: say(56h); return                ; "Sorry, not enough gold crowns!"
gold -= price                                    ; 0x1276E
party[speaker][0x1B] = 0; party[speaker][0x9F] = 0
say(63h)
```

七間修道院的 `payload[0x39]` 是 9…12，也就是每一點傷勢 9 到 12 枚金幣。

### 客棧：價錢是 payload 的一個 16-bit 欄位

```
price = payload[0x38] << 8 | payload[0x39]       ; sub_14DB1:0x14E38
say(37h)                                         ; '"If you want to stay@'
if price == 0: say(38h)                          ; '~H0Djust grab an open room!"@'
else:
    say(53h); print(price); say(54h)
    if answer != YES: return
    if price > gold: say(56h); return
    gold -= price
```

八座城鎮共 10 間客棧，價錢 0、56、57、60、60、65、75、120。
半途之家（TOWN7 門 5）是 0x00 0x38 ＝ 56，與 `town-rest-run1` 畫面上的
「it will cost you 56 gold」相同；同一次以兩人隊伍再跑一次（`town-rest-two-run1`），
答完 `Whole party?` 之後報的仍是 56——**價錢是每人一晚的定價，不隨人數改變**。

隊伍人數大於 1 時先問記錄 0x35 `Whole party?`（`0x127C3`）。
回答 YES 就對六個槽位逐一跑住宿流程，NO 只跑目前發言的那一位。

### 金幣是角色自己的錢包

`sub_14CCF` 與 `sub_14D3A` 是同一個欄位的讀寫：角色記錄位移 0x2A、0x2B、0x2C
是**高位元組在前的 24-bit 金幣**。交易前 `sub_14CCF` 把它讀進
`word_30A6E:word_30A70`，交易後 `sub_14D3A` 寫回去。所以原版沒有共用錢包，
花的是當下那位發言者的錢。

### YES／NO 的回傳值

`sub_144EF` 推兩個圖示（7 大拇指朝上、8 朝下）之後等選擇，回傳值減 7：
**0 是 YES，1 是 NO**。價格問句、`Whole party?`、`Save character to hard disk?`
與 `Do you wish to quit?` 共用這一支。

### 住宿寫進存檔的欄位

付完錢之後 `sub_14DB1` 做的是完整的存檔（`0x14F74` 起）：

1. `int 21h/19h` 取得目前磁碟機。若不是 A:／B:（`var_A > 1`）就問記錄 0x189
   `Save character to hard disk?`；回答 NO 才 `int 21h/0Eh` 切到 A:。
2. 讀 `chardata`（16 × 0x300 byte）到緩衝區，逐槽比對前 16 個 byte 的名字，
   找到同名的槽就覆寫；找不到就用 `byte_30C20` 為 0 的第一個空槽；
   兩者都沒有就講記錄 0x64 並放棄。
3. 把當下的狀態寫進角色記錄：

   | 位移 | 內容 | 來源 |
   |---|---|---|
   | 0x2A–0x2C | 金幣（24-bit，高位在前） | `sub_14D3A` |
   | 0x2E–0x2F | 建築編號 | `payload[0x3A]／[0x3B]` |
   | 0x40–0x49 | 所在地（建築名前 10 字元轉大寫） | 建築名字串 |
   | 0x13 | 入城點 | `word_2DF62` |
   | 0x14／0x15／0x17／0x18 | 鏡頭夾限 | `word_2DF68／6A／66／6C` |
   | 0x90–0x92 | 三字母城鎮代碼 | `byte_349AE／AF／B0` |
   | 0x94 | 顯示模式 | `word_3B266` |
   | 0x97 | 城鎮編號 | `word_2DF2C` |
   | 0x99 | 城內 X | `word_2FB50 + word_2DF36` |
   | 0x9C | 城內 Y | `word_2FB56 + word_2DF38` |

4. `rep movsw` 把 0x300 byte 的角色記錄複製進緩衝區的那一槽，寫回 `chardata`；
   接著用同一個槽號（× 0x2F8）把 `steedpic`、`charpics`、`shldpics`
   三份 16 × 0x2F8 的圖檔一起改寫。
5. `party[slot][0] = 0`、`word_2DF2A -= 1`——**住宿存檔會把角色移出隊伍**。
   隊伍空了之後 `0x128AA` 問記錄 0x36 `Do you wish to quit?`，
   回答 YES 就 `sub_2DB2A("kolbook.exe", 0)` 交棒回書本程式。

這一串就是原版唯一的存檔入口：**住宿即存檔即離隊**。

### 資料側的三個交叉驗證

把上表拿去讀原版的 `CHARDATA` 就對得起來，而且對出兩件原本不知道的事：

* **0x2E／0x2F 就是建築 payload 的 `record_id`。** 14 名角色中 12 名是 86（0x56 ＝
  半途之家）、`place` 欄是 `HALFWAY HO`；另外兩名是 14（0x0E ＝ 巨魔剋星客棧）、
  `TROLLSBANE`。與建築 payload 讀出來的 `record_id` 逐一相同。
* **0x97／0x13／0x99／0x9C 是城鎮、入城點與城內座標。** Theodoric 讀出來是
  城鎮 7、入城點 3、座標 (57, 8)——正是 TOWN7 半途之家那扇門的座標。
* **0x90–0x92 是三字母城鎮代碼，不是日期。** 讀出來是 `MIN`，與角色清單畫面
  右欄顯示的城鎮代碼相同；`sub_14DB1:0x15101` 寫進去的 `byte_349AE/AF/B0`
  因此是當下城鎮的代碼。原版記錄裡目前**沒有**定位到日期欄位。

## RE-074：MIRROR 的角色資料頁——屬性是內部值的三分之二

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_mirror.py`（三輪，
分別倒出 `sub_15ABF`／`sub_16239` 與四支繪圖 helper，以及 DGROUP 的身高體重表）。
原版收據 `docs/verification/character-ui/town-mirror-run1`。
證據等級 `已證實`：每一個欄位都用原版存檔裡 Theodoric 的 byte 算過一次，
與原版畫面上的數字逐項相同。

### 建築內圖示的派工表

`sub_11A93:0x13D06` 是一串 `cmp ax, N` 的比較鏈，`ax` 是圖示編號：

| 編號 | 圖示 | 處理位址 |
|---:|---|---|
| 2 | DOOR | `0x12238` |
| 3 | MOUTH | `0x1223B` |
| 4 | LISTEN | `0x1254E` |
| 5 | REST | `0x1264E`（RE-073） |
| 6 | COMPANION | `0x12903` |
| 9 | 買賣 | `0x1290B` |
| 0Ah | 訓練 | `0x1318D` |
| 0Bh | HEAL | `0x135EF` |
| 0Ch | AWARD | `0x1375C` |
| 0Dh | SCROLL | `sub_17318` |
| 11h | MIRROR | `sub_15ABF`，回來時 `sub_152C6` 重畫城鎮 |

### 八個屬性：畫面值是內部值的 100/150

`sub_15ABF:0x15BCE` 起連續八次 `sub_161FC(值)`，每次都先把游標移到第 12 欄、
往下 8 px，再用 `%3d` 印出來。前六個直接讀角色記錄，後兩個是算出來的：

| 列 | 標籤（KOLDATA 記錄 73h） | 來源 |
|---|---|---|
| 1 | STRENGTH | `char[0x1D] × 100 / 150` |
| 2 | QUICKNESS | `char[0x1E] × 100 / 150` |
| 3 | HEALTH | `char[0x20] × 100 / 150` |
| 4 | FORESIGHT | `char[0x21] × 100 / 150` |
| 5 | CHARISMA | `char[0x22] × 100 / 150` |
| 6 | INTELLECT | `char[0x23] × 100 / 150` |
| 7 | BALANCE | `(char[0x23] + char[0x1D]) / 2 × 100 / 150` |
| 8 | ENDURANCE | `(char[0x20] + char[0x1D]) / 2 × 100 / 150` |

沒有出現在畫面上的 `char[0x1F]` 是體型。七個內部屬性因此是
0x1D–0x23：力量、敏捷、體型、健康、先見、魅力、智力，與說明書
`07-advanced-gamers.md` 的順序相同；說明書也明說「平衡是力量與智力的平均」、
「耐力由力量與健康計算」，與上表兩條算式一致。

Theodoric 的記錄是 130、100、116、102、103、94、109，套上去得
86／66／68／68／62／72／79／77，與 `frame-073` 上的八個數字逐一相同。
**畫面值是內部值的三分之二**——SPEC-005 說的「屬性是內部值，不是畫面值」
在這裡有了確切的換算。

### 抬頭：名字、性別、種族、職業、位階

`sub_16239` 畫的那一塊：

| 欄位 | 來源 | Theodoric |
|---|---|---|
| 名字 | 記錄位移 0 的 Pascal 字串，22 欄置中 | THEODORIC |
| 性別 | KOLDATA 記錄 `char[0x1A] + 4`（4 男 5 女） | MALE |
| 種族 | KOLDATA 記錄 `char[0x24]`（0…3 ＝ Human／Elven／Dwarven／Kelden） | KELDEN |
| 職業 | `sub_14307(char[0x1C] + 6, 槽位)` 回傳的記錄，靠右對齊到第 31 欄 | CLIFF GUARD |
| 位階（左） | KOLDATA 記錄 `0x74 + (char[0x29] / 5) × 5` | PEASANT |
| 位階（右） | `char[0x29] % 5 != 0` 時再畫 `0x74 + char[0x29]` | SERF |

身高與體重共用同一組 DGROUP 表，索引都是 `2 × char[0x1A] + char[0x24]`：

```
基準身高 = [DGROUP+0xB0 + idx]        ; 1C 10 1C 1C 1E 00 14
呎  = (char[0x1F] - 基準身高) / 12
吋  = (char[0x1F] - 基準身高) % 12    ; 印成 "%1d'%2d\""
基準呎 = [DGROUP+0xBC + idx]          ; 03 03 02 03 03 00 03
每呎磅 = [DGROUP+0xC8 + idx]          ; 41 32 4B 46 28 00 23
體重 = (呎 - 基準呎) × 每呎磅 + 吋 × 4  ; 印成 "%3d LBS"
```

Theodoric：idx ＝ 0×2＋3 ＝ 3，體型 116，基準 28 → 88 吋 ＝ 7 呎 4 吋；
體重 ＝ (7 − 3) × 70 + 4 × 4 ＝ 296。畫面上正是 `7' 4"` 與 `296 LBS`。

### 武器熟練度：四列，名字是 5-bit 壓縮

`0x15D2E` 起的迴圈跑 i ＝ 0…3，記錄位移 `0x4A + 13i`：

```
if char[0x4A + 13i] == 0: 跳過這一列
名字 = 5-bit 解碼(char[0x4A + 13i …])
OFF  = (char[0x53 + 13i] << 8 | char[0x54 + 13i]) / 10
DEF  = (char[0x55 + 13i] << 8 | char[0x56 + 13i]) / 10
```

**武器名的 5-bit 版面與城鎮文字不同**：第一個 byte 的高 nibble 是字數，
字元從第 4 個 bit 開始，每 5 個 bit 一個，1…26 對應 A…Z、0 是空白。
`assets/original-derived-data/weapon-templates.json` 的 14 筆記錄全部解得出來：
LONGSWORD、BATTLE AXE、LONG BOW、GREATSWORD、LT CROSSBOW、WAR HAMMER、
SCIMITAR、HALBERD、BROADSWORD、SELF BOW、MACE、SHORT SPEAR、LONG SPEAR、ELF BOW。
Theodoric 的第一列是 GREATSWORD、OFF 10、DEF 10（記錄裡是 100／100，除以 10），
第二列的第一個 byte 是 0，所以畫面上只有一列。

### 右下三欄與兩條長條

| 欄位 | 來源 |
|---|---|
| GOLD CROWNS | `sub_14CCF` 讀出來的 24-bit 錢包（RE-073），`% 9ld` |
| ADVENTURE PTS | `char[0x26] << 8 \| char[0x27]`，`%5u` |
| MOVEMENT RATE | `char[0x1BA]`（0 時當成 20），印出字元 `0x41 + 值` |
| 盔甲 | `char[0x1A0]` 非 0 時畫那個 Pascal 字串 |
| 盾牌 | `char[0x1A8]` 非 0 時畫，欄位是 `0x26 - (值 >> 4)` |

DAMAGE 與 NUTRITION 是兩條紅色長條，長度都是 `值 >> 3` 格（非 0 但不足一格
算一格），矩形寬 `8 × 格數`：DAMAGE 讀傷勢 `char[0x1B]`，NUTRITION 讀 `char[0x9F]`
——RE-073 的修道院治療正是把這兩個位移一起歸零。

### 圖示列

`0x1607E` 起：U-TURN（0Fh）、10h、6，再依條件加兩個——`char[0x200]` 非 0 加 12h，
`char[0x1DD]`…`char[0x1E1]` 任一非 0 加 13h。選 U-TURN 離開回城鎮，
其餘四個各自進另一個畫面（`sub_16986`、`sub_18D61`、`sub_166CC`、`sub_167D9`），
本輪沒有再往下追。

## RE-075：商店——商品表、明細欄位與買下之後東西放哪裡

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_shop.py`（`sub_1AAE8` 清單元件、
`sub_153FF` 商品明細、`sub_15720` 傷害／防護、`sub_1B8B8` 數字、`sub_157C6`、`sub_15896`）。
原版收據 `docs/verification/character-ui/town-shop-run1`／`run2`。
證據等級 `已證實`（商品表 35 棟建築全部解得出來，價錢與明細欄位逐位元讀出）。

### 商品表在建築 payload 的 0x100 起

TRADE 圖示由旗標 `payload[0x3C]` 的 bit1／bit4／bit5 給（RE-067），
按下去之後 `sub_11A93:0x1290B` 先看 bit5：**有 bit5 是馬廄**（走 `0x1291F`），
其餘走 `0x12D17` 的一般商品清單。

一般商品的表：

```
payload 0x100 起，一筆 16 byte，最多 32 筆；第一個 byte 是 0 就結束
  rec[0]  高 nibble 是名稱字數，名稱從第 4 個 bit 起用 5-bit 壓縮（同 RE-074）
  rec[8]  低 nibble 是價錢的高 4 位；高 nibble 是傷害／防護的加值（>7 是負的）
  rec[9]  價錢的低 8 位          → 價錢 = (rec[8] & 0x0F) << 8 | rec[9]
  rec[10] 負重（Enc）
  rec[11] 狀態（Cnd）／營養（Nut）
  rec[12] 低 2 位是類別：0 Item、1 Wield、2 Wear、3 Both
          bit0 另外決定顯示 Dmg 還是 Prt；/4 & 7 是骰面查表、/32 + 1 是骰數
  rec[13] bit1：單手／雙手
  rec[14] 高 nibble 是穿戴部位，低 2 位是防護百分比的檔次
```

馬廄的表另有版面：一筆 0x80 byte，名稱在 +0x70。城 1 的霍克羅夫馬廄讀出
DRAFT HORSE、LGT WARHORSE、MED WARHORSE、HVY WARHORSE 四匹。

八座城 35 棟商店的表全部解得出來，價錢也對得上手冊的武器比較：
匕首 80、短劍 96、長劍 128、闊劍 160、巨劍 400；毛皮帽 32、皮甲兜帽 45。

### 清單是招募同伴用的同一個元件

`sub_1AAE8(1, 0x0D, 0x56, 列數, 筆數, 緩衝區, 0x18)`：起點第 13 欄第 86 列、
一次八列、寬 24 格，捲動箭頭畫在第 `13 + 24 − 1 = 36` 欄、顏色 EGA 4（紅），
上箭頭碼位 0x5E、下箭頭 0x5F。**與 RE-072 的同伴清單逐項相同**——原版兩處共用這一支。
回傳值是 `(圖示編號 << 8) | 反白列`，圖示 0x0F（U-TURN）代表取消。

### 商品明細（`sub_153FF`）

選定之後把那 16 byte 複製到 DGROUP+0x9E28，再逐欄畫出來：

| 欄 | 記錄 | 內容 |
|---|---|---|
| 名稱 | — | 24 格內置中 |
| `Typ:` | 0xA3 ＋ 0xA4 + (rec[12] & 3) | Item／Wield／Wear／Both |
| `Hnd:` | 0xA8 ＋ 0xA9 + ((rec[13] & 2) >> 1) | 類別是 Wield 時才有：One／Two |
| `Pos:` | 0xAB ＋ 0xBB + (rec[14] >> 4) | 類別不是 Item 也不是 Wield 時的穿戴部位 |
| `Dmg:`／`Prt:` | 0xAC／0xAD | `rec[12]` 的 bit0 決定哪一個 |
| 數值 | — | 部位是 6 時印 `(rec[14] & 3) × 10 + 20` 加一個 `%`，否則印骰子範圍 |
| `Enc:` | 0xAE | `rec[10]` |
| `GC :` | 0xAF | 價錢 |
| `Nut:` | 0xB0 ＋ 0xB3 + rec[11] / 25 | 部位是 0x0A 且類別是 Item（食物）時 |
| `Cnd:` | 0xB1 ＋ 0xB7 + 門檻數 | 其餘情形；門檻是 `rec[11]` 超過 0x20／0x40／0x80 各加一 |
| `Fc :` | 0xB2 | 類別是 Wear 時的合身費用（`sub_157C6`） |

骰子範圍（`sub_15720`）是這樣算的：

```
骰數 = rec[12] / 32 + 1
骰面 = [DGROUP+0xD4 + ((rec[12] / 4) & 7)] + 1      ; 表是 02 03 05 07 09 0B 13 63
                                                     ; 加一之後就是 3 4 6 8 10 12 20 100
加值 = rec[8] >> 4；大於 7 就是負的 −(值 − 8)
印出 max(骰數 + 加值, 0) 、`-` 、骰數 × 骰面 + 加值
```

毛皮帽的 `rec[12] = 0x0A`：骰數 1、骰面 6、加值 0 → 防護 `1-6`。

### 買下之後

```
"Do you want this?"（記錄 0x31）→ 是
價錢 > 金幣: 說記錄 0x56「金幣不夠」
否則 金幣 -= 價錢
說記錄 0x32「Eat here?」；只有旗標 bit4 的店（酒館）才問，答是就當場吃掉
找背包空位：角色記錄 0xA0 起、一格 16 byte、共六格，第一個 byte 是 0 就是空的
全滿：說記錄 0xA2「Your backpack is full!」
否則把那 16 byte 抄進去
```

**背包因此是角色記錄 0xA0–0xFF 的六格**，每一格就是一筆商品記錄——
商店賣的東西與背包裡的東西是同一種版面。

### 買完之後的合身與穿戴

背包填好之後還有一整段（`sub_11A93:0x12FE5` 起），只有**會調整的店**
（旗標 byte 1 的 bit1）遇到部位 0…2（頭／軀幹／腿）的東西才走：

```
該有的尺寸 = [DGROUP+0x116 + rec[14] 的低 nibble]
           × [DGROUP+0x120 + 部位] × 角色體型(0x1F) / 100      ; sub_157C6
if 該有的尺寸 == rec[10]: 直接穿上                              ; 回 0
價錢 = [DGROUP+0xE0 + 0x12 × 部位 + 2 × 低 nibble]
說記錄 0x33「If you want this fitted」＋ 0x53 ＋ 價錢 ＋ 0x54
答應且付得起：扣錢、把 rec[10] 改成該有的尺寸、穿上
否則：說記錄 0x5D「I'll pack it for you.」，東西留在背包
```

三張表：

| 表 | 內容 |
|---|---|
| DGROUP+0xE0 | 合身價錢，[部位][低 nibble]：頭 5…80、軀幹 50…800、腿 30…480 |
| DGROUP+0x116 | 尺寸係數，用低 nibble 查：6 5 8 10 10 12 15 15 20 |
| DGROUP+0x120 | 部位係數：頭 1、軀幹 9、腿 6 |

穿上是 `sub_19A87(背包位移)`：把那 16 byte 與角色記錄 `0x100 + 16 × 部位`
那一格**對調**，部位 9（腰帶）例外，換到 0x30；部位 ≥ 10 就什麼都不做。
換完講記錄 0x5C「I'll put this on you.」。**穿戴槽因此是 0x100–0x19F 的九格加上 0x30。**

ddduhh 的體型是 105，毛皮帽的低 nibble 是 0、部位 0，
該有的尺寸 ＝ 6 × 1 × 105 / 100 ＝ 6，與記錄裡的 rec[10] ＝ 6 相同——
原版買完就直接穿上，收據 `town-shop-run2` 的 frame-200 正是那一句。

`sub_15896`（酒館「當場吃掉」）是 `char[0x9F] -= rec[11]`，夾在 0 以上，
再講記錄 0xC6。**0x9F 是飢餓度**：吃東西讓它下降，資料頁上那條 NUTRITION 就是它。

### TRAIN 的兩支：甲冑調整與鍛造

TRAIN 圖示由旗標 byte 1 的 bit1 與 bit3 給（`hi & 0x0A`，RE-067），按下去之後
`sub_11A93:0x1318D` 先看 bit1：

**bit1 是甲冑調整。** 掃過穿戴槽 0x100／0x110／0x120（頭、軀幹、腿），
對每一件跑一次合身計算：

```
if 每一件都合身: 說記錄 0x67「"You don't need to fit!"」，結束
每一件不合身的：說 0x3A ＋ (0xC7 + 部位) ＋ 0x3B
                「"The armor on your head needs to be adjusted!"」，各等一次空白鍵
講完問 0x33 ＋ 0x53 ＋ 總價 ＋ 0x54
答應且付得起：扣錢，把每一件的 rec[10] 改成算出來的尺寸
```

ddduhh 買下毛皮帽穿上之後按 TRAIN，原版說的正是記錄 0x67
（收據 `town-adjust-run1` 的 frame-202）——體型 105 算出來的尺寸 6 與記錄裡的 6 相同。
**不合身那一條路目前只有靜態證據**：原版存檔裡沒有尺寸不合的穿戴品，
沒辦法用正常玩家路徑走到，`"The armor on your` 與部位名之間有沒有空白也還沒對拍過。

**bit3 是鍛造**（`0x133CA`）：說記錄 0x68「"Show me your Ingot."」，
挑一件背包裡的東西，`rec[15] & 0x20` 不是錠就說記錄 0x69，
是的話用 0x3C ＋ 0x53 ＋ 價錢 ＋ 0x54 問要不要打。這一支還沒實作。

## RE-076：AWARD 的訓練——位階閘門、四門技能、數字欄與冒險點數

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_train.py`
（反查記錄 0x3E／0x3F／0x40／0x41／0x6B…0x70／0xD0 的呼叫點，並倒出
`sub_1B5C8` 文字欄、`sub_1B100` 等鍵、`sub_1CA6E`／`sub_1D01D` 游標）。
原版收據 `docs/verification/character-ui/town-train-run1`。
證據等級 `已證實`（整段流程在同一支函式裡讀得完，六張畫面已逐像素對拍）。

訓練沒有自己的函式，整段就在建築內部主迴圈 `sub_11A93` 裡，
從 `0x1375c` 到 `0x13ceb`。

### 先比位階：帳面與實得不同就去競技場

```
sub_14371()                     實得位階（同 RE-074 的計算）
角色 +0x29                      帳面位階
兩者不同 → 記錄 0x6B「"Go to the Arena!"」，結束
```

換句話說，**升階要先去競技場報到，訓練不會順便幫你升**。

### 學費從建築的兩個 byte 來

`byte_2FBFC << 8 | byte_2FBFD`（`0x13799`）就是這棟建築的價錢，
與客棧住宿、修道院治療共用同一組欄位（RE-073）。問句是
記錄 0x3E ＋ 0x53 ＋ 價錢 ＋ 0x54，答應之後比金幣（`word_30A6E`／`0x30A70` 是 32 位元），
不夠就記錄 0x56，夠就先扣錢再問記錄 0x6C「"In which skill?"」。
**學費在選技能之前就扣掉了**——後面任何一種擋下（技能太弱、太強、點數不夠）
都不會退錢。

### 四門技能在建築 payload 的 0x50 起

```
payload 0x50 起，一筆 0x0C byte，掃到 0x80 為止 → 最多四筆；rec[0] 是 0 就跳過
  rec[0]   5-bit 壓縮的名稱（mode 1，每個字第一個字母大寫）
  rec[8]   武器編號；高三位元（& 0xE0）非 0 才問攻擊還是防禦
  rec[0x0A] 這門課收的技能值下限
  rec[0x0B] 上限
```

清單用的還是 `sub_1AAE8`——與招募同伴（RE-072）、商店（RE-075）同一支元件，
第 13 欄第 86 列、一次八列、寬 24 格。

角色的武器熟練度在 `角色 +0x4A` 起、一筆 0x0D byte、掃到 0x7E 為止，同樣四列。
選定的那一門先在這四列裡找：**同武器編號的優先**，沒有就用第一個空列；
四列都滿又沒學過這一門，說記錄 0x6D「"You've learned enough!"」。

### 收不收這個學生，看的是畫面上那個值

`sub_14427(列 + 0x0A + 側 × 2)` 取出攻擊或防禦的內部值，除以 10 得到畫面值，
再與這門課的下限、上限比：低於下限說記錄 0x6E、高於上限說 0x6F。
兩句都由 `sub_172C1` 講，講完直接結束——學費已經收了。

### 「要幾點技能」是一格文字欄，不是圖示

問句是三段拼起來的：記錄 0x3F ＋ (0xCE ＋ 側) ＋ 0x40，下一列是記錄 0x41。
接著 `sub_1B5C8(0x2E3A, 2)` 開一格文字欄：

```
每次呼叫 sub_1B100(1) 等一個鍵
  a…z／A…Z／空白：一律收
  數字：只有 word_2DF58 不是 0 才收（訓練這一格設 1，所以收數字）
  收到就 sub_1C2DA 印出來，寫進緩衝區，筆尖往右一格
  退格／左方向鍵：印兩個空白退回去
  打滿 2 格等同按 Return
Return 才結束，末尾補 0
0x13bfb  ASCII 減 0x30 要落在 1…5，而且第二格必須是 0（只收一位數）
```

**欄尾那條橫槓不是字模。** `sub_1B100(1)` 每 5 拍把 `sub_1CA6E` 的參數
在 0x00／0xFF 之間反相，EGA 版 `sub_1D01D` 把那個 byte 直接寫進
`A000:((列 + 5 + 2 × word_2DF18) × 0x28 + 行)`——**筆尖那一格底下 8 px 寬、
1 px 高的一條線，會閃爍**。所以它比字模表第 64 格那個底線寬一個像素
（字模是 `0xFE`，游標是 `0xFF`）；remake 照著畫成一條 8×1 的線，
`count.png` 才對得起來。

### 點數與技能值

```
要付的點數 = 100 × 級數                      0x13c30
角色 +0x27（16 位元）是冒險點數；不夠說記錄 0xD0，結束
夠就扣掉，再把技能記錄的前 9 個 byte 複製進那一列
技能值 += 100 × 級數 / 10 ＝ 10 × 級數        0x13c7c
最後說記錄 0x70「"Let's train!"」
```

畫面值是內部值除以 10（同 RE-074），所以**一級就是畫面上的 1 點**，
一次最多五級、五百點。

### 對拍

`docs/verification/remake-parity/plan-town-train.json` 六對整頁 320×200，
除了原版自己畫的滑鼠游標之外逐像素相同：學費那一問、四門技能的清單、
攻擊還是防禦、空的數字欄、回顯 1 的數字欄、點數不夠那一句。
ddduhh 的冒險點數是 0，所以正常玩家路徑走到最後一句就是記錄 0xD0；
**加技能那一條路目前只有靜態證據與單元測試**，還沒有原版同狀態畫面。

## RE-077：野外的時刻、日曆、天候與換日

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_wilderness.py`
（`sub_297F2` 時刻推進、`sub_29987` 時段與天候、`sub_17095` 日期框、
`sub_2AAA8` 遭遇、`sub_2B0FF` 特殊地點，以及 DGROUP 的六張表）。
原版收據 `docs/verification/character-ui/world-clock-run1`／`run2`。
證據等級：時刻、日曆、日期框 `已證實`；天候的隨機漫步 `強推論`
（式子逐指令讀出，但沒有做到逐次重現原版的亂數）；遭遇 `未知`。

### 時刻是倒數的

`word_2DF40` 從 4200（`0x1068`，DGROUP+0x70 的初值，也是換日的重設值）往下數。
每走一步扣：

```
時刻 -= 2 × (地形成本 + 隊伍移動率)
```

地形成本表在 DGROUP+0x12B8，索引是地形碼（RE-070 的 0…7）：

| 地形 | DESERT | RIVER | SWAMP | HILL | ROAD | FOREST | SHORE | PLAINS |
|---|---|---|---|---|---|---|---|---|
| 成本 | 10 | 10 | 10 | 20 | 6 | 14 | 30 | 8 |

道路最省、丘陵與岸最貴，與說明書「公路最快、多石丘陵最難走、平原最好走」相符。
隊伍移動率取隊伍裡**最大**的那個（數字越大越慢，`sub_29004:0x290A6`…`0x290D7`），
角色欄位 `+0x1BA` 為 0 時代入 20。

式子由原版實測釘住：`world-clock-run2` 在平原上左右來回，第 15 次按鍵時刻
由 3416 降到 3360，跨過 DAWN 的下限 3367，標籤換成 MID MORN；第 14 次還沒跨。
平原 8 ＋ 移動率 20 得每步 56，只有這個值同時滿足「第 14 步不跨、第 15 步跨」。

**撞到不可通行的地形那一步不耗時間。** 原版擋下之後跳 `0x2958F` 把座標復原，
接著 `jmp 0x295DD` 跳過整支 `sub_297F2`。格內的次格移動則照樣耗時間——
「世界座標沒變」那一條路（`0x295C1`）仍然呼叫它。

### 時段是門檻查表

五段的下限在 DGROUP+0x12CC（五個 word）：3367、2527、1687、847、0，
對應 KOLDATA 記錄 0xD5…0xD9（DAWN／MID MORN／NOON／AFT NOON／TWILIGHT）。
時段是「第一個滿足 `時刻 >= 下限` 的那一段」（`sub_29987:0x29A07`）。
四個門檻把 4200 切成 833 ＋ 840 × 4。

### 換日

時刻扣成負的就換日（`sub_297F2:0x2982D`）：

1. 時刻重設成 4200。
2. 每名在隊上的角色，`+0x9F`（畫面上的 NUTRITION 條）**加上**當月的消耗量，
   上限 255。表在 DGROUP+0x12D6：`4 4 3 2 1 0 1 1 0 1 2 3`——一月最餓、
   六月與九月不餓，與說明書「嚴冬貧瘠、晚春與初秋最豐饒」相符。
   條子越長越紅，所以那個 byte 其實是「餓」不是「飽」。
3. 天候切到第 5 組，標籤印記錄 0xE2「CAMPING!」，畫一張圖，**停下來等一個鍵**；
   收掉之後天候設回第 1 組。
4. 日期推一天：`word_2FA2E`（日，從 0 起算）加一，等於當月天數就歸零並進位月份
   `word_2FBB4`，月份到 12 就歸零並進位年份 `word_30A8E`。每月天數表在
   DGROUP+0x12C0（`31 28 31 30 …`），**沒有閏年**。

### 日期框

`sub_17095` 畫兩行：月名（DGROUP+0x63A 起、一筆 4 byte 的 `Jan`…`Dec`）在 byte 欄 3、
日期（`word_2FA2E + 1`）在欄 4 接著印；`AD` 兩個字元在欄 2、年份（`word_30A8E + 0x44C`）
在欄 6。新遊戲的起點是 **1100 年四月一日**。

### 天候是每月各自有上下限的隨機漫步

只有**時段變換**時才擲（`sub_29987:0x29A83`）：

```
r = rand(0x80) − 1                      ; 0…127
上限 = 表[月] & 7、下限 = 表[月] / 8 & 7  ; 表在 DGROUP+0x1326
r < 表2[月]（DGROUP+0x131A）: 天候還沒到上限就 +1
否則 r < 100:                天候還沒到下限就 −1
表[月] > 0x7F 且天候 == 3:    再 +1
```

`表2` 是「轉壞」的機率門檻，冬天 40、盛夏 15——夏天最穩定。
`表` 把上下限包在一個 byte 裡：四月是 3（0…3）、一月是 139（1…3，而且
大於 0x7F 所以第 3 段還會再往上推一段，那是只有冬天看得到的第 4 組）。

### 遭遇：換時段時擲，怪物由地形與季節挑

`sub_29987:0x29A3E` 在**時段變換之後**呼叫 `sub_2A4E5`，那一支就是野外的隨機遭遇：

```
夜裡（天候第 5 組）表位再加 0x20
季節 = [0x1332 + 月]                     ; 0 冬 1 春 2 夏 3 秋
機率 = payload3[0x5DC + 季節 × 8 + 地形 + 夜裡加值] / 2
rand(100) >= 機率 → 這一次沒有遭遇
隊伍平均位階 = Σ 各隊員的實得位階 / 人數 + 1     ; sub_2A337
平均位階 < 11 且 rand(100) < [0x133E + 平均位階 − 1] → 走任務怪物那一條
否則最多試 0x1F 次：
  怪物 = rand(0x20) − 1
  payload3[0x61C + 怪物] & [0x1312 + 地形]  且
  payload3[0x63C + 怪物] & [0x1316 + 季節]  才收
```

`[0x1312]` 是八個地形的位元遮罩 `80 40 20 10 08 04 02 01`，
`[0x1316]` 是四個季節的 `08 04 02 01`。訊息是記錄 0xE7
「YOU HAVE BEEN AMBUSHED@BY]」＋ 數量 ＋ 0xE8「, LEVEL]」＋ 等級 ＋ 怪物名，
`world-clock-run1` 的 frame-220 就是這一張（1 隻 2 級 BINDERAK）。
全隊有坐騎時用記錄 0xE3「Attempt to ride away?」問要不要逃，
成功是 0xE4、失敗是 0xE5——與說明書〈伏擊〉一節相符。

**這一輪只讀出規則，還沒實作，也還沒對拍。** 怪物表本身（名稱、等級、圖號）
與遭遇之後的戰鬥都還沒解。

### 走到特定格子的兩件事

`sub_29004:0x29416` 比對隊伍座標與 `payload3[0x320]`／`[0x321]`＝(55, 13)，
命中就進 `sub_2AAA8`——那是**競技場與賭場**（記錄 0xEB「The Arena／The Tester」、
0xEC「A Gambling Booth／The Bet Taker」），不是遭遇。
`0x323`／`0x33B` 那 24 個特殊地點命中的是 `sub_2B0FF`，它用記錄 0xE9
「YOU ENCOUNTER THE FEARSOME CREATURES OF YOUR QUEST!」——任務怪物。兩支都還沒實作。

### 對拍

`docs/verification/remake-parity/plan-world-clock.json` 四對全過：
第 14 步與第 15 步的標籤框與日期框都與原版逐像素相同。
天候圖與地圖視窗列為**不宣稱**——換時段會重擲天候而 remake 的亂數不同源，
隊伍那個點還會閃爍。

## RE-078：野外的隨機遭遇——機率、怪物挑法、數量與等級

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_wilderness.py` 追加
`sub_2A4E5`（遭遇）與 `sub_2A337`（隊伍平均位階）；資料由
`tools/extract_monsters.py` 抽成 `assets/original-derived-data/monsters.json`。
證據等級：資料版面 `已證實`（32 隻怪物的名稱與兩張遮罩都讀得出來，
`world-clock-run1` 的 frame-220 對得上）；擲點流程 `強推論`（逐指令讀出，
但沒有逐次重現原版的亂數）。**尚未實作。**

### 什麼時候擲

`sub_29987:0x29A3E`：**時段變換之後**才呼叫 `sub_2A4E5`，一天最多四次。
走進某一格不會擲——那是另外兩件事（見下）。

### 機率表在 TOWN0 的 payload 4

`sub_29004:0x2904F` 把 `TOWN0` 的 payload 4 讀進 `unk_301A0`，
而它正好接在 payload 3 的緩衝區後面，所以程式裡寫成 `byte_2FBC4[0x5DC + …]`。
以 payload 4 自己的位移看：

```
payload4[季節 × 8 + 地形]        白天的機率
payload4[0x20 + 季節 × 8 + 地形]  夜裡（天候第 5 組）的機率
payload4[0x40 + 怪物]            怪物出沒的地形遮罩
payload4[0x60 + 怪物]            怪物出沒的季節遮罩
```

機率要**除以 2** 才是百分比（`rand(100) >= 機率/2` 就沒事）。春天白天的一排是
沙漠 25、河 60、沼澤 35、丘陵 50、道路 30、森林 50、岸 50、平原 40——
除以 2 之後河邊 30%、道路 15%，與說明書「走公路比穿野地安全」相符。
季節由月份查 KOLTOWN DGROUP+0x1332 的 `0 0 1 1 1 2 2 2 3 3 3 0`（冬春夏秋）。

### 怎麼挑怪物

```
平均位階 = Σ 各隊員的實得位階 / 人數 + 1        ; sub_2A337
平均位階 < 11 且 rand(100) < [0x133E + 平均位階 − 1] → 走任務怪物那一條
否則最多試 0x1F 次：
  怪物 = rand(0x20) − 1
  地形遮罩 & [0x1312 + 地形] 且 季節遮罩 & [0x1316 + 季節] 才收
```

`[0x1312]` 是 `80 40 20 10 08 04 02 01`（bit 7 是地形 0），
`[0x1316]` 是 `08 04 02 01`（bit 3 是冬）。試滿 0x1F 次都沒收就當作沒有遭遇。

怪物本身是 `TOWN0` 的 payload 5…36，一筆 512 byte：第 0 個 byte 是名稱長度、
名稱接在後面。三十二隻依序是 MINOTAUR、SKELETON、CYCLOPS、TROLL、ORC、ETTIN、
OGRE、GOBLIN、HILL GIANT、CLIFF TROLL、HOBGOBLIN、GREAT ORC、WALBAR、GREMLIN、
GOLEM、RUFFIAN、MUCK THING、LIZARION、BLOM、BINDERAK、SALAMANDER、MIST GIANT、
ZOMBIE、GHOUL、SYLPH、DJINN、STONE OGRE、SLEDGE、BRIGAND、BANDIT、THUG、MIST GRUB。
說明書〈周遭地形〉說「巨人族在靠近山區的丘陵、殭屍與 Djinn 藏身森林」，
與遮罩對得上。其餘欄位（屬性、武器、圖號）要等戰鬥解出來才有意義，這一輪不寫。

### 數量與等級

```
級距 = rand(0x63) / 0x14                        ; 0…4
等級 = 平均位階 + [0x1348 + 級距 × 2]            ; 五個 word：−2 −1 0 +1 +2
等級 < 1 就當 1
重擲直到 數量 < [0x1352 + 6 × 級距 + 人數 − 1]   ; 數量 = rand(0x0C) − 1
畫面上的數量是 數量 + 1
```

`[0x1352]` 那 30 個 byte 依級距分成五排，每排六格對應隊伍人數 1…6，
所以人多會遇到比較多隻。訊息是記錄 0xE7「YOU HAVE BEEN AMBUSHED@BY]」＋數量
＋0xE8「, LEVEL]」＋等級＋怪物名；數量大於 1 會補 `S`，名稱以 `S` 結尾則補 `ES`。

### 能不能騎走

`word_2FA2C`（隊伍最慢的移動率）等於 20 就是有人沒坐騎，直接開打。
否則問記錄 0xE3「Attempt to ride away?」，答應之後
`rand(100) > 5 × 隊伍移動率` 才成功（記錄 0xE4），失敗是 0xE5。
移動率越小（坐騎越快）越容易脫身。

### 走到特定格子的另外兩件事

`payload3[0x320]`／`[0x321]` ＝ (55, 13) 那一格進 `sub_2AAA8`：
競技場與賭場（記錄 0xEB／0xEC）。`0x323`／`0x33B` 那 24 個特殊地點進 `sub_2B0FF`，
用記錄 0xE9「YOU ENCOUNTER THE FEARSOME CREATURES OF YOUR QUEST!」。兩支都還沒解完。

## RE-079：戰鬥的入口、畫面與模組範圍

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_combat.py`
（`sub_1E3A8` 全函式倒出、兩層呼叫圖、模組用到的 UI 記錄）。
原版收據 `docs/verification/character-ui/world-quest-run1`。
證據等級：入口與畫面版面 `已證實`（原版實跑走到，畫面已擷取）；
模組內部規則 `未知`（這一輪只畫範圍，沒有解玩法）。

### 怎麼走到戰鬥

三個入口都呼叫 `sub_1E3A8`：伏擊（`sub_2A4E5:0x2A94A`，RE-078）、
競技場（`sub_2AAA8:0x2B0D5`）、任務怪物（`sub_2B0FF:0x2B352`）。
另有一處 `sub_10000:0x10D13`：城鎮地圖上按鍵碼 0x63（`c`）會直接開一場
固定 1 級 LIZARION 的測試戰鬥——**原版實跑按下去畫面沒有任何反應**，
那條路在這個版本不通，別再試。

**任務怪物是唯一可重現的入口。** `sub_2B0FF(i)` 先檢查隊上有沒有人帶著這個任務：

```
欄位  = 角色 [0x1E2 + 6 × word_2DF70 + i / 4]
除數  = [0x7C + (i & 3)]        ; 1 4 16 64——一個 byte 塞四個 2-bit 欄位
(欄位 / 除數) & 3 == 1 才算數
```

那個 2-bit 欄位有**四個**狀態：0 沒接、1 接了、2 完成（`sub_28F5D`，RE-101）、
3 回報過（`sub_11A93:0x11CC1`，RE-107）。

原版 `CHARDATA` 裡只有三名角色帶旗標：第 2 名 Poontang 的 `[0x1E3] = 4`
（對應第 5 個特殊地點 (48, 21)）、第 8 與第 10 名的 `[0x1E2] = 1`
（對應第 0 個地點 (49, 40)）。選 Poontang 出城之後往東八格、往北兩格
走到 (48, 21)，原版就會講那一段敘述並問「Will you partake?」，
答應之後畫出滿版的怪物圖與記錄 0xE9「YOU ENCOUNTER THE FEARSOME CREATURES
OF YOUR QUEST!」，再按一次空白鍵就進戰鬥。收據 `world-quest-run1` 走的就是這一條。

### 畫面版面

三個框由 `sub_1BA9D`（左上、左下）與 `sub_1BA9D`（戰場）畫出來：

| 區塊 | 矩形 | 內容 |
|---|---|---|
| 左上 | (16, 2)–(136, 76) | 角色名與裝備欄：記錄 0x101 `W:@H:@T:@L:@S:@B:@`（武器／頭／軀幹／腿／盾／腰帶）|
| 左下 | (8, 82)–(142, 157) | 人形剪影與右側的直條 |
| 戰場 | (151, 7)–(296, 152) | 地形磚塊與參戰者 |

指令圖示列在畫面底部，畫的仍是 `KOLMPICS[編號 + 3]`；出現過的編號有
0x2A…0x2E、0x30、0x35、0x37…0x3A、0x3C…0x40，以及 7（YES）與 0x0F（U-TURN）。

### 任務地點的記錄（`已證實`）

`sub_2B0FF` 讀進來的 `TOWN0` payload `0x25 + i` 是 257 byte：

```
+0x00        怪物編號
+0x01        等級
+0x02        數量
+0x03…0x26   36 byte 的站位：12 敵人 Y、12 敵人 X、6 隊員 Y、6 隊員 X
             （複製到 DGROUP+0x2E0C，戰鬥從那裡取）
+0x80 起八列  一列 16 byte 的敘述：第 0 個 byte 是字數，
             之後從第 8 個 bit 起每 5 個 bit 一個字（`sub_1B94F`）
```

第 5 個地點解出來是「12 隻 3 級 BINDERAK」，敘述是
`From the rim of one of / the hills surrounding / the valley below, you /
see the spidery forms of / Binderaks running about / in what once was the /
druid's keep.  Slowly, / you approach its walls.`——與 `world-quest-run1`
的 frame-862 逐字相同。講完之後問記錄 0xEA「Will you partake?」。

### 戰場（`強推論`）

戰場是 64×64 格的磚塊圖（4,096 byte 的緩衝區），視窗一次看 9×9 格、一格 16×16。
磚圖是 `KOLPICS` 的 payload 8（帶任務時 12），80 塊磚，版面與城鎮磚塊相同。
一般遭遇的地形是**現場亂數生成**的：每一格擲 `rand(4)`，中 1 就填 0x0D、其餘 0x09，
再擲 100 次 `rand(6)` 加總得到一個數，照那個數把隨機格子改成 0x0A。
任務地點走另一條路（`word_34968 = 1`），地形不重新生成。

參戰者站位：一般遭遇的隊員固定在 X `33 31 33 31 33 31`、Y `30 30 32 32 34 34`；
敵人查五組陣型表（`0x0B8C` Y、`0x0BD4` X，一組 12 格），陣型是 `rand(5) − 1`——
西、南、東、北、東西夾擊各一組。任務地點則直接用記錄裡的站位。

### 模組範圍

`sub_1E3A8` 有 4,389 條指令，直接呼叫 62 支函式，其中六支超過 500 條
（`sub_26D76` 901、`sub_2374C` 833、`sub_24131` 699、`sub_2818D` 602、
`sub_228B7` 535、`sub_289EE` 467）。整個模組用到 89 種 KOLDATA 記錄，
涵蓋命中部位（頭、胸、雙臂、雙腿）、傷勢輕重（serious／terrible／grievous）、
昏迷與擊退、法術、戰利品（冒險點數與金幣）、逃走與升級。

### 已讀出的資料結構（`強推論`）

```
參戰者陣列  DGROUP+0x6712，18 筆 × 50 byte
  +0x00  種類（隊員填 0x13）
  +0x01  0 我方、1 敵方、2 空位
  +0x1F  X
  +0x20  Y
戰場磚塊    4096 byte 的緩衝區，隨機填 0x09／0x0D，再灑一些 0x0A
起始站位    隊員 [0x0B60]／[0x0B66]、敵人 [0x0BD4 + 12 × 陣型]，陣型 = rand(5) − 1
怪物記錄    TOWN0 payload 5 + 怪物編號（RE-078）
```

規劃整輪做完問記錄 0x108「Begin next round?」；0x109「Do you want to flee?」
不接在它後面，而是**執行段走完、還沒分勝負時**由 `0x21045` 那一段的條件決定
要不要問（RE-103 訂正了本段最初的順序）。等鍵一律走 `sub_1B100(5)`，
而且會直接比對字母鍵（`y`／`Y`／`n`／`N`）。
指令碼放在 `word_31122`，訊息用 `|A`…`|S` 這類代換碼組出來。**玩法尚未解出。**

## RE-080：戰場的畫面、視線與回合結構

輸入與工具同 RE-070。新增探針輸出 `workplace/analysis/koltown/combat-view.json`、
`combat-fieldraw.json`、`combat-los.json`、`combat-blit.json`、`combat-attack.json`、
`combat-hit.json`、`combat-tables.json`（最後一份含逐塊倒出的 DGROUP 表）。
畫面證據是 `docs/verification/character-ui/world-quest-run1` 的 frame-885…906。

### 戰場視窗是 9×9（`已證實`）

`sub_223C0` 的兩層迴圈都以 9 為界，畫面位移是 `0x280 × 列 + 0x153 + 2 × 欄`：
0x280 是一列磚（16 掃描列 × 40 byte），0x153 = (152, 8)，因此視窗內容區
144×144 剛好九列九欄。之前把最下面 32 px 當成訊息區是錯的——那兩列是
視線之外的黑磚。

格子取法：`buffer[((Y + 欄 − 4) << 6) + (X + 列 − 4)]`，兩軸都夾限在 0…63
（`0x2250C`…`0x2253B`）。**執行期緩衝區是 RLE 解出來的陣列的轉置**：
`sub_2B3D1:0x2B436` 把 `decoded[k × 64 + bx]` 依序寫進去，而繪製又照
`(Y << 6) + X` 取回來，兩次相反的擺法抵銷，所以解出來的位元組本身就是
`(X << 6) + Y`——第一個索引是參戰者的 X，也就是畫面上的列。

任務地點的地形不重新生成，但**灑樹叢那一段照跑**：`0x1E58A` 用 `word_34968`
只擋掉底色生成（`0x1E594`…`0x1E5E2`），分支落在 `0x1E5E5`，也就是灑的起點。
灑法是擲 100 次 1…6 加總得到次數，每次挑 `rand(0x1000) − 1` 一格，
**只有原本是 0x09 或 0x0D 的格子會被換成 0x0A**（`0x1E63C`／`0x1E65F`）。

### 視線（`已證實`）

任務戰場一進去只看得到視線通得到的格子。`sub_223C0:0x22414` 先把 81 格的
可見度（DGROUP 0x2E3A 起 81 byte）全填 0x7F，再對每一格呼叫 `sub_25D7A(3)`：
從行動者走一條 Bresenham 線到那一格，線上每一格若還是 0x7F 就呼叫
`sub_25F8A(3, 步數)` 定值，定出 3 就停，後面的格子維持 0x7F。繪製時
0x7F 的格子畫磚塊 0x4A（全黑）。

擋不擋視線查 DGROUP 0x10CE 起十個 byte 的位元圖，取位元用 `byte_1312`
（`80 40 20 10 08 04 02 01`）：

```
0x10CE: ff 80 01 ff f1 ff 17 f9 76 7f
位元有設 ＝ 擋。例：草地 0x09、石板路 0x0E、磚路 0x24 不擋；牆 0x29、石板 0x2D 擋。
```

例外有兩條，都在 `sub_25F8A:0x2602C`…`0x26095`，只對磚塊落在 0x23…0x2D
這一段時成立：**記錄 +0x1E 非零**（`0x2606C`，騎著的人視線高，整條線都穿得過），
或者行動者**站在 0x24 或 0x25 上**而且是視線的第一步（`arg_2 == 0`）。
（本段原先把 +0x1E 那一條寫成「為 0 才穿得過」，方向相反，見 RE-102。）

拿任務地點 5 的地形照這條規則算出來的 9×9 可見度，與原版 frame-886
逐格相同：最下面兩列只看得到隊伍那一欄，左右被 (16,39)…(16,43) 那排石板擋住。
中心格本身不會被標記（走線長度 0），原版另外畫（見下）。

### 參戰者怎麼畫（`已證實`）

`sub_223C0` 每畫一格就掃一次參戰者陣列（`0x225BF`）：站在那一格的人依
`[0x6AE4 + 格]` 記下要畫的精靈，隊伍前六名是「索引 + 1」、敵人是 7，
倒下的敵人 8、倒下的隊員 9，地上的東西 0x27。畫的時候傳「精靈編號 − 1」，
也就是磚圖 atlas 的第 0…8 塊（1…6 號人形、敵人、倒下敵人、倒下隊員）
與第 0x26 塊（一把劍）。

疊圖是 `sub_1D21D`：把精靈四個平面 OR 成剪影，每一列當成一個 16-bit word
往左右各推一格（`s | s<<1 | s>>1`）加寬，取反當遮罩去挖地形，再把精靈 OR 上去。
**word 是小端序，低 byte 是左半邊**，所以位移在第 7／8 像素之間會跨 byte——
照像素做左右擴張會差幾個點，照 word 做則與原版逐像素相同（已用 frame-906
的中心格驗證：0 px 差）。

目前輪到的那一格另外處理：繪製迴圈把它當成看不到（畫黑磚），
`sub_223C0:0x2278A` 起再畫一次行動者，最後 `sub_1B6DC(0D8h, 48h, 0E5h, 57h, 0Fh)`
在 (216,72)–(229,87) 套一圈白框。

### 參戰者記錄（18 筆 × 50 byte，DGROUP+0x6712）（`強推論`）

| 位移 | 內容 | 出處 |
|---|---|---|
| +0x00 | 種類（隊員 0x13） | `0x1E78D` |
| +0x01 | 0 我方／1 敵方／2 空位 | 各處 |
| +0x02 | 等級：隊員是八個攻防值加總 ÷ 1000 ＋ 1（`0x1EE64`…`0x1EF18`），敵人是 `word_3B26A / 5 + 1`、超過 5 就設 5（`0x1EE12`…`0x1EE5D`）。**這一格與 `word_3B26A` 是兩個尺度**：技能折算與金冠用原始等級，規劃鍵值與 AI 權重用這一格（RE-097） | `0x1EE12`…`0x1EF18` |
| +0x07 | 隊員取角色 [0x1A] | `sub_2198A` |
| +0x08 | 力量 × (510 − 營養) ÷ 510 | `sub_2198A:0x219CF` |
| +0x09 | 快捷 | `0x21A39` |
| +0x0A | 洞察（角色 [0x21]） | `0x21A56` |
| +0x0B | 體型（角色 [0x1F]） | `0x21A73` |
| +0x0C | 智力 | `0x21A90` |
| +0x0E | 累積傷害（上限 0x7F） | `sub_277F4:0x27811` |
| +0x0F | 這一回合招架過幾次 | `sub_26D76:0x2723A` |
| +0x10 | 昏迷門檻 ＝ (體型 ＋ 角色 [0x20]) ÷ 2 | `sub_21AC4:0x21C1F` |
| +0x11／+0x12 | 頭：現值／上限 ＝ 門檻 × 15 % | `0x21C52` |
| +0x13／+0x14 | 胸：15 % | `0x21D55` |
| +0x15…+0x18 | 右臂、左臂：各 20 % | `0x21C8B` |
| +0x19…+0x1C | 右腿、左腿：各 30 % | `0x21CF0` |
| +0x1F／+0x20 | X（畫面的列）／Y（畫面的欄） | `sub_223C0` |
| +0x21 | 出手順序陣列裡的參戰者編號 | `0x1EF58`、`sub_2616C` |
| +0x22 | 這一回合行動過了 | `0x1FAD2` |
| +0x23／+0x24／+0x25 | 攻擊選項／防禦選項／指令碼 | `0x1F11B`…`0x1F14F` |
| +0x28 | 方向 0…7（上、右上、右…左上） | `sub_24BB6`、`0x1F2E7` |
| +0x2D | 疲勞或負重（乘 20 扣攻擊值） | `sub_26D76:0x26E3A` |

`sub_22866(i)` 回報「退出戰鬥」：頭（+0x11）或胸（+0x13）被打到 0xFF 就算。

### 回合結構：規劃與執行兩段（`已證實`）

一個回合分成**規劃**與**執行**兩段，各排一次順序。

**規劃**（`0x1EF34`…`0x1FAE7`）先算鍵值：

```
基準 = (等級 − 1) × 10 + 洞察                      ; 記錄 +0x02、+0x0A
若是隊員且 word_30D00 ≠ 0：基準 ÷= 2 × word_30D00  ; 第二回合起
傷勢 = 50 × 累積傷害 ÷ 上限                        ; sub_23683
基準 −= 基準 × 傷勢 ÷ 200
鍵值 = rand(100) + 基準                            ; 空位固定 0x7FFF
```

`sub_2616C` 是氣泡排序，鍵值由小到大；規劃**從索引 0 往後走**，
所以洞察低的先決定、洞察高的後決定。輪到誰就把指令碼、攻擊選項與防禦選項
寫進那一名的 +0x25／+0x23／+0x24：敵人由 `sub_2374C` 的 AI 填，
隊員畫戰場（`sub_223C0`）等玩家從圖示列選。一名決定完記 +0x22 = 1。

整輪決定完清掉左邊兩個框，印記錄 0x108「Begin next round?」，
圖示列是 YES（7）、U-TURN（0x0F），一般遭遇多一個 0x30
（`word_345DE == 0xFFFF` 才加）。選 U-TURN 跳回 `0x1EF3A` 重新規劃這一回合，
選 0x30 問記錄 0x109「Do you want to flee?」（直接比對 y／Y／n／N），
選 YES 才進執行段。

**執行**（`0x1FD21`…）另外排一次：

```
鍵值 = 快捷 − 快捷 × 傷勢 ÷ 200 − 負重 ÷ 10        ; 負重是 sub_24FD3 的 word_30C88
時間 = 拳腳(0x29) 8、移動(0x27) 6、射擊(0x23) 1，其餘查 [0x10B2 + 攻擊選項]
       （攻擊選項為 1 時再除以 2）
鍵值 −= (word_37CF0 ÷ 2) × 時間 ÷ 10
鍵值 += rand(100)，負數夾成 0
射擊時武器種類不是 0x12／0x16 就再除以 2；移動且攻擊選項為 0 也除以 2
```

排完**從陣列尾端往回走**（`word_2F972` 由 `word_2FBC0` 遞減），
也就是鍵值大的先動——快捷高的先出手。倒下或空位跳過；
輪到誰就把 +0x25／+0x23／+0x24 讀回來執行（`sub_26D76` 等）。

**逃走**（答 y）：離開戰鬥（`word_345DA` 與 `word_34966` 同時變成 0xFF），
收尾走賭金結算與掉東西那一段。條件與掉東西的規則見 RE-103——本段原先寫的
`((力量 + 智力) ÷ 2) × 等級 ÷ 3` 沒有對上位址，掉東西的機率實際上讀
`byte_3AC7D`（`0x21765`），而 `sub_2886C` 是名字比對不是擲骰。

### 指令列（`已證實`）

戰鬥的圖示列位置與城鎮相同：(32 + 32i, 171)，一格 24×20。七格逐像素命中
KOLMPICS 的 10、36、44、39、24、42、8，也就是命令碼（圖號 − 3）7、0x21、0x29、
0x24、0x15、0x27、5。**圖示內容看 frame-886…888 都一樣，但白框位置不同**：
剛進戰場時白框在第 0 格（frame-886），frame-887 與 888 已經各被 `>` 移過一格。
對拍要拿 frame-886 當基準（RE-082）。

| 命令 | 圖 | 意思 | 佐證 |
|---|---|---|---|
| 0x21 | 劍 | 武器攻擊 | `sub_26D76:0x27126` 印 0x130…0x139（whacks／chops…） |
| 0x29 | 拳 | 徒手攻擊 | `0x270F9` 印 0x128…0x12B（kicks／throws／butts／swings） |
| 0x23 | — | 射擊 | `0x270DF` 印 0x125…0x127（fires a missle…） |
| 0x27 | 走路的人 | 移動 | `sub_2374C:0x23B9B` 的 AI 選項 |
| 0x24 | 劍＋下箭頭 | 丟東西 | 記錄 0x10C／0x10D |
| 0x15 | 大劍＋下箭頭 | 撿東西 | 記錄 0x111…0x113 |
| 5 | 床 | 休息 | 記錄 0x119「is resting!」 |
| 7 | 大姆指 | 是／確定 | 與城鎮同一個圖 |

選了攻擊之後圖示列換成 KOLMPICS 18、45、58、59、60、61（frame-899），
也就是 U-TURN 加上 0x2A、0x37…0x3A 這五種攻擊式；命中部位另外用數字鍵選
（記錄 0x16C：「, 1 head, 2 chest, 3 right arm, 4 left arm, 5 right leg or 6 left leg.」）。
左下角那四個標籤 OPT／ATK／LOC／DEF／DIR 是記錄 0xFC…0x100。

方向 0…7 對到 3×3 的格子編號查 DGROUP 0x0B76（`01 02 05 08 07 06 03 00 04`），
也就是上、右上、右、右下、下、左下、左、左上；攻擊的目標格由
`sub_24BB6` 查 0x0FD8／0x0FE6 兩張指標表取位移再加上行動者座標。

### 指令列怎麼組（`已證實`）

圖示列是 `sub_11703(圖號)` 一格一格推出來的（推進 DGROUP 0x2DAE），
`sub_11722` 跑選單並回傳選中的圖號。掃過全部呼叫點之後，戰鬥用到的幾列是：

| 位置 | 圖號 | 用途 |
|---|---|---|
| `sub_228B7` | 7、0x22、0x23、0x21、0x29、0x0D、0x25、0x16、0x24、0x15、0x26、0x28、0x3F、0x27、5 | 主指令列，逐項看裝備與傷勢決定要不要放 |
| `sub_24131` | 7、0x0F | 選目標格：方向鍵移動標記，YES 確定、U-TURN 取消 |
| `sub_1E3A8:0x1F6BA` | 0x0F、0x2A…0x2E | 武器攻擊式，選中後查 [0x0B42 + 圖號 − 0x2A] 換成選項 0、1、2、5、8（RE-082） |
| `sub_1E3A8:0x1F7D9` | 0x0F、0x3C、0x3D、0x3E、0x3F、0x40、0x35 | 移動方式 |
| `sub_1E3A8:0x1F940` | 0x0F、0x2A、0x37…0x3A | 徒手攻擊式，選項 = 圖號 − 0x36（RE-082） |
| `sub_24AF8` | 0x0F、0x2A、0x30…0x34、0x36 | 防禦方式 |
| `sub_1E3A8:0x1FB38` | 7、0x0F、（0x30 只在一般遭遇） | 回合結尾的「Begin next round?」 |

原版 frame-887 的主指令列是 7、0x21、0x29、0x24、0x15、0x27、5，
逐像素命中 KOLMPICS 的 10、36、44、39、24、42、8（圖號 + 3）。

### 攻擊結算（`強推論`）

`sub_26AC2(行動者)` 先把武器的數字讀出來。隊員手上的武器識別在角色記錄 +0x3D，
拿它去比對四列武器（+0x52 + 13 × 列），對到就把那一列的攻防技能讀成
`word_34A12`（攻）與 `word_32380`（防）。武器欄位在角色記錄 +0x30
（徒手則是 +0x100 + [0x0FD2 + 選項]，射擊是 +0x70），欄位裡：

```
+0x08 基礎傷害      +0x0A 重量      +0x0B 耐久      +0x0C 傷害碼
+0x0D 種類          +0x0E 等級（盾牌用低四位）      +0x0F 旗標（bit 5 影響耐久扣法）
```

敵人走另一條：`sub_26C5F(0x17)` 讀怪物記錄的技能，再乘等級除以 5。

`sub_24FD3` 算負重（`word_30C88`）與盾牌等級（`word_34A0C`）：

```
係數 = 70 − (等級 − 1) × 5
重量 = 手上武器 +0x0A；word_37CF0 = 重量 × 係數 ÷ 100     ; 執行順序要用
重量 ÷= 2，再加上三件護甲（+0x100／+0x110／+0x120 的 +0x0A）
盾牌（隊員 +0x160、敵人 +0x60）存在、兩隻手臂都還在、而且武器不是雙手時：
    word_34A0C = 盾牌 +0x0E 的低四位；重量再加盾牌的 +0x0A
word_30C88 = 係數 × 重量 ÷ 100
```

攻擊分數與防禦分數（`sub_26D76`）：

```
攻擊值 = 記錄[+0x08] + word_34A12（武器攻擊技能）
攻擊值 −= 攻擊值 × 傷勢 ÷ 200 + 記錄[+0x0E] × 10 + 記錄[+0x2D] × 20 + 負重
攻擊分數 = 攻擊值 ÷ 10 + rand(100) + [0x110A + 14 × 攻擊選項 + 2 × 防禦選項]（16-bit 有號）

防禦值 = word_32380（武器防禦技能）÷ 這一回合招架次數 + 記錄[+0x09]
防禦值 −= 防禦值 × 傷勢 ÷ 200 + 記錄[+0x0E] × 10 + 負重
若有盾牌：防禦值 += 100 × 盾牌等級 + 200
防禦分數 = 防禦值 ÷ 10 + rand(100)
```

攻擊分數為負就是落空；攻擊分數 ≥ 防禦分數才命中，否則查
[0x11B2 + 11 × 防禦選項 + 攻擊選項] 取招架訊息 0x14F…0x158。

命中之後：

```
骰數 = 傷害碼 ÷ 32 + 1
骰面 = [0x00D4 + ((傷害碼 ÷ 4) & 7)]         ; 02 03 05 07 09 0B 13 63
傷害 = 基礎傷害(+0x08) + Σ rand(骰面) + 力量加成 + [0x1200 + 攻擊選項]
力量加成 = ((記錄[+0x08] − 70) ÷ 8) × (70 − 傷勢) ÷ 100
傷害 += [0x120C + (攻擊分數 − 防禦分數) ÷ 20]      ; 餘裕愈大加愈多
傷害超過上限就扣裝備耐久（`sub_22084`，扣光了那件裝備整個清成 0）並夾回上限
命中部位：擲 rand(100)，在 [0x121C + 6 × 列] 的累積機率裡找第一個 ≥ 的位置，
          列號 = [0x122E + 攻擊選項]（三列：0 是偏上半身、2 是偏腿）
```

`sub_277F4(傷害, 部位, 種類)` 把傷害記到 +0x0E（上限 0x7F），累積到 +0x10
就把頭與胸設成 0xFF（倒下）；部位欄位是 `+0x11 + 2 × 部位`，扣到 0 以下設 0xFF
（那一肢廢掉）。同時 `記錄[+0x2D] += 傷害`——挨打會讓下一次攻擊值變低。
訊息是 0x140「The」＋ 0x141/0x142（blow／spell）＋ 0x143 ＋ 0x144 + 部位，
再依傷害輕重接 0x14A…0x158。

### 還沒走完的

射擊專用的 `sub_268D4`／`sub_26D3C`、法術 `sub_2818D`、戰利品與升級 `sub_289EE`，
以及主指令列 `sub_228B7` 逐項的裝備條件。敵人那一手（`sub_2374C`）見 RE-081。

## RE-081：敵人那一手怎麼決定

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_enemy_ai.py`，輸出
`workplace/analysis/koltown/enemy-ai.json`：`sub_2374C` 全 833 條、十七支被呼叫者
逐條反組譯、呼叫端視窗 `0x1F0C0`…`0x1F170`、十二塊 DGROUP 原樣位元組與七個環的
位移陣列。

### 回傳約定（`已證實`）

`sub_2374C` 不碰參戰者記錄的決策欄位，它把結果放進三個全域，由呼叫端寫回
（`0x1F104`…`0x1F14F`）：

| 全域 | 寫進 | 意思 |
|---|---|---|
| `word_31122` | +0x25 | 指令碼 |
| `word_30A78` | +0x23 | 攻擊選項 |
| `word_30916` | +0x24 | 防禦選項 |

方向（+0x28）與射擊目標則由 `sub_2374C` 自己寫。呼叫的前提是 `+0x01 != 0`
（那一名是敵方），否則規劃段改畫戰場等玩家選（`0x1F155`）。

### 怪物記錄與傷勢（`已證實`）

`sub_22296(i)` 把 `word_39A48` 指到 `DGROUP + 0xCD92 + ((i − 6) << 7)`——執行期的
怪物記錄表，一筆 128 byte，索引是參戰者編號減 6（隊員佔 0…5）。開頭讀兩個
位元組當閾值：

```
word_322F8 = 怪物記錄[+0x0D]      ; 一般的忍受度
word_322FC = 怪物記錄[+0x0E]      ; 已經在撤退時的忍受度
```

`sub_23683(i)` 算傷勢：`word_2FB4A = 50 × [0x322E + 2i] ÷ [0x1C50 + 2i]`
（累積傷害 ÷ 耐力上限，0…50）。整支 AI 的分岔幾乎都在比這個值。

`sub_236AB()` 算行動餘力，兩個全域拿同一個值：

```
剩餘 = 記錄[+0x10] − 記錄[+0x0E]
word_30D06 = word_2FBB6
           = (上限 − 累積傷害) + 上限 ÷ 2 − 累積傷害 × 10 + 剩餘 × 2
```

### 同心環搜尋（`已證實`）

戰場的搜尋單位是以行動者為中心的方環。環號 `word_349B2` 為 0…6，格數在
`[0x0FC0]`＝`08 10 18 20 28 30 38`，位移陣列由 `[0x0FD8 + 2×環]`（X）與
`[0x0FE6 + 2×環]`（Y）兩張指標表指出去。環 0 就是八個相鄰方向，順序與 RE-080 的
方向編號相同：

```
環 0  dx −1 −1  0  1  1  1  0 −1
      dy  0  1  1  1  0 −1 −1 −1     ; 左上起順時針
環 r  起點 (−r, 0)，沿邊順時針走一圈，格數 8r
```

`sub_24BB6(環, 索引, 篩選)` 等於 `sub_24C35(篩選, X + dx, Y + dy)`：掃參戰者陣列，
找站在那一格上、通過篩選的人，命中就把編號寫進 `word_3237E` 並回 1。篩選是位元遮罩：

| bit | 條件 |
|---|---|
| 0x01 | 必須是我方（+0x01 == 0） |
| 0x02 | 必須是敵方（+0x01 != 0） |
| 0x04 | +0x1E 必須非零 |
| 0x08 | +0x1E 必須為零 |
| 0x10 | +0x05 必須為零，而且只在智力擲骰通過時才檢查 |
| 0x20 | 指令碼必須是 0x0D 或 0x23 |

一律排除空位（+0x01 == 2）、自己與已經退出戰鬥的人（`sub_22866`）。實際用到的
遮罩是 0x11、0x18、0x19、0x1A、0x31。

`sub_2531C(上一個索引, 篩選)` 在同一個環上往前找下一個命中格：第一次傳 0xFFFF
就從 `rand(該環格數 − 1) − 1` 起跳並記在 `word_345DC`，繞回起點回 0xFFFF。
**起點是隨機的**，所以同樣的戰況不會總是選同一個目標。

`sub_24E2B(環, 索引)` 判那一格能不能站：磚塊編號查 `[0x0FC8]` 起十個 byte 的
位元圖（取位元同樣用 `byte_1312`），位元有設就是擋路。這張與 RE-080 的擋視線
位元圖（0x10CE）是兩張不同的表：

```
擋路   0x0FC8: 00 10 9a ff f1 ff 17 ff d0 7f
擋視線 0x10CE: ff 80 01 ff f1 ff 17 f9 76 7f
```

### 智力擲骰（`已證實`）

`sub_252CB(值, 門檻)` 回報 `值 × 等級 ÷ 3 > 門檻`。開頭先擲一個
`word_3AC5E = rand(100)` 存起來，整支 AI 共用這一個骰：

```
sub_252CB(記錄[+0x0C] 智力, word_3AC5E)
```

通過才做戰術判斷（`sub_253E7`），也才會在挑目標時檢查 0x10 那一位。等級越高、
智力越高越容易通過——這是原版把「聰明的怪物」做出來的方式。

### 決策樹（`已證實`）

依序判，先成立的先出：

1. **兩隻手臂都廢**（+0x15 與 +0x17 都 ≥ 0x80）→ 直接跳到移動。
2. **撤退狀態**（+0x29 != 0）：傷勢 < `word_322FC` 就清掉旗標回到正常判斷，
   否則繼續逃。
3. **傷勢 ≥ `word_322F8`**：先看八個方向有沒有我方（`sub_24BB6(0, 方向, 0x19)`），
   有人貼著就跳移動、沒人就走撤退。
4. **無路可退**（+0x29 或 +0x2A 為零）→ 指令碼 5（休息），並記 +0x29 = 1。
5. **怪物記錄[+0x20] 為零**：+0x2B 非零就清掉去找方向，否則指令碼 0x26
   （攻擊選項 0、防禦選項 2）。
6. 其餘走**找目標**：智力擲骰過就先 `sub_253E7` 反制掃描；沒有結論就在環上
   找目標，可見（`sub_25D0F`）且環號 < 4 又有右臂就射擊（指令碼 0x23，並把目標
   座標記進 `[0x1A74 + 2i]`／`[0x1B36 + 2i]`、+0x1D = 1），否則貼近打（0x21）。

找不到任何目標時（`sub_2531C` 繞完一圈都是 0xFFFF）改走 `0x240B9`：從隨機方向
起繞八格，第一個能站又沒人的方向就往那裡移動。

`0x23CA7` 另有一條擾動：種類非零的怪物有 1/3 機率（`rand(3) == 1`）放棄攻擊改成
移動，所以敵人不會每回合都撲上來。

### 目標在哪一格 → 方向（`已證實`）

貼近打之前先用 `sub_25D7A(1)` 沿視線走一步，得到下一格座標，再查
`[0x10D8]`＝`07 00 01 00 06 00 02 00 05 04 03`：

```
索引 = (下一格Y − Y + 1) + 4 × (下一格X − X + 1)
```

九格解出來的方向與 RE-080 由 `0x0B76` 反推的定義（上、右上、右、右下、下、
左下、左、左上）完全一致，是獨立的第二個來源。那個方向站不了人就依序試
方向 −1、方向 +1，都不行才走隨機繞圈。

### 移動決定走多遠（`強推論`）

決定移動之後（`0x239D7`）先寫 +0x28 = 方向，再把三個基礎權重抄進暫存陣列
`byte_37CF8`／`37CF9`／`37CFA`（來源 `byte_3496A`／`3496B`／`3496C`），
然後照狀況打折：

```
兩腿都還在（+0x19、+0x1B 都 < 0x80）而且環號 < 7 → byte_37CFA = 0
否則：環號 < 2 → byte_37CFA ÷= 2
傷勢 > 49 → byte_37CFA ÷= 2
sub_251B0(0, 6)：word_2F940 = 6 × word_30C88（負重）
sub_2526C()：把 word_30D06 ÷ word_2F940 夾在 1/3…1
byte_37CFA = byte_37CFA × word_30D06 ÷ word_2F940，做兩次
```

三個權重加總（為零就當 1）之後，第 2 與第 1 項正規化成 `值 × 255 ÷ 總和`，
再擲 `rand(256) − 1` 做輪盤，選出 0、1 或 2：

```
word_31122 = 0x27（移動）
word_30A78 = 選中 + 0x3C          ; 移動方式圖號 0x3C／0x3D／0x3E
word_30916 = 選中 == 0 ? 2 : 0
```

三張基礎權重表的來源**已經由 RE-092 解出來**：戰鬥初始化時從 DGROUP `0x1092`／
`0x10A6`／`0x10AE` 複製過來，內容分別是 `6 6 5 4 6 4 4 6 6 5`、`1 3 4 4 4 4 2`
與 `12 24 72`。當初找不到是因為只搜了**具名參照**——實際的寫入是
`mov [bx+6A9Ah], al` 這種帶基底暫存器的形式，具名符號掃不到，要對位址做範圍掃描。

### 加權輪盤（`已證實`）

`sub_25A3A(n)` 是攻擊式與命中部位共用的抽樣器。權重放在暫存陣列
`[-0x61D8]`（DGROUP 0x9E28）的兩段：`tmp[i]` 是基礎權重、`tmp[0x10 + i]` 是
情境權重。

```
for i in 0…n−1:
    w = tmp[0x10 + i] 自乘「等級」次        ; w^等級
    w ÷= [0x10C6 + 等級]                    ; 01 01 02 04 08 10 20 40 ＝ 2^(等級−1)
    acc[i] = w
總和 = Σ acc[i] × tmp[i]（為零就當 1）
總和反覆 ÷ 255 直到裝得進 16 bit，同樣的次數套回每一項
r = rand(256) − 1；逐項扣 acc[i]，扣成負的就選中；掃完沒中回 0
```

`w^等級 ÷ 2^(等級−1)` 讓等級越高的怪物越偏向權重高的選項，低等級的接近均勻亂選。

`sub_2578F()` 抽武器攻擊式：基礎權重來自 `[0x6AD2]`，情境權重是
`[0x0FF4 + 10 × 防禦選項 + i]`，再乘上「這一式能不能用」——
`[0x109C + i] & 怪物記錄[+0x2D]` 有交集算 2、沒有算 1。`sub_2590F(攻擊式)`
抽命中部位：基礎權重 `[0x1A9A]`，情境權重同一張 `[0x0FF4]` 但列號換成攻擊式，
其中索引 < 2 或 == 6 的三項另外乘 `word_2F940 ÷ word_30D06`。
`[0x0FF4]` 是 16 列 × 10 byte 的權重矩陣，已原樣收在探針輸出裡。

`[0x6AD2]` 與 `[0x1A9A]` 兩張基礎權重表和 `byte_3496A` 同一個來源，
內容見 RE-092：`6 6 5 4 6 4 4 6 6 5` 與 `1 3 4 4 4 4 2`。

### 反制掃描（`已證實`）

智力擲骰過關才跑 `sub_253E7()`：繞著自己那一環，對每個找到的我方算一個優先度，
取最高的，並把對方的攻防選項抄過來。

| 優先度 | 情形 | 抄什麼 |
|---|---|---|
| 5 | 對方已規劃、要用武器攻擊（0x21）、而且方向對著我 | 對方的 +0x23、+0x24 |
| 4 | 對方還沒規劃 | 攻擊選項 0x0B、防禦選項 7；若對方指令是 0x21 且 rand(2)==1 就改抄對方的 |
| 3 | 對方已規劃、要近戰（0x21／0x29）、但方向不是對著我 | 攻擊選項 0x0B、對方的 +0x24 |
| 2 | 對方有防禦選項 | 攻擊選項 0、對方的 +0x24 |
| 1 | 對方要移動（0x27） | 同上，另外呼叫 `sub_25653` 側翼 |

「方向對著我」的判準是 `對方的 +0x28 XOR 我看到他的方向 == 4`——方向編號差 4
就是正對面。掃完一圈有結論就寫 `word_30A78`／`word_30916` 回 1，否則回 0
讓主流程繼續往下判。

`sub_25653(目標)` 是側翼：從自己現在的方向出發，依方向編號的奇偶決定要掃幾格
（奇數 5 格、偶數 6 格），找一個能站、又沒有敵方（遮罩 0x1A）的鄰格，把 +0x28
改過去，繞到目標側面。

### 還沒走完的

三張基礎權重表的來源**已由 RE-092 解出**；`byte_2EF75` 還沒定位。指令碼 0x26
的意思、`sub_25D7A` 各個 mode 的差異，以及參戰者記錄 +0x05／+0x1D／+0x1E／
+0x29／+0x2A／+0x2B 的語意也還沒解——AI 讀它們當旗標，但寫的人在別支。

> **這一段當初寫著「這幾項要靠執行期取樣，不是再多讀幾支函式能補上的」。**
> 那句話錯了：三張表就是再多讀一支函式（`sub_1E3A8` 的初始化那一段）補上的。
> 教訓寫成規則——**找不到 writer 時，先確認搜的是「具名參照」還是「位址」**。
> 帶基底暫存器的寫入（`mov [bx+6A9Ah], al`）不會出現在具名符號的 xref 裡，
> 要對位址做範圍掃描才看得到。

## RE-082：戰場上玩家那一手

輸入與工具同 RE-070。新增探針 `tools/ida_probe_koltown_battle_commands.py`，輸出
`workplace/analysis/koltown/battle-commands.json`：`sub_228B7`（535 條）、
`sub_24131`（699 條）、圖示列元件與四支條件判斷逐條反組譯，主指令列選完之後的
分派段（`0x1F3B0`…`0x1F6BA`）與三列子選單各一個位址視窗，另加八塊 DGROUP 位元組。

### 裝備記錄的基底（`已證實`）

`sub_22278(i)` 與 RE-081 的 `sub_22296(i)` 是一對，設同一個全域：

```
隊員（sub_22278）：word_39A48 = DGROUP + 0xBB88 + 0x300 × i     ; 角色記錄，一筆 768 byte
敵人（sub_22296）：word_39A48 = DGROUP + 0xCD92 + ((i − 6) << 7) ; 怪物記錄，一筆 128 byte
```

`sub_228B7` 另外用 `word_39A52 = 0x300 × i` 當位移，欄位一律以
`[word_39A52 + 位移 − 0x4478]` 取，也就是同一塊角色記錄。

### 主指令列逐項的條件（`已證實`）

`sub_228B7(行動者)` 先清空圖示堆疊，推 YES（7），數一下廢掉的手臂
（+0x15、+0x17 有幾個 > 0x80，記在 `word_2FBB2`），兩隻都廢就跳過所有武器類圖示。
之後逐項判：

| 圖 | 條件 |
|---|---|
| 0x22 | 有武器（[0x30] != 0）、傷害碼 bit0 為 1、種類 bit1 不與單臂衝突、種類高三位元為零、種類不是 6／0x0A／0x0B、+0x1D 為零 |
| 0x23 | 同上但種類是 6／0x0A／0x0B，或 +0x1D 非零；推完設 +0x1D = 1 |
| 0x21 | 種類高三位元非零而且 +0x1E 為零 |
| 0x29 | +0x1E 為零 |
| 0x0D | 沒有武器，而且 0x200…0x2F0 每 0x10 一筆之中有非零的 |
| 0x25 | 有武器、[0x190] 非零、種類高三位元非零 |
| 0x16 | 沒有武器或 [0x190] 為零，而且 [0x190] 非零 |
| 0x24 | 同上但 [0x190] 為零，且種類高三位元非零 |
| 0x15 | 有武器 |
| 0x26 | +0x1E 為零，而且地上東西的座標表（[0x2BAE]／[0x2D9E]，16 筆）有一筆與自己同格 |
| 0x28 | +0x1E 非零、腳下那一格能站（`sub_24E2B(0, 8)`）、而且沒有別人同格（`sub_227C0`） |
| 0x3F | +0x1E 非零時無條件 |
| 0x27 | +0x1E 為零，而且至少一隻腿還在（+0x19 或 +0x1B < 0x80） |
| 5 | 無條件，永遠是最後一格 |

`sub_24E2B(0, 8)` 用的是環 0 的第 9 項：環 0 的位移陣列除了八個方向之外還多一項
`(0, 0)`，所以索引 8 就是「腳下這一格」。`sub_227C0(i)` 掃全部參戰者找有沒有
別人與 i 同格且沒退出戰鬥。

### 選完之後的分派（`已證實`）

`sub_11722` 回傳選中的圖號，`0x1FA02` 起是分派表：

| 圖號 | 去哪 | 做什麼 |
|---|---|---|
| 5、0x15、0x16、0x24、0x25、0x26 | `0x1F3B2` | 攻擊選項與防禦選項都設 0，直接收尾 |
| 0x0D | `0x1F3C1` | 背包清單 → 用法清單 → 選目標格 |
| 0x21 | `0x1F6A3` | 選目標格 → 武器攻擊式 |
| 0x23 | `0x1F7A2` | 防禦選項 0 → 選目標格（mode 2） |
| 0x27、0x3F | `0x1F7C2` | 選目標格 → 移動方式 |
| 0x28 | `0x1F914` | 指令碼直接改成 0x28，選項 0／2 |
| 0x29 | `0x1F929` | 選目標格 → 徒手攻擊式 |

收尾都走 `0x1FA6D`：`word_37CF2`／`word_30A78`／`word_30916` 寫進 +0x25／+0x23／
+0x24。這是 RE-081 那三個全域的鏡像寫回點——敵人在 `0x1F104`、玩家在 `0x1FA6D`，
兩邊寫同樣的三個欄位。

### 三列子選單（`已證實`）

**武器攻擊式**（`0x1F6BA`，圖示 0x0F、0x2A…0x2E）：
`攻擊選項 = [0x0B42 + 圖號 − 0x2A]`，表是 `00 01 02 05 08`。選項為 1 時防禦選項
直接設 2 收尾；選項非 0 會先跑 `sub_24A02`（見下）；之後一律跑防禦式選單。

**徒手攻擊式**（`0x1F940`，圖示 0x0F、0x2A、0x37…0x3A）：
`攻擊選項 = 圖號 − 0x36`，負數夾成 0，值域 0…4。

這兩列的用途與 RE-080 原本的標註相反，判準有三個：分派表把武器（0x21）送到
`0x1F6BA`、徒手（0x29）送到 `0x1F940`；徒手的裝備欄查 `[0x0FD2 + 選項]`
（`00 50 60 00 40 10`，只有六筆），值域必須 ≤ 5，只有 `圖號 − 0x36` 裝得下；
而 `[0x0B42]` 給出的 0…8 正好是 `[0x110A + 14 × 攻擊選項]` 那張 128-byte 表的
九列。徒手的四個圖示（0x37…0x3A）也剛好對上 RE-080 已證實的四句徒手訊息
0x128…0x12B（kicks／throws／butts／swings）。

**移動方式**（`0x1F7D9`，圖示 0x0F、0x3C…0x3E、0x3F、0x40、0x35）分兩組：

```
+0x1E 為零           → 推 0x3C、0x3D；+0x03 非零再推 0x3E
+0x03 == 2 或 +0x1E 非零 → 讀腳下磚（sub_24E2B(0, 8) 順便設 word_3499A）
                          磚是 0x0C 或 +0x1E 非零 → 推 0x3F、0x40、0x35
攻擊選項 = 選中的圖號本身
防禦選項 = 選中 0x3C 時 2，其餘 0
```

防禦選項那一條與 RE-081 敵人 AI 的「輪盤選 0 就填 2、否則 0」獨立吻合——
敵人的三個候選就是 0x3C／0x3D／0x3E。

**防禦方式**（`sub_24AF8`，圖示 0x0F、0x2A、0x30…0x34、0x36）：
`防禦選項 = 圖號 − 0x2F`，負數夾 0，等於 7 時減 1。0x35 不在列上，所以
0x36 補回 6，值域 0…6。

**體型比較那一列**（`sub_24A02`，圖示 0x0F 加 0x41 或 0x2F／0x41／0x42）：
面前那一格沒有符合遮罩 0x18 的人就只推 0x41；有人就比體型
（目標 +0x0B − 自己 +0x0B），差距 < 100 才多推 0x2F，另外推 0x41、0x42。
選中之後 `攻擊選項 += 圖號 − 0x40`（0x2F 先換成 0x40）。

### 選目標格（`已證實`）

`sub_24131(mode)` 是游標互動迴圈，回 0xFFFF 表示取消。

```
mode 1（近戰、移動）：游標起點是面前那一格（方向查 [0x0B76] 換成 3×3 格），
                     夾在 3…5，而且不准停在中心 (4,4)——移到中心就再往同方向走一步
mode 2（射擊、用物品）：起點是上次記住的目標（[0x1A74 + 2i]／[0x1B36 + 2i]），
                     夾在 0…8，而且不准落在 2…6 的內圈——方向鍵移進去會把位移乘三
                     跳到外圈（之後不再夾），滑鼠點進去才還原成上一格（`0x24519`）
```

按鍵除了八個方向鍵（0x4800 上、0x4B00 左、0x4D00 右、0x5000 下與四個斜向）之外，
掃描碼 0xFF 是滑鼠：座標落在 X 0x130…0x250、Y 8…0x98 之內就換算成格
（`(X − 0x130) ÷ 0x20`、`(Y − 8) ÷ 0x10`）。`.`／`>` 與 `,`／`<` 移圖示列的白框，
空白鍵翻 `word_303FC` 重畫。

游標停在有人的格上會切換右側資訊框顯示的對象（`word_37D2A` 改成精靈編號 − 1，
編號 6 時查 `[0x1C90 + 格]`），`sub_2324F` 印那個人的資料——**敵人的細節要智力
檢定才看得到**（`sub_252CB(對方智力, word_3B264)` 不成立就只印簡略）。游標離開
有人的格就還原成自己。

Return 確定：白框不在第一格（`word_2DF3A != 0`，也就是選了 U-TURN）回 0xFFFF；
否則

```
mode 2：[0x1A74 + 2i] = 自己X + 列 − 4；[0x1B36 + 2i] = 自己Y + 欄 − 4，回 0
mode 1：+0x28 = [0x0B6C + 3 × (列 − 3) + (欄 − 3)]，回 0
```

`[0x0B6C]` 是 `07 00 01 06 08 02 05 04 03`，與 `[0x0B76]`（`01 02 05 08 07 06 03 00 04`）
互為反表，八個方向逐一往返一致，中心格對到 8（無效方向）。這是 RE-080 方向定義的
第三個獨立來源。

### 兩段死碼（`已證實`）

`0x247E7` 與 `0x2483F` 兩段都被前一條無條件跳躍跳過，永遠走不到：前者把索引 6
起的參戰者頭部設 0xFF 再結算，後者重畫所有參戰者。看起來是原版留下的除錯鍵
（按 1 清場、按 2 重畫），不進 remake。

### 還沒走完的

角色記錄的 +0x190、+0x200…+0x2F0（0x0D 的清單來源，一筆 0x10 byte、最多 16 筆，
名稱是 5-bit 壓縮）、參戰者記錄的 +0x03（移動方式的門檻）、+0x1D（射擊過了）、
+0x1E（不能動）、+0x26／+0x27（用物品選的物件與用法）語意都還沒定。
圖示 0x15／0x24／0x16／0x25／0x26 的條件已經讀出來，但條件與 RE-080 由 KOLDATA
記錄推出的語意（丟／撿）還沒對上，暫不改動那一份的標註。磚塊 0x0C 為什麼會
解鎖 0x3F／0x40／0x35 三個移動方式也還沒解。

## RE-083：主指令列的原版同狀態對拍

輸入與工具同 RE-070。remake 這一側是
`docs/verification/remake-parity/shots/world-quest/field.png`，由
`tools/run_kol_shots.sh` 依 `workplace/specs/quest-intro.txt` 重生；原版那一側是
`docs/verification/character-ui/world-quest-run1/frames/frame-886.png`。

### 白框初值在第 0 格（`已證實`）

`sub_11722` 先把每一格畫出來、框成亮藍（`sub_1B90B(格, 9)`），迴圈結束再單獨把
第 0 格框成白色（`0x1178D` 的 `sub_1B90B(0, 0x0F)`）。所以剛進戰場時白框一定在
最左邊那一格。

`sub_228B7` 傳給 `sub_11722` 的 `word_2EA1A + 4` 不是起始位置而是模式：
`sub_11722` 只比它等於 1（跳過選單迴圈）與等於 6（設 `word_2DF48`）兩種。

這一點決定了對拍的基準。原版那一輪的按鍵序列（`world-quest-run1` 的 checkpoint
108…110）是「進戰場 → `>` → `>` → Return」，收據記的畫格是**按鍵前**的穩定畫面，
所以 frame-886 才是剛進戰場、白框在第 0 格的那一張；frame-887 與 888 的圖示內容
相同，只有白框分別移到第 1、2 格。

### 七格逐像素相同（`已證實`）

Poontang 手上是 Long Spear（種類 0x9E、傷害碼 bit 0 為 1），四肢健全，穿戴欄的
第十格是空的。照 RE-082 的條件組出來是 YES、武器攻擊、徒手、丟、撿、移動、休息，
KOLMPICS 圖號是命令碼加三。圖示列整區 (32, 170) 216×22 共 4,752 個像素與
frame-886 **零差異**，含七個圖示、六個亮藍框與第 0 格的白框。

這一對用的是原版角色檔裡真實的裝備數字，不是反推出來的最小條件，所以它同時證實
了 RE-082 那幾條條件的取值方向。

### 順帶修掉的一個回歸

裝備 typed data 的抽取（`tools/extract_character_roster.py`）把 `equipment` 放進了
`sheet` 這一層，而 `internal/roster` 的 `Slot.Equipment` 是從槽位這一層讀。兩邊
對不上，所以 `Worn` 一直是十格全 nil——指令列因此只組得出空手的四格，SPEC-009 的
近戰結算也拿不到武器。抽取器已改成把 `equipment` 放在槽位層並重生 typed data。

**這個回歸不會被單元測試抓到**：規則層的測試自己餵 `Member`，走不到 JSON 那一段。
抓到它的是拿原版畫格對拍。

## RE-084：執行段的順序與那一手怎麼做出來

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_battle_commands.py` 的
`execution` 視窗（`0x1FD00`…`0x20900`，1,000 條）、`attack_target` 視窗
（`sub_26D76` 的開頭）與 `sub_26384`（移動，207 條）。

### 鍵值（`已證實`）

規劃整輪做完、玩家在「Begin next round?」答應之後，執行段**另外排一次順序**
（`0x1FD21`…`0x1FF4A`）。每一名先把自己的編號寫進 +0x21，空位跳過，其餘算：

```
sub_23683(i)                                 ; word_2FB4A = 傷勢
鍵值 = 快捷(+0x09) − 快捷 × 傷勢 ÷ 200
sub_24FD3()                                  ; word_30C88 負重、word_37CF0 武器速度
鍵值 −= word_30C88 ÷ 10
時間 = 指令 0x29 → 8
       指令 0x27 → 6
       指令 0x23 → 1
       其餘     → [0x10B2 + 攻擊選項]，而且攻擊選項 == 1 時再 ÷ 2
鍵值 −= (word_37CF0 ÷ 2) × 時間 ÷ 10
鍵值 += rand(100)
鍵值 < 0 就夾成 0
```

`[0x10B2]` 是 `00 0f 07 07 07 06 06 06 08 08 08 00 00 02 01 02`。

射擊（0x23）另外打折：讀射擊武器欄的種類（`word_39A48` ＋ 隊員 0x3D／敵人 0x2D），
**不是 0x12 也不是 0x16 就把鍵值除以 2**（`0x1FEDE`…`0x1FF06`）。

`sub_2616C` 排序是鍵值**小到大**，而執行迴圈從陣列尾端往回走
（`word_2F972` 由 `word_2FBC0` 遞減，`0x1FF5B`），所以**鍵值大的先動**——
快捷高、負重輕、指令快的先出手。

### 兩處原版的 dead code（`已證實`）

- 移動（0x27）且攻擊選項為 0 時，原版算了 `鍵值 ÷ 2`，但 `idiv` 之後
  **沒有寫回** `[0x2D78 + 2i]`（`0x1FF2F`…`0x1FF3E`）。對照射擊那一條
  （`0x1FEF2`）就有寫回。所以這一條折扣實際上不生效。
- 選目標格那一段的起點檢查（RE-082）是同一類：算了、判了，結果沒有用。

### 一手的分派（`已證實`）

執行迴圈取 `actor[尾端][+0x21]` 當行動者，跳過已退出戰鬥的與空位，把 +0x25／
+0x23／+0x24 讀回三個全域，然後：

| 指令 | 做什麼 |
|---|---|
| 0x27、0x3F | `sub_26384` 移動 |
| 0x15 | 撿：掃 `[0x2BAE]` 十六格找空的，把自己的座標寫進去；十六格都滿就印記錄 0x10C |
| 0x21、0x29 | 攻擊選項為 0 就把指令歸零，否則 `sub_26D76(1)` |
| 0x0B | 防禦選項設 2，印記錄 0x118 |
| 0 | 印記錄 0x117 |
| 5 | `sub_2625A()` 為零時印記錄 0x119「is resting!」 |

### 怯戰（`強推論`）

攻擊類的指令（0x21、0x23、0x0D、0x26、0x29 之一）在做之前多一道檢定
（`0x200AA`…`0x2016A`）：

```
條件：word_3AC58 != 0、行動者是我方（+0x01 == 0）、這一手是攻擊類、而且 +0x06 == 0
門檻 = ((力量(+0x08) + 智力(+0x0C)) ÷ 2) × 等級(+0x02) ÷ 3
rand(100) >= 門檻 → 指令改成 0x0B
```

門檻的式子與 RE-080 的逃走判定同型。推論等級是 `強推論`：`word_3AC58` 與參戰者
記錄 +0x06 的語意還沒定，只知道它們是這道檢定的兩個開關。

### 攻擊打到誰（`已證實`）

`sub_26D76(arg_0)` 開頭先取自己的方向（+0x28），`arg_0` 非零時用
`sub_24BB6(0, 方向, 8)` 看面前那一格有沒有人：

- 沒有：徒手（0x29）印記錄 0x11F、其餘印 0x120，整支就結束——**揮空**。
- 有：`sub_24BB6` 已經把那一名的編號寫進 `word_3237E`，那就是目標。

遮罩 8 只有 bit 3（+0x1E 必須為零），**沒有分敵我**——面前站著自己人也打得到。
之後才是 `sub_26AC2`（武器數字）、`sub_23683`（傷勢）與 SPEC-009 的整條數字鏈。

## RE-085：移動那一手

輸入與工具同 RE-070。`sub_26384`（207 條）由探針
`tools/ida_probe_koltown_battle_commands.py` 倒出。

### 一次只走一格（`已證實`）

執行段遇到指令 0x27 或 0x3F 就呼叫 `sub_26384`。整支函式只把座標加一次環 0 的
位移，**沒有點數、沒有步數迴圈**：

```
方式 = 攻擊選項(+0x23) − 0x3C          ; 圖號 0x3C…0x40 → 0…4；0x35 → 負 → 5
方式 > 2 → actor[+0x1E] = 1
方向 = actor[+0x28]
新X = actor[+0x1F] + [0x0C4E + 2 × 方向]
新Y = actor[+0x20] + [0x0C3C + 2 × 方向]
擋住了 → 印記錄 0x10A；沒擋 → 寫回座標、印記錄 0x10B
```

先前把「移動指令 0x27 的點數規則」列為待解（RE-080 末段、SPEC-008）是把移動方式
當成了距離。移動方式只有兩個作用：方式 3…5 會把 +0x1E 設起來，以及在執行段的
鍵值裡當時間成本（固定 6，不看方式）。

### 三種擋法（`已證實`）

依序判，任何一項成立就走不動：

1. **地形**：`sub_24E2B(0, 方向)` 說目的地那一格擋路。但 +0x1E 非零時有例外——
   目的地的磚塊編號（`word_3499A`，`sub_24E2B` 順便設的）落在 0x20…0x22、
   等於 0x2E、或大於 0x37 才真的擋，其餘擋路磚都穿得過去。
2. **腳下**：+0x1E 非零而且腳下那一格的磚是 0x0C 時走不動。
3. **有人**：`sub_24BB6(0, 方向, 遮罩)`，遮罩在 +0x1E 為零時是 8、非零時是 4。
   也就是**沒有 +0x1E 的人擋沒有 +0x1E 的人，有 +0x1E 的人擋有 +0x1E 的人**。
4. **界限**：新座標兩軸都要落在 4…0x3B。戰場陣列是 64×64，但可走的只有中間
   4…59，最外圈四格進不去。

### +0x1E 的行為（`強推論`）

把 RE-082 與這一支合起來，+0x1E 這個旗標的行為是：

| 狀態 | 指令列 | 移動方式 | 地形 | 誰擋你 |
|---|---|---|---|---|
| 為零 | 有 0x29 徒手、0x27 移動、0x26 | 0x3C、0x3D（+0x03 非零再加 0x3E） | 照 `[0x0FC8]` 位元圖 | +0x1E 為零的人 |
| 非零 | 換成 0x28、0x3F | 0x3F、0x40、0x35 | 大部分擋路磚穿得過去 | +0x1E 非零的人 |

解鎖後三種移動方式的條件是「腳下磚是 0x0C 或 +0x1E 已經非零」，而選了它們就會把
+0x1E 設成 1。行為上像騎乘：踩到某種格子才騎得上去、騎著能跨越地形、下來是另一個
指令。**但原版沒有任何文字或圖形證據支持這個命名**，所以 remake 只照行為實作，
欄位名沿用位移。

### 訊息（`已證實`）

`sub_2625A()` 回零時才印：走不動是記錄 0x10A、走了一步是 0x10B，兩者都印完等一個鍵
（`sub_1B100(5)`），走了一步還會重畫戰場再等一次。`sub_2625A` 的語意未定，
它在執行段的三個地方都當成「要不要印訊息」的開關。

## RE-086：訊息的代換碼

輸入與工具同 RE-070。`sub_1A8FD`（178 條，331 個呼叫點）與 `sub_265F3`
（226 條，只有一個呼叫點）由探針
`tools/ida_probe_koltown_battle_commands.py` 倒出。文字取自
`assets/original-derived-data/ui-text.json` 的 KOLDATA 395 筆。

### 兩個記號（`已證實`）

`sub_1A8FD` 逐字掃記錄時只認四種記號（`0x1A94D`…`0x1A9B3`）：

| 記號 | 處理 |
|---|---|
| `0x0D` | 記錄結束 |
| `\|` (0x7C) | 讀下一個字元交給 `sub_265F3` 代換 |
| `]` (0x5D) | 換成一個空白 |
| `@` (0x40) | 換行：X 回到 `word_2FB58`，Y 加 `2 × word_2DF18 + 6` |
| `~` (0x7E) | 後面三個字元是 `Vxx`／`Hxx`／`Cxx` 的參數 |

`]` 只是「保證這裡有一個空白」，它常出現在句尾（`The]`、`, and]`），
因為下一段文字會直接接在後面。

### 十一個代換碼（`已證實`）

`sub_265F3` 是一張 `cmp`／`jmp` 的分派表（`0x267F9` 起）。行動者是
`word_37D2A`、目標是 `word_3237E`，兩邊都用參戰者記錄的 +0x07（性別，
隊員取角色記錄 0x1A）選代名詞：

| 碼 | 代換成 | 出處 |
|---|---|---|
| `\|A` | 目標的所有格：`KOLDATA[0x110 − 性別]`（his／her） | `0x26618` |
| `\|B` | 行動者的所有格 | `0x26641` |
| `\|C` | 目標的受格：`KOLDATA[0x10E + 性別]`（him／her） | `0x2666A` |
| `\|D` | 行動者的受格 | `0x26691` |
| `\|E` | 行動者手上武器的名字：隊員讀角色記錄 +0x30、敵人讀怪物記錄 +0x20，5-bit 解碼 | `0x266B8` |
| `\|F` | 目標的名字：隊員直接印角色記錄開頭的 C 字串；敵人先印 `KOLDATA[0x140]`（`The]`）再印怪物名字表 `0xB967` | `0x26714` |
| `\|G` | 行動者的名字：`sub_2687F`，與 `\|F` 同一套規則，只是換成 `word_37D2A` | `0x2675D` |
| `\|H` | `[0x9D20 + 16 × word_30D02]` 的 5-bit 名稱 | `0x26765` |
| `\|I` | 行動者角色記錄 **+0x190** 那一格的 5-bit 名稱 | `0x26795` |
| `\|N` | `sub_1B8B8(word_30D02)`（還沒讀，看起來是印數字） | `0x267C4` |
| `\|S` | `[0x9E28]` 那一塊暫存區的 5-bit 名稱 | `0x267D3` |

代換之前，`word_345E0` 非零時會先補一個空白（`0x265FC`）。KOLDATA 裡每一句帶代換碼
的訊息都需要這個空白才讀得通（`kicks|B leg` → `kicks his leg`），所以 remake 一律補。

### 每一手開頭先印名字（`已證實`）

執行段每做一手都先呼叫 `sub_2685A`（`0x2005E`）：清一次（`sub_1C1A4`）、印行動者的
名字（`sub_2687F`）、再印一個空白。所以畫面上看到的整句是

```
<行動者的名字> <記錄展開之後的內容>
```

例如「Poontang is resting!」、「The BINDERAK moves!」。`sub_2687F` 與 `|F` 是同一套
規則：隊員直接印角色記錄開頭的 C 字串，敵人先印 `KOLDATA[0x140]`（`The]`）再印
怪物名字表 `0xB967`——差別只在讀的是 `word_37D2A`（行動者）還是 `word_3237E`（目標）。

### 順帶關掉一個未知欄位（`強推論`）

`|I` 讀的是角色記錄 +0x190，而 KOLDATA 0x114 是 `puts|B|I in|B belt`
（「把<某物>放進腰帶」）。RE-082 只知道 +0x190 會影響指令列上 0x16／0x24／0x25
三格要不要出現，語意未定；配上這一句，它是**腰帶上那一格裝備**的可能性很高。
仍標 `強推論`：沒有第二個獨立來源，而且指令列那三格的動作語意也還沒對上。

### 攻擊訊息怎麼組（`已證實`）

RE-080 記的鏈配上原文就完整了：

```
0x140 "The]"          → "The "
0x141 "blow" / 0x142 "spell"
0x143 " hits|F's]"    → " hits <目標名>'s "
0x144 + 部位          → head／chest／right arm／…
0x14A… 依傷害輕重     → " doing little damage!"…
```

其他常見的：`0x117 "does nothing."`、`0x118 "is frozen in terror!"`、
`0x119 "is resting!"`、`0x10A "is blocked from moving!"`、`0x10B "moves!"`、
`0x11F "staggers about."`、`0x120 "whips|B|E through the air."`
（「揮空」那一句會把行動者的武器名填進去）。

## RE-087：訊息框的版面

輸入與工具同 RE-070。`sub_1C1A4`／`sub_1C1BF`／`sub_1C2DA`／`sub_1C311`／
`sub_1103F` 與 `sub_1DE84` 的開頭由探針
`tools/ida_probe_koltown_battle_commands.py` 倒出。

### 先進緩衝區，收尾才排版（`已證實`）

訊息不是邊組邊畫的：

```
sub_1C1A4()      word_2E57C = 1（進緩衝）、word_345E0 = 0（長度歸零）
sub_1C2DA(字元)  緩衝模式下寫進 [0x1B60 + word_345E0++]，否則直接畫
sub_1C1BF()      離開緩衝，把整段斷行之後逐字畫出來
```

執行段每做一手先呼叫 `sub_2685A`（`sub_1C1A4` ＋ 行動者的名字 ＋ 一個空白），
接著 `sub_1A8FD` 把記錄的內容也送進緩衝，最後 `sub_1C1BF` 一次排版。
所以「行動者的名字」與訊息本體是同一段文字，斷行時一起算。

### 版面與斷行（`已證實`）

`sub_1C1BF` 開頭把游標設成固定值：

```
word_2DF14 = 3       ; X，byte 欄（`sub_1DE84:0x1DE9F` 的 shl ax, 3 → 像素 24）
word_2DF16 = 0xA8    ; Y，掃描列 168
word_2DF18 = 0       ; 0 是小字型 8×6，非零是大字型 8×8
sub_1103F(0x23)      ; 剪裁區設成欄 1…0x26、列 0xA4…0xC4
```

斷行的迴圈（`0x1C1FD`…`0x1C29F`）：

```
斷點候選 = 0x21 − word_2DF14        ; 第一行 ＝ 33 − 3 ＝ 30
while 斷點候選 < 長度:
    從候選往回找第一個空白當斷點
    印 起點…斷點−1
    word_2DF16 += 6；word_2DF14 = 3   ; 換行，行高 6
    起點 = 斷點 + 1，再跳過連續的空白
    斷點候選 = 起點 + 0x21            ; 第二行起是 +33，與第一行不對稱
印剩下的
```

第一行 30 個字元、之後每行 33 個，兩者不對稱——原樣照抄。整段找不到空白時原版
會一路往回掃出緩衝區，remake 改成硬斷在行寬處，不讀界外。

### 字元怎麼落地（`已證實`）

`sub_1C311` 依 `word_2DF1E` 選繪製器（0x0D、9、< 8 三種顯示模式），EGA 走
`sub_1DE84`：

```
像素 X = word_2DF14 << 3        ; byte 欄 × 8
像素 Y = word_2DF16             ; 掃描列
先 sub_1C40D(x, y, x + 8, y + 8) 清一塊 8×8，再畫字模
字模位址：word_2DF18 == 0 → 0x252E + 0x2E8 + (字 − 0x20) × 6      ; 小字型
          word_2DF18 != 0 → 0x252E + (字 − 0x20) × 8              ; 大字型
```

清的高度是 8 而行高是 6，所以相鄰兩行會蓋掉彼此頂端兩列；也因為是**逐字元格**
清，沒有字的格子會保留底下原本的畫面——戰鬥的訊息就這樣蓋在圖示列那一塊上。

### remake 的驗證（remake 自我一致，**不是** parity）

`workplace/specs/exec-msg.txt` 走完「進戰場 → 選休息 → Begin next round? → 執行」，
再用 `tools/read_text_from_frame.py` 從 (24, 168) 逐格比對原版字模，讀回
`Poontang is resting!`——字模、起點與行高都對得上。

**這不是同狀態對拍**：`world-quest-run1` 那一輪 oracle 的按鍵序列沒有走進執行段
（checkpoint 108…123 都在規劃與戰場之間），所以沒有原版畫格可比。要 parity 得
另錄一輪走到執行段的 oracle。

## RE-088：攻擊的訊息鏈

輸入與工具同 RE-070。`sub_26D76` 印訊息那三段（`0x270A0`…`0x27620`）與
`sub_277F4`（546 條）由探針 `tools/ida_probe_koltown_battle_commands.py` 倒出。
文字取自 KOLDATA 0x125…0x171。

### 第一句：動作與結果（`已證實`）

一次攻擊會印兩句，中間各等一次鍵。第一句是「誰對誰做了什麼、結果如何」：

```
動作（依指令碼）
  0x21 武器  → [0x12F + 攻擊選項]（whacks／chops／hacks…），再接 0x13A `|B|E at|F`
  0x29 徒手  → [0x127 + 攻擊選項]（kicks|B leg／throws|B body／butts|B head／
               swings|B fist），再接 0x127 ` at|F`
  0x23 射擊  → 武器種類 > 0x7F 印 0x126「throws a boulder」，否則 0x125
               「fires a missle」，再接 0x127

結果
  攻擊分數為負（落空）：取絕對值，> 0x1E 印 0x13B「 but misses|C completely!」、
      > 0x0A 印 0x13C「 but misses|C.」、否則 0x13D「 but narrowly misses|C.」
  被擋：印 0x15E「 who]」再印 [0x15F + 招架碼]（0x15F…0x168 十句）
  命中：印 0x13E「 who cannot parry the blow!」
```

**招架碼查的是 `[0x11B2 + 11 × 目標的防禦選項 + 攻擊選項]`，訊息基底是 0x15F**
——RE-080 原本記成「0x14F…0x158」，那一段是傷害那一句的記錄，不是招架。
另外有一條修正：招架碼算出來是 4 而且目標的體型減行動者的體型大於 100 時，
碼改成 2（`0x27417`）。

武器攻擊的選項值域是 {1, 2, 5, 8}（0 會在執行段被歸零），所以實際只會用到
0x130 whacks、0x131 chops、0x134 thrusts、0x137 slices 四句；0x12C…0x12F
在資料裡是佔位的 `f`。

### 第二句：傷害（`已證實`）

`sub_277F4(傷害, 部位, 種類)` 先把傷害記進去（累積上限 0x7F，累積到 +0x10 就把
頭與胸設成 0xFF 並記下「倒下」），再開一段新的訊息：

```
0x140「The]」＋ [0x141 + 種類]（blow／spell）＋ 0x143「 hits|F's]」＋ [0x144 + 部位]

傷害為 0            → 0x14A「 doing little damage!」
倒下                → 0x14B「 making|C]」，接著
                      sub_22229 為零 → 0x14C「collapse」＋ 0x14D「 from blood loss!」
                      怪物記錄 +0x10 > 0x7F → 0x14E「shatter into pieces!」
                      其餘 → 0x14F「die」＋ 0x14D
那一肢廢了（部位欄 > 0x7F）
    部位 > 1（手腳） → 0x152「 making that limb useless!」，並把目標的指令碼清成 0；
                      部位 > 3（腿）再把目標的 +0x03 也清成 0
    部位 0／1（頭胸）→ 目標是敵方印 0x151「 killing it instantly!」、
                      我方印 0x150「 making|C unconscious!」
其餘                → 0x153「 causing a]」＋ 輕重 ＋ 0x154「 wound!」
                      傷害 < 0x10 → 0x155 serious（並把傷害歸零）
                      傷害 < 0x18 → 0x156 terrible
                      其餘        → 0x157 grievous
```

腿廢掉會清 +0x03——而 +0x03 正是 RE-085 裡「移動方式第三格出不出現」的門檻。
兩邊合起來看，+0x03 與腿的狀態相關，這是 +0x1E／+0x03 那一組旗標的另一條線索。

### remake 的驗證（remake 自我一致，**不是** parity）

`workplace/specs/exec-hit.txt` 從任務地點走進戰場，往下移動四步貼到最近那一隻
BINDERAK（隊員站 (14,41)、那一隻站 (19,41)），再用武器攻擊式選項 1 打一下。
`tools/read_text_from_frame.py` 從 (24, 168) 讀回兩句：

```
第一句  Poontang whacks his Long Spear at The BINDERAK who cannot parry the blow!
第二句  The blow hits The BINDERAK's chest killing it instantly!
```

第一句三行、第二句兩行，斷點都落在空白上，第一行剛好 30 個字元——與 RE-087 的
斷行規則一致。動作句 0x130（`[0x12F + 1]`）、0x13A 的 `|B|E at|F`、命中的 0x13E，
以及第二句的 0x140／0x141／0x143／[0x144 + 1]／0x151 全部對上。

句首的「The blow」與 `|F` 展開的「The BINDERAK」都是 0x140，兩個 The 是原版的
行為，不是重複代換。

**這不是同狀態對拍**：`world-quest-run1` 那一輪 oracle 沒有走進執行段，沒有原版
畫格可比。

### 還沒走完的

`0x27B77` 之後是擊退與昏迷（用到 0x158「The blow stuns|F]」、0x159「The force of
the blow knocks|F]」與 0x15A…0x15C），還沒讀。法術那一套（0x169…0x171）同樣未讀。
倒下那一句的 collapse／shatter 兩個分支條件（`sub_22229` 與怪物記錄 +0x10）語意
未定，實作一律走 die；上面那一次是頭胸廢掉走 0x151，沒有經過這個分支。

## RE-089：執行段的原版畫面

新增 oracle 情境 `world-quest-execute`（`tools/dosbox_character_ui_oracle.py`），
輸出 `docs/verification/character-ui/world-quest-execute-run1`（969 個畫格）。
前綴沿用 `world-quest-site` 走到戰場那一段，接著讓 Poontang 在指令列上往右按到
最後一格、按 Return、答應「Begin next round?」，再連按二十次空白鍵推訊息。

**為什麼不照 remake 的路徑走**：原版的敵人會動（`sub_2374C` 的 AI），每一手都印
一句再等鍵，而 remake 的敵人那一手還沒實作（RE-081）。兩邊的按鍵數與訊息內容因此
對不起來，這一份要的是**訊息框的版面**，那與誰動的無關。

### 五件事（`已證實`）

1. **訊息區整條在執行段被換掉**。規劃段與「Begin next round?」時 (8,164) 起
   296×32 那一條還是紅磚底圖（frame-963／965 主色 `(170,0,0)`）；進執行段之後
   換成白底加一片亮黃的圖案（frame-966：白 5,488、黃 2,900、黑 864）。範圍與
   `sub_1103F(0x23)` 設的剪裁區（欄 1…0x26、列 0xA4…0xC4）完全吻合。
   **那張底圖還沒認出來**：不是 KOLMPICS 圖 2（那一張是紅底 296×32，把兩邊各自的
   主色當背景比形狀也只對得上 5,843/9,472）。
2. **文字是白底黑字**。字模格內非墨點是白色、墨點是黑色——與先前 remake 用的
   黑底白字相反。
3. **訊息整段轉成大寫**。KOLDATA 裡是 `moves!`、`does nothing.`，畫面上是
   `THE RUFFIAN MOVES!`、`POONTANG DOES NOTHING.`。與角色資料頁同一種處理
   （RE-074）。
4. **起點與行高與 RE-087 從反組譯讀出來的一致**：文字從 (24, 168) 起、行高 6。
   `tools/read_text_from_frame.py` 以原版字模逐格比對讀得回來，沒有偏移。
5. **移動成功要按兩次鍵**。`THE RUFFIAN MOVES!` 出現在 frame-966 與 967 兩張
   連續畫格上（中間按了一次空白鍵）——正是 RE-085 讀到的
   「印訊息 → 等鍵 → 重畫戰場 → 再等鍵」。

### 那一輪打的不是任務怪（`已證實`）

畫面上動的那一隻叫 **RUFFIAN**，而任務地點 5 的記錄寫的是 BINDERAK。這不是差異：
**那一輪根本沒走到任務地點**。frame-958 上寫著

```
YOU HAVE BEEN AMBUSHED
BY 1 LEVEL 1
RUFFIAN!
```

也就是路上被隨機遭遇攔下來了。收據的畫格編號也印證：從「往東第 43 步」起畫面就在
frame-961／962 之間反覆，後面那三十幾次方向鍵一步都沒走。挑怪的規則見 RE-090。

### remake 這一輪修掉的

白底黑字、整段轉大寫、整條訊息區換掉、移動成功等兩次鍵。訊息區底圖見 RE-090。

## RE-090：戰場上的敵人是誰

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_monster_name.py`
倒出 `sub_2687F`、`sub_2A4E5`、`sub_2B0FF`、`sub_1A4FE` 與四處把名字緩衝當參數
傳出去的位置。

### 名字只有一份（`已證實`）

`sub_2687F` 印行動者的名字（`|G` 走同一支，`|F` 換成目標）：

```
0x26888  word_37D2A < 6 ？
  是（隊員）→ sub_1A8FD(0, 0xBB88 + 0x300 × word_37D2A)   角色記錄開頭的 C 字串
  否（敵人）→ sub_1A8FD(0x140, 0x9E5C)   KOLDATA 的 `The]`
             sub_1A8FD(0, 0xB967)      名字緩衝區的第 0 筆
```

敵人那一邊**不帶索引**——一場戰鬥裡所有敵人共用 DGROUP+0xB967 這一份名字。原版一場
戰鬥只放同一種怪，這是結構上的保證，不是巧合。

### 那份名字從哪來（`已證實`）

`sub_1A4FE(fmt, 記錄號, ds:0xB966, word_34A0A + 0x400)` 把一筆記錄讀進
DGROUP+0xB966 這塊共用緩衝。全庫只有四處這樣用：

| 位址 | 場合 | `fmt` | 記錄號 |
|---|---|---|---|
| `0x1E412` | 戰鬥初始化（`sub_1E3A8`，`arg_0` 非零時） | `0x1844` | `word_3B262 + 5`，而 `word_3B262` 前一行硬寫 `0x11` |
| `0x2A65F` | 野外隨機遭遇（`sub_2A4E5`） | `0x1884` | 挑中的怪物編號 `+ 5` |
| `0x2B049` | `sub_2AAA8` | `0x1890` | `rand(16) − 1 + 5` |
| `0x2B2C5` | 任務（`sub_2B0FF`） | `0x189C` | 任務記錄的第 0 個 byte `+ 5` |

四處都把編號加五之後才當記錄號——`monsters.json` 的 32 個名字對應的是編號 0…31，
記錄 5…36。抽取器已經是按編號存的，`quest-sites.json` 的 `monster` 欄可以直接用。

### 任務戰鬥的怪物（`已證實`）

`sub_2B0FF(arg_0)` 的 `arg_0` 是任務地點編號 0…23：

- 旗標位址 `char[i][0x1E2 + 6 × word_2DF70 + arg_0 ÷ 4]`，除數
  `[0x7C + (arg_0 & 3)]` ∈ {1, 4, 16, 64}，`值 ÷ 除數 & 3 == 1` 才算這名角色接了
  這個任務（`0x2B108`…`0x2B16D`，掃六名角色，一筆 0x300）
- 任務文字 `sub_1A4FE(0x1896, arg_0 + 0x25, …)`
- 讀回來的緩衝：`+0` 是怪物編號、`+1` → `word_3B26A`（怪物**等級**，RE-097）、
  `+2` → `word_30C6C`（隻數）、`+3` 起 36 byte 複製到 `0x2E0C`、`+0x29` 起 16 byte
  複製到 `0xB956`。抽取器把這兩格分別存成 `quest-sites.json` 的 `level` 與 `count`
- 玩家答應（`sub_144EF`）之後 `word_3B262 = 怪物編號`，
  `sub_1A4FE(0x189C, 編號 + 5, …)` 換成怪物那一筆，再 `sub_2A975(編號)`

這一組數字與 `quest-sites.json` 的 `quest_byte`、`quest_divisor`、`payload`、
`monster`、`count` 五個欄位完全對上（地點 5：byte 483 = `0x1E3`、除數 4、
payload 42 = `0x25 + 5`），抽取器解對了。

### 野外隨機遭遇（見 RE-078）

`sub_2A4E5` 的挑怪規則 **RE-078 已經解出來了**，而且比這一輪讀到的完整：機率表在
`TOWN0` 的 payload 4、地形遮罩 `[0x1312]` 與**季節**遮罩 `[0x1316]`、數量與等級的
擲法、能不能騎走。這一輪只補一件 RE-078 沒有的事——挑中之後
`sub_1A4FE(0x1884, 怪物編號 + 5, ds:0xB966)` 把那一筆讀進名字緩衝，畫面上的名字
就是從這裡來的。

### 訊息區的底圖（`已證實`）

執行段鋪在 (8,164) 296×32 的那一張是 **MISCPICS 的圖 1**——一張卷軸，白紙面配黃色
的捲邊與軸心。拿 `world-quest-execute-run1` 的 frame-966 逐像素比對，9,472 個像素
中 8,986 個相同（94.9%），差的 486 個都是寫在紙面上的字。RE-089 當時比的
KOLMPICS 圖 2 只對得上 0.7%，KOLPICS 圖 2 是 3.8%。

## RE-091：一場戰鬥怎麼結束

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_battle_victory.py` 倒出
`sub_289EE`（467 條）、`sub_28F5D`、`sub_288D4`、`sub_26C5F`、`sub_22278`
與 `sub_1E3A8` 裡的七個窗。訊息記錄取自
`assets/original-derived-data/ui-text.json`。

remake 先前只判「隊伍全倒」，把怪物全部打死不會結束（SPEC-008 的已知限制）。
這一條把四個結束條件、兩種獎勵與三組訊息全部解出來。

### 每一輪之後數兩個數（`已證實`）

`0x20FA9`…`0x21042`，每個回合走完就掃一次全部參戰者：

```
倒下的隊員 = 0 ; 倒下的敵人 = 0
for i = 0 … word_2FBC0:              ; 參戰者總數
    if 參戰者[i][+0x01] == 2: 跳過    ; 這一格已經不算數
    if sub_22866(i) == 0:     跳過    ; 還站著
    if i < 6: 倒下的隊員 += 1
    else:     倒下的敵人 += 1
```

`word_345DA` 是倒下的隊員數、`word_34966` 是倒下的敵人數。**這兩個全域在別處被當
通用暫存重用**（例如 `0x2137D` 拿 `word_34966` 放金冠的骰數），讀碼時不能只看名字。

### 四個結束條件（`已證實`）

依序判斷，先中的先算：

| 條件 | 結果 |
|---|---|
| 倒下的隊員 ＝ `word_2DF2A`（隊伍人數） | 打輸 |
| 倒下的敵人 ＝ `word_30C6C`（敵人數） | 打贏 |
| 倒下的隊員 ＝ 0 | 繼續打 |
| `word_345DE` ≠ `0xFFFF` | 有人升級，先處理 |

`word_345DE` 是「這一輪升級的那個角色的記錄位移」，開場設成 `0xFFFF`
（`0x1E3D2`），`word_37D2A = word_345DE / 0x300` 換回角色索引。**誰設它還沒追。**

`word_345DA` 與 `word_34966` 同時等於 `0xFF` 是**逃走**那一條（`0x21592`）。

### 打贏（`已證實`）

`0x211CB` 起：

1. `sub_28F5D()`、清訊息，`word_3237E = 6`（把 `|F` 的目標指向第一個敵人）
2. 印記錄 **0x174**（`You have defeated|F`）
3. 敵人數 > 1 就補英文複數：名字最後一個字元是 `S` 就印 `ES`（並且
   `word_345E0 -= 1`，那是「代換前補空白」的開關），否則印 `S`
4. 印 `!`，等一次鍵

**冒險點數**（`0x21270`…`0x212CF`）：

```
總額 = Σ 每一個敵人 i：sub_26C5F(0x17) 之後的 word_34A12 + word_32380
每人 = 總額 ÷ 2 ÷ 隊伍人數
```

`sub_26C5F(0x17)` 讀的是怪物記錄 `+0x17` 與 `+0x19` 的兩個大端 16-bit，也就是攻擊
技能與招架技能（拆解見 RE-096）。用的是記錄裡的原值，不是讀技能時再乘等級的那一份。

結果放 `word_30A7A:word_30A7C`，同時複製到 `word_30D02`——那正是 `|N` 要印的數
（RE-086）。隊伍人數不等於 1 才印記錄 **0x175**（`Each player]`），單人跳過。

**金冠**（`0x21360`…`0x21413`）：

```
骰數 = (怪物記錄[+0x0F] ÷ 16) & 7
骰面 = [0x00D4 + (怪物記錄[+0x0F] & 7)]      ; 2 3 5 7 9 11 19 99
總額（32-bit）= Σ word_3B26A（敵人等級）輪：Σ 骰數 + 1 次：rand(骰面)
每人 = sub_2DDF0(總額, 隊伍人數)              ; 32-bit 除法
每人 < 2 就歸零
```

骰式只讀**第一隻怪**的（`sub_22296(6)`），同場其他怪的 `+0x0F` 沒有讀。外層用
`word_3B26A` 遞減當計數並把它減到 0；**那是敵人的等級**（RE-097），所以等級越高
金冠越多。

發錢（`0x2141C`…`0x21472`）：掃六名角色（步進 `0x300`，上限 `0x1200`），角色記錄
開頭非零才算數；每個人 `sub_14427(0x27)` → 加 `word_30A7A` → `sub_14465(0x27, …)`
寫回冒險點數，金冠則是 `word_30A6E:word_30A70 += 每人金額`。

最後 `word_30D02 = 每人金額`，印記錄 **0x177**（`|N gold crowns!`），等鍵，
呼叫 `sub_289EE(0)`。

記錄 **0x176**（`receives|N adventure points, and`）印在 `0x21344`，就接在隊員名字
那個迴圈之後。`已證實`

### 打輸與逃走（`已證實`）

打輸（`0x214DE`）：記錄 **0x140**（`The]`）＋ 敵人名字 ＋ 複數 ＋
**0x178**（`s have`）或 **0x179**（` has`）＋ **0x17A**（` defeated you!`）。
複數的判斷與打贏那邊相同。

逃走（`0x215A8`）：`sub_288D4(0)`、記錄 **0x17B**（`You flee for your]`），
隊伍人數 ＝ 1 印 **0x17C**（`life!`）、否則 **0x17D**（`lives!`）。

升級（`0x21611`）：`sub_288D4(4)`、`sub_2685A()` 印名字、記錄 **0x17E**
（`you have gained a level!`）、`sub_22278(角色)`、**角色**記錄 `[+0x29] += 1`、
`sub_289EE(1)`。

這條路要 `word_345DE != 0xFFFF` 才走得到，而戰鬥初始化把它設成 `0xFFFF`（`0x1E3D2`），
只有城鎮那三支會改。**野外打贏不會升級**，升級屬於城鎮那一側（RE-096）。

### 掉落物品（`已證實`）

`sub_289EE(0)` 的前半：

```
rand(100) >= 0x14 就什麼都不掉        ; 20% 機率
把 item.lst（DGROUP 0x185E 的檔名）載到 word_2FB5C 那個 segment
挑一隻怪：rand(敵人數)
掃 0x9D20 每 16 byte 一格，找第一個開頭為 0 的空槽
取 (rand(0x3D) − 1) × 16 那一筆，複製進空槽      ; 61 種物品
```

`0x9D20` 正是 `|H` 讀的那張表（RE-086），也就是訊息裡的「你找到了 X」。

掉落物**沒有座標**。它們不落在地上，而是直接進 `0x9D20` 這個池子；`sub_289EE` 的
後半就是把那 16 格排成清單讓玩家挑走（RE-096）。

### 還沒解的

- `sub_22866` 怎麼判定「能不能收下一件戰利品」（remake 目前自己用身體點數判倒下）
- `sub_288D4(n)` 那五個編號的語意（0 逃走／2 打輸／3 打贏／4 升級都出現過）
- 參戰者記錄 `+0x01` 等於 2 代表什麼

## RE-092：敵人 AI 的三張基礎權重表其實是靜態的

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_weight_tables.py`：對三張表的
位址做範圍掃描分出讀寫，並倒出 `sub_1E3A8` 的 `0x1EB80`…`0x1ED80`。

RE-081 把 `sub_2374C` 的決策樹解完之後，結論是三張基礎權重表
（`byte_3496A`／`3496B`／`3496C`、`[0x6AD2]`、`[0x1A9A]`）在解包後的靜態影像裡是
`0xFF`，「要執行期取樣才定得了值」。**那個結論是錯的。**BSS 的內容只有兩種來源
——程式碼寫進去，或從檔案載入——兩條都掃得到，而這三張表是戰鬥初始化時複製的。

### 三張表在戰鬥初始化時被填（`已證實`）

`sub_1E3A8` 的 `0x1EC49` 起，由 `byte_399D6` 選兩條路之一：

```
byte_399D6 ≠ 0 →  從怪物的載入緩衝取（0x1EC5F…0x1ECCC）
    for i = 0…9 : [0x6AD2 + i] = [0xB966 + 0x1A0 + i]     ; 攻擊式，10 筆
    for i = 0…6 : [0x1A9A + i] = [0xB966 + 0x1AA + i]     ; 部位，7 筆
    for i = 0…2 : [0x6A9A + i] = [0xB966 + 0x1B1 + i]     ; 移動三式，3 筆

byte_399D6 = 0 →  從 DGROUP 的靜態表取（0x1ECCF…0x1ED38）
    for i = 0…9 : [0x6AD2 + i] = [0x1092 + i]
    for i = 0…6 : [0x1A9A + i] = [0x10A6 + i]
    for i = 0…2 : [0x6A9A + i] = [0x10AE + i]
```

`0xB966` 就是 RE-090 那個共用載入緩衝——戰鬥開場 `sub_1A4FE` 把怪物那一筆讀進去，
所以「怪物自帶的權重」是那一筆記錄的 +0x1A0 起。

### 靜態預設值（`已證實`）

| 表 | DGROUP | 筆數 | 內容 |
|---|---|---:|---|
| 攻擊式基礎權重 | `0x1092` | 10 | `6 6 5 4 6 4 4 6 6 5` |
| 部位基礎權重 | `0x10A6` | 7 | `1 3 4 4 4 4 2` |
| 移動三式基礎權重 | `0x10AE` | 3 | `12 24 72` |

移動那一組 `12 24 72` 加起來是 108，比例約 1 : 2 : 6——**衝刺的權重最重**，
與原版敵人常常直接衝過來的印象相符。部位那一組 `1 3 4 4 4 4 2` 是頭、胸、
右臂、左臂、右腿、左腿的順序（RE-088 的部位表），頭最少、四肢均等。

### 那個旗標恆為零（`強推論`）

`byte_399D6`（DGROUP `0xBB06`）**全庫只有一個使用者**，就是 `0x1EC49` 那一次讀；
逐條掃過所有 operand 落在它附近的指令，**沒有任何 writer**。它落在 BSS
（靜態影像是 `0xFF`），而 DOS 的 C runtime 會把 BSS 清零，所以執行時應該恆為 0
——也就是**永遠走預設表那一條**，怪物自帶的權重走不到。

推論等級是 `強推論` 而不是 `已證實`：靜態掃描抓不到透過指標的間接寫入。要證實得在
執行期讀一次那個 byte。不過即使將來發現有 writer，預設表也還是其中一條真實路徑。

### 順帶：這一段還做了另一件事

`0x1EBA2`…`0x1EC46` 是**把載入緩衝複製成每一隻敵人的怪物記錄**：

```
for i = 6 … word_2FBC0:                     ; 每一個敵人
    序號 = i − 6
    前 32 byte ： [0xCD92 + (序號 << 7) + j] = [0xB966 + j]
    變體 = (rand(4) − 1) × 0x60 + 0x20
    後 96 byte ： [0xCD92 + (序號 << 7) + 0x20 + j] = [0xB966 + 變體 + j]
```

所以一筆 128 byte 的怪物記錄是「32 byte 共通 ＋ 96 byte 從四組變體隨機挑一組」。
同一場戰鬥裡每一隻的後 96 byte 可能不同——這解釋了為什麼原版同種怪的能力會有差異。

### 對 RE-081 的更正

RE-081 的「要執行期取樣」已經不成立，三張表都讀出來了。那一條的推論等級因此可以從
`強推論` 升到 `已證實`——決策樹本身早就是逐指令讀出來的，缺的只有這三張表。

## RE-093：執行段的分派只認十個指令碼

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_battle_items.py` 倒出
`0x2065E`…`0x20A80`。

### 十個，不是全部（`已證實`）

執行段的分派是一串 `cmp word_31122, n`，只比這十個值：

| 指令碼 | 做什麼 |
|---|---|
| `0x21`／`0x29` | 武器攻擊／徒手 → `sub_26D76(1)` |
| `0` | 印記錄 0x117（`does nothing.`） |
| `5` | 休息 → `sub_2625A` 回 0 才印記錄 0x119 |
| `0x28` | 騎乘 → `sub_227C0` |
| `0x22`／`0x23` | 投擲／射擊 → `sub_26D76(0)`，但先看要不要裝填 |
| `0x0D` | **施法** → `sub_2818D` |
| `0x27`／`0x3F` | 移動／下馬 → 第二輪掃描 `sub_26384` |

**`0x15`、`0x16`、`0x24`、`0x25`、`0x26` 一個都不在。**它們在**同一支的前一段**
（`0x20189`…`0x20621`）就處理掉了——`sub_1E3A8` 的「做出這一手」分成前後兩半，
前半是丟／撿／切換，後半才是這張分派表。規則見 RE-094。

### `0x0D` 是施法，不是背包（`已證實`）

`0x208ED` 對 `0x0D` 呼叫的是 `sub_2818D`，而 RE-081 已經把那一支認成法術
（訊息 0x16B 是 `casts|S at|F.`）。remake 原本把 `0x0D` 命名成 `CmdPack`
（背包清單），是從指令列那一格的圖號推的，推錯了。

### 射擊前要先裝填（`已證實`）

`0x20799` 起：讀角色記錄 `+0x3D`（手上武器的種類），**不是** 6／10／11 而且
`+0x1D` 為零時，這一手變成裝填——設 `+0x1D = 1`、印記錄 **0x15D**
（`places a bolt in|B|E!`）、把指令碼歸零，這一回合就沒了。下一回合再選射擊才真的射
（`+0x1D` 已經是 1，走 `sub_26D76(0)`）。

種類 6／10／11 那三種武器不用裝填，直接射。射完 `0x20860` 把 `+0x1D` 清回 0、
指令碼設成 `0x22`。

### 移動方式決定走幾格（`假說`）

`0x209AA` 起的第二輪掃描依攻擊選項（移動方式的圖號）呼叫 `sub_26384` 不同次數：

| 移動方式 | 呼叫次數 |
|---|---|
| `0x3C` 步行／`0x3F` | 0 |
| `0x3D` 奔跑 | 1（要 `+0x04` 非零） |
| `0x3E` 衝刺 | 1 ＋ 2（`+0x04` 非零時） |
| `0x35` | 1 ＋ 1 |

**這和 RE-085 的「一次只走一格」看起來衝突**，但兩者未必矛盾：RE-085 讀的是
`act` 那一輪，這是第二輪掃描，可能是同一格移動的兩段，也可能真的會多走。
`sub_26384` 一次走多遠、`+0x04` 是什麼，都還沒讀，所以整節標**假說**，
**不實作**。要解得先讀 `sub_26384`。

## RE-094：戰場上的丟、撿與切換

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_battle_pickup.py` 用訊息記錄號
反查處理點（比掃指令碼準：七個記錄各只有一個 `mov ax, <記錄>`），
`tools/ida_probe_koltown_battle_items.py` 倒出 `0x20150`…`0x20660`。

### 三個指令碼的語意（`已證實`）

remake 原本從指令列的圖號推語意，**三個都推錯了**：

| 指令碼 | 實際是 | 原本猜的 |
|---|---|---|
| `0x15` | **丟** | 撿（`CmdPickUp`） |
| `0x26` | **撿** | 地面（`CmdGround`） |
| `0x24`／`0x25`／`0x16` | **切換**（手上武器與腰帶武器對調） | 丟（`CmdDrop`）與兩個未知 |

判準是訊息：`0x15` 那一段印記錄 0x10D `drops|B|E!`，`0x26` 那一段印 0x111
`You have found a|H.  Do you want to pick it up?`，切換那一段印 0x114
`puts|B|I in|B belt`——而 `|I` 就是角色記錄 +0x190（腰帶）的名字（RE-086）。

### 戰場上的物品格

十六格，三張平行的表：

| 表 | DGROUP | 一格 | 內容 |
|---|---|---|---|
| 物品 | `0x9D20` | 16 byte | 物品記錄，第一個 byte 為 0 就是空格 |
| X | `0x2BAE` | 1 byte | 丟在哪一格的 X |
| Y | `0x2D9E` | 1 byte | Y |

`0x9D20` 正是 `|H` 讀的那張表（RE-086），也是 `sub_289EE(0)` 掉落物品放進去的地方
（RE-091）。

### 丟（`0x15`，`0x20189`）

```
掃十六格找第一個空的
全滿 → 印記錄 0x10C（cannot drop|B|E!），這一手就沒了
找到第 i 格：
    [0x2BAE + i] = 參戰者記錄[+0x1F]       ; 丟在自己腳下這一格的 X
    [0x2D9E + i] = 參戰者記錄[+0x20]       ; Y
    印記錄 0x10D（drops|B|E!）
    角色記錄 +0x30 起 16 byte（手上那件）複製進 [0x9D20 + 16i]
```

### 撿（`0x26`，`0x202CC`）

```
掃十六格：
    空格跳過
    [0x2D9E + i] 不等於自己的 Y → 跳過
    [0x2BAE + i] 不等於自己的 X → 跳過
    找到腳下這一格的東西：
        word_30D02 = i                     ; |H 要印的物品索引
        是隊員 → 印記錄 0x111 問要不要撿，等鍵；`n`／`N` 就不撿
        是敵人 → 直接撿
```

**只撿得到自己腳下那一格的東西**——X 與 Y 都要對上。

記錄 0x112（`|G, you must drop or belt your|E to pick this up!`）與 0x113
（`|G, this is the quest item!  Run for the exit!`）在同一段的後面，分別是「手上滿了」
與「這是任務物品」兩條分支，細節還沒逐條讀。

### 切換（`0x24`／`0x25`／`0x16`，`0x2050E`）

```
sub_22278(行動者)
for i = 0…15:
    交換 角色記錄[+0x30 + i] 與 角色記錄[+0x190 + i]     ; 手上 ←→ 腰帶
交換完：
    腰帶那一格非空 → 印記錄 0x114（puts|B|I in|B belt）
    手上那一格是空的 → 再印一個 `!`
    否則 → 印記錄 0x115
```

`+0x30` 是手上那件、`+0x190` 是腰帶那件，兩邊都是 16 byte 的物品記錄。
說明書的 `SWITCH`（交換備戰武器與腰帶武器）對得上。

### 還沒讀的

`sub_22278` 的語意（切換與升級那兩條都呼叫它）、0x112／0x113 兩條分支的條件、
撿起來之後物品放進哪一格。

## RE-095：擊退、打暈，以及戰場的兩套座標

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_battle_items.py` 倒出
`0x27B60`…`0x27E20` 與三張表。

### 擊退的判定（`已證實`）

攻擊命中、傷害訊息印完之後（`0x27B77`）：

```
體型差 = 行動者記錄[+0x0B] − 目標記錄[+0x0B]
rand(100) >= 體型差 → 什麼都不做
```

**機率就是體型差的百分比**：差 40 就是四成，同體型或更小完全推不動。

### 往哪個方向推（`已證實`）

先把「目標在我哪一邊」換成 `[0x123A]` 的索引：

```
索引 = (目標欄 < 我欄 ? 0 : 目標欄 == 我欄 ? 4 : 8)
     + (目標列 < 我列 ? 1 : 目標列 > 我列 ? 2 : 0)
```

`[0x123A]` 每三個 byte 一組，`rand(3) − 1` 從三個候選裡挑一個——**推的方向有隨機性**，
不是固定往正後方：

| 索引 | 目標在 | 候選方向 |
|---:|---|---|
| 0 | 正左 | 6 7 0 |
| 1 | 左上 | 5 6 7 |
| 2 | 左下 | 4 5 6 |
| 5 | 正上 | 0 0 0 |
| 6 | 正下 | 3 4 5 |
| 8 | 正右 | 0 1 2 |
| 9 | 右上 | 1 2 3 |
| 10 | 右下 | 2 3 4 |

索引 3、7 與 4（同格）用不到，原版就是 `0 0 0`。

### 兩條路（`已證實`）

挑到的方向那一格沒人（`sub_24BB6(0, 方向, 8)`）而且站得下（`sub_24E2B`）：

```
目標的欄 += [0x0C1C + 2×方向] ; 列 += [0x0C2C + 2×方向]
印 0x159（The force of the blow knocks|F]）
手上是空的 → 0x15C（backwards!）
否則      → 0x15A（backwards,）＋ 0x15B（and|A|H falls to the ground!）
```

推不動：

```
手上是空的 → 什麼都不印
否則      → 0x158（The blow stuns|F]）＋ 0x15B
```

兩條路最後都 `sub_223C0(目標)`（讓對手掉武器）並把目標的指令碼歸零——
**被擊退或打暈的人這一回合還沒做的那一手就沒了**。

### 戰場有兩套座標，差一個轉置（`已證實`）

這一條解開了一個一直沒對過的疑點。

| 表 | 加到 | 環 0 的值 |
|---|---|---|
| `[0x0FD8]`（環的 X 位移） | `+0x1F` | `-1 -1 0 1 1 1 0 -1` |
| `[0x0FE6]`（環的 Y 位移） | `+0x20` | `0 1 1 1 0 -1 -1 -1` |
| `[0x0C1C]`（方向的 X 位移） | `+0x1F` | `0 1 1 1 0 -1 -1 -1` |
| `[0x0C2C]`（方向的 Y 位移） | `+0x20` | `-1 -1 0 1 1 1 0 -1` |

同一個欄位、兩組不同的位移，所以**環索引與方向編號不是同一套**：
`環索引 = (6 − 方向編號) mod 8`。RE-081 寫的「環 0 的順序與方向編號相同」不成立。

由 `[0x0B6C]` 定的方向 0 是「上」（列減一），而 `[0x0C2C][0] = -1` 是加到 `+0x20` 的
——所以 **`+0x1F` 是欄、`+0x20` 是列**。`[0x123A]` 的候選方向逐項對得上（正上推 0、
正右推 2、右下推 3），是獨立的第二個證據。

**remake 的軸是反過來的**：`Actor.X` 是列、`Y` 是欄。戰場格子陣列在抽取時就轉置過
（RE-080），環位移表抄進 `ring0` 時也把原版的 X 欄放進了 `Row`——兩次對調抵消，
所以 remake 的 `ring0` 索引**等於方向編號**，畫面與移動方向都正確（四對戰場對拍
是它的證據）。

> 這種「兩個錯誤互相抵消」的狀態很脆弱：任何一邊被「修正」成與原版一致，另一邊
> 就會壞掉。`ring0` 與 `Cell` 的註解都寫明了，不要單獨改其中一個。

### 還沒讀的

`sub_223C0`（掉武器）與 `sub_269BA` 的細節、`sub_22229`（決定要看 `+0x30` 還是
`+0x20` 那一格）的語意。

## RE-096：戰後結算——冒險點數、金冠與戰利品

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_after_battle.py` 與
`tools/ida_probe_koltown_message_index.py` 倒出 `sub_1E3A8` 的 `0x211C0`…`0x21700`、
`sub_289EE` 的 `0x28B20`…`0x28D60`，以及 `sub_22296`。

### 先修正一個找錯的地方

RE-091 把冒險點數記成「`sub_26C5F(0x17)` 未解讀」，並猜結算在 `sub_289EE`。兩件事
都不對：獎勵整段都在 `sub_1E3A8` 裡（`0x21258`…`0x214B7`），`sub_289EE` 是它結束前
呼叫的**戰利品分配畫面**。

找法值得記下來：`sub_1A8FD(記錄號, 緩衝)` 是全庫唯一的訊息輸出入口，呼叫點前面必定有
一條 `mov ax, imm` 把記錄號推上堆疊。掃過全部 331 個呼叫點回推常數，301 個一次解出，
「哪一段程式印哪一句」就成了一張可查的表。要找某句玩家看得到的字是誰印的，不必再逐支
反組譯。索引在 `workplace/analysis/koltown/message-index.json`。

### `sub_26C5F` 讀的是什麼（`已證實`）

```
sub_26C5F(off):
    word_34A12 = (byte[p + off]     << 8) + byte[p + off + 1]
    word_32380 = (byte[p + off + 2] << 8) + byte[p + off + 3]
```

`p` 是 `word_39A48`，由 `sub_22278`（角色）或 `sub_22296`（怪物）先設好：

```
sub_22296(i) : word_39A48 = DGROUP + 0xCD92 + ((i − 6) << 7)
```

也就是**從當前記錄的 `off` 起讀兩個大端 16-bit**。`off = 0x17` 時讀的是攻擊技能與
招架技能——`extract_monsters.py` 早就把這兩欄抽成 `attack_skill`／`parry_skill` 了。

### 獎勵只在 `word_345DE == 0xFFFF` 時發（`已證實`）

`0x2124E` 的閘門。`word_345DE` 在戰鬥初始化（`0x1E3D2`）設成 `0xFFFF`，城鎮那三支
（`sub_2A4E5`／`sub_2AAA8`／`sub_2B0FF`）會設成別的值。不等於 `0xFFFF` 時整段獎勵
跳過，改走 `0x21611` 的**升級**路徑——那條路把 `word_345DE / 0x300` 當角色索引、
`記錄[+0x29] += 1`，再呼叫 `sub_289EE(1)`。

所以**升級不屬於戰後結算**，野外打贏不會升級。它是城鎮那一側的事，留給城鎮那一輪。

### 冒險點數（`已證實`）

```
總和 = 0
for i = 6 … word_2FBC0:                  ; 每一個參戰的敵人
    sub_22296(i)                          ; 指向第 i−6 隻的怪物記錄
    sub_26C5F(0x17)
    總和 += word_34A12 + word_32380       ; 攻擊技能 ＋ 招架技能
每人 = (總和 / 2) / word_2DF2A            ; 先除 2，再除隊伍人數
```

`word_2DF2A` 是隊伍人數（加入／離開時 ±1）。注意兩件事：**不看敵人死了沒**，參戰過
就算；用的是怪物記錄 `+0x17` 的**原值**，不是 `sub_26AC2` 讀技能時再乘等級的那一份。

發放在 `0x2141C`：對 `0xBB88` 起每 768 byte 一格、第 0 byte 非零的角色，
`sub_14427(0x27)` 讀出目前的冒險點數、加上去、`sub_14465(0x27, …)` 寫回。

### 金冠（`已證實`）

```
sub_22296(6)                              ; 只讀第一隻怪
b      = 記錄[+0x0F]
骰數   = (b / 16) & 7
面數   = [0x00D4 + (b & 7)]               ; 2 3 5 7 9 11 19 99
總額 = 0
while (word_3B26A-- != 0):                ; 敵人的等級（RE-097）
    for k = 0 … 骰數:                     ; 骰數 ＋ 1 顆
        總額 += rand(面數)
每人 = 總額 / 隊伍人數                     ; 32-bit 除法 sub_2DDF0
if 每人 < 2: 每人 = 0
```

**外層次數是 `word_3B26A`，不是敵人數**——那是敵人的等級（RE-097），等級越高金冠
越多；而且**骰式只讀第一隻怪的**，同場其他怪的 `+0x0F` 根本沒讀。

發放同樣在 `0x2141C` 的迴圈裡，`word_30A6E:70` 是角色的金冠（32-bit）。

### 那三句話怎麼排（`已證實`）

```
0x174  "You have defeated|F"      ; |F 展開敵人名字
       敵人數 > 1 → 名字結尾是 'S' 補 "ES"、否則補 "S"
       '!'
       隊伍人數 ≠ 1 → 0x175 "Each player]"
       隊伍人數 = 1 → 逐一印那一位的名字（記錄號 0，緩衝指到 0xBB88 + 768i）
0x176  "receives|N adventure points, and"   ; |N 讀 word_30D02 = 冒險點數
0x177  "|N gold crowns!"                    ; |N 讀 word_30D02 = 每人金冠
```

先前 `RewardRecords` 因為算不出冒險點數而只報金冠，現在兩個數都有了。

### 戰利品分配（`已證實`）

`sub_289EE(0)` 在報完獎勵之後被呼叫（`0x214B2`）。它前半產生掉落物、後半是玩家介面：

```
掉落：件數 = rand(敵人數)
      每件挑一個空的物品格，複製 16 byte 進去，
      再把 +0x0B（狀況）覆寫成 rand(255)

分配：掃 0x9D20 的 16 格，非空的收成一張清單
      清單是空的就直接結束
      畫框 (4,7)-(0x80,0x98)，印 0x181 "~H02~V09Items Found:"
      每一格用 sub_1B94F 展開成名字，接 0x0D 排成選單
      sub_1AAE8(…) 回傳值：低 4 bit 是選第幾項，高 byte 是狀態
      狀態 0x0F 且 word_2DF20 = 0 → 取消，結束
      掃六名隊員，記錄 +0x01 非零且 sub_22866(i) = 0 的收成候選
      候選 0 個 → 結束；1 個 → 直接給；多個 → 印 0x182 "~H02~V09Characters:" 再選一次
```

**掉落物沒有座標**——先前找不到誰寫 `0x2BAE`／`0x2D9E` 是對的，因為它們根本不落在
地上。`0x9D20` 那 16 格在戰鬥中是地面物品、戰鬥結束後就是戰利品池，所以打到一半丟在
地上沒撿的東西，戰後會一起出現在這張清單裡。

### 還沒讀的

`sub_22866`（隊員能不能收下一件戰利品的判準）、`sub_1B94F`（把物品格展開成名字）、
`sub_1AAE8` 的完整選單協定。

## RE-097：隨機遭遇的擲點流程逐條核對，以及金冠輪數的來源

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_ambush.py` 倒出 `sub_2A4E5`
（381 條）、`sub_2A337`，以及 `sub_1E3A8` 的開頭與六張 DGROUP 小表。

RE-078 是靠讀碼推出來的（`強推論`），實作之前逐條再對一次。**流程完全吻合**，
機率、遮罩、試 0x1F 次、級距與數量上限都對得上，等級可以升到 `已證實`。

### 三張小表（`已證實`）

```
0x1312  80 40 20 10 08 04 02 01           ; 地形位元，bit 7 是地形 0
0x1316  08 04 02 01                        ; 季節位元，bit 3 是冬
0x1332  00 00 01 01 01 02 02 02 03 03 03 00 ; 月份 → 季節
0x133E  21 21 21 16 16 16 11 11 11 11      ; 平均位階 1…10 走任務怪那一條的機率
0x1348  FFFE FFFF 0000 0001 0002           ; 級距 → 等級修正（−2 −1 0 +1 +2）
0x1352  2  4  6  8 10 12                   ; 級距 0，隊伍人數 1…6 的數量上限
        2  3  4  5  6  7                   ; 級距 1
        1  2  3  4  5  6                   ; 級距 2
        1  1  2  2  3  3                   ; 級距 3
        1  1  2  2  3  3                   ; 級距 4
```

等級越高（級距越大）遇到的隻數越少，人多則越多。

### `word_3B26A` 是怪物的等級（`已證實`）

這一條推翻了 RE-096 對金冠的一個說法。`sub_2A4E5:0x2A7D1` 在進戰鬥之前
`word_3B26A = 等級`，`word_30C6C = 數量`、`word_345DE = 0xFFFF` 也在同一段設好，
然後才呼叫 `sub_1E3A8(0)`。

`sub_1E3A8` 開頭是一個閘門（`0x1E3B1`）：**參數為 0 就跳過整段預設值**
（`0x1E3BA`…`0x1E417`：怪物 0x11、一隻、等級 1、地圖 `rand(3)−1`）。遭遇與任務
兩條路都傳 0，所以那個 `word_3B26A = 1` 落不到玩家身上——**金冠的外層輪數是敵人的
等級，不是常數 1**。等級越高，同一種怪掉的金冠越多。

remake 的 `GoldReward` 因此改吃 `Actor.Level`。先前實跑量到的 25 與 43 枚是 1 級怪，
一輪，數字不受影響。

### 騎走與任務怪那兩條岔路（`已證實`）

```
var_6 = (rand(3) − 1) × 2                    ; 0、2、4：那一隻的圖要畫哪一張
if word_2FA2C != 20:                          ; 20 表示有人沒坐騎，直接開打
    畫那一隻，印記錄 0xE3「Attempt to ride away?」
    答否 → 往下打
    答是 → 怪物是任務怪（0xFF）或 rand(100) > 5 × 移動率 就脫身（0xE4），
           否則失敗（0xE5）
if 怪物 == 0xFF:                              ; 任務怪那一條
    畫圖（var_6 + 1）；var_6 == 4 就直接結束
    否則印 0xE6，答應之後把六名角色的 [+0x1B] 與 [+0x9F] 清零
    return
```

角色記錄 `+0x1B` 是開場帶的傷（RE-080 用它填 `Wounds`），`+0x9F` 未定。清零的
時機與語意還沒解。

### 開打前設好的其他全域

```
word_30C6C = 數量 + 1        ; 敵人數
word_3B26A = 等級
word_37CF6 = rand(5) − 1     ; 0…4，敵人站位組的索引（RE-098；sub_1E3A8 的預設是 rand(3) − 1）
word_34968 = 0
word_345DE = 0xFFFF          ; 走獎勵那一條（RE-096）
```

訊息是記錄 **0xE7**（`YOU HAVE BEEN AMBUSHED@BY]`）＋數量＋**0xE8**（`, LEVEL]`）
＋等級＋怪物名；數量大於 1 補 `S`，名字以 `S` 結尾改補 `ES`——與打贏那一句
（RE-091）同一套。

### 還沒解的

`sub_2A975`（開打前對怪物編號做的事）、`sub_2A46E`／`sub_2A437`（畫那一隻與收掉）、
角色記錄 `+0x9F`，以及戰場地形圖的來源（`word_37CF6` 已由 RE-098 確定是站位組的
索引，不是地圖編號）。

## RE-098：戰場站位有兩條路，隨機遭遇走的是固定表

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_placement.py`（對 DGROUP 位移做
範圍掃描）與 `tools/ida_probe_koltown_after_battle.py`（`sub_1E3A8` 的
`0x1E41A`…`0x1EA10`）。

### 先記一個找法

`0x2E0C` 這張站位表**沒有具名 xref**——它全部是 `mov [bx+2E0Ch], al` 這種帶基底
暫存器的存取，靜態符號掃描結構上抓不到（RE-092 踩過同一個坑）。改成對**位移範圍**
掃 operand：走過每一支函式的每一條指令，把任何一個 operand 的值落在
`0x2E0C`…`0x2E30` 的都記下來，一次就找到唯一的 writer 與四個 reader。

副作用是會撈到假陽性——`call sub_1CBA9` 的 far 呼叫 operand 剛好落在範圍內。
濾掉分支與呼叫指令就乾淨了。

### `word_34968` 是兩條路的開關（`已證實`）

`sub_1E3A8:0x1E68C`：

```
word_34968 == 0 → 跳到 0x1E8C6，走「固定表」那一條
word_34968 != 0 → 走 0x1E696，先把 0xB956 那 16 byte 複製進 0x9D20（任務地點自帶
                  一件地上的物品），再從 0x2E0C 讀站位
```

任務地點（`sub_2B0FF:0x2B1AD`）設 1，隨機遭遇（`sub_2A4E5:0x2A82B`）設 0。

### 任務地點那一條（`已證實`）

`sub_2B0FF:0x2B204` 把 payload `+3` 起 36 byte 複製到 `0x2E0C`，`sub_1E3A8` 再讀：

```
隊員 n： [+0x1F] = [0x2E0C + 0x1E + n]      ; 欄，payload 的 party_x
         [+0x20] = [0x2E0C + 0x18 + n]      ; 列，payload 的 party_y
敵人 j： [+0x1F] = [0x2E0C + 0x0C + j]      ; 欄，monster_x
         [+0x20] = [0x2E0C + j]             ; 列，monster_y
```

與 `quest-sites.json` 的四個欄位一一對上，抽取器解對了。

`word_345DE >= 0`（訓練戰）時敵人不讀表，改用固定的 `byte_3472D = 0x16`、
`byte_3472E = 0x1F`。

### 隨機遭遇那一條（`已證實`）

`0x1E8C6`…`0x1EA08`，站位全部來自 DGROUP 的四張表：

```
隊員 n： [+0x1F] = [0x0B66 + n]              ; 33 31 33 31 33 31
         [+0x20] = [0x0B60 + n]              ; 30 30 32 32 34 34
敵人 j： [+0x1F] = [0x0BD4 + 12 × word_37CF6 + j]
         [+0x20] = [0x0B8C + 12 × word_37CF6 + j]
```

隊員永遠站在欄 31…33、列 30…34 那一小塊；敵人依 `word_37CF6` 選一組十二格，
每一組都排在另一側。表裡有**六組**，而 `word_37CF6 = rand(5) − 1` 只選得到 0…4，
第六組沒有使用者。

`word_37CF6` 因此是**站位組的索引**，不是先前猜的戰場地圖編號（RE-097 那一條更正）。

### 戰場的地形圖還沒解

`0x1E4C0`：`word_39A4E = 8`，`word_345DE != 0xFFFF` 時改成 `0x0C`；接著
`sub_1A4FE(0x2D50, word_39A4E, …)` 載入那一筆，再 `sub_19B29(word_30C8A + 0x100,
0x8000)` 解一次。記錄號是**固定的**，不隨地點變——這與「24 張地圖對應 24 個任務
地點」（`battle-tiles` 的 manifest）對不上，還要再讀。沒有地形圖就開不了戰場。

## RE-099：戰場的地形是擲出來的，不是載入的

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_after_battle.py` 的
`0x1E41A`…`0x1EA10`，以及 `tools/ida_probe_koltown_ambush.py` 的檔名表。

### 先記一次差點犯的錯

追隨機遭遇的最後一塊時，我一度要下「戰場地形圖的來源還沒解」的結論。**`Generate`
早就寫在 `internal/battle/field.go` 裡**，註解連位址（`0x1E594`…`0x1E5E2`）都標好了。
`rulebook/00` 那條「要說解不出來之前先 grep 自己的 `docs/` 與程式碼」在這裡又成立
一次——這一次是自己幾天前寫的程式碼。

### `word_39A4E` 是磚圖的 payload，不是地圖（`已證實`）

`0x1E4C0`：`word_39A4E = 8`，`word_345DE != 0xFFFF`（訓練戰）時改成 `0x0C`；
接著 `sub_2D3FA(0x2D50, 0x184A, 0x670 + 5 × word_2DF64)` 把檔名組出來、
`sub_1A4FE` 載入那一筆。

檔名表（DGROUP `0x1840` 起）解出來是：

```
0x1844  "town0"
0x184A  "kol%s"
0x1850  "mon%s"
0x1856  "comb%s"
0x185E  "item.lst"
```

`%s` 由 `0x0670` 的副檔名表填：`"pics"`／`"pcga"`／`"herc"`，每個 5 byte，
由顯示模式 `word_2DF64` 選。所以 EGA 下組出來的是 **KOLPICS**，記錄 8 就是
`battle-tiles-08.png` 那張磚圖（訓練戰的 12 是 `battle-tiles-12.png`）——
抽取器早就把兩張都抽出來了。

### 地形逐格擲（`已證實`）

`0x1E58A` 的分支：`word_34968 == 0`（隨機遭遇）才走 `0x1E594`：

```
for i = 0 … 0x0FFF:                  ; 64 × 64
    [地圖 + i] = rand(4) == 1 ? 0x0D : 0x09
```

任務地點跳過這一段（地形從 `TOWN0` 的 payload 61…84 載入，就是 24 張
`battle-map-NN.png`），但**兩條路都會走**接下來的灑點（`0x1E5E5`）：

```
數量 = Σ 100 次 rand(6)              ; 平均 350
```

remake 的 `battle.Generate` 與 `Field.Scatter` 就是這兩段。

### 這一塊補完之後

隨機遭遇的四個必要條件——擲點規則（RE-078／RE-097）、站位（RE-098）、地形
（本條）與磚圖——全部齊了，`battle.NewAmbush` 因此開得起一場戰鬥。

## RE-100：施法

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_spell.py` 一次倒出 `sub_2818D`
（602 條）、`sub_265F3`、`sub_277F4` 與兩張 DGROUP 表。指令碼 `0x0D` 由執行段分派到
這裡（RE-093）。

### 法術放在角色記錄裡（`已證實`）

```
法術記錄 = 角色記錄 + 0x200 + 16 × 參戰者[+0x26]
複製 16 byte 到 0x9E28
```

**角色記錄 `0x200` 起、每 16 byte 一筆、共十六筆就是法術**。`Member.PackCount`
那個「語意未定」的欄位（RE-080 末段）數的正是這個——指令列因此在「有法術」時才
放出施法那一格（`0x22A82`），語意現在對上了。

法術記錄與商品記錄同版面（RE-075），所以名字、`+0x0C` 傷害碼那一套都能沿用。

### 三個欄位（`已證實`）

複製過去之後程式只讀三格：

| 位移 | 內容 |
|---|---|
| +0x0C | 效果編號 |
| +0x0D | **持續回合**的骰式 |
| +0x0E | **效果強度**的骰式；bit 7 是「這一手要不要選目標格」 |

兩個骰式的拆法與武器傷害碼同一套：`(值 ÷ 2) & 0x3F`，低三位查 `[0x00D4]`
（`2 3 5 7 9 11 19 99`）得骰面、`÷ 8 + 1` 得骰數。擲出來的兩個總和分別是持續回合
（`var_10`）與效果強度（`var_C`）。

### 有沒有東西可以施（`已證實`）

`0x281ED`…`0x28281`：`+0x0E` 的 bit 7 決定看哪一邊——沒設就用施法者的方向
（`sub_24BB6`），有設就用選好的目標格（`sub_24C35`）。都沒人就印
**0x169**（`casts a spell, but nothing is there!`）收工。

接著看目標擋不擋得住（`0x28284`…`0x2835C`）：目標倒下（`sub_22866`）就直接算數，
否則拿法術記錄 `+0x08`…`+0x0B` 那四格的高低 nibble 與目標的 `+0x1D`（隊員）或
`+0x28`（敵人）比——高 nibble 相同、而且該格 ≥ 0x50 時低 nibble 也要夠大。
一格都沒中就印 **0x16A**（`casts a spell, but it has no effect!`）。

過了就印 **0x16B**（`casts|S at|F.`）。

### 三種效果（`已證實`）

**效果 6：指定部位**（`0x2845E`）。要選目標格的法術隨機挑一個部位
（`rand(6) − 1`）；否則畫出目標的人形（隊員按性別 `0x47 + [角色+0x1A]`、敵人
`0x49`），印 **0x184**（`Select an area to]`）＋ **0x186**（`hurt`）或
**0x185**（`heal`）＋ **0x16C**（部位清單），等一個 `1`…`6`。

施法者的 `+0x27` 非零是**負面**：目標 `+0x2D` 加上強度，再 `sub_277F4(強度, 部位, 1)`
造成傷害。零是**正面**：那個部位的血量加上強度，上限是它自己的上限
（`+0x11 + 2 × 部位` 與後面那一格），腿被治好而且對象是隊員時再 `sub_21D94`。

**效果 5**（`0x286AB`）：`[0x322E + 2 × 參戰者]` 加或減「強度 × 2」，減不到零以下。
負面是加、正面是減——**那一格是疲勞累計**：

```
sub_23683(i):  word_2FB4A = 50 × [0x322E + 2i] ÷ [0x1C50 + 2i]
攻擊結算（`0x27298`）：扣掉 攻擊值 × word_2FB4A ÷ 200
每回合（`0x20D6E`）：[0x322E + 2i] -= word_30D06 ÷ 10，下限 0
休息（`0x20B02`，指令碼 5 且 +0x2D 為零）：[0x322E + 2i] × 7 ÷ 10
```

所以疲勞越高攻擊被扣得越多，休息一回合恢復三成。`[0x1C50 + 2i]` 是折算的基準，
語意還沒解。remake 這一版對效果 5 什麼都不做——`[0x322E]` 那一條鏈還沒進戰鬥的
規則層（`Actor.Fatigue` 是另一格 `+0x2D`）。`強推論`

**效果 1…4、7、8**（`0x28704`）：查 `[0x12A4 + 效果 − 1]` 得到參戰者記錄的欄位偏移
再加減強度（正面上限 0xFF、負面下限 0）。表是 `08 09 0A 0C 2C 2C 2E 2F`：

| 效果 | 欄位 | 已知語意 |
|---:|---|---|
| 1 | +0x08 | 力量 |
| 2 | +0x09 | 快捷 |
| 3 | +0x0A | 洞察 |
| 4 | +0x0C | 智力 |
| 7 | +0x2E | 未解 |
| 8 | +0x2F | 未解 |

第 5、6 項（`2C`）用不到——那兩個效果在前面就被自己的分支接走了。
效果 0 或大於 8 直接收工，什麼都不做。

### 效果掛在身上（`已證實`）

效果 1…8 那一條會設目標的 `+0x0D = 1` 與 `+0x30 = 持續回合 + 1`；部位那一條
（效果 6）反而把 `+0x0D` 設成 0。`sub_1E3A8:0x20F63` 在回合開始時印
**0x173**（`'s spell wears off!`），所以 `+0x30` 是倒數。

### 最後那一句（`已證實`）

```
0x16D「The spell has a]」
0x16E + min(強度 ÷ 8, 3)     ; mildly／markedly／strongly／powerfully
' '
0x106 + 施法者[+0x27]         ; Positive／Negative
0x172「 effect.」
```

### 參戰者記錄補三格

| 位移 | 內容 | 出處 |
|---|---|---|
| +0x26 | 選中的法術（0…15） | `0x281AE` |
| +0x27 | 效果方向：0 正面、1 負面 | `0x2850D`、`0x28844` |
| +0x30 | 法術剩幾回合 | `0x28730` |

### 還沒解的

`[0x1C50]`（疲勞折算的基準）、參戰者記錄 `+0x2C`／`+0x2E`／`+0x2F`、
擋招那一段比的四格（法術記錄 `+0x08`…`+0x0B`）到底是什麼屬性，以及**法術的母表在哪**
——目前只知道它被存在角色記錄裡，還不知道角色是從哪裡學會的。

## RE-101：任務標成完成

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_quest_flag.py` 倒出 `sub_28F5D`
（58 條）與 DGROUP `0x007C` 的四個除數。旗標的位址算法與 RE-079 解出來的讀取端
（`sub_2B0FF:0x2B108`）是同一套，這一條講的是**寫入端**。

### 誰呼叫（`已證實`）

`sub_28F5D` 只有兩個呼叫點，都在 `sub_1E3A8`（戰鬥主迴圈）裡：

| 呼叫點 | 前置條件 |
|---|---|
| `0x211CB` | `word_34966 == word_30C6C`——擊倒數等於怪物總數，也就是打贏 |
| `0x21176` | 玩家答應記錄 **0x109**（`Do you want to flee?`）而且 `word_39A54 != 0` |

打贏那一條的順序是：先 `sub_28F5D`（標旗標），再 `sub_1C1A4` 與記錄 **0x174**
（`You have defeated|F`），之後才是冒險點數與金冠（RE-096）。remake 的 `State.Step`
照這個順序。

**帶著東西逃掉才算完成。**`0x21045`…`0x21104` 逐一掃六名參戰者，只看隊員
（`+0x01 == 0`），`+0x1F` 或 `+0x20` 等於 4 或 0x3B 的才算數——那兩格是戰場座標，
而可走區就是 4…0x3B（RE-085），所以條件是**站在最外圈、腳已經在戰場邊上**。
`word_3AC5A` 記有幾個人站到了邊上；接著 `sub_22278(i)` 把 `word_39A48` 指到那名
角色的記錄，`sub_2886C(0x30)` 再問一次「他手上拿的是不是那件東西」，是的話記進
`word_39A54`。

`sub_2886C(位移)` 不擲骰：它把 `word_39A48[位移 + i]` 與 `[0xB926 + i]` 逐 byte 比
十六格，全同回 1，並且**隨機遭遇一律回 0**（`0x28875` 先看 `word_34968`）。
位移 0x30 是角色記錄裡手上那件東西的名字（RE-086 的 `|E` 讀的就是這一格），
所以問的是「手上那件是不是任務要的那一件」。`[0xB926]` 那十六個 byte 由誰填
還沒解——任務 payload 已知有一塊 16 byte 複製到 `0xB956`（RE-079），
`0xB926` 是另一塊。

`word_39A54` 非零就問「要逃嗎」；`word_39A54` 為零時還有第二條路——
`word_3AC5A == word_2DF2A`（全隊都站到邊上，`word_2DF2A` 是隊伍人數，記錄
0x17C `life!`／0x17D `lives!` 就是照它選單複數）也問。兩個都不成立連問都不問，
直接回戰鬥。答 `y` 而且真的有人帶著那件東西，就在 `0x21176` 把任務標成完成，
然後從 `0x211AD` 走戰鬥的共同出口。

也就是說任務不必打完——**搶到東西、走到戰場邊上、逃掉，一樣記完成**。
remake 還沒實作戰鬥中的逃走（`ambush.go` 的 `CanRideAway` 是野外遭遇開打前
那一次，RE-098），做的時候這一步不能漏。

### 只有任務戰鬥會做（`已證實`）

`0x28F66` 第一件事就是 `cmp word_34968, 0`，是零就整支跳掉。`word_34968` 是
「這一場是不是任務地點」的開關（RE-098）：隨機遭遇設 0（`sub_2A4E5:0x2A82B`）、
任務地點設 1（`sub_2B0FF:0x2B1AD`）。所以**伏擊打贏不會動任何旗標**。

### 位址算法（`已證實`）

```
起點  = 6 × word_2DF70 + 0x1E2          ; 0x28F70，一組六個 byte
byte  = 起點 + word_30A72 ÷ 4           ; 0x28F80
除數  = [0x7C + (word_30A72 & 3)]       ; 1 4 16 64
角色  = 0xBB88 + 0x300 × i，i 走 0…5    ; 0x28FEF 的 `add var_4, 300h`
```

`word_30A72` 是任務地點編號 0…23、`word_2DF70` 選哪一組六 byte，兩個都與讀取端
（RE-079）用的是同一對全域。**六名角色一律走完**，不看誰在隊上、也不看誰在場。

### 沒接的人也會被標成完成（`已證實`）

`0x28FB9` 是這一段的關鍵，它只比 bit 1：

```
0x28FB9  and ax, 2
0x28FBC  cmp ax, 2
0x28FBF  jnz 0x28FC4      ; 不是 2 就往下寫
0x28FC1  jmp 0x28FEF      ; 已經是 2，跳過這名角色
```

值為 0（沒接這個任務）落在「不是 2」那一邊，所以**照樣被寫成 2**。寫法是
`0x28FCD` 起的「先 `and` 掉那一位、再 `or` 上它的兩倍」：

```
al = 除數 xor 0xFF ; [byte] &= al      ; 0x28FCD、0x28FD9——清掉這個 2-bit 欄位
al = 除數 × 2      ; [byte] |= al      ; 0x28FDD、0x28FEB——填 2 進去
```

`0x28FC4` 順便把 `word_34968` 設成 `0xFFFF`——它本來就是非零（不然整支跳掉了），
所以對「是不是任務地點」的判讀沒有影響。**那個值是給結局用的**：任務地點那一側
在戰鬥回來之後看它等不等於 0xFFFF，等於才檢查要不要播結局（RE-106）。

**「接了」那一位元的作用只在入口**：`sub_2B0FF` 要有人的欄位等於 1 才讓玩家踏上
那個地點（RE-079）。打完之後全隊一起記完成，這是原版的行為，不是簡化。

### remake 的落點

`internal/battle/victory.go` 的 `MarkQuestDone`／`markQuestDone`，由
`execution.go` 的 `Step` 在 `VerdictWon` 時呼叫。`word_34968` 用 `Site.Index >= 0`
表示。`Member.Quests` 只存一組六 byte，等於把 `word_2DF70` 固定在 0——原版
`CHARDATA` 裡的三名帶旗標角色都落在這一組（RE-079），還沒看到會換組的路徑。

### 還沒解的

`word_2DF70` 什麼時候會不是 0、`[0xB926]` 那十六個 byte 是誰填的。
（`word_34968 = 0xFFFF` 已由 RE-106 解出。）

## RE-102：射擊的彈道

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_throw.py` 倒出 `sub_26D76`
（901 條）與戰鬥收尾整段，`tools/ida_probe_koltown_missile.py` 補上
`sub_25F8A`（160 條）與 `sub_24C35`（183 條）。

### 一支函式三種手（`已證實`）

`sub_26D76(arg_0)` 同時是近戰、徒手、投擲與射擊。`arg_0` 只決定一件事：
非零才先問「面前那一格有沒有人」（`sub_24BB6(0, 方向, 8)`），沒有就印
**0x11F**（`staggers about.`，徒手）或 **0x120**（`whips|B|E through the air.`）
收工。射擊那一條傳 0（`0x20824`），所以跳過這個檢查。

之後分岔的是全域 `word_31122`（目前的指令碼），不是參數：

| `word_31122` | 走哪一段 |
|---|---|
| 0x23（射擊）| `0x26EAD` 的彈道 |
| 其餘（0x21 武器、0x29 徒手、0x22）| `0x26FC4` 的攻擊選項修正表 |

**0x22 沒有自己的段落**，因為它走不到這裡。執行段的分派把 0x22 與 0x23 一起送進
`0x20799`（`0x20785`），而 0x22 那一手在那裡就被攔成「裝填」、指令碼歸零，
跳過 `sub_26D76`。它是**還沒上弦的射擊武器**那一格，不是投擲——三段證據串起來
見 RE-105。

### 走線（`已證實`）

`sub_25D7A(mode)` 是 Bresenham，起點 `word_39A3E`／`word_39A40`、游標
`word_39A42`／`word_39A44`。射擊在 `0x26EAD` 把**記住的目標格**
（`[0x1A74 + 2i]`／`[0x1B36 + 2i]`，RE-081）放進起點、把行動者的位置放進游標，
所以線是從行動者往目標走。四種 mode：

| mode | 用途 |
|---|---|
| 0 | 彈道（`0x26EF9`）|
| 1 | 只走一步，給「往目標移動」用（`0x25EDF` 立刻回 0）|
| 2 | 規劃段的游標（RE-081）|
| 3 | 視線，把結果寫進 9×9 的 `[0x2E3A]`（RE-079）|

主軸每步走一格、誤差項滿了副軸才走一格，起點那一格不判——迴圈一開頭就先位移
（`0x25E19`）。remake 的 `walkLine` 是這一段，視線與彈道共用。

### 每一格判什麼（`已證實`）

`sub_25F8A(mode, 第幾步)` 回 1 表示停下來，回 0 表示可以再往前。停下來的原因
寫在 `word_30A74`：

| 值 | 意思 | 出處 |
|---|---|---|
| 1 | 樹 | `0x25FDD` |
| 2 | 半路的人 | `0x2610C` |
| 3 | 牆或碎石 | `0x260A2` |

判的順序是：

1. **樹**（磚塊 0x0A，只有 mode 0 認）：扣 **0x23** 點命中值，扣完還 ≥ 0 就穿過去
2. **擋不擋**：`[0x1312 + (磚 & 7)]` AND `[0x10CE + (磚 ÷ 8)]`，非零就是擋——
   與視線同一張表（RE-079）
3. **例外**（磚塊落在 0x23…0x2D）：記錄 **+0x1E 非零**整條線都穿得過，
   或者站在 0x24／0x25 上而且是第一步
4. **人**：`sub_24C35(0, 游標)` 找站在那一格的，扣掉他**體型的一半**
   （`+0x0B ÷ 2`）。走到目標格本身不算擋路（`0x260BA`）

第 3 條的 `+0x1E` 這次讀反過來了：**非零才穿得過**，也就是騎著的人視線高。
RE-079 原先寫成「+0x1E 為 0 才穿得過」，方向相反，該段已更正；remake 的
`seesOver` 與 `Field.Visible` 跟著補上騎乘那一條。

命中值扣成負之後原版會跳回同一格再判一次（`0x2612B` → `0x260B1`），而
`sub_24C35` 沒有「已經找過」的記憶，每次回同一個人——要靠 16-bit 有號數繞回正值
才出得來。remake 扣一次就停，結果與原版繞完之後相同。

### `sub_24C35(旗標, x, y)`（`已證實`）

掃參戰者找站在 (x, y) 的一名，找到就把編號寫進 `word_3237E` 回 1。空位
（+0x01 == 2）、自己（`word_37D2A`）與已退出的（`sub_22866`）一律跳過，
旗標再篩：

| 位元 | 篩掉誰 |
|---|---|
| 0x01 | 我方（`+0x01 == 0`）|
| 0x02 | 敵方 |
| 0x04 | 沒騎的（`+0x1E == 0`）|
| 0x08 | 騎著的 |
| 0x10 | `+0x05` 為零的（要先 `sub_252CB(智力, word_3AC5E)` 成立）|
| 0x20 | 指令碼不是 0x0D 也不是 0x23 的 |

### 射到哪、印什麼（`已證實`）

彈道停下來之後：

- `word_30A74` 是 1 或 3 → 印 **0x121**（`fires a missle into a]`）接
  **0x122**（`tree!`）或 **0x123**（`wall!`），這一手結束
- 沒被擋 → `sub_24C35(1, 停下來那一格)` 找敵方，沒有再 `sub_24C35(0, 同一格)`
  找任何一名；兩次都沒有就印 **0x124**（`fires a missle but nothing is there!`）
- 有目標 → 印 **0x125**（`fires a missle`）或 **0x126**（`throws a boulder`），
  接 **0x127**（` at|F`）

### 攻防分數（`已證實`）

射擊的攻擊分數是 `0x26E08`…`0x26EA0` 那一段，也就是 `(力量 + 攻擊 − 傷勢 −
累積傷害 × 10 − 疲勞 × 20 − 負重) ÷ 10 + rand(100)`，**沒有攻擊選項修正表**
——那一段在 `0x26FC4`，射擊跳過了。彈道再從這個值扣掉沿路的樹與人。

防禦分數換成 `目標[+0x09] ÷ 2`（快捷的一半，`0x271E3`），然後 `jmp 0x2728C`
接回共用的折算（傷勢、累積傷害、負重、盾牌）。跳過去的那一段裡有「防禦選項為 0
就直接命中」與招架次數 +1，所以**射擊不看防禦選項，也不會讓對手多算一次招架**。

分數不夠時印的是 **0x13F**（`but misses completely!`）而不是招架句
（`0x273A5`）。防禦分數是負的一律當 0 再比（`0x27454`），命中餘裕不會因此灌水
——這一條近戰也適用，remake 原本漏了，這次一併補上。

半路被人接下來的那一發（`word_30A74 == 2`）另有一條：攻防**兩邊都換成
`rand(100)`**（`0x27153`），彈道算出來的命中值不算數。

### remake 的落點

`internal/battle/missile.go` 的 `Shoot`／`MissileShot`、`attack.go` 的
`MissileDefenceValue`／`MissileDefenceScore`／`ResolveMissile`、`execution.go` 的
`shoot`（含 RE-093 的裝填那一步）與 `enemyAt`。走線抽成 `sight.go` 的
`walkLine`，視線與彈道共用。

### 還沒解的

`sub_22229` 回傳的是什麼（`0x27083` 用它決定印射箭還是丟石頭）、
`sub_27FC1` 的飛行動畫。（0x22 已由 RE-105 解出。）

## RE-103：戰鬥中的逃走

輸入與工具同 RE-070。反組譯來自 `tools/ida_probe_koltown_throw.py` 的視窗
`0x20F00`…`0x21800`，`sub_2886C` 的全文在 `battle-drop.json`。

### 什麼時候問（`已證實`）

每個回合走完先數兩邊倒下的人（`0x20FA9`…`0x21042`，RE-091 的四個判斷）。
四個都不成立——也就是還沒分勝負——才落到 `0x21045` 的逃走判定：

```
word_3AC5A = 0 ; word_39A54 = 0
對六名參戰者 i：
    +0x01 != 0（不是隊員）跳過
    +0x1F 或 +0x20 等於 4 或 0x3B 才算數     ; 站在可走區的最外圈
    word_3AC5A += 1
    sub_22278(i)                             ; word_39A48 指到這名角色的記錄
    sub_2886C(0x30) 非零 → word_39A54 += 1
```

`+0x1F`／`+0x20` 是戰場座標，可走區是 4…0x3B（RE-085），所以條件是**腳踩到最
外圈**；兩軸任一個到邊就算，不必站在角上。倒下的人不排除——這一段只比 `+0x01`。

問不問由兩條路決定（`0x21116`）：

| 條件 | 結果 |
|---|---|
| `word_39A54 != 0` | 問 |
| `word_39A54 == 0` 而且 `word_3AC5A == word_2DF2A`（全隊都到邊上）| 問 |
| 兩個都不成立 | 不問，直接排下一回合（`0x2119B`）|

問的是記錄 **0x109**（`Do you want to flee?`），`y`／`Y` 才算答應，`n`／`N` 回
戰鬥，其餘的鍵重問（`0x21149`…`0x21198`）。

### `sub_2886C(位移)` 是名字比對（`已證實`）

```
word_34968 == 0 → 回 0                        ; 隨機遭遇一律不算
for i in 0…15:
    word_39A48[位移 + i] != [0xB926 + i] → 回 0
回 1
```

十六個 byte 全同才算數。位移 0x30 是角色記錄裡**手上那件東西的名字**
（RE-086 的 `|E` 讀的就是這一格），所以問的是「他手裡拿的是不是任務要的那一件」。
`[0xB926]` 那十六個 byte 由誰填還沒解——任務 payload 已知有一塊 16 byte 複製到
`0xB956`（RE-079），`0xB926` 是另一塊。

**任務不必打完。**搶到那件東西、走到戰場邊上、逃掉，`0x21176` 照樣把任務標成
完成（RE-101）。答應之後從 `0x211AD` 走戰鬥的共同出口。

### 逃走的收尾（`已證實`）

出口那一段依 `word_345DA`（倒下的隊員）與 `word_34966`（倒下的敵人）分三條：

| 條件 | 走哪一條 |
|---|---|
| `word_34966 == word_30C6C` | 打贏（`0x211CB`，RE-096）|
| `word_345DA == word_2DF2A` 或 `word_345DE != 0xFFFF` | 打輸，印 0x140 ＋ 名字 ＋ 0x178／0x179 ＋ **0x17A**（` defeated you!`）|
| `word_345DA == 0xFF` 而且 `word_34966 == 0xFF` | 逃走，`sub_288D4(0)` 之後印 **0x17B**（`You flee for your]`）接 **0x17C** `life!`／**0x17D** `lives!`，單複數看 `word_2DF2A`（隊伍人數）|

`word_34966 = 0xFF` 出現在 `sub_1E3A8:0x1FCF5`（operand 掃描找到的唯一一處），
`word_345DA` 那一半還沒對上——它在 `0x20FA9` 每回合都被歸零重數。兩處的上下文
都還沒倒出來。

三條路最後都會走到 `0x21607` 起的收尾：升級（`word_345DE != 0xFFFF` 時
`sub_2685A` ＋ 記錄 **0x17E** ＋ 角色記錄 `[+0x29] += 1` ＋ `sub_289EE(1)`）、
賭金結算、以及掉東西。

### 賭金與掉東西（`已證實`）

賭金（`0x216C9`…`0x2175B`）：`[0x1B2A + 2i]` 是第 i 名的賭注，非零才算——
加進 32 位元的金錢（`word_30A6E`／`word_30A70`），再依正負印 **0x17F**
（`wins a bet!`）或 **0x180**（`loses a bet!`）。輸掉那一場時
（`0x21688`）這張表會先整排取負再結算。

掉東西（`0x21765`…）只在 `word_3AC5A == 0` 時做，機率讀 `byte_3AC7D`：

```
word_39A4E = byte_3AC7D
對六名參戰者 i：
    +0x01 != 0 跳過；sub_22866(i) 為 0（還沒退出）跳過
    sub_22278(i)
    位移 word_2FB5A 從 0xA0 起，跳過 0x170 那一格：
        rand(100) < word_39A4E 才可能掉
        sub_2886C(位移) 非零（這一格是任務要的那件）就不掉
```

也就是**任務物品掉不掉**——同一支比對函式在這裡當保險。`byte_3AC7D` 那個機率
是誰算的還沒對上。

### 升級不在戰鬥模組（`已證實`）

`0x21611` 那一段要 `word_345DE != 0xFFFF` 才做。探針
`tools/ida_probe_koltown_writers.py` 走過 338 支函式的每一條指令比對 operand 的
值，**寫這個全域的只有四處**：

| 位址 | 寫什麼 |
|---|---|
| `sub_1E3A8:0x1E3D2` | 0xFFFF——戰鬥一開始就清掉 |
| `sub_2A4E5:0x2A940` | 0xFFFF——隨機遭遇 |
| `sub_2B0FF:0x2B348` | 0xFFFF——任務地點 |
| **`sub_2AAA8:0x2AFE6`** | `ax`——**競技場** |

也就是說**只有競技場打完會升位階**，任務戰鬥與野外遭遇一律走不進 `0x21611`。
這與訓練那一側的規則對得起來：帳面位階（角色記錄 `+0x29`）與實得位階
（`sub_14371`）不同時，訓練會叫玩家「Go to the Arena!」（RE-074）。
升級本身就是 `[+0x29] += 1`，一次一級。

`0x21688` 那一段（賭注整排取負）也只有競技場才有意義，同一個結論。

### operand 掃描抓不到的兩個

`byte_3AC7D`（掉東西的機率）只有 `0x21765` 一處**讀**，`[0xB926]`（任務要的
那件東西）連一次命中都沒有。兩個都是被整塊複製寫進去的——`sub_19D98` 那一類
搬 payload 的函式用的是指標，operand 裡不會出現目標位址，靜態掃描結構上看不到。
要解得換成別的辦法（追 payload 的搬運，或實跑時看記憶體）。

### remake 的落點

`internal/battle/flee.go` 的 `OnFieldEdge`／`FleeOffer`／`Flee`／`OfferFlee`，
`State.QuestItem` 是 `[0xB926]` 那件東西的名字。問句接在 `Step` 走完之後：
`PromptFlee` ＋ `StageFlee`，只認 `y`／`n` 兩個鍵，別的鍵不動——與原版
`0x21158`…`0x21190` 相同。

**掉東西還沒做**，要先有裝備欄十二格的 typed data。升級照上面的結論歸到競技場
那一側，不放在戰鬥模組。

順帶訂正 RE-079：那裡把 0x109 寫成接在 0x108「Begin next round?」答否之後，
實際上兩者無關。remake 的「Begin next round?」答否因此改成重排這一輪
——U-TURN 那一格的原版行為還沒對上位址。

## RE-104：敵人的射擊與視線

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_enemy_ai.py` 倒出 `sub_2374C`
（833 條）與它呼叫的十九支，另加決策樹那一段的視窗 `0x23C1A`…`0x24060`。

### `sub_25D0F` 回 0 才是通得到（`已證實`）

```
游標已經等於起點 → 回 0
sub_25D7A(2) 回 0（走完沒被擋）→ 回 0
走完之後游標回到起點 → 回 0
其餘 → 回 1
```

**回傳值是反的**：0 表示「線通得到」。兩個呼叫點都拿 `jz` 接著做事——
`0x20F38` 通得到才畫戰場、`0x23EB0` 通得到才考慮射擊。

`sub_25D7A(2)` 是走線的第三種 mode（RE-102）：不比樹、不看人、不畫東西，
只有地形擋不擋視線那一張表算數。remake 把它做成 `Field.LineOfSight`，
並且**把回傳值翻正**——這一側回 true 就是通得到，免得每個呼叫點都要記得反著讀。

### 射擊那一條（`已證實`）

環上找到目標之後（`0x23E4F`…`0x23F53`）：

```
起點 = 目標的 +0x1F／+0x20     ; word_39A3E／word_39A40
游標 = 自己的 +0x1F／+0x20     ; word_39A42／word_39A44
sub_25D0F() != 0 → 回去在同一環找下一個（0x23DEC）
word_349B2 >= 4 → 貼近打
自己的 +0x00 為零 → 貼近打
自己的 +0x15（右臂）>= 0x80 → 貼近打        ; 原版連比兩次同一格
否則：
    +0x1D = 1                               ; 直接算成已裝填
    [0x1A74 + 2i] = word_36A1C              ; 目標那一格
    [0x1B36 + 2i] = word_37BEE
    word_31122 = 0x23                       ; 射擊
    word_30A78 = 0                          ; 攻擊選項 0
```

**敵人不必花一回合裝填**——`+0x1D` 直接設 1，玩家那一側要先花一手（RE-093）。

`word_349B2` 是環號，從 **1** 起算相鄰的那一圈（`0x23DDE`），到 7 為止。
remake 的環號從 0 起算同一圈，所以門檻是 `ring < 3`；環 0 在前面就被貼身攻擊
接走了，真的會射的是環 1 與環 2。

remake 的落點是 `internal/battle/ai.go` 的 `enemyShoot` 與 `EnemyShootRing`。
線通不到時 **remake 換下一環，原版是回頭在同一環繼續找**——`findInRing` 一次只
回一名，兩者只在「同一環有多名目標而先挑到的那名被擋住」時不同。

### 反制掃描 `sub_253E7`（`部分解出`）

`sub_2531C(上一格, 0x19)` 繞相鄰那一圈，對每一名鄰居算一個優先度 `var_10`，
留下最高的那一個，掃完之後把 `var_2` 寫進 `word_30A78`、`var_8` 寫進
`word_30916`，並把那一格的方向寫回自己的 `+0x28`。優先度是：

| 情形 | 優先度 | 攻擊選項 | 防禦選項 |
|---|---:|---|---|
| 鄰居這一回合**還沒行動**（`+0x22 == 0`）| 4 | 0x0B | 7；它的指令碼是 0x21 而且 `rand(2) == 1` 時改用它的 +0x23／+0x24 |
| 已行動、指令碼 0x21，而且**正對著自己**（`+0x28 xor 格號 == 4`）| 5 | 它的 +0x23 | 它的 +0x24 |
| 已行動、指令碼 0x21／0x29 但不是正對 | 3 | 0x0B | 它的 +0x24 |
| 已行動、指令碼 0x27（移動）| 1 | 0 | 它的 +0x24 |
| 已行動、其餘（`+0x24` 非零 → 2，為零 → 4）| 2／4 | 0 | 它的 +0x24 |

掃完一圈之後優先度還是 0 就回 0（沒有結論，主幹改走找目標）；優先度是 1
（鄰居在移動）時另外呼叫 `sub_25653`（側翼）。

`word_30916` **是防禦選項**（`0x1FA6D` 把它寫進 `+0x24`，RE-105）——主幹在別的
分支把同一個全域當成別的東西用（`0x23D8B` 寫的是 `sub_2590F` 的回傳），
那是 Borland 慣見的全域重用。即使如此 remake 仍不實作這一段：挑目標的優先度
規則沒有可觀察的驗證方式，照抄會做出「自洽但無從反駁」的行為。

### 還沒解的

`sub_25653` 側翼那一支、撤退狀態（`+0x29`／`+0x2A`）、以及沒有武器時的
指令碼 0x26。（`word_30916` 是防禦選項，已由 RE-105 定案。）

## RE-105：0x22 不是投擲，是「還沒上弦」

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_plan_dispatch.py` 倒出規劃段的
分派與收尾，`tools/ida_probe_koltown_writers.py` 補掃 `word_37CF2` 的寫入者。

### 三段碼串起來（`已證實`）

| 位址 | 做什麼 |
|---|---|
| `sub_228B7:0x22D44` | 選中的圖號寫進 `word_37CF2` |
| `0x1FA02` | 分派表——**沒有 0x22**，它落 default（`0x1FA6A`）直接走收尾 |
| `0x1FA6D` | `+0x25 = word_37CF2`、`+0x23 = word_30A78`、`+0x24 = word_30916` |
| `0x20785` | 執行段對 **0x22 與 0x23** 都跳 `0x20799` |
| `0x20799` | 種類不是 6／0x0A／0x0B 而且 `+0x1D` 為零 → 裝填，指令碼歸零 |

指令列上 0x22 那一格的出現條件（RE-083）**正好是**「種類不是 6／0x0A／0x0B、
`+0x1D` 為零」；0x23 那一格是「種類是那三種，或 `+0x1D` 非零」，而且推完就設
`+0x1D = 1`。

兩邊接起來的結論是：**選 0x22 一定變成裝填**，這一回合就沒了；下一回合
`+0x1D` 非零，同一格換成 0x23，那時才真的射出去。射完 `0x20860` 又把 `+0x1D`
清回 0，回到 0x22。

所以 **0x22 與 0x23 是同一件事的兩個階段，不是「投擲」與「射擊」**。
`sub_26D76` 裡沒有一處比 0x22 —— 它根本走不到那裡（RE-102 把這件事記成
「0x22 的行為沒有證據」，現在有了）。

### `word_30916` 是防禦選項（`已證實`）

`0x1FA6D` 的三行把 `word_37CF2`／`word_30A78`／`word_30916` 分別寫進
`+0x25`／`+0x23`／`+0x24`，也就是**指令碼／攻擊選項／防禦選項**。RE-104 把
`sub_253E7` 寫進 `word_30916` 的值標為「語意未定」，現在定了：那是要對手用的
防禦選項。反制掃描仍然不實作——它挑目標的優先度規則沒有可觀察的驗證方式。

`0x1F9E0` 另有一條：`word_30A78` 與 `word_30916` 都是 0 時把指令碼改成 5
（休息）。

### remake 的落點

常數 `CmdThrow` 改名成 `CmdLoad`（`commands.go`），規劃段落 `commit(i, icon, 0, 0)`
（原版那兩個選項沿用殘留值，remake 取 0——裝填那一手不看它們），執行段與射擊
共用 `State.shoot`。分派表認不得的圖號改由 `chooseCommand` 的 default 記進
`Plan.Unimplemented`，不再靠 0x22 那一格當唯一的記錄點。

## RE-106：結局的觸發與流程

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_ending.py` 倒出 `sub_2B4A7`
（273 條）、它呼叫的每一支，以及唯一的呼叫點 `0x2B3A9` 前後那一段。
七張 ENDPICS 的來源與幾何見 RE-017。

### 誰觸發（`已證實`）

結局接在**任務戰鬥打完之後**（`sub_2B0FF` 的尾端）：

```
0x2B34E  sub_1E3A8(0)                       ; 打這一場
0x2B36B  word_34968 != 0xFFFF → 跳過
0x2B375  var_4 = 0
0x2B37D    角色記錄[+0x00] 為零 → 下一位     ; 這一格沒有角色
0x2B38F    角色記錄[0x1DD + word_2DF70] == 0x17 才成立
0x2B3A9    sub_2B4A7()                      ; 結局
0x2B3B1  var_4 += 0x300，走完六名
```

**`word_34968 == 0xFFFF` 是 `sub_28F5D` 設的**（`0x28FC4`，RE-101）——那一支在
標任務完成時順手把它從 1 換成 0xFFFF。RE-101 把這個賦值記成「語意還沒對上」，
現在對上了：它是「這一場真的標到了東西」的旗標，回到任務地點那一側用來決定
要不要檢查結局。沒標到就不檢查。

判準是角色記錄 `[0x1DD + word_2DF70]` 等於 **0x17**。任務旗標區在
`0x1E2 + 6 × word_2DF70`（RE-079），這一格在它前面五個 byte；`word_2DF70` 為 0 時
就是 `[0x1DD]`。**那一格是「已回報的任務數」**——回報一次加一（RE-107），
而最終任務要回報滿 0x17（23）個才接得到。二十四個地點編號 0…23，最後一個是
最終任務，打完它就播結局，三邊對得起來。

### 流程（`已證實`）

`sub_2B4A7` 是一條直線，七段：

```
sub_1CBA9(2)                                  ; 進結局的畫面模式
sub_1A4FE(0x18B0, 0, byte_37D2C, …)           ; 載 ENDTXT payload 0
每一段：
    sub_2B845()                               ; 依 word_2DF1E 選清畫面的方式
    word_3499E／word_349A0／word_34A06／word_34A0E = 矩形
    sub_2B80A(圖號)                           ; 檔名 `end%s`，載入並畫
    sub_1A8FD(記錄, 0x9E5C) × 2 或 3
    sub_1B100(0)                              ; 等一個鍵
sub_1103F(0) ; sub_1CBA9(4)
sub_1A4FE(0x18B8, 1, …) ; sub_1CBA9(1)        ; 換回城鎮那一側
```

圖與文字的配對是固定的：

| 圖 | ENDTXT 記錄 | 矩形（左欄, 上列）–（右欄, 下列）| 換算成像素 |
|---:|---|---|---|
| 0 | 0、1 | (1, 13)–(22, 135) | 168×122 |
| 1 | 2、3 | (15, 41)–(39, 152) | 192×111 |
| 2 | 4、5 | (1, 50)–(38, 143) | 296×93 |
| 3 | 6、7、8 | (1, 50)–(23, 147) | 176×97 |
| 4 | 9、0x0A | (17, 12)–(38, 135) | 168×123 |
| 5 | 0x0B、0x0C、0x0D | (0, 46)–(24, 148) | 192×102 |
| 6 | 0x0E、0x0F、0x10 | (16, 45)–(39, 143) | 184×98 |

**四個全域是矩形的四個邊**：`word_3499E` 是左欄、`word_349A0` 是上列、
`word_34A06` 是右欄、`word_34A0E` 是下列。欄乘八得像素，右欄減左欄乘八就是
RE-017 量到的圖寬，下列減上列就是圖高——**七張全部吻合**，這是四個全域語意的
交叉驗證。

ENDTXT 有 18 筆記錄，這裡用掉 17 筆（0…0x10）；**最後一筆沒有任何呼叫點**。

文字自帶 `~Vxx`／`~Hxx` 定位（例如記錄 0 的 `~V08~H17`），所以每一段的字擺在
圖的哪一側是文字自己決定的，程式不另外算。

### 還沒解的

第 18 筆 ENDTXT 給誰用。（`[0x1DD]` 與 0x17 已由 RE-107 解出。）

## RE-107：任務的回報、接取門檻與最終任務

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_writers.py` 的 operand 掃描
找出角色記錄 `[0x1DD]` 的五個使用點，`tools/ida_probe_koltown_progress.py`
倒出每一處的前後文。

### `[0x1DD + 群組]` 是「已回報的任務數」（`已證實`）

回報在建築內的主迴圈裡（`sub_11A93:0x11C28`…`0x11CD0`，訓練也在同一支）：

```
var_8 = byte_2FBFA & 0x7F              ; 這棟建築對應的任務地點編號
var_6 = var_8 ÷ 4                      ; 旗標的第幾個 byte
var_8 = [0x7C + (var_8 & 3)]           ; 除數 1／4／16／64
對六名角色（var_2 一筆 0x300）：
    角色記錄[+0x00] 為零 → 跳過
    欄位 = (角色[0x1E2 + 6 × word_2DF70 + var_6] ÷ 除數) & 3
    欄位 != 2 → 跳過                    ; 只收「已完成」的
    角色[0x1E2 + …] |= 除數             ; 2 or 1 ＝ 3
    角色[0x1DD + word_2DF70] += 1       ; 已回報的任務數加一
```

所以那個 2-bit 欄位有**四個**狀態，不是三個：0 沒接、1 接了、2 完成、**3 回報過**。
RE-079 與 RE-101 只解到 2，這一格是第三個值的來源。

`byte_2FBFA` 是建築自己的欄位：**低七位是任務地點編號、bit 7 是「最終任務」旗標**
（`0x11C33`、`0x1456D`、`0x145C6` 三處都這樣讀）。它與價錢那兩格
（`byte_2FBFC`／`byte_2FBFD`，RE-073）同屬一塊，整塊由建築 payload 複製進來——
operand 掃描只找得到讀取，找不到寫入。

**它對到 payload 的 0x36**（RE-109 驗證）：先用 `icon_flags[1] & 4` 篩出真的收
任務物品的 23 棟，那一格的低七位全部落在 0…23，bit 7 只有一棟設。
先前把 102 棟全列時看到 90／100／110／120 那種超界的值，是因為沒有篩選——
沒有那個旗標的建築，那一格根本不是任務編號。

### 接任務要旗標為 0，最終任務另有門檻（`已證實`）

`sub_14564:0x1459C`…`0x145F2`：

```
欄位 = (角色[0x1E2 + …] ÷ 除數) & 3
欄位 != 0 → 不能接（已經接了、完成了或回報過了）
byte_2FBFA <= 0x7F → 照常給任務（0x145F5）
byte_2FBFA > 0x7F（最終任務）：
    角色[0x1DD + word_2DF70] < 0x17 → 回 4        ; 擋下來
    否則照常給
```

**0x17（23）就是這裡的門檻**，與結局的判準是同一個值（RE-106）：回報滿二十三個
任務才接得到最終那一個，打完它就播結局。二十四個任務地點編號 0…23，
最後一個是最終任務——這樣三邊就對起來了，RE-106 標成 `假說` 的那一句可以升成
`已證實`。

### 另外兩個使用點（`部分解出`）

- `sub_15ABF:0x160CD`：`[0x1DD]` 與 `[0x1DE]` 都是 0 時走一條分支，前面剛畫完
  角色記錄 `[0x200]`（法術）那一格的圖示（`sub_11703(0x12)`），像是角色資料頁上
  「有沒有任務在身」的圖示。
- `sub_167D9:0x167F6`：對 `[0x1DD + i]` 逐格看非零，非零就 `sub_16239`。
  看起來是任務的一覽畫面，還沒逐條讀。

### remake 的落點

`internal/ending` 的 `TriggerValue` 註解改成已證實；接取與回報還沒實作——
那是城鎮那一側的事（建築內主迴圈），等城鎮那一輪再做。

## RE-108：競技場——升階、對手與下注

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_arena.py` 倒出 `sub_2AAA8`
（541 條）與它呼叫的每一支。升級掛在這一側的理由見 RE-103：`word_345DE` 唯一
寫進真值的地方是 `0x2AFE6`。

### 誰進得去（`已證實`）

`sub_2AAA8` 的唯一呼叫點是 `sub_29004:0x29442`。進來先畫圖示列（DOOR 與
U-TURN），選 DOOR（回傳 2）才進場。標題是記錄 **0xEB**
（`The Arena@The Tester`）。

候選由 `0x2AC0F`…`0x2AC77` 掃六名角色決定，兩個條件（`0x2AC4F`、`0x2AC5C`）：

```
角色記錄[+0x00] 為零 → 跳過
帳面位階 = 角色記錄[+0x29]
sub_14371() == 帳面位階 → 跳過        ; 實得與帳面一樣，不用升
帳面位階 >= 0x18（24）→ 跳過          ; 位階上限
合格：[0x9E28 + i] = 1，var_A = 0x300 × i
```

`sub_14371` 是實得位階（RE-074 的計算，訓練那一側用同一支比對，
不同就叫玩家「Go to the Arena!」）。**位階上限是 24。**

沒有人合格就依隊伍人數印記錄 **0xFA**（`None of you are ready`）或
**0xF9**（`You are not ready`）離開。

隊伍只有一人就直接用那一名；多人印記錄 **0xED**（`Which one wishs to fight?`）
讓玩家選，選到不合格的印 0xF9 重選。

選好之後記錄 **0xEE**（`Are you sure you wish to enter?`）確認，
答否印 **0xEF**（`Return when you have gained courage!`）離開。

### 下注（`已證實`）

隊伍只有一人就跳過整段直接開打。多人的話換到賭場：記錄 **0xEC**
（`A Gambling Booth@The Bet Taker`）、**0xF0**（`You follow him down a tunnel…`），
再問 **0xF1**／**0xF2**（單人／多人的「要不要下注」）。答否就開打。

下注迴圈（`0x2AE13`…`0x2AFDB`）：

| 情形 | 結果 |
|---|---|
| 選到上場的那一位（`word_39A52 == var_A`）| 記錄 **0xF4**（`You cannot bet!`）|
| 那一位已經下過注（`[0x1B2A + 2i] != 0`）| 記錄 **0xF5**（`You cannot bet again!`）|
| 其餘 | 記錄 **0xF6**（`How much?`）|

金額由 `sub_1B5C8(0x2E3A, 2)` 讀**兩位數字**，換算是
`10 × (第一位 − 0x30) + (第二位 − 0x30)`，所以**上限是 99**。第一位是零就取消。
金額大於那名角色的錢包（`sub_14CCF` ＋ `word_30A6E`／`word_30A70`）印記錄
**0x56**（與訓練共用的「錢不夠」）。

再問記錄 **0xF7**（`Betting on your friend?`）：

- 答**是** → 金額原樣寫進 `[0x1B2A + 2i]`（`0x2AFBD`）
- 答**否** → 先取負再寫（`0x2AFAF` 的 `neg`），並印記錄 **0xF8**（`Good friends, eh?`）

也就是**負值＝賭他會輸**。戰後結算（`0x216EF`，RE-103）看那一格的正負決定印
**0x17F**（`wins a bet!`）或 **0x180**（`loses a bet!`），而打輸那一場時
`0x21688` 會把整排先取負——兩邊配起來，賭中的一律變成正值。

`sub_144EF` 這個 y／n 問句**回非零是「否」**（`0x2AD65` 拿它接 0xEF 那一句就看得出來）。

### 對手（`已證實`）

`0x2AFE3`…`0x2B00C`：

```
word_345DE = var_A                    ; 打完之後 0x21611 靠它決定誰升階
word_3B26A = sub_14371() + 1          ; 對手等級 ＝ 實得位階 ＋ 1
word_30C6C = 1                        ; 一隻
word_37CF6 = rand(5) − 1              ; 陣型
word_34968 = 1                        ; 走「任務地點」那一條（RE-098）
```

之後 `0x2B0D5` 呼叫 `sub_1E3A8` 開打（RE-079 記的三個戰鬥入口之一）。

打完回來走的是戰鬥的共同收尾（RE-103）：`word_345DE != 0xFFFF` 就升階
——`sub_2685A` ＋ 記錄 **0x17E**（`you have gained a level!`）＋
角色記錄 `[+0x29] += 1` ＋ `sub_289EE(1)`，接著賭金結算。

**升階一次只加一級**，所以帳面追上實得要打好幾場。

### 還沒解的

`sub_14CCF` 取的是哪一份錢包（角色自己的還是隊伍共用的）、`word_37CF6` 那個陣型
怎麼用、以及 `sub_2B454(5)`／`sub_2B454(6)` 畫的是什麼。

## RE-109：任務物品交在哪裡

輸入與工具同 RE-070。探針 `tools/ida_probe_koltown_hooks.py` 倒出
`sub_11A93:0x11B20`…`0x11C33`，也就是 RE-107 那一段回報的**前面**——回報不是
進門就發生，是**交出東西**才發生。

### 條件與流程（`已證實`）

```
0x11B5A  byte_2FC01 & 4 為零 → 整段跳過        ; icon_flags 第二個 byte 的 bit 2
對六名角色（var_2 一筆 0x300）：
    角色記錄[+0x00] 為零 → 跳過
    var_A 從 0x30 走到 0x1A0，每次 +0x10       ; 身上一格十六 byte
        0x40 ≤ var_A ≤ 0x9F 跳過（`0x11B8D`／`0x11B96`）
        逐 byte 比 角色[var_A + i] 與 [0x1CF4 + 0x2E0 + i]，i 走 0…15
        十六格全同 → 把 [0x1CF4 + 0x2F0 + i] 複製回去（換成另一件）
                     再走回報那一段（`0x11C33`，RE-107）
```

`[0x1CF4]` 是**目前這棟建築的 payload** 在 DGROUP 的落點——`0x11BAE` 沒有乘任何
索引，讀的就是剛載入的那一份。所以 `0x2E0` 與 `0x2F0` 是 payload 自己的位移。

### 資料對上了（`已證實`）

拿 `icon_flags[1] & 4` 去篩七座城的 102 棟，得到 **23 棟**，而且：

- payload `0x36` 的低七位全部落在 **0…23**（缺 10，那一個在 TOWN0）
- **bit 7 只有一棟設**：TOWN7 的 `The Assembly Building`，值 0x84
  ——任務 4、而且是**最終任務**（RE-107 的門檻就是為它設的）
- `0x2E0` 起十六 byte 用 5-bit 名稱解得出來：`Truth Sword`、`Village Flag`、
  `White Quill`、`Gavel Of Oak`、`Kelder Crown`、`Parth Flask`…
- `0x2F0` 有的是另一件（`Kelder Crown` → `Flying Cloak`）、有的全零（不給東西）

RE-107 原本把 `byte_2FBFA` 對到 payload 0x36 之後看到 90／100／110／120 這種
超界的值，於是標成未解——**那是因為把 102 棟全列了**。沒有 `icon_flags[1] & 4`
的建築那一格根本不是任務編號。加上篩選就完全對上，那一段可以升成已證實。

### 身上哪幾格會被掃（`已證實`）

`var_A` 從 0x30 起每 0x10 一格，**不到 0x1A0 為止**——`0x11CF2` 的 `jge` 是先加
再比，所以最後一格是 0x190，一共十七格。跳過 0x40…0x9F（`0x11B8D`／`0x11B96`）。
角色記錄 0x30 是手上那件（RE-086）、0x100 起是穿戴欄與背包（RE-080）；
跳掉的那一段是四列武器技能（RE-080 的 `+0x52 + 13 × 列`），不是物品。

### remake 的落點

`tools/extract_town_buildings.py` 多帶四個欄位：`quest_site`、`quest_final`、
`quest_wants`／`quest_wants_bytes`、`quest_gives`／`quest_gives_bytes`。
規則層是 `internal/town/quest.go`：`CarriedOffsets`（要掃的十七格）、
`QuestFlag`／`SetQuestFlag`（四個狀態）、`TurnIn`（交付與回報）、
`CanTakeFinal`（最終任務的門檻）。

**還沒接進建築內那一層**：角色身上那十七格要先有 raw bytes 的 typed data，
抽取器目前只給 typed 的穿戴欄（`equipment`），比對要的是逐 byte 相同。

### 還沒解的

任務 10 的回報處（在 TOWN0，抽取器目前從 TOWN1 起）、
`[0xB926]`（戰鬥中逃走比的那一份）與這裡的 `0x2E0` 是不是同一件東西。
