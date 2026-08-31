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
