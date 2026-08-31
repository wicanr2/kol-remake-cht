# 《武士傳說》重製目前狀態

更新日期：2026-08-31

## 目前真相

- 本專案是 *Knights of Legend* 的繁體中文乾淨重寫，正式 runtime 使用 Go／Ebitengine、
  PNG 圖像與 Ogg Vorbis 音訊。
- GitHub `wicanr2/kol-remake-cht` 已建立為 private；原版與衍生素材不得提交。
- remake 尚未開始。RE 知識庫已建立，SPEC-001 已 READY，只授權固定 KOLTOWN derivative
  作研究 oracle；任何玩法、格式或玩家路徑仍須自己的 DRAFT／READY spec。
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
- 中文手冊 SPEC-003 已 READY：Poppler 22.12.0 對 79 頁玩家手冊與 5 頁參考卡建立逐頁
  雜湊清冊；82 個內容頁都有文字層，玩家手冊 76、77 經頁面 PNG 目視為空白跨頁。
  玩家手冊實體頁 1–55 已逐頁對照掃描並翻譯為 Markdown；目前 reviewed 57、todo 27。
- 圖片 RE 已建立 SPEC-002 DRAFT：44 個實體候選歸併為 30 個邏輯 archive；TWN `.1/.2`
  已證實是 continuation volume，總數由 960 訂正為 1,022 個 payload，全部通過 LZW 機械
  探針。索引／分卷契約已證實，LZW 為強推論；
  尺寸、位平面、palette、透明色與原版視覺仍未知，尚未授權 PNG production 抽取。
- `AWD1PICS.0` 是 KOLBOOK 固定寫出的 4000-byte 工作區快照；KOLTOWN 只讀其中索引
  payload。尾端 520 bytes 已排除於圖片流，但仍是未解析工作區資料，不是 padding。
- `CHARPICS`／`SHLDPICS`／`STEEDPIC` 已由 KOLBOOK／KOLTOWN 讀寫資料流證實為 16 筆、
  每筆 760 bytes 的固定可變角色圖片記錄表；記錄內像素佈局尚未證實。
- COMBPICS EGA 分支已由 IDA 關閉到 296×144、二階 RLE、4-bit 色彩與 BIOS mode 0Dh
  palette；兩張本機測試 PNG 分別對上魔旅士文章的同場景原版 JPEG。這只證實該窄分支，
  不代表全部圖片格式完成。
- TWN 全域圖號 24 已關閉 continuation 垂直鏈：八城均從 `.1` 讀取；town 0 為 192×96，
  其餘為 88×64。八張 ignored PNG 與接觸表通過像素數、4-bit bounds、方向及非空目視。
- MONPICS 已關閉 32 組怪物 sprite 垂直鏈：64 個 payload 依序配成 4-bit EGA 彩圖與
  MSB-first 1-bit 遮罩；IDA 尺寸表、逐張像素數與透明接觸表均吻合。遮罩有效資料後多出的
  單一 LZW byte 不作像素，仍標為壓縮終止／padding 的強推論。
- 固定 ZIP 一般圖片只有 EGA `PICS`，但 `MAKEMODE.EXE` 已證實是官方 CGA／Hercules
  轉換器：它對 28 組 PICS archive 生成 `PCGA`／`HERC` 分卷，28 組圖數與現行清冊完全
  相符。完整轉換像素規則尚未動態驗證；三種 theme 的產品範圍待使用者決定。

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

等待使用者決定 EGA／CGA／Hercules theme 交付範圍；不依賴該決定的下一工作是依
SPEC-003 翻譯玩家手冊實體頁 56–60。MONPICS 已完成，不重開；SPEC-002 READY 前不接 runtime。
