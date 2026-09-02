# remake 對 DOS EGA 原版的畫面對拍

更新日期：2026-09-01

繁體中文重製不可能在文字上逐像素等於原版英文字模，因此對拍分兩層，而不是一個
「像素相同／不同」的是非題：

- **素材與版面層**：允許差異區以外的每個像素都必須相同。差一個像素就是失敗。
- **允許差異區**：由計畫明文列出並附理由（原版畫的滑鼠游標、原版自繪的文字區）。
  允許不等於忽略——收據會記錄每個允許區內實際差了多少像素，數字異常變動就看得出來。
- **必須相同區**：優先於允許差異區，用來在「文字可以不同」的大片區域裡仍然釘住素材
  必須逐像素相同的矩形，例如角色圖像的 blit 位置。

## 產生方式

```
tools/run_kol_headless.sh -shot-dir docs/verification/remake-parity/shots
python3 tools/verify_remake_parity.py \
    --plan docs/verification/remake-parity/plan-book-screens.json \
    --out docs/verification/remake-parity/book-screens.json

tools/run_kol_headless.sh -script "t e s t RET RET RET RET" \
    -shot-dir docs/verification/remake-parity/shots/newchar-steps
python3 tools/verify_remake_parity.py \
    --plan docs/verification/remake-parity/plan-newchar-steps.json \
    --out docs/verification/remake-parity/newchar-steps.json

tools/run_kol_headless.sh \
    -script "z e d RET RET RET RET DOWN RET LEFT$(printf ' RIGHT%.0s' $(seq 81))" \
    -shot-dir docs/verification/remake-parity/shots/newchar-pictures
python3 tools/verify_remake_parity.py \
    --plan docs/verification/remake-parity/plan-newchar-pictures.json \
    --out docs/verification/remake-parity/newchar-pictures.json
```

remake 以 Go 1.26.7 ＋ Ebitengine v2.9.9 建置於固定映像
`eob-remake-go:1.26.7-ebiten2.9.9-audio`，在容器內以 Xvfb ＋ Mesa llvmpipe 軟體算繪，
輸出的是 320×200 原版解析度離屏畫布，因此不受視窗縮放與 GPU 差異影響。

原版側畫面來自 `../character-ui/`，是 `KOL.EXE` 正常玩家路徑的擷取。

## 目前結果

| 收據 | 對數 | 通過 |
|---|---:|---:|
| `book-screens.json`（封面、目錄頁、空白 New Character 頁） | 3 | 3 |
| `newchar-steps.json`（名字→擲屬性，對 `reroll-run1`） | 9 | 9 |
| `newchar-pictures.json`（完整流程＋80 張圖像循環，對 `picture-cycle-run1`） | 92 | 92 |
| `book-open.json`（封面翻到目錄頁，對 `accept-and-save-run1`） | 18 | 18 |
| `book-menu.json`（目錄頁翻到角色選項頁，對 `accept-and-save-run1`） | 16 | 16 |
| `book-roster.json`（角色選項頁翻到角色清單，對 `list-characters-run1`） | 20 | 20 |
| `delete.json`（刪除流程含確認視窗，對 `delete-character-confirm-run1`） | 8 | 8 |
| `delete-move.json`（刪除清單反白條移動，對 `delete-character-run1`） | 5 | 5 |
| `harbor.json`（HARBOR MASTER 訊息頁，對 `harbor-master-enter-run1`） | 1 | 1 |
| `town-inn.json`（旅店畫面，對 `town-icon-move-run1`／`town-recruit-companion-run1`） | 8 | 8 |
| `town-outdoor.json`（戶外地圖與門口，對 `town-outdoor-move-run1`／`town-icon-bar-run1`／`town-enter-building-run1`） | 22 | 22 |
| `town-building.json`（建築進門，對 `town-enter-building-run1`） | 1 | 1 |
| `town-edge.json`（地圖邊界的城門提示，對 `town-map-edge-run1`） | 4 | 4 |
| `town-listen.json`（招呼語與 LISTEN，對 `town-listen-run1`） | 3 | 3 |
| `world.json`（野外地圖畫面，對 `town-leave-run1`） | 4 | 4 |

十五份收據 214 對畫面，允許差異區之外都是**零像素差**。

`world.json` 多了一個欄位 `compare`：每一對只涵蓋畫面上的一塊矩形。這**不是**
允許差異區——矩形之外的部分那一對不宣稱，而計畫的 `not_claimed` 要寫明為什麼。
目前唯一一項是天候圖 (16,76) 64×62：原版的雲層每個 tick 旋轉一次，
緩衝區的位元組版面還沒定案（RE-068），remake 畫的是未旋轉的合成圖。

| 對拍項目 | 允許差異區實際差異 | 結果 |
|---|---|---|
| 書本封面 | 原版滑鼠游標 50 px | 通過 |
| 目錄頁 | 原版滑鼠游標 55 px | 通過 |
| 空白 New Character 頁 | 名字欄殘留字元 31 px | 通過 |

- 書本封面與目錄頁在游標以外**完全相同**：這兩個畫面在原版是整張烘好的 BOOKPICS
  素材，remake 直接畫同一張抽取 PNG。
- 80 張 New Character 畫面的角色圖像 blit 矩形 (208, 106, 40×38) 列為必須相同區，
  80 對全部逐像素相同。
- **空白 New Character 頁是第一個含文字的逐像素對拍**：remake 以抽出的原版字型、
  機械量得的版面座標與原版的繪字規則（逐字模格清白底、空白字元不清底）重繪整頁。

## 角色建立流程逐畫面對拍

remake 以固定腳本 `t e s t RET RET RET RET`（seed 1）重播名字→種族→性別→職業→
擲屬性五層，九張畫面對原版 `character-ui/reroll-run1` 的同狀態畫格。文字、選單、
標籤高亮與輸入游標全部由狀態機畫出來，不是預先烘好的圖。

| 步驟 | 原版畫格 | 允許差異區實際差異 | 結果 |
|---|---|---|---|
| 進入 New Character 頁 | frame-065 | 頁尾殘影 112 px | 通過 |
| 名字輸入 T／TE／TES／TEST | frame-068／071／074／077 | 頁尾殘影 112 px、名字游標 8 px | 通過 |
| 種族層 | frame-079 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 性別層 | frame-080 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 職業層 | frame-081 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 擲屬性層 | frame-082 | 頁尾殘影 112 px、滑鼠 55 px、數值欄 163 px | 通過 |

`newchar-pictures.json` 走得更遠：名字 `zed`、完整五層、進圖像層之後向左回繞一次、
再向右走完 80 張並回到起點，共 92 張畫面。圖像層的右頁說明、`CANCEL   ACCEPT` 提示、
青色雙格外框、名牌白字與右側上下箭頭都由狀態機畫出來，全部逐像素相同；80 張圖像的
blit 矩形 (208,106,40×38) 另外列為必須相同區。

四類允許差異的成因都查證過（RE-045）：

- **頁尾殘影**是原版行為，不是 remake 缺漏。原版把展開跨頁貼到前一個書本畫面之上時，
  左頁底緣有一條未被覆蓋的細帶，帶內每一個像素都等於 `BOOKPICS[0]`。差異是雙向的，
  所以不是素材解碼錯誤；貼圖矩形怎麼切還沒反組譯確認。
- **名字輸入游標**會閃爍，取樣落在哪一相位由畫面靜止的時點決定；游標的格子位置本身
  仍受驗證——跑到別的格子就會落在允許區外。
- **滑鼠指標**是 DOSBox 畫的，remake 沒有這個元素。
- **屬性數值欄**是亂數：原版以 BIOS tick 播種，remake 以固定 seed 重播，骰式相同、
  點數不同。標籤與冒號不在允許區內，因此擲屬性層十列標籤由青轉棕的高亮規則仍逐像素
  受檢。

## 翻書動畫

原版 `KOLBOOK.EXE` 的 `sub_114F5` 是表驅動的逐欄捲動（RE-048）：每一步把畫面上某一段
byte 欄整體左移一欄，再把目錄頁的一欄複製過來，只動第 10 列起的 180 列，每步等 3 個
BIOS tick。remake 依同一張表重播。

原版取樣間隔 0.3 秒、每步 0.165 秒，所以原版每兩步取到一張；18 對裡有 16 對是動畫
中間的畫面，逐像素完全相同，另外兩對是動畫的起點與終點，只差原版滑鼠游標。

原版擷取的 frame-020 沒有進對拍計畫：它是取樣落在第 25 步與第 26 步之間的更新中途
（前 84 列等於第 25 步、後 116 列等於第 26 步），不是 remake 的差異。

## 書中的其他兩段翻頁

同一張動畫參數表也驅動另外兩段翻頁，只是起點與來源不同：

| 段落 | 起點 | 來源 |
|---|---|---|
| 封面 → 目錄頁 | `BOOKPICS[0]` | `BOOKPICS[1]` |
| 目錄頁 → 角色選項頁 | `BOOKPICS[1]` | 還沒反白的角色選項頁 |
| 角色選項頁 → 角色清單 | 已取消反白的角色選項頁 | 角色清單頁 |

兩件由逐像素比對逼出來的原版行為：**反白是動畫結束之後才畫上去的**（動畫期間整份選單
都是青色），而**離開選單進入清單時，原版先把反白拿掉再開始捲動**。這兩點都不是猜的，
是把反白放錯位置時逐張畫面差 281 px、放對之後差 0 px 得到的。

角色清單的內容來自 `CHARDATA` 的 16 個槽位：名字在槽位起點、所在地是位移 64 起的固定
10 bytes、城鎮代碼是位移 144 起的 3 bytes，畫面顯示的是「所在地 空白 城鎮代碼」。
槽位位置由畫面上讀回的 14 個名字在 `CHARDATA` 內唯一命中確認（RE-049）。

## 刪除角色

刪除畫面用的就是角色清單頁，差別只有顏色語意：反白那一列黑字、其餘青字，左右兩欄
同時反白。Return 叫出 (94,58) 的 132×24 確認視窗，兩行暗紅字 `ARE YOU SURE?` 與
`YES/NO`；視窗開著時清單沒有任何一列反白，回答之後反白條重置為第 0 列。

| 步驟 | 原版畫格 | 允許差異區實際差異 | 結果 |
|---|---|---|---|
| 刪除清單（14 人，反白第 1 名） | frame-067 | 頁尾殘影 112 px、滑鼠 55 px | 通過 |
| 反白第 2 名 | frame-068 | 同上 | 通過 |
| 確認視窗 | frame-069 | 同上 | 通過 |
| 回答 YES 後的清單（13 人） | frame-070 | 滑鼠 55 px | 通過 |
| 反白第 2／3 名 | frame-071／072 | 滑鼠 55 px | 通過 |
| 刪除後清單上的確認視窗 | frame-073 | 滑鼠 55 px | 通過 |
| 回答 NO 回到清單 | frame-070 | 滑鼠 55 px | 通過 |

`delete-move.json` 另外對拍反白條的向下兩列、向上一列與 Return，5 對全過。刪除本身
只寫一個 byte，remake 的輸出與原版逐 byte 相同，見 `../chardata-delete/`。

這張清單不只給刪除用：`EDIT PICTURE`、`EDIT SHIELD`、`HARBOR MASTER` 進入後的第一張
畫面和它是同一個 PNG（RE-052），差別只在選定角色之後走到哪裡。

## HARBOR MASTER

這台機器只安裝一個 region，原版因此在展開跨頁上只畫三行黑字（`ONLY ONE REGION`、
`IS INSTALLED.`、空一列、`PRESS ANY KEY.`），沒有標題圖。remake 重繪後在允許差異區
之外零像素差。多個 region 的畫面沒有量測，remake 也沒有實作那條分支。

## 旅店畫面

城鎮畫面不是一張烘好的素材，而是紅磚底圖之上疊了幾塊：隊員頭像格、店主肖像
（`TWN7PICS[12]`）、標題框、對話框與六格圖示列。組成由
`tools/decompose_town_frame.py` 拆出，分群之外殘餘 0 像素（RE-054）。

| 步驟 | 原版畫格 | 允許差異區實際差異 | 結果 |
|---|---|---|---|
| 圖示白框第 1–6 格 | `town-icon-move-run1` frame-071…076 | 滑鼠 31 px | 通過 |
| 單人隊伍（白框在 COMPANION） | `town-recruit-companion-run1` frame-077 | 滑鼠 31 px | 通過 |
| 招募後的兩人隊伍 | 同上 frame-081 | 滑鼠 28 px | 通過 |

兩個原版收據是不同的執行，因此不是同一張畫面對自己。

## 這份收據不宣稱什麼

- 不宣稱角色建立已完成。接受圖像後寫入 CHARDATA、重名與名字欄字元集合都還沒對拍；
  角色清單目前讀的是原版磁碟上的 14 個既有角色，remake 自己建立的角色還不會寫回。
- 刪除流程的確認視窗按 ESC、大寫 `Y`／`N`、滑鼠按鍵與空名冊都沒有量測。
- `EDIT PICTURE`／`EDIT SHIELD` 的點陣圖編輯器沒有實作，也沒有對拍：原版是滑鼠驅動的，
  ESC 沒有作用，現有的鍵盤 oracle 走不完（RE-052）。
- 城鎮已對拍旅店、戶外地圖、門口、建築進門與地圖邊界的城門提示。商店、修道院與
  圖示列其餘五個圖示的行為都還沒量測，remake 也沒有實作。旅店畫面的內容目前來自對拍樣本
  `docs/re/town-inn-sample*.json`，不是從原版資料解出來的遊戲資料模型。
- 動畫已對拍：翻書動畫見上表，TITLE 標題序列見 `../title-sequence/`。尚未對拍的是
  戰鬥與野外地圖的動畫，那些畫面還沒實作——在城門選 DOOR 之後會走到野外地圖，
  remake 目前只把那件事記下來。
- 這十五份收據都是**英文模式**（`-lang en`）的結果。中文模式的版面是 remake 自訂的，
  原版沒有中文畫面，因此不對拍；見 `../cjk-ui/`。
