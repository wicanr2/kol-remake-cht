# 《武士傳說》重製目前狀態

更新日期：2026-09-04

**自評 82 / 100**。算法、權重與逐系統的完成度依據都在
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
  **一場戰鬥雙向打得完了**（RE-091／RE-092）：敵人會照原版的 AI 接近、挑目標、
  選攻擊式與命中部位；每個回合結束數兩邊倒下的人，依四個條件判勝負，
  打輸與逃走各有自己的訊息。
  **戰後結算做完了**（RE-096）：冒險點數是每一個參戰敵人的攻防技能加總除以 2
  再除人數，金冠只讀第一隻怪的骰式，掉落物與戰鬥中丟在地上的東西一起排成
  「Items Found:」清單讓玩家一件一件領走。整條路徑實跑過
  （`docs/verification/battle-spoils/`，用了觀察用的替代輸入）。
  **野外的隨機遭遇打得起來了**（SPEC-011／RE-097…099）：時段一變就擲一次，
  遇到就進戰場——機率、挑怪、隻數、等級、騎走判定、兩邊站位與逐格擲出來的地形
  都對得上原版。實跑收據 `docs/verification/battle-ambush/`。
  **施法、射擊與逃走也做完了**（SPEC-012／013／014）：法術從角色記錄讀、
  兩張清單走原版的版面；射擊拉一條 Bresenham 彈道，樹扣命中值、牆停下、半路的人
  各扣體型的一半，敵人在近距離也會射；每個回合結束若有人帶著任務要的東西站到戰場
  最外圈就問「要逃嗎」，答應就走——**帶著東西逃掉一樣記任務完成**（RE-101…104）。
  SPEC-001／002／003／005／006／007／008／009／010／011／012／013／014 為 READY
  或 CONFORMED（008、009、010 只有已解出的段落 READY），SPEC-004（音訊）仍為 DRAFT。
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

**任務的回報與完成**。戰鬥這一側從指令到結算、從遭遇到戰利品、從近戰到法術都通了
（SPEC-008…012），但任務打完之後沒有下文——怎麼算完成、回去跟誰交、旗標怎麼清，
一條 RE 都還沒有。

之後依序是敵人 AI 剩下的（撤退、反制掃描、側翼、射擊）、施法與投擲、撿那一句的
問話。完整清單見 `WORKLIST.md`。

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
