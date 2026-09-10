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
| `SHLDPICS` | 固定 ZIP；12,160 bytes | 16 個 760-byte 原版角色盾牌圖片槽 |

完整 executable 定位見 `ida-fingerprints/KOLBOOK-chardata-routes-summary.json` 與
`ida-fingerprints/KOLBOOK-rng-summary.json`；含完整指令的 private 摘要與 `.i64` 只留
本機 ignored 證據區。

角色保存 writer 的安全摘要另見
`ida-fingerprints/KOLBOOK-character-writer-summary.json`；KOLTOWN 最小 reader 摘要見
`ida-fingerprints/KOLTOWN-character-reader-summary.json`。
角色建立返回／提交邊界另見
`ida-fingerprints/KOLBOOK-character-ui-summary.json`。

## 玩家可見契約

| 契約 | 證據 | 等級 | 限制 |
|---|---|---|---|
| 建立順序為名字→種族→性別→職業→屬性→圖像→保存 | 手冊實體頁 4–5 | 已證實 | DOS UI 的每個中間畫面尚未逐步 oracle |
| 名字最多 16 個字母與空格；遊戲圖像旁只顯示前四字母 | 手冊實體頁 4；KOLBOOK `0x11907..0x11957`；`docs/verification/character-ui/name-length-run{1,2}.json` | 已證實 | DOS UI 實測容量為 16 個可見字元，輸入第 16 個字元即直接前進到種族層，不需 Return；空名按 Return 不前進。字元集合與大小寫保存方式仍未知 |
| 種族為人類、精靈、矮人、凱爾登 | 手冊實體頁 4 | 已證實 | executable enum 值尚未定位 |
| 人類／精靈可選男性或女性；矮人／凱爾登只可選男性 | 手冊實體頁 4 | 已證實 | 不自行增加現代化例外 |
| 種族與性別限制職業；職業決定屬性可能範圍、基礎技能與初始裝備 | 手冊實體頁 4、27–29 | 已證實 | 表格提供平均值，不是亂數上下界 |
| 可不限次數重新擲骰，接受後才進圖像選擇 | 手冊實體頁 4；參考卡實體頁 2；KOLBOOK `0x11da9..0x11e00` | 已證實 | 選項 0 重擲、取消返回職業；正常 DOS UI 截圖仍待完成 |
| 圖像以方向鍵瀏覽，RETURN 接受；接受圖像前可用 ESC 放棄 | 手冊實體頁 4；參考卡實體頁 2；KOLBOOK `0x11f6a..0x11ff1` | 已證實 | 原版接受左右／上下鍵並在 0..79 循環；模板與職業／性別映射未知 |
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
  最新 SHA-256 為 `ae711e3b…2176`／`f42fbbf1…97bc`。private raw records 不進版控。

## 起始金冠、武器與武器技能

- KOLBOOK 的 39×2-byte 表位於 EA `0x1940c`；writer 將選定職業的 little-endian word
  寫入角色記錄 offset `0x2b..0x2c`。39 筆值全部與中文手冊職業表的起始金冠相符。
- EA `0x19348` 是 39×5-byte 職業 metadata。byte 0 低半位元組為主要武器 selector，
  高半位元組為可選的次要武器 selector；selector 為一基底。`BOOKDATA` 項目 3 解成
  14×16-byte 武器模板加一個零尾 byte，依序為
  `LO, BA, LB, GS, LC, WH, SC, HL, BR, SB, MA, SS, LS, EB`。
- metadata byte 3／4 是主要武器的兩個內部技能值，byte 1／2 是次要武器的兩個內部技能值；
  手冊顯示值為內部值除以十。39 筆金冠、主要／次要武器與技能均通過中文手冊逐列驗證。
- `BOOKDATA` 項目 4 解成 760-byte 初始盾牌模板加一個零尾 byte。這不證實盾牌的遊戲內
  裝備名稱或數值。

## 最小角色寫入欄位索引

| CHARDATA offset | 寫入內容 | 等級／限制 |
|---|---|---|
| `0x00..0x0f` | 名字 16 bytes | 已證實；字串終止與邊界 UI 仍待 oracle |
| `0x1a` | secondary selector | 已證實；只在人類／精靈可對應性別，不泛稱 gender |
| `0x1c` | profession record index | 已證實 |
| `0x1d..0x23` | 七項主要屬性內部值 | 已證實；欄位語意分級沿用上節 |
| `0x24` | race enum | 已證實 |
| `0x28` | `(race enum + 1) × 16` selector | 已證實寫法；下游完整語意未知 |
| `0x29` | `floor(四個職業武器技能 bytes 總和 ÷ 200)` | 已證實寫法；下游完整語意未知 |
| `0x2b..0x2c` | 起始金冠 | 已證實，little-endian |
| `0x30..0x3f` | 主要武器 16-byte 模板 | 已證實 |
| `0x4a..0x51` | 主要武器模板 bytes `0..7` | 已證實；前版「輔助區塊」名稱已訂正 |
| `0x52` | 主要武器模板 byte `13` | 已證實 |
| `0x54`, `0x56` | 主要武器兩個內部技能值 | 已證實 |
| `0x57..0x5e`, `0x5f` | 次要武器模板的已寫入部分 | 已證實；不是完整 16-byte copy |
| `0x61`, `0x63` | 次要武器兩個內部技能值 | 已證實 |
| `0x7f` | 常數 `3` | 寫入已證實；玩法語意未知 |

- writer 清空 768-byte slot 後寫入欄位；選定肖像與初始盾牌各以 760 bytes 寫入同 slot 的
  `CHARPICS`／`SHLDPICS`，最後依序保存 `0x3000`、`0x2f80`、`0x2f80` bytes。這證實
  原版順序，不證實跨三檔案的 atomicity 或 rollback。
- `region%d.dat` 最多載入五筆、每筆 48 bytes；固定 ZIP 只有 `REGION0.DAT`——其餘四筆是
  原版未曾發行的擴充地區（見 [地區擴充片](20260911-region-installer.md)）。固定來源與
  CHARDATA 目的 offset 的路由已收入安全摘要，但欄位語意保持未知。
- 最終兩個全新 IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  最新 `2884a1ae…c129`／`6c00db3a…fdb6`。

## KOLTOWN 最小 reader 交叉查核

- 固定 unpacked KOLTOWN derivative SHA-256 `01f6f9…d1d2`；IDA Pro 9.4 linear EA。
- `sub_147DD:0x14fad..0x14fc1` 從載入的 CHARDATA 逐槽複製 16-byte 名字，
  `0x14fcb..0x14fdd` 再以名字比對目前角色，交叉確認 offset `0x00..0x0f`。
- `sub_14DB1:0x14de7..0x14e23` 讀 offset `0x24`／`0x1c` 作種族／職業 eligibility；
  `sub_14CCF:0x14cf1..0x14d1e` 將 `0x2b..0x2c` 合成 little-endian 金冠；
  `sub_18F3A:0x19167..0x1917e` 讀 `0x30` 判斷主要武器並進入 renderer route。
- writer 在先清空 768-byte slot 後，所有直接 CHARDATA 寫入都已由固定欄位表與 region
  routing 涵蓋。因此 `sub_11820` 沒有其他職業衍生的起始技能／物品寫入；這不代表
  768-byte record 的所有後續遊戲欄位已解析。
- 兩個全新 KOLTOWN IDA DB 的 safe／private JSON 逐 byte 相同，SHA-256 分別為
  `16f8b24d…67ba`／`dcfa8c59…18e1`。

## 角色建立 UI 控制流

- 名字層以 `0x10` 傳入輸入 helper；回傳後第一 byte 為零且未取消時回到同一輸入點，取消
  則直接離開角色建立。正常 DOS UI 收據與這條分支一致：空名按 Return 會停在名字層，
  輸入滿 16 個字元則不需 Return 直接前進到種族層。
- 種族取消回名字；secondary 選單取消回種族。職業取消在人類／精靈 selector 返回
  secondary 選單，其他 selector 返回種族。這是控制流證據，不把 secondary 欄位泛稱成
  所有種族都存在的性別欄位。
- 屬性選單結果 0 回到完整擲骰起點；取消返回職業層。
- 肖像索引為 `0..79`：左／上遞減並由 `-1` 回繞至 79，右／下遞增並由 80 回繞至 0；
  Return 接受，取消回到 commit 前的職業流程。
- slot finder 未命中名字後，以 `0x300` stride 掃描到 `0x3000`。16 槽全滿時顯示訊息並在
  `0x12061` 離開；角色記錄清零從 `0x12064` 才開始，因此滿槽分支不修改待寫記錄。
- 探針以兩個全新 IDA Pro 9.4 DB 重建，safe／private 逐 byte 相同；SHA-256 分別為
  `9ecd4b76…aa82`／`ef84a41e…50c`。`word_19290` 的「取消旗標」是多層返回用途一致所得的
  已證實語意，匯出仍保留原始名稱與 EA。
- 正常玩家入口為 `KOL.EXE`：TITLE 標題序列為循環動畫，按任意鍵離開後進入 KOLBOOK
  書本封面，再按任意鍵翻到目錄頁，按 `5` 進 Character options，反白條下移一格選
  `CREATE CHARACTER` 進入 New Character 頁。整條路徑的逐畫面收據見
  `docs/verification/character-ui/`。
- 動態收據證實的取消目的地與靜態控制流一致：種族→名字、性別→種族、職業→性別、
  屬性→職業，再往上回到性別、種族、名字層。回到名字層時名字與右頁摘要欄位皆已清空。
- 屬性頁的 `REROLL STATS` 會重擲九個顯示值；同一 HUMAN／MALE／BARBARIAN 角色兩次
  reroll 的力量為 88 → 85 → 89。四個 UI 情境全程未接受屬性或圖像，三個角色檔的
  前後 SHA-256 均未變動。

## 尚未關閉的 RE 問題

1. 已排除 `sub_11820` 另有職業衍生的起始技能／物品寫入；region metadata 的玩法語意仍未知。
2. DOS `rand` 的確切 library 序列；reroll／接受控制流已固定，正常 UI 截圖仍待完成。
3. 768-byte record 其餘、非角色建立必要欄位的 reader／writer 與 checksum／狀態語意；
   角色建立最小 reader／writer 索引已交叉查核。
4. 重名、名字接受的字元集合與大小寫保存、角色槽已滿的正常 UI oracle。空名、16 字元
   邊界、ESC 每層返回與 reroll 已有兩輪可重生 receipt；滿槽只有靜態控制流證據。
   固定 ZIP 的 `CHARDATA` 十六槽都非全零，前十四槽有名字，第 15、16 槽名字為空但
   記錄非零，槽位可用性的 UI 判準尚未確認。
5. 圖像模板 selector；三個檔案的寫入順序已證實，但失敗、rollback 與共同 commit 邊界未知。
