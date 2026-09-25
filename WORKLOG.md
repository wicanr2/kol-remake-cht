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
