# RE-262：受阻打暈訊息的舊地面名稱

日期：2026-10-06。狀態：有限來源與既有畫格核對通過。正式程式未改。

路由命中 remake 收尾與 IDA 9.4。工具入口為
`/home/anr2/ida_94_official/skills/use-ida-pro-9-4/SKILL.md`，
工具契約為該專案的 `knowledge-base/ida-94-tools.md`。
本輪只研究已登錄的 H 訊息欄，不重跑整場戰鬥或延伸硬體研究。

## 輸入與定位

| 項目 | 值 |
|---|---|
| 原始輸入 | `workplace/analysis/koltown/KOLTOWN-unpacked.EXE` |
| 輸入 SHA-256 | `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 正式資料庫 | `KOLTOWN-unpacked-ida94-derived.i64`，唯讀掛載 |
| 資料庫 SHA-256 | `5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8` |
| 工具 | IDA Pro 9.4，`ida-pro-9.4-idapython:locked-v1` |
| Image ID | `6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780` |
| 位址基準 | IDA 線性位址，載入基底 `0x10000`；MZ header 9216 bytes |
| H | 原始名稱 `word_30D02`，IDA `0x30D02`，DS 位移 `0x2E32` |
| 地面記錄 | DS 位移 `0x9D20` 起，16 槽，每槽 16 bytes |

## 已證實的讀取鏈

`sub_265F3:0x26769` 讀 H，四次左移得到槽位乘 16。
`0x26774` 加上 `0x9D20`，`0x2677A` 呼叫 `sub_1B94F(ptr,1)`。
這條鏈讀地面槽，不讀受害者手持記錄。

名稱解碼器 `sub_1B94F:0x1B960` 先讀長度，高四位為 mode 1 的字數。
`0x1B973` 從第一個 byte 的 bit 3 開始讀五位元碼。
它先解碼再減長度，沒有先檢查零長度。
`0x1BA2E` 將零碼寫成空白，`0x1BA49` 限制輸出到 24 字元。
因此本次全零地面記錄產生 24 個空白，沒有預設「Short Sword」名稱。
僅外推到此全零輸入，不替其他損壞或殘留記錄訂正式規則。

兩函式的 142 道指令均與重定位後原始 bytes 相符。
新查詢在容器內的 `/tmp/name-query.i64` 一次性副本執行，沒有改名或註記正式資料庫。

## 既有正常畫格核對

| 來源 | H 與地面槽 | 原版可見結果 | 字模差異 |
|---|---|---|---|
| RE-253 第 317 格 | H=0，全零槽 | 物品名稱空白，第二行為 `FALLS TO THE GROUND!` | 0，矩形 24／174／160／6 |
| RE-260 第 730 格 | H=2，`a149e124eef910c020a028ff0f5d9401` | `BROADSWORD FALLS TO THE GROUND!` | 0，矩形 24／174／248／6 |

兩個正常樣本都在執行前固定 setup seed 2，沿既有 dosgolem
`968ca774705853276ae55c8d1163c369481ac71f` 收據。本輪沒有新增原版執行或 GUI 重播。
比較點 seed 分別為 2126174334、3128552693；名稱由完整地面 bytes 推導。
沒有重擲、挑新種子、換畫格來源或注入狀態。

字模用現行原版小字 PNG 與 manifest，逐像素比較六列黑字白底，兩個矩形均零差。
同位置錯填 `SHORT SWORD` 會差 143 像素，錯填 `BATTLE AXE` 會差 125 像素。
這是既有畫格的有限文字核對，不宣稱現行 remake 訊息、完整 UI 或整場戰鬥相同。

## RE-253 追加勘誤

RE-253 曾寫「畫面卻顯示 Short Sword 落地」。原始第 317 格 PNG
實際沒有該名稱，完整地面記錄也全零。
本輪來源 bytes、零長度控制流與字模負對照足以否定該句。
原始收據、當輪文件與錯誤敘述保留，另追加此勘誤。
受阻保留精靈弓及 ReturnScratch=6 的已驗結論不受影響。

RE-260 的 Kenneth 保留戰斧，訊息卻讀到舊地面闊劍，兩者可以同時成立。
現行正式版與既有受阻候選仍把此 H 名稱留空，非空舊槽的呈現缺口仍在。
本輪解除原版讀取語意未知，沒有解除實作與正常重播閘門。
受阻持物規則沿既有問題等待確認，保持 DRAFT。

## 重生與限制

本機工作區為 `workplace/stale-ground-name-20261006/`。
`probe.py` 匯出原始指令，`review.py` 重算來源 bytes、兩個畫格及現行 338 輸入。
`verify.py` 以唯讀輸入重算收據，並核對附件、擁有權與 PNG 校驗碼。
只用上述 IDA image 與 `eob-remake-go:1.26.7-ebiten2.9.9-audio`，
Docker 掛載 `/src` 唯讀、研究輸出 `/out` 可寫，UID/GID 1000:1000、無網路且有資源與逾時限制。
原始 bytes、畫格及來源快照僅本機，不擴張 PNG 的 private 儲存庫例外。

338 正式來源與 Linux `1f31007c19ceac875a12df9b68d8629296764d01194983cf98d10cc8dcdbd274`
保持。沒有新增保存、平台或封包驗收，也沒有新的美術與發行決定。
HD 仍只確認手繪媒材，其他既有問題保持。

收據：[receipt.json](../../workplace/stale-ground-name-20261006/receipt.json)。
重驗：[verify.py](../../workplace/stale-ground-name-20261006/verify.py)。
