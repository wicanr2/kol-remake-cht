# 角色建立證據表

本頁只整理已固定的原版證據。手冊敘述是玩家可見契約，但不能單獨證明 executable 的亂數
分布、資料欄位或所有取消邊界；這些缺口保留為未知，不由 remake 猜補。

## 固定輸入

| 輸入 | SHA-256／定位 | 用途 |
|---|---|---|
| 玩家手冊 | `4c80f5d4…7086a`；實體頁 4–5、27–31 | 建立流程、容量、職業平均值、主要／次要屬性 |
| Apple II 參考卡 | `4815bc6c…abd3`；實體頁 2 | 建立操作、取消與隊伍建議；平台差異證據 |
| `KOLBOOK.EXE` | `2cada028…e5832`；IDA 9.4 loaded MZ EA | `chardata` 路由、亂數種子、職業屬性擲骰與衍生值 |
| `CHARDATA` | `307ae9e2…394e`；12,288 bytes | 16 個 768-byte 原版角色資料槽 |
| `CHARPICS` | `d3875a6c…46c7`；12,160 bytes | 16 個 760-byte 原版角色圖片槽 |

完整 executable 定位見 `ida-fingerprints/KOLBOOK-chardata-routes-summary.json` 與
`ida-fingerprints/KOLBOOK-rng-summary.json`；含完整指令的 private 摘要與 `.i64` 只留
本機 ignored 證據區。

## 玩家可見契約

| 契約 | 證據 | 等級 | 限制 |
|---|---|---|---|
| 建立順序為名字→種族→性別→職業→屬性→圖像→保存 | 手冊實體頁 4–5 | 已證實 | DOS UI 的每個中間畫面尚未逐步 oracle |
| 名字最多 16 個字母與空格；遊戲圖像旁只顯示前四字母 | 手冊實體頁 4 | 已證實 | 大小寫、空字串、重名與非 ASCII 行為未知 |
| 種族為人類、精靈、矮人、凱爾登 | 手冊實體頁 4 | 已證實 | executable enum 值尚未定位 |
| 人類／精靈可選男性或女性；矮人／凱爾登只可選男性 | 手冊實體頁 4 | 已證實 | 不自行增加現代化例外 |
| 種族與性別限制職業；職業決定屬性可能範圍、基礎技能與初始裝備 | 手冊實體頁 4、27–29 | 已證實 | 表格提供平均值，不是亂數上下界 |
| 可不限次數重新擲骰，接受後才進圖像選擇 | 手冊實體頁 4；參考卡實體頁 2 | 已證實 | DOS UI 的 reroll／接受按鍵路徑尚未逐步 oracle |
| 圖像以左右鍵瀏覽，RETURN 接受；接受圖像前可用 ESC 放棄 | 手冊實體頁 4；參考卡實體頁 2 | 已證實 | 可選模板與職業／性別的映射未知 |
| 一個角色儲存區最多 16 名；隊伍最多六名 | 手冊實體頁 4–5；參考卡實體頁 2 | 已證實 | remake 不沿用磁碟交換 UX |
| 開始遊戲先選一名非正式領隊，再於旅店招募其他已建立角色 | 手冊實體頁 4 | 已證實 | 招募與隊伍排序另立 spec |

## 原版資料路由

- `KOLBOOK.EXE` 的 `chardata` 位於 IDA EA `0x1945a`、DGROUP-relative `0x021a`。
- `sub_134C2`（EA `0x134c2`）以 `arg_0 << 4` 選取 16-byte filename table entry；
  `arg_0=0` 路由到 `chardata`。固定 caller 位於 `_main`、`sub_11820`、`sub_1293E`、
  `sub_135DB`，安全摘要列出五個立即值為零的 call sites。
- `CHARDATA` 大小精確等於 `16 × 768`；`CHARPICS` 精確等於 `16 × 760`，與手冊的 16 名
  上限吻合。這證實 slot count／record stride，不證實 768-byte record 內欄位語意。
- 兩個全新 IDA DB 的 safe JSON 逐 byte 相同，SHA-256 為
  `e71a6ca352a0be3bd8684c29a6d3c79bf93c3efe051c3290883b684212beec5d`。

## 屬性亂數契約

- `_main:0x1001e..0x1015b` 以 BIOS `INT 1Ah` 取得低 16-bit tick `DX` 並傳給
  `_srand`。當顯示模式值為 7 時，該全域值先歸零再遞增 16 次，因此傳入 16；本切片不把
  模式值 7 擅自命名成特定顯示卡。
- `_rand` 只有 `sub_14BE4:0x14be4` 一個 caller；該 wrapper 已證實為
  `rand() % N + 1`。固定骰面表位於 EA `0x1929c`，依序為
  `[3, 4, 6, 8, 10, 12, 20, 100]`。
- `sub_11820:0x11c9d..0x11d18` 以 `profession_index × 16` 選記錄。offset `0..6`
  是七個骰式 byte `D`，offset `8..14` 是七個內部基值 `B`；offset 7／15 固定為零。
  每項主要屬性的內部值為：

  `內部值 = B + Σ roll(骰面表[D >> 4])`，擲骰次數為 `D & 0x0f`，每骰落在 `1..N`。

  先前把 offset `8..14` 的 load 漏看成同一位址、因而誤記為「骰式 byte 同時是基值」；
  原始 operand `add bx,8` 足以否定該解釋。內部輸出依序寫入 EA `0x1a474..0x1a47a`。
- `sub_12677:0x126b3..0x126c4` 顯示十項屬性時採
  `floor(內部值 × 100 / 150)`；手冊 0–100 表與內部 byte 不可直接逐值比較。
- 手冊順序與後續公式共同支持主要欄位依序為力量、敏捷、體型、健康、先見、魅力、智力；
  其中力量、體型、健康、智力可由衍生公式交叉確認，其餘三項目前列為強推論。
- 三項衍生值皆採無號整數除二、向下取整：平衡 EA `0x1a47b` =
  `(力量 + 智力) / 2`；耐力 EA `0x1a47c` = `(力量 + 健康) / 2`；身體點數
  EA `0x1a47d` = `(體型 + 健康) / 2`。這補足手冊只說「由兩項屬性計算」而未明載的
  耐力公式。
- 兩個最終全新 IDA DB 的 safe／private JSON 均逐 byte 相同；SHA-256 分別為
  `e3309787e6281c0a0837394b19fbcfc4be493b1e996898f4d695890cf3b71116` 與
  `a6e07f2a8cb5f8ce62ff795a75598eb6958317c9a9a675661f7b1d2b8ee48401`。

## BOOKDATA 職業記錄結構

- `BOOKDATA` 項目 2 壓縮 424 bytes，機械解碼為 625 bytes，SHA-256
  `848ad877…b321`；結構為 39×16-byte 記錄加一個 `0x00` 尾 byte。
- `sub_14101` 證實文字項目以 CR 分列；行 14..19 固定 race／secondary enum，行 21..53
  固定職業選單順序。selector 分組為 12、4、6、6、8、3 筆；男性／女性精靈共用六個
  選單名稱但各自使用不同的六筆記錄。記錄 28..31 分別與 32..35 逐 byte 相同，和手冊把
  八個矮人名稱合併為四列共同數值交叉吻合。
- 可重生安全摘要為 `profession-data-summary.json`；兩輪 safe／private 輸出逐 byte 相同，
  SHA-256 為 `aaa112ef…cb2d`／`7d13f6ea…9b36`。private raw records 不進版控。

## 尚未關閉的 RE 問題

1. 39 個 executable 職業記錄的初始物品／技能 writer；35 個手冊列與 39 個 menu name／record 的對應已關閉。
2. DOS `rand` 的確切 library 序列與 reroll／接受按鍵的正常 UI 資料流；目前已固定 seed 與 bounded-roll caller 契約。
3. 768-byte record 的名字、種族、性別、職業、屬性、技能、裝備與 checksum／狀態欄位。
4. 空名、重名、16 字元邊界、ESC 每層返回、角色槽已滿等正常 UI oracle。
5. 圖像模板 selector 與保存時 `CHARDATA`／`CHARPICS` 的共同 commit 邊界。
