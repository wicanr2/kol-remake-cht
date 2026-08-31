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
