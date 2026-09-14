# SPEC-031：資料頁物品選中後的明細層

狀態：**CONFORMED**（證據審查與同狀態驗證：2026-09-14）

本規格只處理資料頁（`sub_16986`）選中一件非空物品後的畫面重畫，以及與這個
畫面直接相鄰的空格訊息、動作訊息與返回生命週期。它不是整個物品資料模型或
動作規則的重製規格。

## 1. 範圍與排除

納入：

- 選中物品後的 320×200 英文頁面圖層順序。
- 立繪、木屋背板、十七列清單與外框在選中後是否仍存在。
- 明細框 `(62,58)`–`(257,125)` 的填色、框線與右／下四像素投影。
- 空格訊息、動作完成訊息、送禮候選人與 U-turn 返回所需的既有狀態轉換。
- 中文模式沿用同一狀態分層並以 `book.CJKScale` 放大；中文沒有原版 oracle，
  不宣稱中文像素 parity。

排除：

- 物品記錄格式、十七格索引、穿戴／丟棄／送禮／食用規則本身。
- 重新設計英文或繁體中文文字、資料頁座標、圖示列或玩家操作方式。
- 未取得公開散布權的原版素材與執行檔。

## 2. 證據輸入與定位契約

| 輸入／工具 | 定位與雜湊 | 用途 | 證據等級 |
|---|---|---|---|
| `KOLTOWN-unpacked.EXE` | SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`；IDA 線性有效位址，重建 unpacked MZ | 原版控制流與呼叫順序 | 已證實 |
| `workplace/analysis/koltown/KOLTOWN-unpacked.EXE.asm` | IDA Pro 9.4、`ida-pro-9.4-idapython:locked-v1`；保留 `sub_16986`、`sub_1BA9D`、`sub_1B7E9` 原名 | 可回查的反組譯證據 | 已證實 |
| `workplace/analysis/koltown/mark-box.json` | 同一輸入 SHA、IDA 9.4；`sub_1BA9D` `0x1BA9D..0x1BAFF`、`sub_1B7E9` `0x1B7E9..0x1B8B8` | 函式邊界與呼叫端索引 | 已證實 |
| `docs/re/RESEARCH-LOG.md` RE-127、RE-137 | 物品頁控制流、`sub_1BA9D` 複合框與原版畫格觀察 | 證據交接與既有勘誤 | 已證實 |
| `docs/verification/character-ui/town-sheet-items-run1/frames/frame-070.png`、`frame-072.png` | 原版 ZIP SHA-256 `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`；DOSBox-X 2026.06.02、fixed 3000 | 目前可重生的輔助畫面 oracle | 已證實（輔助；非 dosgolem 收據） |

反組譯中的位址都是 **IDA linear effective address for reconstructed unpacked MZ**；
不能與檔案偏移或其他工具的虛擬位址混用。新增語意保留原始函式名、位址與
operand；本規格沒有把推測性名稱寫回 IDA 資料庫。

## 3. 原版行為

### 3.1 完整物品頁

`sub_16986` 的完整頁分支在 `0x16995` 先呼叫 `sub_1103F(0)`，再依
`word_2DF2C` 呼叫 `sub_1103F(0x20)` 或 `sub_1103F(0x21)`／`sub_1103F(0x64)`，
並呼叫 `sub_18F3A(0)` 貼立繪。主迴圈 `0x16A11` 另外畫右側清單欄、外框、分隔線、
外框投影、動作圖示與 `sub_196A9(..., 0x11, ...)` 的十七列清單。

### 3.2 選中非空物品

在 `sub_16986:0x16BBF`，原版：

1. 呼叫 `sub_1103F(0)`，覆蓋完整 320×200 畫布；這一步之後，先前的立繪、木屋
   背板、十七列清單、外框與分隔線不再是畫面內容。
2. 以五個參數呼叫 `sub_1BA9D`：`(0x3E, 0x3A, 0x101, 0x7D, 0x0F)`，其幾何範圍
   是 `(62,58)`–`(257,125)`。
3. 設定 `word_2DF1A=0x0F`、`word_39A56=選中列`；非空記錄複製 16 bytes 後，
   呼叫 `sub_153FF(8)` 與 `sub_18857(選中列)` 印明細並建立動作圖示列。

這條分支沒有再次呼叫 `sub_1103F(0x20)`／`(0x21)`／`(0x64)`、`sub_18F3A(0)`、
`sub_196A9` 或完整頁的外框／分隔線序列。這是「選中後只剩木紋底圖與明細層」的
控制流證據，不是由畫面外觀反推。

### 3.3 `sub_1BA9D` 的複合框

`sub_1BA9D:0x1BA9D..0x1BAFF` 的順序是：

1. 以內縮一像素的參數呼叫 `sub_1B779` 填滿內部。
2. 以原始五個參數呼叫 `sub_1B6DC` 畫框線。
3. 以前四個參數呼叫 `sub_1B7E9`；`sub_1B7E9` 在其 `0x1B7E9..0x1B8B8` 內，
   以 `+4` 的右側與下側迴圈畫出各四像素的投影。

因此「明細框沒有投影」與原版控制流矛盾；這不是待猜的美術選擇。

### 3.4 空格與動作返回

空格會從 `0x16C66` 起印出 `0xCA`、`0xCB`／`0xCC` 與部位名稱，最後在
`0x16CE7` 等待一次按鍵；它與非空明細共用前面的 `sub_1103F(0)`，但不代表要
把選中後的十七列清單重新畫回來。

非空動作回到 `0x16CEC` 檢查 `byte_2DF50`；非零時清零並跳回 `0x16995`，也就是
重新建立完整頁。這是動作完成後的 full redraw，不能與選中當下的 blank-base
分支混用。

## 4. 重製版規格

### 4.1 畫面分層

英文 `drawItemPage` 與中文 `drawItemPageCJK` 都使用下列狀態判斷：

```text
interactionBase = p.Acting || p.Detail != nil || len(p.Message) > 0
```

- `interactionBase == false`：保留既有完整物品頁流程——整頁底圖、木屋背板、立繪、
  右欄、外框／分隔線／投影、十七列與左上小框。
- `interactionBase == true`：只先畫 `drawSheetPageFull`（`sub_1103F(0)` 對應的
  全頁底圖），不畫木屋背板、立繪、右欄、十七列、左上小框或完整頁外框。
- 明細／訊息／Note／送禮候選人的既有文字狀態與輸入結果不改；它們繼續在目前
  的明細區域繪製。`p.Note` 且 `p.Acting == false` 的動作完成後狀態走完整頁，
  對應原版 `byte_2DF50` 的 full redraw。

### 4.2 明細框

繪製 `l.Detail` 時固定依序執行：內部填色、`frame`、`shadow`。`shadow` 的右側與
下側必須各為四像素，使用現有 `town.Shadow`／`shadow` 原語；不得以文字裁切或
全頁差分例外取代。英文框仍為 `(62,58)`–`(257,125)`，中文框為同一份
`town.ItemPageLayout.Detail` 乘 `book.CJKScale`。

英文明細文字使用既有 `book-large` 8×8 字模；frame-072 的第一段字形起於
`y=66`，由 `goodDetail` 的原始列 `102` 經模式 8 的文字基準 `50` 換算取得。
這項位置與字模高度由 frame-072 像素列量測交叉確認；不是為了避開差異而加入的
允許區。

### 4.3 玩家狀態與保存

不改 `internal/town/itempage.go` 的 `AcceptItemSlot`、`afterItemAction`、
`closeItemAction` 或保存格式。選中、空格、動作、給隊友、U-turn 都必須仍由正常
玩家按鍵進入；截圖腳本只固定輸入，不得使用 direct-entry、座標注入或強制勝利。

## 5. 驗收契約

1. 既有 frame-070 路徑 `L R RET SPACE > RET > RET D` 必須維持通過；它保護完整
   物品頁、十七列與左上小框。
2. 新增正常路徑 `L R RET SPACE > RET > RET D RET`，產生選中後 320×200 PNG，
   與原版 `frame-072.png` 整頁逐點比較。
3. 比較只可明列原版滑鼠游標矩形 `(158,100,14,14)`；frame-070 的既有
   `(15,7,1,1)` 灰點只屬完整清單 pair 的允許區。選中 pair 的明細框四邊與
   右／下四像素投影列為必須相同區，允許區外差異必須為 0。
4. 收據記錄重製建置／輸入腳本、兩側 PNG SHA-256、工具版本、尺寸、比較矩形、
   允許區、區外差異、首個差異座標，以及 dosgolem 能力檢查結果。DOSBox-X 畫面
   若仍是唯一原版輸入，標為輔助／bootstrap，不冒稱 strict dosgolem parity。
5. 回歸至少包含：完整頁、選中後清除立繪／清單、四像素投影、明細、Note、Message、
   Giving 與返回；再跑相關 Go 測試、完整測試與 Docker 內建置／GUI smoke。

## 6. 生命週期

2026-09-14 證據審查已完成：

- `sub_16986` 選中分支的 `0x16BBF`、`sub_1BA9D` 呼叫與 `0x16CEC` 返回控制流；
- `sub_1BA9D:0x1BAF1` 對 `sub_1B7E9` 的直接呼叫及四像素幾何；
- dosgolem 的能力檢查結果為 `false`：目前沒有 KOL 專屬適配器，因此 frame-072
  明確分類為 DOSBox-X `dosbox-bootstrap`，不作 strict dosgolem parity；
- 正常玩家路徑、CJK 非 parity 限制與現有 itempage 狀態測試的涵蓋範圍。

以上證據足以授權本切片實作，但不授權把 DOSBox-X 圖片包裝成 dosgolem 收據。
實作與驗收已完成：`internal/ui/itempage.go`／`itempage_cjk.go` 依上述分層繪製，
英文 frame-070／frame-072 正常玩家路徑與原版逐點收據均通過，明細框四邊及右／下
四像素投影列入必須相同區；相關 Go 產品測試、建置與 CJK 文字安全檢查也通過。
收據為 `docs/verification/remake-parity/receipt-town-sheet-items.json`，其中
frame-072 的允許區外差異為 0。由於目前 dosgolem 沒有 KOL 適配器，原版畫面仍是
DOSBox-X `dosbox-bootstrap` 輸入；本規格不宣稱 strict dosgolem parity。未通過的
差異仍應回到 RE／DRAFT，不在程式中默默增加例外。
