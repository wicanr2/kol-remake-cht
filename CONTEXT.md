# 《武士傳說》重製目前狀態

更新日期：2026-09-04

**自評 54 / 100**。算法、權重與逐系統的完成度依據都在
[WORKLIST 的自評分數](WORKLIST.md#自評分數)——狀態只放一處，這裡不複製。

## 目前真相

- 本專案是 *Knights of Legend* 的繁體中文乾淨重寫，正式 runtime 使用 Go／Ebitengine、
  PNG 圖像與 Ogg Vorbis 音訊。
- GitHub `wicanr2/kol-remake-cht` 已建立為 private。原版 ZIP／EXE／archive／PDF 與其他 raw
  素材不得提交；使用者已於 2026-09-01 特別允許已驗證、可重生且有 provenance manifest
  的 PNG 衍生物入庫：抽取器產生的素材 PNG 放 `assets/original-derived-png/`，模擬器
  擷取的原版執行畫面 PNG 與其收據同目錄放 `docs/verification/`。這不授權 GitHub Release
  或公開散布。
- Go／Ebitengine 垂直切片已啟動：`cmd/kol` 以 Go 1.26.7＋Ebitengine v2.9.9 建於固定
  映像 `eob-remake-go:1.26.7-ebiten2.9.9-audio`，只讀 `assets/original-derived-png/ega/`
  的 manifest 與 PNG，畫在 320×200 原版解析度離屏畫布上，並提供截圖模式供對拍。
  書本的六個畫面（封面、目錄頁、角色建立、角色清單、刪除、HARBOR MASTER）、兩段
  翻頁動畫、TITLE 標題序列與城鎮的旅店、戶外地圖（含地形擋路、門口與地圖邊界的
  城門提示、進建築、招呼語與 LISTEN／MOUTH 對話、招募同伴、客棧住宿與修道院治療、
  MIRROR 的角色資料頁、商店買賣、AWARD 訓練、存檔與讀檔）
  與野外地圖畫面（含 ROAD／SIGNPOST 自動循路、進城、時刻與日曆推進）都已實作並對拍。
  任務地點的戰鬥走得完一輪：戰場的 9×9 視窗、逐格拉線的視線、參戰者疊圖與白框，
  指令列、選目標、四列子選單、兩段回合順序、移動、近戰結算與逐句訊息都實作了，
  實跑打得死怪物（RE-079…090／SPEC-008、009、010）。
  **但一場戰鬥還打不完**——`execution.go` 只判隊伍全倒，把怪物全部打死沒有判定。
  野外的隨機遭遇規則已解出（RE-078）但還沒實作。
  SPEC-001／002／003／005／006／007／008／009／010 為 READY 或 CONFORMED
  （008、009、010 只有已解出的段落 READY），SPEC-004（音訊）仍為 DRAFT。
  戰利品、經驗值與升級**連 RE 都還沒開始**，入口 `sub_289EE` 已定位但未解讀。
- 原版內嵌點陣字型已抽出：`BOOKDATA.0`／`KOLDATA.0` 的第一個 payload 是同一份
  1,303-byte 區塊（SHA-256 `819e4f82…d318`），前 744 bytes 是 93 個 8×8 字模、
  `0x2e8` 起是 93 個 8×6 字模，都由 ASCII 32 起算涵蓋 32–124。
  `tools/extract_fonts.py` 產出 `assets/original-derived-png/fonts/`。書本內文用小字型。
- 書本文字版面已機械量得並版控於 `docs/re/text-layout.json`；原版繪字規則為
  逐字模格清成白底、但空白字元不清底。原版字模格是 8×6、字距 8，繁體中文字形放不進
  同一格，因此採雙模式：英文模式維持 320×200 與原版字模供嚴格對拍，中文模式 640×400、
  16×18 中文字格。城鎮的文字則用 8×8 大字型，不是書頁的小字型。
- remake 目前畫面的展示圖在 `docs/screenshots/`，由 `tools/run_kol_shots.sh` 依
  `shots.spec` 在鎖定容器內一次重生，`MANIFEST.json` 記下產生指令、輸入雜湊與 PNG
  SHA-256，並要求有對拍收據的畫面與那張逐 byte 相同。README 的畫面區塊引用同一批圖。
- remake 對拍收據已建立（`docs/verification/remake-parity/`）：二十五份收據 283 對畫面
  全部通過（書本畫面 3、角色建立 9、圖像層 92、三段翻頁 54、刪除流程 13、
  HARBOR MASTER 1、旅店畫面 8、戶外地圖 22、建築進門 1、城門提示 4、建築對話 3、
  招募同伴 9、住宿存檔 11、角色資料頁 8、商店 9、訓練 6、野外地圖 4、野外時刻 4、
  任務地點 4、道路 10、進城 7）。野外那一份用 `compare` 指明每一對涵蓋的矩形，天候圖列在
  `not_claimed`——雲層動畫還沒實作，那一塊不宣稱，也不當成允許差異區。
  書本封面與目錄頁在原版滑鼠游標以外與 DOS EGA 原版逐像素相同；80 張 New Character
  畫面的角色圖像 blit 矩形 (208,106,40×38) 全部逐像素相同；空白 New Character 頁
  以原版字型與版面重繪後，64,000 個像素中只有 31 個不同且全部落在名字欄的原版殘留
  字元。對拍判準分成「素材與版面必須相同」、「明列允許差異區」與「必須相同區」三層。
- `Knights-of-legend.zip` 已在唯讀容器中通過 ZIP CRC 檢查，共 76 個中央目錄項目；可見
  6 個 MZ executable、安裝批次檔、城鎮／角色／物品資料及多組 `*PICS` 圖像候選。
- ZIP 多數遊戲檔時間為 1989，但 `KOLTOWN.EXE`、`CHARDATA`、`CHARPICS`、`SHLDPICS`、
  `STEEDPIC` 等顯示 1998。這份封裝是可運行 DOS 集合的 `強推論`，是否為未修改原始
  發行版仍是 `未知`。
- IDA Pro 9.4 `locked-v1` 已對六個 executable 完成第一層指紋。全部由 IDA loader 判定
  為 16-bit MS-DOS EXE；`KOL`、`KOLBOOK`、`MAKEMODE`、`TITLE` 出現同一個 Microsoft
  Quick C／MSC 5.1 runtime FLIRT 候選，這只證實可套用的 library signature 候選，不能
  鎖定精確 compiler 版本。`TANDYPIN` 尚未辨識函式，仍屬未知。
- `KOLTOWN` entry 已證實會由區塊尾端向前迭代，把 `0x131f` bytes 搬到較高的重疊
  segment，隨後 `retf` 至新 segment 的 `IP=0x0038`；這是 EXEPACK 前置搬移。檔案相對
  MZ 宣告長度多出的 81 bytes 已證實全為 `0x1A`，屬 DOS EOF padding，不是 overlay payload。
- `KOLTOWN` 已證實使用 Microsoft EXEPACK；精確版本未知。以固定 SHA-256 的 CC0
  `unexepack.c` 產生本機 derivative `01f6f9…d1d2`，IDA 9.4 在 unpacked EA `0x2b90c`
  建立真正入口後辨識 337 個函式。解包 EXE、`.i64` 與完整 disassembly 僅留本機 ignored。
- DOSBox-X 第二次乾淨啟動以 `KOLTOWN.EXE A` 在 runtime `23B3:001C`、C startup 前命中；
  2,281 個 relocation 全通過，正規化後 129,312-byte load image 與 derivative 零差異。
- 魔旅士 2010-04-16 的玩家心得與《CRPG 通鑑》中文版已進入 README；只作專案起源，
  不作玩法 oracle。
- 中文手冊 SPEC-003 已 CONFORMED：Poppler 22.12.0 對 79 頁玩家手冊與 5 頁參考卡建立逐頁
  雜湊清冊；82 個內容頁都有文字層，玩家手冊 76、77 經頁面 PNG 目視為空白跨頁。
  玩家手冊 79 頁與 Apple II 參考卡 5 頁已全部 reviewed；84 頁總 gate 通過。
- 圖片 RE：44 個實體候選歸併為 30 個邏輯 archive；TWN `.1/.2` 已證實是 continuation
  volume，總數由 960 訂正為 1,022 個 payload，全部通過 LZW 解碼。索引／分卷契約、
  尺寸、位平面、palette 與透明規則均已逐 archive 證實。
- `AWD1PICS.0` 是 KOLBOOK 固定寫出的 4000-byte 工作區快照；KOLTOWN 只讀其中索引
  payload。尾端 520 bytes 已排除於圖片流，但仍是未解析工作區資料，不是 padding。
- `CHARPICS`／`SHLDPICS`／`STEEDPIC` 已由 KOLBOOK／KOLTOWN 讀寫資料流證實為 16 筆、
  每筆 760 bytes 的固定可變角色圖片記錄表。KOLBOOK `sub_15AEC` 與 4 倍顯示
  `sub_15B6E` 已證實每筆是 40×38、high-nibble-first 的 4-bit EGA 圖；48 張 ignored
  PNG 與接觸表已重生。逐槽語意與未使用槽內容仍未知。
- `CHARPICS.0` 全 80 筆已由 KOLBOOK `sub_13D7D` 索引／解碼鏈與 RE-027 下游 consumer
  證實為 40×38、high-nibble-first 的 4-bit EGA 完整角色模板；每筆第 761 byte 為
  `0x00` 且不進像素。80 張 ignored PNG 與接觸表已重生；逐圖職業／裝備名稱仍未知。
- COMBPICS 全七筆已關閉：0..4 為 296×144 戰鬥場景，5..6 為 88×64 肖像，皆使用
  二階 RLE、4-bit 色彩與 BIOS mode 0Dh palette；圖號 0、1 另對上魔旅士文章同場景 JPEG。
  這只證實 COMBPICS archive，不代表全部圖片格式完成。
- TWN0..7PICS 全 243 筆已關閉：TWN0 全 25 筆為 192×96；TWN1..7 圖號 0 為
  56×57 彩圖加共享 AND-mask、圖號 1 為七筆一-byte placeholder、偶數圖號 2 起為
  88×64、奇數圖號 3 起為 192×96。236 張 ignored PNG 與八張接觸表已兩輪重生；
  TWN0、TWN1、TWN7 人工抽查通過，個別 NPC／地點語意仍未知。
- MONPICS 已關閉 32 組怪物 sprite 垂直鏈：64 個 payload 依序配成 4-bit EGA 彩圖與
  MSB-first 1-bit 遮罩；IDA 尺寸表、逐張像素數與透明接觸表均吻合。遮罩有效資料後多出的
  單一 LZW byte 不作像素，仍標為壓縮終止／padding 的強推論。
- PERPICS 已關閉六組人物圖垂直鏈：前六個 payload 是 4-bit EGA 彩圖，後六個依
  `[i,i+6]` 配成 MSB-first 1-bit 遮罩；六組 IDA 矩形、像素數與透明接觸表全部吻合。
- KOLMPICS 全 83 筆已分類：71 筆 active、12 筆一-byte placeholder。圖號 0 是
  320×200、2 是 296×32；68 筆共同展開為 24×20 UI 圖示，其中非固定 caller 的個別幾何
  位置／裁切不宣稱 caller-exact；canonical extraction shape 已證實。28 是 56×57
  AND-mask。68 圖示接觸表方向與序列均連貫。
- WEAPPICS 全 56 筆已關閉為 28 組角色武器彩圖／AND-mask：`sub_18F3A` 由裝備欄位、
  六值 selector 基表與 family 5 的 variant 2／3 交換規則產生偶數圖號；28 組尺寸、
  像素數與透明棋盤接觸表全部吻合。family／variant 只是導覽標籤，尚未推定遊戲內分類名稱。
- HEADPICS／TORPICS／LEGSPICS 各 108 筆已關閉為各 54 組角色頭部、軀幹與腿部
  彩圖／AND-mask；共享 selector 是 `9*renderer_class+variant`，三份 6×9 接觸表與
  324 筆像素／mask 長度均吻合。指令允許 variant 0..15，實際合法資料範圍仍待另證。
- KOLPICS 全 14 筆已分類：0／8／12 是 `sub_19B29` 轉換的 raw-planar 資料，1／2／4／5／13
  是 direct-blit RLE 圖，3／6／7／9／10／11 是一-byte placeholder。RE-030 已證實圖號 0
  是五個 converter block 的 256×5 renderer-data PNG；8／12 由 40 個完整 stride 與
  `y<<6+x` 固定為 64×160 grid PNG。這些不是 original display rectangles。
- MISCPICS 六筆已完整分類：image 0／1／3／4／5 是五張 direct-blit 4-bit EGA 彩圖；
  image 2 是唯一一-byte、全程未引用的非圖片 placeholder，不產生 PNG。
- 固定 ZIP 一般圖片只有 EGA `PICS`，但 `MAKEMODE.EXE` 已證實是官方 CGA／Hercules
  轉換器：它對 28 組 PICS archive 生成 `PCGA`／`HERC` 分卷，28 組圖數與現行清冊完全
  相符。使用者已於 2026-09-01 決定首版只做 EGA，CGA／Hercules 列為後續選項；
  MAKEMODE 的完整轉換像素規則保留為待辦 RE，不進首版交付範圍。
- SPEC-002 已 READY。完整 EGA extraction set 為 30 個邏輯 archive、1,022 payload：
  996 筆來源已證實、26 筆非圖片 placeholder、0 筆 unknown，形成 752 個 PNG 單位。
  `tools/extract_ega_assets.py` 已把全部 769 張 PNG 產到 `assets/original-derived-png/ega/`，
  逐張重讀驗證並與已版控 visual probe 紀錄的 SHA-256 核對，原子化替換、輸出根目錄受限。
  原版同狀態 parity：80 張 CHARPICS 模板在 blit 座標 (208,106) 兩輪各 80/80 逐像素相同；
  BOOKPICS[0] 對書本封面畫面只差螢幕中央 8×11 的滑鼠游標。
- 音訊 RE 已建立 SPEC-004 DRAFT。固定 ZIP 沒有獨立音訊檔；TITLE `sub_10F75` 已證實為
  兩段、每段 100 次隨機 PIT 除數更新的 PC Speaker cue。`TANDYPIN.EXE` 已證實是 DOS
  `AH=31h` 常駐記憶體程式，不是音訊播放器。六個 executable 的雙新建 IDA DB 直接 I/O
  普查只有 TITLE 命中音訊埠；六個未解析 DX 點均已人工排除為 EGA renderer。正常啟動的
  DOSBox-X `DX-CAPTURE /A` 已取得 38.229333 秒 WAV，四段非靜音窗及 peak 均已固定。
  使用者已於 2026-09-01 決定 OGG 採規格近似合成：依已證實的 PIT 除數序列與公開硬體
  規格自行合成方波再轉 OGG，不散布原版錄音。合成器實作與人耳驗收尚未完成。
- 角色建立已建立 SPEC-005 DRAFT。手冊固定名字→種族→性別→職業→reroll→圖像→保存
  狀態機；`CHARDATA` 為 16×768 bytes、`CHARPICS` 為 16×760 bytes。IDA 固定 KOLBOOK
  `chardata` EA `0x1945a`、路由 `sub_134C2` 與五個立即 arg=0 caller。屬性 RNG 已證實為
  BIOS tick seed、`rand()%N+1` wrapper、八值骰面表、獨立骰式／內部基值與 100/150 顯示
  量尺；三個衍生值均為內部屬性的向下取整平均。`BOOKDATA` 的 39 筆記錄、四種族、兩個
  secondary 值及 33 個 menu names 已建立完整 selector；35 個手冊列與 39 筆記錄的差異由
  男女精靈共名及矮人雙名稱同記錄解釋。39 筆起始金冠、主要／次要武器及武器技能已與
  中文手冊逐列相符；最小 writer offset、CHARPICS／SHLDPICS 同槽寫入及三檔保存順序亦已
  固定。完整 writer coverage 已排除其他職業衍生起始技能／物品，KOLTOWN 最小 reader 亦
  交叉確認名字、種族、職業、金冠與主要武器。KOLBOOK 靜態控制流另已固定輸入 buffer
  首 byte 為零時重試、各層取消返回、選項 0 reroll、80 張肖像循環與 16 槽全滿時在 writer
  前退出。
- SPEC-005 已 READY。正常 DOS 玩家路徑的書本與城鎮開場 UI 有 22 個情境、32 輪的可重生
  收據（`docs/verification/character-ui/`，43 個 run、6,515 個去重畫格；核心情境各兩輪）。
  入口鏈為 `KOL.EXE` → TITLE
  循環標題序列 → 任意鍵 → 書本封面 → 任意鍵 → 目錄頁 → `5` → Character options →
  `CREATE CHARACTER`。名字欄容量 16 個可見字元，輸入第 16 個字元即直接前進種族層、
  不需 Return；空名按 Return 不前進；取消目的地為種族→名字、性別→種族、職業→性別、
  屬性→職業，回到名字層時欄位清空；`REROLL STATS` 重擲九個顯示值；圖像 selector 恰好
  循環 80 張；接受圖像後才寫檔，兩輪都只寫 `CHARDATA` 的 slot 14。跨輪差異只出現在
  TITLE 動畫落點、名字欄游標與屬性數值三類畫面且差異量已量化。16 槽全滿、重名處理、
  名字欄字元集合與大小寫保存方式仍未由 UI 證實，實作必須標為未驗證分支。

## 固定輸入

| 檔案 | Bytes | SHA-256 | 分類 |
|---|---:|---|---|
| `Knights-of-legend.zip` | 1,699,452 | `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46` | 僅本機原版封裝 |
| `knightslegend-manual.pdf` | 35,942,482 | `4c80f5d4c7221f9d3376546d249d52e2fc7fc0f73b48ee9daf6e20ce87d7086a` | 僅本機英文說明書 |
| `knightslegend-refcard.pdf` | 2,253,388 | `4815bc6c3a51ff7abaa610a071f08e05eb2dd180e5130b0674de8b4a94b2abd3` | 僅本機參考卡 |
| `KOL-MAP-001a4s2.jpg` | 92,661 | `9528ca50e14d09d746d2ecce523ac56f78baeaec5cdd4ba5e3351679bd52dff9` | 僅本機參考地圖 |
| `knightslegend.jpg` | 433,061 | `76b08e4b6a6da9b616d7f57ce3e77c2973591379c8897779d849126555defb9b` | 僅本機盒面 |
| `knightslegend-back.jpg` | 525,582 | `10d3e63358d6604b807d4afae3b53acf9a2962ce10871390dd815e325e53b194` | 僅本機盒背 |

## 文件入口

- 執行規範：`AGENTS.md`
- 唯一待辦：`WORKLIST.md`
- 工作歷程：`WORKLOG.md`
- 研究證據：`docs/re/RESEARCH-LOG.md`
- RE 知識庫索引：`docs/re/README.md`
- 重製計畫：`REMAKE-PLAN.md`
- 驗證矩陣：`docs/verification/VERIFICATION-MATRIX.md`
- 交接：`HANDOFF.md`

## 下一個唯一工作

**把選中之後的動作接上**。戰場畫面、視線、站位與回合的規劃順序都已實作並對拍；
兩邊的決策規則都解完了（敵人 RE-081、玩家 RE-082／SPEC-010），玩家那一手也已
實作到能操作：主指令列、選目標格與四列子選單都串起來了，圖示列與原版
frame-886 逐像素相同（RE-083）。

執行段也已實作（RE-084）：整輪規劃完答應之後另外排一次順序（快捷扣傷勢、負重與
指令的時間成本，鍵值大的先動），攻擊那一手接 SPEC-009 的結算——**戰鬥現在真的會
造成傷害、會有人倒下**。

移動那一手也已實作（RE-085）：**一次只走一格**，沒有點數規則——先前把移動方式
當成距離是誤解。走不動的四種情形（地形、腳下、有人、界限）都照原版判，其中擋路
用的是與擋視線不同的另一張位元圖。

訊息也接上了（RE-086／RE-087）：代換碼十一個、`]` 換成空白，先進緩衝區收尾才
斷行，版面在 byte 欄 3（像素 24）、掃描列 0xA8、行高 6。實跑「進戰場 → 選休息 →
執行」之後從畫面讀回 `Poontang is resting!`——**這是 remake 自我一致的驗證，
不是 parity**，那一輪 oracle 沒有走進執行段。

攻擊的訊息鏈也解出來了（RE-088）：一次攻擊分兩句、各等一次鍵——第一句是動作
（whacks／kicks／fires a missle…）加結果（落空三段、被擋、命中），第二句是傷害
（部位、輕重三段、倒下與斷肢）。招架碼的訊息基底是 0x15F，不是 RE-080 原本寫的
0x14F。實跑讀回 `Poontang whips his Long Spear through the air.`。

三條路都實跑過了：走進戰場、往下移動四步貼到最近那一隻 BINDERAK、攻擊，讀回
`Poontang whacks his Long Spear at The BINDERAK who cannot parry the blow!`
與 `The blow hits The BINDERAK's chest killing it instantly!`。
**remake 現在打得死怪物**——規劃、移動、攻擊、傷害與訊息整條通了。

執行段的原版 oracle 已經錄了（RE-089，情境 `world-quest-execute`，969 個畫格）。
量到五件事並照著修：訊息區整條在執行段換成白底、文字**白底黑字**、整段**轉大寫**、
起點 (24,168) 與行高 6 和反組譯讀出來的一致、移動成功要**按兩次鍵**。

當時掛著的兩個差異都收掉了（RE-090）：訊息區的底圖是 MISCPICS 的圖 1，一張卷軸，
remake 已改成畫它，`receipt-battle-exec.json` 那一對允許差異區之外零像素差；敵人的
名字則沒有對不上——原版那一輪沒走到任務地點，是路上被 1 隻 1 級 RUFFIAN 埋伏。

還沒做的是野外的隨機遭遇（`sub_2A4E5` 的挑怪骨架已讀出，但四張表的語意還沒證實）
與撿、丟、背包那三手。

敵人 AI 的規則已經完整：六條分岔（撤退、休息、射擊、貼近打、移動、繞圈）、
以行動者為中心的七個同心方環搜尋、篩選遮罩、智力擲骰決定要不要做戰術判斷、
反制掃描抄鄰居的攻防選項、加權輪盤用 `權重^等級 ÷ 2^(等級−1)` 讓高等級怪物偏向
好選項。**缺三張基礎權重表的執行期值**（移動三式 `byte_3496A…C`、攻擊式
`[0x6AD2]`、部位 `[0x1A9A]`）：靜態影像裡是未初始化區，全庫掃過沒有具名 writer，
要靠執行期取樣才能定。在那之前敵人那一手維持跳過並記為已知差異。

循路的時刻成本已於 2026-09-03 修好：原版量到「往北十二步（PLAINS，每步 56）之後
連按 ROAD，時段標籤在第 9 次才翻成 MID MORN」，往回推一步扣 18…20，取 20
（`world.RoadStepCost`）。機制還沒追出來——`sub_297F2` 只有一處扣時刻，
呼叫前的 `word_39A54`／`word_2FA2C` 對不上這個值，常數旁邊已註明。


角色建立的名字→種族→性別→職業→擲屬性→圖像層已在 Go／Ebitengine 實作，並與原版
同狀態畫格逐畫面對拍通過：`newchar-steps.json` 9 對（對 `reroll-run1`）、
`newchar-pictures.json` 92 對（對 `picture-cycle-run1`，含 80 張圖像循環與兩端回繞）、
`book-screens.json` 3 對，全部在允許差異區之外零像素差。殘餘差異只有四類，成因都已
查證並記於 RE-045、RE-046。

刪除角色（DELETE CHARACTER）已閉環：畫面 13 對逐張通過（`delete.json` 8 對、
`delete-move.json` 5 對），確認視窗的座標與兩行文字都由原版畫面機械量得，刪除本身
只寫一個 byte，remake 的輸出與原版逐 byte 相同（`chardata-delete/`）。

城鎮的第一張畫面已對拍：旅店畫面由紅磚底圖加隊員頭像格、店主肖像（`TWN7PICS[12]`）、
標題框、對話框與六格圖示列組成，八對逐畫面對拍全過，允許區只有原版滑鼠游標
（RE-054、RE-055）。中文模式一併補上。

城鎮的戶外地圖已實作並對拍（RE-056 至 RE-060）。地圖視窗是 12×6 格、每格 16×16，
格子索引是 `(X<<6)+Y`：`TOWNn.0` 的 payload 0／1／2 解碼後接成一塊，前 4,096 byte 是
64×64 的格子陣列（一格一 byte ＝ 磚塊編號），其後 32,768 byte 是 256 塊 16×16 四平面
plane-major 的磚。隊伍標記的精靈在 KOLPICS payload 0（十個 128-byte 單元，城鎮用 0–4
五個畫格來回），它四周那圈黑框是算出來的：四平面 OR 成佔用位圖、左右各膨脹一像素、
反相當 AND 遮罩，而膨脹在 little-endian word 上做因此不跨位元組邊界。

隊伍先在視窗內走、走到一定位置鏡頭才跟著捲一格，兩者分開記（`Camera` 與 `Marker.Row/Col`）。
地形會擋路：`TOWNn.0` payload 3 的位移 0x00–0x1F 是「磚塊編號 → 不可通行」的位元圖，
移動前查目的地那一格；TOWN0（野外）的位元圖全為零，七座城各擋 125–141 個磚塊編號，
而 103 個門的座標全部落在可通行的磚塊上（RE-065）。夾限的是**鏡頭**不是隊伍，
所以隊伍最遠站到界限＋1；夾到就跳出城門提示：地圖視窗換成 MISCPICS 圖 0 的林地圖、
圖示列 DOOR 與 U-TURN，選 DOOR 就離開城鎮（野外地圖本身還沒實作，
remake 記成 `Session.LeftTown`）。圖示列是一個會被清空重建的堆疊：走離門口清空、
踩到門重建成兩格、選 U-TURN 與從建築物出來都只剩一格 DOOR。
`tools/solve_outdoor_camera.py` 從 94 張移動 oracle 反解出 22 張戶外畫格的狀態，
每一張的鏡頭都是唯一解；21 對逐畫面對拍全過，允許區只有原版滑鼠游標。

站在門的座標上會進入門口狀態：地圖視窗換成該建築的 192×96 門面圖，圖示列只剩
DOOR 與 U-TURN（RE-061）。門 *i* 的四份資料同一個索引推出來——門面圖
`TWN{n}PICS[2i+3]`、肖像 `[2i+2]`、名稱與人物在 `TOWNn.0` 的 payload `i+4`。
門口是**狀態不是位置**：掃門迴圈接在移動後面，剛走出建築站在同一格上仍是地圖畫面。

城鎮的對話文字已經解出來（RE-062）：`TOWNn.0` 每棟建築的 payload 從位移 128 起是
16 byte 一格的文字格，第 0 byte 是字數、其後 15 byte 是 24 個 5-bit 字元，
小寫預設、大寫靠一個跳脫記號。字元表與空白規則由解碼器 `sub_1B94F` 直接讀出
（RE-064，已證實）：`!` 與 `.` 各補兩個空白並打開跳脫，跳脫中的 `"` 會往回吃掉空白，
所以 `[1B][1B]` 解出來是 `!"`。七座城 101 棟建築共 9,515 行，
連同 `KOLDATA`／`BOOKDATA`／`ENDTXT` 三份 UI 文字庫，構成中文化的文字母體。

七座城 101 棟建築的名稱、人物與進門敘述都已翻成繁體中文（`docs/i18n/town-zh-Hant.json`），
文字安全矩形 863 項全過。城名由世界地圖記錄的三字母代碼配上城鎮對話定出來（RE-066）：
BRT 布雷特爾、HTN 赫特隆、THM 辛博瓦德、POT 波伊特水閘、SHL 謝勒農、OLN 奧蘭斯、
MIN 霍布（十個礦坑入口共用 TOWN7 的地圖，各佔 16×16 一塊）。原版資料一律拼作
Shellernoon，說明書的 Shellemoon 是異拼。還沒翻的是建築內的後續對話與兩份 UI 文字庫。

建築內的對話機制已解出並實作（RE-067）。每棟建築有六個話題槽（payload 位移
0x17…0x1C），低 6 bit 是話題碼、高 2 bit 是類別。**LISTEN 只講話題碼為 0 的槽**，
隨機挑起點再往回找沒講過的；有關鍵字的槽要玩家用 MOUTH 打字問到那個字才會講。
關鍵字表在 `TOWNn.0` 的 payload 20，四字元字首、字典序排好，七座城共 305 個；
沒有人回答時會用 KOLDATA 的句子指出鎮上誰知道。招呼語依發言者的性別、會不會法術
與聲望等第選（訊息 10…13 或 26 ＋ 等第），後面接 KOLDATA 記錄 45。
409 個非空話題槽的關鍵字與對應訊息交叉檢查零例外。

中文字模表的烘焙修掉一個會影響所有中文畫面的缺陷：ffmpeg 的 drawtext 是依照
**這一次要畫的字串**的最大 ascender 對齊的，因此一個字單獨畫時墨跡頂端一律貼齊格子
上緣——逗號、句號被拉到半空中。現在每一格先畫一個全形驚嘆號當基準字再畫目標字，
兩個字共用基線，目標字的縱向位置換算成「相對基準字墨跡頂端」。

按下 DOOR 之後的進門畫面已實作並對拍：版面沿用旅店畫面，內容完全由三份 typed data
組出來——門座標表給建築名與人物名、`TWN{n}PICS[2i+2]` 給肖像、5-bit 文字給敘述。
一則訊息固定四格（64 byte），訊息 0 與 1 合起來是進門的八行敘述。
畫面規則只有兩條：句號後補兩個空白並把下一個字母轉大寫、逗號後補一個空白。

城鎮文字已開始中文化（RE-063）：譯文表 `docs/i18n/town-zh-Hant.json` 依「城／門／欄位」
分層、譯文自己帶斷行，布瑞提爾 21 棟建築的名稱、人物與進門敘述共 168 行已翻好。
中文字模表因此由 371 字增至 770 字，文字安全矩形檢查 286 項全過。

城鎮這一段已接上輸入迴圈（`internal/town/session.go`）：`ModeMap` 走動、
`ModeDoor` 站在門口、`ModeInside` 進了建築，三種模式互轉。鏡頭跟隨的規則是
「標記在視窗內的偏移夾在 0…1，多出來的由鏡頭吸收」——由 22 張原版畫格反解，
強推論。可重播的玩家路徑收據在 `docs/verification/player-path/`：14 步，
每一步都有可見效果，其中兩步與已通過對拍的截圖逐 byte 相同。

`internal/flow` 把已經各自對拍過的畫面串成一條可玩路徑：標題 → 書本封面 → 目錄頁 →
角色選項 → 角色選擇清單 → 旅店 → DOOR → 戶外地圖 → 門口 → 進建築。它只決定
「按了什麼鍵、畫面走到哪」，各畫面自己的狀態機維持原樣。26 步收據在
`docs/verification/player-path/full.json`，其中兩步與已對拍的截圖逐 byte 相同；
按下去還沒實作的功能會記進 `flow.Unimplemented`，不會默默吞掉。

書本六個選單項目裡五項有收據。`HARBOR MASTER` 在只安裝一個 region 時是一張三行訊息頁，
已實作並對拍通過；`EDIT PICTURE` 與 `EDIT SHIELD` 是同一個滑鼠驅動的點陣圖編輯器，
ESC 沒有作用，現有的鍵盤 oracle 走不完，尚未實作（RE-052）。四個功能進入後的第一張
畫面是同一張角色選擇清單。

接下來依序做三件事：

1. remake 自己的存檔格式。角色寫入已經對拍（RE-050：原版改動的 58 個 byte，兩個職業各
   58/58 逐一相同），起始武器的索引規則也已反追出來，39 個職業都能寫；建立→寫入→清單
   已閉環。remake 自己的存檔格式已於 2026-09-03 定案（見下方「已定案的共同決策」）。
2. 動畫已對拍：TITLE 標題序列（RE-047，原版兩輪觀察的每一張畫格都對上）與翻書動畫
   （RE-048，18 對全通過）都已實作並可實際播放。剩下的是把雷聲接上 Ebitengine 音訊
   路徑，那要等 SPEC-004 通過人耳驗收。
3. SPEC-004 的人耳驗收：`assets/audio/title-thunder.ogg` 是 `sub_10F75` 的規格近似合成
   候選，技術驗收已通過（時長與原版 DOSBox-X 擷取四段的平均值差 0.000581 秒），
   但**需要使用者實際聆聽**才能把 SPEC-004 升 READY，在那之前不接 Ebitengine 音訊路徑。

繁中雙模式已實作：英文模式維持 320×200 與原版字模，供嚴格對拍；中文模式 640×400、
16×18 中文字模格，New Character 頁已全頁中文化（`docs/verification/cjk-ui/`）。
字模表衍生自 Noto Sans CJK TC（OFL 1.1），散布時必須隨附授權全文。

## 已定案的共同決策

| 日期 | 議題 | 決定 |
|---|---|---|
| 2026-09-04 | 授權條款 | **RRSAL-1.0**（復古重製 source-available 授權條款 1.0，SPDX `LicenseRef-RRSAL-1.0`）。非商業免費含修改與再散布、實況與影片明示允許（含平台分潤）、貢獻回授、商業要洽談。`LICENSE` 與 README 授權段已建立；**不涵蓋原版素材**，第 2 條逐項寫明界線。這不是開源授權，對外一律寫 source-available。 |
| 2026-09-03 | 戰鬥忠實度 | **忠實 ＋ 可關閉的加速**。規則、傷害計算與回合結構完全照原版；另提供「重複上一指令」「快速結算未接敵回合」「動畫速度」三個可關閉的選項，預設全開。**對拍一律在三個選項全關的狀態下驗證**，加速只影響玩家的等待，不影響任何結果。 |
| 2026-09-03 | 存檔格式 | **remake 自有格式 ＋ 匯入原版**。存檔是可讀的 JSON，含 remake 才有的狀態（世界時間、任務進度、設定）；另提供從原版 `CHARDATA` 匯入角色。原版檔案一律唯讀，不寫回。已落實為 SPEC-006、`internal/save` 與 `tools/import_original_characters.py`。 |
| 2026-09-03 | 客棧住宿是否離隊 | **預設不離隊，設定開關 `faithful_inn_stay` 切回原版**。原版住宿是「收費 → 寫存檔 → 角色離隊 → 隊伍空了問是否離開遊戲」，那是一人一檔的磁片設計。開關打開就完全照原版跑，**與原版對拍一律在打開的狀態下進行**。 |
| 2026-09-01 | 圖像 theme | 首版 EGA-only；CGA／Hercules 列為後續選項。 |
| 2026-09-01 | PNG 入庫例外 | 已驗證、可重生且有 provenance manifest 的 PNG 衍生物可進 private repo；不含 ZIP／EXE／archive／PDF，也不等於公開散布權。 |

待使用者決策：`assets/original-derived-data/character-creation.json`（39 筆職業骰式、
內部基值與名稱表）是原版衍生的 typed data，不是 PNG。2026-09-01 的 private-repo
例外明文只涵蓋 PNG，因此這份檔案目前留在本機並由 `.gitignore` 排除，改由
`tools/extract_character_data.py` 從固定原版輸入重生（抽取器本身有進版控）。
要不要把例外擴及 typed data，屬權利邊界決策。

未關閉的邊界：16 槽全滿、重名、名字欄字元集合；展開跨頁貼圖殘影的貼圖矩形切法；
其餘 29 個 archive 的原版同狀態畫面收據；PCGA／HERC 轉換規則。
