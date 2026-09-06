# 工作歷程

## 2026-08-31：建立專案骨架

- 目標：參考 `civ1` 經驗，建立《武士傳說》重製規範、工作清單與 private GitHub 儲存庫。
- 變更：新增 `AGENTS.md`、`WORKLIST.md`、`README.md`、`.gitignore`；建立
  `wicanr2/kol-remake-cht` private 遠端。
- 來源查證：確認魔旅士 2010-04-16〈Knights of Legend 武士傳說〉，並搭配《CRPG
  通鑑》中文版評介撰寫 README 專案起源；兩者均只列為玩家觀點，不作規則 oracle。
- 安全：六個原版／參考素材檔均由 `.gitignore` 排除，未上傳 GitHub。
- 驗證：Markdown 必要契約、檔案擁有權、`git diff --check` 與 `.gitignore` 排除規則
  通過；初始提交 `b635703` 已推送。遠端 visibility 與 Docker 容器清理於收尾複核。
- 下一步：以唯讀 Docker 容器建立原版清冊與 SHA-256；尚未授權實作遊戲本體。

## 2026-08-31：固定 M1 原版 build

- 範圍：唯讀盤點 ZIP、PDF、參考圖片；建立目前狀態、研究、計畫、驗證與交接台帳。
- 變更：新增 `tools/inventory_original.py` 與 `docs/re/original-inventory.json`；固定六個
  根目錄輸入及 ZIP 內 75 個檔案的 SHA-256、CRC、timestamp、大小與 MZ header。
- 驗證：ZIP 76 項 CRC 通過；manifest 在兩個一次性 Python 3.13 容器重生後逐 byte 相同，
  SHA-256 `0f8ca5c5ed7579e7a8fb5b1355fb79599c7802db16b6e87b8b3f3aa382295b7f`；
  PDF metadata 由 Poppler 22.12.0 讀取，JPEG metadata 由 Pillow 11 讀取。
- 限制／訂正：不能從 MZ magic 或檔名證實精確發行 revision；1998 timestamp 與音訊資料
  身分仍未知，已明示而未猜補。
- 下一閘門：以 IDA Pro 9.4 盤點六個 executable 的 header、overlay、工具鏈與 driver。
- IDA 最小探針第一次失敗：JSON 已寫出，但腳本呼叫 IDA 9.4 不存在的
  `ida_kernwin.qexit`，一次性 DB 未封裝；不列為成功證據。已改用 `idc.qexit`，須在
  全新暫存輸出以相同 image／input 重跑。
- IDA 探針修正後通過：`locked-v1` image ID、UID 1000、`HOME=/home/ubuntu`、IDA 9.4、
  schema 與輸入 SHA-256 均核對；六份 JSON 由 UID 1000 寫出。四檔出現 Microsoft Quick
  C／MSC runtime FLIRT 候選，`KOLTOWN`／`TANDYPIN` 未辨識函式，未過度宣稱 compiler。
- `KOLTOWN` entry probe 第一次失敗：IDA 9.4 的 `generate_disasm_line` 位於 `idc` 而非
  `ida_kernwin`，未產生 JSON，故不列為成功證據；修正後須再用全新一次性 DB 重跑。
- 第二次 entry probe 已產生解碼，但 MZ page 欄位從錯誤 offset 讀取，導致不可能的
  `trailing_bytes=-384`；依 fail-closed 不採用。修正為從 offset `0x02` 讀 last-page／pages。
- 最終 entry probe 通過：固定 EA `0x2f522`、128 bytes、32 條有限線性解碼與前置資料；
  證實反向搬移 `0x131f` bytes 後 `RETF` 至新 segment `IP=0x0038`。81-byte 尾端全為
  `0x1A`，推翻 overlay payload 候選並訂正為 DOS EOF padding。
- 收尾驗證：inventory 再次重生逐 byte 相同；7 份 IDA JSON schema／雜湊硬閘與三支
  Python 工具語法通過。所有本輪 `/tmp/kol-*` 一次性 DB／輸出已刪除，無遺留 IDA 容器。
- 安全訂正：首次推送因 `KOLTOWN-entry.json` 含原版 bytes 而被拒絕；未繞過審查，改將
  該檔精確加入 `.gitignore`，GitHub 只保存工具、雜湊、metadata 與不含 payload 的摘要。
- 提交／推送：安全版本提交 `c6424fe` 已推送至 private `main`；遠端 visibility 複核為
  `PRIVATE`。原始 bytes JSON 仍只在本機且為 ignored。

## 2026-08-31：KOLTOWN EXEPACK 與 RE 知識庫

- 範圍：沿 stage 1 far-transfer 建立 stage 2 control flow；辨識 packer、重建研究用
  unpacked derivative，並依使用者指示先建立 RE 知識庫與 DRAFT spec。
- IDA 證據：stage 2 source packed EA `0x2f548..0x2f627` 共 104 instructions／12 direct
  xrefs；marker `Packed file is corrupt`、B0/B1 fill、B2/B3 copy、16-section relocation
  與 final transfer 共同證實 Microsoft EXEPACK，精確版本未知。
- 解包：沿用 `/home/anr2/openkb/openkb-code/src/tools/unexepack.c`（CC0，SHA-256
  `0e1da11a…b03c`），GCC 13 原樣編譯。`-Werror` 因舊式 warnings 失敗後，保留
  `-Wall -Wextra` warnings、不修改上游，產生 derivative `01f6f9a9…d1d2`。
- 獨立驗證：IDA 9.4 重載後 unpacked entry EA `0x2b90c`；強制建立入口函式得到 63
  instructions、7 call sites，整個 DB 辨識 337 functions。
- 權利邊界：解包 EXE、`.i64`、含 bytes JSON 只存本機 ignored；Git 只保存 safe summary。
- 文件：新增 `docs/re/` 知識庫索引、build／address-space／toolchain 台帳與 SPEC-001 DRAFT。
- 下一閘門：DOS runtime memory 三點核對與 garbage-byte consumer 審查；未 READY 不實作。
- 驗證／清理：5 支 Python 工具語法通過；safe summary 不含 instruction bytes／disassembly；
  本機 unpacked EXE／`.i64` 存在且由 UID 1000 擁有。本輪 `/tmp/kol-stage2-*` 與 pycache
  已刪除，無遺留 IDA／GCC 容器。
- 提交／推送：RE 知識庫、safe summaries 與 SPEC-001 DRAFT 已以 `5832a45` 推送至
  private `main`；完整 disassembly、unpacked EXE 與 `.i64` 未上傳。

## 2026-08-31：DOSBox-X runtime oracle 啟動探針

- 沿用 `rich2-dosbox-dbg:latest`（DOSBox-X 2026.06.02）與原生唯讀 bridge，新增可重生
  的 KOLTOWN 記憶體擷取探針；原始 ZIP 唯讀、執行期 dump 位於 ignored `workplace/`。
- 啟動勘誤：`-break-start` 停在 BIOS `F000:FFF0`，不是 packed MZ entry；探針改以
  EXEPACK marker 的固定 load offset 反推 DOS 載入段。
- 實測：`KOLTOWN.EXE` 與暫定 town index `0` 都在三秒採樣前退出，沒有產生 runtime
  dump；依失敗即關閉原則，SPEC-001 仍為 DRAFT，三點核對與 garbage-byte gate 未通過。
- 下一步：用 IDA 9.4 追查啟動參數 consumer／主程式 exec contract，不再猜參數；仍不
  開始 Go／Ebitengine implementation。

## 2026-08-31：SPEC-001 runtime oracle 關閉

- IDA 9.4 證實 main 於 unpacked EA `0x103b3..0x103bc` 讀 `argv[1][0]-'A'`；最小合法
  命令為 `KOLTOWN.EXE A`。完整指令保留 private，Git 只收 safe startup summary。
- 保留三項勘誤：短採樣仍在 BIOS／DOS，不是程式退出；DOS 長目錄 `cd` 未成功；MZ
  header `0x0240` 是 paragraph 數，正確 header 為 `0x2400` bytes。
- 修正掛載與位址後，第一輪量得 load segment `0824`；第二次乾淨啟動預設
  `23B3:001C` 斷點並在 C startup 前精確命中。
- runtime dump 129,312 bytes、SHA-256 `df79480d…67f6`。2,281 個 relocation 契約全部
  通過；正規化後全映像零差異，五個分散 32-byte 錨點全吻合。
- SPEC-001 由 DRAFT 升 READY，只授權 derivative 當研究 oracle；下一步進圖像格式 RE／
  PNG 抽取 DRAFT spec，仍不開始 production runtime。

## 2026-08-31：圖片 archive 與 SPEC-002 DRAFT

- 新增固定 ZIP 圖片清冊：42 個候選中 30 個符合 20-byte 固定區、2-byte offset table；
  共 960 個 payload，offset 全部嚴格遞增且有界。
- IDA 9.4 非破壞性探針定位 `sub_1A429`／`sub_1A4FE` 索引 consumer、`sub_1A04E`
  LZW core 與 `sub_1A26B` LSB-first bit reader；完整指令留本機 private JSON。
- 一次性 LZW 探針對 960 個 payload 全部完成機械解碼，輸出只含大小與 SHA-256；尚未
  證實尺寸、位平面、palette、透明色或畫面方向，不產生 production PNG。
- 發現 `AWD1PICS.0` 索引 payload 後有 520 bytes，保留為未知，不當作 padding。
- 新增 `image-formats.md` 與 SPEC-002 DRAFT。下一步完成單一 consumer 的來源到 blit
  垂直鏈與原版視覺交叉驗證；READY 前不接 Go／Ebitengine。

## 2026-08-31：COMBPICS EGA 垂直鏈

- IDA 9.4 證實 KOL launcher 選 BIOS mode 0Dh，KOLTOWN suffix 表為 PICS／PCGA／HERC；
  COMBPICS consumer 使用 37 byte-columns×144 scanlines，即 296×144。
- KOL 第一次新建 DB 的 video-I/O probe 未等待 autoanalysis，錯誤輸出空 hits；原始檔仍有
  三個 `INT 10h` opcode 候選，故空摘要作廢。加入 `ida_auto.auto_wait()` 後以全新 DB
  重跑，得到 3 個 BIOS video 與 6 個 direct-port hits。
- `sub_1C907`／`sub_1A2BB` 二階格式已解出：big-endian stream 分界、high-nibble-first、
  repeat／literal RLE、4-bit 色彩與四 EGA bitplanes。第一次 probe 把兩種 control 分支
  對調而越界，未產生 PNG；回查原始指令後修正並重跑。
- 前三張 ignored PNG 均通過 296×144、4-bit bounds 與非空色彩；其中 index 0／1 分別
  精確對上魔旅士文章的逃跑／戰敗同場景 JPEG。
- 固定 crop 的正配 RGB MAE 為 20.12／19.17，交叉錯配為 62.00／62.88。外部圖片只留
  ignored 工作區；Git 保存 URL、雜湊與矩陣。
- Palette 依 DOSBox-X pinned commit `63a4357` mode 0Dh／EGA table，並標記為
  hardware-spec approximation；KOL／TITLE／KOLTOWN IDA-defined code 未見 palette
  寫入，MAKEMODE 無 video I/O。不深入 DOS 硬體逐週期或類比色彩。
- SPEC-002 仍為 DRAFT。下一步查 AWD1 尾端 520 bytes 與三個無副檔名 runtime 檔。

## 2026-08-31：AWD 與角色圖片固定記錄表

- KOLBOOK 第一版 asset-I/O probe 誤用不存在的 `idc.get_segm_class`，未產生可採證輸出；
  改用 `ida_segment.get_segm_class` 後，以新資料庫重跑成功。
- IDA 證實 KOLBOOK 的 AWD PICS／PCGA／HERC buffer 均由 `sub_137C0` 固定寫出 4000
  bytes；`sub_1374E` 是最大 65535 bytes 的整檔讀取 wrapper。KOLTOWN 的 AWD 玩家路徑
  只經索引 loader 消費 PICS payload，故尾端 520 bytes 排除於圖片流，但不稱為 padding。
- KOLBOOK 與 KOLTOWN 的 reader／writer 證實 `CHARPICS`、`SHLDPICS`、`STEEDPIC` 是
  16×760-byte 的可變角色圖片記錄表；內部像素佈局仍未知。
- 圖片候選規則補入 AWD PCGA／HERC，由 42 增為 44；30 個索引容器數量不變，三個
  無副檔名檔均輸出 16×760-byte 記錄雜湊。清冊在同一無網路容器重生兩次逐 byte 相同，
  SHA-256 `299ce917…75e4`。
- KOLBOOK 安全摘要 SHA-256 `17f26ce0…dddb`；安全摘要檢查證實沒有 `bytes`、完整指令或
  disassembly 欄位。所有寫入檔由 UID/GID 1000:1000 擁有。
- SPEC-002 仍為 DRAFT；下一個唯一工作改為完整失敗政策、命名、manifest schema 與
  權利審查，READY 前仍不實作 production PNG importer 或 Go runtime。
- 新增 `kol-png-asset-manifest-v1` JSON Schema，固定來源追溯、證據等級、排除狀態、PNG
  欄位與權利分類；規格同時定義原子批次、命名碰撞與所有解碼錯誤的失敗即關閉政策。
- 權利審查維持原版衍生 PNG 僅限本機／授權 private 位置；private GitHub 不等於取得公開
  散布權。此 READY gate 已關閉，但其餘像素格式仍未證實，因此 SPEC-002 保持 DRAFT。
- 收尾前已刪除本輪 Python bytecode、KOLBOOK 暫存副本與 KOLBOOK／KOLTOWN 暫存 IDA
  資料庫／log；`docker ps -a` 未見 IDA 或 kol 專案容器。本輪沿用映像，未建立待清理映像。

## 2026-08-31：TWN 分卷清冊勘誤與 continuation 像素鏈

- 發現 v1 清冊把 `TWN*PICS.1/.2` 誤當獨立未知格式，實際 `sub_1A4FE` 由 `.0` header
  的累計門檻、per-volume last-length 與全域 offset table 選取最多五卷。
- 清冊升 v2：44 個實體候選歸併 30 個邏輯 archive；TWN0..6 為兩卷、TWN7 為三卷，
  payload 總數由 960 訂正為 1,022。舊數字保留於歷史並明寫成因，不靜默改寫。
- LZW 摘要升 v2，新增 volume name／index／offset；1,022 個 payload 全部機械解碼成功。
  inventory 與 LZW 摘要均在同一無網路容器重生兩次逐 byte 相同，SHA-256 分別為
  `367b28a2…a4f0`、`9cf2f5ee…3ba1`。
- `sub_2AAA8` 的常數圖號 24 經 `sub_11302`／`sub_10F9C` 完成 continuation → LZW →
  二階 RLE → 4-bit EGA → blit；八組 TWN 都落在 `.1`。town 0 為 192×96，其餘為 88×64。
- 探針最初把 `word_2DF2C` 誤讀成顯示 mode，先預期全部 88×64；town 0 實測 18,432
  pixels 推翻此解釋。回查 string formatter 與分支後訂正它是 town index，改採 town-0
  特例；第二次又以固定 192×96 套到 town 1 而失敗，最終逐 town 契約全數通過。
- 八張 PNG 兩次重生逐 byte 相同；接觸表目視方向正常、內容非空。語意未證實，工具由
  `event` 改名為中性的 `index24`。安全摘要 SHA-256 `642a68c6…ed54`，PNG 不入 Git。
- SPEC-002 仍為 DRAFT；本輪只新增 DRAFT 證據與可丟棄 probe，未建立 production importer。
- 跨台帳驗證確認 30 archive／1,022 outputs 的圖號、volume、檔名、offset 與長度逐筆
  一致；Pillow 重讀八張 PNG 均為預期尺寸、RGB 且非空。Python syntax、全部 JSON parse、
  `git diff --check`、安全摘要 deny-list 與 UID/GID 抽查通過。
- 收尾刪除 Python bytecode 與第二份重複 PNG 目錄；`docker ps -a` 未見 KOL、IDA 或
  Pillow 遺留容器。本輪沿用既有映像，未建立或取代 Docker image。

## 2026-08-31：MAKEMODE 顯示 theme 範圍查證

- 以固定 SHA-256 的 MAKEMODE.EXE 建立全新 IDA 9.4 DB；`locked-v1` image ID、UID 1000、
  `HOME=/home/ubuntu` 與 `idat` 入口先通過最小探針。
- 第一輪匯出定位 mode 提示、`%spics.0`／`%spics.%d` 與 `_main`／`sub_10464`；第二輪
  以另一新 DB 加入 I/O、LZW、二階轉換與重壓縮 helper；第三輪加入 28-call catalog；
  第四輪把安全摘要的長篇原版文字改成 length＋SHA-256。每輪都用新 DB，未污染正式 i64。
- 證實官方工具由 PICS 生成 `basepcga.N` 或 `baseherc.N`。28 組 base／count／transform
  case 已機械匯出，所有 count 與 image inventory v2 逐組一致。
- 靜態資料流已閉合到 PICS header／分卷 → LZW decode → per-case geometry → EGA planes →
  packed 2-bit → recompress → checked write；尚無官方工具動態輸出 round-trip，故完整像素
  轉換維持強推論。
- 此證據把「PCGA／HERC 素材缺失」訂正為「官方可生成衍生 theme」。是否列為首版必要
  交付會實質改變 scope，依 grilling 技能留待使用者選擇，未寫入正式 extraction set。
- 安全摘要 SHA-256 `a5ca5e13…e0a7`；Python syntax、全部 JSON parse、28 組 count 對
  inventory v2、長文不入版控 deny-list、原始 bytes／disassembly deny-list 與 UID/GID
  抽查均通過。
- 收尾已刪除四個一次性 MAKEMODE `.i64` 與 EXE 暫存副本；private JSON 仍由精確
  `.gitignore` 規則排除。`docker ps -a` 未見 KOL／MAKEMODE／IDA 遺留容器，本輪未建立
  或取代 Docker image。

## 2026-08-31：MONPICS 彩圖／遮罩垂直鏈

- 從 KOLTOWN 既有 consumer 清冊選定 `MONPICS`；`sub_22EFC` 與 `sub_2A975` 共同固定
  DGROUP offset `0x1370` 的尺寸表、怪物索引與兩筆 payload 範圍。
- 首次 IDA 探針誤設 64 個怪物，零尺寸硬閘失敗；依下一表 `0x13b0` 與
  `[2*i,2*i+1]` consumer 訂正為 32 個怪物、64 個 payload，保留失敗原因而不覆寫歷史。
- 第一版 PNG probe 對全部 payload 套 COMBPICS 二階 RLE，在奇數 index 1 的非法分界
  失敗。回查第一階長度後，訂正偶數為 4-bit 彩圖、奇數為 1-bit 遮罩。
- 32 組彩圖均精確符合 IDA 寬高，遮罩有效長度均為 `width_bytes*height`；MSB-first
  展開後合成 32 張 RGBA PNG。接觸表人工抽查方向、輪廓與透明邊界正常。
- 遮罩 LZW 輸出各多一個未消費 byte，現只標為強推論的終止／padding artifact；探針
  保存其 SHA-256，不靜默吞掉，也不當成像素。
- 新增安全摘要 `KOLTOWN-monpics-summary.json`、`monpics-ega-visual-probe.json` 與兩支
  可重生研究工具；完整指令、table bytes 與所有 PNG 保持 ignored。
- SPEC-002 只新增 DRAFT 證據，仍未授權 production extractor 或 Go／Ebitengine runtime。
- IDA 摘要以第二個全新 `/tmp` DB 重生逐 byte 相同；32 張 RGBA PNG 與接觸表第二次重生
  逐 byte 相同。首次失敗留下的單一舊命名 PNG 已精確刪除，未清理其他專案資料。
- 安全摘要 SHA-256 為 `1a57c085…7af9`，視覺統計為 `cf2e1287…73d6`，接觸表為
  `54bc3863…a10`。全部 JSON parse、跨台帳尺寸／像素數、RGBA PNG header、Python syntax、
  安全摘要 deny-list、UID/GID 1000:1000 與 `git diff --check` 均通過。
- 本輪暫存 rerun JSON／PNG 目錄已刪除，沒有 `__pycache__` 或懸空 image；`docker ps -a`
  未見 KOL／IDA／MONPICS 容器。清單中的其他執行中／已建立容器不屬本專案，未擅自停止。

## 2026-08-31：中文手冊管線與首批 reviewed 頁

- 依使用者要求唯讀參考 `~/cht/civ1` 的手冊翻譯經驗，沿用既有
  `civ1-manual-pdf:local`，未建立重複 image。採其逐頁 hash、`todo → translated →
  reviewed`、機械 gate＋人工逐頁對照與英文全文不入版控的契約。
- 固定玩家手冊 79 個 PDF 實體頁、參考卡 5 頁；Poppler `pdftotext 22.12.0 -raw` 建立
  84 頁清冊。清冊 SHA-256 `bd27f093…6131`，兩次重生與本機英文 TSV 均逐 byte 相同。
- 初次清冊的 `all_text_layers_nonempty` 因玩家手冊 76、77 只有換頁字元而失敗。以
  `pdftoppm` 120 DPI PNG 目視確認兩頁是真正空白跨頁、只有背頁透印；工具改為明示
  `expected_blank`，其餘 82 頁全部具文字層，不啟動不必要 OCR。
- 新增 SPEC-003 READY、手冊入口、逐頁索引、術語表、狀態 ledger、來源清冊與兩支
  Docker-only 工具。英文全文、掃描頁 PNG 與原 PDF 均留在 ignored 工作區。
- 玩家手冊實體頁 1–5 已翻譯並逐頁對照掃描影像。影像對照抓到實體頁 3 的 OCR 錯序：
  raw 文字把矮人／凱爾登／精靈職業群拆開，初稿目錄頁碼錯配；已依掃描訂正並保留成
  SPEC-003 的正對照。
- 目前狀態為 reviewed 7 頁（內容 1–5、空白 76–77）、translated 0、todo 77。gate 檢查
  key／hash、狀態覆蓋、數字、篇幅、連續英文與 Markdown 連結；狀態重疊及漏數字兩個
  正對照都會非零失敗。
- 翻譯狀態 ledger SHA-256 `e12d8f49…a401`；全部 JSON parse、安全清冊不含英文 `source`
  欄位、Python syntax、全 `docs/` 本地 Markdown 連結、UID/GID 1000:1000 與
  `git diff --check` 均通過。
- ignored 工作區保留實體頁 1–5、76–77 的校對 PNG 與英文逐頁 TSV；沒有重生副本、
  `__pycache__` 或懸空 image。短暫顯示的 `alpine:3.22` Created 容器在 inspect 前已由
  `--rm` 自動消失；重查 `docker ps -a` 未見 KOL／PDF 容器，其他專案容器未動。

## 2026-08-31：中文手冊實體頁 6–10

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 6–15，新增城鎮操作章節；逐頁保留 PDF 實體頁
  marker，涵蓋旅店、角色資料、裝備、物品攜帶、居民交談、同伴招募與城鎮移動。
- 以 100 DPI ignored PNG 逐頁核對掃描；校正 OCR 的 `NUlRITION`、`I/you`、
  `ll-T\\JRN` 等錯字，並把 14 個原版圖示英文名及核准繁中功能名加入共用術語表。
- 初次 gate 報 page 007 缺少數字 `11`；上下文證實它只來自圖示分隔線
  `--1•11-----`／`--11!11-----`。驗證器改為只排除「無字母且含裝飾符號」的 OCR 行；
  正對照確認純數字 `11` 仍會被檢查，沒有用寬泛例外掩蓋目錄頁碼。
- 人工頁界審查發現 Bronk 回答跨實體頁 9→10，初稿卻把完整句子放在 page 009 區塊；
  已以雙端省略號訂正，恢復原文定位。這類段落歸屬無法由數字／篇幅 gate 取代。
- 現況為 reviewed 12（內容 1–10、空白 76–77）、translated 0、todo 72；翻譯 gate 與
  圖示裝飾正對照通過。下一批從實體頁 11–15 開始，不重開已 reviewed 頁。
- 本批章節 SHA-256 `5e91d94d…3e2f`，狀態 ledger `bdd17905…c577`；全部 JSON parse、
  Python syntax、全 `docs/` 本地 Markdown 連結、`git diff --check` 與 ignored 素材邊界
  通過。本輪沒有 `__pycache__`、懸空 image 或 KOL／PDF 容器；其他專案容器未動。

## 2026-08-31：中文手冊實體頁 11–15

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 16–25，新增〈在盔甲店〉與〈進入荒野〉兩章；
  涵蓋盔甲購買與合身、物品轉交與出售、旅店保存、道路旅行、伏擊、地形、天候及營養。
- 以 ignored 掃描頁逐頁核讀，保留實體頁 13 的跨章歸屬，以及 13→14、14→15 的跨頁句界；
  `Coming to Blows` 依掃描影像校正，不沿用 OCR 的 `Corning to Blows`。
- 翻譯 gate 初次發現實體頁 14 的 `you'll` 被 OCR 認成 `you'11`。驗證器只針對英文代名詞
  加縮寫的精確模式正規化，不放寬真正數字檢查；各頁篇幅不足則回查原文，補回遺漏的
  交易操作、道路安全、遭遇可能是治療者與地形細節。
- 現況為 reviewed 17（內容 1–15、空白 76–77）、translated 0、todo 67；全 84 頁狀態
  完整覆蓋，五個 Markdown 輸出均有逐頁來源定位。下一批從實體頁 16–20 開始。
- 兩章 SHA-256 分別為 `8b4fef46…d81a`、`95cafb57…7b1`，狀態 ledger 為
  `c0d7bb0f…3198`；Python syntax、全部 JSON、全專案本地 Markdown 連結、UID/GID、
  `git diff --check` 均通過。`docker ps -a` 未見 KOL／手冊／PDF 專案殘留容器。

## 2026-08-31：中文手冊實體頁 16–20

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 26–35；第 16 頁分別接續荒野營養章並開始
  〈短兵相接〉，其餘頁面涵蓋戰鬥畫面、逐回合命令、地面／飛行移動、四類攻擊、目標
  選擇、武器與徒手招式、瞄準部位、七種防禦、十字弓裝填及武器切換。
- 逐頁對照 100 DPI 掃描，確認章名為 `COMING TO BLOWS`、圖示列與文字欄閱讀順序；
  `Kelderheit`、`FIRE` 被 OCR 錯認成其他字形時，均以掃描及上下文校正。
- 共用術語表新增戰鬥者視窗、戰鬥地圖、移動速度、攻擊、防禦、瞄準、裝填與切換等詞；
  尚需未來遊戲 UI consumer 決定的招式中文名維持「暫譯」，不因手冊翻譯升格為 runtime 規則。
- 現況為 reviewed 22（內容 1–20、空白 76–77）、translated 0、todo 62；全 84 頁狀態
  完整覆蓋，六個 Markdown 輸出均通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 21–25 開始。
- 更新後荒野章、戰鬥章與狀態 ledger SHA-256 分別為 `02e58937…f175`、
  `15879db6…62f9`、`c6c838e5…9b32`；Python syntax、全部 JSON、本地 Markdown 連結、
  UID/GID 與 `git diff --check` 均通過；`docker ps -a` 未見 KOL／手冊／PDF 專案殘留容器。

## 2026-08-31：中文手冊實體頁 21–25

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 36–45；完成〈短兵相接〉後半，涵蓋武器切換、
  戰場物品撿取、疲勞恢復、命令審閱、非輪流式行動順序、傷勢與箭矢管理、三種戰鬥結局、
  戰利品分配、逃跑損失與全隊敗北。
- 新增〈回到城鎮〉，翻譯戰利品交易、武器技能訓練、城鎮魔法服務、酒館飲食與新裝備；
  第 24→25 頁訓練句界與第 25→26 頁新裝備句界均按實體掃描保留省略號。
- 逐頁對照 100 DPI 掃描，確認 `PICK UP`、`PANIC`、`TRAIN` 圖示及原文數字；Hvrad Myth、
  Astimiah Eckhart、Lonely Page 與 Orc 未由 OCR 或人工偏好定案，維持待查／暫譯。
- 現況為 reviewed 27（內容 1–25、空白 76–77）、translated 0、todo 57；全 84 頁狀態
  完整覆蓋，七個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 26–30 開始。
- 更新後戰鬥章、回城章與狀態 ledger SHA-256 分別為 `708fe325…bc73`、
  `374be98e…e788`、`4613eb63…9956`；Python syntax、全部 JSON 及本地 Markdown 連結
  均通過；UID/GID 與 `git diff --check` 正常，`docker ps -a` 未見 KOL／手冊／PDF
  專案殘留容器。

## 2026-08-31：中文手冊實體頁 26–30

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 46–55；完成回城章的新裝備、修道院完整治療與
  旅店保存流程，並新增〈進階玩家〉章的角色職業與主要屬性說明。
- 第 27–29 頁共 35 筆男性／女性人類、精靈、男性矮人及男性凱爾登職業資料，全部按
  100 DPI 掃描影像重建 Markdown 表格；保留職業專名、七項屬性、起始金冠、武器縮寫與
  進攻／防禦技能值，不採用 OCR 的雙欄錯序與 `L0`／`LO` 混淆。
- 第 29–30 頁翻譯力量、敏捷、體型、健康、先見、魅力及智力；說明書公開描述只進入
  文件證據層，未據此實作隱藏公式。Balance、Endurance、Body Points 等尚待 consumer
  複核的詞維持暫譯。
- 現況為 reviewed 32（內容 1–30、空白 76–77）、translated 0、todo 52；全 84 頁狀態
  完整覆蓋，八個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 31–35 開始。
- 更新後回城章、進階玩家章與狀態 ledger SHA-256 分別為 `0317b2a9…e8e6`、
  `da746f3f…fe97`、`3b834824…be77`；Python syntax、全部 JSON 與本地 Markdown 連結
  均通過；UID/GID 與 `git diff --check` 正常，`docker ps -a` 未見 KOL／手冊／PDF
  專案殘留容器。

## 2026-08-31：中文手冊實體頁 31–35

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 56–65；完成智力、平衡、耐力與身體點數，並
  翻譯疲勞成本、戰鬥選項／執行、命中、攻防修正、部位、傷害、擊退與回合結束等公開描述。
- 第 32–33 頁按 100 DPI 掃描重建 23 種近身武器、5 種飛射武器、3 種盾牌、9 種盔甲及
  五個武器分類比較表，保存攻擊類型、負重、傷害、手數與防護值；校正 OCR 雙欄交錯、
  `Dmi:`／`Hods` 等欄名錯誤。
- 戰鬥段落只記錄說明書玩家可見契約；隨機數來源、精確修正順序及 executable 隱藏公式
  未升格為已證實，也未進入 production runtime。新增疲勞、擊退與防護術語。
- 現況為 reviewed 37（內容 1–35、空白 76–77）、translated 0、todo 47；全 84 頁狀態
  完整覆蓋，八個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 36–40 開始。
- 更新後進階玩家章與狀態 ledger SHA-256 分別為 `4b8c8458…31df`、`885f9c7d…5090`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過；UID/GID 與 `git diff --check`
  正常，`docker ps -a` 未見 KOL／手冊／PDF 專案殘留容器。

## 2026-08-31：中文手冊實體頁 36–40

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 66–75；逐項保存從農民、平民、紳士、侍從到
  騎士的 25 個階級與說明，並翻譯競技場晉升條件及一對一、不可撤退的玩家可見契約。
- 依 100 DPI 掃描重建精靈語法術構詞：固定 `種族＋屬性＋強度＋距離＋（持續時間）＋亞種`
  順序、8 個種族類別音節、8 個屬性音節、5 級強度、2 種距離、5 種持續時間、29 個亞種
  映射及 6 個完整範例；原始魔法詞保持不翻譯。
- 修正 OCR 對 Orc／Great Orc、數個雙欄亞種與字母 `O`／數字 `0` 的混淆；教團、護符、
  修改法術與印刷頁 129 交叉引用也已保留。音節表是手冊證據，不直接授權 runtime parser。
- 現況為 reviewed 42（內容 1–40、空白 76–77）、translated 0、todo 42；全 84 頁恰好一半
  已 reviewed，八個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 41–45 開始。
- 更新後進階玩家章與狀態 ledger SHA-256 分別為 `40071778…d8e3`、`41ca879d…8f08`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過；UID/GID 與 `git diff --check`
  正常，`docker ps -a` 未見 KOL／手冊／PDF 專案殘留容器。

## 2026-08-31：中文手冊實體頁 41–45

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 76–85；完成法術選擇、SCROLL、任務、MEDAL 與
  原版規劃的擴充區域操作，並新增〈阿什塔拉瑞亞王國〉世界史章。
- 世界史依 100 DPI 掃描保留 Cullen 敘事框架、矮人出土與 Ashtel Arra、精靈抵達、
  Ghor／Zolod 兩族、三次食人魔戰事、人類跨越 Tyme、城鎮建立、公國化、Crelek 遠征與
  Walbar 抵達；938、952、963、962、972、977、1003 等年代均通過逐頁數字 gate。
- 地名與人名未由 OCR 自行音譯；Ashtel Arra、Walbar 等維持待查，Day of Lighting 暫譯
  「照明日」。手冊世界觀與任務描述仍不直接授權 runtime 或新增不存在的擴充內容。
- 現況為 reviewed 47（內容 1–45、空白 76–77）、translated 0、todo 37；全 84 頁狀態
  完整覆蓋，九個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 46–50 開始。
- 更新後進階玩家章、世界史章與狀態 ledger SHA-256 分別為 `abda34f2…0251`、
  `b033d86f…21d5`、`131693d0…15d7`；Python syntax、全部 JSON 與本地 Markdown 連結
  均通過；UID/GID 與 `git diff --check` 正常，`docker ps -a` 未見 KOL／手冊／PDF
  專案殘留容器。

## 2026-08-31：中文手冊實體頁 46–50

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 86–95；完成 Walbar 貿易興衰、Poitle 海盜與
  黑暗戰爭完整戰史，並新增〈傳說種族〉章的 Blacksteel 敘事及 Brettle Regular 起點。
- 依 100 DPI 掃描保留 1023、1026–1044、1042–1050 年代，Dom Zolod 淹城、Raven's Point、
  Olegar's End、Crelek 率 2,000 名矮人歸來、六軍 Shellernoon 決戰與戰後十五年重建；
  原文 ’43／’44／’48 簡寫亦在譯文並列，避免逐頁數字失去定位。
- 翻譯凱爾登七年前降臨、今日阿什塔拉瑞亞威脅與 Usip 繼承危機；Kelderheit／Kelder、
  Blacksteel 等專名維持暫譯或待查，不由 OCR 定案。世界史不直接成為 runtime 行為。
- 現況為 reviewed 52（內容 1–50、空白 76–77）、translated 0、todo 32；全 84 頁狀態
  完整覆蓋，十個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 51–55 開始。
- 更新後世界史章、種族章與狀態 ledger SHA-256 分別為 `1f240cce…6126`、
  `22e44fa8…4b09`、`23debd18…dbfd`；Python syntax、全部 JSON 與本地 Markdown 連結
  均通過；UID/GID 與 `git diff --check` 正常，`docker ps -a` 未見 KOL／手冊／PDF
  專案殘留容器。

## 2026-08-31：中文手冊實體頁 51–55

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 96–105；接續 Brettle Regular，完成 Dark Guard、
  Drezin Ranger、Duke's Highwaymen、Hobean Squire、Htron Pirate、Krag Barbarian 與
  Krell Warrion，並開始 Lintle Plainsmen。
- 逐頁對照 100 DPI 掃描，確認 OCR 誤讀的 `HTRON PIRATE`、`KRELL WARRION`、
  `Paraman-tathon U'Hi` 等標題與專名；跨頁段落以雙向刪節號保留來源邊界。
- 新專名維持待查，不讓手冊敘事直接授權 runtime 規則；Dark Guard 與 Pildar 的關係、
  Sledge 食樹、Krell Warrion 傳聞等僅記為世界觀文件證據。
- 現況為 reviewed 57（內容 1–55、空白 76–77）、translated 0、todo 27；全 84 頁狀態
  完整覆蓋，十個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 56–60 開始。
- 更新後種族章與狀態 ledger SHA-256 分別為 `57c4e947…6ef1`、`7f605dd2…eea5`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：中文手冊實體頁 56–60

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 106–115；完成 Lintle Plainsmen、
  Olanthan Hunter、Poitle Rogue、Shellernoon Watchmen、Ghor Tigress、
  Lintle Plainswoman、Olanthan Huntress 與 Tegal Amazon，並開始矮人及其採礦職業。
- 逐頁對照 100 DPI 掃描，確認 OCR 的 `Ores` 實為 `Orcs`、`aTroll` 實為 `a Troll`，
  並保留 `48` 小時、七年與逾十年等可機械複核數字。
- `Great Oath` 原文明示外人不知內容，維持待查；各文化敘事只作手冊世界觀證據，
  不直接授權職業、AI、戰鬥或其他 runtime 行為。
- 現況為 reviewed 62（內容 1–60、空白 76–77）、translated 0、todo 22；全 84 頁狀態
  完整覆蓋，十個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 61–65 開始。
- 更新後種族章與狀態 ledger SHA-256 分別為 `93c7d71e…fe47`、`49d3d92f…f422`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：中文手冊實體頁 61–65

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 116–125；完成矮人 tunnelers、
  Ghor Spiderguard／Mytrone Ratguard、Ghor Trollbane／Mytrone Orcbane、
  Ghor Militia／Mytrone Levie，並涵蓋凱爾登三階級與精靈 Brek、Klvar、Melod。
- 逐頁對照 100 DPI 掃描，保留原書 `Levie` 拼法，訂正 OCR 的 `Ores` 為 `Orcs`；
  Klakar 五天、Kelkret 十五天與遠行數千里等數量均保留並通過數字 gate。
- 凱爾登文化是 Blacksteel 有限觀察，明示三階級只是當時所知的簡化；手冊內容不直接授權
  runtime 的職業、儀式、語言或戰鬥規則。
- 現況為 reviewed 67（內容 1–65、空白 76–77）、translated 0、todo 17；全 84 頁狀態
  完整覆蓋，十個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 66–70 開始。
- 更新後種族章與狀態 ledger SHA-256 分別為 `091b2d0e…dc7d`、`54f85c21…3449`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：中文手冊實體頁 66–70

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 126–135；完成 Melod、Pyar、Thism、Usip，
  並開始〈精靈語〉附錄的歷史、發音、三家族、十五根詞、法術構詞及詞彙表。
- 逐頁以 100 DPI 掃描校正附加符號，保留 `Anfël`、`Kafarifël`、`ëavorfël`、`ANÓN`、
  `KUTWË` 與亞類數字音節；OCR 把 `Zebin` 誤作 `7'ebin`，已在譯文留下可追溯校勘。
- 精靈語表格是手冊語料證據，不直接授權 runtime 法術 parser；現代／古代屈折差異及十二字元
  上限仍須由遊戲 consumer 與原版行為另行審查。
- 現況為 reviewed 72（內容 1–70、空白 76–77）、translated 0、todo 12；全 84 頁狀態
  完整覆蓋，十個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。下一批從
  實體頁 71–75 開始。
- 更新後種族章與狀態 ledger SHA-256 分別為 `34050961…eee9`、`ffaf1db1…e130`；
  Python syntax、全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：中文手冊實體頁 71–75

- 依 SPEC-003 READY 翻譯玩家手冊印刷頁 136–144，新增年表、完整索引、製作名單與
  ORIGIN 原版九十日有限保固的歷史中文紀錄。
- 年表逐項保存 604–1100 年代及 `686？`、`980？` 等原書不確定標記；索引維持原書印刷
  頁碼，沒有錯改成 Markdown 頁碼。OCR 的 `HXX>` 依掃描訂正為 `1,000`。
- 保固章明示只是 1989 年原版條款翻譯，不代表重製專案承諾或法律建議；五美元、九十日與
  原書郵政地址均保留供來源核對。
- 現況為 reviewed 77（內容 1–75、空白 76–77）、translated 0、todo 7；全 84 頁狀態
  完整覆蓋，十一個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。玩家手冊只剩
  實體頁 78–79，另有參考卡五頁。
- 新增章與狀態 ledger SHA-256 分別為 `6bf476db…4e14`、`5017bfc8…e884`；Python syntax、
  全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：中文玩家手冊實體頁 78–79 完成

- 依 SPEC-003 READY 翻譯玩家手冊作者介紹與封底識別；第 78 頁掃描為左側空白、右側
  Todd Mitchell Porter 介紹，不能把整個實體頁誤標為空白。
- 保留 1981、1987、四兄弟、兩姊妹、Central College、Penguin Software、ORIGIN 與
  《Ultima》等可核對內容；第 79 頁記錄地址及產品碼 `ORI-3123`，但不宣稱商標授權。
- 封底文字層數字 `11` 經 120 DPI 掃描證實為商標圖形 OCR 誤辨，已留下校勘而未升格為內容。
- 玩家手冊 79 個實體頁至此全部 reviewed；全 84 頁套件現為 reviewed 79、todo 5，剩下
  參考卡五頁。十一個 Markdown 輸出通過來源 hash、數字、篇幅與英文殘留 gate。
- 更新後末章與狀態 ledger SHA-256 分別為 `a1fd4cf1…504e`、`fa0ba16d…7278`；Python syntax、
  全部 JSON 與本地 Markdown 連結均通過。

## 2026-08-31：Apple II 參考卡與 SPEC-003 完成

- 翻譯 Apple II 參考卡五個 PDF 實體頁，涵蓋 ProDOS `/CHARS`、MA／MB／A1／A4／A6
  換片、角色建立、圖像／盾牌編輯、城鎮／荒野移動、戰鬥、保存與完整圖示速查。
- 文件頂端明示這是原版 Apple II 操作，不是 Go／Ebitengine remake 安裝說明；產品碼、
  商標與 1989 著作權只作歷史記錄，不表示重製散布授權。
- 逐頁以 120 DPI 掃描校正；第 4、5 個 PDF 頁各合併原卡印刷頁 5–6、7–8，文字層把圖示
  裝飾誤辨為 `1111`，已留下校勘而未當作操作資料。
- 全套 84 頁現為 reviewed 84、translated 0、todo 0，12 個 Markdown 輸出通過來源 hash、
  數字、篇幅、英文殘留、JSON 與本地連結 gate；SPEC-003 依自身完成條件升為 CONFORMED。
- 參考卡、狀態 ledger 與 SPEC-003 SHA-256 分別為 `c6101885…4cce`、`bc0709c9…8c49`、
  `edc11b66…525b`。原始 PDF、英文 TSV 與掃描 PNG 仍僅在 ignored 本機工作區。

## 2026-08-31：PERPICS EGA 彩圖／遮罩分支

- 第一版 IDA 探針未等待 auto-analysis，只產生 `.i64` 而沒有 JSON；不列成功證據。
  加入 `ida_auto.auto_wait()` 後以全新 DB 重跑，再擴充追查 `sub_1CBA9`、`sub_16FA6`、
  `sub_2D3FA` 與 DGROUP table；每次修改均不用既有 DB 冒充重現。
- 修正舊 fingerprint 限制：near-offset 命中不足以證實 consumer。本輪由
  `sub_18F3A:0x18fdd` 的 `per%s` 格式化資料流、`0x18ff3` 的 `[i,i+6]` 配對，以及
  `sub_16FA6` 兩次 loader call／遮罩 blit，關閉真正垂直鏈；`sub_1CBA9` 排除為媒體提示。
- 六組 IDA 矩形與 12 個 payload 全部吻合：前六筆展開為 4-bit EGA pixels，後六筆為
  MSB-first 1-bit mask 加一個非像素零 byte。六張 ignored RGBA PNG／接觸表目視方向、
  plane 與透明邊界正常；未推測人物身分。
- 新增安全 IDA 摘要、視覺 probe 統計與可重生探針；SPEC-002 只新增已證實 PERPICS
  窄分支，整體仍為 DRAFT，未建立 production importer 或 Go runtime。
- 同一輸入第二次重生的安全摘要與 PNG 接觸表逐 byte 相同；接觸表 SHA-256
  `4ada9b74…6ec3`。Python syntax、全部 JSON、本地 Markdown 連結、UID/GID 與
  `git diff --check` 均通過；private JSON SHA-256 `4e2e9cb0…ee3b` 只作本機收據。

## 2026-08-31：KOLMPICS 圖號 28 EGA AND-mask

- 由 `sub_10000` 的 `kolm%s.0` whole-file load、`word_39820` buffer，追到
  `sub_14047:0x1415c` 以常數 `0x1c` 呼叫 `sub_1A429`；只關閉圖號 28，不泛化 83 筆。
- IDA 矩形固定為 7 byte-columns×57 scanlines，即 56×57。LZW 輸出 400 bytes，前
  399 bytes 以 MSB-first 展開精確填滿矩形，最後一個零 byte 不作像素。
- 初版預覽錯套其他 archive 的 1=opaque，畫面幾乎全空，未採用。回查
  `sub_1BD17:0x1bf4c..0x1bf53` 的 `(destination AND mask) OR color` 後，訂正為 bit 1
  保留背景／透明、bit 0 覆寫／不透明；bit-map 目視為連續切角與圓弧遮罩。
- 新增全新 DB 可重生 IDA 探針、安全摘要、ignored PNG probe 與統計；SPEC-002 仍是
  DRAFT，未實作 production importer 或 Go runtime。
- 兩次全新 IDA DB 的 private／safe JSON 逐 byte 相同，SHA-256 分別為
  `7c39c3df…f25b`、`3f600d9b…716`；兩次視覺摘要與兩張 PNG 也逐 byte 相同，視覺摘要
  SHA-256 `d2b8b32e…ac34`。
- Python syntax、全部 JSON、本地 Markdown 連結、版控摘要與重生摘要、UID/GID、
  `git diff --check` 均通過。本輪三個 `/tmp/kol-kolmpics28-*` 與重複 PNG 目錄已刪除；
  Docker inspect 未見掛載 KOL 工作樹的殘留容器，其他專案容器未動。

## 2026-08-31：MISCPICS 完整 archive 分類

- 先用 discovery probe 盤點 `sub_1A429` caller 與預載 buffer；確認 MISCPICS 正常鏈是
  `sub_1103F → sub_11262 → sub_1A429 → sub_19E9D`，再建立專用 fail-closed 探針。
- request code 34／35／71／72／73 映到 image 0／1／3／4／5，尺寸依序 192×96、
  296×32 與三張 56×70。五筆二階 RLE／4-bit EGA 像素數全部吻合，接觸表方向正常。
- 為避免把「router 未用」誤寫成「全程未用」，另列 `word_2FBB8` 全部 data refs；結果只有
  啟動、cleanup、唯一 consumer 三處，故 image 2 的一-byte payload 可證實為未引用
  `excluded_non_image` placeholder，而非猜測性 PNG。
- 初次專用摘要尚未帶 buffer 排他 refs；補入後以全新 DB 重跑，沒有沿用舊 `.i64`
  冒充驗證。新增安全 IDA 摘要、視覺統計與 ignored 五張 PNG／接觸表；SPEC-002 維持 DRAFT。
- 兩次最終全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  `e96ec3d7…0c3a`、`a6b77e8d…5c79`；兩次視覺摘要與接觸表也逐 byte 相同，SHA-256
  分別為 `67a76253…c456`、`d595b820…a972`。
- Python syntax、全部 JSON、本地 Markdown 連結、版控／重生摘要與 UID/GID 均通過；
  本輪五個 `/tmp/kol-*` 與重複 PNG 目錄已刪除。Docker 只餘已確認屬其他專案或服務的
  容器，未停止或刪除它們。

## 2026-08-31：COMBPICS 全七筆完成

- 盤點 30 個邏輯 archive 時發現 COMBPICS 有七筆，而 RE-006／SPEC-002 只關閉前五筆；
  保留舊限制並追查第二 consumer，沒有把「前五張成功」冒充整包完成。
- `sub_288D4` 五個 caller 完整覆蓋 0..4，固定 296×144；`sub_2B454` 兩個 caller 傳
  6／5，固定 88×64。兩組 caller 與 archive 0..6 一一覆蓋。
- 七筆二階 RLE 全部精確吻合矩形；七張 ignored PNG 接觸表為五張戰鬥場景與兩張肖像，
  無鏡像、旋轉、plane 錯排或裁切。0、1 繼續由魔旅士文章同場景截圖交叉驗證。
- 新增 fail-closed IDA 探針、安全摘要、全 archive 視覺 probe 與統計；COMBPICS 狀態由
  部分完成訂正為全 archive 已證實，SPEC-002 整體仍維持 DRAFT。
- 兩次全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  `765f2223…618c`、`a30919d8…e657`；兩次視覺摘要、七張 PNG 與接觸表均逐 byte 相同，
  視覺摘要與接觸表 SHA-256 分別為 `95c1ca7a…4ced`、`c1b2aaeb…d031`。
- Python syntax、全部 JSON、本地 Markdown 連結、版控／重生摘要與 UID/GID 均通過；
  兩個 `/tmp/kol-combpics-*` 與重複 PNG 目錄已刪除。Docker inspect 未見掛載 KOL
  工作樹的殘留容器，其他專案容器未動。

## 2026-08-31：TITLE SET0／SET1／SET2／FACE 共 24 張

- 以固定 SHA-256 的 `TITLE.EXE` 建立全新 IDA 9.4 DB，定位四個檔名字串、整檔 reader、
  索引 selector、little-endian word reader 與 mode 0Dh blit；早期探針漏掉 `>set0%s.0`
  的有效子字串起點，修正後每次均用全新 DB 重跑，未沿用舊 `.i64`。
- SET0／1／2 與 FACE 各完整覆蓋圖號 0..5；尺寸為 88×70、104×84、160×125、112×80。
  24 筆 LZW／二階 RLE 像素數全部吻合，索引不超過 15。
- 代表性 ignored PNG 目視為三組山谷騎士動畫與正向臉部影格，無鏡像、旋轉、plane 錯排
  或裁切。新增安全 IDA 摘要、視覺統計及可重生探針；SPEC-002 仍維持 DRAFT。
- 第二個最終全新 IDA DB 的安全摘要，以及第二批 24 張 PNG／視覺統計，均與第一批逐
  byte 相同。Python 語法、全部 JSON、Markdown 本地連結與 `git diff --check` 通過。

## 2026-08-31：TITLE BACKPICS 全 14 張

- 專用 IDA 探針列出 `sub_10E9F` 的全部 30 個 caller；僅 `_main` 與 `sub_10C86`，每處
  都是立即值圖號，集合完整覆蓋 0..13，沒有把部分 caller 冒充整包。
- 匯出並核對 TITLE 自己的 LZW、二階 RLE／四 EGA planes 與 mode 0Dh blit pipeline。
  14 筆全部精確得到 320×200、索引 0..15、無透明的全螢幕畫面。
- 代表性 ignored PNG 為山谷雷雨、閃電與製作名單，無鏡像、旋轉、plane 錯排或裁切。
  新增安全 IDA 摘要、視覺統計與可重生探針；SPEC-002 仍為 DRAFT。
- 第二個最終全新 IDA DB 的安全摘要，以及第二批 14 張 PNG／視覺統計，均與第一批逐
  byte 相同。

## 2026-08-31：KOLTOWN ENDPICS 全七張

- 一般字串探針未列 `end%s` consumer；原始 offset 唯一命中 `0x2b821` 後，回查 IDA DB
  證實位於 `sub_2B80A`。未把缺少自動 xref 誤判成未使用資料。
- `sub_2B80A` 全部七個 caller 均在 `sub_2B4A7`，傳入圖號 0..6；專用探針對 call set、
  圖號與 28 筆矩形 assignment 設 fail-closed 斷言。
- 七筆 LZW／二階 RLE 全部精確吻合各自矩形，索引不超過 15；七張 ignored PNG 為結局
  敘事人物／場景，無鏡像、旋轉、plane 錯排或裁切。
- 新增安全 IDA 摘要、視覺統計與可重生探針；ENDPICS 全 archive 升為已證實，
  SPEC-002 整體仍維持 DRAFT。
- 第二個最終全新 IDA DB 摘要，以及第二批七張 PNG／安全統計，均與第一批逐 byte 相同。

## 2026-08-31：KOLBOOK BOOKPICS 全 14 張

- `_main` 以 `book%s.0` 整檔載入 archive；raw offset 另一命中只是 `mov al,5`，已排除。
  `sub_1137B` 全部 caller 只傳 0..13 與 100，後者在函式入口映為圖號 9。
- 幾何分支為 0..3 的 320×200、4..12 的 112×31、13 的 248×22；探針對 alias、
  caller 常數集合及全部幾何 operand 設 fail-closed 斷言。
- 14 筆 LZW／二階 RLE 像素數全部吻合、索引不超過 15；ignored PNG 目視為書本介面、
  章節標題與圖示列，無鏡像、旋轉、plane 錯排或裁切。
- 新增安全 IDA 摘要、視覺統計與可重生探針；BOOKPICS 全 archive 升為已證實，
  SPEC-002 仍維持 DRAFT。
- 第二個最終全新 IDA DB 摘要，以及第二批 14 張 PNG／安全統計，均與第一批逐 byte 相同。

## 2026-08-31：WEATPICS 全 18 筆分層合成

- 由 `weat%s.0` 預載 buffer 追到 `sub_2A0C6`、`sub_2A01C`與 `sub_2A18E`，證實六組
  64×62 圖層依 `(base AND mask) OR overlay` 合成。
- group 0／1／2／5 的第三層為 496-byte MSB-first 1-bit mask；group 3／4 為四平面
  4-bit mask。遮罩 PNG 是運算資料，不是 alpha。
- 六張 composite 目視為晴空、雲景、風暴、雨、雪、夜景。早期僅依 caller 位置所作的
  「戰鬥效果」判斷已由圖像證據否定，勘誤已寫入 RE-019。
- 兩個全新 IDA 9.4 DB 的 safe／private JSON，以及兩輪 18 張圖層 PNG、6 張 composite
  與摘要均逐 byte 相同。safe、private、視覺摘要 SHA-256 分別為
  `11ef4d0e…bc35`、`638bce39…ae4a`、`7bb4e514…bed9`。
- 新增可重生 IDA 探針、PNG 探針、安全證據摘要與格式規格；SPEC-002 仍為 DRAFT，
  未建立 production importer 或 Go runtime。

## 2026-08-31：AWD1PICS 全 24 張勳章圖示

- `sub_167D9` 以 6×4 迴圈完整產生圖號 0..23，再對五個角色槽的 packed 2-bit
  狀態判斷是否畫出；`sub_168F8` 固定 24×21 矩形並走共用 EGA direct-blit 管線。
- 24 筆 payload 均精確展開為 504 個 4-bit indices；接觸表是正向勳章、獎盃、盾牌、
  戒指與紋章圖示，無鏡像、旋轉、plane 錯排或裁切。中文手冊與參考卡的 `Medal`
  命令提供獨立語意交叉證據。
- 兩個全新 IDA DB 的 safe／private JSON，以及兩輪 24 張 PNG、接觸表與摘要均逐 byte
  相同。safe、private、視覺摘要、接觸表 SHA-256 分別為 `5684459b…df8`、
  `445767bb…af2`、`8c9ed5fe…b5fe`、`b11f2e05…14d3`。
- 固定 ZIP 只附 bundled `AWD1PICS`；consumer 會建立 `AWD1..AWD5` 檔名，但未把未出現的
  其餘角色檔猜成相同資料。SPEC-002 仍為 DRAFT，未接 production importer。

## 2026-09-01：SHLDPICS.0 全四組盾牌

- 先將無副檔名 `SHLDPICS` 固定記錄表與 `SHLDPICS.0` 索引 archive 分流。四字元
  `shld` 因 IDA 預設最短字串長度而漏列；改以 raw DGROUP offset 找到真正 operand consumer。
- 第一版失敗即關閉斷言把 immediate 內 byte 誤當指令頭，因此停止且沒有產生證據。
  修正後同時保留 raw EA `0x1929f`與 instruction-head EA `0x1929e`，並排除兩個 far-call bytes 巧合。
- `sub_18F3A` 以 `[2*i,2*i+1]` 選擇彩圖／mask；`sub_16FA6` 載入後交給
  `sub_1BD17`。四組尺寸為 40×29、56×45、48×63、48×61，八個 payload 像素數全部吻合。
- RGBA 依原版 `(destination AND mask) OR color` 用 `alpha=1-mask`。接觸表是兩張圓盾與
  兩張徽章形盾，透明邊界連續，無方向、plane 或裁切錯誤。
- 兩個全新 IDA DB 的 safe／private JSON，以及兩輪四張 PNG、接觸表與摘要均逐 byte 相同。
  IDA safe／private SHA-256 為 `bfd58326…34fe`、`739147f4…6bcd`；路徑無關的最終視覺摘要／接觸表
  為 `13dd97ae…a6ef`、`079a90ad…0da9`。曾寫入暫存 basename 的舊摘要已作廢。
- 本切片曝露 MONPICS／PERPICS 早期探針直接以 mask bit 作 alpha，與已證實極性衝突。
  已在同輪改為 `opaque=1-mask`，並用 8×8 棋盤背景重生 32 張怪物與六張人物。
- MONPICS／PERPICS 兩輪修正版 PNG、接觸表與摘要均逐 byte 相同；摘要／接觸表 SHA-256
  分別為 `7a5a5258…3b29`／`934f1eb3…f2b5` 與 `d6d87f46…9fc5`／`b1a90a2f…9ff8`。
  舊 alpha PNG 與舊摘要已作廢；SPEC-002 仍為 DRAFT，未接 production importer。

## 2026-09-01：HANDPICS.0 全 12 組手部 sprite

- `sub_18F3A` 以 `2*renderer_class` 選擇同類別的兩個手部組，各組再以
  `[2*group,2*group+1]` 取彩圖與 AND-mask；六類×兩位置完整覆蓋 24 個 payload。
- DGROUP `0x2e492` 的 24-byte 幾何表證實尺寸分佈為六張 16×8、兩張 16×6、
  兩張 24×8與兩張 24×7。24 筆 LZW／二階 RLE／1-bit mask 長度全部吻合。
- PNG 依 `sub_1BD17` 使用 `alpha=1-mask`；6×2 棋盤接觸表是六種外觀的成對手部，
  透明輪廓連續，無旋轉、pair-order、plane 或裁切錯誤。文件保留「第一／第二位置」，
  不把畫面方位擅自升格為解剖學左／右手。
- 兩個全新 IDA DB 的 safe／private JSON，以及兩輪 12 張 PNG、接觸表與摘要均逐 byte 相同。
  SHA-256 分別為 `8067fa1f…d88b`、`ddd7637a…ce18`、`a180983f…1b03`、`6ed86cb0…75b2`。
- 新增可重生 IDA／PNG 探針與安全摘要；SPEC-002 維持 DRAFT，未接 production importer。

## 2026-09-01：WEAPPICS.0 全 28 組角色武器 sprite

- 由 `sub_18F3A` 同一角色 renderer 鏈追出裝備欄位 selector、六值基表
  `[0,10,14,28,38,42]`、family 5 variant 2／3 交換、28 組尺寸與六類位置表。
- 第一版 IDA 斷言把 DGROUP `0x17ea` 誤猜成 `weap`，實際為 `%s%s`；失敗輸出未採用。
  修正後固定 `0x0609` 的 `weap` token 與 active renderer suffix，並用全新 DB 重跑。
- 56 筆全部配成 28 組 4-bit EGA 彩圖／MSB-first AND-mask；PNG 採 `alpha=1-mask`。
  4×7 棋盤接觸表目視為弓、刃器、斧、長柄武器、錘與連枷，沒有方向、配對、plane 或裁切錯誤。
- 兩個獨立全新 IDA DB 的 safe／private SHA-256 分別為 `ea84a195…4831`、
  `fa282fe0…17f9`；兩輪 29 個 PNG 與摘要逐 byte 相同，視覺摘要／接觸表為
  `f3f48278…f9fb`／`9af11e5a…7346`。SPEC-002 仍為 DRAFT，沒有進入 production runtime。

## 2026-09-01：HEADPICS／TORPICS／LEGSPICS 全 162 組角色分層 sprite

- 由 `sub_18F3A` 的 component 2→1→0 迴圈固定 legs／tor／head token、存在欄位、
  `9*renderer_class+variant` selector、162 組寬高／位置與三 archive 配對。
- 初版把 `%s%s` 猜成 `%spics`，斷言如預期失敗且沒有安全摘要；修正後以兩個最終全新
  IDA DB 重跑，safe／private SHA-256 為 `87151ef1…1b1b`／`20a82ec0…c932`。
- 324 筆全部解成 162 組 4-bit EGA 彩圖／MSB-first AND-mask；三份 6×9 棋盤接觸表
  顯示協調的頭部、軀幹與腿部衣甲組件，透明片段、方向、配對、plane 與裁切均正常。
- 兩輪 165 個 PNG 與摘要逐 byte 相同；視覺摘要／三份接觸表 SHA-256 為
  `d01b16e1…f3ba`、`3a4bd8a0…92a3`、`3dc7994e…58bc`、`fc0c8382…d53e`。
  實際合法 variant 資料域仍待 writer 證據；SPEC-002 仍為 DRAFT，未接 production runtime。

## 2026-09-01：KOLPICS.0 全 14 筆混合 archive

- 先排除同名 `KOL.EXE`：它只是顯示模式偵測／launcher；真正 consumer 是固定雜湊的
  `KOLTOWN-unpacked.EXE`。IDA 路由證實 active indices 為 0／1／2／4／5／8／12／13，
  其餘六筆均為一-byte placeholder。
- 1／2／4／5／13 是 320×200、296×32、288×155、288×77、288×78 的 RLE EGA 圖。
  0／8／12 經 `sub_19B29` 以每 256-byte source stride 的前 128 bytes、四個 32-byte
  planes 轉成 256 個 chunky pixels；索引 8／12 下游以 `y<<6+x` 證實 64 欄。
- 初次目視以錯誤 256×80 排列出現交錯雜訊，已保留為勘誤並回到 consumer；最終索引
  8／12 採 64×160，寬度已證實、高度強推論；索引 0 的 256×5 仍是強推論。
- 兩個最終全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `af3fe195…9e2`／`25193b3d…58d`；兩輪八張 PNG、接觸表與摘要逐 byte 相同，摘要／接觸表
  為 `09650949…d42`／`d6f3a288…24e`。SPEC-002 仍為 DRAFT，未接 production importer。

## 2026-09-01：KOLMPICS.0 全 83 筆

- `sub_1103F` router 證實 KOLMPICS 候選 request 集合，並分流 KOLPICS／MISCPICS；整包
  最終分類為 71 筆 active 與 12 筆一-byte placeholder。
- 圖號 0／2 的 caller 矩形與 RLE 分別精確得到 320×200／296×32；68 筆圖示全部精確
  得到 480 pixels，與 `sub_11722`／`sub_23638` 的 24×20 helper 契約及全批目視吻合。
  動態 `sub_27E20` 索引的個別幾何仍標強推論，不由像素數單獨升格。
- 圖號 28 沿用已證實的 56×57 AND-mask／`alpha=1-mask`；其餘 68 圖示接觸表呈現連續的
  手勢、武器、狀態、人物與方向箭頭，未見 RLE、位平面、旋轉或裁切錯誤。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `79fceaaf…e5ba`／`c39152e8…71d2`；兩輪全 PNG 與摘要逐 byte 相同，視覺摘要／接觸表為
  `bd8a5858…4bfb`／`2f335771…d15a`。SPEC-002 仍為 DRAFT，未接 production importer。

## 2026-09-01：固定角色圖片表 40×38 格式

- 從 KOLBOOK 的 760-byte 記錄複製點追到原尺寸 `sub_15AEC` 與 4 倍顯示
  `sub_15B6E`；兩者都逐列讀 20 bytes、共 38 列，並以 high nibble、low nibble 順序
  取得 40 個 4-bit EGA pixels。
- 新增可重生 IDA consumer 探針；兩個全新 DB 的 safe／private JSON 逐 byte 相同，
  SHA-256 為 `db71511e…b316`／`8bcbfc6e…815a`。
- 新增固定記錄 PNG 探針，解出 `CHARPICS`、`SHLDPICS`、`STEEDPIC` 各 16 張、共 48 張
  ignored PNG。接觸表目視支持方向與 nibble 次序；空白及雜訊槽不推定語意。
- 安全視覺摘要／接觸表 SHA-256 為 `7eede045…bb55`／`fd552865…37d1`。已更新
  RE-027、圖片格式、SPEC-002、CONTEXT 與 WORKLIST；SPEC-002 維持 DRAFT，未進入實作。
- 首次重生比對因容器輸出路徑未含獨立 `workplace` 目錄而被探針拒絕；該驗證 shell 又未
  啟用 `set -e`，因此留下連續 `cmp` 路徑錯誤。修正為新 `workplace` 路徑並用 `set -eu`
  以同一 image 重跑後，48 張 PNG、接觸表、JSON 與 Python 語法全部通過。
- 本輪 IDA、解碼與驗證容器均使用 `--rm`；本輪 `/tmp` 資料庫、重跑輸出與 Python cache
  已精確刪除。結束檢查未見 KOL 專案容器；其他專案及系統既有容器保持不動。

## 2026-09-01：CHARPICS.0 全 80 張角色模板

- 沿 KOLBOOK `sub_11820` 追出 `CHARPICS.0` → `sub_13D7D` → 目的 buffer →
  `sub_15335` 的完整鏈；索引公式為 `20+2*image_index`。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `b4877736…ff30`／`18d44d22…f74b`。
- 80 筆全部解成 761 bytes；前 760 bytes 依 RE-027 轉成 40×38 EGA PNG，末端共同
  `0x00` 排除於像素。10×8 接觸表顯示完整人物、武器、盾牌、服裝與坐騎組合，逐圖語意
  未由外觀猜定。
- 兩輪 80 張 PNG、接觸表與摘要逐 byte 相同；摘要／接觸表 SHA-256 為
  `0bcb3709…7575`／`cd0a65db…a5a4`。SPEC-002 維持 DRAFT，未接 production runtime。
- 本輪 IDA、抽取與驗證全部使用 `docker run --rm`；`/tmp` 資料庫與重跑輸出已精確移除。
  結束檢查沒有 KOL 專案容器，其他專案與系統容器未動。

## 2026-09-01：TWN0..7PICS 全 243 筆

- 由 KOLTOWN `sub_11302`、`sub_14047`、`sub_15920` 與 RE-008 分卷 loader 建立完整
  TWN 幾何／遮罩契約；兩個全新 IDA DB 的 safe／private 輸出逐 byte 相同，SHA-256
  為 `123cf073…ca52`／`76b2f51d…f360`。
- TWN0 的 25 筆全部解成 192×96；TWN1..7 的圖號 0 為 56×57 彩圖加
  KOLMPICS 圖號 28 AND-mask，圖號 1 為一-byte placeholder，後續偶數／奇數圖號分別
  為 88×64／192×96。合計輸出 236 張 PNG，排除七筆 placeholder。
- 兩輪 236 張 PNG、八張接觸表與摘要逐 byte 相同；視覺摘要 SHA-256 為
  `13a3b6d8…dce3`。TWN0、TWN1、TWN7 接觸表人工抽查通過；不由外觀猜定 NPC 或地點語意。
- 新增 RE-029、IDA／PNG 可重生探針與安全摘要，並更新圖片格式、SPEC-002、CONTEXT、
  WORKLIST 與 RE 索引。SPEC-002 仍為 DRAFT，沒有建立 Go runtime 或 production importer。
- 本輪 IDA、抽取與驗證全部使用 `docker run --rm`；第二輪 PNG、Python cache、`/tmp`
  IDA 資料庫與摘要比對檔已精確移除。保留第一輪 ignored PNG 作本機視覺證據；結束檢查
  沒有 KOL／TWN 專案容器，其他專案及系統容器未動。

## 2026-09-01：SPEC-002 完整 EGA extraction set

- 以 `image-asset-inventory.json` 的 30 個邏輯 archive／1,022 payload 作唯一母體，逐包
  對照現有 IDA 與視覺摘要；新增 `002-ega-extraction-set.md`，不再用「其他 archive」
  這類無法驗收的籠統缺口。
- 新增失敗即關閉 verifier，核對 archive 集合、逐包 payload、disposition 完整性與總和；
  容器驗證結果為 925 筆已證實、71 筆待證實、26 筆非圖片、681 個已證實 PNG 單位。
- 71 筆缺口精確縮為 KOLMPICS 68 筆動態圖示與 KOLPICS 圖號 0／8／12 raw-planar 幾何。
  研究探針雖已有 PNG，仍依 SPEC-002 規則保持 `excluded_unknown`，沒有升格 production。
- 已以全新 IDA DB 重生 KOLMPICS／KOLPICS private 證據，確認下一窄切片應追
  `sub_1103F`／`sub_11722`／`sub_23638`／`sub_27E20` 與 `sub_19B29`；本輪尚未把新的
  分析推論寫成已證實契約。
- 本輪 verifier 與 IDA 均在 `docker run --rm` 執行；兩個暫存 IDA DB 目錄已精確移除，
  結束檢查沒有 KOL／KOLPICS／KOLMPICS 專案容器。

## 2026-09-01：KOLPICS raw-planar renderer-data PNG

- 重新審查 `sub_19B29` 的 256-byte source stride／128-byte planar input／256-index output
  契約，並將 original display geometry 與 production extraction geometry 明確分開。
- 圖號 0 caller endpoint `0x500` 證實五個 converter block，固定為 256×5 data PNG；
  圖號 8／12 各有 40 個完整 stride，配合下游 `y<<6+x` 證實 64×160 source grid。
  三者都不是 direct-blit 螢幕矩形。
- IDA 探針與視覺探針升 v2。兩個全新 DB 的 safe／private JSON 逐 byte 相同，SHA-256
  為 `a8ee8476…e63d`／`c6538475…45a5`；兩輪八張 PNG、接觸表與摘要亦逐 byte 相同，
  摘要／接觸表為 `82281d7a…8f67`／`d6f3a288…524e`。
- 追加 RE-030 勘誤，不覆寫 RE-025 的歷史形成原因。extraction set 更新為 928 筆已證實、
  68 筆待證實、26 筆非圖片與 684 個 PNG 單位；KOLPICS 不再有 `excluded_unknown`。
- 本輪 IDA、抽取與驗證全部使用 `docker run --rm`；第二輪 PNG、Python cache、兩個 IDA DB
  與暫存摘要已精確移除，保留第一輪 ignored PNG 作視覺證據。結束檢查沒有 KOL／KOLPICS
  專案容器。

## 2026-09-01：KOLMPICS 68 筆 canonical 24×20 extraction shape

- 將 lossless extraction shape 與 caller-specific display geometry 分流。`sub_11722` 與
  `sub_23638` 提供兩條固定 24×20 consumer；68 筆圖示各精確展開為 480 indices，完整
  接觸表在 24×20 下維持連續 UI 圖示，三層證據共同關閉 extraction geometry。
- `sub_27E20` 的動態螢幕位置與 clipping 仍明列為未宣稱，不由 PNG shape 冒充 caller-exact
  parity。IDA／視覺探針升 v2，保留這項限制於 safe summary 與逐筆紀錄。
- 兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `93a4a3be…f6ef`／`cf44a187…a160`；兩輪 73 個 PNG 與摘要逐 byte 相同，摘要／接觸表為
  `2647e6a6…572e`／`2f335771…15a`。
- EGA extraction set 更新為 996 筆已證實、0 筆待證實、26 筆非圖片與 752 個 PNG 單位。
  SPEC-002 的 EGA 格式 gate 已通過，但 theme 決策與 production extractor 尚未完成。
- 本輪 IDA、抽取與驗證全部使用 `docker run --rm`；第二輪 PNG、Python cache、兩個 IDA DB
  與暫存摘要已精確移除，保留第一輪 ignored PNG 作視覺證據。結束檢查沒有 KOL／KOLMPICS
  專案容器。

## 2026-09-01：音訊 RE 知識庫與 SPEC-004 DRAFT

- 固定 ZIP 未見獨立 WAV／VOC／MIDI／音訊 archive；建立機器可讀 `audio-inventory.json`。
- IDA 證實 TITLE 的 speaker enable／disable、PIT channel 2 divisor 與兩段隨機音爆控制流；
  依硬體時序停止線，不深挖 busy-wait 的逐週期 wall-clock。
- IDA 證實 `TANDYPIN.EXE` 使用 DOS `INT 21h AH=31h` 常駐記憶體，不是音訊播放器。
- TITLE 與 TANDYPIN 各以兩個全新 IDA DB 重生，safe JSON 逐 byte 相同；SPEC-004 維持
  DRAFT，尚未建立 OGG 或 Ebitengine 正式音訊實作。安全摘要 SHA-256 分別為
  `7bd5e0bc…b526` 與 `c5d62e57…9eb1`。
- 全部分析與驗證皆使用 `docker run --rm --network none`；本輪十個精確暫存目錄已刪除，
  結束檢查沒有 KOL／TITLE／TANDY 容器，其他專案容器未動。

## 2026-09-01：六 executable 音訊母體與 TITLE WAV oracle

- 新增通用 IDA 音訊 I/O 探針；六個 executable 各以兩個全新 DB 重生，safe JSON 逐 byte
  相同。六份 safe SHA-256 依 KOL／KOLBOOK／KOLTOWN／MAKEMODE／TANDYPIN／TITLE 為
  `a8186db9…ce76`／`77769513…eb97`／`809ca2ae…c5ba`／`df479c0c…4d9e`／
  `72449d60…75e7`／`f0858b80…c25f`。只有 TITLE 命中 `0x61`／`0x42`，其餘五個無音訊候選；
  六個未知 DX 點人工排除為 EGA。
- TITLE 八個 cue caller 已固定在正常 presentation flow，安全摘要補入 caller 函式與限制。
- 初次沿用傳統 `Ctrl+F6` 沒有產生 WAV；查官方文件後確認新版 Linux host key 為 F12，且
  官方另提供 `DX-CAPTURE`。改用 `DX-CAPTURE /A /-D TITLE.EXE` 後，同一容器取得有效 WAV。
- WAV 為 48 kHz stereo PCM、38.229333 秒，含四段非靜音窗；只留本機 oracle，Git 保存
  metadata、SHA-256 與證據限制。SPEC-004 仍為 DRAFT，未建立 OGG production 資產。
- IDA、DOSBox-X、Xvfb、FFmpeg 與驗證皆用有界 `docker run --rm`；十二個 IDA 暫存目錄與
  DOSBox-X 暫存目錄已精確刪除，結束檢查沒有 KOL／TITLE／TANDY／DOSBox 容器。

## 2026-09-01：角色建立 RE 與 SPEC-005 DRAFT

- 從已校對手冊建立角色建立狀態機、16 名儲存上限、六人隊伍上限與 35 個職業表的證據入口；
  明列職業數值是平均值，不猜作 RNG range。
- 固定 `CHARDATA` 16×768 與 `CHARPICS` 16×760 stride；新增 IDA 探針定位 KOLBOOK
  `chardata` 字串、`sub_134C2` 路由及五個立即 arg=0 caller。
- 初版探針誤把 little-endian immediate 的位址當 instruction 起點；依原始 bytes
  `b9 1a 02` 訂正為 EA `0x134f7`，保留形成原因後以兩個全新 DB 重跑。safe JSON 逐 byte
  相同，SHA-256 `e71a6ca3…ec5d`。
- SPEC-005 維持 DRAFT；沒有建立角色 production code 或臆測 768-byte 欄位。
- 全部探針與驗證使用 `docker run --rm --network none`；八個精確暫存 IDA 目錄已刪除，
  結束檢查沒有 KOL／KOLBOOK／chardata 容器，其他專案容器未動。

## 2026-09-01：角色主要屬性 RNG 與衍生公式

- 新增 KOLBOOK RNG 探針，窮舉 `_srand`、`_rand`、唯一 bounded-roll wrapper 及二階 caller；
  固定 BIOS tick seed、模式值 7 的 seed 16 分支、八值骰面表與七項主要屬性擲骰公式。
- 回查 `sub_11820` 證實同一職業編碼 byte 同時作起始數值、骰面 selector 與擲骰次數；
  排除「另有 base table／LZW 後處理」的早期假說，並固定三個向下取整的衍生平均公式。
- 使用既有 `ida-pro-9.4-ver2:civ1-py312-uidfix-v2`，兩個全新 DB 的 safe／private JSON
  逐 byte 相同，SHA-256 為 `f8ff1103…187b`／`9575adcf…d7f`。安全摘要已版控，完整
  private 輸出與 `.i64` 不進版控，並已在逐 byte 比對後隨本輪精確暫存目錄清除。
- 更新 RE-035、角色證據表與 SPEC-005 DRAFT；沒有建立 Go／Ebitengine production code。
- 全部分析與驗證使用有界 `docker run --rm --network none`；五個 `/tmp/kol-rng-*` 目錄已
  精確清除，結束檢查沒有 KOL／IDA 專案容器。

## 2026-09-01：職業記錄與屬性量尺勘誤

- 回查 `sub_11820` 發現前輪漏讀 `0x11cb6 add bx,8`，因此「骰式 byte 同時是基值」結論
  作廢；正確記錄 layout 為七個骰式、零、七個內部基值、零。RE-035 保留形成原因，
  RE-036 追加訂正。
- `sub_12677` 證實玩家畫面使用 `floor(internal×100/150)`；新增 archive 探針，固定
  `BOOKDATA` 項目 2 為 39×16＋1 bytes，並以項目 1 的 CR-delimited 文字關閉 race、
  secondary 與 33 個 menu names 的記錄順序。
- 39 筆 selector 分組為 12／4／6／6／8／3；男女精靈共名但分用記錄，四對矮人記錄逐
  byte 相同，解釋手冊只有 35 列。沒有把 raw records 寫入 production typed data。
- 最終雙新建 IDA safe／private SHA-256 為 `e3309787…1116`／`a6e07f2a…8401`；雙重生
  archive safe／private 為 `aaa112ef…cb2d`／`7d13f6ea…9b36`。
- IDA、archive 解碼與驗證全用有界 `docker run --rm --network none`；十個
  `/tmp/kol-prof-*` 目錄已精確清除，結束檢查沒有 KOLBOOK／IDA 專案容器。

## 2026-09-01：角色 writer 與起始武器資料

- 以 KOLBOOK `sub_11820` 建立最小非破壞性 CHARDATA writer offset 索引，並固定
  CHARPICS／SHLDPICS 同槽初始化及三檔保存順序；未把未知 region 欄位擅自命名。
- 解出 39 筆金冠、主要／次要武器 selector 與兩組武器技能，另固定 14 筆武器模板和
  760-byte 初始盾牌模板。機械 verifier 對中文手冊 35 列展開為 39 記錄後全部通過。
- 兩個全新 IDA DB 的 safe／private SHA-256 為 `39beef72…b178`／`099e0d2b…9f73`；
  archive safe／private 為 `ae711e3b…2176`／`f42fbbf1…97bc`。
- SPEC-005 仍為 DRAFT：只關閉起始武器垂直鏈，不冒稱全部技能／物品或存檔 schema 已知，
  也沒有建立 Go／Ebitengine production code。
- IDA、archive 解碼與驗證皆使用有界 `docker run --rm --network none`；本輪精確清除
  `/tmp/kol-writer-*` 與 Python cache，專案容器檢查未見遺留，其他專案容器未動。

## 2026-09-01：角色 writer coverage 與 KOLTOWN reader

- 完整列舉 KOLBOOK `sub_11820` 在 768-byte slot 清零後的直接寫入，補入 `0x28`、
  `0x29`、`0x52`、`0x7f`，並訂正 `0x4a..0x51` 為武器模板 prefix。
- 證實 writer 沒有其他職業衍生起始技能／物品；REGION route 仍只宣稱 byte routing。
- 新增 KOLTOWN 最小 reader 探針，交叉確認名字、race、profession、gold 與主要武器。
- KOLBOOK safe／private 雙重重建 SHA-256 為 `2884a1ae…c129`／`6c00db3a…fdb6`；
  KOLTOWN 為 `16f8b24d…67ba`／`dcfa8c59…18e1`，四組摘要皆逐 byte 一致。
- SPEC-005 仍為 DRAFT，沒有把探索程式接入 Go runtime；下一步轉向正常 DOS UI oracle。
- 全部 IDA 與驗證均使用有界 `docker run --rm --network none`；探索腳本、六個全新 DB
  目錄、摘要副本與 Python cache 已精確清除。容器清冊未見 KOL／IDA 遺留，其他專案容器未動。

## 2026-09-01：角色建立 UI 靜態邊界與正常入口訂正

- 新增失敗即關閉的 KOLBOOK UI 探針，固定名字 16-byte 容量、空名、各層取消、屬性
  reroll、80 張肖像循環與滿 16 槽時在 writer 前退出。
- 兩個全新 IDA Pro 9.4 DB 的 safe／private JSON 逐 byte 相同，SHA-256 為
  `9ecd4b76…aa82`／`ef84a41e…50c`；安全摘要入版控，完整函式只留本機暫存。
- DOSBox-X 證實直接 `KOLBOOK.EXE A` 不是正常玩家入口；改由 `KOL.EXE` 成功進 EGA 開場。
  尚未完成主選單後的角色建立截圖，因此 SPEC-005 保持 DRAFT，沒有建立 production code。
- IDA、DOSBox-X、Xvfb 與格式驗證均在有界 `docker run --rm --network none` 執行；雙 DB、
  writable 原版副本、截圖與一次性 capture script 已精確刪除，專案容器清冊無遺留。

## 2026-09-01：正常 KOLBOOK 路徑初步實驗與空名勘誤

- 回查 KOL launcher `_main`：它在執行 TITLE 前檢查 `_kbhit`；已有緩衝鍵時執行
  `kolbook.exe`，否則執行 `title.exe`。最初用 stdin 重導向提供鍵值，卻讓子程式繼承
  重導向而無法接收後續鍵盤；改用 DOS `PAUSE` 加兩個真實鍵後，成功由 `KOL.EXE e`
  正常進入 KOLBOOK，這是 oracle 工具修正，不是遊戲規則變更。
- 暫存 DOSBox-X 畫面曾依序走到 Contents → Character options → New character，並觀察：
  名字層 ESC 回角色選項；種族 ESC 回名字；人類性別 ESC 回種族；職業 ESC 回性別；
  預設 REROLL STATS 會留在屬性頁且數值改變。各次實驗的角色三檔前後雜湊沒有差異。
- 直接在空名字欄按 Enter 會前進至種族層，推翻上一輪把「buffer 首 byte 為零時重試」解成
  「玩家空名不前進」的結論。另輸入 `ABCDEFGHIJKLMNOPQ` 時只顯示／接受 A 至 O；第一次
  Backspace 即刪除 O，初步支持此 DOS build 的可見 ASCII 上限為 15，而非手冊所寫 16。
- 使用者隨後要求暫停；依主機衛生規則，暫存截圖、可寫原版副本、IDA DB 與 capture script
  均已清除，沒有留下可版控 receipt。因此上述動態結果只列為「待重跑初步觀察」，不可
  升格成正式 oracle；靜態 bytes 契約已立即訂正，SPEC-005 維持 DRAFT。

## 2026-09-01：README 已確認畫面

- README 新增魔旅士文章的兩張外部原版截圖，分別對應已交叉驗證的 COMBPICS 圖號 0／1；
  圖片不複製進 Git，點擊會回到來源文章。
- 沿用 `docs/re/combpics-visual-oracle.json` 的證據限制：只確認場景身分、幾何與方向，不把
  JPEG 配準結果宣稱為逐像素 RGB parity；README 同時明列圖片權利仍屬原權利人。

## 2026-09-01：private GitHub PNG 特例決策

- 使用者明確決定本專案可將原版延伸 PNG 放入目前 private GitHub，取代「所有衍生 PNG
  一律不得進 Git」的舊政策。
- 特例只涵蓋從固定輸入可重生、已通過格式／視覺驗證且有 provenance／權利 manifest 的
  PNG，固定放在 `assets/original-derived-png/`。原版 ZIP、EXE、archive、PDF、音訊、
  反組譯 payload 與未驗證圖片仍禁止提交。
- private 保存不等於公開散布權；repo 改為 public、public fork、GitHub Release 與公開
  `dist-all/.../release/` 仍維持失敗即關閉並要求重新權利審查。

## 2026-09-01：圖像 theme 與 TITLE 音訊決策

- 使用者決定首版重製只涵蓋 EGA 素材，CGA／Hercules（`MAKEMODE.EXE` 產生的
  `PCGA`／`HERC` 分卷）列為後續選項，其完整轉換像素規則保留為待辦 RE。
- 使用者決定 TITLE cue 的 OGG 呈現採規格近似合成：依已證實的兩段 PIT 除數序列與
  公開硬體規格自行合成方波再轉 OGG，不散布原版錄音。DX-CAPTURE 取得的 WAV 仍只作
  本機比對 oracle。

## 2026-09-01：正常 DOS UI 角色建立 oracle 擷取器與四情境收據

- 建立三個版控工具：`tools/dosbox_character_ui_oracle.py`（唯讀 ZIP → 容器 tmpfs
  重建遊戲目錄 → Xvfb ＋ DOSBox-X → 畫格去重擷取 → 三檔前後 SHA-256 收據）、
  `tools/x11_key_inject.py`（純標準函式庫的最小 X11 XTEST 客戶端）、
  `tools/compare_character_ui_runs.py`（兩輪檢查點畫面比對）。
- 先用 DOSBox-X `AUTOTYPE` 的固定時序盲打，同一顆按鍵在不同執行的反應延遲介於
  1.5 秒到 11 秒，無法對齊選單層；容器映像沒有 xdotool／ImageMagick，因此改為自寫
  X11 XTEST 注入加「等畫面靜止再送下一鍵」。改成畫面驅動後四個情境全部穩定完成。
- **勘誤（工具，非遊戲規則）**：X11 客戶端最初把 server 主動送出的 32-byte event 封包
  當成 reply 解析，取到任意長度後永久 block 在 `recv`，表面症狀是「畫面凍結」。
  實際上凍結的是取樣端，模擬器仍在跑。修正為逐包檢查封包型別並加上 socket timeout。
- **勘誤（遊戲行為）**：上一輪暫存實驗記下的「空白名字可前進」與「可見 ASCII 上限 15」
  兩點皆不成立。實測名字欄容量為 16 個可見字元，輸入第 16 個字元即直接前進種族層、
  不需 Return；空名按 Return 停在名字層。當時看到的「15 個字元」是自動前進前的最後
  一格畫面，「Backspace 刪除 O」則是按鍵落在已切換的種族層而無作用。正文已改寫為
  現況，`docs/re/character-creation.md` 與 SPEC-005 同步訂正。
- 四個情境（`name-length`／`blank-name`／`escape-layers`／`reroll`）各執行兩輪全新流程，
  收據存入 `docs/verification/character-ui/`。跨輪比對：`name-length` 26/26 檢查點
  逐 byte 相同；其餘差異只出現在 TITLE 循環動畫落點（13.98%）、名字欄閃爍游標
  （0.0125%）與 BIOS tick 播種的屬性數值（0.175%–0.280%）。
- 四個情境全程未接受屬性或圖像，`CHARDATA`／`CHARPICS`／`SHLDPICS` 前後 SHA-256
  均未變動。畫面 PNG 只留本機；原版執行畫面截圖與 `assets/original-derived-png/`
  是不同素材類別，2026-09-01 的 PNG 特例未涵蓋，入庫前要另行權利決策。
- 全部執行使用有界 `docker run --rm --network none --memory 1g --cpus 1
  --pids-limit 256`，原版 ZIP 唯讀掛載，遊戲副本建在容器 tmpfs。中止的三個探索容器
  已逐一 `docker stop`，結束檢查沒有 DOSBox 容器殘留，其他專案容器未動。

## 2026-09-01：執行畫面 PNG 入庫與 git 作者欄統一

- 使用者決定原版執行畫面截圖 PNG 亦可進目前的 private 儲存庫。852 個去重畫格連同
  `receipt.json`、跨輪比對與 `MANIFEST.json` 移入
  `docs/verification/character-ui/<情境>-run{1,2}/`；收據以相對路徑索引畫格，
  因此畫格與收據同目錄，`AGENTS.md` 的權利邊界章節同步記錄這個位置分工。
- 權利分類維持 private-repo 可保存、不得公開散布；轉 public、public fork、
  GitHub Release 或公開 `dist-all/.../release/` 仍須重新權利審查。
- 使用者決定本 repo 的 git 作者欄一律為 `wicanr2@gmail.com`。已設定 repo-local
  `user.email`（不動全域），本輪兩個新提交 `d114fbb`、`eb357cd` 已使用該位址並推送。
- 既有 64 個提交仍掛 `cy.wang@coretronic.com`。改寫歷史需要 `git filter-branch`／
  `git filter-repo` 加上對 `main` 的 force push，這些動作被本工作階段的自動核准
  分類器擋下，尚未執行。待補：由使用者授權或自行執行後，再回填改寫結果與新的
  HEAD 雜湊。

## 2026-09-01：角色建立 oracle 補齊、EGA production 抽取器與資產 parity

- **勘誤（工具）**：XTEST 注入器對大寫字母沒有真的按 Shift。同一個 keycode 的 keysym
  清單同時含大小寫，直接查大寫 keysym 會拿到同一顆鍵而漏掉 Shift，因此先前情境標籤
  寫 `ABCDEF…`、實際送出的是小寫。已修正 Shift 處理，並把情境改為明寫小寫字母；
  全部六個情境各兩輪重跑，收據以新版取代舊版。本批不對大小寫保存行為下結論。
- 新增 `accept-and-save` 與 `picture-cycle` 兩個情境，補完 SPEC-005 缺口：圖像
  selector 恰好循環 80 張（由 0 向左一次 = 由 0 向右 80 次，再一次回到 0）；接受圖像
  後才寫檔，兩輪都只寫 `CHARDATA` 的 slot 14，`CHARPICS`／`SHLDPICS` 未變。
- 六情境十二輪跨輪比對：`blank-name` 10/10、`name-length` 25/26、`escape-layers`
  21/28、`reroll` 12/20、`accept-and-save` 12/16、`picture-cycle` 10/99 檢查點逐 byte
  相同；差異全部落在 TITLE 循環動畫落點、名字欄游標與 BIOS tick 播種的屬性數值。
- 新增 `tools/extract_ega_assets.py`：由固定 ZIP 產出 30 個 archive、752 張 PNG 到
  `assets/original-derived-png/ega/`。抽取計畫取自已版控的 visual probe 逐張紀錄，
  每張 PNG 重讀驗證並與紀錄的 SHA-256 核對，暫存目錄整批替換，輸出根目錄受限。
- 新增對拍工具 `tools/png_io.py`、`tools/compare_frames.py`、
  `tools/locate_asset_in_frame.py`、`tools/verify_charpics_against_capture.py`、
  `tools/diff_character_files.py`。
- **原版同狀態資產 parity**：80 張 CHARPICS 模板在搜出的 blit 座標 (208,106) 上，
  兩輪擷取各 80/80 逐像素相同；BOOKPICS[0] 對書本封面畫面 64,000 像素中只有 50 個
  不同，全部落在螢幕中央 8×11 的滑鼠游標。
- SPEC-002（含 EGA extraction set）與 SPEC-005 經證據審查升 READY；theme 決策已納入
  gate。SPEC-002 的 26 筆非圖片 placeholder 維持明列排除。
- 全部執行使用有界 `docker run --rm --network none`（DOSBox 擷取 `--memory 1g --cpus 1`，
  抽取與比對 `--memory 2g --cpus 1`）。原版 ZIP 唯讀掛載，遊戲副本建在容器 tmpfs；
  執行後的角色檔只留本機比對，不進版控。結束檢查沒有 DOSBox／KOL 容器殘留。

## 2026-09-01：Go／Ebitengine 垂直切片與首批 remake 對拍

- 鎖定建置環境：既有映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`
  （image ID `sha256:e651beea…64eb`，Go 1.26.7、Ebitengine v2.9.9），本專案只讀不改。
  映像內的 module cache 屬 root，因此包裝腳本每次先把 55 MB cache 複製到容器 tmpfs
  再指向副本，避免以 root 執行而在專案留下 root-owned 檔案。
  `tools/go_docker.sh` 跑 go 指令，`tools/run_kol_headless.sh` 另起 Xvfb 執行 remake。
  映像沒有 `xauth`，`xvfb-run` 不能用，改為直接起 `Xvfb :91` 並在結束後收掉。
- `cmd/kol` 垂直切片：`internal/assets` 只讀 `assets/original-derived-png/ega/` 的
  manifest 與 PNG（拒絕非 PNG、拒絕目錄外路徑、檢查尺寸與 manifest 相符）；
  `internal/screen` 提供固定 320×200 離屏畫布並可寫出 PNG；`internal/book` 只記錄
  已由原版畫面證實的組成事實。截圖模式一次輸出書本封面、目錄頁、空白跨頁與 80 張
  New Character 畫面。
- 由 BOOKPICS 與原版畫面的差異外接矩形量出畫面分解：目錄頁整頁文字都烘在
  BOOKPICS[1] 素材裡（與原版畫面只差 55 px 的滑鼠游標）；New Character 頁面以
  BOOKPICS[2] 為底圖，原版自繪文字落在左頁 (24,16,146×183) 與右頁。
- 新增 `tools/verify_remake_parity.py`：三層判準（素材與版面必須相同、明列允許差異區
  並記錄實際差異量、必須相同區優先於允許差異區）。首批計畫 82 對全部通過：封面與
  目錄頁在原版游標外逐像素相同；80 張 New Character 畫面的角色圖像 blit 矩形
  (208,106,40×38) 全部逐像素相同。remake 截圖與計畫、收據一併入庫，收據可重生。
- 全部建置與執行使用有界 `docker run --rm --network none`（Go `--memory 3g --cpus 2`，
  比對 `--memory 2g --cpus 1`）。中止的一個 Go 容器已 `docker stop`，結束檢查沒有
  eob／DOSBox 容器殘留，其他專案容器未動。

## 2026-09-01：原版字型抽取與含文字畫面的逐像素對拍

- 從原版畫面反推字模再回頭搜尋原始檔：`RACE : HUMAN` 的五個字模逐 byte 取出後，
  在固定 ZIP 的所有檔案與所有 archive payload 中搜尋，命中 `BOOKDATA.0` 與
  `KOLDATA.0` 的第一個 payload（同一份 1,303-byte 區塊）。直接在 EXE 裡掃 1bpp
  圖樣沒有結果，因為字型在壓縮 payload 裡。
- 該區塊含兩套字型：8×8 與 8×6，各 93 個字模、ASCII 32–124。
  `tools/extract_fonts.py` 產出 RGBA 字模表與含逐字模 hex 的 manifest。
- 新增 `tools/locate_text_in_frame.py` 與 `tools/extract_text_layout.py`：用抽出的
  字型在原版畫面上機械定位字串，要求唯一命中，輸出 `docs/re/text-layout.json`。
  初版只比對墨點同色，會在單色背景整片誤命中；加上「文字框內非墨點須大多不同色」
  的對比條件後每個字串都唯一命中。
- Go 端新增 `internal/font`（讀字模表、以 ColorScale 上色）與 `internal/book/layout.go`
  （版面表，座標與顏色全部引用機械量測結果）。
- 逐像素比對逼出兩條原版繪字規則：每個字模格先清成白底再畫墨色，但空白字元不清底；
  最後一列的實際文字是 `BODY PTS.`。修正後空白 New Character 頁與原版只差 31 個像素。
- 那 31 個像素全在名字欄：原版在種族層取消後於該處殘留一個字元，字模逐 byte 命中為
  前次名字的首字 `t`。成因未確認，列為原版狀態而非版面問題。
- 對拍計畫擴為 83 對，全部通過。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：角色建立狀態機與五層逐畫面對拍

- 新增 `tools/extract_character_data.py`，把 39 筆職業記錄、骰面表、種族／secondary
  名稱與 100/150 顯示量尺抽成 `assets/original-derived-data/character-creation.json`；
  規則層只讀這份 typed data，執行期不解析 `BOOKDATA`。
- Go 端新增 `internal/chardata`（typed data loader）、`internal/charcreate`（狀態機：
  名字 16 格滿格自動前進、各層取消目的地、擲骰與衍生公式、80 張肖像循環）與
  `internal/ui`（把狀態畫成像素）。`cmd/kol` 加 `-script`／`-seed`，可用固定腳本逐步
  截圖重播，對拍才可重生。
- 逐像素比對逼出的原版規則，依證據補進版面表：
  - 「New character」標題是 `BOOKPICS[10]`，位置 (32,16)。
  - 冒號不緊接標籤，而在固定字元欄（名字／種族／性別／職業列 x=64，屬性列 x=104）。
  - 名字輸入游標是 8×1 的黑線，畫在下一個字模格的底列（空名字時 x=24、y=63）。
  - 標籤高亮：前四層各自只有一列變棕；**擲屬性層是十列屬性同時變棕**；圖像層沒有
    作用中的列。以整輪畫面的逐列棕／青像素簽章證實，`picture-cycle-run1` 與
    `reroll-run1` 一致。
- 建立 `plan-newchar-steps.json`：9 對逐畫面對拍，9 通過。允許差異區以外零像素差。
- 殘餘差異查證到底，四類全部有機械證據（RE-045）。其中「頁尾殘影」原先看起來像素材
  解碼錯誤，實際是原版行為：貼展開跨頁時左頁底緣有一條未覆蓋的細帶，帶內 112 個像素
  **每一個都等於 `BOOKPICS[0]`**，而且差異是雙向的（有的位置原版紅素材黑，有的相反），
  解碼錯誤不會這樣。另外驗證 `BOOKPICS[2]` 解出 64000 像素且 `pixel_sha256` 相符，
  以及「索引 0 當透明」不成立（2857 個黑像素只有 100 個透出前一畫面）。
- 全部建置與比對使用有界 `docker run --rm --network none`（Go `--memory 3g --cpus 2`，
  比對 `--memory 1–2g --cpus 1–2`）；結束檢查沒有容器殘留，其他專案資源未動。

## 2026-09-01：圖像層版面與完整流程 92 對逐畫面對拍

- 把角色建立的圖像層畫完：右頁五行說明（`USE ARROW KEYS` …）、底部的
  `CANCEL   ACCEPT`、青色雙格外框、名牌白字與右側上下箭頭。字串全部用
  `tools/read_text_from_frame.py` 以抽出的原版字型逐格讀回，座標由原版畫面量得。
- 上下箭頭是 8×8 大字型的 codepoint 94／95，字模逐 byte 命中。`extract_fonts.py` 的
  大字型註記與 manifest 同步改寫，重跑抽取器產出的 manifest 與手改後逐 byte 相同。
- 新增 `plan-newchar-pictures.json`：`z e d` 建到圖像層，先向左回繞一次，再向右走完
  80 張並回到起點，共 92 對；92 對全通過，允許差異區之外零像素差。80 張圖像的
  blit 矩形另列為必須相同區。
- `plan-book-screens.json` 的 80 張圖像對拍由上述計畫取代（原本把左右頁文字整片列為
  允許差異，現在文字全部受檢），該計畫縮為封面、目錄頁與空白 New Character 頁三對，
  對應的 `shots/newchar-0*.png` 一併移除。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：標題序列從觀察改為靜態讀出，並完成畫面對拍

- 新增 `tools/ida_probe_title_sequence.py`，匯出 `TITLE.EXE` 的 `_main` 與其全部直接
  被呼叫者，並解析每個呼叫點的引數。兩個全新 IDA DB 的輸出逐 byte 相同。
  **踩到的坑**：第一次跑加了 `-e HOME=/w`，結果 rc=1、零輸出、沒有 `.i64`——`idapyswitch`
  的選定結果寫在映像自己的 HOME，覆蓋 HOME 就讀不到。拿掉之後一次成功。這正是
  IDA 知識庫說的「靜默失敗」樣態。
- 重建結果寫成 `docs/re/title-sequence.json`：每一步都帶原始呼叫位址，等待是 BIOS tick。
  先前記為「機制未確認的閃爍段」是 SET0／SET1／SET2 的四圈疊圖動畫；疊圖矩形
  (112,53,88×70)／(96,53,104×84)／(48,53,160×125)／FACE (104,80,112×80) 與四個 archive
  的實際尺寸完全吻合。
- 動手實作之前先用抽出的 PNG 做合成假設檢定：原版 observe 兩輪共 81 張畫格中，
  66 張的 SHA-256 直接命中「背景 + 疊圖」的合成結果，確認合成規則在動工前就已成立。
- Go 端新增 `internal/title`（讀腳本、展開成畫面清單，亂數可注入）與
  `-title-shot-dir`／`-title-choice`。`-title-choice first`／`second` 固定取兩處亂數的
  各一個分支，聯集涵蓋序列會出現的所有畫面。
- 新增 `tools/verify_title_parity.py`：判準是集合包含（原版每一張都要能在 remake 的
  畫面集合裡找到逐像素相同的一張），找不到再判斷是不是取樣落在整張搬移途中。
  兩輪各 33 張逐像素相同、3／4 張是更新中途、**0 張未解釋**。
- 過程中一度把「原版有、remake 沒有」的畫面當成缺陷，實際是 remake 的亂數只走了一條
  分支；改成窮舉兩條分支後就對上了。這也是判準要寫成集合包含而不是逐幀位置對位的原因。
- 全部執行使用有界 `docker run --rm --network none`（IDA `--memory 4g --cpus 2`）；
  結束檢查沒有容器殘留，其他專案資源未動。

## 2026-09-01：TITLE 雷聲的規格近似合成

- 補讀音訊與計時原語（`sub_10F33`／`sub_10F48`／`sub_10F5D`／`sub_11904`／`sub_118D1`），
  把它們加進 `tools/ida_probe_title_sequence.py` 的匯出清單。
  `sub_10F5D` 寫的是 **PIT 除數**（低位再高位寫入埠 `0x42`），不是頻率；
  `sub_118D1` 數的是「一個 BIOS tick 內某個迴圈跑得完幾圈」。
- **長度為什麼不能靜態讀出**：`sub_10F75` 的步進間隔用的是校正值除以 8 的忙等，但
  校正迴圈裡有 `int 1Ah`、延遲迴圈沒有，兩個迴圈的每圈成本不同，因此換一台機器就換一個
  長度。這正是 SPEC-004「硬體時序停止線」要擋的東西，所以長度取自 DOSBox-X 擷取量測值。
- 新增 `tools/synthesize_title_thunder.py`：兩段音爆、每段 100 次除數更新、除數為
  `500×段次 + rand()%10000 + 1 + 1000`、方波頻率 `1193182/除數`、兩段間 50% 機率多等
  一個單位。輸出 `assets/audio/title-thunder.ogg`（Vorbis、44.1 kHz、單聲道、`-q:a 5`）。
- 技術驗收：時長 0.961451 秒，與原版四段非靜音視窗平均值 0.960870 秒差 0.000581 秒；
  解碼通過、無靜音段、峰值 −7.5 dB／平均 −10.3 dB。收據在 `docs/verification/audio/`。
- **人耳驗收尚未執行**，需使用者聆聽；在那之前 SPEC-004 維持 DRAFT，也不接
  Ebitengine 音訊路徑。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：繁體中文介面（雙模式）

- 使用者 2026-09-01 決定的雙模式落地：英文模式維持 320×200 與原版 8×6 字模，
  嚴格對拍不受影響（三份收據重跑後畫面逐檔相同）；中文模式改用 640×400 畫布、
  書本美術最近鄰放大、16×18 中文字模格。兩個模式共用同一個角色建立狀態機。
- 新增 `tools/build_cjk_font.py`：用 ffmpeg 的 drawtext 在固定格點上烘字模，再以固定
  門檻二值化。執行期只讀烘好的 PNG，不在執行期做字型排版。
  **踩到兩個坑**：(1) 直接用 `text=` 傳字元會被 drawtext 的參數解析與 text expansion
  吃掉（`%`、`\`、引號），改用 `textfile=` 加 `expansion=none`；(2) 相鄰字模的側邊會
  突出到隔壁格，量共同邊界時被算進來，把格距放大到 32 並內縮 8 px 才乾淨。
- `internal/font` 加上非連續碼位字型的支援（逐字模位置與逐字字寬），原版英文字型
  仍走 first_codepoint 起算的連續索引，行為不變。
- 新增 `docs/i18n/newchar-zh-Hant.json`（狀態沿用術語表三級）、`internal/i18n`、
  `internal/book/layout_cjk.go`（**自訂版面，不是量自原版**）與 `internal/ui` 的中文繪製。
- 新增 `tools/verify_text_fit.py`：版面座標由 `cmd/kol -dump-cjk-layout` 匯出，用實際
  字模表的逐字字寬檢查每個字串的右緣與缺字。98 項全過。
- 字模表衍生自 Noto Sans CJK TC（OFL 1.1），授權全文已隨字模表入庫。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：翻書動畫

- `KOLBOOK.EXE` 的 `sub_114F5` 是封面翻開的動畫，表驅動：兩張參數表放在 DGROUP
  （`+0x64` 17×2、`+0x86` 18×3），每一步把畫面上一段 byte 欄整體左移一欄，再把目錄頁的
  一欄複製過來，只動第 10 列起的 180 列，每步等 3 個 BIOS tick。詳見 RE-048。
- **先驗證再實作**：在寫任何 Go 之前，用 Python 照讀出來的規則跑 36 步，與原版
  `accept-and-save-run1` 的動畫畫格比對——18 張裡 16 張逐像素完全相同，兩端只差滑鼠
  游標，剩下一張是取樣落在兩步之間的更新中途。規則成立才動手實作。
- 新增 `tools/ida_probe_function_script.py`（通用呼叫腳本匯出，取代再寫一支近似的探針）、
  `tools/ida_probe_kolbook_book_animation.py`（讀參數表）與
  `tools/extract_book_animation.py`（以位元組序列在 EXE 內唯一定位，產出 typed data）。
- Go 端 `internal/book/animation.go` 在 CPU 端的 RGBA 緩衝上重播（原版動的是顯示記憶體），
  `cmd/kol -scene book-open` 可實際播放、`-book-shot-dir` 逐步輸出。
- `plan-book-open.json` 18 對全通過。參數表與角色資料同屬原版衍生 typed data，
  一樣留在本機由抽取器重生。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：角色選項頁、角色清單與兩段翻頁

- 新增正常 DOS UI 情境 `list-characters` 並跑兩輪，六個 checkpoint 的畫面 SHA-256 逐一
  相同、三個角色檔全程未變。這是本輪唯一一次新增原版擷取。
- 用抽出的原版字型從清單畫面讀回 14 個名字，再拿這 14 個名字回 `CHARDATA` 搜尋——
  每一個都唯一命中槽位起點，槽位版面因此是**驗證出來的**而不是推定的：名字在位移 0、
  所在地是位移 64 起的固定 10 bytes、城鎮代碼在位移 144。詳見 RE-049。
- 兩段翻頁沿用 RE-048 的同一張表。逐像素比對逼出兩條原版行為：反白是動畫**結束之後**
  才畫上去的；離開選單進入清單時原版**先取消反白再開始捲動**。把反白放錯位置時逐張差
  281 px，放對之後差 0 px——這種差異用眼睛看不出來，只有逐像素比對會講。
- 中文模式補上角色選項頁與角色清單頁。過程中發現 ASCII 用半形固定字距會把相鄰字模切掉
  （Noto 的 ASCII 是比例字），改成逐字取墨跡寬度；名字欄 16 個最寬字元也放不進標籤同一列，
  改成另起一行——原版英文版面本來就是這樣處理名字的。
- 文字安全矩形檢查擴充到 110 項（含右頁 12 項清單、六項選單、16 槽清單的底緣），全過。
  書頁底緣裝飾從原版第 190 列開始，安全矩形因此收到 y=380。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：角色寫入的逐 byte 對拍

- 以 `--save-character-files` 重跑 `accept-and-save`，取得原版寫入前後的 `CHARDATA`。
  原版只改動槽位裡的 58 個 byte。判準因此不是檔案雜湊相同（屬性是擲骰的，不可能一樣），
  而是「原版改動的每一個 byte，remake 都要寫出相同的值，且不能寫錯或多寫」。
- 欄位語意大部分早就在 `KOLBOOK-character-writer-summary.json`；本輪補上兩件事：
  金冠是 **big-endian**（摘要記成 little-endian，該處為誤），以及起始武器樣板來自
  `ITEM.LST`（61 筆 × 16 bytes）。
- **一個被推翻的推論**：只有 BARBARIAN 一筆資料時，`ITEM.LST` 索引看起來等於武器
  selector ＋ 3（兩個點都成立）。特地跑第二個職業（新增 `accept-and-save-ranger` 情境）
  才發現 BR 9→5、LB 3→25 完全不是算術關係。**一筆資料能支持的推論比看起來少得多**；
  現在只登記觀測過的職業，writer 對未觀測職業直接失敗，不寫猜出來的索引。
- 結果：BARBARIAN 與 RANGER 各 58/58 逐 byte 相同、寫錯 0、多寫 0。
- 建立→寫入→清單閉環：用腳本建立 `kai` 寫進槽位 14 之後重新讀出清單，14 列變 15 列。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：起始武器的來源與索引規則

- 上一輪把武器樣板的來源記成 `ITEM.LST`，那是錯的。反追 `sub_11820` 的定址
  （`[bx+79EEh]`）發現真正的來源是 `BOOKDATA` 的第 3 個項目——`_main` 在 `0x10257`
  把它載到 dgroup+0x79EE，解出 225 bytes ＝ 14 筆 16-byte 記錄，與 14 個武器 selector 一致。
- **同樣的武器記錄同時出現在 `ITEM.LST` 與 `BOOKDATA` 第 3 項**，所以拿寫入的 bytes 去
  搜尋會在 `ITEM.LST` 裡唯一命中——一個正確但無關的位置。**位元組相同不等於來源相同**；
  這一次是靠讀定址的程式碼才確定的，不是靠搜尋。
- 索引規則讀出來了：職業 metadata 第一個 byte 的低 nibble ＝ 主武器 selector、
  高 nibble ＝ 副武器 selector（0 表示沒有），記錄索引 ＝ selector − 1。
  39 筆職業裡 21 筆沒有副武器，原版對那些職業不動 `0x57` 與 `0x5f`。
- `tools/extract_item_templates.py` 由 `tools/extract_weapon_templates.py` 取代；
  writer 不再只認觀測過的兩個職業，39 個都能寫。兩個有原版對照的職業仍是 58/58。
- 新增 `internal/chardata` 的單元測試：逐筆檢查 39 個職業寫得出武器樣板、沒有副武器的
  不會誤寫，以及欄位語意（名字 NUL 結尾、種族 selector、屬性、big-endian 金冠、常數 3）。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：DELETE CHARACTER

- 新增三個正常 DOS UI 情境：`delete-character-probe`（進畫面即取消）、`delete-character`
  （反白條移動 ＋ 叫出確認視窗）、`delete-character-confirm`（回答 YES 實際刪除，再回答
  NO 驗證取消分支）。後兩個各跑兩輪，13/13 與 15/15 個 checkpoint 的 PNG SHA-256 逐一
  相同，連 TITLE 標題序列那一張都對上。
- **第一次試跑整串失步**：`delete-character-probe` 的八個 checkpoint 全部停在目錄頁，
  86 張畫格裡只有 29 張不同。成因是目錄頁按 5 之後有約六秒的翻頁動畫，靠「畫面靜止」
  判斷送鍵時機會在動畫途中誤判已靜止，後續按鍵全被吃掉。改用固定等待（8／9／10 秒）
  跨過那段動畫之後一次就過。
- **確認視窗不吃方向鍵與 Return**：這不是推測，是送出這兩個鍵之後畫面逐像素完全沒變
  （`delete-character` 的 checkpoint 10–12 是同一張 PNG）。視窗上寫的是 `YES/NO`，
  改送字母鍵才有反應。
- **刪除只寫一個 byte**：兩輪的 12288 bytes `CHARDATA` 差分完全相同——slot 1 的 offset
  `0x000` 由 `0x50` 變成 `0x00`，其餘欄位與 `CHARPICS`／`SHLDPICS` 都沒動。畫面上名冊
  往上補齊是「只列已建立槽位」的結果，不是資料搬移。
- remake 側：`internal/roster` 的 `DeleteFlow`、`internal/chardata` 的 `DeleteSlot`、
  `internal/ui` 的 `DeleteList`／`DeleteListCJK`，加上 `cmd/kol -delete-script` 的
  逐步截圖。對拍 `plan-delete.json` 8 對、`plan-delete-move.json` 5 對全過；
  `tools/verify_chardata_delete.py` 確認 remake 的輸出與原版**全檔逐 byte 相同**。
- 中文模式一併補上確認視窗（原版視窗放大兩倍再加高），新增七個漢字需要重烘字模表。
  烘焙改由 `tools/run_cjk_font_build.sh` 在有界容器內做：映像本身沒有中文字型，
  把主機的 Noto 目錄唯讀掛進去、在容器 tmpfs 重建 fontconfig 快取。222 字、
  文字安全矩形 113 項全過；重烘之後既有中文畫面的 PNG 逐檔沒有變化。
- 順手修掉一處陳舊標記：`docs/spec/README.md` 的狀態表停在 SPEC-002／003／005 的舊狀態，
  各規格檔本身早已是 READY／CONFORMED。以規格檔為準把索引補齊。
- 八份對拍收據重跑一次確認沒有回歸：171 對全過，既有收據逐檔沒有變化。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：書本剩下三項功能

- 六個新情境（三個 probe、三個 enter）並行跑完，`CHARDATA`／`CHARPICS`／`SHLDPICS`
  全程未變。**四個功能共用同一張角色選擇清單**：`DELETE CHARACTER`、`EDIT PICTURE`、
  `EDIT SHIELD`、`HARBOR MASTER` 進入後的第一張畫面是同一個 PNG（SHA-256 完全相同），
  所以刪除那一版的清單繪製直接就是共用的選擇畫面。
- `HARBOR MASTER` 在只安裝一個 region 時是一張三行訊息頁，已實作並對拍通過
  （允許差異只有滑鼠游標與頁尾殘影）。多個 region 的畫面需要另一張角色磁碟，未量測。
- `EDIT PICTURE` 與 `EDIT SHIELD` 是同一個點陣圖編輯器（左側編輯區、右上預覽、
  16 色色盤、底部 8 個工具圖示，畫在紅磚底上）。**ESC 連送三次畫面逐像素沒變**，
  原版是滑鼠驅動的，現有的鍵盤 oracle 走不完——這一項需要能送滑鼠事件的 oracle 才能繼續。
- 中文模式一併補上訊息頁；原版把訊息斷成兩行是英文字距下的排版，中文模式頁寬放得下
  整句，因此併成一行。字模表重烘為 237 字，文字安全矩形 115 項全過。
- 全部執行使用有界 `docker run --rm --network none`；三個容器並行時佔 3/14 核，
  結束檢查沒有容器殘留。

## 2026-09-01：走出書本，進到城鎮組隊

- oracle 第一次開出 `KOLBOOK` 之外：目錄頁按 `6`（Play the game）→ 共用的角色選擇清單
  → 選領隊 → 城鎮的半途之家旅店。**那張角色選擇清單和書本四個功能是同一個 PNG**，
  所以現在有五個入口共用它。
- **圖示列不吃方向鍵**：Right／Left 送出後畫面逐像素沒變。翻參考卡才看到正解——
  「以小於號 `<` 與大於號 `>` 在圖示間移動」。這一條是先量到「方向鍵無效」，
  再回頭查文件確認，不是先猜。
- 送 `<`／`>` 又踩一個坑：US 排列上它們是逗號／句號的 Shift 位，
  `tools/x11_key_inject.py` 只查 keysym→keycode 會送成 `,`／`.`。改成從 keyboard
  mapping 的奇數位置判斷哪些 keysym 需要 Shift 才送得出去。
- 組隊流程走通：COMPANION（第 6 格）→ `Look for characters on hard disk?` → 該旅店
  可招募清單（半途之家 8 人，與 `CHARDATA` 的所在地欄一致）→ 上下選、Return 加入，
  頭像出現在頂端領隊旁邊。**整段不寫磁碟**，三個角色檔的 SHA-256 全程未變——
  與手冊「只在旅店 REST/SAVE/HEAL 才保存」一致。
- 四份新收據封存於 `docs/verification/character-ui/`（現共 21 個情境、31 輪、
  3,023 個去重畫格）。remake 側還沒有任何城鎮畫面。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-01：旅店畫面的組成

- 新增 `tools/decompose_town_frame.py`。判準不是「差異最小的那張就是底圖」，而是
  **分群之外的殘餘差異必須為零**——只看最小值會接受一張碰巧接近的圖。
- 旅店畫面 ＝ 紅磚底圖 ＋ 11 個矩形，殘餘 0 像素。底圖 `BOOKPICS[3]` 與 `KOLMPICS[0]`
  的 PNG SHA-256 相同，是同一張圖各存一份。
- **第一次搜尋圖示零命中，成因值得記下來**：拿完整的 24×20 KOLMPICS 圖示在整張畫面
  搜尋，六格全部落空。原因是畫面上的外框會蓋掉圖示左右各一欄，比對永遠差那兩欄。
  改成比對框內 22×20 對圖示第 1..22 欄之後，六格全部逐像素零差異。
- 店主肖像仍找不到來源素材：30 個 archive、尺寸 60×40 到 200×160、±6／±4 位移、
  放寬到允許 400 個差異像素，全部零命中。先記成未知，不猜。
- 全部執行使用有界 `docker run --rm --network none` 或純本機讀檔；沒有容器殘留。

## 2026-09-01：旅店畫面的 remake 對拍

- `internal/town`（版面與狀態）＋ `internal/ui.Inn`（繪製）＋ `cmd/kol -scene inn`。
  對拍八對全過：六個圖示白框位置對 `town-icon-move-run1`，單人與招募後兩人隊伍對
  `town-recruit-companion-run1`；允許差異區之外零像素，允許區只有原版滑鼠游標。
- **店主肖像找了三輪才找到**。前兩輪用「尺寸範圍 ＋ 位移搜尋」全部零命中；
  改成從畫面上取一列 60 個像素當指紋、去所有素材的像素緩衝區 `find`，一次命中
  `TWN7PICS[12]`（88×64，畫在 (8,64)，88×64 像素全同）。
  **不確定素材在哪裡時，拿畫面的一列去搜素材比拿素材去搜畫面便宜得多。**
- 第一版對拍差 404 px，兩個成因都在隊員格：名牌下方沒填黑（紅磚透出來），
  以及少了右下的投影。填黑加投影之後降到 31 px，全部落在滑鼠游標。
- 城鎮的文字用 8×8 大字型，不是書頁的 8×6 小字型；名牌則是小字型且只畫前五個字元。
- 中文模式一併補上。**烘字模表時踩到一個坑**：城鎮樣本自帶的中文台詞不在
  `docs/i18n/` 裡，第一次烘出來的表沒有那些字，畫面上會變成空白。
  `tools/build_cjk_font.py` 因此加了 `--extra-json`。
- 十份對拍收據重跑：180 對全過。文字安全矩形 124 項全過（含城鎮台詞的行寬與行數）。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-02：玩家可見的展示截圖

- 新增 `docs/screenshots/`：19 張展示圖 ＋ `shots.spec`（每行一條產生指令）＋
  `MANIFEST.json`。產生走 `tools/run_kol_shots.sh`——**一個容器建置一次、起一次 Xvfb、
  跑完整份規格**，而不是每張圖開一個容器把建置成本重複十九遍。
- `tools/build_screenshot_manifest.py` 就地驗證：PNG 解得開、尺寸符合該模式、顏色數
  不少於 4，且**有對應對拍收據的畫面必須與那張逐 byte 相同**——展示圖不是另外畫一份，
  就是通過對拍的同一批像素。五張有對應的都逐 byte 相同。
- `cmd/kol` 加了三個只為展示圖存在的入口：`-script`／`-delete-script` 搭配 `-shot`
  時只輸出最後一張，以及 `-title-frame N` 輸出標題序列展開後的第 N 張。
- 中文模式的圖像層說明改用自己的一組行。原版把一句話拆成五行是英文字距下的排版，
  逐行硬譯會讀成「用方向鍵／選擇角色／圖像」這種斷句；中文模式合併成三行。
  字模表因此重烘為 344 字，文字安全矩形 122 項全過。
- README 從「尚未有可玩版本、SPEC-002／005 仍為 DRAFT」更新為現況，並加上畫面區塊。
- 十份對拍收據重跑：英文側的截圖逐檔重生後與版控的完全相同，180 對全過。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-02：城鎮戶外地圖的拆解

- 新增情境 `town-outdoor-move`，並用 `tools/decompose_town_frame.py` 拆出戶外畫面的
  五塊。徽章（`TWN7PICS[0]`，2,754 個不透明像素零差異）、圖示（`KOLMPICS[5]` 零差異）、
  日期框（小字型 `Apr 1`／`AD  1100`）與頭像格都對上了。
- **地圖視窗不是一張固定圖**。這一輪特地用三種互不相同的證據去問同一件事，
  而不是換參數再問一次：
  1. 從地圖取八段 24 像素指紋去 752 張 PNG 搜尋，全部零命中；同尺寸素材逐張對位，
     最好的一張差 12,535／18,432 像素。
  2. 移動時視窗以 16 px 為單位捲動（往右一步不動、往右兩步位移 16 px 後 97.9% 相同）。
  3. 以 16×16 切格時，位移 (0,0) 只有 47 種不同磚（65%），其餘 255 種對齊都 87% 以上。
- 因此戶外地圖是 12×6 格、每格 16×16 的捲動磚塊圖，磚圖來源不在抽取集合裡。
  **這一輪就停在這裡不實作**：把 18,432 個像素的地圖區列成「允許差異區」等於用允許
  掩蓋未實作，違反對拍判準。下一步是讀 KOLTOWN 的地圖繪製路徑。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-02：城鎮地圖的磚圖找到了

- **`TOWN0.0`…`TOWN7.0` 七個 50–65 KB 的檔案從來沒被分析過**，因為清冊工具的
  `is_candidate()` 只認名字含 `PIC` 的檔案。這是自己的過濾器有洞——查「地圖像素在哪」
  的時候，答案一直躺在 inventory 之外。
- `TOWN7.0` 用 `*PICS` 同一套容器契約解得開（26 個 payload）。磚圖在 payload 0／1／2，
  每個 12,288 bytes ＝ 96 塊 × 128 bytes。
- **磚圖格式靠搜尋定案，不是靠猜**：把畫面上 72 塊磚各自編成四種候選格式去搜，
  67 塊命中而且**全部命中同一種**——16×16、四個 EGA 平面、plane-major；
  另外三種（high／low nibble 封裝、planar row-major）一塊都沒中，offset 全是 128 的倍數。
  沒命中的五塊是隊伍圖示覆蓋的那一小片。
- 地圖配置的編號還沒解出來：六種編號方式在 2,048 byte 的 payload 裡都零命中，
  288 塊磚也放不進一個 byte。下一步要讀 KOLTOWN 的地圖繪製路徑。
- **這一輪也補上一個一直沒補的缺口**：新增 `tools/run_python_tool.sh`，
  讓純 Python 的解包與分析也走有界容器（原版 ZIP 唯讀掛載），不再在主機上跑。
- 全部執行使用有界 `docker run --rm --network none`；結束檢查沒有容器殘留。

## 2026-09-02：磚圖載入路徑，以及兩個被推翻的假設

- 新增 `tools/ida_probe_koltown_town_map.py`（探索用，不是鎖定斷言）。
  **第一次跑得到「沒有人用這些字串」的假零**：`town%d.0` 等字串的 `DataRefsTo` 全空。
  16-bit real mode 是把 DGROUP 位移當立即值載入，IDA 不產生 data xref；
  改掃立即值之後 18 個使用點一次浮現。
- `sub_15920` 把 `town%d` 的 payload 0／1／2 放到同一 segment 的 0x0000／0x3000／0x6000
  （0x3000 ＝ 12,288，三個首尾相接），再對 `+0x100` 呼叫 `sub_19B29` 轉色盤索引。
  這與 RE-057 的位元組證據互相印證：288 塊磚是平坦陣列，第 n 塊位移 `n×128`。
- 讀出 `sub_1C80F` 的 nibble RLE 版面（big-endian 表頭 ＋ 控制流 ＋ nibble 流）。
- **兩個假設被自己的實作推翻**：(1)`sub_2A0C6` 組檔名用的是 `twn` 那組字串，
  不是 `TOWNn.0`；(2) 把 nibble-RLE 套到 21 個 2,049 byte payload 上，
  每一個的表頭都指向 2,132–16,589，全部超出長度，一個都展不開。
  兩條都寫進 RE-058 的「被推翻的假設」，不是默默改掉。
- 地圖格子陣列仍未定位。下一個錨點已經寫在 WORKLIST 上。
- IDA 在有界容器內執行（`--memory 4g --cpus 2`、`--network none`），沿用既有 `.i64`
  避免重複自動分析；結束檢查沒有容器殘留。

## 2026-09-02：地圖格子陣列——六種問法全部零命中，停止資料側硬搜

- 接續 RE-058，又試了三種問法：偏移不變的差分搜尋（磚塊編號整體加常數也能中）、
  把 `sub_1C80F` 的 nibble RLE 實作出來套上去、掃全部 338 個函式找「乘以 128」的算式。
  **三種都零命中**，連同先前三種共六種，全部記進 RE-058 的表。
- 過程中修正一個誤讀：`word_30C8A` 不是磚塊專用緩衝區，而是**離屏畫面緩衝區**
  （`sub_16986` 先把它設成繪圖目標 `word_2DF12`，之後才換成 `0A000h`）。
  磚圖是載入時借用它當暫存區。
- **停在這裡不再對資料硬搜**：同一個問題已經換六種方式問過，再換第七種也只是抽樣。
  下一步改從城鎮畫面流程（`sub_16183`／`sub_1766F` 一帶）往下追真正的繪製迴圈。
- IDA 全部在有界容器內執行、沿用既有 `.i64`；結束檢查沒有容器殘留。

## 2026-09-02：城鎮戶外地圖，從格子陣列到逐畫面對拍

- **地圖格子陣列解出來了**（RE-059）。前一輪六種資料側搜尋全部零命中，這一輪改追
  繪製迴圈：`sub_115D1` 是 12×6 的雙重迴圈，格子索引 `(X<<6)+Y`，X 下限與 Y 上限在
  程式碼裡靜態就是 `0` 與 `3Fh`。關鍵那一步在 `sub_1D1C6`：`shl ax,7` 是磚距 128，
  來源段是「緩衝段 +0x100 paragraph」——**+0x1000 byte**，所以段內前 4,096 byte
  不是磚圖而是格子陣列。
- **零命中的成因**：前一輪把三個 payload 當成各自 96 塊、合計 288 塊，於是推論
  「一個 byte 放不下」；實際是 256 塊，一格就是一個 byte。六種搜尋又都只掃 21 個
  2,049 byte 的 payload，沒掃 payload 0 的開頭——**範圍與編號同時偏掉**。
- **驗證用計分而不是全有全無**。第一輪要求 67 格全中，回零；改成計分之後
  最佳 60/67、次佳只有 12，一眼就看得出是對的。剩下 7 格的落差追下去是
  磚 193 與磚 210 的 128 byte 完全相同，比對器只回報編號較小的那一個。
  **全有全無的判準會把「幾乎全對」和「完全不對」混成同一個答案。**
- **隊伍標記的遮罩不是存的，是算出來的**（RE-060）。先找不到遮罩資料：
  1-bit、4-plane、KOLPICS 全 payload 都搜過，零命中。讀 `sub_1D21D` 才知道
  遮罩是把精靈四平面 OR 成佔用位圖、左右各膨脹一像素、再反相現算的。
  **找不到資料時要考慮「它根本不存在」，而不是換一種編碼再搜一次。**
- **一個像素卡了一輪**：照「一整列 16 像素」實作膨脹，原版畫面 255/256。
  成因是 8086 讀的是 little-endian word，而畫面左半 8 像素是低位元組——
  位移不跨位元組邊界。補上這條之後 256/256。
- **對拍用了另一輪 oracle**。`town-outdoor-move-run1` 的 94 張畫格裡 22 張是戶外畫面，
  `tools/solve_outdoor_camera.py` 把每一張的鏡頭、標記偏移與動畫畫格都反解出來，
  而且每一張都是唯一解。反解的結果同時揭露兩件原本不知道的事：
  **隊伍先在視窗內走、鏡頭才跟著捲一格**；**圖示列會依情境整格淨空**。
  兩張原版正在重繪的畫格（隊員格／徽章／日期框還沒畫完）排除在計畫外並寫明理由。
- 十一份對拍收據重跑：201 對全過。資產 752 → 769 張 PNG（八個城鎮的地圖與磚塊
  atlas、一張標記 atlas），既有 752 張位元組不變。
- **字模烘焙撞到既有限制**：全域水平位移是 −1，新加的「四月日西元」墨水延伸到第 15 欄，
  聯集變成 17 欄寬而整批失敗。改成全域位移照舊、個別字模超出字格才單獨夾回來，
  並把錯誤訊息從「超出 16x16」改成列出是哪幾個字撐大的。字模表 344 → 371 字。
- 全部執行使用有界 `docker run --rm --network none`；一次因為截圖規格漏了 `-shot`
  讓程式進入互動迴圈而逾時，留下一個容器，已 `docker stop` 自己那一個，未動他人資源。
  收工檢查沒有本專案容器殘留。

## 2026-09-02（續）：門口、建築內部，以及城鎮文字的 5-bit 編碼

- **門的判定來自掃門迴圈**（RE-061）。payload 3 的 `[0x20]` 是門數、`[0x47+2i]` 是座標，
  命中時 `word_2FBC2 = 2i`——門 *i* 的四份資料（門面圖 `2i+3`、肖像 `2i+2`、
  文字 payload `i+4`）全部由同一個索引推出來。八座城的門數與 `TWN{n}PICS` 配對數
  逐一相等，是很強的交叉檢查。
- **門口是狀態不是位置**。掃門迴圈接在移動後面，所以剛走出旅店站在門的格子上仍是
  地圖畫面——原版畫格 068–076 就是這樣。第一版把它寫成位置的函數，
  一render 就發現與既有的 21 對對拍衝突，改成明確狀態才對。
  **「同一個位置有兩種畫面」是狀態存在的證據**，不是資料矛盾。
- **文字編碼繞了一大圈**。明文、XOR、位移、固定字母表的位元打包全部零命中；
  掃 `and ..., 1Fh` 找解碼器在 338 個函式裡**一個都沒有**。
  最後是用畫面上的原文當 **crib 去掃位元偏移**才破的——命中點全落在 `16k+1`，
  而 15 byte × 8 ÷ 5 ＝ 24 正是對話框欄數，版面當場成立。
  **有已知明文時，crib 比找解碼器便宜得多**；而「掃不到解碼器」本身也是資訊，
  它排除了「先讀碼再讀格式」這條路。
- **符號靠統計定案**：`[30][27]` 有 1,721/1,796 次落在整格開頭後接字母（開頭引號），
  `[31]` 後面幾乎只有 s／t／v／m／l（`'s`、`'t`、`'ve`）。
  剩下兩個未定的（引號前空白、`[27][27]` 收尾）已寫明證據與影響範圍，沒有硬猜。
- **畫面規則只有兩條**，由原版畫格 072 的八行逐格讀回比對出來：句號後兩個空白＋
  下一字母大寫、逗號後一個空白。套上之後八行逐字相同。
- 進門畫面**完全由三份 typed data 組出來**，沒有寫死字串；換一個門序號就畫出鐵匠鋪。
  第一版差 219 px 在隊員格，成因是外框顏色——進門當下說話的是建築裡的人，隊員不是
  發言者。
- 十二份對拍收據：203 對全過。展示截圖 26 張。全部執行在有界 `docker run --rm
  --network none`；收工檢查沒有本專案容器殘留。

## 2026-09-02（續二）：城鎮這一段可以實際玩了

- **輸入迴圈接上了**。`internal/town/session.go` 是三種模式的狀態機（地圖／門口／
  建築內），`cmd/kol -scene play` 走 Ebitengine 的 Update：方向鍵與參考卡的斜走鍵
  走動、`<`／`>` 移動白框、Return 執行。狀態機不碰繪製也不碰輸入裝置，可單獨測試。
- **鏡頭跟隨規則是反解出來的**：22 張原版畫格顯示偏移永遠夾在 0…1，多出來的由鏡頭
  吸收。寫成測試直接重播那六步移動，鏡頭與偏移每一步都要相同。
  **這是強推論不是已證實**——規則沒有從程式碼讀出來，收據裡寫明了。
- **玩家路徑驗證器抓到一個真 bug**。`tools/verify_player_path.py` 的第二條檢查是
  「相鄰兩步不得完全相同」，結果 `>` 那一步畫面沒變——門口圖示列的白框位置
  根本沒有畫出來（`ui.Outdoor` 讀的是 `Outdoor` 自己的游標，不是 session 的）。
  **「每個操作都要有可見效果」比「畫面看起來對」抓得到更多東西。**
- **可玩路徑與對拍證據連上了**：玩家走到門口那一步的截圖，與已通過逐像素對拍的
  門口截圖**逐 byte 相同**；進門那一步同理。這條連結是「沒有除錯捷徑」的實證，
  不是宣稱。
- 十二份對拍收據 203 對全過；`internal/town` 新增 session 與鏡頭跟隨測試。
- 全部執行在有界 `docker run --rm --network none`；收工檢查沒有本專案容器殘留。

## 2026-09-02（續）：城鎮地形擋路、地圖邊界的城門提示、圖示列生命週期

- IDA 探針 `tools/ida_probe_koltown_world_map.py`（探索用，不是鎖定斷言），
  以 `ida-pro-9.4-idapython:locked-v1` 有界 `docker run --rm --network none` 執行四輪，
  逐步加大 `DUMP_FUNCTIONS`：`sub_10000`／`sub_15920`／`sub_29004`／`sub_29CA3`／
  `sub_1198E`／`sub_11302`／`sub_11703`／`sub_11722`／`sub_1449C`。
- **走錯過一次方向**：一開始想從「`word_2FBC2 = 0FFFFh` 是哪裡寫的」直接切進去，
  但 0FFFFh 的寫入點全檔有 62 處。真正省時間的是先讀說明書第四章——它把
  「走到城鎮地圖邊緣 → 地圖換成林地圖 → DOOR 離開／U-TURN 回城」整段流程寫清楚了，
  再回頭讀反組譯就只剩對照。
- 新探針 `tools/probe_town_passability.py` 解出不可通行位元圖並做兩項交叉檢查：
  TOWN0 全零、103 個門座標全部落在可通行磚塊上。
- `tools/extract_town_buildings.py` 加出 `blocked_bitmap`／`blocked_tiles`；
  重生 `assets/original-derived-data/town-buildings.json`。
- Go：`Outdoor.Walk` 改寫成與原版同順序（先查地形 → 動標記或鏡頭 → 夾限鏡頭並計數），
  `Outdoor.AtEdge`、`town.ModeGate`、`Session.LeftTown`；圖示列改成會清空重建的堆疊。
  `cmd/kol` 加 `-outdoor-at-edge`／`-outdoor-selected`。
- 原版 oracle 情境 `town-map-edge`（DOSBox-X，一個容器、1 核，約 4 分鐘），
  124 張畫格；`tools/solve_outdoor_camera.py` 反解每一張戶外畫格的鏡頭都是唯一解。
- 對拍 `plan-town-edge.json` 4 對全過（允許區只有原版滑鼠游標）；
  既有 `town-outdoor` 22 對、`town-building` 1 對重跑仍全過。
- 玩家路徑新增 `town-edge.json`（13 步，兩個檢查點與對拍截圖逐 byte 相同）；
  既有兩條路徑重生後仍 0 失敗，變動的畫格正是「走離門口後圖示列清空」那幾張——
  與原版 `frame-074` 一致。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續二）：七座城的建築文字全部翻完，順帶修掉字模基線

- 先解世界地圖的城鎮記錄（RE-066）才動筆：`sub_29CA3` 把 22 筆 16 byte 記錄拆成
  世界座標、進城鏡頭、四個界限、三字母代碼與磁碟號，並確定「記錄 i → TOWN{i+1}，
  記錄 6 之後全部是 TOWN7 的入城點」。城名再由城鎮對話的一手證據對上
  （「north is Htron」「west is Thimblewald」等）。
  **`towns.7.name` 原本寫成「布瑞提爾」是錯的**——TOWN7 是 the Hobe，
  Brettle 是 TOWN1；已改正。
- `docs/i18n/town-zh-Hant.json`：補上 1–6 共 80 棟建築的名稱、人物與進門敘述，
  加上城名與三字母代碼對照。文字安全矩形 863 項全過。
- 說明書術語表把因此定案的十條由「待查」改成「暫譯」並補來源；
  `Shellemoon` 改成資料裡實際的 `Shellernoon`（一手資料勝過說明書拼法）。
- **字模烘焙修了兩個缺陷**：
  1. 整張表一次交給 ffmpeg 會踩到單一參數 128 KB 上限（`Argument list too long`）。
     改成濾鏡圖走檔案，再分批（每批 16 列）烘完拼接。
  2. **drawtext 依「這一次要畫的字串」的最大 ascender 對齊**，所以一個字單獨畫時，
     墨跡頂端一律貼齊格子上緣——逗號、句號被拉到半空中，而且**這個缺陷從一開始就在**，
     既有的中文截圖全部有。改成每格先畫一個全形驚嘆號當基準字，兩個字共用基線；
     基準字本身的墨跡頂端每格量一次，所以目標字比基準字高時的整串位移也會被抵銷。
  診斷方式是先烘一張「單獨畫」與「連在一起畫」的對照圖，看出兩者高度不同。
- 截圖重生 31 張（新增霍布以外兩座城的中文建築畫面與英文城門提示），manifest 0 失敗。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續三）：建築內的話題系統

- 先讀說明書第二章（LISTEN／MOUTH／MIRROR 各是什麼），再回頭讀 `sub_11A93`
  的圖示分派表，兩邊逐條對上；圖示編號用 KOLMPICS 接觸表肉眼指認。
- 關鍵字表在 `TOWNn.0` payload 20。`sub_1B94F` 的 **mode 1** 與 mode 0 不同
  （字數在高 4 bit、位元流從 bit 3 起、跳脫預設打開），一開始照 mode 0 解出來是亂碼，
  讀了 `0x1B962` 那幾行才解對。七座城 305 個關鍵字全部是四字元字首、字典序排好。
- **踩到一次條件讀反**：`0x1258C` 是 `jz` 才往下走，我第一版寫成「話題碼為 0 就跳過」，
  結果 LISTEN 講出應該要用 MOUTH 問才有的內容。對拍時原版的畫面文字對不上才發現。
  正確的是：LISTEN 只講**沒有掛關鍵字**的槽。
- 新抽取器 `tools/extract_town_keywords.py`、`tools/extract_ui_text.py`；
  `extract_town_buildings.py` 加出圖示旗標（payload 位移 0x3C／0x3D）。
- 交叉檢查：101 棟建築共 409 個非空話題槽，每一個的話題碼都在該城的關鍵字表裡，
  而且都有對應的訊息 20 + 槽號，零例外。
- 原版 oracle 情境 `town-listen`（一個容器、1 核，約 4 分鐘）；三對逐畫面對拍全過，
  允許區只有滑鼠游標。圖示列那一段逐 byte 相同。
- 新玩家路徑 `town-talk`（9 步，三步與對拍截圖逐 byte 相同）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續四）：布雷特爾的建築內對話全部翻好

- 譯文表加了 `messages`（以訊息編號為鍵）與 `ui`（KOLDATA 記錄號為鍵）兩個欄位；
  `Session.message` 與新的 `uiRecord` 都改成「譯文優先、沒有就退回原文」，
  因此英文模式的對拍不受影響（重跑仍 3/3、玩家路徑 0 失敗）。
- 布雷特爾 16 棟建築共 303 則對話翻完；`verify_text_fit.py` 加上 `messages` 的檢查，
  檢查項從 863 成長到 1,772，全過。
- 新增中文 LISTEN 的兩張示範截圖，另存一份布雷特爾的戶外樣本
  `docs/re/town1-outdoor-sample.json`（示範用，不是對拍輸入）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續五）：野外地圖畫面

- 新抽取器 `tools/extract_world_map.py`（`TOWN0` payload 3 的地形／道路／不可通行／
  城鎮記錄）；交叉檢查：隊伍離開霍布礦坑之後站的 (40,23) 是 tile 159、地形 PLAINS，
  與原版畫面左下的「DAWN PLAINS」對得上。
- 新套件 `internal/world`：次格移動（0…7 繞回才走一格）、地形／道路查表、
  城鎮入口比對、天候圖合成與 EGA 索引反相。
- **三件事是量出來的，不是猜的**：
  1. 標籤置中靠 KOLDATA 記錄自帶的空白，去掉空白就左移兩個字；
     因此另留一條「保留前後空白」的取字路徑。
  2. 隊伍那個點是把底下四個像素的 EGA 索引**與 15 互斥或**，
     `0 → 15` 與 `8 → 7` 兩組對應排除了「塗白」與「OR 15」。
  3. 雲層動畫**沒有**做出來：照 `sub_2A18E` 的「每 16 byte 向右旋轉一位元」
     用 64 或 128 像素繞回重現，對原版畫格的最佳吻合都還差四百個像素。
     不硬湊，改成在對拍計畫裡明寫不宣稱。
- 對拍工具加了 `compare` 欄位（每一對涵蓋的矩形）與計畫層的 `not_claimed`。
  這是為了「允許不等於忽略」：還沒實作的東西不塞進允許差異區。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續六）：城門接上野外地圖

- 比對 `TOWNn.0` payload 3 的 `0x27 + 2k` 與世界地圖城鎮記錄的座標，16 組全部相同
  （RE-069）。因此 RE-061 把那一欄叫「入城點在城鎮地圖上的位置」是錯的，
  它是**城門在世界地圖上的格子**；正文已改寫，勘誤集中在 RE-069。
- `town.Session` 加 `OnLeaveTown` 回呼；`cmd/kol` 的 play 場景在 `LeftTown` 時
  把隊伍放到那一格並改畫野外。
- 新玩家路徑 `town-leave`（12 步），最後一步與野外對拍截圖**逐 byte 相同**。
  既有四條路徑與五份城鎮／野外對拍全部重跑通過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。


## 2026-09-02（續七）：野外走到城鎮入口就能進城

- IDA 探針包成 `tools/run_ida_probe.sh`（之前每輪手打 `docker run`）。第一次跑
  沒有輸出，把 `>/dev/null 2>&1` 拿掉並加 `-L` 才看到真正的原因：`HOME` 指到 `/tmp`
  會讓 IDAPython 與授權都起不來，而 `idat` 的標準輸出是空的、離開碼是 1。
  包裝現在鎖 `HOME=/home/ubuntu`，失敗時直接印 log 尾端。
- 逐指令讀出 `sub_29BBB`（入口提示）、`sub_29CA3`（進城搬的九個欄位）、
  `sub_29987`（地形與道路）、`sub_29B40`（標籤）、`sub_11302`／`sub_11722`／
  `sub_1B100`／`sub_11390`（畫圖與等鍵盤那一串），寫成 RE-070。
- **一個判讀陷阱花掉三輪探針**：原版的提示畫面比 remake 多一個 14×11 的白箭頭。
  查過 `sub_29BBB` 的每一個呼叫、`sub_11302` 的繪製矩形、標記精靈圖集，
  又把全部已抽出的 PNG 掃過一遍找那個形狀（零命中）——它是 **DOSBox-X 的滑鼠游標**。
  同一輪的城門提示也有它，而更早那輪沒有。**既有的 `plan-world.json` 早就把它
  列成允許差異了**；先查自己的對拍計畫會比再跑一輪探針快。
- **一個真的錯誤被原版抓出來**：野外的地形標籤看的不是鏡頭那一格。
  原版讀的是視窗緩衝區第 0x29 個 byte（第 3 列第 5 欄），而列偏移是 −2，
  所以腳下那一格是 `(X, Y+1)`。地形名稱、道路旗標與地形擋路三處全部改掉；
  之前之所以沒被抓到，是因為 (40,23) 與 (40,24) 剛好同樣是 PLAINS。
- 新增 `internal/world.Session`（地圖／入口提示兩個模式、圖示列、進城回呼）與
  `town.Session.EnterFromWorld`；`cmd/kol` 的 play 場景現在把方向鍵按人在城裡
  或在野外分派給不同狀態機。
- 新玩家路徑 `town-round-trip`（19 步）：出城 → 野外走五步到路上 → 退回入口 →
  選 DOOR 回到 TOWN7。新對拍計畫 `plan-world-town` 七對全過（天候圖不宣稱、
  滑鼠游標列為允許差異）。全部 16 份對拍 221 對、6 條玩家路徑重跑通過。
- 新 oracle 收據 `docs/verification/character-ui/world-enter-town-run1`（383 幀）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續八）：野外畫面的中文模式

- `world.CJK` 版面：英文版把時段／地形擠在 67×15 的框裡，中文兩個字就要 32 px，
  因此左欄整個重排成日期框（114–158）、天候圖（162–292）、標籤框（296–340），
  圖示列往下移到 348。標籤在框內置中，不沿用原版靠記錄自帶空白置中的做法。
- KOLDATA 記錄 0D5h–0E1h（五個時段、八種地形）翻好進 `docs/i18n/town-zh-Hant.json`；
  `verify_text_fit.py` 加上野外標籤的檢查，1,789 項全過。
- **重烘字模表掉了五個字**。第一次只帶了兩個 `--extra-json`，烘出來的表少了
  反、嗅、嗨、戲、暫——`zh-town-inn.png` 的「他嗅了嗅」當場變成兩個空白。
  原因是 `--strings` 只讀字串表的五個鍵，中文台詞其實散在譯文表與四份畫面樣本裡，
  而 README 記的烘焙指令沒有帶齊。修法：把完整的輸入清單寫進 README，並要求
  **重烘之後先比對新舊 manifest 的碼位集合**，確認沒有字消失再往下走。
  補齊之後 1,689 → 1,714 個字模，舊截圖逐 byte 回到原狀。
- 新增展示圖 `en-world-town`（與對拍過的 `world/town-prompt.png` 逐 byte 相同）、
  `zh-world-map`、`zh-world-town`；37 張展示圖的 manifest 重建通過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續九）：野外的道路

- 道路表 `payload3[0x200 + 磚塊]` 存的是**八個方向的位元遮罩**，不是布林值。
  四張 DGROUP 表（位元、X 位移、Y 位移、鍵碼）原樣倒出來寫進 RE-071；
  ROAD 與 SIGNPOST 是同一支 `sub_29DDE`，差別只在要不要排除目前方向。
- **一次白跑的 oracle**：第一版腳本讓隊伍往東走到 (41,23) 才按 ROAD，
  結果圖示列是空的、按鍵完全沒有作用。原因是 RE-070 那條——腳下那一格是 (X, Y+1)，
  (41,24) 是森林不是道路。改成往北十二步站到 (40,21) 才對。
  **這也是自己的實作先發現、原版後印證的一次**：修好 `Tile()` 之後 remake 的圖示列
  就跟原版一樣是空的，反而是舊的截圖記憶誤導了腳本設計。
- 對拍抓到一個真的差異：野外圖示列的框色有**三種**。`sub_11722` 畫的是白與亮藍，
  但主迴圈自己把**離開的那一格**畫成青色（`sub_1B90B(舊格, 3)`）。
  按下 Return 之後白框還會回到第 0 格。兩件事都補上。
- 新對拍 `plan-world-road` 10 對全過（滑鼠游標列為允許差異；時段標籤在第八步之後
  不宣稱——時間推進還沒實作），新玩家路徑 `world-road` 34 步。
  原版畫格挑的是隊伍那個點與 remake 同相的那一張，因此不必為那個點開允許區。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續十）：招募同伴

- 建築內的圖示列現在真的能用：`<`／`>` 移動白框、Return 執行那一格，
  選到還沒實作的圖示會記進 `Session.Unimplemented` 而不是靜靜吞掉。
  原本 `Accept()` 在建築內一律「走出去」，那是對拍時期的捷徑。
- COMPANION 的兩個畫面與清單行為由 `town-recruit-list-run1` 量出來：
  一次八列、反白到底才捲、**兩端夾住不繞回**、右側兩個紅色箭頭一直都在。
- 候選人的規則是角色檔的所在地欄比對建築名稱前十個字母，半途之家 11 位，
  與原版畫面逐一相同。頭像則靠「存檔的 16 張頭像逐像素命中 80 張模板」
  這個檢查取得模板圖號——抽取器對不上會失敗，不會靜靜留白。
- **對拍先綠後紅的一格**：加入之後的招呼語原版是「你受傷了，該去修道院」，
  remake 卻畫一般招呼語。追下去發現 `sub_11A93:0x11F47` 讀的是角色記錄位移 0x1B，
  資料側只有三個人非 0，而被招募的 Poontang 正是其中之一。補上之後九對全過。
- 新對拍 `plan-town-recruit`（9 對）、新玩家路徑 `town-recruit`（15 步）、
  新 oracle 收據 `town-recruit-list-run1`（99 幀）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續十一）：赫特隆的建築內對話翻完

- 城 2（赫特隆）14 棟建築、274 則對話全部翻好。人名、地名沿用既有譯法
  （賽加利昂、皮爾達、公爵、戈爾、拉斯坦灣），新增的地名（費林岬、菲格拉灣、
  風徑、科拉沙嘴、泰加森林、卡尼許谷、德雷津、塔克斯朋）在城 2 內部一致。
- 字模表重烘：1,689 → 1,901 個字模，比對新舊碼位集合確認沒有掉字；
  `verify_text_fit.py` 2,611 項全過（0 超寬、0 缺字）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-02（續十二）：七座城的建築內對話全部翻完

- 城 7（霍布）21 棟、407 則翻完；至此七座城 101 棟建築的 1,938 則建築內對話
  全部有繁中譯文，加上名稱、人物與進門敘述，城鎮文字這一塊收尾。
- 譯法上的三個一致性決定：
  1. 凱爾登人（城 2 門 2、城 3 門 7）與阿瓦隆人（城 7 門 15–17）的破碎語法
     在中文裡保留同樣的殘缺句式，不改寫成通順句子——那是角色設定不是錯字。
  2. 城 4 門 3 的德語腔在中文裡以句式的生硬呈現，不用注音或錯別字模擬口音，
     那會犧牲可讀性。
  3. 專名一律沿用先前定下的譯法（賽加利昂、皮爾達、公爵、戈爾、奧蘭斯…），
     新地名在同一座城內先求一致，跨城再對齊。
- 字模表 2,255 → 2,366 個字模，比對碼位集合確認沒有掉字；
  `verify_text_fit.py` 6,693 項全過（0 超寬、0 缺字）。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03：客棧住宿、修道院治療與 remake 自己的存檔格式

- 追住宿價錢時先讀錯了一支。`sub_11A93` 裡有六處畫 KOLDATA 記錄 0x53（`it will cost
  you`），第一處是修道院、第二處是買馬，客棧那一支根本不在這個函式裡——記錄 0x37
  （`"If you want to stay`）在 `sub_11A93` 一次都沒出現。新探針
  `tools/ida_probe_koltown_rest_save.py` 把每個 `sub_1A8FD` 呼叫點畫的是哪一筆記錄
  一起倒出來，`sub_14DB1` 才浮出來。
- RE-073 收下三條路：修道院是「傷勢 × payload[0x39]」，客棧是
  `payload[0x38] << 8 | payload[0x39]`（0 代表免費），金幣是角色記錄 0x2A–0x2C 的
  24-bit 大端錢包。半途之家讀出來 56，與 `town-rest-run1` 畫面上的數字相同；
  兩人隊伍再跑一次仍是 56，價錢不隨人數改變。
- **匯入工具當場推翻自己寫下的斷言**。我把角色記錄 0x90–0x92 記成日期，
  匯入 Theodoric 之後那三個 byte 讀出來是 77／73／78 ＝ `MIN`——那是三字母城鎮代碼，
  和角色清單畫面右欄同一段。RE-073 的欄位表已改正，存檔格式也把沒有證據的
  `day`／`month`／`year` 拿掉：世界時間還沒有規格，格式裡就不留佔位。
  同一次匯入另外驗證了兩件事：0x2E／0x2F 就是建築 payload 的 `record_id`
  （12 名角色是 86 ＝ 半途之家、2 名是 14 ＝ 巨魔剋星客棧），
  0x97／0x13／0x99／0x9C 讀出來是城 7、入城點 3、座標 (57, 8)，正是那扇門。
- SPEC-006 READY：`kol-save-v1` 是可讀的 JSON，`internal/save` 原子讀寫，
  未知的設定鍵原樣保留。原版角色靠離線工具 `tools/import_original_characters.py`
  搬進來，runtime 不解析 `CHARDATA`，原版檔案一律唯讀。
- 使用者決定客棧住宿**預設不離隊**，做成 `faithful_inn_stay` 開關切回原版行為
  （問「存到硬碟嗎」、角色離隊、隊伍空了問是否離開遊戲）。與原版對拍一律開著跑，
  與戰鬥的「忠實 ＋ 可關閉的加速」是同一個原則。
- 對拍抓到一個沒讀出來的行為：答應「全隊」之後原版每處理一位就重畫隊員格
  （`sub_11A93:0x127E5` 的 `sub_13E82`），白框從剛招募的同伴移回隊長。
  第一版 remake 沒有這一步，隊員格差了 474 個像素；補上之後五對只差滑鼠游標。
- 玩家路徑 `town-rest` 11 步，並且**把存檔本身納入驗證**：
  `verify_player_path.py` 多了一個 `save` 區塊，把寫出來的 JSON 讀回來逐欄比對
  （金幣 1009 − 56 ＝ 953、所在地 `HALFWAY HO`、建築 86、城 7 入城點 3 的 (57,8)）。
  存檔因此是玩家路徑的一部分，不是另一條沒人走的支線。
- 對拍 `town-rest` 11 對全過；十九份收據 251 對全過，九份玩家路徑全過，
  37 張展示截圖重生後逐 byte 相同。
- 譯文補 16 筆 KOLDATA 記錄（住宿、治療、受傷招呼語、兩個硬碟問句）。
  字模表 2,366 → 2,368（新增「棧」「碟」），比對碼位集合確認沒有掉字；
  `verify_text_fit.py` 6,725 項全過。中文的價錢句併成一行——原版把數字與量詞拆成
  兩行是因為框只有 24 個原版字格，中文的框放得下，拆開反而讀不順。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續一）：MIRROR 的角色資料頁

- `sub_11A93:0x13D06` 的比較鏈給出建築內十一個圖示的處理位址；MIRROR 是
  `sub_15ABF`，抬頭在 `sub_16239`。整頁的算式讀出來之後（RE-074），
  用原版存檔裡 Theodoric 的 byte 逐項算過一次，與原版畫面相同。
- **匯入工具推翻了自己剛寫下的斷言。** 角色記錄 0x90–0x92 被我寫成日期，
  匯入出來卻是 77／73／78 ＝ `MIN`——那是三字母城鎮代碼。RE-073 的欄位表已改正，
  存檔格式也把沒有證據的 `day`／`month`／`year` 拿掉。
  教訓寫成規則：**沒有證據的欄位不猜也不留佔位**，寫進格式的每一欄都要有出處。
- 對拍抓到三件反組譯讀得到、但沒先想到的事，每一件都是先看到像素差才回去讀碼：
  1. 五個框都有右下投影（`sub_1BA9D` 末尾的 `sub_1B7E9`），少了就差 3,000 多像素。
  2. 抬頭整段轉大寫（`word_2DF54`），KOLDATA 裡是 `Kelden`、`Cliff Guard`。
  3. 繪字是不透明的，連空白都清底——`GOLD CROWNS  :` 的冒號被值的第一個空白蓋掉，
     原版畫面上正是只有那一行沒有冒號。
- 資料頁用的是 8×6 的小字型，不是城鎮對話的 8×8 大字型：抬頭的列距是 6 px，
  用大字型會互相蓋掉。這一條是先照大字型畫、看到抬頭糊成一團才回頭確認的。
- 對拍 `town-mirror` 八對全過，整頁 320×200 只差原版滑鼠游標那 37 個像素。
  玩家路徑 `town-mirror` 10 步；展示截圖多一張 `en-town-sheet.png`。
- 二十份對拍收據 259 對全過，十份玩家路徑全過，38 張展示截圖重生後舊圖逐 byte 相同。
- 中文模式還沒有資料頁的版面：`Session.CJK` 為真時按 MIRROR 只記一筆未實作，
  不開出一頁排錯的畫面。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續二）：讀檔回到遊戲，存檔的迴圈閉合

- `-load` 把隊伍、金幣、傷勢與所在位置放回去。分工寫進 SPEC-006：**角色的靜態資料
  不進存檔**——頭像、七個內部屬性、武器熟練度仍由 typed data 提供，依名字接回去；
  存檔只帶會變的那幾欄。鏡頭是算回來的：夾限取該入城點在世界地圖記錄裡的那一組。
- 第一版把隊伍放到門的格子上就順手設了門口狀態，結果開場停在門面圖，
  `L`／`R` 在門口模式無效——玩家路徑的「相鄰兩步不得相同」當場抓到。
  規則本來就寫過：原版的掃門迴圈接在**移動**後面，放置不算移動（RE-065）。
- `verify_player_path.py` 多一種檢查點 `same_as_except`：指定矩形之外必須逐像素
  相同、之內必須真的不同。`town-load` 用它證明讀檔只換掉金幣那一欄——
  收據記下內 36 個像素不同、外 0 個。這比「兩張圖不一樣」強得多。
- 二十份對拍收據 259 對全過，十一份玩家路徑全過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續三）：商店的商品表與明細面板

- 商品表在建築 payload 的 0x100 起，一筆 16 byte（RE-075）；價錢是
  `(rec[8] & 0x0F) << 8 | rec[9]`，傷害／防護是骰子，骰面查 DGROUP+0xD4 的表。
  八座城 35 棟商店全部解得出來，價錢對得上手冊的武器比較。
- 走到商店這件事本身要先解：城 7 的十個入城點各自把鏡頭夾在一個 16×16 象限裡，
  半途之家那一區沒有商店。改用第 14 名角色 ddduhh（住城 1、3000 金幣），
  用 BFS 在城 1 的格子陣列上算出 37 步的路徑餵給 oracle。
- **清單與明細第一次跑就逐像素相同**，只差原版的滑鼠游標。版面全部是照
  `sub_1AAE8`／`sub_153FF` 推給繪圖 helper 的引數算的，一格都沒用量的。
- 名稱的大小寫是 `sub_1B94F` 的第二個引數：對話走 0（原樣），名稱走 1
  （每個字的第一個字母轉大寫）。第一版把名稱全轉大寫，畫面上就成了 `FUR CAP`。
  改 `render()` 加 `capitalise` 參數時**先把 mode 0 的行為弄壞了**——
  空白原本不動跳脫狀態，我讓它一律歸零。`town-text.json` 的 SHA 立刻變了，
  改回「只有 capitalise 時才動」才回到原值。**改共用解碼器要先有一個會變的雜湊當哨兵。**
- 買下之後那一段（合身與穿戴，記錄 0x33／0x5C）還沒實作。這是對拍時發現的：
  原版買完顯示 `"I'll put this on you."`，而我讀的那一段程式碼裡沒有這一筆記錄——
  往下追才看到背包填完之後還有一整段。remake 目前買完就回清單，並記進未實作清單。
- `tools/import_original_characters.py` 多一個 `--location`，固定起點的存檔因此可以重生；
  玩家路徑 `town-shop` 13 步就是靠它。
- 二十一份對拍收據 265 對全過，十二份玩家路徑全過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續四）：買完之後的合身與穿戴

- 對拍時發現原版買完顯示 `"I'll put this on you."`，而我讀的那一段程式碼裡
  根本沒有那一筆記錄——**是我讀太早就停了**。往下追才看到背包填完之後還有一整段：
  會調整的店先算「這件衣物穿在這名角色身上該有的尺寸」，剛好就直接穿上，
  不合身才問要不要花錢改（RE-075）。
- 尺寸公式與三張 DGROUP 表都讀出來了。ddduhh 的體型 105、毛皮帽的係數 6 × 1，
  算出來是 6，與記錄裡的 6 相同——所以原版就是直接穿上。remake 算出同一個數字。
- 穿上是把背包那一格與 `0x100 + 16 × 部位` 對調（部位 9 換到 0x30）。
  **穿戴槽因此是 0x100–0x19F 的九格加上 0x30。**
- 三張商店畫面（清單、明細、穿上）整頁 320×200 都只差原版的滑鼠游標。
  玩家路徑 `town-shop` 14 步，三個檢查點都與對拍截圖逐 byte 相同。
- 順手把住宿、治療與合身共用的價錢句抽成 `priceLines`，三處走同一段。
- 二十一份對拍收據 267 對全過，十二份玩家路徑全過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續五）：TRAIN 的甲冑調整

- TRAIN 圖示底下其實有兩支：旗標 byte 1 的 bit1 是把身上的甲冑改到合身，
  bit3 是鍛造。這一輪做 bit1（RE-075）。
- ddduhh 買下毛皮帽穿上之後按 TRAIN，原版說 `"You don't need to fit!"`——
  體型 105 算出來的尺寸 6 與記錄裡的 6 相同。remake 算出同一個答案，
  那張畫面整頁只差原版的滑鼠游標。
- **不合身那一條路只有靜態證據**：原版存檔裡沒有尺寸不合的穿戴品，
  正常玩家路徑走不到，`"The armor on your` 與部位名之間有沒有空白也沒對拍過。
  這件事寫進 RE-075 與收據，沒有假裝已經驗過。
- 單元測試補三條：都合身、逐件講完再收費、錢不夠。
  第一版的「錢不夠」測試自己設錯了——身上 10 枚、改一頂帽子 5 枚，本來就付得起。
- 二十一份對拍收據 268 對全過，十二份玩家路徑全過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續六）：AWARD 的訓練

- 訓練沒有自己的函式，整段在建築內部主迴圈 `sub_11A93` 的 `0x1375c`…`0x13ceb`（RE-076）。
  新探針 `tools/ida_probe_koltown_train.py` 從記錄編號反查呼叫點，一次把
  訓練、文字欄與游標三支倒出來。
- 三個閘門依序是：帳面位階要等於實得位階（否則「去競技場」）、金幣要付得起學費、
  技能值要落在這門課的上下限。**學費在選技能之前就扣掉**，後面被擋下不退錢。
- 「要幾點技能」那一格是 `sub_1B5C8(0x2E3A, 2)` 開的文字欄，不是圖示選單：
  打進去的數字會回顯，還要按 Return 才算數，只收 1…5 的一位數。
  第一版直接拿數字當答案，所以永遠畫不出原版那個中間狀態，`nopoints` 差了 1,661 px。
- **欄尾那條橫槓不是字模。** `sub_1B100(1)` 每 5 拍把 `sub_1CA6E` 的參數在 0x00／0xFF
  之間反相，EGA 版 `sub_1D01D` 把那個 byte 寫進顯示記憶體——8 px 寬、1 px 高，會閃爍。
  字模表第 64 格那條底線是 `0xFE`（7 px），拿它當游標會差最右邊那一個像素；
  改成畫一條 8×1 的線之後 `count` 與 `digit` 兩張才歸零。
- 順帶修掉兩個共用元件的差異：`]` 在 UI 文字裡是空白（`How many Offensive` 少了那一格）、
  清單的上下箭頭只有在筆數多於顯示列數時才畫，且下箭頭畫在最後一列。
- 六張畫面（學費、技能清單、攻擊還是防禦、空數字欄、回顯 1、點數不夠）整頁 320×200
  都只差原版的滑鼠游標。ddduhh 的冒險點數是 0，**加技能那一條路走不到**，
  只有靜態證據與單元測試——這件事寫進 RE-076 與計畫，沒有當成驗過。
- 二十二份對拍收據 274 對全過，十三份玩家路徑全過；38 張展示截圖重生後逐 byte 相同。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續七）：野外的時刻、日曆與天候

- 時刻整段在 `sub_297F2` 與 `sub_29987`（RE-077）。日曆、營養、天候、時段門檻
  六張表都在 KOLTOWN 的 DGROUP，新探針 `ida_probe_koltown_wilderness.py` 一次倒出。
- **每步的成本靜態讀不出唯一解**：式子是 `2 × (表[?] + 隊伍移動率)`，而那個索引在
  反組譯裡是編譯器共用的暫存變數，值追不乾淨。先量界限：舊收據 `world-road-run1`
  給了「20 步之內會跨」，第一次量測 `world-clock-run1` 給了「17 步還沒跨」，
  兩者兜不出唯一解。
- 第二次量測改成**左右來回**（次格在同一格內、地形固定、不換格也不觸發地點）：
  第 15 步跨過 DAWN 的門檻。平原 8 ＋ 移動率 20 得 56，是唯一同時滿足
  「14 步不跨、15 步跨」的值——索引就是地形碼。
- 順帶靜態確認一條：**撞到不可通行的地形那一步不耗時間**（`0x2958F` 之後跳過
  整支 `sub_297F2`），這也解釋了第一次量測為什麼偏慢——那一輪有一半的按鍵被擋住。
- 第一次量測踩進伏擊（frame-220，「1 隻 2 級 BINDERAK」），意外把遭遇的觸發點解出來：
  走進 `byte_2FEE4/5` 那一格就進 `sub_2AAA8`。但那一格**在 KOLTOWN.EXE 裡只有讀
  沒有寫**，是誰填的還沒解，寫成 `未知` 沒有硬湊。
- 二十三份對拍收據 278 對全過，十四份玩家路徑全過；38 張展示截圖重生後逐 byte 相同。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續八）：紮營畫面、遭遇的規則，以及一次量測方法的修正

- 換日那一張（CAMPING!）依 `sub_297F2` 實作：天候第 5 組、標籤第一行換成記錄 0xE2、
  地圖視窗畫腳下地形那一張圖，按任何鍵收掉。**日期是收掉之後才推的**——
  原版的順序是「時刻歸位 → 補營養 → 停在畫面等鍵 → 推日期」，所以紮營那一張上面
  印的還是前一天。第一版寫在 `Advance` 裡，對照原版位址之後搬到 `EndCamping`。
- **紮營沒對拍成**。走到換日要 76 步，中間四次時段變換每次都擲遭遇：
  第一輪第 15 步被 1 隻 2 級 BINDERAK 攔下，第二輪同樣第 15 步被 1 隻 1 級 ORC 攔下。
  遭遇畫面上方向鍵無效，之後的按鍵全部空轉。**這是排序問題不是實作問題**：
  要先把戰鬥打通，能打完再走，才拿得到紮營的同狀態畫面。
- 量測方法修正一次：oracle 每一步預設等 120 秒的靜止逾時，而換過天候之後雲層動畫
  讓畫面永遠不靜止——76 步會變成兩個半小時。野外那幾步的逾時改成 8 秒。
  **判準是「這個等待會重複幾次」**：76 次 × 120 秒，值得停下來改。
- 順帶把遭遇整條規則讀出來並抽成 typed data（RE-078，見上一輪 commit）。
  也因此知道遭遇是**時段變換時擲**，不是走到某一格；走到 (55,13) 的是競技場與賭場。
- 戰鬥的進入點是 `sub_1E3A8`（4,389 條指令，另有 62 支直接被呼叫的函式），
  由伏擊、競技場與任務怪物三處進入。已先畫出模組範圍與 89 種戰鬥訊息記錄。
- 二十三份對拍收據 278 對全過，十四份玩家路徑全過。
- 容器狀態：本輪起過三個 dosbox 容器，兩個中途停掉（量測方法修正與順序調整），
  都已 `docker stop` 並確認離開；未動任何共用 image／volume。

## 2026-09-03（續九）：走到戰鬥的門口——任務地點的提示已對拍

- 先找可重現的入口。原版自己的除錯鍵（城鎮地圖按 `c` 直接開一場測試戰鬥）實跑過，
  畫面完全沒反應；隨機伏擊跑了三輪，兩輪被伏擊攔在半路、一輪只擲出治療者。
  最後從 `sub_2B0FF` 的任務旗標讀出可重現的那一條：原版 `CHARDATA` 的第 2 名角色
  Poontang 帶著第 5 個地點 (48,21) 的旗標，選她出城往東八格往北兩格就會觸發。
- 任務地點的記錄整筆解出來（RE-079）：怪物、等級、數量、36 byte 站位，
  以及 0x80 起八列 5-bit 敘述。第 5 個地點是 12 隻 3 級 BINDERAK，
  敘述與原版 frame-862 逐字相同。
- remake 走同一串按鍵停在同一頁：任務框 (96,55)–(296,152) 內除了原版自己畫的滑鼠游標
  之外逐像素相同。框外不宣稱——兩邊走到那一格的步數不同，時刻與天候因此不同。
- 中途會經過換日，CAMPING! 那一張按任何鍵收掉；這條玩家路徑順帶把紮營走過一次。
- 新增 `internal/battle`：任務地點的資料、參戰者陣列（18 × 50 byte 的版面）、
  戰場（64×64 格、視窗 9×9）與一般遭遇的地形生成。**戰鬥本身還沒實作**，
  答應之後目前直接回地圖。
- 二十四份對拍收據 279 對全過，十五份玩家路徑全過；38 張展示截圖重生後逐 byte 相同。
- 容器狀態：全部 `--rm`，本輪起過的 dosbox 容器都已離開；未動任何共用 image／volume。

## 2026-09-03（續十）：答應之後的滿版怪物圖

- 遭遇那一張整頁 320×200 只差原版自己畫的滑鼠游標。過程中修掉三件事：
  文字是 8×6 的小字型不是大字型（行距 6 不是 8）、白框的矩形是
  `sub_1B6DC(3Dh, 9, 0F9h, 27h)` 而不是黑底往外擴一格、外面那一圈是**青色框**
  加黑底加黑投影（`sub_1BA9D(0Eh, 2, 131h, 9Ch, 3)`），不是單純把那一塊塗黑。
- 怪物照 DGROUP+0x13B0 的十二個位置畫，畫的是 MONPICS 的同一張圖——原版是先畫一隻
  再把那一塊複製到其餘位置，結果相同。
- 二十四份對拍收據 280 對全過，十五份玩家路徑全過。
- 容器狀態：全部 `--rm`，本輪沒有殘留容器；未動任何共用 image／volume。

## 2026-09-03（續）：戰場、視線與戰鬥回合

- 戰場視窗改回 9×9。先前把最下面 32 px 當成訊息區是誤判：那兩列是視線之外的
  黑磚（`sub_223C0:0x22597` 把沒看到的格子填 0x4A）。判準來自 `sub_223C0`
  兩層迴圈都以 9 為界，以及畫面位移 `0x280 × 列 + 0x153 + 2 × 欄`。
- 戰場格子陣列在執行期是 RLE 解出來的轉置（`sub_2B3D1:0x2B436`），繪製又照
  `(Y << 6) + X` 取回去，兩次相反抵銷——所以抽取器輸出的順序不用改。
- 回合結構先寫成「先攻排一次、逐一行動」，讀完 `0x1FD21` 之後修正為
  **規劃與執行兩段**：規劃依洞察排（鍵值小的先決定），整輪決定完問
  「Begin next round?」，YES 才進執行段，執行段依快捷重排、鍵值大的先動。
  正文（RE-080、SPEC-008）已改寫成兩段結構，Go 的 `InitiativeKey` 改名 `PlanKey`。
- 攻擊結算的數字鏈這一輪走完了（`sub_26AC2`、`sub_24FD3`、`sub_26D76`、
  `sub_277F4` 與六張 DGROUP 表），但 remake 還不能實作：缺角色裝備欄的
  typed data（傷害碼、重量、耐久、旗標），敵人 AI 也還沒讀。
- 探針輸出：`combat-view/fieldraw/los/blit/attack/hit/tables/tables2/bar/cmdbar/iconbars.json`。

## 2026-09-03（續二）：攻擊結算、角色裝備與循路時刻的紅燈

- 近戰結算的數字鏈實作完成（SPEC-009）：攻擊值、防禦值、命中判定、傷害骰、
  命中部位與部位血量，七張 DGROUP 表原樣抄進 `internal/battle/attack.go`，
  單元測試逐條驗算。這一份沒有同狀態收據——原版每一步都擲骰，亂數狀態不可重建。
- 角色檔上穿的東西以前沒有讀出來（`Worn` 只有商店買的那一份）。抽取器補上穿戴欄
  十格與怪物的技能／武器／護甲，載入時照角色檔補進去。
- 重跑全部截圖之後發現一個**既有的紅燈**：道路那一份的「循路七步」對不上。
  原因不是這一輪的改動，而是野外時刻上線之後沒有重生那幾張圖——remake 循路一步
  扣 52，原版扣 18…20。兩輪原版量測都在第 9 次循路翻時段（`world-road-clock`、
  `world-road-clock2`）。修法要先追 `sub_29004` 的 ROAD 分支，不是硬塞常數。
- 探針另外倒出 `combat-tables2/bar/cmdbar/iconbars/skill/world-clock/clock-owners.json`。

## 2026-09-03（續三）：循路時刻的紅燈修掉

- `world.RoadStepCost = 20`：循路一步不照一般步伐算。兩輪原版量測都在第 9 次循路
  翻時段，往回推 18…20；取 20 之後 8 步 832、9 步 852 剛好卡在門檻 3367 兩側（差一）。
  單元測試把那一組量測釘住。
- 重生道路、時刻與任務三份截圖之後：對拍 24 份收據 283 對全過，玩家路徑 15 條全過，
  `go vet`／`go test ./...` 全過。
- 機制仍未追出：`sub_297F2` 只有一處扣時刻（`2 × ([0x12B8 + word_39A54] + word_2FA2C)`），
  而循路那一條路的兩個全域湊不出 20。常數旁邊註明了這一點，不假裝已經解出來。

## 2026-09-03（續四）：git 作者信箱改寫與敵人那一手

- 接手先查 git 身分：`git config user.email` 已是 `wicanr2@gmail.com`，但
  `git log --format=%ae | sort -u` 顯示歷史裡 144 個 commit 有 64 個掛
  `cy.wang@coretronic.com`（最後一個 2026-09-01）。經使用者授權改寫：
  `git filter-branch --env-filter` 把 author 與 committer 都換成
  `wicanr2@gmail.com`，改寫後 tree hash 與改寫前相同（`86b725d9`）、
  `git diff` 為空、commit 數不變。無 tag 也無 release，不需重建。
  `--force-with-lease` 推上去，遠端 HEAD `b4a5531`，repo 仍為 private。
  舊 HEAD 留在本地 `backup-pre-email-rewrite`（`29d2ecd`）與 `refs/original/`。
- 敵人那一手（`sub_2374C`，833 條）逐條讀完，寫成 RE-081。新探針
  `tools/ida_probe_koltown_enemy_ai.py`：根函式全倒、十七支被呼叫者逐條、
  呼叫端視窗、十二塊 DGROUP 位元組與七個環的位移陣列。
- 決策的回傳是三個全域（`word_31122`／`word_30A78`／`word_30916`），由呼叫端寫進
  參戰者記錄的 +0x25／+0x23／+0x24；方向自己寫。搜尋單位是以行動者為中心的
  七個同心方環（格數 8r）。智力擲骰（`智力 × 等級 ÷ 3 > 開頭那一個 rand(100)`）
  決定要不要做反制掃描。加權輪盤用 `權重^等級 ÷ 2^(等級−1)`，等級越高越偏向
  高權重選項。
- `[0x10D8]` 的九格方向表與 RE-080 由 `0x0B76` 反推的方向定義獨立吻合，
  等於替那一份加了第二個來源。
- 沒解出的部分寫進 RE-081 末段，不假裝完整：`byte_3496A`／`3496B`／`3496C`、
  `[0x6AD2]`、`[0x1A9A]` 三張基礎權重表在靜態影像裡是未初始化區，全庫掃過
  只有讀沒有具名 writer，要靠執行期取樣才能定值。

### 勘誤

- RE-080 末段原本把 `sub_2374C` 與「缺角色裝備欄 typed data」一起列在還沒走完的
  清單裡。裝備 typed data 已於同日 `1ff7031` 完成，敵人 AI 見 RE-081；該段已改寫
  成現況。
- `WORKLIST.md` 的「下一個唯一工作」原本停在「建築內的後續對話」，而同一份的
  checklist 早已把那一項標成完成（RE-067）。已改寫成戰場上的指令。

## 2026-09-03（續五）：玩家那一手與 SPEC-010

- `sub_228B7`（535 條）、`sub_24131`（699 條）、主指令列選完之後的分派段與三列
  子選單逐條讀完，寫成 RE-082 與 SPEC-010。新探針
  `tools/ida_probe_koltown_battle_commands.py`。
- 主指令列十五個圖示各自的條件都是角色記錄的武器欄、+0x190、0x200 起的物品欄
  與參戰者記錄的手臂／腿／+0x1D／+0x1E 組出來的。腳下那一格用環 0 的第 9 項取
  ——環 0 的位移陣列除了八個方向還多一項 `(0, 0)`。
- 選目標格兩個 mode 的禁區不同：近戰不准停在中心、射擊不准落在 2…6 的內圈。
  游標停在有人的格上會換右側資訊框，而**敵人的細節要智力檢定才顯示**，
  檢定用的是 RE-081 同一支 `sub_252CB`。滑鼠也能直接指格。
- `[0x0B6C]` 與 `[0x0B76]` 互為反表，八個方向逐一往返一致——這是 RE-080 方向
  定義的第三個獨立來源（第二個是 RE-081 的 `[0x10D8]`）。
- 移動方式的防禦選項規則（0x3C 給 2、其餘 0）與 RE-081 敵人 AI 輪盤那一段
  獨立吻合，兩邊指的是同三個候選 0x3C／0x3D／0x3E。
- 找到兩段死碼（`0x247E7` 清場、`0x2483F` 重畫），都被前一條無條件跳躍跳過，
  看起來是原版留下的除錯鍵，不進 remake。

### 勘誤

- RE-080 的「圖示列怎麼組」把兩列子選單標反了：`0x1F6BA`（0x2A…0x2E，查
  `[0x0B42]`）是**武器**攻擊式，`0x1F940`（0x2A、0x37…0x3A，選項 = 圖號 − 0x36）
  是**徒手**。三個判準：分派表把 0x21 送到前者、0x29 送到後者；徒手的裝備欄查
  `[0x0FD2 + 選項]` 只有六筆，值域必須 ≤ 5；而 `[0x0B42]` 的 0…8 正好是
  `[0x110A + 14 × 攻擊選項]` 那張 128-byte 表的九列。徒手的四個圖示也對上
  RE-080 已證實的四句徒手訊息 0x128…0x12B。正文已改，`internal/battle` 不受
  影響——那一層以攻擊選項為輸入，沒有硬編哪一列屬於哪一種攻擊。

## 2026-09-03（續六）：指令列與選目標的規則層

- `internal/battle/commands.go`：`CommandBar` 照 `sub_228B7` 的順序組出主指令列，
  十五格的條件逐一實作；四列子選單的選項換算（武器查 `[0x0B42]`、徒手是圖號減
  0x36、防禦是圖號減 0x2F 帶 0x36 補回 6、移動的防禦選項只有走路那一格是 2）
  與兩張方向反表也在同一支。
- `internal/battle/cursor.go`：`sub_24131` 兩個 mode 的游標。近戰起點是面前那一格、
  夾在 3…5 且不准停在中心；射擊起點是上次記住的目標、夾在 0…8，內圈用方向鍵移進去
  會把位移乘三跳到外面，滑鼠點進去才還原成上一格。
- `Weapon` 補 `Code`（+0x0C 原值，指令列要它的 bit 0，而 Count／Sides 取不回來）；
  `Actor` 補 `Command`／`Facing`／`Flag1D`／`Flag1E`／`Flag03`；`Member` 補
  `HasWeapon`／`HasSecondItem`／`PackCount`。後三對的語意未定，命名沿用位移。
- 二十個單元測試：指令列九組邊界（空手、原版那一列、雙臂全廢、雙手武器缺一臂、
  三種射擊種類、投擲轉射擊、背包、地上東西、Flag1E、無腿）、四列選項換算、
  移動方式的兩組、攻擊收尾與方向反表往返。`go vet`／`go test ./...` 全過。
- 對拍還沒做——指令列的畫面與輸入迴圈還沒接，frame-887／frame-899 的重新對拍
  留給下一輪。反推出原版那一列的最小武器條件只證明「條件能產生那一列」，
  不證明 parity，測試名稱與註解都寫明了這一點。

### 勘誤

- RE-082 與 SPEC-010 初稿把射擊游標內圈的按鍵行為寫反了。原版 `0x24519` 比的是
  「這一次的按鍵是不是 0xFF（滑鼠）」：**方向鍵**移進內圈會把位移乘三跳到外圈
  （乘完不再夾），**滑鼠**點進內圈才還原成上一格。兩份文件已改。
- 另外補記一筆原版的 dead store：射擊游標的起點算完之後有一段 0…8 的檢查
  （`0x2428B`…`0x242B3`），但它唯一的動作是清掉一個馬上被覆寫的區域變數，
  所以越界的起點照樣留在游標上。實作照抄這個行為，另外把「有沒有可用起點」
  當成回傳值報出來。

## 2026-09-03（續七）：指令列接上畫面與輸入，主指令列對拍過了

- `internal/battle/planning.go`：規劃段的選單狀態機（主指令列 → 選目標格 → 子選單
  → 寫回），取消逐層退回。`internal/ui/battle.go` 畫指令列（KOLMPICS 圖號是命令碼
  加三），`cmd/kol/main.go` 的 `updateBattle` 與 `applyBattleToken` 接鍵盤與腳本。
- 十一個規劃段的單元測試：武器攻擊整條、攻擊選項 1 跳過防禦式、移動、射擊記住目標、
  兩層取消、兩個選項都 0 變休息、直接收尾那幾格、未實作的圖示、白框繞回。
- 主指令列已對拍（RE-083）：圖示列整區 (32, 170) 216×22 的 4,752 個像素與
  frame-886 **零差異**。用的是原版角色檔裡 Poontang 真實的 Long Spear
  （種類 0x9E、傷害碼 bit 0 為 1），不是反推出來的最小條件。
- 對拍基準是 frame-886 不是 887。`sub_11722` 迴圈畫完每一格之後單獨把第 0 格框成
  白色（`0x1178D`），所以剛進戰場白框一定在最左邊；原版那一輪的 checkpoint 記的是
  **按鍵前**的穩定畫面，frame-887、888 已經各被 `>` 移過一格。
- 收據 24 份 282 對全過（`plan-world-quest` 從 3 對加到 4 對）。

### 修掉的回歸

`tools/extract_character_roster.py` 把 `equipment` 放進了 `sheet` 這一層，而
`internal/roster` 的 `Slot.Equipment` 是從槽位那一層讀。兩邊對不上，所以 `Worn`
一直是十格全 nil——指令列只組得出空手的四格，SPEC-009 的近戰結算也拿不到武器。
抽取器已改成放在槽位層並重生 typed data。

**單元測試結構上抓不到這一類**：規則層的測試自己餵 `Member`，走不到 JSON。
抓到它的是拿原版畫格對拍——第一次比出來七格裡有六格不對，才往回查到欄位位置。

### 勘誤

- `CONTEXT.md` 原本寫「二十四份收據 283 對」，而同一段列出的明細加起來是 281，
  文件自己前後矛盾。`docs/verification/remake-parity/README.md` 的表格也把
  `world-quest.json` 記成 2 對（實際 3）、總數記成 280。兩份都已改成對齊實際收據
  的 282 對。
- RE-080 的指令列那一段原本拿 frame-887 當七格的佐證。圖示內容 886…888 都相同，
  所以那句話沒有錯，但拿它當白框的基準會錯；已註明基準是 frame-886。

### 觀察（沒有處理）

`docs/verification/remake-parity/` 底下每一份收據都有兩個副本：`X.json` 與
`receipt-X.json`，內容相同（`world-quest` 這一份原本差了兩對，已同步）。
README 引用的是不帶前綴的那一套。要收斂成一處要先確認哪一套是正式的，
這一輪只讓兩者一致，沒有動命名。

## 2026-09-03（續八）：執行段接上，戰鬥會造成傷害了

- `internal/battle/execution.go`：執行段的鍵值（快捷扣傷勢、負重 ÷ 10、
  武器速度 × 指令時間成本 ÷ 10，再加擲骰、夾 0），排序與一手的分派。
  `[0x10B2]` 的時間成本表原樣抄進來。
- 攻擊那一手接上 SPEC-009 的結算。目標的取法是 `sub_26D76` 開頭的
  `sub_24BB6(0, 方向, 8)`——遮罩 8 只看 +0x1E，**不分敵我**，面前站著自己人一樣
  打得到。面前沒人就揮空（徒手記錄 0x11F、其餘 0x120）。
- `Answer(true)` 現在會先跑完整個執行段再排下一回合的規劃；隊伍全倒就結束。
- 找到兩處原版 dead code：移動且攻擊選項為 0 時算了 `鍵值 ÷ 2` 卻沒有寫回
  （`0x1FF2F`），對照射擊那一條（`0x1FEF2`）就有寫回；選目標格的起點檢查
  （RE-082）是同一類。remake 照原版的行為，不實作那一條折扣。
- 十個執行段的單元測試：時間成本表的四種取法、鍵值的四項扣除與夾 0 的時機、
  射擊的減半條件、順序、面前那一格找人（含不分敵我與 Flag1E）、六種分派、
  跳過倒下與空位。`go vet`／`go test ./...` 全過，對拍 4 對仍全過。
- SPEC-008 的執行順序由 DRAFT 升 READY，末段改寫成現況：移動點數、敵人 AI 的
  三張權重表、戰利品與怯戰檢定的兩個開關仍未解。

### 這一輪沒有做的

執行段的訊息**沒有逐句顯示**。原版每一手印一句再等玩家按鍵（`sub_1B100(5)`），
remake 一次跑完把結果收在 `Execution.Log`。結算本身照原版做，差的是呈現。
移動、撿、丟與背包那幾手在 `Act` 裡標成 `Unimplemented`，不假裝做過。

## 2026-09-03（續九）：移動那一手，順便推翻一個假設

- `sub_26384`（207 條）讀完，寫成 RE-085 並實作。**移動一次只走一格**：整支函式
  只把座標加一次環 0 的位移，沒有點數也沒有步數迴圈。
- 走不動的四種情形都照原版判：目的地擋路（位元圖 `[0x0FC8]`，與擋視線的
  `[0x10CE]` 是**不同的兩張表**，例如磚 0 擋視線但不擋路）、+0x1E 非零時腳下踩到
  0x0C、目的地站著同類的人（遮罩 8 與 4：沒有 +0x1E 的擋沒有的、有的擋有的）、
  新座標超出 4…0x3B。
- 走不動印記錄 0x10A、走了一步印 0x10B。
- 十六個新測試（移動方式換算、走一格、四個界限、三種誰擋誰、兩張位元圖的差異）。
  `go vet`／`go test ./...` 全過。

### 推翻的假設

RE-080 末段、SPEC-008 與 WORKLIST 都把「移動指令 0x27 的點數規則」列為待解，
一路寫了好幾輪。實際上**沒有這個規則**——把移動方式（圖號 0x3C…0x40、0x35）
當成了距離。移動方式只有兩個作用：方式 3…5 會把 +0x1E 設起來，以及在執行段的
鍵值裡當時間成本（而時間成本對移動是固定的 6，不看方式）。三份文件已改。

### +0x1E 這個旗標

把 RE-082 與 RE-085 合起來看，它的行為像騎乘：踩到磚 0x0C 才解得開後三種移動
方式、選了就設起來、設起來之後大部分擋路磚穿得過去、指令列從「徒手／移動」換成
「0x28／0x3F」。**但原版沒有任何文字或圖形證據支持這個命名**，所以 remake 照行為
實作，欄位名沿用位移（`Flag1E`），SPEC-008 把它列在「還沒解出來的」。

## 2026-09-04：訊息的代換碼

- `sub_1A8FD`（178 條）與 `sub_265F3`（226 條）讀完，寫成 RE-086 並實作
  `internal/battle/message.go`。
- 記號只有四種：`0x0D` 結束、`|` 代換、`]` 換成一個空白、`@` 換行、`~` 後面三個
  字元是座標或顏色的參數。
- 十一個代換碼：`|A`／`|C` 是目標的所有格與受格、`|B`／`|D` 是行動者的、
  `|E` 是行動者手上武器的名字、`|F`／`|G` 是目標與行動者的名字（敵人前面補
  `KOLDATA[0x140]` 的 `The `）、`|H` 是物品、`|I` 是角色記錄 +0x190 那一格、
  `|N` 是數字、`|S` 是暫存區的名稱。
- 每一手開頭先呼叫 `sub_2685A`（清、印行動者名字、補空白），所以畫面上那一整句是
  「<行動者的名字> <記錄展開>」——`MessageContext.Line` 就是這一條。
- 代換之前一律補一個空白。原版由 `word_345E0` 控制，那個旗標的語意還沒定，但
  KOLDATA 裡每一句帶代換碼的訊息都需要它才讀得通（`kicks|B leg` → `kicks his leg`），
  所以 remake 一律補。第一版漏了這個空白，是測試對照原文時抓到的。

### 順帶收窄了一個未知欄位

`|I` 讀的是角色記錄 +0x190，而 KOLDATA 0x114 是 `puts|B|I in|B belt`。RE-082 只知道
+0x190 會影響指令列上 0x16／0x24／0x25 三格出不出現；配上這一句，它是**腰帶上那一格
裝備**的可能性很高。仍標 `強推論`——沒有第二個獨立來源。

## 2026-09-04（續一）：訊息框的版面，訊息真的畫出來了

- `sub_1C1A4`／`sub_1C1BF`／`sub_1C2DA`／`sub_1C311`／`sub_1103F` 與 `sub_1DE84`
  的開頭讀完，寫成 RE-087 並實作。
- 訊息不是邊組邊畫：`sub_1C1A4` 先進緩衝模式，之後所有字元寫進 `[0x1B60]`，
  `sub_1C1BF` 才一次斷行畫出來。所以行動者的名字與訊息本體是同一段文字。
- 版面固定：byte 欄 3（`word_2DF14 << 3` ＝ 像素 24）、掃描列 0xA8、行高 6
  （小字型 8×6）。斷行從行寬處往回找空白，第一行 30 個字元、之後每行 33 個
  ——兩者不對稱，原樣照抄。
- 每畫一個字元前先清一塊 8×8（`sub_1C40D`），清的高度比行高多兩列，而且是逐格清，
  所以沒有字的格子保留底下的畫面——戰鬥訊息就這樣蓋在圖示列那一塊上。
- 執行段改成逐手停下（`State.Step`），Answer(true) 之後進 `PromptExecuting`。
- 實跑驗證：`workplace/specs/exec-msg.txt` 走完「進戰場 → 選休息 → Begin next
  round? → 執行」，再用 `read_text_from_frame.py` 從 (24, 168) 逐格比對原版字模，
  讀回 `Poontang is resting!`。字模、起點與行高都對得上。

### 這不是 parity

`world-quest-run1` 那一輪 oracle 的按鍵序列沒有走進執行段（checkpoint 108…123
都在規劃與戰場之間打轉），所以**沒有原版畫格可以比**。上面那一條只證明 remake
自己前後一致。要 parity 得另錄一輪走到執行段的 oracle，已列進 WORKLIST。

### 過程中修掉的兩個問題

- 第一版把 `bank.Lines` 拿來取訊息，而 `Lines` 會 `TrimSpace`，`The]` 的尾空白
  被吃掉，畫面上變成 `TheBINDERAK`。改用 `RecordPadded`（`]` 保留成空白），
  並讓 `name()` 一定補上那個空白。
- 敵人那一手的 `+0x25` 從來沒被填過（AI 還沒實作），原本會落到「指令 0 → 記錄
  0x117 does nothing」——畫面上看起來像原版選了什麼都不做。改成標
  `Unimplemented`；`battleStep` 遇到沒有訊息的手直接往下走，不讓玩家停在空白畫面。

## 2026-09-04（續二）：攻擊的訊息鏈

- `sub_26D76` 印訊息那三段與 `sub_277F4`（546 條）讀完，寫成 RE-088 並實作
  `AttackLine`／`DamageLine`。
- 一次攻擊分兩句、各等一次鍵：第一句是動作加結果，第二句是傷害。
  武器的動作句是 `[0x12F + 攻擊選項]`（實際只用得到 whacks／chops／thrusts／
  slices 四句，因為選項值域是 {1,2,5,8}），徒手是 `[0x127 + 選項]` 四句，
  射擊依武器種類 > 0x7F 選 boulder 或 missle。
- 落空分三段（幅度 > 0x1E／> 0x0A／其餘）、被擋查 `[0x11B2]` 的招架碼、
  命中印「who cannot parry the blow!」。
- 傷害那一句：傷害 0 印 little damage；倒下印 die/collapse/shatter；手腳廢印
  limb useless（並清掉目標的指令碼，腿廢還會清 +0x03）；頭胸廢依敵我印
  killing it／unconscious；其餘依 0x10／0x18 兩個門檻印 serious／terrible／
  grievous wound。
- 實跑驗證：`workplace/specs/exec-attack.txt` 走到武器攻擊、面前沒人，
  讀回 `Poontang whips his Long Spear through the air.`——兩行，第一行 29 個
  字元斷在空白，與 RE-087 的斷行規則一致。

### 勘誤

- RE-080 把招架訊息記成「0x14F…0x158」。實際上招架碼查 `[0x11B2 + 11 × 目標防禦
  選項 + 攻擊選項]`，**訊息基底是 0x15F**（0x15F…0x168 十句 does not attempt to
  parry／avoids the blow by covering up…）；0x14F 那一段是傷害那一句的 die。
  另外原版還有一條修正：招架碼是 4 而且目標體型減行動者體型大於 100 時改成 2
  （`0x27417`）——`internal/battle/attack.go` 已補上。

### 沒有做的

攻擊訊息只實跑過「面前沒人」那一條。動作句與傷害那一句要讓隊員走到敵人旁邊才驗得到
（任務地點 5 的隊員與怪物差十一格）。`0x27B77` 之後的擊退與昏迷還沒讀；倒下那一句
的 collapse／shatter 兩個分支條件語意未定，實作一律走 die。

## 2026-09-04（續三）：打死了第一隻怪物

- 這一輪沒有新的反組譯，只把 RE-085…RE-088 的實作實際走一遍。
- `workplace/specs/exec-hit.txt`：進戰場 → 往下移動四步 → 攻擊。任務地點 5 的
  隊員站 (14,41)，最近那一隻 BINDERAK 站 (19,41)，所以四步就貼得上去
  （先前寫成「差十一格」是看錯了怪物清單的第一筆）。
- 讀回兩句，逐字都對：

```
Poontang whacks his Long Spear at The BINDERAK who cannot parry the blow!
The blow hits The BINDERAK's chest killing it instantly!
```

- 這一次驗到的東西：移動四步都成功（RE-085 的四道判斷沒有誤擋）、選目標格的起點
  跟著方向走、攻擊式選項 1 直接收尾（RE-082）、動作句 `[0x12F + 1]`、`|B`／`|E`／
  `|F` 三個代換、命中的 0x13E、傷害那一句的 0x140／0x141／0x143／[0x144+1]／0x151，
  以及兩句的斷行（第一行剛好 30 個字元）。
- 句首的「The blow」與 `|F` 展開的「The BINDERAK」都來自 0x140，兩個 The 是原版
  行為，不是重複代換。

### 仍然不是 parity

執行段的畫面**一張都還沒對過**。`world-quest-run1` 那一輪 oracle 的按鍵序列停在
規劃與戰場之間，沒有原版畫格可比。要對得照這一輪走過的路徑另錄一輪
（`tools/run_character_ui_oracle.sh`），已列為下一個唯一工作。

## 2026-09-04（續四）：授權條款

- 本專案先前**沒有 `LICENSE` 檔**，README 也沒有授權段——等於未授權。已依
  `rulebook/85` 的定案（2026-08-28 裁定、2026-09-03 定名）建立 RRSAL-1.0。
- `LICENSE` 由 `~/.claude/knowledge-base/retro/rrsal-1.0.md` 的範本整份複製，
  只換占位符；條款本文十三條一字未改（改字就是新版號）。中英雙語 327 行。
- 灰色地帶（第 2 條）照本 repo 實際有什麼逐項寫：`assets/original-derived-png/`
  的重生 PNG、`docs/verification/` 的原版執行畫面截圖與收據、`docs/i18n/` 裡作為
  對照保留的原版英文字串、`docs/manual/` 的中文說明書譯文，以及 `docs/re/`／
  `docs/spec/` 中由原版量測得出的位址、偏移、座標與數值表。
- 第三方：Noto Sans CJK TC（SIL OFL 1.1，中文點陣字模由它烘製）與 Ebitengine
  及其相依套件。
- 原權利人寫成「Origin Systems 及其權利繼受人」。Origin 1992 年被 Electronic Arts
  併購，但這一款的權利現況沒有查證，所以不點名 EA。
- README 加了「授權、致謝與聲明」段，明寫**這不是開源授權**（source-available）、
  非商業免費、實況與平台分潤明示允許、商業要洽談，以及授權範圍不含原版素材。

### 順帶更正

README 的現況段還寫著「訓練、戰鬥與任務都還沒實作」。訓練在 RE-076 就做完了，
戰鬥這幾輪也已經走得完一輪並打得死怪物。已改寫成現況。

### 還沒做

`rulebook/85` 要求授權出現在四個地方，目前只有兩個（repo `LICENSE`、README）。
另外兩個是「每一個發行包裡」與「發行根目錄」——`dist-all/` 還沒建立，
打包腳本要記得複製，已列進 WORKLIST 的 M6。

## 2026-09-04：執行段的原版畫面與訊息框 parity

### 做了什麼

錄了新的原版 oracle 情境 `world-quest-execute`（`tools/dosbox_character_ui_oracle.py`，
969 個畫格，10m36s）。前綴沿用 `world-quest-site` 走到第 5 個任務地點進戰場，
接著在指令列上往右按到最後一格、按 Return、答應「Begin next round?」，
再連按二十次空白鍵推完執行段的訊息。收據與畫格在
`docs/verification/character-ui/world-quest-execute-run1/`，已登進 MANIFEST
（43 個 run、6,515 個畫格）。

用 `tools/read_text_from_frame.py` 以原版字模逐格比對，從畫格讀回兩句原文：
frame-966/967 的 `THE RUFFIAN MOVES!` 與 frame-968 的 `POONTANG DOES NOTHING.`。
量到五件事寫成 RE-089，remake 四項照著修：整段轉大寫、白底黑字、訊息區整條鋪成
白底、移動成功等兩次鍵。

### 兩個沒收掉的差異

- **訊息區那張底圖沒認出來**。原版 (8,164) 起 296×32 是白 5,488 + 亮黃 2,900 +
  黑 864；remake 鋪純白是白 9,108 + 黑 364。主色一致，黃色那一片圖案不明——
  不是 KOLMPICS 圖 2（那張是紅底，各自取主色當背景比形狀也只對得上
  5,843/9,472）。
- **敵人的名字對不上**。畫面上動的叫 RUFFIAN（怪物 15），任務地點 5 的記錄寫的是
  BINDERAK（怪物 19）。`|F`／`|G` 印的是 `sub_1A8FD(0, 0xB967)`，還沒追出誰寫那個
  緩衝區。remake 目前用任務記錄裡的名字。

兩項都寫進 SPEC-008 的已知差異與 WORKLIST，不當成做完。

### 為什麼沒有逐手對拍

原版的敵人會動（`sub_2374C` 的 AI），remake 的敵人那一手還沒實作（缺三張基礎權重表
的執行期值），兩邊的按鍵數與訊息內容對不起來。這一輪要的是**訊息框的版面**，
那與誰動的無關；逐手的內容要等敵人 AI 做出來才有得比。

## 2026-09-04（續二）：敵人是誰、訊息區的底圖是什麼

RE-089 掛著的兩個差異都收掉了，寫成 RE-090。

### 敵人的名字沒有對不上

新探針 `tools/ida_probe_koltown_monster_name.py` 追 `0xB967`：`sub_2687F` 對敵人
固定印 `KOLDATA[0x140]`（`The]`）加上 DGROUP+0xB967 的第 0 筆——**不帶索引**，
一場戰鬥裡所有敵人共用同一份名字。那份名字由 `sub_1A4FE` 從四個地方之一讀進
DGROUP+0xB966 的共用緩衝，其中任務那一支（`sub_2B0FF`）的怪物編號取自任務記錄的
第 0 個 byte，旗標位址與除數表跟 `quest-sites.json` 的 `quest_byte`／
`quest_divisor`／`payload`／`monster`／`count` 五個欄位完全對上。抽取器解對了。

那為什麼畫面上是 RUFFIAN？**那一輪根本沒走到任務地點**。frame-958 寫著
`YOU HAVE BEEN AMBUSHED BY 1 LEVEL 1 RUFFIAN!`——路上被隨機遭遇攔下來了。
收據的畫格編號也印證：從「往東第 43 步」起畫面就在 frame-961／962 之間反覆，
後面三十幾次方向鍵一步都沒走。

> 教訓寫成規則：**畫面上的兩個值不一致時，先確認兩邊是不是同一個狀態**。
> RE-089 直接把「顯示的名字」和「任務記錄的名字」對比，跳過了「這一場是不是任務
> 戰鬥」這個前提——而收據裡「畫格不再變化」的訊號當時就在眼前。

### 訊息區的底圖

是 **MISCPICS 的圖 1**：一張 296×32 的卷軸，白紙面配黃色的捲邊與軸心。之前比
KOLMPICS 圖 2 只對得上 0.7%，換成 MISCPICS 圖 1 是 94.9%（差的 486 個像素是寫在
紙面上的字）。`internal/battle/message.go` 加 `MessageScroll`，`internal/ui/battle.go`
改成貼那張圖而不是鋪純白。

### 對拍

`plan-battle-exec.json`／`receipt-battle-exec.json`：remake 的 step-104 對原版的
frame-966，訊息區 (8,164) 296×32，允許差異區之外**零像素差**。允許區只有文字那
一行（144×6）——原版動的是 RUFFIAN、remake 動的是 Poontang，內容本來就不同。
收據總數變成二十五份 283 對。

### 順帶讀出來的（強推論，還沒做）

野外隨機遭遇 `sub_2A4E5` 的挑怪骨架：兩道百分比擲骰決定要不要遇上，然後最多試
32 次 `rand(32) − 1`，每次要同時通過地形與區域兩道遮罩。四張表（`[0x1312]`、
`[0x1332]`、`[0x1CF4]`、`[0x133E]`）與 `sub_2A337` 的語意還沒證實，所以整節標
強推論，列進 WORKLIST 而不是 spec。

## 2026-09-04（續三）：三個 UI 文字庫全部翻完

### 做了什麼

`docs/i18n/ui-zh-Hant.json`：KOLDATA 395 筆、BOOKDATA 73 筆、ENDTXT 18 筆，
共 486 筆全部翻成繁體中文。原本只有 41 筆（散在 `town-zh-Hant.json` 的 `ui` 段），
這一輪補完 445 筆並把那 41 筆搬過來——狀態只放一處。

新增 `tools/verify_ui_translation.py`。判準三條：代換碼 `|A`…`|S` 與掃描列
`~Vxx` 的集合必須一致；欄 `~Hxx`／`~Cxx` 與換行 `@` 可以少不可以多（中文比英文短，
兩行常併一行）；每一行的實際像素寬度不得超過畫面寬。486 筆全過，並用一筆
故意加長的譯文做過正對照，確認檢查器真的會失敗。

字模表重烘：2,368 → 2,434 字，UI 與城鎮兩份譯文的缺字都是 0。
`verify_text_fit.py` 6,685 項全過。既有的中文展示圖重生後逐 byte 相同。

### 翻譯的三個決定

- **一筆記錄對一筆譯文。** 戰鬥訊息在原版是靠記錄號串接的碎片（RE-088），
  中文語序不同，但譯成整句模板會讓譯文與原版記錄失去一一對應、之後對不了帳。
  所以每一筆譯成「接得上去的片段」，例如 `|B|E at|F` → `|B的|E，攻向|F`。
- **人名不翻。** 工作人員名單（BOOKDATA 記錄 0…7）是署名，不是內容。
- **KOLDATA 320（`The]`）譯成空字串。** 它同時當敵人名字的前綴與 `The blow`／
  `The spell` 的冠詞，中文兩處都不需要。`377`（` has`）同理。刻意留空的記錄在
  譯文檔裡都寫了 `note`，驗證器據此和漏譯分開。

### 修掉一個真 bug

譯文路徑**不剝版面碼**——`uiRecord` 直接把譯文回傳，所以譯文裡的 `~H0D` 會被當成
文字印在畫面上。既有 41 筆躲過是因為它們把版面碼整個拿掉了，而那又和原文用
`@`、譯文用 `\n` 兩套語法並存的問題疊在一起。

修法是統一：譯文與原文用同一套語法，`uiLines`／`uiRecord`／`uiRaw` 三個都走
跟原文相同的 `strip`／`splitUILines` 處理。這樣譯文與原文可互換，`ParseSegments`
那種需要座標的記錄（角色資料頁的 KOLDATA 115）也能直接用譯文。

### 還沒接上

BOOKDATA 與 ENDTXT 的譯文**還沒有畫面在讀**：書本目前走 `newchar-zh-Hant.json`
（以英文原文為鍵），結局畫面則還沒實作。譯文備妥了，接上就能用。

UI 譯文的**框內**版面也還沒驗證。`verify_ui_translation.py` 的行寬檢查只擋得住
「畫到畫面外」，擋不住「畫到框外」——每一筆記錄畫在哪個框，要等它接上畫面才知道。

### 順帶

術語表補了 23 列：13 列職業與部族從「待查」改成「暫譯」並填上遊戲畫面要用的譯名
（畫面上沒辦法顯示「待查」），另外新增港務長、丹德爾、皮爾達、塞加利昂、富匡公爵
與 25 個社會階級。

## 2026-09-04（續四）：自評分數、知識庫主題文件與過期斷言清理

### 自評分數

WORKLIST 新增「自評分數」一節：**54 / 100**。權重是判斷（各系統在原版裡佔玩家時間
的份量），完成度盡量指到 spec 狀態、對拍收據或程式位置，指不到的寫「無」。算式
逐項列出來，好讓它可以被反駁而不是只能相信。

明寫這**不是進度百分比**：已完成的 54 分裡做的是基礎設施最重的部分（抽取器、對拍
框架、字模、存檔、UI 分層），剩下的 46 分裡有 18 分（法術＋角色成長＋任務結局）
連一條 RE 都還沒有，不確定性遠高於已完成的部分。

### 知識庫補了四份主題文件

原本只有 `character-creation.md` 與 `image-formats.md` 兩份，其餘全散在 4,452 行的
RESEARCH-LOG 裡。補上 `audio.md`、`combat.md`、`text-and-fonts.md`、
`town-and-world.md`，並把 `docs/re/README.md` 的索引重寫——它原本停在
「RE-001 至 RE-039」，還說 SPEC-002／005 仍是 DRAFT。

音訊那一份是使用者點名的。重點：整份原版**只有一個音效**，而這是把母體窮舉完才敢
說的（六個 executable 的每一條 `IN`／`OUT` 都掃過，只有 TITLE 的六條命中音訊埠）。
`TANDYPIN.EXE` 名字看起來像 Tandy 音效驅動，實際上只是個 TSR——二手推論輸給一手證據
的例子。remake 走規格近似合成而不是錄音，理由是權利：合成波形是 remake 自有、可公開，
錄下來的 WAV 不是。目前遊戲裡一個音都沒有，卡在人耳驗收。

### 一次自己造成的重工

**RE-090 重做了 RE-078 已經解完的事。**寫 RE-090 時我讀 `sub_2A4E5` 讀出一套「隨機
遭遇挑怪」的規則，標成強推論寫進紀錄——而 RE-078 早就把同一支解完了，還更完整
（機率表在 payload 4、地形遮罩 `[0x1312]`、**季節**遮罩 `[0x1316]`、數量與等級的
擲法、能不能騎走）。我不但重做，還把第二層的季節遮罩誤讀成「區域」。

那一節整段刪掉，改成指向 RE-078，只留 RE-078 沒有的那一件事（怪物名字的載入路徑）。

> 規則：**動手解一支函式之前，先 grep 自己的 `docs/`。** `rulebook/00` 的檢查點
> 早就寫著這一條，但它的觸發時機寫成「要說解不出來之前」——而我當時是「解出來了」，
> 那個時機一次都不成立。**解出來的時候也要查**，成本是一次 grep。

### 清掉的其他過期斷言

| 位置 | 原本寫的 | 實際 |
|---|---|---|
| `VERIFICATION-MATRIX.md` | 戰鬥、存讀檔、背包、任務「未開始」 | 都做了；整份改寫 |
| 同上 | 玩家路徑十二份 | 十五份 |
| `README.md` | 「世界地圖與戰鬥都還沒實作」 | 都實作了 |
| `CONTEXT.md` | 「戰鬥的指令與命中結算還沒實作」 | 已實作 |
| `CONTEXT.md` | 3,117 個去重畫格 | 6,515 個、43 個 run |
| `WORKLIST.md` 狀態表 | 字模 2,366 字、安全矩形 6,693 項 | 2,434 字、6,685 項 |
| `SPEC-008`／`spec/README.md` | 戰利品與升級「DRAFT」 | 沒有 RE 證據，連草案都不是；改成「未開始」 |
| `docs/re/README.md` | 索引到 RE-039；SPEC-002／005 是 DRAFT | 到 RE-090；兩份都 READY |

**歷史記錄不動**：RE 條目與 WORKLIST 已完成項裡的數字寫的是「那一輪的成果」，
不是現況宣稱。判準是「這段話會不會影響後續的思考」——會的才改。

`spec/README.md` 另外新增一節「未開始的規格」，把戰利品／經驗值、法術、任務回報與
結局列出來，免得「沒有 DRAFT」被誤讀成「有草案但還沒審」。

## 2026-09-04（續五）：一場戰鬥打得完了

### RE-091

新探針 `tools/ida_probe_koltown_battle_victory.py` 倒出 `sub_289EE`（467 條）與
`sub_1E3A8` 裡的九個窗。解出來的：

- **每一輪之後數兩個數**（倒下的隊員、倒下的敵人），再依序判四個條件。
  順序有意義——打輸排在打贏前面，兩邊同時全倒算打輸。
- **打贏**：記錄 0x174 ＋ 英文複數（名字以 `S` 結尾補 `ES`）＋ `!`；金冠是每隻怪
  擲「骰數 ＋ 1」次，骰數在怪物記錄 `+0x0F` 的高 nibble，骰面查 `[0x00D4]` 的
  `2 3 5 7 9 11 19 99`，總額除以隊伍人數，每人不到 2 枚就歸零。
- **打輸**：`The]` ＋ 名字 ＋ 複數 ＋ `s have`／` has` ＋ ` defeated you!`。
- **逃走**：`You flee for your]` ＋ `life!`／`lives!`。
- **掉落物品**：`sub_289EE(0)` 是 20% 機率從 `item.lst` 的 61 種挑一件放進
  `0x9D20` 的空槽——那正是 `|H` 讀的那張表（RE-086）。

`word_345DA` 與 `word_34966` 這兩個計數器**在別處被當通用暫存重用**（`0x2137D`
拿 `word_34966` 放金冠的骰數），讀碼時不能只看名字。

### 實作

- `tools/extract_monsters.py` 補抽 `+0x0F`，`monsters.json` 重生。
- 新增 `internal/battle/victory.go`：`Verdict` 四態、`RoundVerdict()`、
  `GoldReward()`、`PluralSuffix()` 與三組訊息的記錄清單。
  型別叫 `Verdict` 不叫 `Outcome`——後者已經是「一次攻擊的結果」。
- `Step` 的收尾從「只判隊伍全倒」換成 RE-091 的四個條件，打贏時算好金冠。
- `cmd/kol` 加 `battleVerdictLines()`，打贏分兩句（戰果、金冠）各等一次鍵；
  並加一道「已經結束就不要再 Step」的閘門——少了它會把勝負訊息重排成按不完的迴圈。

**順帶接上戰鬥訊息的中文化。** `battleRecord` 原本直接讀 `bank.RecordPadded`，
也就是說戰鬥訊息一直是英文的，即使 486 筆譯文早就翻好了。新增
`town.UIRaw(bank, trans, ...)`，原文與譯文走同一套版面碼處理，中文模式優先用譯文。

### 沒做的，以及為什麼

- **冒險點數**（記錄 0x176）：算式要 `sub_26C5F(0x17)` 才成立，那一支還沒解讀。
  **寧可不報，也不要報一個錯的數**——所以 `RewardRecords` 只回金冠那一句，
  而且測試明文斷言它不會回冒險點數那一筆。
- **掉落物品**：規則已證實，但 `item.lst` 還沒抽成 typed data。
- **升級**：`sub_289EE(1)` 還沒讀。

### 驗證

`internal/battle` 新增六個測試（四個結束條件、金冠算式、英文複數、記錄清單、
`Step` 的收尾），全庫測試全綠。既有的戰鬥截圖重跑後與對拍收據逐 byte 相同，
沒有回歸。

**打贏的畫面還沒實跑過。**要把整場八隻怪打完才看得到那兩句，手算八隻的移動與攻擊
路徑不可靠，目前沒有那麼長的可重播腳本，也還沒有原版畫格可對拍。這一項寫進
SPEC-008 的已知差異與 WORKLIST，不當成做完。

### 下一步

**敵人那一手。**一場戰鬥打得完了，但敵人不會動手——那一手會被標成
`Unimplemented` 跳過，所以目前只有玩家單方面打、永遠不會輸。決策樹 RE-081 已解，
缺的三張基礎權重表在靜態影像裡是未初始化區，得用 DOSBox-X 在戰鬥中斷點取樣。

## 2026-09-04（續六）：敵人會動手了

### 先更正一個自己下錯的結論

RE-081 當初說三張基礎權重表「要執行期取樣才定得了值」，還加了一句「不是再多讀幾支
函式能補上的」。**使用者的直覺是對的**：BSS 的內容只有兩種來源——程式碼寫進去，
或從檔案載入——兩條都掃得到。

`sub_1E3A8` 的 `0x1EC49` 起由 `byte_399D6` 選兩條路：非零時從怪物的載入緩衝取，
為零時從 DGROUP `0x1092`／`0x10A6`／`0x10AE` 三張**靜態**表複製。而 `byte_399D6`
落在 BSS 且全庫沒有任何 writer，C runtime 清零之後恆為 0。寫成 RE-092。

> 規則：**找不到 writer 時，先確認搜的是「具名參照」還是「位址」。**
> 帶基底暫存器的寫入（`mov [bx+6A9Ah], al`）不會出現在具名符號的 xref 裡，
> 要對位址做範圍掃描才看得到。

### 實作

- `internal/battle/rings.go`：七個同心方環。半徑是**環號 ＋ 1**，起點 `(−R, 0)`
  沿邊順時針、格數 8R。位移用公式產生，`rings_test.go` 拿原版那七張表的 224 對
  逐項驗過——抄 224 個數字抄錯一個很難發現。
- `internal/battle/ai.go`：加權輪盤 `sub_25A3A`、攻擊式 `sub_2578F`、部位
  `sub_2590F`、移動三式，以及決策主幹 `PlanEnemy`。
- `round.go` 的 `AdvanceToParty` 輪到敵人時呼叫 `PlanEnemy`，不再直接跳過。

**中途修正一個實作錯誤**：第一版從環 0 一路掃到環 6，找到就攻擊，結果距離 5 格也
在「攻擊」。回去讀碼才發現 `0x23C1A` 把環號設成 **0**——**環 0 貼著才打**，
環 0 沒人才往外掃環 1…6，而那一條找到目標之後是**朝它走一步**（`0x2401E` 跳回
移動），不是攻擊。是測試先抓到的。

### 實跑抓到兩個真 bug

跑十六個回合的腳本，敵人一路走過來把 Poontang 打倒：

```
THE BINDERAK CUTS HIS SPINEY MACE AT POONTANG WHO CANNOT PARRY THE BLOW!
THE BLOW HITS POONTANG'S CHEST CAUSING A SERIOUS WOUND!
THE BINDERAK STABS HIS SPINEY MACE AT POONTANG WHO CANNOT PARRY THE BLOW!
THE BLOW HITS POONTANG'S CHEST MAKING HIM UNCONSCIOUS!
THE BINDERAKS HAVE DEFEATED YOU!
```

第一版印出來的是 `CUTS HIS  AT POONTANG` 與 `THE BINDERAKSS HAVE`：

1. **怪物的武器名沒帶進訊息**。`battleSpeaker` 只填隊員的 `WeaponName`，敵人的
   `|E` 因此展開成空。怪物記錄 +0x20 的武器名早就抽出來了，接上就好。
2. **打輸那一句的複數多一個 S**。打贏那一句是「名字 ＋ S／ES ＋ `!`」，打輸那一句
   接的記錄 0x178 是 `s have`——**本身已經帶了 s**，所以原版只在名字結尾是 S 時
   補一個 `E`。兩句用同一套後綴就會變成 `BINDERAKSS`。分成
   `PluralSuffix` 與 `DefeatSuffix` 兩支，各自有測試。

**這兩個都是單元測試結構上抓不到的**：規則層自己不組句，組句在 `cmd/kol`，而
訊息對不對只有把畫面上的字讀回來才知道。

### 沒做的

敵人 AI 只做了主幹。撤退狀態、反制掃描 `sub_253E7`、側翼 `sub_25653`、射擊那一條
與沒有武器時的指令 0x26 都還沒做——它們都要參戰者記錄裡幾個旗標的語意
（+0x05／+0x1D／+0x1E／+0x29／+0x2A／+0x2B），而那幾個旗標是別支寫的。
寫進 SPEC-008 第 8 節的「不授權實作的範圍」。

打贏的畫面還是沒實跑到（要把整場八隻怪打完），但**打輸實跑到了**。

自評 56 → 59。

## 2026-09-04（續七）：丟、撿與切換

### 三個指令碼的語意原本全推錯了

RE-093 發現執行段的分派只認十個指令碼，撿與丟都不在其中。用**訊息記錄號反查**
（七個記錄各只有一個 `mov ax, <記錄>`，比掃指令碼準）找到它們在同一支的前一段
`0x20189`…`0x20621`，倒出來之後三個語意都翻案（RE-094）：

| 指令碼 | 實際是 | remake 原本猜的 |
|---|---|---|
| `0x15` | 丟 | 撿（`CmdPickUp`） |
| `0x26` | 撿 | 地面（`CmdGround`） |
| `0x24`／`0x25`／`0x16` | 切換 | 丟（`CmdDrop`）與兩個未知 |

判準是各段印的訊息：`drops|B|E!`、`You have found a|H…`、`puts|B|I in|B belt`。
原版 frame-887 的指令列因此讀成「YES、武器攻擊、徒手、**切換**、**丟**、移動、休息」
——值一個都沒變，變的是名字。

同一輪還更正了 `0x0D`：那是**施法**（`sub_2818D`）不是背包，也是從圖號推錯的。

### 實作

`internal/battle/items.go`：戰場十六格物品（物品在 `0x9D20`、X 在 `0x2BAE`、
Y 在 `0x2D9E` 三張平行的表，remake 併成一個結構）與三手的規則。執行段的 `act`
接上三個 case，`Unimplemented` 只剩投擲。

### 實跑又抓到一個

第一版印出來是 `POONTANG DROPS HIS !`——**原版是先印訊息再搬東西**，`|E` 展開的
時候武器還在手上；remake 的規則層做完才輪到 UI 展開，名字已經沒了。
`ActResult` 因此多一個 `Item`，把丟掉／撿到的名字帶出來。修完是
`POONTANG DROPS HIS LONG SPEAR!`。

**這是這一輪第三個只有實跑才抓得到的 bug**（前兩個是怪物武器名與打輸的複數）。
共通點都一樣：規則層自己不組句，組句在 `cmd/kol`，而訊息對不對只有把畫面上的字
讀回來才知道。

### 沒做的

撿那一句的問話（原版對隊員會先問「要不要撿」再等鍵，remake 一律直接撿——那要在
執行段中間插一次互動）與記錄 0x113（任務物品）那一條分支。兩項都寫進 SPEC-010 的
「與原版的差異」。

自評 59 → 61。

## 2026-09-04（續八）：擊退、打暈，以及一個一直沒對過的疑點

### 擊退

RE-095：攻擊命中之後判一次，**機率就是體型差的百分比**——差 40 就是四成，同體型
或更小完全推不動。推的方向從 `[0x123A]` 的三個候選裡 `rand(3) − 1` 挑一個，
不是固定往正後方。推得動就把目標移過去印 0x159，推不動印 0x158；兩條路最後都讓
對手掉武器，並把它這一回合還沒做的那一手取消掉。

訊息因此要排隊：攻擊命中本來就分兩句，擊退成立還會再多一到兩句，各等一次鍵。
`Exec.Pending` 從單獨一段改成 `[][]string` 的佇列。

### 順帶解開一個一直沒對過的疑點

同一個欄位 `+0x1F` 有兩組不同的位移表：環的 `[0x0FD8]` 與方向的 `[0x0C1C]`。
兩者差一個 `環索引 = (6 − 方向編號) mod 8` 的映射——**環索引與方向編號不是同一套**，
RE-081 寫的「環 0 的順序與方向編號相同」不成立。

再往下推就知道 **`+0x1F` 是欄、`+0x20` 是列**（方向 0 是「上」，而它的位移加在
`+0x20` 上）。`[0x123A]` 的候選方向逐項對得上，是獨立的第二個證據。

**remake 的軸是反過來的**：`Actor.X` 是列、`Y` 是欄。戰場格子在抽取時就轉置過
（RE-080），環位移表抄進 `ring0` 時也把原版的 X 欄放進了 `Row`——兩次對調抵消，
所以 remake 的 `ring0` 索引等於方向編號，畫面與移動方向都正確（四對戰場對拍是
它的證據）。

> 這種「兩個錯誤互相抵消」的狀態很脆弱：任何一邊被「修正」成與原版一致，另一邊就
> 會壞掉。`ring0` 與 `Cell` 的註解都寫明了，並在 RE-095 留了完整推導。
> **查一個對不上的數字，比假設它無關緊要划算。**

### 沒實跑到

擊退要湊到「攻擊者比目標大」才觸發：手上這一份原版存檔裡 Poontang 的體型是 114、
BINDERAK 是 100，敵人打過來永遠推不動，而 Poontang 打過去常常一擊斃命。
規則層有五個測試，實跑列為已知差異寫進 SPEC-009。

自評 61 → 62。

## 2026-09-04（續九）：戰後結算，以及一個把八十四次實跑浪費掉的抽樣錯誤

### 找錯地方

上一輪把冒險點數記成「`sub_26C5F(0x17)` 未解讀」，並猜結算在 `sub_289EE`。兩件事
都不對：獎勵整段在 `sub_1E3A8` 裡（`0x21258`…`0x214B7`），`sub_289EE` 是它結束前
呼叫的戰利品分配畫面。

找對的方法值得留下來。`sub_1A8FD(記錄號, 緩衝)` 是全庫唯一的訊息輸出入口，呼叫點
前面必定有一條 `mov ax, imm` 把記錄號推上堆疊。掃過全部 331 個呼叫點回推常數，
301 個一次解出——「哪一段程式印哪一句」從此是一張可查的表
（`tools/ida_probe_koltown_message_index.py`）。要找某句玩家看得到的字是誰印的，
不必再逐支反組譯。這一份索引以後每一輪都用得到。

### 解出來的

冒險點數、金冠、戰利品清單、升級的歸屬，寫成 RE-096。三個先前記錯的地方一併更正：
金冠的外層輪數是 `word_3B26A`（野外恆為 1）不是敵人數；骰式只讀第一隻怪的；
升級要 `word_345DE != 0xFFFF`，而戰鬥初始化就是 `0xFFFF`——**野外打贏不會升級**，
那是城鎮訓練那一側的事。

掉落物沒有座標這件事上一輪找不到 writer，是對的：它們根本不落在地上。`0x9D20`
那十六格在戰鬥中是地面物品、戰鬥結束後就是戰利品池，所以打到一半丟在地上沒撿的
東西，戰後會一起出現在清單裡。

### 抽樣不是實驗

戰利品的畫面實跑卡了很久。掉落是 19% 的機率，我用 seed 1…84 掃了八十四次，一次
都沒中，一度以為是 bug——加了臨時輸出才看到 `DropSpoils` 每次都有跑，只是那一擲
永遠 ≥ 20。

原因是 LCG：走固定步數 N 之後 `state = A×seed + C`，A 是常數。**連續的種子加上
固定的步數，那一擲的值就沿著一條直線走**，可以整條錯過想要的區間。換成散開的大數
（12345、987654、…、33554393），第十二個就中了。

規則寫成一句：**用同一個亂數源掃機率事件時，種子要散開，不能用連續整數**。這條和
「固定參數重試 N 次不是實驗，是抽樣」是同一件事的另一面——那一次是參數沒變，
這一次是參數變了但變得太整齊。收據 `docs/verification/battle-spoils/receipt.json`
把它記在 `seed_note`。

### 順手抓到的一個真 bug

戰利品清單的標題 `~H02~V09Items Found:` 一開始畫到畫面左上角去了。`town.UIRaw`
會剝掉版面碼——那對「串接成一句話」是對的，對「座標就是內容」的框標題是錯的。
新增 `town.UISegments` 走保留版面碼的那一條，原文與譯文共用。

### 驗證與提交

`go vet`、全部單元測試綠。實跑取到打贏那一句、獎勵那一句與戰利品清單三張畫格，
收據在 `docs/verification/battle-spoils/`。那一次用了觀察用的替代輸入
（`tools/make_observation_inputs.py`：十二隻 3 級怪換成一隻壓弱的），所以標成
**觀察，不是對拍**；戰利品框的內部版面在程式註解與 SPEC-008 都標了未對拍。

容器全部 `--rm`，跑完沒有殘留（`docker ps` 只剩使用者自己的 portainer）。
提交 `398682a`，未推送。

自評 62 → 64。

## 2026-09-04（續十）：隨機遭遇的擲點、站位，以及一個被自己弄壞的 IDA 資料庫

### 做完的

RE-097 把 RE-078 的「強推論」逐條對過 `sub_2A4E5`，流程完全吻合，升到 `已證實`；
順帶解掉兩個誤解（`word_3B26A` 是敵人等級不是常數 1，參戰者記錄那一格的等級是它
÷ 5 ＋ 1、與原始值是兩個尺度）。RE-098 找出戰場站位的兩條路與四張固定表。

規則層與觸發點都實作了：`battle.RollAmbush`／`RollForce`／`RideAway`／
`AmbushPartyPlace`／`AmbushEnemyPlace`，加上 `world.Session.OnTimePart`。
SPEC-011 從 SPEC-007 拆出來。**還打不起來**——戰場的地形圖還沒解。

### 兩個找法值得留下

**訊息記錄的全域索引**（上一輪建的）這一輪又用到：找「哪一段程式印哪一句」不必逐支
反組譯。

**對位移範圍掃 operand**：`0x2E0C` 那張站位表沒有具名 xref，因為它全部是
`mov [bx+2E0Ch], al` 這種帶基底暫存器的存取。改成走過每一支函式的每一條指令、
比對 operand 的**值**落不落在範圍內，一次就找到唯一的 writer 與四個 reader。
這和 RE-092 的教訓是同一件事：**靜態符號掃描結構上抓不到間接存取**。

### 事故：連續開關同一個 `.i64` 把它弄壞了

跑到第十幾支探針時 `idat` 開始報

```
Database is empty -> OK
Database initialization failed with error 4
```

`~/.claude/knowledge-base/retro/ida-pro-9.4.md` 早就寫了這一條：**對同一個 `.i64`
連續快速開關會把資料庫弄壞**，而且「不限於同一支腳本的迴圈，任何連續執行都算」。
我這一輪為了追一條線索，改一個常數就重跑一次，跑了十幾次。

處置：把壞掉的 `.i64` 改名留證（`.i64.broken`），`idat -A -B <EXE>` 重建，再拿
`ida_probe_koltown_ambush.py` 重跑一次，六張表與三支函式的反組譯**逐條相同**才繼續。
規則寫進 `tools/run_ida_probe.sh` 的標頭：**一支探針要一次問完所有想看的位址**，
不要改一個參數重跑一次。

教訓的形狀和 `rulebook/40` 那條「固定參數重試 N 次不是實驗」很像，只是這一次代價
不是浪費時間，是弄壞了工具本身。

### 另一個抽樣錯誤（上一輪的延續）

上一輪用連續整數種子掃 19% 的掉落，八十四次一次都沒中。這一輪確認原因是 LCG 走固定
步數之後狀態對種子是線性的。兩件事合起來寫成一條規則：**掃機率事件時種子要散開**。

### 驗證與提交

`go vet`、全部單元測試綠。提交 `107335f`（等級尺度的修正）與 `f69e1b4`（遭遇的規則
層），未推送。容器全部 `--rm`，沒有殘留。

自評維持 64（野外那一列 85% → 88%，權重 9，加權後不足以進位）。

## 2026-09-04（續十一）：走在野外會被伏擊了

### 最後一塊差點又找錯

追隨機遭遇的地形時，我一路讀 `sub_1E3A8` 找「地圖載入」，正要下「還沒解」的結論——
`battle.Generate` 幾天前就寫好了，註解連位址（`0x1E594`）都標著。
`rulebook/00` 的「要說解不出來之前先 grep 自己的 `docs/` 與程式碼」這一次抓的是
**自己的程式碼**，不是筆記。

順帶把 `word_39A4E` 解掉：它載入的是 KOLPICS 記錄 8，那是**磚圖**（`battle-tiles-08`）
不是地圖；檔名由 `0x184A` 的 `"kol%s"` 加 `0x0670` 的副檔名表組出來。寫成 RE-099。

### 這一輪做完的

RE-097（擲點流程逐條核對）、RE-098（站位的兩條路與四張表）、RE-099（地形是擲出來的）
補齊了隨機遭遇的四個必要條件，`battle.NewAmbush` 因此開得起一場戰鬥。
`main.go` 的 `OnTimePart` 接上：時段一變就擲一次，遇到就進那張怪物圖與
「YOU HAVE BEEN AMBUSHED BY n, LEVEL m NAMES!」。

實跑走九十步被伏擊兩隻 1 級 BINDERAK，收據 `docs/verification/battle-ambush/`。
那一組**沒有換過輸入**，與 `battle-spoils` 不同。

順帶兩個修正：`parseSegments` 把 `]` 丟掉（原版是換成空白，`sub_1A8FD:0x1A990`，
少了它會印成 `BY2`）；戰鬥結束後沒有人放開畫面，打完會卡在戰場。

### 遭遇讓兩條既有的收據走不完

`world-quest`（91 步走到任務地點）與 `world-clock`（28 步走到換日）在實作遭遇之後
被伏擊攔下——**原版量測也是這樣**（WORKLIST 早就記過「兩輪量測都在半路被伏擊攔下」）。

處置不是關掉遭遇，是**釘住亂數種子**：遭遇是機率的，91 步走完不被攔下的機會約一半。
掃了八個種子，world-quest 用 3、world-clock 用 2，兩條的檢查點都與原本已對拍的截圖
**逐 byte 相同**——換種子不影響那兩張，因為一月的天候上下限窄，多數擲骰不會改變天候，
而遭遇擲在天候之後。十五條玩家路徑重跑後全部 `failures: 0`。

這個取捨寫進了那兩份 plan 的描述：**那兩條要驗的不是遭遇**，遭遇本身另有單元測試與
自己的實跑收據。

### 驗證與提交

`go vet`、全部單元測試綠；十五條玩家路徑逐條驗過。提交 `bf64244`，未推送。
容器全部 `--rm`，沒有殘留。

自評 64 → 65（野外 88% → 95%）。

## 2026-09-04（續十二）：法術

### 一次倒完再讀

`sub_2818D` 有 602 條。這一次照上一輪的教訓**一次問完**——函式、呼叫關係、
兩張 DGROUP 表、四個相關的小函式、三個視窗全部列進同一支探針，讀完再補問一次。
`.i64` 沒有再壞。

### 解出來的

RE-100。法術不是獨立的資料檔，是存在**角色記錄 0x200 起、每 16 byte 一筆、
共十六筆**，版面與商品記錄相同。`Member.PackCount` 那個掛了很久的「語意未定」
欄位數的正是這一份——指令列在「有法術**而且手上空著**」時才放出施法那一格
（`0x22A82`），語意現在對上了。

骰式與武器傷害碼差一個位移，各有各的出處，沒有合併。擋招比高低 nibble，敵人看怪物
記錄 `+0x1D`、隊員看角色記錄 `+0x28`。三種效果：6 是指定部位、5 改疲勞、其餘查
`[0x12A4]` 得欄位再加減強度並掛倒數；倒數到 0 把屬性整批重算回原值。

效果 5 改的 `[0x322E]` 也解掉了：那是**疲勞累計**，`sub_23683` 折算成
`50 × 疲勞 ÷ [0x1C50]`，攻擊值再乘它除以 200 扣掉。找法同 RE-098——對 DGROUP 位移
做範圍掃描，因為它全部是帶基底暫存器的存取。

### 實跑要先丟武器

Warder 是原版存檔裡唯一帶法術的角色，但他手上有一把射擊武器，所以第一回合指令列上
**沒有**施法那一格——這不是 bug，是原版的條件。丟掉之後第二回合才施得了法。
收據 `docs/verification/battle-spell/` 的描述把這一段寫進去了。

四張畫格：「WHICH SPELL? / VARTWELYRMI」、「SELECT AFFECT / POSITIVE / NEGATIVE」、
「WARDER CASTS VARTWELYRMI AT WARDER.」、「THE SPELL HAS A MILDLY NEGATIVE EFFECT.」。

### 還沒做的

選部位那一次互動（要在執行段中間插一次等鍵）、效果 5 的疲勞鏈、效果 7／8 的兩格
欄位語意，以及法術的母表（角色是從哪裡學會的）。

### 驗證與提交

`go vet`、全部單元測試綠。提交 `55a9329`（規則層）、`f169949`（資料接線）與這一次的
選單，未推送。容器全部 `--rm`，沒有殘留。

自評 70 → 73。

## 2026-09-04 戰鬥的收尾三塊：任務旗標、射擊、逃走

### 任務標成完成（RE-101）

`sub_28F5D` 的寫入端。位址算法與讀取端（RE-079）同一套，但判斷只比 bit 1
（`0x28FB9`），所以**沒接那個任務的角色也會被寫成 2**，六名角色一律走完。
remake 照原版做，測試的期望值原本寫成「只標接了的那一位」，是我把測試寫錯了，
不是實作錯。

`word_34968` 這個「是不是任務地點」的開關用 `Site.Index >= 0` 表示。

### 射擊（RE-102／SPEC-013）

`sub_25D7A(0)` 是 Bresenham，`sub_25F8A` 逐格判：樹扣 35 點命中值、擋路的磚就停、
半路的人各扣體型的一半。走線與視線是同一段碼，抽成 `walkLine` 共用。

三件順手解出來的事：

1. **`+0x1E` 讀反了**。`sub_25F8A:0x2606C` 是「非零才穿得過牆」——騎著的人視線高。
   RE-079 原本寫成「為 0 才穿得過」，方向相反，`Field.Visible` 跟著補上騎乘那一條。
2. **防禦分數是負的要當 0 再比**（`0x27454`）。近戰原本漏了這一條，命中餘裕會灌水。
3. **投擲（0x22）在 `sub_26D76` 裡一處都沒有**。它不拉彈道也不印動作訊息，
   指令列那一格按下去的實際行為沒有證據，所以維持「還沒實作」，不假裝做得出來。

### 逃走（RE-103／SPEC-014）

`sub_2886C(0x30)` 不是體能檢定——它把角色記錄 0x30 起的十六個 byte 與 `[0xB926]`
逐 byte 比。位移 0x30 是手上那件東西的名字，所以逃走那一段問的是「站在戰場邊上
的人手裡拿的是不是任務要的那一件」。**任務不必打完**：搶到東西、走到最外圈、
逃掉，一樣記完成。

RE-080 對這一段的舊敘述（`((力量 + 智力) ÷ 2) × 等級 ÷ 3`）沒有對上位址，已更正。
RE-079 說 0x109 接在 0x108「Begin next round?」答否之後，也不對——兩者無關，
0x109 由執行段走完之後的 `0x21045` 決定。

### operand 掃描：升級屬於競技場

`tools/ida_probe_koltown_writers.py` 走過 338 支函式的每一條指令比對 operand 的值。
寫 `word_345DE`（誰要升位階）的只有四處，三處清成 0xFFFF，唯一寫真值的是
`sub_2AAA8:0x2AFE6`——競技場。所以升級不屬於戰鬥模組，也不屬於城鎮訓練
（訓練只會叫玩家「Go to the Arena!」，RE-074）。WORKLIST 與規格索引跟著改。

`byte_3AC7D` 與 `[0xB926]` 一次命中都沒有：它們是被整塊複製寫進去的，
指標間接的寫入在 operand 掃描裡結構上看不到。

### 敵人會射擊了（RE-104）

`sub_25D0F` **回 0 才是通得到**，兩個呼叫點都拿 `jz` 接著做事。做成
`Field.LineOfSight` 時把回傳值翻正，免得每個呼叫點都要記得反著讀。

敵人在環 1、2 上看得見、種類非零又有右臂時改成射擊，而且 `+0x1D` 直接設 1
——**敵人不必先花一回合裝填**，玩家那一側要。

反制掃描 `sub_253E7` 的優先度表也讀完了，但 `word_30916` 的語意還沒定
（主幹在別的分支把它當移動方式用），照抄會做出自洽而錯的行為，只寫進 RE 不實作。

### 規格補在後面

射擊與逃走這一輪是先實作才補規格，順序反了。補上 SPEC-013 與 SPEC-014，
狀態 READY，授權範圍逐項限定；SPEC-009 的射擊段移出去。

### 追記：0x22 不是投擲（RE-105）

規劃段的分派表（`0x1FA02`）沒有列 0x22，它落 default 直接收尾，而收尾寫進 `+0x25`
的是 `sub_228B7:0x22D44` 存下來的圖號。執行段 `0x20785` 把 0x22 與 0x23 一起送進
`0x20799`，0x22 在那裡必定變成裝填。

指令列那兩格的出現條件與這個檢查完全對稱（RE-083）：0x22 是「還沒上弦」、
0x23 是「可以射」，推 0x23 那一格時還順手設 `+0x1D = 1`。所以 **0x22 與 0x23 是
同一件事的兩個階段**，`CmdThrow` 這個名字從一開始就取錯了，改成 `CmdLoad`。

一起解掉的是 `word_30916`：`0x1FA6D` 把它寫進 `+0x24`，**是防禦選項**。RE-104 把
反制掃描寫進它的值標為未定，現在定了；反制掃描本身仍不實作，理由改成「優先度
規則沒有可觀察的驗證方式」，不再是「語意未定」。

順帶把 `chooseCommand` 的 default 補上：分派表認不得的圖號記進 `Unimplemented`。
以前那個記錄點掛在 0x22 那一格上，0x22 一實作就沒有人記了。

### 追記：實跑抓到三個畫面層的洞

規則層的單元測試全綠，但畫面上什麼都沒有——這一輪連續踩到三次：

1. **射擊撞樹／撞牆那兩句不見了**。`battleStep` 推訊息的迴圈只看 `Message` 與
   `Outcome`，那一手只有 `Records`，整手被跳過去。
2. **逃走那一問沒畫**。`PromptFlee` 是新的一層，`internal/ui/battle.go` 只認
   `PromptExecuting` 才畫卷軸。畫面停住、按鍵無效，看起來像當掉。
3. **答應逃走之後沒印「You flee for your life!」**，直接回野外。

三個都是「規則層回傳值正確、畫面上沒有」，單元測試結構上看不到。收據
`battle-missile/` 與 `battle-flee/` 就是為了擋這一類而做的。

### 替代輸入的坑：換了站位，畫面逐 byte 相同

做逃走收據時把隊伍站位挪到戰場邊上，跑出來的畫面與沒換的**完全相同**。原因是
那個 seed 在野外那七十步先觸發了伏擊，走到的是隨機遭遇——站位用固定表
（RE-098），地點自帶的座標根本沒被讀。換 seed 才走得到任務地點。

**「換了輸入但輸出沒變」要先問走到的是不是同一條路**，不要先懷疑載入失敗。
這一條寫進了 `docs/verification/battle-flee/receipt.json` 的 `seed_note`。

## 2026-09-04（續）結局做完了

RE-017 早就把七張 ENDPICS 抽出來、也知道 `sub_2B4A7` 連著呼叫它們七次，但那一條
線一直沒有往上追。這一輪把 `sub_2B4A7` 與它**唯一的呼叫點**倒出來，整條就通了。

### 觸發在任務戰鬥的尾端

`sub_2B0FF` 打完那一場回來，先看 `word_34968` 等不等於 0xFFFF——**那正是
`sub_28F5D` 標任務完成時設的值**（RE-101 把它記成「語意還沒對上」）。等於才逐一
走六名角色，其中一名的 `[0x1DD + word_2DF70]` 等於 0x17 就播。

那一格是誰寫的還沒解，所以 remake **不接觸發**，只把七段做出來用 `-scene ending`
播。照著假說接，會做出一條永遠不會成立（或永遠成立）的路。

### 四個全域是矩形的四個邊

每一段畫圖之前設 `word_3499E`／`word_349A0`／`word_34A06`／`word_34A0E`。
右欄減左欄乘八、下列減上列，**七張全部等於 RE-017 從 LZW 像素數反推的尺寸**。
兩條獨立的證據對上，四個全域的語意就定了——這比逐條讀繪製函式快得多。

### 文字要接起來再拆座標

一段兩三筆記錄，第二筆開頭那幾個字**沒有座標**，接在第一筆的最後位置
（原版逐筆送進同一個緩衝區，RE-088）。分開解析會把它畫到左上角——第一次跑出來
就是那樣，`town.JoinedSegments` 是為此加的。

ENDTXT 的譯文因此也接上了畫面，中文化那一項從「BOOKDATA 與 ENDTXT 都還沒接」
變成只剩 BOOKDATA。

自評 76 → 78。

### 任務的回報也解出來了（RE-107）

`[0x1DD]` 的 operand 掃描找到五個使用點，其中一個在建築內的主迴圈——訓練也在
同一支。倒出來一看就是**任務的回報**：

```
建築的 byte_2FBFA & 0x7F  = 這棟建築對應的任務地點編號
旗標欄位 == 2（完成）的角色：
    欄位 |= 除數            ; 2 → 3
    [0x1DD + 群組] += 1     ; 已回報的任務數加一
```

所以那個 2-bit 欄位有**四個**狀態，不是三個：0 沒接、1 接了、2 完成、3 回報過。
RE-079 與 RE-101 只解到 2。

接任務那一側（`sub_14564`）要欄位為 0，而且 `byte_2FBFA` 的 **bit 7 是「最終任務」**
——那一棟要 `[0x1DD] >= 0x17` 才開放。0x17 正是結局的判準（RE-106），
所以「回報滿二十三個才接得到第二十四個，打完它就播結局」三邊對得起來，
RE-106 標成假說的那一句升成已證實。

**這是 operand 掃描第二次直接把一整塊解出來**（第一次是升級屬於競技場）。
帶基底暫存器的存取沒有具名 xref，走過每一條指令比對 operand 的值反而比
從呼叫圖往下追快。

自評 78 → 79。

## 2026-09-04（續）任務交付與發行打包

### 任務交付的物品找到了（RE-109）

RE-107 找到「回報」那一段，但它不是進門就發生——前面還有一整塊：
建築 `icon_flags[1] & 4` 開啟時，逐一比對六名角色身上每十六 byte 一格與 payload
`0x2E0` 的那件物品，全同就換成 `0x2F0` 那件，再走回報。

拿那個旗標篩七座城的 102 棟得到 **23 棟**，payload `0x36` 的低七位全部落在
0…23，bit 7 只有一棟設——TOWN7 的 `The Assembly Building`。物品名字也解得出來：
`Truth Sword`、`Village Flag`、`Kelder Crown` → `Flying Cloak`…

**RE-107 原本把那一格標成未解是因為沒有篩選。**把 102 棟全列時看到 90／100／120
這種超界的值，其實是沒有那個旗標的建築在那一格放了別的東西。加上篩選就完全對上。

負面結果先做正對照：23 件任務物品去比原版十四名角色身上那十七格，一件都沒中
——那是全新存檔，合理。而同一批位元組用 5-bit 名稱解碼器解得出 `Long Spear`、
`Barbut Helm`、`Greatsword`，位移是對的。

規則層 `internal/town/quest.go` 做完了（四個狀態、交付、最終任務的門檻）；
寫測試時才發現原版那個迴圈是**先加再比**，所以最後一格是 0x190 不是 0x1A0。

### 發行打包（M6 的第一版）

`dist-all/0.1.0/` 三個包建得出來也驗得過：

- `release/` **一個原版衍生位元組都不帶**，deny-list 在打包與驗收各擋一次；
  玩家拿自己的原版磁片跑抽取器產生素材（專案契約明文允許的那一條路）
- `full/` 含素材、附 DO-NOT-DISTRIBUTE.txt，只留本機
- 驗收含**啟動**：拿包裡的執行檔畫封面與結局，畫得出來才算過

只做 linux/amd64。Ebitengine 要 CGO，Windows 與 macOS 要另外的工具鏈。

自評 80 → 82。

## 2026-09-05 正常玩家路徑貫通

完成定義要的是「啟動 → 建立隊伍 → 城鎮／世界探索 → 戰鬥規劃與結算 → 任務 →
存檔 → 離開 → 重載 → 結局」走在**同一串按鍵**上。四條收據把它蓋滿了：
`full-chain`（567 步）、`full-chain-reload`（119 步）、`flow-quest`（109 步）、
`full-chain-ending`（28 步）。

### 三個只有實跑才看得見的洞

**flow 根本沒建野外那一套。**`-scene flow` 的初始化只涵蓋 `sceneWorld` 與
`scenePlay`，走到城牆邊選 DOOR 之後畫面就停在那裡——而世界地圖的**預設起點
(40,23) 剛好等於霍布入城點 3 的座標**，所以「出城之後座標對不對」這件事看起來
一直是對的。從別的城門出去才會露餡。

**鍵位有兩份，其中一份少了空白鍵。**`cmd/kol` 的 `applyPlayToken` 與
`internal/flow` 的 `townKey` 各維護一份 switch，後者沒有 `SPACE`——走 flow 進建築
之後敘述翻不了頁，畫面卡在第一段，看起來像當掉。收進 `town.Session.Key` 只留一份；
出城那一步的 `OnLeaveTown` 一起移進來，原本掛在 cmd 的包裝上，走 flow 的鍵繞過了它。

**建築 typed data 的載入器漏抄了任務欄位。**`LoadBuildings` 建 `Door` 時沒有複製
RE-109 那六個 `quest_*` 欄位，所以 `Door.Quest()` 永遠回 nil。抽取器一直有輸出、
規則層對、`quest_test.go` 全綠——但玩家走進收任務物品的建築什麼都不會發生。
**整條任務交付在遊戲裡從來沒生效過。**加了一條讀檔測試擋回歸。

這三個都是 CLAUDE.md 那條「綠色單元測試只證明 remake 自洽」的實例：三處的規則層
與測試都是對的，接線斷在別的地方。

### 長路徑要換一種收據

上千步的路徑不可能逐步存圖。加了兩個旗標：`-trace` 把每一步停在哪個畫面、隊伍在
城裡或世界地圖的哪一格、戰鬥走到哪一階段寫成 JSON；`-shot-steps` 只在指定的幾步
截圖。`tools/verify_full_path.py` 檢查軌跡的里程碑、檢查點截圖與寫出來的存檔。
**軌跡是遊戲自己報的狀態，不是從畫面猜的**——這比多存幾百張圖有用。

### 腳本是怎麼生出來的

戰鬥那 375 步不是手寫的。寫了一支外部驅動器：跑一次腳本 → 讀 trace 的最後一筆 →
依規劃段停在哪一層決定下一批 token → 再跑。產出的仍是一串固定按鍵，遊戲本身沒有
任何自動化旗標。三件事值得記下來：

- **`Prompt` 的編號數錯了一位。**以為 8 是執行段，其實 8 是「Begin next round?」、
  10 才是執行段。行為上碰巧接近（空白鍵在那一層是「答應並走一手」），所以錯了
  二十幾個回合都沒察覺，只是每個回合多耗掉一堆按鍵。
- **選了攻擊卻打不到人。**攻擊式那一列的第一個選項（0x2A）換算出來的攻擊選項是
  **0**，防禦式那一列的第一個也是 0——`ResolveAttack` 看到兩個都是 0 就把整手改成
  休息。二十二個回合裡敵人一點傷都沒有，訊息全是「X is resting!」。要真的打出去
  得挑第二個。
- **六個人朝同一隻走會擠成一列。**後面幾個一直「blocked from moving」。改成依隊伍
  次序分給兩隻敵人，再挑一格沒有人站的相鄰格前進，四個回合就打完了。

探索期間每次重編一遍 Go 是主要成本（一次 20–30 秒）。改用事先編好的執行檔之後
一次 1 秒，快了兩個數量級；正式收據仍由 `run_kol_shots.sh` 從原始碼重建。

### 存檔補了三欄

任務旗標、已回報數與身上帶著的東西都會在遊玩中改變，不寫進存檔重載之後就沒了。
物品存的是原始位元組而不是名字——交任務要逐 byte 比對建築要的那一件。
三欄都是選填，舊存檔照樣讀得開。

### 結局走得到了

觸發條件是已回報數等於 0x17（RE-106）。接任務那一側（`sub_14564`）還沒實作，
玩家沒有辦法在遊戲裡把那一格累積上去，所以用 `tools/make_final_quest_save.py` 把
存檔改成「已回報 22 件、最後一件已完成、東西就在身上」，再用正常按鍵走進霍布的
集會所交出去。七段結局與 `-scene ending` 那組收據**逐 byte 相同**。
這一組因此掛 observation，收據裡列了換過什麼、不宣稱什麼。

自評 82 → 83（加權和 82.90）。這一輪的進展主要是 gate 不是新系統：分數只動了
任務與結局那一列，結局從「只能用 `-scene ending` 直接播」變成走得到。

## 2026-09-05（續）競技場的下注，以及一個實跑才看得見的站位偏差

競技場的其餘部分（選人、開打、升階）上一輪已經做完，這一輪補的是中間那一段賭場。

### 下注

多人隊伍答應進場之後先換到賭場（記錄 0xEC／0xF0），沒上場的人可以押那一位會贏還是
會輸。金額由 `sub_1B5C8(0x2E3A, 2)` 讀兩位數，所以上限是 99；押他會輸就把金額取負
（`0x2AFAF` 的 `neg`）再多印一句「Good friends, eh?」。戰後結算（RE-103）看那一格的
正負決定印 `wins a bet!` 還是 `loses a bet!`，打輸那一場時整排先取負，所以賭中的
一律變成正值。

實跑時卡在賭場那一段敘述推不動：`Continue()` 開頭是

```go
if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
    return
}
```

而競技場不在建築裡，沒有 door。門的檢查移到競技場那個分支後面就通了。**三個單元
測試全綠的情況下卡住的**——鍵位那一層沒有測試覆蓋到「不在建築裡的建築畫面」。

收據 `arena` 從 2,133 步變成 2,153 步，中間押了兩注：Moraine 押 25 賭 Poontang 會贏、
Fighting Bull 押 10 賭他會輸。最後一頁三行：`YOU HAVE GAINED A LEVEL!`、
`MORAINE WINS A BET!`、`FIGHTING BULL LOSES A BET!`，錢包 2149 → 2174 與 442 → 432。

### 站位偏差

收據第 1,784 步的 trace 把六名隊員全部列成 `party`，站位是隨機遭遇那一組。回頭讀
`sub_1E3A8` 才發現 `word_345DE >= 0`（就是競技場與訓練戰）走的是第三條分支：只有
上場那一名參戰（固定站欄 0x14、列 0x1F），其餘隊員的記錄 `+0x01` 設成 2 並排到
DGROUP `0x12AC`／`0x12B2` 那張觀眾席表，唯一那隻敵人固定站欄 0x16、列 0x1F。

那三個敵人常數 IDA 命名成 `byte_3470F`／`byte_3472D`／`byte_3472E` 三個獨立的 byte，
看起來與站位無關；它們其實**就是參戰者記錄 6**——`0x6712 + 0x32 × 6 = 0x683E`，
加 DGROUP 基址 `0x2DED0` 正好是 `0x3470E`，三個位址是它的 `+0x01`／`+0x1F`／`+0x20`。
那幾條指令用絕對位址而不是 `[bx+…]`，所以符號掃描看不出它們同屬一筆記錄
（與 RE-092、RE-098 踩到的是同一類坑）。

分支與三個常數都從既有的反組譯讀得出來，只差兩張觀眾席表各六個 byte 的內容——
留給下一次開 `.i64` 時連同敵人 AI 那三支一起問完。這一項已寫進 `arena` 收據的
`not_claimed` 與 WORKLIST 的「下一個唯一工作」。

自評 85.63 → 85.84（角色成長 85% → 88%）。

## 2026-09-05（續二）競技場改成單挑

上一段留下的站位偏差當場做完了。缺的只有兩張觀眾席表各六個 byte，而它們不必開
`.i64`——**初始化資料本來就在 unpacked 的執行檔裡**，只要定位 DGROUP 的檔內起點。

錨用任務旗標的除數表（`0x7C` = `01 04 10 40`，RE-105），它在檔內只出現一次，
位置 `0x2034C`，所以起點是 `0x202D0`。正對照拿隨機遭遇那六組敵人站位
（144 byte）驗，逐 byte 對上已經寫進 `ambush.go` 的那份表。讀出來是：

```
0x12B2  欄  16 16 47 47 47 0
0x12AC  列  25 35 36 30 31 0
```

上場那一名站 (20,31)、對手站 (22,31)，觀眾在兩側。第六格是 0——隊伍最多六人、
上場一名，觀眾最多五個，那一格永遠排不到人。讀取器留成 `tools/read_dgroup_table.py`，
正對照寫在裡面：定位錯了一樣讀得出「看起來像資料」的東西。

實作是 `battle.NewArena`。順帶修了擲骰的順序——原版先擲陣型（`0x2AFF9`）才擲種類
（`0x2B012`），而這一版反過來。陣型這一場用不到（站位是固定的），但那一次擲骰要擲，
否則之後的亂數序列會偏掉。

### 收據要重跑，而且結果變了

六個人打一隻變成一個人打一隻，原來那 369 步的戰鬥腳本整串作廢。重跑驅動器，
第一次 **Poontang 兩個回合就被 HILL GIANT 打倒**——單挑同級對手本來就會這樣。
那一條路徑仍然是有效的：`Moraine loses a bet!`／`Fighting Bull wins a bet!`
正好驗到「打輸那一場時整排先取負」（RE-103），押他會輸的反而賭中。

不過收據要留住「打贏升一階」這個里程碑，所以改讓 Moraine 上場（清單上按一次 D）。
九個回合打贏，最後一頁是升階那一句接兩行賭金結算，錢包 214 → 239 與 442 → 432。
腳本從 2,153 步縮到 1,950 步。

**沒有為了讓它贏而改任何輸入**——換的只是清單上選誰。

### 順帶露出來的缺口

打贏那一頁印「receives 175 adventure points, and 313 gold crowns!」，但那兩個數字
**沒有寫回角色**：`AdventureReward()`／`GoldAward()` 目前只用來組訊息。錢包上看得到
的變化只有賭金。這是戰後結算共用的缺口，不是競技場獨有，已列進收據的 `not_claimed`
與 WORKLIST。

自評 85.84 → 85.98（角色成長 88% → 90%）。

## 2026-09-05（續三）敵人傷得夠重會撤退了

`sub_2374C` 的撤退分支要三個參戰者記錄的旗標，RE-081 當初標成「AI 讀它們當旗標，
但寫的人在別支」。寫的人確實在別支，但都在同一個執行檔裡讀得到：

| 位移 | 意思 | 誰寫 |
|---|---|---|
| +0x2D | 從上一次輪到自己之後挨的傷 | `sub_26D76:0x2762E`、`sub_2818D:0x285AF` |
| +0x2A | 上一手的 +0x2D | `0x20E64` 一手執行完的收尾搬過來 |
| +0x29 | 撤退狀態 | AI 自己：`0x2381D` 清 0、`0x238C1` 設 1 |

那個收尾同時把 **+0x0F（招架次數）也歸零**——兩格都是「這一手」的計數，不是整場的。
這一版兩格都從來沒有清過，攻擊分數因此一路被 `−Fatigue×20` 拖下去。

兩個門檻在執行期怪物記錄的 `+0x0D`／`+0x0E`。那份記錄開頭 32 byte 是從 DGROUP
`0xB966` 複製來的，而 `0xB966` 就是 `sub_1A4FE` 從 MONSTERS archive 讀進來的怪物
記錄本身——所以那兩格是原版資料檔裡名稱欄之後、金冠骰式之前那兩格。抽出來看：
**SKELETON、GOLEM、ZOMBIE、SALAMANDER、SYLPH、DJINN、MUCK THING 是 100／0**，
而傷勢是 `50 × 累積傷害 ÷ 昏迷門檻`，到不了 100——不死系與元素系不會逃跑。

### 改了戰鬥數值，收據全部要重跑

七份長路徑收據裡，`full-chain`、`full-chain-reload`、`flow-quest`、`quest-take`、
`quest-list` 重跑之後**逐 byte 不變**；`full-chain-ending` 從 1,012 步變成 1,147 步，
`arena` 從 1,950 步變成 2,489 步。變長的原因是敵人開始逃：`arena` 路上第二場伏擊
一個人就打了 1,288 步，追著逃跑的怪補刀本來就是原版的節奏。

`battle-ambush`／`battle-flee`／`battle-missile`／`battle-spoils` 那四份的畫格
**逐 byte 相同**——一擊就倒或還沒受傷的場面走不進撤退那一段。`battle-spell` 的
腳本沒有留下來，還沒驗。

### 三個把時間燒掉的坑

- **`full-chain-ending` 的 `-quest-sites` 只寫在散文裡。**收據的描述寫著「最終任務
  原本是十二隻 10 級 CYCLOPS…壓成一隻 1 級」，但 plan 裡沒有那個旗標，重跑時打的是
  原版那十二隻。它們站在三十格外，追了十九個回合還沒接觸。**plan 要帶齊重跑所需的
  每一個旗標**，散文不算。修法是加 `quest_sites` 欄位與結構化的 `substituted_inputs`，
  再寫一支 `tools/replay_player_path.py` 從 plan 一鍵重生。
- **改了 Go 卻用預編的執行檔跑了半天。**`run_kol_shots_fast.sh` 用的是
  `workplace/tmp/kol`，而那一份是改 AI 之前編的。產出全部「正常」，只是跑的是舊行為
  ——照它下了好幾個結論才發現。修法是那支腳本現在會比 mtime：執行檔比任何 `.go` 舊
  就拒絕跑並印出重編指令。**沉默的舊執行檔比錯誤訊息貴得多。**
- **`battle-*` 收據的重跑指令是散文。**「91 步的前段接八次…」照著組出來的腳本
  對不上任何一張畫格——括號裡的 token 數其實是 96，而真正的前段是另一條路徑的 90 步
  再加一個空白鍵。四份已經補上 `replay.spec` 進版控並重跑對過。

自評 85.98 → 86.48（戰鬥 90% → 92%）。

## 2026-09-05（續四）打贏的冒險點數與金冠會進帳了

`AdventureReward()`／`GoldAward()` 早就算得出來也印進訊息了，但兩個數字從來沒有
寫回角色——玩家看到「receives 60 adventure points, and 12 gold crowns!」，錢包與
資料頁上什麼都沒變。發放在 `0x2141C` 的迴圈：隊伍那六格裡有人的每一名各加一份，
兩個數字本來就是「每人分到多少」。

同一段還有一個**閘門**：`0x2124E` 比 `word_345DE == 0xFFFF`，不成立就整段跳過
（RE-096）。也就是**競技場與訓練戰不發冒險點數也不發金冠**——它給的是位階。
`State.RankFight` 對應那個旗標；`arena` 收據裡升階那一句因此直接接在打贏後面，
中間少了一頁獎勵。

冒險點數跟著進了存檔（`adventure_points`，選填、指標）：它打贏會加、訓練會扣
（RE-076），不存就重載歸零。`full-chain` 的存檔比對加了兩欄——Poontang 從角色檔
的 2,316 變成 2,376，那一場每人 60。

`arena` 的錢包最後疊了三件事：路上兩場伏擊的金冠每人 ＋12、押中的 Poontang 再
＋25、押錯的 Fighting Bull −10。七份長路徑收據全部重跑通過。

自評 86.48 → 86.98（戰鬥 92% → 94%）。

## 2026-09-05（續五）敵人被打時會招架了

`0x23CD6`…`0x23D8B` 整段讀出來才發現這一版少了一大塊：貼近打那一手原版會填三個
選項，而 remake 只填了攻擊式。**`word_30916` 寫進 `+0x24`，而那一格在這一名被攻擊
時就是防禦選項**——敵人選的「命中部位 ＋1」同時就是它的招架姿態。這一版那一格
幾乎都是 0，所以敵人被打時從來不招架（`sub_26D76` 看 `d.Guard != 0` 才算招架次數）。

算式有兩個容易看漏的地方：

- **抽命中部位的列號是「進來時」的攻擊選項**，不是剛抽出來的攻擊式。環 0 那條
  進來時是 `0x0B`（`0x23C81` 同時填防禦選項 7），反制掃描那條是從對手抄來的。
- 攻擊式退回 1、或進來時的攻擊選項是 0，都走另外兩條短路，不抽部位。

`word_34A08`（移動段繞完八個方向都站不下的次數）非零時攻擊式固定 1。這一版沒有
那條重試路徑，列成已知差異。

### 三條收據重生

敵人開始招架，含戰鬥的長路徑全部作廢：`full-chain` 567 → 764 步（那一場從四個
回合變六個）、`full-chain-ending` 1,147 → 1,182 步、`arena` 2,489 → 2,583 步。
`flow-quest`、`full-chain-reload`、`quest-list`、`quest-take` 與 `battle-*` 四份
逐 byte 不變。

重生用的驅動器從 `workplace/` 搬進版控，成了 `tools/drive_battle_script.py`：
暫存目錄改由 `KOL_DRIVE_DIR` 指定，所以三條路徑可以同時跑。走路與城裡那幾段沿用
原腳本的前後綴，只換中間戰鬥那一段。

自評 86.98 → 87.23（戰鬥 94% → 95%）。

## 2026-09-05（續六）反制掃描與側翼

`sub_253E7` 與 `sub_25653` 從 `enemy-ai.json` 逐條讀完。RE-104 當初把反制掃描標成
「部分解出」，因為那張優先度表看起來就是五條互斥的規則；再讀一次才看到**「已規劃」
那幾條之後還有一個共同出口**（`0x2553C`）把優先度整個重算：指令是移動給 1、有防禦
選項給 2、其餘給 4，而且算出來的分數 `>=` 目前最高分就覆蓋。所以表上的 5 與 3 不是
最終分數——5 永遠留得住（共同出口最高只到 4），3 只在共同出口算出 1 或 2 時留得住。
**一張看起來自洽的規則表，漏掉共同出口就會做出「照抄卻不對」的行為。**

側翼那一支只改方向、不改位置：從自己現在的方向往上數，方向編號是奇數就走 2 格、
掃到第 5 格，偶數走 1 格、掃到第 6 格；數到對方的方向時就知道自己站在它的哪一側，
`k == 4`（正對面）放棄，其餘把對方的方向偏一個步幅，再看那一格有沒有磚塊或自己人。
**原版把加完的方向夾在 `> 8` 而不是 `> 7`**（`0x256F7`），偏出 8 會落在 0…7 之外；
這一版照樣算，越界當成擋住直接放棄。

前面那個智力擲骰（`sub_252CB`：`值 × 等級 ÷ 3 > rand(100)`）原本列在 SPEC-008 的
「不授權」裡，理由寫的是「沒有反制掃描就用不到」——現在用得到了。

篩選遮罩 0x10 要的 `+0x05` 仍未解：它是 `sub_21FCD` 依裝備賦予的四格之一，而這一版
沒有任何地方設它，所以那個條件恆真。列成已知缺口，不拿它擋住整段實作。

### 三條長路徑重生

`full-chain` 764 → 938 步（那一場從六個回合變八個）、`full-chain-ending`
1,182 → 1,003 步、`arena` 2,583 → 1,770 步。變化的方向不一致——同一批改動讓
`full-chain` 那一場變長、另外兩條變短。`flow-quest`、`full-chain-reload`、
`quest-list`、`quest-take` 逐 byte 不變。

### 短收據也有重跑工具了

`docs/verification/battle-*/` 那五份原本只能照收據 `command` 欄的散文重跑。
新增 `tools/replay_battle_receipts.py`：跑各自的 `replay.spec`，再拿輸出的 PNG 對
收據 `frames` 的 SHA-256 與版控裡那一張。五份的結果：

| 收據 | 這一輪 |
|---|---|
| `battle-ambush` | 腳本與三張畫格都不變 |
| `battle-missile` | 腳本不變，`blow-hits.png` 換掉（命中部位從右臂變成胸口）|
| `battle-flee` | 腳本 107 → 102 個鍵（逃走之後推回野外少按五個鍵），兩張畫格不變 |
| `battle-spoils` | 腳本 235 → 106 個鍵（原本備了八輪，現在一輪多就打完），三張畫格不變 |
| `battle-spell` | **原本沒有 `replay.spec`**，這一輪從頭重建（117 個鍵）；四張裡兩張逐 byte 不變，`casts.png` 與 `effect.png` 換掉（法術強度的形容詞從 mildly 變成 markedly）|

`battle-spell` 那一份是這一輪才發現的洞：它的重跑指令從來只寫在散文裡，所以改了
戰鬥數值之後根本沒有東西可以重跑，也就一直沒有人發現它該不該重生。**收據要能重跑，
指令就得是機器讀得到的檔案**——`full-chain-ending` 的 `-quest-sites` 是同一個教訓。

自評 87.23 → 87.42（戰鬥 95% → 96%、角色成長 90% → 92%）。

## 2026-09-05（續七）遭遇的四行敘述與騎走那一問

`sub_2A46E` 與 `sub_2A437` 從整份反組譯讀出來，RE-097 對它們的標註是錯的：不是
「畫那一隻與收掉」，而是「畫任務地點那個框」與「印四行敘述」。四行的來源是
DGROUP `0xBB26`——**那就是怪物記錄的 `+0x1C0`**：一筆記錄 512 byte 讀進 `0xB966`，
末端 64 byte 正好是四格文字，格式與城鎮對話完全相同（`sub_1B94F` 的 mode 0，
解碼器 RE-064 早就寫好了）。三十二隻各有一段遇到牠時的敘述，例如 BINDERAK 的
`You notice the spiders are awfully large in this area.  Suddenly, a form bolts
towards you.`

**這一段原本一直看不到**：那個框只在「要不要騎走」那一頁上出現，而問這一句的條件是
全隊都有坐騎（`word_2FA2C != 20`）。原版名冊裡的角色移動率全是 0（用走的），
買坐騎的馬廄商品表又還沒解，所以正常玩法到不了——收據因此換過名冊。

**找到它的方式值得記下來**：`0xBB26` 在整份反組譯裡只出現一次（一個 `mov ax, 0BB26h`），
沒有任何 writer。`tools/read_dgroup_table.py` 讀出它落在 BSS，所以那不是靜態表；
接著才想到 `0xB966` 是怪物記錄載入的位置，而 `0xBB26 − 0xB966 = 0x1C0`，
記錄長度 512 減 64 正好對上。**「找不到 writer」有時不是掃描有洞，是那塊記憶體
根本是別人載進來的資料。**

同一段還改正了 RE-097 的第二處：怪物編號 0xFF 那一條當初叫「任務怪」，記錄 0xE6
實際是 **`Will you be healed?`**——那是治療的遭遇，低位階隊伍才擲得到，答應之後把
六名角色開場帶的傷（`+0x1B`）清零。那一條的敘述走另一個緩衝區（`0x2350`），
來源還沒解，所以仍然不實作。

順手修掉一個一直都在的差異：**那一頁蓋著地圖時方向鍵還走得動**。原版那兩處等的是
`sub_144EF`／`sub_144C3`，答完之前收不進任何輸入。任務地點那一頁有同樣的問題，
只是既有腳本都是「跳出來就馬上回答」，所以一直沒露出來。

### 收據與命名

新增 `docs/verification/battle-rideaway/`（觀察，名冊是換過的）：四行敘述加問句、
答是騎不掉、按一個鍵才開打，以及另一顆種子下答是就跑掉。`replay_battle_receipts.py`
因此支援一份規格檔有好幾行（同一段路的不同分支），畫格用 `spec_line` 指定自己是
哪一行跑出來的。

`world.QuestPrompt` 改名成 `world.Prompt` 並加上 `PromptKind`（任務地點／騎走／
只等一個鍵）：同一個框現在有三種用途，名字沿用「任務」會誤導。

自評 87.42 → 87.51（野外 95% → 96%）。

## 2026-09-05（續八）撿之前會問了

原版對隊員撿東西之前會先印記錄 0x111（`You have found a|H.  Do you want to pick it
up?`）停下來等 y／n（`0x20354` 的 `sub_1B100(5)`），敵人才直接撿。這一版一直是直接撿
——那個問句要在**執行段中間**插一次互動，而執行段原本只會一手一句往下推。

做法是讓 `act` 在那一手停住：填 `State.PickUp`、把 `Plan.Prompt` 換成 `PromptPickUp`，
答完之後把要補印的那一句放進 `State.resume`，下一次取手時先交出去。答否會**繼續問
同一格上的下一件**（`0x20501` 回到同一個迴圈），答是則搬過去——手上已經有東西就印
0x112，撿到任務要的那一件就印 0x113（一場只喊一次），**撿成功本身不印任何一句**。

### 三句都不加動作者的名字

問句自己就是一句完整的話，0x112／0x113 的開頭本來就是 `|G`。原版對它們另起一段訊息
（`sub_1C1A4`），不接在動作者的名字後面。`ActResult.Plain` 就是這一條——之前 0x112
走的是加名字那一條，印出來會是兩次名字，只是沒有收據蓋到。

### 一個實跑才看得見的洞

`Plan.Env`（`CommandEnv`）**從來沒有人填**。地上有東西時該出現的「撿」那一格因此
在遊戲裡從來沒有出現過，而每個單元測試都自己傳一份 `CommandEnv`，一路都是綠的。
補上 `State.CommandEnv(i)` 之後才走得到這一手。**「呼叫端要填的欄位沒有人填」
這一類洞，型別系統與單元測試都擋不住**——只有把整條路走一遍才看得見。

### 順手解掉 RE-103 的一個懸案

`sub_2886C`（手上那件是不是任務要的那一件）比的是 `[bx-46AAh]`，也就是 DGROUP
**`0xB956`**；RE-103 記成 `0xB926`，並且寫著「那十六個 byte 由誰填還沒解」。
位址算對之後 writer 一眼就看得到：`sub_2B0FF:0x2B1C4` 從任務 payload 的 `+0x29`
複製過來。**掃不到 writer 時先把位址重算一次**，再去懷疑「整塊複製看不到」。

同一條線還多帶出一件事：開打時原版會把那十六個 byte 再複製一份到**戰場物品的第 0
格**（`0x1E68C`）——任務要的那一件一開場就躺在戰場上，撿起來再跑到邊上就算完成。
這一版還沒做，列成下一個工作。

### 收據

新增 `docs/verification/battle-pickup/`（沒有換過輸入）：第一回合丟長矛、第二回合
指令列多出撿那一格、執行段停在問句上，答是換回帶武器那一組、答否東西留在地上。
兩行規格檔跑同一段路的兩種回答。

### 任務要的那一件東西是什麼，解出來了

那十六個 byte 的來源既然是任務地點 payload 的 `+0x29`，就直接讀得出來——二十四個
地點各一件：Gavel Of Oak、Village Flag、White Quill、Truth Sword、Shackles、
Kelder Crown、Ship Wheel、Coat Of Arms、Flask Of Oil、Silk Cloak、Stone Mallet、
Blue Vial、Serpent Wand、Gold Chalice、Pirate Hat、Hidden Staff、Parth Flask、
Ruby Choker、Pirate Chest、Statuette、Gold Bracer、Teddy Bear、Shade Ring、
Sheller Ward。`quest-sites.json` 多一個 `quest_item` 欄，`NewQuest` 拿它填
`State.QuestItem`——RE-103 那條「帶著它站到邊上就算完成」這才真的比得出東西。

**還缺的是它一開場放在戰場哪一格。**站位表就在它前面（`+3` 起 12＋12＋6＋6 到
`+0x27` 為止），所以 `+0x27`／`+0x28` 是緊接著的兩個 byte，二十四個地點的值都落在
5…56（可走區是 4…0x3B），`sub_2B0FF:0x2B1F0` 又把它們抄進兩個全域，其中一個在
競技場那一支被設成 `0x80`——與戰場物品那張 X 表的「不在場」哨兵相同。看起來就是
座標，但**讀那兩個全域的地方還沒找到**，所以資料裡先原樣留成 `hint_bytes`，不放。

### 順帶結掉一條「不授權」

SPEC-011 一直把遭遇圖列在不授權裡，理由是「`[0x1370 + 怪物 × 2]` 取的圖偏移還沒
對上」。查下去才發現**那個理由從一開始就抄錯了**：RE-010 早就把 `0x1370` 解成
**尺寸表**（8-pixel 寬度單位與 scanline 高度），不是圖號；`sub_2A975` 畫的就是
`MONPICS[怪物編號]`，沒有中間那一層。三十二筆尺寸與抽出來的 PNG 逐筆相符
（8×47 → 64×47、…、11×64 → 88×64），`0x13B0` 與 `MonsterSpots` 也逐 byte 相同。

**不授權清單上的理由也會過期，而且過期的方式不是「後來解出來了」，是「當初就抄錯了
隔壁那條 RE 的結論」。清單上的每一條都要指得出證據，指不出就是還沒查，不是還沒解。**

自評 87.51 → 87.76（戰鬥 96% → 97%）。

## 2026-09-05（續九）打字那一格的版面

### 「還沒讀」的其實早就讀過了

WORKLIST 上「MOUTH 輸入框的游標（`sub_1B5C8` 還沒讀）」這一條，讀下去第一件事就
發現 **RE-076 已經整段解過**——訓練那一格「要練幾點」用的是同一支函式，游標的畫法、
反相週期與 `sub_1D01D` 的位址算式當時都寫進 RESEARCH-LOG 了，`internal/town` 也已經
有 `TextCaret` 與 `CaretRow` 在畫那條橫槓。真正沒做的是**版面對拍**，不是「還沒讀」。

**這是第二次踩到同一個坑**：清單上的理由會過期，而過期的方式是「當初寫下的那一刻
就不準」或「別處解完了沒有回頭改」。這一條原本會讓人重讀一次三百行反組譯。

### 三處文字欄，兩處畫錯

同一支 `sub_1B5C8` 在遊戲裡有四個呼叫點（MOUTH、賭場金額、訓練點數、有框的那一格），
共用同一個緩衝區 `0x2E3A`。remake 只有訓練那一格用 `TextCaret`，MOUTH 與賭場金額
各自寫死一個字面底線 `"_"`——**同一個原版元件在 remake 裡有兩種畫法**，而且沒有任何
測試會抓到，因為兩處都沒有對拍。三處統一成同一個哨符。

### 原版收據與對拍

`dosbox_character_ui_oracle.py` 新增情境 `town-mouth`：進旅店 → 白框走到 MOUTH →
逐鍵打 `elve` → 退格一次 → 補回去 → 送出。19 個檢查點，68 個畫格。

量出來的版面：第 0 格左緣 x = 104（`word_2DF14 = 0x0D`，一欄 8 px），字模在
y = 120…124，橫槓在 y = 125、寬 8、高 1。與 `sub_1D01D` 的
`A000:((列 + 5 + 2 × word_2DF18) × 0x28 + 行)` 逐項對得上。

`plan-town-mouth.json` 兩對畫面，允許差異區之外**零像素差**。

**閃爍的相位會咬人**：橫槓每 5 個 timer tick 翻一次，同一條腳本跑出來的畫格
亮暗都有（058／062 亮，059／063／065 暗）。拿暗的去比會差出正好 8 個像素——
看起來像版面錯了，其實只是取樣落在另一個相位。對拍只挑亮的那幾張。

順帶量到一件本來沒在找的事：**退格是完全可逆的**。「打了 elve」→ 退格 →
「補回 elve」三張畫格的 SHA-256 兩兩相同，畫面上連一個 byte 的殘留都沒有。

### 圖示列在打字期間不收

按下 MOUTH 前後（畫格 056 → 058）的差異外接矩形只有 `(104, 86, 191, 40)`，
底下那排圖示與白框原樣留著。remake 原本在 `StartAsk` 把它們收掉了——這是**只有
逐像素對拍才看得見的差異**，玩起來不會有人發現，單元測試也不會失敗。

### 驗證

`go vet ./...`、`go test ./...` 全過；二十六份對拍收據 285 對全過；
七份長路徑收據（`replay_player_path.py --all`）與七份戰鬥收據全過。
`quest-take` 與 `arena` 的畫格因為游標與圖示列改動而重生。

自評 87.76 → 87.94（城鎮 91% → 92%）。

## 2026-09-05（續十）清單元件的參數語意

### 一支函式，十三個呼叫點

`sub_1AAE8` 在招募、商店、訓練、任務一覽、戰利品與領取者六個地方都在用，
先前三份 RE 各記了一次呼叫參數的值，但沒有人解語意，所以戰利品那張清單的座標
是「接在標題下面一行」猜的。整支讀完之後七個參數全部對上（RE-115）。

正對照現成就有：招募與商店的呼叫是 `(1, 0x0D, 0x56, 8, n, 緩衝區, 0x18)`，
照解出來的語意算是 x = 104、y = 86、寬 24 欄、一次 8 列、捲動箭頭在第 36 欄，
而 remake 的 `DialogBox`、`ListRows` 與 `ListColumn` 逐項相同——那兩張清單的
對拍本來就是零像素差。**先有一組已經對拍過的呼叫，再去解語意，等於自帶驗證。**

### 選中那一列是換色，不是框

`sub_1B033(顏色, 列, 緩衝區)` 從 `[0x668A + 列 × 2]` 取那一列在緩衝區的起點，
把整列重畫一次——選中傳 0x0F、離開時傳 3 還原。畫字元本來就會先清掉那 8×8 格
（RE-114），所以重畫等於換色。

remake 的戰利品清單原本畫一個白框，改成換色；`SpoilList` 從猜的
`Col: 2, Row: 0x11` 改成 `Col: 1, Row: 0x13`。**招募與商店那兩張本來就是換色**
——同一支原版元件在 remake 裡又出現了兩種畫法，與昨天的橫槓是同一個模式。

### 還沒接上的

清單底下那條圖示列（`sub_11703(7)` ＋ `sub_11703(0x0F)` ＝ YES ＋ U-TURN）。
原版收據裡沒有戰利品畫面——要打贏一場而且擲中兩成的掉落——所以這一整塊到現在
都只有靜態證據，`battle-spoils` 的 `not_claimed` 照實寫。

### 驗證

`go vet ./...`、`go test ./...` 全過；七份戰鬥收據全過（`battle-spoils` 的
`items-found.png` 因版面改動重生，其餘六份逐 byte 相同）。

## 2026-09-05（續十一）戰利品的圖示列與空手的怪物

### 圖示列與執行段的卷軸是同一塊

接戰利品那條圖示列（YES ＋ U-TURN）時發現：原版清圖示列用的是
`sub_1103F(2)`，範圍 `(1, 0xA4)–(0x26, 0xC4)`——換算是 x = 8…304、y = 164…196，
**與執行段那張卷軸的 `MessageBox` (8, 0xA4) 296×32 完全一樣**。所以清單一出來，
報獎勵的那幾句就被蓋掉了；第一版只蓋了圖示列自己那 22 條掃描線，卷軸的上半留在
畫面上，一眼就看得出來不對。

**兩個元件共用同一塊畫布時，「清掉」的範圍要從清的那一支去讀，不是從畫的那一支猜。**

### 空手的怪物：兩層都走不到

決策樹第 5 條「手上那一格空著就去撿」讀完之後，發現它有兩層走不到：

1. `[參戰者 + 0x2B]` 在整份反組譯裡只有兩個寫入點，**兩個都寫 0**，
   所以「非零就改成移動」那一條永遠不成立。
2. 再往外一層：`monsters.json` 的**三十二隻怪全部都有武器**，
   所以「手上那一格空著」本身在原版資料下就不成立。

照樣實作了，理由寫進 RE-116：決策樹要能與原版逐條對照，走不到的分支也要在。
**副作用是它不影響任何既有收據**——長路徑打的是原版名冊裡的怪，`Unarmed` 一路是 false。

順帶修正 RE-081 的一處：那一條寫成「怪物記錄 [+0x20]」，其實 `+0x20` 讀的是
`word_39A48`（隊員 768 byte 的角色記錄／怪物 128 byte 的怪物記錄），
`+0x2B` 才是參戰者記錄（`0x6712 + 0x32 × i`，50 byte）。**兩個偏移分屬兩份記錄。**

### 順手修掉一個「每次重跑都變」的收據欄位

`full-chain` 與 `town-rest` 的收據都記了寫出來那份存檔的 SHA-256，而存檔裡有
`saved_at`——**那個值每跑一次就換一個**。它驗不了任何東西（重生不出同一個值），
只會在每次重跑時多一行 diff。兩支驗證器改成算「拿掉 `saved_at` 之後的正規化
雜湊」，欄位名跟著改成 `sha256_without_saved_at`，兩份收據重生過。

驗證方式是**連跑兩次**：第二次跑完 `full-chain` 的收據與第一次逐 byte 相同。
（先前每一次重跑都會改那一行，只是沒有人把它當成問題。）

### 驗證

`go vet ./...`、`go test ./...` 全過（新增四個測試釘住這兩段）；
七份戰鬥收據與七份長路徑收據全過，`battle-spoils` 的 `items-found.png`
因圖示列而重生，`town-rest` 的截圖十一張逐 byte 沒變。

自評 87.94 → 88.19（戰鬥 97% → 98%）。

## 2026-09-05（續十二）撿到的東西會用自己的骰，移動段解出來但沒接上

### 地上那一格只留名字，撿起來就打不對

RE-116 做完之後留了一句「撿到的東西還是照空手的數字打」。追下去發現**隊員那一側
也有同一個洞**：`TakeItem` 只填 `Member.WeaponName`，`Actor.Weapon` 沒動。原版是
整筆 16 byte 搬進 `+0x20`／`+0x30`，傷害骰、負重與種類本來就在那十六個 byte 裡。

`GroundItem` 加一格 `Weapon`：丟出去時存進去、撿起來時搬回 `Actor.Weapon`，
掉落物由 `SpoilItem` 內嵌的 `Item` 換算。`SpoilItem` 原本只有 index/name/empty，
改成內嵌 `Item`——`ITEM.LST` 的 JSON 本來就是平的。

### 移動段有三層退路，不是「朝目標走一步」

RE-081 把移動段記成「朝目標走一步」。實際上（RE-117）：

```
有目標   往目標那一步 → 被擋就試左一格 → 再試右一格 → 都不通就往下
沒有目標 從 rand(8) − 1 起繞八格，取第一個站得下的
八格都不行 傷勢 < 撤退門檻 → 原地揮武器
           否則 1/3       → 回撤退段
           否則           → word_34A08 加一，跳回「手上有沒有武器」那個分岔重跑
```

而 `word_34A08` 非零之後，貼近打的攻擊式**不再抽，固定第 1 式**。

決策樹因此要能重入：從那個分岔起抽成 `planTactics`，重跑帶一個 depth。原版沒有
重試上限（靠 1/3 那一擲跳出去），remake 加了一道，免得亂數注入器餵固定值時打轉。

### 換成隨機起點之後，一個測試露餡了

`TestEnemyClosesInThenAttacks` 原本用 `&Field{Tiles: make(...)}`——磚 0 走得過去，
**卻擋視線**（`sightBits[0]` 是 0xFF）。那一場敵人根本看不見目標，走的是「沒有
目標」那一條；而舊的 `stepToward(i, -1)` 固定從方向 0 掃起，方向 0 正好朝著目標，
所以測試會過。**測試驗的不是它以為在驗的東西。**換成隨機起點就露餡了，地形改用
`NewField()`。

### 移動段接上去會讓兩條長路徑作廢，所以退回來了

七份戰鬥收據都過，`full-chain` 也過——但 `arena` 與 `full-chain-ending` 從
**第 650 步**開始整條走偏：那是同一場戰鬥，敵人的走法一變，玩家那一側的固定按鍵
序列就對不上了。

用 `drive_battle_script.py` 重生 `arena` 的第一場，跑了約五十分鐘、從 369 個
token 推到 988 個（原本那一場是 573 個 token，現在 911 個）。**然後才發現規模
不對**：`arena` 一共三場（兩場伏擊 ＋ 競技場），中間夾著進場、下注那幾段城鎮
互動，driver 只生得出戰鬥那一段；`full-chain-ending` 是兩場戰鬥再接七頁結局。
兩條都得手工接後綴，接完各要跑一輪十分鐘才知道對不對。

**把移動段的實作退回去**，RE-117 的記錄留著並標成「解出來了，還沒接上」。
留半條走不通的收據比沒接上更糟；腳本重生值得一輪專門處理。

**留下來的兩件**：`GroundItem` 帶武器數字（不碰 AI），以及那個測試的洞——
`TestEnemyClosesInThenAttacks` 用的全 0 地形**擋視線**，敵人根本看不見目標，
舊實作固定從方向 0 掃起才剛好朝著目標。**它驗的不是它以為在驗的東西**，
地形改用 `NewField()`。

### 順帶結掉 `+0x05`

RE-081 把反制掃描的篩選遮罩 0x10 記成「要的 `+0x05` 還沒解，這一版沒有設它，
條件恆真」。追下去發現**原版也恆為零**（RE-118）：那一格只由穿戴欄部位
3／5／7／8 那件東西種類 byte 的最低兩位元決定，而 `ITEM.LST` 裡放得進那四格的
三件（Fur Boots、Fur Cape、Leather Cape）種類 byte 都是 0；`kind = 1` 的物品
全是武器，而武器不在那四格。

**「這一版沒有設它，所以條件恆真」要往前多走一步**——原版有沒有設它是另一個
問題，而答案可能是「也沒有」。掛在待辦上等於替原版發明了一個它沒有的功能。

### 驗證

`go vet ./...`、`go test ./...` 全過；七份戰鬥收據與七份長路徑收據全過。

## 2026-09-05（續十三）馬廄賣什麼、多少錢

`RE-075` 讀出「馬廄的表一筆 0x80 byte、名稱在 +0x70」就停了，抽取器也只抽了名字
與整筆 hex。這一輪把欄位讀完（RE-119）：第 0 個 byte 是 `STEEDPIC` 的圖號、
`+0x78`／`+0x79` 是大端的價錢。

**價錢自己就是正對照**：五間馬廄的表完全相同，四匹是 200／400／800／1600 的
倍增數列。欄位若取反或錯位，讀出來會是 0xC800、0x9001 那種散亂的值。

### 買下之後那一段也讀完了

錢包扣掉價錢，把馬廄那一筆的 `+0x70` 起**十六個 byte** 抄進角色記錄 `+0x1B0`，
再把 `STEEDPIC` 那張圖搬進角色圖像區（一槽 0x2F8 byte）。折抵讀的
`[角色 +0x1B8]`／`[+0x1B9]` 正是那十六個 byte 的第 8、9 格——**以舊換新**，
折到不足 50 就收 50。

### 交叉檢查：原版名冊裡沒有人有坐騎

十六槽的 `+0x1B0`…`+0x1BF` 全是零。這同時證實那十六個 byte 的空判準（全零）
與「馬廄是取得坐騎的唯一入口」。**「全隊都有坐騎才問得出來」那一頁（RE-112）
在正常玩法上到不了，不是因為那一頁沒做，是因為前面那一步還沒接上。**

`tools/extract_town_buildings.py` 的馬廄分支補上 `picture` 與 `price`；
買賣那一段（TRADE 的 bit5 分岔、扣錢、寫進角色、STEEDPIC 的圖）還沒接。

## 2026-09-06 攻防技能的罰值，RE-122 兩個洞填完

`tools/ida_probe_koltown_skill_penalty.py`（新增），輸出
`workplace/analysis/koltown/skill-penalty-private.json`。

RE-122 把效果 7／8 推到「把某兩個全域各減十倍」就停住，因為一直盯著**寫入點**看。
換成把**取用點**倒出來就結束了：`sub_26AC2` 的三個呼叫點全在攻擊結算裡，
`sub_26D76:0x26E1C` 把 `word_34A12` 加上力量——那就是攻擊分數的第一項。
於是 `word_34A12` 是攻擊技能、`word_32380` 是防禦技能，參戰者記錄 `+0x2E`／`+0x2F`
是它們的罰值（一點抵十）。與 `MAGIC.0` 屬性表第 7、8 項 `Off. Skill`／`Def. Skill`
一一對上。

**全庫掃描做了正對照**：直接以 `[reg+2Eh]`／`[reg+2Fh]` 觸碰記錄的指令只有
`sub_26CD5` 那兩條讀取，兩條都出現在結果裡，所以「沒有第二個寫入點」是掃到的、
不是掃不到。法術寫進去走的是 `0x287E2` 那條泛用 `mov [bx], al`，字面上看不到位移。

### 兩件跟著定下來的事

- **原版沒有為罰值那兩格把正負反過來**（`0x2878E` 對六個欄位一視同仁），
  而效果 5 的疲勞（`0x286AB`）有反。照原版跑，「正面」的法術會扣掉對方的攻防技能，
  「負面」的反而是把罰值減回來。remake 照抄，理由與證據都寫進 SPEC-012 第 10 節。
- **罰值不隨法術過期還回來**：`sub_2198A`／`sub_2215A` 只寫 `+0x07`…`+0x0C`。

### 疲勞的分母

`sub_21AC4` 的兩條寫入是「力量 ＋ 本體記錄的某一格」，那一格在同一支的下一段又拿去
算體力總量，所以是**健康**。RE-122 擔心的「`+0x20` 在怪物那一側是手上的武器」是把
兩份記錄搞混了——同一個位移在怪物記錄裡是武器、在角色記錄裡是健康。

### remake

`Actor` 加 `SkillPenaltyOff`／`SkillPenaltyDef`；`OffenceSkill()`／`DefenceSkill()`
折算罰值，`AttackValue()`／`DefenceValue()` 改用它們；`castStat` 的效果 7、8 不再
什麼都不做。新測試三支（欄位、折算、正負方向與過期不還）。八份長路徑收據與戰鬥
收據全部重跑——罰值起始是 0，既有結果不受影響。

## 2026-09-06（續一）疲勞不是傷勢——修掉戰鬥數學的一個對錯

`tools/ida_probe_koltown_fatigue_source.py` 與 `..._fatigue_readers.py`（新增）。

RE-129 定下 `sub_23683` 的分母是「力量 ＋ 健康」之後，那條式子就與 remake 對不
起來了：remake 把它實作成 `50 × 累積傷害 ÷ 昏迷門檻`，而昏迷門檻是
`(體型 ＋ 健康) ÷ 2`，與分母不是同一個東西。分母既然不同，分子也得重查。

**判準是誰在寫它。** 全庫掃 `[bx+322Eh]`：只有開打歸零、法術效果 5、休息 ×7÷10、
每一手 −= 額度÷10。**傷害落到身上那一支 `sub_277F4` 不在名單裡**（它寫 `+0x0E`）。
所以那一格是 `MAGIC.0` 屬性表第 5 項的 Fatigue，與累積傷害是兩個獨立的累計。
正對照是 `sub_23683` 的讀取與 `sub_21AC4` 的歸零，兩條都在名單裡。

再掃 `word_2FB4A` 的讀取點（每一條帶前後文，一次問完）得到六處，remake 六處都有
對應實作、位置全對、值全錯：規劃鍵值、執行鍵值、攻擊值、防禦值、力量加成、
敵人 AI 的撤退門檻。等於替原版加了一條「傷得越重打得越差」的乘法惩罰。

### remake

`Actor` 加 `Fatigue`（`[0x322E+2i]`）與 `Stamina`（力量 ＋ 健康）；`+0x2D` 那一格
改名 `Hurt`（它是「這一手挨的傷」，與疲勞無關）；`Injury()` 換成 `FatigueRatio()`。
法術效果 5 接上（**正負是反過來的**：負面加、正面減；原版下限那一段是死碼，
所以減得到負數），休息一手回復三成也接上。每一手的衰減沒接——它要整套時間額度。

### 收據

`full-chain` 那一場少打了兩回合（8 → 9 回合才結束，步號整條位移），按鍵序列失效，
用 `regen_long_path.py` 重生戰鬥段（957 → 1067 個 token），里程碑與兩張檢查點圖
搬到步 921 與 1066，五張過期截圖刪掉。其餘七份長路徑與七份戰鬥收據不受影響。
十五份全部重跑通過。

順帶：`run_driver.sh` 的「執行檔比原始碼舊就拒跑」那道保險擋下了一次——它比對的是
`*.go` 全部，包含測試檔，所以改測試也會觸發。這一次擋得對（本體真的改過）。

## 2026-09-06（續二）時間額度接得起來，力竭會放倒人

`tools/ida_probe_koltown_time_allowance.py`（新增）。

RE-130 收尾時判斷「每一手的疲勞衰減要整套時間額度才接得起來」，把它擱著。
**那個判斷太保守。** 把 `word_30D06` 從算出來到被讀掉那一整段倒出來就看清楚了：
它不是行動點數，沒有任何地方拿它當「還能不能動」的閘門，只是一個每一手結算一次的
計量，唯一的去處是疲勞與敵人 AI 的權重。所以接得起來，而且**非接不可**——
沒有它，疲勞一整場都是 0，RE-081 那一整套撤退判斷是死碼。

三個扣款點：挨的傷 × 5、指令成本 × 速度 ＋ 姿態乘數 × 負重、施法成本。四張表一次
倒出來，邊界互相接得起來（`0x10C6` 那一張正好是 RE-081 已知的 `01 01 02 04 08
10 20 40`，拿它當正對照）。騎著的人額度另外折半（為負則加倍），然後
`疲勞 −= 額度 ÷ 10`。

**額度為負時疲勞上升**，而額度裡有 `− 累積傷害 × 10`，疲勞本身又再扣額度——
「受傷 → 疲勞 → 撤退」那條迴路就是這樣閉合的。

### 力竭不是印一句話

`word_2FB4A ≥ 100` 那一條（`0x20D9A`）除了印訊息，還把 `+0x11` 與 `+0x13`
（頭與胸）都設成 0xFF——**當場倒下**。三句話是 0x11C `collapses from
exhaustion!`（隊員）、0x11D `dies from blood loss!`（怪物）、0x11E `shatters into
pieces!`（本體記錄 `+0x10` 的 bit 7；三十二隻裡只有 SKELETON 是這一種）。
`extract_monsters.py` 跟著補抽那一格。

### 沒接的一項

`sub_251B0` 在「`+0x22` 非零而且這一手是徒手」時速度那一項改從本體記錄取一格。
`+0x22` 的語意還沒解，所以那一條沒實作。

### 一個工具上的觀察

背景任務跑到一半被中止過三次（重生與收據重跑各一次），改用 `nohup` 分離啟動就不
受影響。長工作要能獨立於呼叫端存活。

## 2026-09-06（續三）SPEC-008 那個「不授權實作」也關掉了

RE-131 解出 `word_30D06` 與 `word_2F940` 之後，SPEC-008 列為不授權的
「部位權重對索引 < 2 或 == 6 的三項要乘 `word_2F940 ÷ word_30D06`」就沒有理由擋著了
——當初擋的理由正是那兩個全域語意未解。

```
sub_251B0(a, b):  word_2F940 = a × 武器速度 + b × 負重
sub_2526C():      把「額度 ÷ 花費」夾進 1/3…1
sub_2578F（攻擊式）：sub_251B0(15, 1)，最後一項 × 額度 ÷ 花費        ; 只會壓低
sub_2590F（部位）  ：先把額度設回完整那一份，sub_251B0([0x10B2 + 式], 2)，
                    索引 0、1、6 各 × 花費 ÷ 額度                    ; 1…3 倍
```

`EnemyStyle`／`EnemyLocation` 的簽章因此改成收整個 `Actor`（要武器速度、負重與
額度）。十五份收據重跑全過，而且**這一次沒有一條需要重生**——權重只是機率上的偏移，
既有的按鍵序列還走得通。

還留著的一項：`sub_251B0` 在「這一回合規劃過了而且這一手是徒手」時速度那一項改從
角色記錄取徒手那一式的物品記錄。remake 的 typed data 沒有那四筆，要先抽出來。

## 2026-09-06（續四）選部位那一次互動

使用者決定回到原版。原版對「不用選目標格」的法術會在執行段中間問玩家挑一個部位
（`0x2845E`），remake 一直走的是原版另一條（隨機部位）。

實作要點是**擲骰的位置**：原版的順序是持續（`0x283B2`）、強度（`0x28420`）、
問（`0x2845E`），所以骰要在問之前擲完。`Cast` 因此多一個 `AreaAsk` 哨兵——擲完就
停下來、效果留到 `AnswerArea` 才套。新的提示層 `PromptArea` 與撿那一問同一種形狀：
執行段中間、只認 `1`…`6`、答完才接效果那一句。GUI、腳本 token 與訊息卷軸三條路
都接上了。

目標的人形沒畫：`sub_27E20` 的圖庫與幾何只到 `強推論`，而戰鬥左下角那個框本來就
整塊沒有內容。

### 還缺實跑收據，理由寫進 SPEC-012

十六名原版角色裡只有 Warder 有法術，而那一筆走不到這一問。要走到得先買一筆效果 6
而且不用選目標格的（The Hostel 的第一筆 Tyanalon 340 就是）。這一輪把路徑摸到了
一半：`hostel6.json`（六人隊、2000 金幣、城 7 入城點 7）買得到 Tyanalon，
走到區域上緣 (59,16) 出得了城，種子 11 在野外第 14 步遇到 GHOUL——**卡在戰鬥那一段
的按鍵序列**，第二回合起手動編的 token 推不動規劃段。

沿路量到的兩件事：一人隊在野外一格要六十幾次按鍵（六人隊十幾次就一格），
所以用一人存檔掃遭遇不切實際；`<` 在圖示列上會繞回最後一格，挑「休息」用 `< RET`
比數格數穩。

## 2026-09-06（續五）徒手拿的是穿戴品

RE-131 把 `sub_251B0` 的徒手分支記成「只影響額度裡的速度項」。往上追
`sub_26AC2:0x26BDA` 才看清楚**那不只是速度**：`var_4` 是「這一手用哪一件東西」的
位移，`0x26BFC` 起用它讀出種類、傷害碼與基礎傷害——整組傷害骰都從那裡來。

`[0x0FD2]` 是 `00 50 60 00 40 10`，加 `0x100` 之後除以一格 `0x10`，槽號對上穿戴欄
十格（HEAD／TORSO／LEGS／BODY／HANDS／FEET／SHIELD／NECK／RING／BELT）：

| 攻擊選項 | 槽 | 也就是 |
|---:|---|---|
| 1 | 5 FEET | 踢 |
| 2 | 6 SHIELD | 盾擊 |
| 3 | 0 HEAD | 頭槌 |
| 4 | 4 HANDS | 拳 |

**四式與四格穿戴品一一對上**——這本身就是表讀對了的證據。remake 之前徒手一律拿
手上那件在算傷害，數字整個不對。

`Actor.UnarmedWeapon[6]` 與 `AttackWeapon(式)`：命中折半那一條、傷害骰與額度的
速度項都改從它取。**怪物那一側沒接**：同一張表對怪物是加 `+0x30`，五個位移裡有
三個落在怪物記錄的已知版面之外，還沒查那幾格是什麼（記在 SPEC-008）。

`arena` 與 `full-chain-ending` 重生（3053 → 2973、2142 → 2063 個 token）——變短是
因為徒手打得比較實在，戰鬥回合數少了。里程碑與檢查點跟著搬，二十張過期截圖刪掉。
十五份收據全部通過。

## 2026-09-06（續六）怪物徒手武器補完，亂數改成強制對拍

怪物那一側的徒手武器接上了（RE-131 留的洞）。怪物記錄 `+0x30` 起那七格就是穿戴欄，
一格 16 byte，順序與角色記錄相同——`sub_26AC2` 對怪物那一條取 `[0x0FD2 + 攻擊選項]
+ 0x30`，與隊員同一張表。八個位移解出來全部是像樣的物品名，這就是版面讀對了的證據。
`extract_monsters.py` 補出七格、`NewQuest` 依 `UnarmedSlots` 填進 `UnarmedWeapon`。
十五份收據全過，主樹 `42c4459`。

### 亂數：不等遊戲自己抽，直接呼叫原版

RE-132 解出原版的 `rand()` 之後，對拍卡在「原版跑不到會抽亂數的地方」——2.6 億道
指令、送 12 批鍵，一次都沒抽。先做了正對照（`_main` 與 `srand` 都觸發得到），確認
那是事實不是 hook 沒掛上。

問題在於**繼續往那個方向走是在解錯的題目**：要走到戰鬥得先解出開機、選單、建隊、
進戰鬥每一步的按鍵序列，而中間任何一步錯了，看到的都是「一次都沒抽」——與「原版
根本不抽」長得一模一樣。花的時間全部不是花在對拍上。

改成在 dosgolem 加 `Call(addr, args…)`：在目前堆疊上安排參數與一個假返回位址，把
`CS:IP` 指到目標，跑到 `retf` 回來，還原暫存器。**把原版當函式庫用**——走的是同一段
機器碼、同一份全域種子，所以驗到的就是原版的實作。

| 對拍 | 結果 |
|---|---|
| `_main` 進入時的種子 | `1`（MS C 初值）——正對照，一次證明映射、DS、種子位址都對 |
| `rand()` | 6 個種子 × 500 抽逐次相同，收尾 32 位元狀態也相同 |
| `rand(n)` | 十種面數各 40 次逐次相同 |
| 耗時 | 0.02 秒；整套 oracle 測試 15.5 → 4.3 秒 |

**這一層驗常式的實作，不驗呼叫時機。** 兩件事不要混為一談：誰在什麼時候抽、抽幾次
還沒驗過，那要 dosgolem 第四層（`apps/kol/`）。

順帶量到原版 `rand(0)` 是除零當機（R6003，回傳碼 255），不是某個安全的預設值。

### dosgolem 分支（`kol-force-call`，`ce78896`）

工作目錄複製成 `~/cht/dosgolem-kolrand`，`tools/go_docker.sh` 的預設掛載指到它。
除了 `Call`，修掉兩處「錯了不會報錯」的：

1. `At()` 改比線性位址。`IDA()` 產生的是正規化位址（offset 一律 0…15），執行期的
   `CS:IP` 幾乎不會長成那樣，所以結構相等的 `At(o.IDA(x))` **永遠不成立而且不噴錯**，
   只會安靜地跑滿上限。far call 的大模式程式必踩。KOL 側自寫的繞過因此刪掉。
2. `DS()` 的 DGROUP 基底補上 `DGroupSeg`／`SetDGroup`／`SyncDGroupFromDS`，文件標明
   預設值是 rich2 專用常數。換執行檔不重設就會讀到別的東西，同樣不報錯。

`Call` 的回歸測試用**合成的 MZ 映像**，不需要玩家自備素材。

### 代價：十三份收據基準失效

換對亂數之後遭遇（`ambush.go:109`）與治療者的觸發步數位移，三種形狀——腳本留在
野外模式、里程碑位置差一兩格、戰鬥截圖雜湊不同。**沒有一項指向 bug。**

使用者 2026-09-06 決定：**先把 RNG 併回 main，收據另案登記**，理由是 RNG 修正已被
原版逐次驗證、不該卡在數小時的重生後面，而 dosgolem 第四層解出真正的抽取時機之後
重生可能要再做一次。失效狀態記在 `WORKLIST.md` 的「收據基準」一節——收據跑起來會
紅，那是正確的狀態，不假裝通過。

### 8.3 截斷也修掉了（`1e204ce`）

合併之後在主樹重跑對拍，`TestBootDrawsNothing` 紅了——主樹的遊戲目錄沒有 worktree
那份手動放的 `steedpics` 別名，原版在 180 萬道指令時「跑出可用記憶體」，而且**在
飛掉的路上抽了 1 次亂數**（錯誤處理路徑，不是正常流程）。

那正是 8.3 截斷這個洞的樣子：FAT 的主檔名只有 8 個字元，真 DOS 開到的是
`STEEDPIC`；dosgolem 回「找不到檔」，而原版不檢查開檔結果。**症狀離原因很遠**，
看起來像模擬器壞了。修在 `internal/dos` 的 `resolve`：原名（含大小寫不分）找不到
才試截斷版，所以 `DATA.PAX` 不會因此開到 `DATA.PAK`。

修好之後主樹九條對拍全過，遊戲目錄不必再放別名。**別名那種本機繞過本來就是一種
債**——它讓下一個 session 在別的機器上看到一個無法解釋的失敗。

## 2026-09-06（續七）規則對拍：寫記憶體、釘死骰值、呼叫原版的常式

使用者指出方向：不用追 `rand()` 到實際結果，改記憶體的狀態去對拍要驗證的事情，
只要 remake 行為一致即可——`rand()` 本來就不容兩個同時一樣。

那句話換掉了整個問題。原本的路是「寫 dosgolem 第四層驅動遊戲走到會抽亂數的地方」，
**那是解錯的題目**：要驗的是規則，而輸入可以直接寫進原版的記憶體、規則可以直接
呼叫原版的常式；亂數序列不必對齊，把骰值釘成常數就好。

於是一條規則的對拍變成三步：**寫記憶體設好輸入 → 釘死骰值 → 呼叫常式讀結果**。

### dosgolem 加了 `Stub`（`633fd66`）

走到某位址的 far 常式不執行，直接用給定的值當 `DX:AX` 返回。關鍵是它要在**原版
自己的呼叫鏈**上生效：攔住底層的 `rand()`，上層的 `rand(n)` 跟著變成決定性的——
不然「上層規則 → 擲骰」那類對拍一條都做不了。測試釘著這一條。

順帶把 `Missing()`（開不起來的檔）與 `VideoMode()`（載入進度的路標）接出來。

### 對拍過的四條（RE-133）

`sub_23683`（疲勞比例）、`sub_26CD5`（技能罰值）、`sub_2526C`（額度夾擠），
加上「stub 掉 `rand()` 之後 `rand(n)` 回固定值」。前三支是 RE-129／130／131 那一輪
解出來的數字鏈，這一輪之前**只有靜態證據**；現在是原版自己在執行期回的數在背書。

`sub_23683` 尤其重要——remake 原本寫成「50 × 累積傷害 ÷ 昏迷門檻」，五處都吃它。

前提是 DGROUP 基底：dosgolem 的預設值是 rich2 專用常數，`prepare()` 跑到 `_main`
之後抄程式自己的 `DS` 再比對筆記說的 `0x2DED0`。**這一步不做的話，底下每一個
`ds:` 位址都在讀別的東西，而且不報錯。**

### 綠燈本身不算數

常式沒被呼叫到、記憶體沒被改，而 remake 剛好算出同一個數——看起來與「對上了」
一模一樣。所以每一支都放了至少一組**沒跑就會紅**的輸入（輸出格先寫 `0xDEAD`、
罰值大於技能、花費超過三倍）。

### 還沒驗的仍然是呼叫時機

誰在什麼時候呼叫、一回合抽幾次骰，這一層問不到。那也是收據基準重生的前置。
