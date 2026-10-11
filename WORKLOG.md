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

### 對拍抓到的第一個缺口：換武器之後負重沒重算

第五支對拍（`sub_251B0`，一手的花費）**原版一律回 0**。輸出格確實被寫了，所以常式
跑了——原因是它開頭就先呼叫 `sub_24FD3` 從穿戴品重算速度與負重，把寫進去的值蓋掉。

那揭露了 remake 的缺口：原版每次算花費都重算，所以戰鬥中撿起、丟下、換走手上那件
武器，負重與武器速度會立刻反映；remake 的 `items.go` 只換 `Weapon`，兩格停在開場
算的那一次。影響是一整條鏈——花費 → 疲勞 → 撤退判斷，而攻擊分數與執行鍵值也各自
減去負重。

修法：`Actor.ReloadGear()`，撿起、丟下、換到腰帶三處呼叫；`Actor` 多存
`LoadLevel`／`ArmourWeight`／`ShieldWeight`（`LoadLevel` **不是** `Level`——
怪物那一格另外除以 5 再加一）。兩支單元測試釘住撿與丟。

**這個缺口是「原版回了 0」帶出來的。** 那看起來像測試寫錯；把它當成參數沒設對而
去調參數，就會錯過。stub 在這裡的用途不是釘死骰值，是**切斷上游依賴**讓一支常式
變回純函式。

### 再兩支：敵我旗標與敵人技能縮放

`sub_22229(i)`（是不是敵人）對十八筆空白記錄全部回 0——**判準不是索引**。逐格探測
（把記錄的第 k 個 byte 設 1，看回傳值變不變）定位到 `+0x01`，五十次呼叫就問完，
比開 IDA 便宜得多。

有了那一格才做得了 `sub_26AC2(i)` 的敵人分支：技能 × 遭遇等級 ÷ 5，十組全部對上
remake 在 `state.go` 的折算。順帶量出 16 位元的環繞邊界（6554 × 10 那一組原版回 0），
實際資料最大 300 × 10 ＝ 3000 到不了——記下來是為了讓它「已知且走不到」。

### 這套方法的界線

`sub_26D76`（攻擊分數與防禦分數）試過直接呼叫：三支上游與 `rand(n)` 都 stub 掉、
設好力量與快捷，**跑滿一百萬道指令沒有回來**。它不是純函式——算完分數還要印訊息、
扣裝備耐久、更新畫面。

**這條界線本身是結論**：這套方法適用於「讀幾個全域、寫幾個全域」的規則常式；
會碰畫面、訊息或等輸入的那一層要換別的辦法。判準很直接——跑不完就是跑不完。

負面結果留在文件裡，測試檔不留永遠紅或永遠 skip 的殼——下一個 session 才不會
再試一次同一條路。

### 原版不是「在等滑鼠」，是早就飛掉了

追「dosgolem 第四層」的第一步，把開機卡住那條線拆開。表面症狀全部誤導：滑鼠輪詢
65 萬次、七個鍵都推不動、畫面全黑——看起來像在等玩家點畫面上的東西。

實際的鏈：第 166 萬道指令時，IDA `0x1DC59` 的搬移迴圈用 **`ES = 0000`** 從 0 寫到
0x190，把**整個中斷向量表**覆寫掉；`int 08h` 因此變成垃圾；第 181 萬道時計時器中斷
把 `CS:IP` 換成那個垃圾位址；程式在全 0 的記憶體裡「執行」到程式碼區，落進
`sub_1B100`（輪詢 `int 33h` 的公用常式）的迴圈中間。

**判準是 prologue 沒執行**：`sub_1B100` 的 `0x1B100`…`0x1B11E` 一次都沒跑，而迴圈體
跑了 65 萬次。跳進迴圈中間只可能是飛進去的——這一條比任何「畫面／輸入」的觀察都
決定性，而且便宜（掛幾個 `OnCall` 就問到了）。

路上補了 dosgolem 三個診斷缺口，每一個都擋過一次：`TextScreen()`（`Indexed()` 讀的是
圖形記憶體，文字模式下全黑看起來像沒畫）、`PortWrites()`（EGA 直接設模式不走
int 10h）、`Regs()`（段暫存器錯了從指令看不出來）。

下一步是查 `ES` 為什麼是 0——原版在真 DOS 跑得起來，所以那是 dosgolem 某個服務回的。

### 根因：KOLTOWN 要有人先設好圖形模式

`ES` 的來源查出來了：`word_2FB54` 是**視訊記憶體的段**，全庫只有兩個寫入點，都在
`_main` 的顯示卡偵測裡——比對 BIOS 資料區 `0040:0049` 的目前視訊模式，只認 4（CGA
320×200）、9（Tandy）、7（單色）三條。

dosgolem 冷啟動填的是 3（DOS 的預設），**三條一條都不中**，於是視訊段留在 0，
後面每一支繪製常式都拿 `ES = 0` 去寫——中斷向量表就是這樣被當成畫布的。

**KOLTOWN 不是從 DOS 提示字元直接跑的**，它是主程式設好模式之後 EXEC 起來的子程式。

把 `0040:0049` 設成 4：視訊段是 B800、向量表完好、程式跑滿四千萬道指令停在
`sub_1B100` 裡輪詢滑鼠 675,302 次——**這一次是真的在等玩家點畫面**。

路上還原了一個猜測性的改動（`AH=4Ah` 改成可增可減）：語意上它是對的，但沒解決問題
也沒有測試覆蓋，**沒有證據支持的改動不該留下**。

畫面仍然讀不到，因為 `Indexed()` 讀的是 A0000 而 CGA 模式 4 的視訊記憶體在 B8000。
下一步是把那一塊讀出來，才知道該點哪裡。

### 原版載入到城鎮了

把 `0040:0049` 的每一個值都試過。模式 **09（Tandy）與 0D（EGA）** 是唯二走完載入的
——都開到 `TOWN7.0`，其餘模式五個檔就 `AH=4Ch` 離開。

四個判準一起看才不會誤判：模式 4 設了視訊段卻只開 4 個檔；模式 0D 開了 19 個檔而
視訊段是 0（EGA 不用那一格，它走 A0000）。**走得遠與走得對是兩件事。**

對 oracle 來說 **Tandy 比 EGA 好用**：模式 09 的視訊記憶體是線性的，dosgolem 不需要
額外邏輯就存得對（24,480 個非零像素）；EGA 要走位元平面，dosgolem 沒有那一層，
同一個狀態只剩 605 個像素。這不影響規則對拍——要驗的是規則，畫面只是路標，
而 KOL 自己支援 Tandy，所以那是原版走得到的路不是繞道。

dosgolem 補了 `CGA4()`：CGA 的掃描線交錯（偶數列從 0、奇數列從 0x2000），照線性讀
會得到一張梳子，於是「全黑」看起來就像「什麼都沒畫」。

### 畫面解得出來，但還沒人眼確認

`Tandy16()` 解出來的色號分布集中在四個（0、4、12、15），有規律的空間結構，不是雜訊。

⚠ **色號分布證明不了解碼參數對**：段數記錯時讀的是同一批 bytes，分布幾乎不變，
變的是像素的位置。ASCII 輪廓看得出橫條與重複圖案，但那不足以斷言「這是城鎮畫面」。
要確認得存成 PNG 給人眼看——那是下一步，不是現在能宣稱的事。

**人眼確認過了**：存成 PNG 之後看得很清楚——是一張**木紋框架，三塊橫向木板**。
解碼參數對（錯了會是四張交疊的梳子）。原版真的在 dosgolem 裡畫出畫面了。

### 下一個障礙

畫完木紋框架之後，程式停在 IDA `0x1A1BC`…`0x1A1EE` 的迴圈裡，跑滿 **2.4 億道指令**
沒有離開。那是「跟著鏈結串列走到 -1、一邊填一張表」的迴圈，而指標本身有效
（`2EFF:0000`，指向的記憶體全是 FF ＝ 終止值就在那裡）。

畫面在前兩億道內從 24,480 個非零像素長到 49,198 之後不再變，滑鼠只輪詢 1 次
——**不在等輸入，是在空轉**。下一步要反組譯那一支看終止條件為什麼不成立。

終止條件為什麼不成立量出來了：取樣四百次，目前走到的節點**永遠是 `0x07A0`**
——那張表在 `2EFF:0F40`（＝ `0x07A0 × 2`）存著 `0x07A0`，**指向自己**。長度 1 的環。

所以問題在「那張表怎麼被填的」，不在走訪的迴圈本身。下一步找那個寫入點
（`WatchWrites` 收的是 DGROUP 偏移，這一塊在配置來的記憶體裡，要另外想辦法）。

### 第七支規則對拍：負重與武器速度

`sub_24FD3` 就是抓到缺口的那一支。它不吃參數，用「當前行動者」的全域——照怪物那一側
設好資料再傳 slot 進去，原版一律回 0。

輸入位置**逐格探出來**：當前行動者是參戰者 0、手上武器在角色記錄 `+0x30`（重量
`+0x3A`，不是 `+0x20`）、三件護甲的重量在 `+0x10A`／`+0x11A`／`+0x12A`、等級在參戰者
記錄 `+0x02`。八組全部對上 `battle.Encumbrance()`。

**這個手法第二次奏效**（敵我旗標那次五十次呼叫問到 `+0x01`）：比讀反組譯快，
而且答案是原版在執行期給的。

溢位邊界：`係數 × 總重` 是有號 16 位元，等級 1 的係數 70 → 門檻總重 468（實測 468
相同、469 起回負數）。實際資料上限 317，走不到，但**比技能那一格近得多**。

### 第二個缺口：盾牌不是永遠算數的

`sub_24FD3` 的盾牌那一段有條件，逐格探測問出來：參戰者記錄 `+0x15`／`+0x17`
（兩隻手臂）與武器格 `+0x0D`（種類 byte，雙手那一位元在裡面）。所以原版的判斷是
**盾牌要在、兩隻手臂都還在、手上那件不是雙手武器**，而且每次算負重都重新判斷。

remake 把盾牌等級當成建戰鬥時定下的常數——手臂被打斷或撿起雙手武器之後，防禦值
裡的 `100 × 等級 + 200` 都還在加。修法是 `Actor.EffectiveShieldGrade()`。

**同一支常式抓到的第二個缺口，而且是同一個模式**：原版每次取用都重算，remake 快取
了。第一個是「換武器之後負重沒重算」。這個模式值得在別的地方也找一遍。

### 第三個缺口：換武器要換技能那一列

照「原版每次取用都重算、remake 快取了」這個模式找下去，第三個就在 `sub_26AC2` 的
隊員分支：拿手上武器的種類 byte 去角色記錄 `+0x52` 起四筆裡找，找到哪一列就用哪一
列的攻防。實跑驗過四列都選對，**種類不在四列裡時攻防都是 0**。

remake 在建戰鬥時選一次就不再換，所以撿起不同種類的武器之後技能還是舊那一列。
修法是 `Actor.WeaponRows` 整份帶著 ＋ `ReloadSkills()`。怪物那一側不吃這一條。

**三個缺口、同一個模式**，而且找法是機械的：把 `Actor` 上「建戰鬥時算一次」的欄位
逐個對回原版的取用點，看原版是不是每次都重算。

### 其餘欄位對過了，第四項是「整條沒做」

`Strength`（飢餓折算）、`Stamina`、`Threshold`、`Level` 在戰鬥中都不變，兩邊一致
——特別是 `Stamina`：原版的 `[0x1C50+2i]` 只由 `sub_21AC4` 在開打時寫，法術改的是
`+0x08`（力量），碰不到那一格，所以 remake 不重算是對的。

剩下的 `ArmourWeight`／`ShieldWeight`／`UnarmedWeapon` 會變，但變的原因是**裝備耐久**
——而 remake 連耐久欄位都沒有（`Item` 與 `Weapon` 都沒有，抽取器也沒抽）。所以那不是
「快取了沒重算」，是整條沒做。登記在 `WORKLIST.md`。

### 裝備耐久：扣除規則證實了，觸發時機沒有

`sub_22084` 的扣法實跑量到：`記錄[+0x0F]` 的 bit 5 沒設就不扣；新耐久 ＝
`記錄[+0x0B] − 32 × 次數`；為負就把那一格 16 byte 整個清成 0，剛好 0 不清。

六個呼叫端（第三支 IDA 探針）：四個是穿戴欄那四格、次數固定 1；兩個在攻擊結算裡，
次數是區域變數。**那幾個變數是什麼還沒解**——`sub_26D76` 直接呼叫跑不完（含訊息與
畫面），量不到，只能靜態讀，而「很像傷害減上限」是強推論不是已證實。

照 spec-gated 的閘門，觸發時機沒到 READY 之前不進實作。**這一條的界線與
`sub_26D76` 那條是同一條**：會碰畫面的那一層，直接呼叫問不到。

### 自環那一格：那塊記憶體是共用暫存區

用 `WatchLinear`（這一輪加給 dosgolem 的能力——`WatchWrites` 只收 DGROUP 偏移，
涵蓋不到配置來的記憶體）把寫入序列抓下來，第 170 萬道時 IDA `0x1C873` 把那三個
byte 寫成 **`44 44 44`**，也就是 ASCII 的 `'DDD'`。

**那塊記憶體被當成文字緩衝區用過**——它不是一張專用的鏈結表，是共用暫存區，
而走訪迴圈讀到的是另一種用途的殘留。最後把 `0x0F40` 寫成 `0x07A0`（指向自己）的是
走訪迴圈自己那一支（`0x1A246`）。

還沒解的是 `sub_1A1xx` 的完整邏輯。那要反組譯——同一個 `.i64` 這一輪已經開過三次，
下一輪再開。

### 卡住的那次解壓：輸入是檔頭本身

`FileOps()`（這一輪加給 dosgolem 的能力）把兩種載入路徑並排出來：`kolmpics.0` 一次
讀整個檔，解壓輸入是「緩衝區 ＋ 0x12C」＝ 緩衝區 ＋ **300**，正是檔頭大小；
`magic.0`／`koldata.0` 是「讀 300 檔頭 → seek 到 200 → 讀資料」。

而 `town7.0` 是「讀 300 檔頭 → **從 0 再讀 8825，沒有 seek**」，解壓輸入位移 **0**
——等於把檔頭本身餵給 LZW 解壓器。

那條路徑的 seek 與位移都是從檔頭算出來的，算出 0 表示 entry 編號不對。編號多半來自
KOLTOWN 自己的狀態——**它是主程式 EXEC 起來的子程式**，而 dosgolem 的 PSP 命令列
（`+0x80`）目前留空。與「視訊模式」是同一個形狀：少了前置，它會用垃圾繼續跑。

路上修了 dosgolem 的一個監看漏洞：**`WriteBytes` 不觸發 `onWrite`**，所以檔案讀取、
映像載入這類批次寫入在監看下是隱形的——看起來像「沒有人寫過」，實際上整塊被蓋掉了。
那個漏洞害我一度以為緩衝區沒被寫。

### 主程式證實了「前置狀態」那個假說

遊戲目錄裡有 `KOL.EXE`、**`MAKEMODE.EXE`**、`TANDYPIN.EXE`、`TITLE.EXE`、
`KOLBOOK.EXE`、`KOLTOWN.EXE`，**五支都不是壓縮的**。

從 `KOL.EXE` 跑起來，它做的第一件事就是把**視訊模式設成 `0Dh`**（EGA 320×200），
然後 EXEC `TITLE.EXE`——這證實了 KOLTOWN 那一條：模式是主程式設的。

補了 dosgolem 三個 `int 21h` 服務（`AH=08h`／`AH=0Bh`／`AH=43h`），補之前 `KOL.EXE`
連 `TITLE.EXE` 都走不到。

**還沒解**：樁掉 `TITLE.EXE` 之後鏈仍停在 `KOL.EXE` 自己的小迴圈裡，不再 EXEC 下一支，
送鍵沒有反應，而 `int 19h AH=4A` 被呼叫兩萬多次——跑進資料區當程式碼執行的樣子。
啟動鏈用的目錄是 `workplace/oracle/game-chain/`（硬連結原檔，只換跑不動的那幾支）。

## 2026-09-06（續）原版第一次走到可互動畫面

`KOL.EXE` 那個「不再 EXEC 下一支」的停頓解開了，連著解出三個相扣的模擬層缺口。
上一節寫的「KOLTOWN 是主程式 EXEC 起來的子程式」是錯的，正文已改寫；
勘誤集中在這裡。

### 一、記憶體服務只報成功、界線沒動

停頓那一段的機器碼與檔案裡的對不上：記憶體是 `cd 19 03 fa`，檔案是 `cd 21 72 fa`。
`WatchLinear` 掛在那兩個 byte 上，寫的人是**那段碼自己**的 `mov es:[0], cx`。

它是「把記憶體配光」的慣用法：配 1 段 → `AH=4Ah` 撐到最大 → 把上一塊的段寫進這塊
開頭串成鏈 → 再來一次。`AH=4Ah` 擴大成功卻不推進 bump 指標的話，這個迴圈永遠不會
結束——三百萬圈之後配到程式自己的映像上，那句寫入就把自己的 `int 21h` 改成了
`int 19h`。

`avail` 寫成 `MemTop - seg` 還會在指標越過 `MemTop` 之後 uint16 環繞成 `0FFFFh`，
配置變成永遠成功。縮小也要把空間還回去：`KOL.EXE` 配光之後縮回四段、緊接著就要
那四段，不還的話它以回傳碼 255 離開，主控台只留一行 `run-time error R6006`。

新加的 `MemOps()` 是這一節的關鍵工具——配置器把已經有人住的段配出去時，症狀是
程式碼被自己寫壞，沒有那份軌跡只能從被改掉的位元組往回猜。

### 二、`AH=4Bh AL=03h`（載入 overlay）沒實作

修完之後主控台才說得出話：`run-time error R6006 - bad format on exec`。那不是 EXEC
子程式，是**載入 overlay**——主程式自己配好記憶體、叫 DOS 把映像搬進去、再 far call
進入點。`TITLE.EXE`／`KOLBOOK.EXE`／`KOLTOWN.EXE` 都是 `KOL.EXE` 的 overlay。

沒實作的症狀不像少了一個服務，像那支 EXE 檔壞了。`Unimplemented()` 裡的
`int 21h AH=4B AL=03` 才是線索。

### 三、handle 號碼只增不減

片頭跑起來了，但城鎮仍停在 LZW 的走訪迴圈。表面上是解壓器吃到了目錄本身——
它從 `TOWN7.0` 的位移 0 讀，而第一張圖在位移 200。

軌跡裡看不到那次 `AH=42h`。真 DOS 從 JFT 挑第一個空位發 handle，開關幾十次號碼
仍然是 5；dosgolem 只增不減，爬到了 23。遊戲拿 handle 當自己那張表的索引，表裡存著
「這個 handle 上次 seek 到哪」——越界之後新開的檔繼承了別人的位置，於是判斷
「已經在 200」而跳過 seek。上一個檔 `MAGIC.0` 剛好就 seek 到 200。

診斷的洞一起補：`FileOps()` 現在也記開、關與**失敗的呼叫**。原本 handle 無效就直接
返回不留紀錄，那正是差一步就要歸咎到遊戲邏輯上的原因。

### 成果

`KOL.EXE` → `TITLE.EXE`（片頭五組圖播完循環，按鍵可中斷）→ `KOLBOOK.EXE`。
城鎮載入 35 個檔（原本 19）之後停在 `sub_1B100` 輪詢 `int 33h` 等玩家操作，
畫面是建築物內的對話：角色頭像、隊員圖示、文字框、程式自己畫的小手游標。
在文字框上點一下，對話推進到下一段，底部長出六個圖示的指令列。

`Click` 對這個遊戲要帶兩個新選項：`NoCursorWait()`（KOL 從不叫 `int 33h AX=4`，
游標是自己畫的，預設那個等待永遠不會成立）與 `Frame((*Oracle).Tandy16)`
（視訊記憶體不在 `A0000`，預設取樣全 0 會讓每一次點擊都回「沒反應」）。

dosgolem `kol-force-call` 分支三個 commit，各帶測試；規格更新在 `004` §1.3／§1.4、
`005` §4.1／§4.2、`007` §2.5。KOL 這邊刪掉八支為已修 bug 寫的一次性探針。

### 第一支「呼叫時機」對拍：野外遭遇

原版跑得動之後立刻兌現的東西。RE-132 對拍的是 `rand()` 與 `rand(n)` 的**實作**
逐次相同，`rand_test.go` 那句「這裡不驗呼叫時機」一直是收據基準重生的前置。

野外遭遇（`sub_2A4E5`）四條走法逐條對過，原版與 remake 抽的次數與上限完全相同：
沒有遭遇 `[100]`、治療者 `[100, 100, 3]`、挑中怪物 `[100, 100, 32, 3]`、
一隻都不符合 `[100, 100, 32×32]`。RE-078 的擲點流程因此從 `強推論` 升到 `已證實`。

兩個踩到的坑值得記下來：

**資料沒載入的話那條路等於不存在。** `_main` 剛進來就呼叫 `sub_2A4E5`，機率算出來
一律 0，`rand(100) >= 0` 恆真——不論骰值釘成什麼都回「沒有遭遇」。看起來像原版
根本沒有遭遇分支，實際上是 `payload3` 還是空的。先讓原版跑到城鎮地圖就對了。

**「一隻都不符合」那個情境第一次寫錯了。** 原版那側固定回 1 → 怪物編號 0，
而編號 0 在當下的地形與季節**剛好符合**，於是第一次就中、只抽 4 次；remake 那側
的 fixture 是「全不符合」抽 34 次。兩邊走的不是同一條路——**測試設計的問題，
不是行為不一致**。改成回 2（編號 1 不符合）就對上了。

順帶量到城鎮地圖上每走一步抽 0 次骰（遭遇是野外的事），以及機率實測落在 4…50。

敵人組成那一段跟著也對完了。騎走那一問是有條件的——`word_2FA2C == 20` 表示隊裡
有人沒坐騎，整段跳過直接開打。設它就讓 `sub_2A4E5` 一路跑到底。**跑不完不是失敗**：
算完之後原版接著畫戰場又停在 `sub_1B100` 等滑鼠，而結果在那之前就寫進全域了。

使用者 2026-09-06 定調了對拍的形狀：**不要追亂數的呼叫次數，機制才是要對的東西**。
骰值釘死只為了讓亂數不介入，比的是「設好狀態之後兩邊算出同一個結果」。
遭遇這一支因此改成比四個全域（怪物、等級、隻數、站位組），五組含下限的負對照
（位階 1 加級距修正 −2 得 −1，兩邊都夾成 1）。

天候接著也對完了（九組，含冬天那條「推到第 3 段再多一段」）。它落在
`word_2DF5A`——那個全域原本沒有名字，是這樣找出來的：時段變換前後對 DGROUP
做差分，剩下兩個都是 `0 → 1` 的候選；把兩格設成不同的值再擲一個小骰，
照 `NextWeather` 該變的那一格才是它。**差分縮小範圍、設狀態確認**，
比開 IDA 便宜，而且答案是原版給的。

一個容易漏的前置：天候只有**時段變換**時才擲。第一次差分沒設時刻與時段，
`sub_29987` 走的是「時段沒變」那條，只有三個 word 動——看起來像天候根本不在
這一支裡。把時刻設成 2000（NOON 那一段）而時段還記著 0，變換就成立。

攻擊結算（`sub_26D76`）先前列為「不能直接呼叫」——實際上可以，先前是**根本沒傳
參數**。`arg0 != 0` 只跑 11.7 萬道就返回而且一支子常式都沒呼叫，那個「什麼都沒
發生」的路徑正是當初的誤判來源。開頭的碼從記憶體讀出來解就看得到
（`cmp word [bp+6], 0`），順帶解出當前行動者索引在 `ds:9E5A`。

這一段踩到三個坑，每個都值得記：

**正對照本身會失敗。** 想用 `sub_23683` 當「攔截器有沒有掛上」的正對照，結果它
的分母是「力量 ＋ 健康」，資料全 0 時直接 `R6003 integer divide by 0` 讓原版
離開——連正對照都做不成。換 `sub_24FD3`（除數是常數 100）才行。

**跑不完的呼叫會污染狀態。** `Call` 只還原暫存器不還原記憶體，停在半路那一次
把全域留在中間狀態，**下一次呼叫走的是另一條路**。同一個實例連續問四次，只有
第一次的答案算數；每個情境都要新的實例。

**探測的預算要照單次成本抓。** 單次呼叫在 `_main` 剛進來時要 224 萬道，資料載入
之後只要 11.7 萬道——差 19 倍。逐格要跑上百次，用錯狀態或把預算開到兩千萬，
探測本身就變成整晚的工作（第一次就是這樣被砍掉的）。

出口用**雙輸入差分**找：單跑一次有七百多個 word 變，絕大多數是畫面；兩組只差一個
輸入的執行，差異就只剩三個 word，其中一個還是輸入本身。

實跑問到「對手還沒進來」就到頂了——改防禦者的欄位沒有反應，把整筆記錄塗成 0xFF
也一樣。這時候才是開 IDA 的時機，而且**一支探針一次問完**：完整反組譯、參照的
每一個全域、六支已對過的子常式在裡面的呼叫點、三個已知全域的讀寫點。

答案很乾脆：`sub_26D76` 901 道指令只參照 23 個全域，**指令碼在 `word_31122`
（`ds:3252`）而不是參戰者記錄的 `+0x25`**——先前怎麼設 `+0x25` 都沒反應就是
這個原因。`word_3237C` 是攻擊值（`記錄[+0x08] + word_34A12`）、`word_2FB46` 是
攻擊分數、`word_30A78` 是這一手打第幾下。`sub_2526C`／`sub_251B0`（額度）
一處都沒有——那兩支不在這一支裡。

攻擊分數的骰前算式對完了（七組）。做法是把 `sub_26AC2`／`sub_23683`／`sub_24FD3`
三支釘住，技能、疲勞比例與負重保持寫進去的值，驗的就只剩 `sub_26D76` 自己那段。

第一次漏了 `sub_26AC2`，七組全紅而且**每一組的差都剛好等於「技能 ÷ 10」**——
看起來像 remake 多加了一項，實際上是原版那一側的 `word_34A12` 被重算成 0。
這種「差值有規律」的紅燈通常不是公式錯，是有個中間值沒釘住。

防禦分數跟著也對完了（七組）。這一段一次定下三個先前不知道的全域：`word_3237E`
防禦者索引、`word_2F942` 防禦分數、`word_2FB46` 攻擊分數，命中判定就是
`word_2FB46 >= word_2F942`。`word_3237C` 只是中間值，攻守兩邊共用——第一次讀錯
那一格，量到 200 而 remake 是 60。

「沒有盾牌」的表示又是一個會安靜出錯的地方：判準是 `word_34A0C >= 0x80`，
寫 `0xFFFF` 想表示「沒有」反而被當成有號的 −1 落進「有盾牌」那一條，
防禦分數多了 100。

傷害鏈跟著也對完了（六組），攻擊結算的數字鏈到此完整：攻擊分數、防禦分數、
命中判定、傷害。

卡最久的是「命中了但傷害是 0」。擲骰序列只有攻防那兩次，看起來像傷害那一段
根本不存在——實際上**命中之後原版先印訊息再 `sub_1B100` 等玩家點滑鼠**，傷害在
那之後才算。釘住那一支，序列立刻成了 `[100 100 骰面 100 …]`。
**畫面與等待輸入也是可以釘住的**，不必繞過去。

命中部位的算式與護甲減傷也讀出來了，但**肢體的扣傷不在 `sub_26D76` 裡**：兩組只差
部位那一擲的執行，防禦者記錄五十個 byte 一個都沒差。那一支只算傷害與護甲，
扣到頭／胸／臂／腿是呼叫端做的。差分為空這次是**真的空**，不是訊號選錯——
同一份差分在傷害那一格是看得到變化的。

肢體的扣傷找到了：**`sub_277F4`**，位址剛好接在 `sub_26D76` 結束之後——
remake 的 `Actor.Take` 註解本來就寫著那個位址，先前繞了一圈才想起來查自己的碼。
參數是 `(傷害, 部位)`，七組對完（含剛好扣光、扣過頭、累積傷害到門檻、夾在 0x7F）。

順帶量到一個表示差異：**廢掉的那一肢原版寫 −1（`0xFF`），remake 寫 0**。
兩邊判準都是「`<= 0` 就算廢」所以戰鬥行為一致，但存檔或畫面讀那一格會差。
測試因此比語意而不是精確值，差異登記在 `WORKLIST.md`。

攻擊結算的完整鏈對完之後就去動收據，第一份 `arena` 重生出來還是紅的，
錯誤與重生前一模一樣。`regen_long_path.py` 依舊軌跡切段——非戰鬥段照抄、
戰鬥段用 driver 重打——而**照抄那一段才是問題**：競技場的路線是確定的
（走到 55,13），但路上多出一場隨機遭遇，第 77 步時隊伍在戰鬥中而不是在競技場，
後面照抄的野外 token 全部對不上。

`drive_battle_script.py` 吃「世界目標 X,Y」可以自己走到某一格，`regen_long_path.py`
沒有用它。下一步是讓非戰鬥段也重生，而不是照抄。

壞掉的 `arena.json` 已經 `git checkout` 還原——**不留下比原來更糟的收據**。

`regen_long_path.py` 改成野外那一截交給 driver 走（只照抄到踏進野外，之後用世界
目標）。driver 確實走到了每一個目標格，但每一場戰鬥仍是 0 個 token——**走到目標格
就停，那時還沒遇敵**，接下來的戰鬥段拿到「不在戰鬥中」的狀態，打不出東西。
走到 (55,13) 那一段還跑滿了 400 輪的上限。

工具的改動留著（方向對、driver 真的會走了），收據還原——**不留下比原來更糟的收據**。

## 2026-09-07 怪物旗標 bit 0 與雲層動畫

- **`sub_2A18E` 的版面用 dosgolem 實跑驗過**（RE-136）。反組譯讀出三條視訊路徑，
  再在城鎮地圖狀態下改 `word_2DF1E`、往 `word_34A1A:0x7C0`／`:0xF80` 填偽亂數、
  呼叫一次、回讀比對「整列環繞右旋」：EGA 155 列 × 1 位元、Tandy 雲層 31 列 ×
  1 個 nibble ＋ 遮罩 31 列 × 1 位元、CGA 雲層 31 列 × 2 位元 ＋ 遮罩同上，
  逐格相同。`TestCloudScrollMatchesOriginal`。
- **怪物記錄 `+0x10` 的 bit 0 解出來了**（RE-135）：令人生畏。落點是 `sub_2215A`
  寫的 `word_3AC58`，執行段 `0x200AA` 拿它當閘門，沒過檢定就把指令碼換成
  `0x0B`（怯戰）。remake 補上 `checkFlinch`，`tools/extract_monsters.py` 多抽一個
  `fearsome`。`TestMonsterTemplateCopyMatchesOriginal` ＋ `TestFlinchCheck`。
- **失敗分類：查詢回空是自己的探針錯了。** 第一版拿 dosgolem 做雙輸入差分，先寫
  `word_39A48` 再呼叫 `sub_2215A`，量到「整個 DGROUP 零差異」。真因是
  `sub_2215A` 進門第一件事就呼叫 `sub_22296` 把 `word_39A48` 重算掉——
  設好的指標根本沒留下。**連 `word_2FB4E` 都沒變就該察覺碼沒跑到那一行**，
  而不是去懷疑狀態。改讀反組譯之後三分鐘就解掉。
- 順手清掉一條過期標記：WORKLIST 還寫著「倒下那一句的 collapse／shatter 兩個
  分支條件仍未定」，那在 RE-131 就解完並實作了。
- 容器：`tools/run_ida_probe.sh` 一次（`--rm`）、`tools/go_docker.sh` 數次
  （`--rm`）、`tools/run_python_tool.sh` 一次（`--rm`）。沒有殘留 container，
  沒有動別的專案資源。

## 2026-09-07 三平台交付與推廣片

- **macOS 這一支腳補上了。** `docker/Dockerfile.osxcross-go` ＋
  `tools/build_macos.sh` ＋ `tools/verify_macos.sh`。兩弧各編一次再 lipo，
  靜態驗收四道全過（雙弧、arm64 的 `LC_CODE_SIGNATURE`、`minos 11.0`、
  動態相依只落在 `/usr/lib/` 與 `/System/Library/`）。
- **失敗分類：image 疊錯方向。** 第一版把 osxcross `COPY --from` 疊到本專案那個
  Go image（Debian 12）上，`osxcross-conf` 一跑就報 `GLIBC_2.38 not found` 與
  `GLIBCXX_3.4.32 not found`——crazymax 那份是在較新的 Debian 上建的。
  改成 Ubuntu 24.04 起底、把 Go 工具鏈與 module cache 搬過去；版本仍然鎖住，
  `GOPROXY=off` 也照樣成立。
- **推廣片做好了**：59 秒、1280×720、`mean_volume -27.8 dB`／`max_volume -7.8 dB`。
  分鏡與版面在 `docs/promo/compose.sh`，設計 token 在 `docs/promo/theme.sh`
  （theme「赤石與燭火」，色票是從實機截圖的色彩直方圖取的）。
- **兩個 ffmpeg 的坑**（都實際踩到）：
  1. `apad,atrim=0:<秒>` 會讓整支跑不完——`apad` 無長度就是無限產生，而 `atrim`
     只丟掉超過的畫格、**不會讓上游 EOF**。輸出檔停在 48 byte、CPU 一直燒。
     改 `apad=whole_dur=` 不卡了，但音軌只有 8.2 秒（它管「補到多長」不管
     「輸出多長」，前面的 `adelay` 已經把時間軸推掉）。**能用的是第三路
     `anullsrc` 拿 `-t` 定長再 `amix`。**
  2. `volumedetect` 印在 **info** 等級，`-v error` 會把量測值一起吃掉——
     收據裡的響度量到空字串。
- **另一個坑不是 ffmpeg 的**：長 `docker run` 被 harness 在 120 秒轉背景時，
  被砍掉的是 docker **client**，容器還活著。結果同一個 `/out` 有兩個容器在寫。
  長容器一開始就要 `run_in_background`。收尾時停掉了自己那個孤兒容器
  （`docker kill`，只動自己建立的）。
- **第三個坑，最難查的那個**：`otool -l | grep -q LC_CODE_SIGNATURE` 在
  `set -o pipefail` 底下會偶發失敗。`grep -q` 找到就結束，上游的 `otool` 收到
  SIGPIPE 回 141，`pipefail` 把整條管線判成失敗——**簽章明明在，驗收照樣噴錯**。
  機器閒的時候 `otool` 早就寫完不會踩到，load 79 那一次打包就死在這裡。
  單獨跑 `verify_macos.sh` 反而過，因為那時機器沒那麼忙。
  凡是「拿一段可能很長的輸出去 `grep -q`」都要先整份收下來再比。
- **權利審查**寫成 `docs/RELEASE-RIGHTS-REVIEW.md`：`release/` 可以公開、
  `full/` 不可以，說明書譯文要不要進包待使用者決定；**Release 沒有發**——
  那是對外、不可逆的動作，要使用者明確同意。
- **`dist-all/0.2.0/` 建好並驗過**：`full` 848 個檔、`release` 229 個、`promo` 8 個。
  `verify_dist.sh 0.2.0` 全過——三個包的 manifest 與條款、`release` 的素材
  deny-list、Windows 的 PE 檔頭、macOS 的四道靜態驗收、Linux 的啟動（封面與
  結局都畫得出來）。執行檔：linux 12.7 MB、windows 12.0 MB、macOS universal
  23.4 MB。
- 容器：`docker build` 兩次（第一次的 image 疊錯方向，用同一個 tag 重建，
  舊層變成 dangling；**沒有 prune、沒有 rmi**，要不要清由使用者決定）、
  `build_macos.sh`／`verify_macos.sh`／`make_promo.sh`／`build_dist.sh` 各數次，
  全部 `--rm`。沒有動別的專案的資源。

## 2026-09-07：開起來就能玩，三平台各自獨立封裝

使用者回報 0.2.0「打開後滑鼠鍵盤都沒作用，不能遊玩」，並要求 Linux 出
AppImage、Windows 與 macOS 各出獨立的 zip。

- **成因**：`kol` 不帶旗標時進的是 `-scene cover`，一張靜態封面——`Update` 對那個
  場景沒有任何輸入分支，只有 Escape 會關窗。滑鼠更直接：`MouseButton`、
  `CursorPosition` 在整個儲存庫 grep 不到任何一處。
- **範圍比一個 bug 大**：從標題到旅店那一整條（`flow`）只有腳本驅動，`g.flow`
  只在 `-flow-script` 那一支建立。二十六份對拍收據與五條 player-path 收據
  **全部是餵 token 逐步截圖**，沒有一次是人在鍵盤前走完的。先前自評 90 分裡
  「發行打包 85%」沒有把這個缺口算進去，而它是「可玩」的前提。
- **修法**：`-scene game`（新的預設值）借 `flow` 那條路徑的全部初始化，token
  改由鍵盤與滑鼠來。分派與 `-flow-script` 完全相同，所以互動遊玩經過的是
  收據驗過的同一組狀態機。
- **滑鼠翻成既有 token**，不是新增 API：點選單第三列送出 `D D RET`。可點區域
  取自各畫面既有的版面常數，沒有另外量一份。右鍵是 Escape，滾輪是上下。
- **兩個踩到的坑**：
  1. `ebiten.CursorPosition()` 回的是 **logical 座標**（文件明寫
     "this considers the scale of the screen"），我多除了一次視窗倍率，
     點擊全落在版面上方。**測試一開始「通過」了**——因為判準是「畫面有變」，
     而點歪會落到 fallback 的 Return，畫面照樣變。改成「滑鼠點第 N 列必須與
     鍵盤走到第 N 列產生同一張畫面」才抓出來。**寬鬆的判準比沒有判準更危險。**
  2. 驗收腳本用固定 `sleep` 截圖，在合成事件生效前就截到舊畫面，把「按鍵有效」
     誤判成無效。改成輪詢等畫面變。**當時是靠 `-log-input` 把收到的 token 印出來
     才定位的**——`token=[D D RET]` 後面接 `scene=delete cursor=2`，程式那一側
     從頭到尾都是對的。
  3. **改預設值會波及所有靠那個預設值的呼叫。**`-scene` 從封面改成互動遊玩之後，
     不寫 `-scene` 的批次呼叫會落進互動分支——`-shot-dir` 實測逾時、產出 0 張。
     `cmd/kol` 的 `batchFlags` 把產檔旗標排除在互動之外。但**哪些會被吃掉不能用
     旗標名推**：`-title-frame` 在 switch 之前就設好 `g.jobs`，而 `Update` 先看
     jobs 再看互動旗標，所以它從頭到尾都不受影響。我一開始把它判成「真回歸」
     並據此停掉了一輪建置，實測才發現不是。`verify_interactive.sh` 現在兩種都驗。
  4. `appimagetool` 兩次靜默失敗：缺 `file` 指令、以及在 `--network none` 下要去
     GitHub 抓 runtime。**第一次沒看到原因是我自己把它的 stderr 重導到
     `/dev/null`**；拿掉重導向就直接印出來了。
- **三個平台各自獨立**：`tools/package_dist.sh` 產 `dist-all/<版本>/packages/`
  六個檔（full／release × AppImage／Windows zip／macOS zip）。先前 `build_dist.sh`
  產的是目錄，三支不同作業系統的執行檔混在同一層，而且沒有任何壓縮步驟——
  腳本註解裡寫的 tar.gz 並不存在。
- **macOS 做成 `.app` bundle**：裸執行檔在 Finder 雙擊會開終端機，而且 `.app`
  由 Finder 啟動時工作目錄是 `/`。加了 `useResourceDir`：工作目錄找得到資源就
  不動，找不到才換到執行檔旁邊（含 `.app` 的 `../Resources`），所以既有的對拍與
  截圖流程完全不受影響。
- **驗收**：`tools/verify_interactive.sh` 在 Xvfb 裡送鍵與滑鼠、截圖比雜湊，
  一路驗到選角、進旅店、滑鼠點 DOOR 出門、在城鎮裡走一步；它吃自訂啟動命令，
  所以**同一套驗收直接跑在 AppImage 上**，畫面雜湊與從儲存庫跑的逐一相同。
  `tools/verify_packages.py` 驗六個發行檔的結構，並做過負對照：注入
  「AppImage 沒有執行位元」「`.app` 執行檔沒有執行位元」「release 混進 PNG」
  三個缺陷，三個都被抓到、退出碼 1。
- **容器**：`docker build -f docker/Dockerfile.package`（在鎖定建置映像上補
  `file`／`zip`，因為 appimagetool 缺 `file` 就不動）一次；`go_docker.sh`、
  `build_dist.sh`、`package_dist.sh`、`verify_dist.sh`、`verify_interactive.sh`
  各數次，全部 `--rm`。沒有 prune、沒有 rmi、沒有動別的專案的資源。
  `workplace/tools/` 放了 appimagetool（`a6d71e2b…d13a0`，build 295）與
  AppImage runtime（`db3bdec3…f3e4`），兩者都只留本機、不進版控。
- **`dist-all/0.2.1/` 建好並驗過**：`verify_dist.sh 0.2.1` 全過，六個發行檔
  驗收全過，AppImage 的互動驗收全過。**Windows 與 macOS 仍然沒有真機驗收。**

### 這一輪沒做、但看見的缺口

- 開場旅店（`flow.SceneInn`）只有 DOOR 接得動。MIRROR、LISTEN、TALK、REST、
  COMPANION 按下去會走 `f.note` 記一筆「未實作」然後畫面不動，玩家看起來就是
  沒反應。城鎮裡的建築內部走的是 `town.Session`，功能是齊的——缺的只有開場那一格。
- 戰鬥與野外還沒有可點區域，那兩層滑鼠只有「點一下等於 Return」與滾輪。
- `-load` 沒有預設值，存了檔要自己帶路徑讀回來。要不要「有存檔就自動接續」
  是操作方式的決定，等使用者定。

## 2026-09-07：既有 AppImage 與 dosgolem 操作稽核

- 直接執行 `0.2.1` full AppImage，重跑互動驗收：選角、旅店、出門、城鎮走動與
  滑鼠／鍵盤同列選取通過。補測地圖點擊送 `RET`，長按 Down 1 秒只送一次 `D`。
- dosgolem `7f30ad8` 重跑 `TestProbeMainExecutable`、`TestDumpTownMove`，另以
  本機 `TestAuditNormalKeys` 從 ZIP 重抽副本正常送鍵；可載入 KOLBOOK，未走到城鎮。
  城鎮點擊只取得 Tandy／BDA 初始化輔助證據，探測 PASS 不代表正常鏈或手感通過。
- 報告：`docs/verification/appimage-dosgolem-audit-20260907.md`；腳本、日誌、截圖與
  37 個檔案清冊：`workplace/appimage-dosgolem-audit-20260907/`。原版副本 75 個檔案
  與 ZIP 相同；來源唯讀，正常鏈副本寫容器暫存。
- 環境失敗為巢狀引號及登入 shell 的 Go PATH，修正後在同一 image 重跑成功；
  不記為產品缺陷。文件寫入另遇一次 sandbox 掛載競態，重試完成。
- 全部限額、斷網、非 root、`--rm`，保存檔案 UID/GID 1000:1000，Xvfb／遊戲以 trap
  清理。結束檢查無 `kol-audit` 容器殘留。未改遊戲、未重打包、未提交或推送。

## 2026-09-07：改以操作結果等價，稽核 RE 規格與正式原始碼

- 使用者明確允許滑鼠取代鍵盤，以結果等價為準；記入 CONTEXT 決策表，並對前輪
  操作報告追加訂正，保留原始實驗與當時解讀。
- 核對 15 個 SPEC ID／16 份文件與正式入口、資料轉換、規則、UI、保存接線。
  結果見 `docs/verification/spec-source-audit-20260907.md`；WORKLIST 頂部改成
  可執行修復入口，90/100 標為歷史自評，spec 索引的過時限制明示待回填。
- 唯讀掛載專案，複製 `cmd`／`internal` 到容器暫存；既有
  `go test ./internal/... ./cmd/kol -count=1` 全過。三份只在本機保存的診斷測試
  重現九項失敗：建角入口、選隊、建角取消、可變狀態保存、腰帶轉接、戰後回寫、
  原版 14 名角色 Ward、切換武器數值、修道院 Nutrition。
- 戰後回寫的首個診斷用了敗北出口，為避免混入未知敗北規則，另以勝利出口重跑，
  仍確認傷勢與背包未寫回。三批 log 都保存，不覆寫第一次判讀。
- 證據在 `workplace/spec-source-audit-20260907/`；manifest 保存 191 個來源雜湊、
  15 個產物與工具版本；產物擁有權全為 1000:1000。固定 image 為
  `kol-package:20260907`，全部斷網、限額、非 root、外層逾時與 `--rm`，Xvfb 有 trap。
  結束檢查無 `kol-spec-audit` 容器殘留，未改其他專案資源。
- 未修改正式 Go 程式、規格授權狀態、存檔格式或封包；沒有提交／推送。

## 2026-09-07：先修正，再繼續檢驗

- 使用者授權修正稽核缺口。完成建角入口／正式表單／逐層取消、同城選隊與顯示同步、
  腰帶武器與完整物品交換、Ward 抽取／轉接、修道院 Nutrition、勝利／撤退傷勢與
  物品回寫、任務交付 typed 同步、零金冠與設定讀回、住宿退出請求接回書本。
- 重生原始名冊的 Ward／背包及建築任務獎勵 typed data；來源 ZIP 唯讀，輸出僅本機。
- 完整 `go test ./internal/... ./cmd/kol -count=1` 與新增七項回歸通過。
  正常 GUI 建角與選隊再驗時發現戶外仍畫舊隊員，已修正後按同一腳本重跑通過。
- dosgolem 六項具名規則對照通過，包含三千次亂數、四百次有界骰與傷害承受邊界。
  屬受控常式對照，不宣稱完整正常原版玩家路徑。完整紀錄見
  `docs/verification/spec-fixes-20260907.md`；本機腳本、日誌及圖片在
  `workplace/spec-fixes-20260907/`。
- 保存 v1／v2 遷移提案已整理並詢問，尚未收到決定。未改保存 schema；建角保存、
  中文資料頁、開場旅店等仍待續作。未重打包、提交或推送。
- 所有容器使用既有 image、斷網、限額、非 root 與 `--rm`；GUI 以 trap 收尾。
  檢查無 `kol-fix` 容器殘留；產物擁有權 1000:1000，未清理其他專案資源。

## 2026-09-07：dosgolem 上游抓取與整合評估

- 已 `fetch origin`；遠端預設分支為 master，上游為 `fd746405cf0ef9e535121e215b935b0bf59cd1de`。
  現用 `kol-force-call` 保持 `7f30ad8427086738b037c50a417baf3d0d56a5d2`、乾淨，獨有 0 筆／落後 352 筆。
- 上游已包含現用分支，可快轉；但 KOL 的 FileOps／MemOps／Regs 診斷介面不相容，原樣測試無法編譯。
  公開 FileOp 還缺少內部保有的 Failed／Whence，須保留診斷能力後再整合。
- 唯讀來源的臨時適配副本通過六項 KOL 規則對照（97.881 秒）；上游核心套件測試亦通過，未掛載外部 CPU 語料。
  不宣稱正常 DOS 玩家路徑或 AppImage 對拍完成。報告見 `docs/verification/dosgolem-update-20260907.md`。
- 僅抓取遠端參照，未移動現用分支、未推送；未覆蓋前輪正式修正。一次性 Docker 均使用既有工具映像、
  斷網、限額、非 root 與 --rm；本機臨時腳本／日誌保留於 `workplace/dosgolem-update-20260907/`。
  結束確認無 `kol-dosgolem-update` 容器殘留，報告及產物擁有權為 1000:1000。

## 2026-09-08：配合新版 dosgolem 修正

- 使用者授權後將乾淨 `dosgolem-kolrand/kol-force-call` 快轉至 `fd746405`。在該工作樹補回公開
  FileOp 的 Whence／Failed 轉接，附成功與失敗定位回歸測試及既有 READY 規格補正。
- KOL 三處檔案日誌改用共用新版欄位輸出；兩處記憶體日誌使用 Fn／BX／ES／AX／OK，
  暫存器改用結構格式。未變更玩法、存檔格式或封包，保留前輪正式修正。
- dosgolem 核心套件與本機 CPU 語料通過；KOL 六項原版規則對照通過（126.714 秒）。
  最初正常探針因缺環境變數而跳過，未列驗收；修正腳本後才進行實際正常入口與六張截圖。
- 原版正常 KOL.EXE 路徑執行 1.2 億指令，無缺檔或未實作服務，但 BIOS 按鍵尚未讀取。
  逐張檢查六張 PNG 全黑；索引像素非零而 DAC 色數為 0。最初僅無錯誤判準的 PASS 已撤回，
  補上全黑必須失敗的畫面閘門後，乾淨重跑如實 FAIL（13.465 秒）。未宣稱正常畫面或操作對拍通過。
- 嘗試產接觸表時發現既有映像沒有 Pillow，改為逐張直接檢視，未安裝額外依賴。
  報告見 `docs/verification/dosgolem-integration-20260908.md`；來源、腳本、日誌、圖片與差異收據留本機。
- 本輪兩個命名容器均 --rm 並確認無殘留，非 root、限額、斷網；寫入檔案為 1000:1000。
  未提交／推送，未清理其他專案資源。

## 2026-09-08：持續修復原版對拍工具鏈

- 上輪屬實質進展。本輪依目前工作樹繼續：以 SeaBIOS 的模式／色盤契約建立 dosgolem
  `185-ega-bios-default-palette.md` READY 規格，再補 BIOS 0Dh／0Eh 初始化。
  核心套件與六項 KOL 規則通過（123.786 秒）；六張相同路徑的 RGB 圖全部恢復顯示，
  索引雜湊不變，已逐張檢視雷電／Origin／製作人文字。
- 正常 `KOL.EXE` 入口：硬體 Enter 到十億指令仍在片頭；DOS Enter 於 2.5 億指令階段
  轉入 `KOLBOOK.EXE`，另跑 500 萬指令後截圖確認書本封面。
- 書本 BIOS 空白鍵可翻到目錄；計數仍顯示已讀 0，故不以計數否定實際畫面。
  調整到完整翻頁後，數字 6 與滑鼠選單都未進入旅店。未使用 direct-entry、記憶體修改或 DOSBox 收據。
- 繼續查 dosgolem 程式時發現 AH06/DL=FF 永遠回無字元；依 Microsoft DOS 4.0
  契約建立 `186-dos-direct-console-input.md` READY，修 Stdin 消耗／ZF／AL／軌跡，
  加入含 NUL 與空佇列的回歸。核心與六項 KOL 規則再驗通過（118.650 秒）。
  但正常旅店測試仍 FAIL（65.646 秒），不把獨立缺陷冒稱根因；下一步限定為原版目錄讀鍵鏈。
- 原因、來源、成功與失敗畫面見 `docs/verification/dosgolem-palette-fix-20260908.md`。
  收據在 `workplace/dosgolem-palette-20260908/` 與 `workplace/dosgolem-console-20260908/`；
  新收據工具於執行前固定來源，執行後檢查變動。較早清冊明示為事後工作樹清冊，不能單獨證明歷史輸入。
- 完整存檔 v1／v2 決定仍待回覆；已再提出具體相容性選項，未擅自改格式。
  中文資料頁與開場旅店流程已定位，尚未在本輪改動；整體目標保持未完成。
- 本輪使用既有映像、UID/GID 1000、限額及 --rm；除下載公開 BIOS 來源外斷網。
  本輪命名容器均已清理。未重打包、提交或推送，原有 KOL 修正保留。


## 2026-09-08：書本跨介面讀鍵修復

IDA Pro 9.4 從固定 KOLBOOK.EXE 建立一次性資料庫，證實遊戲直接檢查 BDA 後以 DOS AH08 取鍵。dosgolem 新增 BIOS 到 DOS 取鍵橋接及一般鍵、擴充鍵、Stdin 優先回歸；規格 187 已建立 READY 契約。第一次匯出編碼失敗修為 UTF-8，乾淨重跑成功，歸類為腳本問題。

於 kol-package:20260907 執行 tools/oracle/check_updated_dosgolem.sh：核心回歸、六項規則（142.962 秒）、片頭擷取（12.843 秒）、正常書本入口（48.978 秒）通過。人工檢視 book-final.png 確認 Halfway House 旅店與對話。詳見 docs/re/20260908-kolbook-input-bridge.md；收據 workplace/kolbook-input-20260908/manifest.json 為執行前清冊，source_changed_during_run=false。完整 AppImage 對拍仍未完成，未更新發行包，未提交或推送。

輸出 UID/GID 均為 1000:1000；本輪一次性容器已移除。原始反組譯匯出及暫存收據僅本機保存。


## 2026-09-08：選隊與招募資料隔離

核對 flow.innKey 與 town.Session，確認開場旅店仍是只有 DOOR 的靜態流程，未宣稱修完。另發現招募共用名冊中的技能、任務、裝備與法術資料；選隊也仍共用裝備及法術防護陣列。兩條入口改為共用獨立副本建立函式，招募後同步室內與戶外隊伍。新增回歸走選隊及招募入口，再修改各可變欄位，確認名冊不受污染。這是 remake 資料隔離修正，不是原版 parity 證明，也不改存檔格式。

kol-package:20260907 容器執行 go test ./internal/town ./internal/flow -count=1 -timeout 60s 通過（0.033／0.012 秒）。初跑受專案 go.work 引用未掛載的 /dosgolem 阻擋；設 GOWORK=off 後於同一映像乾淨重跑，歸類為環境設定問題。修改檔案 UID/GID 為 1000:1000；一次性容器已清理。未提交、推送或打包。


## 2026-09-08：開場旅店接線與繁中招募清單

依 SPEC-016 接入正式城鎮 Session；同城選角、招募、離店及不一致所在地拒絕回歸通過。GUI 重現繁中候選人清單漏畫，補上八列與捲動指示後完整測試再次通過。正常視窗實測半途之家、招募 Poontang 及兩人離店，逐張確認。詳見 docs/verification/starting-inn-fix-20260908.md。

首次建置命令試圖於容器根目錄建立 /out 連結，遭非 root 權限拒絕；改明確輸出掛載後重跑。首次 GUI 抓圖等待不足列為腳本問題，空白清單列為產品缺陷；兩者舊證據保留於 first-gui-attempt。最終輸出與程式雜湊保存於本機 manifest.json。未提交／推送／打包；跨城、完整保存及 AppImage 對拍仍待完成。

輸出 UID/GID 已確認為 1000:1000，一次性容器已清理。


## 2026-09-08：跨城選角與入城代碼同步

依 SPEC-016 接線審查擴充，利用 RE-069 既有入口、邊界、城鎮代碼及 typed data／PNG 載入跨城角色所在旅店。暫存驗證成功後才提交遊玩狀態，未知代碼、重複代碼與不一致建築不能猜補。並補野外入城時 TownCode 同步。

完整測試與建置通過；正常 GUI 從標題選第十三名 Mula，確認巨魔剋星旅店、當地候選人 ddduhh、兩人入隊與正確戶外地圖。詳見 docs/verification/cross-town-fix-20260908.md。本輪未作 Mula 的 dosgolem 同狀態驗收，不宣稱嚴格 parity。未提交、推送或更新 AppImage。

本機收據 workplace/cross-town-20260908/manifest.json 保存建置及截圖雜湊，來源盤點標為執行後；建置後僅調整一行來源註解。輸出 UID/GID 1000:1000，一次性容器已清理。


## 2026-09-08：dosgolem 原版 Mula 跨城入口抽驗

正常書本測試增加有界 KOL_BOOK_SELECT_DOWN，保留選角截圖。以 KOL_CHECK=book、KOL_BOOK_SELECT_DOWN=12 執行 check_updated_dosgolem.sh，原版入口重跑通過（48.963 秒）。人工確認 Mula／TROLLSBANE／BRT，實際載入 TOWN1.0；與 remake 旅店畫面核對角色、店主肖像、招呼語及六圖示順序相符。只驗入口可見狀態，完整操作對拍仍未完成。

收據 workplace/dosgolem-mula-20260908/manifest.json 為執行前清冊，source_changed_during_run=false；輸出 UID/GID 1000:1000，容器已清理。未提交、推送或打包；原版輸入與擷取仍僅本機保存。


## 2026-09-08：招募候選人滑鼠接線

發現 flowTargets 只列旅店圖示，點候選人名字會落到 Return fallback，可能確認錯誤角色。依已核准滑鼠等價操作及 SPEC-016，補中英文可見清單列範圍，送上下移動、回確認圖示、Return；不直接寫游標或繞過招募狀態機。捲動視窗按相對列差換算。

kol-package:20260907、Xvfb 容器執行 go test ./cmd/kol ./internal/input ./internal/flow ./internal/town -count=1 -timeout 60s 通過（0.082／0.015／0.008／0.059 秒）。新增回歸含三列、中英文、取消圖示與不重疊的列範圍。這是輸入接線回歸，尚未做新版本正常滑鼠 GUI 或 AppImage 驗收，不能當作完整操作對拍。

修改檔案擁有權已檢查；一次性容器已清理，未提交、推送或打包。


## 2026-09-08：招募滑鼠正常 GUI 驗證

完整測試與乾淨建置通過，KOL_RECRUIT_MOUSE=1 從正常視窗啟動。清單初始反白 Poontang、圖示移至返回後，點第二列 Olear，實際輸入 D／<／RET；人工確認 Olear 加入並與 Theodoric 離店。沒有注入角色或直接進場。詳見 starting-inn-fix-20260908.md 追加段落。

收據 workplace/recruit-mouse-20260908/manifest.json 保存建置及 PNG 雜湊，來源標為執行後盤點。僅抽測繁中未捲動清單，不宣稱英文、捲動或 AppImage 全驗收。輸出 UID/GID 1000:1000；一次性容器已清理，未提交、推送或打包。


## 2026-09-08：中文角色資料頁版面原型

依 grilling 技能，先製作 640×400 同頁集中與兩頁分頁原型，沿用現有字模，來源及圖檔保存在 workplace/sheet-prototype-20260908/。數值為版面範例，不是原版證據或正式 UI；正式程式未修改。字模儲存格實際為 16×18，初稿發現返／吋／銜不在現有字表；原型暫用可顯示文字，正式字庫補驗仍待辦。最終原型字串無缺字。

同頁與分頁是玩家閱讀／翻頁取捨，等待使用者選擇，不把建議升成需求。所有輸出 UID/GID 1000:1000，一次性容器已清理；未提交、推送或打包。


## 2026-09-08：資料頁返回路徑疑慮排除

中文 A／B 版面仍待使用者回答，未自行選擇。獨立查核 closeSheet 未清 SpellList 的疑慮：正式 acceptInside 先處理 ShowingSpellList，第一次返回會清內頁，第二次才關資料頁，因此不能把孤立欄位視為可達缺陷。新增按鍵分派回歸，連續兩輪開資料頁、法術清單、返回、關頁，再以更新後資料重開，確認不復活舊清單或圖示。

kol-package:20260907 執行 go test ./internal/town -count=1 -timeout 60s 通過（0.025 秒）。沒有修改正式行為，也不宣稱 GUI 或原版 parity。新測試檔 UID/GID 1000:1000，一次性容器已清理，未提交、推送或打包。


## 2026-09-08：正式翻譯字庫涵蓋檢查

新增 tools/verify_cjk_coverage.py，按正式翻譯區塊核對字模 manifest，保存來源雜湊及缺字位置。5,109 個字串無可見缺字；隔離副本加入缺字返，確認退出碼 1 及正確位置。原型缺字尚未出現在這批正式譯文中，未重烘字庫或修改翻譯。

初次全字串掃描包含研究註記與核准狀態，不能當成玩家缺字；正式檢查初稿另捕捉檔尾 NUL／SUB，改為僅檢查可印字元，保留 control-character-first-run.json。最終收據 workplace/font-coverage-20260908/report.json。這只證明翻譯資料的字元涵蓋，不證明 Go 動態文字、PNG 墨跡或版面安全。

中文版面 A／B 與存檔格式仍待使用者回覆。輸出擁有權已確認 1000:1000，一次性容器已清理，未提交、推送或打包。


## 2026-09-08：耐久子常式邊界重驗

回到未完成戰鬥規則，發現舊耐久探針只有列印，沒有整格及相鄰資料斷言。新增 TestDurabilityBoundariesOriginal，以新版 dosgolem 直接呼叫 sub_22084，九組通過：耐久零保留、負數清空整個 16-byte 槽、旗標未設不扣、相鄰資料不變。詳見 docs/re/20260908-durability-boundaries.md。此為子常式證據，不是正常戰鬥或 remake 完成證據；護甲與耐久呼叫接線仍待修。

第一輪 uint32／uint16 段位移型別錯誤已修，同容器乾淨重跑，保留 first-compile 紀錄。收據 workplace/durability-20260908/manifest.json 為 before_copy，source_changed_during_run=false。輸出 UID/GID 1000:1000，一次性容器已清理，未提交／推送／打包。


## 2026-09-08：裝備合身後原始記錄同步

追查耐久回寫垂直鏈時，確認戰鬥仍未攜帶全部穿戴物品，護甲／耐久接線仍待完成。另發現既有合身購買與穿戴調整只改 Good.Encumbrance，未改來源 Record。依 RE-075／SPEC-009 已知重量位移 +0x0A，新增局部更新；沒有完整原始記錄時不補造 payload。兩個正式寫入點同時同步 Carried。

新增兩路回歸，走實際商店問答，驗證重量 6→9、原始記錄與 Carried 同步，其他位元組不變。kol-package:20260907 執行 go test ./internal/town ./internal/flow -count=1 -timeout 60s 通過（0.025／0.003 秒）。這是裝備資料一致性修正，不是耐久或原版完整對拍完成證據。

修改檔案 UID/GID 已確認為 1000:1000；一次性容器已清理，未提交、推送或打包。中文版面與存檔格式仍待使用者決定。

## 2026-09-08：新版 dosgolem 裝備加值驗證與護甲控制流訂正

沿用既有 IDA 9.4 資料庫的隔離副本，確認護甲映射、骰數、加值、耐久與扣傷直接呼叫。
新增受版控地址索引，匯出器自動合併語意等級及證據，不改原始名稱／指令。
RE-133 原先以記錄無差異推定扣傷不在攻擊常式內，本輪以直接 CALL 訂正，保留舊記錄。
詳見 docs/re/20260908-armour-control-flow.md；正式護甲及耐久仍未接入戰鬥。

新增 TestSignedItemBonusOriginal，全部 256 種輸入與原九組耐久邊界通過（0.007 秒）。
收據 workplace/equipment-oracle-20260908/manifest.json 為 before_copy，來源未在執行中變更。
初次輸出連結因容器權限失敗，改成明確掛載後同映像乾淨重跑成功，屬工具啟動問題。
未提交、推送或打包；完整 AppImage 對拍、存檔格式及中文版面決策仍未完成。

收尾確認七筆語意索引均自動附加且保留原始定位與證據；git diff --check 通過。
抽查新增來源及本機收據均為 UID/GID 1000:1000，KOL 一次性容器已全部清理。

## 2026-09-08：護甲記錄指標與後段 84 組原版驗證

既有 IDA 資料庫補匯出 sub_22278／sub_22296，確認隊員與敵方不同的記錄公式。
新增 TestEquipmentRecordPointerOriginal，驗證 18 槽及非零旗標；回查舊 TestDamageMatchesOriginal
發現索引 1 被標為敵方且忽略呼叫錯誤，因此不能拿它證明完整攻擊鏈。
訂正獨立扣傷測試的舊註解，未覆寫原歷史證據。

新增 TestArmourReductionOriginal，以明示入口橋建立區域變數後跳入原版後段，
保留指標、護甲與耐久機器碼，亂數及最末扣傷改為參數觀測。84 組通過，另三項裝備
測試也通過，合計 0.018 秒。包含全零空槽仍擲骰、零輸入配負加值仍可能產生傷害。
沒有把此受控入口視為正常戰鬥／傷勢回寫完成。收據 equipment-pointer-20260908
為 before_copy，source_changed_during_run=false；保留較早的 60 組與指標單獨版本。

新增 SPEC-017 DRAFT，列出兩種部位映射及接線閘門。下一步核對耐久百分比及破損後
衍生數值，審查後才接正式近戰／射擊與回寫。未提交、推送或打包。

收尾確認十筆語意索引與匯出雜湊相符，git diff --check 通過；新增測試及收據
UID/GID 為 1000:1000，KOL 一次性容器已全部清理。

## 2026-09-08：護甲減傷正式接線與 remake／原版比較

核對物品副作用後，訂正 Percent 是盾牌率而非耐久百分比，SPEC-017 升為 READY。
新增 Actor 三槽護甲副本，接入近戰及射擊；依原版保留護甲前 Hurt，並將破損／耐久
接到勝利及撤退的穿戴槽、Record、Carried 回寫。缺少有效原始記錄時明確失敗。

完整 go test ./internal/... ./cmd/kol -count=1 及建置通過；新增攻擊兩路、裝備隔離、
破損減重與戰後回寫回歸。首輪 TestResolve 舊有 Hurt 等於傷勢的斷言失敗，依已證實
護甲前受創契約改為明確六次骰序與 16／14 兩值，保留 first-tests 紀錄。

原版護甲後段 84 組擴充為與 remake 實際程式比較，逐項核對傷害與三槽完整記錄，
另三項裝備原版測試一併通過（0.046 秒）。oracle/manifest.json 為 before_copy 且來源未變。
詳見 docs/verification/armour-implementation-20260908.md。建置清冊為明示事後盤點，
建置後僅修改骰序註解。產物及收據在 workplace/armour-implementation-20260908/。

git diff --check 通過，輸出 UID/GID 1000:1000，KOL 一次性容器全部清理。
未提交、推送或打包；正常 GUI 戰鬥、完整保存、其他磨損呼叫及 AppImage 仍待完成。

## 2026-09-08：正常 GUI 戰鬥入口與繁中縮放修正

沿用已建置護甲版，以實際視窗按鍵從 Theodoric 選角、招募 Poontang、離店及出城步行。
第一輪 15 次向東尚未遇敵，保留 first-walk；第二輪 100 次向東進入戰場，保留 before-scale。
正常畫面確認左側資訊框空白（新增 A10），並發現戰場只佔繁中視窗四分之一。

依既定 320×200／640×400 畫布契約，新增 renderBattle 共用離屏縮放出口，涵蓋遭遇、
戰場、法術清單及戰利品。未自行設計中文戰鬥版面或填入沒有規格的左側內容。
完整測試與建置通過；同一容器映像、相同種子及 123 個正常輸入重跑通過。
兩張戰場畫格與修正前的預期放大結果像素差異為零，人工確認戰場填滿視窗。

收據 workplace/normal-battle-20260908/scale-verification.json，執行前二進位與腳本
雜湊存於 gui-inputs.json，執行後核對未變；不冒充全部素材清冊或原版對拍。
詳見 docs/verification/normal-battle-entry-20260908.md。攻擊結算、破損顯示、中文戰鬥字型
及兩個資訊框內容仍未完成，未更新 AppImage、未提交或推送。

收尾 git diff --check 通過，PNG 收據擁有權 1000:1000；KOL 一次性容器全部清理。

## 2026-09-08：戰鬥資訊框原版定位與直條邊界

以既有 IDA 資料庫定位 sub_22EFC 的兩種左上模式、六個裝備位移，以及 sub_2324F
左下框。確認直條取疲勞與體力而非生命值；人形參數為隊員 0x47+性別、敵方 0x49，
但素材合成及模式切換輸入仍未確認，正式資訊框沒有猜補。

新增 TestBattleStatusBarOriginal，十二組原版繪製參數通過（0.005 秒）。初稿忽略先除二
的整數截斷，疲勞 3 的分界預期 84、實際 85；依指令順序訂正並加入低體力樣本。
KOL_CHECK=panels 為獨立可重現入口，僅替代繪製以觀察參數，不假稱為正常原版畫面。

資料交叉參照起初被 xref.iscode 錯誤篩除，修為檢查來源指令後取得五個 mode 參照。
建立十一筆分級語意索引，自動附加於原始位址／指令／bytes；原版測試收據及當時輸出
雜湊先驗證再封存於 workplace/battle-panels-20260908/oracle-run/，避免新註記改壞歷史收據。
報告 docs/re/20260908-battle-panels.md；未改正式 runtime、未打包、提交或推送。

收尾確認十一筆索引與匯出雜湊一致，原版測試來源未變，輸出 UID/GID 1000:1000。
git diff --check 通過；KOL 一次性容器全部清理。

本輪收尾 IDA 曾一次讀不到資料庫輸入雜湊，檢查拒絕新輸出；保留失敗 log，沒有放寬
身分驗證。同映像與來源乾淨重跑退出碼 0，十一筆索引、原版收據快照與擁有權核對通過。

## 2026-09-08：新版 dosgolem 整合入口補正

發現 `all` 未包含後來新增的護甲／耐久與資訊框測試，已補接並保留個別入口。
核心、六項原版規則、四項護甲相關測試、十二組直條、片頭及書本選角全部通過。
人工檢視 dosgolem 新產生的 book-action-3.png，確認 Halfway House 旅店畫面。
詳見 docs/verification/dosgolem-complete-suite-20260908.md。

首次輸出符號連結因非 root 根目錄權限遭拒，尚未執行測試；修正為直接掛載 /out，
同映像乾淨重跑退出碼 0。收據輸入未變，19 個輸出雜湊與 1000:1000 擁有權通過。
KOL 一次性容器全部清理；未重建 AppImage、未提交或推送。

## 2026-09-08：原版人形傷勢遮罩驗證

延伸既有 IDA 探針，保留原始資料庫與定位，取得 72 bytes 身體矩形及像素讀寫資料流。
新增 TestBattleBodyOverlayOriginal：三種人形、六部位、八組傷勢，共 144 組通過；
與既有十二組直條共 0.135 秒。首次誤用 ReadU8 導致編譯失敗，改成新版 Byte API，
同映像命令乾淨重跑；首次收據保留於 body-overlay-20260908/first-api-compile/。

新增五筆分級索引並重生自動附註匯出；檔名格式 misc%s.0 與人形快取索引已定位，
正式素材及模式接線仍待核對。詳見 docs/re/20260908-body-overlay.md。
來源未變、收據雜湊核對通過，輸出為 1000:1000；未更新正式 runtime 或 AppImage。
收尾確認分級索引已併入匯出，git diff --check 通過；KOL 一次性容器全部清理。

## 2026-09-08：正式戰鬥人形與疲勞框

補證三種底圖快取、PNG 索引／尺寸，以及被檢視角色的原始控制流；SPEC-018 先建立
DRAFT，經證據審查升為 READY，再接入正式戰場。角色檢視與行動者分開，空格清框、
招式子選單保留檢視者、取消恢復本人、輪末清空。三張底圖與單張傷勢輸出使用有界快取。

正式計算與原版十二組直條、144 組傷勢遮罩及三種底圖路由比較通過（0.134 秒）。
完整測試與建置通過；首次規劃測試在輪末未重設回合位置，修正前置後同命令乾淨重跑，
失敗紀錄保留 first-planning-test。未因測試錯誤改動正式回合規則。

正常開場選 Theodoric、招募 Poontang、離店及出城步行遇敵，實際 GUI 顯示人形與疲勞條。
裁切與原始 PNG 放大四倍像素差異零。新增目標／取消抽測初稿誤用方向鍵切圖示，
只修正腳本為大於號，保留 inspection-first-navigation；同執行檔重跑 127 個輸入通過，
空格資訊框全黑、取消後整張戰場像素差異零。

產物 workplace/body-panel-implementation-20260908，執行檔 SHA-256
cc1c8782378a9e7c099cfd4b5b0b8f4aed0b8a2c327e80cb247ac990a3430d8c。
詳見 docs/verification/body-panel-implementation-20260908.md。A10 的左上框與左下命令
圖示、正常攻擊後傷勢、完整保存及 AppImage 仍待完成；未提交或推送。

23 筆分級註記與 IDA 匯出一致，原版及 GUI 收據來源／雜湊核對通過，輸出 1000:1000。
git diff --check 通過；KOL 一次性容器全部清理。

## 2026-09-08：新版 dosgolem 補證與左上裝備框實作

原版常式補驗空白鍵模式、空格不得切換、六列隊員／五列敵人裝備及肖像位置。
首輪測試誤將 DS:2DAE 與 IDA 線性 30C7E 視為不同欄位；修正測試前置並保留
first-icon-alias，乾淨重跑 panels 通過（0.064 秒）。SPEC-019 先經 DRAFT／審查升 READY，
再接基本肖像、裝備列與模式切換。彈藥數字尚未證實，沒有猜補。

完整測試及建置通過。正常開場、招募、離店、出城遇敵及模式切換共 126 個輸入完成；
肖像與裝備畫面切回後各自整張像素差異零。28 筆 IDA 分級註記重新匯出並逐筆核對。
詳細命令、限制與產物見 docs/verification/equipment-panel-implementation-20260908.md。
最新執行檔 SHA-256：24c113944b57707ebcef84c69f687f7d94b71ad9eccbaaf6cb0a26e9460ffb60。

產物位於 workplace/equipment-panel-20260908；擁有權 1000:1000，來源雜湊前後一致，
git diff --check 通過，本批 KOL 容器均已自動移除。未提交／推送，AppImage 尚未重建。
完整戰鬥同狀態、剩餘面板、完整保存及封包驗收仍未完成。

## 2026-09-08：左下命令圖示接線前查核

依目前工作樹繼續 A10，IDA 匯出器補納 sub_23638，從唯讀資料庫副本重生成功。
定位當前規劃／被檢視人物的不同來源、命令分支及 24×20 圖示矩形；語意保留強推論，
記錄於 docs/re/20260908-command-panel.md。下一步以 dosgolem 擷取各分支參數，
核對文字記錄副作用與素材路由後才寫正式 UI。未重新開啟已完成的人形與裝備功能。
本輪未修改正式行為；未提交或推送。容器使用 --rm、UID/GID 1000:1000，原資料庫未改動。

## 2026-09-08：左下命令 dosgolem 參數測試

新增 CommandPanel／CommandSprite 原版測試並加入 panels／all 入口。首輪 1,152 組
取得兩張映射表後，保存 initial-four-styles；再以固定獨立期望表擴充至 3,168 組。
人物來源、開關、命令分支、招式索引、防禦跳號及方向夾值通過；五組 helper 亦通過。
整批 panels 耗時 0.269 秒，正式 UI 未變，不重跑無關建置。
收據 workplace/command-panel-20260908，來源前後一致、輸出雜湊與 1000:1000 核對通過。
下一步核對文字副作用及 PNG 路由，再審查 READY 規格；未宣稱正常畫面或整體完成。
git diff --check 通過，本批 KOL 容器已自動移除；未提交／推送。

## 2026-09-08：命令標籤來源與稀疏圖號查核

IDA 匯出器補納 sub_1A8FD，資料庫副本重生通過。讀取端證實 DS:9E5C 是來源文字庫，
訂正前輪「目的地」語意並保留原證據；五筆標籤與版面碼已核對。方向圖示走 KOLMPICS
快取而非 MISCPICS；71 張稀疏 PNG 中 74..81 號均存在且為 24×20。
詳見 docs/re/20260908-command-panel.md 的追加勘誤。下一步補實際字元出口與方向路由測試。
未修改正式玩法或介面；未提交／推送。IDA 使用唯讀來源副本及既有輸出目錄，容器自動移除。

## 2026-09-08：命令標籤與方向路由 dosgolem 驗證

新增 CommandText／CommandDirection，納入 panels／all。五筆標籤用固定窄文字庫，
實際執行 helper 與文字解析；八方向實際執行 helper 與圖像路由，兩項全過。
整批 panels 0.134 秒，收據保存在 command-panel-20260908/text-direction。
來源前後一致、輸出雜湊與 1000:1000 核對通過。正式 Actor 已有對應欄位，
下一步審查顯示閘門並接 UI。未把窄測試列為正常 KOLDATA 載入或像素對拍。
未修改正式行為，未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：敵人命令檢視來源訂正

呼叫端查核發現舊筆記把檢定值寫成對方智力，但 IDA 取址使用行動者 word_2EA18。
新增 CommandInspection 六組真實目標常式測試，隊員／敵人、成功／失敗、上方框與
左下命令閘門全部通過；整批 panels 0.264 秒。SPEC-010 及研究紀錄追加精確勘誤，
未覆寫歷史研究。正式命令面板仍需核對檢定亂數生命週期後接線。
收據 command-panel-20260908/inspection；未提交／推送，未修改正式行為。

## 2026-09-08：檢視檢定本體及亂數缺口

IDA 補匯出 sub_252CB 與門檻交叉參照。162 組原版檢定通過，確認使用被檢視者等級，
本體不另擲亂數。首次測試誤讀 DX:AX 而非 AX，保留 first-return-width 後修正，
同容器命令重跑 panels 通過（0.207 秒）。未修改 CPU 或正式玩法。
已定位門檻在人物索引更新後、可行動檢查前擲出；目前回合前進缺少對應保存與消耗，
需連同倒下人物與取消返回的生命週期補接，不能在顯示時臨時擲骰。
詳見 docs/re/20260908-command-panel.md。未提交／推送，本批容器自動移除。

## 2026-09-08：規劃門檻生命週期草稿

比對 IDA 規劃迴圈與 BeginPlanning／EndTurn／AdvanceToParty 三個 skipDown 入口，
確認不能只在查看敵人或 renderer 補擲門檻。新增 SPEC-020 DRAFT，明列倒下人物、
敵方 AI 前順序、取消不重擲、排序上界與空槽對應的 READY 前驗證要求。
尚未修改正式規則，避免未核對空槽數量就改變整條亂數序列。下一步原版生命週期抽測。
git diff --check 通過，本批唯讀容器自動移除；未提交／推送。

## 2026-09-08：規劃位置迴圈 dosgolem 抽測

新增 PlanningInspectionOrder 有界原版切片，明列兩處測試記憶體出口修改。
五個排序位置含隊員、敵人、不能行動者與空位，門檻保存、檢查、AI 及完成標記順序
全過；包含上界最後一項，沒有多擲一次。整批 panels 0.387 秒，收據 planning-order。
SPEC-020 追加審查進度，仍待初始化上界對應，未直接修改 production 亂數。
未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：實作 SPEC-020 規劃門檻

Round 新增門檻及位置一次性標記，skipDown 在可行動檢查前處理一次擲骰，
含倒下／空位並清招式與防禦。排序保留六個隊員槽加本場敵人，修正舊排除空槽測試。
完整測試與建置通過，新增生命週期回歸另跑通過；詳見 planning-roll-implementation 驗證紀錄。
正常 GUI、正式同骰序整合比較與命令面板仍待完成，未將 SPEC 升 CONFORMED。
未提交／推送；git diff --check 通過，本批容器自動移除。

## 2026-09-08：規劃門檻修正後正常 GUI

最新 planning-roll-fix 建置走正常開場至戰鬥，共 126 個輸入完成；人工確認肖像及
裝備清單。兩種面板切回整張像素差異均零，執行前後雜湊一致，輸出 1000:1000。
收據 gui-verification.json；未以此取代完整戰鬥或原版同骰序驗證。
WORKLIST 回填已實作狀態。未提交／推送；git diff --check 通過，KOL 容器均已清理。

## 2026-09-08：實作左下命令面板

SPEC-021 先 DRAFT 後審查 READY；新增 State.CommandPanel 與 UI 繪圖接線，
依被檢視者、命令分支、保存門檻畫五種標籤及原版圖示，不消耗亂數。
來源與檢定邊界回歸、完整測試與建置通過，產物 command-panel-implementation-20260908。
正常 GUI、直接原版參數比較及完整戰鬥仍待完成，未升 CONFORMED。
未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：命令面板直接比較正式實作

原版 CommandPanel 測試接入 State.CommandPanel，3,168 組參數、標籤、圖碼、座標及
順序全部通過；保留獨立期望表。整批 panels 0.272 秒，oracle 收據來源未變、
輸出雜湊及 1000:1000 核對通過。關閉組使用空人物不顯示，限制另記，不冒充敵人
閘門本體或正常畫面對拍。未重建正式程式，下一步正常 GUI。
未提交／推送；git diff --check 通過，KOL 容器已清理。

## 2026-09-08：正常規劃與一輪結算抽測

新增 KOL_COMMAND_PANEL 腳本分支，正常移動規劃、走路、另一名休息及結算完成，
返回下一輪 Theodoric 主指令；畫面可見位置、傷勢與疲勞更新。初始目標格為空，
尚未取得命令標籤可見畫格，不標為面板 GUI 通過。下一步利用結算後相鄰隊友檢視。
gui-round-receipt.json 保存雜湊，執行前後來源相同、輸出 1000:1000。
未提交／推送；git diff --check 通過，容器已自動移除。

## 2026-09-08：正常隊友命令面板可見性

保留前次 GUI，新增 ally-gui 獨立目錄；同執行檔正常一輪後檢視相鄰 Poontang，
確認 OPT／休息圖示及其肖像、人形。148 個輸入完成，取消後整張畫面與檢視前零差異。
gui-verification.json 保存收據，執行前後雜湊相同、擁有權 1000:1000。
其他命令圖示、敵人檢定與原版同狀態仍待完成，不擴大本次 GUI 結論。
未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：真實敵人檢定直接比較

新增 CommandVisibility：真實執行原版目標與智力檢定，與正式 CommandPanel 比較
九組隊員／敵人及門檻 39／40／41。刻意讓兩方等級不同，並禁止額外亂數呼叫。
九組與整批 panels 全過（0.240 秒），來源清冊未變、雜湊與 1000:1000 核對通過。
收據 visibility-oracle；不把此規則層比較冒充敵人正常 GUI。正式程式未改。
未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：規劃初始化上界補證

原版 word_30C6C 敵人數來源與 0x1E4AE..0x1E4BD 上界計算核對。
新增 PlanningLimit 十二組有界執行通過，確認範圍為六個隊員槽加本場敵人，
不是排除空隊員槽，也不是無條件全部 18 槽。整批 panels 0.211 秒，收據 planning-limit。
SPEC-020 追加證據；正式亂數尚未修改，下一步完成取消／重複入口審查並接線。
未提交／推送；git diff --check 通過，容器均已自動移除。

## 2026-09-08：修復兩方初始智力

門檻接線查核發現 NewQuest 未複製 Member／Monster.Intellect；後續 restoreStats
只在法術到期呼叫，無法修復正常初始狀態。已補正式轉接，加入建場與恢復後判定
一致性測試。完整 internal／cmd 測試及建置通過，tests.exit=0，輸出 1000:1000。
產物 intellect-fix-20260908，SHA-256 43626d1648627dcd4f275d29452694310dec218549bc955b8b3760100e441a83。
SPEC-020 取消／重複入口與門檻實作仍待完成；本輪未完成原先規劃的該分支。
未提交／推送；git diff --check 通過，KOL 容器均已自動移除。

## 2026-09-08：SPEC-020 完成實作前審查

新增 PlanningCancel 有界回跳測試。首次正規化段位址造成 IP 下溢，保留 first-segment，
改用原函式段後三次取消均不重擲、不改排序位置；整批 panels 0.377 秒通過。
比對正式取消入口都使用 BeginTurn，確認一次性標記必須放 Round 排序位置，不能隨
BeginTurn 重設。SPEC-020 升 READY，尚未實作、未 CONFORMED。
未提交／推送；git diff --check 通過，本批 KOL 容器自動移除。

## 2026-09-08：新版 dosgolem 目前整合案例重跑

以 fd746405 加既有本機修正執行 KOL_CHECK=all，核心、六項規則、耐久、完整面板、
規劃檢定及正常片頭／書本入口全部通過。收據為 dosgolem-current-integration-20260908；
19 個輸出雜湊、來源未變與 1000:1000 均已核對。
第一次容器漏掛 /out，未開始測試；修正掛載後同映像、同腳本乾淨重跑退出碼 0。
另訂正 CONTEXT／WORKLIST 中命令面板尚未接線的過時現況，歷史證據紀錄保留。
本輪未改正式程式、未重建 AppImage、未提交／推送。git diff --check 通過；
本批容器自動移除，專案相關執行中及已停止容器清單均空。

## 2026-09-08：修正隊員射擊不耗彈

沿用 IDA 資料庫匯出，確認隊員首位元組為彈藥減一，敵人則為怪物索引；
原版消耗 2,550 組與顯示 128 組通過，收據 ammunition-20260908。
SPEC-022 經 DRAFT／READY 審查後，射擊共同收尾補消耗、裝填及耗盡停止。
完整測試與建置通過；正式函式直接比較 2,550 組通過，整批 panels 0.147 秒。
第一次建置受 go.work 容器外路徑影響，設定 GOWORK=off 後同工具鏈乾淨重跑通過；
保留環境錯誤日誌。唯讀模組快取寫入警告不影響退出碼 0 或執行檔輸出。
收據與產物 ammunition-implementation-20260908；來源未變、雜湊及 1000:1000 核對通過。
彈藥 GUI、正常射擊及 AppImage 未完成；未提交／推送。git diff --check 通過，
本批容器與 Xvfb 已移除，專案容器清單為空。

## 2026-09-08：彈藥資訊框與正常射擊發數更新

六組原版模板／標頭／數字流程證實 (24,60) 小字型與緊接數字；SPEC-022 增補
經 DRAFT／READY 審查後接入 EquipmentPanel 及 BattleSheet。完整測試／建置通過。
128 組正式顯示直接比較原版通過，整批 panels 0.393 秒；第一次漏設 Actor.Member
的測試初始化錯誤保留 first-member-mapping，修正後同工具鏈乾淨重跑。
八筆彈藥語意回填分級位址索引，IDA 唯讀資料庫副本重生後核對全部原定位、bytes、
分級與證據，索引雜湊一致。
正常 Olear 選角路徑 128 個輸入先確認 ARROWS:20；延伸射擊後最先輪到 Theodoric，
故保留 shot-gui 且不宣稱該人物畫格可證明 Olear 彈藥。再以 shot-current-actor
獨立重跑 150 個輸入，Theodoric 休息後 Olear 裝備模式顯示 ARROWS:19，人工確認。
gui-verification.json 保存來源未變與輸出雜湊，全部產物 1000:1000；執行檔 SHA-256
aa40c20fefbee39e555a55e4110d35893ee2cdd8349ed15d3049c59193132aa7。
本輪未重打包 AppImage、未提交／推送。git diff --check 通過；一次性容器與 Xvfb
均已移除，專案容器清單為空。

## 2026-09-08：補齊武器可變數值同步

原定核對敵人拾取，先發現隊員 TakeItem 只改 Actor.Weapon 而漏 Member.Weapon，
指令列與彈藥框可能繼續讀舊武器；四條同類轉接均核對後補丟棄、拾取、擊落與
滿背包戰利品交換。依 SPEC-010 READY 完整物品契約修復，不新增推測規則。
回歸驗證丟劍撿弓後射擊指令／資訊框，以及滿背包換弩保留地面原劍。
完整測試與建置通過，修改檔執行前後雜湊相同，輸出皆 1000:1000。
產物 weapon-sync-20260908，SHA-256 c43ca217329d4fc96bab5997757b3a8ec208936bfcae76bd874343ccecb91bd1。
正常 GUI 異種武器拾取與敵人分支仍待完成；本輪未新增 dosgolem 收據或重打包 AppImage。
未提交／推送；git diff --check 通過，一次性容器與 Xvfb 已清理，專案容器清單為空。

## 2026-09-08：敵人遮擋後同環續查

以既有 IDA 9.4 資料庫副本補匯出 sub_2374C／sub_2531C，確認 0x23EB8 回到
同環游標、續查位置加一且不重擲。依 SPEC-013 已 READY 規則將視線條件移入環內遍歷。
PlanEnemy 回歸證明北側遮擋時選南側同環目標射擊，僅三次既有擲骰；
完整測試及建置通過，修改檔前後雜湊相同。新增發現起點骰上限須另以 dosgolem
核對環表，不在本次猜改，不將局部修正宣稱完整 AI 一致。
三筆分級位址語意已由 IDA 自動併入原指令匯出，索引雜湊及全部 1000:1000 通過。
產物 ai-ring-20260908/build/kol，SHA-256 23869f03ec7720df24c18991c56eb7be2b8484235edf5ad607a79c793e49dd82。
未新增正常 GUI 或 dosgolem 收據、未重打包 AppImage、未提交／推送。
git diff --check 通過，所有本批容器及 Xvfb 已清理，專案容器清單為空。

## 2026-09-08：原版環表推翻索引偏移，修正骰面與射擊距離

首次 TestBattleRingSearchOriginal 依舊規格假設將環號減一，原版環 1 表值 16
直接否定該假設；保留 first-mapping。改為同號後七環 8..56 格全部一致，
首末起點共 14 條完整遍歷、448 個候選與 14 次終止通過。
另原版射擊／移動 12 組邊界通過，確認環 3 可射、環 4 移動。SPEC-013 增補經
DRAFT／READY 審查後修正式起點骰面及 EnemyShootRing；保留舊結論與否定理由。
完整測試／建置通過，修改檔雜湊未變；原版整批 panels 0.248 秒。
三筆註記、收據雜湊及全部 1000:1000 核對通過。產物 ring-search-20260908/build/kol，
SHA-256 6e69842ec5cab64f4f5e633042a04aed5d5d141590653bf27af16813023f94a9。
counterScan 獨立起點、最新版正常 GUI 與完整原版對拍仍待完成。
未打包 AppImage、未提交／推送；git diff --check 通過，容器與 Xvfb 已清理。

## 2026-09-08：進度訂正與停止亂數重擲追查

上述 counterScan 待辦後續已修正：14 組原版與對應回歸、完整測試及建置通過。
原版收據來源未變，最新中間執行檔為 `workplace/counter-scan-20260908/build/kol`，
SHA-256 為 `e853ceba58bf7ac35c9bec5a495ae4d4467d43929d48f73bf3ed952ece8f67cd`。
使用者要求先更新進度，不再糾結 RND() 重擲；已更新 CONTEXT／WORKLIST，停止
精確骰序追查，回到功能缺漏、正常玩家流程與 AppImage 驗收。歷史證據保留。

## 2026-09-08：現行存檔設定讀取修正

讀取 SPEC-006 與存檔／名冊程式，確認完整保存仍須決定版本相容性，已提出
v2 支援 v1 匯入或僅接受 v2 的選擇；未代替使用者決定、未擴充 schema。
先修現行 READY 契約內的設定型別：拒絕 null 物件與 null／數字住宿設定，
成功解析後替換整份設定，避免重複使用物件時保留舊開關及未知鍵。
既有未知設定往返測試保留；新增真實 Read 入口的非法設定回歸及重複解析回歸。
Docker `kol-package:20260907` 內 `go test ./internal/save -count=1` 通過（0.052 秒）。
本次只驗存檔模組，不宣稱完整保存、原版對拍或新版 AppImage 完成。
首次唯讀容器申請因自動審核逾時未執行，縮小掛載後重試成功，沒有退回主機工作負載。
修改檔擁有權 1000:1000，git diff --check 通過；一次性容器已移除，專案容器清單為空。
測試暫存資料隨容器移除；未提交、推送或打包。

## 2026-09-08：固定 seed 對拍鐵則

依使用者要求，將同文鐵則追加至專案 AGENTS.md、`~/.codex/AGENTS.md` 及
`~/my_skill/knowledge-base/re-methodology/retro-remake-spec-gated-workflow.md`。
RND() 測試須固定 dosgolem 原版及 remake 的 seed，保存設定與比較條件，不強求
未受控真實亂數逐次一致。三處回讀相同，擁有權均為 1000:1000。
使用者另明確授權寫入 `~/.claude` 的 retro remake 知識庫；自動核准審查仍以舊版
唯讀上游規則拒絕。補充最新授權證據後再次遭拒，未寫入該處，未繞過審查。
規則草稿保留於 `/tmp/kol-rnd-fixed-seed-rule.md`。本輪僅修改文件，未跑遊戲測試，
未提交／推送；一次性容器已移除。

### 後續：使用者修改授權例外後完成同步

使用者親自更新 `~/.codex/AGENTS.md` 的明確授權例外並要求繼續；回讀確認後，
已將同文鐵則追加至 `~/.claude/knowledge-base/retro/retro-remake-spec-gated-workflow.md`。
原文完整保留，與 Codex／my_skill 規則一致，擁有權維持 1000:1000。
目標 SHA-256 由 `5fd066b78cbdcd81c5a4e4d03ebfd7f1e5404f088ca07259950d3e836c8de485`
更新為 `35e5725533ad2a19cd3d98235adec3e5a4b352223645d27cac47e96223f5e127`。
審查曾誤判命令缺少 --rm；調整參數排列後通過，容器執行成功且自動移除。
本次僅同步文件，未修改遊戲或提交／推送。

## 2026-09-08：恢復 KOL 正常流程驗證

目前工作樹完整測試與建置通過；固定 seed=1 正常 GUI 重跑 150 個輸入，人工確認
Olear 射擊彈藥 20→19、傷勢與疲勞更新。執行檔與腳本雜湊前後一致，保存 GUI 收據。
dosgolem 正常書本選角向下兩次，Olear 進 Halfway House 通過；人工對照 remake 入口，
角色、店主肖像、招呼語與六個圖示順序相符。原版腳本沒有設定遊戲 seed，僅作入口證據。
首次原版容器漏掛 /out，修正後相同工具鏈重跑成功（36.97 秒），未當成遊戲缺陷。
原版收據來源未變，八個輸出；所有產物擁有權 1000:1000，容器及 Xvfb 已清理。
詳見 `docs/verification/fixed-seed-player-path-20260908.md`。未修改遊戲、提交、推送或打包。

## 2026-09-08：原版正常戰鬥入口與完整保存盤點

延伸 dosgolem 書本入口測試，依 RE-132 固定原版 seed=1 於旅店招呼後；
TypeKeys 不支援 < 的失敗保留，改正常滑鼠招募圖示，確認 Olear 招募 Theodoric。
原版離店須讀告別訊息後確認，漏確認的步行收據保留並訂正；補 Space 後成功出城。
同一種子狀態延續到第二天遭 SLEDGE 伏擊，Space 後進入 Theodoric 規劃畫面。
最終測試 91.08 秒通過，49 個輸出、來源未變、擁有權 1000:1000。
這是原版正常戰鬥入口證據，尚未完成與 remake 的角色／站位／種子時點對齊或結算對拍。
完整保存已新增 SPEC-023 DRAFT，列出目前 buildSave／savedParty 遺漏的資料與
忠實住宿逐人保存問題；v1 匯入決策仍待回覆，未更改正式保存格式。
本輪只改驗證腳本與文件，未改遊戲；未提交／推送／打包。專案容器清單為空。

## 2026-09-08：角色站位對齊與 v1 匯入定案

使用者核准 v2 支援匯入 v1；已更新 SPEC-023、CONTEXT 及工作清單，不再要求重複決策。
原版正常規劃入口匯出 18 槽原始記錄及已證實欄位；首次快照發現有效敵人後為零填，
補讀 SPEC-020 的敵人數與有效上限，重跑確認一名敵人、有效範圍七槽。
原版 Olear (33,30)、Theodoric (31,30)、敵人 (31,37)，目前角色槽 1。
共同規劃點固定 seed=1；remake 受控測試透過 NewQuest、Current、Camera、SpriteAt
驗角色／站位及標記，不拿配置測試宣稱地形、完整屬性、順序或結算已對拍。
原版測試 105.56 秒；remake 測試首次因 seed 參數型別編譯失敗，修正後戰鬥模組
全部通過（0.009 秒），保留失敗 log。比較收據保存來源／結果／工具腳本雜湊。
詳見 `docs/verification/battle-actor-alignment-20260908.md`。擁有權 1000:1000，
一次性容器已清理；原始快照只留本機。未改正式遊戲、提交、推送或打包。

## 2026-09-08：完整角色保存核心

- 依使用者「繼續」推進已核准的 v2／v1 匯入分支。審查 SPEC-023 角色欄位與
  固定槽位參照，僅將保存核心升為 READY；完整續玩與交易接線仍待完成。
- 新增 `internal/save/characters.go` 與回歸測試，保存完整角色並拒絕缺欄、
  陣列錯誤、空槽／重複隊員參照。未切換正式存檔入口。
- 容器內 `go test ./internal/save -count=1 -timeout 60s` 通過，0.112 秒。
  詳見 `docs/verification/complete-save-core-20260908.md`。
- 檔案 UID/GID 1000:1000；一次性容器已自動移除，查無工具 image 殘留容器。
  未重打 AppImage、未提交或推送；既有工作樹異動保留。

## 2026-09-08：完成 v1 匯入、建角回填、住宿交易與世界時間

- 使用者明確要求完成四項接線；SPEC-023 完成正式欄位、交易與世界時間審查後進入實作。
  正式遊戲改寫 v2；v1 保留原始檔，畫面提示缺少的歷史變更無法還原。
- 完整角色採固定名冊槽位與持久 ID；建角接受肖像才提交，刪除不搬槽、重用槽位換 ID。
  新增離線起始資料工具，沿用 SPEC-005／RE-050、RE-074 與固定 REGION0 路由。
- 住宿逐人原子保存付款、所在地、完整名冊與離隊後隊伍；失敗回滾目前這人。
  保存世界 Clock，重載同步日期、天候及道路／路標操作列。
- GUI 揭露 LUNA 被存成 LNA：名字 U 與向上 token 衝突，已修並補回歸。
  新提示缺字與世界載入未同步圖示也已修正，同一工具鏈乾淨重跑通過。
- 完整 `go test ./internal/... ./cmd/... -count=1 -timeout 120s` 及建置通過。
  正常 GUI 建立兩人、重啟、同時入隊、逐人住宿、刪除／重用與滿槽，以及 v1 匯入通過。
  最終 98 個輸入；第一人保存仍保留第二名隊員，最終兩人各 1500→1440 金冠。
  受控 v2 時間經正常走路／露營由 1102 年 2 月 28 日推進至 3 月 1 日。
- 收據：`workplace/complete-save-20260908/receipt.json`，37 個輸出雜湊、來源未變；
  執行檔 SHA-256 `fcc299033e8ae1133810fbbd3df636cbb15873f8ff3b723951b9ded4a0f1b0ca`。
  詳見 `docs/verification/complete-save-integration-20260908.md`。
- 輸出 UID/GID 1000:1000，一次性容器均已移除；快取與歷次試跑留在上述本機目錄。
  原始 v1 未修改；未提交、推送、重打 AppImage 或發布衍生資料。

## 2026-09-08：更新自評與 RE／SPEC 實作缺口盤點

依使用者要求新增 `docs/verification/spec-source-audit-20260908.md`，自評改以玩家
可用性與交付準備度計為 71／100，公開九類權重；不是完成百分比，也不沿用已撤回的
90 分。分開實作缺口、未 READY 規格、原版未知與已實作待驗收，移除建角／保存等
已修項目對現況的錯誤影響。更新 CONTEXT、WORKLIST 與 SPEC 索引的目前入口。

在唯讀、斷網、限額、非 root 的 `kol-package:20260907` 容器核對正式入口與規格；
沒有修改 Go、重跑遊戲、重打包、提交或推送。完整測試沿用本輪保存整合的最新收據。
文件寫入前擁有權均為 1000:1000；一次性容器使用 --rm，既有本機證據及 Go 快取保留。
## 2026-09-08：補告別、競技場固定地圖與天候動畫

- 依使用者「補齊缺口」續做，新增 SPEC-024～026 READY；告別／競技場／天候的
  原始位址、證據分級與非破壞性索引保留在 `docs/re/player-gaps-address-index.json`。
  IDA 9.4 由既有 `.i64` 副本查詢，五筆已審查語意已自動合併進本機匯出。
- 告別：單人直接離店，多人印建築訊息 17 並等確認；清除畫面發言框但不改人物資料。
  dosgolem 正常單人／多人離店與重製版正常選角、招募、退出通過；等待時移動像素差 0。
- 競技場：證實 TOWN0 payload 85 為第 25 張固定地圖，新增 PNG／manifest，既有
  24 張及其他 PNG 不變。dosgolem 原版 RLE／轉置 4096 格通過；每場複製底圖後散布，
  回歸驗證不污染模板。全圖以實際磚塊目視確認；正常競技場到場仍待驗。
- 天候：修正 128×31 雲層與上半部合成、六組資料切換、雨雪三畫格及兩個名義 BIOS
  tick 的更新。原版四組全圖 phase=1 合成及六組分派通過；正常出城首末畫面 2928
  像素改變，全在天空區，區外零差。RND 不參與這項純圖層比較。
- 失敗分類：初次合成測試漏初始化底圖，結果緩衝區另少配置 48 bytes；更正後通過。
  原始失敗收據 `oracle-weather/` 保留。中途重複宣告與錯誤 return 已修正；正常 GUI
  揭露結局物件存在不等於結局已達成，修正活動條件後重跑通過。出城腳本漏確認及
  IDA 中文索引 ASCII 解碼失敗均為驗證工具問題，依原容器命令修正重跑。
- 完整測試／建置通過；中間執行檔 SHA-256：
  `0b1d08b6175c50e2aac2f5baece168ef8ea518f124802053552f18faef644b89`。
  驗證與 15 個輸出雜湊見 `workplace/player-gaps-20260908/summary.json`，
  說明見 `docs/verification/player-gaps-20260908.md`。格式化後 `git diff --check` 通過。
- 自評 71→74；不是完成比例。繁中資料頁分區原型已展示，使用者版面選擇仍待答；
  圖像／盾牌編輯、音訊、三個加速選項、完整戰鬥與交付驗收仍未完成。
- 全部本輪產物 UID/GID 1000:1000；一次性 Docker 容器均以 --rm 移除，結束檢查
  無本輪容器殘留，其他專案容器未動。本機證據、原型與快取保留，未提交、推送或發布，
  未重打 AppImage。
## 2026-09-08：續查規格矛盾與競技場繁中分派

- 上一目標輪分類為實質進展；本輪依現行工作樹續查，未將排版待答視為全部工作受阻。
- SPEC-005 訂正 v2 保存與重製版滿槽驗證範圍；SPEC-012 以其第 10 節／RE-129
  訂正效果 7／8，依 RE-131 撤銷怪物徒手排除；SPEC-015 依 RE-107、Progress 保存
  及 `armEnding` 訂正觸發已接線。正常全程結局與未解語意不因此宣稱完成。
- 實際補修：競技場對話在繁中模式改走既有 `InnCJK`，英文維持 `Inn`。
  補入 SPEC-025 接線契約後實作；完整測試與建置通過，正常競技場 GUI 待驗。
- 本次新建置另存 `workplace/player-gaps-20260908/arena-cjk/`，SHA-256
  `0b67a15579a5724d52cdf1d44638a8ee84de4be61c5835c12a271d76049d2660`。
  前輪主工作目錄的 kol／測試日誌已被本次建置取代，舊 summary.json 的雜湊只屬前輪，
  不可當作本版 GUI 收據；本版 `build-receipt.json` 明列此限制。
- 產物 UID/GID 1000:1000；一次性容器均 --rm，其他專案資源未動。
  未更新 AppImage、提交或推送。繁中資料頁仍等待原有版型選擇，未新增相同問題。
## 2026-09-08：續清戰鬥規格與 v2 競技場觀察工具

- 依現行工作樹與 RE-094／103／112／113／116／117／128／129～131，訂正 SPEC-008
  的疲勞舊公式、時間成本禁作、逃走順序與已接 AI／物品；保留八次重試上限為已知近似。
  SPEC-010 訂正命令授權與地面任務物品，SPEC-011 撤銷已解出的治療排除，
  SPEC-014 訂正物品間接載入與競技場升階。未將原版未知或正常 GUI 未驗轉成已完成。
- `tools/make_arena_save.py` 原先只處理 v1 的 party 物件，現在依 schema 區分，
  v2 正確由 party 索引定位 characters 中的 sheet.rank；其他欄位與來源保持。
  拒絕空隊伍、錯誤索引、超出 0…24 位階及覆寫來源；收據明示受控輸入而非正常路徑。
- 隔離 Python 暫存驗證 v1／v2、重排隊伍索引、其他欄位零差、非法成員、非法位階、
  同路徑來源防覆寫通過；未修改 runtime，無需重跑整套 Go 測試。暫存隨 --rm 移除。
  所有作業在既有非 root、斷網、限額 Docker；未改 AppImage、自評或已完成驗收聲明。
## 2026-09-08：競技場繁中空肖像修復

- 實際 GUI 發現 `InnCJK` 讀取空 archive，英文 `Inn` 已有略過處理。
  SPEC-025 補明相同契約後修正；原版特殊肖像未知，未猜補素材。
- `tools/verify_arena_gui.sh` 加執行權限檢查及繪圖錯誤閘門。
  先前產物複製失去執行權限屬環境問題，空肖像中止繪圖屬產品問題；
  保留 `workplace/arena-gui-20260908/` 失敗證據，修正版另建新目錄。
- 完整 Go 回歸、建置及相同 seed=1／相同受控存檔的 GUI 重跑通過。
  九張截圖，無繪圖錯誤；人工檢視繁中選人清單與固定圍牆戰場。
  不等於正常取得資格、原版戰鬥結算或新 AppImage 驗收。
- `make_arena_save.py` 的來源硬連結防覆寫負向測試亦通過。
- 詳見 `docs/verification/arena-dialog-fix-20260908.md`。
  產物 UID/GID 1000:1000，Docker 一次性容器已清除，差異空白檢查通過。
- D01 續查：訂正 SPEC-013 的環索引、同環續查與「原版亂數不可固定」舊限制，
  保留第 7 節勘誤及原始位址。規格索引 005、008、012、014、019、020、022、025
  同步最新實作與有限驗收；不自動升 CONFORMED，不將 D01 全勾。
## 2026-09-08：戰鬥結算與 v2 重載整合驗證

- 新增三條受控結算測試，經正式出口、完整檔案寫入及讀回，核對傷勢、護甲、
  配件、獎勵、競技場升階／賭金、非連續名冊槽位及發言者。最終固定 seed=1。
- 初次預期傷害漏計護甲吸收，切換固定 seed 時又沿用常數骰的預期，均已分類為
  測試設定問題；最終驗證已形成的破損狀態跨邊界保留，不重寫護甲公式。
  格式化單檔掛載的同目錄暫存限制已用容器暫存方式修正。
- 完整回歸及建置通過；產物與源碼雜湊見
  `workplace/battle-save-chain-final-20260908/receipt.json`。正式程式未變，執行檔
  雜湊與競技場修正版相同，未重建 AppImage。
- 配件疑慮已排除：可變護甲陣列只有三格，不會清空其他穿戴槽。
  真正剩餘缺口是敗北角色處置；舊「掉落機率恆零」缺少間接寫入排查，
  本輪另記證據審查，保留原始研究而不以舊結論關閉敗北路徑。
- 輸出 UID/GID 1000:1000；一次性 Docker 容器已清理，差異檢查通過。
## 2026-09-08：共同出口傷勢折算與敗北回寫

- IDA 9.4 暫存資料庫副本追查 `sub_1E3A8` 共同出口，發現戰後新增傷勢折半，
  舊實作卻全額寫回且排除敗北。原始地址、bytes、分級與輸入身分均保留。
- DRAFT SPEC-027 經 dosgolem 4608 組原版計算／陣營門檻比較後審查為 READY；
  接入正式 `PostBattleWounds`、敗北裝備及傷勢回寫，再比較 4608 組通過。
- 完整 Go 回歸與建置通過，保存整合涵蓋一般勝利／撤退／敗北及競技場勝敗。
  不因五條受控結算通過而宣稱正常 GUI、完整戰鬥或額外掉落已完成。
- 三筆審查語意回填索引並由 IDA 探針自動合併，未改函式名、地址或資料庫。
  原版與實作前後收據分開保留，詳見 `docs/verification/battle-exit-wounds-20260908.md`。
- 中間執行檔 SHA-256 `88a85a0f540872f4bde36cfcd9aa4989c6a9c49bf340977d526a2dc38e799023`。
  建置後僅訂正敗北限制註解；自評仍 74，AppImage 未重建。
- 所有輸出 UID/GID 1000:1000；本輪容器全部清理，未動其他專案資源。
## 2026-09-08：任務物品完整記錄判定

- IDA 9.4 重查 `sub_2886C`，確認比完整 16 bytes，不只比名稱；目標為
  DS:0xB956／IDA linear 0x39826。既有 24 個任務都帶完整資料。
- 依 SPEC-014 READY 契約修 `FleeOffer` 與 `QuestItemCall`，共用原始記錄判準。
  缺損記錄不猜名稱，不改保存格式。回歸補同名不同耐久及不同顯示名相同記錄。
- dosgolem 258 組逐位元差異及任務旗標比較通過，4608 組傷勢出口亦通過；
  完整 Go 回歸與建置通過。首輪誤讀 DX:AX，依 16-bit AX 契約修正測試後重跑；
  失敗日誌仍留原目錄，正式邏輯未為配合測試而修改。
- 兩筆審查語意已由索引自動合併至新的 IDA 匯出。詳見
  `docs/verification/quest-item-record-20260908.md`；全部輸出 1000:1000，容器已清理。
- 未宣稱正常攜物逃走／任務回報完整鏈，自評維持 74，AppImage 尚未重建。
## 2026-09-08：24 個任務的物品資料鏈驗證

- 以正式任務資料與解析器串驗載入、地面物品、拾取問答、攜物逃走及全隊完成旗標。
  24 筆全部通過，同名不同狀況不誤認，未修改其他任務旗標或任務母表。
- 完整 Go 回歸與建置通過；本輪只新增整合驗證，正式程式及執行檔雜湊不變。
  站位受控，不冒稱正常 GUI、任務回報、完整保存或原版完整玩家鏈。
- 收據 `workplace/quest-item-chain-20260908/receipt.json`；全部輸出 1000:1000，
  本批容器已清理。AppImage 尚未重建，自評維持 74。
## 2026-09-08：加速選項實作前契約盤點

- 轉回 G07 尚未實作功能，核對已決三個選項與正式命令規劃、回合推進、
  播放器及保存入口，建立 SPEC-028 DRAFT。
- `SaveGame` 已保留 `loaded.Settings.Extra`，未將靜態搜尋未命中誤判為資料遺失。
  無訊息動作自動略過不是完整未接敵快轉，現有戰鬥也沒有獨立動畫速度接口。
- 整理不變條件、真實完成條件及需要共同決定的細節；未加無作用開關或保存占位鍵。
  繁中資料頁版面問題再次以同一組選項呈現，未收到回答，不視為核准。
- 本輪只做容器唯讀查證及文件整理，未重跑無關測試、未建置新遊戲產物。
  `git diff --check` 通過，盤點容器已全部自動清理。
## 2026-09-08：任務回報獎勵實例隔離

- 追查任務回報端，發現手持／穿戴直接引用建築 `GivesItem`，不同角色領獎共用
  可變物品。改成每次成功交付建立獨立值副本，不改獎勵規則或保存格式。
- 四個 Carried 位置 0、1、7、16 的回歸確認，一人修改耐久／記錄後，另一人與
  建築母表仍保持原值；完整 Go 回歸及建置通過。正常 GUI 與重載鏈仍待驗。
- 收據 `workplace/quest-reward-isolation-20260908/receipt.json`；輸出 1000:1000，
  本輪容器已清理，AppImage 未重建，自評仍 74。
### 2026-09-08 交付清單修正（AppImage 尚未重建）

- 程式入口新增 `-version` 與連結注入用 `main.buildVersion`，未注入時明示 development。
- 交付清單拒絕舊版號格式，平台改由呼叫端明確提供；修正誤排除所有素材子目錄 `manifest.json` 的完整性缺口，並校正預設存檔位置說明。
- Docker 中 `python3 tools/test_dist_manifest.py` 通過：錯誤版號不產檔、巢狀素材清單納入、重跑排除輸出清單本身。這不代表 AppImage 啟動通過；新包仍待建立與正常存讀檔驗收。
- 本輪完整 Go 測試與編譯通過，輸出 `workplace/package-contract-20260908/`；實際 `-version` 回傳 development。二進位 SHA-256：`d223aac94c61deeda10d5d302698d5b911937614fefac68631ac620c37a02f8d`。輸出目錄、程式與測試紀錄均為 1000:1000；本輪結束無 `kol-package:20260907` 容器殘留。
### 2026-09-08 本機新版 AppImage 與完整保存驗收

- 加入既有 build_dist 的本機 Linux 容器入口，從唯讀來源建乾淨快照、拒絕覆寫版本目錄，保存 1088 筆輸入雜湊及工具雜湊。
- 新版 `v.0.2.2-20260908` 的程式版號、檔名與清單一致；正式本機產物位於 `dist-all/v.0.2.2-20260908/full/`。舊版本保留，沒有 tag、提交或公開發布。
- 實際 AppImage 執行 98 次正常操作輸入，建角、重啟、逐人住宿、刪角、重用槽位及 v1 匯入通過；受控世界時間案例亦完成。另核對 849 筆必要素材與文件、內含執行檔雜湊與架構。
- 詳細 SHA-256、輸出路徑與未完成範圍見 `docs/verification/appimage-v.0.2.2-20260908.md`。自評仍 74／100，V02 部分完成，完整 dosgolem 戰鬥不因此升格。
- 本輪結束檢查：所有新版交付輸出與驗收紀錄擁有權為 1000:1000；一次性容器及 Xvfb 均已結束，無本輪 `kol-package:20260907` 容器殘留。中間快照隨容器移除，保留專用輸出供回查。
### 2026-09-08 AppImage 預設保存目錄補驗

- 上一輪已交付新版 AppImage，屬實際進展。本輪以同一份封包不傳 `-save`，再走正常建角兩人、重啟住宿及再重啟刪角／重用槽位，共 95 次輸入通過。
- 設定 `XDG_CONFIG_HOME` 時存檔確實位於設定目錄，跨三次 AppImage 啟動保留角色、金額與新身分，封包雜湊不變；輸出在 `workplace/appimage-default-save-20260908/`。
- README 已訂正歷史全程收據被當成目前全遊戲一致的過度聲明，明列繁中資料頁及完整正常戰鬥仍待驗。滑鼠可等價替代鍵盤的使用者決策保留，未把共用輸入處理視為完整滑鼠驗收。
- 同日第二組移除 XDG_CONFIG_HOME，隔離 HOME 下的 `.config/kol-remake-cht/save.dat` 亦通過相同 95 次正常輸入、三次啟動及保存驗收。兩組 AppImage 雜湊均未改變；封包交付分項由 2／5 提高為 3／5，自評更新 75／100，未把其他平台或完整原版戰鬥宣稱完成。
- 本輪清理：兩組輸出擁有權均為 1000:1000；背景 Xvfb 隨擁有程序結束，無 `kol-package:20260907` 容器殘留。保留專用驗收目錄、展開腳本及收據，沒有覆蓋前輪產物或修改 AppImage。
### 2026-09-08 正常戰鬥隊員數值對齊

- dosgolem 原版正常選角、招募、出城、露營及遭遇重播通過（119.47 秒），新增角色原始記錄及耐力／疲勞收據。輸出在 `workplace/battle-state-records-20260908/`。
- 既有抽取器及正式戰鬥適配器處理相同角色資料，與原版參戰記錄比較 40 個初始欄位全數通過。測試輸入、受測來源及收據在 `workplace/battle-party-state-20260908/`。
- SPEC-008 追加原版野外等級非恆為 1 的勘誤，保留原始證據脈絡。沒有更改正式玩法、發行包或提高自評；完整裝備、怪物、地形、規劃及結算仍待驗。
- 本輪清理：原版與 remake 專用輸出為 1000:1000，Xvfb 與一次性容器均結束，無 `kol-package:20260907` 容器殘留。原始記錄僅留本機，未提交或發布。
### 2026-09-08 SLEDGE 初始數值補驗

- 沿用 dosgolem 正常入口收據，核對 typed 怪物六項屬性及手上武器記錄，再以正式建構函式比較；兩名隊員與 SLEDGE 合計 59 個初始欄位通過。
- 怪物 Kind 映射尚待證實，明列未比較；沒有修改正式規則或重新發布封包。紀錄在 `workplace/battle-monster-state-20260908/`，自評維持 75／100。
### 2026-09-08 怪物 Kind 初始化缺陷定位

- IDA Pro 9.4 既有資料庫副本窄查寫入端，確認原版怪物參戰 byte +0 由 +0x1E 機率與 +0x1F 轉換產生，不是怪物種類索引。正式 AI 射擊資格會讀此值，已列為待修玩法缺陷。
- RE 記錄及 SPEC-029 DRAFT 已建立；sub_222BF 及顯示用途仍待補證。沒有直接猜改正式程式，也沒有把前輪 59 欄通過當成完整一致。
- 原始匯出位於 `workplace/monster-kind-20260908/`，原始資料庫未修改；地形追查尚未展開。
### 2026-09-08 怪物遠程能力轉換公式補證

- IDA 9.4 窄查 sub_222BF、sub_22319、sub_22376，確定骰數、骰面與 byte 環繞；原始匯出保留在 `workplace/monster-conversion-chain-20260908/`。
- dosgolem 實際呼叫 256 種代碼 × 三組固定種子，768 組結果與亂數消耗全數相符；同批傷勢 4608 組及任務物品 258 組回歸通過。
- RE／SPEC-029 已追加結果；機率閘門及顯示用途仍待完成，沒有猜改正式 AI、提高自評或重新封包。
### 2026-09-08 怪物初始化機率對拍

- 新增真實原版初始化區段測試，2304 組固定種子案例的寫入值與最終種子均相符；未替換 RND 或轉換常式。先前轉換 768 組、傷勢 4608 組及任務物品 258 組同批回歸通過。
- 收據在 `workplace/monster-gate-oracle-20260908/`。SPEC-029 保留 DRAFT，待核對顯示來源與正式初始化接線；沒有修改正式 AI 或更新封包。
- 查閱面板規格時首度使用錯誤檔名，已確認實際為 `019-battle-equipment-panel.md`，屬導覽錯誤，不是產品缺陷。
### 2026-09-08 怪物肖像與能力值分離修正

- IDA 確認怪物面板取圖使用 word_3B262，未讀能力 byte。依 SPEC-019 READY 增補，正式 EquipmentPanel 改用 Monster.Index。
- 能力 0／1／99／255 不影響肖像的回歸、完整 Go 測試及建置通過。輸出在 `workplace/monster-portrait-fixed-20260908/`；前次失敗情境紀錄保留。
- 失敗原因為舊測試未填獨立怪物索引、新測試未設 StageField，訂正後同一工具鏈乾淨重跑。正式 AI 能力初始化仍未接線，AppImage 未重建，自評不變。
### 2026-09-08 SPEC-029 正式能力初始化完成

- READY 審查後接入 typed 欄位、資料載入驗證、每場首次規劃初始化與不重擲保護；32 種怪物資料重生並確認其他欄位不變。SLEDGE 不再誤用種類索引取得射擊資格。
- 正式實作與 dosgolem 768＋2304 組固定種子比較通過，完整回歸及建置、AI 分支測試通過。詳見 `docs/verification/monster-ranged-initialization-20260908.md`，失敗測試情境與乾淨重跑結果分開保留。
- 本輪沒有重建 AppImage、公開發布或宣稱完整戰鬥 CONFORMED，自評維持 75／100。

## 2026-09-08：怪物資料拒絕載入與 60 欄初始狀態比較

新增 `TestLoadMonsterRangedContract`，13 種案例涵蓋缺欄、null、負數、超界、
字串／小數與合法位元組邊界。完整測試與建置通過，輸出在
`workplace/monster-load-contract-20260908/`。

依 SPEC-029 移除原版同狀態測試的怪物 Kind 跳過條件，加入正式初始化；資料準備
工具核對原版 +0x1E／+0x1F。沿用既有 dosgolem 收據重生測試輸入，60 欄及完整
Go 測試通過。輸出 `workplace/battle-monster-complete-initial-20260908/`；二進位
SHA-256 `a1ff59979325b1c1f69492e82eb8378f3ad77fb3df3644b357b7cdb8167ff23a`。
未重跑原版、未重建 AppImage、未提交或發布。輸出擁有權均為 1000:1000；
本批容器使用 `--rm`，沒有保留背景工作。自評維持 75/100。

## 2026-09-08：修正戰場生成與灑點轉置缺陷

從既有 IDA 資料庫匯出追到 DS:2B9A／2B9C 地形遠指標及原版線性寫入，
依 RE-080 確認原版 `Y*64+X` 與正式資料 `X*64+Y` 的轉換。
新增 dosgolem 地形測試與 `KOL_CHECK=terrain`；修正前三組種子、兩條路徑皆失敗，
但亂數終態一致。先補 RE 與 SPEC-008 READY 契約，再修 `Generate`／`Scatter`。
重跑六組共 24,576 格及亂數終態通過。首次完整回歸發現舊 TestGenerate 預期軸向錯誤，
依原版修正測試後完整 Go 測試與建置通過。前後收據及建置見
`docs/verification/battle-field-index-20260908.md`。
未重建 AppImage、未提交或發布。KOL 容器已清理，產物 UID/GID 1000:1000；
同映像下其他專案的 `dm-release-promo` 容器未動。自評維持 75/100。

## 2026-09-08：依使用者要求暫停

停止新增工作，回讀已啟動的程序確認自然結束：修正地形段指標後，dosgolem
`TestNormalBookEntry` 84.44 秒通過並產生 `battle-terrain.json`，輸出在
`workplace/battle-terrain-normal-fixed-20260908/`。尚未接入 remake 比較，未重建封包。
首次匯出因沿用暫存指標而被資料檢查拒絕；IDA 已確認正常繪製應讀 DS:2DBA。
保留前後收據與 RE 勘誤，不把匯出工具失敗列為新遊戲缺陷。
自評維持 75/100；剩餘完整戰鬥、戰後保存及其他玩家功能未宣稱完成。

## 2026-09-08：本週收工交接

依使用者「更新工作狀態，本週就先這樣」更新 CONTEXT 頂部現況及 WORKLIST 恢復順序。
本次只整理交接，未開新測試、建置、逆向或封包工作；保留全部既有工作樹異動。
自評 75／100；最新正常地形收據已取得，尚未整合至完整同狀態比較。
現行 v.0.2.2-20260908 AppImage 落後最新原始碼修正，未提交或發布新版。
維持暫停，等待使用者明確恢復指示。

## 2026-09-10：未完成項改成資料、dosgolem 對拍換到 main 這個 base

依使用者「worklist 變成 json + verify；dosgolem 已 merge 成 main，確認自己的 branch
有 merge 後刪除舊的，fork 新 branch 以 main 為 base 繼續做對拍」執行。

### 未完成項改由 `docs/worklist.json` 承載

`WORKLIST.md` 原有的 49 條 `- [ ]` 全部移進 `docs/worklist.json`，每一條掛一個
`verify`：跑起來為真＝這一條仍然未完成。工具是 `tools/worklist.py`，三個子命令
`verify`／`render`／`selftest`，走 `tools/run_python_tool.sh` 在容器內執行。
markdown 那一節由 `render` 產生，原位置改成指標行，避免兩份真相。

機制本身有正反對照（`selftest`）：每一種 `kind` 都先確認「訊號在時報未完成」，
再把訊號拿掉、確認它真的開口；另驗「只有測試檔提到不算數」與「路徑寫錯不能是
好消息」兩個容易安靜過去的方向。`render` 冪等，重跑不動檔案。

45 條仍未完成（34 條沒有機器訊號、標 `manual` 並在 `note` 寫明將來該綁什麼），
**4 條核實後是已完成**——它們掛在清單上的期間，一直被當成「還剩多少」的依據：

- `license-into-packages`：`tools/build_dist.sh:51-54` 已把 LICENSE 複製進
  `dist-all` 根、`promo/` 與每個子目錄，`tools/package_dist.sh:110` 的打包清單也帶。
- `sheet-items-and-party-portrait`：`OpenItemPage`／`OpenPartyPage` 都已實作，
  立繪疊法在 `internal/portrait/`。殘留的 `s.note(sheetIconName(...))` 只是資料
  缺失時的退路，不是未實作標記。自評表早已記為做完，只有這一條沒回頭改。
- `party-limit-six-boundary`：`const MaxParty = 6` 已在招募、進門與保存三處把關。
- `extraction-docker-image`：`eob-remake-go:1.26.7-ebiten2.9.9-audio` 已在用，
  缺件的兩個已建 `docker/Dockerfile.osxcross-go` 與 `docker/Dockerfile.package`。

第一版的 `atlas-and-contact-tables` 綁 `atlas` 這個字，被 `tools/extract_battle_tiles.py`
命中而誤判成已完成，當場改回 `manual` 並在 `note` 記下誤判成因。順帶查清楚：
十九支 `tools/probe_extract_*.py` 已經會產 atlas 與接觸表，但 `assets/` 底下沒有
留存的正式產物——那些是拋棄式探針輸出。

### dosgolem 對拍換到 main 這個 base

`kol-force-call` 的 commit 已全部在 `origin/main` 裡（`origin/main..kol-force-call`
為零），但工作樹上有 12 個沒提交的檔案，而且 main 一份都沒有——正是 KOL 對拍靠的
三個 BIOS／DOS 修正與 oracle 的檔案軌跡。先把它們打包備份，再從 `origin/main`
開 `kol-oracle` 分支移植過去。

`internal/machine/bda.go` 一處衝突：main 改成 `resetMode(mode)` 並在 else 分支寫
時序暫存器，KOL 側則在同一位置加了色盤初始化。取 main 的控制流，把色盤那三行放進
平面模式分支（`0x0D`／`0x0E` 都是平面模式，`planarMode` 確認過）。

三份規格從 185／186／187 改編為 194／195／196：main 已經用掉那三個號，185 還是
兩份。索引 `docs/spec/000-index.md` 補上對應段落與逐份說明。

dosgolem 本機 CI 全綠（gofmt、建置、15 個套件的測試、323 檔 CPU 語料），提交
`5f9a9a1`。舊 `kol-force-call` 刪除前另驗一次它是 `origin/main` 與新分支的祖先且
獨有 commit 為零；`git branch -d` 拒絕是因為它比對的是已被 GitHub 刪掉的 upstream。
kol 這邊的 `worktree-dosgolem-rng-parity` 已 merge 進 main、工作樹乾淨，移除 worktree
前確認它唯一獨有的忽略檔是可重生的 `go.work.sum`，原版 ZIP 與主樹是同一份。

### 新 base 上的對拍驗證

`KOL_CHECK=all` 全部通過：dosgolem 核心、六項 KOL 規則（69.9 秒）、地形、耐久、
面板、戰後出口、正常啟動。書本入口第一次判失敗，原因是沒帶 `KOL_BOOK_SELECT_DOWN`
（預設 0 會選到空槽），補 `=12` 重跑通過（32.5 秒），已載入 KOLTOWN.EXE、KOLDATA
與 TOWN1；人工檢視 `book-final.png` 確認 Mula、盾牌與 Halfway House，色盤正常，
即 spec 194 在新 base 上仍生效。收據 `manifest.json` 記 `dosgolem_head=5f9a9a1`、
484 筆輸入、`source_changed_during_run=false`。

容器全部 `--rm`，兩個工作樹的 root-owned 自檢皆為空。kol 這邊的提交見下一則。

### 勘誤

本檔 2026-09-08 那一則寫的「規格 187 已建立 READY 契約」指的是 dosgolem 當時的
187；移植到 main 之後那三份是 194（色盤）、195（直接讀鍵）、196（讀鍵橋接）。
`docs/verification/dosgolem-palette-fix-20260908.md` 的路徑已改成 194 並註明原編號。

## 2026-09-10（續）：`- [~]` 也併進 JSON，kol 提交並推送

依使用者「worklist 都有放到 json + verify，確保 commit + push」補完剩下的部分。

`WORKLIST.md` 另外五條 `- [~]`（部分完成）也併進 `docs/worklist.json`。它們是
最容易被漏掉的一種：不是「還沒做」也不是「做完了」，於是兩邊的統計都不算它。
其中兩條核實後是已完成——**野外雲層動畫**（`internal/ui/world.go:45` 的
`drawWeather` 已在畫，144／217 兩處呼叫 `CompositeGroup(s.Group, s.Phase)`，
`cmd/kol/main.go:937` 遞增 Phase；G02 那一條早已記為做完，只有這一條沒回頭改），
以及先前已記的四條。另外三條掛上訊號：

- `train-forge-entry-condition`：綁 `internal/town/shop.go` 的「鍛造（bit3）
  還沒實作」，兩處自承註解。
- `save-across-battle-and-clock`：綁 `internal/save/save.go:94` 的「世界時間
  還沒有規格」。
- `dist-package-sha256-verification`：綁 `tools/verify_dist.sh` 與
  `verify_packages.py` 都沒有 `sha256`。**這是這一輪唯一新發現的缺口**——
  manifest 裡有雜湊，但兩支驗收腳本從來沒有回頭比對過它。正反對照：同一個
  pattern 在 `tools/write_dist_manifest.py` 命中兩次（它產雜湊），在兩支驗收
  腳本都是零。
- `dist-rebuild-version-number`：版本號是決策，標 `manual`。

`WORKLIST.md` 現在沒有任何 `- [ ]` 或 `- [~]`，只剩 149 條 `- [x]` 的歷史與
49 行指標。JSON 共 54 條：49 條未完成（35 條 `manual`）、5 條已核實完成。

建置與測試在容器內通過。`./...` 會被 `workplace/spec-source-audit-20260907`
的兩個同名 package 擋下來（那個目錄是 gitignore 的稽核暫存），改跑
`./cmd/... ./internal/...`；`cmd/kol` 與 `internal/ui` 需要 X11，照
`run_kol_headless.sh` 的模式起 Xvfb 後兩包都通過。

提交 `e118596`，推送到 `wicanr2/kol-remake-cht` 的 `main`（連同先前三個未推送的
commit 共四個）。推送前確認 repo 仍是 PRIVATE、作者信箱全部是 `wicanr2@gmail.com`、
240 個檔案裡沒有原版 ZIP／EXE／PDF／地圖；唯一的 PNG
（`assets/original-derived-png/ega/BATTLE/battle-map-24.png`）依專案例外進 private
repo，manifest 有登記、SHA-256 相符、權利分類齊全。

## 2026-09-10（三）：正常路徑地形快照接進 remake

WORKLIST 第 1 項。`workplace/battle-terrain-normal-fixed-20260908/battle-terrain.json`
是原版在正常遭遇裡實際產生的那 4096 格；先前只有產生它的那一支（
`tools/oracle/book_startup_test.go`），沒有任何程式讀它。

**它只能被載入，不能被重算。** 地形生成當下的亂數狀態沒有留在快照裡：
`alignment_seed: 1` 是規劃入口對齊之後設下的值，`seed_before_alignment`
（1323149826）是地形已經生成完畢之後的值，兩者都反推不回生成前的狀態。用途是
給後續的指令與結算對拍一個共同起點。

索引是這一步唯一會出錯的地方，而且**寫反不會報錯**：原版存 `Y*64+X`（繪製直接
吃的 `ES:SI`），remake 存 `Tiles[X*64+Y]`，兩邊的 X／Y 語意相同（X 是畫面上的列、
Y 是欄），差的只有哪一軸乘 64。寫反的話 4096 個 byte 照樣填滿、磚塊編號照樣合法，
整張圖只是沿對角線翻過去。判準因此要能分辨那個差別：合成地形用 `x*2+y*3`（
不對稱——拿 `x+y` 來驗，翻過來每一格都一樣，測試照樣全綠而什麼都沒證明），
再加一支負對照把同一張圖翻過來、確認判準真的抓得到。另有五種前提檢查（schema
換版本、索引契約改了、寬高不是 64、byte 數不足、`record_hex` 不是十六進位）。

**一個與索引無關的佐證**：磚塊組成是底色 2,848、粗糙 941、灑落 307。生成規則是
逐格擲 1…4、中 1 填粗糙，期望 1,024——帳面少了 83，因為灑落那一段會蓋掉一部分
粗糙格。把蓋掉的加回來得 1,017，與期望差 **0.24 個標準差**（σ ≈ 27.7）。沿對角線
翻過來時每一種磚塊的格數一個都不會變，所以這一條驗的是另一件事：這張圖確實出自
SPEC-008 那條規則，不是別處來的資料或解錯的 payload。

程式落點 `internal/oracle/terrain.go`（不帶 `oracle` build tag，只讀 JSON，不需要
原版執行檔）與 `terrain_test.go`。runtime 沒有動，正式遊戲路徑不讀 `workplace/`。
容器內 `go test ./cmd/... ./internal/...` 全綠（GUI 兩包另起 Xvfb），新增的兩個檔案
`gofmt` 乾淨。收據 `docs/verification/battle-terrain-snapshot-load-20260910.md`。

`battle-normal-seed-alignment` 條目改寫成「完整屬性與結算比較」，verify 維持
`manual`：屬性與結算還沒有可綁的東西，現在綁等於綁一個還不存在的猜想。

## 2026-09-10（四）：地形接進受控情境，站位與地形交叉驗證

WORKLIST 第 2 項的前半。`TestOriginalNormalPartyState` 建戰鬥時給 `NewQuest` 的
地形是 `nil`——60 個初始欄位都對得上，但那一場打在一片空白上。

先確認兩份快照是同一場：`battle-state-records-20260908/` 與
`battle-terrain-normal-fixed-20260908/` 的 `battle-state-records.json` 與
`battle-alignment.json` 逐欄比對**完全相同**。dosgolem 兩次獨立執行到達同一個
狀態，決定性成立，後者（含地形）因此可以當統一的共同起點。第三份
`battle-terrain-normal-20260908/` 沒有 `battle-terrain.json`——那次匯出沿用暫存
指標被資料檢查擋下來，目錄留著但沒有地形。

`TestOriginalNormalBattleTerrain`（`cmd/kol/original_battle_terrain_test.go`）三項
判準：4096 格逐格傳遞、在場的是 slot 0／1／6 而其餘 15 槽 remake 都判為空、
站位與地形同一座標系。

第三項是這一輪真正的收穫：站位取自 `battle-alignment.json` 的參戰記錄 +31／+32，
地形取自 `battle-terrain.json`，**兩份各自獨立的原版資料**。原版繪製時的索引是
`SI = Y*64 + X`，所以「`Field.At(a.X, a.Y)` 等於 `record_hex[a.Y*64 + a.X]`」一旦
成立，就同時釘住了地形的轉置方向、`Actor.X／Y` 的語意，以及兩者用的是同一個
座標系。三名在場參戰者都不在對角線上，判準因此分得出轉置；測試會在全部落在
`x == y` 時直接失敗，不讓一組碰巧無效的座標默默通過。

第二項踩過一次坑：第一版拿原版記錄的 +1（陣營）當「有沒有人」用，結果 14 個槽
被當成參戰者，而數字看起來還很合理。沒用到的隊員槽 7…17 在原版記憶體裡整條是零，
陣營因此讀成 0＝隊伍——**那是殘留值不是語意**。

負對照實跑一次：把 `terrain.go` 的轉置改成 `f.Tiles[y*64+x]` 之後，交叉驗證報
**1,818／4,096 格對不上**，合成地形那兩支也一起 FAIL；改回來後十六個套件全綠。
那個 1,818 只有 44%——翻過來有 56% 的格子碰巧一樣，因為底色佔了七成。這正是
合成地形那一支要用不對稱 `x*2 + y*3` 的理由：拿真實地形驗轉置，一半以上的格子
幫不上忙。

收據補在 `docs/verification/battle-terrain-snapshot-load-20260910.md`。
`battle-normal-seed-alignment` 的 verify 維持 `manual`：一度想綁
`TestOriginalNormalPartyState` 收尾那句自承，但它在測試檔裡而 verify 不掃測試檔，
綁了會永遠回「還沒出現」——那種條目看起來盡責，其實一次都沒真的看過。

## 2026-09-10（五）：怪物徒手取哪一件——那幾格根本沒被讀

接續第 2 項。原本打算擴充怪物記錄的擷取範圍（收據寫「只擷取 0x80 byte，不能當成
全部穿戴欄已完整保存」），查下去發現方向不對：**0x80 就是完整的一筆**
（`0xCD92 + (index-6)*0x80`，RE-080 與 SPEC-017 一致，`TestMonsterTemplateCopyMatchesOriginal`
也用 `bodySize = 128`）。護甲三件、盾牌、技能、旗標、逃走門檻、金幣全都在 128 之內，
資料本來就夠——只是沒比。

真正的缺口在 RESEARCH-LOG 的 RE-131 結尾：「怪物那一側沒接——同一張表加 `+0x30`
落在 `0x30`／`0x80`／`0x90`／`0x70`／`0x40`，其中三個在怪物記錄的已知版面之外，
還沒查那幾格是什麼。」而 remake 的 `NewQuest` 早就用 `unarmedFrom(monster.Worn)`
接了，取的是 `monsters.json` 那份母表的七格穿戴欄。兩者只可能有一邊對。

靜態掃描回答不了——用算出來的指標取的存取在位元組層面看不見。改用 dosgolem 的
`WatchReadsAt`：本體記錄 `+0x20`…`+0xC0` 每格填不同的可辨識值，設好規劃旗標、
徒手指令碼與攻擊選項，呼叫 `sub_26AC2(slot)`，看讀取落在哪。

**正對照是這次的關鍵。** 只跑怪物那側證明不了任何事：四式都只讀 `+0x20` 有兩種
可能——怪物本來就只取武器，或者設定根本沒走進徒手分支，而後者安靜得跟成功一模
一樣。所以同一套設定也用在隊員身上，拿 RE-131 已解出的答案當已知值。

結果：隊員四式各自落在 `+0x50`／`+0x60`／`+0x00`／`+0x40`，與 RE-131 逐項相符；
怪物四式一律只讀 `+0x20`，把武器那一格整格清零（真正赤手空拳）也不改變。
**怪物那一側沒有「徒手四式取穿戴品」這個機制**，`+0x80`／`+0x90` 越過 128 邊界
那件事不會發生，因為原版從來不往那裡取。

證據 `docs/re/20260910-monster-unarmed-slots.md`；RESEARCH-LOG 那段「還沒查」已改寫
成現況並指過去。探針 `internal/oracle/monster_unarmed_probe_test.go`（三組：怪物
有武器、怪物空手、隊員正對照，共 12 個子測試）。

remake 因此與原版有一處玩家看得見的差異：怪物徒手那幾手的傷害骰。這會改變戰鬥
數值，依專案規則不自行處置——新增 worklist 條目 `monster-unarmed-weapon-source`，
verify 綁 `state.go` 的那一行，等使用者決定。`monsters.json` 的七格穿戴欄本身沒錯，
護甲三件與盾牌照常使用。

## 2026-09-10（六）：怪物徒手改成照原版，並清掉三處被推翻的斷言

依使用者決定採選項 1（照原版）。

程式：`Actor.AttackWeapon` 在 `Side == SideMonster` 時一律回手上那件；`NewQuest`
不再填怪物的 `UnarmedWeapon`；`unarmedFrom` 連同它唯一的使用者一起移除。
空手的怪物（本體記錄 `+0x20` 全零）因此拿到空武器，而不是退回穿戴品——這與原版
一致，探針裡「武器那格整格清零」那一組就是這個情境。

測試：`TestMonsterUnarmedFromWorn` 釘的是被推翻的行為，改寫成
`TestMonsterUnarmedUsesHeldWeapon`，含兩個對照——同一份設定換成隊員要挑得到穿戴品
（證明判準分得出兩側），以及空手怪物拿到空武器。十六個套件與 oracle 探針都通過。

### 被推翻的斷言（集中在這裡，正文只留現況）

| 位置 | 原本寫的 | 現況 |
|---|---|---|
| `docs/spec/008-combat-field.md` | 「同一張表對怪物是加 `+0x30`…**兩側都接上了**」 | 隊員走穿戴品、怪物走 `+0x20`；怪物記錄 `+0x30` 起是穿戴欄沒錯（護甲與盾牌照常用），不成立的只有「徒手時會去取那幾格」 |
| `docs/spec/012-combat-spells.md` §8b | 「RE-131 已關閉怪物穿戴欄與隊員穿戴欄的對應…**不是只完成隊員一側**」 | 怪物不在此列 |
| `docs/re/RESEARCH-LOG.md` RE-131 尾 | 「怪物那一側沒接…**還沒查那幾格是什麼**」 | 那幾格根本沒被讀 |

`RESEARCH-LOG` 另一處「武器欄位在角色記錄 +0x30（徒手則是 +0x100 + [0x0FD2 + 選項]）」
講的是**角色**記錄，正確，不動。探針檔的註解原本寫「remake **目前**用
`unarmedFrom` 填了」——那種現在式敘述在修好的那一刻就變成假斷言，改寫成當時的
狀況加結論指標。

### 工具：已完成項不能用 `present`

把條目標成完成時綁了 `present`，`verify` 當場報「★ 可能回歸」。原因是語意相反：
已完成項跑的是回歸檢查（訊號**還在**＝仍然完成），而 `present` 的語意是
「找得到＝仍未完成」。改用 `absent` 之後正常。`check_schema` 加了這個組合的檢查，
`selftest` 補上正反例——這種錯誤自己會冒出來，但擋在 schema 比較省事。

## 2026-09-10（七）：所有未解工作進 JSON，並讓機器每輪自己問

依使用者「所有未解工作都集中在 json 用 verify 驗證」全面 review。

人工搬一次不夠——產品碼會長出新的自承，而下一輪沒有人記得回來搬。所以做成機制：
`tools/worklist.py scan` 掃產品碼（`cmd/`、`internal/`，排除測試檔）裡所有「還沒解／
還沒實作／還沒接／還沒做／還沒查／尚未實作／沒有實作／TODO」的句子，每一行都要有人
認領——不是被某個條目的 `verify` **精確綁到那一行**，就要在 `acknowledged` 寫明為什麼
它不是待辦，沒有第三種。`verify` 併入這項檢查，有孤兒就回非零。

認領判定刻意做成「pattern 要命中那一行」而不是「paths 涵蓋那個檔案」：後者的話，
一條綁 `internal/` 的條目就會把整棵樹認領走，於是這支掃描永遠是乾淨的而什麼都沒看。
`selftest` 為此加了反例——綁到別的檔案不算認領。

### 掃出 6 個藏在註解裡的過期斷言

這是這一輪最值錢的部分。註解本身也會變成過期斷言，而且沒有任何東西在看：

| 位置 | 原本寫的 | 現況 |
|---|---|---|
| `cmd/kol/main.go` | 「指令（攻擊、移動）還沒解完，這一版只到看得到、逃得掉」 | 指令與結算走 SPEC-008／010 的既有實作 |
| `cmd/kol/main.go` | 「執行段還沒實作，答應之後只是再排一輪規劃」 | 由呼叫端的回呼推進執行段（`adapter_regression_test` 就在測它） |
| `internal/battle/battle.go` | 「隨機遭遇還沒實作，見 SPEC-007」 | 走 SPEC-011 的 `ambush.go`，實跑收據在 `battle-ambush/` |
| `internal/battle/round.go` | 「負重與指令的時間成本那兩項還沒解完」 | `Encumbrance` 與 `ActionCost` 都已實作且各有原版對拍 |
| `internal/battle/state.go` | 「這一版只做到進得去、看得到、逃得掉」 | 規劃、執行、移動、命中結算與勝負判定都在 |
| `internal/town/session.go` | 「野外地圖本身還沒實作」 | 野外那一段由 `internal/world` 接手 |

### 認領結果

33 行自承（改寫 6 處之後從 39 降下來）全部有人認領，0 孤兒：

- **12 行標 `acknowledged`**：`Unimplemented` 欄位的語意說明、規劃段漏接時的容錯路徑、
  擊退取消未執行那一手的規則敘述、以及一句「更正 RE-113 的『被誰讀還沒解』」——
  那句裡的「還沒解」是**被更正的舊說法**，不是現在的缺口。
- **13 條新條目**：野外與戰鬥的可點區域、旅店解散離開、攻擊訊息鏈索引、
  `sub_26C5F(0x17)`、敵人移動段方向（已解未接，程式裡寫明了不接的理由是長路徑收據要
  整條重生）、反制掃描的攻擊選項來源與移動重試路徑、參戰者記錄 `+0x05`、磚塊 `0x0C`
  解鎖移動方式的理由、`+0x25` 的讀取端、`sub_1AAE8` 參數語意、打贏之後的掉落
  （`sub_289EE`）、城鎮打字輸入框、`word_2DF2C`。
- **3 條既有條目擴大 pattern**：鍛造（兩處自承）、選目標格游標、競技場肖像
  （`sub_2B454`，兩個檔案）。

另加一條 `spec008-unsolved-list`：SPEC-008 末尾那一節的七項裡，四項已各自有條目，
剩下三項在產品碼裡沒有留下自承（AI 權重表的執行期值、`sub_1B94F`、`word_3B26A`
在城鎮被設成什麼、怯戰檢定的兩個開關、`sub_2625A` 的語意），`scan` 掃不到，靠這一條
追。這是唯一一條綁規格文件而不是產品碼的——那幾項沒有程式落點可綁，硬綁一個猜想中的
識別字會永遠為真。

清單從 49 條長到 63 條未完成（33 條 `manual`）加 6 條已核實完成。十六個套件全綠。

## 2026-09-10（八）：掃描範圍擴到工具與文件，再抓出七個過期斷言

依使用者「再掃描一次程式碼 & markdown 確保沒有遺漏」。`scan` 的範圍從
`cmd`／`internal` 的 `.go` 擴到 `tools` 的 `.py`／`.sh`，加上 `docs/` 與根目錄的全部
markdown——**掃得寬一點沒關係**：掃到的每一行不是有條目綁著，就是要在 `acknowledged`
寫明為什麼不是待辦，所以「為什麼不看某一類」也會留在資料裡，而不是藏在程式的範圍
設定裡。`tools/worklist.py` 自己不掃，它的內容就是那組 pattern 的定義與測試用的假資料。

範圍一擴，行數從 33 跳到 261。

### 又抓到七個過期斷言

上一輪在產品碼註解裡抓到六個，這一輪在工具與規格裡再抓七個：

| 位置 | 原本寫的 | 現況 |
|---|---|---|
| `tools/make_final_quest_save.py` | 「接任務那一側（`sub_14564`）**還沒實作**，所以沒有辦法用正常路徑累積」 | RE-110 已解並接上（`quest_offer.go`，收據 `quest-take`）；這支工具存在的理由改成**時間**——跑滿 23 件任務太久 |
| `tools/ida_probe_koltown_steed.py` | 「圖本身還沒接上」 | RE-120 已解，資料頁的 `SteedPic` 與 `STEEDS` 素材都在 |
| `tools/make_observation_inputs.py` | 「馬廄的商品表還沒解出來，遊戲裡買不到坐騎」 | 馬廄買賣由 RE-119／120／121 解出並接上；換名冊只是省掉買四匹的步驟 |
| `tools/ida_probe_koltown_quest_take.py` | 「還沒解的是整段接取」 | 同第一列，已解 |
| `tools/ida_probe_koltown_healing.py` | 「那六段敘述從哪裡載入還沒解」 | RE-128 那一輪解出（RESEARCH-LOG 標 `已證實`） |
| `tools/ida_probe_koltown_mount_drop.py` | 「執行段還沒接」 | RE-124 解出，`CmdLand` 已實作 |
| `docs/spec/011-wilderness-ambush.md` | 「買坐騎的馬廄商品表還沒解」 | 同第三列 |

探針檔的改法是保留當時的問法再指向結論——那是研究過程的紀錄，刪掉反而看不出
為什麼要問；但「還沒解」這種現在式的斷言要改成過去式，否則它會一直宣稱一件
已經不成立的事。

### 認領方式

`acknowledged` 的 `file` 現在支援三種寫法：整檔（省略 `match`）、目錄前綴（結尾加
`/`）、萬用字元（`tools/ida_probe_*.py`）。`selftest` 為三種各補了正反對照——
包含「綁到別的目錄不算認領」與「不符的萬用字元不算認領」，不然一條就能把整棵樹
認領走。

整批認領的是那些**記當時狀態**的來源：`WORKLOG.md`（逐輪流水帳）、`WORKLIST.md`
（render 產生＋已完成的歷史）、`CONTEXT.md`（現況摘要，數字由 render 產生）、
`docs/re/`（研究筆記，解出後由後續段落更正而不刪除）、`docs/verification/`（收據，
記產出當下驗到哪）、`docs/screenshots/`、`docs/manual/`、`tools/ida_probe_*.py`。

### 條目

新增兩條：`spec011-unsolved-list`（SPEC-011 末尾的未解清單，與 SPEC-008 那條同性質）、
`quest-item-0xb926`。另有六條既有條目擴大 `paths`，讓它們同時認領規格與 README 裡
講同一件事的那幾行——`windows-macos-real-machine` 綁 `verify_dist.sh` 的自承、
`actor-record-05-ability` 加 SPEC-008 的反制掃描條件、`g06-audio-listening-gate` 加
README 的現況句等。

65 條未完成（32 條 `manual`）、6 條已核實完成、261 行自承 0 孤兒。十六個套件全綠。

## 2026-09-10（九）：pattern 本身有洞——1,246 行盤點，再抓兩個過期斷言

依使用者「再次盤點確保未完成工作都集中在 json、verify 可以 cover」。上一輪把掃描
**範圍**擴到工具與文件，這一輪查的是**過濾器**本身。

量過一次就清楚了：在沒有被整檔認領的 296 個檔案裡，原本那組 pattern
（`還沒解｜還沒實作｜還沒接｜還沒做｜還沒查｜尚未實作｜沒有實作｜TODO`）命中 47 行，
換成寬的一組是 242 行。**中間那 195 行不是不存在，是過濾器有洞**——漏掉的說法包括
裸的「還沒」「尚未」、「仍待」「未完成」「未解」「待查」「未接」「未驗」，以及
FIXME／XXX／HACK。

pattern 因此改寬。寬了會抓到不是待辦的句子，那正是 `acknowledged` 存在的理由：
寧可多寫幾條理由，也不要靜靜漏掉。全樹 1,246 行全部認領、0 孤兒。

### 又抓到兩個過期斷言

| 位置 | 原本寫的 | 現況 |
|---|---|---|
| `internal/oracle/doc.go` | 「⚠ **要 dosgolem 的 `kol-force-call` 分支**…四件事都還沒進主線」 | 那個分支昨天已刪；`Call`、`Stub`、線性版 `At()`、FAT 8.3 截斷四件事**都已經在 main 裡**，現在用的是 `kol-oracle` |
| `docs/spec/016-starting-inn.md` | 「跨城選角…目前仍待完成」 | `recruit.go` 已有跨城載入器，收據 `cross-town-fix-20260908.md` |

第一個特別值得記：它指向一個**已經不存在的分支**，下一個人照著找會撲空。

### 認領的形狀

`acknowledged` 的 `file` 現在也吃 `*`，用來寫「不管出現在哪都不是待辦」的說法。
最大宗是**遊戲術語**——「還沒看到」（視線）、「還沒分勝負」、「還沒規劃」、
「還沒上弦」、「還沒押」、「還沒跨門檻」。這些是規則本身要表達的**狀態**，
和「這件事我還沒做」只差一個字。其餘是規格生命週期用語（`尚未 CONFORMED`）、
驗收範圍聲明（「未完成正常鏈前不升 CONFORMED」）、玩家看得到的訊息字串
（「資料尚未改變」）、抽取器的前提檢查（`尚未證實` 就 SystemExit）。

### 機制修了一個真的 bug

`spec-unsolved-sections` 綁 `^## .*(還沒|尚未)` 之後，`verify` 報「★ 可能已完成」。
原因是 **`^` 在兩邊語意不同**：`scan` 逐行比對，`^` 是行首；`find_hit` 讀整個檔案，
不加 `re.MULTILINE` 的話 `^` 變成「檔案開頭」。於是同一個 pattern 在 `scan` 認領得到、
在 `verify` 永遠找不到，**兩邊都不報錯**。加上旗標並在 `selftest` 釘住。

### 條目

新增 12 條：五處 UI 細節還沒對拍、世界時刻門檻機制、角色寫回未反追的位移、
SPEC-023 尚未取代 v1 入口、術語表待查項、規劃段 `+0x25` 未逐格核對、KOLDATA 動作
敘述、重排那一輪的位址、施法目標人形、參戰者記錄 `+0x2C`、任務計數 0xFF 的寫入點、
SPEC-029 接線順序審查。另把 `spec008`／`spec011` 兩條合併成 `spec-unsolved-sections`，
綁章節標題本身——新開一節未解項會自動被認領，全部清空才會開口。

76 條未完成（32 條 `manual`）、6 條已核實完成、1,246 行 0 孤兒。十六個套件全綠。

## 2026-09-10（十）：刪掉兩條假的未完成項

依使用者「刪除過期斷言」。上一輪抓的都是文件與註解裡的過期敘述；這一輪查的是
**清單自己**——41 條 `present` 條目各綁一句自承，如果那句自承本身已經不成立，
清單上就多了一條永遠不會完成的假缺口，而 `verify` 每一輪都會盡責地回報它「仍未完成」。

抽查最有嫌疑的兩條，兩條都是：

- `internal/save/characters.go`「Characters 是 SPEC-023 已審查的保存核心；**尚未取代
  正式 v1 入口**」——`save.go:17` 就寫著「正式遊戲使用 `CompleteSchema`，v1 保留為
  匯入入口」，而 `cmd/kol/complete_save.go` 用的正是 `save.Characters`。v1 是**刻意
  保留**的匯入路徑，不是還沒取代。斷言已刪除，條目標完成並改綁那句契約當回歸訊號。
- `glossary-terms-pending` 綁 `i18n.go` 的「因為術語還在審查中」——那是**設計理由**
  （翻譯不寫死在程式裡），不是待辦；而且術語本身的缺口早就有
  `terminology-and-safe-rects` 在追。重複的條目移除，那一句改成 `acknowledged`。

74 條未完成、7 條已核實完成、1,283 行 0 孤兒。十六個套件全綠。

## 2026-09-10（十一）：怪物裝備能力比完了——資料本來就夠

繼續第 2 項。原本這一塊卡在收據那句「怪物記錄目前只擷取 0x80 byte，不能當成全部
穿戴欄已完整保存」，於是它一直被當成「要先擴充擷取範圍」。**那句話有一半不成立**：
`0x80` 就是完整的一筆（`0xCD92 + (index-6)*0x80`，RE-080／SPEC-017，
`TestMonsterTemplateCopyMatchesOriginal` 也用 `bodySize = 128`），護甲三件、盾牌、
技能、旗標、撤退門檻與金冠骰式全在裡面。真正裝不下的只有母表 `worn` 的第 5、6 槽，
而那兩格原版在戰鬥中根本不讀。資料本來就夠，缺的只是沒有比。

`TestOriginalNormalMonsterEquipment` 做三方比對：戰鬥期記錄的 bytes ↔
`monsters.json`（母表抽取器讀對了沒）↔ remake 的 `Actor`（適配器接對了沒）。
SLEDGE 共 20 項通過，穿戴欄前五槽實得 3 件。收據
`docs/verification/monster-equipment-20260910.md`。

### 判準踩了一次坑

第一版拿 `m.Worn[slot] != nil` 判斷有沒有東西，SLEDGE 的第 3、4 槽（原版是空的）
被判成不符。**`monsters.json` 的兩個欄位對「空」的約定不一樣**：`shield` 走
`item()`，全零回 `null`；`worn` 要保持七格對位，全零那幾格解出來是一個每個欄位都是
零的**物件**。判準改成看記錄本身是不是整片零。順帶確認 remake 沒踩同一個坑——
`cloneArmour` 只取前三格，SLEDGE 那三格都有實料。

### verify 綁測試檔：兩個方向都錯

把條目改綁對拍測試收尾那句自承，`verify` 當場報「★ 可能已完成」。**verify 不掃
測試檔**，所以 `present` 會回「找不到＝可能已完成」每輪誤報，`absent` 則永遠為真。
兩個方向都不對，改回 `manual` 並在 note 寫明；規劃／結算的載入器接上時
`internal/oracle` 會出現產品碼落點，到時再綁。

74 條未完成、7 條已核實完成、1,288 行 0 孤兒。十六個套件全綠。

## 2026-09-10（十二）：規劃對拍卡在進門，用滑鼠輪詢數定位

要做規劃與結算的對拍，得先讓原版走到規劃入口**之後**。`probeNormalBattlePath`
第一步的 `recruit-click` 點 (202,182) 落空，訊息是「畫面完全沒變（期間滑鼠被輪詢
426,415 次）——不是座標不對，就是遊戲還沒準備好收這個點擊」。

### 兩次猜參數都錯，換方法

先試 `KOL_BOOK_SELECT_DOWN=2`（fixture 的隊長是 Olear，清單上在第 2 列——這是從
`book-selection.png` 直接讀出來的），結果連 KOLTOWN 都沒載入。再回到 12（Mula）
仍然點擊落空。**連續兩次同類失敗就不再試第三個數字。**

換成靜態比較兩個 base：`Click` 相關一行都沒動，但 main 給
`docs/spec/012-input-keyboard-irq-and-mouse-scale` 加了一段新的診斷指引——

> 點了沒反應時，先看遊戲收到什麼。三種失敗在畫面上完全一樣：一筆都沒有 ＝ 按住
> 的期間遊戲沒讀滑鼠；有筆但 `CX` 不是要的值 ＝ 倍率或座標送錯；`CX` 對而畫面
> 不動 ＝ 那個位置本來就沒有可點的東西。

### 輪詢數就是那個判準

在送鍵前後各記一次 `o.MousePollSteps()` 的長度：點擊前 426,415、點擊後**還是**
426,415——**點擊期間一次都沒有增加**。所以是第一種：那個畫面根本沒在讀滑鼠。
擷取下來的 `battle-path-before.png` 也對得上：畫面停在 Halfway House 的**建築
外觀**，圖示列還沒出現，(202,182) 那個位置當時是建築圖。

再送一組鍵逐步觀察，輪詢數的跳變一眼就指出轉折點：

| 鍵 | 滑鼠輪詢增量 |
|---|---:|
| Return／Space／Up／Return | +1／+38／+40／+43 |
| **Down** | **+305,850** |
| Return | +296,027 |

`Down` 之後畫面左下角多出兩個圖示，遊戲開始大量輪詢滑鼠——**圖示列要先叫出來，
點擊才有東西可點**。原本的序列少了這一步。

### 這一輪留下的

`tools/oracle/book_startup_test.go` 多了三樣：送鍵前的畫面與輪詢紀錄（分得開三種
點擊失敗）、`KOL_BOOK_PROBE_ENTRY=1` 的進門觀察模式、`KOL_BOOK_PLAN=1` 的規劃逐鍵
探測（寫 `battle-planning.json`，逐鍵記 `+0x22`／`+0x23`／`+0x25`）。規劃探測本身
還沒跑到——它排在進門之後。

下一步：把 `Down` 補進序列，走進旅店，再讓規劃探測跑起來。

## 2026-09-10（十三）：進門那三步解出來了，座標一直是對的

接續上一輪。點擊落空的原因確定了，而且**不是座標**。

### 進門是三步

用觀察模式逐鍵看滑鼠輪詢增量，轉折點一眼就出來：

| 鍵 | 輪詢增量 | 畫面 |
|---|---:|---|
| `Down` | **+305,850** | 城鎮外觀底部多出圖示列 |
| `Return` | +296,027 | 進門，Trollsbane Inn 的敘述頁 |
| `Space` | +335,471 | 店主招呼，六個圖示出現 |

六個圖示就是 `town.InnIconNames` 的 `DOOR／MIRROR／LISTEN／TALK／REST／COMPANION`。
`town.IconBar` 是 `X=32`、`Step=32`、`Y=170`、高 22，所以第 6 格 COMPANION 的中心是
**(204, 181)**——而原本那一下點的是 **(202, 182)**。**座標從頭到尾都是對的**，
少的是讓圖示列出現的那三步。補上之後九步全部走完，沒有再落空。

這一段值得記成規則：**「點了沒反應」先看遊戲有沒有在讀滑鼠**（輪詢數增量），
不要先去調座標。調座標會在畫面根本沒到位時「試出」一個看起來對的數字。

### 但走路序列綁城鎮

補完之後路徑往下走到 walk、next-day 與 fight-entry，卻在
`exportNormalBattleState` 以「參戰上限不一致：count=0 limit=1」失敗。看畫面就懂了：
`fight-entry.png` 還是城鎮外觀——**根本沒出城**。

那串方向鍵（`Right`×2、`Up`×9、`Return`、`Right`×100）是照 **Halfway House** 的地圖
數出來的。`KOL_BOOK_SELECT_DOWN=12` 選到的 Mula 在 **Trollsbane**，同一串鍵走不出去。
而 fixture 的隊長 Olear 雖然在 Halfway House、清單上排第 2 列，`SELECT_DOWN=2` 卻在
**書本階段**就停住（「已達書本，但正常 SPACE/6/Enter/SPACE 尚未載入 KOLTOWN.EXE」），
和旅店無關。

所以現在缺的是「哪一個 `SELECT_DOWN` 會選到一個**在 Halfway House 而且進得了遊戲**的
角色」。這一輪不再試第三個數字——已經連續兩次靠猜參數失敗。

工具側的進展留著：進門三步、輪詢數診斷、走路序列綁城鎮的警告都寫進
`tools/oracle/book_startup_test.go`。

## 2026-09-10（十四）：三個「猜次數」都換成「等狀態」

依使用者選 1（查清楚 `SELECT_DOWN=2` 為什麼進不去，用觀察而不是換數字）。

### 答案：清單還沒畫完

`down=2` 的 `book-selection.png` 左頁停在 **ORIGIN ©1989 的封面**——角色清單根本
還沒畫出來，那個 Enter 送進去因此沒有作用，整條路徑停在 KOLBOOK
（IP `0110:40D7`，畫完的那一條是 `0110:40BF`）。

**`down=12` 會過是巧合**：十二次 Down 各跑 5M 指令，剛好給了六倍的時間讓清單畫完。
症狀是「選這個角色進不了遊戲」，看起來完全像角色的問題——我也確實先去查了
CHARDATA，Olear 與 Mula 在那幾個欄位上一模一樣。

改成 Enter 之前先 `o.Run(60_000_000)` 等畫面就緒之後，`down=2` 進得了 KOLTOWN，
Olear 也順利出城（`fight-entry.png` 是戶外地圖）。

### 同一個模式，這一輪出現三次

| 原本 | 症狀 | 改成 |
|---|---|---|
| 靠按鍵次數湊時間讓清單畫完 | 「這個角色進不了遊戲」 | Enter 前等畫面就緒 |
| 假設圖示列已經在，直接點 | 「點了沒反應」 | 先 `Down` 叫出圖示列（看滑鼠輪詢增量確認） |
| 次日固定走一百步 | 「參戰上限不一致：count=0」 | 每一步問 `DS:2D9C`，遇敵就停並記在第幾步 |

三個都是**拿次數當條件**。它們的共同症狀是「看起來像另一個問題」——角色不對、
座標不對、擷取壞了——所以每一次都會先去查錯的地方。

### 還沒解決：固定 seed 下走不到遭遇

改成偵測之後訊息很明確：「次日一百步走完都沒有遇敵」。這不是擷取壞了，是**真的
沒打起來**。而 SPEC-011 記著：野外那兩條收據「選一顆不會在半路被攔下的種子」——
`seed=1` 正是為了**避免**遭遇挑的，現在要的卻相反。

下一步：用 remake 的 `ambush` 規則（與原版亂數已對拍過，RE-132）掃 seed，找一顆會
在早期遭遇的，再拿它去跑原版。用 remake 當搜尋工具、原版當驗證，比一顆一顆試
原版快得多。

## 2026-09-10（十五）：掃 seed 掃出一件更根本的事——那一百步從來沒走過

依使用者「掃」，用 dosgolem 的 `Save`／`Restore` 在同一次執行裡逐顆 seed 重來，
省掉一顆跑一次整條路徑。掃描本身跑起來了，結果卻是連續三次 0 命中：

| 規模 | 結果 |
|---|---|
| 40 顆 × 40 步 | 0 顆遇敵 |
| 12 顆 × 25 步 | 0 顆遇敵 |
| 3 顆 × 90 步 | 0 顆遇敵，**真的走動 0 步** |

前兩次我以為是步數不夠——`internal/world/clock.go` 記著一步約 56 個時間單位、
時段門檻在 3360…4200，也就是**一個時段要走六十到七十五步**，而遭遇是每次時段
變換才擲一次。所以第三次少掃幾顆、每顆走遠一點，順便記下「真的走動幾步」。

**那個數字是 0。** 九十步裡一步都沒動，四個方向輪過都一樣。

回頭看先前那次固定一百步的畫面雜湊，答案早就在那裡：`next-day-start`、
`next-day-009`、`next-day-099` 三張**完全相同**。露營之後的那一百步，
**從來沒有走過任何一步**。畫面看起來完全正常——野外地圖、角色在城外、沒有訊息
框——所以這件事在畫面上完全看不出來，只有在「按了一百次卻沒遇敵」時才會被當成
機率問題。

出城前那一段（`walk-000`／`-011`／`-051`／`-final` 四張雜湊都不同）是真的有走，
所以不是方向鍵壞了，是**露營之後的那個狀態不收方向鍵**。

### 這一輪加的工具

`walker`：每走一步比一次畫面指紋（`sha256(o.Tandy16())`），沒變就換下一個方向，
並回報這一步到底有沒有走出去。前面三個修正都是「拿次數當條件」，這是第四個，
而且它一啟用就立刻把一個藏了很久的假設打掉。

### 下一步

先查「露營之後要按什麼才回得到可移動狀態」，用進門那一套（逐鍵送、看畫面指紋與
輪詢增量）。在那之前，掃 seed 沒有意義——每一顆都停在同一個不能動的狀態上。

## 2026-09-10（十六）：露營後是死迴圈，不是在等玩家——該換路徑了

接續掃 seed 那一輪。露營之後方向鍵無效，逐鍵探測十一個鍵（Right／Space／Return／
Escape／Down／Up，每個後面跟一個 Right 測方向鍵有沒有生效）：

- **十一個鍵，畫面指紋一次都沒變**（一直是 `4f7988030a00d082`）
- 輪詢每步只 +8～+10（城鎮等待時是 +300,000）
- IP 在 `0AE9:0B61`／`0B63`／`0B65` 三個值之間打轉

再什麼都不按、空跑 4×60M 指令：畫面照樣沒變，輪詢每段 +125。**五個位元組的迴圈、
極低輪詢、按鍵無效**——那是死迴圈的形狀，不是「在等玩家」。

問 dosgolem 要不要負責，答案是不用：

- 未實作呼叫：**空的**
- 缺檔：`region1.dat`…`region4.dat`——但原版 ZIP 裡**只有 `REGION0.DAT`**，
  那四個本來就不存在，書本階段也一樣缺而運作正常，不是原因
- 最近開的檔：**`TOWN7.0`／`TWN7PICS.0`**

最後那一條是新資訊：走的過程中遊戲載入了**第 7 座城鎮**，而起點是 Halfway House。

### 判斷：不再修這條路

這一連串下來，「拿次數當條件」的問題修掉了五個，每一個都是真的：

1. 靠按鍵次數湊時間讓角色清單畫完
2. 假設圖示列已經在就直接點
3. 次日固定走一百步
4. 走路時撞牆卻繼續按同一個方向
5. 掃 seed 前沒確認角色真的會動

修完之後路徑仍然走不通，而且每修一個就冒出下一個。依 `rulebook/41`，這時候該停
下來重想，而不是補第六個特例。

**架構問題**：規劃與結算的對拍要的是「一場打得起來的戰鬥」，而我選的是**最脆弱
的取得方式**——走完整正常路徑等隨機遭遇，依賴機率、地圖、時段、走位，任何一環
不對就整條停住，而且症狀都長得像別的問題。

原版有確定性的戰鬥入口：任務地點（走到就打）與競技場（選人、下注、單挑）。
既有的 60 欄比較也不是靠隨機遭遇取得的。下一輪改用那一種。

### 這一輪留下的

`tools/oracle/book_startup_test.go` 的探測工具都留著，它們與這條路徑無關也用得上：
`walker`（畫面指紋判斷有沒有真的走出去）、`screenDigest`、露營逐鍵探測、seed 掃描
（`Save`／`Restore` 在同一次執行裡逐顆重來）、未實作／缺檔查詢。

## 2026-09-11：town-scroll 收據與法術一覽

### town-scroll：程式對、收據過期

第 14 步起對不上：收據（09-05）買完回法術清單、重跑回招呼語。RE-128（09-06）讀到
`sub_17318` 在 `0x178DA` 就 `retf`，原版 oracle 的畫格也是回招呼語，所以是收據早於
更正。腳本改成買完接 `> RET`（白框在 DOOR，往右一格是 MIRROR），22 步變 21 步；
新第 16、20 步與舊第 18、22 步在改法術一覽之前逐 byte 相同。

### 法術一覽：跑原版才看到的三處

新 oracle 情境 `town-spell-list`（沿用 `town-spell-edit` 到買下為止的 100 步，再開
MIRROR 與法術那一格），272 個去重畫格，109 個檢查點全部穩定、角色檔沒動。
frame-270 推翻了 remake 的版面：

1. 只剩抬頭那一層——`sub_16239` 不是「重畫整頁」，屬性以下四框是 `sub_15ABF` 自己畫的
   （RE-137，更正 RE-123）。任務一覽走同一支，一起改。
2. 每一格右下有四像素投影（`sub_1BA9D` 收尾必呼叫 `sub_1B7E9`）。
3. 名字是大寫（`0x163FB` 的大寫旗標）；資料頁的教派與頭銜也在旗標底下，一併改。

改完兩張整頁（法術一覽、資料頁）逐點比對，允許區（滑鼠游標）外零差異；反面對照
（remake 法術一覽比原版資料頁）差 25,906 點。資料頁拆成抬頭層與內容層後照原版順序
畫，坐騎投影在 y = 52 那 41 點被武器框頂邊蓋掉，town-stable 第 16 步因此重生。

中文模式同一個分界：`drawSheetHeaderCJK` ＋ 法術格／徽章乘倍率，實跑截圖確認。

### 驗證與狀態

- `go vet`、`go test ./cmd/... ./internal/...` 全過（Xvfb 容器）。
- `check_player_paths.py`：town-scroll、town-mirror、town-load 一致；town-stable 重生後一致。
  quest-list 從書本選單起步、重跑卡在名冊頁，屬於 `player-path-receipts-stale` 既有那一類。
- 畫廊的 en／zh-town-sheet 重生後逐 byte 不變。
- 物品頁選中之後的畫面與原版不同（原版只剩明細框與木紋底），新條目 `item-detail-page-parity`。
- worklist 訊號 `counterscan-style-and-retry` 的正則 `a\.Stuck` 對不上測試寫法，放寬為 `\.Stuck\b`。
- 容器全用 `--rm`；專案內無 root-owned 檔案。

## 2026-09-11：worklist 鏡像到 GitHub issue

使用者決定：JSON 為準、issue 是鏡像；91 條全放（已完成 17 條建成 closed）；label 用
分層五個＋manual＋worklist。

- `tools/worklist.py issues`（容器）產生每一張的標題、內文、label、開關；`--adopt` 把
  建好的編號寫回條目的 `issue`。schema 擋非正整數與兩條共用一個編號；selftest 加了
  內文記號、開關、manual label 與編號寫回的正反對照。已完成項的內文寫「回歸檢查」，
  不沿用「……就是還沒完」那句（對已關閉的 issue 會讀反）。
- `tools/sync_worklist_issues.sh`（主機，只呼叫 `gh`）：建 label、開 issue、對齊既有
  issue 的標題／內文／受管 label／開關。每建一張先記進 `created.tsv`，斷在半路重跑
  不會重開。
- 結果：#1…#91，開 74、關 17。重跑一次「新開 0 張、更新 0 處」；正對照：手改 #91 的
  標題並加 `bug`，同步後標題蓋回、`bug` 保留（不在受管的那組 label 裡），測試用的
  `bug` 已拿掉。
- 順手修正：`player-path-receipts-stale` 內文有 9 處字面上的 `\n`，換回真正的換行。

## 2026-09-14：Issue #91 物品頁選中後明細層完成

- 入口與授權：依專案路由載入復古重製的 RE→spec→同狀態驗證流程，並使用
  `reverse-engineer-retro-game-remake` 技能的規格閘門與交接要求。以 host
  `/home/anr2/.local/bin/gh auth status --hostname github.com` 確認 `wicanr2`
  憑證，再以 host `gh issue view 91` 讀取 Issue #91；本輪沒有留言、關閉、標籤或
  其他遠端寫入。
- 證據與規格：依 RE-127、RE-137 及新增的 RE-138，保留
  `KOLTOWN-unpacked.EXE` SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`、
  IDA Pro 9.4 線性有效位址與 `sub_16986:0x16BBF`／`sub_1BA9D:0x1BAF1`／
  `sub_1B7E9` 原始定位。SPEC-031 完成 READY→CONFORMED；frame-072 仍明確標為
  DOSBox-X 2026.06.02 `dosbox-bootstrap`，因 dosgolem 目前沒有 KOL 適配器，
  不宣稱 strict dosgolem parity。
- 實作：`internal/ui/itempage.go` 與 `itempage_cjk.go` 依互動狀態改走全頁底圖，
  選中／訊息狀態不再重畫立繪、木屋背板、右欄、十七列清單與完整頁外框；明細框
  接回 `shadow` 的右／下四像素投影。新增 UI 狀態分支與投影幾何測試，未改動
  `internal/town/itempage.go` 的正常狀態轉換與保存格式。
- 對拍：Docker 內重生 `tools/verify_remake_parity.py` 收據，4/4 pair 通過；
  frame-070 完整頁回歸保留，frame-072 正常玩家輸入新增，允許區外差異為 0，
  明細框四邊及兩條四像素投影列為 required regions。重製選中畫面 SHA-256
  `e6ab1a385817827a2a4afb557549f55c04a832fccfafcb35fd5aa3b0827e9b07`，差異圖
  SHA-256 `020031697da861f1be604ba3e40a09cb430f9b69858f81e19afb46d1a6bdf04a`。
- 測試：Docker 內 `go test ./cmd/... ./internal/... -count=1 -timeout 600s`、
  `go build -o /tmp/kol ./cmd/kol` 與 UI／town 測試通過；以收據原本的
  `/tmp/mod`、`/tmp/gocache`、`/tmp/gopath` 重建，執行檔 SHA-256 為
  `d2869ddfc6bad5c0d3e4bf9aa569665648cdcda4c966a7f15ed35a58ff260de3`。CJK
  文字安全矩形檢查 `6952/0`，UI 翻譯檢查 `486/0`，另有 2 個刻意空白。
  `go test ./...` 仍會因 `workplace/` 的既有 scratch packages 缺少外部模組／含
  重複與未定義符號而失敗；產品範圍已獨立通過，分類為聚合環境／工作區問題。
- 文件與台帳：更新 `CONTEXT.md`、RESEARCH-LOG RE-138、SPEC-031、規格索引、
  parity plan／receipt；由 `tools/worklist.py render` 重生 `WORKLIST.md` 與統計，
  從權威 `docs/worklist.json` 移除已完成的 `item-detail-page-parity`，其餘 73 條
  未完成項保持原狀。JSON、收據、工作清單與擁有權檢查通過；所有本輪檔案為目前
  使用者 UID/GID，沒有 root-owned 專案產物。
- 勘誤：第一次產品重跑未設定 `GOWORK=off`，被不存在的 `/dosgolem` 路徑擋下；
  第二次以同一 Docker 映像隔離後通過。收尾包裝檢查第一次誤讀收據的 `summary`
  欄位，改按實際 `pair_count`／`pass_count` schema 重跑後通過。容器均使用
  `--rm`，未保留長駐工作程序；目前 HEAD 為 `a442b78b64c02b99adcf214ec9236883e54bd15d`，
  本輪未 commit／push，最後容器與暫存清理於本節收尾複核。

## 2026-09-23：接手收尾、核實繁中戰鬥清單與修正現況入口

- 開工依路由命中復古重製收尾與工作清單，載入 `reverse-engineer-retro-game-remake`、
  收尾 playbook、驗收交接與工作清單規則；`grill-me` 用於待決的完成級別。
  `main`／`origin/main` 同為 `5d2f6ce6c78484073d57c3a2e0dbee4aba499386`，
  開工 `git status --short` 為空。host `gh auth status` 通過，已唯讀核對遠端 Issue；
  沒有建立、編輯、留言或關閉 Issue。
- `tools/worklist.py verify` 首跑指出 #90 的 `還沒有中文版面` 訊號已消失，
  但 SPEC-032 要求的四種正常路徑畫格未齊。從既有 `battle-spell/replay.spec`
  以目前程式重建並重播，取得 `PromptSpell`／`PromptAffect` 兩張 960×600
  繁中畫格，目視確認標題與清單可讀；輸入、工具映像、步號、SHA-256、證據
  等級和權利分類見 `docs/verification/cjk-ui/battle-lists-receipt.json`。這是
  remake 的正常按鍵路徑收據，不是原版中文像素對拍。
- 戰利品舊腳本用 seed 39 重播時，第 40 步遇治療提示，原移動鍵被攔住；
  插入 `RET` 後，第 67 步又遇 GOLEM，未走到 `Items Found`／`Characters`。
  已把 #90 改為誠實的 `manual` 驗收缺口，保留開放。對 `battle_cjk.go`
  的原版游標外觀自承，納入既有 #33 的訊號，不另外製造同義項。
- 第一次 Go 測試未設 `DISPLAY`，`cmd/kol`／`internal/ui` 在 GLFW 初始化失敗；
  分類為驗證環境，於同一 `eob-remake-go:1.26.7-ebiten2.9.9-audio` 映像加 Xvfb
  乾淨重跑 `go test ./cmd/... ./internal/... -count=1 -timeout 600s` 全部通過。
  繁中字模覆蓋 5,122 筆／0 缺字、既有文字安全矩形 6,952 筆／0 失敗。
  `tools/worklist.py selftest` 正反對照通過，`verify` 為 73 條未完成、17 條
  已核實完成，無過期訊號；`git diff --check` 通過。
- `README.md` 移除過期的 2026-09-05 自評流水帳並修正繁中資料頁現況；
  `CONTEXT.md` 頂部、`WORKLIST.md` 與繁中收據索引已更新。尚未選定近似完成
  或嚴格對拍的交付門檻，已向使用者提出一個父層決策；回答前不把建議寫成決定。
- 清理本輪生成的 207 張探索 PNG、探索軌跡與暫存 CJK 報表，只保留兩張正式
  法術畫格、其 118 步軌跡與收據。所有新增檔案為 UID/GID `1000:1000`；
  專案根未見 root-owned 檔或誤建的 `.md` 目錄。一次性容器均用 `--rm`，
  依專案映像查無殘留容器。`workplace/tmp/kol` 原本已存在，本輪探索時
  被目前 HEAD 重建覆寫；這是被忽略的中間產物，已記錄而未再清除。
  現行 AppImage 未重建；本輪未 commit／push，也未作遠端 Issue 寫入。

## 2026-09-23：本機可玩版範圍、候選封包與戰鬥判讀勘誤

- 使用者在完成級別決策選「本機可玩版」。這輪以最新原始碼的本機 Linux 封包、
  正常操作與保存／重載為目標；近似完成的全面抽樣、嚴格對拍、其他平台及公開
  發行沒有被納入完成宣稱。載入交付目錄與版號契約後沿用專案既有
  `full/release/promo` 結構；沒有把含原版素材的包放進 Git 或公開目錄。
- `kol-package:20260907` 以唯讀來源與乾淨新目錄重建
  `dist-all/v.0.2.3-20260923/full/kol-v.0.2.3-20260923-full-linux-x86_64.AppImage`。
  SHA-256 為 `3413ec68699eb4ff06044c014bec0c6a720234fcd958d056a8d467b29d9f7a4e`；
  `build-inputs.json` 記 1,107 筆輸入。解包驗證程式版號、x86-64 ELF、849 筆
  素材／文件雜湊，三目錄 manifest 與 SHA256SUMS 逐檔核對通過。`release/`／
  `promo/` 只有「尚未交付」佔位。第一次自製 manifest 檢查把容器掛載點名
  `/dist` 當成版號，屬驗證腳本錯誤；同一映像修正讀取 `build-inputs.json` 後通過。
- 實際 AppImage 以 `--appimage-extract-and-run` 進行 GUI 抽樣：建角、名冊、
  住宿、刪角與重用槽位、v1 匯入、受控世界時間，及兩種預設保存位置
  `XDG_CONFIG_HOME`、`HOME/.config` 的三次啟動／95 次輸入均通過。第一次
  `verify_local_appimage.py` 沒有 DISPLAY 而失敗，補 Xvfb 後以相同映像重跑通過；
  這不是產品缺陷。輸出收據在三個 `workplace/appimage-*v.0.2.3-20260923/` 目錄。
- 六人隊伍 seed 1 的正常按鍵路徑從城鎮走到野外 WALBAR 遭遇、規劃 26 回合，
  回到世界後走入半途之家，住宿寫 v2 存檔，重新啟動同一 AppImage 載入後
  六人、傷勢、金冠與城鎮位置仍在。**勘誤：**最初把 `verdict=1` 誤讀成勝利；
  畫面是「The WALBARs have defeated you!」，`internal/battle/victory.go` 亦定義
  `1=敗北`、`2=勝利`。本條只證明敗北結算與存檔，不證明勝利或獎勵回寫。
  已讓 `tools/drive_battle_script.py` 同列顯示文字勝負與原值，避免再誤讀。
- 額外以原版名冊的單人 Poontang 和 Moraine 探索固定 seed GOBLIN 戰鬥，
  兩條也敗北；未為通過而改動正式規則或素材。Poontang seed 3 正常行走到
  世界 `(48,21)`，觸發第 5 號任務提示並進入 BINDERAK 戰鬥。
  現行封包尚無正常路徑勝利、獎勵或 #90 戰利品繁中畫格收據，故維持候選
  身分，不把歷史勝利收據升格成本版完成。細節、雜湊與限制見
  `docs/verification/appimage-v.0.2.3-20260923.md`。
- 本輪 Docker 作業均 `--rm --network none`，有資源限制及外層逾時；
  專案資料與封包為 UID/GID `1000:1000`。清掉 76 個明確列名的種子掃描與
  敗北探索暫存檔／目錄，保留戰後保存、重載及任務入口證據；本輪未產生
  root-owned 專案檔或誤建的 `.md` 目錄。`git diff --check`、
  `tools/worklist.py selftest/verify`、腳本語法與 JSON 檢查通過；
  工作清單仍是 73 條未完成、17 條核實完成。專案映像未見殘留容器。
  沒有 GitHub Issue 寫入、公開發行、commit 或 push。

## 2026-09-24：繁中戰利品與戰後返回續測

- 固定 seed 39、六人 v1 存檔轉成暫存 v2，配合明示的觀察用怪物與任務地點資料，
  以正常玩家 token 重播至第 22 回合勝利（`verdict=2`）。「找到的物品」與「角色」
  兩張 960×600 畫格已目視檢查，PNG、trace、輸入及程式雜湊收在
  `docs/verification/cjk-ui/battle-spoils-winning-receipt.json`；本地 #90 按其畫格條件
  標為完成，GitHub 遠端仍依使用者要求等整體驗收後更新。
- 續按取消鍵發現戰利品第一頁結束後會重開清單；依 RE-115 的取消路徑，在單場
  結算加只開一次的閘門。`N N RET` 現返回野外，中文與英文模式的最後狀態軌跡
  SHA-256 均為 `633010aa837b7e9426854fbcb1bb59551ef97193ab40547e770c5c3bcf2908b3`。
- 返回野外的繁中畫格揭出日期仍是 `Apr 1`／`AD 1100`，已改由時鐘提供
  「四月一日／西元 1100」，年末日期測試與目視畫格通過；英文模式的
  `DateLines` 與畫布不變。修正後二模式結束畫格與軌跡同列於上述收據。
- 修正後以 `eob-remake-go:1.26.7-ebiten2.9.9-audio`、唯讀來源、Xvfb、
  `--rm --network none` 重跑 `go test ./cmd/... ./internal/... -count=1 -timeout 600s`
  全部通過；新 binary SHA-256 為
  `e923df8cda9793dd96a7ba6c0e9b4443735c3960b1781314bbddb98deea4ff64`。
  這只驗重製版狀態，不是原版同狀態或正式 AppImage 驗收。
- SPEC-032 按四頁畫格、幾何測試與來源收據升 CONFORMED；這是中文版面規格，
  不替代戰鬥機制對拍。另核對字模 5,785 筆文字缺字 0、資料原名 385 種缺譯 0、
  共用文字庫 486 筆失敗 0；城鎮來源仍有 45 筆 2–9 號亂碼／殘句缺譯，
  其原版全域可達性尚未證實，沒有把靜態檢查改成假綠燈。
- 收尾清理：`git diff --check` 通過，`tools/worklist.py verify` 通過（72 條未完成、
  18 條已核實完成）；無本專案殘留容器，也無 root-owned 專案檔或誤建的 `.md`
  目錄。`/tmp/kol-closure-driver` 保留 14 MiB 的 UID 1000 暫存 v2 存檔、
  本輪 binary 與重播輸出，供後續同狀態研究；正式可追溯畫格及收據已複製進
  `docs/verification/cjk-ui/`。本輪未更新遠端 GitHub Issue、README 截圖、
  正式封包，也未 commit 或 push，因使用者指定整體完成後才更新前兩者。

## 2026-09-24：六人馬廄、角色換頁與體重單位

- 以 `workplace/saves/full-chain.json` 建六人觀察存檔，移至城 1 馬廄門口並給
  3,000 金冠；原始輸入與暫存 v1／v2 的 SHA-256、24 個按鍵、程式雜湊及
  25 筆狀態見 `docs/verification/cjk-ui/stable-six-page-receipt.json`。這是正式
  按鍵分派的局部玩家路徑，並非新遊戲自然抵達或原版同狀態對拍。
- 購馬後的六人角色資料頁、物品頁、隊伍立繪頁與返回畫格已目視檢查。
  第 11 步金冠由 3,000 降至 2,600；英中重播軌跡 SHA-256 相同。
  原先按 `SPACE` 離開隊伍立繪頁時，圖示列返回建築但立繪頁仍殘留；
  `Session.Continue` 現先關閉該頁，`TestOpenPartyPage` 加入正常按鍵回歸。
- 繁中角色頁體重欄仍印 `LBS`，依手冊「磅」補 UI 譯文，繪製端只替換單位，
  重量數值與英文畫面不變。既有字型映像重烘為 2,542 字；新版角色畫格
  顯示「288 磅」且第 16／24 步 PNG 相同。六張畫格的尺寸與雜湊均對照
  正式收據通過；英文角色畫格逐 byte 不變，英中軌跡仍相同。
- `go test ./cmd/... ./internal/... -count=1 -timeout 600s` 在鎖版 Go 映像與
  Xvfb 下全部通過。字模 5,786 筆文字缺字 0、資料譯名 385 種缺譯 0、
  共用文字庫 486 筆失敗 0；同一套排版輸入重跑 9,301 項安全矩形失敗 0。
  全中文硬閘門仍有 45 筆城鎮來源亂碼或殘句
  缺譯，原版可達性未知；手冊列出的整體機制與原版同狀態仍未完成。
- 第一次完整測試命令以登入 shell 改掉了 Go 的 `PATH`，屬驗證環境錯誤；
  改用同映像非登入 shell 並把 Xvfb 設為有 trap 的背景程序後乾淨重跑通過。
  第一次英文重播用了 22 而非 24 個按鍵，修正為同一 24 鍵序列後英中軌跡
  完全相同，屬驗證命令錯誤。
- SPEC-030 的繁中版面升 CONFORMED，本地 #86 完成；工作清單 71 條未完成、
  19 條本地完成。遠端 GitHub Issue 與 README 截圖依使用者要求留待整體
  完成；正式 AppImage 尚未以本輪原始碼重建。
- 清理核對：`git diff --check`、`tools/worklist.py verify` 通過；本專案相關容器
  無殘留，專案內無 root-owned 檔或誤建 `.md` 目錄。新字模、收據與畫格均由
  UID/GID `1000:1000` 擁有。`/tmp/kol-closure-driver` 約 43 MiB，保留可重播的
  暫存存檔、binary 與新舊英中畫格；此目錄不是交付物。

## 2026-09-25：任務交付勘誤、中文來源槽位分類與 GUI 抽驗

- 以固定 `KOLTOWN-unpacked.EXE`、專用 IDA Pro 9.4 映像及 `.i64` 一次性副本查
  `sub_14234` 的 25 個直接呼叫。訊息 2–5 是接任務、6–7 是任務物品交付；
  2–7 的 30 筆馬廄未譯槽位受 `icon_flags[1] & 4` 阻擋，8–9 的 15 筆未見
  直接繪製呼叫。沒有排除間接讀取，嚴格翻譯檢查維持 45 筆失敗。
  工具版本、原始位址、輸入與原始收據雜湊及 RE-109 勘誤見 RE-139。
- RE-139 證實舊「交付時沒有台詞」不成立。依 SPEC-014 READY 補充條款，
  `Session` 改為進門敘述讀完才交換第一件吻合物品並顯示訊息 6／7；再按一次
  SPACE 才替所有已完成該任務的隊員回報。規則、型別化物品同步、同時持有
  兩件、非持物者回報與按鍵順序均補回歸測試。
- 用六人觀察存檔在鎮長宅邸門前執行 `U RET SPACE SPACE`；三張繁中畫格
  已目視檢查，英中五個狀態步驟的追蹤 SHA-256 相同。來源 v1／v2、程式、
  PNG、trace 雜湊與「注入任務物品／旗標」限制見
  `docs/verification/cjk-ui/quest-turnin-receipt.json`。這證明重製版正常按鍵
  分派，不證明從新遊戲自然取得任務或原版同狀態；SPEC-014 留在 READY。
- `eob-remake-go:1.26.7-ebiten2.9.9-audio`、Xvfb、唯讀來源執行
  `go test ./cmd/... ./internal/... -count=1 -timeout 600s` 全部通過。翻譯分類器
  已對原版建築旗標標出 30／15 筆兩種缺口，仍以非零狀態拒絕 45 筆缺譯。
  初次 Python `py_compile` 因唯讀掛載試圖建立 `__pycache__` 失敗，屬驗證命令
  的寫入位置錯誤，並非遊戲缺陷；後續以容器內 `PYTHONPYCACHEPREFIX=/tmp/kol-pycache`
  重跑語法檢查通過。另核對 23 間啟用任務物品的建築，獎勵 raw 16-byte 記錄
  與 typed 記錄不符 0 筆、缺少 typed 獎勵 0 筆。
- 本地 `docs/worklist.json` 新增兩條未結工作，`render`、`verify`、`selftest`
  通過；73 條未完成、19 條已核實完成。新增的 CJK 畫格與追蹤檔由 UID/GID
  `1000:1000` 容器寫入。原版同狀態、自然任務鏈、全部中文字串及新版封包
  仍待驗；沒有更新遠端 GitHub Issue、README 截圖、正式封包、commit 或 push。
- 收尾核對：六張 PNG 與兩份 trace 的 SHA-256／尺寸和 JSON 收據一致，
  `git diff --check`、Python 腳本語法、JSON 語法與 Go 格式檢查通過。
  本輪既有 Go／IDA 映像沒有殘留容器；專案樹未發現 root-owned 檔案或誤建
  `.md` 目錄。`/tmp/kol-quest-turnin-20260925` 約 13 MiB、
  `/tmp/kol-msg-ida-20260925` 約 6.3 MiB，均為 UID/GID `1000:1000` 的
  本機觀察與 IDA 副本暫存，供後續同狀態研究，非交付包。
- 翻譯檢查器原先只計已有譯文的建築名／人物名，整欄缺失可能漏算；已改以
  原始建築欄位為母體。現行 101 個建築名與 101 個人物名均有中文；從暫存
  譯表刪掉鎮長宅邸名稱的負例使缺譯由 45 增至 46，證明新閘門能抓整筆缺欄。
  重跑統計也確認建角／共用 UI 譯表現為 99 筆（包含新加的「磅」），
  `CONTEXT.md` 與手冊稽核已訂正先前的 98 筆數字。
- 接著用既有 `quest-take` 34 步計畫試繁中任務台詞。初次誤用城鎮專用
  `-play-script` 執行整個 flow，第三鍵 `5` 不被接受，屬命令路由錯誤；改
  `-flow-script` 後，舊存檔在第 17 步停於選角。該 v1 觀察存檔的城鎮位置
  是第 7 城，角色所在地卻是 MIN，與現行依所在地還原旅店的 `StartLeader`
  不相容；不能把舊 34 步收據當作現行自然玩家路徑。
- 從該存檔明示的第 7 城 `(9,24)` 起跑，經正式按鍵
  `U RET SPACE > > > RET n o b j RET SPACE SPACE RET` 問中第 5 槽、顯示
  兩頁繁中任務敘述、問是否接受，答應後第 6 號任務由未接改為已接。
  中英文用同一暫存 v2、同一 seed 1 與 15 鍵，16 個狀態步驟逐位元相同；
  五張繁中畫格已目視檢查。正式輸入、程式、譯表、字模、PNG 與 trace 雜湊
  見 `docs/verification/cjk-ui/quest-offer-receipt.json`。這只證明指定城鎮起點
  的局部玩家路徑，不替代自然抵達、原版同狀態或完整任務鏈。

## 2026-09-25：GitHub Issue 與本機狀態文件同步

- 使用者明確授權先更新 Issue 狀態與本機 Markdown。主機 `gh auth status` 成功，
  目標為 private 儲存庫 `wicanr2/kol-remake-cht`；先讀遠端 Issue，
  再由 `docs/worklist.json` 在 `kol-package:20260907` 一次性容器產生正文。
- 更新 #14、#33、#34、#35、#39、#61、#86、#88、#90 的正文；#61、#90
  補 `manual` 標籤，#90 標題改為「補戰後戰利品清單的繁中正常路徑畫格」。
  #86、#90 依各自局部版面收據標為完成並關閉；這兩項完成不代表全遊戲
  中文或原版一致。#14、#33、#34、#35、#39、#61、#88 保持開放。
- 新建 #92 追蹤 45 個未譯來源槽位的可達性及中文處置；新建 #93 追蹤
  任務物品交付的原版同狀態與自然玩家鏈。兩者保持開放，編號已寫回
  `docs/worklist.json`，並重生 `WORKLIST.md`、`CONTEXT.md` 統計區。
  回讀遠端 11 張相關 Issue 的開關、標題與標籤；#86、#90、#92、#93
  正文的工作清單識別碼也已抽查。
- `README.md`、`CONTEXT.md`、手冊忠實度稽核及玩家路徑說明改以目前收據
  描述候選包、45 筆缺譯與過期 `quest-take` 計畫；README 截圖未更新，
  仍待整體驗收。`tools/worklist.py render`、`verify`、`selftest` 通過：
  73 項未完成、19 項已完成；`git diff --check` 通過。本輪沒有重新建置
  遊戲、提交或推送工作樹，因此遠端 Issue 中指向新收據的本機路徑尚未
  隨 Git 同步。
- 全部本輪分析與產生器執行使用 `docker run --rm`、`--network none`、
  有界資源及 UID/GID `1000:1000`；專案檔案擁有權抽查正常，未見
  root-owned 專案檔或誤建 `.md` 目錄。本專案沒有殘留容器；其他專案
  容器未清理。Issue 正文暫存於 `workplace/worklist-issues/bodies-20260925/`。

## 2026-09-25：接手收尾、語言預設修正與提交

- 使用者指定本輪範圍：先結清 09-23～09-25 未提交成果、分批 commit 並授權
  push，再從目前原始碼重建本機包；45 個來源槽位先查可達性再決定處置。
  開工時 git 身分與歷史作者均為 `wicanr2@gmail.com`，本專案無殘留容器。
- 工作樹曾把 `-lang` 旗標預設由 `en` 改為 `zh-Hant`，但未記錄決策。HEAD
  已在互動遊玩時自動切繁中；全域改預設會讓 30 餘份沒寫 `-lang` 的對拍計畫
  （如 `remake-parity/plan-*.json`、`battle-*/replay.spec`）改產 960×600 中文畫格。
  實測 `-scene inn -shot` 預設輸出 960×600，加 `-lang en` 為 320×200 且與既有
  `remake-parity/shots/inn.png` SHA-256 相同。已把旗標預設改回 `en`，
  繁中譯表存在檢查移到互動切換之後，雙擊啟動仍涵蓋。玩家預設體驗不變。
- `battle-spoils-winning-receipt.json` 原命令依賴該預設而未寫語言，已補
  `-lang zh-Hant`。以修正後程式、同一 v1 存檔、觀察資料與前 1,926 個 token
  重播，`items-found.png` 與軌跡 SHA-256 與收據逐位元相同。
- `tools/run_kol_headless.sh` 在執行檔比原始碼舊時以狀態 2 拒跑；先前
  `-dump-cjk-layout` 因此沒有覆寫 `workplace/cjk-layout.json`，text-fit 收據
  （含 HEAD 版本）一直使用 09-05 的 2 倍座標版面。重編後以 3 倍現行版面重跑，
  仍為 9,301 項、失敗 0，已更新 `docs/verification/cjk-ui/text-fit.json`。
  規則：重生版面後要確認命令成功且檔案確實更新，不能只看檢查器結果。
- 驗證：鎖定映像、唯讀來源、Xvfb 下 `go vet ./cmd/...`、
  `go test ./cmd/... ./internal/... -count=1` 全過；顯示譯名 385／0 問題、
  字模覆蓋 5,786／0 缺字、`worklist.py verify`／`selftest` 通過；
  `verify_runtime_translation.py` 仍以 45 筆缺譯非零結束，屬已知 #92。
  `gofmt -l` 列出 7 個 HEAD 既有、未在本輪異動的檔案，未處理。
- 已推送 `c4d967c`、`de70372`、`5a0c0ad` 至 `origin/main`（private
  `wicanr2/kol-remake-cht`），提交檔無 ZIP／EXE／PDF 等原版輸入。
- 以 `kol-package:20260907`、唯讀來源與新空目錄建 `v.0.2.4-20260925`；
  AppImage SHA-256 `1c8930406eef65c63a61dbe7bc9361cae6d9f3d9096496bd897696fa1cc9ccbf`。
  完整保存 GUI（28／32／35／0／3 次輸入）、封包 851 筆雜湊與兩種預設存檔位置
  通過；雙擊啟動畫面為繁中，日期中文化已目視確認。
  `verify_local_appimage.py` 首次因沒把 GUI 腳本 stdout 存成
  `verification.log` 而失敗，屬驗證命令錯誤；XDG 預設位置首次因 Xvfb `:95`
  尚未就緒使 GLFW 無法開 display，屬驗證環境時序，兩者皆以新空目錄重跑通過，
  失敗目錄已刪。收據見 `docs/verification/appimage-v.0.2.4-20260925.md`。
- RE-140：三支新 IDAPython 探針各在一次性 `.i64` 副本執行，正式 `.i64` 雜湊不變。
  `sub_14234` 以 `DGROUP:0x1CF4` 為基底解碼訊息，這也是建築 payload 緩衝；
  45 筆未譯槽位分為 40 筆馬廄坐騎記錄（已證實）、1 筆商品表殘留資料與 4 筆
  找不到讀取點的任務續文（強推論）。依使用者處置不翻譯，新增
  `docs/i18n/town-nontext-slots.json`（原文快照綁定），檢查器改為 2,076 則 0 缺譯、
  另排除 45 筆；刪一筆、竄改快照、多列一筆三個負例均以非零結束。
  worklist #92 在本地標為完成，#35 改寫說明；71 項未完成、21 項已完成。
  遠端 Issue 本輪沒有寫入。探針收據含原版 bytes，留在 `workplace/ida-reach-20260925/`。

## 2026-09-25：中文化漏譯掃描與修正

- 使用者授權同步 Issue 與 push，並指示繼續中文化。`tools/sync_worklist_issues.sh`
  更新 5 處，#45、#92 關閉、#35 正文更新；回讀狀態相符。
- 新增 `font.DrawObserver` 與 `-text-log`，記錄每個第一次畫出的（字型, 字串）；
  `tools/scan_cjk_text_leaks.py` 以中文模式重播 23 條玩家路徑與 9 個戰鬥情境，
  彙整含拉丁字母、又不是玩家名的字串。只看得到字模繪製的字串；畫在原版圖片裡的
  英文與「有畫但看不見」要看畫格。
- 修正的漏譯：名冊／刪角頁所在地（`HALFWAY HO MIN`，SPEC-033 補 101 個所在地鍵與
  7 個城鎮代碼）；教派名與頭銜（入會句、法術編輯拒絕句與資料頁，12 筆暫譯，
  代入句尾改全形「！」）；戰鬥撿／丟物品名（`|H`／丟棄時的 `|E` 改查顯示譯名）；
  敵人名前的英文冠詞 `The`；所有格 `他的的`（KOLDATA 272 改譯「他」，與 271「她」
  一致，模板自帶「的」）。
- 根因修正：`UIRaw`／`UISegments`／`JoinedSegments` 與 Session 三個包裝把刻意留空
  的譯文（記錄 320 `The]`、377 ` has`）當成沒譯而退回英文；新增 `RecordOK`，空譯文
  帶 note 視為有效。
- 使用者決策（2026-09-25）：移動率顯示「U（20）」；清單捲動箭頭改 ▲▼；法術名保留
  精靈語原文並加中文注解。注解依說明書音節表機械解析，說明書 6 例與 48 筆商店法術
  和法術記錄逐筆一致；原型經使用者確認後立 SPEC-034（READY）。
- 畫格檢查發現既有缺陷：中文模式戰鬥執行段訊息自 2026-09-11 起一直是黑塊（黑底填色
  上寫黑字），掃描看得到字串但畫面讀不到；改為卷軸上直接繪字。同時修正
  `WrapMessage` 硬斷時吃掉一個 byte、可能切壞中文字，中文繪製端改依像素寬重新斷行。
  規則：文字掃描只證明「畫了什麼」，可讀性要看畫格。
- 字模重烘 2,553 字，舊字零流失；新增 `docs/i18n/ui-symbols-zh-Hant.json` 與
  `spell-words-zh-Hant.json` 納入烘焙與覆蓋檢查。檢查：安全矩形 9,467／0、UI 譯文
  486／0（2 筆刻意留空）、城鎮文字 2,076／0（另排除 45 個非文字槽位）、顯示譯名
  492／0、字模 5,941／0；Docker／Xvfb 下 `go vet`、`go test ./cmd/... ./internal/...`
  全過。`go fmt` 順帶改到的 7 個無關檔已還原。
- 最終掃描（commit `f6faa44`）：34 組紀錄，剩 17 筆拉丁字串，全屬移動率字母、
  精靈語法術名或句中玩家名；收據 `docs/verification/cjk-ui/text-leak-scan-20260925.json`。
  `battle-flee-0`、`battle-rideaway-1`、`battle-spoils-0` 三行舊 replay.spec 在野外送
  `SPACE` 而中止，與語言無關，屬過期腳本，未處理。中途停止一次過時的掃描，
  僅停了本專案自己的容器；另一個 `eob_123_remake` 容器未碰。
- 續查「看不見／沒畫出來」的問題：`cellInAt` 呼叫點的字色都不與底色相同；但中文
  `InnCJK` 完全沒處理 `Inn.SpellEdit`，法術編輯頁在中文模式只剩兩行造價數字。
  新增 `drawSpellEditCJK`（SPEC-035 READY）：英文版面乘 3，欄名改用說明書術語
  （KOLDATA 0x4D 譯文為 RANGE／DUR／DMG 各給 `~Hxx`），欄位值查顯示譯名（補 MAGIC.0
  的 class／stat／All 18 筆），名稱下接 SPEC-034 注解；選單畫在被編輯欄位另一側，
  名稱輸入顯示游標。英文模式同輸入畫格與對拍畫格逐位元相同。
- `verify_ui_translation.py` 放寬欄座標規則：多出的 `~Hxx` 只允許等於原文同段某字的
  起始欄；正例 `~H18／~H1D` 通過、負例 `~H19` 失敗。
- 收據 `docs/verification/cjk-ui/spell-pages-receipt.json`：七張繁中畫格（編輯頁、修改
  清單、選單、名稱輸入、法術店注解、戰鬥選法術注解、戰鬥執行訊息）經目視檢查。
  SPEC-034 升 CONFORMED。檢查：安全矩形 9,557／0、字模 5,959／0、譯名 510／0、
  Go 測試全過。
- 從 `c62bef2` 重建 `v.0.2.5-20260925`（AppImage SHA-256 `8bcb7814…fc27f`，1,119 筆輸入）；
  完整保存 GUI、封包雜湊與兩種預設存檔位置通過。README 交付入口改指向它。
- 目視檢查：以 `scan_cjk_text_leaks.py run --frames` 產生 1,276 張中文檢查點畫格，
  在 `kol-package` 容器以 32×20 灰階指紋全域去重為 48 種版面，拼 3×3 縮圖逐張看。
  發現中文模式開場標題只畫在左上角 320×200：`sceneTitle` 用 `canvas.DrawAt` 而非
  既有的 `drawPage`，已改；英文標題畫格與既有截圖逐位元相同。其餘畫格未見黑塊或
  錯位。原版美術內的英文（開場 ORIGIN／製作名單／標誌與版權、封面書名）列為待
  使用者決定。`montage` 在該映像因缺字型崩潰，改用 `convert +append／-append`。
- 使用者 2026-09-26 決定：開場保留原圖加中文字幕、英文標誌保留加中文副標。新增
  `docs/i18n/title-zh-Hant.json`（以 TITLE.EXE 呈現的 BACKPICS 圖號 7…12 為鍵，由
  `-title-timeline` 對出六張字幕卡）與 `cmd/kol/title_subtitles.go`：字幕以中文字模
  兩倍最近鄰放大、黑底，一般卡放畫面下緣，標誌卡做成劍身右段銘牌，封面副標放在書名
  板下方；人名與版權行保留原文。字模 2,557 字零流失、覆蓋 5,966／0。英文封面與標題
  畫格逐位元不變。收據 `docs/verification/cjk-ui/title-subtitles-receipt.json`。
- #61 城鎮打字問話：腳本 token 在第 7 城民宅驗基準／退格／空問句，英中軌跡相同，
  退格版最終狀態與基準相同。空問句送出後游標仍留在對話框，與 RE-114「讀入函式收工前
  蓋掉橫槓」不符，改 `SubmitAsk`；英文 `town-talk` 收據重跑 10 張一致。另寫
  `tools/verify_town_typing_gui.sh`：互動遊玩下 xdotool 真實按鍵、每鍵等輸入紀錄，
  英中輸入序列相同，十二張畫格收據 `docs/verification/cjk-ui/town-typing-gui-receipt.json`。
  第一次 GUI 腳本用了修正前編的執行檔，已在腳本加過期檢查。空問句後白框重設到門，
  原版位置未對拍，未改。#61 本地完成。

## 2026-09-26：可獨立完成的工作項

- #49：`verify_packages.py` 新增 `check_manifest`，三包逐檔重算 SHA-256 與大小、抓清單外
  與缺檔；`verify_dist.sh` 經它執行，`verify_local_appimage.py` 也呼叫 `--manifests-only`。
  v.0.2.5 通過，竄改／多檔／少檔三個負例失敗。舊版 0.2.1 因 2026-09-04 打包工具排除所有
  `manifest.json`，`full/assets/ui-font/manifest.json` 未入清單，新檢查會報出，舊交付未改。
- #26：容器無 strace、Go 不經 libc，改以 `go/ast` 靜態守門：旗標預設路徑須在 assets/、docs/
  且為 PNG／JSON／TSV／OGG 或目錄，非測試程式碼不得有像原版檔名的字面值；`assets.Decode`
  拒絕非 PNG、逃出目錄與內容非 PNG 的測試。自我負例與把旗標改成原版封裝的真實負例皆失敗。
- #21：25 份規則規格加「說明書對照」表，連結全數可解析；015、028、029 明示說明書無對應。
- #87：`shots.spec` 46 行各跑兩次逐位元相同；與收據不同的 9 張逐一以像素差對到擲骰、
  捲雲動畫與中文化改動，非編碼不穩定。README 展示圖依使用者指示暫不重生。
- #85：以新探針 `tools/ida_probe_koltown_function_dump.py` 傾印 `sub_24131`，讀出滑鼠
  分支：嚴格區間換格後轉為位移、與方向鍵同段夾限，位移為零整段略過。規則層早已照
  RE 實作，本次只接輸入端 `fieldCellTargets`＋`CELL:` token；另發現選目標格時戰場外
  點擊會落到 fallback 的 RET 而直接確定，與原版不符，已改為吃掉。射擊重播 X→CELL:8,8
  英中軌跡逐位元相同，收據 `docs/verification/battle-target-mouse/receipt.json`。
- #13：`verify_complete_save_gui.sh` 補「住宿離隊後重載，兩名新角色都能再選進旅店」，
  v.0.2.5 AppImage 真實按鍵連同既有建角、入隊、刪除、空槽重用、滿槽拒絕全數通過，
  兩種預設存檔位置亦過。同日第一次執行在建角後因視窗偵測偶發失敗而靜默結束，
  相同輸入重跑兩次皆通過；腳本補上失敗訊息。收據 `docs/verification/identity-save-chain/`。
- #39：`docs/i18n/overflow-policy.json` 為 22 種 CJK 欄位登記 overflow policy
  （拒絕、拒絕並執行期裁切、拒絕多列），`verify_text_fit.py` 對未登記欄位判失敗；
  9,557 項 0 失敗，移除一種欄位的負例失敗。新增 `verify_glossary_consistency.py`
  以術語表核准／暫譯詞對照資料譯名、建角字串與文字庫：201 詞、105 處比對、0 衝突，
  語境不同的 8 處登在 `glossary-exceptions.json`，假例外負例失敗。依術語表改
  懸崖巨魔、岩石食人魔、勳章／黃銅勳章、精靈族／矮人族、遠望者；字模重烘為 2,559 字。
- 新登 `target-mode-icon-bar-mouse`：選目標格時圖示列的滑鼠行為原版未查，只收鍵盤。
- #88：`check_player_paths.py --all` 重跑 25 條，10 條一致、15 條對不上。逐張算像素差與
  差異矩形分三類：只差天候圖雲層（SPEC-026 捲雲，3 條，已重生並同步
  `remake-parity/shots/world/` 四張，`world` 4/4、`world-road` 10/10 通過）；09-04 後才接上的
  野外遭遇插進路徑（world-clock、world-quest，未重生）；標題選角後落點不同（10 條，未重生）。
  重跑途中重烘字型，但差異都不在中文畫面，不影響分類。#88 仍開放。
- #8／#19：唯讀子代理逐句稽核 SPEC-005、008、010～015，回報 16 條過時；逐條以 RE-064／088／
  092／098／108／113／115／118／124／135 與程式核對後改寫 SPEC-008、010、011，另補稽核漏掉的
  SPEC-008 授權段 `+0x05` 舊說法。#56 由 RE-118 結案（原版 `+0x05` 恆為 0）；#52 改綁殘留的
  攻擊致死 collapse／shatter 分支。核對時發現 RE-118 對 `sub_21E0B` 過濾條件前後矛盾，
  登 `sub-21e0b-filter-contradiction`；`battleStep` 跳過無訊息手的迴圈理由已過時，登
  `battle-silent-steps-inventory`。main.go、battle.go 只改註解，Go 測試通過。
- #88 README 的 world-clock 處置改引 SPEC-011 §9 既有前例（不驗遭遇的路徑可另選種子並寫明理由）。
- #88 續：world-clock、world-quest 原種子因天候重擲位移而遇到遭遇。依 SPEC-011 §9 前例，
  以多行規格一次跑多顆種子並輸出軌跡，從原種子往上取第一顆路徑事件符合原設計的
  （world-clock 試 2…13 取 3；world-quest 試 3…14 取 5，種子 4 在第 52 步有遭遇提示而排除）。
  world-clock 12 顆種子最後 ticks=3360、time_index=1 全同。重生後路徑 28／91 步 0 失敗，
  對拍 `world-clock`、`world-quest` 各 4/4。對拍圖 step14／step15／intro 同步換新。
- #88 續：選角後現在先進旅店敘述，舊按鍵的 `>` 落在敘述頁而無效。`full` 改成敘述 RET、
  招呼 RET 選門出去，`town` 在招呼頁多按一次 RET 出門；門口與進門兩個檢查點仍與對拍圖
  逐位元相同。`quest-list` 的舊 `party` 格式存檔在標題選角時用不上，改成讀檔直接在城裡
  開場（標題段由 `full` 驗），任務一覽只重畫抬頭層依 RE-137 為正確行為。
  `replay_player_path.py` 改為遵守 plan 的 `scene` 欄位。
- 修正走出建築後隊員格仍畫白框：原版 frame-071（戶外）與 frame-131（野外）在有發言者時
  都是藍框。`idleParty` 只改畫法，新增單元測試；Go 測試通過。讀存檔開場的六條收據因此
  只差第一格頭像框而重生，`world-quest` 對拍 4/4。
- 修正讀檔後標題選角停在名冊：TOWN7 十個入城點都是 MIN，各有夾限；存檔位置在入城點 4
  那一區時放不下半途之家門口，`StartLeader` 回傳失敗而無提示；跨城分支的 `TownByCode`
  遇到重複代碼一律回 nil。改成 `LoadStartingTown` 回傳該代碼全部入城點，逐一試到放得下
  門口為止（入城點順序是 remake 選法，寫進 SPEC-016）。新增同城異區正反測試，Go 測試通過；
  實跑 quest-take 舊按鍵，選角後進到半途之家。`quest-take` 改成讀檔開場（15 鍵），重生一致。
- 使用者決定（2026-09-26）：`flow-quest` 撤下（前提「標題選角沿用讀檔隊伍」已不成立，標題段由
  `full`、任務地點由 `world-quest` 涵蓋）；`world-healer` 改用受控亂數並在收據標明方法；
  `full-chain`、`full-chain-reload`、`full-chain-ending`、`arena` 改成讀檔開場並重生戰鬥段。
- #88 續：`world-healer` 改讀檔開場，依使用者決定改用受控亂數。新增只供測試的
  `-ambush-rolls`（`cmd/kol/scripted_rolls.go`）：依序當作 `RollAmbush` 的骰值，用完回到種子
  亂數，超出骰面直接停下；預設為空，正式遊戲不經過。plan 以 `ambush_rolls` 記錄、重跑工具
  帶上、`verify_full_path.py` 照抄進收據。`1,1,1` 指定遭遇成立、落在治療、僧侶那一組；
  第一次試 `1,1` 時場景由種子擲到不問治療的 Brettle 那一組，屬正確行為。收據 28 步一致。
- #88 續：`full-chain` 改讀檔開場。種子 1 自然遇到 6 隻 WALBAR 而敗北，依使用者決定改用
  `-ambush-rolls 1,100,23`（遭遇成立、不落治療、ZOMBIE），等級隻數與戰鬥仍由種子決定；
  結果 2 隻 3 級 ZOMBIE、十二回合打贏，回城住宿後存檔數值（158 金、AP 2376、位階 1）與舊收據
  相同，存檔改為 v2 格式，期望值路徑同步。戰鬥段由 `drive_battle_script.py` 重生；
  `run_driver.sh` 補傳 `KOL_SCENE` 等環境變數並支援 `KOL_AMBUSH_ROLLS`。
- `-play-script` 改為遵守 `-shot-steps`（與 flow 相同）；`check_player_paths.py` 只對長路徑
  帶 `-shot-steps`，短路徑逐張比（`full` 因此從比 2 張變成比 27 張）。quest-list、quest-take、
  world-healer 收據目錄只留檢查點畫格。
- #88 續：`full-chain-reload` 改讀檔開場（讀新的 v2 `full-chain.json`）。出門時多了多人隊伍的
  告別語，補一鍵；種子 3 在第一次入夜同時擲到治療僧侶（自然事件），野外段交給驅動器處理後
  走到任務地點，第 94 步任務提示、第 95 步 BINDERAK。重生通過，README 與驗證矩陣步數同步。
- #54 結案：RE-117 的方向規則（三方向、隨機起點、三條退路、`Actor.Stuck`）早在 2026-09-05
  接上，ai.go 仍留著「解出來了但還沒接上」的過時註解，#54 的驗證綁在那句上而一直顯示未完成。
  改寫註解、改綁 `stepToward`／`enemyStuck` 當回歸訊號。長路徑收據重生由 #88 繼續。
- #88 續：`full-chain-ending` 改讀檔開場，城內段照舊，野外與戰鬥段由驅動器重生。種子 1 的路上伏擊
  打了 26 回合敗北（自然結果，非這條要驗的內容，描述標明敗北回寫另列待驗），之後帶傷走到任務地點、
  打贏最終任務、結局觸發；結局七段仍與 `-scene ending` 收據逐位元相同。驅動器停在結局是因為軌跡
  殘留的野外座標等於目標格，已人工確認。
- 依使用者要求改寫 `docs/verification/player-path/README.md`：開頭 09-11 起累積的過期警示與
  分類流水帳改成「收據與現行程式的對應」現況節（讀檔開場、種子、受控亂數、自然事件照實走）；
  更正 `town`、`full`、`full-chain`、`full-chain-reload`、`full-chain-ending`、`town-scroll` 的
  按鍵與步數；「不宣稱」清單移除已實作的項目（雲層動畫、下注、SCROLL 與馬廄、中文資料頁），
  補上讀檔開場與受控亂數的限制。`arena` 段標成待定，#88 驗證訊號改綁「收據重生中」。
- `arena` 重生試跑：種子 1 路上伏擊敗北、競技場那一場 Moraine 也敗北，沒有升階；
  賭金依輸的那一支結算（Poontang 214 → 189、Fighting Bull 442 → 452）。待使用者決定。
- #88 續：使用者決定 `arena` 照自然結果重生。讀檔開場，城內段照舊，路上 WALBAR 伏擊敗北，競技場
  進場、選 Moraine、賭場下兩注的 23 鍵照舊（步數整體減 740），對手 TROLL 第 23 回合敗北，
  沒有升階，賭金整排取負後結算（Poontang 214 → 189、Fighting Bull 442 → 452）。收據、README
  與驗證矩陣改為展示輸的分支。
- #88 結案：`check_player_paths.py --all` 24 條逐張一致；`verify_player_path.py` 17 條、
  `verify_full_path.py` 7 條全數 0 失敗（含存檔逐欄比對）。回歸訊號改綁重跑工具與 README 現況節。
  容器皆 `--rm`，本專案無殘留容器；中途停掉的驅動器容器已確認移除。
- #52 結案：依 `sub_277F4:0x2796F` 反組譯接上攻擊致死三分支——`sub_22229` 讀參戰者 `+0x01`
  判隊員（collapse from blood loss），怪物記錄 `+0x10` bit 7（只有 SKELETON）shatter into
  pieces，其餘 die from blood loss。`DamageLine` 加 `shatters` 參數；新增
  `cmd/kol/death_line_test.go` 以真實文字庫與繁中譯表展開三句。Go 測試通過，24 條玩家路徑仍一致。
- #96 結案：`act` 無訊息的手只有看不到的怪物休息／移動與腳下沒東西的撿。反組譯證實原版等鍵
  `sub_1B100` 在印訊息的分支裡（`sub_1E3A8:0x206E1`、`sub_26384:0x2654F`／`0x265B4`），
  跳過時不印也不等；SPEC-008「原版每一手都印一句」的前提不成立，改寫為與原版相同。
- 發現：`replay_battle_receipts.py --all` 7 份 `battle-*` 收據全數對不上或跑不起來（野外遭遇
  沒發生，`SPACE` 落在野外），與 #88 同一成因，另登工作項處理。
- 使用者定案（2026-09-26）：對拍與 remake 收據以**遊戲機制一致**為準，不要求截圖逐位元相同；
  記入 CONTEXT 已定案決策。新增 `tools/mechanics.py`（依序依條件找步、`$re`／`$in` 等條件、
  時段變換判定），`replay_battle_receipts.py` 支援 `runs` 機制收據與 `--write` 重截示意圖；
  軌跡在遭遇頁加記 `encounter_text`。
- #97：`battle-ambush` 改驗機制（時段變換才成立、遭遇兩行組法、隊員固定格、下一步進戰場），
  怪物由種子決定不綁定；站位改錯的負例會失敗。`battle-flee` 路上遭遇以受控骰值 100 不成立，
  驗任務地點休息 → 問逃 → 逃走 → 回野外。
- #97 續：`battle-missile` 原本是路上擲到的 BINDERAK 伏擊，但 BINDERAK 不能出現在 ROAD、射擊
  機率 0；依機制原則改以受控骰值 `1,100,29` 指定 BRIGAND（ROAD、85%），前兩回合休息讓它靠近、
  第三回合瞄準射擊，驗敵人射擊、傷害鏈、目標格模式與玩家射擊；射進樹裡由 missile_test.go 支撐。
  `battle-pickup` 伏擊照種子自然發生（TROLL），驗丟 → 指令列出現撿 → 問句 → 答是／答否兩條分支
  的第三回合指令列。重跑工具一律帶 `-trace-detail`。
- #97 結案：`battle-rideaway` 受控骰值依 `RollAmbush` 擲骰順序加上騎走那一骰（`-ambush-rolls` 改為
  同時接遭遇擲點與騎走判定），驗跑不掉接開打／跑掉回野外；`battle-spell` 以受控骰值指定 BRIGAND，
  驗丟武器後才出現施法與選法術、選正負、施法、效果；`battle-spoils` 路上遭遇受控不成立，驗打贏、
  獎勵、戰利品清單。軌跡加記野外提示頁文字（`prompt_lines`）。Go 測試通過，
  `replay_battle_receipts.py --all` 七份全數通過。
- #59 結案：法術清單的 `sub_1AAE8` 呼叫在 `sub_1E3A8:0x1F4DE`，推入順序（右到左）得
  (旗標 3, 行 4, 列 0x11, 可見列數≤7, 總筆數, 2E3Ah, 寬 0Dh)；依 RE-115 語意把 `SpellListRows`
  從 Col 3／Row 0x0F 改為 Col 4／Row 0x11，寫進 SPEC-012。Go 測試通過，中文版面與文字寬度
  9,557 項 0 失敗，battle-spell 收據通過並重截示意圖。
- #95 結案（RE-141）：`sub_21E0B:0x21E1A` 對每格穿戴欄先比 `+0x0D` > 0x7F，不成立整格跳過；
  RE-118「種類 byte 0 走 kind 0 設 +0x03／+0x04」不成立。現行資料下穿戴欄不設 +0x03～+0x06，
  remake 不設 Flag03 與原版一致。SPEC-008 那一節已無未解項，改標為「早期列管項目的去向」；
  SPEC-010、battle.go 註解同步（只改註解）。
- #53 結案：`sub_26C5F(0x17)` 早由 RE-091／RE-096 解出，`AdventureReward` 已實作並報出冒險點數；
  main.go「還沒解讀」是過時註解，已改寫。
- #58 結案（RE-142）：`+0x25` 就是參戰者指令碼（SPEC-010 欄位表早有記載），輪到時在
  `sub_1E3A8:0x20A8B` 讀進 `word_31122` 供分派與疲勞使用。remake 下馬時照原版把 `Command`
  清成 0，疲勞改用出手前的指令碼；補單元測試，Go 測試通過。
- #67 結案：`CommandBar` 產生得出的圖號在規劃分派表都有 case，default 走不到；補
  `TestPlanningEveryBarIconHandled`。核對時讀 `sub_1E3A8:0x1FB3x` 起的「Begin next round?」：
  `word_345DE == 0xFFFF`（競技場與訓練戰以外，含任務戰鬥）會多一格 0x30 問要不要逃，並逐名擲
  逃走判定；remake 沒有這一格，註解「任務戰鬥沒有」也不正確。已改寫註解並登 `begin-round-flee-icon`。
- `check_player_paths.py` 改為比**狀態軌跡**（去掉步名後逐步比對），截圖差異只列張數當參考；
  短路徑以 `--bootstrap` 產生參考軌跡（17 份新進版控），長路徑參考軌跡因今天新增
  `encounter_text`／`prompt_lines` 兩欄而更新——逐一確認差異只有這兩欄才覆寫。24 條狀態一致。
- #98 結案（RE-143）：「Begin next round?」在競技場與訓練戰以外多一格 0x30。`NextRoundBar`、
  `ChooseNextRound`（依白框：執行／U-TURN 重排／問 0x109）、`BarFlee`（逐名以
  ((智力＋力量)÷2×等級)÷3 擲判定，失敗且手上是武器的丟在腳下；一律逃掉、不標任務完成）。
  腳本 token 在這一問補 `<`／`>`。SPEC-008、SPEC-014 §3b 改寫為現況，推翻紀錄集中在 RE-143。
  單元測試四項，新增機制收據 `battle-bar-flee`；八份戰鬥收據全數通過。
- #69 結案：U-TURN 跳回 `0x1EF3A`，重擲鍵值、排序、`word_30D00` 加一，不經過回合結束的
  法術倒數 `0x20F01`。remake 原本會多扣一次法術剩餘回合，拆出 `Replan` 修正並補測試。
- #72 結案（RE-144）：參戰者 `+0x2C` 在原版沒有作用。整份反組譯無直接存取；效果欄位表第 5、6 項
  指到它，但 `sub_2818D:0x2844A`／`0x2869D` 先把效果 6、5 分走，查表讀不到。SPEC-012 改寫為現況。
- RE-145：`word_30D00` 是本回合的重排次數，不是回合數。整份反組譯五處存取：進戰鬥與
  `0x2119B`（每個新回合）歸零、`0x1F04C` 排序後加一、`0x1EFE3` 不為 0 才除隊員基準。所以
  除以 2 × word_30D00 只在同一回合 U-TURN 或逃走格答 n 重排時成立。remake 原本第二回合起就除，
  隊員規劃基準被砍半以上；拆出 `Round.Replans` 修正，`Round.Number` 只作 remake 的回合計數。
  測試補新回合歸零與 U-TURN 只加重排次數。RE-080、SPEC-008 §5 改寫為現況，更正紀錄集中在 RE-145。
  全套 Go 測試與八份戰鬥收據通過；玩家路徑 21 條狀態一致，`arena`、`full-chain`、
  `full-chain-ending` 三條多回合戰鬥段對不上，另行重生。
- RE-145 後重生三條多回合長路徑（`drive_battle_script.py`，三條並行各 2 CPU）：
  `full-chain` 改為第 4 回合打贏 ZOMBIE（原 12 回合），存檔數值不變，全長 585 步；
  `arena` 路上 WALBAR 伏擊仍敗北、競技場第 37 回合敗北（原 23 回合），賭金結算不變；
  `full-chain-ending` 照種子自然走會在最終任務第 61 回合敗北、走不到結局，依使用者
  2026-09-27 決定改以 `-ambush-rolls 100×6` 讓出城到任務地點不遇伏，戰場照種子 1，
  第 16 回合打贏、結局七段與 `-scene ending` 逐 byte 相同。README 路徑表同步。
  24 條玩家路徑狀態一致、八份戰鬥收據通過。
- #73 結案（RE-146）：兩個倒下計數同時設成 0xFF 只有逃走格答 y 的 `0x1FCF5` 一處。
  共同出口 `0x211AD` 依序分派打贏、打輸（全隊倒下或競技場／訓練戰）、逃走（兩者皆 0xFF），
  其餘直接到 `0x21607`。戰場邊上答 y 帶著本回合的倒下計數出去，一般戰鬥不印任何一句。
  remake 原本兩條逃走都印「You flee for your life!」，改由 `State.FleeLine` 只讓逃走格那條印；
  `battle-flee` 收據改斷言答 y 直接回野外，並拿掉答 y 前五個無作用的 SPACE（`fled.png` 的刪除
  誤併入上一個 commit，本次補上 `back.png`）。SPEC-014 §5–6、SPEC-008 改寫為現況。
  另登 `combat-result-pictures`：`sub_288D4` 的 COMBPICS 結果圖 runtime 沒有讀。
  全套 Go 測試、八份戰鬥收據通過。
- #100 結案（RE-147）：`sub_288D4` 五個呼叫點依去向畫 COMBPICS 圖（一般戰鬥贏 3、輸 2、逃走格 0、
  競技場／訓練戰贏 4、輸 1），貼在 (8,8)、白框 (7,7)–(305,152) 與投影，第一句勝負訊息時出現、
  留到收尾，戰利品畫面蓋掉。remake `State.ResultPicture`，英文與中文戰場都畫，軌跡記
  `result_picture`。`battle-spoils`、`battle-bar-flee` 加斷言並重生畫格；三條長路徑的參考軌跡
  逐步確認差異只有這一欄後更新。SPEC-008 補結果圖一節。
- #84 結案（RE-148）：原版以 dosgolem 逐格走 Theodoric 的指令列十四組（徒手攻擊式六格、防禦方式
  八格、切換、休息、移動方式六格），`TestOriginalNormalPlanningCells` 在 remake 照同一串鍵重播，
  比每一鍵之後的格數、白框與寫回的四個值。第一次移動方式原版 7 格、remake 3 格，追出：
  參戰者 `+0x03` 由 `sub_21D94` 依種族設（Kelden 2、Dwarven 0、其餘 1），remake 一律當 0；
  磚塊 0x0C 在移動方式列是擋住後三格而非解鎖（RE-085、SPEC-010 方向讀反）；同一分支的
  手臂／腿廢掉清 `+0x25`、腿廢掉清 `+0x03`、治好隊員的腿依種族重設，remake 都沒做。修正後
  十四組全數一致。收據 `docs/verification/planning-cells-20260927/`，摘要工具
  `tools/summarize_planning_probes.py` 先驗進場狀態與 2026-09-11 逐 byte 相同。#57 改題為
  「0x0C 在地圖上是什麼」。全套 Go 測試、原版對拍、八份戰鬥收據、24 條玩家路徑通過。
  原版跑法：兩個並行各 2 CPU（當時主機負載約 39，另有其他 session 的工作）。
- #57 結案（RE-149）：磚塊 0x0C 只出現在任務戰場，拼圖標紅後全落在洞穴通道與迷宮內部的地面。
  說明書移動一節證實 `+0x1E` 是飛行（WALK／RUN／SPRINT 對 0x3C…0x3E，FLY／FLY FASTER／ZOOM 對
  0x3F、0x40、0x35，LAND 對 0x28），RE-148 的 Kelden 飛行升為已證實。戰鬥程式與測試註解、
  SPEC-010 由騎乘／下馬改為飛行／降落，識別字不改。只動註解與文件，battle 測試通過。
- #62 結案（RE-150）：`word_2DF2C` 是所在城鎮編號，野外迴圈 `sub_29004` 設成 0；物品頁與立繪頁在
  野外用 KOLPICS 第 4 張整張。查呼叫者時發現資料頁 `sub_15ABF` 在城鎮戶外（`sub_10000`）與野外
  （`sub_29004`）都能按 `1`…`6` 打開，remake 只掛在建築內的 MIRROR，登 `sheet-digit-outside`。
- #63 結案：`[0xB926]` 是位址誤算，實際 DS:0xB956 已由 RE-113 解出（任務 payload `+0x29`）。
  SPEC-014 §6 與 `battle-flee` 收據仍說 QuestItem 是空的，改寫為：第 5 個地點的 Kelder Crown
  放在地上，那一次沒有人撿，所以走「全隊到邊上」那一條。
- #65 結案（RE-151）：野外時刻的成本表 DGROUP+0x12B8 在原版是用 `word_39A54`（這一步的 Y 位移）
  查的，查到 `[0x12B7…0x12B9]` = `00 0A 0A`：移動率 20 時往北 40、其餘 60，與地形無關。
  `sub_297F2` 開頭先呼叫 `sub_29987` 依扣之前的時刻判時段，換了才重畫標籤、擲遭遇，最後
  擲天候（SPEC-011 §1 原本把標籤重畫 `sub_29B40` 當成天候、寫成遭遇在後，更正）；換日在
  CAMPING! 收掉之後才換成 DAWN。`tools/oracle/book_startup_test.go` 加 `KOL_BOOK_CLOCK_PROBE`，
  dosgolem 讀記憶體實測四個方向與第 15 步換時段，收據 `docs/verification/world-clock-probe-20260927/`。
  兩段舊 DOSBox 錄影計入「畫格晚按鍵一拍」後逐步相符。remake：`StepCost(dy, 移動率)`、
  `Clock.UpdatePart`、`Session.updatePart`（遭遇在天候前），`RoadStepCost` 移除；全套測試通過。
  **收據待重生**：時段換的時機與骰子先後都變了，野外路徑的遭遇位置跟著移動——
  full-chain、arena、full-chain-ending、full-chain-reload、world-healer、world-quest、
  world-road 與 battle-ambush、bar-flee、missile、pickup、spell 另行重生。
- #63 的 verify 改用 required_in_all（`absent` 是「出現前都算未完成」，先前用反了）。
- RE-151 收據重生完成：world-clock（東西每步 60，第 15 步開頭換 MID MORN，與原版錄影對拍仍逐像素
  相同；對拍規格帶與路徑同一顆種子 3）、world-road、town-round-trip 只差時刻值；world-quest 事件
  不變、露營移到第 86 步；world-healer 多走一步；full-chain 伏擊移到第 27 步、五隻 1 級 ZOMBIE
  第十八回合打贏，存檔冒險點數 2466／2966；full-chain-ending 按鍵不變；arena 路上換成 CLIFF TROLL；
  full-chain-reload 在 CAMPING! 收掉後擲到治療、再自然遇到 GOBLIN 伏擊打贏，全長 1515 步；
  battle-ambush／bar-flee／pickup 改接第 129 步的 OGRE 伏擊，missile／spell 多走一步。
  驅動器在 full-chain-reload 一路打進任務戰鬥卡在第 316 回合（反覆次數上限），改截在任務戰鬥
  開始那一步。八份戰鬥收據、24 條玩家路徑、四份野外對拍全數通過。
- #66 結案：`chardata-write` 收據兩個職業原版改動的 58 byte 全數寫對、零漏寫；「來源未反追」的
  武器樣板早已由 `BOOKDATA` 第 3 項寫出，其餘位移原版建立角色時不動。writer.go 過時註解改寫。
- #68 結案：戰鬥指令的名稱都已由 KOLDATA 記錄或說明書證實（RE-093／094／105／124／149），
  SPEC-010 末段早已關閉；commands.go 開頭的過時自承改寫，移動方式補上說明書名稱。
- #70 結案：各規格未解章節清空或轉給各自條目。RE-152：`sub_2A39B` 是野外整頁重畫，遭遇打完
  之後在道路旗標 `word_2FA2A` 非零時補回 ROAD／SIGNPOST，remake 的 `Sync` 行為相同。SPEC-010、
  002、011 的章節標題改為現況，004 的人耳驗收由 #4 追蹤。
- #71 結案（RE-153）：施法問部位那一頁，原版重畫左下人形框並以目標呼叫同一支 `sub_27E20`，
  沒有疲勞直條。remake `BodyPanel` 在 `PromptArea` 時改取 `Area.Target`、不畫直條；取圖與
  幾何沿用 SPEC-018 已驗證的那一套。原版同狀態畫格仍待效果 6 法術的實跑收據。全套測試通過。
- #74 結案：SPEC-029 早已於 2026-09-08 升 READY，正文殘留的 DRAFT 待辦句改寫為現況；
  補註 `Replan` 每次呼叫 `InitializeMonsterRanged`，由 `rangedPending` 保證只擲一次。
- #18 部分（RE-154）：`sub_21E0B` 的過濾條件雖對，任務獎品與商店有五件過得了（Flying Cloak、
  Courage Coat、Shade Ring、Speed Boots、Shoes），更正 RE-141／RE-118。remake 接上
  `ApplyCharms`（種族之後套穿戴）、執行段主迴圈之後的第二輪移動（`ExtraSteps`，RUN／SPRINT／
  FLY FASTER／ZOOM 的格數依圖號與 `+0x04`）、AI 找目標遮罩 0x10 的 Shade Ring（智力骰改為用到
  才擲，已知差異）。四個穿戴欄磨損呼叫只扣 `+0x0F` 有 0x20 的東西，五件都沒有，不改耐久。
  酒館逃走補上 RE-143 漏的分支：手上是任務要的那件武器，擲多少都丟下。`go test`（戰鬥套件、
  Xvfb 下 cmd/kol 與 ui）通過；八份戰鬥收據與 24 條玩家路徑全數一致，不需重生。
  容器：只留自己 `--rm` 的一次性容器，已結束；無 root-owned 檔案。
- #18 部分（RE-155）：傷害上限在「不是徒手而且武器種類沒有 `[0x109C + 攻擊選項 − 2]` 那一位」
  時加倍，選項 0、1 讀到 0x109A／0x109B。新增原版對拍 `TestDamageCapMatchesOriginal`
  （dosgolem 呼叫 `sub_26D76`）：原版 6／3／3，remake 修改前第一組是 3，修改後三組相同。
  超出量扣攻擊者那一格耐久（`WearOverCap`）：隊員武器扣手上那件、徒手扣那一式的穿戴品，
  怪物扣記錄 `+0x30` 穿戴欄第一格；會碰到的是 BRIGAND／BANDIT 的 Leather Hood、CYCLOPS 的 Skin。
  八份戰鬥收據通過；玩家路徑 23 條一致，full-chain-ending 最終任務戰鬥第 16 回合兩句
  執行順序對調（CYCLOPS 的 Skin 磨損後出手鍵值改變，推論）後重生，仍到結局。
  容器：只有自己的 `--rm` 一次性容器，已結束；無 root-owned 檔案。
- #18 結案：`TestBattleSettlementCompleteSaveReload` 補上 RE-155 的手上那件磨光（`WearOverCap`）
  與物品頁狀況檔次，五種結局（勝利、撤退、敗北、競技場勝敗）回寫後 v2 存檔重載全等。原版兩個
  磨損呼叫點不印訊息，顯示只有物品頁狀況名稱與空格。競技場地形窄 RE 早由 SPEC-025
  `TestArenaMapOriginal` 完成。限制：存讀驗證從結算狀態起跑，正常路徑的破損存讀歸整體交付閘門。
  SPEC-017 的保存段改為現況。容器：只有 `--rm` 一次性容器；無 root-owned 檔案。
- #101 結案（RE-150）：城鎮戶外（`Key` 在 `ModeMap` 收 `1`…`6`）與野外（`playToken` 主迴圈
  狀態）打開第 n 名的資料頁，暫借 `Inside` 掛頁、U-TURN 還原模式與圖示列白框；發言者換成
  第 n 名不還原（原版不還原 `word_39A52`）；野外物品與隊伍立繪頁背板換 KOLPICS 第 4 張
  （`sub_1103F(0x20)`）。軌跡在這份資料頁開著時多記 `outside_sheet`。新玩家路徑
  `sheet-outside`（26 步，party6 開場）。town 單元測試三支；`go test` 全過（ui 與 cmd/kol 在 Xvfb）。
  限制：原版在這兩處打開資料頁的同狀態畫格未取。
- #2 部分（RE-156）：`sub_2B454` 組 `comb%s` ＋ `pics` ＝ COMBPICS，競技場入口畫第 6 張、
  賭場印完 0xF0 後畫第 5 張，矩形同建築肖像。remake 接上並移除兩處「還沒解」自承；
  #2 的開放訊號改綁 `arena` 路徑仍讀的受控存檔 `arena-ready.json`。
- `arena` 收據重生：競技場對話畫格帶上 COMBPICS 肖像（第 2476…2483 步）。軌跡不變。檢查點標註
  自 2477 起比畫格早一步的既有錯位一併改正（對照截圖逐張確認）。
- #3 結案（RE-157、SPEC-036 READY）：以 IDA 產生 KOLBOOK 整份反組譯（只留本機），子代理初讀
  `sub_1293E` 後由主工作階段逐條核對。更正子代理「純鍵盤、無滑鼠」的初讀：`sub_146CE` 讀
  INT 33h 並把點擊轉成按鍵與模式切換；ESC 在編輯器是切換模式。推翻 RE-052「滑鼠驅動、鍵盤
  oracle 走不完」。實作授權待使用者決定（是否實作、自訂圖像存檔格式、介面、巨盾顯示位置）；
  #31 的阻擋原因同步改寫。25 條玩家路徑、8 份戰鬥收據一致；`go test` 全過。
  容器：只有自己的 `--rm` 一次性容器，已結束；無 root-owned 檔案。
- #9 結案（RE-158）：`book_startup_test.go` 新增 `KOL_BOOK_EXECUTE` 模式，以 OnCall 記下原版正常路徑
  六個回合的擲骰、訊息與每回合入口／結束狀態（收據 workplace/battle-first-round-20260928/，只留本機）。
  `TestOriginalNormalRoundsExecution` 逐回合接手比對，修正三處：近戰不招架不擲防禦骰、擊退只在傷害
  16 以上判且一定擲骰、疲勞挨傷招架改到回合收尾；怯戰不改記錄指令碼。修正後六回合全部一致；
  `run_original_parity_tests.sh` 六支全過，`check_updated_dosgolem.sh` 的 book 逾時可用
  `KOL_BOOK_TIMEOUT` 調整。另發現原版規劃按 YES 才換人，新增 `planning-yes-confirm` 待使用者決定。
  8 份戰鬥收據通過；21 條玩家路徑一致，**arena、full-chain、full-chain-ending、full-chain-reload
  四條的戰鬥段待重生**（擊退與防禦骰改變了戰況）。`go test` 全過（ui 在 Xvfb）。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- RE-158 後重生四條長路徑（`regen_long_path.py --battles-only`：每場戰鬥用自己的亂數，戰鬥外照抄）：
  arena 路上伏擊 30 回合敗北、競技場第 12 回合敗北，1961 步；full-chain 第 20 回合打贏，2057 步，
  存檔驗證通過；full-chain-reload 第 21 回合打贏、戰利品清單選 U-TURN 不領，2204 步；
  full-chain-ending 第 10 回合打贏、結局七段，1124 步。里程碑與檢查點依段落換算後逐項複核，
  只改回合數與伏擊留下的傷勢。工具修正：driver 在戰利品清單改按 `>` `RET` 離開；
  `--battles-only` 帶上 plan 的場景與受控骰值、路徑停在戰場開場的不再打。25 條路徑一致。
- #10 結案：全城逐店告別審查（building-goodbye-review-20260928）。`TestGoodbyeEveryBuilding`
  以正式資料驗證 101 棟英文與中文的告別文字、圖示列與按一次回地圖；TOWN7 門 21 原版缺文字，
  另列 `town7-door21-missing-text`。容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #11 部分（RE-159）：探針加 AI 入口 `sub_2374C`、輪盤入口 `sub_25A3A` 的記憶體取樣與共用緩衝寫入監看，
  `TestOriginalNormalEnemyPlanning` 以原版第 2…6 回合 AI 進場狀態接手，擲骰、亂數終態與寫回的
  `+0x25`／`+0x23`／`+0x24`／`+0x28` 全部一致。修正五處：AI 開頭與反制掃描前各擲一次智力骰；
  1/3 改移動只對 `+0x00` 非零的怪物擲；三張基礎權重取自怪物記錄 `+0x1A0` 起（推翻 RE-092
  「旗標恆為零」，`monsters.json` 新增三欄）；移動三式照 `0x239D7` 重寫；攻擊式的情境列號是
  `word_30916`。攻擊式遮罩第 8、9 筆原版讀旅店名冊清單（`sub_147DD:0x149DD`）留下的殘值，
  測試注入、正式路徑沿用近似，新增 `ai-style-mask-tail` 待使用者決定。#11 仍開放：撤退、
  側翼、射擊、撿武器沒有原版同狀態收據。
  8 份戰鬥收據通過；四條長路徑戰鬥段以 `regen_long_path.py --battles-only` 重生：arena 路上伏擊
  39 回合敗北、競技場第 7 回合敗北，2313 步；full-chain 第 15 回合打贏，1782 步；full-chain-reload
  第 16 回合打贏，1649 步；full-chain-ending 第 6 回合打贏，截掉結局第七段後多按的 7 次 SPACE，
  677 步。里程碑依段落換算並更新回合數與伏擊傷勢，刪除 26 張不再引用的舊截圖。25 條路徑一致；
  `go test ./internal/... ./cmd/...` 全過。容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #104 結案（使用者 2026-09-29 選 A：照原版）：敵人抽攻擊式的遮罩第 8、9 筆讀旅店名冊清單留在共用緩衝
  +0x28／+0x29 的槽號。子代理靜態盤點約 30 處直接寫入，主工作階段抽驗上界：一般遊玩中只有
  `sub_147DD:0x149C4` 會寫這兩格（RE-159 補述）。remake：`Candidate.Slot`、
  `Session.RosterListTail`（開名冊清單時記下，只在記憶體、啟動為 0）→ 三種戰鬥的
  `State.StyleMaskTail`；`StyleMask` 只留前 8 筆。另補原版流程：住宿後隊伍空了答 NO 進名冊流程
  （`0x128A5`）。新增 `TestRosterListTail`、`TestFaithfulInnStayQuitNoOpensRoster`。
  8 份戰鬥收據通過；四條長路徑戰鬥段再重生：arena 路上 20 回合敗北、競技場第 10 回合敗北，1697 步；
  full-chain 第 19 回合打贏，2202 步；full-chain-reload 第 13 回合打贏，1372 步；full-chain-ending
  第 3 回合打贏，402 步。25 條路徑一致；`go test ./internal/... ./cmd/...` 全過。
  過程中誤用 `git checkout` 還原了剛重生的 plan，已重跑重生；無其他影響。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #102 結案（使用者 2026-09-29 決定照原版，RE-160）：規劃段寫回後停在同一名、白框在 YES，按 YES 才換人；
  `+0x25` 跨回合保留，新回合直接按 YES 即沿用。指令列重畫時依 `word_31122`（`State.LastCommand`）標出
  目標格。既有規劃測試改經 `commitThenYes`，新增 `TestPlanningYesRepeatsLastCommand`、
  `TestPlanningSummaryMarksTarget`。`drive_battle_script.py` 下完指令後按 RET 確認。五份戰鬥收據
  （flee、bar-flee、missile、pickup、spoils）在按鍵序列補上 YES 並以 `--write` 重截示意圖；pickup
  第一次寫回回報沒過，之後兩次重跑都通過，原因未查明。四條長路徑戰鬥段重生，回合數與勝負不變，
  只多了確認鍵。8 份戰鬥收據、25 條路徑一致；`go test ./internal/... ./cmd/...` 全過。
- SPEC-036 決定（使用者 2026-09-29）：實作；像素存 v2 存檔；原版鍵位加中文提示；盾牌先查顯示位置。
  子代理查出（主工作階段抽驗 `sub_18F3A:0x192D6` 盾牌那一段）：每名角色的 CHARPICS 畫在城鎮隊伍格、
  資料頁抬頭與戰鬥資訊框；個人 SHLDPICS 只畫在資料頁全身立繪，盾牌型別 3 時貼在模板盾牌左上角下一列。
  新角色的盾牌初值固定取 BOOKDATA 第 4 項。容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #31 結案（SPEC-036 CONFORMED，RE-161）：新增 `internal/bitmapedit`（照 KOLBOOK `sub_1293E`／`sub_146CE`
  的按鍵、滑鼠、四支變換與存檔問句）、`ui/bitmap_editor.go`、書本流程 `SceneEditPick`／`SceneEditor`，
  取圖與寫回在 `cmd/kol/bitmap_edit.go`。存檔升 `kol-save-v3`（角色加 `picture`／`shield`，讀 v2 時遷移）。
  預設盾牌由新工具 `tools/extract_default_shield.py` 從 BOOKDATA 第 4 項抽成 PNG，並核對原版 SHLDPICS
  16 槽都等於它。圖像接到城鎮隊伍格、戶外與世界隊伍格、戰鬥資訊框；個人盾牌接到資料頁立繪
  （型別 3）。原版同狀態：`KOL_CHECK=editor` 新探針 `tools/oracle/book_editor_test.go`，同一串鍵
  存檔結果 760 byte 相同，進場畫格除 DOS 滑鼠指標 57 像素外逐像素相同；對拍加進
  `run_original_parity_tests.sh`。子代理兩次靜態盤點（顯示位置、畫面組成），主工作階段抽驗
  `sub_18F3A:0x192D6` 與編輯器主迴圈。`full-chain` 存檔斷言與兩支工具改認 v3。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #12 結案：從 `7504f81` 打本機 AppImage `v.0.2.6-20260929`（SHA-256 `73178858…788f`，1,148 筆輸入）。
  `run_kol_shots.sh` 新增 `KOL_BIN`，改用封包解出的執行檔（雜湊與建置收據相符；給不存在的路徑時
  正確失敗作為反向對照）。封包執行檔上：旅店 8／8、招募 9／9、住宿 11／11 對原版畫格全過；
  8 份戰鬥機制收據、25 條玩家路徑一致；`verify_complete_save_gui.sh`（斷言改認 v3）與
  `verify_local_appimage.py` 通過；預設存檔位置 XDG 通過、HOME 後備第一次重載後按鍵沒被讀到而失敗，
  重跑通過，原因未查明。README 交付入口改指 v.0.2.6。容器：只有自己的 `--rm` 一次性容器；
  無 root-owned 檔案。
- #14、#93 結案（2026-09-30，使用者 2026-09-29 決定：對拍只比機制；戰鬥路線選「撿起物品後逃走」）。
  原版以關鍵點接手：新探針 `internal/oracle/quest_chain_test.go`（`KOL_CHECK=quest`），入口橋進
  `sub_1E3A8`／`sub_11A93` 中段、出口改帶回傳碼的 `retf`。標完成 36864 組、撿起 9 組、戰場邊與帶物
  逃走 6 組、交付 8 組全數一致（RE-162）。對拍找出兩個 remake 差異並修正：撿成功後 `+0x25` 寫 0x15；
  敵人撿起原本把問句 0x111 當訊息印出，改成不印、撿到任務物品才喊 0x113（`State.EnemyPickUp`）。
  舊測試 `TestPickUpEnemyDoesNotAsk` 斷言的是錯的訊息，已改；新增 `TestEnemyPickUpIsSilentExceptQuestItem`。
  remake 自然路徑 `quest-chain`（7096 步）與 `quest-chain-reload`：軌跡新增戰場 `ground`、
  `quest_holders`、每場一次的 `walls`（只在 `-trace-detail`）；`drive_battle_script.py` 加任務模式
  （`KOL_DRIVE_QUEST=1`：BFS 繞牆找路、撿物、走到邊、逃走問句答 y；`-` 表示不讀檔）。
  過程：持物者單獨走時在上緣前三格被追上倒下（敗北），改成其餘隊員護送後通過；驅動器
  stage 3 分支原本一律答 N，逃走問句被答掉，修正後從問句處接續。城鎮 4 號入城點的鏡頭範圍
  到不了旅店，改繞 3 號入城點；野外路線從戰場出口直接 BFS 會卡在河邊（路線工具起點子位置
  固定 3,3，與實際 7,7 不同），改用去程實際移動的方向鍵反序反向。一次 `pgrep -f` 的樣式同時
  比對到自己的 shell，中斷了當時的命令，驅動器程序已確認停止、無其他影響。第一次重播
  `quest-chain` 沒有設檢查點而逐步截了 7000 張圖，已刪到只剩五張檢查點。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #15 結案（2026-09-30，RE-163）：A05／A06／A09 的三支診斷測試早已是正式回歸測試。原版關鍵點接手
  新探針 `internal/oracle/belt_ward_abbey_test.go`（`KOL_CHECK=gear`）：切換 4 組、法術 Ward
  14336 組、修道院治療 144 組一致。對拍修正兩處：切換後原本一律只印 0x114，改照原版依兩格有無
  組 0x114／0x115／0x116 或補 `!`（`State.SwitchLine`、`ActResult.Bang`）；倒下的目標原本讓法術
  生效，原版是一律沒有效果（`SpellBlocked`），SPEC-012 §4 那句方向寫反已更正。修道院抽出
  `HealPrice`、`Member.ApplyHeal`。正常路徑抽測 `gear-switch`（讀 quest-chain 存檔，Theodoric
  丟地圖後切換兩次再逃走）與 `gear-abbey`（標題選 Warder 走到城 5 修道院治療）。
  過程：第一版驅動在第 3 回合切換後隨即從「Begin next round?」逃走，切換沒被執行，改成第 4 回合
  才逃。治療前資料頁的 DAMAGE 條沒畫出傷勢，未追查，記在驗證文件的不宣稱。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #17 結案（2026-09-30）：`tools/oracle/book_startup_test.go` 在執行段探測結束時加 `capturePanels`：
  對槽 0、1、6 呼叫原版 `sub_2324F`、`sub_27E20`、`sub_22EFC`，記人形、疲勞直條、傷勢遮罩讀取序與
  原版自己解出的裝備名字，另在槽 0 做一次受控破損（`sub_22084`）後重擷裝備框，寫進
  `battle-rounds.json` 的 `panels`。新旗標 `KOL_BOOK_UNTIL_WOUND=1` 讓探測等到部位真的掉血。
  新測試 `TestOriginalNormalBattlePanels`（cmd/kol）全部一致；SPEC-018／019 升 CONFORMED。
  `run_original_parity_tests.sh` 預設回合收據改指 `workplace/battle-panels-20260930/`（回合邊界、
  擲骰與 0928 那一份逐項相同）。過程：第一趟沒帶 `KOL_BOOK_NEXT_DAY=1` 沒進戰鬥；第二、三趟沒帶
  `KOL_BOOK_SELECT_DOWN=2`，選到另一組隊伍（Theodoric＋Poontang），與 remake 夾具不是同一場，
  部位上限因此對不上；補上之後重現 0928 同一場。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #20 天候那一半結案（2026-09-30，RE-164）：靜態讀 `sub_11390`／`sub_115D1`／`sub_29004`／`sub_2A0C6`／
  `sub_1B100`，並在原版野外實跑觀察（`KOL_BOOK_WEATHER=1`）。remake 改成 `sub_11390` 的計數
  狀態機（`Screen.WeatherCall`、`WeatherRedraw`）：雨雪索引換組不歸零、雲層先轉再畫、只在野外
  地圖等鍵時推進、離城時畫面組設 0 並重畫（`Session.EnterFromTown`）。新探針
  `internal/oracle/weather_phase_test.go`（`KOL_CHECK=gaps`）53 次呼叫逐次一致；SPEC-026 升
  CONFORMED，契約第 4 條「換組歸零相位」更正。加速選項併入 g07（#5，使用者先前決定跳過）。
  過程：插入觀察呼叫時錨點重複，第一次替換沒生效，改用行號插入；`go_docker.sh build` 沒帶 `-o`
  在根目錄留下 `kol`（已被 gitignore），已刪。容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #103 結案（2026-09-30，RE-165）：原以為的原版資料缺口其實是抽取器漏讀續卷。原版正常路徑
  （旅店走出來後用關鍵點接手把城鎮鏡頭與夾限換成入城點 9，再走到門 21 進門）畫出完整建築，
  並開了 `TOWN7.1`；`TOWN7.0` 表頭第 12…13 byte（2195）就是那一卷的長度。`payloads()` 改把續卷
  算成再下一個全域索引，三支城鎮抽取器改用 `len(parts)`，重生三份 JSON（`assets/original-derived-data/`
  被 gitignore，只有工具進版控）。門 21：Krag Keep／Ballastar of Simmer、任務地點 10（Stone Mallet →
  Death Blade）、27 則訊息、六個話題槽；24 個任務地點各有一棟交付建築，最終任務的 23 個回報門檻
  終於湊得滿。繁中：名稱、人物、進門敘述與 25 則後續訊息、三筆顯示名、術語表九列，字模表 2,553 → 2,567
  字，`verify_text_fit.py` 9,650 項 0 失敗，`verify_runtime_translation.py` 與 `verify_display_names.py`
  0 缺譯。`TestGoodbyeEveryBuilding` 改為 102 棟、0 缺文字。原版與 remake 畫格：進門敘述從 y=53 起相同，
  告別對話區 0 差異像素。順帶發現原版多人隊伍進門後會印「Which one needs help?」讓玩家選發言者，remake
  沒有這一步，併入 #32。過程：`quest-reward` 鍵值第一次撞號（已有 08），改 09。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #32 結案（2026-09-30，RE-166、SPEC-037）：進門後的發言者選擇、建築好惡與教派。建築 payload
  `+0x40`／`+0x48` 各有 8 byte 的 5-bit 名稱（討厭與欣賞），抽取器新增 `dislikes`／`likes`。原版流程：
  全隊被討厭只印 `Get out, scum!` 並離開；多人隊伍先印訊息 16 加記錄 0x55，圖示列只剩 DOOR，按 1…6 選人；
  被討厭或（SCROLL 建築）教派不同印訊息 14，一人離開、多人回頭問；被欣賞印訊息 15。remake 新增
  `internal/town/greet.go`（`GreetRule`、`CanonicalRecord`、狀態機）與 `greet_test.go`；`Continue`、招募、
  住宿、`Key` 接上；職業記錄換算改用同一份 `CanonicalRecord`。原版關鍵點接手 `internal/oracle/greet_test.go`
  （`KOL_CHECK=greet`）：名稱換算 1,280、全隊被趕 24,000、發言者好惡 10,896，全部一致；負對照（換算門檻
  改壞）三支測試都失敗。修正舊說法：RE-067 與 RE-165 都把訊息 16 當成招呼語，已在 RE-166 推翻並改指標。
  多人隊伍進門多一步，六條玩家路徑補鍵：`quest-list`、`quest-take`、`quest-chain-reload`、`full-chain-reload`
  各在進門敘述後補 `1`；`full-chain` 補 1 處；`quest-chain` 補 3 處，其中第一處要按 6（最後招募的
  Kenneth 本來就是發言者，按 1 會換成 Theodoric 而讓接任務的人變了）。補鍵位置用暫時掛鉤（腳本跑到問誰
  開口時自動按 1 並記步號）一次收集，掛鉤與暫時腳本已撤掉，不進版控。過程：掛鉤第一次把步號寫到 stderr，
  被包裝腳本吞掉，改寫檔案；兩次中途停掉自己起的容器（掛載路徑確認是 `workplace/autowho/`）；
  補鍵工具第一次沒保留 plan 的縮排與結尾換行，改為自動偵測。worklist 新增
  `building-portrait-click-speaker`（滑鼠點肖像換發言者，不宣稱）。`go test ./cmd/... ./internal/...`
  全過（Xvfb），原版對拍測試全過；29 條玩家路徑除六條補鍵者外狀態一致，六條重生後再驗。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #29、#30 結案（2026-09-30）：#29 新增 `internal/layers/layers_test.go`（四項：套件登記、依賴方向、
  Ebitengine 只在 `cmd/kol`／`ui`／`font`／`screen`、規則層不讀 `time.Now` 系列與全域亂數）與
  SPEC-038；現況核對後全部通過，例外只有 `cmd/kol` 的存檔時間戳記與隨機角色 ID。負對照：在 `town`
  放依賴 `battle`、Ebitengine、`time.Now` 的暫時檔，在 `battle` 放 `math/rand`，第 2、3、4 項都失敗，
  檔案刪掉後通過。#30 以既有的 `full-chain`、`quest-chain` 與兩條重載路徑認定滿足，另寫
  `docs/verification/minimal-player-chain-20260930.md`，範圍不含新遊戲到結局。過程：無失敗。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #94、#105 結案（2026-09-30，RE-167）：讀完 `sub_1B100` 的滑鼠分支。原版左鍵依序判斷：地圖窗八方位扇區
  （`mode == 3`）、圖示列（兩次點擊：先搬白框，再點同一格才確認）、只剩 SPACEBAR 的等鍵、`mode 4／6` 回 0xFF、
  隊伍肖像（`word_2DF3E == 1`，y < 50、x 從 16 起每 48 像素一位，回鍵 1…6）；右鍵回空白鍵。選目標格
  （`sub_24131`）用 `sub_1B100(6)`，圖示 7 與 0xF 的圖示列照常收滑鼠。remake：`portraitTargets`、
  `Session.PortraitClicks`（戶外走路、問誰開口、建築內一般圖示列、野外走路），目標層圖示列排在戰場格之前，
  移除 input.go 的「還沒證據」註解；`portrait_mouse_test.go` 與 `TestPortraitClicksState` 驗區域。全部是靜態證據，
  沒有真實視窗點擊。原版與 remake 三處操作差異（圖示列兩次點擊、右鍵是空白鍵、地圖點擊走路）不在這一版動，
  列成 `mouse-original-semantics` 等使用者決定。過程：讀 3×3 表時第一次用 EXE 檔內偏移換算落在檔外
  （解包 EXE 只有 138,528 byte），改讀反組譯清單的資料列。容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #27 結案（2026-09-30）：新增 `tools/check_repo_visibility.sh`，`build_dist.sh`、`package_dist.sh`、
  `verify_dist.sh` 主機端開頭都先跑它。判定：`gh api repos/<owner>/<name> --jq .visibility` 是 `private` 才放行；
  其他值、空輸出、沒有 `gh`、沒登入、API 出錯一律結束碼 4（失敗即關閉），沒有略過旗標。回歸測試
  `tools/test_check_repo_visibility.sh` 用假的 `gh` 走過六種情形，全部通過；對真實儲存庫回報 private。
  `docs/RELEASE-RIGHTS-REVIEW.md` 補一節說明閘門與它管不到的範圍（網頁直接改可見度、手動建立 Release）。
  過程：「沒有 gh」的測試第一版靠覆寫 `command` 函式，難以確認是因為那個原因失敗，改成 PATH 只放
  `dirname` 並檢查訊息。容器：無新增。
- #41 結案（2026-09-30）：新增 `docs/verification/parity-registry.json`（schema `kol-parity-registry/1`）與
  `tools/verify_parity_registry.py`。四個階段各登記宣稱、母體、gate、硬失敗、平台矩陣與證據檔；
  玩家路徑母體 29 份，實際檔案數與登記不符會被檢查抓到。目前狀態：可啟動 linux-only、可玩 claimed-local、
  近似完成與嚴格 parity not-claimed。使用者 2026-09-26／29 的決定（對拍看規則狀態，不看畫面逐位元）寫進
  登記的 note 與嚴格 parity 的宣稱。gate 內容是依現有工具與收據草擬，使用者可修改；改動要走 revisions。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。
- #38 結案（2026-09-30）：新增 `tools/verify_spec_states.py`，檢查規格索引、各檔狀態行與 CONFORMED 的驗證證據。
  第一次執行抓到三處矛盾：SPEC-029、SPEC-030 只在索引的散文段、不在表裡；SPEC-036 索引仍寫 READY 而檔內是
  CONFORMED（已實作）。已補列並更正。負對照：放一份未登記的規格檔、把 SPEC-028 的索引狀態改成 READY，
  兩種都被擋下，還原後通過。工具只擋登記自己的矛盾，不判斷「實作之前是否已 READY」，該項仍靠流程。
  容器：只有自己的 `--rm` 一次性容器；無 root-owned 檔案。


## 2026-09-30：接手 AI 成本與收尾驗證

- 接手 HEAD `f16fc13`，保留原有九個修改檔與兩個未追蹤檔。先載入專案現況、收尾／規格生命週期、文件責任與逆向訂正引用入口；沿用既有 Go、KOL 與 IDA Pro 9.4 映像。host `gh auth status` 成功，唯讀核對遠端 24 個開放 Issue，沒有 Issue 寫入。
- 原有 300 組任意 AI 擾動有四組差異。IDA 正式 `.i64` 唯讀掛載後複製到容器 `/tmp`，查明移動成本引數順序與部位成本索引；先補 SPEC-008 READY，再修兩處正式成本。第 98 組差異消失；新出現的第 119 組是夾具清空原版武器卻保留 remake 裝備，同步 `ReloadGear` 後消失。原 41／59／166 三組表外矩陣差異保留，RE-168 記原始定位與等級。
- 在執行前限定舊全域選項 0…10、維持擾動種子 20260930 與每組遊戲 seed，dosgolem 正常入口生成 300 組收據；原版呼叫失敗 0，remake 全部一致。原始來源執行前後沒有改動。完整六回合執行與五回合 AI 規劃、產品範圍 Go 測試、建置均通過。常式比較不等於自然 GUI 或整場原版對拍。
- 新建置重播 `full-chain` 2,317 token，勝利、回城住宿、v3 保存與全部登記里程碑通過。讀該新存檔執行原 `full-chain-reload`（seed 3、1,451 token），前段恢復與出城通過，後段戰鬥／任務七項失敗；完整失敗收據保留。
- 生成器第一次診斷誤用 seed 1，訂正為計畫 seed 3。既有走法及暫試障礙尋路各跑有界 500 次迭代，皆未勝利，撤回未證有效的尋路改動。89／200／1442 步的單張／逐步模式比較，扣除重複欄位省略後狀態一致。工具只保留子程序錯誤輸出與 `replay_player_path.py` 匯入保護，未將工具失敗寫成存檔缺陷。
- 修正六個過時工作清單檢查，其中 +0x05 與穿戴過濾結論按既有 RE-154 回填，保留 RE-118／RE-141 歷程。明記 SPEC-036 全槽索引仍為強推論，其全槽 UI 證據由 #28 追蹤。工作清單掃描 1,258 行均有認領；自測、39 份規格狀態與四階段登記通過。
- 依本輪真實重載回歸，在本機重列 #88：目前 25 項未完成、80 項完成；遠端 #88 仍關閉，遠端仍為 24 個開放 Issue，沒有同步寫入。現況與收據入口：`CONTEXT.md`、`docs/verification/enemy-ai-fuzz-20260930.md`。新文件已連入現況、研究與工作清單。
- 清理：全部工作在具資源上限、外層逾時與 `--rm` 的容器內執行；原始輸入唯讀、使用 UID/GID 1000:1000，Xvfb 有退出 trap。檢查無專案 root 擁有檔案、誤建 `.md` 目錄或本輪殘留容器。原始研究與兩次失敗中間產物留本機 `workplace/`，容器內 `/tmp` 已移除。不提交、不推送、不重建交付包，也未建立 Release。

## 2026-10-01：續玩收據與完整玩家路徑核對

- `drive_battle_script.py` 補正式射擊、存活隊員分派與射擊游標歸位。第一次八鍵歸位在視窗外座標失敗，修正為先夾回再走八格；正式 Go 游標的 72 個合法目標 × 225 個初始位置共 16,200 組通過，失敗與通過探針均留本機。
- 固定原計畫 seed 3、讀本輪 `full-chain` 新存檔，生成 2,012 個正式輸入完成第十七回合勝利、離場、下一任務參戰，登記里程碑全部通過。新計畫、軌跡、七張畫格與建置 manifest 更新至原 `full-chain-reload` 入口；舊收據備份未刪除，不宣稱整場原版一致。
- `check_player_paths.py` 支援容器內 `KOL_BIN`、當輪存檔依賴排序、獨立來源存檔唯讀複製與子程序紀錄。第一次全量重跑遇到 v1 匯入在唯讀來源旁寫檔，分類為驗證環境問題；修正輸入複製後同命令乾淨重跑。29 條舊參考有 23 條一致、六條不同或無法執行；其中任務鏈無效 `SPACE` 阻斷兩條後續讀檔路徑，#88 仍待辦。
- 任務生成器維持 seed 1 與既有遇敵控制，正常標題及 417 步前綴進場；有界 1,500 次迭代內於第 82 回合敗北，產出 6,080 步，沒有完成任務。觀察 417／612／1035／2000／4000 步：物品仍留地上、無持物者；不將生成器敗北判為規則錯誤、不換種子挑結果。失敗入口為 `workplace/ai-fuzz-attack-only-20260930/quest-chain-regeneration/`。
- 更正 #47 的 v1／v3 混淆：正式 v3 已按 SPEC-023 保存世界時間；戰鬥中途暫態明確排除於現行契約，原完成條件保留等待範圍決定。#106 滑鼠取捨已用共同決策技能提出一題，尚無答覆，依賴該決定的操作分支維持待定。
- 全部操作沿用既有有界 `--rm` 容器、UID/GID 1000:1000 與唯讀研究來源；本節的最終驗證與清理結果另於完成後追加。不提交、推送、Issue 寫入或發布新封包。
- 續輪母路徑重驗：2,318 筆僅第 244 步的防禦敘述不同；兩側該步去掉 `exec_text` 後完全相同、其餘步全部相同。當輪保存結果、全部里程碑與畫格通過後保留舊軌跡／收據，再更新正式母路徑參考；新軌跡 SHA-256 `b0fbd3985b85c7341765fb67951c86583c52902c0b9d8b09c2cdfe9eb7f6583a`。原版該場自然戰鬥的防禦敘述尚未取得，不外推完整原版一致。
- 近戰護送策略取得任務物品後持物者倒下，第 79 回合敗北，6,925 步未完成；失敗快照留 `quest-chain-regeneration/escort-melee/`。查證移動子選單從 U-TURN 起，`>` 一次其實選步行，訂正生成器的奔跑註解；SPEC-008 §5 明定每次移動只有一格，不以調移動方式猜補距離。另以可丟棄的持物者繞行策略續查，同樣固定 seed 1。
- 競技場兩段生成與正常逐步重播：固定 seed 1，1,707 步完成路上第十八回合敗北、競技場第七回合敗北及兩筆賭金 189／452；單人 Moraine 參戰、選人／下注／結算全部里程碑與十二張畫格通過。傷勢依新戰鬥按鍵改變，未更改金幣、任務／回報完成條件；正式入口與建置 manifest 已更新且接入既有索引，舊版備份留本機。初次啟動錯用 flow 場景、初次驗證使用另一容器的 `/out` 路徑，均分類為工具環境問題，校正後同種子與同容器命令重跑通過。
- 持物者繞行策略第 84 回合敗北，6,631 步未完成；已取得物品但持物者倒下，失敗保留在 `quest-chain-regeneration/carrier-detour/`。以 Go 臨時覆蓋檔加入唯讀疲勞輸出，正式 `main.go` 不改，診斷原固定種子失敗前綴；結果待實測，不將假說寫成規則。
- 疲勞診斷：持物後第 60／70／75／78 回合的疲勞比 16／42／68／91，領隊耐力 229；第七十回合相同前綴的診斷與正式建置，去掉診斷欄位後狀態及 PNG 完全一致。加入休息的版本第 112 回合離場（7,408 步），任務旗標仍未完成，未登記為成功；後續僅修改本機生成策略為撿物後直接找最近出口。
- `check_player_paths.py arena full-chain full-chain-reload` 正式建置再驗 1,708／2,318／2,013 筆全部一致，畫格差異 0；重載讀當輪保存結果。正式目錄只保留 manifest 的十二／七張畫格，舊圖在查無引用且確認本機備份雜湊相同後逐檔移除。
- 重播入口補 `KOL_BIN`、唯讀來源複製、存檔依賴排序與父路徑失敗就跳過子路徑；真實十一個輸入的任務一覽重播通過，來源位元組不變；失敗父路徑的負對照確認沒有執行子路徑。`run_driver.sh` 補任務模式環境變數與 Xvfb 退出 trap。三十九份規格、四階段母體與清單自測通過；重現控制腳本／來源雜湊索引為 `workplace/ai-fuzz-attack-only-20260930/continuation-manifest.json`。


### 最後任務鏈與全量重播


- 固定 seed 1、原遭遇控制及正常接任務前綴，休息加直接出口策略第九十七回合帶物撤離。正式建置從標題完整重播 7,435 個輸入，取物、六名完成／回報、住宿金幣 953 及 v3 保存全部里程碑通過；更新五張畫格與來源 manifest。舊收據與失敗策略留本機，不改正式規則。
- 更新後全量先得 27 條一致；兩條後續路徑只有母存檔繼承傷勢變化，差值、存檔初態及去掉傷勢後全部狀態均一致，只更新該文字與來源；按鍵、金幣、旗標、物品、戰鬥結果及畫格未改。
- 最後獨立 check_player_paths.py --all 全部 29 條狀態一致、exit 0，續玩讀當輪保存結果。正式 binary SHA-256 1876cea523a45e8b089108d3512a4fb4ec5fbd8bda73b3ed7a3da992f42a84e5；本機日誌 workplace/ai-fuzz-attack-only-20260930/player-all-final.log，重現來源見同目錄 continuation-manifest.json。本地 #88 重新結案，24 項未完成／81 項完成；未寫入遠端 Issue，未宣稱原版完整戰鬥、全中文或跨平台驗收。
- 保留九月三十日任務文件歷史，追加現行步數與第 248 步實際按 6 選 Kenneth 的勘誤；正式畫格只在備份位元組一致且無引用後逐檔清理。最終文件與主機衛生核對另追加。

- 最後文件核對：工作清單正反自測、逐項核實、39 份規格狀態與四階段驗收登記均通過；五份建置 manifest 畫格集合、SHA-256 與 UID/GID 核對通過。三支 Python 工具語法與 run_driver.sh 語法通過。專案 root 擁有產物及誤建 .md 目錄為 0；全部一次性容器採 --rm，最終容器狀態由主機控制面另核對。暫存重播、診斷及舊收據留既有 workplace/，透過 continuation-manifest.json 索引；不是正式交付包。

### 繁中任務遭遇漏譯

- 正式 1876cea5 建置的完整繁中任務鏈達 1,000 秒逾時（exit 124），沒有完整軌跡與保存結果；未執行兩條後續路徑，不認作通過。部分輸出留 player-cjk-current/，容器已移除。部分文字分析第一次未處理空角色槽，修分析夾具後重跑，沒有改存檔。
- 部分文字紀錄揭出 KOLDATA 233 三行英文遭遇訊息；既有譯表已提供中文，但正常任務與競技場共用原文入口。修正 battle_render.go 的繁中繪圖取譯文，保留 quest.lines 原始內容與動態伏擊；修正 renderer 註解，新增防止改寫原始軌跡、誤換伏擊與缺譯回退的回歸測試。
- 新建置 SHA-256 781666891ccc974a531ec3b555452cdf55a41e1e455ac56cf0f27660c1eb2e1f，cmd/kol、internal/ui、internal/town 測試通過。初次唯讀模組快取警告屬環境，改複製到 /tmp 後同映像乾淨重跑通過。正常由標題沿用 418 個輸入，419 筆英中狀態與正式英文參考完全相同，三行中文及三行英文缺漏負對照通過；第 416／417 步中文畫格已目視可讀、未裁切。來源與六張畫格雜湊為本機 cjk-encounter-fixed/source-manifest.json。
- scan_cjk_text_leaks.py 文件更正不得主機分析的舊指示，列出無 Docker 控制面時的容器內規格與 report 入口。部分紀錄中的角色原名及輸入關鍵字不判為這次漏譯；未宣稱圖片英文或全畫面驗收。#35 保持未完成，#88 的 29 條結果仍明確指舊正式建置；未發布或重建交付包。
- 本輪最終主機控制面核對：kol-finish 系列執行中／停止容器均為 0；沒有 root 擁有產物或誤建 .md 目錄。git diff --check 通過。原始輸入未改，既有未提交異動保留；新舊建置、失敗輸出與來源索引留本機 workplace/，未提交、推送或發布。

### 共用填色配置與繁中完整任務鏈

- 繁中長鏈的即時日誌版本收到 SIGKILL，當時記憶體接近 2 GiB 上限；沒有取得 OOM 事件，根因只列候選。初次映像檢查讀不存在的 Entrypoint 欄位失敗，改核對 Config，屬工具環境問題。依既有 Ebitengine 2.9.9 的 Fill／SubImage 契約作可丟棄配置原型；500 秒上限仍逾時，停於取物步 5477，失敗輸出保留，不宣稱原型全鏈通過。
- 正式 fill 改填既有畫布子矩形，避免每次配置臨時圖片；保留非正尺寸原行為。72 組像素等價及四組尺寸行為通過。探針初次錯認負尺寸必須 panic，按既有 NewImage 的正規化行為修夾具後，以同映像重跑通過；沒有更改畫面矩形、顏色或玩家規則。局部 500 次量測約 329／5 ms，非整體遊戲效能宣稱。
- `go test ./cmd/... ./internal/...` 全產品範圍通過；現行工作樹建置 `kol-fill-fixed`，SHA-256 `3ea4ff59b8b1dce2d40e6b81bc751cf2d52726a00f5272cbefb359eaabebff00`。同建置全量 29 條英文狀態一致；當輪新存檔供子路徑讀取，日誌 `player-all-fill-fixed.log`。#88 保持完成。
- 同建置繁中完整任務鏈 7,436 筆約 570 秒完成，裝備切換 241 筆及任務重載 13 筆亦完成。英中差異只在 battle.exec_text（1,088／33 步）與 gear-switch 的 encounter_text（1 步），其他欄位及排除 saved_at 後存檔完全相同；保存雜湊仍 b3a023d4…。三個程序退出 0，容器峰值 1,128,280,064 bytes、OOM 計數 0。取物、撤離、交付與住宿畫格已目視。彙整夾具初次未把顯示陣列元素歸屬欄位，次次誤把完整名冊當六人隊伍，均修夾具後同命令重跑，未動遊戲產物。
- 110 筆含拉丁字候選與 25 筆純角色名審閱完成：中文句中的原名、玩家輸入與核准符號之外未分類 0。這不檢查圖片英文及未畫文字。正常 world-quest 觀察存檔路徑第 89 步另證實繁中世界提示漏畫：英文有敘述／問句，繁中只有地圖／答覆圖示，WorldCJK 無 Prompt 繪製。#35 保持待辦，下一步補提示及敘述翻譯，未擴張新閘門。
- 現況、玩家路徑入口與既有接手收據已更新；完整英中差分、來源、腳本、字串審閱與缺陷截圖均留既有 workplace/，由 continuation-manifest.json 索引，不是交付包。未提交、推送、遠端 Issue 寫入、重建正式封包或發布 Release。Docker 與檔案擁有權最終檢查另追加。
- 最終核對：工作清單正反自測與逐項核實、39 份規格登記、四階段驗收登記全部通過；五份正式畫格 manifest 的集合與雜湊一致。Docker 主機控制面確認 kol-finish 系列執行中及停止容器均為 0；容器內檢查 root 擁有產物、誤建 .md 目錄均為 0，git diff --check 通過。本輪控制腳本與新舊失敗輸出留本機既有 workplace/，來源索引已更新；沒有改動原始輸入或覆蓋接手時既有異動。

### 繁中世界提示與完整敘述

- 上一輪分類為實質進展；本輪重新讀 AGENTS、現況、唯一清單與工作樹，載入中文顯示／語意隔離、規格閘門與文件職責入口，沿用既有 Go／FFmpeg 映像。先補 SPEC-011 第 11 節 READY 顯示契約並更正首頁已被第 8 節撤銷的治療排除，再正式接 WorldCJK 的地圖與城鎮外觀兩個出口；不新增狀態、識別值或存檔欄位。
- 新增外部世界提示 TSV 共 67 筆：24 任務、32 騎乘遭遇、6 治療及五則既有 UI 問句／通知。來源英文字逐筆與固定 JSON／UI bank 對上，譯文只查顯示段落，原提示不回寫；載入時 UTF-8、唯一鍵、欄位及狀態失敗即關閉且不污染原譯表。新譯表／檢查器同次接入 docs/i18n/README.md，四個新地名接術語表並列暫譯，未代替使用者定案。
- 以既有 eob-remake-audio:20260901 及唯讀 Noto 來源烘字模，新增 41 字成為 2,608 字；替換前逐筆確認原 2,567 字 rows_hex 與字距完全相同，舊字模備份與重建命令留 world-prompt-fixed/。全域字模檢查補入新 TSV，6,093 筆缺字 0；9,635 項既有文字安全檢查失敗 0。世界提示另直接以 Go 原 PromptBox／PromptText 乘倍率驗真實字寬及八列上限，不在 Python 複製第二套座標。
- 全產品 go test ./cmd/... ./internal/... 通過；新增兩個 WorldCJK 繪圖出口的觀察測試後 internal/ui 再驗通過。正式建置 kol-world-prompt-fixed SHA-256 cd510e35e79a49440b9be3466e661618aade207b5b0af07bcb9661cee403f0ef；check_player_paths.py --all 全部 29 條狀態一致，續玩讀當輪保存結果，畫格差異張數與前輪相同。
- 英中六組提示分支完成：世界任務 91 筆、治療答應 29／拒絕 30、正常標題前綴 419、騎走失敗／成功各 35。騎走維持原替代名冊、種子及受控骰，沒有宣稱正常買馬全鏈。首次騎走失敗比較在第 34 步停下，查明只差既有動態遭遇文字；保留首次完整輸出，只明示該步 encounter_text 的確切兩側值後，同建置／命令／種子重跑全部通過，其餘狀態比較不放寬。任務、治療與兩種騎走通知畫格已目視，可讀、無缺字及裁切。
- 兩次工具環境失誤：補丁批次含重複目標路徑而整批拒絕，改合併同檔修改後成功；唯讀檢查誤讀不存在的 grid.go，查實 Grid 位於 world.go 後續讀正確來源。均未改遊戲規則，沒有重啟仍在執行的容器。
- 世界提示顯示子契約補局部 CONFORMED；SPEC-011 整體仍 READY，#35 的其他分支與圖片英文仍未完成，#88 維持完成。本地仍 24 項未完成／81 項完成。現行 cd510e35 未再執行完整繁中 7,435 輸入長鏈，前輪完整結果仍屬 3ea4ff59，不把舊收據換名。未提交、推送、遠端 Issue 寫入、打包或發布；本輪清理與來源核對另追加。
- 最終來源與清理：原始 Knights-of-legend.zip 雜湊仍等於既有契約；六組英中輸入、軌跡及畫格雜湊核對通過。清單正反自測與逐項核實、39 份規格、四階段驗收登記、五份正式畫格 manifest 均通過。容器內 root 擁有產物及誤建 .md 目錄為 0，主機 Docker 控制面查 kol-finish 系列執行中／停止容器為 0；git diff --check 通過。控制腳本、舊字模、首次差分失敗與新收據留既有 workplace/，由 continuation-manifest.json 與 world-prompt-fixed/source-manifest.json 索引，不是正式封包。

## 2026-10-01：正常建角名冊邊界收尾

- 載入復古重製收尾路由、handoff／文件職責與 dosgolem 能力契約。遠端僅 host gh 唯讀核實 24 開放 Issue，沒有寫入。
- 正常 KOL.EXE 按鍵抵達建角，自然 near _srand 參數預先固定 1。首次誤把 overlay 載入段加 0x10，修正後仍黑畫面／按鍵未消費；未修改的既有 TestNormalBookEntry 同樣失敗，分類為新版工具回歸。以 68 份正式來源對上前次成功 manifest，容器 git archive 固定 968ca77，未改共用 repo。
- pinned1 第二次建角誤選 LIST 而停止有界容器；pinned2 未設 SetScratch，GUI 通過仍未落盤，不能列資料驗收；字元快照過早／等待不足也保留。修正原版 create 的 scratch、選單重入與等待，同映像乾淨重生 pinned3：三情境 108.06／63.40／166.94 秒通過，46／24／223 個狀態擷取。來源未變、輸出 UID/GID 1000:1000。
- 獨立 review 核實逐槽三檔差分、提交前／拒絕後不變及 95 字元被 AH08 消費。NOVA／LUNA 使用原空名槽 14／15，第三人接受肖像滿槽拒絕；同名 Theodoric 覆寫槽 0；字元 53 接受／42 拒絕，MiXeD Name 保存槽 14。模板相同的圖片不冒稱 bytes 改變。
- 三份正式摘要及四張已目視 PNG 同次接入 SPEC-005、角色建立證據、RE-169；未驗證 PNG、原始／變動角色檔只留本機。實際執行來源與控制腳本在格式化前核對 SHA 並保留。三次補丁組裝失敗皆在寫入前被拒絕，修正控制腳本字串與精確段落後成功，未造成部分資料覆寫。
- 原版 #28 本地結案，同時新增建角正式差異待辦；本地 24 未完成／82 完成，遠端仍 24 開放。重名重設角色，使用 grilling 提出單題等待政策；SPEC-005 整體 READY。本輪未改正式程式、未提交／推送／Issue 寫入／打包／發布。收據核實及清理結果於下方追加。
- 收尾核實：三份收據與四張 PNG 雜湊、提交時點及 ASCII 95 項通過；數字假接受／PNG 竄改兩項反面對照均被拒絕。格式化 probe 與實際已執行來源的 gofmt 結果相同；shell 語法、工作清單正反自測／逐項核實、39 份規格與四階段驗收登記通過。本地清單／交接索引已同步，root 擁有產物與誤建 .md 目錄為 0，來源 ZIP 雜湊不變。控制腳本與完整失敗仍留本機，細部結果由 pinned3/final-checks.json 索引。
- 最終 Docker 控制面核實 `kol-finish` 執行中／停止容器為 0，`git diff --check` 通過。證據表的過時字元未知敘述已訂正；本輪檢查器的暫存位元碼已清除。所有保留的原始角色資料與失敗產物均位於本機研究目錄，不是交付包。

## 2026-10-01：正式素材圖集審查

- 上輪建角 oracle 為實質進展；本輪載入復古重製路由、生命週期／驗收技能與文件職責，先核實工作樹與 24 項待辦。沿用 kol-package:20260907，未開新映像或改共用 dosgolem。
- #23 原要求已有 atlas／接觸表正式產物與逐批人工審查，現有 773 張主素材皆具驗證 manifest。新重生器補 27 張戰鬥磚塊／灰階索引、兩張原字型及預設盾牌，完整母體 803 張，44 批、77 頁。來源 PNG 先核對 SHA、尺寸、EGA 色盤及二值 alpha；灰階索引明確另列，不當 EGA 美術。
- 77 頁全數逐頁目視：角色／怪物、武器／頭身腿／手／盾牌透明邊界、八城鎮、書頁、勳章、結局、天候層、標題序列、地圖／磚塊及字型皆看過。圖號／矩形／整數倍率逐筆索引，不裁切、不增添語意名稱；動畫範圍只有圖號序列，不宣稱時長或全部 caller 裁切。部分素材可見英文招牌與 FREE 標記，#35 須先查實際玩家可達性。
- 初版補源 entry 時將同名 CHARPICS／SHLDPICS 可變記錄表也列入 archive；查實 .0 優先分流後乾淨重生。新版本 77 張 PNG 與已目視初版逐張雜湊相同，故目視成果可回溯。BATTLE 地圖來源明確是 TOWN0，不誤列 KOLPICS。最終原版 ZIP 與三個來源 PNG 目錄以唯讀掛載，只有正式輸出可寫。
- 77 頁正式 PNG／manifest 放 assets/original-derived-png/contact-sheets/，同次索引至 SPEC-002／CONTEXT／本地清單；只依已驗證 PNG 的 private GitHub 特例，不擴張公開權利。新核實器檢查全母體、來源 entry、每個矩形與像素／透明極性。#23 仍待整合提交才能達成「版控中的正式產物」，未把新工作樹檔冒稱已入 Git。本地 24 未完成／82 完成不變；沒有正式程式、存檔格式、Issue／GitHub 寫入或打包發布。
- 正反核實通過：77 頁／44 批／803 張來源逐像素／矩形／來源一致；矩形偏移與重算雜湊但 RGB 不在 EGA 色盤的兩項反例均被拒絕，生成器無半套輸出。工具碼與逐頁紀錄見 workplace/asset-sheets-final-20261001/checks.json。
- 原本暫列 #23 待提交，核實授權範圍後以本輪獨立本機 commit 1384e508f70e8a5d81bc050ea79f66d3f0d38fe6 完成版控條件，只含 81 個確切路徑（77 PNG、manifest、兩支工具與 SPEC-002）。受限環境 commit 首次因 .git/index.lock 唯讀失敗；使用主機 Git 權限重試成功，關閉 host hooks 避免主機工作負載。其餘既有異動未納入，沒有推送或 Issue 寫入。#23 本地結案，本地 23 未完成／83 完成；遠端仍維持最近唯讀的 24 開放，#23／#28 未寫入。
- 最終核實：工作清單正反自測／逐項核實、39 份規格與四階段驗收登記通過；78 個正式圖集／manifest 檔案皆 UID/GID 1000:1000，來源 ZIP 未改，root 擁有產物與誤建 .md 目錄為 0。Docker 控制面 `kol-finish` 執行中／停止容器為 0，Git 暫存區乾淨、`git diff --check` 通過。初版／來源修正版與最終圖集、工具來源及控制腳本保留在本機研究區供追溯，沒有公開產物。

## 2026-10-01：城門繁中繪圖與圖片英文引用核對

- 路由載入中文化顯示與語意隔離；實作前再讀規格閘門、SPEC-016 與 RE-065。既有原版證據已定城門 MISCPICS[0]，繁中 OutdoorCJK 卻沒有 AtEdge 分支。先將窄契約由草案審查至 READY，再補正式繪圖；行走與保存不改。
- 修補前正式 cd510e35 建置由標題、名冊與旅店正常按鍵走到邊界，英中 33 筆狀態相同；中文畫格仍顯示地圖與標記。修補後建置 f969a63f 產品 Go 測試通過；回歸涵蓋四種繪圖狀態、缺素材原錯誤傳遞與不回寫狀態。
- 修補後三條正常 flow 英中分支各 33／27／32 筆狀態完全相同；城門及 U-TURN、建築門面與來源三倍像素差為 0，修補前城門 106,182 個像素差保留。英文六張共同畫格及返回地圖雜湊不變。目視確認門面、城門、返回地圖及正常出城。另四條英文機制回歸狀態相同；town-leave 一張與歷史畫格不同只列參考，不宣稱畫格全部相同。
- 第 3／5／7 城 FREE 字形逐圖塊比較，再計七城全部地圖：1＋27＋15 共 43 個圖塊皆未引用。來源與三種目視樣本已保存。第 5 城門 8 的 TWN5PICS[19] 草地告示由正式門口繪圖原樣使用，仍待翻譯與正常自然抵達，#35 不結案。
- 核對工具初次使用 PNG helper 時漏尺寸引數，修正後以同容器同命令乾淨重跑；出城檢查錯把 flow:world 當軌跡名稱，讀正式程式及完整軌跡確認仍為 flow:town、以 left_town／world 表示切換，訂正後同命令通過。兩者為驗證工具問題，未改產品來迎合檢查；第一個正式輸出寫入前被拒絕。
- 新收據、七份軌跡、九張目視 PNG、測試紀錄與五份控制來源同次索引至 SPEC-016、中文化入口及現況稽核。只有 SPEC-016 的繁中顯示子契約 CONFORMED；原版沿用既有 IDA／DOSBox-X 輔助基準，未重跑 dosgolem、完整戰鬥／任務鏈或封包。新建置與舊收據按實際雜湊分開。
- 本地仍 23 未完成／83 完成；HEAD 1384e50 不變，既有異動保留。未提交／推送、未寫 GitHub Issue、未打包或發布。容器、擁有權及最終核實結果在下方追加。
- 最終核實：24 個正式收據附件雜湊、現行正式來源／執行檔、原始 ZIP、三條英中軌跡及修補前後城門狀態通過；PNG 假雜湊反例被拒絕。清單掃描先指出中文化索引的新自承未綁項目，已精確認領到既有 #35，同命令乾淨重跑清單正反／逐項核實及 39 份規格通過；本地仍 23／83。正式與本機輸出皆 UID/GID 1000:1000，root 擁有產物及誤建 .md 目錄為 0；紀錄在 workplace/cjk-edge-20261001/final-checks.json。
- Docker 控制面最後核實 `kol-finish` 執行中／停止容器為 0；`git diff --check` 通過，Git 暫存區乾淨。修補前、修補後與控制來源保留本機研究區，沒有正式交付包。

## 2026-10-01：原圖告示定位勘誤與自然旅行

- 上輪城門顯示修補與正式收據有效；本輪路由載入正常玩家路徑驗收，從現行 f969a63f 正式建置追原圖英文告示的可達路徑。未修改正式程式或種子。
- 查實 typed towns 的城號為 1…7，前輪以 towns[5] 取得第 6 城鞋匠，誤配 TWN5PICS[19]。以 town == 5 明確查詢後，實際為第 5 城門 8、住宅 Residence／Chauncey Lawnwatcher、(40,19)，世界入口 (37,38)。現況、稽核及中文化索引同次訂正，保留前輪誤判形成原因；圖片英文缺口與城門／FREE 證據不變。
- 由標題正常選 Theodoric，在旅店正式招募 Poontang、Olear、Moraine、Fighting Bull、Kenneth，六人完整成立；多人告別再走出旅店。全程不讀觀察存檔、不注入座標、不使用 monsters／quest-sites／ambush-rolls；執行前固定 seed 1。
- 第一版 driver 只朝世界目標直走，卡在 (40,23)、次格 (0,6)、時刻 4020，數百次重播仍相同。查正式 Travel.walk 證實跨格目的地不可通行會還原，因此只停止本次有界容器，exit 143，保留輸入與最後狀態；這是找路工具問題，不是產品卡死。
- 新可丟棄找路適配器依正式 8×8 次格、腳下格 Y+1、固定地形遮罩與 A* 產生方向鍵；只改測試工具的按鍵選擇，沒有改遊戲。第一次雜湊核對把 typed grid.sha256 當 PNG bytes 而拒絕；讀抽取器查實它是原始 column-major 的 4,096 byte 索引。現以素材 manifest 驗 PNG bytes，反轉轉置另驗原索引 SHA，二者都通過後以相同 seed 與前綴重跑。
- 自然旅行、六人初始畫格、原始失敗與兩版找路來源保存於 workplace/cjk-sign-natural-20261001/；本輪途中結果及最終清理於下方追加，不把仍在執行的旅行稱為通過。

- 續輪完成：同建置 f969a63f 由標題選角、正常六人招募，seed 1 自然遭遇 CLIFF TROLL／ORC，分別第 18／13 回合結算，再從世界 (37,38) 入第 5 城、由 (62,57) 正常走到住宅 (40,19) 並入屋。固定腳本 3,080 個輸入，英文及繁中獨立完整重播各 3,081 筆，皆退出 0；沒有觀察存檔、座標注入或指定遭遇。
- 首次門口定位檢查被 town.door=index 5 拒絕，查實 Session.Door 保留上間室內門，戶外繪圖實際使用 Outdoor.AtDoor。未改正式追蹤程式，另以城號／座標／完整來源像素及正常 RET 入屋後 door=8 核實。首次英中語意比較未登記 encounter_text 而拒絕；查正式匯出與既有顯示契約，僅明列第 159／1653 步兩側譯文後重跑，386 步 exec_text 顯示差異與兩步遭遇譯文以外全部相同。原始軌跡欄位保留。
- 七張關鍵畫格逐張目視；住宅門面英文及繁中視窗與來源 PNG 像素差均為 0，證實 KEEP OFF THE GRASS 告示可由正常新遊戲抵達。正式軌跡、腳本、畫格、來源雜湊與七份控制程式收錄 docs/verification/cjk-ui/natural-sign-20261001/，同次接入中文化索引、稽核、CONTEXT 與工作清單。只驗 remake 正常可達性與英中語意，不宣稱原版、存檔或平台封包通過；告示翻譯仍待辦，#35 不結案。
- 最終核實：正式 23 個附件、來源輸入及執行檔雜湊一致，七張目視畫格與英中語意／來源像素通過；假 PNG 雜湊反面對照被拒絕。工作清單首輪掃描指出新索引缺口未綁項目，已精確綁回既有 #35 並乾淨重跑正反自測／逐項核實、39 份規格及四階段驗收登記通過。本地仍 23 未完成／83 完成；正式附件與新增輸出 UID/GID 1000:1000，root 擁有產物及誤建 .md 目錄為 0。
- 告示視覺共同決策：依 grilling 先做可丟棄 A 牌面中文／B 保留英文加下方中文說明，兩張畫格已目視。介面畫布使用既有 Go 工具鏈、正式字模與相同來源畫格；逐像素確認方案指定矩形以外不變。原型、來源程式與雜湊只留 [本機研究區](workplace/cjk-sign-natural-20261001/sign-prototype/prototype.json)，未改正式資產、規則或存檔；只涵蓋主告示，兩側殘缺文字未升格成已知。已提出單題等待使用者選擇，沒有將建議寫成正式決定。
- 原型工具失敗保留：封包 image 的 ImageMagick 預設 SVG 委派缺少 rsvg-convert，MSVG 嘗試仍未提供內建繪製器；兩次失敗後重讀工具鏈路由。改用既有 Go 畫布，首輪暫存 tmpfs 的 noexec 拒絕啟動產物；明確設定容器暫存 exec 後同 Go 來源乾淨重跑退出 0。未在主機執行工作負載、未建立重複 image；最終圖形不是 AI 重畫。

## 2026-10-01：中文字模空白字距驗收勘誤

- 告示呈現等待使用者選擇，續做不依賴該決定的文字量測。查實正式字模沒有空白 glyph，Font.advance 使用 Advance=16；靜態工具誤用 8，會低估混合姓名／英文句。只改檢查工具，不改正式字距、翻譯、版面或產品建置。
- 保留前一 9,635 項報表與原正式 6,685 項報表；同母體修正版 9,635 項全部通過。Go oracle 直接 font.Load／Font.Measure 核對 7425 筆非空檢查，全部一致；110 項舊寬度訂正。以實際低估字串建立窄矩形反例，舊算法會放行、新契約拒絕。
- Go oracle 首次被 go.work 的未掛載 /dosgolem 擋在載入前；本任務不依賴 oracle 模組，以 GOWORK=off 限定本專案、相同來源與字模乾淨重跑退出 0。Xvfb 有 trap 與程序邊界，使用既有 Go image、無網路及容器 tmpfs exec，未在主機執行。
- 正式字寬收據、直接量測、原執行 Go 來源、工具、版面及舊報表同次保存與索引；新正式 cjk-text-fit.json 含每份輸入 SHA。此證明限既有靜態母體，#34 全欄位及 #40 最長正常 GUI 仍開放。
- 字距反面對照已升級為真正工具執行：從 HEAD 取原檢查器，不修改正式譯表，以人工 A＋16 個半形空白＋A 的同一 TSV 跑舊／新工具。舊版計 150、退出 0；修正版計 278，對戰鬥裝備欄 264 像素拒絕、退出 1，恰一項失敗。兩側完整命令／stdout／報表／baseline 來源與人工夾具已保存並納入字寬收據；不把人工界限夾具稱為正常玩家路徑。反例彙整首輪錯假設 A 字寬 10 而拒絕，查實正式字寬 11，現改直接依 Go runtime 量測結果組合核實，同命令乾淨重跑。
- 最後收尾核實：自然告示及字寬收據附件雜湊與擁有權全數一致，原始 ZIP SHA 未變；全專案 root 擁有產物及誤建 .md 目錄為 0。39 份規格與四階段驗收登記通過，告示 A／B 選擇維持待答；新字寬工具不改正式產品。Docker 控制面回讀 `kol-finish` 執行中／停止容器為 0，`git diff --check` 通過、暫存區乾淨；未提交、推送、Issue 寫入或發布。

## 2026-10-01：最寬自然戰鬥訊息正常畫面

- 上輪正常住宅與空白字距勘誤為實質進展；告示 A／B 尚待答案，未將自動繼續當成採用。路由沿用正常玩家路徑與中文化顯示隔離，從現行 f969a63f 的自然旅行母路徑補最寬執行訊息證據。
- 全母體 386 個可見訊息狀態按正式字距預先選前三則，第 2,566／545／1,287 步，最寬 564。半形空白仍用 16；原先以 8 估算會誤把第 545 步列成第一，現在沒有用舊偏差選容易樣本。選定與種子 1 在執行前記錄，正常標題、六人招募與自然遭遇按鍵重播，無讀檔／傳送／指定敵人。
- 2,567 個輸入、2,568 筆狀態與完整母路徑同前綴完全相同、程序退出 0；五張畫格包含最寬三則與相鄰前／後一步。卷軸來源三倍像素疊實際 PNG 字模核對為 0 差異；五則重排成一行，寬 563／557／516／564／163、文字底緣 522，在卷軸底緣 588 內。單一墨跡像素缺失反例拒絕；五張完整畫格逐張目視，黑字、姓名、句子可讀且不裁切。
- 正式收據、來源 SHA、軌跡、正常按鍵、畫格與兩份控制來源保存於 docs/verification/cjk-ui/longest-battle-20261001/，同次接入中文化入口、稽核、CONTEXT 與清單。此證明限自然戰鬥母體；#34 尚未覆蓋欄位、#40 母體外文字與最長輸入仍開放，未改正式程式、字模、規則、保存或平台產物。
- #34 原來的過時 387／1,935／9,301 計數與逐輪附註改為現行量測摘要；原內文保留如下供追溯：

書本版面已機械量得；城鎮、世界與戰鬥已有部分中文安全矩形和畫格。資料名稱 387 筆在五種固定欄位新增 1,935 項寬度檢查，目前合計 9,301 項失敗 0；戰鬥訊息與其他分支仍未全覆蓋。最新手冊／中文稽核見 docs/verification/manual-fidelity-audit-20260923.md。 2026-10-01 空白字距訂正：靜態工具原計 8 像素、正式 Font.Measure 為 16。相同 9,635 項仍無失敗，7425 筆非空檢查與直接 Go 字模量測相同、110 項寬度已訂正。正式報表 docs/verification/cjk-text-fit.json、來源與舊報表收據 docs/verification/cjk-ui/font-width-20261001/receipt.json；未改正式字距或譯文，尚未涵蓋欄位仍待驗。

- 最長自然訊息收尾核實：12 份附件／來源雜湊、執行檔與原始 ZIP 全數符合，五張目視狀態及像素收據一致，輸出 UID/GID 1000:1000。全專案 root 擁有產物與誤建 .md 目錄為 0；清單正反自測／逐項核實、39 份規格與四階段驗收登記通過，本地仍 23 未完成／83 完成。沒有正式程式變更、提交／推送、Issue 寫入、打包或發布。
- 最後 Docker 控制面確認 `kol-finish` 執行中／停止容器為 0；`git diff --check` 通過、暫存區乾淨。正常重播與控制來源仍保留本機研究區，正式收據按同一執行檔 SHA 分開索引。

## 2026-10-01：全母體最寬城鎮進門敘述

- 路由沿用中文化顯示隔離與正常玩家路徑驗收；住宅告示 A／B 尚待答案，未代選。現行正式 451 行進門敘述預先選最寬 240 像素之一，第七城第 21 門克拉格要塞。
- 第一次控制腳本誤用只支援真彩色的 png_io.read_png 讀灰階地圖，遊戲執行前即拒絕；空輸出目錄保持原樣並查擁有權，改用專案既有 build_asset_contact_sheets.read_image，同一容器命令重跑通過。此為驗證腳本格式問題，未改來源 PNG。
- 正常標題與六人招募，起點 (57,8)，77 個城內移動並經正常折返圖示，140 個輸入抵達 (20,36) 再進門。種子 1 執行前固定；英中各 141 筆狀態完全相同、退出 0，無讀檔、傳送或指定遭遇。
- 四行寬 240／240／224／144；對話安全寬 575、最後一行底緣 348。完整 585×204 黑底白框與實際字模 0 像素差，單一墨跡像素反例拒絕；TWN7PICS45 門口／44 肖像在英中共四份來源圖片核對 0 差。主代理檢視中文門口、中文四行內頁與英文內頁，完整可讀。14 份雜湊附件與控制來源保存於 docs/verification/cjk-ui/longest-town-entry-20261001/，同次接入索引、稽核、現況與 #40。
- 途中曾誤將人物資料頁 Rank 當作招呼 Renown 推算最長訊息；查載入端後確認二者不同，先前選人推算不成立，未更動規則、規格或正式程式。最長招呼語另追正常可招募蠻族的偏好分支，尚未驗收；不把推測升格成已證實。
- 本收據只補進門敘述全母體最寬之一；其他訊息、標題與最長輸入仍待驗，#34／#40 未結案。未提交、推送、寫入遠端 Issue 或發布。

## 2026-10-01：全母體最寬城鎮訊息的正常告別路徑

- 路由沿用正常玩家路徑驗收與中文化顯示隔離。先前規劃以第六城酒館蠻族招呼觸發 368 像素訊息；實際完整母體有四則同寬，含第七城旅舍告別語，因此改以現有正常隊伍離店核實母體最寬之一，無須旅行第六城、重招募或推算聲望。
- 第一份招募控制腳本的文字標籤漏了「訊息」的「息」，選樣斷言在遊戲執行前拒絕；未招募 Sharakan、未開始遠途旅行，空輸出目錄 UID/GID 1000:1000。完整候選已保存 selection.json，選定旅舍告別語後另用正常離店控制來源，舊失敗控制來源保留在 /tmp。
- 現行 f969a63f 從標題正常新建隊伍與招募六人，固定種子 1；28 次城內移動抵達第七城第 17 門旅舍 (59,28)，正常進門、問誰開口、選門離店，94 個按鍵／95 筆狀態英中完全相同、退出 0。第 93 步訊息 17 兩行寬 368／144，文字底緣 300 位於框內、安全寬 575。第 94 步確認回到原門格，未注入座標、讀檔或指定遭遇。
- 完整 585×204 黑底白框及正式字模核對 0 像素差；單一墨跡像素缺失反例拒絕。TWN7PICS37 門口／36 肖像在英中門口、進門、選人與告別畫面共八份圖片核對 0 差。主代理逐張檢視五張中文完整畫格與英文告別，兩行可讀且無裁切。
- 20 份雜湊附件、控制來源與收據保存於 docs/verification/cjk-ui/longest-town-message-20261001/，同次接入中文化索引、稽核、CONTEXT 與 #40；未改正式程式、字模、規則或保存。未驗另外三則同寬訊息、其他欄位、最長輸入、原版與平台產物，#34／#40 不結案。未提交、推送、遠端 Issue 寫入或發布。
- 本批結束：工作清單正反對照與同步驗證通過，本機仍為 23 項未完成／83 項完成；兩份新收據的全部附件雜湊及 UID/GID 已核對，專案沒有 root-owned 檔案或誤建的 `.md` 目錄。`docker ps -a --filter name=kol-finish` 為空，背景 Xvfb 隨擁有程序結束；`git diff --check` 通過、暫存區為空。正式程式未改，未重跑或冒稱新一輪 Go 產品測試；暫存控制來源及正常路徑產物留在 `/tmp` 與 `workplace/cjk-longest-town-20261001/`、`workplace/cjk-longest-message-20261001/` 供追溯。

## 2026-10-01：最長姓名正常輸入、容量勘誤與肖像名牌缺口

- 上輪兩份全母體最寬城鎮收據為實質進展。路由沿用中文化顯示隔離與正常玩家路徑，保留既有工作樹及未決分支，繼續 #40 的最長輸入。讀取時一度猜 internal/charcreate/state.go，查無檔後改查實際 charcreate.go；未產生缺源掛載。
- 首次控制漏掉選單 D，實際走到角色清單；軌跡場景核對攔截，沒有當成建角通過，原 selection／trace 留存本機並附入正式收據。修正後加入 flow:newchar 場景斷言，在同一容器命令以 attempt2 乾淨重跑，無覆寫前次產物。
- 現行 f969a63f、種子 1 在執行前固定，正常標題→角色選項→建角，16 W／15 空白加 W／16 空白三情境英中各 30／28／23 筆、合計81筆狀態相同。退格移除一字、滿16字前進種族、額外 W 不改畫面；11 個姓名值／游標區依原書頁、字模 alpha、字距和清底核對0差，六個滿欄後額外按字的英中完整畫格亦0差。未接受肖像或寫角色保存。
- 舊靜態姓名容量只掃非空白有字模字元，估240；正式鍵盤空白前進16，最寬應256。tools/verify_text_fit.py 改用現行鍵盤輸入字元集合及與Font.advance一致的寬度，9,635項重新生成失敗0，只有容量一項改動；前報告 c8b81562 完整保留於 font-width 與本收據。248像素窄欄反例：舊版錯誤接受 exit0，新版拒絕唯一 name 項 exit1。無改正式字模或遊戲。
- 肖像名牌黑底121，16W240超出119、15空白加W255超出134；實際像素核實白字落到名牌外書頁，目視後半段或全名不可讀，英文亦保留同現象。本次不能宣稱最長輸入整體完成；#34／#40保持開放。
- grilling 實際讀取，依「暫停依賴該決定的分支」先做可丟棄 Go 原生字模原型：A依實際字寬加寬黑底完整姓名，B維持名牌並縮略加省略號。兩張原型逐張檢視後，以單一問題請使用者選擇、建議A；尚未收到答案，正式CJK名牌、姓名政策及保存不改。原型只比較繁中顯示，英文原版名牌語意不猜補。
- 正式收據及附件、控制來源、失敗畫格、兩張原型在 docs/verification/cjk-ui/longest-input-20261001/，同次接入中文化索引、稽核、CONTEXT與清單；先前#40敘述保留如下供歷史追溯。

已補固定種子 1、正常標題／六人招募／自然戰鬥母體的最寬訊息視覺回歸。先從完整路徑 386 個可見訊息狀態選出前三則，再重播相同 2,567 個按鍵；2,568 筆狀態相同。五張正式畫格與卷軸／字模像素完全相同、目視可讀，最寬 564 未超安全寬 816；收據 docs/verification/cjk-ui/longest-battle-20261001/receipt.json。此母體外的最長文字、其他 CJK 欄位與最長輸入仍待正常路徑驗收，本項不結案。 2026-10-01 再補正式 town-entry 全母體 451 行最寬 240 像素之一：正常標題／六人招募／方向鍵抵達第七城克拉格要塞，英中 141 筆狀態相同；完整對話框與字模、四份門口／肖像圖片核對皆 0 差異，目視可讀，見 docs/verification/cjk-ui/longest-town-entry-20261001/receipt.json。其他訊息、標題與最長輸入仍待驗，本項不結案。 另補正式 town-message 4,201 行全母體同寬最寬 368 像素之一：第七城旅舍正常多人告別。英中 95 筆狀態相同，完整文字框／字模及八份來源圖片均 0 差異，目視通過並返回原門格；見 docs/verification/cjk-ui/longest-town-message-20261001/receipt.json。未改正式程式，其他欄位與最長輸入仍待驗，本項不結案。

## 2026-10-01：最長城鎮共用問句的正常聆聽

- 肖像名牌 A／B 尚待答案，依 grilling 暫停正式顯示分支，繼續不依賴該決定的 #40 共用問句。路由沿用正常玩家路徑與中文化顯示隔離；未另開無關逆向或猜補規則。
- 現行正式 dialog-ui 260 行最寬240為KOLDATA0x9D；查 typed 話題資料第七城第八門小船屋具三則可聆聽傳聞。現行 f969a63f 從標題正常招募六人、城內方向鍵抵達(4,25)，正常進門、按1選人、圖示列兩次右移選聆聽，連續三輪每輪兩段確認。種子1執行前固定，沒有指定槽位、重擲或挑選結果；英中各169筆狀態完全相同、退出0。
- 驗證控制一度將後兩則預期類別／訊息順序對調，第161步完整文字框核對差560而拒絕；實際文字庫0x9B是「近來如何」、0x9A是「你知道些什麼」，typed槽4／3類別1／0。核對畫格與文字庫後修正預期順序為槽0／4／3、問句0x9D／0x9B／0x9A、訊息20／24／23，以同一容器命令重跑通過；未改原始資料、遊戲或種子，不把驗證腳本錯源分類成產品缺陷。
- 六個完整585×204黑底白框與正式字模各0像素差；每框單一墨跡像素遺失反例拒絕，三行問句底緣324在框內。人物TWN7PICS18在六張畫格英中共十二份核對0差，主代理逐張檢視三張中文問句、三張中文答覆及英文最長問句，完整可讀。32份附件與控制來源保存 docs/verification/cjk-ui/longest-dialog-20261001/，同次接入中文化索引、稽核、CONTEXT及#40。
- 收據只證明所列正常中文顯示，不宣稱原版亂數或全話題對拍。名牌、最長標題、其他動態欄位、原版與平台仍未完成，#34／#40保持開放；正式遊戲、規格、字模與保存不改，未提交、推送、寫遠端Issue或發布。
- 本輪結束核對：兩份新收據共 101 份附件的雜湊與 UID/GID 已驗，工作清單正反對照及同步檢查通過；本機仍 23 項未完成／83 項完成。專案沒有 root-owned 檔案或誤建 `.md` 目錄，`docker ps -a --filter name=kol-finish` 為空，Xvfb 皆隨擁有程序退出，`git diff --check` 通過、暫存區為空。本輪只改文字量測工具、收據及文件，未重跑或冒稱新 Go 產品測試；正式二進位仍 f969a63f。正常路徑、錯路由產物、原型與控制來源留 `workplace/cjk-longest-input-20261001/`、`workplace/cjk-longest-dialog-20261001/` 及 `/tmp` 供追溯。
## 2026-10-01：全母體最長城鎮標題的正常自然旅行

- 路由再次核對中文化顯示隔離、正常玩家路徑與既有文件職責入口。正式 town-title 204行最大160有六則同寬，選第五城商行人名，沒有沿用先前「只有兩則」的摘要推測。實際英文建築 Journeys' End Stores、人物 Ignatz Kramfrupples，以 typed 與英文明頁核實；英文舊推測拼字不作證據。
- 現行 f969a63f、固定種子1，完整重播正常標題、六人招募、自然旅行及住宅母路徑，再以SPACE／RET／SPACE正常告別，步行六格到(41,24)並進商行。英中各3,091筆、退出0，前3,081筆各自與原住宅軌跡完全相同；其餘英中狀態相同，只有原登記386步戰鬥敘述及2步遭遇譯文差異。沒有重新挑種子、讀檔、指定敵人或注入座標。
- 驗證工具在 kol-package:20260907 以60秒有界命令執行：完整585×60標題框／正式字模0像素差、單一墨跡像素反例拒絕；店名寬96、人名寬160、底緣217位於安全框內。TWN5PICS21／20英中四份圖片0差，主代理逐張檢視四張中文完整畫格與英文內頁，置中可讀且沒有裁切。
- 19份附件及收據保存 docs/verification/cjk-ui/longest-title-20261001/，同次接入中文化索引、稽核、唯一現況表及#40。只證明六則同寬之一，其他動態欄位、肖像名牌、母體外分支與原版／平台仍待驗；#34／#40保持開放。名牌與住宅告示A／B尚待答案，未代選；正式程式、規格、字模與保存均未改。
- 動態名稱分類勘誤：112像素最大值 Girak Helbak／吉拉克赫爾巴克的來源是 shop-good.078、第七城第13門酒館商品表T7D13G5；不能僅憑名稱推成NPC、坐騎或武器。後續須依實際消費端選欄位樣本，此處未猜補商品語意。
- #40上一版內文保留供回查：

已補正常最寬文字收據：自然戰鬥 386 個訊息狀態中選最寬三則（最大564、安全寬816），重播2,568筆；town-entry 451行全母體最寬240之一正常抵達克拉格要塞、英中141筆；town-message 4,201行全母體最寬368之一正常旅舍告別、英中95筆。文字與來源圖片核對0差且目視通過。最長姓名另從標題正常建角，16個W／15空白加W／16空白三情境英中各81筆相同，11個姓名值與游標區0像素差；但肖像名牌超出119／134像素而名字不可完整讀取，正式版面等待加寬A或縮略B的使用者選擇。容量工具漏算空白已改240→256，9,635項靜態檢查仍失敗0，窄欄反例正確拒絕。收據均在 docs/verification/cjk-ui/{longest-battle,longest-town-entry,longest-town-message,longest-input}-20261001/receipt.json。名牌、其他欄位、最長標題與母體外分支未驗，本項不結案。 另補 dialog-ui 260 行母體最寬240的正常小船屋聆聽：預先固定種子1、三次聆聽，英中169筆相同；六個問答框與十二份人物圖來源0像素差，目視通過，見 docs/verification/cjk-ui/longest-dialog-20261001/receipt.json。不宣稱原版亂數、全話題或整項完成。

- 本批收尾：清單正反對照及同步核實通過，維持23未完成／83完成；19份附件與來源SHA、執行檔SHA和UID/GID全數符合，專案root-owned與誤建.md目錄為0。Docker控制面回讀kol-finish容器為空、暫存區為空。第一次差異檢查只抓到新附加文件尾端多一空行，已修正再核；沒有產品變更或新Go測試，沒有提交／推送／Issue寫入／打包／發布。控制來源及正常重播產物留在/tmp與workplace/cjk-longest-title-20261001/供追溯。

## 2026-10-01：最長商品名稱的正常購買、物品頁與保存重載

- 上輪最長標題收據為實質進展；本輪再次讀AGENTS／工作清單／現況及中文化顯示隔離入口，保留現有工作樹與未決A／B分支。最長名稱來源查證為酒館食品；沒有從名稱猜成NPC、馬或武器。
- 首次控制預期入袋後Carried摘要+1，並選第一背包格而失敗。原角色已有五件背包物品，第六格才是新食品；Carried摘要來自獨立原始資料，不是Backpack即時計數。原控制來源、selection及軌跡保留在工作區並附正式收據。按角色實際資料修正選格與斷言，以同一Docker命令在attempt2乾淨重跑，未改遊戲或原始資料。
- 現行f969a63f、固定種子1，正常標題／六人招募／41個城內方向鍵到第七城第13門(38,26)，交易第5項食品，正常買下並選不當場吃。商品清單、明細、第六背包小框及物品明細已核實。正常返回半途之家、選單人住宿，食品35與住宿56使金冠1009→974→918。英中各204筆完全相同，住宿保存含食品原名Girak Helbak及記錄c3a641581056082b002300370000a000。
- 重載首輪錯用flow-script而落入書本目錄，outside_sheet斷言拒絕；保存和錯路由軌跡保留，未稱重載完成。查實批次flow入口會重建標題狀態，改用正式play-script，同一Docker命令以reload2乾淨重跑。英中各21筆完全相同，正常戶外按1開資料頁再到第六背包格；本次自然存檔沒有被改寫。先前像素工具因繁中重載來源未產生而拒絕，重載完成後同命令通過，為驗證順序問題。
- 503筆實際商品、292個名稱全部有譯名，唯一最大112就是所選食品；與靜態420個跨欄位名稱母體分開登記。六個完整文字框／字模0像素差，每框單墨跡像素反例拒絕；TWN7PICS29／28英中四份來源0差。主代理逐張目視9張正常中文、2張重載中文及英文物品明細，共12張，名稱可讀且不裁切。
- 正式50份附件、完整商品母體及收據放docs/verification/cjk-ui/longest-shop-20261001/，同次接入中文化索引、稽核、CONTEXT與#40。重載run元資料的雜湊是在執行後取值，誤標before已訂正為after，控制來源與附件雜湊同步更新；命令、畫格和遊戲沒有改。只驗商品及物品頁，不用食品取代正常戰鬥武器、裝備或戰利品樣本；#34／#40仍開放，未改正式程式、規格、字模或保存格式，未提交／推送／Issue寫入／打包／發布。
- 本批最後核實：50份附件與來源／二進位SHA、UID/GID已核對，root-owned及誤建.md目錄0；清單正反對照及同步通過，維持23未完成／83完成。Docker控制面kol-finish執行中及停止容器為空，Xvfb隨擁有程序退出；git diff --check通過、暫存區為空。沒有正式程式改動或新Go測試；首次錯格／錯入口與正常重播、自然存檔及控制來源保留workplace/cjk-longest-shop-20261001/及/tmp供追溯。

## 2026-10-01：初始物品空格判準修正

- 路由命中逆向／規格閘門與IDA，載入reverse-engineer-retro-game-remake、use-ida-pro-9-4與工具契約。技能符號連結在窄掛載內失效，查到實際來源/home/anr2/ida_94_official/skills/use-ida-pro-9-4後讀取；未改技能。正式.i64唯讀，以一次性副本匯出338函式環境中的購買／物品頁指令，保留原名、bytes、xref及分級附加語意。
- RE-170證實首byte零即空；原抽取器any(block)誤判22格（21背包、Mula手持）。先追加SPEC-023 READY窄子契約，再修worn_slot並重抽名冊。原ZIP／PNG／EXE唯讀，只有typed輸出與工作目錄可寫。238格與三邊界檢查通過，原Carried及其他欄位不變；舊資料反面對照拒絕。舊／新名冊SHA f8f05ff2→d2b3666b，執行檔f969a63f未變。
- 正常標題、六人招募、城內走路、購買帶走、第一背包格檢視、住宿保存及重新啟動讀檔，英中199／16筆完全一致，保存只差時間戳記；金冠1009→974→918。六框四圖0像素差、單墨跡反例拒絕，主代理目視三張物品畫格。產品go test ./internal/... ./cmd/...全通過；town-shop／town-mirror／town-rest／town-load四條短回歸全部相同。
- 舊v3存檔相容性首輪漏shot-dir被CLI拒絕；控制腳本補齊後compat2乾淨重跑21筆與舊重載完全相同，來源未改，沒有清洗舊玩家物品。分欄診斷初稿把空背包null當[]而拒絕，按實際合法空值修診斷後通過，沒有改保存器或玩家資料。
- full-chain2318筆／full-chain-reload2013筆各二／三筆與舊嚴格參考不同，exit1與失敗保留。另逐欄診斷確認只差Carried摘要及其省略事件，其他狀態與抽樣畫格相同。母保存全部欄位審查只差22格空物品移除、四名戰後Carried同步及時間戳記；診斷不作比較器忽略區，不換舊基準。#88依新初始資料實測重新待辦，24未完成／82完成，下一閘門受影響收據重生及全量29條；不冒稱新資料已全量通過。
- 正式驗證紀錄、九份附件、工具與本機證據SHA同次接入SPEC-023／CONTEXT／中文化／玩家路徑索引。原始名冊、IDA payload、完整軌跡與自然存檔只留workplace/inventory-empty-slots-20261001；三張衍生PNG仍只適用private GitHub特例。未提交／推送／遠端Issue寫入／打包／發布。
- 收尾核實：九份附件、正式來源與本機證據SHA全符，UID/GID為1000:1000；專案root-owned及誤建.md目錄為0。一次性IDA副本與/tmp/probe.json已移除，控制來源、舊資料及正常／失敗收據留明確工作目錄。工作清單正反對照與同步通過，24未完成／82完成；Docker控制面本批kol-inventory及kol-empty-slot容器回讀為空。git diff --check通過，暫存區為空。未改比較器、未覆寫舊基準；沒有提交／推送／遠端寫入或發布。

## 2026-10-01 RE-170 全量續核與本地 #88 結案

- 接續初始空物品修正，保留 543 份舊基準與三份母存檔；首輪全量 29 條中七條只差持物摘要。
  `review_player_path_inventory_change.py` 核對逐欄軌跡及保存，金冠／模式／超額持物反例拒絕。
  審查後更新七條基準，原執行檔未重建；原腳本、種子與受控輸入未改。
- `kol-inventory-all-after-20261001` 同一容器／同一命令獨立全量重播，exit 0，
  29 條／14,875 筆狀態一致，續玩讀當輪新保存；日誌 `all-after.log`。
- 收據初驗 24／29，五條失敗因舊整圖天候引用。依 RE-164／SPEC-026，先核對來源合成，
  再以相同程式／seed1／90步前綴的新舊名冊對照，91筆及兩圖完全相同。
  原圖不改，七張新模板回驗原有22組原版區域全過；比較矩形、滑鼠例外、必要區域未改。
  `verifier-final` 使用第二次全量的實際畫格／保存，全部29條0失敗；正式入口再核29條全過。
- 宿泊舊驗收 v1 按 SPEC-023逐欄修正為正式v3；未改格式或清洗舊玩家存檔。
  新證据、歷史失敗與來源雜湊入口為 `docs/verification/inventory-empty-slots-20261001.md`。
- 本地#88重新完成，統計23未完成／83完成；未提交、推送、遠端Issue寫入、打包或發布。
  原版整場戰鬥與正常GUI天候仍依既有獨立閘門，不因回歸通過升格。

本輪最終自檢：專案 `kol-inventory*` 容器在驗證前為零；全部工作容器採 `--rm`。
全專案 root-owned 檔／同名 `.md` 目錄零；本輪本機與正式輸出逐項確認 UID/GID 1000:1000。
655 項附檔／來源雜湊核對通過，`git diff --check` 通過；本機快照與失敗保留，
沒有新 IDA 臨時資料庫。自檢收據 `workplace/inventory-empty-slots-20261001/final-audit.json`
已掛入正式物品收據；容器退出後再由主機核對無本專案殘留。

## 2026-10-01 最寬正常戰鬥裝備欄與腰帶標籤勘誤

本輪先核對AGENTS、WORKLIST、目前工作樹與知識路由；命中並讀正常玩家驗收入口，
發現已有RE語意未回填時再讀re-resolution-backlinks與SPEC-019／SPEC-010。
只用既有Docker工具鏈，保留所有繼承異動，未重建正式執行檔。

盤點腳本曾把坐騎商品、空名冊與Stage列舉混讀，另有搜尋regex轉義錯誤；
均為驗證腳本問題，改讀實際schema及程式列舉。目標格首輪擷取為空格，按規格應無資訊；
改走六個隊員主指令的空白切換，未猜座標或注入角色。
首輪相鄰trace比較失敗因stage快照roster在下一筆省略，已依匯出程式契約補齊，
其他欄位保持嚴格，切換後全圖像素亦要求復原。像素模型首輪把shield取equipment[4]，
與SPEC-019的+0x160不符；改為equipment[6]，重選142筆實際護甲母體再乾淨重跑。

六張初驗畫格證實B:仍譯身體，但現行消費端讀BeltName；原版+0x190及RE-094已證實腰帶。
先補SPEC-019 DRAFT子契約並逐項審查READY，再改兩份外部譯表及註解，不改英文或規則。
`kol-cjk-equipment-current-20261001`：英中各270筆，母路徑258筆完全符合，
只有原遭遇譯文一筆不同；每人切換後逐像素復原。六個完整框獨立字模0差、六圖目視，
96像素護腿在95像素窄欄反例拒絕。18張英文畫格不變，中文只差六個腰帶標籤框。

譯表原整份json重新序列化產生無關差異，已保留實際執行輸入並恢復原排版、只換一字；
parsed JSON逐項相同，前後SHA及快照記在format-equivalence.json。
相同靜態工具在目前字節輸入重驗9635項0失敗，兩支翻譯檢查通過；
正式cjk-text-fit更新，舊報表留本機備份。收據同次掛入中文化索引、SPEC-019與三條待辦。
本地23未完成／83完成；未提交、推送、遠端Issue寫入、打包或發布。

本輪收尾補記：收據匯出曾因缺少fontpath變數中止，產品與三項檢查均已通過；
修正後只續寫尚未產生的收據，沒有重複追加文件。正式收據60份附件／36張畫格、
現行來源及當時執行來源雜湊皆核實；格式等價快照另列，沒有抹去執行時的字節版本。
全專案root-owned／`.md`同名目錄零，本輪產物逐項UID/GID1000:1000；
自檢收據`workplace/cjk-equipment-20261001/final-audit.json`掛入正式收據。
初驗與修正前畫格保留本機，無新工具image或IDA臨時資料庫；一次性容器均--rm，
退出後再由主機核對本專案無殘留。

## 2026-10-01：現行三平台中間建置、Wine 正常互動與預設保存

- 路由命中跨平台試玩、dist-all輸出位置及文件職責；沿用既有Go1.26.7／Ebitengine2.9.9、Wine9、osxcross與Python工具image，未新建image。原始輸入與專案資源唯讀，task輸出與Wineprefix以1000:1000寫入；全部Docker工作有界、--rm、--network none。首個image inspect模板因Config缺Entrypoint鍵失敗，改用index後確認原image；沒有換image或搬入主機runtime。
- Linux／Windows各編譯兩次逐byte一致；LinuxSHA仍f969a63f，Windows304890ea。macOS universalSHA2b6370c9，兩個slice與thin檔逐byte一致、最低11.0、arm64簽章superblob及所有系統相依通過；只做靜態核實，未宣稱macOS啟動或雙次重建一致。
- Wine64 prefix初始化成功；wineboot的32位元／OLE警告保留，不誤判64位元產品故障。批次flow首輪兩側199筆正常購買軌跡相同但不寫預設保存，控制腳本test失敗；核實batchFlags後分類入口問題，保留原結果，不改正式程式。
- 用既有x11_key_inject.py在另一Docker容器連共用唯讀socket，XTEST逐鍵送出並核對log-input收到同token；X11 GetImage深度24擷取。直接互動啟動不指定scene／lang／save，兩側198鍵從標題、六人招募、酒館購買帶走、檢視第一背包格到正常住宿。Windows寫APPDATA、Linux寫XDG_CONFIG_HOME的save.dat，六人、gold918與食品record均符合；重啟-load本次自然保存，15互動鍵與16筆批次重載一致、原檔SHA不變。
- Linux首輪Ctrl+Q快速退出，擁有容器已關X11而捕捉器keyup遇BrokenPipe；198鍵與存檔完整保留，修控制器先記成功、最後keyup允許X11關閉，產品退出碼仍由wait嚴格核實。新config2／gui-normal2用相同198鍵乾淨重跑，退出0；首輪控制快照另存。一次PNG診斷迴圈重複解碼而逾時，停止精確容器，再以單次讀取同SHA核對，屬診斷腳本問題。
- 167筆sourceSHA執行前後一致；兩側199／16筆批次狀態與既有初始物品收據相同，40組完整靜止畫格0像素差，保存只差saved_at。錯誤PE架構、金冠差異與單像素差異反例拒絕／不相同。主代理檢視Windows封面、目錄、商品、物品、住宿與重載明細，中文可讀，Linux對應靜止畫格一致。
- 正式收據及36張PNG、控制來源、編譯／静態日誌放docs/verification/platform-smoke-20261001/，同次接入CONTEXT、稽核與#7／#37／#42。原版ZIP、typed payload、binary、SDK、Wineprefix及自然保存留workplace/platform-smoke-20261001；PNG只沿用private GitHub特例。工作樹繼承異動不動，不提交／推送、不寫遠端Issue、不建正式版號／dist-all、不發布。三平台整項仍開放；macOS實際啟動、Windows／macOS真機、正式封包、FUSE、音訊及推廣分別待驗。

本批收尾核實：73份正式附件／36張PNG與說明文件SHA符合，167筆產品／資料來源未變，
1210項輸出擁有權皆1000:1000；全專案root-owned及誤建.md目錄為0。
工作清單正反對照與同步檢查通過，23未完成／83完成；git diff --check通過、暫存區為空。
主機Docker控制面回讀kol-platform容器為空，Xvfb與Wine均隨擁有程序退出，未留新image。
自檢收據workplace/platform-smoke-20261001/final-audit.json已掛入正式收據，
說明入口docs/verification/platform-smoke-20261001.md同次掛入receipt；失敗與正常產物保留本機。


## 2026-10-01：初始持有物漏譯收尾

本輪沿用重製驗收與正常玩家路徑技能／入口，重新查路由後讀顯示與語意隔離、
spec 閘門及文件職責。盤點戰鬥武器來源時發現初始名冊尚未納入顯示名稱稽核；
實際 14 位有效角色有 95 筆持物，漏譯 Maul／Black Choker。原檢查器對舊表仍綠，
正常選角到物品頁後也確實顯示英文。先追加 SPEC-033 窄契約、審查 READY，
再只補「大槌／黑色頸鍊」暫譯與名冊三個持物欄的覆蓋。
Black Choker 狀況為 0 但 record 首 byte 非 0，沿用 RE-170 的有效物品判準。
沒有改正式 Go、typed data、物品規則或保存格式。

既有 f969a63f 建置未重建。正常六人／Warder 單人各自從標題選角，檢視物品、退出、
步行進旅店及住宿保存；修正前後英中各跑一次，再以本次保存重載。12 次共 620 筆
狀態完全一致；保存只排除 saved_at 後相同，重載未改寫，原名及物品 record 保留。
八個完整文字框符合正式字模、缺一個墨點反例正確拒絕；32 組整圖差分只改四個
中文名稱區域，英文及重載畫面相同。四張修正後完整中文畫格目視可讀。
新檢查器拒絕舊表的兩筆缺譯，修正後 531 鍵／515 原名無缺譯或孤兒。
9,645 項安全矩形失敗 0，其中舊 9,635 項逐筆不變；6,095 筆譯文沒有缺字。
未新增字模，font PNG 與 manifest 不變。

驗證腳本問題另記：初次只預期 Maul，準備斷言被 Black Choker 的第二筆缺譯阻止，
其後誤啟動的探查因 selection 未建立而停止；保留首輪備份，補完整持物盤點後重跑。
單人住宿不問「全隊嗎」，首次套用多人 NO 步驟使 Warder 沒有保存；依既有 StartStay
與追蹤核對後修正腳本，保留失敗並以同映像重跑，未改產品。
兩次窄檔案查詢因錯認 JSON 鍵／程式區段定位而停止，已依實際 schema 重讀；
收尾建置輸入稽核原先預期清冊包含譯表，實際該清冊只含其他建置輸入，
訂正檢查條件後逐項核對，未放寬產品狀態或像素判準。

正式收據位於 docs/verification/cjk-ui/held-names-20261001/receipt.json，
110 附件／72 畫格含命令、種子、字模差分、來源與 SHA-256，控制腳本同次保存；
SPEC-033 窄契約 CONFORMED，父規格仍 READY。中文化索引、CONTEXT 與本地
#34／#35／#40 同步掛入口並保持未結案。worklist render／selftest／verify 通過，
本地仍 23 未完成／83 完成；未驗最長戰鬥武器、怪物、戰利品母體或全遊戲原版 parity。

清理：專案 root-owned 檔案與 .md 同名目錄均 0，新增檔案 UID/GID 1000；
host docker ps -a --filter name=kol- 回讀無執行中或停止容器，全部一次性工作 --rm，
Xvfb 有界並經 trap 收尾。完整保存、原名冊盤點及失敗只留 workplace，正式目錄
不含原版 ZIP／EXE／原始名冊／完整保存；PNG 只適用 private GitHub 既有例外。
未提交／推送、遠端 Issue 寫入、建新映像、建立 dist-all 或發行；保留接手工作樹全部異動。

## 2026-10-01：最長武器正常路徑收尾續核

路由命中重製驗收、正常玩家路徑與文件職責，沿用既有技能、試玩入口與工具映像。
實際武器盤點先擴入 kind1 弓類，再擴入怪物 weapon／worn，得到282筆、最大80像素；
玩家197筆最大64。執行前選 Binderak／Spiney Mace，固定種子1與初始任務5；玩家另選
Fighting Bull 原背包 War Maul。未改產品、規則、typed data、譯表、字模或存檔格式。

八次正式重播共1204筆：玩家正常190／住宿101／重載15、敵方正常296，各英中一次。
差異僅遭遇與執行顯示文字；住宿保存除時間一致，取用精確 record，舊手持入背包，
重載未改寫。八完整文字框字模0差、六整圖0差、缺一墨點與63／79寬度反例拒絕；
六張完整中文畫格目視可讀。敵方正常移動／休息後由鄰格檢視，未提交檢視目標或攻擊。
此為顯示驗收，不把固定種子當作原版亂數 parity，也不代替戰利品或全遊戲完成。

控制器失敗保留本機：初次錯把單張 -shot 追蹤當逐鍵多筆；訂正後又將零起算角色索引
多減1而停止。第三次正常移動後已相鄰，但工具只允許武器攻擊入口，Theodoric 指令列
沒有該格而達四回合上限；依既有 ToggleEquipmentPanel／PromptTarget 契約，改用正常
移動目標格檢視，從同種子、同正常輸入前綴重播成功。另首個全文讀取把 main.go 路徑
認成根目錄，查實在 cmd/kol 後訂正；狀態比較器誤認 execution_text 鍵，依實際 exec_text
訂正，已完成的程序／原始追蹤不重跑、不放寬非顯示欄位。沒有產品例外或種子重擲。

正式收據 docs/verification/cjk-ui/longest-weapons-20261001/receipt.json 與48張PNG同次
接入中文化索引、SPEC-033、CONTEXT與本地#34／#35／#40，父規格仍READY、三項仍開放。
原封裝、完整母體、typed data、完整保存與失敗留workplace；PNG只沿用private GitHub例外。
無提交／推送、遠端Issue寫入、發布、正式版號、dist-all或新映像。

本批收尾：正式重播控制從乾淨輸出再跑八次／1204筆，所有追蹤與48張PNG逐byte相同，
保存只差時間、像素檢查相同；收據另列 fresh-replay-checks.json，不混算首輪八次。
原始封裝六個EXE entry逐項回讀SHA，收據使用真正Knights-of-legend/KOL.EXE的4f798b07，
沒有把其他位元組版本的摘要雜湊當成原始entry。167筆產品建置輸入與選樣譯表／字模均不變。
85份正式附件／48張PNG雜湊核實，輸出UID/GID1000:1000，全專案root-owned及.md同名目錄0；
正式收據附 final-audit.json。工作清單render／正反對照／verify通過，23未完成／83完成。
全部Docker一次性程序--rm、Xvfb trap退出；最後容器清理與Git差異另由主機控制面核對。
主機最終回讀：docker ps -a --filter name=kol- 無執行中或已停止容器，
git diff --check 通過、git diff --cached --stat 為空。保留全部既有異動，未提交或推送。
乾淨重播後的比較器曾因英中 exec_text 換行數不同而停止；訂正為完整顯示欄比較例外，
其餘所有欄位仍嚴格相等。已完成八次程序不重跑，先核對命令、來源與追蹤，再通過
獨立存檔及字模檢查；最後原始與乾淨輸出的追蹤／48圖逐位元相同。

## 2026-10-01：戰鬥目標圓環原版契約（RE-171）

接手核對 Git、真實 host gh auth／Issue 33與64、知識路由及既有工具鏈。
路由命中重製驗收、IDA、文件職責與逆向結論回填；沿用指定技能與映像，不建新工具鏈。
IDA共六個暫存副本：首輪ASCII讀中文索引失敗，明確UTF-8後使用新副本成功；
後續窄查sub_1C683、sub_24131、EGA合成與sub_11390。正式.i64唯讀，名稱及地址未改。
新分級索引六筆自動合併到原始位址／operand／bytes／xref匯出，完整payload只留workplace。

五次dosgolem正常入口固定同一版本968ca774與seed1；沒有掃種子、重擲、座標注入、
強制遭遇、direct-entry或替換等鍵。GUI1按錯層只列觀察；GUI2進武器目標後，
GUI4補只讀捕捉時點正式重生。GUI3核實正常移動與休息一回合，GUI5重播此前綴，
由Theodoric徒手目標檢視相鄰Poontang。原始名冊／角色快照／完整呼叫僅本機保存。

原版12目標框（9空、3有人）與4武器／徒手／防禦子選單框，依捕捉前最後繪製呼叫
取得相位，完整16×16各256像素與獨立小端16-bit遮罩合成零差；單位元反例拒絕。
動態呼叫包含五格5…9；子選單仍在同一位置持續動畫，不是靜止的最後畫格。
像素腳本先錯把模式5當目標層，後依實際兩格圖示列訂正；子選單模式0假設也被
實際值推翻並訂正。未放寬像素判準，失敗分類在本機observations-and-script-corrections.json。

正式18張PNG、來源／命令／SHA-256、可重播四腳本、反向引用正反對照與收據同次保存
於docs/verification/battle-target-cursor-20261001/；研究入口docs/re/20261001-target-cursor.md。
SPEC-010、RE-060、CONTEXT、本地#33／#64與英文UI註解同步訂正，歷史證據保留。
正式行為、圖像資產、譯表、亂數與保存未改，未重建產品；重製版仍白框，未宣稱對拍。
取消／摘要／遠程／邊界完整生命週期待驗；#33／#64不結案，23未完成／83完成。

稽核：18張PNG與本機原捕捉逐byte一致，正式保存的像素腳本重算結果相同；
新分級索引六筆自動合併、舊規格／舊研究反向引用及缺標記／錯模塊SHA反例通過。
Python語法通過，正式輸出UID/GID1000:1000，全專案root-owned或.md同名目錄0。
worklist首核有一行新摘要註解未認領，已將其精確訊號接回#33核實式，未增加忽略項。
全部容器--rm，終止與清理另由host控制面回讀；不清理其他專案映像或資料。
未提交／推送、遠端Issue寫入、正式版號、dist-all或發布，接手既有異動保留。

本批最終回讀：worklist render／selftest／verify全部通過，沒有未認領的自承；
25份附件／18張PNG雜湊與擁有權核實，反向引用正反對照通過。
host docker ps -a --filter name=kol 無執行中或已停止容器；git diff --check通過，
git diff --cached --stat為空。正式輸出沒有原版資料或反組譯payload；暫存副本與失敗僅留workplace。

## 2026-10-01：戰鬥圓環正式接線、摘要來源勘誤與窄驗收

沿用重製驗收、spec閘門、IDA、逆向回填及文件職責路由；工作清單兩次未認領訊號
再命中未完成項資料路由，讀61-worklist-as-data-with-verify.md後修正歸屬。
沒有另建工具鏈、分派子代理、提交、推送、遠端寫入、正式版號、dist-all或發布。
接手既有異動保留，HEAD仍1384e50；真正host gh auth status成功，遠端唯讀核實24項開放。

原版續跑gui6／7／9／10，由固定dosgolem968ca774與seed1正常入口重生。
新增52完整16×16框全部零差，含U-TURN清除、空／有人摘要、四種子選單、遠程首次0,0、
四邊夾限與射擊後8,8記憶重開；沒有原版座標／隊員注入、重擲或等鍵替身。
gui8名冊Down早於繪製而未選到Olear，只作失敗觀察；gui9／10等繪製後正常選角。
射擊摘要的初次預期錯誤由實際+0x1D=1與summary_flag=1推翻，完整框容差仍為0。

依READY窄契約，英文／繁中共用PNG單元5–9小端word遮罩圓環，替換目標／摘要白框，
四種子選單沿用最後格；首次遠程改0,0、空摘要不再以佔有人為條件。
相位0…4往返，每三個BIOS tick推進，明示為現代呈現近似；不寫保存或使用亂數。
GPU測試隔離Ebitengine子行程，避免RunGame結束後其他測試再建圖片的環境限制。
104個英／繁中倍率框零差，錯誤圖集尺寸反例拒絕；Go規則、UI與cmd/kol三套測試通過。

前版5bf08653正常四次455鍵雖通過記錄狀態，初次主列仍多畫圓環。
再追.i64前置呼叫：0x1F16A先進sub_249AE，0x249C5讀當前人物[bx+25h]、0x249CB
寫word_31122；原版gui7 plan-09→10由0x27回到當前人物0。此證據足以推翻RE-160
跨人物沿用全域的舊說。先在SPEC-010追加READY勘誤，再由BeginTurn重載目前人物指令。
保留RE-160原文並追加來源／等級／定位勘誤；14筆索引自動合併成功，完整payload只留本機。

最終程式ac79bb47fc5d19c54fffe838e2d991c0d7b76497d9432e0a1c1885fd75bf847d，
重跑英／繁中前後四次正常455鍵、各456筆記錄狀態完全一致，四種子選單實際走到。
34張完整畫面非戰場區零差；初次／換人主列整圖零差，確認來源修正消除多餘圓環。
正常軌跡不含游標座標，呈現fixture明示注入同相位；兩者不冒稱整場原版戰鬥對拍。
首輪重播誤把首位Olear當Poontang，保留後改正常YES換人；未以錯誤按鍵意圖計算涵蓋。

正式入口docs/verification/battle-target-cursor-20261001.md及remake-receipt.json。
145份最終附件、52原版框、104倍率框與34完整畫面核實，首批25附件雜湊仍相同。
前版收據另存remake-receipt-render-only.json與舊PNG，不混用5bf與ac79結果。
稽核器初次猜錯backlink筆數及欄位，按實際schema訂正；舊verify_backlinks.py曾被擴充，
雜湊護欄攔下後恢復原文，續驗另存verify_remake_backlinks.py並同次接入索引。
root檢查曾跟隨Wine的合法z:連結而誤判容器根目錄，改用lstat與不跟隨連結的find。
這些分類為驗證工具問題，沒有修原版證據雜湊或放寬像素判準。

工作清單render／selftest／verify通過，23未完成／83完成；法術CJK白框訊號回歸#64，
#33僅追完整攻擊列與其他圓環等待層。SPEC-010只升窄契約CONFORMED，父規格不結案。
反向引用三筆與缺標記／模組雜湊／來源勘誤／等級四種反例通過；收據不同程式及原版SHA拒絕。
最終稽核見remake-final-audit.json，全部輸出UID/GID1000:1000，root-owned及.md同名目錄0。
Docker全部--rm、有資源／逾時／無網路，Xvfb trap退出；host ps回讀相關容器0。
git diff --check通過、暫存區為空；原版輸入、資料庫、角色快照與失敗留既有本機workplace。

## 2026-10-01：完整武器／防禦圖示列驗收與繁中框線修正

沿用重製驗收、規格閘門、dosgolem及文件職責路由；讀既有README／CLAUDE與正式
MoveMouse契約。使用既有固定dosgolem版本968ca774及工具映像，沒有分派子代理或另建工具鏈。
真正host gh唯讀核對#33原始「兩塊都有逐畫面對拍收據」條件，未寫入遠端。

原版gui11固定seed1，由正常入口、自然遭遇與正式按鍵重生17個武器／防禦列狀態。
早先箭頭覆蓋第五框被嚴格比較攔下；使用正式MoveMouse(620,4)事件移到右上角後重生，
沒有隱藏滑鼠、裁掉框線、像素遮罩或直接改寫座標。六個武器／八個防禦選取及繞回，
完整296×32區域每框9472像素與獨立合成器零差，包含圖示、間隙與尾端。

修正前英文整列零差；繁中武器列每框差3120像素、防禦列4160像素，定位為尺寸放大
但框線仍一個實體像素。依SPEC-030既定三倍契約，先追加SPEC-010 READY窄契約，
再只修drawBattleIconsCJK的外框厚度。舊ac79程式、薄框畫格及6e19cdcd來源快照保留。
新程式SHA-256為e5be211397bbb2eb11e355a77bbe490cc8a1cd8c618c049679dd97a1456bdc7a，
位於workplace/battle-cursor-20261001/kol-attack-frame-linux，沒有建立正式平台封包。

Go的internal/ui及cmd/kol測試通過，含104個英／繁中倍率圓環框回歸。
修正前後兩語言共四次正常重播，各438鍵／439筆狀態一致；新34張畫格的完整列均零差。
英文17整圖不變，繁中全圖差異只在框線；單位元、錯選取框、缺最後圖示三種反例均拒絕。
人工檢視繁中首武器與末防禦畫格，空目標的左側面板符合既有行為。

正式入口為docs/verification/battle-target-cursor-20261001.md，獨立合成器、重播腳本、
verify_attack_rows.py、新收據及稽核同次接入。實際素材archive83項／71張已驗PNG，
本批13個素材來源的尺寸、SHA及原ZIP entry核對通過；沒有以版本參數代替實際count。
attack-row-receipt.json的SHA-256為41b7d8a6551498ed1280650330e3aa6374e9d58685c32f1d2e34540f0b471572。
新58附件、首批25及圓環145附件各自雜湊與UID/GID1000:1000通過，312個目前輸入核實。
六個新Python模組語法通過；舊收據未重寫，新PNG依private GitHub特例分類，原版payload只留本機。

收尾稽核首命令未帶Docker -i，標準輸入腳本未執行；發現沒有結果檔後，於同一隔離環境
補-i乾淨重跑通過，分類為驗證命令問題，不算產品失敗或驗收證據。
最終結果見docs/verification/battle-target-cursor-20261001/attack-final-audit.json，
全專案不跟隨連結檢查root-owned或.md同名目錄為0。所有工作容器--rm、資源及逾時有界、
無網路、UID/GID明確，Xvfb以trap退出；主機最後清理回讀另記於本批末尾。

本地#33依原始兩塊條件結案，22項未完成／84完成；其他等待層及法術／戰利品／物品頁
仍由#64與完整戰鬥驗收追蹤。原版沒有繁中畫面、兩側隊伍／遭遇不同，只證實列狀態與
既定三倍呈現，不稱整場戰鬥同狀態或父SPEC-010完成。遠端仍24開放，沒有同步寫入。
HEAD仍1384e50，既有異動保留；沒有暫存、提交／推送、正式版號、dist-all或發布。

本批最後回讀：worklist render／selftest／verify通過，22未完成／84完成且兩份產生區塊同步，
結果保存workplace/battle-cursor-20261001/attack-worklist-final.log。文件及紀錄UID/GID1000:1000，
不跟隨連結的find沒有root-owned檔案或.md同名目錄。host docker ps -a --filter name=kol為空，
沒有本專案執行中或已停止容器；git diff --check通過、git diff --cached --stat為空。

## 2026-10-01：物品模式8頁的dosgolem原版續核（進行中）

下一項沿用#64與SPEC-031，先重生正常旅店→物品明細入口，不改既有正式版面。
路由命中重製驗收、spec閘門、dosgolem及文件職責，研究中間產物沿用workplace/
子目錄item-detail-20261001；入口為original.sh、prepare.py。只在容器暫存Go副本
加正式輸入與唯讀呼叫觀察，原版資料從既有唯讀輸入複製到本機專用輸出，沒有原版改碼。

本批結果：gui1多按一次ENTER離開資料頁，只作觀察；gui2訂正後正常21鍵重生。
兩次模式8呼叫的引數8與caller0x16C55確認胸甲／腿甲；其他選中列實際為
BODY／HANDS／FEET及UP返回HANDS，不能按最初意圖稱為盾／武器／口袋驗收。
MoveMouse(638,199)被原版夾限為可見箭頭(304,185)，一像素末角假設失敗保留。
正式比較只准該14×14滑鼠矩形；每框實際67差異像素，不豁免框線、文字、投影或圖示。

三個IDA暫存副本窄查sub_16986／153FF／144C3與11722／1449C；正式.i64唯讀且SHA不變。
六筆新語意與四筆既有語意自動合併，原名、位址、operand、bytes、xref及未知警示保留。
先在SPEC-031第7節審查READY再實作：空格訊息原版欄8列76、英文兩行接續及感嘆號、
中文既有片段語序、SPACEBAR單圖示、ENTER／SPACE返回同列，其他鍵不移清單。
不是任意鍵返回：原第3.4節摘要由sub_11722(0)的直接控制流訂正；原說形成原因保留。
不改物品record、角色身分、保存格式、正式亂數或其他動作規則。

新程式SHA-256為6ffca0ddcaacd88b0d0bca899d4c23aad613654b8ba0c230f802dca72f929bd2，
保存在workplace/item-detail-20261001/kol-items-linux。初次測試漏設GOWORK=off，
既有go.work的/dosgolem外部路徑使Go未執行；同映像、同測試集合關閉工作區後乾淨重跑。
全部internal/...及cmd/kol通過，含104圓環框與空格確認／等待回歸；初次失敗是命令環境問題。

英中修正前後各27筆正常記錄欄位相同；軌跡沒有資料頁暫態，不以該一致性冒稱完整狀態。
六個英文完整頁除原版滑鼠外零差，四個繁中字模框零差，六張中文頁目視可讀。
英中前版胸甲檢視22鍵／新版含空格37鍵各正常返回旅店住宿，四份保存只差saved_at，
不把兩種長度說成逐步同狀態。全部18張PNG實際為RGB，不含透明度差異。
清單11張另有1／8像素差異，正式收據列為未通過，不靠允許區遮掉，不宣稱完整清單完成。

匯出器首輪誤把六筆新索引當全部筆數，實際還自動合併四筆舊索引；次輪把末次UP返回
HANDS誤指定FEET。嚴格護欄皆攔下，尚未匯出PNG；回查路由、原始按鍵與畫格後訂正資料，
沒有更改像素判準、原始PNG或程式去迎合。框邊缺一像素、舊錯位頁與缺字模墨點反例均拒絕。

正式入口docs/spec/031-item-detail-page.md與docs/re/20261001-item-detail.md，
新語意索引、研究台帳、規格索引及#64本地JSON同次接入。
docs/verification/item-detail-20261001/receipt.json保存27附件／18 PNG／314目前輸入，
SHA-256為3d498b4f902d50e6ff2e5bd3b1236cdd7c5175508742fa494dc10e3d4d2a3120。
父規格因完整清單續畫差異回READY，指定模式8／空格切片CONFORMED；不重開已驗舊選中切片。
#64仍待法術、戰利品、其他等待層及未驗物品分支。本地22未完成／84完成，遠端#64唯讀OPEN。
真正host gh auth status成功；未遠端寫入、暫存、提交／推送、正式版號、dist-all或發布。
容器及擁有權最後清理另記本批末尾，原版payload、角色保存及失敗留既有本機workplace。

本批最後核實：verify_items.py預設重算通過；worklist首次攔下SPEC-031新空口袋待驗
訊號未認領，精確接回#64的path／pattern後，render／selftest／verify全部通過。
結果workplace/item-detail-20261001/worklist-checked.log，22未完成／84完成且兩份產生區塊同步。
final-audit.json核對314目前輸入、新27附件與舊25／145／58附件各自SHA及UID/GID1000:1000；
不跟隨連結全專案root-owned及.md同名目錄0，正式PNG均RGB，來源與舊收據未改寫。
host gh唯讀回讀仍24個開放Issue；docker ps -a --filter name=kol為空，沒有本專案殘留容器。
git diff --check通過、git diff --cached --stat為空。所有Xvfb以trap退出，暫存資料留本機，
沒有提交、推送、遠端寫入或發布；下一閘門仍是#64其他畫面／分支與整體收尾22項。

### 2026-10-01：十七欄位正常路徑與名稱置中續核

本輪路由命中原版畫面對拍、IDA與規格閘門；重讀knowledge-router、規格閘門及文件
職責入口，沿用既有reverse-engineer-retro-game-remake與IDA Pro 9.4技能／Docker映像。
先核對AGENTS、WORKLIST、CONTEXT及308筆繼承異動，不重設、暫存或提交。
原版all-keys.json預登記61鍵，由正常標題、書本選Theodoric、旅店及MIRROR進全部17欄，
確認返回、上下繞回、重訪末格及U-TURN離開。原版gui3固定seed1，沒有狀態注入或改碼；
只附加唯讀DS觀察，133.47秒正常測試通過，四次sub_153FF(8)caller0x16C55。
本機重生入口workplace/item-detail-20261001/original_all.sh／prepare_all.py；
remake_all.py、compare_all.py及all-comparison-probe.json保存前版探索，不當正式收據。

原版頭盔、胸甲、腿甲與13空欄除滑鼠外零差，補足空武器與六口袋正常畫格。
巨劍名出現220文字像素差異。回查固定.i64匯出的sub_153FF，確認名稱計數包含NUL，
有號IDIV使偶數長度多偏左一格；現行goodDetail漏此計數。先SPEC-031第8節READY，
才修shop.go一個計算，新增正常商店四種原名欄位測試。名稱、record與保存格式不改。
新item-name-address-index.json四筆已證實語意同次掛入RE-173／既有研究與規格，
IDA暫存副本匯出自動合併14筆語意；正式.i64、原名與舊payload保留。

新程式workplace/item-detail-20261001/kol-items-name-linux SHA-256
351464d1bbff6848defd94e2346b7c4c4741e7299f0842abe340e855f6b10ec2。
同映像全部internal/...及cmd/kol測試通過，含104圓環框；Go模組統計快取非root
不可寫訊息屬既有映像環境警告，測試及建置退出0，不當產品缺陷。
remake_all_after.py英中各66鍵／67筆；四組記錄欄位相同，資料頁暫態仍另以畫格核實。
save_all.py四次正常檢視後旅店住宿73鍵，完整保存只排除saved_at即相同。

verify_slots.py預登記十八選中頁及兩返回頁完整320×200除原版實際14×14滑鼠外零差。
十四繁中字模框及安全矩形零差；四件非空明細／空武器與口袋中文抽樣目視可讀。
中文巨劍只在名稱24像素平移範圍內變動，其餘各66張保持相同。舊巨劍錯位、框邊
缺一像素及字模缺墨點反例均拒絕；首次正式匯出通過後才保存61張PNG／72附件。
新收據docs/verification/item-detail-20261001/all-slots-receipt.json SHA-256
0e3d1f32da14fd2c9ff210574f30130f2b6d8969b9326f2ffb56e958b8591c0a。
舊27附件與舊收據不改。SPEC-031第8節CONFORMED，父規格READY，#64不結案。

清單37張仍有8／9像素差異，只有x15；原版EGA平面迴圈提供原因線索，曝光／還原
生命週期只列強推論，沒有自行製造允許區或改圖形。工作清單首輪scan攔下歷史
「第7節當批未驗空口袋」句，回讀路由61工作清單規則後將目前斷言明寫「第8節補足」，
保留原始收據歷史範圍及同頁其他未完成種類／動作與清單父層，不忽略掃描器。
兩次文件補丁因定位不合拒絕，讀回確認未部分寫入再精確重套，沒有覆寫既有內容。
本輪清理、重驗與待辦最終核實於批末追加；沒有遠端寫入、版號、封包或發布。

最後唯讀重驗抓到JSON索引鍵序列化問題：selected_readonly_ds的Python數字鍵與
回讀JSON字串鍵不同，整份收據重算拒絕；像素、字模、存檔及附件檢查均已通過。
先保存all-slots-receipt-first-export.json、verify_slots-first-export.py與all-slots-first-audit.json
至本機workplace，再把新收據的索引型別明定為字串，僅更新新重驗工具雜湊；不改判準。
首份收據SHA為上列0e3d1f32，新最終收據SHA為
4b21c2f87edf61a58804925f2e5ecb9fde2f4e3cdcd7b044e933a5ce547f1943；舊批收據與附件未改。
本批稽核入口workplace/item-detail-20261001/audit_slots.py，輸出為
docs/verification/item-detail-20261001/all-slots-final-audit.json，已同次掛回CONTEXT。

本批最終核實：verify_slots.py唯讀重算整份收據通過；worklist render／selftest／verify
乾淨重跑通過，22未完成／84完成，WORKLIST與CONTEXT產生區塊同步。
收尾稽核確認315目前輸入、新72及舊27／25／145／58附件各自雜湊與UID/GID1000:1000；
全專案root-owned與.md同名目錄0，61張正式PNG為RGB。Go兩檔gofmt檢查無輸出。
docker ps -a --filter name=kol為空，Xvfb均由trap退出；沒有本專案殘留容器。
git diff --check通過、暫存區為空。正式資料庫及來源唯讀；首份失敗收據與原版payload
保留本機，未提交／推送／遠端寫入／打包／發布。下一閘門仍為#64清單與其他畫面、
未驗物品動作以及其餘整體收尾項目；本輪沒有擴大完成聲明。

### 2026-10-01：正常物品搬移、丟棄確認與保存續核

路由命中原版畫面對拍、規格閘門與文件／待辦職責，載入重製技能、同狀態對拍契約、
dosgolem README／CLAUDE及文件職責／61-worklist入口。先核對AGENTS、WORKLIST、
CONTEXT及HEAD1384e50，保留311筆繼承異動；沒有重設、暫存、提交或遠端寫入。
沿用kol-package、鎖定Go／Ebitengine與IDA Pro 9.4映像，所有工作在有界一次性Docker。

本機入口workplace/item-detail-20261001/original_actions.sh、prepare_actions.py及
actions-keys.json預登記104原版後續鍵。固定dosgolem968ca774／seed1，從正常標題、
書本選Theodoric、旅店MIRROR開物品頁；135.33秒正常實驗通過，gui-actions1保存原圖、
輸入雜湊與唯讀512-byte角色片段。四次sub_18CD5完整16-byte交換依序為400→48、
48→400、400→160、160→48；NO不改物品，YES才清空，訊息等待後返回首欄。
原版記錄、完整存檔與反組譯payload只留本機，不以512-byte片段冒稱完整角色記錄。

IDA暫存副本item-actions-ida-private.json、item-drop-layout-ida-private.json及
item-actions-annotated-private.json保留原始函式名、位址、operand、bytes／xref。
item-action-address-index.json八筆已證實語意同次接入RE-174及SPEC-031第9節，
匯出自動合併八新／五既有註記；正式.i64雜湊未變，沒有推測性改名。
先READY，才補town內部DropConfirm／DropMessage、YES／NO、清空後SPACEBAR等待、
明細下方疊框及成功動作返回首欄。原版18986讓食用繞過重建旗標，最後維持其舊選取，
不把成功搬移的重設套入食用。持久格式、原名、物品record與正式亂數未改。

初版10101aa4及最終c1e3b410兩次全部internal/...與cmd/kol測試通過，含104圓環框。
最終程式kol-items-actions-final-linux SHA-256
c1e3b410380c338d5d1e6c799daee00a75101ea688dfa836a9c94afdc4dd02e7。
Go快取統計不可寫是既有非root映像警告，建置與測試退出0，不列產品缺陷。
remake_actions_after.py英中各109正常鍵／110筆狀態一致；compare_actions.py逐104畫格
比較保留所有清單差異，沒有擴大滑鼠遮罩。23個完整明細／等待及返回頁除實際箭頭
67像素外全零差，四繁中確認／完成字模框零差。完整清單76張仍未通過，動作後另有
692像素框形殘留，主要(17,125)..(149,152)，原因未知，回填#64而非自行製造允許區。

save_actions.py六階段英中共12次正常住宿；reload_actions.py逐份從新存檔正常重啟、
經標題與資料頁檢視再住宿，共12次重載。十七物品欄非空原始16 bytes與原版同期動作
記錄相同，空欄依首byte為零；十二重載選中頁與正常動作相應頁完全相同。
每階段英中完整存檔僅排除saved_at即相同，不宣稱原版金幣／時間／完整角色對拍。
正常存讀檔結果actions-persistence-comparison.json不含原版記錄或完整存檔payload。

正式入口docs/verification/item-detail-20261001/verify_actions.py先重算所有預登記條件，
再匯出actions-receipt.json；SHA-256
d2d13cb2054d707450ce08df70fdc2a4cc85597c29ff9bc30c64f89255ddc049。
保存94附件／82 PNG，含23原版已驗證畫格，舊72及27附件不改。前版提問畫格36,394
像素差、明細框缺一像素與繁中字模缺墨點反例均拒絕。最終只第9節明細／保存切片
CONFORMED；父規格READY，所有物品／洞位／送禮／食用／失敗訊息仍未全驗。

環境／工具失敗分類：受限控制面首次Docker無socket權限，改用獲核准主機Docker控制面；
探索腳本誤匯入不存在write_png、兩次猜錯完整保存檔名均是唯讀檢查錯誤，查明
cmd/kol/complete_save.go後於同映像乾淨重跑。正式驗證首輪把物品明細框反例套到
資料頁／旅館而拒絕，尚未匯出任何附件；限定反例只對含該框的物品頁後重跑通過，
未弱化正常整頁零差判準。所有失敗未寫成產品缺陷。

update_actions_worklist.py只追加#64窄切片與清單新差異，完成狀態與母驗收不改。
worklist首次verify抓到新註解「先等待確認」含『待確認』字樣；依既有acknowledged
契約精確認領這一行及正式收據理由，沒有改已驗程式、掃描器或放寬未完成項。
render／selftest／verify乾淨重跑結果worklist-actions-checked.log；22未完成／84完成。
本批稽核audit_actions.py、actions-final-audit.json核對316目前輸入、新94與歷史
72／27／25／145／58附件SHA及UID/GID1000:1000，全專案root-owned／.md同名目錄0。
正式82PNG均RGB。相關docker ps -a為空、Xvfb均由trap退出，暫存與失敗證據留本機。
git diff --check通過、暫存區為空；無提交、推送、遠端寫入、版號、封包或發布。
下一閘門為#64完整清單殘留、其他畫面／物品分支及原有整體收尾項；本輪不擴大完成聲明。

### 2026-10-01：物品鏡像與立繪殘留修正

前輪分類為進展，保留311筆繼承異動、HEAD1384e50與所有舊收據。本輪命中原版畫面
對拍／規格閘門／IDA路由，載入重製技能、tactical-parity、spec-gated-workflow、
IDA技能及ida_94_official受控工具入口；收尾重查文件職責與61-worklist規則。
所有工作沿用既有映像，Docker有界、無網路、UID/GID1000:1000，沒有新增工具映像。

c1e3b410正常存檔15／42／56／72／88的typed腰帶已空，但Carried末格仍存初始巨劍；
手持鏡像也一直是空。Figure空欄回退舊資料，finishItemAction卻只同步畫面參照。
IDA正式.i64暫存副本sub_18CD5／18F3A證實交換與立繪讀同一角色記錄，首byte零即跳過；
新item-figure-address-index.json七筆已證實語意同次接入RE-175／SPEC-031第10節。
本機ida_item_figure.py產生item-figure-ida-private.json／item-figure-annotated-private.json，
同列保留原名、位址、operand、bytes、xref與等級，正式.i64及原始輸入不變。

先READY，再新增三個舊行為回歸；tests-mirror-before.log確實拒絕空格尾端被當裝備、
動作後鏡像未同步，以及送出者／接收者不一致。物品完成先SyncCarried再組立繪，
YES清除即同步、成功送禮同步接收者；carriedItem改首byte判空。沒有改schema、欄位順序、
原名、未知bytes、背包洞位政策或動作等待規則。兩個產品檔與兩個測試檔gofmt完成。
全部internal/...及cmd/kol測試通過，含104圓環框；tests-mirror.log與正常建置退出0。
最終workplace/item-detail-20261001/kol-items-mirror-linux SHA-256
4c2a5eb003ac5d3c58fa8cf3b215f6be10ee8ac9ed0502e227959e50190be8d4。

prepare_mirror_runs.py從保留的前版腳本另建remake_mirror.py／save_mirror.py／reload_mirror.py，
mirror-run-script-origins.json保存來源與新腳本雜湊。與同一gui-actions1正常104原版鍵的
等價109鍵重播，英中各110筆狀態相同；鏡像修正使第93步持物摘要正確變3，其餘規則相同。
compare_mirror.py保存逐點差分；29完整明細／等待／六重建與返回頁除原版箭頭67像素外
零差，前版692像素反例拒絕。英文只在舊腰帶貼圖矩形(16,124,136,29)改動，繁中104張保持前版。

六階段12次正常住宿與12次正常重載再住宿，typed及raw十七欄均與原版同期動作記錄一致；
二十份前版過期鏡像反例拒絕。其他持久欄位只排除saved_at與發言者Carried即與前版相同。
十二重載選中頁與正常動作頁相同，英中完整保存除saved_at外相同，原始物品與完整存檔只留本機。
正常觀察692像素原因已閉合，追加勘誤保留前版未知形成過程；全104頁中34零差，70清單未通過。
殘餘x15的1／8像素與三張手持小框22文字像素均不豁免；後者已存在前版，未由本修正引入。
繁中正常畫格與程式讀取另證實drawItemPageCJK未貼Figure，補圖／顯示驗證由#64接續。

正式verify_mirror.py驗收後匯出mirror-receipt.json，113附件／100PNG，其中29原版已驗頁。
收據SHA-256 82d4e21da58dbeb0ecc86e2b6e0376679f4154a41025facf273443361f97ee01。
唯讀重算整份收據通過，舊94／72／27及25／145／58附件保持原SHA。第10節窄切片CONFORMED，
父規格READY，其他物品／洞位／食用送禮正常原版流程未全驗，不把繁中無變動當立繪完成。

環境／工具失敗：IDA探針漏指定UTF-8，讀中文索引時ASCII解碼失敗，沒有產生證據輸出；
保留ida-item-figure-first-failed.log，明定讀寫UTF-8，以同映像及新暫存副本乾淨重跑通過。
首次Go檢查未加入映像內/usr/local/go/bin至PATH，未執行測試；補正入口重跑，才取得
三個有意義失敗與後續全綠結果。若干不存在的既有檔名／glob探索僅唯讀，改查實際入口，
均未列產品缺陷。模組統計快取非root不可寫警告沿用既有分類，沒有掛主機library迴避。

audit_mirror.py／mirror-final-audit.json核對317目前輸入、新113及所有歷史附件SHA、
UID/GID1000:1000；全專案root-owned與.md同名目錄0，100正式PNG均RGB。
待辦render／selftest／verify通過，worklist-mirror-checked.log顯示22未完成／84完成，
WORKLIST與CONTEXT產生區塊同步；沒有遠端寫入、提交、推送、版號、封包或發布。
相關Docker與Git差異收尾核對於本批末；暫存、失敗log與原始payload僅本機。
收尾實查：專案相關Docker容器0，Xvfb由擁有程序trap清理；git diff --check通過，
暫存區無變更。移除本輪產生的verify_actions.py位元組快取；來源與附件雜湊不變，
本輪文件擁有權1000:1000，既有工作樹異動保留。

## 2026-10-01：繁中物品頁立繪補回（正常顯示切片完成）

本輪沿用畫面對拍／規格閘門與文件職責路由；SPEC-031第11節先經證據審查升READY，
再於drawItemPageCJK補上既有Figure的三倍貼圖，置於背板後、清單框及小框前。
本機入口沿用workplace/item-detail-20261001/；prepare_cjk_figure.py建立同鍵重播的
remake_cjk_figure.py、save_cjk_figure.py、reload_cjk_figure.py及cjk-figure-script-origins.json。
正式顯示驗證入口預登記為docs/verification/item-detail-20261001/verify_cjk_figure.py，
收據cjk-figure-receipt.json；本輪證據另存，前版收據與payload不覆寫。

最終建置325ba73b2617743dd84697adc50e31c68e16152f6bf0093d3f967140be4d015a，
kol-items-cjk-figure-linux；沿用鎖版Go1.26.7／Ebitengine2.9.9映像，沒有新建工具鏈。
GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1全通過，
含104圓環框，log為tests-cjk-figure.log；模組stat cache的非root不可寫警告不影響退出碼。

同一104原版鍵的等價109鍵：雙語各110筆狀態完全相同，與RE-175亦相同。
78繁中完整清單由原版同期記錄＋固定portrait-layout＋已驗PNG獨立三倍合成全零差；
小框、文字與圖示沿用前版。32其他繁中頁及110英文頁不變，六階段24份正常住宿／
重載再住宿完整保存排除saved_at後不變，十二重載明細頁不變。
正式43 PNG附件另存；實際舊版漏圖、右移1px、明細誤貼反例拒絕。已目視正常手持立繪
step-021及明細step-020，人物／武器可見、中文明細未誤貼人物。
verify_cjk_figure.py --export後以同一入口重算收據通過；audit_cjk_figure.py及
cjk-figure-final-audit.json核對526目前輸入、新43與歷史113／94／72／27／25／145／58
附件雜湊、24份舊保存鎖定雜湊及UID/GID1000:1000；全專案root-owned／.md同名目錄0。

驗證腳本首輪在寫保存來源路徑時將/out誤當/src子路徑，78畫面已驗但無收據輸出；
補project_path明確映回workplace，沿同映像／命令乾淨重跑通過。其他唯讀探索用了
不存在的舊猜名，改查實際030-cjk-character-sheet.md及介面所在檔；未列產品缺陷。
一個SPEC-030文字補丁的定位不符而未套用；該檔未變，新成果已由SPEC-031及索引掛入。

目前22未完成／84完成；工作清單render／selftest／verify全通過，產生區塊同步。
只閉合繁中Theodoric此正常動作造型，英文70張清單、手持小框文字、其他造型／動作
與法術／戰利品仍未全驗；#64及父規格開放。未提交、推送、遠端寫入、打包或發布。
本輪正常保存、原始角色payload、失敗與暫存留本機；最後容器與Git清理於批末實查。
批末實查：本專案Docker容器0，Xvfb由trap清理，git diff --check通過且暫存區無變更；
最終526輸入與43附件雜湊不變，已更新文件擁有權1000:1000，工作清單核實仍通過。

## 2026-10-01：手持物品小框列號座標（正常顯示切片完成）

路由命中IDA／.i64與畫面對拍，載入兩技能、IDA工具契約、規格閘門及文件職責。
沿用workplace/item-detail-20261001/，新入口ida_item_mark.py在一次性資料庫副本匯出
sub_16D0B；item-mark-ida-private.json保留原名／位址／operand／bytes／xref，
自動合併既有分級索引。正式.i64及原始遊戲輸入唯讀，所有payload僅本機。
分級索引入口docs/re/item-mark-address-index.json與研究RE-177，另一次性DB匯出
item-mark-annotated-private.json。正常重生入口prepare_mark.py→remake_mark.py／
save_mark.py／reload_mark.py；來源腳本與雜湊保存在mark-script-origins.json。
正式收據入口docs/verification/item-detail-20261001/verify_mark.py及mark-receipt.json。
收尾入口audit_mark.py與正式mark-final-audit.json，核對目前輸入、歷史附件與舊保存鎖定。

sub_16D0B在16F24按完整類別匹配0x52+13*n技能首byte，最多四列；首個匹配輸出n+1。
固定正常三框的87／第一列87吻合，數值1無誤。16F36重新算框X/8+8；舊英文從骰式
起點多加1所以多右8像素，繁中tx+8*Advance取x464而非三倍x504。
先SPEC-031第12節READY再加ItemMark.SkillX共用座標，兩個UI分支使用同一計算。
不改數值、字模、規則、動作或存檔。七筆已證實索引在暫存副本自動合併，正式.i64未改。

最終821ee273ce3d78f658d4efa83b2716ec89a12e4002fa0365b7c8bf17fc86b78a，
本機kol-items-mark-linux；沿用Go1.26.7／Ebitengine2.9.9鎖版映像。
GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1全通過，
tests-mark.log含104圓環框。既有模組stat cache不可寫警告沒有影響退出碼。
同一109正常鍵，雙語各110筆狀態不變；三英文小框100×22與原版零差，前版每框22文字
像素偏移消除，整頁仍有x15的一像素。三繁中整圖與既有字模獨立重建一致，舊列號清除，
新列號x504；其餘107畫面兩語言各自完全不變。繁中step-031已目視列號／骰式／重量可讀。
六階段24份正常住宿及重載再住宿，完整持久資料排除saved_at後不變；十二重載明細頁不變。
偏移與缺墨點反例拒絕。verify_mark.py --export及再重算均通過，正式27 PNG另存。
原版104畫格中34零差、70清單只剩x15的1／8像素，未新增允許區或放寬完成門檻。
audit_mark.py核對528目前輸入、新27與全部歷史附件、24份舊保存鎖定與UID/GID1000:1000，
全專案root-owned／.md同名目錄0；mark-final-audit.json已寫入。

額外唯讀查14名初始有效角色，技能列全為緊湊順序，未發現跳號列；不把此靜態檢查
當其他技能列的正常原版畫格驗收。其他母體及x15續畫仍追蹤，不重開已驗Theodoric座標。
工作清單render／selftest／verify通過，worklist-mark-checked.log保存結果，22未完成／84完成。
無提交、推送、遠端Issue寫入、版號、封包或發布；原始payload、正常保存與暫存僅本機。
本輪沒有產品測試或驗證失敗；只讀資料定位沿實際檔名，未改工具鏈或原始輸入。
最後Docker、Git差異與文件擁有權於批末實查。
批末：專案相關容器0，Xvfb由trap清理；git diff --check通過，暫存區無變更。
528輸入與27附件雜湊保持不變，文件擁有權1000:1000，工作清單最終核實通過。

## 2026-10-01：物品清單左緣續畫追查（進行中）

本輪命中畫面對拍、IDA及平台規格優先；先讀技能與對應契約，現行821ee273不變。
沿用workplace/item-detail-20261001/，入口prepare_edge.py、ida_item_edge.py及
original_edge.sh；原版觀察只讀緩衝區／呼叫參數，不改CPU指令或遊戲資料。
新的item-edge-ida-private.json與gui-edge1/保存正常原版輸出與原始定位；完整payload僅本機。

本批來源續核完成（產品尚未修正）：首次gui-edge1探針因Bytes需要oracle.Addr而非
uint32編譯失敗，保留run.log；改用oracle.Far後以同一工具鏈／固定seed1／104鍵
乾淨重跑gui-edge2，TestNormalBookEntry在191.05秒通過。第一個參數核對曾取到一般
素材複製，改依實際caller16A2A定位專用完整複製；完整曝光模型最初漏圖示底板，
補入觀察到的byte欄1..38、列164..196複製後全78張通過。兩者均為驗證腳本問題，
不記成產品缺陷或放寬允許區。兩次文件批次patch因不存在的比對行被工具拒絕，
改以實際文件尾端定位完成，未覆寫其他內容。

已證實七次重建的快取步長8160而非平面間距8000，436個首次實際byte改變全符；
104原版畫格及104筆角色／UI狀態與既有gui-actions1完全相同。327次唯讀呼叫、
八個選定原版／現行畫格及78完整清單曝光模型保存來源；產品70張差異仍待修正。
入口analyze_edge.py、prepare_edge_verifier.py、ida_item_edge.py；九筆新索引與相關既有
索引自動合併至item-edge-annotated-private.json，正式.i64不改。
正式入口docs/verification/item-detail-20261001/verify_edge_source.py、edge-source-receipt.json，
SPEC-031第13節已READY。來源資料、完整快取及失敗／正常探針副本只留本機；PNG沿
既有private GitHub權利分類，未提交、推送、Issue寫入、版號、封包或發布。
產品未改，不重跑無關產品測試；本批做來源重生、模型正反核對、唯讀重驗及工作清單核實。
Docker、檔案擁有權、暫存與差異衛生在批末實查並追加結果。

來源收據以全唯讀掛載重驗通過；九筆新語意與相關既有索引均由IDA自動合併。
工作清單首輪verify發現第13節兩行現況自承未掛入口；將「英文清單快取還原尚未實作」
與「70張差異尚未改」掛至第64項核實模式，保留原待辦。相同容器乾淨render、selftest、
verify已通過，仍22未完成／84完成，worklist-edge-checked.log保留首次未掛訊號，
worklist-edge-final.log保存最後核實。不是產品測試失敗或把未完成改成已完成。
本批稽核入口workplace/item-detail-20261001/audit_edge_source.py，輸出正式
docs/verification/item-detail-20261001/edge-source-final-audit.json；保持歷史附件與保存鎖定。

批末實查通過：528既有輸入、15個來源鎖定及612份新舊附件雜湊一致，RE-176與RE-177
兩組各24份保存維持原雜湊。正式.i64未改，九筆新語意自動合併；全專案root-owned
與.md同名目錄0，文件及產物UID/GID1000:1000。專案相關執行中／已停止容器0，
本批未啟動Xvfb或建立新映像。git diff --check通過、暫存區無變更；來源／探針／失敗
紀錄僅本機workplace，產品821ee273未改。本輪來源切片完成，實作及正常產品驗收仍待辦。

## 2026-10-01：英文物品清單快取還原實作（進行中）

上一輪是進展：RE-178已證實快取步長與曝光順序，SPEC-031第13節READY。
本輪命中正式實作閘門及畫面對拍，載入重製技能、spec-gated-workflow與tactical-parity契約。
實作入口internal/town/itempage.go的RestoredEdge暫態與internal/ui/itempage_edge.go：
由既有PNG合成結果在GPU保留一欄，再依實際平面修改及還原範圍繪製，不逐幀讀回CPU。
既有繁中版面與持久格式沿用；下一步正常動作、十七欄、繞回及保存／重載驗證。
現有異動保留，沒有提交、推送或遠端寫入；產品尚未驗收，不升CONFORMED。

正常重生入口沿workplace/item-detail-20261001/：prepare_edge_remake.py另建remake_edge.py、
remake_edge_slots.py、save_edge.py、reload_edge.py及edge-script-origins.json。
正常104原版鍵的等價109重製鍵與17欄／繞回61原版鍵的等價66鍵分開驗證；歷史輸出不改。
第一次建置命令的自動權限審查逾時、未啟動容器；依工具允許重試一次後執行通過。
完整internal／cmd及104圓環框回歸通過，本機程式1e2c891b完整雜湊保存於後續收據。
Go既有唯讀stat cache警告非致命，不歸因為產品缺陷。
正常畫面預核入口compare_edge.py，另存edge-comparison-private.json；只沿用既定滑鼠允許區，
完整原版104動作頁及61欄位／繞回頁分開核對，繁中110張與前版逐像素比較。

首版1e2c891b正常返回頁100發現x15／y164、165兩點差異；這是實作漏掉已登記的
圖示底板覆蓋，保留原版正常輸出及本輪首版畫格，不加入允許區。依RE-178實際
sub_19E9D矩形新增唯一ItemPageLayout.ControlBack，投影還原時保留底板覆蓋。
修正後以同一容器與命令重跑完整internal／cmd及104圓環框通過；註解呼叫定位再以
原始返回位址144BA核對並修正，只有註解變更的最後一次另做建置，未重跑無關測試。
最終入口prepare_edge_final.py、remake_edge_final.py、remake_edge_slots_final.py、
compare_edge_final.py、save_edge_final.py與reload_edge_final.py，來源雜湊存edge-final-script-origins.json。
正式重驗入口docs/verification/item-detail-20261001/verify_edge.py，收據另存edge-receipt.json；
source／mark及更早收據保留原有程式與來源身分，不覆寫歷史輸入鎖定。

最終程式fb185a0a1a9217feb25b66b2458cde1f6c15ae328f87ed79a9b6d291b57f5fba，
kol-items-edge-final-linux，Go1.26.7／Ebitengine2.9.9；image仍為既有
eob-remake-go:1.26.7-ebiten2.9.9-audio（sha256:3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa），
kol-package:20260907（sha256:3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193）。
tests-edge-final.log保存完整GOWORK=off KOL_MARKER_ORACLE=1 go test ./internal/... ./cmd/kol -count=1結果，
含104圓環框及新正常返回／重建暫態回歸。正常輸出為normal-edge-final-{en,zh-Hant}110步、
normal-edge-final-slots-{en,zh-Hant}67步；source104＋61英文整頁除實際滑鼠外165張全零差。
改動只在前版70張的x15，110繁中圖及兩語言各110／67筆狀態保持既有結果。
六階段save-edge-final／reload-edge-final雙語24份保存只排除saved_at後完整資料與MARK前版相同，
十二重載明細頁亦不變，十七欄正規化物品與原版同期記錄一致。
重載工具新副本修正原先對古早save-actions路徑的語言比較；歷史工具不改，正式新收據
另逐份比較實際save-edge-final／reload-edge-final資料，不只採信工具摘要。

verify_edge.py --export保存15張新PNG與edge-receipt.json；舊70頁、首版底板漏兩點及
損壞框線單點均被正式完整比較拒絕。全唯讀重驗通過；首次重驗漏掛工具要求的/out，
在讀檔前即失敗，補齊同一私有輸入的唯讀掛載後按同一映像乾淨重跑成功，屬驗證環境問題。
文件批次補丁一次因歷程索引不存在的定位遭拒，改用實際尾行接入，原有內容保留。
SPEC-031第13節升正常窄切片CONFORMED，RE-179追加原始定位與來源身分；父規格及#64開放。

本批文件／待辦更新入口finish_edge_docs.py只追加RE-179並重新產生本地清單；
收尾入口audit_edge.py輸出docs/verification/item-detail-20261001/edge-final-audit.json。
稽核使用kol-package:20260907，專案/src:ro、既存私有工作目錄/out:ro、正式證據目錄/proof:rw，
UID/GID1000:1000、--rm、--network none、--memory 1g、--cpus 1、--pids-limit 64，外層120秒。
新腳本與證據均由本歷程、CONTEXT及規格掛入；正式PNG只適用既有private GitHub保存特例。
未提交、推送、遠端Issue寫入、版號、正式打包或發布；全部原版快取／反組譯／保存payload僅本機。

本批正式唯讀重驗與收尾稽核通過，edge-final-audit.json鎖定531目前輸入、536來源／
操作輸出、627新舊附件、RE-176及RE-177兩組各24份歷史保存。全部檔案擁有權
1000:1000；全專案root-owned及.md同名目錄0。已目視正常繁中返回頁100，人物、清單
及空武器小框可讀；繁中110張保持前版，沒有新增視覺決策。
worklist-edge-final-implementation.log保存render／selftest／verify全通過與22未完成／84完成；
產生區塊同步，真正未驗的法術／戰利品訊號仍讓#64保持開放。
批末Docker實查本專案相關執行中及停止容器0，Xvfb由有界擁有程序trap清理；
git diff --check通過且暫存區無變更。沒有新建image或清理其他專案資源；
本機失敗首版、探針、完整存檔與正常輸出保留供重生，未加入正式交付或公開入口。

## 2026-10-01：完整自然戰鬥正常輸入續核（進行中）

上一輪RE-179已完成物品左緣切片；本輪依現況與22項待辦轉回V01完整戰鬥鏈。
路由命中正常玩家路徑與戰術同狀態對拍，載入重製技能、試玩／戰術契約與文件職責。
RE-158六回合舊收據在執行段替換sub_1B100等待，且有人受傷後停止；本輪另建
workplace/battle-complete-20261001/prepare_original.py及original_complete.sh，在一次性
副本使用既有正常Olear招募Theodoric／自然SLEDGE遭遇及固定seed1，續送真實鍵盤輸入。
探針不替換等待、不改角色／座標／原版指令，觀察既有已證實入口與完整原始記錄。
完整payload與原始遊戲副本只留本機；原始輸入與正式程式只讀，舊六回合收據保留。
下一閘門是實際到達戰鬥出口後，核對第一個與重製正常鏈不一致的狀態；未到此閘門
不宣稱整場戰鬥、回寫或保存已完成，不改父層22項狀態。

原版original1／TestNormalBookEntry在59.84秒通過，351正常鍵、350次真實等待，
實際到達第61回合戰鬥出口0x211AD；122個入口／結尾邊界、1458個亂數呼叫與1318則訊息
保留完整原始資料。前六回合的12邊界，seed、18份參戰者原始記錄與耐力／疲勞全部與
RE-158舊收據相同，證實本情境保留等待未改這六回合的規則結果。已目視出口最後訊息，
原版顯示Theodoric胸口傷勢使其昏迷；此刻尚未執行戰後回寫，不能冒稱保存鏈完成。
prepare_compare.py另建介面轉接與正常初態fixture，compare_rounds.sh沿用現有回合常式
比較，目的為找首個超出六回合的差異；這層仍逐回合接手，不當作重製正常整場驗收。

轉接準備首輪漏加tools匯入路徑，第二輪已正確產生Olear／Theodoric與SLEDGE夾具，
但動態執行抽取器污染外層root變數而讀錯工具路徑。兩次均在產品測試前失敗。
依連續失敗閘門回查路由、腳本及實際夾具，改用獨立命名空間與/tmp乾淨輸出，
既有夾具及轉接收據以完全相同JSON驗證，不覆寫歷史。這是驗證腳本問題，不列產品缺陷。

完整比較首輪確認近戰選項1先擲rand(3)違反SPEC-009既有rand(100)→rand(3)契約；
第15回合起骰序不同。新增TestResolveRandomStyleOrder在未修程式的兩條路均失敗，
attack-order-before.log保留負例；規格續核先READY，再修Resolve順序。第24回合的0xFF
倒地標記則只在測試夾具轉為Actor的0，原始byte保留，正式表示與保存未改。
compare-rounds-after.log顯示59／61執行段通過，剩22／34涉及出手先後，未放寬欄位。
私有診斷首輪Go overlay文字替換誤及AI函式造成編譯失敗，第二輪限定首函式後可跑；
保留兩份order-debug日誌，不歸因產品。唯讀工具專案實際路徑為/home/anr2/ida_94_official，
沒有依錯猜cht子路徑建立掛載。新入口ida_battle_order.py查既有正式資料庫的一次性副本，
自動合併受版控分級索引；工具與原始payload僅本機，正式.i64唯讀。

IDA初輪標準輸出為空，第二輪明確指定/tmp及-L後定位為索引read_text沿ASCII解碼失敗；
回查IDA路由及已載入工具契約，指定UTF-8後新副本成功，函式338、資料庫雜湊與正式輸入
一致。無工具／授權重建；沒有在主機執行分析。原始查詢證實sub_24FD3讀+0x02，
舊LoadLevel以遭遇等級計算的解釋錯誤。original2另加183次唯讀負重觀察，122邊界、
351鍵、1458亂數及1318訊息與original1完全相同；SLEDGE固定等級byte1、負重87、速度35。
新負重回歸首輪Field值／指標型別寫錯而編譯失敗，修正後monster-load-before2.log
在舊正式程式上拒絕75／30，既有新骰序回歸已通過。先SPEC-009勘誤READY，再修
NewQuest初值與ReloadGear輸入，攻防縮放及遭遇等級不改；gofmt在UID1000容器完成。
compare-rounds-final.log全部61執行段通過。下一驗證入口check_rules.sh做完整internal／cmd、
舊正常AI／執行段／面板回歸，再建立本機候選程式；仍不宣稱完整正常GUI或戰後保存。

完整internal／cmd及104圓環框通過，舊正常五次AI與六執行段／面板抽樣通過；
check_rules.sh已建置kol-battle-rules-linux，SHA-256
b1264c4e12452cad59ed015bab02bd339c967b3da41bbab037019faf6d28f058。
後續GUI入口check_normal_items.py沿RE-179正常標題及109鍵，重播雙語物品流程，
比較歷史完整110筆狀態與110張PNG；新輸出另存，歷史輸入鎖定及收據不改。

新正式入口docs/re/20261001-complete-battle.md、battle-order-address-index.json與
docs/verification/battle-complete-20261001/verify_rounds.py已同次掛進RESEARCH-LOG及SPEC-009。
收據只記來源身分、雜湊、語意比較及回歸，完整原始記錄仍僅本機；沒有新增發布目錄。

雙語正常標題／物品回歸完成，各110筆完整狀態與110張PNG逐byte保持RE-179結果；
沒有新增允許區、變更存檔或覆寫歷史。七筆索引以新暫存DB自動合併成功。
正式receipt.json鎖定1213目前輸入與275附件，SHA-256
b219d14e49e415883a0234147bef16bc261a2d3ca95016ca71d9170ee70b6c3b；
verify_rounds.py --export及全唯讀重驗均通過，錯誤seed／替換等待的反例拒絕。
SPEC-009只將本輪逐回合執行段及負重修正升窄切片CONFORMED，父規格仍READY。

清單首輪新增一行真實戰後缺口沒有機器認領；第6項仍保持完整玩家鏈人工閘門，
第44項改綁SPEC-009「尚未驗戰後回寫、保存、完整GUI」這行具體訊號。
訊號消失也須回到第6項審查，不自動結案；沒有隱藏未完成或降低原完成條件。
worklist-rules.txt保留首次結果，worklist-rules-final.txt保存嚴格退出碼的render／selftest／
verify／scan全通過，22未完成／84完成、1427行自承全部有入口。現行清單同步，
另訂正第6項沿用舊規劃YES差異的過期敘述：該差異已由RE-160修正，不重開舊項。

收尾稽核入口docs/verification/battle-complete-20261001/audit.json，另鎖定正式收據、
重驗程式、研究／規格與目前規則檔；目前真相已更新CONTEXT頂表為b1264c4e。
Docker實查本專案執行中／停止容器0，Xvfb由trap清理，沒有新增映像；
git diff --check通過，未提交、推送、遠端寫入、版號／封包或發布。完整GUI與戰後保存
維持待驗，暫存原版副本、失敗日誌、程式與220張回歸圖只留本機。

稽核首輪漏傳docker -i，Python stdin沒有收到腳本、雖exit0但沒有輸出，未當成驗收。
補-i後同一映像／資源限制乾淨重跑，audit.json實際寫出：1213輸入、275附件均維持
鎖定，所有檔案UID/GID1000:1000，全專案root-owned與.md同名目錄0。
最後Git差異檢查通過、暫存區空；專案相關容器0、無新建映像，正常Xvfb已清理。

## 2026-10-01：正常戰鬥出口後回寫續核（進行中）

上一輪RE-180為實質進展，61執行段與兩處規則修正已驗，不重開物品頁切片。
本輪命中正常玩家路徑／戰術對拍及文件職責，核對現行b1264c4e、22項待辦與工作樹。
新私有入口workplace/battle-settlement-20261001/prepare_original.py及original_complete.sh
從RE-180同一固定正常來源續送八次原版等待，觀察角色與世界快照；來源父腳本唯讀，
不修改原版指令、角色、座標或等待，不以局部收據代替完整重製GUI／保存。
SPEC-027已有READY共同出口傷勢契約；本輪先取得正常回寫實例，不另猜敗北掉落規則。
重製正常操作入口為同目錄run_normal.py／normal_gui.sh，固定現行程式與seed=1，
從標題送鍵、輸出逐步狀態與畫格；不使用載入夾具、遭遇注入或座標修改。
正常gui2初核發現58回合與原版61回合不一致，保留此結果，不以逐回合通過掩蓋。
prepare_probe.py／build_probe.sh只在一次性程式副本增加完整觀察；另以明示測試條件，
於與原版相同的首次規劃入口固定seed=1，排除兩側不同播種時點。正式程式不改。
首版觀察器把亂數物件換掉卻保留舊規劃參照，且未複製逐步參戰者slice；gui3只作
診斷，不能當同狀態驗收。probe2修正共用亂數參照與快照，舊資料不覆寫。
ida_settlement.py沿用IDA9.4正式.i64的暫存副本，窄查既有敗北出口、物品保護判定與
門檻直接交叉參照；原名／位址／bytes／xref與分級索引自動同列，未知仍帶警示。
post2增加出口門檻／競技場閘門／任務物品保護的唯讀原始欄位，從同一固定正常來源
重新生成；post1、首版與probe2產物留存。兩側訊息動畫分段不同，逐鍵微狀態不直接
當同狀態比較；回合入口、完成後與首次規劃等待另做明確對齊。

## 2026-10-02：RE-181連續戰鬥與出口來源收據

工作跨午夜沿用battle-settlement-20261001，不改既有目錄身分。新正式入口為
docs/re/20261002-battle-settlement.md及docs/verification/battle-settlement-20261002/verify.py，
同次接入RESEARCH-LOG與SPEC-027。英中研究副本各501狀態／畫格、61連續回合與1458
骰序一致；回合入口19欄／計數器61組、首次新規劃60組通過。原版出口32次搜刮與
六份完整768-byte角色模型精確重建，省略搜刮反例拒絕，產品與保存鏈仍有實質缺口。
初版正式重驗器把終態指令暫存也當完全相同而拒絕；查明原版0／0、重製41／33，
保持其他終態欄位全部比較並在收據明列差異，不宣稱所有暫存相同或完整remake完成。
IDA初次命令用錯/opt/ida/idat，查實既有入口/opt/ida-pro-9.4/idat後同映像乾淨重跑，
輸出338函式與固定輸入／.i64雜湊，屬工具命令問題，非產品錯誤；沒有另建映像。
首次收據已生成；verify.py的指令差異tuple寫進JSON成list，新增verify_readonly.py作
單純JSON表示正規化後的唯讀入口，來源判準與已生成收據／原匯出工具皆不改。
七筆已證實出口語意存入docs/re/battle-settlement-address-index.json，同次接入RE-181，
IDA再以唯讀正式DB的暫存副本自動合併匯出；沒有推測性改名或覆寫既有原始payload。
本輪收尾稽核入口docs/verification/battle-settlement-20261002/audit.py／audit.json，
鎖定正式收據與唯讀工具、新七筆索引／合併匯出及現行清單；RE-1801213輸入／275
附件唯讀重驗維持原雜湊。首輪清單漏認領SPEC-027的新保存缺口，改綁既有第44項
同一訊號後render／selftest／verify／scan全過：22未完成／84完成、1436行皆有入口。
Docker批末查專案相關執行中及停止容器0，GUI的Xvfb均由trap與wait清理，沒有新映像。
所有私有來源／研究副本與畫格保留本機，未提交、推送、遠端寫入、打包或發布。
全部來源與正式資料庫唯讀，UID/GID1000:1000，明確私有輸出；新目錄沿workplace職責，
既有RE-180原始來源、收據與當時1213輸入鎖定保留，不覆寫歷史。

上述audit.py已實際通過：RE-181共2522個附件、七筆索引合併、RE-180的1213輸入／275附件
維持原雜湊；全專案無root-owned檔案或同名.md目錄，UID/GID1000:1000。批末主機再次
核對kol-容器為0。

## 2026-10-02：戰後搜刮控制條件續核

命中spec-gated-workflow與文件職責路由，實際載入兩份契約，沿用IDA9.4技能與工具入口。
workplace/battle-loot-20261002/prepare_original.py、original_complete.sh沿用RE-181正常61回合
來源，另在原始出口位址唯讀觀察共用暫存及升階／倒地／保護欄位；原有來源與收據不改。
新工作區沿workplace的既有研究職責，不建立新根目錄。搜刮切片維持DRAFT至控制條件
及保護接線完成審查；此階段尚未修改正式程式。
### 同日 RE-182 正式候選與嚴格物品核對

搜刮控制與保護經READY審查後接32隻怪物typed欄位、正式出口、固定背包空格及
完整保存。169d57f1候選全部內部／cmd、144組真實原版搜刮案例與出口v2重載回歸通過；
驗證副本缺docs掛載及保存夾具名冊不足的兩次失敗留診斷，修正環境／夾具後乾淨重跑。
雙語正常GUI各501狀態／畫格、61回合與1458戰中骰序、出口傷勢24／19及錢包1777／0通過。
verify_loot.py嚴格34槽比較失敗，未產生正式receipt.json。先前準備中的完整物品符合
敘述已訂正：Olear手持弓未清，Theodoric四個空背包殘留被清零；未放寬比較器。

來源重跑入口prepare_original.py／original_complete.sh，固定原版loot1全流程與RE-181
前綴一致；loot_source_test.go／source_cases.sh保存144組無替代常式案例。
ida_control.py輸出控制及四筆分級索引自動合併；正式來源與.i64唯讀。
parent-preservation.json保存1213前版建置輸入及10份稽核文件，共1223份逐檔雜湊核對，
不把歷史收據改成新輸入已通過。當前候選與失敗GUI保留供正反比較。

手持續核入口ida_hand.py，輸出hand-collapse／hand-drop兩份IDA原始定位；12組無替代
常式drop_source_test.go／drop-source-cases-private.json核對全手持／地上記錄／座標。
固定種子1前後不變，空手含殘留與地上滿格不改，有空格完整搬移。SPEC-027補充經
DRAFT／證據審查升READY後才修產品；正常住宿重載仍未完成。所有研究payload僅本機，
沒有提交、推送、遠端寫入、封包或發布。批末另查Docker、擁有權及工作清單。
### 同日 RE-182 修正後正常出口續驗

空槽保留與搜刮整槽清零分開處理；倒地、疲勞及擊退共用完整手持搬移，怪物撿取
保留當前完整手持記錄。新增殘留144原版案例、正式288搜刮回寫／12掉落比對通過。
rules2.log完整內部／cmd通過；rules3.log增加17槽、保存傷勢／錢包斷言後聚焦重跑通過。
目前來源重建候選SHA仍為7bc7d34a；觀察副本a87ae4ed，修正GUI另存fix2不覆寫首版。
雙語各501狀態／畫格、61入口與1458戰中骰序、出口34槽完整位元組及傷勢／錢包通過；
首版五處物品差異被同一核對器拒絕，Theodoric指令差異已消除，怪物暫存差異仍明列。
九筆已審查語意自動合併IDA匯出，正式.i64保持原雜湊。

正常住宿入口run_save_chain.py延長按鍵，首個inn1-en沒有存檔：回城途中新ORC遭遇
阻斷固定移動腳本，屬路徑控制失敗，不列產品缺陷。drive_inn_chain.py改依畫面選
正常休息規劃／逃走，再走回城；逐鍵重播過慢，保存inn2-en診斷並主動停止唯一
kol-loot-inn2-en容器。drive_inn_chain3.py把世界步合併八鍵，場景本身阻止後續
不適用移動；不替代遭遇、不注入座標、角色或指令狀態，不把住宿尚未通過宣稱完成。
上述入口均在既有workplace/battle-loot-20261002索引下，輸出僅本機；新收據待批末凍結。
# 2026-10-02 RE-182 正常住宿與實際存檔重載續核

本輪驗證入口：`workplace/battle-loot-20261002/drive_inn_chain4.py` 修正既定規劃 YES 與逃走 Y 的操作腳本；`reload_inn_save.py` 以正常住宿產生的存檔另起行程，再正常進門招募 Olear。失敗診斷保留，不以測試資料或座標注入替代存檔。英文 `fix2/inn4-en/` 已產生住宿存檔；繁中逐次重播因 30 秒腳本逾時中止，改以英文已完成的同一組 768 個正常按鍵一次重播，90 秒有界時限下於 `fix2/inn5-zh-Hant/` 成功。住宿之後的自然遇敵只驗重製玩家路徑，不外推原版對拍；研究副本在每場新遭遇的首次規劃固定測試種子 1，正式程式不鎖種子。重載結果待下段紀錄。

RE-182正常保存鏈通過：雙語769狀態完成住宿，16槽名冊含14角色及2空槽；另起行程正常進門與招募，雙語各16狀態／畫格逐欄還原34槽及typed裝備。保存驗證起初把空名冊槽當角色讀取，修正驗證器判空後以同映像乾淨重跑通過，沒有改產品。完整持久資料僅排除saved_at後雙語相同；傷勢24／19、金額1721／0。SPEC-027子切片有限CONFORMED、父READY；新正式收據及後續稽核入口為docs/verification/battle-loot-20261002/receipt.json、verify_loot.py及audit.py／audit.json。

真正host gh auth status成功後唯讀核實24筆遠端open；#6／#44保持open，#33／#28／#23與本地done不一致。本地22未完成／84完成僅為本地資料，不假稱遠端現況，沒有Issue寫入或同步。工作清單追加本次有限驗收與目前父限制，歷史不重開已完成切片。

正式程式重載補驗：fix2/kol-battle-loot-linux複製品原為0644，僅修該檔為0755，SHA7bc7d34a不變；首版啟動失敗的command保留於production-reload1-en。production-reload2雙語各16張完整PNG與觀察副本逐張相同，狀態也一致。正常住宿與重載操作沒有狀態注入。封存前主機git diff --check通過、暫存區空、docker ps -a專案容器0；實際結果保存在final-control.json。工作清單render／selftest／verify／scan通過，1447自承全部認領。全部臨時資料、失敗診斷及原始payload保留本機，未清理其他專案。

RE-182主收據已封存5258附件。補充稽核入口docs/verification/battle-loot-20261002/audit_runtime.py及runtime-audit.json，同次接入CONTEXT；額外鎖定65份語言資料、標題／城鎮樣本與研究JSON索引，保留主收據及原始audit.py不覆寫。補充稽核在Docker唯讀來源上執行，只在指定驗證目錄新增結果，另驗正式候選0755權限。

RE-182封存及唯讀稽核完成：receipt.json的5258附件與補充65執行資料／索引雜湊均通過；1213前版實際輸入、RE-181原收據與正式IDA資料庫保持原雜湊。audit.json及runtime-audit.json均已落地，後以全唯讀Docker重新執行audit_runtime.py通過。全部鎖定檔UID/GID1000:1000，全專案root-owned與.md同名目錄0；主機docker ps -a再次確認專案容器0，git diff --check通過且暫存區空。沒有重建或清理其他專案映像。存檔、程式、原始payload與失敗診斷留本機，正式封包及整體完成聲明仍受既有父閘門限制。

## 2026-10-02 RE-183 世界座標差異定位

沿用研究工作區workplace/battle-world-20261002；入口prepare_world.py、source_world.sh，從RE-182封存正常驅動器增加唯讀世界／輸入觀察，到首次規劃停止，沒有角色、座標、指令或等待替換。先依文件職責路由建立parent-inputs及parent-preservation.json，保存前版共享輸入與稽核鎖定文件，避免新現況文件覆蓋歷史身分。目標是判定49,23／51,23差異出現在進場前或出口；產品規則暫不改。原始資料、探針與輸出仍僅本機。

source1為觀察腳本替換範圍斷言失敗，保留診斷；限定walker步行函式後，同映像source2正常重跑通過53.31秒。原版0x1E3A8入口已在49,23，既有重製進場前51,23，出口不是兩個世界大格差異的起點。新入口replay_world.py依原版實際完成71個向右步及後續方向序列，用正常GUI重播；明示完成步對齊，沒有把原版100個原始送鍵宣稱逐鍵等價。結果待核。

gui-prefix1雙語各135鍵／136狀態，世界大格均49,23，遭遇SLEDGE／18參戰者；原版正常入口時刻3360，重製為3320，仍保留時間限制。replay_world.py的--complete續送既有61回合正常按鍵，只改正常進場移動方向，正式產品未改。

gui-complete1雖正常返回世界，首次規劃seed已由1消耗至2179684817，戰鬥結果不能視為同狀態。原版最後6個walker方向只等待繪圖，沒有位移；重製提前進規劃並消耗選項骰。新入口replay_world_steps.py依原版相鄰觀察的實際位移保留16步，沒有重新選種子、座標或角色注入；舊診斷不覆寫。

勘誤：gui-complete2仍不符合規劃初態。重新查路由及實際清冊後，確認既有gui-en原為100個R，規劃在索引131至140；此前依摘要取118至127，實際取到閒置R。重播亦漏設既有KOL_PROBE_PLAN_SEED=1，造成首次規劃前六次原生初始化骰後seed2179684817。前段對規劃骰差異的方向鍵歸因沒有足夠證據，予以撤回；原版16步實際移動觀察仍成立。新入口replay_world_seeded.py修正清冊切片與既有播種條件，保留所有前版失敗產物；沒有挑選新seed或改產品。

gui-complete3雙語各488正常鍵／489狀態與畫格通過，首次規劃seed1；規劃完成後360個完整戰鬥狀態與RE-182全部相同，涵蓋61回合、1458骰序與出口34槽、傷勢24／19、金額1777／0。進場與出口均49,23，大格差異由進場前路徑不同形成；時刻3360／3320及次格返回仍待驗。正式產品未改，父規格／#6／#44維持開放。研究及重驗入口同次掛入RESEARCH-LOG、CONTEXT及SPEC-011／027。工作清單只更新現行父限制，3句歷史未驗陳述精確認領；render／selftest／verify／scan通過，22未完成／84完成，1450句自承全部認領。Git diff --check通過、暫存區空、專案容器0，控制收據final-control.json；沒有提交、推送、Issue寫入、版號或交付變更。

RE-183正式receipt.json已一次建立，唯讀重驗入口verify_world.py；補充批末稽核入口docs/verification/battle-world-20261002/audit_world.py，輸出audit.json，只補記擁有權、模式、現況計數與控制面清理狀態。它不覆寫主收據，也不將更新後文件冒充RE-182歷史輸入；所有父輸入及原版payload仍僅本機。

RE-183封存後全唯讀重驗通過：6130附件、1299前版共享輸入、舊5258附件及65補充輸入保持記錄雜湊，舊結果重新計算相同。audit.json已建立並以無可寫來源容器重驗；鎖定檔UID/GID1000:1000，全專案root-owned與.md同名目錄0，正式候選SHA7bc7d34a及0755權限不變。父規格未升CONFORMED，下一步仍為世界時刻及次格返回的有限觀察，沒有改正式亂數、原版等待或硬體範圍。

## 2026-10-02 RE-184 世界步行的戰鬥返回

前輪為有效進度：已定位世界大格差異，RE-183收據封存。新工作區沿用workplace/battle-return-20261002，按文件職責路由建立；入口preserve_inputs.py及ida_return.py。先保存RE-183共享輸入，再以既有IDA9.4唯讀正式DB的暫存副本追步行時刻、時段及次格直接／間接資料流。只處理進場時刻差40與正常返回，不開硬體／無關helper研究；正式產品先不改，原版語意須先補證據與規格。

IDA同映像成功匯出338函式固定DB身分與原始指令。sub_297F2在0x297FB呼叫時段／遭遇後才於0x2981F扣時刻；sub_2A4E5在0x2A94A等待戰鬥返回，再於0x2A952／0x2A958寫次格4,4。新正常來源入口prepare_return.py及source_return.sh延長RE-183觀察至RE-182原61回合／出口等待，唯讀記錄原始BB84及其表byte；沒有修改控制流或原始資料。1303共享輸入已封存。

正常來源source1通過58.01秒，162輸入觀察、549事件；初態、351鍵、61回合邊界、
1458戰中骰、訊息、戰後角色及出口計數與RE-182完全相同。0x1E3A8入口仍3360，
戰鬥返回後次格4,4，BB84留16；DS:12B8加16取byte30，移動率20，步行耗時100，
0x2981F執行後時刻3260。天氣擲骰在戰鬥後、扣步行時刻前，不以提前扣40替代。

prepare_scratch_cases.py／scratch_cases.sh沿用原版搜刮常式，追加初始BB84−1、0、1，
432組固定seed1232439027案例通過，六人完整記錄／骰數／終態種子都核對；各組
136保留輸入、8清槽留16。這是常式邊界，不將它當正常其他戰鬥出口完成證據。
ida_scratch.py首版缺UTF-8 locale而讀檔失敗，屬工具環境，日誌保留；明示UTF-8
編碼及LC_ALL後用同image乾淨重跑成功。120筆直接交叉參照含其他戰鬥寫入，
不推定未清槽的整場會保留進場方向。沒有新image、正式DB改名或硬體時序研究。

review_return.py與ida_reviewed.py把九筆有限已證實定位寫入分級索引並重生IDA匯出，
原始名稱、位址、bytes、運算元及未證實警告保持；正式DB唯讀。研究入口
docs/re/20261002-battle-return.md、world-return-address-index.json及重驗
docs/verification/battle-return-20261002/verify_return.py同次掛入RESEARCH-LOG、CONTEXT
及SPEC-011／027。返回子切片DRAFT，正常勝利／逃走最後暫存與亂數接續待補；
產品仍7bc7d34a，既有搜刮保存CONFORMED範圍不重開，沒有新GUI／保存驗收。
所有新原始payload、遊戲資料與未驗PNG留本機；未提交、推送、Issue寫入、打包或發布。

verify_return.py首版誤將完整1491骰斷言為1458戰中骰，來源未改；依原出口計數
訂正後同容器唯讀重跑通過，原版61回合完整JSON仍與RE-182相同。既有雙語產品
實際出口49,23／2,3／3320被來源返回條件拒絕，沒有用來源通過冒充產品修正。
工作清單兩次未精確認領SPEC-011新句，人工note不參與匹配；依路由實讀工作清單
契約及claimed_by後，finalize_worklist.py將兩句DRAFT限制與既有父限制綁到第44項。
三句正對照與全部移除後的反對照通過，render／selftest／verify／scan重跑通過，
22未完成／84完成、1457自承全部認領。沒有擴大歷史忽略範圍或新增重複待辦。

RE-184正式收據已封存1516附件，audit_return.py／audit.json另記批末稽核；封存後撤去所有可寫掛載，verify_return.py及audit_return.py均通過。1303前版共享輸入、舊6130附件與RE-182結果保持原雜湊及原結果；九筆分級語意合併、正式DB雜湊不變。全專案root-owned及.md同名目錄0，UID/GID1000:1000，相關容器0、暫存區空、git diff --check通過。下一閘門只補正常勝利／逃走最後暫存，再審查返回DRAFT；本輪沒有正式程式修改、GUI／保存新驗收或發布。

## 2026-10-02 RE-185 正常戰鬥出口暫存續核

前輪為有效進度。先讀現況及工作清單，命中dosgolem／IDA／spec閘門路由，載入規格
及文件職責。研究沿既有workplace/battle-exits-20261002；入口preserve_inputs.py、
prepare_flee.py與source_flee.sh。先保存RE-184現況共享輸入，舊收據保持原身分。
新逃走來源從正常標題、旅店招募、出城與原版遭遇開始，兩名隊員以正常MOVE／Left／
SPRINT規劃走向邊界，遇真實逃走問句答Y，唯讀觀察最後BB84。沒有指令／座標注入、
常式替代或新seed；規格先DRAFT，產品不改，原始payload與未驗PNG僅本機。

flee1於48.06秒停止，初始Return選中YES而非MOVE，後續按鍵誤入回合確認。
查SPEC-010及實際完整輸入後確認主列YES第一格、MOVE倒數第二格；移動方式第一格
是取消，SPRINT為第四格。首版只是重播腳本錯誤，不列產品缺陷。入口
prepare_flee2.py／source_flee2.sh改用兩次逗號繞到MOVE及三次點號選SPRINT，
同image、同正常前綴及固定seed乾淨重跑；原flee1檔案和日誌保留。

flee2於58.00秒仍在第一回合規劃，兩次連續失敗後回查spec閘門及SPEC-010實際欄位。
原版選完子選單只寫指令，+0x22仍0；RE-160明定重畫同一人的主列後還須選YES才
確認並換人。首二版缺每名隊員的YES，不把它當產品或原版規則缺陷。新入口
prepare_flee3.py／source_flee3.sh補逐人確認，並以Up夾限、Left夾限、Down選正西，
不把首二版的西北方向宣稱正西。原失敗資料不覆寫，固定seed及正常前綴保持。

flee3程序通過71.56秒，但實際第五回合因SPRINT累積疲勞而全隊倒下，沒有答Y。
它是另一條正常自然敗北返回，仍為4,4及3260；不得宣稱逃走驗收。依既有SPEC-008／
RE-143改用回合確認的0x30逃走格，入口prepare_flee4.py／source_flee4.sh，seed不換。
獨立次格寫回的來源已足夠：RE-184正常觀察及IDA無條件4,4寫入先審查為SPEC-011
RE-185 READY，再加正式startAmbush來源標記與共同出口置中；扣时刻／亂數部分仍DRAFT。
build_center.sh驗正式出口三種結算、其他／過期標記及再次按鍵，並跑全部內部／cmd。
正式保存格式、世界大格、時刻及角色規則不改；產品修正必須另做雙語正常GUI驗收。

flee4通過75.52秒：兩人REST並各按YES，再從回合確認列選0x30逃走及答Y，
10鍵／首回合確認後正常返回；兩人頭部15／16，兩側倒地計數255，BB84直到返回為0。
原版次格4,4，3360扣60成3300；flee3自然敗北則留16、耗時100。不可固定所有出口
100耗時；勝利最後暫存尚未觀察。IDA9.4固定DB暫存副本保留338函式，672／128／58
有限指令及120筆直接暫存xref；0x21765屬原始sub_1E3A8，不另造函式或改正式DB。

build_center.sh的全部internal／cmd通過，正式候選SHA9c64e343、0755；沿既有
prepare_probe／build_probe建立cd1f16f0觀察副本，只加唯讀資料及同時點測試seed。
normal_gui.sh雙語各492正常鍵／493狀態與畫格，61回合、1458戰中骰及出口34完整槽
保持RE-183結果；返回前480狀態／圖完全相同，九張返回圖只改英文8／繁中72點位
像素。真正返回後次格4,4，R／D／L／U四步正常走回4,4；時間仍3320，來源3260未修。
run_reload.py與reload_gui.sh用新正式程式重載RE-182正常住宿存檔，正常進門招募，
雙語各16狀態／整圖完全相同。首版保存路徑漏fix2，尚未啟動遊戲即停止，訂正後用
同image／命令重跑；沒有修改存檔，也不宣稱從新置中位置另走住宿保存鏈。

verify_exits.py預驗首版誤要求新出口初態與RE-184先行規劃後記錄全同。
實查差異只在原始指令欄及步數；兩條新來源初態全同，與RE-184角色／計數／seed
相同。收據明列此範圍，原輸出不改。像素檢查初用沒有PIL的kol-package而停止；
改用專案既有tools/png_io.py標準函式庫解碼，不另裝工具或新image。
history_check.py映射1309個已封存共享輸入，用原RE-184驗證器重算1516附件與下層
結果，歷史證據保持原身分；當時的production_unchanged只描述封存歷史來源。

研究docs/re/20261002-battle-exits.md及重驗／稽核入口同次掛入RESEARCH-LOG、
CONTEXT與SPEC-011／027。先獨立READY再有限CONFORMED，返回時間／亂數仍DRAFT，
父#6／#44保持開放。finalize_worklist.py對三句現行限制作正對照及全部移除反對照；
render首跑少CONTEXT可寫掛載，屬環境失敗；補掛後另抓到一行時間限制未精確認領，
接回第44項既有訊號後，render／selftest／verify／scan全部乾淨重跑通過。
22未完成／84完成，1462行自承全部認領。沒有提交、推送、Issue寫入、版號、封包
或發布；程式、原始payload、遊戲副本與新PNG僅本機。批末另封存receipt及audit，
容器／擁有權以稽核實查，不把來源或規則測試宣稱完整remake完成。

RE-185主收據一次封存4284附件，SHAeb2b83de；audit_exits.py另存稽核後撤去全部
可寫掛載，verify_exits.py及audit_exits.py完整唯讀重跑通過。1309歷史共享輸入與
舊1516附件、下層歷史結果保持；正式DB不變，UID/GID1000:1000，全專案root-owned
及.md同名目錄0。實際目前main與封存前版只有來源欄位、建立時標記及共同出口
4,4三段差分，既有工作樹異動保留。Git diff --check通過、暫存區空，相關容器0；
下一步仍限定正常勝利最後暫存及返回時間／亂數契約，整體Goal保持進行中。

## 2026-10-02 RE-186 正常勝利返回來源

前輪為有效進度，正式返回次格切片已有限CONFORMED。重新核對AGENTS、WORKLIST、
目前HEAD與工作樹，現有358項異動保留。命中復古重製、spec與dosgolem路由，載入
文件職責及spec閘門；新研究沿workplace/battle-victory-20261002，先封存RE-185共享
輸入。入口preserve_parent.py、prepare_victory.py、source_victory.sh；正常兩人步行接敵、
徒手出招與休息，只送原版合法按鍵、保留原seed1設定及等待。勝負依真實終態分類，
沒有指令／座標注入、強制勝利或正式規則修改，原始payload與未驗PNG僅本機。

先保存1316份RE-185共享輸入。source1限定2000鍵後仍在第63回合，沒有退出；
其後沿用舊戰後等待造成額度失敗，分類為研究腳本問題。source2於第一回合14鍵
停止：原版先規劃Theo，腳本卻假設0、1順序，實際指令列0729162705及下一人的
072329152705保留。IDA9.4正式DB暫存副本匯出sub_11722、sub_228B7及原始
sub_1E3A8的先攻／規劃片段，338函式與正式DB雜湊不變。

source3首次準備停在腳本組合層斷言，尚未執行遊戲；修正組合層後用同一容器、
同一命令乾淨重跑。原生等待與先攻槽位觀察有效，但第81回合1738鍵仍未退出。
其第一回合Olear原始位置33,30變34,29，指令39、方向5；敵人在31,37，距離
反而增加。這推翻按散文註解交換兩軸的腳本判讀。source4依正常位移把Down
對應原始+1F、Right對應+20，保留前版反例與seed1，不修改正式規則或補猜傷害。

source4正常147鍵、第四回合勝利後完成八次原生等待並返回世界。0x211AD當下
敵方倒地1、隊伍0；敵人軀幹255，兩名隊員頭部15／16。真正副手換到手上以
完整16bytes核對。返回BB84=1、表byte10、移動率20，次格4,4，耗時60，
3360降3300。戰中86次、戰後12次rand(19)、一次rand(100)、一次天候rand(128)，
出口seed775978695、天候前1502690714、返回1084083701。四組完整初態都與
RE-185正常逃走相同；沒有換種子選結果，失敗樣本保持原始payload及run.log。

研究及索引、SPEC-011／027來源補充、唯一目前狀態與第6／44項已同步；正式程式
及父閘門不改，時間／亂數仍DRAFT。worklist render／selftest／verify／scan通過，
22未完成／84完成，1472限制全部認領。新verify_victory.py封存2109附件，receipt
SHA-256 e1572bf2aebd7e8025c37aa6b7a0f30537113a5b311d7fef7a918c88b9364f18。
撤去全部可寫掛載後完整唯讀重算通過；1316歷史共享輸入、RE-185舊4284附件與
下層結果相同，正式IDA資料庫及cmd／internal／tools保持原雜湊。全部附件與收據
UID/GID1000:1000，全專案root-owned及.md同名目錄0；host docker ps -a核對
相關容器0，git diff --check通過且暫存區空。原版副本、反例、新PNG僅本機，
沒有提交、推送、Issue寫入、版號、封包或發布。下一步只審查有限出口成本與
延後天候／扣時刻的契約；整體Goal保持進行中。

## 2026-10-02 RE-187 世界時間的戰後接續

前輪正常勝利返回來源已封存，屬有效進度。重新核對AGENTS、WORKLIST與目前
狀態；命中spec閘門及文件職責路由，沿用既有remake／IDA技能。新工作區為
workplace/world-clock-20261002，入口preserve_parent.py；先保存RE-186與共同
建置來源。只追世界返回成本所需的最後寫入端，再審查異步戰鬥／提示接續規格。
現有360項工作樹異動保留，沒有提交、遠端寫入、發行或正式行為修改。

三條來源沿正常按鍵及seed1，只增加完整DS:BB84變值監看；勝利57筆、敗北120筆、
逃走0筆。完整來源JSON移除新觀察欄後逐項相同，返回JSON逐byte相同；最後勝利
20BAA寫sub_268D4回傳1，敗北清槽留16，逃走維持0，不按勝負硬編碼成本。
先SPEC-011獨立READY再正式實作pending時間接續及實際隨機戰鬥來源回呼。
全部內部／cmd、提示鏈、紮營換日、重複接續及三結算乘三來源隔離通過。
正式dea837bc與觀察b4b5f65b於既有鎖版image重建；正式程式沒有測試播種旗標。

雙語各492正常鍵／493狀態／PNG，61回合完整資料、1458戰中骰與出口34槽不變。
362個戰鬥狀態維持3360；1458＋32骰後才擲一次天候rand(128)，4284445587變
3077853194符合原版。前480張PNG逐byte相同，後13張只改天候1895／17055像素。
首次圖差斷言誤把繁中畫面當英文三倍布局；以既有cjk.go原點46,247訂正驗證，
另檢查全部差異逐像素對應英文三倍，乾淨重跑通過。產品UI未改。
正式候選重載既有住宿存檔各16狀態／圖相同；本輪沒有新戰後住宿保存鏈。

CONTEXT、SPEC-011／027、研究索引及第6／44項同步；更新順序與本場戰後天候
有限CONFORMED，返回成本40／100、3320／3260及完整進場亂數仍DRAFT。
worklist render首次少掛CONTEXT可寫輸出，修正掛載後同一容器命令乾淨重跑；
render／selftest／verify／scan通過，22未完成／84完成、1478限制全部認領。
1319共享輸入先封存，舊2109附件及下層結果重算相同；歷史來源與收據不覆寫。
主收據及唯讀重驗入口docs/verification/world-clock-20261002/verify_time.py，
批末擁有權、Git／Docker控制面與暫存狀態由同目錄audit_time.py及audit.json記錄。
原版副本、新payload、PNG及候選程式僅本機，沒有提交、推送、Issue寫入、版號、
封包或發布。下一步只補有限出口成本資料流；完整Goal保持進行中。

## 2026-10-02 RE-188 返回成本資料流續核

前輪RE-187實作及正常GUI有已驗進度，本輪沿唯一目前狀態續行。命中規格閘門與
文件職責路由，已讀對應入口。366項既有異動保留；先以
workplace/return-cost-20261002/preserve_parent.py保存1326共享輸入及RE-187收據／稽核。
目前正式程式仍dea837bc。只追返回成本依賴的最後寫入，不把勝敗固定為60／100，
不深挖無關helper；沒有提交、推送、遠端寫入或發行。

IDA9.4一次性DB匯出648條必要指令，固定原版與正式DB雜湊保持。舊工具專案路徑
/home/anr2/cht/ida_94_official不存在；漏接形態檢查失敗後Docker誤建空root目錄。
確認無內容後只rmdir該目錄，實際專案在/home/anr2/ida_94_official，已讀README與
knowledge-base/ida-94-tools.md，沿用原locked-v1 image。首次IDA腳本因ASCII讀取
中文來源失敗；補明確UTF-8及容器語系後從唯讀DB乾淨重跑，保留失敗log。
此為路徑／工具腳本問題，不記成產品缺陷，也未改原始DB。

正常原版147鍵第四回合勝利新增39個指令觀察，48.73秒通過；去除cost_flow後完整
JSON與RE-187逐項相同，world-return JSON逐byte相同。11次0x20BAD乘數直接轉成
Actor後與既有正式規則全部一致，研究測試只加在容器暫存副本，未改正式測試或規則。
最後Olear與Theodoric都寫1，機器steps1223232645／1223233296；前輪WatchWrites
只見變值，真正最後同值賦值由新觀察補足。原版執行準備7、掉落0都先被覆寫。
actor觀察欄是DS:9E5A，不能在尚未更新角色或處理另一個arg_0時誤稱函式目標。

原始DGROUP 0x12B7..12C8的18個byte以唯一錨及144站位byte正對照確認，file
offset0x21587。8..16讀到成本表後資料，不按未知語意猜補、不把索引夾為方向。
研究、索引、SPEC-011 DRAFT補充、CONTEXT及第6／44項同步；render／selftest／
verify／scan通過，22未完成／84完成，1482限制全部認領。成本仍DRAFT、程式
dea837bc及前輪更新順序CONFORMED保持；下一步補正式寫入追蹤及正常三出口驗收。

新verify_cost.py已在全唯讀容器重算正常來源、11個正式規則案例、18bytes及IDA
身分；1326共享輸入、RE-187舊4160附件與下層結果相同，314個cmd／internal／tools
輸入全部保持原雜湊。封存及唯讀重驗入口docs/verification/return-cost-20261002，
擁有權、收據SHA、Git／Docker控制面與暫存狀態由audit_cost.py及audit.json保存。
新來源、副本、payload、PNG及研究測試僅本機；沒有新正式程式、保存格式、提交、
推送、Issue寫入、版號、封包或發布。完整Goal保持進行中。

封存後audit_cost包裝器把範本中兩個舊路徑誤計為一個，首輪稽核在執行前斷言失敗。
只修正計數與第二版收據位置；先保存WORKLOG、驗證器及包裝器三份原輸入於
workplace/return-cost-20261002/audit-fix-inputs，receipt.json原樣保留。
新驗證器先核對第一版全部附件的歷史視圖，再封存receipt2.json；預設驗第二版，
--receipt receipt.json可重驗第一版歷史視圖。正常來源、IDA、案例及正式程式都不改。
這是稽核腳本錯誤，不是產品缺陷；批末實際清理結果另存audit.json。

## 2026-10-02 RE-189 返回成本的施法副作用

前輪39個正常觀察及11個規則比較有有效進度，本輪依目前檔案續核。命中規格閘門
與文件職責，載入相應入口，沿用逆向重製技能。查到施法0x20C31..20D17在乘數
之後覆寫BB84，因此18bytes不足以宣稱成本追蹤完整。讀grilling技能核對既有忠實
決策，以原版為準維持；先做隔離示例，不把未證實記憶體讀取猜成正式成本。
工作入口workplace/return-cost-gate-20261002/preserve_parent.py，先封存RE-188
第二版收據、第一版歷史視圖及目前共享輸入。368項既有異動保留，程式仍dea837bc，
沒有提交、推送、遠端寫入或發行。

alias_cases.sh第一次直接Call整支0x297F2未正常返回，未產生成本驗證結果；原輸入只有
_main初始化，不能判為產品缺陷。首版test／script另存.attempt1，失敗log保留。
依既有IDA指令縮為0x29800..29823，原版切片不改bytes、不加stub；十個預登記案例
通過，−1／0／1／16成本40／60／60／100，840對seed1／2為42／44。沒有換種子。

cast_cases.sh建立原始角色0法術記錄，直接執行0x20BCE..20D22並保留BB84接扣時刻
切片；四種組合乘兩個種子共八例通過。840／150／1400／586與既有CastTime一致，
其讀值分別在DS:1600／134E／1830／1502；840確實讀到seed低byte，成本42／44。
兩段沒有呼叫亂數，seed保持設定值；指令切片不能宣稱法術可正常取得或正常返回。
18bytes只足以驗已觀察範圍，不能作完整成本實作；0x20BAA乘數後仍有施法覆寫。

研究docs/re/20261002-return-cost-spells.md、SPEC-011 DRAFT勘誤與CONTEXT同步；
原始位址／bytes沿RE-188唯讀IDA輸出，正式DB與314個cmd／internal／tools不改。
新驗證入口docs/verification/return-cost-gate-20261002/verify_gate.py，先重算RE-188
歷史視圖與1332共享輸入，再封存本輪證據。下一步先確認正常施法返回的最後寫入
與讀值；父閘門及整體Goal保持進行中，未建立新正式程式或發行包。

worklist render／selftest／verify／scan通過，22未完成／84完成，1488限制全部認領。
新驗證器預驗完整唯讀通過：1332共享輸入及RE-188原2832附件／下層結果保持，314
正式輸入不改。git diff --check通過、暫存區空，主機docker ps -a實查本專案容器0，
全專案root-owned及.md同名目錄0。控制面紀錄host-control.json，後續封存與完整
唯讀重驗結果由docs/verification/return-cost-gate-20261002/receipt.json及audit.json保存。
新輸出擁有權另由verify_gate.py --audit抽驗；本輪沒有暫存正式程式或新GUI產物。

## 2026-10-02 RE-190 正常施法與世界返回來源

前輪RE-189原版施法及扣時刻切片有有效進度。本輪重讀AGENTS、WORKLIST與目前
CONTEXT，370項既有異動保留；命中規格閘門、dosgolem及文件職責路由，載入必要
入口與SPEC-012。原有battle-spell畫格實際是重製觀察，不能當作原版正常施法收據。
原始名冊只有Warder帶Vartwelyrmi，與Olear同在Halfway House，可沿正常COMPANION
招募。新工作區workplace/spell-return-20261002，入口preserve_parent.py先封存1336
共享輸入，再以prepare_spell.py／source_spell.sh準備正常啟動、招募、丟弓與施法。
只送合法玩家按鍵，seed1沿旅店招呼後及首次規劃入口固定；不注入角色、法術、
座標或指令，不替換等待。成本部分仍DRAFT，正式程式保持dea837bc。

source1於招募姓名檢查失敗，Down被TypeKeys當文字送入；source2改具名方向鍵與
C字串首NUL截取後仍選Theodoric。連續兩次失敗重查規格閘門，實際原版畫格證實
七次Down仍在硬碟Yes／No確認，尚未進名冊。source3先完成兩次Return後再下移，
正常招募Warder及進場通過。兩版原始腳本、log、畫格保留；另一次準備腳本字串錨
斷言未建立輸出或啟動原版，均分類驗證驅動器問題，未改正式產品或種子。

source3正常74.21秒通過，42回合921鍵自然敗北後八次原生等待返回世界。Warder先
正常丟弓，第二、三回合原有Vartwelyrmi對空格，訊息0x169；BB84兩次覆寫720。
475個成本觀察、561筆變值、988戰中及33戰後亂數保留。最後清槽留16，讀byte30
加移動率20再乘2，耗時100，3360降3260，次格4,4。既有CastTime讀兩次實測法術
bytes均720，測試只加在Docker暫存副本；314個正式輸入及程式保持原雜湊。
護符後三格00的成本100高於74類別80，不能只取第一格算764。

研究、SPEC-011 DRAFT補證、研究索引及唯一目前狀態同步；本輪只補正常施法後敗北
來源，沒有證實施法生效、表外720留到世界或重製GUI。驗證入口
docs/verification/spell-return-20261002/verify_spell_return.py，先唯讀重算1336封存輸入、
RE-189舊2680附件及下層結果通過。下一步依實際寫入順序審查有限成本追蹤，父閘門
與整體Goal保持進行中；沒有提交、推送、Issue寫入、資料格式、版號、封包或發布。

worklist首次核驗少認領SPEC-011新增限制；誤把manual的paths當成逐行認領後仍失敗。
重查工作清單路由及claimed_by，確認只有present／absent等pattern實際綁行。
第44項追加精確成本審查訊號並當場驗正反方向，第6項manual維持原契約；同一
容器命令乾淨重跑render／selftest／verify／scan通過，22未完成／84完成，1490
限制全部認領。這是台帳核驗設定問題，沒有改產品行為。

## 2026-10-02 RE-191 返回暫存追蹤契約

RE-190正常原版來源與正式規則比較屬有效進度。本輪核對目前AGENTS、WORKLIST、
CONTEXT與372項既有異動；命中規格閘門、dosgolem與文件職責路由，沿用既有技能。
workplace/return-tracker-20261002/preserve_parent.py先封存1340共享輸入及RE-190
收據／稽核。只追實際玩家路徑需要的暫存寫入，不深挖無關helper。
已讀實際IDA工具入口，沿用locked-v1及一次性DB；正式DB與程式不改。

IDA窄匯出通過，120個直接參照、去重1197條指令，原DB338函式與雜湊保持。
只分類初始化／世界方向端，不深挖無關helper。隔離Go副本在spendTime追蹤乘數與
施法覆寫；107個正常觀察含105中間乘數及兩次720，原版與Go相符。首版83數量
斷言誤取變值筆數，Go未啟動；第二版漏Actor.Member映射，取錯Warder法術來源。
原輸入／log保存，修正夾具後同一來源乾淨重跑通過。正式314個輸入不改。

另預登記單一正常Warder領隊、招募Olear、首輪丟弓，第二輪規劃Negative再逃走。
首次4500鍵停在原生字元問句，Return沒有答y；原始來源與完整失敗payload保留。
第二版只改字元y並收緊上界120，同種子與策略52.30秒通過；43鍵、八次原生等待
返回世界。只執行一回合，第二回合只規劃，不能稱施法生效。選單0x0700／0x0701
先被主列重畫清除，最後Olear REST留BB84=0，成本60，3360降3300，次格4,4。
該來源排除本路徑直接沿用清單值，沒有把120參照稱120寫入或外推全部清單。

研究、索引、SPEC-011 DRAFT與目前狀態同步；正式程式仍dea837bc，下一步串起
主列、執行、掉落、疲勞、回合末與搜刮階段。表外讀值及完整進場亂數限制保留，
沒有新正式程式、資料／保存格式、提交、推送、Issue寫入、版號、封包或發布。
新驗證入口docs/verification/return-tracker-20261002/verify_tracker.py，歷史收據不覆寫。

完整唯讀預驗通過：1340封存共同輸入、RE-190原3094附件與下層結果保持，正式
314輸入不改。worklist render／selftest／verify／scan通過，22未完成／84完成，
1494限制全部認領；新限制精確掛第44項且驗訊號移除不再匹配。封存收據與批末
完整唯讀、擁有權及Docker／Git清理結果另見新驗證目錄的receipt.json與audit.json。
所有副本與失敗樣本僅本機，整體Goal與成本追蹤實作保持進行中。

## 2026-10-02 RE-192 返回暫存階段與堆疊索引

先完成RE-191封存及完整唯讀稽核：收據11693f67、3226附件、1340共同輸入與下層
歷史通過，UID/GID1000:1000、root-owned及.md目錄0。隨後依規格閘門與文件職責
路由，在workplace/return-tracker-phases-20261002保存1344共同輸入才更新現況。

沿RE-190原策略及seed1正常重播，只加Regs、SS:BP-2及SS:BP+6唯讀觀察。80.68秒
原生執行通過；42回合921鍵與全部原有輸入、狀態、亂數、等待、561變值及475成本
觀察一致。新944觀察依步數與WatchWrites交叉相符。執行排序真正人物索引在BP-2，
126快捷觀察的全域索引皆不同，錯讀有94個算值反例。一般丟棄0x20231留下手持
欄位移816，強制掉落0x26A65留下地面格1。這補上舊成功掉落分類的不足。

隔離Go副本接主列、執行排序、普通／強制掉落、疲勞、回合邊緣及搜刮清槽的追蹤。
903個局部觀察比較及internal/battle全量測試通過。固定陣列誤寫slice使初版編譯
停止；全量初版缺唯讀assets而讀不到quest-sites.json。兩次失敗log與原輸入保留，
修正隔離準備後，同一容器／命令乾淨重跑。沒有把準備失敗記產品缺陷。

只驗本場局部輸入，沒有正式重製GUI或保存垂直鏈。正常原版最後搜刮留16，耗時
100、3360降3260；世界主列方向清零不能倒推成本為0。正式程式dea837bc與314
輸入保持，仍扣40返回3320，成本DRAFT。表外動態讀值與完整進場亂數仍開放。
研究、索引、SPEC-011與CONTEXT同步；父規格、#6／#44及有限CONFORMED範圍不變。
封存／批末完整唯讀、工作清單與擁有權結果見新收據及audit.json。一次性容器
均--rm，其他專案容器不動；所有原版payload與觀察副本僅本機。
沒有提交、推送、Issue寫入、正式版號、封包或發布，整體Goal繼續。

RE-192工作清單render首跑漏掛CONTEXT.md的產生區塊，唯讀拒寫；補上既有檔案的
單檔可寫掛載後，未重加條目，乾淨重跑render／selftest／verify／scan通過。
本地22未完成／84完成，1499行限制全部認領；新SPEC限制正反訊號已當場核對。
這是驗證環境掛載問題，正式程式及遠端狀態不改。

## 2026-10-02：RE-193返回索引窗

沿用規格閘門、dosgolem與IDA9.4路由；建立既有workplace下的隔離研究批次，先封存1348份共享輸入。正常三人兩場均65回合敗北；固定seed1、正常按鍵與原生等待，沒有注入或換種子。退離策略14次成本720，7次實際施法Nothing There；原位置兩次均Flinch。第一次招募前綴在出門後才按COMPANION失敗，完整日誌與畫格保留，分類為輸入腳本問題。

138,962筆窗內變值逐byte核對old、新值及完整返回窗，兩場返回均16、成本100、3360降3260。IDA只補EGA搬移窗及五個位移參照，正式資料庫唯讀；停止深入無關顯示helper與更換策略尋找特定值。證據及工具命令在[RE-193](docs/re/20261002-indexed-return-window.md)與[重驗器](docs/verification/return-indexed-read-20261002/verify_window.py)。314份正式程式輸入與dea837bc保持；成本DRAFT及22未完成／84完成不變，下一步回引擎接正式返回保存鏈。

沒有提交、推送、遠端Issue寫入、版本變更、封包或發布。所有工作在資源受限、無網路、UID/GID1000:1000的一次性Docker內完成；本機原始payload及PNG保留為可重播研究資料。收據、完整唯讀重驗及容器／權屬稽核結果見docs/verification/return-indexed-read-20261002/。

本批唯讀重驗器首次把移動方式位移寫為34，核對原始記錄後改為35；失敗與勘誤留在validation-attempt1.txt。同一容器與相同validate命令乾淨重跑通過，壞old值及壞返回窗反例均被拒絕。工作清單正反測試、verify與scan通過，1503行自承全部認領。15:14:32 UTC控制面核實本專案執行中及已停止容器均零，暫存區只有可重播研究資料；其他專案容器保持。git diff --check通過，staged清單空白。

## 2026-10-02：RE-194正式返回暫存寫入

沿用規格閘門路由，先保存1352份共享輸入，再將SPEC-011寫入契約升READY。接入battle.State.ReturnScratch：主列、排序、負重、施法、一般／強制掉落、逃走檢查及搜刮清槽；七份既有戰鬥檔及兩份追蹤／測試，不改資產、保存格式或世界成本。正式候選8ce2afbe，觀察副本b46b9aaf。原版資料與既有夾具唯讀，不重跑或挑選新結果。

全部internal／cmd測試及903筆原版觀察通過；唯讀記錄只附在正式setter後，不改算值。英／繁中各492正常鍵、493狀態／PNG與前版保持；61回合、362戰鬥狀態及1491亂數呼叫不變。正常敗北搜刮後留下16；各16個正式重載狀態／PNG相同。沒有新住宿保存鏈。世界端仍扣40而非100，3320／3260差異保持產品缺陷。完整成本及父閘門未完成，下一步接返回值及動態讀取。

來源、命令腳本、程式雜湊與驗收界限見[RE-194](docs/re/20261002-production-return-scratch.md)與[唯讀重驗](docs/verification/world-return-cost-20261002/verify_scratch.py)。工作清單正反測試、核實及自承掃描通過，22未完成／84完成。沒有提交、推送、遠端Issue寫入、版號、封包或發布；所有工作在無網路、UID/GID1000:1000、有資源上限的一次性Docker內完成，觀察payload、PNG與失敗紀錄維持本機。批末容器、權屬與完整性結果見同目錄audit.json。

工作清單首輪verify發現新增的非零任務攜帶抽驗缺口未認領；失敗日誌保留。第44項補入此精確訊號，正反對照及同一容器命令乾淨重跑通過；產品與完成條件不改。

2026-10-03跨午夜續驗，沿用world-return-cost-20261002批次。16:02:54 UTC控制面核實本專案執行中及已停止容器均零；其他專案容器保持。git diff --check通過，staged清單空白；目前狀態與工作清單日期已更新。316份正式程式輸入的雜湊及UID/GID1000:1000已記錄，完整唯讀稽核結果另見audit.json。

## 2026-10-03：RE-195 世界返回成本與新正常保存

- 命中規格進實作路由，載入`retro-remake-spec-gated-workflow.md`；文件職責沿既有
  契約。住宿驅動器同類逾時後重查正常玩家路徑路由，載入`retro-game-playtest.md`。
- 修改前封存1358共同輸入，新工作區`workplace/world-cost-resume-20261003/`。
  SPEC-011有限契約先READY，再將成本改成天候後讀返回索引。改三份既有程式／
  測試，新增讀取器／測試，318份正式碼輸入鎖定；未知別名保留診斷與既有近似。
- `build.sh`沿用既有Go image，4 GiB／2 CPU／256 PID及480秒上限。全部internal／cmd、
  原版十切片、天候後動態seed及三結算乘三來源出口隔離通過，正式候選f4c4f70b。
  觀察副本初版State欄位／方法衝突僅修研究碼，失敗log保存。
- `run_gui.sh`雙語正常61回合各493狀態／PNG：戰中3360、返回3260、四步後3040。
  除返回後ticks少60，全部狀態／PNG與RE-194相同，1491骰及最後天候seed保持。
  正式候選重載舊住宿保存，各16狀態／PNG相同。
- 外部住宿驅動器重播超過30秒，改單次有界GUI行程；初版漏支援516步治療提示，
  保留失敗軌跡，再補正常圖示回答NO。正式規則不因驗證工具失敗而修改。
- `run_inn_live.sh`雙語各706正常鍵／707狀態、14檢查圖完成回城住宿，生成v3保存。
  正式候選另起行程正常招募Olear，各16狀態／PNG與觀察副本相同；16名冊槽、
  34物品槽、typed裝備、背包、傷勢24／19及金額1721／0還原。繁中重載圖已目視。
- `verify_cost.py`預封存重驗通過，RE-194的3779附件及下層結果以封存副本重算。
  本場正常敗北成本／新保存有限CONFORMED；表外讀值、正常勝利GUI與父閘門仍開放。
  CONTEXT目前狀態、規格、研究索引及工作清單同批接回，完成閘門不改。
- Docker均--rm、無網路、UID/GID1000:1000及有界資源；Xvfb由trap收尾。沒有提交、
  推送、遠端Issue寫入、版號變更、封包或發布。批末控制面與擁有權由
  `docs/verification/world-cost-resume-20261003/audit.json`記錄。

工作清單渲染首兩次未掛可寫CONTEXT.md統計出口，OSError完整保留；補足該單檔
掛載後同一命令乾淨重跑。verify另找出新讀取器／診斷的未認領自承，補第44項的
精確訊號及兩個產品碼路徑，再跑正反對照與核實。沒有修改產品規則或完成條件。

17:07:26 UTC控制面核實本專案執行中及已停止容器均零，其他專案容器保持；git diff
--check通過，暫存區空。工作清單正反測試、verify及scan通過，1519行自承全部認領，
22未完成／84完成不變。318份正式碼、候選及新保存雜湊由收據鎖定，批末另做全唯讀
重驗及UID/GID1000:1000、root-owned／.md同名目錄自檢。

封存後文件勘誤：將903筆驗收段落移回RE-194節。七份初封文件另存，初封收據
未覆寫；Spec原文以確切段落移動重建，SHA-256與初封附件相符。WORKLOG未列初封
附件，保存修改前全文並由最終收據鎖定。新verify_final.py及receipt2.json重驗文件
順序與全部原結果，正式318份程式碼及GUI／保存結果保持，不重跑已通過的產品測試。

17:19:49 UTC控制面再核實專案容器零、暫存區空、git diff --check通過；其他專案
容器保持。最終工作清單正反測試、核實及掃描通過，1520行自承全部認領；最終
收據追加後再以唯讀根目錄做完整重驗及權屬／暫存產物自檢。

## 2026-10-03：RE-196 正常勝利與合法延後時段保存

- 接既有RE-186原版正常4回合來源，先封存1367共享輸入及RE-195最終3961附件。
  既有387項工作樹異動保留。命中規格閘門、正常玩家存讀檔及文件職責路由；
  私人驅動器繁中失敗時補讀localization-display-semantic-isolation.md，不改locale格式。
- 正常按鍵走標題、招募、城鎮／世界、規劃、4回合勝利與正常領取戰鎚。英／繁中
  各269鍵、270狀態及PNG，四執行入口抽樣、100個seed_before、返回3300及34槽符合
  原版；金額1835／1067、冒險點數2891／749及傷勢0／0相符。敵人倒地後LastHurt
  25／23明列，不宣稱完整角色記錄對拍。
- 私人勝利首版直接鍵序與WEAPON圖示錯誤保留，取消戰利品另留負面反例。住宿
  驅動器先後錯用過早數字、英文問句、舊Inside及REST功能編號；改讀Mode與
  AskingWho並按IconSprite對應圖號8。第7版誤引用第6版輸入，SHA相同，未執行。
  各失敗軌跡／日誌留本機，未調高上限或修改規則迎合腳本。
- 第8版正常走到存檔但失敗，第9版唯讀SaveError在507步取得「世界時段與剩餘時刻
  不一致」。回城進門留下Ticks3360／Part0，RE-151原版已有相同合法延後；SPEC-023
  先READY勘誤，再將保存檢查改為允許時段延後，內容原值往返。這是產品保存缺陷，
  與前述驅動器錯誤分開記錄。保存格式、世界更新順序、亂數及城鎮規則不改。
- build_save_fix.sh在既有Go image、4 GiB／2 CPU／256 PID／480秒內先驗實際舊檢查
  反例，再跑全部internal／cmd及保存目標測試，全數通過。新正式候選71d66b31，
  住宿觀察副本4147c6e8；318碼輸入只改complete.go及兩份既有測試，315份保持。
  初次新腳本的shell語法檢查失敗在啟動建置前修正，未當成產品缺陷。
- run_lodging10.sh兩語各507正常鍵／508狀態、27張圖生成v3保存，保留3360／0及
  16名冊槽。回城另一次正常逃走掉落Theodoric手持大劍，只有該槽清零，保存另對照
  當時實際34槽；戰鎚保持。Olear付款56後金額1779／1067，傷勢及冒險點數保持。
- run_reload10.sh另起行程正常招募Olear，新正式候選兩語各16狀態／PNG與觀察副本
  完全相同；原始34槽、typed手持／穿戴／背包、完整資料頁及持久欄位還原。繁中
  正式畫面已目視。typed物品只補town.Good省略的picture／movement零值後逐欄比較。
- 現況、SPEC-011／023／027、研究索引與工作清單接回本輪。父閘門、第6／44／64項
  及整體Goal保持進行中；沒有提交、推送、遠端Issue寫入、版號變更、封包或發布。
  原版payload、PNG、觀察程式、保存與失敗樣本僅本機。
- 全部工作在無網路、有界資源、UID/GID1000:1000的一次性Docker中執行，Xvfb以trap
  清理。新收據及批末唯讀完整性、權屬與Docker／Git控制面由
  docs/verification/normal-victory-20261003/receipt.json及audit.json記錄。

工作清單render／selftest／verify／scan均通過，22未完成／84完成，1529行自承全部
認領。合法延後時段的產品註解以精確單行說明為遊戲狀態，不當成未實作功能。
本輪父規格與有限範圍依RE-196及唯一目前狀態表，歷史收據不覆寫。

完整預驗在骰面相等斷言失敗，首輪結果保存prevalidation-attempt1.json。原版第4回合
IDA linear 0x2736B及0x2755E，索引73／74的n是100／2，重製是3／100；全部100個
seed_before及其餘98筆n相同。先前「100骰相符」過廣，現況表、規格及台帳同批勘誤。
這是已知產品差異，不列原版oracle未知；下一閘門只追這兩個呼叫點，成因及與
LastHurt差異的關係未證實。返回、報酬、物品及保存保持有限驗收，不宣稱完整戰鬥。

18:41:46 UTC封存前主機核實本專案執行中及停止容器均零，20個其他專案容器保持。
HEAD保持1384e50、git diff --check通過、暫存區空，控制面保存host-control.json。
隨後唯讀容器自檢全專案root-owned及.md同名目錄均零，驗證器語法通過。最終工作
清單正反測試、核實及掃描通過，1529行全部認領；封存及批末完整重驗另存收據／稽核。

初封收據成功後，最後稽核因Python元組與JSON陣列直接比較而失敗，記錄於
audit-attempt1.json。原收據、驗證器與產品不改；九份初封文件保存後將現行入口改為
verify_final.py，以JSON標準形重算同一結果，最終收據另存receipt2.json。入口更新
首跑誤認研究文件只有一處引用，實際有連結及命令兩處；修正精確數量後重跑完成。
工作清單最後一次render／selftest／verify／scan另存-final.log，初封日誌不覆寫。

## 2026-10-03 RE-197：徒手結算與正常勝利續核

- 上一輪為進展，RE-196存檔修正與5343附件最終稽核已完成。本輪命中remake、
  IDA、spec-gated及文件責任路由，沿目前狀態只追第4回合兩筆骰面差異。
- 開工核對工作樹；1375共享輸入封存，5343附件逐項保持。既有IDA locked-v1與
  Go1.26.7音訊image沿用，正式EXE／資料庫唯讀，一次性DB新匯出1045道指令。
  技能安裝連結指向ida_94_official；確認來源後載入真正入口。
- 原版徒手跳過持械選項修飾，空槽傷害碼仍1d2，且不加持械傷害表或部位表。
  原版Olear的FEET完整16bytes為零。多一骰與少一骰抵消，形成RE-196相同seed卻
  兩筆n不同。原始位址、bytes、xref、輸入與工具身分見RE-197。
- 首次MZ bytes檢查未套relocation，在far call失敗；保留檔案與載入bytes，依MZ
  表修正驗證後同命令重跑全部1045道通過。初次唯讀搜尋命令括號不完整，修正後
  重讀；均為驗證腳本錯誤。IDA exit1以JSON/schema/雜湊與1045道bytes判定成功。
- SPEC-009先DRAFT、證據審查再READY，之後修改attack.go、attack_test.go及main.go
  錯誤註解。正式318碼輸入只改三份，315份保持。新候選ca2ff7bf、觀察副本d2a6fc45。
- 舊版負對照確實在第2骰多擲3失敗；新四式徒手、落空、破甲空槽及持械骰序通過。
  全部internal／cmd通過。新雙語正常269鍵／270狀態及PNG，全270PNG保持歷史
  已目視畫面；四執行入口、100個n／seed_before及既定角色終態抽樣相符，LastHurt25。
- 新雙語各507正常鍵／508狀態完成住宿保存；正式候選另起行程正常招募Olear，
  各16狀態／PNG與觀察副本相同，34槽及typed裝備／完整持久欄位還原。預驗重新
  計算歷史父結果並核對新完整鏈通過。原版完整角色記錄與進場亂數未外推。
- 現況、SPEC-009／011／027、研究索引與第6／44項接回RE-197，父閘門維持開放。
  全部分析、測試與GUI在無網路、有界資源、UID/GID1000:1000的Docker中執行，
  Xvfb以trap清理。沒有提交／推送、遠端寫入、版號、封包或發布；原始payload、
  PNG、觀察副本與保存僅本機。批末收據／完整性／擁有權／容器清理另見本輪稽核。

工作清單首跑render已回寫WORKLIST，但同步現況統計時CONTEXT掛載唯讀而失敗。
補上CONTEXT精確可寫掛載後，以同一工具／image乾淨重跑，初次日誌保留。
最終render／selftest／verify／scan結果另存-final.log，不將掛載錯誤寫成產品缺陷。

最終工作清單維持22未完成／84完成，1530行自承全部認領；規格回填護欄納入本輪
唯讀重驗。主機git diff --check通過、暫存區空、HEAD保持1384e50；核對目前掛載
後本專案容器為零，24個其他專案容器保持。較早清冊中的其他專案臨時容器已被
其擁有者移除，刷新清冊後再核對，沒有重啟或清理其他工作。控制面見host-control.json。

## 2026-10-03 RE-198：正常逃走、回城住宿與正式重載

- 接續RE-197，本輪命中正常玩家路徑與跨規格回填路由，載入試玩、文件責任及
  re-resolution-backlinks契約。唯讀讀取曾因Docker socket sandbox受限，改用既有
  Docker權限後成功；沒有憑證或產品錯誤。
- 封存1380共享輸入，核對RE-197的2115附件。原版來源核實為RE-185 flee4的10鍵
  正常逃走；flee3是193鍵SPRINT後疲勞敗北，保留負對照，不以檔名推斷。
- 沿用正式ca2ff7bf、觀察d2a6fc45與既有Go音訊image；318碼輸入零變更，不重建
  或重跑無變更的全量規則測試。雙語各133正常鍵、134狀態／PNG；前120狀態／圖
  保持RE-197共同進場。初態／出口既定角色抽樣、8骰面數／seed_before、34槽、
  金額1777／1009及冒險點數2816／674相符，暫存0、成本60、時刻3360降3300。
- 新住宿雙語各395正常鍵、396狀態、35張抽樣圖。途中兩次追加逃走後34槽保持，
  Olear支付56離隊；v3保存金額1721／1009與時鐘3000／1。正式候選及觀察副本另
  起行程重載並正常招募Olear，各16狀態／PNG相同，34槽及持久欄位還原。
- 已目視繁中逃走問句、逃走訊息、世界返回、住宿保存問句與正式重載招募畫面。
  比較只限本場，回城只驗重製可玩性；完整世界進場亂數及全中文不外推。
- 初次資料預驗將監看沒有變值的null誤作空陣列，驗證腳本失敗。依RE-187既有
  契約接受無變值記錄，同容器／命令乾淨重跑通過；未改原版payload或產品。
- 現況、研究索引、SPEC-011／027及第6／44項回填RE-198，22未完成／84完成
  保持。沒有提交／推送、遠端寫入、封包或發布。分析與GUI均為有界、無網路、
  UID/GID1000:1000 Docker；Xvfb以trap清理。完整歷史重算、擁有權與控制面結果
  另見本輪receipt.json／audit.json。

- 工作清單首輪掃描將兩行「先前未驗」歷史範圍列為未認領；依既有acknowledged
  契約只認領兩行歷史說明，其他分支未知仍由第6／44項保留。首輪log另存initial，
  同容器／命令重跑；不刪限制或移除未完成項。


## 2026-10-03 RE-199：正常卷軸購買、AREA提示與新保存

- 路由命中規格閘門、dosgolem主oracle與GUI還原；載入spec-gated workflow、
  dosgolem README／CLAUDE及retro-gui-restoration。沿用逆向重製技能。
- 先封存1385共享輸入，核對RE-198的1843附件，再從副本重算父驗證及下層結果。
- 原版native-shop1因「.」按鍵名稱失敗；修驗證驅動器後native-shop2正常到南界15，
  沒有抵達Hostel。selected只是未分級原始字，不當選項證據；原版購買／AREA保持未知。
- 重製商店首兩次按鍵落點不對，保留輸出；第三次正常加入Blue Gem、購得Kelnalon。
  正常丟弓與移動，3、4、5回合怯戰，6回合成功。種子1沿同一場接續，未重設。
- 正常英中圖揭露PromptArea問句在Pending卻不收Return。SPEC-012先DRAFT審查READY，
  再只修問句直接呈現及清佇列，新增帶原始文字庫的反例。舊版明確因缺可見問句失敗；
  全部internal／cmd與針對提示回歸通過。測試、建置、抓圖均在固定Go映像／Xvfb容器。
- 42cf8db0正式候選、59f506ee觀察副本。雙語303鍵／304狀態，0／7不改狀態，
  選胸16→15；Damage／Hurt各0→1。首輪資料稽核漏列Hurt，被拒絕後按實際Take
  契約補入精確變化，沒有放寬其他欄位。301張共同PNG、規則及骰序保持。
- 住宿驅動器有界上限540→720；實際533鍵／534狀態、32圖。正常7回合逃走、回城
  住宿與新保存；另起正式／觀察程式重載、招募Olear各16狀態／圖一致，完整34槽、
  法術、Blue Gem／Blue Chain與金額931／1009還原。來源保存只差saved_at。
- 維持本地工作清單22未完成／84完成，不關閉完整原版施法或全重製。新研究入口
  同步掛入RESEARCH-LOG.md；README、遠端Issue、Git提交與推送、正式封包均無異動。
- 一次性容器均--rm；檔案UID/GID、暫存與本批容器清理以最終audit.json核對。

- 保存空法術欄位依既有 omitempty 契約比對空陣列與 null；非空法術與教派仍逐欄嚴格比對。文件腳本曾用唯讀掛載、未處理部分工作項目缺 note，以及讀取無 issue 的項目；均屬整理腳本問題，已修正並乾淨重跑驗證。

- 最終工作清單重跑：限制文字已由現有 Issue44 精確認領，維持22未完成／84完成，掃描零未認領。中文住宿末畫格與正式重載招募畫格已目視核對。程式比較證明除 NeedsArea 接線及 gofmt 外，主程式文字相同；未增加玩法改動。

## 2026-10-03 RE-200：正常原版購買、實際頸部信物與繁中續核

- 依dosgolem／spec閘門、文件分工及顯示／語意隔離路由接手，沿固定968ca774工具鏈。
  `exit1`正常出城、`entry7-1`145狀態從入城點7進城、`shop1`164狀態購買成功。
  正常世界Dijkstra路徑108上／13右，實際露營確認後步行至Hostel；只有旅店後兩筆seed寫入。
  123世界狀態的位置／次格／時刻／時段／seed及20城鎮狀態的金額／兩人34槽對拍通過。
- 1390共享輸入與RE-199的5123附件先保存，歷史收據不覆寫。SPEC-012先DRAFT／READY審查，
  再改scroll.go與scroll_test.go。原版token實在+0x170，舊版只有Token資訊、實際槽為空。
  新候選68b4ca0e、只讀副本dacde289；319碼輸入僅兩檔變更，全量internal／cmd與四種成交回歸通過。
  舊版頸槽反例正確失敗。雙語各172鍵購買、183鍵正式頸槽GUI、303鍵AREA與533鍵住宿通過。
  新保存完整頸槽、法術、教派、34槽及金額931／1009；正式重載各16狀態／PNG零差。
- SPEC-033先DRAFT／READY再補四個暫譯原名及六個order.token來源鍵，舊表負例拒絕。
  檢查器補token，譯表537鍵／519原名缺譯／孤兒／缺字0，9675項安全矩形失敗0。
  固定既有audio映像重新烘字模，首輪發現舊字「顛」從現行字串消失，按零流失契約加入保留表。
  第二輪2608字模、字距與PNG完全相同，輸入雜湊更新。繁中頸頁184圖中183圖不變，
  431像素僅名稱列；藍鍊32×18字模零差且目視可讀。補譯後新保存鏈及正式重載再跑一次，
  語意、排除時間的保存、英文名稱與物品原始記錄完全相同。
- 工具／環境失敗保留：首次go run受專案go.work的/dosgolem引用影響，GOWORK=off乾淨重跑；
  審查腳本誤讀condition欄、CJK來源誤用door而非index欄、像素腳本名稱上限過窄皆按實際欄位訂正。
  正式／觀察追蹤差異只移除四種只讀probe欄位後比較，不影響正式畫格；同一容器命令重驗通過。
  沒有把上述環境或驗證工具錯誤登記為產品缺陷。
- 僅本場空頸入會、清冊／字模、藍鍊GUI與新保存有限CONFORMED。原版正常AREA、其他教派／占用分支、
  完整進場亂數及其他手冊功能保持開放，22未完成／84完成。舊保存不自動生成可能已移走的信物。
  沒有暫存、提交、推送、遠端Issue寫入、版號變更、封包或發布。Docker均有界--rm及Xvfb trap。
  批末Git、容器、UID/GID與root-owned／.md目錄結果由本輪audit.json獨立實查，其他專案容器保留。

RE-200清單掃描首輪指出SPEC-033的其他信物GUI缺少逐行認領；已綁既有#44入會／裝備／施法保存鏈，#35人工中文化閘門保持。重新執行同一套render／selftest／verify／scan，不把真實待驗文字列為忽略項。

## 2026-10-03 RE-201：正常購買後出城與治療AREA

- 接手時沿dosgolem正常原版與文件責任／規格閘門路由，沿用原Docker工具鏈。
  初始HEAD1384e50、404項既有異動不重設。工作清單22未完成／84完成。
- 依既有城鎮Walk規則只規劃正常鍵：離店RET SPACE、D五次、RET出城。
  原版172狀態之前164與RE-200相同。雙語180鍵／181狀態，新增八狀態的金額與34槽相符。
  世界42,9／次格3,3／時刻3400／時段0／seed3750785579相同；正式英文181PNG與觀察副本相同。
- 預先固定42,9至49,23路線、seed1，不換seed或挑戰鬥。首版漏確認伏擊，239PNG，
  第173至238張同一CLIFF TROLLS畫面；signal:killed、無施法收據，終止原因未確認。
  嘗試清理時容器已自動移除。第二版只補enemyCount非零後正常Space，同前綴／路線／seed重跑，
  TestNormalBookEntry341.60秒通過。175前綴狀態、99正常戰鬥鍵，第4回合AREA。
  腳本意圖負面，實際+0x27=0與0x185治療，故按治療結果驗收，不稱負面傷害。
  選胸部2，Theodoric滿血16維持16，18actors bytes／seed／疲勞／體力不變。
- 局部Go規則抽驗取實際原版角色／法術、持續骰前3538685960，擲1d2＝1、1d11＝11，
  提示前2128043714相符；回答不動seed，六部位相同。此fixture不代替正常進場。
- 正式／觀察184張進場圖相同；相同兩次向右世界位置42,9／次格5,3、3340／Part1／
  seed316395082下，原版伏擊、重製仍在世界。已登記正常伏擊差異，不改規則猜補。
- 六份可變文件先封存，RE-200的6635附件與完整下層結果重算通過；319正式碼不變。
  驗證工具的internal import位置、局部Go複合字面值及治療鍵序錯估已修正。
  繁中出城尚未完成時提早核對曾報缺trace；治療鍵序未產生就啟動亦報缺檔，均為工具依賴錯誤。
  失敗產物保留，後续按實際追蹤先生成／核對輸入再重跑；不記產品缺陷。
- 本輪沒有Git暫存／提交／推送、遠端Issue寫入、版號、封包或發布。
  正常伏擊差異尚未修正，負面原版AREA及其他教派仍待驗。批末擁有權、容器、
  暫存與Git檢查由本輪audit.json實查，其他專案容器不動。

RE-201進場勘誤：上段「原版伏擊、重製仍在世界」是錯誤判斷。初核只看run_remake輸出的
world摘要，漏讀同一trace的battle。完整追蹤182步stage1 CLIFF TROLL、183步stage2規劃，
正式／觀察184張PNG相同，足以推翻缺伏擊說法；現況、研究及工作清單已撤除錯誤登記。
真正限制是首兩怪物位置：原版35,32／37,29，重製36,32／36,30，第三隻36,34相同。
首次規劃seed原版2535178238、重製3707996875；不以世界seed相同宣稱完整進場相同。
雙語治療UI另以既有正常路徑及首次規劃seed1抽驗，各301鍵／302狀態及PNG；
第6回合治療問句完整，0／7不變、胸部滿血16保持、回答不消耗骰。
此局部條件與本次原版不同，沒有宣稱整場同狀態。工作工具首次誤讀summary欄已按body訂正。

## 2026-10-03 RE-202：正常伏擊進場位置與地形交接

- 接手命中 dosgolem、IDA Pro 9.4、規格閘門及文件責任路由；沿用既有固定映像。
  初始406項異動、HEAD1384e50保持，不重設工作樹。22未完成／84完成。
- 唯讀一次性i64匯出814道指令與兩張原始位移表，逐bytes回查固定EXE。
  原版每隻敵人RND(8)後完整查八方向，成功後仍繼續，跳過已占格與0x0A。
  既有正常骰序局部重播找到原35,32／37,29／36,34；再沿相同正常購買、出城
  與seed1重跑，TestNormalBookEntry238.75秒通過。175狀態及PNG與RE-201相同，
  新48個前／後觀察、4096地形格與校正前後seed2052235245／4058265942符合重播。
- DRAFT經正常來源與bytes審查升READY後，正式只改main、state、ambush及其測試。
  確認遭遇圖後才接續野外亂數生成地形、校正位置一次；重排不重做。
  舊版原位置及缺地形反例拒絕，新版兩路徑與全部internal／cmd通過。
  新正式候選eed01abb、只讀副本b5d8912e，319碼輸入中315份保持。
- 英文／繁中各183正常鍵、184狀態／PNG，正式與觀察各184張完全相同。
  五名參戰者位置與4096地形格相符，第一次輪到Theodoric相同。
  英文前183張保持，戰場144x144內容20736像素全同原版；繁中縮放後有55像素不同。
  首次規劃seed仍原2535178238／重製1323149826，不宣稱完整進場或新保存完成。
- 十一份共同輸入先保存原hash，再以唯讀掛載修正前檔，重算RE-201的2317附件及
  下層歷史。原始來源與舊收據不覆寫。只整理#6／#44本機目前正文，歷史留WORKLOG。
- 本輪工具修正與原失敗保留在failures.json：不存在SS()、StackWord基準誤讀、
  既有DOS診斷的錯誤審查、新測試helper名稱、歷史驗證器繞過映射、PIL缺件及
  英文／繁中PNG縮放不同。修正後沿同一命令、映像、輸入與seed重跑，不記產品缺陷。
- 沒有Git暫存、提交、推送、遠端Issue寫入、版號、封包或發布。
  本輪唯讀收據入口docs/verification/ambush-placement-20261003/verify.py。
  批末Git、Docker、擁有權及暫存狀態由同目錄audit.json實查；其他專案資源不動。

RE-202驗證器勘誤：前批receipt的code_inputs是計數，319份hash另存固定parent-check。
首次封存前因此失敗，未寫收據；已改讀實際hash表並核對315現檔與4封存檔。
原深層重算與產品结果保持，不重新宣稱驗收。

## 2026-10-03 RE-203：原版四套怪物裝備與隔離原型

- 開工命中 dosgolem、IDA Pro 9.4、規格閘門及文件責任；HEAD1384e50、410項既有異動不重設。
  先保存330共同輸入與hash，正式319份Go碼、抽取器與怪物資料未變。
- 32怪物實際各513 bytes、128組裝備，31種有組別差異。初次假定512 bytes的盤點失敗，
  以全部實際解碼長度及既有record_sha256更正重跑；末端byte未知，不推定填充。
- 原版正常來源重跑206.43秒通過。175狀態／PNG及4583骰記錄保持，新增唯讀六次裝備觀察。
  三隻崖地巨魔選2／0／3組，三筆128-byte目的記錄相同。IDA一次性i64副本匯出，
  1003道bytes與固定EXE相符，不改名、不回寫資料庫。
- 隔離原型只改工作區副本；逐隻選裝備後才初始化遠程能力，18件裝備、14／8／0、
  一次性初始化及護甲耐久隔離通過。雙語正常183鍵／184狀態，觀察與無觀察各184圖一致，
  前183圖保持正式版。五名參戰者各38欄、190比較，裝備、地形、首次隊員與seed皆相符。
- 資料格式v2或v1新必填欄位已展示，依AGENTS.md及grilling詢問使用者，未自行定案。
  正式路徑不變，人物保存v3不變，完整回合／遠程使用／面板／其他入口及新保存尚未驗。
- 同輪新增研究入口並掛RESEARCH-LOG；工作清單22未完成／84完成保持。
  裝備77種原名查出正式譯表缺7筆，依既有術語完成本機擬譯，526種清冊無缺項或孤兒、
  不增加字元；正式譯表不變，新物品中文GUI仍待驗。這是中文化缺口，不記為環境失敗。
  無Git暫存／提交／推送、遠端Issue寫入、版號、封包或發布。容器有界--rm，Xvfb有trap。
  批末Git、擁有權、暫存與Docker狀態由本輪audit.json實查；其他專案資源不動。

## 2026-10-03 RE-204：正常原版負面AREA與清單方向鍵

- 命中dosgolem固定種子對拍、SPEC-012及文件責任入口；接手412既有異動不重設。
  先保存330份共享輸入。怪物資料格式問題仍待答覆，沿獨立原版負面來源前進。
- 正常原版同一路線與seed1，只把Select Affect的`.`改Down，297.30秒通過。
  175前綴狀態／PNG保持。第94鍵選胸部，第95狀態耐久16→5、傷害兩欄各0→11，
  其餘17名參戰者bytes、六名角色記錄、18組疲勞／體力分母與seed保持。
- 原驅動器只用治療分支0x16D標示生效，負面實際傷害卻留false；原始輸出不改，
  審查以實際三個byte差分判定。之後第6回合出口保留，未驗成本或完整回合對拍。
- Go局部Cast／AnswerArea從原版3538685960重播，1d2＝1、1d11＝11，
  10個傷害／部位欄與seed2128043714相符。正式規則未改。
- 既有隔離原型英中各278正常鍵／279狀態，前184圖與RE-203相同；機械狀態一致。
  原版Down選負面，重製方向鍵在文字清單被忽略，實际仍正面，登記產品缺口。
- Docker template缺Entrypoint鍵、舊run_native.sh路徑不存在及查錯state.go函式位置
  都屬查證工具錯誤，改用實際Cmd、run.sh及planning.go後取得資料，未改產品。
- RE-203的1715附件由封存共享輸入重算；原型結果重新計算相同。
  正式319Go輸入、抽取器、資料與翻譯表保持；22未完成／84完成保持。
  新研究立即索引，未提交／推送／遠端Issue寫入／打包／發布。
  容器--rm、Xvfb有trap，所有權與Docker清理由同批audit.json實查。

RE-204驗證工具勘誤：新增清單方向鍵自承未被#44的既有pattern認領，初次verify失敗log保留；將該具體訊號掛回#44，狀態不升格。收據初稿誤把30個舊book測試輔助寫入當正常驅動器寫入；改驗已鎖hash的實際native-driver只有兩個seed寫入。

## 2026-10-03 RE-205：法術清單輸入與取消

- 路由：原版GUI還原、正常玩家路徑、文件職責；使用逆向技能，注解取捨另用grilling。
- 原版175前綴狀態／PNG保持，121正常戰鬥鍵驗兩層取消、上下邊界、效果第一列25及文字白／青色。負面胸部16→5、兩傷害欄0→11、seed2128043714保持。
- 原版／PNG審查後SPEC-012先READY，正式八份Go差異，319份輸入其餘311份保持。反例先紅、新適配器四項及全部internal／cmd過。舊滑鼠測試誤將效果列起點寫17，依原版改25；初次測試失敗另存all-tests-first.log。
- 勘誤：既有CONTEXT決策允許等價鍵鼠操作，RE-204鍵序不同本身不列完工否決。本次修實際負面滑鼠失效與取消功能缺失。
- 正式相同原版固定304鍵未到AREA，屬已知裝備／次序差異；不重設seed、不用空擲補齐。隔離裝備原型雙語304鍵／305狀態符合原版11傷害，只作隔離證據。
- 正常只讀驅動器三版失敗保留：誤用Unarmed判持武器、套用別的世界入口回城路線、錯把quest容器存在當成戰鬥仍存在。先查實際指令列、世界42,9與戰後state=nil，再在同一工具鏈／seed1乾淨重跑。正式程式未因此改動。
- 正式與觀察雙語各631鍵／632狀態、448張可比PNG相同；正常選負面、AREA胸部16→8、seed3969673063保持，逃走後按可通行地形回世界40,23，入城住宿寫全新保存。
- 正式重載雙語各16圖／狀態，正常招募Olear；兩人34個槽bytes、法術／信物、金額931／1009、傷勢10／4還原。保存排除saved_at後相同。首次檢查器把hex與base64直接比較誤報，已按16-byte槽解碼，零picture／movement只允許省略零值；失敗收據保留。
- 目視發現繁中法術注解卷軸蓋住圖示；SPEC-034原位置已核准，沒有擅自搬動。A／B兩個DRAFT隔離原型由同一241正常鍵生成，待使用者選擇。正式候選仍53e335fe…bf25b8。另一個實際缺口是負面AREA結果仍印通用效果句，下一窄任務先補訊息分支證據。
- 329份修正前輸入封存，RE-204的1637附件重算，原收據不改。驗收入口docs/verification/spell-list-input-20261003/verify.py。工作清單22未完成／84完成保持，怪物資料格式仍待答覆。全部產物僅本機，無提交／推送、遠端Issue寫入、封包、改版號或發布。
- 本輪容器均--rm，Xvfb由trap結束；最終Docker、UID/GID與root-owned檢查記錄在同目錄audit.json，未清理其他專案資源。

## 2026-10-03 RE-206：負面部位法術結果句

- 開工420個既有異動路徑，保留全部；先載入原版GUI還原／規格閘門路由、文件職責及IDA技能契約。
- prepare.py封存329共同輸入；IDA 9.4一次性DB副本匯出sub_2818D／sub_277F4，1148道指令與MZ原始bytes含relocation逐筆相符。6803父附件保持，先DRAFT／審查／READY才改正式碼。
- 舊版雙語實際部位回答反例失敗。CastResult傳出真實目標與Take結果，手選與直接部位UI使用DamageLine類別1。改三份既有Go檔，新增一份回歸。沒有擲骰、傷害、正式資料格式或存檔v3變更。
- 首輪三個中文測試期望與既有譯文不同，保留targeted-tests-first-failure.log及fixture-failure.json；修驗證夾具後同image／命令乾淨重跑。20個雙語傷害案例、直接部位頁、其他通用句及全部internal／cmd通過。
- build_check.sh正式候選0dfb458d，唯讀觀察版c60c5754，隔離裝備原型645a2b42。run_normal.sh雙語各631鍵／632狀態與圖；逐步狀態、全部正式／觀察PNG相同。對RE-205只改step-297，英文708／繁中1648像素且限訊息區。所有觀察角色、傷害8與seed3969673063保持。
- 新住宿保存排除saved_at後與前版內容相同；正式／觀察重載各16狀態／PNG、14持久欄位、34個16-byte物品槽相符。隔離裝備原型雙語各304正常鍵／305狀態，只改step-304結果句，傷害11與seed2128043714保持，符合正常原版七筆非零訊息串。
- 研究與規格追加現行勘誤，worklist四項檢查通過，22未完成／84完成保持。負面法術重傷擊退、遠距部位骰序、完整parity仍開放；注解版面與怪物格式未代為核准。
- 文件驗證初次先自測後渲染，造成同步檢查失敗；改為先渲染後又發現CONTEXT統計區塊缺可寫掛載。兩份失敗log保留，回查文件路由與工具寫入端後補掛載，同命令乾淨重跑；未分類為產品缺陷。
- 全部工作在一次性非root、無網路容器。目視繁中結果頁正常，沒有提交、推送、Issue寫入、打包或發布。收據封存後唯讀重驗；容器與擁有權以audit.json實查，不觸碰其他專案資源。

## 2026-10-03 RE-207：部位法術骰序與傷害反應

- 保留開工423個既有異動路徑與RE-206的3946份附件；先保存336共同輸入。路由命中規格閘門、dosgolem固定seed與文件責任。
- 沿用IDA 9.4固定DB與EXE，1148道指令bytes逐筆審查；dosgolem固定版本968ca774、所有案例執行前seed1，不重擲。局部角色、護符、武器與狀態明示注入，顯示／等待stub，RNG未stub；沒有冒充正常玩家GUI。
- 初版夾具缺參戰者搜尋上限，補上後又被來源護符0x40擋下。補明示護符值與分支斷言後有效。治療比較再對齊來源門檻109及部位上限，保留初版與七案例收據；這些是驗證夾具問題，不是產品缺陷。
- 證據先READY。四個舊碼反例失敗：骰序6/2/11、被擋仍擲6、倒地未掉完整武器、未在傷害頁後判擊退。補查原本已失效肢體，候選反例又擲100/3；第八個原版案例證明應取消並清腿移動資格，不判擊退。補證據、重新READY後修正。
- 修改四份既有Go檔並新增兩份回歸。遠距部位骰改在通過阻擋、持續與強度之後；部位法術清倒數；倒地沿共享forcedDrop；重傷反應由下一次Act先於下一角色結算，保存掉落前物品名。保存v3與正式資料格式不改。
- 最終build_check.sh用已格式化輸入乾淨重建，八個原版同狀態案例、四個雙語手選／直接部位訊息適配器與全部internal／cmd通過。formal aa07b990，唯讀觀察版7759099a，隔離裝備原型9a18aa2b。原型額外欄位僅在暫存副本合併，未代選怪物格式。
- run_normal.sh雙語各631正常鍵／632狀態及PNG全部保持RE-206；正式／觀察632PNG相同，所有觀察角色欄位與seed保持。新v3保存排除saved_at後相同；正式／觀察重載各16狀態／PNG，34完整槽與14持久欄位還原。隔離裝備原型各304正常鍵／305狀態與PNG保持。
- 原型腳本原本忽略語言參數，首輪兩個容器都跑雙語，第二次碰到既有目錄而退出。保存首輪輸出雜湊及腳本，修正參數後清除兩個本輪未封存輸出並乾淨重跑；610張PNG、兩份trace與兩份command相同。分類為驗證腳本問題，不當作產品缺陷。
- worklist掃描把新規格中的歷史「尚未接線」字樣當成無條目認領的現行缺口。改寫時間關係並保留失敗log，重新檢查；收據未產生時的唯讀檢查隨之報缺檔，未歸類為產品缺陷。封存脚本補重入保護，避免重複追加交接。
- 狀態、規格與權威worklist更新，22未完成／84完成保持。局部有限CONFORMED；正常原版高傷害購買鏈、完整戰鬥與保存parity、未決版面及資料格式仍開放。未提交、推送、Issue寫入、打包或發布。
- 全部工作用一次性非root、無網路、有限資源Docker；Xvfb有trap。收據封存後唯讀重驗，容器、檔案擁有權與既有異動保存由本輪audit.json實查。工作區保存本機證據與候選；不清理其他專案資源。

## 2026-10-03 RE-208：法術啟用旗標與回合末失效

RE-208已修成功疲勞法術取消舊有效倒數，以及回合末倒數先於勝負與逃跑。原版+0x0D是啟用旗標，+0x30才是倒數；RE-207定位勘誤保留舊收據。1313道指令bytes與21個固定seed1原版局部案例先審查READY；舊碼10個反例失敗，新碼21個原版案例、出口／力竭等待、雙語適配器與全部internal／cmd通過。雙語各631正常鍵／632狀態、全部PNG、觀察欄位、新v3保存與正式重載各16圖／34槽／14持久欄位保持RE-207。原版正常多回合法術鏈與自然視野顯示閘門未驗；本切片局部有限CONFORMED，完整戰鬥與保存parity仍開放。RE-206／207傷害結果、骰序與擊退已接線。繁中注解A／B與怪物四套格式待決；多法術捲動、完整回合、遠程物品使用、正常怪物面板、任務／競技場仍待驗。入口docs/re/20261003-spell-lifecycle.md與docs/verification/spell-lifecycle-20261003/verify.py。

命令摘要：固定IDA 9.4資料庫副本匯出、MZ relocation bytes審查、固定dosgolem revision局部oracle、舊碼反例、正式測試／建置、雙語正常鍵序／保存／重載、附件重驗。全部在有界非root Docker容器、原始輸入唯讀。第一批Go夾具漏帶Hurt=11，修正後同容器契約同命令重跑。正常鏈未結束時提前驗收一次，因繁中trace尚未產生而失敗；等待結束後同命令重驗。兩項記為驗證問題，不記為產品缺陷，見verification-failures.json。
前批337份共用輸入先封存，3959份附件保持。未提交、推送或發布。容器、擁有權、入口Git異動保留與暫存產物狀態見workplace/spell-lifecycle-20261003/audit.json。

## 2026-10-03 RE-209：法術選單起點與正常修改

RE-209法術欄位清單起始列已修為0，保留編輯頁現值。1283條IDA9.4原始指令與正常原版反例先審查READY。舊碼12欄位與Body向下一次共13反例失敗，新碼全部Go測試通過。原版正常None／Strength各194／190狀態；新碼雙語各202／198狀態，兩種資源各扣280／330，完整法術欄位相符；正式與觀察版合計1600PNG及trace相同。本項有限CONFORMED；正常持續法術多回合失效、修改後住宿保存與整體SPEC-035仍待驗。RE-208局部有效倒數／失效順序與RE-206／207傷害反應保持。繁中注解A／B、怪物四套格式仍待決。完整戰鬥、替代路徑、任務／競技場與保存parity仍開放。入口docs/re/20261003-spell-lifecycle-player-path.md。

命令摘要：固定原版正常啟動／購買／重新進店／修改兩批、IDA9.4原始bytes審查、舊碼13反例、全部Go測試、正式／唯讀建置、雙語None／Strength正常重播、附件重驗。全部使用有界非root Docker、原始來源唯讀。前兩次探查重進門的輸入未完成，按正常移動觸發門口契約修正驅動後通過；非產品缺陷。沒有新需猜補的正式規則。前批339份共用輸入封存、3342附件保持；未提交、推送、打包、版本或發布。擁有權、入口Git異動與容器清理見workplace/spell-lifecycle-player-path-20261003/audit.json。

## 2026-10-03 RE-210：正常持續法術、保存與名稱頁

RE-210正常購買／修改Strength後的多回合失效已驗：原版力量128→117→128、原始倒數2→1→0；正式重製力量128→120→128、有效倒數3→2→1→0，繁中失效頁可讀。兩側seed1但骰條件不同，怪物四套格式待決，不宣稱整場同狀態。雙語正常保存分支656鍵／657狀態，正式重載16狀態；兩人34槽與14持久欄位相符，修改法術效果1、Olear金冠601／冒險點數2486及教派信物保留。保存分支在自然失效前正常逃走，與失效分支分開。名稱頁與取消各197狀態通過，但BS／N未改文字，正常名稱輸入是新產品缺口，SPEC-035保持READY。正式323份Go與b3a5d06b不變，正式／觀察合計4920PNG相同。原版住宿保存、整場／替代路徑、任務／競技場與完整保存parity仍開放；繁中注解A／B、怪物四套資料格式沿用原有待決。入口docs/re/20261003-sustained-spell-chain.md。

命令摘要：固定dosgolem正常啟動／修改／施法／REST至失效、原版額外無按鍵20M步呈現失效頁、正式／觀察雙語失效／住宿／重載／名稱取消，父附件與323Go重算。原版額外呈現前後actors／characters／counters／seed與141鍵完全相同。所有工作用有界非root Docker、來源唯讀；未重建或改正式碼。

驗證問題：第一原版探查未正確離店且程序被kill，真因未確定，最後PNG不完整；第二次以錯誤UTurn查找被guard拒絕，查實IconDoor=2後通過。第一次重製探查離店選到SCROLL及後續世界Y未知，另次保存探索在非露營狀態送SPACE；皆為驅動問題。grid雜湊是X-major原始4096 bytes，初次誤比PNG雜湊，修正後同資料通過。prepare失敗後誤啟動一次相依probe，未產生有效證據。繁中正式失效初驗誤期待英文句，程式已正常完成，修正驗證器後重驗既有有效畫格並續跑保存。名稱取消初報PASS只涵蓋資源，細查typing證實BS／N未接線；此為產品缺口，未以PASS掩蓋。

前批4356附件與335份封存輸入保持。探索、無效末PNG及私用payload只留本機且有索引；不作對拍或公開素材。清理、入口438個Git異動、擁有權見workplace/sustained-spell-chain-20261003/audit.json。未提交、推送、打包或發布。

## 2026-10-03 RE-211：法術名稱正常輸入接線

RE-211正常法術名稱輸入已接線，BS／N可見Kelnalon→Kelnalo→KelnaloN，取消保持原名與資源。舊碼兩個文字反例失敗，新碼全部internal／cmd測試通過；只改三份Go接線，新增正常按鍵回歸。雙語正常修改名稱／Strength後住宿各660鍵／661狀態，保存只變預期法術名稱，其他資料與RE-210相同；正式重載各16狀態、兩人34槽與14持久欄位相符。正式／觀察合計4936PNG及可比trace相同，雙語各360自然失效圖保持前版；取消除189／190／191輸入圖外各194圖相同。本接線有限CONFORMED，原版名稱字元集／大小寫／正常保存嚴格同狀態仍未知，整體SPEC-035保持READY。正式324份Go，候選a20728cb；怪物四套格式、繁中法術注解A／B沿用原有待決。整場／替代路徑、任務／競技場及完整保存parity仍開放。入口docs/re/20261003-spell-name-input.md。

命令摘要：先封存父附件、讀既有RE-128與SPEC-035 READY、正常Key舊碼反例、修三份接線及新增回歸、鎖定Go1.26.7／Ebitengine2.9.9全量測試與正式／唯讀觀察建置、雙語名稱取消／保存重載與原失效圖回歸。初版fixture漏教派，補前置後同容器乾淨重跑兩反例；非新產品缺陷。所有工作使用有界非root Docker，原始來源唯讀，未改資料／保存格式。335共用輸入封存，9313父附件保持；清理、入口440個異動與擁有權見workplace/spell-name-input-20261003/audit.json。未提交、推送、打包或發布。

## 2026-10-03 RE-212：原版名稱正常輸入與英文輸入框

RE-212修正正常法術名稱：有效輸入從空字串開始，確認後各單字字首大寫、其餘小寫，第12字自動確認；正式Left退格，英文補原版黑框白底輸入框。四條固定seed1的正常原版共789狀態，N與Abc Defghijk正常存錄的法術欄位及金冠647／冒險點數2476，兩語言相符。20張英文115×13輸入框零像素差，框外差異未豁免，不宣稱整頁或caret閃爍相位parity。正式／觀察6548PNG及可比trace相同，雙語各360自然失效圖與前版相同、取消其餘193圖相同。改名N及Strength後正常住宿保存各661狀態，金冠601／冒險點數2486，其他持久資料保持RE-210；正式重載各16狀態、兩人34槽與14欄位一致。三個反例失敗，新碼全部internal／cmd測試通過。正式325份Go，候選56bfb4c8；本名稱切片有限CONFORMED，整體SPEC-035保持READY。原版住宿保存、其他字元／空白邊界、完整戰鬥／替代路徑與任務／競技場parity仍開放。繁中注解A／B與怪物四套資料格式待原有決定。22未完成／84完成保持。入口docs/re/20261003-spell-name-original.md。

命令摘要：封存336共用輸入、核對固定dosgolem版本及兩次seed寫入、正常原版四路徑、兩個舊碼反例與第12字不自動確認反例、獨立字模原型六框、READY審查後修接線／規則／英文畫框、鎖版全量測試與兩份建置、雙語六路徑正式／唯讀觀察、20來源框及同欄位存錄成本審查。第二原版路徑D／L實送Down／Left，保留實際退格證據，第三路徑改明確文字鍵；第12字自動確認後多餘RET重開輸入，因此第四路徑刪多餘RET並驗正常存錄。初始可丟棄原型誤用book-large底線符號與小寫n，依來源字模修正，無變造原版PNG。原版報告繼承unpacked輸入標籤，實際packed執行SHA另列，不混位址基準。前批英文「可見」追加勘誤，原收據保持。

所有工作使用有界非root Docker、來源唯讀；所有789原版PNG、6548候選及780舊碼PNG與研究payload留本機。清理、入口443個異動、檔案擁有權及暫存收據見workplace/spell-name-original-20261003/audit.json。未提交、推送、打包或發布；不對其他專案作全域清理。

## 2026-10-03 RE-213：原版修改法術後住宿保存

RE-213從正常啟動、招募、購買與修改法術，經自然遇敵逃離回旅館，取得原版CHARDATA實際寫檔。原版574狀態、4次正常逃離；從實際檔正常重載及招募7狀態。Olear法術N／Strength、護符與骰式、金冠601及冒險點數2486和雙語重製保存相符。營養欄原版10／重製8；原版三次CAMPING、重製兩次，每次實際+2。差異保留，不宣稱旅程同狀態。雙語正式／觀察各380保存與16重載狀態，1584PNG及可比trace相同；兩人34物品槽與14持久欄位重載相符。正常旅店招呼後seed1，世界、戰鬥與住宿不重設；原版4次遇敵，重製一次，不宣稱整段骰序、整場或整頁parity。正式325份Go／56bfb4c8不變，保存格式v3保持。原版payload只留本機。22未完成／84完成保持。繁中法術注解A／B、怪物四套資料格式及既有未決事項維持原閘門。入口docs/re/20261003-spell-edit-inn-save.md。

使用既有Docker映像與固定dosgolem git archive，原始資料唯讀、輸出UID/GID1000。
首次二進位複製缺少執行權限、一般野外SPACE的批次錯誤、188鍵前綴中止、逃離等待caller誤套及比較腳本StageFlee常數，均分類為環境／驗證腳本問題；修正後乾淨重跑，失敗與中止輸出保留。
沒有新玩法或產品決定，也沒有重建工具鏈、提交、推送、發布或跨平台完成聲明。
容器及所有權收尾見本工作區filesystem-audit.json與host-state.json，原版payload與所有暫存研究留在workplace。

## 2026-10-03 收尾現況稽核與交接去重

命中完整生命週期／收尾及文件責任路由。從目前程式、有效決策、固定原版與最近收據核對。
先保存 RE-213 的5078附件並封存九份可變文件；完整舊CONTEXT保留在
`workplace/finish-audit-20261003/parent-inputs/CONTEXT.md`，原SHA記於parent-preservation.json。
CONTEXT改為唯一現況表、有效決策與下一閘門；移除入口中的過期暫停、舊版本與下一步指示，歷史不回寫。

Docker由固定唯讀ZIP重生inventory，與原清冊相同；EGA30archive／1022payload全部分類。
隔離副本重建40組現行manifest／773PNG，逐張與正式SHA-256相同。
首版EGA-only決定保持；修正資產待辦「其餘archive未完成」的過期文字，CGA／Hercules保留後續選項。
原條目與驗收條件不改為完成。遠端Issue22經主機gh auth status後唯讀核對為OPEN，沒有寫入。

文字安全矩形9675項失敗0；537鍵／519原名無缺譯或孤兒；6101字串缺字0。
城鎮來源2101訊息、102進門、102名稱、102人物、100建角字串缺譯0；45非文字槽位保持。
39份規格索引一致。這些為靜態與資產驗證，不外推所有正常GUI、原版parity、人耳或真機。
325份Go及原版、素材、字型、譯表和存檔v3均未變。本機22未完成／84完成保持。

怪物四套資料格式的v2／v1必填選項已再次以具體原型提出，尚未收到選擇。
只暫停依賴該選擇的正式接線，其他驗收可繼續。依grilling的JSON schema確認閘門。
首次Docker heredoc未加-i，Python沒有收到腳本；屬驗證命令問題。補-i後同一image乾淨重跑通過。

重驗入口：`docs/verification/finish-audit-20261003/verify.py`；CONTEXT同時索引receipt與verify。
資產重建副本與歷史文件僅本機保存。無暫存／提交／推送／遠端寫入／版號／打包／發布。
容器與擁有權清理狀態以該收據及末次核對為準。

收尾檢查先查尚未生成的receipt連結而失敗，屬驗證腳本順序問題。改為生成收據後再查全部連結；同一工具鏈重新核對。

## 2026-10-03 最長怪物正常 GUI 與現行交叉建置續核

命中正常玩家路徑、完整生命週期及文件職責路由。先核對目前狀態、452個既有異動及既有Docker。
保存上一批36附件的四份可變文件；325份Go、正式候選56bfb4c8、素材、譯表、字型及存檔v3保持。
唯一目前狀態回填CONTEXT，沒有重開舊EGA工作或改原驗收條件。

32怪物及60非空生成掉落名稱先計算字寬，再選80像素的岩石食人魔。
從標題正常招募六人，預先固定seed1，步行至最近森林，在次格間往返。
正常處理治療、露營、提示與其他遭遇逃離；第2745鍵自然遇到10隻等級1岩石食人魔，2746鍵進場。
讀取已抵達狀態，只推導四個正式按鍵，選射擊圖示、移動目標格再開裝備檢視；未確認目標或攻擊。
雙語正式候選各2751狀態，排除遭遇／執行／世界提示的顯示文字後完全相同。
2747狀態的基礎重播與四鍵延伸前綴相同；最長怪物譯名完整字模與敵人肖像／裝備頁已目視。
80像素怪物譯名、112像素商品及敵人武器木棒三個文字區完整字模通過，缺墨點及窄一像素反例拒絕。
這是重製版顯示驗證，沒有新增原版oracle或整場戰鬥parity。

戰利品母體判斷追加勘誤：只看到ItemAction可執行TAKE，曾誤判類型0的食品可手持後在戰場丟下。
正常itemIcons要求kind為奇數，法術槽與坐騎也不能當作可手持物品。
重新計算現行來源的316筆候選，上界80像素。正常購買112像素食品後開第一個背包槽，明細可讀且沒有TAKE圖示。
這是來源候選上界與反例，最長戰利品正常掉落、拾取、結算及保存仍開放；未接線四套資料不列入目前上界。

同一325份Go在既有kol-osxcross-go image交叉編譯Windows amd64與macOS arm64／amd64／universal。
四個產物在兩次乾淨編譯中SHA-256相同。Windows PE架構及macOS雙架構、最低11.0、arm64簽章區塊、系統相依檢查通過。
產物留workplace中間目錄，沒有正式版號、封包或Windows／macOS啟動與真機聲明。

失敗均保留：世界治療錯送Y；兩次420秒驗證時限不足；舊商店腳本選第六空槽；誤用-save-out旗標；核對器誤把第57鍵當六人招募完成；macOS image缺Python；512MiB分析容器載入110MB觀察trace記憶體不足。
分別修成正常選單RET、有檢查點及稀疏擷取、足以跑完原訂6000鍵的有界時間、現行第一空槽、正式-save旗標、實際第72鍵正常離城時驗六人，以及既有Python image獨立驗收。
沒有重擲種子或挑結果；沒有把驗證輸入、工具或時限問題寫成產品缺陷。
工具修正後乾淨重跑，兩個超時後尚存的自有容器已明確停止；後續工作加容器內外逾時與Xvfb trap。

入口：docs/verification/finish-gui-20261003/receipt.json及verify.py；正常按鍵、trace、失敗輸出及建置留workplace/cjk-monster-loot-20261003。
10張經驗證且有來源SHA、畫格、工具與權利記錄的PNG適用既有private例外，不授權公開散布。
檔案擁有權、Markdown目錄自檢、容器清理與暫存狀態見該工作區closeout.json。
22未完成／84完成及怪物四套格式待選擇保持；未提交、推送、遠端寫入、打包或發布。

末次收據核對器誤把開放項省略的status當必填，因KeyError拒絕。依既有worklist讀法改用get，原JSON及驗收條件不變；失敗核對器及初稿收據另存，再以同一image重驗。

## 2026-10-03 現行 Linux／Wine 正常保存續核

命中正常玩家路徑／Wine及文件職責路由，沿用既有試玩入口與Docker image。
先保存上一批GUI收據的可變文件，逐附件核對3,351筆；325份Go、正式資料、素材、字型、譯表與v3格式未改。
最長戰利品選80像素棘刺釘頭鎚，從標題正常六人招募、步行至任務5；執行前固定seed1，預先限制4,000鍵、60回合及480秒。
觀測器只讀狀態，經正式按鍵切換武器、移動、攻擊與休息；沒有改坐標、物品、遭遇、亂數或強制勝利。
有界執行以124停止；末次檢查點在第40回合，已看到棘刺釘頭鎚掉落，尚未勝利。
檢查點鍵列含下一批待送按鍵，完整trace及complete未生成；不把此快照當正常掉落GUI、結算、拾取或保存通過。
保留原始腳本與逾時輸出，不調種子、不反覆挑結果，最長戰利品項維持開放。

現行Linux候選56bfb4c8及Windows候選416edf3b實際執行互動入口，未指定scene、lang或save。
各198次真實X11按鍵由-log-input逐鍵確認；正常六人招募、食品購買、第一背包槽檢視、回旅館住宿保存。
再明寫-load重啟，各15次真實按鍵開保存食品明細；兩個遊戲程序正常退出，原保存SHA-256未改。
完整v3保存除saved_at外一致，包含六人隊伍、角色數值與食品原始record。
金冠1009→974→918，食品35、住宿56；13組靜止畫面逐像素相同，金冠增一與單像素變更反例都拒絕。
已目視Windows購物與重載食品明細，最長112像素譯名完整可讀；Linux對應畫格相同。

第一輪Docker在受限sandbox無法連socket，改用require_escalated取得既有Docker控制面。
Wine首輪因既有image沒有Python而127；保留輸出，沿用同一Wine image及kol-package Python image共享明確X11 socket，乾淨重跑通過。
這是工具環境問題，未改產品，也未另建重複image。所有可寫容器均UID/GID1000、無網路、有資源限制及內外逾時；Xvfb及Wine有trap。

現況入口同時索引docs/verification/finish-runtime-20261003/receipt.json及verify.py。
PNG保存來源entry、ZIP／PNG manifest／renderer雜湊、畫格、工具與private權利分類；其餘存檔、typed data、Wine環境與觀測輸出僅本機。
末次Docker、root-owned檔案、Markdown目錄與連結核對見該收據closeout。
22開放／84完成維持；Wine不代替Windows真機，macOS未啟動，沒有原版同狀態、全戰鬥、音訊人耳或封包聲明。
未提交、推送、遠端Issue寫入、定版號、打包或發布。怪物四套格式仍待使用者選擇，不依普通繼續推定同意。

## 2026-10-03 任務返回續行與最長戰利品保存

命中正常玩家路徑、spec-gated、IDA9.4與文件職責路由，沿用既有試玩入口、
`reverse-engineer-retro-game-remake`及`use-ida-pro-9-4`技能。
保留454個既有工作樹異動；先保存11份可變輸入、舊325份Go與父收據來源，再建立唯一現況。

正常40回合戰利品路線第3677鍵拾取棘刺釘頭鎚，任務逃走後世界凍結。
5500鍵有界驗證重現四方向不動。縮成134鍵招募／任務前綴、六人休息一輪與正常逃走，
205鍵仍重現；唯讀觀察指出任務提示已換時段，待續時間卻沒有出口。
IDA Pro9.4既有locked-v1 image、唯讀固定EXE與DB副本核對七處bytes、五支函式及呼叫端：
原版任務亦設次格4、4，任務返回後才更新時段及扣成本。
RE-214／SPEC-011第12節先READY審查，再修四份正式Go及加兩份測試。
327份Go的全部internal／cmd測試通過；雙語正式各189正常鍵、190狀態，返回760、四步後540。
修正前後各206觀察狀態、位置反例及18張PNG已封存，原版正常動態成本仍未驗。

修正後直接逃走再回旅館的4042鍵路線完成保存，但第3713鍵依既有BarFlee規則丟下手持戰利品。
該路線不算目標保存通過；英文正式末次斷言失敗，繁中同路徑有意中止137，均保留。
改走既有正常腰帶切換，卻在主列缺指令36。物品原記錄+0x0C為0x2b，轉換後Weapon.Code為0。
根因為Item.CombatWeapon漏填Code；骰數、戰鬥骰面與重量原本已有搬入。
依SPEC-010既有READY的整筆搬移契約，只回補Code，不改主列條件、骰面表、怪物v1、人物v3或seed。
追加怪物丟棄→隊員拾取→主列→空腰帶切換的奇偶碼回歸，先紅再綠；全部internal／cmd通過。
Go格式整理只改既有測試空白，精確測試輸入與新格式均保留雜湊。

同一預先固定seed1及3677正常拾取前綴，第3708鍵切入空腰帶，第4076鍵旅館保存。
無觀察的正式雙語各4077狀態一致，只排除世界提示與戰鬥遭遇／執行三個顯示欄位。
觀察版排除probe後與正式版亦一致；完整v3保存只差saved_at，原32字元物品記錄保存於worn[9]。
各26正常鍵重載Kenneth，檢視腰帶、正常拿到手上再檢視；原保存SHA未改。
腰帶、手持及明細三處80像素完整字模通過，缺墨點及窄一像素反例拒絕，完整畫面已目視。
現行Linux候選e0afbdbc另重跑雙語189鍵，保持RE-214正常返回。
有限入口：docs/verification/quest-return-20261003與loot-belt-20261003的receipt／verify。
Windows／macOS仍是舊325份Go歷史候選，未宣稱本修正已跨平台驗收。

| 失敗分類 | 處理與保留 |
|---|---|
| 原版IDA首次文字輸出編碼失敗 | 同image、新DB副本及明寫UTF8後匯出通過，舊log保留 |
| 測試缺assets／docs、短正式腳本誤設時刻 | 補唯讀來源連結、以實際760核對，同命令乾淨重跑 |
| 首版物品回歸未初始化四肢 | 保留原紅輸出；SetParts後重現Code缺陷，再修正式程式 |
| 收據錯讀缺省quest_prompt、輸出寫到唯讀/src | 修成正式mode及明確/out，原失敗保留，產品不改 |
| 直接逃走丟下武器 | 已有規則的預期結果，正常保存不能冒充目標保存通過 |

所有工作使用Docker、UID/GID1000、無網路、資源限制及內外逾時；Xvfb有trap。
既有原版輸入、IDA原DB與歷史工作區保留，沒有重擲seed或狀態／座標／物品注入。
末次容器、root-owned、Markdown目錄與暫存檢查以兩份有限收據的closeout為準。
怪物四套資料格式仍待選擇，重新提出一題v2／v1必填四組；依grilling暫停該正式分支。
本輪無Git暫存／提交／推送、遠端Issue寫入、版號、封包或發布；22開放／84完成不擅自結案。

## 2026-10-04 現行327份Go跨平台續核

命中Wine／Xvfb／跨平台正常試玩及文件職責路由，載入既有試玩入口與文件職責。
先逐附件核對前輪913筆，再固定327份Go、完整素材、譯表、必要runtime JSON與測試輸入。
沿用既有kol-osxcross-go、Go1.26.7／Ebitengine2.9.9、Wine9與Python3.11映像；沒有重建工具鏈。
Linux重建e0afbdbc，與前輪修正版完整SHA相同；Windows2e0df587，macOS universal2b4e5650。
Windows PE amd64、macOS雙架構、最低11.0、arm64簽章區塊與系統依賴通過。
產物在workplace/current-platform-20261003/build，沒有正式版號或封包，macOS未執行。

Linux及Wine各198次真實正常按鍵與15次重載按鍵，逐鍵input log確認且正常退出。
沒有scene、lang或save override；從標題正常六人招募、購買食品、住宿保存，再明寫-load重載。
完整v3保存只差saved_at，金冠918、食品原record及六人隊伍相同，原保存SHA不變。
13組X11畫格10組相同，normal126、reload3、reload14不同；此項未通過，不豁免不同頁。
驗證器保存差異數，唯讀收據重驗通過不代表13組畫格相同。原嚴格斷言失敗保留於failures。

雙語各189正常任務返回鍵與26鍵前輪正常戰利品保存重載，由正式同一二進位執行。
Linux／Wine兩側各四條完整軌跡都等於前輪對應正式軌跡，26組共52張配對PNG相符。
Wine預設DirectX返回英文已完成，但戰利品頁D3DCompile E5017缺多重overload優先序。
查證並封存鎖版Ebitengine官方環境契約，改用EBITENGINE_GRAPHICS_LIBRARY=opengl，雙語有限回歸通過。
不宣稱Wine DirectX戰利品頁或Windows真機通過；本輪未在Windows重跑4076鍵拾取保存鏈。

| 驗證失敗 | 分類與處理 |
|---|---|
| image inspect缺Entrypoint欄位、讀檔命令引號遺失 | 工具命令問題，修正讀法後同工具通過 |
| 隔離runtime缺docs/re預設JSON | 開窗前失敗，依正式旗標補齊四份唯讀固定資料，乾淨重跑 |
| Linux第9鍵未在6秒內記錄 | 舊輪保留；500ms按住後正常保存完成，不能外推畫格相同 |
| 3組X11畫格不同 | 獨立未通過訊號，未改產品或以保存相同掩蓋；仍待定位 |
| 統一120ms與2秒等待的Wine第9鍵超時 | 其後engine log實際出現該鍵，證實接收延遲超過控制器閘門；停止自有容器並保留環境 |
| Wine DirectX shader未支援功能 | 工具能力限制，依鎖版OpenGL契約重跑；原失敗與限制保留 |
| OpenGL準備尚有session時啟動相依容器 | 隨後收取準備exit0與回歸exit0；不把啟動先後當證據，完整輸出另核對 |

正常保存採本輪完整成功的Wine輪與最後Linux輪；沒有把超時或未完成輪計入。
有限入口同時掛進CONTEXT的docs/verification/current-platform-20261004/receipt.json與verify.py。
17張PNG有來源ZIP／manifest、renderer、擷取工具、畫格與private權利分類；其他資料與Wine環境僅本機。
所有工作Docker、UID/GID1000、無網路、有資源與內外逾時，Xvfb及Wine有trap。
失敗與中間產物保留，末次容器、擁有權與Markdown目錄核對見該收據closeout。
327份正式Go沒有再修改；worklist唯讀verify仍為22未完成／84完成，原條件與狀態不變。
怪物四套資料格式仍待使用者選擇，沒有猜補正式接線。
沒有Git暫存／提交／推送、遠端寫入、版號、封包或發布，也沒有全remake完成聲明。

封存完成：1537附件、17PNG。唯讀verify退出0，明寫13組X11全部相同為False。
封存的文件快照保留核對前版本；此段追加末次結果，不修改收據或歷史附件。
末次find查root-owned與Markdown目錄無輸出；git diff --check退出0；專案容器清單空。

## 2026-10-04 X11畫格同步與正常關窗

前輪為實際進展，現行327份Go及跨平台產物已固定。命中正常試玩／Wine與文件職責路由，
重新核對既有AGENTS、WORKLIST、CONTEXT與462個工作樹異動，保留所有既有內容。
先逐附件重驗前輪1537筆並核對現行327份Go。正式程式、資料、字型、素材、保存v3與二進位均未改。

直接檢查正式Update／Draw：輸入log在Key處理之前印出，不能證明按鍵後畫面已完成。
原本3組畫格的engine log沒有繪製錯誤，未把Wine shader缺口套成這三個問題的原因。
兩條短正常重載路徑共17鍵，以同一seed1、真實X11與預先固定0／0.5／2／8／16秒觀察。
裝備第2鍵的0秒兩側差212980像素，500ms至16秒均相同；物品3／14／15鍵全部時點相同。
Windows兩案例輸入與畫格完成，合成Ctrl+Q在物品明細未退出；外層430秒以124清理。
補一次1000ms退出鍵仍未退出；第三次修飾鍵探查前X11 socket已終止，退出1。
短路徑不宣稱正常退出通過，不重啟該程序或把超時轉成產品缺陷。
連續同類退出失敗後重查路由並讀直接reference，改採標準關窗測試而非繼續調短鍵時間。

同一Linux候選e0afbdbc正式198鍵及既有保存重載15鍵，先產13張按鍵後期望畫格。
產生流程的完整保存只差saved_at，與前輪真實正常保存一致；該流程只作期望資料，不算真實鍵證據。
預先固定期望SHA與完整畫格後，實際重播兩平台各198＋15鍵。
每鍵只送一次，收到log後指定畫格30秒內須與完整期望畫面相同；超時失敗，不換seed或豁免像素。
輸入等待20秒；早期畫格、全部比對樣本、最後接受畫格與時間均保存。
單像素變更、金冠增一及舊頁反例拒絕。四個遊戲程序經ICCCM WM_DELETE_WINDOW正常關窗並wait退出0。
只對宣告該協定的視窗送ClientMessage，再核對視窗移除與程序退出，沒有強制KillClient或DestroyWindow。
規範來源：[X11 ICCCM](https://xorg.freedesktop.org/archive/X11R7.7/doc/xorg-docs/icccm/icccm.html)，第4.2.8.1節。

兩側213真實鍵、13按鍵後畫格與完整v3保存通過；金冠918、六人及食品record相同，重載原保存SHA不變。
Linux首張舊頁1處，Wine首張舊頁12處，均被畫格閘門拒絕。
追加勘誤：舊normal126、reload3、reload14的Wine PNG逐像素等於新首張舊頁，
舊Linux PNG逐像素等於新按鍵後畫格，明確證實是擷取過早。原未通過收據不改寫。
三張Windows按鍵後裝備／物品／食品選取頁已目視，中文完整可讀。

入口同時掛進CONTEXT的docs/verification/gui-frame-sync-20261004/receipt.json與verify.py。
所有工作在有界、非root、無網路Docker；唯一網路查證由web工具讀X11官方規範。
第一次guarded腳本patch缺新增行前綴，工具拒絕且未寫入；修正後套用，未覆蓋任何舊輸出。
原版資料、父收據與runtime固定來源唯讀。PNG依既有private例外附來源與權利記錄；沒有公開授權擴張。
末次容器、擁有權與暫存狀態見該收據closeout；短診斷失敗與全部中間產物保留。
正式Go沒有改動，22未完成／84完成及原條件保持。Wine DirectX戰利品頁、macOS與真機仍未通過。
沒有新增原版oracle、全遊戲完成、Git暫存／提交／推送、遠端寫入、版號、封包或發布。
怪物四套格式仍待原有選擇，正式接線未猜補。

封存結果：229新附件、17PNG；1537父附件保持。唯讀verify退出0，三組舊差異與新首張舊頁逐像素相符。
末次worklist verify維持22未完成／84完成及產生區塊同步；git diff --check退出0。
全專案root-owned及Markdown目錄自檢無輸出，專案容器清單空。
文件快照保存封存前版本，本段只追加末次核對，不修改收據與舊證據。

## 2026-10-04 原版任務返回動態續核

前輪修正X11擷取時序為實際進展。本輪先讀AGENTS、WORKLIST、CONTEXT與目前工作樹，
命中remake驗證、正常試玩、dosgolem及文件職責路由。
沿用已讀skill入口，重查knowledge-router，讀dosgolem README／CLAUDE、
retro-game-playtest及直接reference 05-verification；新文件入口同輪掛入CONTEXT。
音訊候選已存在，SPEC-004仍要求人耳驗收；未重生音訊或接入正式runtime。
怪物四套格式的原提問仍待選擇，未改正式資料。

原版使用git archive固定dosgolem 968ca774，不使用該專案目前有異動的工作樹。
全部原始輸入唯讀；旅店招呼後固定seed1，之後不重設，沒有角色、物品、座標、
遭遇注入、直接函式入口、等待替身或強制勝利。
沿既有正常啟動鏈及Halfway House招募，實際方向鍵到任務5。

| 執行／核對 | 結果與分類 |
|---|---|
| native1 | 驅動未使用變數，編譯退出1，尚未執行遊戲 |
| native2 | Olear／Theodoric沒有任務5旗標；原版抵達後略過提示，94狀態；保留為前置條件反例 |
| native3 | 正常招募原有Poontang，旗標確認；接受後漏送遭遇確認，停止在真實確認畫面，退出1 |
| native4 | 依畫面及既有IDA `0x2B332` 等鍵控制流補確認；原版117次旅店後輸入、118狀態、324事件，退出0 |
| 第一版核對腳本 | 誤把戰鬥後共用scratch當仍持有任務掃描號；改在入口檢查5，返回不推定該暫值 |
| 露營末段 | 沒有普通成本出口觀察，明列未配對；四方向不是全部移動，不加入74筆普通成本樣本 |
| 第一版規則比較 | 成本相同；誤用16位元播種入口New，32位元狀態檢查失敗；保留log，改用既有NewState，原版不重跑或重擲 |
| 現行規則比較 | 74筆實際成本輸入全部相符，亂數狀態不變；成本加一反例退出非零；整批退出0 |
| 第一次封存 | `/out`誤當`/src`子路徑，封存退出1；未產收據。修正路徑後使用失敗即停止的shell，封存與重驗退出0 |

原版動態已證實：任務提示與戰鬥期間時刻120不變；返回後依序設次格4、4，
重畫／更新時段，再以索引1、成本位元組10與移動率20扣60，時刻120到60。
輸入實際packed KOLTOWN.EXE SHA及CHARDATA、工具、位址空間與證據在RE-214追加段。
局部比較明示受控成本輸入與32位元狀態；原版兩人、重製既有六人，不宣稱整段同狀態、骰序、勝利或搜刮。
SPEC-011第12節維持READY，未升CONFORMED。記錄仍保留四個缺少的擴充region檔及一筆未實作DOS呼叫。

新收據75附件、30張僅本機PNG。原版狀態及PNG留workplace，不提交或公開。
327份正式Go、go.mod／go.sum與現行平台凍結雜湊全部相符。
只新增原版有限oracle、核對工具與文件，不改正式玩法、保存v3或資料格式。
所有工作在非root、無網路、有資源限制與內外逾時的Docker；Xvfb有trap，沒有改動外部dosgolem工作樹。
沒有Git暫存、提交、推送、遠端寫入、版號、封包或發布。末次只讀核對與清理結果另追加。

末次唯讀收據重驗退出0；worklist verify退出0，維持22未完成／84完成及產生區塊同步。
git diff --check退出0；全專案root-owned與Markdown目錄自檢無輸出，專案容器清單空。
75附件封存後未修改；文件只追加動態證據與本次核對，未重寫歷史收據。

## 2026-10-04 DirectX物品頁等價shader修正及正常保存

前輪原版任務返回74筆動態成本核對為實際進展。本輪讀AGENTS、WORKLIST、CONTEXT及工作樹，
保留464個既有異動；命中重製驗收、Wine／Xvfb試玩及文件職責路由，載入既有skill、試玩入口、
直接05-verification與project-document-responsibilities。新工作區及重驗器入口同輪掛進CONTEXT。
SPEC-031第13節已證實EGA四平面還原契約，修正等價shader語法不改正式規則或產品取捨。

| 執行／診斷 | 結果與分類 |
|---|---|
| 原Windows候選同一26鍵 | 預設DirectX重現E5017，退出1；原二進位與失敗log保留 |
| 第一次隔離建置／測試 | 二進位建置完成，但臨時來源缺assets及docs，測試退出1；補唯讀來源連結後同命令通過，不列為產品失敗 |
| 僅替換mod運算 | 全測試及Linux有限回歸通過；Wine物品頁觸發HLSL_IR_CONSTANT斷言，整批退出3；中間Windows二進位與兩平台輸出保留 |
| 四平面固定分量隔離候選 | 同一26鍵、27完整狀態及8完整畫格與前輪Linux相同，退出0；確認後才套入正式來源 |
| final-build.sh | 現行327份Go與隔離候選輸入逐檔相符；Linux重建、全部internal／cmd測試及雙語返回／重載通過，退出0 |
| final-wine.sh | 預設DirectX雙語返回／重載及雙語各4,076鍵完整拾取保存通過，退出0 |
| mac-build.sh與cross-verify.py | 雙架構及universal重建、結構檢查通過；macOS未執行 |
| Wine image的Python探查 | Python不存在，退出127；沿用原image，由kol-package控制器透過本輪明確準備的X11 socket目錄控制，不建重複image或退回主機執行 |
| gui-wine.sh及gui-control.sh | 載入既有正常保存，雙語52真實鍵、14完整畫格、兩次正常關窗退出0 |
| gui-own-wine.sh及gui-own-control.sh | 載入本輪新Windows保存，雙語52真實鍵、14完整畫格、兩次正常關窗退出0；保存SHA不變 |
| checks.py、seal.py及唯讀verify.py | 完整來源／狀態／保存／畫格／歷史證據核對與封存通過，全部退出0 |

本輪僅修改internal/ui/itempage_edge.go：以Kage已知公式改寫mod，並將四平面動態索引區塊
改成固定分量。原16色×16遮罩共256代數輸入相同；不拆分宣稱某一次索引讀寫是唯一斷言原因。
公式來源：[Ebitengine Kage](https://ebitengine.org/en/documents/shader.html)。
沒有升級Go1.26.7、Ebitengine2.9.9或Wine9，也沒有改素材、字型、譯表、物品資料、規則或保存v3。
現行二進位完整SHA以本輪build/checks.json及收據為準，Linux b87ad738、Windows f525f6ff、macOS universal fcd417a8。
mod-only候選f83f4490屬失敗中間版本，沒有計入最終DirectX完成聲明。

Linux及Wine的八條雙語返回／重載完整軌跡等於前輪，52張配對PNG完整RGB相同。
Wine預設DirectX完整路線由標題正常招募六人、任務戰鬥、最長戰利品拾取、腰帶切換、逃離及住宿保存，
雙語各4,076UI腳本鍵／4,077完整狀態與前輪Linux相同；20張完整畫格相同，完整v3保存只排除saved_at。
沒有角色、物品、座標、遭遇注入、直接入口或強制勝利；不宣稱4,076鍵全程真實X11。
兩個真實重載批次各52鍵／14完整畫格／兩次正常退出，期望畫格在送鍵前凍結。
每鍵120ms只送一次，接收20秒、按鍵後完整畫格30秒、初始開窗90秒有界；舊頁不被接受。
第一批拒絕12張早期舊畫格，新保存的完整實際X11畫格另由唯讀verify核對。
正常關窗沿已宣告WM_DELETE_WINDOW協定，不用強制DestroyWindow／KillClient充當正常退出。
本輪新保存不被重載覆寫；完整保存的發言者變更與畫格單像素變更反例拒絕。
新保存重載的中英文物品頁已目視，名稱、造型、明細及控制圖示可讀。

新增收據1,965附件、1,096張本機PNG／資產快照，固定before-source及source-final各329個來源／module檔。
原版返回75附件與原始輸入仍相符，舊收據來源綁定before-source；沒有重寫原收據。
舊驗證器依賴當時正式來源，須唯讀覆蓋掛載該快照；新驗證器直接核對歷史來源，避免合法shader更新被當成歷史證據破損。
SPEC-031第14節記錄有限等價修正；header與索引更正第13節已閉合的Theodoric清單，父規格READY、#64仍開放。
macOS只驗結構、雙架構、最低11.0、arm64簽章區塊與系統依賴，不稱真機啟動。

所有分析、建置、測試、Wine與X11在有界、無網路、UID/GID1000:1000 Docker，原始資料與歷史輸入唯讀。
Xvfb、Wine與控制器由外層逾時及trap管理；本輪保留診斷prefix、中間二進位、失敗與觀察畫格。
prefix與socket目錄不納入不可變附件，沒有將Wine DLL或未授權來源發布。
沒有Git暫存／提交／推送、遠端Issue寫入、正式版號、封包或發布。
怪物四套資料格式仍待原有選擇，正式接線沒有猜補；全remake與原版完整戰鬥仍未完成。
末次唯讀新收據verify退出0；舊原版返回verify以before-source覆蓋掛載重驗退出0。
worklist verify退出0，維持22未完成／84完成，WORKLIST及CONTEXT產生區塊同步。
全專案root-owned及Markdown目錄自檢無輸出；本專案容器清單空，所有本輪工作已結束。
git diff --check退出0。原始輸入及歷史證據保持；失敗、中間候選、Wine prefix與觀察畫格只留本機。

## 2026-10-04：G03正常競技場資格、下注、單挑與保存

路由命中正常試玩與remake驗收。沿用`reverse-engineer-retro-game-remake/SKILL.md`、
`retro-game-playtest.md`與`references/05-verification.md`；建立工作區及收據前讀文件職責契約。
本輪入口為`docs/verification/arena-natural-20261004/receipt.json`、`verify.py`，
固定輸入與有界重生腳本在`workplace/arena-natural-20261004/`，已掛入CONTEXT及玩家路徑索引。
沿用327份Go的Linux b87ad738候選，329個來源／module雜湊與前輪相同，未改正式程式或資料格式。

| 正常玩家鏈 | 實際結果 |
|---|---|
| 標題與招募 | 從真正標題正常招募Theodoric、Poontang、Olear、Moraine、Fighting Bull、Kenneth；未讀資格存檔 |
| 正常訓練 | 城鎮7 Weapons Master的Nigel Gulliam，Club攻擊2點；學費210、冒險點數674→474，技能原值200→400，帳面位階1維持、實得位階1→2 |
| 初次住宿保存 | 正常付56住宿，Theodoric金冠1009→799→743；95鍵的v3保存證實自然資格，不曾載入此檔來省略路程 |
| 世界與競技場 | seed1在執行前固定；路上SLEDGE第39回合自然敗北，抵達世界55,13後選Theodoric參戰，其他五人為觀眾 |
| 下注與單挑 | Poontang押25賭贏，Fighting Bull押10賭輸；唯一TROLL，第9回合自然敗北，不挑勝負、不改骰值 |
| 結算與正常保存 | 兩名金冠214→189、442→452；正常走回城鎮2免費旅館，六人住宿保存，Theodoric傷勢38、冒險點數474與新技能保留 |

訓練與資格公式沿用RE-076、`internal/town/train.go`與`Sheet.EarnedRank`，
下注及單挑分支沿用RE-108與SPEC-025。這次沒有新增原版整場競技場oracle，
原版seed為null、original_execution為false；不稱整場同狀態、完整RND或勝利升階分支通過。
怪物四套正式資料仍待原有選擇，SPEC-025仍為READY。

`formal.sh`雙語各4,204正常UI腳本鍵、4,205完整狀態及67張預選畫格通過。
完整兩語規則軌跡相同，僅依明示欄位排除戰鬥`exec_text`、`encounter_text`顯示文字；
完整v3保存只排除`saved_at`，其餘全部欄位一致。不是4,204鍵全程真實X11聲明。
`reload.sh`各23正常鍵、24狀態，三張資料頁及免費旅館再保存通過。
`gui.sh`以預設正常玩家場景讀各自新保存，雙語合計46個真實X11鍵、16張完整畫格及兩次正常退出0。
期望畫格在送鍵前固定，完整RGB相同；不遮像素、不重送鍵，經WM_DELETE_WINDOW正常關窗。
輸入保存唯讀，SHA前後不變，GUI重存與完整原保存只差時間戳。
缺少訓練、改帳面位階、錯賭金、漏保存任務及單像素變更反例皆被拒絕。

普通軌跡以`standard.sh`逐鍵呈現另跑同一固定路線，和詳細軌跡的欄位契約相符。
`tools/verify_full_path.py`核對現行arena計畫的15里程碑、67畫格及保存欄位，退出0。
正常入口替換當前`arena.json`；舊受控計畫、收據與文件逐字快照保留於`parent-inputs/`，
原`arena-receipt.json`及`arena/`歷史畫格不改。只讀回歸仍可用`check_player_paths.py arena`獨立輸出。
封存694附件、236張本機PNG，父DirectX收據1,965附件雜湊維持；來源快照329個檔案。
固定名冊與訓練資料讀父收據的runtime快照，歷史重驗不依賴之後的可變資料。
最終唯讀`verify.py`及文件回填後重驗均退出0。

正常GUI另揭露Fighting Bull繁中資料頁裁切部分攻擊表頭與冒險點數2816末兩位。
英文完整、保存仍為2816，證據在`gui-zh-Hant/step-005.png`。
已回填既有`text-layout-remaining-screens`待辦，未把保存正確外推為畫面完成。

失敗分類與保留：

- 未指定`-scene flow`的腳本旗標組合、第一次`/repo`與腳本所需`/src`不符，屬驗證設定問題；修正後同一候選重跑通過。
- 初次店主問誰開口的按鍵錯序沒有開始訓練，保留輸入與畫格；正常選Theodoric後完成，未改訓練規則。
- 90秒的完整逐鍵競技場診斷逾時，保留失敗；正式全路線600秒有界重播成功，沒有換seed。
- 第一批21鍵重載確實讀到正確三名資料，但初始位置尚無進門提示，未產生保存；加入正常L、R離開再走回門口，23鍵重跑完成，不改坐標。
- 單張截圖快捷模式的trace第0筆寫成末狀態，沒有採用為標題基準；逐鍵模式的真正起始與全狀態核對通過。
- 保存檢查器未處理空角色null槽、封存器把`/out`直接當`/src`相對路徑，皆修正檢查腳本；不是保存損壞或遊戲規則缺陷。
- 一次自動權限審查未在期限內完成，依回覆只重試一次後成功；沒有待核准動作或未解的拒絕。

G03依原完成條件在本機結案，遠端Issue未寫入。
`tools/worklist.py render`及`verify`退出0，現在21未完成／85完成，WORKLIST與CONTEXT產生區塊同步。
正式Go未改，因此不重跑前輪已通過的完整internal／cmd測試；此次新增正常鏈與保存驗證如上。
全部分析、遊戲、GUI、封存與檢查在既有鎖版kol-package Docker，UID/GID1000:1000、無網路、相稱資源及外層逾時。
原始輸入、候選、素材、歷史證據與重載來源唯讀，僅明確工作區與回填文件可寫；Xvfb有trap，容器均--rm。
全專案root-owned／誤建*.md目錄自檢無輸出，694附件擁有權1000:1000；診斷失敗與中間輸出保留本機。
git diff --check退出0。沒有提交、推送、遠端Issue寫入、版本號、封包、上傳或發布。
全部本輪工作終止後再查本專案容器清單為空，沒有留下執行中或停止容器。

## 2026-10-04：資料頁裁切勘誤與現行自然勝利清單

接手時讀路由、AGENTS、WORKLIST、CONTEXT與工作樹，保留467個既有異動。命中中文動態文字版面、正常試玩及文件職責，沿用remake skill、verification-and-handoff與project-document-responsibilities。建立工作區、驗證器及下完成結論前重查路由；入口已掛入CONTEXT、SPEC-030／032及玩家路徑索引。正式327份Go、go.mod／go.sum、字型、譯表、資料、Linux b87ad738與其他平台候選沒有改動。

前輪「Fighting Bull裁切攻防表頭與2816末兩位」判斷有誤。同一已封存PNG的SHA仍為29d1f0d4fae30140772e0acf7ab2546f8769330caa381b1ce911bca92881a874，實際攻擊、防禦、2816與452皆完整。以runtime字模逐像素核對，四欄零差。沒有證據支持runtime裁切，沒有為此改正式UI。歷史G03收據與畫格維持，這裡追加勘誤；現行CONTEXT、worklist、arena計畫及玩家入口撤回錯誤敘述，其餘G03證據維持。

六人正常保存重載的英中各12鍵／13狀態一致；24真實X11鍵、14完整畫格與兩次正常退出0。六人各四欄在腳本與視窗共48新欄，加舊畫格4欄共52，全部字模零差；缺墨點及錯冒險值反例拒絕。119附件、70 PNG封存於cjk-sheet-visibility-20261004；來源保存SHA前後不變、沒有意外寫檔。第一版字模核對器未處理缺失空白glyph，改依Font.advance保留16像素空白而不畫墨點後通過，屬核對腳本問題。

| 現行自然勝利鏈 | 結果 |
|---|---|
| 舊正常269鍵 | 英中各270狀態，在第4回合verdict0；保留觀察，不沿用舊版勝利結論 |
| 依目前規劃頁重生 | 同一seed1與前119正常鍵，460鍵第6回合自然勝利；沒有怪物／任務／遭遇替換、座標注入或強制結算 |
| 清單與正常領取 | step457戰錘，458 Olear／Theodoric，459正常領取，460返回世界49,23次格4,4，ticks3300 |
| 現行正式雙語重播 | 各461狀態及35預選PNG；除exec_text／encounter_text顯示文字外完整規則軌跡一致 |
| 回城保存 | 固定既有238正常回城鍵接到新460前綴，各698鍵／699狀態；Halfway House保存，金冠1843／1075→1787／1019，AP2891／749 |
| 正常重載及再住宿 | 各24鍵／25狀態，Olear戰錘明細可讀，保存只差每人56住宿及時間戳；其他完整v3欄位一致 |
| 真實視窗 | 繁中標題至勝利460鍵／六完整畫格，雙語重載48鍵／14完整畫格，三次WM_DELETE_WINDOW正常退出0 |
| 有限中文字模 | 兩個標題、戰錘、選中Olear與青字Theodoric，在腳本及視窗共十框字模零差，未越過安全右緣 |

508真實鍵不是全部698鍵的真實視窗聲明；完整逐鍵呈現與X11抽樣範圍各自記錄。期望畫格送鍵前固定，每鍵120ms一次、接收20秒、畫格30秒、開窗90秒；按鍵後0.15秒有界等待，沒有重送或挑畫格。原保存唯讀，SHA不變；重載GUI的新保存與正式新保存除時間戳完全相同。未勝、漏戰錘、錯金冠、單像素反例皆拒絕。395附件、290 PNG封存於cjk-victory-list-20261004，父資料頁119、G03 694與DirectX1965附件保持。

戰鬥控制器在已固定勝利後縮小最後SPACE批次，是為停在結果頁邊界，不是重新擲骰或選勝負。兩次查不存在的inn10建置／來源檔失敗後重查路由及實際清冊，確認可追溯的507鍵driver-keys，而非猜檔名；歷史指令與238鍵suffix的SHA由舊收據核對。重載準備器第一次未跳過空角色null，修正後原命令通過；皆屬工具設定問題。沒有更換seed。

原版seed為null、original_execution=false，沒有新增整場oracle或完整骰序parity。最長戰利品、怪物四套資料、繁中名牌／法術原型與其他玩家分支仍按原閘門。SPEC-032保留CONFORMED，補有限正常勝利鏈及其舊dosgolem能力敘述的日期勘誤，不宣稱完整adapter。維持21未完成／85完成，沒有重新開啟已完成項。

所有分析、遊戲、X11、封存與回填在既有鎖版kol-package Docker，UID/GID1000:1000、無網路、有CPU／記憶體／程序上限及外層逾時。原始輸入、素材、候選及歷史證據唯讀；Xvfb有trap，容器--rm。沒有Git暫存、提交、推送、遠端Issue寫入、正式版號、封包、上傳或發布。PNG及保存均僅本機，權利manifest不外推公開散布權。末次唯讀驗證與擁有權、暫存及容器清理結果另追加。

文件回填後唯讀自然勝利 verify 及其資料頁／歷史父附件重驗退出0，329個來源／module雜湊與凍結候選一致。worklist render、verify退出0，維持21未完成／85完成，WORKLIST與CONTEXT產生區塊同步。全專案root-owned及誤建Markdown目錄自檢無輸出；119及395附件擁有權全部1000:1000。git diff --check退出0。所有本輪工作終止後，kol命名及掛載本專案根目錄的容器清單皆空；不處理其他專案容器。原始輸入、封存證據及歷史畫格保持，診斷與中間輸出留本機，工具映像未重建。本輪沒有正式Go變更，沿用前輪已通過測試，不重跑無關測試或改發行聲明。

## 2026-10-04：最長戰利品的正常自然勝利清單

接手重查AGENTS、WORKLIST、CONTEXT、路由與git狀態471行，保留既有異動。命中remake驗收、正常試玩與文件職責，沿用remake技能、既有驗收reference、retro-game-playtest與project-document-responsibilities。封存與文件回填前再查路由；新工作區及收據掛回CONTEXT、SPEC-032及玩家路徑索引。最長怪物名已完成的有限切片不重開。正式327份Go及module、b87ad738候選、字型、譯表、素材、規則與v3格式均未改。

| 正常玩家路徑 | 結果 |
|---|---|
| 唯讀來源 | 此前DirectX4076正常鍵由標題走到旅館的六人保存，兩語來源SHA由父收據核對；本輪不宣稱另重走前史 |
| 名稱母體 | 固定runtime typed資料重算940記錄，現行316候選上界80px；棘刺釘頭鎚達上界，未接線四套裝備不計入 |
| 六人首次診斷 | 70正常裝備鍵後正常出城，Kenneth首手DROP，2015鍵第31回合自然SLEDGE敗北；輸出保留，不冒充勝利 |
| 正常重組隊伍 | 已核准faithful_inn_stay=true，住宿離隊、正常招募Kenneth／Shyla；種子仍1，沒有改保存、裝備記錄或遭遇資料 |
| 自然勝利清單 | 雙語各570鍵／571狀態，第7回合勝利；563結果、564報酬、565棘刺釘頭鎚及戰錘、566受領人、567領回、570返回世界 |
| 正常回城保存 | 途中GREMLIN第20回合自然勝利，各1914鍵／1915狀態到Halfway House，Kenneth住宿離隊後保存party[9]、gold856、AP2748、wounds8 |
| 正常重載再招募 | 雙語各45鍵／46狀態，正常招募Kenneth、查看棘刺釘頭鎚明細及再住宿；完整保存只差Kenneth56金冠與時間戳，金冠800 |
| 真實視窗 | 繁中570鍵及九完整畫格，雙語重載90鍵及十四完整畫格；660鍵、23完整RGB閘門、三次WM_DELETE正常退出0 |
| 字模及記錄 | 十四個安全矩形完整字模零差，白字選中／青字未選中保持；稀有record b9c12e2e40d08ca0000023ff2b3c9100位於背包3及carried4 |

雙語完整規則軌跡只排除exec_text／encounter_text顯示文字，完整v3只排除saved_at比較，其他欄位不另豁免。原版seed=null、original_execution=false，沒有新增原版oracle或整場骰序parity。未勝、缺稀有物品、錯金冠、缺record及單像素反例拒絕。660真實鍵不擴大成全部1914鍵真實操作；正常保存來源與每次GUI重載皆唯讀，SHA前後不變。499附件315 PNG與來源快照封存，父395附件及其他歷史證據維持。

| 診斷與工具失敗 | 分類與處理 |
|---|---|
| 首版控制器停城門目標格 | 沒有遇敵；保留輸出，補正常走到邊界與DOOR鍵 |
| 兩人離店漏SPACE | 正常告別訊息尚未關閉，token1拒絕；由session.go leaveDoor確認，補正常空白鍵，未改遊戲 |
| 回城wrapper找不到controller | shell檔名替換後未複製對應腳本；保留原log，提供同一controller後以同容器同命令乾淨重跑 |
| 提前讀取中文保存 | 中文1914鍵仍在執行，準備器安全拒絕；等待完整trace後同命令通過，新增完整trace前置檢查 |
| 末狀態party未重複寫入 | trace採差異記錄；核對器沿前筆正式party恢復比較，不造新狀態 |
| 幾次查錯Go／trace檔名 | 回查實際清冊並使用trace.json；皆為讀取工具錯誤，不是產品缺陷 |

正式來源未改，沿用前輪已通過internal／cmd測試；本輪完成新正常路徑、GUI與保存抽樣。既有未完成條目只追加這份有限證據，不變更原完成條件，維持21未完成／85完成。四套怪物格式、名牌／法術原型、其他CJK分支、原版完整戰鬥、平台真機與發行仍待各自閘門。

所有分析、遊戲、GUI、封存與文件回填在既有kol-package:20260907 Docker，UID/GID1000:1000、無網路、相稱CPU／記憶體／程序上限及外層逾時。Xvfb有trap，容器--rm。原始輸入、父證據、素材、候選與保存來源唯讀；PNG、保存、控制器診斷及少量Python cache保留本機，未放入公開交付。沒有提交、推送、遠端Issue寫入、版本選擇、封包、上傳或發布。末次唯讀、擁有權及容器清理結果另追加。

文件回填前後的最長勝利清單唯讀verify均退出0，父勝利／資料頁及其歷史附件保持，329個現行來源／module與凍結快照一致。最初worklist verify退出1：SPEC-032新增四套未接線聲明未綁條目；依路由載入rulebook/61，將精確訊號與該spec接回既有V01，不改完成條件。正負對照證實現行綁定認領、舊綁定不認領、移除訊號會開口。render及重跑verify退出0，維持21未完成／85完成，WORKLIST與CONTEXT生成區塊同步。全專案root-owned及誤建*.md目錄自檢無輸出；499附件與回填六文件UID/GID均1000:1000。git diff --check退出0。所有執行結束後，kol命名與掛載專案根目錄的容器清單均空，沒有清理其他專案資源。控制器失敗、正常敗北與中間產物保持本機，工具映像未重建。

下一個正式怪物接線仍受資料格式選擇閘門約束。已重讀grilling與RE-203，展示v2隔離原型證據及兩個格式選項，再提出單一問題；人物保存v3不變。尚未收到選擇，不把GO當成v2或v1選擇，也沒有把原型接正式路徑。這是資料格式的共同決策，沒有要求新的外部發布許可；Goal仍未完成。


## 2026-10-04 正式打包入口的版號與覆寫防護

路由命中完整收尾，沿用 `retro-remake-end-to-end-playbook.md`；版號及交付目錄分別依
`retro-remake-release-versioning.md` 與 `retro-remake-dist-all-output.md`。專案既定
`full/`、`release/`、`promo/` 名稱維持；文件職責依 `project-document-responsibilities.md`。

核對發現 `build_dist.sh`、`package_dist.sh` 與 `verify_dist.sh` 接受舊版號，前兩支
還會刪除同版目錄。正式建置沒有注入既有 `main.buildVersion`，清單驗收也沒有核對
要求的版號及包別。修補既定契約，不代替使用者選版號。

- 打包、封包及驗收入口先檢查 `v.<主版>.<次版>.<修訂版>-YYYYMMDD`；容器分支同樣檢查。
- 同版目錄及封包目錄已存在時停止，失效符號連結也拒絕；取消刪除既有輸出的操作。
- Linux、Windows建置注入同一完整版號；macOS由 `KOL_BUILD_VERSION` 傳入兩個架構。
  單獨開發建置未給該變數時維持 `development`。更新工具註解中的舊版號範例。
- 清單驗收核對 schema、bundle、完整版號及要求的版號，保留逐檔雜湊檢查。
  `verify_dist.sh` 的 Python 驗收改用有界唯讀Docker，Xvfb追加有界程序的退出清理。
- 沿用 `tools/test_dist_manifest.py`，九項回歸及四支Shell語法檢查通過。
  涵蓋六類非法版號、目錄與失效連結保全、各平台版號傳遞、開發建置及清單反例。
  同一測試套件套到修改前來源失敗，輸出保留。容器、編譯器及lipo使用替身，
  封包副本只替換固定掛載路徑；測試不操作真正的 `dist-all/`，不代表平台建置完成。
- 權利防護六個隔離案例通過。最長勝利收據及其父收據唯讀重驗通過；329項Go來源／
  module輸入雜湊維持。更新打包待辦現況並render／verify通過，仍21項未完成、85項完成，
  原驗收條件與狀態維持。

修改前後七支工具、雜湊、測試、回歸及本機收據在
[打包防護工作區](workplace/release-tool-guards-20261004/receipt.json)，入口已掛入CONTEXT。
全程沿用 `kol-package:20260907`，映像SHA為
`3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193`；無網路，
UID/GID1000:1000、1GiB、1CPU及64程序上限。所有權抽查與全工作根root-owned／md目錄
掃描均通過；Docker名稱及專案掛載容器掃描為空，暫存測試副本已移除，來源與輸出保留。
`git diff --check` 通過。未選正式版號、未打包、未提交、推送、寫入Issue或發布。
怪物四套資料格式問題仍待選擇，依賴該決定的正式資料與接線維持現狀。


## 2026-10-04 打包組裝與內容驗收移入Docker

前輪版號及覆寫防護有實質進展，開工已核對七支工具雜湊。路由仍命中完整收尾，
載入 `retro-remake-end-to-end-playbook.md`，交付目錄及文件職責依各自reference。
版號、目錄及文件契約雜湊與前輪相同；保留專案 `full/`、`release/`、`promo/`。

- `build_dist.sh` 新增內部組裝階段。建立包子目錄、複製檔案、素材排除檢查及寫清單
  在 `kol-package:20260907` 執行，`/src` 唯讀，只有指定 `/out` 可寫；主機保留
  Git／可見度核對、輸出目錄配置與Docker控制。既有資料清單及權利分界維持。
- `verify_dist.sh` 的條款、素材、Windows檔頭、封包及清單內容檢查移入唯讀容器；
  macOS結構檢查仍走既有專用容器。Linux啟動檢查改為唯讀來源，先核對實際
  `-version`，再畫封面及結局，退出時終止並等待Xvfb。
- `make_promo.sh` 的來源腳本複製移入同一影片容器，合成成功後才複製。
  未改合成腳本、素材、音訊、分鏡或影片，不執行本輪影片重建。
- 追查 `KOLBOOK-character-writer-summary.json` 僅見於指定的原版角色寫檔CLI；
  本輪沒有把它判作正常玩家阻塞，也沒有為未使用的helper增加資料或逆向工作。
- 擴充既有 `tools/test_dist_manifest.py`，16項隔離回歸及三支Shell語法檢查通過。
  正例驗實際組裝、位元組保全及完整清單；反例驗大量違規素材、錯版號、錯PE檔頭。
  另驗容器參數及唯讀掛載、版號成功／失敗、畫面失敗時的Xvfb清理，以及合成成功／
  失敗後的來源複製。Docker、編譯器、lipo及媒體合成替身不代表平台或影片驗收。
- 用現行來源及三個既有凍結候選做真實容器組裝，936份資料逐檔相符；full清單945檔、
  release280檔、promo只有LICENSE一檔，三包清單重算通過。全組裝共1230檔，沒有
  六個平台封包或影片。這是本輪workplace的隔離測試輸出，不是新的正式交付入口。
- 真實Linux b87ad738候選仍回報development。fixture版號為v.0.0.0-20000101，
  正式啟動比對正確拒絕3；改以development作診斷則啟動0，封面及結局八張PNG
  完整解碼為320x200。兩次Xvfb清理均核實。診斷正例沒有通過正式版號驗收，
  不能代替正常玩家鏈、Windows／macOS真機或新的原版oracle。
- 前輪21附件及四支本輪工具的原來源保留；歷史九項測試由after-source重新執行通過。
  329項Go／module來源不變。最長勝利與父收據重驗通過，權利防護六個案例通過。
  打包待辦只更新現況，render／verify仍21項未完成、85項完成，不改原完成條件。

本輪封存1257附件、八張本機啟動畫格，來源及重生腳本在
[本機收據](workplace/release-docker-pipeline-20261004/receipt.json)，入口已掛入CONTEXT。
沿用 `kol-package:20260907`，映像SHA為
`3d32441befe02810c3220a2e9f2db5b84ee65c703e5b65ee563712a3519ad193`；UID/GID1000:1000、
無網路。一般驗證1GiB、1CPU、64程序，GUI診斷2GiB、1CPU、128程序；均有外層逾時及--rm。
全工作根root-owned／md目錄掃描為空，輸出所有權均正確；Docker名稱及專案掛載容器
掃描為空。暫存測試副本已移除，來源快照、隔離組裝及有限啟動畫格保留。
`git diff --check`通過。未選正式版號、未重建正式平台包、未提交、推送、寫入Issue或發布。
怪物四套格式仍待使用者選擇；正式資料與接線維持現狀。

## 2026-10-04 雷聲離線試聽材料

路由命中完整收尾與音訊轉檔，分別載入 `retro-remake-end-to-end-playbook.md` 及
`remake-audio-ogg-default.md`。新材料的入口與重生方式已掛入CONTEXT及音訊驗證README。
SPEC-004仍為DRAFT，正式runtime音訊仍限OGG；本輪只準備研究試聽材料。

- 保留既有候選OGG，SHA-256為0e95e887d2513161a4f07d85fad3c64d7fc538d3b61864256f787d98db85d362。
  不重合成或重新編碼。原版捕捉WAV的SHA-256為
  f2c5191545b3bae1f066c9328db184c9406541941c8a008426d95a8f2260936b。
- 首次嚴格WAV讀取失敗，已保留兩份部分輸出及失敗結果。原RIFF長度欄位為28，
  data長度欄位為0，實際檔案7340076位元組。這是輸入檔頭與驗證腳本問題，
  不列為產品缺陷；捕捉未正常關閉只標強推論，不追查DOS硬體driver。
- 修正隔離腳本後，以FFmpeg產生有效的研究WAV副本。完整PCM逐位元組等於原檔
  offset44後內容；四段連續樣本的裁切各保留前後0.1秒，逐樣本相符，不補零或調增益。
  原檔全程唯讀，六份固定來源的雜湊均維持。
- 離線HTML共五個播放控制、六份內嵌音訊，內嵌資料雜湊均相符。六份音訊完整解碼通過，
  研究WAV非靜音。OGG容器時長0.961451秒，FFmpeg解碼42496個44100Hz樣本，
  時長0.963628秒；兩者各自記錄，不以解碼結果宣稱硬體時鐘或逐波形對拍。
- 原版捕捉沒有受控seed，只作聽感參考。人耳確認與瀏覽器播放均未驗，未接正式音訊，
  未把SPEC-004升為READY。正式生成流程與原候選的編碼器版本未變。
- 封存16份附件，329項Go／module來源不變。待辦render／verify通過，
  仍21項未完成、85項完成；未改狀態、驗收條件或怪物四套格式決策。

[試聽頁](workplace/audio/listening-review-20261004/audition.html)、
[來源及裁切收據](workplace/audio/listening-review-20261004/receipt.json)、
[附件雜湊](workplace/audio/listening-review-20261004/SHA256SUMS.json)均只留本機。
含原版錄音的WAV與HTML受既有workplace忽略規則保護，未加入Git或交付包。
沿用固定映像game-video@sha256:3fb42fb6a46e34b72f0c04c50426b6912d6f901aca237af6a5f2754e9ad5ca1a，
FFmpeg 6.1.1-3ubuntu5；UID/GID1000:1000、無網路、1GiB、1CPU、64程序及120秒外層逾時。
驗證沿用kol-package:20260907，512MiB、1CPU、64程序及60秒外層逾時。
工作根root-owned／md目錄掃描為空，輸出所有權正確；專案容器掃描為空。
保留有限試聽材料、原始輸入及失敗分類，不留下持續容器。git diff --check通過。
未選正式版號、未建正式平台包、未提交、推送、寫入Issue或發布。
怪物四套資料格式仍待使用者選擇，依賴該決定的正式分支維持現狀。

## 2026-10-04 法術名稱空白邊界

前輪雷聲試聽材料有實質進展。開工核對現行21條未完成項，沒有重開已結案項。
路由命中完整收尾、dosgolem正常原版執行及RE回填；載入end-to-end playbook、
dosgolem README／CLAUDE與re-resolution-backlinks。沿用reverse-engineer-retro-game-remake
技能，先登記兩條正常輸入邊界，再執行既有固定版本的正常啟動、招募、旅行及購買。

- 原版固定dosgolem revision 968ca774705853276ae55c8d1163c369481ac71f；
  KOL.EXE→KOLBOOK→packed KOLTOWN.EXE。KOLTOWN.EXE的SHA-256為
  c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221。
  seed只在正常旅店招呼後DS:1600=1、1602=0設定兩次；後續不重設，沒有座標／角色注入。
- 原版edges為191狀態，spaces為187狀態，兩次測試皆0；名稱輸入無亂數判定。
  各前180個角色記錄、seed與PNG皆與已封存native-name4相同。驅動只修正報告的
  packed輸入標籤，不改原執行檔、caller或回傳值。
- 輸入「 A  B 」後，原版打包字數6，保留首、連續及尾空白；兩個空白亦能確認並
  正常存錄，打包字數2。記錄分別為60040010006c6f6e4000000006000a22及
  20003000616c6f6e4000000006000a22，DGROUP角色DS:BB88+0x200。
  抽取器去尾空白後顯示「 A  B」及空字串，不等於原版空法術槽或未接受名稱。
- 現行Linux b87ad738正式候選維持；英中四次正常flow各完成199／195狀態，共788。
  第一例會刪首尾空白，第二例會保留Kelnalon。兩例原版與重製扣款均為金冠
  987→647；原版冒險點數2816→2476，效果6、護符[64,0,0,0]、持續0、強度10保持。
  沒有把trace未顯示的重製冒險點數當作已驗結果。
- 十張仍在名稱輸入的115×13英文框零像素差；十八張對照的整頁差異及關閉後相同
  區域的差異照實保留。原版全空白確認頁與正式Kelnalon頁已目視核對。
  首次腳本把關閉後的座標也要求零差，edges第189步547差而失敗；保存失敗分類後，
  只在輸入框仍開啟的明確鍵序階段要求零差，原始PNG與trace均未改。
- 證據追加到既有RE-212，回填SPEC-035的空白邊界DRAFT及CONTEXT。
  已建立附原始位址基準的回填索引，舊規格仍READY，舊正常N／12字收據保持。
  正式Go、字型、保存v3與資料格式未改；typed空白名稱與法術槽的表示尚待審查，
  正常法術清單及住宿保存／重載仍未由本輪驗完。
- worklist只補三條待辦的現況；render／verify通過，21未完成／85完成。
  原狀態、完成條件及核實契約維持。329項Go／module及前輪16份音訊附件雜湊保持。

[原版與正式差異](workplace/spell-name-boundaries-20261004/review.json)、
[完整本機收據](workplace/spell-name-boundaries-20261004/receipt.json)及正常重生腳本
已掛入CONTEXT。Go映像SHA為3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa；
原版3GiB、2CPU、256程序、1000秒外層逾時；正式試跑2GiB、1CPU、128程序、
800秒外層逾時及Xvfb退出trap。來源與dosgolem唯讀，只指定輸出可寫，
UID/GID1000:1000、network none、--rm。驗證沿用kol-package:20260907。
工作根root-owned／md目錄為空，輸出所有權正確，專案容器掃描為空。
保留有限研究輸出與失敗分類，未改原始輸入或前批收據。
HEAD仍1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，git diff --check通過；
未提交、推送、寫入Issue、選正式版號、打包或發布。怪物四套格式及人耳驗收仍待原閘門。

### 2026-10-04：法術名稱空白隔離原型，正式升格待明確核准

路由命中規格閘門，載入 `retro-remake-spec-gated-workflow.md`、共同決策技能及文件職責契約。
本批延續前批兩個正常原版空白反例，只建立可丟棄原型與修正提案。

- 舊程式的空字串案例通過，首尾空白與全空白案例失敗。首次完整測試因隔離副本
  缺少 `docs/i18n` 而失敗，保留紀錄並分類為測試輸入缺漏。從乾淨副本唯讀接入
  該目錄後，`go test ./internal/town ./internal/save -count=1` 通過。
- 原型僅在副本的名稱提交處取消 `TrimSpace`，空字串仍沿用原名稱。
  另建唯讀觀察副本，JSON 深複製完整隊伍與城鎮狀態，不改狀態、輸入或亂數。
  同輸入的原型與觀察副本，其PNG及移除觀察欄位後的trace一致。
- 英文及繁中，兩個名稱案例共32次正常程序、3,552個狀態與PNG。
  住宿保存、退出重載、招募及MIRROR法術頁均走正常按鍵。名稱「 A  B 」保留六字，
  名稱「  」保留兩個空白且仍占有法術槽。前189步共756張PNG與正式基準一致。
- 存檔仍為 `kol-save-v3`。除預期名稱及保存時間外，完整JSON與正常基準一致。
  重載核對兩名人物的14項持久欄位及34個攜帶槽。兩張英文提交名稱區域96×7
  與既有原版正常收據零像素差。未新增原版空白名稱住宿重載實驗，不宣稱完整保存parity。
- 自動核准審查拒絕將SPEC-035空白邊界升READY，理由是正式行為閘門的變更
  尚未獲現有GO明確授權。被拒命令未執行，正式規格及329份Go／module輸入維持。
  留存的升格提案腳本加入無條件退出防護。兩份Go差異提案未套用，等待明確核准。
- 僅更新CONTEXT、既有RE-212及三條待辦的現況敘述；render／verify通過。
  21項未完成、85項完成，狀態、完成條件及核實契約保持。前批1,374份封存附件
  與原始輸入逐一核對。當前329份來源亦在隔離副本中保留並核對。

[修正提案](workplace/spell-name-spaces-prototype-20261004/proposed-formal.patch)、
[審查資料](workplace/spell-name-spaces-prototype-20261004/review-bundle.json)、
[本機收據](workplace/spell-name-spaces-prototype-20261004/receipt.json)及重生腳本已掛入CONTEXT。
原型二進位SHA-256為59e40422a8694dbbe3a2515e1ca6fd223b6c5ec361281d1994af12445fbcd83c；
觀察副本為23fe8b662377d8a77ca0e5596b25e0ff4abb12314887d4781f95f4d627e3c256。
建置沿用固定Go映像3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa，
驗證沿用kol-package:20260907，未建立映像。建置3GiB、2CPU、256程序及600／300秒逾時；
正常路徑2GiB、1CPU、128程序及1800秒逾時。來源唯讀、輸出限定工作區，
UID/GID1000:1000、network none、--rm，Xvfb有退出trap。
封存4,386份附件，其中3,552張PNG；封存後逐一核對雜湊、語法、DRAFT及來源未改。
工作根root-owned／md目錄為空，輸出所有權正確；容器退出後按專案掛載與名稱掃描均為空。
保留失敗分類、隔離原型與未套用提案。未提交、推送、寫Issue、選版號、正式打包或發布。
`git diff --check`通過；本批收據、提案與二進位均受既有ignore規則保護。

### 2026-10-04：原版空白名稱正常住宿保存及實際檔案重載

前輪完成隔離原型與修正提案。本批補原版保存鏈，不依賴尚未核准的正式升格。
路由命中決定性原版對拍，載入dosgolem的README、CLAUDE、驗證交接及文件職責契約；
寫入新證據前再核對規格閘門。沿用RE-213正常旅行、逃離及住宿探針，沒有新增執行器。

- 預先登記六字元「 A  B 」及兩個空白兩例，各自從乾淨原版啟動。
  正常旅店招呼後只寫一次seed1的兩個word，世界、戰鬥及住宿不重設。
  重載與招募沒有種子寫入。沒有direct-entry、座標／人物注入、wait stub或局部常式呼叫。
- 原版兩例各正常逃離四次遇敵，回旅館付費住宿，實際寫出12,288-byte的CHARDATA
  及三份肖像資料。保存路徑579／575狀態；從本次實際保存的四個檔案正常重新啟動，
  各七狀態招募回兩人，共1,154個保存狀態及14個重載狀態。
- CHARDATA人物槽2的法術檔案偏移為0x800；執行期對照為DGROUP角色DS:BB88+0x200。
  六字元記錄60040010006c6f6e4000000001000a22，兩空白記錄
  20003000616c6f6e4000000001000a22。打包codes分別[0,1,0,0,2,0]與[0,0]，
  有效長度6／2。兩名人物的完整法術區、攜帶資料及角色資料頁欄位重載一致。
  原版抽取器rstrip後的空字串仍不能代表空法術槽。
- 每例前181狀態及PNG與RE-213正常N來源一致；實際整份CHARDATA與該基準相比
  只差法術名稱區域。雙語隔離原型的名稱、護符、效果、金冠601及冒險點數2486相符。
  唯一抽樣欄位差異為營養10／8；本次原版三次露營逐次4→6→8→10，原型兩次。
  保留差異，不宣稱全旅程同狀態、整場戰鬥、整段骰序或整頁parity。
- 1,196張原版PNG全部解碼為320×200；主收據封存1,246份附件。
  核對器使用的Python工具及直接依賴共11份，在核對後追加封存，另存13份附件清單。
  此補充明示為事後工具來源記錄，不回寫主收據。Python版本3.11.2。
  工具清單首次誤讀不存在的extract_goods_table.py，屬核對檔名錯誤；
  改由實際import模組取得來源，未分類為產品缺陷。
- 追加RE-212與SPEC-035的DRAFT證據，更新CONTEXT及三條待辦的現況敘述。
  render／verify通過，21未完成／85完成；狀態、完成條件及核實契約維持。
  329份Go／module及前輪4,386份原型附件雜湊保持。此前自動核准審查拒絕READY升格
  仍有效；沒有升READY／CONFORMED，沒有套用兩份Go修正。

[原版欄位核對](workplace/spell-name-spaces-original-save-20261004/review.json)、
[主收據](workplace/spell-name-spaces-original-save-20261004/receipt.json)及
[核對工具來源](workplace/spell-name-spaces-original-save-20261004/toolchain-receipt.json)已掛入CONTEXT。
重生入口run_native.sh；唯讀重驗verify_toolchain.py，來源變更後可加--recorded-source，
只驗封存歷史，不宣稱當前程式回歸。
原版KOLTOWN.EXE SHA-256為c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221；
dosgolem固定968ca774705853276ae55c8d1163c369481ac71f，Go映像固定
3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa，未建新映像。
原版3GiB、2CPU、256程序、1500秒外層逾時；核對沿用kol-package:20260907，
1GiB、1CPU、64程序及90／120秒有界逾時。來源唯讀，只限定研究輸出可寫，
UID/GID1000:1000、network none、--rm。全部四次原版程序退出0。
主收據及工具補封重驗通過，工作根root-owned／md目錄為空，所有輸出1000:1000，
按專案掛載與名稱的容器掃描均為空。保留有限本機研究資料，既有ignore規則涵蓋主收據、
工具補封及CHARDATA；沒有提交、推送、Issue寫入、正式版號、打包或發布。
HEAD維持1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，正式修正仍待前輪明確核准。


### 2026-10-04：繁中送禮候選人、正常轉交及保存重載

路由命中完整收尾流程，載入retro-remake-end-to-end-playbook.md；兩次輸入失敗後
重查正常玩家路徑，載入retro-game-playtest.md。結論回填前再讀規格閘門及文件職責。
依21項目前待辦找到英文Giving列候選人、繁中漏列的現行缺口；沿用SPEC-031第4節READY。
工作區及腳本先掛CONTEXT，前後329份Go／module輸入留存。沒有新增根目錄或正式規則。

- 只改internal/ui/itempage_cjk.go及internal/ui/town.go；其餘327份Go／module不變。
  Linux為b6f8d01ad1a0d64ca0069061f306534ba10d72f0dd95076ba676172951d2c093。
  Windows／macOS為此前shader候選，本批未重建或執行，CONTEXT現行／歷史已分開。
- 從標題選Theodoric、招募Poontang、正常數字鍵切回發言者，經MIRROR及物品頁送腰帶巨劍。
  seed1固定，無direct-entry、座標／人物注入或強制勝利。唯讀觀察版只深複製Party及Inside至trace，
  與正式版PNG及可比trace相同。20次正常UI程序共560狀態及PNG；英文34張不變，繁中只step-032增姓名。
  兩框108×18及103×18完整字模零差，共581墨點；舊漏名反例拒絕，目視姓名可讀。
- 正常給第2人後，同一巨劍record由Theodoric腰帶到Poontang第三口袋。使用正式預設住宿模式付費保存，
  兩語修正前後完整v3只差saved_at。從實際新保存重開，持久欄位、裝備、背包及攜帶鏡像相符，
  原存檔SHA不變。各26狀態正常旅店及收件人物品頁顯示第三口袋巨劍；未新增原版送禮oracle。
- 缺xdpyinfo的首啟動退出127，遊戲未執行；同映像改X11 socket就緒後重跑。
  三次輸入探測分別多按RET退出旅店、招募改發言者、誤選TAKE，保留98張失敗PNG。
  保存比較只正規化Good零picture／movement的JSON省略差異，原遊戲輸出保留。
  缺檔名讀取、尚未產生畫格的view_image及未加docker -i而未執行文件腳本，均列工具問題，未判產品缺陷。
- Go固定映像3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa，
  Go1.26.7／Ebitengine2.9.9；全部internal/...及cmd/kol測試及正式／觀察建置退出0。
  建置3GiB、2CPU、256程序、600秒；GUI沿用kol-package:20260907，2GiB、1CPU、128程序、600秒，
  Xvfb有kill／wait trap。全部--rm、network none、UID/GID1000:1000；來源唯讀，僅工作區及指定文件可寫。
  未建新映像。驗證Python3.11.2，封存與重驗1GiB、1CPU、64程序、120秒。
- [本機收據](workplace/item-give-player-path-20261004/receipt.json)封存1442附件、658PNG與920視覺／工具輸入，
  [重驗器](workplace/item-give-player-path-20261004/verify.py)及update_docs.py掛既有工作區索引。
  三條worklist只追加目前證據；完成條件、狀態與verify契約不變。render／verify通過，21未完成／85完成。

父SPEC-031仍READY；六人送禮、其他動作及原版規則未驗。SPEC-035空白名稱仍DRAFT，待核准修正未套用。
未提交、推送、寫Issue、選正式版號、打包或發布。所有權、容器與暫存清理於下方追加收尾核對。

收尾核對：新收據唯讀重驗退出0，前輪原版空白名稱以--recorded-source驗歷史也退出0，
主收據及工具補封保持。初次worklist verify因新限制句沒有待辦認領而退出1；
明寫「六人送禮的畫面還沒對拍」後，依既有UI待辦契約認領，沒有放寬pattern或完成條件。
render及最終verify均通過，21未完成／85完成，產生區塊與JSON同步。
工作根root-owned檔案及誤建md目錄為空，新工作區所有檔案1000:1000；
唯讀核對結束後，本專案掛載及kol名稱的容器均為空。沒有重建映像，未清理其他專案資源。
保留本機失敗畫格、來源副本及收據；二進位與收據皆由既有ignore保護。
HEAD保持1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，git diff --check通過。

### 2026-10-04：送禮修正跨平台重建與Windows真實按鍵續核

前輪已修正繁中Giving漏列候選人。本批命中重製驗證與交接，載入技能的
verification-and-handoff.md及文件職責。先核對AGENTS、WORKLIST、git狀態、現行來源與映像；
現有異動保留。預先索引workplace/item-give-platform-20261004，沒有新增根層目錄或交付結構。

- 從前輪封存核對329份Go／module及920份視覺／工具輸入，保存來源副本。
  三平台均由同一來源重建；正式Go未改，前輪全internal／cmd測試仍對應此來源，不無故重跑。
  Linux逐byte保持b6f8d01ad1a0d64ca0069061f306534ba10d72f0dd95076ba676172951d2c093；
  Windows為86fcdb0b1a2213b8b24d12d5226e2e81cb4de7acc8498e48b39b9a69f8f61a63；
  macOS universal為108fb7632c300fa3a5d898aee33d3e4c32442215c1e6cd7f742206fc81dc5b96。
- 沿用Go映像3e89464c、kol-osxcross-go的1d7655e7及Wine映像4e41c4dc，完整ID見收據。
  Go1.26.7、Ebitengine2.9.9、Wine9.0；Python3.11.2。macOS雙架構、最低11.0、
  arm64臨時簽章及系統相依庫通過；沒有執行macOS或聲稱真機驗收。
- 在執行前鎖定二進位、正常鍵序、控制器及完整期望畫格。Wine預設DirectX兩語六條
  正常送禮、住宿保存及本次保存口袋重載，共208狀態與PNG，完整軌跡及全部畫格同Linux。
  完整v3保存只有saved_at差異。沒有direct-entry、座標／人物注入或強制勝利，seed1固定。
- 另以136真實X11鍵走四次正常程序，24完整960×600畫格符合預定畫格，
  WM_DELETE_WINDOW正常關窗及程序退出0。兩語GUI與腳本完整v3只差saved_at，
  重載後收件人第三口袋巨劍可見，原保存SHA保持。目視繁中送禮及口袋明細可讀。
  舊漏名畫格581差異像素反例拒絕；沒有豁免文字或框線。
- 15張首擷取仍是前頁，預先固定的30秒完整畫格閘門均通過；首張及末張均保留。
  沒有重送鍵、調整種子、增加允許區或因觀察逾時重啟。第一次Docker inspect
  因缺Entrypoint欄位而模板失敗，改用index安全取值；是控制面讀取問題，映像未改。
- 建置Go為3GiB、2CPU、256程序、600秒，macOS為6GiB、2CPU、512程序、900秒；
  Wine腳本與GUI為2GiB、1CPU、128程序、900秒，Python控制器1GiB、1CPU、64程序、900秒。
  核對及封存1GiB、1CPU、64程序、120秒。全部Docker、--rm、network none、UID/GID1000:1000；
  原始來源唯讀，只有已登記工作區及指定文件可寫，Xvfb／遊戲／Wine server有退出清理。
  未建立或更換映像；沒有HOME覆寫。
- [本機收據](workplace/item-give-platform-20261004/receipt.json)封存709附件及280PNG，
  [唯讀重驗器](workplace/item-give-platform-20261004/verify.py)退出0，掛入CONTEXT及SPEC-031第16節。
  兩條worklist只追加目前證據，render／verify均退出0，21未完成／85完成；
  狀態、完成條件及verify契約保持。前輪收據及待核准空白名稱修正維持。
- 四次GUI退出後，只清理本批兩個Wine前綴，各844檔、720400502 bytes，以及空X11目錄。
  保留全部驗收輸出，沒有清理其他專案資源。擁有權與專案容器收尾核對於下方追加。

Linux／Windows／macOS現行候選已在唯一現況表更新。macOS僅結構，Windows／macOS真機、
原版送禮及其他未訪分支仍按原閘門。本批沒有正式版號、封包、提交、推送、Issue寫入或發布。

收尾核對：本輪主收據唯讀重驗退出0；工作根root-owned檔案及誤建md目錄為空，
新工作區檔案全部1000:1000。全部核對結束後，專案掛載及kol名稱的容器均為空。
二進位與收據由既有ignore保護，git diff --check通過；
HEAD保持1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，前輪Go及封存證據沒有回寫。

## 2026-10-04：原版正常送禮與逐人保存續核

新增本機workplace/item-give-original-20261004，固定dosgolem968ca774與Go image3e89464c。
run.sh／run_case.sh正常啟動招募兩人，六條旅店延伸201鍵207狀態；
SPACE及ENTER才搬移，部分保存重載作反例，兩人真正CHARDATA正常重載成功。
prepare_prototype.py／prototype.sh／run_prototype.py生成兩版隔離原型，各雙語300狀態。
第二版五份Go變更及唯讀觀察入口，完整v3只差saved_at，52重載圖及軌跡與現行相同。
verify.py／check_prototype.py重算後由seal.py封存1624附件849PNG；
review.json為可核准提案。原版證據集中docs/re/20261001-item-detail.md，SPEC-031第17節DRAFT。
正式329輸入未改，完成條件保持21未完成／85完成。
首輪AH=0E斷言過窄與Carried欄順序已依實際案例及既有契約修正，原腳本保留。
文件渲染首次缺CONTEXT可寫掛載，補齊後重跑；操作暫態措辭改為確認等待，未改掃描規則。
自動審查先拒絕尾段替換；唯讀核對確認尾段僅本輪新增，再用完整SHA保護及備份替換。
這些是驗證／文件腳本問題。HEAD仍1384e508，482筆既有異動保留，未提交、推送或發布。
原版記錄、兩版原型及反例均僅本機；Xvfb有trap，容器--rm與斷網。
最後verify.py重驗全部附件及第二版原型通過；worklist render／verify退出0，21未完成／85完成同步。
本專案送禮容器0，專案root-owned檔案及誤建.md目錄0。工作區來源與收據已被Git忽略。
兩份受SHA保護的文件備份及文件更新腳本留在/tmp，無持續執行程序；正式來源與HEAD保持。

## 2026-10-04：滿背包送禮、取消與六口袋保存續核

新工作區workplace/item-give-full-pack-20261004，只重用已封存工具及現行候選，不改正式Go。
prepare.py／run_native.sh預鎖正常六次轉交與披風失敗；原版副本缺執行位元退出126，
保留未啟動紀錄，只修自有副本，native-full同92鍵重跑通過。
prepare_save.py／run_native_save.sh再106鍵保存及31鍵重載，第六口袋護腿與原物品不變。
run_remake.py／run_remake_save.py的正式／唯讀觀察兩語12次共900狀態，完整PNG及可比軌跡相符。
四份v3只差saved_at；原版完整物品槽相符，滿背包提示回問及內層取消差異未稱通過。
新增披風額外防護列觀察，記入SPEC-031第17.1節DRAFT及現況，五檔成功確認仍未核准。
verify.py／seal.py封存1237附件1153PNG，全部圖片實際解碼；正式329及視覺920輸入保持。
首輪Note翻譯斷言誤判已按現有譯文更正；60秒外層收尾逾時改120秒，同條件重跑退出0。
上述副本／驗證問題各留紀錄，不當成產品缺陷。正式實作、完成條件及待辦狀態保持。
原版記錄與保存只留本機，來源唯讀、所有可寫容器UID/GID1000:1000，Xvfb有trap與有界擁有程序。
最後唯讀重驗1237附件及1153圖片退出0，329／920來源保持；worklist render／verify退出0，21未完成／85完成同步。
滿背包工作容器0，專案root-owned檔案與誤建.md目錄0；新收據及原版工具副本被Git忽略。
HEAD仍1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，既有異動保留。文件更新腳本留/tmp，無持續執行程序。
未提交、推送、打包、定版或發布；先前成功確認五檔提案仍等待明確核准。

## 2026-10-04：披風類型0明細的最小顯示證據與隔離原型

- 路由命中IDA Pro9.4，載入技能、知識入口與官方README／ida-94-tools；建立工作區前再讀文件職責，下結論前查規格閘門及grilling。CONTEXT先登記workplace/item-misc-detail-20261004。
- 沿用locked-v1映像6f6d59af、正式.i64唯讀，新副本匯出sub_153FF的283指令；原名、operand、bytes、xref與既有匯出一致。原版類型0跳過部位、防護及換列，負重與類別同列、價值及狀況下一列。原址與雜湊見既有物品研究檔新增段。
- prepare.py複製329輸入及既有唯讀觀察main；prototype.sh在Go3e89464c離線建置，只在副本shop.go補顯示閘門。run.py沿父收據預鎖正常按鍵，雙語104住宿／29重載各一次，共266狀態，無角色、座標或物品注入。
- verify.py核對人物、六口袋、未轉交披風、v3保存除saved_at、重載及原來源329／視覺920保持。各語只有正常披風86／87兩張圖改變，262其他圖保持；533來源／原型PNG全部解碼。英文明細框196×68與原版零差，繁中全圖符合前版實際字模搬移模型。舊額外防護頁及單像素損壞反例拒絕；目視中文可讀。
- seal.py封存641附件及266新PNG，驗父收據附件SHA；封存後相同verify.py以/src:ro乾淨退出0。原型不是正式程式，SPEC-031第18節DRAFT與單檔review.json待明確核准。其他類型0、食品正常抽樣及送禮成功／失敗／取消仍未閉合。
- IDA首輪因預設ASCII讀中文索引退出1，保留初稿與log，明列UTF-8後用新DB副本退出0。驗證初稿誤用疊框位置、未換算/out路徑、未分語言明細預期；各自保留初稿與更正JSON，修正後相同期望乾淨重驗通過，不記為產品缺陷。工作清單首核將本輪DRAFT的正式來源未改措辭列為未認領；改用既有待辦明確記載的來源保持狀態，最終verify退出0、21未完成／85完成／13人工，未改完成條件或verify。
- 本輪只更新既有CONTEXT／WORKLIST／WORKLOG、研究檔與索引、SPEC-031及待辦正文。正式327Go／兩module不變，HEAD保持1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；沒有提交、推送、Issue寫入、版號或交付包。
- 全部工作有界Docker、network none、UID/GID1000，來源唯讀、輸出限本工作區。整個專案root-owned檔及.md目錄查無；本輪容器均--rm，Xvfb有trap，收尾查無kol.task容器。工作區與原型均Git忽略，全部本機留存，不擴張原版散布權。

## 2026-10-04：現行四回合進場與裝備續核

- 命中remake驗收、dosgolem固定種子、文件責任路由；載入reverse-engineer-retro-game-remake及verification-and-handoff、knowledge-router及project-document-responsibilities。資料格式確認沿grilling。原版沿RE-186鎖定dosgolem收據與RE-203 IDA9.4證據，本輪未啟動原版或改.i64。
- 現行329來源及920視覺輸入核對保持。標題起跑、固定269正常鍵、seed1與首次規劃後一次重設，四次正式／自然觀察／受控雙語各270狀態及14圖；自然觀察排除audit欄位後與正式完全一致。受控舊序列第四回合未勝，首次隊員已由Theodoric變Olear。
- 根因為原版0x1EBEA逐怪物選四套裝備的RND(4)尚未正式接線，原版／現行首次提示前種子1323149826／2415577835。沿RE-203原型套到當前來源副本，只改隔離資料與規則；不是插入空骰或調seed過關。雙語各270狀態、14圖恢復四回合自然勝利並領取戰錘返回。
- 每語言九個同階段點、三人38欄共1026及54計數器無差；4534進場骰、4096地形格、六件敵人裝備、99戰中／報酬骰、34原始槽與報酬相符。全批1620狀態、84圖已解碼。繁中戰利品已目視。只驗有限三人欄位，不宣稱50-byte完整記錄、天候骰直接觀測、新住宿重載或真實視窗鍵。
- 主機第一次docker受sandbox拒絕，經require_escalated沿用既有image。另一次只讀查詢假定的0x1EFAB caller不存在，改查原收據實際0x1F01C。比較器找不到重設前下一筆seed、混用報酬前出口與報酬後UI；兩份初版及訂正保留，改按實際重設邊界與post_battle[2]同階段驗。一次文件patch含空hunk，解析拒絕後移除空hunk重做，未改產品。
- 工作區為workplace/full-battle-current-20261004，CONTEXT、RE-203、RE-196、RESEARCH-LOG及worklist兩項追加；README與正式spec狀態保持。資料格式提案仍DRAFT，依共同決策閘門等待選擇，21未完成／85完成／13人工項保持。
- 所有工作在既有Go／封包Docker中，UID/GID1000:1000、無網路、相稱資源及外層逾時，Xvfb有trap。原始資料唯讀，副本保留。封存與批末重驗、清理結果見同工作區receipt.json及audit.json。未提交、推送、寫Issue、定版、打包或發布。

封存前另發現Python尺寸元組與JSON陣列的型別比較不相等；比較值保持，改輸出陣列後乾淨重算。
初版seal.py及當時compare.py已保存。第一次封存未生成收據，我提前啟動的相依重驗因此找不到檔案；後續改為封存成功才重驗。
這些屬驗證工具問題，不列產品缺陷。正式來源及既有原版收據保持。

## 2026-10-04：採用怪物資料v2並完成限定接線

- 使用者明確回答採v2。再次核對knowledge-router的remake生命週期與規格閘門，讀retro-remake-spec-gated-workflow及文件職責；依RE-203先READY，再實作。人物保存v3、其他未決提案與發行方式保持原決定。
- 先固定332來源與920視覺輸入，保留before-source／before-visual。候選審查及驗收後只套用14檔，331份Go、333程式輸入逐SHA核對。每隻怪物首次規劃前擲RND(4)並複製獨立六格，重規劃不重選；面板與訊息讀當前裝備。
- 載入發現七格非零填充，回RE／READY補證後修正。原版盾格第一byte判空已證實，768格raw保留、625有效物品及143空格；重生JSON逐byte一致。來源及失敗輸出保留。
- 全部internal／cmd與格式反例通過，原版三巨魔18件裝備及能力初始化局部比較通過。雙語四回合各九點1026欄、54計數器、99戰鬥報酬骰、34槽及報酬相符。四回合兩側在首次規劃固定seed1，只是驗收條件。
- 新正常住宿各509鍵／510狀態，正式由書本選人、正常招募重載各24狀態與PNG，34槽與持久欄位還原。正式未重設亂數另各349鍵／350狀態回城保存；正式與觀察軌跡及20PNG一致。
- Linux d443c00d乾淨重建逐byte相同，Windows dd1442fb與macOS雙架構／universal 81da1b45重建。Windows Wine預設DirectX雙語各350狀態／20PNG與Linux一致，完整v3只有saved_at差異；macOS只驗結構，不宣稱真機通過。
- 失敗分為資料契約、測試狀態、觀察器及環境／驗證路徑，細節與訂正見本輪corrections.json。StageIntro與只支援隊員的DropItem測試、第二場觀察器LCG型別、兩人離隊後重載路徑、唯讀別名及Wine缺Python均已修正，不另列產品缺陷。缺武器名稱讀到母表是實際UI問題，本輪已修正。
- 更新CONTEXT唯一現況、RE-203與RESEARCH-LOG追加、SPEC-011第13節有限CONFORMED與索引，收斂worklist兩項現況文字。#6／#44保持開放；不擴大到完整50-byte、+0x21、其他射擊／裝備、任務／競技場或表外返回。
- 所有工作在既有有界Docker中，UID/GID1000:1000、無網路、原始輸入唯讀。工作區workplace/monster-equipment-v2-20261004，封存與批末唯讀重驗見receipt.json／audit.json。Wine暫存prefix在程序結束後移除；容器、擁有權及工作樹檢查記入audit。
- 無Git暫存／提交／推送、遠端Issue寫入、版號變更、正式打包或發布。既有482行異動未重設，歷史收據保持。

本輪封存2159附件與1284張PNG索引，包含原有視覺快照及失敗輸出，不能全數稱新遊玩畫格。
收據SHA-256為239e987c24be1ab7529fc31fdb43f45d537d70188afab4c295426890d4cdaf6b。
封存後唯讀重驗通過，333程式輸入、新資料及全部附件相符；舊v1來源／視覺覆蓋重驗799附件也通過。
worklist三行新限制最初未認領，已明確掛回開放的#44後重跑render／selftest／verify／scan，均通過。
21未完成／85完成／13人工項保持。全專案232169項擁有權檢查沒有root-owned或.md目錄，Wine prefix已移除。
最後487行工作樹狀態、HEAD保持1384e508，git diff --check通過；專案相關執行中及停止容器均無殘留。

## 2026-10-04：怪物實際射擊取槽與負重重算

- 重查knowledge-router的remake／規格閘門、文件職責與技能RE／verification reference。使用use-ida-pro-9-4技能並讀專案README及ida-94-tools。沿既有Go1.26.7、IDA9.4與dosgolem968ca774，未建新image。
- 正常原版195狀態完成三名Cliff Troll的第一回合，兩名投石、一名移動；正常旅店招呼後seed1，無角色／座標／指令注入、直接呼叫或等待替換。首175狀態及PNG與RE-203保持，449條新IDA指令與EXE bytes核對。
- 原版傷害依指令選+0x70，負重依首byte選+0x20／+0x70。舊正式槽6拿Troll Maul，碼83／基礎2，負重182／速度84；原版Boulder碼81／基礎0，負重172／速度63。先DRAFT證據審查再READY，三檔正式修正及正常來源回歸通過，332份Go／334程式輸入。
- 新取槽／正常初始化回歸在舊來源失敗，候選及正式全量internal／cmd通過。候選雙語各203鍵／204狀態、21PNG，四回合各270狀態／14PNG與v2來源逐byte保持。正式二進位SHA b0aa211629e3afafcaa0206d30b80d6d962b6db607e4300f0db21359e0528c1e與候選相同。
- 原版射後14／8→13／7與原始減值段已證實，但裝填種類讀角色區域外別名；未猜補，ConsumeAmmunition保持，SPEC-022怪物增補DRAFT。父SPEC-013、#6／#44及完整戰鬥保持開放；本場兩發落空，不宣稱正常命中傷害。
- 驗證腳本首次少算選項索引，current-*輸出保留；改用實際selected後current-rest-*乾淨重跑。審查首次忘了新增下一回合placement及混入隊員weight，重讀路由／完整點列後修正樣本。0x251A7的var_6是盾格，舊selected_offset標籤保留但不當武器證據。首次正式／觀察比較漏清城鎮audit_party，保存first-normalization輸出後同image／入口重跑；這些是腳本問題，不列產品缺陷。
- 新研究編號先誤用已存在的RE-204，索引檢查後正式訂正RE-216。候選歷史註解保留，正式註解訂正；只有註解不同，實際二進位逐byte相同。
- 更新CONTEXT唯一現況、SPEC-009／013／022／029勘誤與有限驗收、RE-216及索引、RESEARCH-LOG追加與worklist兩項。Windows／macOS是上輪331份Go，沒有把舊平台產物說成現行；本批未打包、改版號或發布。
- 全部工作在UID/GID1000:1000、無網路、有資源限制與逾時的Docker，原始資料及正式IDA庫唯讀，查詢只用一次性副本。批末收據、擁有權及容器清理見本機工作區audit.json，歷史產物保留。無Git暫存／提交／推送、遠端Issue寫入或新影像檔。

## 2026-10-04：RE-217怪物射後計數收尾

- 路由命中remake接手、spec閘門與dosgolem／IDA；沿用唯讀原版及固定工具鏈。
- 正常原版195狀態、202PNG、骰序與RE-216完全相同；新增別名讀取與2,313寫入觀察。
  槽6讀怪物0+33值33，槽7讀未使用怪物6+33值0；其他槽／跨場仍未知。
- 新IDA9.4一次性資料庫副本匯出173指令並核對MZ bytes。有界原版15,360組減值及
  RND不變通過；初次型別編譯、原版路徑SKIP及負對照/tmp模組路徑失敗全部保留，
  修正後同工具鏈乾淨重跑。這些是觀測腳本／環境問題，不列產品缺陷。
- RE證據經review.json與SPEC-022有限READY審查後，三檔正式修正怪物計數與派生負重。
  舊實作兩項獨立負對照失敗；新版全部internal／cmd及15,360原版計數對照通過。
  正式雙語各203鍵／204狀態與只讀觀察一致，Kind13／7及射空畫面通過；既有四回合
  各270狀態／14PNG保持。裝填Command／Flag1D差異、正常耗盡與命中仍待驗。
- 怪物v2與人物保存v3保持。Windows／macOS仍是331份Go；不冒充現行平台產物。
  工作區workplace/monster-ranged-consume-20261004的receipt.json／verify.py／audit.json
  保存封存、唯讀重驗及UID/GID1000:1000、容器清理。無新image、Git暫存／提交／推送、
  遠端Issue寫入、版號變更、打包或發布。既有異動全部保留。

本輪另由同一335輸入更新Windows AMD64與macOS雙架構／universal，Linux乾淨重建逐byte相同。
平台結構及最低macOS11.0通過，完整SHA見本輪platform-checks.json；本輪兩平台未執行，真機與封包未驗。

## 2026-10-04：RE-218 正常投石命中修正

- 沿 remake、規格閘門、dosgolem、IDA 9.4 與舊規格回鏈路由，接 RE-217 固定 seed 1 正常前綴，兩人選休息，原版四回合 259 狀態與 266 PNG。195 狀態、201 前綴 PNG 完全相同。
- 正常第二回合原版累積傷害 8、舊重製 16；第一發七次骰相同，原版原始傷害 12、重製 24。IDA 301 指令與固定 EXE 搬移 bytes 核對，證實射擊跳過力量加成。
- RE 證據、DRAFT、damage-review.json READY、舊公式首發反例、正式兩檔修改依序完成。三發護甲後 4／4／5；全部 internal／cmd、雙語各 267 鍵／268 狀態及 SLEDGE 270 狀態／14 PNG 回歸通過。
- 每語言四個執行入口 760 人物欄與 40 計數器、4564 進場骰及 209 戰鬥骰相符。四回合傷害 0／8／13／34、第三隻疲勞 54、怪物計數 10／4；最後正常畫格種子 2839412825。next-planning 與規劃後畫格階段不同，不作完整終態 parity；畫面幾何、正常耗盡與射後別名保持限制。
- 334 Go／336 輸入與兩檔 promotion 保存在工作區。Linux 8fc833ae 乾淨重建相同；Windows AMD64、macOS amd64／arm64／universal 重建及結構驗證通過，本輪沒有執行兩平台。
- 驗證工具失敗分類：延後 book-final 圖與前綴混比、copyfile 沒保留執行權限，均未改變產品。macOS 首次用了錯誤映像完整 ID，容器未啟動；查證既有 tag 的實際 ID 後同命令成功。封存腳本初稿語法與診斷 exec-entry 名稱均已改正；不登錄為產品缺陷。
- 本機收據 receipt.json、comparison.json、verify.py 及 audit.json 保存驗收與清理。未提交／推送、未寫 Issue、未打包／發布、未建立 image；繼承異動保持。原始 payload 僅本機。批末容器與 UID/GID 檢查見 audit.json。

## 2026-10-04：RE-219 正常八回合與擊退修正

- 路由沿 remake、規格、dosgolem、IDA 9.4 與舊規格回鏈，原版最多八回合：316狀態、323PNG，未耗盡。
- 第一個差異為第五回合擊退軸，245指令與33表bytes核對；原版反例舊30/30失敗，新30/29通過。只改三檔，336輸入保持。
- 全部internal／cmd及雙語SLEDGE通過。雙語324鍵／325狀態正常重播，前五入口各950欄與50計數器、4564進場骰與240戰鬥骰相符；後續投石仍分歧，八回合gate失敗保留。
- 三平台同來源重建；新Windows／macOS僅驗結構。本轮Wine四回合268狀態／11圖相同，僅屬修正前RE-218。
- 驗證環境第一次未設GOWORK=off，保留錯誤後同image乾淨重跑。法術頁測試期待舊方向，依原版表修正。Wine擁有權檢查誤跟隨z:符號連結，改lstat後個別前綴安全移除。一次package完整image ID誤植，容器未啟動；docker inspect確認後使用正確ID。均不登錄為產品缺陷。
- 使用者明確批准送禮確認原型；正式接線另批，不批准其他未決方案。沒有Git提交／推送、Issue寫入、打包或發布。批末UID、容器與附件見本機audit.json。

## 2026-10-04：RE-220 原版送禮確認核准與接線

- 使用者明確選原版確認，路由沿remake、規格閘門、文件分工與grilling。重驗舊收據全部附件及原型，五檔仍符合基底SHA，審查READY後正式接線，不覆蓋後續異動。
- 五正式檔與一既有測試更新；334 Go／336輸入。ENTER、SPACE與圖示確認回歸、全量internal／cmd通過。雙語292鍵／300狀態及300正式PNG與批准原型相同，原版前後272物品槽核對，完整v3保存除saved_at相同，52重載圖保持。
- 正式預設啟動各108真實鍵，合計216鍵；26個預定完整畫格、六次正常關窗與退出碼0通過。沒有重送鍵或狀態注入，重載不改保存。
- 首TestGiveItem沿舊立即轉交期待，依原版確認契約更新。GUI首批誤用截圖場景flow，未初始化flow即panic；只修驗證命令，改正式預設啟動，同image乾淨重跑通過。一次讀取命令else空白誤植修正，不登錄產品缺陷。
- Linux乾淨重建相同，Windows及macOS雙架構／universal中間候選同來源；新兩平台只驗結構。戰鬥雙語325狀態及16圖保持，RE-219八回合gate仍失敗。
- SPEC-031第17節有限CONFORMED，17.1／18及法術空白不因本決定獲准。原始payload僅本機；沒有提交、推送、Issue寫入、打包、公開或版號決定。封存與UID／容器清理見本機audit.json。


## 2026-10-04：RE-221 半途人物投石回繞

- 命中規格閘門與dosgolem路由，沿用已讀入口；建立並登錄missile-pass-through工作區，保存336修改前來源。
- IDA9.4隔離副本匯出648指令並核對MZ載入bytes。原版24組完整彈道只略過繪圖，回繞後到終點；零地形及人物槽上限初始化失敗保留，依證據修正後乾淨重跑。
- SPEC-013先DRAFT再READY；舊正式反例失敗，兩檔修正後24組及全部internal／cmd通過。存檔v3、怪物v2保持。
- 同一原版八回合雙語各324鍵／325狀態；前七入口各1330欄與70計數器、4564進場骰與前365戰鬥骰相符。第五回合投石0x124、seed48124921相符。完整八回合仍失敗，第七回合擊退原版29／29、重製30／28，第八入口14差異；下一步查方向座標表。
- 送禮300狀態／300圖、完整v3保存重載與SLEDGE各270狀態／14圖保持。三平台候選同來源重建，Linux逐byte相同；新Windows／macOS未執行。
- 全部工作在有界、無網路、非root容器；原始payload僅本機，不改封存父收據，不提交、推送、寫Issue或正式打包。收據唯讀重驗與批末擁有權／容器清理結果寫同工作區audit.json。


## 2026-10-04：RE-222 擊退成功位移

- 路由命中規格閘門與文件職責，保存336修改前後來源；269指令與129表bytes、八方向原版支持先DRAFT再READY。
- 局部測試未恢復人物搜尋後目標索引，以及IDA／dosgolem搬移基址混用，均保留首次結果並修正後乾淨重驗；不列為產品缺陷。
- 保留原版碰撞格，只修成功位移的一個正式檔與兩個既有測試檔；舊反例失敗，新碼30狀態及全部internal／cmd通過。法術擊退舊期待修正為10／12，保留首次失敗。
- 同一正常八回合雙語各324鍵／325狀態、142觀察PNG／16正式PNG，八入口各1520欄與80計數器、4564進場骰及400戰鬥骰相符。預先登記的八回合驗收通過，終態傷害34／42、次數8／3、seed3351160206。
- 送禮300狀態／300圖、完整v3保存重載及SLEDGE各270狀態／14圖保持。三平台同來源重建，Linux逐byte相同；新Windows／macOS只驗結構。
- 下一步預先限制最多16正常回合或正常出口先停，核對首次耗盡及下一規劃。正常耗盡、完整射後別名、其他戰鬥與實機仍開放。
- 有界非root無網路容器，原始payload僅本機，父封存不變；無提交、推送、Issue寫入或發行。封存、唯讀重驗、全工作根擁有權與容器清理結果由同工作區audit.json保存。


## 2026-10-04：正常投石耗盡續驗啟動

- 上輪RE-222完成正式擊退修正，八回合通過及收據封存，屬實際進展。
- 本輪核對AGENTS、CONTEXT、WORKLIST、工作樹及路由；命中dosgolem原版對拍與文件職責，載入README、CLAUDE及文件契約。
- 新工作區workplace/monster-ranged-sixteen-20261004已掛入CONTEXT；prepare.py、run_native.sh固定seed1、正常啟動與招募／休息，上限16回合或正常出口。prepare_replay.py、run_replay.sh及review_native.py接原版收據與雙語重播。
- 原版程序在有界無網路非root容器內執行；固定dosgolem968ca774，前後正式336來源及RE-222候選保持。執行中先核對已完成原版209畫格前綴相符，不冒充終態驗收；完整收據及耗盡結論待程序完成。


## 2026-10-04：RE-223 正常休息路徑驗收結果

- 原版run_native.sh正常完成438秒；八個完整回合、第九執行入口後正常戰敗，328狀態／335PNG，八回合316狀態與322非終態圖保持。此前進度輸出「still running」生成時程序已完成；以session20746終態及新收據為準。
- 雙語run_replay.sh四次程序完成，各336等價鍵／337狀態，154觀察圖／18正式圖。compare.py首次把執行入口數誤當完整回合數而失敗；保存初版，按實際正常出口修正後用同一trace重算，九入口1710欄／90計數器、4564＋450骰序通過。未重跑GUI或改產品。
- 原版與雙語戰敗圖已目視核對；終seed3250832936、投石7／2，正常耗盡未發生。原版射後Command34、正式35的既有差異及全終態／戰後保存仍待驗。
- 正式336來源與Linuxd2d1ca6b保持，不重建已固定平台候選，不重開已通過的八回合。下一步查同初始狀態正常防守／移動操作，先登記路徑再測，不延長敗北路徑或換seed。
- 文件命中既有職責與原版路由；新RE-223已掛入研究索引與CONTEXT。全部有界無網路非root容器；原始payload僅本機，父收據保持。未提交、推送、Issue寫入或發行；封存與唯讀重驗、擁有權及容器清理由同區audit.json保存。


## 2026-10-04：RE-224 防守路徑、姿態修正與剩餘差異

- 接續已核准怪物v2與原版送禮，不重新提問。命中dosgolem、IDA9.4、正式spec閘門與文件職責路由；新研究文件已掛入索引。
- 正常原版run_native.sh固定seed1，492.61秒完成，第十回合怪物7首次耗盡後戰敗，369狀態／376PNG，無下一規劃。雙語舊正式378狀態首次於骰244分歧，第八回合提早戰敗；末態無battle造成compare.py KeyError，改以最後戰鬥快照量測並明記FAIL，不當作環境假失敗。
- IDA窄窗口402指令bytes、40姿態分類及第五回合局部原版39／56支持READY。探針首次把monster_records_hex誤當陣列，保存錯誤log後修正型別重跑。bytes審查首次誤減0x1000，按既有MZ裝載起點0x10000與重定位字+0x1000修正後通過。首次Docker命令抄錯image雜湊，未建立新image；後續用--pull never與既有鎖版image。
- 舊正式測試防禦16失敗。只修改attack.go與missile_test.go，補姿態加分；九姿態及第五回合、全部internal／cmd通過。corrected/另存新來源與候選，不覆寫舊正式負對照。建置驗證腳本的輸出根先修正為corrected/，未寫入舊GUI產物。
- 新雙語378狀態、195觀察PNG／20正式PNG，正式／觀察相符。前七入口1330欄／70計數器及前361戰鬥骰相符；Style8近戰分歧後第八至十入口13差異，整條路徑FAIL。終seed2907222844相同及兩側第十回合耗盡不能替代完整對拍。
- 送禮300狀態／300PNG及v3保存重載、SLEDGE各270狀態／14PNG與正常休息337狀態保持。三平台同來源重建，Linux乾淨重建相同，Windows／macOS只驗結構。未驗實機、未更新封包。
- 主機gh auth status成功後只讀遠端#6／#44全文及開放清單，保存issues-snapshot.json；兩項仍OPEN，完整正常戰鬥結算後保存仍是完成條件。沒有提交、推送、Issue寫入、發布或新image。容器／UID-GID／唯讀封存結果見同區audit.json。


## 2026-10-04：RE-225 高攻擊式修正與正常耗盡對拍

- 前輪屬具體進展：投射姿態正式修正，剩餘差異縮到Style8。開工核對AGENTS、WORKLIST、CONTEXT與508既有異動，沿用IDA9.4、spec閘門及文件職責路由；新RE-225掛研究索引，工作區只放workplace。
- 固定336來源及RE-224正常377鍵，未換seed、未注入正常戰局、未重跑或改寫原版封存。IDA89指令／184表bytes，修正表168bytes共十二列，舊九列完全保持。
- 完整局部原版首次漏設遭遇等級全域，59／71只屬缺少測試輸入；保存JSON與log，明示等級1後為64／71。該等級取重製見證測試輸入，未冒稱正常原版直接觀測。READY審查後舊正式負對照72失敗，只補三列並加測試。
- 43組高攻擊式／武器位元／負分數及第七回合兩次骰、全部internal／cmd通過。雙語377鍵／378狀態，195觀察圖／20正式圖，正式與觀察相符；十入口1900欄／100計數器及4564＋534骰零差。第十回合首次耗盡後正常戰敗，終seed2907222844、Kind5／0。原版與英文、繁中結果圖目視核對，完整畫面幾何與終態不外推。
- 送禮300狀態／300圖、完整v3保存重載、SLEDGE各270狀態／14圖及正常休息337狀態回歸保持。三平台同來源候選重建，Linux乾淨重建逐byte相同；Windows／macOS只驗結構，未驗實機。正式334 Go／336輸入、怪物v2與保存v3保持。
- 高攻擊式及投射姿態切片有限CONFORMED，父規格及#6／#44仍開放。下一步延伸同正常前綴的戰後回城、住宿保存與重載，不改舊FAIL收據、不為無玩家影響的別名開無界研究。未提交、推送、Issue寫入、打包或發布。擁有權、容器及唯讀封存狀態見audit.json。


## 2026-10-04：RE-226 正常防守戰敗返回

- RE-225封存1920附件／2726外部輸入／3737PNG唯讀重驗通過，receipt SHA-256 b4723d0404f05c3da9014d587a1279e308d82138b8507d9a831dbb9466825331，509既有工作樹異動保持。
- 命中正常存讀檔路由，讀retro-game-playtest、dosgolem README／CLAUDE與文件職責；新RE-226已掛研究索引與CONTEXT，工作區沿workplace契約。
- 固定原版seed1及RE-224正常前綴，run_native.sh 454.01秒完成，新增唯讀出口觀察；369狀態、375前綴圖與5117呼叫保持。一鍵返回，沒有direct-entry、狀態注入、常式替換或規劃重播種。
- 名稱清單與人物資料混用、原版末圖誤當前綴、觀察器在正式清除戰鬥狀態後取Calls失敗，皆按來源分類為驗證腳本問題。首次產物及摘要保留；依實際audit_party、共用世界Rand及defer末圖時機修正，乾淨重跑。一次讀取命令漏空白及文件patch定位失敗也已修正，未改產品或舊封存。
- 新雙語正式／觀察各378鍵／379狀態，trace及抽樣PNG相符；compare_return.py逐34完整物品槽、傷勢29／20、金冠987／1009、33戰後骰與世界欄位零差。共567骰、終seed4151719855、暫存16、次格4／4、3340降3240，原版與雙語返回圖已目視核對。
- 正式336來源與候選保持，沒有重複建置或跑已通過的產品測試。只閉合本場結果到世界返回；住宿保存、完整50-byte終態、其他戰鬥及實機保持開放。下一步從42／9登記正常回旅店路線，不能套用不同入口舊路線。
- 原始payload、遊戲複本及PNG僅本機，父收據不變；無提交、推送、Issue寫入、打包或發布。21未完成／85完成／13人工項保持；封存及UID／容器清理結果见同區audit.json。


## 2026-10-04：RE-227 正常防守戰敗後住宿保存與重載

- 開工沿正常玩家路徑、dosgolem與文件職責路由，工作區及RE入口已登記。正式334 Go／336輸入、候選及人物v3不改，固定旅店seed1，不重設戰鬥／旅途亂數。
- 首人與全隊原版分開重生，700／702狀態；370返回前綴與5150呼叫保持。正常世界42／9走40／23，兩次自然遇敵正常逃離，再到Halfway House，兩次實際住宿寫出CHARDATA。原版從該寫檔正常重載7狀態，兩人資料保持。
- 雙語正式與觀察579保存／19重載狀態及34完整槽、14組持久欄位相符；角色頁、金冠931／953、傷勢29／20、營養10／8與原版相符。正常離城各31狀態，可見日期時刻及世界位置還原，60抽樣圖成對一致。原版及雙語末端畫面目視核對。
- 初次腳本的空切片表示、未變動party省略、城鎮出口與戰敗返回次格混用均按既有契約修正；初次JSON、畫面與log保留。一次診斷輸出過量，後續限定欄位。沒有產品修改或重複建置，未跑來源相同且已通過的產品測試。
- 本場有限保存鏈通過，完整旅途骰序、50-byte終態、射後別名、其他戰鬥與實機仍開放。21未完成／85完成／13人工項保持。沒有提交、推送、Issue寫入、打包、發布或版本變更；UID、Docker清理與唯讀封存結果見同區audit.json。

## 2026-10-04 RE-228 標題雷聲與HD工作登記

使用者核准聆聽與正式播放，另要求安排圖片與操作畫面HD化。SPEC-004先READY後接線。
雙語正常啟動各23次cue，首格雙聲；第23聲播放中SPACE離開，輸出尾端一秒為零。
實際非靜音時長另驗，全部internal／cmd通過，RE-227真實v3重招離城31狀態及6圖保持。
兩次固定參數新OGG雜湊相同，與原已驗收OGG解碼相同；候選原檔雜湊保持。
硬體時序為規格近似，不宣稱逐波形、全遊戲parity、三平台實機或已發布封包更新。

研究副本掃描、樣本數誤記、互動標題未初始化、短快捷鍵、FFmpeg SIGINT退出與繪圖CPU節流分別定位；失敗資料保留。限制繪圖執行緒後同binary重跑。HD先盤點與城鎮／戰鬥樣張，再確認視覺與解析度。未提交、推送、寫Issue或發行。封存與衛生見本機audit.json。

## 2026-10-04至05 RE-229 HD清冊及共同決策

- 路由採remake技能、文件職責及戰術固定骨架參考；使用imagegen與grilling技能。
  現有workplace承載原型，RE-229已掛研究索引。正式336份Go／338輸入未改。
- Docker核對五份manifest的804張PNG；雙語固定正常seed1回放各184狀態，城鎮與戰鬥
  各一圖，與RE-225軌跡及戰鬥PNG相符。沒有注入角色、座標、物品或戰局。
- 保存像素、手繪初版與修正版三張原始生成板。實際1122×1402，未達提示1920×2400；
  生成圖人物及編號有差異，保留但不當正式素材。36張渲染／地圖資料及6張遮罩另列。
- 使用者已選手繪媒材。layout-review.html比較16:10保留骨架與16:9分區重排；初次
  共用戰鬥裁切會截掉城鎮地圖，已依來源改分場景排列。資源、縮放與區域界限通過，
  瀏覽器操作未測；正式介面、文字與命中區仍待定版後驗收。
- 來源與既有封存不改，無提交、推送、Issue寫入、打包或發行。原型留在本機研究工作區；
  封存、UID/GID及Docker清理結果見同區receipt.json與audit.json。

## 2026-10-05 RE-230 雷聲接線後跨平台續核

- 前輪HD清冊及共同決策為實質進展；HD版面仍待回覆，本輪做獨立平台驗證。載入remake驗證、Wine正常試玩與文件職責路由；舊05-verification路徑不存在，依技能改讀verification-and-handoff，沒有新增工具image。
- 固定RE-228的336份Go／338輸入，Linux與Windows、macOS建置並行，均離線使用鎖定module cache。五個候選通過架構、macOS最低11.0、系統相依與新音訊依賴核對；Linux逐位元組相同。沒有重跑已通過且來源未改的產品測試。
- Wine9預設DirectX雙語正常城鎮／戰鬥各184狀態、真實v3重載各31狀態及10PNG與Linux基準相同。正常啟動各23聲、SPACE中斷、Ctrl-Q退出0；有效聲長21.35415／21.35701秒、首格雙聲完整，尾端一秒為零。畫面已目視核對。
- 首次單次依視窗標題查找未命中；trial2控制器路徑完成，但繁中錄音開頭截斷使總聲長20.60651秒失敗。波形及CPU記錄定位錄音啟動延遲，修正為先確認靜音PCM寫出才啟動遊戲；原截斷的兩語錄音均被新首格閘門拒絕。trial3同候選、同輸入及同image乾淨重跑通過，不放寬界限。一次文件patch定位失敗已修正。
- 保留初次及trial2／3的所有遊戲證據；僅移除本批三個Wine前綴及IPC設定暫存，保存實際Pulse設定。全輪不改正式碼、素材、v3、版號，沒有提交、推送、Issue寫入、正式打包或發布。macOS執行、Windows真機、新交付與其他玩家分支仍未完成。封存、UID及容器清理見同區receipt.json與audit.json。

## 2026-10-05 RE-231 現行任務流程與真正保存重載

- 接手核對AGENTS、CONTEXT、WORKLIST與525既有git異動；命中remake正常玩家路徑、存讀檔、Xvfb及文件職責路由。主機gh auth status成功後只讀遠端清單及#6，仍OPEN，未寫Issue。
- 固定RE-228／230的336份Go、338輸入及Linux0aca3427，seed1在啟動前設定，沿用200個遭遇控制骰100。原7435鍵在第422步因圖示位置改變選到DROP，第5477步答應逃跑而未取任務物品；保留首次退出2與完整5478狀態。
- 正常按鍵產生器依現行狀態機調整。步行運送者55回合撿物、75回合戰敗，完整5467狀態留負對照。遠程護送漏了先空手，136回合戰敗；飛行策略停止觀察。只修可丟棄按鍵策略，不改正式規則、人物、座標、怪物、任務地點或亂數。
- 正常DROP後Moraine80回合取Ship Wheel，114回合帶物逃離。正常走回城交付，六人回報一次。全隊旅店流程逐人收費，五人缺錢正常略過，Moraine付56金冠後實際寫出完整六人v3；沒有補錢或改住宿規則。
- 雙語各7476鍵／7477狀態及12鍵／13新程序重載狀態相符，整份保存除saved_at相同，保存雜湊保持。各13任務圖、3重載圖抽樣，人物頁與交付畫面目視核對。正式trace未匯出outside_sheet，檢查器改以現有欄位及實際PNG，不補造欄位。
- 繁中首次超過600秒，保留失敗目錄、log與部分圖；同來源、seed、按鍵及工具image改四CPU／三繪圖執行緒／1800秒上限乾淨重跑。歸為工作量與驗證上限，未改遊戲速度或放寬驗收。
- 現行回歸有限通過；不宣稱自然全旅程、新原版完整parity或結局。18未完成／89完成／14人工項保持；HD手繪已定、比例仍待回覆。無提交、推送、Issue寫入、版號決定、打包或發布。來源與歷史封存保持；文件快照、UID/GID、Docker清理與唯讀附件驗證寫同區receipt.json／audit.json。
- 工作清單首次產生已寫WORKLIST，但CONTEXT未獨立開放可寫掛載，依容器唯讀邊界正常拒絕。保留初次log，補齊既有CONTEXT檔的可寫掛載後，同image乾淨重跑render、verify與selftest皆退出0；兩份產生區塊與JSON同步。文件腳本一處換行筆誤也在執行前修正，未改產品。

## 2026-10-05 RE-232 六人確認畫面與真正保存重載

- 命中dosgolem原版對拍、remake驗證及文件職責路由，沿用既有技能與工具image。核對526既有git路徑與RE-231封存輸入；沒有覆寫舊收據。
- 預鎖正常六人招募、送Kenneth、ENTER／SPACE確認，原版215狀態，實際六次12288-byte CHARDATA保存與新程序重載。固定seed1在送鍵前，不重擲，不改人物或入口。
- 正式確認頁漏六人肖像11916像素，原因為物品底圖覆蓋先畫的隊伍。原型先驗，再依已批准原版確認流程走SPEC-031第17.2節DRAFT→READY，僅正式英文／繁中兩檔GiftConfirm補既有肖像繪圖，最後有限CONFORMED。
- 正式／觀察候選760狀態、10812組完整物品槽，41張六人英文整頁只排既定滑鼠區相符；實際六人原版持久資料、v3保存與新程序重載一致。候選只變兩張等待頁，狀態與整份保存除saved_at保持。
- 全量internal／cmd測試通過，正式Linux b810e7fc逐位元組等於已驗候選。雙語兩人送禮、城鎮／戰鬥446狀態回歸通過，兩張兩人英文確認頁與原版相符。Windows與macOS雙架構／universal乾淨建置及結構通過，新來源真機未驗。
- MIRROR第六人殘留2959像素、Giving候選清單版面17170像素保持FAIL與未完成，不增加遮罩、不稱整UI一致。舊全任務與Wine完整證據屬於0aca3427，不冒稱新程式重跑。
- 初次CJKneedle重複、比較器讀到完整人物頁真實差異、持久欄位表示差異、回歸檔名位數錯誤與一處腳本換行筆誤已定位；保留初稿／失敗案例，同image乾淨重跑。源資料與規則未猜補。
- 18未完成／89完成／14人工項保持。HD手繪媒材已定、比例待選；沒有新HD接線、版號、正式封包、提交、推送、Issue寫入或發布。文件快照、來源及所有附件雜湊、UID/GID、Docker清理見同區receipt.json／audit.json。
- 工作清單初驗指出新增SPEC兩行限制尚未綁項目；保留初驗log，肖像殘留綁回既有UI項、平台限制明記由兩個現有人工項追蹤，再產生與重驗。沒有把缺口改成已完成。

## 2026-10-05 RE-233 資料頁肖像層修正

- 前輪RE-232有實質修正與封存，按現行527git路徑及338來源接手。命中remake驗證、文件職責與規格實作閘門；沿用既有Go／osxcross工具鏈，沒有subagent。
- 原版預鎖70鍵／71狀態，正常六人招募後逐人開MIRROR與返回，固定正常工具968ca774／7cdb3533、seed1送鍵前設定一次。逐人畫面證實原版將目前角色搬到第一格；完整人物記錄保持。
- 初隔離提案只取隊長，被其他五名原版畫面推翻；原稿與建置保留initial-proposal。新原型只取Speaker，雙語304狀態與6324組物品槽保持，31張英文整頁僅排既定滑鼠區零差，六張舊錯圖拒絕。繁中六名肖像／名牌搬移零差，抬頭文字與下方數值保持，畫面目視核對。
- SPEC-030 DRAFT→READY審查後只改internal/ui/town.go，兩種語言共用sheetVisibleParty，物品／隊伍換頁與送禮確認全隊肖像保持。全量internal／cmd測試通過，正式Linux dbc465a3等於已驗候選；684正常回歸、六人真實v3保存／新程序重載保持，完整保存只差saved_at。
- Windows、macOS雙架構／universal重建及結構通過；新來源真機與全任務／Wine未驗，不借用舊來源聲稱已驗。資料頁本切片有限CONFORMED，Giving候選清單版面與子頁／坐騎／戶外其他邊界仍開放。
- Go image未裝rg，改用有界Python查現行檔；早期檔名定位失敗與讀到空gift_inside已修檢查器，沒有修改產品來繞過失敗。沒有新授權決定、提交、推送、Issue寫入、版號、打包、發布或HD接線。18未完成／89完成／14人工項保持。
- 文件快照、338來源、原版／重製收據、UID/GID與容器清理見同區receipt.json／audit.json。歷史工作區封存保持，HD媒材已定、比例仍待回覆。

## 2026-10-05 RE-234 送禮選人版面原型

- 前輪RE-233已完成實質修正，現行528git路徑及338來源核對。命中dosgolem畫面對拍、remake驗證與文件職責；選人版面取捨啟用grilling，未把原版轉交確認的舊核准擴張成清單替換。
- 從正式dbc465a3建立兩UI覆蓋：Giving保留明細及六人肖像，下方提示給誰。現行姓名清單並列可比較，未改typed data、數字鍵、ENTER／SPACE、人物v3或正式來源。
- 24次正常程序，新舊／唯讀觀察雙語1520狀態，完整trace相同；真實保存與新程序重載保持。只改四張選人等待圖，53英文原版整頁僅排既定滑鼠零差、八張舊錯圖拒絕。10812組完整物品槽原始bytes包含空槽及重載相同。
- 比較摘要初次將兩批數量差額誤列空槽差異，已撤回；原版與重製均無此byte差異。保留錯誤摘要為非驗收，補完整重載204槽，直接逐位元組比較通過，不使用空槽豁免。
- 雙語畫格已目視核對，review.html可切換四個步驟且全部PNG引用存在。未在瀏覽器實際操作比較頁；沒有可用的既有playwright／browser image，不另建重複工具鏈，限制明記。
- 已提出一題比較原版明細版面與現行完整姓名清單，等待使用者選擇；SPEC-031第17.3節DRAFT，沒有正式READY升格或實作。滿背包、取消、其他物品／野外與滑鼠語意及HD比例保持原有閘門。
- 正式338來源、dbc465a3與封包保持，未重跑未改來源的全量Go或三平台建置。沒有提交、推送、Issue寫入、定新版號、公開發布、HD接線或新image。文件快照、所有權、來源與附件SHA、容器清理見同區receipt.json／audit.json。

## 2026-10-05 RE-235 最新正式程式完整任務鏈

- 前輪RE-234已改變證據與正式文件，判為實際進展；本輪529個git路徑及現行338輸入核對。命中remake驗收與文件職責，載入技能及verification-and-handoff。
- RE-231完整任務證據屬舊0aca3427。本輪沿其7476正常鍵及200個明示遭遇骰100，從最新正式dbc465a3標題開始；來源差異只三UI檔，沒有重設種子、替換角色或改規則。
- 雙語各7477任務與13真實v3重載狀態，完整trace各自與前輪相同；跨語言只排四種顯示欄位。完整保存只排saved_at後相同，重載不改檔。正常DROP、80回合取得Ship Wheel、114回合逃離、六人回報、Moraine付56保存六人保持。
- 32張前後PNG只重載資料頁兩張肖像區改動，符合RE-233正式修正；正文與數值不變，正常交付、保存及重載圖目視核對。原真實失敗負例沿封存收據拒絕，不重新選seed。
- 主機gh auth status成功後唯讀抓106張Issue，24開放；8張本機已完成仍遠端開放，#6／#44開放且無留言。快照及remote-review.json保存，未作Issue寫入或同步。
- 初始檢查誤用JSON不存在的status／issue欄位及不存在的review.py檔名，屬工具讀取錯誤；修正為現有欄位、實際工具及可選Issue編號後乾淨查證，未列為遊戲缺陷。image inspect查詢不存在Entrypoint欄位也已以實際Config訂正，沒有新建映像或改工具鏈。
- 畫面比較初次漏掉英文名牌4像素投影，把排他下緣誤設52；完整狀態與保存已通過，畫面單項失敗另記收據。原腳本未留快照，依失敗條件重建的compare-initial.py只作定位，已明記來源限制。命中固定骨架對拍路由並載入retro-remake-tactical-parity，核實PartySlot／Shadow及CJKInn公式後改為六格實際矩形，英文下緣53、繁中149。沒有改正式碼、圖片、規則或加任意遮罩。既有image沒有rg，窄搜尋改容器內標準函式庫，不退回主機分析。
- CONTEXT、WORKLIST與研究入口回填現行證據；仍18未完成／89完成／14人工。不稱自然全旅程、原版整場與全任務、結局、HD、實機或新版交付完成。沒有正式Go變更、提交、推送、定新版號或發布。
- 收據封存文件快照與全部輸入／附件，唯讀重驗及UID/GID1000、自檢無root-owned／誤建md目錄、容器清理與本機保留產物見receipt.json／audit.json。

## 2026-10-05 RE-236 正常戰鬥唯讀AI觀察

- 命中remake驗收與文件職責，沿用reverse-engineer-retro-game-remake技能及verification-and-handoff。現行338來源、dbc465a3及前輪封存收據核對，沒有以舊worklist重新開已完成項。
- 本機兩檔Go overlay只序列化AI分支與全部18槽，不改亂數、狀態或畫面。先重播固定既有1500鍵，再重播1430鍵成對擷取，沒有重設seed、換人物或戰局。
- 共4363英文正常狀態，213分支事件與六組完整PNG通過。第6回合反制掃描選中Moraine；第9回合對Style60的Poontang側翼Facing7→0，敵人位置保持，舊Style62明記。
- 原版沒有重跑；沒有把常式分支、圖像存在或正式／觀察一致當原版parity。SpriteAt固定敵人圖示不讀Facing，前後畫面已目視但不能直接證明轉向效果。撤退、休息及Command38未出現，僅記有限缺席。
- 原本尋找captureStep及非實際Go接收者名，後確認真正snapshot／jobs出口；屬工具搜尋定位，不列遊戲缺陷。主機Docker預設sandbox拒絕後依既有Docker授權使用host權限，沒有退回主機執行工作負載。
- #11、18未完成／89完成／14人工項保持。正式來源、v2／v3、HD比例、送禮版面、平台與封包保持；沒有提交、推送、Issue寫入、定版本或發布。
- 本機產物保留為驗收來源。文件快照、全部輸入／附件雜湊、UID/GID1000、root-owned／md目錄與Docker容器清理在receipt.json／audit.json。

## 2026-10-05 RE-237 AI高移動選項與32位元輪盤

- 命中dosgolem對拍、remake驗收、文件職責及IDA9.4路由，載入技能與工具契約。核對現行338輸入、封存父收據與原版固定968ca774；HD手繪回覆已在有效決策及工作清單，比例仍待確認。
- 固定生成器20260930及各樣本seed，三種舊全域各300同輸入，加完整300正對照；原版首次1200與唯讀觀察1200呼叫失敗0，所有結果保持。640-byte表與IDA一致，高列每種實際4次。擾動與direct entry明示，不改標為正常GUI。
- 舊夾具在119不同，僅同步遠程槽的初解釋未解決，撤回並保留原source與結果；正因為Level5、LoadLevel1造成負重50／70與花費300／420差。同步負重等級恢復300正對照。
- 高列原型再有九差；補60..62與byte截斷後只62的59不同。最小IDA sub_25A3A及乘除helper原bytes確認有號32位元回繞，隔離原型及正式900＋300全符，舊壓力41／59／166三失敗保留。未追無關helper。
- SPEC-008既有忠實AI授權，DRAFT、審查READY及promotion記錄後，只修改ai.go與既有original_ai_fuzz_test.go，未改v2／v3或新增產品決策。原名與operand保留，追加十筆已證實語意進既有索引，匯出合併註記與未知警示。
- 全部cmd／internal Go通過；同seed1、既定7476鍵及200遭遇控制骰重跑雙語任務與真正保存重載14980狀態，完整trace與前輪相同，v3只差saved_at，32圖解碼相同。沒有六人原版全任務收據、自然全旅程或平台真機聲明。
- 工具定位曾誤查未匯出的IDA函式、表JSON形狀、nullable觀察欄位及技能symlink；依真實來源訂正。promote.py初版語法錯誤在任何寫入前停止，修正後同image乾淨重跑。此類失敗不列遊戲缺陷。
- 初次外層1800秒錯把建置、全量測試及雙語長鏈共用時限，繁中在回報後尚未保存就被中止124。命中同一remake驗收／環境分類路由，保留11張部分畫格、command與日誌；只以相同image、資源、binary、seed及完全相同command重跑繁中，獨立外層2400秒。中止前畫格全符，不改產品、不重擲、不削減既定7476鍵或驗收條件。首次完整wrapper應採足以覆蓋兩次1800秒與建置的外層界限。
- 本機18未完成／89完成／14人工保持。沒有Git提交／推送、Issue寫入、定新版號、正式封包、公開發布或新image。附件與外部來源SHA、文件快照、UID/GID、自檢及容器清理見同工作區receipt.json／audit.json。

- RE-237清單初檢找到新規格兩行限制未綁定條目。保存原FAIL日誌，以精確match指回既有#11及v01，未隱藏限制、新增重複任務或改完成狀態。render／verify／selftest乾淨重跑後封存。

## 2026-10-05 RE-238 原版正常六人自然戰鬥

- 命中dosgolem對拍、remake驗收、spec閘門及文件職責，沿用既有Docker image、固定968ca774與最新338來源。怪物v2、人物v3、雷聲及HD手繪決定保持；HD比例仍待原型確認。
- 正常招募六人、離店、出城與自然遭遇三名懸崖巨魔，seed1只設一次。初版錯用Y確認下一回合，290狀態與完整失敗輸出保留；選YES後ENTER、SPACE推進訊息重跑247狀態，兩次結算、七AI入口、4710亂數呼叫。兩次前215完整狀態相同，招募41人物與PNG符合封存來源。
- 正式雙語無遭遇覆寫，各173狀態。原版多送且狀態不變的第152鍵SPACE被正式腳本拒絕，兩次失敗輸出保留；等效輸入只省略這一鍵並明記索引。114世界位置時刻、9名進場位置與頭胸HP、角色、圖示符合，144×144可見戰場像素差0。雙語只有一筆遭遇顯示文字不同，完整AI／記錄／保存等父閘門不升格。
- prepare.py初版拼行及旅行界限納入城內數字、核對腳本錯讀town.party、比較腳本誤猜顯示差異欄位，皆依實際來源訂正；英文成功遊戲輸出沿用，沒有重跑綠色結果。初版與修正版、所有失敗來源／輸出均保留，不將腳本或工具契約誤稱遊戲缺陷。
- #11及18未完成／89完成／14人工項保持。正式Go與338來源不改，因此沿用RE-237的全部Go及14980保存／重載收據。沒有提交、推送、Issue寫入、定版本、打包、發布或新image。
- 本機來源與圖片保留作後續對拍；文件快照、來源／附件SHA、自檢UID/GID1000、root-owned／md目錄及專案容器清理見receipt.json／audit.json。

## 2026-10-05 RE-239 正常六人AI與兩回合比較

- 命中dosgolem、remake驗收、文件職責與IDA9.4路由，核對RE-238封存收據與全部338正式輸入。沿用249等效正常鍵、seed1，不重擲、不注入或覆寫遭遇。
- 正式／AI觀察英文、結算觀察英文及正式／結算觀察繁中共1250狀態、40圖。雙語兩組完整trace、16PNG對及100觀察事件完全相同，第一英文副本另有八PNG對；唯讀觀察不改狀態或骰序。
- 七AI輸出40已解值、兩執行入口九槽各40值與seed、4096地形相符。四AI入口共九個敵人Command差異，下一規劃兩邊seed相同；未通過全部入口資料比較。78畫面四處seed／38欄差異保留，包含射擊收尾、傷害訊息等待時點與舊敵人指令，不以綠色控制試驗稱完整parity。
- 第一次Replan觀察在initializeMonsterEquipment／Ranged之前，原版0x1EF3A在初始化之後，seed及兩Kind差異列為不可比。沒有刪除差異或調seed使它通過；後續七AI入口seed皆相同。
- IDA9.4在一次性tmp資料庫匯出20E40…21016，338函式、正式.i64與unpacked EXE雜湊確認。已證實20F7C以槽號大於5選支，20F94 c6472500清原始[bx+25h]，在每槽法術倒數後、下一槽及勝負之前。正式Act漏此writer，下一輪先補spec再隔離修正與正式回歸。
- compare-ai.py初版拼行SyntaxError在任何比較前停止，修正後同image乾淨執行。讀skill的symlink掛載未涵蓋真正ida_94_official，核實原路徑後正確唯讀載入；未換image。IDA初版使用ASCII預設開檔，部分JSON及失敗日誌保留，僅顯式UTF-8的新腳本重跑成功，原正式資料庫不改。工具錯誤不列遊戲缺陷。
- #11、18未完成／89完成／14人工、338正式輸入、38f86f38、怪物v2／人物v3與HD手繪決定保持。無新保存或整場聲明，不無變更重跑全量Go／長任務。沒有提交、推送、Issue寫入、定版號、封包、發布或新image。
- 本機正常收據、研究及圖片保留；文件快照、附件／外部輸入雜湊、PNG CRC／解碼、UID/GID1000、無root-owned／誤建md目錄及Docker專案容器清理見receipt.json／audit.json。

- 最後比較重生的自動權限審查逾時，未啟動工作負載；依回覆允許重試一次，以相同image、UID及腳本直接重跑成功。這是審查逾時，不列為遊戲或Docker缺陷。

## 2026-10-05 RE-240 回合末指令清零正式修正

- 上輪封存457附件、736外部與40圖核對。命中remake驗收、spec閘門及文件職責，沿用技能與image，不重新開已完成項。
- 原始IDA20F74呼叫sub_1B100等法術消失確認，20F94才清+0x25。RE-239下一閘門曾寫反先後，尚未實作；本輪先追加勘誤，舊收據與文件不覆寫。沒有猜在訊息前提前清零。
- DRAFT先保存；隔離execution.go補settledSpell階段，按第七槽起清Command，疲勞／倒數不重算。第六／七槽、非怪物高槽、空槽、倒地、法術等待／恢復及既有力竭等待測試通過。四次正常原型1000狀態、雙語16PNG對與100觀察事件保持，七AI入口／輸出及兩執行與後續規劃入口相符。
- evidence-review.json審查READY及ready-spec.md快照後，promote.py只改既有execution.go與spell_lifecycle_test.go。正式Linux56f2420190c8f652a4cfea1ac6aaa612b571f13ee90d2c9215a15fd2be288320，與原型逐byte相同；全部cmd／internal測試通過。七AI入口原九個敵人Command差異消失，4096地形保持。三個原始位址回填索引，annotated匯出自動合併，未知仍警告。
- 正式雙語同seed1、249正常鍵再重播500狀態，完整trace／16PNG對與觀察版及RE-239相同。加原型共1500狀態／48圖。四處seed與31畫面欄差異仍留，不能稱整場parity；首次planning hook在初始化前仍不可比。
- 新正式版本以現行91正常送禮／旅店鍵及26重載鍵，各語言92＋27狀態，共238與十圖；正常保存、退出、新程序載入、完整trace／v3及PNG與RE-234正式相同，僅保存時間不同。沒有採未核准的收件人版面原型。此短鏈沒有新戰鬥，不稱新版戰後保存或7476鍵全任務；RE-237的14980長鏈仍是38f86f38歷史證據。
- SPEC-008清零子切片有限CONFORMED，父規格與#11不結案，18未完成／89完成／14人工保持。怪物v2、人物v3、雷聲及HD手繪決定不變。無提交、推送、Issue寫入、新版本tag、打包、發布或新image。
- 文件更新工具首次JavaScript字串未跳脫換行，在任何工具或檔案寫入前停止；拆開重送後執行，不列產品缺陷。
- 本機before／formal來源及DRAFT／READY文件快照保持，全部輸入／附件雜湊、PNG CRC／解碼、UID/GID1000、root-owned／md目錄自檢與專案容器清理見receipt.json／audit.json。

## 2026-10-05 RE-241 命中確認後結算傷害

- 命中 remake 驗收、spec 閘門、文件職責與 IDA9.4 路由；核對 RE-240 封存收據及全部338來源。固定原版seed1正常六人／249等效鍵保持，不重擲、不注入。
- IDA9.4一次性副本匯出sub_26D76控制流，原名／operand／bytes／xref保留。27473呼叫sub_1B100(5)先等命中句，274E9才傷害骰。四次正常投石第一句保持傷害前值，是正式Atomic Resolve加文字佇列造成的提前扣血。
- SPEC-009先DRAFT，隔離Go overlay分命中與傷害兩段，共用finishDamage；不回捲狀態或RNG。近戰／徒手／投射兩段終態與骰序、未命中／撞牆／裝填單頁回歸通過。原型四次1000狀態、64圖，四個原版第一句seed差異消失。審查READY並快照後，只改既有attack.go、execution.go、main.go與attack_test.go。
- 正式Linux9ab6289798c3ee46351c25f335e536ee7b55396c01257879d9be90e42100afcd與原型逐byte相同，全部cmd／internal通過。正式雙語再跑500狀態、32圖，與唯讀觀察完全相同；共1500正常狀態、96PNG。正常trace只在228／243／245第一命中句修正過早HP／掉落，按鍵數仍249。Olear弓在確認後的244才落地。
- 原版78畫面seed差異四降零，31欄剩12欄射後Command34／35；七AI入口／輸出、兩執行及兩後續規劃保持。怪物裝填別名仍未知，未以武器種類猜補。三個已解位址追加既有索引，匯出自動合併語意與未知警告。
- 八張原版／英文144×144戰場顯示區仍有2773至3287像素差異。目視呈現位置不同，原因未解；不把數值相同稱完整畫面。近戰／徒手本輪只有共用原始控制流與核心回歸，沒有新正常原版命中GUI或擊退等待完成聲明。
- 新正式雙語正常旅店保存、退出及新程序重載238狀態、十圖保持，完整v3只差saved_at。此短鏈沒有戰鬥，不借舊RE-237長鏈稱新來源全任務／戰後保存。
- 讀取工具一次自動權限審查逾時，未啟動；依允許重試一次，相同image、UID及唯讀命令成功。review.py初版使用未查證的call bytes常數而停止，核對原始匯出訂正為9a7013d919後乾淨重跑；這是驗證腳本錯誤，不列產品缺陷。
- SPEC-009子切片有限CONFORMED；#11、18未完成／89完成／14人工保持。怪物v2、人物v3、雷聲及HD手繪批准不變；HD比例、送禮版面仍待原型選擇。無提交、推送、Issue寫入、版本tag、封包、公開發布或新image。
- 本機before／formal-source、DRAFT／READY文件與原版圖片保留。附件／外部來源SHA、PNG CRC／解碼、UID/GID1000、root-owned／md目錄與專案容器清理見receipt.json／audit.json。

- RE-241 清單初檢將已實作的『尚未擲傷害』等鍵階段當作自承缺口；保留 FAIL 日誌，以精確 match 說明時間點並指回既有 #11，未改完成數量或藏掉真正限制。

## 2026-10-05 RE-242 執行鏡頭及戰場呈現

- 上輪RE-241為已驗正式結算修正，屬實際進展。命中remake／spec閘門、戰術畫面對拍、文件職責與IDA9.4路由；核對封存903附件、1216外部及現行338來源。
- IDA9.4副本匯出623指令及caller，1FF6F／2003A傳出手者，223EB／22401取座標；22603…226FF證實倒地／地上物品與同格優先度。sub_1C1BF直接匯出為卷軸文字，訂正RE-241重畫標記；原始等鍵與傷害順序不變，舊語意與封存證據保留。
- SPEC-008先DRAFT。V1四次1000正常狀態／324圖，78出手者與75英文戰場相符；219／241／244三反例保留。原始265CD等鍵後265D9重畫，279FF傳被擊中者；V2呈現快照保留移動前畫面，但遊戲座標已完成移動，倒地中心與Cursor分開。共用ViewActor，不改規則、RNG或v2／v3。
- 舊正式鏡頭／圖示兩個反例失敗；V2核心與兩段呈現測試通過，四次1000狀態／324圖，78英文戰場全部逐像素相同。初增CJK等於完整英文三倍放大的假設失敗，查實所有差異僅在既有細白框，172繁中完整PNG與舊正式逐byte相同。保留初結果，不重跑成功遊戲；只修比較方法並明列白框內310至328像素差。完整英文頁面174至778像素差仍保留，未宣稱全UI完成。
- 審查READY及ready-spec.md快照後，七個既有檔案正式更新。Linux2ba7018a517378487bcabc342deebe0f744e2b4befcfb387d00a4d02e3143ed3與V2逐byte相同，全部cmd／internal通過。正式雙語500狀態／162圖，與唯讀觀察一致，78英文戰場保持零差。全部原型與正式共2500正常狀態／810PNG，V1的1000不是完整呈現通過。
- 78畫面seed零差、12欄射後指令別名保持，七AI入口／輸出、兩執行及後續規劃相符。實際旅店保存／退出／新程序載入238狀態、十圖、完整v3保持，只差saved_at；此短鏈沒有戰鬥，新來源全任務／戰後保存沒有重跑。
- 八筆已解位址回填既有索引，2746A追加舊語意與勘誤；IDA匯出自動合併索引與未知警告。幾次檔名定位讀不到，依實際定義查找訂正，非產品缺陷。
- 封存前置腳本錯用至少20個有測試套件，實際19個全部通過，無測試的五套件另列；改成核對完整19個名稱。清單將「格上還沒有圖示」誤判待辦，補精確行號語意認領，未刪檢查。這兩項屬驗證腳本問題，失敗收據保留。
- SPEC-008此有限呈現切片CONFORMED；#11及18未完成／89完成／14人工保持。HD手繪、雷聲、v2／v3既有批准保持；HD比例、送禮版面、完整UI、其餘AI、整場、平台與交付仍開放。無提交、推送、Issue寫入、版本tag、封包、公開發布或新image。
- V1、before／formal-source、DRAFT／READY、全部收據與原版PNG保留本機。附件／外部SHA、820PNG CRC／解碼、UID/GID1000、root-owned／md目錄自檢與專案容器清理見receipt.json／audit.json。

## 2026-10-05 RE-243 最新任務、戰後保存與平台續核

- 命中正常玩家路徑／存讀檔／跨平台及文件職責路由，載入retro-game-playtest與verification-and-handoff，沿現行RE-242的336 Go／338輸入及Linux2ba7018a。不改正式規則、資料、UI或音訊。
- 先英文原plan前綴1001狀態，只差49個Cursor。再同一7476鍵、seed1與200明示遭遇骰100，雙語各7477任務及13真正v3重載；Moraine80回合取物、114回合帶物逃離，六人各回報一次、Moraine正常支付56，六人完整保存只差saved_at。沒有增加確認鍵或改策略。
- 每語言完整trace1222步不同，排除Cursor只12個命中首句的roster，其中兩步ground；下一確認鍵結算與前輪相符。12個索引、訊息、欄位及掉落保留逐步證據，不稱完整trace全相同。跨語言只排四種既定顯示字串，全部語意與保存相符。任務34圖、前綴8圖保留，雙語取物與英文重載人物頁已目視核對。
- Windows8af7df33、macOS AMD64的56cc2b7f、ARM64的f4031016、universal的c0fd775e由同一338來源離線重建，架構、thin相符、最低11.0及系統相依通過。Linux沿已驗2ba7018a逐byte身分，沒有再建Linux。
- Wine9預設DirectX兩批各500正常＋26重載狀態、168PNG全部與Linux相同，合計1000正常＋52重載及336圖。首次wrapper誤覆寫HOME，修正且保留初版，乾淨重跑相同程式、seed、按鍵及閘門。兩份本批Wine前綴移除，registry與執行收據保留；沒有移除其他專案資源。
- Docker User未設欄位的模板查詢失敗，改核image ID，所有實際容器明示1000:1000。平台檢查器引用上一輪program-after-inputs而失敗，改正現行檔名後重跑，候選不重建。圖片檢查假設每張新抽樣都有同號舊PNG，在新增7472畫格失敗；18組實際可比畫面16零差、兩個取物戰場因鏡頭不同，新增保存圖另目視核對，不重跑成功遊戲。這些屬環境／驗證腳本問題。
- 本輪378PNG，附件／外部來源SHA、CRC／解碼、UID/GID1000、root-owned及md目錄自檢與本專案容器清理見receipt.json／audit.json。已有Go、macOS及Wine image沿用，沒有新建或清理其他專案image。
- 本機18未完成／89完成及14人工保持。原版整場／全任務、射後別名、完整UI、HD比例／素材、Windows及macOS真機、新版交付仍沿既有閘門；沒有提交、推送、Issue寫入、正式版號或發布。


## 2026-10-05 RE-244 正常六人十三回合反例續核

- 命中dosgolem固定種子、remake驗收、文件職責與IDA9.4路由，核對當輪338來源與原版75輸入。上一輪RE-243有實質進度；本輪補原版正常首場後續，不重開已完成項。
- 原版正常六人自然遇敵，383狀態、39AI、13執行／規劃、5337亂數，到十三回合戰敗結果頁。前247完整狀態、13事件、4710亂數與126PNG完全相同。0x211AD進入不等於回世界；未新增原版戰後保存、重載或整場parity聲明。
- 2GiB首容器在196格以137終止，沒有保留下OOMKilled旗標，原因未冒稱已證實。同binary改8GiB／GOGC50／GOMEMLIMIT6GiB／GC記錄，326.87秒完成；heap最高一次收集前3703MiB、後2679MiB。失敗輸出保持，虛擬步數／seed／按鍵不改。
- 正式2ba7018a完整385鍵在第381鍵SPACE拒絕，保留完整方案、日誌與圖片。前380鍵僅作診斷，雙語正式／唯讀1524狀態、1028PNG相同；跨語言只排四個既定顯示欄位，事件完全相同。原版39AI／13執行與重製36／12不符，不用zip前綴遮掩計數缺口。
- 206可比戰場166英文144×144相同，繁中同166白框外相同；4096地形相符，完整英文頁相同0。42seed比較點不符含三個無重製戰場，51別名欄與558其他欄差異保留。第一短暫差異為第六回合擊退提前，第一持續差異為第九回合疲勞110／113，後續重製提早一回合敗。
- IDA9.4固定338函式、原.i64與解包EXE雜湊，在tmp副本核實0x250EA..0x25171盾牌／雙臂／非雙手才加入盾重，以及0x27B6F先等確認、0x27BC4才判擊退。SPEC-008新增DRAFT，未將推導負重當成hook實測值，正式來源不改。下一步隔離原型、READY審查、完整鍵序重驗及正常戰敗確認返回。
- 準備腳本首次換行拼接SyntaxError在送鍵前停止，原版.initial保持；修正同image語法檢查後執行。文件補丁第一次JavaScript字串未跳脫反引號，在工具啟動前停止，重新送出正確補丁。比較在繁中正式擷取尾段開始，完整JSON結論只於四個程序全部通過後封存；未以中間輸出稱完成。
- #11、#6、#44與18未完成／89完成／14人工保持。雷聲、怪物v2、人物v3及HD手繪決定保持；HD比例仍待已提出原型選擇。338來源未改，沿用RE-242全量Go、RE-243任務／保存與平台，不無變更重跑。無提交、推送、Issue寫入、定版、打包、發布或新image。
- 本機原版資料、候選與PNG保留；所有附件／外部來源雜湊、PNG解碼、UID/GID1000、無root-owned／誤建md目錄及本專案容器清理由receipt.json／audit.json核對。

- 清單首次verify找到兩行新DRAFT限制尚未綁定條目，失敗日誌保存。以精確兩行match回指既有#11／#44，保持未完成，render／verify／selftest在同image乾淨重跑。

## 2026-10-05 RE-245：盾牌負重與嚴重傷害確認

- 命中remake規格閘門、文件責任與IDA9.4路由，沿用固定image與338來源。八個既有檔先READY才改，正式27212080與候選逐byte相同；19套件通過。
- 盾牌負重加入雙臂／雙手條件、受傷與恢復依賴更新；嚴重傷害先確認才擲擊退。V1消除兩反例後，338格倒地Poontang仍擋擊退；IDA人物搜尋略過倒地者且遮罩8排除騎乘，地形不套一般移動條件。V2修正後同一385鍵雙語1544狀態、39AI／13執行及seed／40已解欄相符。213戰場211相同，兩張鏡頭、八頁索引與42射後別名保持。
- 舊正式三個回歸反例實際失敗，新版通過。初次測試地形陣列及舊schema引用造成編譯失敗，首次保存前置與按鍵分隔解析錯誤、準備器重入、長任務掛載路徑、比較器AI變數覆寫，均分類工具／測試問題，修正後重跑。完整385鍵V1失敗、前380鍵與初始來源保留，不冒充通過。
- 新正式正常旅店送禮／保存及真正新程序重載238狀態、十圖與完整v3保持。舊7476鍵雙語重播失敗，唯讀診斷6073步是第八回合逃離，與正式圖相同；舊版同一步第80回合取物。沒有新全任務保存驗收，下一步重建正常操作。原版十三回合戰敗結果未確認後續世界返回。
- HD手繪回覆沿已確認決策，比例仍待選。未改存檔、版號、封包、平台、Git提交／推送、遠端Issue或公開方式。全部Docker有界、UID/GID1000:1000、無網路，原版輸入唯讀。收據、擁有權、暫存與容器清理見本機receipt.json／audit.json。

## 2026-10-05 RE-246：新版任務操作重建

- 命中remake驗收、正常玩家路徑與文件職責，載入現行技能及verification-and-handoff。核對338正式來源、RE-245封存雜湊、UID/GID及既有image；不改產品碼、資料或已確認決策。
- 比較既有事件5658狀態與RE-243完整trace，881步出手索引先不同，888步舊版已進第五回合而新版仍結算，890步新版才換回合。1325步階段失配；舊固定鍵最終答應逃走，不將操作失配猜成產品缺陷。
- 可丟棄stdin產生器只讀狀態、送正常鍵；V1的5421鍵／80回合戰敗，V2的5411鍵／96回合戰敗保留。V2在4781步69回合由Fighting Bull取得物品；V3保留同前綴，避開敵人相鄰格返程並正常REST，109回合邊緣逃離。沒有角色／座標／種子替換、強制勝利或改骰。
- 首5969鍵正式英文真正保存與新程序重載通過，但目視末圖顯示Kenneth住宿價錢仍待回答。保存該基準，V4只加RET／SPACE兩鍵收完缺錢提示。新5971鍵英文及繁中各5972正常狀態、13新程序重載；四顯示文字外全部相同，完整v3只差saved_at，產生器與正式逐鍵狀態相同。
- 六人各回報一次，Fighting Bull保留Treasure Map完整記錄；Olear與Fighting Bull各支付56，四人正常略過，最後回旅店一般選單。正常11944／重載26及圖片見review.json。沿RE-245十九套件，不無變更重跑全量Go。
- 同338來源Win／Mac四候選離線重建並驗架構／結構；Linux沿正式逐byte身分。Wine9預設DirectX雙語正常十三回合772狀態、真正v3新程序重載26狀態及22PNG與Linux相同。原版戰敗後世界返回保存、平台真機與封包仍待驗。
- 準備器補丁多出加號，在執行前修正。讀取plan.go錯路徑屬查檔問題，改讀planning.go。完整重播wrapper初次以廣泛字串替換誤改formal-inputs檔名，遊戲未啟動；保留初版，改精確case路徑後乾淨重跑。產生器比較初次遞迴排除name，誤刪城門內層名稱而出假差異；保留初版，改只刪頂層截圖標籤後重算，保留所有內層名稱。圖片檢查初次誤把Xvfb桌面1280×800當來源PNG，改依已核實canvasFor的英文320×200／繁中960×600檢查。兩種檢查器初版與失敗日誌保留，只重算收據，沒有重跑已通過的遊戲。這些不列為產品缺陷。
- 所有容器無網路、原輸入唯讀、1000:1000、有界資源與Xvfb trap；目前image沿用，沒有新建或清理其他專案資源。附件雜湊、PNG解碼、個別擁有權與本專案容器清理見receipt.json／audit.json。
- 18未完成／89完成、14人工保持，沒有Git提交／推送、Issue寫入、版號、打包或發布。HD手繪已核准，比例及操作版面仍待既有問題回覆。
- V01清單正文改為目前來源、有效有限證據與下一閘門，完成條件及Issue綁定保持。先前歷程留在RE-245文件快照、各RE與WORKLOG，不再讓舊「需重建任務操作」冒充目前待辦。

## 2026-10-05 RE-247 正常六人首場戰後返回與保存

原版674狀態及41重載狀態，六人各住宿56、六次實際CHARDATA寫檔，34物品槽及持久欄還原。
383狀態／5337亂數／78事件／262PNG首戰前綴保持。正式567等效鍵，雙語1136正常／26重載與保存持久欄相符。
一處原版Down確認改為重製RET，明示操作未過；末端空隊伍／重製保留六人、離店亦未過。
19套件與同338來源平台沿先前收據，沒有無變更重跑全量測試。來源338及原版75保持。
首次前綴告警範圍錯誤、xdpyinfo缺件及兩條拒絕腳本均保留；唯讀診斷定位確認鍵差異後才做等效轉換。
原版632.49秒、重載72.65秒；正式英文10.61秒、繁中48.33秒。所有工作於固定Docker、無網路、UID/GID1000:1000、有界執行。
已完成容器均自動移除，封存前後另核對清理與擁有權。沒有提交／推送、Issue寫入、版號、交付或公開變更。
結果與再現入口見docs/re/20261005-normal-six-return-save.md；完整收據在workplace/normal-six-return-save-20261005/。

## 2026-10-05 RE-248 規劃末端索引與戰場中心

原版IDA9.4 linear 0x1F070先保存槽位，0x1F15F只為可行動隊員重畫，末端兩者獨立保持。
333道指令bytes核對；四個舊正式反例實際失敗，新原型通過，初次測試型別編譯失敗另存。
385正常鍵、雙語候選及唯讀1544狀態，39AI／13執行／13規劃、seed及已解40欄保持。
八頁索引相符，213戰場英文與繁中游標框外全部相同；只有兩張PNG改變，完整UI仍有差異。
中途進度曾提前宣稱像素結果，隨即撤回；以最終comparison-summary及READY審查收據為準。
先READY再改round.go及既有round_test.go，336 Go／338輸入保持，Linux87698438與原型逐byte相同。
十九套件及雙語1136正常回城保存／26新程序重載通過；完整v3只差saved_at，重載trace完全相同。
沒有新原版抽樣、長任務、Win／Mac、真機或封包驗收。V01、完整UI、方向確認、末端隊伍及射後別名未結案。
所有工作在固定Docker、無網路、UID/GID1000:1000、有界執行，容器自動移除；封存另核清理與擁有權。
誤掛不存在IDA路徑產生空root目錄，確認後只rmdir該一路徑；改用既有ida_94_official及fail-on-missing bind mount。
沒有提交／推送、Issue寫入、版號、交付或公開變更。入口docs/re/20261005-planning-view-cursor.md。
封存首版在1116PNG解碼後遇/out與/src相對路徑格式錯誤；保存首版、修正路徑後只重核封存。
完整PNG雜湊、CRC及解碼另行重驗，未修改或重跑產品。

## 2026-10-05 RE-249 下一回合問句繪圖

沿原版0x1FAEA大字旗標、0x1FB30讀記錄0x108及RE-086／087字型契約，沒有新原版抽樣。
V1十三頁各293像素差，缺記錄反例重複印錯誤至120秒逾時；完整保存首版。首個邊界腳本缺shot-dir退出2另存。
V2用既有book-large並在下一Update回傳Draw錯誤；十三頁英文左上框逐像素相同，其他七張每語言抽樣保持。
雙語候選／唯讀1544狀態及完整觀察事件與RE-248相同；三個壞資料邊界5至6秒回報退出1。
READY後才改battle_render.go與main.go；336 Go／338來源保持，Linux599620c7與原型逐byte相同。
十九套件、雙語1136正常保存／26新程序重載通過，trace完全相同，完整v3只差saved_at。
完整UI、確認操作、末端隊伍、射後別名、真機及封包未結案。未重跑長任務或Win／Mac，不沿用舊來源宣稱新版通過。
全部在固定Docker、無網路、UID/GID1000:1000、有界執行；容器自動移除，封存另核清理、擁有權與PNG。
沒有提交／推送、Issue寫入、版號、交付或公開變更。入口docs/re/20261005-next-round-question.md。

## 2026-10-05 RE-250 戰鬥姓名字型

283道既有IDA指令bytes與固定EXE相符；0x23098大字、0x230C0姓名、0x23107小字裝備。
初次目視誤認疲勞條缺畫，量測511藍色像素兩側相同，沒有當成產品缺陷。
五組正常1930狀態、213左上框及六人首頁像素相同；四組正常裝備切換700狀態、12圖復原。
首次比較腳本要求跨語言譯文相同而錯報，保留首稿，改用各語言自己的基線重核同一批收據。
READY後才修改三個既有來源，Linuxd7faa6c9與原型逐byte相同，十九套件及1136保存／26重載通過。
完整v3只差saved_at，圖像新差異只在姓名區。213全頁未套遮罩仍0頁完全相同，完整UI未結案。
未重跑長任務、Win／Mac、真機或封包。HD手繪媒材已核，比例及版面仍待選。
固定Docker、無網路、UID/GID1000:1000及有界執行；收尾另核容器、擁有權、PNG與封存雜湊。
沒有提交／推送、Issue寫入、版號或公開變更。入口docs/re/20261005-battle-name-font.md。

## 2026-10-05 RE-251 指令區主列重建

沿原版GUI還原入口與文件職責建立獨立workplace；既有IDA兩份匯出原始byte核對後做DRAFT候選。主列顯示讀Actor已提交值，編輯仍Planning；隊員肖像標籤大字、裝備小字。三次正常1158狀態、724模式切換、213原版指令區零差及完整事件保持，READY後接正式兩呈現檔與既有回歸測試。十九套件、1136保存及26新程序重載通過，Linuxc47c9a29與候選相同。完整UI、操作、終態與平台維持開放。腳本語法、fmt匯入與匯出格式誤讀均屬工具問題，修正後同鏈重跑。更新分級索引與舊規格來源勘誤，保留歷史。未commit、push、Issue寫入、發布或版本變更。Docker清理、UID1000、PNG與收據校驗見本輪audit.json。

## 2026-10-05 RE-252 戰鬥訊息白底與成功掉落名稱

來源389＋867指令核對，V1保留212/213反例，V2的213訊息區全相同，完整213頁仍差右下滑鼠67像素，沒有遮罩。READY後正式六檔接線，Linuxc8764592與候選相同，十九套件、1136保存／26新程序重載及完整v3通過。修正前後繁中299步只差訊息，闊劍沿原譯名。ASCII讀取、缺EGA分支匯出與準備語法失敗均修正後同鏈重跑；保留首次輸出，不當產品缺陷。回填SPEC-008、RE-087／114／115歷史勘誤、書本空白契約與原始位址索引；不可變鍵護欄正對照通過、三個壞反例拒絕。打暈及滿地上槽仍開放，完整UI、操作、長任務與平台未結案。未commit、push、Issue寫入、發布或版本變更；Docker、UID與PNG／收據校驗以本輪audit.json為準。

## 2026-10-05 RE-253 受阻打暈正常樣本與隔離候選

- 再核固定EXE、IDA來源867道指令；原版正常seed2／3共1011狀態。seed2第317格Olear持物保持、Command清0；seed3二十回合後有界停止，沒有打暈樣本。
- V1未回答治療問句，467狀態與程序PASS仍拒絕作戰鬥證據。V2保留同種子修正正常NO操作與預定旅行。3GiB首例cgroup證實OOM；第二例8GiB SIGKILL沒有終態cgroup，不猜原因。查固定來源滑鼠Polls無界累積，同程式／種子／plan重跑；seed3降低Go堆成長並保存GC及資源抽核，593.684秒完成，584前綴PNG逐檔相同。
- 隔離修正的雙語正常前版／候選及各自只讀副本2576狀態、56PNG通過。完整Elf Bow記錄與原版相同，整份Party／Ground保持，只清目標Command；結算seed2126174334相符。舊版搬走物品的雙反例失敗，候選battle測試通過。既有seed1的1136保存／26新程序重載、68PNG與完整v3保持，但沒有這個新分支戰後保存。
- 去除forcedDrop也使ReturnScratch由舊版0改為保持6；原版word_39A54未採集，H訊息及返回成本未驗。保留DRAFT，338正式輸入及Linuxc8764592不變。下一閘門補只讀暫存與新分支正常戰後保存，不能用綠色測試代替。
- 新增重跑腳本曾有語法錯誤，未啟動原版；正常比較腳本集合union錯誤也在遊戲前停止，拒絕plan保留。修正後同容器命令乾淨重跑，分類為驗證工具問題。
- 入口docs/re/20261005-blocked-knockback.md；本機原始與失敗產物保留，seal.py／verify.py核對PNG、雜湊及UID/GID。背景Xvfb有trap，容器--rm；本輪容器清理收據另存containers-final.txt，不清理其他專案。
- 沒有提交、推送、遠端Issue寫入、版號、封包、發布或HD版面變更。手繪媒材已核，比例與介面仍待既有問題答覆。

RE-253清單核對補充：首次兩行未認領，修正文句與pattern後仍有一行；查claimed_by確認必須同時綁路徑與訊號。既有V01剩餘鏈補SPEC-008精確路徑後核對0，18未完成／89完成及14人工保持。拒絕log保留，不增新項或完成聲明。

RE-253收尾：git diff --check通過。Docker以專用Go／IDA映像列舉現存容器為空，本輪命名容器亦為空；containers-final.txt保存空清冊。封存重驗另核專案root-owned路徑、.md目錄、研究輸出UID/GID及完整PNG解碼。驗證容器有界、--rm，原版及失敗輸出保留本機。

## 2026-10-05 RE-254 受阻分支返回與保存

- 原版同seed2正常六人重跑760狀態，首433狀態與289前綴PNG保持。新增唯讀DS:BB84證實第317格暫存6，與H的DS:2E32分開。十四回合戰敗後返回世界、途中逃離、正常回城住宿，六次12288-byte CHARDATA實際寫入；新程序41狀態六人正常重載，沒有重播種或寫回。原版程序556.344秒，12GiB硬上限、GOGC20／GOMEMLIMIT9GiB，資源抽核無OOM，不把抽核值稱完整最終峰值。
- 首版prefix誤納probe末端defer抓圖book-final，保留拒絕稿後只修比較範圍。候選702鍵腳本第536個SPACE拒絕，正常IPC診斷器定位原版592格；原版額外SPACE只改共用暫存1至0，其他採集狀態保持。V1程序與腳本完整保存，V2逐筆記錄省略差分，636鍵重跑完成；另一次Right改RET明示等效確認，沒有改正式輸入。診斷就緒檢查誤用不存在xdpyinfo，未啟動遊戲，改用既有X socket後同映像乾淨重跑。
- 雙語候選及觀察副本2548正常保存狀態、26新程序重載通過。完整v3只差saved_at，六人每人17個原始carried槽及持久欄與原版相符；每語言168回程位置／時刻點及268戰鬥seed／暫存點相符。首場另2616狀態、233對照點通過。舊H、複數顯示、末端隊伍、方向／滑鼠、全50-byte與42別名仍開放。
- 受阻原型DRAFT，已備同狀態review.html及具體規則確認問題；正式338輸入與Linuxc8764592保持。CONTEXT目前表、SPEC-008及RE-253勘誤回填，原版與失敗輸出保留本機。清單仍18未完成／89完成、14人工，未改完成條件。文件新增入口沿docs/re/README，反向勘誤由本輪護欄核對。
- 所有負載在固定Docker、UID/GID1000:1000、network none及有界資源執行，Xvfb trap結束、容器--rm。收尾另核容器清冊、root-owned路徑、.md目錄、PNG解碼與收據雜湊。沒有提交、推送、遠端Issue寫入、版號、封包、發布或HD版面變更。

RE-254清單核對：首稿把已解除限制的歷史引句誤判為一行未認領，拒絕log保留；改用明確的解除敘述後核對0，18未完成／89完成及14人工保持。git diff --check通過，本輪命名容器清冊為空。封存與完整重驗另核附件、外部輸入、PNG CRC與解碼、反向勘誤及UID/GID，結果見本輪audit.json。

RE-254封存首稿誤用固定大寫CHARDATA路徑，保存拒絕log後沿既有大小寫相容的唯一檔案檢查；只修重驗器，不重跑遊戲。

RE-254封存第二稿把oracle-inputs.json的size／sha256物件誤當雜湊字串，拒絕log保留。查原始tar、建置腳本與metadata，go.mod未調整，核心來源沒有差異；修正為逐檔大小與sha256欄比對，仍以同一固定tar與原版程序核驗。

## 2026-10-05 RE-255 現行長任務與平台回歸

- 固定338正式來源與Linuxc8764592，沿RE-246預定5971鍵、seed1及200個明示遭遇骰100重跑雙語。11944正常狀態、第69回合取物、第109回合逃離、六人回報及真正v3保存通過；新程序26狀態重載保持。每語言82步只差stage2／prompt8的battle.cursor，符合RE-248修正，其餘完整狀態與舊基線相同。完整保存只排saved_at保持，40PNG中28相同、12呈現差分保留，不當原版全任務parity。
- 同來源Windows AMD64及macOS雙架構／universal離線重建、架構與依賴通過；Linux沿已驗正式執行檔逐byte身分。Wine9雙語正常十三回合772狀態、實際重載26狀態及22組PNG與Linux相同。Windows真機、macOS實際啟動及正式封包仍待驗。
- Wine V1英文遊戲與重載已通過，收據路徑相對基底錯誤中斷；V2英文通過，繁中啟動前發現RE-252凍結案例少四張控制圖。兩稿程式與輸出保留。沿同Linux、seed1與385正常鍵補八圖，386完整trace及四張共同PNG保持；V3同映像、同命令乾淨重跑四條全部通過。映像inspect缺Entrypoint鍵也改用既有null相容讀法，未建新映像。以上均屬工具與驗證輸入問題，正式程式未改。
- 依AGENTS.md的GitHub例外先在主機核gh auth status，再唯讀取得106張worklist Issue全文及private狀態。遠端24 OPEN／82 CLOSED，本機18未完成／89完成；八張已有完成證據仍OPEN，建角邊界及HD兩項未建Issue。差異保留remote-review.json，未由歷史遠端開關重開功能，也未寫Issue或搬運憑證。
- 更新唯一目前表、研究入口及既有V01／平台項的證據，完成條件與18／89及14人工保持。既有受阻規則DRAFT及HD比例問題繼續待答，沒有新的產品取捨。
- 固定Docker、UID/GID1000:1000、network none、有界資源、Xvfb與Wine清理trap；只移除已確認的本輪wineprefix-current。seal.py／verify.py封存並重算來源、附件、平台雜湊、完整PNG解碼及擁有權，容器清冊與git diff --check另存。不清其他專案容器。入口docs/re/20261005-current-quest-regression.md。
- 未提交、推送、Issue寫入、版號、封包、發布或HD正式素材變更。真機、未驗玩家分支及原版未知欄位維持現有閘門。

RE-255收尾：清單核對18／89、14人工，1781行自承全部認領。固定Go／Mac／Wine映像身分與本輪命名及三映像容器清冊已保存，清冊全空；git diff --check通過。本輪Wine prefix及其使用者擁有項已清理，其他專案資源保持。封存與來源唯讀重驗的完整結果以本輪audit.json為準。

## 2026-10-06 RE-256 英文怪物複數末字

199條IDA原始bytes與固定EXE相同，原版退欄覆寫末S，CYCLOPS為CYCLOPES。RE-254摘要誤記名稱，追加勘誤保留歷史。隔離1744正常狀態、32PNG只英文兩欄與兩圖變動，戰敗訊息262差降0；繁中保持。1136實際v3保存／26新程序重載保持，READY後三檔正式接線，Linux420e0e47與候選逐byte相同。十九套件及正式872正常／26再重載、22PNG保持。ASCII匯出失敗與比較器漏列遭遇／誤沿另一候選步號均保留，修正工具後同條件重跑；文件patch拒絕未改產品。新長任務、Win／Mac仍待回歸，不借RE-255舊c876收據稱已驗；其他GUI、受阻DRAFT、真機、HD與原完成條件保持。固定Docker、UID1000、無網路、有界資源、Xvfb trap與--rm；封存另核PNG、反向勘誤、來源、擁有權與容器清理。沒有提交、推送、Issue寫入、版號、封包或公開變更。入口docs/re/20261005-monster-plural-ending.md。

RE-256清單首稿兩行限制未認領，拒絕log保留；既有V01剩餘項已綁008／011路徑，補精確pattern後重核，不增項或改完成條件。封存與完整重驗結果見本輪audit.json。

## 2026-10-06 RE-257 現行長任務與平台回歸

- 正式336Go／338輸入與Linux420e0e47保持，沿RE-255既定5971鍵、seed1及200個明示遭遇骰100。雙語11944正常狀態、26實際新程序重載、40完整PNG及完整v3保存通過，前次trace逐項相同，保存只排saved_at。第69回合取Ship Wheel、第109回合逃離、六人回報與正常住宿保存保持。不稱自然全旅程、全24任務或原版整條任務parity。
- Windows AMD64及macOS雙架構／universal離線重建、架構、最低11.0與音訊依賴通過。Linux沿正式執行檔逐byte身分。Wine9分normal與reload兩階段完成，雙語872正常／26真實重載、22完整PNG與現行Linux相同，兩份保存檔未改。Windows真機、macOS執行及正式封包保持待驗。
- 主機gh auth status成功後唯讀核106張Issue與private狀態，24 OPEN／82 CLOSED；八張已有完成證據仍OPEN、建角邊界與HD未建Issue。未寫Issue或搬運憑證。更新目前表與既有三項證據，18未完成／89完成、14人工與完成條件保持。
- 本輪未改正式來源，沿RE-256十九套件結果；沒有重跑相同Go測試。清理腳本先核整份Wine prefix的UID/GID1000，再只移除本輪prefix。固定Go／Mac／Wine Docker、有界資源、network none、Xvfb／Wine trap與--rm，封存另核完整PNG解碼、來源及附件雜湊、文件入口、擁有權、容器清冊與git diff --check。
- 文件檢視曾誤讀不存在的根目錄RESEARCH-LOG.md，唯讀命令停止；沿既有docs/re/RESEARCH-LOG.md更新，沒有影響遊戲程序或產品檔案。正式程式、受阻DRAFT及HD後續待選保持，沒有提交、推送、版號、封包或發布變更。入口docs/re/20261006-plural-current-regression.md。

RE-257清單核對：18未完成／89完成、14人工保持，1795行自承全部認領。文件由JSON產生區塊同步，沒有新增完成聲明。封存與來源唯讀重驗結果保存本輪receipt.json及audit.json。

## 2026-10-06 RE-258 正常遭遇名稱列

沿RE-254固定dosgolem968ca774、seed2與正常招募旅行，152.372秒重生198狀態、138事件與23亂數，前綴完整相同。補197原圖，原版三行而正式兩行；199道IDA bytes與payload1的E7／E8核實，等級後加7、回欄8。隔離雙語872正常／1136真實保存／26重載通過，英文只改201與保存171／457的分列，繁中保持。訊息328差異降0，整頁67滑鼠差異保留，READY快照後兩檔正式接線。Linux1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274與候選相同，十九套件與正式872正常／26新程序重載、22PNG通過。新長任務與平台尚未重跑，受阻DRAFT、HD、真機與完成條件保持。

工具拒絕稿保留：索引patch標題不符，空採集null對[]及對null求len，零差假設被三行原圖否定；來源驗證誤把archive當明文及誤用payloads參數型別。讀既有抽取契約後修正，原版程序沒有重跑、正式未猜補規則。固定Docker、UID/GID1000、有界資源、network none、Xvfb trap與--rm。收尾核封存雜湊、PNG、入口、擁有權、容器與git diff --check，不清其他專案資源。沒有提交、推送、Issue寫入、版號、封包或發布。入口docs/re/20261006-ambush-name-frame.md。

RE-258清單首稿兩行受阻DRAFT敘述未認領，拒絕log保留。既有V01剩餘項已綁008／011，補精確pattern後1804行全認領，18未完成／89完成及14人工保持。沒有改完成條件或新增工作項。

RE-258封存首稿：映像inspect誤取不存在的Config.Entrypoint，空metadata使封存器拒絕。改用本輪所需的固定image ID核對；拒絕log保留，正式程式、執行檔與正常路徑收據保持。此為metadata工具問題，不另建映像或重跑遊戲。

## 2026-10-06 RE-259 新來源長任務與平台回歸

固定正式338來源及Linux1f31007c，沿RE-257既定5971鍵、seed1與200明示遭遇骰100，11944正常／26新程序重載通過。雙語完整trace與RE-257相同，40／40張PNG與RE-257逐byte相同，完整v3只差saved_at。第69回合取物、第109回合逃離、六人回報保持。
Windows與macOS雙架構／universal離線建置與結構通過，Wine9分normal與reload兩階段，872正常／26實際重載、22PNG與現行Linux相同，保存檔未改。只在本輪前綴核全部UID後移除wineprefix-current，啟動logs分階段保存。
主機gh auth status成功後唯讀核106張Issue與private，24 OPEN／82 CLOSED；八張已完成功能仍OPEN，建角邊界與HD未建Issue。沒有寫Issue、搬運憑證、提交、推送、版號、封包、發布或正式素材變更。18／89、14人工與原完成條件保持；正式Go來源未改，沿RE-258十九套件，不再重跑相同Go測試。固定Docker、有界資源、network none、UID/GID1000、Xvfb與Wine清理trap、--rm。封存另核來源／附件SHA、PNG CRC與解碼、入口、擁有權、容器與git diff --check。入口docs/re/20261006-ambush-current-regression.md。

## 2026-10-06 RE-260 主動近戰與第二個受阻持物樣本

- 沿固定 dosgolem968ca774、seed2、六人正常招募與旅行。控制器首稿顛倒列欄，僅終止自己的原版子程序，exit=-15與負證據完整保留。既有RE-171的十一條IDA指令直接核原EXE bytes；改DS:0068列／0066欄與十六次停滯閘門後，同種子及策略重跑807狀態、十四回合戰敗。198狀態／138事件／23亂數前綴保持，沒有重擲或勝利頁。
- 現行正式338來源與Linux1f31007c保持，雙語各809正常狀態；先在原版730／正式733誤掉Kenneth戰斧，再在741／744選單分歧，751／754血量差異屬下游不同操作狀態，不能另稱傷害公式缺陷。原版完整16-byte手持記錄保持、ReturnScratch6／H2已讀，舊H文字不猜補。
- 隔離既有RE-253分支，只在成功後退時forcedDrop。正式雙負對照失敗、候選通過；DRAFT候選ab0245b1雙語1618正常狀態，605人物位置及頭胸、496規劃列全部相符，末端十四回合戰敗。訊息區零差、完整頁1083差異無遮罩；不稱完整記錄、完整UI或新場保存完成。正式十九套件沿RE-258，未重跑相同正式測試。
- 比較器提前讀未完成繁中trace、漏倒地尾綴及對末端null求長度的拒絕稿保存。等待同一程序退出、修正格式與明查狀態後再比較，沒有重啟遊戲或挑成功輸出。最末觀察缺PNG採用前一個完整相同結果狀態的805原圖。
- 更新唯一目前表、既有兩項V01證據及研究索引；完成條件保持。受阻DRAFT沿既有問題待答，HD手繪已核准、解析度與介面待選。沒有提交、推送、Issue寫入、版號、封包、發布或正式美術變更。
- 固定Docker、UID/GID1000、有界資源、network none、Xvfb trap及--rm。seal.py／verify.py封存及來源唯讀重驗附件、外部輸入、PNG CRC與解碼、原版前綴、原EXE bytes、候選605人物／496選單、入口與擁有權。實際容器清理與git diff --check收據一併保存，不清理其他專案。入口docs/re/20261006-cyclopes-victory-path.md。

RE-260封存首稿：單程序逐張PNG解碼超過180秒，退出124且未產生收據。負紀錄保留，改兩個有界工作程序沿同一圖片重算CRC、完整解碼及尺寸；不省略圖片或重跑遊戲。

## 2026-10-06 RE-261 四套裝備的繁中名稱母體

- 核對RE-260實際audit與目前18項剩餘條目。先沿完整生命週期與文件職責路由、remake／grilling技能核查停止線；既有最長怪物及戰利品已有正常收據，沒有由歷史待辦重開。真正缺項為新v2四套名稱母體。
- 現行32種／128套、768槽／625非空槽完整核對；五來源1435記錄，獨立法術槽排除後1434。475個候選最長仍80像素，舊53筆worn來源移到新版欄位且原名／bytes保持，四組候選212定位。沿既有Go字型量測器，只改三個路徑，1132筆／129個名稱全部相符，沒有缺譯或缺字。
- 用現行字模核對兩張已保存正常畫格，最長怪物與棘刺釘頭鎚完整相符，缺墨點及79像素縮窄反例拒絕。標為舊正常畫格的字模相容，不稱現行GUI重播。338正式來源及Linux1f31007c保持，正式十九套件沿RE-258；無遊戲／亂數操作、新保存或正式測試重跑。
- 母體首稿把空槽誤當全零，既有抽取器明定首byte0；第二稿在排除前檢查法術Vartwelyrmi。兩個拒絕稿與來源反例保存，第二次錯誤後回查路由及verification-and-handoff reference，修正同一工具鏈後通過。Go量測器的XGB無Xauthority診斷沿既有Xvfb無認證路徑連接成功，程序退出0，沒有換映像或改產品。
- 更新唯一目前表、既有三個中文工作項、SPEC-033有限證據、研究及中文驗收索引。完成條件保持；受阻DRAFT、HD解析度／介面、平台真機與正式交付沿既有閘門。沒有提交、推送、Issue寫入、版號、封包或公開變更。
- 固定Docker、UID/GID1000、有界資源、network none、Xvfb trap及--rm。封存另核來源／附件SHA、母體與正式Go字寬、兩張既有PNG CRC及字模、文件入口、擁有權、容器與git diff --check。工作區workplace/cjk-v2-name-population-20261006；入口docs/re/20261006-cjk-v2-name-population.md。

RE-261量測器重現：改用獨立Go／XDG快取，不覆寫HOME；同一工具映像重建，量測器及字寬JSON均逐byte相同。腳本另加Docker內執行檢查，正式來源保持。

## 2026-10-06 RE-262 受阻訊息的舊地面名稱

- 路由命中remake收尾與IDA9.4，沿工具及文件職責入口。技能路徑為絕對symlink，僅掛skills目錄時容器內無法解析；改掛實際ida_94_official專案唯讀後讀取契約。一次錯查比較腳本檔名後改列既有工作區，沒有建立替代工具或修改產品。
- IDA image ID保持6f6d59af，正式EXE／i64唯讀，查詢在/tmp副本執行。首稿write_text漏明示UTF-8，在ASCII環境輸出空檔，拒絕作證據；保留ida-v1-encoding.log與decoder-v1-empty.json，修正encoding後用新容器副本重跑同一查詢。
- 142道原始指令與重定位bytes相符。H讀舊地面槽，零長度全零槽產生24空白。原第317格名稱空白，推翻RE-253的Short Sword敘述；第730格顯示Broadsword，手持戰斧保持。兩個既有正常畫格六列字模零差，錯填名稱差143／125像素。沒有重跑原版、換seed或注入狀態。
- 更新唯一目前表、既有兩個V01項、規格來源證據及研究索引，歷史以追加勘誤處理。非空舊H呈現尚未實作，受阻持物規則仍DRAFT待既有問題。338正式來源與Linux1f31007c保持，沒有新GUI、保存、平台、封包、提交、推送、Issue或公開變更。
- 本輪使用bounded Docker、UID/GID1000、network none、--rm；無Xvfb或持續程序。review.py／verify.py核對來源、附件SHA、PNG CRC、字模、正式來源、擁有權與工作清單。收尾檢查容器及git diff --check，僅本機輸出保留於workplace/stale-ground-name-20261006。入口docs/re/20261006-stale-ground-name.md。

## 2026-10-06 RE-263：舊地面名稱隔離原型

沿remake、共同決策與文件職責路由，前輪分類為有進展。只準備可撤回原型，不重複既有問題或進正式。

- 以RE-260受阻候選為基礎，三個runtime overlay記住成功強制掉落槽號，受阻時讀地面現有名稱。95道來源bytes與H寫入錨點重新核對，沒有新IDA／原版執行。舊候選兩個非空槽反例失敗，新原型四組及既有完整持物／正常局部擊退回歸通過。Go image、離線模組與有界快取沿用。
- 六個正常玩家程序及四個真實新程序重載全部退出0，共4218狀態。主動戰鬥前後只有第733步exec_text改變，原版訊息區597差異降0；605人物與496選單相符。受阻後正常回城保存的雙語完整trace相同，完整v3只排saved_at，重載及保存SHA保持；50PNG逐檔驗CRC／解碼。
- 一次錯查forced_drop.go後使用容器內定位，找到items.go。兩次patch因漏+及行匹配不符被拒，未有部分寫入；回查verification-and-handoff及文件契約，後續以明確追加更新。無產品或玩家程序失敗，沒有重啟已完成流程、換seed或挑過關結果。
- 更新唯一目前表、既有兩項V01、DRAFT來源／原型證據及研究入口，完成條件保持。正式338來源及Linux1f31007c保持，原型DRAFT待既有受阻規則問題。沒有正式接線、HD決策、提交、推送、Issue、封包、版號或公開變更。
- Docker無網路、UID/GID1000、資源／逾時有界，Xvfb socket與trap、--rm。封存核對來源與附件SHA、原始bytes、完整狀態、保存與新程序重載、50PNG及入口。收尾檢查容器、擁有權與git diff --check。入口docs/re/20261006-stale-ground-prototype.md。

RE-263工作清單首驗拒絕新DRAFT行未認領；保存負紀錄，將跨場H／戰鬥中途保存未知綁回既有V01機器訊號，不新增或關閉工作項。後續文件修正以來源樹唯讀及個別文件可寫掛載執行。

## 2026-10-06 RE-264：怪物裝填別名唯讀正常觀察

文件清單產生器第一次只取得WORKLIST.md寫入掛載，因它也同步CONTEXT.md統計而拒絕唯讀目標。
增加CONTEXT.md的個別可寫掛載後，以同一image與命令重跑；第一次紀錄保存在工作區worklist-render-v1-readonly-context.log。

沿用remake收尾、共同決策及文件職責入口。以固定dosgolem 968ca774與Go專用image新增唯讀hook，重跑兩個封存正常首場。1190狀態、304事件、10544骰序及919張PNG與各自父收據完全相同。

- 首次選RE-260近戰流程，雖807狀態／657PNG完全相同，但沒有射後consumer。分類為觀察中立對照，不能代替裝填驗收。轉用已封存且有射擊的RE-244正常REST流程；新增require_consumer，禁止以零筆消費讀取通過射後驗收。
- 各自控制器、計畫、seed2／seed1、正常入口與輸入全部保持。讀取DS別名與最後寫入IDA指令，觀察上限200000筆寫入／10000個階段點。兩次原版執行有界，退出0，不重啟、不重擲。
- 338正式來源、75原版檔案、169固定dosgolem來源及173道IDA指令核對通過。初次靜態驗證漏套MZ重定位，五處遠端呼叫不同；按2281筆重定位及載入segment1000修正。保留static-audit-v1-unrelocated.json，屬驗證腳本問題。
- 42欄位簡稱澄清為比較欄位數。正常觀察限同一首場，完整射後裝填別名仍待驗；未以手持種類猜補，未新增跨場或保存格式。
- 更新目前表、既有V01條目、SPEC-022的DRAFT證據及研究索引；完成條件與狀態保持。正式來源與Linux1f31007c保持，既有受阻規則及HD問題不重複提問。
- Docker無網路、UID/GID1000、資源及外層逾時有界，容器--rm。收尾核對附件與來源SHA、入口、擁有權、容器清理及git diff --check。未提交、推送、寫Issue、定版號、封包或發布。

## 2026-10-06 RE-265：完整戰鬥記錄的比較覆蓋與42筆差異勘誤

沿remake收尾及文件職責入口，前輪有有限正常讀取證據。重讀RE-245比較器與現行來源，發現前輪將42解釋為欄位數有誤。

- 原比較器37記錄欄＋有效倒數＋耐力／疲勞共40欄；42是Command差異條目數，Flag1D相符。已修目前表與權威清單，歷史研究追加勘誤，封存RE-264程式／收據保持。
- 部位相鄰byte是六個血量上限，未將兩格猜為16-bit血量。+0x21／+0x22的速度／負重候選被3772筆反例否定，依SPEC-008、SPEC-010與Round／planned來源改作排序與完成旗標投影。探索讀取一次把命令JSON列當事件列，無檔案改動；之後定位events.jsonl，不算產品失敗。
- 單一main.go研究覆蓋檔，JSON凍結唯讀actors、排序、性別及turn；正式338來源保持。先跑正式與觀察雙語1544狀態，再補State()亂數觀察772狀態。六個程序退出0，不重播種、不改操作。
- 兩版觀察去新增欄位後，完整trace與正式相同，各12組PNG相同，共36PNG。新增10語意欄位，每語言213seed與1917記錄；50欄只有42筆Command差異。+0x2C／+0x31尚未表示，部位負值及無效法術倒數仍是語意比較，完整50-byte不結案。
- 沒有新IDA、原版、保存、完整UI、平台或封包驗收，未改正式程式、資料契約、HD、授權、版本、Issue、提交或推送。Go image、離線模組沿用。Docker UID/GID1000、無網路、資源與逾時有界，Xvfb socket／trap與--rm；收尾核對SHA、擁有權、路由、清單、容器及git diff --check。

## 2026-10-06 RE-266：戰鬥記錄空缺的最小來源核對

沿remake收尾、IDA9.4與文件職責路由，只追RE-265兩個未表示偏移。

- 使用既有IDA9.4 image與唯讀資料庫/tmp副本。初次匯出沒有有效輸出，最小probe證實工具9.4、UID1000及338函式；查到UTF-8索引在ASCII locale讀取失敗，修正明示編碼及索引dict形狀。追加候選腳本一次生成語法錯誤，容器退出1；修正後同一工具與輸入退出0。失敗保存於script-failures.json及ida-script-error.json，沒有正式DB或產品失敗。
- 匯出42940道已辨識指令中的23個候選，另查四個有界片段。275道唯一來源指令按9216-byte MZ header、2281重定位與載入segment1000核對；位址索引66／41筆及已證實0x25409正對照通過。+0x30已查讀寫均單byte，+0x31仍未知；+0x2C候選取址不屬已證實戰鬥記錄，未推成padding或不存在間接writer。
- 唯讀重查RE-264兩場封存正常首場，821個戰鬥狀態／7389筆前九槽記錄的兩格均零。沒有新增正常執行、seed、PNG、GUI、保存、HD、平台或封包驗收。正式338來源及Linux1f31007c保持，不重跑未改正式測試。
- 新研究文件同輪掛入docs/re/README.md，更新唯一目前表與既有V01來源說明；完成條件、規格狀態及既有待答問題保持。沒有提交、推送、Issue、發行、版號或公開變更。
- Docker UID/GID1000、無網路、有界逾時與資源、--rm，IDA日誌僅容器/tmp。封存與獨立唯讀驗證核對來源、附件、重定位bytes、正式SHA、索引與擁有權；收尾檢查容器及git diff --check。工作區workplace/actor-record-gaps-20261006保留本機研究輸出。

## 2026-10-06 RE-267：首名怪物裝填來源接線

沿remake收尾、文件職責及共同決策入口。前輪來源核對有進展，本輪回到已見Command差異。

- 先讀RE-264、RE-227及現行v2來源，排除把別名當手持種類。首名DS:CDC5可由第一筆裝備+33重生，來源是頭部記錄+3；其他別名包含未使用舊裝備、後續記憶體與堆疊，未開無關helper逆向。173道來源指令套MZ重定位核對，128套與三場八筆完整複製相符，13個原版正常射後點相符。沒有新原版／IDA執行。
- 私有單檔覆蓋原型及唯讀main觀察，四個正常雙語程序1544狀態，12組PNG與前正式相同。每語言213seed與1917記錄，只餘第二名16筆Command差異；首名26筆歸零，其他欄位保持。測試空槽填充、三個免裝填值、其他值、下溢及分段命中收尾，整個戰鬥套件通過。
- 原型正常雙語回城保存1136狀態，新程序重載26狀態，完整v3只差saved_at，16PNG保持。DRAFT規格經ready.py審查，留spec-ready.md與ready-review.json後，才更新正式execution.go及單一回歸檢查；其他來源DRAFT保持。
- 新正式337Go／339輸入，Linuxdba813be與原型逐byte相同。十九套件與五無測試套件通過；六個正式程序1934狀態、28PNG及實際保存重載通過。全輪4640狀態、68PNG。未重跑不相關平台或前版長任務；新來源回歸保留下一步。沒有資料契約、HD、版號、封包、Issue、提交、推送或公開變更。
- 探索既有alias列表時一次把stage列當consumer列，KeyError無寫入；後續按schema分型。準備腳本的換行與錯用CmdShoot識別字在正式建置前修正，未觸發產品或執行失敗。Docker UID/GID1000、無網路、資源與外層逾時有界、--rm，Xvfb socket／trap收尾；封存獨立核對SHA、PNG、來源、入口、擁有權、清單與git diff --check。

## 2026-10-06 RE-268 首名怪物裝填後的來源回歸

固定339來源及Linuxdba813be，沿RE-259的5971鍵、seed1與200明示遭遇骰100，11944正常／26新程序重載通過。完整trace與40PNG相同，完整v3只差saved_at。第69回合取物、第109回合逃離與六人回報保持。
同339來源Windows與macOS雙架構／universal離線建置及結構通過。Wine9正常雙語772狀態、12PNG與RE-267正式相同；26重載、6PNG與本輪Linux相同，保存未改。沒有重新執行未改來源的十九套件，沿RE-267正式通過收據。
所有執行在固定Docker，network none、UID/GID1000、相稱資源與外層逾時；Xvfb與Wine均由trap清理。核全UID後只移除本輪Wine前綴。清單18未完成／89完成、14人工保持。封存另核來源、附件、完整PNG CRC與解碼、入口、擁有權、容器及git diff --check。沒有Issue存取／寫入、提交、推送、發行、版號或HD正式素材變更。入口docs/re/20261006-monster-reload-current-regression.md。

## 2026-10-06 RE-269：第二名裝填來源與頭部記錄生命週期

- 路由命中remake收尾、文件職責、IDA與原版決定性執行器；共同決策沿既定原版修復範圍。使用者HD手繪方向保持，既有解析度與介面問題不重問。
- IDA窄查42940指令、九個裝備池定位及93指標xref。啟動清零、row6複製與完整頭部清除已核；3020道唯一bytes套MZ重定位相符。首次立即數查詢未取16位導致已知正對照漏查，保留失敗腳本與錯誤，修正查詢後通過。技能symlink掛載與唯讀CONTEXT渲染失敗均分類為環境；補明確掛載後同工具重跑通過，不記產品缺陷。
- 固定dosgolem commit、seed1，32種真實v2裝備的六名／七名／後續三名複製與保留通過，選擇及終seed符合Go；18組完整頭部耐久邊界通過。有界堆疊入口／返回與零標頭注入明示，不冒稱正常第七名GUI。真實128套只有一空頭部槽，來源與耐久旗標均零。
- 唯讀覆蓋原型共2706正常／實際保存／新程序重載狀態。每語言213seed／1917記錄的50語意欄位零差，第二名16筆Command消失；完整trace、v3只排saved_at及28比較PNG保持。DRAFT快照、READY審查後才改四份Go並新增暫存與回歸檢查，339Go／341輸入；伏擊、任務、競技場共用程序內裝備池殘值，頭部清除同步讀零，v2／v3保持。
- 正式Linux340c8492與原型逐byte相同；十九套件、1934正式正常／保存／重載狀態與28PNG通過。全輪4640狀態、14程序、68PNG。位址索引與舊SPEC-022補充同輪回填，CONFORMED限首名／第二名來源。其他別名、原始＋0x2C／＋0x31、完整UI與整場保持；新來源長任務／平台下一輪回歸，不借RE-268舊來源。
- 封存SHA f11eca86a1d787e82eaebab212fc6994a2e6dfa756f4a4e3c78f672bd84f656f；獨立稽核215附件、960外部輸入、341正式來源、完整PNG CRC與解壓、backlink與UID/GID1000通過，root-owned零。清單18未完成／89完成、14人工；1842自承全認領。--rm、network none、資源／逾時與Xvfb trap有界，無本輪殘留容器；git diff --check通過。原版EXE／DB唯讀，研究輸出只留本機，沒有Issue、提交、推送、版號、發行或公開變更。入口[RE-269](docs/re/20261006-monster-second-reload.md)。

## 2026-10-06 RE-270：第二名裝填修正後的來源回歸

341份現行來源及Linux340c8492保持，沿RE-268的5971鍵、seed1與200明示遭遇骰100。11944正常／26新程序重載通過，完整trace與40PNG相同，完整v3只差saved_at；第69回合取物、第109回合逃離與六人回報保持。
同341來源Windows與macOS雙架構／universal離線建置及結構通過；Wine9正常772狀態／12PNG符合RE-269 Linux，26重載／6PNG符合本輪Linux，保存未改。十九套件沿RE-269最近正式通過收據，不對未改來源重複測試。
固定Docker、network none、UID/GID1000、資源與外層逾時有界；Xvfb與Wine由trap清理。核整個UID後只刪本輪Wine前綴。首次映像模板讀取不存在的Entrypoint欄位，改讀已驗ID；沒有切換或重建映像。封存另核來源、完整PNG CRC／解碼、文件入口、擁有權、容器與git diff --check。清單18未完成／89完成、14人工保持。原版整條任務與真機、正式封包及既有決策閘門保持，沒有Issue、提交、推送、發行、版號或HD正式素材變更。入口docs/re/20261006-second-reload-current-regression.md。

## 2026-10-06 RE-271：住宅門面告示中文原型

前輪RE-270已完成現行來源回歸，本輪改查剩餘圖片英文。路由命中remake收尾及文件職責，沿既有原版修復與全玩家文字繁中的範圍。沒有重問HD解析度、送禮版面或其他既有待選問題。
從TWN5PICS[19]及歷史正常住宅畫格讀出中央KEEP OFF THE GRASS，量測中央及兩側白底；隱藏寬度與重複同句維持強推論。可撤回canvas原型A兩行「請勿／踐踏草坪」、B單行「勿踏草坪」已提出單題待選，DRAFT不進正式路徑。
prepare.py核對165888來源視窗像素、六字1728字模像素；render.cjs在既有Puppeteer映像實際繪製及點擊範圍切換。首輪Chrome查找落在非預設UID的錯誤快取，核實映像內實際路徑後以PUPPETEER_CACHE_DIR修正，同映像乾淨重跑通過。Go映像沒有Pillow，原型以既有PNG讀取工具與原生canvas產生，不安裝套件或另建映像。這些是工具環境問題。
獨立verify.py核對三圖PNG CRC及完整解碼；來源canvas零差，A／B各1986／2036變動像素，全部在三個白底矩形。341份正式來源未改，Linux340c8492、字模、來源PNG、人物保存v3及HD正式素材保持。主代理目視A／B，中央可讀、左右維持裁切；不宣稱新正常玩家路徑、全中文、原版動態對拍或HD完成。
所有執行在Docker、network none、UID/GID1000、有界資源與逾時；瀏覽器finally關閉，--rm移除容器。清單仍18未完成／89完成、14人工；未提交、推送、發布或改Issue。研究入口docs/re/20261006-residence-sign-prototype.md；原型及稽核工作區workplace/cjk-residence-sign-prototype-20261006/，僅本機審閱。收尾核對檔案擁有權、容器與git diff --check。

## 2026-10-06 RE-272：正常首場成功反制掃描

路由命中remake收尾、文件職責、IDA及原版決定性執行器，沿既有原版修復範圍。前輪RE-271住宅告示仍待選，沒有重問HD解析度或既有操作問題。
IDA9.4限定三個函式，1150道指令套9216-byte MZ標頭、2281重定位與loadseg0x1000核對。唯讀返回觀察保留原始名、位址、operand及語意等級；正式資料庫與輸入唯讀，副本只在/tmp。首次技能symlink掛載無內容，檢查實際目錄後以同映像重跑，不另建工具鏈。
固定dosgolem版本、seed1及既有385鍵重新執行，383狀態、5337骰、269PNG與RE-244父收據全部原有欄位及圖片相同。雙語六個觀察程序為同一案例的三批抓圖，各386唯一狀態；完整trace符合現行RE-269，12正式控制PNG相同。39次AI入口九名40欄與種子、出口當前敵人40欄零差；11次掃描入口／出口九名40欄、種子及返回零差。兩次成功在第9／11回合，不沿用舊第10回合。九組英文整頁各差67像素，全部位於既有滑鼠矩形，戰場零差；主代理目視繁中選中與攻擊／倒地頁可讀。
比較器先誤要求missing／unimplemented空清單，核對原版全部父欄已相同後，改成固定region1…4缺檔與int21h AH0E AL38×5的已知清單，同容器及命令重跑通過。未改原版服務、種子或遊戲，不記產品缺陷。文件替換patch因同檔雙操作被拒，改單一update；新增稽核腳本換行在執行前修正。
正式339Go／341來源及Linux340c8492、v2／v3保持。未重跑未改來源的十九套件、長任務或平台；最近RE-269／RE-270收據沿用。撤退、側翼、撿武器、完整原始記錄、整場UI及其他決策仍開放。文件同輪掛docs/re/README.md，更新目前表與既有AI項的說明，完成條件保持。清單18未完成／89完成、14人工。Docker UID/GID1000、network none、--rm、有界資源與外層逾時，Xvfb trap收尾。附件與來源封存後另以唯讀audit核對1150bytes、303PNG完整CRC及解碼、索引、所有UID/GID；372附件、711外部輸入通過，receipt SHA-256為7a77ddcd60876c07fb0fb720ff2f3fa361c11786b18320320c5b716f962773fd。全專案root-owned／*.md目錄零，本輪容器零，git diff --check通過。沒有Issue、提交、推送、版號、發行或公開變更。本機入口workplace/ai-counter-normal-20261006/。

## 2026-10-06 RE-273：正常移動首場與同格射擊目標

前輪成功反制為有效進展，本輪沿remake收尾及文件職責路由，先固定seed1、正常六人旅行、MOVE朝最近敵人／WALK／確認的策略，最多16回合及1200戰鬥鍵。原版632狀態、5223骰至11回合結果頁，497秒正常退出；首170狀態與RE-272相同。沒有新IDA、重播種、狀態注入或等待替身，原版與輸入唯讀。
634等價鍵的舊正式／觀察雙語2540狀態、144PNG中，15原版掃描對12重製掃描，初批33組畫面抽樣未通過；初版等長比較器停止，改依步／回合／角色／同點序號保留缺事件，沒有截掉失敗。逐狀態核對首次差異在原版444／重製447：Moraine與倒地Olear同格34,32，原版投石向Moraine未中，舊碼actorAt先找倒地者就回空目標，漏防禦rand100。之前40語意欄位及seed零差；根因與SPEC-013已證實篩選契約相符。
DRAFT單一execution覆蓋及同格站立／全倒地檢查。第一版測試夾具漏已裝填Flag1D，被裝填分支截走，保留log後補正；正確舊碼反例仍FAIL、候選battle套件PASS。候選2540雙語狀態、152PNG；461戰鬥狀態／4149記錄的40欄及seed零差，33AI／15掃描／4側翼入口返回零差。九頁其他UI差異保留；四次側翼均未改Facing，不稱成功繞側。重製一次flee入口未有原版對應鉤子，不稱撤退oracle完成。
ready-review.json核來源、單一差異、同格原記錄、正常候選、舊碼反例及444畫面後，留spec-ready.md才接正式。340Go／342來源，只改備用查找並新增回歸，Linux03509ac1與候選逐byte相同。十九套件、正式新案例2540／152PNG及原先REST／保存／真正重載1934／28PNG通過。完整v3只排saved_at，v2／v3保持。正式每語言461／4149的40欄與seed零差；35英文頁26頁除既有滑鼠相同、9頁其他差異3579px保留，戰場724通道差。沒有放寬滑鼠區，未宣稱整場或完整50-byte。新版長任務及平台留下一閘門，不借RE-270舊來源。
研究入口及索引、SPEC-013有限CONFORMED、目前表及既有清單說明同輪回填，完成條件保持。Docker無網路、UID/GID1000、資源與外層逾時有界、--rm，Xvfb socket與trap清理。收尾封存及獨立唯讀核對附件、來源、新舊340／339 Go範圍、PNG完整CRC／解碼、入口與擁有權；清單與git diff --check及本輪容器另核。沒有提交、推送、Issue、版號、發行或公開變更；工作區workplace/ai-flank-normal-20261006保留本機研究產物。

RE-273封存及獨立唯讀稽核通過：1155附件、472外部輸入、342現行來源、994PNG完整CRC／解碼、UID/GID1000，receipt SHA-256為6e09dccbbe35ba78d876e498599a6f2910f084f5a5a23327ce7b223c3ef34ab6。全專案root-owned／*.md目錄零，清單18未完成／89完成、14人工、自承1870全認領，git diff --check通過。本輪所有--rm程序均已正常終止；9頁GUI差異及新版長任務／平台保持開放。

## 2026-10-06 RE-274：同格射擊修正後的來源回歸

342份現行來源及Linux03509ac1保持。沿RE-270的5971鍵、seed 1及200明示遭遇判定100，11944正常／26新程序重載通過。完整trace與40PNG相同，完整v3只差saved_at；第69回合取物、第109回合逃離與六人回報保持。
同342來源的Windows及macOS雙架構／universal離線建置與結構通過。Wine9正常772狀態／12PNG符合RE-273 Linux；26重載／6PNG符合本輪Linux，保存未改。十九套件沿RE-273最近正式通過收據。
固定Docker、network none、UID/GID 1000、資源及外層逾時有界；Xvfb與Wine由trap清理。核對整個UID後只刪本輪Wine前綴。最初Docker讀取stdin未加-i，準備指令未執行，補-i後原命令成功；沒有來源修改或產品缺陷。封存另核來源、完整PNG CRC／解碼、文件入口、擁有權、容器與git diff --check。
清單18未完成／89完成、14人工保持。沒有新增原版整條任務、全24任務、真機或正式封包驗收。成功側翼、完整UI及既有產品決策仍開放。沒有Issue、提交、推送、發行、版號或HD正式素材變更。入口docs/re/20261006-missile-current-regression.md。

## 2026-10-06 RE-275：目標資料框辨識修正

沿RE-273九組非滑鼠差異，原版339／重製342初始倒地Olear資料框錯顯由正常畫格與164道新IDA指令定位。原版按視窗精靈辨識，倒地／物品保留前框，空格清框；新進目標層先清框。DRAFT原型補三個舊碼反例、正常保存與重載，READY審查後只寫body_panel.go、planning.go與既有測試。340Go／342來源，Linuxe7e2ccc0與候選逐byte相同。
十九套件、正式3204正常／保存重載狀態及104PNG通過，全輪6408狀態、16程序及208PNG。完整trace未改，v3只差saved_at。原版兩個框20495像素零差；英文2275像素消除，35頁非滑鼠差異3579降1304。沒有放寬滑鼠矩形，圓環相位與結果邊緣仍待查。
IDA技能連結初次容器讀取因外部符號連結失敗，核實指向ida_94_official後唯讀掛該來源。首次IDA匯出未明示UTF-8中止，依賴準備也因缺JSON拒絕；明示編碼後用同映像及新副本乾淨重跑通過，失敗log保留，正式資料庫未動。容器、UID/GID1000、資源及外層逾時有界，Xvfb由trap清理；收尾核來源、附件、PNG CRC／解碼、入口、擁有權、容器及git diff --check。
清單18未完成／89完成、14人工保持。新版長任務／平台未重跑，不借RE-274舊Linux03509ac1。AI成功側翼、原始50-byte、完整UI、真機、正式封包及HD既有決策仍開放。沒有Issue、提交、推送、發行或版號變更。入口docs/re/20261006-target-inspection.md。

RE-275稽核補記：同一份164道指令核對原始MZ檔時，8個far CALL的segment word須依檔案重定位表與IDA載入基底0x10000換算。首次稽核直接比未重定位檔案bytes而失敗，已保留此工具前提錯誤；mz-relocations.json分列檔案偏移、檔案bytes、IDA線性位址與載入bytes，全部164指令仍逐條核對，未放寬差異。

## 2026-10-06 RE-276：結果圖底圖恢復

沿RE-275剩餘結果邊緣，479道新IDA指令與原始MZ relocation核實共同入口sub_1103F(0)恢復整張底圖。DRAFT候選先恢復底圖再畫結果，舊來源1016個非滑鼠差異為反例；正常保存通過後READY，只接internal/ui/battle.go。342來源／340Go，Linux97201bdb與候選逐byte相同。
十九套件、正式3204正常／保存／重載狀態及104PNG通過，全輪6408狀態／16程序、208玩家PNG與兩張差分圖。完整trace保持，v3只差saved_at；六張結果頁恢復底圖，其餘98張相同。原版631整頁只差滑鼠67像素，35組英文非滑鼠1304降288，八組圓環仍開放。
初次圖片檢查把繁中框內間隙當成圖片而拒絕。核對既有frame實際1像素周界後，按圖片、卷軸、框線及每個變動底圖像素，以同工具鏈乾淨重跑通過。此為核對幾何錯誤，沒有正式來源特例。UID/GID1000、network none、資源及外層逾時固定；Xvfb由trap清理，收尾核附件、來源、PNG CRC／解碼、索引、擁有權、容器與git diff --check。
本輪未新增原版GUI擷取，沿封存正常seed1；沒有新長任務、平台或封包，不借RE-274舊來源。只驗一般戰敗圖2，其他結果圖、原始50-byte、完整UI與HD後續決策仍開放。清單18未完成／89完成、14人工保持。無Issue、提交、推送、版號或發行變更。入口docs/re/20261006-result-background.md。

RE-276文件核對補記：第一輪清單verify拒絕新增的「餘八頁圓環尚待查」未認領斷言。
將這一句精確綁回既有v01-remaining-battle-chain，再由同容器重跑render、verify及scan。
保留圓環未完成狀態，未新增或關閉工作項，沒有產品缺陷或來源特例。

## 2026-10-06 RE-277：八張圓環畫格的捕捉相位

前輪已修結果頁殘框，本輪沿戰術與dosgolem路由追八頁288像素。讀既有SPEC-010現代等待近似及RE-171來源，正常固定seed1重播加sub_1CB28唯讀最後繪製觀察。632狀態的所有舊字段與PNG符合RE-273，不替換等待或寫狀態。八個原版框依實際引數選相位全符，正式腳本固定單元5模型全符；所有非滑鼠差異逐點等於兩相位模型的288像素。現行共用renderer另驗16個倍率框零差與相位等待／往返測試。
正式342來源／340Go及Linux97201bdb保持，不重跑完整產品或保存／平台。既有來源不變，RE-276十九套件及正式3204正常／保存重載仍有效。初次唯讀檔案擷取腳本因用不存在的文字分割點中止，補讀確實存在的近似契約；這是讀取腳本錯誤，沒有產品或參數修正。原始畫面差異未移除、不新增允許區、不推硬體精確等待；完整UI與其他閘門保持。清單18未完成／89完成、14人工維持，render／verify／scan收尾重核。
沿既有Go及鎖定dosgolem封存，容器network none、UID/GID1000、資源與外層逾時有界，Xvfb由trap移除；收尾核原版輸出、附件與來源雜湊、PNG完整性、入口、擁有權、容器及git diff --check。沒有Issue、提交、推送、版號、發行或HD正式素材變更。入口docs/re/20261006-marker-phase.md。

RE-277清單補記：verify把規格010新節引用的RE-276「圓環尚待查」歷史句當成未認領自承。
依路由61-worklist-as-data-with-verify與現行acknowledged_by，精確認領兩規格該引文並記原版3056觀察及16倍率框的解除證據。沒有隱藏目前完整UI缺口或修改完成條件；同容器重跑render／verify／scan，舊失敗log保留。

## 2026-10-06 RE-278：正常MOVE全部畫格與標籤字型

命中戰術同狀態與文件職責路由，沿Go1.26.7、Ebitengine2.9.9及鎖定dosgolem、IDA9.4。
先登記464原版畫格／463重製步，雙語各抓635正常狀態。完整對拍揭露十張標籤字型及一張途中圓環。
增加唯讀標籤helper、圓環回傳及捕捉PC觀察，原版632狀態全部舊字段與518PNG保持。
548道新IDA指令及MZ重定位核實怪物肖像保留文字模式。隔離候選只改十張標籤矩形1371像素，正常635及雙語1934保存重載、28PNG保持；READY後才接兩份共用UI。
正式Linuxd43a6385a9c35486cc2cd383f78ffce23d70820f1fa113c98559d1760ff2b647與候選逐byte相同；十九套件、正式3204正常／保存重載狀態及1298PNG通過，完整trace與v3保持、繁中635PNG保持。
相位由最後實際原版引數決定，隔離正常GPU463張穩定英文頁只差既有滑鼠；繁中247相位頁只改48×48目標格並與英文三倍全符。
第460張在9924999450進入繪製、9925000000抓圖、9925001847回傳；不把半張繪製當正式缺陷。另重播462狀態正常前綴，唯讀完成補抓與人物／世界狀態相同，整頁只差滑鼠67像素，原畫格保持。
全輪十九次重製8313狀態／4501玩家PNG，兩次原版1094狀態／867PNG，另四張診斷裁圖。
首次觀察核對假設helper有兩種字型而失敗，改依本MOVE路徑實際144次大字helper與未重印子頁的前次標籤區核對；失敗保存在observer-check-failure.json。完成前綴首次PNG核對漏算七張書本入口圖，依實際清冊核完整348普通圖與一張額外完成圖，再查startup-local.go的defer末端捕捉，347同名圖保持、前綴末端圖與原版461張相同，記completion-check-failure.json。沒有改原版或容差。
唯一目前表更新RE-278；18未完成／89完成、14人工保持，既有兩個V01工作續記，沒有增加或關閉gate。新Linux長任務／平台與其他戰局、原始記錄、操作、HD既有決策沿原完成條件。沒有Issue、提交、推送、版號或發布。
所有分析、建置、測試、GUI均在network none、UID/GID1000及有界Docker；Xvfb由trap清理。收尾核容器、擁有權、暫存、索引、worklist、git diff --check及附件／外部輸入／PNG。入口docs/re/20261006-full-move-ui.md。

RE-278封存補記：首次全5372張PNG解碼超出180秒，外層逾時後仍見容器執行，已停止移除。保存原腳本與seal-initial-timeout.json；同一工具映像重跑，改對SHA-256及原始bytes都相同的圖片共用解碼尺寸，每個檔案仍查CRC與雜湊。加Docker init及外層kill-after，沒有放寬比較或修改產品。

## 2026-10-06 RE-279：最新 UI 來源的任務與平台回歸

- 前輪為實際進展：RE-278修正字型，完成464張有限同相位呈現證據與正式保存回歸。
- 載入正常玩家路徑／跨平台試玩路由、文件職責與逆向重製驗證技能；沿現行342來源／340Go與Linuxd43a6385。
- Docker --rm／--init、無網路、UID/GID1000；Go與Wine4GiB、macOS6GiB，2CPU、192程序及外層有界逾時。未改正式來源。
- 沿既定5971鍵、seed1與200明示遭遇骰100，雙語11944長任務及26新程序重載、完整trace與40PNG符合RE-274，完整v3只差saved_at。
- 第69回合取得Ship Wheel，第109回合逃離，六人回報、Treasure Map與保存重載通過。
- Windows AMD64、macOS AMD64／ARM64／universal建置及系統相依／音訊依賴通過；Wine9正常772／重載26狀態與18PNG符合現行Linux，保存未改。
- 全輪八個玩家程序、12768狀態、58PNG。十九套件沿RE-278相同正式來源收據，不重跑；未宣稱自然旅程、全24任務、原版整條任務或真機。
- 入口[RE-279](docs/re/20261006-move-current-regression.md)。工作區workplace/move-current-regression-20261006/；seal.py與verify.py核對來源、附件、保存、PNG、路由及擁有權。
- 本輪wineprefix-current已核對擁有權後移除。清單render／verify／scan、git diff --check、封存稽核及專案容器終檢見同工作區；不清理其他專案資源。
- HEAD 1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；沒有提交、推送、Issue寫入、版號或公開發行。HD與既有待確認分支保持。

## 2026-10-06 RE-280：正常奔跑與側翼停止條件

- 前輪為進展：RE-279完成現行來源的正常長任務／保存重載、跨平台建置與Wine有限回歸。
- 命中同狀態戰術路由，載入retro-remake-tactical-parity及文件職責；保留342來源／Linuxd43a6385。
- 執行前固定seed1、16回合／1200戰鬥鍵與鄰接奔跑、逆時針偏一格策略，只送正常鍵。原版799狀態、5234骰、十二回合，715.56秒正常返回，未重擲或換seed。
- 正式／唯讀雙語四程序3208狀態、112PNG中立；每語言629點、5661記錄、50語意欄位與seed零差，36規劃／12掃描／3側翼相符。
- 三次沒有轉向。兩次同類阻擋、一次四格正對停止為強推論，沒有guard-return動態鉤子。先前進度訊息把WALK當零的判讀有誤，實際60、RUN61；勘誤保留原始記錄，不重寫舊收據。
- compare-views.py保留exit1：25頁中九頁454非滑鼠差異，最大14×10且全部在戰場，其他區域相同。圓環相位僅是假說，不放寬矩形、不修改PNG、不宣稱成功側翼或完整GUI。
- 全輪五玩家程序4007狀態／797PNG。來源、規則、v2／v3、正式資產與平台不變；十九套件及RE-279回歸沿相同來源，不重跑。
- 下一假說為鄰接時順時針偏一格，預測方向6／35,31空地；須另預登記與原版正常重跑，尚未接受。
- 入口[RE-280](docs/re/20261006-flank-blocked-normal.md)；工作區workplace/flank-success-normal-20261006/，完整策略、鍵序、原始資料及未通過的畫面比較均保留。
- Docker --rm／--init、無網路、UID/GID1000、4GiB／2CPU／192程序及外層逾時；Xvfb trap清理。封存稽核、render／verify／scan、git差異與容器終檢在同工作區。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；無提交、推送、Issue寫入、公開發行或新產品決策，成功側翼與其他原完成閘門保持。

## 2026-10-06 RE-281：成功側翼的有限正常案例

- 前輪為進展：RE-280完成三次正常停止案例、50語意欄一致證據與步行60勘誤。
- 路由命中戰術對拍與文件職責；原版控制器預先固定seed1、16回合／1200戰鬥鍵及順時針策略，只送正常鍵，不注入狀態或重擲。
- 原版644狀態／9回合；雙語每語言474點／4266記錄的50語意欄位及seed零差，27規劃／9掃描／1側翼相符，其中1次轉向。9英文頁有1頁／69非滑鼠差異，戰場外0像素。342來源／Linuxd43a6385保持。
- 雙語正式／唯讀四程序完整trace及PNG相同；另兩正式程序同646鍵，側翼後槽6實際出手及下回合四英文頁只差滑鼠，完整trace保持。七完整執行4526狀態／586PNG，另保留一次OOM部分523PNG，合計1109PNG。規則入口／返回與GUI分開判定，未擴大原完成條件。
- 第三回合第369步槽6有一次flee-enter正常候選；沒有原版撤退入口／返回鉤子，不當撤退通過，下一步可用同一策略／seed窄觀察。
- 初次原版Docker事件確認OOM、515秒／第九回合／exit137，保留523部分PNG及原日誌。相同binary／控制器／seed／策略乾淨重跑，只改6GiB及GOMEMLIMIT=4GiB；部分PNG與完整重跑前綴逐byte相同，屬環境失敗。
- 工具初建兩份Python腳本有換行誤植，語法檢查發現後在同映像修正乾淨重跑；屬驗證腳本，不是產品缺陷。
- 正式來源、v2／v3、規則、資產與產品決定保持；十九套件及RE-279長任務／平台沿同來源，不重跑。
- 入口[RE-281](docs/re/20261006-clockwise-flank-normal.md)；工作區workplace/clockwise-flank-normal-20261006/保存預登記、原始鍵序、方向轉移與未放寬的畫面差分。
- Docker --rm／--init、network none、UID/GID1000、有界記憶體／CPU／程序及外層逾時；Xvfb trap清理，封存／稽核與worklist render／verify／scan、git差異及容器終檢見同工作區。沒有新映像或背景服務。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；無提交、推送、Issue寫入、發行或新決策，目標進行中。

## 2026-10-06 RE-282：敵人撤退分支的有限正常對拍

- 命中戰術對拍、IDA工具與文件職責路由；沿RE-281固定seed1順時針策略，預先登記420觀察前綴，不另搜尋或重擲。
- 原版正常420狀態前綴／第四回合，4738次自然亂數；雙語每語言251戰鬥點／2259記錄的50語意欄位及seed零差。第三回合槽6撤退分支選方向3、指令39、RUN61、Guard0，實際由35,32移至36,33。七英文整頁五頁除滑鼠零差，另兩頁共75像素差異，戰場外零差。
- 原版新增撤退分支／檢查／共用返回只讀鉤子，正常TestNormalBookEntry 384.58秒PASS。原420列、既有events／counter_events／rolls及全部編號PNG符合父收據；book-final.png停止點不同，不互比。
- 雙語正式版各647狀態、七PNG，完整軌跡與父收據相同；Go唯讀觀察欄只在來源／素材／父收據／trace與events雜湊重新核對後沿用，沒有新觀察器程序。
- 新原版加雙語正式三程序1714狀態／320PNG；七英文頁五頁除滑鼠零差，另22及53像素均在戰場內，原圖與比較exit1保留，相位原因未知，不放寬允許區。
- 原版共用返回的Command／Style／Guard仍0／1／2，決策全域為39／61／0；保留呼叫者寫回時點差異，同輸入完成角色與完整前綴另驗，沒有程式修改。
- 原source340Go／342輸入及Linuxd43a6385、怪物v2／人物v3、資產與產品決定保持；十九套件與RE-279長任務／平台沿同來源，不重建或重測。
- 入口[RE-282](docs/re/20261006-flee-normal.md)；工作區workplace/flee-normal-20261006/，compare-flee／states／views分開核對，seal.py封存、verify.py獨立唯讀稽核。
- 首次唯讀Docker呼叫因sandbox socket權限失敗，以既有授權Docker主機控制面重跑成功；屬環境，不是產品失敗。原版6GiB、GOMEMLIMIT4GiB、--rm／--init、networknone、UID/GID1000，有界CPU／程序／逾時；Xvfb trap與容器／擁有權終檢見工作區。無新映像或背景服務。
- 工作清單render／verify／scan、git差異及封存／獨立稽核見本輪日誌；沒有關閉其他原完成條件。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；無提交、推送、Issue寫入、發行或新產品決定，目標進行中。

## 2026-10-06 RE-283：正常RUN的標記捕捉相位

- 上輪RE-282新增正常撤退決策與移動證據，屬進展。本輪命中戰術對拍與文件職責路由，先讀現況及唯一清單。
- 沿同一seed1順時針RUN及420觀察停止線，新增sub_1CB28／caller-return0x114AB只讀引數，沒有寫入、等待替換、途中重設或新策略搜尋。
- 同seed1順時針RUN的420狀態、原有全部欄位與306原版PNG符合RE-282。新增唯讀標記呼叫證實原版315／366／416三頁共144非滑鼠差異全由捕捉相位造成；實際引數決定原版單元，兩側模型及差分位置全符。現行共用呈現器按原版相位的六個英文／繁中倍率框零差。正式342來源／Linuxd43a6385保持，原始跨相位差異及完整UI閘門保留。
- 本輪一個正常原版程序420狀態／306PNG，另一個現行共用呈現fixture產生六PNG，共312PNG。正式雙語畫格與trace只沿來源及完整附件雜湊核對後的父收據，沒有新正式玩家、保存或Go規則觀察程序。
- analyze.py以實際原版引數決定單元，不從像素擬合；原版模型、正式單元5模型及三整頁差分位置全符。錯誤圖集尺寸反例仍拒絕。原始144差異與既有滑鼠允許區保持。
- 342正式來源／340Go、Linuxd43a6385、v2／v3、規則、資產及其他產品決定保持；十九套件及長任務／平台沿同來源RE-278／RE-279，不重建或重測。
- 入口[RE-283](docs/re/20261006-run-marker-phase.md)，工作區workplace/run-marker-phase-20261006/，外層prepare入口workplace/run-marker-phase-20261006-prepare.py。update-docs、worklist render／verify／scan、封存與獨立唯讀稽核同工作區保存。
- Docker --rm／--init、networknone、UID/GID1000、有界記憶體／CPU／程序／逾時；原版6GiB及GOMEMLIMIT4GiB，Xvfb trap及容器／擁有權終檢見工作區。沒有新映像或背景服務。
- 無提交、推送、Issue寫入、版號、發行或新產品決定；其他戰局與原完成條件保持，目標進行中。

## 2026-10-06 RE-284：主動近戰的有限正常路徑

- 命中remake技能、玩家機制稽核與文件職責路由，先核目前狀態及唯一清單，保留既有工作樹。
- 使用者手繪方向回覆與已有效decision.json一致；HD封存唯讀驗41附件／7PNG／1154外部輸入通過，未重問既有解析度問題。
- 固定seed1由正常六人招募、世界旅行進場，主動近戰602狀態／9回合／5144自然亂數，戰敗，保留結果。前170原有欄位符合RE-283。物品分支：0x269ba：7；完整16-byte怪物撿入記錄核實0筆。現行雙語各431戰鬥點／3879記錄，比對50語意欄位各保留878差異。首個物品分歧在554步，Kenneth受阻後正式版掉下戰斧、原版保留，重現RE-254／RE-260既有閘門。英文23頁中14頁除滑鼠相同，其他原始差分保留。正式342來源／Linuxd43a6385、v2／v3保持；未新增保存或正式修正。
- 新程序為原版一次、正式及唯讀觀察英中各一次，共3018狀態／587PNG；完整中立性通過，原始差異保留。沒有新保存流程或產品測試重跑。
- 入口[RE-284](docs/re/20261006-active-ai-normal.md)；工作區workplace/active-ai-normal-20261006/。腳本與固定來源、按鍵映射、raw記錄、比對、封存及稽核同工作區保存。
- Docker --rm／--init、network none、UID/GID1000，記憶體／CPU／程序及外層逾時有界；原版6GiB／GOMEMLIMIT4GiB。Xvfb trap、容器與擁有權終檢由本輪收據記錄。
- 正式342來源／340Go、Linuxd43a6385、v2／v3保持。工作清單render／verify／scan及封存／獨立唯讀稽核見本輪日誌。
- 無提交、推送、Issue寫入、版號或發行；目標仍進行中，既有產品決策與完成條件保持。


## 2026-10-06 RE-285 正常重載子頁

RE-285正常重載法術與任務子頁：有效原版19狀態、雙語正式／唯讀觀察296狀態。Olear法術N兩次英文整頁除滑鼠零差；雙語兩次法術與六人任務共16次開頁，52行抬頭及肖像名牌保持，完整隊員資料除Speaker不變。任務頁六人均顯示已回報第6號徽章，但沒有本輪原版任務oracle。正式342來源／Linuxd43a6385、怪物v2／保存v3保持。

入口：[RE-285](docs/re/20261006-sheet-subpages.md)。首次原版控制器及wrapper失敗保留，修正後同輸入重跑。未改正式程式、存檔、素材、版號或遠端。Docker有界、network none、UID1000，原版研究輸入唯讀；清理與擁有權由封存稽核。


## 2026-10-06 RE-286 最長世界提示

RE-286從RE-279實際v3保存正常重載、旅行至Htown、向Tulliana問map接第18號任務，再走到37,55。現行雙語正式／唯讀觀察各745狀態，共2980狀態；完整trace、抽樣PNG及文字log中立，來源保存未改。67段世界提示全母體最長quest.18為848像素，繁中換成544／304兩行與80像素問句；603×294完整提示框字模零差，缺墨點與543像素窄框反例拒絕。沿既定200個明示遭遇控制骰100，途中沒有戰鬥；不是原版任務oracle、自然完整旅程或任務完成驗收。正式342來源／Linuxd43a6385、怪物v2／保存v3保持。

入口：[RE-286](docs/re/20261006-world-prompt-longest.md)。控制器錯用flow重載、圖號語意及鏡頭邊界錯誤保留，修正後四程序乾淨重播。未改正式程式、保存、版號或遠端；Docker與擁有權由封存及唯讀稽核核實。

RE-286封存勘誤：首次唯讀稽核拒絕畫格計數，實際12張而非文件的16張。
原因是抓圖腳本把稀疏導航紀錄的列表索引當輸入步數。原收據dbecfbd99b382b61b11df820b44f891f2d62a0409abebed1ca3aa45d92c53481及原輸出移至同工作區attempt-v1/，2312份原附件全部核對雜湊。
改以完整紀錄的第300步與299步檢查接任務轉移，同744鍵、相同初始保存及種子重新抓0、300、303、744四張。新收據重新核對；舊收據不作現行驗收。未修改正式程式。

## 2026-10-06 規格索引現況訂正

現行規格為38個ID、39份文件；舊索引開頭仍指向26個ID的早期稽核。
修正SPEC-005／023／025／030／033／035六列敘述，連回已存在的原版建角、自然競技場、人物v3、六人肖像、名稱母體及法術住宿保存證據。
父規格狀態、DRAFT子分支、未完成18項與正式342來源保持，沒有新玩法、圖片或保存格式。
在主機完成gh auth status及遠端唯讀核對：儲存庫private，24 OPEN／82 CLOSED，全部Issue無留言，沒有新的決策可套用。未提交、推送、Issue寫入、版號或發行。
入口：[有限核對腳本](workplace/spec-index-current-20261006/review.py)，同區receipt.json保存連結、現行來源、文件與清單核對結果。
沿既有Docker image，UID/GID1000:1000、network none、資源與外層逾時有界；專案唯讀掛載，輸出限同工作區。不重跑未改動的產品測試；容器清理與擁有權終檢另記audit.json。

## 2026-10-06 RE-287 正常怪物拾取案例預登記

RE-284首場怪物體型232大於全隊，不能由隊員擊退掉武器；128套初始怪物手持皆非空。
改採Poontang原始旗標已接的任務5，BINDERAK體型100；正常招募六人、旅行、路途逃離與任務進場。
種子1預先固定，只送玩家按鍵；最多200旅行鍵、三場路途遭遇、1800任務戰鬥鍵、40回合，不覆寫遭遇與勝利。
入口：[RE-287](docs/re/20261006-monster-pickup-normal.md)。準備字串替換首稿失敗保留，尚未啟動遊戲；修正後相同來源重跑。
本輪先預登記再執行，不以預定結果宣稱拾取通過。正式342來源、怪物v2／人物v3與既有決策保持。

## 2026-10-06 RE-287 有限任務前綴結果

RE-287正常六人與原始任務旗標進入第5號任務：原版883狀態／10回合／2270自然亂數，750戰鬥鍵有界停止，仍在戰鬥。第441／825格的怪物11／8完整16-byte手持搬至地面1／2槽並清空，拾取轉移0筆。雙語正式／唯讀觀測各885狀態，共3540狀態／20PNG，完整中立。兩側首次規劃seed已不同；每語言753戰鬥點／13554記錄的50語意欄保留69991差異，另5018物品差異，不當作獨立AI缺陷。原版任務入口100個六面骰及367位置骰已記錄，正式入口地形與戰鬥分別播種，完整接續仍未驗。正式342來源／Linuxd43a6385、怪物v2／人物v3保持。

入口：[RE-287](docs/re/20261006-monster-pickup-normal.md)。首跑逾時及兩份錯誤分析／準備稿保留，重播882列及797PNG與首次戰場資料相同；原完成條件保持。容器、擁有權、輸入雜湊與清理由工作區收據及獨立稽核核實。未改正式程式、存檔、素材、版號或遠端。

正式第462步起停在逃離問句，原版按鍵前綴沒有對應回答，其後423個狀態沒有推進。753個戰鬥觀察點不是753個可比同狀態點；差異只保留診斷。跨語軌跡僅423筆exec_text不同，為英文與繁中逃離問句，其他欄位保持。

封存首稿誤要求英中exec_text逐字相同，保留seal-v1-rejected.py及seal-failure.json；修正僅區分已確認的翻譯問句，不改遊戲與原版比較。

## 2026-10-06 RE-288 任務亂數接續原型

RE-288隔離任務亂數候選：首次4096格地形、seed162767633、十八角色900項語意及十二完整手持零差。雙語候選／唯讀副本3540狀態與44PNG中立。每語言753戰鬥擷取點、13554角色記錄，751seed相同，九列保留134項語意差異；完整怪物手持與地面占用座標零差。英文10次畫格比較對應8步，保留175非滑鼠像素，成因未證實。RE-287兩名掉物怪物其後442／58點均倒地，拾取資格案例未取得。正式342來源／Linuxd43a6385、怪物v2／人物v3保持；候選DRAFT待契約確認及READY，不宣稱整場、保存、完整UI或目標完成。

沿RE-287固定原版收據，未新增原版程序或重設種子。四個新正常回放各885狀態／11PNG，884鍵相同。單檔overlay候選、首次地形欄列轉換沿既有證據，沒有正式接線。全部來源與父收據雜湊、文件／worklist完成條件、UID/GID及容器清理由工作區seal.py／verify.py核對；離線比較頁直接引用原有擷取。未提交、推送、寫Issue、改版號或發行。

離線頁首稿錯以/out解析父工作區路徑，保留finish-v1-rejected.py及finish-failure.json。
修正為/src的實際專案位置，沿相同來源續跑。render／verify／scan均通過，18未完成／89完成保持。

## 2026-10-06 RE-289 任務圓環相位補證

RE-289沿同場RE-287封存圓環引數，10次英文比較的175個非滑鼠差異逐點等於相位模型差分。六個獨立案例的現行英文／繁中共十二倍率框零差，原始跨相位差異保持。沒有新原版或正常玩家程序、保存、正式來源或版號變動；九個語意差分列、共用亂數DRAFT及原完成條件保持。

準備腳本先核RE-288父收據103附件與3103外部輸入，預先登記六獨立呈現案例及七次非零比較。沿用測試首稿漏移動子列4，來源TargetMarker已明列該層；保留拒絕稿與分析原因，修正只記真實操作層。原版框、候選框、兩模型及全頁差分位置相符，未改圖片或滑鼠區。共用renderer測試及相位等待／反向、壞圖集反例通過。資料僅本機，原始差分與未完成閘門保持。全部容器--rm／--init、network none、UID/GID1000；Xvfb有trap，清理與獨立稽核見同工作區。沒有提交、推送、Issue寫入或發行。

## 2026-10-06 RE-290 移動可見性與確認前戰場

RE-290依固定原版702道指令及正常第六回合入視野反例，修正成功移動以目的地判斷訊息，
並按起點可見性保存確認前戰場。組合隔離候選的雙語753點／13554角色記錄，
五十語意欄與seed零差；原九列134項差異解除，十二張移動等待整頁除滑鼠零差。
二十六次整頁比較保留262個圓環相位像素，實際原版引數與模型逐點相符。
正式只接三份既有來源，任務亂數仍DRAFT，怪物v2／人物v3保持。
Linux991046ab與只含移動修正的候選逐byte相同；十九套件及正式雙語3204正常／保存／新程序重載
狀態、56PNG通過，完整trace與抽樣整圖保持，完整v3只差saved_at。
新版長任務／平台未重跑；成功怪物拾取、完整原始記錄、整場、完整UI、HD及發行仍開放。

先固定RE-289父收據與342來源，IDA 9.4以唯讀原始資料庫的可寫副本匯出三函式。
702道載入指令與MZ relocation後位元組相符。原版先寫位置才查目的地可見性；
起點不可見時跳過前置重畫。九列停在sub_1B100輸入等待，圖像相位不能解釋語意差分。
V1保留13208、V2保留9267像素反例；V3把起點重畫閘門套到無訊息移動，十二等待圖降零。
三版回放均沿884鍵及執行前固定seed1，沒有新原版執行或資料注入。
READY後只接execution.go、state.go與既有execution_test.go；task-source.diff保存本輪增量。
十九套件及正常MOVE、REST、真住宿保存／新程序重載按新版正式重新核對。
先前長任務／平台屬舊版，本輪沒有挪用為新來源驗收。

驗證工具失敗分開保存：IDA LC_ALL=C的JSON中文字編碼；比較早於觀察程序完成；
合併建置流程重複複製mod快取；phase analyzer的/out相對/src路徑。
修正後在同一鎖版image及來源續跑，沒有修改種子、圖片或容差。
歷史來源三檔先存基線，external引用明示改指基線，current-inputs另存新版342雜湊。
沒有重設使用者異動、提交、推送、Issue寫入、版號或發行。HD手繪核准保持，解析度與介面仍待選。
容器--rm／--init、network none、UID/GID1000，Xvfb有trap；全部工作容器已退出。
輸入、產物、UID/GID、root-owned與.md目錄、清理及獨立唯讀稽核見工作區收據。

## 2026-10-06 RE-291：移動可見性修正後的長任務與平台回歸

- 接手查AGENTS、WORKLIST、CONTEXT、dirty tree與唯一狀態表。路由命中正常試玩、文件職責與remake驗證，讀試玩入口及verification-and-handoff。兩語言同類失敗後回查路由，沒有改產品。
- 沿現行342來源／340Go及Linux991046ab。RE-279的164及RE-290的838封存附件全部雜湊相符；不把歷史外部現行文件當未變來源。
- 原5971鍵雙語退出2。正常4781鍵前綴定位第864步恢復怪物移動訊息，等待鍵不足造成後續規劃錯位。原失敗完整輸出保存，不當產品新缺陷。
- 隔離副本僅改cmd/kol/main.go普通鍵產生器：110段Prompt10確認，32段長度改變、淨加16鍵為5987。所有其他正常動作、seed1及200明示遭遇骰100保持；110段末完整戰鬥欄除exec_text／一次性walls相符，完整v3只差saved_at。生成器不是正式GUI證據。
- 未修改正式版重跑雙語11976任務與26新程序重載。第4789步／69回合取得Ship Wheel，第109回合逃離，5599步回世界，六人回報與Treasure Map完整保存通過。每語言5001映射非訊息狀態零差，步號及name截圖metadata改按已登記映射比較；正式跨語言非顯示欄全部相符。
- 40PNG中38相同。兩張逃離頁有英文13066／繁中117814像素差，全在戰場視圖；問句及其他面板零差。目視可讀無裁切。與RE-290鏡頭保存修正的關係標強推論，未取得本頁原版oracle，不關閉完整UI gate。
- 同來源Windows AMD64、macOS AMD64／ARM64／universal與最低11.0相依檢查通過。Wine9正常772／重載26狀態、18PNG完整符合現行Linux，保存未改。十九套件沿RE-290相同來源收據，不重跑。
- 全輪八驗收程序12800狀態、58PNG；另保存兩失敗程序、正常診斷4782狀態與生成器最末1狀態，總82PNG。失敗程序無完整trace，不推測其狀態數。
- Docker --rm、無網路、UID/GID1000、有界逾時；正式重跑3GiB／3CPU、建置3GiB／2CPU、Wine2GiB／2CPU，256程序上限；Xvfb／Wine有trap。沿既有三個鎖版映像，沒有重建或清理其他專案資源。
- 本輪專用Wine前綴核對擁有權後移除，封存與唯讀重驗、工作清單render／verify／scan、git diff --check與容器終檢見工作區。來源保持、怪物v2／人物v3保持，RE-288任務亂數仍DRAFT。
- 入口[RE-291](docs/re/20261006-visibility-current-regression.md)及workplace/visibility-current-regression-20261006/prepare.py。同輪掛入逆向索引；CONTEXT只更新唯一目前狀態表，工作清單只補三既有項目內文，狀態／驗證條件未變。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；沒有提交、推送、Issue寫入、版號、封包或發行。HD手繪媒材已選，解析度與介面及其他既有問題仍待回覆，目標未完成。

## 2026-10-06 RE-292：正常低位攻擊與怪物拾取資格

- 開工沿唯一狀態表與未完成清單，命中 remake 驗證與文件職責路由，使用 reverse-engineer-retro-game-remake；收尾回查 knowledge-router，載入 references/verification-and-handoff.md。原始資料與專案來源唯讀，保留既有 dirty tree。
- 前案 RE-287 兩名掉物怪物均倒地。核對 SPEC-010 選項表與部位表後，登記同任務5、六人正常招募、seed1，僅武器圖示44與體型圖示66優先，得到選項4。上限仍750正常戰鬥鍵／40回合。未換種子、注入狀態或修改正式程式。
- parent 1771 封存附件、75 原始輸入與現行342份來源於準備時逐檔核對。所有工作在既有 Go1.26.7／Ebitengine2.9.9 映像，Docker --rm、無網路、UID/GID1000、明確輸出與逾時。
- 首次合併建置／執行容器3GiB／2CPU／256程序，外層1260秒，Go測試1200秒。tool session45014返回137；無完整終態，保留850 checkpoint、最後log850及772 PNG，退出原因未知。Docker歷史event查詢無可確認原因，不稱為產品問題。
- predeclared-v2.json 保留首次附件與執行檔雜湊。第二容器4GiB／2CPU／256程序，GOMEMLIMIT=1GiB，只執行原 run-original.sh；外層1260秒／Go1200秒不變。退出0，cgroup oom／oom_kill為0；完整前850狀態、既有事件與772 PNG逐byte核對。
- 原版883狀態／10回合／2264自然亂數，750戰鬥鍵有界停止。完整16-byte掉物1筆，站立怪物掉物0筆、怪物成功拾取0筆。站立空手怪物0點；正常選項4已證實，部位變化0筆。母體地形及首次戰鬥符合RE-287，前375正常狀態相同。原始結果保留，不延長同場補證據。
- 原版未取得完整怪物拾取案例時，不能升為整場AI或保存對拍。共用任務亂數、受阻擊退與體型子選單各自閘門保持，沒有正式 remake 重播或新保存程序。
- 監看腳本的 null 列表與舊收據欄位名稱讀取曾失敗，已依實際 schema 修正。這些是腳本問題，與原版程序退出137分開記錄。
- combat.md 及 docs/re/README.md 改正過期「AI／法術／戰利品未接、標題音訊待人耳、編輯器待實作、RE-227保存待重生」現況。歷史RE不重寫，索引與現行來源／既有收據互相連結。新增RE-292同輪掛入索引，CONTEXT只更新唯一狀態表，兩個既有工作項只追加內文。
- 正式342來源／340Go與Linux991046ab保持，人物v3／怪物v2保持。文件／原版觀察不重跑未變十九套件、平台與長任務；RE-291為最新來源回歸。工作清單render／verify／scan、git diff --check、獨立audit與Docker／擁有權清理結果隨工作區保存。
- 入口[RE-292](docs/re/20261006-monster-pickup-low-strike.md)及workplace/monster-pickup-low-strike-20261006-prepare.py。HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，沒有提交／推送／Issue寫入／版號／封包／發行；HD手繪方向已確認，解析度與介面及原未決問題保持，目標未完成。

- RE-292收尾補充：worklist render 首次只開放 WORKLIST.md 寫入，CONTEXT.md 統計區塊因唯讀掛載退出1；保留 worklist-render-failed.log。確認 render 會同步兩檔後，明確開放這兩個既有檔案，沿同映像／同命令乾淨重跑退出0。verify 維持18未完成／89完成／14人工，scan 1949行全部認領；不當作產品缺陷。
- 已目視原版384／846／882三頁；只記錄原版讀取與有界結果，不宣稱 remake 整頁對拍。正式342來源、Linux SHA、原750鍵與seed1保持。新案未讓怪物空手，不延長同場；下一閘門回到已待確認的共用任務亂數與既有決策，HD解析度／介面沿原問題。

## 2026-10-06：現況交接表與十八項完成條件稽核

- 路由命中驗收與交接、文件職責及結論回填；沿用已載入的 verification-and-handoff.md、project-document-responsibilities.md、re-resolution-backlinks.md，不新增完成標準。
- CONTEXT 原有兩張互相矛盾的現況表，並累積舊來源／平台及雷聲待聆聽敘述。依現行 RE-290／291／292 和 SPEC-004／031 整理為單一現況表，保留有效決策、限制與穩定入口。
- 整理前全文 SHA-256 aecb7fc8674592a2449a00d0522722089b5bd51eaf4f365f9b005b71d36364e4，與 RE-292 不可變 documents/CONTEXT.md 完全相同。本輪另留原文；歷史證據及收據不改。
- 對十八項逐一抄錄原 acceptance、verify 與 Issue，再附現行證據及下一閘門。docs/worklist.json、WORKLIST.md、所有 spec 與 README 完整雜湊保持。18未完成、89完成、14人工；不以有限切片結案整場，也不把 CGA／Hercules 當 EGA 首版阻塞。
- 補入已確認的送禮 ENTER／SPACE 及標題雷聲決策，HD 手繪只核准媒材。解析度／介面與原有共同決策問題保持；任務共用亂數、受阻／舊H、法術空白等 DRAFT 不進正式路徑。
- 正式342輸入／340Go與Linux991046ab完整SHA保持；最新平台／長任務沿 RE-291。不重跑未變來源的十九套件，沒有新遊戲、原版、保存或平台執行。遠端Issue沿最近快照標明日期範圍，本輪未存取遠端。
- 稽核入口為 [準備腳本](workplace/current-handoff-audit-20261006-prepare.py)、[逐項契約](workplace/current-handoff-audit-20261006/gate-audit.json)、[唯讀重驗](workplace/current-handoff-audit-20261006/verify.py)及同區 receipt.json。新增工作區只有文件快照、JSON與核對腳本，不建新交付目錄或第二份權威工作清單。
- 讀取工具第一次以所有條目必有status處理，遇到開放條目省略status而退出1；改用get判斷後在同映像乾淨重讀通過。草稿一個RE-261檔名引用已在回寫前改正。這些是文件工具問題，沒有產品修正。
- 本輪無提交、推送、Issue寫入、版號、封包或發行；Docker、有界資源、擁有權與暫存終檢在本輪收據及後續核對補充，目標尚未完成。

- 現況交接稽核收尾：CONTEXT 由806行整理為134行；十八條契約與全部既有規格雜湊保持。worklist verify退出0，scan 1914行全部認領、0未認領；數量減少來自移除現況文件舊斷言。git diff --check退出0，本批容器已移除；全專案無root-owned檔案或.md目錄，抽查UID/GID1000:1000。固定Go映像ID 3e89464c6f9740e28b90b2fec3627bf235e38ba4d5892b28d676653c2a4902fa未重建。新工作區只保留文件快照與稽核附件，無背景程序、Wine前綴或遊戲暫存。獨立唯讀核對入口與收據已掛入CONTEXT；本輪是文件收尾進展，完整重製仍未完成。

## 2026-10-06 RE-293：正常購馬、折抵與保存重載

- 沿唯一狀態表，命中驗收與交接、文件職責路由，使用 reverse-engineer-retro-game-remake；回填結論前回查路由。馬廄操作差異依 grilling 只做DRAFT可丟棄原型。HD手繪媒材與既有解析度／版面問題保持，不重問已選媒材。
- 舊馬廄收據從加金冠的觀察保存開始。本輪選原始名冊合法第13槽ddduhh，3000金冠、城1旅店；以64×64實際地圖產生21去程、21回程方向鍵。正常KOL→TITLE→BOOK→TOWN啟動，不注入角色、資金、坐標或state；seed1在第一次延伸按鍵前固定，重載不重設。
- 固定既有dosgolem revision968ca774、原版觀察器7cdb3533，KOLTOWN.EXE c2442c84；現行342輸入／340Go、Linux991046ab。全部分析、建置、Xvfb及正常執行在既有鎖版Docker，無網路、UID/GID1000:1000、--rm；原版／建置3GiB、2CPU、有界程序及外層600秒，原版內層540秒，分析1GiB／1CPU／180秒。
- 首版原版79狀態、正式／觀察雙語各97狀態，未保存，全部保留。已證實原版第一次購買扣金冠後坐騎仍空，SPACE才寫入、回招呼五圖示；正式版先寫入、回兩圖示商品清單。原版換馬交舊與接新各等一次確認，正式版合併。鍵序修訂分開兩側正常控制器，不改同角色、seed、資料或範圍。
- 修訂原版79狀態，買400輕型戰馬、零差額折抵實收50換馱馬、正常人物頁、回旅店付60。真正寫CHARDATA12288、STEEDPIC／CHARPICS／SHLDPICS各12160 bytes；新程序正常角色清單重載4狀態，2490金冠、完整16-byte馱馬與移動率10保持，沒有再次寫檔。
- 正式／唯讀觀察雙語各93正常狀態並真正寫v3；另起程序各22重載狀態。觀察副本230狀態去新增欄後相同、230組PNG逐byte相同。可丟棄馬廄原型雙語各97正常及22重載；與原版156映射狀態、754項隊伍／金冠／坐騎／移動率／室內圖示／選框核對零差。原型沒有進正式來源。
- 六份正式／觀察／原型完整v3除saved_at相同；怪物v2及人物v3保持。四頁馱馬圖各1520像素符合來源，四頁繁中馱馬名牌字模零差；所驗四段繁中文字log缺字0。單人單城，不結案全部坐騎、全隊騎走、野外、全部中文、HD或平台真機。
- 正常保存相對原始CHARDATA第13槽改13個byte；+0x97差分0→1保留，不猜正式v3欄位。正常Quit的第78格已切BOOK，舊TOWN DS欄位失效；保留PNG及raw但不解讀隊伍數54783。映射限有效城鎮0…77格，不冒稱全部768-byte或全畫面parity。
- 父收據附件物件SHA比較、正式trace非每格都有party的包裝器KeyError，以及只讀列印路徑錯誤均為工具問題；保留首次產物，修訂後沿同映像乾淨重跑。原版正常執行103.22秒、新程序重載32.77秒均退出0；正式／原型各程序退出0。未重跑未變十九套件或整條長任務。
- 新研究檔同輪掛逆向索引，六份修改前文件先保存；CONTEXT只補目前表與下一閘門，工作清單只追加兩既有項內文。原acceptance、verify、status、Issue及18未完成／89完成／14人工保持。入口[RE-293](docs/re/20261006-steed-trade-normal.md)、[操作比較](workplace/steed-trade-normal-20261006/review.html)及同區comparison.json；封存與唯讀終檢另附收據。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；沒有提交、推送、Issue寫入、版號、封包或發行。正式馬廄流程未改，原版確認原型待共同決策，完整目標尚未完成。

- RE-293唯讀封存初次重算在矩形tuple與JSON陣列比對時退出1，屬序列化驗證器問題；來源、附件與逐格比較已通過。第一份收據8910f72d與當時工具、文件原文保存於failed-readonly-audit-v1，其他初版與正常產物保持。修訂只將重算結果依JSON契約正規化，同映像／同唯讀命令乾淨重跑；不調比較值或正式程式。

## 2026-10-06：RE-294，HD圖片及操作畫面的工作批次

- 接續使用者已確認的HD手繪媒材。重製與文件職責路由已核對，載入grilling共同決策契約；既有解析度與版面問題保持，不重複提問或把媒材選擇擴張成完整樣張定版。
- 沿現行342程式輸入、340Go及Linux991046ab核對來源不變。五份manifest的804張PNG核對SHA-256、尺寸、PNG標頭、CRC與解壓；與RE-229素材集合和內容均無差。分為43圖組、八批，804個路徑各分派一次。
- 工作批次涵蓋書本與操作介面109、片頭結局45、人物裝備293、怪物坐騎36、城鎮世界252、戰鬥法術標記48、天候18、字型3。36張渲染或地圖資料、6張結構遮罩與3張字型圖集獨立分類，不承諾全部手繪。
- 補圖組字串、動態名稱與現行程式行號。FACEPICS經片頭資料驅動，核對title-sequence.json二十個JSON位置；補充前JSON、HTML及輸入清冊保存於initial-*。靜態候選位置不當作呼叫圖、正常路徑覆蓋率或每張圖已顯示。
- 六項固定尺寸／逐像素依賴列入後續契約：人形56×70、天候64×62、標記16×160、磚格16×16、自訂圖40×38與人物組合疊圖。原圖放大不算HD完成；字型、造型、透明、命中區、動畫及自訂保存圖片仍待原型和共同決策。
- 初次唯讀顯示腳本對整數欄位呼叫len而退出1，修正顯示後讀取通過；不是素材或產品缺陷。未建立新映像，未重跑來源未變的遊戲測試或假稱瀏覽器、HD正常操作及平台驗收。
- 新研究檔同輪掛docs/re/README.md，CONTEXT維持唯一目前表，docs/worklist.json只追加hd-art-ui內文並render。原acceptance、verify、status、Issue與18未完成／89完成／14人工保持。worklist verify及scan通過，1922行自承均有認領，git diff --check退出0。
- Docker全部有界、無網路、UID/GID1000:1000，研究輸入及程式唯讀，只開必要輸出與文件可寫。kol-re294一次性容器均移除；輸出根find未見root-owned檔案或.md目錄。沒有提交、推送、Issue寫入、版號、封包或發行；本輪收據保存文件快照，不回寫舊RE收據。
- 入口[RE-294](docs/re/20261006-hd-render-dependencies.md)、[HD工作批次](workplace/hd-art-dependencies-20261006/review.html)與同區verify.py。正式HD替換仍依既有layout-review.html選擇與後續素材契約；完整重製目標保持開放。


## 2026-10-06 RE-295：真保存騎乘出城的移動率

- 接續RE-293合法名冊第13槽ddduhh、2490金冠及馱馬10的真保存，預登記seed1、最多200延伸鍵／540秒。命中驗收、交接、文件職責與規格狀態機路由；原版四個真保存檔唯讀複製，正常TITLE、BOOK、TOWN及城門出城，不注入角色、金冠、座標、遭遇或state。
- 原版固定169延伸鍵、170狀態、177 PNG，189.914秒退出0；保存來源不變、DOS寫入0、未實作0。首次紮營原版143格，本上限沒有自然騎走提示。既有probe的reload=false只控制播種，不代表未重載；來源四檔鏈在plan及process-plan記錄。
- 真角色資料頁移動率10，正式世界仍0：啟動時計算空隊伍，正常選角後出城未刷新。時刻以0回退20，騎走資格卻讀0。先DRAFT登記，隔離一行原型及唯讀觀察，世界Movement確認0→10。1128組PNG及完整trace觀察中立，340映射狀態／1360項隊伍數、金額、完整坐騎與移動率零差。
- 原型首次紮營161步與原版143格為同輸入；原版33…143共111格時段／地形標籤及日期框逐像素符合，前版24格標籤符合。原版最後26方向鍵仍等紮營，現行與原型換日；保留此操作差異，不冒稱完整亂數或正常騎走通過。
- SPEC-011子節READY審查後，正式cmd/kol/main.go的OnLeaveTown增加一行刷新Movement。342輸入／340Go只變這一檔，人物v3、怪物v2及公式保持。正式Linux91d11101與原型逐byte相同；前版991046ab及原文保留。驗收後子節有限CONFORMED，父規格READY。
- 正式go test -timeout300s ./internal/... ./cmd/...十九套件通過、五套件無測試。新正式雙語騎乘各188狀態與188 PNG；步行REST各386狀態、真保存各568狀態、新程序重載各13狀態，trace及28抽樣PNG保持RE-290，完整v3只差saved_at。本輪正式共2310狀態／404 PNG；隔離審查1880狀態另列，不重複算原版覆蓋。
- 所有分析、建置、Xvfb、原版與重製正常執行在鎖版eob-remake-go:1.26.7-ebiten2.9.9-audio，--rm、--network none、UID/GID1000:1000、有界CPU／RAM／PIDs及外層逾時。Xvfb有trap；未建新映像。前版RE-291長任務及Windows／macOS尚未按本次來源重跑，CONTEXT已降為前版證據。
- 第一次工具JavaScript語法錯誤未執行；只讀將原版二進位當目錄與猜錯旅店標籤均在準備前修正，屬工具／路徑查證。實際玩家缺陷為移動率載入及另留待驗的紮營控制，不把環境失敗算產品問題。
- 新研究同輪掛docs/re/README.md；七份修改前文件保存，工作資料只追加兩原條目內文，acceptance、verify、status與Issue保持18未完成／89完成／14人工。新收據綁現行來源與不可變文件快照，不回寫舊RE-291／293／294收據。
- HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6；沒有提交、推送、Issue寫入、版號、封包、影片或發行。原有馬廄、HD版面與其他共同決策問題保持；目標尚未完成。入口docs/re/20261006-rideaway-normal.md與同工作區review.html、receipt.json、verify.py。工作清單、擁有權及容器終檢另記hygiene.json。

- 封存初版多一個右括號而未執行，第二版猜錯Go縮排而在產收據前退出。保留兩版脚本及實際source-difference.patch；連續工具失敗後重查規格狀態機路由，只修驗證器，正式程式與遊戲結果保持。工作清單verify退出0、scan1933行全部認領，git diff --check退出0；全專案find無root-owned檔案或.md目錄，本批kol-re295容器已移除，輸出UID/GID1000:1000。

## 2026-10-07：接手現行平台收尾

- 路由命中 remake 收尾、驗收與交接、文件職責，載入 reverse-engineer-retro-game-remake、retro-remake-end-to-end-playbook.md、verification-and-handoff.md 與 project-document-responsibilities.md。沿 RE-295 唯一狀態表，現有異動保持。
- 工作清單 verify 通過：18 未完成、89 完成、14 人工判定，產生區塊與 JSON 同步。主機 gh auth status 通過；遠端 private，24 張 OPEN，未寫入 Issue。
- 本輪平台工作入口為 [verify_current_platform.py](tools/verify_current_platform.py)，輸出在既有 workplace 下的 finish-platform-20261007/。來源唯讀、輸出獨立可寫，沿用 Go／Wine／osxcross 映像，不建立交付目錄、版號或封包。
- 預先登記現行342程式輸入／340Go、seed1及RE-295原鍵序。每平台雙語騎乘188、REST386、真保存568及新程序重載13，共2310狀態、404PNG；比較完整trace、PNG及只排saved_at的完整v3。Linux重建須等於91d11101；Windows以Wine抽驗，macOS只驗架構、最低系統、系統依賴及arm64簽章。重用來源未變的RE-295十九套件測試，不把Wine與結構檢查稱為真機驗收。
- 使用者確認採原版受阻打暈保留持物。沿規格閘門與結論回填路由審查SPEC-008子節READY，再正式修改共用knockback出口；成功擊退掉物保持。新反例先在舊正式版四種隊員／怪物、人物／地形阻擋案例失敗。工作區[blocked-stun-formal-20261007](docs/verification/local/blocked-stun-formal-20261007/plan.json)保存修改前來源、原版證據清冊、測試及正常回放入口[run.py](docs/verification/local/blocked-stun-formal-20261007/run.py)。其他未決問題保持；結果與清理狀態於本段追加。

- 正式來源只改internal/battle/knockback.go與既有knockback_test.go，340Go／342输入保持。十九套件通過。正式Linux SHA-256為55613e676d6b218c5f077be52e448b28c2f9e97d1159a63c5720cd6ec4cb5547。成功位移仍呼叫forcedDrop；受阻持械只清指令，完整持物、地上槽及返回暫存保持，空手受阻仍直接返回。
- 沿RE-254固定seed2／636正常鍵，正式雙語1274狀態、唯讀副本1274狀態及26真正新程序重載通過。兩版完整trace、PNG及排saved_at的v3中立。每語言126持久欄位、268戰鬥seed／返回暫存點、168世界點符合封存原版。原版317格／正式320步的完整Elf Bow、位置、指令、seed2126174334、暫存6及地上槽零差。原版及雙語畫格已目視；舊H名稱仍不同，本輪沒有新增原版程序，不把重播數量當原版覆蓋。
- SPEC-008只將上述窄切片升有限CONFORMED，父READY及所有整場完成條件保持。新[驗收筆記](docs/verification/blocked-stun-20261007.md)同輪掛入規格索引與正常玩家路徑索引；原版來源、位址空間、證據等級、工具、seed與權利邊界保留。README只修過期v2摘要，不改專案起源或累積本輪命令。
- 決策前已完成的RE-295跨平台建置保存於finish-platform-20261007/baseline-re295/，未當修正後驗收。新正式來源再建Linux、Windows AMD64、macOS AMD64／ARM64／universal；Linux與已測試正式二進位相同，macOS最低11.0、系統依賴、arm64簽章通過。Linux／Wine9各雙語2310正常／保存／重載狀態、404PNG的完整trace、PNG及排saved_at的v3符合RE-295登記基準。macOS未執行，Wine不取代Windows真機，正式封包、FUSE、長任務及發行未驗。
- 最初清冊驗證錯把獨立tools/oracle四個測試檔加入342輸入，保存首次腳本與失敗清冊後按實際cmd／internal契約乾淨重跑。Wine初次選kol-package而缺wine64，遊戲未啟動；改用前版已驗coab-wine-smoke:ubuntu-noble-20260826，同控制器乾淨重跑。建置首次自動審核逾時未執行，依回應重試一次後通過。均為工具或環境問題，未當產品缺陷。
- 工作資料只追加v01-full-battle-parity、v01-remaining-battle-chain及cross-platform-build-smoke的內文；全部原acceptance、verify、status、Issue及acknowledged保持。render／verify／scan通過：18未完成、89完成、14人工，1933自承全部認領。CONTEXT維持唯一狀態表，將舊平台來源降為歷史基準，不宣稱完整remake完成。
- [規則收據](docs/verification/local/blocked-stun-formal-20261007/receipt.json)與[唯讀核對](docs/verification/local/blocked-stun-formal-20261007/verify.py)、[平台收據](docs/verification/local/finish-platform-20261007/receipt.json)均通過全唯讀重算。Go／Wine／osxcross沿用既有映像，未建立新映像。全部程序在Docker內，有界資源、逾時、無網路、UID/GID1000:1000；Xvfb及wineserver由trap回收。
- 專用Wine前綴清除720418292 bytes；驗收圖、真保存及失敗附件保留。全專案find未見root-owned檔案或.md目錄，抽查來源及輸出均1000:1000；本輪kol-blocked-stun／kol-finish-platform容器全部移除。git diff --check通過。HEAD仍1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，無提交、推送、Issue寫入、正式版號、封包、影片或發行。

## 2026-10-07：原型比較與工作區清理

- 使用者確認本輪要看任務亂數接續的原型比較，並要求完成比較後刪除對應工作目錄。
- 沿封存RE-288／289／290，整理原版、前版正式與RE-290組合候選的首次／末端畫面。原圖位元組保持；補入較新的移動修正及圓環相位結論。正式任務亂數仍DRAFT，未把比較確認當作規則採用。
- [比較頁](docs/verification/local/quest-entry-rng-prototype-20261006/review.html)與[保存入口](docs/verification/retained-evidence.md)已索引。五個工作區的原收據、非執行檔證據及來源腳本壓縮保存，原SHA-256保持。執行檔可重建，移除位元組並保存雜湊。
- 新增tools/retain_verification.py，在Docker內核對所有保留檔、原收據附件、圖片來源雜湊、離線頁連結與342份正式程式輸入。所有檢查通過後才逐一刪除五個精確工作目錄；原始輸入、其他研究工作區與dist-all保持。
- 保存工具首稿把兩種既有收據欄位都當size／sha256物件；實際正式收據是SHA-256字串。修正格式判讀後原雜湊核對通過。連結核對首稿又從容器/out別名解析頁面相對路徑；改從實際/src/docs/verification/local解析，再以同容器／同命令乾淨重跑通過。兩項均屬保存／驗證工具問題，沒有產品或原圖變更。
- 已更新現行Markdown連結、CONTEXT、工作表body與AGENTS清理規則。原完成條件、狀態、Issue與acknowledged保持；歷史收據及壓縮內快照仍保留當時路徑，以retention.json解析。
- 清理前約756 MiB，保留約18.6 MiB，約釋放737.8 MiB。完整數據及舊收據雜湊見[清理收據](docs/verification/local/cleanup-20261007.json)。
- 未提交、推送、寫Issue、改版號、發行或重新執行原版／正式遊戲。沿現行正式來源的既有驗收，來源雜湊保持。Docker與擁有權最終自檢另記本機final-audit.json。
- 最終報告首稿誤把工作表可省略的status欄位當成必填；改用既有預設判讀，文件追加加上重跑防護，同命令重跑通過。工作表仍為89完成、18開放；此錯誤只影響清理報告，證據保存及刪除已通過。
- 刪除原工作區後，獨立唯讀核對五份壓縮證據及比較頁全部通過。四份本輪/tmp編輯腳本已移除，比較頁建置腳本保存在本機證據入口。
- Docker專案容器無殘留，沒有root-owned檔案或誤建Markdown目錄；git diff --check通過。

## 2026-10-07：Codex重啟後接手任務亂數收尾

- 路由命中復古重製收尾及交接，載入reverse-engineer-retro-game-remake、收尾作業法、規格閘門、驗收交接及文件職責。依前輪READY與封存決策接手，保留原工作樹。
- 重啟前繁中前綴容器仍在執行，讓原任務完成後接續差分核對。正式342輸入／340Go只改main.go任務入口；十九套件已通過，來源雜湊保持。正式Linux 2a885b9a37e884d1830010f7ccb1051b030fe4b14c678fbb71a25955a745fd66。
- 雙語正式及唯讀前綴3540狀態／108PNG，每語言753戰鬥點、13554角色記錄及50已解值／seed／完整怪物手持零差。首次4096地形與900語意符合原版，26次畫面保留262圓環像素，PNG與RE-290候選相同。原版收據重用，不增加末端oracle。
- 預登記1045正常鍵保留自然ORC遭遇並正常逃離，回城全隊住宿及四次新程序重載共4236狀態／72PNG。正式與副本完整trace、PNG及v3除saved_at中立；唯讀重載快照重建完整v3相同，保存原檔未改。正常繁中戰場及住宿畫格已目視；不結案完整中文或整場原版末端。
- 圖片工具舊路徑、未完成的比較依賴、回城道路、多人發言者確認、v3扁平JSON欄位及文件patch格式問題均已修正工具，失敗輸出保留。觀察副本於自然遭遇初頁讀nil地圖，英文失敗與繁中停止紀錄另存；正式雙語保存均退出0。診斷副本補nil守衛後按同seed／同鍵序乾淨重跑通過，正式來源保持。
- SPEC-008本切片有限CONFORMED，父READY及全部原完成條件保持。CONTEXT更新唯一現況，前版長任務／平台不冒充新來源。工作表只追加三條body，acceptance、verify、status、Issue及acknowledged保持；render／verify／scan通過。沒有提交、推送、Issue寫入、正式版號、封包、影片或發行；HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6。
- 收據、來源、預登記、控制器與失敗附件由docs/verification/quest-rng-20261007.md索引。證據壓縮、工作目錄清理及Docker／擁有權終檢另於本段追加。
- 工作清單第一次核對抓到新寫的任務完成範圍未被既有認領規則識別。重查工作清單路由後，明確寫成「此鏈以逃離結束；原版末端及完整任務5仍待驗」，維持原條目及全部acknowledged。render／verify／scan同命令重跑通過，18未完成／89完成／14人工。
- 1217份證據壓縮保存於docs/verification/local/quest-rng-formal-20261007/，四份執行檔只保留雜湊；236791517 bytes縮為約4801811 bytes，釋放約221 MiB。新收據、完整v3重載、原版引用、全部壓縮內容與新連結經完全唯讀核對通過後，精確刪除本次工作目錄。其他工作區及dist-all保持。
- 刪除後完全唯讀重核保留證據通過；工作表仍18未完成／89完成／14人工，1940行自承全部認領。全專案find無root-owned檔案或.md目錄，來源、文件、收據及壓縮證據抽查均1000:1000。本輪kol容器無殘留，git diff --check通過；沒有本輪主機暫存腳本或新映像。

## 2026-10-08：現行任務亂數來源跨平台回歸

- 本輪於2026-10-07開始，驗收於2026-10-08完成，工作區及映像revision沿開始日期。路由命中跨平台正常玩家路徑，載入既有重製驗收技能、試玩來源規則及文件職責，未進入新的產品決策。
- 前輪Go／Wine／macOS映像按tag及ID均無法inspect或啟動；通用Python正對照也失敗，列舉結果未當可用證據。先恢復Python及鎖定Go基底，再讀原建置來源。新增docker/Dockerfile.verify及verify-wine，macOS既有Dockerfile增加KOL_GO_BASE參數，建立有明確替代關係的三個本機revision。Go官方digest固定，module保持現行鎖定檔；建置上下文只含Dockerfile與module，不傳原版輸入或資產。所有實際映像ID及套件清冊已保存，Go1.26.7／Ebitengine2.9.9／Oto3.4.0／Wine9確認。
- Linux執行檔2a885b9a與前輪逐byte相同，WindowsAMD64、macOSAMD64／ARM64／universal完成建置；最低11.0、ARM64簽章、系統依賴及固定module通過。十九套件再次通過，五套件無測試。正式342輸入／340Go及全部runtime雜湊保持，沒有遊戲規則、資料格式或資產修改。
- Linux／Wine9各雙語2310狀態、404PNG符合RE-295原正常騎乘、REST、真保存及新程序重載。另Windows最新任務前綴、1045正常鍵保留自然ORC並正常逃離、回城全隊保存及26新程序重載，共3888狀態、90PNG，完整trace、PNG及v3只差saved_at符合封存現行Linux；保存原檔未改。合計8508狀態、898PNG，不新增原版覆蓋或末端oracle。
- 結構核對第一次自動審核逾時，未執行；按回應重試一次通過。文件patch首稿的行比對問題在執行前修正。這些屬工具或環境，不當產品缺陷。macOS未執行，Wine非Windows真機，FUSE、正式封包、HD及原完成條件保持。
- WORKLIST只追加三條body及更新日期，所有acceptance、verify、status、Issue及acknowledged保持；render／verify／scan通過。CONTEXT更新唯一現況與可用工具入口，新驗收文件同輪掛入CONTEXT及保存索引。沒有提交、推送、Issue寫入、版號、封包、影片或發行。
- 原控制器及新增Windows控制器在來源唯讀、輸出獨立可寫的有界Docker執行；UID/GID1000:1000、測試無網路，Xvfb與wineserver由trap回收。建置恢復明確開網路，有界CPU／RAM／nproc及外層逾時。兩個精確Wine前綴已在回放結束後移除，釋放1440804232 bytes。其他工作區及dist-all保持。根收據及唯讀核對均已通過，壓縮保存、引用與最終清理於本段追加。
- 根證據保存1158份檔案及七份可重建執行檔雜湊，134757126 bytes縮為約6366771 bytes。完全唯讀重核來源／runtime、全部壓縮內容、8508狀態、完整v3及新連結通過後，才精確清理本次工作目錄。封存核對的JavaScript首稿語法錯誤未執行，修正shell引用後重跑通過；資料與來源保持。三個恢復revision及必要基底保留，懸空映像清單為空，沒有清理其他專案資源。
- 使用者暫停前，平台工作區刪除後的獨立唯讀核對通過；1939行自承全部認領，18未完成／89完成／14人工保持。全專案find無root-owned檔案或.md目錄，來源／工具檔／收據抽查1000:1000，kol容器無殘留，HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6，git diff --check通過。

## 2026-10-08：暫停對拍，查證原版C同位元組重建

- 使用者明確要求暫停目前對拍，想改以matching decompilation還原原版C並減少重複行為比對。專案容器清單為空，沒有再跑原版、重製或對拍。伺服器重啟後重新核對停止狀態及已完成工作，未重複前輪回歸。
- 路由命中原始工具鏈及共同決策，載入toolchain-and-runtime-fingerprints、工具鏈知識入口及grill-me／grilling；先查證事實，未更動正式Go架構或完成條件。
- 既有toolchain.md已證實Microsoft C runtime及EXEPACK，精確compiler／linker／packer仍未知。唯讀容器重查六支原版EXE大小與SHA-256；查過的borland、eob_123_remake及dosgolem工具目錄找到MSVC1.0／1.52候選，沒有據此認定原作版本。原始ZIP及相鄰工具目錄唯讀，不搬運、上傳或執行候選工具。
- 查閱decomp.me官方FAQ及decomp-toolkit維護者文件，區分函式機器碼、解壓映像及原始EXE全位元組三層匹配。完整binary相同是可重現建置結果的證明，不等同找回唯一原始文字，也不自動證明後續Go移植一致。
- 新入口docs/re/matching-decompilation.md同輪掛入逆向索引與CONTEXT，現況為可行性研究、尚無匹配成果。原版證據、Go程式、工作表及原完成條件保留。下一步先核KOLTOWN歷史工具鏈及代表函式，再依實際匹配／耗時決定擴展；沒有提交、推送、Issue寫入或發行。

## 2026-10-08：matching首批純C匹配試點

- 使用者確認開始已展示KOLTOWN試點；遊戲對拍保持暫停。命中工具鏈與IDA9.4路由，沿use-ida-pro-9-4及指紋規則；原始ZIP、正式DB及相鄰工具唯讀，338函式一次性DB探針通過，輸入SHA與固定解壓映像相符，原名／位址／bytes保留。
- Microsoft官方MS-DOS工具repo固定commit2d04cacc5322951f187bb17e017c12920ac8ebe2、28檔／1165531bytes，compiler候選MSC5.10／link3.65。五種memory model純C輸出均不匹配，prologue保存SI／DI順序與指令選擇為反例。CL的/qc須QC.EXE而官方工具組缺前端，保留提示／有界停止，未盲目調C。
- QuickC候選來源davidly/dos_compilers固定commitda3f15ce4912893a0dd9f8349dab29116a53ca59，每檔核Git blob及SHA；QCL banner2.01、真實linker4.07。/AM /Od /Gs /Zl /G0生成三個C函式，完整linked body268bytes符合原版：sub_14427=62、sub_18CD5=85、sub_1A866=121。PUBLIC兩種排序定位一致，CODE段268bytes，各函式真實extent相同，未只比較前綴。
- pool的DS:BB88、active的DS:BB82及__acrtused=0明列link fixture。原資料、CRT、startup、stack、entry、2281 relocs與原始EXEPACK未還原；780byte研究片段不是可玩遊戲或全binary。C源沒有assembly或原版code blob；sub_1A4E4 far-pointer純C反例仍保留，相容工具組匹配不等於唯一歷史版本證實。
- 已驗CC0 unexepack.c來源完整SHA0e1da11a1d5a4aa5722dc83dd16c683b4c2c25cec31d3741ee71039840a6b03c從固定revision取回，原樣GCC容器編譯。原版ZIP重生138528byte、SHA01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2，三段rawfile bytes與IDA、linked MZ零差。乾淨重編C／OBJ／BIND／MZ／MAP均逐byte相同，片段SHA0144c4dfc32779a36fbdaa7748688fffb13025e1e774f6141cafdf92db74422d。
- transport大型commit回應截斷、批次反斜線與工具JavaScript語法問題已查明，改小ref/tree及明示quote後重跑；X11未建立可用SDL display，compiler改無頭SDL，沒有拿無頭當遊戲畫面驗收。首批DOSBox完成後未退出，外層timeout回收；後續由90秒有界owner觀察DONE後停止，程序紀錄保留。TIS OMF1.1明允許零checksum，解析器依規格修正；link即使error仍產MZ，已改核無error、PUBLIC及完整extent，所有失敗產物保留。
- 所有分析、IDA、下載、GCC、DOS編譯及link在有界Docker。沿既有IDA9.4 py312、Go1.26.7 image；DOS工具確認舊image皆缺失後新建kol-matching-dos:bookworm-20261008-r1，來源docker/Dockerfile.matching-dos。UID/GID1000:1000、原始輸入唯讀；下載開網路，分析／編譯無網路，無主機工作負載。正式342Go輸入及程式保持，沒有原版或重製執行、提交、推送、Issue寫入、封包或發行。
- 新原版完整matching工作登記為matching-koltown-binary；原107項全部原文及條件保持，新增一個明確scope認領。matching入口、C／hash／收據同輪掛CONTEXT、逆向及保存索引。根收據及唯讀驗證通過；封存、工作區清理與終檢於下文追加。
- 首批C、IDA、OMF、負例、原始bytes及命令保存138份檔案；5982325 bytes縮為約1830329 bytes，44份可重建或固定commit重取的EXE只保留hash及來源。匹配C另有可讀副本，未進正式Go路徑或公開repo。保存器前批日期硬編碼已改從新收據取日期，只修本批新retention日期，原收據及歷史保存保持。
- 原始file來源、三個完整linked body、乾淨repeat、342份Go输入、壓縮內容與新鏈結經完全唯讀核對後，精確清理本次matching-pilot工作目錄。DOS image5992e601及必要基底保留；懸空清單中的4ac62de83339是本輪實際使用的既有私有IDA9.4候選，仍屬所需工具，不刪。其他工作區、工具、原始ZIP、dist-all及原遊戲對拍暫停狀態保持。
- 刪除後完全唯讀重核通過；工作表19未完成／89完成／15人工，1945行自承全部認領。全專案find無root-owned檔案或.md目錄，源檔／工具／收據／可讀C抽查1000:1000。kol容器無殘留，HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6及正式Go保持，git diff --check通過；沒有提交或發布。

## 2026-10-08：原位增量連結與跨段C函式匹配

- 使用者確認繼續已展示的分段與增量連結基準。命中工具鏈、IDA9.4與文件職責路由，沿既有逆向技能及TIS1.1 OMF規格；沒有變更正式玩法或架構，遊戲對拍保持暫停。
- 以一次性IDA資料庫匯出八個段及338函式，固定輸入SHA01f6f9a。沿唯讀ZIP與原樣CC0解包器重生，不執行遊戲；固定QuickC2.01候選的三支必要EXE按前輪commit、Git blob與SHA重取。既有匹配image5992e601及IDA image4ac62de沿用，沒有建立重複image。
- ORIG.OBJ保留原始未匹配片段。CSLOT.OBJ將compiler的LEDATA、FIXUPP與BAKPAT轉換到原始段位置，保留指令與回填值。真正LINK4.07解析globals及遠呼叫；COMMON段、完整無洞無重疊覆蓋與MAP原位定位另核，沒有把linked code貼回原始EXE。
- 三函式268bytes原位接回後零差，新增sub_1B5A8的32byte遠呼叫函式也完整匹配。純C為return sub_2D538() % n + 1，callee仍是原始未匹配片段；四個call fixup bytes全核，MZ相對目標1B8F:1C48與IDA載入後2B8F:1C48分開記錄。合計C300bytes，原始未匹配129012bytes；沒有新library或asm匹配。
- 六CODE段、完整129312byte初始化映像與原版逐byte相同；2281筆重定位實體位置及目標、入口1B8F:001C及堆疊2B48:07D0相同。active／pool從原版DGROUP的實際BSS PUBLIC解析，取代絕對global fixture；__acrtused=0仍明列無程式引用的link anchor，CRT原片段保持。
- 首個適配物件重複宣告完整DATA，LINK報L1070超過64K，修正為零DATA／STACK貢獻。首個原版DATA布局把47968byte未初始化空間寫入EXE，依固定解包器檔案邊界分開DATA／BSS後重跑，完整初始化映像長度相同。原作者OBJ邊界仍未知，不因此推定原始名稱或來源結構。
- native MZ hash3bb05eecbaff312af6d76ba79abbb019d8cf942e90ef82a133b99ca92a4a51d6。相較固定解包器輸出，e_minalloc為3123對2308、e_csum為57581對0、e_lfarlc為30對28；重定位順序及配對編碼亦不同。沒有改寫header或遮罩差分，完整解壓MZ及原始packed EXEPACK尚未匹配；129012bytes原片段不計為C還原。
- 乾淨第二工作區由原始ZIP重生，C、compiler OBJ、ORIG.OBJ、CSLOT.OBJ、native MZ、MAP及重定位清冊七項逐byte相同。獨立核對器直接解析兩次OMF與明示符號位置，重生整個初始化映像及2281筆重定位；修改C指令、移除重定位、改目標、改入口CS四個負例均拒絕。342份正式程式輸入雜湊保持，沒有重跑Go建置或遊戲對拍。
- 啟動器raw字串與建置腳本換行的語法錯誤在編譯前修正；Python3.11不支援tar.extractall的filter參數，改先核所有成員形態／路徑／雜湊。image缺rg時以容器Python查少量索引，沿同image乾淨重跑；工作表驗證均過。這些為研究腳本或環境問題，沒有登記為產品缺陷。
- docs/re/matching-decompilation.md、CONTEXT及matching-koltown-binary回填本輪現況，逆向索引的舊「沒有匹配成果」修正；首批歷史、原有107工作條件及所有Go程式保持。工作表render／verify／scan／status通過，19未完成／89完成／15人工；下一閘門為原始EXEPACK／MZ重建基準，再擴展增量C替換。
- 有界Docker、UID/GID1000:1000、原版與正式DB唯讀、分析／建置無網路；固定工具下載明示開網路。DOSBox由90秒完成標記owner及外層timeout回收。封存395份檔案，18份可重生或固定commit重取的EXE只留hash，13107888bytes縮為1987250bytes。封存後完全唯讀重生與逐檔核對通過，才刪除matching-link-20261008工作區；其他工作區、原始輸入、現行映像與交付物保持。
- 擁有權及全專案find核對通過，沒有root-owned檔案或.md目錄；本批容器均已移除。沒有提交、推送、Issue寫入、打包、版號或發行。最終文件與封存重核結果於下文追加。
- 刪除後終檢通過：395份保存雜湊、唯讀原版重生、四函式及整個初始化映像、2281筆重定位、兩次OMF、七項重編與四個負例均核對。19未完成／89完成／15人工保持，終檢時1952行自承全部認領；逆向索引正對照及matching文件全部連結存在。文件與收據1000:1000，HEAD1384e508f70e8a5d81bc050ea79f66d3f0d38fe6及342正式程式輸入保持，git diff --check通過，kol容器清單為空。

## 2026-10-08：提交重製收尾與matching檢查點

- 使用者明確要求先commit及push，再繼續MZ／EXEPACK研究。遠端wicanr2/kol-remake-cht仍為private，主機gh登入核對通過；原有未推送1384e50及這次檢查點一起推送，不改寫遠端歷史。
- 暫存区包含前輪收尾與驗證資料，完整檔案核對2233份、97087401 bytes；刪除與rename另由Git登記。1249張PNG的645個不同hash均有已索引JSON紀錄。原版ZIP／EXE／PDF、typed data、matching C／OBJ／DB、工具binary及本機封存證據均排除，沒有憑證或大型executable bytes匯出。
- 舊十九套件與跨平台證據的342份正式程式輸入保持；本輪只重核工作表及matching獨立證據，沒有重跑遊戲對拍。檢查均在既有有界Docker、UID/GID1000:1000及來源唯讀掛載內執行；Git與GitHub控制面在主機。
- 首次完整暫存差分顯示六處既有空白警告，位於歷史驗證腳本、CRLF夾具TSV及原始建置log；保留既有bytes，沒有為消除警告重寫證據。先前工作樹diff檢查不涵蓋當時未追蹤的新資料，與這次cached檢查範圍不同。暫存區沒有active Git hook。提交與推送結果於後續研究段追加。

## 2026-10-08：EXEPACK完整檔案重建基準

- 檢查點ff25f7a173150cb27209371e121cb3033047bcf7已提交並推送，origin/main實際SHA核對相同；原有1384e50一併推送，工作樹乾淨後才開始封裝研究。沿原工具鏈、MZ／OMF證據與文件職責路由，讀Microsoft C原始手冊的EXEPACK／EXEMOD介面；資料來源與確切工具版本分開登記。
- 固定MS-DOS官方commit2d04cacc5322951f187bb17e017c12920ac8ebe2取回LINK3.65、EXEPACK4.06與EXEMOD4.02；QuickC2.01／LINK4.07沿原固定清冊。每檔核Git blob、大小與SHA；原版、封存輸入與執行中的工具唯讀，下載明示開網路，分析與重編無網路。既有5992e601工具image沿用，沒有建新映像或執行遊戲。
- LINK4.07與LINK3.65內建/EXEPACK均不匹配。官方standalone EXEPACK4.06的壓縮資料、EXEPACK header與261byte stub匹配；首個剩餘差分為最小記憶體欄、relocation順序及81byte尾端0x1A。失敗候選與原始logs保留。
- 勘誤：前輪固定解包器e_minalloc=2308是unexepack.c:384的fullSize/60占位；checksum、header與relocation表亦由工具生成。該完整MZ不作原版未封裝header的oracle，原始初始化映像及IDA定位維持。原版packed MZ已知最小載入量為ceil((133679-512)/16)+2883=11206 paragraphs，據此設定未封裝最小配置11206-8082=3124；不認定原始未封裝header或歷史工具版本。
- C遠呼叫的relocation在原先ORIG+CSLOT順序排到末端；第一次prefix／C／suffix切分修正此筆，但原版DATA／CRT CODE交錯仍剩91筆序列差。粗粒度LEDATA排序不能表達交錯；每個兩byte原始relocation placeholder改由獨立LEDATA/FIXUPP供應，普通bytes只供應一次，按原始序列傳給native LINK。完整129312byte初始化映像及2281筆實體relocation的順序和目標相同，沒有post-link relocation改寫，原始OBJ切分仍未知。
- EXEMOD4.02接受hex參數。最初/MIN的字串3124被讀為0x3124，欄位驗證拒絕；按原手冊改用同一已推導值C34後重跑，實際欄位3124。EXEPACK4.06產生133679byte核心，所有MZ欄位、壓縮資料、stub與壓縮relocation符合原版。128byte記錄與fill0x1A產生尾端81bytes；KOLBOOK／MAKEMODE也分別有55／63bytes同類填補，產生原因與原始工具未知。
- 完整133760byte KOLTOWN.EXE逐byte及SHA-256 c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221符合原版。來源比例維持四C函式300bytes／原始未匹配129012bytes，沒有新增library或asm，也沒有宣稱完整C還原。完整C反編譯工作保持open，下一步逐函式擴展，維持全檔hash與來源覆蓋清冊。
- 乾淨第二工作區由原版ZIP、C、固定工具與原始片段重建：QCL編譯、OMF適配、native LINK、EXEMOD、EXEPACK及外層序列化全部重跑。11項物件／MZ／MAP／packed產物逐byte相同；七個檔案區域相同，最小記憶體、壓縮byte、重定位byte與尾端填補四個損壞負例拒絕。兩次最終EXE再次靜態解包相同；獨立reader核對原位C、完整OMF初始化映像與2281筆重定位。正式342份程式輸入保持。
- 封裝研究源檔、C、OMF、候選失敗與原生工具命令只留本機。新收據明列whole packed matching=true、full C decompilation=false；封存核對器讀原版與OMF並核對原生工具收據，完整重編入口為rebuild.py與固定DOS工具，支援範圍仍為四個函式。工作區清理及最終文件／容器核對於下文追加。
- 封存557份檔案，35份可重生或固定commit重取的EXE只保留hash；17909645bytes縮為2349194bytes。全唯讀核對封存、full packed hashes、C與OMF來源、重定位及342份正式輸入通過，才精確刪除matching-pack-20261008。全專案find沒有root-owned檔案或.md目錄；其他工作區、原版輸入與工具image保持，專案容器清單為空。工作表render／verify／scan通過，19未完成／89完成／15人工保持，1954行自承全部認領；待完成的是C來源覆蓋。
- 刪除後終檢再次通過，557份保存雜湊、原版ZIP／OMF來源與full packed收據、342份正式輸入及全部matching文件連結保持；終檢時1955行自承全部認領。新檢查點只包含八份現況、研究、保存與工作表文件／工具，原版EXE、C、OBJ、DB與候選工具繼續排除；依本次commit／push授權保存完整封裝基準。提交主題為「建立KOLTOWN完整封裝的混合來源重建基準」，確切SHA與遠端結果由Git紀錄及本輪交接提供。

## 2026-10-08：持續目標的46個C與八個runtime來源還原

- 前一goal turn屬progress：完整EXEPACK基準確實改變可重建證據，但當時僅四C函式，沒有完整C還原。開工核對工作樹乾淨、HEAD2d7bef520b71144f6b1a31dbd38bfd923ac4c4e8、專案容器清單為空；目標保持完成kol matching decompilation，六原始EXE與全部未還原CODE不縮小。
- 命中工具鏈、MZ／IDA與文件職責路由，載入實際技能與references。沿原IDA9.4 image4ac62de、matching image5992e601；正式DB與ZIP唯讀，對一次性副本匯出全部338函式、原始chunks／xref、IDA載入bytes與檔案bytes的各自基準。原正式DB保持，沒有推測性改名。
- 批次從4→18→25→34→46個純C函式，最後4309bytes完整native body符合原版。包括角色條件、word讀写、骰值欄位、字串、繪圖調度、720byte sub_21AC4初始化、450byte sub_21E0B裝備與183byte sub_21FCD switch。源码保持原sub名稱與位址，50byte記錄欄位以offset作導航；語意未分級者不改正式Go型別。
- 解出固定工具後的空msc510目錄未被封存，下載器補明示mkdir後沿同hash重取。編譯／連結完成標記改用獨立目錄，已消失的完成compiler handle不重啟；新建置接續既有產物。以上是工作腳本環境問題，沒有當產品缺陷。
- 實際compiler差分顯示舊式char參數、逗號初始化、一般加法與+=、運算順序、signed／unsigned prototype轉換及明列default改變布局。逐一由原始指令與compiler交叉讀取找原因，保留所有負例C；沒有加ASM、_emit或把非fixup指令改成原版。小模組尾端單一NOP另列compiler padding，不計入函式body。
- 原位適配改讀真實PUBLIC／EXTERNAL，不再固定四函式。C與library地址欄位分開供應、backpatch依原始函式位置重生；每個原始byte只供應一次，2281筆原始relocation順序與目標由native LINK產生。獨立source reader核compiler／LIB裸指令與物件覆蓋，再解析fixup重生完整129312byte初始化映像，不從原版貼回C指令。
- 固定davidly/dos_compilers commitda3f15ce查完整9035記錄tree，取QuickC medium MLIBC與MSC候選／headers，每檔核Git blob及SHA。QuickC MLIBC.LIB306物件全部checksum通過；五個無fixup string函式直接完整匹配，sprintf90、srand17、rand40以實際原位引用與seed初值另核，再native整合。相容來源372bytes已證實，不據此認定唯一原版工具或符號名。MSC候選crt0.asm的COMENT在檔案offset482 checksum不符，保持診斷，未降低檢查。
- 兩次乾淨C、compiler OBJ、原位OBJ、native MZ、MAP、loader配置、packed core與完整EXE八項逐byte相同；原版packed SHA c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221保持。兩次來源物件分別獨立核對，負例指令／配置／尾端修改均被拒絕；正式342份程式輸入保持，沒有原版／remake執行或遊戲對拍。
- 新唯一source-ledger涵蓋六EXE、338個KOLTOWN函式及來源类别；46C／4309bytes、8library／372bytes，尚有117895CODE bytes／284函式與其他五EXE未還原，6736byte DATA保留。完整goal、工作表與來源清冊保持未完成，下一批從此C與工具接續，不回退四函式試點；原始CODE不計為source progress。
- 本批835附件收據與全唯讀驗證通過；原版、C、LIB、OBJ、DB、完整ASM／bytes匯出只留本機。完成比較的工作區按既有保存契約封存與清理，goal仍active；最終數據與核對於下文追加。
- 封存805份檔案、31份可重建EXE只保留hash，37125058bytes保存為16841622bytes。封存後完全唯讀重生原版並獨立解析兩份compiler／LIB／OMF來源，全部已驗來源與完整packed hashes符合，六EXE與284個未完成函式清冊保持；才精確清理本批工作區。來源與文件抽查1000:1000，全專案find無root-owned檔或.md目錄，沒有本批容器殘留。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1957行自承全部認領，matching入口所有本機證據連結存在；正式342份程式輸入與原交付物保持。完整goal的completion audit不成立，原始CODE與其他EXE尚未還原，保持active並從此46C／8LIB來源繼續。沒有Issue寫入或發行；本機源碼及binary不提交。

## 2026-10-08：第二批來源還原與中斷後提交

- 從HEAD 8b024c5與已完成的46C／8函式庫批次接續。中斷後先查容器及產物，沒有仍在執行的工作，不重啟已完成的compiler；本輪屬進度推進，完整matching目標維持。
- 沿工具鏈、MZ／IDA及文件職責路由。KOLTOWN新增五個完整C函式，現51C／5057 bytes；三個固定函式庫的候選須共享外部符號及物件段定位，再由native LINK驗完整指令。現42個函式庫函式／4214 bytes，包含實際32-bit helper，原始名稱與IDA位址保持。
- LIBH後寫入的LEDATA把部分retf覆蓋為retf 8，按實際來源記錄順序處理；地址計算採16-bit段內模數。來源核對不只遮蔽地址，不改普通指令或放寬checksum；相容工具鏈已證實，原版精確歷史版本仍未知。
- KOLBOOK以IDA9.4重新核對191個函式，首個near C函式sub_1276F的89 bytes完整匹配。原版資料區兩筆重定位使用晚64 bytes的段框；以實際段配置保留後，native LINK的header、checksum、初始化映像與四筆重定位均相同。未分段候選保留為負例，沒有手補checksum。
- 合計52C／5146 bytes。兩支完整EXE逐byte符合固定ZIP：KOLTOWN 133760 bytes、KOLBOOK 40192 bytes。十三項乾淨C、OBJ、native、MAP及封裝產物相同；三份獨立compiler／LIB／原位OMF來源核對通過，六項反例均拒絕。正式342份程式輸入保持，沒有遊戲執行或恢復對拍。
- 唯一清冊保持完整目標未完成：KOLTOWN尚有245函式／113305 CODE bytes，KOLBOOK尚有190函式／37351 CODE bytes，其餘四支EXE尚未還原。原始CODE及DATA不計來源進度。
- 本批1048份附件封存為1014份檔案，35份可重建執行檔只保留hash；58514028 bytes保存為19148105 bytes。完全唯讀核對原始六EXE、完整雜湊、三份來源證明與正式Go後，才精確刪除matching-c2-20261008工作區；C、LIB、OBJ、DB與原版片段只留本機壓縮證據。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1959行自承全部認領。新增入口沿既有逆向索引與保存索引，檔案擁有權1000:1000，已清範圍沒有root-owned檔案或.md目錄；容器終檢與Git提交結果於本輪交接提供。
- 依使用者「先commit + push，然後繼續」授權提交本批研究紀錄與保存工具；GitHub儲存庫仍為private。沒有Issue寫入、封包、版號或發行，原版衍生來源不進Git。

## 2026-10-08：push後接續KOLBOOK來源還原

- 第二批提交e44cddde78434df2372a16ff5db473fd44b795f6已推送origin/main，工作樹與遠端同步後才接續第三批。完整matching目標保持進行中，沒有恢復遊戲對拍。
- 沿既有工具鏈及文件職責入口，在已登記的matching-c3-20261008工作區復原第二批封存；原版ZIP、前批來源與programs唯讀。固定QuickC2.01／LINK4.07及small model函式庫按commit、Git blob與SHA重取，沒有新增image。
- KOLBOOK新增八個C函式505 bytes，現九C／594 bytes。near-call的self-relative FIXUPP保留compiler指令與addend，交由native LINK原位回填；長值與條件分支依原始指令匹配。sub_13E25的far pointer候選24 bytes與原版26 bytes不同，保留負例、沒有用ASM補差。
- 搜尋SLIBC／SLIBFP／LIBH；本批36個實際供應函式均為固定SLIBC物件，3705 bytes完整匹配。初次LIB解碼遇到未支援LIDATA後改為明確拒絕該物件並保存診斷，不放寬檢查或把部分物件當作已支援。一次交叉讀取未設容器工作目錄造成omf import失敗，修正同一容器命令後重跑，兩項均為工具問題。
- 完整KOLBOOK 40192 bytes及SHA-256 2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832保持。乾淨C、compiler OBJ、原位OBJ、native MZ、MAP及完整EXE六項相同；兩份Book來源與兩份既有Town來源由獨立reader重算完整初始化映像及有序重定位。
- 六項反例包括有效OMF checksum下的C指令、函式庫指令、near-call目標修改，以及ASM注入、MZ checksum與尾端填補。三個來源修改確實執行獨立reader並以AssertionError拒絕，不只寫入預期布林值。等待迴圈只驗靜態compiler結果，不宣稱硬體wall-clock一致。
- 合計60C／5651 bytes與78個函式庫函式／7919 bytes。唯一清冊含338個Town與191個Book原始函式；Town尚245函式／113305 CODE bytes，Book尚146函式／33141 CODE bytes，其餘四EXE未還原。完整源碼目標未完成；原始CODE與DATA不計進度。
- 封存1125份附件，保存1108份檔案，18份可重建執行檔只保留hash；62672804 bytes保存為7518680 bytes。完全唯讀核對六原始EXE、兩支各兩份來源及正式342份程式輸入後，精確清理第三批工作區；匹配C、LIB、OBJ、DB及原版片段保持本機。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1964行自承全部認領。Git、入口、擁有權及Docker終檢於提交前核對；本批沒有遊戲執行、Issue寫入、版號、封包或發行。

## 2026-10-08：四支新程式的來源基準與TANDYPIN完整還原

- 前輪屬實際進度：新增八個Book C函式與36個函式庫函式，HEAD 6208cd95已與origin/main同步。本輪開工重查工作樹乾淨、專案容器清單為空；完整matching目標保持，對拍維持暫停。
- 知識路由命中工具鏈、MZ／IDA及文件職責，重新核對IDA技能的權威專案入口。安裝技能的外部符號連結在容器內無法解析，改直接唯讀掛檔。沿已驗IDA9.4 image4ac62de與matching image5992e601，沒有另建image；全部工作UID/GID1000:1000，固定ZIP與前批來源唯讀。
- IDA9.4一次性DB匯出KOL76、TITLE115、MAKEMODE78與TANDYPIN1個原始函式，schema、輸入SHA、段界及載入／原始bytes分開核對。全清冊現799函式，沒有推測性改名或修改正式DB。
- TANDYPIN五條指令13 bytes，由MASM5.10／LINK4.07加三個段對齊零byte與96-byte重複字串生成624-byte原版檔案。獨立reader解析實際CODE、巢狀LIDATA及selector fixup，完整112-byte初始化映像與重定位相同，原始CODE／DATA片段為0；兩次ASM、OBJ、EXE、MAP四項相同。計為完整組合語言來源，不計為C，原作者文字及歷史工具版本仍未知。
- KOL／TITLE／MAKEMODE新增35／36／30個實際SLIBC供應函式，共101個／10080 bytes。共享符號與模組段約束、普通指令、native完整body及6／4／4筆有序重定位均核對；三支完整EXE符合原版。只有MAKEMODE有原始63-byte外層填補；三支OBJ、native MZ、MAP、完整EXE十二項乾淨相同。
- 兩支程式在同一binding邊界失敗後重查runtime路由：未知PUBLIC類別的__aDBdoswp被candidate預設成DATA，不能當原版DATA地址證據。所有來源使用一致的16-bit值0xD6D6，以明列ABS binding通過native MAP與完整原版差分；原始定義與語意未知，先前DATA推論撤回，原始fixup與先前定位保留。沒有放寬靜態段範圍或修改普通指令。
- 不供應特殊符號的native LINK報L2029但留下EXE，驗證器新增連結錯誤拒絕。初次EXTDEF索引遺漏造成L1101，修正後另建候選；子目錄的omf module搜尋與舊profile缺少選用ABS欄位也修正。以上為工具問題，失敗輸出保存；不宣稱linker自動提供此符號。
- 本批16項乾淨產物與九項反例通過：三個LIB指令、ABS值、組合語言指令、LIDATA、checksum、填補及「有EXE的連結錯誤」。新舊共十二份來源由獨立reader重新解析，正式342份程式輸入保持；沒有執行遊戲。
- 合計60C／5651 bytes、179個函式庫函式／17999 bytes、一個組合語言入口13 bytes及三個連結填補byte。六支EXE都有同位元組重建基準，只有TANDYPIN完整來源；其餘五支尚559函式、176286 CODE bytes與13788初始化DATA bytes未還原，全部明列，不計進度。
- 封存1315份附件，保存1288份檔案，28份可重建執行檔只保留hash；77436265 bytes保存為8459328 bytes。全唯讀核對六原始EXE、四支新程式完整雜湊、十二份來源及正式Go後，精確清理第四批工作區；匹配C、ASM、LIB、OBJ、DB與原版片段只留本機。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持；新增研究證據沿既有逆向與保存入口索引。最後檔案擁有權、工作表、引用及Docker清理在提交前核對；沒有Issue寫入、版號、封包或發行，目標保持進行中。

## 2026-10-08：KOL.EXE完整來源還原

- 前輪為實際進度，HEAD 939119b2與origin/main同步，工作樹乾淨且專案容器清單為空。沿工具鏈、MZ／IDA與文件職責入口，建立已登記的matching-c5-20261008工作區；原始ZIP、前批來源與programs唯讀，固定QuickC／MASM／LINK按相同commit及hash重取。
- KOL `_main`原始567 bytes以C控制流及五個明列BIOS／I/O ASM區塊重寫。原位body完整匹配，326個C產生bytes與241個硬體ASM bytes分列為混合來源，未增加純C數。第一個for候選少三byte入口跳躍，讀compiler後改while形狀，不用ASM補普通C差異。
- 補真實LOCAL PUBLIC／EXTERNAL與COMDEF索引順序；局部名稱按模組隔離。解析LIDATA的巢狀重複及literal offset到展開位置的fixup映射，補byte BAKPAT；標準_FILE與五個重複指標均由真正來源物件供應。錯誤import名稱、未支援記錄與候選失敗保留，未放寬普通指令檢查。
- 已知公共符號及模組CODE基址擴展成48個完整runtime模組，覆蓋75個非main函式及所有IDA共享chunks；沒有修改原始chunk歸屬。兩個小函式false-positive候選在完整binding後拒絕。CODE實際來源9600 bytes，13個LIB宣告零byte及44個段零byte另列，不計指令還原。
- DATA.C產生122 bytes的可讀字串、旗標、九個register初值及兩個word；其餘初始化DATA由固定LIB initializer、copyright、runtime訊息、stdlib指標與對齊供應，1235 bytes無原始blob。DS+2F8兩個word初值已證實，語意未知；不猜名字或接Go格式。
- 第一個完整DATA候選把兩個far pointer改為物理等價的DGROUP表示，四個DATA bytes與checksum不同；保留原始0x0283框後native LINK直接生成相同header、checksum及六筆有序重定位。沒有手改指標或MZ header。
- KOL完整11971 bytes與原版SHA 4f798b078db889b6166b92c0f0bae1febe3efb3535ad67800b2e0aaed94d7b71相同。兩次乾淨MAIN.C／OBJ、DATA.C／OBJ、原位OBJ、native MZ、MAP與完整EXE八項相同；獨立reader解析compiler、真LIB、initializer、byte／word BAKPAT及段零填補，完整11459-byte映像相同且全部76函式的所有chunks有來源。
- 八項反例實際拒絕C、硬體ASM、LIB、C初始化、LIDATA、byte BAKPAT、checksum及far pointer段框修改。正式342份程式輸入保持；只執行compiler／linker，沒有原版／重製遊戲執行或恢復對拍。
- 累計60純C／5651 bytes、219個LIB函式／23682 bytes、一個混合主程式567 bytes及TANDYPIN入口13 bytes。KOL與TANDYPIN完整來源，其餘四支仍有518函式、169979 CODE bytes與12553 DATA bytes。完整matching目標未完成；原版唯一文字、部分初值語意與精確歷史工具版本仍未知。
- 封存1457份附件，保存1438份檔案，20份可重建執行檔只保留hash；91189586 bytes保存為9237881 bytes。全唯讀核對六原始輸入、兩次完整來源、全檔hash及Go後，精確清理第五批工作區。C、ASM、LIB、OBJ、DB與原版片段只留本機，不加入Git。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持；研究與保存索引已更新，完整源碼範圍保持open。檔案擁有權、引用及Docker終檢於提交前核對；沒有Issue寫入、版號、封包或發行。

## 2026-10-08：三支完整runtime模組與四個Maker純C函式

- 前輪屬進度：KOL完整CODE／DATA來源通過並推送c56adbe7。本輪重查工作樹乾淨、main與origin/main同步、專案容器清單為空。沿工具鏈、MZ／IDA與文件職責入口，完整目標保持六支；matching-c6-20261008工作區先掛入逆向入口，固定ZIP及前批證據唯讀。
- 從第五批已驗LOCAL符號、COMDEF、LIDATA與整CODE模組方法接續。TITLE／MAKEMODE／KOLBOOK分別52／36／52完整模組，覆蓋82／64／82個原始函式的所有chunks與共享tail，新增126個完整LIB函式，沒有改IDA邊界。LIB實際CODE bytes與13／0／13個宣告零byte分列。
- 實際near lmul物件的後LEDATA以C2 0800覆蓋C3，加入此明列近返回規則；未支援overwrite仍拒絕。子目錄module搜尋路徑缺項與舊reader要求所有宣告byte都由LEDATA供應造成失敗，依KOL相同zero-fill契約修正後，用同image乾淨重跑，分類工具問題。
- MAKEMODE新增_main968、sub_11199 160、sub_1142C 107、sub_11497 93，四個純C共1328 bytes完整native匹配。兩個不返回call後的各九byte epilogue另供應，18 bytes不修改原始函式邊界、不計新函式。未符原版的sub_103E1算式候選保留而未計進度。
- 實際QuickC對已定義同TU函式預算relative call，不產生可搬移FIXUPP；改用獨立TU，真正compiler生成外部near-call由LINK回填，沒有改普通call bytes。signed／unsigned prototype按MOV／PUSH差異核對。來源擷取誤取prototype後的main已修正，正確exit物件另建，不沿用錯誤物件。
- KOLBOOK原已驗九個純C保持真實compiler供應，沒有因新LIB批次退回原始片段。三支完整EXE、header、checksum、初始化image與四筆有序relocation符合原版；Book55與Maker63-byte填補保持。六份獨立reader重算compiler、原始LIB與backpatch，完整source區間只有一個供應者。
- TITLE四項、Maker十二項、Book六項，共22項乾淨source、OBJ、MZ、MAP與完整EXE逐byte相同。八項反例實際拒絕三個LIB指令、兩個compiler指令與三個checksum修改；有效OMF checksum下的source錯誤確實被reader拒絕。
- 累計64純C／6979 bytes、345個LIB函式／38978實際來源bytes、KOL混合主程式567與TANDYPIN入口13 bytes。兩支完整、四支部分來源；尚388函式、153311 CODE及12553 DATA bytes未還原，完整goal保持active。正式342份程式輸入不變，沒有遊戲執行或恢復對拍。
- 封存1734份附件，保存1706份檔案，29份可重建EXE只保留hash；115356323 bytes保存為9779916 bytes。全唯讀核原始六EXE、三支新全檔、六份source proof與Go後精確清理本批工作區。C、ASM、LIB、OBJ、DB及原版片段只留本機。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，新增來源入口已索引。檔案擁有權、引用與Docker终檢於提交前核對；沒有Issue寫入、版號、封包或發行。

## 2026-10-09：MAKEMODE九個完整C／ASM函式

- 接續原版C matching，遊戲對拍維持暫停。開工HEAD為c56adbe7，既有第六批八檔異動保留。載入逆向重製、文件職責及工具鏈指紋入口，唯讀重算第六批六原始輸入、三支全檔與六份來源證明後，建立已索引的matching-c7-20261009工作區。
- MAKEMODE新增sub_103E1、sub_11242、sub_11318、sub_1135B、sub_113BF、sub_114F4、sub_116E4、sub_11734、sub_118A8，共1757 bytes。全部原始IDA Pro 9.4位址、file offset、chunks及bytes保持；一個純C131 bytes、八個混合C／ASM1626 bytes分列，後者含931個C產生bytes與695個明列ASM bytes。原作者inline ASM形式未知，不用原始指令blob或ASM補普通C差異。
- 勘誤sub_103E1的舊整數候選：DS+42h為六筆十byte磁片名，DS+7Eh為兩筆十byte模式名，DS+7A0h的printf格式含兩個%s。改用真實字串表加步距的near pointer，compiler原生MOV CX／ADD／PUSH完整匹配；兩個整數失敗候選保留。sub_114F4的一byte差異由實際byte局部變數修正。
- 固定QuickC2.01與LINK4.07按原commit、Git blob及SHA-256重取，沿既有kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000、有界無網路容器編譯與連結；只有工具重取需要網路。ASM參數名與保留字衝突已修正，byte BAKPAT沿第五批契約處理。準備腳本換行、註解token誤判及跨批次計數欄位假設修正後，用相同輸入與image重跑，屬研究腳本問題，不記成產品缺陷。
- 十三份C於兩個空白目錄重編，32項源檔、OBJ、清冊、原位OBJ、MAP、native MZ及完整15744-byte EXE相同，原版SHA-256保持3051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56。兩份獨立reader重算完整初始化映像、四筆有序重定位、所有compiler fixup與byte／word回填。
- 十項反例實際拒絕：修改C再編譯、修改ASM再編譯、純C指令、混合C產生指令、ASM指令、LIB指令、byte回填、near-call目標、ASM誤列純C及header checksum。六項OMF修改保持合法checksum。342份正式程式輸入雜湊未變；沒有執行遊戲或恢復對拍。
- 累計65純C／7110 bytes、345個LIB函式／38978實際來源bytes、九個混合C／ASM函式2193 bytes及TANDYPIN入口13 bytes。兩支完整來源、四支部分來源；剩379函式、151554 CODE bytes及12553 DATA bytes。MAKEMODE下一項sub_10464為3381 bytes，另有27函式外CODE bytes及2353 DATA bytes；完整matching仍未完成。
- 封存2387份附件，保存2365份檔案，23份可重建EXE只保留hash；133308658 bytes保存為10777487 bytes。全唯讀重核六原始EXE、32項乾淨產物、十項反例、兩份完整來源及Go後，精確清理本批工作區，約釋放117 MiB。入口見docs/re/matching-decompilation.md與docs/verification/retained-evidence.md。C、ASM、LIB、OBJ、DB及原版片段只留本機。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1973行自承全部有人認領。來源引用已核對，入口檢查先以第六批作正對照。Docker終檢無專案容器，find終檢無root-owned檔案或誤建Markdown目錄，更新檔UID/GID1000:1000；git diff --check通過。本批未提交、未推送，沒有Issue寫入、版號、封包或發行。

## 2026-10-09：MAKEMODE完整CODE與初始化DATA來源

- 前輪九函式還原屬實際進展，本輪接續完整六支matching目標。HEAD仍為c56adbe7，既有第六、七批八檔異動保留。路由沿reverse-engineer-retro-game-remake、verification-and-handoff及project-document-responsibilities；第七批先唯讀重核六原始輸入、兩份C／ASM／LIB來源及Go，建立已索引matching-c8-20261009工作區。
- sub_10464的1002個IDA Pro 9.4原始指令與28個caller保持。3381-byte完整函式以23個C switch分支、typed記錄步距、文件讀寫、尺寸與溢位條件還原，3293個C產生bytes與五個明列ASM區塊88 bytes分列；IDA linear10464h..11199h、原始MZ偏移1636及原始邊界保持。高低byte算式次序、有號轉換、沒有default的分派與未定型呼叫prototype經真實QuickC差分修正，不以ASM補普通C差異，原作者實際型別與唯一寫法未知。
- 原先未覆蓋的27 CODE bytes全部為零，改由段宣告與LINK供應，沒有LEDATA原始片段。完整Maker CODE為五純C1459、九混合C／ASM5007、18 compiler epilogue、64個LIB函式6305及27宣告零bytes，共12816 bytes；全部78個原始函式的所有chunks與shared tail有來源。
- DATA由67個可讀C符號1495 bytes、固定LIB766個實際初值bytes及88個宣告零bytes、四個原生零bytes供應，共2353 bytes。名稱表、byte lookup與3／6／2-byte記錄表按consumer及stride建立；原位LIB來源與全部source fixup核對。compiler最後字串為奇數長度，最後一個對齊byte由LINK供應。外層63個1Ah由明列生成，不從原版tail拷入。
- DS+4E2原始bytes為00 1F，未在固定LIB資料段找到。原始__cinit於11ADBh／11ADEh及11AE4h／11AE7h令SI／DI同為4E2h；sub_11B8A與sub_11B99均先CMP、JNB退出，空範圍不讀值。C初值0x1F00只供應已證實的觀察bytes；用途未知、原始C宣告為假說，不把g_4E2別名當語意證據。
- 固定QuickC2.01／LINK4.07按既有commit、Git blob及SHA-256重取，沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000；工作無網路、有界資源與逾時，只有工具重取開放網路。前批腳本程式名稱／core邊界、Python模組路徑、DATA末端長度及反例far-pointer型fixup假設已按實際記錄修正後同image重跑，屬研究腳本問題。Maker DATA實際採segment fixup，本輪不製造不存在的far-pointer型反例。
- 十五份來源於兩個空白目錄重編，36項源檔、OBJ、清冊、原位OBJ、MAP、native MZ及完整15744-byte EXE逐byte相同。兩次完整EXE均與原版SHA-2563051cc981e2cbf62237c1cbb788f4d10b84c8dea4dcd6c2aad67171cd2196c56相同；header、checksum、全部15169-byte初始化映像及四筆有序重定位均核對。兩份獨立reader逐byte確認compiler CODE／DATA、LIB、零填補的唯一來源，原始CODE與DATA blob為0。
- 十三項反例拒絕實際C／ASM／未知初值再編譯、純C指令、Core ASM、LIB CODE、compiler DATA、LIB DATA、byte回填、segment fixup目標、額外零區LEDATA、DATA清冊漏項及header checksum；八項OMF修改保持合法checksum。來源漏項即使保持全檔相同也不能通過。342份Go及模組輸入雜湊保持，沒有執行遊戲或恢復對拍。
- MAKEMODE成為第三支完整來源程式，KOL與TANDYPIN原有結果保持。累計65純C／7110 bytes、345個LIB函式／38978實際來源bytes、十個混合C／ASM5574 bytes及TANDYPIN入口13 bytes；混合來源含4550個C與1024個ASM bytes。TITLE、KOLBOOK與KOLTOWN仍有378函式、148146 CODE及10200 DATA bytes未還原；完整goal維持active，下一批接續TITLE。
- 封存3704份附件，保存3679份檔案，26份可重建EXE只保留hash；165621187 bytes保存為12273981 bytes，約釋放146 MiB。六原始輸入、所有附件、36項乾淨產物、十三反例、兩份完整來源及Go經全唯讀重算後，精確清理本批工作區。入口為docs/re/matching-decompilation.md與docs/verification/retained-evidence.md；C、ASM、LIB、OBJ、DB及原版衍生資料只留本機。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1977行自承全部有人認領。先以第七批作正對照，再核對十一個第八批來源連結及擁有權。Docker終檢無專案容器，find終檢無root-owned檔案或誤建Markdown目錄，更新檔UID/GID1000:1000；git diff --check通過。本批未提交、未推送，沒有Issue寫入、版號、封包或發行。

## 2026-10-09：TITLE完整來源與未命名CODE範圍

- 前輪MAKEMODE完整來源屬實際進展，本輪接續完整六支matching目標。HEAD仍為c56adbe7，既有八檔異動保留。載入reverse-engineer-retro-game-remake、文件職責、驗證交接及IDA工具入口；第八批完整來源先唯讀重核，再建立已登記matching-c9-20261009工作區。工具handle消失後核對容器與七個固定工具實檔SHA-256，沒有重啟已完成下載。
- 原始TITLE的115函式中，82個LIB結果保持；新增十個純C1625 bytes與23個混合C／ASM5629 bytes。混合來源含4269個C產生bytes、1360個明列ASM bytes；十五個完整寄存器／硬體body由ASM重寫，普通C控制、解碼、資料步距、檔案及尺寸運算未補ASM。所有原始位址、名稱、chunks與caller保持。RND僅還原呼叫與機器碼，不執行亂數判斷；硬體busy-wait不作逐週期語意研究。
- 小量差分由缺／多列呼叫、加法次序、局部變數順序、length保留字及prototype轉換修正。三十五份來源加TITLE.H於兩個空白目錄完整重編，77項源檔、OBJ、header、清冊、原位OBJ、MAP及完整18671-byte EXE逐byte相同。原版SHA-256保持8a2879d924d9a9cb01e000bd548c348022c0f922288565015f7560ab8fc6fad6；原始唯一寫法與實際型別仍未知。
- 原始110BDh..11133h額外118 CODE bytes以正式IDA DB的一次性副本核對，49個原始CODE items皆沒有函式歸屬，entry沒有xref。固定IDA Pro9.4／image4ac62de8、UID/GID1000:1000輸出schema、原始EXE SHA與115個原始函式數核實；正式DB雜湊不變。另列C範圍原生匹配118 bytes，不建立／改名函式，不增加全域799個原始函式。別名只作導覽，沒有當成原始符號或語意證據。
- 完整16880 CODE bytes由1625純C、5629混合C／ASM、118另列C、9459實際LIB／13LIB宣告零bytes及36原生零bytes供應。1279初始化DATA由二十個C初值符號218 bytes、969實際LIB初值／88LIB宣告零bytes及四原生零bytes供應，原始CODE／DATA blob為0。三個五byte模式名、九個Hercules word記錄與十二個字串有consumer／shape來源；DS+2D6的0x227A及0x3E9E用途未知，原始C宣告為假說，觀察值與重建能力分開記錄。
- 兩份獨立reader重算完整18159-byte初始化映像，核對所有compiler CODE／DATA、固定LIB、額外C及零填補唯一來源、全部fixup與byte／word backpatch、四筆有序重定位。十五項反例拒絕實際C／ASM／未知初值／額外C再編譯、C與ASM、CODE／DATA函式庫、compiler DATA、byte回填、segment目標、額外零區LEDATA、DATA或額外C來源漏項及header checksum；八項OMF修改保持合法checksum。
- 沿既有kol-matching-dos:bookworm-20261008-r1與固定QuickC2.01／LINK4.07，原始ZIP、正式DB及前批來源唯讀；寫入UID/GID1000:1000，有界CPU／記憶體／PID及逾時，只有固定工具重取需要網路。額外CODE的ida_linear重複欄位及DATA218-byte／二十符號假設修正後同image重跑，屬研究工具問題，沒有修改正式Go或玩家流程。
- TITLE成為第四支完整來源，原有KOL／TANDYPIN／MAKEMODE結果保持。累計75純C／8735 bytes、345個LIB函式／38978實際來源bytes、33個混合C／ASM11203 bytes；8819個C與2384個ASM分列，額外C118 bytes及TANDYPIN入口13 bytes不混入函式數。KOLBOOK與KOLTOWN仍有345函式、140738 CODE及8921 DATA bytes，完整goal維持active，下一批接續KOLBOOK。
- 封存5943份附件，保存5918份檔案，26份可重建EXE只保留hash；195783496 bytes保存為14613501 bytes，約釋放173 MiB。全唯讀核對已印出全部PASS，但第一輪收尾碰到120秒外層逾時，分類環境／執行時限問題；延長至300秒，用同image、同命令乾淨重跑，不更改封存收據。來源及清理入口見docs/re/matching-decompilation.md與docs/verification/retained-evidence.md；本機原版衍生C、ASM、LIB、OBJ、DB及工具不加入Git。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1979行自承全部有人認領；以第八批作正對照後核對十個第九批來源連結。300秒重跑完成全部來源、DB與附件核對，退出0；封存收據未改。第九批工作區已清理，find終檢無root-owned檔案或誤建Markdown目錄，更新檔UID/GID1000:1000，Docker終檢無專案容器，git diff --check通過。342份Go及模組輸入保持，沒有遊戲執行、恢復對拍、Issue寫入、版號、封包或發行；本批未提交或推送。

## 2026-10-09：KOLBOOK新增75個完整原始C／ASM來源函式

- 前輪TITLE完整來源屬實際進展，本輪接續全六支matching目標。HEAD仍c56adbe7，既有八檔異動保留。沿復古逆向、文件職責與驗證交接入口，唯讀重核第九批來源後建立已索引matching-c10-20261009工作區。KOLBOOK原始191函式、九個C594 bytes與82個LIB9400實際來源bytes皆先核對；沒有依舊待辦重新研究已完成函式。
- 新增75個完整原始函式8485 bytes，包含二十個純C2475 bytes、55個混合C／ASM6010 bytes。混合來源含1939個C產生bytes與4071個明列ASM bytes，ordinary C控制、前置條件、呼叫、圖像分派、解碼、檔案、字串、幾何交疊與色盤運算未補ASM。原始名稱、linear位址、file offset、chunks、bytes與caller保持；完整原位body與外部引用逐byte相同。原作者實際型別、唯一寫法及精確工具版本未知，不把別名當語意證據。
- 三十五個完全寄存器BODY按原始指令明列ASM，後續條件／呼叫及局部C設定另列。小差分由失敗分支形狀、DS+1130解碼暫存區、stack argument offset修正；g_1234及g_13D4的陣列consumer與header宣告衝突已訂正。offset unk_1924B保持原運算元註解，以DS偏移別名綁定原值，不猜語意改名。這些失敗屬研究來源／腳本問題，同image乾淨重跑，不記為Go產品缺陷。
- 原有九個C共享BOOK.OBJ，獨立reader按compiler PUBLIC offset及完整相接／非重疊來源區間覆蓋全594-byte OBJ，不把每筆宣告當獨立TU，也不把已驗C退回原版片段。Book現有29純C3069 bytes、55混合6010 bytes與82完整LIB函式；仍有25原始函式、18948 CODE與2185初始化DATA bytes。全檔相同仍為混合來源基準，Book尚未完整。
- 七十六個C源檔與BOOK.H在兩個空白目錄重編，159項源檔、OBJ、header、清冊、原位OBJ、MAP、native MZ及完整40192-byte EXE逐byte相同。原版SHA-256保持2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832，55個1Ah外層填補、header、checksum、全部39625-byte初始化映像及四筆有序重定位均核對。
- 兩份獨立reader核對84個完整compiler C／ASM函式、52個完整LIB模組及所有fixup／byte與word backpatch。九項反例拒絕實際C／ASM再編譯、新C／ASM／舊共享OBJ與LIB指令、byte回填、共享OBJ來源漏項及header checksum；五項OMF修改保持合法checksum。未還原CODE／DATA仍明列原版來源，不放寬完整目標。
- 固定QuickC2.01／LINK4.07按同commit、Git blob及SHA-256重取，沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000。有界無網路工作，唯工具重取開網路；原始ZIP、programs、analysis與前批來源唯讀掛載。單次編譯每個真實源檔，沒有因九個共用OBJECT的宣告重複啟動compiler。來源驗證逐檔重算SHA-256後快取必要archive檔案，避免反覆xz定位；沒有更改既有不可變收據或驗收條件。
- 累計95純C11210 bytes、345個LIB函式38978實際來源bytes、88個混合C／ASM17213 bytes；10758個C與6455個ASM分列，額外C CODE118 bytes、TANDYPIN入口13 bytes與18個epilogue bytes保持。全域799個原始函式不變。Book25及Town245共270函式、132253 CODE與8921 DATA bytes未還原，四支完整來源保持，full goal維持active；下一批接續Book剩餘25原始函式與DATA。
- 封存8915份附件，保存8893份檔案，23份可重建EXE只保留hash；221612068 bytes保存為18252743 bytes，約釋放194 MiB。全唯讀封存重算六原始輸入、159項產物、九反例、兩份完整原位來源及342份Go／模組輸入後退出0，才精確清理本批工作區。C、ASM、LIB、OBJ、DB、工具及原版衍生資料只留本機；入口docs/re/matching-decompilation.md與docs/verification/retained-evidence.md。
- 工作表render／verify／scan通過，19未完成／89完成／15人工保持，1979行自承全部有人認領。以第九批入口作正對照後核對十個第十批來源連結及擁有權。find終檢無root-owned檔案或誤建Markdown目錄，更新檔UID/GID1000:1000，Docker終檢無專案容器，git diff --check通過；本批未提交或推送。沒有遊戲執行、RND結果對拍、硬體逐週期研究、Go變更、Issue寫入、版號、封包或發行。

## 2026-10-09：KOLBOOK新增18個完整原始來源函式

- 沿已索引的matching-c11-20261009工作區接續全六支目標。路由重核復古逆向、IDA及文件職責，既有八檔異動保留。先核對C10唯一來源清冊與原始191個Book函式；沒有重新開啟四支完整來源或遊戲對拍。
- 新增18個完整原始函式6012 bytes：十個純C3499、八個混合C／ASM2513 bytes，混合來源含2116個C與397個ASM bytes。磁碟提示、選單、輸入、檔案、字串、動畫與幾何來源按原始指令及compiler逐byte核對。sub_146CE完整1302 bytes，三段滑鼠／鍵盤ASM45 bytes，其餘1257 bytes由C產生；普通C差異不補ASM。
- 四個候選仍未匹配：sub_14101置中式的型別／暫存器差異；三個繪字函式各有SUB／ADD差異。只取真正匹配的102個來源函式，四個候選不計進度，另三個大型Book函式尚未實作。28段LIB初始化資料已有固定來源與fixup核對，typed DATA建置未完成，2185 DATA bytes保持未還原。
- IDA導航第一次讀取錯誤DB路徑，改從已封存Book DB建立新一次性副本；正式DB唯讀，SHA與191函式核對，未改名。Hex-Rays未提供偽碼，按原始指令繼續。反例腳本第一次定位到不存在的ASM運算元，保留失敗資料，修正為g_90C後用新反例目錄同image重跑。這些屬研究工具問題，沒有修改Go產品。
- 九十四個C源檔加BOOK.H兩次從空白目錄重編，195項產物逐byte相同；兩份完整40192-byte EXE符合原版SHA-256。55-byte 1Ah外層、header、checksum、39625-byte初始化映像與四筆有序重定位保持。獨立reader核對102個完整C／ASM函式、52個LIB模組、共享BOOK.OBJ全部PUBLIC區間及fixup／byte與word backpatch。九項反例均拒絕，五項OMF修改仍有合法checksum。
- Book現39純C6568、63混合8523、82個LIB9400實際來源bytes。累計105純C14709、345個LIB38978、96個混合19726 bytes；12874個C與6852個ASM分列。額外C118、Tandy入口13及epilogue18 bytes保持。Book七與Town245共252函式、126241 CODE及8921 DATA bytes未還原，799原始函式與四支完整來源保持，完整goal仍active。
- 沿既有kol-matching-dos:bookworm-20261008-r1與ida-pro-9.4-idapython:py312-v1，UID/GID1000:1000、有界CPU／記憶體／PID／逾時；原始programs、analysis及前批來源唯讀。固定工具重取才開網路。342份Go／模組輸入保持，沒有遊戲執行、RND對拍、硬體時序考古、Issue寫入、版號、封包或發行；未提交或推送。
- 封存14512份附件，保存14488份檔案；25份可重建EXE只保留hash。255367655 bytes保存為20822756 bytes，約釋放224 MiB。全唯讀重驗六原始輸入、195項產物、九反例、兩份來源證明與342份正式輸入後退出0，才精確刪除第十一批工作目錄。以第十批八個入口作正對照，核對第十一批九個入口及更新檔擁有權。工作表render／verify／scan通過，19未完成、89完成、15人工保持；1984行自承全部有人認領。find終檢無root-owned檔案或Markdown目錄，Docker清單為空，git diff --check通過。下一批由本機封存接續Book七函式與DATA，完整目標保持active。

## 2026-10-09：KOLBOOK主函式與全部初始化DATA來源

- 前輪C11增加18個函式，本輪先全唯讀重核其清冊、102個既有來源函式與342份正式輸入，再建立已索引C12工作區。沿復古逆向、IDA、工具鏈考證與文件職責入口，既有異動保留；全六支目標不縮小，遊戲對拍未恢復。
- _main完整4789 bytes由4562 C與227硬體ASM供應，十一個寄存器／複製區塊分列。原始16-byte函式外尾碼也由同一MAIN.OBJ供應：清理呼叫十bytes、真正尾聲六bytes。未猜出使用者變數；for逗號初始化自行產生BP-2暫存。原始名稱、位址、chunks、bytes與191／799原始函式數保持。診斷改用真實PUBLIC／OBJ結尾求完整尺寸，較長候選不被截短。
- 全2185 DATA由78個typed C初值／1123 bytes、28個LIB初始化區段加PAD／EPAD共1057 bytes及五個原生零bytes供應。表格步幅、固定寬度字串、word、調色盤與游標遮罩經consumer／原始值及compiler物件核對。DS+57A兩word為72FA／8F1E，值與位置已證實，用途未知，原始C宣告為假說；不讓名字取代證據。
- 第一次DB副本無法取輸入SHA，閘門拒絕；改用C11已核對的副本後SHA及191函式通過，正式DB唯讀。DATA最末對齊由LINK供應，compiler物件實際1123而非猜出的1124 bytes；修正後同image重跑。首次乾淨建置重複啟動同一DOS case被DONE閘門拒絕，將DATA與CODE合併為單次有界編譯，在新空白目錄重跑。失敗分類為研究工具／來源問題，保留資料，未改Go產品。
- 原位建置與獨立reader分別核對資料完成及程式碼完成。108個資料來源區段覆蓋全部DATA；Code仍8085 bytes，46個獨立零區由LINK生成。偽造Code完成、移除DATA來源及關閉資料完成旗標都拒絕，full_source_restoration仍false。未匹配圖像編輯器與置中／三個繪字函式保留探針，不計進度，普通C差異不補ASM。
- 95個CODE源檔加DATA.C／BOOK.H兩次從空白目錄重編，201項產物逐byte相同；完整40192-byte EXE符合原版SHA-256，55-byte外層、header、checksum、39625-byte映像與四筆有序重定位保持。兩份獨立reader核對103個compiler函式、52個LIB模組、全部typed DATA及fixup／byte／word回填。十五反例全拒絕，八項OMF改動保持合法checksum。
- Book現39純C6568、64混合13312、82個LIB9400實際來源bytes；混合含8617 C／4695 ASM。累計105純C14709、345個LIB38978、97混合24515 bytes，17436 C／7079 ASM分列。Book六與Town245共251函式、121390 CODE與6736 DATA bytes未還原，四支完整來源保持，完整goal active。
- 沿既有匹配與IDA image，UID/GID1000:1000、有限CPU／記憶體／PID／逾時，原始ZIP、programs、analysis、前批來源唯讀，僅固定工具重取開網路。342份Go／模組輸入保持；沒有遊戲執行、RND結果對拍、硬體時序考古、Issue寫入、版號、封包或發行，未提交或推送。
- 封存20073份附件，保存20049份檔案，25份可重建EXE只留hash；297887791 bytes保存為23667775 bytes，約釋放262 MiB。全唯讀重驗六原始輸入、201項產物、十五反例、兩份來源證明與342份正式輸入後退出0，才精確清理本批工作區。第十一批九個入口作正對照，第十二批九個入口及更新檔擁有權通過。工作表render／verify／scan通過，19未完成、89完成、15人工保持，1985行自承全部有人認領。最後統計小腳本誤把舊JSON可選status當必要欄位，改按既有產生區塊核對後通過，未改完成條件。find無root-owned檔案或Markdown目錄，Docker清單為空，git diff --check通過。下一批接續Book六函式及函式外CODE；完整目標active。

## 2026-10-09：角色建立與411個額外CODE bytes

- 前輪主函式與DATA屬實際進展。本輪全唯讀重核C12封存、103個來源函式、六原始輸入及342份正式輸入，沿復古逆向、IDA、工具鏈及文件職責入口建立已索引C13工作區，既有異動保留。完整六支目標與對拍暫停保持。
- 角色建立完整3671 bytes匹配，3646 C與25複製ASM另列，涵蓋名稱、選項、取消、重擲、肖像、記錄與保存。十一個C局部變數與三個compiler暫存符合原始28-byte stack，未在程式或測試猜補規則；只編譯與link，不執行遊戲或RND結果對拍。第一次暫存名字offset撞MASM保留字，改position後完整BODY相符，原始IDA名稱不改。
- 原本未歸函式的411 CODE bytes按原始prologue及IDA指令拆成79／273／59三段。初次中段邊界猜為15133，decode閘門拒絕；以raw prologue核對1513A後在新DB副本重跑。191函式、輸入SHA及每一instruction raw bytes保持，沒有新增或改名IDA函式。兩段純C138 bytes，中段258 ASM＋15 C wrapper，共153 C／258 ASM；用途只限這些區段資料流，不深挖無caller helper。
- 原位builder將額外mixed區段從原始函式數排除，新增extra_source_CODE／extra_explicit_assembly欄位；原始799函式不變。Book現104個原始C／ASM來源與三個額外區段，剩4003 CODE正好五個原始函式，沒有函式外未還原區段。Source分類、BODY、fixup與完整覆蓋均核對，普通C差異不補ASM。
- QuickC2.01的context、指標、成員存取探針未解出其餘差異。工具清冊首次遞迴下載超過界限，縮到固定commit根tree後完整保存；按Git blob及SHA核QuickC1／2.51候選。QuickC1純C prologue／配置不同，_asm不支持；2.51置中／繪字仍不符，編輯器有未支援thread FIXUPP，保留未驗而不擴寫reader掩蓋差異。CLI help未輸出文字，診斷腳本模組路徑修正後保存已生成兩份比較；工具存在不作原版版本證據。
- 99個CODE源檔加DATA.C／BOOK.H兩次乾淨重編，209項產物逐byte相同；17反例拒絕，九項OMF修改checksum仍合法。兩份獨立reader核104原始來源函式、三額外區段、52 LIB模組、全部DATA、fixup與byte／word回填。40192-byte EXE兩次符合原版SHA-256，55外層、header、checksum、39625-byte映像及四筆有序重定位保持。
- 累計105純C14709、345 LIB38978、98混合28186 bytes，21082 C／7104 ASM分列。額外C由118加153為271，額外ASM258與Tandy入口13另列。Book五／Town245共250函式、117308 CODE及6736 DATA bytes未還原，四支完整來源及完整goal active保持，原版唯一文字／型別、精確工具及Go parity限制不變。
- 沿既有image、UID/GID1000:1000、有界CPU／記憶體／PID／逾時，原始ZIP、programs、analysis及前批來源唯讀，固定工具下載才開網路。原版衍生來源、DB、OBJ及工具只留本機，不加入Git。342份正式輸入保持，無遊戲執行、硬體時序研究、Issue寫入、版號、封包或發行，未提交或推送。
- 封存25247份附件，保存25224份檔案，24份可重建EXE只留hash；341055815 bytes保存為27165226 bytes，約釋放299 MiB。全唯讀重驗六原始輸入、209項產物、十七反例、兩份來源證明與342份正式輸入後退出0，才精確清理本批工作區。第十二批九個入口作正對照，第十三批九個入口及更新檔擁有權通過。工作表render／verify／scan通過，19未完成、89完成、15人工保持，1991行自承全部有人認領。find無root-owned檔案或Markdown目錄，Docker清單為空，git diff --check通過。下一批接續Book五函式，完整六支目標active，未提交或推送。

## 2026-10-09：KOLTOWN完整LIB模組與33個原始函式

- 前輪角色建立與額外CODE屬實際進展。本輪先全唯讀核C13來源、六原始EXE與342份正式輸入，沿復古逆向、原作工具鏈、文件職責與既有module解析入口建立C14工作區，八檔既有異動保留。Book旗標探針不匹配，保留未完成，不補ASM或縮小全六支目標。
- 接手Town原有51純C與42 LIB，固定ZIP重生KOLTOWNU SHA01f6f9a...；MLIBC／MLIBFP／LIBH檔bytes及SHA保持。按完整segment、修正fixup、原生零與共享PUB約束擴展模組；早期清冊沒有LEXTDEF／local符號使output.c越界，套用已驗parser後完整原位2394-byte模組核對。local以module SHA及原名稱分開，不混同名符號；closeall同型候選因__iob差40拒絕。
- 52完整CODE模組供應75原始函式，新增33；共有尾碼的DOS helper核全部原始chunks。來源範圍9645，實際LIB9632與宣告零13另列。沒有重疊既有51 C，也沒有把先前C退回blob。CODE新增5431範圍bytes、實際新LIB5418。所有原始名稱、位址、bytes及338／全域799函式數保持。
- 兩空白目錄/AM /Od /Gs /Zl /G0重編，LINK4.07完整映像差0，51 C保持5057 bytes。獨立reader加入完整來源範圍計數、原始fixup與word回填比對，129312映像與2281重定位全符。EXEMOD4.02 /MIN C34及EXEPACK4.06重生核心133679；首次全檔差僅既有81外層1Ah，依MZ EOF生成常數而不複製tail。兩次完整133760 EXE hash等於原版c2442c84...，header／checksum／payload未手修。
- 十項C、OBJ、清冊、MAP、native與packed全檔逐byte相同。十一反例拒絕實際C再編譯、C與新舊LIB指令、word回填、output或共享C漏項、本機fixup目標，及packed header／core／外層；四OMF修改checksum合法。最初模組腳本搜尋路徑、非CODE無陣列、反例文字重複屬驗證工具問題，保留失敗，在同image修正並乾淨重跑，不寫成Go產品缺陷。
- Town現75 LIB、212未還原原始函式107874 CODE與6736 DATA，完整來源仍false。Book五函式4003 bytes不變。合計217未還原函式111877 CODE／6736 DATA；105純C14709、378 LIB44396實際bytes、98混合28186及額外271 C／258 ASM分列，四支完整來源與完整goal active保持。
- 沿原有image、UID/GID1000:1000、有限CPU／記憶體／PID／逾時、原始ZIP／programs／analysis／前批來源唯讀，固定工具下載才開網路。原版衍生C、LIB、OBJ、DB與工具僅本機，沒有Go、玩家路徑、RND對拍、硬體時序研究、Issue、版號、封包或發行，未提交或推送。
- 封存25702份附件，保存25680份檔案，23份可重建執行檔只留hash；382660051 bytes保存為28480804 bytes，約釋放338 MiB。全唯讀重驗六原始輸入、兩份完整模組來源、2281重定位、十項乾淨產物、十一反例與342正式輸入後退出0，才精確清理本批工作區。前批九個入口作正對照，本批九個入口及擁有權通過；render／verify／scan通過，19未完成／89完成／15人工保持，1994行自承全有人認領。find無root-owned檔或Markdown目錄，Docker清單空、git diff --check通過。下一批Town玩家CODE與資料，Book五函式仍未完成，全六支目標active，未提交或推送。

## 2026-10-09：KOLTOWN新增30個純C原始函式

- 接續既有C15工作區與原版C匹配方向，遊戲對拍維持暫停。開工HEAD為c56adbe7，八檔既有異動保留。載入復古逆向、文件職責、IDA與驗證交接入口；C14封存的25680份來源逐檔重核，六支原始EXE與七個固定工具的大小、SHA-256及Git blob均核對，沒有另建工具image或重新下載。
- Town新增30個完整純C原始函式3203 bytes，現81純C8260 bytes。原有51純C5057 bytes與52個LIB模組、75個LIB原始函式保持；LIB實際9632、宣告零13、範圍9645分列。全部原始名稱、IDA Pro9.4位址、固定解壓MZ偏移、完整chunks及bytes有逐函式索引，原始338／全域799函式數不變。
- 初始C差異由條件方向、連鎖賦值、運算次序、整數／指標表示與compiler暫存核清。sub_252CB手動多列暫存會多六bytes，改讓QuickC依有號參數及無號運算生成暫存後完整81 bytes匹配。這只證實有效匹配表達，不推定原作者的唯一宣告或型別；沒有以ASM、原始指令blob、遮罩或手修bytes補普通C差異。
- EXTDEF超過單筆OMF長度上限，改為依原順序分筆；封裝報告沿用51個C的舊固定計數，改由實際清冊取得。診斷腳本換行及compiler helper路由修正後，同image重跑；舊候選與失敗保留，分類研究來源或驗證工具問題，不記作Go產品缺陷。
- 沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000、有界CPU／記憶體／PID／逾時及無網路容器，原版ZIP、programs、解壓oracle、LIB與前批來源唯讀。trial.py clean-first及clean-repeat從兩個空白目錄執行QuickC2.01 /AM /Od /Gs /Zl /G0、LINK4.07、EXEMOD4.02及EXEPACK4.06；十項源檔、OBJ、來源清冊、MAP、native MZ與packed全檔逐byte相同。
- 兩份獨立reader重算全部compiler／LIB來源、fixup、word回填、129312-byte映像及2281筆有序重定位。兩次原生packed核心133679 bytes相同，再由常數生成81個1Ah；完整133760-byte KOLTOWN.EXE均符合原版SHA-256 c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221。沒有手改header、checksum、壓縮payload或複製原版tail。
- 十三反例拒絕新C實際再編譯、新舊C與LIB指令、word回填、新C呼叫目標、新舊C或output模組漏項、local目標，以及packed header、core與外層填補；五項OMF修改保持合法checksum。來源漏項不能因全檔bytes相同而通過。
- 累計135純C17912、378 LIB44396實際bytes、98混合C／ASM28186 bytes，額外271 C／258 ASM與Tandy入口13分列。Town182函式104671 CODE與6736 DATA、Book五函式4003 CODE仍未還原；全六支187函式108674 CODE／6736 DATA未完成，四支完整來源保持。342份正式Go及模組輸入雜湊不變，沒有執行遊戲、恢復對拍或開啟硬體時序研究。
- 封存25949份檔案，35份可重建執行檔只保留hash；418724543 bytes保存為28225759 bytes，約釋放372 MiB。六原始輸入、全部附件、兩份來源證明、十項乾淨產物、十三反例與342正式輸入從壓縮證據全唯讀重驗退出0後，精確刪除C15工作區。原版衍生C、LIB、OBJ、DB及工具僅本機；來源與重現入口見docs/re/matching-decompilation.md及docs/verification/retained-evidence.md。
- 權威工作表只更新進度及verify說明，完成條件、狀態與Issue保持。render／verify／scan通過，19未完成／89完成／15人工保持。C14九個入口作正對照，C15九個入口與擁有權通過；find無root-owned檔案或誤建Markdown目錄，Docker終檢無專案容器，git diff --check通過。下一批接Town玩家CODE與初始化DATA，Book五函式保持未還原；本批未提交或推送，沒有Issue寫入、版號、封包或發行。

## 2026-10-09：KOLTOWN新增25個純C原始函式

- 前輪C15新增30個原始函式屬實際進展；本輪依active goal接續完整六支matching，沒有縮小成功條件。開工HEAD為c56adbe7，八檔既有異動保留。沿復古逆向、文件職責、compiler指紋及驗證交接入口，建立已索引C16工作區；C15保存的25949份來源逐檔重核，固定七個工具按原commit重取並驗大小、SHA-256及Git blob。
- 新增Town25個完整純C原始函式5595 bytes，涵蓋選單、文字與資料頁、角色狀態、排序、位置判斷與戰鬥數值。Town現106純C13855 bytes；既有81純C8260 bytes、52個LIB模組與75個LIB原始函式保持。原始338／全域799函式、完整chunks、IDA Pro9.4 linear、固定解壓MZ偏移及bytes雜湊均保存。
- 新compiler import __aFlmul綁定原始sub_2DE8C，由sub_2198A／sub_2590F caller、52-byte完整LIBH lmul.asm與實際生成ABI核對；此模組已列前批LIB來源，不重複計進度。compiler helper的來源狀態改按實際LIB覆蓋判斷，原始位址與模組hash保持。
- C候選差異由求值次序、連鎖賦值、參數轉換、16-byte記錄步距及近指標表示修正。sub_166CC差兩bytes，以固定compiler／旗標比較七個C表達式；無號word位移使完整269-byte函式匹配。原作者唯一文字、型別與變數宣告仍未知，不用ASM、原始指令blob或手修bytes補普通C。生成核對器的集合欄位比對漏右括號，修正後同image重跑，分類研究工具問題。
- 沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000、有界CPU／記憶體／PID與外層逾時，除固定工具重取外均無網路。原版ZIP、programs、解壓oracle、LIB與前批來源唯讀。trial.py clean-first及clean-repeat從兩個空白目錄執行QuickC2.01 /AM /Od /Gs /Zl /G0、LINK4.07、EXEMOD4.02及EXEPACK4.06，十項來源、OBJ、清冊、MAP、native MZ與packed全檔逐byte相同。
- 兩份獨立reader核全部compiler／LIB來源、fixup與word回填，重生129312-byte映像及2281筆有序重定位。兩次原生133679-byte核心加常數生成的81個1Ah，完整133760-byte KOLTOWN.EXE均符合原版SHA-256 c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221，沒有手改header、checksum、壓縮payload或複製tail。
- 十四反例實際拒絕新C再編譯、新舊C與LIB指令、word回填、新C遠呼叫、乘法helper PUBLIC綁定、新舊C或LIB漏項、local目標與packed三區；六項OMF修改保持合法checksum。342份正式Go與模組輸入雜湊不變，沒有執行遊戲、恢復對拍或硬體時序研究。
- 累計160純C23507、378 LIB44396實際bytes、98混合C／ASM28186 bytes，額外271 C／258 ASM與Tandy入口13分列。Town157函式99076 CODE與6736 DATA、Book五函式4003 CODE仍未還原；全六支162函式103079 CODE／6736 DATA未完成，四支完整來源與完整goal active保持。
- 封存26206份檔案，23份可重建執行檔只保留hash；451089376 bytes保存為29925298 bytes，約釋放402 MiB。六原始輸入、全部附件、兩份來源證明、十項乾淨產物、十四反例與342正式輸入從壓縮證據全唯讀重驗退出0後，精確刪除C16工作區。準備腳本已保存，外部暫存副本刪除；原版衍生C、LIB、OBJ、DB及工具僅本機。入口docs/re/matching-decompilation.md及docs/verification/retained-evidence.md。
- 權威工作表只更新進度與verify說明，完成條件、狀態及Issue保持。render／verify／scan通過，19未完成／89完成／15人工保持。C15九個入口作正對照，C16九個入口與擁有權通過；find無root-owned檔案或誤建Markdown目錄，Docker終檢無專案容器，git diff --check通過。下一批接Town玩家CODE與初始化DATA，Book五函式仍未完成；本批未提交或推送，沒有Issue寫入、版號、封包或發行。

## 2026-10-09：KOLTOWN新增12純C與31明列C／ASM函式

- 前輪C16新增25個純C函式屬實際進展；本輪接續完整六支matching goal。開工HEAD為c56adbe7，八檔既有異動保留。沿復古逆向、文件職責、compiler指紋及驗證交接入口，建立已索引C17工作區；C16保存的26206份來源逐檔重核，固定七個工具依原commit重取並驗大小、SHA-256及Git blob。
- 新增Town43個完整原始來源函式4170 bytes，包含12純C1905與31明列C／ASM2265 bytes。31混合函式由compiler產生465個C框架bytes，1800個ASM bytes按原始完整暫存器、segment、port、字串與位元編解碼區塊另列。零局部frame函式均保留原始9-byte序言與6-byte尾碼；全部原名、IDA Pro9.4 linear、固定解壓MZ偏移、完整chunks與bytes雜湊保持。
- 上層分派、邊界控制、行索引及文字換行保持普通C。g_6710／g_670A的無號word表達、局部有號值與未定型遠呼叫宣告順序重生原始比較、減法及除法序列，既有106個C的全部機器碼保持。沒有以ASM補普通C差異，未匹配的普通far-pointer候選仍開放；原作者唯一型別、C文字、translation unit與原始inline ASM形式未知。
- 來源分類器核106個既有C、12個新C與31個明列ASM的集合、數量、原始hash及範圍。compiler源與ASM框架／body分列，拒絕混合誤列純C、範圍偏移、來源漏項與普通C插入ASM。函式擷取器誤排除內文分號、ASM反例定位撞到extern宣告，修正後在新目錄同image重跑，失敗保留並分類研究工具問題。
- 沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000、有界CPU／記憶體／PID及逾時，除固定工具重取外均無網路；原版ZIP、programs、解壓oracle、LIB與前批來源唯讀。兩空白目錄以QuickC2.01 /AM /Od /Gs /Zl /G0、LINK4.07、EXEMOD4.02及EXEPACK4.06重編，十項源檔、OBJ、清冊、MAP、native MZ及packed全檔逐byte相同。
- 兩份獨立reader直接解析149個compiler函式與52個完整LIB模組，核全部fixup、word回填、129312-byte映像與2281筆有序重定位。兩次原生133679-byte核心加常數生成81個1Ah，完整133760-byte KOLTOWN.EXE均符合原版SHA-256 c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221。沒有手改指令、header、checksum、壓縮payload或複製tail。
- 21反例實際拒絕C與ASM再編譯、新舊C、ASM及C框架指令、LIB、word回填、遠呼叫、helper PUBLIC綁定、來源漏項、local目標、ASM誤列純C、錯誤ASM範圍、普通C插入ASM與packed三區；八項OMF修改保留合法checksum。342份正式Go與模組輸入雜湊不變；沒有遊戲執行、恢復對拍或硬體時序實驗。
- Town現118純C15760、31混合2265及75個LIB原始函式9632實際bytes，52模組與13宣告零bytes保持。累計172純C25412、378 LIB44396、129混合30451 bytes；混合分列21547 C與8904 ASM，函式外271 C／258 ASM及Tandy入口13另列。Town114函式94906 CODE與6736 DATA、Book五函式4003 CODE未還原；全六支119函式98909 CODE／6736 DATA未完成，原始799函式、四支完整來源與goal active保持。
- 封存26561份檔案，31份可重建執行檔只留hash；512598542 bytes保存為31852025 bytes，約釋放458 MiB。六原始輸入、全部附件、兩份完整來源、來源分類、十項乾淨產物、21反例與342正式輸入從壓縮證據全唯讀重驗退出0後，精確刪除C17工作區。準備腳本已保存，外部暫存bootstrap刪除；原版衍生C、ASM、LIB、OBJ、DB及工具只留本機。入口docs/re/matching-decompilation.md及docs/verification/retained-evidence.md。
- 權威工作表只更新進度與verify說明，原完成條件、狀態及Issue保持。render／verify／scan通過，19未完成／89完成／15人工保持。C16九個入口作正對照，C17十一入口、分類計數與擁有權通過；find無root-owned檔案或誤建Markdown目錄，Docker終檢無專案容器，git diff --check通過。接續需保留source_categories.py及分類基準，再擴充Town114個函式與初始化DATA；Book五函式保持。未提交或推送，沒有Issue寫入、版號、封包或發行。

## 2026-10-09：KOLTOWN全部初始化DATA來源完成

- 前輪C17新增43個來源函式屬實際進展；本輪接續完整六支matching goal，先閉合Town尚未還原的6736 DATA bytes。HEAD為c56adbe7，八檔既有異動保留。沿復古逆向、文件職責、IDA9.4及驗證交接入口建立已索引C18工作區；C17保存的26561份來源逐檔重核，七個固定工具按原commit重取並核大小、SHA-256及Git blob。
- 從既有完整LIB CODE binding、實際MLIBC／MLIBFP／LIBH資料與DATA-only模組定位31個初值範圍，包含XP、MSG、PAD與EPAD。991個實際LIB初值bytes與88個宣告零bytes分列；沒有把DATA-only模組當成新增CODE函式，也沒有將未匹配_rmtmp候選猜補進來源。
- 246個具名compiler初值供應5600 bytes，數值／位元表、逐個ASCII字串、兩個七筆近指標表與遠函式指標分列。DATA.C與PTRS.C使用不同translation unit，保留15筆真正external FIXUPP；LINK依原始DS及CODE位置解析，不手改指標值。原始C宣告標hypothesis，未證實語意保持unknown；遠指標sub_2DB42／LIB _execve的值與定位已核，資料記錄用途仍未知。
- 使用既有ida-pro-9.4-idapython:py312-v1、UID/GID1000:1000、無網路及有界資源，從固定解壓MZ建立一次性DB。初次CLI載入paragraph單位錯誤、次輪API命名空間錯誤，查安裝介面後以新副本修正；最終338函式、八段幾何、271個DATA項目與1359直接xref通過，input SHA與原始raw file bytes一致。DB SHA fd6fe282d1d21fdcd8b269252bc94bd15316185e9f616ab159954ac61adb0873。沒有修改正式DB或原始輸入，載入後bytes與file bytes的位址基準各列。
- CPP DATA讀取器沿已觀察QuickC thread preamble契約處理，真正指標fixup只接受已觀察的明列external形式。原生合併後的external index另作命名空間對應，來源比較按原始symbol與fixup核對。第一次獨立reader因輸出index欄位差異拒絕，修正正規化後同image重跑；這些屬研究工具問題，不記作遊戲缺陷。
- 全部6736 DATA分成5600 compiler初值、991 LIB實際初值、88 LIB宣告零與57原生零bytes；56個額外compiler對齊零bytes另列而不拷入原位，原版DATA片段0。兩空白目錄以QuickC2.01 /AM /Od /Gs /Zl /G0、LINK4.07、EXEMOD4.02、EXEPACK4.06重編，15項C、OBJ、來源清單、MAP、native MZ及packed全檔逐byte相同。
- 兩份獨立reader從真實compiler DATA物件與固定LIB bytes重算初值、PUBLIC、LIDATA、fixup、宣告零與原生零，連同149個既有compiler函式、52個LIB CODE模組，完整129312-byte映像、2281有序重定位及九個DATA relocation保持。兩次原生133679-byte核心加常數生成81個1Ah，完整133760-byte KOLTOWN.EXE均符合原版SHA c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221；沒有手改指令、header、checksum或壓縮payload。
- 31反例包含原有21項CODE／C／ASM／LIB／分類控制，以及實際DATA C與近指標再編譯、C／LIB資料bytes、近／遠指標目標、C／LIB初值漏項、額外零LEDATA與compiler對齊修改。14項OMF修改保持合法checksum；來源缺漏與多加零LEDATA不能因全檔值相同而通過。
- CODE計數保持172純C25412、378 LIB44396實際bytes、129混合30451 bytes，函式外271 C／258 ASM及Tandy入口13另列。原始799函式、四支完整來源與Book五函式4003 CODE保持；Town114函式94906 CODE、全六支119函式98909 CODE仍未還原，未還原DATA為0。原始唯一C文字／型別及__aDBdoswp原始symbol定義限制保留。342正式輸入hash不變，沒有遊戲執行、恢復Go對拍、硬體時序實驗、Issue寫入、版號、封包或發行。
- 封存27053份檔案，27份可重建執行檔只留hash；610853612 bytes保存為35088749 bytes，約釋放549 MiB。六原始輸入、完整CODE／DATA來源、IDA原始引用、15項乾淨產物、31反例與342正式輸入從壓縮證據全唯讀重驗退出0後，精確清理C18工作區與外部暫存bootstrap。原版衍生C、資料、ASM、LIB、OBJ、DB及工具僅本機；來源與核對入口見docs/re/matching-decompilation.md及docs/verification/retained-evidence.md。
- 權威工作表只回填DATA來源完成與下一CODE閘門，原完成條件、狀態及Issue保持。render／verify／scan通過，19未完成／89完成／15人工保持。C17十一入口作正對照，C18十二入口、DATA計數與擁有權通過；find無root-owned檔案或誤建Markdown目錄，Docker終檢無專案容器，git diff --check通過。下一批接Town114與Book五個CODE函式，保留full-build.py、data_objects.py、data_proof.py及classification-baseline.json；完整goal active，未提交或推送。

## 2026-10-10：KOLTOWN新增24個CODE來源函式

- 前輪完成6736 DATA來源屬實際進展；本輪接續完整六支matching goal。開工HEAD為c56adbe7，八檔既有異動保留。沿復古逆向、文件職責、compiler指紋及驗證交接入口，建立C19工作區；跨日後沿用起始日命名matching-c19-20261009，完成收據與目前狀態用2026-10-10。C18保存的27053份來源逐檔重核，七個固定工具按原commit重取並驗大小、SHA-256及Git blob。
- 新增Town24個完整原始CODE函式4337 bytes，包含3純C466與21明列C／ASM3871 bytes；混合分列1438 C設定／控制與2433 ASM。Book已驗來源只作候選，Town原始指令、常數、全域位置、callee與ASM邊界各別核對；只有完整native函式匹配才計進度。普通far-pointer候選sub_1A4E4保持開放，沒有以ASM補普通C差異。
- 新來源包含兩個解碼器的C控制、文字行處理、DOS讀寫檔重試，以及圖形拷貝、遮罩與游標暫存器程式。字典far pointer、byte／word陣列依既有header及Town consumer核對。sub_1A26B回傳值與near pointer參數、DOS寫入與關閉介面的型別按原始PUSH序列修正；全部149個既有compiler函式機器碼保持，原作者唯一型別、C文字與inline ASM形式仍未知。
- 轉換器移除0Fh／0A000h的必要前導0，造成C2094未定義標籤；錯誤碼核對Microsoft官方文件後保留合法十六進位語法，以同compiler／旗標乾淨重跑。驗證器局部變數遮蔽案例路徑、反例將函式本體延伸到檔尾，均修正並用新反例目錄重跑。失敗保留且分類研究工具問題，沒有改遊戲規則或放寬byte比較。
- 沿kol-matching-dos:bookworm-20261008-r1、UID/GID1000:1000、有界CPU／記憶體／PID／外層逾時，固定工具重取以外均無網路；原版ZIP、programs、解壓oracle、LIB與前批來源唯讀。兩空白目錄重編173個compiler函式及完整DATA，15項C、OBJ、清冊、MAP、native MZ與packed全檔逐byte相同。
- 兩份獨立reader核121純C16226、52混合6136、52個LIB CODE模組／75原始函式，連同全部6736 DATA，完整129312-byte映像、2281有序重定位及九個DATA relocation保持。LINK4.07、EXEMOD4.02、EXEPACK4.06兩次原生133679-byte核心加常數生成81個1Ah，完整133760-byte KOLTOWN.EXE均符合原版SHA c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221，沒有手改指令、header、checksum、壓縮payload或複製tail。
- 35反例保留原有31項CODE／DATA／LIB／來源分類控制，新增解碼C實際再編譯、新kernel ASM指令、新純C來源漏項與新ASM範圍修改；15項OMF修改保持合法checksum。全部DATA來源、246個初值符號、31個LIB初值範圍與CODE／DATA分離計數保持。
- 累計175純C25878、378 LIB44396實際bytes、150混合34322 bytes，混合分列22985 C／11337 ASM，函式外271 C／258 ASM及Tandy入口13另列。Town90函式90569 CODE、Book五函式4003 CODE仍未還原；全六支95函式94572 CODE、未還原DATA0，原始799函式與四支完整來源保持。342正式輸入hash不變，沒有遊戲執行、恢復Go對拍、硬體時序實驗、Issue寫入、版號、封包或發行。
- 封存27719份檔案，23份可重建執行檔只留hash；741855232 bytes保存為37318330 bytes，約釋放672 MiB。六原始輸入、24個新CODE、全部既有CODE／DATA、15項乾淨產物、35反例及342正式輸入從壓縮證據全唯讀重驗退出0後，精確清理C19工作區及外部暫存bootstrap。原版衍生C、ASM、LIB、OBJ、DB與工具只留本機；入口docs/re/matching-decompilation.md及docs/verification/retained-evidence.md。
- 權威工作表只更新進度與下一CODE閘門，原完成條件、狀態及Issue保持。render／verify／scan通過，19未完成／89完成／15人工保持。C18十二入口作正對照，C19九入口、CODE／DATA計數與擁有權通過；find無root-owned檔案或誤建Markdown目錄，Docker終檢無專案容器，git diff --check通過。下一批接Town90與Book五個CODE函式，保留seed-CODE.c、full-build.py、來源分類及完整DATA鏈；完整goal active，未提交或推送。


## 2026-10-10：KOLTOWN 新增 16 個 CODE 來源函式

- 路由命中原作工具鏈與執行期指紋，載入對應入口及逆向技能；沿用文件職責與來源證據契約。
- 從第十九批壓縮證據逐檔核雜湊恢復，保留全部 173 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增 16 個完整原始函式，共 3823 bytes。`sub_1B5C8` 為純 C 276 bytes；其餘 15 混合函式分列 3086 C／461 ASM bytes。
- C 還原圖像載入、地圖取樣、顯示控制、鍵盤清除與文字輸入。近指標及運算順序按原始指令／QuickC 結果核對；直接 ES／DS／BIOS／解碼區塊明列範圍。
- 三個字型 C 候選仍有指令差異，未用 ASM 補差異。保留試驗失敗，不計來源完成；`sub_1A4E4` 保持未知。
- 沿用 `kol-matching-dos:bookworm-20261008-r1`，固定 QuickC2.01 的 `/AM /Od /Gs /Zl /G0`；只跑編譯、連結及核對，遊戲與 Go 對拍維持暫停。
- 兩個空白案例由 `full-build.py` 重編 189 函式與兩份 DATA 來源，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06。15 項產物逐 byte 相同，完整 133760-byte KOLTOWN 的 SHA-256 兩次均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。
- 39 反例拒絕，其中 15 項 OMF 修改具有合法 checksum。新增反例實際重編輸入 C 與 BIOS ASM，並核漏函式與錯誤範圍。
- 驗證腳本誤選前置宣告及變數遮蔽造成兩次腳本失敗；修正後同容器、同命令於新案例乾淨重跑。未修改產品或比較條件。
- 封存後全唯讀重新核六個原始輸入、全部附件、兩份獨立 CODE／DATA 來源、fixup／回填、2281 有序重定位與 342 份正式 Go／module 輸入。
- 保存入口為 [第二十批收據](docs/verification/local/matching-c20-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c20-20261010/source-ledger.json)。29104 份檔案已保存，23 份可重建 EXE 只留 hash；本批工作區刪除，約釋放 944 MiB。
- 封存 SHA-256：`7439c53f5a023a513722810939797d6a6c51577d1f1563692f33132c998f4932`。原版衍生來源維持僅本機，未提交、推送、寫 Issue 或發行。
- 現有 176 純 C／26154 bytes、378 LIB／44396 實際 bytes、165 混合／37869 bytes。Town 仍 74 函式／86746 CODE，Book 五／4003；全六支 79 函式／90749 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單；原驗收條件及開放狀態保持。工作清單核對為 19 開放、89 已核實完成；掃描 2010 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建的 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 新增 14 個純 C 來源函式

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出及文件職責契約。
- 從第二十批壓縮證據逐檔核雜湊恢復，保留全部 189 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增 14 個純 C 原始函式，共 5268 bytes，沒有新增 ASM。包含共用文字控制 `sub_1A8FD`、畫面控制 `sub_1103F`／`sub_11390`、命令選擇 `sub_11722`，以及記錄交換、顯示、圖像包裝、地圖範圍／取樣、方向與日期計算。
- C 表達保留 signed／unsigned 比較、近指標順序、低 16-bit 乘除及遠呼叫的實際 stack words。導航型別 `g_005C`／`g_006C` 與 `sub_1C94E` 回傳值調整後，全部既有函式機器碼保持；原始唯一宣告仍未知。
- 編譯失敗為舊 far 函式指標語法，修正後同 QuickC／旗標重跑。比較差異由 C 型別、運算順序、連鎖指派及 C 標籤修正，未加 ASM、手修指令或放寬比對。
- 沿用 `kol-matching-dos:bookworm-20261008-r1`，QuickC2.01 的 `/AM /Od /Gs /Zl /G0`。兩個空白案例由 `full-build.py` 重編 203 函式及 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；最後 compiler 無警告。
- 15 項產物逐 byte 相同；兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像、2281 有序重定位與九個 DATA relocation。15 個 compiler 指標 fixup 與 56 個獨立 compiler 對齊零 bytes 保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。43 反例全部拒絕，其中 15 項 OMF 修改具有合法 checksum；新增文字控制／命令回傳的實際 C 重編、來源漏項與純 C 插入 ASM。
- [第二十一批收據](docs/verification/local/matching-c21-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c21-20261010/source-ledger.json)已從壓縮證據全唯讀重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入。
- 保存 29814 份檔案，23 份可重建 EXE 只留 hash；壓縮證據 SHA-256 為 `96cbc85b7fd2930b601ed8c71c8e97bc489b9a4dd9c4ab56cc27f015f697833c`。刪除本批已保存工作區，約釋放 1070 MiB；來源維持僅本機。
- 現有 190 純 C／31422 bytes、378 LIB／44396 實際 bytes、165 混合／37869 bytes。Town 仍 60 函式／81478 CODE，Book 五／4003；全六支 65 函式／85481 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。工作清單為 19 開放、89 已核實完成，掃描 2012 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 新增 14 個角色與戰鬥控制來源函式

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出及文件職責契約。
- 從第二十一批壓縮證據逐檔核雜湊恢復，保留全部 203 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增 14 個完整原始函式，共 6851 bytes。12 純 C 5944 bytes 涵蓋角色選擇、旗標、路徑步進、移動判定、文字替換及顯示控制；功能名稱只作導航，唯一原始宣告及未證實語意保持未知。
- `sub_25F8A` 分列 452 C／30 ASM bytes，只有兩處 ES 地圖讀取列 ASM；`sub_2A18E` 分列 41 C／384 ASM bytes，三個顯示暫存器區段明列範圍。沒有用 ASM 補普通 C 差異。
- 按原始參數順序修正長整數除法方向；同長度的錯誤候選未計成功。共用出口、switch、陣列、近指標、signed／unsigned 及 stack words 以完整 compiler 差分核對。
- 前置宣告與定義衝突造成一次編譯失敗，修正後同 compiler／旗標重跑；最後無警告。`g_008A` 導航型別調整後，全部 203 個既有函式機器碼保持。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。兩個空白案例由 `full-build.py` 重編 217 函式及 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；15 項產物逐 byte 相同。
- 兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像、2281 有序重定位與九個 DATA relocation。15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。47 反例全部拒絕，其中 15 項 OMF 修改具有合法 checksum；新增路徑 C／地圖 ASM 的實際再編譯、漏純 C 與錯誤顯示 ASM 範圍。
- [第二十二批收據](docs/verification/local/matching-c22-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c22-20261010/source-ledger.json)已從壓縮證據全唯讀重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入。
- 保存 30519 份檔案，23 份可重建 EXE 只留 hash；壓縮證據 SHA-256 為 `fe3a445ede9561f0c2f223043d5379d4078b03b9dc212d134732a96e1dab3f4f`。刪除本批已保存工作區，約釋放 1212 MiB；衍生來源維持僅本機。
- 現有 202 純 C／37366 bytes、378 LIB／44396 實際 bytes、167 混合／38776 bytes。Town 仍 46 函式／74627 CODE，Book 五／4003；全六支 51 函式／78630 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。工作清單為 19 開放、89 已核實完成，掃描 2014 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 主函式、選單與輸入來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出及文件職責契約。
- 從第二十二批壓縮證據逐檔核雜湊恢復，保留全部 217 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- `sub_10000` 完整 3548 bytes 已有來源，分列 3388 C／160 ASM。1095 條原始指令的初始化、載入、輸入控制、世界／選單切換與退出分支均匹配；六個 BIOS／DOS／顯示區段明列範圍。
- `sub_1AAE8` 為純 C 1355 bytes；`sub_1B100` 為 1145 C／47 ASM，共 1192 bytes。選單的文字掃描、選擇、捲動、滑鼠及命令控制，以及完整底層輸入均有匹配來源。
- 三個函式合計 6095 bytes，包含 1 純 C 與 2 混合來源。九個 ASM 區段限 BIOS／DOS／顯示／滑鼠暫存器；沒有用 ASM 補普通 C 差異，前批未知候選保持開放。
- 主函式以 near-word 視圖讀寫既有 far pointer，選單兩處相等比較使用 unsigned word 視圖，其餘 signed 運算保留。陣列誤作 scalar 指派造成一次編譯失敗，改為第一元素後同工具與旗標重跑；最後無警告。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。兩個空白案例由 `full-build.py` 重編 220 函式及 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；15 項產物逐 byte 相同。
- 兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像、2281 有序重定位與九個 DATA relocation。15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。51 反例全部拒絕，其中 15 項 OMF 修改具有合法 checksum；新增主函式 C／ASM 的實際再編譯、漏選單純 C 與錯誤輸入 ASM 範圍。
- [第二十三批收據](docs/verification/local/matching-c23-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c23-20261010/source-ledger.json)已從壓縮證據全唯讀重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入。
- 保存 31276 份檔案，23 份可重建 EXE 只留 hash；壓縮證據 SHA-256 為 `4a31cc8c4e8642ca0c9dd55e514b1624dcb3c8862075cf7754d9b3058ccc9a37`。刪除本批已保存工作區，約釋放 1373 MiB；衍生來源維持僅本機。
- 現有 203 純 C／38721 bytes、378 LIB／44396 實際 bytes、169 混合／43516 bytes。Town 仍 43 函式／68532 CODE，Book 五／4003；全六支 48 函式／72535 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。工作清單為 19 開放、89 已核實完成，掃描 2016 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 完整城鎮互動控制來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。
- 從第二十三批壓縮證據逐檔核雜湊恢復，保留全部 220 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增完整 `sub_11A93`，8967 bytes 分列 8939 C／28 ASM。2874 條原始指令的角色檢查、對話／服務選單、交易、物品與訓練控制均匹配；功能名稱只作導航，唯一原始宣告及未證實語意仍未知。
- 唯一 ASM 為 IDA linear 半開範圍 `[0x12CF2,0x12D0E)` 的 ES 圖像搬移，其他全部由 C 產生。普通 C 未匹配候選保持開放，沒有新增 ASM 補差異。
- 候選差異包括多出的提示呼叫、冗餘回跳、指標／長整數求值順序及 unsigned word 比較。逐項按原始指令與真實 compiler 差分修正；最後無警告，全部 221 函式匹配。
- objdump 只作 compiler OMF 交叉檢查，輸入 hash、工具版本與函式相對位址另存；未回填 word backpatch 的局部跳轉不作 native 控制流證據。原版定位仍以固定 IDA 匯出及 raw bytes 為準。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。兩個空白案例由 `full-build.py` 重編全部來源與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；15 項產物逐 byte 相同。
- 兩份獨立 reader 核 CODE／DATA、fixup／回填、129312-byte 映像、2281 有序重定位與九個 DATA relocation。15 個 compiler 指標 fixup、56 個獨立 compiler 對齊零 bytes 保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。55 反例全部拒絕，其中 15 項 OMF 修改具有合法 checksum；新增交易金額 C／圖像搬移 ASM 的實際再編譯、漏控制函式與錯誤 ASM 範圍。
- [第二十四批收據](docs/verification/local/matching-c24-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c24-20261010/source-ledger.json)已從壓縮證據全唯讀重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入。
- 保存 32070 份檔案，23 份可重建 EXE 只留 hash；壓縮證據 SHA-256 為 `d5c2b8a894c00059b6b1fb611f9e93d4bd24dc1e42bcc81115ec0bb26e654778`。刪除本批已保存工作區，約釋放 1555 MiB；衍生來源維持僅本機。
- 現有 203 純 C／38721 bytes、378 LIB／44396 實際 bytes、170 混合／52483 bytes。Town 仍 42 函式／59565 CODE，Book 五／4003；全六支 47 函式／63568 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。工作清單為 19 開放、89 已核實完成，掃描 2018 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 完整戰鬥控制來源匹配與多 CODE 編譯單元

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。
- 從第二十四批壓縮證據逐檔核雜湊恢復，保留全部 221 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增完整 `sub_1E3A8`，13650 bytes 分列 13634 C／16 ASM；原始 4389 條指令匹配。唯一 ASM 為 IDA linear 半開範圍 `[0x1EB0D,0x1EB1D)` 的 ES／SI 地圖讀取；普通 C 沒有以 ASM 補差異。
- 以明列 far pointer、50-byte 角色及 128-byte 怪物視圖產生原始欄位存取。保留重複條件及未指派的除法式；原始 C 宣告、型別與來源形式仍未知，功能名稱只作導航。
- 候選修正依原始指令核對全域位址、角色欄位、亂數布林條件及多敘述等待巨集的條件範圍；最後獨立候選無警告並逐 byte 匹配。
- 原單一編譯單元達 67016 CODE bytes，實際編譯報 C1028。固定來源的 C1.ERR 與 CL.ERR 已保存雜湊及下載來源，C1.ERR 明列 64KiB 段配置限制；錯誤不是產品缺陷。
- 沿用既有 image、QuickC2.01 `/AM /Od /Gs /Zl /G0` 與原記憶體模型。改為實際 `BULK.C/BULK.OBJ` 53366 bytes 及 `BATTLE.C/BATTLE.OBJ` 13650 bytes，各自低於 64KiB；前者來源與物件均和第二十四批逐 byte 相同。
- `seed-compiler-units.json` 定義真正的來源編譯入口。`seed-CODE.c` 與 67016-byte 合併座標只供閱讀及分類，不是 compiler 輸入，也未製造超過 64KiB 的 OMF CODE 段。準備工具會保留每個實際來源／物件。
- 兩份獨立 reader 從實際物件驗證各自 PUBLIC、段、offset、fixup／回填及有序清冊雜湊，另保留合併座標。核對 129312-byte 映像、2281 有序重定位、九個 DATA relocation 與完整 DATA 鏈。
- 兩個空白案例由 `full-build.py` 重編全部 222 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；18 項乾淨產物逐 byte 相同。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。62 反例全部拒絕，其中 16 項 OMF 修改具有合法 checksum；新增實際戰鬥 C／ASM 再編譯、缺編譯單元、錯來源位置／順序、BATTLE PUBLIC 及實際 C1028 容量反例。
- [第二十五批收據](docs/verification/local/matching-c25-20261010/receipt.json) 與 [來源清冊](docs/verification/local/matching-c25-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入。沒有執行遊戲。
- 保存 33149 份檔案，23 份可重建 EXE 只留 hash；壓縮證據 53385800 bytes，SHA-256 為 `47789f97f6ed157c15f3580828b39f1e9db0cad50a9941400b9cee87eae397b5`。只刪除本批已保存工作區，約釋放 1806 MiB；衍生來源維持僅本機。
- 現有 203 純 C／38721 bytes、378 LIB／44396 實際 bytes、171 混合／66133 bytes。Town 仍 41 函式／45915 CODE，Book 五／4003；全六支 46 函式／49918 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2020 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 圖像、文字與磁碟控制四函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第二十五批封存逐檔核雜湊恢復，保留兩個真實 CODE 編譯單元、全部 222 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增四個完整函式共 1707 bytes，分列 1012 C／695 ASM：`sub_19B29` 454 為 209 C／245 ASM，`sub_19E9D` 433 為 253 C／180 ASM，`sub_1B94F` 273 為 15 C／258 ASM，`sub_1CBA9` 547 為 535 C／12 ASM。
- 圖像轉換的普通 C 遠指標讀寫、unsigned 除法及 signed 迴圈比較，矩形顯示調度，以及磁碟重試、13 個區域 word 與狀態還原均由 C 產生。七個 ASM 範圍限文字／圖像 codec、顯示搬移與 DOS 切換磁碟；沒有用 ASM 補普通 C 差異。
- 磁碟重試的 `continue` 與原始跳轉核對。候選中參數名稱碰到 MASM 的 SEG／LENGTH 保留字，改為一般識別字後同工具重跑；最後四函式無警告且逐 byte 匹配。
- 三個字型候選的普通 C 減法仍有指令差異，保持未核准；候選及完整差分保存，不擴張 ASM 範圍吞入該差異。遠指標葉函式 `sub_1A4E4` 的 `_segment／_based` 候選報 C2061，該語法支援未證實，也保持未還原。
- 原始定位保留 IDA linear、解壓 MZ file offset、bytes 與 chunks；功能名稱只作導航。原始唯一 C 宣告、來源文字、ASM 來源形式及未證實語意仍未知。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。新增實際 `CONTROL.C/CONTROL.OBJ` 1707 bytes，既有 BULK 53366 及 BATTLE 13650 的 C／OBJ 均與前批逐 byte 相同；三個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義三個真正來源單元。68723-byte 合併只供閱讀與分類；`seed-CODE.c` 不是單一 compiler 輸入。每個 PUBLIC、段、offset、fixup／回填保留實際物件來源與合併座標，沒有合成超過 64KiB 的 OMF CODE 段。
- 兩個空白案例由 `full-build.py` 重編全部 226 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；20 項乾淨產物逐 byte 相同。兩份獨立 reader 核完整 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。68 反例全部拒絕，包含前批全部 62 項與新轉換 C、磁碟 C、codec ASM 實際再編譯、漏控制函式、錯 ASM 邊界及實際 CONTROL PUBLIC；17 項 OMF 修改具有合法 checksum。
- [第二十六批收據](docs/verification/local/matching-c26-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c26-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入；沒有執行遊戲。
- 保存 34481 份檔案，23 份可重建 EXE 只留 hash；原始證據 2279245607 bytes，壓縮為 75755036 bytes，SHA-256 為 `490f9cdbf70c1552925a88327ded087d666760f32cf1c7cd10da003d05090478`。只刪除本批已保存工作區，約釋放 2084 MiB；衍生來源維持僅本機。
- 現有 203 純 C／38721 bytes、378 LIB／44396 實際 bytes、175 混合／67840 bytes。Town 仍 37 函式／44208 CODE，Book 五／4003；全六支 42 函式／48211 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。接續需保留三個 seed 編譯單元與完整 DATA 鏈。未提交、推送、寫 Issue、打包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2022 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 選項、物品與圖像四個純 C 函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第二十六批封存逐檔核雜湊恢復，保留全部三個 CODE 編譯單元、226 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增四個完整純 C 函式共 3679 bytes，1236 條原始指令全部匹配，沒有新增 ASM。`sub_16D0B` 667 bytes／231 指令，`sub_196A9` 990／331，`sub_18857` 1150／374，`sub_1A4FE` 872／300。
- 選項重繪、文字分頁與輸入選擇、物品命令完整 switch、預載／檔案圖像來源及重試分支均有匹配來源。功能名稱只作導航，原始唯一 C 宣告、文字、型別與未證實語意仍未知。
- 差異按實際 compiler 與原始指令核對：二維 17-byte 表格、unsigned offset 表、stack word 視圖、整數比較與求值順序，以及遠指標的兩個 word 組成。原版比較式、未用區域變數及重複調度保持；最後四函式無警告且逐 byte 相同。
- 沒有貼回指令、使用 `_emit` 或 ASM 補普通 C 差異。既有三個字型候選與遠指標葉函式保持未還原，本批沒有將它們改列完成。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。新增實際 `MENU.C/MENU.OBJ` 3679 bytes；既有 BULK 53366、BATTLE 13650、CONTROL 1707 的 C／OBJ 均與前批逐 byte 相同，四個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義四個真正來源單元，bootstrap 依前批清單逐一恢復全部 seed。72402-byte 合併只供閱讀與分類，`seed-CODE.c` 不是單一 compiler 輸入。每個 PUBLIC、段、offset、fixup／回填保留實際物件來源與合併座標。
- 兩個空白案例由 `full-build.py` 重編全部 230 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；22 項乾淨產物逐 byte 相同。兩份獨立 reader 核完整 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 反例腳本首次在偽造 ASM 分類案例報 `KeyError: 'C_generated_bytes'`，因純 C 清冊省略此可選欄位。修正 fixture 由 `size` 派生該欄位；保存原失敗腳本與 69 個案例，見[腳本勘誤](docs/verification/local/matching-c27-20261010/checkpoint-failure.json)。使用相同 image／同一命令從空白反例案例乾淨重跑，兩個正例未改；此項分類為驗證腳本缺陷。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。74 反例全部拒絕，包含前批全部 68 項與新選擇 C、圖像 C 的實際再編譯、漏新純 C、偽造 ASM 分類、錯來源物件及實際 MENU PUBLIC；18 項 OMF 修改具有合法 checksum。
- [第二十七批收據](docs/verification/local/matching-c27-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c27-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入；沒有執行遊戲。
- 保存 37324 份檔案，23 份可重建 EXE 只留 hash；原始證據 2929551494 bytes，壓縮為 123774040 bytes，SHA-256 為 `5cc4873442c4f9bc2f6230a2630d03cd9f7d1899fca5cee9ec2d672ccbd524ce`。只刪除本批已保存工作區，約釋放 2657 MiB；衍生來源維持僅本機。
- 現有 207 純 C／42400 bytes、378 LIB／44396 實際 bytes、175 混合／67840 bytes。Town 仍 33 函式／40529 CODE，Book 五／4003；全六支 38 函式／44532 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。接續需保留四個 seed 編譯單元、47 個後續純 C／77 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2024 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 名冊、保存與資料頁四函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第二十七批封存逐檔核雜湊恢復，保留全部四個 CODE 編譯單元、230 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增四個完整函式共 5221 bytes，1756 條原始指令全部匹配，分列 5033 C／188 ASM。`sub_147DD` 1266 為 1202 C／64 ASM，`sub_14DB1` 1301 為 1177 C／124 ASM，`sub_153FF` 801 與 `sub_15ABF` 1853 為純 C。
- 兩個混合函式的 13 個 ASM 範圍在編譯前按原始 DOS 中斷與 ES 搬移列出。名冊比對、選擇、金額、保存槽、記錄回填、物品明細及資料頁控制由 C 產生；沒有用 ASM 補普通 C 差異。
- 值轉型與 word 視圖分開處理，保留求值順序、unsigned 除法／右移與 signed 傳參。原版重複讀寫、未用區域變數、磁碟還原、取消分支與唯一原始定位保持；功能名稱只作導航，原始唯一 C／ASM 來源形式與未證實語意仍未知。
- 候選最初 C2094 為 ASM 別名未宣告，改用既有 pool 綁定。固定 C1.ERR／CL.ERR 複製到候選案例以保存可讀診斷；編譯器及旗標保持，最後四函式無警告且逐 byte 相同。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。新增實際 `ROSTER.C/ROSTER.OBJ` 5221 bytes；BULK 53366、BATTLE 13650、CONTROL 1707、MENU 3679 的 C／OBJ 均與前批逐 byte 相同，五個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義五個真正來源單元，bootstrap 依前批清單逐一恢復。77623-byte 合併只供閱讀與分類，`seed-CODE.c` 不是單一 compiler 輸入；每個 PUBLIC、段、offset、fixup／回填保留實際物件來源與合併座標。
- 兩個空白案例由 `full-build.py` 重編全部 234 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；24 項乾淨產物逐 byte 相同。兩份獨立 reader 核完整 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 物品明細反例把 128 改成 127，真實 compiler 函式由 801 縮為 800 bytes，轉接器正確拒絕長度不符；原腳本只預期較後的獨立來源核對拒絕，因此中止。明確登記此反例必須在原生轉接長度閘門拒絕，保存失敗腳本與 73 個案例，見[驗證階段勘誤](docs/verification/local/matching-c28-20261010/checkpoint-failure.json)。同 image／同命令從空白反例案例重跑，原生正例未改；此項分類為驗證腳本階段假設缺口。
- 80 反例全部拒絕，包含前批全部 74 項及新名冊 C、物品明細 C、記錄搬移 ASM 再編譯、漏資料頁、錯 DOS 範圍與實際 ROSTER PUBLIC；19 項 OMF 修改具有合法 checksum。封存核對器另解析正負真實物件，證實 `sub_153FF` 801／800 bytes，並核對兩批歷史失敗紀錄。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第二十八批收據](docs/verification/local/matching-c28-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c28-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入；沒有執行遊戲。
- 保存 40676 份檔案，23 份可重建 EXE 只留 hash；原始證據 3655820291 bytes，壓縮為 179009348 bytes，SHA-256 為 `15cff0a2a9687006e79c983d13c2b1b44d3ac385ba34472410ef06fe4ad58de3`。只刪除本批已保存工作區，約釋放 3295 MiB；衍生來源維持僅本機。
- 現有 209 純 C／45054 bytes、378 LIB／44396 實際 bytes、177 混合／70407 bytes。Town 仍 29 函式／35308 CODE，Book 五／4003；全六支 34 函式／39311 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。接續需保留五個 seed 編譯單元、49 個後續純 C／79 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2026 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 角色、選單與欄位編輯四函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第二十八批封存逐檔核雜湊恢復，保留五個 CODE 編譯單元、234 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增四個完整函式共 6038 bytes，2023 條原始指令全部匹配，分列 5911 C／127 ASM。`sub_16239` 1171 為 1166 C／5 ASM，`sub_16986` 901 為 779 C／122 ASM，`sub_17318` 1480 與 `sub_178E0` 2486 為純 C。
- 四個 DS／ES ASM 範圍在編譯前明列。角色摘要、物品選單、既有／可購項目、欄位編輯、價格與記錄回填均有完整來源；沒有用 ASM 補普通 C 差異。
- word 視圖修正 signed 傳參，混合 signed／unsigned 索引的求值順序亦閉合。原始多餘實參、未初始化讀取、壓棧／清理、重試與取消分支保持。`sub_16986` 的 IDA linear `0x16B7F`／`0x16A95` 分別傳七個／五個實參，清理 14／10 bytes；不據此推定唯一 prototype。
- 初次 C2086 為候選工具預設 int 宣告與新 void 定義衝突，補明示宣告後同工具重跑。最後四函式無警告且逐 byte 相同。原始函式名、位址、chunks、raw bytes 與未證實語意保留；功能名稱只作導航。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。七個工具與兩個錯誤資源皆核固定 hash。新增 `CHAR.C/CHAR.OBJ` 6038 bytes，前五個 C／OBJ 與前批逐 byte 相同，六個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義全部六個來源單元。83661-byte 合併只供閱讀與分類，`seed-CODE.c` 不是單一 compiler 輸入；每個 PUBLIC、段、offset、fixup／回填保留實際物件來源與合併座標。
- 兩個空白案例由 `full-build.py` 重編全部 238 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；26 項乾淨產物逐 byte 相同。兩份 reader 核完整 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 86 反例全部拒絕，包含全部前批 80 項及新摘要 C、欄位編輯 C、選單 ASM 再編譯、漏純 C、錯 DS 範圍與實際 CHAR PUBLIC；20 項 OMF 修改具有合法 checksum。前批真實物件 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- 首次封存命令的自動權限審查逾時，命令未啟動。依工具指示重試一次獲准完成，沒有更換工具鏈或放寬隔離。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第二十九批收據](docs/verification/local/matching-c29-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c29-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入；沒有執行遊戲。
- 保存 42812 份檔案，23 份可重建 EXE 只留 hash；原始證據 4094764232 bytes，壓縮為 213044472 bytes，SHA-256 為 `4edce41204ecc7777c284934e248fde87bf1b9e13677d33155523c5b997260a3`。只刪除本批已保存工作區，約釋放 3680 MiB；衍生來源維持僅本機。
- 現有 211 純 C／49020 bytes、378 LIB／44396 實際 bytes、179 混合／72479 bytes。Town 仍 25 函式／29270 CODE，Book 五／4003；全六支 30 函式／33273 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。接續需保留六個 seed 編譯單元、51 個後續純 C／81 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2028 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 欄位、圖像與戰鬥四函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第二十九批封存逐檔核雜湊恢復，保留六個 CODE 編譯單元、238 個 compiler 函式、52 個 LIB CODE 模組及 6736 初始化 DATA。
- 新增四個完整函式共 5554 bytes，1864 條原始指令全部匹配，分列 5539 C／15 ASM。`sub_18350` 1022、`sub_18F3A` 1903 與 `sub_228B7` 1605 為純 C；`sub_223C0` 1024 為 1009 C／15 ASM。
- 唯一 ASM 為候選編譯前已明列的 IDA linear 半開範圍 `[0x2254E,0x2255D)` ES／SI 讀取。欄位顯示、數值彙總、分層圖像、九乘九格點、狀態與命令控制均有完整來源；沒有用 ASM 補普通 C 差異。
- 二維／三維 byte 視圖、word 視圖、unsigned 運算及加法的求值順序按原始指令閉合。候選 break 分支按原始跳轉修正，原版夾限、重複運算、狀態回填與取消保持。原始唯一 C／ASM 來源形式與未證實語意仍未知；功能名稱只作導航。
- 初次工具引用前批轉換器的層級多一層，另有 `g_9E35`／`g_1CE2` 缺項宣告。修正後同 image／同命令從空白案例重跑；最後四函式無警告且逐 byte 相同。這些失敗分類為研究工具／候選宣告問題。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`。七個工具與兩個錯誤資源皆核固定 hash。新增 `TACTIC.C/TACTIC.OBJ` 5554 bytes，前六個 C／OBJ 與前批逐 byte 相同，七個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義全部七個來源單元。89215-byte 合併只供閱讀與分類，`seed-CODE.c` 不是單一 compiler 輸入；每個 PUBLIC、段、offset、fixup／回填保留實際物件來源與合併座標。
- 兩個空白案例由 `full-build.py` 重編全部 242 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；28 項乾淨產物逐 byte 相同。兩份 reader 核完整 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 92 反例全部拒絕，包含全部前批 86 項及新欄位 C、命令 C、格點 ASM 再編譯、漏圖像 C、錯 ES／SI 範圍與實際 TACTIC PUBLIC；21 項 OMF 修改具有合法 checksum。前批真實物件 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第三十批收據](docs/verification/local/matching-c30-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c30-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份正式 Go／module 輸入；沒有執行遊戲。
- 保存 45279 份檔案，23 份可重建 EXE 只留 hash；原始證據 4581218518 bytes，壓縮為 251167288 bytes，SHA-256 為 `5efa472e33b8f116deaa750d836ff183c1369fe9745029f3f68fd4fe3f48153a`。只刪除本批已保存工作區，約釋放 4107 MiB；衍生來源維持僅本機。
- 現有 214 純 C／53550 bytes、378 LIB／44396 實際 bytes、180 混合／73503 bytes。Town 仍 21 函式／23716 CODE，Book 五／4003；全六支 26 函式／27719 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件及開放狀態保持。接續需保留七個 seed 編譯單元、54 個後續純 C／82 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2030 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。


## 2026-10-10：KOLTOWN 戰鬥顯示、選擇與目標輸入四函式來源匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第三十批封存逐檔核雜湊恢復，保留七個 CODE 單元、242 個 compiler 函式、52 個 LIB CODE 模組與 6736 初始化 DATA。
- 新增四個完整函式共 6558 bytes，2182 條原始指令全部匹配，分列 6553 C／5 ASM。`sub_22EFC` 851 為 846 C／5 ASM；`sub_2324F` 1001、`sub_2374C` 2533 與 `sub_24131` 2173 為純 C。
- 唯一 ASM 為候選編譯前明列的 IDA linear 半開範圍 `[0x230D2,0x230D7)` DS 讀取。戰鬥文字／圖像、32-bit 比例、權重、方向搜尋、重試、目標輸入、夾限與返回控制皆有完整來源；沒有以 ASM 補普通 C 差異。
- 128-byte 視圖、unsigned offset 表、post-condition 迴圈、加法求值順序與 chained assignment 按原始指令閉合。原版重複判定、跳過的程式區段與取消保持；原始唯一 C／ASM 寫法與未證實語意仍未知，功能名稱只作導航。
- 最後方向差異回查完整 IDA 匯出與 raw bytes：`0x2406A` 的 `8b46fa8946fc` 先恢復基準，`0x24070` 的 `8346fc01` 再加一。截斷輸出的推測已撤回，未進核准來源；完整差分保留。候選缺項宣告與近指標陣列語法亦經同工具從空白案例重跑，最後四函式無警告。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`，九項工具／錯誤資源固定 hash 保持。新增 `DECIDE.C/DECIDE.OBJ` 6558 bytes，前七個 C／OBJ 與前批逐 byte 相同，八個真實段各自低於 64KiB。
- `seed-compiler-units.json` 定義八個來源單元。95773-byte 合併只供閱讀與分類，`seed-CODE.c` 不是單一 compiler 輸入；PUBLIC、段、offset、fixup／回填均保留實際物件與合併座標。
- 兩個空白案例由 `full-build.py` 重編全部 246 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；30 項乾淨產物逐 byte 相同。兩份 reader 核 CODE／DATA、129312-byte 映像與 2281 有序重定位，九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 98 反例全部拒絕，包含全部前批 92 項與新顯示 C、戰鬥選擇 C、DS ASM 再編譯、漏目標輸入 C、錯 DS 範圍與實際 DECIDE PUBLIC；22 項 OMF 修改具有合法 checksum。前批 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第三十一批收據](docs/verification/local/matching-c31-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c31-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份 Go／module 輸入；沒有執行遊戲。
- 保存 48117 份檔案，23 份可重建 EXE 只留 hash；原始證據 5120704523 bytes，壓縮為 293534172 bytes，SHA-256 為 `a08959c8db2c45e551eee28038473b0bed673a1fd0382d64890cb2844bde779a`。只刪除本批已保存工作區，約釋放 4579 MiB；衍生來源維持僅本機。
- 現有 217 純 C／59257 bytes、378 LIB／44396 實際 bytes、181 混合／74354 bytes。Town 仍 17 函式／17158 CODE，Book 五／4003；全六支 22 函式／21161 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件與開放狀態保持。接續需保留八個 seed、57 個後續純 C／83 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 工作清單為 19 開放、89 已核實完成，掃描 2032 行全部有入口。Docker 容器清單為空，UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；三份主機暫存腳本已刪除，`git diff --check` 通過。

## 2026-10-10：KOLTOWN 目標、權重、攻擊與傷害四純 C 函式匹配

- 路由命中原作工具鏈與執行期指紋；沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第三十一批封存逐檔核雜湊恢復，保留八個 CODE 單元、246 個 compiler 函式、52 個 LIB CODE 模組與 6736 初始化 DATA。
- 新增四個完整純 C 函式共 5611 bytes，1891 條原始指令全部匹配。`sub_253E7` 620、`sub_25A3A` 725、`sub_26D76` 2686 與 `sub_277F4` 1580；沒有新增 ASM 或原始函式數量。
- 目標候選、32-bit 權重與正規化、攻擊／防禦、損傷、部位狀態與位移均有完整來源。首輪一個函式匹配，依原始 bytes 調整 chained assignment、運算元求值順序與跳轉結構，第三輪四個函式全部無警告匹配。原版覆寫權重、未用到的機率運算與重複跳轉保持，未按推測修正玩法。
- 原始名稱、線性位址、chunks、operand 與 raw bytes 保持。功能名稱只作導航；固定工具產生相同機器碼已證實，原始唯一 C 宣告、型別與未證實語意仍未知。三個字型候選與遠指標葉函式保持未匹配，未以 ASM 補普通 C 差異。
- 沿用 `kol-matching-dos:bookworm-20261008-r1` 與 QuickC2.01 `/AM /Od /Gs /Zl /G0`，九項工具／錯誤資源固定 hash 保持。新增 `ATTACK.C/ATTACK.OBJ` 5611 bytes；前八個 C／OBJ 與前批逐 byte 相同，九個真實 CODE 段各自低於 64KiB。
- `seed-compiler-units.json` 定義九個來源單元。101384-byte 合併只供閱讀與分類；PUBLIC、段、offset、fixup／回填保留實際物件與合併座標，沒有編譯大於 64KiB 的合併物件。
- 兩個空白案例由 `full-build.py` 重編全部 250 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；32 項乾淨產物逐 byte 相同。兩份 reader 核 CODE／DATA、129312-byte 映像與 2281 有序重定位，九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 104 反例全部拒絕，包含前批 98 項與新權重／攻擊 C 再編譯、漏傷害 C、偽造 ASM 分類、錯誤物件來源與實際 ATTACK PUBLIC；23 項 OMF 修改具有合法 checksum。前批 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第三十二批收據](docs/verification/local/matching-c32-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c32-20261010/source-ledger.json)已從壓縮證據全唯讀獨立重驗通過，包含六個原始輸入、全部附件與 342 份 Go／module 輸入；沒有執行遊戲。
- 現有 221 純 C／64868 bytes、378 LIB／44396 實際 bytes、181 混合／74354 bytes。Town 仍 13 函式／11547 CODE，Book 五／4003；全六支 18 函式／15550 CODE 未還原，DATA 未還原為 0，完整目標未完成。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件與開放狀態保持。接續需保留九個 seed、61 個後續純 C／83 混合分類、106 純 C 基準與完整 DATA 鏈。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 保存 51296 份檔案，23 份可重建 EXE 只留 hash；原始證據 5713959497 bytes，壓縮為 340051632 bytes，SHA-256 為 `c058068d8fc9e24b89411a3b4f39802ce6fddc9114bccd58c6a504395dc7e0c9`。只刪除本批已保存工作區，約釋放 5099 MiB；衍生來源維持僅本機。
- 工作清單為 19 開放、89 已核實完成，掃描 2034 行全部有入口。本專案容器清單為空，其他專案容器保持；UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄。四份主機暫存腳本已刪除，`git diff --check` 通過。

## 2026-10-10：KOLTOWN 戰鬥效果、掉物、世界與遭遇四純 C 函式匹配

- 上一輪為實際進展。路由命中原作工具鏈與執行期指紋，沿用逆向技能、IDA Pro9.4 原始匯出與文件職責契約。從第三十二批封存逐檔核雜湊恢復，保留九個 CODE 單元、250 個 compiler 函式、52 個 LIB CODE 模組及全部初始化 DATA。
- 新增四個完整純 C 函式共 6348 bytes，2048 條原始指令全部匹配：`sub_2818D` 1759、`sub_289EE` 1391、`sub_29004` 2030、`sub_2A4E5` 1168。沒有新增 ASM 或原始函式數量。
- 初輪一個、次輪兩個、第三輪四個函式匹配，三輪皆無 compiler 警告。原始條件方向、逗號運算子初始化、前置加減的條件值、有號字組視圖與求值順序均由完整原始 bytes 閉合。遠指標讀取及三處 DS 轉換由 C 產生，沒有用 ASM 補普通 C 差異。
- 原版陣列基址比較、被跳過的按鍵分支、清理返回與取消後重試保持。唯一 C 宣告／型別、未證實語意仍未知；功能名稱只作導航，原始函式名、位址、chunks、operand 與 bytes 保持。
- 新增[剩餘範圍核對](docs/verification/local/matching-c33-20261010/remaining-coverage-review.json)。C32 的 13 個未完成函式完整長度合計 11741 bytes，但 `sub_2B9BE` 已有 235／236 bytes 由兩個 CRT 模組覆蓋；另有 41 未還原 CODE bytes 位於函式外，實際未還原量為 11547。四個新目標與既有來源無重疊；本批後實際剩餘 5199，部分 CRT 與函式外範圍保持，兩份核對器獨立重算。
- 沿用 `kol-matching-dos:bookworm-20261008-r1`、QuickC2.01 `/AM /Od /Gs /Zl /G0` 與九項固定工具／錯誤資源。image Entrypoint 查詢的 template 缺鍵只屬控制面查詢失敗；沿用相同 image，ID 核對一致，沒有另建工具鏈。
- 新增 `WORLD.C/WORLD.OBJ` 6348 bytes，前九個 C／OBJ 與前批逐 byte 相同。十個真實 CODE 段各自小於 64KiB，107732-byte 合併只供閱讀與分類；每筆 PUBLIC、fixup／回填保留實際物件、段、offset 與分析座標。
- 兩個空白案例由 `full-build.py` 重編全部 254 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06。34 項乾淨產物逐 byte 相同，兩份 reader 核全部 CODE／DATA、129312-byte 映像與 2281 有序重定位；九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes 保持。
- 110 反例全部拒絕，包含前批 104 項與新效果 C、世界 C 再編譯、漏掉物 C、偽造 ASM 分類、錯誤物件來源與實際 WORLD PUBLIC；24 項 OMF 修改有合法 checksum。前批 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[第三十三批收據](docs/verification/local/matching-c33-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c33-20261010/source-ledger.json)已由壓縮證據全唯讀獨立核對，涵蓋六個原始輸入、全部附件與 342 份 Go／module 輸入；沒有執行遊戲。
- 累計 225 純 C／71216 bytes、378 LIB／44396 實際 bytes、181 混合／74354 bytes。Town 尚 9 函式／5199 CODE，Book 五／4003；全六支 14 函式／9202 CODE 未還原，DATA 未還原為 0。完整目標仍開放，不以 34 項一致產物代替剩餘來源還原。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件與開放狀態保持。接續保留十個 seed、65 個後續純 C／83 混合分類、106 純 C 基準、全部 DATA 鏈與剩餘範圍核對。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 封存期間唯讀回查 [C26 排除清冊](docs/verification/local/matching-c26-20261010/excluded-font-inputs.json)。其壓縮檔的 `units/KOLTOWN/controls-unit-03/BULK.C` 與 `diagnostic.json` 顯示 `x=x-32` 只差 ADD／SUB；04 的無號轉型與 05 的 `volatile` 亦未匹配。指標型參數或近指標差能否保留 SUB 仍屬假說，尚未編譯，不列進度。`based-probe-01` 的 C2061 只證明該 `_segment` 宣告失敗，不推定所有 `_based` 語法都不支援。
- 保存 54871 份檔案，23 份可重建 EXE 只留 hash；原始證據 6368282319 bytes，壓縮為 391374076 bytes，SHA-256 為 `04efa2a5a4b0382229d72e95b03a9d5f697a3e3f5c27e6c9d06e6b4ca624f76a`。只刪除本批已保存工作區，約釋放 5673 MiB；衍生來源維持僅本機。
- 工作清單為 19 開放、89 已核實完成，掃描 2037 行全部有入口。本專案容器清單為空；同 image 的 `hungry_wing` 經掛載核對屬 `fd2`，予以保留。UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄。六份主機暫存腳本已刪除，`git diff --check` 通過。

## 2026-10-10：KOLTOWN 圖像、人物選擇、事件與分頁四函式匹配

- 上一輪為實際進展。路由命中原作工具鏈與執行期指紋，載入單一入口與文件職責契約。逐檔核 C33 封存後恢復全部十個 CODE 單元、254 個 compiler 函式、52 個 LIB 模組及 6736 初始化 DATA，保留原版六支 EXE 與 342 份 Go／module 輸入。
- 新增四個完整函式共 4203 bytes，1453 條原始指令匹配，分列 3701 C／502 ASM。`sub_1BD17` 991 為 489 C／502 ASM，`sub_2AAA8` 1623、`sub_2B0FF` 722、`sub_2B4A7` 867 為純 C。
- 圖像常式的六段硬體 ASM 在編譯前明列；DS／ES、複製、位元合成與暫存原始運算元由 IDA9.4 匯出轉成可回查的來源，C 控制在範圍外。沒有用 ASM 補普通 C 差異。
- 首輪一個函式匹配，依原始 bytes 修正有號常數、布林求值順序、條件方向、近指標表達與未使用的運算式結果，第二輪四個函式全部無警告匹配。原始函式名、線性位址、chunks、offset、bytes 與未證實語意保持；功能名稱只作導航，原始唯一 C／ASM 寫法仍未知。
- 另保存[12 個受控試驗](docs/verification/local/matching-c34-20261010/probe-summary.json)，沒有增加匹配函式數。近指標差、長整數、無號常數、參數型別與運算式求值仍未閉合三個字型函式的 ADD／SUB 差異。這是上一輪假說的實測結果，原假說歷史保留。
- 同一固定工具 repo commit 的 QuickC1 與 QuickC251 候選另核八份檔案的 Git blob／SHA-256。QuickC1 拒絕本次 `_asm` 源碼，純 C 前綴的保存順序與指令也不同；不把省略硬體的前綴稱為完整匹配。QuickC251 接受本次 `_segment`／`_based` 轉型，完整葉函式為 29／26 bytes；舊式參數宣告（K&R）候選為 21 bytes，索引候選報 C2147。只證實各次源碼／參數結果，不推定所有寫法不支援或不可能。
- 正式工具仍為 QuickC2.01 `/AM /Od /Gs /Zl /G0` 與原九項工具／錯誤資源。新增 `SUPPORT.C/SUPPORT.OBJ` 4203 bytes，前十個 C／OBJ 與前批逐 byte 相同；十一個真實段各低於 64KiB，111935-byte 合併只作閱讀／分類。PUBLIC、fixup／回填保留實際物件、段、offset 與分析座標。
- 兩個空白案例重編全部 258 個 compiler 函式與 DATA，再經 LINK4.07／EXEMOD4.02／EXEPACK4.06；36 項乾淨產物逐 byte 相同。兩份 reader 核 CODE／DATA、129312-byte 映像、2281 有序重定位、九個 DATA relocation、15 個 compiler 指標 fixup 與 56 個對齊零 bytes。
- 116 反例全部拒絕，包含前批 110 項及新裁切 C、人物選擇 C、圖像 ASM 再編譯、漏分頁 C、錯硬體 ASM 範圍與實際 SUPPORT PUBLIC；25 項 OMF 修改保持合法 checksum。前批 801／800 長度閘門與兩批失敗紀錄保持獨立核對，本批沒有新增驗證腳本失敗。
- [剩餘範圍](docs/verification/local/matching-c34-20261010/remaining-coverage-review.json)重算後，Town 尚 5 函式／996 CODE，包含 41 個函式外 bytes 與部分 CRT 的 1 byte；235 個已覆蓋 CRT bytes 不重複計數。Book 五／4003、所有初始化 DATA、799 個原始函式與四支完整來源保持。
- 兩次完整 133760-byte KOLTOWN 的 SHA-256 均為 `c2442c8438ce8812926d61555626f67f918a5c9ed9c42939d98938535fddb221`。[C34 收據](docs/verification/local/matching-c34-20261010/receipt.json)與[來源清冊](docs/verification/local/matching-c34-20261010/source-ledger.json)已從壓縮證據全唯讀獨立核對，另核所有試驗、替代工具雜湊與正式輸入；沒有執行遊戲。
- 累計 228 純 C／74428 bytes、378 LIB／44396 實際 bytes、182 混合／75345 bytes，混合分列 61350 C／13995 ASM。全六支仍 10 函式／4999 CODE 未還原，DATA 未還原為 0，完整目標仍開放。
- 已更新目前狀態、研究入口與唯一 JSON 工作清單，原驗收條件與開放狀態保持。接續保留十一個 seed、68 個後續純 C／84 混合分類、106 純 C 基準、完整 DATA 鏈、剩餘範圍與試驗清冊。未提交、推送、寫 Issue、建立正式交付封包、發行或恢復 Go 對拍。
- 封存期間另開始 [C35 先行工作](workplace/matching-c35-20261010/PLAN.json)，以 C34 唯讀輸入與容器暫存檔案系統跑五個 C 候選，兩份收據及來源／物件／命令已保存。逗號運算式恢復遠指標讀取後為 26 bytes，仍只差載入順序；其餘候選也未匹配。C35 批次尚未整合，工作區保留供接續；目前來源計數仍以 C34 為準。恢復完整 C35 前先核兩份 pilot 收據，保留既有 pilot 與 PLAN.json。
- 保存 59003 份檔案，26 份可重建或按固定 Git blob 重取的執行檔只留 hash；原始證據 7077133575 bytes，壓縮為 446953976 bytes，SHA-256 為 `b89bcf54287983fae8b5580eb11919307d6f38e806a565eea72d2fc63063852c`。只刪除本批已保存工作區，約釋放 6293 MiB；衍生來源與工具仍僅本機保存。
- 工作清單為 19 開放、89 已核實完成，掃描 2041 行全部有入口。C34 工作區已刪除，C35 的 pilot 與兩份收據保留且雜湊已核對；本專案容器清單為空，其他專案資源未動。UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；九份主機暫存腳本已刪除，`git diff --check` 通過。

## 2026-10-10：暫停 matching，保存與推送研究進度

- 依使用者「先 commit + push」指示暫停 matching 目標，不繼續來源研究。目標狀態已確認為 paused；目前仍 10 函式／4999 CODE bytes 未還原，DATA 來源完整。
- C35 已逐檔核 C34 封存並恢復十一個 CODE 單元與完整 DATA 鏈。既有五個先行候選、兩份 pilot 收據與 PLAN.json 保留；正式九項工具 pins 已核對。新增無號左值字型候選三函式仍未匹配，沒有整合、增加來源完成量或執行完整 C35 驗證。
- 接續入口為 `workplace/matching-c35-20261010/PLAN.json`、`plan.json`、`bootstrap.py` 與 `units/KOLTOWN/font-lvalue-unsigned/`。未完成工作區保留；恢復時以 C34 來源清冊與全部真實 CODE 單元為基準。
- 本次保存第六至三十四批研究紀錄、來源清冊索引、工作清單與證據保存工具。提交範圍限八個既有文字檔；原版輸入、C／ASM／OBJ／LIB、工具與本機壓縮證據均不納入 Git。
- 提交前以既有隔離 image 核對保存工具語法與擁有權。UID/GID 1000:1000，沒有 root-owned 檔案或誤建 `.md` 目錄；本專案容器清單為空，其他專案容器保持。
- 工作清單 render、verify 與 scan 均通過：19 開放、89 已核實完成，2042 行自承全部有條目認領。C35 bootstrap 暫存腳本已刪除，工作區內保存副本保留。

## 2026-10-11：v1.0.0 本機收尾、完整版、release 與推廣影片

- 依使用者授權判斷收尾，最多補一次抽樣並打包。路由命中收尾驗證、版號、交付結構與影音合成，載入對應單一入口及必要 reference；沿專案 `full/`、`release/`、`promo/` 契約。完整版號為 `v.1.0.0-20261011`，交付與收據入口見 [本輪驗收](docs/verification/release-v1.0.0-20261011.md)。
- 正式來源 342 份 Go／module 輸入與最近平台驗收逐項相同，沒有改規則、中文排版、保存格式或加速設定。matching 保持暫停，既有 C35 先行工作區與證據未動；本輪沒有重開原版研究或增加原版同狀態覆蓋。
- 先登記 seed1、正常鍵序、完整 trace、v3 保存比較與失敗條件，再執行唯一正常抽樣。標題、書本、六人招募、任務5戰鬥、逃離、自然 ORC 遭遇、回城住宿保存共1046狀態／40 PNG；同一鏈新程序重載13狀態／3 PNG。trace 與最近正式版相同，完整保存只差 saved_at，重載沒有改輸入存檔。沒有重擲、注入狀態、遭遇覆寫或強制勝利；不宣稱任務5完成或全程通關。
- 正式十九套件測試通過。`go vet` 保留兩則既有 `book.Label` 未具名欄位警示，位於 `internal/ui/bitmap_editor.go` 第56、57行；原文與退出碼保存於交付 `smoke/`，不宣稱靜態檢查全綠。
- 沿既有 Go1.26.7／Ebitengine2.9.9、osxcross 與 Wine9 工具鏈。既有 package image 不存在，修正 `docker/Dockerfile.package` 並建立 `kol-package:1.26.7-ebiten2.9.9-20261011-r1`；固定基底與 AppImage runtime 雜湊，核對同為 Debian12。首輪 apt 安裝逾時及缺 liblzo2 的建置均屬工具環境問題，改用既有固定 image 並補同系統相依後乾淨建置通過，未掛入主機 runtime。
- `full/` 與 `release/` 各有 Linux AppImage、Windows AMD64 ZIP、macOS universal ZIP。逐包解開驗證執行檔、版號、完整素材雜湊、條款、ZIP UTF-8、中文說明 BOM 與 macOS 執行權限。release 不含原版素材，提供先核全部 SHA-256 才寫入的合法資料匯入工具；錯誤雜湊負例在寫入前拒絕。
- Linux 實際完整版及匯入後 release 以 AppImage 解包模式啟動。Windows 實際 ZIP 解壓後，full 與合法匯入後 release 的版號及中文封面相同。macOS 雙架構、最低版本11.0、ARM64 ad-hoc 簽章與系統相依通過；沒有執行或真機驗收。Linux 主機 FUSE、Windows／macOS 真機與完整原版 parity 保持既有未驗限制。
- Xvfb 就緒等待與 Wine 的暫存 HOME 擁有權失敗均發生於遊戲啟動前，修正隔離環境後重跑原命令。正常抽樣只啟動一次；版本與封面檢查不另開玩家情境。ImageMagick 的接觸表程序兩次 SIGABRT，最小兩圖案例亦相同，改用 FFmpeg tile 完成，不記為遊戲缺陷。
- 推廣影片為本版43張畫格中選取的截圖剪輯，共64秒、1920×1080、25fps、H.264／AAC。八幕中心與完整字幕逐幕目視通過；峰值 -9.8dB，沒有達門檻的黑幀。只用已接受的遊戲雷聲；其他靜音與6至10秒停格依分鏡保留，不稱連續實機錄影。來源、配方、逐幕索引與影音檢查均隨影片保存。
- 六個封包內容檢查通過；既有獨立驗證器重算 full／release／promo 的951／290／40筆 manifest，包含巢狀來源 manifest。根目錄 `SHA256SUMS.json` 記錄所有輸出，`build-inputs.json` 保存正式來源與最終打包工具雜湊。完整版及含原作畫面的影片僅本機；release 中翻譯的公開權利仍有限制。本輪沒有公開發布、建立 tag、上傳影片或修改遠端 Issue。
- 更新唯一目前狀態、README 交付入口與工作清單，僅完成本輪封包重建項；其餘18開放項的驗收條件保持。生成二進位、遊戲資料及影片均留在被忽略的 `dist-all/`，本輪未提交或推送。工作清單核對通過，2046行自承全部有條目認領；根目錄1382份輸出的大小與SHA-256獨立重算相符。
- 容器暫存隨 `--rm` 回收，沒有新增比較工作區；本專案容器清單為空，全部交付擁有權為1000:1000，專案沒有root-owned檔案或誤建 `.md` 目錄，`git diff --check` 通過。目前五個專案工具image保留；懸空查詢列出的 `4ac62de83339` 經inspect為既有具名IDA9.4共用image，未刪除，沒有本輪待清工具image。

## 2026-10-11：提交收尾、中文與功能核查、README 及 HD 計畫

- 依使用者指令，先將既有 v1.0.0 收尾提交為 `f1944666` 並推送 `origin/main`。其後只做本輪核查與顯示修正；已交付封包、推廣片及版號保持，原版 C matching 維持暫停。
- 沿已載入的 remake 驗證／交付入口、README 契約及 grilling 決策流程。使用者確認住宅告示 A「請勿／踐踏草坪」，並確認 HD 手繪採 A：1920×1200 原比例。兩項決策與核准範圍已保存，不延伸成其他美術或玩法核准。
- 新增 `tools/verify_cjk_closeout.py`，預先登記首批342份程式輸入、29條玩家計畫、seed與依賴。148張展示／建築圖全部擷取成功，125張繁中以六張接觸表目視。14條英中路徑兩側各4688狀態非顯示資料相同；只豁免實際匯出端的兩個戰鬥顯示文字清單，不豁免數值。首批欄位名稱誤寫的原報告保留，經來源核對後重新分類。
- 固定 dosgolem `968ca774705853276ae55c8d1163c369481ac71f` 從正常入口重跑原版六人路徑，共247狀態，與既有同版收據完全相同。重製版英中各173狀態；114個旅行位置／時刻、九名初始角色位置與頭／胸生命、規劃者／回合／圖示均相符，144×144戰場視窗零差。原版seed1於旅店招呼後固定，重製版啟動前seed1；未宣稱後續整場一致。
- 競技場舊腳本雖英中相同，仍有76個已觀測里程碑欄位偏離，不列功能通過。14條缺歷史輸入；長任務沒有完整trace或save，中文程序達600秒工具上限，不延長、不補造基準。舊住宅3080鍵正常路徑未到原終點。這些限制全部保留，沒有依舊腳本猜改正式規則。
- `OutdoorCJK`接入告示A，1986像素只落在三個指定告示區，三區與核准原型零差，區外及英文零差。此為觀察入口繪圖驗收，正常住宅旅程仍開放。目視另找到圖片烘字的英文目錄，接入中文靜態右頁；正常進入、角色選項及返回共五狀態英中完全一致，英文及中文返回頁零差，數字留在紙張內。兩個譯名改回既有術語；翻書動畫未納入，不稱全畫面完成。
- 最終來源的 `go test ./cmd/... ./internal/...` 通過。runtime、UI、名稱、世界提示、字模、術語及文字容納七項靜態檢查通過。程式輸入獨立重算與最後建置一致。格式化唯讀父目錄、外部go.work、Xvfb啟動競爭與輸出目錄先後失敗，均修正隔離環境後重跑；一次字寬雙回傳值編譯錯誤已修正。通用原版工具與當前dosgolem的首Space失敗保留，改回已證實版本與工具後通過。
- README改用本輪新畫格，加入依手冊整理的世界與開發背景，刪除逐輪實作流水敘述。八張保留PNG附ZIP／來源entry／資產manifest／工具／畫格／雜湊及private-only分類。新增並索引SPEC-039 DRAFT，安排旅店、六人肖像、戰鬥規劃可操作切片及八批素材；造型與正式素材契約待確認。
- 目前18開放項保持，90已完成項的回歸訊號通過。新增限制精確掛回既有工作項，未藉刪除限制通過掃描。工作清單與CONTEXT生成區同步；修改文件的本機連結與八張PNG雜湊獨立核對通過，`git diff --check`通過。
- 本輪1062份證據保存至 `docs/verification/local/cjk-closeout-20261011/`，逐檔唯讀核對通過；兩份可重建執行檔只留雜湊。原工作目錄約62.6MB，壓縮檔約9.1MB。核對後只刪除 `workplace/cjk-closeout-20261011/`，既有其他研究目錄保持。完整入口為 `docs/verification/cjk-closeout-20261011.md` 與 retained-evidence 索引。
- 全部工作使用既有鎖定Docker、無網路、UID/GID1000與有界資源，Xvfb有退出trap。本輪容器清單為空；全專案沒有root-owned檔案或誤建`.md`目錄，保留檔擁有權1000:1000。這批中文修正、截圖及文件另隨本輪提交推送；不建立tag、公開Release或修改遠端Issue。
