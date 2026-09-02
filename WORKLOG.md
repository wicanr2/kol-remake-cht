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

