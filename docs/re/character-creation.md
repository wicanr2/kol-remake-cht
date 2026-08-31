# 角色建立證據表

本頁只整理已固定的原版證據。手冊敘述是玩家可見契約，但不能單獨證明 executable 的亂數
分布、資料欄位或所有取消邊界；這些缺口保留為未知，不由 remake 猜補。

## 固定輸入

| 輸入 | SHA-256／定位 | 用途 |
|---|---|---|
| 玩家手冊 | `4c80f5d4…7086a`；實體頁 4–5、27–31 | 建立流程、容量、職業平均值、主要／次要屬性 |
| Apple II 參考卡 | `4815bc6c…abd3`；實體頁 2 | 建立操作、取消與隊伍建議；平台差異證據 |
| `KOLBOOK.EXE` | `2cada028…e5832`；IDA 9.4 loaded MZ EA | `chardata` 檔名路由與 caller |
| `CHARDATA` | `307ae9e2…394e`；12,288 bytes | 16 個 768-byte 原版角色資料槽 |
| `CHARPICS` | `d3875a6c…46c7`；12,160 bytes | 16 個 760-byte 原版角色圖片槽 |

完整 executable 定位見 `ida-fingerprints/KOLBOOK-chardata-routes-summary.json`；含完整指令的
private 摘要與 `.i64` 只留本機 ignored 證據區。

## 玩家可見契約

| 契約 | 證據 | 等級 | 限制 |
|---|---|---|---|
| 建立順序為名字→種族→性別→職業→屬性→圖像→保存 | 手冊實體頁 4–5 | 已證實 | DOS UI 的每個中間畫面尚未逐步 oracle |
| 名字最多 16 個字母與空格；遊戲圖像旁只顯示前四字母 | 手冊實體頁 4 | 已證實 | 大小寫、空字串、重名與非 ASCII 行為未知 |
| 種族為人類、精靈、矮人、凱爾登 | 手冊實體頁 4 | 已證實 | executable enum 值尚未定位 |
| 人類／精靈可選男性或女性；矮人／凱爾登只可選男性 | 手冊實體頁 4 | 已證實 | 不自行增加現代化例外 |
| 種族與性別限制職業；職業決定屬性可能範圍、基礎技能與初始裝備 | 手冊實體頁 4、27–29 | 已證實 | 表格提供平均值，不是亂數上下界 |
| 可不限次數重新擲骰，接受後才進圖像選擇 | 手冊實體頁 4；參考卡實體頁 2 | 已證實 | RNG 演算法、種子與分布未知 |
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

## 尚未關閉的 RE 問題

1. 七項主要屬性的職業亂數上下界、分布、seed 與 reroll 資料流。
2. 35 個職業的 executable enum、種族／性別 eligibility table 與初始物品／技能 writer。
3. 768-byte record 的名字、種族、性別、職業、屬性、技能、裝備與 checksum／狀態欄位。
4. 空名、重名、16 字元邊界、ESC 每層返回、角色槽已滿等正常 UI oracle。
5. 圖像模板 selector 與保存時 `CHARDATA`／`CHARPICS` 的共同 commit 邊界。
